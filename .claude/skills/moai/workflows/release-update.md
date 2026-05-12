---
name: moai-workflow-release-update
description: >
  Claude Code upstream change tracker for moai-adk-go maintainers (dev-only; user-invocable: false).
  NOT distributed to user projects. Tracks CC release notes since last analyzed version,
  classifies changes by impact tier (Hook/Agent/MCP/TUI/Platform), cross-references official docs,
  generates update plan or umbrella SPEC, synchronizes docs-site 4-locale + README via manager-docs,
  and commits via manager-git. Activated by /release-update command.
user-invocable: false
allowed-tools: Read, Write, Edit, Bash, WebFetch, WebSearch, Glob, Grep, Agent, ToolSearch, TaskCreate, TaskUpdate, TaskList, TaskGet
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-05-12"
  tags: "release-update, cc-update, upstream, changelog, docs-sync, dev-only"
  aliases: "cc-update, release-track"
  dev_only: "true"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 120
  level2_tokens: 8000

# MoAI Extension: Triggers
triggers:
  keywords: ["release-update", "cc-update", "release-track", "upstream changes", "claude code update", "cc changelog"]
  agents: []
  phases: []
---

# Workflow: release-update — CC Upstream Change Tracker

> **[DEV-ONLY]** This workflow is exclusively for moai-adk-go maintainers.
> It MUST NOT be added to `internal/template/templates/` or any user-facing artifact.
> Entry point: `.claude/commands/97-release-update.md`

## 1. Purpose & Scope

Automates the cycle of:
1. Detecting new Claude Code releases since the last tracked version
2. Classifying upstream changes by impact on moai-adk-go
3. Generating an actionable update plan (or umbrella SPEC directory for large diffs)
4. Synchronizing docs-site (4-locale) + README files
5. Opening a PR via manager-git

**In scope**: CC release notes analysis, moai-adk-go documentation update, SPEC stub generation, state file maintenance.

**Out of scope**: Implementing code changes (delegate to /moai run SPEC-XXX), modifying `internal/template/templates/` (template changes require their own SPEC).

## 2. Activation

### Flags

| Flag | Behavior |
|------|----------|
| `--since vX.Y.Z` | Override start version (ignores state file) |
| `--dry` | Analyze and report only; skip Phases 6-7 (no file edits, no commits) |
| `--report-only` | Alias for `--dry` |
| `--docs-only` | Skip Phase 4 plan generation; jump to Phase 6 using existing plan |
| `--master-spec` | Force umbrella SPEC directory even if diff < 10 items |

### Invocation examples

```
/release-update
/release-update --since v2.1.100
/release-update --dry
/release-update --docs-only
/release-update --master-spec
```

## 3. Phase-by-Phase Specification

### Phase 0 — Load State

**Goal**: Determine the `since_version` baseline.

Steps:
1. Read `.moai/state/last-cc-version.json`
   - If file missing: default `since_version = "2.1.0"`, emit warning
   - If `--since` flag provided: override with flag value (ignore state)
   - Otherwise: use `last_analyzed_version` from state file
2. Log resolved `since_version` for audit trail
3. Create TaskList entries for Phases 1-8 if TaskCreate is available

State file schema:
```json
{
  "last_analyzed_version": "2.1.139",
  "last_analyzed_date": "2026-05-12",
  "last_master_research": ".moai/research/cc-update-20260512.md",
  "analysis_history": [
    {
      "version_range": "2.1.0..2.1.139",
      "date": "2026-05-12",
      "spec_id": null,
      "items_found": 47
    }
  ]
}
```

### Phase 1 — Collect Release Notes

**Goal**: Obtain the raw CC changelog text.

Strategy (in priority order):

**Option A — `/release-notes` session command** (preferred):
- `/release-notes` is an interactive CC session command that shows the current release notes
- The orchestrator should instruct the user to run `/release-notes` and paste the output
- Use AskUserQuestion (after ToolSearch preload) to request the paste

**Option B — Cache file** (if available):
- Check `~/.claude/RELEASE_NOTES.md` or `~/.claude/release-notes.txt`
- Read directly if present and recent (mtime within 7 days)

**Option C — WebSearch fallback**:
- Search: `"Claude Code release notes" site:code.claude.com OR site:anthropic.com 2026`
- Fetch changelog page via WebFetch

Steps:
1. Check for cache file first (Bash: `ls -la ~/.claude/RELEASE_NOTES.md 2>/dev/null`)
2. If cache hit → Read cache, proceed to Phase 2
3. If no cache → ToolSearch preload → AskUserQuestion requesting paste from `/release-notes`
4. Write received text to `/tmp/cc-release-notes-$(date +%Y%m%d).txt` for parsing

