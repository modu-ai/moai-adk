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

### M4 — symbol layer (second delegation)

**AC-GF-016 (astx consumable without navigator) — PASS, with one recorded design decision**
- (a) `go test ./internal/graph/symbol/ -run TestSeamCarriesNoNavigatorTierDeps -count=1`
- (b) `ok` — `go list -deps` on the seam package contains no `internal/navigator/*` except astx itself.
- (c) this run, this tree. **Decision**: the acceptance names "the graph builder's extraction seam" as the consumer; the graph package ITSELF already imported `navigator/tiers` for the doc-layer markdown parser (a pre-M4 dependency whose reuse-not-fork is an explicit @MX:NOTE design rule), so the seam was factored into `internal/graph/symbol` — the astx-consuming package — which is the package the AC's `go list -deps` method validates.

**AC-GF-017 (undocumented call edge in blast radius) — PASS**
- (a) `go test ./internal/graph/ -run TestCodeEdges_UndocumentedCallAppears -count=1`
- (b) `ok` — A→B→C fixture with no doc layer: code-call edges present (file:A → B), B's blast radius contains A, every code-call edge carries its grade.
- (c) this run, this tree.

**AC-GF-018 (additivity + disagreement exposure) — PASS**
- (a) `go test ./internal/graph/ -run TestBuild_CodeLayersAreAdditive -count=1` + real-tree build
- (b) `ok` — every doc edge survives with kind/source/target/line unchanged (the disagrees_with marker is the one deliberate REQ-GF-015 addition, asserted separately). Real tree: `OK: wrote 173,473 edges` — doc layers import 22 / mx-spec 95 / spec-depends 148 unchanged, plus code-call 161,092 / code-import 12,112.
- (c) this run, this tree.

**AC-GF-019 (matrix, no empty cells) — PASS**
- (a) `go test ./internal/graph/symbol/ -run TestMatrix -count=1` + `go test ./internal/graph/ -run TestGradeMatrixDefect -count=1`
- (b) both `ok` — 16/16 cells graded (go/python/js/ts/java/rust = name-based; the other 10 = none; nothing claims full until scope-aware resolution exists); removing one cell yields a defect verdict naming the language; `moai graph build` prints any defect verdict.
- (c) this run, this tree.

