// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.1
// source: poster.proto

package poster

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
	Poster_SendSummary_FullMethodName = "/flight.Poster/SendSummary"
)

// PosterClient is the client API for Poster service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PosterClient interface {
	SendSummary(ctx context.Context, in *SendSummaryRequest, opts ...grpc.CallOption) (*SendSummaryResponse, error)
}

type posterClient struct {
	cc grpc.ClientConnInterface
}

func NewPosterClient(cc grpc.ClientConnInterface) PosterClient {
	return &posterClient{cc}
}

func (c *posterClient) SendSummary(ctx context.Context, in *SendSummaryRequest, opts ...grpc.CallOption) (*SendSummaryResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SendSummaryResponse)
	err := c.cc.Invoke(ctx, Poster_SendSummary_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PosterServer is the server API for Poster service.
// All implementations must embed UnimplementedPosterServer
// for forward compatibility.
type PosterServer interface {
	SendSummary(context.Context, *SendSummaryRequest) (*SendSummaryResponse, error)
	mustEmbedUnimplementedPosterServer()
}

// UnimplementedPosterServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedPosterServer struct{}

func (UnimplementedPosterServer) SendSummary(context.Context, *SendSummaryRequest) (*SendSummaryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendSummary not implemented")
}
func (UnimplementedPosterServer) mustEmbedUnimplementedPosterServer() {}
func (UnimplementedPosterServer) testEmbeddedByValue()                {}

// UnsafePosterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PosterServer will
// result in compilation errors.
type UnsafePosterServer interface {
	mustEmbedUnimplementedPosterServer()
}

func RegisterPosterServer(s grpc.ServiceRegistrar, srv PosterServer) {
	// If the following call pancis, it indicates UnimplementedPosterServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Poster_ServiceDesc, srv)
}

func _Poster_SendSummary_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendSummaryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PosterServer).SendSummary(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Poster_SendSummary_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PosterServer).SendSummary(ctx, req.(*SendSummaryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Poster_ServiceDesc is the grpc.ServiceDesc for Poster service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Poster_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "flight.Poster",
	HandlerType: (*PosterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendSummary",
			Handler:    _Poster_SendSummary_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "poster.proto",
}
