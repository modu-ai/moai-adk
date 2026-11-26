---
id: SPEC-20261126-hook-write-boundary
title: "moai hook write-boundary: PreToolUse file-write boundary check (H3) and hook conformance tests"
status: draft
revision: 2
created: 2026-11-26
---

# moai hook write-boundary: PreToolUse file-write boundary check (H3) and hook conformance tests

## Intent

새 MoAI 훅 설계의 단계 1-d, H3 예요 (최종 계획서 `~/.moai/reports/moai-adk/hooks-rules-final-plan-2026-11-23.md` §4 H3, §7 1-d). 카드 the hooks redesign card 의 첫 번째 기준("각 훅 이벤트가 공식 문서대로 동작하고, 자동 허용·자동 완료가 없고, WorktreeCreate/Remove 는 기본으로 설치되지 않고, 훅이 꺼져도 필수 관문은 거절한다")을 이 조각이 닫아요. 두 가지를 해요.

1. `moai hook write-boundary` 의 처리기를 채워요. Claude Code 가 `Write`, `Edit`, `NotebookEdit` 를 부르기 전에 쓸 경로와 쓸 내용을 읽고, 네 가지 경계를 넘으면 `permissionDecision: "deny"` 를 내요. 지금 이 자리는 아무것도 하지 않는 빈 함수예요 (`src/hook_entry.rs` 의 `write_boundary`).
2. 세 새 항목(`session-context`, `bash-boundary`, `write-boundary`)이 공식 문서의 입력·출력 계약대로 동작하는지 보는 적합성 시험을 만들어요: 이벤트 이름이 다르면 판단하지 않고, 어떤 출력에도 `allow`·`ask` 가 없고, 어떤 훅도 원장을 바꾸지 않고, 설치 기본 목록에 WorktreeCreate·WorktreeRemove·Stop 이 없고, 훅을 전부 끈 일회용 프로젝트에서도 필수 관문이 거절해요.

이 조각은 처리기와 시험만 해요. 옛 처리기(`src/hook_guards.rs`, `src/hook_engine.rs`)의 판정 코드는 다시 쓰지 않고 읽기만 했어요(사용자 지시 msg_3dc06e21df4c, 계획서 §2). 다만 옛 쓰기 보호가 지금 거절하는 종류(보호 경로, 비밀 내용)는 하나도 빠뜨리지 않아요. 단계 3(the retirement of the old handlers)이 옛 처리기를 걷어낼 때 보호가 약해지면 안 되기 때문이에요(리더 지시: 기존 보호 훅을 약하게 하지 않기).

### Sources

- 공식 hooks 문서 `https://code.claude.com/docs/en/hooks.md` 4,452줄 전체: 2026-11-25T20:48~20:50Z, 입력 표 재수신 2026-11-25T21:36:17Z (sha256 앞 16자 `d16003851f60562e`). 카드 the hooks redesign card 노트 `note:310:efc3646a9f371dcfa0a81d504c5889e5` 에 확인 결과가 있어요. 이 SPEC 이 쓰는 사실: `Write` 입력은 `file_path`(절대 경로)와 `content`, `Edit` 입력은 `file_path`, `old_string`, `new_string`, `replace_all`; 파일 도구의 `file_path` 는 훅이 받기 전에 `~` 와 상대 경로가 펼쳐져요; `permissionDecision: "deny"` 는 `bypassPermissions` 에서도 도구를 막아요; 매처 `Edit.*` 는 `NotebookEdit` 에도 맞아요 (정규식 부분 일치).
- `NotebookEdit` 의 입력 키는 훅 문서에 없어요. 이 세션의 도구 스키마(ToolSearch `select:NotebookEdit`, 2026-11-26)가 `notebook_path`(절대 경로, 필수)와 `new_source`(필수), `cell_id`, `cell_type`, `edit_mode` 를 정의해요. 훅 입력의 `tool_input` 은 도구의 인자 그대로라는 문서 서술에 기대어 이 키들을 쓰되, 실제 훅 입력으로는 확인하지 못했어요(미확인, R-3).
- 계획서 §4 H3 (matcher, 입력, 출력, fail-open, 타임아웃 5 s, 목표 p50 ≤ 20 ms · p90 ≤ 40 ms), §7 1-d, §6 비밀 행.
- 저장소 지도 조사(2026-11-25T20:49~20:58Z, 읽기 전용): H3 는 `src/hook_entry.rs:186-189` 의 빈 함수, 옛 쓰기 보호는 `src/hook_guards.rs` (`write_guard`, `is_protected_path`, `carries_secret`, `TOKEN_SHAPES`), 카드 예약 확인은 `moai card check` 이고 훅 호출자가 없음.

