# @SPEC:SKILLS-REDESIGN-001 마이그레이션 계획

> **목표**: 46개 → 44개 스킬을 4개 Tier로 계층화 (4주)
> **작업량**: 44개 SKILL.md 재구성
> **난이도**: 중간 (자동화 가능한 대부분 작업)

---

## 📅 Phase별 계획

### Phase 1: Foundation 스킬 재구성 (1주)

**목표**: Tier 1 (6개) 완성

#### 작업 1: Alfred 6개 스킬 재명명

```bash
# .claude/skills/ 디렉토리에서

1. moai-alfred-trust-validation → moai-foundation-trust
2. moai-alfred-tag-scanning → moai-foundation-tags
3. moai-alfred-spec-metadata-validation → moai-foundation-specs
4. moai-alfred-ears-authoring → moai-foundation-ears
5. moai-alfred-git-workflow → moai-foundation-git
6. moai-alfred-language-detection → moai-foundation-langs
```

#### 작업 2: SKILL.md 표준화 (6개 모두)

**변경 전**:
```yaml
---
name: moai-alfred-trust-validation
description: ...
version: 0.1.0
author: MoAI Skill Factory
license: MIT
tags:
  - trust
  - quality
---
```

**변경 후**:
```yaml
---
name: moai-foundation-trust
description: Validates TRUST 5-principles (Test 85%+, Readable, Unified, Secured, Trackable)
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Foundation: TRUST Validation

## What it does
Validates MoAI-ADK's TRUST 5-principles compliance...

## When to use
- `/alfred:3-sync` 자동 호출
- "TRUST 원칙 확인", "품질 검증"

## Works well with
- moai-foundation-tags (TAG 체인 검증)
- moai-foundation-specs (SPEC 메타데이터 검증)
```

**변경사항**:
- ❌ 제거: version, author, license, tags
- ✅ 추가: allowed-tools, "Works well with"
- ✅ 구조 정리: SKILL.md <500 words 목표

#### 작업 3: Templates 동기화

```bash
# templates/.claude/skills/ 디렉토리도 동기화
cp -r .claude/skills/moai-foundation-* templates/.claude/skills/
```

**검증 명령어**:
```bash
# 6개 Foundation 스킬 확인
ls -la .claude/skills/moai-foundation-*
ls -la templates/.claude/skills/moai-foundation-*
# 같은 개수 확인
```

---

### Phase 2: Essentials 스킬 재구성 + 삭제 (1주)

**목표**: Tier 2 (4개) 완성 + 2개 삭제

#### 작업 1: Alfred 나머지 4개 재명명

```bash
7. moai-alfred-code-reviewer → moai-essentials-review
8. moai-alfred-debugger-pro → moai-essentials-debug
9. moai-alfred-refactoring-coach → moai-essentials-refactor
10. moai-alfred-performance-optimizer → moai-essentials-perf
```

#### 작업 2: 2개 스킬 삭제

```bash
# .claude/skills/ 에서 삭제
rm -rf moai-alfred-template-generator
rm -rf moai-alfred-feature-selector

# templates/.claude/skills/ 에서도 삭제
rm -rf templates/.claude/skills/moai-alfred-template-generator
rm -rf templates/.claude/skills/moai-alfred-feature-selector
```

**기능 이관 확인**:
- ✅ template-generator → moai-claude-code/templates/ 디렉토리
- ✅ feature-selector → /alfred:1-plan 명령어 내부 로직

#### 작업 3: SKILL.md 표준화 (4개 모두)

Phase 1과 동일한 표준화 적용 (version, author, license, tags 제거)

#### 작업 4: Templates 동기화

```bash
cp -r .claude/skills/moai-essentials-* templates/.claude/skills/
rm -rf templates/.claude/skills/moai-alfred-template-generator
rm -rf templates/.claude/skills/moai-alfred-feature-selector
```

**검증**:
```bash
# Alfred 12개 → 10개 확인
ls -la .claude/skills/ | grep moai-alfred | wc -l  # 0개 (모두 재명명됨)
ls -la .claude/skills/ | grep moai-foundation | wc -l  # 6개
ls -la .claude/skills/ | grep moai-essentials | wc -l  # 4개
```

