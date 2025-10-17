---
name: alfred:0-project
description: 프로젝트 문서 초기화 - product/structure/tech.md 생성 및 언어별 최적화 설정
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Grep
  - Glob
  - TodoWrite
  - Bash(ls:*)
  - Bash(find:*)
  - Bash(cat:*)
  - Task
---

# 📋 MoAI-ADK 0단계: 범용 언어 지원 프로젝트 문서 초기화/갱신

## 🎯 커맨드 목적

프로젝트 환경을 자동 분석하여 product/structure/tech.md 문서를 생성/갱신하고 언어별 최적화 설정을 구성합니다.

## 📋 실행 흐름

1. **환경 분석**: 프로젝트 유형(신규/레거시) 및 언어 자동 감지
2. **인터뷰 전략 수립**: 프로젝트 특성에 맞는 질문 트리 선택
3. **사용자 확인**: 인터뷰 계획 검토 및 승인
4. **프로젝트 문서 작성**: product/structure/tech.md 생성
5. **설정 파일 생성**: config.json 자동 구성

## 🔗 연관 에이전트

- **Primary**: project-manager (📋 기획자) - 프로젝트 초기화 전담
- **Quality Check**: trust-checker (✅ 품질 보증 리드) - 초기 구조 검증 (선택적)
- **Secondary**: None (독립 실행)

## 💡 사용 예시

사용자가 `/alfred:8-project` 커맨드를 실행하여 프로젝트 분석 및 문서 생성/갱신을 수행합니다.

## 명령어 개요

프로젝트 환경을 분석하고 product/structure/tech.md 문서를 생성/갱신하는 체계적인 초기화 시스템입니다.

- **언어 자동 감지**: Python, TypeScript, Java, Go, Rust 등 자동 인식
- **프로젝트 유형 분류**: 신규 vs 기존 프로젝트 자동 판단
- **고성능 초기화**: TypeScript 기반 CLI로 0.18초 초기화 달성
- **2단계 워크플로우**: 1) 분석 및 계획 → 2) 사용자 승인 후 실행

## 사용법

사용자가 `/alfred:8-project` 커맨드를 실행하여 프로젝트 분석 및 문서 생성/갱신을 시작합니다.

**자동 처리**:
- 기존 `.moai/project/` 문서가 있으면 갱신 모드
- 문서가 없으면 신규 생성 모드
- 언어 및 프로젝트 유형은 자동 감지

## ⚠️ 금지 사항

**절대 하지 말아야 할 작업**:

- ❌ `.claude/memory/` 디렉토리에 파일 생성
- ❌ `.claude/commands/alfred/*.json` 파일 생성
- ❌ 기존 문서 불필요한 덮어쓰기
- ❌ 날짜와 수치 예측 ("3개월 내", "50% 단축" 등)
- ❌ 가상의 시나리오, 예상 시장 규모, 미래 기술 트렌드 예측

**사용해야 할 표현**:

- ✅ "우선순위 높음/중간/낮음"
- ✅ "즉시 필요", "단계적 개선"
- ✅ 현재 확인 가능한 사실
- ✅ 기존 기술 스택
- ✅ 실제 문제점

## 🚀 STEP 1: 환경 분석 및 인터뷰 계획 수립

프로젝트 환경을 분석하고 체계적인 인터뷰 계획을 수립합니다.

### 1.1 프로젝트 환경 분석 실행

**자동 분석 항목**:

1. **프로젝트 유형 감지**
   Alfred는 디렉토리 구조를 분석하여 신규 vs 기존 프로젝트를 분류합니다:
   - 빈 디렉토리 → 신규 프로젝트
   - 코드/문서 존재 → 기존 프로젝트

2. **언어/프레임워크 자동 감지**: 파일 패턴을 기반으로 프로젝트의 주요 언어를 감지합니다
   - pyproject.toml, requirements.txt → Python
   - package.json, tsconfig.json → TypeScript/Node.js
   - pom.xml, build.gradle → Java
   - go.mod → Go
   - Cargo.toml → Rust
   - backend/ + frontend/ → 풀스택

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
- **예상 소요시간**: [시간 산정]
- **우선순위 영역**: [중점적으로 다룰 영역]

### ⚠️ 주의사항
- **기존 문서**: [덮어쓰기 vs 보완 전략]
- **언어 설정**: [자동 감지 vs 수동 설정]
- **설정 충돌**: [기존 config.json과의 호환성]

### ✅ 예상 산출물
- **product.md**: [비즈니스 요구사항 문서]
- **structure.md**: [시스템 아키텍처 문서]
- **tech.md**: [기술 스택 및 정책 문서]
- **config.json**: [프로젝트 설정 파일]

