---
id: SPEC-20261122-headless-run-contract
title: "Every claude -p run moai starts pins its permission mode, MCP config, model, output schema and budgets, and the result is checked before it is trusted"
status: draft
revision: 2
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# Every claude -p run moai starts pins its permission mode, MCP config, model, output schema and budgets, and the result is checked before it is trusted

## Intent

흐름 재설계의 헤드리스 조각. moai가 스스로 띄우는 모든 `claude -p` 실행(감사 실행기, `-p` 작업자, 깨우기 응답 실행, GLM 게이트웨이 실행)이 한 실행기 함수를 거치도록 묶고, 그 실행기가 권한 모드·MCP 구성·모델·출력 스키마·예산을 항상 고정하고, 결과를 믿기 전에 검사해요. 근거가 된 공식 문서 사실(2026-11-22 받은 원문에서 확인): (1) Claude Code v2.1.302부터 아무것도 모드를 정하지 않은 `claude -p`는 수동 모드로 시작해 승인이 필요한 호출을 거절해요(permission-modes.md "Tool calls denied in `claude -p` and Agent SDK runs"). (2) `-p`는 프로젝트 `.mcp.json`의 서버를 묻지 않고 읽어요(mcp.md). (3) `--bare`가 `-p`의 미래 기본이 될 예정이라 예고됐고(headless.md "`--bare` … will become the default for `-p` in a future release") `--bare`는 훅·스킬·CLAUDE.md·자동 기억을 건너뛰어요. (4) `-p`에서는 Fable 사용 크레딧 청구에 동의를 구하지 않고 청구해요(model-config.md "Claude Code bills it without asking"). (5) 결과의 `success` subtype만으로는 답이 나왔다는 뜻이 아니에요(agent-sdk structured-outputs.md "A `success` subtype doesn't by itself mean the run produced an answer"). (6) 무인 실행은 `CLAUDE_CODE_RETRY_WATCHDOG=1`로 용량 오류를 기다려야 하고(env-vars.md), 백그라운드로 넘어간 명령 때문에 `-p`가 명령 끝을 안 기다리고 exit 0할 수 있어 `BASH_DEFAULT_TIMEOUT_MS`를 올려야 해요(headless.md). (7) `-p`에서는 `ANTHROPIC_API_KEY`가 있으면 항상 그 키로 청구해요(authentication.md "the key is always used when present"). 사용자 선택(2026-11-22) ③: `permissions.denyInsteadOfPrompting`은 레인 실행에만 `--settings`로 주고 주 세션은 지금처럼 물어요. 이 조각은 SPEC 만 쓰고(코드는 이어지는 조각), 테스트는 Rust 만이며 돈이 드는 실제 모델 호출은 하지 않아요.

## Non-goals

- 인터랙티브 세션과 오르카 레인 띄우기(lane start/stop), 작업트리 생성 — 각자의 카드.
- 감사 에이전트 정의의 앞머리 글과 `/code-review` 증거 채택 — 에이전트·스킬 카드.
- 원장에 판정을 쓰는 형식화된 명령의 문법(typed verdict 행) — 3단계 조각; 이 SPEC 은 실행 한 번의 기록 행만 정해요.
- GLM 게이트웨이 프로필(기본 URL·모델 별칭·키 위임)의 세부 값 — 별도 카드; 이 SPEC 은 GLM 실행도 같은 실행기를 지난다는 규칙만 정해요.
- 원격·클라우드 세션(`--cloud`, `--environment`)과 CI 러너 — 이 실행기가 띄우는 것은 로컬 `claude -p`뿐이에요.
- 상태줄·mods·출력 스타일.

## Requirements

