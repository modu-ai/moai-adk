# パフォーマンス最適化上級ガイド

MoAI-ADKプロジェクトのパフォーマンスを最大化する実践的なテクニックです。

## パフォーマンス最適化の原則

1. **測定優先**: 推測せず、プロファイリングでボトルネックを特定
2. **大きいものから**: 影響の大きいものから最適化
3. **段階的改善**: 一度に一つずつ変更して測定
4. **テスト維持**: 最適化後もテストが通ることを保証

## ボトルネックの特定

### Pythonパフォーマンス分析

```bash
# CPUプロファイリング
python -m cProfile -s cumtime app.py | head -20

# メモリプロファイリング
python -m memory_profiler app.py

# Line profiler
kernprof -l -v app.py
```

### 結果の解釈

```
Function              Calls  Time    Time(%)
expensive_func        100    5.234   85%  ← ボトルネック！
normal_func          1000    0.523   8%
helper_func          5000    0.443   7%
```

## 最適化テクニック

### 1. アルゴリズムの改善

```python
# ❌ O(n²)アルゴリズム
def find_duplicates(numbers):
    for i in range(len(numbers)):
        for j in range(i+1, len(numbers)):
            if numbers[i] == numbers[j]:
                return True

# ✅ O(n)アルゴリズム
def find_duplicates(numbers):
    seen = set()
    for num in numbers:
        if num in seen:
            return True
        seen.add(num)
```

### 2. キャッシング

```python
from functools import lru_cache

# ❌ 遅いバージョン（毎回計算）
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

# ✅ 速いバージョン（キャッシュ）
@lru_cache(maxsize=128)
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)
```

### 3. データベース最適化

```python
# ❌ N+1クエリ問題
for user in users:
    posts = db.query(Post).filter(Post.user_id == user.id).all()

# ✅ JOINで一度に
users_with_posts = db.query(User).joinedload(User.posts).all()
```

### 4. 非同期処理

```python
# ❌ 同期（7秒）
for url in urls:
    response = requests.get(url)
    process(response)

# ✅ 非同期（1秒）
import asyncio
tasks = [fetch_and_process(url) for url in urls]
await asyncio.gather(*tasks)
```

## 一般的な最適化

| 問題          | 解決策             | パフォーマンス向上       |
| ------------- | ------------------ | --------------- |
| 反復的な計算 | @lru_cache         | 10-100倍        |
| N+1クエリ      | JOIN/eager loading | 100倍           |
| 同期I/O      | async/await        | 10倍            |
| 大きなリスト     | ジェネレータ         | メモリ50%削減 |
| 反復検索     | Set/Dict           | O(n²) → O(n)    |

## パフォーマンスモニタリング

### リアルタイムモニタリング

```bash
# CPU/メモリモニタリング
top -p $(pgrep -f app.py)

# プロセス状態
ps aux | grep python

# ポート使用状況
lsof -i :8000
```

### APM (Application Performance Monitoring)

```python
# Prometheusメトリクス収集
from prometheus_client import Counter, Histogram

request_count = Counter('requests', 'Total requests')
request_time = Histogram('request_duration', 'Request duration')

@request_time.time()
def handle_request():
    request_count.inc()
    # ... 処理
```

______________________________________________________________________

**次**: [セキュリティ上級ガイド](security.md) または [拡張とカスタマイズ](extensions.md)



