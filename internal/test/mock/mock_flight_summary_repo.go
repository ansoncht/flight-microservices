// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/repository/flight_summary.go
//
// Generated by this command:
//
//	mockgen -source pkg/repository/flight_summary.go -destination=internal/test/mock/mock_flight_summary_repo.go -package=mock
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	model "github.com/ansoncht/flight-microservices/pkg/model"
	gomock "go.uber.org/mock/gomock"
)

// MockSummaryRepository is a mock of SummaryRepository interface.
type MockSummaryRepository struct {
	ctrl     *gomock.Controller
	recorder *MockSummaryRepositoryMockRecorder
	isgomock struct{}
}

// MockSummaryRepositoryMockRecorder is the mock recorder for MockSummaryRepository.
type MockSummaryRepositoryMockRecorder struct {
	mock *MockSummaryRepository
}

// NewMockSummaryRepository creates a new mock instance.
func NewMockSummaryRepository(ctrl *gomock.Controller) *MockSummaryRepository {
	mock := &MockSummaryRepository{ctrl: ctrl}
	mock.recorder = &MockSummaryRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSummaryRepository) EXPECT() *MockSummaryRepositoryMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockSummaryRepository) Get(ctx context.Context, id string) (*model.DailyFlightSummary, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, id)
	ret0, _ := ret[0].(*model.DailyFlightSummary)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockSummaryRepositoryMockRecorder) Get(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockSummaryRepository)(nil).Get), ctx, id)
}

// Insert mocks base method.
func (m *MockSummaryRepository) Insert(ctx context.Context, summary model.DailyFlightSummary) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", ctx, summary)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Insert indicates an expected call of Insert.
func (mr *MockSummaryRepositoryMockRecorder) Insert(ctx, summary any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*MockSummaryRepository)(nil).Insert), ctx, summary)
}
