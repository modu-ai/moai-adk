# 🎓 MoAI-ADK Skills Management

> **Skill 시스템의 3단계 구조** - JIT Retrieval을 통한 컨텍스트 최적화

---

## 📋 디렉토리 구조

```
.claude/skills/
├─ active/          # 활성화된 Skill (기본값: 7개 Foundation)
│  └─ moai-foundation-*/ (7개)
├─ available/       # 전체 가능한 Skill (57개)
│  ├─ moai-foundation-*/ (7개)
│  ├─ moai-lang-*/ (20개)
│  ├─ moai-domain-*/ (10개)
│  └─ moai-essentials-*/ (13개)
├─ archived/        # 사용 중단된 Skill (선택)
└─ README.md        # 이 파일
```

---

## 🎯 Skill 선택 정책

### Default: Foundation Skills (7개)

프로젝트 시작 시 **기본 활성화** Skill:

1. **moai-foundation-specs** - SPEC 메타데이터 검증
2. **moai-foundation-ears** - EARS 요구사항 작성
3. **moai-foundation-tags** - @TAG 시스템 관리
4. **moai-foundation-trust** - TRUST 5원칙 검증
5. **moai-foundation-langs** - 언어 감지
6. **moai-foundation-git** - Git 워크플로우

```
목적: 새로운 프로젝트를 위한 최소한의 필수 Skill
효과: 컨텍스트 로드 57개 → 7개 (87% 감소)
```

### Custom: 프로젝트별 선택 (3~9개)

`/alfred:0-project` 또는 수동으로 선택:

```bash
# Example: Python + Backend 프로젝트
moai-foundation-specs
moai-foundation-ears
moai-foundation-tags
moai-foundation-trust
moai-lang-python       # ← 언어별 추가
moai-domain-backend    # ← 도메인별 추가
moai-essentials-debug

# 7개 선택 (최적)
```

---

## 🔄 Skill 활성화 방법

### 1. 자동 활성화 (권장)
```bash
/alfred:0-project
→ feature-selector 에이전트가 3~9개 최적 선택
→ 자동으로 active/ 에 심볼릭 링크 생성
```

### 2. 수동 활성화
```bash
# 특정 Skill 활성화
ln -s ../available/moai-lang-python active/

# 전체 Skill 활성화 (개발 용도)
ln -s available/* active/
```

### 3. Skill 비활성화
```bash
# Skill 제거
rm active/moai-lang-go

# 또는 archived/ 로 이동
mv active/moai-lang-go archived/
```

---

## 📊 Skill Tier 구조

### Tier 1: Foundation (7개) ⭐
기초 구성요소, 모든 프로젝트에 필수:
- specs, ears, tags, trust, langs, git, (essentials-debug)

### Tier 2: Languages (20개) 🎨
언어별 TDD 도구:
- Python, TypeScript, Java, Go, Rust, Ruby, Kotlin, Swift, Dart 등

### Tier 3: Domains (10개) 🏗️
도메인별 전문성:
- Backend, Frontend, ML, Mobile, Security, Database, DevOps 등

### Tier 4: Essentials (13개) 🔧
필수 도구:
- Debug, Refactor, Review, Perf, Git 등

---

## 🚀 성능 최적화 가이드

### 컨텍스트 로드 비용

| 상태 | Skill 개수 | 로드 크기 | 응답 속도 | 추천 상황 |
|------|-----------|---------|---------|---------|
| **Minimal** | 3-4개 | ~50KB | 매우 빠름 | 신규 프로젝트 |
| **Optimal** | 5-9개 | ~100-150KB | 빠름 | **권장 (기본값)** |
| **Extended** | 10-20개 | ~200-300KB | 보통 | 중간 프로젝트 |
| **Full** | 57개 | ~500-600KB | 느림 | 개발/테스트 |

### 권장 설정

```json
// .claude/settings.json (추가 항목)
{
  "skills": {
    "mode": "auto",              // auto|manual
    "active-count": 7,           // 기본 활성 개수
    "max-count": 9,              // 최대 활성 개수
    "cache-enabled": true,       // Skill 메타데이터 캐싱
    "lazy-load": true            // 필요 시에만 로드
  }
}
```

---

## 🔍 Skill 발견

### 전체 Skill 목록 조회
```bash
ls -la .claude/skills/available/
```

### 특정 언어 Skill
```bash
ls .claude/skills/available/ | grep moai-lang-
```

### 특정 도메인 Skill
```bash
ls .claude/skills/available/ | grep moai-domain-
```

---

## 📝 Skill 관리 체크리스트

### 프로젝트 초기화
- [ ] `/alfred:0-project` 실행
- [ ] 언어/도메인 선택
- [ ] active/ 에 Skill 심볼릭 링크 생성
- [ ] 확인: `ls active/` (3-9개)

### 주기적 검토
- [ ] 월 1회: 사용되지 않는 Skill 식별
- [ ] 분기별: active/ 최적화
- [ ] 매년: 전체 Skill 업그레이드 확인

### 문제 해결
- [ ] Skill 로드 실패 → 메타데이터 검증 (SKILL.md)
- [ ] Skill 충돌 → `rg "name:" active/*/SKILL.md` 중복 확인
- [ ] 성능 저하 → active/ 개수 줄이기

---

## 🎯 다음 단계

### Immediate (지금)
- [x] Skills 디렉토리 구조 생성
- [ ] active/ 에 Foundation Skills 설정
- [ ] README.md 작성 ✓

### Short-term (1주)
- [ ] `.claude/settings.json` 에 skills 설정 추가
- [ ] Feature-selector 에이전트 최적화
- [ ] Skill 캐싱 구현

### Long-term (1달)
- [ ] Skill 활용도 분석 (로깅)
- [ ] 미사용 Skill 식별 및 아카이브
- [ ] Skill 성능 모니터링 대시보드

---

**관련 문서**:
- [SKILL_INTEGRATION_TEST_REPORT.md](../../SKILL_INTEGRATION_TEST_REPORT.md)
- [.claude/settings.json](./../settings.json)
- [.moai/memory/development-guide.md](../../.moai/memory/development-guide.md)

**작성자**: @agent-cc-manager
**최종 업데이트**: 2025-10-20
