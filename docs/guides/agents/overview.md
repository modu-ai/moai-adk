# Alfred Agent Ecosystem

Alfred가 조율하는 9개의 전문 에이전트 생태계입니다.

## Alfred: Your MoAI SuperAgent

**Alfred**는 모두의AI(MoAI)가 설계한 MoAI-ADK의 공식 SuperAgent입니다.

### Alfred 페르소나

- **정체성**: 모두의 AI 집사 ▶◀
- **성격**: 정확하고 예의 바르며 체계적
- **역할**: MoAI-ADK 워크플로우의 중앙 오케스트레이터
- **책임**: 사용자 요청 분석 → 에이전트 위임 → 결과 통합
- **목표**: SPEC-First TDD 방법론을 통한 완벽한 코드 품질 보장

### Core Responsibilities

1. **중앙 오케스트레이터**
   - 사용자 요청 분석 및 본질 파악
   - 적절한 전문 에이전트에게 작업 위임
   - 단일/순차/병렬 실행 전략 결정

2. **품질 게이트 관리**
   - TRUST 5원칙 자동 검증
   - TAG 체인 무결성 확인
   - 예외 발생 시 debug-helper 자동 호출

3. **워크플로우 자동화**
   - 3단계 개발 사이클 조율 (1-spec → 2-build → 3-sync)
   - Git 워크플로우 자동화
   - PR 생성 및 머지 관리 (Team 모드)

---

## Orchestration Strategy

Alfred의 작업 분배 전략:

```
사용자 요청
    ↓
Alfred 분석 (요청 본질 파악)
    ↓
작업 분해 및 라우팅
    ├─→ 직접 처리 (간단한 조회, 파일 읽기)
    ├─→ Single Agent (단일 전문가 위임)
    ├─→ Sequential (순차 실행: 1-spec → 2-build → 3-sync)
    └─→ Parallel (병렬 실행: 테스트 + 린트 + 빌드)
    ↓
품질 게이트 검증
    ├─→ TRUST 5원칙 준수 확인
    ├─→ @TAG 체인 무결성 검증
    └─→ 예외 발생 시 debug-helper 자동 호출
    ↓
Alfred가 결과 통합 보고
```

---

## 9 Specialized Agents

Alfred가 조율하는 9개의 전문 에이전트. 각 에이전트는 IT 전문가 직무에 매핑되어 있습니다.

### Primary Workflow Agents (3개)

핵심 3단계 워크플로우를 담당하는 에이전트들:

| Agent | Persona | Expertise | Command |
|-------|---------|-----------|---------|
| 🏗️ **spec-builder** | 시스템 아키텍트 | SPEC 작성, EARS 명세 | `/alfred:1-spec` |
| 💎 **code-builder** | 수석 개발자 | TDD 구현, 코드 품질 | `/alfred:2-build` |
| 📖 **doc-syncer** | 테크니컬 라이터 | 문서 동기화, Living Doc | `/alfred:3-sync` |

### Support Agents (4개)

온디맨드로 호출되는 지원 에이전트들:

| Agent | Persona | Expertise | Invocation |
|-------|---------|-----------|------------|
| 🏷️ **tag-agent** | 지식 관리자 | TAG 시스템, 추적성 | `@agent-tag-agent` |
| 🚀 **git-manager** | 릴리스 엔지니어 | Git 워크플로우, 배포 | `@agent-git-manager` |
| 🔬 **debug-helper** | 트러블슈팅 전문가 | 오류 진단, 해결 | `@agent-debug-helper` |
| ✅ **trust-checker** | 품질 보증 리드 | TRUST 검증, 성능/보안 | `@agent-trust-checker` |

### System Agents (2개)

시스템 관리를 담당하는 에이전트들:

| Agent | Persona | Expertise | Command |
|-------|---------|-----------|---------|
| 🛠️ **cc-manager** | 데브옵스 엔지니어 | Claude Code 설정 | `@agent-cc-manager` |
| 📋 **project-manager** | 프로젝트 매니저 | 프로젝트 초기화 | `/alfred:0-project` |

