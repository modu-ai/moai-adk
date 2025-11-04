---
id: CLAUDE-COMMANDS-001
version: 0.1.0
status: completed
created: 2025-10-18
updated: 2025-10-18
author: @Goos
priority: high
category: bugfix
labels:
  - claude-code
  - slash-commands
  - plugin
related_issue: "https://github.com/modu-ai/moai-adk/discussions/30"
scope:
  packages:
    - .claude/commands/
    - src/moai_adk/cli/
    - src/moai_adk/core/diagnostics/
---

# @SPEC:CLAUDE-COMMANDS-001: Claude Code 슬래시 커맨드 로드 실패 문제 해결

## HISTORY

### v0.1.0 (2025-10-18)
- **COMPLETED**: TDD 구현 완료 (RED → GREEN → REFACTOR)
- **TESTED**: 17/17 테스트 통과, 커버리지 90.24%
- **CODE**: 진단 도구 구현 완료
  - src/moai_adk/core/diagnostics/slash_commands.py (160 LOC)
  - src/moai_adk/cli/commands/doctor.py 확장 (+48 LOC)
  - tests/unit/test_slash_commands.py (17 tests)
- **VERIFIED**: TRUST 5원칙 준수
- **DISCOVERY**: alfred/2-build.md 파일에 YAML 파싱 오류 발견 (Discussion #30 원인 가능성)
- **AUTHOR**: @Goos (Alfred 오케스트레이션, tdd-implementer 구현)

### v0.0.1 (2025-10-18)
- **INITIAL**: Claude Code 슬래시 커맨드 0개 로드 문제 진단 및 해결 명세 작성
- **AUTHOR**: @Goos
- **RELATED**: Discussion #30

---

## PROBLEM STATEMENT

### 현상

```
[DEBUG] Total plugin commands loaded: 0
[DEBUG] Slash commands included in SlashCommand tool:
```

사용자가 제공한 로그에 따르면 `.claude/commands/` 디렉토리의 슬래시 커맨드가 전혀 로드되지 않고 있습니다.

### 영향

- `/alfred:0-project`, `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync` 등 모든 Alfred 커맨드 사용 불가
- MoAI-ADK 핵심 워크플로우 완전 차단
- 사용자는 수동으로 에이전트를 호출해야 하는 불편함

### 우선순위

**High**: 슬래시 커맨드는 MoAI-ADK의 핵심 UX이며, 이 기능이 작동하지 않으면 전체 워크플로우 사용 불가

---

## REQUIREMENTS (EARS)

### Ubiquitous Requirements (기본 요구사항)

- 시스템은 `.claude/commands/` 디렉토리의 모든 `.md` 파일을 슬래시 커맨드로 인식해야 한다
- 시스템은 커맨드 로드 실패 시 명확한 오류 메시지를 제공해야 한다

### Event-driven Requirements (이벤트 기반)

- WHEN Claude Code가 시작되면, 시스템은 `.claude/commands/` 디렉토리를 스캔해야 한다
- WHEN 커맨드 파일 형식이 잘못되었으면, 시스템은 해당 파일을 건너뛰고 경고를 출력해야 한다
- WHEN 모든 커맨드가 로드되면, 시스템은 로드된 커맨드 개수를 로그에 출력해야 한다

### State-driven Requirements (상태 기반)

- WHILE 커맨드 로드 중일 때, 시스템은 진행 상황을 DEBUG 레벨 로그로 기록해야 한다
- WHILE 커맨드가 로드된 상태일 때, 시스템은 `/` 접두사로 커맨드를 호출 가능하게 해야 한다

### Constraints (제약사항)

- IF 커맨드 파일에 필수 메타데이터가 없으면, 시스템은 해당 파일을 무시해야 한다
- 커맨드 파일명은 반드시 `.md` 확장자를 가져야 한다

---

## ROOT CAUSE ANALYSIS

### 4가지 가설

1. **파일 형식 문제**
   - Front Matter YAML 형식 오류
   - 필수 필드 누락 (name, description)
   - 인코딩 문제 (BOM, line endings)

2. **디렉토리 구조 문제**
   - `.claude/commands/` 경로 불일치
   - 중첩 디렉토리 구조
   - 심볼릭 링크 문제

3. **파일 권한 문제**
   - 읽기 권한 부족
   - 소유자 불일치

4. **Claude Code 버전 호환성**
   - 구버전 Claude Code 사용
   - 메타데이터 스펙 변경

---

## SOLUTION DESIGN

### 진단 도구 추가

```python
# @CODE:CLAUDE-COMMANDS-001:DIAGNOSTIC
def diagnose_slash_commands():
    """Diagnose slash command loading issues"""
    commands_dir = Path(".claude/commands")

    # Check 1: Directory exists and readable
    if not commands_dir.exists():
        return {"error": "Commands directory not found"}

    # Check 2: Find all .md files
    md_files = list(commands_dir.glob("**/*.md"))

    # Check 3: Validate each file
    results = []
    for file in md_files:
        validation = validate_command_file(file)
        results.append({
            "file": str(file),
            "valid": validation["valid"],
            "errors": validation.get("errors", [])
        })

    return {
        "total_files": len(md_files),
        "valid_commands": sum(1 for r in results if r["valid"]),
        "details": results
    }
```

### 파일 검증 기준

```python
def validate_command_file(file_path: Path) -> dict:
    """Validate slash command file format"""
    try:
        content = file_path.read_text(encoding="utf-8")

        # Check Front Matter
        if not content.startswith("---"):
            return {"valid": False, "errors": ["Missing YAML front matter"]}

        # Parse YAML
        parts = content.split("---", 2)
        if len(parts) < 3:
            return {"valid": False, "errors": ["Invalid front matter format"]}

        metadata = yaml.safe_load(parts[1])

        # Check required fields
        required = ["name", "description"]
        missing = [f for f in required if f not in metadata]
        if missing:
            return {"valid": False, "errors": [f"Missing required field: {', '.join(missing)}"]}

        return {"valid": True}

    except Exception as e:
        return {"valid": False, "errors": [str(e)]}
```

---

## RISKS & MITIGATIONS

| Risk                    | Impact | Mitigation                  |
| ----------------------- | ------ | --------------------------- |
| Claude Code 버전 비호환 | High   | 최신 버전으로 업데이트 권장 |
| 파일 권한 문제          | Medium | chmod 644 적용              |
| 인코딩 문제             | Low    | UTF-8 BOM 제거              |

---

## RELATED DOCUMENTS

- **Development Guide**: `.moai/memory/development-guide.md`
- **SPEC Metadata**: `.moai/memory/spec-metadata.md`
- **Discussion**: https://github.com/modu-ai/moai-adk/discussions/30

---

**작성일**: 2025-10-18
**작성자**: @Goos (spec-builder 에이전트)
**상태**: Draft (v0.0.1)
