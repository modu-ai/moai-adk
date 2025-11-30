---
id: SPEC-WORKTREE-001
version: "1.0.0"
status: "draft"
created: "2025-11-27"
updated: "2025-11-27"
---

# SPEC-WORKTREE-001 구현 계획

## 개요

Git Worktree CLI를 6단계(Phase 1-6)에 걸쳐 구현합니다. 각 Phase는 독립적으로 테스트 가능하며, TDD 원칙에 따라 진행됩니다.

**총 예상 범위**: 6 Phases, 8 core commands, ~650 lines of code

---

## Phase 1: Core Infrastructure (핵심 인프라 구축)

### 목표
- WorktreeManager 클래스 구현
- WorktreeRegistry 클래스 구현
- CLI 엔트리포인트 설정

### 주요 작업

**1.1 WorktreeManager 클래스 기본 구조**
```python
# src/moai_adk/cli/worktree/manager.py

class WorktreeManager:
    def __init__(self, repo_path: Path, worktree_root: Path)
    def create(spec_id: str, branch_name: str | None, base_branch: str) -> WorktreeInfo
    def remove(spec_id: str, force: bool) -> None
    def list() -> list[WorktreeInfo]
```

**핵심 기능**:
- GitPython을 통한 Git 저장소 초기화
- Worktree 생성 (`git worktree add`)
- Worktree 제거 (`git worktree remove`)
- 예외 처리 (WorktreeExistsError, GitCommandError)

**1.2 WorktreeRegistry 클래스 구현**
```python
# src/moai_adk/cli/worktree/registry.py

class WorktreeRegistry:
    def __init__(self, worktree_root: Path)
    def register(info: WorktreeInfo) -> None
    def unregister(spec_id: str) -> None
    def get(spec_id: str) -> WorktreeInfo | None
    def list_all() -> list[WorktreeInfo]
```

**핵심 기능**:
- JSON 기반 레지스트리 파일 관리
- CRUD 작업 (Create, Read, Update, Delete)
- 동기화 메커니즘

**1.3 데이터 모델 정의**
```python
# src/moai_adk/cli/worktree/models.py

@dataclass
class WorktreeInfo:
    spec_id: str
    path: Path
    branch: str
    created_at: str
    last_accessed: str
    status: str
```

**1.4 CLI 엔트리포인트 설정**
```python
# src/moai_adk/cli/worktree/cli.py

@click.group()
def worktree():
    """Git Worktree management"""
    pass

@worktree.command()
def new():
    """Create new worktree"""
    pass
```

**1.5 main.py에 worktree 명령어 등록**
```python
# src/moai_adk/cli/main.py

from moai_adk.cli.worktree.cli import worktree

cli.add_command(worktree, name="worktree")
```

### 테스트 전략

**단위 테스트**:
```python
# tests/test_cli/test_worktree_manager.py

def test_create_worktree()
def test_create_duplicate_worktree_raises_error()
def test_remove_worktree()
def test_remove_with_uncommitted_changes_raises_error()
```

```python
# tests/test_cli/test_worktree_registry.py

def test_register_worktree()
def test_unregister_worktree()
def test_get_existing_worktree()
def test_get_nonexistent_worktree_returns_none()
```

### 완료 기준

- [x] WorktreeManager 클래스 구현 완료 (~100 lines)
- [x] WorktreeRegistry 클래스 구현 완료 (~80 lines)
- [x] WorktreeInfo 데이터 모델 정의 (~30 lines)
- [x] CLI 엔트리포인트 설정 완료 (~20 lines)
- [x] 단위 테스트 커버리지 ≥85%
- [x] `ruff check` 통과
- [x] `mypy` 타입 검사 통과

---

## Phase 2: Advanced Commands (고급 명령어 구현)

### 목표
- `moai-worktree list` 명령어 구현
- `moai-worktree switch` 명령어 구현
- `moai-worktree remove` 명령어 구현
- `moai-worktree status` 명령어 구현

### 주요 작업

**2.1 moai-worktree list 구현**
```python
@worktree.command()
@click.option("--format", type=click.Choice(["table", "json"]), default="table")
def list(format: str):
    """List all active worktrees"""
    manager = get_worktree_manager()
    worktrees = manager.list()

    if format == "table":
        display_table(worktrees)  # Rich 테이블
    else:
        print(json.dumps([w.to_dict() for w in worktrees]))
```

