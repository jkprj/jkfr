package handlers

import (
	"context"

	pb "github.com/jkprj/jkfr/protobuf/demo"
)

// NewService returns a naïve, stateless implementation of Service.
func NewService() pb.HelloServer {
	return helloService{}
}

type helloService struct{}

// SayHello implements Service.
func (s helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	var resp pb.HelloReply
	resp = pb.HelloReply{
		// Message:
	}
	return &resp, nil
}

// GetPersons implements Service.
func (s helloService) GetPersons(ctx context.Context, in *pb.PersonRequest) (*pb.PersonsReply, error) {
	var resp pb.PersonsReply
	resp = pb.PersonsReply{
		// Persons:
	}
	return &resp, nil
}

// GetPersonByName implements Service.
func (s helloService) GetPersonByName(ctx context.Context, in *pb.HelloRequest) (*pb.Person, error) {
	var resp pb.Person
	resp = pb.Person{
		// Name:
		// Sex:
		// Age:
	}
	return &resp, nil
}
