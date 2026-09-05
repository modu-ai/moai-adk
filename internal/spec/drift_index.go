package spec

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/execerr"
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
		// execerr.StatusDetail, not %w: a raw *exec.ExitError chain would be
		// mistaken for an intentional ExitCoder at the cmd/moai seam (t130).
		return nil, fmt.Errorf("git log failed: %s", execerr.StatusDetail(err))
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
		return "", fmt.Errorf("git rev-parse HEAD failed: %s", execerr.StatusDetail(err))
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

// inMemBodyDeclaredClose is the THIRD judgment axis in the FALLBACK-ONLY slot: it
// recognizes a close that was declared in a commit's BODY because the subject could
// not carry the target SPEC-ID (SPEC-DRIFT-CLOSE-BODY-001, REQ-DCB-001).
//
// The measured instance is e979a4d13, `chore(SPEC group C): Mx-phase close (...)`:
// the subject names an arbitrary combined scope, so ExtractSPECIDs yields nothing
// and the stage-2 re-filter of inMemImpliedStatus drops the commit. The walker keeps
// descending, meets an older `docs(...)` commit, and reports `in-progress` for a
// SPEC that is genuinely closed.
//
// It sits beside inMemCombinedScopeClose rather than replacing it: that fallback
// answers the scope-PREFIX shape (`chore(SPEC-CCSYNC): ... (CLAUDEMD + TOOLCAT)`),
// which is reachable from the subject alone. This one answers the shape where the
// subject derives from no SPEC-ID at all and only the body names the SPEC.
//
// Two gates bound the candidate window:
//
//	gate (a) the commit SUBJECT must NOT name specID. A subject that does name it
//	         belongs to the primary walk, and re-deciding it here would breach the
//	         behavior-preservation invariant (REQ-DCB-006, §5.4 condition 2).
//	gate (b) the body-line predicate (bodyDeclaresClose) must find a declaration.
//
// There is deliberately NO subject-side close-signal filter. closeInfixMatch was the
// obvious candidate and was REJECTED BY MEASUREMENT: it admits only `3-phase close` /
// `4-phase close` / `mx-phase audit-ready`, and the confirmed target subject reads
// `Mx-phase close` — 2 of the 6 measured close commits, this card's own target among
// them, fail that filter (spec.md §5.4). Widening the constant set would change
// shouldSkipCommitTitle, the combined-scope gate and ClassifyPRTitle at once, which
// is a REQ-DCB-006 violation. The discriminating weight therefore rests entirely on
// the body-line predicate, as §5.4 condition 1 requires; gate (a) is a cheap
// candidate reducer, not a discriminator.
//
// The candidate window mirrors the primary walk bit-for-bit — full-message substring
// match, newest-first, capped at gitLogWindowSize matches — so no new subprocess is
// introduced (SPEC-SESSIONSTART-PERF-001 REQ-SSP-001).
//
// @MX:NOTE: [AUTO] body-declared close — third FALLBACK-ONLY axis; the output is
//
//	`completed` or no verdict, never any other status (REQ-DCB-005).
//
// @MX:REASON: SPEC-DRIFT-CLOSE-BODY-001 REQ-DCB-001/002/003/006 — the reverse-direction
//
//	risk (a false `completed` frontmatter absolved by one body mention) is held by
//	bodyDeclaresClose alone, so any widening of that predicate reopens candidate A,
//	which was rejected for exactly that reason (spec.md §5, §5.2).
func inMemBodyDeclaredClose(commits []commitRecord, specID string) bool {
	matched := 0

	for _, c := range commits { // newest-first
		if !strings.Contains(c.fullMsg, specID) {
			continue
		}

		matched++
		if matched > gitLogWindowSize {
			break
		}

		// gate (a): a subject naming the full ID is the primary walk's to classify.
		if commitMatchesSPECID(c.subject, specID) {
			continue
		}

		// gate (b): the body-line predicate.
		if bodyDeclaresClose(c.subject, c.fullMsg, specID) {
			return true
		}
	}

	return false
}

