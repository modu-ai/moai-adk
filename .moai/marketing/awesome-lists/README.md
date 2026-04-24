# MoAI-ADK Marketing Materials — Awesome Lists + Release Strategy

This directory contains marketing preparation materials for MoAI-ADK v2.12.0, focusing on Awesome Lists discovery and GitHub Release enhancement.

## 📋 Files Overview

### 1. **submission-plan.md** (Primary Strategy Document)

**Purpose**: Complete plan for submitting MoAI-ADK to 7+ Awesome Lists over the next 4 weeks.

**Key Sections**:
- ✅ **2 Verified ACTIVE Lists** (ready to submit immediately):
  - `awesome-ai-agents` (e2b-dev) — CONFIRMED active (last push 2026-02-26)
  - `awesome-go` (avelino) — CONFIRMED active (last push 2026-04-22)

- ⚠️ **5 Unverified Candidates** (require manual verification before submission):
  - `awesome-claude-code` (W&B?) — Highest priority if confirmed
  - `awesome-llm-tools` (taishi-i?)
  - `awesome-developer-tools` (johnwulp?)
  - `awesome-mcp` (dirk1983?)
  - `awesome-ai-coding` (phodal?)

**Workflow**:
1. Phase 1: Manual verification of unverified lists (2026-04-22 to 2026-04-23)
2. Phase 2: Draft PR templates for each list (2026-04-24 to 2026-04-26)
3. Phase 3: Submit PRs in priority order (2026-04-27 to 2026-05-05)
4. Phase 4: Track merges and celebrate wins (2026-05-06 onwards)

**Expected Outcome**: 6–8 merged Awesome Lists by 2026-05-31, contributing 75–150 new stars toward the 1,000-star milestone.

---

### 2. **github-release-v2.12.0-enhanced.md** (Marketing Appendix)

**Purpose**: Ready-to-paste marketing content for GitHub v2.12.0 release notes.

**Contents** (~400 words):
- 📚 **Documentation Links** — 4 languages (en/ko/ja/zh) with quick links to docs, getting started guides, design workflow, and database docs
- 🌏 **Community Section** — Discord [PLACEHOLDER], GitHub Discussions [PLACEHOLDER], Issues
- ⭐ **Milestone Moment** — Frame reaching 1K stars (currently 937, ~60 away)
- 🗓️ **Roadmap Teaser** — Pull 5 highlights from [Unreleased] section in CHANGELOG.md to preview v2.13.0
- 📖 **Multilingual READMEs** — Links to en/ko/ja/zh README files

**Status**: Ready to paste into GitHub release editor (no edits needed).

**Implementation**: See `next-release-strategy.md` — **Recommendation is Option A (in-place update)**.

---

### 3. **next-release-strategy.md** (Release Timing Decision)

**Purpose**: Analysis of three strategies for deploying the marketing appendix:

| Strategy | Implementation | Risk | Effort | Result |
|----------|-----------------|------|--------|--------|
| **Option A: In-Place Update** (RECOMMENDED) | Edit v2.12.0 release notes, append marketing block | Low | 20 min | No new release; immediate impact |
| **Option B: v2.12.1 Patch** | Create empty v2.12.1 release (same binary) with enhanced notes | Medium | 25 min | Fresh notification to subscribers; minor SemVer purist concern |
| **Option C: v2.13.0 Fast-Track** | Accelerate v2.13.0 release with extra marketing push | High | 120+ min | Strongest signal; risky compressed timeline |

**Recommendation**: **Option A (In-Place Update)**
- Immediate execution (no Git machinery)
- Cleanest messaging ("Enhanced release notes")
- Compatible with Awesome Lists submissions
- Reversible if adjustments needed

**Action**: Go to GitHub release editor for v2.12.0 and append the content from `github-release-v2.12.0-enhanced.md` at the bottom.

---

### 4. **pr-drafts/** (To Be Created — Phase 2)

**Purpose**: Individual PR body templates for each verified Awesome List.

**Expected Files**:
- `PR-awesome-ai-agents.md` — Formatted PR body ready to copy into GitHub PR creation form
- `PR-awesome-go.md` — Go CLI angle, alphabetical ordering notes
- `PR-awesome-claude-code.md` — (if verified)
- `PR-awesome-llm-tools.md` — (if verified)
- ... etc.

