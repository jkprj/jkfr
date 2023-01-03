package counter

import (
	"jkfr/prometheus/utils"

	"github.com/prometheus/client_golang/prometheus"
)

var CounterCollector *utils.UCollector = utils.NewUCollector(newCounterVec)

func newCounterVec(name string, labelNames []string) prometheus.Collector {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: name,
		},
		labelNames)
}

func GetCounterVec(name string, labelNames []string) *prometheus.CounterVec {
	return CounterCollector.GetCollector(name, labelNames).(*prometheus.CounterVec)
}

func GetCounter(name string, labels map[string]string) prometheus.Counter {
	return GetCounterVec(name, utils.GetLabels(labels)).With(prometheus.Labels(labels))
}

func Inc(name string, labels map[string]string) {
	GetCounter(name, labels).Inc()
}

func Add(name string, labels map[string]string, val float64) {
	GetCounter(name, labels).Add(val)
}
