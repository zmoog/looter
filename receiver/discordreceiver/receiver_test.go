package discordreceiver

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestMessageCreateToLogs(t *testing.T) {
	ts := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	event := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			ID:        "msg-123",
			ChannelID: "chan-456",
			GuildID:   "guild-789",
			Content:   "hello world",
			Timestamp: ts,
			Author: &discordgo.User{
				ID:       "user-111",
				Username: "testuser",
			},
		},
	}

	logs := messageCreateToLogs("bot-999", event)

	require.Equal(t, 1, logs.ResourceLogs().Len())
	rl := logs.ResourceLogs().At(0)

	botID, ok := rl.Resource().Attributes().Get("discord.bot_id")
	require.True(t, ok)
	assert.Equal(t, "bot-999", botID.Str())

	svcName, ok := rl.Resource().Attributes().Get("service.name")
	require.True(t, ok)
	assert.Equal(t, "discordreceiver", svcName.Str())

	require.Equal(t, 1, rl.ScopeLogs().Len())
	sl := rl.ScopeLogs().At(0)
	require.Equal(t, 1, sl.LogRecords().Len())
	lr := sl.LogRecords().At(0)

	assert.Equal(t, pcommon.NewTimestampFromTime(ts), lr.Timestamp())
	assert.Equal(t, "hello world", lr.Body().Str())
	assert.Equal(t, plog.SeverityNumberInfo, lr.SeverityNumber())

	checkAttr := func(key, expected string) {
		t.Helper()
		v, ok := lr.Attributes().Get(key)
		require.True(t, ok, "missing attribute %s", key)
		assert.Equal(t, expected, v.Str())
	}
	checkAttr("discord.event_type", "message_create")
	checkAttr("discord.guild_id", "guild-789")
	checkAttr("discord.channel_id", "chan-456")
	checkAttr("discord.message_id", "msg-123")
	checkAttr("discord.author_id", "user-111")
	checkAttr("discord.author_name", "testuser")
	checkAttr("discord.emoji", "")
}

func TestMessageReactionAddToLogs(t *testing.T) {
	event := &discordgo.MessageReactionAdd{
		MessageReaction: &discordgo.MessageReaction{
			UserID:    "user-111",
			MessageID: "msg-123",
			ChannelID: "chan-456",
			GuildID:   "guild-789",
			Emoji:     discordgo.Emoji{Name: "👍"},
		},
	}

	logs := messageReactionAddToLogs("bot-999", event)

	require.Equal(t, 1, logs.ResourceLogs().Len())
	rl := logs.ResourceLogs().At(0)
	require.Equal(t, 1, rl.ScopeLogs().Len())
	lr := rl.ScopeLogs().At(0).LogRecords().At(0)

	assert.Equal(t, plog.SeverityNumberInfo, lr.SeverityNumber())
	assert.Equal(t, "", lr.Body().Str())

	checkAttr := func(key, expected string) {
		t.Helper()
		v, ok := lr.Attributes().Get(key)
		require.True(t, ok, "missing attribute %s", key)
		assert.Equal(t, expected, v.Str())
	}
	checkAttr("discord.event_type", "message_reaction_add")
	checkAttr("discord.guild_id", "guild-789")
	checkAttr("discord.channel_id", "chan-456")
	checkAttr("discord.message_id", "msg-123")
	checkAttr("discord.author_id", "user-111")
	checkAttr("discord.author_name", "")
	checkAttr("discord.emoji", "👍")
}

func TestAllowEvent(t *testing.T) {
	t.Run("empty filters allow everything", func(t *testing.T) {
		cfg := &Config{}
		assert.True(t, allowEvent(cfg, "any-guild", "any-chan"))
	})

	t.Run("guild filter: matching guild allowed", func(t *testing.T) {
		cfg := &Config{GuildIDs: []string{"guild-1"}}
		assert.True(t, allowEvent(cfg, "guild-1", "any-chan"))
	})

	t.Run("guild filter: non-matching guild blocked", func(t *testing.T) {
		cfg := &Config{GuildIDs: []string{"guild-1"}}
		assert.False(t, allowEvent(cfg, "guild-2", "any-chan"))
	})

	t.Run("channel filter: matching channel allowed", func(t *testing.T) {
		cfg := &Config{ChannelIDs: []string{"chan-1"}}
		assert.True(t, allowEvent(cfg, "any-guild", "chan-1"))
	})

	t.Run("channel filter: non-matching channel blocked", func(t *testing.T) {
		cfg := &Config{ChannelIDs: []string{"chan-1"}}
		assert.False(t, allowEvent(cfg, "any-guild", "chan-2"))
	})

	t.Run("both filters: must match both", func(t *testing.T) {
		cfg := &Config{GuildIDs: []string{"guild-1"}, ChannelIDs: []string{"chan-1"}}
		assert.True(t, allowEvent(cfg, "guild-1", "chan-1"))
		assert.False(t, allowEvent(cfg, "guild-1", "chan-2"))
		assert.False(t, allowEvent(cfg, "guild-2", "chan-1"))
	})
}
