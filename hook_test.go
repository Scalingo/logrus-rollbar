package logrus_rollbar

import (
	"net/http"
	"sync"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stvp/rollbar"
)

type mockSender struct {
	sync.Mutex
	calls []senderParams
}

type senderParams struct {
	severity string
	req      *http.Request
	error    error
	stack    rollbar.Stack
	fields   []*rollbar.Field
}

func (s *mockSender) RequestErrorWithStack(severity string, req *http.Request, err error, stack rollbar.Stack, fields ...*rollbar.Field) error {
	s.Lock()
	defer s.Unlock()
	s.calls = append(s.calls, senderParams{severity: severity, req: req, error: err, stack: stack, fields: fields})
	return nil
}

func TestHook_Fire(t *testing.T) {
	sender := &mockSender{}
	hook := hook{Sender: sender}
	logger := logrus.New()
	entry := logrus.NewEntry(logger)
	entry.Data["msg"] = "line of log"

	err := hook.Fire(entry)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if len(sender.calls) != 1 {
		t.Errorf("expected 1 call, got %v", len(sender.calls))
	}

	params := sender.calls[0]
	if params.severity != rollbar.ERR {
		t.Errorf("expected severity error, got %v", params.severity)
	}

	if params.error.Error() != "line of log" {
		t.Errorf("expected '%v' error, got '%v'", entry.Data["msg"], params.error.Error())
	}
}

func TestHook_FireWithReq(t *testing.T) {
	sender := &mockSender{}
	hook := hook{Sender: sender}
	logger := logrus.New()
	entry := logrus.NewEntry(logger)
	entry.Data["msg"] = "line of log"

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "password")
	entry.Data["req"] = req

	err := hook.Fire(entry)
	if err != nil {
		t.Errorf("expected nil error, got '%v'", err)
	}

	params := sender.calls[0]
	if params.req.Header.Get("Authorization") != "" {
		t.Errorf("expected Authorization header to be cleared, got", params.req.Header.Get("Authorization"))
	}
}

func TestHook_WithExtraField(t *testing.T) {
	sender := &mockSender{}
	hook := hook{Sender: sender}
	logger := logrus.New()
	entry := logrus.NewEntry(logger)
	entry.Data["msg"] = "line of log"
	entry.Data["extra"] = "extrafield"

	err := hook.Fire(entry)
	if err != nil {
		t.Errorf("expected nil error, got '%v'", err)
	}

	params := sender.calls[0]
	if len(params.fields) != 1 {
		t.Errorf("expected 1 extra field, got %v", len(params.fields))
	}
	extra, ok := params.fields[0].Data.(string)
	if !ok {
		t.Errorf("expected extra field string, got %t", params.fields[0].Data)
	}
	if extra != "extrafield" {
		t.Errorf("expected 'extrafield' value for extra field, got %v", extra)
	}
}

// Test when a request object is used multiple times in error handling (example multiple goroutine)
// This test case aims at testing race conditions
func TestHook_FireWithReq_Concurrent(t *testing.T) {
	sender := &mockSender{}
	hook := hook{Sender: sender}
	logger := logrus.New()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "password")

	// Entries are only created and sent once, it is not safe to send twice the
	// same entry through the hook
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		entry := logrus.NewEntry(logger)
		entry.Data["msg"] = "line of log"
		entry.Data["req"] = req

		hook.Fire(entry)
		wg.Done()
	}()
	go func() {
		entry := logrus.NewEntry(logger)
		entry.Data["msg"] = "line of log"
		entry.Data["req"] = req

		hook.Fire(entry)
		wg.Done()
	}()
	wg.Wait()

	if len(sender.calls) != 2 {
		t.Errorf("expected 2 call, got %v", len(sender.calls))
	}
}
