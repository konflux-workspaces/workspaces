package hook

import (
	"context"
	"fmt"
	"os"

	"github.com/cucumber/godog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
)

// Test Namespace
func createAndInjectTestNamespace(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	cli := tcontext.RetrieveHostClient(ctx)

	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Labels: map[string]string{
				"scope": "test",
			},
		},
	}
	if err := cli.Create(ctx, &ns); err != nil {
		panic(fmt.Sprintf("error creating test namespace %v: %v", ns.Name, err))
	}

	return tcontext.InjectTestNamespace(ctx, ns.Name), nil
}

func deleteTestNamespace(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	// skip if error is not nil
	if err != nil {
		return ctx, err
	}

	cli := tcontext.RetrieveHostClient(ctx)
	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
	}
	if err := cli.Delete(ctx, &ns); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// Workspace Namespace
func injectWorkspacesNamespace(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	ns, ok := os.LookupEnv("WORKSPACES_NAMESPACE")
	if !ok {
		ns := tcontext.RetrieveTestNamespace(ctx)
		return tcontext.InjectWorkspacesNamespace(ctx, ns), nil
	}

	return tcontext.InjectWorkspacesNamespace(ctx, ns), nil
}

// Kubespace Namespace
func injectKubespaceNamespace(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	ns, ok := os.LookupEnv("KUBESPACE_NAMESPACE")
	if !ok {
		ns := tcontext.RetrieveTestNamespace(ctx)
		return tcontext.InjectKubespaceNamespace(ctx, ns), nil
	}

	return tcontext.InjectKubespaceNamespace(ctx, ns), nil
}
