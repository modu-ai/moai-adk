---
id: SPEC-20261126-project-next-flow-setup
title: "Picking NeXT-Flow in the project setup makes the flow run: the integration branch, the card ledger and the protection hook are set up in one fixed order"
status: draft
revision: 1
created: 2026-11-26
---

# Picking NeXT-Flow in the project setup makes the flow run: the integration branch, the card ledger and the protection hook are set up in one fixed order

## Intent

`/moai project` 에서 NeXT-Flow 를 골라도 지금은 프로필에 `work_flow {flow, integration_branch}` 와 선택적인 워크트리 기준만 써요. 흐름 문서가 요구하는 것 중 글과 과정에 없는 것이 있어요: 통합 브랜치 만들기, 보호 훅 설치, 착지 뒤 원격으로 올릴지 묻기, 독립 감사 확인을 기본으로 켤지 묻기, 원장 등록을 흐름 설정의 일부로 하기, 리더와 레인을 어떻게 시작하는지 최종 보고에 한 줄 적기. 사용자 지시(2026-11-25T00:43:00Z, 카드 본문, 근거 `~/.moai/reports/moai-adk/2026-11-25-moai-project-survey.md` B1): `/moai project` 전체 과정을 전수 조사해 모두 고친다. 이 SPEC 은 NeXT-Flow 를 고른 사용자에게만 새 설정 단계(9단계)와 후속 질문 둘을 더해요. 필요한 명령은 모두 이미 있어요: `work-flow-plan`(`missing-target`, `occupied-worktree`, `ok`), `git-hooks-plan` 과 `git-hooks-apply`(`SPEC-20261125-integration-guard-install` 의 설치 명령), `ledger-register-plan` 과 `ledger-register-apply`, `project-plan` 의 `remote_copy` 와 `work_flow.land_verdict` 입력. 그래서 이 일은 글(`skills/moai/workflows/project.md`, `skills/moai/references/internal/interview.md`)과 시험이에요. 승인된 `SPEC-20261124-moai-project-interview` 의 REQ-1(질문 은행은 목적과 언어 넷만 묻는다)에는 예외가 하나 생겨요: 후속 질문은 NeXT-Flow 를 고른 뒤에만 한 번의 후속 호출로 물어요(REQ-5, REQ-11). 훅이 설치되면 통합 브랜치를 손으로 만들거나 옮기는 것은 모두 거절되므로 순서는 통합 브랜치, 원장, 훅이에요.

## Interview

- Q1: 이 변경의 tier 와 SPEC 이 필요한가요? → A: 경로로는 tier 1(글과 시험)이지만 훅 설치, 실제 HOME 의 원장 쓰기, 원격 쓰기 스위치를 부르는 글이라 tier 2 로 보고 작은 SPEC 을 써요. 승인은 리더가 위임으로 기록해요. 근거: 카드 노트 9683, 9684(D-1).
- Q2: 시작 순서는? → A: `skills/moai/workflows/project.md` 와 시험 네 개를 다른 진행 중 작업이 live 예약으로 쥐고 있어요. 이 SPEC 을 먼저 쓰고, 예약이 풀리면 `card scope` 로 범위를 넓히고 새 NeXT tip 을 한 번 합친 뒤 구현해요. m-flow 마무리 커밋 문제는 접어 넣지 않고 순서만 이 일 다음이에요. 근거: 노트 9684(D-2).
- Q3: 원장 등록은 어디에서? → A: NeXT-Flow 설정 단계에서 등록하고, 끝의 점검 항목은 NeXT-Flow 가 아닌 프로젝트용으로 남겨요(D-3).
- Q4: 착지 확인(독립 감사 확인)의 기본값과 말은? → A: 켜는 쪽을 첫 선택지로 두되 "PASS 없이 착지 안 됨"이라고 쓰지 않아요. 닫는 착지는 모든 모드에서 이미 독립 PASS 가 필요하고, 켜면 더해지는 것은 부분 착지(`--keep-open`), 읽을 수 없는 검사, 빠진 스캔 영수증의 거절이에요(D-4).
- Q5: 원격 사본의 기본값은? → A: "아니요, 로컬에 둬요"를 첫 선택지로 두고 아무것도 미리 고르지 않아요. 원격 쓰기는 사람이 고를 때만이에요(D-5).
- Q6: 현재 브랜치가 기본 브랜치인 프로젝트는? → A: 사용자의 체크아웃이 잠기지 않도록 새 전용 통합 브랜치를 첫 선택지로 권해요. 기본 브랜치를 통합 브랜치로 쓰는 선택지는 남겨요(D-6).

