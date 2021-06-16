// Copyright (c) 2021 The Jaeger Authors.
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
	"encoding/json"
	"fmt"
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

type jsonMarshaler interface {
	marshal(writer io.Writer, response interface{}) error
}

// protoJSONMarshaler is a protobuf-friendly JSON marshaler that knows how to handle protobuf-specific
// field types such as "oneof" as well as dealing with NaNs which are not supported by JSON.
type protoJSONMarshaler struct {
	marshaler *jsonpb.Marshaler
}

// structJSONMarshaler uses the built-in encoding/json package for marshaling basic structs to JSON.
type structJSONMarshaler struct {
	marshaler func(v interface{}) ([]byte, error)
}

func newProtoJSONMarshaler(prettyPrint bool) jsonMarshaler {
	marshaler := new(jsonpb.Marshaler)
	if prettyPrint {
		marshaler.Indent = prettyPrintIndent
	}
	return &protoJSONMarshaler{
		marshaler: marshaler,
	}
}

func newStructJSONMarshaler(prettyPrint bool) jsonMarshaler {
	marshaler := json.Marshal
	if prettyPrint {
		marshaler = func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", prettyPrintIndent)
		}
	}
	return &structJSONMarshaler{
		marshaler: marshaler,
	}
}

func (pm *protoJSONMarshaler) marshal(w io.Writer, response interface{}) error {
	return pm.marshaler.Marshal(w, response.(proto.Message))
}

func (sm *structJSONMarshaler) marshal(w io.Writer, response interface{}) error {
	resp, err := sm.marshaler(response)
	if err != nil {
		return fmt.Errorf("failed marshalling HTTP response to JSON: %w", err)
	}
	_, err = w.Write(resp)
	return err
}
