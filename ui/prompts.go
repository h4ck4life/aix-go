package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// InputModel is a reusable text input prompt
type InputModel struct {
	prompt      string
	placeholder string
	textInput   textinput.Model
	done        bool
	cancelled   bool
}

// NewInput creates a new input prompt
func NewInput(prompt, placeholder string) *InputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return &InputModel{
		prompt:      prompt,
		placeholder: placeholder,
		textInput:   ti,
	}
}

func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m InputModel) View() string {
	if m.done || m.cancelled {
		return ""
	}
	return fmt.Sprintf("%s\n%s\n", Info(m.prompt), m.textInput.View())
}

// Value returns the input value
func (m InputModel) Value() string {
	return m.textInput.Value()
}

// Cancelled returns true if the user cancelled
func (m InputModel) Cancelled() bool {
	return m.cancelled
}

// SecureInputModel is a password input prompt
type SecureInputModel struct {
	prompt    string
	textInput textinput.Model
	done      bool
	cancelled bool
}

// NewSecureInput creates a new secure input prompt
func NewSecureInput(prompt string) *SecureInputModel {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.CharLimit = 512
	ti.Width = 50

	return &SecureInputModel{
		prompt:    prompt,
		textInput: ti,
	}
}

func (m SecureInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SecureInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m SecureInputModel) View() string {
	if m.done || m.cancelled {
		return ""
	}
	return fmt.Sprintf("%s\n%s\n", Info(m.prompt), m.textInput.View())
}

// Value returns the input value
func (m SecureInputModel) Value() string {
	return m.textInput.Value()
}

// Cancelled returns true if the user cancelled
func (m SecureInputModel) Cancelled() bool {
	return m.cancelled
}

// ConfirmModel is a yes/no confirmation prompt
type ConfirmModel struct {
	prompt    string
	defaultYes bool
	value     bool
	done      bool
	cancelled bool
}

// NewConfirm creates a new confirmation prompt
func NewConfirm(prompt string, defaultYes bool) *ConfirmModel {
	return &ConfirmModel{
		prompt:     prompt,
		defaultYes: defaultYes,
		value:      defaultYes,
	}
}

func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		default:
			s := strings.ToLower(msg.String())
			switch s {
			case "y":
				m.value = true
				m.done = true
				return m, tea.Quit
			case "n":
				m.value = false
				m.done = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m ConfirmModel) View() string {
	if m.done || m.cancelled {
		return ""
	}
	prompt := m.prompt
	if m.defaultYes {
		prompt += " [Y/n]"
	} else {
		prompt += " [y/N]"
	}
	return fmt.Sprintf("%s ", Info(prompt))
}

// Value returns the confirmation result
func (m ConfirmModel) Value() bool {
	return m.value
}

// Cancelled returns true if the user cancelled
func (m ConfirmModel) Cancelled() bool {
	return m.cancelled
}

// RunPrompt runs a Bubble Tea program and returns the result
func RunPrompt(p tea.Model) (tea.Model, error) {
	return tea.NewProgram(p).Run()
}
