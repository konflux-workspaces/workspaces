package kubeconfig

import (
	"os"

	"k8s.io/client-go/rest"
)

const EnvVarUseInsecure string = "E2E_USE_INSECURE_TLS"

func BuildRESTConfigFromEnv() (*rest.Config, error) {
	overrides := []OverrideKubeconfigFunc{}
	if os.Getenv(EnvVarUseInsecure) == "true" {
		overrides = append(overrides, WithInsecureTLS)
	}

	return parseEnvOrHomeKubeconfig(overrides...)
}
