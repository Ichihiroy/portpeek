package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Ichihiroy/portpeek/internal/ports"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("245"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("212"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	okStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	warnStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
)

type mode int

const (
	modeList mode = iota
	modeFilter
	modeConfirm
)

type portsMsg struct {
	ports []ports.Port
	err   error
}

type Model struct {
	all     []ports.Port
	visible []ports.Port
	cursor  int
	filter  string
	mode    mode
	status  string
	isErr   bool
	loading bool
	width   int
}

func New() Model {
	return Model{loading: true}
}

func (m Model) Init() tea.Cmd {
	return fetch
}

func fetch() tea.Msg {
	p, err := ports.List()
	return portsMsg{ports: p, err: err}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case portsMsg:
		m.loading = false
		if msg.err != nil {
			m.setStatus(msg.err.Error(), true)
			return m, nil
		}
		m.all = msg.ports
		m.applyFilter()
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modeFilter:
			return m.updateFilter(msg)
		case modeConfirm:
			return m.updateConfirm(msg)
		default:
			return m.updateList(msg)
		}
	}
	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Pasted or burst input arrives as one KeyMsg; "/abc" should open the filter with "abc".
	if len(msg.Runes) > 1 && msg.Runes[0] == '/' {
		m.mode = modeFilter
		m.filter = string(msg.Runes[1:])
		m.applyFilter()
		return m, nil
	}
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < len(m.visible)-1 {
			m.cursor++
		}
	case "r":
		m.loading = true
		m.setStatus("", false)
		return m, fetch
	case "/":
		m.mode = modeFilter
		m.setStatus("", false)
	case "esc":
		if m.filter != "" {
			m.filter = ""
			m.applyFilter()
		}
	case "k":
		if len(m.visible) > 0 {
			m.mode = modeConfirm
		}
	}
	return m, nil
}

func (m Model) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.mode = modeList
	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.applyFilter()
		}
	case "ctrl+c":
		return m, tea.Quit
	default:
		if len(msg.Runes) > 0 {
			m.filter += string(msg.Runes)
			m.applyFilter()
		}
	}
	return m, nil
}

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.mode = modeList
		p := m.visible[m.cursor]
		if err := ports.Kill(p.PID); err != nil {
			m.setStatus(err.Error(), true)
			return m, nil
		}
		m.setStatus(fmt.Sprintf("sent SIGTERM to %s (pid %d)", p.Process, p.PID), false)
		m.loading = true
		return m, fetch
	case "ctrl+c":
		return m, tea.Quit
	default:
		m.mode = modeList
	}
	return m, nil
}

func (m *Model) setStatus(s string, isErr bool) {
	m.status = s
	m.isErr = isErr
}

func (m *Model) applyFilter() {
	q := strings.ToLower(m.filter)
	m.visible = m.visible[:0]
	for _, p := range m.all {
		if q == "" ||
			strings.Contains(strings.ToLower(p.Process), q) ||
			strings.Contains(strconv.Itoa(p.Port), q) ||
			strings.Contains(strconv.Itoa(p.PID), q) {
			m.visible = append(m.visible, p)
		}
	}
	if m.cursor >= len(m.visible) {
		m.cursor = max(0, len(m.visible)-1)
	}
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("portpeek"))
	b.WriteString(dimStyle.Render("  listening TCP ports"))
	b.WriteString("\n\n")

	compact := m.width > 0 && m.width < 70
	rowFmt := func(port, pid, proc, user, addr string) string {
		if compact {
			return fmt.Sprintf("  %-6s  %-7s  %s", port, pid, proc)
		}
		return fmt.Sprintf("  %-6s  %-8s  %-20s  %-12s  %s", port, pid, proc, user, addr)
	}
	b.WriteString(headerStyle.Render(rowFmt("PORT", "PID", "PROCESS", "USER", "ADDR")))
	b.WriteString("\n")

	switch {
	case m.loading && len(m.all) == 0:
		b.WriteString(dimStyle.Render("  loading...\n"))
	case len(m.visible) == 0:
		b.WriteString(dimStyle.Render("  nothing listening\n"))
	default:
		for i, p := range m.visible {
			row := rowFmt(strconv.Itoa(p.Port), strconv.Itoa(p.PID), truncate(p.Process, 20), truncate(p.User, 12), p.Addr)
			if i == m.cursor {
				b.WriteString(selectedStyle.Render(row))
			} else {
				b.WriteString(row)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	switch m.mode {
	case modeFilter:
		b.WriteString(fmt.Sprintf("  / %s█\n", m.filter))
	case modeConfirm:
		p := m.visible[m.cursor]
		b.WriteString(warnStyle.Render(fmt.Sprintf("  kill %s (pid %d) on :%d? [y/N]", p.Process, p.PID, p.Port)))
		b.WriteString("\n")
	default:
		if m.filter != "" {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  filter: %s  (esc clears)\n", m.filter)))
		}
		if m.status != "" {
			if m.isErr {
				b.WriteString(errStyle.Render("  " + m.status))
			} else {
				b.WriteString(okStyle.Render("  " + m.status))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString(dimStyle.Render("\n  ↑/↓ move   k kill   r refresh   / filter   q quit\n"))
	if m.width > 0 {
		return lipgloss.NewStyle().MaxWidth(m.width).Render(b.String())
	}
	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
