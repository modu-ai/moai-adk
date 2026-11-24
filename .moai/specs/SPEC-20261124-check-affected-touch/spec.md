---
id: SPEC-20261124-check-affected-touch
title: "check affected 가 빌드 전에 바뀐 파일의 시각을 갱신해 옛 파일 시각으로 되돌린 복구가 낡은 빌드를 남기지 않아요"
status: approved
revision: 2
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 8002 (2026-11-23T23:29:48Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 8000·8001 에 답함): 개정 2 승인, 커밋 0cfb2e5577, 선택지 (a); tests/agent_defs_package.rs 는 범위에 더했고 source-manifest.json 은 줄 임대(card line-lease)로 써요; 개정 1 의 조건(링크 대상 mtime 그대로)은 그대로예요; 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:49c100611bf42d78375e0bd7ef4cb159ecdf186ea011c31f9a5a14bedb26171c"
---

# check affected 가 빌드 전에 바뀐 파일의 시각을 갱신해 옛 파일 시각으로 되돌린 복구가 낡은 빌드를 남기지 않아요

## Interview

- none: 리더 배차가 요구사항을 전달했어요 — 카드 C-433 본문(재발 방지: 빌드 전에 바뀐 파일의 시각 갱신, 스킬 글의 복구 규칙)과 배차 지시가 범위를 정했고, 별도 사용자 인터뷰는 없었어요.

## Intent

같은 원인이 두 번 레인의 영수증을 망쳤어요(카드 C-433 본문). 레인이 guard-mutant 를 돌린 뒤 원본을 `cp -p` 로 되돌리면 파일 내용은 원본으로 돌아오지만 수정 시각도 옛 값으로 돌아와요. cargo 는 소스 파일의 수정 시각이 마지막 빌드의 dep-info 파일보다 새로울 때만 다시 빌드하므로(내용 해시는 보지 않아요), 되돌린 파일은 "안 바뀐 것"으로 읽혀 마지막 mutant 가 든 빌드가 그대로 남아요. `moai check affected` 는 그 낡은 빌드를 시험해서 2 of 3985 테스트가 실패했어요(소스는 `cmp` 로 같고 `git status` 도 깨끗했어요). 첫 번째 발생은 메모리 `restore_mtime_stale_build` 에 있어요.

이 SPEC 은 원인을 두 곳에서 닫아요.

- 제품 쪽: `moai check affected` 가 단계를 시작하기 전에 `base...head` 에서 바뀐 파일(체크아웃에 있는 일반 파일)의 수정 시각을 지금으로 갱신해서, cargo 가 그 파일을 반드시 새것으로 봐요. 파일 내용·권한·git 추적 내용은 바꾸지 않아요.
- 안내 쪽: 스킬의 guard-mutant 안내(`skills/moai/references/internal/self-check.md` 의 Tests 2번)가 복구는 `git checkout -- <path>` 나 그냥 `cp` 로 하고 `cp -p` 는 쓰지 말라고 적어요.

코드에서 확인한 사실(2026-11-24, NeXT `4872c62870`):

- `src/check_affected.rs` 의 `run` 은 `read_facts` 로 `facts.changed`(`git diff --name-only --no-renames -z base...head`)와 `facts.present`(그 중 체크아웃에 파일로 있는 것)를 만들고, 깨끗함 검사(`git status --porcelain --untracked-files=no`) 뒤에 단계 반복문을 시작해요. 시각을 건드리는 코드는 없어요.
- `cargo fmt`, `cargo clippy`, 테스트 단계는 모두 같은 `<root>/target` 을 써요(`CARGO_TARGET_DIR` 가 없을 때). 그래서 이전 빌드의 낡은 결과가 그대로 쓰여요.
- guard-mutant 안내는 `skills/moai/references/internal/self-check.md` 의 Tests 2번과 `skills/moai/references/flows/next-flow.md:23` 에 있고, 복구 방법은 어디에도 적혀 있지 않아요.

## Non-goals

