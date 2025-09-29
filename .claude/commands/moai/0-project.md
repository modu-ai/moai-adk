---
name: moai:0-project
description: 범용 언어 지원 프로젝트 킥오프. product/structure/tech 문서 생성 with JSON 기반 tags.json 최적화
argument-hint: "PROJECT_NAME"
tools: Read, Write, Edit, MultiEdit, Grep, Glob, TodoWrite, Bash
---

# MoAI-ADK 0단계: 범용 언어 지원 프로젝트 문서 초기화/갱신

**프로젝트 초기화 대상**: $ARGUMENTS

## 🔍 STEP 1: 환경 분석 및 프로젝트 인터뷰 시작

프로젝트 환경을 분석하고 체계적인 인터뷰를 즉시 시작합니다.

### 환경 분석 및 인터뷰 진행

1. **프로젝트 유형 감지**
   - 신규 프로젝트 vs 기존 프로젝트 분류
   - 언어/프레임워크 자동 감지
   - 프로젝트 구조 및 복잡도 평가

2. **문서 현황 분석**
   - 기존 문서 상태 확인
   - 부족한 정보 영역 식별
   - 보완 필요 항목 정리

3. **프로젝트 인터뷰 실행**
   - 프로젝트 유형별 질문 트리 적용
   - Product Discovery: 미션, 사용자, 해결 문제
   - Structure Blueprint: 모듈 전략, 책임 구분, 외부 연동
   - Tech Stack Mapping: 언어/런타임, 프레임워크, 품질 정책

---

## 📋 STEP 2: 문서 생성 승인 및 완료

모든 인터뷰가 완료된 후 문서 생성에 대한 최종 승인을 받습니다.

### 문서 생성 승인 단계

인터뷰 완료 후 다음 중 선택하세요:
- **"생성"** 또는 **"진행"**: 수집된 정보로 문서 생성
- **"수정 [내용]"**: 특정 내용 수정 후 문서 생성
- **"중단"**: 프로젝트 초기화 중단

### 프로젝트 문서 생성 실행

사용자 승인 후 project-manager 에이전트가 **언어별 최적화**된 프로젝트 문서를 생성합니다.

## 기능 (언어별 최적화)

`.moai/project/{product,structure,tech}.md`를 **프로젝트 언어별 최적화**된 템플릿으로 생성·갱신합니다.
- **NEW**: 프로젝트 언어 감지 기반 고성능 분석 및 설정
- **NEW**: 프로젝트 초기화 성능 최적화 (고성능 CLI 182ms 달성)
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

## 🚀 범용 언어 지원 원칙

**⚡ 핵심**: `/moai:0-project`는 언어별 최적화된 고속 프로젝트 초기화로 JSON 기반 tags.json 통합을 제공합니다.

### 🔗 언어별 최적화 전략

```python
# 언어별 최적화된 프로젝트 초기화 구현
from moai_adk.core.project import ProjectInitializer
from moai_adk.core.language_detector import detect_project_language
from moai_adk.core.tag_system import TagDatabase

async def optimal_project_init(project_name: str) -> ProjectResult:
    """프로젝트 언어를 감지하고 최적화된 초기화 수행"""

    # 프로젝트 언어 감지
    language = detect_project_language()
    initializer = ProjectInitializer(language=language)
    tag_db = TagDatabase()

    # 언어별 최적화된 고성능 초기화
    result = await initializer.initialize({
        'name': project_name,
        'language': language,
        'enable_json_tagging': True
    })

    # tags.json 자동 초기화
    await tag_db.initialize_project_tags(project_name)

    return result
}
```

### 실행 성능 비교

| 구현 방식 | 초기화 시간 | 시스템 검증 | 확장성 |
|-----------|-------------|-------------|--------|
| **TypeScript 최적화** | ~0.18초 | 완전 자동 | 최고 |
| **기존 Python** | ~4.6초 | 수동 | 중간 |
| **SQLite3 tags.db** | 즉시 | 실시간 | 최고 |

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

