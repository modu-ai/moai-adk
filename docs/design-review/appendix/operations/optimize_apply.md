# `optimize-apply`

operation `optimize.apply` · 코드 `src/ops/optimize_apply.rs`, `src/worktree_cap.rs` · SPEC-20261120-worktree-cap-and-cleanup

`optimize-apply PROJECT --integration-ref REF --pick PATH@ITEM_DIGEST (반복) --permission-reference R` 는 `optimize-plan` 의 `cleanup.candidates` 에서 사용자가 고른 worktree 만 지운다. `PATH` 와 `ITEM_DIGEST` 는 미리보기가 준 `path` 와 `item_digest` 다.

- 허가 참조가 비었거나 쓸 수 없으면 아무것도 지우지 않고 거절한다(참조는 허가 검증기가 아니고, 승인은 호스트와 사용자의 몫이다).
- 고른 항목마다 지우기 직전에 다시 판정한다. 후보가 아니면(`not_a_candidate`, 이유 목록 포함), 항목 다이제스트가 다르면(`digest_mismatch`), 지우기가 실패하면(`remove_failed`) 그 항목만 `refused` 에 이유와 함께 넣고 다음 항목으로 간다.
- 지우기는 `git worktree remove` 를 `--force` 없이 한 항목씩, 경로가 깊은 것부터(다른 worktree 의 `target/` 안에 중첩된 것이 먼저) 한다. 가지는 남는다.
- 끝나면 `<git 공통 폴더>/.moai-cli/worktree-cleanup/<now_unix>-<digest12>.json` 에 지운 항목의 경로·가지·HEAD·복구 명령(`git worktree add …`)을 쓴다. 기록을 쓸 수 없으면 지운 경로를 모두 대며 실패를 알린다. 무시 파일과 빌드 출력은 복구되지 않는다.
- 회수한 공간은 `df` 여유 바이트의 변화(`free_bytes_delta`)로만 말한다. 크기 합(`allocated_bytes`)은 APFS 복제 때문에 실제보다 클 수 있다.
