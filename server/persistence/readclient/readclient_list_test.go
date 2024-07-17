package readclient_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
	"github.com/konflux-workspaces/workspaces/server/persistence/readclient"
	"github.com/konflux-workspaces/workspaces/server/persistence/readclient/mocks"
)

var _ = Describe("List", func() {
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

	AfterEach(func() {
		ctrl.Finish()
	})

	DescribeTable("Filter by label", func(unfilteredInternalWorkspaces []workspacesv1alpha1.InternalWorkspace, expectedObjectMetas []metav1.ObjectMeta) {
		// given
		user := "user"
		// internalLabel := workspacesv1alpha1.LabelInternalDomain + "whatever"

		// internal client expected to be called once.
		// It returns no error so we can test the filtering by label.
		frc.EXPECT().
			ListAsUser(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, iww *workspacesv1alpha1.InternalWorkspaceList) error {
				iww.Items = unfilteredInternalWorkspaces
				return nil
			}).
			Times(1)

			// mapper expects to be called once.
			// It returns the list of mapped workspaces (just objectmeta)
		mp.EXPECT().
			InternalWorkspaceListToWorkspaceList(gomock.Any()).
			Times(1).
			DoAndReturn(func(iww *workspacesv1alpha1.InternalWorkspaceList) (*restworkspacesv1alpha1.WorkspaceList, error) {
				ww := restworkspacesv1alpha1.WorkspaceList{Items: []restworkspacesv1alpha1.Workspace{}}
				for _, w := range iww.Items {
					ww.Items = append(ww.Items, restworkspacesv1alpha1.Workspace{ObjectMeta: w.ObjectMeta})
				}
				return &ww, nil
			})

		// when
		actualWorkspaces := restworkspacesv1alpha1.WorkspaceList{}
		err := rc.ListUserWorkspaces(ctx, user, &actualWorkspaces, client.MatchingLabels{"hit": "hit"})

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(actualWorkspaces.Items).To(
			And(
				HaveLen(len(expectedObjectMetas)),
				WithTransform(func(ww []restworkspacesv1alpha1.Workspace) []metav1.ObjectMeta {
					oww := make([]metav1.ObjectMeta, len(ww))
					for i, w := range actualWorkspaces.Items {
						oww[i] = w.ObjectMeta
					}
					return oww
				}, BeEquivalentTo(expectedObjectMetas))))
	},
		Entry("unfiltered InternalWorkspaces are nil returns empty list", nil, []metav1.ObjectMeta{}),
		Entry("unfiltered InternalWorkspaces are empty returns empty list", []workspacesv1alpha1.InternalWorkspace{}, []metav1.ObjectMeta{}),
		Entry("no matching InternalWorkspaces returns empty list", []workspacesv1alpha1.InternalWorkspace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "miss",
					Namespace: "miss",
					Labels: map[string]string{
						"miss": "miss",
					},
				},
			},
		}, []metav1.ObjectMeta{}),
		Entry("single matching InternalWorkspaces returns one workspace", []workspacesv1alpha1.InternalWorkspace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hit",
					Namespace: "hit",
					Labels: map[string]string{
						"hit": "hit",
					},
				},
			},
		}, []metav1.ObjectMeta{
			{
				Name:      "hit",
				Namespace: "hit",
				Labels: map[string]string{
					"hit": "hit",
				},
			},
		}),
		Entry("multiple matching InternalWorkspaces returns multiple workspaces", []workspacesv1alpha1.InternalWorkspace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "miss",
					Namespace: "miss",
					Labels: map[string]string{
						"miss": "miss",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hit-1",
					Namespace: "hit-1",
					Labels: map[string]string{
						"hit": "hit",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hit-2",
					Namespace: "hit-2",
					Labels: map[string]string{
						"hit": "hit",
					},
				},
			},
		}, []metav1.ObjectMeta{
			{
				Name:      "hit-1",
				Namespace: "hit-1",
				Labels: map[string]string{
					"hit": "hit",
				},
			},
			{
				Name:      "hit-2",
				Namespace: "hit-2",
				Labels: map[string]string{
					"hit": "hit",
				},
			},
		}),
	)

	DescribeTable("InternalClient returns an error", func(rerr error, expectedErrorFunc func(error) bool) {
		// given
		frc.EXPECT().
			ListAsUser(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(rerr).
			Times(1)

		// when
		err := rc.ListUserWorkspaces(ctx, "", nil)

		// then
		Expect(err).To(HaveOccurred())
		Expect(expectedErrorFunc(err)).To(BeTrue())
	},
		Entry("unauthorized -> internal error", iwclient.ErrUnauthorized, kerrors.IsInternalError),
	)

	It("handles mapper error and returns InternalError", func() {
		// given
		merr := fmt.Errorf("mapper error")
		user := "user"

		// internal client expected to be called once.
		// It returns no error so we can test the filtering by label.
		frc.EXPECT().
			ListAsUser(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

			// mapper expects to be called once.
			// It returns an error so we can test error handling.
		mp.EXPECT().
			InternalWorkspaceListToWorkspaceList(gomock.Any()).
			Return(nil, merr).
			Times(1)

		// when
		err := rc.ListUserWorkspaces(ctx, user, nil)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(kerrors.IsInternalError, "IsInternalError"))
	})

	Describe("ListOptions are mapped", func() {
		It("returns an error if labels with reserved domain are used", func() {
			// given
			user := "user"
			internalLabel := workspacesv1alpha1.LabelInternalDomain + "whatever"

			// internal client expected to be called once.
			// It returns no error so we can test the filtering by label.
			frc.EXPECT().
				ListAsUser(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)

			// when
			actualWorkspaces := restworkspacesv1alpha1.WorkspaceList{}
			err := rc.ListUserWorkspaces(ctx, user, &actualWorkspaces, client.MatchingLabels{internalLabel: "whatever"})

			// then
			Expect(err).To(MatchError(fmt.Errorf("invalid label selector: key '%s' is reserved", internalLabel)))
		})
	})
})
