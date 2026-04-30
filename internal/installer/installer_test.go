package installer

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/anuj/lfg/internal/preset"
)

// fakeRun records every invocation so tests can assert on call order
// and command contents. Returns the configured error for that command.
type fakeCall struct {
	tool, cmdline string
}

func withFakeExec(t *testing.T, errFor map[string]error) *[]fakeCall {
	t.Helper()
	calls := []fakeCall{}
	var mu sync.Mutex
	prev := execRun
	execRun = func(ctx context.Context, tool, cmdline string, out chan<- Line) error {
		mu.Lock()
		calls = append(calls, fakeCall{tool, cmdline})
		mu.Unlock()
		out <- Line{Tool: tool, Stream: "stdout", Text: "fake output"}
		if errFor != nil {
			if err, ok := errFor[tool]; ok {
				return err
			}
		}
		return nil
	}
	t.Cleanup(func() { execRun = prev })
	return &calls
}

func drainLines(out <-chan Line, done <-chan struct{}) []Line {
	var got []Line
	for {
		select {
		case l := <-out:
			got = append(got, l)
		case <-done:
			return got
		}
	}
}

func TestPlan_BootstrapDeduped(t *testing.T) {
	bundles := []preset.Bundle{
		{ID: "a", Tools: []preset.Tool{
			{Name: "git", Source: "brew", InstallMac: "brew install git"},
			{Name: "fzf", Source: "brew", InstallMac: "brew install fzf"},
			{Name: "node", Source: "mise", InstallMac: "mise use -g node@lts"},
		}},
	}
	sel := map[string]bool{"a/git": true, "a/fzf": true, "a/node": true}
	plan := Plan(bundles, sel)

	if len(plan) != 5 { // brew bootstrap + 2 brew tools + mise bootstrap + 1 mise tool
		t.Fatalf("plan len: got %d, want 5: %+v", len(plan), plan)
	}
	if !plan[0].Bootstrap || plan[0].Backend != "brew" {
		t.Errorf("step 0: expected brew bootstrap, got %+v", plan[0])
	}
	if !plan[3].Bootstrap || plan[3].Backend != "mise" {
		t.Errorf("step 3: expected mise bootstrap, got %+v", plan[3])
	}
}

func TestRun_DispatchesPerBackend(t *testing.T) {
	calls := withFakeExec(t, nil)
	out := make(chan Line, 32)

	tools := []preset.Tool{
		{Name: "git", Source: "brew", InstallMac: "brew install git", InstallLinux: "sudo apt-get install -y git"},
	}
	bundles := []preset.Bundle{{ID: "a", Tools: tools}}
	plan := Plan(bundles, map[string]bool{"a/git": true})

	done := make(chan struct{})
	var failed []FailedStep
	go func() {
		failed = Run(context.Background(), plan, out)
		close(done)
	}()
	_ = drainLines(out, done)

	if len(failed) != 0 {
		t.Errorf("unexpected failures: %+v", failed)
	}
	// brew bootstrap may short-circuit when brew is already installed on
	// the host running tests, so we only require the git install itself.
	foundGit := false
	for _, c := range *calls {
		if c.tool == "git" {
			foundGit = true
		}
	}
	if !foundGit {
		t.Errorf("git install never executed; calls=%+v", *calls)
	}
}

func TestRun_CapturesFailureContinues(t *testing.T) {
	calls := withFakeExec(t, map[string]error{"bad": errors.New("boom")})
	out := make(chan Line, 32)

	bundles := []preset.Bundle{{ID: "a", Tools: []preset.Tool{
		{Name: "good", Source: "custom", InstallMac: "echo ok", InstallLinux: "echo ok"},
		{Name: "bad", Source: "custom", InstallMac: "false", InstallLinux: "false"},
		{Name: "after", Source: "custom", InstallMac: "echo after", InstallLinux: "echo after"},
	}}}
	plan := Plan(bundles, map[string]bool{"a/good": true, "a/bad": true, "a/after": true})

	done := make(chan struct{})
	var failed []FailedStep
	go func() {
		failed = Run(context.Background(), plan, out)
		close(done)
	}()
	_ = drainLines(out, done)

	if len(failed) != 1 || failed[0].Step.Tool.Name != "bad" {
		t.Errorf("expected single failure on 'bad', got %+v", failed)
	}
	// All three tools should still have been attempted.
	if len(*calls) < 3 {
		t.Errorf("queue aborted early: only %d calls", len(*calls))
	}
}
