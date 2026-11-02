# Plan — SPEC-AUDIT-ENGINE-CONVERGENCE-001

## §A Context

The operator decided on 2026-11-03 to repair six defects in the audit engine; the leader relayed the product rulings as R287 and R288. This plan also carries the round-1 plan-audit repair (`.moai/reports/t2017/plan-audit-1.md`), and per leader ruling R290 the card is in the rc.39 integration candidate.

- Requirements: `spec.md` §C. Reconciliation with completed SPECs: `spec.md` §G.
- Acceptance: `acceptance.md`.
- Decisions: `decision-index.md`.
- Measured trees: `cb30e05b3e5f64774b8dabd87e80c24332f4551d`, and `36148aeec99fd8d80e76645432f5a58802732c3e` for the round-4 entries (same code).

## §B Tier classification

**Tier M, at both ceilings.**

| Measure | Value | Tier M ceiling | Note |
|---|---|---|---|
| Requirements | 16 (REQ-AEC-001..016) | 16 | at the ceiling |
| Acceptance criteria | 6 logical ACs, 16 sub-criteria (AC-AEC-001a..006) | 16 | at the ceiling, counting every sub-criterion |
| Production files touched | about 18 | — | see milestones |
| All files touched | about 34 | 5-15 guidance | over the file guidance |
| Estimated LOC | 1000-1400 | 300-1000 | over the upper bound |

Reasoning: both budgets are counted independently and both sit exactly at 16. The round-1 audit repair added no requirement and no criterion. Every repair is a condition inside an existing requirement or a fixture inside an existing criterion, and no pinned case was dropped. The file and LOC estimates are over the Tier M guidance; about half of the files are tests, template mirrors, and rendered agent copies.

**B4 stays in this SPEC (leader rulings R288, R300).** Any new requirement or criterion would push the SPEC over a ceiling and is reported to the leader rather than added. The only way out the leader has named is a B4 split (REQ-AEC-011..013, AC-AEC-004a..c), issued by the leader as a separate SPEC and card. That would leave this SPEC at 13 and 13. Re-tiering is ruled out.

## §C Pre-flight

- Re-read `spec.md` §B before touching any cited line; the run tree may have moved.
- Fix-confirm re-audits for this card are limited to 3, each scoped to the diff since the previous verdict's `audited_sha`, passed as `delta_base` explicitly. Until M1 lands there is no chain gate, so such a delta is a t1708 delta, not a chain-verified one.
- Verification runs scoped tests only, for example `go test ./internal/cli/... -run '^(TestName1|TestName2)$'` with every branch anchored. The same applies to `./internal/auditreceipt/...`, `./internal/auditverdict/...`, `./internal/runtime/...`, `./internal/closure/...`, `./internal/homestate/...`, `./internal/hook/...`, and `./internal/template/agentemit/...`. The full suite runs in CI after push. Take a slot lease before the `internal/cli` suite where other lanes may run it.

## §D Constraints

- Forbidden (REQ-AEC-016):
  - a concurrently writing session or agent added to the tree;
  - replacing the installed binary;
  - weakening any refusal;
  - deleting any check.
- RED first per item: each milestone opens with the failing test named in its AC, committed before the fix.
- Template-First:
  - A file with a mirror under `internal/template/templates/` changes in the same commit as the mirror, with the neutrality strip: no internal SPEC IDs, REQ tokens, card ids, dates, or SHAs in the mirror.
  - An agent definition is edited in its source under `internal/template/agentemit/source/`, then `make agents-emit` regenerates the copies it renders. For the plan-auditor and sync-auditor those are the template sets only; their four local definitions (`.claude/agents/moai/{plan,sync}-auditor.md` and `.codex/agents/moai/{plan,sync}-auditor.toml`) are not emitter targets (`internal/template/agentemit/sets.go:48-50`, `:62-63`) and are updated by hand in M6 (leader rulings R308, R313).
  - Then `make build`, which runs `agents-emit-check`, exits 0, and `git status --short` shows no unexpected tracked change.
- Push happens only at sync, through the PR edge.

