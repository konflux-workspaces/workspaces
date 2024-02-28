package hook

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func injectUnauthKubeconfig(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
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

	c := &rest.Config{
		ContentConfig: cfg.ContentConfig,
		Host:          cfg.Host,
		APIPath:       cfg.APIPath,
		TLSClientConfig: rest.TLSClientConfig{
			CAFile:     cfg.CAFile,
			CAData:     cfg.CAData,
			ServerName: cfg.ServerName,
			Insecure:   cfg.Insecure,
		},
		Timeout: cfg.Timeout,
	}

	return tcontext.InjectUnauthKubeconfig(ctx, c), nil
}
