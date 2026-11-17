---
id: SPEC-20261117-edit-diagnostics
title: 편집 직후 진단 hook (LSP), 켜야만 동작하는 설정
status: draft
revision: 2
created: 2026-11-17
parent: SPEC-20261117-legacy-verdicts
---

## Intent

C-11 재설계 3번. 에이전트가 파일을 쓰거나 고친 직후, 그 파일의 언어 진단(타입·컴파일 오류)을 바로 세션에 돌려주는
hook을 설계해요. 사용자가 켜야만 동작하고(opt-in), 도구를 설치하지 않고, 네트워크를 쓰지 않고, 편집을 막지
않아요(사용자가 막기 모드를 켰을 때만 모델에게 "고쳐야 한다"고 알려요). 이 문서는 설계만 담아요. 코드는 아래
Plan의 조각(slice)마다 따로 SPEC 승인 뒤에 써요.

배경 — Go 세대(개념만 참고, 옮기지 않음): `legacy-asset-map.json`에서 `deferred(design:edit-time-diagnostics-hook)`인
10행(`moai-workflow-loop` SKILL.md·references 2개, `workflows/fix.md`·`loop.md`, 각각 저장소 사본과 템플릿 사본).
그 설계는 `handle-post-tool.sh`(PostToolUse)가 Write/Edit 뒤 LSP 진단을 `additionalContext`로 돌려주고, 오류가
있으면 exit 2, `ralph.yaml`의 `hooks.post_tool_lsp.enabled`(기본 켜짐)·`severity_threshold: error`·
`lsp.timeout_seconds: 30`, LSP가 없으면 linter → 컴파일러 → 시험 순으로 대신하고, `MOAI_DISABLE_LSP_DIAGNOSTIC`로
끄는 방식이었어요. 여기서는 기본 꺼짐, 설치 없음, 짧은 시간 한도로 바꿔요.

지금 있는 것(이 저장소, 기준 `d2b93a420c`):
- `src/hooks.rs` `HOOK_SET`: PostToolUse는 matcher `Bash`의 receipt만 있고, 모든 handler에 `timeout` 10초
  (`HOOK_TIMEOUT_SECS`).
- `src/hook_engine.rs`: 의무(duty) 4개(session-check, tool-preflight, tool-observe, turn-validation). tool-observe는
  PostToolUse에서 기록만 하고 막지 않아요. 모델에 가는 글은 고정 문장과 열거 값만 쓰고, 도구 입력·파일 내용은
  되풀이하지 않아요. 네트워크 없음.
- `src/hook_config.rs`: HOME 상태 폴더의 `hook-config.toml`에 `[hooks.outside_project_writes]`,
  `[hooks.primary_checkout]`, `[hooks.commit_identity]`가 있고 모두 기본 꺼짐. 파일이 없거나 링크·64 KiB 초과·
  TOML 오류면 꺼짐(fail-open)이고 `config_error` 한 줄을 남겨요.
- `src/project_detect.rs` `LSP` 표: 언어 13개 → 서버 이름(`rust-analyzer`, `pyright`, `typescript-language-server`,
  `gopls`, `jdtls`, `kotlin-language-server`, `ruby-lsp`, `intelephense`, `clangd`, `omnisharp`, `sourcekit-lsp`).
  "설치됐다는 뜻은 아님". `commands` 사실은 build·test 명령만 있고 파일 하나를 검사하는 명령은 없어요.
- 품질 참고 문서(quality-guides 가지)는 원격에 없어 읽지 않았어요(`git fetch` 결과 "couldn't find remote ref").

공식 문서(2026-11-17 09:43–09:45Z에 읽음):
- Claude Code, https://code.claude.com/docs/en/hooks — PostToolUse는 "After a tool call succeeds"에 돌고, matcher는
  도구 이름(`Edit|Write` 예시)이에요. 파일 도구 경로는 "always absolute"예요. 출력은 `hookSpecificOutput.additionalContext`
  ("next to the tool result")와 최상위 `decision: "block"`+`reason`("adds the reason next to the tool result"; 편집은
  되돌리지 않음)이에요. exit 2는 "Shows stderr to Claude; the tool already ran"이에요. `additionalContext`와 stdout은
  "capped at 10,000 characters"이고, 넘으면 파일로 저장되고 2,000자 미리보기만 가요. command hook `timeout` 기본값은
  600초, 같은 이벤트의 hook들은 "run in parallel"이에요. `async: true`면 다음 턴에 `additionalContext`를 전달하고
  막을 수 없고, timeout을 강제하지 않으며, `-p`에서는 끝날 때 취소돼요.
