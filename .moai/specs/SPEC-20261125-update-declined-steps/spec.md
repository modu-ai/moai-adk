---
id: SPEC-20261125-update-declined-steps
title: "A choice step the user declined stays declined while its source is the same; the merge steps say where their text comes from and what blocked them"
status: draft
revision: 1
created: 2026-11-25
---

# A choice step the user declined stays declined while its source is the same; the merge steps say where their text comes from and what blocked them

## Intent

`/moai update` 의 합치기 단계(update 경로 ⑤)는 사용자가 "아무것도 바꾸지 않기"를 골라도 update 때마다 같은 물음을 다시 내고, 그 물음이 어디서 왔는지도 말하지 않아요. 실제로 9일 전 백업(`.moai/backups/<UTC>/root-files/AGENTS.local.md`)을 `CLAUDE.local.md` 에 합치자는 `agents-local-merge` 가 거절 뒤에도 `needs_decision` 으로 남아서 `all_unchanged`(`src/project_update.rs` 의 `plan`)가 영영 `false` 예요. 같은 화면에서 `agents-picks` 는 "AGENTS.md 가 없다"며 `skip` 인데 합치기는 AGENTS.local.md 를 합치자고 해서 둘이 어긋나 보이고, `claude-settings-merge` 는 `conflict` 라는 말만 하고 어느 키가 왜 막혔는지도, 다음 길도 말하지 않아요. 이 SPEC 은 (1) 선택 단계를 거절하는 명령을 두어 같은 출처인 동안 그 단계를 `skip` 으로 읽게 하고, (2) 출처가 바뀌면 다시 묻게 하고, (3) 거절을 되돌리는 길을 주고, (4) 합치기 단계의 이유 문장이 출처(백업이면 그 경로)를, 설정 합치기의 충돌이 바뀐 키 이름과 고를 수 있는 길(내 값 유지, 건너뛰기)을 말하게 해요.

## Interview

- Q1: 거절을 어디에 기억할까요? → A: 아직 묻지 않았어요. 기본: 프로젝트 상태 폴더의 `.moai/manifests/update-declined.json` 한 파일(리더 지시 2026-11-25). HOME 은 건드리지 않아요(D-1).
- Q2: 거절은 무엇에 묶을까요? → A: 아직 묻지 않았어요. 기본: 단계 이름과 그 단계가 제안하는 글의 출처 sha256. 출처가 바뀌면 거절은 읽히지 않고 단계는 다시 물어요(D-2).
- Q3: 거절을 되돌릴 수 있어야 할까요? → A: 아직 묻지 않았어요. 기본: 예. 같은 명령의 `--ask-again` 이 그 단계의 기록을 지워요(D-3).
- Q4: 설정 충돌의 "내 값 유지"는 무엇을 하나요? → A: 아직 묻지 않았어요. 기본: 충돌한 키는 파일의 현재 값을 그대로 두고 나머지 키만 합쳐요. 그 키의 충돌은 다음 미리보기에도 남고 `skip` 으로 끝내요. 명령은 그 계획이 충돌 없이 만들어질 때만 보여요(D-4).

## Non-goals

- 거절 기록이 없는 단계, 그리고 `generation`·`profile`·`claude-settings`·`hooks`·`github`·`retired-hosts`·`go-*`·`reports` 단계의 동작: 그대로예요. `generation` 이 `needs_decision` 인 까닭(요청 기록 없음)은 update 경로 ①②의 몫이에요.
- 물음 횟수, `next`/`say`/`ask`/`chain` 필드, `update.md` 의 문장(update 경로 ⑥·⑦).
- 백업 위치와 HOME 표식(update 경로 ⑧), `retired-hosts` 의 삭제(다른 조각).
- 사용자가 고친 파일을 덮어쓰거나 합치는 일. 거절은 파일을 쓰지도 지우지도 않아요.

## Requirements

