package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
)

func colorizeLevel(level string) string {
	switch strings.ToLower(level) {
	case "info":
		return ColorGreen + level + ColorReset
	case "error":
		return ColorRed + level + ColorReset
	case "fatal":
		return ColorYellow + level + ColorReset
	default:
		return ColorBlue + level + ColorReset
	}
}

type keyMap struct {
	filterEmail  key.Binding
	filterNextjs key.Binding
	clearOrigin  key.Binding
	filterInfo   key.Binding
	filterError  key.Binding
	filterFatal  key.Binding
	clearLevel   key.Binding
	quit         key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		filterEmail:  key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "origin: email")),
		filterNextjs: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "origin: nextjs")),
		clearOrigin:  key.NewBinding(key.WithKeys("0"), key.WithHelp("0", "clear origin")),
		filterInfo:   key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "level: info")),
		filterError:  key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "level: error")),
		filterFatal:  key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "level: fatal")),
		clearLevel:   key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "clear level")),
		quit:         key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

type logItem struct {
	entry LogEntry
}

func (i logItem) Title() string {
	prefix := i.entry.Message
	if idx := strings.Index(i.entry.Message, ":"); idx != -1 {
		prefix = i.entry.Message[:idx]
	}
	origin := fmt.Sprintf("[%s]", i.entry.Source)
	level := fmt.Sprintf("[%s]", colorizeLevel(i.entry.Level))
	return fmt.Sprintf("%-8s %-10s %s", origin, level, prefix)
}

func (i logItem) Description() string {
	return i.entry.Timestamp
}

func (i logItem) FilterValue() string {
	return i.entry.Message
}

type model struct {
	list         list.Model
	keys         *keyMap
	activeOrigin string
	activeLevel  string
	allItems     []list.Item
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.filterEmail):
			m.activeOrigin = "email"
			m.applyFilters()
		case key.Matches(msg, m.keys.filterNextjs):
			m.activeOrigin = "nextjs"
			m.applyFilters()
		case key.Matches(msg, m.keys.clearOrigin):
			m.activeOrigin = ""
			m.applyFilters()
		case key.Matches(msg, m.keys.filterInfo):
			m.activeLevel = "info"
			m.applyFilters()
		case key.Matches(msg, m.keys.filterError):
			m.activeLevel = "error"
			m.applyFilters()
		case key.Matches(msg, m.keys.filterFatal):
			m.activeLevel = "fatal"
			m.applyFilters()
		case key.Matches(msg, m.keys.clearLevel):
			m.activeLevel = ""
			m.applyFilters()
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	width := m.list.Width()
	height := m.list.Height()

	logListView := m.list.View()

	var detailsView string
	if selected, ok := m.list.SelectedItem().(logItem); ok {
		entry := selected.entry

		var logMap map[string]interface{}
		if err := json.Unmarshal([]byte(entry.Raw), &logMap); err != nil {
			detailsView = fmt.Sprintf("Failed to parse log JSON: %v", err)
		} else {
			builder := &strings.Builder{}
			builder.WriteString("\n\n\n")

			coreFields := []string{"level", "timestamp", "message"}

			maxKeyLen := 0
			for k := range logMap {
				if len(k) > maxKeyLen {
					maxKeyLen = len(k)
				}
			}

			for _, k := range coreFields {
				val, ok := logMap[k]
				if !ok {
					continue
				}

				keyLabel := fmt.Sprintf("%-*s", maxKeyLen, k)
				var valueStr string
				if k == "level" {
					valueStr = colorizeLevel(fmt.Sprint(val))
				} else {
					valueStr = fmt.Sprint(val)
				}

				fmt.Fprintf(builder, "%s: %s\n\n", keyLabel, valueStr)
				delete(logMap, k)
			}

			fmt.Fprintln(builder, strings.Repeat("─", maxKeyLen+40))
			fmt.Fprintln(builder)

			keys := make([]string, 0, len(logMap))
			for k := range logMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				val := logMap[k]
				keyLabel := fmt.Sprintf("%-*s", maxKeyLen, k)
				valStr := fmt.Sprint(val)

				if strings.Contains(valStr, "\n") {
					fmt.Fprintf(builder, "%s:\n", keyLabel)
					for _, line := range strings.Split(valStr, "\n") {
						fmt.Fprintf(builder, "  %s\n", line)
					}
					fmt.Fprintln(builder)
				} else {
					fmt.Fprintf(builder, "%s: %s\n\n", keyLabel, valStr)
				}
			}

			detailsView = builder.String()
		}
	} else {
		detailsView = "\n\nNo log selected"
	}

	// === Styles ===
	columnGap := 1
	listWidth := width * 35 / 100
	detailWidth := width - listWidth - columnGap - 1 // -1 for separator width

	listStyle := lipgloss.NewStyle().Width(listWidth).Height(height)
	detailStyle := lipgloss.NewStyle().
		Width(detailWidth).
		Height(height).
		PaddingLeft(1)

	separatorLines := make([]string, height)
	for i := range separatorLines {
		separatorLines[i] = "│"
	}
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		PaddingLeft(0).PaddingRight(1).
		Render(strings.Join(separatorLines, "\n"))

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listStyle.Render(logListView),
		separator,
		detailStyle.Render(detailsView),
	)

	return docStyle.Render(content)
}

func (m *model) applyFilters() {
	var filtered []list.Item
	for _, item := range m.allItems {
		entry := item.(logItem).entry
		if m.activeOrigin != "" && entry.Source != m.activeOrigin {
			continue
		}
		if m.activeLevel != "" && strings.ToLower(entry.Level) != m.activeLevel {
			continue
		}
		filtered = append(filtered, item)
	}
	m.list.SetItems(filtered)
}
