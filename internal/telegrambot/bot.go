// Package telegrambot implements an interactive Telegram Bot that mirrors
// the Console TUI, synchronized via the shared EventBus.
package telegrambot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"

	"github.com/atlasdev/polytrade-bot/internal/api/gamma"
	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/markets"
	"github.com/atlasdev/polytrade-bot/internal/tui"
)

// OrderCanceler is the subset of TradesMonitor used by the bot for order management.
type OrderCanceler interface {
	CancelOrder(id string) error
	CancelAllOrders() error
}

// WalletMutator allows the Telegram Bot to toggle wallet state.
// Implemented by *wallet.Manager.
type WalletMutator interface {
	Toggle(id string, enabled bool) error
	WalletEnabled(id string) bool
}

// MarketsProvider allows the bot to query markets data.
// Implemented by *markets.Service.
type MarketsProvider interface {
	GetByTag(slug string) []gamma.Market
	GetMarket(conditionID string) (gamma.Market, bool)
	Tags() []gamma.Tag
	AddAlert(rule markets.AlertRule) string
}

// Bot is the interactive Telegram Bot.
type Bot struct {
	api      *tgbotapi.BotAPI
	bus      *tui.EventBus
	state    *BotState
	canceler OrderCanceler  // optional; nil if TradesMonitor not running
	wallets  WalletMutator  // optional; nil if wallet manager unavailable
	mkts     MarketsProvider // optional; nil if Markets service not running
	log      zerolog.Logger

	cfgMu   sync.RWMutex
	cfg     *config.Config
	cfgPath string

	adminID int64 // 0 means no admin configured
}

