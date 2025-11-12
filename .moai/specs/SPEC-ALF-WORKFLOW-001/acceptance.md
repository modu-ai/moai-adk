
> **4단계 워크플로우 로직 검증 시나리오**
>
> Given-When-Then 형식의 상세 테스트 케이스

---

## 인수 기준 개요

### 검증 목표

1. **의도 파악 단계**: AskUserQuestion이 모호한 요청에서 자동 실행되는지 확인
2. **계획 수립 단계**: Plan Agent가 작업을 올바르게 분석하고 TodoWrite 형식으로 출력하는지 확인
3. **작업 실행 단계**: TodoWrite가 실시간으로 작업 진행 상황을 추적하는지 확인
4. **보고/커밋 단계**: 보고서가 선택적으로, 커밋이 필수로 생성되는지 확인

### 검증 범위

- **기능적 검증**: 4단계 워크플로우의 각 단계가 올바르게 실행됨
- **통합 검증**: 에이전트 간 협업이 원활함
- **사용자 경험 검증**: 사용자가 진행 상황을 명확하게 파악 가능
- **품질 검증**: TAG 체인, 문서 일관성, Git 커밋 구조가 올바름

---

## 테스트 시나리오

### 시나리오 1: 명확한 요청 (AskUserQuestion 건너뛰기)

#### Given (사전 조건)
- MoAI-ADK v0.8.2+ 환경
- Python 3.13+ 프로젝트
- feature/SPEC-ALF-WORKFLOW-001 브랜치
- 사용자가 Claude Code에서 Alfred와 대화 중

#### When (실행)
```
사용자 입력: "JWT 인증 시스템 만들어줘. 30분 토큰 만료, refresh token 지원."
```

#### Then (기대 결과)

**1단계: 의도 파악**
- ✅ Alfred가 요청의 명확성을 HIGH로 평가
- ✅ AskUserQuestion을 건너뛰고 바로 Plan Agent 호출
- ❌ 사용자에게 추가 질문하지 않음

**2단계: 계획 수립**
- ✅ Plan Agent가 호출됨
- ✅ Plan Agent가 작업을 5-6개로 분해
  - SPEC 문서 작성
  - 테스트 코드 작성 (RED)
  - 구현 (GREEN)
  - 리팩토링 (REFACTOR)
  - 문서 동기화
  - Git 커밋 생성
- ✅ TodoWrite에 6개 작업 항목 생성
- ✅ 모든 작업의 초기 상태가 `pending`

**3단계: 작업 실행**
- ✅ 첫 번째 작업 "SPEC 문서 작성"이 `in_progress`로 변경
- ✅ spec-builder 호출
- ✅ `.moai/specs/SPEC-AUTH-001/` 디렉토리 생성
- ✅ 3개 파일 생성 (spec.md, plan.md, acceptance.md)
- ✅ 첫 번째 작업이 `completed`로 변경
- ✅ 두 번째 작업 "테스트 코드 작성"이 `in_progress`로 변경
- ✅ tdd-implementer 호출 (RED 단계)
- ✅ `tests/auth/test_jwt_service.py` 생성
- ✅ pytest 실행 → FAIL (예상됨)
- ✅ 두 번째 작업이 `completed`로 변경
- ✅ (나머지 작업 동일 패턴 반복)

**4단계: 보고/커밋**
- ✅ 모든 작업이 `completed` 상태
- ✅ git-manager 호출
- ✅ 3개 커밋 생성:
  - Commit 1 (RED): "test: add failing tests for JWT authentication"
  - Commit 2 (GREEN): "feat: implement JWT authentication service"
  - Commit 3 (REFACTOR): "refactor: improve JWT service code quality"
- ✅ 모든 커밋에 Alfred co-authorship 포함
- ❌ 보고서 파일 생성 안 됨 (명시적 요청 없음)

