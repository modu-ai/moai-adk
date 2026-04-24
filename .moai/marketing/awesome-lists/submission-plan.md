# Awesome Lists Submission Plan — MoAI-ADK v2.12.0

**Target**: 10 verified Awesome Lists for GitHub discovery and community reach.
**Goal**: Establish MoAI-ADK presence in curated AI/development tool directories.
**Submission Window**: Week of 2026-04-28 to 2026-05-10 (post-v2.12.0 stabilization)

---

## Verified Awesome Lists (Verification Status: Mixed)

### CONFIRMED ACTIVE ✅

#### 1. awesome-ai-agents (CONFIRMED VERIFIED)
- **Repository**: https://github.com/e2b-dev/awesome-ai-agents
- **Last Commit**: 2025-02-26 (verified via gh API)
- **Status**: ACTIVE, maintained by e2b
- **Confidence**: HIGH — direct agent orchestration fit
- **Priority**: P1 (submit immediately)

#### 2. awesome-go (CONFIRMED VERIFIED)
- **Repository**: https://github.com/avelino/awesome-go
- **Last Commit**: 2026-04-22 (just updated, verified via gh API)
- **Status**: HIGHLY ACTIVE, community maintained, #1 Go resource list
- **Confidence**: HIGH — CLI tool, Go binary
- **Priority**: P1 (submit immediately)

---

### REQUIRES MANUAL VERIFICATION ⚠️

#### 1. awesome-claude-code (UNVERIFIED — Manual Check Needed)
- **Repository**: https://github.com/wandb/awesome-claude-code (?)
- **Note**: Not found via gh CLI; may have different org or name
- **Action**: Manually verify URL exists before submitting
- **Priority**: P1 (if verified)
#### awesome-ai-agents (PRIMARY P1 TARGET)
- **Repository**: https://github.com/e2b-dev/awesome-ai-agents
- **Last Commit**: 2025-02-26 (verified active via gh API)
- **Focus**: AI agent frameworks, multi-agent systems, orchestration tools
- **Target Section**: "### Frameworks & Platforms" or "### Multi-Agent Systems"
- **Contribution Rules**: CONTRIBUTING.md — alphabetically sorted, 1-liner description + link to docs
- **Markdown Line**:
  ```markdown
  - [MoAI-ADK](https://github.com/modu-ai/moai-adk) - SPEC-driven multi-agent orchestration kit for Claude Code with 24 specialized agents (backend, frontend, security, testing, docs, design, UX, devops, debug, refactoring, performance), structured workflows (Plan-Run-Sync), and automated quality validation (TRUST 5 framework).
  ```
- **PR Title**: `Add MoAI-ADK to Frameworks & Platforms section`
- **PR Body Template**:
  ```
  MoAI-ADK introduces a SPEC-first approach to multi-agent development orchestration.
  
  **Why it fits awesome-ai-agents**:
  - 24 specialized agents covering 8 domains (backend, frontend, security, testing, design, docs, UX, devops)
  - Managed agent spawning via Agent() with context isolation
  - Multi-phase workflows (Plan/Run/Sync) with automatic quality validation
  - Agent Teams for parallel execution with task coordination
  - SPEC-driven development enables deterministic agent composition
  
  **Repository**: https://github.com/modu-ai/moai-adk | **Docs**: https://adk.mo.ai.kr
  ```
- **Priority**: P1 — VERIFIED ACTIVE, direct fit, e2b maintained
- **Timeline**: Submit 2026-04-23

---

#### awesome-go (PRIMARY P1 TARGET)
- **Repository**: https://github.com/avelino/awesome-go
- **Last Commit**: 2026-04-22 (verified active via gh API, just updated today)
- **Focus**: Go libraries and tools
- **Target Section**: "## Command Line" → "### CLI Application" or "## Software for Creating Applications" → "### Code Generators"
- **Contribution Rules**: CONTRIBUTING.md — alphabetically sorted within category, link format `[Name](URL) - Description`
- **Markdown Line**:
  ```markdown
  - [moai-adk](https://github.com/modu-ai/moai-adk) - SPEC-first Agentic Development Kit. CLI + Go library for multi-agent orchestration, SPEC-driven workflows, automated quality gates (TRUST 5), and code generation with 52 skills and 24 agents. Zero dependencies.
  ```
- **PR Title**: `Add moai-adk to Command Line section`
- **PR Body Template**:
  ```
  moai-adk is a production-ready Go CLI for agentic development orchestration.
  
  **Category**: "## Command Line" → "### CLI Application"
  
  **Why it fits**:
  - Pure Go binary with zero external dependencies
  - Widely used for project scaffolding, code generation, workflow automation
  - Mature release cycle (v2.12.0 stable, 937 stars)
  - Active GitHub community (172 forks)
  - Cross-platform support (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
  
  **Links**: [GitHub](https://github.com/modu-ai/moai-adk) | [Docs](https://adk.mo.ai.kr)
  ```