## Non-goals

- 새 명령 만들기(통합 브랜치 만들기 명령은 없어요: 글이 사용자가 고른 뒤 `git branch <이름> <끝>` 을 써요).
- `src/` 의 변경: 필요한 명령과 프로필 입력이 모두 이미 있어요.
- NeXT-Flow 가 아닌 흐름의 설정과 끝 점검 항목의 변경.
- m-flow 마무리 커밋 문제와 훅의 동작 변경(`SPEC-20261125-integration-guard-install` 의 몫).
- 어떤 것도 원격으로 보내는 일: 설정 단계는 `git push` 를 하지 않아요.

## Requirements

- REQ-1: When the user picks NeXT-Flow in the project setup, the project flow shall run a NeXT-Flow setup step after the profile is applied, in the order integration branch, card ledger, protection hook. · failure: none — a project that picks another flow runs no such step.
- REQ-2: When the integration branch the profile names does not exist, the setup step shall offer to create it from the current tip with no creation before the user picks it. · failure: none — without the pick nothing is created.
- REQ-3: IF the integration branch is held by another worktree, THEN the setup step shall create nothing while relaying the answer of `work-flow-plan` as printed.
- REQ-4: When the integration branch exists and the user picks the hook, the setup step shall install the protection hook through `git-hooks-plan` and `git-hooks-apply` followed by reading the guard record back with the way back. · failure: none — without the pick no hook is installed.
- REQ-5: When the user picked NeXT-Flow, the project flow shall ask in one follow-up call whether landed work is copied to the remote and whether the landing check is turned on, in plain words with no internal key name. · failure: none — another flow asks neither question.
- REQ-6: The remote copy question shall list "No, keep it local" first with no option preselected.
- REQ-7: The landing check question shall list turning it on first without saying that only that mode needs an audit PASS to land.
- REQ-8: While the current branch is the default branch, the integration branch follow-up of NeXT-Flow shall list a new dedicated branch first.
- REQ-9: When the user answers both follow-up questions, the project flow shall record the answers in the profile through `project-plan` and `project-apply`. · failure: none — an unanswered question records nothing.
- REQ-10: The final report of the project flow shall end with one line that says how to start the leader and the lanes.
- REQ-11: The question bank shall keep the purpose as the only row of its asked table while the follow-up questions stand outside that table.
- REQ-12: The setup step shall not run `git push`.

## Acceptance

- AC-1 (REQ-1, REQ-4): 임시 HOME 과 한 커밋짜리 임시 저장소에서 설정 순서(프로필 적용, 통합 브랜치, 원장 등록, 훅 설치) 뒤에 통합 브랜치 기록, 실행 가능한 훅, 원장 파일이 있고 `update-ref`, `branch -f`, `branch -D` 가 거절돼요; 훅이 먼저면 브랜치 만들기가 거절돼요(순서의 이유) · verify: `cargo test --locked --offline --test project_next_flow_setup setup_makes_record_hook_ledger_and_refuses_hand_moves`
  verify: `cargo test --locked --offline --test project_next_flow_setup hook_first_then_branch_create_is_refused`