---
**승인 요청**: 위 계획으로 인터뷰를 진행하시겠습니까?
("진행", "수정 [내용]", "중단" 중 선택)
```

### 1.4 사용자 확인 대기

**반응에 따른 분기**:
- **"진행"** 또는 **"시작"**: STEP 2로 진행
- **"수정 [내용]"**: 계획 수정 후 재제시
- **"중단"**: 프로젝트 초기화 중단

---

## 🚀 STEP 2: 프로젝트 초기화 실행 (사용자 승인 후)

사용자 승인 후 project-manager 에이전트가 초기화를 수행합니다.

### 2.1 project-manager 에이전트 호출

Alfred는 project-manager 에이전트를 호출하여 프로젝트 초기화를 시작합니다. 다음 정보를 기반으로 진행합니다:
- 감지된 언어: [언어 목록]
- 프로젝트 유형: [신규/기존]
- 기존 문서 상태: [존재/부재]
- 승인된 인터뷰 계획: [계획 요약]

에이전트는 체계적인 인터뷰를 진행하고 product/structure/tech.md 문서를 생성/갱신합니다.

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

**STEP 1: 전체 프로젝트 구조 파악**

Alfred는 전체 프로젝트 구조를 파악합니다:
- tree 명령어 또는 find 명령어를 사용하여 디렉토리 구조 시각화
- node_modules, .git, dist, build, __pycache__ 등 빌드 산출물 제외
- 주요 소스 디렉토리 및 설정 파일 식별

**산출물**:
- 프로젝트 전체 폴더/파일 계층 구조 시각화
- 주요 디렉토리 식별 (src/, tests/, docs/, config/ 등)
- 언어/프레임워크 힌트 파일 확인 (package.json, pyproject.toml, go.mod 등)

**STEP 2: 병렬 분석 전략 수립**

Alfred는 Glob 패턴으로 파일 그룹을 식별합니다:
1. **설정 파일들**: *.json, *.toml, *.yaml, *.yml, *.config.js
2. **소스 코드 파일들**: src/**/*.{ts,js,py,go,rs,java}
3. **테스트 파일들**: tests/**/*.{ts,js,py,go,rs,java}, **/*.test.*, **/*.spec.*
4. **문서 파일들**: *.md, docs/**/*.md, README*, CHANGELOG*

**병렬 Read 전략**:
- 여러 파일을 동시에 Read 도구로 읽어 분석 속도 향상
- 각 파일 그룹별로 배치 처리
- 우선순위: 설정 파일 → 핵심 소스 → 테스트 → 문서

**STEP 3: 파일별 특성 분석 및 보고**

각 파일을 읽으면서 다음 정보를 수집:

1. **설정 파일 분석**
   - 프로젝트 메타데이터 (이름, 버전, 설명)
   - 의존성 목록 및 버전
   - 빌드/테스트 스크립트
   - 언어/프레임워크 확정

2. **소스 코드 분석**
   - 주요 모듈 및 클래스 식별
   - 아키텍처 패턴 추론 (MVC, 클린 아키텍처, 마이크로서비스 등)
   - 외부 API 호출 및 통합점 파악
   - 도메인 로직 핵심 영역

3. **테스트 코드 분석**
   - 테스트 프레임워크 확인
   - 커버리지 설정 파악
   - 주요 테스트 시나리오 식별
   - TDD 준수 여부 평가

4. **문서 분석**
   - 기존 README 내용
   - 아키텍처 문서 존재 여부
   - API 문서 현황
   - 설치/배포 가이드 완성도

**보고 형식**:
```markdown
## 파일별 분석 결과

### 설정 파일
- package.json: Node.js 18+, TypeScript 5.x, Vitest 테스트
- tsconfig.json: strict 모드, ESNext 타겟
- biome.json: 린터/포매터 설정 존재

### 소스 코드 (src/)
- src/core/: 핵심 비즈니스 로직 (3개 모듈)
- src/api/: REST API 엔드포인트 (5개 라우터)
- src/utils/: 유틸리티 함수 (로깅, 검증 등)
- 아키텍처: 계층형 (controller → service → repository)

### 테스트 (tests/)
- Vitest + @testing-library 사용
- 유닛 테스트 커버리지 약 60% 추정
- E2E 테스트 미비

### 문서
- README.md: 설치 가이드만 존재
- API 문서 부재
- 아키텍처 문서 부재
```

**STEP 4: 종합 분석 및 product/structure/tech 반영**

수집된 정보를 바탕으로 3대 문서에 반영:

1. **product.md 반영 내용**
   - 기존 README/문서에서 추출한 프로젝트 미션
   - 코드에서 추론한 주요 사용자층 및 시나리오
   - 해결하는 핵심 문제 역추적
   - 기존 자산을 "Legacy Context"에 보존

2. **structure.md 반영 내용**
   - 파악된 실제 디렉토리 구조
   - 모듈별 책임 분석 결과
   - 외부 시스템 통합점 (API 호출, DB 연결 등)
   - 기술 부채 항목 (@CODE 태그로 표기)

3. **tech.md 반영 내용**
   - 실제 사용 중인 언어/프레임워크/라이브러리
   - 기존 빌드/테스트 파이프라인
   - 품질 게이트 현황 (린터, 포매터, 테스트 커버리지)
   - 보안/배포 정책 파악
   - 개선 필요 항목 (TODO 태그로 표기)

**보존 정책**:
- 기존 문서를 덮어쓰지 않고 부족한 부분만 보완
- 충돌하는 내용은 "Legacy Context" 섹션에 보존
- @CODE, TODO 태그로 개선 필요 항목 표시

**최종 보고서 예시**:
```markdown
## 기존 프로젝트 분석 완료