---

### Phase 3: Language/Domain 검증 및 표준화 (1주)

**목표**: Tier 3-4 (33개) 검증 및 Progressive Disclosure 적용

#### 작업 1: Language 스킬 (24개) 검증

```bash
# Tier 3 마크 추가 (SKILL.md에)
for skill in .claude/skills/moai-lang-*; do
  # 각 SKILL.md에 다음 추가:
  # tier: 3
  # auto-load: "true"  (moai-foundation-langs가 감지할 수 있게)
done
```

#### 작업 2: Domain 스킬 (9개) 검증

```bash
# Tier 4 마크 추가
for skill in .claude/skills/moai-domain-*; do
  # 각 SKILL.md에 다음 추가:
  # tier: 4
  # auto-load: "false"  (사용자 요청 기반)
done
```

#### 작업 3: "Works well with" 섹션 추가 (33개 모두)

**예시 (moai-lang-python)**:
```markdown
## Works well with
- moai-foundation-langs (자동 로드 트리거)
- moai-essentials-review (코드 리뷰)
- moai-essentials-debug (디버깅)
- moai-essentials-refactor (리팩토링)
- moai-essentials-perf (성능 최적화)
```

**예시 (moai-domain-backend)**:
```markdown
## Works well with
- moai-lang-typescript (TypeScript 프로젝트)
- moai-lang-python (Python FastAPI)
- moai-lang-go (Go Gin)
- moai-essentials-perf (API 성능 최적화)
```

#### 작업 4: SKILL.md 크기 검증

```bash
# 모든 스킬의 SKILL.md 크기 확인
for skill in .claude/skills/moai-*/SKILL.md; do
  lines=$(wc -l < "$skill")
  chars=$(wc -c < "$skill")
  echo "$skill: $lines lines, $chars chars"
done

# 제약사항 검증:
# - Language/Domain: <100줄
# - Foundation: 총 <500줄
# - description: <200 chars
```

#### 작업 5: Templates 동기화

```bash
# Language/Domain 동기화
cp -r .claude/skills/moai-lang-* templates/.claude/skills/
cp -r .claude/skills/moai-domain-* templates/.claude/skills/
```

---

### Phase 4: 통합 테스트 및 검증 (1주)

**목표**: 전체 4-Tier 구조 검증

#### 작업 1: 구조 검증

```bash
# 총 44개 스킬 확인
find .claude/skills -name "SKILL.md" | wc -l  # 44개

# Tier별 개수 확인
ls -la .claude/skills/moai-foundation-* | wc -l  # 6개
ls -la .claude/skills/moai-essentials-* | wc -l  # 4개
ls -la .claude/skills/moai-lang-* | wc -l  # 24개
ls -la .claude/skills/moai-domain-* | wc -l  # 9개
ls -la .claude/skills/moai-claude-code | wc -l  # 1개
# 합계: 44개
```

#### 작업 2: 워크플로우 통합 테스트

```bash
# 테스트 1: /alfred:1-plan 시 Tier 1 로드 확인
/alfred:1-plan "테스트 기능"
# 확인: moai-foundation-ears, moai-foundation-specs, moai-foundation-langs 로드됨
# 확인: Tier 3/4 스킬은 로드되지 않음

# 테스트 2: /alfred:2-run 시 Language 자동 로드
/alfred:2-run AUTH-001
# 프로젝트 언어: Python
# 확인: moai-lang-python 자동 로드됨
# 확인: 다른 23개 언어 스킬은 로드되지 않음

# 테스트 3: /alfred:3-sync 시 Foundation 조합
/alfred:3-sync
# 확인: moai-foundation-trust, tags, specs, git 함께 작동
# 확인: Essentials는 필요 시만 호출됨

# 테스트 4: Domain 스킬 선택적 로드
# 사용자: "백엔드 API 설계"
# 확인: moai-domain-backend 로드
# 확인: 다른 8개 도메인은 로드 안 됨
```

#### 작업 3: 문서 검증

