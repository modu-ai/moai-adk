---
name: moai-alfred-reporting
version: 1.0.0
created: 2025-11-05
updated: 2025-11-05
status: active
description: Report generation standards, output formatting rules, and sub-agent report examples
keywords: ['reporting', 'formatting', 'documentation', 'output', 'style']
allowed-tools:
  - Read
  - Bash
---

# Alfred Reporting Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-reporting |
| **Version** | 1.0.0 (2025-11-05) |
| **Allowed tools** | Read, Bash |
| **Auto-load** | On demand during report generation |
| **Tier** | Alfred |

---

## What It Does

Establishes comprehensive standards for report generation, output formatting, and documentation structure within Alfred's workflow. Ensures consistent, professional communication across all user interactions and internal documentation.

## Core Principle: Output Format Distinction

**CRITICAL RULE**: Always distinguish between screen output (user-facing) and internal documents (files).

### Screen Output to Users (Plain Text)
- **Format**: Plain text with NO markdown syntax
- **Purpose**: Direct chat responses, real-time updates
- **Style**: Clear, concise, immediate readability

### Internal Documents (Markdown Format)
- **Format**: Structured markdown with proper syntax
- **Purpose**: Files in `.moai/docs/`, `.moai/reports/`, `.moai/analysis/`
- **Style**: Professional documentation with structure

## Screen Output Standards

### Plain Text Format

When responding directly to users in chat:

```
Detected Merge Conflict:

Root Cause:
- Commit c054777b removed language detection from develop
- Merge commit e18c7f98 re-introduced the line

Impact Range:
- .claude/hooks/alfred/shared/handlers/session.py
- src/moai_adk/templates/.claude/hooks/alfred/shared/handlers/session.py

Proposed Actions:
- Remove detect_language() import and call
- Delete language display line
- Synchronize both files
```

**Key Characteristics**:
- NO markdown headers (`##`, `###`)
- NO markdown tables (`| column | value |`)
- NO markdown formatting (`**bold**`, `*italic*`)
- Simple text with clear indentation
- Human-readable hierarchy through spacing

### When to Use Plain Text

**Always use plain text for:**
- Direct responses to user questions
- Real-time progress updates
- Error messages and debugging output
- Quick status reports
- Analysis summaries in chat

## Internal Document Standards

### Markdown Format

When creating files in `.moai/` directories:

```markdown
## 🎊 Task Completion Report

### Implementation Results
- ✅ Feature A implementation completed
- ✅ Tests written and passing (47/47)
- ✅ Documentation synchronized

### Quality Metrics
| Item | Result | Status |
|------|--------|--------|
| Test Coverage | 95% | ✅ |
| Linting | 0 issues | ✅ |
| Type Checking | Passed | ✅ |

### Next Steps
1. Run `/alfred:3-sync` for final synchronization
2. Create and review Pull Request
3. Merge to main branch

### Important Notes
⚠️ Database migration required for production deployment
```

**Key Characteristics**:
- Proper markdown headings (`##`, `###`)
- Structured tables for data presentation
- Emojis for visual status indicators (✅, ❌, ⚠️, 🎊, 📊)
- Lists and bullet points for organization
- Clear section separation

### Document Locations

| Document Type | Location | File Pattern |
|---------------|----------|--------------|
| Implementation Guides | `.moai/docs/` | `implementation-{SPEC}.md` |
| Analysis Reports | `.moai/analysis/` | `{topic}-analysis.md` |
| Sync Reports | `.moai/reports/` | `sync-report-{date}.md` |
| Exploration Results | `.moai/docs/` | `exploration-{topic}.md` |
| Strategic Planning | `.moai/docs/` | `strategy-{topic}.md` |

## Report Writing Guidelines

### 1. Structured Sections

