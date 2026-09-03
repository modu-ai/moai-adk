# Sync-Phase Audit Verdict — SPEC-CODEX-SKILL-PATH-001 (card t468)

Auditor: sync-auditor (independent, this run). Cross-model fan-out NOT invoked — after the
mechanical verification below, no verdict-critical question remained open that a second backend
opinion could settle (the one open item, Windows runtime classification, is structurally
unreachable from a darwin host by any backend and is already a named CI gap).

---

## Claim

**FINAL VERDICT: PASS-WITH-DEBT — overall score 91/100** (harmonic mean of dimensions;
per-dimension: Functionality 95, Security 95, Craft 82, Consistency 93).

- All 8 acceptance criteria (AC-CSP-001-01 .. -08) verified PASS — 7 by fresh test execution on
  this tree, 1 (AC-07, structural) by fresh source inspection + fixture sweep.
- All 7 PRESERVE rows re-verified fresh (code read + green tests).
- Must-pass firewall: Functionality PASS (95), Security PASS (95, zero Critical/High findings).
- The debt: **D1** — one factually wrong clause in the user-facing CHANGELOG entry (committed,
  unmerged branch). Code is correct; the doc contradicts REQ-CSP-004's Windows-host semantics.
  One-clause reword fixes it. The confirming re-audit, if routed, is scoped to D1's delta only.

## Evidence

### Dimension scores (mechanical verification, all this run)

| Dimension | Score | Verdict | Evidence (verbatim) |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | `go test ./internal/cli/ -run 'CodexSkillPath|TestCheckCodexWiring' -count=1 -v` → 27 tests: 19 `TestCheckCodexWiring_*` + 8 `TestCodexSkillPath_*`, **27 `--- PASS`, 0 FAIL, 0 SKIP**; `ok github.com/modu-ai/moai-adk/internal/cli 1.064s`. `go test ./internal/codexwiring/ -count=1` → `ok github.com/modu-ai/moai-adk/internal/codexwiring 0.745s`. Deductions: Windows test-run verdict pending CI (see Gaps); AC-04's Windows-native sub-assertion rests on code structure until that leg runs. |
| Security (25%) | 95/100 | PASS | Production-diff filesystem/exec surface, added lines only: `git diff d592b0551..b37fe44b4 -- internal/cli/doctor_codex.go internal/codexwiring/skills.go \| grep -E '^\+' \| grep -nE 'os\.|exec\.|Setenv|WriteFile|MkdirAll|Remove'` → exactly one hit: `90:+		_, serr := os.Stat(statPath)` — the diff adds a READ and nothing else. No exec of config-derived strings, no writes, no env mutation, no secrets. Config path strings (untrusted user input) are classified and stat'ed, never executed or interpolated into commands. Fail-open preserved (unresolvable home → indeterminate `doctor_codex.go:405-410`; absent config → silent `:382-385`). |
| Craft (20%) | 82/100 | PASS | `go vet ./internal/cli/ ./internal/codexwiring/` → exit 0. `golangci-lint run --timeout=4m ./internal/cli/... ./internal/codexwiring/...` → `0 issues.` exit 0 (fresh, this tree). Test craft strong: ordered-phrase assertions with forbidden transposed forms (`doctor_codex_test.go:227-235`), positive census discipline, skip-with-explicit-reason on symlink-host limits. Deduction: coverage % NOT measured this run (dispatch prohibits instrumentation runs); internal/cli package coverage also unmeasured in run-phase (10m timeout fail recorded, progress.md §E.3). |
| Consistency (15%) | 93/100 | PASS | `gofmt -l` on the 3 changed files → empty, exit 0. `grep -c 'SPEC-CODEX-SKILL-PATH-001' CHANGELOG.md` → `1` (no duplicate). `grep -c '^- AC-CSP-001-' .moai/specs/SPEC-CODEX-SKILL-PATH-001/spec.md` → `8` (Tier S ceiling respected). Seam reuse (no second resolver), English comments, file-local naming conventions all match. Deduction: D1 (deliverable-accuracy defect in CHANGELOG). |

Harmonic mean: 4 / (1/95 + 1/95 + 1/82 + 1/93) = **90.9 ≈ 91**.

### Per-AC results (8/8 PASS)