### 이미 있는 것

- `src/hook_entry.rs` 는 `moai hook write-boundary` 를 이미 받아요 (SPEC-20261123-hook-entry-skeleton): 표준 입력 한 번 읽기, 비MoAI 폴더는 출력 0 · exit 0, 입력 오류는 영수증 한 줄. 이 조각은 그 빈 함수 자리에 호출 한 줄을 더해요.
- 형제 처리기 H2 (`src/hook_bash_boundary.rs`)가 "세션 체크아웃"을 `cwd` 에서 가장 가까운 `.git` 항목으로 정하고, 프로세스 없이 `.git` 파일과 `commondir`, `worktrees/*/gitdir` 를 읽어요. H3 는 같은 정의를 써요(D-3). 그 파일의 도우미는 비공개이고 그 파일은 the boundary-detail change 이 고치는 중이라, H3 는 같은 읽기 규칙을 새 모듈에 자기 몫으로 두고 합치는 일은 the boundary-detail change 착지 뒤의 후속으로 미뤄요(D-12).
- 카드 예약 겹침 거절은 사용자가 승인한 `SPEC-20261125-card-no-file-locks` 로 없어졌어요. 그래서 계획서 H3 의 "다른 카드가 예약한 경로" 조항은 이 SPEC 에 넣지 않아요(D-10).

## Non-goals

- 훅 설치, settings 항목, `onFailure`, 타임아웃 등록, 레인 워크트리로의 하네스 전파, 항목 단위 소유 (단계 1-e). 이 조각은 settings 를 읽지도 쓰지도 않아요. the hooks redesign card 의 둘째 기준(the earlier hook-wiring work 의 남은 일: 실제 훅 연결과 MoAI 가 아닌 저장소의 허용 경로)도 1-e 와 함께 다음 SPEC 이에요(D-8).
- SPEC 이 없을 때의 안내와 거절. 이 조각은 둘 다 하지 않고 옛 쓰기 보호의 SPEC 먼저 규칙이 계속 맡아요. 따라서 이 SPEC 이 닿아도 그 옛 거절은 걷어낼 수 없어요(D-9).
- 커밋에 비밀이 들어가는 것을 막는 Rust 사전 커밋 검사(계획서 §6). 이 조각은 만들지 않아요. 비밀 내용과 보호 경로 쓰기는 훅이 꺼지면 아무것도 거절하지 않는 부류이고, 이 빈틈은 이름을 붙여 기록해요(R-2).
- Bash 로 파일을 쓰는 일(리다이렉션, `tee`, `cp`, `mv`), `MultiEdit` 와 그 밖의 도구. 훅 문서에 `MultiEdit` 입력이 없어서 그 도구는 판단하지 않고 조용해요.
- 훅 이벤트 네 개 이상 확장, 새 이벤트 등록, PowerShell, Windows 확인(검증은 macOS 에서만 해요).
- 진짜 Claude Code 세션에서 항목이 로드되는지 확인하는 일(설치가 1-e 라서 이 조각은 `not_run` 으로 남겨요).
- `src/hook_bash_boundary.rs`, `tests/hook_bash_boundary_cli.rs` 수정(the boundary-detail change 이 착지하기 전).

## Requirements

