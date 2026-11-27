---
id: SPEC-20261121-self-improvement-loop
title: "Self-improvement loop: `moai improve` proposes small gated changes to project skills, agents and CLAUDE.local.md from observed signals; the first change adds the read-only `improve-signals`"
status: approved
revision: 2
created: 2026-11-21
approval:
  by: user
  at: 2026-11-27
  basis: "사용자가 직접 골랐어요(2026-11-27, 리더 세션의 질문 도구; 리더 노트 10683, 2026-11-27T06:28:45Z, 레인은 그 노트를 인용해 옮겨 적었어요): 개정 1 을 한 SPEC 으로 승인(D11), D1 개인 파일 쓰기 허용(관리 블록 안만, 미리보기와 선택 뒤), D4 평평한 이름(improve-signals·improve-plan·improve-eval·improve-apply·improve-revert, 스킬 입구 /moai improve), D6 어림 세기 받아들임(출력에 어림값 표시). 개정 2: 사용자가 직접 \"개정 2 승인 (Recommended)\"을 골랐어요(2026-11-27, 리더 세션의 질문 도구; 리더 노트 10814, 2026-11-27T12:10:26Z, 레인은 그 노트를 인용해 옮겨 적었어요): 요구 11줄(REQ-26~REQ-36)과 기준 9줄(AC-10~AC-18)을 더하고 결정 본문(Decisions)은 바꾸지 않음"
  digest: "sha256:c584a562d011c102b02585269cfab91943644c4b648ccfbe9d39076954a0cc55"
---

# 자기 개선 루프 (`moai improve`)

## Interview

- Q1: 무엇을 만드나요? → A: 사용자 지시(2026-11-20) — 재귀적 자기 개선(RIS) 논문과 모범 사례를 모아, moai-adk 개발(로컬)과 사용자 설치(프로젝트 스킬·에이전트·`CLAUDE.md`·`CLAUDE.local.md`)에서 AI 가 스스로 개선하는 기능을 추가해요.
- Q2: 근거는 어디에 있나요? → A: 리더의 보고서(저장소 밖) `self-improvement-research-2026-11-21.md`(기준 커밋 `f2eb8c61e3`, 2026-11-20T21:23Z, 54,221 바이트, sha256 `81158291de773da147b8cc7e1107a5f824185206f583ae0b418df80ab2f39403`). 이 개정은 보고서를 다시 읽지 않았고 체크섬만 2026-11-27T04:54Z 에 계산했어요. 인용은 2026-11-20 초안이 읽은 것이에요. 저장소 밖 파일이라 쓰는 근거를 Decisions 끝의 "근거"에 줄여 적었어요.
- Q3: 열린 결정은요? → A: 보고서 §6.3 의 셋은 리더가 2026-11-20T21:30Z 에 정했어요(D1~D3). D4~D12 는 이 초안의 제안이고 승인은 SPEC 이 병합될 때 사용자·리더가 골라요. 이 문서는 승인을 스스로 적지 않아요.
- Q4: 사용자에게 남은 질문은요? → A: 있어요. 선택지와 기본값을 카드의 `LANE-BLOCKED kind=decision` 노트 한 건에 모았고(SPEC-READY 노트), 리더가 한 번에 여쭤요. 기본값으로 진행하면 아래 Decisions 그대로예요.
- Q5: 2026-11-21 초안 뒤에 무엇이 바뀌었나요? → A: (1) Codex·OMP 철회(사용자 결정 2026-11-21, `SPEC-20261121-remove-codex-support`, `SPEC-20261121-remove-codex-project-layer`): 호스트는 Claude Code 하나예요. (2) 프로젝트 지침 파일은 `CLAUDE.md`(팀)와 `CLAUDE.local.md`(개인)이고 `AGENTS.md`·`AGENTS.local.md` 는 쓰지 않아요(사용자 결정 2026-11-22, `CLAUDE.md` 를 정본으로). (3) 테스트는 Rust 만 써요(`cargo nextest`). (4) NeXT-Flow: 레인은 커밋만 하고 독립 리뷰 뒤 리더가 랜딩해요. (5) 공유 파일(`src/help.rs`, `tests/fixtures/help/*.txt` 등)은 줄 임대로만 써요. (6) 운영 레지스트리는 경로 없는 형태예요(`SPEC-20261122-ops-registry-path-free`).

## Intent

AI 코딩 도구가 같은 실수를 되풀이하면 사람이 같은 정정을 반복해요. 이 기능은 **관찰된 반복 문제를 모아 → 프로젝트 지침·도구의 작은 변경을 제안하고 → 독립 평가로 걸러 → 사용자가 고른 것만 적용·기록하고 → 되돌릴 수 있게** 하는 루프예요. 보고서의 결론 그대로 살아남는 원칙은 하나예요: **제안 → 독립 평가 → 수용(또는 거절) → 기록**(Darwin Gödel Machine, ADAS, SICA, OPRO 가 같은 모양). MoAI 는 이미 이 뼈대를 갖고 있어요(plan digest → 사용자 선택 → apply → receipt, "구현자는 감사하지 않는다"). 새 쓰기 엔진은 만들지 않고 기존 `generation-plan/apply`(프로젝트 스킬·에이전트)와 `agents-plan/apply`(`CLAUDE.md` 관리 구간)를 써요.

가장 큰 위험은 **평가 장치를 속이는 개선**이에요(DGM 이 탐지 표식을 지우고 테스트 통과 로그를 위조, METR·ImpossibleBench 의 테스트 변조). 그래서 평가 세트와 게이트 파일은 개선 루프가 쓸 수 없는 자리에 두고, 지침은 통째로 다시 쓰지 않고 항목 단위 델타 + 결정적(LLM 이 아닌) 병합으로만 고쳐요(ACE: 통째 재작성은 18,282 → 122 토큰, 정확도 66.7 → 57.1). 신호는 **외부에서 관찰된 사실**만 쓰고 에이전트의 주장은 신호가 아니에요(Huang et al.: 외부 피드백 없는 자기 교정은 믿기 어려움).

이 SPEC 은 루프 전체의 설계를 한 번에 정하고(요구 REQ-1~REQ-25), 구현은 조각(slice) 다섯 개로 나눠요(Plan). 조각과 요구와 카드의 대응은 아래 "조각과 카드의 경계" 표예요. **첫 번째로 착지할 변경은 조각 S1 하나**예요: 읽기 전용 명령 `moai improve-signals` — 관찰된 신호를 모아 문제 "부류(class)"로 묶고 횟수를 세어 보여 줘요. 쓰기, 제안, 평가는 S1 에 없어요.