```bash
# SKILL.md 형식 검증
for skill in .claude/skills/moai-*/SKILL.md; do
  # 필수 필드 확인
  grep -q "^name:" "$skill" || echo "MISSING: name in $skill"
  grep -q "^description:" "$skill" || echo "MISSING: description in $skill"
  grep -q "^allowed-tools:" "$skill" || echo "MISSING: allowed-tools in $skill"
done

# 크기 검증
for skill in .claude/skills/moai-*/SKILL.md; do
  lines=$(wc -l < "$skill")
  if [[ $lines -gt 100 ]]; then
    echo "TOO LARGE: $skill ($lines lines)"
  fi
done
```

#### 작업 4: Progressive Disclosure 검증

```bash
# Tier 3 자동 로드 메커니즘 검증
grep -r "auto-load.*true" .claude/skills/moai-lang-*
# 24개 모두 "auto-load: true" 설정 확인

# Tier 4 수동 로드 메커니즘 검증
grep -r "auto-load.*false" .claude/skills/moai-domain-*
# 9개 모두 "auto-load: false" 설정 확인
```

#### 작업 5: Git 커밋

```bash
# Phase 별 커밋

# Phase 1: Foundation 재구성
git add .claude/skills/moai-foundation-* templates/.claude/skills/moai-foundation-*
git commit -m "🟢 Foundation Skills 표준화 (6개, Tier 1 완성)"

# Phase 2: Essentials 재구성 + 삭제
git add .claude/skills/moai-essentials-* templates/.claude/skills/moai-essentials-*
git rm -r .claude/skills/moai-alfred-template-generator
git rm -r .claude/skills/moai-alfred-feature-selector
git rm -r templates/.claude/skills/moai-alfred-template-generator
git rm -r templates/.claude/skills/moai-alfred-feature-selector
git commit -m "🟢 Essentials Skills 표준화 + 2개 삭제 (Tier 2 완성)"

# Phase 3: Language/Domain 검증
git add .claude/skills/moai-lang-* .claude/skills/moai-domain-*
git commit -m "🟢 Language/Domain Skills 표준화 (Tier 3-4 완성)"

# Phase 4: 최종 검증
git add .moai/reports/
git commit -m "📚 Skills 재설계 완료 - 4-Tier 아키텍처 (46→44개)"
```

---

## 📊 작업 일정

| Phase | 작업 | 일수 | 상태 |
|-------|------|------|------|
| **1** | Foundation 6개 | 2-3일 | 예정 |
| **2** | Essentials 4개 + 2개 삭제 | 2-3일 | 예정 |
| **3** | Language/Domain 33개 검증 | 2-3일 | 예정 |
| **4** | 통합 테스트 + Git 커밋 | 1-2일 | 예정 |
| **합계** | - | **7-11일** | 약 2주 |

---

## 🛠️ 자동화 스크립트

### 스크립트 1: SKILL.md 표준화 자동화

**파일**: `scripts/standardize_skills_v2.py`

```python
#!/usr/bin/env python3
"""
Skills v0.4.0 표준화 스크립트

작업:
1. YAML frontmatter 정리 (version, author, license, tags 제거)
2. allowed-tools 자동 추가
3. "Works well with" 섹션 자동 생성
4. Tier 메타데이터 추가
"""

import re
import os
from pathlib import Path

def standardize_skill(skill_dir):
    """특정 스킬 표준화"""

    skill_md = skill_dir / "SKILL.md"
    if not skill_md.exists():
        return False

    content = skill_md.read_text()

    # 1. YAML frontmatter 추출
    yaml_match = re.match(r'^---\n(.*?)\n---\n(.*)', content, re.DOTALL)
    if not yaml_match:
        return False

    yaml_content = yaml_match.group(1)
    body_content = yaml_match.group(2)

    # 2. 필수 필드 추출
    name = re.search(r'^name:\s*(.+)$', yaml_content, re.MULTILINE)
    desc = re.search(r'^description:\s*(.+)$', yaml_content, re.MULTILINE)

    # 3. 새 YAML 생성 (간소화)
    new_yaml = f"""---
name: {name.group(1) if name else 'unknown'}
description: {desc.group(1) if desc else 'No description'}
allowed-tools:
  - Read
"""

    # tier 자동 결정
    skill_name = skill_dir.name
    if 'foundation' in skill_name:
        new_yaml += "  - Bash\n  - Write\n  - Edit\n  - TodoWrite\n"
    elif 'essentials' in skill_name:
        new_yaml += "  - Bash\n  - Write\n  - Edit\n  - TodoWrite\n"
    else:
        new_yaml += "  - Bash\n"

    new_yaml += "---\n"

    # 4. 새 파일 저장
    new_content = new_yaml + body_content
    skill_md.write_text(new_content)

    return True

# 실행
skills_dir = Path(".claude/skills")
for skill_dir in skills_dir.iterdir():
    if skill_dir.is_dir() and not skill_dir.name.startswith('.'):
        print(f"Standardizing {skill_dir.name}...")
        standardize_skill(skill_dir)
```

