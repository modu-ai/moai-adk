# v2.12.0 Release Notes Enhancement Strategy

**Decision Date**: 2026-04-22  
**Current Status**: v2.12.0 released 2026-04-17 (5 days ago)  
**Stars**: 937 (target: 1,000 by Q2 end)

---

## Context

v2.12.0 published with **technical release notes** covering major features:
- Design workflow absorption (SPEC-AGENCY-ABSORB-001)
- Database integration (SPEC-DB series)
- Profile setup hardening
- 128 total commits since v2.11.0

**Gap**: Current release notes are engineering-focused. A **marketing appendix** can significantly increase discoverability and community engagement without waiting for v2.13.0.

---

## Three Options Evaluated

### Option A: Update v2.12.0 Notes In Place (Recommended)

**Approach**: Append the marketing appendix to the existing v2.12.0 release page at GitHub (no new release/tag needed).

**Pros**:
- ✅ Immediate implementation (30 minutes to edit + verify)
- ✅ No Git operations or release machinery
- ✅ Audience already sees v2.12.0 as latest
- ✅ GitHub shows release notes prominently in repo header
- ✅ Links to roadmap (v2.13.0 teaser) drive engagement toward next release
- ✅ Community discord/email can link to "updated release notes" without confusion

**Cons**:
- ⚠️ In-place edits can break GitHub notification tracking (subscribers won't be re-notified)
- ⚠️ Late-stage marketing may feel secondary to early technical readers
- ⚠️ Release notes editor has 5-day history; very old edits show as "updated 5 days ago" (UX signal noise)

**Implementation**:
1. Copy the appendix block from `github-release-v2.12.0-enhanced.md`
2. Log into GitHub repo as maintainer
3. Edit the v2.12.0 release page → add appendix at bottom
4. Preview markdown rendering (GitHub's editor shows live preview)
5. Save (no new release created)
6. Post announcement in Discussions/Discord: "Updated v2.12.0 release notes with roadmap and community links"

**Estimated Effort**: 20 minutes  
**Risk**: Low (edit can be reverted; no breaking changes)

---

### Option B: Publish v2.12.1 Patch with Enhanced Notes

**Approach**: Create a v2.12.1 patch release (no code changes, same binary as v2.12.0) with expanded release notes.

**Pros**:
- ✅ Clean separation: v2.12.0 = technical notes, v2.12.1 = marketing + community
- ✅ Freshness signal: newer release date (2026-04-22 vs 2026-04-17) attracts attention
- ✅ Social media angle: "We listened — here's what's next" message
- ✅ Subscribers get re-notified of new release (GitHub notification UX)
- ✅ Badge/Release UI shows "latest" as v2.12.1 for ~1 week
- ✅ Awesome Lists PRs can reference v2.12.1 as "latest" without looking stale

**Cons**:
- ⚠️ Semantic versioning purists dislike "marketing-only" patches (not aligned with SemVer philosophy)
- ⚠️ Dilutes release history with non-code version (confuses some users)
- ⚠️ Requires git tag, GitHub release, potentially CI/CD runs (hostvn dependencies)
- ⚠️ 10–15 minute Git + release workflow overhead
- ⚠️ Risk of accidentally bumping binary version if CI auto-builds

**Implementation**:
1. Create commit with updated `.moai/config/sections/system.yaml` version line (cosmetic, v2.12.1)
2. Create git tag `v2.12.1` and push
3. Create GitHub Release with full enhanced notes
4. Verify CI does **not** trigger release workflow (mark as "pre-release" or "draft" first, then publish)

**Estimated Effort**: 25 minutes (tag + GH release editor)  
**Risk**: Medium (requires git operations; verify CI behavior)

---

### Option C: Fast-Track v2.13.0 as "Marketing Cut"

**Approach**: Accelerate v2.13.0 release to next week (2026-04-28), pack with early features from [Unreleased] section, market as "community-driven roadmap cut."

**Pros**:
- ✅ Strongest signal: "We listen to community feedback"
- ✅ Momentum: Fresh release week after Awesome Lists submissions compounds discovery
- ✅ Roadmap visibility: v2.13.0 features land 4 weeks earlier than normal
- ✅ Social media story: "v2.12.0 feedback → fast-tracked v2.13.0" narrative
- ✅ Extended release notes can promote v2.14.0 roadmap (4 months out)

**Cons**:
- ❌ Compressed testing cycle (1 week instead of typical 3–4 weeks)
- ❌ Risk of QA gaps: Unreleased features may have undiscovered edge cases
- ❌ LSP refinements (Unreleased) are major; pulling forward increases bug risk
- ❌ Requires full CI/CD, signing, documentation updates, changelog finalization
- ❌ Messaging complexity: "What's in v2.13.0 if v2.12.0 was 5 days ago?" (confuses timing-sensitive users)
- ❌ High effort: 2–3 hours of release process, testing, docs sync

**Implementation**:
1. Stabilize Unreleased features in short QA sprint
2. Bump version to v2.13.0 in Go code + CHANGELOG
3. Create release PR, run full test suite
4. Tag and publish (full CI/CD run)
5. Finalize all 4-language docs and README updates

**Estimated Effort**: 120–180 minutes (full release cycle)  
**Risk**: High (compressed timeline, QA pressure)

---

## Evaluation Matrix

| Criterion | Option A (In-Place Update) | Option B (v2.12.1 Patch) | Option C (v2.13.0 Fast-Track) |
|-----------|---------------------------|-------------------------|-------------------------------|
| **Implementation Speed** | 20 min | 25 min | 150 min |
| **Risk Level** | Low | Medium | High |
| **Community Benefit** | High | High | Highest |
| **Awesome Lists Alignment** | Good | Best | Best |
| **Semantic Correctness** | ✅ (no version bump) | ⚠️ (empty patch) | ✅ (feature release) |
| **Subscriber Re-notification** | ❌ | ✅ | ✅ |
| **Marketing Lift** | Good | Good | Excellent |
| **Team Bandwidth** | Minimal | Minimal | High |
| **Release Pace Signal** | Maintains 3-4 week cycle | Slightly accelerated | Doubled pace (risky) |

---

## Recommendation

### **PRIMARY: Option A (In-Place Update)**

**Why**: Best risk-to-reward ratio for immediate action.

1. **Immediate execution** — No Git machinery, no CI/CD risk, edit completed today
2. **Clean messaging** — "Enhanced release notes" is straightforward to communicate
3. **Awesome Lists compatible** — v2.12.0 is stable, PR text references it; no freshness concerns
4. **Reversible** — If marketing appendix underperforms or needs tweaks, easy to refine
5. **Upstream roadmap** — Links to v2.13.0 features keep community engaged without forcing early release

**Execution Timeline**:
- 2026-04-22 (today): Copy appendix, edit GitHub release, post announcement
- 2026-04-23: Verify page rendering, share on social media
- 2026-04-24+: Submit Awesome Lists PRs (reference updated v2.12.0 notes)

---

### **SECONDARY: Option B (v2.12.1 Patch)**

**When to use Option B instead**:
- If analytics show v2.12.0 release page has low traffic (< 100 views/day) — a refresh (v2.12.1) would re-surface it
- If social media engagement on v2.12.0 announcement was weak — a new release generates fresh notifications
- If team wants to test release machinery before v2.13.0 (dry-run value)

**If choosing Option B, use this workflow**:
```bash
# Ensure clean git state
git status

# Create patch version bump (cosmetic)
# Update version in internal/cli/version.go or .moai/config/sections/system.yaml

# Tag and push
git tag -a v2.12.1 -m "Marketing and community link enhancements for v2.12.0"
git push origin v2.12.1

# Go to GitHub, create release from tag with full enhanced notes
# Mark as "Latest release" (default), DO NOT mark as "Pre-release"
```

---

### **NOT RECOMMENDED: Option C (v2.13.0 Fast-Track)**

**Rationale for rejection**:
- Unreleased features (LSP refinements, Agent Profiling) have not undergone full QA cycle
- v2.13.0 historically takes 3–4 weeks; cramming into 1 week is a quality risk
- Awesome Lists PRs can reference v2.12.0 (stable, 5 days old); no staleness concern
- Team velocity: v2.13.0 work is underway; forcing release steals focus from ongoing development

**When Option C becomes viable**:
- After v2.13.0 reaches code-complete state (in ~2 weeks)
- When risk assessment shows Unreleased features are stable enough to promote
- If community feedback prioritizes early release (monitor Discussions/Discord)

---

## Final Decision: Option A (In-Place Update)

**Action Items**:

- [ ] **Today (2026-04-22, 4 PM UTC)**: Edit v2.12.0 release page, append marketing block, save
  - Go to https://github.com/modu-ai/moai-adk/releases/edit/v2.12.0
  - Scroll to bottom, paste content from `github-release-v2.12.0-enhanced.md`
  - Click "Update Release"
  
- [ ] **Today (6 PM UTC)**: Post announcement in Discussions
  - Title: "🎉 Updated v2.12.0 Release Notes with Community Links & Roadmap"
  - Body: Highlight new docs links, community channels, v2.13.0 teaser
  - Pin the discussion
  
- [ ] **Today (6:30 PM UTC)**: Social media push
  - Twitter: "Just updated v2.12.0 release notes with community links and v2.13.0 roadmap. We're 60 stars from 1K — thanks for the support! Join our Discord [LINK]"
  - LinkedIn: Professional tone, emphasize team milestone and roadmap transparency

- [ ] **2026-04-23**: Verify page rendering on GitHub (screenshot for records)

- [ ] **2026-04-24+**: Begin Awesome Lists submissions (reference updated v2.12.0 in all PR bodies)

---

## Post-Submission Metrics

Track these KPIs over 4 weeks post-Option A implementation:

- **v2.12.0 page views**: GitHub can show traffic via /insights/network (baseline: establish from 4 PM today onwards)
- **Discord referrals**: Add `?ref=release-notes` to Discord link, track join source
- **GitHub stars**: Project-level growth rate (target: 10–20 stars/week)
- **Awesome Lists merges**: PRs merged into accepted lists (target: 5+ by 2026-05-10)
- **Documentation traffic**: adk.mo.ai.kr page views (baseline from GA if available)

---

**Status**: APPROVED for implementation as Option A  
**Owner**: Maintainer (GitHub release editor access required)  
**Timeline**: 2026-04-22 (today), execution window < 1 hour  
**Fallback**: If issues with release page editing, use Option B (v2.12.1 patch, 25 min overhead)
