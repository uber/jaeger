// Copyright (c) 2017 Uber Technologies, Inc.
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
	"sync"

	"github.com/uber/jaeger-lib/metrics"

	"github.com/jaegertracing/jaeger/model"
)

const (
	maxServiceNames      = 2000
	otherServices        = "other-services"
	defaultTransportType = "undefined"
)

// SpanProcessorMetrics contains all the necessary metrics for the SpanProcessor
type SpanProcessorMetrics struct {
	//TODO - initialize metrics in the traditional factory way. Initialize map afterward.
	// SaveLatency measures how long the actual save to storage takes
	SaveLatency metrics.Timer
	// InQueueLatency measures how long the span spends in the queue
	InQueueLatency metrics.Timer
	// SpansDropped measures the number of spans we discarded because the queue was full
	SpansDropped metrics.Counter
	// BatchSize measures the span batch size
	BatchSize metrics.Gauge // size of span batch
	// QueueLength measures the size of the internal span queue
	QueueLength metrics.Gauge
	// SavedOkBySvc contains span and trace counts by service
	SavedOkBySvc  metricsBySvc  // spans actually saved
	SavedErrBySvc metricsBySvc  // spans failed to save
	serviceNames  metrics.Gauge // total number of unique service name metrics reported by this collector
	spanCounts    map[string]CountsBySpanType
}

type countsBySvc struct {
	counts          map[string]metrics.Counter // counters per service
	debugCounts     map[string]metrics.Counter // debug counters per service
	factory         metrics.Factory
	lock            *sync.Mutex
	maxServiceNames int
	category        string
}

type metricsBySvc struct {
	spans  countsBySvc // number of spans received per service
	traces countsBySvc // number of traces originated per service
}

// CountsBySpanType measures received, rejected, and receivedByService metrics for a format type
type CountsBySpanType struct {
	// ReceivedBySvc maintain by-service metrics for a format type
	ReceivedBySvc metricsBySvc
	// RejectedBySvc is the number of spans we rejected (usually due to blacklisting) by-service
	RejectedBySvc metricsBySvc
}

// NewSpanProcessorMetrics returns a SpanProcessorMetrics
func NewSpanProcessorMetrics(serviceMetrics metrics.Factory, hostMetrics metrics.Factory, otherFormatTypes []string) *SpanProcessorMetrics {
	spanCounts := map[string]CountsBySpanType{
		ZipkinFormatType:  newCountsBySpanType(serviceMetrics.Namespace(metrics.NSOptions{Name: "", Tags: map[string]string{"format": ZipkinFormatType}})),
		JaegerFormatType:  newCountsBySpanType(serviceMetrics.Namespace(metrics.NSOptions{Name: "", Tags: map[string]string{"format": JaegerFormatType}})),
		UnknownFormatType: newCountsBySpanType(serviceMetrics.Namespace(metrics.NSOptions{Name: "", Tags: map[string]string{"format": UnknownFormatType}})),
	}
	for _, otherFormatType := range otherFormatTypes {
		spanCounts[otherFormatType] = newCountsBySpanType(serviceMetrics.Namespace(metrics.NSOptions{Name: "", Tags: map[string]string{"format": otherFormatType}}))
	}
	m := &SpanProcessorMetrics{
		SaveLatency:    hostMetrics.Timer(metrics.TimerOptions{Name: "save-latency", Tags: nil}),
		InQueueLatency: hostMetrics.Timer(metrics.TimerOptions{Name: "in-queue-latency", Tags: nil}),
		SpansDropped:   hostMetrics.Counter(metrics.Options{Name: "spans.dropped", Tags: nil}),
		BatchSize:      hostMetrics.Gauge(metrics.Options{Name: "batch-size", Tags: nil}),
		QueueLength:    hostMetrics.Gauge(metrics.Options{Name: "queue-length", Tags: nil}),
		SavedOkBySvc:   newMetricsBySvc(serviceMetrics.Namespace(metrics.NSOptions{Name: "", Tags: map[string]string{"result": "ok"}}), "saved-by-svc"),
		SavedErrBySvc:  newMetricsBySvc(serviceMetrics.Namespace(metrics.NSOptions{Name: "", Tags: map[string]string{"result": "err"}}), "saved-by-svc"),
		spanCounts:     spanCounts,
		serviceNames:   hostMetrics.Gauge(metrics.Options{Name: "spans.serviceNames", Tags: nil}),
	}

	return m
}

