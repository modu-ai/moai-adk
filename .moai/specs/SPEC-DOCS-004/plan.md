# Implementation Plan: README.md Documentation Update to v0.8.2

**@PLAN:DOCS-004**

## Overview

Systematic update of README.md and all translated versions to reflect MoAI-ADK v0.8.2, including version references, changelog expansion, Skills count standardization, coverage badge refresh, and multilingual translation updates.

## TAG BLOCK

```
@PLAN:DOCS-004
@SPEC:DOCS-004
@DOC:README-UPDATE-PLAN-001
```

## Phase 1: Pre-Update Validation

**Priority**: Primary (Blocking)

### Task 1.1: Verify Release Existence
- **Action**: Confirm v0.8.2 tag exists on GitHub
- **Command**: `gh release view v0.8.2` or check https://github.com/modu-ai/moai-adk/releases/tag/v0.8.2
- **Exit Criteria**: Release page accessible, tag confirmed

### Task 1.2: Measure Test Coverage
- **Action**: Run comprehensive coverage measurement
- **Commands**:
  ```bash
  pytest --cov=moai_adk --cov-report=term --cov-report=html
  # Extract percentage from output
  ```
- **Exit Criteria**: Coverage percentage obtained (target: ≥85%)

### Task 1.3: Count Skills Files
- **Action**: Verify actual Skills count
- **Command**: `ls .claude/skills/moai-*.md | wc -l`
- **Exit Criteria**: Confirm "55+" notation is accurate

### Task 1.4: Backup Current READMEs
- **Action**: Create backup of all README files before modification
- **Files**:
  - README.md
  - README.ko.md
  - README.ja.md
  - README.zh.md
- **Exit Criteria**: Backup copies stored safely

## Phase 2: English README.md Updates

**Priority**: Primary (Core Update)

### Task 2.1: Version Reference Updates

**Files**: `README.md`

#### Update Locations:
1. **Line 1970**: Install command
   ```markdown
   # BEFORE
   > 📦 **Install Now**: `uv tool install moai-adk==0.4.11` or `pip install moai-adk==0.4.11`

   # AFTER
   > 📦 **Install Now**: `uv tool install moai-adk==0.8.2` or `pip install moai-adk==0.8.2`
   ```

2. **Line 1992**: PyPI package link
   ```markdown
   # BEFORE
   | **PyPI Package**         | https://pypi.org/project/moai-adk/ (Latest: v0.4.11)     |

   # AFTER
   | **PyPI Package**         | https://pypi.org/project/moai-adk/ (Latest: v0.8.2)     |
   ```

3. **Line 1993**: GitHub release link
   ```markdown
   # BEFORE
   | **Latest Release**       | https://github.com/modu-ai/moai-adk/releases/tag/v0.4.11 |

   # AFTER
   | **Latest Release**       | https://github.com/modu-ai/moai-adk/releases/tag/v0.8.2 |
   ```

### Task 2.2: Version History Expansion

**Location**: Lines 1960-1968 ("Latest Updates" section)

**Action**: Expand table to include 6 missing versions

```markdown
| Version     | Key Features                                                                                     | Date       |
| ----------- | ------------------------------------------------------------------------------------------------ | ---------- |
| **v0.8.2**  | 📖 EARS terminology update: "Constraints" → "Unwanted Behaviors" for clarity                     | 2025-10-29 |
| **v0.8.1**  | 🔄 Command rename: `/alfred:9-help` → `/alfred:9-feedback` + User feedback workflow improvements | 2025-10-28 |
| **v0.8.0**  | 🏷️ @DOC TAG auto-generation system + SessionStart version check enhancement                      | 2025-10-27 |
| **v0.7.0**  | 🌍 Complete language localization system (English, Korean, Japanese, Chinese, Spanish)           | 2025-10-26 |
| **v0.6.3**  | ⚡ 3-Stage update workflow: 70-80% performance improvement via parallel operations               | 2025-10-25 |
| **v0.6.0**  | 🏗️ Major architecture refactor + Enhanced SPEC metadata structure (7 required + 9 optional)      | 2025-10-24 |
| **v0.5.7**  | 🎯 SPEC → GitHub Issue automation + CodeRabbit integration + Auto PR comments                    | 2025-10-27 |
| **v0.4.11** | ✨ TAG Guard system + CLAUDE.md formatting improvements + Code cleanup                           | 2025-10-23 |
```

