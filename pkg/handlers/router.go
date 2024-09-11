package handlers

import (
	"github.com/gorilla/mux"
)

type Router struct {
	*mux.Router
}

func NewRouter() *Router {
	r := &Router{
		Router: mux.NewRouter(),
	}
	r.SkipClean(true)
	return r
}

func (r *Router) WithCreateProjectHandler(handler CreateProjectHandler) *Router {
	r.Handle("/projects", handler).Methods("POST")
	return r
}

func (r *Router) WithDeleteProjectHandler(handler DeleteProjectHandler) *Router {
	r.Handle("/projects/{id}", handler).Methods("DELETE")
	return r
}

func (r *Router) WithCreateTaskHandler(handler CreateTaskHandler) *Router {
	r.Handle("/projects/{id}/tasks", handler).Methods("POST")
	return r
}

func (r *Router) WithListTasksHandler(handler ListTasksHandler) *Router {
	r.Handle("/projects/{id}/tasks", handler).Methods("GET")
	return r
}

func (r *Router) WithCreateFileHandler(handler CreateFileHandler) *Router {
	r.Handle("/projects/{id}/files", handler).Methods("POST")
	return r
}

func (r *Router) WithListFilesHandler(handler ListFilesHandler) *Router {
	r.Handle("/projects/{id}/files", handler).Methods("GET")
	return r
}

func (r *Router) WithReadFileHandler(handler ReadFileHandler) *Router {
	r.Handle("/projects/{id}/files/{path:.*}", PathValidator(handler)).Methods("GET")
	return r
}

func (r *Router) WithUpdateFileHandler(handler UpdateFileHandler) *Router {
	r.Handle("/projects/{id}/files/{path:.*}", PathValidator(handler)).Methods("PUT")
	return r
}

func (r *Router) WithDeleteFileHandler(handler DeleteFileHandler) *Router {
	r.Handle("/projects/{id}/files/{path:.*}", PathValidator(handler)).Methods("DELETE")
	return r
}

func (r *Router) WithSearchFileHandler(handler SearchFilesHandler) *Router {
	r.Handle("/projects/{id}/search", handler).Queries("type", "content", "query", "").Methods("GET")
	return r
}

func (r *Router) WithSearchSymbolsHandler(handler SearchSymbolsHandler) *Router {
	r.Handle("/projects/{id}/search", handler).Queries("type", "symbol", "query", "").Methods("GET")
	return r
}

func (r *Router) Build() *mux.Router {
	return r.Router
}
