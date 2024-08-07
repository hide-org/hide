package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/artmoskvin/hide/pkg/files"
	"github.com/artmoskvin/hide/pkg/project"
)

type ReadFileHandler struct {
	ProjectManager project.Manager
}

func (h ReadFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID, err := getProjectID(r)
	if err != nil {
		http.Error(w, "invalid project ID", http.StatusBadRequest)
	}

	filePath, err := getFilePath(r)
	if err != nil {
		http.Error(w, "invalid file path", http.StatusBadRequest)
	}

	queryParams := r.URL.Query()

	showLineNumbers, err := parseBoolQueryParam(queryParams, "showLineNumbers", files.DefaultShowLineNumbers)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startLine, err := parseIntQueryParam(queryParams, "startLine", files.DefaultStartLine)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	numLines, err := parseIntQueryParam(queryParams, "numLines", files.DefaultNumLines)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, err := h.ProjectManager.ReadFile(r.Context(), projectID, filePath, files.NewReadProps(
		func(props *files.ReadProps) {
			props.ShowLineNumbers = showLineNumbers
			props.StartLine = startLine
			props.NumLines = numLines
		},
	))

	if err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

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
