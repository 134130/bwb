package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fmt"
)

type ExplorerPage struct {
	Light  bool
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
	h := max(0, size.Height-3)
	if size.Width < 85 {
		if e.Focus == "detail" {
			return Size{}, Size{max(0, size.Width), h}
		}
		return Size{max(0, size.Width), h}, Size{}
	}
	left := max(28, size.Width*40/100)
	return Size{left, h}, Size{max(0, size.Width-left-1), h}
}
func (e *ExplorerPage) Update(msg tea.Msg, size Size) tea.Cmd {
	e.Search.setTheme(e.Light)
	if key, ok := msg.(tea.KeyPressMsg); ok {
		if e.Focus == "search" {
			if key.String() == "esc" || key.String() == "enter" || key.String() == "tab" || key.String() == "shift+tab" {
				e.Search.Blur()
				e.Focus = "tree"
				msg = LayoutMsg{}
			} else {
				cmd := e.Search.Update(msg, Size{max(0, size.Width-4), 1})
				left, _ := e.layout(size)
				e.Tree.rebuild(e.Search.Value(), inside(left))
				return cmd
			}
		} else {
			switch key.String() {
			case "/":
				e.Detail.clearValue()
				e.Focus = "search"
				return e.Search.Focus()
			case "tab", "shift+tab":
				if e.Focus == "tree" {
					e.Focus = "detail"
				} else {
					e.Focus = "tree"
				}
				msg = LayoutMsg{}
			case "esc":
				if e.Focus == "detail" {
					if e.Detail.Expanded {
						e.Detail.Expanded = false
						return nil
					}
					e.Focus = "tree"
				} else {
					e.Search.SetValue("")
					left, _ := e.layout(size)
					e.Tree.rebuild("", inside(left))
				}
				msg = LayoutMsg{}
			case "s":
				e.Tree.Reverse = !e.Tree.Reverse
				left, _ := e.layout(size)
				e.Tree.rebuild(e.Search.Value(), inside(left))
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
	e.Tree.Light, e.Detail.Light = e.Light, e.Light
	e.Search.Update(LayoutMsg{}, Size{max(0, size.Width-4), 1})
	// Changed layouts synchronize both stateful panes before rendering.
	if left.Width > 0 {
		e.Tree.Update(LayoutMsg{}, inside(left))
	}
	if right.Width > 0 {
		e.Detail.Update(LayoutMsg{}, inside(right))
	}
	if e.Focus == "tree" {
		return e.Tree.Update(msg, inside(left))
	}
	if e.Focus == "detail" {
		return e.Detail.Update(msg, inside(right))
	}
	return e.Search.Update(msg, Size{max(0, size.Width-4), 1})
}
func (e *ExplorerPage) View(size Size) string {
	left, right := e.layout(size)
	parts := []string{}
	ids := map[string]bool{}
	for _, row := range e.Tree.Rows {
		if row.Node.Item != nil {
			ids[row.Node.Item.ID] = true
		}
	}
	if left.Width > 0 {
		parts = append(parts, panel(fmt.Sprintf("Explorer · %d items", len(ids)), e.Tree.View(inside(left)), left, e.Focus == "tree", e.Light))
	}
	if right.Width > 0 {
		if left.Width > 0 {
			parts = append(parts, box("", Size{1, right.Height}))
		}
		parts = append(parts, panel("Item", e.Detail.View(inside(right)), right, e.Focus == "detail", e.Light))
	}
	search := panel("Search", e.Search.View(Size{max(0, size.Width-4), 1}), Size{size.Width, 3}, e.Focus == "search", e.Light)
	return box(search+"\n"+lipgloss.JoinHorizontal(lipgloss.Top, parts...), size)
}
func (e *ExplorerPage) hints() string {
	switch e.Focus {
	case "search":
		return "Enter return · Esc leave search · Tab pane"
	case "detail":
		return "j/k field · y copy · v reveal item · Enter expand · Tab pane · Esc back"
	default:
		return "↑↓ navigate · ←→ tree · / search · Enter open · v reveal item · Tab pane"
	}
}
