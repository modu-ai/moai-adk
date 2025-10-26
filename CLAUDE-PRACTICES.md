# CLAUDE-PRACTICES.md

> MoAI-ADK Practical Workflows & Examples

---

## Alfred를 위해: 이 문서가 필요한 이유

Alfred가 이 문서를 읽는 시점:
1. 실제 작업을 수행할 때 - "구체적으로 어떻게 실행할 것인가?"
2. Context 관리가 필요할 때 - "Explore를 어떻게 효율적으로 사용할 것인가?"
3. 문제를 해결할 때 - "이 에러/문제를 어떻게 진단하고 해결할 것인가?"
4. 새 개발자가 온보딩할 때 - "MoAI-ADK 워크플로우를 실전으로 배우기"

Alfred의 의사결정:
- "이 작업을 수행하기 위한 구체적인 단계는?"
- "필요한 컨텍스트를 어떻게 최소한으로 수집할 것인가?"
- "문제 발생 시 어디서 문제를 진단할 것인가?"

이 문서를 읽으면:
- JIT (Just-in-Time) 컨텍스트 관리 전략 습득
- Explore agent를 효율적으로 사용하는 방법 학습
- SPEC → TDD → Sync의 구체적 실행 명령어 숙달
- 자주 발생하는 문제별 해결 방법 참조