**출력 예시 (table)**:
```
┏━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━┓
┃ SPEC ID        ┃ Path                      ┃ Branch               ┃ Status   ┃
┡━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━┩
│ SPEC-AUTH-001  │ ~/worktrees/MoAI/SPEC-... │ feature/SPEC-AUTH... │ active   │
│ SPEC-PAY-002   │ ~/worktrees/MoAI/SPEC-... │ feature/SPEC-PAY-... │ active   │
└────────────────┴───────────────────────────┴──────────────────────┴──────────┘
```

**2.2 moai-worktree switch 구현**
```python
@worktree.command()
@click.argument("spec_id")
def switch(spec_id: str):
    """Switch to another worktree"""
    manager = get_worktree_manager()
    worktree = manager.registry.get(spec_id)

    if not worktree:
        console.print(f"[red]Error:[/red] Worktree {spec_id} not found")
        sys.exit(1)

    # 새 셸 실행
    subprocess.run([os.environ.get("SHELL", "/bin/bash")], cwd=worktree.path)
```

**2.3 moai-worktree remove 구현**
```python
@worktree.command()
@click.argument("spec_id")
@click.option("--force", "-f", is_flag=True, help="Force removal")
def remove(spec_id: str, force: bool):
    """Remove a worktree"""
    manager = get_worktree_manager()

    try:
        manager.remove(spec_id, force=force)
        console.print(f"[green]✓[/green] Worktree {spec_id} removed")
    except UncommittedChangesError:
        console.print(f"[red]Error:[/red] Uncommitted changes. Use --force to override")
```

**2.4 moai-worktree status 구현**
```python
@worktree.command()
def status():
    """Show worktree status"""
    manager = get_worktree_manager()

    # 레지스트리와 Git 동기화
    manager.registry.sync_with_git(manager.repo)

    worktrees = manager.list()
    console.print(f"Total worktrees: {len(worktrees)}")

    for wt in worktrees:
        console.print(f"  {wt.spec_id}: {wt.status} ({wt.branch})")
```

### 테스트 전략

**통합 테스트**:
```python
# tests/test_cli/test_worktree_commands.py

def test_list_command_table_format()
def test_list_command_json_format()
def test_switch_command()
def test_remove_command()
def test_remove_command_with_force()
def test_status_command()
```

### 완료 기준

- [x] `list` 명령어 구현 완료 (~40 lines)
- [x] `switch` 명령어 구현 완료 (~30 lines)
- [x] `remove` 명령어 구현 완료 (~35 lines)
- [x] `status` 명령어 구현 완료 (~25 lines)
- [x] Rich 테이블 출력 구현
- [x] 통합 테스트 작성 및 통과
- [x] 테스트 커버리지 ≥85%

---

## Phase 3: MoAI Integration (MoAI 명령어 통합)

### 목표
- `/moai:1-plan` 명령어에 `--worktree` 플래그 추가
- 3가지 시나리오 지원 (SPEC only, SPEC + branch, SPEC + worktree)

### 주요 작업

**3.1 1-plan.md 수정**
```markdown
# .claude/commands/moai/1-plan.md

## Usage

/moai:1-plan "description"                 # Scenario 1: SPEC only
/moai:1-plan "description" --branch        # Scenario 2: SPEC + branch
/moai:1-plan "description" --worktree      # Scenario 3: SPEC + worktree
```

**3.2 Scenario 3 로직 구현**
```python
# manager-spec 내부 로직 수정

if flags.get("worktree"):
    # SPEC 생성
    spec = create_spec(description)

    # Worktree 생성
    from moai_adk.cli.worktree.manager import WorktreeManager
    manager = WorktreeManager(repo_path, worktree_root)
    worktree = manager.create(spec.id)

    # 사용자 안내
    console.print(f"\n[green]✓[/green] SPEC created: {spec.id}")
    console.print(f"[green]✓[/green] Worktree created: {worktree.path}")
    console.print(f"\n[yellow]Next steps:[/yellow]")
    console.print(f"  1. Switch to worktree: [cyan]moai-worktree switch {spec.id}[/cyan]")
    console.print(f"  2. Or use shell eval: [cyan]eval $(moai-worktree go {spec.id})[/cyan]")
```

