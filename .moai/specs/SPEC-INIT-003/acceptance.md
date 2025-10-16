# SPEC-INIT-003 인수 기준 (v0.2.1)

> **@SPEC:INIT-003 Acceptance Criteria - 백업 조건 완화**
>
> 이 문서는 SPEC-INIT-003 "Init 백업 및 병합 옵션" 기능의 완료 조건을 정의합니다.

---

## 📋 Definition of Done (완료 조건)

### 필수 조건

**Phase A (moai init)**:
- ✅ OR 조건 백업 생성 테스트 통과 (1개라도 존재 시)
- ✅ 선택적 백업 테스트 통과 (Case 2~5)
- ✅ 메타데이터 `backed_up_files` 배열 검증 통과
- ✅ 테스트 커버리지 ≥85%

**Phase B (/alfred:8-project)**:
- ✅ 백업 감지 테스트 통과
- ✅ 긴급 백업 생성 테스트 통과
- ✅ 부분 백업 분석 테스트 통과
- ✅ 병합 로직 테스트 통과
- ✅ 테스트 커버리지 ≥85%

**공통 조건**:
- ✅ 타입 체크 통과 (`tsc --noEmit`)
- ✅ 린트 통과 (`biome check`)
- ✅ TRUST 5원칙 준수
  - **T**est First: Phase A/B 모든 로직에 테스트 존재
  - **R**eadable: 백업 및 병합 알고리즘 명확한 주석
  - **U**nified: TypeScript 타입 안전성 100%
  - **S**ecured: 파일 경로 검증, 백업 메타데이터 무결성
  - **T**rackable: `@CODE:INIT-003:BACKUP`, `@CODE:INIT-003:MERGE` TAG 부여
- ✅ TAG 체인 무결성: `@SPEC:INIT-003` → `@TEST:INIT-003` → `@CODE:INIT-003`

---

## 🧪 시나리오별 인수 기준

### AC-INIT-003-01: Phase A 백업 생성 (v0.2.1 업데이트)

**Given**: `.claude/`, `.moai/`, `CLAUDE.md` 중 **1개 이상** 존재하는 프로젝트
**When**: `moai init .` 실행
**Then**:

1. ✅ 백업 디렉토리 생성
   - **검증**: `.moai-backups/{timestamp}/` 존재
   - **내용**: 존재하는 파일/폴더만 선택적 백업

2. ✅ 백업 메타데이터 생성
   - **경로**: `.moai/backups/latest.json`
   - **내용**:
     ```json
     {
       "timestamp": "2025-10-07T14:30:00.000Z",
       "backup_path": ".moai-backups/20251007-143000",
       "backed_up_files": [".claude/", ".moai/", "CLAUDE.md"],
       "status": "pending",
       "created_by": "moai init"
     }
     ```

3. ✅ 사용자 안내 메시지 출력
   - **메시지**:
     ```
     ✅ 백업 완료: .moai-backups/20251007-143000
     📋 백업된 파일: .claude/, .moai/, CLAUDE.md

     ✅ MoAI-ADK 설치 완료!

     📦 기존 설정이 백업되었습니다:
        경로: .moai-backups/20251007-143000
        파일: .claude/, .moai/, CLAUDE.md

     🚀 다음 단계:
        1. Claude Code를 실행하세요
        2. /alfred:8-project 명령을 실행하세요
        3. 백업 내용을 병합할지 선택하세요
     ```

4. ✅ 템플릿 복사 완료
   - **검증**: `.claude/`, `.moai/`, `CLAUDE.md` 신규 템플릿으로 복사됨

**테스트 파일**: `__tests__/core/installer/phase-executor.test.ts`

---

### AC-INIT-003-02: Phase B 백업 감지 (/alfred:8-project)

**Given**: 백업 메타데이터 존재 (status: pending)
**When**: `/alfred:8-project` 실행
**Then**:

1. ✅ 백업 감지
   - **검증**: `.moai/backups/latest.json` 읽기 성공
   - **조건**: `status === 'pending'`

2. ✅ 백업 내용 분석 및 요약 표시
   - **출력**:
     ```
     📦 기존 설정 백업 발견

     **백업 시각**: 2025-10-07 14:30:00
     **백업 경로**: .moai-backups/20251007-143000

     **백업된 파일**:
     - .claude/
     - .moai/
     - CLAUDE.md
     ```

