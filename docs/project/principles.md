---
title: Design principles
description: The five rules that drive every decision in lfg.
---

1. **Wrap, don't reinvent.** chezmoi, mise, brew, age already work.
   `lfg` is the opinionated interactive layer on top.

2. **Use Charm components.** This project lives in the Charm
   ecosystem; if `huh`/`bubbles` ships a primitive, use it instead
   of hand-rolling. Keybindings, theming, focus management, and
   accessibility stay consistent for free.

3. **Never silently touch user data.** Private SSH keys stay
   on-device by default. Secrets are encrypted client-side; the
   master key never leaves the user's control.

4. **Beautiful beats clever.** First five minutes are the product.
   If it doesn't look great, nobody runs step two.

5. **Ship-able MVP.** UX-first prototype before installers,
   installers before sync, sync before the kitchen sink.

These aren't aspirational — they're how we decide between two
plausible implementations on every PR. When in doubt, the principle
that says *"no"* wins.
