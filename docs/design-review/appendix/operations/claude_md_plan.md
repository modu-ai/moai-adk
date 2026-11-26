# `claude-md-plan`

operation `claude_md.plan` · 코드 `src/ops/claude_md_plan.rs`, `src/project_update/claude_md.rs`

`claude-md-plan [PROJECT] [--keep <id>=shared|personal]...` 는 Go 세대에서 올라온 프로젝트의 옛 지침 파일에 무엇이 있고 그 가운데 어느 줄이 이 프로젝트만의 것인지, 고른 묶음으로 새 파일이 어떻게 되는지 보여 준다. 쓰지 않고, 링크를 따라가지 않으며, git 은 읽기 명령(`ls-files`, `log`, `cat-file`)만 부른다. 프로젝트 줄은 묶음의 글로 출력하고, 비밀 모양 값이 있는 줄은 개수만 내며 어느 출력에도 싣지 않는다.

필드:

- `read_only`: 항상 `true`.
- `go_claude_md`: 루트 `CLAUDE.md` 가 Go 세대 지침 파일인 근거. 일반 파일이 Go 큐레이터 표지(`moai:learned-start`, `moai:harness-start`, `moai:learned-local-start`)를 가지면 `marker`, moai 관리 블록이 없고 찾은 Go 매니페스트가 `CLAUDE.md` 를 기록하면 `manifest`, 그 밖은 `null`.
- `go_manifest`: 찾은 Go 매니페스트의 경로(`go-residue` 단계와 같은 순서로 찾는다). 없으면 `null`.
- `sources`: `CLAUDE.md`, `AGENTS.md`, `AGENTS.local.md`, `CLAUDE.local.md`, 상태 폴더의 `docs/local-rules/` 아래 모든 파일(경로 순) 가운데 있는 것. 각 항목은 `path`, `sha256`, `line_count`, `git`(`tracked` | `untracked` | `unknown`), `comparison`, `project_line_count`, `project_lines`(1부터 세는 줄 번호).
- `refused`: 읽지 않는 출처의 `path` 와 `reason`. 링크(`is a link`), 링크 폴더 아래(`lies below a link folder`), UTF-8 아님(`is not UTF-8 text`), 1 MiB 초과(`is larger than 1 MiB`), 일반 파일 아님(`is not a regular file`), 읽지 못함(`cannot be read`), 목록을 못 읽는 폴더(`cannot be listed`), UTF-8 이 아닌 이름(`has a name that is not UTF-8 text`).
- `hunks`, `picks`, `results`, `warnings`, `apply_blocked`, `plan_digest`, `apply`: 아래에서 설명한다.

`comparison.used` 와 프로젝트 줄:

1. `manifest_equal`: 지금 바이트가 매니페스트가 그 경로에 기록한 sha256 과 같다. 프로젝트 줄은 0개.
2. `moai_first_text`: `AGENTS.local.md` 나 `CLAUDE.local.md` 가 moai 가 처음 만들 때 쓰는 글과 똑같다. 프로젝트 줄은 0개.
3. `git_history`: 그 경로의 git 기록(모든 ref, 최신 200개 버전까지)에 sha256 이 기록과 같은 버전이 있다. 그 버전(`blob`)에 없는 줄만 센다.
4. `all_lines`: 그 밖은 모든 줄을 센다. `reason` 은 `no_record`(매니페스트가 그 경로를 기록하지 않음), `not_a_repository`, `git_failed`, `no_matching_version`.

어느 경우든 완결된 `moai-workflow` 관리 블록의 줄과 Go 큐레이터 표지 한 줄(시작·끝 표지만 있는 줄)은 프로젝트 줄에서 뺀다.

## 묶음 (`hunks`)

