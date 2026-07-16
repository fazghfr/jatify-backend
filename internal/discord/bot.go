package discord

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"job-tracker/internal/config"
	"job-tracker/internal/dto"
	"job-tracker/internal/entity"
	"job-tracker/internal/repository"
	"job-tracker/internal/service"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Bot is the Discord inbound adapter. It reuses the existing service/repository
// layer; no service signatures change. userID is a single configured jatify
// user (MVP — no Discord→user mapping).
type Bot struct {
	session    *discordgo.Session
	appSvc     service.ApplicationService
	jobRepo    repository.JobRepository
	statusRepo repository.StatusRepository
	userID     int
	prefix     string
}

func New(cfg *config.Config, appSvc service.ApplicationService, jobRepo repository.JobRepository, statusRepo repository.StatusRepository) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, err
	}
	// Message Content is a privileged intent — must be enabled in the Discord
	// Developer Portal for this bot, or messageCreate content arrives empty.
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent | discordgo.IntentsGuildMessageReactions

	b := &Bot{
		session:    session,
		appSvc:     appSvc,
		jobRepo:    jobRepo,
		statusRepo: statusRepo,
		userID:     cfg.DiscordDefaultUserID,
		prefix:     cfg.DiscordPrefix,
	}
	session.AddHandler(b.messageCreate)
	session.AddHandler(b.reactionAdd)
	session.AddHandler(b.reactionRemove)
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
	cmd, args, _ := strings.Cut(strings.TrimSpace(strings.TrimPrefix(m.Content, b.prefix)), " ")
	switch cmd {
	case "ping":
		b.reply(m.ChannelID, "pong")
	case "add":
		b.handleAdd(m.ChannelID, args)
	case "edit":
		b.handleEdit(m.ChannelID, args)
	case "view":
		b.handleView(m.ChannelID, args)
	}
}

// addArgs is the parsed !add payload.
type addArgs struct{ title, company, desc, status string }

// parseAdd splits "title | company | description [| status]".
// ponytail: description must not contain '|' — documented MVP ceiling; switch to
// a quoted/flag parser only if descriptions with pipes show up.
func parseAdd(args string) (addArgs, bool) {
	parts := strings.Split(args, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	var a addArgs
	switch len(parts) {
	case 4:
		a.status = parts[3]
		fallthrough
	case 3:
		a.title, a.company, a.desc = parts[0], parts[1], parts[2]
	default:
		return a, false
	}
	if a.title == "" || a.company == "" || a.desc == "" {
		return a, false
	}
	return a, true
}

func (b *Bot) handleAdd(channelID, args string) {
	a, ok := parseAdd(args)
	if !ok {
		b.reply(channelID, "usage: "+b.prefix+"add <title> | <company> | <description> [| <status>]")
		return
	}

	// Always create a new job — same title+company can be a genuinely separate
	// opening, so duplicates are allowed (no Company entity; company lives on Job).
	job := &entity.Job{UserID: b.userID, Company: a.company, Position: a.title, Description: a.desc, UUID: uuid.New()}
	if err := b.jobRepo.Create(job); err != nil {
		log.Printf("discord add: job create failed: %v", err)
		b.reply(channelID, "failed: could not create job")
		return
	}

	sid, err := b.resolveStatusID(a.status)
	if err != nil {
		log.Printf("discord add: status lookup failed: %v", err)
		b.reply(channelID, "failed: status lookup error")
		return
	}

	app, err := b.appSvc.Create(b.userID, &dto.CreateApplicationRequest{JobID: job.ID, StatusID: sid, Text: ""})
	if err != nil {
		log.Printf("discord add: create application failed: %v", err)
		b.reply(channelID, "failed: could not create application")
		return
	}

	b.reply(channelID, fmt.Sprintf("Application added successfully\napplication_id: %d\njob_id: %d", app.ID, job.ID))
}

// parseEdit splits "<id> <status words…>" into the application id + status name.
func parseEdit(args string) (id int, status string, ok bool) {
	idStr, status, _ := strings.Cut(strings.TrimSpace(args), " ")
	status = strings.TrimSpace(status)
	id, err := strconv.Atoi(idStr)
	if err != nil || status == "" {
		return 0, "", false
	}
	return id, status, true
}

func (b *Bot) handleEdit(channelID, args string) {
	id, status, ok := parseEdit(args)
	if !ok {
		b.reply(channelID, "usage: "+b.prefix+"edit <application_id> <status>")
		return
	}

	sid, err := b.resolveStatusID(status)
	if err != nil {
		log.Printf("discord edit: status lookup failed: %v", err)
		b.reply(channelID, "failed: status lookup error")
		return
	}

	// appSvc.Update looks up by id and enforces ownership, so no repo call here.
	if _, err := b.appSvc.Update(b.userID, id, &dto.UpdateApplicationRequest{StatusID: &sid}); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			b.reply(channelID, "failed: application not found")
		case errors.Is(err, service.ErrForbidden):
			b.reply(channelID, "failed: not your application")
		default:
			log.Printf("discord edit: update failed: %v", err)
			b.reply(channelID, "failed: could not update application")
		}
		return
	}

	b.reply(channelID, "success")
}

