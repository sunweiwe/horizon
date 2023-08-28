package prometheus

import (
	"time"

	"github.com/spf13/pflag"
)

type Options struct {
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	ClusterControllerResyncPeriod time.Duration `json:"clusterControllerResyncPeriod,omitempty" yaml:"clusterControllerResyncPeriod,omitempty"`

	HostClusterName string `json:"hostClusterName,omitempty" yaml:"hostClusterName,omitempty"`

	ClusterName string `json:"clusterName,omitempty" yaml:"clusterName,omitempty"`
}

func NewPrometheusOptions() *Options {
	return &Options{
		Endpoint: "",
	}
}

func (s *Options) AddFlags(fs *pflag.FlagSet, opt *Options) {
	fs.StringVar(&s.Endpoint, "prometheus-endpoint", opt.Endpoint, ""+
		"Prometheus service endpoint which stores horizon monitoring data, if left "+
		"blank, will use builtin metrics-server as data source.")
}
