# SPEC-HANDOFF-RESUME-ASK-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-21

Basis for the explicit signal: `moai spec lint` exited 0 on the SPEC directory in both the plain and the `--strict` form (record below). This is the author's own lint, not an audit verdict; no plan audit has run yet.

**Plan-phase entry (authoring).** spec.md, plan.md, acceptance.md written by manager-spec in the card worktree for card t1672 (Tier M — rationale in `spec.md` §A; Class C design change, P1, operator origin; the operator approved the three-layer design before dispatch). Tree `8249bff8c3933d647b4f99e59048595f9808e83a`, branch `WT-clear-cc-codex`. Not yet audited. Artifact set: 3 Tier M artifacts + this file.

**Source of authority.**

1. The operator-approved design plan: `.moai/reports/moai-handoff-resume-plan-20261010.md` (primary checkout; the approved SSOT for the design — three layers, config keys, file set, completion criteria, out-of-scope items).
2. Card t1672 text (the dispatch authority; Korean card body quoted in the dispatch, [핸드오프 재개·운영자 승인 10-10·Class C·P1]).
3. The code as it exists in this tree — every structural claim in `spec.md` §C re-read at `8249bff8c3933d647b4f99e59048595f9808e83a` before authoring; where the plan doc and the code disagree, the code wins for as-is statements and the divergence is recorded below.

