# acceptance.md — SPEC-WORKFLOW-CACHE-OPT-001

> Reachability over token presence: every doctrine AC pairs the text change with the mechanical surface it invokes (per-surface discriminating checks, baseline → post). Go ACs require actual test-execution evidence (named tests, verbatim exit/output), never grep-only. No compound vacuous greps.

## §D — Acceptance Criteria Matrix

| AC | REQ | Verifies | Verification (single discriminating check) | Baseline → Post |
|----|-----|----------|--------------------------------------------|-----------------|
| AC-WCO-001 | REQ-SNAP-001 | Snapshot schema fields + loop-verdict-compatible conditions block | `go test -run 'TestSnapshotSchema' ./internal/verify/...` exit 0; test asserts presence of check id / command / exit / counts / timestamp / duration / key AND decodes a loop-verdict-shaped `conditions` block | no pkg → PASS |
| AC-WCO-002 | REQ-SNAP-002 | Key changes on any tracked-content change | `go test -run 'TestSnapshotKey' ./internal/verify/...` exit 0; table cases: clean tree, dirty tracked edit, HEAD advance — all yield distinct keys | no pkg → PASS |
| AC-WCO-003 | REQ-SNAP-003 | Freshness accept/stale E2E | `go test -run 'TestFreshness' ./internal/verify/...` exit 0; E2E: record → same-tree check accepts → mutate tracked file → check reports stale | no pkg → PASS |
| AC-WCO-004 | REQ-SNAP-004 | CLI verb exists AND is registered (cross-file reachability) | `go run ./cmd/moai verify --help` exit 0 and usage lists `record` + `check` | verb absent → exit 0 |
| AC-WCO-005 | REQ-SNAP-005 | gate.md consumes+produces via the real verb | gate.md names `moai verify` and `.moai/state/verify/` in a consumption step (each `grep -c` 0 → ≥1) AND AC-WCO-004 PASS (the named verb is invocable) | 0 → ≥1 (both) |
| AC-WCO-006 | REQ-SNAP-006 | run Phase 2.75 snapshot wiring | `grep -c "moai verify" .claude/skills/moai/workflows/run/task-decomposition.md` within the Phase 2.75 section (windowed read confirms placement) | 0 → ≥1 |
| AC-WCO-007 | REQ-SNAP-007 | sync Phase 0 snapshot consumption | quality-gates-context.md Phase 0 section names snapshot consumption + citation-as-evidence wording | 0 → ≥1 |
| AC-WCO-008 | REQ-SNAP-008 | loop Step 3 writes / Step 1 reads shared schema | loop.md Step 3 names the snapshot write via `moai verify record`; Step 1 references the persisted snapshot (two separate greps, each 0 → ≥1) | 0 → ≥1 (both) |
| AC-WCO-009 | REQ-SNAP-009 | Step 1.5 independence carve-out lands IN Step 1.5 | windowed read of loop.md Step 1.5 contains the shall-NOT-consume-same-run-snapshot sentence | 0 → ≥1 |
| AC-WCO-010 | REQ-SNAP-010 | stop-goal evaluator reuse path | `go test -run 'TestEvaluateSnapshot' ./internal/goal/...` exit 0; cases: (a) fresh-hit reuses exit w/o CmdRunner call (fake runner call-count 0), (b) stale → CmdRunner executes, (c) verdict payload carries snapshot attribution | no path → PASS |
| AC-WCO-011 | REQ-SNAP-011 | Stale-never-cited + attribution | `go test -run 'TestFreshness.*Stale' ./internal/verify/...` proves stale check returns non-reusable; doctrine: loop.md/gate.md snapshot sections carry path+key+command citation wording (grep ≥1 per file) | 0 → ≥1 |
| AC-WCO-012 | REQ-GATE-001 | Kickoff + 11.5 one call; gate preserved | moai.md: Step 11.3/11.5 region states single AskUserQuestion call carrying both questions (0 → ≥1) AND "score-independent" + "exactly once per pipeline entry" text still present (present → present) | see check |
| AC-WCO-013 | REQ-GATE-002 | Stage B single multi-question call | codebase-analysis.md Stage B: "separate AskUserQuestion" per-axis wording removed (≥1 → 0) AND single-call multi-question wording present (0 → ≥1) | paired |
| AC-WCO-014 | REQ-GATE-003 | harness proposal+approval merged | harness-build-entry.md: one-round proposal+approval wording present; two-sequential-rounds flow absent | 0 → ≥1 |
| AC-WCO-015 | REQ-GATE-004 | Full-pipeline success closes w/o manufactured question | moai.md completion step: full-pipeline no-next-step-question clause present; single-phase "(Recommended)" chain text preserved (present → present) | 0 → ≥1 |
| AC-WCO-016 | REQ-GATE-005 | feedback single round | feedback.md: the 3 collection items ride one AskUserQuestion round (windowed read); per-item sequential-round instructions removed | ≥3 rounds → 1 |
| AC-WCO-017 | REQ-GATE-006 | No information reduction | Diff review of M2 surfaces: every question/option present pre-merge appears post-merge (manual diff evidence in §E.2, per-surface) | diff evidence |
| AC-WCO-018 | REQ-DELEG-001 | fix Level-1 orchestrator-direct | fix.md Phase 3: Level-1 row maps to orchestrator-direct formatter (0 → ≥1); mandate re-scoped "Level 2+" (0 → ≥1); `go test -run 'TestAgentlessUtilityNoLLMControlFlow' ./...` exit 0 | paired + test |
| AC-WCO-019 | REQ-DELEG-002 | loop Step 6 Level-1 direct | loop.md Step 6: Level-1 orchestrator-direct exception present | 0 → ≥1 |
| AC-WCO-020 | REQ-DELEG-003 | codemaps ≤1 spawn | codemaps.md Agent Chain Summary lists exactly 1 Agent() spawn phase; Phase 2/3 delegation-to-manager-docs [HARD] lines removed (≥2 → 0) | 3 → ≤1 |
| AC-WCO-021 | REQ-DELEG-004 | clean ≤2 spawns | clean.md Agent Chain Summary lists exactly 2 spawn phases (combined 1+2, combined 4+5); per-phase spawn count in Execution Summary consistent | 4 → 2 |
| AC-WCO-022 | REQ-DELEG-005 | mx <5-item direct edit | mx.md Pass 3: <5-item orchestrator-direct clause present | 0 → ≥1 |
| AC-WCO-023 | REQ-DELEG-006 | Approval semantics untouched | Level-3 AskUserQuestion approval (fix.md), clean Phase 3 removal-plan AskUserQuestion, @MX:ANCHOR protection lines all still present post-edit (present → present, per-surface grep) | present → present |
| AC-WCO-024 | REQ-AUDIT-001 | Tier S single-pass default | run.md + run/phase-execution.md: Tier S single-audit-pass default clause present AND "audit always runs (once)" preserved for all tiers (no skip-audit wording introduced) | 0 → ≥1 + guard |
| AC-WCO-025 | REQ-AUDIT-002 | Defect-list + delta re-check | plan-auditor.md + sync-auditor.md: FAIL verdict defect-list format (finding id / location / severity / required fix) present in output contract (each 0 → ≥1); owning workflow body documents delta-scoped re-audit | 0 → ≥1 (per file) |
| AC-WCO-026 | REQ-AUDIT-003 | project retry cap 1 | doc-generation.md Phase 3.1: retry ceiling text reads 1 (iteration=3 escalation wording updated consistently) | 3 → 1 |
| AC-WCO-027 | REQ-AUDIT-004 | Verdict authority preserved | Auditor-verdict-authority sentence present in the defect-list protocol text; no orchestrator-self-verdict wording introduced | 0 → ≥1 |
| AC-WCO-028 | REQ-BOOK-001 | loop bookkeeping 1 batch/iteration | loop.md Steps 5-7: batched-single-turn bookkeeping wording replaces per-fix TaskUpdate mandates (per-fix [HARD] lines ≥2 → 0; batch clause 0 → ≥1) | paired |
| AC-WCO-029 | REQ-BOOK-002 | MX_TAG_REPORT once at exit | loop.md: per-iteration MX_TAG_REPORT emission removed from Step 7.5 (≥1 → 0); aggregate-at-exit clause present (0 → ≥1) | paired |
| AC-WCO-030 | REQ-BOOK-003 | review 4 lenses Mode-4 parallel | review.md Phase 2: parallel read-only fan-out wording present (0 → ≥1); "sequentially" single-pass instruction removed (≥1 → 0); sync-auditor synthesis/verdict ownership preserved (present → present) | paired |
| AC-WCO-031 | REQ-BOOK-004 | Incremental secrets scan | review.md: last-SHA checkpoint under `.moai/state/` + `<last-sha>..HEAD` incremental command present (0 → ≥1); full `--all` scan retained as first-run/flag fallback (present → present) | paired |
| AC-WCO-032 | REQ-BOOK-005 | No record loss | M5 diff review: task records, tag actions, finding classes all preserved (per-surface diff evidence in §E.2) | diff evidence |
| AC-WCO-033 | REQ-GUARD-001 | Kickoff invariant repo-wide | `grep -rn "score-independent" .claude/skills/moai/workflows/moai.md .claude/skills/moai/workflows/run.md` — count non-decreasing vs baseline; no edited file conditions Kickoff on audit score | present → present |
| AC-WCO-034 | REQ-GUARD-002 | Channel monopoly intact | All merged flows still name AskUserQuestion as the channel (per-surface grep ≥1); no free-form prose-question instruction introduced in any edited file | present → present |
| AC-WCO-035 | REQ-GUARD-003 | VCI intact | Snapshot doctrine sections cite `verification-claim-integrity.md` (0 → ≥1 per consumer file); no edited surface permits evidence-free claims (M6 review) | 0 → ≥1 |
| AC-WCO-036 | REQ-GUARD-004 | Template parity + build | Post-edit sweep: `cmp -s` local vs template for EVERY touched `.claude/**` file → all EQ; `make build` exit 0; `go test ./internal/template/...` exit 0 | all EQ + exit 0 |

