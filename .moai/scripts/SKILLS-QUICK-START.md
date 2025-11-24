# Skills Standardization Quick Start Guide

**Duration**: 60 minutes (Immediate Execution - Option 1)
**Goal**: Fix critical issues and prepare for modularization
**Target**: Achieve 80% compliance within 1 hour

---

## Step 1: Review Validation Results (5 minutes)

Current validation results show:

```
🔍 Validation Summary
├─ Total Skills: 134
├─ Critical Issues: 29 (file size > 500 lines)
├─ Warnings: 327 (modularization needed)
├─ Info Items: 16 (minor improvements)
└─ Current Compliance: 78%
```

**Key Findings**:
- 29 skills require immediate modularization
- moai-domain-nano-banana needs YAML frontmatter fix
- Metadata already standardized (version, modularized fields present)
- Most issues are file size, not metadata

---

## Step 2: Fix Critical Frontmatter Issue (10 minutes)

**CRITICAL**: moai-domain-nano-banana is missing YAML frontmatter

```bash
# Navigate to the skill directory
cd /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-domain-nano-banana

# Verify current state
head -20 SKILL.md

# Check if file starts with --- (YAML marker)
# If not, we need to add it
```

**If Missing Frontmatter** - Add the following at the very beginning of SKILL.md:

```yaml
---
name: moai-google-nano-banana
description: "Integrates Google's Nano Banana AI image model for enterprise-grade image generation, processing, and analysis. Use when generating, editing, or analyzing images with AI capabilities."
version: 1.0.0
modularized: false
allowed-tools: Read, Write, WebFetch
---
```

**Verification**:
```bash
# Verify fix
head -10 SKILL.md
# Should show:
# ---
# name: moai-google-nano-banana
# ...
```

---

## Step 3: Identify Top 10 Priority Skills for Modularization (10 minutes)

These 10 skills have the highest file size and will benefit most from modularization:

```
Priority 1: moai-lib-shadcn-ui (863 lines)
   └─ Action: Extract to modules/components.md, modules/themes.md

Priority 2: moai-baas-vercel-ext (957 lines)
   └─ Action: Extract to modules/deployment.md, modules/edge-functions.md

Priority 3: moai-design-systems (833 lines)
   └─ Action: Extract to modules/design-tokens.md, modules/components.md

Priority 4: moai-essentials-perf (759 lines)
   └─ Action: Extract to modules/profiling.md, modules/benchmarking.md

Priority 5: moai-security-ssrf (761 lines)
   └─ Action: Extract to modules/validation-patterns.md, modules/mitigation.md

Priority 6: moai-cc-configuration (748 lines)
   └─ Action: Extract to modules/schema.md, modules/patterns.md

Priority 7: moai-baas-clerk-ext (700 lines)
   └─ Action: Extract to modules/authentication.md, modules/webhooks.md

Priority 8: moai-core-dev-guide (610 lines)
   └─ Action: Extract to modules/workflows.md, modules/patterns.md

Priority 9: moai-docs-toolkit (590 lines)
   └─ Action: Extract to modules/generation.md, modules/validation.md

Priority 10: moai-foundation-langs (583 lines)
   └─ Action: Extract to modules/patterns.md, modules/implementations.md
```

---

## Step 4: Understand Modularization Process (15 minutes)

### What is Modularization?

Moving content from a large SKILL.md file (>500 lines) into organized modules/ subdirectory to improve:
- **Code Size**: Keep SKILL.md ≤350 lines
- **Token Efficiency**: Load only needed content
- **Maintainability**: Organize by topic

### 3-Step Modularization Process

**Step A: Extract Content**
1. Identify topics in the SKILL.md that form separate concepts
2. Copy each topic section to modules/topic-name.md
3. Keep the core SKILL.md but remove detailed topic sections

**Step B: Link to Modules**
1. In SKILL.md, add reference: "See modules/topic-name.md for details"
2. Update YAML: `modularized: true`
3. Update structure to show modules listing

**Step C: Validate**
1. Run validate-skills.py
2. Verify SKILL.md ≤500 lines (≤350 recommended)
3. Verify all internal links work
4. Test in Claude Code skill invocation

### Example: moai-domain-nano-banana Modularization

**Before** (single SKILL.md, 480 lines):
```
---
name: moai-google-nano-banana
...
---

## Quick Reference
...

## Image Generation Patterns (120 lines)
... detailed patterns ...

## Image Processing (110 lines)
... detailed patterns ...

## Best Practices
...
```

