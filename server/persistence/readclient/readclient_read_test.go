package readclient_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
	"github.com/konflux-workspaces/workspaces/server/persistence/mapper"
	"github.com/konflux-workspaces/workspaces/server/persistence/readclient"
	"github.com/konflux-workspaces/workspaces/server/persistence/readclient/mocks"
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

	DescribeTable("Mapper returns an error", func(rerr error, expectedErrorFunc func(error) bool) {
		// given
		frc.EXPECT().
			GetAsUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		mp.EXPECT().
			InternalWorkspaceToWorkspace(gomock.Any()).
			Return(nil, rerr).
			Times(1)

		// when
		err := rc.ReadUserWorkspace(ctx, "", "", "", nil)

		// then
		Expect(err).To(HaveOccurred())
		Expect(expectedErrorFunc(err)).To(BeTrue())
	},
		Entry("display label not found -> internal error", mapper.ErrLabelDisplayNameNotFound, kerrors.IsInternalError),
		Entry("owner label not found -> internal error", mapper.ErrLabelOwnerNotFound, kerrors.IsInternalError),
	)
})
