---
id: SPEC-20261126-hook-unreadable-input
title: "PreToolUse hook entries refuse an input they cannot read"
status: draft
revision: 2
created: 2026-11-26
---

# PreToolUse hook entries refuse an input they cannot read

## Intent

MoAI 프로젝트에서 `moai hook bash-boundary` 와 `moai hook write-boundary` 는 PreToolUse 관문이에요. 지금은 입력을 읽지 못하면 영수증 한 줄만 남기고 아무것도 출력하지 않아서, Claude Code 가 그 도구 호출을 그대로 실행해요. SPEC-20261126-hook-write-boundary(아래에서 "쓰기 경계 SPEC")가 `write-boundary` 의 4 MiB 초과 하나만 거절로 닫았고, 나머지는 그 SPEC 의 R-10 과 `docs/later.md` 한 줄로 남겼어요. 이 SPEC 은 그 남은 부분을 닫아요: 읽지 못한 입력은 두 이름 모두 거절해요. 판정은 더 엄격해지는 방향으로만 바뀌어요.

| 입력 | `bash-boundary` 지금 → 이 SPEC | `write-boundary` 지금 → 이 SPEC |
| --- | --- | --- |
| 4 MiB 초과 | 통과 → 거절 | 거절 → 그대로 (쓰기 경계 SPEC REQ-13) |
| UTF-8 아님 | 통과 → 거절 | 통과 → 거절 |
| JSON 아님 (빈 입력 포함) | 통과 → 거절 | 통과 → 거절 |
| JSON 객체 아님 | 통과 → 거절 | 통과 → 거절 |

### Sources

- 공식 hooks 문서 `https://code.claude.com/docs/en/hooks.md`: 2026-11-26T04:01:27Z 수신, 4,452줄, 301,732 바이트, sha256 `d16003851f60562eaa793e6270316b17550f8e24691214d7d798e5a503537c79`. 읽은 부분은 216행, 803~1001행, 1088~1127행, 1861~1890행, 2085~2144행이에요. 이 기기의 Claude Code 는 2.1.332 예요.
  - 2091행: PreToolUse 의 `permissionDecision: "deny"` 는 도구 호출을 막아요. 2096행: 여러 판정 중 `deny` 가 `defer`, `ask`, `allow` 보다 앞서요.
  - 216행: 종료 코드 0 에 출력이 없으면 훅의 판정이 없는 것이고 보통의 권한 흐름으로 도구 호출이 이어져요. 즉 지금의 침묵은 통과예요.
  - 809·831행: 훅이 실패하거나 출력이 해석되지 않으면 동작이 그대로 진행돼요. 848~856행·2098행: 종료 코드 2 도 막는 방법이에요.
  - 908행: 시간 초과한 PreToolUse 명령 훅은 막지 않아요(`onFailure: "block"` 은 예외). 942행: `onFailure` 는 v2.1.295 이상이 필요해요.
- **문서에 없는 것**: 훅이 받는 *입력*이 깨졌을 때의 동작은 저장한 문서 전체에서 찾지 못했어요(`malformed`, `invalid input`, `utf-8`, `surrogate` 검색 0건, `stdin` 이 나오는 행은 입력을 stdin 으로 준다는 설명과 예제뿐). 문서가 정하는 것은 훅의 *출력*이므로, 어떤 입력을 받았든 유효한 deny 객체와 종료 코드 0 이면 호출이 막힌다고 읽어요.
- **미확인**: 실제 Claude Code 가 UTF-8 이 아니거나, 객체가 아니거나, 짝 없는 대리 문자(lone surrogate) 이스케이프가 든 입력을 보내는 일이 있는지는 확인하지 않았어요.
- 코드: 기준 `41c16e60ac` 의 `src/hook_entry.rs` (`parse`, `failure`)와 착지 전 쓰기 경계 SPEC 의 head `3b69e2fd13` 의 같은 파일(`TOO_LARGE` 갈래, `expected_event`)·`src/hook_write_boundary.rs` (`over_limit`, `deny`).
- 설치 상태: 2026-11-26 에 이 저장소의 `.claude/settings.json`·`.claude/settings.local.json` 과 `~/.claude/settings.json` 에서 `bash-boundary`·`write-boundary` 는 0건이에요. 그래서 지금 돌고 있는 세션의 동작은 이 변경으로 바뀌지 않아요(설치는 훅 재설계의 단계 1-e).

## Non-goals