```markdown
## 🎯 Key Achievements
- Core accomplishments completed
- Major milestones reached

## 📊 Statistics Summary
| Metric | Value | Status |
|--------|-------|--------|
| Performance | +25% | ✅ |
| Quality | 95% | ✅ |

## ⚠️ Important Notes
- Critical information user needs to know
- Warnings or considerations

## 🚀 Next Steps
1. Recommended immediate action
2. Follow-up tasks
3. Long-term considerations
```

### 2. Length Management

**Short Reports (<500 characters)**:
- Output in single response
- Lead with summary
- Include key metrics

**Long Reports (>500 characters)**:
- Split into logical sections
- Lead with executive summary
- Follow with detailed analysis
- Use emojis for visual organization

### 3. Language and Tone

**User's Language**:
- Explanations and guidance in user's `conversation_language`
- Clear, accessible language
- Cultural context awareness

**Technical Terms**:
- Keep technical terms in English
- Provide explanations in user's language
- Use consistent terminology

## Sub-Agent Report Templates

### spec-builder Report Template

```markdown
## 📋 SPEC Creation Complete

### Generated Documents
- ✅ `.moai/specs/SPEC-XXX-001/spec.md` - Requirements specification
- ✅ `.moai/specs/SPEC-XXX-001/plan.md` - Implementation plan
- ✅ `.moai/specs/SPEC-XXX-001/acceptance.md` - Acceptance criteria

### EARS Validation Results
- ✅ All requirements follow EARS format
- ✅ @TAG chain created and linked
- ✅ Acceptance criteria are measurable
- ✅ Scope boundaries clearly defined

### Quality Metrics
| Aspect | Status | Details |
|--------|--------|---------|
| Clarity | ✅ | Requirements are unambiguous |
| Completeness | ✅ | All user needs captured |
| Testability | ✅ | Each requirement can be tested |
| Feasibility | ✅ | Implementation is achievable |

### Next Steps
1. Review SPEC with stakeholders
2. Run `/alfred:2-run SPEC-XXX-001` to begin implementation
3. Schedule acceptance criteria review
```

### tdd-implementer Report Template

```markdown
## 🚀 TDD Implementation Complete

### Implementation Files
- ✅ `src/feature.py` - Core implementation (245 lines)
- ✅ `tests/test_feature.py` - Comprehensive test suite (156 lines)
- ✅ `docs/api.md` - API documentation updated

### TDD Phases Results
| Phase | Status | Tests | Description |
|-------|--------|-------|-------------|
| RED | ✅ | 5/5 failing | Failure confirmed for all requirements |
| GREEN | ✅ | 5/5 passing | Implementation successful |
| REFACTOR | ✅ | 5/5 passing | Code optimized, tests maintained |

### Quality Metrics
- **Test Coverage**: 95% (47/49 lines covered)
- **Code Quality**: 0 linting issues
- **Performance**: 15% improvement over baseline
- **Security**: No vulnerabilities detected

### Implementation Highlights
- Used dependency injection for testability
- Implemented comprehensive error handling
- Added logging for production monitoring
- Created extensible architecture for future features

### Next Steps
1. Run integration tests with related systems
2. Schedule code review with team
3. Prepare documentation updates
4. Run `/alfred:3-sync` for final synchronization
```

### doc-syncer Report Template

```markdown
## 📚 Documentation Sync Complete

### Updated Documents
- ✅ `README.md` - Usage examples and installation guide
- ✅ `.moai/docs/architecture.md` - System architecture updated
- ✅ `CHANGELOG.md` - v0.8.0 entries added
- ✅ `docs/api/` - API documentation refreshed

### @TAG Verification Results
- ✅ SPEC → CODE connection verified (all links valid)
- ✅ CODE → TEST connection verified (test coverage complete)
- ✅ TEST → DOC connection verified (documentation current)

### Documentation Quality
| Document | Status | Updates | Review Required |
|----------|--------|---------|-----------------|
| README.md | ✅ | Installation guide added | No |
| API Docs | ✅ | New endpoints documented | Yes |
| Architecture | ✅ | New components added | No |
| CHANGELOG | ✅ | Release notes current | No |

### Synchronization Metrics
- **Files Updated**: 12 documents
- **Links Verified**: 47 @TAG connections
- **Content Added**: 2,341 words
- **Accuracy Score**: 98% (minor outdated items found)

### Improvements Made
- Added code examples to README
- Updated architecture diagrams
- Enhanced API documentation with examples
- Standardized formatting across all docs

### Next Steps
1. Review documentation updates with team
2. Publish updated documentation
3. Schedule next documentation sync (recommended: weekly)
```

