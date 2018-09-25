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

package consumer

import (
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
)

// seppukuFactory is a factory for seppukuWorkers
type seppukuFactory struct {
	metricsFactory metrics.Factory
	logger         *zap.Logger
	interval       time.Duration
	panicFunc      func(int32)
}

type seppukuWorker struct {
	msgConsumed    *uint64
	ticker         *time.Ticker
	closePartition chan struct{}
}

func newSeppukuFactory(factory metrics.Factory, logger *zap.Logger, interval time.Duration) seppukuFactory {
	return seppukuFactory{
		metricsFactory: factory,
		logger:         logger,
		interval:       interval,
		panicFunc: func(partition int32) {
			factory.Counter("seppuku", map[string]string{"partition": strconv.Itoa(int(partition))}).Inc(1)
			time.Sleep(time.Second) // Allow time to flush metric

			buf := make([]byte, 1<<20)
			logger.Panic("No messages processed in the last check interval",
				zap.Int32("partition", partition),
				zap.String("stack", string(buf[:runtime.Stack(buf, true)])))
		},
	}
}

// startMonitoringForPartition monitors the messages consumed by the partition and signals for the partition to by
// closed by sending a message on the closePartition channel.
//
// Closing the partition should result in a rebalance, which alleviates the condition. This means that rebalances can
// happen frequently if there is no traffic on the Kafka topic. This shouldn't affect normal operations.
//
// If the message send isn't processed within the next check interval, a seppuku is attempted.This hack relies on a
// container management system (k8s, aurora, marathon, etc) to reschedule
// the dead instance.
//
// This hack protects jaeger-ingester from issues described in  https://github.com/jaegertracing/jaeger/issues/1052
//
func (s *seppukuFactory) startMonitoringForPartition(partition int32) *seppukuWorker {
	var msgConsumed uint64
	w := &seppukuWorker{
		msgConsumed:    &msgConsumed,
		ticker:         time.NewTicker(s.interval),
		closePartition: make(chan struct{}, 1),
	}

	go func() {
		for range w.ticker.C {
			if atomic.LoadUint64(w.msgConsumed) == 0 {
				select {
				case w.closePartition <- struct{}{}:
					s.logger.Warn("Signalling partition close due to inactivity", zap.Int32("partition", partition))
				default:
					// If closePartition is blocked, attempt seppuku
					s.panicFunc(partition)
				}
			} else {
				atomic.StoreUint64(w.msgConsumed, 0)
			}
		}
	}()

	return w
}

func (w *seppukuWorker) getClosePartition() chan struct{} {
	return w.closePartition
}

func (w *seppukuWorker) close() {
	w.ticker.Stop()
}

func (w *seppukuWorker) incrementMsgCount() {
	atomic.AddUint64(w.msgConsumed, 1)
}
