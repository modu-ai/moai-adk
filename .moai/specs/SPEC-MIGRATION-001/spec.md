---
id: SPEC-MIGRATION-001
version: 1.0.0
status: draft
created: 2025-11-18
updated: 2025-11-18
author: GoosLab
priority: high
---

# SPEC-MIGRATION-001: Alfred → Moai 폴더 구조 자동 마이그레이션

## HISTORY

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| v1.0.0 | 2025-11-18 | GoosLab | 초기 SPEC 생성 - alfred → moai 마이그레이션 자동화 |

---

## Environment (환경)

- **Target System**: MoAI-ADK v0.26.0+
- **Python Version**: >= 3.11
- **Platform**: macOS, Linux, Windows
- **Execution Context**: `moai-adk update` 명령어 실행 시
- **Dependencies**: BackupManager, TemplateProcessor, shutil, json, pathlib, datetime

---

## Assumptions (가정)

1. 사용자는 v0.26.x 이하의 MoAI-ADK를 사용 중이며 alfred/ 폴더가 존재
2. 사용자는 `moai-adk update` 명령어로 마이그레이션을 실행
3. 프로젝트의 `.moai/config/config.json` 파일이 쓰기 가능한 상태
4. `.moai/backups/` 디렉토리가 쓰기 가능한 상태
5. 마이그레이션 실패 시 전체 프로젝트 복원이 가능해야 함
6. 패키지 템플릿에는 moai/ 폴더 구조가 v0.27.0부터 존재

---

## Requirements (EARS 형식 요구사항)

### 1. Ubiquitous Requirements (항상 적용)

**R1.1**: 시스템은 마이그레이션 프로세스 시작 전 항상 전체 프로젝트를 자동으로 백업해야 함
- Location: `.moai/backups/alfred_to_moai_migration_{timestamp}/`
- Format: 전체 프로젝트 디렉토리 복사
- Action: 백업 완료 후에만 마이그레이션 진행

**R1.2**: 시스템은 마이그레이션 완료 후 `config.json`에 마이그레이션 상태를 기록해야 함
- Field: `migration.alfred_to_moai.migrated = true`
- Include: `timestamp`, `folders_installed`, `folders_removed`, `package_version`
- Action: 중복 실행 방지를 위한 상태 기록

---

### 2. Event-Driven Requirements (이벤트 발생 시)

**R2.1**: WHEN alfred 폴더 감지 → alfred 폴더가 존재하는 프로젝트에서 update 명령어 실행 시
- THEN: 자동 백업 생성 및 마이그레이션 프로세스 시작
- Action: alfred 폴더 존재 여부 자동 감지

**R2.2**: WHEN 백업 완료 → 마이그레이션 진행
- THEN: 패키지 템플릿의 moai/ 폴더를 신규 설치
- Action: template → project/.claude/ 복사

**R2.3**: WHEN moai 폴더 설치 완료 → alfred 폴더 정리
- THEN: alfred 폴더 삭제 (백업은 보존)
- Action: 레거시 폴더 자동 정리

---

### 3. Unwanted Requirements (원치 않는 상황)

**R3.1**: IF 백업 생성 실패 → 마이그레이션 중단
- Condition: `.moai/backups/` 쓰기 권한 없음 또는 디스크 공간 부족
- THEN: 에러 메시지 출력 + 마이그레이션 전체 중단
- Action: 어떤 파일도 수정하지 않음 (롤백 불필요)

**R3.2**: IF moai 템플릿 설치 실패 → 자동 롤백 실행
- Condition: 복사 중 I/O 에러, 경로 오류 등
- THEN: 백업에서 전체 프로젝트 복원
- Action: 자동 롤백 후 에러 로그 출력

**R3.3**: IF settings.json 업데이트 실패 → 마이그레이션 롤백
- Condition: JSON 파싱 실패 또는 파일 접근 오류
- THEN: 백업에서 복원 + 에러 로그
- Action: 부분 성공 상태 방지

---

### 4. State-Driven Requirements (특정 상태)

**R4.1**: WHILE 마이그레이션 진행 중 → 진행 상황을 실시간으로 표시
- State: 진행 중
- Display: 콘솔에 단계별 메시지 출력
- Format: "✅ Migrated: .claude/commands/alfred/" 형식

