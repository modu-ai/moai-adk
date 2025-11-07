# MoAI-ADK: AI 기반 SPEC-First TDD 개발 프레임워크

> **신뢰할 수 있고 유지보수하기 쉬운 소프트웨어를 AI의 도움으로 빌드하세요.** 요구사항부터 문서화까지 모든 산출물이 완벽하게 추적되고, 자동으로 테스트되며, 항상 동기화됩니다.

---

## 🎯 우리가 해결하는 문제

### 기존 AI 기반 개발의 6가지 문제

| 문제 | 영향 |
|------|------|
| **모호한 요구사항** | 개발자가 40% 시간을 요구사항 명확화에 사용 |
| **부족한 테스트** | 테스트되지 않은 코드로 인한 프로덕션 버그 |
| **동기화되지 않는 문서** | 구현과 맞지 않는 문서 |
| **잃어버린 컨텍스트** | 팀원들 간 반복적인 설명 필요 |
| **불가능한 영향 분석** | 요구사항 변경 시 영향받는 코드 파악 불가 |
| **일관성 없는 품질** | 수동 QA로 인한 엣지 케이스 누락 |

### MoAI-ADK의 해결책

✅ **SPEC-First**: 코드 작성 전 명확한 요구사항 정의
✅ **보증된 테스트**: 자동 TDD를 통해 87%+ 테스트 커버리지 달성
✅ **살아있는 문서**: 자동 동기화되어 절대 떨어지지 않는 문서
✅ **지속적인 컨텍스트**: Alfred가 프로젝트 이력과 패턴을 기억
✅ **완전한 추적성**: `@TAG` 시스템으로 모든 산출물 연결
✅ **품질 자동화**: TRUST 5 원칙을 자동으로 강제

---

## ⚡ 핵심 기능

### 1. SPEC-First 개발
- **EARS 형식 명세서**: 구조화되고 명확한 요구사항
- **구현 전 명확화**: 비용이 큰 재작업 방지
- **자동 추적성**: 요구사항에서 코드, 테스트까지 연결

### 2. 자동화된 TDD 워크플로우
- **RED → GREEN → REFACTOR** 사이클 자동 관리
- **테스트 우선 보증**: 테스트 없는 코드는 없음
- **87%+ 커버리지**: 체계적 테스팅으로 달성

### 3. Alfred 슈퍼에이전트
- **19개의 전문 AI 에이전트** (spec-builder, tdd-implementer, doc-syncer 등)
- **69개 이상의 프로덕션급 스킬** (모든 개발 영역 커버)
- **적응형 학습**: 프로젝트 패턴으로부터 자동 학습
- **스마트 컨텍스트 관리**: 프로젝트 구조와 의존성 이해

### 4. @TAG 시스템 (완전한 추적성)

모든 산출물을 연결하는 추적성 시스템:

```
@SPEC:AUTH-001 (요구사항)
    ↓
@TEST:AUTH-001 (테스트)
    ↓
@CODE:AUTH-001:SERVICE (구현)
    ↓
@DOC:AUTH-001 (문서)
```

### 5. 살아있는 문서
- **실시간 동기화**: 코드와 문서가 항상 일치
- **수동 업데이트 불필요**: 자동 생성
- **다중언어 지원**: Python, TypeScript, Go, Rust 등
- **자동 다이어그램 생성**: 코드 구조에서 자동 생성

### 6. 품질 보증
- **TRUST 5 원칙**: Test-first, Readable, Unified, Secured, Trackable
- **자동화된 품질 게이트** (린팅, 타입 체크, 보안 검사)
- **Pre-commit 검증**: 위반 사항 사전 차단
- **종합 리포팅**: 실행 가능한 메트릭

---

## 🚀 빠른 시작

### 설치 (권장: uv tool)

```bash
# uv tool을 사용하여 moai-adk를 전역 명령어로 설치
uv tool install moai-adk

# 설치 확인
moai-adk --version

# 새 프로젝트 초기화
moai-adk init my-awesome-project
cd my-awesome-project
```

### 프로젝트 구성 (필수)

설치 후 **반드시** 프로젝트를 구성해야 합니다:

```bash
# 프로젝트 메타데이터 및 환경 초기화
/alfred:0-project
```

### 5분 빠른 시작

```bash
# 1. 새 기능 계획 - SPEC 자동 생성
/alfred:1-plan "사용자 인증 기능 (JWT 토큰)"

# 2. TDD 실행 - 자동으로 테스트 → 구현 → 리팩토링
/alfred:2-run SPEC-AUTH-001

# 3. 문서 동기화 및 품질 검증
/alfred:3-sync
```

**결과**: 요구사항 명확화 → 테스트 우선 구현 → 자동 문서화 → 품질 보증까지 완료!

---

## 📊 프로젝트 통계

| 항목 | 수치 |
|------|------|
| **테스트 커버리지** | 87%+ |
| **지원 언어** | 10개 (Python, JavaScript, TypeScript, Go, Rust, Kotlin, Ruby, PHP, Java, C#) |
| **AI 에이전트** | 19명 전문가팀 |
| **프로덕션급 스킬** | 69개+ |
| **오픈소스 라이선스** | MIT |

---

## 🌟 주요 특징

### 다국어 지원
- **4개 언어 지원**: 한국어, 영어, 일본어, 중국어
- **자동 번역**: AI 기반 고품질 번역
- **실시간 동기화**: 모든 언어로 최신 문서 제공

### 지원되는 기술 스택
- **프론트엔드**: React, Vue, Angular (TypeScript)
- **백엔드**: Node.js, Python, Go, Rust
- **데이터베이스**: SQL, NoSQL (MongoDB, PostgreSQL)
- **배포**: Docker, Kubernetes, AWS, Vercel

### 팀 협업
- **개인 모드**: 자유로운 로컬 개발
- **팀 모드**: Feature branch, PR 자동 관리, 자동 merge
- **실시간 컨텍스트**: 팀원 간 완벽한 문서 공유

---

## 📚 학습 자원

### 공식 문서
- **[시작하기](getting-started/installation.md)**: 설치 및 기본 설정
- **[사용 가이드](guides/alfred/index.md)**: Alfred 워크플로우 완벽 가이드
- **[API 참조](reference/cli/index.md)**: 명령어 및 스킬 API
- **[개발자 가이드](contributing/index.md)**: 프로젝트 기여 및 확장

### 핵심 가이드
- **[SPEC 작성](guides/specs/basics.md)**: SPEC-First 개발 방법론
- **[TDD 실행](guides/tdd/red.md)**: RED → GREEN → REFACTOR 사이클
- **[TAG 시스템](guides/specs/tags.md)**: 완전한 추적성 관리

---

## ✨ 커뮤니티

- **GitHub**: [modu-ai/moai-adk](https://github.com/modu-ai/moai-adk)
- **Issues**: [버그 리포트 및 기능 요청](https://github.com/modu-ai/moai-adk/issues)
- **라이선스**: MIT (상업적 사용 가능)

---

## 🎬 다음 단계

<div align="center">

### 지금 바로 시작하세요!

[빠른 시작 가이드](getting-started/installation.md) · [Alfred 워크플로우](guides/alfred/index.md) · [GitHub 저장소](https://github.com/modu-ai/moai-adk)

---

**MoAI-ADK**로 SPEC-First TDD 개발의 강력함을 경험하세요!

</div>
