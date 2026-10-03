package verify

import (
	"strconv"
	"strings"
	"time"
)

// This file holds the pure decision pieces of `moai verify run`
// (SPEC-VERIFY-RUN-REUSE-001): the canonical command string, the environment
// digest, and the reuse predicate. Execution, timeouts and recording live in
// the CLI (internal/cli/verify_run.go); nothing here touches the filesystem or
// the clock, so every input is injected.

// unsetEnvMarker and setEnvPrefix keep an unset variable distinct from one set
// to the empty string — and from one whose value happens to spell the marker.
const (
	unsetEnvMarker = "unset"
	setEnvPrefix   = "set:"
)

// CanonicalCommand renders argv as the single command string a snapshot entry
// stores and FindCommand matches byte for byte: elements joined by one space,
// an element that is empty or holds whitespace or a quote rendered with
// strconv.Quote so that ["a b"] and ["a", "b"] never collide.
func CanonicalCommand(argv []string) string {
	parts := make([]string, len(argv))
	for i, a := range argv {
		if a == "" || strings.ContainsAny(a, " \t\r\n\v\f\"'") {
			parts[i] = strconv.Quote(a)
			continue
		}
		parts[i] = a
	}
	return strings.Join(parts, " ")
}

// EnvDigest is the config digest of the environment variables a result depends
// on: one `env:NAME` input per name, whose value tells an unset variable from
// one set to the empty string. lookup is os.LookupEnv in production. With no
// names it is the digest of the empty input set, so a run that binds no
// environment still carries a non-empty digest.
func EnvDigest(names []string, lookup func(string) (string, bool)) string {
	inputs := make(map[string]string, len(names))
	for _, name := range names {
		if v, ok := lookup(name); ok {
			inputs["env:"+name] = setEnvPrefix + v
		} else {
			inputs["env:"+name] = unsetEnvMarker
		}
	}
	return ConfigDigest(inputs)
}

// @MX:ANCHOR: [AUTO] verify-run reuse decision — the single place that decides whether a recorded result stands in for re-running a command
// @MX:REASON: loosening any leg turns a stale or failed recording into a fabricated pass; callers are the verb, its tests, and any later gate wiring
// @MX:SPEC: SPEC-VERIFY-RUN-REUSE-001

// DecideReuse reports whether snap holds an entry that may stand in for
// running state.Command now (REQ-VRR-002). Reuse needs all of: the snapshot
// key equals state's Head:TreeDigest; an entry whose command matches byte for
// byte; CheckReceipt accepts the five bound fields and the TTL; and the entry
// itself passed (exit code 0 and verdict "pass"). The returned reason names
// the first leg that failed, and the entry is returned only on a hit. now and
// ttl are injected; ttl <= 0 selects DefaultTTL.
func DecideReuse(snap *Snapshot, state ReceiptState, now time.Time, ttl time.Duration) (bool, *CheckEntry, string) {
	if snap == nil {
		return false, nil, "no snapshot recorded for the current tree"
	}
	if want := state.Head + ":" + state.TreeDigest; snap.Key != want {
		return false, nil, "snapshot key differs from the current tree"
	}
	entry := snap.FindCommand(state.Command)
	if entry == nil {
		return false, nil, "no entry recorded for this command"
	}
	receipt := &Receipt{
		CheckID:      entry.CheckID,
		Head:         state.Head,
		TreeDigest:   state.TreeDigest,
		ConfigDigest: entry.ConfigDigest,
		Command:      entry.Command,
		ToolVersion:  entry.ToolVersion,
		ExitCode:     entry.ExitCode,
		Verdict:      entry.Verdict,
		RecordedAt:   entry.RecordedAt,
	}
	if check := CheckReceipt(receipt, state, now, ttl); !check.Run {
		return false, nil, check.Reason
	}
	if entry.ExitCode != 0 || entry.Verdict != "pass" {
		return false, nil, "recorded run did not pass (exit " + strconv.Itoa(entry.ExitCode) + ", verdict " + strconv.Quote(entry.Verdict) + ")"
	}
	return true, entry, ""
}
