---
spec_id: SPEC-04-GROUP-A
title: Phase 4 Skill Modularization - LANGUAGE Skills (18개)
description: 18개 프로그래밍 언어 스킬의 포괄적인 모듈화 및 표준화
phase: Phase 4
category: LANGUAGE
week: 4-5
status: PLANNED
priority: HIGH
owner: GOOS행
created_date: 2025-11-22
updated_date: 2025-11-22
total_skills: 18
modularization_target: 100%
---

# SPEC-04-GROUP-A: Phase 4 Skill Modularization - LANGUAGE Skills

## Given (알려진 조건)

### 완료된 작업
- **Phase 1-3**: 15개 스킬 모듈화 완료
  - Week 1: 10개 (ruby, php, scala, cpp, kotlin, html-css, rust, frontend, figma, monitoring)
  - Week 2: 5개 (go, java, javascript, typescript, python)
- **표준 모듈화 패턴** 확립됨:
  - SKILL.md: ≤400줄 (Quick Reference, 3-Level Learning Path, Best Practices)
  - examples.md: 550-700줄 (10-15개 실제 예제, Problem-Solution 형식)
  - modules/advanced-patterns.md: 400-500줄 (언어별 고급 패턴, 메타프로그래밍)
  - modules/optimization.md: 300-500줄 (성능 최적화, 메모리 관리)
- **Context7 Integration**: 모든 스킬에 라이브러리 문서 링크 포함
- **기준 날짜**: 2025-11-22 (최신 버전 기준)

### 대상 스킬 목록 (18개)

#### 기존 9개 (Phase 1-3에서 모듈화됨)
1. moai-lang-ruby (Week 1 완료)
2. moai-lang-php (Week 1 완료)
3. moai-lang-scala (Week 1 완료)
4. moai-lang-cpp (Week 1 완료)
5. moai-lang-kotlin (Week 1 완료)
6. moai-lang-html-css (Week 1 완료)
7. moai-lang-rust (Week 1 완료)
8. moai-lang-go (Week 2 완료)
9. moai-lang-java (Week 2 완료)

#### 신규 9개 (Phase 4에서 모듈화할 대상)
10. **moai-lang-c** - 정적 타입 시스템, 메모리 관리, C99/C11/C17 표준
11. **moai-lang-csharp** - OOP, LINQ, Async/Await, .NET 최신 버전
12. **moai-lang-dart** - 객체지향, Hot Reload, Flutter 통합
13. **moai-lang-elixir** - 함수형, 불변성, Erlang/OTP, 동시성
14. **moai-lang-r** - 데이터 분석, 통계, ggplot2, tidyverse
15. **moai-lang-shell** - 스크립팅, bash/zsh/fish, 시스템 자동화
16. **moai-lang-sql** - 데이터베이스 쿼리, 최적화, NoSQL vs SQL
17. **moai-lang-swift** - iOS/macOS 개발, SwiftUI, Concurrency
18. **moai-lang-tailwind-css** - Utility-first CSS, 디자인 시스템, Performance

---

## When (실행 조건)

### 선행 조건
- 남은 토큰 예산 충분함 (≥350K 예상)
- Skill Factory 에이전트 (`moai-cc-skill-factory`) 활용 가능
- 표준 모듈화 템플릿 재사용 가능
- Context7 라이브러리 접근 가능
- 2025-11-22 기준 최신 라이브러리 버전 정보 확인됨

### 실행 가능한 경우
- 단일 세션에서 1-2개 스킬 완료 가능
- 배치 스크립트로 자동화 가능
- 병렬 처리로 효율성 증대 가능

---

## What (명확한 요구사항)

### 각 스킬마다 생성/모듈화할 파일 구조

#### 1. SKILL.md (≤400줄)
**필수 섹션**:
- Quick Reference (30초 요약)
- What It Does / When to Use
- Three-Level Learning Path
  - Level 1: Fundamentals (examples.md 참조)
  - Level 2: Advanced Patterns (reference.md 참조)
  - Level 3: Production Deployment (전문 스킬 참조)
