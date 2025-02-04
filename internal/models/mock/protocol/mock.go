// Code generated by MockGen. DO NOT EDIT.
// Source: protocol/repository.go

// Package mock_protocol is a generated GoMock package.
package mock_protocol

import (
	protocol "github.com/baking-bad/bcdhub/internal/models/protocol"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockRepository is a mock of Repository interface
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// GetProtocol mocks base method
func (m *MockRepository) GetProtocol(arg0, arg1 string, arg2 int64) (protocol.Protocol, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProtocol", arg0, arg1, arg2)
	ret0, _ := ret[0].(protocol.Protocol)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProtocol indicates an expected call of GetProtocol
func (mr *MockRepositoryMockRecorder) GetProtocol(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProtocol", reflect.TypeOf((*MockRepository)(nil).GetProtocol), arg0, arg1, arg2)
}

// GetSymLinks mocks base method
func (m *MockRepository) GetSymLinks(arg0 string, arg1 int64) (map[string]struct{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSymLinks", arg0, arg1)
	ret0, _ := ret[0].(map[string]struct{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSymLinks indicates an expected call of GetSymLinks
func (mr *MockRepositoryMockRecorder) GetSymLinks(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSymLinks", reflect.TypeOf((*MockRepository)(nil).GetSymLinks), arg0, arg1)
}
