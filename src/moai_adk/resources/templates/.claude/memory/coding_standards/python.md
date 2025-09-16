# Python 규칙(요약)

- 도구: black, ruff, mypy(strict), pytest, coverage
- 타입: type hints 필수, Protocol/TypedDict로 명세적 인터페이스
- 예외: 구체 타입 사용, 메시지 친절, 예외 그룹 사용 가능(Py3.11+)
- 동시성: asyncio 우선, contextvars/timeout 관리, 리소스 정리(contextmanager)
- 구조: 계층/모듈 경계 유지, 순환 의존 금지, 설정은 env+pydantic
- 테스트: pytest parametrize, fixture 최소, 속도 고려(marker)
