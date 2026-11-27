---
id: SPEC-20261126-home-tidy-apply
title: "홈 moai 폴더 정리의 적용: 고른 항목만 압축 보관한 뒤 지우고, 목록 파일로 되돌려요"
status: draft
revision: 1
created: 2026-11-26
parent: SPEC-20261122-home-tidy
---

# 홈 moai 폴더 정리의 적용: 고른 항목만 압축 보관한 뒤 지우고, 목록 파일로 되돌려요

## Interview

- Q1: 사용자가 정한 것은요? -> A: 2026-11-22 카드 본문이에요. 홈 moai 폴더를 분석·정리·최적화해요. 정리 계획에서 사용자가 압축 보관(archive compressed), 삭제, 유지를 고르고 고른 것만 실행해요. 삭제보다 보관을 먼저 제안하고, 보관본에는 목록 파일이 있어 다시 꺼낼 수 있어요. 파괴적인 단계마다 사용자의 선택이 필요하고, 모든 실행은 사본에서 먼저 증명해요. 홈 moai 폴더와 프로젝트 밖은 건드리지 않아요.
- Q2: 어디까지 만들어졌나요? -> A: 조각 2(`tidy-inventory`, `tidy-plan`, 읽기 전용)가 head 73c70c2c03 으로 착지했어요. 적용·보관·복원이 없어 카드 기준 2 가 열려 있어요.
- Q3: 이 SPEC 을 쓰면서 사용자에게 새로 물은 것은요? -> A: 없어요. Decisions 의 D-1~D-3 은 승인할 때 고르는 선택지예요.

## Intent

`tidy-plan` 은 무엇을 보관하거나 지울 수 있는지 보여 주기만 해요. 이 SPEC 은 부모 SPEC-20261122-home-tidy 의 조각 3 이에요: 사용자가 고른 항목만 실행하는 `tidy-apply` 와, 보관본을 목록 파일로 되돌리는 `tidy-restore` 를 더해요. 원본은 보관본을 다시 읽어 모든 파일의 해시가 맞은 뒤에만, 목록에 있는 파일만 지워요. 읽는 사람: 사용자(승인), 레인, 감사자, 착지 담당. 부모의 요구사항과 용어(항목 class, 계획 digest)는 그대로 쓰고 여기서 반복하지 않아요.

## Non-goals

- `moai update` 의 `tidy` 블록과 미리보기(카드 기준 3), 세션 시작 알림, 프로필 키. 부모의 조각 4 예요. `src/update*.rs` 와 `src/profile.rs` 를 쓰지 않아요.
- 워크트리·빌드 폴더·evidence·receipts 처럼 git 증명이 필요한 항목, 깨진 등록, `claude rm`, `git worktree remove`. 계획이 `proof_pending` 으로 두는 것은 적용이 건드리지 않아요.
- A26(프로젝트 폴더 정리)과 `MEMORY.md` 수정.
- 보관본의 자동 삭제나 보관 기간 정책. 보관본은 이 변경이 지우지 않아요.
- 실제 HOME 에 대한 실행. 모든 테스트는 임시 HOME 에서 해요. 실제 HOME 에서의 실행은 사용자가 자기 선택으로 해요.
- Windows 에서의 동작 확인(not_run).

## Requirements

