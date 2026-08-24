# Progress — SPEC-V3R6-GRAPH-FRESHNESS-001

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set: Tier L 5-file set (spec.md, plan.md, acceptance.md, design.md, research.md) + progress.md — Tier L justified by 5 milestones spanning 4 subsystems (internal/graph, internal/cli, internal/navigator/astx, internal/hook/quality + CI + MCP) and 2 cross-cutting conventions; the per-layer metric and cache-anchoring decisions need design.md, and the graft-analysis provenance + anchor re-verification live in research.md so the run phase never re-derives them.
- Requirements: 20 REQ (GEARS), all pattern-annotated; 22 AC (Given-When-Then) + §D.5 mutant-coverage table; REQ↔AC traceability 100% both directions (acceptance.md §D.2).
- Frontmatter: 12 canonical fields + `tier: L` + `era: V3R6`; SPEC ID regex-validated (`PASS`) before write; ID uniqueness verified against the catalog (research.md §3).
- Split decision: single SPEC — 20 REQ / 22 AC sit inside the Tier L ceilings (25/25) with slack; M1-M3 and M4-M5 share the provenance/cache substrate (M2's cache feeds M4/M5), so a split would duplicate that substrate in both halves.
- Evidence base: `.moai/reports/graft/graft-analysis-20260824.md` (read, not re-derived) + anchor re-verification at `baa100ce5` (research.md §2) — drift re-measured 740 commits; mx-index provenance absence and fresh-worktree artifact absence are directly measured facts.

## §E.2 Run-phase Evidence

Scope of this delegation: **M1-M3 only** (M4 symbol layer, M5 MCP code queries are a separate later delegation). All evidence below was measured in this run, against this tree (WT-graph-freshness), at the HEAD SHAs named per item.

### M1 — drift gate

**AC-GF-001 (numeric per-layer report) — PASS**
- (a) `go test ./internal/graph/ -run TestCheckFreshness_AllFresh -count=1`
- (b) `ok  github.com/modu-ai/moai-adk/internal/graph	3.707s` — asserts 3 layer reports, each carrying layer name + metric kind + integer value 0 + threshold + verdict `fresh`, and `Failed() == false`.
- (c) this run, this tree, at M1 GREEN (pre-commit `a1b1ca696..` → M1 commit below).

**AC-GF-002 (per-layer metric correctness) — PASS**
- (a) `go test ./internal/graph/ -run TestCheckFreshness_PerLayerMetrics -count=1`
- (b) part of `ok ... 3.707s`: codemaps value must equal exactly 3 (endpoint diff), mx-index exactly 1 (inventory diff — only the one inventoried file), edges stale via its own fingerprint metric. Each layer's value derives from its own metric column, not a shared formula.
- (c) this run, this tree, M1 GREEN.

**AC-GF-003 (provenance blocks present) — PASS**
- (a) `go test ./internal/graph/ -run TestCheckFreshness_NoProvenanceIsNotFresh -count=1` + real-tree run below
- (b) no-provenance codemaps reports `absent` (unjudgeable), never fresh. Real-tree: `go run ./cmd/moai graph stamp codemaps` → `provenance: tree=...wt250 commit=5b88fd0ab0d0`; sidecar blocks carry tree root + commit-or-dirty + per-layer data (FileInventory / SourceFingerprints / DescribedRoots).
- (c) this run; stamp output verbatim above.

**AC-GF-004 / MUTANT A (exit-code discipline) — PASS, mutant killed with observed evidence**
- (a) `go run ./cmd/moai graph check --root /tmp/t250mutA` (real process, fixture aged past a threshold-2 gate.yaml override)
- (b) verbatim:
  ```
  codemaps  metric=described-source-diff value=2 threshold=2 verdict=stale
  graph check: layer codemaps verdict=stale value=2 threshold=2 —
  exit status 1
  ```
  Restored input → codemaps fresh; absent untracked layers still exit 1 (absent clause). Process exit code 1 observed directly, not asserted. CLI-level: `TestGraphCheckCmd_StaleExitsOne` asserts `errors.As` ExitCoder == 1.
- (c) this run, real binary, /tmp/t250mutA fixture.
- **(a) red-when**: any layer over threshold or absent · **(b) failing input**: codemaps stamped at HEAD + 2 described-source files changed (threshold 2 via fixture gate.yaml) · **(c) reachability**: CI job and `moai gate` consume only the exit code; a reporting-only mutant silently disarms both (740-commit silent-drift class).

**AC-GF-005 (absent is distinct + failing) — PASS**
- (a) `go test ./internal/graph/ -run TestCheckFreshness_AbsentLayers -count=1` + `go test ./internal/cli/ -run TestGraphCheckCmd_AbsentExitsOne -count=1`
- (b) both `ok`; absent verdicts carry explicit reason strings ("untracked runtime artifact — fresh worktree state"), JSON carries `"absent"`, exit 1.
- (c) this run, this tree.

**AC-GF-006 (gate integration, notice-not-silence) — PASS**
- (a) `go test ./internal/hook/quality/ -run TestGateGraphFreshness -count=1`
- (b) `ok  github.com/modu-ai/moai-adk/internal/hook/quality	1.899s` — blocking+stale fails naming the step; advisory+stale passes with the verdict emitted; disabled emits an explicit skip notice; the step runs before language detection (no-marker fixture still notices). Deviation recorded: the UNCONFIGURED (nil) posture stays silent — the pre-existing unknown-project silent-pass contract (TestQualityGate_Run_UnknownProjectPasses) is preserved; production paths always populate the config.
- (c) this run, this tree.

**AC-GF-007 (CI job actually turns red / healthy head green) — job landed; local halves verified**
- (a) workflow: `.github/workflows/graph-freshness.yml` (bootstrap → check). Local healthy-head reproduction: `go run ./cmd/moai mx scan --quiet && go run ./cmd/moai graph build && go run ./cmd/moai graph check` on the real worktree
- (b) verbatim:
  ```
  codemaps  metric=described-source-diff value=0 threshold=40 verdict=fresh
  mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=fresh
  edges     metric=source-fingerprint-mismatch value=0 verdict=fresh
  ```
  Red half: the MUTANT A fixture above is the same check the CI step runs, exit 1 observed. **Gap**: the job's red conclusion is verified on the local binary, not on a pushed non-main branch CI run (push is out of this delegation's scope — "do NOT push"); the workflow triggers on PR so the delivering PR's CI exercises it.
