package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

func getProjectID(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	projectID, ok := vars[key]
	if !ok {
		return "", errors.New("invalid project ID")
	}

	return projectID, nil
}
