# progress.md — SPEC-DIAGRAM-PROFILE-IMPORT-001

Tier M · phase "v3.1.3 target" · created 2026-08-22

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-22
artifacts: spec.md v0.2.0 (16 REQ, audit iter-1 fixes D1/D2/D3/D5 applied) · plan.md (M1–M3) · acceptance.md (16 AC) · progress.md (this skeleton)
tier: M
depends_on: SPEC-SVG-QUALITY-ABSORB-001 (status: completed — verified this session)

## §E.2 Run-phase Evidence

Baseline: HEAD `70b69732d` (WT-diagram-profile-import, tree clean), all
measurements this run against this tree. `TPL` = `internal/template/templates/.claude/skills`.

### RED-analog pre-flight (captured BEFORE any edit; plan §C)

| Command | Observed | Expected |
|---|---|---|
| `grep -rniE "\.design-dna\|profile[ -]?(marker\|snapshot)\|active profile marker" TPL/moai-domain-design-dna/ \| wc -l` | `0` | 0 ✓ |
| `grep -rn "drawio" TPL/moai-domain-svg-infographic/ \| wc -l` | `0` | 0 ✓ |
| `grep -rn "import-mermaid\|import-drawio" TPL/moai-domain-svg-infographic/ \| wc -l` | `0` | 0 ✓ |
| `grep -c "paint-order\|attachment\|mask.*gap" TPL/moai-domain-svg-infographic/scripts/check-svg.mjs` | `0` (exit 1) | t166 absent ✓ |
| `wc -l` both SKILL.md | `192` / `385` | matches plan §A |
| catalog hashes (pre) | design-dna `ec96c86b…`, svg-infographic `221fd9be…` | baseline for AC-TPL-001 |

### E1 — AC matrix (all commands run against final tree, HEAD after M2)

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-DPI-001 | PASS | `grep -c "marker-first" TPL/moai-domain-design-dna/references/diagram-profiles.md` → `3`; "in place" → `3`; `grep -c "no profile is guessed"` → `1`; slug-grammar line → `1`; `grep -c "project-scoped"` → `1`; `grep -c "not observed"` → `2` | all ≥ 1 (components i–iv) |
| AC-DPI-002 | PASS | `grep -c "diagram-profiles.md" TPL/moai-domain-design-dna/SKILL.md` → `2`; `grep -c "Phase 2 — Analyze"` → `1`; `grep -c "Phase 3 — Generate"` → `1` | pointer present; phases intact (M1 diff was +12 lines, 0 deletions) |
| AC-DPI-003 | PASS | `grep -ci "confirm before overwrite"` → `1`; `grep -ci "re-read"` → `2` | both ≥ 1 |
| AC-IMP-001 | PASS | `grep -c "untrusted"` on import-mermaid.md → `3`; `grep -c "coordinates, colors, fonts"` → `1`; `grep -c "theme, or layout"` → `1` | untrusted + full never-carry-over list |
| AC-IMP-002 | PASS | `grep -c "base64"` → `2`; `grep -c "deflate"` → `1`; `.drawio.png` → `1`; `.drawio.svg` → `1`; `grep -ci "decode before parse"` → `1`; `grep -c "never read as structure"` → `1` (import-drawio.md) | four containers + decode-first + never-read-as-structure |
| AC-IMP-003 | PASS | `grep -c "ledger\|merged\|collapsed\|dropped\|check-svg.mjs\|zero errors"` → mermaid `12`, drawio `14`; per-term greps land in both (ledger/merged/collapsed/dropped each ≥ 1; `check-svg.mjs` + `zero errors` present in both ledger tables) | ledger family complete in both files |
| AC-IMP-004 | PASS | `grep -c "one-home\|one home"` → `2` in each file; `grep -c "same change"` in the replacement sections; `grep -c "keep-both"` → `2` (both negations: "no keep-both option") | same-change replacement; no keep-both offered |
| AC-IMP-005 | PASS | per file: `grep -c "r = 8"` → `1`; `grep -c "6–10"` → `1`; `grep -c "≥ 12"` → `1`; `grep -c "paint order"` → `1`; `grep -c "six mandatory connector rules"` → `2`; `grep -c "§2.5"` → `1` | all obligations in both files |
| AC-IMP-006 | PASS | bundle-table grep (SKILL.md:294-295) shows both importer rows prefixed `Opt-in`; `git diff HEAD~1` on SKILL.md = 2 added table rows + 1 amended sentence only; six-step headings, routing table, four-dial table unmodified (no hunks touch them) | opt-in rows; default flow unchanged outside one sentence |
| AC-IMP-007 | PASS | `git diff` SKILL.md shows exactly ONE amended sentence (Step-0 no-migration sentence gains the parenthetical caller-invoked exception; tail "no diagram should ever exist in both forms…" retained verbatim); no other non-bundle hunk | amendment confined to the qualification |
| AC-ATTR-001 | PASS | per new file: `grep -c "cathrynlavery/diagram-design"` → `1`; `grep -c "MIT"` → `1` (all 3 files) | attribution in each |
| AC-TPL-001 | PASS | `make build` → exit 0, `catalog.yaml updated successfully (12899 bytes)`; post-hashes: design-dna `b35f9264…` (≠ `ec96c86b…`), svg-infographic `1b4c1a05…` (≠ `221fd9be…`) | build green; both hashes changed |
| AC-TPL-002 | PASS | 5 × `diff -q` template vs local mirror → all exit 0 (`all-5-diffs-exit-0`) | byte-identical |
| AC-TPL-003 | PASS | `grep -rn "SPEC-\|t167\|lane-"` over 3 new files → `0`; date-shaped `grep -rnE "[0-9]{4}-[0-9]{2}-[0-9]{2}"` → `0`; `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` → `ok … 20.773s` | neutrality mechanically clean |
| AC-TPL-004 | PASS | `wc -l`: design-dna SKILL.md `204`, svg SKILL.md `390` (both ≤ 500); git diffs show no frontmatter hunks (description/when_to_use byte-identical) | budget + routing surfaces untouched |
| AC-VERIFY-001 | PASS-WITH-GAP (documented fallback) | re-check on final tree: `grep -c "paint-order\|attachment\|mask.*gap" scripts/check-svg.mjs` → `0`; `grep -c "SVG07\|SVG08"` → `0` — t166 NOT landed on this branch | gap recorded; ledger's check-svg pass verified against the t165 checker surface (codes SVG001–SVG064 present; geometry checks absent). Importer constraints cite authoring.md §2.5 numbers only; no diagnostic codes invented |

