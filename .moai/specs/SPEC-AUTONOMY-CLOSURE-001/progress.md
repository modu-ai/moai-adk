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

### M6 — Hook wiring (commit `9f38b7593`)

RED (`go test -count=1 ./internal/hook/ -run 'TestAC_CLOSURE'`, before implementation): build failure — `checkContractVerdict`/`checkClosurePush` undefined.

GREEN: `ok github.com/modu-ai/moai-adk/internal/hook` — AC-015 (five push forms each deny `CLOSURE_PUSH_STOP:` with `SPEC-FIXTURE-001=` + `second_review_not_performed`; performed pass clears; completed-in-pushed-commit still a candidate; `git push origin WT-feature` leaves the evaluation seam at 0; out-of-range SPEC-002 not evaluated while ready c1 is; another card's SPEC-directory edit re-admits SPEC-002 with `closure_report_missing` + `second_review_not_performed`), AC-017 (--all/--mirror/sh -c/`$BR`/`$( )`/bare push without upstream/missing remote ref → `push_check_undetermined`), AC-021 (verdict deny under guided AND contract; no closure-verdict writes outside the human path — grep guard over the cli package), AC-023 (guided/absent/bogus mode → no A4 decision, zero evaluations, Handle output byte-identical to `NewSafeDefaultOutput`).

### M7 — Auditor instructions and mirrors (commit `f2a8783cd`)

Marker blocks between `<!-- moai:closure-second-review:start/end -->` in the template and local copies of the skill and the sync-auditor agent (card argument + target baseBranch + after the last governed-path commit); `make agents-emit` regenerated `templates/.codex/agents/moai/sync-auditor.toml`; `make build` re-embedded. Neutrality regexes (`SPEC-[A-Z]`, `\bt[0-9]{3,5}\b`, dates, `A-Q[0-9]`, 9-40 hex): 0 matches over the block content.

### Run-phase commit list (branch `WT-completion-report`)

| Commit | Content |
|---|---|
| `5bc8a0296` | pre-flight evidence + draft→in-progress |
| `158c3be69` | M1 record shapes and report model |
| `5d079fb28` | M2 readiness rule and reason codes |
| `47bfdbeb8` | M3 report builder and renderer |
| `c55208625` | M4 CLI surfaces |
| `8bd821972` | M5 audit_multi second-review record |
| `9f38b7593` | M6 hook wiring |
| `f2a8783cd` | M7 auditor instructions and mirrors |
| `e51c87002` | AC-025 named test |
| `8c746ccc9` | in-package push evaluation coverage |
| `6611fcae5` | lint findings over the A4 packages |

### Self-verification (all commands run in this worktree at HEAD `6611fcae5`)

| Item | Command | Result |
|---|---|---|
| E1 | `go test -v -count=1 -run '^TestAC_CLOSURE_(…\|…)$'` per package | closure 15/15 PASS, cli 5/5 PASS, hook 4/4 PASS, template 1 subtest PASS (`TestAC_CLOSURE_025`) — **25/25**, zero `--- FAIL`, zero `[no tests to run]` |
| E2 | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 both (build=0, win=0) |
| E3 | `go test -cover ./internal/closure/...` | closure **88.6%**, gitio **88.1%** (≥85); `closuretest` 0.0% — test-support package by convention (the escalationtest precedent), exercised by the closure/cli/hook test suites |
| E4 | `golangci-lint run ./internal/closure/... ./internal/cli/... ./internal/hook/... ./internal/template/...` (golangci-lint **v2.1.6**, the CI judge) | **0 issues**. NEW findings during the run: 7 (2 errcheck `f.Close`, 2 unused test helpers, 1 ineffassign, 2 staticcheck QF) — all fixed in `6611fcae5`; baseline of the touched packages was clean pre-SPEC |
| E5 | `make agents-emit-check` | exit 0 |
| E6 | `moai spec lint SPEC-AUTONOMY-CLOSURE-001` | exit 0, "No findings — all SPEC documents are valid" |
| E7 | RED evidence | per-milestone verbatim captured above (M1 build-failure output, M2 undefined symbols, M4/M5 undefined symbols, M6 undefined checks) |
| E8 | blockers | none — no §31 decision required operator input beyond the choices recorded above |

Boundary greps: `grep -rn 'AskUserQuestion' internal/closure internal/hook internal/cli/contract_*` non-test → 1 hit, the pre-existing blocklist line in `pre_tool.go` (denies the tool; predates this SPEC). `grep -rn "Retired\|superseded" internal/closure …` → only this SPEC's own STALE rendering strings. `verdict.md` is written by no code path (the constant catalogue deliberately omits it).

### Sync-audit repair — F1-F5 + F10 (lead-side Opus sync-audit FAIL 63.3; report `.moai/reports/t1237/sync-audit-opus.md` §7)

Repairs landed one commit per defect; every fix was preceded by its failing regression test (auditor §8.1 reproduce-first). All measurements below: this run, this tree, HEAD `819847831`.

| Defect | Fix | Commit | Regression evidence |
|---|---|---|---|
| F1 [High] | `classifyPushCommand` gates on a push candidate (`\bgit\b[^;&\|\n]*\bpush\b`) before the unprovable judgment; pushless `$VAR`/`$(…)`/wrapper-shell commands get no A4 decision | `14a09d162` | RED (pre-fix): `command "echo $HOME": decision "deny" reason "CLOSURE_PUSH_STOP: push_check_undetermined (command substitution, eval, a wrapper shell, or a variable operand)", want no A4 decision`; GREEN post-fix. Test: `TestAC_CLOSURE_017/pushless_unprovable_commands_are_not_judged` |
| F2 [High] | design.md §C.1 rows 2-3: bare remote → upstream judgment; colon-less HEAD/@ → destination = current branch; matched operand quotes stripped; wrapped pushes (env assignment, subshell, `time`/`command`) → undetermined; `\|` and `&` join the segment splitters; the caller-resolved tree is threaded into the classifier | `1d8d9fad2` | RED (pre-fix): 11 new AC-015/017 forms all fail — 6 evaluate forms `decision = "", want deny` + 5 undetermined forms — matching probe E6; GREEN post-fix |
| F3 [High] | `appendSecondReviewRecord` receives the REQUESTED target from `runMultiAudit` and records it verbatim (was the literal `"baseBranch"`) | `cb64d1c6e` | RED (pre-fix): `recorded target = "baseBranch", want the requested "uncommittedChanges"` (and the `""` default case), matching probe E7; GREEN incl. the `SelectSecondReview` scope-not-covered assertion |
| F4 [Med] | AC-023 fixture gains `writeGitFlowConfig(t, f)` after `queueC1` so `integrationBranchFor` resolves and the mode gate is reachable | `de302c7c2` | Mutant check (`go test -overlay`, guided-gate deleted): `--- FAIL: TestAC_CLOSURE_023` + `mode "guided": A4 evaluations ran (0 -> 1)` — mutant dead; original code + strengthened fixture passes (E8 contrast) |
| F5 [Med] | card_id validated against `contract.CardPattern` before any path construction; failure rides `second_review_record_error`, audit result unaltered | `20527c978` | RED (pre-fix): `second_review_record_error empty, want the invalid card_id rejection` (traversal id `../../../../tmp/zzt1237`); GREEN post-fix: error set, nothing written outside the evidence home |
| F10 | `@MX:ANCHOR` on `ResolveEvidenceHome` and `gitio.Head` (4 non-test callers each, auditor E9) | `819847831` | n/a (annotation) |

Remeasurement batch (env-scrubbed single compound invocations, this run, this tree, HEAD `819847831`):

- `go test -timeout 30m -count=1 ./internal/closure/...` → `ok … internal/closure 14.054s` + `ok … internal/closure/gitio 2.566s` (`closuretest` `[no test files]` — test-support package).
- `go test -timeout 30m -count=1 -v -run '^TestAC_CLOSURE_' ./internal/hook/ ./internal/template/` → 6 `--- PASS` top-level, `ok` both packages.
- `go test -timeout 30m -count=1 -v -run '^TestAC_CLOSURE_\|AuditMulti\|Convergence\|Contract' ./internal/cli/` → 58 `--- PASS`, `ok … internal/cli 68.852s`.
- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `golangci-lint run ./internal/closure/... ./internal/cli/... ./internal/hook/... ./internal/template/...` (PATH **v2.1.6**, the CI judge) → `0 issues.`, exit 0.
- `moai spec lint SPEC-AUTONOMY-CLOSURE-001` with a TREE-BUILT binary (`go build -o bin/moai ./cmd/moai`; the installed `~/go/bin/moai` is stale at `a8a9b9376`, older than this tree's VacuousTestAssertion rule, and is FORBIDDEN as a verdict source — F6's cause): `0 error(s), 2 warning(s)` — `acceptance.md:13` (run-pattern anchoring) and `acceptance.md:14` (outcome-assertion delimiter). Both recorded below; acceptance.md not modified.

Debt list — recorded as debt per the lead's disposition (all Low/optional in audit §7; no repair this round):

- **F6** [Low] The §E.2 self-verification E6 row above cites the stale installed binary's "No findings"; the tree binary emits the 2 warnings above. `acceptance.md:13`'s unanchored run-pattern is real debt in SPEC body prose (amending it is manager-spec's domain). `acceptance.md:14` is a likely VacuousTestAssertion rule false-positive (the pattern already carries a space after the `<NNN>` placeholder) — `/moai:feedback` candidate.
- **F7** [Low] `contract_pushcheck.go` carries its own pre-F2 classifier (REQ-019 parity: bare-remote / colon-less-HEAD forms judge differently there than in the hook). Follow-up: share the hook classifier.
- **F8** [Low] `pushcheck.go` silently drops undecodable contracts and ignores closure-verdict.jsonl read errors — consider REQ-017 undetermined routing.
- **F9** [Low] `report_sha256` is the generated_at-masked canonical hash, not the file-byte SHA-256 (deliberate, coupled to REQ-002 determinism); the REQ-020/022 wording amendment is manager-spec's.
- **F11** [Low] New APIs stay `not observed` (plan.md R2 disposition: A2's `newAPIAdditions` is unexported) — follow-up card material.
- **F12** [Low] README 4-locale carries no `moai contract` verbs — manager-docs scope.
- **F13** [Low] AC-012's top-level test count grew by 2 (`TargetRecorded`, `PathTraversalCardID`; family now 6) against the acceptance.md one-top-level-test convention — the same accepted deviation class the audit recorded.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_commit_sha: "6611fcae5"
run_complete_at: "2026-09-27"
ac_pass_count: 25
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "n/a (worktree lane, no push — lead batch-pushes develop)"
l44_post_push_fetch: "n/a (same)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass
  windows: pass
total_run_phase_files: "internal/closure (16 files incl. gitio + closuretest), internal/cli (5), internal/hook (3), internal/template (4 + regenerated toml/catalog)"
m1_to_m7_commit_strategy: "per-milestone conventional commits with t1237 in every subject, no amend, no push"
milestones:
  - {id: M1, commit: "158c3be69", subject: "record shapes and report model"}
  - {id: M2, commit: "5d079fb28", subject: "readiness rule and reason codes"}
  - {id: M3, commit: "47bfdbeb8", subject: "report builder and renderer"}
  - {id: M4, commit: "c55208625", subject: "CLI surfaces"}
  - {id: M5, commit: "8bd821972", subject: "audit_multi second-review record"}
  - {id: M6, commit: "9f38b7593", subject: "hook wiring"}
  - {id: M7, commit: "f2a8783cd", subject: "auditor instructions and mirrors"}
follow_up_cards:
  - "export a read-only class-4 comparison from internal/escalation (newAPIAdditions) so the closure report's New APIs section can run the live comparison (R2 disposition)"
known_debt:
  - "canonical report hash (generated_at-masked SHA-256) is a §31 design decision recorded in §E.2 M3 — reconciled against REQ-CLOSURE-002 determinism; flagged for sync-auditor review"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-27
sync_commit_sha: "77508c33a"
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-AUTONOMY-CLOSURE-001' CHANGELOG.md = 0 (pre-emission) — no duplicate entry"
b12_self_test_b: "acceptance.md distinct AC = 25 (AC-CLOSURE-001..025, zero [RETIRED]/[REF] markers — all live); CHANGELOG entry references 25 AC-CLOSURE criteria — match"
b12_self_test_c: "all paths in CHANGELOG entry verified via ls: internal/closure/ (11 files), internal/closure/gitio/gitio.go, internal/closure/closuretest/closuretest.go, internal/cli/contract_{report,verdict,pushcheck}.go, internal/cli/mcp_audit_multi_record.go, internal/hook/closure_push.go, .claude/agents/moai/sync-auditor.md, .claude/skills/moai-ref-cross-model-audit/SKILL.md + template mirrors"
changelog_entry_position: "Added, first entry of [Unreleased]"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (merged 3-phase close, single sync commit)"
  plan_md: n/a (no frontmatter status field)
  acceptance_md: n/a (no frontmatter status field)
  progress_md: n/a (progress carries §E.4 signal, not frontmatter status)
  updated_field: "2026-09-27 (unchanged — sync same day)"
mx_tag_validation:
  existing: "build.go @MX:ANCHOR on Build (single assembly point, [AUTO] + @MX:REASON) — kept"
  gaps_reported_not_retaged: "ResolveEvidenceHome (4 non-test callers) and gitio.Head (3 non-test callers) lack @MX:ANCHOR per the fan_in>=3 gate — reported to the lead, not bulk-retagged in sync phase"
canary_compliance_check:
  spec_body_untouched: true
  forbidden_files_touched: false
  mx_delta: "0 tags added/removed in sync phase; run-phase [AUTO] tags left as-is"
```

Docs surfaces: README 4-locale set deliberately NOT extended — no README command table lists the `moai contract` family (it was absent when A1 landed `verify|show|sign` and remains absent), so adding only the three A4 verbs would be both inconsistent and beyond a one-line change; flagged to the lead as a follow-up instead. Codemaps scoped refresh: `.moai/project/codemaps/modules.md` + `entry-points.md` (`data-flow.md`/`dependencies.md` left for the next full regeneration).

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
