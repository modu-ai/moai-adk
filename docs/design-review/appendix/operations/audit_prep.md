# `audit-prep`

operation `audit.prep` · 코드 `src/ops/audit_prep.rs`, `src/audit_prep.rs` · SPEC-20261124-audit-prep

감사를 카드 이름으로 시작해요. 고정 감사 자리 `.claude/worktrees/audit-1`…`audit-3` 중 하나를 카드의 자리로 잡고, 감사할 head 로 옮기고, 레인이 이미 만든 빌드를 이어받고, 바꾼 대상만 증분 빌드를 한 번 돌려요. 자리와 빌드를 다루는 일은 `audit-slot prep` 이 하고, 이 명령은 카드를 holder 로 불러 주고 빌드를 더해요. 감사 자리 단위의 일(리더의 `--discard` 정리 등)은 `audit-slot` 이 그대로 해요.

`audit-prep CARD --head COMMIT [--lane DIR] [--test NAME]… [--pid PID] [--lease-secs N] [--project DIR] [--max-worktrees N] [--min-free-percent N] [--slot-dir DIR] [--max N] [--min-free-gb N]`

- `CARD` 는 `C-` 와 숫자(`C-364`)예요. 그 카드가 자리의 holder 예요. 세션은 비교하지 않아요: 리더가 준비하고 감사 에이전트가 `audit-done` 을 쳐도 같은 holder 예요. 그래서 holder 기록의 세션은 비어 있어요.
- `--head` 는 40자 소문자 커밋 id 여야 해요(짧은 id, 브랜치, 태그는 거부).
- 자리: 카드가 이미 자리를 쥐고 있으면 그 자리를 다시 준비해요(앞 자리가 비어 있어도 옮기지 않아요). 없으면 살아 있는 holder 가 없는 첫 자리를 잡고, 셋이 모두 쥐여 있으면 holder 를 모두 대며 거부해요. 거부하면 빌드는 돌리지 않아요.
- 자리 쪽 동작은 `audit-slot prep` 과 같아요: 자리 확인(링크, 자기 작업트리), 이전 감사가 남긴 변경 거부, `--lane` 의 `target/` 복제(`cp -c -R -p`), 추적 소스 touch(레인의 뮤턴트 빌드가 head 로 통과하는 일을 막음). 자세한 것은 `audit_slot.md`.
- 빌드: 자리 안에서 `cargo test --no-run --locked --offline --message-format json` 에 `--test NAME`(최대 두 개, 셋이면 아무것도 바꾸기 전에 거부) 또는 `--lib` 를 붙여 한 번 돌려요. `CARGO` 환경 변수가 있으면 그 cargo, 없으면 `cargo` 예요. `RUSTC_WRAPPER`, `CARGO_BUILD_RUSTC_WRAPPER`, `CARGO_TARGET_DIR`, `CARGO_BUILD_TARGET_DIR` 는 지우고 자리 자기 `target/` 에 빌드해요(증분 빌드와 sccache 가 같이 켜지면 거부해요, 2026-11-22 감사에서 exit 101). `CARGO_BUILD_JOBS` 가 없으면 무거운 자리 수로 나눈 값을 넣어요.
- 무거운 자리: `--slot-dir`(기본은 이 컴퓨터의 자리 폴더), `--max`(기본 4, 리더의 최신 줄을 따라요), `--min-free-gb`(기본 50) 로 `check affected` 와 같은 규칙으로 하나 잡고 빌드가 끝나면 놓아요. 기다린 줄은 `<자리>/target/audit-prep-<카드>.heavy-slot.log` 에 남아요.
- 답: 자리의 답(`slot`, `path`, `head`, `build`, `holder`, `took_over`, `clean`)에 `card` 와 `cargo` 가 붙어요. `cargo` 는 `exit`, `log`(cargo 의 표준 오류), `units`(`name`, `kind`, `workspace`, `fresh`, `executable`), 그리고 개수 `compiled`, `fresh`, `compiled_workspace`, `compiled_dependencies` 예요. `workspace` 는 그 단위의 `Cargo.toml` 이 자리 안에 있다는 뜻이에요. 이 저장소 밖 의존 크레이트가 `fresh` 이고 `compiled_dependencies` 가 0 이면 이어받은 빌드가 쓰였다는 뜻이에요.
- 빌드가 실패하면 0 이 아닌 상태로 끝나고, 컴파일러의 마지막 줄을 메시지에 담고, 자리는 카드가 쥔 채로 남아요. 감사자는 그래도 `audit-done` 을 돌려요.
- 컴파일 때 박히는 경로(`env!("CARGO_MANIFEST_DIR")` 등)는 고치지 않아요. 추적 소스를 touch 했으므로 워크스페이스 단위가 자리의 소스로 다시 컴파일되고, 다시 빌드한 테스트 실행 파일은 레인이 아니라 자리 폴더를 읽어요(시험 `a_rebuilt_test_executable_reads_the_slot_and_holds_no_byte_of_the_lane_folder`).
- 하지 않는 것: 감사의 테스트나 뮤턴트를 돌리는 일, head 의 `check affected` 를 다시 돌리는 일(감사자는 레인의 영수증을 인용해요), 레인과 nextest 아카이브를 공유하는 일(테스트 163개 파일이 컴파일 때 `env!` 로 경로를 읽어서, 실행 시점 경로로 바꾼 뒤).