- REQ-1: When moai starts any headless Claude run, the runner shall build the `claude -p` command in one function that every caller goes through and that refuses to build a command missing a pinned flag or an audit run whose requesting session is the change's author session. · failure: none — a caller that cannot name its pinned values, or an audit caller that is the author or cannot name the author, gets a build error before any process starts, and the requesting session is recorded on the built run.
- REQ-2: The runner shall always pass exactly one of `--permission-mode auto` or `--permission-mode dontAsk` together with `--permission-prompts none` and `--allowedTools` entries only as narrow rules that survive auto mode's broad-rule drop, recording the chosen mode and its reason in the run's ledger row. · failure: none — a blanket `Bash` or `Bash(*)`-style entry is a build error, and no run starts in the Manual default where approval-needing calls are denied.
- REQ-3: The runner shall pass `--strict-mcp-config` and no `--mcp-config` when no platform `managed-mcp.json` is deployed and neither of the two flags when one is. · failure: none — with the platform file deployed both flags make Claude Code exit at startup, and without the flag pair a `-p` run loads project `.mcp.json` servers without asking.
- REQ-4: The runner shall always pass `--model` with a value the caller names and `--agent` with the agent definition name for audit runs. · failure: none — a run without a pinned model is a build error, because `-p` bills a Fable request to usage credits without asking, and the audit contract lives in the agent definition.
- REQ-5: The runner shall pass `--output-format json` together with `--json-schema` for every run whose answer is read by code. · failure: none — a verdict read from free text is a build error.
- REQ-6: The runner shall always pass `--max-turns` and `--max-budget-usd` with positive values that leave headroom above the expected answer cost. · failure: none — the CLI rejects a `0` budget at startup and spend can pass the cap by one response, so `0` and unset are both build errors.
- REQ-7: The runner shall always pass `--setting-sources user,project,local` and never `--bare`. · failure: none — a run that relies on the default sources or a future bare default is a build error, because `--bare` skips hooks, skills, CLAUDE.md and auto memory and stops the session from receiving messages.
- REQ-8: The runner shall pass exactly one `--settings` value whose inline JSON object carries `crossSessionInbound: accept` for worker runs and additionally `permissions.denyInsteadOfPrompting: true` for lane runs only. · failure: none — a plain audit run that carries the lane key or a lane run without it is a build error, and the main session keeps asking because the key never enters a settings file.
- REQ-9: When an audit is reproduced in a repository moai does not own, the runner shall replace the setting sources with `user` only or add `disableAllHooks: true` to the `--settings` object so that tree's hooks do not run. · failure: none — the default sources stay for moai's own repository.
- REQ-10: The runner shall set `CLAUDE_CODE_RETRY_WATCHDOG=1` and a raised `BASH_DEFAULT_TIMEOUT_MS` in the child environment for every run and, when `ANTHROPIC_API_KEY` is present, print one warning that names the key's presence and billing consequence without the key's value or a file that holds it, keeping every credential value out of the child argv and environment except a credential the caller names for a gateway run. · failure: none — an unattended run that fails on a transient capacity error or reports success before a long command finishes is a finding, the warning text is checked to contain no credential material, and the scrubbed credential names are recorded without their values.
- REQ-11: The runner shall treat a run as answered only when the process exited zero and the JSON result carries `is_error` false and the expected `structured_output` field. · failure: none — a `success` subtype without `structured_output`, a refusal `stop_reason` and a non-zero exit are all recorded as no verdict.
- REQ-12: The runner shall record one ledger event row per run with the kind `claude.run` and a JSON payload holding the verdict, `session_id`, `total_cost_usd` marked as an estimate, `num_turns`, the terminal reason and `permission_denials`. · failure: none — the row fits the existing free-text event table with no schema change.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10): a stand-in `claude` binary placed first on the test PATH records the exact argv and environment of every moai-started run, each pinned flag of REQ-2 through REQ-10 is present in what it recorded, and neither the recorded argv nor the recorded environment dump contains the credential value planted in the parent environment.
  verify: `cargo test --locked --offline --test headless_run_cli standin_records_every_pinned_flag` — red first in the code slice
- AC-2 (REQ-3): with a faked platform root carrying a `managed-mcp.json`, the built command carries neither `--strict-mcp-config` nor `--mcp-config`, and without the file it carries `--strict-mcp-config` and no `--mcp-config`.
  verify: `cargo test --locked --offline --test headless_run_cli managed_mcp_adapts_the_flags` — red first
