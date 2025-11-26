# SPEC-UPDATE-001: 수락 기준 및 테스트 시나리오

> **TAG**: ACCEPT-UPDATE-001
> **연관 SPEC**: SPEC-UPDATE-001
> **작성일**: 2025-11-27

---

## 📋 수락 기준 개요

이 문서는 SPEC-UPDATE-001의 구현이 완료되었는지 검증하기 위한 상세한 수락 기준과 테스트 시나리오를 정의합니다. 모든 시나리오는 Given-When-Then 형식으로 작성되었습니다.

---

## 🎯 전체 수락 기준

### 기능적 수락 기준

- ✅ **AC-1**: Commands, Agents, Hooks 파일이 자동으로 탐지되어야 한다
- ✅ **AC-2**: 탐지된 사용자 정의 파일이 `.moai/backups/update/` 하위에 백업되어야 한다
- ✅ **AC-3**: 업데이트 후 백업된 파일이 원래 위치로 복원되어야 한다
- ✅ **AC-4**: 멀티셀렉트 UI를 통해 복원할 파일 유형을 선택할 수 있어야 한다
- ✅ **AC-5**: --yes 플래그 사용 시 모든 백업 파일이 자동으로 복원되어야 한다
- ✅ **AC-6**: 템플릿 파일과 충돌하는 파일은 복원되지 않아야 한다

### 비기능적 수락 기준

- ✅ **AC-7**: 테스트 커버리지가 85% 이상이어야 한다
- ✅ **AC-8**: 모든 함수에 type hints가 적용되어야 한다
- ✅ **AC-9**: 모든 함수에 docstring이 작성되어야 한다
- ✅ **AC-10**: Ruff 린터를 통과해야 한다
- ✅ **AC-11**: mypy 타입 검사를 통과해야 한다
- ✅ **AC-12**: 실행 시간이 1초 이내여야 한다 (파일 수십 개 기준)

---

## 🧪 테스트 시나리오

### 시나리오 1: Commands 백업 및 복원

**목적**: 사용자 정의 Commands 파일이 정확히 백업되고 복원되는지 검증

**Given** (사전 조건):
```
.claude/commands/moai/
├── 1-plan.md          (템플릿 파일)
└── custom-cmd.md      (사용자 정의 파일)
```

**When** (실행):
```bash
moai-adk update
# UI에서 "Commands" 선택
```

**Then** (예상 결과):
```
1. custom-cmd.md가 .moai/backups/update/commands/에 백업됨
2. 업데이트 후 custom-cmd.md가 .claude/commands/moai/에 복원됨
3. 1-plan.md는 백업/복원 대상이 아님
4. 콘솔 출력:
   [green]✓[/green] Backed up: custom-cmd.md
   [green]✓[/green] Restored: custom-cmd.md
```

**검증 방법**:
```python
def test_commands_backup_and_restore():
    # Given
    custom_file = Path(".claude/commands/moai/custom-cmd.md")
    custom_file.write_text("# Custom Command")
    template_file = Path(".claude/commands/moai/1-plan.md")
    template_file.write_text("# Template Command")

    # When
    custom_commands = detect_custom_commands()
    backup_files(custom_commands, "commands")
    restore_files(["commands"])

    # Then
    backup_path = Path(".moai/backups/update/commands/custom-cmd.md")
    assert backup_path.exists()
    assert custom_file.exists()
    assert custom_file.read_text() == "# Custom Command"
    assert template_file.read_text() == "# Template Command"
```

**성공 기준**:
- ✅ 백업 파일이 정확한 경로에 생성됨
- ✅ 복원된 파일 내용이 원본과 동일함
- ✅ 템플릿 파일이 백업/복원되지 않음

---

### 시나리오 2: Agents 백업 및 복원

**목적**: 사용자 정의 Agents 파일이 정확히 백업되고 복원되는지 검증

**Given** (사전 조건):
```
.claude/agents/
├── manager-spec.md        (템플릿 파일)
└── custom-agent.md        (사용자 정의 파일)
```

**When** (실행):
```bash
moai-adk update
# UI에서 "Agents" 선택
```

