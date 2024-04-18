package main

import (
	"fmt"
	"net/http"

	"github.com/artmoskvin/hide/pkg/handlers"
)

func main() {
	http.HandleFunc("/project", handlers.CreateProject)
	http.HandleFunc("/exec", handlers.ExecCmd)

	port := ":8080"
	err := http.ListenAndServe(port, nil)

	if err != nil {
		fmt.Println("Error starting server: ", err)
	}

	fmt.Println("Server started on port", port)
}
