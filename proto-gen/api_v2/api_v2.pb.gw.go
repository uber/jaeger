// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: api_v2.proto

/*
Package api_v2 is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package api_v2

import (
	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray

func request_CollectorService_PostSpans_0(ctx context.Context, marshaler runtime.Marshaler, client CollectorServiceClient, req *http.Request, pathParams map[string]string) (CollectorService_PostSpansClient, runtime.ServerMetadata, error) {
	var metadata runtime.ServerMetadata
	stream, err := client.PostSpans(ctx)
	if err != nil {
		grpclog.Printf("Failed to start streaming: %v", err)
		return nil, metadata, err
	}
	dec := marshaler.NewDecoder(req.Body)
	handleSend := func() error {
		var protoReq PostSpansRequest
		err = dec.Decode(&protoReq)
		if err == io.EOF {
			return err
		}
		if err != nil {
			grpclog.Printf("Failed to decode request: %v", err)
			return err
		}
		if err = stream.Send(&protoReq); err != nil {
			grpclog.Printf("Failed to send request: %v", err)
			return err
		}
		return nil
	}
	if err := handleSend(); err != nil {
		if cerr := stream.CloseSend(); cerr != nil {
			grpclog.Printf("Failed to terminate client stream: %v", cerr)
		}
		if err == io.EOF {
			return stream, metadata, nil
		}
		return nil, metadata, err
	}
	go func() {
		for {
			if err := handleSend(); err != nil {
				break
			}
		}
		if err := stream.CloseSend(); err != nil {
			grpclog.Printf("Failed to terminate client stream: %v", err)
		}
	}()
	header, err := stream.Header()
	if err != nil {
		grpclog.Printf("Failed to get header from client: %v", err)
		return nil, metadata, err
	}
	metadata.HeaderMD = header
	return stream, metadata, nil
}

func request_SamplingManager_GetSamplingStrategy_0(ctx context.Context, marshaler runtime.Marshaler, client SamplingManagerClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq SamplingStrategyParameters
	var metadata runtime.ServerMetadata

	if req.ContentLength > 0 {
		if err := marshaler.NewDecoder(req.Body).Decode(&protoReq); err != nil {
			return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
		}
	}

	msg, err := client.GetSamplingStrategy(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

// RegisterCollectorServiceHandlerFromEndpoint is same as RegisterCollectorServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterCollectorServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Printf("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Printf("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterCollectorServiceHandler(ctx, mux, conn)
}

// RegisterCollectorServiceHandler registers the http handlers for service CollectorService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterCollectorServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterCollectorServiceHandlerClient(ctx, mux, NewCollectorServiceClient(conn))
}

// RegisterCollectorServiceHandler registers the http handlers for service CollectorService to "mux".
// The handlers forward requests to the grpc endpoint over the given implementation of "CollectorServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "CollectorServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "CollectorServiceClient" to call the correct interceptors.
func RegisterCollectorServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client CollectorServiceClient) error {

	mux.Handle("POST", pattern_CollectorService_PostSpans_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		if cn, ok := w.(http.CloseNotifier); ok {
			go func(done <-chan struct{}, closed <-chan bool) {
				select {
				case <-done:
				case <-closed:
					cancel()
				}
			}(ctx.Done(), cn.CloseNotify())
		}
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_CollectorService_PostSpans_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_CollectorService_PostSpans_0(ctx, mux, outboundMarshaler, w, req, func() (proto.Message, error) { return resp.Recv() }, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_CollectorService_PostSpans_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"api", "v2", "spans"}, ""))
)

var (
	forward_CollectorService_PostSpans_0 = runtime.ForwardResponseStream
)

// RegisterSamplingManagerHandlerFromEndpoint is same as RegisterSamplingManagerHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterSamplingManagerHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Printf("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Printf("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterSamplingManagerHandler(ctx, mux, conn)
}

// RegisterSamplingManagerHandler registers the http handlers for service SamplingManager to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterSamplingManagerHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterSamplingManagerHandlerClient(ctx, mux, NewSamplingManagerClient(conn))
}

// RegisterSamplingManagerHandler registers the http handlers for service SamplingManager to "mux".
// The handlers forward requests to the grpc endpoint over the given implementation of "SamplingManagerClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "SamplingManagerClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "SamplingManagerClient" to call the correct interceptors.
func RegisterSamplingManagerHandlerClient(ctx context.Context, mux *runtime.ServeMux, client SamplingManagerClient) error {

	mux.Handle("POST", pattern_SamplingManager_GetSamplingStrategy_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		if cn, ok := w.(http.CloseNotifier); ok {
			go func(done <-chan struct{}, closed <-chan bool) {
				select {
				case <-done:
				case <-closed:
					cancel()
				}
			}(ctx.Done(), cn.CloseNotify())
		}
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_SamplingManager_GetSamplingStrategy_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_SamplingManager_GetSamplingStrategy_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_SamplingManager_GetSamplingStrategy_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"api", "v2", "samplingStrategy"}, ""))
)

var (
	forward_SamplingManager_GetSamplingStrategy_0 = runtime.ForwardResponseMessage
)
