package tui

import (
	"cmp"
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/134130/bwb/internal/vaultapi"
)

type Node struct {
	Key, Name   string
	ContainerID string
	Item        *vaultapi.ListEntry
	Children    []*Node
}
type Row struct {
	Node  *Node
	Depth int
}
type Tree struct {
	Light                        bool
	Roots                        []*Node
	Rows                         []Row
	Index, Offset                int
	Query                        string
	collapsed, filteredCollapsed map[string]bool
	savedKey                     string
	savedOffset                  int
	Reverse                      bool
}

func buildTree(catalog map[string][]vaultapi.ListEntry) []*Node {
	personal := &Node{Key: "personal", Name: "Personal"}
	roots := []*Node{personal}
	orgs := map[string]*Node{}
	folders := map[string]*Node{}
	collections := map[string]*Node{}
	for _, entry := range catalog["organization"] {
		n := &Node{Key: "org/" + entry.ID, Name: entry.Name}
		orgs[entry.ID] = n
		roots = append(roots, n)
	}
	directory := func(parent *Node, name, id string) *Node {
		parts := strings.Split(name, "/")
		for index, part := range parts {
			key := fmt.Sprintf("%s/path/%x", parent.Key, sha256.Sum256([]byte(part)))
			var found *Node
			for _, child := range parent.Children {
				if child.Key == key {
					found = child
					break
				}
			}
			if index == len(parts)-1 && found != nil && found.ContainerID != "" && found.ContainerID != id {
				key += "/id/" + id
				found = nil
				part += " [" + id[:min(8, len(id))] + "]"
			}
			if found == nil {
				found = &Node{Key: key, Name: part}
				parent.Children = append(parent.Children, found)
			}
			if index == len(parts)-1 {
				found.ContainerID = id
			}
			parent = found
		}
		return parent
	}
	for _, entry := range catalog["folder"] {
		folders[entry.ID] = directory(personal, entry.Name, entry.ID)
	}
	for _, entry := range catalog["collection"] {
		if entry.Organization != nil {
			if org := orgs[entry.Organization.ID]; org != nil {
				collections[entry.ID] = directory(org, entry.Name, entry.ID)
			}
		}
	}
	for _, entry := range catalog["item"] {
		targets := []*Node{}
		if entry.Organization == nil {
			parent := personal
			if entry.Folder != nil && folders[entry.Folder.ID] != nil {
				parent = folders[entry.Folder.ID]
			}
			targets = append(targets, parent)
		} else {
			for _, c := range entry.Collections {
				if n := collections[c.ID]; n != nil {
					targets = append(targets, n)
				}
			}
			if len(targets) == 0 {
				if org := orgs[entry.Organization.ID]; org != nil {
					targets = append(targets, org)
				}
			}
		}
		seen := map[*Node]bool{}
		for _, parent := range targets {
			if !seen[parent] {
				item := entry
				parent.Children = append(parent.Children, &Node{Key: parent.Key + "/item/" + entry.ID, Name: entry.Name, Item: &item})
				seen[parent] = true
			}
		}
	}
	return roots
}

