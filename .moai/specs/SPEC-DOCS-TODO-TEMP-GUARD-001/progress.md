# SPEC-DOCS-TODO-TEMP-GUARD-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-DOCS-TODO-TEMP-GUARD-001
tier: S
status: draft
card: t575
worktree: .claude/worktrees/t575
branch: WT-docs-todo-locales
authoring_only: true
artifacts_emitted:
  - .moai/specs/SPEC-DOCS-TODO-TEMP-GUARD-001/spec.md
  - .moai/specs/SPEC-DOCS-TODO-TEMP-GUARD-001/plan.md
  - .moai/specs/SPEC-DOCS-TODO-TEMP-GUARD-001/acceptance.md
  - .moai/specs/SPEC-DOCS-TODO-TEMP-GUARD-001/progress.md
grounding_evidence:
  truth_table: spec.md §2 (code-cited: internal/kanban/todo_root.go, state_dir.go, temp_origin.go)
  false_sentences: docs-site/content/<locale>/utility-commands/moai-todo.md lines 59 and 248, all 4 locales (verified present)
  heading_parity_baseline: 29 per file (verified equal across ko/en/ja/zh)
  tracking_owner_gap: confirmed — .moai/docs/followup-candidates.md absent; only CHANGELOG line 440 mentions the doc follow-up
  spec_id_check: "ID=SPEC-DOCS-TODO-TEMP-GUARD-001 regex ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ → PASS (verbatim Bash output); uniqueness → no existing directory"
premise_correction: card quotes the stale path ~/.moai/todo/<key>/; current pages say ~/.moai/db/<key>/todo/backlog.db (post-t621 rekey) — defect class unchanged
doc_pages_edited: false   # authoring only per delegation prompt
plan_audit_fixes: v1.1.0 — D1 lifecycle enum corrected to spec-anchored (blocking, SSOT-confirmed); D2 census renumber 13-20 with per-row multiplier; D3 plan.md AC-section pointer §3→§4; D4 REQ-001 five-row-table parenthetical
```

## §E.2 Run-phase Evidence

Measured on this tree, branch WT-docs-todo-locales, pre-edit HEAD e83b4ec20 (docs files identical to HEAD before edit — `git status` showed only untracked `.moai/reports/t575/`).

**(a) AC-003 evidence — pre-edit line-248 content captured before edits (per delegation, en shown; ko/ja/zh identical line numbers):**

```
$ git show HEAD:docs-site/content/en/utility-commands/moai-todo.md | sed -n '248p'
Run it inside a linked worktree and the queue still **resolves to the primary checkout's project key** — the contract is one repository, one queue. A `moai todo add` from a card worktree lands in the same database the lead and the foreman loop read. Projects without git metadata use the same `~/.moai/db/<project-key>/todo/backlog.db` layout.
```

Post-edit line 248 (en): same sentence with the final sentence replaced by "Projects without git metadata use the same home layout — unless they originate in a temporary directory, in which case the queue is project-local at `<base>/.moai/state/todo/backlog.db` (an absolute `MOAI_HOME` override applies even then)."

**(b) AC-002 — temp-origin carve-out present, per locale (`grep -c '.moai/state/todo/' <file>`):**

```
ko: 2
en: 2
ja: 2
zh: 2
```

**(c) AC-003 — old line-248 sentences absent, per locale (0 each):**

```
$ grep -c 'backlog.db` 구조를 사용합니다\|backlog.db` layout\|backlog.db` の構成を使います\|backlog.db` 结构' <4 files>
zh: 0 / ko: 0 / en: 0 / ja: 0
```

**(d) AC-004 — old unconditional line-59 shapes absent (line-start anchors, 0 each; first probe with substring anchors hit the new conditioned sentences in ko/ja/zh — re-anchored):**

```
$ grep -c '^대기열은 `~/.moai/db' <ko>   → 0
$ grep -c '^The queue is stored in one SQLite database at `~/.moai/db' <en> → 0
$ grep -c '^キューは `~/.moai/db/<project-key>/todo/backlog.db` という' <ja> → 0
$ grep -c '^队列保存在 `~/.moai/db/<project-key>/todo/backlog.db` 这一个' <zh> → 0
```

**(e) AC-005 — heading parity post-edit (`grep -c '^#'`):**

```
ko: 29 / en: 29 / ja: 29 / zh: 29
```

**(f) AC-007 — diff scope (`git diff --stat`):**

```
 docs-site/content/en/utility-commands/moai-todo.md | 4 ++--
 docs-site/content/ja/utility-commands/moai-todo.md | 4 ++--
 docs-site/content/ko/utility-commands/moai-todo.md | 4 ++--
 docs-site/content/zh/utility-commands/moai-todo.md | 4 ++--
 4 files changed, 8 insertions(+), 8 deletions(-)
```

(Plus the census artifact `.moai/reports/t575/docs-census.md` and this SPEC directory — allowed by AC-007.)

**(g) AC-006** — census artifact landed at `.moai/reports/t575/docs-census.md` with the 21-claim per-claim table (8 false in-scope FIXED, 4 TRUE, 8 factory/README not-verified, 1 template follow-up).

## §E.3 Run-phase Audit-Ready Signal

```yaml
phase: run
spec: SPEC-DOCS-TODO-TEMP-GUARD-001
tier: S
card: t575
ac_matrix:
  AC-001: PASS   # four files exist (ls pre-edit batch output)
  AC-002: PASS   # carve-out ≥1 hit ×4 — observed 2 per locale
  AC-003: PASS   # old line-248 sentence 0 hits ×4
  AC-004: PASS   # old line-59 unconditional shape 0 hits ×4 (line-start anchors)
  AC-005: PASS   # heading parity 29/29/29/29
  AC-006: PASS   # census at .moai/reports/t575/docs-census.md, cited from verdict
  AC-007: PASS   # diff = 4 docs files + census + SPEC dir only
run_status: audit-ready
pre_edit_head: e83b4ec20
gaps:
  - hugo build not run (CI/Vercel owns it — recorded per delegation)
  - moai-web-console.md:108, factory-mode.md:93, README line 80, template todo-queue-storage.md: not edited (out of scope; t704/t706 own them)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
phase: sync
spec: SPEC-DOCS-TODO-TEMP-GUARD-001
tier: S
card: t575
sync_status: audit-ready
sync_commit_sha: pending-backfill-sync   # placeholder — commit cannot cite its own hash; backfilled in the following commit
changelog_entry: CHANGELOG.md [Unreleased] → Fixed (single entry, first position)
changelog_b12_self_test:
  pre_emission_grep: 0   # grep -c 'SPEC-DOCS-TODO-TEMP-GUARD-001' CHANGELOG.md → 0 before emission
  ac_count_match: 7 == 7 # acceptance.md distinct AC ids = 7; entry references 7
  file_path_verify: PASS # docs-site/content/{ko,en,ja,zh}/utility-commands/moai-todo.md, .moai/reports/t575/docs-census.md all resolve
mx_tag_validation: not-applicable   # docs-only SPEC — no exported Go symbols introduced or modified
canary_compliance_check:
  close_subject_single_full_id: PASS   # chore/commit subject names SPEC-DOCS-TODO-TEMP-GUARD-001 exactly once
  heading_parity_post_close: 29/29/29/29
frontmatter_status_transitions:
  draft_to_in_progress: "55adec04f (run-phase first commit)"
  in_progress_to_completed: "this sync commit (pending-backfill-sync)"
out_of_scope_followups: [t704, t705, t706]
```
