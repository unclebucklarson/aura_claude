package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ReadMessage reads one LSP message from r.
// LSP framing: "Content-Length: N\r\n\r\n" followed by N bytes of JSON.
func ReadMessage(r *bufio.Reader) (map[string]json.RawMessage, error) {
	// Read headers until the blank line.
	contentLength := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break // blank line = end of headers
		}
		if strings.HasPrefix(line, "Content-Length: ") {
			val := strings.TrimPrefix(line, "Content-Length: ")
			n, err := strconv.Atoi(strings.TrimSpace(val))
			if err != nil {
				return nil, fmt.Errorf("bad Content-Length: %q", val)
			}
			contentLength = n
		}
		// Ignore any other headers (Content-Type, etc.)
	}
	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("reading message body: %w", err)
	}

	var msg map[string]json.RawMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("parsing message JSON: %w", err)
	}
	return msg, nil
}

// WriteMessage serialises v as JSON and writes it with LSP framing to w.
func WriteMessage(w io.Writer, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshalling message: %w", err)
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if _, err := io.WriteString(w, header); err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

// UnmarshalParams decodes the "params" field of a raw message into dst.
func UnmarshalParams(raw map[string]json.RawMessage, dst any) error {
	p, ok := raw["params"]
	if !ok {
		return nil // params is optional
	}
	return json.Unmarshal(p, dst)
}

// MessageID extracts the request ID from a raw message.
// Returns nil if no id field is present (i.e. it's a notification).
func MessageID(raw map[string]json.RawMessage) any {
	idRaw, ok := raw["id"]
	if !ok {
		return nil
	}
	var id any
	_ = json.Unmarshal(idRaw, &id)
	return id
}

// MessageMethod extracts the method string from a raw message.
func MessageMethod(raw map[string]json.RawMessage) string {
	methodRaw, ok := raw["method"]
	if !ok {
		return ""
	}
	var method string
	_ = json.Unmarshal(methodRaw, &method)
	return method
}

// OKResponse builds a successful ResponseMessage.
func OKResponse(id, result any) ResponseMessage {
	return ResponseMessage{JSONRPC: "2.0", ID: id, Result: result}
}

// ErrResponse builds an error ResponseMessage.
func ErrResponse(id any, code int, msg string) ResponseMessage {
	return ResponseMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCErr{Code: code, Message: msg},
	}
}

// Notification builds a NotificationMessage.
func Notification(method string, params any) NotificationMessage {
	return NotificationMessage{JSONRPC: "2.0", Method: method, Params: params}
}
