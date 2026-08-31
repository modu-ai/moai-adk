# SPEC-GITIGNORE-ROOT-GUARD-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored in worktree `.claude/worktrees/t377`, branch `WT-gitignore-parity`,
Tier S, era V3R6. Artifact set: `spec.md` only (Tier S inlines AC in-body per §D — no `plan.md` /
`acceptance.md` were produced as separate files for this card; the SPEC's own §D carries both
requirements and acceptance criteria).

Plan-auditor converged over two iterations:

- iter1 — FAIL 0.69 (`.moai/reports/t377/plan-audit-iter1.md`, gitignored local artifact)
- iter2 — **PASS 0.8125** (Tier S threshold 0.75) (`.moai/reports/t377/plan-audit-iter2.md`)

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-31
```

## §E.2 Run-phase Evidence

**Reconciliation note (read before relying on this section).** The run phase executed and produced
verbatim evidence, but the `draft → in-progress` status transition that run-phase owns was never
performed at the time, and no `progress.md` existed to carry this section — both are being written
now, in the sync commit, rather than at the point they should have landed. The evidence itself is
not reconstructed after the fact: it is the run phase's own captured output, redirected verbatim to
`.moai/reports/t377/run-evidence.txt` (237 lines) as the delegation instructed. What is being
repaired here is the *record* of the run phase, not the run phase's work.

Full verbatim evidence: `.moai/reports/t377/run-evidence.txt`. Summary below; every command, its
exit code, and its stdout are in that file — this section cites it rather than re-pasting it.

Run-phase commits (both already landed on this branch before this sync commit):

- `9a8a99667` — `test(SPEC-GITIGNORE-ROOT-GUARD-001): assert declared gitignore rules on both surfaces (t377)` — the guard (`internal/template/embed_gitignore_generated_test.go`, function `TestGitignoreDeclaredRulesOnBothSurfaces`) plus the `**/.mink/auth/` line added to the root `.gitignore` (REQ-GRG-007).
- `006ad5943` — `docs(SPEC-GITIGNORE-ROOT-GUARD-001): pin each §B measurement to the tree it was taken on (t377)` — spec.md §B measurement refresh only; no code change.

### AC PASS/FAIL matrix (all 7, each with observed output — cited from run-evidence.txt)

| AC | Status | Verification | Observed |
|----|--------|---------------|----------|
| AC-GRG-001 | PASS | `go test ./internal/template/ -run '^TestGitignoreDeclaredRulesOnBothSurfaces$' -count=1 -v` | `--- PASS: TestGitignoreDeclaredRulesOnBothSurfaces (0.00s)`; zero-selection baseline (PRE-1) confirmed distinct — `[no tests to run]`, no `--- PASS:` line |
| AC-GRG-002 | PASS | root-surface mutant (remove `.moai/project/graph/` from root `.gitignore` only) | `--- FAIL: TestGitignoreDeclaredRulesOnBothSurfaces` with message naming surface `"root"` and rule `".moai/project/graph/"`; restored, `git diff --stat .gitignore` shows only the run-phase mink addition remaining |
| AC-GRG-003 | PASS — **RED observed for the first time in this run** | template-surface mutant + `make build` re-embed | went RED (per run-evidence.txt tail), restored + rebuilt, went PASS. spec.md §D marked this "not yet observed" at plan-audit time for THIS guard (the single-surface t373 guard had observed its own equivalent RED, but this guard had not) — the run phase is where it was first observed for `TestGitignoreDeclaredRulesOnBothSurfaces` |
| AC-GRG-004 | PASS | divergence re-measured (root 179 rules, template 136, root-only 44, template-only 1, `cmp` rc=1 — not byte-identical) THEN guard run | `--- PASS: TestGitignoreDeclaredRulesOnBothSurfaces` while divergence is non-zero and byte-inequality holds |
| AC-GRG-005 | PASS | `grep -rn '^var generatedArtifactIgnoreRules' internal/template/ --include='*.go'` | exactly one line: `internal/template/embed_gitignore_generated_test.go:28` |
| AC-GRG-006 | PASS | AC-GRG-002/003 mutant failure output | both the missing-rule string and a surface token (`"root"` / `"embedded"`) present in the failure message (see AC-GRG-002 row) |
| AC-GRG-007 | PASS (both halves) | (a) `git check-ignore -v .mink/auth/token` → `.gitignore:182:**/.mink/auth/  .mink/auth/token`, exit 0; (b) `grep -c 'mink' internal/template/embed_gitignore_generated_test.go` → `0` | both observed as required — the rule is ignored AND not declared |

### Supporting evidence (run-evidence.txt)

- Preservation: `TestEmbeddedGitignoreCoversGeneratedArtifacts` (t373's original single-surface guard) still PASS — unmodified, unbroken by this SPEC's addition.
- Package suite: `go test ./internal/template/...` — all PASS (`internal/template` 24.648s, `internal/template/agentemit` 0.659s, `scripts` no test files).
- `go vet ./internal/template/...` — exit 0.
- `golangci-lint run ./internal/template/...` — 0 issues.
- Cross-platform: `GOOS=windows go vet ./internal/template/...` — exit 0.

```yaml
run_complete_at: 2026-08-31
run_status: complete
```

## §E.3 Run-phase Audit-Ready Signal

All 7 acceptance criteria PASS with observed evidence (§E.2 above); no FAIL, no GAP. The one
open item recorded in run-evidence.txt / this section is not a defect: AC-GRG-003's RED was
observed for the first time in this run (spec.md's plan-phase text had explicitly marked it
unobserved), and that observation is now on record.

```yaml
run_audit_status: audit-ready
```

## §E.4 Sync-phase Audit-Ready Signal

### Scope of this commit

- Created this file (`progress.md`) — did not exist before this commit. Run-phase evidence section
  (§E.2) backfilled from `.moai/reports/t377/run-evidence.txt`, which the run phase itself produced
  and which was already committed at `9a8a99667`.
- `spec.md` frontmatter: `status: draft → completed` (single-commit transition — the intermediate
  `in-progress` / `implemented` states were never recorded in a separate commit at the time they
  occurred; this sync commit performs the full `draft → in-progress → implemented → completed`
  collapse in one step, per the 3-phase close convention. `updated:` refreshed to this commit's
  date.
- No `plan.md` / `acceptance.md` frontmatter to transition — Tier S, neither file exists as a
  separate artifact for this SPEC.
- CHANGELOG.md `[Unreleased]` — **no entry added.** Judgment recorded below.

### CHANGELOG judgment

Card t373 (the guard this SPEC extends) received a full narrative `[Unreleased]` entry because it
closed a **user-visible** hole: `.moai/project/graph/` was neither tracked nor ignored on either
surface, in every project deployed from the template, and a `git add -A` could commit a six-figure
generated index. That entry documents a behavior change reaching every downstream `moai init`
project.

This SPEC does not carry that shape. It:

- adds one test file (`internal/template/embed_gitignore_generated_test.go`,
  `TestGitignoreDeclaredRulesOnBothSurfaces`) — a regression guard over this repository's own two
  `.gitignore` files, not a template behavior change;
- adds one line (`**/.mink/auth/`) to the **root** `.gitignore` only — the template side already
  carried the rule (per spec.md §B, "The `**/.mink/auth/` rule" subsection), so no project deployed
  from the template observes any change at all.

No user-facing behavior changes, and no template content changes (`internal/template/templates/`
is untouched by this SPEC — confirmed by `git diff --stat 9328a5242..006ad5943 -- internal/template/templates/`,
which returns nothing). B12's own criterion for an entry — a change reaching the template or a
downstream project — is not met. **No CHANGELOG entry is added.**

### Docs-site

No docs-site work. This SPEC changes no user-facing behaviour (an internal test guard plus one
root `.gitignore` line, confirmed above); the 4-locale obligation in `.claude/rules/moai/development/docs-site-i18n-rules.md`-class doctrine binds user-facing documentation changes only, which this SPEC has none of.

### Baseline-attribution

Measured in worktree `.claude/worktrees/t377`, branch `WT-gitignore-parity`, at HEAD `006ad5943`
(the tree as of the two prior run-phase commits; this sync commit's own SHA is not yet known at
authoring time — see `sync_commit_sha` placeholder below).

### Gaps — what was explicitly NOT observed in this sync commit

- **CI has not run.** This branch is unpushed (per dispatch instruction — "Do NOT push"); no
  clean-environment or darwin/windows-matrix verdict exists for any commit on this branch,
  including this one.
- **The `draft → in-progress` and `in-progress → implemented` transitions were never independently
  observed or recorded** — they are being collapsed into this single sync commit's frontmatter
  change rather than reconstructed as separate historical events, because no record of them exists
  to reconstruct from.
- **The docs-site grep-based non-applicability judgment above was not independently re-verified by
  a second reader** — it rests on this session's own `git diff --stat` check.

### Residual-risk

- **The reconciliation record above depends on this session's own account of what happened.** No
  one else observed the missing `draft → in-progress` transition at the time it should have
  occurred; if the run phase's actual sequence of events differed from what run-evidence.txt and
  the two commit messages suggest, that difference is not recoverable from what exists on disk.
- **CHANGELOG omission is a judgment call**, not a mechanically-derived result — a future reader
  who disagrees that a regression-guard-only change warrants no entry can revisit it; the reasoning
  and the measurements it rests on are recorded above for that purpose.

```yaml
sync_complete_at: 2026-08-31
sync_commit_sha: pending-backfill-sync
sync_status: complete
b12_self_test_a: 0 (grep -c 'SPEC-GITIGNORE-ROOT-GUARD-001' CHANGELOG.md — no prior entry, no duplicate)
b12_self_test_b: 7 (AC count in spec.md matches AC PASS/FAIL matrix row count)
b12_self_test_c: not-applicable (no CHANGELOG entry emitted this commit)
changelog_entry_position: none (judged not warranted — see CHANGELOG judgment above)
frontmatter_status_transitions.spec_md: draft -> completed (single-commit collapse; in-progress/implemented never independently recorded)
frontmatter_status_transitions.plan_md: not-applicable (no separate plan.md for this Tier-S SPEC)
frontmatter_status_transitions.acceptance_md: not-applicable (no separate acceptance.md for this Tier-S SPEC)
```