- 절은 마크다운 제목 줄(`#` 1~6개, 앞 공백 3칸까지, 코드 펜스 안은 제외)부터 다음 제목 앞까지다. 첫 제목 앞의 줄도 한 절이다. 펜스는 파일 전체를 읽어 판정한다.
- 빈 줄이 아닌 프로젝트 줄이 있는 절마다 묶음 하나. 묶음은 그 절의 프로젝트 줄만 순서대로 글자 그대로 담고, 문맥 줄은 더하지 않는다. 줄 끝 문자(CRLF 의 CR 포함)와 파일 첫머리의 BOM 은 담지 않는다.
- `id`: `<path>:<첫 프로젝트 줄 번호>`. `path`, `heading`(그 절의 제목 줄, 첫 제목 앞이면 `null`), `line_count`, `text`(줄을 `\n` 으로 이은 글).
- `same_as`: 앞선 묶음과 줄이 똑같으면 그 `id`. 이때 `text` 와 `heading` 은 `null` 이다. 이런 묶음도 고를 수 있다.
- `hidden_lines`: 출력하지 않은 비밀 모양 줄 수. 알려진 접두사의 토큰, `password`·`token` 류 대입, `Authorization` 머리글, 비밀번호가 든 URL, 사용자 부분이 토큰 모양(8자 이상이고 숫자나 대문자가 있으며 자리표시가 아님)인 URL(`scheme://<값>@호스트`, `ssh://git@…` 는 보인다), 개인 키의 `BEGIN` 줄부터 `END` 줄(없으면 파일 끝)까지다. `BEGIN` 줄을 글 속에서 말만 해도 같은 규칙이 적용된다. 보이지 않는 문자로 나눈 값도 잡는다. 제목 줄이 이런 줄이면 `heading` 은 `null` 이다.
- `moved_paths`: 묶음의 글이 말하는, 이 단계가 옮기는 경로(읽은 출처와 `docs/local-rules` 폴더). 더 긴 경로나 이름의 일부(`.claude/CLAUDE.md`, `AGENTS.md.bak`)는 세지 않는다.

## 선택과 결과 (`picks`, `results`)

- `--keep <id>=shared` 는 새 `CLAUDE.md` 에, `--keep <id>=personal` 은 `CLAUDE.local.md` 에 묶음을 둔다. 모르는 id, 두 번 준 id, 두 낱말이 아닌 목적지는 명령 오류다. 오류에 되돌려 보이는 값이 비밀 모양이면 가린다.
- `results` 는 프로필을 쓸 수 없으면 빈 목록이다. `CLAUDE.md` 는 프로필에서 렌더한 관리 블록 뒤에 `shared` 묶음을 출처 순서로 붙인다. `personal` 선택이 하나라도 있으면 `CLAUDE.local.md` 가 moai 가 처음 만들 때 쓰는 글 뒤에 그 묶음들을 붙인 것으로 더해진다. 언어는 프로필의 `language.instruction_files`, 없으면 `user_chat`, 없으면 `en` 이다(생성 매니페스트는 요청 언어를 기록하지 않는다).
- 묶음 앞에는 빈 줄 하나를 두고 묶음 양 끝의 빈 줄은 버린다. 옮기는 파일을 가리키는 import(`@<경로>`, 프로젝트 루트나 그 출처의 폴더 기준, 코드 펜스와 백틱 밖)가 있는 줄과 Go 큐레이터 시작·끝 표지가 든 줄은 뺀다. 원래 펜스 안이던 import 줄도 새 파일에서 펜스 밖이 되면 뺀다.
- 각 결과는 `path`, `line_count`, `sha256`(쓸 파일 전체 기준), `text`(비밀 모양 줄을 뺀 글), `hidden_lines`, `hunks`(고른 묶음의 id), `left_out`(뺀 줄의 `<path>:<줄>` 과 `reason`: `imports_a_moved_file` | `go_curator_marker`).

## 경고와 적용 명령

- `warnings`: 결과 파일이 200줄을 넘으면 파일 이름과 줄 수. `.claude/CLAUDE.md` 가 있으면 알린다(읽지도 옮기지도 않는다). 경고는 적용 명령을 막지 않는다.
- `apply_blocked`: 적용 명령을 내지 않는 이유들. 거절한 출처, 없거나 읽지 못하는 프로필, 첫 실행의 자리표시 프로필, 관리 표지 문법이 든 프로필 값, 비밀 모양 줄이 있는 묶음의 `shared` 선택.
- `plan_digest` 와 `apply`: `apply_blocked` 가 비었을 때만 값이 있다(그 밖은 `null`). digest 는 모든 출처의 경로와 sha256, 프로필 바이트, 정렬한 선택, 결과 글을 묶는다. `apply.argv` 는 `moai claude-md-apply <프로젝트> [--keep ...] --plan-digest <digest> --permission-reference <permission-reference>` 낱말 목록이고 선택이 없어도 낸다.
