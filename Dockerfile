# Multi-stage Dockerfile for testing lfg in a clean Linux env.
# Stage 1: build static binary inside Go image.
# Stage 2: Ubuntu runtime with Homebrew so brew installer paths can be exercised.
#
# Note: Homebrew on Linux refuses to install as root, so we provision a
# non-root user `dev`. Image size ~1.5GB due to brew + its toolchain.

FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/lfg ./cmd/lfg

FROM ubuntu:24.04

# NOTE: INCLUDE_BREW is declared LATER, just before the brew RUN. We
# deliberately don't `ARG` it here at the top of the stage because
# BuildKit folds every in-scope ARG value into the cache key of every
# subsequent RUN — even ones that don't reference the ARG. Putting it
# at the top meant `make docker-bare` (INCLUDE_BREW=0) and `make
# docker` (INCLUDE_BREW=1) produced different cache keys for the
# expensive apt-get + useradd layers, forcing 2-min full rebuilds on
# every variant switch. Declaring it later lets the apt + user layers
# share cache across both build variants.

# Brew prerequisites + general dev basics. `procps` ships `ps` which the
# Homebrew installer probes; `file` is also required even when we skip
# the brew install — keeps the prereqs in place so a later in-container
# `lfg` run can install brew without a missing-dep surprise.
RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        curl \
        file \
        git \
        procps \
        sudo \
        zsh \
    && rm -rf /var/lib/apt/lists/*

# Non-root user. Passwordless sudo so brew (which refuses root) and
# post-install steps (e.g. `sudo apt install chromium` for
# agent-browser on arm64) run without prompting.
RUN useradd -m -s /bin/bash dev \
    && echo 'dev ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/dev \
    && chmod 0440 /etc/sudoers.d/dev

USER dev
WORKDIR /home/dev

# INCLUDE_BREW toggles whether linuxbrew is preinstalled.
#   1 (default) — fast iteration; brew shows under ALREADY INSTALLED.
#   0           — bare ubuntu; lfg's brew bootstrap path is exercised.
# Override per-build:  docker build --build-arg INCLUDE_BREW=0 -t lfg-bare .
#
# Declared HERE (not at top of stage) so it only affects this RUN's
# cache key, not the apt + useradd layers above. Lets a brewed and a
# bare build share the heavy ubuntu setup.
ARG INCLUDE_BREW=1
RUN if [ "$INCLUDE_BREW" = "1" ]; then \
        export NONINTERACTIVE=1 && \
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" && \
        echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /home/dev/.bashrc && \
        echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /home/dev/.zshrc && \
        echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /home/dev/.profile; \
    else \
        echo "skip brew install (INCLUDE_BREW=0)"; \
    fi

# Brew env vars baked in unconditionally — paths resolve only when the
# directories exist, so they're harmless when brew was skipped. Saves
# the conditional dance for ENV (which isn't ARG-aware anyway) and
# means `lfg` will pick up brew if it's installed later inside the
# container.
ENV PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:${PATH}"
ENV HOMEBREW_PREFIX="/home/linuxbrew/.linuxbrew"
ENV HOMEBREW_CELLAR="/home/linuxbrew/.linuxbrew/Cellar"
ENV HOMEBREW_REPOSITORY="/home/linuxbrew/.linuxbrew/Homebrew"

# Drop the lfg binary in.
USER root
COPY --from=build /out/lfg /usr/local/bin/lfg
RUN chmod +x /usr/local/bin/lfg

USER dev
ENV TERM=xterm-256color
CMD ["lfg"]
