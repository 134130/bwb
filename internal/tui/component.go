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
	i.SetWidth(max(1, size.Width-ansi.StringWidth(i.Prompt)))
	var cmd tea.Cmd
	i.Model, cmd = i.Model.Update(msg)
	return cmd
}
func (i *Input) View(size Size) string { return box(i.Model.View(), size) }

// Colors are local to each model; terminal background reports never mutate globals.
type palette struct{ foreground, muted, border, selection string }

func colors(light bool) palette {
	if light {
		return palette{"#252a31", "#606873", "#cbd0d6", "#e0e8ef"}
	}
	return palette{"#dce0e5", "#9299a3", "#424953", "#303a46"}
}
func subdued(s string, light bool) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(colors(light).muted)).Render(s)
}
func selection(s string, width int, light bool) string {
	p := colors(light)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(p.foreground)).Background(lipgloss.Color(p.selection)).Render(box(ansi.Strip(s), Size{width, 1}))
}
func inside(size Size) Size { return Size{max(0, size.Width-4), max(0, size.Height-2)} }
func panel(title, body string, size Size, active, light bool) string {
	if size.Width < 5 || size.Height < 3 {
		return box("", size)
	}
	p := colors(light)
	color := p.border
	if active {
		color = p.foreground
	}
	border := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	label := ansi.Truncate(" "+safe(title)+" ", size.Width-4, "")
	labelColor := p.muted
	if active {
		labelColor = p.foreground
	}
	top := border.Render("╭─") + lipgloss.NewStyle().Foreground(lipgloss.Color(labelColor)).Render(label) + border.Render(strings.Repeat("─", max(0, size.Width-3-ansi.StringWidth(label)))+"╮")
	lines := []string{top}
	for _, line := range strings.Split(box(body, inside(size)), "\n") {
		lines = append(lines, border.Render("│")+" "+line+" "+border.Render("│"))
	}
	lines = append(lines, border.Render("╰"+strings.Repeat("─", size.Width-2)+"╯"))
	return box(strings.Join(lines, "\n"), size)
}
func dialog(title, content, background string, size Size, light bool) string {
	width := min(72, size.Width)
	height := min(size.Height, lipgloss.Height(lipgloss.Wrap(content, max(1, width-4), ""))+4)
	foreground := panel(title, lipgloss.Wrap(content, max(1, width-4), ""), Size{width, height}, true, light)
	return box(lipgloss.NewCompositor(lipgloss.NewLayer(subdued(ansi.Strip(background), light)), lipgloss.NewLayer(foreground).X(max(0, (size.Width-width)/2)).Y(max(0, (size.Height-height)/2))).Render(), size)
}

func (i *Input) setTheme(light bool) {
	p := colors(light)
	style := textinput.DefaultStyles(!light)
	style.Focused.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color(p.foreground))
	style.Focused.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color(p.muted))
	style.Focused.Text = lipgloss.NewStyle().Foreground(lipgloss.Color(p.foreground))
	style.Blurred = style.Focused
	style.Cursor.Color = lipgloss.Color(p.foreground)
	i.SetStyles(style)
}
