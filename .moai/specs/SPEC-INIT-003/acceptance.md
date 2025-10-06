# SPEC-INIT-003 인수 기준

> **@SPEC:INIT-003 Acceptance Criteria**
>
> 이 문서는 SPEC-INIT-003 "Init 백업 및 병합 옵션" 기능의 완료 조건을 정의합니다.

---

## 📋 Definition of Done (완료 조건)

### 필수 조건

- ✅ 모든 테스트 통과 (Unit, Integration, E2E)
- ✅ 테스트 커버리지 ≥85%
- ✅ 타입 체크 통과 (`tsc --noEmit`)
- ✅ 린트 통과 (`biome check`)
- ✅ TRUST 5원칙 준수
  - **T**est First: 모든 병합 로직에 테스트 존재
  - **R**eadable: 병합 알고리즘 명확한 주석
  - **U**nified: TypeScript 타입 안전성 100%
  - **S**ecured: 파일 경로 검증, 백업 무결성
  - **T**rackable: `@CODE:INIT-003` TAG 부여
- ✅ TAG 체인 무결성: `@SPEC:INIT-003` → `@TEST:INIT-003` → `@CODE:INIT-003`

---

## 🧪 시나리오별 인수 기준

### Scenario 1: 병합 모드 선택 및 실행

**Given**: 기존 `.claude/`, `.moai/`, `CLAUDE.md`가 존재하는 프로젝트
**When**: 사용자가 `moai init .` 실행 후 "병합" 옵션 선택
**Then**:

1. ✅ 백업 디렉토리 생성
   - **검증**: `.moai-backup-{timestamp}/` 존재
   - **내용**: 기존 `.claude/`, `.moai/`, `CLAUDE.md` 복사본

2. ✅ JSON 파일 깊은 병합
   - **대상**: `.claude/settings.json`, `.moai/config.json`
   - **검증**:
     - 신규 필드 추가됨
     - 기존 사용자 값 유지됨 (예: `mode: "personal"` 보존)
     - 중첩 객체 재귀 병합됨 (예: `hooks` 객체)

3. ✅ Markdown 파일 섹션 병합
   - **대상**: `CLAUDE.md`, `.moai/project/*.md`
   - **검증**:
     - HISTORY 섹션에 신규 버전 누적됨
     - 기존 HISTORY 항목 보존됨
     - 중복 섹션 제거됨

4. ✅ Hooks 파일 버전 비교
   - **대상**: `.claude/hooks/**/*.cjs`
   - **검증**:
     - 신규 버전이 높으면 업데이트됨
     - 동일 버전이면 기존 유지됨
     - 버전 로그 출력됨

5. ✅ Commands 파일 커스터마이징 보존
   - **대상**: `.claude/commands/**/*.md`
   - **검증**:
     - 사용자 커스터마이징 파일은 보존됨
     - 템플릿 원본은 신규로 교체됨

6. ✅ 변경 내역 리포트 생성
   - **경로**: `.moai/reports/init-merge-report-{timestamp}.md`
   - **내용**:
     - 병합된 파일 목록
     - 덮어쓴 파일 목록
     - 보존된 파일 목록
     - 변경 사유 설명

7. ✅ 성공 메시지 표시
   - **메시지**: "병합 완료! 변경 내역: .moai/reports/init-merge-report-{timestamp}.md"
   - **백업 경로**: `.moai-backup-{timestamp}/`

**테스트 파일**: `__tests__/cli/init/merge-mode.e2e.test.ts`

---

### Scenario 2: 새로 설치 모드 선택 및 실행

**Given**: 기존 `.claude/`, `.moai/`, `CLAUDE.md`가 존재하는 프로젝트
**When**: 사용자가 `moai init .` 실행 후 "새로 설치" 옵션 선택
**Then**:

1. ✅ 백업 디렉토리 생성
   - **검증**: `.moai-backup-{timestamp}/` 존재
   - **내용**: 기존 파일 전체 복사본

2. ✅ 신규 템플릿으로 덮어쓰기
   - **대상**: `.claude/`, `.moai/`, `CLAUDE.md`
   - **검증**: 모든 파일이 템플릿 원본과 동일

3. ✅ 변경 내역 리포트 생성
   - **경로**: `.moai/reports/init-merge-report-{timestamp}.md`
   - **내용**:
     - "덮어쓴 파일" 목록에 모든 파일 포함
     - 백업 경로 안내

4. ✅ 성공 메시지 표시
   - **메시지**: "새로 설치 완료! 백업: .moai-backup-{timestamp}/"
   - **안내**: "기존 설정 복원이 필요하면 백업 디렉토리를 참조하세요."

**테스트 파일**: `__tests__/cli/init/reinstall-mode.e2e.test.ts`

---

### Scenario 3: 취소 선택