- Best Practices (DO/DON'T 리스트)
- Tool Versions (2025-11-22 기준)
- Installation & Setup (간단한 예제)
- Context7 Integration (관련 라이브러리 링크)

#### 2. examples.md (550-700줄)
**필수 사항**:
- 10-15개 실제 사용 예제
- Problem-Solution 형식
- 각 언어/도메인 특성 반영
- 단계별 복잡도 증가
- 실행 가능한 코드 (테스트 완료)
- 일반적인 함정과 해결책 포함

#### 3. modules/advanced-patterns.md (400-500줄)
**필수 사항**:
- 언어별/도메인별 고급 패턴
- 메타프로그래밍 기법
- 동시성 모델
- 성능 최적화 패턴
- 프로덕션 레디 코드
- 아키텍처 설계 패턴

#### 4. modules/optimization.md (300-500줄)
**필수 사항**:
- 성능 최적화 기법
- 메모리 관리
- 컴파일 최적화
- 프로파일링 및 튜닝
- 최적 실행 방법
- 일반적인 성능 함정

#### 5. reference.md (30-40줄)
- CLI 레퍼런스 또는 기존 파일 유지
- 주요 API 링크
- 공식 문서 참조

### 품질 기준

#### 파일 규격
- 모든 파일 마크다운 (.md) 형식
- UTF-8 인코딩
- 한국어/영문 혼합 가능
- Code block 포함 (문법 강조)

#### 내용 기준
- Context7 Integration 섹션 필수 포함
- 2025-11-22 최신 버전 정보 기재
- 모든 예제 실행 가능해야 함
- 안전성 및 보안 고려

#### 문서 규격
- YAML 헤더 포함 (메타데이터)
- 목차 자동 생성 가능하도록 구성
- 크로스-레퍼런스 명확
- 이미지/다이어그램 (선택사항, 복잡한 개념)

### 처리 순서 (우선도 기반)

#### Session 1 (Week 4, 초반)
**대상**: C, C#, Swift (정적 타입 시스템)
- moai-lang-c: 메모리 관리, 포인터, 구조체
- moai-lang-csharp: OOP, LINQ, Async/Await
- moai-lang-swift: iOS/macOS, SwiftUI, 안전성

**예상 토큰**: 80-100K
**체크포인트**: 3개 스킬 100% 완료

#### Session 2 (Week 4, 후반)
**대상**: Dart, Elixir, R (다양한 패러다임)
- moai-lang-dart: 객체지향, Hot Reload, Flutter
- moai-lang-elixir: 함수형, 불변성, 패턴 매칭
- moai-lang-r: 벡터화, 데이터프레임, 통계

**예상 토큰**: 80-100K
**체크포인트**: 6개 스킬 누계 완료

#### Session 3 (Week 5, 초반)
**대상**: Shell, SQL, Tailwind-CSS (시스템/쿼리/스타일)
- moai-lang-shell: bash/zsh/fish, 자동화, 스크립팅
- moai-lang-sql: 쿼리 최적화, 인덱싱, NoSQL
- moai-lang-tailwind-css: Utility-first, 성능, 디자인 시스템

**예상 토큰**: 80-100K
**체크포인트**: 9개 스킬 누계 완료

#### Session 4 (Week 5, 후반) - 예비
**대상**: 추가 스킬 또는 검수/개선
- 우선도 낮은 스킬 처리
- 기존 스킬 검수 및 버전 업데이트

**예상 토큰**: 40-50K

---

## Then (완료 기준)

### SPEC 완료 후 상태

#### 산출물
- 18개 스킬 100% 모듈화 완료
- 각 스킬당 5개 파일 생성/수정:
  - SKILL.md
  - examples.md
  - modules/advanced-patterns.md
  - modules/optimization.md
  - reference.md (기존 파일 유지 또는 신규 생성)

#### 검증 항목
- [ ] 모든 파일이 마크다운 형식 (.md)
- [ ] Context7 Integration 섹션 모두 포함
- [ ] 2025-11-22 버전 정보 정확성 확인
- [ ] 모든 코드 예제 실행 가능 확인
- [ ] 스킬별 파일 크기 범위 준수
- [ ] 문법 및 스타일 일관성 확인

#### 완료 체크리스트
- [ ] SPEC-04-GROUP-A 작업 완료
- [ ] 4 세션 모두 완료 (또는 계획된 세션 수)
- [ ] 각 스킬당 예상 크기 (총 ~1.5-2MB)
- [ ] 템플릿 동기화 완료
- [ ] Git commit 1개 작성
- [ ] 다음 그룹(GROUP-B) 진행 가능 상태

### 누적 진행률

```
Phase 4 Overall Progress:
- Week 1-2: 15개 스킬 완료 (11.1%)
- Week 4-5 (이 SPEC): +18개 스킬 (13.3%)
- 누계: 33개 스킬 (24.4%)
- 남은 작업: 102개 스킬 (75.6%)

Weekly Breakdown:
- Week 4: 6개 (3 sessions)
- Week 5: 6개 (3 sessions)
- Week 6-7+: 나머지 그룹 (B, C, D, E)
```

---

## 자동화 지시사항

### Skill Factory 사용 명령어

#### 명령어 1: 단일 스킬 모듈화
```bash
# 사용법
Skill("moai-cc-skill-factory", {
  "skill_name": "moai-lang-c",
  "action": "modularize",
  "target_version": "2025-11-22",
  "context7_integration": true,
  "files_to_create": ["SKILL.md", "examples.md", "modules/advanced-patterns.md", "modules/optimization.md"]
})
```

#### 명령어 2: 배치 모듈화 (여러 스킬)
```bash
# Session 1 배치 (C, C#, Swift)
Skill("moai-cc-skill-factory", {
  "action": "batch_modularize",
  "skills": ["moai-lang-c", "moai-lang-csharp", "moai-lang-swift"],
  "target_version": "2025-11-22",
  "parallel": true,
  "context7_integration": true
})
```

#### 명령어 3: 버전 검증
```bash
# 모든 스킬 버전 정보 확인
Skill("moai-cc-skill-factory", {
  "action": "validate_versions",
  "skills": ["moai-lang-c", "moai-lang-csharp", "moai-lang-dart", ...],
  "target_date": "2025-11-22"
})
```

### 수동 세션별 작업 흐름

#### Session 1 작업 프로세스
1. **준비**: C, C#, Swift 선택
2. **생성**: SKILL.md 작성 (각 ~300줄)
3. **예제**: examples.md 작성 (각 ~600줄)
4. **고급**: modules/advanced-patterns.md 작성
5. **최적화**: modules/optimization.md 작성
6. **검증**: 모든 코드 실행 확인
7. **통합**: Context7 라이브러리 링크 확인
8. **Commit**: 3개 스킬 완료 후 git commit

#### Session 간 체크포인트
- 이전 세션 3개 스킬 모두 완료 확인
- 기준 버전 정보 최신화
- Context7 링크 유효성 검증
- 템플릿 일관성 확인

---

## 리소스 예산

### 토큰 예산 분배

| Session | 스킬 | 예상 토큰 | 비고 |
|---------|-----|----------|------|
| Session 1 | C, C#, Swift | 80-100K | 정적 타입 시스템 |
| Session 2 | Dart, Elixir, R | 80-100K | 다양한 패러다임 |
| Session 3 | Shell, SQL, Tailwind | 80-100K | 시스템/쿼리/스타일 |
| Session 4 | 예비/검수 | 40-50K | 보충 및 개선 |
| **합계** | **18개** | **280-350K** | 예상 총액 |

### 세션별 시간 예상
- 각 Session: ~1.5-2시간 (3개 스킬)
- 총 기간: ~6-8주 (주 2-3 세션)
- 유연성: 병렬 처리 가능

---

## 추가 참고 사항

### Context7 라이브러리 통합
각 스킬의 SKILL.md에 포함될 Context7 라이브러리:

| 스킬 | Context7 라이브러리 | 우선 토픽 |
|-----|-------------------|---------|
| moai-lang-c | `/gnu-gcc/docs`, `/posix-standard/docs` | Memory Management, Pointers |
| moai-lang-csharp | `/microsoft/dotnet`, `/dotnet/csharp` | LINQ, Async/Await, Modern C# |
| moai-lang-dart | `/google/dart`, `/flutter/docs` | Dart Fundamentals, Hot Reload |
| moai-lang-elixir | `/elixir-lang/docs`, `/erlang/otp` | Pattern Matching, Concurrency |
| moai-lang-r | `/r-project/docs`, `/tidyverse/r4ds` | Data Analysis, ggplot2 |
| moai-lang-shell | `/gnu-bash/manual`, `/shell-scripting/best-practices` | Bash, Zsh, Fish |
| moai-lang-sql | `/mysql/docs`, `/postgresql/docs` | Query Optimization, Indexing |
| moai-lang-swift | `/apple/swift`, `/swift-org/docs` | SwiftUI, Concurrency Models |
| moai-lang-tailwind-css | `/tailwindlabs/tailwindcss` | Utility-First Design, Performance |

### 성공 지표 (TRUST 5 원칙)
- **Test**: 모든 코드 예제 실행 가능
- **Readable**: 명확한 구조, 주석 포함
- **Unified**: 모든 스킬 형식 일관성
- **Secured**: 보안 고려사항 포함
- **Trackable**: 버전 및 태그 관리

---

## 다음 단계

### SPEC-04-GROUP-A 완료 후
1. ✅ GROUP-A 모든 스킬 완료 확인
2. → `/clear` 명령으로 컨텍스트 리셋
3. → SPEC-04-GROUP-B 시작 (DOMAIN Skills)

### 전체 Phase 4 로드맵
```
Week 4-5: GROUP-A (LANGUAGE Skills, 18개) ← 현재
Week 5-6: GROUP-B (DOMAIN Skills, 17개)
Week 6-7: GROUP-C (Infrastructure Skills, 20개)
Week 7-8: GROUP-D (Platform/BaaS Skills, 10개)
Week 8-9: GROUP-E (Specialty Skills, 40+개)

최종 목표: 135개 모든 스킬 모듈화 완료 ✅
```

---

**SPEC ID**: SPEC-04-GROUP-A
**생성일**: 2025-11-22
**상태**: PLANNED
**우선도**: HIGH
**기대 효과**: 18개 프로그래밍 언어 스킬 100% 모듈화로 개발자 경험 향상
