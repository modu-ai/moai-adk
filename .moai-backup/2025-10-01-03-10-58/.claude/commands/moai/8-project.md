---
name: moai:8-project
description: Use PROACTIVELY for 프로젝트 문서 초기화 - product/structure/tech.md 생성 및 언어별 최적화 설정
argument-hint: [PROJECT_NAME] [update]
tools: Read, Write, Edit, MultiEdit, Grep, Glob, TodoWrite, Bash, Agent
---

# MoAI-ADK 8단계: 범용 언어 지원 프로젝트 문서 초기화/갱신

**프로젝트 초기화 대상**: $ARGUMENTS

## 명령어 개요

프로젝트 환경을 분석하고 product/structure/tech.md 문서를 생성/갱신하는 체계적인 초기화 시스템입니다.

- **언어 자동 감지**: Python, TypeScript, Java, Go, Rust 등 자동 인식
- **프로젝트 유형 분류**: 신규 vs 기존 프로젝트 자동 판단
- **고성능 초기화**: TypeScript 기반 CLI로 0.18초 초기화 달성
- **2단계 워크플로우**: 1) 분석 및 계획 → 2) 즉시 인터뷰 진행

## 사용법

```bash
/moai:8-project                    # 프로젝트 문서 생성
/moai:8-project update            # 기존 설정 재조정
/moai:8-project MyProject         # 프로젝트 이름 지정
```

**인수 처리**:
- `$ARGUMENTS` 없음: 현재 디렉토리 분석
- `$ARGUMENTS` = "update": 기존 문서 갱신 모드
- `$ARGUMENTS` = 프로젝트명: 신규 프로젝트로 초기화

## 🚀 STEP 1: 환경 분석 및 인터뷰 계획 수립

프로젝트 환경을 분석하고 체계적인 인터뷰 계획을 수립합니다.

### 1.1 프로젝트 환경 분석 실행

**자동 분석 항목**:

1. **프로젝트 유형 감지**
   ```bash
   # 신규 vs 기존 프로젝트 분류
   - 빈 디렉토리 → 신규 프로젝트
   - 코드/문서 존재 → 기존 프로젝트
   ```

2. **언어/프레임워크 자동 감지**
   ```bash
   # 파일 패턴 기반 언어 감지
   - pyproject.toml, requirements.txt → Python
   - package.json, tsconfig.json → TypeScript/Node.js
   - pom.xml, build.gradle → Java
   - go.mod → Go
   - Cargo.toml → Rust
   - backend/ + frontend/ → 풀스택
   ```

3. **문서 현황 분석**
   - 기존 `.moai/project/*.md` 파일 상태 확인
   - 부족한 정보 영역 식별
   - 보완 필요 항목 정리

4. **프로젝트 구조 평가**
   - 디렉토리 구조 복잡도
   - 단일 언어 vs 하이브리드 vs 마이크로서비스
   - 코드 기반 크기 추정

### 1.2 인터뷰 전략 수립

**프로젝트 유형별 질문 트리 선택**:

| 프로젝트 유형 | 질문 카테고리 | 중점 영역 |
|-------------|-------------|----------|
| **신규 프로젝트** | Product Discovery | 미션, 사용자, 해결 문제 |
| **기존 프로젝트** | Legacy Analysis | 코드 기반, 기술 부채, 통합점 |
| **TypeScript 전환** | Migration Strategy | 기존 프로젝트의 TypeScript 전환 |

**질문 우선순위**:
- **필수 질문**: 핵심 비즈니스 가치, 주요 사용자층 (모든 프로젝트)
- **기술 질문**: 언어/프레임워크, 품질 정책, 배포 전략
- **거버넌스**: 보안 요구사항, 추적성 전략 (선택적)

### 1.3 인터뷰 계획 보고서 생성

**사용자에게 제시할 계획서 포맷**:

```markdown
## 📊 프로젝트 초기화 계획: [PROJECT-NAME]

### 환경 분석 결과
- **프로젝트 유형**: [신규/기존/하이브리드]
- **감지된 언어**: [언어 목록]
- **현재 문서 상태**: [완성도 평가 0-100%]
- **구조 복잡도**: [단순/중간/복잡]

### 🎯 인터뷰 전략
- **질문 카테고리**: Product Discovery / Structure / Tech
- **예상 질문 수**: [N개 (필수 M개 + 선택 K개)]
- **우선순위 영역**: [중점적으로 다룰 영역]

### 🚨 주의사항
- **기존 문서**: [덮어쓰기 vs 보완 전략]
- **언어 설정**: [자동 감지 vs 수동 설정]
- **설정 충돌**: [기존 config.json과의 호환성]

### ✅ 예상 산출물
- **product.md**: [비즈니스 요구사항 문서]
- **structure.md**: [시스템 아키텍처 문서]
- **tech.md**: [기술 스택 및 정책 문서]
- **config.json**: [프로젝트 설정 파일]

---
**인터뷰를 시작합니다.**
```

