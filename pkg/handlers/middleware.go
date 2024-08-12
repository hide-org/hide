package handlers

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

func PathCheckerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("Invoking PathCheckerMiddleware")

		filePath, err := getFilePath(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid file path: %s", err), http.StatusBadRequest)
			return
		}

		if len(filePath) < 1 {
			http.Error(w, "Invalid file path: path is empty", http.StatusBadRequest)
			return
		}

		if filePath[0:1] == "/" {
			http.Error(w, "Invalid file path: path starts with '/'", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
