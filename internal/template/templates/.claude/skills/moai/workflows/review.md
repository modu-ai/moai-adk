---
description: >
  Multi-perspective code review with security, performance, quality, and UX analysis.
  Supports staged changes, branch comparison, and security-focused review.
  Team mode available for parallel multi-perspective review.
  Use when performing code review, security audit, or quality assessment.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "review, code-review, security, performance, quality, ux, audit"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["review", "code review", "security audit", "quality check", "code analysis"]
  agents: ["sync-auditor"]
  phases: ["review"]
---

# Workflow: Review - Code Review

Purpose: Multi-perspective code review analyzing security, performance, quality, and UX dimensions. Produces a consolidated, prioritized report of findings.

Flow: Identify Changes -> Analyze Perspectives -> Consolidate -> Report

## Supported Flags

- --staged: Review only staged (git add) changes
- --branch BRANCH: Compare current branch against BRANCH (default: main)
- --security: Focus primarily on security review (OWASP, injection, auth)
- --file PATH: Review specific file(s) only
- --design: Extract design patterns from UI code and create/update `.moai/design/system.md`
- --critique: Post-build craft review focusing on subtle layering, surface elevation, token architecture, and typography hierarchy
- --team: Use parallel multi-perspective review team (see ${CLAUDE_SKILL_DIR}/team/review.md)
- --lean: Over-engineering-ONLY lean audit mode. Short-circuits the comprehensive 4-perspective analysis (Security / Performance / Quality / UX) and runs ONLY the over-engineering scan with the 5-tag finding format + net-reduction summary. Read-only and advisory: applies no fixes, modifies no files, renders no PASS/FAIL verdict. See the "--lean Mode" section below.
- --repo: Repo-wide scope. With --lean, sweeps the WHOLE tree instead of the diff-scope default. Ignored without --lean.

## Phase 1: Identify Changes

Determine the scope of code to review:

If --staged: Use `git diff --staged` to get staged changes.
If --branch: Use `git diff {BRANCH}...HEAD` to get branch changes.
If --file: Read the specified file(s) directly.
If no flag: Use `git diff HEAD~1` for the most recent commit changes.

Collect:
- List of modified files with change types (added, modified, deleted)
- Diff summary with line counts
- Affected modules and their responsibilities

## Phase 2: Multi-Perspective Analysis

If --lean flag: SHORT-CIRCUIT this phase entirely. Skip the comprehensive 4-perspective analysis (Perspectives 1-4 below) and jump directly to the "--lean Mode — Over-Engineering-Only Lean Audit" section. The narrowness IS the feature: correctness, security, and performance findings stay in the default (non-`--lean`) comprehensive review.

[HARD] Delegate review to the sync-auditor subagent with all perspectives (independent skeptical quality scoring per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 2).

If --team flag: Route to ${CLAUDE_SKILL_DIR}/team/review.md for parallel multi-perspective review with 4 dedicated reviewers.

If no --team flag (default single-agent mode): Delegate to the sync-auditor subagent with instructions to review from all 4 perspectives sequentially.

At the finding stage, report every issue you find, including ones you are uncertain about or consider low-severity, each with a confidence level and an estimated severity. Do not filter for importance or confidence while finding — the verdict stage (must-pass thresholds + harmonic scoring) does the filtering downstream. The goal at this stage is coverage: surfacing a finding that later gets filtered out is preferable to silently dropping a real bug.

### Perspective 1: Security Review

- OWASP Top 10 compliance check
- Input validation and sanitization
- Authentication and authorization logic
- Secrets exposure (API keys, passwords, tokens)
- Injection risks (SQL, command, XSS, CSRF)

#### Dependency Vulnerability Scan

Enumerate project manifest files and run a vulnerability scan for each detected file:
`go.mod`, `package.json`, `requirements.txt`, `Cargo.toml`, `pyproject.toml`, `Gemfile`, `composer.json`, `mix.exs`, `Package.swift`, `pubspec.yaml`.

