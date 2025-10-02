---
id: STRUCTURE-001
version: 0.0.1
status: active
created: 2025-10-01
updated: 2025-10-01
authors: ["@goos"]
---

# @DOC:ARCHITECTURE-001 MoAI-ADK 아키텍처 설계

## HISTORY

### v0.0.1 (2025-10-01)
- **INITIAL**: MoAI-ADK 아키텍처 초기 문서화
- **AUTHOR**: @goos
- **SCOPE**: CLI 및 핵심 아키텍처 모듈 정의

## @DOC:ARCHITECTURE-OVERVIEW-001 시스템 아키텍처 전략

### 아키텍처 접근법
MoAI-ADK는 유연하고 확장 가능한 플러그인 기반 CLI 아키텍처를 채택했습니다:

```
MoAI-ADK Architecture
├── CLI Layer               # 사용자 인터페이스 및 명령어 처리
│   ├── Commander           # CLI 명령 라우팅
│   └── Inquirer            # 대화형 프롬프트
│
├── Core Layer              # 핵심 비즈니스 로직
│   ├── Project Detector    # 프로젝트 언어/구조 감지
│   ├── Config Manager      # 설정 및 환경 관리
│   ├── Git Workflow        # Git 통합 및 브랜치 관리
│   └── Agent Orchestrator  # Alfred SuperAgent
│
├── Agent Layer             # 9개 전문 에이전트
│   ├── spec-builder        # SPEC 및 요구사항 관리
│   ├── code-builder        # TDD 구현 및 테스트
│   ├── doc-syncer          # 문서 동기화
│   ├── tag-agent           # @TAG 시스템
│   ├── git-manager         # Git 작업 자동화
│   ├── debug-helper        # 오류 진단
│   ├── trust-checker       # 코드 품질 검증
│   ├── cc-manager          # Claude Code 설정 관리
│   └── project-manager     # 프로젝트 초기화
│
└── Utility Layer           # 공통 유틸리티
    ├── Logger              # 로깅 시스템
    ├── File Processor      # 파일 및 디렉토리 작업
    └── System Detector     # 시스템/언어 감지
```

**선택 이유**:
- 고도의 모듈성과 확장성
- 각 레이어의 명확한 책임 분리
- 플러그인 기반 아키텍처로 새로운 언어/도구 쉽게 추가 가능

## @DOC:MODULES-001 모듈별 책임과 상호작용

### 1. CLI Layer
- **책임**: 사용자 입력 처리, 명령어 라우팅
- **주요 컴포넌트**:
  - `Commander`: CLI 명령 정의 및 처리
  - `Inquirer`: 대화형 설정 마법사

### 2. Core Layer
- **책임**: 프로젝트 초기화, 설정 관리, Git 워크플로우
- **주요 컴포넌트**:
  - `Project Detector`: 프로젝트 언어/프레임워크 자동 감지
  - `Config Manager`: 프로젝트 및 전역 설정 관리
  - `Git Workflow`: 브랜치, PR, 커밋 관리
  - `Agent Orchestrator`: Alfred SuperAgent

### 3. Agent Layer
- **책임**: 개발 생명주기의 전문화된 작업 수행
- **상호작용 패턴**:
  - 순차적 실행 (의존성 있는 작업)
  - 병렬 실행 (독립적인 작업)
  - 중앙 오케스트레이션 (Alfred)

### 4. Utility Layer
- **책임**: 크로스 레이어 공통 기능 제공
- **주요 기능**:
  - 로깅
  - 파일 처리
  - 시스템/언어 감지
  - 타입 안전성 유틸리티

## @DOC:INTEGRATION-001 외부 시스템 통합

### CI/CD 통합
- **GitHub Actions/Jenkins**: 빌드, 테스트, 배포 자동화
- **인증 방식**: OAuth/GitHub App
- **데이터 교환**: JSON, YAML 메타데이터
- **장애 대응**: 실패 시 롤백, 알림 시스템

### 외부 개발 도구
- **IDE 플러그인**: VS Code, JetBrains 계열
- **코드 분석 도구**: SonarQube, CodeClimate
- **의존성 관리**: npm, pip, Maven, Cargo

## @DOC:TRACEABILITY-001 @TAG 추적성 전략

### TAG 체인 검증
- **방법**: 정규식 기반 코드 직접 스캔
- **검증 범위**: src/, tests/, docs/
- **검증 주기**: 코드 변경 시 실시간

### 추적성 검증 예시
```bash
# TAG 체인 무결성 검사
rg '@(SPEC|TEST|CODE|DOC):' -n \
   .moai/specs/ \
   tests/ \
   src/ \
   docs/
```

## Legacy Context

### 기존 시스템 현황
- TypeScript CLI 프로토타입
- 초기 @TAG 시스템 구현
- 단일 언어(TypeScript) 지원

### 마이그레이션 고려사항
1. 멀티 언어 지원 아키텍처로 전환
2. 플러그인 시스템 도입
3. 코드 품질 및 추적성 강화

## TODO:STRUCTURE-001 구조 개선 계획

1. **플러그인 메커니즘 강화**
   - 동적 에이전트 로딩
   - 커스텀 에이전트 지원

2. **성능 최적화**
   - 캐싱 전략 개선
   - 비동기 에이전트 처리

3. **보안성 강화**
   - 샌드박스 에이전트 실행
   - 권한 관리 시스템

## EARS 아키텍처 요구사항

### Ubiquitous Requirements
- 시스템은 모듈화된 아키텍처를 제공해야 한다
- 시스템은 크로스 플랫폼 호환성을 유지해야 한다

### Event-driven Requirements
- WHEN 새로운 언어/프레임워크가 감지되면, 시스템은 자동으로 지원 도구를 로드해야 한다
- WHEN 에이전트 간 의존성 충돌이 발생하면, 시스템은 자동으로 해결해야 한다

### State-driven Requirements
- WHILE 프로젝트가 진행 중이면, 시스템은 지속적으로 아키텍처 무결성을 검증해야 한다
- WHILE 개발 모드일 때, 시스템은 동적 플러그인 로딩을 지원해야 한다

### Optional Features
- WHERE CI/CD 파이프라인이 구성되면, 시스템은 자동 아키텍처 진단을 제공할 수 있다
- WHERE 외부 플러그인이 연동되면, 시스템은 확장된 기능을 지원할 수 있다

### Constraints
- IF 메모리 사용량이 임계값을 초과하면, 시스템은 에이전트 실행을 제한해야 한다
- 각 에이전트의 실행 시간은 5분을 초과하지 않아야 한다

---

_이 문서는 `/alfred:2-build` 시 아키텍처 구현의 기준이 됩니다._