package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// SpinnerModel wraps a spinner with a message
type SpinnerModel struct {
	spinner spinner.Model
	message string
	done    bool
	result  string
	err     error
}

// NewSpinner creates a new spinner model
func NewSpinner(message string) *SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = InfoStyle

	return &SpinnerModel{
		spinner: s,
		message: message,
	}
}

func (m SpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m SpinnerModel) View() string {
	if m.done {
		if m.err != nil {
			return Error(fmt.Sprintf("%s failed: %v\n", m.message, m.err))
		}
		return Success(fmt.Sprintf("%s done%s\n", m.message, m.result))
	}
	return fmt.Sprintf("%s %s...", m.spinner.View(), m.message)
}

// SetDone marks the spinner as complete
func (m *SpinnerModel) SetDone(result string, err error) {
	m.done = true
	m.result = result
	m.err = err
}

// RunSpinner runs a spinner until the done function returns true
func RunSpinner(message string, done func() (string, error)) error {
	m := NewSpinner(message)
	p := tea.NewProgram(m)

	go func() {
		result, err := done()
		m.SetDone(result, err)
		p.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	}()

	_, err := p.Run()
	return err
}
