# `archive-plan`

operation `archive.plan` · 코드 `src/ops/archive_plan.rs` · D1 #2187

`archive-plan PROJECT [--grace-days N]` 는 `.moai-cli/specs/` 아래 SPEC 폴더 중 보관할 수 있는 것과 나머지가 빠진 이유를 보여 준다(REQ-D1-01). `--grace-days` 는 프로필의 비활성 기간을 덮어쓴다. 쓰지 않고 stdin 도 읽지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R2a).
