# SPEC-DOCS-TODO-TEMP-GUARD-001 — Acceptance Criteria

## §D AC Matrix

| AC | Requirement | Criterion (binary-testable) | Severity |
|----|-------------|------------------------------|----------|
| AC-001 | REQ-001 | All four files `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-todo.md` exist. | MUST |
| AC-002 | REQ-002 | Each of the four files contains a temp-origin carve-out naming the project-local queue location `.moai/state/todo/` for temporary-directory bases. Grep per locale ≥ 1 hit of a per-locale pattern derived from the canonical wording. | MUST |
| AC-003 | REQ-003 | The old unconditional sentence — en: `Projects without git metadata use the same ~/.moai/db/<project-key>/todo/backlog.db layout.` — and its ko/ja/zh equivalents are absent (0 hits per file). Equivalent per-locale anchor: the line-248 sentence text as of authoring. | MUST |
| AC-004 | REQ-001 | The line-59 storage statement in each locale carries the same carve-out (the unconditional home-path sentence shape is absent; per-locale anchor = the line-59 sentence text as of authoring). | MUST |
| AC-005 | REQ-004 | `grep -c '^#'` returns the identical count for all four files (baseline 29). | MUST |
| AC-006 | REQ-005 | `.moai/reports/t575/docs-census.md` exists, contains the census table (hit, location, truth status), and the run verdict references its path. | MUST |
| AC-007 | REQ-006 | `git diff --stat` for the change lists only the four docs-site files (plus `.moai/reports/t575/` and the SPEC directory). | MUST |

## §D.1 Given-When-Then Scenarios

- **AC-001**: Given the worktree at the run-phase HEAD, When the verifier lists
  `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-todo.md`, Then all
  four paths resolve to existing files.
- **AC-002**: Given the four edited pages, When the verifier greps each for the
  locale's temp-origin carve-out wording (canonical en anchor:
  `.moai/state/todo/` co-occurring with a temporary-directory condition), Then
  each file yields ≥ 1 hit.
- **AC-003**: Given the four edited pages, When the verifier greps each for the
  authoring-time line-248 sentence (per-locale text fixed in the census
  artifact), Then each file yields 0 hits.
- **AC-004**: Given the four edited pages, When the verifier greps each for the
  authoring-time line-59 unconditional sentence, Then each file yields 0 hits.
- **AC-005**: Given the four edited pages, When the verifier runs
  `grep -c '^#'` on each, Then the four counts are identical.
- **AC-006**: Given the completed run, When the verifier reads
  `.moai/reports/t575/docs-census.md` and the verdict, Then the census table is
  present and the verdict cites its path.
- **AC-007**: Given the change's diff, When the verifier inspects
  `git diff --stat`, Then no Go, test, template, or README path appears.

## §D.2 Edge Cases

- A locale file drifting from the others before the edit (heading counts were
  verified equal at plan-phase — 29 each; re-baseline if a sibling card touched
  the pages).
- The absolute-`MOAI_HOME`-override branch: the corrected wording must not
  claim temp bases are always project-local — the override wins
  (`state_dir.go:28-29,37-38`).
- A git repository physically under a temp root: resolution is project-local
  while the CLI guidance reports not-refused; wording keys on the temporary
  origin of the project directory, not on the absence of git metadata alone.

## §D.3 Quality Gates

- Docs-only change: no lint/test gates fire on Go packages; the §17 verify
  recipe's 4-locale checks (file existence, section-count parity) apply.
- Language register: new sentences read as native written prose in each locale.

## §D.4 Definition of Done

- AC-001..007 all PASS with verbatim command outputs recorded in progress.md
  §E.2; census artifact landed; no out-of-scope file touched.

## §D.5 Census (to be reproduced into `.moai/reports/t575/docs-census.md`)

Plan-phase census of queue-path claims (hits + truth status at authoring time):

| Hit | Location | Claim | Truth status |
|---|---|---|---|
| 1-4 | `docs-site/content/<locale>/utility-commands/moai-todo.md` line 59 (ko/en/ja/zh) | queue stored at `~/.moai/db/<project-key>/todo/backlog.db` (unconditional) | FALSE for temporary-origin bases (in scope — AC-004) |
| 5-8 | same files, line 248 | projects without git metadata use the same home layout | FALSE for temporary-origin bases (in scope — AC-003) |
| 9-12 | `docs-site/content/<locale>/advanced/moai-web-console.md` line 108 | kanban search-order list: project-local `.moai/state/todo` first, then home `~/.moai/db/<project-key>/todo` | TRUE (describes lookup candidates in order) |
| 13-16 | `docs-site/content/<locale>/advanced/factory-mode.md` line 93 + `README*.md` line 80 | factory lane ownership at `~/.moai/db/<project-key>/factory/factory.db` | NOT VERIFIED this pass — factory store, separate from the todo queue; follow-up candidate |
| 17 | `internal/template/templates/.moai/docs/todo-queue-storage.md` lines 4, 106 | home layout explained unconditionally | Same defect class; template-shipped doc, out of card scope — follow-up finding |

Census method: `grep -rn 'moai/todo\|moai/db'` over `docs-site/content/` and
the four README files, then per-hit truth classification against the spec.md §2
truth table. README carries no todo-queue claim (only the factory line above).

## §D.6 Follow-up Findings (not silently expanded into this SPEC)

1. Template `.moai/docs/todo-queue-storage.md` (lines 4, 106) carries the same
   false-for-temp-origin claim and ships to every initialized project.
2. Factory queue path truth status under temporary origins — unverified.
3. `followup-candidates.md` tracking file does not exist in this tree; the
   tracking-owner gap this SPEC closes for the todo pages remains open as a
   mechanism for future guard-vs-doc gaps.
