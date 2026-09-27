# progress.md — SPEC-AUTONOMY-GATE-REWIRE-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-26
- spec_version: 0.3.6 (2026-09-27, plan-audit iter-6 final FAIL 0.868 → lead decision PASS-WITH-DEBT; D56 narrow fix + shell measurability probe, D57-D59 fixed — .moai/reports/t1236/verdict.md); prior 0.3.5 (2026-09-27, plan-audit iter-5 delta FAIL 0.841 repaired in place — D47-D55, plan.md §L; report .moai/reports/t1236/plan-audit-5.md); prior 0.3.4 (2026-09-27, M0 re-anchor per lead decisions B1-B5 — plan.md §K, research.md §10.6; REQ 25 / AC 25 unchanged); prior 0.3.3 (operator-approved one-time 4th audit exception limited to D43-D46; base 710530d67); prior 0.3.2 (revision base 1b071a573 = v0.3.1; plan-audit iter-2 FAIL 0.77 → `.moai/reports/t1236/plan-audit-2.md`, dispositions D26-D42 in plan.md §J; iter-1 dispositions in plan.md §I)
- tier: L (spec.md, plan.md, acceptance.md, design.md, research.md)
- requirements: 25 (REQ-GR-001..025, contiguous, no new IDs in v0.3.2) / acceptance criteria: 25 (AC-GR-001..025)
- a1_reference_baseline: 25283ebf8 (SPEC-AUTONOMY-CONTRACT-001 0.5.2, schema owner; conditional interim-rule wording, A3 owns the AC-CONTRACT-016 (t) replacement test); history 8f77d9a33 / 98cb7879d / 4208a3a3b / 652243c72 / 6d98ca466 / 67a2f55cb / 65e0a9167
- a2_reference: A2 0.4.3 (implemented, BASE 7fe658815) — escalation/<class>-<fingerprint>.md + YAML header + revoke kind confirmed at M0; revoke records are status: resolved in A2, so resume blocking is owned by the A3 revoke reader (v0.3.4, B3); earlier withdrawn JSON format d8926ff9a is history
- a2b_reference: t1245 owns push serialization, agent-origin sign deny, decide refusal in MOAI_FACTORY_ROLE=agent sessions (SPEC not yet written; marker kept)
- store: $MOAI_HOME/db/<project-key>/contract/{receipts.jsonl,events.jsonl} (lead decision R10)
- deciders: guided=human; contract autonomous = llm (default) | llm+jev; jev refused (R5); outcome in {approve, reject, human} derived by A3 rules R1-R5; reject and human both = sign refused, human decision required
- linkage (REQ-GR-025): Jev-principle amendment (incl. CLAUDE.local.md §29) + R3 release + A1 interim-rule release in ONE commit via single constant jevDoctrineAmended; assembly per lead decision 2026-09-26 (D32): manager-spec then manager-develop, lane orchestrator commits by explicit pathspec; whole commit held without operator §29 confirmation (D33); proposed §29 text in design.md §11.1
- open_clarifications: none (defaults D-2, D-4 recorded; operator may override)
- lead_confirmations_pending: A1 acceptance of the jevDoctrineAmended gating of signer step (1) (marker, research.md §10.4); a producer of `plan_artifact_hash:` for card-path plan-audit reports before M8 (plan.md §C)
- run_preconditions: t1234 (A1 >= 0.5.2), t1235 (A2, with final escalation format), t1245 (A2b) merged into develop; t1175 merged and absorbed; BASE recorded in §E.2; pre-flight: re-measure reject/human consumers after the A2b SPEC lands
- operator_confirmation_29: approved 2026-09-26, text verbatim as design.md §11.1 (relayed by lead); edit lands only on the develop copy of CLAUDE.local.md in the card worktree, via the develop merge
- autonomous_kickoff_activation: see design.md §7.1 (single source)
- Implementation Kickoff Approval: not requested at plan phase

## §E.2 Run-phase Evidence

### M0 — absorb and re-anchor (2026-09-27, manager-develop) — HALTED with blocker report