**검증 방법**:
```bash
# TodoWrite 작업 수 확인 (6개)
# Git 커밋 수 확인 (3개)

# 보고서 파일 미존재 확인
ls *.md | grep -E "(GUIDE|REPORT|ANALYSIS)"  # 결과: 빈 출력
```

---

### 시나리오 2: 모호한 요청 (AskUserQuestion 실행)

#### Given (사전 조건)
- MoAI-ADK v0.8.2+ 환경
- React + TypeScript 프로젝트
- main 브랜치
- 사용자가 Claude Code에서 Alfred와 대화 중

#### When (실행)
```
사용자 입력: "대시보드 페이지 추가해줘"
```

#### Then (기대 결과)

**1단계: 의도 파악**
- ✅ Alfred가 요청의 명확성을 LOW로 평가
- ✅ AskUserQuestion 실행
- ✅ 4-5개 질문 제시:
  1. "데이터 소스는?" → [REST API | GraphQL | Local state]
  2. "주요 차트 유형은?" → [Time series | Category comparison | Distribution]
  3. "실시간 업데이트 필요?" → [Yes | No | Every 10 seconds]
  4. "접근 제한은?" → [Admin only | Logged-in users | Public]
- ✅ 사용자가 각 질문에 응답 선택
- ✅ Alfred가 선택 사항 요약 표시
- ✅ "이 선택으로 진행할까요?" 최종 확인

**사용자 응답 예시**:
- 데이터 소스: REST API
- 차트 유형: Time series
- 실시간 업데이트: No
- 접근 제한: Logged-in users

**2단계: 계획 수립**
- ✅ Plan Agent 호출 (사용자 응답 포함)
- ✅ TodoWrite에 작업 항목 생성 (5-7개)
- ✅ 작업 항목 예시:
  - "SPEC-DASHBOARD-001 문서 작성"
  - "Dashboard 컴포넌트 테스트 코드 작성"
  - "Dashboard 컴포넌트 구현"
  - "REST API 호출 로직 구현"
  - "Time series 차트 통합"
  - "인증 미들웨어 추가"
  - "문서 동기화"

**3단계: 작업 실행**
- ✅ 각 작업이 순차적으로 `pending → in_progress → completed` 전이
- ✅ 사용자 응답이 SPEC 문서에 반영됨

**4단계: 보고/커밋**
- ✅ git-manager가 TDD 커밋 생성
- ❌ 보고서 파일 생성 안 됨

**검증 방법**:
```bash
# SPEC 문서에 사용자 응답 반영 확인
rg "REST API|Time series|Logged-in users" .moai/specs/SPEC-DASHBOARD-001/spec.md

# Git 커밋 확인
git log --oneline --grep="DASHBOARD-001"
```

---

### 시나리오 3: 차단 요인 발생 및 해결

#### Given (사전 조건)
- MoAI-ADK v0.8.2+ 환경
- Python 3.13 프로젝트
- feature/SPEC-CACHE-001 브랜치
- Redis가 설치되지 않은 환경

#### When (실행)
```
사용자 입력: "FastAPI 앱에 Redis 캐싱 추가해줘"
```

#### Then (기대 결과)

**1단계: 의도 파악**
- ✅ 명확성 HIGH → AskUserQuestion 건너뛰기

**2단계: 계획 수립**
- ✅ Plan Agent 호출
- ✅ TodoWrite에 5개 작업 생성

**3단계: 작업 실행 (차단 발생)**
- ✅ 작업 1 "SPEC 문서 작성" → `completed`
- ✅ 작업 2 "테스트 코드 작성" → `in_progress`
- ✅ tdd-implementer 호출
- ✅ 구현 중 Redis 패키지 미설치 오류 감지
- ✅ tdd-implementer가 차단 요인 보고
- ✅ 작업 2가 `in_progress` 상태 유지 (completed로 변경 안 됨)
- ✅ 새 차단 작업 생성:
  ```
  {
    content: "차단 요인 해결: Redis 패키지 설치 (redis-py)",
    activeForm: "차단 요인 해결 중",
    status: "in_progress"
  }
  ```
