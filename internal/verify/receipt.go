package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"
)

// Receipt is the record a check leaves when it runs outside a hook because it
// cannot finish within the hook timeout (SPEC-DUAL-HARNESS-HOOK-PARITY-001
// design §D3.6, REQ-HPR-019). The hook never re-runs the check; it compares
// the five bound fields — Head, TreeDigest, ConfigDigest, Command, ToolVersion
// — against the current state, and accepts the recorded outcome only when all
// five are equal.
//
// A receipt is stored as a snapshot CheckEntry under the verify.Key of the
// tree it measured (Head + ":" + TreeDigest), so it shares the existing
// snapshot store and its atomic write; there is no separate receipt store.
type Receipt struct {
	CheckID      string
	Head         string
	TreeDigest   string
	ConfigDigest string
	Command      string
	ToolVersion  string
	ExitCode     int
	// Verdict is the check's own outcome word where it has one ("pass",
	// "fail", "inconclusive"); the gate that reads the receipt interprets it.
	Verdict    string
	RecordedAt time.Time
}

// ReceiptState is the current value of each bound field, as the Stop chain
// observes it.
type ReceiptState struct {
	Head         string
	TreeDigest   string
	ConfigDigest string
	Command      string
	ToolVersion  string
}

// ReceiptCheck is the comparison result. Run is true only when the receipt
// was accepted; every other outcome — absent, truncated, any field differing
// or unbound, past the TTL — is "not run", never "passed". Reason names what
// failed, for the continuation text the gate emits.
type ReceiptCheck struct {
	Run     bool
	Reason  string
	Receipt *Receipt
}

// @MX:ANCHOR: [AUTO] receipt comparison predicate — the one place the Stop chain decides whether an out-of-hook check counts as run
// @MX:REASON: consumed by the goal member, the sync gate, and the codex review gate (member 6) on the Codex Stop path; loosening any field comparison turns a stale receipt into a pass

// CheckReceipt compares a stored receipt with the current state. A nil stored
// receipt is absent. Each bound field must be non-empty on both sides and
// equal; an unbound field is not evidence even when the current value is also
// empty. The wall-clock TTL (Fresh's second leg; ttl <= 0 selects DefaultTTL)
// is kept as an extra staleness bound on top of the field comparison.
func CheckReceipt(stored *Receipt, current ReceiptState, now time.Time, ttl time.Duration) ReceiptCheck {
	if stored == nil {
		return ReceiptCheck{Reason: "receipt absent"}
	}
	fields := []struct {
		name          string
		stored, value string
	}{
		{"head", stored.Head, current.Head},
		{"tree_digest", stored.TreeDigest, current.TreeDigest},
		{"config_digest", stored.ConfigDigest, current.ConfigDigest},
		{"command", stored.Command, current.Command},
		{"tool_version", stored.ToolVersion, current.ToolVersion},
	}
	var problems []string
	for _, f := range fields {
		switch {
		case f.stored == "" || f.value == "":
			problems = append(problems, f.name+" unbound")
		case f.stored != f.value:
			problems = append(problems, f.name+" differs")
		}
	}
	if len(problems) > 0 {
		return ReceiptCheck{Reason: "receipt not valid for the current state: " + strings.Join(problems, ", ")}
	}
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	if now.Sub(stored.RecordedAt) > ttl {
		return ReceiptCheck{Reason: "receipt older than " + ttl.String()}
	}
	return ReceiptCheck{Run: true, Receipt: stored}
}

// RecordReceipt stores r as a snapshot entry under the key of the tree it
// measured, through RecordCheck (per-key lock, atomic rename).
func RecordReceipt(projectRoot string, r Receipt) error {
	_, err := RecordCheck(projectRoot, r.Head+":"+r.TreeDigest, CheckEntry{
		CheckID:      r.CheckID,
		Command:      r.Command,
		ExitCode:     r.ExitCode,
		RecordedAt:   r.RecordedAt,
		ConfigDigest: r.ConfigDigest,
		ToolVersion:  r.ToolVersion,
		Verdict:      r.Verdict,
	})
	return err
}

// LoadReceipt returns the receipt recorded for the current tree and command,
// or nil. Every failure — no snapshot, an unreadable or truncated snapshot
// file, no entry for the command — is absence, so a partially written receipt
// can never be read as a valid or failing one.
func LoadReceipt(projectRoot string, current ReceiptState) *Receipt {
	s, err := Load(projectRoot, current.Head+":"+current.TreeDigest)
	if err != nil || s == nil {
		return nil
	}
	e := s.FindCommand(current.Command)
	if e == nil {
		return nil
	}
	head, digest, _ := strings.Cut(s.Key, ":")
	return &Receipt{
		CheckID:      e.CheckID,
		Head:         head,
		TreeDigest:   digest,
		ConfigDigest: e.ConfigDigest,
		Command:      e.Command,
		ToolVersion:  e.ToolVersion,
		ExitCode:     e.ExitCode,
		Verdict:      e.Verdict,
		RecordedAt:   e.RecordedAt,
	}
}

// ConfigDigest hashes the configuration inputs a check reads (section file
// contents, switch variables), keyed by a name the producer and the Stop
// chain agree on. The result is independent of map order; names and values
// are length-prefixed so bytes cannot move between them unnoticed. Which
// inputs belong to a check is declared by that check, not here.
func ConfigDigest(inputs map[string]string) string {
	names := make([]string, 0, len(inputs))
	for name := range inputs {
		names = append(names, name)
	}
	sort.Strings(names)
	h := sha256.New()
	for _, name := range names {
		writeLenPrefixed(h, name)
		writeLenPrefixed(h, inputs[name])
	}
	return hex.EncodeToString(h.Sum(nil))[:digestHexLen]
}

func writeLenPrefixed(h interface{ Write([]byte) (int, error) }, s string) {
	var n [8]byte
	l := uint64(len(s))
	for i := range n {
		n[i] = byte(l >> (8 * i))
	}
	_, _ = h.Write(n[:])
	_, _ = h.Write([]byte(s))
}
