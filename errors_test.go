package logrus_rollbar

import (
	stderrors "errors"
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
