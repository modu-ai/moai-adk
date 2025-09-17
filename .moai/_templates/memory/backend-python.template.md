# $PROJECT_NAME 백엔드(Python) 메모

> Python 기반 백엔드 작업 시 따라야 할 환경 구성과 품질 체크 항목을 정리했습니다.

## 1. 환경 구성
- uv 권장: `pipx install uv` → `uv init --package $PROJECT_NAME` 또는 `uv venv`
- poetry 대안: `pipx install poetry` → `poetry init` → `poetry config virtualenvs.in-project true`
- 공통 의존성: `ruff`, `black`, `mypy`, `pytest`, `pytest-cov`
- 환경 변수는 `.env` → `.env.example` 패턴으로 관리하고, 민감정보는 절대 커밋하지 않는다.

## 2. pre-commit Hook
```bash
pre-commit install
cat <<'EOF' > .pre-commit-config.yaml
repos:
  - repo: https://github.com/charliermarsh/ruff-pre-commit
    rev: v0.6.8
    hooks:
      - id: ruff
      - id: ruff-format
  - repo: https://github.com/psf/black
    rev: 24.8.0
    hooks:
      - id: black
  - repo: https://github.com/pre-commit/mirrors-mypy
    rev: v1.10.0
    hooks:
      - id: mypy
  - repo: https://github.com/pytest-dev/pytest
    rev: 7.4.4
    hooks:
      - id: pytest
        additional_dependencies: [pytest-cov]
EOF
```
- CI에서도 동일한 lint/test 조합을 실행하여 로컬·원격 결과 일관성을 유지합니다.

## 3. 테스트 & 품질
- 테스트는 pytest, `tests/` 디렉터리 구조 유지, 커버리지 목표 80% 이상 (@.claude/memory/tdd_guidelines.md)
- typing은 필수, interface는 `Protocol`/`TypedDict`로 명시, 런타임 검증은 pydantic/attrs 활용
- 로깅은 구조화(json) + 상관관계 ID 전파(@.claude/memory/security_rules.md)
- long-running 작업은 asyncio/taskgroup, context cancellation 처리

## 4. 배포 & 운영 체크
- 패키지 빌드: `uv build` 또는 `poetry build`
- 컨테이너 이미지 작성 시 멀티스테이지 + slim 베이스 권장, health check 라우트 구현
- 관측성: OpenTelemetry + OTLP, Prometheus exporter, Sentry/Log aggregation 연동

## 5. 참고 문서
- 코딩 표준(세부): @.claude/memory/coding_standards/python.md
- 공통 체크리스트: @.claude/memory/shared_checklists.md
- 보안 규칙: @.claude/memory/security_rules.md
- 운영 가이드: @.claude/memory/project_guidelines.md
