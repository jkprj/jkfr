package histogram

import (
	"jkfr/prometheus/utils"

	"github.com/prometheus/client_golang/prometheus"
)

var HistogramCollector *utils.UCollector = utils.NewUCollector(newHistogramVec)

func newHistogramVec(name string, labelNames []string) prometheus.Collector {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: name,
			Help: name,
		},
		labelNames)
}

func GetHistogramVec(name string, labelNames []string) *prometheus.HistogramVec {
	return HistogramCollector.GetCollector(name, labelNames).(*prometheus.HistogramVec)
}

func GetObserver(name string, labels map[string]string) prometheus.Observer {
	return GetHistogramVec(name, utils.GetLabels(labels)).With(prometheus.Labels(labels))
}

func Observe(name string, labels map[string]string, val float64) {
	GetObserver(name, labels).Observe(val)
}
