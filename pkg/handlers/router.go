package handlers

import (
	"github.com/artmoskvin/hide/pkg/project"
	"github.com/gorilla/mux"
)

func Router(pm project.Manager) *mux.Router {
	router := mux.NewRouter()
	router.SkipClean(true)

	createProjectHandler := CreateProjectHandler{Manager: pm}
	deleteProjectHandler := DeleteProjectHandler{Manager: pm}
	createTaskHandler := CreateTaskHandler{Manager: pm}
	listTasksHandler := ListTasksHandler{Manager: pm}
	createFileHandler := CreateFileHandler{ProjectManager: pm}
	readFileHandler := ReadFileHandler{ProjectManager: pm}
	updateFileHandler := UpdateFileHandler{ProjectManager: pm}
	deleteFileHandler := DeleteFileHandler{ProjectManager: pm}
	listFilesHandler := ListFilesHandler{ProjectManager: pm}

	router.Handle("/projects", createProjectHandler).Methods("POST")
	router.Handle("/projects/{id}", deleteProjectHandler).Methods("DELETE")
	router.Handle("/projects/{id}/tasks", createTaskHandler).Methods("POST")
	router.Handle("/projects/{id}/tasks", listTasksHandler).Methods("GET")
	router.Handle("/projects/{id}/files", createFileHandler).Methods("POST")
	router.Handle("/projects/{id}/files", listFilesHandler).Methods("GET")
	router.Handle("/projects/{id}/files/{path:.*}", PathCheckerMiddleware(readFileHandler)).Methods("GET")
	router.Handle("/projects/{id}/files/{path:.*}", PathCheckerMiddleware(updateFileHandler)).Methods("PUT")
	router.Handle("/projects/{id}/files/{path:.*}", PathCheckerMiddleware(deleteFileHandler)).Methods("DELETE")

	return router
}
