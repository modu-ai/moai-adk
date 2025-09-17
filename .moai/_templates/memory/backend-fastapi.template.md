# $PROJECT_NAME 백엔드(FastAPI) 메모

> FastAPI 기반 서비스 구현 시 필요한 설정과 체크 항목입니다. Python 공통 메모(`backend-python.md`)와 함께 사용하세요.

## 1. 서비스 구성
- 앱 구조: `app/main.py`, `app/api`, `app/services`, `app/models`, `app/core`
- 의존성 주입: `fastapi.Depends`, DI 컨테이너 사용 시 `punq`, `wired` 등 활용
- 설정 관리: `pydantic-settings`/`BaseSettings`로 환경변수 검증, `settings.py` 모듈화
- 웹서버: 개발 `uvicorn app.main:app --reload`, 운영 `uvicorn ... --workers $(nproc)` 또는 Gunicorn/uvicorn worker

## 2. 검증 & 문서화
- 스키마: Pydantic v2, `Annotated`/`Field` 활용, request/response 모델 분리
- 라우팅: prefix + tags 명시, 버저닝(`/api/v1/`) 적용
- 문서: OpenAPI 자동 문서 + `fastapi.openapi.utils.get_openapi` 커스터마이징

## 3. 테스트
- pytest + `TestClient`, 비동기 테스트는 `pytest-asyncio`
- DB 세션은 테스트 픽스처로 격리(Session `override_get_db`), Testcontainers 권장
- 계약 테스트: `schemathesis`/`pytest`로 OpenAPI contract 검증

## 4. 미들웨어 & 관측성
- CORS, GZip, TrustedHost, Rate limiting(FastAPI-Limiter) 구성
- 로깅: `structlog` 또는 `loguru` + request id, ASGI middleware로 tracing(OpenTelemetry) 연동
- 헬스체크/ready 엔드포인트 제공, Prometheus exporter(`prometheus_fastapi_instrumentator`) 사용

## 5. 배포 체크
- Docker 이미지: slim python base, multi-stage (builder + runtime), non-root user
- CI 파이프라인: lint(ruff/black), mypy, pytest, coverage 리포트, `uv pip compile`로 lock 갱신
- Zero-downtime: rolling update, DB migration(`alembic upgrade head`) → healthcheck 순서 정의

## 6. 참고 링크
- 공통 Python 메모: `./backend-python.md`
- FastAPI 공식: https://fastapi.tiangolo.com/
- 보안/TDD/체크리스트: @.claude/memory/security_rules.md, @.claude/memory/tdd_guidelines.md, @.claude/memory/shared_checklists.md
