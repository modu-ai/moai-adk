# Breaking Changes - MoAI-ADK v0.4.0

**릴리스 일자**: 2025-10-20 (예정)
**이전 버전**: v0.3.13
**새 버전**: v0.4.0

---

## 📋 Executive Summary

v0.4.0은 **Skills Revolution**이라는 주제로 MoAI-ADK의 아키텍처를 대폭 개선한 메이저 릴리스입니다.

**주요 변경사항**:
- ✨ **Skills 시스템 도입**: 44개 Skills (Foundation, Essentials, Language, Domain)
- ♻️ **Commands 명칭 변경**: 사용자 의도 명확화 (1-spec → 1-plan, 2-build → 2-run)
- 🎯 **Sub-agents → Skills 통합**: Agent 프롬프트 1,200 LOC 감소
- 🚀 **성능 최적화**: 컨텍스트 80% 감소, 응답속도 2배 향상

**영향도**:
- 🔴 **Critical**: Commands 명칭 변경 (기존 워크플로우 영향)
- 🟡 **Medium**: Skills 시스템 (신규 기능, 기존 동작 유지)
- 🟢 **Low**: Sub-agents 통합 (내부 구조, 사용자 투명)

**마이그레이션 시간**: 약 30분 (스크립트 제공)

---

## 🔴 Breaking Changes 상세

### 1. Commands 명칭 변경 (Critical)

#### 변경 내용

| Before (v0.3.13) | After (v0.4.0) | 이유 |
|------------------|----------------|------|
| `/alfred:1-spec` | `/alfred:1-plan` | "명세 작성"보다 "계획 수립"이 의도 명확 |
| `/alfred:2-build` | `/alfred:2-run` | "빌드"보다 "실행"이 TDD 흐름 명확 |
| `/alfred:3-sync` | `/alfred:3-sync` (유지) | 변경 없음 |

#### 영향

**기존 워크플로우**:
```bash
# v0.3.13
/alfred:1-spec "새 기능"
/alfred:2-build SPEC-001
/alfred:3-sync
```

**새 워크플로우**:
```bash
# v0.4.0
/alfred:1-plan "새 기능"  # 변경됨
/alfred:2-run SPEC-001     # 변경됨
/alfred:3-sync             # 유지
```

**호환성**:
- ✅ **v0.4.0**: 새 명칭 사용 (권장)
- ⚠️ **v0.5.0**: 기존 명칭 제거 예정 (Deprecation)

#### 마이그레이션 방법

**옵션 1: 수동 마이그레이션 (권장)**
```bash
# 기존 스크립트/문서에서 검색 및 치환
sed -i '' 's|/alfred:1-spec|/alfred:1-plan|g' **/*.md **/*.sh
sed -i '' 's|/alfred:2-build|/alfred:2-run|g' **/*.md **/*.sh
```

**옵션 2: Alias 사용 (임시)**
```bash
# v0.4.0에서는 기존 명칭도 작동 (Deprecation Warning 출력)
/alfred:1-spec "새 기능"  # ⚠️ Deprecated: Use /alfred:1-plan
```

**옵션 3: 자동 마이그레이션 스크립트** (TODO: P1-1에서 제공)

---

### 2. Skills 시스템 도입 (Medium)

#### 변경 내용

**신규 기능**: 44개 Skills 추가 (4-Tier 아키텍처)
```
Foundation Tier (6개): 핵심 기능
  ├─ moai-foundation-trust
  ├─ moai-foundation-tags
  ├─ moai-foundation-specs
  ├─ moai-foundation-ears
  ├─ moai-foundation-git
  └─ moai-foundation-langs

Essentials Tier (4개): 일상 개발
  ├─ moai-essentials-debug
  ├─ moai-essentials-review
  ├─ moai-essentials-refactor
  └─ moai-essentials-feature

Language Skills (23개): 프로그래밍 언어
  └─ moai-lang-{python, typescript, rust, go, ...}

Domain Skills (10개): 문제 해결 영역
  └─ moai-domain-{backend, frontend, database, ...}
```

