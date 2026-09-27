package escalation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"time"
)

// Card audit log entry kinds (design.md §C.6). The log is authoritative for
// arming: a card is armed exactly while its most recent armed entry has no
// later disarmed entry.
const (
	LineArmed      = "armed"
	LineState      = "state"
	LineDisarmed   = "disarmed"
	LineNotChecked = "not-checked"
	// LineNotObserved lists detections that could not be completed
	// (REQ-AE-022): at a write whose judgment needs an undetermined root or
	// an unreadable field, and at every commit checkpoint.
	LineNotObserved = "not-observed"
	// LineNotArmed and LineWarning are declared with the resolver.
)

// LogEntry is one line of a card audit log.
type LogEntry struct {
	Kind string `json:"kind"`
	Time string `json:"time"`
	Card string `json:"card"`
	// armed
	Spec           string `json:"spec,omitempty"`
	Contract       string `json:"contract,omitempty"`
	ContractSHA256 string `json:"contract_sha256,omitempty"`
	// state
	StateSHA256 string `json:"state_sha256,omitempty"`
	// disarmed
	Reason      string `json:"reason,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	// Accounted lists the evidence of arming a state-tamper judgment
	// consumed, so the same evidence is not judged again.
	Accounted []string `json:"accounted,omitempty"`
	// not-armed, not-checked, warning
	Class   string   `json:"class,omitempty"`
	Cause   string   `json:"cause,omitempty"`
	Detail  string   `json:"detail,omitempty"`
	Specs   []string `json:"specs,omitempty"`
	Reasons []string `json:"reasons,omitempty"`
	// not-observed
	NotObserved []string `json:"not_observed,omitempty"`
	// Prev is the SHA-256 of the previous line of the same file ("" first).
	Prev string `json:"prev"`
}

// CardLog is a read card audit log.
type CardLog struct {
	Entries []LogEntry
	// ChainIntact is false when a line does not parse or a prev does not
	// match its predecessor. An absent or empty log is intact.
	ChainIntact bool
	// lastLine is the raw last line, the input of the next prev.
	lastLine []byte
	// lines counts the non-empty lines read or appended.
	lines int
	// lastBreak is the index of the last line that broke the chain, -1 when
	// none; lastTamper is the line index of the last state-tamper disarmed
	// entry, -1 when none. A break before lastTamper has been judged.
	lastBreak, lastTamper int
	// missingNewline is true when the file does not end in a newline.
	missingNewline bool
}

// Armed reports whether the log shows the card armed: its most recent armed
// entry has no later disarmed entry.
func (l CardLog) Armed() bool {
	armed := false
	for _, e := range l.Entries {
		switch e.Kind {
		case LineArmed:
			armed = true
		case LineDisarmed:
			armed = false
		}
	}
	return armed
}

// unaccountedBreak reports a chain break no later state-tamper disarmed entry
// has judged, and the line it is on.
func (l CardLog) unaccountedBreak() (int, bool) {
	return l.lastBreak, l.lastBreak >= 0 && l.lastBreak > l.lastTamper
}

// latest returns the latest entry of kind.
func (l CardLog) latest(kind string) (LogEntry, bool) {
	for i := len(l.Entries) - 1; i >= 0; i-- {
		if l.Entries[i].Kind == kind {
			return l.Entries[i], true
		}
	}
	return LogEntry{}, false
}

// accounts reports whether the log accounts for an evidence id: an armed
// entry for a contract digest ("sha:<digest>"), or a disarmed entry carrying
// the fingerprint ("fp:<fp>") or listing the id as accounted.
func (l CardLog) accounts(id string) bool {
	for _, e := range l.Entries {
		switch {
		case e.Kind == LineArmed && "sha:"+e.ContractSHA256 == id:
			return true
		case e.Kind == LineDisarmed && ("fp:"+e.Fingerprint == id || slices.Contains(e.Accounted, id)):
			return true
		}
	}
	return false
}

// SHA256Hex returns the lowercase-hex SHA-256 of b.
func SHA256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ReadCardLog reads and chain-checks a card audit log. An absent log reads as
// an empty, intact log. Only an I/O error is returned as an error; a broken
// chain is reported through ChainIntact.
func ReadCardLog(path string) (CardLog, error) {
	lg := CardLog{ChainIntact: true, lastBreak: -1, lastTamper: -1}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return lg, nil
	}
	if err != nil {
		return CardLog{}, fmt.Errorf("escalation: read card log: %w", err)
	}
	lg.missingNewline = len(data) > 0 && data[len(data)-1] != '\n'
	prev := ""
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var e LogEntry
		if err := json.Unmarshal(line, &e); err != nil {
			lg.ChainIntact = false
			lg.lastBreak = lg.lines
		} else {
			if e.Prev != prev {
				lg.ChainIntact = false
				lg.lastBreak = lg.lines
			}
			lg.addEntry(e)
		}
		prev = SHA256Hex(line)
		lg.lastLine = append([]byte(nil), line...)
		lg.lines++
	}
	return lg, nil
}

// addEntry records a parsed entry at the current line index.
func (l *CardLog) addEntry(e LogEntry) {
	l.Entries = append(l.Entries, e)
	if e.Kind == LineDisarmed && e.Reason == DisarmStateTamper {
		l.lastTamper = l.lines
	}
}

// AppendLog appends one entry to the card log at path, chaining it to the
// file's current last line. Callers hold the card lock (design.md §C.6).
func AppendLog(path string, e LogEntry, now time.Time) error {
	lg, err := ReadCardLog(path)
	if err != nil {
		return err
	}
	return appendAfter(path, &lg, e, now)
}

// appendAfter appends e after the already-read log lg and updates lg.
func appendAfter(path string, lg *CardLog, e LogEntry, now time.Time) error {
	if e.Time == "" {
		e.Time = now.UTC().Format(time.RFC3339)
	}
	e.Prev = ""
	if lg.lastLine != nil {
		e.Prev = SHA256Hex(lg.lastLine)
	}
	line, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("escalation: encode log entry: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("escalation: create store: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("escalation: open card log: %w", err)
	}
	out := append(line, '\n')
	if lg.missingNewline {
		out = append([]byte{'\n'}, out...)
		lg.missingNewline = false
	}
	if _, err := f.Write(out); err != nil {
		_ = f.Close()
		return fmt.Errorf("escalation: append card log: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("escalation: close card log: %w", err)
	}
	lg.addEntry(e)
	lg.lastLine = line
	lg.lines++
	return nil
}
