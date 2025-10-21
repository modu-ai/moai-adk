# Awesome 릴리즈 플레이북

## 워크플로 개요

- **품질 게이트**: 최신 `develop`에서 `pytest --cov`, `ruff check`, `mypy`, `bandit`, `pip-audit`를 실행한다. 테스트 실패는 릴리즈를 중단시키며, 커버리지가 85% 이하라면 사유를 기록한다.
- **브랜치 정렬**: 릴리즈 대상 커밋이 `develop`에 모두 포함됐는지 확인한다. `origin/develop`을 최신 상태로 맞춘 뒤 `develop` → `main` PR을 생성하고 머지한다. 릴리즈는 `main`의 머지 커밋을 기준으로 진행한다.
- **버전 결정**: 시맨틱 버전 증가폭을 `patch`, `minor`, `major` 중에서 선택한다. 버전 정보는 `pyproject.toml` 한 곳만 수정한다(SSOT).

## 릴리즈 실행

1. **시크릿 로드**  
   `.env`를 로드해 `PYPI_API_TOKEN`을 확보하고, PyPI 배포용으로 `UV_PUBLISH_TOKEN`에 할당한다.  
   ```bash
   source .env
   export UV_PUBLISH_TOKEN="$PYPI_API_TOKEN"
   ```
   GitHub CLI 권한도 확인한다: `gh auth status`.

2. **자동화 커맨드 실행**  
   `main` 브랜치에서 Awesome 릴리즈 커맨드를 실행한다.  
   ```bash
   /awesome:release-new <patch|minor|major>
   ```
   이 커맨드는 품질 검증, `pyproject.toml` 버전 갱신, 커밋·태그 생성, `uv publish`를 통한 PyPI 배포(환경 변수 `UV_PUBLISH_TOKEN` 사용), GitHub 릴리즈 드래프트 작성까지 자동으로 수행한다. 표시되는 계획을 확인하고 승인 후 진행한다.

3. **수동 플로우 대응**  
   자동화가 불가능할 경우 다음과 같이 수동으로 동일 작업을 수행한다.
   - `pyproject.toml` 버전을 목표 버전으로 수정한다.
   - 변경분을 커밋하고 `v<version>` 형식의 주석 태그를 만든다.
   - `.env`에서 불러온 `UV_PUBLISH_TOKEN`을 그대로 사용하여 PyPI에 배포한다.  
     ```bash
     uv publish --token "$UV_PUBLISH_TOKEN"
     ```
   - `main`과 태그를 원격 저장소로 푸시한다.

## GitHub 릴리즈 마무리

- 자동화 커맨드가 생성한 드래프트 릴리즈(태그 `v<version>` )를 열고, 주요 변경사항과 PR 링크를 정리한다.
- 검토 후 릴리즈를 게시한다. 필요 시 `gh release create v<version> --notes-file <notes.md>`로 수동 작성한다.

## 릴리즈 후 체크리스트

- 패키지 설치 확인: `uv pip install moai-adk==<version>`.
- PyPI 메타데이터와 업로드 파일 검사.
- 팀 공지(Slack·Notion 등) 및 문서 업데이트.
- 향후 개발을 위해 `main`에서 새로운 `develop` 브랜치를 생성한다.
