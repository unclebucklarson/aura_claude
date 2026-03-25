package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/unclebucklarson/aura/pkg/checker"
	"github.com/unclebucklarson/aura/pkg/lexer"
	"github.com/unclebucklarson/aura/pkg/parser"
)

// Server is the LSP server. It reads requests from in, writes responses to out,
// and logs diagnostics to logger (stderr by default).
type Server struct {
	in     *bufio.Reader
	out    io.Writer
	logger *log.Logger

	// docs maps document URI -> current text content.
	docs map[string]string

	// shutdown tracks whether the client sent shutdown.
	shutdown bool
}

// NewServer creates a new LSP server reading from in and writing to out.
func NewServer(in io.Reader, out io.Writer) *Server {
	return &Server{
		in:     bufio.NewReader(in),
		out:    out,
		logger: log.New(os.Stderr, "[aura-lsp] ", log.LstdFlags),
		docs:   make(map[string]string),
	}
}

// Run is the main dispatch loop. It reads messages until EOF or exit.
func (s *Server) Run() error {
	for {
		msg, err := ReadMessage(s.in)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			}
			return fmt.Errorf("read: %w", err)
		}
		if err := s.dispatch(msg); err != nil {
			s.logger.Printf("dispatch error: %v", err)
		}
	}
}

// dispatch routes a single raw message to the appropriate handler.
func (s *Server) dispatch(msg map[string]json.RawMessage) error {
	method := MessageMethod(msg)
	id := MessageID(msg)

	switch method {
	case "initialize":
		return s.handleInitialize(id, msg)
	case "initialized":
		return nil // client notification, no response needed
	case "shutdown":
		s.shutdown = true
		return WriteMessage(s.out, OKResponse(id, nil))
	case "exit":
		if s.shutdown {
			os.Exit(0)
		}
		os.Exit(1)
		return nil
	case "textDocument/didOpen":
		return s.handleDidOpen(msg)
	case "textDocument/didChange":
		return s.handleDidChange(msg)
	case "textDocument/didClose":
		return s.handleDidClose(msg)
	case "textDocument/hover":
		return s.handleHover(id, msg)
	case "textDocument/definition":
		return s.handleDefinition(id, msg)
	case "$/cancelRequest", "$/setTrace", "$/logTrace":
		return nil // silently ignore
	default:
		if id != nil {
			// Unknown request — respond with MethodNotFound.
			return WriteMessage(s.out, ErrResponse(id, CodeMethodNotFound,
				fmt.Sprintf("method not supported: %s", method)))
		}
		// Unknown notification — ignore.
		return nil
	}
}

// handleInitialize responds to the initialize request.
func (s *Server) handleInitialize(id any, _ map[string]json.RawMessage) error {
	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync:   1, // full sync
			HoverProvider:      true,
			DefinitionProvider: true,
		},
		ServerInfo: &ServerInfo{Name: "aura-lsp", Version: "1.0.0"},
	}
	return WriteMessage(s.out, OKResponse(id, result))
}

// handleDidOpen stores the document and publishes diagnostics.
func (s *Server) handleDidOpen(msg map[string]json.RawMessage) error {
	var params DidOpenTextDocumentParams
	if err := UnmarshalParams(msg, &params); err != nil {
		return err
	}
	s.docs[params.TextDocument.URI] = params.TextDocument.Text
	return s.publishDiagnostics(params.TextDocument.URI, params.TextDocument.Text)
}

// handleDidChange updates the document and re-publishes diagnostics.
func (s *Server) handleDidChange(msg map[string]json.RawMessage) error {
	var params DidChangeTextDocumentParams
	if err := UnmarshalParams(msg, &params); err != nil {
		return err
	}
	if len(params.ContentChanges) == 0 {
		return nil
	}
	// We use full sync (TextDocumentSync: 1), so the last change is the full text.
	text := params.ContentChanges[len(params.ContentChanges)-1].Text
	s.docs[params.TextDocument.URI] = text
	return s.publishDiagnostics(params.TextDocument.URI, text)
}

// handleDidClose removes the document from the buffer and clears diagnostics.
func (s *Server) handleDidClose(msg map[string]json.RawMessage) error {
	var params DidCloseTextDocumentParams
	if err := UnmarshalParams(msg, &params); err != nil {
		return err
	}
	delete(s.docs, params.TextDocument.URI)
	// Clear diagnostics for the closed document.
	return WriteMessage(s.out, Notification("textDocument/publishDiagnostics",
		PublishDiagnosticsParams{URI: params.TextDocument.URI, Diagnostics: []Diagnostic{}}))
}

// handleHover responds to a hover request.
func (s *Server) handleHover(id any, msg map[string]json.RawMessage) error {
	var params TextDocumentPositionParams
	if err := UnmarshalParams(msg, &params); err != nil {
		return err
	}
	text, ok := s.docs[params.TextDocument.URI]
	if !ok {
		return WriteMessage(s.out, OKResponse(id, nil))
	}
	filePath := uriToPath(params.TextDocument.URI)
	hover := computeHover(text, filePath, params.Position)
	return WriteMessage(s.out, OKResponse(id, hover))
}

