# PR Code Review Multi-Agent Architecture

The PR Code Review process uses a multi-agent architecture following the official Claude Code plugin pattern.

## Agent Roles

### Eligibility Check Agent (Haiku)

The Haiku agent performs lightweight filtering to avoid unnecessary reviews. It checks PR state and metadata to determine if review is warranted.

Skip conditions:

- Closed PRs
- Draft PRs
- PRs already reviewed by bot
- Trivial changes (typo fixes, comment-only)
- Automated dependency updates

### Context Gathering

Before launching review agents:

- Find CLAUDE.md files in directories containing modified code to understand project-specific coding standards
- Generate a concise summary of PR changes (files modified, lines added/removed, overall impact)

### Parallel Review Agents (5 Sonnet instances)

Five Sonnet agents run in parallel, each focusing on a specific review dimension:

| Agent | Focus |
|-------|-------|
| Agent 1 | CLAUDE.md compliance — violations of documented coding standards and conventions |
| Agent 2 | Obvious bugs — logic errors, null reference risks, resource leaks |
| Agent 3 | Git blame and history context — recent changes, evolving patterns |
| Agent 4 | Previous PR comments — recurring issues, unresolved feedback |
| Agent 5 | Code comment compliance — comments accurate and helpful |

## Confidence Scoring System

Each detected issue receives a confidence score from 0 to 100:

| Score | Meaning |
|-------|---------|
| 0 | False positive — no confidence |
| 25 | Somewhat confident — might be real |
| 50 | Moderately confident — real but minor |
| 75 | Highly confident — very likely real |
| 100 | Absolutely certain — definitely real |

## Filter and Report Stage

Issues below the 80 confidence threshold are filtered out to reduce noise. Remaining issues are formatted and posted to the PR via the GitHub CLI.

Output format:

- Code review header
- Count of found issues
- Numbered list of issues with descriptions
- Direct links to code with specific commit SHA and line range

## Example PR Review Output

```markdown
## Code review

Found 3 issues:

1. CLAUDE.md compliance — `internal/auth/login.go:45-52` violates the
   error-wrapping convention (Section 7). Use `fmt.Errorf("auth: %w", err)`.
   [link](https://github.com/org/repo/pull/123/files#diff-abc#L45-L52)

2. Obvious bug — `internal/api/handler.go:88` dereferences `user` before
   the nil check on line 90. NPE risk on cold path.
   [link](https://github.com/org/repo/pull/123/files#diff-def#L88)

3. Code comment compliance — `pkg/parser/parse.go:120` comment claims
   "returns error on empty input" but function returns nil — comment is
   stale since refactor commit `a3b2c1d`.
   [link](https://github.com/org/repo/pull/123/files#diff-ghi#L120)
```
