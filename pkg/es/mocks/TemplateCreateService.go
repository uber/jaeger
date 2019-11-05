// Code generated by mockery v1.0.0. DO NOT EDIT.

// Copyright (c) 2019 The Jaeger Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package mocks

import (
	context "context"

	elastic "github.com/olivere/elastic"
	mock "github.com/stretchr/testify/mock"

	es "github.com/jaegertracing/jaeger/pkg/es"
)

// TemplateCreateService is an autogenerated mock type for the TemplateCreateService type
type TemplateCreateService struct {
	mock.Mock
}

// Body provides a mock function with given fields: mapping
func (_m *TemplateCreateService) Body(mapping string) es.TemplateCreateService {
	ret := _m.Called(mapping)

	var r0 es.TemplateCreateService
	if rf, ok := ret.Get(0).(func(string) es.TemplateCreateService); ok {
		r0 = rf(mapping)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(es.TemplateCreateService)
		}
	}

	return r0
}

// Do provides a mock function with given fields: ctx
func (_m *TemplateCreateService) Do(ctx context.Context) (*elastic.IndicesPutTemplateResponse, error) {
	ret := _m.Called(ctx)

	var r0 *elastic.IndicesPutTemplateResponse
	if rf, ok := ret.Get(0).(func(context.Context) *elastic.IndicesPutTemplateResponse); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*elastic.IndicesPutTemplateResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
