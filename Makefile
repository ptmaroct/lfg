# lfg — build, test, snapshot, demo
.PHONY: build run test snap snap-update widths demo docker docker-run clean

build:
	go build -o lfg ./cmd/lfg

run: build
	./lfg

test:
	go test ./...

# Snapshot tests — diff current View() output against testdata/*.golden
snap:
	go test ./internal/tui

# Regenerate goldens after intentional UI change
snap-update:
	go test ./internal/tui -update

# Per-width PNG snapshots via charm freeze (install: brew install charmbracelet/tap/freeze)
widths:
	./tools/snap-widths.sh

# Animated GIF of full flow via vhs (install: brew install vhs)
demo:
	vhs demos/welcome.tape

# Page through every snapshot golden in less. Filters via args.
#   make preview                      # all
#   make preview ARGS=welcome         # filter
#   make preview ARGS="welcome lfg"   # multiple filters
preview:
	./tools/preview.sh $(ARGS)

docker:
	docker build -t lfg .

docker-run:
	docker run --rm -it lfg

clean:
	rm -f lfg lfg-linux
	rm -rf snaps demos/*.gif
