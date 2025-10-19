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

### 1.0 백업 디렉토리 확인 (최우선)

**moai-adk init 재초기화 후 백업 파일 처리**

Alfred는 먼저 `.moai-backups/` 디렉토리를 확인합니다:

```bash
# 최신 백업 타임스탬프 확인
ls -t .moai-backups/ | head -1

# config.json의 optimized 플래그 확인
grep "optimized" .moai/config.json
```

**백업 존재 조건**:
- `.moai-backups/` 디렉토리 존재
- 최신 백업 폴더에 `.moai/project/*.md` 파일 존재
- `config.json`의 `optimized: false` (재초기화 직후)

**백업 존재 시 사용자 선택 (AskUserQuestion)**:

```typescript
AskUserQuestion({
  questions: [{
    question: "백업 파일(.moai-backups/{timestamp}/)이 발견되었습니다. 어떻게 처리하시겠습니까?",
    header: "백업 처리 방식",
    options: [
      {
        label: "병합",
        description: "백업 파일의 사용자 커스터마이징을 최신 템플릿에 병합 (권장)"
      },
      {
        label: "새로 작성",
        description: "백업 무시하고 새로운 인터뷰 시작 (처음부터 다시 작성)"
      },
      {
        label: "건너뛰기",
        description: "현재 파일 유지 (변경 없음, 작업 종료)"
      }
    ],
    multiSelect: false
  }]
})
```

**응답 처리**:
- **"병합"** → Phase 1.1 (백업 병합 워크플로우)로 진행
- **"새로 작성"** → Phase 1.2 (프로젝트 환경 분석)로 진행 (기존 프로세스)
- **"건너뛰기"** → 작업 종료

**백업 없음 또는 optimized: true**:
- Phase 1.2 (프로젝트 환경 분석)로 바로 진행

---

### 1.1 백업 병합 워크플로우 (사용자가 "병합" 선택 시)

**목적**: 최신 템플릿 구조를 유지하면서 사용자 커스터마이징 내용만 복원

**STEP 1: 백업 파일 읽기**

Alfred는 최신 백업 디렉토리에서 파일을 읽습니다:
```bash
# 최신 백업 디렉토리 경로
BACKUP_DIR=.moai-backups/$(ls -t .moai-backups/ | head -1)

# 백업 파일 읽기
Read $BACKUP_DIR/.moai/project/product.md
Read $BACKUP_DIR/.moai/project/structure.md
Read $BACKUP_DIR/.moai/project/tech.md
Read $BACKUP_DIR/CLAUDE.md
```

**STEP 2: 템플릿 기본값 탐지**

다음 패턴은 "템플릿 기본값"으로 간주 (병합하지 않음):
- "주요 사용자층을 정의하세요"
- "해결하려는 핵심 문제를 설명하세요"
- "프로젝트의 강점과 차별점을 나열하세요"
- "{{PROJECT_NAME}}", "{{PROJECT_DESCRIPTION}}" 등 변수 형식
- "예시:", "샘플:", "Example:" 등 가이드 문구

**STEP 3: 사용자 커스터마이징 추출**

백업 파일에서 **템플릿 기본값이 아닌 내용**만 추출:
- `product.md`:
  - USER 섹션의 실제 사용자층 정의
  - PROBLEM 섹션의 실제 문제 설명
  - STRATEGY 섹션의 실제 차별점
  - SUCCESS 섹션의 실제 성공 지표
- `structure.md`:
  - ARCHITECTURE 섹션의 실제 설계
  - MODULES 섹션의 실제 모듈 구조
  - INTEGRATION 섹션의 실제 통합 계획
- `tech.md`:
  - STACK 섹션의 실제 기술 스택
  - FRAMEWORK 섹션의 실제 프레임워크
  - QUALITY 섹션의 실제 품질 정책
- `HISTORY` 섹션: **전체 보존** (모든 파일)

**STEP 4: 병합 전략**

```markdown
최신 템플릿 구조 (v0.4.0+)
    ↓
사용자 커스터마이징 삽입 (백업 파일에서 추출)
    ↓
HISTORY 섹션 업데이트
    ↓
버전 업데이트 (v0.1.x → v0.1.x+1)
```

**병합 원칙**:
- ✅ 템플릿 구조는 최신 버전 유지 (섹션 순서, 헤더, @TAG 형식)
- ✅ 사용자 커스터마이징만 삽입 (실제 작성한 내용)
- ✅ HISTORY 섹션 누적 보존 (기존 이력 + 병합 이력)
- ❌ 템플릿 기본값은 최신 버전으로 교체

**STEP 5: HISTORY 섹션 업데이트**

병합 완료 후 각 파일의 HISTORY 섹션에 이력 추가:
```yaml
### v0.1.x+1 (2025-10-19)
- **UPDATED**: 백업 파일 병합 (자동 최적화)
- AUTHOR: @Alfred
- BACKUP: .moai-backups/20251018-003638/
- REASON: moai-adk init 재초기화 후 사용자 커스터마이징 복원
```

