---
name: moai:0-project
description: 프로젝트 킥오프. product/structure/tech 문서 생성
argument-hint: "PROJECT_NAME"
tools: Read, Write, Edit, MultiEdit, Grep, Glob, TodoWrite, Bash
---

# /moai:0-project — 프로젝트 문서 초기화/갱신

## 기능

- ULTRATHINK: `.moai/project/{product,structure,tech}.md`를 **간결하고 실용적인 템플릿**으로 생성·갱신합니다.
- 완료 후 CLAUDE.md의 `@.moai/project/*` 임포트를 통해 메모리에 자동 반영됩니다.

## 사용법

```bash
/moai:0-project                    # 프로젝트 문서 생성
/moai:0-project update            # 기존 설정 재조정
```

### 권한 및 설정

- `.moai/project` 경로는 Guard가 승인하므로 별도의 스크립트 실행이나 권한 토큰이 필요하지 않습니다.
- 프로젝트 이름을 명시하려면 moai:0-project 뒤에 원하는 이름을 덧붙입니다.
- 개인/팀 모드, 출력 스타일을 함께 확인·수정할 수 있습니다.

## 🚀 최적화된 실행 원칙

**⚡ 핵심**: `/moai:0-project`는 병렬 분석으로 빠른 프로젝트 초기화를 제공합니다.

### 병렬 처리 구조 (성능 최적화)

**Phase 1: 동시 분석 실행**

```
Task 1 (haiku): 빠른 환경 스캔
├── 언어/프레임워크 자동 감지
├── 프로젝트 유형 판단 (신규/레거시)
└── config.json 설정 준비

Task 2 (haiku): project-manager 호출
├── 기존 문서 상태 확인
├── 사용자 인터뷰 진행
└── 문서 작성/갱신
```

**성능 향상**: 환경 감지와 문서 분석을 병렬 처리하여 초기화 시간 40% 단축

### 통합 처리

1. **병렬 분석 결과 통합**: 언어 감지 + 사용자 인터뷰 결과 결합
2. **문서 생성/갱신**: product/structure/tech.md 생성 또는 업데이트
3. **config.json 저장**: 감지된 언어/도구 설정을 자동 저장

### 금지 사항

- ❌ `.claude/memory/` 디렉토리에 파일 생성
- ❌ `.claude/commands/moai/*.json` 파일 생성
- ❌ 기존 문서 불필요한 덮어쓰기

## 진행 순서

0. **프로젝트 유형 및 언어 감지**
   - **프로젝트 유형**: `moai init project-name`(신규) vs `moai init .`(기존) 분류
   - **🚀 언어 자동 감지**: 파일 구조 분석을 통한 언어/도구 자동 감지

     ```python
     # 단일 언어 감지
     if detect_python(): config['project_language'] = 'python'
     if detect_node(): config['project_language'] = 'javascript'

     # 풀스택 감지
     if has_backend_frontend():
         config['project_type'] = 'fullstack'
         config['languages'] = {
             'backend': analyze_backend_dir(),
             'frontend': analyze_frontend_dir()
         }
     ```

   - **감지 규칙**:
     - `pyproject.toml`, `requirements.txt` → Python
     - `package.json`, `tsconfig.json` → Node.js/TypeScript
     - `backend/` + `frontend/` 구조 → 풀스택
     - `go.mod` → Go, `Cargo.toml` → Rust
   - 감지 결과를 사용자에게 설명하고 최종 확인

1. **상태 스냅샷 공유** – `.moai/project/*.md`, README, CLAUDE.md, 소스 디렉터리 구조와 모든 파일을 읽어 현재 문서·코드·테스트가 어떤 상태인지 요약합니다.
2. **프로젝트 유형별 인터뷰 진행** – 확인된 유형(신규/레거시)에 따라 적절한 질문 트리를 선택하고 대화를 이끕니다.
3. **문서 작성 및 config.json 저장** – 응답을 토대로 product/structure/tech 문서를 갱신하고, **언어 감지 결과를 .moai/config.json에 저장**합니다.

   **🚀 config.json 자동 생성/업데이트**:

   ```json
   {
     "project_name": "detected-name",
     "project_type": "single|fullstack|microservice",

     // 단일 언어 프로젝트
     "project_language": "python",
     "test_framework": "pytest",
     "linter": "ruff",
     "formatter": "black",

     // 풀스택 프로젝트 (선택적)
     "languages": {
       "backend": {
         "language": "python",
         "path": "backend/",
         "test_framework": "pytest"
       },
       "frontend": {
         "language": "typescript",
         "framework": "react",
         "path": "frontend/",
         "test_framework": "jest"
       }
     },

     "coverage_target": 85,
     "mode": "personal"
   }
   ```

   기존 내용과 충돌하는 부분은 주석이나 @TODO로 표시해 후속 검토를 돕습니다.

