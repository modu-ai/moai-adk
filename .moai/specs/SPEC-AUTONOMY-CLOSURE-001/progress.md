# Progress — SPEC-AUTONOMY-CLOSURE-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_audit_iter1: "FAIL 0.77 — repaired D1-D18 in v0.2.0"
plan_audit_iter2: "FAIL 0.84 — repaired D19-D25 in v0.3.0"
plan_audit_iter3: "FAIL 0.87 — blocker D26 repaired in v0.3.1 (with D27, D28); operator-approved 4th-iteration exception (D26-scoped)"
plan_audit_iter4: "FAIL 0.87 (binding, lead-side Opus claude-opus-5-5[1m]) — blocking D31 repaired via disposition (b) in v0.3.2 (+D29, D33); delta audit pending"
plan_audit_delta: "PASS 0.90 (binding, lead-side Opus claude-opus-5-5[1m]) at d76d460e0 — D31/D29/D33 resolved, N1/N2 folded; plan closed"
plan_complete_at: "2026-09-26"
plan_status: audit-ready
tier: L
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md, progress.md]
requirement_count: 25
ac_count: 25
open_questions: []   # OQ-1, OQ-2 resolved by lead decision 2026-09-26 (plan.md §H)
a1_baseline: "WT-contract-schema, v0.5.1 at 65e0a9167"
a2_baseline: "8c9ee29b7"
```

## §E.2 Run-phase Evidence

### Pre-flight (plan.md §C — run 2026-09-27, worktree `.claude/worktrees/t1237`, HEAD `7d5159a54`)

| # | Check | Command | Observed result |
|---|---|---|---|
| C1 | A1 landed + absorbed | `git merge-base --is-ancestor b1a62fb2b HEAD` | exit 0 |
| C1 | A2 landed + absorbed | `git merge-base --is-ancestor 19daa9dcf HEAD` | exit 0 |
| C2 | A1 verify core present | `go doc ./internal/contract Verify` | `func Verify(in Inputs) Report` — pure, never errors |
| C2 | A1 loader present | `go doc ./internal/contract LoadDir` | `func LoadDir(specDir string) (Inputs, error)` |
| C3 | A2 record reader (R2) | `go doc ./internal/escalation` | `ParseRecord(data []byte) (Record, error)` (`internal/escalation/record.go:118`); `ReadCardState(path)`; `NeedsDecision(worktreeRoot, card)`; `RecordDir(worktreeRoot, card)` = `.moai/reports/<card>/escalation` |
| C3 | A2 class-4 comparison (R2) | `go doc ./internal/escalation NewAPIAdditions` | `doc: no symbol NewAPIAdditions` — `newAPIAdditions` (`newapi.go:150`) is UNEXPORTED → per R2: the production New APIs live comparison renders `not observed` + `not_performed: new-api-comparison-unavailable`; AC-CLOSURE-007 is covered through the injected comparison seam. Follow-up card owed: export a read-only comparison from A2. A2 logic is NOT copied. |
| C4 | A1 design delta (R1) | `diff` of `.moai/specs/SPEC-AUTONOMY-CONTRACT-001/design.md` at HEAD tree (v0.5.2) vs pin `65e0a9167` (v0.5.1) | wording only (provenance comment, A2→A2b attributions, interim-rule phrasing) — no field, code, or rule A4 consumes changed; receipt field set (v0.5.1) unchanged |
| C5 | Hook anchor re-measured | `grep -n 'checkBashCommand(input.ToolInput)' internal/hook/pre_tool.go` | `:426` (escalationOptions.denylisted closure) and `:525` (main path; `@MX:ANCHOR` at `:523-524` forbids a conditional return above it). design.md §F cited `:507` — the line drifted with the develop absorption; anchor semantics unchanged. Branch guard `:552`, integration lock `:570`, slot lease `:587`, push serializer `:611`, contract-sign guard `:634`; A4's two checks go after the contract-sign guard, before the Write/Edit block (`:645`). |

Baseline-attribution: all five rows measured in this run, in this tree, at HEAD `7d5159a54`.

### §31 autonomous decisions (card t1237 run phase)

| Decision | Choice | Rationale |
|---|---|---|
| R2 disposition | New APIs live comparison renders `not observed` + `new-api-comparison-unavailable` in production; the AC seam carries the test path | A2's `newAPIAdditions` is unexported; copying A2 logic is the plan.md §G anti-pattern. Follow-up card: export the read-only comparison from `internal/escalation`. |
| AC-CLOSURE-007 seam shape | section builder takes an injected `NewAPICompareFunc` returning `([]escalation.Addition, error)`; nil/unavailable → `not observed` | keeps AC-CLOSURE-007 testable both ways (two additions / error) without re-implementing A2 |
| Verdict TTY seam | reuse A1's `contractStdinIsTerminalFn` / `newContractLineReader` / `contractAgentMarkers()` seams | same refusal semantics as `sign` (design.md §E); no new marker set |

Milestone evidence follows as each milestone lands.

### M1 — Record shapes and the report model (commit `158c3be69`)

RED (`go test -count=1 ./internal/closure/`, before implementation, HEAD `5bc8a0296`, this run):

```
internal/closure/model_test.go:12:7: undefined: NewReport
internal/closure/model_test.go:81:12: undefined: NotPerformedTokens
internal/closure/model_test.go:99:6: undefined: IsNotPerformedToken
internal/closure/model_test.go:112:20: undefined: NotPerformedProgressMissing
internal/closure/receipt_view_test.go:21:21: undefined: DecodeKickoffReceipt
FAIL	github.com/modu-ai/moai-adk/internal/closure [build failed]
```

GREEN (`go test -count=1 -v ./internal/closure/`): 14 tests PASS — `ok github.com/modu-ai/moai-adk/internal/closure 0.251s`. Files: `model.go` (fixed key order per REQ-CLOSURE-002, closed 15-token not-performed catalogue, CountValue), `records.go` (strict schema_version decoders; unknown-schema/malformed lines skipped AND listed), `receipt_view.go` (v0.5.1 display decoder over A1 `contract.KickoffReceipt`; unknown fields listed, never dropped).

### M2 — Readiness rule and reason codes (commit pending)

RED (`go test -count=1 ./internal/closure/ -run 'TestAC_CLOSURE_013|TestAC_CLOSURE_016|TestAC_CLOSURE_018|…'`, before implementation, this run):

```
internal/closure/readiness_test.go:13:68: undefined: CommitPaths
internal/closure/readiness_test.go:92:66: undefined: SecondReviewState
internal/closure/readiness_test.go:93:10: undefined: SelectSecondReview
FAIL	github.com/modu-ai/moai-adk/internal/closure [build failed]
```

GREEN (`go test -count=1 ./internal/closure/`): `ok github.com/modu-ai/moai-adk/internal/closure 0.086s` — 29 tests incl. `TestAC_CLOSURE_013` (11 subtests: all six not-performed causes, performed pass/fail, stale with superseding commit, sync-commit non-staling, filter-order survivor), `TestAC_CLOSURE_016` (nine-code closed set + per-code fixtures + failed≠not-performed + accept/none ready + sorted dedup), `TestAC_CLOSURE_018` (advisory/off → no code), currency-undetermined honesty (`CurrencyUndetermined` flag → `push_check_undetermined`), substitute-backend flag. Files: `readiness.go` (`GitFacts` injection, `SelectSecondReview` §D filters, `EvaluateReadiness` closed nine-code set, currency rule over `contract.MatchGlob`).

### M3 — Report builder and renderer (commit pending)

RED (`go test -count=1 ./internal/closure/ -run 'TestAC_CLOSURE_00[2-9]'`, before implementation): `sign refused: verify_failed` on an intentionally incomplete draft contract — the fixture signing seam exercised before any builder existed; builders did not compile.

GREEN (`go test -count=1 ./internal/closure/...`): `ok ... internal/closure 19.316s` + `ok ... internal/closure/gitio 2.650s`. `go test -cover`: closure **88.9%** (≥85 target), gitio 68.6% (thin subprocess wrapper; M6 raises it). Files: `gitio/gitio.go` (bounded-timeout git subprocess: Head/IsAncestor with errors.As exit-1 discrimination/NonMergeCommits/WorktreeList/CommonDir/CurrentBranch/UpstreamBranch), `evidence.go` (§C.7 card evidence home with symlink canonicalization, evidence paths, canonical report hash, §E.2 row + §E.3 YAML parsers, REQ-CLOSURE-011 plan-audit search, `LoadA2Records` via A2's `ParseRecord`), `build.go` (BuildInput, Build assembly, atomic pair write via `atomicfile`), `sections_{core,escalation,review}.go` (12 section builders per design.md §G), `render_md.go` (Markdown from the JSON model only). AC tests landed: 002 (order+parity+determinism), 003 (missing→listed), 004 (reconciliation), 005 (invariants, armed/disarmed), 006 (ownership counts), 007 (New APIs seam + escalation-dir byte-identity), 008 (escalations + needs-decision), 009 (first verdict + mismatch), 010 (five receipt fixtures via `signtest`), 011 (both plan-audit streams), 014 (NOT PERFORMED/STALE/FAILED/substitute renders), 022 (verdict current/stale/latest-line).

§31 design decision — **canonical report hash**: `CanonicalReportHash` = SHA-256 of the report JSON with `generated_at` masked to "". Both the verdict recorder (M4) and the Human Verdict currency check use it; hashing raw file bytes instead would stale every verdict on the next rebuild, violating REQ-CLOSURE-002 determinism. `CountValue.UnmarshalJSON` added so the file round-trips through the struct (prerequisite of the same property).

### M4 — CLI surfaces (commit pending)

Files: `internal/cli/contract_report.go` (`moai contract report <card-id>`: queue lookup → §C.7 evidence home → A1 verify over the home's SPEC → Build → atomic pair write → prints the md path; refusals exit 2 naming the cause; the three A4 subcommands attach to A1's `contract` tree via file-name-ordered init — A1's files untouched), `contract_verdict.go` (`verdict <card> <verdict>`: closed refusal set agent_marker/not_tty/confirmation_mismatch/report_missing/git_identity_missing → exit 1 without writing; appends one `interactive-tty` record), `contract_pushcheck.go` (`push-check`: mode gate → inactive exit 0 under guided; `--all`/`--mirror`/unresolvable → `push_check_undetermined` exit 1; unknown flag exit 2; per-SPEC codes line). Shared core: `internal/closure/pushcheck.go` (`EvaluatePush` — REQ-CLOSURE-015 own-card candidacy over the source commit's tree, evidence from the live worktree, `EvaluateReadiness` per candidate) + `closuretest` package (the acceptance.md §A fixture, shared by closure/cli/hook tests).

RED (`go test -count=1 ./internal/cli/ -run 'TestAC_CLOSURE_001|…'`, before implementation): build failure — `runContractReport` etc. undefined.

GREEN (`go test -count=1 ./internal/cli/ -run 'TestAC_CLOSURE_001|TestAC_CLOSURE_019|TestAC_CLOSURE_020|TestAC_CLOSURE_024'`): `ok github.com/modu-ai/moai-adk/internal/cli 13.774s` — AC-001 (pair written at the c1 evidence dir; c2/c3/c4/c5 each exit 2 naming the cause; nothing created under any `.moai/reports`), AC-019 (ready/not-ready naming SPEC+codes/refspec/--all undetermined/missing remote ref undetermined/unknown flag 2/guided inactive), AC-020 (four refusals exit 1 with no file + success line with operator/method/report_sha256 = canonical hash), AC-024 (evidence lands only in the c1 worktree from both trees; sources name the home; removed worktree → primary). `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Logged by the lane orchestrator before the first run-phase `Agent()` spawn (2026-09-27).

