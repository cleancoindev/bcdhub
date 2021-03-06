// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go

// Package mock_tzkt is a generated GoMock package.
package tzkt

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockService is a mock of Service interface
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// GetMempool mocks base method
func (m *MockService) GetMempool(address string) ([]MempoolOperation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMempool", address)
	ret0, _ := ret[0].([]MempoolOperation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMempool indicates an expected call of GetMempool
func (mr *MockServiceMockRecorder) GetMempool(address interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMempool", reflect.TypeOf((*MockService)(nil).GetMempool), address)
}
