---
id: SPEC-MIGRATION-001
type: implementation-plan
version: 1.0.0
created: 2025-11-18
updated: 2025-11-18
author: GoosLab
total_duration: 10 days
---

# SPEC-MIGRATION-001 구현 계획

## 개요

Alfred → Moai 폴더 구조 자동 마이그레이션을 4단계로 구현합니다.

**총 소요 시간**: 10 작업일 (병렬 작업 가능)

**브랜치**: release/0.26.0 (새 브랜치 없음)

**타겟 버전**: v0.27.0 (마이그레이션 로직 포함)

---

## Phase 1: 패키지 템플릿 Moai 구조 생성 (3일)

**목표**: Package template에 moai/ 폴더 구조를 생성하고 모든 파일을 복사

**소요 시간**: 3 작업일

### 1.1 Commands 폴더 생성 및 복사

**위치**: `src/moai_adk/templates/.claude/commands/moai/`

**작업**:
- ✅ 디렉토리 생성: `commands/moai/`
- ✅ Alfred에서 5개 파일 복사 및 내용 수정:
  - `0-project.md` - 진행상황 표시바 설명 업데이트 (moai 용어로)
  - `1-plan.md` - Alfred → Moai 용어 변경
  - `2-run.md` - SPEC-First TDD 프로세스 설명
  - `3-sync.md` - 문서화 동기화 프로세스
  - `9-feedback.md` - 피드백 폼 변경 없음

**변경 사항**:
- Command 설명에서 "🎩 Alfred" → "🗿 MoAI" 변경
- Hook 경로: `.claude/hooks/alfred/` → `.claude/hooks/moai/` 변경
- Agent 경로: `.claude/agents/alfred/` → `.claude/agents/moai/` 변경

**예상 라인 수**: 50-100 라인 변경

### 1.2 Agents 폴더 생성 및 복사

**위치**: `src/moai_adk/templates/.claude/agents/moai/`

**작업**:
- ✅ 디렉토리 생성: `agents/moai/`
- ✅ Alfred에서 31개 Agent 파일 복사 및 내용 수정:
  - `cc-manager.md` - Claude Code 설정 관리
  - `git-manager.md` - Git 작업 관리
  - `spec-builder.md` - SPEC 문서 생성
  - `tdd-implementer.md` - TDD 구현
  - ... (총 31개 Agent)

**변경 사항**:
- Agent 설명에서 alfred 참조 제거
- Hook 경로: `.claude/hooks/alfred/` → `.claude/hooks/moai/` (모든 참조)
- Command 경로: `.claude/commands/alfred/` → `.claude/commands/moai/` (해당하는 경우)

**예상 라인 수**: 200-300 라인 변경

### 1.3 Hooks 폴더 생성 및 복사

**위치**: `src/moai_adk/templates/.claude/hooks/moai/`

**작업**:
- ✅ 디렉토리 생성: `hooks/moai/`
- ✅ Alfred에서 39개 Hook Python 파일 복사 및 내용 수정:
  - `session_start__*.py` - 세션 시작 훅
  - `pre_tool__*.py` - 도구 실행 전 훅
  - `post_tool__*.py` - 도구 실행 후 훅
  - `on_error__*.py` - 에러 처리 훅
  - ... (총 39개 Hook)

**변경 사항**:
- Import 경로: Alfred 참조 제거
- 환경 변수: ALFRED → MOAI (필요한 경우)
- 로깅 메시지: "Alfred" → "MoAI" 또는 "Moai" 용어로 변경
- 설정 파일 참조: `.moai/config/moai_settings.json` 같은 명확한 네이밍

**예상 라인 수**: 300-400 라인 변경

### 1.4 Settings.json 업데이트

**파일**: `src/moai_adk/templates/.claude/settings.json`

**변경 사항**:
```json
{
  "hooks": [
    {
      "name": "moai",  // 새로운 훅 프로파일
      "enabled": true,
      "config": {
        "hooks": {
          "SessionStart": [
            {
              "type": "command",
              "command": "uv run {{PROJECT_DIR}}/.claude/hooks/moai/session_start__*.py"
            }
          ],
          "PreToolUse": [
            {
              "type": "command",
              "command": "uv run {{PROJECT_DIR}}/.claude/hooks/moai/pre_tool__*.py"
            }
          ]
          // ... 기타 훅 설정
        }
      }
    }
  ]
}
```

**체크리스트**:
- ✅ Alfred 훅 경로를 Moai로 변경
- ✅ 모든 `{{PROJECT_DIR}}/.claude/hooks/alfred/` → `{{PROJECT_DIR}}/.claude/hooks/moai/`
- ✅ Command 참조도 동일하게 업데이트

