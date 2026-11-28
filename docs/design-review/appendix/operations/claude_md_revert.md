# `claude-md-revert`

operation `claude_md.revert` · 코드 `src/ops/claude_md_revert.rs`, `src/project_update/claude_md.rs`

`claude-md-revert [PROJECT] --receipt <id> --permission-reference <ref>` 는 `claude-md-apply` 하나를 receipt 를 보고 되돌린다. `<id>` 는 apply 가 답한 `receipt`(백업 시각 폴더 이름, `YYYYMMDDThhmmssZ`)다. 실패했거나 중간에 멈춘 apply 도 같은 명령으로 되돌린다. 옮긴 원본을 지우지 않고, 있는 파일 위에 옮기지 않으며, 링크를 따라가지 않는다.

## 처음 쓰기 전의 거절

아래는 모두 아무것도 쓰기 전에 거절하고 이유를 낸다(`claude-md-revert refused: <이유>; nothing was written`). 이유에는 receipt 나 인자의 값을 되돌려 싣지 않는다.

- `--permission-reference` 가 받는 형식(1~128바이트, 제어 문자 없음, `null`·`placeholder` 아님)이 아니거나, 자리표시 `<permission-reference>` 그대로이거나, 비밀 모양 값일 때. 오류 종류는 `permission_reference_missing`.
- `--receipt` 가 시각 이름(16글자 `YYYYMMDDThhmmssZ`)이 아닐 때.
- receipt 가 없을 때(`receipt not found`, 오류 종류 `not_found`), receipt 경로의 어느 부분이 링크이거나 일반 파일이 아니거나 16 MiB 를 넘을 때.
- receipt 가 이 형식으로 읽히지 않을 때(JSON 아님, 모르는 필드 등: `is no receipt of this step`).
- 읽히지만 apply 가 쓰지 않는 값이 있을 때(`is not a receipt an apply writes`):
  - `schema_version` 이 1 이 아니거나 `operation` 이 `claude_md.apply` 가 아닐 때, `project` 가 이 프로젝트의 절대 경로와 다를 때, `stamp` 가 `--receipt` 와 다를 때, `state` 가 아는 다섯 값이 아닐 때.
  - `moves` 의 경로가 보통 상대 경로가 아니거나, apply 가 옮기는 자리(루트의 네 지침 파일, 상태 폴더의 `docs/local-rules/` 아래)가 아니거나, 이름이 `.gitignore` 이거나, 두 번 나올 때. 루트 `CLAUDE.md` 가 `moves` 에 없을 때.
  - `backup` 이 그 시각 폴더의 `claude-md/<path>` 와 글자까지 같지 않을 때.
  - digest 가 `sha256:<64자리 16진수>` 가 아니거나, `mode` 가 0777 보다 크거나, `git` 이 세 낱말이 아닐 때.
  - `writes` 가 `CLAUDE.md` 하나, 또는 `CLAUDE.md` 와 `CLAUDE.local.md` 순서의 둘이 아닐 때.
  - `ignore` 가 `.gitignore` 에 `CLAUDE.local.md` 한 줄을 더한 규칙이 아니거나, `CLAUDE.local.md` 를 쓰지 않는데 규칙이 있을 때.

## 되돌리는 규칙과 순서

receipt 의 플래그만 믿지 않고 디스크의 바이트를 digest 와 비교한다. 그래서 멈춘 apply, 멈춘 revert, 두 번째 실행에 같은 규칙이 통한다. 순서는 apply 가 쓴 개인 글이 git 에 보이는 때가 없도록 정했다.

1. apply 가 쓴 파일(뒤에 쓴 것부터: `CLAUDE.local.md`, `CLAUDE.md`).
   - 일반 파일이고 바이트가 쓴 그대로이며, 그 경로에 원본이 없었거나 원본이 백업에서 돌아올 수 있으면 지운다. 이 명령이 지우는 파일은 이것뿐이다(`.gitignore` 는 아래 3).
   - 원본이 이미 그 자리에 있으면 그대로 둔다.
   - 백업 원본이 없거나 바뀌었으면, 쓴 그대로인 새 파일도 남기고 `conflicts` 에 넣는다.
   - 바뀐 파일, 링크, 그 밖의 것이 있으면 남기고 `conflicts` 에 넣는다.
