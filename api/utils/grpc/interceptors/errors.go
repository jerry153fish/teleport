// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interceptors

import (
	"context"

	"github.com/gravitational/trace"
	"github.com/gravitational/trace/trail"
	"google.golang.org/grpc"
)

// grpcServerStreamWrapper wraps around the embedded grpc.ServerStream
// and intercepts the RecvMsg and SendMsg method calls converting errors
// to the appropriate gRPC status error.
type grpcServerStreamWrapper struct {
	grpc.ServerStream
}

// SendMsg wraps around ServerStream.SendMsg and adds metrics reporting
func (s *grpcServerStreamWrapper) SendMsg(m interface{}) error {
	return trace.Unwrap(trail.FromGRPC(s.ServerStream.SendMsg(m)))
}

// RecvMsg wraps around ServerStream.RecvMsg and adds metrics reporting
func (s *grpcServerStreamWrapper) RecvMsg(m interface{}) error {
	return trace.Unwrap(trail.FromGRPC(s.ServerStream.RecvMsg(m)))
}

// grpcClientStreamWrapper wraps around the embedded grpc.ClientStream
// and intercepts the RecvMsg and SendMsg method calls converting errors
// to the appropriate gRPC status error.
type grpcClientStreamWrapper struct {
	grpc.ClientStream
}

// SendMsg wraps around ClientStream.SendMsg
func (s *grpcClientStreamWrapper) SendMsg(m interface{}) error {
	return trace.Unwrap(trail.FromGRPC(s.ClientStream.SendMsg(m)))
}

// RecvMsg wraps around ClientStream.RecvMsg
func (s *grpcClientStreamWrapper) RecvMsg(m interface{}) error {
	return trace.Unwrap(trail.FromGRPC(s.ClientStream.RecvMsg(m)))
}

// GRPCServerUnaryErrorInterceptor is a gRPC unary server interceptor that
// handles converting errors to the appropriate gRPC status error.
func GRPCServerUnaryErrorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	return resp, trace.Unwrap(trail.ToGRPC(err))
}

// GRPCClientUnaryErrorInterceptor is a gRPC unary client interceptor that
// handles converting errors to the appropriate grpc status error.
func GRPCClientUnaryErrorInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return trace.Unwrap(trail.FromGRPC(invoker(ctx, method, req, reply, cc, opts...)))
}

// GRPCServerStreamErrorInterceptor is a gRPC server stream interceptor that
// handles converting errors to the appropriate gRPC status error.
func GRPCServerStreamErrorInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	serverWrapper := &grpcServerStreamWrapper{ss}
	return trace.Unwrap(trail.ToGRPC(handler(srv, serverWrapper)))
}

// GRPCClientStreamErrorInterceptor is gRPC client stream interceptor that
// handles converting errors to the appropriate gRPC status error.
func GRPCClientStreamErrorInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, trace.Unwrap(trail.ToGRPC(err))
	}
	return &grpcClientStreamWrapper{s}, nil
}
