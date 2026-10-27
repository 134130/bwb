package tui

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/vaultapi"
	"github.com/charmbracelet/x/ansi"
)

type copyFieldMsg struct{ ID, Field string }
type Detail struct {
	Node                      *Node
	Fields                    []bitwarden.ItemField
	Index, Offset, Scroll     int
	Revealed, Expanded, Light bool
	Error                     string
	OTP                       *OTPPage
	otpPending                bool
}

func (d *Detail) clearValue() {
	for i := range d.Fields {
		if d.Fields[i].Sensitive {
			d.Fields[i].Value = nil
			d.Fields[i].State = "hidden"
		}
	}
	d.Revealed, d.Expanded, d.Scroll = false, false, 0
	d.OTP = nil
	d.otpPending = false
}
func fieldGroup(key string) string {
	switch {
	case key == "notes":
		return "NOTES"
	case strings.HasPrefix(key, "custom/"):
		return "CUSTOM FIELDS"
	case strings.HasPrefix(key, "url/"):
		return "WEBSITES"
	default:
		return "FIELDS"
	}
}
func (d *Detail) setItem(item vaultapi.ItemResponse, reveal bool) {
	oldKey := ""
	if len(d.Fields) > 0 {
		oldKey = d.Fields[d.Index].Key
	}
	d.Fields = nil
	d.Error = item.Restriction
	d.Revealed = reveal && item.Restriction == ""
	for _, f := range item.Fields {
		if f.Key == "name" {
			continue
		}
		if f.Key == "totp-secret" {
			d.Fields = append(d.Fields, bitwarden.ItemField{Key: "totp", Label: "Authentication code", Sensitive: true, State: "hidden"})
		}
		d.Fields = append(d.Fields, f)
	}
	order := map[string]int{"FIELDS": 0, "WEBSITES": 1, "CUSTOM FIELDS": 2, "NOTES": 3}
	slices.SortStableFunc(d.Fields, func(a, b bitwarden.ItemField) int {
		return cmp.Compare(order[fieldGroup(a.Key)], order[fieldGroup(b.Key)])
	})
	d.Index = min(d.Index, max(0, len(d.Fields)-1))
	for i, f := range d.Fields {
		if f.Key == oldKey {
			d.Index = i
		}
	}
}
func (d *Detail) fieldValue(f bitwarden.ItemField) string {
	if f.Key == "totp" && d.Revealed {
		if d.OTP == nil {
			return "Loading…"
		}
		if d.OTP.Stopped {
			return "Unavailable · hide/reveal to retry"
		}
		if d.OTP.Code.Code == "" || !d.OTP.Now.Before(d.OTP.Code.ExpiresAt) {
			return "Refreshing…"
		}
		seconds := int((d.OTP.Code.ExpiresAt.Sub(d.OTP.Now) + time.Second - 1) / time.Second)
		return fmt.Sprintf("%s · %ds", d.OTP.Code.Code, seconds)
	}
	if f.Value != nil {
		if *f.Value == "" {
			return "[empty]"
		}
		return *f.Value
	}
	return "••••••••••••"
}
func (d *Detail) lines(size Size) ([]string, []int) {
	if d.Node == nil {
		return []string{"Select an item"}, nil
	}
	lines := []string{lipgloss.NewStyle().Bold(true).Render(safe(d.Node.Name))}
	if d.Node.Item == nil {
		return append(lines, "", "Folder / collection", "←→ navigate · Enter open"), nil
	}
	item := d.Node.Item
	meta := item.Type
	if item.Organization != nil {
		meta += " · " + item.Organization.Name
	}
	lines = append(lines, subdued(safe(meta), d.Light))
	names := []string{}
	for _, c := range item.Collections {
		names = append(names, safe(c.Name))
	}
	if len(names) > 0 {
		lines = append(lines, subdued(strings.Join(names, ", "), d.Light))
	}
	visibility := "Secrets hidden · v reveal item"
	if d.Revealed {
		visibility = "Secrets revealed · v hide item"
	}
	lines = append(lines, "", subdued(visibility, d.Light))
	if d.Error != "" {
		return append(lines, "", "Fields unavailable: "+safe(d.Error)), nil
	}
	if len(d.Fields) == 0 {
		return append(lines, "", "Loading fields…"), nil
	}
	group := ""
	positions := make([]int, len(d.Fields))
	labelWidth := min(20, max(8, size.Width/3))
	for i, f := range d.Fields {
		section := fieldGroup(f.Key)
		if section != group {
			lines = append(lines, "", subdued(section, d.Light))
			group = section
		}
		value := d.fieldValue(f)
		if before, _, found := strings.Cut(value, "\n"); found {
			value = before + " ↵"
		}
		label := ansi.Truncate(safe(f.Label), labelWidth, "…")
		label += strings.Repeat(" ", max(0, labelWidth-ansi.StringWidth(label)))
		row := "  " + label + "  " + safe(value)
		positions[i] = len(lines)
		if i == d.Index {
			row = selection("> "+label+"  "+safe(value), size.Width, d.Light)
		}
		lines = append(lines, row)
	}
	return lines, positions
}
func (d *Detail) Update(msg tea.Msg, size Size) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		if d.Expanded {
			switch key.String() {
			case "esc", "q", "enter":
				d.Expanded = false
			case "down", "j":
				d.Scroll++
			case "up", "k":
				d.Scroll = max(0, d.Scroll-1)
			case "pgdown":
				d.Scroll += max(1, size.Height/2)
			case "pgup":
				d.Scroll = max(0, d.Scroll-max(1, size.Height/2))
			}
		} else {
			switch key.String() {
			case "up", "k":
				d.Index = max(0, d.Index-1)
			case "down", "j":
				d.Index = min(max(0, len(d.Fields)-1), d.Index+1)
			case "pgdown":
				d.Index = min(max(0, len(d.Fields)-1), d.Index+max(1, size.Height/2))
			case "pgup":
				d.Index = max(0, d.Index-max(1, size.Height/2))
			case "enter":
				if len(d.Fields) > 0 {
					d.Expanded = true
					d.Scroll = 0
				}
			}
		}
		if key.String() == "y" && d.Node != nil && d.Node.Item != nil && len(d.Fields) > 0 {
			request := copyFieldMsg{d.Node.Item.ID, d.Fields[d.Index].Key}
			return func() tea.Msg { return request }
		}
	}
	if d.Expanded {
		d.Scroll = min(d.Scroll, max(0, len(d.expandedLines(size))-max(1, size.Height-2)))
		return nil
	}
	lines, positions := d.lines(size)
	if len(positions) > 0 {
		d.Offset = clampOffset(positions[d.Index], d.Offset, size.Height)
	}
	d.Offset = min(d.Offset, max(0, len(lines)-size.Height))
	return nil
}
func (d *Detail) expandedLines(size Size) []string {
	if len(d.Fields) == 0 {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(d.fieldValue(d.Fields[d.Index]), "\n") {
		lines = append(lines, strings.Split(lipgloss.Wrap(safe(line), max(1, size.Width), ""), "\n")...)
	}
	return lines
}
func (d *Detail) View(size Size) string {
	if d.Expanded && len(d.Fields) > 0 {
		lines := d.expandedLines(size)
		start := min(d.Scroll, len(lines))
		return box(safe(d.Fields[d.Index].Label)+" · Esc back\n\n"+strings.Join(lines[start:], "\n"), size)
	}
	lines, _ := d.lines(size)
	return box(strings.Join(lines[min(d.Offset, len(lines)):], "\n"), size)
}