**3.3 AskUserQuestion 통합**
```python
# Worktree 생성 여부를 사용자에게 확인

if not flags.get("worktree") and not flags.get("branch"):
    response = AskUserQuestion({
        "questions": [{
            "question": "SPEC 생성 후 Worktree를 생성하시겠습니까?",
            "header": "Worktree 옵션",
            "multiSelect": false,
            "options": [
                {
                    "label": "SPEC만 생성",
                    "description": "SPEC 문서만 생성하고 브랜치/Worktree는 생성하지 않습니다"
                },
                {
                    "label": "브랜치 생성",
                    "description": "SPEC 생성 후 Git 브랜치를 생성합니다"
                },
                {
                    "label": "Worktree 생성",
                    "description": "SPEC 생성 후 독립적인 Worktree 환경을 생성합니다"
                }
            ]
        }]
    })
```

### 테스트 전략

**통합 테스트**:
```python
# tests/test_commands/test_plan_worktree_integration.py

def test_plan_with_worktree_flag()
def test_plan_without_flags_shows_prompt()
def test_plan_with_branch_flag_no_worktree()
```

### 완료 기준

- [x] 1-plan.md 업데이트 완료
- [x] Scenario 3 로직 구현 완료 (~50 lines)
- [x] AskUserQuestion 통합 완료
- [x] 통합 테스트 작성 및 통과
- [x] 사용자 가이드 출력 완료

---

## Phase 4: Additional Commands (추가 명령어)

### 목표
- `moai-worktree go` 명령어 구현 (shell eval 패턴)
- `moai-worktree sync` 명령어 구현
- `moai-worktree clean` 명령어 구현
- `moai-worktree config` 명령어 구현

### 주요 작업

**4.1 moai-worktree go 구현**
```python
@worktree.command()
@click.argument("spec_id")
def go(spec_id: str):
    """Print cd command for shell eval"""
    manager = get_worktree_manager()
    worktree = manager.registry.get(spec_id)

    if not worktree:
        console.print(f"echo 'Error: Worktree {spec_id} not found'", file=sys.stderr)
        sys.exit(1)

    # Shell에서 eval로 실행 가능한 명령어 출력
    print(f"cd {worktree.path}")
```

**사용 예시**:
```bash
# Bash/Zsh
eval $(moai-worktree go SPEC-AUTH-001)

# Fish
moai-worktree go SPEC-AUTH-001 | source
```

**4.2 moai-worktree sync 구현**
```python
@worktree.command()
@click.argument("spec_id")
@click.option("--base", default="main", help="Base branch")
def sync(spec_id: str, base: str):
    """Sync worktree with base branch"""
    manager = get_worktree_manager()

    try:
        manager.sync(spec_id, base_branch=base)
        console.print(f"[green]✓[/green] Synced {spec_id} with {base}")
    except MergeConflictError as e:
        console.print(f"[red]Error:[/red] Merge conflict: {e}")
```

**4.3 moai-worktree clean 구현**
```python
@worktree.command()
def clean():
    """Remove worktrees for merged branches"""
    manager = get_worktree_manager()

    # 병합된 브랜치 탐지
    merged = manager.clean_merged()

    if not merged:
        console.print("[yellow]No merged worktrees found[/yellow]")
        return

    # 확인 프롬프트
    console.print(f"Found {len(merged)} merged worktrees:")
    for spec_id in merged:
        console.print(f"  - {spec_id}")

    if questionary.confirm("Remove these worktrees?").ask():
        for spec_id in merged:
            manager.remove(spec_id)
        console.print(f"[green]✓[/green] Cleaned {len(merged)} worktrees")
```

**4.4 moai-worktree config 구현**
```python
@worktree.command()
@click.argument("key")
@click.argument("value", required=False)
def config(key: str, value: str | None):
    """Get/set worktree configuration"""
    config_path = Path(".moai/worktree-config.json")

    if value is None:
        # Get
        config_data = json.loads(config_path.read_text())
        console.print(f"{key}: {config_data.get(key, 'Not set')}")
    else:
        # Set
        config_data = json.loads(config_path.read_text()) if config_path.exists() else {}
        config_data[key] = value
        config_path.write_text(json.dumps(config_data, indent=2))
        console.print(f"[green]✓[/green] Set {key} = {value}")
```