const viewPageSize = 10

// Navigation reactions on a !view message.
const (
	prevEmoji = "◀️"
	nextEmoji = "▶️"
)

func (b *Bot) handleView(channelID, args string) {
	page := 1
	if fields := strings.Fields(args); len(fields) > 0 {
		if n, err := strconv.Atoi(fields[0]); err == nil && n > 0 {
			page = n
		}
	}

	res, err := b.appSvc.GetPage(b.userID, page, viewPageSize)
	if err != nil {
		log.Printf("discord view: getpage failed: %v", err)
		b.reply(channelID, "failed: could not load applications")
		return
	}

	msg, err := b.session.ChannelMessageSend(channelID, formatView(res))
	if err != nil {
		log.Printf("discord view: reply failed: %v", err)
		return
	}
	if res.Pagination.TotalPages <= 1 {
		return
	}
	for _, emoji := range []string{prevEmoji, nextEmoji} {
		if err := b.session.MessageReactionAdd(channelID, msg.ID, emoji); err != nil {
			log.Printf("discord view: add reaction failed: %v", err)
			return
		}
	}
}

// formatView renders a page of applications as plain text. The id is shown so
// it can be copied into !edit.
// ponytail: a full page (20 rows) can approach Discord's 2000-char message
// limit with long company/position names; lower viewPageSize or paginate the
// reply if that bites.
func formatView(res *dto.PaginatedApplicationsResponse) string {
	p := res.Pagination
	if len(res.Data) == 0 {
		return fmt.Sprintf("No applications on page %d.", p.Page)
	}
	var sb strings.Builder
	for _, app := range res.Data {
		company, position, status := "?", "?", "?"
		if app.Job != nil {
			company, position = app.Job.Company, app.Job.Position
		}
		if app.Status != nil {
			status = app.Status.Text
		}
		fmt.Fprintf(&sb, "#%d | %s @ %s | %s\n", app.ID, position, company, status)
	}
	fmt.Fprintf(&sb, "page %d/%d (%d total)", p.Page, p.TotalPages, p.Total)
	return sb.String()
}

var viewPageRe = regexp.MustCompile(`^page (\d+)/(\d+) \(`)

// parseViewPage reads the footer formatView writes back off a sent message.
// That footer is the only place the current page lives — there is no session
// state, so navigation survives restarts and needs no eviction.
func parseViewPage(content string) (page, total int, ok bool) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	m := viewPageRe.FindStringSubmatch(lines[len(lines)-1])
	if m == nil {
		return 0, 0, false
	}
	page, _ = strconv.Atoi(m[1])
	total, _ = strconv.Atoi(m[2])
	return page, total, true
}

func (b *Bot) reactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	b.navigate(r.MessageReaction)
}

func (b *Bot) reactionRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	b.navigate(r.MessageReaction)
}

// navigate re-renders a !view message one page in the direction of the emoji.
// ponytail: reacting and un-reacting both step a page — clearing the user's
// reaction so ▶ can be re-clicked would need MANAGE_MESSAGES and a delete call
// per click. Switch to reaction-removal if the toggle behaviour confuses people.
func (b *Bot) navigate(r *discordgo.MessageReaction) {
	self := b.session.State.User
	if self == nil || r.UserID == self.ID {
		return
	}
	var delta int
	switch r.Emoji.Name {
	case prevEmoji:
		delta = -1
	case nextEmoji:
		delta = 1
	default:
		return
	}

	msg, err := b.session.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		log.Printf("discord nav: fetch message failed: %v", err)
		return
	}
	if msg.Author == nil || msg.Author.ID != self.ID {
		return
	}
	page, total, ok := parseViewPage(msg.Content)
	if !ok {
		return
	}
	target := page + delta
	if target < 1 || target > total {
		return
	}

	res, err := b.appSvc.GetPage(b.userID, target, viewPageSize)
	if err != nil {
		log.Printf("discord nav: getpage failed: %v", err)
		return
	}
	if _, err := b.session.ChannelMessageEdit(r.ChannelID, r.MessageID, formatView(res)); err != nil {
		log.Printf("discord nav: edit failed: %v", err)
	}
}

// resolveStatusID maps a status name to its id, defaulting to 1 (Applied) when
// empty or unknown. Mirrors the notion_service add-flow, kept local.
func (b *Bot) resolveStatusID(name string) (int, error) {
	if strings.TrimSpace(name) == "" {
		return 1, nil
	}
	statuses, err := b.statusRepo.FindAll()
	if err != nil {
		return 0, err
	}
	for _, st := range statuses {
		if strings.EqualFold(st.Text, name) {
			return st.ID, nil
		}
	}
	return 1, nil
}

func (b *Bot) reply(channelID, msg string) {
	if _, err := b.session.ChannelMessageSend(channelID, msg); err != nil {
		log.Printf("discord: reply failed: %v", err)
	}
}
