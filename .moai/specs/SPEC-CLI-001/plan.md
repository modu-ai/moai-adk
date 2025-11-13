
> **TDD 기반 3단계 구현 전략**
>
> RED → GREEN → REFACTOR 사이클을 엄격히 준수하여 CLI 명령어를 고도화합니다.

---

## 📋 구현 개요

**SPEC ID**: CLI-001
**목표**: doctor/status/restore 명령어 고도화
**예상 기간**: 5일 (각 Phase 1.5일 + 통합 0.5일)
**TDD 전략**: 각 Phase별 RED-GREEN-REFACTOR 독립 실행
**테스트 커버리지 목표**: ≥85% 유지

---

## Phase 1: doctor 명령어 고도화 (2일)

### 🔴 RED: 실패하는 테스트 작성 (0.5일)

**테스트 파일**: `tests/unit/test_doctor_advanced.py`

```python

def test_doctor_detects_python_tools():
    """Python 프로젝트에서 pytest/mypy/ruff 감지"""
    # Given: Python 프로젝트 설정
    # When: doctor 실행
    # Then: pytest, mypy, ruff 체크 결과 반환
    assert "pytest" in result
    assert "mypy" in result
    assert "ruff" in result

def test_doctor_verbose_shows_versions():
    """--verbose 옵션으로 도구 버전 표시"""
    # Given: pytest 8.4.2 설치됨
    # When: doctor --verbose 실행
    # Then: "pytest ✓ (8.4.2)" 표시
    assert "pytest ✓" in output
    assert "8.4.2" in output

def test_doctor_fix_suggests_installation():
    """--fix 옵션으로 설치 명령어 제안"""
    # Given: mypy 미설치
    # When: doctor --fix 실행
    # Then: "pip install mypy" 제안
    assert "pip install mypy" in output

def test_doctor_language_detection():
    """프로젝트 언어 자동 감지"""
    # Given: pyproject.toml 존재
    # When: doctor 실행
    # Then: "Language: Python" 표시
    assert "Language: Python" in output

def test_doctor_timeout_under_5_seconds():
    """진단 실행 시간 5초 이하"""
    # When: doctor --verbose 실행
    # Then: 실행 시간 < 5초
    assert duration < 5.0
```

**실패 예상**: `AttributeError: 'LanguageDetector' object has no attribute 'detect_tools'`

### 🟢 GREEN: 최소 구현 (1일)

**1. 언어별 도구 매핑 추가**

`src/moai_adk/core/project/checker.py`:
```python

LANGUAGE_TOOLS = {
    "python": {
        "required": ["python3", "pip"],
        "recommended": ["pytest", "mypy", "ruff"],
        "optional": ["black", "pylint"]
    },
    "typescript": {
        "required": ["node", "npm"],
        "recommended": ["vitest", "biome"],
        "optional": ["typescript", "eslint"]
    },
    # ... 20개 언어 전체
}

class LanguageToolChecker:
    """언어별 도구 체인 검증"""

    def detect_project_language(self) -> str:
        """프로젝트 언어 감지"""
        if Path("pyproject.toml").exists():
            return "python"
        elif Path("package.json").exists():
            return "typescript"
        # ... 언어 감지 로직
        return "unknown"

    def check_language_tools(self, language: str) -> dict[str, dict]:
        """언어별 도구 체크"""
        tools = LANGUAGE_TOOLS.get(language, {})
        result = {}

        for category in ["required", "recommended", "optional"]:
            for tool in tools.get(category, []):
                result[tool] = {
                    "installed": shutil.which(tool) is not None,
                    "version": self._get_version(tool),
                    "category": category
                }

        return result
```

**2. doctor 명령어 옵션 확장**

`src/moai_adk/cli/commands/doctor.py`:
```python

@click.command()
@click.option("--verbose", is_flag=True, help="Show detailed version information")
@click.option("--fix", is_flag=True, help="Suggest installation commands")
@click.option("--export", type=click.Path(), help="Export results to JSON")
def doctor(verbose: bool, fix: bool, export: str | None) -> None:
    """Advanced system diagnostics with language-specific tool checking"""

    checker = LanguageToolChecker()
    language = checker.detect_project_language()

    console.print(f"[cyan]Language:[/cyan] {language}")

    # 언어별 도구 체크
    tools = checker.check_language_tools(language)

    # Rich 테이블 출력
    table = create_tool_table(tools, verbose)
    console.print(table)

    # --fix 옵션 처리
    if fix:
        suggest_fixes(tools, language)
```

**3. 테스트 통과 확인**

```bash
pytest tests/unit/test_doctor_advanced.py -v
# 예상: 5개 테스트 모두 통과
```

