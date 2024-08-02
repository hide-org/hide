package languageserver

import (
	"context"
	"encoding/json"
	"io"
	"log"

	"github.com/sourcegraph/jsonrpc2"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

type lspHandler struct {
	diagnosticsHandler func(protocol.PublishDiagnosticsParams)
}

func (h *lspHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	switch req.Method {
	case "textDocument/publishDiagnostics":
		var params protocol.PublishDiagnosticsParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		h.diagnosticsHandler(params)
	}
	return nil, nil
}

type Client interface {
	Initialize(ctx context.Context, params protocol.InitializeParams) (protocol.InitializeResult, error)
	NotifyInitialized(ctx context.Context) error
	NotifyDidOpen(ctx context.Context, params protocol.DidOpenTextDocumentParams) error
	NotifyDidChange(ctx context.Context, params protocol.DidChangeTextDocumentParams) error
	NotifyDidChangeWorkspaceFolders(ctx context.Context, params protocol.DidChangeWorkspaceFoldersParams) error
	// TODO: check if any LSP server supports this
	// PullDiagnostics(ctx context.Context, params DocumentDiagnosticParams) (DocumentDiagnosticReport, error)
}
type ClientImpl struct {
	conn Connection
}

func NewClient(ctx context.Context, rwc io.ReadWriteCloser, diagnosticsChannel chan protocol.PublishDiagnosticsParams) Client {
	handler := &lspHandler{
		diagnosticsHandler: func(params protocol.PublishDiagnosticsParams) {
			log.Printf("Received diagnostics for %s: %+v", params.URI, params.Diagnostics)
			diagnosticsChannel <- params
			log.Printf("Sent diagnostics for %s", params.URI)
		},
	}
	conn := NewConnection(ctx, rwc, jsonrpc2.HandlerWithError(handler.Handle))
	return &ClientImpl{conn: conn}
}

func (c *ClientImpl) Initialize(ctx context.Context, params protocol.InitializeParams) (protocol.InitializeResult, error) {
	var result protocol.InitializeResult
	err := c.conn.Call(ctx, "initialize", params, &result)
	return result, err
}

func (c *ClientImpl) NotifyInitialized(ctx context.Context) error {
	return c.conn.Notify(ctx, "initialized", nil)
}

func (c *ClientImpl) NotifyDidOpen(ctx context.Context, params protocol.DidOpenTextDocumentParams) error {
	return c.conn.Notify(ctx, "textDocument/didOpen", params)
}

func (c *ClientImpl) NotifyDidChange(ctx context.Context, params protocol.DidChangeTextDocumentParams) error {
	return c.conn.Notify(ctx, "textDocument/didChange", params)
}

func (c *ClientImpl) NotifyDidChangeWorkspaceFolders(ctx context.Context, params protocol.DidChangeWorkspaceFoldersParams) error {
	return c.conn.Notify(ctx, "workspace/didChangeWorkspaceFolders", params)
}

// func (c *ClientImpl) PullDiagnostics(ctx context.Context, params DocumentDiagnosticParams) (DocumentDiagnosticReport, error) {
// 	var result DocumentDiagnosticReport
// 	err := c.conn.Call(ctx, "textDocument/diagnostic", params, &result)
// 	return result, err
// }
