package main

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestRunSelfUpdateSuccess(t *testing.T) {
	oldExec := execCommand
	t.Cleanup(func() { execCommand = oldExec })

	execCommand = func(name string, args ...string) *exec.Cmd {
		if name != "go" {
			t.Fatalf("unexpected command: %s", name)
		}
		if len(args) != 2 || args[0] != "install" || args[1] != updateModule {
			t.Fatalf("unexpected args: %v", args)
		}
		return exec.Command("sh", "-c", "exit 0")
	}

	if err := runSelfUpdate(); err != nil {
		t.Fatalf("runSelfUpdate() error = %v", err)
	}
}

func TestRunSelfUpdateGoMissing(t *testing.T) {
	oldExec := execCommand
	t.Cleanup(func() { execCommand = oldExec })

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("/definitely-missing-go-binary")
	}

	err := runSelfUpdate()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Go is not installed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSelfUpdateCommandFailureIncludesOutput(t *testing.T) {
	oldExec := execCommand
	t.Cleanup(func() { execCommand = oldExec })

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "echo 'network failure' 1>&2; exit 1")
	}

	err := runSelfUpdate()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "network failure") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestErrorsIsExecNotFound(t *testing.T) {
	err := &exec.Error{Name: "go", Err: exec.ErrNotFound}
	if !errorsIsExecNotFound(err) {
		t.Fatal("expected exec.ErrNotFound to be detected")
	}
	if errorsIsExecNotFound(errors.New("other")) {
		t.Fatal("did not expect generic error to match")
	}
}