- 바뀌지 않은 파일(범위 밖)의 복구: 범위 밖 파일을 mutant 로 만들었다면 그 파일을 `base...head` 가 모르므로 갱신하지 않아요. 안내(REQ-4)가 그 경우를 줄여요.
- `--list` 실행의 변화: `--list` 는 읽기 전용이라 시각을 건드리지 않아요(REQ-3).
- cargo 의 빌드 폴더를 지우거나 `cargo clean` 하는 일: 빌드 결과를 지우지 않고 시각만 갱신해요.
- `git diff --check` 와 `cargo fmt` 단계의 변화, 요약 줄(`check-affected base=… result=…`)의 모양 변화: 요약 줄은 그대로예요(영수증 판정이 읽어요).
- 레인이 mutant 를 만드는 방법 자체나 `tools/self_check.sh` 의 변경.

## Requirements

- REQ-1: WHEN `moai check affected` runs a range, it shall set the modification time of every regular file that `base...head` lists as changed and the checkout holds to the current time before its first step, with the bytes, the mode and the git state of those files left as they were.
- REQ-2 (REQ-1): IF a changed path is no regular file (a link, a folder) or its time cannot be set, THEN `moai check affected` shall skip that path without following it, name it in the answer's `refresh.skipped`, and run its steps as before.
- REQ-3: WHEN `moai check affected` runs with `--list`, it shall set the time of no file · failure: none — `--list` reads git only and a time that was set could not be told from a change.
- REQ-4: The guard-mutant guidance in `skills/moai/references/internal/self-check.md` shall tell a lane to restore a mutated file with `git checkout -- <path>` or a plain `cp`, never with `cp -p`, with the reason in the same line.
- REQ-5 (REQ-4): A test shall read that guidance line and fail when the restore rule or the `cp -p` ban is missing from it.

## Acceptance

- AC-1 (REQ-1): in a temporary crate, a source file changed in the range that was built, replaced by a failing mutant, built again and restored with its old time gives a `check affected` run whose test step exits 0 (the restored source was rebuilt) and whose summary ends in `result=PASS`. · verify: `cargo nextest run --offline --locked --test check_affected_touch -E 'test(a_restore_with_the_old_file_time_is_rebuilt)'`
- AC-2 (REQ-1): a run moves the time of a changed file to a time after the run started, keeps its bytes and mode, and leaves the time of a file outside the range alone. · verify: `cargo nextest run --offline --locked --test check_affected_touch -E 'test(the_run_refreshes_only_the_changed_files_time)'`
- AC-3 (REQ-2): a changed path that is a symlink keeps its own time and its target's time, is named in `refresh.skipped`, and the run still finishes with the same steps. · verify: `cargo nextest run --offline --locked --test check_affected_touch -E 'test(a_changed_link_is_skipped_and_named)'`
- AC-4 (REQ-3): `check affected --list` leaves the time of every changed file as it was. · verify: `cargo nextest run --offline --locked --test check_affected_touch -E 'test(list_mode_touches_no_file)'`
- AC-5 (REQ-4, REQ-5): the guard-mutant line of `self-check.md` names `git checkout`, a plain `cp` and the ban of `cp -p` with its reason, and the pin test fails when one of them is removed. · verify: `cargo nextest run --offline --locked --test agent_defs_package -E 'test(the_guard_mutant_line_names_the_restore_rule)'`

## Scope

- write: `.moai/specs/SPEC-20261124-check-affected-touch/spec.md`, `src/check_affected.rs`, `tests/check_affected_touch.rs` (new), `skills/moai/references/internal/self-check.md`, `skills/moai/references/source-manifest.json`, `tests/agent_defs_package.rs`
- read-only: `skills/moai/references/flows/next-flow.md`, `tests/check_affected_one_slot.rs`, `tests/common/`, `Cargo.toml`
- forbidden: `tools/repo_scans.json`, `CLAUDE.md`, the cargo and rustup state of the user's home

## Plan