**설정 항목**:
- `worktree_root`: Worktree 루트 디렉토리 (기본: `~/worktrees/{{PROJECT_NAME}}`)
- `default_base_branch`: 기본 base 브랜치 (기본: `main`)
- `auto_sync`: 자동 동기화 여부 (기본: `false`)

### 테스트 전략

**단위 테스트**:
```python
# tests/test_cli/test_worktree_advanced_commands.py

def test_go_command()
def test_sync_command()
def test_sync_command_with_conflict()
def test_clean_command()
def test_clean_command_no_merged()
def test_config_get()
def test_config_set()
```

### 완료 기준

- [x] `go` 명령어 구현 완료 (~20 lines)
- [x] `sync` 명령어 구현 완료 (~30 lines)
- [x] `clean` 명령어 구현 완료 (~40 lines)
- [x] `config` 명령어 구현 완료 (~30 lines)
- [x] 단위 테스트 작성 및 통과
- [x] 테스트 커버리지 ≥85%

---

## Phase 5: Documentation & Skills (문서화 및 스킬)

### 목표
- README.ko.md에 Git Worktree 섹션 추가
- Skill 문서 작성 (moai-domain-worktree.md)
- 사용 가이드 및 예시 작성

### 주요 작업

**5.1 README.ko.md 업데이트**
```markdown
## 🔀 Git Worktree CLI (병렬 개발 지원)

MoAI-ADK는 Git Worktree를 활용한 병렬 SPEC 개발을 지원합니다.

### 핵심 기능
- 여러 SPEC을 동시에 독립적으로 개발
- 각 SPEC은 별도의 디렉토리와 브랜치 소유
- 빠른 전환 및 자동 정리

### 시작하기

1. SPEC 생성 + Worktree 생성
```bash
/moai:1-plan "User authentication" --worktree
```

2. Worktree 목록 확인
```bash
moai-worktree list
```

3. Worktree 전환
```bash
moai-worktree switch SPEC-AUTH-001
# 또는
eval $(moai-worktree go SPEC-AUTH-001)
```

4. 작업 완료 후 정리
```bash
moai-worktree clean
```
```

**5.2 Skill 문서 작성**
```markdown
# .claude/skills/moai-domain-worktree/SKILL.md

## Quick Reference (30s)

Git Worktree 기반 병렬 SPEC 개발 전문가

**Core Commands**:
- `moai-worktree new <spec-id>` - 새 Worktree 생성
- `moai-worktree list` - Worktree 목록 조회
- `moai-worktree switch <spec-id>` - Worktree 전환
- `moai-worktree remove <spec-id>` - Worktree 제거

...
```

**5.3 예시 문서 작성**
```markdown
# .claude/skills/moai-domain-worktree/examples.md

## Example 1: 병렬로 2개 SPEC 개발

# SPEC 1: 인증 기능
/moai:1-plan "User authentication with JWT" --worktree
moai-worktree switch SPEC-AUTH-001

# 작업 중...

# SPEC 2: 결제 기능 (다른 터미널)
/moai:1-plan "Payment integration with Stripe" --worktree
moai-worktree switch SPEC-PAY-002

# 작업 중...

## Example 2: Main 브랜치와 동기화

moai-worktree sync SPEC-AUTH-001
```

### 완료 기준

- [x] README.ko.md 업데이트 완료 (~100 lines)
- [x] Skill 문서 작성 완료 (~300 lines)
- [x] 예시 문서 작성 완료 (~200 lines)
- [x] 사용자 가이드 검토 완료

---

## Phase 6: Polish & Testing (최종 다듬기 및 테스트)

### 목표
- 종합 통합 테스트
- 에러 메시지 개선
- 성능 최적화
- 최종 문서 검토

### 주요 작업

**6.1 종합 통합 테스트**
```python
# tests/test_integration/test_worktree_workflow.py

def test_full_workflow():
    """전체 워크플로우 테스트"""
    # 1. SPEC 생성 + Worktree
    # 2. 다른 SPEC 생성 + Worktree
    # 3. 목록 확인
    # 4. 전환
    # 5. 동기화
    # 6. 정리
    pass

def test_error_scenarios():
    """에러 시나리오 테스트"""
    # 중복 생성, 존재하지 않는 worktree, 충돌 등
    pass
```

