package workspace_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ = Describe("WorkspaceList", func() {
	var (
		ctrl    *gomock.Controller
		ctx     context.Context
		lister  *MockWorkspaceLister
		request workspace.ListWorkspaceQuery
		handler workspace.ListWorkspaceHandler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.Background()
		lister = NewMockWorkspaceLister(ctrl)
		request = workspace.ListWorkspaceQuery{}
		handler = *workspace.NewListWorkspaceHandler(lister)
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
		lister.EXPECT().
			ListUserWorkspaces(ctx, username, &restworkspacesv1alpha1.WorkspaceList{}, gomock.Any()).
			Return(nil)

		// when
		response, err := handler.Handle(ctx, request)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(response).To(Equal(&workspace.ListWorkspaceResponse{
			Workspaces: restworkspacesv1alpha1.WorkspaceList{
				TypeMeta: metav1.TypeMeta{},
			},
		}))
	})

	It("should forward errors from the workspace reader", func() {
		// given
		username := "foo"
		ctx := context.WithValue(ctx, ccontext.UserSignupComplaintNameKey, username)
		error := fmt.Errorf("Failed to create workspace!")
		lister.EXPECT().
			ListUserWorkspaces(ctx, username, &restworkspacesv1alpha1.WorkspaceList{}, gomock.Any()).
			Return(error)

		// when
		response, err := handler.Handle(ctx, request)

		// then
		Expect(response).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(error))
	})
})
