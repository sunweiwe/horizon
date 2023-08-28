package app

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sunweiwe/horizon/cmd/controller-manager/app/options"
	"github.com/sunweiwe/horizon/pkg/apis"
	"github.com/sunweiwe/horizon/pkg/informers"
	"github.com/sunweiwe/horizon/pkg/simple/client/k8s"
	"k8s.io/component-base/term"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	controllerConfig "github.com/sunweiwe/horizon/pkg/apiserver/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliFlag "k8s.io/component-base/cli/flag"
	ctrl "sigs.k8s.io/controller-runtime"
)

func NewControllerManagerCommand() *cobra.Command {
	s := options.NewHorizonControllerManagerOptions()
	conf, err := controllerConfig.TryLoadFromDisk()

	if err == nil {
		s = &options.HorizonControllerManagerOptions{
			KubernetesOptions:   conf.KubernetesOptions,
			LeaderElect:         s.LeaderElect,
			LeaderElection:      s.LeaderElection,
			WebhookCertDir:      s.WebhookCertDir,
			MonitoringOptions:   conf.MonitoringOptions,
			MultiClusterOptions: conf.MultiClusterOptions,
			ControllerGates:     []string{"*"},
		}
	} else {
		klog.Fatalf("Failed to load configuration from disk: %v", err)
	}

	cmd := &cobra.Command{
		Use:  "controller-manager",
		Long: `Horizon controller manager is a daemon that embeds the control loops shipped with horizon.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := Run(s, controllerConfig.WatchConfigChange(), signals.SetupSignalHandler()); err != nil {
				klog.Error(err)
				os.Exit(1)
			}
		},
		SilenceUsage: true,
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags(allControllers)
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliFlag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	return cmd
}

func Run(s *options.HorizonControllerManagerOptions, configCh <-chan controllerConfig.Config, ctx context.Context) error {
	ictx, cancelFunc := context.WithCancel(context.TODO())
	errCh := make(chan error)
	defer close(errCh)
	go func() {
		if err := run(s, ictx); err != nil {
			errCh <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			cancelFunc()
			return nil
		case cfg := <-configCh:
			cancelFunc()
			s.MergeConfig(&cfg)
			ictx, cancelFunc = context.WithCancel(context.TODO())
			go func() {
				if err := run(s, ictx); err != nil {
					errCh <- err
				}
			}()
		case err := <-errCh:
			cancelFunc()
			return err
		}
	}
}

func run(s *options.HorizonControllerManagerOptions, ctx context.Context) error {
	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		klog.Errorf("Failed to create kubernetes clientSet %v", err)
		return err
	}

	informerFactory := informers.NewInformerFactories(
		kubernetesClient.Kubernetes(),
		kubernetesClient.Horizon(),
	)

	mgrOptions := manager.Options{
		CertDir: s.WebhookCertDir,
		Port:    8443,
	}

	if s.LeaderElect {
		mgrOptions = manager.Options{
			CertDir:              s.WebhookCertDir,
			Port:                 8443,
			LeaderElection:       s.LeaderElect,
			LivenessEndpointName: "horizon-system",
			LeaderElectionID:     "hz-controller-manager-leader-election",
			LeaseDuration:        &s.LeaderElection.LeaseDuration,
			RetryPeriod:          &s.LeaderElection.RetryPeriod,
			RenewDeadline:        &s.LeaderElection.RenewDeadline,
		}
	}

	klog.V(0).Info("setting up manager")
	ctrl.SetLogger(klog.NewKlogr())

	mgr, err := manager.New(kubernetesClient.Config(), mgrOptions)
	if err != nil {
		klog.Fatalf("unable to set up overall controller manager: %v", err)
	}

	if err = apis.AddToScheme(mgr.GetScheme()); err != nil {
		klog.Fatalf("unable add APIs to scheme: %v", err)
	}

	metav1.AddToGroupVersion(mgr.GetScheme(), metav1.SchemeGroupVersion)

	if err = addAllControllers(mgr,
		kubernetesClient,
		informerFactory,
		s,
		ctx.Done()); err != nil {
		klog.Fatalf("unable to register controllers to the manager: %v", err)
	}

	klog.V(0).Info("Starting cache resource from apiServer...")
	informerFactory.Start(ctx.Done())

	// klog.V(0).Info("Starting the controllers.")
	// if err = mgr.Start(ctx); err != nil {
	// 	klog.Fatalf("unable to run the manager: %v", err)
	// }

	return nil
}
