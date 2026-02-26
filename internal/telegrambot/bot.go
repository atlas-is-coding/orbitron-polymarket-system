// Package telegrambot implements an interactive Telegram Bot that mirrors
// the Console TUI, synchronized via the shared EventBus.
package telegrambot

import (
	"context"
	"strconv"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"

	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/tui"
)

// OrderCanceler is the subset of TradesMonitor used by the bot for order management.
type OrderCanceler interface {
	CancelOrder(id string) error
	CancelAllOrders() error
}

// Bot is the interactive Telegram Bot.
type Bot struct {
	api      *tgbotapi.BotAPI
	bus      *tui.EventBus
	state    *BotState
	canceler OrderCanceler // optional; nil if TradesMonitor not running
	log      zerolog.Logger

	cfgMu   sync.RWMutex
	cfg     *config.Config
	cfgPath string

	adminID int64 // 0 means no admin configured
}

// New creates a new Bot.
// canceler may be nil if order management is not needed.
// log may be nil (uses nop logger).
// Returns (nil, nil) if bot_token is empty — caller must check.
func New(cfg *config.Config, cfgPath string, bus *tui.EventBus, canceler OrderCanceler, log *zerolog.Logger) (*Bot, error) {
	var adminID int64
	if cfg.Telegram.AdminChatID != "" {
		if id, err := strconv.ParseInt(cfg.Telegram.AdminChatID, 10, 64); err == nil {
			adminID = id
		}
	}

	l := zerolog.Nop()
	if log != nil {
		l = log.With().Str("component", "telegram-bot").Logger()
	}

	b := &Bot{
		bus:      bus,
		state:    NewBotState(),
		canceler: canceler,
		log:      l,
		cfg:      cfg,
		cfgPath:  cfgPath,
		adminID:  adminID,
	}

	if cfg.Telegram.BotToken == "" {
		return nil, nil //nolint:nilnil // intentional: no token = bot disabled
	}

	api, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, err
	}
	b.api = api
	return b, nil
}

// IsAllowed reports whether chatID is authorized to use the bot.
// Only the admin is allowed.
func (b *Bot) IsAllowed(chatID int64) bool {
	return b.adminID != 0 && chatID == b.adminID
}

// IsAdmin reports whether chatID has admin privileges (exported for testing).
func (b *Bot) IsAdmin(chatID int64) bool {
	return b.isAdmin(chatID)
}

func (b *Bot) isAdmin(chatID int64) bool {
	return b.adminID != 0 && chatID == b.adminID
}

// Run starts the bot. Blocks until ctx is cancelled.
func (b *Bot) Run(ctx context.Context) error {
	if b.api == nil {
		b.log.Warn().Msg("Telegram bot token not set, bot disabled")
		<-ctx.Done()
		return nil
	}

	b.log.Info().Str("username", b.api.Self.UserName).Msg("Telegram bot started")

	// Subscribe to EventBus — tap channel receives copies of all messages
	tap := b.bus.Tap()

	// EventBus consumer goroutine
	go b.consumeEvents(ctx, tap)

	// Telegram long-polling loop (blocks)
	return b.pollTelegram(ctx)
}

// consumeEvents reads from the EventBus tap and updates BotState.
func (b *Bot) consumeEvents(ctx context.Context, tap <-chan tea.Msg) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-tap:
			if !ok {
				return
			}
			b.processBusMsg(msg)
		}
	}
}

// processBusMsg applies an incoming EventBus message to BotState.
func (b *Bot) processBusMsg(msg tea.Msg) {
	switch m := msg.(type) {
	case tui.BalanceMsg:
		b.state.SetBalance(m.USDC)

	case tui.SubsystemStatusMsg:
		b.state.SetSubsystem(m.Name, m.Active)

	case tui.BotEventMsg:
		b.state.AddLog(m.Level + " " + m.Message)

	case tui.OrdersUpdateMsg:
		b.state.SetOrders(m.Rows)

	case tui.PositionsUpdateMsg:
		b.state.SetPositions(m.Rows)

	case tui.ConfigReloadedMsg:
		if m.Config != nil {
			b.cfgMu.Lock()
			b.cfg = m.Config
			b.cfgMu.Unlock()
		}
	}
}

// pollTelegram runs the getUpdates long-polling loop.
func (b *Bot) pollTelegram(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			b.log.Info().Msg("Telegram bot stopped")
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			b.handleUpdate(ctx, update)
		}
	}
}

// handleUpdate routes an incoming Telegram update.
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		if !b.IsAllowed(update.Message.Chat.ID) {
			b.log.Debug().Int64("chat_id", update.Message.Chat.ID).Msg("ignoring message from unauthorized chat")
			return
		}
		if update.Message.IsCommand() {
			b.handleCommand(ctx, update.Message)
		}
	case update.CallbackQuery != nil:
		if !b.IsAllowed(update.CallbackQuery.Message.Chat.ID) {
			return
		}
		b.handleCallback(ctx, update.CallbackQuery)
	}
}

// sendText sends a plain HTML text message.
func (b *Bot) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := b.api.Send(msg); err != nil {
		b.log.Warn().Err(err).Int64("chat_id", chatID).Msg("failed to send message")
	}
}

// sendWithKeyboard sends an HTML text message with an inline keyboard.
func (b *Bot) sendWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard
	if _, err := b.api.Send(msg); err != nil {
		b.log.Warn().Err(err).Int64("chat_id", chatID).Msg("failed to send keyboard message")
	}
}
