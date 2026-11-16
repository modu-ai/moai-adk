# `project-detect`

operation `project.detect` · 코드 `src/ops/project_detect.rs` · #2496

`project-detect PROJECT` 는 `/moai project` 경로를 위해 프로젝트 사실(새 프로젝트인지, manifest, 언어, 언어 서버, 설정된 MCP 서버 이름, 데이터베이스)을 찾는다. 읽기만, `detected` 가 채워진 프로필을 `project-plan` 에 넘긴다. `src/main.rs` 에서 그대로 옮김(#2834 R4a).

버저닝(card C-69): 결과에 `detected` 옆으로 `versioning` 이 붙는다(프로필에는 들어가지 않는다). `scheme` 은 `semver`·`calver`·`headver`·`mixed`·`none` 중 하나, `latest` 는 버전으로 읽힌 가장 높은 값(날짜순이 아니라 버전순. 없으면 `null`), `evidence` 는 버전순으로 가장 높은 태그 이름 최대 5개와 루트 `Cargo.toml`·`package.json`·`pyproject.toml` 의 버전을 `source`·`value`·`shape` 로 적은 목록, `tags` 는 태그를 읽은 결과(`read`·`no_repository`·`unavailable`)이고 `unavailable` 이면 `tags_reason` 이 이유를 적는다. 태그 중 하나라도 모양이 읽히면 태그가 정하고, 아니면 매니페스트가 정한다. 두 가지로 읽히는 모양(`26.4.1`, `16.4`)은 `shape: ambiguous` 와 `note` 로 적고 세지 않는다. HeadVer 모양은 모두 SemVer 로도 유효하므로 `note` 에 그렇게 적고, SemVer 태그와 함께 있으면 SemVer 로 센다.

태그 이름은 파일에서만 읽는다. 프로그램을 시작하지 않고, git 객체와 git 설정을 읽지 않으며, 아무것도 쓰지 않고 네트워크를 쓰지 않는다. `<project>/.git` 이 실제 폴더이면 `refs/tags` 아래 일반 파일의 이름(내용은 읽지 않음. 링크는 태그가 아니며 따라가지 않음. 디렉터리 항목 최대 2,000개, 이름은 최대 4마디)과 `packed-refs` 의 `<id> refs/tags/<이름>` 줄(일반 파일, 최대 4 MiB, `^` 줄은 건너뜀)을 읽는다. `.git` 이 파일이면 연결된 작업 트리일 때만 따라간다: `gitdir: <경로>` 가 가리키는 폴더에 일반 파일 `commondir` 가 있고 그 폴더의 `gitdir` 파일이 `<project>/.git` 을 되가리킬 때, `commondir` 가 가리키는 공통 폴더에서 같은 규칙으로 읽는다. 두 포인터가 가리키는 폴더 자체는 실제 폴더인지(링크가 아닌지) 확인하지만, 거기까지 가는 경로는 확인하지 않는다. 그 밖의 경우(`.git` 이 링크, 서브모듈 포인터, 이 폴더를 되가리키지 않는 저장소를 향한 포인터, `refs`·`refs/tags`·`packed-refs` 가 링크나 특수 파일, 한도 초과)는 `tags: unavailable` 과 이유를 내고 태그를 읽지 않는다. `.git` 이 없으면 `no_repository` 이고 상위 폴더 저장소는 읽지 않는다. 짧고 평범한 이름(ASCII 영숫자와 `. - _ + / @`, 64바이트 이하)만 후보가 되고 출력에 나온다. 모양 판정 규칙은 `src/versioning.rs` 와 `skills/moai/references/internal/versioning.md`.
