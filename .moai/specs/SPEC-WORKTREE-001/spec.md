---
id: SPEC-WORKTREE-001
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
| 1.0.0 | 2025-11-27 | GOOS | Git Worktree CLI 초안 - Phase 1-6 완전 구현 |

# SPEC-WORKTREE-001: Git Worktree CLI for Parallel SPEC Development

## 개요

MoAI-ADK에 Git Worktree 기반 병렬 개발 환경을 구축하는 CLI 도구를 추가합니다. 여러 SPEC을 동시에 개발할 수 있도록 독립적인 워킹 디렉토리를 생성하고 관리하는 8개의 핵심 명령어를 제공합니다.

**핵심 가치**:
- 🔀 **병렬 개발**: 여러 SPEC을 동시에 작업 가능
- 🔒 **격리 환경**: 각 SPEC은 독립적인 브랜치와 디렉토리 소유
- 🚀 **빠른 전환**: `moai-worktree switch` 또는 `moai-worktree go`로 즉시 이동
- 🧹 **자동 정리**: 병합된 SPEC의 worktree 자동 정리

---

## Environment (실행 환경 및 전제 조건)

### 시스템 환경

- **Python**: 3.9 이상
- **Git**: 2.30 이상 (Worktree 기능 지원)
- **필수 패키지**:
  - `GitPython>=3.1.43` - Git 작업 자동화
  - `click>=8.1.7` - CLI 프레임워크
  - `rich>=13.9.4` - 터미널 UI 포맷팅
  - `questionary>=2.0.0` - 대화형 프롬프트
  - `pathlib` - 경로 처리 (표준 라이브러리)
  - `json` - 레지스트리 저장 (표준 라이브러리)

### 프로젝트 구조

```
~/worktrees/{{PROJECT_NAME}}/
├── SPEC-AUTH-001/           # Worktree 디렉토리
│   ├── .git                 # Worktree Git 메타데이터
│   ├── src/                 # 독립적인 소스 코드
│   └── .moai/
│       └── specs/SPEC-AUTH-001/  # SPEC 문서
├── SPEC-PAYMENT-002/        # 또 다른 Worktree
│   └── ...
└── .moai-worktree-registry.json  # Worktree 레지스트리

{{PROJECT_ROOT}}/
├── .moai/
│   └── worktree-config.json      # Worktree 설정
└── .git/
    └── worktrees/                # Git Worktree 메타데이터
```

### 기존 코드 베이스

- **Integration Point**: `.claude/commands/moai/1-plan.md` - SPEC 생성 시 worktree 옵션 추가
- **New Module**: `src/moai_adk/cli/worktree/` - Worktree 관리 모듈
- **New Command**: `moai-worktree` - CLI 엔트리포인트

---

## Assumptions (가정 사항)

### 기술적 가정

1. **Git 저장소**: 사용자는 Git 저장소에서 MoAI-ADK를 사용한다고 가정
2. **디스크 공간**: Worktree 생성에 충분한 디스크 공간 (프로젝트 크기 × N)
3. **파일 시스템 접근**: `~/worktrees/` 디렉토리 생성 권한 보유
4. **GitPython 호환성**: GitPython이 시스템 Git과 정상 동작

### 운영 가정

1. **브랜치 전략**: 3-Mode Git 전략과 독립적으로 동작
2. **SPEC 명명**: SPEC-{DOMAIN}-{NUMBER} 형식 준수
3. **동시 작업**: 최대 5-10개 worktree 동시 관리 (성능 고려)
4. **셸 환경**: Bash/Zsh에서 `eval` 기반 디렉토리 이동 지원

---

## Requirements (요구사항)

### Ubiquitous Requirements (항상 활성화)

**REQ-WORKTREE-001**: 시스템은 SPEC ID를 기반으로 새로운 Git Worktree를 생성해야 한다.
- **조건**: `~/worktrees/{{PROJECT_NAME}}/{{SPEC-ID}}/` 경로에 생성
- **출력**: Worktree 생성 완료 메시지 + 경로

**REQ-WORKTREE-002**: 시스템은 생성된 모든 worktree를 레지스트리에 기록해야 한다.
- **조건**: `~/worktrees/{{PROJECT_NAME}}/.moai-worktree-registry.json` 파일
- **출력**: SPEC ID, 경로, 브랜치, 생성일, 상태 기록

