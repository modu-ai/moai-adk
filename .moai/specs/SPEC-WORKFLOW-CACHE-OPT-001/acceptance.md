# acceptance.md — SPEC-WORKFLOW-CACHE-OPT-001

> Reachability over token presence: every doctrine AC pairs the text change with the mechanical surface it invokes (per-surface discriminating checks, baseline → post). Go ACs require actual test-execution evidence (named tests, verbatim exit/output), never grep-only. No compound vacuous greps.

## §D — Acceptance Criteria Matrix

| AC | REQ | Verifies | Verification (single discriminating check) | Baseline → Post |
|----|-----|----------|--------------------------------------------|-----------------|
| AC-WCO-001 | REQ-SNAP-001 | Snapshot schema fields + loop-verdict-compatible conditions block | `go test -run 'TestSnapshotSchema' ./internal/verify/...` exit 0; test asserts presence of check id / command / exit / counts / timestamp / duration / key AND decodes a loop-verdict-shaped `conditions` block | no pkg → PASS |
| AC-WCO-002 | REQ-SNAP-002 | Key = HEAD SHA + porcelain-v2 digest + diff-HEAD content hash (D13); any tree change invalidates | `go test -run 'TestSnapshotKey' ./internal/verify/...` exit 0; table cases: clean tree, dirty tracked edit, RE-EDIT of an already-dirty tracked file (D13 boundary: porcelain-v2 byte-identical, diff-hash leg must change the key → no reuse), staged edit, ADD-untracked-file, HEAD advance — each yields a distinct key | no pkg → PASS |
| AC-WCO-003 | REQ-SNAP-003 | Freshness = key-equality AND 10-min TTL, E2E | `go test -run 'TestFreshness' ./internal/verify/...` exit 0; E2E: record → same-tree in-TTL check accepts → (a) mutate tracked file → stale, (b) add untracked file → stale, (c) injected clock past TTL (default 10 min; configurable value honored) → stale | no pkg → PASS |
| AC-WCO-004 | REQ-SNAP-004 | CLI verb exists AND is registered (cross-file reachability) | `go run ./cmd/moai verify --help` exit 0 and usage lists `record` + `check` | verb absent → exit 0 |
| AC-WCO-005 | REQ-SNAP-005 | gate.md consumes+produces via the real verb + defines force-fresh | gate.md names `moai verify` and `.moai/state/verify/` in a consumption step (each `grep -c` 0 → ≥1) AND gate.md documents the `--fresh` no-reuse mode (`grep -c '\-\-fresh' gate.md` 0 → ≥1; measured baseline: `grep -c fresh` = 0) AND AC-WCO-004 PASS (the named verb is invocable) | 0 → ≥1 (all legs) |
| AC-WCO-006 | REQ-SNAP-006 | run Phase 2.75 snapshot wiring | `grep -c "moai verify" .claude/skills/moai/workflows/run/task-decomposition.md` within the Phase 2.75 section (windowed read confirms placement) | 0 → ≥1 |
| AC-WCO-007 | REQ-SNAP-007 | sync Phase 0 snapshot consumption | quality-gates-context.md Phase 0 section names snapshot consumption + citation-as-evidence wording | 0 → ≥1 |
| AC-WCO-008 | REQ-SNAP-008 | loop Step 3 writes / Step 1 reads shared schema (discriminating tokens — NOT the pre-existing "snapshot" prose, 11 baseline matches) | Step-3 leg: `grep -c 'moai verify record' loop.md` (measured baseline 0) → ≥1 within the Step 3 window; Step-1 leg: `grep -c '\.moai/state/verify/' loop.md` (measured baseline 0) → ≥1 within the Step 1 window (windowed reads confirm placement) | 0 → ≥1 (both legs, verified-0 baselines) |
| AC-WCO-009 | REQ-SNAP-009 | Step 1.5 independence incl. force-fresh gate invocation (D7 transitive path) | windowed read of loop.md Step 1.5: (a) shall-NOT-consume-same-run-snapshot sentence present (0 → ≥1), (b) the gate invocation names `--fresh` (`grep -c '\-\-fresh'` in the Step 1.5 window, baseline 0 → ≥1); gate-side `--fresh` definition pinned by AC-WCO-005 | 0 → ≥1 (both legs) |
| AC-WCO-010 | REQ-SNAP-010 | stop-goal evaluator exact-match reuse path | `go test -run 'TestEvaluateSnapshot' ./internal/goal/...` exit 0; cases: (a) exact byte-string command match + fresh → reuses exit w/o CmdRunner call (fake runner call-count 0), (b) stale key or TTL-expired → CmdRunner executes, (c) near-miss command variant (e.g. added flag) → NO reuse, CmdRunner executes, (d) verdict payload carries snapshot attribution | no path → PASS |
| AC-WCO-011 | REQ-SNAP-011 | Stale-never-cited + attribution | `go test -run 'TestFreshness.*Stale' ./internal/verify/...` proves stale check returns non-reusable; doctrine: loop.md/gate.md snapshot sections carry path+key+command citation wording (grep ≥1 per file) | 0 → ≥1 |
| AC-WCO-012 | REQ-GATE-001 | Kickoff + 11.5 one call; gate preserved | moai.md: Step 11.3/11.5 region states single AskUserQuestion call carrying both questions (0 → ≥1) AND "score-independent" + "exactly once per pipeline entry" text still present (present → present) | see check |
| AC-WCO-013 | REQ-GATE-002 | Stage B single multi-question call (windowed — L152 PRESERVE collision avoided) | Removal leg WINDOWED to the Stage B axes section (`sed -n '/### Stage B/,/## Stage B Round 4/p'` or equivalent): the exact per-axis phrase "as a separate AskUserQuestion with up to 4 options" ≥1 → 0 AND single-call multi-question wording 0 → ≥1; PRESERVE leg: the L152 scope-boundary phrase "as a SEPARATE AskUserQuestion" (documentation-priority second question, [HARD]) present → present | paired + preserve |
| AC-WCO-014 | REQ-GATE-003 | harness proposal+approval merged | harness-build-entry.md: one-round proposal+approval wording present; two-sequential-rounds flow absent | 0 → ≥1 |
| AC-WCO-015 | REQ-GATE-004 | Full-pipeline success closes w/o manufactured question — BOTH surfaces | Leg 1 (moai.md): completion step carries the full-pipeline no-next-step-question clause (0 → ≥1); Leg 2 (sync/delivery.md): windowed to "### Phase 4: Completion and Next Steps" (L357 anchor) — the same suppression clause present (0 → ≥1); both surfaces: single-phase "(Recommended)" chain text preserved (present → present) | 0 → ≥1 (per surface) |
| AC-WCO-016 | REQ-GATE-005 | feedback single round | feedback.md: the 3 collection items ride one AskUserQuestion round (windowed read); per-item sequential-round instructions removed | ≥3 rounds → 1 |
| AC-WCO-017 | REQ-GATE-006 | No information reduction — per-surface expected inventories | Per-surface post-edit inventory checks: (a) moai.md merged round names BOTH questions (kickoff run-entry/review/abort decision + execution-shape worktree/sub-agent choice); (b) Stage B single call names all 4 axis fields (`verification`, `external_systems`, `ui_surface`, `team_sharing` — grep each token in the merged-round window, 4/4 present); (c) harness merged round names proposal AND approval; (d) feedback single round names all 3 fields (type, title, description — 3/3 tokens in the round window). Each inventory item is a grep-checkable token; verbatim outputs cited in §E.2 | inventory 4 surfaces, all tokens present |
| AC-WCO-018 | REQ-DELEG-001 | fix Level-1 orchestrator-direct | fix.md Phase 3: Level-1 row maps to orchestrator-direct formatter (0 → ≥1); mandate re-scoped "Level 2+" (0 → ≥1); `go test -run 'TestAgentlessUtilityNoLLMControlFlow' ./...` exit 0 | paired + test |
| AC-WCO-019 | REQ-DELEG-002 | loop Step 6 Level-1 direct | loop.md Step 6: Level-1 orchestrator-direct exception present | 0 → ≥1 |
| AC-WCO-020 | REQ-DELEG-003 | codemaps ≤1 spawn | codemaps.md Agent Chain Summary lists exactly 1 Agent() spawn phase; Phase 2/3 delegation-to-manager-docs [HARD] lines removed (≥2 → 0) | 3 → ≤1 |
| AC-WCO-021 | REQ-DELEG-004 | clean ≤2 spawns worst-case incl. Phase 5.5 | clean.md Agent Chain Summary lists exactly 2 spawn phases (combined 1+2, combined 4+5); per-phase spawn count in Execution Summary consistent; Phase 5.5 leg: the Phase 5.5 / Agent Chain Summary line reads orchestrator-direct ONLY — its "or a per-spawn `Agent(general-purpose)` refactoring specialist" alternative removed (≥1 → 0 in the Phase 5.5 window) | 4 → 2 (worst case incl. 5.5) |
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
| AC-WCO-032 | REQ-BOOK-005 | No record loss — per-surface expected inventories | Per-surface post-edit inventory checks: (a) loop.md batched-bookkeeping clause still names TaskCreate AND both TaskUpdate transitions (`in_progress`, `completed` — 3/3 tokens present); (b) loop.md aggregate MX_TAG_REPORT still names Tags Added / Removed / Updated classes (3/3 tokens); (c) review.md Phase 2 still names all 4 perspective headings (Security / Performance / Quality / UX — 4/4); (d) review.md secrets scan retains the 3 credential pattern classes (PRIVATE KEY / AKIA / ghp_ — 3/3). Each item grep-checkable; verbatim outputs cited in §E.2 | inventory 4 surfaces, all tokens present |
| AC-WCO-033 | REQ-GUARD-001 | Kickoff invariant repo-wide | `grep -rn "score-independent" .claude/skills/moai/workflows/moai.md .claude/skills/moai/workflows/run.md` — count non-decreasing vs baseline; no edited file conditions Kickoff on audit score | present → present |
| AC-WCO-034 | REQ-GUARD-002 | Channel monopoly intact | All merged flows still name AskUserQuestion as the channel (per-surface grep ≥1); no free-form prose-question instruction introduced in any edited file | present → present |
| AC-WCO-035 | REQ-GUARD-003 | VCI intact (delta-framed per measured baselines) | Snapshot doctrine sections cite `verification-claim-integrity.md` — per-file delta framing: gate.md 0 → ≥1, run/task-decomposition.md 0 → ≥1, sync/quality-gates-context.md 0 → ≥1, loop.md 2 → ≥3 (measured baseline 2: Step 9 + Ceiling-Exit contract already cite it); no edited surface permits evidence-free claims (M6 review) | per-file delta (see check) |
| AC-WCO-036 | REQ-GUARD-004 | Template split parity + build | Post-edit sweep, two legs: (a) workflow files — `cmp -s` local vs template for EVERY touched `workflows/**` file → all EQ; (b) agent files — `diff <(sed 's/(SPEC-V3R2-HRN-003)/(HRN-003)/' .claude/agents/moai/sync-auditor.md) internal/template/templates/.claude/agents/moai/sync-auditor.md` → empty (sanitized-parity; same transform pattern for any new internal-SPEC-ID line, each documented), plan-auditor.md `cmp -s` EQ unless a sanitization line is introduced (then its transform documented + diff-empty after transform); `grep -rn 'SPEC-V3R2-HRN-003' internal/template/templates/` → 0 (neutrality guard); `make build` exit 0; `go test ./internal/template/...` exit 0 | split parity (see check) + exit 0 |

