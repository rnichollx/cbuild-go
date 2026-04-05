package csetupapp

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() failed: %v", err)
	}
	os.Stdout = w

	var buf bytes.Buffer
	done := make(chan error, 1)
	go func() {
		_, copyErr := io.Copy(&buf, r)
		done <- copyErr
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("stdout close failed: %v", err)
	}
	os.Stdout = origStdout

	if err := <-done; err != nil {
		t.Fatalf("stdout capture failed: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("stdout reader close failed: %v", err)
	}

	return buf.String()
}

func TestCSetupHelpSubcommandShowsSubcommandUsage(t *testing.T) {
	output := captureStdout(t, func() {
		if err := CSetup.Run(context.Background(), []string{"help", "init"}); err != nil {
			t.Fatalf("CSetup.Run() failed: %v", err)
		}
	})

	if !strings.Contains(output, "Usage: csetup init <workspace>") {
		t.Fatalf("expected init usage in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Prefer `csetup help init` for command-specific help.") {
		t.Fatalf("expected help preference note in output, got:\n%s", output)
	}
}
