package internal

import (
	"os"

	"k8s.io/client-go/rest"
)

const envVarUseInsecure string = "E2E_USE_INSECURE_TLS"

func setTLSFromEnv(cfg *rest.Config) {
	if os.Getenv(envVarUseInsecure) == "true" {
		cfg.Insecure = true
		cfg.CAData = nil
		cfg.CAFile = ""
	}
}

type configMutationFunc func(*rest.Config)

var mutationFuncs = []configMutationFunc{
	setTLSFromEnv,
}

// MutateConfig adjusts a pre-populated rest.Config to adapt to the
// configuration we've been given from the environment.  For now, this only
// disables TLS validation if the environment variable `E2E_USE_INSECURE_TLS`
// is set.
func MutateConfig(cfg *rest.Config) {
	for _, f := range mutationFuncs {
		f(cfg)
	}
}