3. ✅ 병합 프롬프트 표시
   - **선택지**:
     - 병합 (Merge): 기존 설정 보존 + 신규 기능 추가
     - 새로 설치 (Reinstall): 백업 보존, 신규 템플릿 사용

**테스트 파일**: `__tests__/cli/commands/project/backup-merger.test.ts`

---

### AC-INIT-003-03: 병합 실행 (Phase B)

**Given**: 사용자가 "병합" 선택
**When**: 병합 프로세스 실행
**Then**:

1. ✅ JSON Deep Merge 성공
   - **대상**: `.claude/settings.json`, `.moai/config.json`
   - **검증**:
     - 신규 필드 추가됨
     - 기존 사용자 값 유지됨 (예: `mode: "personal"` 보존)

2. ✅ Markdown HISTORY 누적
   - **대상**: `CLAUDE.md`, `.moai/project/*.md`
   - **검증**: HISTORY 섹션에 신규 버전 추가, 기존 항목 보존

3. ✅ Hooks 버전 비교
   - **대상**: `.claude/hooks/**/*.cjs`
   - **검증**: 신규 버전이 높으면 업데이트

4. ✅ 병합 리포트 생성
   - **경로**: `.moai/reports/init-merge-report-{timestamp}.md`
   - **내용**: 병합된 파일, 보존된 파일 목록

5. ✅ 메타데이터 status 업데이트
   - **검증**: `.moai/backups/latest.json`의 `status: "merged"`

**테스트 파일**: `__tests__/cli/commands/project/merge-strategies/merge-orchestrator.test.ts`

---

### AC-INIT-003-04: 새로 설치 선택 (Phase B)

**Given**: 사용자가 "새로설치" 선택
**When**: 새로설치 프로세스 실행
**Then**:

1. ✅ 백업 보존
   - **검증**: `.moai-backups/{timestamp}/` 디렉토리 삭제되지 않음

2. ✅ 메타데이터 status 업데이트
   - **검증**: `.moai/backups/latest.json`의 `status: "ignored"`

3. ✅ 신규 템플릿 유지
   - **검증**: `.claude/`, `.moai/`, `CLAUDE.md`가 신규 템플릿 그대로 유지

4. ✅ 안내 메시지 표시
   - **메시지**: "새로 설치가 선택되었습니다. 백업은 보존되었습니다: .moai-backups/{timestamp}/"

**테스트 파일**: `__tests__/cli/commands/project/backup-merger.test.ts`

---

### AC-INIT-003-05: 부분 백업 시나리오 (v0.2.1 신규)

**Given**: `.claude/`만 존재 (`.moai/`, `CLAUDE.md` 없음)
**When**: `moai init .` 실행
**Then**:

1. ✅ 백업 디렉토리 생성
   - **검증**: `.moai-backups/{timestamp}/` 디렉토리 생성

2. ✅ 선택적 백업
   - **검증**: `.claude/`만 백업됨
   - **확인**: `.moai/`, `CLAUDE.md` 백업 시도 안 함

3. ✅ 메타데이터 `backed_up_files` 배열
   - **검증**: `backed_up_files: [".claude/"]`

4. ✅ 안내 메시지
   - **메시지**: "백업된 파일: .claude/"

**테스트 파일**: `__tests__/core/installer/phase-executor.test.ts`

---

### AC-INIT-003-06: /alfred:8-project 긴급 백업 (v0.2.1 신규)

**Given**: 백업 메타데이터 없음 AND `.moai/`만 존재
**When**: `/alfred:8-project` 실행
**Then**:

1. ✅ 기존 파일 감지
   - **검증**: `.moai/` 존재 확인

2. ✅ 긴급 백업 안내 메시지
   - **메시지**: "⚠️ 기존 MoAI-ADK 설정이 감지되었으나 백업이 없습니다."

3. ✅ 긴급 백업 생성
   - **검증**: `.moai-backups/{timestamp}/` 디렉토리 생성
   - **내용**: `.moai/`만 백업

