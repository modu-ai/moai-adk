package hygiene

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Signal is the three-valued per-signal state (REQ-HYG-007): affirmative,
// negatively measured, or unmeasured. An unmeasured signal never feeds a
// DEAD verdict.
type Signal int

const (
	SignalUnmeasured Signal = iota
	SignalAffirmative
	SignalNegative
)

// String renders the signal for audit evidence.
func (s Signal) String() string {
	switch s {
	case SignalAffirmative:
		return "affirmative"
	case SignalNegative:
		return "negative"
	default:
		return "unmeasured"
	}
}

// Verdict is the session-liveness classification.
type Verdict int

const (
	VerdictIndeterminate Verdict = iota
	VerdictLive
	VerdictDead
)

// String renders the verdict for audit rows and reports.
func (v Verdict) String() string {
	switch v {
	case VerdictLive:
		return "LIVE"
	case VerdictDead:
		return "DEAD"
	default:
		return "INDETERMINATE"
	}
}

// RegistryEntry is the minimal registry row the liveness evaluator reads
// (the registry JSON schema is frozen upstream; only these fields are
// consumed here).
type RegistryEntry struct {
	SessionID     string    `json:"session_id"`
	PID           int       `json:"pid"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

// Liveness classifies a candidate session's liveness from three independent
// signal families (REQ-HYG-007): the registry pid probe (with its process-
// start-time fingerprint check), transcript activity under the scanned
// profile roots, and the registry heartbeat recency. Any affirmative
// signal ⇒ LIVE; every signal measured and negative ⇒ DEAD; any unmeasured
// signal with none affirmative ⇒ INDETERMINATE (kept, reason reported).
//
// In this tree the registry schema records no per-entry process
// fingerprint, so a live pid reads unmeasured in production — an
// affirmative pid signal requires a fingerprint comparison the install
// does not record yet (fail-closed by design); the match/mismatch arms
// remain reachable through the seam for the verdict matrix and for a
// future writer that records fingerprints.
type Liveness struct {
	// TranscriptWindow bounds transcript recency (production default from
	// internal/config; never a call-site literal).
	TranscriptWindow time.Duration
	// HeartbeatWindow bounds registry heartbeat recency.
	HeartbeatWindow time.Duration
	// TranscriptRoots are the profile roots scanned for transcripts.
	TranscriptRoots []string
	// RegistryPath is the active-sessions registry file path.
	RegistryPath string

	now func() time.Time

	// registryPresent lets tests model an absent vs present registry file.
	registryPresent bool
	// registryEntry returns the entry for a key (nil = no entry).
	registryEntry func(key string) *RegistryEntry
	// pidProbe classifies the registry pid signal.
	pidProbe func(pid int) Signal
	// transcriptProbe classifies the transcript signal for a key.
	transcriptProbe func(key string) Signal
	// heartbeatProbe classifies the heartbeat signal for a key.
	heartbeatProbe func(key string) Signal
}

// lnow returns the evaluator's clock.
func (l *Liveness) lnow() time.Time {
	if l.now != nil {
		return l.now()
	}
	return time.Now()
}

// Evaluate computes the verdict for one session key and returns the
// verdict, the names of the unmeasured signals, and per-signal evidence
// for the audit row (REQ-HYG-009's signal-evidence requirement).
//
// @MX:ANCHOR: [AUTO] Liveness.Evaluate — the three-signal fail-closed verdict
// @MX:REASON: every apply-mode deletion is gated on this verdict; a DEAD
// misclassification deletes live session state, so the unmeasured-never-
// DEAD rule is the load-bearing safety property (REQ-HYG-007/008).
func (l *Liveness) Evaluate(key string) (Verdict, []string, map[string]string) {
	evidence := map[string]string{}
	unmeasured := []string{}

	pid := l.pidSignal(key, evidence)
	if pid == SignalUnmeasured {
		unmeasured = append(unmeasured, "pid")
	}
	transcript := l.transcriptSignal(key, evidence)
	if transcript == SignalUnmeasured {
		unmeasured = append(unmeasured, "transcript")
	}
	heartbeat := l.heartbeatSignal(key, evidence)
	if heartbeat == SignalUnmeasured {
		unmeasured = append(unmeasured, "heartbeat")
	}

	switch {
	case pid == SignalAffirmative || transcript == SignalAffirmative || heartbeat == SignalAffirmative:
		return VerdictLive, unmeasured, evidence
	case pid == SignalNegative && transcript == SignalNegative && heartbeat == SignalNegative:
		return VerdictDead, unmeasured, evidence
	default:
		return VerdictIndeterminate, unmeasured, evidence
	}
}

// pidSignal resolves the registry entry and classifies the pid signal.
// A registry-absent or entry-missing candidate — the majority case — has
// an unmeasured pid signal by construction.
func (l *Liveness) pidSignal(key string, evidence map[string]string) Signal {
	if l.pidProbe != nil {
		s := l.pidProbe(registryPIDFor(l, key))
		evidence["pid"] = s.String()
		return s
	}
	entry := l.lookupEntry(key)
	if entry == nil {
		evidence["pid"] = "unmeasured"
		return SignalUnmeasured
	}
	s := probePidAlive(entry.PID)
	evidence["pid"] = s.String()
	return s
}

// transcriptSignal classifies transcript activity for a key.
func (l *Liveness) transcriptSignal(key string, evidence map[string]string) Signal {
	if l.transcriptProbe != nil {
		s := l.transcriptProbe(key)
		evidence["transcript"] = s.String()
		return s
	}
	s := l.scanTranscripts(key)
	evidence["transcript"] = s.String()
	return s
}

// heartbeatSignal classifies registry heartbeat recency.
func (l *Liveness) heartbeatSignal(key string, evidence map[string]string) Signal {
	if l.heartbeatProbe != nil {
		s := l.heartbeatProbe(key)
		evidence["heartbeat"] = s.String()
		return s
	}
	entry := l.lookupEntry(key)
	if entry == nil {
		evidence["heartbeat"] = "unmeasured"
		return SignalUnmeasured
	}
	window := l.HeartbeatWindow
	if window <= 0 {
		evidence["heartbeat"] = "unmeasured"
		return SignalUnmeasured
	}
	age := l.lnow().Sub(entry.LastHeartbeat)
	if age < 0 {
		// A heartbeat stamped in the future is treated as fresh (clock skew
		// errs toward keeping).
		evidence["heartbeat"] = "affirmative"
		return SignalAffirmative
	}
	if age <= window {
		evidence["heartbeat"] = "affirmative"
		return SignalAffirmative
	}
	evidence["heartbeat"] = "negative"
	return SignalNegative
}

// lookupEntry resolves the registry entry for a key, loading the registry
// once per evaluation. A missing or unparseable registry file reads as
// registry-absent (both signals unmeasured — fail-closed).
func (l *Liveness) lookupEntry(key string) *RegistryEntry {
	if l.registryEntry != nil {
		return l.registryEntry(key)
	}
	if l.RegistryPath == "" {
		return nil
	}
	entries, err := readRegistryEntries(l.RegistryPath)
	if err != nil {
		return nil
	}
	l.registryPresent = true
	for i := range entries {
		if entries[i].SessionID == key {
			entry := entries[i]
			return &entry
		}
	}
	return nil
}

// registryPIDFor resolves the pid for a key through the seam path.
func registryPIDFor(l *Liveness, key string) int {
	if entry := l.lookupEntry(key); entry != nil {
		return entry.PID
	}
	return 0
}

// scanTranscripts looks for a transcript of the session key under the
// configured profile roots: found with fresh mtime ⇒ affirmative, found
// stale ⇒ negative, and — the REQ-HYG-007 rule — absent under resolvable
// roots ⇒ unmeasured, never negative (the session may live under a root
// this install does not scan).
func (l *Liveness) scanTranscripts(key string) Signal {
	window := l.TranscriptWindow
	if window <= 0 || len(l.TranscriptRoots) == 0 || key == "" {
		return SignalUnmeasured
	}
	now := l.lnow()
	resolvable := false
	for _, root := range l.TranscriptRoots {
		realRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue // unresolvable root — unmeasured, never negative
		}
		resolvable = true
		found, fresh := findTranscript(realRoot, key, now.Add(-window))
		if found {
			if fresh {
				return SignalAffirmative
			}
			return SignalNegative
		}
	}
	if !resolvable {
		return SignalUnmeasured
	}
	// Resolvable roots carried no transcript for this key: unmeasured —
	// the session may live under an unscanned profile root.
	return SignalUnmeasured
}

// findTranscript checks a profile root for a transcript file named for the
// session key, reporting whether one was found and whether its mtime is
// inside the window. The scan is bounded two levels deep
// (<root>/<key>.jsonl and <root>/<project-dir>/<key>.jsonl — the shape the
// transcript writers stamp) so a per-key probe stays cheap under the hook
// budget (REQ-HYG-014).
func findTranscript(root, key string, freshAfter time.Time) (found, fresh bool) {
	sessionJSON := key + ".jsonl"
	check := func(p string) (bool, bool) {
		info, err := os.Stat(p)
		if err != nil {
			return false, false
		}
		return true, info.ModTime().After(freshAfter)
	}
	if ok, isFresh := check(filepath.Join(root, sessionJSON)); ok {
		return ok, isFresh
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return false, false
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if ok, isFresh := check(filepath.Join(root, e.Name(), sessionJSON)); ok {
			return ok, isFresh
		}
	}
	return false, false
}

// readRegistryEntries decodes the active-sessions registry file.
func readRegistryEntries(path string) ([]RegistryEntry, error) {
	blob, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []RegistryEntry
	if err := json.Unmarshal(blob, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