<!-- This SPEC adds a check that blocks a tool call: a violation is a deny, never an ask or an allow; the final enforcement of paths and secrets stays with the ledger, the landing check and the git guards. -->
- REQ-1: When `moai hook write-boundary` runs in a MoAI project with a call whose `tool_name` is `Write` or `Edit` and whose `tool_input.file_path` is a string, or whose `tool_name` is `NotebookEdit` and whose `tool_input.notebook_path` is a string, the write boundary handler shall judge that path and its content against the checks of REQ-4 to REQ-6 · failure: none — a call whose `tool_name` is missing or is any other name, or whose path is missing or is not a string, gets no decision.
- REQ-2: When the call breaks a check, the write boundary handler shall print exactly one JSON object `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":<text>}}` on stdout and exit 0, and a call that breaks none prints nothing on stdout and stderr and exits 0 · failure: none — the handler never prints `ask`, `allow`, `updatedInput` or `additionalContext`.
- REQ-3: The write boundary handler shall name the session checkout as the checkout that holds the hook input `cwd` (the nearest ancestor with a `.git` entry, read from the file system without starting a process), resolve the target path after `..` and symbolic links and judge a path of a file that does not exist yet by its nearest existing ancestor · failure: a `cwd` that cannot be resolved leaves no session checkout, so only REQ-5 is skipped, and a `.git` file that cannot be parsed names no checkout, so the search goes on with the parent folder (D-15).
- REQ-4: IF the target path, as written or as resolved, has a `.git` component, a credential folder component (`.ssh`, `.aws`, `.azure`, `.kube`, `.gnupg`, or `.config/gcloud`) or a private-key file name (`id_rsa`, `id_ed25519`, `id_ecdsa`, `id_dsa`) as its last component, with names compared as a file system that ignores letter case spells them (D-13), or its symbolic links cannot be followed (D-14), THEN the write boundary handler shall deny the call with the class `protected-path`.
- REQ-5: IF the target path lies in a checkout of the session's repository other than the session checkout (the primary checkout or a linked worktree, a registered worktree folder that lost its `.git` file included, D-16), THEN the write boundary handler shall deny the call with the class `other-checkout`, and a path in no checkout of that repository (a temporary folder, `scratchpad_dir`, the home state folder, the user's Claude folder) is not this check's.
- REQ-6: IF the content of the call (`content` of `Write`, `new_string` of `Edit`, `new_source` of `NotebookEdit`) holds a private-key block header or an access token of the shape table of D-4 (prefix, minimum length, character set and word-start rule), THEN the write boundary handler shall deny the call with the class `secret-content`, and a prefix without the minimum length after it, in prose or in a test fixture of that shorter size, is allowed.
- REQ-7: When the reason text is written, the write boundary handler shall begin it with `moai-boundary/<class>:`, name the violated rule and one allowed way, keep it under 400 bytes, and carry no byte of the path or the content · failure: none — a reason that cannot be built still denies with the class word alone.
- REQ-8: When the entry reads a call whose `hook_event_name` is present and differs from the event the registered name belongs to (`PreToolUse` for `bash-boundary` and `write-boundary`, `SessionStart` for `session-context`), the entry shall judge nothing, print nothing and write one receipt line with the reason `event_mismatch`, and a call with no `hook_event_name` is judged as before · failure: none — a receipt that cannot be written is the existing one stderr line.
- REQ-9: The write boundary handler shall make no network call, start no process, read no environment variable beyond what the entry reads (`CLAUDE_PROJECT_DIR`, `HOME`), write no file and not depend on `permission_mode` for its decision.
- REQ-10: When the conformance test runs the three registered names over its fixture set, the entry shall print no `allow` or `ask` decision and leave every card state, note and reservation of the project unchanged · failure: none — a fixture that cannot be run fails the test.
- REQ-11: The hook set generator of `moai hooks-plan` shall name no `WorktreeCreate`, `WorktreeRemove` or `Stop` event in its default plan.
- REQ-12: When Claude Code hooks are disabled in a throwaway project (`disableAllHooks` true in its settings, and no hook entry registered), the guards that do not read a hook (the reference-transaction guard of the integration branch, `moai card check` scope, `moai trailer-check`) shall still refuse their violation · failure: none — the two classes that only a hook refuses (REQ-4, REQ-6) are the named gap of R-2.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4, REQ-5, REQ-6): the table of allowed and denied cases is fixed as tests, at least 8 allowed and 8 denied cases for each of the three checks and each of the three tools, every case running the real binary in a temporary project and HOME and comparing the class and the decision · verify: `cargo nextest run --test hook_write_boundary_cli -E 'test(/^ac1_/)'`
- AC-2 (REQ-2): outside a MoAI project the hook prints nothing on either stream and exits 0 for every case of the table, and in a MoAI project a denied case prints one JSON object with exactly the three fields of REQ-2, never `ask` or `allow`, and an allowed case prints nothing · verify: `cargo nextest run --test hook_write_boundary_cli -E 'test(/^ac2_/)'`
- AC-3 (REQ-1): input with another `tool_name`, no `tool_name`, a missing, numeric, null or array path, and an empty string path each exit 0 with nothing printed, and `NotebookEdit` judged by `file_path` alone is not judged · verify: `cargo nextest run --test hook_write_boundary_cli -E 'test(/^ac3_/)'`
- AC-4 (REQ-3, REQ-5): a path inside the session checkout, a path in a linked worktree session's own checkout, a new file in a folder that does not exist yet, a path in a temporary folder, in `scratchpad_dir`, in the home state folder and in the user's Claude folder are allowed, while the primary checkout seen from a linked worktree, a sibling linked worktree, a registered worktree folder without its `.git` file, a link that leads into another checkout and a path reached through `..` are denied, while a session in a worktree nested inside the primary writes its own files and a separate repository nested inside the primary is not another checkout · verify: `cargo nextest run --test hook_write_boundary_cli -E 'test(/^ac4_/)'`
- AC-5 (REQ-4): every protected place of REQ-4 is denied in each spelling (written, through `..`, through a link, as the new file of a missing folder), and a folder or file that only contains the same letters (`.gitignore`, `.github/x`, `my.ssh.md`, `id_rsa.pub`) is allowed, the same places in another letter case (the long s, the Kelvin sign and the sharp s included) are denied, a path with more than 32 links or a link loop is denied, and a broken `.git` file does not hide the checkout around it · verify: `cargo nextest run --test hook_write_boundary_cli -E 'test(/^ac5_/)'`
- AC-6 (REQ-6): each token shape of the table of D-4 at its minimum length, and a private-key block header, are denied in `content`, `new_string` and `new_source`, the same prefix one character shorter, and prose that only names a prefix, are allowed, and the word-start rule of the OpenAI-style prefix holds (an ordinary hyphenated word that ends in those two letters is allowed) · verify: `cargo nextest run --test hook_write_boundary_cli -E 'test(/^ac6_/)'`
- AC-7 (REQ-7): every denial reason starts with `moai-boundary/` and one of the three class words, is under 400 bytes and carries no byte of the path or the content (a marker word in both never appears in the output) · verify: `cargo nextest run --test hook_write_boundary_cli -E 'test(/^ac7_/)'`
- AC-8 (REQ-8): a `hook_event_name` that differs from the name's event prints nothing and writes one `event_mismatch` receipt line for each of the three names, a matching name and an absent name are judged, and the receipt holds no input byte · verify: `cargo nextest run --test hook_entry_cli -E 'test(/^ac8_/)'`
- AC-9 (REQ-9): `src/hook_write_boundary.rs` names no process-start, network, environment-read or file-write word, no file is created or changed in the project or HOME by any case, the answer is the same for every `permission_mode` value, and the diff touches no settings file, `src/hook_bash_boundary.rs`, `src/hook_guards.rs`, `src/hook_engine.rs` or `src/main.rs` · verify: `cargo nextest run --test hook_write_boundary_cli -E 'test(/^ac9_/)'`
- AC-10 (REQ-10): the conformance test runs the three names over the fixture set (allowed, denied, malformed, wrong event, wrong tool) and finds no `allow` or `ask` decision in any stdout, and the project's ledger, notes and reservations hold the same bytes before and after · verify: `cargo nextest run --test hook_conformance_cli -E 'test(/^ac10_/)'`
- AC-11 (REQ-11): the plan and the written default settings of `moai hooks-plan` hold none of the three events, and no settings file of the temporary project names `write-boundary`, `bash-boundary` or `session-context` after the run · verify: `cargo nextest run --test hook_conformance_cli -E 'test(/^ac11_/)'`
- AC-12 (REQ-12): in a throwaway project with hooks disabled, a hand move of the integration branch is refused by the installed reference-transaction guard, a path outside the card's reserved scope answers `allowed: false`, a commit message without the trailers fails `moai trailer-check`, and the test prints the two hook-only classes as the named gap · verify: `cargo nextest run --test hook_conformance_cli -E 'test(/^ac12_/)'`
- AC-13 (REQ-2): the measuring test (D-11) runs the real binary 200 times for an allowed and a denied call in a temporary project, prints p50 and p90 in milliseconds with the build profile and the machine load, and fails only when p90 is over 10 times the target · verify: `cargo nextest run --test hook_write_boundary_cli -E 'test(/^ac13_/)'`
- AC-14 (REQ-2, REQ-8): the entry skeleton, session-start, latency and dispatch tests still pass, so the other registered names and the non-MoAI answer are unchanged · verify: `cargo nextest run --test hook_entry_cli --test hook_latency_cli --test hook_dispatch_cli --test hook_session_start_cli`
- AC-15 (REQ-1…REQ-12): the SPEC itself passes the spec check · verify: `moai spec check SPEC-20261126-hook-write-boundary`

