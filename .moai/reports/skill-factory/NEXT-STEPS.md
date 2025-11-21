# IMMEDIATE NEXT STEPS

**Priority**: HIGH  
**Timeline**: 5 hours for Phase 2 completion  
**Updated**: 2025-11-21

---

## Phase 2 Completion Tasks

### Task 1: Process 5 Remaining Skills (75 minutes)

**Skills to Process**:
1. `xlsx` - Excel file processing
2. `moai-document-processing-unified` - Document orchestrator
3. `pptx` - PowerPoint processing
4. `docx` - Word document processing
5. (1 additional file utility)

**Required Structure**:
```markdown
# Quick Reference (Level 1)
- What it does (2-3 sentences)
- When to use (bullet points)
- Core concepts (3-5 items)

# Implementation (Level 2)
- Step-by-step patterns
- Code examples
- Integration guides

# Advanced (Level 3)
- Expert patterns
- Edge cases
- Performance optimization
```

**Command**:
```bash
# For each skill
cd .claude/skills/[skill-name]
# Apply 3-level structure to SKILL.md
# Copy to templates directory
# Verify synchronization
```

---

### Task 2: Reduce 9 Oversized Skills (180 minutes)

**Skills Requiring Reduction** (Current → Target):
1. `moai-translation-korean-multilingual` (1209 → 500 lines)
2. `moai-baas-vercel-ext` (925 → 500 lines)
3. `moai-security-auth` (915 → 500 lines)
4. `moai-project-documentation` (905 → 500 lines)
5. `moai-core-todowrite-pattern` (893 → 500 lines)
6. `moai-docs-unified` (850 → 500 lines)
7. `moai-baas-firebase-ext` (839 → 500 lines)
8. `moai-lib-shadcn-ui` (837 → 500 lines)
9. `moai-design-systems` (806 → 500 lines)

**Reduction Strategy**:
- Extract examples to `examples.md` (save 100-200 lines)
- Move references to `reference.md` (save 50-100 lines)
- Compress redundant explanations (save 100-150 lines)
- Apply progressive token disclosure (save 50-100 lines)

**Command**:
```bash
# For each oversized skill
cd .claude/skills/[skill-name]

# Create examples.md if needed
touch examples.md

# Create reference.md if needed
touch reference.md

# Extract content and reduce SKILL.md
# Verify line count: wc -l SKILL.md
# Target: <500 lines

# Copy to templates
# Verify synchronization
```

---

### Task 3: Final Synchronization (30 minutes)

**Directories to Sync**:
- Main: `.claude/skills/*/SKILL.md`
- Templates: `src/moai_adk/templates/.claude/skills/*/SKILL.md`

**Verification Commands**:
```bash
# Count skills in both directories
find .claude/skills -name "SKILL.md" | wc -l
find src/moai_adk/templates/.claude/skills -name "SKILL.md" | wc -l

# Compare file sizes
diff -r .claude/skills src/moai_adk/templates/.claude/skills

# Verify identical content
for skill in .claude/skills/*/SKILL.md; do
  template_skill="src/moai_adk/templates/${skill}"
  if ! diff -q "$skill" "$template_skill" > /dev/null; then
    echo "MISMATCH: $skill"
  fi
done
```

---

### Task 4: Quality Validation (15 minutes)

**Automated Checks**:
```bash
# Check all skills have Progressive Disclosure
grep -L "Progressive Disclosure\|Level 1:\|Quick Reference" .claude/skills/*/SKILL.md

# Verify line counts
find .claude/skills -name "SKILL.md" -exec wc -l {} \; | awk '$1 > 500 {print $2, $1}'

# Check YAML frontmatter
for skill in .claude/skills/*/SKILL.md; do
  if ! head -n 1 "$skill" | grep -q "^---$"; then
    echo "MISSING YAML: $skill"
  fi
done
```

---

## Phase 3 Preparation Tasks

### Task 5: Create Library Mapping File (30 minutes)

**Location**: `.moai/config/context7-libraries.yaml`

**Template**:
```yaml
python_libraries:
  fastapi: "/tiangolo/fastapi"
  django: "/django/django"
  pydantic: "/pydantic/pydantic"
  # ... 15+ libraries

javascript_libraries:
  react: "/facebook/react"
  nextjs: "/vercel/next.js"
  typescript: "/microsoft/TypeScript"
  # ... 15+ libraries

go_libraries:
  gin: "/gin-gonic/gin"
  gorm: "/go-gorm/gorm"
  # ... 7+ libraries

rust_crates:
  axum: "/tokio-rs/axum"
  tokio: "/tokio-rs/tokio"
  # ... 6+ libraries
```

---

### Task 6: Begin Context7 Integration (Week 1)

**High-Priority Skills** (20 skills/day):
- Day 1: Language skills (Python, TypeScript, Go, Rust, JavaScript)
- Day 2: Domain skills (Backend, Frontend, Database, Testing)
- Day 3: Framework skills (Next.js, React, FastAPI, Django)
- Day 4: BaaS skills (Vercel, Firebase, Supabase)

**Documentation Section Template**:
```markdown
## Documentation

For up-to-date library documentation:
Skill("moai-context7-lang-integration")

### Context7 Integration

**Supported Libraries**:
- **FastAPI**: `/tiangolo/fastapi`
  - Topic: `routing`, `dependency-injection`, `async`
  - Token budget: 3000-5000

**Example**:
\```python
docs = await get_library_docs(
    context7_library_id="/tiangolo/fastapi",
    topic="routing patterns",
    tokens=3000
)
\```
```

---

## Success Criteria

**Phase 2 Complete When**:
- ✅ 132/132 skills with Progressive Disclosure (100%)
- ✅ 132/132 skills under 500 lines (100%)
- ✅ Both directories synchronized (100%)
- ✅ All quality checks passing

**Phase 3 Ready When**:
- ✅ Library mapping file created
- ✅ Documentation section template validated
- ✅ Integration patterns tested
- ✅ First 20 skills updated

---

## Timeline Summary

**Today (Day 1)**:
- ✅ Phase 2 analysis complete
- ✅ Phase 3 planning complete
- ⏳ 5 skills processing (Task 1)
- ⏳ 9 skills reduction (Task 2)

**Tomorrow (Day 2)**:
- Final synchronization (Task 3)
- Quality validation (Task 4)
- Library mapping creation (Task 5)

**Week 1 (Days 3-7)**:
- Context7 integration rollout (Task 6)
- 20 skills/day = 100 skills/week

**Weeks 2-4**:
- Testing + validation
- Synchronization
- Final deployment