- AC-2 (REQ-5, REQ-6, REQ-7, REQ-9, REQ-11): 두 후속 질문의 질문과 선택지 칸에 내부 키 이름과 금지어가 없고, 추천 선택지가 첫째이고, 아무것도 미리 고르지 않고, 질문 은행의 물음 표는 `purpose` 하나로 그대로이고, 두 답이 프로필에 읽혀요 · verify: `cargo test --locked --offline --test project_next_flow_setup follow_up_rows_use_plain_words_and_answers_reach_the_profile`
- AC-3 (REQ-2, REQ-3): 통합 브랜치가 없거나 다른 워크트리가 쥔 임시 저장소에서 글이 `work-flow-plan` 의 `kind` 를 모두 다루고, 안내는 `state.message` 를 찍힌 대로 사용자의 말로 옮기고, 만들기는 `missing-target` 에서 고른 뒤에만 권하고 `occupied-worktree` 는 아무것도 만들지 않아요 · verify: `cargo test --locked --offline --test project_next_flow_setup setup_guidance_is_work_flow_plans_answer`
- AC-4 (REQ-1, REQ-8, REQ-10, REQ-12): 9단계가 8단계 뒤 최종 보고 앞에 있고, 훅이 프로필 적용과 통합 브랜치 뒤에 오고, 훅과 원격 사본과 원장 선택에 미리 고른 것이 없고, 보고 끝 줄이 `moai cc -l` 과 `moai cc -a` 를 말하고 `/moai -l ` 이나 `/moai -a ` 는 말하지 않고, 8단계부터 `## Procedure` 끝까지 `git push` 가 없어요; 기본 브랜치에서는 새 전용 통합 브랜치가 첫 권유예요 · verify: `cargo test --locked --offline --test project_next_flow_setup step_9_order_and_closing_line`
- AC-5 (REQ-1, REQ-5, REQ-12): 글이 바뀌어도 고정돼 있던 기존 시험이 같은 뜻으로 통과해요 · verify: `cargo test --locked --offline --test project_flow_closing`
  verify: `cargo test --locked --offline --test project_interview_minimal`
  verify: `cargo test --locked --offline --test project_preview_order`
  verify: `cargo test --locked --offline --test update_route`
  verify: `cargo test --locked --offline --test generic_flow_questions`
  verify: `cargo test --locked --offline --test project_applied_table`
  verify: `cargo test --locked --offline --test agent_defs_package`

## Scope

- write: `.moai/specs/SPEC-20261126-project-next-flow-setup/spec.md`
- write: `skills/moai/workflows/project.md`, `skills/moai/references/internal/interview.md`
- write: `tests/project_next_flow_setup.rs`
- write: `tests/project_flow_closing.rs`, `tests/project_interview_minimal.rs`, `tests/project_preview_order.rs`, `tests/update_route.rs`, `tests/generic_flow_questions.rs`, `tests/project_applied_table.rs`, `tests/agent_defs_package.rs`
- write: `skills/moai/references/source-manifest.json`, `tools/repo_scans.json`

위 두 번째 줄의 시험 일곱은 글이 바뀌어서 같은 뜻으로 고치는 것이고, 예약이 풀린 뒤 범위를 넓혀요. 마지막 줄의 두 파일은 공유 파일이라 한 줄씩 임대해서 쓰고 커밋 뒤 바로 반납해요(매니페스트는 `moai check fix` 가 새 sha256 을 쓰고, 목록 파일에는 새 시험 한 줄을 더해요).
- read-only: `src/work_flow.rs`, `src/ops/work_flow_plan.rs`, `src/card/guard.rs`, `src/card/remote_copy.rs`, `src/card/land/check.rs`, `src/card/land.rs`, `src/git_hooks/**`, `src/ops/git_hooks_*.rs`, `src/ops/ledger_register_*.rs`, `src/profile.rs`, `install/git-hooks/reference-transaction`
- forbidden: `src/` 의 변경, 실제 HOME 과 그 안의 Claude 설정 폴더와 MoAI 상태 폴더(시험은 임시 HOME 과 임시 저장소만), 실제 GitHub 으로 보내는 모든 것, 원장 스키마 변경, 파이썬

