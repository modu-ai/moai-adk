# SPEC-WEB-CODEX-PANEL-001 — progress

Card: t509 (axis B1) · Tree: `.claude/worktrees/t509` · Branch: `WT-codex-model-config`

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, this file.

- SPEC ID `SPEC-WEB-CODEX-PANEL-001` — regex check executed as Bash, output `PASS`; collision check
  against `.moai/specs/` in this worktree and in the primary checkout returned no existing
  directory.
- Mechanism fixed by operator ruling (option A, read-only mirror) — not reopened.
- Three dispatch premises re-measured against HEAD `8a6e21d98`; two corrected in spec.md §B, both
  in the direction of less work. Both corrections are reported to the lead alongside these
  artifacts.
- Status: `draft`. Awaiting plan audit and Implementation Kickoff Approval.

Plan-audit iteration 1 (`.moai/reports/t509/plan-audit.md`, FAIL 0.69, no must-pass failure) —
repaired, artifacts at version 0.2.0:

- D1 → AC-WCP-005 restated as the per-field/per-panel invariant; `boolSegment`'s radio pair
  re-read in this tree, spec.md §B.1 mechanism corrected.
- D2 → AC-WCP-011 split into ref-resolution, non-empty-diff, and filter steps. The `rc=1` trap was
  reproduced independently in this tree before rewriting.
- D3 → AC-WCP-008 scoped to `panelHTML(t, html, "codex")`; whole-body arm repurposed to prove the
  MCP surface unchanged.
- D4/D5 → `workflow.audit.model` decision closed in spec.md §C.1 as one declared exception;
  AC-WCP-006 now compares against an independent pinned list in both directions, AC-WCP-014 covers
  the exception, the "≥12" floor is gone.
- D6 → `internal/web/settings_shell.go` named in spec.md §F, rail count decided as zero in §C.2,
  REQ-WCP-012 + AC-WCP-013 + MU-6 added.
- D7 → AC-WCP-012 and AC-WCP-009 now compare extracted function bodies across revisions.
- D8/D9/D10 → mutant-split rationale corrected, `name="` + control-tag sweep and the not-last
  placement dependency pinned, icon step made actionable.

Plan-audit iteration 2 (`.moai/reports/t509/plan-audit-iter2.md`, 0.84 over the 0.80 threshold;
FAIL on the retry-contract regression clause, not on score) — two edits, both in `acceptance.md`:

- D2-1 → AC-WCP-012's extractor anchor made receiver-tolerant AND gated on a non-zero extraction
  count per side per target, with MU-8 pinning the mis-anchor mutant. `handleSave` is a method
  (`internal/web/handlers.go:350`), so the prior `^func handleSave\(` anchor matched 0 and both
  sides extracted nothing — a vacuous `IDENTICAL` on the function guarding REQ-WCP-011. The
  extraction loop was **run in this tree**: at merge-base `c068667ad`, 209/209, 67/67, 53/53.
- D2-2 → the baseline is now an explicitly computed `git merge-base origin/develop HEAD`, not a
  `git show origin/develop:<file>` read of a moving tip.

Nothing else was touched: `spec.md` and `plan.md` are unchanged at 0.2.0, and no criterion that
passed iteration 2 was edited.

Iteration-2 reinforcement (lead, same round, still `acceptance.md` only):

- All three targets re-measured in both directions (method-form and plain-form controls), each row
  summing to exactly 1: `handleSave` is a method, `parseSchemaForm` and `ApplySchemaEdits` are plain
  functions. The non-zero assertion binds all three — the anchor matching two of them today is a
  coincidence of shape, not a property.
- The vacuous pass is now an **observation**: naive anchor on `handleSave`, base 0 / head 0 /
  `diff` exit 0. The loop itself remains unrunnable in a worktree session — refused twice, first
  for a compound `git` form, then for a non-literal `awk` program — so the criterion is written as
  six plain per-side-per-target commands, all of which ran.
- Genealogy line added to AC-WCP-012: the sibling class is "the fact the verdict rests on does not
  yet exist"; this one is "the thing the predicate points at does not exist in that shape".
- New prose in this round cites function names and section numbers, never `file:line`. The
  `handlers.go:350` citation introduced in the previous round was converted; the six pre-existing
  `file:line` citations are left untouched for the lead's post-absorption re-measurement.

Plan-audit iteration 3 — **PASS 0.89** (Tier M threshold 0.80; trajectory 0.69 → 0.84 → 0.89, no
STOP). Three residual findings closed, spec.md now at 0.3.0:

