package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/artmoskvin/hide/pkg/model"
	"github.com/artmoskvin/hide/pkg/project"
)

const (
	queryKey      = "query"
	searchTypeKey = "search"
)

type SearchFilesHandler struct {
	ProjectManager project.Manager
}

func (h SearchFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID, err := getProjectID(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid project ID: %s", err), http.StatusBadRequest)
		return
	}

	query := r.URL.Query().Get(queryKey)
	if query == "" {
		http.Error(w, "Query not specified", http.StatusBadRequest)
		return
	}

	typ := r.URL.Query().Get(searchTypeKey)

	check, err := getChecker(typ, query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Bad query: %s", err), http.StatusInternalServerError)
	}

	listFiles := func(ctx context.Context, showHidden bool) ([]*model.File, error) {
		return h.ProjectManager.ListFiles(ctx, projectID, showHidden)
	}

	result, err := findInFiles(r.Context(), listFiles, check)
	if err != nil {
		var projectNotFoundError *project.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			http.Error(w, projectNotFoundError.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("Failed to search: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func getChecker(typ, query string) (check func(s string) bool, err error) {
	switch typ {
	case "exact":
		return exactSearch(query)
	case "grep":
		return grepSearch(query)
	default:
		return caseInsensitiveSearch(query)
	}
}

func caseInsensitiveSearch(query string) (check func(s string) bool, err error) {
	q := strings.ToLower(query)
	check = func(s string) bool {
		return strings.Contains(strings.ToLower(s), q)
	}

	return
}

func exactSearch(query string) (check func(s string) bool, err error) {
	return func(s string) bool {
		return strings.Contains(s, query)
	}, nil
}

func grepSearch(query string) (check func(s string) bool, err error) {
	re, err := regexp.Compile(query)
	if err != nil {
		return nil, err
	}

	return func(s string) bool {
		return re.MatchString(s)
	}, nil
}

func findInFiles(ctx context.Context, listFiles func(ctx context.Context, showHidden bool) ([]*model.File, error), check func(s string) bool) ([]model.File, error) {
	files, err := listFiles(ctx, false)
	if err != nil {
		return nil, err
	}

	resultFiles := make([]model.File, 0)
	for _, file := range files {
		resultLines := make([]model.Line, 0)

		for _, line := range file.Lines {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled")
			default:
				if check(line.Content) {
					resultLines = append(resultLines, line)
				}
			}
		}

		if len(resultLines) != 0 {
			resultFiles = append(resultFiles, model.File{
				Path:  file.Path,
				Lines: resultLines,
			})
		}
	}

	return resultFiles, nil
}
