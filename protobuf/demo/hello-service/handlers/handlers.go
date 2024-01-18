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

func (s helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	var resp pb.HelloReply
	return &resp, nil
}

func (s helloService) GetPersons(ctx context.Context, in *pb.PersonRequest) (*pb.PersonsReply, error) {
	var resp pb.PersonsReply
	return &resp, nil
}

func (s helloService) GetPersonByName(ctx context.Context, in *pb.HelloRequest) (*pb.Person, error) {
	var resp pb.Person
	return &resp, nil
}
