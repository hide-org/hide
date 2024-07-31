package languageserver

import (
	"context"
	"encoding/json"
	"io"
	"log"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// DocumentDiagnosticParams represents the parameters for a textDocument/diagnostic request
// type DocumentDiagnosticParams struct {
// 	TextDocument     lsp.TextDocumentIdentifier `json:"textDocument"`
// 	PreviousResultId string                     `json:"previousResultId,omitempty"`
// }

// DocumentDiagnosticReport represents the response from a textDocument/diagnostic request
// type DocumentDiagnosticReport struct {
// 	Kind             string                              `json:"kind"`
// 	ResultId         string                              `json:"resultId,omitempty"`
// 	Items            []lsp.Diagnostic                    `json:"items,omitempty"`
// 	RelatedDocuments map[string]DocumentDiagnosticReport `json:"relatedDocuments,omitempty"`
// }

type lspHandler struct {
	diagnosticsHandler func(lsp.PublishDiagnosticsParams)
}

func (h *lspHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	switch req.Method {
	case "textDocument/publishDiagnostics":
		var params lsp.PublishDiagnosticsParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		h.diagnosticsHandler(params)
	}
	return nil, nil
}

type Client interface {
	Initialize(ctx context.Context, params lsp.InitializeParams) (lsp.InitializeResult, error)
	PullDiagnostics(ctx context.Context, params DocumentDiagnosticParams) (DocumentDiagnosticReport, error)
}
type ClientImpl struct {
	conn Connection
}

func NewClient(ctx context.Context, rwc io.ReadWriteCloser) Client {
	handler := &lspHandler{
		diagnosticsHandler: func(params lsp.PublishDiagnosticsParams) {
			log.Printf("Received diagnostics for %s: %+v", params.URI, params.Diagnostics)
		},
	}
	conn := NewConnection(ctx, rwc, jsonrpc2.HandlerWithError(handler.Handle))
	return &ClientImpl{conn: conn}
}

func (c *ClientImpl) Initialize(ctx context.Context, params lsp.InitializeParams) (lsp.InitializeResult, error) {
	var result lsp.InitializeResult
	err := c.conn.Call(ctx, "initialize", params, &result)
	return result, err
}

func (c *ClientImpl) PullDiagnostics(ctx context.Context, params DocumentDiagnosticParams) (DocumentDiagnosticReport, error) {
	var result DocumentDiagnosticReport
	err := c.conn.Call(ctx, "textDocument/diagnostic", params, &result)
	return result, err
}
