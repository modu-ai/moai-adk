# Implementation Plan: @SPEC:UPDATE-REFACTOR-001

## Overview

**목표**: `/alfred:9-update` Option C 하이브리드 리팩토링 - 문서-구현 불일치 해소 및 Alfred 중앙 오케스트레이션 복원

**우선순위**: P0 (Critical)

**예상 영향 범위**:
- 파일 제거: 1개 (`template-copier.ts`)
- 파일 수정: 2개 (`update-orchestrator.ts`, `9-update.md`)
- 테스트 추가: 3개 (통합 테스트)

---

## Milestones

### 1차 목표: 코드 정리 및 Orchestrator 수정

**범위**:
- `template-copier.ts` 제거
- `update-orchestrator.ts` Phase 4 로직 제거
- TypeScript 컴파일 오류 해결

**완료 조건**:
- TypeScript 빌드 성공
- 기존 테스트 통과 (Phase 4 제외)

---

### 2차 목표: 9-update.md 문서 업데이트

**범위**:
- Phase 4 Section 전면 재작성 (A-I 카테고리 상세화)
- Phase 5 검증 로직 강화
- Phase 6 품질 검증 옵션 추가
- 오류 복구 시나리오 추가

**완료 조건**:
- 모든 P0, P1 요구사항이 문서에 반영됨
- Alfred 실행 방식이 정확히 명세됨

---

### 3차 목표: 통합 테스트 및 검증

**범위**:
- 로컬 테스트 환경 구성
- `/alfred:9-update` 실행 테스트 (정상 시나리오)
- 오류 시나리오 테스트 (파일 누락, 권한 오류 등)
- --check-quality 옵션 테스트

**완료 조건**:
- 모든 AC (Acceptance Criteria) 통과
- 오류 복구 전략이 정상 작동

---

### 최종 목표: 배포 및 문서 동기화

**범위**:
- CHANGELOG.md 업데이트
- moai-adk 패키지 배포 (v0.0.3 or v0.1.0)
- Living Document 생성 (`/alfred:3-sync`)

**완료 조건**:
- 패키지 배포 완료
- 사용자 가이드 업데이트
- @TAG 체인 무결성 검증

---

## Technical Approach

### 아키텍처 변경 전략

#### Before (현재)
```
Alfred (CLAUDE.md)
  → [Bash] UpdateOrchestrator.executeUpdate()
      ├─ Phase 1: VersionChecker
      ├─ Phase 2: BackupManager
      ├─ Phase 3: NpmUpdater
      ├─ Phase 4: TemplateCopier (Node.js fs) ❌
      └─ Phase 5: UpdateVerifier
```

#### After (리팩토링)
```
Alfred (CLAUDE.md)
  ├─ Phase 1-3: [Bash] UpdateOrchestrator.executeUpdate()
  │   ├─ VersionChecker
  │   ├─ BackupManager
  │   └─ NpmUpdater
  │
  ├─ Phase 4: Alfred 직접 실행 (Claude Code 도구) ✅
  │   ├─ [Bash] npm root
  │   ├─ [Glob] 템플릿 파일 검색
  │   ├─ [Read] 파일 읽기
  │   ├─ [Grep] {{PROJECT_NAME}} 패턴 검색
  │   ├─ [Write] 파일 복사
  │   └─ [Bash] chmod +x 권한 부여
  │
  └─ Phase 5: [Bash] UpdateVerifier.verifyUpdate()
      ├─ [Glob] 파일 개수 검증
      ├─ [Read] YAML frontmatter 파싱
      └─ [Grep] 버전 정보 확인
```

---

### 제거 대상: template-copier.ts

**현재 역할**:
- `copyTemplates(templatePath)`: Node.js fs로 파일 복사
- `copyDirectory()`: 재귀적 디렉토리 복사
- `countFiles()`: 파일 개수 카운트

**제거 이유**:
1. MoAI-ADK 철학 위배 (Claude Code 도구 우선 원칙)
2. Alfred의 중앙 오케스트레이션 역할 약화
3. 프로젝트 문서 보호 로직 부재 (Grep 검증 없음)
4. Output Styles 복사 누락

**영향 분석**:
- `update-orchestrator.ts`에서 import 제거
- Phase 4 로직 전체 삭제 (라인 121-123)
- 테스트 파일 수정 필요 (`update-orchestrator.test.ts`)

---

### 수정 대상: update-orchestrator.ts

#### 변경 내용

**삭제 (Phase 4 로직)**:
```typescript
// Phase 4: Template file copy
const npmRoot = await this.npmUpdater.getNpmRoot();
const templatePath = path.join(npmRoot, 'moai-adk', 'templates');
const filesUpdated = await this.templateCopier.copyTemplates(templatePath);
```