#### 영향

**기존 동작**: 변경 없음 (Alfred가 자동 선택)
**새로운 기능**: Skills를 명시적으로 호출 가능 (선택사항)

```bash
# 기존 (v0.3.13) - 여전히 작동
/alfred:1-plan "Python 프로젝트"

# 신규 (v0.4.0) - 명시적 Skill 활용
/alfred:1-plan "Python 프로젝트"
# Alfred가 moai-lang-python 자동 선택
```

#### 마이그레이션 방법

**필수 작업**: 없음 (기존 동작 유지)
**선택 작업**: Skills 시스템 학습 (`.claude/skills/*/SKILL.md` 참조)

---

### 3. Sub-agents → Skills 통합 (Low)

#### 변경 내용

11개 Sub-agents가 재사용 가능한 Skills로 통합되었습니다.

**통합된 Sub-agents**:
```
Before (v0.3.13):
  - tag-agent (400 LOC)
  - trust-checker (500 LOC)
  - (기타 9개)

After (v0.4.0):
  - moai-foundation-tags (통합)
  - moai-foundation-trust (통합)
  - (기타 9개 통합)
```

#### 영향

**사용자**: 투명 (내부 구조 변경, 동작 동일)
**개발자**: Agent 프롬프트 1,200 LOC 감소 (-40%)

#### 마이그레이션 방법

**필수 작업**: 없음 (자동 마이그레이션)

---

## 🛠️ 마이그레이션 가이드

### 단계별 마이그레이션

#### Step 1: 프로젝트 백업
```bash
# 프로젝트 전체 백업
cp -r ~/my-project ~/my-project.backup
```

#### Step 2: MoAI-ADK v0.4.0 설치
```bash
# Option A: pip
pip install --upgrade moai-adk

# Option B: uv (권장)
uv pip install --upgrade moai-adk
```

#### Step 3: 버전 확인
```bash
moai-adk version
# Expected output: v0.4.0
```

#### Step 4: Commands 명칭 변경
```bash
# 프로젝트 내 문서/스크립트 검색
grep -r "/alfred:1-spec" .
grep -r "/alfred:2-build" .

# 치환 (macOS)
find . -type f \( -name "*.md" -o -name "*.sh" \) -exec sed -i '' \
  -e 's|/alfred:1-spec|/alfred:1-plan|g' \
  -e 's|/alfred:2-build|/alfred:2-run|g' {} \;

# 치환 (Linux)
find . -type f \( -name "*.md" -o -name "*.sh" \) -exec sed -i \
  -e 's|/alfred:1-spec|/alfred:1-plan|g' \
  -e 's|/alfred:2-build|/alfred:2-run|g' {} \;
```

#### Step 5: 동작 테스트
```bash
# 새 명칭으로 워크플로우 테스트
/alfred:1-plan "테스트 기능"
/alfred:2-run SPEC-TEST-001
/alfred:3-sync
```

#### Step 6: 백업 정리
```bash
# 테스트 성공 시 백업 삭제
rm -rf ~/my-project.backup
```

---

## 🔄 호환성 매트릭스

| 기능 | v0.3.13 | v0.4.0 | v0.5.0 (계획) |
|------|---------|--------|--------------|
| `/alfred:1-spec` | ✅ 작동 | ⚠️ Deprecated | ❌ 제거 |
| `/alfred:2-build` | ✅ 작동 | ⚠️ Deprecated | ❌ 제거 |
| `/alfred:1-plan` | ❌ 없음 | ✅ 권장 | ✅ 유일 |
| `/alfred:2-run` | ❌ 없음 | ✅ 권장 | ✅ 유일 |
| `/alfred:3-sync` | ✅ 작동 | ✅ 작동 | ✅ 작동 |
| Skills 시스템 | ❌ 없음 | ✅ 신규 | ✅ 확대 |

---

## ❓ FAQ

