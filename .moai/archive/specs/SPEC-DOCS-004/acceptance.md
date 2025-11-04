# Acceptance Criteria: README.md Documentation Update to v0.8.2

**@ACCEPTANCE:DOCS-004**

## Overview

Comprehensive acceptance criteria for verifying README.md updates to v0.8.2, including version references, changelog expansion, Skills count standardization, coverage badge accuracy, and multilingual translation parity.

## TAG BLOCK

```
@ACCEPTANCE:DOCS-004
@SPEC:DOCS-004
@PLAN:DOCS-004
@TEST:DOCS-004-README-VALIDATION
```

## Acceptance Criteria Format

All criteria follow **Given-When-Then** format for clarity and testability.

---

## AC-1: Version Reference Updates

### AC-1.1: Install Command Version

**@TEST:DOCS-004-AC-001-INSTALL**

**GIVEN** README.md contains installation instructions
**WHEN** user reads the install command at line 1970
**THEN** the command shall display:
- `uv tool install moai-adk==0.8.2`
- `pip install moai-adk==0.8.2`
**AND** no references to v0.4.11 shall exist in install commands

**Validation Method**:
```bash
# Should return exactly 1 match showing v0.8.2
grep -n "uv tool install moai-adk==0.8.2" README.md

# Should return 0 matches
grep -n "uv tool install moai-adk==0.4.11" README.md
```

### AC-1.2: PyPI Package Link

**@TEST:DOCS-004-AC-002-PYPI**

**GIVEN** Community & Support section exists in README.md
**WHEN** user views PyPI package information at line 1992
**THEN** the link text shall display:
- "https://pypi.org/project/moai-adk/ (Latest: v0.8.2)"
**AND** the link shall be accessible (HTTP 200)

**Validation Method**:
```bash
# Verify text update
grep -n "Latest: v0.8.2" README.md

# Verify link accessibility
curl -I https://pypi.org/project/moai-adk/ | grep "HTTP/2 200"
```

### AC-1.3: GitHub Release Link

**@TEST:DOCS-004-AC-003-RELEASE**

**GIVEN** Community & Support section exists in README.md
**WHEN** user views Latest Release information at line 1993
**THEN** the link shall point to:
- "https://github.com/modu-ai/moai-adk/releases/tag/v0.8.2"
**AND** the release page shall be accessible (HTTP 200)

**Validation Method**:
```bash
# Verify link update
grep -n "releases/tag/v0.8.2" README.md

# Verify release exists
gh release view v0.8.2 --json name,tagName
```

---

## AC-2: Version History Completeness

### AC-2.1: Six New Version Entries

**@TEST:DOCS-004-AC-004-HISTORY**

**GIVEN** README.md "Latest Updates" section exists (lines 1960-1968)
**WHEN** user views the version history table
**THEN** the table shall contain entries for:
1. v0.8.2 (EARS terminology update)
2. v0.8.1 (Command rename)
3. v0.8.0 (@DOC TAG auto-generation)
4. v0.7.0 (Language localization)
5. v0.6.3 (3-Stage workflow)
6. v0.6.0 (Architecture refactor)
**AND** each entry shall have:
- Version number in bold
- 3-5 feature descriptions
- Date in YYYY-MM-DD format
- Appropriate emoji icon

**Validation Method**:
```bash
# Count version entries in table (should be ≥8 including v0.5.7, v0.4.11)
grep -c "^\| \*\*v0\." README.md

# Verify specific versions exist
grep "v0.8.2" README.md
grep "v0.8.1" README.md
grep "v0.8.0" README.md
grep "v0.7.0" README.md
grep "v0.6.3" README.md
grep "v0.6.0" README.md
```

### AC-2.2: Feature Description Accuracy

**@TEST:DOCS-004-AC-005-FEATURES**

**GIVEN** version history table contains 6 new entries
**WHEN** user reads feature descriptions
**THEN** each version shall describe accurate, implemented features:

| Version | Required Keywords | Validation |
|---------|------------------|------------|
| v0.8.2 | "EARS", "Unwanted Behaviors", "Constraints" | ✅ |
| v0.8.1 | "Command rename", "9-help", "9-feedback" | ✅ |
| v0.8.0 | "@DOC TAG", "auto-generation", "SessionStart" | ✅ |
| v0.7.0 | "language", "localization", "5 languages" or "English, Korean" | ✅ |
| v0.6.3 | "3-Stage", "70-80%", "performance" | ✅ |
| v0.6.0 | "architecture", "SPEC metadata", "7 required" | ✅ |

