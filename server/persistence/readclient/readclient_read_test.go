package readclient_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/konflux-workspaces/workspaces/server/persistence/readclient/mocks"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
	"github.com/konflux-workspaces/workspaces/server/persistence/readclient"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ = Describe("Read", func() {
	var ctx context.Context
	var ctrl *gomock.Controller
	var frc *mocks.MockFakeIWReadClient
	var mp *mocks.MockFakeIWMapper
	var rc *readclient.ReadClient

	BeforeEach(func() {
		ctx = context.Background()
		ctrl = gomock.NewController(GinkgoT())
		frc = mocks.NewMockFakeIWReadClient(ctrl)
		mp = mocks.NewMockFakeIWMapper(ctrl)
		rc = readclient.New(frc, mp)
	})

	Describe("valid request", func() {
		// happy path
		It("returns a copy of the mapped value", func() {
			// given

			// internal client expected to be called once.
			// It returns no error so we can test the mapper invocation.
			frc.EXPECT().
				GetAsUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)

			// mapper expects to be called once.
			// It returns a valid workspace so we can test handler's result.
			mappedWorkspace := restworkspacesv1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "workspace",
					Namespace: "owner",
				},
				Spec: restworkspacesv1alpha1.WorkspaceSpec{
					Visibility: restworkspacesv1alpha1.WorkspaceVisibilityCommunity,
				},
				Status: restworkspacesv1alpha1.WorkspaceStatus{
					Space: &restworkspacesv1alpha1.SpaceInfo{
						Name: "space",
					},
				},
			}
			mp.EXPECT().
				InternalWorkspaceToWorkspace(gomock.Any()).
				Return(&mappedWorkspace, nil).
				Times(1)

			// when
			returnedWorkspace := restworkspacesv1alpha1.Workspace{}
			err := rc.ReadUserWorkspace(ctx, "", "", "", &returnedWorkspace)

			// then
			Expect(err).NotTo(HaveOccurred())
			Expect(returnedWorkspace).To(Equal(mappedWorkspace))
			// test data is deepcopied
			Expect(returnedWorkspace).NotTo(BeIdenticalTo(mappedWorkspace))
		})

		DescribeTable("should set the is-owner label on owned workspaces", func(owner, is_owned string) {
			// internal client expected to be called once.
			// It returns no error so we can test the mapper invocation.
			frc.EXPECT().
				GetAsUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)

			// mapper expects to be called once.
			// It returns a valid workspace so we can test handler's result.
			mappedWorkspace := restworkspacesv1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "workspace",
					Namespace: "owner",
				},
				Spec: restworkspacesv1alpha1.WorkspaceSpec{
					Visibility: restworkspacesv1alpha1.WorkspaceVisibilityCommunity,
				},
				Status: restworkspacesv1alpha1.WorkspaceStatus{
					Space: &restworkspacesv1alpha1.SpaceInfo{
						Name: "space",
					},
				},
			}
			mp.EXPECT().
				InternalWorkspaceToWorkspace(gomock.Any()).
				Return(&mappedWorkspace, nil).
				Times(1)

			// when
			returnedWorkspace := restworkspacesv1alpha1.Workspace{}
			err := rc.ReadUserWorkspace(ctx, owner, "", "", &returnedWorkspace)

			// then
			Expect(err).NotTo(HaveOccurred())
			Expect(returnedWorkspace.Labels).To(HaveKeyWithValue(restworkspacesv1alpha1.LabelIsOwner, is_owned))
		},
			Entry("non-owner", "another", "false"),
			Entry("owner", "owner", "true"),
		)
	})

	// error handling
	DescribeTable("InternalClient returns an error", func(rerr error, expectedErrorFunc func(error) bool) {
		// given
		frc.EXPECT().
			GetAsUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(rerr).
			Times(1)

		// when
		err := rc.ReadUserWorkspace(ctx, "", "", "", nil)

		// then
		Expect(err).To(HaveOccurred())
		Expect(expectedErrorFunc(err)).To(BeTrue())
	},
		Entry("not found -> not found", iwclient.ErrWorkspaceNotFound, kerrors.IsNotFound),
		Entry("unauthorized -> not found", iwclient.ErrUnauthorized, kerrors.IsNotFound),
		// TODO: should we use here a different error? like InternalServerError?
		Entry("more than one found -> not found", iwclient.ErrMoreThanOneFound, kerrors.IsNotFound),
	)

	It("handles mapper error and returns InternalError", func() {
		// given
		// internal client expected to be called once.
		// It returns no error so we can test the mapper invocation and error handling.
		frc.EXPECT().
			GetAsUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		// mapper expects to be called once.
		// It returns an error so we can test error handling.
		mp.EXPECT().
			InternalWorkspaceToWorkspace(gomock.Any()).
			Return(nil, fmt.Errorf("mapper error")).
			Times(1)

		// when
		err := rc.ReadUserWorkspace(ctx, "", "", "", nil)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(kerrors.IsInternalError, "IsInternalError"))
	})
})