**Feature Descriptions Rationale**:
- **v0.8.2**: Focus on EARS terminology standardization (user-facing clarity improvement)
- **v0.8.1**: Highlight command rename and workflow UX improvement
- **v0.8.0**: Emphasize automation features (@DOC TAG generation, version checking)
- **v0.7.0**: Showcase internationalization milestone (5 languages)
- **v0.6.3**: Performance improvement metrics (quantified 70-80% gain)
- **v0.6.0**: Architecture foundation changes (metadata structure overhaul)

### Task 2.3: Skills Count Standardization

**Search Pattern**: `\b(55|56|58)\s+(Claude\s+)?Skills?\b`

**Update Locations**:
1. **Line 218**: `.claude/skills/` comment
   ```markdown
   # BEFORE
   │   ├── skills/                     # 58 Claude Skills

   # AFTER
   │   ├── skills/                     # 55+ Claude Skills
   ```

2. **Line 1978**: Additional Resources table
   ```markdown
   # BEFORE
   | Skills detailed structure | `.claude/skills/` directory (58 Skills)                         |

   # AFTER
   | Skills detailed structure | `.claude/skills/` directory (55+ Skills)                         |
   ```

3. **Line 2002**: Philosophy section
   ```markdown
   # BEFORE
   MoAI-ADK is not simply a code generation tool. Alfred SuperAgent with its 19-member team and 56 Claude Skills together guarantee:

   # AFTER
   MoAI-ADK is not simply a code generation tool. Alfred SuperAgent with its 19-member team and 55+ Claude Skills together guarantee:
   ```

**Note**: Line 1290 already shows "55 Claude Skills" (acceptable variant)
**Note**: Line 2019 already shows "55+ Production-Ready Guides" (correct)

### Task 2.4: Coverage Badge Update

**Location**: Line 10

**Action**: Update badge with current coverage percentage

```markdown
# BEFORE
[![Coverage](https://img.shields.io/badge/coverage-87.84%25-brightgreen)](https://github.com/modu-ai/moai-adk)

# AFTER (example with new measurement: 88.42%)
[![Coverage](https://img.shields.io/badge/coverage-88.42%25-brightgreen)](https://github.com/modu-ai/moai-adk)
```

**Badge Color Decision**:
- ≥90%: `brightgreen`
- 80-89%: `green`
- 70-79%: `yellowgreen`
- <70%: `yellow`

### Task 2.5: EARS Terminology Update

**Search Pattern**: `Constraints` (in EARS context)

**Action**: Replace with "Unwanted Behaviors" where applicable
- Check sections discussing EARS methodology
- Update any SPEC examples showing old terminology
- Verify consistency with `.moai/memory/ears-template.md` (if exists)

## Phase 3: Translated README Updates

**Priority**: Primary (Parallel with Phase 2)

### Task 3.1: Korean README Update (README.ko.md)

**Translated Content**:
```markdown
| 버전     | 주요 기능                                                                                     | 날짜       |
| ----------- | ------------------------------------------------------------------------------------------------ | ---------- |
| **v0.8.2**  | 📖 EARS 용어 업데이트: "Constraints" → "Unwanted Behaviors" (명확성 개선)                     | 2025-10-29 |
| **v0.8.1**  | 🔄 명령어 변경: `/alfred:9-help` → `/alfred:9-feedback` + 사용자 피드백 워크플로우 개선 | 2025-10-28 |
| **v0.8.0**  | 🏷️ @DOC TAG 자동 생성 시스템 + SessionStart 버전 체크 강화                      | 2025-10-27 |
| **v0.7.0**  | 🌍 완전한 언어 지역화 시스템 (영어, 한국어, 일본어, 중국어, 스페인어)           | 2025-10-26 |
| **v0.6.3**  | ⚡ 3단계 업데이트 워크플로우: 병렬 작업을 통한 70-80% 성능 개선               | 2025-10-25 |
| **v0.6.0**  | 🏗️ 주요 아키텍처 리팩터링 + SPEC 메타데이터 구조 개선 (필수 7개 + 선택 9개)      | 2025-10-24 |
```

**Update Actions**:
- Version numbers: Keep as-is (v0.8.2)
- Install commands: Update to `0.8.2`
- PyPI/Release links: Update to v0.8.2
- Skills count: "55개 이상의 스킬" or "55+ 스킬"
- Coverage badge: Same percentage as English

### Task 3.2: Japanese README Update (README.ja.md)

