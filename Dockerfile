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

# Brew prerequisites + general dev basics. `procps` ships `ps` which the
# Homebrew installer probes; `file` is also required.
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

# Non-root user for brew. Passwordless sudo so the user can install
# system packages during testing if needed.
RUN useradd -m -s /bin/bash dev \
    && echo 'dev ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/dev \
    && chmod 0440 /etc/sudoers.d/dev

USER dev
WORKDIR /home/dev

# Install Homebrew non-interactively.
ENV NONINTERACTIVE=1
RUN /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Wire brew into the user's PATH for both bash and zsh sessions.
RUN echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /home/dev/.bashrc \
    && echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /home/dev/.zshrc \
    && echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /home/dev/.profile

# Make brew visible to non-login shells used by `docker run` CMD.
ENV PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:${PATH}"
ENV HOMEBREW_PREFIX="/home/linuxbrew/.linuxbrew"
ENV HOMEBREW_CELLAR="/home/linuxbrew/.linuxbrew/Cellar"
ENV HOMEBREW_REPOSITORY="/home/linuxbrew/.linuxbrew/Homebrew"
ENV MANPATH="/home/linuxbrew/.linuxbrew/share/man:${MANPATH}"
ENV INFOPATH="/home/linuxbrew/.linuxbrew/share/info:${INFOPATH}"

# Drop the lfg binary in.
USER root
COPY --from=build /out/lfg /usr/local/bin/lfg
RUN chmod +x /usr/local/bin/lfg

USER dev
ENV TERM=xterm-256color
CMD ["lfg"]