- 객체로 읽히는 입력의 판정. `hook_event_name` 이 다른 호출은 계속 판정 없이 영수증만 남겨요(쓰기 경계 SPEC REQ-8, 승인된 규칙). `command` 나 경로가 없거나 문자열이 아닌 객체는 각 처리기의 침묵 그대로예요.
- `session-context`. SessionStart 는 도구 호출을 막는 자리가 아니라서, 읽지 못한 입력에도 지금처럼 출력 없이 영수증만 남겨요.
- 훅 프로세스 자체의 실패(작업 폴더를 읽지 못함, 패닉, 시간 초과, 실행 불가). `src/main.rs` 는 바꾸지 않아요. `onFailure: "block"` 등록은 단계 1-e 예요.
- 훅 설치와 settings 항목, `src/hook_bash_boundary.rs` 의 판정, 입력 한도 4 MiB 값의 변경.

## Requirements

<!-- This SPEC makes a check that blocks a tool call stricter: a violation is a deny, never an ask or an allow. -->
- REQ-1: IF the standard input of `moai hook bash-boundary` or `moai hook write-boundary` in a MoAI project is not valid UTF-8, is not valid JSON (an empty input included) or is JSON that is not an object, THEN the entry shall print exactly one JSON object `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":<text>}}` on stdout and exit 0.
- REQ-2: IF the standard input of `moai hook bash-boundary` in a MoAI project is larger than 4 MiB, THEN the entry shall print the deny object of REQ-1 and exit 0, while the denial of an over-limit `write-boundary` input stays as it is.
- REQ-3: When the entry writes the reason of a denial of REQ-1 or REQ-2, the entry shall begin it with `moai-boundary/unreadable:` for `bash-boundary` and `moai-boundary/protected-path:` for `write-boundary`, name the cause, name one allowed way and that a refusal of every call is to be reported to the user, keep it under 400 bytes and carry no byte of the input · failure: none — the reason is fixed text per name and cause.
- REQ-4: When the entry denies under REQ-1 or REQ-2, the entry shall also record the one receipt line it records today (name, reason word, time, no input byte).
- REQ-5 (REQ-4): IF the receipt line of REQ-4 cannot be written, THEN the entry shall print the deny all the same and write the existing one stderr line.
- REQ-6: IF the project folder has no regular file `.moai/project.yaml`, THEN the entry shall print nothing on stdout and stderr and write nothing for every input of REQ-1 and REQ-2.
- REQ-7: IF the registered name is `session-context`, or the input is a readable object that names another event or lacks what its handler judges, THEN the entry shall deny nothing it does not deny today.

## Acceptance

- AC-1 (REQ-1, REQ-3): in a temporary MoAI project, for both `PreToolUse` names each of empty input, plain text, a lone surrogate escape, nesting deeper than the parser reads, a JSON array, a JSON scalar and invalid UTF-8 exits 0 and prints exactly one object with the three fields of REQ-1, a reason that starts with the class of its name, differs by cause, is under 400 bytes and holds no marker word of the input, with stderr empty · verify: `cargo nextest run --test hook_entry_cli -E 'test(/^unreadable_ac1_/)'`
- AC-2 (REQ-2, REQ-3): `bash-boundary` with an object of 4194305 bytes prints one deny object whose reason names the 4 MiB limit, the same object of 4194304 bytes gets no denial from the entry, and the `write-boundary` over-limit reason is the one it prints today · verify: `cargo nextest run --test hook_entry_cli -E 'test(/^unreadable_ac2_/)'`
- AC-3 (REQ-4, REQ-5): each denial adds exactly one receipt line with the name and the reason word and no input byte, and with no project folder under HOME, with a receipt path that is a link and with a full receipt log the call is still denied, one stderr line is written and nothing outside is touched · verify: `cargo nextest run --test hook_entry_cli -E 'test(/^unreadable_ac3_/)'`
- AC-4 (REQ-6): outside a MoAI project (no state folder, `project.yaml` a folder, a link to nothing, and `CLAUDE_PROJECT_DIR` naming such a folder) every input of AC-1 and AC-2 prints nothing on either stream and changes no file · verify: `cargo nextest run --test hook_entry_cli -E 'test(/^unreadable_ac4_/)'`
- AC-5 (REQ-7): `session-context` prints nothing for every refused input of AC-1 and AC-2 and denies nothing for the object of exactly 4194304 bytes, an object that names another event is silent for all three names, and a readable object without `tool_input` is silent for the two `PreToolUse` names · verify: `cargo nextest run --test hook_entry_cli -E 'test(/^unreadable_ac5_/)'`
- AC-6 (REQ-1, REQ-2, REQ-7): the entry, session start, write boundary, conformance and latency targets pass, with the earlier silent expectations changed only for the cases this SPEC denies · verify: `cargo nextest run --test hook_entry_cli --test hook_session_start_cli --test hook_write_boundary_cli --test hook_conformance_cli --test hook_latency_cli`
- AC-7 (REQ-1, REQ-6): the temp-home regression of the hook entry expects one deny object for bad input of the two `PreToolUse` names in the MoAI project, and silence for `session-context` and the plain folder · verify: `cargo nextest run --test e2e_temp_home_cli -E 'test(/^scenario_hook_entry/)'`
- AC-8 (REQ-1…REQ-7): the SPEC itself passes the spec check · verify: `moai spec check SPEC-20261126-hook-unreadable-input`