### ♻️ REFACTOR: 코드 품질 개선 (0.5일)

**개선 항목**:
1. **함수 분리**: `check_language_tools` → `_check_tool_installed` + `_get_tool_version`
2. **타입 힌트 강화**: `dict` → `ToolCheckResult` TypedDict
3. **에러 처리**: 도구 버전 파싱 실패 시 graceful fallback
4. **성능 최적화**: 병렬 도구 체크 (asyncio 활용)
5. **린터 경고 제거**: ruff 실행 및 수정

**리팩토링 후 테스트**:
```bash
pytest tests/unit/test_doctor_advanced.py -v
ruff check src/moai_adk/cli/commands/doctor.py
mypy src/moai_adk/cli/commands/doctor.py
```

---

## Phase 2: status 명령어 고도화 (2일)

### 🔴 RED: 실패하는 테스트 작성 (0.5일)

**테스트 파일**: `tests/unit/test_status_advanced.py`

```python

def test_status_detail_shows_tag_chain():
    """--detail 옵션으로 SPEC 체인 무결성 표시"""
    # Given: 1개 SPEC, 1개 CODE TAG
    # When: status --detail 실행
    # Then: "SPEC Chain: 100% (0 orphans, 0 broken)"
    assert "SPEC Chain: 100%" in output

def test_status_detail_shows_coverage():
    """--detail 옵션으로 테스트 커버리지 표시"""
    # Given: pytest-cov 결과 85.61%
    # When: status --detail 실행
    # Then: "Test Coverage: 85.61% ✓ (goal: 85%)"
    assert "Test Coverage: 85.61%" in output

def test_status_detail_shows_quality():
    """--detail 옵션으로 코드 품질 지표 표시"""
    # Given: 0개 린터 경고
    # When: status --detail 실행
    # Then: "Code Quality: 0 warnings"
    assert "Code Quality: 0 warnings" in output

def test_status_detects_broken_tag_chain():
    """끊어진 SPEC 체인 감지"""
    # When: status --detail 실행
    # Then: "⚠ 1 broken chain" 표시
    assert "1 broken chain" in output

def test_status_json_output():
    """--json 옵션으로 JSON 출력"""
    # When: status --json 실행
    # Then: 파싱 가능한 JSON 반환
    data = json.loads(output)
    assert "mode" in data
    assert "specs" in data
```

**실패 예상**: `KeyError: 'tag_chain_integrity'`

### 🟢 GREEN: 최소 구현 (1일)

**1. SPEC 체인 검증 로직**

`src/moai_adk/core/project/tag_checker.py` (신규 생성):
```python

import subprocess
from dataclasses import dataclass

@dataclass
class TagChainResult:
    """SPEC 체인 검증 결과"""
    total_tags: int
    orphans: list[str]  # SPEC 없이 CODE만 있는 TAG
    broken: list[str]   # CODE 없이 SPEC만 있는 TAG
    integrity: float    # (total - orphans - broken) / total

class TagChainChecker:
    """SPEC 체인 무결성 검증"""

    def check_integrity(self) -> TagChainResult:
        """전체 SPEC 체인 검증"""

        orphans = code_tags - spec_tags
        broken = spec_tags - code_tags

        total = len(spec_tags | code_tags)
        valid = total - len(orphans) - len(broken)
        integrity = (valid / total * 100) if total > 0 else 100.0

        return TagChainResult(
            total_tags=total,
            orphans=list(orphans),
            broken=list(broken),
            integrity=integrity
        )

    def _scan_tags(self, path: str, pattern: str) -> set[str]:
        """ripgrep로 TAG 스캔"""
        result = subprocess.run(
            ["rg", pattern, "-n", path],
            capture_output=True,
            text=True
        )
        # TAG ID 추출 로직
        tags = set()
        for line in result.stdout.splitlines():
            tag_id = extract_tag_id(line)
            tags.add(tag_id)
        return tags
```

**2. 커버리지 및 품질 지표 수집**

`src/moai_adk/core/project/quality_checker.py` (신규 생성):
```python

import json
from pathlib import Path

class QualityChecker:
    """코드 품질 지표 수집"""

    def get_test_coverage(self) -> float | None:
        """pytest-cov 결과 파싱"""
        coverage_file = Path(".coverage.json")
        if not coverage_file.exists():
            return None

        with open(coverage_file) as f:
            data = json.load(f)
            return data["totals"]["percent_covered"]

    def get_linter_warnings(self) -> int:
        """ruff 경고 개수"""
        result = subprocess.run(
            ["ruff", "check", "src", "--output-format=json"],
            capture_output=True,
            text=True
        )
        warnings = json.loads(result.stdout)
        return len(warnings)
```