---

## Agent Collaboration Principles

### 1. Single Responsibility (단일 책임 원칙)

각 에이전트는 자신의 전문 영역만 담당합니다.

**예시**:

- `spec-builder`는 SPEC 작성만 담당 (구현은 `code-builder`가 담당)
- `doc-syncer`는 문서 동기화만 담당 (TAG 관리는 `tag-agent`가 담당)

### 2. Central Orchestration (중앙 조율)

**오직 Alfred만이** 에이전트 간 작업을 조율합니다. 에이전트 간 직접 호출은 금지됩니다.

**올바른 예**:

```
User → Alfred → spec-builder → Alfred → code-builder → Alfred → User
```

**잘못된 예**:

```
User → spec-builder → code-builder → User  # 직접 호출 금지
```

### 3. Quality Gates (품질 게이트)

각 단계 완료 시 Alfred가 자동으로 검증합니다:

- **1-spec 완료 후**: SPEC 메타데이터 검증
- **2-build 완료 후**: TRUST 원칙 검증
- **3-sync 완료 후**: TAG 체인 무결성 검증

### 4. Command Priority (커맨드 우선순위)

**커맨드 지침** > **에이전트 지침**

충돌 시 커맨드 지침을 우선 적용합니다.

---

## Primary Workflow Agents

### 🏗️ spec-builder (시스템 아키텍트)

**전문 영역**: SPEC 작성, EARS 요구사항 명세

**책임**:

- EARS 방식으로 요구사항 작성
- SPEC 메타데이터 관리 (id, version, status 등)
- Git 브랜치 생성 (feature/SPEC-XXX)
- Draft PR 생성 (Team 모드)

**실행**:

```bash
/alfred:1-spec "JWT 인증 시스템"
```

**워크플로우**:

1. Phase 1: 프로젝트 분석 및 SPEC 후보 제안
2. Phase 2: SPEC 문서 작성 및 Git 작업

**출력물**:

- `.moai/specs/SPEC-{ID}/spec.md`
- Git 브랜치: `feature/SPEC-{ID}`
- Draft PR (Team 모드)

**관련 문서**: [Stage 1: SPEC Writing](guides/workflow/1-spec.md)

---

### 💎 code-builder (수석 개발자)

**전문 영역**: TDD 구현, 코드 품질

**책임**:

- RED-GREEN-REFACTOR 사이클 실행
- 언어별 최적 TDD 패턴 적용
- TRUST 원칙 준수
- TDD 이력 주석 추가

**실행**:

```bash
/alfred:2-build AUTH-001
```

**워크플로우**:

1. Phase 1: SPEC 분석 및 TDD 계획 수립
2. Phase 2: RED (테스트) → GREEN (구현) → REFACTOR (개선)

**출력물**:

- `tests/` 디렉토리: `@TEST:ID` 태그가 포함된 테스트
- `src/` 디렉토리: `@CODE:ID` 태그가 포함된 구현
- Git 커밋: RED, GREEN, REFACTOR 단계별 커밋

**언어별 지원**:

- **TypeScript**: Vitest, Biome/ESLint
- **Python**: pytest, ruff/black
- **Java**: JUnit, Maven/Gradle
- **Go**: go test, golangci-lint
- **Rust**: cargo test, clippy

**관련 문서**: [Stage 2: TDD Implementation](guides/workflow/2-build.md)

---

### 📖 doc-syncer (테크니컬 라이터)

**전문 영역**: 문서 동기화, Living Document

**책임**:

- TAG 체인 스캔 및 검증
- Living Document 자동 생성
- PR 상태 Draft → Ready 전환
- CI/CD 확인 및 자동 머지 (Team 모드)

**실행**:

```bash
# 기본 동기화
/alfred:3-sync

# 자동 머지 (Team 모드)
/alfred:3-sync --auto-merge

# 검증만 수행
/alfred:3-sync --check
```

**워크플로우**:

