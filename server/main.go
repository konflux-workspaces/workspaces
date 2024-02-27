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
	log.Printf("Workspaces namespace is: %s", wns)

	kns, ok := os.LookupEnv("KUBESAW_NAMESPACE")
	if !ok {
		return fmt.Errorf("required Environment Variable KUBESAW_NAMESPACE not found")
	}
	log.Printf("KubeSaw namespace is: %s", kns)

	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	// setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setup read model
	log.Println("setting up cache")
	c, err := cache.New(cfg, wns, kns)
	if err != nil {
		return err
	}

	// setup write model
	// TODO

	// setup REST over HTTP server
	log.Println("setting up REST over HTTP server")
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
			log.Fatalf("error gracefully shutting down the HTTP server: %v", err)
		}
	}()

	// start the cache
	go func() {
		log.Println("Start cache")
		if err := c.Start(ctx); err != nil {
			if ctx.Err() == nil {
				cancel()
			}
			log.Printf("error starting cache: %s", err)
		}
	}()

	log.Println("Waiting for cache to sync...")
	if !c.WaitForCacheSync(ctx) {
		return fmt.Errorf("error synching cache")
	}

	// start HTTP server
	log.Printf("Start HTTP server at %s", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		return fmt.Errorf("error running server: %v", err)
	}
	return nil
}
