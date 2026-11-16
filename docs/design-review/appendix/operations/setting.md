# `setting`

operation `setting.read` · 코드 `src/ops/setting.rs`, `src/setting.rs`, `src/user_config.rs` · #2871 후속 S1·S3

`setting list` 는 사용자 CLI 기본값을 닫힌 키 표로 보여 주고, `setting get KEY` 는 키 하나를 보여 준다. 읽기만 하고 쓰지 않는다(`writes:false`). 프로젝트 설정은 다루지 않는다(`/moai project` 몫).

| 키 | 값 | 없을 때 |
| --- | --- | --- |
| `install.hosts` | `claude`·`codex`·`omp` 의 비지 않은 부분집합, 각각 한 번, 소문자 그대로 | `["claude","codex"]` (지금 `moai update` 가 사용자 단계에서 쓰는 호스트) |
| `launcher.role_access` | `full` 또는 `prompt` | `full` |
| `launcher.advisor` | `none`·`fable`·`opus`·`sonnet`·`claude-*` | `none` |

- 파일은 HOME 상태 폴더의 `config/launcher.yaml`(JSON 부분집합, 사용자 전용)이다. 읽기는 `src/user_config.rs` 하나이고 런처(`src/session.rs`)도 같은 읽기를 쓴다. 디렉터리·심볼릭 링크·특수 파일·깨진 JSON·객체가 아닌 뿌리는 내용을 인용하지 않고 거절한다.
- 각 항목은 `source`(`file`/`default`)와 `valid` 를 가진다. 허용 집합 밖의 값은 `valid:false` 와 `error` 만 나오고 값은 인용하지 않는다. 표에 없는 키는 개수만 `unlisted_keys` 로 알린다(이름·값은 보이지 않는다).
- `selected_hosts(home)` (S3): `install.hosts` 가 있으면 그 호스트들을 `claude`·`codex`·`omp` 순서로, 없으면 위 기본값을 돌려준다. 깨진 파일·값은 오류이고 추측하지 않는다. 이 조각에서는 아무 명령도 이 함수를 아직 부르지 않는다(S4 가 `moai update` 에 연결한다).