2. 옮긴 파일(`moves` 순서).
   - 원래 자리에 기록한 바이트의 파일이 이미 있으면 아무것도 하지 않는다. 백업 사본이 함께 있어도 지우지 않는다.
   - 백업에 기록한 바이트가 있고 원래 자리가 비었으면, 없는 폴더를 만들고 옮긴 뒤 receipt 의 권한을 준다.
   - 원래 자리에 다른 것이 있거나 그 경로에 링크가 있으면 남기고 `conflicts` 에 넣는다. 원본은 백업에 남는다.
   - 기록한 바이트가 두 곳 어디에도 없고 receipt 가 옮겼다고 적었으면(또는 자리가 비었는데 백업 경로에 다른 것이 있으면) `missing` 에 넣는다. 옮긴 적이 없거나 앞선 실행이 돌려놓은 파일을 그 뒤에 누가 바꾸거나 치운 것은 세지 않는다.
3. `.gitignore`(apply 가 규칙을 더한 경우).
   - 파일이 apply 가 남긴 바이트 그대로이고, 원본을 돌려놓은 뒤 루트 `CLAUDE.local.md` 가 아래 둘 가운데 하나일 때만 그 한 줄을 뺀다. apply 가 만든 파일이면 지운다. receipt 가 뭐라고 적었든, "이전 바이트 + 그 한 줄"이 지금 파일과 같을 때만 뺀다.
     - 없다. 원본이 백업에서 돌아오지 못했고 apply 가 쓴 파일도 없는 경우를 포함한다.
     - 기록한 원본이다(링크가 아닌 일반 파일이고 바이트가 `moves` 의 digest 와 같으며, apply 가 쓰는 글의 digest 와 다르다). apply 는 그 파일을 무시하는 규칙이 없을 때만 줄을 더하므로, 이 원본은 apply 전에도 git 에 보였다. 줄을 빼면 적용 전 상태가 된다.
   - 그 밖(apply 가 쓴 글이 남음, 다른 파일, 링크, 폴더, 읽지 못하는 것)이면 규칙을 남기고 `conflicts` 에 넣는다. 그것을 치우고 다시 실행하거나 줄을 손으로 지운다. apply 가 쓴 글은 git 에 보이게 두지 않는다.
   - apply 뒤 바뀌었으면 남기고 `conflicts` 에 넣는다.

폴더는 지우지 않는다. 백업 시각 폴더와 receipt 도 남는다.

## receipt 상태

- 되돌릴 것이 남지 않으면 `reverted`, 남으면 `partial` 로 저장한다. `partial` 은 원인을 없앤 뒤 같은 명령을 다시 실행한다.
- `reverted` 나 `restored`(실패한 apply 가 스스로 모두 되돌림) receipt 는 아무것도 쓰지 않고 `state: reverted` 와 `note` 를 답한다.
- 되돌린 뒤 receipt 를 저장하지 못하면 `partial apply: ...`(오류 종류 `partial`)이고, 같은 명령을 다시 실행하면 디스크를 보고 끝낸다.

## 답

`state`(`reverted` | `conflict`), `receipt`, `receipt_path`, `receipt_state`, `restored`(돌려놓은 원본, 이전 바이트로 돌린 `.gitignore`), `removed`(지운 새 파일), `conflicts`(그대로 둔 경로), `missing`(기록한 바이트가 어디에도 없는 원본), `reasons`(`conflicts` 와 `missing` 경로마다의 이유), `note`(할 일이 없었을 때만 글, 그 밖은 `null`), `notice`(설정은 바로 적용되고, 실행 중인 세션과 lane 은 다시 시작한 뒤에야 바뀐 지침을 읽는다).

## 한계

- 같은 파일 시스템 안에서 하드 링크를 만든 뒤 원래 이름을 지우는 방식으로 옮긴다.
- receipt 는 승인이 아니다. receipt 를 고칠 수 있는 사람은 그 안의 digest 도 고칠 수 있으므로, 백업 폴더의 권한이 첫 번째 보호다.
- apply 뒤에 사용자가 다른 무시 규칙을 지웠으면, 돌아온 개인 파일이 git 에 보일 수 있다. receipt 는 apply 전의 무시 상태를 담지 않는다.
- Windows 에서는 권한을 되살리지 않는다(확인하지 않음).
