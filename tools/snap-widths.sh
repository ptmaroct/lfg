#!/usr/bin/env bash
# Render every screen × width as a static PNG via `freeze`.
# Useful for visual regression review without launching the TUI manually.
#
# install: brew install charmbracelet/tap/freeze
# usage:   ./tools/snap-widths.sh [theme]   (default: lfg)
# output:  ./snaps/<screen>_<width>.png

set -euo pipefail

THEME="${1:-lfg}"
OUT="snaps"
mkdir -p "$OUT"

# Sizes mirror internal/tui/screens_test.go. Edit there first if you
# change the matrix; this is a thin wrapper.
declare -a SIZES=(
  "60 20 xs"
  "80 24 sm"
  "100 30 md"
  "120 36 lg"
  "160 44 xl"
  "200 50 xxl"
)

# We render each screen by emitting its View() to stdout via a tiny helper.
# `freeze` then turns ANSI text → PNG.
#
# The helper isn't built yet; for now we render through the running binary
# via `script`/expect-like flow. Simplest path: print captured goldens.

if ! command -v freeze >/dev/null 2>&1; then
  echo "install freeze first:  brew install charmbracelet/tap/freeze" >&2
  exit 1
fi

for sz in "${SIZES[@]}"; do
  read -r W H NAME <<<"$sz"
  for screen in welcome bundles backup; do
    GOLDEN="internal/tui/testdata/${screen}_${THEME}_${NAME}_${W}x${H}.golden"
    if [[ ! -f "$GOLDEN" ]]; then
      # fallback for screens that don't iterate themes
      GOLDEN="internal/tui/testdata/${screen}_lfg_${NAME}_${W}x${H}.golden"
    fi
    if [[ ! -f "$GOLDEN" ]]; then
      echo "skip: $GOLDEN missing"
      continue
    fi
    OUTPUT="$OUT/${screen}_${THEME}_${NAME}.png"
    freeze --output "$OUTPUT" --window --width "$W" --height "$H" "$GOLDEN" >/dev/null
    echo "wrote $OUTPUT"
  done
done

echo "done. open $OUT/ to review."
