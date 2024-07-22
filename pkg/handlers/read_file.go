package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/artmoskvin/hide/pkg/filemanager"
	"github.com/artmoskvin/hide/pkg/project"
)

type ReadFileHandler struct {
	Manager     project.Manager
	FileManager filemanager.FileManager
}

func (h ReadFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectId := r.PathValue("id")
	filePath := r.PathValue("path")
	queryParams := r.URL.Query()

	showLineNumbers, err := parseBoolQueryParam(queryParams, "showLineNumbers", filemanager.DefaultShowLineNumbers)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startLine, err := parseIntQueryParam(queryParams, "startLine", filemanager.DefaultStartLine)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	numLines, err := parseIntQueryParam(queryParams, "numLines", filemanager.DefaultNumLines)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	project, err := h.Manager.GetProject(projectId)

	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	file, err := h.FileManager.ReadFile(os.DirFS(project.Path), filePath, filemanager.NewReadProps(
		func(props *filemanager.ReadProps) {
			props.ShowLineNumbers = showLineNumbers
			props.StartLine = startLine
			props.NumLines = numLines
		},
	))

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(file)
}

func parseIntQueryParam(params url.Values, paramName string, defaultValue int) (int, error) {
	param := params.Get(paramName)

	if param == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(param)

	if err != nil {
		return 0, fmt.Errorf("Failed to parse %s: %w", paramName, err)
	}

	return value, nil
}

func parseBoolQueryParam(params url.Values, paramName string, defaultValue bool) (bool, error) {
	param := params.Get(paramName)

	if param == "" {
		return defaultValue, nil
	}

	value, err := strconv.ParseBool(param)

	if err != nil {
		return false, fmt.Errorf("Failed to parse %s: %w", paramName, err)
	}

	return value, nil
}