---
→ 관련 문서:
- [규칙 확인은 CLAUDE-RULES.md](./CLAUDE-RULES.md#skill-invocation-rules)를 참조하세요
- [Agent 선택은 CLAUDE-AGENTS-GUIDE.md](./CLAUDE-AGENTS-GUIDE.md#agent-선택-결정-트리)를 참조하세요

---

## Context Engineering Strategy

### 1. JIT (Just-in-Time) Retrieval

- Pull only the context required for the immediate step.
- Prefer `Explore` over manual file hunting.
- Cache critical insights in the task thread for reuse.

#### Efficient Use of Explore

- Request call graphs or dependency maps when changing core modules.
- Fetch examples from similar features before implementing new ones.
- Ask for SPEC references or TAG metadata to anchor changes.

### 2. Layered Context Summaries

1. **High-level brief**: purpose, stakeholders, success criteria.
2. **Technical core**: entry points, domain models, shared utilities.
3. **Edge cases**: known bugs, performance constraints, SLAs.

### 3. Living Documentation Sync

- Align code, tests, and docs after each significant change.
- Use `/alfred:3-sync` to update Living Docs and TAG references.
- Record rationale for deviations from the SPEC.

---

## On-demand Agent Usage

### Debugging & Analysis

- Use `debug-helper` for error triage and hypothesis testing.
- Attach logs, stack traces, and reproduction steps.
- Ask for fix-forward vs rollback recommendations.

### TAG System Management

- Assign IDs as `<DOMAIN>-<###>` (e.g., `AUTH-003`).
- Update HISTORY with every change.
- Cross-check usage with `rg '@TAG:ID' -n` searches.

### Backup Management

- `/alfred:0-project` and `git-manager` create automatic safety snapshots (e.g., `.moai-backups/`) before risky actions.
- Manual `/alfred:9-checkpoint` commands have been deprecated; rely on Git branches or team-approved backup workflows when additional restore points are needed.

---

## 실전 워크플로우 예제

### 시나리오 1: 새 기능 구현 (USER-DASHBOARD-001)

**상황**: 사용자가 "사용자 대시보드 추가"를 요청함

**Alfred의 실행 순서**:

1. **모호함 감지 → AskUserQuestion 실행**
   ```
   질문 1: 데이터 소스는 무엇인가요?
   선택지: [REST API | GraphQL | Local State]

   질문 2: 갱신 주기는 어떻게 되나요?
   선택지: [실시간 (WebSocket) | 10초마다 | 수동 새로고침]

   질문 3: 접근 제한이 필요한가요?
   선택지: [Admin만 | 로그인한 사용자 | Public]

   사용자 답변:
   - 데이터 소스: REST API
   - 갱신 주기: 수동 새로고침
   - 접근 제한: 로그인한 사용자
   ```

2. **SPEC 작성 (사용자 답변 기반)**
   ```bash
   /alfred:1-plan "User Dashboard Feature - Display user stats with manual refresh, authenticated access only"
   ```

   **산출물**: `.moai/specs/SPEC-USER-DASHBOARD-001/spec.md`
   - YAML metadata: id, version: 0.0.1, status: draft
   - @SPEC:USER-DASHBOARD-001 TAG
   - EARS 문법 요구사항:
     - "The system must display user statistics dashboard"
     - "WHEN user clicks refresh button, THEN fetch latest data from REST API"
     - "IF user not authenticated, THEN redirect to login page"

3. **TDD 구현 (RED → GREEN → REFACTOR)**
   ```bash
   /alfred:2-run USER-DASHBOARD-001
   ```

   **Alfred 내부 실행**:
   - **implementation-planner** (Phase 1):
     - 구현 전략 수립: React component + fetch API + auth guard
     - 라이브러리 선택: react-query (data fetching), @tanstack/react-query (caching)
     - TAG 설계: @CODE:USER-DASHBOARD-001:UI, @CODE:USER-DASHBOARD-001:API

   - **tdd-implementer** (Phase 2):
     - **RED**: `tests/features/dashboard.test.tsx` 작성 (실패하는 테스트)
     - **GREEN**: `src/features/Dashboard.tsx` 구현 (테스트 통과)
     - **REFACTOR**: 코드 정리, hook 분리, 재사용성 향상

4. **문서 동기화**
   ```bash
   /alfred:3-sync
   ```

   **Alfred 내부 실행**:
   - TAG 체인 검증: @SPEC ↔ @TEST ↔ @CODE
   - Living Document 업데이트: README.md, CHANGELOG.md
   - PR 상태 변경: Draft → Ready

**최종 산출물**:
- SPEC: `.moai/specs/SPEC-USER-DASHBOARD-001/spec.md` (@SPEC:USER-DASHBOARD-001)
- TEST: `tests/features/dashboard.test.tsx` (@TEST:USER-DASHBOARD-001)
- CODE: `src/features/Dashboard.tsx` (@CODE:USER-DASHBOARD-001:UI)
- CODE: `src/api/dashboard.ts` (@CODE:USER-DASHBOARD-001:API)
- DOCS: `docs/features/USER-DASHBOARD-001.md` (@DOC:USER-DASHBOARD-001)

**예상 소요 시간**: 30-45분 (SPEC 10분 + TDD 20분 + Sync 10분)

---

### 시나리오 2: 버그 수정 (BUG-AUTHENTICATION-TIMEOUT)

**상황**: 사용자가 "5분 후 인증이 자동으로 끊김" 버그를 보고

**Alfred의 실행 순서**:

1. **에러 분석 (debug-helper)**
   ```bash
   @agent-debug-helper "Authentication timeout after 5 minutes - expected 30 minutes"
   ```

   **debug-helper 분석 결과**:
   - 어느 함수에서 timeout 발생? → `src/auth/token.ts:validateToken()`
   - 현재 timeout 값은 얼마? → `300000 ms` (5분)
   - 정상 값은 얼마? → `1800000 ms` (30분)
   - 원인: JWT 토큰 만료 시간이 잘못 설정됨

2. **SPEC 작성 (버그 수정용)**
   ```bash
   /alfred:1-plan "Fix AUTH-TIMEOUT-001: JWT token expiration should be 30 minutes, not 5 minutes"
   ```

   **산출물**: `.moai/specs/SPEC-AUTH-TIMEOUT-001/spec.md`
   - Bug description: JWT expiration 5분 → 30분으로 수정
   - Root cause: `expiresIn` 값 오류 (`300` → `1800`으로 수정)
   - Test case: Token validity를 30분 동안 검증

3. **TDD 구현 (RED → GREEN → REFACTOR)**
   ```bash
   /alfred:2-run AUTH-TIMEOUT-001
   ```

   **Alfred 내부 실행**:
   - **RED**: `tests/auth/token.test.ts` 추가
     ```typescript
     it('should keep token valid for 30 minutes', () => {
       const token = generateToken();
       const now = Date.now();
       const futureTime = now + 30 * 60 * 1000;
       expect(isTokenValid(token, futureTime)).toBe(true);
     });
     ```

   - **GREEN**: `src/auth/token.ts` 수정
     ```typescript
     const JWT_EXPIRATION = 1800; // 30 minutes (was 300)
     ```

   - **REFACTOR**: 상수화
     ```typescript
     const JWT_EXPIRATION_MINUTES = 30;
     const JWT_EXPIRATION = JWT_EXPIRATION_MINUTES * 60;
     ```

4. **검증**
   - **TRUST 5 확인**:
     - Test First: ✅ 새 테스트 케이스 추가
     - Readable: ✅ ruff lint 통과
     - Unified: ✅ mypy type safety 통과
     - Secured: ✅ trivy 보안 스캔 통과
     - Trackable: ✅ @TAG 체인 정상

   - **TAG 체인 검증**:
     ```bash
     rg '@(SPEC|TEST|CODE):AUTH-TIMEOUT-001' -n
     ```
     - @SPEC:AUTH-TIMEOUT-001 → `.moai/specs/SPEC-AUTH-TIMEOUT-001/spec.md`
     - @TEST:AUTH-TIMEOUT-001 → `tests/auth/token.test.ts`
     - @CODE:AUTH-TIMEOUT-001 → `src/auth/token.ts`

**최종 산출물**:
- SPEC 업데이트
- TEST 추가
- CODE 수정 (1줄)
- Git commit: `fix(auth): Extend JWT expiration to 30 minutes (was 5 minutes) - Refs: @AUTH-TIMEOUT-001`

**예상 소요 시간**: 15-20분 (분석 5분 + SPEC 5분 + TDD 5분 + 검증 5분)

---

### 시나리오 3: 문서 동기화 (자동)

**상황**: 코드 수정 후 문서를 최신 상태로 유지

**Alfred의 실행 순서**:

1. **변경된 파일 확인**
   ```bash
   git diff develop...HEAD
   ```

   **결과**:
   - `src/features/Dashboard.tsx` (modified)
   - `src/api/dashboard.ts` (new)
   - `tests/features/dashboard.test.tsx` (new)

2. **Living Document 검증**
   ```bash
   /alfred:3-sync status
   ```

   **doc-syncer 분석**:
   - README.md 업데이트 필요: Features 섹션에 "User Dashboard" 추가
   - CHANGELOG.md 생성 필요: v0.4.2 릴리즈 노트
   - TAG 무결성 확인: 모든 @CODE에 @SPEC 연결됨

3. **TAG 무결성 확인**
   ```bash
   rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
   ```

   **결과**:
   - ✅ @SPEC:USER-DASHBOARD-001 → @TEST:USER-DASHBOARD-001 ✅
   - ✅ @TEST:USER-DASHBOARD-001 → @CODE:USER-DASHBOARD-001:UI ✅
   - ✅ @CODE:USER-DASHBOARD-001:UI → @DOC:USER-DASHBOARD-001 ✅
   - 🎉 No orphan TAGs detected

4. **PR 상태 변경 (Draft → Ready)**
   ```bash
   @agent-git-manager "Move PR #42 from Draft to Ready"
   ```

   **git-manager 실행**:
   - PR 검증: 모든 tests 통과, coverage ≥85%
   - PR label 업데이트: `draft` → `ready-for-review`
   - Reviewer 자동 할당: GOOS오라버니
   - PR 설명 업데이트: CHANGELOG.md 내용 반영

**최종 산출물**:
- README.md 자동 업데이트 (Features 섹션)
- CHANGELOG.md 자동 생성 (v0.4.2 entry)
- TAG 체인 검증 완료
- PR #42 상태: Draft → Ready for Review

**예상 소요 시간**: 5-10분 (자동화)

---

**마지막 업데이트**: 2025-10-26
**문서 버전**: v1.0.0 (Option A Refactoring)
