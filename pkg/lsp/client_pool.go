package lsp

import "sync"

type ClientPool interface {
	Get(projectId ProjectId, languageId LanguageId) (Client, bool)
	GetAllForProject(projectId ProjectId) (map[LanguageId]Client, bool)
	Set(projectId ProjectId, languageId LanguageId, client Client)
	Delete(projectId ProjectId, languageId LanguageId)
	DeleteAllForProject(projectId ProjectId)
}

// In memory store for clients. Applies mutex locking for concurrent access.
type ClientPoolImpl struct {
	clients map[ProjectId]map[LanguageId]Client
	mu      sync.Mutex
}

func (c *ClientPoolImpl) Get(projectId ProjectId, languageId LanguageId) (Client, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if clients, ok := c.clients[projectId]; ok {
		if client, ok := clients[languageId]; ok {
			return client, true
		}
	}

	return nil, false
}

func (c *ClientPoolImpl) GetAllForProject(projectId ProjectId) (map[LanguageId]Client, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if clients, ok := c.clients[projectId]; ok {
		return clients, true
	}

	return nil, false
}

func (c *ClientPoolImpl) Set(projectId ProjectId, languageId LanguageId, client Client) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.clients[projectId]; !ok {
		c.clients[projectId] = make(map[LanguageId]Client)
	}

	c.clients[projectId][languageId] = client
}

func (c *ClientPoolImpl) Delete(projectId ProjectId, languageId LanguageId) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if clients, ok := c.clients[projectId]; ok {
		delete(clients, languageId)
	}
}

func (c *ClientPoolImpl) DeleteAllForProject(projectId ProjectId) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.clients, projectId)
}

func NewClientPool() *ClientPoolImpl {
	return &ClientPoolImpl{
		clients: make(map[ProjectId]map[LanguageId]Client),
	}
}
