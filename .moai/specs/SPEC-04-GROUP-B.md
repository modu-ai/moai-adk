---
spec_id: SPEC-04-GROUP-B
title: Phase 4 Skill Modularization - DOMAIN Skills (17개)
description: 17개 도메인 스킬의 포괄적인 모듈화 및 표준화
phase: Phase 4
category: DOMAIN
week: 5-6
status: PLANNED
priority: HIGH
owner: GOOS행
created_date: 2025-11-22
updated_date: 2025-11-22
total_skills: 17
modularization_target: 100%
---

# SPEC-04-GROUP-B: Phase 4 Skill Modularization - DOMAIN Skills

## Given (알려진 조건)

### 완료된 작업
- **Phase 1-3**: 15개 스킬 모듈화 완료
- **표준 모듈화 패턴**: 5개 파일 구조 확립
- **Context7 Integration**: 모든 스킬에 라이브러리 문서 링크
- **기준 날짜**: 2025-11-22 (최신 버전 기준)

### 대상 스킬 목록 (17개)

#### Architecture & Design
1. **moai-domain-backend** - REST API, 마이크로서비스, 확장성, 보안
2. **moai-domain-web-api** - HTTP, REST, GraphQL, API 설계 패턴
3. **moai-domain-cloud** - 클라우드 아키텍처, 멀티클라우드, 서비스 호환성
4. **moai-domain-devops** - CI/CD, 인프라 자동화, 배포 파이프라인
5. **moai-domain-database** - SQL/NoSQL, 스키마 설계, 최적화, 복제

#### Data & Computing
6. **moai-domain-ml-ops** - 모델 배포, A/B 테스팅, 모니터링
7. **moai-domain-iot** - IoT 장치, 실시간 데이터, 엣지 컴퓨팅
8. **moai-domain-monitoring** - 관찰성, 메트릭, 로깅, 추적

#### Tools & Utilities
9. **moai-domain-cli-tool** - CLI 설계, 아규먼트 파싱, 사용자 경험
10. **moai-domain-testing** - 테스트 전략, 테스트 피라미드, 자동화
11. **moai-domain-frontend** - UI/UX, 상태 관리, 성능, 접근성

#### Mobile & Specialized
12. **moai-domain-mobile-app** - 네이티브/크로스플랫폼, 성능, 배터리
13. **moai-domain-security** - OWASP, 암호화, 인증, 권한부여
14. **moai-domain-figma** - 디자인 시스템, 컴포넌트, 개발자 핸드오프

#### Content & Integration
15. **moai-domain-notion** - Notion API, 데이터베이스, 자동화
16. **moai-domain-toon** - 웹툰/만화 플랫폼, 콘텐츠 관리
17. **moai-domain-nano-banana** - 나노 스케일 프로젝트, 미니 MVP, 경량 솔루션

---

## When (실행 조건)

### 선행 조건
- SPEC-04-GROUP-A 완료 후 진행
- 남은 토큰 예산 충분함 (≥250K)
- Skill Factory 에이전트 활용 가능
- Context7 라이브러리 접근 가능

### 실행 가능한 경우
- 도메인별 관련 스킬을 묶어서 처리 가능
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

#### Session 1 (Week 5, 초반)
**대상**: Backend, Web-API, Cloud (아키텍처/설계)
- moai-domain-backend: 마이크로서비스, API 설계
- moai-domain-web-api: REST, GraphQL, API 패턴
- moai-domain-cloud: 멀티클라우드, 확장성

**예상 토큰**: 80-100K

#### Session 2 (Week 5, 후반)
**대상**: Database, DevOps, Monitoring (데이터/운영)
- moai-domain-database: SQL/NoSQL, 스키마 최적화
- moai-domain-devops: CI/CD, 자동화
- moai-domain-monitoring: 관찰성, 메트릭

**예상 토큰**: 80-100K

#### Session 3 (Week 6, 초반)
**대상**: ML-Ops, IoT, Testing (데이터/테스트)
- moai-domain-ml-ops: 모델 배포, 모니터링
- moai-domain-iot: IoT 장치, 실시간 데이터
- moai-domain-testing: 테스트 전략, 피라미드

**예상 토큰**: 80-100K

#### Session 4 (Week 6, 중반)
**대상**: CLI-Tool, Frontend, Mobile (도구/UI)
- moai-domain-cli-tool: CLI 설계, UX
- moai-domain-frontend: UI/UX, 상태 관리
- moai-domain-mobile-app: 크로스플랫폼, 성능

**예상 토큰**: 80-100K

#### Session 5 (Week 6, 후반)
**대상**: Security, Figma, Notion, Toon, Nano (보안/통합/특화)
- moai-domain-security: OWASP, 암호화
- moai-domain-figma: 디자인 시스템
- moai-domain-notion: 자동화, API
- moai-domain-toon: 콘텐츠 관리
- moai-domain-nano-banana: 경량 솔루션

**예상 토큰**: 80-100K

---

## Then (완료 기준)

### SPEC 완료 후 상태
- 17개 스킬 100% 모듈화 완료
- 각 스킬당 5개 파일 생성/수정
- 모든 검증 항목 완료

### 누적 진행률
```
Phase 4 Overall Progress:
- Week 1-2: 15개 (11.1%)
- Week 4-5: +18개 (GROUP-A) (13.3%)
- Week 5-6: +17개 (GROUP-B) (12.6%) ← 현재
- 누계: 50개 (37.0%)
- 남은 작업: 85개 (63.0%)
```

---

## 자동화 지시사항

### Skill Factory 배치 모듈화 명령어
```bash
# Session 1: Backend, Web-API, Cloud
Skill("moai-cc-skill-factory", {
  "action": "batch_modularize",
  "skills": [
    "moai-domain-backend",
    "moai-domain-web-api",
    "moai-domain-cloud"
  ],
  "target_version": "2025-11-22",
  "parallel": true,
  "context7_integration": true
})
```

### 도메인별 그룹 처리
- **Architecture**: backend, web-api, cloud, devops
- **Data**: database, ml-ops, iot, monitoring
- **Tools**: cli-tool, testing, figma, notion
- **Specialized**: security, mobile-app, toon, nano-banana, frontend

---

## 리소스 예산

| Session | 스킬 | 예상 토큰 |
|---------|-----|----------|
| Session 1 | Backend, Web-API, Cloud | 80-100K |
| Session 2 | Database, DevOps, Monitoring | 80-100K |
| Session 3 | ML-Ops, IoT, Testing | 80-100K |
| Session 4 | CLI-Tool, Frontend, Mobile | 80-100K |
| Session 5 | Security, Figma, Notion, Toon, Nano | 80-100K |
| **합계** | **17개** | **400-500K** |

---

**SPEC ID**: SPEC-04-GROUP-B
**생성일**: 2025-11-22
**상태**: PLANNED
**우선도**: HIGH
