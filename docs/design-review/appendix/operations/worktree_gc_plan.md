# `worktree-gc-plan`

operation `worktree.gc.plan` · 코드 `src/ops/worktree_gc_plan.rs` · WT-4 #2380

`worktree-gc-plan PROJECT --integration-ref REF` 는 연결된 worktree 마다 `remove`·`trim`(cargo 빌드 출력만)·`skip` 과 이유, digest 를 보여 준다. 지우지 않고, fetch 하지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R2c).

고정 자리와 상한(SPEC-20261123-worktree-slots 조각 2, 리더가 기록한 결정의 (b)안): 아래 규칙은 건너뛰거나 알려 주기만 하고, 새로 지울 수 있게 되는 경로는 없다. `.claude/worktrees/` 보존(Q8)은 그대로다.

- `fixed_slot`: 본 체크아웃의 `.claude/worktrees/` 에 있는 고정 자리(`lane-01`…`lane-10`, `audit-1`…`audit-3`, `seed`)는 지우지 않는다.
- `nested_worktree`: 안에 다른 연결 worktree 가 든 worktree 는 건너뛴다. 지우면 안쪽 체크아웃도 함께 사라진다.
- `live_claim`: 활성 카드 예약(`held`, `expired_candidate`, `held_for_review`)이 가리키는 worktree 는 건너뛴다. 임대가 끝난 `held` 예약도 센다. 긴 실행 중에 임대가 끝날 수 있어서다.
- `claim_check_unavailable`: 원장 파일은 있는데 읽지 못하면, 예약이 없다고 단정할 수 없으므로 연결 worktree 를 모두 건너뛴다. 프로필이나 원장이 없는 프로젝트에는 예약이 없다.
- `cap`: `limit`(`--max-worktrees`, 기본 25), `count`(연결 worktree 수), `over`, `oldest_first`(`remove` 후보를 오래된 변경부터), `needed`(상한 아래로 돌아오는 데 필요한 개수), `enough`. 보고용이라 digest 에 묶이지 않는다.