- BASE: `7fe658815eb0d4110b9acadad56e5a85bee3ed3f` (`git merge-base develop HEAD`; last absorbed develop commit, t1175 close). Branch HEAD at measurement: `016a3965e` (merge of develop into WT-contract-gate-rewire). `MOAI_GR_BASE=7fe658815`.
- Preconditions (plan.md §C): A1 `SPEC-AUTONOMY-CONTRACT-001` status completed v0.5.2 (internal/contract present); A2 `SPEC-AUTONOMY-ESCALATION-001` status implemented v0.4.3 (internal/escalation present); A2b `SPEC-AUTONOMY-PRECONDITION-001` status completed v0.1.3 (internal/hook/contract_sign_guard.go carries the decide guard); t1175 `SPEC-ALWAYS-LOADED-DIET-002` status completed v0.10.0, absorbed.
- research.md §1.1 recount on this tree: local 33 files / 116 lines, template 32 files / 111 lines (identical to plan-phase); per-file counts identical to §1.2 table; Kickoff-line diff local↔template across the 32 paired files: 0. E/R/H classification unchanged.
- Anchors: all skill-file anchors (rows 6-14) present. Two cited anchors moved by t1175 — `askuser-protocol.md` `### The Five Exceptions` (now `askuser-protocol-reference.md:229`; `askuser-protocol.md` keeps only a 1-paragraph `## Ambiguity Triggers and Exceptions` stub at :206) and the `jev_ask` row of `moai-mcp-tools.md` (row now only in `moai-mcp-tools-catalogue.md:138,216`).
- A2 cross-check: record path/YAML frontmatter/`kind: revoke` match the lead-stated final format, BUT A2 spec §I.2 states a revoke record carries `status: resolved` and never makes a card needs-decision, and `escalation.NeedsDecision` counts only `contract`/`operational` kinds — contradicts REQ-GR-022/AC-GR-023 (revoke → needs-decision, "open record" via the A2 reader) and the REQ-GR-012 second-witness path.
- A1 cross-check: signer step (1) (interim rule) lives in `internal/contract/receipt.go` `ReceiptOutcome` (package `contract`, L208-231), not in `internal/contract/sign/` (design.md §2 row 23 allowlist).
- Constitution baseline: `moai constitution validate` on this tree → exit 1, `FAILED — 9 error(s)` (DRIFT CONST-V3R2-013/014/015/016/017 on CLAUDE.md §7, 033, 049, 152, 153); same with `go run ./cmd/moai`. AC-GR-003's second command (expects exit 0 `OK`) is red at BASE for reasons outside this SPEC's edit allowlist.
- Always-loaded sizes at BASE (`LC_ALL=en_US.UTF-8 wc -m`): CLAUDE.md 14039, askuser-protocol.md 18684, goal-directive.md 5628, moai-mcp-tools.md 3841; orchestration-mode-selection.md 37140.
- Disposition: M0 stop rule fired (cited anchors gone + A2/A1 differences). No M1+ work started; blocker report returned to the orchestrator for manager-spec re-delegation.

### M0 (resumed, v0.3.6 at `33d21cbfe`)

- Re-delegated after B1-B5 were fixed in the SPEC. Anchors of v0.3.4 (`askuser-protocol.md` stub paragraph before `## Free-form Circumvention Prohibition`; catalogue rows) confirmed present on this tree. BASE unchanged: `MOAI_GR_BASE=7fe658815` (read from the environment by every base test; never pinned in code except the BASE-derived sets listed as debt below).
- BASE caveat (lead, 2026-09-27): BASE is the pre-t1175-repair tree. BASE-derived artifacts committed by this run — regenerate on absorb: (1) `grBaseDriftIDs` in `internal/template/contract_mode_guided_test.go` (the EV-6 nine DRIFT ids, asserted as the base premise), (2) `grKickoffClasses` in the same file (research §1.2 E/R/H/L classification of the 33 base Kickoff documents).

### M9 RED — observed before any block existed

- How obtained: the two guard files were written first and run on the working tree while it still had zero contract-mode blocks and no SSOT (before M1-M5 edits were made); the blocks were inserted only after this run. Not a revert — the working tree was genuinely pre-block at that moment. Raw log: `.moai/state/verify/t1236/m9-red.txt` (gitignored evidence dir).
- Command: `MOAI_GR_BASE=7fe658815 go test ./internal/template/ -run '^(TestContractModeBlocksWellFormed|…|TestContractModeEmitterSites)$' -count=1 -v` → exit 1.
- Verbatim excerpt:

