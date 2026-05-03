# lfg — build, test, snapshot, demo
.PHONY: build run test snap snap-update widths demo docker docker-bare docker-run docker-shell docker-test docker-test-bare clean release release-snapshot docs-install docs docs-build

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

# Bare ubuntu image — no preinstalled brew. Use to exercise lfg's full
# brew bootstrap path. Slower per build (~2min on first run) so default
# `make docker` keeps brew baked in for fast iteration.
docker-bare:
	docker build --build-arg INCLUDE_BREW=0 -t lfg-bare .

docker-run:
	docker run --rm -it lfg

# Fresh blank Ubuntu shell with `lfg` on PATH. Type `lfg` to launch.
# Builds image on first run (slow, ~2min — pulls ubuntu base + brew).
# After that, layer cache makes rebuild instant. Re-run `make docker`
# manually after source changes to refresh `lfg`.
docker-shell:
	@docker image inspect lfg >/dev/null 2>&1 || $(MAKE) docker
	docker run --rm -it --entrypoint bash lfg

# One-shot test loop: rebuild image (so latest binary is baked), launch
# fresh container, auto-run `lfg --debug`, drop into bash on exit so
# the user can poke around (cat the install log, run the freshly
# installed CLIs, etc.). Tightest iteration loop for end-to-end QA.
docker-test:
	docker build -t lfg .
	docker run --rm -it --entrypoint bash lfg -lc 'lfg --debug; echo; echo "--- lfg exited; debug log + transcripts in ~/.config/lfg/logs/ ---"; exec bash'

# Same one-shot loop but on the bare ubuntu image (no preinstalled
# brew). Use when verifying changes that touch the brew bootstrap.
docker-test-bare:
	docker build --build-arg INCLUDE_BREW=0 -t lfg-bare .
	docker run --rm -it --entrypoint bash lfg-bare -lc 'lfg --debug; echo; echo "--- lfg exited; debug log + transcripts in ~/.config/lfg/logs/ ---"; exec bash'

# Local snapshot release via goreleaser — builds all platforms into
# ./dist/ without uploading anywhere. Useful for smoke-testing the
# release config before pushing a tag.
#   brew install goreleaser/tap/goreleaser
release-snapshot:
	goreleaser release --snapshot --clean

# Real release — only run from CI on a tagged push (see
# .github/workflows/release.yml). Local invocation requires a
# GITHUB_TOKEN with repo:write scope.
release:
	goreleaser release --clean

clean:
	rm -f lfg lfg-linux
	rm -rf snaps demos/*.gif dist

# Docs site (Astro Starlight). Lives in docs/, deploys to GitHub Pages
# via .github/workflows/docs.yml on push to main.
docs-install:
	cd docs && npm install

docs:
	cd docs && npm run dev

docs-build:
	cd docs && npm run build
