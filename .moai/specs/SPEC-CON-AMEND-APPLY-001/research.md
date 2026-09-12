# research.md — SPEC-CON-AMEND-APPLY-001

Tier L research artifact, added in revision 0.1.4 (`.moai/reports/t659/verdict.md` §13.3). It collects the codebase findings the design in `design.md` rests on. Almost all of it was measured earlier by lane-6 and recorded in the verdict and its evidence files; this file cites those measurements rather than re-running them.

Provenance labels used below:

- **Executed** — a command was run and its output is recorded in the named evidence file or verdict section.
- **Code reading** — the source was read; nothing was executed. A code reading is a hypothesis about behaviour, not an observation of it.
- **This revision** — run by the 0.1.4 author at HEAD `699bedd7c`, read-only.

## §A Measurement trees

| Tree | Role |
|---|---|
| `987eb7e40` | origin/develop the branch was created from (verdict header) |
| `ed71054d3` | local develop absorbed with `--no-ff` (verdict header) |
| `5a066994b` | absorb merge; the repro, key probe, and clause census ran here (verdict §2, §3) |
| `034d55c56` | code coordinates read for the 0.1.0 draft (`spec.md` preamble) |
| `ff11e752f` | test-file facts read for 0.1.1 (`plan.md` §B, `spec.md` §E.3) |
| `578afca87` | first read-only scan of the registry `file:` shape (`plan.md` §B) |
| `4e9273d0b` | same scan repeated (`acceptance.md` AC-CAA-025) |
| `e693f0583`, `fa966740d` | trees judged by `moai spec lint` for 0.1.2 and 0.1.3 (verdict §10.2, §12.3) |
| `699bedd7c` | HEAD for this revision |

**Unchanged since measurement — this revision, executed.**

```
git diff --stat 4e9273d0b 699bedd7c -- .claude/rules/moai/core/zone-registry.md             -> (no output)
git diff --stat 5a066994b 699bedd7c -- .moai/research/evolution-log.md                      -> (no output)
git diff --stat 5a066994b 699bedd7c -- internal/constitution internal/cli/constitution.go   -> (no output)
git diff --stat 5a066994b 699bedd7c                                                         -> 15 files changed, 1291 insertions(+)   (control)
```

The registry, the real log, and the code this SPEC changes are byte-identical at HEAD to the trees they were measured on. The control shows the command does report a difference where one exists. A `grep -n` at `699bedd7c` confirmed the code lines cited in §B, §G, §H, and §L.

## §B Repro — the apply step is a stub (verdict §2.1) — executed

Pre-check (`.moai/reports/t659/precheck-repro.txt`): 2026-09-10T18:43:21Z, load averages 11.78 14.89 19.07, one other lane's `go test` running (`internal/hook`).

Command (verdict §2.1) and output (`.moai/reports/t659/repro-package.txt`), tree `5a066994b`, `go1.26.8 darwin/arm64`:

```
unset … && go test ./internal/constitution/ -count=1 -v -timeout 300s -run '^(TestPipeline_Execute_DryRun_Success|TestPipeline_Execute_NonDryRun_AmendmentStubError|TestPipeline_applyAmendment_StubError|TestUpdateSourceFile_StubError|TestUpdateRegistryClause_StubError)$'
--- PASS: TestPipeline_Execute_DryRun_Success (0.00s)
--- PASS: TestPipeline_Execute_NonDryRun_AmendmentStubError (0.00s)
--- PASS: TestPipeline_applyAmendment_StubError (0.00s)
--- PASS: TestUpdateSourceFile_StubError (0.00s)
--- PASS: TestUpdateRegistryClause_StubError (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.357s
EXIT=0
```

These five are characterization tests of the unimplemented state, so their passing is the observation of the failure point:

- `NonDryRun_AmendmentStubError`: on an isolated fixture (a `t.TempDir()` registry, an approving double, a pinned lock path), `Execute(dryRun=false)` returns `amendment application error`.
- `applyAmendment_StubError`: the only reachable error is `source file update error`.