### E2 — Cross-platform build

```
$ go build ./...                          → exit 0 (GO-BUILD-DARWIN-EXIT-0)
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0 (GO-BUILD-WINDOWS-EXIT-0)
```

### E3 — Coverage

n/a — no Go code in this SPEC (template markdown + catalog only; the spec's
entire verification surface is grep/diff/build/read, per acceptance.md §D.3).

### E5 — Lint + template gates

```
$ golangci-lint run --timeout=2m   → 0 issues. (exit 0)
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...
  → ok  github.com/modu-ai/moai-adk/internal/template  20.773s
```

One NEW-lint event occurred and was repaired within the run: the first strict
run FAILED with `class=S1-internal-date | match=2026-05-14` in
diagram-profiles.md (an example date in the metadata-header JSON, which the
plan had classified as user-data-shaped but the mechanical C1–C8 guard rejects
regardless). Fix: placeholder `(ISO date the profile was extracted)` +
prose alignment; strict suite then green. Final tree carries no date-shaped
literals (grep → 0).

### E4 — boundary greps

n/a (no Go package under change; no `internal/harness` / `internal/hook`
surface). Subagent boundary: no AskUserQuestion invoked (subagent contract held).

### E6 — commits (NOT pushed — factory lane)

| Milestone | Commit | Files |
|---|---|---|
| M1 | `28e7d76c5` | 5 (design-dna SKILL.md + diagram-profiles.md ×2 mirrors + spec.md status draft→in-progress) |
| M2 | `dd9487719` | 6 (svg SKILL.md + import-mermaid.md + import-drawio.md ×2 mirrors) |
| M3 | (this commit) | catalog.yaml + diagram-profiles.md neutrality repair (template+mirror) + progress.md §E |
| M3b | (follow-up) | run_commit_sha backfill (D3 placeholder pattern) |

### Plan Audit Gate re-verify (post-verdict plan.md touch)

The run executed against the post-iter-2 plan.md (the R1/R2 pointer renumber
touched after the PASS 0.92 verdict, per §F). Re-verified at run start: the
touch-up renumbered pointer references only — milestone structure (M1 design-dna
→ M2 importers → M3 closure), §A file list, §C pre-flight, §D constraints all
match the audited content; no requirement semantics changed. The skip condition
(hash unchanged) did not hold, so this §E.2 record IS the re-verify against the
touched plan, as §F required.

