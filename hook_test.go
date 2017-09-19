package logrus_rollbar

import (
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
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

func (s *mockSender) ErrorWithStack(severity string, err error, stack rollbar.Stack, fields ...*rollbar.Field) error {
	s.Lock()
	defer s.Unlock()
	s.calls = append(s.calls, senderParams{severity: severity, error: err, stack: stack, fields: fields})
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

// Test that the stack is correctly sent when an errors from github.com/pkg/errors is sent
func TestHook_WithPkgErrors(t *testing.T) {
	sender := &mockSender{}
	hook := hook{Sender: sender}
	logger := logrus.New()
	entry := logrus.NewEntry(logger)
	entry.Data["error"] = errorsFoo()

	err := hook.Fire(entry)
	if err != nil {
		t.Errorf("expected nil error, got '%v'", err)
	}

	if len(sender.calls) != 1 {
		t.Errorf("expected 1 call, got %v", len(sender.calls))
	}

	params := sender.calls[0]
	// github.com/Scalingo/logrus-rollbar.errorsBar
	// 	/home/leo/Projects/Go/src/github.com/Scalingo/logrus-rollbar/hook_pkgerrors_test.go:12
	// github.com/Scalingo/logrus-rollbar.errorsFoo
	// 	/home/leo/Projects/Go/src/github.com/Scalingo/logrus-rollbar/hook_pkgerrors_test.go:8
	// github.com/Scalingo/logrus-rollbar.TestHook_WithPkgErrors
	// 	/home/leo/Projects/Go/src/github.com/Scalingo/logrus-rollbar/hook_test.go:155
	// testing.tRunner
	// 	/opt/go/src/testing/testing.go:657
	// runtime.goexit
	// 	/opt/go/src/runtime/asm_amd64.s:2197
	if len(params.stack) != 5 {
		t.Errorf("expecting a stack of 5 levels, got %v: \n===\n%v", len(params.stack), params.stack)
	}

	stack := rollbar.Stack{
		{
			Method:   "goexit",
			Filename: "asm_amd64.s",
		}, {
			Method:   "tRunner",
			Filename: "testing.go",
		}, {
			Line:     156,
			Method:   "TestHook_WithPkgErrors",
			Filename: "hook_test.go",
		}, {
			Line:     8,
			Method:   "errorsFoo",
			Filename: "hook_pkgerrors_test.go",
		}, {
			Line:     12,
			Method:   "errorsBar",
			Filename: "hook_pkgerrors_test.go",
		},
	}

	for i, _ := range params.stack {
		if params.stack[i].Method != stack[i].Method {
			t.Errorf("expected method %v, got %v", stack[i].Method, params.stack[i].Method)
		}
		if !strings.HasSuffix(params.stack[i].Filename, stack[i].Filename) {
			t.Errorf("expected filename %v, got %v", stack[i].Filename, params.stack[i].Filename)
		}
		if stack[i].Line != 0 && stack[i].Line != params.stack[i].Line {
			t.Errorf("expected line %v, got %v", stack[i].Line, params.stack[i].Line)
		}
	}
}
