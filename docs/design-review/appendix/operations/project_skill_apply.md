# `project-skill-apply`

operation `project_skill.apply` · 코드 `src/ops/project_skill_apply.rs` · I5 #2242

`project-skill-apply PROJECT` 는 준비된 `moai` 스킬 패키지를 프로젝트의 host 폴더로 복사한다. stdin 의 정확한 요청을 받고, `opt_in` 이 있어야 하며, 호출자가 준 프로젝트 루트 아래에만 쓴다. `src/main.rs` 에서 그대로 옮김(#2834 R2a).

`--plan-digest D --permission-reference P` (#2936 R9 S5c): 둘 다 주거나 둘 다 생략한다(하나만 주면 사용 오류, 아무것도 쓰지 않음). 주면 permission reference 가 있어야 하고 `D` 가 지금 `project-skill-plan` 의 `plan_digest` 와 같아야 쓴다 — 다르면 `stale or altered plan; re-preview required`, 모두 쓰기 전이다. 둘 다 생략한 옛 호출 모양은 그대로 동작한다.
