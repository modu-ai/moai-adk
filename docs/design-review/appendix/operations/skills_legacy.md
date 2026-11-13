# `skills-legacy`

operation `skills.legacy` · 코드 `src/ops/skills_legacy.rs`, `src/skills_legacy.rs` · Issue #2592

`skills-legacy HOME` 은 도구별 사용자 스킬 폴더(`.agents/skills`·`.claude/skills`·`.codex/skills`·`.omp/agent/skills`·`.cursor/skills`) 바로 아래에서 `moai` 를 뺀 `moai-*` 항목을 찾는다. 결과는 다음과 같다.

- `count` 와 `summary`(없으면 `0건: …`).
- 폴더별 `state`(`present`·`absent`·`link_not_followed`·`not_a_folder`·`unreadable`)와 `read_by`.
- `legacy` 항목마다 `folder`, `path`, `kind`, `skill_md`, `name`. `skill_md` 는 `file`·`absent`·`link`·`not_a_file`·`too_large`·`unreadable`, 폴더가 아니면 `not_checked` 다.
- Codex 가 읽는 `.agents/skills` 에서 `skill_md` 가 `file` 인 항목마다, 그 항목을 끄는 `[[skills.config]] path = "…/SKILL.md" enabled = false` 문구(`codex_disable_entries`). 경로는 TOML 기본 문자열로 escape 한다.

링크는 따라가지 않는다. HOME 과 스킬 폴더 사이의 폴더, 항목, 항목의 `SKILL.md` 모두 마찬가지다. `SKILL.md` 는 64 KiB 이하의 일반 파일일 때만 읽는다(링크를 따라가지 않고, 막히지 않게 연다). 그래서 FIFO 하나로 멈추지 않고, HOME 밖 파일 내용이 출력에 나오지 않는다(#2592 사전 감사 6852075845 B1·B2). 읽기만 하고, 옮기거나 지우거나 설정을 쓰지 않는다.
