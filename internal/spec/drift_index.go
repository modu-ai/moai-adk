package spec

import (
	"fmt"
	"os/exec"
	"strings"
)

// drift_index.go — single-pass commit index for drift detection
// (SPEC-SESSIONSTART-PERF-001 M1, REQ-SSP-001 / REQ-SSP-002).
//
// The previous drift path spawned one `git log --grep=<specID>` subprocess per
// active SPEC (plus a second `--grep=<scope-prefix>` subprocess for the
// combined-scope fallback), making session-start cost O(n) in the SPEC count.
// This file replaces those per-SPEC subprocesses with ONE `git log` pass whose
// result is walked in memory, so the subprocess count is constant.
//
// The classification helpers themselves (shouldSkipCommitTitle,
// commitMatchesSPECID, ClassifyPRTitle, combinedScopeCloseMatches,
// deriveScopePrefix) are REUSED VERBATIM. Only the commit SOURCE changes — from
// a per-SPEC subprocess to a shared in-memory index. The matching basis
// (full-message candidate window + subject re-filter) is preserved bit-for-bit.

// Field and record separators for the single-pass git log format. ASCII unit
// separator (0x1f) and record separator (0x1e) are used because commit bodies
// routinely contain newlines, so a newline-delimited format cannot be parsed
// unambiguously.
const (
	gitLogFieldSep  = "\x1f"
	gitLogRecordSep = "\x1e"
)

// gitLogFullMessageFormat captures, per commit: the SUBJECT (%s) followed by the
// FULL RAW MESSAGE (%B).
//
// %B — not %s+%b — is the load-bearing choice: `git log --grep=<pat>` matches
// against the full raw message, and %B is exactly that. Reconstructing the
// message from %s + %b would silently drop any line that sits between the
// subject and the first blank line, which would shrink the candidate set and
// could flip the drift count.
const gitLogFullMessageFormat = "--format=%s" + "%x1f" + "%B" + "%x1e"

// commitRecord holds the two views of one commit the drift walker needs:
//
//   - subject: the stage-2 classification input (what `--oneline` printed before)
//   - fullMsg: the stage-1 candidate-match input (what `--grep` matched against)
//
// Keeping both is what allows the in-memory walk to reproduce the original
// two-stage semantics exactly.
type commitRecord struct {
	subject string
	fullMsg string
}

// gitLogAllFullMessage performs THE single git log pass: one subprocess that
// streams the whole (non-merge) history of branch, newest-first, carrying each
// commit's subject and full raw message.
//
// The pass is deliberately UNBOUNDED (no -N): the per-SPEC gitLogWindowSize cap
// is applied in memory instead. A bounded global window could truncate history
// before an older-but-still-active SPEC's own 50-commit window was reachable,
// silently changing its classification.
//
// @MX:NOTE: [AUTO] one subprocess replaces ~2×N per-SPEC `git log --grep` spawns.
// @MX:REASON: SPEC-SESSIONSTART-PERF-001 REQ-SSP-001 — session-start blocked ~15s
//
//	on process-spawn thrash; the cost must be independent of the SPEC count.
func gitLogAllFullMessage(branch string) ([]commitRecord, error) {
	cmd := exec.Command("git", "log", branch, "--no-merges", gitLogFullMessageFormat)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	return parseCommitRecords(string(output)), nil
}

// parseCommitRecords splits the single-pass git log output into commitRecords,
// preserving git's newest-first ordering.
func parseCommitRecords(output string) []commitRecord {
	raw := strings.Split(output, gitLogRecordSep)
	records := make([]commitRecord, 0, len(raw))

	for _, entry := range raw {
		// git separates log entries with a newline, which lands at the head of
		// every entry after the first.
		entry = strings.TrimLeft(entry, "\r\n")
		if entry == "" {
			continue
		}

		subject, fullMsg, found := strings.Cut(entry, gitLogFieldSep)
		if !found {
			// Defensive: git always emits the separator, so this is unreachable
			// in practice. Skipping is safer than guessing at the field split.
			continue
		}

		subject = strings.TrimSpace(subject)
		if subject == "" && strings.TrimSpace(fullMsg) == "" {
			continue
		}

		records = append(records, commitRecord{subject: subject, fullMsg: fullMsg})
	}

	return records
}