- ✅ Alfred가 `pip install redis` 실행
- ✅ 설치 성공
- ✅ 차단 작업이 `completed`로 변경
- ✅ 작업 2 재개
- ✅ 테스트 코드 작성 완료
- ✅ 작업 2가 `completed`로 변경
- ✅ 나머지 작업 계속 진행

**4단계: 보고/커밋**
- ✅ git-manager가 커밋 생성
- ✅ 차단 해결도 커밋 메시지에 포함:
  ```
  chore: install Redis dependencies for caching feature

  - Add redis-py package to requirements.txt
  - Resolve blocker for SPEC-CACHE-001 implementation

  ```

**검증 방법**:
```bash
# 차단 작업이 TodoWrite에 추가되었는지 확인 (로그 분석)
# Redis 패키지 설치 확인
pip show redis

# 차단 해결 커밋 확인
git log --oneline --grep="Redis dependencies"
```

---

### 시나리오 4: 병렬 작업 실행

#### Given (사전 조건)
- MoAI-ADK v0.8.2+ 환경
- Python 프로젝트
- feature/SPEC-DOCS-001 브랜치
- README.md와 CHANGELOG.md 파일 존재

#### When (실행)
```
사용자 입력: "README.md에 새 기능 추가하고, CHANGELOG.md에 v0.8.3 항목 추가해줘"
```

#### Then (기대 결과)

**1단계: 의도 파악**
- ✅ 명확성 HIGH → AskUserQuestion 건너뛰기

**2단계: 계획 수립**
- ✅ Plan Agent 호출
- ✅ Plan Agent가 병렬 실행 가능 판단:
  - 이유: 두 파일이 독립적, 충돌 없음
- ✅ TodoWrite에 2개 작업 생성 (병렬 표시):
  ```
  [
    {
      content: "README.md 업데이트 (새 기능 섹션 추가)",
      activeForm: "README.md 업데이트 중",
      status: "pending",
      parallel_group: "docs_update"
    },
    {
      content: "CHANGELOG.md 업데이트 (v0.8.3 항목 추가)",
      activeForm: "CHANGELOG.md 업데이트 중",
      status: "pending",
      parallel_group: "docs_update"
    }
  ]
  ```

**3단계: 작업 실행 (병렬)**
- ✅ 두 작업이 **동시에** `in_progress`로 변경
- ✅ 병렬 호출:
  - Thread 1: doc-syncer → README.md 업데이트
  - Thread 2: doc-syncer → CHANGELOG.md 업데이트
- ✅ 두 작업이 독립적으로 진행
- ✅ 두 작업이 **거의 동시에** `completed`로 변경 (1-2초 차이)

**4단계: 보고/커밋**
- ✅ git-manager가 단일 커밋 생성 (두 파일 수정 포함):
  ```
  docs: update README and CHANGELOG for v0.8.3

  - Add new feature section to README.md
  - Add v0.8.3 release notes to CHANGELOG.md

  ```

**검증 방법**:
```bash
# 두 파일 수정 확인
git diff HEAD~1 README.md CHANGELOG.md

# 병렬 실행 시간 확인 (로그 타임스탬프 분석)
# 예상: 두 작업이 1-2초 차이로 완료
```

---

### 시나리오 5: 보고서 명시적 요청

#### Given (사전 조건)
- MoAI-ADK v0.8.2+ 환경
- Python 프로젝트
- main 브랜치
- 기존 인증 시스템 코드 존재

#### When (실행)
```
사용자 입력: "기존 인증 시스템 분석하고, 보고서 만들어줘. 개선 방안도 포함해줘."
```

#### Then (기대 결과)

