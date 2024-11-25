package lsp

import (
	"sync"

	lang "github.com/hide-org/hide/pkg/lsp/v2/languages"
)

type ClientPool interface {
	Get(languageId lang.LanguageID) (Client, bool)
	GetAll() map[lang.LanguageID]Client
	Set(languageId lang.LanguageID, client Client)
	Delete(languageId lang.LanguageID)
}

// In memory store for clients. Applies mutex locking for concurrent access.
type ClientPoolImpl struct {
	clients map[lang.LanguageID]Client
	mu      sync.Mutex
}

func (c *ClientPoolImpl) Get(languageId lang.LanguageID) (Client, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	client, ok := c.clients[languageId]
	return client, ok
}

func (c *ClientPoolImpl) GetAll() map[lang.LanguageID]Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.clients
}

func (c *ClientPoolImpl) Set(languageId lang.LanguageID, client Client) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.clients[languageId] = client
	return
}

func (c *ClientPoolImpl) Delete(languageId lang.LanguageID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.clients, languageId)
}

func NewClientPool() *ClientPoolImpl {
	return &ClientPoolImpl{
		clients: make(map[lang.LanguageID]Client),
	}
}
