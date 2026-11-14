# `home-retire-plan`

operation `home_retire.plan` · 코드 `src/ops/home_retire_plan.rs`, `src/home_retire.rs` · Issue #2871 (U2)

`home-retire-plan [--home DIR] [--now UTC]` 는 판정이 `go`·`mixed`·`rust-first` 인 `~/.moai` 의 맨 위 항목마다 HOME 이름 표(`home_layout::HOME_TABLE`)로 동작을 정한다: `retire`(`~/.moai-go-retired/<UTC>/` 로 이동), `keep`(이유 `user_data`·`owner_unknown`·`not_in_table`·`this_generation`·`merge_candidate_kept`·`shared_name`·`link`·`special_file`·`nested_checkout`), `skip-in-use`(돌고 있는 프로그램이 씀). 두 세대가 같이 쓰는 이름(`state`·`config`·`backups`)은 판정이 `go` 일 때만 이동한다. 비밀은 이름과 종류만 보고 열지도 걷지도 않는다. 막는 것: 살아 있는 Go 세션, `MOAI_HOME` 이 다른 곳, 은퇴 폴더가 폴더가 아님, 같은 stamp 가 이미 있음, 항목 500개 초과, 읽을 수 없는 이름. 다른 판정은 `none`(또는 `refuse`·`interrupted` 는 `blocked`). 아무것도 쓰지 않고 `plan_digest` 를 돌려준다.
