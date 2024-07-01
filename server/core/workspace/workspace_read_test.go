package workspace_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"sigs.k8s.io/controller-runtime/pkg/client"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/core/workspace/mocks"
)

var _ = Describe("WorkspaceRead", func() {
	var (
		ctrl    *gomock.Controller
		ctx     context.Context
		reader  *mocks.MockWorkspaceReader
		request workspace.ReadWorkspaceQuery
		handler workspace.ReadWorkspaceHandler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.Background()
		reader = mocks.NewMockWorkspaceReader(ctrl)
		request = workspace.ReadWorkspaceQuery{}
		handler = *workspace.NewReadWorkspaceHandler(reader)
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
		username := "foo"
		ctx := context.WithValue(ctx, ccontext.UserSignupComplaintNameKey, username)
		reader.EXPECT().
			ReadUserWorkspace(ctx, username, request.Owner, request.Name, &restworkspacesv1alpha1.Workspace{}, []client.GetOption{}).
			Return(nil)

		// when
		response, err := handler.Handle(ctx, request)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(response).To(Equal(&workspace.ReadWorkspaceResponse{
			Workspace: &restworkspacesv1alpha1.Workspace{},
		}))
	})

	It("should forward errors from the workspace reader", func() {
		// given
		username := "foo"
		ctx := context.WithValue(ctx, ccontext.UserSignupComplaintNameKey, username)
		error := fmt.Errorf("Failed to create workspace!")
		reader.EXPECT().
			ReadUserWorkspace(ctx, username, request.Owner, request.Name, &restworkspacesv1alpha1.Workspace{}, []client.GetOption{}).
			Return(error)

		// when
		response, err := handler.Handle(ctx, request)

		// then
		Expect(response).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(error))
	})
})
