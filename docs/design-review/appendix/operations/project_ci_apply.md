# `project-ci-apply`

operation `project.ci_apply` · 코드 `src/ops/project_ci_apply.rs`, `src/project_ci_plan.rs` · SPEC-20261122-project-ci-design (REQ-7)

`project-ci-apply PROJECT --plan-digest D --permission-reference R` 는 stdin 의 승인된 CI 계획을 적용한다. 권한 참조가 없거나, 지문이 다르거나, 미리 본 뒤 프로젝트가 바뀌어 계획이 낡았거나, 충돌이 있으면 아무것도 쓰기 전에 계획 전체를 거절한다.

- 쓰는 것: 계획의 `create`·`merge` 파일, 사용자가 고른 `remove`, 그리고 기준을 적은 `manifests/ci.json`. 쓰기 전 내용은 영수증에 보관하고 복구 연산이 되돌린다. 아무것도 쓰지 않은 실행은 영수증을 남기지 않는다.
- 결과: `request_id`, `written`, `removed`, `adopted_baselines`, `unchanged`, `preserved`, `manifest`.
