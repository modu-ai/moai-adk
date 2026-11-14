# `hygiene-apply`

operation `hygiene.apply` · 코드 `src/ops/hygiene_apply.rs`, `src/hygiene.rs` (`apply`) · Issue #2934 (S4b), #2930 R-4

`hygiene-apply PROJECT --integration-ref REF --plan-digest <sha256:…> --permission-reference <ref> [--tmp DIR] [--home DIR] [--now UNIX]` 는 `hygiene-plan` 이 `hold` 로 낸 항목을 보관 폴더 `<HOME>/.moai/holding/<UTC 날짜>/` 로 옮긴다. 입력은 `hygiene-plan` 과 같고(`--home` 은 보관 폴더의 HOME, 기본 HOME), `worktree-gc-apply` 와 같은 약속을 따른다.

- 쓰기 전에 거부: 권한 참조가 비었거나, 같은 입력으로 다시 계산한 계획의 `plan_digest` 가 `--plan-digest` 와 다르거나(사이에 사용 중이 된 항목도 계획이 달라져 여기서 걸림), `holding`·날짜 폴더·`manifest.tsv` 가 심볼릭 링크면 아무것도 쓰지 않고 오류.
- 항목마다 옮기기 직전 다시 확인: 그 안을 작업 폴더로 둔 프로세스, 명령 줄(Linux 는 환경도)에 그 경로를 담은 프로세스(`hygiene-plan` 과 같은 판정), 그 안의 파일을 열어 둔 프로세스(`lsof -Fn`, Linux 는 `/proc/*/fd`). 하나라도 있거나 읽지 못하면 그 항목은 건너뜀(`in_use`, `file_open_in_use`, `in_use_unknown`, `open_file_check_unavailable`).
- 기록 먼저: 첫 이동 전에 `manifest.tsv`(머리줄 `original\tholding\tkind\tbytes\treason`)를 만들고, 항목마다 줄을 덧붙여 디스크에 내린 뒤에 옮긴다. 그래서 중간에 멈춰도 옮겨진 항목은 모두 manifest 에 있다. 줄을 못 쓰면 그 항목은 옮기지 않음(`manifest_write_failed`).
- 옮기기: 같은 파일 시스템 안의 `rename` 하나로 `<날짜 폴더>/<n>-<이름>`. 복사 뒤 지우기는 하지 않는다. 다른 파일 시스템(`other_file_system`)이거나 rename 이 실패하면 건너뛰고 그 manifest 줄을 되돌린다.
- 옮긴 `worktree` 가 있으면 프로젝트에서 `git worktree prune` 을 돌려 git 목록에서 뺀다(`pruned`).
- 결과: `holding`, `manifest`, `moved`[{`original`, `holding`, `kind`, `bytes`, `reason`}], `skipped`[{`path`, `kind`, `reason`}], `pruned`, `status`(`unchanged`·`applied`·`partial`), `writes`. 건너뛴 항목은 다시 시도하지 않는다. 지우기는 `hygiene-purge`(S4c) 몫.