## Scope

- write: `.moai/specs/SPEC-20261126-hook-write-boundary/**`, `src/hook_write_boundary.rs` (new), `src/hook_entry.rs` (the `write_boundary` call and the event table of REQ-8), `src/lib.rs` (one `pub mod` line, through a line lease), `tests/hook_write_boundary_cli.rs` (new), `tests/hook_conformance_cli.rs` (new), `tests/hook_entry_cli.rs` (the silent write-boundary assertion and the event cases), `tests/hook_session_start_cli.rs` (the two PreToolUse names get a PreToolUse input, REQ-8), `tools/repo_scans.json` (one scan entry, through a line lease), `tests/repo_scans_table.rs` (only an exemption row when a new test walks temporary folders), `tests/areas.toml` (only when a new test-file word needs a line; `hook` is already listed), `docs/later.md` (two lines, D-12 and R-2).
- read-only: `src/hook_bash_boundary.rs`, `src/hook_guards.rs`, `src/hook_engine.rs`, `src/hooks.rs`, `src/hook_latency.rs`, `src/card/guard.rs`, `src/card/work.rs`, `.moai/project.yaml`, the hooks plan and the official hooks documentation.
- forbidden: any settings file (`.claude/**` of the home and of the project), the installed hooks of this clone, `src/hook_bash_boundary.rs` and `tests/hook_bash_boundary_cli.rs` (the boundary-detail change), `src/hook_guards.rs`, `src/hook_engine.rs`, `src/claim_verify.rs`, `src/main.rs`, `src/help.rs`, the real HOME and its `.moai/projects` folder, and any network call. Tests use temporary HOMEs, folders and repositories.