- (c) this run, real worktree + /tmp/t250mutA.

**REQ-GF-002 no-mtime discipline (supports AC-GF-002)**
- Static inspection: `grep -n "ModTime\|mtime" internal/graph/check.go internal/mx/provenance.go internal/mx/refresh.go` → no match; all signals are content hashes / git endpoint diffs / fingerprints.

**Threshold calibration (acceptance §D.7) — measured**
- (a) `git log -10 --name-only --pretty=format: -- internal cmd pkg | sort -u | wc -l` → `137`; same for `-50` → `233`
- (b) verbatim counts above.
- (c) this run, this tree (HEAD 5b88fd0ab). Decision: threshold 40 retained — ≈2-3 commits of typical described-source churn, so routine small PRs pass while accumulated drift (the 740-commit failure mode) reds. No threshold was raised to make a check pass.

### M2 — query-time refresh

**AC-GF-008 / MUTANT B (refresh actually re-reads) — PASS, mutant killed with observed evidence**
- (a) `go test ./internal/mx/ -run TestRefreshIndex_ReflectsUncommittedEdits -count=1` (GREEN) and, with the stamp-only mutation injected (drop freshly parsed tags):
- (b) verbatim RED under the mutant:
  ```
  --- FAIL: TestRefreshIndex_ReflectsUncommittedEdits (0.14s)
      refresh_test.go:124: refreshed index lost the b.go tag entirely
  ```
  Graph-side guard: `TestGraphQuery_RefreshesStaleEdges` fails on an unchanged artifact after a source moved ("MUTANT B (graph side): edges artifact unchanged").
- (c) this run, this tree. **(a) red-when**: content changed between queries · **(b) failing input**: uncommitted edit to a scanned file between two queries · **(c) reachability**: agents act on query answers; stamp-only refresh propagates stale answers into audits (init.go:773/:898 lineage).

**AC-GF-009 (changed-files-only, no LLM) — PASS**
- (a) `go test ./internal/mx/ -run TestRefreshIndex_ChangedFilesOnly -count=1`
- (b) `ok` (within full-package `ok ... 19.348s`): zero-change refresh parses 0 files; exactly 2 changed files → parses exactly those 2. Interpretation note: change DETECTION hashes every walked file (content-hash discipline; mtime is banned); FilesParsed counts the re-PARSED files — the expensive change-consumption the AC's instrument targets. No-LLM/no-network: dependency inspection `go list -deps ./internal/mx` contains no LLM-client package; the refresh path is file I/O + hashing only.
- (c) this run, this tree.

**AC-GF-010 (per-tree cache isolation) — PASS**
- (a) `go test ./internal/cli/ -run TestGraphQuery_PerTreeIsolation -count=1` + `go test ./internal/mx/ -run TestRefreshIndex_WrongTreeFullRescan -count=1`
- (b) both `ok`: two fixture trees with distinct uncommitted edits answer from their own edits and name their own tree root in provenance; a wrong-tree index forces full rescan (never incremental trust).
- (c) this run, this tree.