**Given**: 기존 `.claude/`, `.moai/`, `CLAUDE.md`가 존재하는 프로젝트
**When**: 사용자가 `moai init .` 실행 후 "취소" 옵션 선택
**Then**:

1. ✅ 설치 중단
   - **검증**: 프로세스 즉시 종료
   - **파일 변경 없음**: 기존 파일이 전혀 수정되지 않음

2. ✅ 백업 생성 안 함
   - **검증**: `.moai-backup-*` 디렉토리가 생성되지 않음

3. ✅ 취소 메시지 표시
   - **메시지**: "설치가 취소되었습니다. 기존 파일은 변경되지 않았습니다."

**테스트 파일**: `__tests__/cli/init/cancel-mode.e2e.test.ts`

---

### Scenario 4: 백업 생성 실패 처리

**Given**: 디스크 공간 부족 또는 권한 문제
**When**: 사용자가 병합 또는 새로 설치 선택
**Then**:

1. ✅ 에러 감지
   - **검증**: 백업 생성 중 에러 발생 시 catch

2. ✅ 설치 중단
   - **검증**: 이후 병합/덮어쓰기 로직 실행 안 함

3. ✅ 에러 메시지 표시
   - **메시지**: "백업 생성 실패: {에러 메시지}"
   - **안내**: "디스크 공간을 확인하거나 권한을 확인하세요."

4. ✅ 부분 백업 롤백
   - **검증**: 불완전한 백업 디렉토리 삭제

**테스트 파일**: `__tests__/cli/init/backup-failure.test.ts`

---

### Scenario 5: 병합 중 충돌 발생

**Given**: 자동 병합이 불가능한 JSON 스키마 충돌
**When**: 병합 엔진이 충돌을 감지
**Then**:

1. ✅ 충돌 감지 로그
   - **메시지**: "충돌 감지: .claude/settings.json"

2. ✅ 충돌 파일 목록 표시
   - **출력**:
     ```
     ⚠️ 다음 파일에서 충돌이 발견되었습니다:
     - .claude/settings.json (JSON 스키마 불일치)
     ```

3. ✅ 수동 해결 가이드 제공
   - **안내**:
     ```
     수동 해결이 필요합니다:
     1. 백업: .moai-backup-{timestamp}/.claude/settings.json
     2. 신규 템플릿: {template-path}/.claude/settings.json
     3. 원하는 설정으로 직접 병합하세요.
     ```

4. ✅ 부분 병합 적용
   - **검증**: 충돌 없는 파일은 정상 병합됨

5. ✅ 리포트에 충돌 기록
   - **섹션**: "충돌 파일 (Conflicts)"
   - **내용**: 충돌 파일 목록 및 해결 방법

**테스트 파일**: `__tests__/cli/init/merge-conflict.test.ts`

---

### Scenario 6: 병합 중 오류 발생 시 롤백

**Given**: 병합 중 파일 쓰기 오류 (권한 문제 등)
**When**: 병합 엔진에서 치명적 오류 발생
**Then**:

1. ✅ 오류 감지 및 로깅
   - **메시지**: "병합 중 오류 발생: {에러 메시지}"

2. ✅ 자동 롤백 시작
   - **메시지**: "백업에서 복원 중..."
   - **동작**: 백업 디렉토리 내용을 원래 위치로 복사

3. ✅ 롤백 완료 확인
   - **검증**: 모든 파일이 병합 시작 전 상태로 복원됨

4. ✅ 롤백 성공 메시지
   - **메시지**: "롤백 완료. 기존 파일이 복원되었습니다."
   - **로그**: 에러 원인 및 해결 방법 안내

**테스트 파일**: `__tests__/cli/init/merge-rollback.test.ts`

---

## 🧪 테스트 매트릭스

### Unit Test 커버리지

| 모듈 | 테스트 케이스 수 | 커버리지 목표 | 상태 |
|------|-----------------|--------------|------|
| `installation-detector.ts` | 5 | ≥90% | ⏳ |
| `merge-prompt.ts` | 3 | ≥85% | ⏳ |
| `json-merger.ts` | 8 | ≥90% | ⏳ |
| `markdown-merger.ts` | 6 | ≥90% | ⏳ |
| `hooks-merger.ts` | 5 | ≥85% | ⏳ |
| `merge-orchestrator.ts` | 10 | ≥85% | ⏳ |
| `report-generator.ts` | 4 | ≥85% | ⏳ |

### Integration Test

| 통합 시나리오 | 검증 항목 | 상태 |
|--------------|----------|------|
| Phase 1 + Phase 2 | 감지 → 병합 | ⏳ |
| Phase 2 + Phase 3 | 병합 → 리포트 | ⏳ |
| Phase 1~4 통합 | 전체 플로우 | ⏳ |

### E2E Test