- Codex, https://learn.chatgpt.com/docs/hooks — PostToolUse는 Bash와 `apply_patch`(matcher `Edit`·`Write`도 받지만
  `tool_name`은 `apply_patch`)에 돌고, 경로 전용 필드는 없어요(`tool_input.command`). 출력은
  `hookSpecificOutput.additionalContext`, 일반 stdout은 무시, exit 2+stderr 또는 `decision: "block"`(도구 결과를
  피드백으로 바꿈)이에요. `timeout` 기본 600초, hook 출력은 "roughly 2,500 tokens"로 제한돼요. 프로젝트 hook은
  검토·신뢰 뒤에만 돌아요.
- OMP, https://raw.githubusercontent.com/can1357/oh-my-pi/main/docs/extensions.md — `tool_result`는 "post-exec, may
  patch content/details/isError or return passive `additionalContext`"이에요. payload 필드 목록과 handler 시간 한도는
  문서에 없어요.

## Non-goals

- 도구 설치나 내려받기(언어 서버, linter, 컴파일러).
- 저장소 전체 검사, 시험 실행, 반복 수정 루프(Go `/moai loop`·`fix`의 goal 엔진). 그것은 VERIFY·RUN 경로 몫이에요.
- 이 SPEC 안에서 코드나 다른 파일 쓰기.
- PreToolUse 보호 규칙(tool-preflight) 바꾸기.

## Requirements

- REQ-1: Where the `[hooks.edit_diagnostics]` table is absent or its `enabled` is not `true`, the edit-diagnostics duty shall run no checker and exit 0 with nothing on stdout. · failure: none — off is the safe state.
- REQ-2: When the host reports a finished file edit inside the project, the duty shall run the one checker the setting names for that file's language, with the project root as its working folder and no shell.
- REQ-3 (REQ-1): IF the setting file is unusable or the table is malformed, THEN the duty shall treat the feature as off and write one `config_error` line to the engine log.
- REQ-4 (REQ-2): IF the edited path is a symbolic link, has a linked folder on its way, or resolves outside the project, THEN the duty shall run no checker and exit 0.
- REQ-5 (REQ-2): IF the named checker is not found or cannot start, THEN the duty shall return one context line naming the checker and exit 0.
- REQ-6 (REQ-2): IF the checker runs past `timeout_secs`, THEN the duty shall stop its whole process group and return one timeout line and exit 0.
- REQ-7 (REQ-2, REQ-8): IF the checker output passes the read bound or holds bytes that are not UTF-8, THEN the duty shall read no more than the bound, replace the bad bytes and say so in its summary line.
- REQ-8: When the checker ends, the duty shall return at most `max_lines` diagnostic lines and at most `max_chars` characters through the host's documented context field.
- REQ-9: Where `mode` is `block` and the checker reports an error, the duty shall answer with the host's documented PostToolUse block field. · failure: none — a failing checker never blocks (REQ-5 to REQ-7 exit 0 with context only).
- REQ-10: The duty shall install nothing and open no network connection. · failure: none — it only starts the command the user named.
- REQ-11: The duty shall finish within 50 ms of its own time when the feature is off and add at most 200 ms to the checker's own time when it is on. · failure: none — a budget measured by AC-6.
- REQ-12: `moai hooks-plan` shall list the duty for each host with its event and answer and what it does not cover. · failure: none — text in the capabilities answer.

## Acceptance

S1이 맡는 것: AC-1~AC-4, AC-6, AC-7(Claude Code 항목과 Codex·OMP의 "지원 안 함" 항목). AC-5는 S2예요. 시험 파일은
`tests/edit_diagnostics_cli.rs`이고, 아래 verify의 시험 이름이 그 안에 있어요.