**추가 (Alfred 호출 주석)**:
```typescript
// Phase 4: Template file copy (Alfred가 직접 실행)
// → /alfred:9-update.md Phase 4 참조
// → Claude Code 도구: [Glob], [Read], [Grep], [Write], [Bash]
logger.log(chalk.cyan('\n📄 Phase 4는 Alfred가 Claude Code 도구로 직접 실행합니다...'));
logger.log(chalk.gray('   (템플릿 파일 복사는 9-update.md 명령어 참조)'));
```

**수정 (반환값)**:
```typescript
// Before
return {
  success: true,
  filesUpdated,  // ❌ template-copier에서 반환
  ...
};

// After
return {
  success: true,
  filesUpdated: 0,  // ✅ Alfred가 별도로 카운트
  ...
};
```

**라인 수 변화**:
- 현재: 168 LOC
- 리팩토링 후: ~120 LOC (Phase 4 로직 삭제)

---

### 업데이트 대상: 9-update.md

#### Phase 4 Section 재작성

**현재 문제점**:
- Claude Code 도구 명령이 너무 추상적
- Grep 검증 로직 누락
- Output Styles 복사 누락
- 오류 복구 전략 부족

**개선 내용**:

1. **카테고리별 복사 절차 상세화** (A-I):
   - A: 명령어 파일 (.claude/commands/alfred/)
   - B: 에이전트 파일 (.claude/agents/alfred/)
   - C: 훅 파일 + 권한 (.claude/hooks/alfred/)
   - D: Output Styles (**신규 추가**)
   - E: 개발 가이드 (.moai/memory/)
   - F-H: 프로젝트 문서 (Grep 검증)
   - I: CLAUDE.md (Grep 검증)

2. **Grep을 통한 프로젝트 문서 보호**:
   ```text
   [Grep] "{{PROJECT_NAME}}" -n .moai/project/product.md
   → 결과 있음: 템플릿 상태 → 덮어쓰기
   → 결과 없음: 사용자 수정 → 백업 후 덮어쓰기
   ```

3. **chmod +x 권한 부여**:
   ```bash
   [Bash] chmod +x .claude/hooks/alfred/*.cjs
   ```

4. **오류 처리 강화**:
   - Write 실패 → mkdir -p 후 재시도
   - Grep 실패 → 무조건 백업 모드
   - chmod 실패 → 경고만 출력

---

#### Phase 5 검증 로직 강화

**추가 검증 항목**:

1. **파일 개수 검증** (동적 계산):
   ```text
   [Glob] .claude/commands/alfred/*.md → 실제 개수
   [Glob] {npm_root}/moai-adk/templates/.claude/commands/alfred/*.md → 예상 개수
   → 비교 후 불일치 시 경고
   ```

2. **YAML Frontmatter 검증**:
   ```text
   [Read] .claude/commands/alfred/1-spec.md (첫 10줄)
   → YAML 파싱 시도
   → 성공: ✅ / 실패: ❌ (손상 감지)
   ```

3. **버전 정보 확인**:
   ```bash
   [Grep] "version:" -n .moai/memory/development-guide.md
   [Bash] npm list moai-adk --depth=0
   → 버전 일치 확인
   ```

---

#### Phase 6 품질 검증 옵션 추가

**실행 조건**: `--check-quality` 플래그 제공 시

**실행 흐름**:
```text
Phase 6: 품질 검증
  → [Alfred] @agent-trust-checker "Level 1 빠른 스캔"
  → 결과: Pass / Warning / Critical
  → 결과별 조치:
     - Pass: 업데이트 완료
     - Warning: 경고 표시 후 완료
     - Critical: 롤백 제안
```

**검증 항목**:
- 파일 무결성 (YAML 유효성)
- 설정 일관성 (config.json ↔ development-guide.md)
- TAG 체계 (문서 내 @TAG 형식)
- EARS 구문 (SPEC 템플릿)

---

#### 오류 복구 시나리오 추가

**시나리오 1: 파일 복사 실패**
- 원인: 디스크 공간 부족, 권한 오류
- 조치: 실패 파일 목록 수집 → 재시도 제안

**시나리오 2: 검증 실패 (파일 누락)**
- 원인: 템플릿 손상, 네트워크 오류
- 조치: Phase 4 재실행 제안

**시나리오 3: 버전 불일치**
- 원인: npm 캐시 손상
- 조치: Phase 3 재실행 제안 (npm 재설치)

---

## Testing Strategy

