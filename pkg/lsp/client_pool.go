package lsp

import "sync"

// In memory store for clients. Applies mutex locking for concurrent access.
type ClientPool struct {
	clients map[ProjectId]map[LanguageId]Client
	mu      sync.Mutex
}

func (c *ClientPool) Get(projectId ProjectId, languageId LanguageId) (Client, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if clients, ok := c.clients[projectId]; ok {
		if client, ok := clients[languageId]; ok {
			return client, true
		}
	}

	return nil, false
}

func (c *ClientPool) GetAllForProject(projectId ProjectId) (map[LanguageId]Client, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if clients, ok := c.clients[projectId]; ok {
		return clients, true
	}

	return nil, false
}

func (c *ClientPool) Set(projectId ProjectId, languageId LanguageId, client Client) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.clients[projectId]; !ok {
		c.clients[projectId] = make(map[LanguageId]Client)
	}

	c.clients[projectId][languageId] = client
}

func (c *ClientPool) Delete(projectId ProjectId, languageId LanguageId) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if clients, ok := c.clients[projectId]; ok {
		delete(clients, languageId)
	}
}

func (c *ClientPool) DeleteAllForProject(projectId ProjectId) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.clients, projectId)
}

func NewClientPool() *ClientPool {
	return &ClientPool{
		clients: make(map[ProjectId]map[LanguageId]Client),
	}
}
