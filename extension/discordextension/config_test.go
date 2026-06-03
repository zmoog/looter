package discordextension

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("empty bot_token returns error", func(t *testing.T) {
		cfg := &Config{BotToken: ""}
		assert.Error(t, cfg.Validate())
	})

	t.Run("non-empty bot_token is valid", func(t *testing.T) {
		cfg := &Config{BotToken: "mytoken"}
		assert.NoError(t, cfg.Validate())
	})
}