- REQ-1: When `project-update-decline` is run for one of the steps `agents-picks`, `agents-local-merge` and `claude-settings-merge` with the source sha256 the update preview printed for it, the command shall record the step name and that sha256 in `.moai/manifests/update-declined.json`.
- REQ-2 (REQ-1): IF the given sha256 is not the source sha256 the step has now, or the step name is not one of those three, THEN the command shall write nothing and name the reason.
- REQ-3 (REQ-1): While a record of a step holds the source sha256 the step has now and the step would read `needs_decision` or `conflict`, the update preview shall read that step as `skip` with a reason that says the user declined it and names the command that asks again.
- REQ-4 (REQ-3): When every step reads `unchanged`, `skip` or `optional`, the update preview shall report `all_unchanged` as true, a declined step counted among the `skip` ones. · failure: none — the flag is computed from the steps already read.
- REQ-5 (REQ-1): IF a record's sha256 differs from the source sha256 the step has now, THEN the update preview shall read the step as it would without the record.
- REQ-6 (REQ-1): When `project-update-decline` is run with `--ask-again` for a step, the command shall remove that step's record and leave the other records as they are. · failure: none — a step with no record is left as it is.
- REQ-7: When a step of the update preview can be declined, the preview shall carry for it the source sha256, the effect of declining in one sentence and the argv of the decline command. · failure: none — a step with no source prints no decline object.
- REQ-8: Where the source of `agents-local-merge` is a backup copy, the step's reason shall name the backup path and say that the file is a copy moved there earlier.
- REQ-9: While the `claude-settings-merge` plan has conflicting keys, the step's reason shall name those keys and the file.
- REQ-10 (REQ-9): While the `claude-settings-merge` plan has conflicting keys, the step's preview shall list the choices `keep-mine` and `skip`.
- REQ-11 (REQ-10): The choice `keep-mine` shall carry the argv of the settings apply that leaves each conflicting key at the file's current value and merges the other earlier settings.
- REQ-12 (REQ-10): IF the plan made with `keep-mine` still has a conflict, THEN the step shall leave the choice `keep-mine` out of its preview.
- REQ-13 (REQ-10): The choice `skip` shall carry the argv of the decline command of REQ-1 for the step.
- REQ-14: The decline command shall write no file other than `.moai/manifests/update-declined.json`.
- REQ-15 (REQ-14): IF the state folder or the manifest path is a symbolic link, THEN the decline command shall write nothing and name the link.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-14, REQ-15): 임시 프로젝트에서 decline 명령이 올바른 sha 로 기록을 만들고, 틀린 sha·모르는 단계 이름·링크로 된 manifest 경로는 아무것도 쓰지 않고 이유를 말해요 · verify: `cargo test --locked --offline --test project_update_plan_cli declined_record`
- AC-2 (REQ-3, REQ-4, REQ-7): 합치기를 거절하면 다음 `project-update-plan` 에서 그 단계가 `skip` 이고 이유가 거절을 말하며 `all_unchanged` 가 true 예요 · verify: `cargo test --locked --offline --test project_update_plan_cli declined_skip`
- AC-3 (REQ-5): 거절한 뒤 출처 파일(백업 사본)의 바이트가 바뀌면 단계는 다시 `needs_decision` 이에요 · verify: `cargo test --locked --offline --test project_update_plan_cli declined_source_changes`
- AC-4 (REQ-6): `--ask-again` 뒤 단계가 다시 물어요 · verify: `cargo test --locked --offline --test project_update_plan_cli declined_ask_again`
- AC-5 (REQ-8): 백업이 출처인 합치기 단계의 이유가 그 백업 경로를 담아요 · verify: `cargo test --locked --offline --test project_update_plan_cli declined_reason_names_backup`
- AC-6 (REQ-9, REQ-10, REQ-11, REQ-12, REQ-13): 설정 합치기의 충돌이 바뀐 키 이름(예: `statusLine`)과 `keep-mine`·`skip` 을 말하고, `keep-mine` 의 명령을 실제로 실행하면 그 키의 값은 그대로이고 나머지가 합쳐져요 · verify: `cargo test --locked --offline --test project_update_plan_cli declined_settings_conflict`
- AC-7: 기존 update 계획 시험이 그대로 통과해요 · verify: `cargo test --locked --offline --test project_update_plan_cli`

## Scope

- write: `.moai/specs/SPEC-20261125-update-declined-steps/spec.md`
- write: `src/project_update.rs`, `src/project_update/declined.rs`, `src/agents_local_merge.rs`
- write: `src/ops/project_update_decline.rs`, `tests/project_update_plan_cli.rs`
- shared files by line lease, only as far as a new operation needs them: `src/help.rs`, `tests/areas.toml`, `tools/repo_scans.json`
- read-only: `src/claude_settings_merge.rs`, `src/claude.rs`, `src/apply.rs`, `src/state_root.rs`
- forbidden: 실제 HOME, 설치된 moai, 다른 프로젝트, 사용자 파일을 쓰거나 지우는 코드, `src/update*` 경로

## Plan