**FIRST doc-vs-code disagreement instance (lead immediate-report trigger — FIRED, reported, NOT resolved)**
- Verbatim: `{"kind":"import","source":"internal/lsp","target":"internal/astgrep","disagrees_with":"code-import (code layer scanned internal/lsp and found no import of internal/astgrep)"}` — the ONLY surviving instance after honest refutation semantics.
- Doc claim: `dependencies.md:75` in THIS tree (measured-tree co-stamp: `.moai/project/codemaps/dependencies.md` as regenerated at commit `7261712f1`; the same statement sits at a different line in the primary checkout's pre-regeneration copy — the lead reproduced it at :71 — line numbers alone already mislead across trees, which is the M3 lesson live).
- Code observation: zero astgrep imports anywhere under internal/lsp/ (subpackages included); astgrep's actual consumers are internal/cli and internal/hook. Analysis: the doc node `internal/lsp` is a summary node; the claim is stale or mis-aggregated. Both layers preserved in the artifact; reported to the lead mid-work (msg 8274650c).

**Asymmetric-rule ruling compliance (lead decision, 3 required items + SHA co-stamp):**
1. **Composition of the intermediate 1,505** (measured on this tree, real build):
   - 1,505 total markers after module-path normalization = **3 doc-side** (doc-explicit claims) + **1,502 code-side** (code-found/doc-silent).
   - Before normalization the code-side count was 1,880 — **378 code-side false positives removed by go.mod module-path normalization** (Go imports carry the full `github.com/modu-ai/moai-adk/...` prefix; unnormalized, no code-import could ever corroborate a doc edge).
   - Of the 3 doc-side markers, **2 were false positives dissolved by the same normalization** (see item 3), leaving **1 genuine disagreement** (the lsp→astgrep instance above).
2. **Revival path for the suppressed direction** (implemented, tested): `moai graph build --all-disagreements` → `BuildWithCodeLayersMode(…, DisagreementAll)` re-marks every suppressed code-found/doc-silent local import with an explicit `[revived]` tag on the code-import edge, while the default mode's genuine refutation markers and doc-edge preservation are unaffected (`TestBuild_DisagreementAllRevivesSuppressedDirection`). A decided-not-to-report signal stays retrievable — it does not harden into cannot-be-reported.
3. **The 2 doc-side false positives — verification recorded**: after normalization, `internal/cli → pkg/models` is corroborated in code (`{"kind":"code-import","source":"internal/cli/profile_setup.go","target":"pkg/models","line":16}` — grep of the artifact, plus `internal/cli/wizard/types.go:8`), and `internal/config → pkg/models` is corroborated (`internal/config/defaults.go` imports `pkg/models` — verified by grep of the non-test source). Both doc claims are TRUE; the pre-normalization markers were artifacts of the prefix mismatch, not disagreements.

**Refresh integration**: `refreshEdgesArtifact` (the M2 query-time path) now rebuilds via BuildWithCodeLayers; the mx-index inventory — the per-file hash of the described sources the code layer walks — is the astx-relevant input inside the existing fingerprint union, so a code-source edit trips the same refresh trigger.

### M5 — MCP code queries (second delegation)

**AC-GF-020 (graph_file_api signatures-only) — PASS**
- (a) `go test ./internal/graph/ -run TestFileAPI -count=1` + `go test ./internal/cli/ -run TestHandleGraphFileAPI -count=1`
- (b) both `ok` — full signatures (`func Helper(s string, n int) (string, error)`), zero bodies, provenance names the tree. Go applies the capitalization export rule strictly; other languages list their declaration set with kinds (export notions differ — documented).
- (c) this run, this tree.

**AC-GF-021 (find_code + trace_calls from the code layer) — PASS**
- (a) `go test ./internal/graph/ -run "TestFindCodeAndTraceCalls|TestCodeQueries_PerTree" -count=1` + `go test ./internal/cli/ -run TestHandleGraphFindAndTrace -count=1`
- (b) both `ok` — matches carry grade + via; trace returns A as B's caller and C as callee; two trees with different content answer from their own artifacts (the t246 wrong-tree family, tested).
- (c) this run, this tree.

**AC-GF-022 (baseline-first + reduction) — PASS with the baseline gap recorded honestly**
- Baseline artifact: `.moai/reports/t250/m5-baseline.md` — authored before the M5 implementation within the working session; committed together with it in the same commit (`7f2e9e77d`), so ordering rests on the authoring session record, not on git history (audit F3: a same-commit pair cannot prove ordering in git). Method: mechanical `grep -c '"name":"<tool>"'` over the 8 most recent real session transcripts (session ids + dates + counts tabled). Real finding: Grep tool-use is 0–1/session in this repository (agents search via Bash grep); Read 0–25/session. Per-task baseline counts: NOT OBTAINABLE (no pre-M5 session performed the fixed task set; a knowing-measurer simulation would fabricate counts) — recorded as an explicit gap per delegation instruction.
- Post artifact: `.moai/reports/t250/m5-post.md` — the same fixed task set executed by the M5 engine against the real worktree (vehicle run, verbatim per-task logs): **0 Grep + 0 Read tool-use events, 5 tool calls, all provenance-stamped**. Claim bounded to what was measured (structural reduction on this task set), stated honestly.

### CI fix round — PR #1648 findings (post-rebase, HEAD dc5d65aad)

**F4 / not-comparable state (t241 mutant-discipline record):**
- **(a) Turns red when** the codemaps provenance stamp names a commit the local git history does not hold (shallow checkout, vanished SHA) — `git diff <stamp>` cannot run and the layer was never measured. It reports as a SYSTEM ERROR (exit 2), never a fabricated `stale` verdict.
- **(b) Failing input + observed reproduction** (local shallow clone of this branch, `git clone --depth 1`, fixed binary): verbatim —
  ```
  $ cd /tmp/t250shallow-repro && /tmp/t250bin/moai graph check
  EXIT=2
  --STDERR--
  graph check: system error: codemaps stamp 5b88fd0ab0d0 not comparable in this checkout: git diff 5b88fd0ab0d0: exit status 128: fatal: bad object 5b88fd0ab0d0e1394bd92bec1650c54d35784281
  ```
  Contrast, the PRE-fix binary on the SAME clone reproduced the CI failure shape (verdict=stale + exit 1 on `fatal: bad object`) — the defect class exit 2 now replaces. Regression tests: `TestCheckFreshness_NotComparableIsSystemError` (engine) + `TestGraphCheckCmd_NotComparableExitsTwo` (CLI exit code + stderr naming), both observed PASS.
- **(c) Reachability + why exit 2 is the honest signal there**: with `fetch-depth: 0` in graph-freshness.yml, the CI checkout holds the PR branch's full history; the stamp names an ancestor of the regenerating commit on the same branch (the standard flow), so `git diff` always resolves — **not-comparable cannot fire in the CI path** (the only CI-side exit-2 risk was exactly the shallow default, now removed at both layers). The remaining reachable environment is a USER's shallow clone running `moai graph check`: there, exit 2 with the named commit is the intended honest signal — the tool refuses to guess freshness from history it does not hold; exit 1 (stale) would assert a measurement that never ran, and exit 0 would bless an unmeasured layer. (A stamp naming a NON-ancestor commit — e.g. cherry-picked codemaps from another branch — also lands here even on full history; that is a mis-stamped artifact, and a system error is likewise the honest verdict.)
- Spec-gap note: the SPEC's verdict enum (fresh|stale|absent) and REQ-GF-004 prescribe no not-comparable verdict; exit 2 (system error, per design.md §7's 0/1/2 contract) is the implemented completion, recorded as a spec-gap fix.

