// Command aura-lsp is the Aura Language Server.
//
// It implements a subset of LSP 3.17 over stdin/stdout:
//   - textDocument/publishDiagnostics (run checker on every change)
//   - textDocument/hover              (show signature + doc comment)
//   - textDocument/definition         (go to top-level definition)
//
// Usage: pipe stdin/stdout to your editor's LSP client.
// The server reads JSONRPC messages from stdin and writes to stdout.
// Diagnostic/debug output goes to stderr.
package main

import (
	"os"

	"github.com/unclebucklarson/aura/pkg/lsp"
)

func main() {
	server := lsp.NewServer(os.Stdin, os.Stdout)
	if err := server.Run(); err != nil {
		os.Stderr.WriteString("aura-lsp: " + err.Error() + "\n")
		os.Exit(1)
	}
}