```text
    contract_mode_blocks_test.go:365: template tree carries zero contract-mode blocks — the guard swept nothing
--- FAIL: TestContractModeBlocksWellFormed (0.03s)
    --- FAIL: TestContractModeBlocksWellFormed/template-tree (0.03s)
--- FAIL: TestContractModeLocalTemplateParity (0.00s)
    --- FAIL: TestContractModeSSOTSections/ssot (0.00s)
    --- FAIL: TestContractModeLifecycleOrder/tree (0.00s)
    --- FAIL: TestContractModeLifecycleEvidence/ssot (0.00s)
    contract_mode_blocks_test.go:692: no blocks found — empty sweep
    --- FAIL: TestContractModeBlockCondition/tree (0.00s)
    --- FAIL: TestContractModeAuditRetryBlocks/tree (0.00s)
    --- FAIL: TestContractModeSyncBlocks/tree (0.00s)
    --- FAIL: TestContractModeSigningBlocks/tree (0.00s)
    contract_mode_guided_test.go:121: no block was stripped — empty sweep
    --- FAIL: TestContractModeGuidedPreservation/tree (1.93s)
    contract_mode_guided_test.go:616: emitter CLAUDE.md carries no contract-mode block
--- FAIL: TestContractModeEmitterSites (0.18s)
FAIL	github.com/modu-ai/moai-adk/internal/template	19.431s
```

- Falsifier subtests (known-bad fixtures) all FAILED the checker as required in the same run, e.g. `--- PASS: TestContractModeBlocksWellFormed/falsifier/forbidden-internal-token`, and AC-GR-003's two falsifiers:

```text
    contract_mode_guided_test.go:452: removed CONST-V3R2-001: drift 10 → findings [new constitution finding not present at the base: DRIFT CONST-V3R2-001 drift_count 10 > base 9]
    contract_mode_guided_test.go:465: observed: [.claude/rules/moai/core/askuser-protocol.md: unregistered [HARD] line not present at the base: "[HARD] Probe-only unregistered rule inserted by the falsifier."]
```

- Invariant guards green at RED time by design (no change yet): InheritedDivergence, ChangeSetAllowlist, ConstitutionDriftNotIncreased, AlwaysLoadedBudget.

### M1-M5 — SSOT and blocks

- M1: `contract-autonomy.md` (local + template, byte-identical). M2-M5: 20 blocks across 13 documents × 2 copies (ids per design.md §2).
- Cascade: `internal/template/catalog.yaml` moai-skill whole-tree hash regenerated (`go run ./internal/template/scripts/gen-catalog-hashes.go --all`, 1 line) — required by `TestManifestHashFormat`/`TestCatalogHashCoversSkillSubfiles`; not on design.md §2 allowlist, admitted in the allowlist test as a mechanical cascade (deviation reported to lead).
- GREEN: `.moai/state/verify/t1236/m9-green1.txt` — all 15 guard tests `--- PASS`; `go test ./internal/template/ -count=1` → `ok … 71.294s`.

### M6 — revoke, revoke reader, store chain

- RED (stub package returning zero values; `.moai/state/verify/t1236/m6-red.txt`), `go test ./internal/contract/revoke/ -run '^(TestRevoke|TestRevokeLeavesRepositoryUntouched)$' -count=1 -v` → exit 1:

```text
    revoke_test.go:108: revoke: status "" err <nil>
    revoke_test.go:174: status "" err <nil>, want not-signed
    revoke_test.go:200: err = <nil>, want integrity
    revoke_test.go:214: err = <nil>, want usage
    revoke_test.go:283: Blocked = false, err <nil>; want true, err false
    --- FAIL: TestRevoke/reader/r1_revoke_record (0.00s)
    --- FAIL: TestRevoke/reader/r4_status_open (0.00s)
    --- FAIL: TestRevoke/reader/r5_broken_header (0.00s)
--- FAIL: TestRevokeLeavesRepositoryUntouched (0.44s)
```

- CLI RED obtained by temporarily removing the `newContractRevokeCmd()` registration from `contract.go` (restored right after; `m6-cli-red.txt`): `contract_revoke_test.go:41: revoke: exit=1 reads=0` / `--- FAIL: TestContractRevoke`.
- GREEN: all 14 subtests + `TestRevokeLeavesRepositoryUntouched` `--- PASS` (`m6-green.txt`); `--- PASS: TestContractRevoke`. Lint (golangci-lint v2.1.6) on receipt/revoke/cli/template: `0 issues.`

