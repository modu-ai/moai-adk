# MoAI-ADK 제품 개요

> **최종 업데이트**: 2026-01-24
> **버전**: 4.1.0

---

## 개요

**MoAI-ADK** (Agentic Development Kit)는 SPEC-First 개발, DDD(도메인 주도 개발), AI 에이전트를 결합한 오픈소스 프레임워크로 완전하고 투명한 개발 수명주기를 제공합니다.

---

## 핵심 가치 제안

### 문제 정의

전통적인 개발 워크플로우의 문제점:
- 불명확한 요구사항으로 인한 빈번한 재작업
- 코드와 동기화되지 않는 문서화
- 지연된 테스트로 인한 품질 저하
- 반복적인 보일러플레이트 코드 작성

### 해결책

MoAI-ADK는 다음을 제공합니다:
- **SPEC-First 개발**: 명확한 SPEC 문서로 오해 해소
- **자동 문서화 동기화**: 모든 것을 자동으로 최신 상태 유지
- **DDD 강제화**: ANALYZE-PRESERVE-IMPROVE 사이클로 85%+ 테스트 커버리지 보장
- **AI 에이전트**: 20개 전문 에이전트로 반복 작업 자동화

---

## 주요 기능

### 1. SPEC-First 방법론
- 모든 개발은 명확한 명세서부터 시작
- EARS 형식으로 요구사항 문서화
- 요구사항 변경으로 인한 재작업 90% 감소

### 2. DDD 강제화
- manager-ddd 에이전트를 통한 자동화된 ANALYZE-PRESERVE-IMPROVE 사이클
- 85%+ 커버리지로 버그 70% 감소
- 행위 보존 테스트를 통한 안전한 리팩토링

### 3. Alfred SuperAgent 오케스트레이션
- Mr.Alfred가 20개 전문 AI 에이전트 지휘 (7-Tier 아키텍처)
- 세션당 평균 5,000 토큰 절약
- 수동 개발 대비 60-70% 시간 절약

### 4. 다국어 라우팅
- 4개 언어(EN/KO/JA/ZH) 자동 에이전트 선택
- XLT(Cross-Lingual Thought) 프로토콜로 의미적 매칭
- 에이전트 호출 100% 언어 커버리지

### 5. Ralph 엔진
- 통합 코드 품질 보증 시스템
- LSP(Language Server Protocol) 통합으로 실시간 진단
- AST-grep 기반 구조적 코드 분석
- `/moai:loop`, `/moai:fix` 명령으로 자동화된 피드백 루프

### 6. AST-Grep 통합
- 구조적 코드 검색, 보안 스캔, 리팩토링
- 40개 이상 프로그래밍 언어 지원
- 패턴 기반 코드 분석 (텍스트 기반 정규식 아님)

### 7. 자동 문서화
- `/moai:3-sync`를 통한 코드 변경 시 자동 문서 동기화
- 100% 문서화 최신성
- 수동 문서 작성 제거

### 8. Progressive Disclosure 시스템
- 3단계 지식 전달 (Quick, Implementation, Advanced)
- 초기 토큰 소비 67% 감소
- 필요할 때만 전체 스킬 콘텐츠 로딩

### 9. TRUST 5 품질
- Test, Readable, Unified, Secured, Trackable
- 엔터프라이즈급 품질 보증
- 배포 후 긴급 패치 99% 감소

### 10. 워크트리 병렬 개발
- Git worktree 기반 독립 작업 공간
- 여러 작업 동시 진행
- 컨텍스트 전환 없는 병렬 개발

### 11. MoAI Rank 리더보드
- 개발 기여도 시각화
- 품질 지표 추적
- 팀 협업 인센티브 (게이미피케이션)

### 12. 멀티레이어 모듈러 아키텍처
- CLI 레이어 (Click + Rich)
- 코어 레이어 (50개 이상 모듈)
- 파운데이션 레이어 (TRUST 5, Git)
- 품질 보증 레이어 (Ralph 엔진, LSP, AST-grep)
- 자동화 레이어 (루프 컨트롤러, 후크, 이벤트 시스템)
- Claude Code 통합 레이어 (20개 에이전트, 48개 스킬)

---

## 대상 사용자

### 주요 사용자
- Claude Code를 사용하는 소프트웨어 개발자
- DDD 실천을 도입하는 팀
- SPEC 기반 개발이 필요한 조직

### 사용 사례
- DDD를 통한 신규 프로젝트 개발
- AI 지원을 통한 기존 프로젝트 개선
- TRUST 5 프레임워크를 통한 품질 개선
- 다국어 팀 협업

---

## 차별화 요소

| 측면 | 전통적 방식 | MoAI-ADK |
|------|-------------|----------|
| 요구사항 명확성 | 임시 논의 | EARS 형식의 SPEC-First |
| 개발 방법론 | 종종 건너뛰거나 지연됨 | 시작부터 DDD 강제화 |
| 문서화 | 수동, 종종 구식 | 코드와 자동 동기화 |
| 코드 품질 | 가변적 | TRUST 5 + Ralph 엔진 검증 |
| 에이전트 지원 | 단일 목적 AI | 20개 전문 에이전트 |
| 토큰 효율성 | 최적화 없음 | Progressive Disclosure로 67% 절감 |

---

## 제품 메트릭

- **~71,800 LOC**: Python 코드 라인 수
- **20개 전문 에이전트**: Backend, Frontend, Database, Security, DDD 등
- **48개 스킬**: 도메인 지식 및 모범 사례
- **50개 이상 모듈**: 핵심 비즈니스 로직
- **4개 언어 지원**: English, Korean, Japanese, Chinese
- **85% 커버리지 목표**: 기본 테스트 커버리지 임계값
- **Python 3.11-3.14**: 지원되는 Python 버전
- **40개 언어 AST-Grep**: 구조적 코드 검색 지원
- **7-Tier 아키텍처**: Alfred 오케스트레이션 계층 구조
- **67% 토큰 절감**: Progressive Disclosure 시스템 효과

---

## 라이선스

MIT License - 오픈 소스

---

## 링크

- **저장소**: https://github.com/modu-ai/moai-adk
- **PyPI**: https://pypi.org/project/moai-adk/
- **이슈**: https://github.com/modu-ai/moai-adk/issues
