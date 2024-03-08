package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/persistence/cache"
	"github.com/konflux-workspaces/workspaces/server/persistence/kube"
	"github.com/konflux-workspaces/workspaces/server/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const DefaultAddr string = ":8080"

func main() {
	l := slog.Default()

	if err := run(); err != nil {
		l.Error("error configuring and running the server", "error", err)
		os.Exit(1)
	}
}

func run() error {
	l := slog.Default()

	// fetch configuration
	wns, ok := os.LookupEnv("WORKSPACES_NAMESPACE")
	if !ok {
		return fmt.Errorf("required Environment Variable WORKSPACES_NAMESPACE not found")
	}
	l.Debug("retrieving configuration from env variables", "workspaces namespace", wns)

	kns, ok := os.LookupEnv("KUBESAW_NAMESPACE")
	if !ok {
		return fmt.Errorf("required Environment Variable KUBESAW_NAMESPACE not found")
	}
	l.Debug("retrieving configuration from env variables", "kubesaw namespace", kns)

	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	// setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setup read model
	l.Debug("setting up cache")
	c, err := cache.New(ctx, cfg, wns, kns)
	if err != nil {
		return err
	}

	// setup write model
	writer := kube.New(kube.BuildClient(cfg), wns)

	// setup REST over HTTP server
	l.Debug("setting up REST over HTTP server")
	s := rest.New(
		l,
		DefaultAddr,
		workspace.NewReadWorkspaceHandler(c).Handle,
		workspace.NewListWorkspaceHandler(c).Handle,
		workspace.NewCreateWorkspaceHandler(writer).Handle,
		workspace.NewUpdateWorkspaceHandler(writer).Handle,
	)

	// HTTP Server graceful shutdown
	go func() {
		<-ctx.Done()

		sctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := s.Shutdown(sctx); err != nil {
			l.Error("error gracefully shutting down the HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	// start the cache
	go func() {
		l.Debug("starting cache")
		if err := c.Start(ctx); err != nil {
			if ctx.Err() == nil {
				cancel()
			}
			l.Error("error starting cache", "error", err)
		}
	}()

	l.Debug("waiting for cache to sync...")
	if !c.WaitForCacheSync(ctx) {
		return fmt.Errorf("error synching cache")
	}

	// start HTTP server
	l.Debug("starting HTTP server", "address", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		return fmt.Errorf("error running server: %w", err)
	}

	return nil
}