**이력(S1 은 한 번 구현된 적이 있어요)**: 이 SPEC 의 2026-11-21 초안과 조각 S1 구현(`improve_signals.rs`, `improve_signals_cli.rs`, 운영 문서 조각)은 로컬 브랜치 `spec/self-improvement-loop-20261121`(헤드 `a4f4f177ca`, 옛 바탕 `0989b3fc58`)에 있고 NeXT 에는 들어오지 않았어요. 그 구현은 SPEC 승인 전에 쓰였고 그 뒤 NeXT 가 크게 앞서 갔어요. 이번 개정은 새 브랜치 `spec/self-improvement-loop-20261121-r2`(NeXT tip `e98798e41f` 에서 땀)에서 SPEC 초안 파일만 옮겨 지금 사실에 맞춘 것이고, 코드는 옮기지 않았어요. 옛 브랜치는 지우지도 고치지도 않고 승인 뒤 S1 을 다시 붙일 때 참고로만 써요. 그 구현의 결과(Result 의 옛 기록)는 이 브랜치의 증거가 아니에요.

## Non-goals

- S1 에서: 제안 만들기, 어떤 파일이든 쓰기, 평가 실행, 네트워크, 호스트 메모리(Claude 자동 메모리) 읽기, 스킬 본문(`skills/moai/**`)·`agents/*.md` 문구 변경(`improve` 루트 문서는 조각 S4).
- 루프 전체에서(보고서 §3.9·§6.2): 세션 중 자동 적용, 백그라운드 상시 개선, 다른 사람 파일 수정, 외부에서 스킬 가져오기, 모델 가중치·호스트 설정·권한·훅의 자동 변경, 진화 탐색(AlphaEvolve·DGM 식 집단 탐색), 평가 세트 자동 확장, "감사 FAIL 수 줄이기" 같은 감독 신호를 점수로 직접 최적화.
- 팀 공유 `CLAUDE.md` 를 개선 명령이 직접 쓰기(항상 SPEC 경로).
- Codex·OMP 같은 철회된 호스트의 대상 경로(`.agents`, `.codex`, `.omp`)와 `AGENTS.md`·`AGENTS.local.md` 를 대상으로 삼기. 보고서의 Codex 항목은 이 SPEC 이 쓰지 않아요.

## Requirements

요구 REQ-1~REQ-11 은 조각 S1 이고, REQ-12~REQ-25 는 이 SPEC 이 정해 두는 루프 전체의 설계로 뒤 조각이 구현해요. REQ-26~REQ-36 은 조각 S2 가 개정 2 로 더한 줄이에요(`improve-plan` 의 정적 게이트와 쓰기 없음, `improve-note` 의 쓰기).

- REQ-1: When `moai improve-signals <project>` runs, the command shall read only observed facts from four sources — audit-FAIL notes and `LANE-BLOCKED kind=<kind>` notes of the project's card ledger, cards titled `flaky: <test>`, and the project's `improve/notes.jsonl` — and print each fact as a signal grouped into classes with a count, the first and last time and at most five example signals per class.
- REQ-2 (REQ-1): IF the card ledger cannot be read or the notes file is a symbolic link, THEN the command shall refuse with an error and print no classes.
- REQ-3: Where the card ledger or the notes file is absent, the command shall report that source as absent, read nothing from it and exit 0.
- REQ-4: When a class has been observed once, the command shall mark it `observed_once` and propose nothing · failure: none
- REQ-5: When a class has been observed two or more times, the command shall mark it `candidate` and propose nothing · failure: none
- REQ-6: When a note or a line is an agent's own claim (a `LANE-READY`, `LANE-DONE`, `CANDIDATE`, `MERGE-READY` or `MERGED-LOCAL` note, a commit message, a "fixed" report, or any note whose first line matches no source rule), the command shall skip it and count it under `skipped` with the reason `no_rule` · failure: none
- REQ-7: When a signal is built, the command shall give the class `false_claim` only to improve notes of the kind `false_claim` and to no other source · failure: none
- REQ-8: While `moai improve-signals` runs, the command shall write no file, no ledger row and no folder, also when the card ledger or the state folder is absent, and the result says `writes: false`.
- REQ-9: The command shall read no host memory of any tool and no file under the real HOME other than the project's own card ledger, and its result carries `host_memory: "off"`; reading host memory is a later slice behind a per-project opt-in (D3).
- REQ-10: When the command prints text taken from a note or an improve note, the text shall be cut to 120 characters and pass through `redact::redact_outgoing` and `redact::redact_secret_shapes`, and no class key carries a secret-shaped word · failure: none
- REQ-11: When the ledger holds more than 20,000 notes, the command shall read the newest 20,000 notes, cut each body at 8,192 bytes and say `truncated: true` · failure: none
- REQ-12: When a later command proposes a change, it shall touch only grade A (project skills and agents with the `prj-` prefix under `.claude/skills` and `.claude/agents`), grade B (the managed block `moai-lessons` of `CLAUDE.local.md`, D1) or a SPEC draft for `CLAUDE.md` (grade C) · failure: none
- REQ-13 (REQ-12): IF a proposal targets the installed `moai` skill, host settings, permissions, keys, hooks, MCP settings, a test deletion or weakening, or the evaluation set or gate files, THEN the plan shall be refused with the reason and the error kind `refused`.
- REQ-14: When a proposal is built, it shall be one target file with one to three items, each with an `id`, one rule sentence with one reason line, `source` signal ids, the status `trial` and the counters `helped` and `hurt`, merged by deterministic code inside the managed block by item id · failure: none
- REQ-15 (REQ-14): IF the target file would pass its size budget (200 lines for CLAUDE.md-class files, 500 lines and a 1,024-character description for a skill), THEN the proposal shall be refused unless it also retires one old `trial` item (N is accepted, N+1 is refused).
- REQ-16: When a proposal is evaluated, the evaluation shall run the static gate, then replay the triggering case before and after with a new sub-agent that cannot write, then the project's regression set, then the project's own tests where a script or tool changed, and record `accept`, `reject` or `needs_human`, with any LLM opinion kept as advice and never as the gate · failure: none
- REQ-17 (REQ-16): IF a case that passed before the change fails after it, or the triggering case still fails, THEN the verdict shall be `reject`.
- REQ-18 (REQ-16): IF a checker cannot decide, THEN the verdict shall be `needs_human` and never a pass.
- REQ-19: When a proposal is applied, the command shall require the plan digest, a caller-supplied permission reference and a user pick made after a preview (the diff, the signals with counts and dates, the evaluation table and the revert command), write a receipt and one ledger line with the decision record fields, and never edit the team `CLAUDE.md` itself · failure: none
- REQ-20 (REQ-19): IF the digest or the permission reference is missing or stale, THEN the apply shall be refused and nothing is written.
- REQ-21: When a change is reverted, the command shall restore the bytes of the receipt in the reverse order of applying · failure: none
- REQ-22: When a new item is applied, it shall start as `trial` for five sessions and then be proposed as `kept` or `retire`, within the rate limits of one round a day, three proposals a round, three items and twenty added lines a proposal, no second change of one file within seven days (a revert excepted) and twenty-four replays a round, each accepted at N and refused at N+1 · failure: none
- REQ-23: Where the project's `CLAUDE.local.md` holds a managed block `moai-lessons`, the writing code shall change only the lines inside the block markers, and `src/agents.rs` names the block as the one exception to "created once and never modified" · failure: none
- REQ-24 (REQ-16): IF an evaluation file changed since the plan was built, THEN the evaluation shall be `blocked`.
- REQ-25: Where a maintainer adapter or a maintainer evaluation case exists (`tools/improve/maintainer-signals.json`), it shall stay outside the shipped package, and a scan fails when the count of maintainer-only tokens in shipped text rises above the base commit's · failure: none
- REQ-26: When `moai improve-plan` runs, the command shall write no file, no ledger row and no folder, and its result says `writes: false` · failure: none
- REQ-27: When a plan is built, the command shall read the target only as a regular file below the project root and print the target, the sha-256 of the old and of the new bytes, the item lines it would write, the target's generation state (`unchanged` when the bytes equal the manifest's `baseline_sha256`, `conflict` when they differ) and a `plan_digest` · failure: none
- REQ-28 (REQ-27): IF an item rule or reason names a project path that does not exist, or a `moai` command that the binary does not have, THEN the plan shall be refused with the reason `drift`.
- REQ-29 (REQ-27): IF the target is not a regular file below the project root (a link, a path with `..`, a folder), or does not start with front matter that has a `name` and a `description` of at most 1,024 characters, THEN the plan shall be refused with the reason `not_allowed` or `front_matter`.
- REQ-30 (REQ-27): IF an item rule, reason or source id is shaped like a secret (`redact::redact_secret_shapes` changes it), THEN the plan shall be refused with the reason `secret`.
- REQ-31: When the same arguments and the same bytes of the target and the manifest are given, the plan digest shall be the same, whatever the project folder or the clock is · failure: none
- REQ-32: When `moai improve-note` runs, the command shall append exactly one line of the D7 shape to the project's `improve/notes.jsonl`, with `at` read from the clock in UTC, creating the folder and the file when they are absent · failure: none
- REQ-33 (REQ-32): IF the kind is outside `correction`, `failure` and `false_claim`, or the key is over 80 characters, or the text is over 200 characters, or the append would take the file over 1 MiB (the read limit of `improve-signals`), THEN the command shall refuse and write nothing.
- REQ-34: When the key or the text of a note carries a secret-shaped word, the command shall write `[redacted]` in its place · failure: none
- REQ-35 (REQ-32): IF the `improve` folder or the notes file is a symbolic link, or the notes file is not a regular file, or the state folder above them is a link, THEN the command shall refuse and write nothing.
- REQ-36: When the notes file is opened for the append, the command shall open it without following a link · failure: none