<!-- This SPEC adds destructive operations, not a check that blocks a user's work; the refusals below are the safety fences of the new operations. -->
- REQ-1: When `tidy-apply <home>` runs with one or more `--pick <path>`, `--confirm <digest>` and `--permission-reference <reference>`, the apply operation shall recompute the plan with the same threshold arguments and act only on the picked items, each by the action the recomputed plan gives it · failure: none — an item that is not picked is never touched.
- REQ-2 (REQ-1): IF no item is picked, a pick names no plan item or an item whose action is `keep`, the digest differs from the recomputed plan, the permission reference is not usable, or `<home>/.moai` or `<home>/.moai/archive` is a link or the latter is not a folder, THEN the apply operation shall refuse before any write, naming the cause · failure: none — the home folder is unchanged.
- REQ-3: When the apply operation acts on an `archive` item, the apply operation shall write one compressed archive and one `list.json` (relative path, type, size, sha256, mode and modification time of every file and folder) under `<home>/.moai/archive/<UTC time>/` with exclusive creation, folders at mode 0700 and files at mode 0600, read the archive back, compare every file with the list and with the source, and only then remove the original, removing only the files named in the list and a folder only when it is empty · failure: none — a file that appeared after the listing stays, with the folders that hold it, and is named.
- REQ-4 (REQ-3): IF writing the archive fails, a read-back hash or size differs, or the source changed since it was listed, THEN the apply operation shall remove nothing of that item's original, remove only the unfinished files this run created for that item, and report the item · failure: none — the other picked items are still run and the answer says the run is partial.
- REQ-5: IF a picked item holds a secret-shaped name, a link, a special file, a part that cannot be read, or passes the walk's entry or time limit, THEN the apply operation shall skip the whole item with the reason, open no secret-shaped file, follow no link in any component of a path it reads, writes or removes, and change nothing of that item · failure: none — the skipped item is named.
- REQ-6: IF a Go session may be live (its lock is held, a registered process is alive, or either cannot be told), THEN the apply operation shall skip every picked item of the retired-tool class with the reason `in_use` or `in_use_unknown` · failure: none — items of other classes are still run.
- REQ-7: When the plan lists `releases`, `db`, `backups` or `user-assets.json`, the plan operation shall keep the entry with the reason `go_run_data_kept` (Decision D-3) · failure: none — a pick of such an entry is the case of REQ-2.
- REQ-8: When the apply operation acts on a `delete` item, the apply operation shall remove it only as an empty folder, without any recursive removal · failure: none — a folder that is not empty at that moment stays and is reported.
- REQ-9: When `tidy-restore <home> --archive <UTC time>` runs with `--permission-reference <reference>` and optional `--pick <path>`, the restore operation shall verify every listed file of an item against the archive before writing any file of that item, create each file only where nothing exists, and set the listed mode and modification time on what it created · failure: none — an existing path is named, skipped and left as it is, and the rest of the pick is restored; the archive is kept.
- REQ-10 (REQ-9): IF the archive folder or the list is missing, not readable or a link, a listed path is absolute, holds `..`, leaves the item or is secret-shaped, the archive holds an entry the list lacks, or a hash or size differs, THEN the restore operation shall write no file of that item and report it · failure: none — nothing is written outside `<home>/.moai`.
- REQ-11: When the apply operation runs, the apply operation shall keep a receipt in the archive folder that is `pending` before the first removal and names each item's state after it, and answer the free bytes of the file system of `<home>` before and after the run · failure: none — where free space cannot be read the answer says `not_checked` and the run goes on.
- REQ-12: When either operation runs, the operation shall start no command and change nothing outside `<home>/.moai` · failure: none — a stand-in `git` and `claude` on the path record no call.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 HOME 에 후보가 둘일 때 하나만 고르면 그 항목만 바뀌어요. 선택 없음, 모르는 선택, `keep` 항목 선택, 계획 뒤에 바뀐 폴더의 옛 digest, 빈 값·`placeholder` 권한 참조, 링크인 `.moai/archive` 는 0 이 아닌 종료 코드이고 홈 전체의 목록·바이트·시각이 그대로예요 · verify: `cargo test --locked --offline --test home_tidy_cli apply_`
- AC-2 (REQ-3, REQ-4): 보관한 항목의 `list.json` 이 모든 파일의 경로·크기·sha256·mode·시각을 담고, 보관 폴더는 0700·파일은 0600 이며, 원본은 사라져요. 읽어 확인하기 전에 보관본의 한 바이트를 뒤집거나 원본을 바꾸면 원본이 하나도 안 지워지고 이번 실행이 만든 미완성 파일만 사라져요. 목록을 만든 뒤에 생긴 파일은 남아요 · verify: `cargo test --locked --offline --test home_tidy_cli archive_`
- AC-3 (REQ-5): 안에 mode 000 인 비밀 이름 파일, 홈 밖을 가리키는 링크, 이름 있는 파이프가 있는 항목은 이유와 함께 건너뛰고, 그 항목과 홈 밖 폴더가 그대로예요 · verify: `cargo test --locked --offline --test home_tidy_cli protected_`
- AC-4 (REQ-6): 살아 있는 pid 가 등록돼 있으면 은퇴 도구 항목이 `in_use`, 등록 파일을 읽을 수 없으면 `in_use_unknown` 으로 남고, 같이 고른 reports 항목은 보관돼요 · verify: `cargo test --locked --offline --test home_tidy_cli inuse_`
- AC-5 (REQ-7): `releases`, `db`, `backups`, `user-assets.json` 이 계획에서 `keep`/`go_run_data_kept` 이고 고르면 거절돼요 · verify: `cargo test --locked --offline --test home_tidy_cli kept_`
- AC-6 (REQ-8): 빈 은퇴 도구 폴더는 지워지고, 지우기 직전에 파일이 생긴 폴더는 파일과 함께 남아 보고돼요 · verify: `cargo test --locked --offline --test home_tidy_cli delete_`
- AC-7 (REQ-9, REQ-10): 보관 뒤 복원하면 바이트·mode·시각이 같고, 이미 있는 경로는 이름이 나오고 덮어쓰지 않아요. `..` 경로, 목록에 없는 항목, 해시가 다른 보관본은 그 항목의 파일을 하나도 쓰지 않아요. 원본 삭제가 중간에 멈춘 상태는 복원이 빠진 파일만 채워요 · verify: `cargo test --locked --offline --test home_tidy_cli restore_`
- AC-8 (REQ-11, REQ-12): 첫 삭제 전에 영수증이 `pending` 이고 끝나면 항목별 상태가 남으며, 답에 실행 전후의 남은 바이트가 있어요. `git`·`claude` 대역에는 호출이 없어요 · verify: `cargo test --locked --offline --test home_tidy_cli receipt_`
- AC-9 (REQ-1, REQ-9): 두 연산의 도움말이 무엇을 지우고 무엇이 필요한지 말하고, 도움말 표에 한 번씩 예시와 함께 있으며, 연산 문서가 코드와 맞고 SPEC 검사를 통과해요 · verify: `cargo test --locked --offline --test home_tidy_cli help_`, `cargo test --locked --offline --bin moai-cli help_rules`, `cargo test --locked --offline --bin moai-cli operations_doc`, `moai spec check SPEC-20261126-home-tidy-apply --project .`

