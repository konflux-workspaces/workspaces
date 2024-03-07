// Code generated by MockGen. DO NOT EDIT.
// Source: interface_test.go
//
// Generated by this command:
//
//	mockgen -source=interface_test.go -destination=mocks/writer.go -package=mocks FakeResponseWriter
//

// Package mocks is a generated GoMock package.
package mocks

import (
	http "net/http"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockFakeResponseWriter is a mock of FakeResponseWriter interface.
type MockFakeResponseWriter struct {
	ctrl     *gomock.Controller
	recorder *MockFakeResponseWriterMockRecorder
}

// MockFakeResponseWriterMockRecorder is the mock recorder for MockFakeResponseWriter.
type MockFakeResponseWriterMockRecorder struct {
	mock *MockFakeResponseWriter
}

// NewMockFakeResponseWriter creates a new mock instance.
func NewMockFakeResponseWriter(ctrl *gomock.Controller) *MockFakeResponseWriter {
	mock := &MockFakeResponseWriter{ctrl: ctrl}
	mock.recorder = &MockFakeResponseWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFakeResponseWriter) EXPECT() *MockFakeResponseWriterMockRecorder {
	return m.recorder
}

// Header mocks base method.
func (m *MockFakeResponseWriter) Header() http.Header {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Header")
	ret0, _ := ret[0].(http.Header)
	return ret0
}

// Header indicates an expected call of Header.
func (mr *MockFakeResponseWriterMockRecorder) Header() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Header", reflect.TypeOf((*MockFakeResponseWriter)(nil).Header))
}

// Write mocks base method.
func (m *MockFakeResponseWriter) Write(arg0 []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockFakeResponseWriterMockRecorder) Write(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockFakeResponseWriter)(nil).Write), arg0)
}

// WriteHeader mocks base method.
func (m *MockFakeResponseWriter) WriteHeader(statusCode int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "WriteHeader", statusCode)
}

// WriteHeader indicates an expected call of WriteHeader.
func (mr *MockFakeResponseWriterMockRecorder) WriteHeader(statusCode any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteHeader", reflect.TypeOf((*MockFakeResponseWriter)(nil).WriteHeader), statusCode)
}