**Validation Method**:
```bash
# Extract v0.8.2 row and verify keywords
grep "v0.8.2" README.md | grep -q "EARS" && \
grep "v0.8.2" README.md | grep -q "Unwanted Behaviors"

# Extract v0.7.0 row and verify language count
grep "v0.7.0" README.md | grep -qE "(5 languages|English.*Korean.*Japanese)"
```

---

## AC-3: Skills Count Standardization

### AC-3.1: Global "55+" Standardization

**@TEST:DOCS-004-AC-006-SKILLS-COUNT**

**GIVEN** README.md contains multiple references to Skills count
**WHEN** searching for Skills count mentions
**THEN** ALL occurrences shall use "55+" notation
**AND** NO occurrences of "56 Skills", "58 Skills" shall exist

**Known Update Locations**:
- Line 218: `.claude/skills/` directory comment
- Line 1978: Additional Resources table
- Line 2002: Philosophy section

**Validation Method**:
```bash
# Should return 0 matches (no old variants)
grep -nE "\b(56|58)\s+(Claude\s+)?Skills?\b" README.md

# Should return ≥3 matches (standardized notation)
grep -n "55+" README.md | grep -i skills
```

### AC-3.2: Consistent Notation Format

**@TEST:DOCS-004-AC-007-SKILLS-FORMAT**

**GIVEN** Skills count is standardized to "55+"
**WHEN** user encounters Skills references
**THEN** acceptable formats are:
- "55+ Skills" ✅
- "55+ Claude Skills" ✅
- "55+ Production-Ready Guides" ✅
**AND** unacceptable formats are:
- "55 Skills" ❌ (missing "+")
- "56 Claude Skills" ❌ (wrong number)
- "More than 55 Skills" ❌ (inconsistent phrasing)

**Validation Method**:
```bash
# Should return 0 matches (no inconsistent phrasing)
grep -nE "\bmore than \d+ skills\b" README.md -i

# Verify "+" symbol used with 55
grep -n "55+" README.md
```

---

## AC-4: Coverage Badge Accuracy

### AC-4.1: Badge Reflects Current Coverage

**@TEST:DOCS-004-AC-008-COVERAGE-BADGE**

**GIVEN** project test suite has measurable coverage
**WHEN** coverage measurement runs: `pytest --cov=moai_adk --cov-report=term`
**THEN** README.md line 10 badge shall display coverage within ±0.5% of measured value
**AND** badge color shall match coverage level:
- ≥90%: `brightgreen`
- 80-89%: `green`
- 70-79%: `yellowgreen`
- <70%: `yellow`

**Validation Method**:
```bash
# Run coverage measurement
coverage_value=$(pytest --cov=moai_adk --cov-report=term | grep "TOTAL" | awk '{print $4}' | sed 's/%//')

# Extract badge value from README
badge_value=$(grep "badge/coverage-" README.md | sed -E 's/.*coverage-([0-9.]+).*/\1/')

# Compare (allow 0.5% tolerance)
echo "Measured: ${coverage_value}%, Badge: ${badge_value}%"
```

### AC-4.2: Badge Link Accessibility

**@TEST:DOCS-004-AC-009-BADGE-LINK**

**GIVEN** coverage badge is displayed in README.md
**WHEN** user clicks badge link
**THEN** the link shall navigate to valid coverage report
**AND** if using Codecov: link shall return HTTP 200

**Validation Method**:
```bash
# Extract badge link URL
badge_link=$(grep -oP 'https://codecov\.io/gh/modu-ai/moai-adk[^)]*' README.md)

# Test accessibility
curl -I "$badge_link" | grep "HTTP"
```

---

## AC-5: Translation Parity

### AC-5.1: Korean README Update

**@TEST:DOCS-004-AC-010-KOREAN**

**GIVEN** README.ko.md exists as Korean translation
**WHEN** comparing README.ko.md to README.md
**THEN** README.ko.md shall contain:
- Version numbers: v0.8.2 (unchanged from English)
- Install command: `moai-adk==0.8.2`
- 6 new version entries (v0.6.0 - v0.8.2)
- Translated feature descriptions in Korean
- Skills count: "55+" or "55개 이상"
- Same coverage percentage as README.md

