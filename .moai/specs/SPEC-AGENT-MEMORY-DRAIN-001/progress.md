# SPEC-AGENT-MEMORY-DRAIN-001 — Progress

> Tier M · card t223 · status: draft (plan-phase v0.2.0; plan-audit iter-1 PASS-WITH-DEBT 0.94, fixes applied — awaiting kickoff)

## §E.1 Plan-phase Audit-Ready Signal

- **Claim**: plan-phase artifact set complete for SPEC-AGENT-MEMORY-DRAIN-001 (Tier M:
  spec.md + plan.md + acceptance.md + progress.md) at v0.2.1, status `draft`, authored
  2026-09-02 on card tree `c0c36c421` (branch `WT-agent-memory-drain`). Plan-audit
  iteration 1 returned **PASS-WITH-DEBT 0.94** (Tier M threshold 0.80 met); audit
  defects D1-D3, D5, D7, D8 folded in at v0.2.0. Kickoff approved 2026-09-02 (design (c)
  adopted, AUTONOMOUS progression); v0.2.1 carries the non-transition `phase` correction
  to `"v3.1.4"` (card joins the v3.1.4 close, PR #1685).
- **Evidence**:
  - SPEC ID pre-write check: `SPEC-AGENT-MEMORY-DRAIN-001` → `PASS` (executed Bash).
  - Uniqueness: `ls .moai/specs | grep -c AGENT-MEMORY` → 0 pre-existing.
  - RED-now cells (single-invocation carrier form, verification-completeness §2.1):
    `moai memory drain` → `Unknown command "drain" for "moai memory".` exit 1;
    `comm -12 /tmp/t223-wt-topics.txt /tmp/t223-primary-topics.txt` → empty stdout,
    exit 0 (EV-002); `grep -n "mirror" …/internal/hook/post_tool.go` → empty stdout,
    exit 1 (EV-005); per-worktree native `memory/` dirs → 0 (baseline AC-AM-009).
  - Scale re-measured: 186 worktrees / 88 agent-memory trees / 26 file-bearing / 70
    files at the 2026-09-02 snapshot (30 per-agent `MEMORY.md` indices + 40 topics;
    breakdown manager-develop 12, manager-spec 9, plan-auditor 6, sync-auditor 1,
    manager-lead 1, manager-docs 1). Drifting population — re-count minutes later read
    73 (31 + 42). Stale 5-of-156 superseded.
- **Baseline-attribution**: all commands run 2026-09-02 from
  `.claude/worktrees/t223` at `c0c36c421`, outputs observed verbatim in this session.
- **Gaps**: mirror runtime behavior and drain execution are design-only at plan phase
  (green-path cells are run-phase fixture tests per verification-completeness §2); the
  former `phase`-target assumption was resolved at kickoff — `phase: "v3.1.4"` (operator
  decision 2026-09-02, card joins the v3.1.4 close, PR #1685); the t209 disposal path is
  inference (tree absent, reaper never fired — which unhookable path killed it is
  unobservable post hoc).
- **Residual-risk**: the write-time mirror sees Write/Edit tool calls only (Bash-written
  memory bypasses it — accepted blind spot, spec.md §G); the index-append race window is
  tolerated and doctor-detectable; the t209 concrete instance is already lost and this
  SPEC prevents recurrence, not that loss.

## §E.2 Run-phase Evidence

Milestones M1→M3 executed autonomously (kickoff 2026-09-02, design (c), serial in-session
progression). Commits on `WT-agent-memory-drain`: `05fd81ad0` (M1), `47986a7af` (M2),
`2da662690` (M3). Branch never pushed (no upstream — integration is the lead's window).

### AC binary matrix

| AC | Status | Deciding command (this tree) | Observed |
|----|--------|------------------------------|----------|
| AC-AM-001a (preview default) | PASS | `go test ./internal/cli/ -run 'TestMemoryDrainListedInHelp|TestMemoryDrainPreviewDefaultWritesNothing' -count=1 -v` | both `--- PASS`; preview names the copy set, announces preview, primary store absent after (stat ENOENT). Full runs archived: `.moai/reports/t223/e1-cli-tests.txt` |
| AC-AM-001b (json) | PASS | `go test ./internal/cli/ -run TestMemoryDrainJSONRecords -count=1 -v` | `--- PASS` — single JSON array, record carries path/agents/files/copied/collided/skipped |
| AC-AM-002 (backfill copies) | PASS | fixture `TestDrainTreeCopiesTopicFileAndKeepsSource` (`--- PASS`, e1-hook-tests.txt) + real backfill | `moai memory drain --yes --json` (binary from this tree): 25 trees, **38 copied, 0 collided, 0 skipped, 38 index lines**; re-run preview → "38 already present, 0 to copy" (idempotent). Post-run worktree↔primary overlap **33** (pre-run 0; drifting population — some source trees disposed between runs). Evidence: `.moai/reports/t223/backfill.json`, `backfill-preview.txt` |
| AC-AM-003 (never overwrite) | PASS | `go test ./internal/hook/ -run TestDrainTreeCollisionNeverOverwrites -count=1 -v` | `--- PASS` — primary keeps content A, `feedback_dup.wt-<tree>.md` carries B; refresh test pins slot semantics |
| AC-AM-004 (exactly one index line) | PASS | `go test ./internal/hook/ -run TestDrainTreeAppendsExactlyOneIndexLine -count=1 -v` | `--- PASS` — delta exactly 1, worktree-index-derived, pre-existing line untouched, second drain adds 0. Regression guard: post-backfill doctor → 8 orphans, **zero overlap with the 38 copied files** (`.moai/reports/t223/post-backfill-doctor.txt`) |
| AC-AM-005 (write-time mirror) | PASS | `go test ./internal/hook/ -run TestMirrorAgentMemoryCopiesToPrimary -count=1 -v` + live dogfood | `--- PASS`; live: real binary from this tree, `printf '{...Write...}' \| moai hook post-tool` against a /tmp fixture worktree → copy landed in fixture primary + index appended, `HOOK_EXIT=0` (`.moai/reports/t223/dogfood-mirror-live.txt`) |
| AC-AM-006 (fail open) | PASS | `go test ./internal/hook/ -run TestMirrorFailsOpenOnUnresolvablePrimary -count=1 -v` + live | `--- PASS` (stderr notice, no propagation); live non-repo fixture → `[memory-mirror] ... (fail-open; the write itself succeeded)`, `HOOK_EXIT=0`, no block decision (`dogfood-failopen-live.txt`) |
| AC-AM-007 (never deletes) | PASS | `go test ./internal/hook/ -run TestDrainTreeCopiesTopicFileAndKeepsSource -count=1 -v` | whole-fixture-tree checksums identical before/after (`requireEqualMaps`) |
| AC-AM-008 (primary no-op) | PASS | `go test ./internal/hook/ -run TestMirrorNoOpInPrimarySession -count=1 -v` | `--- PASS` — mirrored=false, zero `.wt-*` self-copies |
| AC-AM-009 (native store untouched) | PASS (regression-guard) | `find <profile>/projects -maxdepth 2 -type d -name memory \| grep -c worktree` | **0** (grep exit 1), re-measured after M2 — baseline 0 preserved |

### Full suites (touched packages)

- `go test ./internal/hook/... -count=1` → all `ok` (0 non-ok lines), this tree `2da662690`.
- `go test ./internal/cli/ -count=1` → `ok ... 330.5s`.
- `go test ./internal/template/... -count=1` → all `ok` (mirror-parity incl.).

### Cross-platform / vet / lint / coverage / boundary

- `go build ./...` exit 0; `GOOS=windows|linux|darwin go vet ./internal/hook/ ./internal/cli/` → no findings (3-OS).
- `golangci-lint run ./internal/hook/... ./internal/cli/...` → `0 issues.` (baseline was 0 — zero NEW).
- Coverage (go tool cover -func, new code): hook agent-memory files **88.6%** (15 fns), cli memory_drain.go **92.5%** (5 fns).
- Subagent boundary: `grep 'AskUserQuestion\|mcp__askuser'` over the new non-test files → 0 matches.
- SPEC-WORKTREE-REAPER-001 untouched: change set vs base `c0c36c421` contains 0 files under `internal/cli/worktree`, `internal/session/`, `internal/core/git`.

### RED evidence (E8, verbatim archives)

- `.moai/reports/t223/red-m1-hook.txt` (12 FAILs: all drain-core tests against the stub), `red-m1-cli.txt` (5 FAILs: unknown command/flags), `red-m2.txt` (undefined seam/wrapper + wiring test).

### Touched packages (for merge-tree re-measure)

`internal/hook`, `internal/cli`, `internal/template` (template mirror + catalog rebuild via `make build`).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-02
run_commit_sha: 2da662690
run_status: audit-ready
ac_pass_count: 10
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run (card worktree, integration is the lead's window; branch unpushed by design)
l44_post_push_fetch: not-run (no push — operator-forbidden)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  goos_windows: pass
  goos_linux: pass
  goos_darwin: pass
total_run_phase_files: 12
m1_to_mN_commit_strategy: one commit per milestone (05fd81ad0 M1, 47986a7af M2, 2da662690 M3)
```

Notes: AC count = 10 rows (001a/001b counted separately). Backfill executed operator-approved
after M1 (plan M1 step 4); real primary store received 38 topic files + 38 index lines, 0
collisions, source worktrees untouched.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-02
sync_commit_sha: "pending-backfill-t223"
sync_status: audit-ready
b12_self_test_a: 0 pre-emission grep hits for SPEC-AGENT-MEMORY-DRAIN-001 in CHANGELOG.md (grep -c, exit 1)
b12_self_test_b: 9 distinct AC ids in acceptance.md (AC-AM-001..009) — CHANGELOG entry references all 9 (7 blocker + 1 major + 1 regression-guard)
b12_self_test_c: all claimed file paths verified via ls (CHANGELOG.md, spec/plan/acceptance/progress.md, internal/hook/agentmemory.go, internal/cli/memory_drain.go)
changelog_entry_position: [Unreleased] > ### Added > first bullet
frontmatter_status_transitions:
  spec_md: in-progress -> completed (updated: 2026-09-02; status+updated only, zero body modifications)
mx_tag_compliance_check:
  agentmemory_go: @MX:ANCHOR x2 (IsAgentMemoryMDPath predicate, DrainTree) + @MX:REASON x2 — present
  memory_mirror_go: @MX:NOTE (write-time mirror wrapper) — present
  additions_required: none
```

## §F Phase 4 Mode Selection

- Recorded: 2026-09-02, lane session (card t223; lead dispatch + operator kickoff decisions relayed 2026-09-02).
- Implementation Kickoff Approval: GRANTED by the operator (2026-09-02, relayed via the lead) — run-phase entry approved, progression mode **AUTONOMOUS** (run→sync continuous; no inter-milestone approval pauses). Operator decisions recorded: (1) design (c) write-time mirror + backfill ADOPTED; (2) `phase` corrected to `"v3.1.4"` (card joins the v3.1.4 close, PR #1685) — applied at `f14f0c569`; (3) M3 docs surface = moai-memory.md WITH the Template-First mirror duty (template-source edit + `make build`); (4) backfill auto-runs immediately after M1 lands (the operator's literal "M1 착지 후" — M1 delivers `moai memory drain`, plan M1 step 4 already schedules the backfill with `--json` evidence archived).
- Plan Audit Gate: SKIP taken — final verdict PASS 1.00 (iter-2 delta re-check at `f671d6f6b`; threshold Tier M 0.80). Artifact hash: plan artifacts changed AFTER iter-1 (the fix commits `f671d6f6b` + `f14f0c569`) but the iter-2 delta re-check ran ON the fixed artifacts and passed — the most-recent verdict (PASS 1.00) is on the current hash, so the three skip conditions hold.
- Input parameters: tier M · scope ~2 Go packages (internal/cli new `memory drain` subcommand + internal/hook PostToolUse mirror) + docs (moai-memory.md + template mirror) · domains 2 (Go source, markdown docs+template) · language mix Go+markdown · concurrency benefit LOW (sequential dependency: M1 reconciliation core → M2 mirror shares the path predicate → M3 docs) · agent-teams prereqs: not requested.
- Mode evaluation: `direct` — not selected (multi-file Go implementation with AC discipline); `serial` — SELECTED; `fanout` — not selected (2 domains, coding-heavy, M2 reuses M1's shared predicate — sequential dependency, write-capable parallel fan-out not sanctioned); `sweep` — not selected (authored code, not a mechanical-uniform transform).
- Decision: serial
- Justification: coding-heavy Go work with in-milestone sequential dependencies (M2's mirror anchor is M1's shared path-predicate constant; M3 documents both). Per Anthropic's coding-task caveat, one writer via a single sequential manager-develop delegation covering M1→M3. AUTONOMOUS progression honored by continuous in-session execution; the goal engine is deliberately NOT armed (worktree goal-keying friction on record) — progression is managed by the lane session across the run→sync boundary.
