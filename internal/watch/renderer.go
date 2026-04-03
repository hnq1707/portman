package watch

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/hnq1707/portman/internal/port"
)

// --- Styles ---

var (
	wTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B")).
			Background(lipgloss.Color("#1a1a2e")).
			Padding(0, 2)

	wSubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	wTableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444"))

	wNewStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF88")).
			Bold(true)

	wGoneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4444")).
			Bold(true)

	wSwapStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Bold(true)

	wTimeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	wPortStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D4FF")).
			Bold(true)

	wDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	wHelpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	wHelpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	wCounterNewStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF88"))

	wCounterGoneStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF4444"))

	wCounterSwapStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFAA00"))
)

// --- Key bindings ---

type watchKeyMap struct {
	Quit    key.Binding
	Pause   key.Binding
	Clear   key.Binding
	Refresh key.Binding
}

func (k watchKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Pause, k.Refresh, k.Clear, k.Quit}
}

func (k watchKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

var watchKeys = watchKeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Pause: key.NewBinding(
		key.WithKeys("p", " "),
		key.WithHelp("p/space", "pause"),
	),
	Clear: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clear log"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh now"),
	),
}

// --- Messages ---

type watchTickMsg time.Time

type watchPollResult struct {
	events []Event
	ports  []port.PortInfo
	err    error
}

// --- Model ---

// WatchModel is the bubbletea model for the watch TUI.
type WatchModel struct {
	watcher  *Watcher
	table    table.Model
	help     help.Model
	ports    []port.PortInfo
	events   []Event
	width    int
	height   int
	paused   bool
	interval time.Duration

	// Event counters
	newCount  int
	goneCount int
	swapCount int
}

// NewWatchModel creates a new watch model.
func NewWatchModel(interval time.Duration) WatchModel {
	columns := []table.Column{
		{Title: "PROTO", Width: 6},
		{Title: "PORT", Width: 7},
		{Title: "PID", Width: 8},
		{Title: "PROCESS", Width: 20},
		{Title: "SERVICE", Width: 14},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(8),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("#FF6B6B"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#FF6B6B")).
		Bold(true)
	t.SetStyles(s)

	h := help.New()
	h.Styles.ShortKey = wHelpKeyStyle
	h.Styles.ShortDesc = wHelpDescStyle
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))

	return WatchModel{
		watcher:  NewWatcher(),
		table:    t,
		help:     h,
		interval: interval,
		width:    100,
		height:   30,
	}
}

// Init starts the model.
func (m WatchModel) Init() tea.Cmd {
	return tea.Batch(
		m.pollCmd(),
		m.tickCmd(),
	)
}

// Update handles messages.
func (m WatchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Adjust table height: reserve space for header + timeline + help
		tableH := max(msg.Height/3, 5)
		m.table.SetHeight(tableH)
		if msg.Width > 70 {
			m.table.SetColumns([]table.Column{
				{Title: "PROTO", Width: 6},
				{Title: "PORT", Width: 7},
				{Title: "PID", Width: 8},
				{Title: "PROCESS", Width: (msg.Width - 50) / 2},
				{Title: "SERVICE", Width: (msg.Width - 50) / 3},
			})
		}

	case watchTickMsg:
		if !m.paused {
			return m, tea.Batch(m.pollCmd(), m.tickCmd())
		}
		return m, m.tickCmd()

	case watchPollResult:
		if msg.err == nil {
			m.ports = msg.ports
			if len(msg.events) > 0 {
				m.events = append(m.events, msg.events...)
				// Cap at 50 visible events
				if len(m.events) > 50 {
					m.events = m.events[len(m.events)-50:]
				}
				// Update counters
				for _, e := range msg.events {
					switch e.Type {
					case PortAppeared:
						m.newCount++
					case PortDisappeared:
						m.goneCount++
					case ProcessChanged:
						m.swapCount++
					}
				}
			}
			m.updateTable()
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, watchKeys.Quit):
			return m, tea.Quit
		case key.Matches(msg, watchKeys.Pause):
			m.paused = !m.paused
		case key.Matches(msg, watchKeys.Clear):
			m.events = nil
			m.newCount = 0
			m.goneCount = 0
			m.swapCount = 0
		case key.Matches(msg, watchKeys.Refresh):
			return m, m.pollCmd()
		}
	}

	return m, nil
}

