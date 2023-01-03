package transport

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/kit/sd/lb"

	"github.com/go-kit/kit/endpoint"
)

type RetryError struct {
	RawErrors []error // all errors encountered from endpoints directly
	Final     error   // the final, terminating error
}

func (e RetryError) Error() string {
	var suffix string
	if len(e.RawErrors) > 1 {
		a := make([]string, len(e.RawErrors)-1)
		for i := 0; i < len(e.RawErrors)-1; i++ { // last one is Final
			a[i] = e.RawErrors[i].Error()
		}
		suffix = fmt.Sprintf(" (previously: %s)", strings.Join(a, "; "))
	}
	return fmt.Sprintf("%v%s", e.Final, suffix)
}

type Callback func(n int, received error) (keepTrying bool, replacement error)

func Retry(max int, intervalMS int, b lb.Balancer) endpoint.Endpoint {
	return RetryWithCallback(b, max, intervalMS)
}

func RetryWithCallback(b lb.Balancer, max, intervalMS int) endpoint.Endpoint {
	if b == nil {
		panic("nil Balancer")
	}

	return func(ctx context.Context, request interface{}) (response interface{}, err error) {

		var (
			final RetryError
			e     endpoint.Endpoint
		)

		for i := 1; ; i++ {

			if e, err = b.Endpoint(); err != nil {
				final.RawErrors = append(final.RawErrors, err)
				if max > i {
					time.Sleep(time.Duration(intervalMS) * time.Millisecond)
					continue
				} else {
					return nil, final
				}
			}

			response, err := e(ctx, request)

			if err != nil {
				final.RawErrors = append(final.RawErrors, err)
				if max > i {
					time.Sleep(time.Duration(intervalMS) * time.Millisecond)
					continue
				} else {
					final.Final = err
					return nil, final
				}
			} else {
				return response, nil
			}

		}
	}
}
