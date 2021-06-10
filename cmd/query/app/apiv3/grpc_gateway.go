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

package apiv3

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"github.com/jaegertracing/jaeger/proto-gen/api_v3"
)

// RegisterGRPCGateway registers api_v3 endpoints into provided mux.
func RegisterGRPCGateway(r *mux.Router, basePath string, grpcEndpoint string) error {
	jsonpb := &runtime.JSONPb{}
	grpcGatewayMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, jsonpb),
	)
	r.PathPrefix("/v3/").Handler(http.StripPrefix(basePath, grpcGatewayMux))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	return api_v3.RegisterQueryServiceHandlerFromEndpoint(context.Background(), grpcGatewayMux, grpcEndpoint, opts)
}
