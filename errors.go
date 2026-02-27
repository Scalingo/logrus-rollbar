package logrus_rollbar

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rollbar/rollbar-go"
)

// Rollbar package expect such error:
// type CauseStacker interface {
//   error
//   Cause() error
//   Stack() Stack
// }

var (
	_ rollbar.CauseStacker = wrappedError{}
)

type wrappedError struct {
	err error
	msg string
}

func (err wrappedError) Error() string {
	return err.msg
}

type causer interface {
	Cause() error
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func Wrap(msg string, err error) wrappedError {
	return wrappedError{msg: msg, err: err}
}

func (err wrappedError) Cause() error {
	return err.err
}

func (werr wrappedError) Stack() []runtime.Frame {
	var stack []runtime.Frame
	err := werr.err

	// We're going to the deepest call
	for {
		c, ok := err.(causer)
		if !ok {
			break
		}
		err = c.Cause()
	}

	// Return nil stack so rollbar-go can fallback to runtime.Callers
	// The relevant code in rollbar-go: https://github.com/rollbar/rollbar-go/blob/4c2ee8c66b8ae695aff08fc331445a4639036e68/rollbar.go#L89
	// `stackTracer` is implemented by errors that carry a captured stack trace for example:
	// - github.com/pkg/errors via errors.New/Wrap/Wrapf has a full stack trace,
	// - github.com/go-errgo/errgo via errgo.New which includes the caller location,
	// Plain errors.New from the standard library do not include a stack trace.
	tracer, ok := err.(stackTracer)
	if !ok {
		return nil
	}

	errorsStack := tracer.StackTrace()
	for i := len(errorsStack) - 1; i >= 0; i-- {
		f := errorsStack[i]
		line, _ := strconv.Atoi(fmt.Sprintf("%d", f))
		frame := runtime.Frame{
			File:     fmt.Sprintf("%+s", f),
			Line:     line,
			Function: fmt.Sprintf("%n", f),
		}
		stack = append([]runtime.Frame{frame}, stack...)
	}

	return stack
}