## §E Self-verification

Each milestone closes with its AC's green cell observed, verbatim, in `progress.md` §E.2 (owned by manager-develop).

## §F Milestones (ordered: highest decision-reversibility first)

### M1 — B1 delta chain gate, the delta record, and its persistence (highest-risk decision)

RED: AC-AEC-001c, 001d, 001e, 001f.
Files:
- `internal/cli/audit_delta_chain.go` (new):
  - reads the card's second-review records;
  - checks every record, the starting record included, for a pass or fail from every required backend, iterating over the required set rather than the entries present, so an absent entry counts as not run;
  - cross-checks each record's receipt against the receipt store (tool, tree root, head, empty `gate_unmet`, codex verdict not inconclusive);
  - builds the closure set as the union of `blocking_defect:` ids over every FAIL verdict file tied to a chain record, and checks `blocking_count` on each;
  - treats an id as inherited when the record that extended its FAIL file lists it among its confirmed ids, deciding this per naming rather than per id: when a FAIL file tied to a later chain record names the id again, the earlier inheritance is void, and the id is inherited only if the record extending that latest naming confirmed it; it checks every id not inherited for a closing commit inside `delta_base..HEAD` that comes after the id's latest naming.
- `internal/cli/mcp_convergence.go`:
  - change the record branch at `:1059` so an admitted delta files a record;
  - change the persistence condition at `:1046` so an admitted delta persists whatever its verdict, while a `card_id`-less delta stays fail-only (t1873);
  - pass the confirmation to every leg.
- `internal/cli/mcp_audit_multi_record.go` and `internal/closure/records.go` — additive fields: delta marker, delta base, extended receipt id, confirmed ids, relied-on verdict-file digests. `SelectSecondReview` logic stays unchanged.
- `internal/cli/mcp_audit_multi.go`, `internal/cli/mcp_server.go` — the confirmation input and its schema description; an explicit `delta_base` with a `card_id` goes through the gate.
- `internal/auditverdict/verdict.go` — parse `blocking_defect:` lines.
- Tests: `internal/cli/audit_delta_chain_test.go`, `internal/cli/mcp_audit_multi_record_test.go`, a persistence test beside the existing t1873 test.

### M2 — B4 operator approval path (security boundary)

RED: AC-AEC-004a..c.
Files:
- `internal/auditreceipt/store.go` — one function that removes exactly one named receipt-kind rejection of one tree and appends one approval record (an append-only log inside `.moai/state/audit-receipts/`).
- `internal/cli/verify_clear_rejection.go` (new) — the verb under `moai verify`. It reuses `contractAgentMarkers()` (`internal/cli/contract.go:76-78`), the terminal seam, and the typed confirmation of the contract signing path.
- `internal/cli/verify.go` — register the subcommand.
- Tests: `internal/cli/verify_clear_rejection_test.go` (including the set-equality test against `contractAgentMarkers()`), `internal/auditreceipt/store_approval_test.go`.

### M3 — B1 default proposal and the delta codex leg

RED: AC-AEC-001a, 001b.
Files:
- `internal/cli/mcp_audit_multi.go` — the default proposal from the newest FAIL verdict file; the `"full"` opt-out.
- `internal/cli/audit_delta_scope.go` — remove the prose-only path for the leg.
- `internal/cli/mcp_convergence.go` — while a delta is in effect, `performCodexAudit` keeps its adversarial `turn/start` and embeds the server-collected `delta_base...HEAD` diff in the turn. It does not switch to the native target.
- Reuse the newest-verdict selection used by `factory rebind-evidence` (`internal/cli/factory_card_rebind.go`) rather than writing a second reader.
- Tests: `internal/cli/audit_delta_scope_test.go`, `internal/cli/mcp_audit_multi_test.go`.

### M4 — B2 classified unrun legs, the INCONCLUSIVE label and every consumer

RED: AC-AEC-002a..c.
Files:
- `internal/cli/mcp_convergence.go`:
  - `error_class` on `PerBackendVerdict`;
  - head commit, tree digest, and scope fields on the persisted result;
  - re-running only the named legs.
