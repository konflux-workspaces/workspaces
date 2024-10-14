package workspace_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"

	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ = Describe("", func() {
	var (
		ctrl    *gomock.Controller
		ctx     context.Context
		reader  *MockWorkspaceReader
		updater *MockWorkspaceUpdater
		request workspace.PatchWorkspaceCommand
		handler workspace.PatchWorkspaceHandler
		w       workspacesv1alpha1.Workspace
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.Background()
		w = workspacesv1alpha1.Workspace{
			ObjectMeta: v1.ObjectMeta{
				Name:      "default",
				Namespace: "user",
			},
			Spec: workspacesv1alpha1.WorkspaceSpec{
				Visibility: workspacesv1alpha1.WorkspaceVisibilityPrivate,
			},
		}
		updater = NewMockWorkspaceUpdater(ctrl)
		reader = NewMockWorkspaceReader(ctrl)
		request = workspace.PatchWorkspaceCommand{
			Workspace: w.Name,
			Owner:     w.Namespace,
		}
		handler = *workspace.NewPatchWorkspaceHandler(reader, updater)
	})

	AfterEach(func() { ctrl.Finish() })

	It("should not allow unauthenticated requests", func() {
		// don't set the "user" value within ctx

		response, err := handler.Handle(ctx, request)
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(fmt.Errorf("unauthenticated request")))
		Expect(response).To(BeNil())
	})

	It("should allow authenticated requests", func() {
		// given
		request.PatchType = types.MergePatchType
		request.Patch = []byte(`{"spec":{"visibility":"community"}}`)
		username := "foo"
		ctx := context.WithValue(ctx, ccontext.UserSignupComplaintNameKey, username)
		opts := &client.UpdateOptions{}
		reader.EXPECT().
			ReadUserWorkspace(ctx, username, w.Namespace, w.Name, gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, user, owner, workspace string, rw *workspacesv1alpha1.Workspace, opts ...client.GetOption) error {
				w.DeepCopyInto(rw)
				return nil
			})
		updater.EXPECT().
			UpdateUserWorkspace(ctx, username, gomock.Any(), opts).
			Return(nil)

		// when
		response, err := handler.Handle(ctx, request)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(response).NotTo(BeNil())
		expectedWorkspace := w.DeepCopy()
		expectedWorkspace.Spec.Visibility = workspacesv1alpha1.WorkspaceVisibilityCommunity
		Expect(response.Workspace).To(BeEquivalentTo(expectedWorkspace))
	})

	DescribeTable("Unsupported patch types are rejected",
		func(patchType types.PatchType) {
			// given
			username := "foo"
			ctx := context.WithValue(ctx, ccontext.UserSignupComplaintNameKey, username)
			request.PatchType = patchType
			reader.EXPECT().
				ReadUserWorkspace(ctx, username, w.Namespace, w.Name, gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, user, owner, workspace string, rw *workspacesv1alpha1.Workspace, opts ...client.GetOption) error {
					w.DeepCopyInto(rw)
					return nil
				})
			updater.EXPECT().
				UpdateUserWorkspace(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(0)

			// when
			response, err := handler.Handle(ctx, request)

			// then
			Expect(response).To(BeNil())
			Expect(err).To(MatchError(fmt.Errorf("unsupported patch type: %s", patchType)))
		},
		Entry("empty patchType", types.PatchType("")),
		Entry("invalid patchType", types.PatchType("bar")),
		Entry("strategic merge", types.StrategicMergePatchType),
		Entry("apply patch", types.ApplyPatchType),
	)
})
