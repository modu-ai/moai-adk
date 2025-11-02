# .claude 인프라 동기화 정책

**Document Language**: 한국어
**Last Updated**: 2025-11-02
**Status**: Active Policy

## 개요

MoAI-ADK 패키지의 `.claude/` 인프라 파일들은 **패키지 템플릿을 Source of Truth**로 운영합니다.

## 원칙

### 1. Source of Truth (진실의 원천)

```
패키지 템플릿 (Priority 1)
    src/moai_adk/templates/.claude/

↓ Sync (단방향)

로컬 프로젝트 (Priority 2)
    .claude/
```

- **패키지 템플릿이 항상 우선**
- 로컬 파일은 추적 제외 (`.gitignore`)
- 변경 사항은 패키지 템플릿에서만 적용

### 2. 파일 범위

**추적 제외 (Git 제외):**
```
.claude/commands/      # Alfred 커맨드 정의
.claude/agents/        # Sub-agent 지침
.claude/hooks/         # Claude Code 훅
.claude/skills/        # Skill 라이브러리
```

**추적 포함 (Git 포함):**
```
.claude/settings.json       # 로컬 설정
.claude/settings.local.json # 개발자 설정 (개인용)
```

## 동기화 워크플로우

### 시나리오 1: 패키지 업그레이드 후 로컬 동기화

```bash
# 1. 패키지 업그레이드
uv pip install --upgrade moai-adk

# 2. 패키지 템플릿 확인 (필요시)
ls src/moai_adk/templates/.claude/

# 3. 로컬 동기화 (수동)
cp -r src/moai_adk/templates/.claude/* .claude/

# 4. settings.json 확인 (덮어씌워지지 않음)
# 기존 .claude/settings*.json 파일은 보존됨
```

### 시나리오 2: 로컬 프로젝트에서 .claude 파일 수정 필요

**❌ 하지 말아야 할 일:**
```bash
# 로컬 .claude 파일을 직접 수정하면 안 됩니다
# 패키지 업그레이드 시 덮어씌워집니다
vim .claude/commands/alfred/1-plan.md  # ❌ 금지
```

**✅ 해야 할 일:**
```bash
# 1. 패키지 템플릿 수정 (Pull Request)
# 2. 패키지 배포 대기
# 3. 패키지 설치 후 로컬 동기화
```

## 로컬 커스터마이징

로컬 커스터마이징이 필요한 경우:

```bash
# .moai/local-overrides/ 디렉토리 사용
mkdir -p .moai/local-overrides/commands/

# 필요한 파일만 복사하여 수정
cp .claude/commands/alfred/1-plan.md .moai/local-overrides/commands/
vim .moai/local-overrides/commands/1-plan.md

# .gitignore에 추가
echo ".moai/local-overrides/" >> .gitignore
```

## 문제 해결

### Q: "패키지를 업그레이드했는데 로컬 .claude 파일이 구 버전입니다"

**A:** 동기화 필요합니다:

```bash
cp -r src/moai_adk/templates/.claude/* .claude/
```

### Q: "로컬 .claude 파일을 수정했는데 반영되지 않습니다"

**A:** `.claude/` 파일은 Git 추적 제외입니다. 패키지 템플릿을 수정해야 합니다:

```bash
# 1. 패키지 템플릿 수정
vim src/moai_adk/templates/.claude/commands/alfred/1-plan.md

# 2. 변경사항 커밋 및 패키지 배포
git add src/moai_adk/templates/
git commit -m "feat: Update command prompt"
```

### Q: ".claude/settings.json이 덮어씌워졌습니다"

**A:** `.claude/` 전체 복사를 피하고 필요한 파일만 복사하세요:

```bash
# 제외: settings.json, settings.local.json
rsync -av --exclude="settings*.json" \
  src/moai_adk/templates/.claude/ .claude/
```

## 자동 동기화 (향후)

향후 다음 중 하나 구현 예정:

### Option A: 스크립트 기반 동기화

```bash
#!/bin/bash
# .moai/scripts/sync-claude-infrastructure.sh

TEMPLATE_DIR="src/moai_adk/templates/.claude"
LOCAL_DIR=".claude"

# settings.json 제외하고 동기화
rsync -av --exclude="settings*.json" \
  "$TEMPLATE_DIR/" "$LOCAL_DIR/"
```

### Option B: Pre-commit Hook

```python
# .git/hooks/pre-commit
# 커밋 전 자동 동기화 확인
```

### Option C: Symlink 기반

```bash
# .claude 디렉토리를 템플릿으로의 심링크로 대체
# (권장: 가장 안전하고 간단함)
```

## 체크리스트

패키지 릴리스 전 확인:

- [ ] `src/moai_adk/templates/.claude/` 모든 변경 커밋
- [ ] TAG 중복 없음 (pre-commit hook)
- [ ] 릴리스 노트 업데이트

로컬 동기화 후 확인:

- [ ] `cp -r src/moai_adk/templates/.claude/* .claude/` 실행
- [ ] `.claude/settings*.json` 보존됨 확인
- [ ] Alfred 커맨드 정상 작동 확인

## 참고

- **TAG 정책**: 패키지 템플릿 내 @TAG만 유지 (로컬 TAG 제외)
- **Skill 정책**: `src/moai_adk/templates/.claude/skills/` 내 모든 Skill 유지
- **Hook 정책**: 패키지 배포 시에만 로컬 hook 업데이트