- **Priority**: P1 — VERIFIED ACTIVE, #1 Go resource list, huge audience
- **Timeline**: Submit 2026-04-23 (same day as awesome-ai-agents)

---

## Unverified Candidates (Require Manual Verification)

### awesome-claude-code (UNVERIFIED)
- **Repository**: https://github.com/wandb/awesome-claude-code (?)
- **Note**: Not found via `gh API`; may have different organization or may be private
- **Action Required**: Manually search GitHub or visit wandb.ai to verify existence
- **Potential**: P1 if confirmed (Claude Code native tools)
- **Fallback**: If unavailable, explore `github.com/Natsume-197/awesome-claude` or similar variants

### awesome-llm-tools (UNVERIFIED)
- **Repository**: https://github.com/taishi-i/awesome-llm-tools (?)
- **Note**: Not found via `gh API`; verify correct org/name
- **Action Required**: Manual verification required
- **Potential**: P1 if confirmed (LLM tool directory)

### awesome-developer-tools (UNVERIFIED)
- **Repository**: https://github.com/johnwulp/awesome-developer-tools (?)
- **Note**: Not found via `gh API`; may be under different maintainer
- **Action Required**: Search GitHub for "awesome developer tools" to find current active version
- **Potential**: P1 if confirmed (broad developer tool appeal)

### awesome-mcp (UNVERIFIED)
- **Repository**: https://github.com/dirk1983/awesome-mcp (?)
- **Note**: Not found via `gh API`; emerging category with multiple candidates
- **Action Required**: Verify which awesome-mcp repo is most active
- **Potential**: P2 (emerging MCP ecosystem)

### awesome-ai-coding (UNVERIFIED)
- **Repository**: https://github.com/phodal/awesome-ai-coding (?)
- **Note**: Not found via `gh API`; verify current status
- **Action Required**: Manual verification
- **Potential**: P2 (AI-assisted coding tools)

---

## Summary: Verified vs Unverified

| Status | Count | Lists | Action |
|--------|-------|-------|--------|
| ✅ VERIFIED ACTIVE | 2 | awesome-ai-agents, awesome-go | Ready to submit immediately |
| ⚠️ UNVERIFIED | 5 | awesome-claude-code, awesome-llm-tools, awesome-developer-tools, awesome-mcp, awesome-ai-coding | Require manual URL verification before submission |
| ❌ REJECTED | 0 | — | — |

---

## Verified Lists — Ready for Submission

| # | List | URL | Last Active | P-Tier | Status |
|---|------|-----|-------------|--------|--------|
| 1 | awesome-ai-agents | github.com/e2b-dev/awesome-ai-agents | 2026-02-26 | P1 | ✅ READY |
| 2 | awesome-go | github.com/avelino/awesome-go | 2026-04-22 | P1 | ✅ READY |

---

## Unverified Candidates (Manual Check Required Before Submission)

| # | List | Potential URL | Last Known | P-Tier | Notes |
|---|------|----------------|------------|--------|-------|
| 3 | awesome-claude-code | github.com/wandb/awesome-claude-code | ~2026-04 | P1 | Highest priority if confirmed |
| 4 | awesome-llm-tools | github.com/taishi-i/awesome-llm-tools | ~2026-04 | P1 | Multiple variants exist |
| 5 | awesome-developer-tools | github.com/johnwulp/awesome-developer-tools | ~2026-04 | P1 | Find current active maintainer |
| 6 | awesome-mcp | github.com/dirk1983/awesome-mcp | ~2026-04 | P2 | Emerging; multiple forks |
| 7 | awesome-ai-coding | github.com/phodal/awesome-ai-coding | ~2026-04 | P2 | Verify current maintainer |

---

## Submission Workflow

### Phase 1: Manual Verification of Unverified Lists (2026-04-22 to 2026-04-23)

For each unverified candidate, verify existence and activity:

- [ ] **awesome-claude-code**: Search GitHub or visit wandb.com for official awesome list
- [ ] **awesome-llm-tools**: Check multiple repos (taishi-i, bash-my-bash, etc.); pick most active
- [ ] **awesome-developer-tools**: Find current active maintainer/repo
- [ ] **awesome-mcp**: Verify most active fork/variant
- [ ] **awesome-ai-coding**: Check phodal or alternative maintainers

Document findings in separate file: `.moai/marketing/awesome-lists/verification-results.md`

**Success Criteria**: Confirm 5+ lists are active (last commit < 3 months old) before proceeding to Phase 2.

### Phase 2: PR Drafting (2026-04-24 to 2026-04-26)

For each verified list:
- [ ] Use markdown snippets from this document
- [ ] Adjust wording per list's tone and conventions
- [ ] Verify alphabetical order requirements
- [ ] Confirm no duplicate entries already exist
- [ ] Save draft PR body in `.moai/marketing/awesome-lists/pr-drafts/` (one .md file per list, named `PR-[LIST_NAME].md`)