**Validation Method**:
```bash
# Verify version references
grep -q "0.8.2" README.ko.md

# Verify 6 new versions exist
grep -c "v0.8.[012]" README.ko.md  # Should return ≥3
grep -c "v0.6.[03]" README.ko.md   # Should return ≥2
grep -q "v0.7.0" README.ko.md      # Should return success
```

### AC-5.2: Japanese README Update

**@TEST:DOCS-004-AC-011-JAPANESE**

**GIVEN** README.ja.md exists as Japanese translation
**WHEN** comparing README.ja.md to README.md
**THEN** README.ja.md shall contain identical updates as Korean README
**AND** feature descriptions shall be in Japanese

**Validation Method**: (Same as AC-5.1, replace `.ko.md` with `.ja.md`)

### AC-5.3: Chinese README Update

**@TEST:DOCS-004-AC-012-CHINESE**

**GIVEN** README.zh.md exists as Chinese translation
**WHEN** comparing README.zh.md to README.md
**THEN** README.zh.md shall contain identical updates as Korean README
**AND** feature descriptions shall be in Chinese

**Validation Method**: (Same as AC-5.1, replace `.ko.md` with `.zh.md`)

### AC-5.4: Translation Consistency Matrix

**@TEST:DOCS-004-AC-013-TRANSLATION-MATRIX**

**GIVEN** all 4 README files have been updated
**WHEN** performing consistency check
**THEN** verification matrix shall show 100% completion:

| Element | README.md | README.ko.md | README.ja.md | README.zh.md |
|---------|-----------|--------------|--------------|--------------|
| Version: v0.8.2 | ✅ | ✅ | ✅ | ✅ |
| 6 new versions | ✅ | ✅ | ✅ | ✅ |
| Skills: 55+ | ✅ | ✅ | ✅ | ✅ |
| Coverage badge | ✅ | ✅ | ✅ | ✅ |
| Install: 0.8.2 | ✅ | ✅ | ✅ | ✅ |
| PyPI: v0.8.2 | ✅ | ✅ | ✅ | ✅ |
| Release: v0.8.2 | ✅ | ✅ | ✅ | ✅ |

**Validation Method**:
```bash
# Automated consistency check script
for file in README.md README.ko.md README.ja.md README.zh.md; do
  echo "Checking $file..."
  grep -q "0.8.2" "$file" && echo "✅ Version OK" || echo "❌ Version missing"
  grep -q "55+" "$file" && echo "✅ Skills OK" || echo "❌ Skills inconsistent"
done
```

---

## AC-6: Link Validation

### AC-6.1: PyPI Package Accessibility

**@TEST:DOCS-004-AC-014-PYPI-ACCESS**

**GIVEN** README.md contains PyPI package link
**WHEN** HTTP GET request is made to https://pypi.org/project/moai-adk/
**THEN** response status shall be HTTP 200
**AND** page shall display v0.8.2 as latest version

**Validation Method**:
```bash
curl -I https://pypi.org/project/moai-adk/ | grep "HTTP/2 200"
```

### AC-6.2: GitHub Release Accessibility

**@TEST:DOCS-004-AC-015-RELEASE-ACCESS**

**GIVEN** README.md contains GitHub release link
**WHEN** HTTP GET request is made to https://github.com/modu-ai/moai-adk/releases/tag/v0.8.2
**THEN** response status shall be HTTP 200
**AND** release page shall display v0.8.2 tag

**Validation Method**:
```bash
gh release view v0.8.2 --json name,tagName,publishedAt
```

### AC-6.3: Coverage Service Accessibility

**@TEST:DOCS-004-AC-016-COVERAGE-ACCESS**

**GIVEN** README.md contains Codecov badge link
**WHEN** HTTP GET request is made to Codecov URL
**THEN** response status shall be HTTP 200 (if service used)
**OR** badge shall use static img.shields.io URL (if self-hosted)

**Validation Method**:
```bash
# Extract coverage badge URL
badge_url=$(grep -oP 'https://[^)]*coverage[^)]*' README.md | head -1)
curl -I "$badge_url" | grep "HTTP"
```

---

## AC-7: Content Quality

### AC-7.1: No Version Inconsistencies

**@TEST:DOCS-004-AC-017-NO-OLD-VERSIONS**

**GIVEN** README.md has been updated to v0.8.2
**WHEN** searching for old version references
**THEN** zero occurrences of "v0.4.11" shall exist in:
- Install commands
- Badge links
- PyPI references
- GitHub release references