func newMetricsBySvc(factory metrics.Factory, category string) metricsBySvc {
	spansFactory := factory.Namespace(metrics.NSOptions{Name: "spans", Tags: nil})
	tracesFactory := factory.Namespace(metrics.NSOptions{Name: "traces", Tags: nil})
	return metricsBySvc{
		spans:  newCountsBySvc(spansFactory, category, maxServiceNames),
		traces: newCountsBySvc(tracesFactory, category, maxServiceNames),
	}
}

func newCountsBySvc(factory metrics.Factory, category string, maxServiceNames int) countsBySvc {
	return countsBySvc{
		counts: map[string]metrics.Counter{
			otherServices: factory.Counter(metrics.Options{Name: category, Tags: map[string]string{"svc": otherServices, "debug": "false"}}),
		},
		debugCounts: map[string]metrics.Counter{
			otherServices: factory.Counter(metrics.Options{Name: category, Tags: map[string]string{"svc": otherServices, "debug": "true"}}),
		},
		factory:         factory,
		lock:            &sync.Mutex{},
		maxServiceNames: maxServiceNames,
		category:        category,
	}
}

func newCountsBySpanType(factory metrics.Factory) CountsBySpanType {
	return CountsBySpanType{
		RejectedBySvc: newMetricsBySvc(factory, "rejected"),
		ReceivedBySvc: newMetricsBySvc(factory, "received"),
	}
}

// GetCountsForFormat gets the countsBySpanType for a given format. If none exists, we use the Unknown format.
func (m *SpanProcessorMetrics) GetCountsForFormat(spanFormat string) CountsBySpanType {
	c, ok := m.spanCounts[spanFormat]
	if !ok {
		return m.spanCounts[UnknownFormatType]
	}
	return c
}

// reportServiceNameForSpan determines the name of the service that emitted
// the span and reports a counter stat.
func (m metricsBySvc) ReportServiceNameForSpan(span *model.Span, endpoint string) {
	serviceName := span.Process.ServiceName
	if serviceName == "" {
		return
	}
	m.countSpansByServiceName(serviceName, span.Flags.IsDebug(), endpoint)
	if span.ParentSpanID() == 0 {
		m.countTracesByServiceName(serviceName, span.Flags.IsDebug(), endpoint)
	}
}

// countSpansByServiceName counts how many spans are received per service.
func (m metricsBySvc) countSpansByServiceName(serviceName string, isDebug bool, endpoint string) {
	m.spans.countByServiceName(serviceName, isDebug, endpoint)
}

// countTracesByServiceName counts how many traces are received per service,
// i.e. the counter is only incremented for the root spans.
func (m metricsBySvc) countTracesByServiceName(serviceName string, isDebug bool, endpoint string) {
	m.traces.countByServiceName(serviceName, isDebug, endpoint)
}

// countByServiceName maintains a map of counters for each service name it's
// given and increments the respective counter when called. The service name
// are first normalized to safe-for-metrics format.  If the number of counters
// exceeds maxServiceNames, new service names are ignored to avoid polluting
// the metrics namespace and overloading M3.
//
// The reportServiceNameCount() function runs on a timer and will report the
// total number of stored counters, so if it exceeds say the 90% threshold
// an alert should be raised to investigate what's causing so many unique
// service names.
func (m *countsBySvc) countByServiceName(serviceName string, isDebug bool, endpointType string) {
	serviceName = NormalizeServiceName(serviceName)
	counts := m.counts
	if isDebug {
		counts = m.debugCounts
	}
	var counter metrics.Counter
	m.lock.Lock()
	if c, ok := counts[serviceName]; ok {
		counter = c
	} else if len(counts) < m.maxServiceNames {
		debugStr := "false"
		if isDebug {
			debugStr = "true"
		}
		tags := map[string]string{"svc": serviceName, "debug": debugStr, "transport": defaultTransportType}
		if endpointType != "" {
			tags["transport"] = endpointType
		}
		c := m.factory.Counter(metrics.Options{Name: m.category, Tags: tags})
		counts[serviceName] = c
		counter = c
	} else {
		counter = counts[otherServices]
	}
	m.lock.Unlock()
	counter.Inc(1)
}
