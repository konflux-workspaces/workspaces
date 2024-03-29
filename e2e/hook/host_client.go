package hook

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/e2e/pkg/cli"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesiov1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

func injectHostClient(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	p := func() string {
		e := os.Getenv("KUBECONFIG")
		if e != "" {
			return e
		}
		return filepath.Join(homedir.HomeDir(), ".kube", "config")
	}()

	cfg, err := clientcmd.BuildConfigFromFlags("", p)
	if err != nil {
		panic(fmt.Sprintf("error building config: %v", err))
	}

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(workspacesiov1alpha1.AddToScheme(scheme))
	utilruntime.Must(toolchainv1alpha1.AddToScheme(scheme))

	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		panic(fmt.Sprintf("error building client: %v", err))
	}

	tc := cli.New(c, sc.Id)
	return tcontext.InjectHostClient(ctx, tc), nil
}