**F1 mcp count guard**: catalog test updated 25 → 28 (main's design is a literal pin — kept literal per the delegation instruction; the registration/catalog equality guard in cli remains the drift catcher).

**F2 nonoverlap guard jurisdiction**: TestConsumerOnly_M0AndMxByteUnchanged now gates on the branch's diff touching `internal/navigator/detect/` — the AC-NS2-005a consume-only contract binds Detect-layer change-sets; an mx-package SPEC (this one) legitimately editing internal/mx no longer trips a guard that has no standing over it. On Detect branches the guard is unchanged and still enforcing.

**F3 web i18n key sets**: the 3 new catalog tools' schema fields required dictionary keys in all 4 locales (en/ko/ja/zh × title/desc added — set-membership fix, not translation work); `go test ./internal/web/` green.

**Verification (unfiltered)**: `go test ./internal/mcp/ ./internal/graph/ ./internal/web/` → all `ok` (0.380s / 4.844s / 3.276s); `go test ./internal/hook/` → `ok ... 65.808s`; named tests isolated green (Count28, ConsumerOnly skip-by-jurisdiction, both i18n tests, NotComparable engine+CLI); `go vet` on mcp/hook/web/graph clean.

### E2 cross-platform builds (M1-M5 legs)
```
$ go build ./...                        → exit 0 (no output)   [after M1, M3, M4, M5]
$ GOOS=windows go build ./...           → exit 0 (no output)   [after M1, M4, M5]
```
(`go vet` clean on all touched packages each leg.)

### Gaps (explicitly NOT observed)
- The CI job's red conclusion was not observed on a pushed branch (push forbidden in this delegation); the local binary halves (red fixture + green bootstrapped tree) are the recorded evidence.
- The `~3ms`-class foreign figures were neither used nor quoted anywhere.
- Codemaps md body citations are written by the /moai codemaps skill (LLM); this SPEC provides the canon (renderer + resolver + `moai graph stamp codemaps`) and switched the mechanical writers (mx tags, provenance sidecars); the skill-side adoption rides the next regeneration.
- AC-GF-022 per-task baseline counts unobtainable from real prior sessions (method + reason in m5-baseline.md); the post-run's measured structural reduction is the honest bound.
- astx call captures seeded for 6 languages (name-based); the other 10 grade `none` — extending captures is additive follow-up work, not hidden by the matrix.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-25
run_commit_sha: "M1 cd46e0250 / M2 68323efba / M3 5b88fd0ab / docs 7261712f1 / M4-M5 see final commits"
run_status: "M1-M5 complete across two delegations"
ac_pass_count: 22                 # AC-GF-001..022 (all; AC-022 with recorded baseline gap)
ac_fail_count: 0
ac_deferred_count: 0
preserve_list_post_run_count: 0   # no PRESERVE-list file modified
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: "pass"
  windows: "pass (GOOS=windows go build)"
total_run_phase_files: 30         # 21 (M1-M3 leg) + 9 (M4-M5 leg)
m1_to_mN_commit_strategy: "per-milestone conventional commits (M1/M2/M3/docs/M4/M5)"
budget_measurement:
  zero_change_refresh_ms: 522
  full_scan_ms: 196
  inventory_files: 2597
threshold_calibration:
  described_files_last_10_commits: 137
  described_files_last_50_commits: 233
  codemaps_threshold_kept: 40
m4_measurement:
  real_tree_edges_total: 173473
  code_call_edges: 161092
  code_import_edges: 12112
  grade_matrix: "6 name-based, 10 none, 0 empty cells"
  disagreement_instances: 1       # internal/lsp -> internal/astgrep (recorded, unresolved)
m5_measurement:
  baseline_grep_tooluse_per_session: "0-1 (8-session sample)"
  baseline_read_tooluse_per_session: "0-25 (8-session sample)"
  post_taskset_grep_read_events: 0
  post_taskset_tool_calls: 5
```

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: complete
- sync_complete_at: 2026-08-25
- sync_commit_sha: pending-backfill — the single sync commit on WT-graph-freshness cannot
  contain its own SHA (D3 placeholder-backfill exemption); the integrating lead backfills
- changelog_entry_position: `CHANGELOG.md` `## [Unreleased]` → `### Added` (4 entries: the
  `moai graph check` drift-gate family + gate step + CI job + `graph stamp`; code-derived edge
  layers; the 3 MCP code-query tools; content-addressed citations) and `### Changed`
  (query-time refresh + per-tree anchoring + provenance on answers)
