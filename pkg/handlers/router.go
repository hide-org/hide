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
	r.Handle("/projects", handler).Methods("POST")
	return r
}

func (r *Router) WithDeleteProjectHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}", handler).Methods("DELETE")
	return r
}

func (r *Router) WithCreateTaskHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/tasks", handler).Methods("POST")
	return r
}

func (r *Router) WithListTasksHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/tasks", handler).Methods("GET")
	return r
}

func (r *Router) WithCreateFileHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/files", handler).Methods("POST")
	return r
}

func (r *Router) WithListFilesHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/files", handler).Methods("GET")
	return r
}

func (r *Router) WithReadFileHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/files/{path:.*}", handler).Methods("GET")
	return r
}

func (r *Router) WithUpdateFileHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/files/{path:.*}", handler).Methods("PUT")
	return r
}

func (r *Router) WithDeleteFileHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/files/{path:.*}", handler).Methods("DELETE")
	return r
}

func (r *Router) WithSearchFileHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/search", handler).Queries("type", "content", "query", "").Methods("GET")
	return r
}

func (r *Router) WithSearchSymbolsHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/search", handler).Queries("type", "symbol", "query", "").Methods("GET")
	return r
}

func (r *Router) WithDocumentOutlineHandler(handler http.Handler) *Router {
	r.Handle("/projects/{id}/outline/{path:.*}", handler).Methods(http.MethodGet)
	return r
}

func (r *Router) Build() *mux.Router {
	return r.Router
}
