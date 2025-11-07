# 性能优化高级指南

最大化 MoAI-ADK 项目性能的实战技巧。

## 🎯 性能优化原则

1. **先测量**：不要猜测，用性能分析找出瓶颈
2. **从大到小**：先优化影响最大的部分
3. **逐步改进**：一次只改变一项，然后测量
4. **保持测试**：优化后仍要确保测试通过

## 📊 识别瓶颈

### Python 性能分析

```bash
# CPU 性能分析
python -m cProfile -s cumtime app.py | head -20

# 内存性能分析
python -m memory_profiler app.py

# Line profiler
kernprof -l -v app.py
```

### 结果解释

```
Function              Calls  Time    Time(%)
expensive_func        100    5.234   85%  ← 瓶颈！
normal_func          1000    0.523   8%
helper_func          5000    0.443   7%
```

## 🚀 优化技巧

### 1. 算法改进

```python
# ❌ O(n²) 算法
def find_duplicates(numbers):
    for i in range(len(numbers)):
        for j in range(i+1, len(numbers)):
            if numbers[i] == numbers[j]:
                return True

# ✅ O(n) 算法
def find_duplicates(numbers):
    seen = set()
    for num in numbers:
        if num in seen:
            return True
        seen.add(num)
```

### 2. 缓存

```python
from functools import lru_cache

# ❌ 慢速版本（每次计算）
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

# ✅ 快速版本（缓存）
@lru_cache(maxsize=128)
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)
```

### 3. 数据库优化

```python
# ❌ N+1 查询问题
for user in users:
    posts = db.query(Post).filter(Post.user_id == user.id).all()

# ✅ 使用 JOIN 一次查询
users_with_posts = db.query(User).joinedload(User.posts).all()
```

### 4. 异步处理

```python
# ❌ 同步（7秒）
for url in urls:
    response = requests.get(url)
    process(response)

# ✅ 异步（1秒）
import asyncio
tasks = [fetch_and_process(url) for url in urls]
await asyncio.gather(*tasks)
```

## 🔧 常见优化

| 问题     | 解决方案           | 性能提升        |
| -------- | ------------------ | --------------- |
| 重复计算 | @lru_cache         | 10-100倍        |
| N+1 查询 | JOIN/eager loading | 100倍           |
| 同步 I/O | async/await        | 10倍            |
| 大列表   | 生成器             | 内存减少 50%    |
| 重复搜索 | Set/Dict           | O(n²) → O(n)    |

## 📈 性能监控

### 实时监控

```bash
# CPU/内存监控
top -p $(pgrep -f app.py)

# 进程状态
ps aux | grep python

# 端口使用情况
lsof -i :8000
```

### APM (应用性能监控)

```python
# Prometheus 指标收集
from prometheus_client import Counter, Histogram

request_count = Counter('requests', 'Total requests')
request_time = Histogram('request_duration', 'Request duration')

@request_time.time()
def handle_request():
    request_count.inc()
    # ... 处理
```

______________________________________________________________________

**下一步**：[安全高级指南](security.md) 或 [扩展和自定义](extensions.md)