- AC-1 (REQ-1, REQ-3): 설정이 없음·`enabled = false`·다른 표만 있음 → hook 줄 없음, checker 안 돎, stdout 빔, exit 0.
  표 형식 오류(`mode = "block"`, 범위 밖 숫자 N/N+1, 모르는 키, 잘못된 checker, 불린이 아닌 `enabled`) → 꺼짐과
  `config_error`(`invalid_edit_diagnostics`) 한 줄; 링크된 설정 파일 → 꺼짐과 `not_regular_file`(임시 HOME).
  verify: `cargo test --locked --offline --test edit_diagnostics_cli off_without_the_setting`; `cargo test --locked --offline --test edit_diagnostics_cli a_malformed_setting`
- AC-2 (REQ-2, REQ-8): 가짜 checker(시험이 쓴 스크립트)가 진단 세 줄을 내면 Claude Code 형식
  `hookSpecificOutput.additionalContext`에 프로젝트 기준 경로로 세 줄과 요약 한 줄이 가고, 파일 내용은 없어요.
  `max_lines` N과 N+1, `max_chars` 경계(요약 포함 전체)에서 잘림과 "not shown" 표시가 맞아요.
  verify: `cargo test --locked --offline --test edit_diagnostics_cli diagnostics_come_back_as_context`
- AC-3 (REQ-4): 링크 파일, 링크 폴더를 지나는 경로, `..`, 상대 경로, 없는 파일, 폴더, 프로젝트 밖 절대 경로 —
  모두 checker가 돌지 않아요(가짜 checker가 표시 파일을 만들지 않음).
  verify: `cargo test --locked --offline --test edit_diagnostics_cli files_outside_the_project`
- AC-4 (REQ-5, REQ-6, REQ-7): checker 없음 → 이름 한 줄; `timeout_secs` 1에서 30초 자는 checker → 시간 초과 한 줄과
  백그라운드 자식 프로세스까지 끝남; 2 MiB 출력과 UTF-8 아닌 바이트 → 한도까지만 읽고 대체 표시; 비밀 모양 줄은
  가려짐; 같은 프로젝트에서 두 번째 검사는 "건너뜀" 한 줄. 모두 exit 0.
  verify: `cargo test --locked --offline --test edit_diagnostics_cli failures_are_one_line`; `cargo test --locked --offline --test edit_diagnostics_cli credential_shaped`; `cargo test --locked --offline --test edit_diagnostics_cli one_check_runs`
- AC-5 (REQ-9): `mode = "block"`에서 오류가 있으면 Claude Code는 최상위 `decision: "block"`+`reason`, Codex는
  문서의 block 답(조각 S2), 오류가 없거나 checker가 실패하면 block하지 않아요.
  verify: `cargo test --locked --offline --test edit_diagnostics_cli block_`
- AC-6 (REQ-10, REQ-11): 시험 중 네트워크 없는 환경에서 통과하고, 소스에 네트워크 API 사용이 없으며(기존
  `home_isolation_scan` 방식의 스캔), 꺼진 상태 21회 실행의 중앙값이 50 ms 아래예요(프로세스 시작 포함, 측정값은
  Result에 적어요). 켜진 상태의 추가 시간 200 ms 측정은 S3(LSP)에서 해요.
  verify: `cargo test --locked --offline --test edit_diagnostics_cli budget_ -- --ignored`; `cargo test --locked --offline --test home_isolation_scan`
- AC-7 (REQ-12): `hooks-plan` capabilities에 세 host의 edit-diagnostics 항목(이벤트, 답, 다루지 않는 것)이 있어요.
  verify: `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --lib hook_engine`

## Scope

- write (S1): `.moai/specs/SPEC-20261117-edit-diagnostics/spec.md`, `src/hook_engine.rs`, `src/hook_config.rs`,
  `src/hooks.rs`, `src/edit_diagnostics.rs`(새 모듈), `src/lib.rs`, `src/main.rs`(`hook-dispatch` 도움말),
  `tests/edit_diagnostics_cli.rs`. S2~S4가 더 쓸 곳: hook 참고 문서, `tools/legacy_verdicts.py`
- read-only: `src/hook_guards.rs`, `src/project_detect.rs`, `.moai-cli/manifests/legacy-asset-map.json`,
  `tools/legacy_verdicts.py`, Go 저장소의 `moai-workflow-loop`·`workflows/fix.md`·`workflows/loop.md`(개념만)
- forbidden: 도구 설치, 네트워크, 실제 HOME, S1에서 막기 모드·Codex·OMP·LSP

## Plan