## 📋 STEP 1 실행 가이드: 환경 분석 및 인터뷰 즉시 시작

### 🎯 인터뷰 사용법 안내

**인터뷰가 시작되면 다음 방법으로 답변해주세요:**

1. **선택지 답변 방법**:
   - 질문에 대해 1, 2, 3 등의 숫자로 답변
   - 예: "1" 또는 "2번을 선택합니다"
   - 다중 선택 가능한 경우: "1,3" 또는 "1과 3번"

2. **추가 설명이 필요한 경우**:
   - 숫자 선택 후 추가 설명 제공 가능
   - 예: "2번. 그런데 추가로 설명하면..."

3. **모르거나 확실하지 않은 경우**:
   - **"모르겠음"** 또는 **"추천"** 입력
   - AI가 프로젝트 상황에 맞는 최적의 선택지를 추천해드립니다
   - 추천 후에도 다시 선택할 수 있습니다

4. **질문을 건너뛰려면**:
   - **"패스"** 또는 **"스킵"** 입력
   - 나중에 문서에서 수정 가능합니다

---

### 1. 프로젝트 환경 분석

다음을 우선적으로 실행하여 프로젝트 환경을 분석합니다:

```bash
# 프로젝트 구조 및 언어 감지
@agent-project-manager --mode=analysis --project=$ARGUMENTS
```

#### 분석 체크리스트

- [ ] **프로젝트 유형**: 신규 vs 기존 프로젝트 분류
- [ ] **언어 감지**: Python, Node.js, 풀스택 등 자동 감지
- [ ] **문서 현황**: 기존 문서 상태 및 부족한 영역 확인
- [ ] **구조 복잡도**: 단일 언어 vs 하이브리드 vs 마이크로서비스

### 2. 프로젝트 인터뷰 즉시 실행

#### 프로젝트 유형별 질문 전략

| 프로젝트 유형 | 질문 카테고리 | 중점 영역 |
|-------------|-------------|----------|
| 신규 프로젝트 | Product Discovery | 미션, 사용자, 해결 문제 |
| 기존 프로젝트 | Legacy Analysis | 코드 기반, 기술 부채, 통합점 |
| TypeScript 전환 | Migration Strategy | 기존 프로젝트의 TypeScript 전환 |

#### 인터뷰 진행 방식

- **Product Discovery**: 핵심 비즈니스 가치, 주요 사용자층
- **Structure Blueprint**: 모듈 전략, 외부 시스템 연동
- **Tech Stack Mapping**: 언어/프레임워크, 품질 정책, 배포 전략

---

## 📋 STEP 2 실행 가이드: 문서 생성 승인 및 완료

모든 인터뷰 완료 후 수집된 정보를 바탕으로 문서 생성을 진행합니다.

### 문서 생성 승인 요청

인터뷰 완료 후 다음 형식으로 승인을 요청합니다:

```
## 프로젝트 문서 생성 요청: [PROJECT-NAME]

### 📊 수집된 정보 요약
- **프로젝트 유형**: [신규/기존/하이브리드]
- **감지된 언어**: [언어 목록]
- **핵심 미션**: [핵심 가치 제안]
- **주요 사용자**: [타겟 사용자층]
- **해결 문제**: [핵심 해결 문제들]
- **기술 스택**: [선택된 기술들]

### 🎯 생성할 문서
- **product.md**: 비즈니스 요구사항 및 사용자 분석
- **structure.md**: 시스템 아키텍처 및 모듈 설계
- **tech.md**: 기술 스택 및 품질 정책
- **config.json**: 프로젝트 설정 파일

---
**문서 생성 승인**: 위 정보로 프로젝트 문서를 생성하시겠습니까?
("생성", "수정 [내용]", "중단" 중 선택)
```

### 문서 생성 실행

