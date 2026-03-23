package ui

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

	"github.com/nay-kia/portman/internal/port"
)

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D4FF")).
			Background(lipgloss.Color("#1a1a2e")).
			Padding(0, 2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333")).
			Background(lipgloss.Color("#00D4FF")).
			Padding(0, 1).
			Bold(true)

	helpBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF88")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4444")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Bold(true)

	tableBaseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(0, 0)

	confirmStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Bold(true).
			Padding(0, 1)
)

// --- Key bindings ---

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Kill    key.Binding
	Filter  key.Binding
	Refresh key.Binding
	Quit    key.Binding
	Escape  key.Binding
	Confirm key.Binding
	Deny    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Kill, k.Filter, k.Refresh, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Kill: key.NewBinding(
		key.WithKeys("K", "enter"),
		key.WithHelp("K/⏎", "kill"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "yes"),
	),
	Deny: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "no"),
	),
}

// --- Messages ---

type tickMsg time.Time

type portsRefreshedMsg struct {
	ports []port.PortInfo
}

type killResultMsg struct {
	port    int
	name    string
	success bool
	err     string
}

// --- Model ---

// DashboardModel is the bubbletea model for the TUI dashboard.
type DashboardModel struct {
	table    table.Model
	help     help.Model
	ports    []port.PortInfo
	filtered []port.PortInfo
	width    int
	height   int

	// State
	filterMode bool
	filterText string
	message    string
	messageAt  time.Time

	// Kill confirmation
	confirming    bool
	confirmPort   int
	confirmPID    int
	confirmName   string
}

// NewDashboard creates a new dashboard model.
func NewDashboard() DashboardModel {
	columns := []table.Column{
		{Title: "PROTO", Width: 6},
		{Title: "PORT", Width: 7},
		{Title: "ADDRESS", Width: 24},
		{Title: "PID", Width: 8},
		{Title: "PROCESS", Width: 22},
		{Title: "NOTE", Width: 16},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("#00D4FF"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#00D4FF")).
		Bold(true)
	t.SetStyles(s)

	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4FF")).Bold(true)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))

	return DashboardModel{
		table: t,
		help:  h,
		width: 90,
	}
}

// Init starts the initial commands.
func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		m.refreshPorts(),
		m.tickCmd(),
	)
}