**3. status 명령어 --detail 옵션 추가**

`src/moai_adk/cli/commands/status.py`:
```python

@click.command()
@click.option("--detail", is_flag=True, help="Show quality metrics")
@click.option("--json", "json_output", is_flag=True, help="JSON output")
def status(detail: bool, json_output: bool) -> None:
    """Show project status with optional quality metrics"""

    # 기본 정보 수집
    config = load_config()
    spec_count = count_specs()

    if detail:
        # SPEC 체인 무결성
        tag_checker = TagChainChecker()
        tag_result = tag_checker.check_integrity()

        # 코드 품질 지표
        quality_checker = QualityChecker()
        coverage = quality_checker.get_test_coverage()
        warnings = quality_checker.get_linter_warnings()

        # Rich 테이블에 추가
        table.add_row("TAG Chain", f"{tag_result.integrity:.0f}% ({len(tag_result.orphans)} orphans)")
        table.add_row("Coverage", f"{coverage:.2f}%" if coverage else "N/A")
        table.add_row("Quality", f"{warnings} warnings")

    if json_output:
        # JSON 출력
        data = {
            "mode": config["mode"],
            "specs": spec_count,
            "tag_chain": tag_result if detail else None,
            "coverage": coverage if detail else None
        }
        console.print_json(data=data)
    else:
        # Rich UI 출력
        console.print(panel)
```

**4. 테스트 통과 확인**

```bash
pytest tests/unit/test_status_advanced.py -v
# 예상: 5개 테스트 모두 통과
```

### ♻️ REFACTOR: 코드 품질 개선 (0.5일)

**개선 항목**:
1. **TAG 스캔 최적화**: ripgrep 병렬 실행
2. **캐싱 전략**: TAG 결과 5분 캐싱
3. **에러 처리**: pytest-cov 없을 때 graceful fallback
4. **타입 안전성**: TypedDict → Pydantic 모델
5. **테스트 추가**: 엣지 케이스 (TAG 0개, 커버리지 파일 없음)

---

## Phase 3: restore 명령어 고도화 (1일)

### 🔴 RED: 실패하는 테스트 작성 (0.3일)

**테스트 파일**: `tests/unit/test_restore_advanced.py`

```python

def test_restore_list_shows_backups():
    """--list 옵션으로 백업 목록 표시"""
    # Given: 3개 백업 존재
    # When: restore --list 실행
    # Then: 백업 ID, 타임스탬프, 파일 개수 표시
    assert len(backups) == 3
    assert "backup-001" in output

def test_restore_dry_run_preview():
    """--dry-run 옵션으로 복원 미리보기"""
    # Given: backup-001 존재
    # When: restore --dry-run backup-001 실행
    # Then: 복원될 파일 목록 표시
    assert "config.json" in output
    assert "Will restore 5 files" in output

def test_restore_select_files():
    """--select 옵션으로 선택적 복원"""
    # Given: backup-001에 10개 파일
    # When: restore backup-001 --select config.json
    # Then: config.json만 복원
    assert restored_count == 1

def test_restore_detects_git_dirty():
    """Git dirty state 감지"""
    # Given: 미커밋 변경사항 존재
    # When: restore backup-001 실행
    # Then: 경고 표시 및 확인 요청
    assert "Uncommitted changes" in output
```

**실패 예상**: `AttributeError: 'RestoreCommand' object has no attribute 'list_backups'`

### 🟢 GREEN: 최소 구현 (0.5일)

**1. 백업 메타데이터 확장**

`src/moai_adk/core/template/backup.py`:
```python

@dataclass
class BackupMetadata:
    """백업 메타데이터"""
    backup_id: str
    timestamp: str
    files: list[str]
    size_bytes: int

def create_backup_with_metadata(backup_dir: Path) -> BackupMetadata:
    """백업 생성 및 메타데이터 기록"""
    files = list(Path(".moai").rglob("*"))
    metadata = BackupMetadata(
        backup_id=generate_backup_id(),
        timestamp=datetime.now().isoformat(),
        files=[str(f) for f in files],
        size_bytes=sum(f.stat().st_size for f in files)
    )

    # metadata.json 저장
    with open(backup_dir / "metadata.json", "w") as f:
        json.dump(asdict(metadata), f, indent=2)

    return metadata
```

**2. restore 명령어 옵션 추가**

