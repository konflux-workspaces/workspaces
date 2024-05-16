package hook

import (
	"context"
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/konflux-workspaces/workspaces/e2e/pkg/dump"
)

func dumpResources(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	// skip if no error
	if err == nil {
		return ctx, nil
	}

	// dump resources
	if derr := dump.DumpAll(ctx); derr != nil {
		return ctx, errors.Join(err, fmt.Errorf("error dumping resources: %w", derr))
	}
	return ctx, err
}