각 단계는 **질문 → 사용자 응답 → MoAI 요약/작성** 흐름으로 진행되며, 응답 형태는 자유입니다. 아래 인터뷰 트리를 참고해 프로젝트 유형에 맞는 질문을 선택하세요.

## 프로젝트 유형별 인터뷰 트리

### 1) 신규 프로젝트(그린필드) 전용 트리

#### Product Discovery

| 단계          | 핵심 질문                                | 선택 분기                                           | 후속 질문/결정                                                           |
| ------------- | ---------------------------------------- | --------------------------------------------------- | ------------------------------------------------------------------------ |
| A1. 미션 정의 | 제품이 약속하는 핵심 가치는 무엇인가요?  | 전략: 생산성 ∙ 품질 ∙ 학습 ∙ 커뮤니티               | 선택한 가치에 따라 성공 지표 후보를 미리 메모합니다.                     |
| A2. 사용자층  | 주요 사용자는 누구인가요?                | 개인 개발자 ∙ 팀/조직 ∙ 플랫폼 운영자 ∙ 교육 기관   | 각 사용자군에 대해 “즉시 얻고 싶은 결과”를 1문장씩 작성합니다.           |
| A3. 해결 문제 | 가장 시급한 문제 3가지는 무엇인가요?     | 워크플로우 혼란 ∙ 품질 격차 ∙ 협업 지연 ∙ 규정 준수 | 선택 항목마다 “현재 실패 사례”를 기록해 이후 SPEC 우선순위로 연결합니다. |
| A4. 차별점    | 경쟁 솔루션 대비 강점은 무엇인가요?      | 자동화 심도 ∙ 문서 동기화 ∙ 추적성 ∙ 거버넌스       | 강점이 발휘되는 대표 시나리오를 product.md에 포함합니다.                 |
| A5. 성공 지표 | 성공을 입증할 첫 번째 지표는 무엇인가요? | 채택률 ∙ 반복 속도 ∙ 품질 ∙ 생태계 기여             | KPI에 측정 주기와 베이스라인을 기입합니다.                               |

#### Structure Blueprint

| 단계            | 질문                                | 선택 분기                                                                                        | 후속 결정                                            |
| --------------- | ----------------------------------- | ------------------------------------------------------------------------------------------------ | ---------------------------------------------------- |
| S1. 모듈 전략   | 시스템을 몇 개 모듈로 나눌까요?     | 1) 핵심 3모듈(Planning/Development/Integration) 2) 도메인 별 확장 3) 마이크로서비스 4) 단일 핵심 | 선택 이유와 예외 규칙을 structure.md에 명시합니다.   |
| S2. 책임 구분   | 각 모듈이 담당할 책임은 무엇인가요? | 명세/요구 ∙ 구현/TDD ∙ 문서/PR ∙ 데이터/분석                                                     | 책임별로 입력→처리→출력 흐름을 표로 정리합니다.      |
| S3. 외부 연동   | 어떤 외부 시스템과 통신하나요?      | Claude Code ∙ Git/GitHub ∙ GitLab/CVS ∙ CI/CD ∙ 패키지 레지스트리 ∙ 모니터링                     | 연동당 인증 방식, 장애 시 대체 절차를 기록합니다.    |
| S4. 추적성 전략 | 요구사항 추적을 어떻게 보장하나요?  | TAG 체계 ∙ Issue 템플릿 ∙ 계약 테스트 ∙ 감사 로그                                                | 선택한 전략의 유지 주기를 structure.md에 서술합니다. |

