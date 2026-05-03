---
title: Backup & restore
description: Snapshot your dotfiles and configs into a portable archive.
---

```sh
lfg backup                      # locked tar.age (default)
lfg backup --encrypt=false      # plain tar.gz
lfg backup --include-ssh-keys   # only with encrypt; opt-in
lfg backup --restore <path>     # extract a previous archive
```

## What's in the archive

Source groups (each one is silently skipped if nothing matches):

| Group | Paths |
|---|---|
| Shell config | `.zshrc`, `.zprofile`, `.zshenv`, `.bashrc`, `.bash_profile`, `.profile` |
| Editors / dotfiles | `.gitconfig`, `.tmux.conf`, `.vimrc`, `.editorconfig`, `.inputrc` |
| Starship + dev tools | `~/.config/{starship,mise,bat,btop,lazygit,yazi,glow}` |
| Editor configs | `~/.config/{nvim,zed,ghostty,zellij}` |
| AI tool settings | `~/.claude/{settings.json,CLAUDE.md,agents,commands}`, `~/.codex/config.toml` |
| SSH | `~/.ssh/` (config + public keys; **private keys never copied** unless `--include-ssh-keys` *and* encryption) |

The TUI shows you this exact list with `●`/`○` presence dots before
asking you to confirm — no surprise inclusions.

## Why tar (not zip)

`tar` preserves UNIX file modes and symlinks, which matter for SSH
config and dotfile hierarchies. Zip mangles both. Output is `tar.gz`
(plain) or `tar.age` (locked with a key from
[age](https://github.com/FiloSottile/age)).

## Lock it, or skip the lock

**Locked (`tar.age`)** — encrypted with a key at
`~/.config/lfg/key.txt`. Only that key can open it. Pick this if the
backup will leave your machine (cloud sync, USB, email).
**Back up the key file separately** — without it, even you can't
recover the archive.

**Plain (`tar.gz`)** — anyone with the file can read it. Pick this
when the archive stays on this machine and you want to peek inside
(`tar -tzf <file>`).

## Restore

The result screen shows the exact `lfg backup --restore <path>`
command for whichever option you picked. Restore writes files
back to their original home-relative paths and refuses to clobber
existing files unless you pass `--force`.

For an `.age` archive, `lfg backup --restore` will look for the key
at `~/.config/lfg/key.txt` automatically. Override with `--key <path>` if you stashed it elsewhere.
