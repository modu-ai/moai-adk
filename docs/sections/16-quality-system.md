# Quality System - Python 코드 품질 개선 시스템

> **@FEATURE:QUALITY-GUIDELINES** Python 코드 TRUST 원칙 자동 검증 엔진

## 개요

MoAI-ADK v0.2.2에서 완성된 Python 코드 품질 개선 시스템은 TRUST 5원칙을 기반으로 코드 품질을 실시간으로 검증하고 개선하는 완전 자동화된 시스템입니다.

### 핵심 기능

- **실시간 품질 게이트**: 코드 작성 중 TRUST 원칙 위반 자동 감지
- **TDD 지원**: Red-Green-Refactor 사이클 자동화
- **성능 최적화**: AST 캐싱, 병렬 처리로 대용량 코드베이스 지원
- **설정 가능**: 프로젝트별 품질 기준 커스터마이징

## GuidelineChecker API 레퍼런스

### 클래스 초기화

```python
from pathlib import Path
from moai_adk.core.quality.guideline_checker import GuidelineChecker, GuidelineLimits

# 기본 설정으로 초기화
checker = GuidelineChecker(project_path=Path("/path/to/project"))

# 커스텀 제한값으로 초기화
custom_limits = GuidelineLimits(
    MAX_FUNCTION_LINES=60,  # 기본: 50
    MAX_FILE_LINES=400,     # 기본: 300
    MAX_PARAMETERS=7,       # 기본: 5
    MAX_COMPLEXITY=15       # 기본: 10
)
checker = GuidelineChecker(project_path, limits=custom_limits)

# 설정 파일에서 로드
checker = GuidelineChecker(
    project_path,
    config_file=Path("quality-config.yaml")
)
```

### 개별 검증 메서드

#### 함수 길이 검증

```python
violations = checker.check_function_length(file_path)

# 반환 형식
[
    {
        "function_name": "long_function",
        "line_count": 75,
        "start_line": 10,
        "file_path": "/path/to/file.py",
        "max_allowed": 50
    }
]
```

#### 파일 크기 검증

```python
result = checker.check_file_size(file_path)

# 반환 형식
{
    "file_path": "/path/to/file.py",
    "line_count": 350,
    "violation": True,
    "max_allowed": 300
}
```

#### 매개변수 개수 검증

```python
violations = checker.check_parameter_count(file_path)

# 반환 형식
[
    {
        "function_name": "func_with_many_params",
        "parameter_count": 8,
        "line_number": 25,
        "file_path": "/path/to/file.py",
        "max_allowed": 5
    }
]
```

#### 복잡도 검증

```python
violations = checker.check_complexity(file_path)

# 반환 형식
[
    {
        "function_name": "complex_function",
        "complexity": 15,
        "line_number": 45,
        "file_path": "/path/to/file.py",
        "max_allowed": 10
    }
]
```

### 프로젝트 전체 스캔

```python
# 순차 처리
violations = checker.scan_project(parallel=False)

# 병렬 처리 (권장)
violations = checker.scan_project(parallel=True, max_workers=4)

# 반환 형식
{
    "function_length": [...],
    "file_size": [...],
    "parameter_count": [...],
    "complexity": [...]
}
```

### 종합 리포트 생성

```python
report = checker.generate_violation_report(parallel=True)

# 리포트 구조
{
    "violations": {...},
    "summary": {
        "total_violations": 15,
        "files_checked": 42,
        "files_with_violations": 8,
        "compliant": False,
        "compliance_rate": 80.95,
        "violation_breakdown": {
            "function_length": 5,
            "file_size": 3,
            "parameter_count": 4,
            "complexity": 3
        }
    },
    "performance": {
        "scan_duration_seconds": 2.341,
        "files_per_second": 17.94,
        "parallel_processing": True,
        "cache_stats": {
            "cache_hits": 28,
            "cache_misses": 14,
            "hit_rate": 0.667,
            "cache_size": 42
        }
    },
    "trust_guidelines": {
        "limits": {...},
        "worst_violations": {...}
    }
}
```

## 설정 관리

### 설정 파일 형식 (YAML)

