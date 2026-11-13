# `labels-check`

operation `labels.check` · 코드 `src/ops/labels_check.rs`, `src/labels_check.rs` · Issue #2740

`labels-check PROJECT --repo OWNER/REPO` 는 `PROJECT/.github/moai/labels.yml`(이 저장소 모양: `reused`·`added`·`not_used`, `schema_version: 1`)에서 축을 읽고, 열린 Issue 를 GET 으로만 모두 읽어(페이지 100개씩, pull request 는 건너뜀) 축 규칙을 어긴 Issue 를 낸다. 위반 종류: `type_count`(type 이 1개가 아님), `priority_count`(priority 가 1개가 아님), `area_missing`(area 0개), `area_too_many`(area 3개 이상), `not_used`(쓰지 않는 라벨), `unlisted`(파일에 없는 라벨). 라벨을 달거나 떼지 않는다. 결과: `repo`, `open_issues_read`, `violating`, `by_kind`, `issues`(`number`, `labels`, `violations`). 배차·병합 판정의 입력으로 쓰지 않는다(라벨은 표시일 뿐).