**Then** (예상 결과):
```
1. custom-agent.md가 .moai/backups/update/agents/에 백업됨
2. 업데이트 후 custom-agent.md가 .claude/agents/에 복원됨
3. manager-spec.md는 백업/복원 대상이 아님
4. 콘솔 출력:
   [green]✓[/green] Backed up: custom-agent.md
   [green]✓[/green] Restored: custom-agent.md
```

**검증 방법**:
```python
def test_agents_backup_and_restore():
    # Given
    custom_agent = Path(".claude/agents/custom-agent.md")
    custom_agent.write_text("# Custom Agent")
    template_agent = Path(".claude/agents/manager-spec.md")
    template_agent.write_text("# Template Agent")

    # When
    custom_agents = detect_custom_agents()
    backup_files(custom_agents, "agents")
    restore_files(["agents"])

    # Then
    backup_path = Path(".moai/backups/update/agents/custom-agent.md")
    assert backup_path.exists()
    assert custom_agent.exists()
    assert custom_agent.read_text() == "# Custom Agent"
    assert template_agent.read_text() == "# Template Agent"
```

**성공 기준**:
- ✅ 백업 파일이 정확한 경로에 생성됨
- ✅ 복원된 파일 내용이 원본과 동일함
- ✅ 템플릿 파일이 백업/복원되지 않음

---

### 시나리오 3: Hooks 백업 및 복원

**목적**: 사용자 정의 Hooks 파일이 정확히 백업되고 복원되는지 검증

**Given** (사전 조건):
```
.claude/hooks/moai/
├── session_start__show_project_info.py   (템플릿 파일)
└── custom_hook.py                        (사용자 정의 파일)
```

**When** (실행):
```bash
moai-adk update
# UI에서 "Hooks" 선택
```

**Then** (예상 결과):
```
1. custom_hook.py가 .moai/backups/update/hooks/에 백업됨
2. 업데이트 후 custom_hook.py가 .claude/hooks/moai/에 복원됨
3. session_start__show_project_info.py는 백업/복원 대상이 아님
4. 콘솔 출력:
   [green]✓[/green] Backed up: custom_hook.py
   [green]✓[/green] Restored: custom_hook.py
```

**검증 방법**:
```python
def test_hooks_backup_and_restore():
    # Given
    custom_hook = Path(".claude/hooks/moai/custom_hook.py")
    custom_hook.write_text("# Custom Hook")
    template_hook = Path(".claude/hooks/moai/session_start__show_project_info.py")
    template_hook.write_text("# Template Hook")

    # When
    custom_hooks = detect_custom_hooks()
    backup_files(custom_hooks, "hooks")
    restore_files(["hooks"])

    # Then
    backup_path = Path(".moai/backups/update/hooks/custom_hook.py")
    assert backup_path.exists()
    assert custom_hook.exists()
    assert custom_hook.read_text() == "# Custom Hook"
    assert template_hook.read_text() == "# Template Hook"
```

**성공 기준**:
- ✅ 백업 파일이 정확한 경로에 생성됨
- ✅ 복원된 파일 내용이 원본과 동일함
- ✅ 템플릿 파일이 백업/복원되지 않음

---

### 시나리오 4: 여러 파일 유형 동시 백업 및 복원

**목적**: 여러 파일 유형을 동시에 백업하고 복원할 수 있는지 검증

**Given** (사전 조건):
```
.claude/
├── commands/moai/custom-cmd.md        (사용자 정의)
├── agents/custom-agent.md             (사용자 정의)
└── hooks/moai/custom-hook.py          (사용자 정의)
```

**When** (실행):
```bash
moai-adk update
# UI에서 "Commands", "Agents", "Hooks" 모두 선택 (멀티셀렉트)
```

**Then** (예상 결과):
```
1. 3개 파일이 모두 백업됨:
   - .moai/backups/update/commands/custom-cmd.md
   - .moai/backups/update/agents/custom-agent.md
   - .moai/backups/update/hooks/custom-hook.py

2. 업데이트 후 3개 파일이 모두 복원됨

3. 콘솔 출력:
   [green]✓[/green] Backed up: custom-cmd.md
   [green]✓[/green] Backed up: custom-agent.md
   [green]✓[/green] Backed up: custom-hook.py
   [green]✓[/green] Restored: custom-cmd.md
   [green]✓[/green] Restored: custom-agent.md
   [green]✓[/green] Restored: custom-hook.py
   [green]✓[/green] Restoration complete: 3 files restored, 0 skipped.
```

