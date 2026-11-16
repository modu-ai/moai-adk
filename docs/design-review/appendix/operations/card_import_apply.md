# `card-import-apply`

operation `card.import.apply` · 코드 `src/ops/card_import_apply.rs`, `src/forge/card_import.rs`, `src/ledger/import.rs` · Issue #3071 H3b

`card-import-apply PROJECT --from github --repo OWNER/REPO [--home DIR] --plan-digest D --permission-reference REF` 는 미리 본 열린 이슈를 로컬 원장에 카드로 쓴다. 호스트를 다시 읽어 얻은 이슈 집합의 digest 가 미리 보기의 `plan_digest` 와 같을 때만 쓰고(아니면 `the open issues changed since the preview`, 아무것도 쓰지 않음), 한 트랜잭션에서 전부 쓰거나 아무것도 쓰지 않는다. 같은 `(호스트, 프로젝트, 번호)` 의 카드가 이미 있으면 그 카드는 노트·기준·사건까지 그대로 두고 `already_there` 에 센다. 카드는 `draft` 이고 예약·상태 변경·링크는 만들지 않는다. 원장이 없으면 호스트를 읽기 전에 `ledger not found` 로 멈춘다(원장은 `ledger-register-apply` 만 만든다). 호스트의 이슈를 닫거나 바꾸지 않는다.
