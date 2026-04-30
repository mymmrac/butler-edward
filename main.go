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

	"github.com/mymmrac/butler-edward/pkg/handler/agent"
	"github.com/mymmrac/butler-edward/pkg/module/logger"
	"github.com/mymmrac/butler-edward/pkg/module/platform/channel"
	"github.com/mymmrac/butler-edward/pkg/module/platform/channel/telegram"
	"github.com/mymmrac/butler-edward/pkg/module/platform/channel/terminal"
	"github.com/mymmrac/butler-edward/pkg/module/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/module/platform/provider/openai"
	"github.com/mymmrac/butler-edward/pkg/module/platform/session/inmemory"
	"github.com/mymmrac/butler-edward/pkg/module/platform/tool"
	"github.com/mymmrac/butler-edward/pkg/module/platform/tool/filesystem"
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
	telegramChannel, err := telegram.NewTelegram(ctx, v.GetString("telegram.bot-token"))
	if err != nil {
		return fmt.Errorf("new telegram channel: %w", err)
	}

	groqProvider, err := openai.NewOpenAI(v.GetString("groq.base-url"), v.GetString("groq.api-key"))
	if err != nil {
		return fmt.Errorf("new groq provider: %w", err)
	}

	root, err := os.OpenRoot(v.GetString("workspace.root"))
	if err != nil {
		return fmt.Errorf("open workspace root: %w", err)
	}
	defer func() { _ = root.Close() }()

	agentInstance, err := agent.NewAgent(
		[]channel.Channel{
			telegramChannel,
			terminal.NewTerminal(),
		},
		[]provider.Provider{
			groqProvider,
		},
		[]tool.Tool{
			filesystem.NewReadDirTool(root),
			filesystem.NewReadFileTool(root),
			filesystem.NewWriteFileTool(root),
		},
		inmemory.NewInMemory(),
	)
	if err != nil {
		return fmt.Errorf("new agent: %w", err)
	}

	if err = agentInstance.SelectProviderAndModel(ctx, "api.groq.com", "llama-3.1-8b-instant"); err != nil {
		return fmt.Errorf("select provider and model: %w", err)
	}

	if err = agentInstance.Run(ctx); err != nil {
		return fmt.Errorf("run agent: %w", err)
	}

	return nil
}