```yaml
# quality-config.yaml
limits:
  max_function_lines: 60
  max_file_lines: 400
  max_parameters: 7
  max_complexity: 15
  min_docstring_length: 15
  max_nesting_depth: 5

file_patterns:
  include: ["*.py"]
  exclude: ["*test*", "*__pycache__*", "*.pyc"]

enabled_checks:
  function_length: true
  file_size: true
  parameter_count: true
  complexity: true

output_format: "json"
parallel_processing: true
max_workers: 4
```

### 설정 파일 형식 (JSON)

```json
{
  "limits": {
    "max_function_lines": 60,
    "max_file_lines": 400,
    "max_parameters": 7,
    "max_complexity": 15
  },
  "enabled_checks": {
    "function_length": true,
    "file_size": true,
    "parameter_count": true,
    "complexity": true
  },
  "parallel_processing": true
}
```

### 동적 설정 변경

```python
# 검사 활성화/비활성화
checker.set_enabled_checks({
    "function_length": True,
    "file_size": False,
    "parameter_count": True,
    "complexity": True
})

# 현재 설정 확인
enabled = checker.get_enabled_checks()

# 설정 내보내기
checker.export_config(Path("current-config.yaml"))

# 런타임 설정 업데이트
checker.update_config(
    max_function_lines=60,
    parallel_processing=False
)
```

## 사용 예시

### 기본 사용법

```python
from pathlib import Path
from moai_adk.core.quality.guideline_checker import GuidelineChecker

def validate_project_quality():
    """프로젝트 전체 품질 검증"""

    # 체커 초기화
    project_path = Path.cwd()
    checker = GuidelineChecker(project_path)

    # 전체 스캔 실행
    print("프로젝트 품질 검증 중...")
    report = checker.generate_violation_report(parallel=True)

    # 결과 출력
    summary = report["summary"]
    print(f"✅ 검사 완료: {summary['files_checked']}개 파일")
    print(f"📊 준수율: {summary['compliance_rate']:.1f}%")
    print(f"⚠️  총 위반: {summary['total_violations']}건")

    # 성능 정보
    perf = report["performance"]
    print(f"⚡ 스캔 시간: {perf['scan_duration_seconds']}초")
    print(f"💾 캐시 히트율: {perf['cache_stats']['hit_rate']*100:.1f}%")

    return summary["compliant"]

# 실행
if __name__ == "__main__":
    is_compliant = validate_project_quality()
    exit(0 if is_compliant else 1)
```

### 단일 파일 검증

```python
def validate_single_file(file_path: Path) -> bool:
    """단일 파일 품질 검증"""

    checker = GuidelineChecker(file_path.parent)

    try:
        # 모든 규칙 검증
        is_valid = checker.validate_single_file(file_path)

        if is_valid:
            print(f"✅ {file_path.name}: 모든 품질 규칙 통과")
        else:
            print(f"❌ {file_path.name}: 품질 규칙 위반 발견")

            # 상세 위반 내역 출력
            violations = checker.check_function_length(file_path)
            for v in violations:
                print(f"  ⚠️  함수 '{v['function_name']}': {v['line_count']}줄 (한계: {v['max_allowed']}줄)")

        return is_valid

    except Exception as e:
        print(f"❌ 검증 오류: {e}")
        return False
```

### CI/CD 통합

```python
def ci_quality_gate() -> bool:
    """CI/CD 파이프라인용 품질 게이트"""

    checker = GuidelineChecker(Path.cwd())

    # 빠른 검증 (병렬 처리)
    violations = checker.scan_project(parallel=True)

    # 위반 건수 확인
    total_violations = sum(len(v) for v in violations.values())

    if total_violations == 0:
        print("✅ 품질 게이트 통과: 모든 규칙 준수")
        return True
    else:
        print(f"❌ 품질 게이트 실패: {total_violations}건 위반")

        # 주요 위반 내역 출력
        for violation_type, violation_list in violations.items():
            if violation_list:
                print(f"  {violation_type}: {len(violation_list)}건")

        return False

# CI 스크립트에서 사용
if __name__ == "__main__":
    success = ci_quality_gate()
    exit(0 if success else 1)
```

### 사용자 정의 규칙

