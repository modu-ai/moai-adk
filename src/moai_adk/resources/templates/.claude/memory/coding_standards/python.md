# Python 규칙

## ✅ 필수
- 코드 포맷/정적 분석: black, ruff, mypy(strict) 적용; CI에서 동일 설정 사용
- 타입 힌트 100% 작성, 외부 인터페이스는 Protocol/TypedDict 등으로 명시화
- 예외는 구체 타입을 사용하고 친절한 메시지를 제공(스택 노출 금지)
- 계층/모듈 경계를 지키고 순환 의존을 금지(env 설정은 pydantic 등으로 관리)
- 테스트는 pytest 기반, 커버리지 80% 이상, @.claude/memory/shared_checklists.md 참조

## 👍 권장
- asyncio 우선 사용, contextvars/timeout/contextmanager 로 자원 누수 방지
- 의존성 주입/팩토리/Adapter 패턴으로 테스트 용이성 확보
- pytest parametrize/fixture 재사용을 통해 케이스 추가, marker로 속도 레벨 분리
- 데이터 클래스/attrs/pydantic으로 불변 데이터 모델링

## 🚀 확장/고급
- Python 3.11+ ExceptionGroup/TaskGroup 등 최신 기능 활용
- mypy plugin/ruff rule customization으로 도메인 규칙 확장
- Hypothesis/Property 기반 테스트, coverage threshold 차등화
- uv/poetry 등 패키지 관리 고도화와 pre-commit hook(black/ruff/mypy/pytest) 구성