**검증 방법**:
```python
def test_multiple_file_types_backup_and_restore():
    # Given
    custom_files = {
        "commands": Path(".claude/commands/moai/custom-cmd.md"),
        "agents": Path(".claude/agents/custom-agent.md"),
        "hooks": Path(".claude/hooks/moai/custom-hook.py")
    }
    for file_type, file_path in custom_files.items():
        file_path.parent.mkdir(parents=True, exist_ok=True)
        file_path.write_text(f"# Custom {file_type}")

    # When
    all_custom = {
        "commands": detect_custom_commands(),
        "agents": detect_custom_agents(),
        "hooks": detect_custom_hooks()
    }
    for file_type, files in all_custom.items():
        backup_files(files, file_type)

    results = restore_files(["commands", "agents", "hooks"])

    # Then
    assert results["restored"] == 3
    assert results["skipped"] == 0
    for file_path in custom_files.values():
        assert file_path.exists()
```

**성공 기준**:
- ✅ 3개 파일이 모두 백업됨
- ✅ 3개 파일이 모두 복원됨
- ✅ 복원 결과가 정확히 보고됨 (3 restored, 0 skipped)

---

### 시나리오 5: --yes 플래그로 자동 복원

**목적**: --yes 플래그 사용 시 사용자 입력 없이 자동으로 복원되는지 검증

**Given** (사전 조건):
```
.moai/backups/update/
├── commands/custom-cmd.md
├── agents/custom-agent.md
└── hooks/custom-hook.py
```

**When** (실행):
```bash
moai-adk update --yes
```

**Then** (예상 결과):
```
1. questionary 프롬프트가 표시되지 않음
2. 모든 백업 파일이 자동으로 복원됨
3. 콘솔 출력:
   [green]✓[/green] Restored: custom-cmd.md
   [green]✓[/green] Restored: custom-agent.md
   [green]✓[/green] Restored: custom-hook.py
   [green]✓[/green] Restoration complete: 3 files restored, 0 skipped.
```

**검증 방법**:
```python
def test_yes_flag_auto_restore(monkeypatch):
    # Given
    backup_dir = Path(".moai/backups/update")
    (backup_dir / "commands").mkdir(parents=True, exist_ok=True)
    (backup_dir / "commands/custom-cmd.md").write_text("# Custom")

    # Mock questionary to ensure it's not called
    questionary_called = False
    def mock_questionary(*args, **kwargs):
        nonlocal questionary_called
        questionary_called = True

    monkeypatch.setattr("questionary.checkbox", mock_questionary)

    # When
    _handle_restoration(yes=True)

    # Then
    assert not questionary_called
    assert Path(".claude/commands/moai/custom-cmd.md").exists()
```

**성공 기준**:
- ✅ questionary 프롬프트가 호출되지 않음
- ✅ 모든 백업 파일이 자동 복원됨
- ✅ CI/CD 환경에서 사용 가능

---

### 시나리오 6: 템플릿 파일과 충돌 시 복원 스킵

**목적**: 템플릿 파일과 동일한 이름의 백업 파일은 복원되지 않는지 검증

**Given** (사전 조건):
```
.moai/backups/update/commands/
├── custom-cmd.md        (사용자 정의)
└── 1-plan.md            (템플릿 파일, 충돌)

src/moai_adk/templates/.claude/commands/moai/
└── 1-plan.md            (템플릿 파일)
```

**When** (실행):
```bash
moai-adk update
# UI에서 "Commands" 선택
```

**Then** (예상 결과):
```
1. custom-cmd.md만 복원됨
2. 1-plan.md는 복원 스킵됨
3. 콘솔 출력:
   [green]✓[/green] Restored: custom-cmd.md
   [yellow]⚠[/yellow] Skipping 1-plan.md: conflicts with template
   [green]✓[/green] Restoration complete: 1 files restored, 1 skipped.
```

