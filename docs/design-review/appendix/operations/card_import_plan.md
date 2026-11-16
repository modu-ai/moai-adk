# `card-import-plan`

operation `card.import.plan` · 코드 `src/ops/card_import_plan.rs`, `src/forge/card_import.rs` · Issue #3071 H3b

`card-import-plan PROJECT --from github --repo OWNER/REPO [--home DIR]` 는 코드 호스트의 열린 이슈를 로컬 원장의 카드로 가져오면 무엇이 되는지 미리 보여 준다. 모든 페이지를 읽고(끝까지 읽지 못한 목록은 오류), 이슈마다 만들어질 카드(제목, 종류와 그 출처, 읽어 낸 합격 기준 행 수, 구성원 댓글에서 남길 노트 수, 걸러 낼 제어 줄 수)를 `cards` 로 낸다. 합격 기준 표를 읽지 못한 이슈는 `criteria_unreadable`, 같은 번호로 두 번 온 이슈는 `duplicate_numbers`, 건너뛴 풀 리퀘스트 수는 `pull_requests_skipped` 에 있다. 카드마다 `source`(`호스트 owner/repo#번호`)가 있다. `ledger.state` 는 원장이 `ready`·`absent`·`unavailable` 중 어느 쪽인지 읽기 전용으로 알리고, `ledger.sources` 는 원장 카드의 출처 저장소별 장수, `ledger.notice` 는 원장에 이 가져오기와 다른 저장소의 카드가 이미 있을 때의 한 줄 알림이다(막지 않는다). 답에 `plan_digest` 가 있다. 읽기(GET)만 하고 원장·폴더·파일을 만들지 않는다. `--from gitlab` 은 통신 코드가 아직 없어 이유를 말하고 0 이 아닌 코드로 끝난다.