Stub test locations at `699bedd7c` (this revision, `grep -n`): `internal/constitution/pipeline_test.go:81`, `:145`, `:335`, `:405`, `:415`. Their replacements are fixed in `plan.md` §C.2.

Code coordinates (verdict §2.1, confirmed at `699bedd7c`):

| Location | What it shows |
|---|---|
| `internal/constitution/pipeline.go:190-216` | `applyAmendment` returns on the first `updateSourceFile` error |
| `pipeline.go:256-260` (message at `:259`) | `updateSourceFile` returns `not yet implemented` |
| `pipeline.go:264-267` (message at `:266`) | `updateRegistryClause` returns `not yet implemented` |
| `pipeline.go:133-137` | the dry-run branch returns `createLogEntry` and calls no apply function |
| `internal/cli/constitution.go:542-544` | the CLI wraps the error as `amendment failed: %w` |

Design consequence: `design.md` §E, §F, §H, §I.

## §C Evolution-log key probe (verdict §2.2) — executed

Output (`.moai/reports/t659/probe-log-keys.txt`). The probe test was deleted after the run; its copy is `.moai/reports/t659/tools/probe_log_keys_test.go.txt`, whose sha256 (`e05b230c…`) equals the original's (verdict §3).

```
READ id="LEARN-20260428-001" ruleID="" clauseBefore="" approvedBy="" approvedAtZero=true
WRITTEN:
  id: LEARN-20260911-001
  ruleid: CONST-V3R2-003
  zonebefore: 0
  zoneafter: 0
  clausebefore: a
  clauseafter: b
  canaryverdict: ""
  contradictions: []
  approvedby: human
  approvedat: 2026-09-11T00:00:00Z
  rolledback: false
  rollbackreason: ""
  rollbackat: null
ROUNDTRIP id="LEARN-20260911-001" ruleID="CONST-V3R2-003" approvedAtZero=false
EXIT=0
```

Findings:

1. A snake_case entry reads back with only `id` populated.
2. The writer emits concatenated-lowercase keys and integer zones.
3. The writer's own output round-trips, so the two sides agree only with each other.

Code reading at `699bedd7c`:

- `AmendmentLog` (`internal/constitution/amendment.go:192-220`) has no yaml tags, which is why keys become lowercase field names.
- The writer opens the log with `O_APPEND` (`internal/constitution/evolution_log.go:67`) and wraps each entry in `---` lines.
- `TestLoadEvolutionLogs` uses snake_case fixtures (`rule_id:` at `evolution_log_test.go:30`, `approved_at:` at `:38`) but asserts only `ID` (`:53-54`), so the mismatch passes today.

Design consequence: `design.md` §G.1, §G.2; REQ-CAA-014's added assertions.

## §D The real evolution log parses to zero entries (verdict §2.3) — executed

From `.moai/reports/t659/clause-census.txt` (tool `.moai/reports/t659/tools/clause_census.py`, read-only):

```
evolution_log_dash_split_parts 14
odd_segments_with_top_level_id_key 0
odd_segment_heads ['', '|', '|', '--|', '', '-|\n| 0.1.0 | 2026-05-13 | SPEC', '_End of evolution log._']
```

Structure of `.moai/research/evolution-log.md` (this revision, read-only `grep -n`):

- 61 lines.
- `## HISTORY` at line 6.
- `---` horizontal rules at lines 12 and 59.
- `## EVO-HRN-002` at line 14, followed by a fenced yaml block from line 16 to line 57.
- One line containing `|---` (the HISTORY table separator).
- The last commit touching the file is `0ac27ee4e` (2026-05-13).

The splitter cuts the table separator and the horizontal rules, so no segment carries a top-level `id`.

Code reading at `699bedd7c`:

