// Copyright (c) 2018 The Jaeger Authors.
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

package app

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/proto-gen/api_v2"
)

type mockSpanProcessor struct {
	expectedError error
	mux           sync.Mutex
	spans         []*model.Span
}

func (p *mockSpanProcessor) ProcessSpans(spans []*model.Span, spanFormat string) ([]bool, error) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.spans = append(p.spans, spans...)
	oks := make([]bool, len(spans))
	return oks, p.expectedError
}

func (p *mockSpanProcessor) getSpans() []*model.Span {
	p.mux.Lock()
	defer p.mux.Unlock()
	return p.spans
}

func initializeGRPCTestServer(t *testing.T, expectedError error) (*grpc.Server, *mockSpanProcessor, net.Addr) {
	server := grpc.NewServer()
	processor := &mockSpanProcessor{expectedError: expectedError}
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	handler := NewGRPCHandler(logger, processor)
	api_v2.RegisterCollectorServiceServer(server, handler)
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	go func() {
		err := server.Serve(lis)
		require.NoError(t, err)
	}()
	return server, processor, lis.Addr()
}

func newClient(t *testing.T, addr net.Addr) (api_v2.CollectorServiceClient, *grpc.ClientConn) {
	conn, err := grpc.Dial(addr.String(), grpc.WithInsecure())
	require.NoError(t, err)
	return api_v2.NewCollectorServiceClient(conn), conn
}

func TestPostSpans(t *testing.T) {
	server, processor, addr := initializeGRPCTestServer(t, nil)
	defer server.Stop()
	client, conn := newClient(t, addr)
	defer conn.Close()
	r, err := client.PostSpans(context.Background(), &api_v2.PostSpansRequest{
		Batch: model.Batch{
			Spans: []*model.Span{
				{
					OperationName: "test-operation",
				},
			},
		},
	})
	require.NoError(t, err)
	require.False(t, r.GetOk())
	require.Len(t, processor.getSpans(), 1)
}

func TestPostSpansWithError(t *testing.T) {
	expectedError := errors.New("test-error")
	server, processor, addr := initializeGRPCTestServer(t, expectedError)
	defer server.Stop()
	client, conn := newClient(t, addr)
	defer conn.Close()
	r, err := client.PostSpans(context.Background(), &api_v2.PostSpansRequest{
		Batch: model.Batch{
			Spans: []*model.Span{
				{
					OperationName: "fake-operation",
				},
			},
		},
	})
	require.Error(t, err)
	require.Nil(t, r)
	require.Contains(t, err.Error(), expectedError.Error())
	require.Len(t, processor.getSpans(), 1)
}
