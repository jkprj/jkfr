package gauge

import (
	"jkfr/prometheus/utils"

	"github.com/prometheus/client_golang/prometheus"
)

var GaugeCollector *utils.UCollector = utils.NewUCollector(newGaugeVec)

func newGaugeVec(name string, labelNames []string) prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: name,
		},
		labelNames,
	)
}

func GetGaugeVec(name string, labelNames []string) *prometheus.GaugeVec {
	return GaugeCollector.GetCollector(name, labelNames).(*prometheus.GaugeVec)
}

func GetGauge(name string, labels map[string]string) prometheus.Gauge {
	return GetGaugeVec(name, utils.GetLabels(labels)).With(prometheus.Labels(labels))
}

func Add(name string, labels map[string]string, val float64) {
	GetGauge(name, labels).Add(val)
}

func Dec(name string, labels map[string]string) {
	GetGauge(name, labels).Dec()
}

func Inc(name string, labels map[string]string) {
	GetGauge(name, labels).Inc()
}

func Set(name string, labels map[string]string, val float64) {
	GetGauge(name, labels).Set(val)
}

func SetToCurrentTime(name string, labels map[string]string) {
	GetGauge(name, labels).SetToCurrentTime()
}

func Sub(name string, labels map[string]string, val float64) {
	GetGauge(name, labels).Sub(val)
}
