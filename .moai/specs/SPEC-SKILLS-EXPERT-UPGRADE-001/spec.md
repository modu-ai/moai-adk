---
id: SKILLS-EXPERT-UPGRADE-001
version: 1.0.0
status: draft
created: 2025-11-11
updated: 2025-11-11
author: @user
priority: high
category: enhancement
labels:
  - skills
  - expert-level
  - context7
  - web-search
  - documentation
scope:
  packages:
    - .claude/skills/
  files:
    - 186개 모든 SKILL.md 파일
---

# @SPEC:SKILLS-EXPERT-UPGRADE-001: 전문가 수준 스킬 업그레이드 (186개)

> **비전**: context7 MCP 도구와 웹 검색을 통한 186개 스킬의 전문가 수준 업그레이드
>
> **목표**: 모든 스킬을 공식 문서 기반 전문가 수준으로 업그레이드 (요약/생략 없음)
>
> **기간**: 3단계 병렬 실행 (각 스킬 2-3시간 소요)
>
> **영향도**: 전체 MoAI-ADK 생태계 (모든 사용자 및 개발자)

---

## EARS 요구사항 명세

### Ubiquitous Requirements (기본 요구사항)

- **U-1**: 시스템은 186개 스킬 각각을 **전문가 수준으로 완전히 재작성**해야 한다
- **U-2**: 시스템은 각 스킬에 **공식 문서와 베스트 프랙티스**를 통합해야 한다
- **U-3**: 시스템은 **절대 요약하거나 생략하지 않고** 모든 내용을 상세히 작성해야 한다
- **U-4**: 시스템은 각 스킬에 **최신 버전(2025년 기준) 정보**를 포함해야 한다
- **U-5**: 시스템은 **실제 코드 예제와 실용적인 패턴**을 모든 스킬에 포함해야 한다

### Event-driven Requirements (이벤트 기반)

- **E-1**: WHEN 스킬 업그레이드 시작 시, THEN **context7 MCP 도구로 공식 문서 검색**을 먼저 수행해야 한다
- **E-2**: WHEN 공식 문서 획득 시, THEN **웹 검색으로 최신 베스트 프랙티스**를 추가 조사해야 한다
- **E-3**: WHEN 문서 조사 완료 시, THEN **전문가 수준으로 스킬 완전 재작성**을 수행해야 한다
- **E-4**: WHEN 스킬 재작성 완료 시, THEN **품질 검증 및 다른 스킬과 통합성**을 확인해야 한다

### State-driven Requirements (상태 기반)

- **S-1**: WHILE 93개 템플릿 스킬 업그레이드 중일 때, 시스템은 **로컬 프로젝트 스킬도 동시에 업그레이드**해야 한다
- **S-2**: WHILE 특정 카테고리 스킬 업그레이드 중일 때, 시스템은 **관련 스킬들과의 통합성**을 유지해야 한다
- **S-3**: WHILE 업그레이드 진행 중일 때, 시스템은 **각 스킬의 독립성과 완결성**을 보장해야 한다

### Unwanted Behaviors (원치 않는 동작 방지)

- **UW-1**: 시스템은 **요약하거나 단순화하지 않아야 한다** - 모든 내용을 상세히 포함
- **UW-2**: 시스템은 **구식 정보나 deprecated 기능**을 포함하지 말아야 한다
- **UW-3**: 시스템은 **일관성 없는 포맷이나 스타일**을 사용하지 말아야 한다
- **UW-4**: 시스템은 **실제 불가능하거나 비현실적인 코드 예제**를 포함하지 말아야 한다
- **UW-5**: 시스템은 **스킬 간의 중복이나 모순**을 허용하지 말아야 한다

### Optional Features

- **O-1**: 시스템은 각 스킬에 **실제 프로젝트 예제 케이스**를 포함할 수 있다
- **O-2**: 시스템은 각 스킬에 **성능 벤치마크 데이터**를 포함할 수 있다
- **O-3**: 시스템은 각 스킬에 **일반적인 피트폴과 해결책**을 포함할 수 있다

---

## 상세 요구사항

### Phase 1: Foundation & Essential 스킬 업그레이드 (Tier 1-2)

#### F-1.1: Foundation 스킬 (10개) 전문가 수준 업그레이드

**대상 스킬**: moai-foundation-*, moai-essentials-*
- **moai-foundation-ears**: EARS v2.1 최신 명세 + 실제 요구사항 예제 20개
- **moai-foundation-specs**: YAML frontmatter 7필드 + 실제 SPEC 템플릿 15개
- **moai-foundation-tags**: @TAG 시스템 + 실제 추적 패턴 30개
- **moai-foundation-trust**: TRUST 5원칙 + 실제 검증 체크리스트 50개
- **moai-foundation-git**: GitFlow + 실제 커밋 메시지 예제 100개
- **moai-foundation-langs**: 언어 감지 + 24개 언어 signature
- **moai-essentials-debug**: 디버깅 패턴 + 에러 분석 케이스 50개
- **moai-essentials-perf**: 성능 최적화 + 벤치마크 예제 30개
- **moai-essentials-refactor**: 리팩토링 패턴 + 실제 적용 사례 40개
- **moai-essentials-review**: 코드 리뷰 + SOLID 위반 사례 60개

