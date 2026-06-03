package discordextension

import "fmt"

type Config struct {
	BotToken string `mapstructure:"bot_token"`
}

func (cfg *Config) Validate() error {
	if cfg.BotToken == "" {
		return fmt.Errorf("bot_token is required")
	}
	return nil
}
