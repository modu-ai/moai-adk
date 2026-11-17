# `home-routing-plan`

operation `home.routing_plan` · 코드 `src/ops/home_routing_plan.rs`, `src/home_routing.rs` (#2638)

`home-routing-plan HOME PACKAGE` 는 꾸러미의 `references/home/routing.md` 를 도구별로 그려(`{invoke}` → Claude `/moai`, Codex `$moai`) `HOME/.claude/CLAUDE.md`·`HOME/.codex/AGENTS.md` 의 `moai-routing` 관리 구간과 3자 비교(기준 = 마지막 적용이 기록한 구간 sha256, 현재 = 디스크의 구간, 새 것 = 렌더 결과)한다. 도구마다 `create`, `append`, `merge`, `unchanged`, `conflict`(고친 구간·깨진 표시). 폴더·파일이 링크면 그 도구는 `blocked`. OMP(`~/.omp/agent/AGENTS.md` 는 다른 사용자 지침을 가림)·Kiro·Cursor CLI 는 `not_supported` 에 사유와 함께 적고 쓰지 않는다. 아무것도 쓰지 않는다. 결과: `schema_version`, `home`, `package_digest`, `source_sha256`, `hosts`(`host`, `file`, `preflight`, `blocked_reason`, `section`), `not_supported`, `plan_digest`.

도구 폴더는 `src/host_home.rs` 하나가 정한다(카드 C-79, SPEC-20261117-host-home-resolver): `HOME` 이 이 계정의 홈(비밀번호 데이터베이스, 시험은 `MOAI_ACCOUNT_HOME`)일 때 `CLAUDE_CONFIG_DIR` 이 `HOME/.claude` 를, `CODEX_HOME` 이 `HOME/.codex` 를 대신한다(절대 경로만). 빈 값·상대 경로면 그 도구는 `blocked` 이고 사유가 변수 이름을 말한다(값은 되풀이하지 않음). 다른 홈(임시 HOME 등)은 변수와 상관없이 그 아래 폴더를 쓴다. 옮겨진 폴더 자체가 링크면 `HOME/.claude` 와 같이 `blocked`.
