// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/EricHripko/pack.yaml/pkg/packer2llb (interfaces: Plugin)

// Package packer2llb_mock is a generated GoMock package.
package packer2llb_mock

import (
	context "context"
	cib "github.com/EricHripko/pack.yaml/pkg/cib"
	gomock "github.com/golang/mock/gomock"
	llb "github.com/moby/buildkit/client/llb"
	dockerfile2llb "github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	client "github.com/moby/buildkit/frontend/gateway/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	reflect "reflect"
)

// MockPlugin is a mock of Plugin interface
type MockPlugin struct {
	ctrl     *gomock.Controller
	recorder *MockPluginMockRecorder
}

// MockPluginMockRecorder is the mock recorder for MockPlugin
type MockPluginMockRecorder struct {
	mock *MockPlugin
}

// NewMockPlugin creates a new mock instance
func NewMockPlugin(ctrl *gomock.Controller) *MockPlugin {
	mock := &MockPlugin{ctrl: ctrl}
	mock.recorder = &MockPluginMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPlugin) EXPECT() *MockPluginMockRecorder {
	return m.recorder
}

// Build mocks base method
func (m *MockPlugin) Build(arg0 context.Context, arg1 *v1.Platform, arg2 cib.Service) (*llb.State, *dockerfile2llb.Image, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Build", arg0, arg1, arg2)
	ret0, _ := ret[0].(*llb.State)
	ret1, _ := ret[1].(*dockerfile2llb.Image)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Build indicates an expected call of Build
func (mr *MockPluginMockRecorder) Build(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Build", reflect.TypeOf((*MockPlugin)(nil).Build), arg0, arg1, arg2)
}

// Detect mocks base method
func (m *MockPlugin) Detect(arg0 context.Context, arg1 client.Reference, arg2 *cib.Config) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Detect", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Detect indicates an expected call of Detect
func (mr *MockPluginMockRecorder) Detect(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Detect", reflect.TypeOf((*MockPlugin)(nil).Detect), arg0, arg1, arg2)
}
