package tui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/orbitron/internal/i18n"
)

const maxLogLines = 500

// LogWriter implements io.Writer for zerolog; feeds lines into the EventBus.
type LogWriter struct {
	mu  sync.Mutex
	bus *EventBus
}

// NewLogWriter creates a LogWriter that sends log lines to the EventBus.
func NewLogWriter(bus *EventBus) *LogWriter {
	return &LogWriter{bus: bus}
}

func (w *LogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	line := strings.TrimRight(string(p), "\n")
	level := detectLevel(line)
	w.bus.Send(BotEventMsg{Level: level, Message: line})
	return len(p), nil
}

func detectLevel(line string) string {
	switch {
	case strings.Contains(line, `"ERR"`) || strings.Contains(line, ` ERR `):
		return "error"
	case strings.Contains(line, `"WRN"`) || strings.Contains(line, ` WRN `):
		return "warn"
	case strings.Contains(line, `"DBG"`) || strings.Contains(line, ` DBG `):
		return "debug"
	case strings.Contains(line, `"TRC"`) || strings.Contains(line, ` TRC `):
		return "trace"
	default:
		return "info"
	}
}

// LogsModel is the Logs tab sub-model.
type LogsModel struct {
	viewport viewport.Model
	lines    []BotEventMsg
	filter   string
	freeze   bool
	width    int
	height   int
	tick     int
}

// NewLogsModel creates a new LogsModel.
func NewLogsModel(width, height int) LogsModel {
	vp := viewport.New(width-4, max(height-6, 1))
	return LogsModel{viewport: vp, width: width, height: height}
}

// Resize updates the viewport size without destroying buffered log data.
func (m *LogsModel) Resize(w, h int) {
	m.width = w
	m.height = h
	m.viewport.Width = w - 4
	m.viewport.Height = max(h-6, 1)
	m.viewport.SetContent(m.renderLines())
}

func (m LogsModel) Init() tea.Cmd { return nil }

func (m LogsModel) Update(msg tea.Msg) (LogsModel, tea.Cmd) {
	switch msg.(type) {
	case animTickMsg:
		m.tick++
		return m, nil
	}
	switch msg := msg.(type) {
	case BotEventMsg:
		m.lines = append(m.lines, msg)
		if len(m.lines) > maxLogLines {
			m.lines = m.lines[len(m.lines)-maxLogLines:]
		}
		if !m.freeze {
			m.viewport.SetContent(m.renderLines())
			m.viewport.GotoBottom()
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+f":
			m.freeze = !m.freeze
			return m, nil
		case "t", "T":
			m.toggleFilter("trace")
			m.viewport.SetContent(m.renderLines())
			return m, nil
		case "d", "D":
			m.toggleFilter("debug")
			m.viewport.SetContent(m.renderLines())
			return m, nil
		case "i", "I":
			m.toggleFilter("info")
			m.viewport.SetContent(m.renderLines())
			return m, nil
		case "w", "W":
			m.toggleFilter("warn")
			m.viewport.SetContent(m.renderLines())
			return m, nil
		case "e", "E":
			m.toggleFilter("error")
			m.viewport.SetContent(m.renderLines())
			return m, nil
		case "c", "C":
			m.lines = m.lines[:0]
			m.viewport.SetContent("")
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *LogsModel) toggleFilter(level string) {
	if m.filter == level {
		m.filter = ""
	} else {
		m.filter = level
	}
}

func (m LogsModel) renderLines() string {
	var sb strings.Builder
	for _, l := range m.lines {
		if m.filter != "" && l.Level != m.filter {
			continue
		}
		sb.WriteString(" ")
		sb.WriteString(colorLogLine(l))
		sb.WriteString("\n")
	}
	return sb.String()
}

func colorLogLine(l BotEventMsg) string {
	switch l.Level {
	case "error":
		return StyleNegative.Render(l.Message)
	case "warn":
		return StyleWarning.Render(l.Message)
	case "info":
		return StyleBody.Render(l.Message)
	case "debug", "trace":
		return StyleMuted.Render(l.Message)
	default:
		return StyleBody.Render(l.Message)
	}
}

// colorLogLineStr colors a raw log line string by scanning for level keywords.
func colorLogLineStr(line string) string {
	upper := strings.ToUpper(line)
	switch {
	case strings.Contains(upper, "ERROR") || strings.Contains(upper, "FATAL"):
		return StyleNegative.Render(line)
	case strings.Contains(upper, "WARN"):
		return StyleWarning.Render(line)
	case strings.Contains(upper, "DEBUG"):
		return StyleMuted.Render(line)
	default:
		return StyleBody.Render(line)
	}
}

func (m LogsModel) View() string {
	t := i18n.T()
	freeze := ""
	if m.freeze {
		freeze = StyleWarning.Render(t.LogsFrozen)
	}
	filter := ""
	if m.filter != "" {
		filter = fmt.Sprintf("  %s%s", t.LogsFilter, m.filter)
	}

	logsPanel := renderPanel("Logs", m.viewport.View(), m.width, true)
	helpPanel := renderHelpPanel("↑↓=scroll | /=filter | c=clear | Tab=next-tab | q=quit"+freeze+filter, m.width)
	return lipgloss.JoinVertical(lipgloss.Left, " ", logsPanel, " ", helpPanel)
}
