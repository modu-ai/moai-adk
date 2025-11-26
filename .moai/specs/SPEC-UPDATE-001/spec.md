---
id: SPEC-UPDATE-001
version: "1.0.0"
status: "draft"
created: "2025-11-27"
updated: "2025-11-27"
author: "GOOS"
priority: "HIGH"
---

## HISTORY

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2025-11-27 | GOOS | Initial draft - moai-adk update 사용자 정의 파일 복원 기능 |

# SPEC-UPDATE-001: moai-adk update 사용자 정의 파일 복원 기능 확장

## 개요

moai-adk update 명령어를 실행할 때, 사용자가 커스터마이징한 Commands, Agents, Hooks 파일들을 자동으로 백업하고 업데이트 후 복원하는 기능을 추가합니다. 기존에는 Skills만 백업/복원이 가능했으나, 이를 확장하여 더 많은 사용자 정의 파일을 보호합니다.

---

## Environment (실행 환경 및 전제 조건)

### 시스템 환경

- **Python**: 3.9 이상
- **필수 패키지**:
  - `questionary>=2.0.0` - 대화형 프롬프트
  - `rich>=13.0.0` - 콘솔 출력 포맷팅
  - `pathlib` - 경로 처리 (표준 라이브러리)
  - `shutil` - 파일 복사 (표준 라이브러리)

### 프로젝트 구조

```
.claude/
├── commands/
│   └── moai/           # 사용자 정의 명령어 (백업 대상)
├── agents/             # 사용자 정의 에이전트 (백업 대상)
├── skills/             # 사용자 정의 스킬 (기존 백업 대상)
└── hooks/
    └── moai/           # 사용자 정의 훅 (백업 대상)

.moai/
└── backups/
    └── update/
        ├── commands/   # 백업 저장 위치
        ├── agents/     # 백업 저장 위치
        ├── skills/     # 기존 백업 저장 위치
        └── hooks/      # 백업 저장 위치
```

### 기존 코드 베이스

- **Target File**: `src/moai_adk/cli/commands/update.py`
- **Current Status**: Skills 백업/복원 기능만 구현됨
- **Extension Point**: `_handle_restoration()` 함수

---

## Assumptions (가정 사항)

### 기술적 가정

1. **파일 시스템 접근 권한**: `.claude/` 및 `.moai/` 디렉토리에 읽기/쓰기 권한이 있다고 가정
2. **파일 충돌 없음**: 백업 파일과 템플릿 파일 간 충돌이 발생하지 않는다고 가정 (detection 로직으로 검증)
3. **사용자 정의 파일 식별**: `.claude/` 하위 파일이 템플릿 패키지에 없으면 사용자 정의로 간주
4. **Questionary UI 가용성**: 터미널 환경에서 대화형 UI가 정상 작동한다고 가정

### 운영 가정

1. **업데이트 중단 시나리오**: 복원 단계 전 중단 시, 백업은 `.moai/backups/update/`에 유지되며 수동 복원 가능
2. **네트워크 독립성**: 백업/복원은 로컬 파일 시스템 작업이므로 네트워크 상태와 무관
3. **멱등성**: 동일한 백업을 여러 번 복원해도 안전 (덮어쓰기 확인)

---

## Requirements (요구사항)

### Ubiquitous Requirements (항상 활성화)

**REQ-UPDATE-001**: 시스템은 update 실행 시 사용자 정의 Commands, Agents, Hooks를 자동으로 탐지해야 한다.
- **조건**: 템플릿 패키지(`src/moai_adk/templates/.claude/`)에 없는 파일
- **출력**: 탐지된 파일 경로 리스트

**REQ-UPDATE-002**: 시스템은 탐지된 사용자 정의 파일을 `.moai/backups/update/` 하위에 백업해야 한다.
- **조건**: 백업 디렉토리 구조는 원본 구조를 유지 (`commands/`, `agents/`, `hooks/`)
- **출력**: 백업 완료 메시지 및 백업 경로

**REQ-UPDATE-003**: 시스템은 백업된 파일을 업데이트 완료 후 원래 위치로 복원해야 한다.
- **조건**: 복원 시 파일 충돌 검사 (템플릿 파일과 동일 경로)
- **출력**: 복원 완료 메시지 및 복원된 파일 리스트

### Event-driven Requirements (이벤트 기반)

**REQ-UPDATE-004**: WHEN 사용자가 `moai-adk update` 명령어를 실행할 때, THEN 시스템은 복원할 파일 유형을 선택하는 멀티셀렉트 프롬프트를 표시해야 한다.
- **이벤트**: `moai-adk update` 실행
- **액션**: questionary checkbox 프롬프트 표시
- **선택지**: "Commands", "Agents", "Hooks", "Skills" (멀티셀렉트)

