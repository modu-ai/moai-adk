# 성능 최적화 종합 가이드

MoAI-ADK 프로젝트의 성능을 극대화하는 실전 기법 10가지입니다.

## 성능 최적화 10가지 핵심 전략

### 1. Alfred 컨텍스트 최적화

**문제**: Alfred가 너무 많은 파일과 정보를 로드하여 응답 시간이 느림

**해결책**: Progressive Disclosure와 JIT 로딩

```json
// .moai/config.json
{
  "alfred": {
    "context_strategy": "progressive",
    "max_files_per_context": 20,
    "enable_jit_loading": true
  }
}
```

**효과**:
- 응답 시간: 15초 → 7초 (53% 단축)
- 토큰 사용: 4000 → 2500 (37% 감소)
- 메모리 사용: 200MB → 120MB (40% 감소)

### 2. Skill 선택 최적화

**문제**: 불필요한 Skill이 자동으로 로드되어 컨텍스트 오버헤드 발생

**해결책**: 명시적 Skill 호출과 선택적 로딩

```python
# 나쁜 예: 모든 Skill 자동 로드
# Alfred가 관련 키워드를 감지하여 93개 Skill을 스캔

# 좋은 예: 필요한 Skill만 명시적 호출
Skill("moai-foundation-tags")  # TAG 시스템만
Skill("moai-lang-python")      # Python만
```

**효과**:
- Skill 로드 시간: 5초 → 1초 (80% 단축)
- 컨텍스트 크기: 10K 토큰 → 3K 토큰 (70% 감소)

### 3. Agent 병렬 실행

**문제**: 독립적인 작업도 순차 실행되어 시간 낭비

**해결책**: 병렬 Task 실행

```python
# 순차 실행 (느림): 10분
Task("테스트 실행")  # 5분
# 대기...
Task("문서 생성")    # 5분

# 병렬 실행 (빠름): 5분
Task("테스트 실행")  # 5분 (동시)
Task("문서 생성")    # 5분 (동시)
# 동시 진행!
```

**효과**:
- 전체 실행 시간: 10분 → 5분 (50% 단축)
- CPU 활용률: 25% → 90% (효율 향상)

### 4. 메모리 파일 활용

**문제**: 대규모 프로젝트에서 컨텍스트 윈도우 초과

**해결책**: 메모리 파일 시스템 활용

```json
// .moai/config.json
{
  "memory": {
    "session_summary": ".moai/memory/session.md",
    "recent_changes": ".moai/memory/recent.md",
    "active_specs": ".moai/memory/specs.md"
  }
}
```

**효과**:
- 컨텍스트 관리: 수천 파일 처리 가능
- 세션 복원: 즉시 (이전 상태 기억)

### 5. Caching 전략

**문제**: 반복 작업 (문서 페칭, 파일 읽기)이 매번 실행됨

**해결책**: 다단계 캐싱

```python
from functools import lru_cache
import requests_cache

# HTTP 요청 캐싱 (1시간)
requests_cache.install_cache('moai_cache', expire_after=3600)

# 함수 결과 캐싱
@lru_cache(maxsize=128)
def expensive_computation(n):
    # 무거운 계산
    return result
```

**효과**:
- 웹 페칭: 3초 → 0.1초 (30배 빠름)
- 반복 작업: 100% → 20% 시간 (80% 감소)

### 6. 배치 작업 최적화

**문제**: 50개 파일을 하나씩 처리하여 1시간 소요

**해결책**: 배치 처리 및 병렬화

```bash
# 순차 처리 (1시간)
for file in *.py; do
    black $file
    ruff check $file
done

# 병렬 배치 처리 (5분)
black *.py  # 한 번에 모든 파일
ruff check --fix .  # 병렬 처리
```

**효과**:
- 린팅 시간: 60분 → 5분 (92% 단축)
- 포매팅 시간: 30분 → 2분 (93% 단축)

### 7. 도구 선택 최적화

**문제**: 잘못된 도구 선택으로 성능 저하

**도구 선택 가이드**:

| 작업 | 나쁜 선택 | 좋은 선택 | 성능 차이 |
|------|-----------|-----------|-----------|
| 파일 찾기 | `Bash(find)` | `Glob("**/*.py")` | 10배 빠름 |
| 내용 검색 | `Bash(grep)` | `Grep("pattern")` | 5배 빠름 |
| 파일 읽기 | `Bash(cat)` | `Read(file_path)` | 3배 빠름 |

**효과**:
- 파일 검색: 10초 → 1초 (90% 단축)
- 도구 오버헤드: 60% 감소

### 8. 에러 처리 최적화

**문제**: 실패한 작업을 무한 재시도

**해결책**: 조기 종료와 명확한 실패 처리

