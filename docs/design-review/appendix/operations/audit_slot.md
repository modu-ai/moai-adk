# `audit-slot`

operation `audit.slot` · 코드 `src/ops/audit_slot.rs`, `src/audit_slot.rs` · SPEC-20261123-audit-slot

감사는 고정 감사 자리 `.claude/worktrees/audit-1`…`audit-3` 에서 한다. 새 격리 작업트리나 여러 트리가 함께 쓰는 빌드 폴더는 쓰지 않는다. 감사할 head 로 자리를 옮기고, 레인이 이미 만든 빌드에서 의존 크레이트를 물려받아 head 를 처음부터 다시 빌드하지 않는다.

공통 옵션: `--project DIR`(저장소의 아무 체크아웃), `--max-worktrees N`(자리를 새로 만들 때의 작업트리 상한, 기본은 `worktree-slot` 과 같다), `--min-free-percent N`(자리를 새로 만들거나 데울 때의 디스크 여유 하한, 기본은 `worktree-slot` 과 같다).

- `audit-slot prep SLOT --head COMMIT --holder NAME [--lane DIR] [--pid PID] [--lease-secs N]`
  - `SLOT` 은 `audit-1`, `audit-2`, `audit-3`, 또는 `auto`(살아 있는 점유자가 없는 첫 자리)다. `COMMIT` 은 40자 소문자 커밋 id 여야 한다. 짧은 id, 브랜치, 태그, `HEAD` 는 저장소를 읽기 전에 거부한다.
  - 자리가 없으면 `worktree-slot ensure` 와 같은 방식으로 만든다(상한·디스크 여유 검사 포함). 상한이나 여유 때문에 거부되면 `invalid_input` 이다.
  - 점유자 기록을 git 공통 디렉터리의 `moai-audit-slots/<자리>.json` 에 남긴다: 이름, 호스트가 알려 주는 세션, `--pid`(주면), head, 시각, `lease_until`(`--lease-secs`, 기본 3시간, 1초~하루). 임대가 끝나지 않았고 `--pid` 를 줬다면 그 프로세스가 살아 있는 동안 살아 있는 점유자다. 다른 이름(또는 둘 다 세션이 있는데 다른 세션)이 오면 점유자를 대며 거부한다. 같은 점유자는 다시 prep 할 수 있다. 죽은 점유자의 자리는 넘겨받고, 답의 `took_over` 에 예전 기록을 담는다. `auto` 는 세 자리가 모두 점유 중이면 점유자를 모두 대며 거부한다.
  - 그다음 `git switch --detach` 로 head 에 두고 빌드를 고른다.
    - `--lane` 이 이 저장소의 작업트리이고, 그 HEAD 가 감사할 head 이고, 추적 파일 변경이 없고, `target` 이 진짜 폴더이면: 레인의 `target/` 을 자리 안 `.target-inherit-<pid>` 로 복제(macOS 는 `cp -c -R -p`, 다른 곳은 `cp -R -p`)한 뒤 자리의 `target/` 과 바꾼다. 답의 `build.source` 는 `lane` 이다. 복사가 끝난 뒤 레인의 HEAD 를 다시 읽어 바뀌었으면 복사본을 지우고 실패한다. 실패하면 자리의 원래 `target/` 은 그대로 남는다.
    - 레인이 없으면: `target/` 이 없는 자리는(새 자리든 이미 있는 자리든) `seed` 에 빌드가 있으면 거기서 데우고(`seed`), `target/` 이 있는 자리는 자기 빌드를 그대로 쓴다(`kept`). 둘 다 아니면 `none` 이다.
  - 빌드를 둔 뒤 자리의 추적 파일(`git ls-files`) 전부에 지금 시각을 수정 시각으로 준다(`build.sources_touched` 가 그 수). 레인이 HEAD 를 head 에 두고 추적 변경이 없어도, 그 `target/` 이 뮤턴트를 빌드한 결과일 수 있다. 소스 시각이 빌드보다 오래되면 Cargo 는 그 빌드를 `Fresh` 로 본다. 그래서 우리 크레이트는 감사할 소스로 한 번 다시 컴파일되고, 저장소 밖 의존 크레이트는 시각이 그대로라 `Fresh` 다.
  - 스위치나 빌드가 실패하면 점유자 기록은 그 전 상태로 되돌린다.
- `audit-slot done SLOT --holder NAME [--discard]`
  - 다른 살아 있는 점유자의 자리는 거부한다. 죽은 점유자의 자리는 넘겨받는다(`took_over`).
  - 변경이 없는 자리는 바로 끝난다. 추적 파일 변경(뮤턴트, 스테이지한 새 파일)이나 `target/` 밖의 추적 안 된 파일이 있으면 그 이름을 대며 거부한다. `--discard` 를 주면 자리 안에서만 `git reset --hard HEAD` 로 추적 파일을 되돌리고, 추적 안 된 파일을 지우고, `target/` 은 남긴다. 답의 `reverted`·`removed` 가 그 경로들이다.
  - 끝난 자리가 깨끗하면 점유자 기록을 지운다(`released`). 되돌린 뒤에도 변경이 남으면 0 이 아닌 상태로 실패하고 점유자는 그대로 둔다.
  - 무시된 파일(`target/` 말고)은 지우지 않는다. 사용자의 전역 ignore 파일이 맞추는 파일도 포함해서, 다음 감사가 그 파일을 본다.
