// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/konflux-workspaces/workspaces/server/core/workspace (interfaces: WorkspaceUpdater,WorkspaceReader,WorkspaceLister,WorkspaceCreator)
//
// Generated by this command:
//
//	mockgen -destination=mocks_generated_test.go -package=workspace_test . WorkspaceUpdater,WorkspaceReader,WorkspaceLister,WorkspaceCreator
//

// Package workspace_test is a generated GoMock package.
package workspace_test

import (
	context "context"
	reflect "reflect"

	v1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	gomock "go.uber.org/mock/gomock"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockWorkspaceUpdater is a mock of WorkspaceUpdater interface.
type MockWorkspaceUpdater struct {
	ctrl     *gomock.Controller
	recorder *MockWorkspaceUpdaterMockRecorder
}

// MockWorkspaceUpdaterMockRecorder is the mock recorder for MockWorkspaceUpdater.
type MockWorkspaceUpdaterMockRecorder struct {
	mock *MockWorkspaceUpdater
}

// NewMockWorkspaceUpdater creates a new mock instance.
func NewMockWorkspaceUpdater(ctrl *gomock.Controller) *MockWorkspaceUpdater {
	mock := &MockWorkspaceUpdater{ctrl: ctrl}
	mock.recorder = &MockWorkspaceUpdaterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWorkspaceUpdater) EXPECT() *MockWorkspaceUpdaterMockRecorder {
	return m.recorder
}

// UpdateUserWorkspace mocks base method.
func (m *MockWorkspaceUpdater) UpdateUserWorkspace(arg0 context.Context, arg1 string, arg2 *v1alpha1.Workspace, arg3 ...client.UpdateOption) error {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateUserWorkspace", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUserWorkspace indicates an expected call of UpdateUserWorkspace.
func (mr *MockWorkspaceUpdaterMockRecorder) UpdateUserWorkspace(arg0, arg1, arg2 any, arg3 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUserWorkspace", reflect.TypeOf((*MockWorkspaceUpdater)(nil).UpdateUserWorkspace), varargs...)
}

// MockWorkspaceReader is a mock of WorkspaceReader interface.
type MockWorkspaceReader struct {
	ctrl     *gomock.Controller
	recorder *MockWorkspaceReaderMockRecorder
}

// MockWorkspaceReaderMockRecorder is the mock recorder for MockWorkspaceReader.
type MockWorkspaceReaderMockRecorder struct {
	mock *MockWorkspaceReader
}

// NewMockWorkspaceReader creates a new mock instance.
func NewMockWorkspaceReader(ctrl *gomock.Controller) *MockWorkspaceReader {
	mock := &MockWorkspaceReader{ctrl: ctrl}
	mock.recorder = &MockWorkspaceReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWorkspaceReader) EXPECT() *MockWorkspaceReaderMockRecorder {
	return m.recorder
}

// ReadUserWorkspace mocks base method.
func (m *MockWorkspaceReader) ReadUserWorkspace(arg0 context.Context, arg1, arg2, arg3 string, arg4 *v1alpha1.Workspace, arg5 ...client.GetOption) error {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2, arg3, arg4}
	for _, a := range arg5 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ReadUserWorkspace", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReadUserWorkspace indicates an expected call of ReadUserWorkspace.
func (mr *MockWorkspaceReaderMockRecorder) ReadUserWorkspace(arg0, arg1, arg2, arg3, arg4 any, arg5 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2, arg3, arg4}, arg5...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadUserWorkspace", reflect.TypeOf((*MockWorkspaceReader)(nil).ReadUserWorkspace), varargs...)
}

// MockWorkspaceLister is a mock of WorkspaceLister interface.
type MockWorkspaceLister struct {
	ctrl     *gomock.Controller
	recorder *MockWorkspaceListerMockRecorder
}

// MockWorkspaceListerMockRecorder is the mock recorder for MockWorkspaceLister.
type MockWorkspaceListerMockRecorder struct {
	mock *MockWorkspaceLister
}

// NewMockWorkspaceLister creates a new mock instance.
func NewMockWorkspaceLister(ctrl *gomock.Controller) *MockWorkspaceLister {
	mock := &MockWorkspaceLister{ctrl: ctrl}
	mock.recorder = &MockWorkspaceListerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWorkspaceLister) EXPECT() *MockWorkspaceListerMockRecorder {
	return m.recorder
}

// ListUserWorkspaces mocks base method.
func (m *MockWorkspaceLister) ListUserWorkspaces(arg0 context.Context, arg1 string, arg2 *v1alpha1.WorkspaceList, arg3 ...client.ListOption) error {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListUserWorkspaces", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// ListUserWorkspaces indicates an expected call of ListUserWorkspaces.
func (mr *MockWorkspaceListerMockRecorder) ListUserWorkspaces(arg0, arg1, arg2 any, arg3 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUserWorkspaces", reflect.TypeOf((*MockWorkspaceLister)(nil).ListUserWorkspaces), varargs...)
}

// MockWorkspaceCreator is a mock of WorkspaceCreator interface.
type MockWorkspaceCreator struct {
	ctrl     *gomock.Controller
	recorder *MockWorkspaceCreatorMockRecorder
}

// MockWorkspaceCreatorMockRecorder is the mock recorder for MockWorkspaceCreator.
type MockWorkspaceCreatorMockRecorder struct {
	mock *MockWorkspaceCreator
}

// NewMockWorkspaceCreator creates a new mock instance.
func NewMockWorkspaceCreator(ctrl *gomock.Controller) *MockWorkspaceCreator {
	mock := &MockWorkspaceCreator{ctrl: ctrl}
	mock.recorder = &MockWorkspaceCreatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWorkspaceCreator) EXPECT() *MockWorkspaceCreatorMockRecorder {
	return m.recorder
}

// CreateUserWorkspace mocks base method.
func (m *MockWorkspaceCreator) CreateUserWorkspace(arg0 context.Context, arg1 string, arg2 *v1alpha1.Workspace, arg3 ...client.CreateOption) error {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateUserWorkspace", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateUserWorkspace indicates an expected call of CreateUserWorkspace.
func (mr *MockWorkspaceCreatorMockRecorder) CreateUserWorkspace(arg0, arg1, arg2 any, arg3 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUserWorkspace", reflect.TypeOf((*MockWorkspaceCreator)(nil).CreateUserWorkspace), varargs...)
}
