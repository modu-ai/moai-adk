Translate the following Korean markdown document to Japanese.

**CRITICAL RULES:**
1. Preserve ALL markdown structure (headers, code blocks, links, tables, diagrams)
2. Keep ALL code blocks and technical terms UNCHANGED
3. Maintain the EXACT same file structure and formatting
4. Translate ONLY Korean text content
5. Keep ALL @TAG references unchanged (e.g., @SPEC:AUTH-001)
6. Preserve ALL file paths and URLs
7. Keep ALL emoji and icons as-is
8. Maintain ALL frontmatter (YAML) structure

**Source File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ko/advanced/performance.md
**Target Language:** Japanese
**Target File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ja/advanced/performance.md

**Content to Translate:**

# 성능 최적화 고급 가이드

MoAI-ADK 프로젝트의 성능을 극대화하는 실전 기법들입니다.

## :bullseye: 성능 최적화 원칙

1. **측정 먼저**: 추측하지 말고 프로파일링으로 병목 확인
2. **큰 것부터**: 영향도가 큰 것부터 최적화
3. **점진적 개선**: 한 번에 한 가지만 변경 후 측정
4. **테스트 유지**: 최적화 후에도 테스트 통과 보장

## 📊 병목 지점 식별

### Python 성능 분석

```bash
# CPU 프로파일링
python -m cProfile -s cumtime app.py | head -20

# 메모리 프로파일링
python -m memory_profiler app.py

# Line profiler
kernprof -l -v app.py
```

### 결과 해석

```
Function              Calls  Time    Time(%)
expensive_func        100    5.234   85%  ← 병목!
normal_func          1000    0.523   8%
helper_func          5000    0.443   7%
```

## 🚀 최적화 기법

### 1. 알고리즘 개선

```python
# :x: O(n²) 알고리즘
def find_duplicates(numbers):
    for i in range(len(numbers)):
        for j in range(i+1, len(numbers)):
            if numbers[i] == numbers[j]:
                return True

# ✅ O(n) 알고리즘
def find_duplicates(numbers):
    seen = set()
    for num in numbers:
        if num in seen:
            return True
        seen.add(num)
```

### 2. 캐싱

```python
from functools import lru_cache

# :x: 느린 버전 (매번 계산)
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

# ✅ 빠른 버전 (캐싱)
@lru_cache(maxsize=128)
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)
```

### 3. 데이터베이스 최적화

```python
# :x: N+1 쿼리 문제
for user in users:
    posts = db.query(Post).filter(Post.user_id == user.id).all()

# ✅ JOIN으로 한 번에
users_with_posts = db.query(User).joinedload(User.posts).all()
```

### 4. 비동기 처리

```python
# :x: 동기 (7초)
for url in urls:
    response = requests.get(url)
    process(response)

# ✅ 비동기 (1초)
import asyncio
tasks = [fetch_and_process(url) for url in urls]
await asyncio.gather(*tasks)
```

## 🔧 일반적인 최적화

| 문제          | 해결책             | 성능 향상       |
| ------------- | ------------------ | --------------- |
| 반복적인 계산 | @lru_cache         | 10-100배        |
| N+1 쿼리      | JOIN/eager loading | 100배           |
| 동기 I/O      | async/await        | 10배            |
| 큰 리스트     | 제너레이터         | 메모리 50% 감소 |
| 반복 검색     | Set/Dict           | O(n²) → O(n)    |

## 📈 성능 모니터링

### 실시간 모니터링

```bash
# CPU/메모리 모니터링
top -p $(pgrep -f app.py)

# 프로세스 상태
ps aux | grep python

# 포트 사용 현황
lsof -i :8000
```

### APM (Application Performance Monitoring)

```python
# Prometheus 메트릭 수집
from prometheus_client import Counter, Histogram

request_count = Counter('requests', 'Total requests')
request_time = Histogram('request_duration', 'Request duration')

@request_time.time()
def handle_request():
    request_count.inc()
    # ... 처리
```

______________________________________________________________________

**다음**: [보안 고급 가이드](security.md) 또는 [확장 및 커스터마이제이션](extensions.md)


**Instructions:**
- Translate the content above to Japanese
- Output ONLY the translated markdown content
- Do NOT include any explanations or comments
- Maintain EXACT markdown formatting
- Preserve ALL code blocks exactly as-is
