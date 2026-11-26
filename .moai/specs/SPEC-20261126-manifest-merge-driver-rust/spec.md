---
id: SPEC-20261126-manifest-merge-driver-rust
title: "the source-manifest merge driver is a moai command, so a merge runs no Python"
status: draft
revision: 2
created: 2026-11-26
branch: spec/manifest-driver-rust-20261126
---

# the source-manifest merge driver is a moai command, so a merge runs no Python

## Intent

`skills/moai/references/source-manifest.json` 은 `.gitattributes` 가 합치기 드라이버 `moai-manifest` 를 지정한 파일이에요. 지금 그 드라이버는 `tools/merge_driver_manifest.py` 이고, 공유 체크아웃의 로컬 설정 `merge.moai-manifest.driver` 가 그 복사본을 `python3` 로 돌려요. 그래서 양쪽에서 이 파일이 바뀐 모든 merge 와 `merge-tree`(감사·착지 포함)가 Python 을 돌려요. 사용자 규칙은 Python 을 쓰지 않는 것이에요(2026-11-25).

이 SPEC 은 같은 일을 하는 Rust 명령 `moai manifest-merge-driver %O %A %B %P` 를 만들고, `.gitattributes` 가 새 드라이버 이름 `moai-manifest-rs` 를 가리키게 하고, 등록 안내를 그 명령으로 바꾸고, Python 드라이버를 없애요. 이름이 바뀌면 공유 설정의 옛 `merge.moai-manifest.driver` 줄은 새 `.gitattributes` 를 가진 체크아웃에서 더 불리지 않아요. 새 이름을 등록하는 일은 사람의 몫이에요.

## Non-goals

- 매니페스트를 쓰고 검사하는 `tools/source_manifest.py` 의 이관(`SPEC-20261121-rust-only-inventory` 의 몫). 이 SPEC 은 그 파일에서 드라이버를 말하는 글만 고쳐요.
- git 설정을 moai 가 쓰는 설치 명령. 등록은 문서에 적힌 `git config` 한 줄이에요.
- GitHub 서버의 merge(커스텀 드라이버를 돌리지 않아요).
- `git merge -X…` 옵션 감지(드라이버에 전달되지 않아요; 저장소의 매니페스트 시험이 안전망이에요).

## Requirements

- REQ-1: The `manifest-merge-driver` command shall take the four values git passes (`%O`, `%A`, `%B`, `%P`), write the merged manifest into the `%A` file and exit zero when it can decide every entry. · failure: none — an entry it cannot decide is REQ-2
- REQ-2 (REQ-1): IF the command cannot decide the merged manifest or fails unexpectedly, THEN the command shall leave git's text merge of the three manifests in the `%A` file, name the reason and the command `moai check fix` on standard error, and exit with a status from 1 to 127.
- REQ-3: The command shall choose every manifest value three-way (equal sides, or the side that changed from the common version; `origin` one key at a time) and write `files` in sorted key order with the same bytes whatever order the sides list their entries in. · failure: none — a value both sides changed differently is REQ-2
- REQ-4: When both sides changed a skill file, the command shall record the sha256 of git's histogram file merge of that file's three versions, found by their sha256 in the commit history. · failure: none — a version not found or a file merge with a conflict is REQ-2
- REQ-5 (REQ-2): IF a skill file is deleted on one side and changed on the other, is added on both sides with different contents, has its own `merge` attribute, or lies in a folder that disappears on one side while the other side adds or changes a file in it, THEN the command shall leave the manifest as a conflict.
- REQ-6 (REQ-2): IF a side holds a key twice or is not a package manifest, `diff.algorithm` is set to another value than histogram, or `merge.renormalize` is on, THEN the command shall leave the manifest as a conflict.
- REQ-7: The registration command written in the repository's document shall drop the standard output of `manifest-merge-driver` and fall back to git's own text merge on a machine whose `moai` cannot run it. · failure: none — the text merge is the fallback itself
- REQ-8: The `.gitattributes` line and `docs/agents-md-maintenance.md` shall name the driver `moai-manifest-rs` and the `moai` command as its registration. · failure: none — a scan test reads both
- REQ-9: The repository shall hold no Python merge driver and no instruction to register one, outside the history records under `.moai/specs/` and the decisions log. · failure: none — a scan test reads the tracked files
- REQ-10 (REQ-8): IF a clone registers only the former driver name `moai-manifest`, THEN a merge in a checkout whose `.gitattributes` names `moai-manifest-rs` shall run git's own text merge of the manifest and start no Python.