**STEP 6: config.json 업데이트**

병합 완료 후 최적화 플래그 설정:
```json
{
  "project": {
    "optimized": true,
    "last_merge": "2025-10-19T12:34:56+09:00",
    "backup_source": ".moai-backups/20251018-003638/"
  }
}
```

**STEP 7: 완료 보고**

```markdown
✅ 백업 병합 완료!

📁 병합된 파일:
- .moai/project/product.md (v0.1.4 → v0.1.5)
- .moai/project/structure.md (v0.1.1 → v0.1.2)
- .moai/project/tech.md (v0.1.1 → v0.1.2)
- .moai/config.json (optimized: false → true)

🔍 병합 내역:
- USER 섹션: 백업 파일의 사용자 정의 내용 복원
- PROBLEM 섹션: 백업 파일의 문제 설명 복원
- STRATEGY 섹션: 백업 파일의 차별점 복원
- HISTORY 섹션: 병합 이력 추가 (누적 보존)

💾 백업 파일 위치:
- 원본 백업: .moai-backups/20251018-003638/
- 보존 기간: 영구 (수동 삭제 전까지)

📋 다음 단계:
1. 병합된 문서를 검토하세요
2. 필요 시 추가 수정
3. /alfred:1-spec으로 첫 번째 SPEC 작성

---
**작업 완료: /alfred:0-project 종료**
```

**병합 후 작업 종료**: 인터뷰 없이 바로 완료

---

### 1.2 프로젝트 환경 분석 실행 (사용자가 "새로 작성" 선택 시 또는 백업 없음)

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

### 1.3 인터뷰 전략 수립 (사용자가 "새로 작성" 선택 시)

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

### 1.4 인터뷰 계획 보고서 생성 (사용자가 "새로 작성" 선택 시)

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

### 1.5 사용자 승인 대기 (AskUserQuestion) (사용자가 "새로 작성" 선택 시)

Alfred는 project-manager의 인터뷰 계획 보고서를 받은 후, **AskUserQuestion 도구를 호출하여 사용자 승인을 받습니다**:

```typescript
AskUserQuestion({
  questions: [{
    question: "project-manager가 제시한 인터뷰 계획으로 프로젝트 초기화를 진행하시겠습니까?",
    header: "Phase 2 승인",
    options: [
      { label: "진행", description: "승인된 계획대로 인터뷰 및 문서 생성 시작" },
      { label: "수정", description: "계획 재수립 (Phase 1 반복)" },
      { label: "중단", description: "프로젝트 초기화 중단" }
    ],
    multiSelect: false
  }]
})
```

**응답 처리**:
- **"진행"** (`answers["0"] === "진행"`) → Phase 2 실행
- **"수정"** (`answers["0"] === "수정"`) → Phase 1 반복 (project-manager 재호출)
- **"중단"** (`answers["0"] === "중단"`) → 작업 종료

---

## 🚀 STEP 2: 프로젝트 초기화 실행 (사용자 "새로 작성" 승인 후)

**주의**: 이 단계는 사용자가 **"새로 작성"을 선택한 경우**에만 실행됩니다.
- "병합" 선택 시: Phase 1.1 (백업 병합)에서 작업 종료
- "건너뛰기" 선택 시: 작업 종료
- "새로 작성" 선택 시: 아래 프로세스 진행

사용자 승인 후 project-manager 에이전트가 초기화를 수행합니다.

### 2.1 project-manager 에이전트 호출 (사용자가 "새로 작성" 선택 시)

Alfred는 project-manager 에이전트를 호출하여 프로젝트 초기화를 시작합니다. 다음 정보를 기반으로 진행합니다:
- 감지된 언어: [언어 목록]
- 프로젝트 유형: [신규/기존]
- 기존 문서 상태: [존재/부재]
- 승인된 인터뷰 계획: [계획 요약]

에이전트는 체계적인 인터뷰를 진행하고 product/structure/tech.md 문서를 생성/갱신합니다.

### 2.2 Alfred Skills 자동 활성화 (선택적)

project-manager가 문서를 생성 완료한 후, **Alfred는 선택적으로 Skills를 호출할 수 있습니다** (사용자 요청 시).

**자동 활성화 조건** (선택적):

| 조건 | 자동 선택 Skill | 목적 |
|------|----------------|------|
| 사용자 "품질 검증" 요청 | moai-alfred-trust-validation | 초기 프로젝트 구조 검증 |

**실행 흐름** (선택적):
```
1. project-manager 완료
    ↓
2. 사용자 선택:
   - "품질 검증 필요" → moai-alfred-trust-validation (Level 1 빠른 스캔)
   - "건너뛰기" → 바로 완료
```

