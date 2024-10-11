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
	"k8s.io/apimachinery/pkg/types"

	"github.com/konflux-workspaces/workspaces/server/rest/workspace/mocks"

	coreworkspace "github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/rest/marshal"
	"github.com/konflux-workspaces/workspaces/server/rest/workspace"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ = Describe("Patch tests", func() {
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

		request = buildPatchRequest(w)
		fake = mocks.NewMockFakeResponseWriter(ctrl)
	})

	AfterEach(func() { ctrl.Finish() })

	DescribeTable("workspace PATCH handler",
		func(
			mapperFunc workspace.PatchWorkspaceMapperFunc,
			patchHandler workspace.PatchWorkspaceCommandHandlerFunc,
			marshaler marshal.MarshalerProvider,
			prepare func() http.ResponseWriter,
		) {
			response := prepare()
			handler := workspace.NewPatchWorkspaceHandler(mapperFunc, patchHandler, marshaler)
			handler.ServeHTTP(response, request)
		},
		Entry("failure in marshal provider", workspace.MapPatchWorkspaceHttp, nopPatchHandler, errorMarshalProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusBadRequest)
			return fake
		}),
		Entry("no Content-Type in request", workspace.MapPatchWorkspaceHttp, nopPatchHandler, marshal.DefaultMarshalerProvider, func() http.ResponseWriter {
			request.Body = io.NopCloser(bytes.NewReader([]byte{}))
			request.Header.Del("Content-Type")
			fake.EXPECT().WriteHeader(http.StatusBadRequest).Times(1)
			return fake
		}),
		Entry("failure in patch handler", workspace.MapPatchWorkspaceHttp, badPatchHandler, marshal.DefaultMarshalerProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusInternalServerError)
			return fake
		}),
		Entry("failure marshaling response", workspace.MapPatchWorkspaceHttp, nopPatchHandler, badMarshalProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusInternalServerError)
			return fake
		}),
		Entry("failure to write response", workspace.MapPatchWorkspaceHttp, nopPatchHandler, marshal.DefaultMarshalerProvider, func() http.ResponseWriter {
			fake.EXPECT().Header().Return(http.Header{})
			fake.EXPECT().Write(gomock.Any()).Return(0, fmt.Errorf("failed to write response body"))
			fake.EXPECT().WriteHeader(http.StatusInternalServerError)
			return fake
		}),
		Entry("failure to write response", workspace.MapPatchWorkspaceHttp, nopPatchHandler, marshal.DefaultMarshalerProvider, func() http.ResponseWriter {
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

func badPatchHandler(ctx context.Context, cmd coreworkspace.PatchWorkspaceCommand) (*coreworkspace.PatchWorkspaceResponse, error) {
	return nil, fmt.Errorf("bad patch handler")
}

func nopPatchHandler(_ctx context.Context, cmd coreworkspace.PatchWorkspaceCommand) (*coreworkspace.PatchWorkspaceResponse, error) {
	return &coreworkspace.PatchWorkspaceResponse{}, nil
}

func buildPatchRequest(workspace *restworkspacesv1alpha1.Workspace) *http.Request {
	byteSlice, err := marshal.DefaultMarshal.Marshal(workspace)
	Expect(err).NotTo(HaveOccurred())

	url := fmt.Sprintf("/apis/workspaces.io/v1alpha1/namespaces/%s/workspaces", workspace.GetNamespace())

	request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(byteSlice))
	Expect(err).NotTo(HaveOccurred())
	request.Header.Add("Content-Type", string(types.MergePatchType))
	request.Header.Add("Accept", marshal.DefaultMarshal.ContentType())
	return request
}
