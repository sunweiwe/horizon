package prometheus

import (
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/sunweiwe/horizon/pkg/simple/client/monitoring"

	"github.com/prometheus/client_golang/api"
)

type prometheus struct {
	client v1.API
}

func NewPrometheus(options *Options) (monitoring.Interface, error) {
	cfg := api.Config{
		Address: options.Endpoint,
	}

	client, err := api.NewClient(cfg)
	return prometheus{client: v1.NewAPI(client)}, err
}