### 1.5 검증

**검증 항목**:
```bash
# Phase 1 완료 검증
- [ ] commands/moai/ 디렉토리 생성 확인
- [ ] agents/moai/ 디렉토리 생성 확인
- [ ] hooks/moai/ 디렉토리 생성 확인
- [ ] 모든 파일 복사 확인 (commands: 5, agents: 31, hooks: 39)
- [ ] settings.json 훅 경로 업데이트 확인
- [ ] 경로 치환 누락 확인 (alfred → moai)
```

---

## Phase 2: 마이그레이션 로직 구현 (4일)

**목표**: 기존 프로젝트에서 자동으로 Alfred → Moai 마이그레이션을 수행하는 로직 구현

**소요 시간**: 4 작업일

### 2.1 AlfredToMoaiMigrator 클래스 생성

**파일**: `src/moai_adk/core/migration/alfred_to_moai_migrator.py`

**구조**:
```python
class AlfredToMoaiMigrator:
    """Alfred → Moai 폴더 구조 자동 마이그레이션"""

    def __init__(self, project_root: Path):
        """마이그레이터 초기화"""
        self.project_root = project_root
        self.claude_root = project_root / ".claude"
        self.config_path = project_root / ".moai" / "config" / "config.json"

    def needs_migration(self) -> bool:
        """마이그레이션 필요 여부 확인"""
        # alfred 폴더 존재 확인
        # config.json에서 마이그레이션 상태 확인
        # 이미 migrated=true 면 False 반환

    def execute_migration(self, backup_path: Path) -> bool:
        """마이그레이션 실행"""
        # 백업 완료 확인
        # Alfred 폴더 감지 및 기록
        # Moai 템플릿 설치 (package에서)
        # settings.json 훅 경로 업데이트
        # config.json 마이그레이션 상태 기록
        # 검증 수행

    def _delete_alfred_folders(self) -> None:
        """Alfred 폴더 삭제"""
        # .claude/commands/alfred/ 삭제
        # .claude/agents/alfred/ 삭제
        # .claude/hooks/alfred/ 삭제

    def _update_settings_json_hooks(self) -> None:
        """settings.json Hook 경로 업데이트"""
        # 파일 로드
        # alfred → moai 경로 치환
        # 파일 저장

    def _verify_migration(self) -> bool:
        """마이그레이션 검증"""
        # moai 폴더 존재 확인
        # settings.json 경로 업데이트 확인
        # config.json 상태 기록 확인
```

**예상 라인 수**: 150-200 라인

### 2.2 TemplateProcessor에 moai_folders 추가

**파일**: `src/moai_adk/core/template/processor.py`

**변경 사항**:
```python
class TemplateProcessor:
    # 기존 코드
    alfred_folders = [
        ".claude/commands/alfred",
        ".claude/agents/alfred",
        ".claude/hooks/alfred",
    ]

    # 새로운 코드 추가
    moai_folders = [
        ".claude/commands/moai",
        ".claude/agents/moai",
        ".claude/hooks/moai",
    ]

    def _copy_claude(self, target_dirs: list[str] | None = None) -> None:
        """템플릿 복사 (moai 폴더 포함)"""
        # moai_folders도 처리
```

**예상 라인 수**: 20-30 라인 변경

### 2.3 update.py에 마이그레이션 로직 통합

**파일**: `src/moai_adk/cli/commands/update.py`

**변경 위치**: `_sync_templates()` 함수 내부 (Stage 1.5)

**변경 코드**:
```python
def _sync_templates(self) -> None:
    """템플릿 동기화"""

    # Stage 1: 버전 확인 (기존)
    # ...

    # Stage 1.5: Alfred → Moai 마이그레이션 (NEW)
    migrator = AlfredToMoaiMigrator(self.project_root)
    if migrator.needs_migration():
        try:
            if not migrator.execute_migration(backup_path):
                self.logger.error("Alfred → Moai 마이그레이션 실패")
                return False
            self.logger.info("✅ Alfred → Moai 마이그레이션 완료")
        except Exception as e:
            self.logger.error(f"마이그레이션 중 에러: {e}")
            # 롤백 수행
            backup.restore_backup(backup_path)
            return False

    # Stage 2: 템플릿 동기화 (기존, moai 폴더 포함)
    # ...
```

**예상 라인 수**: 20-30 라인 추가

### 2.4 테스트 스위트 생성

**파일**: `tests/unit/test_alfred_to_moai_migrator.py`

