package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

func PathValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("Invoking PathChecker")

		filePath, err := getFilePath(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid file path: %s", err), http.StatusBadRequest)
			return
		}

		if len(filePath) < 1 {
			http.Error(w, "Invalid file path: path is empty", http.StatusBadRequest)
			return
		}

		if strings.HasPrefix(filePath, "/") {
			http.Error(w, "Invalid file path: path starts with '/'", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