**1단계: 의도 파악**
- ✅ "보고서 만들어줘" 키워드 감지
- ✅ 명확성 MEDIUM (분석 범위는 명확, 세부 사항 일부 모호)
- ✅ AskUserQuestion 실행:
  1. "분석 범위는?" → [인증 로직만 | 인증 + 권한 | 전체 보안 시스템]
  2. "보고서 형식은?" → [기술 분석 | 비즈니스 리포트 | 개발자 가이드]
  3. "개선 방안 우선순위는?" → [보안 강화 | 성능 최적화 | 사용자 경험]

**사용자 응답 예시**:
- 분석 범위: 인증 + 권한
- 보고서 형식: 기술 분석
- 우선순위: 보안 강화

**2단계: 계획 수립**
- ✅ Plan Agent 호출
- ✅ TodoWrite에 작업 생성:
  - "기존 코드 분석 (인증 + 권한 시스템)"
  - "보안 취약점 식별"
  - "개선 방안 도출 (보안 강화 중심)"
  - "**보고서 문서 작성 (.moai/analysis/auth-system-analysis.md)**"

**3단계: 작업 실행**
- ✅ 각 작업 순차 실행
- ✅ 마지막 작업에서 보고서 파일 생성:
  - 위치: `.moai/analysis/auth-system-analysis.md`
  - 내용: 코드 분석, 취약점, 개선 방안, 우선순위
- ✅ 프로젝트 루트에 파일 생성 안 됨 (`.moai/analysis/`에만 생성)

**4단계: 보고/커밋**
- ✅ git-manager가 커밋 생성:
  ```
  docs: add authentication system analysis report

  - Analyze existing auth and authorization system
  - Identify security vulnerabilities
  - Propose security-focused improvements

  Report: .moai/analysis/auth-system-analysis.md
  ```

**검증 방법**:
```bash
# 보고서 파일 존재 확인
ls .moai/analysis/auth-system-analysis.md

# 루트 디렉토리에 보고서 없음 확인
ls *.md | grep -E "(ANALYSIS|REPORT|GUIDE)" | wc -l  # 결과: 0

# Git 커밋 확인
git log --oneline --grep="analysis report"
```

---

### 시나리오 6: 보고서 미요청 (기본 동작)

#### Given (사전 조건)
- MoAI-ADK v0.8.2+ 환경
- TypeScript 프로젝트
- feature/SPEC-API-001 브랜치

#### When (실행)
```
사용자 입력: "REST API 엔드포인트 10개 만들어줘. CRUD 작업 포함."
```

#### Then (기대 결과)

**1단계: 의도 파악**
- ✅ 명확성 HIGH
- ✅ "보고서" 키워드 없음 → 보고서 생성 안 함

**2단계: 계획 수립**
- ✅ Plan Agent 호출
- ✅ TodoWrite에 작업 생성 (SPEC, 테스트, 구현, 문서 동기화)

**3단계: 작업 실행**
- ✅ 모든 작업 순차 실행
- ✅ 10개 API 엔드포인트 구현

**4단계: 보고/커밋**
- ✅ git-manager가 TDD 커밋 생성
- ❌ 보고서 파일 생성 안 됨 (명시적 요청 없음)
- ❌ 프로젝트 루트에 `IMPLEMENTATION_GUIDE.md` 생성 안 됨
- ❌ `.moai/docs/`에도 자동 보고서 생성 안 됨

**검증 방법**:
```bash
# 보고서 파일 미존재 확인 (전체 프로젝트)
find . -name "*GUIDE*.md" -o -name "*REPORT*.md" -o -name "*ANALYSIS*.md" | wc -l
# 결과: 0 (README.md 등 공식 문서 제외)

# Git 커밋만 확인
```

---

## 품질 게이트

### Gate 1: 문서 일관성

**기준**:
- ✅ CLAUDE.md에 4단계 워크플로우 설명 존재
- ✅ CLAUDE-RULES.md에 AskUserQuestion, TodoWrite 규칙 존재
- ✅ 모든 명령 템플릿 (1-plan, 2-run, 3-sync)이 워크플로우 언급
- ✅ 5개 에이전트 모두 TodoWrite 업데이트 로직 포함