## Scope

- write: `.moai/specs/SPEC-20261126-hook-unreadable-input/**`, `src/hook_entry.rs` (the parse-failure branch, the Bash boundary reason and the module text), `src/hook_write_boundary.rs` (one function that builds the unreadable-input denial; the judgment is not changed), `tests/hook_entry_cli.rs` (new tests and the changed expectations), `tests/hook_session_start_cli.rs` (changed expectations only), `tests/e2e_temp_home_cli.rs` (the one hook entry scenario), `docs/later.md` (the line the write boundary SPEC added about unreadable input is removed, one line for R-3 is added).
- read-only: `src/hook_bash_boundary.rs`, `src/main.rs`, `src/hook_latency.rs`, `tests/hook_write_boundary_cli.rs`, `tests/hook_conformance_cli.rs`, `tests/hook_bash_boundary_cli.rs`, the official hooks documentation.
- forbidden: any settings file (`.claude/**` of the home and of the project), installed hooks, `src/hook_guards.rs`, `src/hook_engine.rs`, the real HOME and its `.moai/projects` folder. Tests use temporary HOMEs and folders.

## Plan

1. (this slice) SPEC, status draft; the approval is not written by a lane. Until it is approved nothing below starts.
2. SPEC-20261126-hook-write-boundary lands first; then the integration tip is brought into this branch once (it holds only this SPEC until then).
3. Tests first (RED): the `unreadable_ac*` tests and the changed expectations.
4. `src/hook_write_boundary.rs` builder, then the parse-failure branch of `src/hook_entry.rs` for the two `PreToolUse` names.
5. Guard mutants for each new guard test (`guard-mutant:` lines in READY).
6. Dangerous-path run under `env -i`: the built binary against a throwaway project and a plain folder.
7. `cargo fmt --all`, targeted nextest runs, one bare `check affected --card` at READY.

## Risks

- R-1: A cause that is not the call itself (a host that changes its input form) makes every `Bash`, `Write`, `Edit` and `NotebookEdit` call of a MoAI project denied once the hooks are installed; the model then cannot repair it with those tools. The reason asks for the user to be told, and the user can disable or remove the hook entry. This is the price of failing closed.
- R-2: The parser is stricter than a JavaScript host in two places: a lone surrogate escape and nesting past its depth limit read as invalid JSON, so such a call is denied. AC-1 fixes the behaviour; whether Claude Code ever sends either is unverified.
- R-3: A hook process that cannot read its working folder, panics or times out still lets the call run (documentation lines 809 and 908). Named gap, one line in `docs/later.md`.
- R-4: A person who runs the hook name by hand in a terminal (no piped input) sees a deny object. It judges nothing and writes one receipt line.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | All four read failures deny for both `PreToolUse` names, JSON that is not an object included. | The write boundary SPEC's R-10 and its `docs/later.md` line name all of them; a call that cannot be judged is refused; the order was the stricter direction only. The order's wording names JSON and UTF-8; the non-object case is this SPEC's addition. | Remove the non-object arm. |
| D-2 | This replaces the empty-stdout answer of the entry skeleton (SPEC-20261123-hook-entry-skeleton REQ-5 and its AC-4) for the two `PreToolUse` names and widens the write boundary SPEC's REQ-13 from the over-limit input of one name to every read failure of both. | Documentation line 216: silence lets the call run. | Print nothing again for these inputs. |
| D-3 | No new class word: `unreadable` for `bash-boundary` (its existing class for text it cannot read), `protected-path` with a reason of its own for `write-boundary`. | The write boundary SPEC's D-2 keeps three write classes and its D-18 put the over-limit input in `protected-path`; the Bash hook already has `unreadable`. | A fourth write class by a SPEC revision. |
| D-4 | The write denial is built in `src/hook_write_boundary.rs`, the Bash denial in `src/hook_entry.rs`. | The Bash module's class type is private and the file needs no change. | Move the Bash text into its module. |
| D-5 | A deny object with exit 0, not exit code 2. | The entry exits 0 on every path (entry skeleton) and the write boundary SPEC's REQ-13 already denies this way; the documentation allows both. | Exit 2 with the reason on stderr. |
| D-6 | The deny is printed even when the receipt cannot be written. | Same rule as the write boundary SPEC's REQ-13: the record never decides the guard. | None. |

