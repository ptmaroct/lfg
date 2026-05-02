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

# Bare-bones runtime. We deliberately do NOT preinstall Homebrew here —
# the whole point of the docker test loop is to exercise lfg's brew
# bootstrap path on a clean Ubuntu box. Tools that brew installs (mise,
# node, etc.) all flow from lfg's own install steps.
#
# `procps` ships `ps` which the Homebrew installer probes; `file` is
# also required. Both stay so when lfg DOES install brew, the
# prereqs are already there.
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

# Non-root user. Passwordless sudo so lfg's brew bootstrap (which
# refuses root) and any post-install steps that need apt packages
# (e.g. `sudo apt install chromium` for agent-browser on arm64) can
# run without prompting.
RUN useradd -m -s /bin/bash dev \
    && echo 'dev ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/dev \
    && chmod 0440 /etc/sudoers.d/dev

USER dev
WORKDIR /home/dev

# Drop the lfg binary in.
USER root
COPY --from=build /out/lfg /usr/local/bin/lfg
RUN chmod +x /usr/local/bin/lfg

USER dev
ENV TERM=xterm-256color
ENV NONINTERACTIVE=1
CMD ["lfg"]