**테스트 항목**:
```python
class TestAlfredToMoaiMigrator:
    def test_needs_migration_with_alfred_folders(self):
        """Alfred 폴더가 있으면 마이그레이션 필요"""

    def test_needs_migration_already_migrated(self):
        """이미 마이그레이션되면 불필요"""

    def test_execute_migration_success(self):
        """마이그레이션 성공"""

    def test_execute_migration_backup_fail(self):
        """백업 실패 시 중단"""

    def test_execute_migration_install_fail_rollback(self):
        """설치 실패 시 롤백"""

    def test_update_settings_json_paths(self):
        """settings.json 경로 업데이트"""

    def test_config_json_state_recording(self):
        """config.json 상태 기록"""
```

**예상 라인 수**: 200-250 라인

### 2.5 통합 테스트

**파일**: `tests/integration/test_update_with_migration.py`

**테스트 시나리오**:
```python
class TestUpdateWithMigration:
    def test_update_command_with_alfred_migration(self):
        """update 명령어가 마이그레이션 수행"""

    def test_end_to_end_migration_flow(self):
        """전체 마이그레이션 플로우"""
```

**예상 라인 수**: 100-150 라인

### 2.6 검증

**검증 항목**:
```bash
# Phase 2 완료 검증
- [ ] AlfredToMoaiMigrator 클래스 생성 확인
- [ ] update.py에 마이그레이션 로직 통합 확인
- [ ] TemplateProcessor moai_folders 추가 확인
- [ ] 모든 테스트 통과 확인
- [ ] 수동 테스트: Alfred 폴더 있는 프로젝트에서 마이그레이션 확인
```

---

## Phase 3: 문서화 및 검증 (2일)

**목표**: 마이그레이션 프로세스 문서화 및 comprehensive 테스트

**소요 시간**: 2 작업일

### 3.1 MIGRATION_GUIDE.md 작성

**파일**: `docs/migration/MIGRATION_GUIDE.md`

**내용**:
```markdown
# Alfred → Moai 폴더 구조 마이그레이션 가이드

## 개요
v0.27.0부터 MoAI-ADK는 폴더 구조를 Alfred → Moai로 변경합니다.

## 자동 마이그레이션
`moai-adk update` 명령어 실행 시 자동으로 마이그레이션됩니다.

## 마이그레이션 과정
1. 자동 백업 생성
2. Moai 템플릿 설치
3. Alfred 폴더 삭제
4. settings.json 훅 경로 업데이트
5. config.json 상태 기록

## 마이그레이션 실패 시
자동으로 백업에서 복원됩니다.

## 수동 롤백
`moai-adk rollback <backup-path>`
```

**예상 라인 수**: 100-150 라인

### 3.2 CHANGELOG.md 업데이트

**파일**: `CHANGELOG.md`

**변경 사항**:
```markdown
## v0.27.0 (2025-12-15)

### ✨ Features
- **Auto Migration**: Alfred → Moai 폴더 구조 자동 마이그레이션
  - `moai-adk update` 시 자동으로 기존 Alfred 폴더 마이그레이션
  - 실패 시 자동 롤백
  - `config.json`에 마이그레이션 상태 기록

### 📦 Package Changes
- New: `.claude/commands/moai/`, `.claude/agents/moai/`, `.claude/hooks/moai/`
- Deprecated: `.claude/commands/alfred/`, `.claude/agents/alfred/`, `.claude/hooks/alfred/` (v0.28.0에서 제거 예정)

### 📝 Documentation
- Added: MIGRATION_GUIDE.md - 마이그레이션 프로세스 상세 가이드

### 🚀 Breaking Changes
- None (역호환성 유지)
```

### 3.3 종합 테스트 실행

**테스트 항목**:

**Unit Tests**:
```bash
uv run pytest tests/unit/test_alfred_to_moai_migrator.py -v
# 예상 결과: All 8+ tests passed
```

**Integration Tests**:
```bash
uv run pytest tests/integration/test_update_with_migration.py -v
# 예상 결과: All 3+ tests passed
```

**Manual E2E Test**:
```bash
# 테스트 프로젝트 생성
mkdir test-migration-project
cd test-migration-project
moai-adk init

# Alfred 폴더 생성 (old version 시뮬레이션)
mkdir -p .claude/{commands,agents,hooks}/alfred
touch .claude/commands/alfred/test.md

# 마이그레이션 실행
moai-adk update

# 검증
ls -la .claude/commands/moai/  # Moai 폴더 확인
ls -la .claude/commands/alfred/ # Alfred 폴더 삭제 확인
grep -r "moai" .claude/settings.json # 경로 변경 확인
```

### 3.4 성능 테스트

**테스트 항목**:
```bash
# 마이그레이션 소요 시간 측정
time moai-adk update

# 예상: < 30초 (백업 포함)
```