1. 시험 먼저: `tests/project_update_plan_cli.rs` 에 AC-1~AC-6 의 시험을 임시 프로젝트로 쓰고 지금 코드에서 실패하는 이유를 기록해요.
2. `src/project_update/declined.rs`(`src/project_update.rs` 가 `pub mod declined;` 로 선언해서 공유 파일 `src/lib.rs` 를 건드리지 않아요): 기록 읽기·쓰기(`.moai/manifests/update-declined.json`, `apply::publish_replace`, 링크 거부), 단계별 출처 sha256, 단계를 `skip` 으로 읽는 함수, 거절·다시 묻기 함수를 만들어요.
3. `src/ops/project_update_decline.rs`: 명령 한 개. 옵션은 `--step`, `--source-sha256`, `--ask-again`, `--permission-reference`, 프로젝트 경로예요.
4. `src/project_update.rs`: `merge_steps` 의 세 단계 결과에 거절을 반영하고 단계의 preview 에 `decline` 객체를 넣어요. 설정 합치기의 이유 문장과 `choices` 를 바꿔요.
5. `src/agents_local_merge.rs`: 출처가 백업이면 이유 문장에 경로와 "moved copy" 를 넣어요.
6. 새 명령의 도움말·등록 한 줄은 공유 파일이라 `card line-lease` 로만 써요.
7. `moai check affected` 의 카드 영수증으로 마무리하고 Result 에 기준마다 한 줄을 적어요.

## Risks

- 거절 기록은 프로젝트 안의 한 파일이라 다른 체크아웃에는 따라가지 않아요(`.moai/manifests/` 는 팀이 git 에 올리는 목록에 없어요). 다른 체크아웃에서는 한 번 더 물어요.
- 거절한 출처의 바이트가 한 번 바뀌면 거절이 풀려요. 백업 사본은 보통 바뀌지 않으므로 의도한 동작이에요.
- `keep-mine` 은 나머지 키를 합치지만 설정 적용기는 값이 이미 같은 키의 기준값을 새로 쓰지 않아서(`src/claude.rs`, 이 SPEC 의 범위 밖), 다음 미리보기에도 그 키의 충돌이 남고 끝내는 길은 `skip` 이에요. 선택지의 `effect` 가 그렇게 말해요.
- `agents-picks` 가 `skip` 이고 합치기가 `needs_decision` 인 어긋남은 이유 문장에 출처를 말하는 것으로만 풀어요. 단계를 합치는 일은 이 SPEC 이 아니에요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 거절은 `.moai/manifests/update-declined.json` 에 기억해요. 대안: HOME 에 기억하기(다른 프로젝트와 섞여요). | 리더 지시 2026-11-25. | 파일을 지우거나 `--ask-again` 을 실행해요. |
| D-2 | 거절은 단계 이름과 출처 sha256 에 묶어요. 대안: 단계 이름만(출처가 바뀌어도 영영 안 물어요). | 카드 기준 1·2. | `project_update/declined.rs` 를 지우면 이전 동작이에요. |
| D-3 | 거절을 되돌리는 길은 `--ask-again` 이에요. | 리더 지시 "기억된 거절은 다시 물어 달라고 할 수 있어야 해요". | 해당 없음. |
| D-4 | 설정 충돌의 `keep-mine` 은 충돌한 키를 파일의 현재 값으로 두고 나머지를 합쳐요. | 카드 기준 4 의 "내 값 유지". | 선택지를 지우면 이전 동작이에요. |

## Result

구현 head 8ee4c79ce7(base dfe34f874c). 아래는 이 실행에서 본 값이에요.

- AC-1 met: `cargo test --locked --offline --test project_update_plan_cli declined_record` 가 통과해요(전체 파일 60개 통과, exit 0).
- AC-2 met: `declined_skip` 통과. 거절 뒤 단계는 `skip`, 이유가 거절을 말해요.
- AC-3 met: `declined_source_changes` 통과. 출처 바이트가 바뀌면 다시 물어요.
- AC-4 met: `declined_ask_again` 통과.
- AC-5 met: `declined_reason_names_backup` 통과. 이유가 백업 경로를 담아요.
- AC-6 met: `declined_settings_conflict` 세 시험 통과. `keep-mine` 은 그 키를 두고 나머지를 합치지만, 설정 적용기가 같은 값의 기준을 새로 쓰지 않아서 그 키의 충돌은 다음 미리보기에도 남고 `skip` 으로 끝내요(Risks).
- AC-7 met: `project_update_plan_cli` 전체 60개 통과.
- 영수증: `moai check affected --card dfe34f874c..HEAD`, `result=PASS`, targets=93, diff_check=0 fmt=0 ops_rustfmt=0 clippy=0 test=0.
- AC-1 의 링크 거부(REQ-15)는 기록 파일이 링크일 때와 `manifests` 폴더가 링크일 때 둘 다 시험해요.
- not_run: 전체 `cargo test`, Windows 확인.