## Result

구현 커밋은 `e65f4f588c` 이고 시험을 먼저 쓴 커밋은 `12ee94e800` 이에요(`moai tdd red` 는 exit 100 의 `test-failure`, `moai tdd green` 은 exit 0). 합친 tip 은 `f48ea9d815` 이에요. 구현 커밋 뒤의 커밋은 이 SPEC 파일만 바꿔요. 개정 2 의 승인은 카드 노트 10395 예요. 확인은 2026-11-26(UTC) macOS 에서 `e65f4f588c` 의 코드로 돌렸고, cargo 명령 앞에는 `moai heavy-slot-acquire --name ivan --max 2 --` 를 붙였어요.

- AC-1: met — verify 명령 exit 0, 7개 통과(두 이름의 빈 입력·일반 글·짝 없는 대리 문자·깊은 중첩·배열·스칼라·UTF-8 아님, 원인별 문구, 환경 변수를 비운 실행).
- AC-2: met — verify 명령 exit 0, 3개 통과(4194305 바이트 거절, 4194304 바이트는 진입점 거절 없음, `write-boundary` 의 한도 초과 문구는 그대로).
- AC-3: met — verify 명령 exit 0, 4개 통과(영수증 한 줄, 등록 폴더 없음·링크·가득 찬 영수증에서도 거절과 stderr 한 줄).
- AC-4: met — verify 명령 exit 0, 2개 통과(MoAI 프로젝트가 아닌 폴더와 `CLAUDE_PROJECT_DIR` 이 그런 폴더일 때 두 스트림 모두 비고 파일 변화 없음).
- AC-5: met — verify 명령 exit 0, 3개 통과(`session-context` 는 거절 대상 입력에 침묵, 정확히 한도인 객체에는 판정 없음, 다른 이벤트와 `tool_input` 없는 객체는 침묵).
- AC-6: met — verify 명령 exit 0, 다섯 대상 63개 통과.
- AC-7: met — verify 명령 exit 0, 1개 통과.
- AC-8: met — `moai spec check SPEC-20261126-hook-unreadable-input` 오류 0, 경고 0.

위험 경로 확인(Plan 6): `env -i` 셸 실행 대신 Rust 통합 시험 `unreadable_ac1_a_cleared_environment_denies_in_a_project_and_is_silent_outside` 로 했어요(결정 10013). 빌드한 바이너리를 환경을 비우고 임시 HOME 으로 실행해, 임시 MoAI 프로젝트에서는 거절하고 일반 폴더에서는 침묵하며 아무것도 쓰지 않아요.

변이 아홉 개(거절 객체 빼기, 한도 비교 바꾸기, 프로젝트 확인 빼기, `CLAUDE_PROJECT_DIR` 무시, `session-context` 도 거절, 다른 이벤트 거절, 판정 없는 객체 거절, 영수증이 없으면 거절하지 않기, 객체 아님 문구를 JSON 아님 문구로 바꾸기)는 모두 `hook_entry_cli` 를 exit 100 으로 실패시켰고, 새 시험 19개가 각각 하나 이상의 변이에 실패했어요.

문서: `docs/later.md` 에서 읽지 못한 입력 줄을 지우고 R-3 줄을 더했어요.

돌리지 않은 것: Windows 와 Linux. 실제 Claude Code 가 이런 입력을 보내는지는 Sources 에 적은 대로 미확인이에요. `moai check affected` 는 이 커밋 뒤의 head 에서 돌리고 결과 줄은 병합 준비 노트에 적어요.