- `LoadEvolutionLogs` splits on the substring `---` (`evolution_log.go:29`), reads segments pairwise, and skips a segment whose yaml does not unmarshal.
- `rateLimiter.Admit` (`internal/constitution/rate_limiter.go:41-48`) returns any loader error other than not-exist, and `Execute` wraps a limiter error as `layer 4 (RateLimiter) failed` (`pipeline.go:116`).

Two consequences follow from that reading, neither executed:

- **Today** the limiter admits because it sees 0 entries (verdict §5).
- **Under REQ-CAA-009** a reader error would surface as a Layer 4 failure.

Design consequence: `design.md` §G.3, §G.4, §G.5.

## §E Registry clause census (verdict §2.3) — executed

From `.moai/reports/t659/clause-census.txt`:

```
registry_entries 101
buckets {'missing_file': 0, 'clause_0': 4, 'clause_1': 97, 'clause_2plus': 0}
anchor_slug_found 97
clause_2plus_ids []
clause_0_ids ['CONST-V3R2-021', 'CONST-V3R2-022', 'CONST-V3R2-023', 'CONST-V3R2-024']
normalized_count CONST-V3R2-021 0 '[SUPERSEDED by worktree-opt-in policy — see CLAUDE.md §14 + '
(022, 023, 024 identical)
EXIT=0
```

Findings:

- Every one of the 97 live clauses occurs exactly once in its source file.
- The 4 zero-occurrence entries are the superseded CONST-V3R2-021 … 024.
- No clause occurs twice or more, and no `file:` is missing.
- The registry holds one yaml fence with 101 entries, and every clause is a one-line double-quoted string (verdict §2.3).

`anchor_slug_found 97` uses the tool's approximate slug rule. Whether it matches the validator's rule is unverified (verdict §2.3, §4).

Design consequence: exact-once replacement is attainable without anchor narrowing (`design.md` §E.3), and a one-line clause rewrite fits every current entry (`design.md` §F.1).

## §F Registry `file:` shape — executed

| Measure | Value | Source |
|---|---|---|
| `file:` lines | 101 | `plan.md` §B (`578afca87`), repeated at `4e9273d0b` (AC-CAA-025) |
| Distinct values | 17 | same |
| Beginning with `.claude/` | 87 | same; also verdict §10.3 |
| Equal to `CLAUDE.md` | 14 | same |
| Absolute (beginning with `/`) | 0 | same; also verdict §10.3 |
| Containing `..` | 0 | same; also verdict §10.3 |
| Existing on disk | 101 | same |

§A shows the registry is unchanged from `4e9273d0b` to `699bedd7c`.

Design consequence: the containment check refuses nothing on the real registry (`design.md` §C.4), and AC-CAA-025 carries these figures as a drift witness.

## §G Containment coverage gap (verdict §10.3) — code reading

No execution reproduced either escape (verdict §10.3 "실행 재현 없음"). The verdict's lane confirmed the cited lines; this revision re-read them at `699bedd7c`.

1. **Relative registry path.** `LoadRegistry` (`internal/constitution/loader.go:78`) checks containment only inside `if filepath.IsAbs(cleanPath)` (`:82`), returning `registry path %q escapes project dir %q` (`:86`). A relative value from `MOAI_CONSTITUTION_REGISTRY` or `CLAUDE_PROJECT_DIR` such as `../other` skips the check and is read relative to the process working directory.
2. **Absolute or `..` `file:`.** `applyAmendment` joins a relative `file:` with `projectDir` and uses an absolute one as given, with no containment check (`pipeline.go:192-195`, join at `:194`).
3. **Evolution-log path.** `Execute` and `applyAmendment` build the log path with a `projectDir` join and no containment check (`pipeline.go:115`, `:206`). The verdict names shapes 1 and 2; ruling G7 extends the check to this path (verdict §11 item 3).

The verdict also records that the loader's refusal holds for an absolute registry path escaping `projectDir` (`loader.go:79-88`, verdict §9). That is the boundary ruling G6 adopted and G7 generalized.

