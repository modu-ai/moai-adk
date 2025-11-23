# Claude Code Memory System - 실전 예제

## 예제 1: 세션 메모리 저장

```python
# memory/session_store.py
from moai_cc_memory import MemoryStore, MemoryKey

class SessionMemory(MemoryStore):
    """사용자 세션 메모리"""
    
    async def save_user_context(self, user_id: str, context: dict):
        """사용자 컨텍스트 저장"""
        
        key = MemoryKey(
            scope="session",
            user_id=user_id,
            component="context"
        )
        
        await self.set(key, context, ttl=3600)  # 1시간
    
    async def load_user_context(self, user_id: str) -> dict:
        """사용자 컨텍스트 로드"""
        
        key = MemoryKey(
            scope="session",
            user_id=user_id,
            component="context"
        )
        
        return await self.get(key)
```

## 예제 2: 토큰 예산 추적

```python
# memory/token_budget.py
from moai_cc_memory import TokenBudgetTracker

class ContextualTokenBudget(TokenBudgetTracker):
    """상황별 토큰 예산 추적"""
    
    async def track_request(self, request_id: str, tokens_used: int):
        """요청별 토큰 추적"""
        
        # 예산 확인
        remaining = await self.get_remaining_budget()
        
        if tokens_used > remaining:
            raise BudgetExceededError(
                f"토큰 초과: {tokens_used} > {remaining}"
            )
        
        # 토큰 기록
        await self.record_usage(
            request_id=request_id,
            tokens=tokens_used,
            timestamp=datetime.now()
        )
```

## 예제 3: 대화 이력 관리

```python
# memory/conversation_history.py
from moai_cc_memory import ConversationMemory

class ContextAwareConversation(ConversationMemory):
    """맥락 인식 대화 메모리"""
    
    async def add_message(self, role: str, content: str, context: dict = None):
        """메시지 추가"""
        
        message = {
            "role": role,
            "content": content,
            "context": context or {},
            "timestamp": datetime.now()
        }
        
        await self.append(message)
        
        # 길이 제한 (최근 50개만 유지)
        if len(self.history) > 50:
            await self.trim_old_messages(keep=50)
    
    async def get_context_summary(self) -> str:
        """현재 맥락 요약"""
        
        # 최근 메시지들로부터 맥락 추출
        recent = self.history[-10:]
        
        summary = "\n".join([
            f"{msg['role']}: {msg['content'][:100]}"
            for msg in recent
        ])
        
        return summary
```

## 예제 4: 캐시 레이어

```python
# memory/cache_layer.py
from moai_cc_memory import MemoryCache

class LRUMemoryCache(MemoryCache):
    """LRU 캐시 구현"""
    
    def __init__(self, max_size: int = 1000):
        super().__init__()
        self.max_size = max_size
        self.access_times = {}
    
    async def get(self, key: str):
        """캐시에서 조회"""
        
        value = await super().get(key)
        
        if value is not None:
            # 접근 시간 업데이트
            self.access_times[key] = datetime.now()
        
        return value
    
    async def set(self, key: str, value):
        """캐시에 저장"""
        
        # 크기 초과 시 가장 오래된 항목 제거
        if len(self._data) >= self.max_size:
            oldest_key = min(
                self.access_times,
                key=self.access_times.get
            )
            await self.delete(oldest_key)
        
        await super().set(key, value)
        self.access_times[key] = datetime.now()
```

## 예제 5: 상태 머신 메모리

```python
# memory/state_machine.py
from moai_cc_memory import StatefulMemory

class WorkflowState(StatefulMemory):
    """워크플로우 상태 메모리"""
    
    async def transition(self, state: str, data: dict = None):
        """상태 전환"""
        
        # 유효한 전환 확인
        if not self.is_valid_transition(self.current_state, state):
            raise InvalidTransitionError(
                f"{self.current_state} -> {state} 불가"
            )
        
        # 상태 저장
        await self.set_state(state, data or {})
        
        # 전환 이력 기록
        await self.record_transition(
            from_state=self.current_state,
            to_state=state,
            timestamp=datetime.now()
        )
```

## 예제 6: 분산 메모리 동기화

```python
# memory/distributed_sync.py
from moai_cc_memory import DistributedMemory

class RedisBackedMemory(DistributedMemory):
    """Redis 기반 분산 메모리"""
    
    async def publish_state(self, channel: str, state: dict):
        """상태 발행"""
        
        await self.redis.publish(
            channel,
            json.dumps(state)
        )
    
    async def subscribe_to_state(self, channel: str, callback):
        """상태 구독"""
        
        async def on_message(message):
            state = json.loads(message['data'])
            await callback(state)
        
        await self.redis.subscribe(channel, callback=on_message)
```

---

**Last Updated**: 2025-11-22  
**Status**: Production Ready
