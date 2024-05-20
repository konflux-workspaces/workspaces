package hook

import (
	"os"

	"k8s.io/client-go/rest"
)

const envVarUseInsecure string = "E2E_USE_INSECURE_TLS"

func setTLSFromEnv(cfg *rest.Config) {
	cfg.Insecure = os.Getenv(envVarUseInsecure) != "false"
}

type configMutationFunc func(*rest.Config)

var mutationFuncs = []configMutationFunc{
	setTLSFromEnv,
}

func ProcessConfig(cfg *rest.Config) {
	for _, f := range mutationFuncs {
		f(cfg)
	}
}
