// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package builder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"

	. "github.com/uber/jaeger/cmd/builder"
	"github.com/uber/jaeger/pkg/cassandra"
	"github.com/uber/jaeger/pkg/es"
	"github.com/uber/jaeger/storage/spanstore/memory"
)

type mockElastic struct {
}

func (*mockElastic) NewClient() (es.Client, error) {
	return nil, nil
}

type mockCassandra struct {
}

func (*mockCassandra) NewSession() (cassandra.Session, error) {
	return nil, nil
}

func TestApplyOptions(t *testing.T) {
	opts := ApplyOptions(
		Options.CassandraSessionBuilder(&mockCassandra{}),
		Options.LoggerOption(zap.NewNop()),
		Options.MetricsFactoryOption(metrics.NullFactory),
		Options.MemoryStoreOption(memory.NewStore()),
		Options.ElasticsearchClientBuilder(&mockElastic{}),
	)
	assert.NotNil(t, opts.ElasticClientBuilder)
	assert.NotNil(t, opts.CassandraSessionBuilder)
	assert.NotNil(t, opts.Logger)
	assert.NotNil(t, opts.MetricsFactory)
}

func TestApplyNoOptions(t *testing.T) {
	opts := ApplyOptions()
	assert.NotNil(t, opts.Logger)
	assert.NotNil(t, opts.MetricsFactory)
}