**AC-GF-011 (update-cost budget warning) — PASS**
- (a) `go test ./internal/cli/ -run TestGraphQuery_BudgetOverrunWarns -count=1`
- (b) `ok`: a 1ms configured budget triggers a warning naming measured cost and budget; the query still answers. Budget calibration (REQ-GF-009, local measurement — no foreign figures):
  - (a) measurement vehicle (temporary test, deleted after): `T250_MEASURE_ROOT=<this worktree> go test ./internal/mx/ -run TestMeasureRefreshCost -v`
  - (b) verbatim: `zero-change refresh: parsed=0 duration=522.41925ms` / `full scan: duration=195.894875ms inventory=2597 files`
  - (c) this run, real worktree. Default 2000ms = ~4× the measured ceiling (headroom for slower machines); overrun warns, never blocks.

### M3 — citation convention switch

**AC-GF-012 (citation canon) — PASS**
- (a) `go test ./internal/graph/ -run TestCitationRenderCarriesCanon -count=1`
- (b) `ok`: rendered citation carries file + region hash + convenience L<n>; NewCitation rejects an empty excerpt (line-only anchor is not the canon).
- (c) this run, this tree.

**AC-GF-013 (mx-index hash anchoring) — PASS**
- (a) `go test ./internal/mx/ -run TestRefreshIndex_TagHashSurvivesLineDrift -count=1`
- (b) `ok` (within full-package run): 4 blank lines inserted above a tag → ContentHash identical, convenience Line tracks the physical position (3→7).
- (c) this run, this tree.

**AC-GF-014 (two-tree identical-target guarantee) — PASS**
- (a) `go test ./internal/graph/ -run TestCitationTwoTreeResolution -count=1`
- (b) `ok`: one citation resolves in both trees by content — tree A at line 3, tree B (5 blank lines inserted) at its true physical line 8; a line-anchored resolver would return 3 in both and point at the wrong place in B. Edited-region mismatch reports honestly (TestCitationRegionEditedReportsMismatch).
- (c) this run, this tree.

**AC-GF-015 (measured-tree SHA co-stamp) — PASS**
- Provenance blocks (all three artifacts) carry TreeRoot + CommitSHA-or-dirty+fingerprint; answers print `provenance: tree=... commit=...` (observed in the stamp output and query stderr assertions); Citation.TreeSHA co-stamps alongside the region hash.

### Day-one codemaps regeneration (plan.md M1 item 6 posture)

- Regenerated `.moai/project/codemaps/{dependencies,data-flow}.md` for the current tree: dependencies.md gains the `internal/graph` node + edges (cli→graph, hook→graph, graph→mx) and the infra-package table row; data-flow.md gains §7 (graph freshness flow). Both files previously pre-dated internal/graph entirely — the drift the gate exists to catch, measured as content-stale.
- Stamped via `moai graph stamp codemaps` → `provenance: tree=...wt250 commit=5b88fd0ab0d0` (clean stamp at the M3 commit; the stamp commit itself touches only codemaps, keeping described-source drift 0).
- Post-bootstrap real-tree check: all three layers `fresh` (verbatim output under AC-GF-007).

### E2 cross-platform builds
```
$ go build ./...                        → exit 0 (no output)
$ GOOS=windows go build ./...           → exit 0 (no output)
```
(Measured after M3; `go vet` clean on all touched packages.)

### Gaps (explicitly NOT observed)
- The CI job's red conclusion was not observed on a pushed branch (push forbidden in this delegation); the local binary halves (red fixture + green bootstrapped tree) are the recorded evidence.
- The `~3ms`-class foreign figures were neither used nor quoted anywhere.
- Codemaps md body citations are written by the /moai codemaps skill (LLM); this SPEC provides the canon (renderer + resolver + `moai graph stamp codemaps`) and switched the mechanical writers (mx tags, provenance sidecars); the skill-side adoption rides the next regeneration.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-25
run_commit_sha: "5b88fd0ab0d0"   # M3 code commit; codemaps-regen commit follows in the same delegation
run_status: "M1-M3 complete (M4-M5 out of this delegation's scope)"
ac_pass_count: 15                 # AC-GF-001..015 within M1-M3 scope
ac_fail_count: 0
ac_deferred_count: 7              # AC-GF-016..022 (M4/M5 — separate delegation)
preserve_list_post_run_count: 0   # no PRESERVE-list file modified
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: "pass"
  windows: "pass (GOOS=windows go build)"
total_run_phase_files: 21         # 13 new + 8 modified (see §E.2 commits)
m1_to_mN_commit_strategy: "per-milestone conventional commits (M1/M2/M3 + codemaps regen)"
budget_measurement:
  zero_change_refresh_ms: 522
  full_scan_ms: 196
  inventory_files: 2597
threshold_calibration:
  described_files_last_10_commits: 137
  described_files_last_50_commits: 233
  codemaps_threshold_kept: 40
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
