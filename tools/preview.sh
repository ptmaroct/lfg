#!/usr/bin/env bash
# Page through every snapshot golden as colored ANSI.
# Usage:
#   ./tools/preview.sh                # all goldens
#   ./tools/preview.sh welcome        # filter by screen
#   ./tools/preview.sh welcome lfg    # filter by screen + theme
#   ./tools/preview.sh -- xl          # filter by size token
#
# Keys (less): j/k scroll · / search · q quit · g/G top/bottom

set -euo pipefail

DIR="internal/tui/testdata"
if [[ ! -d "$DIR" ]]; then
  echo "no goldens at $DIR — run: make snap-update" >&2
  exit 1
fi

# Build filter from positional args (any-substring match on filename).
FILTERS=("$@")

shopt -s nullglob
FILES=("$DIR"/*.golden)

selected=()
for f in "${FILES[@]}"; do
  base="$(basename "$f")"
  match=true
  for pat in "${FILTERS[@]}"; do
    [[ "$pat" == "--" ]] && continue
    if [[ "$base" != *"$pat"* ]]; then
      match=false
      break
    fi
  done
  $match && selected+=("$f")
done

if [[ ${#selected[@]} -eq 0 ]]; then
  echo "no goldens matched filter: ${FILTERS[*]:-<none>}" >&2
  exit 1
fi

# Stitch each golden with a header banner. Pipe through less -R so ANSI
# colors render. Without -R less escapes them.
{
  for f in "${selected[@]}"; do
    base="$(basename "$f" .golden)"
    # Pretty header in bold magenta, full-width rule.
    width="$(tput cols 2>/dev/null || echo 80)"
    rule="$(printf '─%.0s' $(seq 1 "$width"))"
    printf '\n\033[1;35m%s\033[0m\n' "$base"
    printf '\033[2;90m%s\033[0m\n\n' "$rule"
    cat "$f"
    printf '\n\n'
  done
} | less -R
