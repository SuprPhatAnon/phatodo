package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunAcceptsTrekkerCompatibleTaskCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--toon", "task", "list", "--status", "in_progress"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "task list") {
		t.Fatalf("expected accepted command output, got %q", stdout.String())
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"unknown"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("expected unknown command error, got %q", stderr.String())
	}
}