func fuzzy(query, value string) bool {
	q := []rune(strings.ToLower(query))
	if len(q) == 0 {
		return true
	}
	pos := 0
	for _, r := range strings.ToLower(value) {
		if r == q[pos] {
			pos++
			if pos == len(q) {
				return true
			}
		}
	}
	return false
}
func (t *Tree) current() *Node {
	if len(t.Rows) == 0 {
		return nil
	}
	return t.Rows[min(t.Index, len(t.Rows)-1)].Node
}
func (t *Tree) key() string {
	if n := t.current(); n != nil {
		return n.Key
	}
	return ""
}
func (t *Tree) rebuild(query string, size Size) {
	oldKey := t.key()
	if t.Query == "" && query != "" {
		t.savedKey, t.savedOffset = oldKey, t.Offset
		t.filteredCollapsed = map[string]bool{}
	}
	if t.Query != "" && query == "" {
		oldKey, t.Offset = t.savedKey, t.savedOffset
	}
	if query != t.Query && query != "" {
		t.filteredCollapsed = map[string]bool{}
	}
	t.Query = query
	if t.collapsed == nil {
		t.collapsed = map[string]bool{}
	}
	collapsed := t.collapsed
	if query != "" {
		collapsed = t.filteredCollapsed
	}
	var walk func([]*Node, int, bool) []Row
	walk = func(nodes []*Node, depth int, inherited bool) []Row {
		ordered := slices.Clone(nodes)
		slices.SortFunc(ordered, func(a, b *Node) int {
			if (a.Item == nil) != (b.Item == nil) {
				if a.Item == nil {
					return -1
				}
				return 1
			}
			n := cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
			if n == 0 {
				n = cmp.Compare(a.Key, b.Key)
			}
			if t.Reverse {
				return -n
			}
			return n
		})
		result := []Row{}
		for _, n := range ordered {
			matches := inherited || fuzzy(query, n.Name)
			children := walk(n.Children, depth+1, matches && n.Item == nil)
			if query != "" && !matches && len(children) == 0 {
				continue
			}
			result = append(result, Row{n, depth})
			if !collapsed[n.Key] {
				result = append(result, children...)
			}
		}
		return result
	}
	t.Rows = walk(t.Roots, 0, false)
	t.Index = 0
	for i, row := range t.Rows {
		if row.Node.Key == oldKey {
			t.Index = i
			break
		}
	}
	t.Offset = clampOffset(t.Index, t.Offset, size.Height)
}
func (t *Tree) Update(msg tea.Msg, size Size) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok && len(t.Rows) > 0 {
		switch key.String() {
		case "up", "k":
			t.Index = max(0, t.Index-1)
		case "down", "j":
			t.Index = min(len(t.Rows)-1, t.Index+1)
		case "left", "h":
			row := t.Rows[t.Index]
			c := t.collapsed
			if t.Query != "" {
				c = t.filteredCollapsed
			}
			if row.Node.Item == nil && !c[row.Node.Key] {
				c[row.Node.Key] = true
				t.rebuild(t.Query, size)
			} else {
				for i := t.Index - 1; i >= 0; i-- {
					if t.Rows[i].Depth < row.Depth {
						t.Index = i
						break
					}
				}
			}
		case "right", "l", "enter":
			row := t.Rows[t.Index]
			c := t.collapsed
			if t.Query != "" {
				c = t.filteredCollapsed
			}
			if row.Node.Item == nil {
				if c[row.Node.Key] {
					delete(c, row.Node.Key)
					t.rebuild(t.Query, size)
				} else if t.Index+1 < len(t.Rows) && t.Rows[t.Index+1].Depth > row.Depth {
					t.Index++
				}
			}
		}
	}
	t.Offset = clampOffset(t.Index, t.Offset, size.Height)
	return nil
}
func (t *Tree) View(size Size) string {
	if len(t.Rows) == 0 {
		return box("No matching items", size)
	}
	lines := []string{}
	c := t.collapsed
	if t.Query != "" {
		c = t.filteredCollapsed
	}
	for i := t.Offset; i < min(len(t.Rows), t.Offset+max(0, size.Height)); i++ {
		row := t.Rows[i]
		icon := "  "
		if row.Node.Item == nil {
			icon = "▾ "
			if c[row.Node.Key] {
				icon = "▸ "
			}
		}
		label := strings.Repeat(" ", row.Depth) + icon + highlight(t.Query, row.Node.Name, t.Light)
		if row.Node.Item != nil {
			label += "  " + subdued(safe(row.Node.Item.Type), t.Light)
		}
		if i == t.Index {
			label = selection(">"+label, size.Width, t.Light)
		} else {
			label = " " + label
		}
		lines = append(lines, label)
	}
	return box(strings.Join(lines, "\n"), size)
}

func highlight(query, value string, light bool) string {
	value = safe(value)
	if query == "" || !fuzzy(query, value) {
		return value
	}
	q := []rune(strings.ToLower(query))
	pos := 0
	var out strings.Builder
	for _, r := range value {
		if pos < len(q) && unicode.ToLower(r) == q[pos] {
			out.WriteString(lipgloss.NewStyle().Underline(true).Render(string(r)))
			pos++
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}
