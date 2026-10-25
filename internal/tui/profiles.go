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
	Rows            []profileRow
	Index, Offset   int
	Path, FileError string
}

func (p *ProfilesPage) Update(msg tea.Msg, size Size) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "up", "k":
			p.Index = max(0, p.Index-1)
		case "down", "j":
			p.Index = min(max(0, len(p.Rows)-1), p.Index+1)
		}
	}
	p.Offset = clampOffset(p.Index, p.Offset, size.Height)
	return nil
}
func (p *ProfilesPage) current() *profileRow {
	if len(p.Rows) == 0 {
		return nil
	}
	return &p.Rows[min(p.Index, len(p.Rows)-1)]
}
func (p *ProfilesPage) View(size Size) string {
	left := min(28, max(0, size.Width/3))
	lines := []string{}
	for i := p.Offset; i < min(len(p.Rows), p.Offset+max(0, size.Height)); i++ {
		row := p.Rows[i]
		line := safe(row.Name) + "  " + row.Status.State
		if i == p.Index {
			line = selected.Render("> " + line)
		}
		lines = append(lines, line)
	}
	detail := "No profiles"
	if row := p.current(); row != nil {
		d := row.Status
		detail = fmt.Sprintf("%s\nState: %s\nServer: %s\nEmail: %s\nApplied timeout: %s\nLock at: %s\n\nConfig: %s", safe(row.Name), safe(d.State), safe(d.Server), safe(d.Email), safe(d.LockTimeout), safe(d.LockAt), safe(p.Path))
		if row.File != nil {
			detail += fmt.Sprintf("\nFile timeout: %s", safe(row.File.LockTimeout))
			if row.File.Server != d.Server || row.File.Email != d.Email || row.File.LockTimeout != d.LockTimeout {
				detail += "\nFile differs · reload required"
			}
		}
		if p.FileError != "" {
			detail += "\nFile error: " + safe(p.FileError)
		}
		detail += "\n\nEnter use · u login/unlock\nl lock · x logout · r reload whole file"
	}
	return box(lipgloss.JoinHorizontal(lipgloss.Top, box(strings.Join(lines, "\n"), Size{left, size.Height}), box(detail, Size{max(0, size.Width-left-1), size.Height})), size)
}