// View renders the UI.
func (m WatchModel) View() string {
	var b strings.Builder

	// ── Header ──
	title := wTitleStyle.Render("  👁 PORTMAN WATCH  ")
	portCount := wSubtitleStyle.Render(fmt.Sprintf("  %d port(s)", len(m.ports)))
	header := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", portCount)

	if m.paused {
		pauseBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000")).
			Background(lipgloss.Color("#FFAA00")).
			Bold(true).
			Padding(0, 1).
			Render("⏸ PAUSED")
		header = lipgloss.JoinHorizontal(lipgloss.Center, header, "  ", pauseBadge)
	}

	b.WriteString(header)
	b.WriteString("\n\n")

	// ── Compact port table ──
	b.WriteString(wTableStyle.Render(m.table.View()))
	b.WriteString("\n")

	// ── Event counters ──
	counters := fmt.Sprintf("  Events: %s  %s  %s",
		wCounterNewStyle.Render(fmt.Sprintf("🟢 %d new", m.newCount)),
		wCounterGoneStyle.Render(fmt.Sprintf("🔴 %d gone", m.goneCount)),
		wCounterSwapStyle.Render(fmt.Sprintf("🔄 %d swap", m.swapCount)),
	)
	b.WriteString(counters)
	b.WriteString("\n")

	// ── Timeline ──
	timelineHeader := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6B6B")).
		Render("  ── Activity Timeline ──")
	b.WriteString(timelineHeader)
	b.WriteString("\n")

	if len(m.events) == 0 {
		b.WriteString(wDimStyle.Render("  Watching for changes...\n"))
	} else {
		// Calculate how many events we can show
		maxEvents := m.height - 20
		if maxEvents < 5 {
			maxEvents = 5
		}
		start := 0
		if len(m.events) > maxEvents {
			start = len(m.events) - maxEvents
		}

		for _, e := range m.events[start:] {
			b.WriteString(m.renderEvent(e))
			b.WriteString("\n")
		}
	}

	// ── Help ──
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(m.help.View(watchKeys)))

	return b.String()
}

// renderEvent renders a single timeline event.
func (m WatchModel) renderEvent(e Event) string {
	timeStr := wTimeStyle.Render(e.Timestamp.Format("15:04:05"))
	portStr := wPortStyle.Render(fmt.Sprintf(":%d", e.Port))

	switch e.Type {
	case PortAppeared:
		label := wNewStyle.Render("🟢 NEW ")
		detail := fmt.Sprintf("%s  %s — %s (PID %d)", label, portStr, e.Process, e.PID)
		svc := port.GetPortLabel(e.Port)
		if svc != "" {
			detail += wDimStyle.Render(" [" + svc + "]")
		}
		return fmt.Sprintf("  %s  %s", timeStr, detail)

	case PortDisappeared:
		label := wGoneStyle.Render("🔴 GONE")
		detail := fmt.Sprintf("%s  %s — %s (PID %d)", label, portStr, e.Process, e.PID)
		return fmt.Sprintf("  %s  %s", timeStr, detail)

	case ProcessChanged:
		label := wSwapStyle.Render("🔄 SWAP")
		detail := fmt.Sprintf("%s  %s — %s → %s", label, portStr, e.OldProc, e.Process)
		return fmt.Sprintf("  %s  %s", timeStr, detail)

	default:
		return ""
	}
}

// --- Internal ---

func (m *WatchModel) updateTable() {
	rows := make([]table.Row, len(m.ports))
	for i, p := range m.ports {
		svc := port.GetPortLabel(p.Port)
		rows[i] = table.Row{
			p.Proto,
			strconv.Itoa(p.Port),
			strconv.Itoa(p.PID),
			p.ProcessName,
			svc,
		}
	}
	m.table.SetRows(rows)
}

func (m WatchModel) pollCmd() tea.Cmd {
	return func() tea.Msg {
		events, ports, err := m.watcher.Poll()
		return watchPollResult{events: events, ports: ports, err: err}
	}
}

func (m WatchModel) tickCmd() tea.Cmd {
	return tea.Tick(m.interval, func(t time.Time) tea.Msg {
		return watchTickMsg(t)
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
