# ⚠️ BREAKING CHANGE: Alfred Skills Naming Migration (v0.26.0)

## Overview

**Version**: 0.26.0
**Release Date**: 2025-11-18
**Type**: BREAKING CHANGE - Hard Break (No Backward Compatibility)

v0.26.0부터 모든 `moai-alfred-*` Skills가 `moai-core-*`로 이름이 변경되었습니다.

**중요**: 이 변경은 하드 브레이크입니다. 구 이름(`moai-alfred-*`)은 v0.26.0부터 **즉시 제거**되었으며, 백워드 호환성이 없습니다.

---

## What Changed?

### All 21 Skills Renamed

모든 alfred-prefixed Skills가 `moai-core-*`로 통일되었습니다:

| Old Name | New Name |
|----------|----------|
| moai-alfred-workflow | moai-core-workflow |
| moai-alfred-personas | moai-core-personas |
| moai-alfred-context-budget | moai-core-context-budget |
| moai-alfred-agent-factory | moai-core-agent-factory |
| moai-alfred-agent-guide | moai-core-agent-guide |
| moai-alfred-ask-user-questions | moai-core-ask-user-questions |
| moai-alfred-clone-pattern | moai-core-clone-pattern |
| moai-alfred-code-reviewer | moai-core-code-reviewer |
| moai-alfred-config-schema | moai-core-config-schema |
| moai-alfred-dev-guide | moai-core-dev-guide |
| moai-alfred-env-security | moai-core-env-security |
| moai-alfred-expertise-detection | moai-core-expertise-detection |
| moai-alfred-feedback-templates | moai-core-feedback-templates |
| moai-alfred-issue-labels | moai-core-issue-labels |
| moai-alfred-language-detection | moai-core-language-detection |
| moai-alfred-practices | moai-core-practices |
| moai-alfred-proactive-suggestions | moai-core-proactive-suggestions |
| moai-alfred-rules | moai-core-rules |
| moai-alfred-session-state | moai-core-session-state |
| moai-alfred-spec-authoring | moai-core-spec-authoring |
| moai-alfred-todowrite-pattern | moai-core-todowrite-pattern |

### Naming Policy Rationale

**Before**: 21개 Skills가 모두 `moai-alfred-` prefix 사용 (persona-specific)
- ❌ Persona에 종속적
- ❌ 네이밍이 일관성 없음
- ❌ Alfred가 다른 이름으로 변경될 경우 혼란

**After**: 모든 Skills가 `moai-core-*` prefix 사용 (category-based)
- ✅ Persona-independent
- ✅ 명확한 카테고리 (core = MoAI-ADK의 핵심 Skills)
- ✅ 향후 유지보수 단순화
- ✅ 다른 persona 추가 시에도 코드/문서 영향 최소화

---

## Migration Guide

### For Package Users (Automatic)

v0.26.0으로 업그레이드하면 **자동으로** 새 이름이 적용됩니다:

```bash
# Package 업그레이드
uv sync

# Claude Code 재시작
# Skills는 자동으로 새 이름으로 로드됩니다
```

**Action Required**: None - Automatic migration

---

### For Local Projects (Manual Migration)

로컬 프로젝트에서 모든 참조를 업데이트해야 합니다.

#### Option 1: Automatic Migration Script (권장)

```bash
# Package의 마이그레이션 스크립트 사용
cd /path/to/your/project
uv run python .moai/scripts/migrate-naming-v026.py --execute

# 또는 dry-run으로 먼저 확인
uv run python .moai/scripts/migrate-naming-v026.py --dry-run
```

**실행되는 작업**:
1. Skill 디렉토리 리네임 (모두 `moai-core-*`로)
2. SKILL.md 메타데이터 업데이트
3. 모든 `Skill("moai-alfred-*")` 호출 → `Skill("moai-core-*")`로 변환
4. 에이전트, 커맨드, 문서 파일 자동 업데이트
5. 로그 파일 생성 (`.moai/logs/migration-v026.log`)