**검증 방법**:
```python
def test_template_conflict_skip():
    # Given
    backup_dir = Path(".moai/backups/update/commands")
    backup_dir.mkdir(parents=True, exist_ok=True)
    (backup_dir / "custom-cmd.md").write_text("# Custom")
    (backup_dir / "1-plan.md").write_text("# Backup of template")

    template_file = Path(".claude/commands/moai/1-plan.md")
    template_file.write_text("# Updated Template")

    # When
    results = restore_files(["commands"])

    # Then
    assert results["restored"] == 1
    assert results["skipped"] == 1
    assert Path(".claude/commands/moai/custom-cmd.md").exists()
    # Template file should remain unchanged
    assert template_file.read_text() == "# Updated Template"
```

**성공 기준**:
- ✅ 사용자 정의 파일만 복원됨
- ✅ 템플릿 파일은 복원 스킵됨
- ✅ 경고 메시지가 출력됨
- ✅ 템플릿 파일이 덮어써지지 않음

---

### 시나리오 7: 백업 파일이 없을 때 처리

**목적**: 백업 파일이 없을 때 복원 단계를 건너뛰는지 검증

**Given** (사전 조건):
```
.claude/commands/moai/
└── 1-plan.md   (템플릿 파일만 존재, 사용자 정의 파일 없음)
```

**When** (실행):
```bash
moai-adk update
```

**Then** (예상 결과):
```
1. detection 단계에서 사용자 정의 파일 0개 탐지
2. 백업 단계 스킵
3. questionary 프롬프트 표시 안 됨
4. 콘솔 출력:
   [yellow]No custom files detected. Skipping backup.[/yellow]
   [yellow]No backups found. Skipping restoration.[/yellow]
```

**검증 방법**:
```python
def test_no_backups_skip_restoration(monkeypatch):
    # Given: 사용자 정의 파일 없음
    custom_commands = detect_custom_commands()
    assert len(custom_commands) == 0

    # Mock questionary to ensure it's not called
    questionary_called = False
    def mock_questionary(*args, **kwargs):
        nonlocal questionary_called
        questionary_called = True

    monkeypatch.setattr("questionary.checkbox", mock_questionary)

    # When
    _handle_restoration(yes=False)

    # Then
    assert not questionary_called
    # Restoration should be skipped
```

**성공 기준**:
- ✅ 백업 단계가 스킵됨
- ✅ questionary 프롬프트가 표시되지 않음
- ✅ 복원 단계가 스킵됨
- ✅ 정보 메시지가 출력됨

---

## 🔍 품질 게이트

### 코드 품질 검증

**QG-1: 테스트 커버리지**
```bash
pytest tests/test_cli/test_update*.py --cov=src/moai_adk/cli/commands/update --cov-report=term-missing
# 기준: ≥85%
```

**검증 명령**:
```bash
# 커버리지 확인
coverage run -m pytest tests/test_cli/
coverage report --include="*/update.py"
coverage html  # HTML 리포트 생성

# 결과 예시:
Name                                   Stmts   Miss  Cover
------------------------------------------------------------
src/moai_adk/cli/commands/update.py      200      5    97%
------------------------------------------------------------
TOTAL                                    200      5    97%
```

**수락 조건**: Coverage ≥ 85%

---

**QG-2: 타입 힌트 검증**
```bash
mypy src/moai_adk/cli/commands/update.py --strict
# 기준: 0 errors
```

**수락 조건**: mypy 에러 0개

---

**QG-3: 린팅 검증**
```bash
ruff check src/moai_adk/cli/commands/update.py
# 기준: 0 errors, 0 warnings
```

**수락 조건**: ruff 에러 및 경고 0개

---

**QG-4: Docstring 검증**
```bash
# 모든 함수에 docstring 존재 확인
python -c "
import ast
import inspect
import src.moai_adk.cli.commands.update as module

for name, obj in inspect.getmembers(module, inspect.isfunction):
    if not obj.__doc__:
        print(f'Missing docstring: {name}')
"
# 기준: 0 missing docstrings
```

**수락 조건**: 모든 함수에 docstring 존재

---

### 성능 검증