// bodyListMarkerPattern matches the leading list markers a declaration line may
// carry in a commit body. The set is deliberately narrow: `-`, `*` and `+` followed
// by whitespace, and nothing else. A blockquote `>` is not a list marker, and
// stripping one would let quoted text be read as a declaration.
var bodyListMarkerPattern = regexp.MustCompile(`^[ \t]*[-*+][ \t]+`)

// conventionalSubjectPattern requires a line to LOOK like a conventional-commit
// subject before it is fed to the classification chain: a lowercase type, an
// optional parenthesized scope, then `: `.
//
// This precondition is load-bearing rather than cosmetic, and fixture line 7 proves
// it. That line is prose from 80dea9684 which both names SPEC-INTERNAL-TEST-002 (as
// the FOLLOW-UP owner of residual debt — the opposite of a close) and contains the
// literal `3-phase close`. ClassifyPRTitle checks the close infix BEFORE the prefix
// loop (transitions.go), so feeding raw prose to the chain returns `completed` for
// it. Requiring the conventional-commit shape first is what makes spec.md §5.3's
// "the line is itself a conventional-commit subject" a mechanical condition rather
// than an implicit one.
var conventionalSubjectPattern = regexp.MustCompile(`^[a-z]+(\([^)]*\))?: `)

// bodyDeclaredCloseDenyKeys are leading keys that can never introduce a close
// declaration (REQ-DCB-004). The shape predicates below already reject them, so this
// is a second lock rather than the only one: it exists so a later widening of either
// shape cannot silently readmit a dependency reference.
var bodyDeclaredCloseDenyKeys = []string{"depends_on:", "related:", "supersedes:", "blocked_by:"}

// subjectCloseSignal is the shape-A-only narrowing found by the M3 corpus sweep.
//
// Shape A recognizes `<full-ID>: <anything>`, and "anything" turned out to include
// a record that is not a close. ac3e38a0b carries
// `- SPEC-MOAI-MCP-SERVER-001: REQ-MCP-002 opt-in->default-on, ... amended
// (0.1.0 -> 0.2.0)` under a subject that closes nothing
// (`feat(SPEC-MCP-DEFAULT-ON-001): moai MCP server as first-class default`); the
// commit body says in as many words that the status was preserved, not
// transitioned. No text predicate separates that line from a genuine verdict line
// (`SPEC-XXX-001: Mx verdict EVALUATE-PASS`) without keyword matching, which
// spec.md §5.3 rules out as the discriminator. So the narrowing is placed on the
// CONTAINING COMMIT: a shape-A line counts only inside a commit whose subject says
// it is closing something.
//
// The predicate is a bare `close` substring, deliberately broader than
// closeInfixMatch, which spec.md §5.4 rejected by measurement for admitting only
// three literals and dropping this card's own target. Measured against the close
// subjects on record, all pass: `Close out 2 SPECs ...` (7beda68a5),
// `chore(SPEC group C): Mx-phase close ...` (e979a4d13), `docs(specs): batch
// sync-phase close ...` (2f449e189), `docs: close out 4 A-tier SPECs ...`
// (cd21df594), `chore(spec): close KANBAN-RENAME-001 ...` (cd80f0644),
// `docs(SPEC-INTERNAL-TEST-001): sync-phase artifacts + 3-phase close`, and
// `feat(SPEC-HIERARCHICAL-TEAM-001): ... (Tier M, 3-phase close)` — satisfying
// §5.4 condition 3.
//
// It binds shape A ONLY. Shape B already requires the line itself to classify as
// `completed` through the validated chain, and 6da952899 is the measurement that
// makes the distinction load-bearing: its subject
// (`feat(factory+epic): Factory Mode multi-session bootstrap ...`) carries no close
// signal while two of its body lines are genuine 3-phase closes. Extending this
// gate to shape B would drop them.
//
// A narrowing can only ever reduce the set of cleared rows, so it cannot introduce
// a regression in the AC-DCB-005 sense; measured on the corpus it removed exactly
// one row (the false positive) and kept the other 18.
func subjectCloseSignal(subject string) bool {
	return strings.Contains(strings.ToLower(subject), "close")
}

