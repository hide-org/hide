package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hide-org/hide/pkg/files"
)

func getProjectID(r *http.Request) (string, error) {
	return getPathValue(r, "id")
}

func getTimeOutSeconds(r *http.Request) int {
	var timeOut int
	if timeoutStr := r.Header.Get("X-Timeout-Seconds"); timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil {
			if t > 0 {
				timeOut = t
			}
		}
	}
	return timeOut
}

func GetFilePath(r *http.Request) (string, error) {
	return getPathValue(r, "path")
}

func getPathValue(r *http.Request, key string) (string, error) {
	vars := mux.Vars(r)
	value, ok := vars[key]
	if !ok || value == "" {
		return "", errors.New("invalid path value")
	}

	return value, nil
}

func getListFilesOptions(r *http.Request) []files.ListFileOption {
	opts := []files.ListFileOption{files.ListFilesWithFilter(getPatternFilter(r))}
	if r.URL.Query().Has("showHidden") {
		opts = append(opts, files.ListFilesWithShowHidden())
	}
	return opts
}

func getPatternFilter(r *http.Request) files.PatternFilter {
	filter := files.PatternFilter{}

	if r.URL.Query().Has("include") {
		filter.Include = r.URL.Query()["include"]
	}

	if r.URL.Query().Has("exclude") {
		filter.Exclude = r.URL.Query()["exclude"]
	}

	return filter
}

func parseIntQueryParam(params url.Values, paramName string) (int, bool, error) {
	param := params.Get(paramName)
	if param == "" {
		return 0, false, nil
	}

	value, err := strconv.Atoi(param)
	if err != nil {
		return 0, true, fmt.Errorf("invalid value for %s: %w", paramName, err)
	}

	return value, true, nil
}

func getAcceptFormat(r *http.Request) string {
	return r.Header.Get("Accept")
}