**REQ-UPDATE-005**: WHEN 사용자가 `--yes` 플래그를 사용할 때, THEN 시스템은 모든 백업 파일을 자동으로 복원해야 한다.
- **이벤트**: `--yes` 플래그 감지
- **액션**: 사용자 입력 없이 전체 복원 실행
- **조건**: 백업 파일이 존재하는 경우만

**REQ-UPDATE-006**: WHEN 복원 중 파일 충돌이 감지될 때 (템플릿에 동일 파일 존재), THEN 시스템은 경고 메시지를 표시하고 복원을 건너뛰어야 한다.
- **이벤트**: 충돌 파일 탐지
- **액션**: 경고 로그 출력 + 해당 파일 복원 스킵
- **출력**: `[WARNING] Skipping restoration of {file_path}: conflicts with template`

### State-driven Requirements (상태 기반)

**REQ-UPDATE-007**: WHILE 백업 디렉토리(`~/.moai/backups/update/`)가 존재하지 않는 경우, THEN 시스템은 백업 디렉토리를 생성해야 한다.
- **상태**: 백업 디렉토리 부재
- **액션**: `Path.mkdir(parents=True, exist_ok=True)`

**REQ-UPDATE-008**: WHILE 백업할 사용자 정의 파일이 없는 경우, THEN 시스템은 백업/복원 단계를 건너뛰고 정보 메시지를 표시해야 한다.
- **상태**: 사용자 정의 파일 0개
- **액션**: 백업/복원 스킵 + `[INFO] No custom files detected. Skipping backup.`

### Unwanted Behaviors (금지사항)

**REQ-UPDATE-009**: 시스템은 템플릿 파일을 절대로 백업해서는 안 된다.
- **금지**: 패키지에 포함된 기본 파일을 백업 대상으로 간주
- **검증**: 파일 경로를 템플릿 패키지와 비교하여 사용자 정의 여부 확인

**REQ-UPDATE-010**: 시스템은 복원 시 기존 템플릿 파일을 덮어써서는 안 된다.
- **금지**: 충돌 파일 강제 복원
- **검증**: 복원 전 템플릿 파일 존재 여부 검사

**REQ-UPDATE-011**: 시스템은 백업 실패 시 업데이트를 중단해서는 안 된다.
- **금지**: 백업 에러 → 업데이트 롤백
- **대응**: 백업 실패 경고 출력 + 업데이트 계속 진행

### Optional Requirements (선택사항)

**REQ-UPDATE-012**: 시스템은 백업 날짜를 기록하여 여러 버전의 백업을 유지할 수 있다 (추후 확장).
- **선택사항**: 타임스탬프 기반 백업 디렉토리 (`backups/update-2025-11-27-143022/`)
- **현재**: 단일 백업만 유지 (덮어쓰기)

**REQ-UPDATE-013**: 시스템은 복원 후 복원 결과를 `.moai/logs/update.log`에 기록할 수 있다.
- **선택사항**: 복원 성공/실패 기록
- **현재**: 콘솔 출력만

---

## Specifications (기술 사양)

### 아키텍처 설계

```
┌─────────────────────────────────────────────────────────┐
│               moai-adk update 워크플로우                  │
└─────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│ 1. Detection Phase (탐지)                                │
│    - detect_custom_commands()                            │
│    - detect_custom_agents()                              │
│    - detect_custom_hooks()                               │
│    - detect_custom_skills() [기존]                       │
└─────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│ 2. Backup Phase (백업)                                   │
│    - backup_files(files, backup_type)                    │
│    - .moai/backups/update/{type}/ 생성                   │
└─────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│ 3. Update Phase (업데이트)                               │
│    - sync_templates() [기존 로직]                        │
└─────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│ 4. Restoration Phase (복원)                              │
│    - show_restoration_prompt() [--yes 플래그 없을 때]    │
│    - restore_files(selected_types)                       │
│    - conflict_detection()                                │
└─────────────────────────────────────────────────────────┘
```

### 주요 함수 명세

#### 1. detect_custom_commands() → List[Path]

**목적**: 사용자 정의 Commands 파일 탐지

**로직**:
```python
def detect_custom_commands() -> list[Path]:
    """
    템플릿에 없는 사용자 정의 Commands를 탐지합니다.

    Returns:
        List[Path]: 사용자 정의 Commands 파일 경로 리스트
    """
    custom_files = []
    commands_dir = Path(".claude/commands/moai")
    template_commands = get_template_files("commands/moai")

    if not commands_dir.exists():
        return custom_files

    for file in commands_dir.glob("*.md"):
        if file.name not in template_commands:
            custom_files.append(file)

    return custom_files
```

