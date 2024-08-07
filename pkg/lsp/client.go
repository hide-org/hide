package lsp

import (
	"context"
	"encoding/json"

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
	NotifyDidClose(ctx context.Context, params protocol.DidCloseTextDocumentParams) error
	// TODO: check if any LSP server supports this
	// PullDiagnostics(ctx context.Context, params DocumentDiagnosticParams) (DocumentDiagnosticReport, error)
	StopServer() error
}
type ClientImpl struct {
	conn               Connection
	server             Process
	diagnosticsChannel chan protocol.PublishDiagnosticsParams
}

func NewClient(ctx context.Context, server Process, diagnosticsChannel chan protocol.PublishDiagnosticsParams) Client {
	handler := &lspHandler{
		diagnosticsHandler: func(params protocol.PublishDiagnosticsParams) {
			diagnosticsChannel <- params
		},
	}
	conn := NewConnection(ctx, server.ReadWriteCloser(), jsonrpc2.HandlerWithError(handler.Handle))
	return &ClientImpl{conn: conn, server: server, diagnosticsChannel: diagnosticsChannel}
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

func (c *ClientImpl) NotifyDidClose(ctx context.Context, params protocol.DidCloseTextDocumentParams) error {
	return c.conn.Notify(ctx, "textDocument/didClose", params)
}

// func (c *ClientImpl) PullDiagnostics(ctx context.Context, params DocumentDiagnosticParams) (DocumentDiagnosticReport, error) {
// 	var result DocumentDiagnosticReport
// 	err := c.conn.Call(ctx, "textDocument/diagnostic", params, &result)
// 	return result, err
// }

func (c *ClientImpl) StopServer() error {
	return c.server.Stop()
}
