package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// sendAndReceive sends a request to a Server and returns the first response.
func sendAndReceive(t *testing.T, s *Server, method string, id any, params any) map[string]json.RawMessage {
	t.Helper()
	req := map[string]any{"jsonrpc": "2.0", "id": id, "method": method}
	if params != nil {
		req["params"] = params
	}
	var inBuf bytes.Buffer
	if err := WriteMessage(&inBuf, req); err != nil {
		t.Fatalf("write request: %v", err)
	}
	s.in = bufio.NewReader(&inBuf)

	msg, err := ReadMessage(s.in)
	if err != nil {
		t.Fatalf("read request: %v", err)
	}
	if err := s.dispatch(msg); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	resp, err := ReadMessage(bufio.NewReader(s.out.(*bytes.Buffer)))
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return resp
}

// newTestServer creates a Server with a bytes.Buffer as output.
func newTestServer() *Server {
	var out bytes.Buffer
	s := &Server{
		out:  &out,
		docs: make(map[string]string),
	}
	return s
}

// --- initialize ---

func TestInitialize(t *testing.T) {
	s := newTestServer()
	var outBuf bytes.Buffer
	s.out = &outBuf

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]any{},
	}
	var inBuf bytes.Buffer
	if err := WriteMessage(&inBuf, req); err != nil {
		t.Fatal(err)
	}
	s.in = bufio.NewReader(&inBuf)
	msg, _ := ReadMessage(s.in)
	if err := s.dispatch(msg); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	resp, err := ReadMessage(bufio.NewReader(&outBuf))
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	// Should have a result, no error.
	if _, ok := resp["result"]; !ok {
		t.Error("missing result in initialize response")
	}
	if _, ok := resp["error"]; ok {
		t.Error("unexpected error in initialize response")
	}

	// Result should have capabilities.
	var result InitializeResult
	_ = json.Unmarshal(resp["result"], &result)
	if result.Capabilities.TextDocumentSync != 1 {
		t.Errorf("TextDocumentSync: got %d, want 1", result.Capabilities.TextDocumentSync)
	}
	if !result.Capabilities.HoverProvider {
		t.Error("HoverProvider should be true")
	}
}

func TestUnknownMethodReturnsError(t *testing.T) {
	s := newTestServer()
	var outBuf bytes.Buffer
	s.out = &outBuf

	req := map[string]any{"jsonrpc": "2.0", "id": 9, "method": "noSuchMethod"}
	var inBuf bytes.Buffer
	_ = WriteMessage(&inBuf, req)
	s.in = bufio.NewReader(&inBuf)
	msg, _ := ReadMessage(s.in)
	_ = s.dispatch(msg)

	resp, _ := ReadMessage(bufio.NewReader(&outBuf))
	if _, ok := resp["error"]; !ok {
		t.Error("expected error response for unknown method")
	}
}

func TestUnknownNotificationIsIgnored(t *testing.T) {
	s := newTestServer()
	var outBuf bytes.Buffer
	s.out = &outBuf

	// A notification has no "id".
	req := map[string]any{"jsonrpc": "2.0", "method": "$/unknownNotification"}
	var inBuf bytes.Buffer
	_ = WriteMessage(&inBuf, req)
	s.in = bufio.NewReader(&inBuf)
	msg, _ := ReadMessage(s.in)
	if err := s.dispatch(msg); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	// No response should have been written.
	if outBuf.Len() > 0 {
		t.Errorf("expected no response for unknown notification, got %d bytes", outBuf.Len())
	}
}

// --- diagnostics ---

func TestDidOpenPublishesDiagnostics(t *testing.T) {
	s := newTestServer()
	var outBuf bytes.Buffer
	s.out = &outBuf

	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.aura",
			LanguageID: "aura",
			Version:    1,
			Text:       "module test\n\nfn add(a: Int, b: Int) -> Int:\n    return a + b\n",
		},
	}
	req := map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": params}
	var inBuf bytes.Buffer
	_ = WriteMessage(&inBuf, req)
	s.in = bufio.NewReader(&inBuf)
	msg, _ := ReadMessage(s.in)
	if err := s.dispatch(msg); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	// Should receive a publishDiagnostics notification.
	notif, err := ReadMessage(bufio.NewReader(&outBuf))
	if err != nil {
		t.Fatalf("read notification: %v", err)
	}
	method := MessageMethod(notif)
	if method != "textDocument/publishDiagnostics" {
		t.Errorf("expected publishDiagnostics, got %q", method)
	}

	var diagParams PublishDiagnosticsParams
	_ = json.Unmarshal(notif["params"], &diagParams)
	if diagParams.URI != "file:///test.aura" {
		t.Errorf("URI: %q", diagParams.URI)
	}
	// Valid code should have no diagnostics.
	if len(diagParams.Diagnostics) != 0 {
		t.Errorf("expected 0 diagnostics for valid code, got %d: %v", len(diagParams.Diagnostics), diagParams.Diagnostics)
	}
}