**REQ-WORKTREE-003**: 시스템은 현재 활성화된 모든 worktree 목록을 조회할 수 있어야 한다.
- **조건**: Git과 레지스트리 정보 동기화
- **출력**: 테이블 또는 JSON 형식 출력

**REQ-WORKTREE-004**: 시스템은 worktree 간 빠른 전환을 지원해야 한다.
- **조건**: `moai-worktree switch <spec-id>` 명령어
- **출력**: 새 셸 실행 또는 디렉토리 이동 안내

**REQ-WORKTREE-005**: 시스템은 더 이상 필요하지 않은 worktree를 제거할 수 있어야 한다.
- **조건**: `moai-worktree remove <spec-id>` 명령어
- **출력**: Worktree 삭제 + 레지스트리 업데이트

### Event-driven Requirements (이벤트 기반)

**REQ-WORKTREE-006**: WHEN 사용자가 `/moai:1-plan "description" --worktree` 실행 시, THEN 시스템은 SPEC 생성 후 자동으로 worktree를 생성해야 한다.
- **이벤트**: `--worktree` 플래그 감지
- **액션**: SPEC 생성 → Worktree 생성 → 사용자에게 이동 안내

**REQ-WORKTREE-007**: WHEN 사용자가 `moai-worktree new <spec-id> --branch <branch-name>` 실행 시, THEN 시스템은 지정된 브랜치를 생성하고 worktree를 만들어야 한다.
- **이벤트**: `--branch` 플래그 감지
- **액션**: 브랜치 생성 → Worktree 생성

**REQ-WORKTREE-008**: WHEN 사용자가 `moai-worktree switch <spec-id>` 실행 시, THEN 시스템은 해당 worktree로 이동해야 한다.
- **이벤트**: `switch` 명령어 실행
- **액션**: 새 셸 실행 또는 `eval` 패턴으로 디렉토리 이동

**REQ-WORKTREE-009**: WHEN 사용자가 `moai-worktree sync <spec-id>` 실행 시, THEN 시스템은 main 브랜치의 최신 변경사항을 가져와야 한다.
- **이벤트**: `sync` 명령어 실행
- **액션**: `git fetch origin main && git merge origin/main` 실행

**REQ-WORKTREE-010**: WHEN 사용자가 `moai-worktree clean` 실행 시, THEN 시스템은 병합된 브랜치의 worktree를 자동 정리해야 한다.
- **이벤트**: `clean` 명령어 실행
- **액션**: 병합된 브랜치 탐지 → Worktree 제거 확인 프롬프트

### State-driven Requirements (상태 기반)

**REQ-WORKTREE-011**: WHILE worktree 디렉토리가 존재하지 않는 경우, THEN 시스템은 기본 worktree 루트 디렉토리를 생성해야 한다.
- **상태**: `~/worktrees/{{PROJECT_NAME}}/` 부재
- **액션**: `Path.mkdir(parents=True, exist_ok=True)`

**REQ-WORKTREE-012**: WHILE 레지스트리 파일이 존재하지 않는 경우, THEN 시스템은 빈 레지스트리를 초기화해야 한다.
- **상태**: `.moai-worktree-registry.json` 부재
- **액션**: 빈 JSON 객체 `{}` 생성

**REQ-WORKTREE-013**: WHILE worktree가 이미 존재하는 경우, THEN 시스템은 중복 생성을 방지하고 경고 메시지를 표시해야 한다.
- **상태**: 동일 SPEC ID의 worktree 존재
- **액션**: 생성 중단 + 경고 출력

### Unwanted Behaviors (금지사항)

**REQ-WORKTREE-014**: 시스템은 메인 저장소의 `.git` 디렉토리를 절대로 수정해서는 안 된다 (worktree 메타데이터 제외).
- **금지**: 메인 저장소의 브랜치/커밋 직접 변경
- **검증**: GitPython을 통한 안전한 worktree 작업만 허용

**REQ-WORKTREE-015**: 시스템은 커밋되지 않은 변경사항이 있는 worktree를 강제 삭제해서는 안 된다.
- **금지**: `--force` 플래그 없이 변경사항이 있는 worktree 삭제
- **검증**: `git status --porcelain` 검사

**REQ-WORKTREE-016**: 시스템은 레지스트리와 실제 Git worktree 상태가 불일치할 경우 자동으로 동기화해야 한다.
- **금지**: 불일치 상태 방치
- **대응**: `moai-worktree status` 실행 시 자동 동기화

