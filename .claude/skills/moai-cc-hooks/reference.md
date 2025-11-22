# Claude Code Hooks - 참고 자료

## 훅 타입

| 훅 | 시점 | 용도 |
|-----|------|------|
| `pre-commit` | 커밋 전 | 코드 검증, 린팅 |
| `post-commit` | 커밋 후 | 로깅, 알림 |
| `pre-push` | 푸시 전 | 최종 검증 |
| `post-merge` | 병합 후 | 자동화 작업 |
| `pre-build` | 빌드 전 | 환경 준비 |
| `post-deploy` | 배포 후 | 알림, 모니터링 |

## Hook 인터페이스

```python
class Hook:
    event: str  # 훅 이벤트
    priority: int  # 실행 우선순위
    async_execution: bool  # 비동기 실행
    
    async def execute(context: dict) -> dict:
        """훅 실행"""
    
    async def on_error(error: Exception) -> dict:
        """에러 처리"""
```

## 컨텍스트 객체

```python
{
    "branch": "main",
    "files_changed": ["src/main.py", "tests/test_main.py"],
    "commit_message": "feat: add new feature",
    "author": "user@example.com",
    "timestamp": "2025-11-22T10:00:00Z"
}
```

## 반환값 형식

```python
{
    "success": True,        # 성공 여부
    "message": "...",       # 메시지
    "data": {...}          # 추가 데이터
}
```

---

**Last Updated**: 2025-11-22
