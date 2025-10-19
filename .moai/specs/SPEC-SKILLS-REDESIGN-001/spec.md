---
id: SKILLS-REDESIGN-001
version: 0.1.0
status: completed
created: 2025-10-19
updated: 2025-10-20
author: @Alfred
priority: high
category: refactor
labels:
  - skills
  - architecture
  - progressive-disclosure
scope:
  packages:
    - .claude/skills/
  files:
    - 44개 모든 SKILL.md
---

# @SPEC:SKILLS-REDESIGN-001: Skills 4-Tier 아키텍처 재설계

> **비전**: "Layered Skills Architecture"를 통한 Progressive Disclosure 최적화
>
> **목표**: 46개 → 44개 스킬 (2개 삭제), 10개 재구성
>
> **기간**: 4주 (1주당 1-2개 Phase)
>
> **영향도**: 모든 MoAI-ADK 사용자 (전역 Skills 재구성)

---

## EARS 요구사항 명세

### Ubiquitous Requirements (기본 요구사항)

- **U-1**: 시스템은 44개 스킬을 **4개 Tier로 명확하게 계층화**해야 한다
- **U-2**: 시스템은 모든 스킬의 SKILL.md를 **<500 words로 표준화**해야 한다
- **U-3**: 시스템은 모든 스킬에 **allowed-tools 필드를 명시**해야 한다
- **U-4**: 시스템은 각 스킬에 **"Works well with" 섹션**을 포함해야 한다

### Event-driven Requirements (이벤트 기반)

- **E-1**: WHEN `/alfred:1-plan` 실행 시, THEN **Tier 1 (Foundation) 스킬만 로드**해야 한다
- **E-2**: WHEN `/alfred:2-run` 실행 시, THEN **프로젝트 언어 감지 후 해당 Tier 3 스킬만 로드**해야 한다
- **E-3**: WHEN `/alfred:3-sync` 실행 시, THEN **Tier 1의 모든 스킬이 함께 작동**해야 한다
- **E-4**: WHEN 사용자가 "백엔드 API 설계" 요청 시, THEN **Tier 4 (moai-domain-backend)만 로드**해야 한다

### State-driven Requirements (상태 기반)

- **S-1**: WHILE 프로젝트가 Python인 상태일 때, 시스템은 **moai-lang-python만 로드**해야 한다 (다른 23개 언어는 로드하지 않음)
- **S-2**: WHILE Foundation 스킬이 실행 중일 때, 시스템은 **Tier 2 Essentials 스킬도 사용 가능**해야 한다
- **S-3**: WHILE 사용자가 디버깅 중일 때, 시스템은 **moai-essentials-debug (Tier 2) 자동 호출**할 수 있어야 한다

### Constraints Requirements (제약사항)

- **C-1**: Tier 1 (Foundation) 6개 스킬의 SKILL.md 총 크기는 **≤500줄**이어야 한다
- **C-2**: 각 Tier 3 (Language) 스킬의 SKILL.md는 **≤100줄**이어야 한다
- **C-3**: 각 Tier 4 (Domain) 스킬의SKILL.md는 **≤100줄**이어야 한다
- **C-4**: 모든 스킬의 description은 **≤200 characters**여야 한다
- **C-5**: 삭제되는 2개 스킬 (template-generator, feature-selector)의 기능은 **다른 컴포넌트로 흡수**되어야 한다

### Optional Features

- **O-1**: 스킬 생성 시 **스킬 간 의존성 그래프 생성**할 수 있다
- **O-2**: 스킬 선택 시 **추천 스킬 자동 제시**할 수 있다
- **O-3**: 스킬 업데이트 시 **자동 버전 관리**할 수 있다

---

## 상세 요구사항

### Tier 1: Foundation Skills (6개)

#### F-1.1: moai-foundation-trust (기존: trust-validation)
- **기능**: TRUST 5원칙 (Test/Readable/Unified/Secured/Trackable) 검증
- **트리거**: `/alfred:3-sync` 자동 호출
- **Output**: quality-report.md
- **Works with**: moai-foundation-tags, moai-foundation-specs