Auto-detect language from project markers; run the dependency vulnerability scan via a per-spawn `Agent(general-purpose)` security reviewer (security whitelist + OWASP instructions per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 9) with the detected manifest.
Full OWASP checklist: load the retained `moai-ref-owasp-checklist` skill (OWASP Top 10 + dependency-scan + secrets patterns), which supplements the inline dependency and secrets scans above.

#### Secrets Scan (Full Git History)

Scan the full git history — not just the working tree — for credential leaks:

```bash
git log -p --all -G '(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})'
```

Cross-reference findings against `.gitignore` to distinguish historical leaks from working-tree exposure.
This scan is separate from working-tree-only scanners and must cover all commits reachable via `--all`.

#### Data Isolation Check

Verify the following boundaries are intact:
- **Multi-tenant**: No cross-tenant data flow; tenant ID is enforced at every query boundary.
- **PII separation**: PII is never written to logs, metrics, or telemetry endpoints.
- **Shared-state leakage**: No mutable globals that carry request-scoped data across concurrent requests.

For all three subsections above, the canonical security procedure is `Skill("moai")` `security` workflow.

If --security flag: This perspective receives primary focus with deeper analysis.

### Perspective 2: Performance Review

- Algorithmic complexity analysis (O(n) considerations)
- Database query efficiency (N+1 queries, missing indexes)
- Memory usage patterns (leaks, excessive allocation)
- Caching opportunities
- Bundle size impact (frontend changes)
- Concurrency safety (race conditions, deadlocks)

### Perspective 3: Quality Review

- TRUST 5 compliance (Tested, Readable, Unified, Secured, Trackable)
- Naming conventions and code readability
- Error handling completeness
- Test coverage for changed code
- Documentation for public APIs
- Consistency with project patterns and conventions

### Perspective 4: UX Review

- User flow integrity (do changes break existing flows?)
- Error states and edge cases from user perspective
- Accessibility compliance (WCAG, ARIA)
- Loading states and feedback mechanisms
- Breaking changes in public interfaces

## Phase 3: MX Tag Compliance Check

After perspective analysis, check @MX tag compliance for changed files:

- New exported functions: Should have @MX:NOTE or @MX:ANCHOR
- High fan_in functions (>=3 callers): Must have @MX:ANCHOR
- Dangerous patterns: Should have @MX:WARN
- Untested public functions: Should have @MX:TODO

Report missing or outdated @MX tags as findings.

## Phase 4: Report Consolidation

Produce a consolidated review report organized by severity:

### Report Structure

```markdown
## Code Review Report - {target}

### Critical Issues (must fix)
- [SECURITY] file:line: Description
- [PERFORMANCE] file:line: Description

### Warnings (should fix)
- [QUALITY] file:line: Description
- [UX] file:line: Description

### Suggestions (nice to have)
- [QUALITY] file:line: Description

### MX Tag Compliance
- Missing tags: N
- Outdated tags: N
- Compliant files: N/M

### Overall Assessment
- Security: PASS/FAIL
- Performance: PASS/WARN
- Quality: PASS/WARN
- UX: PASS/WARN
- TRUST 5 Score: N/5
```

## Phase 5: Next Steps

Present options via AskUserQuestion:

- Auto-fix issues (Recommended): Run /moai fix to automatically resolve Level 1-2 issues found in the review. Critical and complex issues will require manual attention.
- Create fix tasks: Create TaskList items for each finding so they can be addressed individually. Useful for team coordination.
- Export report: Save the review report to .moai/reports/ for future reference and tracking.
- Dismiss: Acknowledge the review without taking immediate action.

## --lean Mode — Over-Engineering-Only Lean Audit

When the --lean flag is present, the review runs ONLY this mode and nothing else. It is a narrow, single-purpose over-engineering audit — a focused "what can be cut" lens. The hard scope boundary is the mechanism's whole value: by EXCLUDING correctness bugs, security findings, and performance findings (those stay in the default comprehensive 4-perspective review), the lean audit gives a high-signal leanness scan that the broad review dilutes across four perspectives. The narrowness IS the feature.

