package main

import (
	"fmt"
	"os"

	"github.com/hackisonjd/http-server/pkg/http"
)

func main() {
	// Init a new server on just localhost interfaces for now
	fmt.Println("creating new server!")
	server := http.NewServer("127.0.0.1", "8080")

	fmt.Println("starting the server")
	// Start the server
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("could not start server: %v\n", err)
		os.Exit(1)
	}
}
