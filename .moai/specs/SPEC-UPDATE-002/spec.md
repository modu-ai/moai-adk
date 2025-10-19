---
id: UPDATE-002
version: 0.0.1
status: draft
created: 2025-10-19
updated: 2025-10-19
author: @Goos
priority: high
category: feature
labels:
  - update
  - template
  - merge
  - optimizer
  - settings
related_specs:
  - UPDATE-REFACTOR-001
  - INIT-003
scope:
  packages:
    - src/moai_adk/cli/commands
    - src/moai_adk/core/template
  files:
    - update.py
    - merger.py
    - config.py
    - 0-project.md
---

# @SPEC:UPDATE-002: 스마트 템플릿 업데이트 및 병합 시스템

## HISTORY

### v0.0.1 (2025-10-19)
- **INITIAL**: 스마트 템플릿 업데이트 및 병합 시스템 명세 작성
- **AUTHOR**: @Goos
- **REASON**: moai-adk update와 /alfred:0-project update 역할 분리 및 settings.json 병합 지원

---

## 개요

MoAI-ADK의 템플릿 업데이트 로직을 2단계로 분리하여, 패키지 업데이트와 프로젝트 최적화를 독립적으로 관리합니다.

**핵심 변경사항**:
1. **moai-adk update**: PyPI 최신 버전 업데이트, `optimized=false` 설정
2. **/alfred:0-project update**: 백업과 신규 템플릿 비교 후 스마트 병합, `optimized=true` 설정
3. **settings.json 병합 지원**: 사용자 환경 변수 및 권한 설정 보존

---

## Ubiquitous Requirements (기본 요구사항)

### R1: 2단계 업데이트 프로세스
- 시스템은 템플릿 업데이트를 **패키지 업데이트**와 **프로젝트 최적화** 2단계로 분리해야 한다

### R2: moai-adk update 기본 동작
- 시스템은 `moai-adk update` 명령 실행 시 `optimized` 상태와 관계없이 항상 템플릿 파일을 업데이트해야 한다

### R3: /alfred:0-project update 신규 커맨드
- 시스템은 `/alfred:0-project update` 커맨드를 통해 백업과 신규 템플릿 비교 및 병합 기능을 제공해야 한다

### R4: 병합 대상 파일
- 시스템은 다음 파일들의 스마트 병합을 지원해야 한다:
  - `CLAUDE.md` (루트)
  - `.moai/config.json`
  - `.claude/settings.json`
  - `.moai/memory/development-guide.md`

---

## Event-driven Requirements (이벤트 기반)

### E1: moai-adk update 실행 시
- WHEN `moai-adk update` 명령을 실행하면, 시스템은 다음을 수행해야 한다:
  1. PyPI에서 최신 버전 확인
  2. 백업 생성 (unless --force)
  3. 템플릿 파일 복사
  4. `config.json`의 `project.optimized`를 `false`로 설정
  5. 사용자에게 `/alfred:0-project update` 실행 안내

### E2: /alfred:0-project update 실행 시
- WHEN `/alfred:0-project update` 명령을 실행하면, 시스템은 다음을 수행해야 한다:
  1. `config.json`의 `optimized` 필드 확인
  2. 최근 백업 디렉토리 탐색
  3. 백업과 신규 템플릿 비교
  4. 병합 계획 생성 및 사용자 확인
  5. 사용자 승인 후 병합 실행
  6. `optimized=true` 설정

### E3: CLAUDE.md 병합 시
- WHEN `CLAUDE.md` 병합을 수행하면, 시스템은 다음을 보존해야 한다:
  - "## 프로젝트 정보" 섹션 (백업에서 추출)
  - 나머지는 신규 템플릿 내용 적용

### E4: config.json 병합 시
- WHEN `config.json` 병합을 수행하면, 시스템은 깊은 병합(deep merge)을 통해 사용자 설정을 보존해야 한다:
  - `project.*` 필드 (name, mode, locale, custom_fields 등)
  - 사용자 추가 설정

### E5: settings.json 병합 시
- WHEN `settings.json` 병합을 수행하면, 시스템은 다음을 보존해야 한다:
  - `env` 객체 (사용자 커스텀 환경 변수)
  - `permissions.allow` 배열 (사용자 추가 권한)
  - `permissions.ask` 배열 (사용자 추가 확인 항목)

---

## State-driven Requirements (상태 기반)

### S1: optimized=false 상태일 때
- WHILE `config.json`의 `optimized=false`일 때, `/alfred:0-project update` 실행 시 병합 프로세스를 진행해야 한다

