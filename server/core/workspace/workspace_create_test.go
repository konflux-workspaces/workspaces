package workspace_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"sigs.k8s.io/controller-runtime/pkg/client"

	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ = Describe("", func() {
	var (
		ctrl    *gomock.Controller
		ctx     context.Context
		creator *MockWorkspaceCreator
		request workspace.CreateWorkspaceCommand
		handler workspace.CreateWorkspaceHandler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.Background()
		creator = NewMockWorkspaceCreator(ctrl)
		request = workspace.CreateWorkspaceCommand{Workspace: restworkspacesv1alpha1.Workspace{}}
		handler = *workspace.NewCreateWorkspaceHandler(creator)
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
		opts := &client.CreateOptions{}
		creator.EXPECT().
			CreateUserWorkspace(ctx, username, &request.Workspace, opts).
			Return(nil)

		// when
		response, err := handler.Handle(ctx, request)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(response).To(Equal(&workspace.CreateWorkspaceResponse{
			Workspace: &request.Workspace,
		}))
	})

	It("should forward errors from the workspace creator", func() {
		// given
		username := "foo"
		ctx := context.WithValue(ctx, ccontext.UserSignupComplaintNameKey, username)
		opts := &client.CreateOptions{}
		error := fmt.Errorf("Failed to create workspace!")
		creator.EXPECT().
			CreateUserWorkspace(ctx, username, &request.Workspace, opts).
			Return(error)

		// when
		response, err := handler.Handle(ctx, request)

		// then
		Expect(response).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(error))
	})
})