// handleDefinition responds to a go-to-definition request.
func (s *Server) handleDefinition(id any, msg map[string]json.RawMessage) error {
	var params TextDocumentPositionParams
	if err := UnmarshalParams(msg, &params); err != nil {
		return err
	}
	text, ok := s.docs[params.TextDocument.URI]
	if !ok {
		return WriteMessage(s.out, OKResponse(id, nil))
	}
	filePath := uriToPath(params.TextDocument.URI)
	loc := computeDefinition(text, filePath, params.Position)
	return WriteMessage(s.out, OKResponse(id, loc))
}

// publishDiagnostics runs the lexer, parser, and type checker on src and
// pushes results to the client as textDocument/publishDiagnostics.
func (s *Server) publishDiagnostics(uri, src string) error {
	diags := checkSource(src, uriToPath(uri))
	params := PublishDiagnosticsParams{URI: uri, Diagnostics: diags}
	return WriteMessage(s.out, Notification("textDocument/publishDiagnostics", params))
}

// checkSource runs lex+parse+typecheck on src and returns LSP Diagnostics.
func checkSource(src, filePath string) []Diagnostic {
	var diags []Diagnostic

	// Lex.
	l := lexer.New(src, filePath)
	tokens, lexErrs := l.Tokenize()
	for _, e := range lexErrs {
		diags = append(diags, errorToDiagnostic(e.Error()))
	}
	if len(lexErrs) > 0 {
		return diags
	}

	// Parse.
	p := parser.New(tokens, filePath)
	mod, parseErrs := p.Parse()
	for _, e := range parseErrs {
		diags = append(diags, errorToDiagnostic(e.Error()))
	}
	if len(parseErrs) > 0 {
		return diags
	}

	// Type-check.
	c := checker.New(mod)
	checkErrs := c.Check()
	for _, e := range checkErrs {
		diags = append(diags, checkerErrToDiagnostic(e))
	}
	return diags
}

// errorToDiagnostic converts a "line:col: message" string to a Diagnostic.
// Falls back to line 0 col 0 if the string doesn't parse.
func errorToDiagnostic(msg string) Diagnostic {
	line, col, text := parseErrorPosition(msg)
	return Diagnostic{
		Range:    pointRange(line, col),
		Severity: SeverityError,
		Source:   "aura",
		Message:  text,
	}
}

// checkerErrToDiagnostic converts a *checker.CheckError to an LSP Diagnostic.
func checkerErrToDiagnostic(e *checker.CheckError) Diagnostic {
	startLine := e.Line - 1   // LSP is 0-based
	startCol := e.Column - 1
	endLine := e.EndLine - 1
	endCol := e.EndCol - 1
	if startLine < 0 {
		startLine = 0
	}
	if startCol < 0 {
		startCol = 0
	}
	if endLine < startLine {
		endLine = startLine
		endCol = startCol
	}
	if endCol < 0 {
		endCol = startCol
	}
	sev := SeverityError
	if e.Severity == checker.SeverityWarning {
		sev = SeverityWarning
	}
	return Diagnostic{
		Range: Range{
			Start: Position{Line: startLine, Character: startCol},
			End:   Position{Line: endLine, Character: endCol},
		},
		Severity: sev,
		Source:   "aura",
		Message:  e.Message,
	}
}

// parseErrorPosition extracts (line, col, message) from "line:col: message".
// Returns (0, 0, msg) if no position prefix is found.
func parseErrorPosition(msg string) (line, col int, text string) {
	// Try "file:line:col: message" or "line:col: message"
	// Strip a leading "filename:" prefix if present (contains a dot or slash)
	rest := msg
	if idx := strings.Index(rest, ":"); idx >= 0 {
		candidate := rest[:idx]
		if strings.ContainsAny(candidate, "./\\") {
			rest = rest[idx+1:] // strip filename prefix
		}
	}
	var l, c int
	var suffix string
	if n, _ := fmt.Sscanf(rest, "%d:%d:", &l, &c); n == 2 {
		// Find the ": " separator after line:col
		if i := strings.Index(rest, ": "); i >= 0 {
			suffix = strings.TrimSpace(rest[i+2:])
		} else {
			suffix = rest
		}
		return l - 1, c - 1, suffix // convert to 0-based
	}
	return 0, 0, msg
}

// pointRange creates a zero-length Range at the given (0-based) line/col.
func pointRange(line, col int) Range {
	if line < 0 {
		line = 0
	}
	if col < 0 {
		col = 0
	}
	return Range{
		Start: Position{Line: line, Character: col},
		End:   Position{Line: line, Character: col},
	}
}

// uriToPath converts a file:// URI to a filesystem path.
func uriToPath(uri string) string {
	if strings.HasPrefix(uri, "file://") {
		return uri[7:]
	}
	return uri
}

// pathToURI converts a filesystem path to a file:// URI.
func pathToURI(path string) string {
	if strings.HasPrefix(path, "/") {
		return "file://" + path
	}
	return "file:///" + path
}
