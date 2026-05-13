# Versioning & supply-chain safety

`lfg` ships a default preset that touches every developer's most
sensitive surface — package registries, install scripts, AI CLIs.
The 2025–2026 incident wave (Shai-Hulud, chalk/debug, axios,
tj-actions, TanStack, Bitwarden CLI, Trivy) made it clear that
"install latest" is no longer an acceptable default. This doc
explains how `lfg` keeps its default preset fresh AND safe, and how
that pipeline reaches end users.

## What we pin

Every tool in `internal/preset/preset.go` gets a corresponding entry
in `internal/preset/pins.toml`. There are two pin types and a tool can
carry either or both:

- `version = "x.y.z"` — an exact registry version. Applied at install
  time by `preset.Tool.ResolvedInstall(goos)`, which rewrites the
  install command (e.g. `npm install -g @openai/codex` →
  `npm install -g @openai/codex@0.128.0`, `mise use -g node@lts` →
  `mise use -g node@24.15.0`). Only npm + mise sources get textual
  substitution; brew commands are left alone because brew formulas
  can't be pinned through plain `brew install <name>@x.y.z`.
- `sha256 = "<hex>"` — the digest of an install script body
  (`curl -fsSL <url> | sh` patterns). The installer downloads the
  script, refuses to execute it unless the digest matches, then runs
  the verified-on-disk copy. See `internal/installer/verify.go`.

```toml
# internal/preset/pins.toml (excerpt)
bumped_at = "2026-05-13T13:55:44Z"

[pins."barebones/node-lts"]
version = "24.15.0"

[pins."barebones/uv"]
sha256 = "14b6b88d74b9300d906d9754118afb4d4151e78ad294f026047b7227f5611760"

[pins."barebones/mise"]
version = "2026.5.6"
sha256  = "2ac97541052b681a139d12b4cf841e05474db8c9df05c45ecbb83fd1a105cc8b"
```

The key format is `<bundleID>/<toolName>`.

## The weekly bumper

`cmd/preset-bump/main.go` is a small Go program that:

1. Walks every tool in `preset.All()`.
2. For each npm-distributed tool, hits
   `https://registry.npmjs.org/<pkg>` and reads the full `time` map.
   Picks `dist-tags.latest` IF its publish timestamp is older than
   the quarantine window AND it isn't deprecated / a pre-release.
   Otherwise walks back to the next eligible stable version.
3. For `node-lts`, reads `https://nodejs.org/dist/index.json` and
   picks the first LTS entry whose `date` is older than the
   quarantine.
4. For brew formulas, takes `versions.stable` from
   `formulae.brew.sh` (brew's tap is curated; we don't add an
   additional quarantine on top).
5. For every curl-piped install script in either `install_mac` or
   `install_linux`, re-downloads the URL and computes its SHA256.
   Drift relative to the previous pin gets flagged in the PR body
   with `⚠ install script CONTENT CHANGED`.
6. Writes the new `pins.toml` if anything moved.

Run locally:

```sh
go run ./cmd/preset-bump --min-age-days=7 --dry-run   # preview
go run ./cmd/preset-bump --min-age-days=7             # rewrite pins.toml
```

`.github/workflows/preset-bump.yml` runs it every Monday at 06:00 UTC
and uses `peter-evans/create-pull-request` (SHA-pinned) to open a
labelled PR. A human is the security gate.

## Quarantine window

Default is **7 days** (`--min-age-days=7`). Rationale:

| Incident | Time live before unpublished |
|---|---|
| chalk/debug qix wave (Sep 2025) | ~2 hours |
| Shai-Hulud 1.0 (Sep 2025) | hours; first wave 24 h |
| axios @ 1.14.1 / 0.30.4 (Mar 2026) | ~3 hours |
| Shai-Hulud 2.0 (Nov 2025) | hours; some packages days |
| TanStack Mini Shai-Hulud (May 2026) | hours |

