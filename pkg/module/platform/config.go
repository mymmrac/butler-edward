package platform

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// Config represents the main configuration structure for the application.
type Config struct {
	Workspace struct {
		Root string `mapstructure:"root"`
	} `mapstructure:"workspace"`
	Channels struct {
		Terminal struct {
			Enabled bool `mapstructure:"enabled"`
		} `mapstructure:"terminal"`
		Telegram struct {
			Enabled  bool   `mapstructure:"enabled"`
			BotToken string `mapstructure:"bot-token"`
		} `mapstructure:"telegram"`
	} `mapstructure:"channels"`
	Providers struct {
		OpenAICompatible map[string]struct {
			Enabled bool   `mapstructure:"enabled"`
			BaseURL string `mapstructure:"base-url"`
			APIKey  string `mapstructure:"api-key"`
		} `mapstructure:"openai-compatible"`
	} `mapstructure:"providers"`
}

// ParseConfig parses the configuration from the given viper instance.
func ParseConfig(v *viper.Viper) (*Config, error) {
	var raw map[string]any
	if err := v.Unmarshal(&raw); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	normalized := normalizeKeys(raw)

	var cfg Config
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &cfg,
		TagName: "mapstructure",
	})
	if err != nil {
		return nil, fmt.Errorf("new decoder: %w", err)
	}

	if err = decoder.Decode(normalized); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	return &cfg, nil
}

func normalizeKeys(v any) any {
	switch values := v.(type) {
	case map[string]any:
		out := make(map[string]any)
		for key, value := range values {
			out[snakeToKebab(key)] = normalizeKeys(value)
		}
		return out
	case []any:
		for i := range values {
			values[i] = normalizeKeys(values[i])
		}
	}
	return v
}

func snakeToKebab(s string) string {
	var result []rune
	for _, r := range s {
		if r == '_' {
			result = append(result, '-')
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