## Acceptance

Acceptance lines AC-1…AC-9 exist for REQ-1…REQ-11, the requirements of slice S1; AC-10…AC-18 are the lines slice S2 added in revision 2 for REQ-12…REQ-15 and REQ-26…REQ-36. REQ-16…REQ-25 are fixed by this SPEC; each later slice adds its own acceptance lines (revision bump) from the boundary table. The host is Claude Code only (D10).

- AC-1 (REQ-1…REQ-3): in a temporary HOME and project with a ledger, two `AUDIT … verdict: FAIL` notes that cite `src/a.rs`, one that cites `src/b.rs`, two `LANE-BLOCKED … kind=ci` notes, one card titled `flaky: some_test` and improve notes of the kind `correction` give the classes `audit_fail:src/a.rs`, `audit_fail:src/b.rs`, `block:ci`, `flake:some_test` and `correction:<topic>` with counts, times and at most five examples; an absent ledger and an absent notes file report `absent` with exit 0; an unreadable ledger and a symbolic link as the notes file are refused.
  verify: `cargo nextest run --locked --offline --test improve_signals_cli`
- AC-2 (REQ-4…REQ-5): boundary: one occurrence is `observed_once` and two are `candidate`, for each of the four sources.
  verify: `cargo nextest run --locked --offline --test improve_signals_cli`
- AC-3 (REQ-6…REQ-7): `LANE-READY`, `LANE-DONE`, `CANDIDATE`, `MERGE-READY`, `MERGED-LOCAL` and unmatched notes that cite paths are never counted and appear under `skipped` with `no_rule`; a `false_claim` improve note is counted; a READY note that says "fixed src/a.rs" never raises `audit_fail:src/a.rs`.
  verify: `cargo nextest run --locked --offline --test improve_signals_cli`
- AC-4 (REQ-8): the tree of the temporary HOME and project (`TempHome::snapshot`, folder modification times included) is the same before and after a run, with the ledger and the notes file present, and also with both absent (no `improve` folder and no ledger file appears).
  verify: `cargo nextest run --locked --offline --test improve_signals_cli`
- AC-5 (REQ-9): with a named pipe (a read would block) where Claude auto memory (`.claude/projects/<name>/memory/MEMORY.md`) would be, the run ends within its limit with exit 0 and `host_memory: "off"`.
  verify: `cargo nextest run --locked --offline --test improve_signals_cli`
- AC-6 (REQ-10): a note and an improve note that carry a secret-shaped word print `[redacted]` in its place and the raw word appears nowhere in the output; a text over 120 characters is cut at 120.
  verify: `cargo nextest run --locked --offline --test improve_signals_cli`
- AC-7 (REQ-11): the collecting function with a limit of N notes reads N and says `truncated: false`, with N+1 notes it reads N and says `truncated: true`; a body over 8,192 bytes is cut at 8,192.
  verify: `cargo nextest run --locked --offline --bin moai-cli improve_signals`
- AC-8 (REQ-1): the command is a file-op with its operations document, help row, example and goldens.
  verify: `cargo nextest run --locked --offline --bin moai-cli operations_doc_matches_code help_table_names_every_command_once every_command_has_examples`; `cargo nextest run --locked --offline --test help_layout`
- AC-9 (REQ-1…REQ-11): at most five mutants of the new code, each turning a test red, are restored.
  verify: manual: the mutant table in the Result, one `guard-mutant:` line per mutant in the READY note.