4. ✅ 백업 메타데이터 생성
   - **검증**: `.moai/backups/latest.json` 생성
   - **내용**:
     ```json
     {
       "timestamp": "2025-10-07T14:30:00.000Z",
       "backup_path": ".moai-backups/20251007-143000",
       "backed_up_files": [".moai/"],
       "status": "pending",
       "created_by": "/alfred:8-project (emergency backup)"
     }
     ```

5. ✅ 백업 완료 메시지
   - **메시지**: "✅ 긴급 백업 완료: .moai-backups/{timestamp}"
   - **메시지**: "📋 백업된 파일: .moai/"

6. ✅ 병합 프롬프트 표시
   - **검증**: 긴급 백업 완료 후 자동으로 병합 프롬프트 진행

**테스트 파일**: `__tests__/cli/commands/project/backup-merger.test.ts`

---

### AC-INIT-003-07: 신규 설치 케이스 (v0.2.1 신규)

**Given**: `.claude/`, `.moai/`, `CLAUDE.md` 모두 없음
**When**: `moai init .` 실행
**Then**:

1. ✅ 백업 생략
   - **검증**: `.moai-backups/{timestamp}/` 디렉토리 생성 안 함

2. ✅ 신규 설치 메시지
   - **메시지**: "✨ 신규 프로젝트 설치"

3. ✅ 템플릿 직접 복사
   - **검증**: 백업 과정 없이 템플릿 즉시 복사

4. ✅ 백업 메타데이터 생성 안 함
   - **검증**: `.moai/backups/latest.json` 생성 안 함

**테스트 파일**: `__tests__/core/installer/phase-executor.test.ts`

---

### AC-INIT-003-08: 백업 생성 실패 처리 (Phase A)

**Given**: 디스크 공간 부족 또는 권한 문제
**When**: `moai init .` 실행 중 백업 생성 시도
**Then**:

1. ✅ 에러 감지
   - **검증**: 백업 생성 중 에러 발생 시 catch

2. ✅ 설치 중단
   - **검증**: 이후 템플릿 복사 로직 실행 안 함

3. ✅ 에러 메시지 표시
   - **메시지**: "백업 생성 실패: {에러 메시지}"
   - **안내**: "디스크 공간을 확인하거나 권한을 확인하세요."

4. ✅ 부분 백업 정리
   - **검증**: 불완전한 백업 디렉토리 삭제

**테스트 파일**: `__tests__/core/installer/phase-executor.test.ts`

---

### AC-INIT-003-09: 병합 중 오류 발생 시 롤백 (Phase B)

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

**테스트 파일**: `__tests__/cli/commands/project/merge-strategies/merge-orchestrator.test.ts`

---

## 🧪 테스트 매트릭스 (v0.2.1)

### Unit Test 커버리지

**Phase A 모듈**:
| 모듈 | 테스트 케이스 수 | 커버리지 목표 | 상태 |
|------|-----------------|--------------|------|
| `phase-executor.ts` (백업) | 7 (Case 2~5 추가) | ≥90% | ⏳ |
| `backup-metadata.ts` | 5 (`backed_up_files` 추가) | ≥90% | ⏳ |

**Phase B 모듈**:
| 모듈 | 테스트 케이스 수 | 커버리지 목표 | 상태 |
|------|-----------------|--------------|------|
| `backup-merger.ts` | 8 (긴급 백업 추가) | ≥85% | ⏳ |
| `json-merger.ts` | 8 | ≥90% | ⏳ |
| `markdown-merger.ts` | 6 | ≥90% | ⏳ |
| `hooks-merger.ts` | 5 | ≥85% | ⏳ |
| `merge-orchestrator.ts` | 10 | ≥85% | ⏳ |
| `merge-report.ts` | 4 | ≥85% | ⏳ |

### Integration Test

| 통합 시나리오 | 검증 항목 | 상태 |
|--------------|----------|------|
| Phase A 전체 | moai init → 선택적 백업 + 메타데이터 | ⏳ |
| Phase B 전체 | 백업 감지 → 긴급 백업 → 병합 → 리포트 | ⏳ |

### E2E Test

