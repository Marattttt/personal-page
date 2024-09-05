// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.3
// source: gorunner.proto

package grpcgen

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	GoRunner_RunCode_FullMethodName = "/GoRunner/RunCode"
)

// GoRunnerClient is the client API for GoRunner service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// The GoRunner service definition
type GoRunnerClient interface {
	// RunCode runs the submitted code and returns the result
	RunCode(ctx context.Context, in *RunCodeRequest, opts ...grpc.CallOption) (*RunCodeResponse, error)
}

type goRunnerClient struct {
	cc grpc.ClientConnInterface
}

func NewGoRunnerClient(cc grpc.ClientConnInterface) GoRunnerClient {
	return &goRunnerClient{cc}
}

func (c *goRunnerClient) RunCode(ctx context.Context, in *RunCodeRequest, opts ...grpc.CallOption) (*RunCodeResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RunCodeResponse)
	err := c.cc.Invoke(ctx, GoRunner_RunCode_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GoRunnerServer is the server API for GoRunner service.
// All implementations must embed UnimplementedGoRunnerServer
// for forward compatibility.
//
// The GoRunner service definition
type GoRunnerServer interface {
	// RunCode runs the submitted code and returns the result
	RunCode(context.Context, *RunCodeRequest) (*RunCodeResponse, error)
	mustEmbedUnimplementedGoRunnerServer()
}

// UnimplementedGoRunnerServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedGoRunnerServer struct{}

func (UnimplementedGoRunnerServer) RunCode(context.Context, *RunCodeRequest) (*RunCodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RunCode not implemented")
}
func (UnimplementedGoRunnerServer) mustEmbedUnimplementedGoRunnerServer() {}
func (UnimplementedGoRunnerServer) testEmbeddedByValue()                  {}

// UnsafeGoRunnerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GoRunnerServer will
// result in compilation errors.
type UnsafeGoRunnerServer interface {
	mustEmbedUnimplementedGoRunnerServer()
}

func RegisterGoRunnerServer(s grpc.ServiceRegistrar, srv GoRunnerServer) {
	// If the following call pancis, it indicates UnimplementedGoRunnerServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&GoRunner_ServiceDesc, srv)
}

func _GoRunner_RunCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RunCodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoRunnerServer).RunCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GoRunner_RunCode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoRunnerServer).RunCode(ctx, req.(*RunCodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GoRunner_ServiceDesc is the grpc.ServiceDesc for GoRunner service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GoRunner_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "GoRunner",
	HandlerType: (*GoRunnerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RunCode",
			Handler:    _GoRunner_RunCode_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "gorunner.proto",
}