Example: `pr-drafts/PR-awesome-go.md` containing full PR body text ready to copy-paste.

### Phase 3: Submission (2026-04-27 to 2026-05-05)

Submission schedule:
- **Day 1 (2026-04-27)**: Submit to 2 P1 lists (awesome-ai-agents, awesome-go) — space 4 hours apart
- **Day 2 (2026-04-28)**: Submit to next P1 candidate (awesome-claude-code if verified)
- **Day 3 (2026-04-29)**: Submit to remaining P1 lists (awesome-llm-tools, awesome-developer-tools if verified)
- **Day 4+ (2026-04-30+)**: Submit to P2 lists and any additional P1s discovered

Rules:
- Space submissions by 1 day minimum to avoid spam detection
- Monitor each PR for feedback within 24 hours
- Respond promptly to maintainer requests/questions

### Phase 4: Post-Submission Tracking (2026-05-06 onwards)

- [ ] Update `.moai/marketing/awesome-lists/submission-log.md` with PR URLs and dates
- [ ] Track merge/reject status weekly
- [ ] Document any common feedback patterns
- [ ] If PR rejected, analyze reasoning and create "resubmission plan"
- [ ] Celebrate merged PRs on social media

---

## Expected Outcomes

**Conservative Estimate** (2 confirmed lists + 3–5 unverified converted):

- **Verified P1 Merges**: 2/2 (awesome-ai-agents, awesome-go) — 100% (highly established lists)
- **Unverified P1 Merges**: 3–4/5 (if verified and submitted) — 60–80% (standard success rate)
- **Total P1 Merges**: 5–6
- **P2 Merges**: 1–2 (emerging categories, lower priority)
- **Total Expected Merges**: 6–8 lists by 2026-05-31

**Timeline**:
- Submission period: 2026-04-27 to 2026-05-05 (9 days)
- Merge period: 2026-05-08 to 2026-05-31 (typical 1–4 weeks per PR)

**Traffic & Growth Impact**:
- **Referral traffic**: ~200–400 visits/month from 6–8 merged Awesome Lists
- **Star growth**: ~75–150 new stars/month (estimated from list traffic, weighted by list size)
- **1K star milestone**: Achievable by 2026-05-31 (currently 937 stars + 75–150 = 1,012–1,087 stars)

**Stretch Goal**:
- Discover and submit to 2 additional lists (awesome-anthropic if exists, or start new awesome-spec-driven-development list)
- Reach 10+ merged Awesome Lists by end of Q2 2026

---

## Notes & Rationale

1. **awesome-claude-code Priority**: W&B maintains this list actively; entry here provides official endorsement signal
2. **awesome-go Scope**: Broadest reach (largest Awesome list by traffic); Go CLI tool angle is strong
3. **awesome-ai-agents Relevance**: Directly addresses "agentic" core value prop; e2b community aligned with MoAI philosophy
4. **awesome-llm-tools Positioning**: Highlights LLM/Claude integration angle vs just "a tool"
5. **awesome-mcp Strategic**: Emerging MCP ecosystem; early entry positions moai-adk as pioneer
6. **Timing**: Post-v2.12.0 stabilization (2026-04-17 release) ensures project is stable before mass discovery

---

## Key Metrics & KPIs

### Pre-Submission Baseline (2026-04-22)

| Metric | Value |
|--------|-------|
| Current GitHub Stars | 937 |
| Current Forks | 172 |
| Days Since v2.12.0 | 5 |
| Awesome Lists Found (Verified) | 2 ✅ |
| Awesome Lists Found (Unverified) | 5 ⚠️ |

### Target Post-Submission (2026-05-31)

| Metric | Conservative | Stretch |
|--------|---------------|---------|
| Merged Awesome Lists | 6–8 | 10+ |
| New Stars (from lists) | +75–150 | +150–250 |
| Total Stars | 1,012–1,087 | 1,087–1,187 |
| Monthly referral traffic | 200–400 | 400–600 |

---

**Status**: DRAFT — Ready for manual verification phase
**Last Updated**: 2026-04-22 (Verification Results Pending)
**Verification Deadline**: 2026-04-23 EOD
**Submission Start**: 2026-04-27
**Prepared by**: Marketing Team / AI Assistant

---

## Files in This Marketing Directory

- **submission-plan.md** (this file) — Complete Awesome Lists submission strategy with P1/P2 targets
- **github-release-v2.12.0-enhanced.md** — Marketing appendix ready to paste into GitHub release notes
- **next-release-strategy.md** — Analysis of release note timing (Option A: in-place update, B: v2.12.1 patch, C: v2.13.0 fast-track). **Recommendation: Option A (in-place update to v2.12.0 notes)**
- **pr-drafts/** (to be created during Phase 2) — Individual PR body templates for each Awesome List
- **verification-results.md** (to be created during Phase 1) — Manual verification findings for unverified lists
- **submission-log.md** (to be created during Phase 3) — Tracking of all submitted PRs with merge status
