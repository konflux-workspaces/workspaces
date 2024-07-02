package workspace_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/konflux-workspaces/workspaces/server/rest/workspace/mocks"

	coreworkspace "github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/rest/marshal"
	"github.com/konflux-workspaces/workspaces/server/rest/workspace"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ = Describe("Read", func() {
	var (
		ctrl    *gomock.Controller
		w       *restworkspacesv1alpha1.Workspace
		request *http.Request
		fake    *mocks.MockFakeResponseWriter
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		w = &restworkspacesv1alpha1.Workspace{}
		w.Name = "foo"
		w.Namespace = "bar"

		request = buildGetRequest(w)
		fake = mocks.NewMockFakeResponseWriter(ctrl)
	})

	AfterEach(func() { ctrl.Finish() })

	DescribeTable("workspace GET handler: read",
		func(
			mapperFunc workspace.ReadWorkspaceMapperFunc,
			readHandler workspace.ReadWorkspaceQueryHandlerFunc,
			marshaler marshal.MarshalerProvider,
			responseFunc func() http.ResponseWriter,
		) {
			response := responseFunc()
			handler := workspace.NewReadWorkspaceHandler(mapperFunc, readHandler, marshaler)
			handler.ServeHTTP(response, request)
		},
		Entry("failure in marshal provider", workspace.MapReadWorkspaceHttp, nopReadHandler, errorMarshalProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusBadRequest)
			return fake
		}),
		Entry("failure in read handler", workspace.MapReadWorkspaceHttp, badReadHandler, marshal.DefaultMarshalerProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusInternalServerError)
			return fake
		}),
		Entry("failure marshaling response", workspace.MapReadWorkspaceHttp, nopReadHandler, badMarshalProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusInternalServerError)
			return fake
		}),
		Entry("failure to write response", workspace.MapReadWorkspaceHttp, nopReadHandler, marshal.DefaultMarshalerProvider, func() http.ResponseWriter {
			fake.EXPECT().Header().Return(http.Header{})
			fake.EXPECT().Write(gomock.Any()).Return(0, fmt.Errorf("failed to write response body"))
			fake.EXPECT().WriteHeader(http.StatusInternalServerError)
			return fake
		}),
		Entry("failure to write response", workspace.MapReadWorkspaceHttp, nopReadHandler, marshal.DefaultMarshalerProvider, func() http.ResponseWriter {
			fake.EXPECT().Header().Return(http.Header{})
			fake.EXPECT().Write(gomock.Any()).DoAndReturn(func(a any) (int, error) {
				slice, ok := a.([]byte)
				Expect(ok).To(BeTrue())
				return len(slice), nil
			})
			return fake
		}),
	)
})

func badReadHandler(ctx context.Context, cmd coreworkspace.ReadWorkspaceQuery) (*coreworkspace.ReadWorkspaceResponse, error) {
	return nil, fmt.Errorf("bad create handler")
}

func nopReadHandler(_ctx context.Context, cmd coreworkspace.ReadWorkspaceQuery) (*coreworkspace.ReadWorkspaceResponse, error) {
	return &coreworkspace.ReadWorkspaceResponse{
		Workspace: nil,
	}, nil
}

func buildGetRequest(workspace *restworkspacesv1alpha1.Workspace) *http.Request {
	byteSlice, err := marshal.DefaultMarshal.Marshal(workspace)
	Expect(err).NotTo(HaveOccurred())

	url := fmt.Sprintf("/apis/workspaces.io/v1alpha1/namespaces/%s/workspaces", workspace.GetNamespace())

	request, err := http.NewRequest("POST", url, bytes.NewReader(byteSlice))
	Expect(err).NotTo(HaveOccurred())
	request.Header.Add("Content-Type", marshal.DefaultUnmarshal.ContentType())
	request.Header.Add("Accept", marshal.DefaultMarshal.ContentType())
	return request
}
