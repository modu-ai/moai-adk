# `project-detect`

operation `project.detect` · 코드 `src/ops/project_detect.rs` · #2496

`project-detect PROJECT` 는 `/moai project` 경로를 위해 프로젝트 사실(새 프로젝트인지, manifest, 언어, 언어 서버, 설정된 MCP 서버 이름, 데이터베이스)을 찾는다. 읽기만, `detected` 가 채워진 프로필을 `project-plan` 에 넘긴다. `src/main.rs` 에서 그대로 옮김(#2834 R4a).

버저닝(card C-69): 결과에 `detected` 옆으로 `versioning` 이 붙는다(프로필에는 들어가지 않는다). `scheme` 은 `semver`·`calver`·`headver`·`mixed`·`none` 중 하나, `latest` 는 버전으로 읽힌 가장 새 값(없으면 `null`), `evidence` 는 가장 새 태그 이름 최대 5개와 루트 `Cargo.toml`·`package.json`·`pyproject.toml` 의 버전을 `source`·`value`·`shape` 로 적은 목록, `tags` 는 태그를 읽은 결과(`read`·`no_repository`·`unavailable`)다. 태그 중 하나라도 모양이 읽히면 태그가 정하고, 아니면 매니페스트가 정한다. 두 방식에 다 맞는 모양(`26.4.1`)은 `shape: ambiguous` 와 `note` 로 적고 세지 않는다. 태그는 프로젝트 자신의 `.git` 에 대해 읽기 전용 `git for-each-ref` 한 번으로 읽고(git 환경 변수 정리 적용, 상위 폴더 저장소는 읽지 않음), 아무것도 쓰지 않으며 네트워크를 쓰지 않는다. 모양 판정 규칙은 `src/versioning.rs` 와 `skills/moai/references/internal/versioning.md`.
