package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/filariow/workspaces/server/core/workspace"
	"github.com/filariow/workspaces/server/persistence/cache"
	"github.com/filariow/workspaces/server/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const DefaultAddr string = ":8080"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// fetch configuration
	wns, ok := os.LookupEnv("WORKSPACES_NAMESPACE")
	if !ok {
		return fmt.Errorf("required Environment Variable WORKSPACES_NAMESPACE not found")
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	// setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setup read model
	c, err := cache.New(cfg, wns)
	if err != nil {
		return err
	}

	// setup write model
	// TODO

	// setup REST over HTTP server
	s := rest.New(
		DefaultAddr,
		workspace.NewReadWorkspaceHandler(c).Handle,
		workspace.NewListWorkspaceHandler(c).Handle,
	)

	// HTTP Server graceful shutdown
	go func() {
		<-ctx.Done()

		sctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := s.Shutdown(sctx); err != nil {
			log.Fatal(err)
		}
	}()

	// start the cache
	go func() {
		if err := c.Start(ctx); err != nil {
			if ctx.Err() == nil {
				cancel()
			}
		}
	}()

	// start HTTP server
	if err := s.ListenAndServe(); err != nil {
		return fmt.Errorf("error running server: %v", err)
	}
	return nil
}