1. Phase 1: TAG 체인 스캔 및 TRUST 검증
2. Phase 2: Living Document 생성 및 PR 처리

**출력물**:

- `.moai/reports/sync-report-YYYY-MM-DD.md`
- `docs/features/` 디렉토리: Feature Document (선택)
- PR 상태 업데이트 (Team 모드)

**관련 문서**: [Stage 3: Document Sync](guides/workflow/3-sync.md)

---

## Support Agents

### 🏷️ tag-agent (지식 관리자)

**전문 영역**: TAG 시스템, 추적성 관리

**책임**:

- TAG 목록 조회 및 검색
- 고아 TAG 탐지
- 끊어진 링크 감지
- TAG 관계 시각화

**호출**:

```bash
# TAG 목록 조회
@agent-tag-agent "AUTH 도메인 TAG 목록 조회"

# 고아 TAG 탐지
@agent-tag-agent "고아 TAG 및 끊어진 링크 감지"

# 특정 TAG 추적
@agent-tag-agent "AUTH-001 TAG 체인 추적"
```

**기능**:

- TAG 검색: `rg '@SPEC:AUTH' -n`
- TAG 체인 검증: `@SPEC → @TEST → @CODE → @DOC`
- 고아 TAG 탐지: SPEC 없는 CODE/TEST 발견

**관련 문서**: [TAG System](guides/concepts/tag-system.md)

---

### 🚀 git-manager (릴리스 엔지니어)

**전문 영역**: Git 워크플로우, 배포

**책임**:

- Git 브랜치 생성/삭제/머지
- 커밋 메시지 생성 (Locale 기반)
- 체크포인트 생성 (백업)
- 특정 커밋으로 롤백

**호출**:

```bash
# 체크포인트 생성
@agent-git-manager "체크포인트 생성"

# 특정 커밋으로 롤백
@agent-git-manager "abc1234 커밋으로 롤백"

# 브랜치 정리
@agent-git-manager "머지된 브랜치 정리"
```

**Git 브랜치 정책**:

- 모든 브랜치 생성/머지는 **사용자 확인 필수**
- Personal 모드: 로컬 머지
- Team 모드: PR 기반 머지

**Locale 기반 커밋 메시지**:

- `ko` (한국어): 🔴 RED: 테스트 작성
- `en` (English): 🔴 RED: Test written
- `ja` (日本語): 🔴 RED: テスト作成
- `zh` (中文): 🔴 RED: 测试编写

---

### 🔬 debug-helper (트러블슈팅 전문가)

**전문 영역**: 오류 진단, 해결

**책임**:

- 에러 메시지 분석
- 원인 진단 및 해결 방법 제시
- TAG 체인 무결성 검증
- TRUST 원칙 위반 사항 검출

**호출**:

```bash
# 에러 진단
@agent-debug-helper "TypeError: 'NoneType' object has no attribute 'name'"

# TAG 체인 검증
@agent-debug-helper "TAG 체인 검증을 수행해주세요"

# TRUST 원칙 확인
@agent-debug-helper "TRUST 원칙 준수 여부 확인"
```

**자동 호출 시나리오**:

- TRUST 검증 실패 시
- TAG 체인 끊김 발견 시
- 테스트 실패 시
- 예외 발생 시

**에러 메시지 표준**:

```
[심각도] [컨텍스트]: [문제 설명]
  → [권장 조치]
```

**심각도 아이콘**:

- **❌ Critical**: 작업 중단, 즉시 조치 필요
- **⚠️ Warning**: 주의 필요, 계속 진행 가능
- **ℹ️ Info**: 정보성 메시지, 참고용

**예시**:

```
❌ SPEC 문서 작성 실패: .moai/specs/ 디렉토리 권한 거부
  → chmod 755 .moai/specs 실행 후 재시도

⚠️ 테스트 커버리지 부족: 현재 78% (목표 85%)
  → 추가 테스트 케이스 작성 권장
```

---

### ✅ trust-checker (품질 보증 리드)