## Interview

- Q1: Is this slice (H3 and the conformance tests) the right size for card the hooks redesign card, with step 1-e (item-level harness ownership, lane worktree propagation) and the card's second criterion (real hook wiring S3b, the non-MoAI allow path) moved to a following SPEC? → A: pending — the leader decides, because splitting a card's criteria changes its scope; the default is yes, since step 1-e edits the settings generator and overlaps the retirement of the old handlers, the lane-worktree command change, the worktree-settings check change and the project-flow change, which read the same worktree settings.
- Q2: Does H3 refuse the secret and protected-path classes that the live old write guard refuses, so that retiring that guard (the retirement of the old handlers) keeps them? → A: yes, as written in REQ-4 and REQ-6, by the reviewer's reading of the leader's order not to weaken an existing protection hook; the SPEC-first refusal is the one class that stays with the old guard (D-9).
- Q3: Is a pre-commit secret scan part of this slice? → A: pending — the default is no (named gap R-2, a following card), because it is a git hook installer change with its own SPEC.

## Plan

1. (this slice) SPEC, status draft; the approval is not written by a lane. Until it is approved nothing below starts.
2. Tests first (RED): `tests/hook_write_boundary_cli.rs` with the case table and the fixtures (temporary repository, linked worktrees made by `git worktree add`, link-in, `..` and missing-folder shapes) against the empty function; `tests/hook_conformance_cli.rs` and the event cases of `tests/hook_entry_cli.rs`.
3. `src/hook_write_boundary.rs`: the checkout reader, the path resolver, the protected-path and other-checkout checks, the token table and the reason writer; `src/lib.rs` gets one line, `src/hook_entry.rs` the call and the event table.
4. Guard mutants: remove each guard in turn and show the test that fails (`guard-mutant:` lines in READY, one per new guard test).
5. Dangerous-path run in a throwaway: the real binary against a throwaway project, a plain folder at a registered path, a link into another checkout and a missing deep folder.
6. Measurement run: p50 and p90 of the real binary for an allowed and a denied call (D-11).
7. Targeted tests with nextest `--test` names, `heavy-slot --max 4`, new targets under nextest before the receipt (HOME-GUARD), one bare `check affected` at READY.
8. After the boundary-detail change lands: merge the tip once; this slice touches none of the boundary-detail change's files.