- AC-3 (REQ-8): a lane run's single `--settings` JSON carries `crossSessionInbound` accept and `permissions.denyInsteadOfPrompting` true, a plain audit run's carries only the inbound key, and no run writes the key into any settings file.
  verify: `cargo test --locked --offline --test headless_run_cli settings_object_is_one_and_role_scoped` — red first
- AC-4 (REQ-6): a budget of `0` or an unset budget refuses to build the command before any process starts.
  verify: `cargo test --locked --offline --test headless_run_cli zero_or_missing_budget_refuses_to_build` — red first
- AC-5 (REQ-11): a result with subtype `success` and no `structured_output` field is treated as no verdict, as are a refusal `stop_reason` and a non-zero exit.
  verify: `cargo test --locked --offline --test headless_run_cli success_without_structured_output_is_no_verdict` — red first
- AC-6 (REQ-12): one finished run writes exactly one `claude.run` event row whose payload names the verdict, `session_id`, the cost estimate marker, `num_turns`, the terminal reason and `permission_denials`.
  verify: `cargo test --locked --offline --test headless_run_cli one_run_one_ledger_row` — red first
- AC-7 (REQ-10): with `ANTHROPIC_API_KEY` set in the test environment the runner prints the presence warning and the warning contains none of the key's value.
  verify: `cargo test --locked --offline --test headless_run_cli api_key_warning_names_presence_only` — red first
- AC-8: the wider READY set passes.
  verify: `cargo test --locked --offline --bins` and `--lib`
- AC-9 (REQ-1): an audit run whose requesting session equals the author session refuses to build before any process starts, and so does an audit run that names no author.
  verify: `cargo test --locked --offline --test headless_run_cli author_session_audit_refuses_to_build` — red first

## Scope

- write: `.moai/specs/SPEC-20261122-headless-run-contract/**`, `src/headless_run.rs`, `tests/headless_run_cli.rs`, and the one `pub mod headless_run;` registration line in `src/lib.rs` (added under `card line-lease`, the shared-file lease, when the path is free)
- later code slices on this card (joined by card scope when their turn comes): the `claude.run` ledger event writer, and the call sites that adopt the runner (the audit runner first, the worker wake and the GLM gateway runner on their own cards)
- read-only: `src/ledger/schema.rs` (the event table the run row fits), `.moai/specs/SPEC-20261121-flow-skeleton/**`
- forbidden: a real `claude` process or any model call inside tests; printing or logging any credential value or key file; `--bare`; `--dangerously-skip-permissions`; force pushes; pushes of the integration branch

## Plan

1. This slice: this SPEC, committed on its branch, READY note with head and base; a moved tip is not re-merged.
2. Code slice A: the runner module and the stand-in test target, red tests first — the pinned argv and environment, the author-session refusal and the credential scrub (REQ-1–REQ-10, AC-1–AC-4, AC-7, AC-9); the `src/lib.rs` registration line joins the slice by `card line-lease` when the shared file is free.
3. Code slice B: the result check and the ledger row (REQ-11–REQ-12, AC-5–AC-6).
4. Code slice C: the audit runner call site adopts the common function; the worker wake and the GLM gateway runner adopt it on their own cards.

## Risks