| E2E 시나리오 | 테스트 파일 | 상태 |
|-------------|------------|------|
| AC-INIT-003-01 (Phase A) | `phase-executor.e2e.test.ts` | ⏳ |
| AC-INIT-003-02 (백업 감지) | `backup-merger.e2e.test.ts` | ⏳ |
| AC-INIT-003-03 (병합) | `merge-mode.e2e.test.ts` | ⏳ |
| AC-INIT-003-04 (새로설치) | `reinstall-mode.e2e.test.ts` | ⏳ |
| AC-INIT-003-05 (부분 백업) | `selective-backup.test.ts` | ⏳ |
| AC-INIT-003-06 (긴급 백업) | `emergency-backup.test.ts` | ⏳ |
| AC-INIT-003-07 (신규 설치) | `fresh-install.test.ts` | ⏳ |
| AC-INIT-003-08 (백업 실패) | `backup-failure.test.ts` | ⏳ |
| AC-INIT-003-09 (롤백) | `merge-rollback.test.ts` | ⏳ |

---

## 🔍 품질 게이트 (v0.2.1)

### 자동 검증 항목

#### 1. 코드 품질
```bash
# 타입 체크
bun run type-check

# 린트
bun run lint

# Phase A 테스트 (Case 2~7 포함)
bun test phase-executor backup-metadata

# Phase B 테스트 (긴급 백업 포함)
bun test backup-merger merge-strategies

# 전체 커버리지
bun run test:coverage
```

#### 2. TAG 무결성
```bash
# Phase A TAG 체인 검증
rg '@CODE:INIT-003:BACKUP' -n moai-adk-ts/src/core/installer/

# Phase B TAG 체인 검증
rg '@CODE:INIT-003:MERGE' -n moai-adk-ts/src/cli/commands/project/

# 예상 결과:
# .moai/specs/SPEC-INIT-003/spec.md:@SPEC:INIT-003
# __tests__/core/installer/*.test.ts:@TEST:INIT-003:BACKUP
# __tests__/cli/commands/project/*.test.ts:@TEST:INIT-003:MERGE
# src/core/installer/*.ts:@CODE:INIT-003:BACKUP
# src/cli/commands/project/*.ts:@CODE:INIT-003:MERGE
```

#### 3. 백업 메타데이터 무결성
```bash
# 메타데이터 스키마 검증
bun test backup-metadata

# 예상 메타데이터 형식:
# {
#   "timestamp": "2025-10-07T14:30:00.000Z",
#   "backup_path": ".moai-backups/20251007-143000",
#   "backed_up_files": [".claude/", ".moai/", "CLAUDE.md"],
#   "status": "pending" | "merged" | "ignored",
#   "created_by": "moai init" | "/alfred:8-project (emergency backup)"
# }
```

### 수동 검증 항목

#### 1. Phase A 사용자 경험
- [ ] 백업 완료 메시지가 명확한가? (백업된 파일 목록 표시)
- [ ] 다음 단계 안내가 명확한가? (Claude Code → /alfred:8-project)
- [ ] 백업 실패 시 에러 메시지가 해결 방법을 제시하는가?

#### 2. Phase B 사용자 경험
- [ ] 백업 분석 요약이 이해하기 쉬운가?
- [ ] 긴급 백업 메시지가 명확한가?
- [ ] 병합 프롬프트가 선택지를 명확히 설명하는가?
- [ ] 병합 리포트가 가독성이 좋은가?

#### 3. 성능
- [ ] Phase A 백업 생성: < 2초 (일반 프로젝트)
- [ ] Phase B 백업 감지: < 100ms
- [ ] Phase B 긴급 백업 생성: < 3초
- [ ] Phase B 병합 실행: < 5초 (10개 파일 기준)

---

## 📊 성능 기준 (v0.2.1)

### 실행 시간 목표

**Phase A (moai init)**:
| 작업 | 목표 시간 | 측정 방법 |
|------|----------|----------|
| 백업 조건 감지 (OR) | < 50ms | `fs.existsSync()` 3회 |
| 선택적 백업 생성 | < 1초 | 실제 파일 크기 기준 |
| 메타데이터 저장 | < 100ms | `performance.now()` |
| 전체 Phase A | < 2초 | 일반 프로젝트 기준 |