**참고**: 프로젝트 초기화 단계에서는 품질 검증이 선택사항입니다.

### 2.3 Sub-agent AskUserQuestion (Nested)

**project-manager 에이전트는 내부적으로 AskUserQuestion을 호출**하여 세부 작업을 확인할 수 있습니다.

**호출 시점**:
- 기존 프로젝트 문서 덮어쓰기 전
- 언어/프레임워크 선택 시
- 중요한 설정 변경 시

**예시** (project-manager 내부):
```typescript
AskUserQuestion({
  questions: [{
    question: "기존 product.md 파일이 존재합니다. 어떻게 처리하시겠습니까?",
    header: "파일 덮어쓰기 확인",
    options: [
      { label: "덮어쓰기", description: "기존 파일 백업 후 새 내용으로 교체" },
      { label: "병합", description: "기존 내용과 새 내용 병합" },
      { label: "건너뛰기", description: "기존 파일 유지" }
    ],
    multiSelect: false
  }]
})
```

**Nested 패턴**:
- **커맨드 레벨** (Phase 승인): Alfred가 호출 → "Phase 2 진행할까요?"
- **Sub-agent 레벨** (세부 확인): project-manager가 호출 → "파일 덮어쓸까요?"

### 2.4 프로젝트 유형별 처리 방식

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

---

## /alfred:0-project update: 템플릿 최적화 (서브커맨드)

> **목적**: moai-adk update 실행 후 백업과 신규 템플릿을 비교하여 사용자 커스터마이징을 보존하면서 템플릿을 최적화합니다.

### 실행 조건

이 서브커맨드는 다음 조건에서 실행됩니다:

1. **moai-adk update 실행 후**: `config.json`의 `optimized=false` 상태
2. **템플릿 업데이트 필요**: 백업과 신규 템플릿 간 차이가 있을 때
3. **사용자 명시적 요청**: 사용자가 직접 `/alfred:0-project update` 실행

### 실행 흐름

#### Phase 1: 백업 분석 및 비교

1. **최신 백업 확인**:
   ```bash
   # .moai-backups/ 디렉토리에서 최신 백업 탐색
   ls -lt .moai-backups/ | head -1
   ```

2. **변경 사항 분석**:
   - 백업의 `.claude/` 디렉토리와 현재 템플릿 비교
   - 백업의 `.moai/project/` 문서와 현재 문서 비교
   - 사용자 커스터마이징 항목 식별

3. **비교 보고서 생성**:
   ```markdown
   ## 📊 템플릿 최적화 분석

   ### 변경 항목
   - CLAUDE.md: "## 프로젝트 정보" 섹션 보존 필요
   - settings.json: env 변수 3개 보존 필요
   - product.md: 사용자 작성 내용 있음

   ### 권장 조치
   - 스마트 병합 실행
   - 사용자 커스터마이징 보존
   - optimized=true 설정
   ```

4. **사용자 승인 대기** (AskUserQuestion):
   - 질문: "템플릿 최적화를 진행하시겠습니까?"
   - 옵션:
     - "진행" → Phase 2 실행
     - "미리보기" → 상세 변경 내역 표시 후 재확인
     - "건너뛰기" → optimized=false 유지

#### Phase 2: 스마트 병합 실행 (사용자 승인 후)

1. **스마트 병합 로직 실행**:
   - `TemplateProcessor.copy_templates()` 실행
   - CLAUDE.md: "## 프로젝트 정보" 섹션 보존
   - settings.json: env 변수 및 permissions.allow 병합

2. **optimized=true 설정**:
   ```python
   # config.json 업데이트
   config_data["project"]["optimized"] = True
   ```

3. **최적화 완료 보고**:
   ```markdown
   ✅ 템플릿 최적화 완료!

   📄 병합된 파일:
   - CLAUDE.md (프로젝트 정보 보존)
   - settings.json (env 변수 보존)

   ⚙️ config.json: optimized=true 설정 완료
   ```

### Alfred 자동화 전략

**Alfred 자동 판단**:
- project-manager 에이전트 자동 호출
- 백업 최신성 확인 (24시간 이내)
- 변경 사항 자동 분석

**Skills 자동 활성화**:
- moai-alfred-tag-scanning: TAG 체인 검증
- moai-alfred-trust-validation: TRUST 원칙 준수 확인

### 실행 예시

```bash
# moai-adk update 실행 후
moai-adk update

# 출력:
# ✓ Update complete!
# ℹ️  Next step: Run /alfred:0-project update to optimize template changes

# Alfred 실행
/alfred:0-project update

# → Phase 1: 백업 분석 및 비교 보고서 생성
# → 사용자 승인 대기
# → Phase 2: 스마트 병합 실행, optimized=true 설정
```

### 주의사항

