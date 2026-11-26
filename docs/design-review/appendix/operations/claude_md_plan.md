# `claude-md-plan`

operation `claude_md.plan` · 코드 `src/ops/claude_md_plan.rs`, `src/project_update/claude_md.rs`

`claude-md-plan [PROJECT]` 는 Go 세대에서 올라온 프로젝트의 옛 지침 파일에 무엇이 있고 그 가운데 어느 줄이 이 프로젝트만의 것인지 보여 준다. 쓰지 않고, 링크를 따라가지 않으며, git 은 읽기 명령(`ls-files`, `log`, `cat-file`)만 부른다. 줄 글자는 출력하지 않고 줄 번호와 개수만 낸다.

필드:

- `read_only`: 항상 `true`.
- `go_claude_md`: 루트 `CLAUDE.md` 가 Go 세대 지침 파일인 근거. 일반 파일이 Go 큐레이터 표지(`moai:learned-start`, `moai:harness-start`, `moai:learned-local-start`)를 가지면 `marker`, moai 관리 블록이 없고 찾은 Go 매니페스트가 `CLAUDE.md` 를 기록하면 `manifest`, 그 밖은 `null`.
- `go_manifest`: 찾은 Go 매니페스트의 경로(`go-residue` 단계와 같은 순서로 찾는다). 없으면 `null`.
- `sources`: `CLAUDE.md`, `AGENTS.md`, `AGENTS.local.md`, `CLAUDE.local.md`, 상태 폴더의 `docs/local-rules/` 아래 모든 파일(경로 순) 가운데 있는 것. 각 항목은 `path`, `sha256`, `line_count`, `git`(`tracked` | `untracked` | `unknown`), `comparison`, `project_line_count`, `project_lines`(1부터 세는 줄 번호).
- `refused`: 읽지 않는 출처의 `path` 와 `reason`. 링크(`is a link`), 링크 폴더 아래(`lies below a link folder`), UTF-8 아님(`is not UTF-8 text`), 1 MiB 초과(`is larger than 1 MiB`), 일반 파일 아님(`is not a regular file`), 읽지 못함(`cannot be read`), 목록을 못 읽는 폴더(`cannot be listed`), UTF-8 이 아닌 이름(`has a name that is not UTF-8 text`).

`comparison.used` 와 프로젝트 줄:

1. `manifest_equal`: 지금 바이트가 매니페스트가 그 경로에 기록한 sha256 과 같다. 프로젝트 줄은 0개.
2. `moai_first_text`: `AGENTS.local.md` 나 `CLAUDE.local.md` 가 moai 가 처음 만들 때 쓰는 글과 똑같다. 프로젝트 줄은 0개.
3. `git_history`: 그 경로의 git 기록(모든 ref, 최신 200개 버전까지)에 sha256 이 기록과 같은 버전이 있다. 그 버전(`blob`)에 없는 줄만 센다.
4. `all_lines`: 그 밖은 모든 줄을 센다. `reason` 은 `no_record`(매니페스트가 그 경로를 기록하지 않음), `not_a_repository`, `git_failed`, `no_matching_version`.

어느 경우든 완결된 `moai-workflow` 관리 블록의 줄과 Go 큐레이터 표지 한 줄(시작·끝 표지만 있는 줄)은 프로젝트 줄에서 뺀다.