Design consequence: `design.md` §C.

## §H Registry path resolution differs between the CLI and `Execute` — code reading

`resolveRegistryPath` (`internal/cli/constitution.go:144-154`) resolves the registry path in this order:

1. `MOAI_CONSTITUTION_REGISTRY` when non-empty — read through the constant `constitutionRegistryEnvKey` (`constitution.go:21`).
2. `<CLAUDE_PROJECT_DIR>/.claude/rules/moai/core/zone-registry.md` when `CLAUDE_PROJECT_DIR` is non-empty — read as a string literal (`:149`). A constant `config.EnvClaudeProjectDir` already exists (`internal/config/envkeys.go:353`).
3. `<cwd>/.claude/rules/moai/core/zone-registry.md`.

`Execute` instead builds `<projectDir>/.claude/rules/moai/core/zone-registry.md` itself (`internal/constitution/pipeline.go:66`).

When the environment is set, the CLI therefore validates `--before` against one registry while `Execute` would load and write another. The verdict treats this as a write accident once writing is enabled (verdict §8 scope addition (a)).

Design consequence: `design.md` §D.

## §I Absence checks with controls (verdict §2.4) — executed

| Probe | Result | Control | Control result |
|---|---|---|---|
| `grep -rn 'rejected-amendments' internal cmd pkg --include='*.go'` | 0 lines | `grep -rn 'evolution-log' … \| grep -v _test.go \| wc -l` | 13 |
| `grep -rln 'runConstitutionAmend\|newConstitutionAmendCmd\|"amend"' internal/cli --include='*_test.go'` | 0 files | `grep -rln 'newConstitutionCmd\|runConstitution' internal/cli --include='*_test.go'` | 4 files |
| `grep -rn 'SentinelAnchorNotFound' internal --include='*.go'` | 2 lines, definition only (`validator.go:27-28`) | `grep -rn 'SentinelDrift' internal --include='*.go' \| wc -l` | 6 |

Findings:

- REQ-CON-002-012 (rejected-amendment logging) is unimplemented.
- No `internal/cli` test exercises the amend command.
- The anchor sentinel is defined but never emitted, so there is no existing anchor-range rule to reuse (verdict §4).

SPEC-V3R2-CON-002 carries `status: implemented`, although its REQ-CON-002-011 (three-file atomic apply) is not implemented (verdict §2.4).

## §J Test-suite facts — executed earlier, partly repeated in this revision

| Fact | Source |
|---|---|
| No `t.Parallel()` in `internal/constitution/pipeline_test.go`, `evolution_log_test.go`, `internal/cli/constitution_test.go`, `constitution_integration_test.go`, `constitution_guard_test.go`, so `t.Setenv` is usable | `grep -c` at `ff11e752f` (`spec.md` §E.3); the two `internal/constitution` files repeated in this revision at `699bedd7c`: 0 and 0 |
| Every existing `Execute` test passes a `Before` equal to its fixture clause, so REQ-CAA-017 breaks none | code reading of `pipeline_test.go` lines 108-262 at `ff11e752f` (`plan.md` §B) |
| `writeTestRegistry` points every entry at a non-existent `dummy.md`, so dry-run tests turn RED once validation runs | code reading (`plan.md` §C.2, R-1) |
| `NewHumanOversight` reads `os.Stdin` | code reading, `internal/constitution/human_oversight.go:22` at `699bedd7c` (`spec.md` §E.4 cites `:20-24`) |

## §K Lint evidence for earlier revisions (verdict §10.2, §12.3) — executed by the lane

`moai spec lint SPEC-CON-AMEND-APPLY-001` printed `0 error(s), 0 warning(s)` with `LINT_EXIT=0` and one INFO (`OwnershipTransitionUnmeasured`) on both occasions:

- 0.1.2 on tree `e693f0583` (`lint-0.1.2.txt`)
- 0.1.3 on tree `fa966740d` (`lint-0.1.3.txt`)