### S2: optimized=true 상태일 때
- WHILE `config.json`의 `optimized=true`일 때, `/alfred:0-project update` 실행 시 "이미 최적화 완료" 메시지를 출력하고 종료해야 한다

### S3: 백업이 존재할 때
- WHILE 최근 백업 디렉토리가 존재할 때, 백업 파일과 신규 템플릿을 비교하여 병합 계획을 수립해야 한다

### S4: 백업이 없을 때
- WHILE 백업 디렉토리가 없을 때, 병합 없이 신규 템플릿을 그대로 복사해야 한다

---

## Optional Features (선택적 기능)

### O1: 파일별 병합 전략 선택
- WHERE 사용자가 파일별 병합 전략을 선택하면, 시스템은 다음 옵션을 제공할 수 있다:
  1. 스마트 병합 (권장)
  2. 백업 우선
  3. 템플릿 우선
  4. 수동 병합 (diff 표시)
  5. 건너뛰기

### O2: diff 표시
- WHERE 사용자가 수동 병합을 선택하면, 시스템은 백업과 신규 템플릿의 차이를 표시할 수 있다

---

## Constraints (제약사항)

### C1: 사용자 데이터 보호
- IF 병합 작업을 수행하면, 다음 디렉토리는 절대 수정하지 않아야 한다:
  - `.moai/specs/`
  - `.moai/reports/`
  - `.moai/project/` (사용자 확인 없이)

### C2: 백업 필수
- IF 병합 작업을 수행하면, 반드시 백업을 먼저 생성해야 한다 (unless --force)

### C3: settings.json 병합 규칙
- IF `settings.json` 병합을 수행하면, 다음 규칙을 준수해야 한다:
  - `env` 객체는 shallow merge (사용자 변수 우선)
  - `permissions.allow`는 배열 병합 (중복 제거)
  - `permissions.deny`는 템플릿 우선 (보안 강화)

### C4: CLAUDE.md 프로젝트 정보 보존
- IF `CLAUDE.md` 병합을 수행하면, "## 프로젝트 정보" 섹션은 반드시 보존되어야 한다

### C5: 버전 체계
- 본 SPEC의 버전은 다음 규칙을 따라야 한다:
  - v0.0.1: INITIAL (draft)
  - v0.1.0: TDD 구현 완료 (completed)
  - v1.0.0: 정식 안정화 (사용자 승인 필수)

---

## 구현 세부사항

### 파일별 수정 계획

#### 1. update.py
**위치**: `src/moai_adk/cli/commands/update.py`

**변경사항**:
```python
# 수정: 라인 102-127
# 기존: optimized 상태 확인 후 종료
# 신규: optimized 상태 무시, 항상 업데이트 수행

# 추가: set_optimized_false() 함수
def set_optimized_false(project_path: Path) -> None:
    """config.json의 optimized 필드를 false로 설정"""
    config_path = project_path / ".moai" / "config.json"
    if config_path.exists():
        config = json.loads(config_path.read_text())
        config.setdefault("project", {})["optimized"] = False
        config_path.write_text(json.dumps(config, indent=2, ensure_ascii=False))

# 수정: 라인 148 (완료 메시지)
console.print("ℹ️  Next step: Run /alfred:0-project update to optimize template changes")
```

#### 2. 0-project.md
**위치**: `.claude/commands/alfred/0-project.md`

**추가 섹션**:
```markdown
## 🚀 STEP 3: 템플릿 최적화 (update 서브커맨드)

### 3.1 최적화 필요 여부 확인
- config.json의 optimized 필드 확인

### 3.2 백업과 신규 템플릿 비교
- 최근 백업 디렉토리 탐색
- 주요 파일 비교:
  - CLAUDE.md (루트)
  - .moai/config.json
  - .claude/settings.json
  - .moai/memory/*.md

### 3.3 병합 계획 생성 및 사용자 확인

### 3.4 병합 실행
- 자동 병합: config.json, settings.json
- 사용자 선택: CLAUDE.md, development-guide.md

### 3.5 최적화 완료
- optimized=true 설정
```

#### 3. merger.py
**위치**: `src/moai_adk/core/template/merger.py`