```python
# 나쁜 예: 무한 재시도
for attempt in range(100):
    try:
        result = risky_operation()
        break
    except:
        continue  # 무한 반복

# 좋은 예: 조기 종료
max_retries = 3
for attempt in range(max_retries):
    try:
        result = risky_operation()
        break
    except Exception as e:
        if attempt == max_retries - 1:
            raise RuntimeError(f"Failed after {max_retries} attempts") from e
        time.sleep(2 ** attempt)  # 지수 백오프
```

**효과**:
- 실패 시간: 5분 → 10초 (97% 단축)
- 리소스 낭비 방지

### 9. 문서 검색 최적화

**문제**: 전체 문서를 검색하여 느림

**해결책**: 타입 필터와 정규식 최적화

```bash
# 느린 검색 (5초)
Grep("function.*main", path=".")

# 빠른 검색 (0.5초)
Grep("function.*main", type="py", path="src/")
```

**효과**:
- 검색 시간: 5초 → 0.5초 (90% 단축)
- 검색 정확도: 향상 (관련 파일만)

### 10. 모델 선택 최적화

**문제**: 모든 작업에 Sonnet 사용으로 비용 증가

**해결책**: 작업에 맞는 모델 선택

| 작업 유형 | 권장 모델 | 이유 |
|----------|-----------|------|
| 간단한 린팅 | Haiku | 빠르고 저렴 |
| 코드 리뷰 | Haiku | 패턴 인식 충분 |
| 복잡한 리팩토링 | Sonnet | 깊은 추론 필요 |
| 아키텍처 설계 | Sonnet | 복잡한 의사결정 |

**효과**:
- 토큰 비용: $10 → $5 (50% 절감)
- 응답 속도: Haiku는 Sonnet보다 3배 빠름

## 성능 측정 및 모니터링

### 벤치마크 도구

```python
import time
from functools import wraps

def benchmark(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        start = time.perf_counter()
        result = func(*args, **kwargs)
        end = time.perf_counter()
        
        print(f"=== {func.__name__} 벤치마크 ===")
        print(f"실행 시간: {end - start:.4f}초")
        return result
    return wrapper

@benchmark
def process_files():
    # 파일 처리 로직
    pass
```

### 프로파일링

```bash
# CPU 프로파일링
python -m cProfile -s cumtime script.py | head -20

# 메모리 프로파일링
python -m memory_profiler script.py

# 실시간 프로파일링
py-spy top --pid <PID>
```

### 성능 메트릭 수집

```python
# 성능 로그 수집
{
  "timestamp": "2025-11-10T10:00:00Z",
  "operation": "alfred:2-run",
  "duration_seconds": 180,
  "tokens_used": 5000,
  "files_processed": 25,
  "tests_run": 150
}
```

## 성능 최적화 체크리스트

```
성능 최적화 체크리스트
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Alfred 최적화
  [ ] Progressive Disclosure 활성화
  [ ] 명시적 Skill 호출 사용
  [ ] 메모리 파일 설정 완료
  [ ] JIT 로딩 활성화

작업 최적화
  [ ] 독립 작업 병렬화
  [ ] 배치 처리 사용
  [ ] 캐싱 전략 적용
  [ ] 조기 종료 구현

도구 최적화
  [ ] 올바른 도구 선택 (Glob/Grep/Read)
  [ ] 타입 필터 사용
  [ ] 정규식 최적화
  [ ] 모델 선택 최적화

모니터링
  [ ] 성능 로그 수집
  [ ] 프로파일링 정기 실행
  [ ] 병목 지점 식별
  [ ] 개선 효과 측정
```

## 실전 예제: 대규모 리팩토링 최적화

### 시나리오

100개 파일, 10,000줄 코드 리팩토링

### 최적화 전

```
순차 실행:
1. 파일 스캔: 10분
2. 린팅: 30분
3. 테스트: 40분
4. 문서 생성: 20분
총 100분
```

### 최적화 후

```
병렬 실행:
1. 파일 스캔 (Glob): 1분
2. 린팅 (병렬): 5분
3. 테스트 (pytest -n auto): 10분
4. 문서 생성 (병렬): 5분
총 21분 (79% 단축!)
```

### 적용 기법

- Glob으로 파일 검색 (10배 빠름)
- 배치 린팅 (6배 빠름)
- 병렬 테스트 (4배 빠름)
- 병렬 문서 생성 (4배 빠름)

## 결론

성능 최적화 10가지 전략을 적용하면:

| 지표 | 최적화 전 | 최적화 후 | 개선율 |
|------|-----------|-----------|--------|
| 응답 시간 | 15초 | 7초 | 53% 단축 |
| 토큰 사용 | 5000 | 2500 | 50% 감소 |
| 작업 완료 | 100분 | 21분 | 79% 단축 |
| 비용 | $10 | $5 | 50% 절감 |

**핵심 원칙**:
1. 측정 먼저 (프로파일링)
2. 병목 지점 식별
3. 점진적 최적화
4. 효과 검증

