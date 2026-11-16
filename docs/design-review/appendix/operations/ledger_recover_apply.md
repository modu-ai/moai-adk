# `ledger-recover-apply`

operation `ledger.recover.apply` · 코드 `src/ops/ledger_recover_apply.rs`, `src/ledger_recover.rs`, `src/ledger/recover.rs` · Issue #3068 P6

`ledger-recover-apply PROJECT --source FILE [--home DIR] [--journal DIR] [--observations FILE] [--now UTC] --plan-digest D --permission-reference REF` 는 같은 입력으로 다시 계산한 미리보기의 digest 가 같을 때만 움직인다(아니면 `stale_plan`, 아무것도 쓰지 않음). 새 원장을 `<원장>.recovering` 에 만들어 `integrity_check`·`foreign_key_check` 를 통과시킨 뒤, 망가진 파일을 `<원장>.damaged-<UTC>` 로 남기고(하드 링크: 원래 이름이 빈 적이 없다) 새 파일을 원래 이름으로 바꿔 놓는다. 아무것도 지우지 않는다. 결과 `state`: `applied` | `unchanged`(이 계획의 결과가 이미 제자리에 있음; 아무것도 쓰지 않음). 이 계획의 결과가 아닌 쓸 수 있는 원장(`target_healthy`), 링크·폴더인 원장 경로, 이미 쓰인 `damaged` 이름은 거절한다.
