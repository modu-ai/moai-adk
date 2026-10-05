# Progress — SPEC-TODO-ANALYZER-CONFORMANCE-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-TODO-ANALYZER-CONFORMANCE-001
status: draft
tier: S
artifacts: [spec.md, plan.md]   # Tier S set; this progress.md is uncounted
req_count: 5
ac_count: 6
spec_id_regex_check: PASS        # verbatim output: `PASS` — Bash ERE check, this run
head: 68e37864a
branch: WT-jev-enum-backfill
worktree: .moai/worktrees/t1311
decision_req2: "option (b) — abolition of the pick-time promise; spec_id stays record-only"
decision_rationale: spec.md §A.3 (9/9 + 53/53 measured; carriers exist; backfill machinery absent)
red_now_cells_observed:
  - AC-TAC-001 (gtd.md:258-259 two-value enum; jev only at :330)
  - AC-TAC-002 (mirror :258-259 same wording)
  - AC-TAC-003 (neutrality scan passing today — guard precondition)
  - AC-TAC-004 (grep "filled in when the item is picked" → 1 hit per surface at :252, exit=0)
  - AC-TAC-005 (no doc-coverage test for the enum/promise in internal/cli)
  - AC-TAC-006 (clean tree at plan HEAD)
open_questions: 2   # spec.md §F — priority delta; carrier-naming granularity
```

## §E.2 Run-phase Evidence

Run-phase executed 2026-09-29 in worktree `.moai/worktrees/t1311`, branch `WT-jev-enum-backfill`,
plan HEAD `3a4c558f7`, run commits `0cc2f16ac` (M1+M2) + `ed51e78a4` (M3). Baseline-attribution:
all commands below ran in this run, against this tree, on these commits.

### AC binary matrix

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-TAC-001 | PASS | `grep -n "jev" .claude/skills/moai/workflows/gtd.md \| head -5` | `263:  \`agent\` (a judgement written through \`relate\`), or \`jev\` — emitted at card` / `264:  admission only, when the \`workflow.jev.enabled\` gate is on, as a model` / `268:  it. A \`jev\` finding is a third thing: its confidence renders as \`p=\` (a` |
| AC-TAC-002 | PASS | `grep -n "jev" internal/template/templates/.claude/skills/moai/workflows/gtd.md \| head -5` | identical hits at :263/:264/:268 — both-surface parity confirmed |
| AC-TAC-003 | PASS | `go test ./internal/cli/ -run '^(TestTodoSkillDocumentsHistoryVerb\|TestTodoSkillDocumentsJevSource\|TestJevFinding_WrittenAtAdmission)$' -count=1` (post-commit) | `ok  github.com/modu-ai/moai-adk/internal/cli 1.189s` — neutrality scan green post-edit |
| AC-TAC-004 | PASS | `grep -n "filled in when the item is picked" .claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai/workflows/gtd.md; echo exit=$?` | no output lines; `promise-grep exit=1` (zero matches on both paths) |
| AC-TAC-005 | PASS | seeded-regression proof: regex-remove the jev clause from the live doc → `go test ./internal/cli/ -run '^TestTodoSkillDocumentsJevSource$' -count=1` → restore → re-run | RED: `--- FAIL: TestTodoSkillDocumentsJevSource ... live gtd.md findings-source enum does not name \`jev\`` / `FAIL  github.com/modu-ai/moai-adk/internal/cli 0.676s`; GREEN (restored): `ok  github.com/modu-ai/moai-adk/internal/cli 0.588s` |
| AC-TAC-006 | PASS (with deviation, see below) | `go test ./internal/cli/ -run '^(TestTodoSkillDocumentsHistoryVerb\|TestTodoSkillDocumentsJevSource\|TestJevFinding_WrittenAtAdmission)$' -count=1` + `go test ./internal/kanban/ -run '^TestSchemaFreezeRecordsTransitionStamps$' -count=1` + `git diff --name-only 3a4c558f7..HEAD` | `ok ... internal/cli 1.189s` / `ok ... internal/kanban 0.330s`; diff names 4 files (deviation below) |

### Deviation — AC-TAC-006 diff scope: 4 files, not 3

`git diff --name-only 3a4c558f7..HEAD` names the expected 3 files **plus**
`internal/template/catalog.yaml`. That file is the mechanical output of the spec §D-mandated
`make build` (Template-First re-embed): `make build` regenerates the content hash of the
`moai` skill tree inside `catalog.yaml`. Verified by reading the diff — the single hunk changes
only the `moai` skill-tree `hash:` line (`789dd982…` → `15b41b6b…`); no hand edit. The AC's
3-file set and the Template-First `make build` obligation cannot hold simultaneously; the
constraint (§D, [HARD]) governs, and the catalog regeneration rides commit `0cc2f16ac` with its
own body note. Not a scope expansion: no hand-written change outside the expected 3 files.