- AC-10 (REQ-12, REQ-13): `improve-plan` accepts a target that the manifest `owned_files` lists with the kind `generation:project-skill:claude:prj-*` or `generation:native-role:claude:prj-*` and refuses, with the reason `not_allowed` and no write, each of: a file the manifest does not list, the installed `moai` skill, `.claude/settings.json`, a hook file, `.mcp.json`, a file under `tests/`, a file under the evaluation set folder `improve/evals/`, `CLAUDE.md`, `CLAUDE.local.md` (grade B waits for slice S4), a path with `..`, an absolute path and a link.
  verify: `cargo nextest run --locked --offline --test improve_plan_cli`
- AC-11 (REQ-14): a proposal of one, two or three items is accepted and of zero or four is refused; each item has an `id`, one rule sentence, one reason line, one to five `source` signal ids, the status `trial` and the counters `helped` and `hurt` at 0; the merge adds, replaces or retires by item id inside the managed block `moai-lessons` of the target and every byte outside the block is the same before and after.
  verify: `cargo nextest run --locked --offline --test improve_plan_cli`; `cargo nextest run --locked --offline --bin moai-cli improve_plan`
- AC-12 (REQ-15): the budgets hold at N and refuse at N+1 on the new bytes: 200 lines for `CLAUDE.md`-class files and for project agent files (D10 names no agent budget; this line fixes it at 200), 500 lines for a skill and 1,024 characters for its `description`; a proposal that would pass a line budget is accepted when it also retires one `trial` item that brings the file back to N.
  verify: `cargo nextest run --locked --offline --test improve_plan_cli`; `cargo nextest run --locked --offline --bin moai-cli improve_plan`
- AC-13 (REQ-26…REQ-30): one failing input per gate (test deletion or weakening named by the target path, `drift` for a missing path and for a missing `moai` command, `secret`, `front_matter`) is refused with its reason, and the reasons of every failing gate are listed; `improve-plan` writes nothing (`TempHome::snapshot`, folder modification times included, equal before and after) and says `writes: false`; a conflict state is shown when the target differs from its `baseline_sha256`.
  verify: `cargo nextest run --locked --offline --test improve_plan_cli`
- AC-14 (REQ-31): the same arguments in two temporary projects with the same file bytes give the same `plan_digest`, and a changed item, a changed target byte or a changed `baseline_sha256` gives another one.
  verify: `cargo nextest run --locked --offline --test improve_plan_cli`
- AC-15 (REQ-32…REQ-34): `improve-note` appends one D7 line (the `at` field is UTC RFC 3339), creates the folder and the file when absent, and the line is read back by `improve-signals` as the class `<kind>:<key>`; a kind outside the three, a key of 81 characters, a text of 201 characters and an append over the 1 MiB read limit are refused with nothing written, and a key of 80 characters, a text of 200 characters and a file that ends exactly at the limit are accepted; a secret-shaped word is written as `[redacted]`.
  verify: `cargo nextest run --locked --offline --test improve_note_cli`
- AC-16 (REQ-35, REQ-36): with the real binary in a temporary project, `improve/` as a symbolic link, `improve/notes.jsonl` as a symbolic link and as a named pipe, and the state folder above them as a link each refuse the append, the link target is not written, and nothing else under the temporary HOME and project changes.
  verify: `cargo nextest run --locked --offline --test improve_note_cli`
- AC-17 (REQ-26, REQ-32): both commands are file-ops with their operations documents, help rows, examples and goldens.
  verify: `cargo nextest run --locked --offline --bin moai-cli operations_doc_matches_code help_table_names_every_command_once every_command_has_examples`; `cargo nextest run --locked --offline --test help_layout`
- AC-18 (REQ-12…REQ-15, REQ-26…REQ-36): one mutant per gate and per refusal path of the new code (allow-list, test path, `drift`, `secret`, `front_matter`, each budget, item count, digest input, link refusal, size cap), each turning a named test red, is restored.
  verify: manual: the mutant table in the Result, one `guard-mutant:` line per mutant in the READY note.

### 조각과 카드의 경계

한 요구는 한 조각이 맡고, 한 조각은 한 카드가 맡아요. 카드 본문은 승인 뒤 리더가 이 표에 맞춰 고쳐 주세요(C-229~C-232 의 본문은 아직 Codex 경로와 `AGENTS.md` 를 말해요).

| 조각 | 카드 | 요구 | 하는 일 | 쓰는 곳 | 상태 |
| --- | --- | --- | --- | --- | --- |
| S1 신호 수집 | C-228 | REQ-1~REQ-11 | 읽기 전용 `improve-signals` 와 이 SPEC | `src/ops/improve_signals.rs`, `tests/improve_signals_cli.rs`, 운영 문서 조각, 줄 임대로 `src/help.rs` 와 `tests/fixtures/help/*.txt` | 이 SPEC 승인 뒤 시작. 옛 구현은 `a4f4f177ca` 에만 있어요 |
| S2 제안과 정적 게이트 | C-229 | REQ-12~REQ-15, REQ-26~REQ-36 | `improve-plan`(등급 A 델타 제안, `generation-plan` 미리보기, 정적 게이트, `plan_digest`)과 신호를 쓰는 `improve-note`(D7). `improve-note` 의 요구 줄은 개정 2 의 REQ-32~REQ-36 이에요 | 새 `src/ops/improve_*.rs`, `src/generation.rs` 는 읽기만 | S1 착지 뒤 |
| S3 평가 게이트 | C-230 | REQ-16~REQ-18, REQ-24 | `improve-eval`: 사례 재생, 결정적 검사기, 판정 `accept`·`reject`·`needs_human`·`blocked` | 새 `src/ops/improve_eval.rs`, 평가 세트 파일(루프가 못 쓰는 자리) | S2 착지 뒤 |
| S4 적용·되돌리기 | C-231 | REQ-19~REQ-23 | `improve-apply`·`improve-revert`, `skills/moai/workflows/improve.md` 와 루트, `CLAUDE.local.md` 관리 블록 예외(`src/agents.rs`) | 새 `src/ops/improve_*.rs`, `skills/moai/**`(소스 매니페스트 줄 임대), `src/agents.rs` | S3 착지 뒤. 스킬 변경이라 사본 동기화 규칙을 따라요 |
| S5 유지보수 루프 | C-232 | REQ-25 | 저장소 전용 어댑터, 배포 텍스트 ratchet 스캔, 평가 사례 네 개 | `tools/improve/**`, 스캔 표(줄 임대) | 카드 상태 `hold`. 스캔과 사례는 지금도 시작할 수 있지만 엔진 연결은 S1 뒤 |