## Scope

- write: `.moai/specs/SPEC-20261126-home-tidy-apply/**`
- write (승인 뒤): `src/tidy/` (`mod.rs`, `rules.rs`, `plan.rs`, `inventory.rs` 와 새 파일 `apply.rs`, `archive.rs`, `restore.rs`), `src/ops/tidy_apply.rs`, `src/ops/tidy_restore.rs`, `tests/home_tidy_cli.rs`, `docs/design-review/appendix/operations/tidy_apply.md`, `docs/design-review/appendix/operations/tidy_restore.md`, `docs/design-review/appendix/operations/tidy_plan.md`, `Cargo.toml`·`Cargo.lock`(D-1 이 크레이트를 고를 때만)
- shared (line lease): `src/help.rs`(`GROUPS`·`EXAMPLES` 행), `tests/fixtures/help/*.txt`(골든이 바뀔 때)
- read-only: `src/home_state.rs`(`is_secret`, `session_lock_held`, `live_sessions`), `src/archive.rs`, `src/update_cleanup.rs`, `src/contract.rs`(`permission_reference_ok`), 부모와 조각 2 SPEC
- forbidden: `src/update.rs`, `src/update_cleanup.rs` 의 변경, `src/profile.rs`, `src/hook_session_start.rs`, 실제 HOME, 이 목록 밖의 파일

## Plan