### 스크립트 2: Tier 메타데이터 추가

**파일**: `scripts/add_tier_metadata.py`

```python
#!/usr/bin/env python3
"""
각 스킬에 Tier 메타데이터 추가
"""

import re
from pathlib import Path

def add_tier(skill_dir):
    """스킬에 Tier 메타데이터 추가"""

    skill_name = skill_dir.name
    skill_md = skill_dir / "SKILL.md"

    # Tier 결정
    if 'foundation' in skill_name:
        tier = 1
        auto_load = 'true'
    elif 'essentials' in skill_name:
        tier = 2
        auto_load = 'true'
    elif 'lang' in skill_name:
        tier = 3
        auto_load = 'true'
    elif 'domain' in skill_name:
        tier = 4
        auto_load = 'false'
    else:
        return False

    content = skill_md.read_text()

    # YAML에 tier 추가
    yaml_pattern = r'(---\n.*?^allowed-tools:.*?\n(?:  - .*\n)+)(---)'
    replacement = f'\\1tier: {tier}\nauto-load: "{auto_load}"\n\\2'

    new_content = re.sub(yaml_pattern, replacement, content, flags=re.MULTILINE | re.DOTALL)
    skill_md.write_text(new_content)

    return True

# 실행
skills_dir = Path(".claude/skills")
for skill_dir in skills_dir.iterdir():
    if skill_dir.is_dir() and not skill_dir.name.startswith('.'):
        print(f"Adding tier to {skill_dir.name}...")
        add_tier(skill_dir)
```

---

## ✅ 검증 체크리스트

### 구조 검증
- [ ] 총 44개 스킬 확인 (46 - 2 = 44)
- [ ] Tier 1 (Foundation): 6개
- [ ] Tier 2 (Essentials): 4개
- [ ] Tier 3 (Language): 24개
- [ ] Tier 4 (Domain): 9개
- [ ] Claude Code: 1개

### 네이밍 검증
- [ ] moai-alfred-* (0개 - 모두 재명명됨)
- [ ] moai-foundation-* (6개)
- [ ] moai-essentials-* (4개)
- [ ] moai-lang-* (24개)
- [ ] moai-domain-* (9개)

### SKILL.md 검증
- [ ] 모든 SKILL.md에 allowed-tools 필드 있음
- [ ] 모든 SKILL.md에 "Works well with" 섹션 있음
- [ ] version, author, license, tags 필드 제거됨
- [ ] description 필드 <200 chars
- [ ] 각 파일 <100줄 (Foundation 제외, 총 <500줄)

### Progressive Disclosure 검증
- [ ] Tier 3 스킬에 "auto-load: true" 설정
- [ ] Tier 4 스킬에 "auto-load: false" 설정
- [ ] moai-foundation-langs 설정 확인

### 워크플로우 통합 검증
- [ ] /alfred:1-plan 실행 시 Tier 1만 로드
- [ ] /alfred:2-run 실행 시 Tier 3 자동 로드
- [ ] /alfred:3-sync 실행 시 Tier 1 조합 작동

---

**작성**: SPEC-SKILLS-REDESIGN-001 마이그레이션 계획
**최종 업데이트**: 2025-10-19
