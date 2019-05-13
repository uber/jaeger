// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"
import mock "github.com/stretchr/testify/mock"
import storage_v1 "github.com/jaegertracing/jaeger/proto-gen/storage_v1"

// SpanReaderPluginServer is an autogenerated mock type for the SpanReaderPluginServer type
type SpanReaderPluginServer struct {
	mock.Mock
}

// FindTraceIDs provides a mock function with given fields: _a0, _a1
func (_m *SpanReaderPluginServer) FindTraceIDs(_a0 context.Context, _a1 *storage_v1.FindTraceIDsRequest) (*storage_v1.FindTraceIDsResponse, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *storage_v1.FindTraceIDsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *storage_v1.FindTraceIDsRequest) *storage_v1.FindTraceIDsResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage_v1.FindTraceIDsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *storage_v1.FindTraceIDsRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindTraces provides a mock function with given fields: _a0, _a1
func (_m *SpanReaderPluginServer) FindTraces(_a0 *storage_v1.FindTracesRequest, _a1 storage_v1.SpanReaderPlugin_FindTracesServer) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(*storage_v1.FindTracesRequest, storage_v1.SpanReaderPlugin_FindTracesServer) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetOperations provides a mock function with given fields: _a0, _a1
func (_m *SpanReaderPluginServer) GetOperations(_a0 context.Context, _a1 *storage_v1.GetOperationsRequest) (*storage_v1.GetOperationsResponse, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *storage_v1.GetOperationsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *storage_v1.GetOperationsRequest) *storage_v1.GetOperationsResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage_v1.GetOperationsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *storage_v1.GetOperationsRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetServices provides a mock function with given fields: _a0, _a1
func (_m *SpanReaderPluginServer) GetServices(_a0 context.Context, _a1 *storage_v1.GetServicesRequest) (*storage_v1.GetServicesResponse, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *storage_v1.GetServicesResponse
	if rf, ok := ret.Get(0).(func(context.Context, *storage_v1.GetServicesRequest) *storage_v1.GetServicesResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage_v1.GetServicesResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *storage_v1.GetServicesRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTrace provides a mock function with given fields: _a0, _a1
func (_m *SpanReaderPluginServer) GetTrace(_a0 *storage_v1.GetTraceRequest, _a1 storage_v1.SpanReaderPlugin_GetTraceServer) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(*storage_v1.GetTraceRequest, storage_v1.SpanReaderPlugin_GetTraceServer) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