### Gaps / Residual risk

- AC-VERIFY-001 gap: t166 not landed here (0-match grep recorded above).
  Constraint numbers in the importer files cite authoring.md §2.5 only; when
  t166 lands on the integration base, re-read its checks against these numbers
  (acceptance.md D.6: no importer text names t166 codes, so no sync debt is
  created by the gap).
- Milestone commits M1/M2 landed before the strict-leak repair; the final tree
  (M3) is the neutrality-clean state. CI on the branch tip will see only the
  final state.
- Runtime behavior of profiles/importers is prose-contract only (§D.3): first
  real exercise happens the next time a session uses either skill.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-22
run_commit_sha: "e61a7397a"              # M3 (backfilled in M3b — self-SHA placeholder, D3 pattern)
run_status: complete-across-16-ac (15 binary PASS + AC-VERIFY-001 documented gap path per acceptance.md D.1)
ac_pass_count: 15
ac_fail_count: 0
ac_gap_recorded_count: 1        # AC-VERIFY-001, gated on lane-6 landing; fallback clause exercised
preserve_list_post_run_count: 0 # no PRESERVE surface touched: frontmatter, six-step workflow,
                                # routing table, dials, budgets, scripts/, authoring+archetypes+sketch intact
l44_pre_commit_fetch: "not-run (factory lane: no push; branch is the work's only instance)"
l44_post_push_fetch: "n/a — not pushed by design"
new_warnings_or_lints_introduced: 0   # 1 S1-internal-date finding raised AND repaired within the run
cross_platform_build:
  darwin_amd64: pass
  windows_amd64: pass
total_run_phase_files: 13        # 5 template + 5 mirrors + catalog.yaml + spec.md + progress.md (this file)
m1_to_mN_commit_strategy: per-milestone commits (M1 28e7d76c5, M2 dd9487719, M3 this, M3b sha-backfill); no push, no amend
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-22
sync_commit_sha: "pending-backfill-SPEC-DIAGRAM-PROFILE-IMPORT-001"  # backfilled in follow-up commit (D3 self-SHA pattern)
sync_status: complete
b12_self_test_a: pass   # pre-emission grep: 'SPEC-DIAGRAM-PROFILE-IMPORT-001' in CHANGELOG.md → 0 before emission
b12_self_test_b: pass   # AC count match: distinct AC IDs in acceptance.md = 16; entry references 16 (15 PASS + 1 gap annotation)
b12_self_test_c: pass   # file paths verified via ls: diagram-profiles.md, import-mermaid.md, import-drawio.md, SKILL.md ×2, spec.md, progress.md
changelog_entry_position: "[Unreleased] > Added (first bullet)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (this sync commit; status: + updated: only)"
  plan_md: "n/a — no transition owned at sync (status untouched)"
  acceptance_md: "n/a — no transition owned at sync (status untouched)"
  progress_md: "n/a — §E.4 only (this block)"
canary_compliance_check:
  required: false   # SPEC defines no forward-looking policy of its own (template-content absorption; no new gate/hook/policy to self-test)
rider: "t165 label fix folded into this sync commit — SPEC-SVG-QUALITY-ABSORB-001 spec.md frontmatter phase: \"v3.2\" → \"v3.1.3 target\" (one line, frontmatter only; no correction-note section exists in its progress.md → note skipped per lead instruction)"
scope: "lightweight sync per lead scoping: CHANGELOG + §E.4 + frontmatter close only; README/docs-site unchanged (no user-facing docs surface touched by this SPEC)"
```

## §F Phase 4 Mode Selection

Decision: **serial** (manager-develop, cycle_type=tdd as applicable to template-content work — RED here means pre-flight discriminating greps captured before edits). Justification: markdown/template-content surface (no Go code), two separable-but-sequenced groups (M1 design-dna, M2 importers, M3 verification sweep); single-author serial fits; coding is procedural-document authorship with grep/build gates. Kickoff: lead batch dispatch 2026-08-22 + plan-audit iter-2 PASS 0.92. Plan Audit Gate skip: PASS 0.92 ≥ 0.80 + artifact hash unchanged since the verdict (this §F addition and the R1/R2 touch-up of plan.md are post-verdict edits — NOTE: plan.md pointer touch-up changes the hash; the gate re-execution at run Phase 1 is therefore NOT skipped; manager-develop records the re-verify against the touched plan).
