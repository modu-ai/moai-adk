---
spec_id: SPEC-04-GROUP-C
title: Phase 4 Skill Modularization - Infrastructure Skills (20개)
description: 20개 인프라 및 핵심 기술 스킬의 포괄적인 모듈화 및 표준화
phase: Phase 4
category: INFRASTRUCTURE
week: 6-7
status: PLANNED
priority: HIGH
owner: GOOS행
created_date: 2025-11-22
updated_date: 2025-11-22
total_skills: 20
modularization_target: 100%
---

# SPEC-04-GROUP-C: Phase 4 Skill Modularization - Infrastructure Skills

## Given (알려진 조건)

### 완료된 작업
- **Phase 1-3**: 15개 스킬 모듈화 완료
- **표준 모듈화 패턴**: 5개 파일 구조 확립
- **Context7 Integration**: 모든 스킬에 라이브러리 문서 링크
- **기준 날짜**: 2025-11-22 (최신 버전 기준)

### 대상 스킬 목록 (20개)

#### Core Architecture (5개)
1. **moai-core-agent-factory** - 에이전트 생성, 설정, 배포
2. **moai-core-workflow** - 워크플로우 엔진, 오케스트레이션
3. **moai-core-config-schema** - 설정 스키마, 검증, 로딩
4. **moai-core-context-budget** - 토큰 예산 관리, 최적화
5. **moai-core-expertise-detection** - 사용자 전문성 감지, 개인화

#### Foundation (5개)
6. **moai-foundation-git** - Git 워크플로우, 브랜치 전략
7. **moai-foundation-specs** - SPEC 라이프사이클, 상태 관리
8. **moai-foundation-trust** - TRUST 5 원칙, 품질 게이트
9. **moai-foundation-ears** - EARS 패턴, 요구사항 엔지니어링
10. **moai-foundation-langs** - 언어 감지, 다국어 지원

#### Claude Code (5개)
11. **moai-cc-skill-factory** - 스킬 생성, 모듈화, 자동화
12. **moai-cc-commands** - MoAI 커맨드 구현, 실행
13. **moai-cc-configuration** - Claude Code 설정, 최적화
14. **moai-cc-memory** - 세션 메모리, 상태 추적
15. **moai-cc-hooks** - 라이프사이클 훅, 이벤트 처리

#### Essentials (5개)
16. **moai-essentials-debug** - 디버깅 기법, 도구, 자동화
17. **moai-essentials-perf** - 성능 최적화, 프로파일링, 튜닝
18. **moai-essentials-refactor** - 리팩토링 패턴, 안전 전략
19. **moai-essentials-review** - 코드 리뷰, 품질 검사
20. **moai-core-dev-guide** - 개발 가이드, Best Practices, 참조

---

## When (실행 조건)

### 선행 조건
- SPEC-04-GROUP-A, GROUP-B 완료 후 진행
- 남은 토큰 예산 충분함 (≥200K)
- Skill Factory 에이전트 활용 가능
- Context7 라이브러리 접근 가능

### 실행 가능한 경우
- 그룹별 관련 스킬을 묶어서 처리 가능
- 병렬 처리로 효율성 증대 가능

---

## What (명확한 요구사항)

### 각 스킬마다 생성할 파일 구조
- SKILL.md (≤400줄): 개요, 3단계 학습 경로, Best Practices
- examples.md (550-700줄): 10-15개 실제 예제
- modules/advanced-patterns.md (400-500줄): 고급 패턴 및 아키텍처
- modules/optimization.md (300-500줄): 성능 최적화 기법
- reference.md (30-40줄): 주요 링크 및 문서

### 품질 기준
- 모든 파일 마크다운 (.md) 형식
- Context7 Integration 섹션 필수
- 2025-11-22 최신 버전 정보 기재
- 모든 예제 실행 가능해야 함

### 처리 순서 (우선도 기반)

#### Session 1 (Week 6, 후반)
**대상**: Core Architecture (에이전트, 워크플로우, 설정)
- moai-core-agent-factory: 에이전트 생성, 설정
- moai-core-workflow: 워크플로우 엔진
- moai-core-config-schema: 설정 스키마

**예상 토큰**: 80-100K

#### Session 2 (Week 7, 초반)
**대상**: Foundation (Git, SPECS, TRUST, EARS)
- moai-foundation-git: Git 워크플로우
- moai-foundation-specs: SPEC 관리
- moai-foundation-trust: TRUST 5 원칙
- moai-foundation-ears: EARS 패턴

**예상 토큰**: 100-120K

#### Session 3 (Week 7, 중반)
**대상**: Claude Code (스킬 팩토리, 커맨드, 메모리)
- moai-cc-skill-factory: 스킬 자동화
- moai-cc-commands: 커맨드 구현
- moai-cc-configuration: 설정 최적화
- moai-cc-memory: 세션 메모리

**예상 토큰**: 100-120K

#### Session 4 (Week 7, 후반)
**대상**: Essentials (디버깅, 성능, 리팩토링)
- moai-essentials-debug: 디버깅 기법
- moai-essentials-perf: 성능 최적화
- moai-essentials-refactor: 리팩토링 패턴
- moai-essentials-review: 코드 리뷰
- moai-core-dev-guide: 개발 가이드

**예상 토큰**: 100-120K

---

## Then (완료 기준)

### SPEC 완료 후 상태
- 20개 스킬 100% 모듈화 완료
- 각 스킬당 5개 파일 생성/수정
- 모든 검증 항목 완료

### 누적 진행률
```
Phase 4 Overall Progress:
- Week 1-2: 15개 (11.1%)
- Week 4-5: +18개 (GROUP-A) (13.3%)
- Week 5-6: +17개 (GROUP-B) (12.6%)
- Week 6-7: +20개 (GROUP-C) (14.8%) ← 현재
- 누계: 70개 (51.9%)
- 남은 작업: 65개 (48.1%)
```

---

## 자동화 지시사항

### Skill Factory 배치 모듈화 명령어
```bash
# Session 1: Core Architecture
Skill("moai-cc-skill-factory", {
  "action": "batch_modularize",
  "skills": [
    "moai-core-agent-factory",
    "moai-core-workflow",
    "moai-core-config-schema"
  ],
  "target_version": "2025-11-22",
  "category": "core-architecture",
  "parallel": true
})

# Session 2: Foundation
Skill("moai-cc-skill-factory", {
  "action": "batch_modularize",
  "skills": [
    "moai-foundation-git",
    "moai-foundation-specs",
    "moai-foundation-trust",
    "moai-foundation-ears"
  ],
  "target_version": "2025-11-22",
  "category": "foundation",
  "parallel": true
})
```

---

## 리소스 예산

| Session | 그룹 | 스킬 수 | 예상 토큰 |
|---------|-----|--------|----------|
| Session 1 | Core Architecture | 3 | 80-100K |
| Session 2 | Foundation | 4-5 | 100-120K |
| Session 3 | Claude Code | 4-5 | 100-120K |
| Session 4 | Essentials | 5-6 | 100-120K |
| **합계** | **4개 그룹** | **20개** | **380-460K** |

---

**SPEC ID**: SPEC-04-GROUP-C
**생성일**: 2025-11-22
**상태**: PLANNED
**우선도**: HIGH
