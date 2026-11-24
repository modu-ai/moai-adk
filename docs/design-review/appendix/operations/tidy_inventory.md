# `tidy-inventory`

operation `tidy.inventory` · 코드 `src/ops/tidy_inventory.rs`, `src/tidy/` · SPEC-20261124-home-tidy-slice2 (부모 SPEC-20261122-home-tidy)

`tidy-inventory HOME` 는 `HOME/.moai` 의 맨 위 항목을 모두 나열한다. 쓰지 않고, 일반 파일을 열지 않고, 링크를 따라가지 않고, 명령을 시작하지 않는다.

- `entries`: 항목마다 `name`, `class`, `reason`, `type`(`file`·`dir`·`symlink`·`fifo`·`socket`·`device`), `bytes`, `newest`(최근 변경, 초), `over_limit`, 보호 항목이 아니면 `files`. 크기는 일반 파일 길이의 합이고(폴더 자체 크기는 더하지 않음), `du` 와 다를 수 있다(APFS 복제는 부풀려 센다).
- `class`: `owned_live`(30일 안에 변경)·`owned_stale`·`retired_tool`(`home_state::SELECTIVE_TABLE` 의 옛 세대 항목과 `db`)·`other_product`(`claude-profiles`·`gpt-appserver`·`local-plugins`·`mcp`·`studio`)·`unknown`(표에 없거나 쓰는 쪽을 모르는 `state`·`logs`·`dev-login.sh`·`thread-media*`, 읽을 수 없는 폴더는 이유 `unreadable`)·`protected`.
- `protected`: 비밀 이름(`.env*`·`credentials*`·`secrets`·`gateway-auth`), 링크, 파이프·소켓·장치. 이름과 자기 디렉터리 데이터(종류, 크기, 시간)만 쓰고 열기·따라가기·해시를 하지 않는다. 계획의 후보가 되지 않는다.
- `kinds`: 자라는 폴더 종류(`reports`·`logs`·`evidence`·`worktrees`·`build`·`receipts`)마다 폴더가 있을 때만 한 줄. `count`, `bytes`, `oldest`, `newest`, `thresholds`(부모 SPEC 의 D-5 기본값에 `--threshold` 를 덮은 값), `passed`(넘은 필드). 항목은 `reports`·`logs` 에서는 파일, `worktrees` 에서는 `<프로젝트>/<이름>`, 나머지에서는 바로 아래 자식이다. `free_gib` 는 보여 주기만 하고(남은 디스크는 이 조각이 재지 않음) 넘은 것으로 세지 않는다.
- `--threshold <종류>.<필드>=<값>`: 0 이상의 수(소수 가능). 알 수 없는 종류·필드나 수가 아닌 값은 홈을 읽기 전에 인자 이름을 대고 거절한다.
- `--memory-index <경로>`: 파일 하나만 읽어(1 MiB 넘으면 읽지 않고 `over_limit`) `lines`, `bytes`, 호스트 한도(200줄 또는 25 KB) 대비 `percent` 와 `over_80`, 목표(100줄·12 KB·줄당 120바이트)를 `memory_index` 에 낸다. 없거나 읽을 수 없는 파일은 `found`·`reason` 으로 보고하고 실행은 계속한다. 파일을 쓰지 않고, `~/.claude*` 를 걷지 않는다.
- 상한: `--entry-limit`(맨 위 항목 하나 아래 방문 수, 기본 200000)과 `--time-limit-secs`(한 번의 실행, 기본 30). 넘으면 그 항목은 `over_limit` 이고 크기는 본 만큼이다.
- 이 조각에서 하지 않는 것: 요약 파일 쓰기, 보관, 삭제, 워크트리·빌드의 병합 증명(조각 3·4).