- **백업 필수**: `.moai-backups/` 디렉토리에 백업이 없으면 실행 불가
- **수동 검토 권장**: 중요한 커스터마이징이 있다면 미리보기 확인 필수
- **충돌 해결**: 병합 충돌 발생 시 사용자 선택 요청

---

## 🚀 STEP 3: 프로젝트 맞춤형 최적화 (선택적)

**실행 조건**:
- Phase 2 (프로젝트 초기화) 완료 후
- 또는 Phase 1.1 (백업 병합) 완료 후
- 사용자가 명시적으로 요청하거나 Alfred가 자동 판단

**목적**: 프로젝트 특성에 맞는 Commands, Agents, Skills만 선택하여 경량화 (37개 스킬 → 3~5개)

### 3.1 Feature Selection 자동 실행

**Alfred는 moai-alfred-feature-selector 스킬을 자동 호출**합니다:

**스킬 입력**:
- `.moai/project/product.md` (프로젝트 카테고리 힌트)
- `.moai/project/tech.md` (주 언어, 프레임워크)
- `.moai/config.json` (프로젝트 설정)

**스킬 출력**:
```json
{
  "category": "web-api",
  "language": "python",
  "framework": "fastapi",
  "commands": ["1-spec", "2-build", "3-sync"],
  "agents": ["spec-builder", "code-builder", "doc-syncer", "git-manager", "debug-helper"],
  "skills": ["moai-lang-python", "moai-domain-web-api", "moai-domain-backend"],
  "excluded_skills_count": 34,
  "optimization_rate": "87%"
}
```

**실행 방법**:
```
Alfred: Skill("moai-alfred-feature-selector")
```

---

### 3.2 Template Generation 자동 실행

**Alfred는 moai-alfred-template-generator 스킬을 자동 호출**합니다:

**스킬 입력**:
- `.moai/.feature-selection.json` (feature-selector 출력)
- `CLAUDE.md` 템플릿
- 전체 commands/agents/skills 파일

**스킬 출력**:
- `CLAUDE.md` (맞춤형 에이전트 테이블 - 선택된 에이전트만)
- `.claude/commands/` (선택된 commands만)
- `.claude/agents/` (선택된 agents만)
- `.claude/skills/` (선택된 skills만)
- `.moai/config.json` (`optimized: true` 업데이트)

**실행 방법**:
```
Alfred: Skill("moai-alfred-template-generator")
```

---

### 3.3 최적화 완료 보고

**보고 형식**:
```markdown
✅ 프로젝트 맞춤형 최적화 완료!

📊 최적화 결과:
- **프로젝트**: {{PROJECT_NAME}}
- **카테고리**: web-api
- **주 언어**: python
- **프레임워크**: fastapi

🎯 선택된 기능:
- Commands: 4개 (0-project, 1-spec, 2-build, 3-sync)
- Agents: 5개 (spec-builder, code-builder, doc-syncer, git-manager, debug-helper)
- Skills: 3개 (moai-lang-python, moai-domain-web-api, moai-domain-backend)

💡 경량화 효과:
- 제외된 스킬: 34개
- 경량화: 87%
- CLAUDE.md: 맞춤형 에이전트 테이블 생성

📋 다음 단계:
1. CLAUDE.md 파일 확인 (5개 에이전트만 표시)
2. /alfred:1-spec "첫 기능" 실행
3. MoAI-ADK 워크플로우 시작
```

---

### 3.4 Phase 3 건너뛰기 (선택적)

**사용자는 Phase 3를 건너뛸 수 있습니다**:

**건너뛰기 조건**:
- 사용자가 명시적으로 "건너뛰기" 선택
- Alfred 자동 판단 시 "간단한 프로젝트" (기본 기능만 필요)

**건너뛰기 효과**:
- 전체 37개 스킬 유지 (경량화 없음)
- CLAUDE.md 템플릿 기본 9개 에이전트 유지
- config.json의 `optimized: false` 유지

---

## 다음 단계

**권장사항**: 다음 단계 진행 전 `/clear` 또는 `/new` 명령으로 새로운 대화 세션을 시작하면 더 나은 성능과 컨텍스트 관리를 경험할 수 있습니다.

초기화 완료 후:

- **신규 프로젝트**: `/alfred:1-spec`을 실행해 설계 기반 SPEC 백로그 생성
- **레거시 프로젝트**: product/structure/tech 문서의 @CODE/@CODE/TODO 항목 검토 후 우선순위 확정
- **설정 변경**: `/alfred:0-project`를 다시 실행하여 문서 갱신
- **템플릿 최적화**: `moai-adk update` 후 `/alfred:0-project update` 실행

## 관련 명령어

- `/alfred:1-spec` - SPEC 작성 시작
- `/alfred:9-update` - MoAI-ADK 업데이트
- `moai doctor` - 시스템 진단
- `moai status` - 프로젝트 상태 확인