1. 이 SPEC 초안을 커밋하고 digest 를 `LANE-BLOCKED kind=decision` 으로 올려요. 레인은 승인을 기록하지 않아요. 승인 전에는 SPEC 폴더 밖에 쓰지 않아요.
2. 승인 뒤 범위를 넓히고, RED: `tests/home_tidy_cli.rs` 에 `apply_`, `archive_`, `protected_`, `inuse_`, `kept_`, `delete_`, `restore_`, `receipt_` 를 임시 HOME 으로 써요. 위험 경로는 임시 HOME 에서 실제 바이너리를 돌리는 Rust 통합 테스트로만 확인해요(리더 결정 10013). 실행 중간에 보관본이나 원본을 바꾸는 경우만 D-4 의 이음새를 써요.
3. GREEN: `archive.rs`(압축 쓰기·다시 읽기·목록) → `apply.rs` → `restore.rs` → 연산 두 파일과 문서. 상태 폴더 이름은 글자로 쓰지 않고 `home_root` 상수를 써요.
4. 새 가드 테스트마다 가드 뮤턴트 한 줄, `cargo fmt --all`, READY 에서 `moai check affected --card` 한 번, READY 노트. tier 2 라서 이 변경을 쓰지 않은 세션의 전체 감사를 받아요.

## Risks

- 삭제 버그는 사용자의 파일을 잃어요. 완화: 목록에 있는 파일만 하나씩 지우고 폴더는 비었을 때만 지워요(재귀 삭제 없음). 다시 읽기 확인 뒤에만 지우고, 원본이 바뀌면 멈춰요.
- 확인과 삭제 사이에 다른 프로세스가 파일을 바꿀 수 있어요. 삭제 직전에 파일마다 해시를 다시 보지만 틈이 0 은 아니에요. 은퇴 도구 항목은 Go 세션 확인으로 한 번 더 막아요.
- 압축 크레이트가 오프라인 cargo 캐시에 있는지는 확인하지 않았어요(캐시를 보는 명령을 가드가 거절했고 다시 하지 않았어요). 없으면 `Cargo.lock` 갱신에 네트워크 내려받기가 필요해요. 라이선스 확인은 구현 때 하고 Result 에 적어요.
- 계획의 변경(REQ-7)은 조각 2 의 `plan_` 테스트 픽스처를 고칠 수 있어요.
- 보관본은 압축해도 자리를 차지해요. 비어 가는 디스크에서는 쓰기가 실패하고 원본은 남아요(REQ-4). 보관본을 언제 지울지는 이 SPEC 이 정하지 않아요.
- 조각 2 의 AC-1 은 열기 호출을 직접 관찰하는 확인이 not_run 이었어요. 이 SPEC 의 AC-3 도 mode 000 대역으로 확인하고 열기 호출 자체는 관찰하지 않아요.

## Decisions