경계 규칙: (1) S1 은 어떤 파일도 쓰지 않아요(REQ-8), 쓰는 명령은 S2 부터예요. (2) S2 는 평가를 실행하지 않고 S3 는 적용하지 않아요. (3) 호스트 메모리 읽기(D3)는 어느 조각에도 없어요. 켜는 조각은 따로 SPEC 개정으로 더해요. (4) S5 의 어댑터는 배포 패키지 매니페스트에 들어가지 않아요(REQ-25). (5) C-229~C-232 의 본문에 남은 Codex·`AGENTS.md`·`AGENTS.local.md` 말은 이 SPEC 이 우선해요.

## Scope

- write (slice S1, 승인 뒤): `.moai/specs/SPEC-20261121-self-improvement-loop/**`, `src/ops/improve_signals.rs` (로직은 이 파일에 자기 단위 테스트와 함께 두어 `src/lib.rs` 를 건드리지 않아요), `docs/design-review/appendix/operations/improve_signals.md`, `tests/improve_signals_cli.rs`
- write (slice S2, 개정 2): `.moai/specs/SPEC-20261121-self-improvement-loop/**`, `src/ops/improve_plan.rs`, `src/ops/improve_note.rs` (각자 자기 단위 테스트와 함께), `docs/design-review/appendix/operations/improve_plan.md`, `docs/design-review/appendix/operations/improve_note.md`, `tests/improve_plan_cli.rs`, `tests/improve_note_cli.rs`. `src/generation.rs` 와 `src/contract.rs` 는 읽기만 해요(S2 는 허용 경로를 매니페스트의 `owned_files` 에서 읽고, 거절은 S1 처럼 `improve-plan refused: <이유>` 메시지와 종류 `invalid_input` 이에요).
- 줄 임대로만(공유 파일, 카드 범위에 넣지 않아요): `src/help.rs`(`GROUPS` 행 하나, `EXAMPLES` 행 하나), `tests/fixtures/help/*.txt`(그 명령을 담는 골든), 새 테스트 파일이 체크인 파일을 읽으면 `tools/repo_scans.json`
- read-only: `src/ledger/open.rs`(`open_read_only`), `src/card/export_db.rs`(`read_snapshot`), `src/card/inbox.rs`, `src/lane_next.rs`, `src/redact.rs`, `src/state_root.rs`, `src/home_root.rs`, `tests/common/home.rs`, `skills/moai/references/internal/check-rules.md`, `skills/moai/workflows/audit.md`, `skills/moai/workflows/work.md`
- forbidden: `skills/moai/**`(`improve` 루트는 조각 S4), `agents/*.md`, `CLAUDE.md`, `src/agents.rs`(D1 예외는 조각 S4), `src/generation.rs`, 철회된 호스트의 경로, 카드·장부·실제 HOME 에 쓰기, 네트워크
- 읽은 사실(2026-11-27, 이 tip `e98798e41f`): 명령은 `src/ops/<word>.rs` 한 파일이 `build.rs` 의 생성 목록으로 하위 명령이 돼요(`improve_signals.rs` → `improve-signals`, `src/ops/mod.rs` 머리말, 레지스트리는 이름과 도움말 줄만 담는 경로 없는 형태). 새 명령은 운영 문서 조각 `docs/design-review/appendix/operations/<word>.md`(`operations_doc_matches_code`, `src/main.rs:3353` 이 코드와 맞는지 봐요), `src/help.rs` 의 `GROUPS`·`EXAMPLES` 행, 도움말 golden 이 필요해요(`tests/fixtures/help/` 에 `all*.txt` 와 `root*.txt`, 그 명령을 담는 것만 바뀌어요). 장부는 `ledger::open_read_only`(`src/ledger/open.rs:120`, 없으면 `Absent`, 아무것도 안 만듦), 모든 노트 본문은 `card::export_db::read_snapshot`(`src/card/export_db.rs:33`). 가림은 `redact::redact_secret_shapes`(`src/redact.rs:42`)와 `redact::redact_outgoing`(`src/redact.rs:76`). 프로젝트 상태 폴더 아래 위치는 `state_root::path(root, sub)`(`src/state_root.rs:847`)로 구하고 상태 폴더 이름을 직접 쓰지 않아요(상태 폴더 리터럴 스캔). 감사 판정을 읽는 기존 파서는 PR 댓글용(`AUDIT #<n>` 로 시작, `src/lane_next.rs:346-349`)이라 카드 노트용 파서는 새로 필요해요. 카드 노트의 판정 첫 줄은 `AUDIT <card> (<scope>) head <40hex> base <40hex> — verdict: PASS|FAIL` 예요(`CLAUDE.md` 흐름 6). `LANE-BLOCKED kind=` 를 읽는 파서는 없어요(`src/card/inbox.rs:56` 은 접두사만 봄). 플레이키 기록은 카드 제목 `flaky: <test>`(`skills/moai/references/internal/check-rules.md:37`, 다시 나면 카드에 `--kind report` 노트, 제목 접두사 필터는 `src/ops/card.rs:57`). 점검 실패의 종료 코드를 남기는 곳은 없어요(`src/verify.rs:1-5`, `check affected` 로그는 워크트리 `target/`). 테스트 스냅샷은 `TempHome::snapshot`(`tests/common/home.rs:154`). 2026-11-21 초안이 읽었던 `verify.md`, `mode-agent.md`, `tests/card_inbox_cli.rs` 는 지금 없어요.

## Plan

조각 순서(보고서 §6.1)는 위 경계 표예요. 이 개정의 일은 SPEC 초안을 지금 사실에 맞추는 데까지이고 승인 전에는 코드를 쓰지 않아요.

1. **지금(이 개정)**: 초안을 NeXT 사실에 맞추고 `moai spec check` 를 돌려 SPEC-READY 에서 멈춰요. 사용자에게 물을 갈림길은 `LANE-BLOCKED kind=decision` 노트에 모았어요. 승인은 레인이 적지 않아요.
2. **S1(승인 뒤, C-228)**: 새 워크트리에서 RED 부터. `tests/improve_signals_cli.rs` 를 `TempHome` 으로 먼저 쓰고(부류별 픽스처, 경계, 주장 제외, 스냅샷, 이름 붙은 파이프, 비밀 가림, 심볼릭 링크) 실패를 확인해요. GREEN 은 `src/ops/improve_signals.rs` 안에 순수 함수(분류 키 추출, 집계, 경계 상수)와 단위 테스트, 읽는 쪽은 얇은 함수로 둬요. 운영 문서 조각과 도움말 행·예시·golden 은 공유 파일이라 줄 임대로 더해요. 옛 구현(`a4f4f177ca`)의 세 파일은 참고로 읽을 수 있지만 결과는 새로 돌려서 적어요(D12).
3. **검증(S1)**: AC 의 verify 명령을 적힌 그대로, 슬롯 안에서. 바뀐 테스트 대상은 nextest 로도, 가드 뮤턴트 5개 이하를 하나씩(`guard-mutant:` 줄), 맨 마지막에 `moai check affected --card <base>..<head>` 한 번.
4. **S2~S5**: 경계 표의 카드가 각자 수용 줄을 개정으로 더하고 자기 SPEC 승인을 받아요. S1 이 착지하기 전에는 시작하지 않아요(S5 의 스캔과 사례만 예외).