**추가 메서드**:
```python
def merge_settings_json(self, template_path: Path, existing_path: Path, backup_path: Path) -> None:
    """settings.json 스마트 병합

    Rules:
    - env: 사용자 환경 변수 보존 (shallow merge)
    - permissions.allow: 배열 병합 (중복 제거)
    - permissions.deny: 템플릿 우선 (보안)
    """
    # 백업에서 사용자 설정 추출
    backup_data = json.loads(backup_path.read_text())
    template_data = json.loads(template_path.read_text())

    # env shallow merge (사용자 우선)
    merged_env = {**template_data.get("env", {}), **backup_data.get("env", {})}

    # permissions.allow 배열 병합 (중복 제거)
    template_allow = set(template_data.get("permissions", {}).get("allow", []))
    backup_allow = set(backup_data.get("permissions", {}).get("allow", []))
    merged_allow = sorted(template_allow | backup_allow)

    # permissions.deny는 템플릿 우선 (보안)
    merged_deny = template_data.get("permissions", {}).get("deny", [])

    # 최종 병합
    merged = {
        "env": merged_env,
        "hooks": template_data.get("hooks", {}),
        "permissions": {
            "defaultMode": template_data.get("permissions", {}).get("defaultMode", "default"),
            "allow": merged_allow,
            "ask": template_data.get("permissions", {}).get("ask", []),
            "deny": merged_deny
        }
    }

    existing_path.write_text(json.dumps(merged, indent=2, ensure_ascii=False))
```

#### 4. config.py
**위치**: `src/moai_adk/core/template/config.py`

**추가 메서드**:
```python
@staticmethod
def set_optimized(project_path: Path, value: bool) -> None:
    """config.json의 optimized 필드 설정"""
    config_path = project_path / ".moai" / "config.json"
    if not config_path.exists():
        return

    config = json.loads(config_path.read_text())
    config.setdefault("project", {})["optimized"] = value
    config_path.write_text(json.dumps(config, indent=2, ensure_ascii=False))
```

---

## 테스트 시나리오

### 시나리오 1: 정상적인 업데이트
```bash
# 1단계: 패키지 업데이트
$ moai-adk update
→ 템플릿 복사, optimized=false

# 2단계: 템플릿 최적화
$ /alfred:0-project update
→ 백업 비교, 병합 계획, 사용자 확인, 병합 수행, optimized=true
```

### 시나리오 2: settings.json 사용자 설정 보존
```bash
# 사용자가 .claude/settings.json에 커스텀 환경 변수 추가
env:
  CUSTOM_VAR: "my_value"

# moai-adk update 후
→ settings.json이 새 템플릿으로 대체됨

# /alfred:0-project update 실행
→ CUSTOM_VAR 보존됨 (백업에서 복원)
```

### 시나리오 3: CLAUDE.md 프로젝트 정보 보존
```bash
# 사용자가 CLAUDE.md에 프로젝트 정보 작성
## 프로젝트 정보
- 이름: My Project
- 버전: 1.0.0

# moai-adk update 후
→ CLAUDE.md가 새 템플릿으로 대체됨

# /alfred:0-project update 실행
→ 프로젝트 정보 섹션 보존됨
```

---

## 성공 기준

### 기능 검증
- [ ] moai-adk update가 optimized 상태와 관계없이 항상 실행됨
- [ ] moai-adk update 완료 후 optimized=false 설정됨
- [ ] /alfred:0-project update가 2단계 워크플로우로 실행됨 (Phase 1: 분석, Phase 2: 실행)
- [ ] 백업과 신규 템플릿 비교가 정확하게 수행됨
- [ ] CLAUDE.md의 "프로젝트 정보" 섹션이 보존됨
- [ ] config.json의 사용자 설정이 보존됨 (깊은 병합)
- [ ] settings.json의 env 환경 변수가 보존됨
- [ ] settings.json의 permissions.allow가 병합됨 (중복 제거)
- [ ] 병합 완료 후 optimized=true 설정됨

### 품질 검증
- [ ] 테스트 커버리지 ≥85%
- [ ] 모든 파일 ≤300 LOC
- [ ] 모든 함수 ≤50 LOC
- [ ] mypy 타입 검사 통과
- [ ] ruff 린터 검사 통과

---

## 관련 문서

- `SPEC-UPDATE-REFACTOR-001`: 기존 업데이트 리팩토링
- `SPEC-INIT-003`: 템플릿 처리기
- `.moai/memory/development-guide.md`: 개발 가이드
- `.moai/memory/spec-metadata.md`: SPEC 메타데이터 표준

---

**다음 단계**: `/alfred:2-build UPDATE-002` 실행하여 TDD 구현 시작
