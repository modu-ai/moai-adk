---
id: SPEC-DOCS-WTRGO-001
title: "Fix inaccurate moai worktree go navigation examples in worktree documentation"
version: "0.1.0"
status: completed
created: 2026-07-19
updated: 2026-07-20
author: manager-spec
priority: P1
phase: "docs-site maintenance"
module: "docs-site/content/en/worktree/"
lifecycle: spec-anchored
tier: S
tags: "docs-site, worktree, cli-usage, fix"
---

# SPEC-DOCS-WTRGO-001 — Fix inaccurate `moai worktree go` navigation examples

## HISTORY

| Version | Date       | Author       | Change |
|---------|------------|--------------|--------|
| 0.1.0   | 2026-07-19 | manager-spec | Initial draft — documentation-accuracy fix for `moai worktree go` usage across four worktree docs pages |

## §A Context

The docs-site pages under `docs-site/content/en/worktree/` present `moai worktree go SPEC-ID` as a navigation command that changes the shell's working directory into a worktree. This is factually incorrect.

The actual CLI behavior, verified from `internal/cli/worktree/go.go`, is that `moai worktree go` **only prints the resolved worktree path to stdout** and never changes directory:

```go
// Output only the path for shell eval: cd $(moai wt go branch)
_, _ = fmt.Fprintln(out, wt.Path)
```

The CLI's own `Long` description states the correct idiom explicitly:

> "Use with shell command substitution to change directory: cd $(moai worktree go my-branch)"

Users following the current documentation would run `moai worktree go SPEC-ID`, see a path printed, and remain in their original directory — a broken navigation experience. The correct shell idiom is:

```bash
cd "$(moai worktree go SPEC-ID)"
```

This is a documentation-only correction. No Go source code changes are involved.

## §B Requirements (GEARS)

### REQ-WTRGO-001 — Correct bare-navigation examples (Pattern A)

Where the docs-site presents a bare `moai worktree go <SPEC-ID>` invocation implying a directory change, the documentation shall present the command wrapped in shell command substitution: `cd "$(moai worktree go <SPEC-ID>)"`.

### REQ-WTRGO-002 — Correct chained-navigation examples (Pattern B)

Where the docs-site chains a bare `moai worktree go <SPEC-ID> && <next-command>`, the documentation shall present the substitution form: `cd "$(moai worktree go <SPEC-ID>)" && <next-command>`.

### REQ-WTRGO-003 — Correct tmux-session examples (Pattern C)

Where the docs-site launches a tmux session whose command is `moai worktree go <SPEC-ID>` (which would only print the path and exit), the documentation shall set the tmux start directory via `-c "$(moai worktree go <SPEC-ID>)"`.

### REQ-WTRGO-004 — Correct diagram labels

When a mermaid diagram node or sequence label depicts `moai worktree go <SPEC-ID>` as a navigation step, the label shall be revised so it does not imply the bare command changes directory (e.g., depict `cd "$(moai worktree go ...)"` or an "enter worktree" description consistent with the substitution idiom).

### REQ-WTRGO-005 — Preserve already-correct examples

The documentation shall not modify examples that already use the correct `cd "$(moai worktree go <SPEC-ID>)"` substitution form.

### REQ-WTRGO-006 — Documentation-only change

The system shall not modify any Go source file, template, or CLI behavior as part of this correction; the change scope is restricted to markdown files under `docs-site/content/en/worktree/`.

### REQ-WTRGO-007 — Buildable docs-site

When the corrections are applied, the docs-site shall build cleanly without warnings introduced by the edits.

## §C Affected Surfaces

Four pages, with the discrepancy inventory carried in `plan.md` §F:

- `docs-site/content/en/worktree/_index.md`
- `docs-site/content/en/worktree/guide.md`
- `docs-site/content/en/worktree/examples.md`
- `docs-site/content/en/worktree/faq.md`

## §D Exclusions

This section records what is deliberately NOT built or changed under this SPEC.

### Out of Scope — Go source and CLI behavior

- No change to `internal/cli/worktree/go.go` or any other Go file — the CLI already behaves correctly; only the documentation is wrong.
- No change to the `moai worktree go` command semantics, flags, or output format.

### Out of Scope — Non-English locales

- Only `docs-site/content/en/worktree/` is corrected in this SPEC. Propagation of the fix to ko/ja/zh locales is deferred; the 4-locale sync obligation is tracked separately per the docs-site i18n rules and is not part of this SPEC's scope.

### Out of Scope — Undocumented flags

- The `moai worktree new` flag surface (including the absence of a `-b` flag) is not documented or corrected here; this SPEC is scoped strictly to `moai worktree go` navigation accuracy.

### Out of Scope — Already-correct examples

- The six verified-correct `cd "$(moai worktree go ...)"` usages (enumerated in `acceptance.md` AC-05 / Scenario 4) are explicitly preserved, not rewritten.
