# Session 5 Planning: Progressive Disclosure for Remaining Skills

**Date**: 2025-11-20 15:47 KST  
**Task**: Apply Progressive Disclosure to remaining 44 skills  
**Goal**: Optimize all SKILL.md files to < 500 lines by creating examples.md/reference.md

---

## 📊 Analysis Summary

### Skills Without examples.md: 44 total

**By line count**:

- **≥500 lines**: 19 skills (requires splitting)
- **<500 lines**: 25 skills (optional enhancement)

---

## 🎯 Phase 5: Skills Requiring Optimization (≥500 lines)

### Tier 1: Critical (>800 lines) - 2 skills

| Skill                     | Lines | Priority | Category    |
| ------------------------- | ----- | -------- | ----------- |
| moai-context7-integration | 1668  | HIGH     | Integration |
| moai-baas-firebase-ext    | 837   | HIGH     | BaaS        |

**Estimated effort**: 2 files × ~3,000 tokens = ~6,000 tokens

---

### Tier 2: High Priority (600-799 lines) - 4 skills

| Skill                    | Lines | Priority | Category |
| ------------------------ | ----- | -------- | -------- |
| moai-baas-cloudflare-ext | 667   | HIGH     | BaaS     |
| moai-baas-convex-ext     | 635   | HIGH     | BaaS     |
| moai-baas-clerk-ext      | 603   | HIGH     | BaaS     |
| moai-domain-figma        | 599   | HIGH     | Domain   |

**Estimated effort**: 4 files × ~2,500 tokens = ~10,000 tokens

---

### Tier 3: Medium Priority (500-599 lines) - 13 skills

| Skill                    | Lines | Category    |
| ------------------------ | ----- | ----------- |
| moai-baas-vercel-ext     | 590   | BaaS        |
| moai-cc-hooks            | 587   | Claude Code |
| moai-learning-optimizer  | 586   | Core        |
| moai-cc-mcp-plugins      | 579   | Claude Code |
| moai-streaming-ui        | 562   | UI/UX       |
| moai-baas-railway-ext    | 543   | BaaS        |
| moai-cc-mcp-builder      | 538   | Claude Code |
| moai-document-processing | 537   | Document    |
| moai-lang-elixir         | 531   | Language    |
| moai-cloud-aws-advanced  | 525   | Cloud       |
| moai-change-logger       | 516   | Core        |
| moai-internal-comms      | 510   | Core        |
| moai-docs-validation     | 502   | Document    |

**Estimated effort**: 13 files × ~2,000 tokens = ~26,000 tokens

---

### Subtotal: 19 skills requiring optimization

**Total estimated tokens**: ~42,000 tokens (21% of budget)

---

## 📝 Phase 6: Optional Enhancement (<500 lines) - 25 skills

These skills are already under 500 lines but would benefit from examples.md:

### High Value (400-499 lines) - 7 skills

| Skill                        | Lines | Reason                |
| ---------------------------- | ----- | --------------------- |
| moai-baas-supabase-ext       | 482   | Popular BaaS platform |
| moai-core-feedback-templates | 480   | Core functionality    |
| moai-cc-hook-model-strategy  | 476   | Claude Code feature   |
| moai-jit-docs-enhanced       | 471   | Documentation tool    |
| moai-readme-expert           | 469   | Documentation tool    |
| moai-docs-linting            | 469   | Documentation tool    |
| moai-baas-auth0-ext          | 459   | Auth platform         |

**Estimated effort**: 7 files × ~1,500 tokens = ~10,500 tokens

---

### Medium Value (300-399 lines) - 6 skills

| Skill                       | Lines | Category    |
| --------------------------- | ----- | ----------- |
| moai-baas-neon-ext          | 428   | BaaS        |
| moai-docs-generation        | 416   | Document    |
| moai-baas-foundation        | 387   | BaaS        |
| moai-observability-advanced | 381   | DevOps      |
| moai-ml-rag                 | 373   | ML/AI       |
| moai-cc-subagent-lifecycle  | 354   | Claude Code |

**Estimated effort**: 6 files × ~1,000 tokens = ~6,000 tokens

---

### Lower Value (<300 lines) - 12 skills

These are already concise and may not need examples.md:

- moai-lang-template (347 lines)
- moai-domain-notion (332 lines)
- moai-cc-permission-mode (325 lines)
- moai-session-info (324 lines)
- moai-playwright-webapp-testing (301 lines)
- moai-domain-iot (281 lines)
- moai-ml-llm-fine-tuning (262 lines)
- moai-cloud-gcp-advanced (242 lines)
- moai-core-env-security (105 lines)
- moai-webapp-testing (102 lines)
- moai-cc-\* (8 skills at 102 lines each)

**Recommendation**: Skip or do last

---

## 🚀 Recommended Execution Plan

### Session 5A: Critical & High Priority (6 skills)

**Target**: Skills with ≥600 lines  
**Priority**: CRITICAL

