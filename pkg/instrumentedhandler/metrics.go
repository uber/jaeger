// Copyright (c) 2019 The Jaeger Authors.
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

package instrumentedhandler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/uber/jaeger-lib/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}

func NewMetricsHandler(metricsFactory metrics.Factory) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w}
			h.ServeHTTP(recorder, r)

			requestDuration := metricsFactory.Timer(metrics.TimerOptions{
				Name: "http.request.duration",
				Help: "Duration of HTTP requests",
				Tags: map[string]string{
					"status": strconv.Itoa(recorder.status),
					"path":   r.URL.Path,
					"method": r.Method,
				},
			})
			requestDuration.Record(time.Since(start))
		})
	}
}
