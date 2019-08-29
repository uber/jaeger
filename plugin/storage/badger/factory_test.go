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

package badger

import (
	"bytes"
	"encoding/binary"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/dgraph-io/badger"
	assert "github.com/stretchr/testify/require"
	"github.com/uber/jaeger-lib/metrics"
	"github.com/uber/jaeger-lib/metrics/metricstest"
	"go.uber.org/zap"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/pkg/config"
)

func TestInitializationErrors(t *testing.T) {
	f := NewFactory()
	v, command := config.Viperize(f.AddFlags)
	dir := "/root/this_should_fail" // If this test fails, you have some issues in your system
	keyParam := fmt.Sprintf("--badger.directory-key=%s", dir)
	valueParam := fmt.Sprintf("--badger.directory-value=%s", dir)

	command.ParseFlags([]string{
		"--badger.ephemeral=false",
		"--badger.consistency=true",
		keyParam,
		valueParam,
	})
	f.InitFromViper(v)

	err := f.Initialize(metrics.NullFactory, zap.NewNop())
	assert.Error(t, err)
}

func TestForCodecov(t *testing.T) {
	// These tests are testing our vendor packages and are intended to satisfy Codecov.
	f := NewFactory()
	v, _ := config.Viperize(f.AddFlags)
	f.InitFromViper(v)

	err := f.Initialize(metrics.NullFactory, zap.NewNop())
	assert.NoError(t, err)

	// Get all the writers, readers, etc
	_, err = f.CreateSpanReader()
	assert.NoError(t, err)

	_, err = f.CreateSpanWriter()
	assert.NoError(t, err)

	_, err = f.CreateDependencyReader()
	assert.NoError(t, err)

	// Now, remove the badger directories
	err = os.RemoveAll(f.tmpDir)
	assert.NoError(t, err)

	// Now try to close, since the files have been deleted this should throw an error
	err = f.Close()
	assert.Error(t, err)
}

func TestMaintenanceRun(t *testing.T) {
	// For Codecov - this does not test anything
	f := NewFactory()
	v, command := config.Viperize(f.AddFlags)
	// Lets speed up the maintenance ticker..
	command.ParseFlags([]string{
		"--badger.maintenance-interval=10ms",
	})
	f.InitFromViper(v)
	// Safeguard
	mFactory := metricstest.NewFactory(0)
	_, gs := mFactory.Snapshot()
	assert.True(t, gs[lastMaintenanceRunName] == 0)
	f.Initialize(mFactory, zap.NewNop())

	waiter := func(previousValue int64) int64 {
		sleeps := 0
		_, gs := mFactory.Snapshot()
		for gs[lastMaintenanceRunName] == previousValue && sleeps < 8 {
			// Wait for the scheduler
			time.Sleep(time.Duration(50) * time.Millisecond)
			sleeps++
			_, gs = mFactory.Snapshot()
		}
		assert.True(t, gs[lastMaintenanceRunName] > previousValue)
		return gs[lastMaintenanceRunName]
	}

	runtime := waiter(0) // First run, check that it was ran and caches previous size

	// This is to for codecov only. Can break without anything else breaking as it does test badger's
	// internal implementation
	vlogSize := expvar.Get("badger_vlog_size_bytes").(*expvar.Map).Get(f.tmpDir).(*expvar.Int)
	currSize := vlogSize.Value()
	vlogSize.Set(currSize + 1<<31)

	waiter(runtime)
	_, gs = mFactory.Snapshot()
	assert.True(t, gs[lastValueLogCleanedName] > 0)

	err := io.Closer(f).Close()
	assert.NoError(t, err)
}