[HARD] AskUserQuestion preload: Always call `ToolSearch(query: "select:AskUserQuestion")` before AskUserQuestion invocations (CLAUDE.md §19.2).

### Phase 2 — Diff & Categorize

**Goal**: Filter and classify entries newer than `since_version`.

Version comparison:
- Parse semver segments: `major.minor.patch`
- Include entries where `version > since_version` (strict greater-than)
- If no entries found: emit "No new versions since vX.Y.Z" and stop (exit Phase 2 early)

Classification tiers and keywords:

| Tier | Impact | Keywords / Signals |
|------|--------|-------------------|
| Tier 1 — Critical | Hooks, agents, skills, plugins, sub-agents, MCP protocol, permissions model, settings.json schema changes | hook, agent, skill, plugin, subagent, mcp, permissions, settings, frontmatter |
| Tier 2 — Important | TUI, CLI, statusline, worktree, headless, session management, memory | tui, statusline, worktree, headless, session, memory, /clear, slash command |
| Tier 3 — Minor | Voice, remote, platform-specific, UI-only, analytics | voice, remote, windows, linux, mac, ui, analytics, telemetry |

Output: structured Markdown table:
```markdown
| Version | Category | Tier | Summary | Impact on moai-adk-go |
|---------|----------|------|---------|----------------------|
| v2.1.140 | Hook system | 1 | New PostToolUse event fields | Update hooks.md, CHANGELOG |
```

Also emit: `total_items`, `tier1_count`, `tier2_count`, `tier3_count`.

### Phase 3 — Cross-Reference Official Docs

**Goal**: Enrich each Tier 1/2 candidate with stable doc references and API signatures.

[HARD] Execute ALL WebFetch calls in parallel (CLAUDE.md §1 Parallel Execution rule):

```
Parallel WebFetch (all independent, single message):
  - https://code.claude.com/docs/en/hooks
  - https://code.claude.com/docs/en/sub-agents
  - https://code.claude.com/docs/en/skills
  - https://code.claude.com/docs/en/plugins
  - https://code.claude.com/docs/en/mcp
  - https://code.claude.com/docs/en/settings
```

Fallback if WebFetch fails: note "doc unavailable at fetch time" — do not block Phase 3.

For each Tier 1/2 item: annotate with `doc_url` and `stable_signature` (key field names, event names, config keys).

### Phase 4 — Generate Update Plan

**Goal**: Produce a structured research document or umbrella SPEC directory.

Size decision:
- `tier1_count + tier2_count < 10` AND `--master-spec` flag absent → **Small plan**: save to `.moai/research/cc-update-YYYYMMDD.md`
- `tier1_count + tier2_count >= 10` OR `--master-spec` flag present → **Umbrella SPEC**: create `.moai/specs/SPEC-V3R4-CC2X-ADOPT-NNN/research.md`
  - NNN: next sequential number (Bash: `ls .moai/specs/ | grep CC2X | tail -1` → increment)

Plan document structure (`.moai/research/cc-update-YYYYMMDD.md` or `.moai/specs/.../research.md`):

```markdown
# CC Upstream Update Research — vX.Y.Z..vA.B.C
Date: YYYY-MM-DD
Analyzed by: moai-workflow-release-update v1.0.0

## Executive Summary
- Version range: vX.Y.Z → vA.B.C (N releases)
- Tier 1 (Critical): N items
- Tier 2 (Important): N items
- Tier 3 (Minor): N items

## Tier 1 — Critical Impact

| Item | Version | Affected File(s) | Doc Ref | Action Required |
|------|---------|-----------------|---------|----------------|

## Tier 2 — Important Impact

| Item | Version | Affected File(s) | Doc Ref | Action Required |
|------|---------|-----------------|---------|----------------|

## Tier 3 — Minor Impact
(summary only, no per-item table)

## Cross-Cutting Concerns
List items that affect multiple files or require coordinated changes.

## Recommended Child SPEC Decomposition
(Only for umbrella SPEC path)
- SPEC-V3R4-CC2X-ADOPT-NNN-T1: [Tier 1 batch description]
- SPEC-V3R4-CC2X-ADOPT-NNN-T2: [Tier 2 batch description]

## References
- CC Release Notes: /tmp/cc-release-notes-YYYYMMDD.txt
- Official docs fetched: YYYY-MM-DD
- State baseline: vX.Y.Z (from .moai/state/last-cc-version.json)

## Open Questions
```

Skip Phase 4 if `--docs-only` flag is set.

### Phase 5 — User Report & Approval

**Goal**: Present findings and obtain explicit approval before modifying docs.

[HARD] Preload: `ToolSearch(query: "select:AskUserQuestion")` before AskUserQuestion call.