## §D.1 — Given-When-Then Scenarios

### Scenario 1 — Snapshot reuse across layers (happy path)
- **Given** run Phase 2.75 executed `go test ./...` (exit 0) and recorded it via `moai verify record` under key K for the current tree
- **When** sync Phase 0 (`gate-sync-1`) starts on the unchanged tree, within the 10-minute TTL, and runs `moai verify check`
- **Then** the check reports fresh, sync Phase 0 cites snapshot path + key K + original command + exit 0 as its full-suite evidence, and the test suite is NOT re-executed.

### Scenario 2 — Stale detection (safety path)
- **Given** a snapshot recorded under key K
- **When** any tracked-content change (including a re-edit of an already-dirty file — caught by the `git diff HEAD` content-hash leg), staged or untracked-file change (caught by the porcelain-v2 leg) lands OR the 10-minute TTL elapses, and a consumer runs `moai verify check`
- **Then** the check exits stale, and the consumer re-executes the check instead of reusing — a stale snapshot is never cited as evidence.

### Scenario 3 — Loop success-exit independence (incl. gate-mediated path)
- **Given** a `/moai loop` iteration whose Step 3 snapshot satisfies the completion predicate at Step 1 (tree unchanged ⇒ the Step-3 snapshot key is still fresh)
- **When** Step 1.5 (Independent Final Pass) runs
- **Then** Step 1.5 invokes `/moai gate --fresh`, so the gate consumes NO snapshot — the same-run Step-3 snapshot cannot flow back through the gate layer; only on independent confirmation does the loop declare success-exit — and Step 1.5's force-fresh results MAY be recorded for downstream (sync) consumption.