- 죽은 prep 이 남긴 복사본(`.target-inherit-<pid>`, `.target-old-<pid>`, `.target-warming-<pid>`)은 파일 하나하나가 아니라 폴더 이름 하나로 보이고, `--discard` 는 폴더째 지운다. 이름은 `git status` 가 아니라 자리 폴더를 읽어 찾는다. 그래서 git 이 무시하는 복사본(`info/exclude` 의 `.target-*` 줄 등)도 거부 메시지에 이름이 나오고 `done --discard` 가 지운다. 복사본 이름이 링크이면 링크만 지우고, 가리키는 폴더는 건드리지 않는다. `worktree-slot ensure` 는 감사 자리의 이 복사본을 지우지 않는다(`cleaned` 는 `[]`): 그 이름을 대고 지우는 것은 이 명령이다. 감사 자리인지는 이름 목록 `AUDIT_SLOTS`(`audit-1`, `audit-2`, `audit-3`)로 판단한다.
- 거부(이유를 메시지에 대고, 아무것도 바꾸지 않음): 감사 자리 이름이 아님(`lane-01`, `seed`, `audit-4`, `../x` 등), head 가 40자 커밋 id 가 아니거나 커밋이 아님, 레인을 쓸 수 없음(이 저장소의 작업트리가 아님, 다른 head, 추적 파일 변경, `target/` 없음, `target` 이 링크나 파일), 자리에 이전 감사의 변경이 남음, 다른 점유자. `ensure` 가 자리를 만들거나 데운 뒤의 거부는 "nothing was changed" 대신 무엇을 했는지 말한다.
- 자리 확인(첫 쓰기 전, 그리고 잠금 안에서 다시): 자리 경로나 그 위의 `.claude`, `.claude/worktrees` 가 링크이면 거부한다. 자리 폴더가 그 자리 자신의 작업트리가 아니면 거부한다. 자기 작업트리란 그 안에서 git 이 자리 경로를 최상위로, 이 저장소의 git 공통 디렉터리를 공통 디렉터리로 대고, 자기 git 디렉터리의 `gitdir` 파일이 자리의 `.git` 을 가리키는 것이다. 그래야 평범한 폴더(git 이 위로 올라가 주 체크아웃에서 동작), 다른 저장소, 다른 작업트리의 git 디렉터리를 빌린 `.git` 에서 `switch`·`reset --hard` 가 돌지 않는다. `worktree-slot ensure` 도 이미 있는 자리를 데우기 전에 같은 확인을 한다.
- 자리 자신의 `target/` 확인(같은 두 자리, 점유자 기록을 쓰기 전): 자리에 `target/` 이 있는데 진짜 폴더가 아니라 링크나 파일이면 `--lane` 이 있든 없든 거부하고 아무것도 쓰지 않는다. 링크를 따라가면 `prep` 과 이어지는 빌드의 모든 쓰기가 가리키는 폴더(예: 레인의 빌드)로 가기 때문이다. 시드의 `target/` 과 같은 규칙이고 `symlink_metadata` 로 읽는다(코드 `real_target`, 시험 `prep_refuses_a_slot_whose_target_is_a_link_and_writes_nothing_through_it`). 위의 `--lane` 쪽 확인은 레인의 `target` 을 보는 별개의 규칙이다.
- 둘 다 `worktree-slot` 과 같은 잠금(git 공통 디렉터리의 `moai-worktree-slot.lock`)을 잡고 자리를 바꾼다. 기다리는 시간은 기본 300초이고, `MOAI_WORKTREE_SLOT_LOCK_WAIT_SECS` 로 줄일 수 있다(테스트용).
- 물려받은 빌드에서 다시 컴파일되는 것(macOS 에서 잼, 테스트 `a_slot_prepared_from_the_lane_does_not_compile_the_dependency`, 뮤턴트 빌드는 `an_inherited_mutant_build_is_compiled_again_from_the_head`): 저장소 밖 의존 크레이트는 `Fresh` 다. 우리 크레이트는 소스 시각을 갱신했으므로 한 번 `Dirty` 가 된다. 그 뒤 뮤턴트 한 줄은 우리 크레이트만 다시 컴파일한다.
- 아직 안 하는 것: 레인의 nextest 아카이브 재사용(테스트 135개 파일이 컴파일 시점 `env!` 로 경로를 읽어서, 실행 시점 경로로 바꾼 뒤), 판정 노트를 쓸 때 증거 로그 복사, 레인 `target/` 복제 전 디스크 여유 검사(SPEC Risks). `build.rs` 가 만드는 파일에는 절대 경로가 없다: ops 목록은 이름만 담고, bundle-packs 는 `env!("CARGO_MANIFEST_DIR")` 를 컴파일 때 풀어 그 자리의 파일을 읽는다. 그래서 복사한 빌드가 레인 쪽 파일을 가리키지 않는다.
