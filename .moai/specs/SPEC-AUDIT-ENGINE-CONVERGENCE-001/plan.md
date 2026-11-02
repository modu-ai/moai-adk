# Plan — SPEC-AUDIT-ENGINE-CONVERGENCE-001

## §A Context

Six repairs to the audit engine, decided by the operator on 2026-11-03, with the product rulings relayed by the leader as R287. Requirements: `spec.md` §C. Acceptance: `acceptance.md`. Decisions: `decision-index.md`. Measured tree: `cb30e05b3e5f64774b8dabd87e80c24332f4551d`.

## §B Tier classification

**Tier M, at both ceilings.**

| Measure | Value | Tier M ceiling | Note |
|---|---|---|---|
| Requirements | 16 (REQ-AEC-001..016) | 16 | at the ceiling |
| Acceptance criteria | 6 logical ACs, 16 sub-criteria (AC-AEC-001a..006) | 16 | at the ceiling, counting every sub-criterion |
| Production files touched | about 12 | — | see milestones |
| All files touched | about 24 | 5-15 guidance | over the file guidance |
| Estimated LOC | 800-1100 | 300-1000 | at or slightly over the upper bound |

Reasoning: both budgets are counted independently and both sit exactly at 16. R287's additions to B1 (the chain gate and the delta record) were absorbed by folding the old B1 fallback requirement into REQ-AEC-001 and by merging the B4 negative sub-criteria, not by exceeding a ceiling. The file and LOC estimates are at or over the Tier M guidance. About half of the files are tests and template mirrors.

**B4 stays in this SPEC (leader ruling R288).** The operator decision bundles B1-B5. Any new requirement or criterion would push the SPEC over a ceiling and is reported to the leader rather than added. The only split the leader has named is B4 (REQ-AEC-011..013, AC-AEC-004a..c), and only if the plan audit raises server-side BLOCKING-id storage as BLOCKING; that split is the leader's act, not this plan's. B4 shares no code with B1-B3, so the cut would be clean.

The R288 positive `§E.3` case fits within the existing sub-criteria: it is a second fixture inside AC-AEC-003a, and the single-line `§E.2` negative tightens AC-AEC-003b. No pinned case was removed.

## §C Pre-flight

- Re-read `spec.md` §B before touching any cited line; the run tree may have moved.
- Fix-confirm re-audits for this card are limited to 3, and each is scoped to the diff since the previous verdict's `audited_sha`. Pass that SHA as `delta_base` explicitly until M3 lands the default.
- Verification runs scoped tests only, for example `go test ./internal/cli/... -run '^(TestName1|TestName2)$'` with every branch anchored. The same applies to `./internal/auditreceipt/...`, `./internal/auditverdict/...`, `./internal/runtime/...`, `./internal/closure/...`, and `./internal/hook/...`. The full suite runs in CI after push. Take a slot lease before the `internal/cli` suite where other lanes may run it.

## §D Constraints

- Forbidden: a concurrently writing session or agent added to the tree, replacing the installed binary, weakening any refusal, deleting any check (REQ-AEC-016).
- RED first per item: each milestone opens with the failing test named in its AC, committed before the fix.
- Template-First: a file with a mirror under `internal/template/templates/` changes in the same commit as the mirror, with the neutrality strip (no internal SPEC IDs, REQ tokens, card ids, dates, or SHAs in the mirror). Then `make build` exits 0, and `git status --short` shows no unexpected tracked change.
- Push happens only at sync, through the PR edge.

## §E Self-verification

Each milestone closes with its AC's green cell observed, verbatim, in `progress.md` §E.2 (owned by manager-develop).

## §F Milestones (ordered: highest decision-reversibility first)

### M1 — B1 delta chain gate and the delta second-review record (highest-risk decision)

RED: AC-AEC-001c, 001d, 001e, 001f.
Files:
- `internal/cli/mcp_convergence.go`:
  - change the second-review branch at `:1059` so a delta admitted under REQ-AEC-004 files a record;
  - pass the BLOCKING-closure confirmation to every leg.
- `internal/cli/mcp_audit_multi_record.go` — the delta record: scope base equals the delta base, a delta marker, and the receipt id of the chain record it extends (additive JSON fields).
- `internal/closure/records.go` — the additive fields on `SecondReviewRecord` and `SecondReviewScope`. The `SelectSecondReview` logic stays unchanged.
- `internal/cli/mcp_audit_multi.go`, `internal/cli/mcp_server.go` — the BLOCKING-closure confirmation input and its schema description.
- `internal/cli/audit_delta_chain.go` (new) — reads the card's second-review records, cross-checks each against the receipt store, and evaluates the confirmation against the newest FAIL verdict's `blocking_defect:` lines.
- `internal/auditverdict/verdict.go` — parse `blocking_defect:` lines.
- Tests: `internal/cli/audit_delta_chain_test.go`, `internal/cli/mcp_audit_multi_record_test.go`.

### M2 — B4 operator approval path (security boundary)

RED: AC-AEC-004a..c.
Files:
- `internal/auditreceipt/store.go` — one function that removes exactly one named receipt-kind rejection of one tree and appends one approval record (an append-only log inside `.moai/state/audit-receipts/`).
- `internal/cli/verify_clear_rejection.go` (new) — the verb under `moai verify`, reusing the `internal/contract/sign` human-path checks.
- `internal/cli/verify.go` — register the subcommand.
- Tests: `internal/cli/verify_clear_rejection_test.go`, `internal/auditreceipt/store_approval_test.go`.