### Optional Requirements (선택사항)

**REQ-WORKTREE-017**: 시스템은 worktree 생성 시 의존성 자동 설치를 지원할 수 있다.
- **선택사항**: `--install-deps` 플래그로 `uv sync` 실행
- **현재**: 수동 설치

**REQ-WORKTREE-018**: 시스템은 worktree 간 파일 변경 비교를 지원할 수 있다.
- **선택사항**: `moai-worktree diff <spec-id1> <spec-id2>`
- **현재**: 미지원

---

## Specifications (기술 사양)

### 아키텍처 설계

```
┌─────────────────────────────────────────────────────────┐
│               moai-worktree CLI 아키텍처                  │
└─────────────────────────────────────────────────────────┘
                         │
      ┌──────────────────┼──────────────────┐
      │                  │                  │
      ▼                  ▼                  ▼
┌──────────┐      ┌──────────┐      ┌──────────┐
│ CLI Layer│      │ Manager  │      │ Registry │
│  (Click) │─────▶│  Layer   │─────▶│  Layer   │
└──────────┘      └──────────┘      └──────────┘
      │                  │                  │
      │                  ▼                  ▼
      │           ┌──────────┐      ┌──────────┐
      │           │   Git    │      │   JSON   │
      └──────────▶│ Python   │      │  Store   │
                  └──────────┘      └──────────┘
```

**Layer 1: CLI Layer** (`src/moai_adk/cli/worktree/cli.py`)
- Click 기반 명령어 정의
- 사용자 입력 검증
- Rich 기반 출력 포맷팅

**Layer 2: Manager Layer** (`src/moai_adk/cli/worktree/manager.py`)
- Worktree 생성/삭제/전환 비즈니스 로직
- GitPython을 통한 Git 작업 추상화
- 에러 핸들링 및 검증

**Layer 3: Registry Layer** (`src/moai_adk/cli/worktree/registry.py`)
- Worktree 메타데이터 관리
- JSON 기반 영구 저장
- 동기화 및 충돌 해결

**Layer 4: Integration Layer** (`.claude/commands/moai/1-plan.md`)
- `/moai:1-plan` 명령어와 통합
- 3가지 시나리오 지원 (SPEC only, SPEC + branch, SPEC + worktree)

### 주요 함수 명세

#### 1. WorktreeManager 클래스

```python
class WorktreeManager:
    """
    Git Worktree 생성, 관리, 삭제를 담당하는 핵심 클래스
    """

    def __init__(self, repo_path: Path, worktree_root: Path):
        """
        Args:
            repo_path: 메인 Git 저장소 경로
            worktree_root: Worktree 루트 디렉토리 (~/worktrees/{project}/)
        """
        self.repo = git.Repo(repo_path)
        self.worktree_root = worktree_root
        self.registry = WorktreeRegistry(worktree_root)

    def create(
        self,
        spec_id: str,
        branch_name: str | None = None,
        base_branch: str = "main"
    ) -> WorktreeInfo:
        """
        새 Worktree 생성

        Args:
            spec_id: SPEC ID (예: SPEC-AUTH-001)
            branch_name: 생성할 브랜치 이름 (None이면 spec_id 기반)
            base_branch: 기준 브랜치 (기본: main)

        Returns:
            WorktreeInfo: 생성된 worktree 정보

        Raises:
            WorktreeExistsError: 이미 존재하는 worktree
            GitError: Git 작업 실패
        """

    def remove(self, spec_id: str, force: bool = False) -> None:
        """
        Worktree 제거

        Args:
            spec_id: 제거할 SPEC ID
            force: 변경사항 무시하고 강제 삭제

        Raises:
            WorktreeNotFoundError: 존재하지 않는 worktree
            UncommittedChangesError: 커밋되지 않은 변경사항 (force=False)
        """

    def list(self) -> list[WorktreeInfo]:
        """
        모든 worktree 목록 반환

        Returns:
            List[WorktreeInfo]: Worktree 정보 리스트
        """

    def sync(self, spec_id: str, base_branch: str = "main") -> None:
        """
        Worktree를 base_branch의 최신 상태로 동기화

        Args:
            spec_id: 동기화할 SPEC ID
            base_branch: 동기화 기준 브랜치

        Raises:
            MergeConflictError: 병합 충돌 발생
        """

    def clean_merged(self) -> list[str]:
        """
        병합된 브랜치의 worktree 정리

        Returns:
            List[str]: 정리된 SPEC ID 리스트
        """
```