```python
def setup_custom_rules():
    """프로젝트별 사용자 정의 규칙 설정"""

    # 레거시 프로젝트용 완화된 규칙
    legacy_limits = GuidelineLimits(
        MAX_FUNCTION_LINES=100,  # 기존 코드 고려
        MAX_FILE_LINES=500,      # 대용량 모듈 허용
        MAX_PARAMETERS=8,        # API 호환성 유지
        MAX_COMPLEXITY=20        # 단계적 개선
    )

    checker = GuidelineChecker(
        project_path=Path.cwd(),
        limits=legacy_limits
    )

    # 일부 검사만 활성화
    checker.set_enabled_checks({
        "function_length": True,
        "file_size": False,      # 파일 크기 검사 비활성화
        "parameter_count": True,
        "complexity": True
    })

    return checker
```

## 성능 최적화

### 대용량 프로젝트 최적화

```python
def optimize_for_large_projects():
    """대용량 프로젝트 최적화 설정"""

    from moai_adk.core.quality.guideline_checker import GuidelineConfig

    config = GuidelineConfig.create_default()
    config.parallel_processing = True
    config.max_workers = 8  # CPU 코어 수에 맞게 조정

    checker = GuidelineChecker(
        project_path=Path.cwd(),
        config=config
    )

    # 캐시 통계 모니터링
    report = checker.generate_violation_report()
    cache_stats = report["performance"]["cache_stats"]

    print(f"캐시 효율성: {cache_stats['hit_rate']*100:.1f}%")

    # 메모리 정리 (필요시)
    if cache_stats["cache_size"] > 1000:
        checker.clear_cache()
        print("캐시 정리 완료")
```

## Claude Code 통합

### 실시간 품질 검증 훅

```python
# .claude/hooks/moai/quality_guard.py
from pathlib import Path
from moai_adk.core.quality.guideline_checker import GuidelineChecker

def pre_write_quality_check(file_path: str, content: str) -> bool:
    """파일 저장 전 품질 검증"""

    if not file_path.endswith('.py'):
        return True

    # 임시 파일로 저장하여 검증
    temp_path = Path(file_path)
    temp_path.write_text(content)

    try:
        checker = GuidelineChecker(temp_path.parent)
        is_valid = checker.validate_single_file(temp_path)

        if not is_valid:
            print("⚠️  품질 규칙 위반이 감지되었습니다.")
            # 사용자에게 계속 진행할지 확인
            return False

        return True

    finally:
        if temp_path.exists():
            temp_path.unlink()
```

## 문제 해결

### 일반적인 문제들

#### 파싱 오류

```python
# 문법 오류가 있는 파일 처리
violations = checker.scan_project()
# 파싱 실패한 파일은 자동으로 건너뜀
```

#### 성능 문제

```python
# 병렬 처리 비활성화
checker.update_config(parallel_processing=False)

# 캐시 정리
checker.clear_cache()

# 검사 범위 제한
checker.set_enabled_checks({
    "function_length": True,
    "file_size": False,
    "parameter_count": False,
    "complexity": True
})
```

#### 메모리 사용량

```python
# 주기적인 캐시 정리
cache_stats = checker.get_cache_stats()
if cache_stats["cache_size"] > 500:
    checker.clear_cache()
```

## 확장성

### 사용자 정의 검증 규칙

GuidelineChecker는 확장 가능한 아키텍처로 설계되어 있어, 새로운 품질 규칙을 추가할 수 있습니다:

```python
class CustomGuidelineChecker(GuidelineChecker):
    """사용자 정의 품질 검증기"""

    def check_naming_conventions(self, file_path: Path) -> List[Dict]:
        """네이밍 컨벤션 검증"""
        violations = []
        tree = self._parse_python_file(file_path)

        for node in ast.walk(tree):
            if isinstance(node, ast.FunctionDef):
                if not node.name.islower():
                    violations.append({
                        "function_name": node.name,
                        "violation": "함수명은 소문자여야 함",
                        "line_number": node.lineno
                    })

        return violations
```

---

**관련 문서**

- [TRUST 5원칙](03-principles.md)
- [아키텍처 가이드](04-architecture.md)
- [테스트 가이드](docs/test-guide.md)

**@TAG 추적성**

- @REQ:QUALITY-002 → @DESIGN:QUALITY-SYSTEM-002 → @TASK:IMPLEMENT-002 → @TEST:ACCEPTANCE-002
