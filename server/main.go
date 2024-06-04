package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
	"github.com/konflux-workspaces/workspaces/server/persistence/readclient"
	"github.com/konflux-workspaces/workspaces/server/persistence/writeclient"
	"github.com/konflux-workspaces/workspaces/server/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const DefaultAddr string = ":8080"
const EnvLogLevel = "LOG_LEVEL"

func main() {
	l := constructLog()
	if err := run(l); err != nil {
		l.Error("error configuring and running the server", "error", err)
		os.Exit(1)
	}
}

func run(l *slog.Logger) error {
	log.SetLogger(logr.FromSlogHandler(l.Handler()))

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
	l.Info("setting up cache")
	c, crc, err := readclient.NewDefaultWithCache(ctx, cfg, wns, kns)
	if err != nil {
		return err
	}

	// setup write model
	iwcli := iwclient.New(crc, wns, kns)
	writer := writeclient.NewWithConfig(cfg, wns, iwcli)

	// setup REST over HTTP server
	l.Info("setting up REST over HTTP server")
	s := rest.New(
		l,
		DefaultAddr,
		crc,
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
		l.Info("starting cache")
		if err := crc.Start(ctx); err != nil {
			if ctx.Err() == nil {
				cancel()
			}
			l.Error("error starting cache", "error", err)
		}
	}()

	l.Info("waiting for cache to sync...")
	if !crc.WaitForCacheSync(ctx) {
		return fmt.Errorf("error synching cache")
	}

	// start HTTP server
	l.Info("starting HTTP server", "address", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		return fmt.Errorf("error running server: %w", err)
	}

	return nil
}

// constructLog constructs a new instance of the logger
func constructLog() *slog.Logger {
	logLevel := getLogLevel()

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	return slog.New(handler)
}

// getLogLevel fetches the log level from the appropriate environment variable
func getLogLevel() slog.Level {
	env := os.Getenv(EnvLogLevel)
	level, err := strconv.Atoi(env)
	if err != nil {
		return slog.LevelError
	}
	return slog.Level(level)
}