## Risks

- R-1: Over-blocking. A legitimate write into another checkout of the repository is denied (D-3). The reason names the allowed way (write in the session's own worktree). Relaxing a rule is a decision of this SPEC, never a silent special case.
- R-2: Two classes are refused only by a hook and by nothing else: a secret in file content reaching a commit (the plan's Rust pre-commit scan does not exist) and a write into a protected path. With hooks disabled, both pass. Named gap; criterion 1 of the hooks redesign card is met for the guards that do not read a hook and not met for these two until the scan card lands.
- R-3: The `NotebookEdit` key `notebook_path` comes from the tool schema, not from a real hook input. If the real key differs, that tool is judged by nothing (fail-open, REQ-1) and the real-load check of 1-e must catch it. A `MultiEdit` call is not judged at all.
- R-4: `cwd` follows `cd` between calls (official documentation). The hook sees only the current `cwd`; the ledger's scope check and the landing check remain the second line.
- R-5: The token table is a copy of the shape data of the old guard. Two copies can drift until the retirement of the old handlers removes the old one (D-12).
- R-6: Whether a PreToolUse deny holds in `bypassPermissions`, `dontAsk` and `auto` modes is documented for `bypassPermissions` only (Sources); D-5 makes the handler mode-independent, so the handler needs no change either way.
- R-7: Latency depends on machine load. The numbers are reported with the load; a hard bound would be flaky (D-11).
- R-8: A registry entry of a removed worktree that was not pruned makes writes under that old folder count as another checkout (D-16) until `git worktree prune`. The handler reads the registry on every judged call, so its cost grows with the number of worktrees.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The judgment lives in a library module `src/hook_write_boundary.rs` (`pub mod` in `src/lib.rs`); the entry's `write_boundary` calls it in one line. | H2 does the same; a library module is testable by name and keeps the merge with the boundary-detail change small. | Inline it in `src/hook_entry.rs`. |
| D-2 | Three classes only (`protected-path`, `other-checkout`, `secret-content`); the handler denies and never asks, allows or adds context. | Plan §2 incident #8 (no `ask`); documentation: silence does not approve a call, an allow cannot lift deny and ask rules. | Add a class by a SPEC revision. |
| D-3 | The session checkout is the nearest `.git` ancestor of `cwd`; the other checkouts are the primary checkout and every registered linked worktree except it. A path outside every checkout of the repository is allowed. | Same definition as H2 rule 2, so the two boundaries never disagree; the plan's "outside the session worktree" would refuse reports in `~/.moai/reports`, memory files and the scratchpad that a session must write. | Narrow or widen the allowed roots by a revision. |
| D-4 | Secret content is judged by a table of token shapes (prefix, minimum length, character set, word-start rule: GitHub, GitLab, Slack, AWS, Google and OpenAI/Anthropic-style keys) and a private-key block header; the table is the data of the old guard's `TOKEN_SHAPES`, copied as data. `crate::secret_shaped` is not used. | `crate::secret_shaped` refuses `token=` and `password=` anywhere, which would refuse ordinary code; the old table has minimum lengths so that prose naming a prefix passes (SPEC-20261120-secret-token-shapes, SPEC-20261120-secret-project-keys). | Switch to a shared table once the retirement of the old handlers retires the old guard. |
| D-5 | The decision does not depend on `permission_mode`, and the handler prints nothing else on stdout. | Documentation: a deny holds in `bypassPermissions`; a mode-independent handler needs no change if another mode differs. | None. |
| D-6 | `NotebookEdit` is judged by `notebook_path` and `new_source`, and a call is judged only when `tool_name` is one of the three, not by the matcher. | Tool schema read 2026-11-26; documentation: `Edit.*` also matches `NotebookEdit`, so a handler that trusts the matcher misjudges. | Add a tool by a revision. |
| D-7 | The event-name check lives in `src/hook_entry.rs`, one table for the three names, not in each module. | Nothing reads `hook_event_name` today; one table covers H1, H2 and H3 without touching the boundary-detail change's file. | Remove the table. |
| D-8 | Step 1-e and the card's second criterion are not in this SPEC. | They edit the settings generator (`src/hooks.rs`, `src/claude.rs`) and overlap the retirement of the old handlers, the lane-worktree command change, the worktree-settings check change and the project-flow change; one SPEC for both would be too large to audit. | A following SPEC under the same card or a new card, the leader's pick. |
| D-9 | SPEC-first stays a refusal of the old write guard; H3 neither refuses nor advises it. | The research found no landing-side SPEC check, so retiring the old refusal would drop the rule; the `additionalContext` form is not verified for PreToolUse. | Add an advisory by a revision after the landing check exists. |
| D-10 | H3 does not read the card ledger. | `SPEC-20261125-card-no-file-locks` (user-approved) removed the cross-card overlap refusal; `moai card check` stays the scope predicate. | None. |
| D-11 | Latency is measured and reported with the load, not asserted tightly: targets p50 ≤ 20 ms, p90 ≤ 40 ms, failing only past 10 times. | Machine load of 80 to 90 was observed on 2026-11-23; same rule as H2 (D-10 there). | A bound after the load is steady. |
| D-12 | The checkout reader is written again in the new module; one line in `docs/later.md` names merging it with H2's after the boundary-detail change lands, and the token table with the old guard's after the retirement of the old handlers. | `src/hook_bash_boundary.rs` is the boundary-detail change's file until it lands. | Merge now. |
| D-13 | Protected names are compared as folded spellings (ASCII lower case, plus the long s, the Kelvin sign, the sharp s and the `st` ligatures), and the resolved path and every checkout root use the spelling the disk holds. | Revision 2: the review of the committed code showed that `.GIT/config` or `ID_RSA` passed on the case-ignoring file system of a Mac. On a file system that honors case this refuses a few more names. | Compare exact bytes again. |
| D-14 | A path whose links cannot be followed (more than 32, a loop or an unreadable link) is denied in the class `protected-path` with a reason of its own, not skipped. | Revision 2: skipping left a path judged by no check; the three class words stay (D-2). | Skip such a path (fail-open) and name it in `docs/later.md`. |
| D-15 | A `.git` file that cannot be read or parsed names no checkout and the search goes on with the parent folder. | Revision 2: stopping there let a broken file hide the checkout around it. | Stop at the broken file. |
| D-16 | The other checkouts include the registered worktrees of the common folder (`worktrees/*/gitdir`); the deepest root holding the path decides. | Revision 2: D-3 names every registered linked worktree; walking up alone allowed a registered folder that lost its `.git` file. Same registry as H2. | Walk up only. |

## Result

(written at READY)