### Additional measurements

- Baseline (pre-edit, plan §C): `go test ./internal/cli/ -run '^(TestTodoSkillDocumentsHistoryVerb|TestJevFinding_WrittenAtAdmission)$' -count=1` → `ok ... 1.347s`; `go test ./internal/kanban/ -run '^TestSchemaFreezeRecordsTransitionStamps$' -count=1` → `ok ... 0.322s`.
- `go vet ./internal/cli/` → exit 0.
- `make build` → clean (agents-emit-check + templ-generate + catalog-hash + go build; build commit `3a4c558f7` stamped into binary).
- Pre-commit absorb-check: `git rev-parse --short develop` → `2b1233b13` (descendant of `68e37864a` — proceed); worktree clean at session start; HEAD re-read `3a4c558f7` immediately before the first commit.
- Seeded-regression scratch: the RED-seed edit was applied to the working copy only and restored from a `/tmp` backup before the GREEN run; the committed tree carries the intended wording (post-commit re-measure green).

## §E.3 Run-phase Audit-Ready Signal

```yaml
phase: run
spec: SPEC-TODO-ANALYZER-CONFORMANCE-001
status: implemented
tier: S
run_commits: [0cc2f16ac, ed51e78a4]
head: ed51e78a4
branch: WT-jev-enum-backfill
worktree: .moai/worktrees/t1311
ac_matrix: 6/6 PASS   # AC-TAC-006 PASS with the catalog.yaml deviation recorded in §E.2
files_changed:
  - .claude/skills/moai/workflows/gtd.md
  - internal/template/templates/.claude/skills/moai/workflows/gtd.md
  - internal/template/catalog.yaml   # mechanical make build output (§D Template-First)
  - internal/cli/todo_skill_doc_parity_test.go   # M3 new
deviations:
  - AC-TAC-006 diff scope 3→4 files; cause + verification recorded in §E.2
blockers: none
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
phase: sync
spec: SPEC-TODO-ANALYZER-CONFORMANCE-001
sync_status: ready
sync_commit_sha: "547f3d61658f8d594d9854ad5d6c02ae34983e55"
b12_self_test_a: "grep -c SPEC-TODO-ANALYZER-CONFORMANCE-001 CHANGELOG.md → 0 (exit 1) — emission allowed"
b12_self_test_b: "AC count in CHANGELOG entry = 6 (AC-TAC-001..006), matches spec.md §C — live identifiers only"
b12_self_test_c: "all file paths named in the CHANGELOG entry verified to exist this tree"
changelog_entry_position: "top of [Unreleased] § Added"
frontmatter_status_transitions:
  draft_to_in_progress: consolidated into the sync commit, disclosed in HISTORY (t1180 precedent)
  implemented_to_completed: carried by this sync commit (consolidated draft → completed)
```

### Sync evidence (5 sections)

- **Claim** — the doc conformance deliverable is synced: CHANGELOG entry emitted under B12 discipline, spec.md closed `completed`, §E.2/§E.3 evidence committed.
- **Evidence** — duplicate check: `grep -c 'SPEC-TODO-ANALYZER-CONFORMANCE-001' CHANGELOG.md` → `0`, exit 1. Run-phase §E.2 matrix 6/6 PASS (this file). Commits: `0cc2f16ac` (M1+M2, 3 files), `ed51e78a4` (M3 guard), `f9f0fcee7` (§E.2/§E.3). Absorb check at sync: `develop` = `02ad57bbe`, descendant of `68e37864a`.
- **Baseline-attribution** — all checks this run, against this worktree (`WT-jev-enum-backfill`), HEAD `f9f0fcee7` at sync start.
- **Gaps** — no sync-auditor verdict yet (runs post-merge, lead-owned); CI not read (no push from the lane).
- **Residual-risk** — `pending-backfill-sync` placeholder until the lane backfills; develop absorb at merge window may require catalog.yaml hash regeneration if other template edits landed first.
- **MX tag surface** — N/A: deliverable is markdown (gtd.md ×2) plus one `_test.go`; no production Go source touched, no exported functions added. Measured fact: `git diff --name-only 3a4c558f7..ed51e78a4` names only the two gtd.md surfaces, `internal/template/catalog.yaml` (build output), and `internal/cli/todo_skill_doc_parity_test.go`.
- **Codemap rotation** — N/A: no Go source change beyond the new test file; no module/entry-point surface changed. Measured, not assumed — same diff basis as above.
