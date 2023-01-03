// Code generated by truss. DO NOT EDIT.
// Rerunning truss will overwrite this file.
// Version: 8907ffca23
// Version Date: Wed Nov 27 21:28:21 UTC 2019

package svc

// This file contains methods to make individual endpoints from services,
// request and response types to serve those endpoints, as well as encoders and
// decoders for those types, for all of our supported transport serialization
// formats.

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"

	pb "jkfr/protobuf/demo"
)

// Endpoints collects all of the endpoints that compose an add service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
//
// In a server, it's useful for functions that need to operate on a per-endpoint
// basis. For example, you might pass an Endpoints to a function that produces
// an http.Handler, with each method (endpoint) wired up to a specific path. (It
// is probably a mistake in design to invoke the Service methods on the
// Endpoints struct in a server.)
//
// In a client, it's useful to collect individually constructed endpoints into a
// single type that implements the Service interface. For example, you might
// construct individual endpoints using transport/http.NewClient, combine them into an Endpoints, and return it to the caller as a Service.
type Endpoints struct {
	SayHelloEndpoint        endpoint.Endpoint
	GetPersonsEndpoint      endpoint.Endpoint
	GetPersonByNameEndpoint endpoint.Endpoint
}

// Endpoints

func (e Endpoints) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	response, err := e.SayHelloEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.HelloReply), nil
}

func (e Endpoints) GetPersons(ctx context.Context, in *pb.PersonRequest) (*pb.PersonsReply, error) {
	response, err := e.GetPersonsEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.PersonsReply), nil
}

func (e Endpoints) GetPersonByName(ctx context.Context, in *pb.HelloRequest) (*pb.Person, error) {
	response, err := e.GetPersonByNameEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.Person), nil
}

// Make Endpoints

func MakeSayHelloEndpoint(s pb.HelloServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.HelloRequest)
		v, err := s.SayHello(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeGetPersonsEndpoint(s pb.HelloServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.PersonRequest)
		v, err := s.GetPersons(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeGetPersonByNameEndpoint(s pb.HelloServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.HelloRequest)
		v, err := s.GetPersonByName(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

// WrapAllExcept wraps each Endpoint field of struct Endpoints with a
// go-kit/kit/endpoint.Middleware.
// Use this for applying a set of middlewares to every endpoint in the service.
// Optionally, endpoints can be passed in by name to be excluded from being wrapped.
// WrapAllExcept(middleware, "Status", "Ping")
func (e *Endpoints) WrapAllExcept(middleware endpoint.Middleware, excluded ...string) {
	included := map[string]struct{}{
		"SayHello":        struct{}{},
		"GetPersons":      struct{}{},
		"GetPersonByName": struct{}{},
	}

	for _, ex := range excluded {
		if _, ok := included[ex]; !ok {
			panic(fmt.Sprintf("Excluded endpoint '%s' does not exist; see middlewares/endpoints.go", ex))
		}
		delete(included, ex)
	}

	for inc, _ := range included {
		if inc == "SayHello" {
			e.SayHelloEndpoint = middleware(e.SayHelloEndpoint)
		}
		if inc == "GetPersons" {
			e.GetPersonsEndpoint = middleware(e.GetPersonsEndpoint)
		}
		if inc == "GetPersonByName" {
			e.GetPersonByNameEndpoint = middleware(e.GetPersonByNameEndpoint)
		}
	}
}

// LabeledMiddleware will get passed the endpoint name when passed to
// WrapAllLabeledExcept, this can be used to write a generic metrics
// middleware which can send the endpoint name to the metrics collector.
type LabeledMiddleware func(string, endpoint.Endpoint) endpoint.Endpoint

// WrapAllLabeledExcept wraps each Endpoint field of struct Endpoints with a
// LabeledMiddleware, which will receive the name of the endpoint. See
// LabeldMiddleware. See method WrapAllExept for details on excluded
// functionality.
func (e *Endpoints) WrapAllLabeledExcept(middleware func(string, endpoint.Endpoint) endpoint.Endpoint, excluded ...string) {
	included := map[string]struct{}{
		"SayHello":        struct{}{},
		"GetPersons":      struct{}{},
		"GetPersonByName": struct{}{},
	}

	for _, ex := range excluded {
		if _, ok := included[ex]; !ok {
			panic(fmt.Sprintf("Excluded endpoint '%s' does not exist; see middlewares/endpoints.go", ex))
		}
		delete(included, ex)
	}

	for inc, _ := range included {
		if inc == "SayHello" {
			e.SayHelloEndpoint = middleware("SayHello", e.SayHelloEndpoint)
		}
		if inc == "GetPersons" {
			e.GetPersonsEndpoint = middleware("GetPersons", e.GetPersonsEndpoint)
		}
		if inc == "GetPersonByName" {
			e.GetPersonByNameEndpoint = middleware("GetPersonByName", e.GetPersonByNameEndpoint)
		}
	}
}
