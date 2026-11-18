---
id: SPEC-20261118-project-decide-model
title: "The project interview asks for the decision model and the key without the key reaching the session"
status: draft
revision: 1
created: 2026-11-18
---

# The project interview asks for the decision model and the key without the key reaching the session

## Intent

사용자 지시(2026-11-16): `/moai project`가 "자율 결정 모델을 쓸까요"(안 씀 / Jev / OpenAI)를 선택지로 묻고, 고른 공급자의 키를 "이미 있는 키 / 새 키" 선택으로 받아야 해요. 켜는 일은 결정 요청 글이 외부 API로 나가는 일이라 질문이 그 사실을 말해야 하고, 기본은 안 씀이에요. 키 값은 세션에 들어오면 안 돼요: 사용자가 자기 터미널의 숨김 입력(`moai key-set <provider> --prompt`, 이미 tip에 있어요)에 직접 치고, 세션은 `key-status`로 있음/없음만 읽어요.
질문 은행에 행이 없고, 워크플로에 단계가 없고, "키가 질문 도구나 명령 인자로 가지 않는다"를 고정한 시험이 없어요(리더 대조 note 2072: 남은 것 seq 1, 3, 5, 그리고 seq 4의 scan 시험). 이 SPEC은 그 글과 시험만 맞춰요. `moai decide`의 코드 쪽 설정 읽기는 하지 않아요.

## Non-goals

- `moai decide` 명령과 그 설정(켜고 끄기, 한도): 코드는 다른 카드의 몫이에요(Decisions D-2).
- `key-set --prompt` 구현: 이미 있어요(`src/key_prompt.rs`). 여기서는 문서가 그 명령을 부르는 방식만 정해요.
- Go 세대의 Jev 키 파일(`~/.moai-go-retired/` 아래)을 읽는 일: 세션은 그 파일을 조용히 읽지 않아요.

## Requirements

- REQ-1: The question bank shall carry a `decision_model` row that offers not used, Jev and OpenAI and says that a provider pick sends decision-request text to an external API.
- REQ-2: The `decision_model` row shall list not used first as the recommended and default option.
- REQ-3: When the user picks a provider, the PROJECT workflow shall ask one selection between using the existing key and entering a new key for that provider only..
- REQ-4: The PROJECT workflow shall offer the existing key only when `key-status` reads present for the picked provider.
- REQ-5: When the user enters a new key, the PROJECT workflow shall show the local command `moai key-set <provider> --prompt` for the user's own terminal and then verify the result with `key-status`..
- REQ-6: No step of the PROJECT workflow or the question bank shall pass a key value through the question tool, a command argument or this session's output.
- REQ-7: The PROJECT workflow shall record the pick as the `decision_model` entry of `decisions` with the value `off` as the default that makes no network call.
- REQ-8 (REQ-3): IF the key question is cancelled or left unanswered, THEN the PROJECT workflow shall keep the step open and keep `decision_model` at `off`.
- REQ-9 (REQ-5): IF `key-status` still reads absent after the user's command, THEN the PROJECT workflow shall report the key as absent and keep `decision_model` at `off`.

## Acceptance

- AC-1 (REQ-1, REQ-2): 질문 은행에 `decision_model` 행이 있고 선택지는 안 씀(권장·기본), Jev, OpenAI이며 외부 전송을 말해요.
  verify: `cargo test --locked --offline --test project_decide_model decision_model_row_offers_off_jev_and_openai_with_the_transmission_note`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-8, REQ-9): 워크플로가 고른 공급자에 대해서만 "이미 있는 키 / 새 키"를 묻고, 이미 있는 키는 `key-status`가 있음일 때만 내놓고, 새 키는 사용자 터미널의 `moai key-set <provider> --prompt`와 `key-status` 확인으로 안내해요.
  verify: `cargo test --locked --offline --test project_decide_model project_workflow_asks_existing_or_new_key_for_the_picked_provider_only`
- AC-3 (REQ-6): 스킬 글 어디에도 키 값이 질문 도구·인자·세션 출력으로 가는 단계가 없어요(`key-set`은 `--prompt`로만 나오고 값 자리표가 없어요).
  verify: `cargo test --locked --offline --test project_decide_model no_skill_text_passes_a_key_through_the_question_tool_or_argv`
- AC-4 (REQ-7): 워크플로가 `decisions`의 `decision_model` 항목과 기본값 `off`(네트워크 호출 0)를 말하고, 그 항목이 실제 `project-plan`에 받아들여져요.
  verify: `cargo test --locked --offline --test project_decide_model decision_model_pick_is_recorded_as_a_decisions_entry_that_defaults_to_off`
- AC-5 (REQ-1…REQ-9): 이웃 검사.
  verify: `cargo test --locked --offline --test agent_defs_package`; `cargo test --locked --offline --test question_bank_hosts`; `cargo test --locked --offline --test update_route`; `python3 tools/source_manifest.py check`; `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261118-project-decide-model/spec.md`, `skills/moai/references/internal/interview.md`, `skills/moai/workflows/project.md`, `skills/moai/references/source-manifest.json`, `tests/project_decide_model.rs`, `tests/agent_defs_package.rs`(질문 은행 id 목록에 한 줄)