#### F-1.2: moai-foundation-tags (기존: tag-scanning)
- **기능**: @TAG 직접 스캔, 인벤토리 생성 (CODE-FIRST)
- **트리거**: `/alfred:3-sync` 자동 호출
- **Output**: tag-inventory.md
- **Works with**: moai-foundation-trust, moai-foundation-specs

#### F-1.3: moai-foundation-specs (기존: spec-metadata-validation)
- **기능**: SPEC YAML frontmatter 7개 필수 필드 검증
- **트리거**: `/alfred:1-plan` 자동 호출
- **Output**: validation-report.md
- **Works with**: moai-foundation-ears, moai-foundation-tags

#### F-1.4: moai-foundation-ears (기존: ears-authoring)
- **기능**: EARS 요구사항 작성 가이드 (Ubiquitous/Event/State/Optional/Constraints)
- **트리거**: SPEC 작성 시 호출
- **Output**: SPEC 템플릿
- **Works with**: moai-foundation-specs

#### F-1.5: moai-foundation-git (기존: git-workflow)
- **기능**: Git 브랜치/커밋/PR 자동화 (TDD 커밋: 🔴→🟢→♻️)
- **트리거**: 모든 Alfred 커맨드
- **Output**: Git 작업 로그
- **Works with**: 모든 Tier

#### F-1.6: moai-foundation-langs (기존: language-detection)
- **기능**: 언어/프레임워크 자동 감지 (package.json, pyproject.toml 등)
- **트리거**: `/alfred:2-run` 실행 시
- **Output**: 감지된 언어 정보
- **Triggers**: Tier 3 Language 스킬

### Tier 2: Developer Essentials (4개)

#### F-2.1: moai-essentials-review (기존: code-reviewer)
- **기능**: SOLID 원칙, 코드 스멜, 언어별 Best Practice 자동 리뷰
- **트리거**: "코드 리뷰", `/alfred:3-sync` 후
- **Output**: code-review-report.md
- **Works with**: moai-foundation-specs, moai-essentials-refactor

#### F-2.2: moai-essentials-debug (기존: debugger-pro)
- **기능**: 스택 트레이스 분석, 에러 패턴 탐지, 수정 제안
- **트리거**: "에러 해결", NullPointerException 등
- **Output**: debug-analysis.md
- **Works with**: Tier 3 Language 스킬

#### F-2.3: moai-essentials-refactor (기존: refactoring-coach)
- **기능**: 리팩토링 가이드 (Extract Method, Replace Conditional, 디자인 패턴)
- **트리거**: "리팩토링", "개선", `/alfred:3-sync` 후
- **Output**: refactoring-plan.md
- **Works with**: moai-essentials-review, moai-foundation-specs

#### F-2.4: moai-essentials-perf (기존: performance-optimizer)
- **기능**: 성능 분석, 병목지점 탐지, 최적화 제안 (Big-O, 캐싱 전략)
- **트리거**: "느려요", "성능 최적화", "프로파일링"
- **Output**: performance-report.md
- **Works with**: Tier 3 Language 스킬, moai-essentials-refactor

### Tier 3: Language Experts (24개) - 유지

#### F-3.1: Progressive Disclosure 메커니즘
- moai-foundation-langs가 프로젝트 언어 감지
- 감지된 언어 스킬만 자동 로드 (예: Python → moai-lang-python)
- 나머지 23개 언어 스킬은 로드하지 않음 → **토큰 비용 0**

#### F-3.2: 24개 언어 스킬
```
Python, TypeScript, JavaScript, Java, Go, Rust, Ruby,
Dart, Swift, Kotlin, Scala, Clojure, Elixir, Haskell,
C, C++, C#, PHP, Lua, Shell, SQL, Julia, R
```

#### F-3.3: 각 스킬 내용
- 언어별 테스트 프레임워크 (pytest, vitest, go test 등)
- 린터/포매터 (ruff, biome, gofmt 등)
- 타입 검사 (mypy, tsc 등)
- TRUST 5원칙 구현 방법