1. S0 (이 SPEC): 설계와 열린 질문. revision 2에서 리더 결정(Q1~Q6, D8~D13)을 넣었어요.
2. S1 — 설정과 Claude Code(이 revision에서 구현): `[hooks.edit_diagnostics]` 읽기(REQ-1, REQ-3), 새 의무
   `edit-diagnostics`와 `hook-dispatch --event edit-diagnostics`, 설정이 켜져 있을 때만 생성하는 PostToolUse matcher
   `Write|Edit|MultiEdit` 줄(D11), 사용자가 적은 확장자별 checker 명령 실행, 경로 검사, 시간 한도, 출력 상한, 문맥 답
   (REQ-2, REQ-4 ~ REQ-8, REQ-10 ~ REQ-12). AC-1 ~ AC-4, AC-6, AC-7.
3. S2 — 막기 모드와 다른 host: `mode = "block"`(REQ-9), Codex `apply_patch` 경로(preflight가 이미 쓰는
   `*** Add/Update File:` 줄 읽기, 문서에 정의 없음), OMP `tool_result`의 `additionalContext`. AC-5, AC-7.
4. S3 — LSP: D9에 따라 측정 뒤 언어 서버 연결(`project_detect` `LSP` 표의 서버가 PATH에 있을 때만, 설치 없음).
   `textDocument/didOpen` 뒤 `textDocument/publishDiagnostics`를 시간 한도 안에서 기다려요.
5. S4 — 마무리: hook 참고 문서, `legacy-asset-map.json`의 10행 판정을 이 설계로 닫기(`tools/legacy_verdicts.py`).

## Risks

- checker 출력에 지시처럼 보이는 글(예: "이전 지시를 무시하라")이 섞여도 걸러 낼 수 없어요 → 진단 첫 줄 200자,
  `max_lines`·`max_chars`, 제어 문자·ANSI 제거, 비밀 모양 가리기로 양을 줄일 뿐이에요. 사용자가 고른 도구만 돌아요(감사 N1).
- 파일 이름이 `-`로 시작하면 checker가 옵션으로 읽을 수 있어요(CWE-88) → 늘 `./<rel>`을 넘겨요(감사 B1).

- 언어 서버는 처음 뜰 때 느려요(큰 Rust 저장소의 `rust-analyzer`는 수십 초) → 매 편집마다 새로 띄우면 10초 안에
  못 끝나요. 그래서 S1은 사용자가 고른 파일 단위 checker로 시작하고, LSP는 측정 뒤 S3에서 해요(D9).
- 같은 이벤트의 hook은 병렬로 돌아서(Claude Code 문서), 빠른 연속 편집이면 checker가 겹쳐 돌 수 있어요 → 프로젝트당
  하나만 돌리고 나머지는 "건너뜀" 한 줄로 답해요(S1 시험에 포함).
- checker 출력에 소스 줄이 섞일 수 있어요(예: rustc 진단의 코드 조각) → 진단마다 첫 줄만, 줄당 200자로 잘라요.
  비밀이 섞일 가능성은 남아요(사용자 코드의 문자열) → 비밀 모양 줄은 가려요(D12).
- Codex는 `apply_patch`의 경로 필드를 정의하지 않아요 → S2에서 patch 본문을 읽는 방식은 문서 밖 형식에 기대요.
- Claude Code `async: true`는 막기 모드를 쓸 수 없고 `-p`에서 취소돼요 → 동기 실행(D10).
- 사용자가 적은 checker 명령 자체가 네트워크를 쓸 수 있어요(예: 의존성을 내려받는 빌드 도구) → MoAI는 그 명령에
  아무것도 더하지 않고, 설명서에 `--offline` 같은 선택을 권해요. MoAI 자체는 네트워크를 열지 않아요(REQ-10).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 기본 꺼짐(opt-in) | 카드 C-11 재설계 3번, Go의 기본 켜짐을 바꿈 | `enabled` 기본값을 true로 |
