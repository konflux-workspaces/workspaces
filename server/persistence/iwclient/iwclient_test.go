package iwclient_test

import (
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/persistence/internal/cache"
	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
)

func buildCache(wsns, ksns string, objs ...client.Object) *iwclient.Client {
	var err error
	scheme := runtime.NewScheme()
	err = workspacesv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = toolchainv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	fcb := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...)

	for key, indexer := range cache.UserSignupIndexers {
		fcb.WithIndex(&toolchainv1alpha1.UserSignup{}, key, indexer)
	}
	for key, indexer := range cache.InternalWorkspacesIndexers {
		fcb.WithIndex(&workspacesv1alpha1.InternalWorkspace{}, key, indexer)
	}

	return iwclient.New(fcb.Build(), wsns, ksns)
}

func generateName(namePrefix string) string {
	return namePrefix + "-jjdjk"
}
