package main

import (
	"fmt"
	"log"

	"github.com/filariow/workspaces/server/rest"
)

const DefaultAddr string = ":8080"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	s := rest.New(DefaultAddr)

	if err := s.ListenAndServe(); err != nil {
		return fmt.Errorf("error running server: %v", err)
	}
	return nil
}
