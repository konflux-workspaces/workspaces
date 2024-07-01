package workspace_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/konflux-workspaces/workspaces/server/core/workspace/mocks"

	"sigs.k8s.io/controller-runtime/pkg/client"

	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ = Describe("", func() {
	var (
		ctrl    *gomock.Controller
		ctx     context.Context
		updater *mocks.MockWorkspaceUpdater
		request workspace.UpdateWorkspaceCommand
		handler workspace.UpdateWorkspaceHandler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.Background()
		updater = mocks.NewMockWorkspaceUpdater(ctrl)
		request = workspace.UpdateWorkspaceCommand{Workspace: restworkspacesv1alpha1.Workspace{}}
		handler = *workspace.NewUpdateWorkspaceHandler(updater)
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
		opts := &client.UpdateOptions{}
		updater.EXPECT().
			UpdateUserWorkspace(ctx, username, &request.Workspace, opts).
			Return(nil)

		// when
		response, err := handler.Handle(ctx, request)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(response).To(Equal(&workspace.UpdateWorkspaceResponse{
			Workspace: &request.Workspace,
		}))
	})

	It("should forward errors from the workspace creator", func() {
		// given
		username := "foo"
		ctx := context.WithValue(ctx, ccontext.UserSignupComplaintNameKey, username)
		opts := &client.UpdateOptions{}
		error := fmt.Errorf("Failed to create workspace!")
		updater.EXPECT().
			UpdateUserWorkspace(ctx, username, &request.Workspace, opts).
			Return(error)

		// when
		response, err := handler.Handle(ctx, request)

		// then
		Expect(response).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(error))
	})
})