**복잡도**:
- `create()`: O(1) - Git 작업
- `remove()`: O(1) - 파일 시스템 작업
- `list()`: O(n), n = worktree 개수
- `sync()`: O(1) - Git 작업
- `clean_merged()`: O(n*m), n = worktree 개수, m = 브랜치 개수

**예상 라인 수**: ~200 lines

#### 2. WorktreeRegistry 클래스

```python
class WorktreeRegistry:
    """
    Worktree 메타데이터 영구 저장 및 관리
    """

    def __init__(self, worktree_root: Path):
        """
        Args:
            worktree_root: 레지스트리 파일이 위치할 디렉토리
        """
        self.registry_path = worktree_root / ".moai-worktree-registry.json"
        self._load()

    def register(self, info: WorktreeInfo) -> None:
        """
        Worktree 정보 등록

        Args:
            info: 등록할 worktree 정보
        """

    def unregister(self, spec_id: str) -> None:
        """
        Worktree 정보 제거

        Args:
            spec_id: 제거할 SPEC ID
        """

    def get(self, spec_id: str) -> WorktreeInfo | None:
        """
        특정 worktree 정보 조회

        Args:
            spec_id: 조회할 SPEC ID

        Returns:
            WorktreeInfo | None: Worktree 정보 (없으면 None)
        """

    def list_all(self) -> list[WorktreeInfo]:
        """
        모든 worktree 정보 반환

        Returns:
            List[WorktreeInfo]: 전체 worktree 정보
        """

    def sync_with_git(self, repo: git.Repo) -> None:
        """
        레지스트리를 Git worktree 상태와 동기화

        Args:
            repo: Git 저장소 객체
        """
```

**복잡도**: O(1) - 대부분 JSON 읽기/쓰기
**예상 라인 수**: ~150 lines

#### 3. CLI 명령어 함수

```python
@click.group()
def worktree():
    """Git Worktree management for parallel SPEC development"""
    pass

@worktree.command()
@click.argument("spec_id")
@click.option("--branch", "-b", help="Branch name")
@click.option("--base", default="main", help="Base branch")
def new(spec_id: str, branch: str | None, base: str):
    """Create a new worktree for SPEC"""
    pass

@worktree.command()
@click.option("--format", type=click.Choice(["table", "json"]), default="table")
def list(format: str):
    """List all active worktrees"""
    pass

@worktree.command()
@click.argument("spec_id")
def switch(spec_id: str):
    """Switch to another worktree (new shell)"""
    pass

@worktree.command()
@click.argument("spec_id")
@click.option("--force", "-f", is_flag=True)
def remove(spec_id: str, force: bool):
    """Remove a worktree"""
    pass

@worktree.command()
def status():
    """Show worktree status and sync registry"""
    pass

@worktree.command()
@click.argument("spec_id")
def go(spec_id: str):
    """Print cd command for shell eval"""
    # Shell에서 `eval $(moai-worktree go SPEC-001)` 형태로 사용
    pass

@worktree.command()
@click.argument("spec_id")
@click.option("--base", default="main")
def sync(spec_id: str, base: str):
    """Sync worktree with base branch"""
    pass

@worktree.command()
def clean():
    """Remove worktrees for merged branches"""
    pass

@worktree.command()
@click.argument("key")
@click.argument("value", required=False)
def config(key: str, value: str | None):
    """Get/set worktree configuration"""
    pass
```

**예상 라인 수**: ~250 lines (9개 명령어 × 평균 25-30 lines)

#### 4. 통합 함수 (1-plan.md 수정)

```python
# .claude/commands/moai/1-plan.md 내부 로직 수정

def handle_plan_with_worktree(description: str, flags: dict):
    """
    /moai:1-plan의 3가지 시나리오 처리

    Scenario 1: SPEC only (no flags)
        → SPEC 생성만

    Scenario 2: SPEC + branch (--branch flag)
        → SPEC 생성 + 브랜치 생성

    Scenario 3: SPEC + worktree (--worktree flag)
        → SPEC 생성 + Worktree 생성 + 사용자 이동 안내
    """
    # SPEC 생성 (기존 로직)
    spec = create_spec(description)

    # Scenario 분기
    if flags.get("worktree"):
        # Worktree 생성
        manager = WorktreeManager(repo_path, worktree_root)
        worktree = manager.create(spec.id)

        console.print(f"[green]✓[/green] Worktree created: {worktree.path}")
        console.print(f"[yellow]→[/yellow] Switch to worktree: moai-worktree switch {spec.id}")
        console.print(f"[yellow]→[/yellow] Or use shell eval: eval $(moai-worktree go {spec.id})")

    elif flags.get("branch"):
        # 브랜치 생성만 (기존 로직)
        create_branch(spec.id)

    # Scenario 1은 기존 로직 유지
```

