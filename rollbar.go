package logrus_rollbar

import (
	"net/http"

	"github.com/stvp/rollbar"
)

type Sender interface {
	RequestErrorWithStack(string, *http.Request, error, rollbar.Stack) error
}

type RollbarSender struct{}

func (s RollbarSender) RequestErrorWithStack(severity string, req *http.Request, err error, stack rollbar.Stack) error {
	rollbar.RequestErrorWithStack(severity, req, err, stack)
	return nil
}
