package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/hide-org/hide/pkg/project/v2"
)

type CreateProjectHandler struct {
	Manager   project.Manager
	Validator *validator.Validate
}

func (h CreateProjectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request project.CreateProjectRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed parsing request body", http.StatusBadRequest)
		return
	}

	err := h.Validator.StructCtx(r.Context(), request)
	if err != nil {

		if _, ok := err.(*validator.InvalidValidationError); ok {
			http.Error(w, fmt.Sprintf("Validation error: %s", err), http.StatusInternalServerError)
			return
		}

		if errs, ok := err.(validator.ValidationErrors); ok {
			http.Error(w, fmt.Sprintf("Validation error: %s", errs), http.StatusBadRequest)
			return
		}

		http.Error(w, fmt.Sprintf("Unknown validation error: %s", err), http.StatusInternalServerError)
		return
	}

	p, err := h.Manager.CreateProject(r.Context(), request)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create project: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}
