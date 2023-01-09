package endpoint

import (
	"context"
	"math"
	"time"

	uprometheus "github.com/jkprj/jkfr/gokit/prometheus"
	jklog "github.com/jkprj/jkfr/log"
	jkos "github.com/jkprj/jkfr/os"
	ucounter "github.com/jkprj/jkfr/prometheus/counter"
	uhistogram "github.com/jkprj/jkfr/prometheus/histogram"

	"golang.org/x/time/rate"

	"github.com/go-kit/kit/endpoint"
)

type ActionMiddleware func(action string, next endpoint.Endpoint) endpoint.Endpoint

func Chain(next endpoint.Endpoint, action string, middlewares ...ActionMiddleware) endpoint.Endpoint {

	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i](action, next)
	}

	return next
}

func DefaultMiddleware(prometheusNameSpace, role string, rateLimit rate.Limit) []ActionMiddleware {
	return []ActionMiddleware{
		MakePrometheusResponeTimeMiddleware(prometheusNameSpace, role),
		MakePrometheusActionRequestMiddleware(prometheusNameSpace, role),
		MakePrometheusTotalRequestMiddleware(prometheusNameSpace, role),
		MakePrometheusTotalRequestErrorMiddleware(prometheusNameSpace, role),
		MakePrometheusActionRequestErrorMiddleware(prometheusNameSpace, role),
		MakeRateLimitMiddleware(rateLimit),
	}
}

func MakeRateLimitMiddleware(limit rate.Limit) ActionMiddleware {

	rateLimiter := rate.NewLimiter(limit, int(math.Min(float64(limit), 10000)))

	return func(action string, next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if limit > 0 {
				rateLimiter.Wait(ctx)
			}
			return next(ctx, request)
		}
	}
}

func MakePrometheusResponeTimeMiddleware(nameSpace, role string) ActionMiddleware {
	return func(action string, next endpoint.Endpoint) endpoint.Endpoint {

		if "" == action {
			return next
		}

		labels := map[string]string{"APP": jkos.AppName(), "Role": role}
		observer := uhistogram.GetObserver(nameSpace+"_Action_ResponeTime", labels)

		return func(ctx context.Context, request interface{}) (response interface{}, err error) {

			bg := time.Now()

			response, err = next(ctx, request)

			if uprometheus.Running {
				longTime := time.Since(bg).Seconds()
				observer.Observe(longTime)
				// jklog.Debugw("Prometheus Respone Time", "action", action, "use_time", longTime)
			}

			return response, err
		}

	}
}

func MakePrometheusActionRequestMiddleware(nameSpace, role string) ActionMiddleware {
	return func(action string, next endpoint.Endpoint) endpoint.Endpoint {

		if "" == action {
			return next
		}

		labels := map[string]string{"APP": jkos.AppName(), "Role": role, "Action": action}
		counter := ucounter.GetCounter(nameSpace+"_Action", labels)

		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if uprometheus.Running {
				counter.Inc()
			}
			// jklog.Debugw("Prometheus Action Request", "action", action)

			return next(ctx, request)
		}

	}
}

func MakePrometheusTotalRequestMiddleware(nameSpace, role string) ActionMiddleware {
	return func(action string, next endpoint.Endpoint) endpoint.Endpoint {

		labels := map[string]string{"APP": jkos.AppName(), "Role": role}
		counter := ucounter.GetCounter(nameSpace+"_Request_Total", labels)

		return func(ctx context.Context, request interface{}) (response interface{}, err error) {

			if uprometheus.Running {
				counter.Inc()
			}

			return next(ctx, request)
		}
	}

}

func MakePrometheusActionRequestErrorMiddleware(nameSpace, role string) ActionMiddleware {
	return func(action string, next endpoint.Endpoint) endpoint.Endpoint {

		if "" == action {
			return next
		}

		labels := map[string]string{"APP": jkos.AppName(), "Role": role, "Action": action}
		counter := ucounter.GetCounter(nameSpace+"_Action_Error", labels)

		return func(ctx context.Context, request interface{}) (response interface{}, err error) {

			response, err = next(ctx, request)
			if nil != err {
				if uprometheus.Running {
					counter.Inc()
				}
				jklog.Debugw("Prometheus Action Request error", "action", action)
			}

			return response, err
		}

	}
}

func MakePrometheusTotalRequestErrorMiddleware(nameSpace, role string) ActionMiddleware {
	return func(action string, next endpoint.Endpoint) endpoint.Endpoint {

		labels := map[string]string{"APP": jkos.AppName(), "Role": role}
		counter := ucounter.GetCounter(nameSpace+"_Request_Error_Total", labels)

		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			response, err = next(ctx, request)
			if nil != err {
				if uprometheus.Running {
					counter.Inc()
				}
			}

			return response, err
		}
	}

}
