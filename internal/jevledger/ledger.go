// Package jevledger records one JSON Lines row per Jev question to the project's
// local decision ledger, .moai/state/jev/decisions.jsonl (SPEC-JEV-LEDGER-001,
// card t1762, REQ-JEVL-009..012, 018).
//
// The recorder is deliberately small and fail-open. It is constructed from a
// project root and does no I/O until Append is called; Append never panics,
// never prints, never blocks, and reports a dropped row only as a returned
// value. It never opens the ledger for reading, so no code path in this
// package can feed a row back into anything it names.
//
// This package imports the standard library only. It sends no request and
// reads no credential; it does not import internal/jev.
package jevledger

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// dirMode and fileMode are the POSIX modes of the ledger directory and
	// file (REQ-JEVL-011).
	dirMode  os.FileMode = 0o700
	fileMode os.FileMode = 0o600

	// modeShadow is the mode whose rows always carry action "none".
	modeShadow = "shadow"
	// actionNone and outcomeUnlabeled are the two values the recorder owns.
	actionNone       = "none"
	outcomeUnlabeled = "unlabeled"
)

// errNotRegular is the one refusal for an opened handle that is not a regular
// file: a FIFO with or without a reader, a device, a directory.
var errNotRegular = errors.New("ledger path is not a regular file")

// Row is one ledger line. The caller supplies every field except OutcomeLabel
// and, under shadow, Action: the recorder owns those two (REQ-JEVL-010) and
// overwrites whatever the caller put there.
//
// Probability and Threshold are pointers WITHOUT omitempty so an absent value
// is written as null and a supplied 0 stays the number 0.
type Row struct {
	TS             string   `json:"ts"`
	Consumer       string   `json:"consumer"`
	Event          string   `json:"event"`
	Harness        string   `json:"harness"`
	CallID         string   `json:"call_id"`
	QuestionID     string   `json:"question_id"`
	RubricVersion  string   `json:"rubric_version"`
	ModelID        string   `json:"model_id"`      // the response's interpreted model id; empty when unanswered
	RequestModel   string   `json:"request_model"` // the request alias; never copied into ModelID
	Availability   string   `json:"availability"`
	Probability    *float64 `json:"probability"`
	Choice         string   `json:"choice,omitempty"`
	Score          *float64 `json:"score,omitempty"`
	InputTokens    int      `json:"input_tokens"`
	Mode           string   `json:"mode"`            // the effective mode
	ConfiguredMode string   `json:"configured_mode"` // the mode as configured
	Threshold      *float64 `json:"threshold"`
	ConfigWarning  string   `json:"config_warning,omitempty"`
	Action         string   `json:"action"`
	OutcomeLabel   string   `json:"outcome_label"`
}

// Result names whether a row landed or was dropped, and why. It is the only
// report of a drop: the recorder writes nothing to stdout or stderr.
type Result struct {
	Landed bool
	// Reason is empty when the row landed, otherwise the cause of the drop.
	Reason string
}

// Recorder appends rows to the ledger under one project root.
type Recorder struct {
	root string
}

// New returns a recorder for projectRoot. It performs no I/O: nothing is
// created under the root until Append runs (REQ-JEVL-001).
func New(projectRoot string) *Recorder {
	return &Recorder{root: projectRoot}
}

// Append writes row as one complete line at the end of the ledger.
//
// The row is dropped, with the cause returned, when the directory or file
// cannot be created or written, or when the opened handle is not a regular
// file. A drop leaves the object at the ledger path and any link target
// unchanged and delivers no byte to it.
//
// @MX:ANCHOR: [AUTO] Append is the single entry point through which every Jev decision row reaches disk; the -002 hook adapter, the tests and the -003 report fixtures all depend on its fail-open contract
// @MX:REASON: A change here (blocking, printing, panicking, or reading the ledger) breaks the hook-budget and display-only guarantees of SPEC-JEV-LEDGER-001/-002
// @MX:SPEC: SPEC-JEV-LEDGER-001
func (r *Recorder) Append(row Row) Result {
	row.OutcomeLabel = outcomeUnlabeled
	if row.Mode == modeShadow || row.Action == "" {
		row.Action = actionNone
	}

	line, err := json.Marshal(row)
	if err != nil {
		return dropped("encode row", err)
	}
	line = append(line, '\n')

	path := r.path()
	if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		return dropped("create ledger directory", err)
	}
	f, err := openLedger(path)
	if err != nil {
		return dropped("open ledger", err)
	}
	// One Write on an O_APPEND handle: concurrent writers never interleave a
	// row on a local filesystem (plan.md section H states the limit).
	_, werr := f.Write(line)
	cerr := f.Close()
	if werr != nil {
		return dropped("write ledger", werr)
	}
	if cerr != nil {
		return dropped("close ledger", cerr)
	}
	return Result{Landed: true}
}

// path is the ledger file under the recorder's project root.
func (r *Recorder) path() string {
	return filepath.Join(r.root, ".moai", "state", "jev", "decisions.jsonl")
}

func dropped(step string, err error) Result {
	return Result{Reason: fmt.Sprintf("%s: %v", step, err)}
}