**전문 영역**: TRUST 검증, 성능/보안

**책임**:

- TRUST 5원칙 자동 검증
- 코드 품질 메트릭 수집
- 보안 취약점 스캔
- 성능 병목 지점 탐지

**호출**:

```bash
# TRUST 검증
@agent-trust-checker "SPEC-001 TRUST 검증"

# 특정 원칙 검증
@agent-trust-checker "AUTH-001 보안 검증"

# 전체 프로젝트 검증
@agent-trust-checker "전체 TRUST 점수 확인"
```

**TRUST 5원칙 검증**:

1. **T - Test First**
   - 테스트 커버리지 ≥85%
   - 모든 테스트 통과

2. **R - Readable**
   - 린터 통과 (0 issues)
   - 파일 크기 ≤300 LOC
   - 함수 크기 ≤50 LOC
   - 복잡도 ≤10

3. **U - Unified**
   - 타입 체크 통과
   - 일관된 아키텍처 패턴

4. **S - Secured**
   - 보안 스캔 통과 (0 vulnerabilities)
   - 입력 검증 구현

5. **T - Trackable**
   - TAG 체인 무결성 확인
   - 고아 TAG 없음

**관련 문서**: [TRUST Principles](guides/concepts/trust-principles.md)

---

## System Agents

### 🛠️ cc-manager (데브옵스 엔지니어)

**전문 영역**: Claude Code 설정 관리

**책임**:

- `.claude/` 디렉토리 구조 관리
- 커맨드 파일 업데이트
- 에이전트 설정 동기화
- 출력 스타일 커스터마이징

**호출**:

```bash
# 설정 확인
@agent-cc-manager "Claude Code 설정 확인"

# 업데이트
@agent-cc-manager "템플릿 최신화"

# 복원
@agent-cc-manager "설정 복원"
```

**관리 대상**:

- `.claude/commands/`: Alfred 커맨드
- `.claude/agents/`: 전문 에이전트
- `.claude/hooks/`: Git 훅
- `.claude/output-styles/`: 출력 스타일

---

### 📋 project-manager (프로젝트 매니저)

**전문 영역**: 프로젝트 초기화

**책임**:

- 프로젝트 구조 생성
- `product.md`, `structure.md`, `tech.md` 작성
- 언어별 최적화 설정
- Git 초기화

**실행**:

```bash
/alfred:0-project
```

**워크플로우**:

1. 프로젝트 정보 수집 (대화형)
2. `.moai/` 디렉토리 구조 생성
3. 프로젝트 문서 작성
4. 언어별 설정 적용
5. Git 초기화

**생성 파일**:

- `.moai/config.json`
- `.moai/project/product.md`
- `.moai/project/structure.md`
- `.moai/project/tech.md`
- `.moai/memory/development-guide.md`
- `CLAUDE.md`

---

## Quality Gates

### Automatic Verification

Alfred가 각 단계 완료 후 자동으로 검증:

**After 1-spec (SPEC 작성 후)**:

```bash
# SPEC 메타데이터 검증
rg "^id:|^version:|^status:" .moai/specs/SPEC-{ID}/spec.md

# YAML Front Matter 완전성 확인
# - id, version, status (필수)
# - created, updated, author, priority (필수)
```

**After 2-build (TDD 구현 후)**:

```bash
# TRUST 원칙 자동 검증
bun test --coverage  # T - Test
biome check src/  # R - Readable
tsc --noEmit  # U - Unified
npm audit  # S - Secured
rg '@(SPEC|TEST|CODE):' -n  # T - Trackable
```

**After 3-sync (문서 동기화 후)**:

```bash
# TAG 체인 무결성 검증
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# 고아 TAG 탐지
# 끊어진 TAG 체인 탐지
# 중복 TAG 탐지
```

### Exception Handling

**검증 실패 시 자동 조치**:

1. **debug-helper 자동 호출**
   - 문제 진단
   - 해결 방법 제시

