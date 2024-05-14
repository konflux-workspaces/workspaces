package kubeconfig

import (
	"cmp"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type OverrideKubeconfigFunc func(*rest.Config)

func WithInsecureTLS(cfg *rest.Config) {
	cfg.TLSClientConfig = rest.TLSClientConfig{Insecure: true}
}

func parseEnvOrHomeKubeconfig(overrides ...OverrideKubeconfigFunc) (*rest.Config, error) {
	p := getKubeconfigFromEnvOrHome()

	cfg, err := clientcmd.BuildConfigFromFlags("", p)
	if err != nil {
		return nil, err
	}

	for _, f := range overrides {
		f(cfg)
	}

	return cfg, nil
}

func getKubeconfigFromEnvOrHome() string {
	return cmp.Or(
		os.Getenv("KUBECONFIG"),
		filepath.Join(homedir.HomeDir(), ".kube", "config"),
	)
}