- D3-5 → `version:` bumped to 0.3.0 and HISTORY rows added for iterations 2 and 3. At 0.2.0 the
  document described a state two rounds behind its own content.
- D3-3 → AC-WCP-012's command block now writes every deciding step literally, `diff` included, as
  three per-target triples plus a stated pass rule (non-zero `wc -l` pair AND silent `diff`, counts
  read first).
- D3-2 → MU-8 now names which mechanism bites per target: receiver-drop empties `handleSave` only
  (measured — `parseSchemaForm` still 67, `ApplySchemaEdits` still 53), so the two plain functions
  need the typo mechanism.

Carried into run-phase as known-unverified (auditor's own note, not a defect of these artifacts):
AC-WCP-012's green-build arm has never been executed — no `go test`, `go build`, `go vet`, or
`templ-generate` has run against this SPEC at any point in plan-phase. Run-phase is its first
execution.

## §E.2 Run-phase Evidence

Baseline re-measured in this tree at HEAD `52fe9a67c` immediately before the first edit, so every
red from that point is attributable to this work: `go build ./...` rc=0 · `go test ./internal/web/
./internal/settings/ -count=1` → `ok 4.374s` / `ok 0.954s` · `templ generate` from `internal/web`
then `git status --porcelain` → 0 lines.

Merge-base for every revision comparison: `git merge-base origin/develop HEAD` →
`0b1e27877259fef70079188bf7714029cbaa7ded`.

| AC | Status | Deciding command | Actual output |
|---|---|---|---|
| AC-WCP-001 | PASS | `go test ./internal/web/ -run TestConsoleTabsOrder -count=1` | `--- PASS: TestConsoleTabsOrder (0.00s)` — `wantTabOrder` updated to place `codex` 8th, after `audit` |
| AC-WCP-002 | PASS | `go test -run 'TestEveryTabHasAPanel\|TestTabPanelRenderOrderMatchesTabs\|TestConsoleTabsOrder'` + `grep -c 'case "codex"' internal/web/root.templ` + the not-last assertion | all three `--- PASS`; grep → `1`; not-last arm green inside `TestCodexPanel_NoNamedFormElements` |
| AC-WCP-003 | PASS | `go test -run TestCodexPanel_NoNamedFormElements -count=1` | `ok` — 0 × `name="`, 0 × `<input`/`<select`/`<textarea` in the panel region; non-vacuity arm counts 12 mirror rows |
| AC-WCP-004 | PASS | the `__present` sub-assertion in the same test | `ok` — 0 companions |
| AC-WCP-005 | PASS | `go test -run TestCodexMirrorFieldsStayOnOwningPanel -count=1` | `ok` — for all 12 mirrored names, whole-page count == owning-panel count and > 0 |
| AC-WCP-006 | PASS | `go test -run TestCodexMirrorCoverage -count=1` | `ok` — predicate == pinned 11 both directions; registry sweep (`contains "codex"`) == pinned list |
| AC-WCP-007 | PASS | `go test -run TestCodexMirrorRowLinksToOwningTab -count=1` | `ok` — seeded `sentinel-codex-model-x7` present; `mcp.tools.codex_audit.enabled` row shows `false`, `workflow.codex.task.allow_write` shows `true`; every row links to ITS owner |
| AC-WCP-008 | PASS | `go test -run TestCodexPanel_ProbeSentinel -count=1` | `ok` — binary/version/auth-provider sentinels inside the codex region, the same three still on the MCP region, not-installed branch renders |
| AC-WCP-009 | PASS | `go test -run 'TestAuditTabFields\|TestMCP'` · `go test ./internal/settings/ -run Audit` · function-body comparison | `--- PASS: TestAuditTabFields`, `TestMCPConsoleRendersAllTools`, `TestMCPConsoleToolCountMatchesCatalog`, `TestMCPConsoleWriteCapableTextDistinction`; settings `TestAuditPinFields_ExistWithTypeAndPanel`, `TestAuditPinFields_SeamRoundTrip`. `partitionWorkflowFields` 16/16 lines, `diff` silent; `isCodexToggleFieldName` 4/4 lines, `diff` silent |
| AC-WCP-010 | PASS | `grep -c '"tab.codex.title"' internal/web/assets/i18n.js` + the four governance tests | grep → `4`; `TestI18nUntranslatedValues`, `TestI18nKeyCoverageForward`, `TestI18nKeyCoverageReverse`, `TestDataI18nKeysSubsetOfDictionary` all `--- PASS`. 12 new keys × 4 locales; no allowlist entry added |
| AC-WCP-011 | PASS | the three steps, each observed separately | `ref_resolves_rc=0`; `git diff --name-only origin/develop...HEAD \| wc -l` → 20; config-surface filter → `filter_rc=1` |
| AC-WCP-012 | PASS | six extractions + three `diff`s (counts read first), then build/vet/test | `handleSave` 209/209, `parseSchemaForm` 67/67, `ApplySchemaEdits` 53/53 — all non-zero, all `diff` silent. `go build ./...` rc=0; `go vet` rc=0; `go test ./internal/web/ ./internal/settings/ -count=1` → `ok 3.845s` / `ok 0.434s`; `golangci-lint run` → `0 issues.` |
| AC-WCP-013 | PASS | `go test -run TestCodexTabRailCount -count=1` + `grep -c 'case "codex"' internal/web/settings_shell.go` | `ok` — names empty, count 0, no `panel__meta` in the panel header; grep → `1` |
| AC-WCP-014 | PASS | `go test -run TestCodexMirrorDeclaredException -count=1` | `ok` — `workflow.audit.model` row present, labelled `tab.codex.shared_backend`, linked to `/settings?tab=audit`, and absent from the predicate's output |

### Mutants (§D.2) — each applied, observed RED, reverted

| # | Observed RED | Failing assertion |
|---|---|---|
| MU-1 | yes | `codex panel region contains 1 occurrences of name="; want 0` **and** `control "workflow.audit.codex.model" renders 2 times page-wide but 1 times inside panel "audit"` |
| MU-2 | yes | `codex panel region contains 1 __present companions; want 0` (plus the AC-WCP-003 arms, as §D.2 predicts) |
| MU-3 | yes | six lines `mirror predicate misses codex field "mcp.tools.codex_*.enabled"` — by disagreement with the pinned list, naming each field |
| MU-4 | yes | `mirror row "workflow.audit.codex.effort" links to the wrong owning tab (want href="/settings?tab=audit")` |
| MU-5 | yes | `codex panel region missing probe sentinel "sentinel-provider-x7"` |
| MU-6 | yes, on the explicit-case arm only | `grep -c 'case "codex"' internal/web/settings_shell.go` → `0` (want 1) while `TestCodexTabRailCount` stayed `ok` — exactly the split §D.2 predicts |
| MU-7 | yes | `codex panel region missing the declared exception row workflow.audit.model` |
| MU-8 | yes, on all three targets, as `EXTRACTION_EMPTY` | receiver-drop on `handleSave` → 0/0 lines, `diff_rc=0`; misspelled `parseSchemaFrom` → 0/0, `diff_rc=0`; misspelled `ApplySchemaEdit` → 0/0, `diff_rc=0`. In every case the `diff` is silently green — the count arm is what catches it |

### Observations handed onward, not repaired here

- **Not a t517 sighting.** No `moai web` write-safety behaviour was observed: no `.moai/config/**`
  path appears in `git status --short` at any point of this work. The run never started a real
  server (every test builds its own `t.TempDir()` project root), so this is an absence of
  observation, not evidence that the defect is gone.
- **One existing test needed a scope repair caused by this change** (`internal/web/mcp_console_test.go`,
  `TestMCPConsoleWriteCapableTextDistinction`): it anchors its row window on
  `strings.Index(body, chip)` — the FIRST page-wide occurrence of a tool's key chip — and the codex
  mirror, which renders earlier in tab order, repeats those chips as read-only rows. Unscoped it
  reported the MCP surface degraded when it had not changed. The repair scopes the body to
  `panelHTML(t, renderConsolePage(t), "mcp")`; no production code and no save path was touched.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: dc817ff65        # the implementation commit; this line is a follow-up backfill commit
run_status: complete
ac_pass_count: 14
ac_fail_count: 0
mutants_observed_red: 8
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0   # golangci-lint run ./internal/web/... ./internal/settings/... → "0 issues."
cross_platform_build:
  darwin_arm64: pass                   # go build ./... rc=0 on this host
  other_platforms: not_measured        # CI owns the matrix; no GOOS cross-build run in this lane
total_run_phase_files: 11              # code only: 7 modified + 4 added (2 of them templ-generated); spec.md + progress.md ride the same commit
m1_to_mN_commit_strategy: single commit — the card is one milestone; M1..M6 land together
verification_scope: ./internal/web/ ./internal/settings/   # full suite is CI's, per the lane rule
templ_generate_drift: 0                # run from internal/web (from the repo root it rewrites every FileName)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
