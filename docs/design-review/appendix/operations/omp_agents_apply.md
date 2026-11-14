# `omp-agents-apply`

operation `omp_agents.apply` · 코드 `src/ops/omp_agents_apply.rs`, `src/omp_agents.rs` · Issue #2936 (R9 S3·S5d)

`omp-agents-apply PROJECT PACKAGE AGENTS --plan-digest D --permission-reference P` 는 확인된 `omp-agents-plan` 의 `create`·`update` 파일을 `PROJECT/.omp/agents/` 에 쓰고, 쓴 파일의 sha256 을 상태 폴더 `manifests/omp-agents.json` 에 남긴다. 지금 계획의 `plan_digest` 가 `D` 와 다르면 `stale or altered plan; re-preview required`, permission reference 가 없으면 `permission reference required`, `preflight` 가 `blocked` 면 이유를 내고 거절하며, 모두 쓰기 전이다. 미리보기 뒤 파일이 바뀌었으면 그 파일 앞에서 멈춘다. 사용자가 고친 파일은 `preserved`·`conflict` 로 두고, 지우는 코드는 없다. 결과: `state`(`applied`·`unchanged`), `plan_digest`, `written`, `state_path`.
