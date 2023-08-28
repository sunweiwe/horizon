package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/sunweiwe/horizon/pkg/constants"
	"github.com/sunweiwe/horizon/pkg/simple/client/k8s"
	"github.com/sunweiwe/horizon/pkg/simple/client/monitoring/prometheus"
	"github.com/sunweiwe/horizon/pkg/simple/client/multicluster"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
)

var (
	// singleton instance of config package
	_config = defaultConfig()
)

const (
	defaultConfigurationName = "horizon"

	defaultConfigurationPath = "/etc/horizon"
)

type config struct {
	cfg         *Config
	cfgChangeCh chan Config
	watchOnce   sync.Once
	loadOnce    sync.Once
}

func (c *config) watchConfig() <-chan Config {
	c.watchOnce.Do(func() {
		viper.WatchConfig()
		viper.OnConfigChange(func(in fsnotify.Event) {
			cfg := New()
			if err := viper.Unmarshal(cfg); err != nil {
				klog.Warningf("config reload error: %v", err)
			} else {
				c.cfgChangeCh <- *cfg
			}
		})
	})

	return c.cfgChangeCh
}

func (c *config) loadFromDisk() (*Config, error) {
	var err error
	c.loadOnce.Do(func() {
		if err = viper.ReadInConfig(); err != nil {
			return
		}
		err = viper.Unmarshal(c.cfg)
	})
	return c.cfg, err
}

type Config struct {
	KubernetesOptions *k8s.KubernetesOptions `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes"`

	MonitoringOptions *prometheus.Options `json:"monitoring,omitempty" yaml:"monitoring,omitempty" mapstructure:"monitoring"`

	MultiClusterOptions *multicluster.Options `json:"multicluster,omitempty" yaml:"multicluster,omitempty" mapstructure:"multicluster"`
}

func New() *Config {
	return &Config{
		KubernetesOptions:   k8s.NewKubernetesClientOptions(),
		MonitoringOptions:   prometheus.NewPrometheusOptions(),
		MultiClusterOptions: multicluster.NewOptions(),
	}
}

func defaultConfig() *config {
	viper.SetConfigName(defaultConfigurationName)
	viper.AddConfigPath(defaultConfigurationPath)

	viper.AddConfigPath(".")

	viper.SetEnvPrefix("horizon")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return &config{
		cfg:         New(),
		cfgChangeCh: make(chan Config),
		watchOnce:   sync.Once{},
		loadOnce:    sync.Once{},
	}
}

func WatchConfigChange() <-chan Config {
	return _config.watchConfig()
}

func TryLoadFromDisk() (*Config, error) {
	return _config.loadFromDisk()
}

func GetFromConfigMap(cm *corev1.ConfigMap) (*Config, error) {
	c := &Config{}
	value, ok := cm.Data[constants.HorizonConfigMapDataKey]
	if !ok {
		return nil, fmt.Errorf("Failed to get configmap horizon.yaml value")
	}

	if err := yaml.Unmarshal([]byte(value), c); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal value from configmap. err: %s", err)
	}

	return c, nil
}
