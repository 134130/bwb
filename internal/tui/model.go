package tui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/config"
	"github.com/134130/bwb/internal/vaultapi"
	"github.com/134130/bwb/internal/vaultclient"
	kit "github.com/134130/go-daemonkit"
	"github.com/atotto/clipboard"
)

type tickMsg time.Time
type resultMsg struct {
	Epoch uint64
	Kind  string
	Value any
	Err   error
}
type statusMsg struct {
	Epoch  uint64
	Status vaultapi.StatusResponse
	Err    error
}
type promptMsg struct {
	Context context.Context
	Request vaultclient.AuthPrompt
	Reply   chan []byte
}
type promptEnded struct{ Reply chan []byte }
type authDialog struct {
	Prompt promptMsg
	Input  Input
}

// Root alone stores terminal size. Every child receives content constraints.
type Root struct {
	otp           *OTPPage
	standaloneID  string
	workKind      string
	ctx           context.Context
	client        *kit.Client
	executor      vaultclient.Executor
	size          Size
	profile, page string
	state         vaultapi.StatusResponse
	explorer      ExplorerPage
	profiles      ProfilesPage
	auth          *authDialog
	epoch         uint64
	cancel        context.CancelFunc
	busy          bool
	message       string
	confirmLogout string
	help          bool
	lastActivity  time.Time
}

