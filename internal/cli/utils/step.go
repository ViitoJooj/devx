package utils

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

type stepDone struct{ err error }

type stepModel struct {
	spinner spinner.Model
	label   string
	err     error
	done    bool
}

func (m stepModel) Init() tea.Cmd { return m.spinner.Tick }

func (m stepModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stepDone:
		m.done, m.err = true, msg.err
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Sequence(tea.Quit, func() tea.Msg { os.Exit(130); return nil })
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m stepModel) View() string {
	switch {
	case !m.done:
		return m.spinner.View() + " " + m.label + "\n"
	case m.err != nil:
		return "✗ " + m.label + "\n"
	}
	return "✓ " + m.label + "\n"
}

// Step runs fn while showing a spinner with label, and leaves a ✓ or ✗ line behind.
// Without a terminal (CI, pipes) it prints plain lines instead of the animation.
func Step(label string, fn func() error) error {

	if !term.IsTerminal(int(os.Stdout.Fd())) {
		err := fn()
		if err != nil {
			fmt.Println("failed: " + label)
			return err
		}
		fmt.Println("done: " + label)
		return nil
	}

	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	program := tea.NewProgram(stepModel{spinner: s, label: label})

	go func() { program.Send(stepDone{err: fn()}) }()

	final, err := program.Run()
	if err != nil {
		return err
	}

	return final.(stepModel).err
}
