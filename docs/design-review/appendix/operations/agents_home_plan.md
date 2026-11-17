# `agents-home-plan`

operation `agents.home_plan` · 코드 `src/ops/agents_home_plan.rs`, `src/agent_install.rs` (#2592 S2)

`agents-home-plan HOME PACKAGE AGENTS` 는 스킬 밖 에이전트 원본 폴더 `AGENTS`(저장소·릴리스 폴더의 `agents/`, `skills/` 옆, #2721)의 `*.md` 를 host 별로 그려(`src/agent_defs.rs`; `reads:` 항목은 스킬 꾸러미 `PACKAGE/references/` 에서 확인), `HOME/.claude/agents/`·`HOME/.codex/agents/`·`HOME/.omp/agent/agents/`(#2633) 의 `moai-<역할>` 파일과 3자 비교(기준 = 마지막 적용이 기록한 sha256, 현재 = 디스크, 새 것 = 렌더 결과)한다. 파일마다 `create`, `update`, `unchanged`, `preserved`(사용자 수정), `conflict`, `remove_blocked`. 그 밖의 파일은 읽지 않는다. 폴더·파일이 링크면 그 host 는 `blocked`. 아무것도 쓰지 않는다. 결과: `schema_version`, `home`, `package_digest`, `agents_digest`(원본 폴더의 트리 digest, 폴더가 없으면 `null`), `hosts`(`host`, `dir`, `preflight`, `blocked_reason`, `files`, `not_supported`), `plan_digest`.

`HOME/.claude`·`HOME/.codex` 는 `home-routing-plan` 과 같은 규칙으로 `CLAUDE_CONFIG_DIR`·`CODEX_HOME` 폴더가 될 수 있다(`src/host_home.rs`, 카드 C-79). OMP 는 옮김 변수가 문서에 없어 언제나 `HOME/.omp`. 링크 검사는 도구 폴더부터 에이전트 폴더까지 한 단계씩이다. 옮겨진 도구 폴더의 `agents` 가 링크면, 계정 홈의 같은 기본 폴더(`HOME/.claude/agents`·`HOME/.codex/agents`)와 실제 경로가 같을 때만 따라가서 `dir` 이 그 기본 폴더가 되고, 다른 곳으로 가는 링크는 링크 경로를 말하며 `blocked`(사용자 결정 2026-11-17 "기본 폴더 링크만 허용", SPEC 결정 D10).