### M7 — kickoff-check, decide, signing events, A1 conditioned in place

- Scope: `internal/contract/kickoff/` (new: Check, Decide, Jev seam, constants `autonomousKickoffEnabled = false`, `JevDoctrineAmended = false`); `internal/contract/receipt/` (event store verbs, `StoreDir` reproduction of the escalation resolver + parity test); `internal/contract/sign/` (signing records `sign-human` / `sign-receipt` / `reseal` before the contract write; the interim `llm+jev` refusal now follows the doctrine flag); `internal/contract/receipt.go` (`ReceiptOutcome(r, doctrineAmended)` — A1 step (1) conditioned in place, not duplicated in `sign/`); CLI `moai contract kickoff-check` / `moai contract decide` (`contract.go` lines 403 and 517).
- RED (stubs returning zero values; `.moai/state/verify/t1236/m7-red.txt`, `m7-cli-red.txt`) → exit 1:

```text
    check_test.go:147: pass=false reason="" reasons=[] state=, want pass=true reason=""
    decide_test.go:274: outcome "" reason "", want human "precondition:a"
    decide_test.go:278: Jev called 0 times, want 1
    decide_test.go:280: receipts 0→0 events 0→0, want +1/+1
    activation_test.go:138: llm signature under the compiled state: pass=false reason="", want inactive
--- FAIL: TestKickoffCheck
--- FAIL: TestDecidePreconditions
--- FAIL: TestDecideJevFallback
--- FAIL: TestSignRecordsEvent
--- FAIL: TestSignInterimRuleFollowsDoctrine
--- FAIL: TestEventStore
    contract_decide_test.go:95: help lacks "decide":
--- FAIL: TestContractDecide
--- FAIL: TestContractKickoffCheck
```

  `TestJevAmendmentLinkage` passed at RED by design (all markers absent is the consistent state).
- GREEN: `ok …/internal/contract/kickoff`, `ok …/internal/contract/receipt` (`m7-green1.txt`); `ok …/internal/cli` for `^(TestContractDecide|TestContractKickoffCheck|TestContractRevoke)$` (`m7-cli-green.txt`); guards with `MOAI_GR_BASE` set → `ok …/internal/template` (`m7-guards.txt`, 70 changed paths). `go build ./cmd/moai` and `GOOS=windows go build ./cmd/moai` exit 0. golangci-lint v2.1.6: `0 issues.`
- Mutant kills (D47/D48; mutate → run → restore, restore confirmed by `cmp`):
  - kickoff reader bypass → `check_test.go:147: pass=true reason="" reasons=[] state=signed-valid, want pass=false reason="revoked"` / `--- FAIL: TestKickoffCheck/17_revoke_record_only`; HEAD `ok …/internal/contract/kickoff`.
  - decide reader bypass → `decide_test.go:274: outcome "human" reason "jev-doctrine-not-amended", want human "precondition:e"` / `--- FAIL: TestDecidePreconditions`; HEAD `ok`.
  - unregistered `[HARD]` line → `contract_mode_guided_test.go:422: .claude/rules/moai/core/askuser-protocol.md: unregistered [HARD] line not present at the base: "[HARD] Probe-only unregistered rule inserted by the mutant probe."` / `--- FAIL: TestContractModeConstitutionDriftNotIncreased`; HEAD `ok …/internal/template`.
- Coverage (combined, `-coverpkg` receipt/revoke/kickoff over contract + cli tests; `cover-a3.txt`): total 81.8% of statements; per-package function average kickoff 79.6%, revoke 74.6%, receipt 86.0%. Below the 85% target for kickoff and revoke — recorded as a gap, not claimed.

### M7b / M8 — held

