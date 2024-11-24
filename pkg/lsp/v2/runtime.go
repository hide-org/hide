package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"

	lang "github.com/hide-org/hide/pkg/lsp/v2/languages"
	"github.com/rs/zerolog/log"
)

func SetupServers(ctx context.Context, delegate lang.Delegate) error {
	var g errgroup.Group
	for _, adapter := range lang.Adapters {
		adapter := adapter // capture loop variable
		g.Go(func() error {
			err := runtime.setupServer(ctx, adapter, delegate)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to setup server")
			}
			return err
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// TODO: check concurrency safety
	runtime.delegate = delegate

	return nil
}

var runtime = run{
	support:   make(map[lang.LanguageID]lang.Adapter),
	bins:      make(map[lang.ServerName]*lang.Binary),
	processes: make(map[lang.ServerName]Process),
}

type run struct {
	sync.RWMutex

	delegate  lang.Delegate
	support   map[lang.LanguageID]lang.Adapter
	bins      map[lang.ServerName]*lang.Binary
	processes map[lang.ServerName]Process // I think it's better to register cmd type here

	// in the future we can add routines tha monitor liveliness of language server
}

func (r *run) setupServer(ctx context.Context, adapter lang.Adapter, delegate lang.Delegate) error {
	srv := adapter.Name()

	// TODO: think about this
	if ok := r.isReady(srv); ok {
		return nil
	}

	version, err := adapter.FetchLatestServerVersion(ctx, delegate)
	if err != nil {
		return err
	}

	bin, err := adapter.FetchServerBinary(ctx, version, delegate)
	if err != nil {
		return err
	}

	if err := r.registerBin(srv, bin); err != nil {
		return err
	}

	if err := r.registerSupport(srv, adapter); err != nil {
		return err
	}

	return nil
}

func (r *run) startServer(_ context.Context, language lang.LanguageID) (Process, error) {
	// TODO: check if process already running and if so return is running error
	command, err := r.getBin(language)
	if err != nil {
		return nil, err
	}

	// Start the language server
	process, err := NewProcess(*command)
	if err != nil {
		return nil, fmt.Errorf("failed to create language server process: %w", err)
	}

	if err := process.Start(); err != nil {
		return nil, fmt.Errorf("failed to start language server: %w", err)
	}

	r.Lock()
	defer r.Unlock()
	v, ok := r.support[language]
	if !ok {
		return nil, errors.New("language not found")
	}
	r.processes[v.Name()] = process

	return process, nil
}

func (r *run) getBin(language lang.LanguageID) (*lang.Binary, error) {
	r.RLock()
	defer r.Unlock()

	srv, ok := r.support[language]
	if !ok {
		return nil, errors.New("language is not supported")
	}

	bin, ok := r.bins[srv.Name()]
	if !ok {
		return nil, errors.New("corrupt runtime state: language support exist, binary not found")
	}

	return bin, nil
}

func (r *run) registerBin(srv lang.ServerName, bin *lang.Binary) error {
	r.Lock()
	defer r.Unlock()

	// should never happen but let's check
	if _, ok := r.bins[srv]; ok {
		return errors.New("server already registered")
	}
	r.bins[srv] = bin

	return nil
}

func (r *run) registerSupport(srv lang.ServerName, adapter lang.Adapter) error {
	r.Lock()
	defer r.Unlock()

	for _, v := range adapter.Languages() {
		// should never happen but let's check
		if _, ok := r.support[v]; ok {
			return errors.New("language already registered")
		}
		r.support[v] = adapter
	}

	return nil
}

func (r *run) isReady(srv lang.ServerName) bool {
	r.RLock()
	defer r.RUnlock()

	_, ok := r.bins[srv]
	return ok
}

func (r *run) serverInitOptions(ctx context.Context, lang lang.LanguageID) (json.RawMessage, error) {
	r.RLock()
	defer r.RUnlock()

	adapter, ok := r.support[lang]
	if !ok {
		return nil, nil
	}

	return adapter.InitializationOptions(ctx, r.delegate), nil
}
