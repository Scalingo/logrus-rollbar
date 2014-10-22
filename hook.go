package logrus_rollbar

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/Soulou/errgo-rollbar"
	"github.com/stvp/rollbar"
)

type Hook struct{ SkipLevel int }

func (hook Hook) Fire(entry *logrus.Entry) error {
	var req *http.Request

	if r, ok := entry.Data["req"]; ok {
		req, ok = r.(*http.Request)
		if ok {
			// We don't want to log credentials
			req.Header.Del("Authorization")

			entry.Data["req"] = fmt.Sprintf(
				"%s %s %s %s",
				req.Method, req.URL, req.UserAgent(), req.RemoteAddr,
			)
		}
	} else {
		// If there is no request, we build one in order to send
		// all the variables to rollbar
		req = new(http.Request)
		req.Header = make(http.Header)
		req.URL = new(url.URL)
	}

	// All the fields which aren't level|msg|error|time|req are added
	// to the headers of the request which will be sent to Rollbar
	// The main goal is to be able to see all the values on Rollbar dashboard
	for val, key := range entry.Data {
		if val != "level" && val != "msg" && val != "error" && val != "time" && val != "req" {
			req.Header.Add("log-"+val, fmt.Sprintf("%v", key))
		}
	}

	// If there is an error field, we want it to be part of Rollbar ticket name
	var errorMsg error
	var err error
	if entry.Data["error"] != nil {
		err = entry.Data["error"].(error)
		errorTxt := new(bytes.Buffer)
		errorTxt.WriteString(err.Error())
		if msg, ok := entry.Data["msg"]; ok && msg != nil {
			if strMsg, ok := entry.Data["msg"].(string); ok {
				errorTxt.WriteString("- " + strMsg)
			}
		}
		errorMsg = fmt.Errorf(errorTxt.String())
	} else {
		errorMsg = errors.New(entry.Data["msg"].(string))
	}

	rollbar.RequestErrorWithStack(rollbar.ERR, req, errorMsg, errgorollbar.BuildStackWithSkip(err, 5+hook.SkipLevel))
	return nil
}

func (hook Hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}
