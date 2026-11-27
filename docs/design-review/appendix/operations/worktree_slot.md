# `worktree-slot`

operation `worktree.slot` · 코드 `src/ops/worktree_slot.rs`, `src/worktree_slot.rs` · SPEC-20261123-worktree-slots

작업트리를 고정 자리로 둔다: 본 체크아웃의 `.claude/worktrees/lane-01`…`lane-10`, `audit-1`…`audit-3`, `seed`. 일회용 작업트리는 하위 에이전트만 쓴다.

- `worktree-slot plan [--project DIR]`: 자리 14개를 순서대로, 있는지·HEAD·브랜치(또는 `detached`)·`target/` 유무와 함께 보여 주고, 상한(작업트리 25개, 디스크 여유 15%, `.claude/worktrees` 300 GB)을 잰 값과 함께 보여 준다. 300 GB 상한은 `du` 가 APFS 복제를 통째로 세서 아직 재지 않는다(`measured: false`). 아무것도 쓰지 않는다.
- `worktree-slot ensure SLOT --at COMMIT [--project DIR]`: 자리가 없으면 `git worktree add --detach` 로 그 커밋에 만든다(강제 없음). 자리에 `target/` 이 없고 `seed` 에 있으면 `seed/target` 을 복사해 데운다: macOS 는 APFS 복제(`cp -c -R -p`, 디스크를 거의 쓰지 않고 수정 시각을 지켜서 의존 크레이트는 `Fresh`, 우리 크레이트만 다시 컴파일), 다른 곳은 `cp -R -p`. 답의 `warmed` 가 `clone`·`copy`·`none` 을 알린다. 데우기가 실패하면 답 대신 명령이 실패한다. 이미 있는 자리는 체크아웃도 `target/` 도 그대로 두고 `present` 로 알린다.
- 자리를 새로 만들면 본 체크아웃의 로컬 제외 파일(`info/exclude`)에 `/.claude/worktrees/` 한 줄을 남겨요(git 이 이미 무시하면 쓰지 않아요). 쓰지 못해도 자리는 그대로 만들어지고 답의 `place_exclude_error` 가 알려요.
- 관찰(아직 거절하지 않음): `.claude` 나 `.claude/worktrees` 가 심볼릭 링크면 전처럼 진행하고 답에 `would_refuse` {`code`: `place_is_a_link`, `why`} 를 더해요. 자리를 만든 호출은 영수증 기록 `worktree-place.jsonl` 에 한 줄을 남기고, 남기지 못하면 `would_refuse_record_error` 가 알려요.
- 거부: 고정 자리 이름이 아니면(`lane-11`, `seed2`, `../x` 등), 연결된 작업트리가 이미 상한이면, 디스크 여유가 바닥 아래면, 같은 경로에 작업트리가 아닌 폴더가 있으면 아무것도 만들지 않고 거부한다. 메시지에 상한과 잰 값이 둘 다 나온다. `--max-worktrees`, `--min-free-percent` 로 상한을 바꿀 수 있다(테스트용 손잡이, 기본 25·15).
- 한 저장소에서 `ensure` 는 한 번에 하나만 돈다: git 공통 디렉터리의 `moai-worktree-slot.lock` 을 배타적으로 잡고(최대 300초 기다림) 확인·추가·데우기를 한다. 같은 자리를 동시에 `ensure` 하면 하나만 만들고 나머지는 `present` 로 답한다.
- 데우기는 자리 옆 `.target-warming-<pid>` 로 복사한 뒤, `cp` 가 성공했을 때만 `target/` 으로 이름을 바꾼다. 실패하면 임시 복사본을 지우고 명령이 실패하며, 자리에는 `target/` 이 없다. 다음 `ensure` 가 그 자리를 데운다(있는 자리라도 `target/` 이 없으면 데운다). 이미 `target/` 이 있는 자리에는 절대 복사하지 않는다.
- `seed` 가 이 저장소의 작업트리가 아니거나 `seed/target` 이 진짜 폴더가 아니면(링크, 파일) 아무것도 만들지 않고 거부한다. 등록은 됐는데 폴더가 사라진 자리는 `plan` 에서 `missing: true` 로 보이고 `ensure` 가 거부한다. `df` 가 실패하면 여유 비율을 모르므로 거부한다.
- 있는 자리에서 죽은 실행이 남긴 임시 복사본(`.target-warming-*`, `.target-inherit-*`, `.target-old-*`)은 `ensure` 가 잠금 안에서 지우고 이름을 답의 `cleaned` 에 담는다. 예외는 감사 자리 `audit-1`…`audit-3`(`audit_slot::AUDIT_SLOTS` 의 이름)다: 그 자리의 복사본은 `ensure` 가 그대로 두고, `audit-slot` 이 폴더를 읽어(git 이 무시하는 복사본도) 이름을 대고 `done --discard` 가 지운다. 그 자리에서 `cleaned` 는 항상 `[]` 다.
- 정리 규칙(고정 자리는 지우지도 줄이지도 않음, 안에 든 작업트리·클레임, 3일, 상한 초과)은 `worktree-gc-plan` 에 있다. 정리가 시작되는 때:
  - 착지: `card land` 가 카드를 닫은 뒤 그 카드 예약이 가리킨 작업트리만 `worktree_cleanup::run` 으로 정리한다(plan 을 만들고, 그 plan 의 digest 로 apply). 카드가 열린 채 남으면 하지 않는다. 정리가 실패해도 착지는 그대로이고 답의 `cleanup.error` 에 남는다. 새로 지울 수 있는 경로는 없다: 안전 기준과 `.claude/worktrees` 보존은 그대로다.
  - 브랜치 삭제: `reference-transaction` 훅이 `committed` 상태에서 지운 ref 이름 한 줄을 git 공통 디렉터리의 `moai/worktree-cleanup-queue` 에 덧붙이기만 하고 아무것도 지우지 않는다. 다음 정리가 큐를 읽는다. ref 가 아직 있으면 그 줄은 버리고, 없으면 그 브랜치를 쥐었던 카드의 작업트리를 같은 방식으로 정리한다.
  - 판정: 같은 함수를 `Trigger::Verdict` 로 부른다. 판정 명령이 아직 없어서 어느 명령에도 연결하지 않았다.
  - seed 다시 빌드와 진입 문은 다음 조각에서 더한다.