### 환경 정보
- **언어**: TypeScript 5.x (Node.js 18+)
- **프레임워크**: Express.js
- **테스트**: Vitest (커버리지 ~60%)
- **린터/포매터**: Biome

### 주요 발견사항
1. **강점**:
   - 타입 안전성 높음 (strict 모드)
   - 모듈 구조 명확 (core/api/utils 분리)

2. **개선 필요**:
   - 테스트 커버리지 85% 미달 (TODO:TEST-COVERAGE-001)
   - API 문서 부재 (TODO:DOCS-API-001)
   - E2E 테스트 미비 (@CODE:TEST-E2E-001)

### 다음 단계
1. product/structure/tech.md 생성 완료
2. @CODE/TODO 항목 우선순위 확정
3. /alfred:1-spec으로 개선 SPEC 작성 시작
```

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
2. /alfred:1-spec으로 첫 번째 SPEC 작성
3. 필요 시 /alfred:8-project update로 재조정
```

### 2.5: 초기 구조 검증 (선택적)

프로젝트 초기화 완료 후 선택적으로 품질 검증을 실행할 수 있습니다.

**실행 조건**: 사용자가 명시적으로 요청한 경우에만

**검증 목적**:
- 프로젝트 문서와 설정 파일 기본 검증
- 초기 구조의 TRUST 원칙 준수 확인
- 설정 파일 유효성 검증

**실행 방식**:
사용자가 명시적으로 요청한 경우에만 Alfred가 trust-checker 에이전트를 호출하여 프로젝트 초기 구조 검증을 수행합니다.

**검증 항목**:
- **문서 완성도**: product/structure/tech.md 필수 섹션 존재 확인
- **설정 유효성**: config.json JSON 구문 및 필수 필드 검증
- **TAG 체계**: 문서 내 @TAG 형식 준수 확인
- **EARS 구문**: SPEC 작성 시 사용할 EARS 템플릿 검증

**검증 실행**: Level 1 빠른 스캔 (3-5초)

**검증 결과 처리**:

✅ **Pass**: 다음 단계 진행 가능
- 문서와 설정 모두 정상

⚠️ **Warning**: 경고 표시 후 진행
- 일부 선택적 섹션 누락
- 권장사항 미적용

❌ **Critical**: 수정 필요
- 필수 섹션 누락
- config.json 구문 오류
- 사용자 선택: "수정 후 재검증" 또는 "건너뛰기"

**검증 건너뛰기**:
- 기본적으로 검증은 실행되지 않음
- 사용자가 명시적으로 요청할 때만 실행


## 프로젝트 유형별 인터뷰 가이드

### 신규 프로젝트 인터뷰 영역

**Product Discovery** (product.md)
- 핵심 미션 및 가치 제안
- 주요 사용자층 및 니즈
- 해결할 핵심 문제 3가지
- 경쟁 솔루션 대비 차별점
- 측정 가능한 성공 지표

**Structure Blueprint** (structure.md)
- 시스템 아키텍처 전략
- 모듈 분리 및 책임 구분
- 외부 시스템 통합 계획
- @TAG 기반 추적성 전략

**Tech Stack Mapping** (tech.md)
- 언어/런타임 선택 및 버전
- 프레임워크 및 라이브러리
- 품질 게이트 정책 (커버리지, 린터)
- 보안 정책 및 배포 채널

### 기존 프로젝트 인터뷰 영역

**Legacy Analysis**
- 현재 코드 구조 및 모듈 파악
- 빌드/테스트 파이프라인 현황
- 기술 부채 및 제약사항 식별
- 외부 연동 및 인증 방식
- MoAI-ADK 전환 우선순위 계획

**보존 정책**: 기존 문서는 "Legacy Context" 섹션에 보존하고 @CODE/TODO 태그로 개선 필요 항목 표시

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

## 다음 단계

**권장사항**: 다음 단계 진행 전 `/clear` 또는 `/new` 명령으로 새로운 대화 세션을 시작하면 더 나은 성능과 컨텍스트 관리를 경험할 수 있습니다.

초기화 완료 후:

- **신규 프로젝트**: `/alfred:1-spec`을 실행해 설계 기반 SPEC 백로그 생성
- **레거시 프로젝트**: product/structure/tech 문서의 @CODE/@CODE/TODO 항목 검토 후 우선순위 확정
- **설정 변경**: `/alfred:8-project`를 다시 실행하여 문서 갱신

## 관련 명령어

- `/alfred:1-spec` - SPEC 작성 시작
- `/alfred:9-update` - MoAI-ADK 업데이트
- `moai doctor` - 시스템 진단
- `moai status` - 프로젝트 상태 확인