### M3 — B1 default proposal and mechanical codex scope

RED: AC-AEC-001a, 001b.
Files:
- `internal/cli/mcp_audit_multi.go` — the default proposal from the newest FAIL verdict file; the `"full"` opt-out.
- `internal/cli/audit_delta_scope.go` — remove the prose-only path for the leg.
- `internal/cli/mcp_convergence.go` — when a delta is in effect, the codex leg uses the native review target (as `codex_audit` native already does at `mcp_codex_audit.go:2323-2330`) or the collected diff.
- Reuse the newest-verdict selection used by `factory rebind-evidence` (`internal/cli/factory_card_rebind.go`) rather than writing a second reader.
- Tests: `internal/cli/audit_delta_scope_test.go`, `internal/cli/mcp_audit_multi_test.go`.

### M4 — B2 classified unrun legs, the INCONCLUSIVE label and its consumers

RED: AC-AEC-002a..c.
Files:
- `internal/cli/mcp_convergence.go` — `error_class` on `PerBackendVerdict`; re-running only the named legs.
- `internal/cli/mcp_codex_audit.go`, `internal/cli/mcp_glm.go`, `internal/cli/mcp_claude.go` — map each could-not-run path to a class (Claude's existing `Provenance.ErrorCode` values map; codex and GLM gain a mapping).
- `internal/cli/mcp_audit_multi.go`, `internal/cli/mcp_server.go` — the legs-to-re-run input.
- `internal/auditreceipt/store.go` — the receipt carries `error_class`.
- `internal/auditverdict/verdict.go` — parse `INCONCLUSIVE` and the class; admission refuses it and names the class.
- `internal/runtime/audit_counter.go` — skip `INCONCLUSIVE` iteration files; keep counting FAIL.
- `.moai/docs/audit-artifact-convention.md` + `internal/template/templates/.moai/docs/audit-artifact-convention.md` — the label vocabulary, the `required_backend` line shape, the `blocking_defect:` line, and the consumer rules (same commit). The `blocking_defect:` line is defined here because M1 consumes it; if M1 lands first, its tests use fixtures.
- Tests in the matching `_test.go` files.

### M5 — B3 receipt binding across evidence-only commits

RED: AC-AEC-003a, 003b.
Files:
- `internal/auditreceipt/store.go`:
  - the new receipt field for the uncommitted state without the head;
  - the ancestry, filename-family, and `progress.md` section checks in the reuse arm at `:1578-1616`. The section check refuses only a changed line under `§E.2`; changes under `§E.1`, `§E.3` and `§E.4` are allowed (leader ruling R288).
- `internal/auditreceipt/store_binding_range_test.go` (new).

### M6 — B5 prompt corrections and A1 rule paragraph (mechanical text)

RED: AC-AEC-005, AC-AEC-006.
Files (each pair in one commit):
- `.claude/agents/moai/sync-auditor.md` lines 201, 203 + `internal/template/templates/.claude/agents/moai/sync-auditor.md`.
- `.claude/agents/moai/plan-auditor.md` lines 251, 253 + `internal/template/templates/.claude/agents/moai/plan-auditor.md`. Edit the matching lines only; the line-64 divergence stays.
- `.moai/docs/audit-artifact-convention.md` + mirror — the A1 paragraph (may share the M4 commit).
- Then `make build`; confirm `git status --short` shows only the intended files.

## §G Risks

| Risk | Effect | Mitigation |
|---|---|---|
| A delta chain passes a card whose pre-existing defect sits outside every delta and the full review missed it | a defect lands | stated in spec §E B1; the chain must start at a full-scope record |
| Forged second-review record plus receipt pair | a fake chain admits a delta | receipt cross-check raises the cost; stated as residual |
| The closure reader treats a delta record as full coverage | readiness reads a delta PASS as coverage | intended by R287 Q1; the record is filed only behind the chain gate |
| `INCONCLUSIVE` read by a consumer that knows three labels | treated as unknown | `AdmitLabel` admits only PASS-family, so unknown refuses; AC-AEC-002b pins it |
| `error_class` mapping misreads codex text | a quota failure labelled `other` | `other` still refuses; AC fixtures pin the three named classes |
| Pseudo-terminal forgery by a local process | B4 boundary bypassed | stated as residual; the hook deny is a follow-up Gap |
| A section-boundary check that drifts to a whole-file or `§E.3`-stale rule | a normal `§E.3` rewrite goes stale, or an `§E.2` evidence rewrite is inherited | AC-AEC-003a pins the `§E.3` acceptance and AC-AEC-003b the single-line `§E.2` refusal |
| Any further addition | a ceiling exceeded | report it to the leader; B4 stays in this SPEC (R288), and only the leader splits it |

## §H Anti-patterns to avoid

- Grepping for the fixed text as the only proof of a fix; every AC pairs the text check with a behavior test.
- Building the delta chain from verdict files or call arguments.
- Clearing a rejection through any path other than the one verb.
- Folding the head commit back into the new uncommitted-state field.

## §I Cross-references

- `spec.md` §B (verified citations), §D (A1/B4 tension), §E (B1 residual risk, B2 counter, B3/B4/A1 choices)
- `decision-index.md` (Q1-Q8, leader ruling R287)
- `.claude/rules/moai/development/verification-completeness.md` §2 (two-cell adoption)
- `.moai/docs/audit-artifact-convention.md`
