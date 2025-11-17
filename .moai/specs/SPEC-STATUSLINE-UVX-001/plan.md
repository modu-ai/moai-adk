---
spec_id: SPEC-STATUSLINE-UVX-001
title: "uvx 기반 Statusline 실행 구현 계획"
version: 1.0
status: draft
created_at: 2025-11-17
---

## 구현 개요

SPEC-STATUSLINE-UVX-001을 구현하기 위한 단계별 로드맵입니다. 본 계획은 우선순위 기반으로 구성되며, 각 마일스톤은 독립적으로 진행 가능합니다.

---

## 마일스톤 구조

### Primary Goal (필수)

uvx 기반 statusline 실행을 모든 OS에서 지원하기 위한 핵심 구현

#### M1: CLI 진입점 추가
**목표**: `moai-adk statusline` 명령어 실행 가능하게 함

**기술 접근**:
1. `pyproject.toml`의 `[project.scripts]` 섹션에 새로운 진입점 추가
   ```toml
   [project.scripts]
   moai-adk = "moai_adk.__main__:main"
   "moai-adk-statusline" = "moai_adk.statusline.main:main"
   ```

2. 또는 기존 `moai-adk` 명령어에 서브커맨드로 통합
   ```python
   # moai_adk/__main__.py
   @click.group()
   def cli():
       pass

   @cli.command()
   def statusline():
       from moai_adk.statusline.main import main
       main()
   ```

**선택 기준**:
- **방식 A (권장)**: 독립 CLI 진입점 → 더 빠른 실행
- **방식 B**: 서브커맨드 → 더 통합적 설계

**산출물**:
- 수정된 `pyproject.toml`
- 테스트 코드: `test_statusline_cli_entry_point()`

#### M2: 로컬 설정 업데이트 (.claude/settings.json)
**목표**: MoAI-ADK 로컬 프로젝트에서 uvx 방식 적용

**기술 접근**:
1. `.claude/settings.json`의 `statusLine` 필드 수정
   ```json
   {
     "statusLine": {
       "type": "command",
       "command": "uvx moai-adk statusline",
       "padding": 0
     }
   }
   ```

2. 대체 명령어 추가 (향후 폴백용)
   ```json
   {
     "statusLine": {
       "type": "command",
       "command": "uvx moai-adk statusline",
       "fallback": "uv run python -m moai_adk.statusline.main",
       "padding": 0
     }
   }
   ```

**산출물**:
- 수정된 `.claude/settings.json`
- 수정 전후 비교 문서

#### M3: 템플릿 동기화
**목표**: 신규 프로젝트가 uvx 방식으로 자동 설정되도록 함

**기술 접근**:
1. `src/moai_adk/templates/.claude/settings.json` 업데이트
   - M2와 동일한 내용으로 변경
   - 신규 프로젝트 생성 시 자동 적용

2. 마이그레이션 가이드 문서 작성
   - 기존 프로젝트를 uvx로 업그레이드하는 방법
   - 폴백 메커니즘 설명

**산출물**:
- 수정된 템플릿 파일
- 마이그레이션 가이드 문서 (`.moai/docs/migration-uvx.md`)

#### M4: 테스트 구현
**목표**: 모든 OS에서 statusline이 정상 작동하는지 확인

**기술 접근**:
1. 단위 테스트 (단일 함수)
   ```python
   def test_statusline_main_function():
       """Test main() function execution"""
       # Mock stdin, verify output

   def test_statusline_entry_point():
       """Test CLI entry point is callable"""
       # Run: moai-adk statusline
   ```

2. 통합 테스트 (CLI 레벨)
   ```python
   def test_uvx_statusline_command():
       """Test: uvx moai-adk statusline"""
       result = subprocess.run(['uvx', 'moai-adk', 'statusline'], ...)
       assert result.returncode == 0

   def test_uv_run_fallback():
       """Test fallback: uv run ..."""
       result = subprocess.run(['uv', 'run', 'python', '-m', '...'], ...)
       assert result.returncode == 0
   ```

3. 플랫폼별 테스트
   - Windows: Process handle 누수 확인
   - macOS/Linux: 신호 처리 확인

**산출물**:
- 테스트 코드: `tests/statusline/test_uvx_entry_point.py`
- 테스트 리포트 (coverage ≥ 85%)

---

### Secondary Goal (선택사항)

사용자 경험 및 안정성 개선

#### M5: 폴백 메커니즘 구현 (향후)
**목표**: uvx 미지원 환경에서도 statusline 작동하도록 함

**기술 접근**:
1. 워퍼 스크립트 작성 (`.claude/bin/statusline.sh`)
   ```bash
   #!/bin/bash
   # Try uvx first
   if command -v uvx &> /dev/null; then
       uvx moai-adk statusline "$@"
   else
       # Fallback to uv run
       uv run python -m moai_adk.statusline.main
   fi
   ```

2. Python 기반 워퍼 (더 나은 옵션)
   ```python
   # .claude/bin/statusline.py
   def find_statusline_command():
       if shutil.which('uvx'):
           return ['uvx', 'moai-adk', 'statusline']
       elif shutil.which('uv'):
           return ['uv', 'run', 'python', '-m', 'moai_adk.statusline.main']
       else:
           raise RuntimeError("Neither uvx nor uv found")
   ```

3. Claude Code 설정에서 워퍼 호출
   ```json
   {
     "statusLine": {
       "command": "python .claude/bin/statusline.py",
       "padding": 0
     }
   }
   ```

**산출물**:
- 워퍼 스크립트: `.claude/bin/statusline.sh` 또는 `.py`
- 수정된 설정

