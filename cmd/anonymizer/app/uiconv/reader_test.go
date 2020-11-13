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

package uiconv

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const spanData = `[{"traceID":"2be38093ead7a083","spanID":"7606ddfe69932d34","flags":1,"operationName":"a071653098f9250d","references":[{"refType":"CHILD_OF","traceID":"2be38093ead7a083","spanID":"492770a15935810f"}],"startTime":1605223981761425,"duration":267037,"tags":[{"key":"span.kind","type":"string","value":"server"}],"logs":[],"process":{"serviceName":"16af988c443cff37","tags":[]},"warnings":null},
{"traceID":"2be38093ead7a083","spanID":"7bd66f09ba90ea3d","flags":1,"operationName":"471418097747d04a","references":[{"refType":"CHILD_OF","traceID":"2be38093ead7a083","spanID":"7606ddfe69932d34"}],"startTime":1605223981965074,"duration":32782,"tags":[{"key":"span.kind","type":"string","value":"client"},{"key":"error","type":"bool","value":"true"}],"logs":[],"process":{"serviceName":"3c220036602f839e","tags":[]},"warnings":null}
]`

func TestReader(t *testing.T) {
	f, err := ioutil.TempFile("", "captured-spans.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.Write([]byte(spanData))
	require.NoError(t, err)

	r, err := NewReader(
		f.Name(),
		zap.NewNop(),
	)
	require.NoError(t, err)

	s1, err := r.NextSpan()
	require.NoError(t, err)
	assert.Equal(t, "a071653098f9250d", s1.OperationName)

	s2, err := r.NextSpan()
	require.NoError(t, err)
	assert.Equal(t, "471418097747d04a", s2.OperationName)

	_, err = r.NextSpan()
	require.Equal(t, io.EOF, err)
}
