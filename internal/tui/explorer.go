package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/134130/bwb/internal/bitwarden"
)

type revealMsg struct{ ID, Field string }
type Detail struct {
	Node                  *Node
	Fields                []bitwarden.FieldInfo
	Index, Offset, Scroll int
	Value                 string
	Revealed              bool
	Error                 string
}

func (d *Detail) clearValue() { d.Value = ""; d.Revealed = false; d.Scroll = 0 }
func (d *Detail) Update(msg tea.Msg, size Size) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "up", "k":
			d.clearValue()
			d.Index = max(0, d.Index-1)
		case "down", "j":
			d.clearValue()
			d.Index = min(max(0, len(d.Fields)-1), d.Index+1)
		case "pgdown":
			d.Scroll += max(1, size.Height/2)
		case "pgup":
			d.Scroll = max(0, d.Scroll-max(1, size.Height/2))
		case "enter":
			if d.Revealed {
				d.clearValue()
			} else if d.Node != nil && d.Node.Item != nil && len(d.Fields) > 0 {
				m := revealMsg{d.Node.Item.ID, d.Fields[d.Index].Path}
				return func() tea.Msg { return m }
			}
		}
	}
	if d.Revealed {
		d.Scroll = min(d.Scroll, max(0, len(d.valueLines(size))-max(1, size.Height-5)))
	}
	d.Offset = clampOffset(d.Index, d.Offset, max(1, size.Height-6))
	return nil
}
func (d *Detail) View(size Size) string {
	if d.Node == nil {
		return box("Select an item", size)
	}
	title := selected.Render(safe(d.Node.Name))
	if d.Node.Item == nil {
		return box(title+"\n\nFolder / collection\nID: "+safe(d.Node.ContainerID)+"\n→ expand  / search", size)
	}
	item := d.Node.Item
	title += "\n" + safe(item.Type)
	if item.Organization != nil {
		title += " · " + safe(item.Organization.Name)
	}
	names := []string{}
	for _, c := range item.Collections {
		names = append(names, safe(c.Name))
	}
	if len(names) > 0 {
		title += "\nCollections: " + strings.Join(names, ", ")
	}
	if d.Error != "" {
		return box(title+"\n\n"+safe(d.Error), size)
	}
	if d.Revealed {
		content := d.valueLines(size)
		start := min(d.Scroll, max(0, len(content)-1))
		return box(title+"\nEnter hide · PgUp/PgDown scroll\n\n"+strings.Join(content[start:], "\n"), size)
	}
	lines := []string{title, "", "Enter reveals selected field"}
	for i := d.Offset; i < min(len(d.Fields), d.Offset+max(1, size.Height-6)); i++ {
		line := safe(d.Fields[i].Label) + "  ••••••••"
		if i == d.Index {
			line = selected.Render("> " + line)
		}
		lines = append(lines, line)
	}
	return box(strings.Join(lines, "\n"), size)
}

type ExplorerPage struct {
	Search Input
	Tree   Tree
	Detail Detail
	Focus  string
}

func newExplorer() ExplorerPage {
	input := newInput(false)
	input.Placeholder = "Search vault…"
	input.Prompt = "/ "
	input.CharLimit = 256
	return ExplorerPage{Search: input, Focus: "tree"}
}
func (e *ExplorerPage) layout(size Size) (Size, Size) {
	h := max(0, size.Height-2)
	if size.Width < 85 {
		if e.Focus == "detail" {
			return Size{}, Size{max(0, size.Width), h}
		}
		return Size{max(0, size.Width), h}, Size{}
	}
	left := max(0, size.Width/2)
	return Size{left, h}, Size{max(0, size.Width-left-1), h}
}
func (e *ExplorerPage) Update(msg tea.Msg, size Size) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		if e.Focus == "search" {
			if key.String() == "esc" || key.String() == "enter" {
				e.Search.Blur()
				e.Focus = "tree"
				msg = LayoutMsg{}
			} else {
				cmd := e.Search.Update(msg, Size{size.Width, 1})
				left, _ := e.layout(size)
				e.Tree.rebuild(e.Search.Value(), left)
				return cmd
			}
		} else {
			switch key.String() {
			case "/":
				e.Focus = "search"
				return e.Search.Focus()
			case "tab":
				if e.Focus == "tree" {
					e.Focus = "detail"
				} else {
					e.Focus = "tree"
				}
				msg = LayoutMsg{}
			case "esc":
				if e.Focus == "detail" {
					e.Detail.clearValue()
					e.Focus = "tree"
				} else {
					e.Search.SetValue("")
					left, _ := e.layout(size)
					e.Tree.rebuild("", left)
				}
				msg = LayoutMsg{}
			case "s":
				e.Tree.Reverse = !e.Tree.Reverse
				left, _ := e.layout(size)
				e.Tree.rebuild(e.Search.Value(), left)
				msg = LayoutMsg{}
			case "enter":
				if e.Focus == "tree" && e.Tree.current() != nil && e.Tree.current().Item != nil {
					e.Focus = "detail"
					msg = LayoutMsg{}
				}
			}
		}
	}
	left, right := e.layout(size)
	e.Search.Update(LayoutMsg{}, Size{size.Width, 1})
	// Changed layouts synchronize both stateful panes before rendering.
	e.Tree.Update(LayoutMsg{}, left)
	e.Detail.Update(LayoutMsg{}, right)
	if e.Focus == "tree" {
		return e.Tree.Update(msg, left)
	}
	if e.Focus == "detail" {
		return e.Detail.Update(msg, right)
	}
	return e.Search.Update(msg, Size{size.Width, 1})
}
func (e *ExplorerPage) View(size Size) string {
	left, right := e.layout(size)
	parts := []string{}
	if left.Width > 0 {
		parts = append(parts, e.Tree.View(left))
	}
	if right.Width > 0 {
		if left.Width > 0 {
			parts = append(parts, box(strings.Repeat("│\n", max(0, right.Height)), Size{1, right.Height}))
		}
		parts = append(parts, e.Detail.View(right))
	}
	return box(e.Search.View(Size{size.Width, 1})+"\n"+strings.Repeat("─", max(0, size.Width))+"\n"+lipgloss.JoinHorizontal(lipgloss.Top, parts...), size)
}

func (d *Detail) valueLines(size Size) []string {
	lines := []string{}
	for _, line := range strings.Split(d.Value, "\n") {
		lines = append(lines, lipgloss.Wrap(safe(line), max(1, size.Width), ""))
	}
	return strings.Split(strings.Join(lines, "\n"), "\n")
}
