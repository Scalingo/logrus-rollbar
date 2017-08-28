package logrus_rollbar

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"gopkg.in/errgo.v1"

	"github.com/Sirupsen/logrus"
	"github.com/Soulou/errgo-rollbar"
	"github.com/stvp/rollbar"
)

var (
	SeverityCritical = "critical"
)

type hook struct {
	Sender
	SkipLevel int
}

func New(skipLevel int) logrus.Hook {
	return hook{Sender: RollbarSender{}, SkipLevel: skipLevel}
}

func (h hook) Fire(entry *logrus.Entry) error {
	var req *http.Request

	if r, ok := entry.Data["req"]; ok {
		upstreamReq, ok := r.(*http.Request)
		if ok {
			req, _ = http.NewRequest(upstreamReq.Method, upstreamReq.URL.String(), nil)
			req.RemoteAddr = upstreamReq.RemoteAddr
			for key, val := range upstreamReq.Header {
				// We don't want to log credentials
				if key == "Authorization" {
					continue
				}
				req.Header[key] = val
			}

			// Replacing the request struct by something simpler in the entry fields
			entry.Data["req"] = fmt.Sprintf(
				"%s %s %s",
				req.Method, req.URL, req.RemoteAddr,
			)
		}
	}

	// All the fields which aren't level|msg|error|time|req are added
	// to the headers of the request which will be sent to Rollbar
	// The main goal is to be able to see all the values on Rollbar dashboard
	fields := []*rollbar.Field{}
	for val, key := range entry.Data {
		if val != "level" && val != "msg" && val != "error" && val != "time" && val != "req" {
			fields = append(fields, &rollbar.Field{Name: val, Data: key})
		}
	}

	// If there is an error field, we want it to be part of Rollbar ticket name
	var (
		errorMsg error
		err      error
		stack    rollbar.Stack
	)

	if entry.Data["error"] != nil {
		err = entry.Data["error"].(error)

		// skip level is to avoid stack trace frame from the hook itself
		switch err.(type) {
		case *errgo.Err:
			stack = errgorollbar.BuildStackWithSkip(err, 5+h.SkipLevel)
		default:
			stack = BuildStackWithSkip(err, 6+h.SkipLevel)
		}

		errorTxt := new(bytes.Buffer)
		errorTxt.WriteString(err.Error())
		if msg, ok := entry.Data["msg"]; ok && msg != nil {
			if strMsg, ok := entry.Data["msg"].(string); ok {
				errorTxt.WriteString("- " + strMsg)
			}
		}
		errorMsg = fmt.Errorf(errorTxt.String())
	} else {
		errorMsg = errors.New(fmt.Sprintf("%v", entry.Data["msg"]))
	}

	severity := rollbar.ERR
	if entry.Data["severity"] == SeverityCritical {
		severity = rollbar.CRIT
	}

	if req == nil {
		h.Sender.ErrorWithStack(severity, errorMsg, stack, fields...)
	} else {
		h.Sender.RequestErrorWithStack(severity, req, errorMsg, stack, fields...)
	}
	return nil
}

func (h hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}