- M7b (doctrine amendment + `JevDoctrineAmended = true`, one commit): held on instruction. The CLAUDE.local.md §29 line is returned to the lead for operator confirmation; the SPEC-JEV-CORE-001 body change is returned as a manager-spec blocker. Nothing of M7b is in the tree (`TestJevAmendmentLinkage/tree` PASS with all markers absent).
- M8 (activation): held. design.md §7.1 row 6 is unmet — `orchestration-mode-selection.md` line 18 (`[ZONE:Frozen] [HARD] All Phase 4 execution modes are strictly downstream of Implementation Kickoff Approval …`) names no non-human signer. Row 4: `git merge-base --is-ancestor fb5901251 HEAD` → exit 0 (A2b merge is in this tree; its tests were not re-run here). Row 3 (A1 fallback-receipt signing) not independently re-verified in this run.

### M10 — AC matrix (tree `ec051a27b332a7e4bb3cee2715bb9762680092df`, HEAD `0e1f2edb9`, `MOAI_GR_BASE=7fe658815` where required)

Logs: `.moai/state/verify/t1236/ac/<AC>.txt`. Columns: exit / `--- PASS` count / `--- FAIL` / `--- SKIP` / `[no tests to run]`.

| AC | exit | PASS | FAIL | SKIP | no-tests | verdict |
|---|---|---|---|---|---|---|
| 001 | 0 | 3 | 0 | 0 | 0 | PASS |
| 002 | 0 | 2 | 0 | 0 | 0 | PASS |
| 003 | 0 | 6 | 0 | 0 | 0 | PASS |
| 004 | 0 | 1 | 0 | 0 | 0 | PASS |
| 005 | 0 | 3 | 0 | 0 | 0 | PASS |
| 006 | 0 | 3 | 0 | 0 | 0 | PASS |
| 007 | 0 | 3 | 0 | 0 | 0 | PASS |
| 008 | 0 | 7 | 0 | 0 | 0 | PASS |
| 009 | 0 | 16 | 0 | 0 | 0 | PASS |
| 010 | 0 | 1 | 0 | 0 | 0 | PASS |
| 011 | 0 | 3 | 0 | 0 | 0 | PASS |
| 012 | 0 | 3 | 0 | 0 | 0 | PASS |
| 013 | 0 | 3 | 0 | 0 | 0 | PASS |
| 014 | 0 | 3 | 0 | 0 | 0 | PASS |
| 015 | `make build` exit 0; `go test ./internal/template/... ./internal/contract/...` exit 0 (9 `ok`); cli exit 0, 12 PASS | | | | | PASS |
| 016 | 0 | 18 | 0 | 0 | 0 | PASS |
| 017 | 0 | 14 | 0 | 0 | 0 | PASS (order half + interim-rule test; linkage half passes in the all-absent state — the linkage-true state is deferred to M7b) |
| 018 | 0 | 14 | 0 | 0 | 0 | PASS |
| 019 | 0 | 18 | 0 | 0 | 0 | PASS |
| 020 | 0 | 7 | 0 | 0 | 0 | PASS |
| 021 | 0 | 10 | 0 | 0 | 0 | PASS |
| 022 | 0 | 0 | 0 | 0 | 2 | DEFERRED — empty selection; `TestJevDoctrineAmendment` belongs to M7b |
| 023 | 0 | 15 | 0 | 0 | 0 | PASS |
| 024 | 0 | 1 | 0 | 0 | 0 | PASS |
| 025 | 0 | 11 | 0 | 0 | 0 | PASS |

- `moai spec lint SPEC-AUTONOMY-GATE-REWIRE-001` (tree-built `bin/moai`) → `✓ No findings — all SPEC documents are valid`, exit 0. `go vet ./internal/contract/... ./internal/cli/ ./internal/template/` exit 0. `make build` exit 0, working tree clean afterwards.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: null            # run held at M7b / M8
run_commit_sha: 0e1f2edb9        # last run-phase code commit (M7)
run_status: held-m7b-m8
ac_pass_count: 24
ac_fail_count: 0
ac_deferred: [AC-GR-022, AC-GR-017-linkage-true-state]
new_warnings_or_lints_introduced: 0   # golangci-lint v2.1.6, go vet
cross_platform_build: {darwin: pass, windows: pass}
coverage_gap: {kickoff: 79.6, revoke: 74.6}   # function average, below 85
base_derived_debt: [grBaseDriftIDs, grKickoffClasses, EV-6 ids]  # BASE is the pre-t1175-repair tree — regenerate on absorb
m1_to_mN_commit_strategy: milestone-per-commit
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