## Acceptance

- AC-1 (REQ-1, REQ-4): in temporary repositories with the built binary registered, `git merge`, a merge started in a subfolder and `git merge-tree --write-tree` end clean with every hash equal to the merged file: one file edited apart on both sides, different files, new files on both sides, a rename on one side, a folder renamed on one side alone, and a triple on which the histogram and myers merges differ · verify: `cargo test --locked --offline --test manifest_merge_driver_cli merged_`
- AC-2 (REQ-2, REQ-5, REQ-6): each refusal case of REQ-4, REQ-5 and REQ-6 leaves the manifest unmerged with conflict markers (or git's text merge) in it, names the reason on standard error, and an unexpected failure inside the command does the same with a status below 128 · verify: `cargo test --locked --offline --test manifest_merge_driver_cli refused_`
- AC-3 (REQ-7): the registration command read from the document merges cleanly when `moai` is on `PATH`, gives git's text conflict when `moai` is not on `PATH` or does not know the command, and adds nothing to the output of `git merge-tree` · verify: `cargo test --locked --offline --test manifest_merge_driver_cli registered_`
- AC-4 (REQ-3): the merged bytes do not change with the order of the entries of either side, over seeded cases, and equal the bytes `moai check fix` writes for the merged tree · verify: `cargo test --locked --offline --test manifest_merge_driver_cli order_`
- AC-5 (REQ-8, REQ-9): `.gitattributes` and the document name `moai-manifest-rs` and the command, and no tracked file outside the history records names the Python driver or a `python3` registration of it · verify: `cargo test --locked --offline --test manifest_merge_driver_cli text_`
- AC-6 (REQ-10): with only the former key registered and a stand-in `python3` first on `PATH` that leaves a marker, a merge with the new attribute line leaves no marker and gives git's text merge, while the same setup with the former attribute line leaves the marker · verify: `cargo test --locked --offline --test manifest_merge_driver_cli registered_a_clone_with_only_the_old_driver_line_runs_no_python`

## Scope

- write: `.moai/specs/SPEC-20261126-manifest-merge-driver-rust/`, `src/manifest_merge.rs`(새), `src/ops/manifest_merge_driver.rs`(새), `docs/design-review/appendix/operations/manifest_merge_driver.md`(새), `tests/manifest_merge_driver_cli.rs`(새), `.gitattributes`, `docs/agents-md-maintenance.md`
- write (D-4): `tools/merge_driver_manifest.py`(삭제), `tests/source_manifest_test.py`·`tests/prop_manifest_test.py`(드라이버 시험만 삭제), `tools/source_manifest.py`(드라이버를 말하는 글만)
- write (D-2): `src/scope.rs`(주석의 드라이버 이름만), `tests/source_manifest_facts.rs`(속성 시험의 드라이버 이름과 머리 설명), `tests/card/card_generated_scope.rs`(속성 이름을 읽는 한 줄; 지금 다른 작업이 `tests/card/**` 를 예약해 두어, 허락을 받은 뒤에만 써요)
- write (한 줄 자리 예약으로만): `src/lib.rs`, `src/help.rs`, `tests/areas.toml`, `tools/repo_scans.json`, `tests/fixtures/help/all*.txt`
- read-only: `src/check_fix.rs`(지금 다른 작업이 예약해 두었어요); `src/bundle.rs`(설정에 등록된 이름은 모두 글 merge 로 덮어써서 새 이름도 이미 덮여요), `tests/bundle_*_cli.rs`(저마다 자기 `.gitattributes` 와 설정을 써요)
- forbidden: 공유 체크아웃과 사용자의 git 설정 쓰기, Python 을 새로 쓰거나 실행하기

## Plan

1. SPEC 과 `research.md`, 승인 요청(merge 동작이라 tier 2).
2. 시험 먼저: `tests/manifest_merge_driver_cli.rs`(빨강: 명령이 없음). Python 시험의 드라이버 사례를 하나씩 옮겨요.
3. `src/manifest_merge.rs` 와 `src/ops/manifest_merge_driver.rs`, 등록, 문서와 `.gitattributes`, Python 드라이버와 그 시험 삭제.
4. 영역 실행, 임시 저장소 위험 경로 실행, 가드 변이, READY.

## Risks

- 틀린 해시를 깨끗한 merge 로 내면 착지 뒤에야 드러나요. 판단할 수 없는 모든 경우는 충돌로 남기고, 저장소의 매니페스트 시험(`source_manifest_facts`)이 안전망으로 남아요.
- 설치된 `moai` 가 이 명령을 모르는 동안(착지·설치 전) 등록 명령은 git 의 글 merge 로 떨어져요: 지금의 "등록 안 함"과 같은 동작이에요.
- 옛 설정 줄과 `.git` 안의 복사본은 남아요. 옛 `.gitattributes` 줄을 가진 체크아웃(이 변경이 착지하기 전에 만든 브랜치)에서 하는 merge 는 여전히 그 줄을 쓸 수 있어요. 어느 쪽 `.gitattributes` 가 드라이버를 고르는지는 시험으로 재서 Result 에 적어요.
- 새 이름을 등록하기 전에는 매니페스트가 git 의 글 merge 로 합쳐져서, 양쪽이 바꾼 merge 는 충돌로 남아요(`moai check fix` 로 다시 그려요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 해시는 작업 폴더의 파일이 아니라 그 파일의 세 판을 합친 결과에서 계산해요("합친 트리의 파일에서 다시 계산"이라는 인수 기준을 이렇게 읽어요) | git 은 합친 파일을 쓰기 전에 드라이버를 불러요(이전 드라이버의 git 2.54 측정, `research.md`), `merge-tree` 에는 작업 폴더가 없어요. 파일에서 읽으면 우리 쪽 해시가 들어가요 | — |
| D-2 | 명령 이름은 `manifest-merge-driver`, 드라이버 이름은 새 이름 `moai-manifest-rs` (revision 2) | `src/ops/` 한 파일 한 명령 규칙. 이름이 그대로면 사람이 설정을 바꾸기 전까지 merge 마다 python3 가 돌아요. 작업 세션의 `git -c merge.…` 덮어쓰기는 경계 훅이 거절해요(장부 노트 9787) | 속성 한 줄과 시험의 이름 |
| D-3 | 복사·`--install`·`--check` 단계를 없애요 | 설치된 바이너리는 브랜치 안에 없어서, 브랜치가 자기 드라이버를 돌릴 수 없어요(옛 복사 단계가 있던 이유) | 설치 명령 추가 |
| D-4 | Python 드라이버와 그 시험을 이 SPEC 에서 지우고 같은 사례의 Rust 시험으로 바꿔요. 다른 Python 도구는 그대로 둬요 | "python3 를 부르는 드라이버 안내가 저장소에 남지 않아요"라는 인수 기준은 그 파일이 남으면 채울 수 없어요(파일 머리글이 안내예요). 대안 (b): 파일은 남기고 스캔을 좁혀 Rust 전환 작업에 넘겨요 | 파일 복원 |
| D-5 | 그리는 함수는 드라이버 모듈에 두고, 바이트가 `check fix` 의 것과 같음을 시험이 고정해요 | `src/check_fix.rs` 를 지금 다른 작업이 예약해 두었어요 | 함수 합치기 |

## Result

(아직 구현 전이에요.)
