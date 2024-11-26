package handlers

import (
	"net/http"

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

func (r *Router) WithCreateProjectHandler(handler http.Handler) *Router {
	r.Handle("/projects", handler).Methods(http.MethodPost)
	return r
}

func (r *Router) WithDeleteProjectHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}", handler).Methods(http.MethodDelete)
	return r
}

func (r *Router) WithCreateTaskHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/tasks", handler).Methods(http.MethodPost)
	return r
}

func (r *Router) WithGetProjectHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}", handler).Methods(http.MethodGet)
	return r
}

func (r *Router) WithGetProjectsHandler(handler http.Handler) *Router {
	r.Handle("/projects", handler).Methods(http.MethodGet)
	return r
}

func (r *Router) WithListTasksHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/tasks", handler).Methods(http.MethodGet)
	return r
}

func (r *Router) WithCreateFileHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/files", handler).Methods(http.MethodPost)
	return r
}

func (r *Router) WithListFilesHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/files", handler).Methods(http.MethodGet)
	return r
}

func (r *Router) WithReadFileHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/files/{path:.*}", handler).Methods(http.MethodGet)
	return r
}

func (r *Router) WithUpdateFileHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/files/{path:.*}", handler).Methods(http.MethodPut)
	return r
}

func (r *Router) WithDeleteFileHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/files/{path:.*}", handler).Methods(http.MethodDelete)
	return r
}

func (r *Router) WithSearchFileHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/search", handler).Queries("type", "content", "query", "").Methods(http.MethodGet)
	return r
}

func (r *Router) WithSearchSymbolsHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/search", handler).Queries("type", "symbol", "query", "").Methods(http.MethodGet)
	return r
}

func (r *Router) WithDocumentOutlineHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/outline/{path:.*}", handler).Methods(http.MethodGet)
	return r
}

func (r *Router) Build() *mux.Router {
	return r.Router
}