### 1.4 프로젝트 초기화 진행

계획 보고 후 즉시 STEP 2로 진행합니다.

---

## 🚀 STEP 2: 프로젝트 초기화 실행

project-manager 에이전트가 체계적인 인터뷰를 통해 초기화를 수행합니다.

### 2.1 project-manager 에이전트 호출

```bash
# 에이전트 호출 패턴
@agent-project-manager 프로젝트 초기화를 시작합니다. 다음 정보를 기반으로 진행합니다:
- 프로젝트명: $ARGUMENTS
- 감지된 언어: [언어 목록]
- 프로젝트 유형: [신규/기존]
- 인터뷰 계획: [계획 요약]

체계적인 인터뷰를 진행하고 product/structure/tech.md 문서를 생성해주세요.
```

### 2.2 프로젝트 유형별 처리 방식

#### A. 신규 프로젝트 (그린필드)

**인터뷰 흐름**:

1. **Product Discovery** (product.md 작성)
   - 핵심 미션 정의 (@DOC:MISSION-001)
   - 주요 사용자층 파악 (@SPEC:USER-001)
   - 해결할 핵심 문제 식별 (@SPEC:PROBLEM-001)
   - 차별점 및 강점 정리 (@DOC:STRATEGY-001)
   - 성공 지표 설정 (@SPEC:SUCCESS-001)

2. **Structure Blueprint** (structure.md 작성)
   - 아키텍처 전략 선택 (@DOC:ARCHITECTURE-001)
   - 모듈별 책임 구분 (@DOC:MODULES-001)
   - 외부 시스템 통합 계획 (@DOC:INTEGRATION-001)
   - 추적성 전략 정의 (@DOC:TRACEABILITY-001)

3. **Tech Stack Mapping** (tech.md 작성)
   - 언어 & 런타임 선택 (@DOC:STACK-001)
   - 핵심 프레임워크 결정 (@DOC:FRAMEWORK-001)
   - 품질 게이트 설정 (@DOC:QUALITY-001)
   - 보안 정책 정의 (@DOC:SECURITY-001)
   - 배포 채널 계획 (@DOC:DEPLOY-001)

**config.json 자동 생성**:
```json
{
  "project_name": "detected-name",
  "project_type": "single|fullstack|microservice",
  "project_language": "python|typescript|java|go|rust",
  "test_framework": "pytest|vitest|junit|go test|cargo test",
  "linter": "ruff|biome|eslint|golint|clippy",
  "formatter": "black|biome|prettier|gofmt|rustfmt",
  "coverage_target": 85,
  "mode": "personal"
}
```

#### B. 기존 프로젝트 (레거시 도입)

**Legacy Snapshot & Alignment**:

1. **코드 기반 요약**
   - 디렉토리 구조 분석
   - 주요 모듈 및 책임 파악
   - 기존 테스트/빌드 파이프라인 확인

2. **기술 부채 및 제한 식별**
   - 오래된 프레임워크/의존성
   - 배포 환경 제약사항
   - 품질 지표 현황

3. **통합 지점 분석**
   - 외부 시스템 연동 현황
   - 인증 방식 및 보안 정책
   - 변경 불가 시스템 파악

4. **마이그레이션 전략 수립**
   - 명세 부재 영역 식별
   - 테스트 미비 항목 정리
   - 단계별 개선 계획 (@CODE:MIGRATION-XXX)

**보존 정책**:
- 기존 문서를 덮어쓰지 않고 부족한 부분만 보완
- 충돌하는 내용은 "Legacy Context" 섹션에 보존
- @CODE, TODO 태그로 개선 필요 항목 표시

### 2.3 문서 생성 및 검증

**산출물**:
- `.moai/project/product.md` (비즈니스 요구사항)
- `.moai/project/structure.md` (시스템 아키텍처)
- `.moai/project/tech.md` (기술 스택 및 정책)
- `.moai/config.json` (프로젝트 설정)

**품질 검증**:
- [ ] 모든 필수 @TAG 섹션 존재 확인
- [ ] EARS 구문 형식 준수 확인
- [ ] config.json 구문 유효성 검증
- [ ] 문서 간 일관성 검증

### 2.4 완료 보고