2. **사용자에게 보고**
   - 문제 상세 설명
   - 권장 조치 안내

3. **작업 중단**
   - 다음 단계 진행 금지
   - 문제 해결 후 재시도 권장

**예시**:

```markdown
❌ TRUST 검증 실패

### T - Test First
- ❌ 테스트 커버리지: 72% (목표 85% 미만)

### R - Readable
- ❌ 린터 오류: 5개

**권장 조치**:
1. 누락된 테스트 케이스 추가
2. biome check src/ --apply 실행
3. /alfred:2-build 재실행

debug-helper가 자동으로 호출되었습니다.
```

---

## Best Practices

### 1. Alfred 중심 작업 흐름

✅ **권장사항**:

```bash
# Alfred를 통한 작업 요청
/alfred:1-spec "새 기능"
/alfred:2-build SPEC-ID
/alfred:3-sync
```

❌ **피해야 할 것**:

```bash
# 에이전트 직접 호출 (Alfred 우회)
@agent-code-builder "코드 작성"  # 금지
```

### 2. 적절한 에이전트 선택

✅ **권장사항**:

```bash
# TAG 관련 작업 → tag-agent
@agent-tag-agent "AUTH 도메인 TAG 조회"

# 에러 진단 → debug-helper
@agent-debug-helper "TypeError 분석"

# Git 작업 → git-manager
@agent-git-manager "체크포인트 생성"
```

### 3. 품질 게이트 존중

✅ **권장사항**:

- TRUST 검증 실패 시 즉시 수정
- TAG 체인 끊김 발견 시 보완
- debug-helper의 권장 사항 따르기

❌ **피해야 할 것**:

- 검증 실패를 무시하고 다음 단계 진행
- 경고 메시지 무시

---

## Troubleshooting

### Issue 1: 에이전트 응답 없음

**증상**:

```bash
$ /alfred:1-spec "새 기능"
# 응답 없음
```

**해결**:

1. Alfred가 활성화되어 있는지 확인
2. 커맨드 문법 확인
3. 프로젝트 초기화 여부 확인 (`moai init .`)

### Issue 2: 품질 게이트 실패

**증상**:

```bash
❌ TRUST 검증 실패
- Test: 커버리지 72%
```

**해결**:

1. debug-helper의 권장 조치 확인
2. 문제 수정 (테스트 추가 등)
3. 해당 단계 재실행

### Issue 3: TAG 체인 끊김

**증상**:

```bash
⚠️ 불완전한 TAG 체인
- SPEC-UPLOAD-003: SPEC → CODE (TEST 누락)
```

**해결**:

1. tag-agent로 TAG 체인 확인
2. 누락된 TAG 추가 (TEST 작성)
3. doc-syncer로 재검증

---

## 관련 문서

### 워크플로우 가이드

- **[Stage 1: SPEC Writing](guides/workflow/1-spec.md)** - spec-builder 상세 워크플로우
- **[Stage 2: TDD Implementation](guides/workflow/2-build.md)** - code-builder 상세 워크플로우
- **[Stage 3: Document Sync](guides/workflow/3-sync.md)** - doc-syncer 상세 워크플로우
- **[Stage 0: Project Setup](guides/workflow/0-project.md)** - project-manager 상세 워크플로우
- **[Stage 9: Update & Upgrade](guides/workflow/9-update.md)** - 패키지 업데이트 가이드

### 핵심 개념

- **[SPEC-First TDD](guides/concepts/spec-first-tdd.md)** - 개발 방법론
- **[TRUST Principles](guides/concepts/trust-principles.md)** - 품질 원칙
- **[TAG System](guides/concepts/tag-system.md)** - 추적성 시스템
- **[Hooks System](guides/hooks/overview.md)** - Hook 시스템 가이드

---

<div style="text-align: center; margin-top: 40px;">
  <p><strong>Alfred와 함께하는 완벽한 개발</strong> 🤖</p>
  <p>9개의 전문 에이전트가 당신의 코드 품질을 책임집니다!</p>
</div>