사용자가 **"생성"** 또는 **"진행"**을 선택한 경우에만 다음을 실행합니다:

```bash
# 프로젝트 문서 생성
@agent-project-manager --mode=generate --project=$ARGUMENTS --approved=true
```

### TypeScript 초기화 단계별 가이드

1. **환경 감지**: 언어/프레임워크 자동 감지 및 확인
2. **인터뷰 실행**: 프로젝트 유형별 체계적 질문 진행
3. **문서 생성**: product/structure/tech.md 생성 또는 갱신
4. **설정 저장**: config.json에 감지된 언어/도구 설정 저장

### 프로젝트 유형별 처리 방식

#### 신규 프로젝트 (그린필드)

**0. 프로젝트 유형 및 언어 감지**
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
| 1. 미션 정의 | 제품이 약속하는 핵심 가치는 무엇인가요?  | 1) 생산성 향상 2) 품질 개선 3) 학습 지원 4) 커뮤니티 구축 5) 모르겠음 (AI 추천 요청) | 선택한 가치에 따라 성공 지표 후보를 미리 메모합니다.                     |
| 2. 사용자층  | 주요 사용자는 누구인가요?                | 1) 개인 개발자 2) 팀/조직 3) 플랫폼 운영자 4) 교육 기관 5) 기타 6) 모르겠음 (AI 추천 요청) | 각 사용자군에 대해 "즉시 얻고 싶은 결과"를 1문장씩 작성합니다.           |
| 3. 해결 문제 | 가장 시급한 문제 3가지는 무엇인가요?     | 1) 워크플로우 혼란 2) 품질 격차 3) 협업 지연 4) 규정 준수 5) 기타 6) 모르겠음 (AI 추천 요청) | 선택 항목마다 "현재 실패 사례"를 기록해 이후 SPEC 우선순위로 연결합니다. |
| 4. 차별점    | 경쟁 솔루션 대비 강점은 무엇인가요?      | 1) 자동화 심도 2) 문서 동기화 3) 추적성 4) 거버넌스 5) 기타 6) 모르겠음 (AI 추천 요청) | 강점이 발휘되는 대표 시나리오를 product.md에 포함합니다.                 |
| 5. 성공 지표 | 성공을 입증할 첫 번째 지표는 무엇인가요? | 1) 채택률 2) 반복 속도 3) 품질 지표 4) 생태계 기여 5) 기타 6) 모르겠음 (AI 추천 요청) | KPI에 측정 주기와 베이스라인을 기입합니다.                               |

#### Structure Blueprint

| 단계            | 질문                                | 선택 분기                                                                                        | 후속 결정                                            |
| --------------- | ----------------------------------- | ------------------------------------------------------------------------------------------------ | ---------------------------------------------------- |
| 1. 모듈 전략   | 시스템을 몇 개 모듈로 나눌까요?     | 1) 핵심 3모듈(Planning/Development/Integration) 2) 도메인별 확장 3) 마이크로서비스 4) 단일 핵심 5) 모르겠음 (AI 추천 요청) | 선택 이유와 예외 규칙을 structure.md에 명시합니다.   |
| 2. 책임 구분   | 각 모듈이 담당할 책임은 무엇인가요? | 1) 명세/요구사항 2) 구현/TDD 3) 문서/PR 4) 데이터/분석 5) 기타 6) 모르겠음 (AI 추천 요청) | 책임별로 입력→처리→출력 흐름을 표로 정리합니다.      |
| 3. 외부 연동   | 어떤 외부 시스템과 통신하나요?      | 1) Claude Code 2) Git/GitHub 3) GitLab/CVS 4) CI/CD 5) 패키지 레지스트리 6) 모니터링 7) 기타 8) 모르겠음 (AI 추천 요청) | 연동당 인증 방식, 장애 시 대체 절차를 기록합니다.    |
| 4. 추적성 전략 | 요구사항 추적을 어떻게 보장하나요?  | 1) @AI-TAG 체계 2) Issue 템플릿 3) 계약 테스트 4) 감사 로그 5) 기타 6) 모르겠음 (AI 추천 요청) | 선택한 전략의 유지 주기를 structure.md에 서술합니다. |