#### Tech Stack Mapping

| 단계              | 질문                                     | 선택 분기                                                                                                                                                                   | 후속 질문                                                            |
| ----------------- | ---------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| T1. 언어 & 런타임 | 어떤 언어/런타임 조합이 필요한가요?      | Python ≥3.11 ∙ Node.js ≥18 ∙ JVM (Java/Kotlin) ∙ .NET ∙ Go ∙ Rust ∙ Swift/Kotlin Mobile ∙ C/C++ 임베디드                                                                    | 다중 언어 선택 시 주요 모듈과 매핑 관계를 tech.md에 표로 기록합니다. |
| T2. 프레임워크    | 핵심 프레임워크/라이브러리는 무엇인가요? | Web(React, Next, Vue, Angular) ∙ Backend(FastAPI, Spring, Nest, Express) ∙ Data(PyTorch, TensorFlow, Spark) ∙ Mobile(Compose, SwiftUI, Flutter) ∙ DevOps(Terraform, Pulumi) | 프레임워크별 빌드/테스트 명령과 CI 파이프라인 요구사항을 적습니다.   |
| T3. 품질 정책     | 어떤 품질 게이트를 적용하나요?           | 테스트 커버리지 목표 ∙ 정적 분석(ruff, eslint, detekt 등) ∙ 포매터 ∙ 계약/E2E 테스트 ∙ 카나리 배포                                                                          | 실패 시 대응 규칙과 책임자를 명시합니다.                             |
| T4. 보안·운영     | 보안 및 운영 정책은 무엇인가요?          | 비밀 관리(Secrets Manager, Vault) ∙ 접근 제어(Role 기반) ∙ 로깅/관측(OpenTelemetry, ELK) ∙ 감사 대응(ISO/SOC 기준)                                                          | 각 정책의 감사 대상 데이터와 보존 기간을 tech.md에 기록합니다.       |
| T5. 배포 채널     | 배포 타깃은 어디인가요?                  | 클라우드(IaaS/PaaS) ∙ 온프레미스 ∙ 모바일 스토어 ∙ 에지/임베디드 ∙ 패키지 레지스트리(PyPI/NPM)                                                                              | 채널별 릴리스 절차와 롤백 전략을 정의합니다.                         |

> 💡 팁: 신규 프로젝트는 모든 선택지가 열려 있으므로, 인터뷰 중 분기가 추가되면 표의 “선택 분기”에 표시하고 product/structure/tech 문서에 동일한 순서로 정리하세요. 다중 스택을 선택할 경우 환경 변수·테스트 스위트·운영 책임자를 명시합니다.

### 2) 기존 프로젝트(레거시 도입) 전용 트리

#### Legacy Snapshot & Alignment

| 단계                       | 핵심 질문                                           | 체크 포인트                                          | 후속 작업                                                                                                                     |
| -------------------------- | --------------------------------------------------- | ---------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| L1. 코드 기반 요약         | 현재 저장소의 주요 디렉터리와 모듈은 무엇인가요?    | src/app/packages 폴더 구조, 언어별 비중, 핵심 진입점 | structure.md에 기존 모듈 구조와 책임을 먼저 기록하고, 필요 시 리팩터링 후보에 @DEBT 태그 추가                                 |
| L2. 빌드/테스트 파이프라인 | 어떤 빌드·테스트 절차가 존재하나요?                 | CI 설정, 스크립트, 커버리지·린트 현황                | tech.md 품질 섹션에 현재 상태와 목표 차이를 @DEBT(부채) 또는 @TODO 항목으로 명시하고, 필요한 경우 `/moai:2-build` 계획에 연결 |
| L3. 기술 부채 및 제한      | 유지해야 할 규칙이나 레거시 제약은 무엇인가요?      | 오래된 프레임워크, 의존성 잠금, 배포 환경 제한       | product.md 또는 structure.md에 “현존 제약” 블록 작성, 해결 계획은 @TASK 로 분리                                               |
| L4. 통합 지점              | 외부 시스템과의 연동 관계는 어떻게 구성되어 있나요? | 인증 방식, 사용 중인 API/Queue/DB, 변경 불가 시스템  | structure.md Integration 섹션에 현재 연결 방식과 위험도를 기록                                                                |
| L5. 마이그레이션 전략      | MoAI-ADK로 전환하며 즉시 해야 할 작업은 무엇인가요? | 명세 부재, 테스트 미비, 배포 자동화 공백             | 각 문서 끝에 “Initial Migration Plan” 목록을 추가하고 `/moai:1-spec` 후보 SPEC과 연결                                         |