**업그레이드 기준**:
- 각 스킬 최소 2,000-5,000 words
- 실제 코드 예제 최소 30개
- 최신 2025년 기준 표준 및 베스트 프랙티스
- Industry-leading 프로젝트 사례 통합

#### F-1.2: Context7 기반 공식 문서 통합

**문서 소스**:
- Python: /python/cpython (최신 3.13+)
- FastAPI: /tiangolo/fastapi
- React: /facebook/react
- Docker: /docker/docker
- AWS: /aws/aws-sdk-js
- 기타: 각 기술의 공식 저장소

**통합 방법**:
```bash
# 각 스킬 업그레이드 프로세스
1. Context7으로 공식 문서 검색 (최소 5,000 tokens)
2. WebSearch로 2025년 베스트 프랙티스 조사
3. 실제 프로젝트 예제 분석
4. 전문가 수준으로 완전 재작성 (요약 없음)
```

### Phase 2: Language 스킬 업그레이드 (Tier 3)

#### F-2.1: 언어별 전문가 수준 업그레이드 (24개)

**대상 언어**: Python, TypeScript, JavaScript, Java, Go, Rust, C++, C#, Ruby, PHP, Swift, Kotlin, Dart, Scala, Shell, SQL, R, Julia, Lua, Elixir, Haskell, Clojure, C, F#

**각 스킬별 요구사항**:
- **언어 최신 버전**: 2025년 11월 기준 최신 안정 버전
- **생태계 현황**: 주요 라이브러리 및 프레임워크 (최소 20개)
- **개발 환경**: IDE, 빌드 도구, 테스트 프레임워크
- **성능 특성**: 벤치마크 데이터 및 최적화 기법
- **실제 프로젝트**: Production-ready 코드 예제 (최소 50개)
- **일반적인 패턴**: idiomatc 코드 및 베스트 프랙티스
- **피트폴 및 해결**: 흔한 실수와 해결책 (최소 30개)

**Python 스킬 예시 (moai-lang-python)**:
- Python 3.13+ 최신 기능 (typing improvements, error handling)
- FastAPI, Django, Flask 생태계
- Pydantic v2, SQLAlchemy 2.0, Celery
- uv, ruff, mypy 도구체인
- AI/ML 통합 (OpenAI, TensorFlow, PyTorch)
- 실제 API 서버 예제 (완전한 코드)
- 성능 최적화 및 프로파일링
- 배포 및 운영 (Docker, Kubernetes)

#### F-2.2: Progressive Disclosure 최적화

**로딩 전략**:
- 프로젝트 언어 자동 감지 시 해당 스킬만 로드
- 각 스킬은 독립적이면서도 상호 연동 가능
- 필요에 따라 관련 스킬 자동 제안

### Phase 3: Domain & 전문 스킬 업그레이드 (Tier 4)

#### F-3.1: 도메인별 전문가 수준 업그레이드 (9개)

**대상 도메인**:
- **Backend**: 마이크로서비스, API 디자인, 데이터베이스 패턴
- **Frontend**: React, Vue, Angular, 상태 관리, 성능
- **Database**: SQL, NoSQL, 데이터 모델링, 마이그레이션
- **DevOps**: CI/CD, Docker, Kubernetes, 모니터링
- **Security**: 인증/인가, 암호화, 보안 스캐닝
- **Web API**: REST, GraphQL, gRPC, API 문서화
- **CLI Tool**: 커맨드 라인 디자인, 인터페이스
- **Data Science**: 데이터 분석, 시각화, 머신러닝
- **Mobile App**: React Native, Flutter, 네이티브 개발

**각 도메인별 요구사항**:
- 아키텍처 패턴 및 디자인 원칙
- 주요 기술 스택 및 도구 (최소 15개)
- 실제 프로젝트 구조 및 코드 예제
- 성능, 확장성, 보안 고려사항
- 일반적인 문제 및 해결 패턴
- Industry standard 및 베스트 프랙티스

#### F-3.2: CC 스킬 및 BaaS 스킬 업그레이드

**CC 스킬 (Claude Code)**:
- Commands: 슬래시 커맨드 디자인 패턴
- Agents: 서브에이전트 오케스트레이션
- Skills: 스킬 제작 및 관리
- Hooks: 이벤트 기반 자동화
- Memory: 컨텍스트 관리 및 최적화
- Configuration: 설정 관리 및 검증

**BaaS 스킬 (Backend-as-a-Service)**:
- 9개 플랫폼: Supabase, Vercel, Neon, Clerk, Railway, Convex, Firebase, Cloudflare, Auth0
- 8개 아키텍처 패턴 및 실제 구현 예제
- 마이그레이션 전략 및 베스트 프랙티스
- 비용 분석 및 의사결정 프레임워크

### Phase 4: 품질 보증 및 통합

