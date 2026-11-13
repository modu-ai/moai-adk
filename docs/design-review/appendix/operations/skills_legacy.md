# `skills-legacy`

operation `skills.legacy` · 코드 `src/ops/skills_legacy.rs`, `src/skills_legacy.rs` · Issue #2592

`skills-legacy HOME` 은 도구별 사용자 스킬 폴더(`.agents/skills`·`.claude/skills`·`.codex/skills`·`.omp/agent/skills`·`.cursor/skills`) 바로 아래에서 `moai` 를 뺀 `moai-*` 항목을 찾는다. 결과는 `count`, `summary`(없으면 `0건: …`), 폴더별 `state`(`present`·`absent`·`link_not_followed`·`not_a_folder`·`unreadable`)·`read_by`·`legacy`(`folder`, `path`, `kind`, `skill_md`, `name`), 그리고 Codex 가 읽는 `.agents/skills` 의 각 항목을 끄는 `[[skills.config]] path = "…/SKILL.md" enabled = false` 문구(`codex_disable_entries`) 다. 링크는 따라가지 않는다. 읽기만 하고, 옮기거나 지우거나 설정을 쓰지 않는다.