승인 전 제안이에요. 사용자나 위임받은 리더가 고르면 revision 을 올려 `approval` 을 기록해요(레인은 승인을 기록하지 않아요). 기본값은 권장안이에요.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 압축 형식(기본): 새 의존 크레이트 `tar` 와 `flate2`(순수 Rust 백엔드)를 정확한 버전으로 고정해 gzip tar 를 써요. 대안 A: 시스템 `tar` 명령을 실행해요(REQ-12 의 "명령 없음"을 바꿔야 하고 Windows 확인이 필요해요). 대안 B: 압축 없이 검증된 사본으로 보관해요(사용자의 "압축 보관" 결정을 바꾸므로 사용자만 고를 수 있어요) | 카드 본문이 압축 보관이고 부모 D-6 이 압축 tar(zstd, 없으면 gzip)예요. `Cargo.lock` 에 `tar`·`flate2`·`zstd` 가 없어요. 새 의존은 범위와 공급망의 선택이에요. 오프라인 캐시에 있는지는 확인하지 않았어요: 없으면 기본안은 크레이트 내려받기와 `Cargo.lock` 갱신(네트워크 한 번)을 포함하고, 그 뒤의 확인 명령은 `--offline` 으로 돌아요. 이 선택을 승인하면 그 내려받기도 함께 승인하는 것으로 적어 주세요 | 의존 두 줄과 보관 모듈을 되돌려요. 만든 보관본은 시스템 도구로도 풀려요 |
| D-2 | 새 연산은 `tidy-apply`, `tidy-restore`(하이픈 형태, 파일 하나씩)예요 | 조각 2 의 D-1 과 같아요. `src/main.rs` 를 건드리지 않아요 | `tidy` 앞 명령을 더해요 |
| D-3 | `releases`, `db`, `backups`, `user-assets.json` 은 적용하지 않고 계획에서 유지해요(기본). 대안: 부모 D-2 대로 은퇴 도구로 보고 Go 세션이 없을 때 보관을 허용해요 | `src/update_cleanup.rs` 의 `GO_RUNTIME` 은 이 이름들(`releases`, `db`, `backups`)을 절대 보관하지 않고, `user-assets.json` 은 `cleanup-plan` 이 파일 소유를 증명하는 목록이에요. 더 엄격한 규칙을 따라요. 대신 부모 Intent 가 후보로 센 옛 Go 릴리스(1.84 GiB)는 남아요 | `rules.rs` 의 목록에서 이름을 빼요 |
| D-4 | 실행 중간에 보관본·원본을 바꾸는 테스트만 라이브러리의 숨은 이음새(`src/archive.rs` 의 `copy_verified_with` 방식)를 임시 HOME 에서 불러요. 나머지는 모두 실제 바이너리예요 | 실제 바이너리만으로는 쓰기와 다시 읽기 사이에 바이트를 바꿀 수 없어요. 바이너리에 환경 변수 우회로를 두는 것보다 안전해요 | 이음새와 그 테스트를 지워요 |
| D-5 | 복원은 확인 digest 없이 권한 참조만 받고, 보관본을 지우지 않아요 | 복원은 아무것도 덮어쓰거나 지우지 않아요(REQ-9) | 확인 토큰을 더해요 |

## Result

2026-11-26, macOS(Darwin 27.2.0), 레인 `moai-adk-agent-irad`. 승인은 카드 원장의 노트 10214 에 있어요(revision 1). 실제 HOME 에는 실행하지 않았어요.

- 커밋: `a705b4ba53`(열기 호출 관찰 테스트), `437d111435`(RED), `2e674cfd2b`(구현), `696ecf9c84`(테스트 보강), `2bfc6ae1f4`(도움말 표와 골든), 그 뒤의 이 절 커밋.
- TDD: RED 는 `437d111435` 에서 `test-failure`(51개 중 28개 실패, 23개 통과)로 기록했고, GREEN 은 `2e674cfd2b` 에서 같은 검사가 통과했어요. 지우기 직전의 `in_use` 재확인 테스트가 없어 `tdd reopen` 으로 테스트를 더했어요.

확인 명령은 적힌 그대로 실행했어요.

- AC-1: met — `home_tidy_cli apply_` exit 0, 5개 통과
- AC-2: met — `home_tidy_cli archive_` exit 0, 6개 통과
- AC-3: met — `home_tidy_cli protected_` exit 0, 8개 통과(감사 뒤 수리 `eca9e19570` 에서 다시 실행)
- AC-4: met — `home_tidy_cli inuse_` exit 0, 3개 통과
- AC-5: met — `home_tidy_cli kept_` exit 0, 3개 통과
- AC-6: met — `home_tidy_cli delete_` exit 0, 3개 통과
- AC-7: met — `home_tidy_cli restore_` exit 0, 7개 통과
- AC-8: met — `home_tidy_cli receipt_` exit 0, 2개 통과
- AC-9: met — `home_tidy_cli help_` 2개, `--bin moai-cli help_rules` 16개, `--bin moai-cli operations_doc` 6개 통과(모두 exit 0), `moai spec check` 오류 0