## §D.1 — Given-When-Then Scenarios

### Scenario 1 — Snapshot reuse across layers (happy path)
- **Given** run Phase 2.75 executed `go test ./...` (exit 0) and recorded it via `moai verify record` under key K for the current tree
- **When** sync Phase 0 (`gate-sync-1`) starts on the unchanged tree and runs `moai verify check`
- **Then** the check reports fresh, sync Phase 0 cites snapshot path + key K + original command + exit 0 as its full-suite evidence, and the test suite is NOT re-executed.

### Scenario 2 — Stale detection (safety path)
- **Given** a snapshot recorded under key K
- **When** any tracked file's content changes and a consumer runs `moai verify check`
- **Then** the recomputed key ≠ K, the check exits stale, and the consumer re-executes the check instead of reusing — a stale snapshot is never cited as evidence.

### Scenario 3 — Loop success-exit independence
- **Given** a `/moai loop` iteration whose Step 3 snapshot satisfies the completion predicate at Step 1
- **When** Step 1.5 (Independent Final Pass) runs
- **Then** Step 1.5 re-executes a fresh gate (never consuming the same run's snapshot); only on independent confirmation does the loop declare success-exit — and Step 1.5's fresh results MAY be recorded for downstream (sync) consumption.

### Scenario 4 — Merged kickoff round (gate preserved)
- **Given** a default pipeline whose plan-audit gate returned PASS
- **When** the plan→run boundary is reached
- **Then** ONE AskUserQuestion call presents both the Kickoff approval question and the execution-shape question; declining the Kickoff option halts run-phase entry exactly as before — no score, snapshot, or merged layout bypasses the human gate.

### Scenario 5 — stop-goal turn-end reuse
- **Given** an armed goal with Tier-1 condition `go test ./... (expect 0)` and a fresh snapshot entry for that command on the current tree
- **When** the `stop-goal` Stop hook evaluates at turn-end
- **Then** the evaluator reuses the recorded exit 0 without spawning the command, and the verdict payload records the snapshot attribution; if the tree changed since recording, the evaluator executes the command exactly as today.

## §D.2 — Edge Cases

1. **TOCTOU window**: tree mutates between `verify check` and the consumer's citation — freshness check runs at consumption time; doctrine instructs re-check on any intervening write step.
2. **Flaky test recorded PASS**: bounded by the freshness-acceptance rule (clarification #2); Residual-risk section of consumer reports names flake risk when reusing.
3. **Concurrent writers (two sessions, one checkout)**: atomic temp+rename write (goal/state.go pattern); last-writer-wins on identical key is benign (same tree ⇒ equivalent results); pre-spawn sync check remains the session-level guard.
4. **`.moai/state/` unwritable**: fail-open — recording is skipped with an explicit note; consumers fall back to re-execution (never block, never fabricate).
5. **Stop-hook deadline exceeded during key computation**: time-boxed per Advisory-Check Discipline; evaluator falls back to command re-execution (correctness preserved, optimization skipped).
6. **First run / no snapshot**: every consumer's absent-snapshot path is plain re-execution (today's behavior) — the contract is strictly additive.
7. **Empty mx tag set / empty loop queue**: <5-item and batching rules degrade to no-op; empty-queue immediate exit unchanged.

## §D.3 — Quality Gate Criteria

- New `internal/verify` package coverage ≥ 85%; `internal/goal` coverage non-regressing.
- `golangci-lint run` — zero NEW findings vs pre-flight baseline.
- Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- Subagent-boundary grep (B3) → 0 matches in `internal/verify` + `internal/goal`.
- `moai spec lint` clean on this SPEC directory (strict).

## §D.4 — Definition of Done

- [ ] AC-WCO-001..036 all PASS with per-AC verbatim evidence in progress.md §E.2
- [ ] All 3 [NEEDS CLARIFICATION] markers resolved and recorded in plan.md Settled Decisions before run-phase entry
- [ ] All touched `.claude/**` files byte-equal in `internal/template/templates/` (AC-WCO-036)
- [ ] `make build` + `go test ./...` + lint green at M6
- [ ] Guard-rail sweep: Kickoff mandatory, channel monopoly, VCI — all confirmed intact (AC-WCO-033/034/035)
- [ ] progress.md §E.2/§E.3 populated by manager-develop with the 5-section evidence format