```markdown
✅ 프로젝트 초기화 완료!

📁 생성된 문서:
- .moai/project/product.md (비즈니스 정의)
- .moai/project/structure.md (아키텍처 설계)
- .moai/project/tech.md (기술 스택)
- .moai/config.json (프로젝트 설정)

🔍 감지된 환경:
- 언어: [언어 목록]
- 프레임워크: [프레임워크 목록]
- 테스트 도구: [도구 목록]

📋 다음 단계:
1. 생성된 문서를 검토하세요
2. /moai:1-spec으로 첫 번째 SPEC 작성
3. 필요 시 /moai:8-project update로 재조정
```

## 프로젝트 유형별 인터뷰 트리

### 1) 신규 프로젝트(그린필드) 전용 트리

#### Product Discovery

| 단계 | 핵심 질문 | 선택 분기 | 후속 질문/결정 |
|------|----------|----------|---------------|
| **A1. 미션 정의** | 제품이 약속하는 핵심 가치는 무엇인가요? | 생산성 ∙ 품질 ∙ 학습 ∙ 커뮤니티 | 선택한 가치에 따라 성공 지표 후보 메모 |
| **A2. 사용자층** | 주요 사용자는 누구인가요? | 개인 개발자 ∙ 팀/조직 ∙ 플랫폼 운영자 | 각 사용자군의 "즉시 얻고 싶은 결과" 작성 |
| **A3. 해결 문제** | 가장 시급한 문제 3가지는 무엇인가요? | 워크플로우 혼란 ∙ 품질 격차 ∙ 협업 지연 | 선택 항목마다 "현재 실패 사례" 기록 |
| **A4. 차별점** | 경쟁 솔루션 대비 강점은 무엇인가요? | 자동화 심도 ∙ 문서 동기화 ∙ 추적성 | 강점이 발휘되는 대표 시나리오 포함 |
| **A5. 성공 지표** | 성공을 입증할 첫 번째 지표는 무엇인가요? | 채택률 ∙ 반복 속도 ∙ 품질 | KPI에 측정 주기와 베이스라인 기입 |

#### Structure Blueprint

| 단계 | 질문 | 선택 분기 | 후속 결정 |
|------|------|----------|----------|
| **S1. 모듈 전략** | 시스템을 몇 개 모듈로 나눌까요? | 핵심 3모듈 ∙ 도메인별 확장 ∙ 마이크로서비스 ∙ 단일 핵심 | 선택 이유와 예외 규칙 명시 |
| **S2. 책임 구분** | 각 모듈이 담당할 책임은 무엇인가요? | 명세/요구 ∙ 구현/TDD ∙ 문서/PR | 책임별로 입력→처리→출력 흐름 표 정리 |
| **S3. 외부 연동** | 어떤 외부 시스템과 통신하나요? | Claude Code ∙ Git/GitHub ∙ CI/CD ∙ 패키지 레지스트리 | 연동당 인증 방식, 장애 시 대체 절차 기록 |
| **S4. 추적성 전략** | 요구사항 추적을 어떻게 보장하나요? | TAG 체계 ∙ Issue 템플릿 ∙ 계약 테스트 | 선택한 전략의 유지 주기 서술 |

#### Tech Stack Mapping

| 단계 | 질문 | 선택 분기 | 후속 질문 |
|------|------|----------|----------|
| **T1. 언어 & 런타임** | 어떤 언어/런타임 조합이 필요한가요? | Python ≥3.11 ∙ Node.js ≥18 ∙ JVM ∙ Go ∙ Rust | 다중 언어 선택 시 주요 모듈과 매핑 관계 표 작성 |
| **T2. 프레임워크** | 핵심 프레임워크/라이브러리는 무엇인가요? | Web(React/Vue) ∙ Backend(FastAPI/Spring) ∙ Mobile | 프레임워크별 빌드/테스트 명령과 CI 파이프라인 기록 |
| **T3. 품질 정책** | 어떤 품질 게이트를 적용하나요? | 테스트 커버리지 목표 ∙ 정적 분석 ∙ 포매터 | 실패 시 대응 규칙과 책임자 명시 |
| **T4. 보안·운영** | 보안 및 운영 정책은 무엇인가요? | 비밀 관리 ∙ 접근 제어 ∙ 로깅/관측 | 각 정책의 감사 대상 데이터와 보존 기간 기록 |
| **T5. 배포 채널** | 배포 타깃은 어디인가요? | 클라우드 ∙ 온프레미스 ∙ 모바일 스토어 ∙ 패키지 레지스트리 | 채널별 릴리스 절차와 롤백 전략 정의 |