| D2 | 설정은 기존 hook 설정 파일의 새 표 `[hooks.edit_diagnostics]` (키: `enabled`, `mode` = `context`/`block`, `timeout_secs` 기본 10·최대 30, `max_lines` 기본 20, `max_chars` 기본 4,000, 언어별 `checkers`) | 다른 opt-in 규칙과 같은 자리와 같은 fail-open 규칙(`src/hook_config.rs`); 위치는 Q1에서 확정 | 표 이름·위치 바꾸기 |
| D3 | `max_chars` 4,000 | Claude Code 상한 10,000자, Codex 약 2,500 token 안에 들어가게 | 값 바꾸기 |
| D4 | `timeout_secs` 기본 10, 최대 30 | 기존 handler `timeout` 10초(`HOOK_TIMEOUT_SECS`), Go `lsp.timeout_seconds` 30 | 값 바꾸기 |
| D5 | checker 실패·시간 초과·없음은 언제나 exit 0과 문맥 한 줄, 막지 않음 | 엔진의 fail-open 규칙; 편집은 이미 끝남 | 실패도 막기 |
| D6 | 모델에 가는 글은 프로젝트 기준 경로, 줄·열, 심각도, 진단 첫 줄(200자)만 | `hook_engine`의 "파일 내용 되풀이 안 함" 규칙 | 전체 메시지 보내기 |
| D7 | 프로젝트당 checker 하나만 동시에 | 병렬 hook으로 같은 도구가 겹쳐 도는 것 막기 | 잠금 빼기 |
| D8 | Q1 → (a) 설정은 HOME 상태 폴더의 `hook-config.toml` 한 곳 | 리더 결정 2026-11-17(권장안) | 프로젝트별 설정 추가 |
| D9 | Q2 → (a) S1은 사용자가 적은 파일 단위 checker 명령만, LSP는 측정 뒤 S3 | 리더 결정 2026-11-17(권장안) | S1에서 언어 서버 연결 |
| D10 | Q3 → (a) Claude Code 동기 실행, 결과가 같은 턴에 도구 결과 옆에 | 리더 결정 2026-11-17(권장안) | handler에 `async: true` |
| D11 | Q4 → (a) 설정이 켜져 있을 때만 hook 줄 생성 | 리더 결정 2026-11-17(권장안) | 늘 생성 |
| D12 | Q5 → (a) 비밀 모양 진단 줄은 가려짐으로 바꿈 | 리더 결정 2026-11-17(권장안) | 그대로 보냄 |
| D13 | Q6 → (a) S1은 Claude Code만; Codex·OMP는 engine error(exit 1, 막지 않음)와 capabilities의 "지원 안 함" | 리더 결정 2026-11-17(권장안) | S1에서 세 host |
| D14 | S1의 `mode`는 `context`만; `block`은 형식 오류로 보고 꺼짐+`config_error` | S1 범위(막기는 S2) | `block`을 받아 문맥으로 처리 |
| D15 | 범위: `timeout_secs` 1–30, `max_lines` 1–200, `max_chars` 200–9,000(요약 포함 전체); checker는 소문자 확장자(1–16자) → argv, 편집한 파일을 마지막 인자로 붙임, 모르는 키는 형식 오류 | Claude Code 10,000자 상한 아래, 오타가 조용히 무시되지 않게 | 범위·키 바꾸기 |
| D16 | 가리기는 `redact::contains_secret_shape`(어디서든 셈) | 이 tip에는 C-90의 `crate::secret_shaped`가 아직 없음; 로그·문맥 쪽은 C-90 수리 뒤에도 어디서든 세는 쪽(D4 of C-90) | 다른 판정 함수 |
| D18 | checker에 늘 `./<rel>`을 넘김 | 감사 B1(CWE-88) | 상대 경로 그대로 |
| D19 | checker 출력에서 ANSI escape와 탭 밖의 제어 문자를 지움 | 감사 N1 | 그대로 넘김 |
| D20 | checker가 스스로 끝나도 그 process group을 SIGKILL | 감사 N2(뒤에 남은 자식과 열린 pipe) | 시간 초과 때만 |
| D21 | 잠금 파일은 HOME 상태 폴더의 `edit-diagnostics/`(0700, 링크 따라가지 않음); 상태 폴더가 없으면 잠금 없이 돌림. 프로젝트당 빈 파일 하나를 지우지 않고 다시 씀(SPEC-20261117-edit-diagnostics-followups D1: 지우면 경쟁) | 감사 N3(공용 임시 폴더 대신) | 시스템 임시 폴더 |
| D17 | 생성된 handler의 `timeout`은 `timeout_secs + 5` | host가 MoAI보다 먼저 hook을 끊지 않게 | 고정 10초 |

