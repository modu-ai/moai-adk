# `reports-move-apply`

operation `reports_move.apply` · 코드 `src/ops/reports_move_apply.rs`, `src/reports_move.rs` · 카드 C-126 (SPEC-20261119-update-moves-reports)

`reports-move-apply --plan-digest D --permission-reference REF [PROJECT]` 는 사용자가 확인한 `reports-move-plan` 그대로 옮긴다. 계획(JSON)은 표준입력으로 받는다.

- 거절: 권한 참조가 없거나, 표준입력 계획의 `plan_digest` 가 인자와 다르거나 계획 내용과 맞지 않거나, 계획을 지금 다시 만든 것과 digest 가 다르면(미리 본 뒤 파일이 생기거나 사라지거나 추적됐다) 아무것도 옮기지 않는다. 계획이 `blocked` 여도 같다.
- 옮김: 계획의 각 파일을 같은 상대 경로로 옮긴다. 같은 파일 시스템이면 하드 링크를 만든 뒤 원본을 지우고, 아니면 새 파일로 복사한 뒤 원본을 지운다. 목적지에 같은 이름이 이미 있으면 건너뛰고 세며 덮어쓰지 않는다.
- 권한: 새로 만드는 폴더는 0700 이고 옮긴 파일은 그룹·다른 사람 권한 비트를 지운다(Unix).
- 다시 실행: 옮긴 뒤 새로 만든 `reports-move-plan` 은 `unchanged`(옮길 것 없음)이고, 옮기기 전의 옛 계획을 다시 적용하면 파일이 이미 옮겨졌으니 낡은 계획으로 거절되어 아무것도 바꾸지 않는다.
- 중간에 실패하면 `partial apply: N file(s) moved; failed at …` 으로 알리고, 계획을 다시 만들어 이어 간다.

결과 필드: `destination`, `moved`, `skipped_existing`, `tracked_left_alone`, `unmovable`, `items` (`{from, to}`). 되돌리기: 옮긴 파일을 `items` 의 `to` 에서 `from` 으로 되옮기면 된다(원본은 지워졌다). 추적 파일은 건드리지 않는다.
