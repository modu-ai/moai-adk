# `optimize-plan`

operation `optimize.plan` · 코드 `src/ops/optimize_plan.rs`, `src/optimize.rs`, `src/worktree_gc.rs`(`plan_with`, `squash_merged`) · Issue #2724(부모 #2380)

`optimize-plan PROJECT --integration-ref REF [--build-root DIR]… [--observations FILE] [--retention-days N] [--recent-minutes N] [--budget-ms N]` 는 저장소에 남은 것을 한 표로 보여 준다. 읽기만 하고, 지우지도 쓰지도 않으며, 네트워크를 쓰지 않는다.

- 행 종류: `worktree`(연결 worktree — `worktree-gc-plan` 을 스쿼시 병합 판정을 켜고 다시 쓴다), `build_folder`(호출자가 `--build-root` 로 넘긴 절대 경로 폴더의 자식 폴더 — 기본 뿌리 없음, 링크·파일인 뿌리는 거절), `report`(`.moai-cli/reports`, 항상 `keep`).
- 행마다: `bytes`(+`bytes_partial`), `modified_unix`, `age_secs`, `retention_until_unix`, `action`(`remove`/`trim`/`skip`/`keep`), `reasons`, `restore`(복구 방법), `apply`(어느 명령이 적용하는지).
- 스쿼시 병합: HEAD 가 통합 ref 의 조상이 아니어도, 가지가 merge-base 이후 바꾼 경로(이름 바꾸기 끔)가 통합 tip 에서 HEAD 와 내용이 모두 같으면 `merged_by: squash`. 바꾼 경로가 없거나 git 이 실패하면 병합으로 보지 않는다(실패는 `merged_unknown`).
- `remove` 후보 조건: gc 판정이 `remove` 이고, 관측 파일의 `audits` 에 그 HEAD 40자와 같은 `PASS` 가 있을 것. 없으면 `audit_unconfirmed`(관측 자체가 없으면 `audits_unobserved`)로 `trim`(cargo 산출물이 있을 때) 또는 `skip`.
- `skip` 아닌 worktree 행은 `claims`·`holds` 관측이 있어야 한다. 없으면 `claims_unobserved`/`holds_unobserved`, `issue/<n>-…` 가지의 n 이 살아 있는 claim 이면 `live_claim`, HOLD 면 `hold` 로 `skip`(#2669 규칙: 로컬로 알 수 없는 것은 출처를 단 관측 파일에서 받고, 관측하지 않은 부분은 행을 남긴다).
- 빌드 폴더: cargo 소유 항목만 있고(다른 파일이 있으면 `other_files_in_folder`), 그 안에서 일하는 프로세스가 없고, 최근 변경(`--recent-minutes`) 밖이며 보존 기한(`--retention-days`, 기본 7일)이 지났을 때만 `trim`. 이 Issue 에서는 적용 단계가 없다.
- 기간 경과만으로 worktree 를 후보로 만들지 않는다: 보존 기한은 보여 줄 뿐이다.
- 적용: `optimize-plan` 은 아무것도 적용하지 않는다. worktree 는 `worktree-gc-apply` 가 자기 plan(조상 병합만 병합으로 봄)을 다시 계산해 사용자 확인 뒤에만 적용한다 — 그래서 스쿼시 병합만 된 worktree 는 거기서 `skip` 으로 남는다. `preview_digest` 는 행의 kind·path·head·action·reasons·trim_paths 를 묶는다.