## Design

글은 `skills/moai/workflows/project.md` 의 단계 구조를 따라요. 통합 브랜치 이름은 지금처럼 3단계의 후속 선택에서 물어요. NeXT-Flow 이고 현재 브랜치가 기본 브랜치면 그 선택의 첫 항목이 새 전용 통합 브랜치이고, 지금의 첫 항목들(프로필 값, 원격 기본 브랜치, 현재 브랜치)은 그 뒤에 남아요. 3단계(질문)에는 NeXT-Flow 를 고른 사용자용 후속 질문 한 줄(한 번의 후속 호출, 세 질문 이하)을 더하고, 4단계(프로필 계획)의 필드 목록에 원격 사본 답을 더하고, 8단계 뒤에 9단계 "NeXT-Flow 설정"을 더해요.

9단계(기본 작업 폴더에서, git 저장소 밖이면 건너뛰어요) 순서:

1. `work-flow-plan` 으로 통합 브랜치 상태를 읽어요. `ok` 면 다음으로. `missing-target` 이면 현재 끝에서 `git branch <이름> <끝>` 을 만들지 사용자에게 묻고 고른 뒤에만 만들어요. `occupied-worktree` 이면 아무것도 만들지 않고 `state.message` 를 찍힌 대로 사용자의 말로 옮겨요.
2. 원장 등록을 미리 보여 주고 사용자가 고르면 `ledger-register-apply` 를 해요(끝 점검 항목은 NeXT-Flow 가 아닌 프로젝트용 대안으로 그대로 남아요).
3. `git-hooks-plan` 으로 훅 설치를 미리 보여 주고 사용자가 고르면 `git-hooks-apply` 를 하고, 보호 기록(`guard.record`)을 읽어 알리고 되돌리는 길 `moai git-hooks-revert <stamp>` 를 말해요. 훅이 설치되면 통합 브랜치의 손 이동은 모두 거절되고 착지 명령만 그 브랜치를 옮겨요.

후속 질문 둘은 일반적인 말이에요. 원격 사본: "일이 끝나 합쳐지면 원격에도 복사해 둘까요?"(첫 선택지 "아니요, 로컬에 둬요", 미리 고르지 않음). 착지 확인: 켜는 쪽이 첫 선택지이고, 켜면 더 엄격히 거절하는 것(부분 착지, 읽을 수 없는 검사, 빠진 스캔 영수증)을 말로 풀어요. 답은 `project-plan` 의 `remote_copy` 와 `work_flow.land_verdict` 입력으로 프로필에 가요. 질문 칸에는 내부 키 이름이나 `enforce`, `observe` 를 쓰지 않아요.

최종 보고에는 통합 브랜치, 훅, 원격 사본, 착지 확인, 원장의 행을 더하고 마지막 줄에 리더와 레인을 시작하는 방법(`moai cc -l`, `moai cc -a`)을 적어요.

## Plan

1. 이 SPEC 의 승인을 기다려요(tier 2). 승인과 `card scope` 확장(예약이 풀린 뒤) 뒤에 새 NeXT 끝을 한 번 합쳐요.
2. 시험 먼저(RED): `tests/project_next_flow_setup.rs` 에 AC-1 에서 AC-4 시험 다섯을 써요. 임시 HOME 과 임시 저장소를 쓰고 `tests/integration_guard_install.rs` 의 조각(`:62-117`, `:543-571`)과 `tests/work_flow_plan_cli.rs`(`:243-282`)를 모델로 삼아요. 빨갛게 도는 것(컴파일 오류 아님)을 로그로 남겨요.
3. 글(GREEN): `interview.md` 에 두 후속 줄, `project.md` 에 3단계 한 줄, 4단계 필드, 9단계, 최종 보고 줄.
4. 글이 바뀌어서 깨지는 기존 시험을 같은 뜻으로 고쳐요(AC-5 목록). `cargo fmt --all`, `moai check fix` 로 매니페스트, `card line-lease` 로 `repo_scans.json` 한 줄.
5. SPEC Result, `moai check affected --card` 한 번. READY 에 임시 프로젝트에서의 위험 경로 실행(AC-1)과 `guard-mutant:` 줄(순서 문장 빼기, 훅 선택 빼기 등 새 가드 시험마다 한 줄)을 담아요.

