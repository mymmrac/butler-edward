package platform

// Config represents the main configuration structure for the application.
//
//revive:disable:nested-structs
type Config struct {
	Defaults struct {
		Provider string `mapstructure:"provider"`
		Model    string `mapstructure:"model"`
	} `mapstructure:"defaults"`
	Workspace struct {
		Root   string `mapstructure:"root"`
		BoltDB string `mapstructure:"bolt-db"`
	} `mapstructure:"workspace"`
	Channels struct {
		Terminal struct {
			Enabled bool `mapstructure:"enabled"`
		} `mapstructure:"terminal"`
		Telegram struct {
			Enabled        bool    `mapstructure:"enabled"`
			BotToken       string  `mapstructure:"bot-token"`
			AllowedChatIDs []int64 `mapstructure:"allowed-chat-ids"`
		} `mapstructure:"telegram"`
	} `mapstructure:"channels"`
	Providers struct {
		OpenAICompatible map[string]struct {
			Enabled   bool   `mapstructure:"enabled"`
			BaseURL   string `mapstructure:"base-url"`
			ChatAPI   string `mapstructure:"chat-api"`
			ModelsAPI string `mapstructure:"models-api"`
			APIKey    string `mapstructure:"api-key"`
			Models    []struct {
				Name string `mapstructure:"name"`
			} `mapstructure:"models"`
		} `mapstructure:"openai-compatible"`
	} `mapstructure:"providers"`
}
