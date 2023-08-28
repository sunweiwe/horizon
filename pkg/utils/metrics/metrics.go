package metrics

import (
	"net/http"
	"sync"

	"github.com/emicklei/go-restful/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	compbasemetrics "k8s.io/component-base/metrics"
)

var (
	registerOnce sync.Once

	Defaults DefaultMetrics

	defaultRegistry compbasemetrics.KubeRegistry
	MustRegister    func(...compbasemetrics.Registerable)
	// RawMustRegister = defaultRegistry.RawMustRegister
)

func init() {
	defaultRegistry = compbasemetrics.NewKubeRegistry()
	MustRegister = defaultRegistry.MustRegister
}

type DefaultMetrics struct{}

func (d DefaultMetrics) Install(c *restful.Container) {
	registerOnce.Do(d.registerMetrics)
	c.Handle("/hapis/metrics", Handler())
}

func (d DefaultMetrics) registerMetrics() {
	// RawMustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// RawMustRegister(collectors.NewGoCollector())
}

func Handler() http.Handler {
	return promhttp.InstrumentMetricHandler(prometheus.NewRegistry(), promhttp.HandlerFor(defaultRegistry, promhttp.HandlerOpts{}))
}