#### Tech Stack Mapping

| 단계              | 질문                                     | 선택 분기                                                                                                                                                                   | 후속 질문                                                            |
| ----------------- | ---------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| 1. 언어 & 런타임 | 어떤 언어/런타임 조합이 필요한가요?      | 1) Python ≥3.11 2) Node.js ≥18 3) JVM (Java/Kotlin) 4) .NET 5) Go 6) Rust 7) Swift/Kotlin Mobile 8) C/C++ 임베디드 9) 기타 10) 모르겠음 (AI 추천 요청) | 다중 언어 선택 시 주요 모듈과 매핑 관계를 tech.md에 표로 기록합니다. |
| 2. 프레임워크    | 핵심 프레임워크/라이브러리는 무엇인가요? | 1) Web(React, Next, Vue, Angular) 2) Backend(FastAPI, Spring, Nest, Express) 3) Data(PyTorch, TensorFlow, Spark) 4) Mobile(Compose, SwiftUI, Flutter) 5) DevOps(Terraform, Pulumi) 6) 기타 7) 모르겠음 (AI 추천 요청) | 프레임워크별 빌드/테스트 명령과 CI 파이프라인 요구사항을 적습니다.   |
| 3. 품질 정책     | 어떤 품질 게이트를 적용하나요?           | 1) 테스트 커버리지 목표 2) 정적 분석(ruff, eslint, detekt 등) 3) 코드 포매터 4) 계약/E2E 테스트 5) 카나리 배포 6) 기타 7) 모르겠음 (AI 추천 요청) | 실패 시 대응 규칙과 책임자를 명시합니다.                             |
| 4. 보안·운영     | 보안 및 운영 정책은 무엇인가요?          | 1) 비밀 관리(Secrets Manager, Vault) 2) 접근 제어(Role 기반) 3) 로깅/관측(OpenTelemetry, ELK) 4) 감사 대응(ISO/SOC 기준) 5) 기타 6) 모르겠음 (AI 추천 요청) | 각 정책의 감사 대상 데이터와 보존 기간을 tech.md에 기록합니다.       |
| 5. 배포 채널     | 배포 타깃은 어디인가요?                  | 1) 클라우드(IaaS/PaaS) 2) 온프레미스 3) 모바일 스토어 4) 에지/임베디드 5) 패키지 레지스트리(PyPI/NPM) 6) 기타 7) 모르겠음 (AI 추천 요청) | 채널별 릴리스 절차와 롤백 전략을 정의합니다.                         |

> 💡 팁: 신규 프로젝트는 모든 선택지가 열려 있으므로, 인터뷰 중 분기가 추가되면 표의 “선택 분기”에 표시하고 product/structure/tech 문서에 동일한 순서로 정리하세요. 다중 스택을 선택할 경우 환경 변수·테스트 스위트·운영 책임자를 명시합니다.

### 2) 기존 프로젝트(레거시 도입) 전용 트리

#### Legacy Snapshot & Alignment