**복잡도**: O(n), n = commands 파일 개수
**예상 라인 수**: ~20 lines

#### 2. detect_custom_agents() → List[Path]

**목적**: 사용자 정의 Agents 파일 탐지

**로직**:
```python
def detect_custom_agents() -> list[Path]:
    """
    템플릿에 없는 사용자 정의 Agents를 탐지합니다.

    Returns:
        List[Path]: 사용자 정의 Agents 파일 경로 리스트
    """
    custom_files = []
    agents_dir = Path(".claude/agents")
    template_agents = get_template_files("agents")

    if not agents_dir.exists():
        return custom_files

    for file in agents_dir.glob("*.md"):
        if file.name not in template_agents:
            custom_files.append(file)

    return custom_files
```

**복잡도**: O(n), n = agents 파일 개수
**예상 라인 수**: ~20 lines

#### 3. detect_custom_hooks() → List[Path]

**목적**: 사용자 정의 Hooks 파일 탐지

**로직**:
```python
def detect_custom_hooks() -> list[Path]:
    """
    템플릿에 없는 사용자 정의 Hooks를 탐지합니다.

    Returns:
        List[Path]: 사용자 정의 Hooks 파일 경로 리스트
    """
    custom_files = []
    hooks_dir = Path(".claude/hooks/moai")
    template_hooks = get_template_files("hooks/moai")

    if not hooks_dir.exists():
        return custom_files

    for file in hooks_dir.glob("*.py"):
        if file.name not in template_hooks:
            custom_files.append(file)

    return custom_files
```

**복잡도**: O(n), n = hooks 파일 개수
**예상 라인 수**: ~20 lines

#### 4. backup_files(files: List[Path], backup_type: str) → None

**목적**: 파일을 백업 디렉토리로 복사

**로직**:
```python
def backup_files(files: list[Path], backup_type: str) -> None:
    """
    사용자 정의 파일을 백업 디렉토리로 복사합니다.

    Args:
        files: 백업할 파일 경로 리스트
        backup_type: 백업 유형 ("commands", "agents", "hooks", "skills")
    """
    backup_dir = Path(".moai/backups/update") / backup_type
    backup_dir.mkdir(parents=True, exist_ok=True)

    for file in files:
        target = backup_dir / file.name
        shutil.copy2(file, target)
        console.print(f"[green]✓[/green] Backed up: {file.name}")
```

**복잡도**: O(n), n = 백업 파일 개수
**예상 라인 수**: ~15 lines

#### 5. show_restoration_prompt() → List[str]

**목적**: 복원할 파일 유형을 사용자로부터 선택받기

**로직**:
```python
def show_restoration_prompt() -> list[str]:
    """
    복원할 파일 유형을 사용자에게 선택받습니다.

    Returns:
        List[str]: 선택된 파일 유형 리스트 (예: ["commands", "agents"])
    """
    choices = []
    if has_backups("commands"):
        choices.append(questionary.Choice("Commands", value="commands"))
    if has_backups("agents"):
        choices.append(questionary.Choice("Agents", value="agents"))
    if has_backups("hooks"):
        choices.append(questionary.Choice("Hooks", value="hooks"))
    if has_backups("skills"):
        choices.append(questionary.Choice("Skills", value="skills"))

    if not choices:
        console.print("[yellow]No backups found. Skipping restoration.[/yellow]")
        return []

    selected = questionary.checkbox(
        "Select file types to restore:",
        choices=choices
    ).ask()

    return selected or []
```

**복잡도**: O(1) - 최대 4개 선택지
**예상 라인 수**: ~25 lines

#### 6. restore_files(selected_types: List[str]) → Dict[str, int]

**목적**: 선택된 파일 유형을 복원하고 결과 반환

**로직**:
```python
def restore_files(selected_types: list[str]) -> dict[str, int]:
    """
    백업된 파일을 원래 위치로 복원합니다.

    Args:
        selected_types: 복원할 파일 유형 리스트

    Returns:
        Dict[str, int]: 복원 결과 {"restored": 3, "skipped": 1}
    """
    results = {"restored": 0, "skipped": 0}

    for backup_type in selected_types:
        backup_dir = Path(".moai/backups/update") / backup_type
        target_dir_map = {
            "commands": Path(".claude/commands/moai"),
            "agents": Path(".claude/agents"),
            "hooks": Path(".claude/hooks/moai"),
            "skills": Path(".claude/skills")
        }
        target_dir = target_dir_map[backup_type]

        for backup_file in backup_dir.glob("*"):
            target_file = target_dir / backup_file.name

            # 충돌 검사: 템플릿 파일과 동일 경로인지 확인
            if is_template_file(backup_type, backup_file.name):
                console.print(f"[yellow]⚠[/yellow] Skipping {backup_file.name}: conflicts with template")
                results["skipped"] += 1
                continue

            shutil.copy2(backup_file, target_file)
            console.print(f"[green]✓[/green] Restored: {backup_file.name}")
            results["restored"] += 1

    return results
```

