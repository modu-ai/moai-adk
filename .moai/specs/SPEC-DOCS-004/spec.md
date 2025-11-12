---
id: DOCS-004
version: 0.0.1
status: completed
created: 2025-10-29
updated: 2025-10-29
priority: high
category: docs
labels:
  - documentation
  - readme
  - version-update
  - translation
scope:
  files:
    - README.md
    - README.ko.md
    - README.ja.md
    - README.zh.md
---

# SPEC: README.md Documentation Update to v0.8.2


## Overview

Update README.md and all translated versions (ko, ja, zh) to accurately reflect MoAI-ADK v0.8.2, including:
- Version references: v0.4.11 → v0.8.2
- Missing version history: v0.6.0, v0.6.3, v0.7.0, v0.8.0, v0.8.1, v0.8.2
- Coverage badge re-measurement
- Skills count consistency (standardize to "55+")
- Command rename: `/alfred:9-help` → `/alfred:9-feedback`
- EARS terminology update: "Constraints" → "Unwanted Behaviors"

## TAG BLOCK

```
```

## Environment

**WHEN** MoAI-ADK reaches v0.8.2 milestone,
**THE SYSTEM** shall provide accurate, up-to-date README documentation reflecting all feature improvements from v0.4.11 to v0.8.2.

**Context**:
- Current README shows outdated v0.4.11 references
- 6 versions missing from changelog: v0.6.0, v0.6.3, v0.7.0, v0.8.0, v0.8.1, v0.8.2
- Coverage badge may be outdated
- Skills count inconsistency across document
- Recent EARS terminology change not reflected
- Command rename not documented

## Assumptions

1. **Version Authority**: PyPI package version is the source of truth
2. **Translation Parity**: All language versions (en, ko, ja, zh) must have equivalent updates
3. **Coverage Data**: Codecov or local coverage report available for badge update
4. **Skills Count**: Actual count is 55+ (verified from `.claude/skills/` directory)
5. **EARS Standard**: Official terminology change approved: "Constraints" → "Unwanted Behaviors"

## Requirements

### Ubiquitous Requirements

- **DESCRIPTION**: All version references throughout README.md must be updated to v0.8.2
- **RATIONALE**: Users must see consistent, accurate version information
- **ACCEPTANCE**: Zero occurrences of "v0.4.11" in install commands, badges, or links

- **DESCRIPTION**: All translated READMEs (ko, ja, zh) must reflect identical updates
- **RATIONALE**: Non-English users deserve equal information quality
- **ACCEPTANCE**: All 4 README files show v0.8.2 and complete version history

- **DESCRIPTION**: Standardize all Skills count references to "55+" throughout document
- **RATIONALE**: Eliminates confusion (56, 58, 55 variants found)
- **ACCEPTANCE**: All Skills count references use "55+" notation

### Event-Driven Requirements

- **WHEN** user reads installation instructions
- **THE SYSTEM** shall display `uv tool install moai-adk==0.8.2` and `pip install moai-adk==0.8.2`
- **RATIONALE**: Ensure users install correct version
- **LOCATION**: README.md line 1970

- **WHEN** user views README badges section
- **THE SYSTEM** shall display PyPI package link with "(Latest: v0.8.2)" text
- **RATIONALE**: Accurate package version visibility
- **LOCATION**: README.md line 1992

- **WHEN** user views community links
- **THE SYSTEM** shall display GitHub release link: `https://github.com/modu-ai/moai-adk/releases/tag/v0.8.2`
- **RATIONALE**: Direct users to correct release notes
- **LOCATION**: README.md line 1993

### State-Driven Requirements

- **WHILE** README displays "Latest Updates" section
- **THE SYSTEM** shall include detailed entries for:
  - v0.8.2: EARS terminology update ("Constraints" → "Unwanted Behaviors")
  - v0.8.1: Command rename `/alfred:9-help` → `/alfred:9-feedback`, improved user feedback workflow
  - v0.7.0: Language localization system (5 languages: en, ko, ja, zh, es)
  - v0.6.3: 3-Stage update workflow with 70-80% performance improvement
  - v0.6.0: Major architecture refactor, enhanced SPEC metadata structure
- **RATIONALE**: Complete changelog transparency for users
- **LOCATION**: README.md lines 1960-1968

- **WHILE** README displays badge section
- **THE SYSTEM** shall show current test coverage percentage from latest measurement
- **RATIONALE**: Accurate quality metrics for users
- **LOCATION**: README.md line 10

## Specifications

### SPEC-1: Version Reference Updates

```bash
# OLD
uv tool install moai-adk==0.4.11

# NEW
uv tool install moai-adk==0.8.2
```

```markdown
# OLD
https://pypi.org/project/moai-adk/ (Latest: v0.4.11)

# NEW
https://pypi.org/project/moai-adk/ (Latest: v0.8.2)
```

```markdown
# OLD
https://github.com/modu-ai/moai-adk/releases/tag/v0.4.11

# NEW
https://github.com/modu-ai/moai-adk/releases/tag/v0.8.2
```

### SPEC-2: Version History Expansion

