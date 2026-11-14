# `project-skill-plan`

operation `project_skill.plan` · 코드 `src/ops/project_skill_plan.rs` · I5 #2242 (P04 HOLD)

`project-skill-plan PROJECT` 는 프로젝트 단위 `moai` 스킬 사본 계획을 읽기만으로 만든다. stdin 의 typed 요청(64 KiB 이하)을 받고 판정 봉투를 stdout 에 낸다. 쓰지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R2a).

결과의 `plan_digest`(#2936 R9 S5c)는 프로젝트(정규 경로)·패키지 digest·모든 항목·경고를 묶는다. 다른 프로젝트나 미리보기 뒤 바뀐 패키지에는 쓸 수 없다(`project-skill-apply --plan-digest`).
