# `labels-plan`

operation `labels.plan` · 코드 `src/ops/labels_plan.rs` · #2434

`labels-plan PROJECT` 는 `.github/moai/labels.yml` 로 프로젝트 라벨 동기화를 계획한다(결정 40). stdin 의 typed 요청 `{schema_version, owner, repo}`, GET 만, 지우지 않는다. stderr 가 터미널이면 진행 줄을 보인다. `src/main.rs` 에서 그대로 옮김(#2834 R4b).
