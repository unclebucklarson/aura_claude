package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// --- RPC framing ---

func TestWriteAndReadMessage(t *testing.T) {
	var buf bytes.Buffer
	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
	}
	if err := WriteMessage(&buf, msg); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	// Verify the framing looks correct.
	s := buf.String()
	if !strings.HasPrefix(s, "Content-Length: ") {
		t.Errorf("missing Content-Length header: %q", s[:min(len(s), 50)])
	}

	// Read it back.
	r := bufio.NewReader(&buf)
	got, err := ReadMessage(r)
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	var method string
	_ = json.Unmarshal(got["method"], &method)
	if method != "initialize" {
		t.Errorf("method: got %q, want %q", method, "initialize")
	}
}

func TestReadMessageMissingContentLength(t *testing.T) {
	// A message with no Content-Length header.
	raw := "\r\n{\"jsonrpc\":\"2.0\"}"
	r := bufio.NewReader(strings.NewReader(raw))
	_, err := ReadMessage(r)
	if err == nil {
		t.Fatal("expected error for missing Content-Length, got nil")
	}
}

func TestMessageIDString(t *testing.T) {
	raw := map[string]json.RawMessage{
		"id": json.RawMessage(`"req-1"`),
	}
	id := MessageID(raw)
	if id != "req-1" {
		t.Errorf("MessageID: got %v, want %q", id, "req-1")
	}
}

func TestMessageIDInt(t *testing.T) {
	raw := map[string]json.RawMessage{
		"id": json.RawMessage(`42`),
	}
	id := MessageID(raw)
	// JSON numbers unmarshal to float64 via any.
	if id.(float64) != 42 {
		t.Errorf("MessageID: got %v, want 42", id)
	}
}

func TestMessageIDMissing(t *testing.T) {
	raw := map[string]json.RawMessage{"method": json.RawMessage(`"shutdown"`)}
	id := MessageID(raw)
	if id != nil {
		t.Errorf("expected nil ID for notification, got %v", id)
	}
}

func TestMessageMethod(t *testing.T) {
	raw := map[string]json.RawMessage{"method": json.RawMessage(`"textDocument/hover"`)}
	m := MessageMethod(raw)
	if m != "textDocument/hover" {
		t.Errorf("method: got %q", m)
	}
}

func TestOKResponse(t *testing.T) {
	resp := OKResponse(1, map[string]string{"key": "value"})
	if resp.JSONRPC != "2.0" {
		t.Errorf("JSONRPC: %q", resp.JSONRPC)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error in OKResponse")
	}
}

func TestErrResponse(t *testing.T) {
	resp := ErrResponse(1, CodeMethodNotFound, "not found")
	if resp.Error == nil {
		t.Fatal("expected error in ErrResponse, got nil")
	}
	if resp.Error.Code != CodeMethodNotFound {
		t.Errorf("code: got %d, want %d", resp.Error.Code, CodeMethodNotFound)
	}
}

func TestRoundTripMultipleMessages(t *testing.T) {
	var buf bytes.Buffer
	for i := 0; i < 3; i++ {
		msg := map[string]any{"jsonrpc": "2.0", "id": i, "method": "ping"}
		if err := WriteMessage(&buf, msg); err != nil {
			t.Fatalf("WriteMessage %d: %v", i, err)
		}
	}
	r := bufio.NewReader(&buf)
	for i := 0; i < 3; i++ {
		got, err := ReadMessage(r)
		if err != nil {
			t.Fatalf("ReadMessage %d: %v", i, err)
		}
		var method string
		_ = json.Unmarshal(got["method"], &method)
		if method != "ping" {
			t.Errorf("msg %d: method %q", i, method)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