Present summary:
- Version range analyzed
- Tier counts (Tier 1: N, Tier 2: N, Tier 3: N)
- Plan file path
- Flag `--dry` or `--report-only` status

AskUserQuestion structure (4 options max, first = recommended per CLAUDE.md §8):

```
Question: "CC 업스트림 분석 완료. 다음 단계를 선택하세요."
Options:
  A. "전체 동기화 진행 (권장)" — Phase 6(docs) + Phase 7(commit+PR) 실행. docs-site 4개 locale + README 업데이트 후 PR 오픈.
  B. "플랜만 생성, 문서 업데이트 없음" — Phase 4 플랜 파일 저장 후 종료. 문서/커밋 없음.
  C. "SPEC 스텁 추가 생성" — 플랜 + Phase 6 문서 업데이트 + 빈 child SPEC 디렉토리 생성.
  D. "중단" — 상태 파일 유지, 종료.
```

If `--dry` or `--report-only` flag was set: skip AskUserQuestion, auto-select Option B.

### Phase 6 — Documentation Updates

**Precondition**: User selected Option A or C in Phase 5.

**[HARD] §17.3 rule**: same-PR 4-locale sync required for non-bulk content changes. All 4 locales MUST be updated in a single delegation.

**[HARD] §17.1 URL blacklist**: `docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr` are forbidden. Use only `adk.mo.ai.kr`.

Delegation to manager-docs subagent (foreground, `run_in_background: false`):

```
Delegate to manager-docs subagent with explicit scope:

Files to update (per plan Tier 1/2 items):
  - docs-site/content/ko/...    (canonical Korean source, §17.3)
  - docs-site/content/en/...    (English translation)
  - docs-site/content/ja/...    (Japanese translation, parallel with zh)
  - docs-site/content/zh/...    (Chinese translation, parallel with ja)
  - README.md                   (project root, English)
  - README.ko.md                (project root, Korean)

Constraints:
  - URL blacklist: docs.moai-ai.dev / adk.moai.com / adk.moai.kr are FORBIDDEN
  - ko is canonical source; en/zh/ja are translations
  - Mermaid diagrams: TD direction only (§17.2)
  - Run scripts/docs-i18n-check.sh after edits to verify 4-locale consistency
  - Bulk changes (>5,000 words): ok to use translation_status: pending + 48h grace
```

If Option C was selected: also create child SPEC stub directories under `.moai/specs/` with empty `spec.md` stubs (headers only, no content). These are created by the orchestrator directly, not delegated.

### Phase 7 — Persist State & PR

**Goal**: Update state file, create branch, open PR.

Step 7a — Update state file:
```json
{
  "last_analyzed_version": "<latest analyzed version>",
  "last_analyzed_date": "<today YYYY-MM-DD>",
  "last_master_research": "<path to plan file>",
  "analysis_history": [
    ... (prepend new entry)
  ]
}
```
Write via Edit tool (read-first pattern per agent-common-protocol.md).

Step 7b — Delegate to manager-git subagent (`run_in_background: false`):

```
Branch naming:
  - chore/cc-update-YYYYMMDD   (for small plans, squash merge)
  - feat/cc-update-YYYYMMDD    (for umbrella SPECs, squash merge)

Commit message (Conventional Commits, §18 pattern):
  chore(release-update): track CC vX.Y.Z..vA.B.C upstream changes

PR body template:
  ## Summary
  - Version range: vX.Y.Z → vA.B.C (N releases)
  - Tier 1 Critical: N items | Tier 2 Important: N items | Tier 3 Minor: N items
  - Plan: <plan file path>

  ## Changes
  - Updated docs-site 4-locale for Tier 1/2 items
  - Updated README.md + README.ko.md
  - Updated .moai/state/last-cc-version.json

  ## Merge Strategy
  - [x] squash (chore PR per §18.3)

  ## Verification
  - [ ] scripts/docs-i18n-check.sh PASS
  - [ ] 4 locales updated in same PR
  - [ ] URL blacklist clear

  🗿 MoAI <email@mo.ai.kr>

Merge strategy: squash (chore-typed per CLAUDE.local.md §18.3)
Labels: type:chore, area:docs-site, priority:P2
```

Mark TaskList items complete after PR creation.

### Phase 8 — Completion Marker

Emit: `<moai>DONE</moai>`

Print to user (Korean if conversation_language=ko):
```
release-update 완료.
- 분석 범위: vX.Y.Z → vA.B.C
- 플랜 파일: <path>
- PR: <url or "skipped (--dry)">
- 다음 단계: Tier 1 항목 구현은 /moai plan SPEC-V3R4-CC2X-ADOPT-NNN 으로 진입하세요.
```

If multi-session is needed (plan file > 20 items), emit a paste-ready resume message per `.claude/rules/moai/workflow/session-handoff.md` canonical format.

