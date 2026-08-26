// lsel-drain-loop.js — `/loop` REMINDER recipe for the LSEL drain
// (SPEC-LSEL-LOCAL-EVOLUTION-001 M2; truth-in-advertising corrected by
// SPEC-LSEL-DRAIN-STALL-001 M1, REQ-LDS-010).
//
// WHAT THIS ACTUALLY IS: a MODEL-MEDIATED REMINDER. Its body only PRINTS commands
// via console.log — it executes nothing, and never has (the earlier header claim
// that it ran drain.sh was false and masked the 3-week drain stall). The durable
// mechanical trigger is the session-start wrapper `session_drain.sh` (exclusive
// lock -> unconditional clusters.json archive -> drain.sh -> one-line status ->
// fail-open), which runs on every Claude session start once wired locally
// (spec.md section E local deliverable). This recipe stays as the in-session
// mid-loop reminder for long sessions: SessionStart fires once per session, and
// mid-session inbox accumulation has no other nudge surface.
//
// Invocation: /loop 30m lsel-drain   (or any interval; the recipe is read-only)
//
// It prints the backlog-check command, the wrapper drain command, and the
// candidate-count query, then stops. It does NOT draft proposals (the curator's
// model-mediated PROPOSE stage reads archived clusters-history copies) and it
// does NOT apply anything.

// The runtime requires `export const meta` to be the FIRST statement in the
// script; a file in this directory without it is skipped at scan time with a
// warning, which is how this recipe went unregistered despite AC-LSEL-007
// recording it as registered. Keep meta first — comments above it are fine,
// executable statements are not.
export const meta = {
	name: 'lsel-drain-loop',
	description: 'Read-only LSEL drain REMINDER: prints backlog-check + session_drain.sh wrapper commands; executes nothing (the mechanical trigger is the session-start wrapper)',
	whenToUse: 'Scheduled via /loop on an interval. Read-only — never commits, pushes, or enters run-phase.',
	phases: [
		{ title: 'Remind', detail: 'print backlog-check + wrapper drain commands + archived candidate count' },
	],
}

// cadence-bridge invariant: this recipe is READ-ONLY. No commit, no push, no
// run-phase entry, and no command execution — output only. If a future edit adds
// a write, it violates the bridge.
const INBOX = ".moai/lessons-inbox.jsonl";
const STATE_DIR = ".moai/state/lsel";
const SESSION_DRAIN = ".claude/skills/hns-lsel-curator/session_drain.sh";
const BACKLOG_CHECK = ".claude/skills/hns-lsel-curator/backlog_check.sh";

// Step 1 (PRINTED): advisory backlog check (emits a system-reminder if overflow).
// Step 2 (PRINTED): the wrapper drain — ALL drains route through session_drain.sh
// (exclusive lock + unconditional archive-before-overwrite); a direct drain.sh
// call bypasses archiving and can silently discard staged candidates.
// Step 3 (PRINTED): candidate count from the newest ARCHIVED copy — the live
// clusters.json is ephemeral (a no-op session-start drain overwrites it with
// candidates: []).
//
// This recipe is invoked by the native `/loop` scheduler on an interval. The
// orchestrator reads its printed output and runs the commands itself; proposal
// drafting stays on-demand (curator PROPOSE), never scheduled.
console.log("lsel-drain-loop: reminder only — prints commands, executes nothing");
console.log("backlog-check: " + BACKLOG_CHECK + " --inbox " + INBOX + " --state-dir " + STATE_DIR);
console.log("drain (wrapper-mediated): " + SESSION_DRAIN + " --inbox " + INBOX + " --state-dir " + STATE_DIR);
console.log("report: LATEST=$(ls -t " + STATE_DIR + "/clusters-history/*.json | head -1); jq '.candidates | length' \"$LATEST\"");
console.log("lsel-drain-loop: read-only complete; proposal drafting is on-demand (curator PROPOSE), not scheduled.");
