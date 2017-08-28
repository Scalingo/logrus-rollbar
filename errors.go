package logrus_rollbar

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/stvp/rollbar"
)

type causer interface {
	Cause() error
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func BuildStack(err error) rollbar.Stack {
	stack := rollbar.Stack{}

	// We're going to the deepest call
	for {
		c, ok := err.(causer)
		if !ok {
			break
		}
		err = c.Cause()
	}

	// Return an empty stack
	tracer, ok := err.(stackTracer)
	if !ok {
		return stack
	}

	errorsStack := tracer.StackTrace()
	for _, f := range errorsStack {
		line, _ := strconv.Atoi(fmt.Sprintf("%d", f))
		frame := rollbar.Frame{
			Filename: fmt.Sprintf("%+s", f),
			Line:     line,
			Method:   fmt.Sprintf("%n", f),
		}
		stack = append([]rollbar.Frame{frame}, stack...)
	}

	return stack
}

// BuildStackWithSkip concatenates the stack given by the current execution flow
// and the stack determined by the pkg/errors error
func BuildStackWithSkip(err error, skip int) rollbar.Stack {
	errStack := BuildStack(err)
	execStack := rollbar.BuildStack(skip)
	return append(errStack, execStack...)
}
