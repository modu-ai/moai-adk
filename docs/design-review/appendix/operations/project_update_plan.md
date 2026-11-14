# `project-update-plan`

operation `project.update_plan` · 코드 `src/ops/project_update_plan.rs`, `src/project_update.rs` · Issue #2936 (S1, 계획 6876732470, 리더 결정 6876725450)

`project-update-plan PROJECT` 는 프로젝트 `/moai update` 의 전체 미리보기를 한 번에 보여 준다. 쓰지 않는다. 각 단계는 이미 있는 plan 을 같은 프로세스 안에서 부르고, 그 plan 의 apply 명령을 그대로 적어 준다.

- `phases`: 사용자 UPDATE 흐름(#2936 6876720048)의 세 단계, 순서대로 `backup`·`install`·`merge`. 단계마다 `phase`, `state`(그 안의 가장 나쁜 step 상태: `unchanged` < `skip` < `ready` < `needs_decision` < `conflict` < `blocked`), `reason`, `steps`.
  - `backup`: `go-state` 한 step(`go_state::plan` 그대로). 옮길 것이 있으면 `needs_decision`(사용자에게 먼저 묻는다, PROJECT 2단계). `.agents`·`.codex`·설정·지침 파일은 아직 다루지 않는다(S2, `reason` 에 적음).
  - `install`: `profile`(저장된 프로필 + 지금 감지한 사실로 `project-plan`), `generation`(AGENTS.md 관리 구간과 생성 파일, `generation-plan`), `claude-settings`(MoAI 가 이미 가진 키에서 `hooks` 를 뺀 묶음, Q2 A), `github`(`github-plan` `{"protocol_version":1}`; 고르지 않은 프로젝트는 `skip`).
  - `merge`: `{"phase":"merge","state":"not_run","reason":…,"candidates":[]}` — 이후 조각(S1f).
- step: `step`, `state`(`ready`·`unchanged`·`conflict`·`needs_decision`·`blocked`·`skip`; `skip_in_use` 는 S2b 용으로 예약만 하고 내보내지 않음), `reason`, `preview`, `plan_digest`, `apply`. `apply` 는 적용할 수 있는 step 에만 `{"argv":[…],"stdin":<json|null>}` 이고 나머지는 `null`. `argv` 는 `moai` 로 시작하는 낱말 목록이고 자리표시자는 `<permission-reference>` 하나뿐이다.
- 막힘 규칙: 한 step 의 오류는 그 step 만 `blocked`(이유에 오류 문구). `backup` 에 옮길 것이 남아 있으면 `install` 의 모든 step 이 `blocked: re-run after the backup phase`. 프로필이 없으면 `install` 의 모든 step 이 `needs_decision: run /moai project`(프로필은 만들지 않음, Q5). `profile` 이 `ready` 이면 `generation` 은 `blocked: re-run after the profile apply`(generation plan 이 프로필 바이트를 묶음).
- `generation` 요청: 언어는 프로필 `language.instruction_files` 가 `en`·`ko` 면 그것, 아니면 `en`; `workflow_reference`·`workflow_id` 는 생성 manifest 의 기록, 없으면 `<state>/workflows/project/`·`project`; `local_file` 은 manifest 가 `.gitignore` 의 `moai-local` 구간을 가지면 `ignored`, 아니면 `none`. manifest 가 요청이 기록되지 않은 프로젝트 에이전트·스킬(`generation:native-role:`·`generation:project-skill:`·`generation:skill-projection:`)을 가지면 `needs_decision`(Q1 A: `remove` 를 계획하지 않음, `preview.not_rebuilt` 에 경로).
- 그 밖: `project`(정규 경로), `detected`(감지 사실, 실패 시 `null`), `state_root`, `all_unchanged`(`backup`·`install` 의 모든 step 이 `unchanged` 또는 `skip`; `merge` 는 세지 않음).