**검증 방법**:
```bash
rg "4-step workflow|four-step workflow|Intent Understanding|Plan Creation|Task Execution|Report & Commit" CLAUDE.md CLAUDE-RULES.md .claude/commands/ .claude/agents/
```

### Gate 2: TAG 체인 무결성

**기준**:
- ✅ 각 파일에 고유한 서브 TAG 존재 (예: :ALFRED, :RULES, :CMD-PLAN)
- ✅ 모든 TAG가 spec.md의 Traceability 섹션에 문서화됨

**검증 방법**:
```bash
# TAG 존재 확인
rg "@(SPEC|CODE):ALF-WORKFLOW-001" -n CLAUDE.md CLAUDE-RULES.md .claude/

# TAG 수 확인 (10개 이상)
# 예상 결과: 10
```

### Gate 3: Git 커밋 구조

**기준**:
- ✅ TDD 단계별 커밋 분리 (test, feat, refactor)
- ✅ 모든 커밋에 Alfred co-authorship 포함
- ✅ 커밋 메시지가 Conventional Commits 형식 준수

**검증 방법**:
```bash
# TDD 단계별 커밋 확인

# Alfred co-authorship 확인
```

### Gate 4: TodoWrite 상태 전이

**기준** (수동 검증):
- ✅ 각 작업이 pending → in_progress → completed 순서로 전이
- ✅ 동시에 최대 하나의 작업만 in_progress (병렬 제외)
- ✅ 차단 요인 발생 시 새 작업 생성 및 원래 작업 in_progress 유지
- ✅ 모든 작업 completed 후 커밋 생성

**검증 방법**:
- 통합 테스트 실행 중 TodoWrite 상태 변화 모니터링
- 로그에서 상태 전이 타임스탬프 분석
- 예상 패턴과 실제 패턴 비교

### Gate 5: 보고서 생성 규칙

**기준**:
- ✅ "보고서 만들어줘" 명시 시에만 생성
- ✅ 생성 위치: `.moai/docs/`, `.moai/reports/`, `.moai/analysis/`
- ❌ 프로젝트 루트에 자동 생성 금지
- ❌ 명시적 요청 없이 자동 생성 금지

**검증 방법**:
```bash
# 시나리오 5 (보고서 요청) → 파일 존재 확인
ls .moai/analysis/auth-system-analysis.md

# 시나리오 1, 6 (보고서 미요청) → 파일 미존재 확인
find . -maxdepth 1 -name "*GUIDE*.md" -o -name "*REPORT*.md"
# 예상 결과: 빈 출력
```

---

## 비기능적 요구사항

### 성능 기준

| 단계 | 목표 시간 | 측정 방법 |
|------|-----------|-----------|
| AskUserQuestion 실행 | <2초 | 질문 표시까지 시간 측정 |
| Plan Agent 호출 | <5초 | 호출부터 TodoWrite 생성까지 |
| TodoWrite 상태 업데이트 | <100ms | 상태 변경 응답 시간 |
| Git 커밋 생성 | <3초 | git-manager 호출부터 커밋 완료까지 |

### 사용자 경험 기준

| 항목 | 목표 | 측정 방법 |
|------|------|-----------|
| AskUserQuestion 질문 수 | 3-5개 | 질문 개수 카운트 |
| TodoWrite 진행 상황 가시성 | 실시간 | 사용자가 매 작업 변화를 확인 가능 |
| 오류 메시지 명확성 | 명확함 | 차단 요인 발생 시 원인과 해결 방법 표시 |
| 커밋 메시지 가독성 | 명확함 | Conventional Commits 준수, 맥락 포함 |

---

## 회귀 테스트

### 기존 기능 영향 확인