**R4.2**: WHILE 롤백 진행 중 → 복원 상태를 실시간으로 표시
- State: 실패 후 복원 중
- Display: 콘솔에 롤백 단계별 메시지
- Format: "🔄 Rolling back..."

**R4.3**: WHILE config.json 마이그레이션 상태 기록 중 → 중단 방지
- State: 최종 상태 기록
- Action: 기록 중 에러 발생 시 로그 출력만 하고 계속 진행

---

### 5. Optional Requirements (선택적)

**R5.1**: WHERE `settings.json`에 alfred Hook 경로 존재 → 자동 업데이트
- Condition: `.claude/hooks/alfred/session_start__*` 등의 경로 존재
- THEN: 모든 alfred → moai 경로로 자동 변경
- Action: 스마트 경로 치환 (정확도 99% 이상)

**R5.2**: WHERE 이미 마이그레이션이 완료됨 (config.json에 기록) → 스킵
- Condition: `migration.alfred_to_moai.migrated == true`
- THEN: "이미 마이그레이션 완료됨" 메시지 출력
- Action: 불필요한 재실행 방지

**R5.3**: WHERE `.moai/backups/` 디렉토리에 이전 백업 존재 → 정리 옵션 제공
- Condition: 30일 이상 된 백업 폴더 감지
- THEN: 사용자에게 정리 여부 문의 (선택 사항)
- Action: 선택에 따라 정리 또는 유지

---

## Specifications (기술 명세)

### 1. 마이그레이션 클래스 구현

**파일**: `src/moai_adk/core/migration/alfred_to_moai_migrator.py`

**주요 메서드**:
```python
class AlfredToMoaiMigrator:
    def __init__(self, project_root: Path)
    def needs_migration(self) -> bool
    def execute_migration(self, backup_path: Path) -> bool
    def _delete_alfred_folders(self) -> None
    def _update_settings_json_hooks(self) -> None
    def _verify_migration(self) -> bool
```

**책임**:
- alfred 폴더 존재 여부 감지
- 마이그레이션 상태 확인 및 기록
- 폴더 삭제 및 경로 업데이트
- 마이그레이션 검증

### 2. update.py 통합

**파일**: `src/moai_adk/cli/commands/update.py`

**변경 지점**: `_sync_templates()` 함수 내부

**통합 위치**:
```
1. 백업 생성 (기존)
2. → 마이그레이션 로직 추가 (새로운 위치)
3. 템플릿 동기화 (기존)
4. 검증 및 정리 (기존)
```

**추가 코드**:
- `migrator = AlfredToMoaiMigrator(project_path)` 인스턴스 생성
- `if migrator.needs_migration()` 조건 체크
- `migrator.execute_migration(backup_path)` 실행
- 실패 시 롤백: `backup.restore_backup(backup_path)`

### 3. 설정 및 상태 관리

**config.json 업데이트 필드**:
```json
{
  "migration": {
    "alfred_to_moai": {
      "migrated": true,
      "timestamp": "2025-11-18 18:30:00",
      "folders_installed": 3,
      "folders_removed": 3,
      "package_version": "0.27.0"
    }
  }
}
```

### 4. 에러 처리 및 롤백

**백업 실패**:
- 에러 메시지 출력
- 마이그레이션 중단 (아무것도 수정하지 않음)

**마이그레이션 중 실패**:
- 백업 복원
- 에러 로그 출력
- 사용자에게 안내

**롤백 검증**:
- 복원된 파일 확인
- 마이그레이션 상태 롤백 (config.json 초기화)

---

## Acceptance Criteria (수용 기준)

### 테스트 시나리오 1: 정상 마이그레이션 ✅

**시나리오**: alfred 폴더가 존재하는 v0.26.x 프로젝트에서 update 명령어 실행

**Given**:
- `.claude/commands/alfred/` 폴더 존재
- `.claude/agents/alfred/` 폴더 존재
- `.claude/hooks/alfred/` 폴더 존재
- `settings.json`에 alfred Hook 경로 존재
- `config.json`에 마이그레이션 기록 없음

**When**:
- `moai-adk update` 명령어 실행

