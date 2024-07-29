package workspace_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

var _ = Describe("Update tests", func() {
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

		request = buildPutRequest(w)
		fake = mocks.NewMockFakeResponseWriter(ctrl)
	})

	AfterEach(func() { ctrl.Finish() })

	DescribeTable("workspace PUT handler",
		func(
			mapperFunc workspace.UpdateWorkspaceMapperFunc,
			updateHandler workspace.UpdateWorkspaceCommandHandlerFunc,
			marshaler marshal.MarshalerProvider,
			unmarshaler marshal.UnmarshalerProvider,
			responseFunc func() http.ResponseWriter,
		) {
			response := responseFunc()
			handler := workspace.NewUpdateWorkspaceHandler(mapperFunc, updateHandler, marshaler, unmarshaler)
			handler.ServeHTTP(response, request)
		},
		Entry("failure in marshal provider", workspace.MapPutWorkspaceHttp, nopUpdateHandler, errorMarshalProvider, marshal.DefaultUnmarshalerProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusBadRequest)
			return fake
		}),
		Entry("failure in unmarshal provider", workspace.MapPutWorkspaceHttp, nopUpdateHandler, marshal.DefaultMarshalerProvider, errorUnmarshalProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusBadRequest)
			return fake
		}),
		Entry("no body sent in request", workspace.MapPutWorkspaceHttp, nopUpdateHandler, marshal.DefaultMarshalerProvider, marshal.DefaultUnmarshalerProvider, func() http.ResponseWriter {
			request.Body = io.NopCloser(bytes.NewReader([]byte{}))
			fake.EXPECT().WriteHeader(http.StatusBadRequest)
			return fake
		}),
		Entry("failure unmarshaling request", workspace.MapPutWorkspaceHttp, nopUpdateHandler, marshal.DefaultMarshalerProvider, badUnmarshalProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusBadRequest)
			return fake
		}),
		Entry("failure in update handler", workspace.MapPutWorkspaceHttp, badUpdateHandler, marshal.DefaultMarshalerProvider, marshal.DefaultUnmarshalerProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusInternalServerError)
			return fake
		}),
		Entry("failure marshaling response", workspace.MapPutWorkspaceHttp, nopUpdateHandler, badMarshalProvider, marshal.DefaultUnmarshalerProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusInternalServerError)
			return fake
		}),
		Entry("failure to write response", workspace.MapPutWorkspaceHttp, nopUpdateHandler, marshal.DefaultMarshalerProvider, marshal.DefaultUnmarshalerProvider, func() http.ResponseWriter {
			fake.EXPECT().Header().Return(http.Header{})
			fake.EXPECT().Write(gomock.Any()).Return(0, fmt.Errorf("failed to write response body"))
			fake.EXPECT().WriteHeader(http.StatusInternalServerError)
			return fake
		}),
		Entry("failure to write response", workspace.MapPutWorkspaceHttp, nopUpdateHandler, marshal.DefaultMarshalerProvider, marshal.DefaultUnmarshalerProvider, func() http.ResponseWriter {
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

func badUpdateHandler(ctx context.Context, cmd coreworkspace.UpdateWorkspaceCommand) (*coreworkspace.UpdateWorkspaceResponse, error) {
	return nil, fmt.Errorf("bad update handler")
}

func nopUpdateHandler(_ctx context.Context, cmd coreworkspace.UpdateWorkspaceCommand) (*coreworkspace.UpdateWorkspaceResponse, error) {
	return &coreworkspace.UpdateWorkspaceResponse{
		Workspace: &cmd.Workspace,
	}, nil
}

func buildPutRequest(workspace *restworkspacesv1alpha1.Workspace) *http.Request {
	byteSlice, err := marshal.DefaultMarshal.Marshal(workspace)
	Expect(err).NotTo(HaveOccurred())

	url := fmt.Sprintf("/apis/workspaces.io/v1alpha1/namespaces/%s/workspaces", workspace.GetNamespace())

	request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(byteSlice))
	Expect(err).NotTo(HaveOccurred())
	request.Header.Add("Content-Type", marshal.DefaultUnmarshal.ContentType())
	request.Header.Add("Accept", marshal.DefaultMarshal.ContentType())
	return request
}