**Phase B (/alfred:8-project)**:
| 작업 | 목표 시간 | 측정 방법 |
|------|----------|----------|
| 백업 감지 | < 100ms | `performance.now()` |
| 긴급 백업 생성 | < 3초 | 실제 파일 크기 기준 |
| 백업 분석 | < 500ms | 파일 읽기 기준 |
| JSON 병합 | < 500ms | 파일당 평균 |
| Markdown 병합 | < 500ms | 파일당 평균 |
| 전체 Phase B | < 5초 | 10개 파일 기준 |

### 메모리 사용량 목표

- **Phase A**: 50MB 이하 (백업 중)
- **Phase B**: 100MB 이하 (병합 중)
- **백업 디스크 사용**: 원본 크기의 1.1배 이하

---

## 🎯 최종 체크리스트 (v0.2.1)

### 기능 완성도
- [ ] Phase A 모든 AC (01, 05, 07, 08) 테스트 통과
- [ ] Phase B 모든 AC (02, 03, 04, 06, 09) 테스트 통과
- [ ] 통합 테스트 100% 통과
- [ ] E2E 테스트 100% 통과 (9개 시나리오)

### 코드 품질
- [ ] Phase A 테스트 커버리지 ≥85%
- [ ] Phase B 테스트 커버리지 ≥85%
- [ ] 타입 체크 통과
- [ ] 린트 통과
- [ ] TRUST 5원칙 준수

### 문서화
- [ ] `spec.md` v0.2.1 업데이트 완료
- [ ] `plan.md` v0.2.1 업데이트 완료
- [ ] `acceptance.md` (본 문서) v0.2.1 업데이트 완료
- [ ] TAG 체인 무결성 확인 (BACKUP, MERGE 분리)

### 사용자 경험
- [ ] Phase A 안내 메시지 명확성 (백업된 파일 목록)
- [ ] Phase B 긴급 백업 메시지 명확성
- [ ] Phase B 병합 프롬프트 명확성
- [ ] 리포트 가독성
- [ ] 성능 기준 충족

### 다음 단계 준비
- [ ] `/alfred:3-sync` 실행 가능
- [ ] README 업데이트 (긴급 백업 시나리오 설명)
- [ ] CHANGELOG v0.2.1 작성 (백업 조건 완화 강조)

---

## 🚀 배포 전 최종 검증 (v0.2.1)

### 실제 프로젝트 테스트 (Case 2~5)

**Case 2: .claude만 존재**
1. **Phase A**: `moai init .` 실행
   - 백업: `.claude/`만 백업 확인
   - 메타데이터: `backed_up_files: [".claude/"]` 확인
2. **Phase B**: `/alfred:8-project` 실행
   - 백업 감지 및 분석 확인
   - 병합 결과 확인

**Case 3: .moai, CLAUDE.md만 존재**
1. **Phase A**: `moai init .` 실행
   - 백업: `.moai/`, `CLAUDE.md`만 백업 확인
   - 메타데이터: `backed_up_files: [".moai/", "CLAUDE.md"]` 확인

**Case 4: CLAUDE.md만 존재**
1. **Phase A**: `moai init .` 실행
   - 백업: `CLAUDE.md`만 백업 확인
   - 메타데이터: `backed_up_files: ["CLAUDE.md"]` 확인

**Case 5: 신규 설치**
1. **Phase A**: `moai init .` 실행
   - 백업 생략 확인
   - 메타데이터 생성 안 함 확인

**긴급 백업 시나리오**
1. **Phase B**: moai init 없이 `/alfred:8-project` 직접 실행
   - 긴급 백업 자동 생성 확인
   - 메타데이터 `created_by: "/alfred:8-project (emergency backup)"` 확인

### 회귀 테스트
- [ ] 기존 init 기능 (신규 프로젝트, 백업 없음) 정상 작동
- [ ] 다른 CLI 명령어(`doctor`, `status`) 영향 없음

---

**최종 승인 조건**: 위의 모든 체크리스트 항목이 ✅ 상태일 때 SPEC-INIT-003 v0.2.1 완료로 간주합니다.

---

_이 인수 기준은 `/alfred:2-build INIT-003` 구현 중 Phase A → Phase B 순차적으로 검증되어야 합니다._