#### Option 2: Manual Migration

```bash
# 1. Skill 디렉토리 리네임
mv .claude/skills/moai-alfred-workflow .claude/skills/moai-core-workflow
mv .claude/skills/moai-alfred-personas .claude/skills/moai-core-personas
# ... (21개 모두)

# 2. 파일 내 참조 업데이트
# .claude/agents/**/*.md
# .claude/commands/**/*.md
# .claude/skills/**/SKILL.md
# CLAUDE.md

# 예: sed 사용
sed -i '' 's/Skill("moai-alfred-/Skill("moai-core-/g' .claude/agents/**/*.md
sed -i '' 's/Skill("moai-alfred-/Skill("moai-core-/g' .claude/commands/**/*.md
sed -i '' 's/Skill("moai-alfred-/Skill("moai-core-/g' .claude/skills/**/SKILL.md
sed -i '' 's/moai-alfred-/moai-core-/g' CLAUDE.md
```

---

## Impact Analysis

### What Will Break?

v0.26.0로 업그레이드한 후 다음이 **작동하지 않습니다**:

```python
# ❌ BROKEN (v0.26.0+)
Skill("moai-alfred-workflow")
# → Error: SkillNotFound: moai-alfred-workflow

# ✅ WORKS
Skill("moai-core-workflow")
```

### Files Affected

마이그레이션이 영향을 미치는 파일들:

1. **Skill Directories** (21개)
   - `src/moai_adk/templates/.claude/skills/moai-alfred-*` → `moai-core-*`
   - `.claude/skills/moai-alfred-*` → `moai-core-*`

2. **Agent Files** (~23개)
   - `src/moai_adk/templates/.claude/agents/alfred/*.md`
   - `.claude/agents/alfred/*.md`

3. **Command Files** (~4개)
   - `src/moai_adk/templates/.claude/commands/alfred/*.md`
   - `.claude/commands/alfred/*.md`

4. **Documentation** (2개)
   - `src/moai_adk/templates/CLAUDE.md`
   - `CLAUDE.md`

5. **Memory Files** (~5개)
   - `.moai/memory/*.md`

**Total**: 130+ files modified, 160+ changes applied

---

## Pre-Migration Checklist

업그레이드 전에 확인하세요:

- [ ] 현재 모든 로컬 변경사항 커밋됨
- [ ] 로컬 프로젝트가 up-to-date 상태
- [ ] `git status`에 uncommitted changes 없음
- [ ] 백업 브랜치 생성 (`git branch backup-pre-v026`)

---

## Post-Migration Validation

마이그레이션 후 검증하세요:

```bash
# 1. 구 이름 참조 확인 (없어야 함)
grep -r "moai-alfred-" src/moai_adk/templates/ .claude/ CLAUDE.md
# → No output (zero matches)

# 2. 새 이름 확인 (많아야 함)
grep -r "moai-core-" src/moai_adk/templates/ .claude/ CLAUDE.md
# → Multiple matches

# 3. Skill 로드 테스트
python3 -c "from moai_adk.skills import Skill; s = Skill('moai-core-workflow'); print('✅ OK')"

# 4. Claude Code 테스트
# Claude Code를 재시작하고 Skill() 호출 테스트
Skill("moai-core-workflow")  # Should work
Skill("moai-alfred-workflow")  # Should fail
```

---

## Rollback Plan

마이그레이션을 취소하려면:

```bash
# Option 1: Git 사용
git reset --hard HEAD~1

# Option 2: 마이그레이션 스크립트 사용
uv run python .moai/scripts/migrate-naming-v026.py --rollback

# Option 3: 백업 브랜치 복원
git checkout backup-pre-v026
```

---

## FAQ

### Q: 왜 이런 변경을 했나요?

A: 다음 이유들로 인해 네이밍 정책을 변경했습니다:

