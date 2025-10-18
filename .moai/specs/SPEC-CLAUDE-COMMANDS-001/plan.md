# 구현 계획 - CLAUDE-COMMANDS-001

> **SPEC**: `.moai/specs/SPEC-CLAUDE-COMMANDS-001/spec.md`

---

## 우선순위 기반 마일스톤

### 1차 목표: 문제 진단 및 근본 원인 파악

**범위**:
- `.claude/commands/` 디렉토리 구조 및 파일 검증
- 각 `.md` 파일의 Front Matter 형식 검증
- 필수 필드 (name, description) 존재 여부 확인
- 파일 권한 및 인코딩 검증

**산출물**:
- 진단 보고서 (JSON 형식)
- 문제 파일 목록 및 오류 원인
- 권장 수정 방법

### 2차 목표: 진단 도구 구현

**범위**:
- `python -m moai_adk doctor --check-commands` 명령어 추가
- 파일별 검증 로직 구현
- 상세한 오류 메시지 및 해결 방법 제시

**산출물**:
- `src/moai_adk/cli/commands/doctor.py` 업데이트
- 진단 결과 포맷터 (human-readable + JSON)

### 3차 목표: 문제 파일 수정 및 검증

**범위**:
- 진단 도구로 식별된 문제 파일 수정
- Front Matter YAML 형식 정규화
- 필수 필드 추가
- 인코딩 통일 (UTF-8, BOM 제거)

**산출물**:
- 수정된 커맨드 파일 목록
- Claude Code 재시작 후 검증 결과

### 최종 목표: 자동화 및 문서화

**범위**:
- Pre-commit hook에 슬래시 커맨드 검증 추가
- 개발자 가이드 업데이트 (커맨드 작성 규칙)
- CI/CD 파이프라인에 검증 단계 통합

**산출물**:
- `.moai/hooks/pre-commit` 업데이트
- `docs/slash-commands-guide.md` 작성
- CI workflow 파일 업데이트

---

## 기술적 접근 방법

### 1. 진단 도구 구현 전략

**기술 스택**:
- Python 3.11+
- `pathlib` (파일 시스템 탐색)
- `pyyaml` (YAML 파싱)
- `rich` (터미널 출력 포맷팅)

**알고리즘**:
1. `.claude/commands/` 디렉토리 순회
2. 모든 `.md` 파일 수집 (`glob("**/*.md")`)
3. 각 파일에 대해:
   - UTF-8로 읽기 시도
   - Front Matter 존재 여부 확인
   - YAML 파싱 및 필수 필드 검증
   - 결과 수집 (valid/invalid + errors)
4. 집계 및 보고서 생성

### 2. 파일 검증 규칙

**필수 Front Matter 구조**:
```yaml
---
name: command-name
description: "Command description"
---
```

**선택 필드**:
- `aliases`: 커맨드 별칭 (배열)
- `arguments`: 인자 정의 (배열)
- `examples`: 사용 예시 (문자열)

**검증 항목**:
- Front Matter 시작/종료 (`---`)
- YAML 파싱 가능 여부
- 필수 필드 존재 여부
- 필드 타입 검증 (name: string, description: string)

### 3. 오류 복구 전략

**자동 수정 가능**:
- 인코딩 변환 (UTF-8 BOM 제거)
- 공백 문자 정규화 (trailing spaces 제거)
- Line ending 통일 (LF)

**수동 수정 필요**:
- Front Matter 누락
- 필수 필드 누락
- YAML 구문 오류

---

## 아키텍처 설계 방향

### 모듈 구조

```
src/moai_adk/
├── cli/
│   └── commands/
│       └── doctor.py  # 진단 명령어
├── core/
│   └── validation/
│       ├── __init__.py
│       ├── slash_command_validator.py  # 커맨드 검증기 (신규)
│       └── report_formatter.py         # 보고서 포맷터 (신규)
└── utils/
    └── file_utils.py  # 파일 읽기/쓰기 유틸리티
```

### 주요 클래스 설계