| 단계                       | 핵심 질문                                           | 체크 포인트                                          | 후속 작업                                                                                                                     |
| -------------------------- | --------------------------------------------------- | ---------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| 1. 코드 기반 요약         | 현재 저장소의 주요 디렉터리와 모듈은 무엇인가요?    | **("패스" 입력 시 자동 분석 진행)** | structure.md에 기존 모듈 구조와 책임을 먼저 기록하고, 필요 시 리팩터링 후보에 @DEBT 태그 추가                                 |
| 2. 빌드/테스트 파이프라인 | 어떤 빌드·테스트 절차가 존재하나요?                 | 1) CI 설정 있음 2) 스크립트만 있음 3) 커버리지 없음 4) 린트 없음 5) 전부 없음 6) 모르겠음 (AI 추천 요청) | tech.md 품질 섹션에 현재 상태와 목표 차이를 @DEBT(부채) 또는 @TODO 항목으로 명시하고, 필요한 경우 `/moai:2-build` 계획에 연결 |
| 3. 기술 부채 및 제한      | 유지해야 할 규칙이나 레거시 제약은 무엇인가요?      | 1) 오래된 프레임워크 2) 의존성 잠금 3) 배포 환경 제한 4) 특별한 제약 없음 5) 모르겠음 (AI 추천 요청) | product.md 또는 structure.md에 "현존 제약" 블록 작성, 해결 계획은 @TASK 로 분리                                               |
| 4. 통합 지점              | 외부 시스템과의 연동 관계는 어떻게 구성되어 있나요? | 1) API 연동 있음 2) DB 연동 있음 3) Queue 시스템 4) 인증 시스템 5) 연동 없음 6) 모르겠음 (AI 추천 요청) | structure.md Integration 섹션에 현재 연결 방식과 위험도를 기록                                                                |
| 5. 마이그레이션 전략      | MoAI-ADK로 전환하며 즉시 해야 할 작업은 무엇인가요? | 1) 명세 작성 2) 테스트 추가 3) 문서화 4) 배포 자동화 5) 전부 필요 6) 모르겠음 (AI 추천 요청) | 각 문서 끝에 "Initial Migration Plan" 목록을 추가하고 `/moai:1-spec` 후보 SPEC과 연결                                         |

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
- **TypeScript 전환 프로젝트**: 기존 프로젝트의 TypeScript 마이그레이션 전략과 단계별 전환 계획 수립

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

## 에이전트 역할 분리

### project-manager 전담 영역

- 프로젝트 환경 분석 및 언어/프레임워크 자동 감지
- 프로젝트 유형별 체계적 인터뷰 진행
- product/structure/tech.md 문서 생성/갱신
- config.json 설정 파일 관리
- TypeScript 환경 최적화 설정 및 SQLite3 tags.db 통합

## 2단계 워크플로우 실행 순서

### Phase 1: 분석 및 인터뷰 단계

**환경 분석기 및 project-manager**가 **연속적으로** 수행:

1. **프로젝트 구조 스캔**: 디렉토리 구조, 기존 파일, 언어 감지
2. **문서 현황 확인**: 기존 문서 상태 및 부족한 영역 식별
3. **프로젝트 인터뷰 실행**: 프로젝트 유형별 체계적 질문 진행
   - Product Discovery (미션, 사용자, 문제)
   - Structure Blueprint (모듈, 책임, 연동)
   - Tech Stack Mapping (언어, 프레임워크, 품질)

### Phase 2: 문서 생성 승인 및 완료 단계

인터뷰 완료 후 **사용자 승인**을 받아 문서 생성:

1. **인터뷰 결과 요약**: 수집된 모든 정보의 구조화된 요약 제시
2. **문서 생성 승인**: 최종 문서 생성에 대한 사용자 확인
3. **문서 생성**: product/structure/tech.md 생성/갱신
4. **설정 저장**: config.json에 프로젝트 설정 저장
5. **품질 검증**: 문서 완성도 및 설정 유효성 검사

## 다음 단계

- 신규 프로젝트: `/moai:1-spec`을 실행해 설계 기반 SPEC 백로그를 생성합니다.
- 레거시 프로젝트: `/moai:1-spec` 실행 전, product/structure/tech 문서의 @DEBT/@TASK/@TODO 항목을 검토해 우선순위를 확정합니다.
- 설정 변경(개인 ↔ 팀, 출력 스타일 등)이 필요하면 `/moai:0-project update`로 인터뷰를 다시 진행하세요. CLI 기반 직접 변경이 필요한 경우에만 `moai config --mode team|personal` 등 명령을 사용합니다.