`src/moai_adk/cli/commands/restore.py`:
```python

@click.command()
@click.argument("backup_id", required=False)
@click.option("--list", "list_backups", is_flag=True, help="List all backups")
@click.option("--dry-run", is_flag=True, help="Preview restore without applying")
@click.option("--select", help="Comma-separated files to restore")
def restore(backup_id: str | None, list_backups: bool, dry_run: bool, select: str | None) -> None:
    """Restore from backup with advanced options"""

    if list_backups:
        # 백업 목록 표시
        backups = load_all_backups()
        table = create_backup_table(backups)
        console.print(table)
        return

    if dry_run:
        # 복원 미리보기
        metadata = load_backup_metadata(backup_id)
        console.print(f"Will restore {len(metadata.files)} files:")
        for file in metadata.files:
            console.print(f"  - {file}")
        return

    # Git dirty state 체크
    if is_git_dirty():
        console.print("[yellow]⚠ Uncommitted changes detected[/yellow]")
        if not click.confirm("Proceed?"):
            return

    # 선택적 복원
    files_to_restore = select.split(",") if select else None
    perform_restore(backup_id, files_to_restore)
```

**3. 테스트 통과 확인**

```bash
pytest tests/unit/test_restore_advanced.py -v
# 예상: 4개 테스트 모두 통과
```

### ♻️ REFACTOR: 코드 품질 개선 (0.2일)

**개선 항목**:
1. **백업 목록 정렬**: 최신 백업 상단 표시
2. **diff 표시**: 복원 전후 변경 사항 diff
3. **에러 처리**: 백업 메타데이터 손상 시 처리
4. **테스트 추가**: 백업 0개, Git repo 아님

---

## 통합 및 마무리 (0.5일)

### 통합 테스트

**테스트 파일**: `tests/integration/test_cli_advanced_integration.py`

```python

def test_full_cli_workflow():
    """전체 CLI 워크플로우 통합 테스트"""
    # Given: 초기화된 프로젝트
    # When: doctor → status --detail → restore --list 순차 실행
    # Then: 모든 명령어 정상 작동

    # Phase 1: doctor
    result = runner.invoke(cli, ["doctor", "--verbose"])
    assert result.exit_code == 0

    # Phase 2: status
    result = runner.invoke(cli, ["status", "--detail"])
    assert result.exit_code == 0
    assert "TAG Chain" in result.output

    # Phase 3: restore
    result = runner.invoke(cli, ["restore", "--list"])
    assert result.exit_code == 0
```

### 최종 검증

```bash
# 전체 테스트 실행
pytest tests/ -v --cov=src/moai_adk/cli/commands --cov-report=term-missing

# 커버리지 목표 확인
# 예상: ≥85%

# 린터 검증
ruff check src/
mypy src/

# 성능 테스트
pytest tests/integration/test_cli_advanced_integration.py --durations=10
```

### 문서 업데이트

**README.md CLI Reference 섹션**:
```markdown
### doctor 명령어

```bash
moai doctor                 # 기본 진단
moai doctor --verbose       # 상세 진단 (도구 버전 표시)
moai doctor --fix           # 자동 수정 제안
moai doctor --export report.json  # JSON 저장
```

### status 명령어

```bash
moai status                 # 기본 정보
moai status --detail        # 품질 지표 추가
moai status --json          # JSON 출력 (CI/CD용)
```

### restore 명령어

```bash
moai restore --list         # 백업 목록
moai restore --dry-run backup-001  # 복원 미리보기
moai restore backup-001 --select config.json  # 선택적 복원
```
```

---

## 예상 위험 및 대응

| 위험                       | 확률 | 영향 | 대응 방안                            |
| -------------------------- | ---- | ---- | ------------------------------------ |
| **도구 버전 파싱 실패**    | 중간 | 낮음 | graceful fallback, "unknown" 표시    |
| **ripgrep 의존성**         | 낮음 | 높음 | 설치 체크 및 대체 로직 (Python 파싱) |
| **진단 시간 초과**         | 중간 | 중간 | 병렬 실행, 타임아웃 5초 강제         |
| **백업 메타데이터 호환성** | 낮음 | 중간 | 버전 필드 추가, 마이그레이션 로직    |

---

## 성공 기준

- ✅ 모든 단위 테스트 통과 (85% 커버리지 유지)
- ✅ 통합 테스트 통과
- ✅ 린터 경고 0개
- ✅ 타입 체크 통과 (mypy)
- ✅ 성능 기준: doctor < 5초, status < 2초, restore --list < 1초
- ✅ 문서 업데이트 완료

---

## 다음 단계

**완료 후 실행**:
```bash
/alfred:2-run CLI-001    # TDD 구현 시작
/alfred:3-sync             # 문서 동기화 및 TAG 검증
```

**권장사항**: 다음 단계(`/alfred:2-run`) 진행 전 `/clear` 또는 `/new` 명령으로 새로운 대화 세션을 시작하면 더 나은 성능과 컨텍스트 관리를 경험할 수 있습니다.
