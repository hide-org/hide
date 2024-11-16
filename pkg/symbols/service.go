package symbols

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/hide-org/hide/pkg/lsp"
)

type Service interface {
	Search(ctx context.Context, query string, symbolFilter lsp.SymbolFilter) ([]lsp.SymbolInfo, error)
}

type ServiceImpl struct {
	lsp lsp.Service
}

func NewService(lsp lsp.Service) Service {
	return &ServiceImpl{lsp: lsp}
}

func (s *ServiceImpl) Search(ctx context.Context, query string, symbolFilter lsp.SymbolFilter) ([]lsp.SymbolInfo, error) {
	log.Debug().Str("query", query).Msg("Searching symbols")

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled")
	default:
	}

	symbols, err := s.lsp.GetWorkspaceSymbols(ctx, query, symbolFilter)
	if err != nil {
		log.Error().Err(err).Msg("failed to get workspace symbols")
		return nil, fmt.Errorf("failed to get workspace symbols: %w", err)
	}

	log.Debug().Str("query", query).Msgf("found %d symbols", len(symbols))
	return symbols, nil
}
