package app

import (
	"context"
	"net/http"

	"github.com/google/gops/agent"
	"github.com/spf13/cobra"
	"github.com/sunweiwe/horizon/cmd/hz-apiserver/app/options"
	apiserverconfig "github.com/sunweiwe/horizon/pkg/apiserver/config"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func NewAPIServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()

	conf, err := apiserverconfig.TryLoadFromDisk()
	if err == nil {
		s.Config = conf
	} else {
		klog.Fatalf("Failed to load configuration from disk: %v", err)
	}

	cmd := &cobra.Command{
		Use: "hz-apiserver",
		Long: `The Horizon API server validates and configures data for the API objects. 
		The API Server services REST operations and provides the frontend to the
		cluster's shared state through which all other components interact.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			if s.DebugMode {
				if err := agent.Listen(agent.Options{}); err != nil {
					klog.Fatal(err)
				}
			}

			return Run(s, apiserverconfig.WatchConfigChange(), signals.SetupSignalHandler())
		},
		SilenceUsage: true,
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	return cmd
}

func Run(s *options.ServerRunOptions, configCh <-chan apiserverconfig.Config, ctx context.Context) error {
	ictx, cancel := context.WithCancel(context.TODO())
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
			cancel()
			return nil
		case cfg := <-configCh:
			cancel()
			s.Config = &cfg
			ictx, cancel = context.WithCancel(context.TODO())
			go func() {
				if err := run(s, ictx); err != nil {
					errCh <- err
				}
			}()
		case err := <-errCh:
			cancel()
			return err
		}
	}

}

func run(s *options.ServerRunOptions, ctx context.Context) error {
	klog.V(0).Info("Cmd Apiserver run")
	apiserver, err := s.NewAPIServer(ctx.Done())
	if err != nil {
		return err
	}

	err = apiserver.PrepareRun(ctx.Done())
	if err != nil {
		return err
	}

	err = apiserver.Run(ctx)
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}