// bodyDeclaresClose reports whether body contains a line declaring specID closed.
//
// It sweeps the WHOLE body and returns true on the first QUALIFYING line. Stopping
// at the first line of the right SHAPE would be wrong, and 7beda68a5 is the measured
// proof: its body names SPEC-WORKTREE-ENTRY-STRATEGY-001 on eight lines, of which
// exactly one — line 117 of 181 — is the close. A first-shape-match implementation
// stops at line 1 (`* fix(...): M1 web auto-toggles ...`), reads `implemented`, and
// never sees the close. No qualifying line means NO VERDICT, which is not the same
// as a verdict of "not closed".
//
// Two measured declaration shapes (spec.md §5.3); one hit on either is enough:
//
//	A — verdict list: `- SPEC-XXX-001: Mx verdict EVALUATE-PASS`. After the list
//	    marker the line starts with the full ID and the very next byte is `:`. The
//	    colon is not decoration: 51d18d3fe carries
//	    `SPEC-INTERNAL-SECURITY-001 f3193bac8 / SPEC-HANDOFF-GOALFIX-001`, a line that
//	    exists only because an enumeration wrapped at column 72 — it starts with a
//	    full ID and declares nothing. Without the colon requirement the verdict would
//	    turn on where a line happened to wrap.
//
//	B — squash sub-subject: `* docs(SPEC-XXX-001): sync-phase artifacts — 3-phase
//	    close`. A squash merge moves each original subject into the body, so these are
//	    lines the primary walk WOULD have adopted had they stayed subjects. They are
//	    therefore re-classified by the existing, already-validated chain
//	    (shouldSkipCommitTitle → commitMatchesSPECID → ClassifyPRTitle) rather than by
//	    a second text predicate, which keeps the judgment surface this SPEC opens as
//	    small as possible. A line qualifies ONLY when that chain returns `completed`;
//	    `implemented`, `in-progress`, `draft`, a skip and an unknown prefix are all
//	    non-declarations.
//
// The skip filter is applied, NOT bypassed (plan.md §B-3). `chore(spec):` /
// `chore(specs):` metadata sweeps and SPEC-ID-scoped backfill chores are excluded
// from lifecycle inference by AC-LSCSK-003 and REQ-DCA-002; letting a body line reach
// a conclusion the same text would be denied as a subject would reopen that guard
// through the back door.
func bodyDeclaresClose(subject, body, specID string) bool {
	subjectDeclaresClose := subjectCloseSignal(subject)

	for _, raw := range strings.Split(body, "\n") {
		line := strings.TrimRight(raw, " \t\r")

		stripped := strings.TrimSpace(bodyListMarkerPattern.ReplaceAllString(line, ""))
		if stripped == "" {
			continue
		}

		if hasDeclarationDenyKey(stripped) {
			continue
		}

		// shape A — full ID at line start, immediately followed by ':', inside a
		// commit whose subject says it is closing something (subjectCloseSignal).
		if subjectDeclaresClose && strings.HasPrefix(stripped, specID+":") {
			return true
		}

		// shape B — a conventional-commit subject the existing chain calls a close.
		if !conventionalSubjectPattern.MatchString(strings.ToLower(stripped)) {
			continue
		}
		if shouldSkipCommitTitle(stripped) {
			continue
		}
		if !commitMatchesSPECID(stripped, specID) {
			continue
		}
		if _, status, err := ClassifyPRTitle(stripped); err == nil && status == "completed" {
			return true
		}
	}

	return false
}

// hasDeclarationDenyKey reports whether the line opens with a key that references
// another SPEC rather than closing one (REQ-DCB-004).
func hasDeclarationDenyKey(stripped string) bool {
	lower := strings.ToLower(stripped)
	for _, key := range bodyDeclaredCloseDenyKeys {
		if strings.HasPrefix(lower, key) {
			return true
		}
	}
	return false
}
