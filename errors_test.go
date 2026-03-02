package logrus_rollbar

import (
	stderrors "errors"
	"runtime"
	"strings"
	"testing"

	pkgerrors "github.com/pkg/errors"
)

func TestWrapStack_returnNilForErrorWithoutStackTrace(t *testing.T) {
	wrapped := Wrap("message", stderrors.New("boom"))
	stack := wrapped.Stack()
	if stack != nil {
		t.Fatalf("expected nil stack, got %d frame(s)", len(stack))
	}
}

func TestWrapStack_returnFramesForErrorWithStackTrace(t *testing.T) {
	wrapped := Wrap("message", pkgerrors.New("boom"))
	stack := wrapped.Stack()
	if len(stack) == 0 {
		t.Fatal("expected non-empty stack")
	}
}

func TestWrapStack_returnFramesForErrorWithStackTraceCheckStack(t *testing.T) {
	wrapped := Wrap("message", errorsFoo())
	stack := wrapped.Stack()
	if stack == nil {
		t.Fatal("expected non-nil stack")
	}
	if len(stack) == 0 {
		t.Fatal("expected non-empty stack")
	}

	if stack[0].Function != "errorsBar" {
		t.Fatalf("expected first frame function errorsBar, got %q", stack[0].Function)
	}
	if !strings.HasSuffix(stack[0].File, "hook_pkgerrors_test.go") {
		t.Fatalf("expected first frame file to end with hook_pkgerrors_test.go, got %q", stack[0].File)
	}
	if stack[0].Line == 0 {
		t.Fatal("expected first frame line number to be set")
	}

	if !stackContainsFunction(stack, "errorsFoo") {
		t.Fatal("expected stack to contain errorsFoo frame")
	}
}

func stackContainsFunction(stack []runtime.Frame, function string) bool {
	for _, frame := range stack {
		if frame.Function == function {
			return true
		}
	}
	return false
}
