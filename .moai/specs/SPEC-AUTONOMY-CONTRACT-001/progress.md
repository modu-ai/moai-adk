# progress.md — SPEC-AUTONOMY-CONTRACT-001

> Plan-phase skeleton. §E.2 / §E.3 / §E.4 are placeholder headings only; they are populated by
> manager-develop (run-phase, §E.2/§E.3) and manager-docs (sync-phase, §E.4).

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts emitted (Tier L): spec.md, plan.md, acceptance.md, design.md, research.md, progress.md.
- Frontmatter: `id: SPEC-AUTONOMY-CONTRACT-001`, `status: draft`, `tier: L`.
- Card: t1234 (AUTONOMY-A1). Base: develop `ca1d5dc43`, branch `WT-contract-schema`.
- Schema draft for A2-A4: `design.md` § Contract Schema.
- Open decisions for plan-audit: plan.md §B D1-D9.
- Revision 0.2.0 after plan-audit iteration 1 (FAIL 0.74, report `.moai/reports/plan-audit/SPEC-AUTONOMY-CONTRACT-001-review-1.md`).
- Revision 0.3.0 after plan-audit iteration 2 (FAIL 0.84, report `.moai/reports/plan-audit/SPEC-AUTONOMY-CONTRACT-001-review-2.md`) plus lead-approved schema additions B1-B9.

## §E.2 Run-phase Evidence

Coordinator: manager-lead (Role A). Leaf workers run sequentially in this worktree (one writer at a time); evidence under `.moai/state/verify/t1234-run/` (gitignored) and `.moai/reports/t1234/` (gitignored).

M1: AC-CONTRACT-001=PASS, AC-CONTRACT-002=PASS, AC-CONTRACT-003=PASS, AC-CONTRACT-004=PASS | RED b93af204a, GREEN f40dea185 | evidence: .moai/state/verify/t1234-run/M1.* + lead.M1.* (lead re-run), RED .moai/reports/t1234/M1-red.txt | leaf model: opus | fold-at: 2026-09-26T13:15+09:00

SPEC text revised twice during the run by manager-spec under operator decision (v0.5.0 `67a2f55cb`, v0.5.1 `65e0a9167`); both landed on this branch between run-phase commits and touched only the SPEC artifacts. v0.5.1 adds the required `card` field, which changes committed M1 behavior (plan.md §B D10) — repaired as a separate RED→GREEN pair ("M1 repair").

M4: AC-CONTRACT-005=PASS, AC-CONTRACT-006=PASS | RED acb641d21, GREEN 086dfb2fd | evidence: .moai/state/verify/t1234-run/M4.* + lead.M4.* (lead re-run: corpus-size=823 non-zero=774; raw-crlf-count=2 normalized-count=2), RED .moai/reports/t1234/M4-red-00{5,6}.txt | leaf model: opus | fold-at: 2026-09-26T13:45+09:00

M1 repair (card, plan.md D10): AC-CONTRACT-001=PASS, AC-CONTRACT-004=PASS | RED de8aee456, GREEN 200683e45 | RED .moai/reports/t1234/M1R-red.txt | leaf model: opus

M3: AC-CONTRACT-008=PASS, AC-CONTRACT-021=PASS, AC-CONTRACT-023=PASS (AC-001..005 re-verified PASS) | RED b7f9cc44b, GREEN 7b0d20494 | evidence: .moai/state/verify/t1234-run/M3.* + lead.M3.* (lead re-run, all 8 exit=0 passlines=1 bad=1); hook pin TestFrozenInstructionFilesPinnedInContract PASS; AC-013 shell block list/deps/nonempty/noexecnet/noimport all exit=0; internal/contract coverage 97.7% | RED .moai/reports/t1234/M3-red.txt | leaf model: opus | SPEC v0.5.2 wording commit 25283ebf8 landed mid-milestone (text only) | fold-at: 2026-09-26T14:30+09:00

M2: AC-CONTRACT-019=PASS (config sub-cases; the sign-behavior clause is exercised in M5/M6), AC-CONTRACT-020=PASS | RED 47408a664, GREEN cb67f353c | evidence: .moai/state/verify/t1234-run/M2.* + lead.M2.* (lead re-run exit=0 bad=1 both); AC-020 shell checks: template scan exit=1 (no line), `^ *decider:` exit=1 (no line), local grep printed exactly 2 lines (228 batch_sign, 231 push_develop); make build exit 0; full ./internal/config/... and ./internal/template/... exit 0 | guard data: shipped_key_inventory.yaml +8 keys, configCacheSchemaVersion 6→7 | RED .moai/reports/t1234/M2-red-0{19,20}.txt | leaf model: opus | fold-at: 2026-09-26T14:50+09:00