The mode is inspired by the "lazy senior dev" minimalist-coding review pattern: a read-only, one-shot audit that applies NO fixes, hunts ONLY over-engineering, emits findings under 5 fixed tags, and closes with a net-reduction summary.

### Scope (two variants — review vs audit split)

The --lean mode supports two scopes, mirroring the diff-vs-repo split of the minimalist-coding audit pattern:

- Diff-scope (default): audit ONLY the changed code, reusing the existing Phase 1 scope selection (`--staged` / `--branch BRANCH` / `--file PATH`, or the most recent commit when no scope flag is given). This is the "review the changes" variant.
- Repo-scope (with --repo): sweep the WHOLE tree. This is the "sweep everything" variant. The `--repo` flag is honored only in --lean mode.

Both variants run the identical 5-tag scan and net-reduction summary below; only the set of files scanned differs.

### The 5 finding tags

The mode emits findings under exactly these 5 fixed tags, and no others:

| Tag | What it flags |
|-----|---------------|
| `delete:` | Unused or speculative code — dead branches, never-called helpers, write-only config |
| `stdlib:` | Reimplemented standard library — hand-rolled logic that the language's standard library already provides |
| `native:` | A dependency or code duplicating a platform-native feature the platform already provides |
| `yagni:` | Premature generality — single-implementation abstraction, single-caller indirection layer, dead config knob |
| `shrink:` | Logic reducible to fewer lines without loss of clarity |

Language neutrality: `stdlib:` names "the language's standard library" and `native:` names "a platform-native feature" generically — across all supported languages. Do NOT reference any single language's standard library module, package-manager name, or platform feature by name; the same 5 tags apply whatever language the audited code is written in.

### Output format

Emit one finding per line in this format:

```
L<line>: <tag> <what to cut>. <replacement>. [path]
```

Report every over-engineering finding you observe; the mode is finding-only and does not filter for importance. Do NOT report correctness, security, or performance issues here — those belong to the default comprehensive review (Perspectives 1, 2 above), not the lean audit.

### `@MX:DEBT` one-directional cross-link

A `yagni:` finding (single-implementation abstraction, single-caller layer, dead config) is exactly the deliberate-simplification case that an `@MX:DEBT` marker records. Before reporting a `yagni:` finding as a fresh discovery, consult the existing deferred-debt harvest:

- Read the `@MX:DEBT` harvest via `moai mx query --kind DEBT` (read-only — never write or modify markers).
- When a `yagni:` finding lands on a site that ALREADY carries an `@MX:DEBT` marker, annotate it as already-tracked deferred debt rather than re-flagging it as new:

  ```
  L<line>: yagni: <site> [already tracked @MX:DEBT — deferred]. [path]
  ```

  This avoids re-surfacing a simplification the author already recorded and deliberately deferred — the noise the `@MX:DEBT` mechanism exists to suppress.

The link is one-directional: the lean audit READS the `moai mx query --kind DEBT` harvest but NEVER creates, modifies, or removes an `@MX:DEBT` marker. Authoring `@MX:DEBT` markers remains the run-phase author's responsibility; an advisory read-only lens must not mutate source.

### Closing summary

Close the audit with exactly one of these forms:

- When at least one removal is warranted (diff-scope): `net: -<N> lines possible`
- When at least one removal is warranted (repo-scope, dependencies also removable): `net: -<N> lines, -<M> deps possible`
- When nothing warrants removal: the literal line `Lean already. Ship.`

### Read-only, advisory, no verdict

The --lean mode is read-only and advisory. It applies NO fixes, modifies NO files, and renders NO PASS/FAIL verdict — it is distinct from `/moai clean` (which removes code) and from the sync-auditor gate (which scores a verdict). It produces only a "what can be cut" list plus the net-reduction estimate. Remediation routes through the existing Phase 5 Next Steps (run `/moai fix`, create fix tasks, export the report, or dismiss) — the same as the rest of this skill.

### Doctrine cross-references (reuse, do not duplicate)