Input parameters:

- tier: L (25 REQ / 25 AC, > 15 files)
- scope (estimated files): ~20 — new `internal/closure` package (+ `internal/closure/gitio`), 3 new CLI files, 2 MCP server files, 2 hook files, 4 Markdown mirrors + regenerated Codex TOML, plus tests
- domain count: 4 (Go source, MCP server surface, PreToolUse hook, template mirrors)
- file language mix: Go + Markdown + generated TOML
- concurrency benefit: LOW — coding-heavy; milestones are strictly ordered by decision reversibility (M1 record shapes → M2 readiness rule → M3 report builder → M4 CLI → M5 MCP record → M6 hook wiring → M7 mirrors)
- Agent Teams prereqs: not requested (no `--team`; no operator request)

Mode evaluation:

| Mode | Selected | Rationale (one line) |
|---|---|---|
| direct | no | multi-file, multi-milestone implementation |
| serial | **yes** | coding-heavy Tier L; per Anthropic's coding-task parallelism caveat the sequential delegation is the safe default; M1→M7 is a strict dependency chain |
| fanout | no | not research-heavy; write-capable parallelism would need isolated worktrees and the milestone chain is serial |
| sweep | no | new-code multi-rule work, not a mechanical uniform transform |

Decision: serial

Justification: plan.md §F orders the milestones by decision reversibility and each builds on the previous (record shapes → evaluator → builder → CLI → MCP append → hook wiring → mirrors); nothing is genuinely parallel. One manager-develop delegation (cycle_type=tdd) with per-milestone commits.

Kickoff: applied autonomously per CLAUDE.local.md §31 (operator policy, card t1266); progression mode autonomous. Kickoff gate basis: binding delta-audit PASS 0.90 (lead-side Opus `claude-opus-5-5[1m]`) at `d76d460e0`; the post-PASS commit `1aabfa37c` carries only the auditor-prescribed N1/N2 wording (lead-sanctioned).