**Plan-phase decision note (operator approval — no open operator questions).** The operator approved the design on **2026-10-10** (board record **`d-20261010T113836Z-a35a`**), covering the three-layer design (L1 launcher question / L2 in-session ask mode / L3 /clear machine capture), the config key set (`handoff.mode` manual|auto|ask, `handoff.capture_on_clear`, `handoff.launcher_prompt`) with template-vs-project defaults, the file set, and the out-of-scope items (Codex `/new` source value; machine-capture summary quality). **No open operator question blocks plan-audit**: the four decision points in `spec.md` §E are design defaults applied under that approval (Q1 ask-notice sources startup+clear; Q2 ask-mode stale expiry = auto's; Q3 empty/EOF launcher answer = fresh start; Q4 launcher locale = project conversation_language, fallback en). The two remaining open items are measurement obligations inside run phase (Codex SessionStart `source` value; Codex launcher question — AC-HRA-014), not operator decisions; the Codex in-session path is best-effort by approved design and the launcher question is its guaranteed path.

**SPEC ID self-check, executed as Bash** (pre-write): `ID="SPEC-HANDOFF-RESUME-ASK-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` printed `PASS`. Uniqueness at this tree: `ls .moai/specs | grep -i 'RESUME-ASK'` matched only the new directory (no pre-existing SPEC carries the ID).

**Counts.** 14 requirements (Tier M ceiling 16; patterns: 8 event-driven, 4 ubiquitous, 1 state-driven, 1 event-detected) and 15 acceptance criteria (ceiling 16); 6 milestones (M1–M6).

**Frontmatter schema.** All 12 canonical fields present with canonical spellings (`created`/`updated`/`tags`/`id` — no snake_case aliases); `status: draft`; `phase: "v3.2.0 target"` (a release target, not a lifecycle stage); optional fields used: `tier: M`, `card: t1672`, `related_specs` (SPEC-HANDOFF-AUTORESUME-001, SPEC-HANDOFF-NEUTRAL-001, SPEC-INFINITE-GOAL-001 — all three directories exist in this tree). plan.md and acceptance.md are stateless (no `status:` field).

**RED-now evidence (acceptance.md ledger, run in this plan phase at the pinned tree).** Empty sweeps, each exit 0 under the runner, each a FAIL under the swept-count rule:

```text
E1  go test ./internal/hook -run '^(TestHandoffInjectBranchMatrix|TestHandoffAskNoticeContentAndLocales|TestHandoffAskNeverConsumes)$' -count=1 -v
    → testing: warning: no tests to run / PASS / ok .../internal/hook 0.739s [no tests to run]
E2  go test ./internal/hook -run '^(TestSessionEndClearCaptureCreatesSingleAutoCaptureRow|TestSessionEndCaptureSkipsWhenLivePendingExists|TestSessionEndNonClearReasonCreatesNoRecord|TestCaptureWithinBudgetAndFailOpen)$' -count=1 -v
    → testing: warning: no tests to run / PASS / ok .../internal/hook 0.489s [no tests to run]
E3  go test ./internal/cli -run '^(TestHandoffResumeClaimsPrintsConsumes|TestHandoffResumeNoPendingExitsNonZero|TestHandoffResumeConcurrentClaimLoses|TestLauncherPromptGateOnTTYSeam|TestLauncherPromptRendererAndParser|TestLauncherResumePassesBodyAsFirstPrompt|TestLauncherFreshStartDeclineEvent|TestLauncherViewPrintsWithoutClaim|TestLauncherNonTTYSkipsSilently)$' -count=1 -v
    → testing: warning: no tests to run / PASS / ok .../internal/cli 1.328s [no tests to run]
E4  go test ./internal/config -run '^(TestHandoffModeValidationRejectsUnknown|TestHandoffModeValidationAcceptsAsk|TestHandoffConfigTemplateDefaults)$' -count=1 -v
    → testing: warning: no tests to run / PASS / ok .../internal/config 0.328s [no tests to run]
```

No behavioral RED was observed (the tests do not exist yet; manager-develop records behavioral REDs at each milestone in §E.2).

**Existing-guard baseline (regression-guard half of AC-HRA-013), observed GREEN in this plan phase:**

```text
E5  go test ./internal/hook -run '^(TestHandoffConfig_Fallbacks|TestHandoffStale_ZeroTime|TestRenderHandoffContext_AllDirectives|TestHandoffInjectRender)$' -count=1 -v
    → --- PASS: TestHandoffConfig_Fallbacks (3 subtests) + the remaining named guards; ok .../internal/hook 1.094s
E5b go test ./internal/hook/handoff -count=1  → ok .../internal/hook/handoff 14.257s (pending_test.go + persist_test.go)
E5c go test ./internal/cli -run '^(TestHandoffShow_PendingSource|TestHandoffSave_Schema|TestHandoffClear|TestHandoffCmdRegistered)$' -count=1 -v
    → 4 --- PASS lines; ok .../internal/cli 3.558s
```

**Lint (command, judging build, tree).** `moai spec lint .moai/specs/SPEC-HANDOFF-RESUME-ASK-001` printed `✓ No findings — all SPEC documents are valid`, exit 0; `moai spec lint --strict` on the same directory printed the same line, exit 0. Judging build: the installed `moai` reports `v3.2.0-rc.37 (commit: 8249bff8c3, built: 2026-10-19)` — its build commit equals this tree's HEAD (`8249bff8c3933d647b4f99e59048595f9808e83a`) exactly, so the judging build was made from the tree under measurement (verification-claim-integrity §2.2: no lag to state).

**Hygiene sweep over the three artifacts:** duration-style time-estimate patterns — grep exit 1 (no hits); `origin/` refs — grep exit 1 (no hits). Line numbers cited in the SPEC are coordinates at the tree above and stated as such in `spec.md` §C and `plan.md` §A.

**Divergences measured between the plan doc / card text and this tree** (the code wins for as-is statements):

1. **Output-style "banner 7" pending.json drift does not exist here.** The card expects an output-style banner pending.json drift to correct; measured: `grep -rn 'pending.json' .claude/output-styles/` returns zero hits at this tree (either already fixed by an earlier card or the drift never lived there). The live pending.json drift is the single hit at `.claude/rules/moai/workflow/session-handoff.md:25` (the plan doc's plural "rules docs" claim measured as exactly 1 hit; `-format.md` and `-examples.md` carry zero). The output-style obligation for this SPEC is therefore the ask-mode amendment of the auto-flow section (`.claude/output-styles/moai/moai.md` + its template mirror), not a pending.json fix — encoded in REQ-HRA-014 / AC-HRA-015(d).
2. **`resume_handoffs` has no provenance `kind` column** (`internal/homestate/factory.go:93-118`). The plan doc's `kind: auto-capture` marker requires a schema addition not mentioned there; the `ALTER TABLE` migration block at `factory.go:396` is the precedent (plan.md §C.3).
3. **Unknown `handoff.mode` is coerced, not rejected, today** (`handoffConfig`, `internal/hook/handoff_inject.go:196-210`, observed live in E5: `WARN ... unknown mode; treating as manual mode=bogus`). The card's "알 수 없는 mode 거부" is encoded as a config-loader validation requirement (REQ-HRA-013) with the hook coercion retained as fail-open defense-in-depth — two layers, both stated.
4. Everything else in the plan doc's "지금 구조" verified unchanged: the single inject+consume cell (`clear ∧ auto`), startup deliberately excluded, the `save|clear|show` verb set, the two-key `HandoffConfig`, both yaml copies at `manual`/`false`, the closed `{manual, auto}` seam enum, the unused `input.Reason` field, and the Codex positional first-prompt channel.

**Open questions carried (none blocks plan-audit; full list in `spec.md` §G).** Codex SessionStart `source` value (measured once in run, AC-HRA-014); Codex `/new` source value (out of scope by card mandate); capture's last-prompt read mechanism (fail-open by requirement); the card-text divergences above (recorded, no action beyond REQ-HRA-014 as encoded).

**Residual judgment calls.** (1) The auto-capture is skipped — not queued — when another session's live pending row exists (REQ-HRA-011; the one-pending unique index makes "queue" unrepresentable). (2) The capture budget is enforced as a code-path ceiling with a `MOAI_HANDOFF_CAPTURE_BUDGET` env override the Codex hook configuration sets (plan.md §C.3) — the harness split (10s/3s) is dispatch-fixed, the mechanism is plan-level. (3) Tier M rather than L: the Go run scope is ~11 files / well under 1000 LOC and the approved design doc removes the design uncertainty Tier L's design.md/research.md would carry; the plan auditor may propose tier-up, and no artifact here depends on it not happening.

## §E.2 Run-phase Evidence

_<pending run-phase>_ — manager-develop records, per milestone, the verbatim RED output before GREEN, the GREEN output, mutant-probe results, and the four manual observations of AC-HRA-014 (Codex SessionStart `source` value; launcher TTY question; resume-as-first-prompt; Codex launcher question).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending run-phase>_