## Risks

- 감사 노트의 경로 토큰은 거친 신호예요. 다른 문장 속 경로(예: 예시로 든 파일)도 세어 부풀릴 수 있어서, 이 명령은 부류와 횟수와 예시 다섯 개만 보여 주고 아무것도 제안하지 않아요(제안은 S2, 사람이 선택). 노이즈가 크면 D5 의 규칙을 바꾸는 것이 되돌리기예요.
- 판정 줄 형식이 바뀌면 신호가 조용히 줄어요. 맞는 줄은 `skipped` 에 `no_rule` 로 세어져 보이고, 형식의 출처(`skills/moai/workflows/audit.md`, `CLAUDE.md`)를 운영 문서 조각에 적어요.
- 플레이키 횟수(D6)는 근사값이에요. 이 근사값이 "두 번째 플레이키"를 잘못 만들 수 있으나 이 명령은 읽기 전용이라 해가 없고, S2 에서 정확한 기록으로 바꿔요.
- 노트 본문에는 개인 정보나 비밀이 있을 수 있어요. 출력은 120자·가림을 거치고(REQ-10), 파일에 저장하지 않으며(REQ-8), 네트워크로 보내지 않아요(비목표).
- 큰 장부는 느려요. 최근 20,000 개·본문 8,192 바이트 상한(REQ-11)과 경계 테스트로 막아요.
- 호스트 메모리는 실제 HOME 읽기라 S1 은 읽지 않고(REQ-9, 이름 붙은 파이프 테스트로 증명), 켜는 것은 D3 에 따라 사용자 옵트인 뒤 별도 조각에서 해요.
- 옛 구현이 SPEC 승인 전에 쓰였어요(이력). 이번에는 승인 전에 코드를 쓰지 않아요. 옛 브랜치의 결과를 이 SPEC 의 증거로 세지 않아요(D12).
- 이 SPEC 은 요구가 25개라 `moai spec check` 가 `size_requirements` 경고를 내요(12개 초과). 쪼개는 길은 D11 의 선택지예요.
- 루프 전체의 위험(보고서 §5): 평가기 속이기, 체크리스트 변조, 자기 강화 오류, 맥락 붕괴, 지침 비대화, 지침은 강제가 아님, 감독 신호 최적화, 경로 의존성·비용 폭주, 팀 지침 충돌, 비밀 유출, 외부 스킬 공급망, 메모리를 규칙으로 착각, 샌드박스 탈출. 각 장치는 REQ-8~REQ-15 에 들어 있고, 구현하는 조각이 자기 수용 기준으로 증명해요.

## Decisions

각 결정은 (무엇 · 선택지 · 기본값 · 되돌릴 수 있음 · 범주 · 기한 · 결정자) 순서예요. D1~D3 은 리더가 2026-11-20T21:30Z(카드 노트 4592)에 정했어요. D4~D12 는 이 초안의 제안이에요(결정자: SPEC 승인 때 사용자·리더, 기한: 이 SPEC 의 병합).

