// Package lsp provides types and utilities for the Language Server Protocol.
//
// Implements a subset of LSP 3.17 sufficient for diagnostics, hover, and
// go-to-definition. No external dependencies — pure stdlib JSON.
package lsp

// --- JSON-RPC 2.0 message types ---

// RequestMessage is an LSP request from client to server.
type RequestMessage struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"` // int | string | null
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// ResponseMessage is an LSP response from server to client.
type ResponseMessage struct {
	JSONRPC string  `json:"jsonrpc"`
	ID      any     `json:"id"`
	Result  any     `json:"result,omitempty"`
	Error   *RPCErr `json:"error,omitempty"`
}

// NotificationMessage is a one-way LSP message (no ID, no response expected).
type NotificationMessage struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// RPCErr represents a JSON-RPC error object.
type RPCErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Standard JSON-RPC error codes.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// --- Core LSP position/range types ---

// Position is a 0-based line and character offset (UTF-16 code units per spec,
// but we treat source as UTF-8 byte offsets for simplicity).
type Position struct {
	Line      int `json:"line"`      // 0-based
	Character int `json:"character"` // 0-based
}

// Range is a start/end Position pair.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location is a URI + Range.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// --- Diagnostic ---

// DiagnosticSeverity mirrors the LSP DiagnosticSeverity enum.
type DiagnosticSeverity int

const (
	SeverityError       DiagnosticSeverity = 1
	SeverityWarning     DiagnosticSeverity = 2
	SeverityInformation DiagnosticSeverity = 3
	SeverityHint        DiagnosticSeverity = 4
)

// Diagnostic represents a compiler error/warning at a location.
type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity,omitempty"`
	Source   string             `json:"source,omitempty"`
	Message  string             `json:"message"`
}

// --- initialize ---

// ClientCapabilities (minimal — we only inspect what we need).
type ClientCapabilities struct {
	TextDocument *TextDocumentClientCapabilities `json:"textDocument,omitempty"`
}

// TextDocumentClientCapabilities holds text-document feature flags.
type TextDocumentClientCapabilities struct {
	Hover *HoverClientCapabilities `json:"hover,omitempty"`
}

// HoverClientCapabilities signals markdown support.
type HoverClientCapabilities struct {
	ContentFormat []string `json:"contentFormat,omitempty"`
}

// InitializeParams is sent by the client at startup.
type InitializeParams struct {
	ProcessID    *int               `json:"processId,omitempty"`
	RootURI      string             `json:"rootUri,omitempty"`
	Capabilities ClientCapabilities `json:"capabilities"`
}

// InitializeResult is the server's response to initialize.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

// ServerCapabilities advertises what the server supports.
type ServerCapabilities struct {
	TextDocumentSync   int  `json:"textDocumentSync"`   // 1 = full sync
	HoverProvider      bool `json:"hoverProvider"`
	DefinitionProvider bool `json:"definitionProvider"`
}

// ServerInfo is the server's name/version, sent in the initialize response.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// --- textDocument/didOpen, didChange, didClose ---

// TextDocumentItem carries the full content of a document.
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// VersionedTextDocumentIdentifier identifies a document + version.
type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

// TextDocumentIdentifier identifies a document by URI only.
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// DidOpenTextDocumentParams is the params for textDocument/didOpen.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// TextDocumentContentChangeEvent carries the new full text on change.
type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

// DidChangeTextDocumentParams is the params for textDocument/didChange.
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier   `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent  `json:"contentChanges"`
}

// DidCloseTextDocumentParams is the params for textDocument/didClose.
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// PublishDiagnosticsParams is the params for textDocument/publishDiagnostics.
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// --- textDocument/hover ---

// TextDocumentPositionParams identifies a document + cursor position.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// MarkupContent is a markdown or plaintext content value.
type MarkupContent struct {
	Kind  string `json:"kind"`  // "markdown" or "plaintext"
	Value string `json:"value"`
}

// Hover is the result of a hover request.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// --- textDocument/definition ---

// DefinitionParams is the params for textDocument/definition.
type DefinitionParams = TextDocumentPositionParams
