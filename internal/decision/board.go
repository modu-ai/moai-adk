// Package decision implements the decision board: one append-only,
// one-record-per-line store per project under the moai home state directory.
// The leader records rulings there; lanes read them instead of waiting on a
// chat reply.
package decision

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Scope and kind vocabularies. Both are closed: a value outside them is refused.
const (
	ScopeStanding    = "standing"
	scopeCardPrefix  = "card:"
	KindRuling       = "ruling"
	KindStandingRule = "standing-rule"
	KindWaitResolve  = "wait-resolution"
	KindHold         = "hold"
	KindSplitAck     = "split-proposal-ack"
	KindCeilingExcpt = "ceiling-exception"
	KindReleaseScope = "release-scope"
	KindSupersede    = "supersede"
)

var kinds = map[string]bool{
	KindRuling: true, KindStandingRule: true, KindWaitResolve: true, KindHold: true,
	KindSplitAck: true, KindCeilingExcpt: true, KindReleaseScope: true, KindSupersede: true,
}

// Read statuses reported instead of an empty success.
const (
	StatusAbsent = "absent"
	StatusEmpty  = "empty"
	StatusOK     = "ok"
)

// Record is one board line.
type Record struct {
	ID           string   `json:"id"`
	Scope        string   `json:"scope"`
	Kind         string   `json:"kind"`
	DecidedBy    string   `json:"decided_by"`
	EvidenceRefs string   `json:"evidence_refs"`
	LadderPath   string   `json:"ladder_path"`
	Body         string   `json:"body"`
	Predicate    string   `json:"predicate,omitempty"`
	Created      string   `json:"created"`
	Supersedes   string   `json:"supersedes,omitempty"`
	Resolves     string   `json:"resolves,omitempty"`
	Release      string   `json:"release,omitempty"`
	Cards        []string `json:"cards,omitempty"`
}

// Line renders the record in the one-line decision-record form so existing
// greps over decision records keep working.
func (r Record) Line() string {
	return fmt.Sprintf("decision record: decided_by=%s evidence_refs=%s ladder_path=%s id=%s scope=%s kind=%s",
		r.DecidedBy, r.EvidenceRefs, r.LadderPath, r.ID, r.Scope, r.Kind)
}

// AppendOptions carries the inputs Append cannot derive itself.
type AppendOptions struct {
	// Now stamps the record; zero means time.Now().
	Now time.Time
	// WaitFile is the progress record that holds the wait a Resolves names.
	// A record that resolves a wait is refused without it, because the board
	// cannot otherwise tell a real wait id from a typo.
	WaitFile string
}

// ReadOptions selects records.
type ReadOptions struct {
	// Scope "card:<id>" returns that card's and every standing record;
	// empty returns everything.
	Scope string
	// All includes superseded records.
	All bool
}

// ReadResult is the outcome of Read, with explicit statuses.
type ReadResult struct {
	Status      string
	Unparseable int
	Records     []Record
}

// StatusLine renders the read status for humans and greps.
func (r ReadResult) StatusLine() string {
	return fmt.Sprintf("board=%s unparseable=%d records=%d", r.Status, r.Unparseable, len(r.Records))
}

// BoardPath is the project's board file under the moai home state directory,
// keyed by the primary checkout, so every linked worktree reads one board.
func BoardPath(projectRoot string) (string, error) {
	dir, err := homestate.ProjectDir(projectRoot)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "decisions", "board.jsonl"), nil
}

var waitIDPattern = regexp.MustCompile(`^w-[A-Za-z0-9._:-]+$`)