**Validation Method**:
```bash
# Should return exit code 1 (no matches found)
grep -r "0\.4\.11" README*.md
echo $?  # Expected: 1 (no matches)
```

### AC-7.2: Emoji Icons Consistency

**@TEST:DOCS-004-AC-018-EMOJI-CONSISTENCY**

**GIVEN** version history table contains entries with emoji icons
**WHEN** reviewing emoji usage
**THEN** each version entry shall have exactly 1 emoji icon
**AND** emojis shall be semantically appropriate:
- 📖 for documentation updates
- 🔄 for workflow changes
- 🏷️ for tagging/automation
- 🌍 for internationalization
- ⚡ for performance improvements
- 🏗️ for architecture changes

**Validation Method**:
```bash
# Count emoji icons in version table (should match version count)
grep -oP "^\| \*\*v0\.\d+\.\d+\*\*  \| [^\|]*" README.md | wc -l
```

### AC-7.3: Date Format Consistency

**@TEST:DOCS-004-AC-019-DATE-FORMAT**

**GIVEN** version history table contains date column
**WHEN** reviewing date values
**THEN** all dates shall follow YYYY-MM-DD format
**AND** dates shall be chronologically consistent (newer versions have newer dates)

**Validation Method**:
```bash
# Extract dates from version table
grep "^\| \*\*v0\." README.md | awk -F'|' '{print $4}' | sed 's/ //g'

# Verify format: YYYY-MM-DD
grep "^\| \*\*v0\." README.md | grep -oP "\d{4}-\d{2}-\d{2}" | wc -l
```

---

## AC-8: EARS Terminology Update

### AC-8.1: "Constraints" Replaced with "Unwanted Behaviors"

**@TEST:DOCS-004-AC-020-EARS-TERMINOLOGY**

**GIVEN** EARS methodology section exists in README.md (if applicable)
**WHEN** searching for EARS-related terminology
**THEN** "Unwanted Behaviors" shall be used instead of "Constraints"
**AND** v0.8.2 changelog entry shall mention this terminology change

**Validation Method**:
```bash
# Verify v0.8.2 mentions EARS update
grep "v0.8.2" README.md | grep -qi "EARS\|Unwanted Behaviors"

# Check for outdated "Constraints" in EARS context (should be minimal or zero)
grep -n "Constraints" README.md | grep -i "ears"
```

---

## AC-9: Git Workflow Compliance

### AC-9.1: Proper Commit Message

**@TEST:DOCS-004-AC-021-COMMIT-MESSAGE**

**GIVEN** README updates have been staged for commit
**WHEN** commit is created
**THEN** commit message shall include:
- Conventional commit format: `docs(readme): ...`
- Description of changes (version update, changelog expansion, etc.)
- TAG references: `@SPEC:DOCS-004`, `@DOC:README-VERSION-UPDATE-001`
- MoAI-ADK signature: "🤖 Generated with Claude Code"
- Co-author line: "Co-Authored-By: Claude <noreply@anthropic.com>"

**Validation Method**:
```bash
# After commit, check message
git log -1 --pretty=%B | grep "@SPEC:DOCS-004"
git log -1 --pretty=%B | grep "Co-Authored-By: Claude"
```

### AC-9.2: Correct Files Staged

**@TEST:DOCS-004-AC-022-FILES-STAGED**

**GIVEN** documentation update is complete
**WHEN** staging files for commit
**THEN** exactly these files shall be staged:
- README.md
- README.ko.md
- README.ja.md
- README.zh.md
**AND** no other unrelated files shall be included

**Validation Method**:
```bash
git diff --name-only --cached | sort
# Expected output:
# README.ja.md
# README.ko.md
# README.md
# README.zh.md
```

---

## AC-10: Regression Prevention

### AC-10.1: No Broken Markdown Formatting

**@TEST:DOCS-004-AC-023-MARKDOWN-VALID**

**GIVEN** README.md has been updated
**WHEN** rendering README in markdown viewer
**THEN** all tables shall render correctly
**AND** all links shall be clickable
**AND** all code blocks shall have proper syntax highlighting

**Validation Method**:
```bash
# Use markdown linter (if available)
markdownlint README.md

# Or GitHub API to preview rendering
gh api repos/modu-ai/moai-adk/readme --jq .content | base64 -d > /tmp/readme_preview.md
```