| AC | Test / method | Result | Evidence anchor |
|---|---|---|---|
| AC-CSP-001-01 | `TestCodexSkillPath_HomeRelativeExistingNotMissing` | PASS | fresh run this tree (`--- PASS (0.00s)`) |
| AC-CSP-001-02 | `TestCodexSkillPath_HomeRelativeMissingStillCounted` | PASS | fresh run (`--- PASS (0.01s)`) — expansion creates no false negatives |
| AC-CSP-001-03 | `TestCodexSkillPath_RelativeNotMissingDistinctClassification` + `_RelativeOnlyNonDestructiveFinding` | PASS | fresh run (both `--- PASS`); remove directive never attaches to relative-only configs |
| AC-CSP-001-04 | `TestCodexSkillPath_BackslashAndOtherUserHomeNotMissing` | PASS (darwin leg) | fresh run (`--- PASS (0.01s)`); backslash fragment + `~otheruser` both oddly-formed, not missing. Windows-native-absolute sub-assertion verified structurally: `filepath.IsAbs` runs first (`doctor_codex.go:334` before `:340`); runtime Windows leg = CI gap (named below) |
| AC-CSP-001-05 | `TestCodexSkillPath_AbsoluteExistingAndMissing` + `StaleHomeSkillsReported` + `HealthyHomeSkillsNoFinding` | PASS | fresh run (all `--- PASS`) |
| AC-CSP-001-06 | `TestCodexSkillPath_RealMissingWithSymlinkLoopIndeterminate` + `IndeterminateStatNotMissing` | PASS | fresh run (both `--- PASS`, **0 SKIP** on this darwin host — loops produced ELOOP and ran) |
| AC-CSP-001-07 | structural — source scan + fixture sweep | PASS | inspection this run: all 8 new tests use `writeCodexHomeConfig` (t.TempDir root) + `stubCodexHome` (pins BOTH the `codexUserHomeDir` seam and CODEX_HOME); zero real-home reads, zero writes outside temp roots. `grep -n '/nonexistent' internal/cli/doctor_codex_test.go` → no output, exit 1 |
| AC-CSP-001-08 | `TestCodexSkillPath_ExpansionUsesUserHomeSeam` | PASS | fresh run (`--- PASS (0.00s)`) — expansion observably through the USER-home seam, incl. bare `~` |

### PRESERVE fidelity (7/7 rows re-verified fresh)

| PRESERVE row | Verification |
|---|---|
| CODEX_HOME precedence via `resolveCodexHomeDir` | `doctor_codex.go:377`; `CodexHomeHonoured` + `CodexHomeConfigRead` PASS fresh |
| only `fs.ErrNotExist` feeds missing | `doctor_codex.go:433` `errors.Is(serr, fs.ErrNotExist)`; other stat errors → indeterminate `:442-449`; loop tests PASS fresh |
| silent at `missing==0 && unresolvedShape==0` | `doctor_codex.go:454-459` (includes indeterminate-only configs — t451 posture verbatim); `HealthyHomeSkillsNoFinding` + `AbsentHomeConfigSilent` PASS fresh |
| no second home resolver | `expandCodexHomeRelativePath` calls the existing seam `codexUserHomeDir` (`:353`; seam var at `internal/cli/mcp_codex.go:1753`); `grep 'os\.UserHomeDir'` in doctor_codex.go hits a comment only |
| read-only / fail-open | production diff adds only `os.Stat`; unresolvable home → indeterminate never missing; absent config → silent |
| unspecified-enabled declared split | `doctor_codex.go:434-441` tri-state; `UnspecifiedEnabledReportedSeparately` PASS fresh |
| Message/rendered-panel width bands | `MessageWidthStaysInBand` + `RenderedPanelStaysInBand` PASS fresh; new summaries routed through `joinCodexSummaries` bound |

### Interrogation of the named stress points

- **(a) `~user` heuristic** — after `~`/`~/` peel, any residual `~`-prefix → oddly-formed
  (`:340`). `~tmp/file` (literal tilde dir) also lands oddly-formed, which matches POSIX shell
  user-lookup semantics; both dispositions are non-destructive, so no false-missing path exists.
  `foo~/bar` (legal mid-segment tilde) → relative. `~` and `~/` exact forms expand to home
  correctly (`Join(home, p[2:])` handles the empty remainder). No defect found.
- **(b) `IsAbs`-first ordering** — verified in code (`:334` before `:340`). darwin leg proven by
  fresh tests (`C:\...` and `skills\foo\SKILL.md` both oddly-formed, not missing). Windows-host
  leg rests on structure until CI. Residual (not a defect): a POSIX `/abs` declaration on a
  Windows host classifies relative per the SPEC's own shape table — safe direction.
- **(c) missing==0-with-unresolved-shapes advisory** — cannot advise deletion: the remove phrase
  exists only inside the `missing > 0` branch (`:477-481`); cannot fold into missing: relative
  and oddly-formed `continue` before any stat (`:417-424`). The non-destructive summary branch
  (`:497-502`) carries no directive. `RelativeOnlyNonDestructiveFinding` asserts the absence
  fresh. No defect found.
- **(d) fixture retrofit** — `/nonexistent` sweep exit 1 (this tree); every fixture path derives
  from `t.TempDir()` (`absentSkillPath`, `liveSkillFile`, `absentRoot`, `writeCodexHomeConfig`,
  `envHome`). The only remaining absolute literals are fake `LookPath` return values
  (`/usr/local/bin/moai`), which are never stat'ed — the path result is discarded
  (`_, lerr := codexWiringLookPath(...)`). No leak.