**QG-5: 실행 시간**
```bash
time moai-adk update --yes
# 기준: ≤1초 (파일 수십 개)
```

**수락 조건**: 실행 시간 ≤ 1초

---

### 기능 검증

**QG-6: E2E 테스트 통과**
```bash
pytest tests/test_cli/test_update_e2e.py -v
# 기준: 모든 시나리오 통과
```

**수락 조건**: 7개 시나리오 모두 통과 (시나리오 1-7)

---

## 📊 테스트 실행 계획

### 1단계: 단위 테스트 (Unit Tests)

**파일**: `tests/test_cli/test_update_detection.py`

**테스트 케이스**:
- `test_detect_custom_commands_success()`
- `test_detect_custom_commands_empty()`
- `test_detect_custom_agents_success()`
- `test_detect_custom_hooks_success()`
- `test_get_template_files()`
- `test_is_template_file()`

**실행**:
```bash
pytest tests/test_cli/test_update_detection.py -v
```

---

### 2단계: 통합 테스트 (Integration Tests)

**파일**: `tests/test_cli/test_update_backup.py`, `tests/test_cli/test_update_restore.py`

**테스트 케이스**:
- `test_backup_files_success()`
- `test_backup_files_no_directory()`
- `test_restore_files_success()`
- `test_restore_files_conflict_skip()`

**실행**:
```bash
pytest tests/test_cli/test_update_backup.py tests/test_cli/test_update_restore.py -v
```

---

### 3단계: UI 테스트

**파일**: `tests/test_cli/test_update_ui.py`

**테스트 케이스**:
- `test_show_restoration_prompt_with_backups()`
- `test_show_restoration_prompt_no_backups()`
- `test_has_backups()`

**실행**:
```bash
pytest tests/test_cli/test_update_ui.py -v
```

---

### 4단계: E2E 테스트

**파일**: `tests/test_cli/test_update_e2e.py`

**테스트 케이스**: 시나리오 1-7

**실행**:
```bash
pytest tests/test_cli/test_update_e2e.py -v
```

---

## ✅ Definition of Done (완료 기준)

### 코드 완료 기준

- [x] 모든 함수에 type hints 적용
- [x] 모든 함수에 docstring 작성
- [x] detect_custom_commands() 구현 완료
- [x] detect_custom_agents() 구현 완료
- [x] detect_custom_hooks() 구현 완료
- [x] backup_files() 구현 완료
- [x] restore_files() 구현 완료
- [x] show_restoration_prompt() 구현 완료
- [x] is_template_file() 구현 완료

### 테스트 완료 기준

- [x] 단위 테스트 작성 완료 (≥10 테스트)
- [x] 통합 테스트 작성 완료 (≥5 테스트)
- [x] E2E 테스트 작성 완료 (7 시나리오)
- [x] 테스트 커버리지 ≥85%
- [x] 모든 테스트 통과 (0 failures)

### 품질 게이트 통과

- [x] `pytest --cov` ≥85%
- [x] `mypy --strict` 0 errors
- [x] `ruff check` 0 errors/warnings
- [x] Docstring coverage 100%
- [x] 실행 시간 ≤1초

### 문서화 완료

- [x] spec.md 작성 완료
- [x] plan.md 작성 완료
- [x] acceptance.md 작성 완료 (이 문서)
- [x] README 업데이트 (새 기능 설명)

---

## 🚀 배포 전 체크리스트

### 코드 리뷰

- [ ] 코드 리뷰 완료 (1명 이상)
- [ ] 리뷰 피드백 반영 완료
- [ ] TRUST 5 원칙 준수 확인

### 테스트 검증

- [ ] 로컬 환경 테스트 통과
- [ ] CI/CD 파이프라인 테스트 통과
- [ ] --yes 플래그 자동화 테스트 통과

### 문서화

- [ ] 변경 사항 CHANGELOG에 기록
- [ ] API 문서 업데이트
- [ ] 사용자 가이드 업데이트

### 배포 준비

- [ ] 버전 번호 업데이트
- [ ] 릴리스 노트 작성
- [ ] 배포 승인 완료

---

**END OF ACCEPTANCE CRITERIA**
