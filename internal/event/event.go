// Package event defines the normalized Zyphr event. It is the only shape that
// crosses from a harness adapter or the proxy into the core.
package event

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// SchemaVersion is the current event schema version.
const SchemaVersion = 1

// Source says which part of Zyphr produced an event.
type Source string

const (
	SourceAdapter Source = "adapter" // a harness adapter, from a native hook
	SourceProxy   Source = "proxy"   // the scoped egress proxy
	SourceGap     Source = "gap"     // a declared hole in visibility
	SourceSystem  Source = "system"  // Zyphr itself, such as run start and end
)

// Action is what happened, in harness-independent terms.
type Action string

const (
	ActionSessionStart  Action = "session_start"
	ActionSessionEnd    Action = "session_end"
	ActionPrompt        Action = "prompt"
	ActionResponse      Action = "response"
	ActionFileRead      Action = "file_read"
	ActionFileWrite     Action = "file_write"
	ActionFileDelete    Action = "file_delete"
	ActionCommandExec   Action = "command_exec"
	ActionMCPCall       Action = "mcp_call"
	ActionWebFetch      Action = "web_fetch"
	ActionWebSearch     Action = "web_search"
	ActionSubagentStart Action = "subagent_start"
	ActionSubagentStop  Action = "subagent_stop"
	// ActionToolUse is a native tool that the adapter could not map with
	// certainty. It is the explicit "unmapped" marker. NativeTool names it.
	ActionToolUse    Action = "tool_use"
	ActionNetRequest Action = "net_request"
	ActionGap        Action = "gap"
)

// Stage separates what an agent asked for from what was reported afterwards.
type Stage string

const (
	StageIntent Stage = "intent" // before the action; not proof that it ran
	StageResult Stage = "result" // reported after the action
)

// Decision is the outcome of policy evaluation for an event.
type Decision string

const (
	DecisionObserve Decision = "observe" // recorded only; the default
	DecisionAllow   Decision = "allow"   // a rule matched and allowed it
	DecisionDeny    Decision = "deny"    // a rule matched and denied it
)

// Capture states how much of the underlying activity this event holds.
type Capture string

const (
	CaptureComplete    Capture = "complete"
	CapturePartial     Capture = "partial"
	CaptureDisabled    Capture = "disabled"
	CaptureUnavailable Capture = "unavailable"
)

// Event is one normalized record. Fields that do not apply stay empty and are
// omitted from JSON. A nil pointer means unknown, never zero.
type Event struct {
	// Identity
	SchemaVersion int       `json:"schema_version"`
	ID            string    `json:"id"`
	TS            time.Time `json:"ts"`
	RunID         string    `json:"run_id,omitempty"`
	SessionID     string    `json:"session_id,omitempty"`

	// Origin
	Harness        string `json:"harness,omitempty"`
	HarnessVersion string `json:"harness_version,omitempty"`
	Source         Source `json:"source"`
	NativeHook     string `json:"native_hook,omitempty"`
	NativeTool     string `json:"native_tool,omitempty"`

	// Action
	Action Action `json:"action"`
	Stage  Stage  `json:"stage,omitempty"`

	// Tool detail
	Target    string `json:"target,omitempty"`
	Command   string `json:"command,omitempty"`
	Cwd       string `json:"cwd,omitempty"`
	MCPServer string `json:"mcp_server,omitempty"`
	MCPTool   string `json:"mcp_tool,omitempty"`
	ExitCode  *int   `json:"exit_code,omitempty"`

	// Network detail
	Method   string `json:"method,omitempty"`
	Host     string `json:"host,omitempty"`
	URL      string `json:"url,omitempty"`
	Status   *int   `json:"status,omitempty"`
	BytesOut *int64 `json:"bytes_out,omitempty"`
	BytesIn  *int64 `json:"bytes_in,omitempty"`

	// Decision
	Decision Decision `json:"decision,omitempty"`
	RuleID   string   `json:"rule_id,omitempty"`

	// Evidence
	Capture Capture `json:"capture,omitempty"`
	Note    string  `json:"note,omitempty"`

	// Opt-in text. The collector drops these unless capture is enabled.
	Prompt   string `json:"prompt,omitempty"`
	Response string `json:"response,omitempty"`
}

// NewID returns a random event or run identifier.
func NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("event: no randomness: " + err.Error())
	}
	return hex.EncodeToString(b[:])
}

// Validate reports whether e is a well-formed schema v1 event.
func (e *Event) Validate() error {
	if e.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %d", e.SchemaVersion)
	}
	if e.ID == "" {
		return errors.New("id is required")
	}
	if e.TS.IsZero() {
		return errors.New("ts is required")
	}
	if e.RunID == "" && e.SessionID == "" {
		return errors.New("run_id or session_id is required")
	}
	if !validSource[e.Source] {
		return fmt.Errorf("unknown source %q", e.Source)
	}
	if !validAction[e.Action] {
		return fmt.Errorf("unknown action %q", e.Action)
	}
	if e.Stage != "" && e.Stage != StageIntent && e.Stage != StageResult {
		return fmt.Errorf("unknown stage %q", e.Stage)
	}
	if e.Decision != "" && !validDecision[e.Decision] {
		return fmt.Errorf("unknown decision %q", e.Decision)
	}
	if e.Capture != "" && !validCapture[e.Capture] {
		return fmt.Errorf("unknown capture %q", e.Capture)
	}
	if e.Action == ActionToolUse && e.NativeTool == "" {
		return errors.New("tool_use needs native_tool")
	}
	return nil
}

// Actions lists every valid action in a stable order.
func Actions() []Action {
	return []Action{
		ActionSessionStart, ActionSessionEnd, ActionPrompt, ActionResponse,
		ActionFileRead, ActionFileWrite, ActionFileDelete, ActionCommandExec,
		ActionMCPCall, ActionWebFetch, ActionWebSearch, ActionSubagentStart,
		ActionSubagentStop, ActionToolUse, ActionNetRequest, ActionGap,
	}
}

var validSource = map[Source]bool{
	SourceAdapter: true, SourceProxy: true, SourceGap: true, SourceSystem: true,
}

var validAction = func() map[Action]bool {
	m := map[Action]bool{}
	for _, a := range Actions() {
		m[a] = true
	}
	return m
}()

var validDecision = map[Decision]bool{
	DecisionObserve: true, DecisionAllow: true, DecisionDeny: true,
}

var validCapture = map[Capture]bool{
	CaptureComplete: true, CapturePartial: true, CaptureDisabled: true, CaptureUnavailable: true,
}