### Scenario 4 — Merged kickoff round (gate preserved)
- **Given** a default pipeline whose plan-audit gate returned PASS
- **When** the plan→run boundary is reached
- **Then** ONE AskUserQuestion call presents both the Kickoff approval question and the execution-shape question; declining the Kickoff option halts run-phase entry exactly as before — no score, snapshot, or merged layout bypasses the human gate.

### Scenario 5 — stop-goal turn-end reuse
- **Given** an armed goal with Tier-1 condition `go test ./... (expect 0)` and a fresh snapshot entry whose recorded command is the exact byte-string `go test ./...`
- **When** the `stop-goal` Stop hook evaluates at turn-end
- **Then** the evaluator reuses the recorded exit 0 without spawning the command, and the verdict payload records the snapshot attribution; if the tree changed, the TTL elapsed, or the condition command differs by even one byte (no normalization in M1), the evaluator executes the command exactly as today.

## §D.2 — Edge Cases

1. **TOCTOU window**: tree mutates between `verify check` and the consumer's citation — freshness check runs at consumption time; doctrine instructs re-check on any intervening write step.
2. **Flaky test recorded PASS**: bounded by the 10-minute TTL (Settled Decisions — key-equality alone is insufficient by design); Residual-risk section of consumer reports names flake risk when reusing.
2b. **In-place edit of an already-listed untracked file**: outside all three key inputs — HEAD, porcelain-v2 digest, AND `git diff HEAD` content hash (untracked content appears in none of them; accepted limitation, Settled Decisions) — mitigated by the TTL bound; named as Residual-risk in consumer reports. (Re-edits of already-dirty TRACKED files are NOT a limitation — the diff-hash leg catches them per D13.)
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
- [ ] Clarification gate clean: zero `NEEDS CLARIFICATION` markers in plan.md/research.md (all 3 resolved by user decision, plan-audit iter-1 D1 — recorded in plan.md Settled Decisions)
- [ ] All touched `.claude/**` files satisfy the split parity rule — workflow files byte-equal, agent files sanitized-parity (AC-WCO-036)
- [ ] `make build` + `go test ./...` + lint green at M6
- [ ] Guard-rail sweep: Kickoff mandatory, channel monopoly, VCI — all confirmed intact (AC-WCO-033/034/035)
- [ ] progress.md §E.2/§E.3 populated by manager-develop with the 5-section evidence format