### AC-10.2: No Accidental Content Deletion

**@TEST:DOCS-004-AC-024-NO-DELETION**

**GIVEN** README.md before and after update
**WHEN** comparing line counts
**THEN** new README.md shall have MORE lines (due to changelog expansion)
**AND** no critical sections shall be missing:
- Installation instructions
- Quick start guide
- Core features
- Architecture overview

**Validation Method**:
```bash
# Compare line counts (after should be > before)
git show HEAD:README.md | wc -l  # Before
wc -l README.md                  # After

# Verify critical sections exist
grep -q "## Installation" README.md
grep -q "## Quick Start" README.md
grep -q "## Architecture" README.md
```

---

## Test Execution Summary

### Automated Test Suite

**Test Script**: `.moai/tests/test_readme_update_docs_004.sh`

```bash
#!/bin/bash
# Automated acceptance test suite for SPEC-DOCS-004

echo "🧪 Running SPEC-DOCS-004 Acceptance Tests..."

# AC-1: Version references
echo "[AC-1.1] Install command version..."
grep -q "moai-adk==0.8.2" README.md && echo "✅ PASS" || echo "❌ FAIL"

echo "[AC-1.2] PyPI link version..."
grep -q "Latest: v0.8.2" README.md && echo "✅ PASS" || echo "❌ FAIL"

echo "[AC-1.3] Release link version..."
grep -q "releases/tag/v0.8.2" README.md && echo "✅ PASS" || echo "❌ FAIL"

# AC-2: Version history
echo "[AC-2.1] Six new versions present..."
for version in "v0.8.2" "v0.8.1" "v0.8.0" "v0.7.0" "v0.6.3" "v0.6.0"; do
  grep -q "$version" README.md || { echo "❌ FAIL: Missing $version"; exit 1; }
done
echo "✅ PASS"

# AC-3: Skills count
echo "[AC-3.1] Skills count standardized..."
if grep -qE "\b(56|58)\s+Skills\b" README.md; then
  echo "❌ FAIL: Found non-standard Skills count"
  exit 1
else
  echo "✅ PASS"
fi

# AC-5: Translation parity
echo "[AC-5.4] Translation consistency..."
for file in README.ko.md README.ja.md README.zh.md; do
  grep -q "0.8.2" "$file" || { echo "❌ FAIL: $file missing v0.8.2"; exit 1; }
done
echo "✅ PASS"

# AC-7: No old versions
echo "[AC-7.1] No v0.4.11 references..."
if grep -q "0\.4\.11" README*.md; then
  echo "❌ FAIL: Found v0.4.11 references"
  exit 1
else
  echo "✅ PASS"
fi

echo ""
echo "✅ All SPEC-DOCS-004 acceptance tests passed!"
```

### Manual Verification Checklist

**Human Review Required**:
- [ ] Version feature descriptions are accurate and meaningful
- [ ] Translated content reads naturally (native speaker review)
- [ ] Emoji icons are semantically appropriate
- [ ] Links open correctly in browser
- [ ] Coverage badge displays correctly in GitHub
- [ ] README renders properly on GitHub/PyPI

---

## Definition of Done

**SPEC-DOCS-004 is COMPLETE when**:
1. ✅ All version references updated to v0.8.2
2. ✅ 6 new versions documented in changelog with 3-5 features each
3. ✅ Skills count standardized to "55+" across all occurrences
4. ✅ Coverage badge reflects latest measurement (±0.5% accuracy)
5. ✅ All 4 README files (en, ko, ja, zh) updated identically
6. ✅ All links validated and accessible (HTTP 200)
7. ✅ Zero v0.4.11 references remain
8. ✅ Automated test suite passes 100%
9. ✅ Git commit includes proper TAG references
10. ✅ No markdown formatting regressions

---

## Traceability

**Parent SPEC**: @SPEC:DOCS-004
**Implementation Plan**: @PLAN:DOCS-004
**Test Coverage**: @TEST:DOCS-004-README-VALIDATION

**Related Tests**:
- @TEST:DOCS-004-AC-001-INSTALL - Install command version
- @TEST:DOCS-004-AC-004-HISTORY - Version history completeness
- @TEST:DOCS-004-AC-006-SKILLS-COUNT - Skills count standardization
- @TEST:DOCS-004-AC-013-TRANSLATION-MATRIX - Translation parity

---

**Last Updated**: 2025-10-29
**Author**: @GOOS
