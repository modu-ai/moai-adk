# `claude-md-apply`

operation `claude_md.apply` · 코드 `src/ops/claude_md_apply.rs`, `src/project_update/claude_md.rs`

`claude-md-apply [PROJECT] [--keep <id>=shared|personal]... --plan-digest <digest> --permission-reference <ref>` 는 사용자가 확인한 `claude-md-plan` 대로 옛 지침 파일을 백업 폴더로 옮기고 새 파일을 쓴다. 인자는 미리보기가 낸 `apply.argv` 그대로다. 합성과 digest 는 미리보기와 같은 코드를 쓴다. 파일을 지우지 않고, 있는 파일 위에 쓰지 않는다. 예외는 한 줄이 더해지는 `.gitignore` 하나다.

## 처음 쓰기 전의 거절

아래는 모두 아무것도 쓰기 전에 거절하고 이유를 낸다(`claude-md-apply refused: <이유>; nothing was written`).

- `--permission-reference` 가 받는 형식(1~128바이트, 제어 문자 없음, `null`·`placeholder` 아님)이 아니거나, 미리보기에 적힌 자리표시 `<permission-reference>` 그대로이거나, 비밀 모양 값일 때. 오류 종류는 `permission_reference_missing`.
- 루트 `CLAUDE.md` 가 Go 세대 지침 파일이 아닐 때(`go_claude_md` 가 `null`).
- 프로필이 없거나 읽지 못하거나 자리표시일 때, 관리 표지 문법이 든 프로필 값이 있을 때.
- 거절한 출처가 있을 때(링크, 링크 폴더 아래, UTF-8 아님, 1 MiB 초과 등).
- 출처 경로가 보통 상대 경로가 아니거나, 이름이 `.gitignore` 인 출처가 있을 때(백업 안으로 옮기면 git 이 무시하는 백업 파일이 바뀐다).
- 모르는 id, 두 번 준 id, `shared`·`personal` 이 아닌 목적지.
- `shared` 로 고른 묶음에 비밀 모양 줄이 있을 때.
- 백업 폴더(상태 폴더의 `backups/`)가 링크이거나 링크 아래에 있거나 폴더가 아닐 때.
- git 작업 트리 안에서 receipt 나 백업 파일 하나라도 git 이 무시하지 않을 때. git 이 대답하지 못하면(실패, 저장소가 망가짐, `.git` 이 있는데 저장소가 아니라는 대답) 무시하지 않는 것으로 본다.
- `personal` 선택이 있는데 git 이 `CLAUDE.local.md` 를 추적하거나, 사용자가 쓴 `!` 규칙이 그 파일의 무시를 풀었거나, 규칙을 더해야 하는데 `.gitignore` 가 링크이거나 바꿀 수 없을 때(일반 파일 아님, 256 KiB 초과).
- `--plan-digest` 가 지금 합성한 digest 와 다를 때(출처·프로필·선택이 바뀜). 오류 종류는 `stale_plan`.
- 앞으로 120초 안에 쓰지 않은 백업 시각 폴더 이름이 없을 때.

git 작업 트리 밖에서는 git 확인 없이 진행하고 무시 규칙을 더하지 않으며, 답의 `backup_ignore` 가 `not_a_repository` 다.

## 적용 순서

1. `<상태 폴더>/backups/<시각>/` 을 소유자만 쓰는 폴더(0700)로 새로 만든다. 시각은 UTC `YYYYMMDDThhmmssZ` 이고, 이 이름이 receipt id 다.
2. 모든 플래그가 `false` 인 `pending` receipt 를 `<상태 폴더>/backups/<시각>/claude-md-receipt.json` 에 저장한다(0600). 저장하지 못하면 아무 파일도 옮기지 않는다.
3. 출처를 순서대로 `<상태 폴더>/backups/<시각>/claude-md/<원래 상대 경로>` 로 옮긴다. 옮기기 직전에 링크가 아닌지와 바이트가 미리보기 그대로인지 다시 확인한다. 하드 링크를 만든 뒤 원래 이름을 지우는 방식이라 있는 파일을 덮지 않고, 파일 권한은 그대로다. 하나 옮길 때마다 receipt 를 다시 저장한다. 비게 된 폴더는 남는다.
4. 옮긴 파일을 모두 다시 읽어 digest 가 같고 원래 경로가 비어 있는지 확인한다.
5. 필요하면 `.gitignore` 에 `CLAUDE.local.md` 한 줄을 더한다(파일이 CRLF 면 CRLF). 개인 파일을 쓰기 전에 하고, git 이 이제 무시하는지 확인한다. 다른 규칙이 이미 무시하면 더하지 않는다.
6. 새 `CLAUDE.md` 를, `personal` 선택이 있으면 이어서 `CLAUDE.local.md` 를 파일이 없을 때만 만든다. 쓴 뒤 다시 읽어 확인한다.
7. receipt 를 `committed` 로 저장한다.