### Tier 4: Domain Experts (9개) - 유지

#### F-4.1: 9개 도메인 스킬
```
Backend, Frontend, Database, DevOps, Web API,
Security, CLI Tool, Data Science, ML, Mobile App
```

#### F-4.2: 선택적 로드 메커니즘
- 사용자 요청 기반 로드 (예: "백엔드 API" → moai-domain-backend)
- 대부분 상호 배타적 (Backend ≠ Mobile App)

#### F-4.3: 각 스킬 내용
- 도메인 아키텍처 패턴
- 디자인 패턴 및 Best Practice
- 성능/보안/스케일링 고려사항

### Claude Code Skill (1개)

#### F-5.1: moai-claude-code (유지)
- 5개 컴포넌트 템플릿 (Agent, Command, Skill, Plugin, Settings)
- 기존 구조 유지

---

## 삭제 대상 (2개)

### D-1: moai-alfred-template-generator
- **이유**: moai-claude-code 스킬로 기능 흡수 (templates/ 디렉토리)
- **대체**: moai-claude-code/templates/

### D-2: moai-alfred-feature-selector
- **이유**: Commands로 기능 흡수 (/alfred:1-plan 내부 로직)
- **대체**: Commands 워크플로우

---

## 네이밍 컨벤션

### Tier 1-2: Alfred 스킬 재명명

| 현재명 | 재설계명 | Tier |
|--------|---------|------|
| moai-alfred-trust-validation | moai-foundation-trust | 1 |
| moai-alfred-tag-scanning | moai-foundation-tags | 1 |
| moai-alfred-spec-metadata-validation | moai-foundation-specs | 1 |
| moai-alfred-ears-authoring | moai-foundation-ears | 1 |
| moai-alfred-git-workflow | moai-foundation-git | 1 |
| moai-alfred-language-detection | moai-foundation-langs | 1 |
| moai-alfred-code-reviewer | moai-essentials-review | 2 |
| moai-alfred-debugger-pro | moai-essentials-debug | 2 |
| moai-alfred-refactoring-coach | moai-essentials-refactor | 2 |
| moai-alfred-performance-optimizer | moai-essentials-perf | 2 |

### Tier 3-4: 기존 네이밍 유지

```
Tier 3: moai-lang-{language}
Tier 4: moai-domain-{domain}
```

---

## HISTORY

### v0.1.0 (2025-10-20)
- **COMPLETED**: Skills 4-Tier 아키텍처 구현 완료
- **AUTHOR**: @Alfred
- **CHANGES**:
  - 모든 스킬 재구성: 46개 → 44개 (2개 삭제)
  - Tier 1: Foundation (6개) - 명명 및 구조 완성
  - Tier 2: Essentials (4개) - 용도별 스킬 재조직
  - Tier 3: Language (24개) - 언어별 전문 스킬 유지
  - Tier 4: Domain (9개) - 도메인별 전문 스킬 유지
  - Claude Code Skill (1개) - 템플릿 구조 유지
  - 모든 스킬 SKILL.md 표준화 (<500 words)
  - allowed-tools 필드 모든 스킬에 추가
  - "Works well with" 섹션 모든 스킬에 추가
  - Progressive Disclosure 메커니즘 구현
  - 테스트 작성 및 통과
  - 문서 동기화 완료

### v0.0.1 (2025-10-19)
- **INITIAL**: MoAI-ADK Skills 4-Tier 아키텍처 재설계 명세 작성
  - Tier 1: Foundation (6개)
  - Tier 2: Essentials (4개)
  - Tier 3: Language (24개)
  - Tier 4: Domain (9개)
  - Claude Code (1개)
  - 삭제: 2개
- EARS 기반 40+ 요구사항 정의
- Progressive Disclosure 및 Anthropic 원칙 준수 명시
- UPDATE-PLAN-0.4.0.md 철학 통합

---

**완료**: SPEC-SKILLS-REDESIGN-001 (v0.1.0)
**작성자**: @Alfred SuperAgent
**상태**: 구현 완료, 문서 동기화 완료
