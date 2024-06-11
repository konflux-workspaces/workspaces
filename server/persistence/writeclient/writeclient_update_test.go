package writeclient_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/konflux-workspaces/workspaces/server/persistence/internal/cache"
	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
	"github.com/konflux-workspaces/workspaces/server/persistence/writeclient"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ = Describe("WriteclientUpdate", func() {
	var ctx context.Context
	var fakeClient client.WithWatch
	var scheme *runtime.Scheme
	var fakeClientBuilder *fake.ClientBuilder
	var cli *writeclient.WriteClient

	workspacesNamespace := "workspaces-system"
	kubesawNamespace := "toolchain-host"

	user := "foo"
	namespace := "bar"
	workspace := restworkspacesv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: user,
			Name:      "workspace-foo",
		},
		Spec: restworkspacesv1alpha1.WorkspaceSpec{},
		Status: restworkspacesv1alpha1.WorkspaceStatus{
			Space: &restworkspacesv1alpha1.SpaceInfo{
				Name: "space",
			},
		},
	}

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(toolchainv1alpha1.AddToScheme(scheme)).ToNot(HaveOccurred())
		Expect(restworkspacesv1alpha1.AddToScheme(scheme)).ToNot(HaveOccurred())
		Expect(workspacesv1alpha1.AddToScheme(scheme)).ToNot(HaveOccurred())
		fakeClientBuilder = fake.NewClientBuilder().WithScheme(scheme)
	})

	When("updating a non existing workspace", func() {
		BeforeEach(func() {
			fakeClient = fakeClientBuilder.Build()

			clientFunc := func(string) (client.Client, error) {
				return fakeClient, nil
			}
			iwcli := iwclient.New(fakeClient, namespace, namespace)
			cli = writeclient.New(clientFunc, namespace, iwcli)
		})

		It("should fail", func() {
			// when
			err := cli.UpdateUserWorkspace(ctx, user, &workspace)

			// then
			Expect(err).To(HaveOccurred())
		})
	})

	When("updating an existing workspace", func() {
		var internalWorkspace workspacesv1alpha1.InternalWorkspace
		BeforeEach(func() {
			internalWorkspace = workspacesv1alpha1.InternalWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workspace.Name + "-fddjk",
					Namespace: workspacesNamespace,
				},
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{
					Visibility:  workspacesv1alpha1.InternalWorkspaceVisibilityPrivate,
					DisplayName: workspace.Name,
					Owner: workspacesv1alpha1.UserInfo{
						JwtInfo: workspacesv1alpha1.JwtInfo{},
					},
				},
				Status: workspacesv1alpha1.InternalWorkspaceStatus{
					Space: workspacesv1alpha1.SpaceInfo{
						IsHome: true,
						Name:   "space",
					},
					Owner: workspacesv1alpha1.UserInfoStatus{
						Username: user,
					},
				},
			}
		})

		beforeInitializeCli := func(objs ...client.Object) {
			fcb := fakeClientBuilder.
				WithObjects(objs...)
			for key, indexer := range cache.UserSignupIndexers {
				fcb.WithIndex(&toolchainv1alpha1.UserSignup{}, key, indexer)
			}
			for key, indexer := range cache.InternalWorkspacesIndexers {
				fcb.WithIndex(&workspacesv1alpha1.InternalWorkspace{}, key, indexer)
			}
			fakeClient = fcb.Build()

			clientFunc := func(string) (client.Client, error) {
				return fakeClient, nil
			}
			iwcli := iwclient.New(fakeClient, workspacesNamespace, kubesawNamespace)
			cli = writeclient.New(clientFunc, namespace, iwcli)
		}

		When("updating a non-owned workspace", func() {
			BeforeEach(func() { beforeInitializeCli(&internalWorkspace, &workspace) })

			It("should fail with 404", func() {
				// given
				w := workspace.DeepCopy()
				w.Spec.Visibility = restworkspacesv1alpha1.WorkspaceVisibilityCommunity

				// when
				err := cli.UpdateUserWorkspace(ctx, user, w)

				// then
				Expect(err).To(HaveOccurred())
				Expect(kerrors.IsNotFound(err)).To(BeTrue())
			})
		})

		When("updating an owned workspace", func() {
			BeforeEach(func() {
				spaceBinding := toolchainv1alpha1.SpaceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      internalWorkspace.Name,
						Namespace: kubesawNamespace,
						Labels: map[string]string{
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            internalWorkspace.Name,
							toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: user,
						},
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						Space:            internalWorkspace.Name,
						SpaceRole:        "admin",
						MasterUserRecord: user,
					},
				}
				userSignup := toolchainv1alpha1.UserSignup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      user,
						Namespace: kubesawNamespace,
					},
					Status: toolchainv1alpha1.UserSignupStatus{
						CompliantUsername: workspace.Namespace,
					},
				}

				beforeInitializeCli(&internalWorkspace, &spaceBinding, &userSignup)
			})

			It("should update if the user is the owner", func() {
				// given
				w := workspace.DeepCopy()
				w.Spec.Visibility = restworkspacesv1alpha1.WorkspaceVisibilityCommunity

				// when
				err := cli.UpdateUserWorkspace(ctx, user, w)

				// then
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
