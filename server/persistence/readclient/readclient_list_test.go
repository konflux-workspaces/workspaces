package readclient_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

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
})