- `internal/cli/mcp_codex_audit.go`, `internal/cli/mcp_glm.go`, `internal/cli/mcp_claude.go` — map each could-not-run path to a class.
- `internal/cli/mcp_audit_multi.go`, `internal/cli/mcp_server.go` — the legs-to-re-run input.
- `internal/auditreceipt/store.go`:
  - the receipt carries `error_class`;
  - the Stop-hook last-line grammar at `:173` accepts `INCONCLUSIVE`.
- `internal/hook/audit_receipt_guard.go` — an `INCONCLUSIVE` last line is a parsed line: no `CauseVerdictLineMissing`, no clear, stage entry blocked with the class.
- `internal/homestate/card_evidence_readers.go` — the verdict-line pattern at `:20` accepts `INCONCLUSIVE` and holds the card.
- `internal/auditverdict/verdict.go`:
  - `convergence_overall: inconclusive`;
  - the 3-token `required_backend` line;
  - the new refusal step and its reason, in the order of `spec.md` §G;
  - `AdmitLabel` unchanged.
- `internal/runtime/audit_counter.go` and `internal/runtime/audit_ceiling.go` (`CountPlanAuditRounds`, `:1186-1191`) — both skip `INCONCLUSIVE` iteration files and keep counting FAIL. `Latest` and `SelectLatestVerdict` still select an `INCONCLUSIVE` file.
- `.moai/docs/audit-artifact-convention.md` + `internal/template/templates/.moai/docs/audit-artifact-convention.md` (same commit) — the label vocabulary, the `required_backend` line shape, the `required_backend_fail:` producer rule, the `blocking_defect:` line on plan and sync FAIL files, and the consumer rules. The `blocking_defect:` line is defined here because M1 consumes it; if M1 lands first, its tests use fixtures.
- Tests in the matching `_test.go` files, one per consumer.

### M5 — B3 receipt binding across evidence-only commits

RED: AC-AEC-003a, 003b.
Files:
- `internal/auditreceipt/store.go`:
  - the new receipt field for the uncommitted state without the head;
  - the cited SPEC ID passed into `CheckCitedReceiptsSince` (`:1514`) from the citing line's `spec=`;
  - the ancestry, filename-family, and `progress.md` checks in the reuse arm at `:1578-1616`.
  The `§E.2` section is computed on both versions and must be byte-identical, with its heading unchanged and in place. Changes under `§E.1`, `§E.3`, and `§E.4` are allowed (leader ruling R288). Only the cited SPEC's `progress.md` qualifies.
- `internal/hook/audit_receipt_guard.go` — pass the parsed `spec=` value to the check.
- `internal/auditreceipt/store_binding_range_test.go` (new).

### M6 — B5 agent sources and A1 rule paragraph (mechanical text)

RED: AC-AEC-005, AC-AEC-006.
Files:
- `internal/template/agentemit/source/sync-auditor.md` (lines 101, 201, 203, 210) and `internal/template/agentemit/source/plan-auditor.md` (lines 193, 204, 251, 253, 260):
  - the narrowed gate statements;
  - `INCONCLUSIVE` in both machine-line templates;
  - the `blocking_count` and `blocking_defect:` obligation on every FAIL, which the sync-auditor source gains for the first time.
- `make agents-emit`, which regenerates the template copies of the two agents (`internal/template/templates/.claude/agents/moai/{plan,sync}-auditor.md` and `internal/template/templates/.codex/agents/moai/{plan,sync}-auditor.toml`). It does not write the local definitions: the emitter's project targets are limited to `CarriedFour` (`internal/template/agentemit/sets.go:62-63`), and the comment at `:48-50` says the other project copies stay hand-managed.
- The four local definitions, updated by hand to match the corrected source text (leader rulings R308, R313):
  - `.claude/agents/moai/plan-auditor.md` (wrong lines 251, 253) and `.claude/agents/moai/sync-auditor.md` (201, 203);
  - `.codex/agents/moai/plan-auditor.toml` (248, 250) and `.codex/agents/moai/sync-auditor.toml` (190, 192).
  Each gets the same edit set as its source: the three wrong sentences removed (two lines per file), the positive statements, `INCONCLUSIVE` in both machine-line templates, and the `blocking_count` / `blocking_defect:` obligation. In the TOML files the text lives inside the `developer_instructions = '''` literal string (`sync-auditor.toml:12-228`, `plan-auditor.toml:10-772`), so the edit must not introduce a `'''` sequence. The local plan-auditor Markdown keeps its pre-existing line 64 and 150 differences from the template copy (R287).
