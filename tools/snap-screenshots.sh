#!/usr/bin/env bash
# Generate the README screenshots in assets/.
#
# Uses the cmd/snap helper to print each screen with TrueColor ANSI,
# then pipes through `freeze` (https://github.com/charmbracelet/freeze)
# to render PNGs at a font size + nerd font that survives README zoom.
#
# install: brew install charmbracelet/tap/freeze
# usage:   ./tools/snap-screenshots.sh
# output:  ./assets/{welcome,tree,welcome-dracula,welcome-catppuccin}.png

set -euo pipefail

if ! command -v freeze >/dev/null 2>&1; then
  echo "install freeze first:  brew install charmbracelet/tap/freeze" >&2
  exit 1
fi

# Build the snap helper if missing.
if [[ ! -x ./bin/snap ]]; then
  go build -o ./bin/snap ./cmd/snap
fi

FONT="${LFG_SNAP_FONT:-FiraCode Nerd Font Mono}"
SIZE=22
LH=1.0
PAD=40

mkdir -p assets

snap_to_png() {
  local screen="$1" w="$2" h="$3" theme="$4" out="$5" bg="$6"
  ./bin/snap "$screen" "$w" "$h" "$theme" > "/tmp/lfg-snap-$$.txt"
  freeze --output "$out" "/tmp/lfg-snap-$$.txt" \
    --window --background "$bg" --padding "$PAD" \
    --font.size "$SIZE" --line-height "$LH" \
    --font.family "$FONT" >/dev/null
  rm -f "/tmp/lfg-snap-$$.txt"
  echo "wrote $out"
}

snap_to_png welcome 100 30 lfg        assets/welcome.png             "#0A0A0F"
snap_to_png tree    100 30 lfg        assets/tree.png                "#0A0A0F"
snap_to_png welcome 100 30 dracula    assets/welcome-dracula.png     "#282A36"
snap_to_png welcome 100 30 catppuccin assets/welcome-catppuccin.png  "#1E1E2E"