### 2) 기존 프로젝트(레거시 도입) 전용 트리

#### Legacy Snapshot & Alignment

| 단계 | 핵심 질문 | 체크 포인트 | 후속 작업 |
|------|----------|------------|-----------|
| **L1. 코드 기반 요약** | 현재 저장소의 주요 디렉터리와 모듈은 무엇인가요? | src/app/packages 폴더 구조, 언어별 비중 | structure.md에 기존 모듈 구조 기록, 리팩터링 후보에 @CODE 태그 추가 |
| **L2. 빌드/테스트 파이프라인** | 어떤 빌드·테스트 절차가 존재하나요? | CI 설정, 스크립트, 커버리지·린트 현황 | tech.md 품질 섹션에 현재 상태와 목표 차이를 @CODE 또는 TODO 항목으로 명시 |
| **L3. 기술 부채 및 제한** | 유지해야 할 규칙이나 레거시 제약은 무엇인가요? | 오래된 프레임워크, 의존성 잠금 | product.md 또는 structure.md에 "현존 제약" 블록 작성 |
| **L4. 통합 지점** | 외부 시스템과의 연동 관계는 어떻게 구성되어 있나요? | 인증 방식, 사용 중인 API/Queue/DB | structure.md Integration 섹션에 현재 연결 방식과 위험도 기록 |
| **L5. 마이그레이션 전략** | MoAI-ADK로 전환하며 즉시 해야 할 작업은 무엇인가요? | 명세 부재, 테스트 미비 | 각 문서 끝에 "Initial Migration Plan" 목록 추가 |

## 🏷️ TAG 시스템 적용 규칙

**섹션별 @TAG 자동 생성**:

- 미션/비전 → @DOC:MISSION-XXX, @DOC:STRATEGY-XXX
- 사용자 정의 → @SPEC:USER-XXX, @SPEC:PERSONA-XXX
- 문제 분석 → @SPEC:PROBLEM-XXX, @SPEC:SOLUTION-XXX
- 아키텍처 → @DOC:ARCHITECTURE-XXX, @SPEC:PATTERN-XXX
- 기술 스택 → @DOC:STACK-XXX, @DOC:FRAMEWORK-XXX

**레거시 프로젝트 태그**:

- 기술 부채 → @CODE:REFACTOR-XXX, @CODE:TEST-XXX, @CODE:MIGRATION-XXX
- 해결 계획 → @CODE:MIGRATION-XXX, TODO:SPEC-BACKLOG-XXX
- 품질 개선 → TODO:TEST-COVERAGE-XXX, TODO:DOCS-SYNC-XXX

## 오류 처리

### 일반적인 오류 및 해결 방법

**오류 1**: 프로젝트 언어 감지 실패
```
증상: "언어를 감지할 수 없습니다" 메시지
해결: 수동으로 언어 지정 또는 언어별 설정 파일 생성
```

**오류 2**: 기존 문서와 충돌
```
증상: product.md가 이미 존재하며 내용이 다름
해결: "Legacy Context" 섹션에 기존 내용 보존 후 새 내용 추가
```

**오류 3**: config.json 작성 실패
```
증상: JSON 구문 오류 또는 권한 거부
해결: 파일 권한 확인 (chmod 644) 또는 수동으로 config.json 생성
```

## 금지 사항

**절대 하지 말아야 할 작업**:

- ❌ `.claude/memory/` 디렉토리에 파일 생성
- ❌ `.claude/commands/moai/*.json` 파일 생성
- ❌ 기존 문서 불필요한 덮어쓰기
- ❌ 날짜와 수치 예측 ("3개월 내", "50% 단축" 등)
- ❌ 가상의 시나리오, 예상 시장 규모, 미래 기술 트렌드 예측

**사용해야 할 표현**:

- ✅ "우선순위 높음/중간/낮음"
- ✅ "즉시 필요", "단계적 개선"
- ✅ 현재 확인 가능한 사실
- ✅ 기존 기술 스택
- ✅ 실제 문제점

## 다음 단계

초기화 완료 후:

- **신규 프로젝트**: `/moai:1-spec`을 실행해 설계 기반 SPEC 백로그 생성
- **레거시 프로젝트**: product/structure/tech 문서의 @CODE/@CODE/TODO 항목 검토 후 우선순위 확정
- **설정 변경**: `/moai:8-project update`로 인터뷰 다시 진행

## 관련 명령어

- `/moai:1-spec` - SPEC 작성 시작
- `/moai:9-update` - MoAI-ADK 업데이트
- `moai doctor` - 시스템 진단
- `moai status` - 프로젝트 상태 확인