- **(e) scope** — `git diff --stat d592b0551..b37fe44b4` shows 8 files: 3 code (declared set) +
  3 SPEC artifacts (this SPEC's own directory) + CHANGELOG.md (sync-phase deliverable) +
  `.moai/reports/t468/verdict.md` (card evidence path convention). No undeclared code file.
  `internal/codexwiring/skills.go` is a doc-comment-only change, within plan.md §B10's explicit
  allowance.

## Baseline-attribution

All Evidence above was produced **in this run, against this tree**:
`b37fe44b480c76ef1a03f9aadadb6cba920d6106` (branch `WT-codex-path-parser`, working tree clean —
`git status --porcelain` → 0 lines). Recorded run-phase measurements (matrix 10/10 at `7789510d2`
via `aa6c97580`; coverage 88.2% codexwiring; windows/darwin builds exit 0) are consumed only
through the attribution bridge `git diff --stat aa6c97580..b37fe44b4 -- internal/ cmd/ pkg/` →
**empty** (range touches only verdict.md, progress.md, spec.md, CHANGELOG.md) — i.e. the code
content is identical between the recorded tree and HEAD. Everything verdict-critical (both test
selectors, vet, gofmt, golangci-lint, all greps) was **re-executed fresh at HEAD** regardless.
Tool provenance: `go test` / `go vet` / `golangci-lint` compiled from this tree's module state,
so the judging build is the tree under measurement.

## Defects

- **D1 [MINOR] [blocking]** `CHANGELOG.md:330` — the clause "while relative and odd-form
  declarations — including Windows-native paths on a Windows host — are reported as
  non-destructive declared classifications and never counted missing" places the Windows-host
  case in the wrong set. Per spec.md REQ-CSP-004 (`:70`) and `classifyCodexSkillPath`
  (`doctor_codex.go:334`), a Windows-native `C:\...` absolute on a Windows host is classified
  ABSOLUTE and, when genuinely absent, IS counted missing with the remove directive — that is
  the SPEC's explicit dual-OS exception, and the published sentence asserts the opposite.
  User-facing release note contradicting the normative requirement. **Required fix**: reword to
  scope the odd-form example to a non-Windows host and state the Windows-host disposition
  (native absolute via `filepath.IsAbs`, checked normally) — a one-clause edit on the unmerged
  branch.
- **D2 [INFO] [optional]** `internal/cli/doctor_codex_test.go` — `classifyCodexSkillPath` /
  `expandCodexHomeRelativePath` are exercised only through `codexStaleSkillFinding` behavior; no
  isolated table-driven unit test. Behavior-level coverage is adequate at Tier S (arguably
  stronger). Discretionary; no action required.
- **D3 [INFO] [optional]** `internal/cli/doctor_codex_test.go` — the silent-indeterminate-only
  cell (`missing==0 && unresolved==0 && indeterminate>0` → no finding, `doctor_codex.go:454`)
  has no direct test (the symlink-loop fixtures always pair the loop with a real missing entry).
  The posture is preserved in code; a one-fixture test would pin it. Discretionary.

## Gaps

1. **Coverage % not measured this run** — the dispatch prohibits coverage instrumentation runs.
   Recorded (consumed via the empty code-diff bridge): codexwiring 88.2% at `aa6c97580`;
   internal/cli package coverage unmeasured anywhere (run-phase attempt timed out at 601s,
   deferred to CI per progress.md §E.3).
2. **Windows test-run verdict pending CI** — structurally unreachable from a darwin host
   (windows test binaries cannot execute here); `GOOS=windows` build-only exit 0 is compilation
   evidence, not classification evidence (plan.md §E D1c acknowledges exactly this). AC-04's
   Windows-native sub-assertion is code-structural until the CI windows leg runs.
3. **No CI has run on any commit of this card** — branch unpushed per lane protocol (lead
   batch-push pending); `origin/develop` CI is the integration verdict.
4. Pre-existing full-repo lint baseline (1 errcheck in `internal/template/catalog_tree_hash.go`)
   not re-measured — outside this card's scope (file untouched); scoped lint over both touched
   packages is fresh `0 issues.`

## Residual-risk

- The residual `~`-prefix heuristic labels any tilde-prefixed non-`~/` relative (e.g. `~tmp/x`)
  oddly-formed rather than relative — a label-only divergence, non-destructive in both buckets,
  and consistent with shell user-lookup semantics.
- A POSIX-style `/abs` declaration on a Windows host reads as "relative (base not observed)" —
  safe direction (never counted missing), and faithful to the SPEC's own shape table, but the
  Detail label could mislead a cross-platform config sharer.
- The Windows CI leg could still surprise (symlink-privilege skips are handled with explicit
  reasons; classification itself is host-pure stdlib, so the surprise surface is small).
- The "0 false findings on the dev machine" premise is lead-carried (spec.md §B.3), not
  re-verified — by design the fixture is the only reproduction surface.
- D1's fix (when routed) touches CHANGELOG only; no re-audit of code is needed beyond the D1
  delta.
