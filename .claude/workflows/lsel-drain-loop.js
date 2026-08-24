// lsel-drain-loop.js — default `/loop` recipe for the LSEL drain trigger
// (SPEC-LSEL-LOCAL-EVOLUTION-001 M2, AC-LSEL-007 / REQ-LSEL-007).
//
// This is a READ-ONLY scheduled recipe (cadence-bridge compliant: scheduled runs
// never commit, never push, never enter run-phase). It triggers the LSEL drain
// (hns-lsel-curator/drain.sh) on a schedule so the loop does NOT depend on the
// orchestrator remembering to invoke it (the audit's exact failure mode, report
// §11 mustFix B#1).
//
// Invocation: /loop 30m lsel-drain   (or any interval; the recipe is read-only)
//
// The recipe runs drain.sh, reports the candidate count, and stops. It does NOT
// draft proposals (that is the curator's model-mediated PROPOSE stage, M2+) and
// does NOT apply anything (M3). Proposal drafting from clusters.json is an
// on-demand model-mediated step, not a scheduled one.

// The runtime requires `export const meta` to be the FIRST statement in the
// script; a file in this directory without it is skipped at scan time with a
// warning, which is how this recipe went unregistered despite AC-LSEL-007
// recording it as registered. Keep meta first — comments above it are fine,
// executable statements are not.
export const meta = {
	name: 'lsel-drain-loop',
	description: 'Read-only LSEL drain trigger: advisory backlog check, mechanical drain, candidate count',
	whenToUse: 'Scheduled via /loop on an interval. Read-only — never commits, pushes, or enters run-phase.',
	phases: [
		{ title: 'Drain', detail: 'backlog check + mechanical inbox drain, then report candidate count' },
	],
}

// cadence-bridge invariant: this recipe is READ-ONLY. No commit, no push, no
// run-phase entry. If a future edit adds a write, it violates the bridge.
const INBOX = ".moai/lessons-inbox.jsonl";
const STATE_DIR = ".moai/state/lsel";
const DRAIN = ".claude/skills/hns-lsel-curator/drain.sh";
const BACKLOG_CHECK = ".claude/skills/hns-lsel-curator/backlog_check.sh";

// Step 1: advisory backlog check (emits a system-reminder if overflow).
// Step 2: run the mechanical drain (read-only over the inbox; advances offset).
// Step 3: report candidate count (read-only; no proposal drafting here).
//
// This recipe is invoked by the native `/loop` scheduler on an interval. The
// orchestrator reads its output and, if candidates warrant, invokes the curator
// skill's PROPOSE stage on demand (model-mediated, not scheduled).
console.log("lsel-drain-loop: advisory backlog check + mechanical drain (read-only)");
console.log("backlog-check: " + BACKLOG_CHECK + " --inbox " + INBOX + " --state-dir " + STATE_DIR);
console.log("drain: " + DRAIN + " --inbox " + INBOX + " --state-dir " + STATE_DIR);
console.log("report: jq '.candidates | length' " + STATE_DIR + "/clusters.json");
console.log("lsel-drain-loop: read-only complete; proposal drafting is on-demand (curator PROPOSE), not scheduled.");