**테스트 항목**:
1. ✅ `/alfred:0-project` 여전히 작동 (프로젝트 초기화)
2. ✅ `/alfred:1-plan` 기존 SPEC 생성 로직 정상 동작
3. ✅ `/alfred:2-run` 기존 TDD 구현 로직 정상 동작
4. ✅ `/alfred:3-sync` 기존 문서 동기화 로직 정상 동작
5. ✅ 기존 SPEC 문서 (28개) 영향 없음
6. ✅ 기존 Git 워크플로우 (branch, commit, PR) 정상 동작

**검증 방법**:
```bash
# 기존 SPEC 개수 확인 (28개 + 1개 신규)
ls .moai/specs/ -d SPEC-* | wc -l
# 예상 결과: 29

# 기존 명령 실행 테스트
/alfred:0-project
/alfred:1-plan "테스트 기능"
# 예상: 오류 없이 정상 실행
```

---

## Definition of Done

### SPEC 문서 완료 기준
- ✅ spec.md, plan.md, acceptance.md 3개 파일 존재
- ✅ YAML frontmatter 7개 필수 필드 작성
- ✅ HISTORY 섹션에 v0.0.1 INITIAL 항목 존재
- ✅ EARS 요구사항 18개 작성
- ✅ Traceability 섹션에 TAG 매핑 완료

### 구현 완료 기준
- ✅ 10개 파일 모두 수정 완료
- ✅ 6개 시나리오 모두 통과
- ✅ 5개 품질 게이트 모두 통과
- ✅ Git 커밋 TDD 단계별 분리
- ✅ PR #118 Ready for Review 상태
- ✅ 회귀 테스트 통과

### 문서화 완료 기준
- ✅ CLAUDE-PRACTICES.md 업데이트 (4단계 워크플로우 예시 추가)
- ✅ CLAUDE-AGENTS-GUIDE.md 업데이트 (Plan Agent 추가)
- ✅ README.md 업데이트 (Four-Step Workflow 섹션 추가)

### 릴리스 준비 완료 기준
- ✅ main 브랜치에 머지
- ✅ SPEC 문서 버전 v0.1.0으로 업데이트
- ✅ CHANGELOG.md에 v0.8.3 항목 추가
- ✅ Git 태그 생성 (v0.8.3)

---

## 검증 체크리스트

### 사전 준비
- [ ] MoAI-ADK v0.8.2+ 설치 확인
- [ ] Python 3.13+ 환경 확인
- [ ] Claude Code 최신 버전 확인
- [ ] feature/SPEC-ALF-WORKFLOW-001 브랜치 생성 확인

### 시나리오 실행
- [ ] 시나리오 1: 명확한 요청 → 통과
- [ ] 시나리오 2: 모호한 요청 → 통과
- [ ] 시나리오 3: 차단 요인 → 통과
- [ ] 시나리오 4: 병렬 작업 → 통과
- [ ] 시나리오 5: 보고서 요청 → 통과
- [ ] 시나리오 6: 보고서 미요청 → 통과

### 품질 게이트
- [ ] Gate 1: 문서 일관성 → 통과
- [ ] Gate 2: TAG 체인 무결성 → 통과
- [ ] Gate 3: Git 커밋 구조 → 통과
- [ ] Gate 4: TodoWrite 상태 전이 → 통과
- [ ] Gate 5: 보고서 생성 규칙 → 통과

### 회귀 테스트
- [ ] 기존 명령 정상 동작 확인
- [ ] 기존 SPEC 문서 영향 없음 확인
- [ ] 기존 Git 워크플로우 정상 동작 확인

### 최종 확인
- [ ] 모든 커밋에 Alfred co-authorship 포함
- [ ] PR #118 Ready for Review 상태
- [ ] 문서 업데이트 완료 (CLAUDE-PRACTICES.md, CLAUDE-AGENTS-GUIDE.md, README.md)

---

**마지막 업데이트**: 2025-10-29
**문서 버전**: v0.0.1