func validate(r Record) error {
	switch {
	case r.Scope == ScopeStanding:
	case strings.HasPrefix(r.Scope, scopeCardPrefix) && strings.TrimSpace(strings.TrimPrefix(r.Scope, scopeCardPrefix)) != "":
	default:
		return fmt.Errorf("scope %q: want card:<card-id> or standing", r.Scope)
	}
	if !kinds[r.Kind] {
		return fmt.Errorf("kind %q is outside the closed enumeration", r.Kind)
	}
	for name, v := range map[string]string{"decided_by": r.DecidedBy, "evidence_refs": r.EvidenceRefs, "ladder_path": r.LadderPath, "body": r.Body} {
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if r.Kind == KindReleaseScope {
		if r.Scope != ScopeStanding {
			return errors.New("a release-scope record must be standing")
		}
		if strings.TrimSpace(r.Release) == "" || len(r.Cards) == 0 {
			return errors.New("a release-scope record requires a release id and a card-id list")
		}
	} else if r.Scope == ScopeStanding && strings.TrimSpace(r.Predicate) == "" {
		return errors.New("a standing record requires the predicate that selects the situations it governs")
	}
	if r.Resolves != "" && !waitIDPattern.MatchString(r.Resolves) {
		return fmt.Errorf("resolves %q is not a wait id (w-...)", r.Resolves)
	}
	return nil
}

func newID(now time.Time) (string, error) {
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "d-" + now.UTC().Format("20060102T150405Z") + "-" + hex.EncodeToString(b[:]), nil
}

// waitFileHolds reports whether the progress record carries a wait line
// whose id equals waitID.
func waitFileHolds(path, waitID string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	needle := "wait record: id=" + waitID
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == needle || strings.HasPrefix(line, needle+" ") {
			return true, nil
		}
	}
	return false, nil
}

// Append validates r and appends it as one line under an exclusive lock.
// A refused record leaves the board untouched.
//
// @MX:ANCHOR: [AUTO] the only writer of the decision board; the CLI and any future MCP surface route here
// @MX:REASON: append-only and reference validation (supersedes, resolves) are enforced at this one point; a second writer would bypass them
func Append(path string, r Record, opts AppendOptions) (Record, error) {
	if err := validate(r); err != nil {
		return Record{}, fmt.Errorf("decision record: %w", err)
	}
	if r.Resolves != "" {
		if opts.WaitFile == "" {
			return Record{}, errors.New("decision record: resolves requires the progress record holding the wait (wait file)")
		}
		ok, err := waitFileHolds(opts.WaitFile, r.Resolves)
		if err != nil {
			return Record{}, fmt.Errorf("decision record: wait file: %w", err)
		}
		if !ok {
			return Record{}, fmt.Errorf("decision record: wait id %q is not in %s", r.Resolves, opts.WaitFile)
		}
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return Record{}, err
	}
	unlock, err := lockBoard(path + ".lock")
	if err != nil {
		return Record{}, fmt.Errorf("decision board lock: %w", err)
	}
	defer unlock()
	if r.Supersedes != "" {
		existing, _, err := load(path)
		if err != nil {
			return Record{}, err
		}
		found := false
		for _, e := range existing {
			if e.ID == r.Supersedes {
				found = true
				break
			}
		}
		if !found {
			return Record{}, fmt.Errorf("decision record: supersedes %q names no record on the board", r.Supersedes)
		}
	}
	id, err := newID(now)
	if err != nil {
		return Record{}, err
	}
	r.ID = id
	r.Created = now.UTC().Format(time.RFC3339)
	line, err := json.Marshal(r)
	if err != nil {
		return Record{}, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return Record{}, err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		_ = f.Close()
		return Record{}, err
	}
	return r, f.Close()
}

// load parses every line; unparseable non-blank lines are counted, never fatal.
func load(path string) ([]Record, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	var out []Record
	bad := 0
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var r Record
		if err := json.Unmarshal([]byte(line), &r); err != nil || r.ID == "" {
			bad++
			continue
		}
		out = append(out, r)
	}
	return out, bad, sc.Err()
}

// Read returns the records for the requested scope, newest first, with an
// explicit status: absent, empty, or ok plus the count of unparseable lines.
func Read(path string, opts ReadOptions) (ReadResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ReadResult{Status: StatusAbsent}, nil
		}
		return ReadResult{}, err
	}
	if info.Size() == 0 {
		return ReadResult{Status: StatusEmpty}, nil
	}
	all, bad, err := load(path)
	if err != nil {
		return ReadResult{}, err
	}
	superseded := map[string]bool{}
	for _, r := range all {
		if r.Supersedes != "" {
			superseded[r.Supersedes] = true
		}
	}
	var out []Record
	for _, r := range all {
		if opts.Scope != "" && r.Scope != opts.Scope && r.Scope != ScopeStanding {
			continue
		}
		if !opts.All && superseded[r.ID] {
			continue
		}
		out = append(out, r)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Created > out[j].Created })
	return ReadResult{Status: StatusOK, Unparseable: bad, Records: out}, nil
}
