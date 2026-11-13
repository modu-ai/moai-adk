# `home-routing-plan`

operation `home.routing_plan` · 코드 `src/ops/home_routing_plan.rs`, `src/home_routing.rs` (#2638)

`home-routing-plan HOME PACKAGE` 는 꾸러미의 `references/home/routing.md` 를 도구별로 그려(`{invoke}` → Claude `/moai`, Codex `$moai`) `HOME/.claude/CLAUDE.md`·`HOME/.codex/AGENTS.md` 의 `moai-routing` 관리 구간과 3자 비교(기준 = 마지막 적용이 기록한 구간 sha256, 현재 = 디스크의 구간, 새 것 = 렌더 결과)한다. 도구마다 `create`, `append`, `merge`, `unchanged`, `conflict`(고친 구간·깨진 표시). 폴더·파일이 링크면 그 도구는 `blocked`. Kiro·Cursor CLI 는 `not_supported` 에 사유와 함께 적고 쓰지 않는다. 아무것도 쓰지 않는다. 결과: `schema_version`, `home`, `package_digest`, `source_sha256`, `hosts`(`host`, `file`, `preflight`, `blocked_reason`, `section`), `not_supported`, `plan_digest`.
