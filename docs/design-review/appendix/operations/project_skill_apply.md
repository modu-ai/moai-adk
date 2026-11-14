# `project-skill-apply`

operation `project_skill.apply` · 코드 `src/ops/project_skill_apply.rs` · I5 #2242

`project-skill-apply PROJECT` 는 준비된 `moai` 스킬 패키지를 프로젝트의 host 폴더로 복사한다. stdin 의 정확한 요청을 받고, `opt_in` 이 있어야 하며, 호출자가 준 프로젝트 루트 아래에만 쓴다. `src/main.rs` 에서 그대로 옮김(#2834 R2a).
