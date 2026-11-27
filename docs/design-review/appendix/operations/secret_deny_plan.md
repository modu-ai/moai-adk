# `secret-deny-plan`

`operation secret_deny.plan` · 코드 `src/secret_deny.rs`, `src/ops/secret_deny_plan.rs` · SPEC-20261122-secret-file-deny-rules

`moai secret-deny-plan [프로젝트 폴더] [--pick 규칙]... [--retire-claudeignore]` 는 비밀 파일을 읽지 못하게 하는 `Read` 거절 규칙을 미리 보여 줘요. 아무것도 쓰지 않아요.

- `key_rules`: 자격 증명 파일 모양(`.env`, `.env.*`, `*.pem`, `*.key`, `*-service-account.json`, `.npmrc`, `.pypirc`) 가운데 파일이 실제로 맞은 것만 규칙으로 내요.
- `claudeignore`: Claude Code 는 `.claudeignore` 를 읽지 않아요. 빈 줄과 주석을 뺀 항목마다 규칙 하나를 내고, 끝이 `/` 인 항목은 폴더 안 내용을 막는 꼴(`Read(**/<폴더>/**/*)`, 중간에 `/` 가 있으면 `Read(<경로>/**/*)`)로 바꿔요. 자격 증명 모양인 항목만 `recommended` 예요.
- 규칙마다 맞은 파일(`matched`)과 그중 git 이 추적하는 파일(`tracked`, git 이 답하지 못하면 `null`), 같은 경로의 편집도 막는다는 경고가 있어요.
- 파일은 이름으로만 맞춰요. 열거나 해시하지 않아요. 대소문자는 구분해요.
- 걷기는 git 이 무시하는 파일도 보고, `.git` 과 자기 `.git` 을 가진 하위 폴더는 들어가지 않고, 심볼릭 링크를 따라가지 않아요. 항목 100,000개에서 멈추면 `walk.complete` 가 `false` 이고 `walk.note` 가 목록이 일부라고 말해요.
- `settings`: 루트 `.claude/settings.json` 을 링크를 따라가지 않고 읽기만 해요. `file_state`(`absent`·`present`·`not_read`), git 이 그 파일을 무시하는지(`settings_ignored`)와 워크트리 세션에 닿는지(`worktree_sessions`), 거절 목록에 이미 있는 `!` 항목(`negations`: 목록 안 위치 `index` 와 `entry`).
- `carve_outs`: 키 파일 규칙이 맞은 파일을 고르는 `!` 항목마다 한 줄이에요. `negation`, 그 파일을 맞춘 키 파일 규칙 `rule`, 파일 `files`, 출처 `source`(`settings` 또는 `claudeignore`). 고르기 전에 보여 줘요.
- `--pick 규칙`(여러 번 쓸 수 있어요)을 주면 `apply` 도 실려요: `plan_digest`, 쓴 뒤의 `permissions.deny` 목록 `final_deny`, 새로 더해지는 규칙 `adds`, `retire_claudeignore`. 미리보기가 내지 않은 규칙을 고르면 거절하고 아무것도 쓰지 않아요. 키 파일 규칙은 목록의 마지막 `!` 항목 뒤에 서도록 옮겨 써요. 다른 항목의 순서와 값은 그대로예요.
- `--retire-claudeignore` 는 `--pick` 이 하나 이상일 때만 되고, `.claudeignore` 가 링크이거나 보통 파일이 아니면 거절해요. 실제 이동은 `secret-deny-apply` 가 해요.