func TestDidOpenWithErrorPublishesDiagnostics(t *testing.T) {
	s := newTestServer()
	var outBuf bytes.Buffer
	s.out = &outBuf

	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///bad.aura",
			Text: "this is not valid aura !!!\n",
		},
	}
	req := map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": params}
	var inBuf bytes.Buffer
	_ = WriteMessage(&inBuf, req)
	s.in = bufio.NewReader(&inBuf)
	msg, _ := ReadMessage(s.in)
	_ = s.dispatch(msg)

	notif, _ := ReadMessage(bufio.NewReader(&outBuf))
	var diagParams PublishDiagnosticsParams
	_ = json.Unmarshal(notif["params"], &diagParams)
	if len(diagParams.Diagnostics) == 0 {
		t.Error("expected diagnostics for invalid code, got none")
	}
}

func TestDidCloseClearsDiagnostics(t *testing.T) {
	s := newTestServer()
	s.docs["file:///test.aura"] = "module test\n"
	var outBuf bytes.Buffer
	s.out = &outBuf

	params := DidCloseTextDocumentParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.aura"},
	}
	req := map[string]any{"jsonrpc": "2.0", "method": "textDocument/didClose", "params": params}
	var inBuf bytes.Buffer
	_ = WriteMessage(&inBuf, req)
	s.in = bufio.NewReader(&inBuf)
	msg, _ := ReadMessage(s.in)
	_ = s.dispatch(msg)

	notif, _ := ReadMessage(bufio.NewReader(&outBuf))
	var diagParams PublishDiagnosticsParams
	_ = json.Unmarshal(notif["params"], &diagParams)
	if len(diagParams.Diagnostics) != 0 {
		t.Errorf("expected empty diagnostics on close, got %d", len(diagParams.Diagnostics))
	}
	if _, ok := s.docs["file:///test.aura"]; ok {
		t.Error("doc should be removed from buffer after didClose")
	}
}

// --- hover ---

func TestHoverOnFunctionName(t *testing.T) {
	s := newTestServer()
	src := "module test\n\nfn add(a: Int, b: Int) -> Int:\n    return a + b\n"
	s.docs["file:///test.aura"] = src
	var outBuf bytes.Buffer
	s.out = &outBuf

	params := TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.aura"},
		Position:     Position{Line: 2, Character: 3}, // on "add"
	}
	req := map[string]any{"jsonrpc": "2.0", "id": 1, "method": "textDocument/hover", "params": params}
	var inBuf bytes.Buffer
	_ = WriteMessage(&inBuf, req)
	s.in = bufio.NewReader(&inBuf)
	msg, _ := ReadMessage(s.in)
	_ = s.dispatch(msg)

	resp, _ := ReadMessage(bufio.NewReader(&outBuf))
	if resp["result"] == nil {
		t.Fatal("expected hover result, got nil")
	}
	var hover Hover
	if err := json.Unmarshal(resp["result"], &hover); err != nil {
		t.Fatalf("unmarshal hover: %v", err)
	}
	if !strings.Contains(hover.Contents.Value, "add") {
		t.Errorf("hover should mention 'add', got: %q", hover.Contents.Value)
	}
}

// --- locate helpers ---

func TestWordAt(t *testing.T) {
	src := "module test\n\nfn myFunc(x: Int) -> Int:\n    return x\n"
	// Line 2 (0-based), col 3 — inside "myFunc"
	word, r := wordAt(src, Position{Line: 2, Character: 5})
	if word != "myFunc" {
		t.Errorf("word: got %q, want %q", word, "myFunc")
	}
	if r.Start.Character != 3 || r.End.Character != 9 {
		t.Errorf("range: %+v", r)
	}
}

func TestWordAtEmpty(t *testing.T) {
	src := "module test\n\n    return x\n"
	// Line 1 is blank
	word, _ := wordAt(src, Position{Line: 1, Character: 0})
	if word != "" {
		t.Errorf("expected empty word on blank line, got %q", word)
	}
}

func TestCheckSourceValidCode(t *testing.T) {
	src := "module test\n\nfn add(a: Int, b: Int) -> Int:\n    return a + b\n"
	diags := checkSource(src, "test.aura")
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for valid code, got %d: %v", len(diags), diags)
	}
}

func TestCheckSourceInvalidCode(t *testing.T) {
	src := "not valid aura !!!"
	diags := checkSource(src, "test.aura")
	if len(diags) == 0 {
		t.Error("expected diagnostics for invalid code, got none")
	}
}