## 파일 권한

- 백업한 파일은 원래 권한을 그대로 갖고, 소유자만 들어가는 폴더 안에 있다.
- 새 `CLAUDE.md` 는 옛 `CLAUDE.md` 의 권한을 그대로 쓴다(사용자가 `shared` 로 고른 줄은 더 좁은 권한의 출처에서 왔어도 같다).
- `CLAUDE.local.md` 와 receipt 는 0600 이다.
- Windows 에서는 권한 비트를 읽지 않아 receipt 의 `mode` 가 0600 으로 적힌다(확인하지 않음).

## 중간 실패

처음 쓴 뒤의 단계가 실패하면 receipt 는 `pending` 으로 남고, 오류가 멈춘 곳과 이유, 옮긴 개수, 쓴 경로, receipt 경로, 백업 위치를 말한다. 옮기거나 쓴 것이 하나라도 있으면 `partial apply: ...`(오류 종류 `partial`), 하나도 없으면 `claude-md-apply refused: ... no file was moved or written` 이다. 이 명령은 자동으로 되돌리지 않고 아무것도 지우지 않는다. 옮긴 파일은 백업 폴더의 같은 상대 경로에 있다. 자동 복원과 되돌리기 명령은 아직 없다.

receipt 는 단계마다 저장하므로, 한 단계를 한 뒤 저장하기 전에 멈추면 파일이 플래그보다 한 단계 앞설 수 있다. receipt 를 읽는 쪽은 플래그만 믿지 말고 디스크를 digest 와 비교한다.

## 답

`state`(`committed`), `receipt`(id), `receipt_path`, `backup`(옮긴 파일의 폴더), `plan_digest`, `moved`(옮긴 경로), `tracked_moved`(그 가운데 git 이 추적하던 경로, 작업 트리에서 삭제로 보인다), `written`(쓴 경로, `.gitignore` 가 있으면 맨 앞), `ignore_rule`(`{path, line}` 또는 `null`), `backup_ignore`(`git` | `not_a_repository`), `notice`(설정은 바로 적용되고, 실행 중인 세션과 lane 은 다시 시작한 뒤에야 바뀐 지침을 읽는다).

## receipt 형식 (`schema_version` 1)

파일 내용은 담지 않고 경로, digest, 권한만 담는다. 모르는 필드가 있으면 읽지 않는다.

- `schema_version`: `1`.
- `operation`: `claude_md.apply`.
- `project`: 프로젝트의 절대 경로.
- `stamp`: receipt id(백업 시각 폴더 이름).
- `plan_digest`, `permission_reference`: 받은 값 그대로(참조는 승인 자체가 아니다).
- `state`: `pending` | `committed`.
- `moves`: 옮길 순서의 목록. 항목은 `path`, `backup`(둘 다 프로젝트 기준 상대 경로), `sha256`, `mode`, `git`(`tracked` | `untracked` | `unknown`), `moved`.
- `writes`: 쓸 순서의 목록. 항목은 `path`, `sha256`, `mode`, `written`.
- `ignore`: `.gitignore` 변경이 없으면 `null`, 있으면 `path`, `line`, `before_sha256`(파일이 없었으면 `null`), `before_len`, `after_sha256`, `written`. 이전 바이트는 새 바이트의 앞 `before_len` 바이트다.
