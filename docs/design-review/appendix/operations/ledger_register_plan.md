# `ledger-register-plan`

operation `ledger.register.plan` · 코드 `src/ops/ledger_register_plan.rs`, `src/ledger/command.rs` · Issue #3068 P1

`ledger-register-plan PROJECT [--home DIR]` 는 프로젝트의 로컬 SQLite 원장을 `<HOME>/.moai/projects/<project_id>/` 에 만들면 무엇이 생기는지 미리 보인다. `project_id` 는 프로파일의 `project_id` 다. 쓰지 않는다. 결과: `project_id`, `ledger_path`, `state`(`absent`|`registered`|`unavailable`), `action`(`register`|`none`|`blocked`), `problem`(원장 오류 코드 또는 null), `would_create`(새로 생길 폴더와 파일), `plan_digest`. 이미 있는 파일은 읽기 전용으로 확인만 하고 바꾸지 않는다. HOME 이 Go 세대의 `~/.moai` 이면 `blocked`(`home_not_ready`).