#### M6: 성능 벤치마크 (향후)
**목표**: uvx vs uv run 성능 비교 검증

**기술 접근**:
1. 벤치마크 스크립트 작성
   ```python
   # Test 100회 실행 시간 측정
   - uvx moai-adk statusline
   - uv run python -m moai_adk.statusline.main
   ```

2. 결과 분석
   - Windows: 30-40% 성능 개선 확인
   - macOS/Linux: 10-20% 개선 확인

**산출물**:
- 벤치마크 스크립트: `tests/benchmarks/statusline_performance.py`
- 성능 리포트

#### M7: 문서 및 가이드 (향후)
**목표**: 사용자가 쉽게 이해하고 적용할 수 있도록 함

**산출물**:
- 마이그레이션 가이드
- 트러블슈팅 가이드
- README 업데이트

---

## 기술 결정 사항

### 1. CLI 진입점 설계 (M1)

**결정**: 독립 CLI 진입점 추가 (방식 A)

**근거**:
- `uvx moai-adk-statusline` → 더 명확한 의도
- 실행 속도 향상 (불필요한 click 파싱 제거)
- Python 모듈 직접 호출 가능

**대안 검토**:
| 방식 | 장점 | 단점 | 추천도 |
|------|------|------|--------|
| 독립 진입점 | 빠름, 명확 | 다소 중복 | ⭐⭐⭐⭐⭐ |
| 서브커맨드 | 통합적, 일관성 | 약간 느림 | ⭐⭐⭐ |
| 래퍼 스크립트 | 유연함 | 복잡도 높음 | ⭐⭐ |

### 2. 명령어 이름 결정

**결정**: `uvx moai-adk statusline` (서브커맨드)

**근거**:
- 기존 사용자가 이미 `moai-adk` 명령어 사용
- 일관성 유지
- 향후 다른 서브커맨드 추가 가능

**구현**:
```python
# pyproject.toml
[project.scripts]
moai-adk = "moai_adk.cli:main"  # 기존

# moai_adk/cli.py
@click.group()
def main():
    pass

@main.command('statusline')
def statusline_command():
    from moai_adk.statusline.main import main as statusline_main
    statusline_main()
```

### 3. 폴백 전략 (M5 - 향후)

**결정**: Python 기반 워퍼 스크립트 + 환경 변수

**근거**:
- 모든 OS에서 동일하게 작동
- 디버깅 용이
- 추가 라이브러리 불필요

**구현 우선순위**:
1. Primary Phase: uvx 만 지원
2. Secondary Phase: 폴백 구현 (필요 시)
3. Tertiary Phase: 워퍼 최적화

---

## 구현 순서

### 단계 1: 코드 변경 (1-2일)
1. M1: CLI 진입점 추가
2. M2: 로컬 설정 업데이트
3. M3: 템플릿 동기화
4. M4: 테스트 작성

### 단계 2: 검증 (1일)
- 모든 OS에서 테스트 (Windows, macOS, Linux)
- 캐시 동작 확인
- 성능 측정

### 단계 3: 배포 (1일)
- PyPI에 새 버전 푸시 (0.25.11)
- GitHub Release 작성
- 마이그레이션 가이드 공지

### 단계 4: 향후 개선 (선택)
- M5: 폴백 메커니즘
- M6: 성능 벤치마크
- M7: 문서 확장

---

## 의존성 및 위험

### 의존성
- `uv` ≥ 0.4.0 (uvx 명령어 필요)
- `pyproject.toml` 수정 권한
- PyPI 배포 권한

### 위험 및 완화

| 위험 | 영향도 | 확률 | 완화책 |
|------|--------|------|--------|
| uvx 미설치 환경 | 높음 | 낮음 | 폴백 메커니즘 + 명확한 오류 |
| 캐시 문제 | 중간 | 낮음 | 캐시 정리 가이드 |
| 버전 호환성 | 중간 | 중간 | 명확한 버전 요구사항 표시 |
| 템플릿 미동기 | 낮음 | 중간 | 마이그레이션 가이드 제공 |

---

## 수행 검사목록

- [ ] M1: `pyproject.toml`에 새로운 진입점 추가
- [ ] M1: `moai_adk/cli.py` (또는 `__main__.py`) 서브커맨드 구현
- [ ] M2: `.claude/settings.json`의 statusLine 필드 수정
- [ ] M3: `src/moai_adk/templates/.claude/settings.json` 동기화
- [ ] M4: 테스트 코드 작성 (최소 5개 테스트)
- [ ] M4: 모든 OS에서 테스트 실행
- [ ] M4: 테스트 커버리지 ≥ 85% 확인
- [ ] M5 (선택): 폴백 메커니즘 구현 (필요 시)
- [ ] 모든 코드 변경사항 검토
- [ ] PyPI 배포 (새 버전)
- [ ] GitHub Release 작성
- [ ] 마이그레이션 가이드 문서 작성

---

## 예상 산출물

### 코드 수정 파일
- `pyproject.toml` (스크립트 진입점 추가)
- `moai_adk/__main__.py` 또는 `moai_adk/cli.py` (서브커맨드 추가)
- `.claude/settings.json` (statusLine 명령어 업데이트)
- `src/moai_adk/templates/.claude/settings.json` (템플릿 동기화)

### 테스트 파일
- `tests/statusline/test_uvx_entry_point.py` (새 파일)
- `tests/test_cli.py` (기존 파일에 추가)

### 문서
- `.moai/docs/migration-uvx.md` (마이그레이션 가이드)
- `README.md` 업데이트 (uvx 정보 추가)
- GitHub Release Notes

---

