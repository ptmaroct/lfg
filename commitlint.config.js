// Conventional Commits config used by both the server-side
// commitlint GitHub Action and (optionally) any local Node-based
// pre-commit hook a contributor wires up.
//
// Types are aligned with what release-please understands so the
// stable + beta release flows can derive sensible version bumps.
//
// See:
//   https://www.conventionalcommits.org/
//   https://github.com/googleapis/release-please?tab=readme-ov-file#how-should-i-write-my-commits
module.exports = {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "type-enum": [
      2,
      "always",
      [
        "feat", // user-facing feature → minor bump
        "fix", // user-facing bug fix → patch bump
        "perf", // performance improvement → patch bump
        "refactor", // internal refactor, no behavior change
        "docs", // documentation only
        "test", // test changes only
        "build", // build system, dependencies (go.mod, Dockerfile, Makefile)
        "ci", // CI config (.github/workflows, goreleaser)
        "chore", // anything else that doesn't fit
        "style", // formatting, whitespace, lint fixes
        "revert", // revert of an earlier commit
        "release", // release-please / goreleaser machinery commits
      ],
    ],
    // Subject line: at least 4 chars, max 100, no trailing period.
    "subject-min-length": [2, "always", 4],
    "subject-max-length": [2, "always", 100],
    "subject-full-stop": [2, "never", "."],
    // Header itself capped slightly higher to leave room for a long scope.
    "header-max-length": [2, "always", 120],
    // Body & footer line wrapping is a courtesy, not a hard rule.
    "body-max-line-length": [0],
    "footer-max-line-length": [0],
  },
};
