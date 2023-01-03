package prometheus

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run(servAddr string) error {
	if len(servAddr) == 0 || !strings.Contains(servAddr, ":") {
		errorInfo := fmt.Sprintf("servAddr is invalid")
		return errors.New(errorInfo)
	}

	// Can customize the http handle function

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(servAddr, nil)
}
