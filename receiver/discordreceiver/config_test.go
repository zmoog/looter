package discordreceiver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("empty config is valid (no required fields)", func(t *testing.T) {
		cfg := &Config{}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("guild_ids and channel_ids filters are optional", func(t *testing.T) {
		cfg := &Config{
			GuildIDs:   []string{"guild-1"},
			ChannelIDs: []string{"chan-1"},
		}
		assert.NoError(t, cfg.Validate())
	})
}
