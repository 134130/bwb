// Package tui owns terminal presentation; session policy remains in the daemon.
package tui

import (
	"strings"
	"unicode"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type Size struct{ Width, Height int }
type Component interface {
	Update(tea.Msg, Size) tea.Cmd
	View(Size) string
}
type LayoutMsg struct{}

var selected = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
var muted = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

// Vault data is untrusted terminal input, including escape and bidi controls.
func safe(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.In(r, unicode.Cf) {
			return '�'
		}
		return r
	}, s)
}
func box(s string, size Size) string {
	if size.Width <= 0 || size.Height <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) > size.Height {
		lines = lines[:size.Height]
	}
	for i := range lines {
		lines[i] = ansi.Truncate(lines[i], size.Width, "")
	}
	return lipgloss.NewStyle().Width(size.Width).Height(size.Height).Render(strings.Join(lines, "\n"))
}
func clampOffset(index, offset, height int) int {
	if index < offset {
		return index
	}
	if index >= offset+max(1, height) {
		return max(0, index-max(1, height)+1)
	}
	return max(0, offset)
}

// Input synchronizes the widget's cached width only during Update.
type Input struct{ textinput.Model }

func newInput(password bool) Input {
	m := textinput.New()
	m.CharLimit = 256
	m.KeyMap.Paste.SetEnabled(false)
	if password {
		m.CharLimit = 4096
		m.EchoMode = textinput.EchoPassword
		m.Prompt = "Password: "
	}
	return Input{m}
}
func (i *Input) Update(msg tea.Msg, size Size) tea.Cmd {
	i.Model.SetWidth(max(1, size.Width-12))
	var cmd tea.Cmd
	i.Model, cmd = i.Model.Update(msg)
	return cmd
}
func (i *Input) View(size Size) string { return box(i.Model.View(), size) }