### 3.5 검증

**검증 항목**:
```bash
# Phase 3 완료 검증
- [ ] MIGRATION_GUIDE.md 작성 완료
- [ ] CHANGELOG.md 업데이트 완료
- [ ] 모든 Unit 테스트 통과 (8+ tests)
- [ ] 모든 Integration 테스트 통과 (3+ tests)
- [ ] E2E 수동 테스트 완료
- [ ] 성능 테스트 완료 (< 30초)
- [ ] 문서 검수 완료
```

---

## Phase 4: 레거시 정리 v0.28.0 (1일)

**목표**: Alfred 폴더 구조 완전 삭제

**타겟 버전**: v0.28.0 (Future)

**소요 시간**: 1 작업일

### 4.1 Package 템플릿 정리

**삭제 대상**:
- `src/moai_adk/templates/.claude/commands/alfred/` (완전 삭제)
- `src/moai_adk/templates/.claude/agents/alfred/` (완전 삭제)
- `src/moai_adk/templates/.claude/hooks/alfred/` (완전 삭제)

### 4.2 TemplateProcessor 업데이트

**파일**: `src/moai_adk/core/template/processor.py`

**변경 사항**:
```python
class TemplateProcessor:
    # Alfred 폴더 제거
    # alfred_folders 삭제

    # Moai 폴더로 변경
    moai_folders = [
        ".claude/commands/moai",
        ".claude/agents/moai",
        ".claude/hooks/moai",
    ]
    # 기존 self.alfred_folders → self.moai_folders 참조 변경
```

### 4.3 update.py 정리

**파일**: `src/moai_adk/cli/commands/update.py`

**변경 사항**:
```python
def _sync_templates(self) -> None:
    # AlfredToMoaiMigrator 코드 완전 제거
    # 마이그레이션 로직이 더 이상 필요 없음 (v0.27.0+ 사용자는 이미 마이그레이션됨)

    # 기존 moai 템플릿 처리만 유지
    self._copy_claude()  # moai 폴더만 처리
```

### 4.4 마이그레이션 로직 제거

**삭제 대상**:
- `src/moai_adk/core/migration/alfred_to_moai_migrator.py` (완전 삭제)
- `tests/unit/test_alfred_to_moai_migrator.py` (완전 삭제)
- `tests/integration/test_update_with_migration.py` (완전 삭제)

### 4.5 검증

**검증 항목**:
```bash
# Phase 4 완료 검증
- [ ] Alfred 폴더 완전 삭제 확인
- [ ] TemplateProcessor 업데이트 확인
- [ ] update.py 정리 확인
- [ ] 마이그레이션 로직 파일 삭제 확인
- [ ] 모든 테스트 통과 확인
```

---

## 의존성 및 위험 관리

### 의존성

| Phase | 선행 Phase | 설명 |
|-------|-----------|------|
| 1 → 2 | Phase 1 완료 | Moai 템플릿 구조가 있어야 마이그레이션 로직 구현 가능 |
| 2 → 3 | Phase 2 완료 | 구현 완료 후 테스트 및 문서화 |
| 3 → 4 | Phase 3 검증 완료 | v0.28.0 릴리스 타겍 (최소 2개월 후) |

### 병렬 작업 가능

**Phase 1과 2는 병렬 진행 가능**:
- Phase 1.1-1.3: 파일 복사 작업 (누군가 담당)
- Phase 2.1-2.4: 마이그레이션 로직 구현 (다른 사람 담당)
- 동기화 지점: Phase 1 완료 후 Phase 2.2 (TemplateProcessor 업데이트)

### 위험 요인 및 대응

| 위험 | 영향 | 대응 |
|------|------|------|
| 백업 실패 | 전체 마이그레이션 중단 | BackupManager 검증 + 에러 처리 |
| 부분 복사 | 불완전한 마이그레이션 | 모든 파일 존재 확인 검증 |
| 경로 치환 오류 | Hook 실행 실패 | settings.json 검증 + 정규식 테스트 |
| Rollback 실패 | 프로젝트 손상 | 자동 재시도 + 사용자 알림 |

---

## 성공 기준

**모든 Phase 완료 후**:

- ✅ 기존 Alfred 폴더가 있는 프로젝트가 `moai-adk update` 시 자동 마이그레이션
- ✅ 마이그레이션 실패 시 자동 롤백
- ✅ 모든 테스트 통과 (Unit + Integration + E2E)
- ✅ 성능: 마이그레이션 < 30초
- ✅ 중복 실행 방지: config.json 상태 확인
- ✅ 사용자 피드백: 진행 상황 콘솔 메시지 출력

---

## End of Plan