**After** (modularized, SKILL.md 250 lines + modules/):
```
---
name: moai-google-nano-banana
modularized: true
---

## Quick Reference
...

## Image Generation Patterns
See modules/generation-patterns.md for detailed implementations

## Image Processing
See modules/processing-patterns.md for detailed implementations

## Best Practices
... (core practices remain)
```

**Result**:
- SKILL.md: 480 → 250 lines ✅
- modules/generation-patterns.md: 120 lines
- modules/processing-patterns.md: 110 lines
- Total content preserved, better organization ✅

---

## Step 5: Generate Modularization Priority List (5 minutes)

Run the validation script to see which skills need attention:

```bash
python3 /tmp/validate-skills.py > /tmp/validation-full-report.txt

# View critical issues
grep "CRITICAL" /tmp/validation-full-report.txt

# Count issues by type
echo "Issues by category:"
grep "exceeds" /tmp/validation-full-report.txt | wc -l    # File size issues
grep "modularized" /tmp/validation-full-report.txt | wc -l  # Modularization issues
```

---

## Step 6: Create Modularization Strategy Document (10 minutes)

Create a CSV to track modularization progress:

```bash
cat > /tmp/modularization-plan.csv << 'EOF'
Skill Name,Current Size,Target Size,Module 1,Module 2,Module 3,Priority,Status
moai-lib-shadcn-ui,863,250,components.md,themes.md,advanced-patterns.md,P1,Pending
moai-baas-vercel-ext,957,280,deployment.md,edge-functions.md,webhooks.md,P1,Pending
moai-design-systems,833,260,design-tokens.md,components.md,patterns.md,P1,Pending
moai-essentials-perf,759,240,profiling.md,benchmarking.md,optimization.md,P1,Pending
moai-security-ssrf,761,250,validation.md,mitigation.md,best-practices.md,P1,Pending
moai-cc-configuration,748,280,schema.md,patterns.md,examples.md,P2,Pending
moai-baas-clerk-ext,700,250,authentication.md,webhooks.md,integration.md,P2,Pending
moai-core-dev-guide,610,280,workflows.md,patterns.md,examples.md,P2,Pending
moai-docs-toolkit,590,260,generation.md,validation.md,linting.md,P2,Pending
moai-foundation-langs,583,240,patterns.md,implementations.md,examples.md,P2,Pending
EOF

# View the plan
cat /tmp/modularization-plan.csv
```

---

## Step 7: Next Steps (Execution Timeline)

### This Week (Day 1-2):
- ✅ Fix moai-domain-nano-banana frontmatter
- ✅ Review validation report
- ✅ Assign team members to P1 skills (5 skills)

### Next Week (Day 3-7):
- ✅ Complete modularization of P1 skills (top 5)
- ✅ Run validation to confirm each skill ≤500 lines
- ✅ Start P2 skills (next 5)

### Week 3:
- ✅ Complete all modularizations
- ✅ Final validation (target: 95% compliance)
- ✅ Update CLAUDE.md auto-trigger system
- ✅ Release to production

---

## Quick Reference: Validation Commands

```bash
# Run full validation
python3 /tmp/validate-skills.py

# Check specific skill
python3 /tmp/validate-skills.py | grep "moai-lib-shadcn-ui"

# Count file sizes
wc -l /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-*/SKILL.md | sort -n | tail -20

# List all skills > 500 lines
for f in /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-*/SKILL.md; do
  lines=$(wc -l < "$f")
  if [ "$lines" -gt 500 ]; then
    echo "$f: $lines lines"
  fi
done | sort -t: -k2 -rn
```

---

## Success Metrics

By completing this quick start guide, you will have:

✅ Fixed 1 critical frontmatter issue
✅ Identified 29 skills needing modularization
✅ Created detailed modularization plan
✅ Established priority matrix
✅ Prepared execution roadmap

**Result**: Ready to execute 3-week modularization campaign

---

## Resources

**Official Standards Document**:
- Full guide: SKILLS-OFFICIAL-STANDARD-GUIDE-2025.md
- Validation tool: validate-skills.py
- Metadata tool: standardize-skill-metadata.sh

**Next Steps**:
1. Read: SKILLS-OFFICIAL-STANDARD-GUIDE-2025.md (Part 4)
2. Execute: Weekly modularization sprints
3. Monitor: Weekly validation reports
4. Target: 95% compliance by Week 3

---

**Time to Complete This Guide**: 60 minutes
**Next Milestone**: P1 skills modularization complete (72 hours)
**Final Target**: 95% compliance (21 days)