**복잡도**: O(n*m), n = 파일 유형 개수, m = 파일 개수
**예상 라인 수**: ~35 lines

#### 7. is_template_file(file_type: str, file_name: str) → bool

**목적**: 파일이 템플릿에 포함되어 있는지 확인

**로직**:
```python
def is_template_file(file_type: str, file_name: str) -> bool:
    """
    파일이 템플릿 패키지에 존재하는지 확인합니다.

    Args:
        file_type: 파일 유형 ("commands", "agents", "hooks", "skills")
        file_name: 파일명

    Returns:
        bool: 템플릿 파일이면 True, 사용자 정의 파일이면 False
    """
    template_base = Path(__file__).parent.parent / "templates/.claude"
    type_map = {
        "commands": template_base / "commands/moai",
        "agents": template_base / "agents",
        "hooks": template_base / "hooks/moai",
        "skills": template_base / "skills"
    }
    template_dir = type_map.get(file_type)
    if not template_dir:
        return False

    return (template_dir / file_name).exists()
```

**복잡도**: O(1) - 파일 시스템 조회
**예상 라인 수**: ~20 lines

### 통합 워크플로우

```python
def _handle_restoration(yes: bool = False) -> None:
    """
    백업 파일 복원 처리 (확장된 버전)

    Args:
        yes: --yes 플래그 여부 (자동 복원)
    """
    # 1. Detection Phase
    custom_commands = detect_custom_commands()
    custom_agents = detect_custom_agents()
    custom_hooks = detect_custom_hooks()
    custom_skills = detect_custom_skills()  # 기존 함수

    all_custom_files = {
        "commands": custom_commands,
        "agents": custom_agents,
        "hooks": custom_hooks,
        "skills": custom_skills
    }

    # 2. Backup Phase
    for file_type, files in all_custom_files.items():
        if files:
            backup_files(files, file_type)

    # 3. Update Phase (sync_templates 호출, 기존 로직)
    sync_templates()

    # 4. Restoration Phase
    if yes:
        # --yes 플래그: 모든 백업 자동 복원
        selected_types = [k for k, v in all_custom_files.items() if v]
    else:
        # 대화형 프롬프트
        selected_types = show_restoration_prompt()

    if selected_types:
        results = restore_files(selected_types)
        console.print(f"\n[green]✓[/green] Restoration complete: {results['restored']} files restored, {results['skipped']} skipped.")
    else:
        console.print("[yellow]No files selected for restoration.[/yellow]")
```

**예상 라인 수**: ~50 lines

### 에러 핸들링

| 에러 시나리오 | 처리 방법 |
|--------------|----------|
| 백업 디렉토리 생성 실패 | `Path.mkdir()` 예외 처리 → 경고 출력 + 계속 진행 |
| 파일 복사 실패 (권한 없음) | `shutil.copy2()` 예외 처리 → 해당 파일 스킵 + 경고 |
| 템플릿 파일과 충돌 | `is_template_file()` 검사 → 복원 스킵 + 경고 |
| 백업 파일 없음 | 복원 프롬프트 표시 안 함 + 정보 메시지 |

### 성능 특성

- **Detection**: O(n), n = 전체 파일 개수 (일반적으로 <100)
- **Backup**: O(n), n = 사용자 정의 파일 개수
- **Restore**: O(n*m), n = 파일 유형 개수 (최대 4), m = 파일 개수

**예상 실행 시간**: <1초 (파일 수십 개 기준)

---

## Traceability (추적성)

### TAG System

- **SPEC-UPDATE-001**: 본 명세서
- **IMPL-UPDATE-001**: `src/moai_adk/cli/commands/update.py` 구현
- **TEST-UPDATE-001**: `tests/test_cli/test_update_restoration.py` 테스트

### Cross-References

- **관련 SPEC**: 없음 (신규 기능)
- **의존 모듈**:
  - `questionary`: 대화형 UI
  - `rich`: 콘솔 출력
  - `pathlib`, `shutil`: 파일 시스템 작업

### 검증 기준

- ✅ 모든 함수 docstring 작성 완료
- ✅ Type hints 100% 적용
- ✅ 테스트 커버리지 ≥85%
- ✅ `ruff check` 통과
- ✅ `mypy` 타입 검사 통과

---

**END OF SPECIFICATION**
