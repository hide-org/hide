package lsp

import "sync"

type ClientPool interface {
	Get(languageId LanguageId) (Client, bool)
	GetAll() map[LanguageId]Client
	Set(languageId LanguageId, client Client)
	Delete(languageId LanguageId)
}

// In memory store for clients. Applies mutex locking for concurrent access.
type ClientPoolImpl struct {
	clients map[LanguageId]Client
	mu      sync.Mutex
}

func (c *ClientPoolImpl) Get(languageId LanguageId) (Client, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	client, ok := c.clients[languageId]
	return client, ok
}

func (c *ClientPoolImpl) GetAll() map[LanguageId]Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.clients
}

func (c *ClientPoolImpl) Set(languageId LanguageId, client Client) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.clients[languageId] = client
	return
}

func (c *ClientPoolImpl) Delete(languageId LanguageId) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// TODO: can this thing fail if the languageId is not found?
	delete(c.clients, languageId)
}



func NewClientPool() *ClientPoolImpl {
	return &ClientPoolImpl{
		clients: make(map[LanguageId]Client),
	}
}
