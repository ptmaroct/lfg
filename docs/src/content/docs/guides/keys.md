---
title: Key bindings
description: Every keystroke the TUI listens to.
---

| Key | Action |
|---|---|
| `↑` `↓` `j` `k` | move |
| `→` `←` | expand / collapse (tree) |
| `space` `x` | toggle option |
| `a` | toggle all (tree) |
| `i` | tool info (description, homepage, install command) |
| `enter` | confirm / continue |
| `1`–`5` | jump to action (welcome) |
| `esc` | back |
| `q` | quit (confirm dialog) |
| `Ctrl+T` | cycle theme |
| `Ctrl+C` | force quit |

## Notes

- `q` always opens a confirm dialog so you don't lose progress mid-install.
- `Ctrl+C` is hard-quit; the installer process keeps running detached.
- `1`–`5` numeric jumps only fire on the welcome screen.
- The footer of every screen shows the relevant subset of keys —
  no need to memorise the whole table.

If you'd like to rebind any of these, open an issue on GitHub —
we're tracking demand before adding a config field.