M5: AC-CONTRACT-016=PASS (20 sub-cases a–t), AC-CONTRACT-019=PASS (adds the deferred sign clause: decider jev → human path signs, receipt path refuses kickoff_decider_jev_sole) | RED a5d8f2b24, GREEN 17c3fb69a | evidence: .moai/state/verify/t1234-run/M5.* + lead.M5.* (lead re-run exit=0 bad=1 both); coverage internal/contract 97.7%, internal/contract/sign 89.5%; -race ok; AC-013 shell block all exit=0 | RED .moai/reports/t1234/M5-red-01{6,9}.txt | leaf model: opus | fold-at: 2026-09-26T15:25+09:00

M6: AC-CONTRACT-007, 009, 010, 011, 012, 014, 015, 017, 018, 022, 024, 025=PASS (internal/cli); AC-CONTRACT-013=PASS (internal/contract) | RED ac12012a8, GREEN 628325f75 | evidence: .moai/state/verify/t1234-run/M6.*; RED .moai/reports/t1234/M6-red.txt; AC-013 observed failure by a reverted, uncommitted mutant (LoadDir writing a file) .moai/reports/t1234/M6-ac013-mutant.txt — the read-only behavior pre-existed, so AC-013's RED is a mutant observation, not a pre-implementation commit | leaf model: opus | fold-at: 2026-09-26T16:05+09:00

Final lead verification (non-author re-execution, HEAD 628325f75): all 25 AC pass checks exit=0, exactly one `--- PASS:` line, bad=1 — matrix at .moai/reports/t1234/final-ac-matrix.txt, logs .moai/state/verify/t1234-run/final.*; go vet exit 0; golangci-lint (contract, config, cli, spec, template, hook) `0 issues.`; windows cross-build exit 0; coverage internal/contract 97.7%, internal/contract/sign 89.5%; -race ok; AC-013 block list/deps/nonempty/noexecnet/noimport all exit=0; AC-022 diff empty; `moai spec lint SPEC-AUTONOMY-CONTRACT-001` → no findings.

Peer cross-validation (read-only agent, not an author; logs .moai/state/verify/t1234-run/peer/): 25/25 §A pass checks hold; sub-case review 23 PASS, 2 PARTIAL (AC-014: terminal message not asserted in the `</dev/null` sub-case; AC-025: "both unset" not asserted). Closed by test-only commit 1dce24707 (mutant evidence .moai/reports/t1234/peerfix-mutant.txt); lead re-run AC-014 and AC-025 exit=0 passlines=1 bad=1.

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready
- run_head: 1dce24707 (branch WT-contract-schema; not pushed, not merged into develop)
- ACs: 25/25 PASS on the acceptance.md §A convention (final matrix .moai/reports/t1234/final-ac-matrix.txt)
- Milestone commits (RED → GREEN): M1 b93af204a → f40dea185; M4 acb641d21 → 086dfb2fd; M1 repair (card) de8aee456 → 200683e45; M3 b7f9cc44b → 7b0d20494; M2 47408a664 → cb67f353c; M5 a5d8f2b24 → 17c3fb69a; M6 ac12012a8 → 628325f75; peer-gap test fix 1dce24707. Execution order was M1, M4, M1 repair, M3, M2, M5, M6: M2/M3/M5 waited for the v0.5.1 decider decision, and M4 is decider-independent.
- Deviations: AC-013's RED is a reverted mutant observation (behavior existed by construction); CLI reads the registry at `<root>/.claude/rules/moai/core/zone-registry.md` or `MOAI_CONSTITUTION_REGISTRY` rather than `ResolveRegistryPath` (worktree containment); test-support package `internal/contract/sign/signtest` added (0% own coverage, imported only by tests).
- Gaps: full `internal/cli` suite not run locally (machine-load rule; CI is the verdict); no pty-backed TTY test (the TTY seam is exercised with pipes); awk parity measured with darwin `/usr/bin/awk` only; `/compact` fold not available in the coordinator subagent context (fold rows recorded here instead).
- Next: sync phase (manager-docs) — CLI docs for `moai contract sign|show|verify` and the `workflow.autonomy` keys.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- Input parameters: tier L; scope ~18 files (internal/contract, internal/contract/sign, internal/cli, internal/config, internal/spec, internal/hook, internal/template + local config); domains 3+ (schema/integrity, CLI/TTY, config/template); language mix Go + YAML; concurrency benefit LOW (coding-heavy, milestones sequentially dependent).
- Mode evaluation: `direct` not selected (non-trivial); `fanout` not selected (coding-heavy, single worktree, one writer); `sweep` not selected (not mechanical/uniform); `serial` selected.
- Decision: serial — delegated to manager-lead (Role A entry: 6 milestones, >=10 files, cross-domain), which runs one write-capable leaf worker at a time in this worktree.
- Justification: milestones M1→M6 are data-dependent (types → config → verify core → acceptance binding → signing → CLI); one worktree admits one writer; Anthropic's coding-task parallelism caveat favours sequential workers.
- Kickoff: Implementation Kickoff Approval granted by the operator in the lane window before run-phase entry.