The judging binary was `v3.2.0-rc.7 moai_cp/20260910_130400-275-ged71054d3-dirty` (`lint-binary.txt`, `lint-binary-0.1.3.txt`). Neither result is evidence for 0.1.4; the lane runs lint after this revision's commit.

## §L Out-of-scope observations

| # | Observation | Location | Provenance | Disposition |
|---|---|---|---|---|
| 1 | No code writes `.moai/research/rejected-amendments/` (REQ-CON-002-012) | §I | executed | separate card candidate (verdict §7.1 item 1; `spec.md` §F) |
| 2 | SPEC-V3R2-CON-002 reports `status: implemented` while REQ-CON-002-011 is not | §I | executed | separate card candidate (verdict §7.1 item 2) |
| 3 | The CLI reads `MOAI_CONSTITUTION_DRY_RUN == "true"` while SPEC-V3R2-CON-002 REQ-CON-002-031 and AC-CON-002-07 say `=1` | `internal/cli/constitution.go:470` | code reading, line confirmed at `699bedd7c` | follow-up candidate (verdict §8.1 item 3) |
| 4 | EVO-HRN-002 names `const_registry_entry: CONST-V3R2-153`, whose registry `file:` is `.claude/rules/moai/workflow/session-handoff.md`, while its `target_file` is `.claude/rules/moai/design/constitution.md` | `zone-registry.md:678-681` | code reading (verdict §8.1 item 4) | follow-up candidate; REQ-CAA-008 maps the value as written |
| 5 | `MarkRolledBack` / `rewriteEvolutionLog` rewrite the whole log in machine format, so once human entries are readable a rollback would lose their other fields | `evolution_log.go:88` | agent reading, unmeasured per verdict §8.1 item 5; `spec.md` §F cites a grep at `034d55c56` printing only the definition | follow-up candidate; latent (0 production callers by that grep) |
| 6 | The default lock path `.moai/research/.amendment.lock` is relative to the process working directory, not `projectDir` | `pipeline.go:227` | code reading | recorded, not fixed (`spec.md` §F); contributes to the G5 reduction |
| 7 | A symbolic link swapped between the containment check and the write | — | design consideration, not measured | not a requirement (`plan.md` R-9) |
| 8 | The installed `moai` binary exited 137 on `moai version` after a replacement | `/Users/goos/go/bin/moai` | reported in verdict §9 | lead checks; the lane does not reinstall; not SPEC content |

## §M Gaps carried into plan-audit

- **CLI-level execution not observed.** No binary ran `moai constitution amend` against a fixture; CLI output wording and exit status are code reading (verdict §4). The non-dry-run CLI path remains the approved gap G5 (`spec.md` §E.4).
- **The two containment escape shapes of §G were not reproduced by execution** (verdict §10.3).
- **The anchor slug count is approximate** (§E).
- **The lint binary's `-dirty` build content is unobserved** (verdict §10.2, §12.3).
- **The `MarkRolledBack` caller count** rests on a grep recorded at `034d55c56` and an agent reading (§L row 5).
- **Symbolic-link rows may skip on platforms that refuse links**; such skips are gaps, not passes (verdict §12.5).

## §N Cross-references

- `.moai/reports/t659/verdict.md` §2 (repro evidence), §3 (baseline attribution), §4 (gaps), §7.1 and §8.1 (follow-up candidates), §10 and §12 (lane checks and lint), §13 (Tier L)
- `.moai/reports/t659/precheck-repro.txt`, `repro-package.txt`, `probe-log-keys.txt`, `clause-census.txt`, `tools/clause_census.py`, `tools/probe_log_keys_test.go.txt`
- `design.md` — the decisions these findings support
- `spec.md` §A (problem), §E.3 and §E.4 (verification route and gaps), §F (exclusions)
- `plan.md` §B (known issues), §C (pre-flight)