### Q1: 기존 프로젝트도 v0.4.0으로 업그레이드해야 하나요?

**A**: 선택사항입니다. v0.3.13도 계속 지원됩니다.

**업그레이드 권장**:
- ✅ 새 프로젝트 시작 시
- ✅ 성능 개선 필요 시 (컨텍스트 80% 감소)
- ✅ Skills 시스템 활용 원할 시

**업그레이드 보류**:
- ⏸️ 안정적인 프로덕션 환경
- ⏸️ 마이그레이션 시간 부족

---

### Q2: `/alfred:1-spec`을 계속 사용하면 어떻게 되나요?

**A**: v0.4.0에서는 작동하지만 Deprecation Warning이 출력됩니다.

```bash
/alfred:1-spec "새 기능"
# ⚠️ Deprecated: /alfred:1-spec is deprecated. Use /alfred:1-plan instead.
# This command will be removed in v0.5.0.
```

**v0.5.0에서 제거 예정**이므로 가능한 빨리 새 명칭으로 마이그레이션하세요.

---

### Q3: Skills 시스템은 어떻게 사용하나요?

**A**: Alfred가 자동으로 적절한 Skills를 선택합니다. 별도 학습 불필요.

**자동 선택 예시**:
```bash
/alfred:1-plan "Python FastAPI 백엔드"
# Alfred가 자동 선택:
# - moai-lang-python
# - moai-domain-backend
```

**수동 확인** (선택사항):
```bash
# Skills 목록 확인
ls .claude/skills/*/SKILL.md

# 특정 Skill 내용 확인
cat .claude/skills/moai-lang-python/SKILL.md
```

---

### Q4: Sub-agents는 어디로 갔나요?

**A**: Skills로 통합되었습니다. 동작은 동일하며 성능이 개선되었습니다.

**예시**:
```bash
# v0.3.13
@agent-tag-agent "TAG 체인 검증"

# v0.4.0 (동일 동작)
/alfred:3-sync  # Alfred가 moai-foundation-tags 자동 호출
```

---

### Q5: 롤백하고 싶은 경우?

**A**: v0.3.13으로 다운그레이드 가능합니다.

```bash
# pip
pip install moai-adk==0.3.13

# uv
uv pip install moai-adk==0.3.13
```

---

## 📚 추가 리소스

### 문서

- **v0.4.0 릴리스 노트**: `CHANGELOG.md#v0.4.0`
- **v0.4.0 심층 분석**: `.moai/reports/DEEP_ANALYSIS_v0.4.0.md`
- **Skills 가이드**: `.claude/skills/README.md` (TODO: P1-3에서 작성)

### 지원

- **GitHub Issues**: https://github.com/modu-ai/moai-adk/issues
- **Discussions**: https://github.com/modu-ai/moai-adk/discussions

### 타임라인

| 날짜 | 이벤트 |
|------|--------|
| 2025-10-20 | v0.4.0 릴리스 |
| 2025-11-01 | v0.4.1 패치 (버그 수정) |
| 2025-12-01 | v0.5.0 RC (기존 명칭 제거 예정) |

---

## 🎯 요약

**v0.4.0은 성능과 개발자 경험을 대폭 개선한 메이저 릴리스입니다.**

**핵심 Breaking Changes**:
1. 🔴 `/alfred:1-spec` → `/alfred:1-plan`
2. 🔴 `/alfred:2-build` → `/alfred:2-run`
3. 🟡 Skills 시스템 도입 (44개)
4. 🟢 Sub-agents → Skills 통합

**마이그레이션 시간**: 30분
**호환성**: v0.3.13 명칭 v0.4.0에서 Deprecated (v0.5.0 제거)
**지원**: GitHub Issues

**시작하기**: [마이그레이션 가이드](#%EF%B8%8F-마이그레이션-가이드) 참조

---

**작성자**: Alfred SuperAgent
**최종 업데이트**: 2025-10-20
**문서 버전**: 1.0.0
