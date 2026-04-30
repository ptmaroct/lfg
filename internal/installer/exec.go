package installer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// execRun is the default implementation. Tests swap it for a fake by
// assigning to the package-level variable.
var execRun = realExecRun

// Run executes a shell command line and pipes stdout/stderr line-by-line
// into `out`. Honors ctx cancellation by killing the underlying process.
// Returns the exec error (non-nil on non-zero exit).
//
// `tool` is attached to every emitted Line so the TUI can group output
// per tool when several installs run back-to-back.
func runCmd(ctx context.Context, tool, cmdline string, out chan<- Line) error {
	return execRun(ctx, tool, cmdline, out)
}

func realExecRun(ctx context.Context, tool, cmdline string, out chan<- Line) error {
	if cmdline == "" {
		return fmt.Errorf("no install command for %s", tool)
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdline)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go scanLines(&wg, stdout, "stdout", tool, out)
	go scanLines(&wg, stderr, "stderr", tool, out)
	wg.Wait()

	return cmd.Wait()
}

// scanLines reads `r` line-by-line, forwarding each as a Line. Buffer
// size is bumped to 1MB so a single very long line (e.g. brew progress
// bar mid-line update) doesn't break the scanner.
func scanLines(wg *sync.WaitGroup, r io.Reader, stream, tool string, out chan<- Line) {
	defer wg.Done()
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		out <- Line{Tool: tool, Stream: stream, Text: sc.Text()}
	}
}
