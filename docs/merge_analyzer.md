# MergeAnalyzer - Intelligent Template Merge Analysis

## 개요

`MergeAnalyzer`는 moai-adk 업데이트 시 **백업된 사용자 설정**과 **새로운 템플릿**을 비교하여 충돌 위험도를 분석하고 사용자 확인을 받은 후 병합을 진행하는 모듈입니다.

**Claude Code headless mode**를 활용하여 지능형 분석을 수행하며, Claude를 사용할 수 없는 경우 difflib 기반 폴백 분석으로 안전성을 보장합니다.

## 기능

### 1. Claude 기반 지능형 분석
- **Claude Sonnet 4.5** 모델 사용 (`claude-sonnet-4-5-20250929`)
- 안전한 읽기 전용 도구만 사용 (`Read`, `Glob`, `Grep`)
- 120초 타임아웃으로 수행 시간 제어
- JSON 형식 응답으로 파싱 가능

### 2. 안전한 폴백 메커니즘
- Claude 호출 실패 시 자동으로 difflib 기반 분석으로 전환
- 어떤 상황에서도 병합 의사결정 제공

### 3. 사용자 대화형 확인
- Rich 라이브러리로 포맷팅된 테이블 형식 출력
- 변경 파일, 위험도, 권장 조치를 명확히 표시
- `click.confirm()` 으로 사용자 승인 요청
- 취소 시 자동으로 백업에서 롤백

## 분석 대상 파일

MergeAnalyzer는 다음 4개 파일만 분석합니다:

| 파일 | 우선도 | 설명 |
|------|--------|------|
| `CLAUDE.md` | 높음 | 프로젝트 정보, 사용자 커스터마이징 |
| `.claude/settings.json` | 높음 | 보안 정책, 도구 권한 설정 |
| `.moai/config/config.json` | 높음 | 프로젝트 메타데이터, 버전 정보 |
| `.gitignore` | 낮음 | 무시 파일 목록 |

## 병합 규칙

각 파일별 권장 병합 전략:

### CLAUDE.md
- **전략**: `smart_merge` (스마트 병합)
- **규칙**:
  - 사용자 "Project Information" 섹션 보존
  - 새로운 기능 설명 추가
  - 글로벌 정책과 프로젝트 정보 병합

### .claude/settings.json
- **전략**: `smart_merge` 또는 `use_template`
- **규칙**:
  - 기존 도구 권한 유지
  - 새로운 보안 설정 추가
  - 사용자 커스텀 설정 보존

### .moai/config/config.json
- **전략**: `keep_existing` + 버전만 업데이트
- **규칙**:
  - `project.name`, `project.version` 보존 (사용자 식별자)
  - `moai.version` 업데이트 (패키지 버전)
  - 신규 필드는 템플릿에서 추가

### .gitignore
- **전략**: `smart_merge` (항목만 추가)
- **규칙**:
  - 기존 항목 보존
  - 새로운 무시 항목 추가
  - 중복 제거

## 사용 흐름

### 1. 백업 생성
```
moai-adk update
  ↓
백업 생성: .moai-backups/backup/
  ↓
차이점 수집
```

### 2. Claude 분석
```
MergeAnalyzer 초기화
  ↓
분석 프롬프트 생성
  ↓
claude -p "prompt" --output-format json --tools Read,Glob,Grep
  ↓
JSON 응답 파싱
```

### 3. 폴백 (Claude 불가 시)
```
Claude 호출 실패 감지
  ↓
difflib 기반 폴백 분석
  ↓
간단한 위험도 평가
```

### 4. 사용자 확인
```
분석 결과 표시 (Rich 테이블)
  ↓
위험도 및 권장사항 표시
  ↓
사용자 확인: "병합을 진행하시겠습니까?"
  ↓
취소 시: 백업에서 복구
진행 시: 템플릿 동기화
```

## 위험도 등급

### Low (낮음, 녹색)
- `.gitignore` 같은 단순한 추가 항목
- 문서 업데이트
- 자동으로 안전하게 병합 가능

