# `worktree-slot`

operation `worktree.slot` · 코드 `src/ops/worktree_slot.rs`, `src/worktree_slot.rs` · SPEC-20261123-worktree-slots

작업트리를 고정 자리로 둔다: 본 체크아웃의 `.claude/worktrees/lane-01`…`lane-10`, `audit-1`…`audit-3`, `seed`. 일회용 작업트리는 하위 에이전트만 쓴다.

- `worktree-slot plan [--project DIR]`: 자리 14개를 순서대로, 있는지·HEAD·브랜치(또는 `detached`)·`target/` 유무와 함께 보여 주고, 상한(작업트리 25개, 디스크 여유 15%, `.claude/worktrees` 300 GB)을 잰 값과 함께 보여 준다. 300 GB 상한은 `du` 가 APFS 복제를 통째로 세서 아직 재지 않는다(`measured: false`). 아무것도 쓰지 않는다.
- `worktree-slot ensure SLOT --at COMMIT [--project DIR]`: 자리가 없으면 `git worktree add --detach` 로 그 커밋에 만든다(강제 없음). 자리에 `target/` 이 없고 `seed` 에 있으면 `seed/target` 을 복사해 데운다: macOS 는 APFS 복제(`cp -c -R -p`, 디스크를 거의 쓰지 않고 수정 시각을 지켜서 의존 크레이트는 `Fresh`, 우리 크레이트만 다시 컴파일), 다른 곳은 `cp -R -p`. 답의 `warmed` 가 `clone`·`copy`·`none`·`failed` 를 알린다. 이미 있는 자리는 체크아웃도 `target/` 도 그대로 두고 `present` 로 알린다.
- 거부: 고정 자리 이름이 아니면(`lane-11`, `seed2`, `../x` 등), 연결된 작업트리가 이미 상한이면, 디스크 여유가 바닥 아래면, 같은 경로에 작업트리가 아닌 폴더가 있으면 아무것도 만들지 않고 거부한다. 메시지에 상한과 잰 값이 둘 다 나온다. `--max-worktrees`, `--min-free-percent` 로 상한을 바꿀 수 있다(시험용 손잡이, 기본 25·15).
- 정리 규칙(자리는 지우지 않음, 안에 든 작업트리·클레임, 3일, 상한 초과)과 정리 시점(착지, 판정, 브랜치 삭제 대기열, seed 다시 빌드)은 다음 조각에서 더한다.