// TestSchemaVersionWritten ensures that the Factory does call the SchemaManager correctly
func TestSchemaVersionWritten(t *testing.T) {
	f := NewFactory()
	v, _ := config.Viperize(f.AddFlags)
	f.InitFromViper(v)
	mFactory := metricstest.NewFactory(0)
	f.Initialize(mFactory, zap.NewNop())

	schemaVersion := -1
	err := f.store.View(func(txn *badger.Txn) error {
		schemaKey := []byte{0x11}
		item, err := txn.Get(schemaKey)
		if err != nil {
			return err
		}

		val, err := item.Value()
		if err != nil {
			return err
		}
		schemaVersion = int(binary.BigEndian.Uint32(val))
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, schemaVersion > 0)
}

func TestBadgerMetrics(t *testing.T) {
	// The expvar is leaking keyparams between tests. We need to clean up a bit..
	eMap := expvar.Get("badger_lsm_size_bytes").(*expvar.Map)
	eMap.Init()

	f := NewFactory()
	v, command := config.Viperize(f.AddFlags)
	command.ParseFlags([]string{
		"--badger.metrics-update-interval=10ms",
	})
	f.InitFromViper(v)
	mFactory := metricstest.NewFactory(0)
	f.Initialize(mFactory, zap.NewNop())
	assert.NotNil(t, f.metrics.badgerMetrics)
	_, found := f.metrics.badgerMetrics["badger_memtable_gets_total"]
	assert.True(t, found)

	waiter := func(previousValue int64) int64 {
		sleeps := 0
		_, gs := mFactory.Snapshot()
		for gs["badger_memtable_gets_total"] == previousValue && sleeps < 8 {
			// Wait for the scheduler
			time.Sleep(time.Duration(50) * time.Millisecond)
			sleeps++
			_, gs = mFactory.Snapshot()
		}
		assert.True(t, gs["badger_memtable_gets_total"] > previousValue)
		return gs["badger_memtable_gets_total"]
	}

	vlogSize := waiter(0)
	_, gs := mFactory.Snapshot()
	assert.True(t, vlogSize > 0)
	assert.True(t, gs["badger_memtable_gets_total"] > 0) // IntVal metric

	_, found = gs["badger_lsm_size_bytes"] // Map metric
	assert.True(t, found)

	err := f.Close()
	assert.NoError(t, err)
}

func TestFailToInitializeByCorruptingDB(t *testing.T) {
	f := NewFactory()
	v, command := config.Viperize(f.AddFlags)
	dir, _ := ioutil.TempDir("", "badger")
	defer os.RemoveAll(dir)

	keyParam := fmt.Sprintf("--badger.directory-key=%s", dir)
	valueParam := fmt.Sprintf("--badger.directory-value=%s", dir)

	command.ParseFlags([]string{
		"--badger.ephemeral=false",
		"--badger.consistency=true",
		keyParam,
		valueParam,
	})
	f.InitFromViper(v)

	err := f.Initialize(metrics.NullFactory, zap.NewNop())
	assert.NoError(t, err)

	sw, err := f.CreateSpanWriter()
	assert.NoError(t, err)

	err = sw.WriteSpan(createDummySpan())
	assert.NoError(t, err)

	err = f.Close()
	assert.NoError(t, err)

	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	store, err := badger.Open(opts)
	assert.NoError(t, err)

	// Corrupt the data
	err = store.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		spanKey := []byte{0x80}
		it.Seek(spanKey)
		if it.Item() != nil && bytes.HasPrefix(it.Item().Key(), spanKey) {
			val := []byte{}
			val, err = it.Item().ValueCopy(val)
			assert.NoError(t, err)

			val[0] = 0 // Corrupt the proto
			keyCopy := []byte{}
			keyCopy = it.Item().KeyCopy(keyCopy)
			err = txn.Set(keyCopy, val)
			assert.NoError(t, err)
		}

		// Delete the schema key to cause proto unmarshalling
		err = txn.Delete([]byte{0x11})
		assert.NoError(t, err)

		return nil
	})
	assert.NoError(t, err)

	err = store.Close()
	assert.NoError(t, err)

	err = f.Initialize(metrics.NullFactory, zap.NewNop())
	assert.Error(t, err)
}

func createDummySpan() *model.Span {
	tid := time.Now()

	dummyKv := []model.KeyValue{
		{
			Key:   "key",
			VType: model.StringType,
			VStr:  "value",
		},
	}

	testSpan := model.Span{
		TraceID: model.TraceID{
			Low:  uint64(0),
			High: 1,
		},
		SpanID:        model.SpanID(0),
		OperationName: "operation",
		Process: &model.Process{
			ServiceName: "service",
			Tags:        dummyKv,
		},
		StartTime: tid.Add(time.Duration(1 * time.Millisecond)),
		Duration:  time.Duration(1 * time.Millisecond),
		Tags:      dummyKv,
		Logs: []model.Log{
			{
				Timestamp: tid,
				Fields:    dummyKv,
			},
		},
	}

	return &testSpan
}