- The 5 lean tags are the OPERATIONAL scan surface for the over-engineering anti-patterns already catalogued in `.claude/skills/moai/references/anti-patterns.md` (the Premature Abstraction and Over-Engineering categories, mapped to Agent Core Behavior #4 Enforce Simplicity). Consult that catalogue for the wrong/right examples; this section does not restate it.
- The lean audit is the post-hoc DETECTION counterpart to the pre-code PREVENTION ladder in `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors #4 Enforce Simplicity (the simplicity decision ladder). The ladder prevents over-engineering before code is written; the lean audit detects what slipped through. Consult that ladder for the ordered prevention steps; this section does not restate it.

## Task Tracking

[HARD] Task management tools mandatory:
- Each critical finding tracked as a pending task via TaskCreate
- Warnings grouped by file as aggregate tasks
- Suggestions listed in report but not tracked as tasks

## Team Mode

When --team flag is provided, review delegates to the team-based multi-perspective review workflow.

Team composition: 4 review agents (security, performance, quality, UX) analyzing in parallel.

For detailed team orchestration steps, see ${CLAUDE_SKILL_DIR}/team/review.md.

Fallback: If team mode is unavailable, standard single-agent sequential review continues.

Team Prerequisites:
- workflow.team.enabled: true in .moai/config/sections/workflow.yaml
- CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 in environment
- If prerequisites not met: Falls back to single-agent review

## Phase 4.5: Design Review (Conditional)

When to run: --design or --critique flag is present, OR changed files include UI components (tsx, jsx, vue, svelte, css, scss)

### --design: Extract Design Patterns

Agent: per-spawn `Agent(general-purpose)` frontend specialist (with moai-design-craft skill; frontend whitelist per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 8)

Tasks:
1. Scan UI files for repeated patterns: spacing values, radius values, color tokens, button/card patterns, depth strategy (borders vs shadows)
2. Identify existing design conventions and inconsistencies
3. If `.moai/design/system.md` exists: Compare extracted patterns against system.md, report deviations
4. If `.moai/design/system.md` does not exist: Create system.md from extracted patterns
5. Present extraction summary with option to update system.md

Output: Design pattern report with deviation list (file:line references)

### --critique: Post-Build Craft Review

Agent: per-spawn `Agent(general-purpose)` frontend specialist (with moai-design-craft skill)

Tasks:
1. Read `.moai/design/system.md` for design direction context
2. Review built UI against craft principles:
   - **Composition**: Layout rhythm, proportions, focal point clarity
   - **Craft**: Spacing grid adherence, typography hierarchy, surface elevation consistency
   - **Content**: String coherence, data truthfulness
   - **Structure**: CSS quality (no negative margin hacks, no absolute positioning escapes)
3. Run quality checks: swap test, squint test, signature test, token test
4. Identify specific locations where defaults won over intentional design decisions
5. Provide actionable rebuild recommendations with file:line references

Output: Craft critique report with severity-ranked findings and rebuild suggestions

## Agent Chain Summary

- Phase 1: MoAI orchestrator (change identification via git)
- Phase 2-3: sync-auditor subagent (multi-perspective analysis) OR a per-spawn `Agent(general-purpose)` security reviewer (if --security)
- Phase 4-5: MoAI orchestrator (consolidation and user interaction)
- Phase 4.5 (conditional): per-spawn `Agent(general-purpose)` frontend specialist (if --design or --critique)

## Execution Summary

1. Parse arguments (extract flags: --staged, --branch, --security, --file, --design, --critique, --team)
2. If --team: Route to ${CLAUDE_SKILL_DIR}/team/review.md workflow
3. Identify code changes (git diff based on flags)
4. Delegate multi-perspective review to the sync-auditor subagent
5. Check @MX tag compliance for changed files
6. If --design or --critique: Run design review phase 4.5 (per-spawn `Agent(general-purpose)` frontend specialist with moai-design-craft)
7. Consolidate findings by severity
8. Present report with next step options

---

Version: 1.0.0
