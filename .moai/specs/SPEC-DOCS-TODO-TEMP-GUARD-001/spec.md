---
id: SPEC-DOCS-TODO-TEMP-GUARD-001
title: "docs-site 4-locale moai-todo pages: correct the queue-location claim made false by the temporary-origin guard"
version: "1.1.0"
status: completed
created: 2026-09-13
updated: 2026-09-13
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs, docs-site, todo-queue, temp-guard, i18n"
tier: S
related_specs: [SPEC-TODO-HOME-TEMP-GUARD-001]
---

# SPEC-DOCS-TODO-TEMP-GUARD-001 — docs-site 4-locale moai-todo queue-location correction

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-09-13 | 1.0.0 | Initial draft (card t575). Authoring only — docs pages are NOT edited in plan-phase. |
| 2026-09-13 | 1.1.0 | Plan-audit D1-D4 fixes: `lifecycle` corrected to the canonical enum value `spec-anchored` (was invalid `spec-first`); census row renumbering for per-claim counts; §-reference and REQ-001 parenthetical clarifications. |

## 1. Background

SPEC-TODO-HOME-TEMP-GUARD-001 introduced a temporary-origin guard: when a launch
base classifies as a temporary-directory origin, no home queue is resolved to and
none is created — the substitute root is the launch base itself
(`internal/kanban/temp_origin.go`, `internal/kanban/state_dir.go:29-31`). The
four docs-site `utility-commands/moai-todo.md` locale pages (ko/en/ja/zh) were
not updated and still state, unconditionally, that projects without git metadata
keep their queue under `~/.moai/db/<project-key>/todo/backlog.db`. For a
temp-origin base the queue is project-local at `<base>/.moai/state/todo/`, so
the sentence is false for that project class.

The tracking-owner gap: the guard's close record (CHANGELOG line 440) noted the
doc sentence as a "doc-accuracy follow-up", but no tracking entry existed
(`.moai/docs/followup-candidates.md` does not exist in this tree). This SPEC is
that follow-up's owner.

Premise correction vs card text: the card quotes the docs as claiming
`~/.moai/todo/<project-key>/`. That was the sentence at the guard's close time;
the home-state migration (t621 rekey) has since updated the path in the pages to
`~/.moai/db/<project-key>/todo/backlog.db` while keeping the claim unconditional.
The defect class is unchanged; the literal path in the card premise is stale.

## 2. Ground Truth — queue location per project class (code-cited)

| Project class | Queue file | Code path |
|---|---|---|
| Git repository (from any linked worktree) | `~/.moai/db/<primary-checkout-key>/todo/backlog.db` | `ResolveTodoQueueRoot` → `primaryCheckoutRoot` (`internal/kanban/todo_root.go:81-86,174-179`) → home branch of `StateDirForRoot` (`internal/kanban/state_dir.go:32-40`) |
| Non-git base, not temporary-origin, home resolvable, no absolute `MOAI_HOME` override | `~/.moai/db/<key>/todo/backlog.db` | same home branch |
| Non-git base classified temporary-origin (inside `os.TempDir()`, `/tmp`, or `/var/folders`; no absolute `MOAI_HOME` override) | `<base>/.moai/state/todo/backlog.db` — project-local | temp branch of `StateDirForRoot` (`state_dir.go:29-31`); discriminant `TempOriginReason` (`temp_origin.go:60-62,75-107`); root resolution stays at the launch base (`todo_root.go:135-144`) |
| Absolute `MOAI_HOME` override set | `<override>/db/<key>/todo/backlog.db`, even for a temp base | `state_dir.go:28-29,37-38` (override beats the guard) |
| Home unresolvable, non-temp base | `<root>/.moai/state/todo/backlog.db` | `state_dir.go:32-34` |

The false doc sentences (verified present, identical line numbers across all four
locales):

- Line 59 — states the `~/.moai/db/<project-key>/todo/backlog.db` location
  unconditionally ("The queue is stored in one SQLite database at …" and its
  ko/ja/zh equivalents). False for the temp-origin class.
- Line 248 — "Projects without git metadata use the same
  `~/.moai/db/<project-key>/todo/backlog.db` layout." and its ko/ja/zh
  equivalents. False for the temp-origin class.

Neither page mentions the temporary-origin behavior at all (grep for
temporary/임시/一時/临时 across the four pages: 0 hits).

## 3. Requirements (GEARS)

- REQ-001 (Ubiquitous): The four docs-site `utility-commands/moai-todo.md` pages
  shall state the queue location for every project class the queue-root resolver
  distinguishes — git project, non-git non-temporary base, and temporary-origin
  base — per the truth table in §2 (full five-row §2 table).
- REQ-002 (Event-driven): **When** a launch base is classified as a
  temporary-directory origin without an absolute `MOAI_HOME` override, the pages
  shall describe the queue as project-local at `<base>/.moai/state/todo/`
  rather than under the home directory.
- REQ-003 (Unwanted): The pages shall not state the
  `~/.moai/db/<project-key>/todo/backlog.db` location as unconditional for
  projects without git metadata — the temp-origin carve-out must accompany the
  claim in the same file.
- REQ-004 (Ubiquitous): The four locale pages shall keep section parity: the
  same heading count (29 at authoring time) and the same section order across
  ko/en/ja/zh, per the §17 4-locale sync duty. All four files are edited in the
  same change.
- REQ-005 (Ubiquitous): The change shall record a census of queue-path claims
  across the docs-site and the README 4-locale set, with each hit's truth
  status, as an artifact under `.moai/reports/t575/`, referenced from the SPEC
  verdict.
- REQ-006 (Ubiquitous): Documentation edits shall not change any Go code, test,
  or template file; the SPEC is docs-only.

## 4. Acceptance Criteria (summary — full Given-When-Then in acceptance.md)

- AC-001: all four `docs-site/content/<locale>/utility-commands/moai-todo.md`
  files exist after the change.
- AC-002: the corrected temp-origin claim (project-local
  `.moai/state/todo/` for temporary bases) is present in all four files.
- AC-003: the old unconditional sentence (line-248 shape) is absent from all
  four files.
- AC-004: the line-59 storage statement carries the same carve-out in all four
  files.
- AC-005: heading count is identical across the four files (parity preserved;
  29 at authoring time).
- AC-006: the census artifact exists under `.moai/reports/t575/` and is
  referenced from the run verdict.
- AC-007: the diff touches only the four docs-site files (plus the census
  artifact and SPEC directory) — no Go, template, or README changes.

## 5. Out of Scope

### Out of Scope — README 4-locale set

- The README files carry no todo-queue path claim (census: only the factory
  `.db` path at line 80). No README edit.

### Out of Scope — template-shipped storage doc

- `internal/template/templates/.moai/docs/todo-queue-storage.md` (lines 4 and
  106) carries the same unconditional home-path claim and is the same defect
  class, but the card scopes docs-site only. Recorded as a follow-up finding;
  not edited here.

### Out of Scope — factory queue path truth-status verification

- `docs-site/content/<locale>/advanced/factory-mode.md` line 93 and README
  line 80 claim `~/.moai/db/<project-key>/factory/factory.db`. Whether a
  temporary-origin guard exists for the factory store was not verified in this
  pass. Recorded as an open follow-up; not fixed here.

### Out of Scope — Go code, CLI guidance, and template changes

- The temporary-origin guard itself, its CLI guidance line, and any template
  mirror are unchanged. This SPEC edits documentation only.