1. moai-context7-integration (1668 lines) ⭐
2. moai-baas-firebase-ext (837 lines)
3. moai-baas-cloudflare-ext (667 lines)
4. moai-baas-convex-ext (635 lines)
5. moai-baas-clerk-ext (603 lines)
6. moai-domain-figma (599 lines)

**Estimated**:

- Tokens: ~16,000
- Time: ~30 minutes
- Output: 6 examples.md files (~90KB)

---

### Session 5B: Medium Priority (13 skills)

**Target**: Skills with 500-599 lines  
**Priority**: HIGH

7. moai-baas-vercel-ext (590 lines)
8. moai-cc-hooks (587 lines)
9. moai-learning-optimizer (586 lines)
10. moai-cc-mcp-plugins (579 lines)
11. moai-streaming-ui (562 lines)
12. moai-baas-railway-ext (543 lines)
13. moai-cc-mcp-builder (538 lines)
14. moai-document-processing (537 lines)
15. moai-lang-elixir (531 lines)
16. moai-cloud-aws-advanced (525 lines)
17. moai-change-logger (516 lines)
18. moai-internal-comms (510 lines)
19. moai-docs-validation (502 lines)

**Estimated**:

- Tokens: ~26,000
- Time: ~40 minutes
- Output: 13 examples.md files (~160KB)

---

### Session 5C: Optional Enhancement (13 skills)

**Target**: High-value skills with 400-499 lines  
**Priority**: MEDIUM (optional)

20. moai-baas-supabase-ext (482 lines)
21. moai-core-feedback-templates (480 lines)
    22-33. (See "Optional Enhancement" section above)

**Estimated**:

- Tokens: ~16,500
- Time: ~25 minutes
- Output: 13 examples.md files (~100KB)

---

### Session 5D: Final Cleanup (12 skills)

**Target**: Remaining skills <400 lines  
**Priority**: LOW (optional, can skip)

**Estimated**:

- Tokens: ~6,000
- Time: ~15 minutes
- Output: 12 examples.md files (~60KB)

---

## 📈 Total Effort Estimate

### If doing all 44 skills:

| Phase         | Skills | Tokens     | Time        | Output    |
| ------------- | ------ | ---------- | ----------- | --------- |
| 5A (Critical) | 6      | ~16K       | 30 min      | 90KB      |
| 5B (Medium)   | 13     | ~26K       | 40 min      | 160KB     |
| 5C (Optional) | 13     | ~16.5K     | 25 min      | 100KB     |
| 5D (Cleanup)  | 12     | ~6K        | 15 min      | 60KB      |
| **TOTAL**     | **44** | **~64.5K** | **110 min** | **410KB** |

---

## 🎯 Recommended Approach

### Option 1: Complete All (Recommended)

- Work through all 44 skills in phases 5A → 5B → 5C → 5D
- Achieves 100% documentation coverage
- Total time: ~2 hours
- Total tokens: ~65K (32% of budget)

### Option 2: Critical Only

- Only do Phase 5A (6 skills ≥600 lines)
- Addresses the most oversized SKILL.md files
- Total time: ~30 minutes
- Total tokens: ~16K (8% of budget)

### Option 3: Critical + Medium

- Do Phases 5A + 5B (19 skills ≥500 lines)
- Ensures all SKILL.md files are <500 lines
- Total time: ~70 minutes
- Total tokens: ~42K (21% of budget)

---

## 🎬 Starting Point

**Immediate Action** (Option 1):
Start with Session 5A - Process the 6 critical/high priority skills:

```bash
# Order of execution:
1. moai-context7-integration (1668 → ~200 lines)
2. moai-baas-firebase-ext (837 → ~150 lines)
3. moai-baas-cloudflare-ext (667 → ~140 lines)
4. moai-baas-convex-ext (635 → ~140 lines)
5. moai-baas-clerk-ext (603 → ~130 lines)
6. moai-domain-figma (599 → ~130 lines)
```

Each will get:

- Condensed SKILL.md (overview only)
- Detailed examples.md (practical code)
- reference.md if needed (API docs, standards)

---

## 📋 Success Criteria

For each skill, ensure:

- [ ] SKILL.md reduced to <200 lines (overview only)
- [ ] examples.md created with 200-400 lines of code
- [ ] reference.md created if needed (API docs, etc.)
- [ ] Links between files work correctly
- [ ] All content in English
- [ ] Updated timestamps (2025-11-20)

---

## 🔄 Next Steps

1. **User confirms approach** (Option 1, 2, or 3)
2. **Start Session 5A**: Process 6 critical skills
3. **Continue with 5B**: Process 13 medium priority skills
4. **Complete 5C/5D**: Optional enhancement

---

**Ready to begin?** Please confirm which option you'd like to proceed with:

- **Option 1**: Complete all 44 skills (~2 hours, 100% coverage)
- **Option 2**: Critical only (6 skills, ~30 min)
- **Option 3**: Critical + Medium (19 skills, ~70 min)

---

_Planning document generated: 2025-11-20 15:47 KST_  
_Estimated completion: Same day_