**예상 라인 수**: ~50 lines (기존 1-plan.md 수정)

### 데이터 모델

```python
@dataclass
class WorktreeInfo:
    """Worktree 메타데이터"""
    spec_id: str                # SPEC-AUTH-001
    path: Path                  # ~/worktrees/MoAI-ADK/SPEC-AUTH-001/
    branch: str                 # feature/SPEC-AUTH-001
    created_at: str             # ISO 8601 timestamp
    last_accessed: str          # ISO 8601 timestamp
    status: str                 # active, merged, stale

    def to_dict(self) -> dict:
        """JSON 직렬화"""
        return {
            "spec_id": self.spec_id,
            "path": str(self.path),
            "branch": self.branch,
            "created_at": self.created_at,
            "last_accessed": self.last_accessed,
            "status": self.status
        }

    @classmethod
    def from_dict(cls, data: dict) -> "WorktreeInfo":
        """JSON 역직렬화"""
        return cls(
            spec_id=data["spec_id"],
            path=Path(data["path"]),
            branch=data["branch"],
            created_at=data["created_at"],
            last_accessed=data["last_accessed"],
            status=data["status"]
        )
```

### 에러 핸들링

| 에러 시나리오 | 예외 클래스 | 처리 방법 |
|--------------|-----------|----------|
| Worktree 중복 생성 | `WorktreeExistsError` | 경고 출력 + 기존 worktree 경로 안내 |
| Worktree 없음 | `WorktreeNotFoundError` | 에러 메시지 + 사용 가능한 worktree 목록 표시 |
| 커밋되지 않은 변경사항 | `UncommittedChangesError` | `--force` 플래그 안내 |
| Git 작업 실패 | `GitCommandError` | 상세 에러 메시지 + Git 명령어 로그 |
| 병합 충돌 | `MergeConflictError` | 충돌 파일 목록 + 수동 해결 안내 |
| 레지스트리 불일치 | `RegistryInconsistencyError` | 자동 동기화 시도 + 실패 시 경고 |

### 성능 특성

- **Worktree 생성**: ~1-2초 (Git 작업 + 파일 복사)
- **Worktree 삭제**: ~0.5초 (파일 시스템 작업)
- **목록 조회**: <0.1초 (레지스트리 읽기)
- **동기화**: 1-5초 (네트워크 속도 의존)
- **정리**: O(n), n = worktree 개수 (일반적으로 <10개)

**디스크 사용량**: 프로젝트 크기 × worktree 개수
- 예: 500MB 프로젝트 × 5 worktrees = 2.5GB

---

## Traceability (추적성)

### TAG System

- **SPEC-WORKTREE-001**: 본 명세서
- **IMPL-WORKTREE-001**: `src/moai_adk/cli/worktree/` 구현
- **TEST-WORKTREE-001**: `tests/test_cli/test_worktree.py` 테스트
- **DOC-WORKTREE-001**: `README.ko.md` 문서 업데이트

### Cross-References

- **관련 SPEC**: 없음 (신규 기능)
- **의존 모듈**:
  - `GitPython>=3.1.43`: Git 작업 자동화
  - `click>=8.1.7`: CLI 프레임워크
  - `rich>=13.9.4`: 터미널 UI
  - `questionary>=2.0.0`: 대화형 프롬프트
- **통합 지점**:
  - `.claude/commands/moai/1-plan.md`: SPEC 생성 시 worktree 옵션
  - `src/moai_adk/cli/main.py`: CLI 엔트리포인트 등록

### 검증 기준

- ✅ 모든 함수 docstring 작성 완료
- ✅ Type hints 100% 적용
- ✅ 테스트 커버리지 ≥85%
- ✅ `ruff check` 통과
- ✅ `mypy` 타입 검사 통과
- ✅ GitPython 3.1.43+ 버전 호환성 확인
- ✅ 모든 Git 작업 로컬 테스트 통과

---

**END OF SPECIFICATION**