```markdown
| Version     | Key Features                                                                                     | Date       |
| ----------- | ------------------------------------------------------------------------------------------------ | ---------- |
| **v0.8.2**  | 📖 EARS terminology update: "Constraints" → "Unwanted Behaviors" for clarity                     | 2025-10-29 |
| **v0.8.1**  | 🔄 Command rename: `/alfred:9-help` → `/alfred:9-feedback` + User feedback workflow improvements | 2025-10-28 |
| **v0.7.0**  | 🌍 Complete language localization system (English, Korean, Japanese, Chinese, Spanish)           | 2025-10-26 |
| **v0.6.3**  | ⚡ 3-Stage update workflow: 70-80% performance improvement via parallel operations               | 2025-10-25 |
| **v0.6.0**  | 🏗️ Major architecture refactor + Enhanced SPEC metadata structure (7 required + 9 optional)      | 2025-10-24 |
| **v0.5.7**  | 🎯 SPEC → GitHub Issue automation + CodeRabbit integration + Auto PR comments                    | 2025-10-27 |
| **v0.4.11** | ✨ TAG Guard system + CLAUDE.md formatting improvements + Code cleanup                           | 2025-10-23 |
```

### SPEC-3: Skills Count Standardization

- Line 218: "58 Skills" → "55+ Skills"
- Line 1290: "55 Claude Skills" → "55+ Claude Skills" (already correct)
- Line 1978: "58 Skills" → "55+ Skills"
- Line 2002: "56 Claude Skills" → "55+ Claude Skills"
- Line 2019: "55+ Production-Ready Guides" → Keep as-is (correct)

**Standardization Rule**: Use "55+" notation to indicate "55 or more" (accounts for growth)

### SPEC-4: Coverage Badge Update

1. Run local coverage measurement: `pytest --cov=moai_adk --cov-report=term`
2. Extract percentage from coverage report
3. Update badge: `[![Coverage](https://img.shields.io/badge/coverage-XX.XX%25-brightgreen)](https://github.com/modu-ai/moai-adk)`
4. Badge color logic:
   - ≥90%: `brightgreen`
   - 80-89%: `green`
   - 70-79%: `yellowgreen`
   - <70%: `yellow`

### SPEC-5: Translation Updates

1. **README.ko.md** (Korean):
   - Translate version history descriptions to Korean
   - Update version numbers (language-neutral)
   - Maintain Korean technical terminology standards

2. **README.ja.md** (Japanese):
   - Translate version history descriptions to Japanese
   - Update version numbers (language-neutral)
   - Maintain Japanese technical terminology standards

3. **README.zh.md** (Chinese):
   - Translate version history descriptions to Chinese
   - Update version numbers (language-neutral)
   - Maintain Chinese technical terminology standards

**Translation Priorities**:
- Version numbers: NO translation (keep as v0.8.2)
- Feature descriptions: FULL translation
- Command names: NO translation (keep as `/alfred:9-feedback`)
- Technical terms: Use established localized equivalents

## Unwanted Behaviors

### UB-1: Version Inconsistency

- **SCENARIO**: Some references updated to v0.8.2 while others remain v0.4.11
- **IMPACT**: User confusion, incorrect installations
- **PREVENTION**: Global search-replace verification before commit

- **SCENARIO**: English README shows v0.8.2 but translated versions still show v0.4.11
- **IMPACT**: Non-English users receive outdated information
- **PREVENTION**: Parallel update of all 4 README files in same commit

### UB-2: Content Errors

- **SCENARIO**: Version history describes features not actually in that version
- **IMPACT**: Misleading users about capabilities
- **PREVENTION**: Cross-reference with actual git tags and CHANGELOG.md

- **SCENARIO**: Some "56" or "58" references missed during standardization
- **IMPACT**: Document appears unprofessional
- **PREVENTION**: Regex search for all numeric Skills references: `\d{2}\s+(Claude\s+)?Skills`

### UB-3: Technical Errors

- **SCENARIO**: PyPI/Release links point to non-existent v0.8.2 tag before release
- **IMPACT**: 404 errors for users
- **PREVENTION**: Verify GitHub release exists before updating links

- **SCENARIO**: Badge shows outdated coverage percentage
- **IMPACT**: Inaccurate quality metrics
- **PREVENTION**: Re-run coverage measurement immediately before update

## Implementation Notes

### Critical Path
1. ✅ ID duplication check (COMPLETED - No SPEC-DOCS-004 exists)
2. Verify v0.8.2 release exists on GitHub
3. Run coverage measurement
4. Update English README.md (all 6 categories)
5. Update translated READMEs (ko, ja, zh)
6. Validate all links (PyPI, GitHub releases)
7. Commit with proper TAG references

### Validation Checklist
- [ ] All "0.4.11" replaced with "0.8.2"
- [ ] 6 new version entries added to history table
- [ ] Skills count standardized to "55+"
- [ ] Coverage badge reflects latest measurement
- [ ] All 4 README files updated (en, ko, ja, zh)
- [ ] PyPI link verified: https://pypi.org/project/moai-adk/
- [ ] GitHub release verified: https://github.com/modu-ai/moai-adk/releases/tag/v0.8.2
- [ ] No broken links in document
- [ ] EARS terminology updated where applicable

## Traceability

**Parent SPECs**: None (documentation maintenance)
**Related SPECs**:
- SPEC-DOCS-001: Initial README.md creation
- SPEC-DOCS-002: Translation system setup
- SPEC-DOCS-003: Badge automation

**Downstream**:

## Success Criteria

1. **Completeness**: All 6 version updates documented with 3-5 features each
2. **Accuracy**: Zero version inconsistencies across all 4 README files
3. **Validation**: All links return HTTP 200 status
4. **Consistency**: Skills count shows "55+" in all occurrences
5. **Coverage**: Badge reflects actual test coverage within ±1%
6. **Translation**: All 4 language versions show identical information

## History