**Then**:
- ✅ 백업 생성됨 (`.moai/backups/alfred_to_moai_migration_{timestamp}/`)
- ✅ moai 템플릿 설치됨 (`.claude/commands/moai/`, `.claude/agents/moai/`, `.claude/hooks/moai/`)
- ✅ alfred 폴더 삭제됨 (모두 3개)
- ✅ `settings.json` Hook 경로 업데이트됨 (alfred → moai)
- ✅ `config.json`에 마이그레이션 상태 기록됨
- ✅ 사용자에게 완료 메시지 표시됨

---

### 테스트 시나리오 2: 백업 실패 시 중단 ✅

**시나리오**: 백업 디렉토리 접근 불가

**Given**:
- `.moai/backups/` 디렉토리 쓰기 권한 없음

**When**:
- `moai-adk update` 명령어 실행

**Then**:
- ✅ 백업 실패 에러 메시지 출력
- ✅ 마이그레이션 중단 (alfred 폴더 유지)
- ✅ 어떤 파일도 수정되지 않음

---

### 테스트 시나리오 3: 마이그레이션 중 실패 및 롤백 ✅

**시나리오**: moai 템플릿 설치 중 에러 발생

**Given**:
- alfred 폴더 존재
- 백업 완료됨
- moai 템플릿 복사 중 I/O 에러 발생 (시뮬레이션)

**When**:
- 마이그레이션 진행 중 에러 발생

**Then**:
- ✅ 자동 롤백 트리거됨
- ✅ 백업에서 전체 프로젝트 복원됨
- ✅ alfred 폴더 유지됨
- ✅ `config.json` 마이그레이션 상태 롤백됨
- ✅ 에러 로그 출력됨

---

### 테스트 시나리오 4: 중복 실행 방지 ✅

**시나리오**: 이미 마이그레이션된 프로젝트에서 재실행

**Given**:
- `config.json`에 `migration.alfred_to_moai.migrated = true`
- alfred 폴더 없음 (이미 삭제됨)
- moai 폴더 존재함

**When**:
- `moai-adk update` 명령어 재실행

**Then**:
- ✅ 마이그레이션 로직 스킵됨
- ✅ "이미 마이그레이션 완료" 메시지 표시
- ✅ 불필요한 재실행 방지

---

### 테스트 시나리오 5: settings.json Hook 경로 업데이트 ✅

**시나리오**: settings.json의 모든 alfred Hook 경로가 moai로 변경

**Given**:
- `settings.json`에 다음 경로 존재:
  ```
  uv run {{PROJECT_DIR}}/.claude/hooks/alfred/session_start__*.py
  uv run {{PROJECT_DIR}}/.claude/hooks/alfred/pre_tool__*.py
  ```

**When**:
- 마이그레이션 실행

**Then**:
- ✅ 모든 `alfred` 경로가 `moai`로 변경됨
- ✅ 경로 구조는 유지됨 (`.claude/hooks/moai/`)
- ✅ 변경 없는 부분은 유지됨

---

## Context & Notes (컨텍스트 및 주의사항)

### 기술적 도전과제

1. **정확한 경로 치환**: settings.json 내 경로를 정확하게 찾고 변경해야 함
2. **롤백 안정성**: 부분 실패 상황에서 완전한 복원 보장
3. **대용량 파일 처리**: 프로젝트 크기가 클 경우 백업/복원 시간 고려

### 기존 코드 재사용

- `BackupManager`: 백업 및 복원 기능
- `TemplateProcessor`: 템플릿 구조 및 복사 로직
- `VersionMigrator`: 마이그레이션 패턴 참고

### 점진적 배포 전략

1. **v0.27.0-alpha**: 마이그레이션 로직 구현 (내부 테스트)
2. **v0.27.0-beta**: 사용자 베타 테스트
3. **v0.27.0**: 정식 릴리스 (자동 마이그레이션 적용)
4. **v0.28.0**: alfred 템플릿 폴더 삭제 (레거시 정리)

---

## Related SPECs & Documents

- SPEC-TEMPLATE-001: Package Template moai/ Structure (선행 작업)
- CHANGELOG: v0.27.0 변경사항 기록
- MIGRATION_GUIDE: 사용자 마이그레이션 가이드

---

## End of Spec