1. **Persona 독립성**: Alfred가 다른 이름(Yoda, R2-D2 등)으로 변경되어도 코드 영향 최소화
2. **명확성**: `moai-core-*`는 "MoAI-ADK의 핵심 Skills"를 명확히 의미
3. **확장성**: 향후 새로운 카테고리 추가 시 체계적 확장 가능
4. **간결성**: 21개 Skills가 6개 카테고리 대신 1개 카테고리로 통일

### Q: 로컬 프로젝트는 어떻게 되나요?

A: 로컬 프로젝트도 업데이트가 필요합니다:

1. **자동**: `migrate-naming-v026.py --execute` 스크립트 실행
2. **수동**: 문서의 "Manual Migration" 섹션 참고

### Q: 롤백할 수 있나요?

A: 네, 언제든지 롤백할 수 있습니다:

```bash
# Option 1: Git revert
git reset --hard <backup-commit>

# Option 2: Migration script
python .moai/scripts/migrate-naming-v026.py --rollback
```

### Q: 이미 커밋한 코드는?

A: 이전 버전의 코드는 여전히 작동합니다 (그 버전의 Skills 사용). 새 버전으로 업그레이드할 때만 마이그레이션이 필요합니다.

### Q: 다른 버전에서는?

A:
- **v0.25.x 이하**: `moai-alfred-*` 사용 (변경 없음)
- **v0.26.0+**: `moai-core-*` 사용 (필수 마이그레이션)

---

## Migration Statistics

### Automated Process

마이그레이션 스크립트 실행 결과:

- **Total Skills Renamed**: 21
- **Total Changes Applied**: 160
- **Errors**: 0
- **Execution Time**: ~5 seconds
- **Log File**: `.moai/logs/migration-v026.log`

### Files Modified

| File Type | Count | Details |
|-----------|-------|---------|
| Skill Directories | 21 | Renamed from `moai-alfred-*` to `moai-core-*` |
| Skill Metadata | 21 | Updated SKILL.md name fields |
| Agent Files | 23 | Updated Skill() references |
| Command Files | 4 | Updated Skill() references |
| CLAUDE.md | 2 | Updated documentation |
| Memory Files | 5 | Updated cross-references |
| Other Skills | 75+ | Updated depends_on references |
| **Total** | **130+** | **160+ individual changes** |

---

## Support & Troubleshooting

### Common Errors

**Error**: `SkillNotFound: moai-alfred-workflow`

**Cause**: 구 이름 사용 (v0.26.0 이상에서 제거됨)

**Solution**:
```python
# ❌ Wrong
Skill("moai-alfred-workflow")

# ✅ Correct
Skill("moai-core-workflow")
```

---

**Error**: Migration script not found

**Solution**:
```bash
# .moai/scripts/migrate-naming-v026.py 확인
ls -la .moai/scripts/migrate-naming-v026.py

# 없으면 package에서 복사
cp src/moai_adk/templates/.moai/scripts/migrate-naming-v026.py .moai/scripts/
```

---

**Error**: Permission denied when running script

**Solution**:
```bash
chmod +x .moai/scripts/migrate-naming-v026.py
python3 .moai/scripts/migrate-naming-v026.py --execute
```

---

## Additional Resources

- **Changelog**: See `CHANGELOG.md` for v0.26.0 release notes
- **Migration Script**: `.moai/scripts/migrate-naming-v026.py`
- **Migration Log**: `.moai/logs/migration-v026.log` (after running script)
- **Related Issues**: Check GitHub issues with label `naming-migration-v026`

---

## Timeline

| Date | Event |
|------|-------|
| 2025-11-18 | v0.26.0 released with breaking change |
| 2025-11-18 | Migration script provided |
| TBD | v0.27.0 planning (no more alfred references) |

---

## Acknowledgments

이 마이그레이션은 MoAI-ADK의 장기적 유지보수성과 확장성을 개선하기 위해 진행되었습니다.

**Questions?** GitHub Issues를 통해 문의하세요.

---

**Last Updated**: 2025-11-18
**Version**: v0.26.0
**Status**: BREAKING CHANGE APPLIED
