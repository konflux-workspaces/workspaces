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

var _ = Describe("Creation tests", func() {
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

		request = buildPostRequest(w)
		fake = mocks.NewMockFakeResponseWriter(ctrl)
	})

	AfterEach(func() { ctrl.Finish() })

	DescribeTable("workspace POST handler",
		func(
			mapperFunc workspace.PostWorkspaceMapperFunc,
			createHandler workspace.CreateWorkspaceCommandHandlerFunc,
			marshaler marshal.MarshalerProvider,
			unmarshaler marshal.UnmarshalerProvider,
			responseFunc func() http.ResponseWriter,
		) {
			response := responseFunc()
			handler := workspace.NewPostWorkspaceHandler(mapperFunc, createHandler, marshaler, unmarshaler)
			handler.ServeHTTP(response, request)
		},
		Entry("failure in marshal provider", workspace.MapPostWorkspaceHttp, nopCreateHandler, errorMarshalProvider, marshal.DefaultUnmarshalerProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusBadRequest)
			return fake
		}),
		Entry("failure in unmarshal provider", workspace.MapPostWorkspaceHttp, nopCreateHandler, marshal.DefaultMarshalerProvider, errorUnmarshalProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusBadRequest)
			return fake
		}),
		Entry("no body sent in request", workspace.MapPostWorkspaceHttp, nopCreateHandler, marshal.DefaultMarshalerProvider, marshal.DefaultUnmarshalerProvider, func() http.ResponseWriter {
			request.Body = io.NopCloser(bytes.NewReader([]byte{}))
			fake.EXPECT().WriteHeader(http.StatusBadRequest)
			return fake
		}),
		Entry("failure unmarshaling request", workspace.MapPostWorkspaceHttp, nopCreateHandler, marshal.DefaultMarshalerProvider, badUnmarshalProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusBadRequest)
			return fake
		}),
		Entry("failure in create handler", workspace.MapPostWorkspaceHttp, badCreateHandler, marshal.DefaultMarshalerProvider, marshal.DefaultUnmarshalerProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusInternalServerError)
			return fake
		}),
		Entry("failure marshaling response", workspace.MapPostWorkspaceHttp, nopCreateHandler, badMarshalProvider, marshal.DefaultUnmarshalerProvider, func() http.ResponseWriter {
			fake.EXPECT().WriteHeader(http.StatusInternalServerError)
			return fake
		}),
		Entry("failure to write response", workspace.MapPostWorkspaceHttp, nopCreateHandler, marshal.DefaultMarshalerProvider, marshal.DefaultUnmarshalerProvider, func() http.ResponseWriter {
			fake.EXPECT().Header().Return(http.Header{})
			fake.EXPECT().Write(gomock.Any()).Return(0, fmt.Errorf("failed to write response body"))
			fake.EXPECT().WriteHeader(http.StatusInternalServerError)
			return fake
		}),
		Entry("failure to write response", workspace.MapPostWorkspaceHttp, nopCreateHandler, marshal.DefaultMarshalerProvider, marshal.DefaultUnmarshalerProvider, func() http.ResponseWriter {
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

func errorMarshalProvider(*http.Request) (marshal.Marshaler, error) {
	return nil, fmt.Errorf("bad marshaler provider")
}

func errorUnmarshalProvider(*http.Request) (marshal.Unmarshaler, error) {
	return nil, fmt.Errorf("bad unmarshaler provider")
}

type badMarshaler struct{}

var _ marshal.Marshaler = &badMarshaler{}

func (b *badMarshaler) ContentType() string {
	return "application/json"
}

func (b *badMarshaler) Marshal(any) ([]byte, error) {
	return nil, fmt.Errorf("unable to marshal input!")
}

func badMarshalProvider(*http.Request) (marshal.Marshaler, error) {
	return &badMarshaler{}, nil
}

type badUnmarshaler struct{}

func (b *badUnmarshaler) ContentType() string {
	return "application/json"
}

func (b *badUnmarshaler) Unmarshal([]byte, any) error {
	return fmt.Errorf("unable to unmarshal input!")
}
func badUnmarshalProvider(*http.Request) (marshal.Unmarshaler, error) {
	return &badUnmarshaler{}, nil
}

func badCreateHandler(ctx context.Context, cmd coreworkspace.CreateWorkspaceCommand) (*coreworkspace.CreateWorkspaceResponse, error) {
	return nil, fmt.Errorf("bad create handler")
}

func nopCreateHandler(_ctx context.Context, cmd coreworkspace.CreateWorkspaceCommand) (*coreworkspace.CreateWorkspaceResponse, error) {
	return &coreworkspace.CreateWorkspaceResponse{
		Workspace: &cmd.Workspace,
	}, nil
}

func buildPostRequest(workspace *restworkspacesv1alpha1.Workspace) *http.Request {
	byteSlice, err := marshal.DefaultMarshal.Marshal(workspace)
	Expect(err).NotTo(HaveOccurred())

	url := fmt.Sprintf("/apis/workspaces.io/v1alpha1/namespaces/%s/workspaces", workspace.GetNamespace())

	request, err := http.NewRequest("POST", url, bytes.NewReader(byteSlice))
	Expect(err).NotTo(HaveOccurred())
	request.Header.Add("Content-Type", marshal.DefaultUnmarshal.ContentType())
	request.Header.Add("Accept", marshal.DefaultMarshal.ContentType())
	return request
}