1. 이 SPEC 초안을 커밋하고 리더 승인을 기다려요(`LANE-BLOCKED kind=decision`, digest 포함). 승인 전에는 코드를 쓰지 않아요.
2. `card scope` 로 쓰기 범위를 넓혀 달라고 `SCOPE-REQUEST` 를 적어요(`src/check_affected.rs`, `tests/check_affected_touch.rs`, `skills/moai/references/internal/self-check.md`).
3. RED: 새 테스트 파일을 먼저 써서 AC-1~AC-5 가 현재 트리에서 실패하는 것을 봐요(AC-1 은 낡은 빌드로 `result=FAIL`).
4. 구현: `refresh_changed(root, present)` 하나와 `run` 의 호출 한 줄, 답의 `refresh` 필드, 스킬 글 한 줄.
5. guard-mutant: 호출을 지우고(AC-1·AC-2 가 빨개져야 해요), 링크 건너뛰기를 지우고(AC-3), `--list` 에서도 갱신하게 하고(AC-4), 스킬 글의 금지 문구를 지워요(AC-5). 복구는 `git checkout -- <path>` 로 해요.
6. `moai check affected` 를 맨몸으로 한 번, READY 노트.

## Risks

- 갱신은 범위에서 바뀐 모든 파일을 다시 컴파일하게 해요. 따뜻한 빌드가 있는 레인에서도 그 파일이 든 크레이트는 한 번 더 빌드돼요. 범위가 넓으면(base 가 멀면) 시간이 더 들어요. 정확함이 속도보다 먼저예요(낡은 빌드로 영수증이 틀린 것이 이 카드의 원인이에요). 알아채는 길: 요약 줄은 그대로고, 답의 `refresh.touched` 가 갱신한 개수를 보여줘요.
- 범위 밖 파일의 mutant 복구는 이 SPEC 이 못 막아요(Non-goals). 안내(REQ-4)가 그 길을 줄여요.
- 갱신하는 시각이 파일 시스템의 시계 해상도보다 거칠면 같은 초 안의 빌드와 구분되지 않아요. macOS APFS·Linux ext4 는 나노초라 영향이 없어요; 낮은 해상도의 파일 시스템은 이 SPEC 이 다루지 않아요.
- 시각만 바꾸는데도 git 은 다음 `git status` 에서 그 파일을 다시 읽어 봐요(내용이 같아 깨끗해요). 깨끗함 검사는 갱신보다 앞에 있어서 이 실행의 판정에는 영향이 없어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 갱신은 `facts.present`(범위에서 바뀐 것 중 체크아웃에 있는 파일)만, 실행 직전에 한 번, 현재 시각으로 해요 | 카드 C-433 본문(바뀐 모든 파일을 마지막 빌드보다 새롭게); cargo 는 소스 mtime > dep-info mtime 일 때만 다시 빌드해요 | 호출 한 줄을 지워요 |
| D-2 | 갱신에 못 쓰는 경로(링크, 폴더, 시각 설정 실패)는 실패가 아니라 `refresh.skipped` 로 알려요 | 링크를 따라가면 트리 밖 파일의 시각이 바뀔 수 있어요; 한 경로 때문에 영수증 전체를 막지 않아요 | `skipped` 를 실패로 바꿔요 |
| D-3 | 시각은 읽기 전용 파일 핸들의 `File::set_modified` 로 바꿔요(내용·권한 불변) | Rust 1.94.1 표준 라이브러리; 새 의존성이 없어요 | 다른 방법(`touch`)으로 바꿔도 같은 결과예요 |
| D-4 | 요약 줄은 바꾸지 않고 답에 `refresh` 필드만 더해요 | 영수증 판정(`src/card/ready_receipt.rs`)이 요약 줄 모양을 읽어요 | 필드를 지워요 |
| D-5 | 안내는 `self-check.md` Tests 2번에 한 문장을 더해요; `next-flow.md:23` 은 그대로 둬요 | 카드가 말한 "guard-mutant 안내"는 mutant 를 만들고 되돌리는 방법이 있는 자리예요; 영수증 모양의 줄은 복구를 다루지 않아요 | 문장을 지워요 |
| D-6 | 개정 2: 핀 테스트(AC-5)는 `tests/agent_defs_package.rs` 에 두고, `source-manifest.json` 의 `self-check.md` sha256 한 줄을 새 값으로 바꿔요 | 새 테스트 파일이 스킬 글을 `include_str!` 로 읽으면 `tools/repo_scans.json` 의 `always` 스캔에 올려야 하는데 그 파일은 다른 카드가 쥐고 있어요(`repo_scans_table` 이 짚었어요); `agent_defs_package.rs` 는 이미 스킬 글 읽기로 등록돼 있어요. 스킬 글이 바뀌면 원본 목록의 sha256 이 낡아져요(`source_manifest_facts`, `product_cli` 가 짚었어요). 개정 1 승인(노트 7975)은 바뀌지 않은 부분의 근거예요 | 핀 테스트를 `tests/check_affected_touch.rs` 로 되돌리고 `repo_scans.json` 에 항목을 더해요 |