### Unit Tests (단위 테스트)

**대상**: 없음 (template-copier.ts 제거로 단위 테스트 불필요)

**이유**: Phase 4는 Alfred가 Claude Code 도구로 직접 실행 (코드가 아닌 명령어 기반)

---

### Integration Tests (통합 테스트)

#### Test 1: 정상 시나리오 - 템플릿 상태 파일

**Given**:
- 프로젝트 문서에 {{PROJECT_NAME}} 패턴이 존재 (템플릿 상태)
- moai-adk@latest가 설치되어 있음

**When**:
- `/alfred:9-update` 실행

**Then**:
- Phase 1-3 정상 완료 (Orchestrator)
- Phase 4 Alfred 직접 실행:
  - 명령어 파일 10개 복사 ✅
  - 에이전트 파일 9개 복사 ✅
  - 훅 파일 4개 복사 + chmod +x ✅
  - Output Styles 4개 복사 ✅
  - 프로젝트 문서 3개 덮어쓰기 (백업 없음) ✅
  - CLAUDE.md 덮어쓰기 (백업 없음) ✅
- Phase 5 검증 통과 ✅

---

#### Test 2: 사용자 수정 파일 보호

**Given**:
- 프로젝트 문서에 {{PROJECT_NAME}} 패턴이 없음 (사용자 수정 상태)

**When**:
- `/alfred:9-update` 실행

**Then**:
- Phase 4에서 Grep 검증 수행:
  - [Grep] "{{PROJECT_NAME}}" → 결과 없음
  - 백업 생성: `.moai-backup/{timestamp}/.moai/project/product.md` ✅
  - 새 템플릿 복사 ✅
- 완료 메시지: "✅ .moai/project/product.md (백업: yes)"

---

#### Test 3: Output Styles 복사 확인

**Given**:
- .claude/output-styles/alfred/ 디렉토리가 없음

**When**:
- `/alfred:9-update` 실행