// gitHeadSHA resolves the current HEAD commit SHA — the drift cache key.
//
// This is deliberately NOT routed through the drift seam (driftDeps.logAll):
// it is an O(1) query, and a cache HIT must be able to answer without doing any
// git-log work at all (AC-SSP-004).
func gitHeadSHA() (string, error) {
	output, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// inMemImpliedStatus is the in-memory equivalent of getGitImpliedStatus: it
// infers a SPEC's lifecycle status from the shared commit index instead of from
// a dedicated `git log --grep=<specID>` subprocess.
//
// It reproduces the original TWO-STAGE semantics exactly:
//
//	stage 1 (candidate set) — the newest gitLogWindowSize commits whose FULL raw
//	  message (subject OR body) contains specID as a substring. This mirrors
//	  `git log --grep=<specID> -50`: `--grep` matches the full message, and git
//	  applies the -N window AFTER that filter. The candidate set is therefore
//	  BODY-DEPENDENT, and the window counts full-message matches — not subject
//	  matches.
//	stage 2 (re-filter + classify) — each candidate, newest-first, runs through
//	  shouldSkipCommitTitle (chore-skip, LSCSK-001) → commitMatchesSPECID (SUBJECT
//	  word-boundary, LSGF-001) → ClassifyPRTitle. The first non-empty
//	  classification wins.
//
// Because `--grep` takes a POSIX basic regular expression and a SPEC-ID contains
// no BRE metacharacters (letters, digits and hyphens only), it degrades to a
// literal, case-sensitive substring match — which strings.Contains reproduces.
//
// @MX:ANCHOR: [AUTO] inMemImpliedStatus — the drift behavior-preservation invariant.
// @MX:REASON: SPEC-SESSIONSTART-PERF-001 REQ-SSP-005 [HARD] + design.md §M1.3.
//
//	Collapsing stage 1 to a SUBJECT-token index (e.g. keying on ExtractSPECIDs of
//	the subject) builds a DIFFERENT candidate set and can flip the drift count: a
//	SPEC-ID appearing in the BODIES of >= gitLogWindowSize commits newer than its
//	own subject-close commit exhausts the original walker's window but would NOT
//	exhaust a subject-only index. Any change here MUST be re-validated against the
//	real-corpus record equivalence check (AC-SSP-005c), not just the unit fixtures.
func inMemImpliedStatus(commits []commitRecord, specID string) (string, error) {
	matched := 0

	for _, c := range commits { // newest-first, as git emitted them
		// stage 1: full-message substring match (the `--grep` equivalent).
		if !strings.Contains(c.fullMsg, specID) {
			continue
		}

		matched++
		if matched > gitLogWindowSize {
			// Window exhausted. The original walker only ever saw the newest N
			// full-message matches, because git applied -N after --grep.
			break
		}

		// stage 2: subject-scoped re-filter, then classification.
		if shouldSkipCommitTitle(c.subject) {
			continue
		}
		if !commitMatchesSPECID(c.subject, specID) {
			continue
		}

		_, status, err := ClassifyPRTitle(c.subject)
		if err != nil || status == "" {
			// Unknown prefix or classification failure — keep walking.
			continue
		}

		return status, nil
	}

	// Both exhaustion paths preserve the original error text, so callers that
	// distinguish them (and the lint rule that treats an error as "skip") keep
	// behaving identically.
	if matched == 0 {
		return "", fmt.Errorf("no git history found for %s", specID)
	}

	return "", fmt.Errorf("no classifiable commit within window of %d for %s", gitLogWindowSize, specID)
}

// inMemCombinedScopeClose is the in-memory equivalent of resolveCombinedScopeClose:
// the FALLBACK-ONLY secondary lookup that finds a combined-scope close commit
// naming only the scope-prefix (e.g. `chore(SPEC-CCSYNC): ... 3-phase close
// (CLAUDEMD + TOOLCAT)`), which the per-SPEC full-ID walk can never reach.
//
// It mirrors the original `git log --grep=<scope-prefix> -50` exactly: the newest
// gitLogWindowSize commits whose FULL raw message contains the scope-prefix, each
// re-filtered through the combinedScopeCloseMatches 3-gate matcher (which
// preserves LSGF-001 by requiring a word-boundary token match on the
// distinguishing segment).
func inMemCombinedScopeClose(commits []commitRecord, specID string) bool {
	prefix := deriveScopePrefix(specID)

	// broad-prefix guard (identical to resolveCombinedScopeClose): a single-domain
	// ID collapses to a bare `SPEC`, which would match every SPEC family. A
	// combined-scope close only ever occurs in a multi-segment family.
	if prefix == "SPEC" || prefix == "" || !strings.HasPrefix(prefix, "SPEC-") {
		return false
	}

	matched := 0

	for _, c := range commits { // newest-first
		if !strings.Contains(c.fullMsg, prefix) {
			continue
		}

		matched++
		if matched > gitLogWindowSize {
			break
		}

		if combinedScopeCloseMatches(c.subject, specID) {
			return true
		}
	}

	return false
}