#### F-4.1: 스킬 품질 표준

**모든 스킬 공통 요구사항**:
- **길이**: 최소 1,500-5,000 words (요약 없음)
- **구조**: 일관된 섹션 구조 (What It Does, When to Use, Examples, Best Practices)
- **코드 예제**: 실행 가능하고 실용적인 예제 최소 20개
- **최신성**: 2025년 11월 기준 최신 정보
- **실용성**: 실제 프로젝트에서 즉시 적용 가능한 내용
- **완결성**: 해당 분야의 전문가 수준 지식 포함

#### F-4.2: 통합성 검증

**스킬 간 통합**:
- 상호 참조 및 연동 패턴
- 중복 최소화 및 시너지 최대화
- 일관된 용어 및 스타일 가이드
- 버전 호환성 및 의존성 관리

**품질 검증 체크리스트**:
- [ ] 공식 문서 기반 정확성
- [ ] 최신 버전 호환성
- [ ] 실제 실행 가능한 코드 예제
- [ ] 일관된 포맷 및 스타일
- [ ] 다른 스킬과의 통합성
- [ ] 실제 프로젝트 적용 가능성
- [ ] 전문가 수준 깊이의 내용

---

## 구현 전략

### 병렬 실행 계획

#### Team 1: Foundation & Essential (10개 스킬)
**담당**: foundation-specialist
**기간**: 2-3일
**우선순위**: 가장 높음 (모든 다른 스킬의 기반)

#### Team 2: Language Experts (24개 스킬)
**담당**: language-specialists (각 언어별 전문가)
**기간**: 5-7일 (병렬 처리)
**전략**: 언어 그룹별 동시 업그레이드

#### Team 3: Domain & Advanced (50+ 스킬)
**담당**: domain-experts (도메인별 전문가)
**기간**: 7-10일 (병렬 처리)
**전략**: 도메인별 동시 업그레이드

### 작업 프로세스

#### 각 스킬 업그레이드 프로세스
```yaml
1. 분석 단계 (30분):
   - 현재 스킬 상태 평가
   - context7로 공식 문서 검색
   - 웹 검색으로 베스트 프랙티스 조사

2. 설계 단계 (30분):
   - 전문가 수준 구조 설계
   - 실용적인 예제 및 케이스 선정
   - 다른 스킬과의 통합성 계획

3. 구현 단계 (1-2시간):
   - 전문가 수준으로 완전 재작성
   - 실제 코드 예제 개발 및 통합
   - 품질 검증 및 피드백 반영

4. 검증 단계 (30분):
   - 실행 가능성 확인
   - 다른 스킬과의 통합성 테스트
   - 전문가 레벨 검증
```

### 성공 측정 지표

#### 정량적 지표
- **스킬 업그레이드율**: 186개 중 186개 (100%)
- **평균 스킬 길이**: 2,000+ words (현재 500-1,000 words)
- **코드 예제 수**: 각 스킬 30+개 (현재 5-10개)
- **정확성**: 공식 문서 기반 100% 정확성

#### 정성적 지표
- **전문가 수준**: 해당 분야 전문가가 인정하는 수준
- **실용성**: 실제 프로젝트에 즉시 적용 가능
- **완결성**: 요약이나 생략 없는 완전한 내용
- **통합성**: 다른 스킬과의 원활한 연동

---

## 리스크 관리

### 고위험 요소
- **R-1**: 시간 소요 - 각 스킬당 예상보다 오래 걸릴 수 있음
- **R-2**: 품질 일관성 - 여러 전문가가 작성 시 스타일 불일치
- **R-3**: 기술 변화 - 업그레이드 중 최신 정보 변경 가능성

### 완화 전략
- **M-1**: 충분한 시간 할당 및 예상 시간의 150% 확보
- **M-2**: 명확한 가이드라인 및 템플릿 제공
- **M-3**: 지속적인 업데이트 프로세스 마련

---

## HISTORY

### v1.0.0 (2025-11-11)
- **INITIAL**: 186개 스킬 전문가 수준 업그레이드 SPEC 작성
- **AUTHOR**: @user
- **SCOPE**:
  - Foundation & Essential 스킬 (10개)
  - Language 스킬 (24개)
  - Domain 및 전문 스킬 (150+개)
  - Context7 MCP 도구 활용
  - 웹 검색 기반 베스트 프랙티스 통합
- **REQUIREMENTS**:
  - 모든 스킬 요약/생략 없이 완전한 전문가 수준으로 업그레이드
  - 공식 문서 기반 정확성 보장
  - 실제 코드 예제 및 실용적인 패턴 통합
  - 3단계 병렬 실행 전략
- **NEXT STEPS**:
  - 전문가 팀 구성 및 작업 분배
  - Context7 및 웹 검색 도구 세팅
  - Phase 1부터 순차적 실행

---

**작성**: @user (2025-11-11)
**상태**: draft (승인 대기)
**다음 단계**: /alfred:2-run SPEC-SKILLS-EXPERT-UPGRADE-001