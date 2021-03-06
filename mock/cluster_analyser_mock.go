// Code generated by MockGen. DO NOT EDIT.
// Source: analytic/cluster_analyser.go

// Package mock is a generated GoMock package.
package mock

import (
	sarama "github.com/Shopify/sarama"
	sarama_cluster "github.com/bsm/sarama-cluster"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockClusterAnalyser is a mock of ClusterAnalyser interface
type MockClusterAnalyser struct {
	ctrl     *gomock.Controller
	recorder *MockClusterAnalyserMockRecorder
}

// MockClusterAnalyserMockRecorder is the mock recorder for MockClusterAnalyser
type MockClusterAnalyserMockRecorder struct {
	mock *MockClusterAnalyser
}

// NewMockClusterAnalyser creates a new mock instance
func NewMockClusterAnalyser(ctrl *gomock.Controller) *MockClusterAnalyser {
	mock := &MockClusterAnalyser{ctrl: ctrl}
	mock.recorder = &MockClusterAnalyserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClusterAnalyser) EXPECT() *MockClusterAnalyserMockRecorder {
	return m.recorder
}

// Messages mocks base method
func (m *MockClusterAnalyser) Messages() <-chan *sarama.ConsumerMessage {
	ret := m.ctrl.Call(m, "Messages")
	ret0, _ := ret[0].(<-chan *sarama.ConsumerMessage)
	return ret0
}

// Messages indicates an expected call of Messages
func (mr *MockClusterAnalyserMockRecorder) Messages() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Messages", reflect.TypeOf((*MockClusterAnalyser)(nil).Messages))
}

// Errors mocks base method
func (m *MockClusterAnalyser) Errors() <-chan error {
	ret := m.ctrl.Call(m, "Errors")
	ret0, _ := ret[0].(<-chan error)
	return ret0
}

// Errors indicates an expected call of Errors
func (mr *MockClusterAnalyserMockRecorder) Errors() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Errors", reflect.TypeOf((*MockClusterAnalyser)(nil).Errors))
}

// Notifications mocks base method
func (m *MockClusterAnalyser) Notifications() <-chan *sarama_cluster.Notification {
	ret := m.ctrl.Call(m, "Notifications")
	ret0, _ := ret[0].(<-chan *sarama_cluster.Notification)
	return ret0
}

// Notifications indicates an expected call of Notifications
func (mr *MockClusterAnalyserMockRecorder) Notifications() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Notifications", reflect.TypeOf((*MockClusterAnalyser)(nil).Notifications))
}

// MarkOffset mocks base method
func (m *MockClusterAnalyser) MarkOffset(message *sarama.ConsumerMessage, metadata string) {
	m.ctrl.Call(m, "MarkOffset", message, metadata)
}

// MarkOffset indicates an expected call of MarkOffset
func (mr *MockClusterAnalyserMockRecorder) MarkOffset(message, metadata interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkOffset", reflect.TypeOf((*MockClusterAnalyser)(nil).MarkOffset), message, metadata)
}

// Close mocks base method
func (m *MockClusterAnalyser) Close() error {
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockClusterAnalyserMockRecorder) Close() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockClusterAnalyser)(nil).Close))
}