- D1 (리더 결정, 되돌릴 수 있음, 범주 guarded, 기한 없음): 개인 파일 `CLAUDE.local.md` 에 `moai-lessons` 관리 블록을 허용해요. 처음 결정(2026-11-20)은 `AGENTS.local.md` 였고, 사용자가 정본을 `CLAUDE.md` 로 바꾼 뒤(2026-11-22) 이름만 따라가요. 사용자 글은 블록 밖에서 절대 건드리지 않고, `src/agents.rs` 의 "한 번 만들고 고치지 않음" 규칙은 블록 안만 예외로 고쳐요(REQ-23, 조각 S4). 사용자의 개인 파일을 쓰는 일이라 이름을 바꾼 지금 사용자 확인을 받아요(결정 노트의 Q1). 대안: 제안 파일만 만들고 사용자가 붙임 / 개인 파일은 대상에서 뺌. 기본값: 허용.
- D2 (리더 결정, 되돌릴 수 있음, 범주 none, 기한 없음): 명령 이름은 `moai improve`, 하위는 signals·plan·eval·apply·revert. 대안: `learn`, `lessons`.
- D3 (리더 결정, 되돌릴 수 있음, 범주 guarded, 기한 없음): 호스트 메모리(Claude 자동 메모리) 읽기는 기본 꺼짐이고 프로젝트별 옵트인이에요. 실제 HOME 을 읽는 일이라 켜는 것은 사용자의 선택이에요. S1 은 호스트 메모리를 전혀 읽지 않아요(REQ-9). Codex 메모리는 호스트 철회로 대상이 아니에요.
- D4 (제안, 되돌릴 수 있음, 범주 none): D2 의 이름 규칙을 이 저장소의 명령 규칙에 맞춰 평평한 이름으로 써요: `improve-signals`, `improve-plan`, `improve-eval`, `improve-apply`, `improve-revert`, `improve-note`. 이유: `src/ops/<word>.rs` 한 파일이 하위 명령이 되는 구조라 중첩 하위 명령(`moai improve signals`)은 `src/main.rs` 의 명령 정의를 바꿔야 해요. 사용자에게 보이는 이름은 스킬 루트 `/moai improve` 로 같아요. 대안: 중첩 하위 명령. 기본값: 평평한 이름. 되돌리기: 이름만 바꾸면 돼요(아직 배포 전).
- D5 (제안, 되돌릴 수 있음, 범주 none): 부류 키 규칙(결정적, LLM 없음). 감사 노트: 첫 줄이 `AUDIT ` 로 시작하고 `verdict: FAIL` 을 담은 노트(카드 노트 모양 `AUDIT <card> (<scope>) head … — verdict: FAIL`, PR 댓글 모양 `AUDIT #<n> …` 도 받아요)만 신호이고, 본문에서 `[A-Za-z0-9_./-]+\.(rs|md|py|json|yml|yaml|toml|sh|txt)` 모양의 경로 토큰(뒤의 `:줄` 은 뗌)마다 `audit_fail:<경로>` 한 신호, 경로가 없으면 `audit_fail:uncited`. 같은 노트가 같은 경로를 두 번 말해도 한 번이에요. 차단 노트: 첫 줄에 `LANE-BLOCKED` 가 있고 `kind=<단어>` 가 있으면 `block:<kind>`(없으면 `block:unknown`). 플레이키: 제목이 `flaky: ` 로 시작하는 카드의 테스트 이름(제목에서 접두사 뒤 첫 공백 앞)으로 `flake:<test>`. 메모: `improve/notes.jsonl` 줄의 `kind`·`key` 로 `<kind>:<key>`. 대안: 감사 본문을 LLM 으로 분류(결정적이 아니라 거절). 기본값: 경로 토큰.
- D6 (제안, 되돌릴 수 있음, 범주 none): 플레이키 횟수 규칙. 플레이키 기록은 테스트마다 카드 하나(`check-rules.md:37`)이고 재발은 그 카드의 `report` 노트로 남아요. 그래서 횟수는 "그 카드 1개 + 그 카드에 달린 `report`·`observation` 종류 노트 수"로 세요. 노트의 형식을 정한 문서가 없어 이 규칙은 근사값이에요. 대안: 카드 수만 셈(한 테스트는 항상 1). 기본값: 위 규칙. 되돌리기: 조각 S2 가 `improve-note` 로 정확한 재발을 남기면 이 규칙을 대체해요.
- D7 (제안, 되돌릴 수 있음, 범주 none): `improve/notes.jsonl` 형식. 줄마다 JSON 한 개: `{"v":1,"at":"<UTC RFC 3339>","kind":"correction|failure|false_claim","key":"<주제 또는 명령, 80자 이하>","text":"<200자 이하>","by":"<관찰한 쪽>"}`. 읽을 때 알 수 없는 `kind`·깨진 줄은 `skipped`(`bad_line`)로 세고 멈추지 않아요. 위치는 프로젝트 상태 폴더 아래 `improve/notes.jsonl`(`state_root::path(root, "improve/notes.jsonl")`, 철자를 직접 쓰지 않음), 관리형 `.gitignore` 블록이 이미 상태 폴더의 나머지를 무시해요. 이 파일을 쓰는 명령(`improve-note`)은 조각 S2 예요.
- D8 (제안, 되돌릴 수 있음, 범주 none): 점검 실패의 종료 코드는 저장된 곳이 없어서(`src/verify.rs:1-5`, `check affected` 로그는 워크트리 `target/` 아래) 조각 S1 은 읽지 않아요. 같은 정보는 `improve-note` 의 `failure` 종류로만 들어와요(조각 S2). 대안: `target/check-affected-*/` 로그를 읽기(워크트리마다 달라 불안정). 기본값: 읽지 않음.
- D9 (제안, 되돌릴 수 있음, 범주 none): 신호의 두 번 규칙(보고서 P9)은 조각 S1 에서 고정 2 예요(경계 N=1/N=2). 프로필로 바꾸는 것은 조각 S4 의 속도 제한과 함께.
- D10 (제안, 되돌릴 수 있음, 범주 none): 호스트는 Claude Code 하나예요(Codex·OMP 철회). 지침 로딩: Claude Code 는 `CLAUDE.md` 를 읽고 개인 파일은 `CLAUDE.local.md` 예요. 프로젝트 스킬은 `.claude/skills/prj-*`(파일 감시로 즉시 반영), 에이전트는 `.claude/agents/prj-*.md`(재생은 Agent 도구 + `isolation: worktree`, `disallowedTools` 로 쓰기 막기)예요. 강제 규칙: 지침은 "맥락이지 강제 설정이 아님"이라 반복 위반은 훅·테스트 승격을 제안해요(등급 D, `guarded` 라 사용자 결정). 2026-11-21 초안의 Codex 쪽 항목(루트부터 이어 붙이기, 32 KiB 한도, `.agents/skills`, `.codex/agents`)은 뺐어요.
- D11 (제안, 되돌릴 수 있음, 범주 none): 이 SPEC 은 요구 25개로 `size_requirements` 경고를 내요. 선택지: (a) 한 SPEC 으로 두고 경고를 받아들임(조각마다 개정으로 수용 줄을 더함), (b) 설계 SPEC(REQ-12~REQ-25)과 조각 SPEC 을 `parent` 로 이어 쪼갬. 기본값: (a). 이유: 승인 한 번으로 루프 전체의 설계를 고정하고 S1 을 바로 시작할 수 있어요. 되돌리기: 승인 전에는 파일만 쪼개면 돼요.
- D12 (리더 결정 2026-11-27, 되돌릴 수 있음, 범주 none): S1 코드는 옛 브랜치(`a4f4f177ca`)에서 옮기지 않아요. 승인 뒤 새 워크트리에서 다시 붙이되(참고로 읽거나 세 파일을 골라 옮김) 모든 verify 와 뮤턴트를 새로 돌려 그 결과만 Result 에 적어요. 옛 브랜치는 지우지도 고치지도 않아요.

### 근거(보고서 인용, 읽은 날 2026-11-20)

- 루프와 평가: Darwin Gödel Machine(arXiv 2505.22954, §1·§3·§4.2·§5), ADAS(2408.08435), SICA(2504.15228), OPRO(2309.03409), GEPA(2507.19457).
- 반성과 신호: Reflexion(2303.11366), Self-Refine(2303.17651), Huang et al.(2310.01798, 외부 피드백 없는 자기 교정 실패), ExpeL(2308.10144), Agent Workflow Memory(2409.07429).
- 맥락·지침 관리: ACE(arXiv 2510.04618v3, 델타·결정적 병합·맥락 붕괴), Dynamic Cheatsheet(2504.07952), Anthropic "Effective context engineering"(주의 예산), Claude Code Memory 문서(CLAUDE.md 200줄 아래, 팀 공유와 개인 파일).
- 스킬·에이전트 공식 문서: Claude Code Skills·Subagents, Anthropic Agent Skills 모범 사례("평가를 먼저 만들어라", 제안자와 평가자 분리).
- 보상 해킹과 안전: Sakana "DGM and AI Safety"(탐지 표식 삭제·로그 위조), METR(2025-06-05), ImpossibleBench(2510.20270), Sycophancy to Subterfuge(2406.10162), OpenAI CoT 모니터링(2503.11926), STOP(2310.02304).
- 확인하지 못한 것(보고서 부록 B): DGM 부록 H 원문, Claude 자동 메모리 저장 경로, "검색" 표시 행의 수치(초록 기준).
- 이 개정에서 다시 확인하지 않은 것: 위 논문·문서 인용은 모두 2026-11-20 초안이 읽은 그대로예요. 이 개정은 저장소 사실(위 "읽은 사실")만 다시 읽었어요.

## Result