열린 질문 Q1~Q6은 2026-11-17 리더 결정으로 닫혔어요(D8~D13, 모두 권장안).

## Result

S1 — 구현 에이전트가 2026-11-17에 채웠어요. 감사가 다시 확인해요. 상태는 `draft` 그대로예요(승인은 사용자 몫).
기준 tip `6be96b4fc7`, 설계 커밋 `63efddb4f2`을 cherry-pick(`276d936c98`). 빌드 폴더는 이 worktree의 `target/`,
`CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, 무거운 작업 슬롯 안에서 cargo 한 번에 하나(2026-11-17 10:42–10:49Z).
RED는 mutation M1로 보였어요: 새 의무를 엔진에서 빼면 새 시험 7개가 실패해요.

- AC-1: met — `off_without_the_setting_no_hook_line_and_no_run`, `a_malformed_setting_is_off_and_logged_and_its_bounds_hold`.
  mutation M4(꺼져 있어도 hook 줄 생성) KILLED
- AC-2: met — `diagnostics_come_back_as_context_within_the_caps`(max_lines N/N+1, max_chars 200, 경로 숨김)
- AC-3: met — `files_outside_the_project_or_behind_a_link_are_not_checked`. mutation M2(`is_symlink` 검사 빼기)는 살아남았어요:
  `symlink_metadata`로 본 링크는 파일도 폴더도 아니라서 형 검사가 이미 거절해요(같은 동작의 mutant). 대신 M2b(링크를
  따라가는 `metadata`) KILLED
- AC-4: met — `failures_are_one_line_and_never_block`, `credential_shaped_diagnostics_are_masked`,
  `one_check_runs_per_project_at_a_time`. mutation M3(시간 초과 때 process group을 죽이지 않음) KILLED, M5(가리기 빼기) KILLED
- AC-6: met — 꺼진 상태 21회 중앙값 50 ms 안(측정 debug 9–21 ms; `budget_the_off_path_is_fast -- --ignored`, 프로세스 시작 포함);
  `home_isolation_scan` 16
- AC-7: met(S1 범위) — capabilities에 Claude Code `edit-diagnostics` 항목, Codex·OMP는 `supported: false`;
  `product_cli` 75, `--lib` edit_diagnostics·hook_config·hook_engine·hooks 42
- 이웃 검사: `edit_diagnostics_cli` 8(+1 ignored), `hooks_cli` 22, `hook_dispatch_cli` 56, `hooks_pre_p4_fixtures` 6,
  `help_layout` 10, `git_env_scrub` 3, `spec_cli` 17, `--bin moai` 27; fmt 0, `rustfmt src/ops` 0, clippy 0; Python
  `repo_scans`·`ci_scope` 44개 OK; `moai spec check --all` 0

실행 안 함: 전체 `cargo test`, Linux·Windows, 실제 Claude Code 세션에서 hook이 도는지(host 실행 없음), AC-5(S2).

S1 감사 수리(B1, N1~N3; D18~D21): `bcf1eae9b6`, tip `17c19a2759` 병합 `cd7954b50f`(manifest는 write로 다시 만들었고,
tip에 두 번 있던 `references/internal/build-folder.md` 키가 하나로 정리됨). RED(2026-11-18 16:27–16:28Z): 새 시험 4개가
고치기 전 코드에서 실패 — 옵션 모양 이름이 그대로 넘어감, 제어 문자·ESC가 그대로 감, 뒤에 남은 자식이 pipe를 잡아
출력이 사라짐, 잠금이 HOME에 없음. 병합된 head에서(무거운 작업 슬롯 안, 16:30–16:37Z): `edit_diagnostics_cli` 12(+1 ignored),
`hooks_cli` 22, `hook_dispatch_cli` 56, `hooks_pre_p4_fixtures` 6, `home_isolation_scan` 16, `product_cli` 75, `help_layout` 10,
`git_env_scrub` 3, `spec_cli` 17, `--lib` 43, `--bin moai` 27; fmt·`rustfmt src/ops`·clippy 0; Python 44개 OK; `spec check --all` 0;
꺼진 상태 중앙값 15.58 ms. mutation B1(`./` 빼기)·N1(정리 빼기)·N2(정상 종료 뒤 group 남김) 모두 KILLED.