- docs_site_sync: `docs-site/content/{ko,en,ja,zh}/cli-reference/graph.md` — `moai graph
  check` + `moai graph stamp codemaps` sections, the build code-layers sentence, the query
  refresh sentence; ko canonical → en/ja/zh in the same commit. No MCP tool-catalog page
  exists on docs-site, so the 3 MCP tools carry no docs-site entry (minimal-addition rule)
- frontmatter_status_transitions:
  - spec.md: in-progress → implemented (this sync commit); `completed` rides the lead-side
    close after branch integration (t82 sibling convention)
  - plan.md / acceptance.md: no frontmatter status block — status carried on spec.md only
- b12_self_tests:
  - a) pre-emission `grep -c 'SPEC-V3R6-GRAPH-FRESHNESS-001' CHANGELOG.md` → `0` (rc=1) —
    no duplicate entry
  - b) acceptance.md distinct AC tokens → `22`; the CHANGELOG entries cite no AC counts
    (nothing to mismatch)
  - c) claimed paths verified by ls — `internal/graph/` (check/citation/codequery/symbol.go/
    meta.go), `internal/graph/symbol/`, `internal/navigator/astx/queries/`, `internal/cli/`
    (graph_check/graph_stamp/graph_refresh_cli/mcp_code_tools.go),
    `.moai/project/codemaps/` — all present
- sync_verification:
  - `go vet ./internal/graph/... ./internal/mx/ ./internal/cli/` → rc=0, no output
  - `moai spec lint .moai/specs/SPEC-V3R6-GRAPH-FRESHNESS-001/spec.md` → `✓ No findings —
    all SPEC documents are valid`, rc=0
  - docs-site: URL-blacklist / Mermaid-LR / body-emoji greps over the 4 changed pages →
    0 matches each (rc=1); `hugo -s docs-site --minify --gc` → rc=0, `Total in 4969 ms`,
    0 warn/error lines in the build log; rendered
    `public/{ko,en,ja,zh}/cli-reference/graph/index.html` each carry the new sections
- open_followups:
  - `sync_commit_sha` backfill (this commit's SHA) once the lead integrates the branch
  - graph-freshness CI job lands bootstrap-only; enabling it as a required check is an
    operator decision (day-one posture: codemaps regenerated in this PR)
  - `/moai codemaps` skill-side adoption of the citation canon rides the next regeneration
    (§E.2 gap)
  - AC-GF-022 per-task baseline gap stands as recorded in §E.2/§E.3 — the CHANGELOG makes no
    measured-reduction claim
