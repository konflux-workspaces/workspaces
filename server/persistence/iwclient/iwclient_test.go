package iwclient_test

import (
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
)

func buildCache(wsns, ksns string, objs ...client.Object) *iwclient.Client {
	var err error
	scheme := runtime.NewScheme()
	err = workspacesv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = toolchainv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	fc := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	return iwclient.New(fc, wsns, ksns)
}

func generateName(namePrefix string) string {
	return namePrefix + "-jjdjk"
}