- 공식 문서가 실험 단계에서 바뀌어요(headless·agent-sdk 문서의 안내 문구): 플래그 세트가 테스트에 고정되므로 행동이 바뀌면 테스트가 먼저 알려줘요.
- `managed-mcp.json`은 실행 순간에 배포되거나 사라질 수 있어요: 실행기는 프로세스를 띄우기 직전에 검출하고 검출 결과를 실행 행에 남겨요(REQ-3, REQ-12).
- `--settings` 인라인 JSON은 셸 인용 규칙에 걸려요: 실행기가 인자 배열로 곧장 전달해 셸을 거치지 않게 해요(테스트가 녹음한 argv 가 증거).
- 문장만 보면 `success` 인 척하는 실패가 늘어요: 판정 없음 규칙(REQ-11)이 원장 행의 `terminal reason` 과 함께 잡아요.
- 키 값이 경고·로그·원장 어디에도 새는 것: AC-7 의 문구 검사와 감사가 지켜요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 권한 모드의 기본 배정: 읽기 전용 실행(감사 등)은 `dontAsk`+좁은 allow 규칙, 레인 작업 실행은 `auto`; 호출자가 다른 값을 주면 사유를 실행 행에 기록한다. | permission-modes.md (dontAsk = 미리 승인된 것만, 잠긴 CI용; auto = 분류기 검사), 카드 본문 "--permission-mode auto or dontAsk" | 카드 열림 |
| D-2 | `managed-mcp.json` 배포 검출 후 적응: 파일이 있으면 `--strict-mcp-config`·`--mcp-config` 둘 다 붙이지 않는다(둘 다 시작 거부). | managed-mcp.md "The `--strict-mcp-config` flag … exits at startup", "--mcp-config … exits at startup" — 이 SPEC 조사에서 정확히 잡은 사실(보고서 8-1 의 해당 줄을 정확화) | 카드 열림 |
| D-3 | 실행 설정은 `--settings` 하나의 인라인 JSON 객체로 합친다: 작업자는 `crossSessionInbound: accept`, 레인 실행만 여기에 `permissions.denyInsteadOfPrompting: true` 를 더한다. | cross-session-messaging.md "start it with `crossSessionInbound` set to `accept` in its `--settings` value", 사용자 선택 ③(2026-11-22) | 카드 열림 |
| D-4 | 원장의 비용은 `total_cost_usd` 를 "추정"으로 표시한다. | agent-sdk cost-tracking.md "client-side estimates, not authoritative billing data" | 카드 열림 |
| D-5 | 테스트는 가짜 claude 이진(스탠드인)만 쓰고 실제 모델 호출·네트워크를 하지 않는다. | "테스트는 Rust 만" 결정, 카드 기준 1 | 카드 열림 |
| D-6 | `--bare` 는 금지, `--setting-sources user,project,local` 을 항상 명시한다. | headless.md "`--bare` … will become the default for `-p` in a future release"; `--bare` 는 수신 소켓도 안 묶는다(cross-session-messaging.md) | 카드 열림 |
| D-7 | 실행 기록은 자유 이벤트 행(kind `claude.run` + JSON payload)에 남긴다 — 스키마 변경 없음. | `src/ledger/schema.rs` event 표(kind TEXT, payload JSON) | 카드 열림 |
| D-8 | `ANTHROPIC_API_KEY` 경고는 존재와 과금 결과만 말하고 값·파일은 절대 말하지 않는다(GLM 레인 규칙과 같은 선). | authentication.md "the key is always used when present", 리더 지시 "GLM 키 값·키 파일은 절대 출력하지 말기" | 카드 열림 |
| D-9 | 감사 실행은 작성자 세션에서 오면 거절한다(fail closed): 요청 세션과 작성자 세션이 같거나 작성자를 못 대면 둘 다 빌드 거부, 요청 세션은 실행 기록에 남긴다. | 리더 MERGED-LOCAL 노트(2026-11-22) "the runner records the requesting session and refuses an audit run requested by the card's author session", "자기 작업은 자기가 감사하지 않기" 규칙 | 카드 열림 |
| D-10 | 자식 환경에서 자격 증명 값(`ANTHROPIC_API_KEY`·`ANTHROPIC_AUTH_TOKEN`·`CLAUDE_CODE_OAUTH_TOKEN`)과 세션 정체 변수(`CLAUDECODE`·`CLAUDE_CODE_CHILD_SESSION`·`CLAUDE_CODE_SESSION_ID`)를 지우고, 호출자가 게이트웨이용으로 명시한 자격 증명 하나만 다시 넣는다; 지워진 이름(값 아님)을 기록한다. `BASH_DEFAULT_TIMEOUT_MS` 는 600000(기본 120000 의 5배, 문서 예시 300000 과 상한 기본 600000 의 위)로 올린다. | 리더 MERGED-LOCAL 노트(2026-11-22) "no credential value appears in argv or the recorded environment dump", headless.md BASH_DEFAULT_TIMEOUT_MS=300000 예시, env-vars.md 기본 120000·상한 관계 | 카드 열림 |

## Result

(status draft — written when the slice lands)