| E2E 시나리오 | 테스트 파일 | 상태 |
|-------------|------------|------|
| Scenario 1 (병합) | `merge-mode.e2e.test.ts` | ⏳ |
| Scenario 2 (재설치) | `reinstall-mode.e2e.test.ts` | ⏳ |
| Scenario 3 (취소) | `cancel-mode.e2e.test.ts` | ⏳ |
| Scenario 4 (백업 실패) | `backup-failure.test.ts` | ⏳ |
| Scenario 5 (충돌) | `merge-conflict.test.ts` | ⏳ |
| Scenario 6 (롤백) | `merge-rollback.test.ts` | ⏳ |

---

## 🔍 품질 게이트

### 자동 검증 항목

#### 1. 코드 품질
```bash
# 타입 체크
bun run type-check

# 린트
bun run lint

# 테스트
bun run test

# 커버리지
bun run test:coverage
```

#### 2. TAG 무결성
```bash
# TAG 체인 검증
rg '@(SPEC|TEST|CODE):INIT-003' -n moai-adk-ts/

# 예상 결과:
# .moai/specs/SPEC-INIT-003/spec.md:1:@SPEC:INIT-003
# __tests__/cli/init/*.test.ts:@TEST:INIT-003
# src/cli/commands/init/*.ts:@CODE:INIT-003
```

#### 3. 파일 무결성
```bash
# 병합 후 파일 검증
moai doctor

# 예상 결과:
# ✅ .claude/ 디렉토리 존재
# ✅ .moai/ 디렉토리 존재
# ✅ CLAUDE.md 존재
# ✅ Hooks 버전 일관성
```

### 수동 검증 항목

#### 1. 사용자 경험
- [ ] 프롬프트 메시지가 명확하고 이해하기 쉬운가?
- [ ] 진행 상황 표시가 적절한가?
- [ ] 에러 메시지가 해결 방법을 제시하는가?

#### 2. 리포트 가독성
- [ ] 변경 내역 리포트가 이해하기 쉬운가?
- [ ] 파일별 변경 사유가 명확한가?
- [ ] 다음 단계 안내가 포함되어 있는가?

#### 3. 성능
- [ ] 병합 시간이 5초 이내인가? (일반적인 프로젝트)
- [ ] 백업 생성이 2초 이내인가?
- [ ] 리포트 생성이 1초 이내인가?

---

## 📊 성능 기준

### 실행 시간 목표

| 작업 | 목표 시간 | 측정 방법 |
|------|----------|----------|
| 기존 설치 감지 | < 100ms | `performance.now()` |
| 백업 생성 | < 2초 | 실제 파일 크기 기준 |
| JSON 병합 | < 500ms | 파일당 평균 |
| Markdown 병합 | < 500ms | 파일당 평균 |
| 전체 병합 플로우 | < 5초 | 10개 파일 기준 |
| 리포트 생성 | < 1초 | - |

### 메모리 사용량 목표

- **최대 메모리**: 100MB 이하 (병합 중)
- **백업 디스크 사용**: 원본 크기의 1.1배 이하

---

## 🎯 최종 체크리스트

### 기능 완성도
- [ ] 모든 시나리오 (1~6) 테스트 통과
- [ ] 통합 테스트 100% 통과
- [ ] E2E 테스트 100% 통과

### 코드 품질
- [ ] 테스트 커버리지 ≥85%
- [ ] 타입 체크 통과
- [ ] 린트 통과
- [ ] TRUST 5원칙 준수

### 문서화
- [ ] `spec.md` 최신화
- [ ] `plan.md` 완료 상태 업데이트
- [ ] `acceptance.md` (본 문서) 검증 완료
- [ ] TAG 체인 무결성 확인

### 사용자 경험
- [ ] 에러 메시지 명확성
- [ ] 리포트 가독성
- [ ] 성능 기준 충족

### 다음 단계 준비
- [ ] `/alfred:3-sync` 실행 가능
- [ ] `moai doctor` 검증 통과
- [ ] README 업데이트 준비

---

## 🚀 배포 전 최종 검증

### 실제 프로젝트 테스트
1. **기존 MoAI-ADK 프로젝트**에서 `moai init .` 실행
2. 병합 모드 선택 및 결과 확인
3. `git diff`로 변경사항 검토
4. `moai doctor` 실행하여 무결성 확인

### 회귀 테스트
- [ ] 기존 init 기능 (신규 프로젝트) 정상 작동
- [ ] 다른 CLI 명령어(`doctor`, `status`) 영향 없음

---

**최종 승인 조건**: 위의 모든 체크리스트 항목이 ✅ 상태일 때 SPEC-INIT-003 완료로 간주합니다.

---

_이 인수 기준은 `/alfred:2-build INIT-003` 구현 중 지속적으로 검증되어야 합니다._
