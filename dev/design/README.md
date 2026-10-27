# Explorer UI design study

Open `tui.html` in a browser. This is a local, dependency-free visual prototype with
synthetic fixtures; it does not connect to a daemon or write to the clipboard.
The Go TUI implements this direction; this file remains a standalone design reference.

- Every panel has the same border width. The focused border uses a neutral foreground;
  inactive borders are dimmed. Neither uses an accent-colored or heavier border.
- Dark/light palettes follow the system preference initially and can be switched manually.
  Only the border/title dims; unfocused content remains readable.
- Selection and focus are independent: the selected item and field retain their background
  when another panel owns keyboard input.
- The Explorer keeps collection ancestry while fuzzy filtering. Matching a directory
  includes its descendants. An item can appear in more than one collection.
- `v` reveals/hides the entire selected item. Field or pane navigation keeps reveal;
  changing item or entering search hides it. `y` demonstrates copying without reveal.
- Narrow mode shows the focused content pane; search returns to Explorer. `Tab` switches
  content panes. Status and key hints stay on separate lines.

Preview variants:

- `tui.html?theme=dark`
- `tui.html?theme=light&focus=detail&revealed`
- `tui.html?theme=light&focus=detail&narrow`

`?test=1` runs browser assertions for focus, border width, item reveal lifetime,
filtering, empty results, themes, and narrow panes. Success sets `body[data-test=passed]`.

This study covers Explorer only. Profiles, authentication dialogs, expanded field views,
real TOTP refresh, terminal theme detection and clipboard behavior are implemented in Go
and are not simulated here. The preview TOTP is deliberately static and labeled as such.
