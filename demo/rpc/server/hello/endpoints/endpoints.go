package endpoints

import (
	"context"
	"fmt"

	"jkfr/demo/rpc/server/hello"

	"github.com/go-kit/kit/endpoint"
)

type Hello struct {
	helloEndpoint     endpoint.Endpoint
	howAreYouEndpoint endpoint.Endpoint
	whatNameEndpoint  endpoint.Endpoint
}

func NewService() *Hello {
	h := hello.NewHelloRpc()

	eps := new(Hello)
	eps.helloEndpoint = makeHelloEndPoint(h)
	eps.howAreYouEndpoint = makeHowAreYouEndPoint(h)
	eps.whatNameEndpoint = makeWhatNameEndPoint(h)

	return eps
}

func (e *Hello) Hello(request hello.URequest, response *hello.URespone) error {

	resp, err := e.helloEndpoint(context.Background(), request)
	if nil != err {
		return err
	}

	*response = resp.(hello.URespone)

	return nil
}

func (e *Hello) HowAreYou(request hello.URequest, response *hello.URespone) error {

	resp, err := e.howAreYouEndpoint(context.Background(), request)
	if nil != err {
		return err
	}

	*response = resp.(hello.URespone)

	return nil
}

func (e *Hello) WhatName(request hello.URequest, response *hello.URespone) (err error) {

	resp, err := e.whatNameEndpoint(context.Background(), request)
	if nil != err {
		return err
	}

	*response = resp.(hello.URespone)

	return nil
}
func (e *Hello) WrapAllLabeledExcept(middleware func(string, endpoint.Endpoint) endpoint.Endpoint, excluded ...string) {
	included := map[string]struct{}{
		"Hello":     struct{}{},
		"HowAreYou": struct{}{},
		"WhatName":  struct{}{},
	}

	for _, ex := range excluded {
		if _, ok := included[ex]; !ok {
			panic(fmt.Sprintf("Excluded endpoint '%s' does not exist; see middlewares/endpoints.go", ex))
		}
		delete(included, ex)
	}

	for inc, _ := range included {
		if inc == "Hello" {
			e.helloEndpoint = middleware("Hello", e.helloEndpoint)
		}
		if inc == "HowAreYou" {
			e.howAreYouEndpoint = middleware("HowAreYou", e.howAreYouEndpoint)
		}
		if inc == "WhatName" {
			e.whatNameEndpoint = middleware("WhatName", e.whatNameEndpoint)
		}
	}
}

func makeHelloEndPoint(hl *hello.HelloRpc) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		resp := hello.URespone{}
		err = hl.Hello(request.(hello.URequest), &resp)

		return resp, err
	}
}

func makeHowAreYouEndPoint(hl *hello.HelloRpc) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		resp := hello.URespone{}
		err = hl.HowAreYou(request.(hello.URequest), &resp)

		return resp, err
	}
}

func makeWhatNameEndPoint(hl *hello.HelloRpc) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		resp := hello.URespone{}
		err = hl.WhatName(request.(hello.URequest), &resp)

		return resp, err
	}
}