**Translated Content**:
```markdown
| バージョン     | 主な機能                                                                                     | 日付       |
| ----------- | ------------------------------------------------------------------------------------------------ | ---------- |
| **v0.8.2**  | 📖 EARS用語更新：「Constraints」→「Unwanted Behaviors」（明確性向上）                     | 2025-10-29 |
| **v0.8.1**  | 🔄 コマンド名変更：`/alfred:9-help` → `/alfred:9-feedback` + ユーザーフィードバックワークフロー改善 | 2025-10-28 |
| **v0.8.0**  | 🏷️ @DOC TAG自動生成システム + SessionStartバージョンチェック強化                      | 2025-10-27 |
| **v0.7.0**  | 🌍 完全な言語ローカライゼーションシステム（英語、韓国語、日本語、中国語、スペイン語）           | 2025-10-26 |
| **v0.6.3**  | ⚡ 3段階更新ワークフロー：並列操作による70-80%のパフォーマンス向上               | 2025-10-25 |
| **v0.6.0**  | 🏗️ 主要アーキテクチャリファクタリング + SPEC メタデータ構造改善（必須7個 + オプション9個）      | 2025-10-24 |
```

**Update Actions**: (Same as Korean README)

### Task 3.3: Chinese README Update (README.zh.md)

**Translated Content**:
```markdown
| 版本     | 主要功能                                                                                     | 日期       |
| ----------- | ------------------------------------------------------------------------------------------------ | ---------- |
| **v0.8.2**  | 📖 EARS术语更新："Constraints" → "Unwanted Behaviors"（提高清晰度）                     | 2025-10-29 |
| **v0.8.1**  | 🔄 命令重命名：`/alfred:9-help` → `/alfred:9-feedback` + 用户反馈工作流程改进 | 2025-10-28 |
| **v0.8.0**  | 🏷️ @DOC TAG自动生成系统 + SessionStart版本检查增强                      | 2025-10-27 |
| **v0.7.0**  | 🌍 完整的语言本地化系统（英语、韩语、日语、中文、西班牙语）           | 2025-10-26 |
| **v0.6.3**  | ⚡ 3阶段更新工作流程：通过并行操作实现70-80%的性能提升               | 2025-10-25 |
| **v0.6.0**  | 🏗️ 主要架构重构 + SPEC元数据结构增强（7个必需 + 9个可选）      | 2025-10-24 |
```

**Update Actions**: (Same as Korean README)

## Phase 4: Validation & Quality Assurance

**Priority**: Secondary (Post-Update)

### Task 4.1: Link Validation

**Action**: Test all updated links

**URLs to Verify**:
```bash
# PyPI package page
curl -I https://pypi.org/project/moai-adk/

# GitHub release page (v0.8.2)
curl -I https://github.com/modu-ai/moai-adk/releases/tag/v0.8.2

# Coverage badge (if using external service)
curl -I https://codecov.io/gh/modu-ai/moai-adk
```

**Expected**: All return HTTP 200 status

### Task 4.2: Version Consistency Check

**Action**: Verify no lingering v0.4.11 references

**Commands**:
```bash
# Search all README files for old version
grep -n "0\.4\.11" README*.md

# Should return: no matches (exit 1)
```

### Task 4.3: Skills Count Consistency Check

**Action**: Verify "55+" standardization complete

**Commands**:
```bash
# Find any variant Skills counts
grep -nE "\b(56|58)\s+(Claude\s+)?Skills?\b" README*.md

# Should return: no matches (exit 1)
```

### Task 4.4: Translation Parity Check

**Action**: Confirm all 4 README files have equivalent updates

**Verification Matrix**:
| Element | README.md | README.ko.md | README.ja.md | README.zh.md |
|---------|-----------|--------------|--------------|--------------|
| Version: v0.8.2 | ✅ | ✅ | ✅ | ✅ |
| 6 new versions | ✅ | ✅ | ✅ | ✅ |
| Skills: 55+ | ✅ | ✅ | ✅ | ✅ |
| Coverage badge | ✅ | ✅ | ✅ | ✅ |

## Phase 5: Git Workflow

**Priority**: Final (Completion)

### Task 5.1: Stage Changes

**Commands**:
```bash
git add README.md README.ko.md README.ja.md README.zh.md
git status
```

### Task 5.2: Commit with TAG References