**Then**:
- Phase 4에서 Output Styles 복사:
  - [Glob] {npm_root}/moai-adk/templates/.claude/output-styles/alfred/*.md → 4개
  - [Write] beginner-learning.md ✅
  - [Write] pair-collab.md ✅
  - [Write] study-deep.md ✅
  - [Write] moai-pro.md ✅
- [Glob] .claude/output-styles/alfred/*.md → 4개 ✅

---

#### Test 4: 훅 파일 권한 확인

**Given**:
- Unix 계열 시스템 (chmod 지원)

**When**:
- `/alfred:9-update` 실행

**Then**:
- Phase 4에서 chmod 실행:
  - [Bash] chmod +x .claude/hooks/alfred/*.cjs ✅
  - [Bash] ls -l .claude/hooks/alfred/ → `-rwxr-xr-x` (실행 권한 확인)

---

#### Test 5: 오류 복구 - 파일 누락

**Given**:
- 템플릿 디렉토리에서 일부 파일 누락 (시뮬레이션)

**When**:
- `/alfred:9-update` 실행
- Phase 5 검증 실패 (파일 개수 불일치)

**Then**:
- 오류 메시지: "❌ 검증 실패: 2개 파일 누락"
- 복구 옵션 제안:
  - "Phase 4 재실행"
  - "백업 복원"
  - "무시하고 진행"

---

#### Test 6: --check-quality 옵션

**Given**:
- 업데이트 완료 상태

**When**:
- `/alfred:9-update --check-quality` 실행

**Then**:
- Phase 6 실행:
  - [Alfred] @agent-trust-checker "Level 1" 호출
  - 결과: Pass / Warning / Critical
  - 결과별 조치 실행

---

### Error Scenario Tests (오류 시나리오 테스트)

#### Error 1: Write 도구 실패 (디렉토리 없음)

**Given**:
- .claude/commands/alfred/ 디렉토리가 없음

**When**:
- [Write] .claude/commands/alfred/1-spec.md 시도 → 실패

**Then**:
- [Bash] mkdir -p .claude/commands/alfred 실행 ✅
- [Write] 재시도 → 성공 ✅

---

#### Error 2: chmod 실패 (Windows)

**Given**:
- Windows 환경 (chmod 지원 안 함)

**When**:
- [Bash] chmod +x .claude/hooks/alfred/*.cjs 실행 → 실패

**Then**:
- 경고 메시지: "⚠️ chmod 실패 (Windows 환경에서는 정상)"
- 계속 진행 ✅ (치명적 오류 아님)

---

#### Error 3: Grep 도구 사용 불가

**Given**:
- Grep 도구가 설치되어 있지 않음

**When**:
- [Grep] "{{PROJECT_NAME}}" 시도 → 실패

**Then**:
- 자동으로 "무조건 백업 후 덮어쓰기" 모드로 전환 ✅
- 경고 메시지: "Grep을 사용할 수 없어 모든 파일을 백업합니다."

---

## Risks and Dependencies

### Critical Risks (P0)

#### Risk 1: Claude Code 도구 성능 저하

**위험**: 40개 파일 복사 시 시간 초과 (60초 제한)

**영향**: 업데이트 실패, 사용자 불만

**대응**:
- 파일별 타임아웃: 3초
- 전체 Phase 4 타임아웃: 60초
- 타임아웃 발생 시 백업 복원 제안

**우선순위**: Medium (성능 허용 범위 내)

---

#### Risk 2: 하위 호환성 문제

**위험**: 기존 사용자가 업데이트 후 호환성 문제 발생

**영향**: 기존 워크플로우 중단

**대응**:
- `/alfred:9-update` 인터페이스 동일 유지
- Phase 1-3, 5는 기존 로직 유지
- 테스트 시나리오에 하위 호환성 포함

**우선순위**: Critical (P0)

---

### Medium Risks (P1)

#### Risk 3: Grep 도구 미지원 환경

**위험**: Windows 등에서 Grep 사용 불가

**영향**: 프로젝트 문서 보호 기능 제한

**대응**:
- Grep 실패 시 자동으로 "무조건 백업" 모드
- 경고 메시지 출력

**우선순위**: Low (대안 존재)

---

### Dependencies

#### External Dependencies
- npm (버전 확인 및 패키지 설치)
- Claude Code 도구 ([Glob], [Read], [Write], [Bash], [Grep])
- trust-checker 에이전트 (--check-quality 옵션)

#### Internal Dependencies
- UpdateOrchestrator (Phase 1-3, 5)
- VersionChecker, BackupManager, NpmUpdater, UpdateVerifier
- 9-update.md 문서 (Alfred 실행 방식 명세)

---

## Rollback Strategy

### 롤백 조건

1. **Phase 4 실패**: 파일 복사 중 치명적 오류 발생
2. **Phase 5 검증 실패**: 파일 누락, 버전 불일치, 내용 손상
3. **사용자 요청**: 업데이트 결과에 만족하지 못함

### 롤백 절차

**자동 롤백** (Phase 4 실패 시):
```text
1. 오류 감지: "❌ Phase 4 실패: {오류 메시지}"
2. 사용자 확인: "백업에서 복원하시겠습니까? (Y/n)"
3. Y 선택 시:
   → [Bash] moai restore --from={timestamp}
   → "✅ 롤백 완료"
4. n 선택 시:
   → "⚠️ 불완전한 상태로 계속 진행합니다."
```

**수동 롤백** (사용자 요청):
```bash
# 백업 목록 확인
moai restore --list

# 특정 백업으로 복원
moai restore --from=2025-10-02-15-30-00

# 최근 백업으로 복원
moai restore --latest
```

---

## Next Steps

### After Implementation

1. **CHANGELOG.md 업데이트**:
   - v0.0.3 (or v0.1.0) 릴리스 노트 작성
   - Breaking Changes 섹션 (없음)
   - 개선 사항 나열

2. **패키지 배포**:
   ```bash
   npm version patch  # 0.0.2 → 0.0.3
   npm publish
   ```

3. **Living Document 생성**:
   ```bash
   /alfred:3-sync
   ```

4. **사용자 가이드 업데이트**:
   - `docs/cli/update.md` 문서 갱신
   - `--check-quality` 옵션 추가
   - 오류 복구 시나리오 예시 추가

5. **@TAG 체인 검증**:
   ```bash
   rg "@(SPEC|CODE|TEST):UPDATE-REFACTOR-001" -n
   ```

---

## Success Criteria

**리팩토링 완료 조건**:

1. ✅ `template-copier.ts` 제거 완료
2. ✅ `update-orchestrator.ts` Phase 4 로직 제거
3. ✅ `9-update.md` 문서 업데이트 (P0, P1 요구사항 반영)
4. ✅ 모든 AC (Acceptance Criteria) 통과
5. ✅ 테스트 커버리지 85% 이상 유지
6. ✅ @TAG 체인 무결성 검증
7. ✅ 하위 호환성 유지

**품질 기준**:
- 성능 저하 없음 (Claude Code 도구 사용으로 인한 속도는 허용)
- 오류 복구 전략이 모든 시나리오에 존재
- 사용자 경험 개선 (실시간 로그, 명확한 오류 메시지)

---

**END OF PLAN**