조각 S1(읽기 전용 `improve-signals`)의 결과예요. 이 브랜치(`spec/self-improvement-loop-20261121-r2`)에서 새로 구현하고 모두 다시 돌렸어요(D12). 옛 브랜치 `spec/self-improvement-loop-20261121`(`a4f4f177ca`)의 기록은 이 결과의 증거가 아니고, 그 브랜치는 건드리지 않았어요(코드는 참고로만 읽고 NeXT 사실에 맞게 다시 붙였어요).

단계: RED 커밋 `7081b1711c`(`tests/improve_signals_cli.rs` 11개, `moai tdd red` 기록: 구현 없이 돌려 11개 모두 단언 실패, 종료 코드 100), GREEN 커밋 `0c9580e8fc`(`src/ops/improve_signals.rs`, 운영 문서 조각, `src/help.rs` 행 둘, 도움말 골든 4개; `moai tdd green` 통과), 커밋 `44842d25fe`(심볼릭 링크 거절 단위 테스트). 아래 verify 결과는 NeXT tip `cf2c0b1374` 를 머지한 헤드 `4f6751b7fa` 에서 적힌 그대로 슬롯 안에서 돌린 것이에요.

이번 조각이 SPEC 본문을 넘은 곳: D6 의 사용자 답("출력에 어림값임을 표시해요")에 따라 `flake:` 부류에만 `"approximate": true` 를 넣었어요(단위 테스트 `improve_signals_only_the_flake_count_is_marked_approximate`). 요구 줄과 AC 는 바꾸지 않았어요.

- AC-1: met — `cargo nextest run --locked --offline --test improve_signals_cli` 11개 통과(exit 0). `every_source_becomes_a_class_with_count_times_and_examples`(다섯 부류, 횟수, 시각, 예시), `a_class_shows_at_most_five_examples_and_the_newest_ones`, `nothing_is_written_with_or_without_a_ledger_and_an_absent_source_is_reported`(없는 장부와 노트 파일은 `absent`, exit 0), `a_broken_ledger_and_an_unsafe_notes_file_are_refused`(링크·이름 붙은 파이프는 `invalid_input`, 망가진 장부는 `blocked`)가 이 줄을 맡아요.
- AC-2: met — 같은 11개 중 `one_signal_is_observed_once_and_two_are_candidates_for_every_source`(네 출처 모두 1 은 `observed_once`, 2 는 `candidate`)와 `bad_lines_of_the_notes_file_are_counted_and_skipped`. 단위 테스트 `improve_signals_status_flips_at_two`.
- AC-3: met — `an_agents_own_claim_is_never_a_signal`: READY·DONE·CANDIDATE·MERGE-READY·MERGED-LOCAL·"고쳤다"·판정 글자만 든 노트 일곱은 `skipped.no_rule` 7 이고, `false_claim` 메모만 부류가 돼요. `an_audit_note_cites_each_path_once_and_a_note_without_a_path_is_uncited`(경로 한 번, 경로 없으면 `uncited`, PASS 판정은 안 셈).
- AC-4: met — `nothing_is_written_with_or_without_a_ledger_and_an_absent_source_is_reported`: 장부와 노트 파일이 있을 때와 둘 다 없을 때 모두 `TempHome::snapshot`(폴더 수정 시각 포함)이 실행 전후 같고 `improve` 폴더와 장부 파일이 생기지 않아요. 결과에 `writes: false`.
- AC-5: met — `host_memory_is_never_opened`: Claude 자동 메모리 자리(`.claude/projects/demo/memory/MEMORY.md`)의 이름 붙은 파이프가 있어도 30초 안에 exit 0 이고 `host_memory: "off"`. 철회된 호스트의 파이프 자리는 이 개정에서 뺐어요.
- AC-6: met — `texts_are_masked_and_cut_to_120_characters`: `password=…`·`token=…` 모양 글은 `[redacted]` 로 나가고 원문이 출력 어디에도 없으며 300자 첫 줄은 120자로 잘려요.
- AC-7: met — `cargo nextest run --locked --offline --bin moai-cli improve_signals` 9개 통과(exit 0): `improve_signals_reads_the_newest_n_notes_and_says_when_it_cut`(N 개는 `truncated: false`, N+1 은 N 개를 읽고 `true`), `improve_signals_cuts_a_body_at_the_byte_limit`(8,192 바이트).
- AC-8: met — `cargo nextest run --locked --offline --bin moai-cli operations_doc_matches_code help_table_names_every_command_once every_command_has_examples` 3개 통과, `cargo nextest run --locked --offline --test help_layout` 11개 통과(골든 4개를 `MOAI_HELP_GOLDEN_WRITE=1` 로 머지된 트리에서 다시 만들었어요).
- AC-9: met — 가드 뮤턴트 다섯, 하나씩 걸고 이름 붙은 테스트를 슬롯에서 돌린 뒤 되돌렸고 마지막에 `git diff` 가 의도한 테스트 추가만 남겼어요.

| 뮤턴트 | 빨갛게 만든 테스트 | 종료 코드 |
| --- | --- | --- |
| `THRESHOLD` 2 → 3 | `one_signal_is_observed_once_and_two_are_candidates_for_every_source` | 100 |
| `is_audit_fail` 이 `verdict: FAIL` 을 보지 않음 | `an_audit_note_cites_each_path_once_and_a_note_without_a_path_is_uncited` | 100 |
| 심볼릭 링크 거절 제거 | `improve_signals_a_symbolic_link_is_refused_by_name_and_never_followed` | 100 (첫 시도는 살아남아 exit 0 이었어요: 다음 검사 "일반 파일이 아님"이 같은 링크를 같은 오류 종류로 거절해서예요. 그래서 이 단위 테스트를 더했고 다시 걸었을 때 100) |
| `excerpt` 가 가림을 전혀 안 함 | `texts_are_masked_and_cut_to_120_characters` | 100 |
| `read_notes` 의 `truncated` 가 `> limit` → `>= limit` | `improve_signals_reads_the_newest_n_notes_and_says_when_it_cut` | 100 |

한 가림 단계만 빼는 뮤턴트(`redact_secret_shapes` 만 제거)는 `redact_outgoing` 이 같은 모양을 가려서 살아남아요. 두 단계가 겹치는 것이라 동작 차이가 없어서(같은 입력에 같은 출력) 테스트를 더하지 않았어요. 둘 다 빼면 위 네 번째 줄처럼 빨개져요.

알려진 한계: 점검 실패의 종료 코드는 읽지 않아요(D8). 플레이키 횟수는 어림값이에요(D6, 출력에 표시). Windows 는 `not_run`(두 테스트 대상이 unix 전용이에요). `improve-note` 와 제안·평가·적용은 뒤 조각(S2~S4)이에요.
