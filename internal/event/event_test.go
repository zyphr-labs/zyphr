package event

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func valid() Event {
	return Event{
		SchemaVersion: SchemaVersion,
		ID:            NewID(),
		TS:            time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		SessionID:     "native-session",
		Source:        SourceAdapter,
		Action:        ActionFileRead,
		Stage:         StageIntent,
	}
}

func TestValidate(t *testing.T) {
	e := valid()
	if err := e.Validate(); err != nil {
		t.Fatalf("valid event rejected: %v", err)
	}

	cases := map[string]func(*Event){
		"schema_version": func(e *Event) { e.SchemaVersion = 2 },
		"id":             func(e *Event) { e.ID = "" },
		"ts":             func(e *Event) { e.TS = time.Time{} },
		"run_id":         func(e *Event) { e.SessionID = "" },
		"source":         func(e *Event) { e.Source = "hook" },
		"action":         func(e *Event) { e.Action = "file_open" },
		"stage":          func(e *Event) { e.Stage = "during" },
		"decision":       func(e *Event) { e.Decision = "block" },
		"capture":        func(e *Event) { e.Capture = "full" },
		"native_tool":    func(e *Event) { e.Action = ActionToolUse },
	}
	for want, mutate := range cases {
		e := valid()
		mutate(&e)
		err := e.Validate()
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("mutation %q: got error %v", want, err)
		}
	}
}

func TestRunIDAloneIsEnough(t *testing.T) {
	e := valid()
	e.SessionID, e.RunID = "", "run-1"
	if err := e.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestUnknownStaysOutOfJSON(t *testing.T) {
	b, err := json.Marshal(valid())
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"exit_code", "status", "bytes_out", "decision", "capture", "prompt"} {
		if strings.Contains(string(b), field) {
			t.Errorf("unknown %s must be omitted, got %s", field, b)
		}
	}
}

func TestNewIDIsUnique(t *testing.T) {
	if a, b := NewID(), NewID(); a == b || len(a) != 32 {
		t.Fatalf("bad ids %q %q", a, b)
	}
}
