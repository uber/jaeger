// Copyright (c) 2020 The Jaeger Authors.
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

package jaegerreceiver

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-collector/config/configerror"
	"github.com/open-telemetry/opentelemetry-collector/receiver/jaegerreceiver"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/jaegertracing/jaeger/plugin/sampling/strategystore/static"
)

func TestDefaultValueFromViper(t *testing.T) {
	v := viper.New()
	v.Set(static.SamplingStrategiesFile, "config.json")
	jr := &jaegerreceiver.Factory{}

	f := &Factory{
		Wrapped: jr,
		Viper:   v,
	}

	cfg := f.CreateDefaultConfig().(*jaegerreceiver.Config)
	assert.Equal(t, "config.json", cfg.RemoteSampling.StrategyFile)
}

func TestType(t *testing.T) {
	f := &Factory{
		Wrapped: &jaegerreceiver.Factory{},
	}
	assert.Equal(t, "jaeger", f.Type())
}

func TestCreateMetricsExporter(t *testing.T) {
	f := &Factory{
		Wrapped: &jaegerreceiver.Factory{},
	}
	mReceiver, err := f.CreateMetricsReceiver(zap.NewNop(), nil, nil)
	assert.Equal(t, err, configerror.ErrDataTypeIsNotSupported)
	assert.Nil(t, mReceiver)
}
