# Discord Extension + Receiver Design

**Date:** 2026-06-04
**Status:** Approved

## Overview

Add real-time Discord event ingestion to the looter custom OTel Collector. A shared `discordextension` owns the bot Gateway connection; a `discordreceiver` registers event handlers on it and emits OTel log records. A future `discordexporter` will reuse the same extension to post messages back.

## Module Structure

```
looter/
├── extension/
│   └── discordextension/       # github.com/zmoog/looter/extension/discordextension
│       ├── go.mod
│       ├── config.go
│       ├── factory.go
│       └── extension.go
├── receiver/
│   └── discordreceiver/        # github.com/zmoog/looter/receiver/discordreceiver
│       ├── go.mod
│       ├── config.go
│       ├── factory.go
│       └── receiver.go
├── go.work                     # adds both modules
└── builder-config.yaml         # references both via file:// gomod paths
```

Both modules are local to the looter repo, referenced in `go.work` and `builder-config.yaml` with `file://` paths. This matches the pattern used in the collector repo.

## Extension: discordextension

### Purpose

Owns the `*discordgo.Session` lifecycle. Exposes it through a narrow interface so the receiver and future exporter can use it without importing the extension package directly.

### Interface

```go
// Discord is the interface other components use to access the shared session.
// Defined in discordextension so consumers import the interface, not the concrete type.
type Discord interface {
    Session() *discordgo.Session
}
```

### Config

```go
type Config struct {
    BotToken string `mapstructure:"bot_token"`
}
```

Collector config:
```yaml
extensions:
  discord:
    bot_token: ${env:DISCORD_BOT_TOKEN}
```

### Lifecycle

- **Start**: `discordgo.New("Bot " + cfg.BotToken)`, set intents (`GuildMessages`, `GuildMessageReactions`, `MessageContent`), call `session.Open()`
- **Shutdown**: `session.Close()`

## Receiver: discordreceiver

### Purpose

Subscribes to `MessageCreate` and `MessageReactionAdd` Discord Gateway events and emits each as one OTel `LogRecord` to the downstream consumer.

### Config

```go
type Config struct {
    GuildIDs   []string `mapstructure:"guild_ids"`    // empty = all guilds
    ChannelIDs []string `mapstructure:"channel_ids"`  // empty = all channels
}
```

Collector config:
```yaml
receivers:
  discord:
    guild_ids: []
    channel_ids: []
```

### Startup

On `Start`, the receiver:
1. Iterates `host.GetExtensions()` looking for a component implementing `discordextension.Discord`
2. Returns an error if none is found
3. Calls `session.AddHandler(r.onMessageCreate)` and `session.AddHandler(r.onMessageReactionAdd)`, storing the returned remove functions

On `Shutdown`, calls both remove functions.

### Data Model

Each event produces one `plog.Logs` with a single `ResourceLogs → ScopeLogs → LogRecord`.

**Resource attributes** (set once per `ResourceLogs`):

| Attribute | Value |
|---|---|
| `discord.bot_id` | `session.State.User.ID` |
| `service.name` | `"discordreceiver"` |

**Log record fields:**

| OTel field | `MessageCreate` | `MessageReactionAdd` |
|---|---|---|
| `Timestamp` | `message.Timestamp` | time.Now() |
| `Body` | message content | `""` |
| `SeverityNumber` | `INFO` | `INFO` |
| `discord.event_type` | `message_create` | `message_reaction_add` |
| `discord.guild_id` | `message.GuildID` | `reaction.GuildID` |
| `discord.channel_id` | `message.ChannelID` | `reaction.ChannelID` |
| `discord.message_id` | `message.ID` | `reaction.MessageID` |
| `discord.author_id` | `message.Author.ID` | `reaction.UserID` |
| `discord.author_name` | `message.Author.Username` | `""` |
| `discord.emoji` | `""` | `reaction.Emoji.Name` |

### Filtering

Before emitting, the receiver checks:
- If `guild_ids` is non-empty and the event's guild ID is not in the list → skip
- If `channel_ids` is non-empty and the event's channel ID is not in the list → skip

### Backpressure

Events are forwarded immediately (no buffering). discordgo calls handlers in a goroutine per event, so a slow consumer blocks that goroutine briefly. Acceptable at personal-project scale.

## Integration

### builder-config.yaml additions

```yaml
extensions:
  - gomod: github.com/zmoog/looter/extension/discordextension file://./extension/discordextension

receivers:
  - gomod: github.com/zmoog/looter/receiver/discordreceiver file://./receiver/discordreceiver
```

### collector.yaml additions

```yaml
extensions:
  discord:
    bot_token: ${env:DISCORD_BOT_TOKEN}

receivers:
  discord:
    guild_ids: []
    channel_ids: []

service:
  extensions: [health_check, pprof, zpages, discord]
  pipelines:
    logs:
      receivers: [otlp, discord]
      processors: [memory_limiter, batch]
      exporters: [debug]
```

## Out of Scope

- `discordexporter` — deferred to a follow-up
- Message edits / deletes (`MessageUpdate`, `MessageDelete`)
- Voice events, member join/leave
- Rate-limit handling beyond discordgo's built-in
- Persistent cursor / replay of missed events across restarts