> ℹ️ 레거시 모드에서는 기존 문서를 덮어쓰지 말고 부족한 부분을 덧붙이거나 TODO로 분리합니다. 자동 분석(`.moai/scripts/project_initializer.py --analyze`) 결과가 있다면 참고하되, 최종 결정은 사용자 확인 후 기록합니다.

## 산출물

- `.moai/project/product.md`
- `.moai/project/structure.md`
- `.moai/project/tech.md`

### 🏷️ TAG 시스템 적용 규칙

**섹션별 @TAG 자동 생성**:

- 미션/비전 → @VISION:MISSION-XXX, @VISION:STRATEGY-XXX
- 사용자 정의 → @REQ:USER-XXX, @REQ:PERSONA-XXX
- 문제 분석 → @REQ:PROBLEM-XXX, @DESIGN:SOLUTION-XXX
- 아키텍처 → @STRUCT:ARCHITECTURE-XXX, @DESIGN:PATTERN-XXX
- 기술 스택 → @TECH:STACK-XXX, @TECH:FRAMEWORK-XXX

**레거시 프로젝트 태그**:

- 기술 부채 → @DEBT:REFACTOR-XXX, @DEBT:TEST-XXX, @DEBT:MIGRATION-XXX
- 해결 계획 → @TASK:MIGRATION-XXX, @TODO:SPEC-BACKLOG-XXX
- 품질 개선 → @TODO:TEST-COVERAGE-XXX, @TODO:DOCS-SYNC-XXX

### 📋 문서별 생성 전략

- **신규 프로젝트**: mission/target/problem/strategy를 기반으로 한 완전한 설계 초안을 작성하고, 미정 항목은 @TODO로 표시합니다.
- **레거시 프로젝트**: 기존 자산 요약 → 발견된 공백을 @DEBT/@TODO 로 표현 → 전환 계획을 @TASK/@DEBT 순으로 구조화합니다. 기존 문서와 충돌하는 경우 "Legacy Context" 섹션을 남겨 추후 비교가 가능하도록 합니다.
- **하이브리드 프로젝트**: 신규 기능과 레거시 유지 영역을 명확히 구분하여 각각 다른 전략 적용

### ⚠️ 에이전틱 코딩 지침

**날짜와 수치 예측 금지**:

- ❌ "3개월 내", "2024 Q4", "50% 단축" 등 구체적 시간/수치 예측 사용 금지
- ✅ "우선순위 높음/중간/낮음", "즉시 필요", "단계적 개선" 등 상대적 표현 사용
- **이유**: AI 개발 환경에서는 개발자의 시간 개념과 다르며, 예측이 불가능함

**추상적/가상 내용 제거**:

- ❌ 가상의 시나리오, 예상 시장 규모, 미래 기술 트렌드 예측 금지
- ✅ 현재 확인 가능한 사실, 기존 기술 스택, 실제 문제점에만 집중

### 참고사항

- `.moai/project` 경로는 MoAI Guard가 자동 허용하므로 별도 토큰이나 스크립트가 필요하지 않습니다.
- 설정을 다시 조정하려면 `/moai:0-project update`로 인터뷰를 다시 진행하세요.

## 다음 단계

- 신규 프로젝트: `/moai:1-spec`을 실행해 설계 기반 SPEC 백로그를 생성합니다.
- 레거시 프로젝트: `/moai:1-spec` 실행 전, product/structure/tech 문서의 @DEBT/@TASK/@TODO 항목을 검토해 우선순위를 확정합니다.
- 설정 변경(개인 ↔ 팀, 출력 스타일 등)이 필요하면 `/moai:0-project update`로 인터뷰를 다시 진행하세요. CLI 기반 직접 변경이 필요한 경우에만 `moai config --mode team|personal` 등 명령을 사용합니다.