**SlashCommandValidator**:
```python
class SlashCommandValidator:
    """Validates slash command .md files"""

    def validate_directory(self, path: Path) -> ValidationReport:
        """Validate all .md files in directory"""

    def validate_file(self, file_path: Path) -> FileValidation:
        """Validate single command file"""

    def check_front_matter(self, content: str) -> dict:
        """Check YAML front matter format"""

    def check_required_fields(self, metadata: dict) -> list[str]:
        """Check for missing required fields"""
```

**ReportFormatter**:
```python
class ReportFormatter:
    """Format validation results"""

    def format_json(self, report: ValidationReport) -> str:
        """Format as JSON"""

    def format_table(self, report: ValidationReport) -> str:
        """Format as rich table"""

    def format_summary(self, report: ValidationReport) -> str:
        """Format as summary text"""
```

---

## 리스크 및 대응 방안

### High Priority Risks

1. **Claude Code 버전 비호환**
   - **리스크**: 구버전 Claude Code가 새로운 메타데이터 형식을 지원하지 않음
   - **영향**: 파일을 수정해도 커맨드가 로드되지 않음
   - **대응**:
     - Claude Code 버전 체크 추가
     - 공식 문서 참조하여 버전별 스펙 확인
     - 최신 버전으로 업데이트 권장

2. **파일 시스템 권한 문제**
   - **리스크**: `.claude/commands/` 디렉토리 또는 파일에 읽기 권한 없음
   - **영향**: 진단 도구가 파일을 읽을 수 없음
   - **대응**:
     - 권한 체크 추가 (`os.access(path, os.R_OK)`)
     - 권한 수정 가이드 제공 (`chmod 644`)

### Medium Priority Risks

3. **인코딩 문제**
   - **리스크**: UTF-8 BOM, CP949 등 다양한 인코딩 혼재
   - **영향**: YAML 파싱 실패
   - **대응**:
     - 여러 인코딩으로 읽기 시도 (UTF-8, UTF-8-sig, CP949)
     - 자동 변환 도구 제공

4. **중첩 디렉토리 구조**
   - **리스크**: `.claude/commands/sub/command.md` 같은 중첩 구조
   - **영향**: Claude Code가 중첩 디렉토리를 지원하지 않을 수 있음
   - **대응**:
     - Flat 구조 권장
     - 경고 메시지 출력

---

## 의존성 관리

### 새로운 의존성 (필요 시 추가)

```toml
[project.dependencies]
pyyaml = ">=6.0"
rich = ">=13.0"
```

### 기존 의존성 활용

- `pathlib`: 파일 시스템 탐색
- `json`: JSON 출력
- `argparse`: CLI 인자 파싱

---

## 테스트 전략

### 단위 테스트

**테스트 케이스**:
1. 유효한 커맨드 파일 (Front Matter + 필수 필드)
2. Front Matter 누락
3. 필수 필드 누락 (name 또는 description)
4. YAML 구문 오류
5. 인코딩 문제 (UTF-8 BOM)
6. 권한 문제 (읽기 불가)

**테스트 파일**:
```
tests/
└── unit/
    └── validation/
        └── test_slash_command_validator.py
```

### 통합 테스트

**시나리오**:
1. `.claude/commands/` 디렉토리 전체 검증
2. 진단 도구 실행 및 출력 확인
3. Claude Code 재시작 후 커맨드 로드 확인

---

## 완료 기준 (Definition of Done)

- ✅ 진단 도구 구현 완료 (`python -m moai_adk doctor --check-commands`)
- ✅ 모든 `.md` 파일 검증 및 오류 수정
- ✅ Claude Code 재시작 시 커맨드 정상 로드 (개수 > 0)
- ✅ 단위 테스트 커버리지 ≥ 85%
- ✅ 통합 테스트 통과
- ✅ 문서 업데이트 (slash-commands-guide.md)
- ✅ PR 승인 및 머지

---

**작성일**: 2025-10-18
**작성자**: spec-builder 에이전트