## Report Generation Triggers

### Automatic Report Generation

**Always generate reports for:**

1. **Command Completion** (Alfred commands)
   - `/alfred:0-project` complete → Project setup report
   - `/alfred:1-plan` complete → SPEC creation report
   - `/alfred:2-run` complete → Implementation report
   - `/alfred:3-sync` complete → Synchronization report

2. **Sub-agent Task Completion**
   - spec-builder: SPEC creation done
   - tdd-implementer: Implementation done
   - doc-syncer: Documentation sync done
   - tag-agent: TAG validation done

3. **Quality Verification Complete**
   - TRUST 5 verification passed
   - Test execution complete
   - Linting/type checking passed

4. **Git Operations Complete**
   - After commit creation
   - After PR creation
   - After merge completion

### Conditional Report Generation

**Generate reports when:**

1. **User explicitly requests**
   - "Create a report for this work"
   - "Generate analysis document"
   - "Write implementation guide"

2. **Significant milestones**
   - Major feature completion
   - Architecture changes
   - Performance improvements

3. **Quality issues discovered**
   - Test failures requiring investigation
   - Performance regressions
   - Security vulnerabilities

## Report Quality Standards

### Content Requirements

**Every report must include:**
1. **Clear Title**: Descriptive and searchable
2. **Executive Summary**: Key findings and outcomes
3. **Detailed Sections**: Structured information presentation
4. **Metrics and Data**: Quantifiable results where applicable
5. **Next Steps**: Actionable recommendations

**Formatting Requirements:**
1. **Consistent Structure**: Use standardized templates
2. **Visual Organization**: Emojis, tables, lists for clarity
3. **Language Consistency**: User's language for explanations
4. **Technical Accuracy**: Verify all technical details

### Quality Checklist

```markdown
## Report Quality Checklist

### Content Quality
- [ ] Title is descriptive and accurate
- [ ] Executive summary captures key points
- [ ] All sections are complete and relevant
- [ ] Data and metrics are accurate
- [ ] Next steps are actionable

### Formatting Standards
- [ ] Markdown syntax is correct
- [ ] Tables are properly formatted
- [ ] Emojis used appropriately for status
- [ ] Section hierarchy is logical
- [ ] Language is consistent (user's language)

### Technical Accuracy
- [ ] @TAG references are correct
- [ ] File paths are accurate
- [ ] Technical details are verified
- [ ] Code examples are tested
- [ ] Links are functional
```

## Integration with Alfred Workflow

### 4-Step Workflow Integration

- **Step 1**: Intent understanding determines if report needed
- **Step 2**: Planning includes report deliverables
- **Step 3**: Execution generates report data
- **Step 4**: Report creation and dissemination

### Tool Usage Guidelines

**Appropriate Tool Usage:**
- **Read tool**: For reading file content and analysis
- **Bash tool**: For system operations and file management
- **Write tool**: For creating reports and documentation

**Prohibited Patterns:**
- **Bash for file content**: Use `cat file.txt` → Use Read tool instead
- **Complex file wrapping**: Avoid heredoc for report generation
- **Markdown in chat**: Use plain text for user responses

## References

- Skill("moai-alfred-doc-management"): Document placement rules
- Skill("moai-alfred-workflow"): 4-Step workflow logic
- Skill("moai-alfred-personas"): Communication style adaptation