**Commit Message**:
```
docs(readme): Update documentation to v0.8.2

- Update all version references: v0.4.11 → v0.8.2
- Add 6 missing versions to changelog (v0.6.0 - v0.8.2)
- Standardize Skills count to "55+" notation
- Update coverage badge with latest measurement
- Translate updates to Korean, Japanese, Chinese

@SPEC:DOCS-004
@DOC:README-VERSION-UPDATE-001
@TEST:DOCS-004-README-VALIDATION

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

### Task 5.3: Push to Remote (if team mode)

**Commands**:
```bash
git push origin feature/SPEC-DOCS-004
```

## Technical Approach

### File Modification Strategy

**Option A: Manual Edit (Precise Control)**
- Use Edit tool for targeted line-by-line updates
- Pros: Surgical precision, no accidental changes
- Cons: More tool calls, slower

**Option B: Regex Search-Replace (Batch Updates)**
- Use sed or similar for pattern-based updates
- Pros: Fast, consistent
- Cons: Risk of over-replacement

**Recommended**: Hybrid approach
- Use Edit tool for critical sections (version history table)
- Use search-replace for repetitive updates (version numbers)

### Coverage Measurement Integration

**Process**:
1. Run pytest with coverage plugin
2. Parse output: `grep "TOTAL" coverage_report.txt`
3. Extract percentage: `awk '{print $4}'`
4. Update badge URL with new percentage

**Automation Potential**:
- Script: `.moai/scripts/update_coverage_badge.sh`
- Hook: Pre-commit hook to auto-update badge

### Translation Workflow

**Consistency Mechanisms**:
1. **Template Alignment**: Ensure table structures identical across languages
2. **Version Numbers**: Never translate (keep as v0.8.2)
3. **Command Names**: Never translate (keep as `/alfred:9-feedback`)
4. **Technical Terms**: Use established localized equivalents (maintain glossary)

## Risks & Mitigations

### Risk 1: Release Tag Not Yet Created
- **Impact**: Broken GitHub release link (404 error)
- **Likelihood**: Medium (if updating docs before release)
- **Mitigation**: Verify `gh release view v0.8.2` succeeds before updating links
- **Fallback**: Use draft release link or delay README update

### Risk 2: Coverage Measurement Fails
- **Impact**: Outdated or inaccurate badge
- **Likelihood**: Low
- **Mitigation**: Run multiple measurement attempts, use cached value if necessary
- **Fallback**: Keep existing badge value with note "pending update"

### Risk 3: Translation Inconsistency
- **Impact**: Non-English users see outdated information
- **Likelihood**: Medium (manual translation prone to errors)
- **Mitigation**: Use translation checklist, parallel updates in same commit
- **Fallback**: Mark translated READMEs with "Translation in progress" banner

### Risk 4: Skills Count Confusion
- **Impact**: Users see conflicting numbers (55 vs 56 vs 58)
- **Likelihood**: Low (addressed by "55+" standardization)
- **Mitigation**: Global search for ALL numeric Skills references
- **Fallback**: Add clarification note: "55+ Skills (continuously expanding)"

## Dependencies

**Blocking Dependencies**:
- v0.8.2 release tag created on GitHub
- Coverage measurement infrastructure working
- Access to all 4 README files (en, ko, ja, zh)

**Non-Blocking Dependencies**:
- CodeRabbit integration (optional quality check)
- Automated link checker (can be manual)
- Translation review by native speakers (optional polish)

## Success Metrics

1. **Completeness**: All 6 new versions documented with ≥3 features each ✅
2. **Accuracy**: Zero v0.4.11 references remaining ✅
3. **Consistency**: All Skills counts show "55+" ✅
4. **Validation**: All links return HTTP 200 ✅
5. **Coverage**: Badge accuracy within ±0.5% ✅
6. **Translation**: All 4 languages updated identically ✅

## Timeline Estimation

**Note**: MoAI-ADK does NOT provide time estimates. This is a **priority-based** workflow.

**Priority Sequence**:
1. **Primary Goal**: English README.md complete (Phase 2) + Pre-validation (Phase 1)
2. **Secondary Goal**: Translated READMEs complete (Phase 3)
3. **Final Goal**: Validation complete (Phase 4) + Git workflow (Phase 5)

**Dependencies**:
- Phase 2 → Phase 4 (English README must be done before validation)
- Phase 3 can run parallel with Phase 2
- Phase 5 depends on Phase 2, 3, 4 completion

## Next Steps

After implementation completion:
1. Run `/alfred:3-sync` to verify documentation consistency
2. Submit PR for review (if team mode)
3. Monitor PyPI/GitHub for release availability
4. Update internal documentation references if needed

---

**Last Updated**: 2025-10-29
**Author**: @GOOS