### Medium (중간, 노란색)
- 새로운 보안 설정 추가
- 부분적인 구조 변경
- 사용자 확인 후 진행

### High (높음, 빨강)
- 프로젝트 메타데이터 변경 (이름, 버전)
- 주요 설정 파일 변경
- 사용자 데이터 손실 가능성
- **반드시 사용자 확인 필수**

## API 사용

### 기본 사용법

```python
from pathlib import Path
from moai_adk.core.merge import MergeAnalyzer

project_path = Path(".")
backup_path = Path(".moai-backups/backup")
template_path = Path("src/moai_adk/templates")

# 분석기 초기화
analyzer = MergeAnalyzer(project_path)

# 분석 수행
analysis = analyzer.analyze_merge(backup_path, template_path)

# 사용자 확인
if analyzer.ask_user_confirmation(analysis):
    # 병합 진행
    print("병합 진행 중...")
else:
    # 취소 - 백업에서 복구
    print("병합 취소됨")
```

### 반환 값 구조

```python
{
    "files": [
        {
            "filename": "CLAUDE.md",
            "changes": "메타데이터 수정",
            "recommendation": "smart_merge",  # use_template, keep_existing, smart_merge
            "conflict_severity": "low",      # low, medium, high
            "note": "추가 설명 (선택사항)"
        }
    ],
    "safe_to_auto_merge": False,    # 자동 병합 안전 여부
    "user_action_required": True,    # 사용자 개입 필요 여부
    "summary": "종합 요약",
    "risk_assessment": "위험도 평가",
    "fallback": False               # 폴백 사용 여부
}
```

## 에러 처리

### Claude 호출 실패
- **TimeoutExpired**: 120초 초과
- **FileNotFoundError**: Claude Code 설치 안 됨
- **JSONDecodeError**: 응답 파싱 실패

모든 경우 자동으로 **폴백 분석**으로 전환됩니다.

### 사용자 취소
- 병합 확인에서 "아니오" 선택
- 자동으로 백업에서 복구 (`backup.restore_backup()`)
- 업데이트 중단

## 성능 특성

| 작업 | 시간 | 비용 |
|------|------|------|
| 백업 생성 | ~1초 | 무료 |
| Claude 분석 | ~30-60초 | ~$0.10-0.20 |
| 폴백 분석 | ~1초 | 무료 |
| 사용자 확인 | ~30초 | 무료 |
| **전체** | **~2분** | **~$0.10-0.20** |

## 테스트

### 단위 테스트 (18개)
```bash
uv run pytest tests/test_merge_analyzer.py -v
```

- 초기화, 파일 수집, 프롬프트 생성
- 폴백 분석, 명령어 구성
- 결과 표시, Mock 테스트

### 통합 테스트 (6개)
```bash
uv run pytest tests/test_merge_analyzer_integration.py -v
```

- 전체 워크플로우 (Claude + 폴백)
- 다중 파일 변경 분석
- 사용자 확인 플로우

## 마이그레이션 가이드

### update.py 통합

```python
# update.py의 _sync_templates() 함수에 추가

from moai_adk.core.merge import MergeAnalyzer

if not force and backup_path:
    analyzer = MergeAnalyzer(project_path)
    template_path = Path(__file__).parent.parent.parent / "templates"

    analysis = analyzer.analyze_merge(backup_path, template_path)

    if not analyzer.ask_user_confirmation(analysis):
        console.print("[yellow]⚠️ 사용자가 업데이트를 취소했습니다.[/yellow]")
        backup.restore_backup(backup_path)
        return False
```

## 향후 개선 사항

- [ ] 다국어 지원 (현재: 한국어)
- [ ] 커스텀 병합 규칙 설정 가능
- [ ] 병합 미리보기 기능
- [ ] 부분 병합 선택 기능
- [ ] 병합 로그 저장

## 참고

- **Claude Code 문서**: https://claude.com/claude-code
- **headless mode**: `claude -p "prompt"` 형식
- **Read-only tools**: 안전한 읽기 전용 도구만 사용