## Result

개정 2, NeXT `4872c62870` 에서 자른 브랜치예요(구현 커밋 `743f901265`, 수리 커밋 `afce2dd216`). 승인 근거는 개정 1 은 리더 결정 노트 7975(조건: 체크아웃 밖을 가리키는 경로와 링크 대상은 건드리지 않아요), 개정 2 는 노트 8002 예요. 개정 2 는 첫 `check affected`(head `1b07af8ef4`, 3894 중 3 실패)가 짚은 두 가지를 고쳤어요: `source-manifest.json` 의 `self-check.md` sha256 한 줄(줄 임대로 한 줄만 바꿈)과, 새 테스트 파일의 `include_str!` 읽기(핀 테스트를 `tests/agent_defs_package.rs` 로 옮김). 숫자는 이 결과를 위해 이 워크트리에서 돌린 실행에서 나왔어요.

- RED: 구현 전에 `check_affected_touch` 의 테스트 5개를 돌렸어요. 실제 cargo 테스트는 선행 조건(cargo 혼자서는 mutant 빌드를 그대로 써서 `cargo test --lib` 가 실패해요)을 넘고 `check affected` 가 `test=101 result=FAIL` 로 끝났어요. 링크·시각·스킬 글 테스트 3개도 실패했고, `--list` 테스트는 구현 전에도 통과했어요(아무것도 건드리지 않던 때라서; 아래 M4 가 지켜요).
- AC-1 met (REQ-1): `a_restore_with_the_old_file_time_is_rebuilt` 통과. 임시 크레이트를 빌드하고, 실패하는 mutant 를 빌드하고, 원본 바이트를 옛 시각(2001-09-09)으로 되돌린 뒤 `check affected` 가 `result=PASS` 로 끝나요(툴체인의 `cargo`, `clippy`, `rustfmt` 를 직접 써서 실제 빌드를 해요).
- AC-2 met (REQ-1): `the_run_refreshes_only_the_changed_files_time` 통과. 바뀐 파일 2개는 실행 시작 뒤 시각을 얻고(`refresh.touched=2`), 바이트와 모드(0o640)는 그대로이고, 범위 밖 `docs/keep.md` 는 옛 시각 그대로예요.
- AC-3 met (REQ-2): `a_changed_link_is_skipped_and_named` 통과. 링크 `link.md`, 링크 폴더 `a` 를 지나는 `a/b` 는 `refresh.skipped` 에 이름이 올랐고, 체크아웃 밖 대상 `target.md` 와 `dir/b` 의 시각은 그대로이고 링크 자신의 시각도 그대로예요. 리더 조건의 위험 경로 실행이에요.
- AC-4 met (REQ-3): `list_mode_touches_no_file` 통과.
- AC-5 met (REQ-4, REQ-5): `agent_defs_package` 의 `the_guard_mutant_line_names_the_restore_rule` 통과(`self-check.md` Tests 2번이 `git checkout -- <path>`, 그냥 `cp`, `cp -p` 금지와 이유를 말해요).
- 가드 뮤턴트(각각 `git checkout -- <path>` 로 복구, 복구 뒤 `git status` 깨끗, nextest exit 100): M1 `refresh_changed` 호출을 지움 → AC-1 과 AC-2 의 테스트가 실패, M2 링크를 따라가게 함(`plain` 과 `within` 검사를 뺌) → AC-3 실패, M3 링크 폴더를 지나는 경로를 건드리게 함(`within` 만 뺌) → AC-3 실패, M4 `--list` 도 갱신하게 함 → AC-4 실패, M5 스킬 글에서 `cp -p` 금지를 뺌 → AC-5 실패(핀 테스트를 옮긴 뒤 `agent_defs_package` 에서 다시 돌렸어요).
- 남은 일: 없어요. 요약 줄 모양은 그대로예요(D-4); 갱신한 개수는 답의 `refresh.touched` 가 보여줘요.
