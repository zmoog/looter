package discordreceiver

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zmoog/looter/extension/discordextension"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type discordReceiver struct {
	cfg       *Config
	settings  receiver.Settings
	consumer  consumer.Logs
	session   *discordgo.Session
	removeFns []func()
	ctx       context.Context
	cancel    context.CancelFunc
}

func newReceiver(cfg *Config, settings receiver.Settings, consumer consumer.Logs) *discordReceiver {
	return &discordReceiver{cfg: cfg, settings: settings, consumer: consumer}
}

func (r *discordReceiver) Start(_ context.Context, host component.Host) error {
	r.ctx, r.cancel = context.WithCancel(context.Background())
	var ext discordextension.Discord
	for _, e := range host.GetExtensions() {
		if d, ok := e.(discordextension.Discord); ok {
			ext = d
			break
		}
	}
	if ext == nil {
		r.cancel()
		return fmt.Errorf("discordreceiver requires the discord extension to be configured")
	}
	r.session = ext.Session()
	r.removeFns = []func(){
		r.session.AddHandler(r.onMessageCreate),
		r.session.AddHandler(r.onMessageReactionAdd),
	}
	return nil
}

func (r *discordReceiver) Shutdown(_ context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}
	for _, fn := range r.removeFns {
		fn()
	}
	return nil
}

func (r *discordReceiver) onMessageCreate(s *discordgo.Session, e *discordgo.MessageCreate) {
	if !allowEvent(r.cfg, e.GuildID, e.ChannelID) {
		return
	}
	logs := messageCreateToLogs(botUserID(s), e)
	if err := r.consumer.ConsumeLogs(r.ctx, logs); err != nil {
		r.settings.Logger.Error("failed to consume message_create event", zap.Error(err))
	}
}

func (r *discordReceiver) onMessageReactionAdd(s *discordgo.Session, e *discordgo.MessageReactionAdd) {
	if !allowEvent(r.cfg, e.GuildID, e.ChannelID) {
		return
	}
	logs := messageReactionAddToLogs(botUserID(s), e)
	if err := r.consumer.ConsumeLogs(r.ctx, logs); err != nil {
		r.settings.Logger.Error("failed to consume message_reaction_add event", zap.Error(err))
	}
}

func botUserID(s *discordgo.Session) string {
	if s.State != nil && s.State.User != nil {
		return s.State.User.ID
	}
	return ""
}

func messageCreateToLogs(botID string, e *discordgo.MessageCreate) plog.Logs {
	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr("discord.bot_id", botID)
	rl.Resource().Attributes().PutStr("service.name", "discordreceiver")

	lr := rl.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	lr.SetTimestamp(pcommon.NewTimestampFromTime(e.Timestamp))
	lr.Body().SetStr(e.Content)
	lr.SetSeverityNumber(plog.SeverityNumberInfo)

	attrs := lr.Attributes()
	attrs.PutStr("discord.event_type", "message_create")
	attrs.PutStr("discord.guild_id", e.GuildID)
	attrs.PutStr("discord.channel_id", e.ChannelID)
	attrs.PutStr("discord.message_id", e.ID)
	authorID, authorName := "", ""
	if e.Author != nil {
		authorID = e.Author.ID
		authorName = e.Author.Username
	}
	attrs.PutStr("discord.author_id", authorID)
	attrs.PutStr("discord.author_name", authorName)
	attrs.PutStr("discord.emoji", "")

	return logs
}

func messageReactionAddToLogs(botID string, e *discordgo.MessageReactionAdd) plog.Logs {
	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr("discord.bot_id", botID)
	rl.Resource().Attributes().PutStr("service.name", "discordreceiver")

	lr := rl.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	lr.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	lr.Body().SetStr("")
	lr.SetSeverityNumber(plog.SeverityNumberInfo)

	attrs := lr.Attributes()
	attrs.PutStr("discord.event_type", "message_reaction_add")
	attrs.PutStr("discord.guild_id", e.GuildID)
	attrs.PutStr("discord.channel_id", e.ChannelID)
	attrs.PutStr("discord.message_id", e.MessageID)
	attrs.PutStr("discord.author_id", e.UserID)
	attrs.PutStr("discord.author_name", "")
	attrs.PutStr("discord.emoji", e.Emoji.Name)

	return logs
}

func allowEvent(cfg *Config, guildID, channelID string) bool {
	if len(cfg.GuildIDs) > 0 {
		found := false
		for _, id := range cfg.GuildIDs {
			if id == guildID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(cfg.ChannelIDs) > 0 {
		for _, id := range cfg.ChannelIDs {
			if id == channelID {
				return true
			}
		}
		return false
	}
	return true
}
