package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ptmaroct/lfg/internal/preset"
)

func TestBuildAliasBlock_Bash(t *testing.T) {
	aliases := []preset.Alias{
		{Name: "gd", Command: "git checkout develop && git pull"},
		{Name: "reload", Command: "exec $SHELL", FishCommand: "exec fish"},
	}
	got := buildAliasBlock("/home/u/.bashrc", aliases)
	wantLines := []string{
		shellAliasMarker,
		`alias gd='git checkout develop && git pull'`,
		`alias reload='exec bash'`,
		shellAliasEndMarker,
	}
	want := strings.Join(wantLines, "\n")
	if got != want {
		t.Errorf("bash block mismatch\n--- want\n%s\n--- got\n%s", want, got)
	}
}

func TestBuildAliasBlock_Zsh(t *testing.T) {
	aliases := []preset.Alias{
		{Name: "reload", Command: "exec $SHELL"},
	}
	got := buildAliasBlock("/home/u/.zshrc", aliases)
	if !strings.Contains(got, "alias reload='exec zsh'") {
		t.Errorf("expected zsh reload line, got:\n%s", got)
	}
}

func TestBuildAliasBlock_Fish(t *testing.T) {
	aliases := []preset.Alias{
		{Name: "gd", Command: "git diff"},
		{Name: "reload", Command: "exec $SHELL", FishCommand: "exec fish"},
	}
	got := buildAliasBlock("/home/u/.config/fish/config.fish", aliases)
	if !strings.Contains(got, "alias gd 'git diff'") {
		t.Errorf("expected fish-style alias line, got:\n%s", got)
	}
	if !strings.Contains(got, "function reload; exec fish; end") {
		t.Errorf("expected fish reload function, got:\n%s", got)
	}
}

func TestBuildAliasBlock_Empty(t *testing.T) {
	if got := buildAliasBlock("/home/u/.bashrc", nil); got != "" {
		t.Errorf("empty alias list should produce empty block, got: %q", got)
	}
}

func TestPosixSingleQuote_EscapesQuote(t *testing.T) {
	got := posixSingleQuote(`echo 'hi'`)
	want := `'echo '\''hi'\'''`
	if got != want {
		t.Errorf("posix quote escape: want %q, got %q", want, got)
	}
}

func TestEnsureShellAliases_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	rc := filepath.Join(dir, ".bashrc")
	if err := os.WriteFile(rc, []byte("# pre-existing rc\nalias old='echo old'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/bin/bash")

	aliases := []preset.Alias{
		{Name: "gd", Command: "git status"},
	}
	if _, err := EnsureShellAliases(aliases); err != nil {
		t.Fatalf("EnsureShellAliases: %v", err)
	}
	data, err := os.ReadFile(rc)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	if !strings.Contains(body, shellAliasMarker) {
		t.Errorf("rc missing alias start marker:\n%s", body)
	}
	if !strings.Contains(body, "alias gd='git status'") {
		t.Errorf("rc missing alias line:\n%s", body)
	}
	if !strings.Contains(body, "alias old='echo old'") {
		t.Errorf("rc lost pre-existing alias:\n%s", body)
	}

	// Second pass with empty list should remove the block.
	if _, err := EnsureShellAliases(nil); err != nil {
		t.Fatalf("EnsureShellAliases empty: %v", err)
	}
	data, _ = os.ReadFile(rc)
	body = string(data)
	if strings.Contains(body, shellAliasMarker) {
		t.Errorf("empty input should strip block, still present:\n%s", body)
	}
	if !strings.Contains(body, "alias old='echo old'") {
		t.Errorf("pre-existing alias lost on empty pass:\n%s", body)
	}
}

func TestExistingAliases_SkipsManagedBlock(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	rc := filepath.Join(dir, ".bashrc")
	body := strings.Join([]string{
		"alias user_owned='echo'",
		shellAliasMarker,
		"alias managed='inside lfg block'",
		shellAliasEndMarker,
		"alias post='after'",
	}, "\n")
	if err := os.WriteFile(rc, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/bin/bash")

	got := ExistingAliases()
	if _, ok := got["managed"]; ok {
		t.Errorf("ExistingAliases must not surface aliases inside lfg block: %+v", got)
	}
	if _, ok := got["user_owned"]; !ok {
		t.Errorf("ExistingAliases must surface user-owned aliases: %+v", got)
	}
	if _, ok := got["post"]; !ok {
		t.Errorf("ExistingAliases must surface aliases after the block: %+v", got)
	}
}