func Run(ctx context.Context, profile string, client *kit.Client, input io.Reader, output io.Writer) error {
	return run(ctx, profile, "", client, input, output)
}
func RunTOTP(ctx context.Context, profile, id string, client *kit.Client, input io.Reader, output io.Writer) error {
	return run(ctx, profile, id, client, input, output)
}
func run(ctx context.Context, profile, id string, client *kit.Client, input io.Reader, output io.Writer) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	root := &Root{ctx: ctx, client: client, profile: profile, page: "explorer", explorer: newExplorer()}
	var program *tea.Program
	root.standaloneID = id
	root.executor = vaultclient.Executor{Client: client, Prompt: func(ctx context.Context, request vaultclient.AuthPrompt) ([]byte, error) {
		reply := make(chan []byte)
		program.Send(promptMsg{ctx, request, reply})
		select {
		case password := <-reply:
			if password == nil {
				return nil, context.Canceled
			}
			return password, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}}
	program = tea.NewProgram(root, tea.WithContext(ctx), tea.WithInput(input), tea.WithOutput(output))
	_, err := program.Run()
	root.purge()
	return err
}
func tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}
func (r *Root) Init() tea.Cmd     { return tea.Batch(r.load(), tick()) }
func (r *Root) contentSize() Size { return Size{max(0, r.size.Width), max(0, r.size.Height-3)} }
func (r *Root) operation(kind string, call func(context.Context) (any, error)) tea.Cmd {
	if r.cancel != nil {
		r.cancel()
	}
	r.epoch++
	epoch := r.epoch
	ctx, cancel := context.WithCancel(r.ctx)
	r.cancel = cancel
	r.busy = true
	r.workKind = kind
	r.message = ""
	return func() tea.Msg { defer cancel(); value, err := call(ctx); return resultMsg{epoch, kind, value, err} }
}
func (r *Root) load() tea.Cmd {
	if r.standaloneID != "" {
		return r.openOTP(r.standaloneID)
	}
	return r.loadCatalog(false)
}
func (r *Root) loadCatalog(sync bool) tea.Cmd {
	profile := r.profile
	kind := "catalog"
	if sync {
		kind = "sync"
		r.explorer.Detail.clearValue()
	}
	return r.operation(kind, func(ctx context.Context) (any, error) {
		if sync {
			if _, err := r.executor.Sync(ctx, vaultapi.ProfileRequest{Profile: profile}); err != nil {
				return nil, err
			}
		}
		catalog := map[string][]vaultapi.ListEntry{}
		for _, kind := range []string{"organization", "collection", "folder", "item"} {
			entries, err := r.executor.List(ctx, vaultapi.ListRequest{Profile: profile, Kind: kind, Background: true})
			if err != nil {
				return nil, err
			}
			catalog[kind] = entries
		}
		return catalog, nil
	})
}
func (r *Root) loadProfiles() tea.Cmd {
	current := r.profile
	return r.operation("profiles", func(ctx context.Context) (any, error) {
		result, err := r.client.Invoke(ctx, vaultapi.Profiles, vaultapi.Empty{})
		if err != nil {
			return nil, err
		}
		daemon, err := r.client.Invoke(ctx, vaultapi.DaemonStatus, vaultapi.Empty{})
		if err != nil {
			return nil, err
		}
		page := ProfilesPage{Path: daemon.ConfigPath}
		file, err := config.Load(daemon.ConfigPath, false)
		if err != nil {
			page.FileError = err.Error()
		}
		names := append(result.Names, current)
		if file != nil {
			for name := range file.Profiles {
				names = append(names, name)
			}
		}
		slices.Sort(names)
		names = slices.Compact(names)
		for _, name := range names {
			state, err := r.client.Invoke(ctx, vaultapi.Status, vaultapi.ProfileRequest{Profile: name})
			if err != nil {
				return nil, err
			}
			row := profileRow{Name: name, Status: state}
			if file != nil {
				if policy, ok := file.Profiles[name]; ok {
					row.File = &policy
				}
			}
			if name == current {
				page.Index = len(page.Rows)
			}
			page.Rows = append(page.Rows, row)
		}
		return page, nil
	})
}
func (r *Root) poll() tea.Cmd {
	epoch, profile := r.epoch, r.profile
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(r.ctx, 3*time.Second)
		defer cancel()
		status, err := r.client.Invoke(ctx, vaultapi.Status, vaultapi.ProfileRequest{Profile: profile})
		return statusMsg{epoch, status, err}
	}
}
func (r *Root) purge() {
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
	r.epoch++
	r.busy = false
	r.otp = nil
	if r.auth != nil {
		r.auth.Input.SetValue("")
		select {
		case r.auth.Prompt.Reply <- nil:
		default:
		}
		r.auth = nil
	}
	// Preserve opaque navigation IDs, never labels, queries or secret values.
	key, offset, collapsed := r.explorer.Tree.key(), r.explorer.Tree.Offset, r.explorer.Tree.collapsed
	r.explorer = newExplorer()
	r.explorer.Tree.savedKey = key
	r.explorer.Tree.savedOffset = offset
	r.explorer.Tree.collapsed = collapsed
	r.profiles = ProfilesPage{}
}
func (r *Root) describe() tea.Cmd {
	n := r.explorer.Tree.current()
	r.explorer.Detail = Detail{Node: n}
	if n == nil || n.Item == nil {
		return nil
	}
	profile, id := r.profile, n.Item.ID
	return r.operation("describe", func(ctx context.Context) (any, error) {
		return r.executor.Describe(ctx, vaultapi.DescribeRequest{Profile: profile, ID: id})
	})
}
func (r *Root) lock(profile string) tea.Cmd {
	r.purge()
	r.state.State = "locked"
	return r.operation("locked", func(ctx context.Context) (any, error) {
		return r.client.Invoke(ctx, vaultapi.Lock, vaultapi.ProfileRequest{Profile: profile})
	})
}
func (r *Root) activity() tea.Cmd {
	if r.state.State != "unlocked" || time.Since(r.lastActivity) < time.Second {
		return nil
	}
	r.lastActivity = time.Now()
	epoch, profile := r.epoch, r.profile
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(r.ctx, 3*time.Second)
		defer cancel()
		_, err := r.client.Invoke(ctx, vaultapi.Activity, vaultapi.ProfileRequest{Profile: profile})
		return resultMsg{Epoch: epoch, Kind: "activity", Err: err}
	}
}
func (r *Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	size := r.contentSize()
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		r.size = Size{max(0, m.Width), max(0, m.Height)}
		size = r.contentSize()
		r.explorer.Update(LayoutMsg{}, size)
		r.profiles.Update(LayoutMsg{}, size)
		if r.auth != nil {
			r.auth.Input.Update(LayoutMsg{}, Size{size.Width, 1})
		}
		return r, nil
	case tickMsg:
		if r.state.State == "disconnected" {
			return r, tick()
		}
		if r.otp != nil {
			r.otp.Update(m, size)
			if !r.busy && !r.otp.Stopped && !r.otp.Now.Before(r.otp.Code.ExpiresAt) {
				cmd := r.refreshOTP(false)
				return r, tea.Batch(tick(), cmd, r.poll())
			}
		}
		return r, tea.Batch(tick(), r.poll())
	case statusMsg:
		if m.Epoch != r.epoch {
			return r, nil
		}
		if m.Err != nil {
			r.purge()
			r.state.State = "disconnected"
			r.message = "Daemon disconnected; quit and reopen TUI"
			return r, nil
		}
		old := r.state.State
		r.state = m.Status
		if old == "unlocked" && m.Status.State != "unlocked" {
			r.purge()
			r.message = "Vault locked · Enter to unlock"
			if r.page == "profiles" {
				return r, r.loadProfiles()
			}
		}
		return r, nil
	case promptMsg:
		if m.Context.Err() != nil {
			return r, nil
		}
		if r.auth != nil {
			select {
			case m.Reply <- nil:
			default:
			}
			return r, nil
		}
		input := newInput(true)
		input.Update(LayoutMsg{}, Size{size.Width, 1})
		r.auth = &authDialog{Prompt: m, Input: input}
		return r, tea.Batch(r.auth.Input.Focus(), func() tea.Msg { <-m.Context.Done(); return promptEnded{m.Reply} })
	case promptEnded:
		if r.auth != nil && r.auth.Prompt.Reply == m.Reply {
			r.auth.Input.SetValue("")
			r.auth = nil
		}
		return r, nil
	case resultMsg:
		if m.Epoch != r.epoch {
			return r, nil
		}
		if m.Kind == "activity" {
			if m.Err != nil {
				r.purge()
				r.state.State = "locked"
			}
			return r, nil
		}
		r.busy = false
		if m.Err != nil {
			if !errors.Is(m.Err, context.Canceled) {
				r.message = safe(m.Err.Error())
			}
			if m.Kind == "otp" && r.otp != nil {
				r.otp.Code = vaultapi.TOTPResponse{}
				r.otp.Message = r.message
				r.otp.Stopped = true
			}
			if m.Kind == "describe" {
				r.explorer.Detail.Error = r.message
			}
			return r, nil
		}
		switch m.Kind {
		case "otp":
			r.epoch++
			if r.otp != nil {
				r.otp.Code = m.Value.(vaultapi.TOTPResponse)
				r.otp.Now = time.Now()
				r.otp.Stopped = false
				r.state.State = "unlocked"
			}
		case "catalog", "sync":
			r.epoch++ // Discard status polls begun before authentication completed.
			r.state.State = "unlocked"
			t := &r.explorer.Tree
			restore := len(t.Rows) == 0 && t.Query == ""
			t.Roots = buildTree(m.Value.(map[string][]vaultapi.ListEntry))
			left, _ := r.explorer.layout(size)
			t.rebuild(t.Query, left)
			if restore && t.savedKey != "" {
				for i, row := range t.Rows {
					if row.Node.Key == t.savedKey {
						t.Index = i
						t.Offset = t.savedOffset
						break
					}
				}
				t.Update(LayoutMsg{}, left)
			}
			return r, r.describe()
		case "describe":
			r.explorer.Detail.Fields = nil
			for _, f := range m.Value.(vaultapi.DescribeResponse).Fields {
				if f.Path == "login/totp" {
					r.explorer.Detail.Fields = append(r.explorer.Detail.Fields, bitwarden.FieldInfo{Path: "totp", Label: "Authentication code"})
					f.Label = "TOTP setup key"
				}
				r.explorer.Detail.Fields = append(r.explorer.Detail.Fields, f)
			}
			r.explorer.Update(LayoutMsg{}, size)
		case "read":
			r.explorer.Detail.Value = m.Value.(vaultapi.ReadResponse).Value
			r.explorer.Detail.Revealed = true
			r.explorer.Detail.Scroll = 0
		case "profiles":
			r.profiles = m.Value.(ProfilesPage)
			r.profiles.Update(LayoutMsg{}, size)
		case "reload", "logout":
			return r, r.loadProfiles()
		case "locked":
			if r.page == "profiles" {
				return r, r.loadProfiles()
			}
		}
		return r, nil
	case revealMsg:
		if r.otp != nil || r.busy && r.workKind == "sync" {
			return r, nil
		}
		n := r.explorer.Tree.current()
		if n == nil || n.Item == nil || n.Item.ID != m.ID || r.explorer.Focus != "detail" || len(r.explorer.Detail.Fields) == 0 || r.explorer.Detail.Fields[r.explorer.Detail.Index].Path != m.Field {
			return r, nil
		}
		if m.Field == "totp" {
			return r, r.openOTP(m.ID)
		}
		profile := r.profile
		return r, r.operation("read", func(ctx context.Context) (any, error) {
			return r.executor.Read(ctx, vaultapi.ReadRequest{Profile: profile, ID: m.ID, Field: m.Field})
		})
	case tea.KeyPressMsg:
		key := m.String()
		if key == "ctrl+c" {
			r.purge()
			return r, tea.Quit
		}
		if key == "ctrl+l" {
			return r, r.lock(r.profile)
		}
		if r.auth != nil {
			if key == "esc" {
				r.purge()
				r.message = "Authentication canceled"
				return r, nil
			}
			if key == "enter" {
				password := []byte(r.auth.Input.Value())
				r.auth.Input.SetValue("")
				if len(password) == 0 || len(password) > 4096 {
					clear(password)
					r.message = "Password must contain 1–4096 bytes"
					return r, nil
				}
				select {
				case r.auth.Prompt.Reply <- password:
				case <-r.auth.Prompt.Context.Done():
					clear(password)
				}
				r.auth = nil
				return r, nil
			}
			return r, r.auth.Input.Update(msg, Size{size.Width, 1})
		}
		if r.standaloneID != "" && r.otp == nil {
			if key == "q" || key == "esc" {
				r.purge()
				return r, tea.Quit
			}
			if key == "enter" && !r.busy {
				return r, r.load()
			}
			return r, nil
		}
		if r.otp != nil {
			switch key {
			case "q", "esc":
				r.closeOTP()
				if r.standaloneID != "" {
					return r, tea.Quit
				}
			case "y":
				r.copyOTP(clipboard.WriteAll)
				return r, r.activity()
			case "enter":
				if r.otp.Stopped {
					return r, r.openOTP(r.otp.ID)
				}
			}
			return r, nil
		}
		if r.confirmLogout != "" {
			name := r.confirmLogout
			r.confirmLogout = ""
			if key == "y" {
				r.purge()
				return r, r.operation("logout", func(ctx context.Context) (any, error) {
					return r.client.Invoke(ctx, vaultapi.Logout, vaultapi.ProfileRequest{Profile: name})
				})
			}
			return r, nil
		}
		if r.explorer.Focus != "search" || r.page != "explorer" {
			if key == "q" {
				r.purge()
				return r, tea.Quit
			}
			if key == "?" {
				if r.busy && r.workKind == "read" {
					if r.cancel != nil {
						r.cancel()
					}
					r.epoch++
					r.busy = false
				}
				r.explorer.Detail.clearValue()
				r.help = !r.help
				return r, nil
			}
		}
		if r.help {
			return r, nil
		}
		if r.busy && r.workKind == "sync" {
			return r, nil
		}
		if key == "ctrl+p" {
			r.explorer.Detail.clearValue()
			r.confirmLogout = ""
			if r.page == "profiles" {
				r.page = "explorer"
				if r.state.State == "unlocked" {
					return r, r.load()
				}
				return r, nil
			}
			r.page = "profiles"
			return r, r.loadProfiles()
		}
		if r.page == "profiles" {
			row := r.profiles.current()
			if row != nil {
				switch key {
				case "enter", "u":
					name := row.Name
					r.purge()
					r.profile = name
					r.page = "explorer"
					r.state = vaultapi.StatusResponse{}
					return r, r.load()
				case "l":
					return r, r.lock(row.Name)
				case "x":
					r.confirmLogout = row.Name
					return r, nil
				case "r":
					r.explorer.Detail.clearValue()
					return r, r.operation("reload", func(ctx context.Context) (any, error) { return r.client.Invoke(ctx, vaultapi.Reload, vaultapi.Empty{}) })
				}
			}
			return r, r.profiles.Update(msg, size)
		}
		if r.state.State != "unlocked" {
			if key == "enter" && !r.busy {
				return r, r.load()
			}
			return r, nil
		}
		if key == "r" && r.explorer.Focus != "search" {
			return r, r.loadCatalog(true)
		}
		before := r.explorer.Tree.key()
		oldFocus := r.explorer.Focus
		field := r.explorer.Detail.Index
		cmd := r.explorer.Update(msg, size)
		r.explorer.Update(LayoutMsg{}, size)
		if before != r.explorer.Tree.key() {
			cmd = r.describe()
		} else if (field != r.explorer.Detail.Index || oldFocus == "detail" && r.explorer.Focus != "detail") && r.busy && r.workKind == "read" {
			if r.cancel != nil {
				r.cancel()
			}
			r.epoch++
			r.busy = false
		}
		if oldFocus == "detail" && r.explorer.Focus != "detail" {
			r.explorer.Detail.clearValue()
		}
		return r, tea.Batch(cmd, r.activity())
	}
	// Cursor blink and paste messages belong only to the focused input.
	if r.auth != nil {
		return r, r.auth.Input.Update(msg, Size{size.Width, 1})
	}
	if r.page == "explorer" && r.explorer.Focus == "search" && r.state.State == "unlocked" {
		before := r.explorer.Tree.key()
		cmd := r.explorer.Update(msg, size)
		left, _ := r.explorer.layout(size)
		r.explorer.Tree.rebuild(r.explorer.Search.Value(), left)
		if before != r.explorer.Tree.key() {
			cmd = tea.Batch(cmd, r.describe())
		}
		return r, cmd
	}
	return r, nil
}
func (r *Root) View() tea.View {
	title := fmt.Sprintf("bwb · %s    Explorer / Profiles [Ctrl+P]    %s", safe(r.profile), safe(r.state.State))
	if r.standaloneID != "" {
		title = fmt.Sprintf("bwb · %s · TOTP · %s", safe(r.profile), safe(r.state.State))
	}
	footer := "/ search  ←→ tree  Tab pane  Enter open  s sort  r sync  Ctrl+L lock  ? help  q quit"
	body := "Vault locked · Enter to login/unlock"
	size := r.contentSize()
	if r.page == "profiles" {
		body = r.profiles.View(size)
	} else if r.state.State == "unlocked" {
		body = r.explorer.View(size)
	}
	if r.busy {
		footer = "Working…  Esc in password dialog cancels · Ctrl+L lock · Ctrl+C quit"
	}
	if r.message != "" {
		footer = r.message
	}
	if r.help {
		body = "Explorer: / fuzzy search; directory matches include descendants.\nEnter/Esc leave search editing; Esc again restores the original tree.\nArrows or hjkl navigate; Tab changes pane; s reverses name order.\nExplorer: r syncs and refreshes, preserving search and tree position.\nEnter on a field reveals it; Enter hides it; PgUp/PgDown scroll value.\nCtrl+P opens Profiles; u authenticates; l locks; x logs out; r reloads.\nUser navigation renews idle time; polling does not.\nq quits without locking the daemon; Ctrl+L locks.\n? closes help."
	}
	if r.otp != nil {
		body = r.otp.View(size)
		footer = "y copy · q close · Ctrl+L lock"
	}
	if r.confirmLogout != "" {
		body = "Remove saved authentication for " + safe(r.confirmLogout) + "? [y/N]"
	}
	if r.auth != nil {
		p := r.auth.Prompt.Request
		body = fmt.Sprintf("%s %s\n%s at %s\n\n%s\n\nEnter submit · Esc cancel", safe(p.Action), safe(p.Profile), safe(p.Email), safe(p.Server), r.auth.Input.View(Size{size.Width, 1}))
	}
	view := tea.NewView(box(title+"\n"+strings.Repeat("─", max(0, r.size.Width))+"\n"+box(body, size)+"\n"+safe(footer), r.size))
	view.AltScreen = true
	return view
}
