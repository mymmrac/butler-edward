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
	"github.com/mymmrac/butler-edward/pkg/handler/platform"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/channel"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/channel/telegram"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/channel/terminal"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider/openai"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/session/inmemory"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool/filesystem"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool/system"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool/web"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool/web/duckduckgo"
	"github.com/mymmrac/butler-edward/pkg/module/collection"
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
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
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
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var cfg platform.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	var channels []channel.Channel

	if cfg.Channels.Terminal.Enabled {
		channels = append(channels, terminal.NewTerminal())
	}

	if cfg.Channels.Telegram.Enabled {
		telegramChannel, err := telegram.NewTelegram(
			ctx, cfg.Channels.Telegram.BotToken, cfg.Channels.Telegram.AllowedChatIDs,
		)
		if err != nil {
			return fmt.Errorf("new telegram channel: %w", err)
		}
		channels = append(channels, telegramChannel)
	}

	var providers []provider.Provider

	for name, config := range cfg.Providers.OpenAICompatible {
		models := collection.MakeSlice[provider.Model](len(config.Models))
		for _, model := range config.Models {
			models = append(models, provider.Model{
				Name: model.Name,
			})
		}

		openAIProvider, err := openai.NewOpenAI(
			name, config.BaseURL, config.ChatAPI, config.ModelsAPI, config.APIKey, models,
		)
		if err != nil {
			return fmt.Errorf("new OpenAI compatible %q provider: %w", name, err)
		}

		providers = append(providers, openAIProvider)
	}

	root, err := os.OpenRoot(cfg.Workspace.Root)
	if err != nil {
		return fmt.Errorf("open workspace root: %w", err)
	}
	defer func() { _ = root.Close() }()

	agentInstance, err := agent.NewAgent(
		channels, providers,
		[]tool.Tool{
			filesystem.NewReadDirTool(root),
			filesystem.NewReadFileTool(root),
			filesystem.NewWriteFileTool(root),
			system.NewTimeTool(),
			web.NewSearchTool(duckduckgo.NewProvider()),
			web.NewFetchTool(),
		},
		inmemory.NewInMemory(),
	)
	if err != nil {
		return fmt.Errorf("new agent: %w", err)
	}

	if err = agentInstance.SelectProviderAndModel(ctx, cfg.Defaults.Provider, cfg.Defaults.Model); err != nil {
		return fmt.Errorf("select provider and model: %w", err)
	}

	if err = agentInstance.Run(ctx); err != nil {
		return fmt.Errorf("run agent: %w", err)
	}

	return nil
}