- Sources, template copies, and the four local definitions are committed together.
- `internal/cli/mcp_convergence_test.go` — the FAIL-plus-inconclusive convergence test; `internal/cli/audit_gates_failclosed_test.go` (where `applyGateUnmet` is already exercised) — the two `applyGateUnmet` tests.
- `.moai/docs/audit-artifact-convention.md` + mirror — the A1 paragraph (may share the M4 commit).
- Then `make build`; confirm `git status --short` shows only the intended files.

## §G Risks

| Risk | Effect | Mitigation |
|---|---|---|
| A delta chain passes a card whose pre-existing defect sits outside every delta, and the full review missed it | a defect lands | stated in spec §E B1; the chain must start at a full-scope record that ran every required backend |
| A forged second-review record plus receipt pair | a fake chain admits a delta | the receipt cross-check raises the cost; stated as residual (R288 threat model) |
| A chain-admitted delta PASS opens the Stop gate | intended by R287 Q1 | only behind the chain gate; a `card_id`-less delta keeps the t1873 rule (AC-AEC-001f) |
| One of the `INCONCLUSIVE` consumers is missed | an unrun audit is read as a missing line, or counted as a round | AC-AEC-002b asserts each consumer separately; §B lists all of them |
| The completed-SPEC projection and the new label disagree | the admission reason stops at "overall fail" and never names the class | spec §G sets `convergence_overall: inconclusive` and the refusal order; AC-AEC-002b pins the reason |
| The `error_class` mapping misreads codex text | a quota failure is labelled `other` | `other` still refuses; AC fixtures pin the three named classes |
| B4 bypassed by an agent Bash call, a Codex session, or direct file deletion | a rejection cleared without the operator | stated as residual in spec §E: an operator-mistake guard, not a boundary |
| Template agent copies edited by hand | `make build` fails at `agents-emit-check` | M6 edits the sources and regenerates the template copies |
| The hand-updated local definitions drift from the sources | a local auditor keeps or regains a wrong sentence, and no check sees it | AC-AEC-005 checks the four local files at M6; later drift is residual (spec §E B5) |
| Only the `.claude` copies are fixed | the Codex lanes keep the wrong gate text | AC-AEC-005 and EV-20 name all four local files |
| A section-boundary check that drifts to whole-file, new-version-only, or §E.3-stale | a rewrite of the evidence is inherited, or the normal flow goes stale | AC-AEC-003a/003b pin each direction |
| Any further addition | a ceiling exceeded | report to the leader; only the leader splits B4 (R300) |

## §H Anti-patterns to avoid

- Grepping for the fixed text as the only proof of a fix; every AC pairs the text check with a behavior test.
- Building the delta chain from verdict files or call arguments.
- Clearing a rejection through any path other than the one verb.
- Folding the head commit back into the new uncommitted-state field.
- Editing a template agent copy instead of its source.
- Assuming `make agents-emit` updates the local plan-auditor and sync-auditor definitions, in `.claude` or in `.codex`.

## §I Cross-references

- `spec.md` §B (verified citations), §D (A1/B4 tension), §E (residual risk and design choices), §G (reconciliation with completed SPECs)
- `decision-index.md` (Q1-Q8, leader rulings R287, R288)
- `.moai/reports/t2017/plan-audit-1.md` (round-1 audit)
- `.claude/rules/moai/development/verification-completeness.md` §2 (two-cell adoption)
- `.moai/docs/audit-artifact-convention.md`
