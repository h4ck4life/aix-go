package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(ColorInfo)
)

// item represents a selectable list item
type item struct {
	title       string
	description string
}

func (i item) FilterValue() string { return i.title }
func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }

// SelectModel is a list selection prompt
type SelectModel struct {
	list      list.Model
	choice    string
	done      bool
	cancelled bool
}

// NewSelect creates a new selection model
func NewSelect(title string, options []string) *SelectModel {
	items := make([]list.Item, len(options))
	for i, opt := range options {
		items[i] = item{title: opt}
	}

	l := list.New(items, list.NewDefaultDelegate(), 40, 20)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &SelectModel{list: l}
}

func (m *SelectModel) Init() tea.Cmd {
	return nil
}

func (m *SelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if i, ok := m.list.SelectedItem().(item); ok {
				m.choice = i.title
				m.done = true
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *SelectModel) View() string {
	if m.done || m.cancelled {
		return ""
	}
	return m.list.View()
}

// Choice returns the selected option
func (m SelectModel) Choice() string {
	return m.choice
}

// Cancelled returns true if the user cancelled
func (m SelectModel) Cancelled() bool {
	return m.cancelled
}

// RunSelect runs a selection prompt
func RunSelect(title string, options []string) (string, error) {
	m := NewSelect(title, options)
	result, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}
	if m, ok := result.(*SelectModel); ok {
		if m.Cancelled() {
			return "", fmt.Errorf("cancelled")
		}
		return m.Choice(), nil
	}
	return "", fmt.Errorf("unexpected result type")
}