7 d gives the npm security team, Socket, Snyk, Sonatype, and the
maintainer themselves enough time to detect + unpublish a poisoned
release before it can become a pin in our default preset. Versions
within the window stay un-pinned (the bumper just keeps the previous
pin), so users still install the older known-good version.

## Release cycle — how fresh pins reach users

```
Mon 06:00 UTC   preset-bump opens a PR
Mon–Wed         maintainer reviews + merges
Wed (auto)      release-on-pins.yml fires on the path-trigger,
                computes next patch tag, pushes it
Wed (auto)      release.yml runs goreleaser → binaries + brew bottle
Wed–Sun         delivery:
                  (a) net-new installs get fresh binary via brew/curl
                  (b) existing binaries fetch newer pins.toml from
                      raw.githubusercontent.com on launch (2 s timeout)
                  (c) staleness chip on welcome screen nudges users
                      past 30 d to `lfg update`
```

This means a binary released in January can still install May's
pins. The runtime path is `internal/preset/pins.go`'s `FetchRemotePins`
+ `SetRemotePins`. It is fail-closed: any error (no network, 4xx,
parse failure) keeps the embedded pins.

Cosign signature verification on the remote pins file is on the
roadmap; until then the remote fetch's only privilege is to swap the
in-memory pin map. The install step's own SHA256 check (npm
registries verify via GPG-signed dist-tags; curl scripts verified
against `pin_sha256`) is the actual gate.

## Override

If you don't trust our pin set, pass a `--config` of your own:

```toml
[[bundles]]
id = "minimal"
name = "minimal"
default = true

  [[bundles.tools]]
  name        = "yarn"
  source      = "npm"
  install_mac = "npm install -g yarn"
  install_linux = "npm install -g yarn"
  pin         = "1.22.22"           # exact npm version

  [[bundles.tools]]
  name        = "uv"
  source      = "curl"
  install_mac = "curl -LsSf https://astral.sh/uv/install.sh | sh"
  install_linux = "curl -LsSf https://astral.sh/uv/install.sh | sh"
  pin_sha256  = "14b6b88d74b9300d906d9754118afb4d4151e78ad294f026047b7227f5611760"
```

When `pin_sha256` is set, the same `runVerifiedCurl` code path runs
— there is no special-case for custom presets.

## What about GitHub Actions in this repo?

Every `uses:` in `.github/workflows/*` is pinned to a full commit SHA
because tags can be rewritten — that's exactly the mechanism behind
`tj-actions/changed-files` (CVE-2025-30066). The human-readable tag
follows each SHA as a comment so reviewers can see the intent.
`.github/dependabot.yml` keeps the SHAs fresh with weekly reviewed
PRs.

## Threat model summary

| Attack | Mitigation |
|---|---|
| npm maintainer account compromised, malicious version published | 7-day quarantine + pin substitution = we don't install the fresh malicious version |
| Install script body swapped on a CDN | SHA256 verification in `runVerifiedCurl` before exec |
| GitHub Action tag rewritten to point at malicious commit | Every `uses:` SHA-pinned |
| pins.toml on main tampered with | Remote fetch fail-closed to embedded copy; install-time SHA256 still runs |
| Bumper PR contains a poisoned version slipping past the 7-day window | Human review of every bumper PR; install-time SHA256 + npm integrity hashes catch tampering |
| Cosign-signed pins.toml | **Open** — planned follow-up; until then, trust boundary is the maintainer's commit signing |

## Open items

- Cosign signature on `pins.toml` so the remote fetch can verify
  before adopting.
- Mirror critical install scripts (brew, mise.run, uv) into the lfg
  repo so we read from a hash-pinned commit instead of upstream CDN.
- Wrap npm install steps with the [Socket CLI](https://socket.dev)
  for behavioural analysis pre-install.
- Surface a `--quarantine-days` flag on `lfg apply` so security-
  conscious users can demand stricter posture (14 d, 30 d).
