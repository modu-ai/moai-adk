# `check-tree`

operation `check.tree` · 코드 `src/ops/check_tree.rs`, `src/check_tree.rs` · Issue #3009 (S3a, 설계 댓글 6887461482 §2 "(가) 검사")

`check-tree [PROJECT]`(기본 `.`) 는 프로젝트에서 git 이 **추적하는** 파일 가운데 첫 폴더가 `.moai` 또는 `.moai-cli` 인 것(대소문자 무시, 폴더 바로 밑 파일은 제외)을 읽고 다섯 가지를 경고한다. 팀이 함께 보는 폴더에는 사람이 고치는 작은 글만 두고 기계가 쓰는 것은 HOME 에 둔다는 규칙(§2)을 기계가 지키게 하는 첫 단계다. 경고 단계라서 어떤 경고가 있어도 성공으로 답하고(`stage: "warning"`), 실패로 바꾸는 단계는 뒤 카드의 몫이다. 쓰기 0: `git ls-files -z` 한 번과 읽기뿐이다.

| `kind` | 무엇을 잡나 | 수량 칸 |
| --- | --- | --- |
| `large_file` | 파일이 1 MiB(1 048 576 바이트)를 넘음. 정확히 1 MiB 는 경고가 아님 | `bytes` |
| `absolute_path` | 글에 `/Users/<이름>` 또는 `/home/<이름>` 이 있음. 앞은 줄 처음·공백·따옴표·괄호·`=`·`:`·`,` 중 하나여야 하고(`docs/Users/x` 같은 상대 경로는 아님) 뒤에 이름 글자가 하나 이상 와야 함. 앞 8 KiB 안에 NUL 이 있는 파일(바이너리)은 읽지 않음 | `hits` |
| `session_id` | 경로에 세션 UUID(8-4-4-4-12 십육진수, 더 긴 낱말의 일부가 아님)가 있거나, 글의 한 줄에 `session`(대소문자 무시)과 UUID 가 함께 있음 | `hits` |
| `secret_name` | 파일 이름이 `.env`·`.env.*`(`.env.example` 류 제외)·`*.env`·`*.key`·`*.pem`·`*.p12`·`*.pfx`·`credentials`·`id_rsa` 같은 자격 증명 이름. 기존 `worktree_include::secret_name` 규칙에 설계가 말한 `*.env` 를 더함 | — |
| `project_id_not_random` | 프로필(`project.yaml`)의 `project_id` 가 S1 의 무작위 형태(소문자 `a-z2-7` 26자)가 아님. 사람이 고른 이름 id 는 다른 프로젝트와 같을 수 있다는 경고(결정 A, 댓글 6889383420)이며 값은 바꾸지 않는다 | — |
| `project_id_unusable` | 프로필이 있는데 쓸 수 있는 `project_id` 를 줄 수 없음(JSON 이 아님·링크·너무 큼 등) | — |

답: `stage`, `status`(`clean` / `warnings`), `tracked`(추적 파일 수), `in_scope`(대상 폴더 안 추적 파일 수), `missing`(추적돼 있으나 작업 폴더에서 지워진 파일), `not_read`(링크·일반 파일 아님·읽기 실패), `size_limit_bytes`, `counts`(종류별 개수), `findings`(`path`·`kind`·수량, 경로·종류 순, 최대 200개), `omitted`(200개를 넘어 싣지 않은 수), `writes: false`. 경고에는 경로와 개수만 있고 맞은 글자(사용자 이름·UUID·id)는 되풀이하지 않는다.

오류: `<project>` 폴더가 없으면 `not_found`(`PROJECT_PATH_UNAVAILABLE`), git 작업 폴더가 아니면 `invalid_input`(`not a git work tree`), git 을 못 실행하면 `io`.

하지 않는 것: `.gitignore` 관리 블록(프로젝트 파일에 쓰기, S3b), 경고를 실패로 바꾸는 `--strict` 류 단계, 실제 HOME 읽기.