## Risks

- tier 2, SPEC 승인은 사용자 위임으로 리더가 기록해요. 감사는 작성하지 않은 세션의 전체 감사예요.
- `skills/moai/workflows/project.md` 와 시험 넷은 다른 진행 중 작업이 쥐고 있고 그 가지가 같은 구간(2단계 삭제, 8단계, 끝 점검)을 고쳐요: 병합 충돌이 날 수 있어서 예약이 풀린 뒤 새 끝에서 한 번 합쳐요.
- 훅이 설치되면 통합 브랜치의 손 이동이 모두 거절돼서 사용자 자신의 커밋과 마무리 커밋도 막혀요. 9단계가 8단계 뒤에 있고, 글이 "이제부터 그 브랜치는 착지 명령만 옮긴다"고 분명히 말해야 해요. 같은 문제를 다루는 m-flow 마무리 커밋 일은 이 일 다음이에요.
- 시험은 글을 읽어서 모델이 실행 시점에 실제로 쓰는 말은 확인하지 못해요(글의 고정 문장으로만 못박아요).
- 안내가 `work-flow-plan` 의 영어 메시지를 사용자의 말로 옮기므로 바이트 같음은 시험할 수 없고, `kind` 들과 "찍힌 대로" 문장만 못박아요.
- `tools/repo_scans.json` 은 감사 바닥 경로이고 성장만 허용인지 읽지 못했어요(unverified): 한 줄 추가가 그 규칙 안에 드는지 구현할 때 확인해요.
- 연결된 워크트리에서는 보호 기록이 `not_written` 이에요(읽은 코드 `src/card/guard.rs:237-240`): 9단계는 기본 작업 폴더에서 돌려요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | tier 2 로 보고 작은 SPEC 을 써요 | 훅 설치, 실제 HOME 의 원장, 원격 쓰기 스위치를 부르는 글; 노트 9684 | tier 1 로 두고 예외를 카드 노트에만 적기 |
| D-2 | 지금 take 하고 SPEC 폴더만 쓰고, 예약이 풀린 뒤 범위를 넓혀 새 끝을 한 번 합쳐요 | 노트 9684; 충돌 줄이기 | 앞선 작업의 착지를 기다렸다 시작하기 |
| D-3 | 원장은 NeXT-Flow 설정 단계에서 등록하고 끝 점검은 그 밖의 프로젝트용으로 남겨요 | 흐름 설정의 일부; 고정된 시험 문장 유지 | 끝 점검으로 완전히 옮기고 시험 문장을 다시 쓰기 |
| D-4 | 착지 확인은 켜는 쪽이 첫 선택지이고 "PASS 없이 착지 안 됨" 문구는 쓰지 않아요 | 닫는 착지는 모든 모드에서 이미 PASS 가 필요(`src/card/land.rs:515-523`); 거짓 약속 금지 | 꺼진 쪽을 첫 선택지로 |
| D-5 | 원격 사본은 "아니요, 로컬에 둬요"가 첫 선택지이고 미리 고르지 않아요 | 원격 쓰기는 사람이 고를 때만 | 켜는 쪽을 첫 선택지로(사용자 결정) |
| D-6 | 현재 브랜치가 기본 브랜치면 새 전용 통합 브랜치를 첫 선택지로 권해요 | 사용자의 체크아웃이 잠기지 않아요 | 기본 브랜치를 통합 브랜치로 쓰는 것을 첫 선택지로 |
