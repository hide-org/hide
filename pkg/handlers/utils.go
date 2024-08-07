package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

func getProjectID(r *http.Request) (string, error) {
	return getPathValue(r, "id")
}

func getFilePath(r *http.Request) (string, error) {
	return getPathValue(r, "path")
}

func getPathValue(r *http.Request, key string) (string, error) {
	vars := mux.Vars(r)
	value, ok := vars[key]
	if !ok {
		return "", errors.New("invalid path value")
	}

	return value, nil
}