## 4. Agent Delegation Map

| Phase | Delegated to | Mode | Notes |
|-------|-------------|------|-------|
| 1 | Orchestrator (direct) | — | State file read, AskUserQuestion for paste |
| 2 | Orchestrator (direct) | — | Parsing + classification |
| 3 | Orchestrator (parallel WebFetch) | — | 6 docs fetched in parallel |
| 4 | Orchestrator (direct Write) | — | Plan file creation |
| 5 | Orchestrator (AskUserQuestion) | — | User approval gate |
| 6 | manager-docs subagent | foreground | 4-locale + README update |
| 6 | Optional: plan-auditor | foreground | Spot-check generated plan quality |
| 7 | manager-git subagent | foreground | Branch + PR |
| 8 | Orchestrator (direct) | — | Completion output |

## 5. Output Artifacts

| Artifact | Path | Phase |
|----------|------|-------|
| Raw release notes (temp) | `/tmp/cc-release-notes-YYYYMMDD.txt` | 1 |
| Update plan (small) | `.moai/research/cc-update-YYYYMMDD.md` | 4 |
| Update plan (large) | `.moai/specs/SPEC-V3R4-CC2X-ADOPT-NNN/research.md` | 4 |
| State file (updated) | `.moai/state/last-cc-version.json` | 7 |
| docs-site updates | `docs-site/content/{ko,en,ja,zh}/...` | 6 |
| README updates | `README.md`, `README.ko.md` | 6 |
| PR | GitHub (via gh pr create) | 7 |

## 6. Verification (CLAUDE.md §16 Self-Check)

Before starting Phase 6, the orchestrator MUST verify:

1. Is this a workflow agent task? → Yes (SPEC analysis + docs delegation). **Delegate to manager-docs for doc writes.**
2. Does manager-docs exist in the catalog? → Yes. **Use it.**
3. Does delegation improve quality/independence over direct file edits? → Yes (4-locale consistency, i18n-check script). **Delegate.**

Before completing Phase 7:
- [ ] `.moai/state/last-cc-version.json` updated (Read-verify after Write)
- [ ] `scripts/docs-i18n-check.sh` result shown in PR description
- [ ] All 4 locales included in the same commit/PR (§17.3)
- [ ] URL blacklist: no forbidden URLs in any updated file (Grep check)
- [ ] Merge strategy: squash (§18.3 chore PR rule)

## 7. Anti-Patterns

| Anti-Pattern | Why Forbidden | Correct Approach |
|--------------|--------------|-----------------|
| Skipping AskUserQuestion in Phase 5 | CLAUDE.md §1 HARD rule: all user questions via AskUserQuestion | Always preload via ToolSearch then AskUserQuestion |
| Updating only `docs-site/content/ko/` | §17.3 requires same-PR 4-locale sync | Delegate to manager-docs with all 4 locales in scope |
| Writing directly to `internal/template/templates/` | DEV-ONLY workflow; template changes corrupt user projects | This workflow NEVER touches templates/ |
| Using `--rebase` merge strategy | §18.10 forbidden | Use `--squash` for chore PRs |
| Force-pushing to main | §18.10 forbidden | Always use PR flow via manager-git |
| Creating a `develop` branch | §18.10 Gitflow pattern forbidden | Use `chore/cc-update-YYYYMMDD` branch |
| Hardcoding absolute paths in agent spawn prompts | Worktree isolation violation | Use project-root-relative paths |
| Skipping ToolSearch before AskUserQuestion | InputValidationError (CLAUDE.md §19.2) | Always call ToolSearch first |
| Adding forbidden URLs (docs.moai-ai.dev, adk.moai.com, adk.moai.kr) | §17.1 URL blacklist | Use only `adk.mo.ai.kr` |

## 8. References

- CLAUDE.local.md §17 — docs-site 4-locale sync rules
- CLAUDE.local.md §18 — Enhanced GitHub Flow, merge strategies, branch naming
- CLAUDE.md §1 — HARD rules (parallel execution, AskUserQuestion-only)
- CLAUDE.md §8 — AskUserQuestion architecture, deferred tool preload
- CLAUDE.md §16 — Orchestrator self-check (delegation trigger)
- CLAUDE.md §19 — AskUserQuestion enforcement protocol
- `.claude/rules/moai/core/askuser-protocol.md` — Socratic interview structure
- `.claude/rules/moai/workflow/session-handoff.md` — paste-ready resume format
- `.moai/state/last-cc-version.json` — state file (schema in Phase 0)
- `scripts/docs-i18n-check.sh` — 4-locale consistency verifier
- `.github/labels.yml` — label definitions (type:chore, area:docs-site, priority:P2)