// New creates a new Bot.
// canceler, wallets, and mkts may be nil.
// log may be nil (uses nop logger).
// Returns (nil, nil) if bot_token is empty — caller must check.
func New(cfg *config.Config, cfgPath string, bus *tui.EventBus, canceler OrderCanceler, wallets WalletMutator, mkts MarketsProvider, log *zerolog.Logger) (*Bot, error) {
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
		wallets:  wallets,
		mkts:     mkts,
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

	case tui.WalletAddedMsg:
		b.state.UpsertWallet(WalletEntry{ID: m.ID, Label: m.Label, Enabled: m.Enabled})

	case tui.WalletRemovedMsg:
		b.state.RemoveWallet(m.ID)

	case tui.WalletChangedMsg:
		wallets := b.state.Wallets()
		for _, w := range wallets {
			if w.ID == m.ID {
				w.Enabled = m.Enabled
				b.state.UpsertWallet(w)
				break
			}
		}

	case tui.WalletStatsMsg:
		b.state.UpsertWallet(WalletEntry{
			ID:      m.ID,
			Label:   m.Label,
			Enabled: m.Enabled,
			Balance: m.BalanceUSD,
			PnL:     m.PnLUSD,
		})

	case tui.LanguageChangedMsg:
		// no-op: Telegram bot uses hardcoded strings; handler present for completeness

	case tui.MarketAlertMsg:
		// Forward triggered market price alert as a notification message.
		text := fmt.Sprintf(
			"🔔 <b>Market Alert</b>\n\n"+
				"Price went <b>%s</b> threshold %.3f\n"+
				"Current: <b>%.3f</b>\n\n"+
				"<code>%s</code>",
			m.Direction, m.Threshold, m.CurrentPrice, m.ConditionID,
		)
		if m.Question != "" {
			text = fmt.Sprintf(
				"🔔 <b>Market Alert</b>\n\n"+
					"%s\n\nPrice went <b>%s</b> %.3f\nCurrent: <b>%.3f</b>",
				m.Question, m.Direction, m.Threshold, m.CurrentPrice,
			)
		}
		if b.adminID != 0 {
			b.sendText(b.adminID, text)
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
			b.state.ClearPending() // commands always reset conversation
			b.handleCommand(ctx, update.Message)
		} else if pi, _ := b.state.Pending(); pi != "" {
			b.handlePendingInput(ctx, update.Message)
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

// sendOrEdit edits the active menu message if menuMsgID is set,
// otherwise sends a new message and stores its ID.
func (b *Bot) sendOrEdit(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	if mid := b.state.MenuMsgID(); mid != 0 {
		edit := tgbotapi.NewEditMessageText(chatID, mid, text)
		edit.ParseMode = tgbotapi.ModeHTML
		edit.ReplyMarkup = &keyboard
		if _, err := b.api.Send(edit); err != nil {
			// Edit failed (e.g. message deleted) — fall back to new message
			b.state.SetMenuMsgID(0)
			b.sendOrEdit(chatID, text, keyboard)
		}
		return
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard
	sent, err := b.api.Send(msg)
	if err != nil {
		b.log.Warn().Err(err).Int64("chat_id", chatID).Msg("failed to send menu message")
		return
	}
	b.state.SetMenuMsgID(sent.MessageID)
}

// handlePendingInput routes incoming text to the active conversation step.
func (b *Bot) handlePendingInput(ctx context.Context, msg *tgbotapi.Message) {
	input, data := b.state.Pending()
	text := strings.TrimSpace(msg.Text)

	switch input {
	case "addtrader_addr":
		if text == "" {
			b.sendText(msg.Chat.ID, RenderError("Адрес не может быть пустым."))
			return
		}
		b.state.SetPending("addtrader_label", text)
		b.sendText(msg.Chat.ID, "🏷 Введите label (отображаемое имя) или <code>-</code> для пропуска:")

	case "addtrader_label":
		addr := data
		label := text
		if label == "-" {
			label = ""
		}
		b.state.SetPending("addtrader_alloc", addr+"|"+label)
		b.sendText(msg.Chat.ID, "📊 Введите allocation % (например <code>5</code>) или <code>-</code> для значения по умолчанию (5%):")

	case "addtrader_alloc":
		parts := strings.SplitN(data, "|", 2)
		addr := parts[0]
		label := ""
		if len(parts) > 1 {
			label = parts[1]
		}
		allocPct := 5.0
		if text != "-" && text != "" {
			if v, err := strconv.ParseFloat(text, 64); err == nil {
				allocPct = v
			}
		}
		b.state.ClearPending()
		args := []string{addr, label, strconv.FormatFloat(allocPct, 'f', 1, 64)}
		b.doAddTrader(ctx, msg.Chat.ID, args)
		b.sendCopytrading(msg.Chat.ID)

	case "alert_threshold":
		parts := strings.SplitN(data, "|", 2)
		if len(parts) != 2 {
			b.state.ClearPending()
			b.sendText(msg.Chat.ID, RenderError("Внутренняя ошибка: неверный формат данных алерта."))
			return
		}
		direction := parts[0]
		condID := parts[1]
		threshold, err := strconv.ParseFloat(text, 64)
		if err != nil || threshold <= 0 || threshold >= 1 {
			b.sendText(msg.Chat.ID, RenderError("Введите число от 0.01 до 0.99"))
			return
		}
		b.state.ClearPending()
		if b.mkts == nil {
			b.sendText(msg.Chat.ID, RenderError("Markets service unavailable"))
			return
		}
		alertID := b.mkts.AddAlert(markets.AlertRule{
			ConditionID: condID,
			Direction:   direction,
			Threshold:   threshold,
		})
		dirIcon := "📈"
		if direction == "below" {
			dirIcon = "📉"
		}
		b.sendText(msg.Chat.ID, RenderSuccess(fmt.Sprintf(
			"Алерт создан! %s Price %s <b>%.3f</b>\n<code>ID: %s</code>",
			dirIcon, direction, threshold, alertID,
		)))

	case "market_view":
		// User typed something while on market detail — ignore silently.

	default:
		// Generic setting edit: "edit:some.key"
		if strings.HasPrefix(input, "edit:") {
			key := strings.TrimPrefix(input, "edit:")
			b.state.ClearPending()
			b.doSetSetting(ctx, msg.Chat.ID, key, text)
		}
	}
}
