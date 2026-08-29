# SPEC-CI-TEST-OBSERVABILITY-001 — Progress

Card: **t358** · Branch: `WT-ci-test-observability` · Base: `origin/develop c6aa61346`

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M).
- SPEC ID regex self-check executed as Bash: `[[ "SPEC-CI-TEST-OBSERVABILITY-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`.
- ID uniqueness confirmed against `.moai/specs/` (no prior `SPEC-CI-TEST-OBSERVABILITY-*`).
- Frontmatter carries all 12 canonical fields; `phase: "v3.1.4 target"` is a release target, not a lifecycle stage.
- Terminal AC is the positive control (`acceptance.md` AC-CTO-007), not a YAML-content assertion.
- Addendum folded in (coordinator, post-authoring): single-predicate design gain recorded as the deciding reason (`spec.md` §C.1); failure-body survival and `-json`+coverage coexistence recorded as observed, not assumed; scope stated at three call sites [**SUPERSEDED** — a fourth, `ci.yml:329`, was found in plan audit iter-1 and brought in scope by lead ruling. Preserved rather than rewritten per `verification-claim-integrity.md` §2: silently rewriting a measured-false claim would erase the record that this SPEC once asserted a completeness it did not have]; ONE-run dispatch constraint written into AC-CTO-007; house conventions (`upload-artifact@v7`, `retention-days: 7`, `jq` established) recorded.
- **Plan audit round 1: FAIL 0.81.** Nine blocking defects repaired (D1-D9) plus two citation corrections; D10/D11 addressed. Repair detail in the return report. AC-CTO-007 survived the adversarial pass unchanged and was NOT weakened — the auditor independently verified its two premises (`internal/statusline/usage_test.go:186` unconditional `t.Skip`; `profile_bench_test.go:305-307` env-gated with neither workflow setting the var), so the single dispatch is guaranteed to contain a real skip.
- **Scope corrected to FOUR call sites.** `ci.yml:329` (`test-integration`, 3-OS matrix, `-tags=integration`) was missed in the first draft and is in scope by lead ruling.
- **Named debt, carried openly (never reported as passes):**
  1. AC-CTO-003's CI-level red-path confirmation — verified locally instead; the CI red path is not exercised pre-merge.
  2. AC-CTO-005b (artifact present on a FAILED run) — `if: always()` is a declaration of intent, not observed behaviour.
  Both discharge on the first genuinely red CI run after this lands. Closing either by manufacturing a second dispatch is prohibited (lead's ONE-run constraint is [HARD]).
- If the fallback observation path is taken, AC-CTO-007 closes post-merge and that too is debt.
- Full-suite artifact size remains an EXTRAPOLATION from two packages; the real figure is recorded at AC-CTO-005a from the approved dispatch, which also supersedes the extrapolation label.
- **Plan audit iter-2: PASS-WITH-DEBT 0.895** (+0.085 monotonic; Tier M threshold 0.80 — also clears Tier L's 0.85). Final text-only pass applied (N1-N9): stale 3-site count corrected, REQ order fixed, AC-CTO-003 split so each criterion is executable exactly as written, DoD reconciled with §I, remote-ref cleanup added with its ordering constraint, plus four precision fixes.
- **Pattern-label correction**: iter-1's D11 was **withdrawn** — `Unwanted` is an EARS-legacy name absent from the canonical GEARS five (Ubiquitous, Event-driven, State-driven, Capability gate, Event-detected), so iter-1 had pushed away from the canonical name toward a non-canonical one. REQ-CTO-004 is restored to `(Event-detected)`; the three always-active prohibitions (REQ-CTO-003, -009, -011) carry `(Ubiquitous)`.
- Status: `draft`. Awaiting Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
