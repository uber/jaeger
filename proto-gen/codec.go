package proto_gen

import (
	"reflect"
	"strings"

	gogoproto "github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"
)

const jaegerProtoGenPkgPath = "github.com/jaegertracing/jaeger/proto-gen"
const jaegerModelPkgPath = "github.com/jaegertracing/jaeger/model"

func init() {
	encoding.RegisterCodec(newCodec())
}

// gogoCodec forces the use of gogo proto marshalling/unmarshalling for
// Jaeger proto types (package jaeger/gen-proto).
type gogoCodec struct {
}

var _ encoding.Codec = (*gogoCodec)(nil)

func newCodec() *gogoCodec {
	return &gogoCodec{}
}

// Name implements encoding.Codec
func (c *gogoCodec) Name() string {
	return "proto"
}

// Marshal implements encoding.Codec
func (c *gogoCodec) Marshal(v interface{}) ([]byte, error) {
	t := reflect.TypeOf(v)
	elem := t.Elem()
	// use gogo proto only for Jaeger types
	if useGogo(elem) {
		return gogoproto.Marshal(v.(gogoproto.Message))
	}
	return proto.Marshal(v.(proto.Message))
}

// Unmarshal implements encoding.Codec
func (c *gogoCodec) Unmarshal(data []byte, v interface{}) error {
	t := reflect.TypeOf(v)
	elem := t.Elem()
	// use gogo proto only for Jaeger types
	if useGogo(elem) {
		return gogoproto.Unmarshal(data, v.(gogoproto.Message))
	}
	return proto.Unmarshal(data, v.(proto.Message))
}

func useGogo(t reflect.Type) bool {
	return t!= nil && (strings.HasPrefix(t.PkgPath(), jaegerProtoGenPkgPath) || strings.HasPrefix(t.PkgPath(), jaegerModelPkgPath))
}
