# `manifest-merge-driver`

operation `manifest.merge_driver` · 코드 `src/ops/manifest_merge_driver.rs`, `src/manifest_merge.rs` · `SPEC-20261126-manifest-merge-driver-rust`

`manifest-merge-driver %O %A %B %P` 는 `skills/moai/references/source-manifest.json` 의 git 머지 드라이버다. 사람이 아니라 git 이 부른다. 세 판(`%O` 공통 조상, `%A` 우리 쪽, `%B` 상대 쪽)을 3-way 로 합쳐 `%A` 파일에 쓰고 0 으로 끝난다. 드라이버 이름은 `.gitattributes` 의 `moai-manifest-rs` 이고, 등록하는 `git config` 한 줄은 `docs/agents-md-maintenance.md` 에 있다. moai 는 git 설정을 쓰지 않는다.

| 대상 | 고르는 방법 |
| --- | --- |
| 매니페스트의 값 | 양쪽이 같으면 그 값, 아니면 공통 조상에서 바뀐 쪽. `origin` 은 하위 키 하나씩. `null` 인 공통 값은 모르는 값으로 보아 양쪽이 같을 때만 고른다 |
| `files` | 한쪽에서만 더한 항목은 넣고, 한쪽에서만 지우고 다른 쪽이 그대로인 항목은 뺀다. 키는 정렬 순서로 쓴다 |
| 한쪽만 바꾼 파일의 `sha256` | 바꾼 쪽의 값 |
| 양쪽이 바꾼 파일의 `sha256` | 세 판의 blob 을 커밋 기록(`git log --all` 과 머지 중인 쪽의 커밋)에서 `sha256` 으로 찾아 `git merge-file --diff-algorithm=histogram` 으로 합친 결과의 `sha256` |

작업 폴더의 파일은 읽지 않는다. git 은 합친 파일을 쓰기 전에 드라이버를 부르고, `merge-tree` 에는 작업 폴더가 없다. 쓰는 바이트(`files` 정렬, 두 칸 들여쓰기, 끝 줄바꿈 하나)는 `moai check fix` 가 같은 트리에 쓰는 것과 같다.

정할 수 없으면 **거절**한다: 스킬 파일의 실제 충돌, 커밋 기록에 없는 판, 한쪽은 지우거나 이름을 바꾸고 다른 쪽은 바꾼 파일, 양쪽이 서로 다른 내용으로 더한 파일, 한쪽에서 사라지는 폴더에 다른 쪽이 파일을 더하거나 바꾼 경우, 어느 한쪽이라도 바꾼 스킬 파일에 자기 `merge` 속성이 있는 경우, 양쪽이 다르게 바꾼 값, 객체가 아닌 항목(공통 조상의 `null` 항목 포함), 키가 두 번 있거나 JSON 이 아니거나 `files` 객체가 없는 쪽, UTF-8 이 아닌 입력, 패키지 매니페스트가 아닌 `%P`, histogram 이 아닌 `diff.algorithm`, 켜진 `merge.renormalize`, 실패한 git 호출. 거절·오류·panic 은 모두 `%A` 에 git 의 글 머지(`git merge-file -L ours -L base -L theirs`)를 남기고, 표준 오류에 이유와 `moai check fix` 를 말하고, 1 로 끝난다. 글 머지가 깨끗해도 매니페스트는 충돌로 남는다. 결과는 임시 파일에 쓴 뒤 이름을 바꿔 넣으므로, 쓰기가 실패해도 잘린 `%A` 가 글 머지의 입력이 되지 않는다.

git 호출은 git 이 드라이버에 준 환경(`GIT_DIR`, `GITHEAD_*`, `GIT_CONFIG_PARAMETERS`)을 그대로 물려받는다. HOME·identity·장부는 읽지 않는다.

답: 성공하면 `path`, `merged: true`. 등록한 줄은 표준 출력을 버린다.

하지 않는 것: 매니페스트 검사와 다시 쓰기(`moai check fix`, `tools/source_manifest.py`), git 설정 쓰기, `git merge -X…` 옵션 감지(드라이버에 전달되지 않음), GitHub 서버의 머지(커스텀 드라이버를 돌리지 않음). 각 쪽의 `sha256` 은 믿으므로 매니페스트를 다시 쓰지 않고 커밋한 쪽은 틀린 해시를 남기고, 저장소의 `tests/source_manifest_facts.rs` 가 그것을 잡는다.