**6.2 에러 메시지 개선**
```python
# 명확한 에러 메시지 및 해결 방법 안내

[red]Error:[/red] Worktree SPEC-AUTH-001 already exists
  Path: ~/worktrees/MoAI-ADK/SPEC-AUTH-001
  Tip: Use 'moai-worktree switch SPEC-AUTH-001' to navigate to it
```

**6.3 성능 최적화**
- 레지스트리 캐싱
- Git 작업 최소화
- 대용량 프로젝트 테스트 (1GB+)

**6.4 최종 문서 검토**
- 모든 명령어 예시 실행 확인
- 스크린샷 추가 (테이블 출력 등)
- FAQ 섹션 작성

### 완료 기준

- [x] 종합 통합 테스트 통과
- [x] 에러 메시지 개선 완료
- [x] 성능 최적화 완료
- [x] 최종 문서 검토 완료
- [x] 전체 테스트 커버리지 ≥85%
- [x] `ruff check` 통과
- [x] `mypy` 타입 검사 통과

---

## 전체 완료 기준 (Definition of Done)

### 코드 품질

- [x] 모든 함수에 docstring 작성
- [x] Type hints 100% 적용
- [x] 테스트 커버리지 ≥85%
- [x] `ruff check` 통과
- [x] `mypy` 타입 검사 통과
- [x] 모든 예외 처리 완료

### 기능 완성도

- [x] 8개 핵심 명령어 구현 완료
- [x] `/moai:1-plan` 통합 완료
- [x] 레지스트리 동기화 완료
- [x] 에러 핸들링 완료
- [x] 성능 최적화 완료

### 문서화

- [x] README.ko.md 업데이트
- [x] Skill 문서 작성 (moai-domain-worktree)
- [x] 예시 문서 작성
- [x] 명령어 도움말 작성

### 테스트

- [x] 단위 테스트 작성
- [x] 통합 테스트 작성
- [x] 종합 통합 테스트 작성
- [x] 에러 시나리오 테스트

---

## 주요 기술 제약 사항

### 라이브러리 버전

- **GitPython**: 3.1.43 이상 (2025-11 최신 안정 버전)
- **Click**: 8.1.7 이상 (2025-11 최신 안정 버전)
- **Rich**: 13.9.4 이상 (2025-11 최신 안정 버전)
- **Questionary**: 2.0.0 이상 (안정 버전)

### 기술적 제약

1. **GitPython 호환성**: Git 2.30 이상 필요
2. **디스크 공간**: 프로젝트 크기 × N worktrees
3. **파일 시스템**: POSIX 호환 (macOS, Linux)
4. **셸 환경**: Bash, Zsh, Fish 지원

### 보안 고려사항

1. **Git 작업 격리**: Worktree는 독립적이지만 Git 히스토리는 공유
2. **파일 권한**: Worktree 디렉토리 권한 검사
3. **레지스트리 보호**: 레지스트리 파일 무결성 검사

---

## 리스크 및 대응 방안

### 리스크 1: GitPython 버전 호환성 문제

**발생 확률**: 중간
**영향도**: 높음

**대응 방안**:
- 최소 버전 명시 (≥3.1.43)
- CI/CD에서 다양한 Git 버전 테스트
- 에러 발생 시 명확한 메시지 제공

### 리스크 2: 디스크 공간 부족

**발생 확률**: 낮음
**영향도**: 중간

**대응 방안**:
- Worktree 생성 전 디스크 공간 확인
- `moai-worktree clean` 명령어로 정리 유도
- 최대 worktree 개수 제한 (설정 가능)

### 리스크 3: 레지스트리 불일치

**발생 확률**: 중간
**영향도**: 낮음

**대응 방안**:
- `moai-worktree status` 실행 시 자동 동기화
- 레지스트리 복구 명령어 제공
- Git worktree 상태를 소스 오브 트루스로 활용

---

## 다음 단계 (After Phase 6)

1. **사용자 피드백 수집**: 실제 사용 후 개선점 파악
2. **성능 모니터링**: 대규모 프로젝트에서의 성능 측정
3. **추가 기능 고려**:
   - `moai-worktree diff <spec1> <spec2>`: Worktree 간 차이 비교
   - `moai-worktree backup`: Worktree 백업 기능
   - `moai-worktree restore`: Worktree 복원 기능

---

**END OF IMPLEMENTATION PLAN**
