// Code generated by truss. DO NOT EDIT.
// Rerunning truss will overwrite this file.
// Version: a2b01cac16
// Version Date: Thu Oct 20 18:44:52 UTC 2022

// Package grpc provides a gRPC client for the Hello service.
package grpc

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/endpoint"
	grpctransport "github.com/go-kit/kit/transport/grpc"

	// This Service
	pb "github.com/jkprj/jkfr/protobuf/demo"
	"github.com/jkprj/jkfr/protobuf/demo/hello-service/svc"
)

// New returns an service backed by a gRPC client connection. It is the
// responsibility of the caller to dial, and later close, the connection.
func New(conn *grpc.ClientConn, options ...ClientOption) (pb.HelloServer, error) {
	var cc clientConfig

	for _, f := range options {
		err := f(&cc)
		if err != nil {
			return nil, errors.Wrap(err, "cannot apply option")
		}
	}

	clientOptions := []grpctransport.ClientOption{
		grpctransport.ClientBefore(
			contextValuesToGRPCMetadata(cc.headers)),
	}
	var sayhelloEndpoint endpoint.Endpoint
	{
		sayhelloEndpoint = grpctransport.NewClient(
			conn,
			"proto.Hello",
			"SayHello",
			EncodeGRPCSayHelloRequest,
			DecodeGRPCSayHelloResponse,
			pb.HelloReply{},
			clientOptions...,
		).Endpoint()
	}

	var getpersonsEndpoint endpoint.Endpoint
	{
		getpersonsEndpoint = grpctransport.NewClient(
			conn,
			"proto.Hello",
			"GetPersons",
			EncodeGRPCGetPersonsRequest,
			DecodeGRPCGetPersonsResponse,
			pb.PersonsReply{},
			clientOptions...,
		).Endpoint()
	}

	var getpersonbynameEndpoint endpoint.Endpoint
	{
		getpersonbynameEndpoint = grpctransport.NewClient(
			conn,
			"proto.Hello",
			"GetPersonByName",
			EncodeGRPCGetPersonByNameRequest,
			DecodeGRPCGetPersonByNameResponse,
			pb.Person{},
			clientOptions...,
		).Endpoint()
	}

	return svc.Endpoints{
		SayHelloEndpoint:        sayhelloEndpoint,
		GetPersonsEndpoint:      getpersonsEndpoint,
		GetPersonByNameEndpoint: getpersonbynameEndpoint,
	}, nil
}

// GRPC Client Decode

// DecodeGRPCSayHelloResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC sayhello reply to a user-domain sayhello response. Primarily useful in a client.
func DecodeGRPCSayHelloResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.HelloReply)
	return reply, nil
}

// DecodeGRPCGetPersonsResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC getpersons reply to a user-domain getpersons response. Primarily useful in a client.
func DecodeGRPCGetPersonsResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.PersonsReply)
	return reply, nil
}

// DecodeGRPCGetPersonByNameResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC getpersonbyname reply to a user-domain getpersonbyname response. Primarily useful in a client.
func DecodeGRPCGetPersonByNameResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.Person)
	return reply, nil
}

// GRPC Client Encode

// EncodeGRPCSayHelloRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain sayhello request to a gRPC sayhello request. Primarily useful in a client.
func EncodeGRPCSayHelloRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.HelloRequest)
	return req, nil
}

// EncodeGRPCGetPersonsRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain getpersons request to a gRPC getpersons request. Primarily useful in a client.
func EncodeGRPCGetPersonsRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.PersonRequest)
	return req, nil
}

// EncodeGRPCGetPersonByNameRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain getpersonbyname request to a gRPC getpersonbyname request. Primarily useful in a client.
func EncodeGRPCGetPersonByNameRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.HelloRequest)
	return req, nil
}

type clientConfig struct {
	headers []string
}

// ClientOption is a function that modifies the client config
type ClientOption func(*clientConfig) error

func CtxValuesToSend(keys ...string) ClientOption {
	return func(o *clientConfig) error {
		o.headers = keys
		return nil
	}
}

func contextValuesToGRPCMetadata(keys []string) grpctransport.ClientRequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		var pairs []string
		for _, k := range keys {
			if v, ok := ctx.Value(k).(string); ok {
				pairs = append(pairs, k, v)
			}
		}

		if pairs != nil {
			*md = metadata.Join(*md, metadata.Pairs(pairs...))
		}

		return ctx
	}
}
