package options

import (
	"time"

	"github.com/spf13/pflag"
	controllerconfig "github.com/sunweiwe/horizon/pkg/apiserver/config"
	"github.com/sunweiwe/horizon/pkg/simple/client/k8s"
	"github.com/sunweiwe/horizon/pkg/simple/client/monitoring/prometheus"
	"github.com/sunweiwe/horizon/pkg/simple/client/multicluster"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/component-base/cli/flag"
)

type HorizonControllerManagerOptions struct {
	KubernetesOptions *k8s.KubernetesOptions

	MonitoringOptions   *prometheus.Options
	MultiClusterOptions *multicluster.Options

	LeaderElect    bool
	LeaderElection *leaderelection.LeaderElectionConfig

	WebhookCertDir string

	ApplicationSelector string

	ControllerGates []string
}

func NewHorizonControllerManagerOptions() *HorizonControllerManagerOptions {
	s := &HorizonControllerManagerOptions{
		KubernetesOptions:   k8s.NewKubernetesClientOptions(),
		MultiClusterOptions: multicluster.NewOptions(),
		LeaderElect:         false,
		LeaderElection: &leaderelection.LeaderElectionConfig{
			LeaseDuration: 30 * time.Second,
			RenewDeadline: 15 * time.Second,
			RetryPeriod:   5 * time.Second,
		},
		WebhookCertDir:  "",
		ControllerGates: []string{"*"},
	}

	return s
}

func (s *HorizonControllerManagerOptions) MergeConfig(cfg *controllerconfig.Config) {
	s.KubernetesOptions = cfg.KubernetesOptions
}

func (s *HorizonControllerManagerOptions) GetControllerEnabled(name string) bool {
	hasStar := false

	for _, ctrl := range s.ControllerGates {
		if ctrl == name {
			return true
		}
		if ctrl == "-"+name {
			return false
		}
		if ctrl == "*" {
			hasStar = true
		}
	}

	return hasStar
}

func (s *HorizonControllerManagerOptions) Flags(allControllerNameSelectors []string) flag.NamedFlagSets {
	fss := flag.NamedFlagSets{}

	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)

	fs := fss.FlagSet("leaderelection")
	s.bindLeaderElectionFlags(s.LeaderElection, fs)

	return fss
}

func (s *HorizonControllerManagerOptions) bindLeaderElectionFlags(l *leaderelection.LeaderElectionConfig, fs *pflag.FlagSet) {
	fs.DurationVar(&l.LeaseDuration, "leader-elect-lease-duration", l.LeaseDuration, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	fs.DurationVar(&l.RenewDeadline, "leader-elect-renew-deadline", l.RenewDeadline, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than or equal to the lease duration. "+
		"This is only applicable if leader election is enabled.")
	fs.DurationVar(&l.RetryPeriod, "leader-elect-retry-period", l.RetryPeriod, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")
}
