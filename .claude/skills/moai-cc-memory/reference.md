# Claude Code Memory - 참고 자료

## 메모리 계층 구조

```
┌─────────────────────────────────────────┐
│   Application Layer                     │
│   (Session, State Management)           │
├─────────────────────────────────────────┤
│   Memory API Layer                      │
│   (MemoryStore, MemoryCache)            │
├─────────────────────────────────────────┤
│   Backend Storage                       │
│   (Redis, Memory, Disk)                 │
└─────────────────────────────────────────┘
```

## API Reference

### MemoryStore 인터페이스

| 메서드 | 설명 | 반환값 |
|--------|------|--------|
| `get(key)` | 값 조회 | `T \| None` |
| `set(key, value, ttl)` | 값 저장 | `None` |
| `delete(key)` | 값 삭제 | `bool` |
| `exists(key)` | 존재 확인 | `bool` |

### TokenBudgetTracker

```python
tracker = TokenBudgetTracker(max_tokens=200000)
remaining = await tracker.get_remaining_budget()
await tracker.record_usage(request_id, tokens)
```

## 메모리 키 규약

```python
# 키 형식: scope:user_id:component:detail
MemoryKey(
    scope="session",     # 범위
    user_id="user123",   # 사용자
    component="context", # 컴포넌트
    detail="state"       # 상세
)
```

## 사용 사례

| 사용 사례 | 메모리 유형 | TTL |
|---------|-----------|-----|
| 세션 데이터 | MemoryStore | 1-24시간 |
| 토큰 추적 | TokenBudgetTracker | 세션 기간 |
| 캐시 | MemoryCache | 5분-1시간 |
| 상태 | StatefulMemory | 워크플로우 기간 |

---

**Last Updated**: 2025-11-22
