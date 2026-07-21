---
id: SPEC-DOCS-WTRGO-001
title: "Fix inaccurate moai worktree go navigation examples — acceptance criteria"
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

# Acceptance Criteria — SPEC-DOCS-WTRGO-001

## §D AC Matrix

| AC ID  | Requirement | Criterion |
|--------|-------------|-----------|
| AC-01  | REQ-WTRGO-001 | All Pattern A (bare navigation) occurrences corrected to `cd "$(moai worktree go X)"` |
| AC-02  | REQ-WTRGO-002 | All Pattern B (`&&` chaining) occurrences corrected to `cd "$(moai worktree go X)" && Y` |
| AC-03  | REQ-WTRGO-003 | All Pattern C (tmux) occurrences corrected to `tmux new-session -d -s name -c "$(moai worktree go X)"` |
| AC-04  | REQ-WTRGO-004 | Mermaid/sequence diagram labels no longer imply the bare command navigates |
| AC-05  | REQ-WTRGO-005 | The six already-correct usages remain byte-unchanged |
| AC-06  | REQ-WTRGO-006 | No Go source, template, or CLI file modified |
| AC-07  | REQ-WTRGO-007 | docs-site builds cleanly with no new warnings from the edits |

## §D.1 Given-When-Then Scenarios

### Scenario 1 — Pattern A bare navigation corrected

- **Given** a reader viewing `docs-site/content/en/worktree/_index.md`
- **When** they reach a line that previously read `$ moai worktree go SPEC-AUTH-001`
- **Then** the line reads `$ cd "$(moai worktree go SPEC-AUTH-001)"`, so following it actually changes the shell into the worktree.

### Scenario 2 — Pattern B chained command corrected

- **Given** a reader viewing `docs-site/content/en/worktree/guide.md`
- **When** they reach a line that previously read `moai worktree go SPEC-001 && moai glm`
- **Then** the line reads `cd "$(moai worktree go SPEC-001)" && moai glm`, so `moai glm` runs inside the worktree.

### Scenario 3 — Pattern C tmux session corrected

- **Given** a reader viewing `docs-site/content/en/worktree/examples.md`
- **When** they reach a `tmux new-session` that previously ran `'moai worktree go SPEC-USER-001'`
- **Then** the session is created with `-c "$(moai worktree go SPEC-USER-001)"` so the pane starts in the worktree directory rather than printing the path and exiting.

### Scenario 4 — Correct usages preserved

- **Given** the six verified-correct examples (`_index.md` ~L135, `guide.md` ~L114-151, `examples.md` ~L557-560, `faq.md` ~L122-127, `faq.md` ~L525-527, and `faq.md` ~L553)
- **When** the SPEC's edits are applied
- **Then** those examples are unchanged.

### Scenario 5 — Docs-only guarantee

- **Given** the full change diff for this SPEC
- **When** filtered to `internal/`, `pkg/`, `cmd/`
- **Then** there are zero matches — no Go/CLI/template files were touched.

## §D.2 Edge Cases

- Automation-script usage (`examples.md` ~L576, `moai worktree go $SPEC_ID`) must use the substitution form with the variable preserved: `cd "$(moai worktree go "$SPEC_ID")"`.
- Prose comments that describe the old (wrong) behavior (e.g., "A new terminal opens and moves into the Worktree", `examples.md` ~L69-71) must be reconciled so the text matches the corrected command.
- A mermaid label containing `<br/>` line breaks must keep valid mermaid syntax after the label revision.

## §D.3 Quality Gate Criteria

- **Verify command (residual bare-command scan)** — expect no navigation-implying bare occurrences remain:
  ```bash
  grep -rn 'moai worktree go' docs-site/content/en/worktree/ | grep -v 'cd "\$(moai worktree go' | grep -v -- '-c "\$(moai worktree go'
  ```
  (Remaining matches must be only intentional diagram/prose references that no longer imply navigation.)
- **Docs-only check** — expect empty output:
  ```bash
  git diff --name-only | grep -E '^(internal/|pkg/|cmd/)'
  ```
- **Build check** — docs-site `hugo` build completes with no new warnings.

## §D.4 Definition of Done

- [ ] AC-01 through AC-07 all satisfied.
- [ ] All four files (`_index.md`, `guide.md`, `examples.md`, `faq.md`) edited per `plan.md` §F.
- [ ] Six preserved-correct usages verified unchanged.
- [ ] Residual bare-command scan returns only intentional non-navigation references.
- [ ] docs-only diff confirmed (no Go/CLI/template changes).
- [ ] docs-site builds cleanly.
