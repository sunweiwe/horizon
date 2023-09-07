package multicluster

import "time"

const (
	DefaultResyncPeriod    = 120 * time.Second
	DefaultHostClusterName = "host"
)

type Options struct {
	ProxyPublishService string `json:"proxyPublishService,omitempty" yaml:"proxyPublishService,omitempty"`

	ProxyPublishAddress string `json:"proxyPublishAddress,omitempty" yaml:"proxyPublishAddress,omitempty"`

	AgentImage string `json:"agentImage,omitempty" yaml:"agentImage,omitempty"`

	Enable bool `json:"enable" yaml:"enable"`

	ClusterName string `json:"clusterName,omitempty" yaml:"clusterName,omitempty"`

	ClusterControllerResyncPeriod time.Duration `json:"clusterControllerResyncPeriod,omitempty" yaml:"clusterControllerResyncPeriod,omitempty"`

	HostClusterName string `json:"hostClusterName,omitempty" yaml:"hostClusterName,omitempty"`
}

func NewOptions() *Options {
	return &Options{
		Enable:                        false,
		ProxyPublishAddress:           "",
		ProxyPublishService:           "",
		ClusterControllerResyncPeriod: DefaultResyncPeriod,
		HostClusterName:               DefaultHostClusterName,
	}
}