**Status**: Created during Phase 2 (2026-04-24 to 2026-04-26).

---

### 5. **verification-results.md** (To Be Created — Phase 1)

**Purpose**: Document the manual verification process and findings.

**Content**:
- URLs verified (with links)
- Last commit dates for each unverified list
- CONTRIBUTING.md review (contribution rules, alphabetical requirements, etc.)
- Confidence rating for each list (High/Medium/Low)
- Final decision: Submit (✅) or Skip (❌)

**Status**: Created during Phase 1 (2026-04-22 to 2026-04-23).

---

### 6. **submission-log.md** (To Be Created — Phase 3)

**Purpose**: Tracking of all submitted PRs with merge status and outcomes.

**Columns**:
- List name
- PR URL
- Submission date
- Status (Merged / In Review / Rejected / Draft)
- Merge date (if applicable)
- Feedback/Notes

**Status**: Created during Phase 3 (2026-04-27 onwards).

---

## 🎯 Quick Start

### Immediate Action (Today — 2026-04-22)

1. **Verify unverified lists** (Phase 1):
   - Open each unverified list candidate in browser
   - Confirm repo exists, has active maintainer, accepts contributions
   - Document findings in `verification-results.md`

2. **Decide on release note strategy**:
   - Read `next-release-strategy.md` (3-minute read)
   - If agree with Option A (recommended): Schedule GitHub release editor edit for tomorrow
   - If prefer Option B or C: Plan git/release workflow

### Week 1 (2026-04-24 to 2026-04-26)

3. **Draft PR templates** (Phase 2):
   - Create `pr-drafts/` directory
   - Generate one `.md` file per verified list
   - Each file contains formatted PR title and body, ready to copy-paste

### Week 2 (2026-04-27 to 2026-05-05)

4. **Submit PRs** (Phase 3):
   - P1 lists first (awesome-ai-agents, awesome-go, awesome-claude-code if verified)
   - Space submissions by 1 day minimum
   - Monitor PRs for feedback
   - Update `submission-log.md` with PR URLs

### Week 3+ (2026-05-06 onwards)

5. **Track & celebrate** (Phase 4):
   - Weekly check of PR merge status
   - Share wins on social media
   - Analyze traffic impact

---

## 📊 Success Metrics

### Target Outcomes by 2026-05-31

**Conservative**: 6–8 merged Awesome Lists, 75–150 new stars  
**Stretch**: 10+ merged Awesome Lists, 150–250 new stars  

**Milestone**: Currently at 937 stars; 1,000-star goal achievable with Awesome Lists + organic growth.

---

## 💡 Notes for the Team

1. **Two verified lists are immediately ready**: awesome-ai-agents and awesome-go. These can be submitted today if desired (no verification step needed).

2. **Unverified lists require 1–2 hours of manual research**: Search GitHub, check last commits, review contribution rules. Most will likely be confirmed as active.

3. **Release note enhancement is separate from Awesome Lists**: Can implement Option A today (20 minutes) while verification runs in parallel.

4. **Awesome Lists have high success rates for mature projects**: 60–80% typical. MoAI-ADK's maturity (937 stars, v2.12.0 stable, active maintenance) suggests 70–100% merge rate.

5. **Discord / Discussions placeholders**: Replace `[PLACEHOLDER]` in `github-release-v2.12.0-enhanced.md` with actual links when Discord server and GitHub Discussions are set up.

---

## 📝 File Checklist

- [x] `submission-plan.md` — Complete, with verified lists separated from unverified
- [x] `github-release-v2.12.0-enhanced.md` — Ready to paste, 4 languages, ~400 words
- [x] `next-release-strategy.md` — Recommendation: Option A with detailed pros/cons
- [ ] `pr-drafts/` — To be created Phase 2
- [ ] `verification-results.md` — To be created Phase 1
- [ ] `submission-log.md` — To be created Phase 3

---

**Status**: Marketing materials prepared and ready for execution  
**Last Updated**: 2026-04-22  
**Timeline**: Phase 1 (Today) → Phase 4 (2026-05-31)  
**Contact**: Maintainer (for GitHub release edit access)
