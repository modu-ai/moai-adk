# `audit-slot`

operation `audit.slot` · 코드 `src/ops/audit_slot.rs`, `src/audit_slot.rs` · SPEC-20261123-audit-slot

감사는 고정 감사 자리 `.claude/worktrees/audit-1`…`audit-3` 에서 한다. 새 격리 작업트리나 여러 트리가 함께 쓰는 빌드 폴더는 쓰지 않는다. 감사할 head 로 자리를 옮기고, 레인이 이미 만든 빌드를 물려받아 head 를 처음부터 다시 빌드하지 않는다.

- `audit-slot prep SLOT --head COMMIT [--lane DIR] [--project DIR]`: 자리가 없으면 `worktree-slot ensure` 와 같은 방식으로 만든다(상한·디스크 여유 검사 포함). 그다음 `git switch --detach` 로 head 에 둔다. 빌드는 이렇게 고른다.
  - `--lane` 이 이 저장소의 작업트리이고, 그 HEAD 가 감사할 head 이고, 추적 파일 변경이 없고, `target` 이 진짜 폴더이면: 레인의 `target/` 을 자리 안 `.target-inherit-<pid>` 로 복제(macOS 는 `cp -c -R -p`, 다른 곳은 `cp -R -p`)한 뒤 자리의 `target/` 과 바꾼다. 답의 `build.source` 는 `lane` 이다. 복사가 끝난 뒤 레인의 HEAD 를 다시 읽어 바뀌었으면 복사본을 지우고 실패한다. 실패하면 자리의 원래 `target/` 은 그대로 남는다.
  - 레인이 없으면: 새로 만든 자리는 `seed` 에서 데우고(`seed`), 이미 있는 자리는 자기 `target/` 을 그대로 쓴다(`kept`). 둘 다 없으면 `none` 이다.
- 거부(아무것도 바꾸지 않음): 감사 자리 이름이 아니면(`lane-01`, `seed`, `audit-4`, `../x` 등), head 가 커밋이 아니면, 레인이 쓸 수 없으면(이 저장소의 작업트리가 아님, 다른 head, 추적 파일 변경, `target/` 없음, `target` 이 링크나 파일), 자리에 이전 감사의 변경이 남아 있으면 거부한다. 메시지에 이유와 파일 이름이 나온다. 레인 검사는 자리를 만들기 전에 한다.
- `audit-slot done SLOT [--discard] [--project DIR]`: 변경이 없는 자리는 바로 끝난다. 추적 파일 변경(뮤턴트, 스테이지한 새 파일)이나 `target/` 밖의 추적 안 된 파일이 있으면 그 이름을 대며 거부한다. `--discard` 를 주면 자리 안에서만 `git reset --hard HEAD` 로 추적 파일을 되돌리고, 추적 안 된 파일을 지우고, `target/` 은 남긴다. 답의 `reverted`·`removed` 가 그 경로들이고, `clean`·`target_kept` 가 결과다. 무시된 파일(`target/` 말고)은 그대로 둔다.
- 둘 다 `worktree-slot` 과 같은 잠금(git 공통 디렉터리의 `moai-worktree-slot.lock`)을 잡고 자리를 바꾼다.
- 물려받은 빌드에서 다시 컴파일되는 것(macOS 에서 잼, 테스트 `a_slot_prepared_from_the_lane_does_not_compile_the_dependency`): 저장소 밖 의존 크레이트는 `Fresh` 다. 우리 크레이트는 한 번 `Dirty` 가 된다. head 로 전환할 때 git 이 쓴 소스 파일의 수정 시각이 레인 빌드보다 새롭기 때문이다. 그래서 head 에서 증분 빌드가 한 번 일어나고, 그 뒤 뮤턴트 한 줄은 우리 크레이트만 다시 컴파일한다.
- 아직 안 하는 것: 레인의 nextest 아카이브 재사용(테스트 135개 파일이 컴파일 시점 `env!` 로 경로를 읽어서, 실행 시점 경로로 바꾼 뒤), 판정 노트를 쓸 때 증거 로그 복사. `build.rs` 가 만드는 파일에는 절대 경로가 없다: ops 목록은 이름만 담고, bundle-packs 는 `env!("CARGO_MANIFEST_DIR")` 를 컴파일 때 풀어 그 자리의 파일을 읽는다. 그래서 복사한 빌드가 레인 쪽 파일을 가리키지 않는다.
