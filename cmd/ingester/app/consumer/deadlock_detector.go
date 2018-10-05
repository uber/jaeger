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

// deadlockDetector monitors the messages consumed and wither signals for the partition to be closed by sending a
// message on closePartition, or triggers a panic if the close fails. It triggers a panic if there are no messages
// consumed across all partitions.
//
// Closing the partition should result in a rebalance, which alleviates the condition. This means that rebalances can
// happen frequently if there is no traffic on the Kafka topic. This shouldn't affect normal operations.
//
// If the message send isn't processed within the next check interval, a panic is issued.This hack relies on a
// container management system (k8s, aurora, marathon, etc) to reschedule
// the dead instance.
//
// This hack protects jaeger-ingester from issues described in  https://github.com/jaegertracing/jaeger/issues/1052
//
type deadlockDetector struct {
	metricsFactory                metrics.Factory
	logger                        *zap.Logger
	interval                      time.Duration
	allPartitionsDeadlockDetector *allPartitionsDeadlockDetector
	panicFunc                     func(int32)
}

type partitionDeadlockDetector struct {
	msgConsumed                   *uint64
	logger                        *zap.Logger
	partition                     int32
	closePartition                chan struct{}
	done                          chan struct{}
	allPartitionsDeadlockDetector *allPartitionsDeadlockDetector
}

type allPartitionsDeadlockDetector struct {
	msgConsumed *uint64
	logger      *zap.Logger
	done        chan struct{}
}

func newDeadlockDetector(factory metrics.Factory, logger *zap.Logger, interval time.Duration) deadlockDetector {
	panicFunc := func(partition int32) {
		factory.Counter("deadlockdetector.panic-issued", map[string]string{"partition": strconv.Itoa(int(partition))}).Inc(1)
		time.Sleep(time.Second) // Allow time to flush metric

		buf := make([]byte, 1<<20)
		logger.Panic("No messages processed in the last check interval",
			zap.Int32("partition", partition),
			zap.String("stack", string(buf[:runtime.Stack(buf, true)])))
	}

	return deadlockDetector{
		metricsFactory: factory,
		logger:         logger,
		interval:       interval,
		panicFunc:      panicFunc,
	}
}

func (s *deadlockDetector) startMonitoringForPartition(partition int32) *partitionDeadlockDetector {
	var msgConsumed uint64
	w := &partitionDeadlockDetector{
		msgConsumed:    &msgConsumed,
		partition:      partition,
		closePartition: make(chan struct{}, 1),
		done:           make(chan struct{}),
		logger:         s.logger,
		allPartitionsDeadlockDetector: s.allPartitionsDeadlockDetector,
	}

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-w.done:
				s.logger.Info("Closing ticker routine", zap.Int32("partition", partition))
				return
			case <-ticker.C:
				if atomic.LoadUint64(w.msgConsumed) == 0 {
					select {
					case w.closePartition <- struct{}{}:
						s.logger.Warn("Signalling partition close due to inactivity", zap.Int32("partition", partition))
					default:
						// If closePartition is blocked, the consumer might have deadlocked - kill the process
						s.panicFunc(partition)
					}
				} else {
					atomic.StoreUint64(w.msgConsumed, 0)
				}
			}
		}
	}()

	return w
}

// start monitors that the sum of messages consumed across all partitions is non zero for the given interval
// If it is zero when there are producers producing messages on the topic, it means that sarama-cluster hasn't
// retrieved partition assignments. (This case will not be caught by startMonitoringForPartition because no partitions
// were retrieved).
func (s *deadlockDetector) start() {
	var msgConsumed uint64
	detector := &allPartitionsDeadlockDetector{
		msgConsumed: &msgConsumed,
		done:        make(chan struct{}),
		logger:      s.logger,
	}

	go func() {
		s.logger.Debug("Starting global deadlock detector")
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-detector.done:
				s.logger.Debug("Closing global ticker routine")
				return
			case <-ticker.C:
				if atomic.LoadUint64(detector.msgConsumed) == 0 {
					s.panicFunc(-1)
				} else {
					atomic.StoreUint64(detector.msgConsumed, 0)
				}
			}
		}
	}()

	s.allPartitionsDeadlockDetector = detector
}

func (s *deadlockDetector) close() {
	s.logger.Debug("Closing all partitions deadlock detector")
	s.allPartitionsDeadlockDetector.done <- struct{}{}
}

func (s *allPartitionsDeadlockDetector) incrementMsgCount() {
	atomic.AddUint64(s.msgConsumed, 1)
}

func (w *partitionDeadlockDetector) closePartitionChannel() chan struct{} {
	return w.closePartition
}

func (w *partitionDeadlockDetector) close() {
	w.logger.Debug("Closing deadlock detector", zap.Int32("partition", w.partition))
	w.done <- struct{}{}
}

func (w *partitionDeadlockDetector) incrementMsgCount() {
	w.allPartitionsDeadlockDetector.incrementMsgCount()
	atomic.AddUint64(w.msgConsumed, 1)
}