- 추가 확인: `cargo nextest run` 의 `home_tidy_cli`·`ac_command_lint`·`state_folder_literal_scan` 80개 통과. 첫 실행은 `state_folder_literal_scan` 1건 실패(서식 정리가 `fx.moai` 의 `.moai` 를 한 줄에 남김)였고 줄 모양을 바꿔 고쳤어요.
- 카드 기준 1(열기 호출 직접 관찰): `inventory_and_plan_open_no_protected_entry` exit 0, 1개 통과. macOS `sandbox-exec` 가 막은 경로를 읽으려고 여는 순간 프로세스를 SIGKILL 해요. 막은 경로마다 실제로 여는 프로그램(`cat`, `ls`)이 죽는 것을 먼저 확인하고, 두 읽기 연산은 끝까지 돌아요. 스캔에 열기 호출을 넣은 뮤턴트는 exit 101 이었어요. 적용 쪽은 `protected_files_are_never_opened_by_the_apply` 가 같은 방법으로 봐요. Linux·Windows 에서는 not_run 이에요. 그래서 위 Risks 의 "열기 호출 자체는 관찰하지 않아요"는 이 결과로 바뀌었어요.
- 위험 경로 실행: 임시 HOME 에서 실제 바이너리와 D-4 이음새를 부르는 Rust 테스트로 링크(절대·형제 폴더), 이름 있는 파이프, mode 000 파일, 목록 뒤에 생긴 파일, 지우기 직전의 파일 변경·링크 바꿔치기·세션 시작, 보관본 훼손을 실행했어요.
- 가드 뮤턴트 25개(`m01`~`m25`)는 모두 해당 테스트 1개가 exit 101 이었어요. 링크 테스트 넷은 처음에 라이브러리(`cap-std`)가 대신 막는 링크만 써서, 우리 가드가 없으면 따라가는 형제 폴더 링크를 더한 뒤에 돌렸어요.
- 의존성(D-1): `tar =0.4.45`(`default-features = false`), `flate2 =1.1.10`(`rust_backend`, `miniz_oxide`). `Cargo.lock` 에 함께 들어온 것은 `filetime`·`miniz_oxide`·`adler2`·`crc32fast`·`simd-adler32` 예요. 라이선스는 `cargo metadata` 기준으로 모두 MIT·Apache-2.0·0BSD·Zlib 중에서 고를 수 있어요. 네트워크는 `cargo fetch` 한 번이었고 내려받은 것은 `tar 0.4.45` 하나였어요.
- SPEC 문장의 해석과 더한 것: Go 세션 확인(REQ-6)은 지우기 직전에 한 번 더 해요. 걸으면 한도를 넘는 은퇴 도구 폴더는 계획이 `keep` 이라 고르면 REQ-2 로 거절해요. 복원에서 이미 있는 폴더도 `existing` 에 이름이 나와요.
- not_run: 보관본 쓰기 실패(디스크 가득 등)의 직접 재현(같은 버림 경로는 다시 읽기 실패 테스트가 지나가요), 적용 도중의 시간 한도 초과, 전원 끊김, Linux·Windows. `src/tidy` 는 조각 2 부터 `cfg` 없이 unix API 를 써요: Windows 빌드는 확인하지 않았어요.
- 남은 위험: 마지막 확인과 삭제 사이의 틈은 0 이 아니에요(위 Risks).
- 카드 기준 3(`moai update` 미리보기)은 이 SPEC 의 범위 밖이고 후속 카드로 남아요.
- 감사 10420 의 N3 수리(리더 결정 10422, 커밋 `eca9e19570`): 비밀 이름 규칙을 `go-state` 미리보기와 같게 넓혔어요(`src/go_state.rs` 의 `secret_name`). `.pem`·`.key`·`.p12`·`.pfx` 파일, `.npmrc`·`.netrc`·`.pgpass`·`service-account.json`·`id_rsa`·`id_ecdsa`·`id_ed25519` 가 더해져요. 목록에 없는 이름(`id_dsa`, `.git-credentials` 등)은 여전히 비밀로 보지 않아요. 새 테스트 `protected_private_key_names_skip_their_item_whole`, `protected_private_keys_are_never_opened_by_the_apply`(macOS 전용)는 예전 규칙으로 되돌린 뮤턴트에서 둘 다 exit 101 이었고, `home_tidy_cli` 전체는 54개 통과예요.
