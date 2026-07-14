package discord

import (
	"log"
	"strings"

	"job-tracker/internal/config"

	"github.com/bwmarrin/discordgo"
)

// Bot is the Discord inbound adapter. M0 is plumbing-only: it connects and
// answers !ping. Commands (add/edit/view) and their service/repo dependencies
// are added in later milestones.
type Bot struct {
	session *discordgo.Session
	prefix  string
}

func New(cfg *config.Config) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, err
	}
	// Message Content is a privileged intent — must be enabled in the Discord
	// Developer Portal for this bot, or messageCreate content arrives empty.
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent

	b := &Bot{session: session, prefix: cfg.DiscordPrefix}
	session.AddHandler(b.messageCreate)
	return b, nil
}

// Start opens the gateway. discordgo runs its own goroutine, so this returns
// once connected and the caller keeps running.
func (b *Bot) Start() error {
	return b.session.Open()
}

func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	if !strings.HasPrefix(m.Content, b.prefix) {
		return
	}
	// ponytail: prefix parsing is fragile (splits on whitespace, no quoting) —
	// good enough for MVP; revisit if commands need multi-word args beyond !add.
	cmd := strings.Fields(strings.TrimPrefix(m.Content, b.prefix))
	if len(cmd) == 0 {
		return
	}
	switch cmd[0] {
	case "ping":
		if _, err := s.ChannelMessageSend(m.ChannelID, "pong"); err != nil {
			log.Printf("discord: reply failed: %v", err)
		}
	}
}
