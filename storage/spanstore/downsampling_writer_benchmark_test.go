package spanstore

import (
	"math"
	"math/rand"
	"sync"
	"testing"

	"github.com/uber/jaeger-lib/metrics"

	"github.com/jaegertracing/jaeger/model"
	"github.com/stretchr/testify/assert"
)

type noopWriteSpanStore struct{}

func (n *noopWriteSpanStore) WriteSpan(span *model.Span) error {
	return nil
}

// Benchmark result:
// numGoroutines = 10000, writeTimes = 10
// BenchmarkDownSamplingWriter_WriteSpan-12    	     100	  25536211 ns/op	 9110205 B/op	  210443 allocs/op
func BenchmarkDownSamplingWriter_WriteSpan(b *testing.B) {
	trace := model.TraceID{
		Low:  uint64(0),
		High: uint64(1),
	}
	span := &model.Span{
		TraceID: trace,
	}
	c := NewDownSamplingWriter(&noopWriteSpanStore{}, DownSamplingOptions{
		Ratio:    0.5,
		HashSalt: "jaeger-test",
	})
	var wg sync.WaitGroup
	numGoroutines := 10000
	writeTimes := 10
	wg.Add(numGoroutines * b.N)
	b.ResetTimer()
	b.ReportAllocs()

	for j := 0; j < b.N; j++ {
		for n := 0; n < numGoroutines; n++ {

			go func() {
				for i := 0; i < writeTimes; i++ {
					c.WriteSpan(span)
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()
}

func TestDownSamplingWriter_RandomHash(t *testing.T) {
	ratioThreshold := uint64(math.MaxUint64 / 2)
	countSmallerThanRatio := 0
	downSamplingOptions := DownSamplingOptions{
		Ratio:          1,
		HashSalt:       "jaeger-test",
		MetricsFactory: metrics.NullFactory,
	}
	c := NewDownSamplingWriter(&noopWriteSpanStore{}, downSamplingOptions)
	for i := 0; i < 100000; i++ {
		low := rand.Uint64()
		high := rand.Uint64()
		span := &model.Span{
			TraceID: model.TraceID{
				Low:  low,
				High: high,
			},
		}
		hash := c.hashBytes([]byte(span.TraceID.String()))
		if hash < ratioThreshold {
			countSmallerThanRatio++
		}
	}
	assert.True(t, math.Abs(float64(countSmallerThanRatio)/100000-0.5) < 0.05)
}