- read-only: `skills/moai/references/internal/questions.md`, `src/key_prompt.rs`, `src/key_vault.rs`, `src/decide/provider.rs`, `tests/key_set_prompt_tty_test.py`
- forbidden: 키 값과 실제 HOME의 보관함, 실제 endpoint 호출, 네트워크

## Plan

1. 시험 먼저: `tests/project_decide_model.rs`의 네 시험이 글이 아직 없어서 실패하는 것(RED)을 읽어요.
2. `interview.md`에 `decision_model` 행을, `project.md`에 절차 항목을 더해 GREEN으로 만들고, 매니페스트를 다시 만들어요.
3. 이웃 검사를 돌려요.

## Risks

- 은행 행을 세는 다른 시험이 있을 수 있어요 → 이웃 검사 AC-5가 잡아요.
- 키가 세션으로 새는 문구가 나중에 들어와요 → AC-3의 scan 시험이 스킬 글 전체를 읽어 잡아요.
- `moai decide`가 설정을 다른 곳에서 읽으면 이 기록과 어긋나요 → D-2에 읽는 쪽이 맞춰야 할 자리(`decisions`의 `decision_model`)를 적고 `moai decide` 카드에 알려요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Jev의 키 이름은 `typesafe`, OpenAI는 `openai`로 해요 | `src/decide/provider.rs`의 `Provider::name()`이 그 이름을 쓰고 키 보관함은 공급자 이름으로 항목을 찾아요 | 이름 바꾸기는 후속 SPEC |
| D-2 | 고른 공급자는 프로필 `decisions`의 항목 `{"id": "decision_model", "value": "off"\|"typesafe"\|"openai"}`로 기록하고, `moai decide`가 읽는 쪽(설정 명령)은 만들지 않아요 | 이 카드의 쓰기 범위는 글과 시험이고, `moai decide`의 설정 명령은 다른 카드의 기준이에요; `decisions`는 프로필이 이미 받아들이는 형식이라 새 코드가 필요 없어요 | `moai decide` 쪽이 다른 자리를 정하면 후속 SPEC에서 바꾸기 |
| D-3 | "이미 있는 키"는 `key-status`가 있음일 때만 내놓고 Go 세대의 키 파일은 조용히 읽지 않아요 | 사용자 지시(2026-11-16 19:4xZ): 보관함 가져오기를 거쳐서만 | 가져오기 단계가 생기면 문구에 그 단계를 이름으로 더하기 |
| D-4 | 질문 은행 id 목록을 고정한 `tests/agent_defs_package.rs`에 `decision_model` 한 줄을 더해요 | 새 행이 그 시험을 깨뜨려요(측정: left에 `decision_model` 하나 더, exit 101); 쓰기 범위는 리더가 넓혔어요(2026-11-18, 카드 노트로 승인) | 그 한 줄을 지우면 은행 행도 함께 지워야 해요 |

## Result

- AC-1: met — `cargo test --locked --offline --test project_decide_model` (RED 먼저: 4 failed, exit 101, 글이 아직 없어서). `interview.md`의 `decision_model` 행이 안 씀(권장·기본) → Jev → OpenAI 순서로 있고 질문이 외부 API 전송을 말해요.
- AC-2: met — 같은 시험. `project.md`의 "Decision model" 단계가 고른 공급자에 대해서만 이미 있는 키/새 키를 묻고, 이미 있는 키는 `key-status`가 있음일 때만 내놓고, 새 키는 사용자 자기 터미널의 `moai key-set <provider> --prompt`와 `key-status` 확인이에요. 취소·무응답·여전히 없음이면 `off` 그대로예요.
- AC-3: met — 같은 시험. 질문하는 경로의 스킬 파일 6개(이름으로 읽어요)에서 `moai key-set` 명령은 `--prompt`로만 나오고 값 자리표·`--value`·파이프가 없어요; 단계가 키를 질문 도구·명령 인자·세션 출력으로 보내지 않는다고 말해요.
- AC-4: met — 같은 시험. 선택은 `decisions`의 `decision_model` 항목(`off`/`typesafe`/`openai`)이고 `off`는 기본이며 네트워크 호출이 없다고 말해요; 세 값 모두 실제 `project-plan`이 받아들여요(ok: true).
- AC-5: met — `agent_defs_package` 19 passed(RED 먼저: `interview_question_bank_covers_the_project_axes` 1 failed, exit 101); `question_bank_hosts` 1 passed; `update_route` 8 passed; `product_cli` 79 passed; `project_update_plan_cli` 59 passed; `python3 tools/source_manifest.py check` ok; `cargo fmt --check` exit 0.
- 남은 것: `moai decide`가 이 기록(`decisions`의 `decision_model`)을 읽는 쪽(설정 명령)은 다른 카드의 기준이에요(D-2). SPEC은 draft예요: 승인은 사용자의 선택이라 이 레인이 하지 않았어요.
