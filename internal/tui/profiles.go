package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/134130/bwb/internal/config"
	"github.com/134130/bwb/internal/vaultapi"
)

type profileRow struct {
	Name   string
	Status vaultapi.StatusResponse
	File   *config.Profile
}
type ProfilesPage struct {
	Light           bool
	DetailFocus     bool
	Scroll          int
	Rows            []profileRow
	Index, Offset   int
	Path, FileError string
}

func (p *ProfilesPage) Update(msg tea.Msg, size Size) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "tab", "shift+tab":
			p.DetailFocus = !p.DetailFocus
		case "esc":
			p.DetailFocus = false
		case "up", "k":
			if p.DetailFocus {
				p.Scroll = max(0, p.Scroll-1)
			} else {
				p.Index = max(0, p.Index-1)
				p.Scroll = 0
			}
		case "down", "j":
			if p.DetailFocus {
				p.Scroll++
			} else {
				p.Index = min(max(0, len(p.Rows)-1), p.Index+1)
				p.Scroll = 0
			}
		case "pgdown":
			if p.DetailFocus {
				p.Scroll += max(1, size.Height/2)
			}
		case "pgup":
			if p.DetailFocus {
				p.Scroll = max(0, p.Scroll-max(1, size.Height/2))
			}
		}
	}
	left, right := p.layout(size)
	if left.Width > 0 {
		p.Offset = clampOffset(p.Index, p.Offset, inside(left).Height)
	}
	if right.Width > 0 {
		p.Scroll = min(p.Scroll, max(0, len(p.detailLines(inside(right)))-inside(right).Height))
	}
	return nil
}

func (p *ProfilesPage) current() *profileRow {
	if len(p.Rows) == 0 {
		return nil
	}
	return &p.Rows[min(p.Index, len(p.Rows)-1)]
}
func (p *ProfilesPage) View(size Size) string {
	left, right := p.layout(size)
	listSize := inside(left)
	lines := []string{}
	for i := p.Offset; i < min(len(p.Rows), p.Offset+max(0, listSize.Height)); i++ {
		row := p.Rows[i]
		line := safe(row.Name) + "  " + row.Status.State
		if i == p.Index {
			line = selection("> "+line, listSize.Width, p.Light)
		}
		lines = append(lines, line)
	}

	list := panel("Profiles", strings.Join(lines, "\n"), left, !p.DetailFocus, p.Light)
	detail := p.detailLines(inside(right))
	settings := panel("Applied configuration", strings.Join(detail[min(p.Scroll, len(detail)):], "\n"), right, p.DetailFocus, p.Light)
	if left.Width == 0 {
		return settings
	}
	if right.Width == 0 {
		return list
	}
	return box(lipgloss.JoinHorizontal(lipgloss.Top, list, " ", settings), size)
}
func (p *ProfilesPage) layout(size Size) (Size, Size) {
	if size.Width < 85 {
		if p.DetailFocus {
			return Size{}, size
		}
		return size, Size{}
	}
	left := min(32, max(0, size.Width/3))
	return Size{left, size.Height}, Size{max(0, size.Width-left-1), size.Height}
}

func (p *ProfilesPage) detailLines(size Size) []string {
	detail := "No profiles"
	if row := p.current(); row != nil {
		d := row.Status
		detail = fmt.Sprintf("%s\nState: %s\nServer: %s\nEmail: %s\nApplied timeout: %s\nLock at: %s\n\nConfig: %s", safe(row.Name), safe(d.State), safe(d.Server), safe(d.Email), safe(d.LockTimeout), safe(d.LockAt), safe(p.Path))
		detail += fmt.Sprintf("\nApplied insecure_skip_reprompt: %t", d.InsecureSkipReprompt)
		if row.File != nil {
			detail += fmt.Sprintf("\nFile insecure_skip_reprompt: %t", row.File.InsecureSkipReprompt)
			detail += fmt.Sprintf("\nFile timeout: %s", safe(row.File.LockTimeout))
			if row.File.LockTimeout != d.LockTimeout || row.File.InsecureSkipReprompt != d.InsecureSkipReprompt {
				detail += "\nFile differs · reload required"
			}
		}
		if p.FileError != "" {
			detail += "\nFile error: " + safe(p.FileError)
		}
		detail += "\n\nEnter use · u login/unlock\nl lock · x logout · r reload whole file"
	}
	return strings.Split(lipgloss.Wrap(detail, max(1, size.Width), ""), "\n")
}