// Update handles messages.
func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetHeight(max(msg.Height-10, 5))
		// Adjust column widths
		remaining := msg.Width - 12 // borders + padding
		if remaining > 90 {
			m.table.SetColumns([]table.Column{
				{Title: "PROTO", Width: 6},
				{Title: "PORT", Width: 7},
				{Title: "ADDRESS", Width: remaining/4},
				{Title: "PID", Width: 8},
				{Title: "PROCESS", Width: remaining/4},
				{Title: "NOTE", Width: remaining/5},
			})
		}

	case tickMsg:
		if !m.confirming {
			cmds = append(cmds, m.refreshPorts())
		}
		cmds = append(cmds, m.tickCmd())

	case portsRefreshedMsg:
		m.ports = msg.ports
		m.applyFilter()

	case killResultMsg:
		if msg.success {
			m.message = successStyle.Render(fmt.Sprintf("  ✓ Killed %s (PID %d) on port %d", msg.name, 0, msg.port))
		} else {
			m.message = errorStyle.Render(fmt.Sprintf("  ✗ Failed: %s", msg.err))
		}
		m.messageAt = time.Now()
		m.confirming = false
		cmds = append(cmds, m.refreshPorts())

	case tea.KeyMsg:
		// Clear old messages
		if !m.messageAt.IsZero() && time.Since(m.messageAt) > 5*time.Second {
			m.message = ""
		}

		// Handle confirm dialog
		if m.confirming {
			switch {
			case key.Matches(msg, keys.Confirm):
				m.confirming = false
				cmds = append(cmds, m.killPort(m.confirmPort))
			default:
				m.confirming = false
				m.message = subtitleStyle.Render("  ↩ Cancelled")
				m.messageAt = time.Now()
			}
			return m, tea.Batch(cmds...)
		}

		// Handle filter mode
		if m.filterMode {
			switch {
			case key.Matches(msg, keys.Escape):
				m.filterMode = false
				m.filterText = ""
				m.applyFilter()
			case msg.Type == tea.KeyBackspace:
				if len(m.filterText) > 0 {
					m.filterText = m.filterText[:len(m.filterText)-1]
					m.applyFilter()
				}
			case msg.Type == tea.KeyEnter:
				m.filterMode = false
			case msg.Type == tea.KeyRunes:
				m.filterText += string(msg.Runes)
				m.applyFilter()
			}
			return m, nil
		}

		// Normal mode keys
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Filter):
			m.filterMode = true
			m.filterText = ""
			return m, nil
		case key.Matches(msg, keys.Escape):
			m.filterText = ""
			m.applyFilter()
		case key.Matches(msg, keys.Refresh):
			m.message = subtitleStyle.Render("  ↻ Refreshing...")
			m.messageAt = time.Now()
			cmds = append(cmds, m.refreshPorts())
		case key.Matches(msg, keys.Kill):
			if row := m.table.SelectedRow(); row != nil {
				portNum, _ := strconv.Atoi(row[1])
				pid, _ := strconv.Atoi(row[3])
				m.confirming = true
				m.confirmPort = portNum
				m.confirmPID = pid
				m.confirmName = row[4]
			}
			return m, nil
		}
	}

	// Update table
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the UI.
func (m DashboardModel) View() string {
	var b strings.Builder

	// Header
	title := titleStyle.Render("  🚀 PORTMAN DASHBOARD  ")
	count := subtitleStyle.Render(fmt.Sprintf("  %d port(s)", len(m.filtered)))
	header := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", count)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Filter bar
	if m.filterMode {
		filterBar := warningStyle.Render(fmt.Sprintf("  🔍 Filter: %s▋", m.filterText))
		b.WriteString(filterBar)
		b.WriteString("\n")
	} else if m.filterText != "" {
		filterBar := subtitleStyle.Render(fmt.Sprintf("  🔍 Filter: %s", m.filterText))
		b.WriteString(filterBar)
		b.WriteString("\n")
	}

	// Table
	b.WriteString(tableBaseStyle.Render(m.table.View()))
	b.WriteString("\n")

	// Confirm dialog
	if m.confirming {
		confirm := confirmStyle.Render(
			fmt.Sprintf("  ⚠ Kill %s (PID %d) on port %d? [y/n]",
				m.confirmName, m.confirmPID, m.confirmPort))
		b.WriteString(confirm)
		b.WriteString("\n")
	}

	// Status message
	if m.message != "" {
		b.WriteString(m.message)
		b.WriteString("\n")
	}

	// Help bar
	b.WriteString("\n")
	b.WriteString(helpBarStyle.Render(m.help.View(keys)))

	return b.String()
}

// --- Helpers ---

// Well-known ports for notes
var knownPorts = map[int]string{
	80: "HTTP", 443: "HTTPS", 3000: "Dev",
	3306: "MySQL", 5432: "Postgres", 6379: "Redis",
	8080: "HTTP Alt", 8443: "HTTPS Alt", 9200: "Elastic",
	9092: "Kafka", 27017: "MongoDB",
}

func (m *DashboardModel) applyFilter() {
	if m.filterText == "" {
		m.filtered = m.ports
	} else {
		lower := strings.ToLower(m.filterText)
		var filtered []port.PortInfo
		for _, p := range m.ports {
			if strings.Contains(strings.ToLower(p.ProcessName), lower) ||
				strings.Contains(strconv.Itoa(p.Port), m.filterText) ||
				strings.Contains(strings.ToLower(p.Proto), lower) {
				filtered = append(filtered, p)
			}
		}
		m.filtered = filtered
	}
	m.updateTableRows()
}

func (m *DashboardModel) updateTableRows() {
	rows := make([]table.Row, len(m.filtered))
	for i, p := range m.filtered {
		note := ""
		if label, ok := knownPorts[p.Port]; ok {
			note = "⭐ " + label
		}
		rows[i] = table.Row{
			p.Proto,
			strconv.Itoa(p.Port),
			p.LocalAddr,
			strconv.Itoa(p.PID),
			p.ProcessName,
			note,
		}
	}
	m.table.SetRows(rows)
}

func (m DashboardModel) refreshPorts() tea.Cmd {
	return func() tea.Msg {
		ports, err := port.ScanPorts()
		if err != nil {
			return portsRefreshedMsg{ports: nil}
		}
		return portsRefreshedMsg{ports: ports}
	}
}

func (m DashboardModel) killPort(portNum int) tea.Cmd {
	return func() tea.Msg {
		results := port.KillByPort(portNum)
		if len(results) == 0 {
			return killResultMsg{port: portNum, success: false, err: "no process found"}
		}
		r := results[0]
		return killResultMsg{
			port:    portNum,
			name:    r.ProcessName,
			success: r.Success,
			err:     r.Error,
		}
	}
}

func (m DashboardModel) tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
