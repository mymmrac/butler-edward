package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mymmrac/butler-edward/pkg/module/logger"
	"github.com/mymmrac/butler-edward/pkg/module/version"
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-signals:
			cancel() // Graceful shutdown
		}

		<-signals  // Force shutdown
		os.Exit(2) //revive:disable:deep-exit
	}()

	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvPrefix("BUTLER_EDWARD")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.SetDefault("logger.level", "debug")

	logger.SetLevel(v.GetString("logger.level"))

	rootCmd := &cobra.Command{
		Use:  "butler-edward",
		Args: cobra.NoArgs,
		Version: fmt.Sprintf("%s (%s %s), built at %s",
			version.Version(), version.Revision(), version.Modified(), version.BuildTime(),
		),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			runCtx := cmd.Context()

			logger.Infow(runCtx, "starting",
				"version", version.Version(),
				"build-time", version.BuildTime(),
				"revision", version.Revision(),
			)

			if err := run(runCtx, v); err != nil {
				return fmt.Errorf("run: %w", err)
			}

			logger.Info(runCtx, "shutting down ")
			return nil
		},
	}

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1) //nolint:gocritic //revive:disable:deep-exit
	}
}

func run(ctx context.Context, v *viper.Viper) error {
	_, _ = ctx, v
	//nolint:forbidigo
	//revive:disable:unhandled-error
	fmt.Println("Hello World!")
	return nil
}
