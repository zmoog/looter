package discordextension

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"
	"go.opentelemetry.io/collector/component"
)

// Discord is the interface other components use to access the shared session.
// Defined here so consumers import only this interface, not the concrete type.
type Discord interface {
	Session() *discordgo.Session
}

type discordExtension struct {
	cfg     *Config
	mu      sync.RWMutex
	session *discordgo.Session
}

func newExtension(cfg *Config) *discordExtension {
	return &discordExtension{cfg: cfg}
}

func (e *discordExtension) Start(_ context.Context, _ component.Host) error {
	s, err := discordgo.New("Bot " + e.cfg.BotToken)
	if err != nil {
		return err
	}
	s.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsMessageContent
	if err := s.Open(); err != nil {
		return err
	}
	e.mu.Lock()
	e.session = s
	e.mu.Unlock()
	return nil
}

func (e *discordExtension) Shutdown(_ context.Context) error {
	e.mu.Lock()
	s := e.session
	e.session = nil
	e.mu.Unlock()
	if s != nil {
		return s.Close()
	}
	return nil
}

func (e *discordExtension) Session() *discordgo.Session {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.session
}
