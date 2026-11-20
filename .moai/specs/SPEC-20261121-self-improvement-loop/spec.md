---
id: SPEC-20261121-self-improvement-loop
title: "Self-improvement loop: `moai improve` proposes small gated changes to project skills, agents and AGENTS.local.md from observed signals; the first change adds the read-only `improve-signals`"
status: draft
revision: 1
created: 2026-11-21
---

# 자기 개선 루프 (`moai improve`)

## Interview

- Q1: 무엇을 만드나요? → A: 사용자 지시(2026-11-20) — 재귀적 자기 개선(RIS) 논문과 모범 사례를 모아, moai-adk 개발(로컬)과 사용자 설치(프로젝트 스킬·에이전트·AGENTS.md·AGENTS.local.md)에서 AI 가 스스로 개선하는 기능을 추가해요.
- Q2: 근거는 어디에 있나요? → A: 리더의 보고서 폴더(저장소 밖) `moai-adk/self-improvement-research-2026-11-21.md`(기준 커밋 `f2eb8c61e3`, 2026-11-20T21:23Z). 저장소 밖 파일이라 쓰는 근거를 Decisions 끝의 "근거"에 줄여 적었어요.
- Q3: 열린 결정은요? → A: 보고서 §6.3 의 셋은 리더가 2026-11-20T21:30Z 에 정했어요(D1~D3). D4~D10 은 이 초안의 제안이고 승인은 SPEC 이 병합될 때 사용자·리더가 골라요. 이 문서는 승인을 스스로 적지 않아요.
- Q4: 사용자에게 남은 질문은요? → A: 없어요. 명령 등록 방식, 노트 파서 유무, 상태 폴더 위치는 코드를 읽어 답했어요(Scope 의 "읽은 사실").

## Intent

AI 코딩 도구가 같은 실수를 되풀이하면 사람이 같은 정정을 반복해요. 이 기능은 **관찰된 반복 문제를 모아 → 프로젝트 지침·도구의 작은 변경을 제안하고 → 독립 평가로 걸러 → 사용자가 고른 것만 적용·기록하고 → 되돌릴 수 있게** 하는 루프예요. 보고서의 결론 그대로 살아남는 원칙은 하나예요: **제안 → 독립 평가 → 수용(또는 거절) → 기록**(Darwin Gödel Machine, ADAS, SICA, OPRO 가 같은 모양). MoAI 는 이미 이 뼈대를 갖고 있어요(plan digest → 사용자 선택 → apply → receipt, "구현자는 감사하지 않는다"). 새 쓰기 엔진은 만들지 않고 기존 `generation-plan/apply`(프로젝트 스킬·에이전트)와 `agents-plan/apply`(AGENTS.md 관리 구간)를 써요.

가장 큰 위험은 **평가 장치를 속이는 개선**이에요(DGM 이 탐지 표식을 지우고 테스트 통과 로그를 위조, METR·ImpossibleBench 의 테스트 변조). 그래서 평가 세트와 게이트 파일은 개선 루프가 쓸 수 없는 자리에 두고, 지침은 통째로 다시 쓰지 않고 항목 단위 델타 + 결정적(LLM 이 아닌) 병합으로만 고쳐요(ACE: 통째 재작성은 18,282 → 122 토큰, 정확도 66.7 → 57.1). 신호는 **외부에서 관찰된 사실**만 쓰고 에이전트의 주장은 신호가 아니에요(Huang et al.: 외부 피드백 없는 자기 교정은 믿기 어려움).

이 SPEC 은 루프 전체의 설계를 한 번에 정하고(요구 REQ-1~REQ-15), 구현은 조각(slice) 다섯 개로 나눠요(Plan). **이번 변경은 조각 S1 하나**예요: 읽기 전용 명령 `moai improve-signals` — 관찰된 신호를 모아 문제 "부류(class)"로 묶고 횟수를 세어 보여 줘요. 쓰기, 제안, 평가는 이번 변경에 없어요.

## Non-goals

- 이번 변경에서: 제안 만들기, 어떤 파일이든 쓰기, 평가 실행, 네트워크, 호스트 메모리(Claude 자동 메모리·Codex memories) 읽기, 스킬 본문(`skills/moai/**`)·`agents/*.md` 문구 변경(`improve` 루트 문서는 조각 S4).
- 루프 전체에서(보고서 §3.9·§6.2): 세션 중 자동 적용, 백그라운드 상시 개선, 다른 사람 파일 수정, 외부에서 스킬 가져오기, 모델 가중치·호스트 설정·권한·훅의 자동 변경, 진화 탐색(AlphaEvolve·DGM 식 집단 탐색), 평가 세트 자동 확장, "감사 FAIL 수 줄이기" 같은 감독 신호를 점수로 직접 최적화.
- 팀 공유 `AGENTS.md` 를 개선 명령이 직접 쓰기(항상 SPEC 경로).

## Requirements

요구 REQ-1~REQ-11 은 이번 변경(조각 S1)이고, REQ-12~REQ-25 는 이 SPEC 이 정해 두는 루프 전체의 설계로 뒤 조각이 구현해요.

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
- REQ-12: When a later command proposes a change, it shall touch only grade A (project skills and agents with the `prj-` prefix under `.claude/skills`, `.agents/skills`, `.claude/agents` and `.codex/agents`), grade B (the managed block `moai-lessons` of `AGENTS.local.md`, D1) or a SPEC draft for `AGENTS.md` (grade C) · failure: none
- REQ-13 (REQ-12): IF a proposal targets the installed `moai` skill, host settings, permissions, keys, hooks, MCP settings, a test deletion or weakening, or the evaluation set or gate files, THEN the plan shall be refused with the reason and the error kind `refused`.
- REQ-14: When a proposal is built, it shall be one target file with one to three items, each with an `id`, one rule sentence with one reason line, `source` signal ids, the status `trial` and the counters `helped` and `hurt`, merged by deterministic code inside the managed block by item id · failure: none
- REQ-15 (REQ-14): IF the target file would pass its size budget (200 lines for CLAUDE.md-class files, 500 lines and a 1,024-character description for a skill, 32 KiB for AGENTS.md), THEN the proposal shall be refused unless it also retires one old `trial` item (N is accepted, N+1 is refused).
- REQ-16: When a proposal is evaluated, the evaluation shall run the static gate, then replay the triggering case before and after with a new sub-agent that cannot write, then the project's regression set, then the project's own tests where a script or tool changed, and record `accept`, `reject` or `needs_human`, with any LLM opinion kept as advice and never as the gate · failure: none
- REQ-17 (REQ-16): IF a case that passed before the change fails after it, or the triggering case still fails, THEN the verdict shall be `reject`.
- REQ-18 (REQ-16): IF a checker cannot decide, THEN the verdict shall be `needs_human` and never a pass.
- REQ-19: When a proposal is applied, the command shall require the plan digest, a caller-supplied permission reference and a user pick made after a preview (the diff, the signals with counts and dates, the evaluation table and the revert command), write a receipt and one ledger line with the decision record fields, and never edit the team `AGENTS.md` itself · failure: none
- REQ-20 (REQ-19): IF the digest or the permission reference is missing or stale, THEN the apply shall be refused and nothing is written.
- REQ-21: When a change is reverted, the command shall restore the bytes of the receipt in the reverse order of applying · failure: none
- REQ-22: When a new item is applied, it shall start as `trial` for five sessions and then be proposed as `kept` or `retire`, within the rate limits of one round a day, three proposals a round, three items and twenty added lines a proposal, no second change of one file within seven days (a revert excepted) and twenty-four replays a round, each accepted at N and refused at N+1 · failure: none
- REQ-23: Where the project's `AGENTS.local.md` holds a managed block `moai-lessons`, the writing code shall change only the lines inside the block markers, and `src/agents.rs` names the block as the one exception to "created once and never edited" · failure: none
- REQ-24 (REQ-16): IF an evaluation file changed since the plan was built, THEN the evaluation shall be `blocked`.
- REQ-25: Where a maintainer adapter or a maintainer evaluation case exists (`tools/improve/maintainer-signals.json`), it shall stay outside the shipped package, and a scan fails when the count of maintainer-only tokens in shipped text rises above the base commit's · failure: none

## Acceptance

Acceptance lines exist for REQ-1…REQ-11, the requirements of this change (slice S1). REQ-12…REQ-25 are fixed by this SPEC; each later slice adds its own acceptance lines (revision bump) from the Plan, and the host differences are fixed in D10.

- AC-1 (REQ-1…REQ-3): in a temporary HOME and project with a ledger, two `AUDIT … verdict: FAIL` notes that cite `src/a.rs`, one that cites `src/b.rs`, two `LANE-BLOCKED … kind=ci` notes, one card titled `flaky: some_test` and improve notes of the kind `correction` give the classes `audit_fail:src/a.rs`, `audit_fail:src/b.rs`, `block:ci`, `flake:some_test` and `correction:<topic>` with counts, times and at most five examples; an absent ledger and an absent notes file report `absent` with exit 0; an unreadable ledger and a symbolic link as the notes file are refused.
  verify: `cargo test --offline --locked --test improve_signals_cli`
- AC-2 (REQ-4…REQ-5): boundary: one occurrence is `observed_once` and two are `candidate`, for each of the four sources.
  verify: `cargo test --offline --locked --test improve_signals_cli`
- AC-3 (REQ-6…REQ-7): `LANE-READY`, `LANE-DONE`, `CANDIDATE`, `MERGE-READY`, `MERGED-LOCAL` and unmatched notes that cite paths are never counted and appear under `skipped` with `no_rule`; a `false_claim` improve note is counted; a READY note that says "fixed src/a.rs" never raises `audit_fail:src/a.rs`.
  verify: `cargo test --offline --locked --test improve_signals_cli`
- AC-4 (REQ-8): the tree of the temporary HOME and project (`TempHome::snapshot`, folder modification times included) is the same before and after a run, with the ledger and the notes file present, and also with both absent (no `improve` folder and no ledger file appears).
  verify: `cargo test --offline --locked --test improve_signals_cli`
- AC-5 (REQ-9): with named pipes (a read would block) where Claude auto memory (`.claude/projects/<name>/memory/MEMORY.md`) and Codex memories (`.codex/memories/m.md`) would be, the run ends within its limit with exit 0 and `host_memory: "off"`.
  verify: `cargo test --offline --locked --test improve_signals_cli`
- AC-6 (REQ-10): a note and an improve note that carry a secret-shaped word print `[redacted]` in its place and the raw word appears nowhere in the output; a text over 120 characters is cut at 120.
  verify: `cargo test --offline --locked --test improve_signals_cli`
- AC-7 (REQ-11): the collecting function with a limit of N notes reads N and says `truncated: false`, with N+1 notes it reads N and says `truncated: true`; a body over 8,192 bytes is cut at 8,192.
  verify: `cargo test --offline --locked --bin moai-cli improve_signals`
- AC-8 (REQ-1): the command is a file-op with its operations document, help row, example and goldens.
  verify: `cargo test --offline --locked --bin moai-cli operations_doc_matches`; `cargo test --offline --locked --bin moai-cli help_table_names_every_command_once`; `cargo test --offline --locked --bin moai-cli every_command_has_examples`; `cargo test --offline --locked --test help_layout`
- AC-9 (REQ-1…REQ-11): at most five mutants of the new code, each turning a test red, are restored.
  verify: manual: the mutant table in the Result.

## Scope

- write (slice S1): `.moai/specs/SPEC-20261121-self-improvement-loop/**`, `src/ops/improve_signals.rs` (the logic stays in this file with its own unit tests, as `src/ops/claim_timeout_check.rs` does, so `src/lib.rs` is not touched), `src/help.rs` (one `GROUPS` row, one `EXAMPLES` row), `tests/fixtures/help/all.txt`, `tests/fixtures/help/all-60.txt`, `tests/fixtures/help/all-80.txt`, `tests/fixtures/help/all-dumb.txt`, `docs/design-review/appendix/operations/improve_signals.md`, `tests/improve_signals_cli.rs`
- read-only: `src/ledger/open.rs` (`open_read_only`), `src/card/export_db.rs` (`read_snapshot`), `src/card/inbox.rs`, `src/lane_next.rs`, `src/redact.rs`, `src/state_root.rs`, `src/hooks.rs` (`log_lines`), `tests/common/home.rs`, `tests/card_inbox_cli.rs`, `skills/moai/workflows/mode-agent.md`, `skills/moai/workflows/verify.md`
- forbidden: `skills/moai/**` (the `improve` route is slice S4), `agents/*.md`, `src/agents.rs` (D1 correction is slice S4), `src/generation.rs`, any write to a card, a ledger, the real HOME or the network
- 읽은 사실(2026-11-20, 이 tip `0989b3fc58`): 명령은 `src/ops/<word>.rs` 한 파일이 `build.rs` 로 하위 명령이 돼요(`improve_signals.rs` → `improve-signals`, `src/ops/mod.rs:3-4`). 새 명령은 운영 문서 조각 `docs/design-review/appendix/operations/<word>.md`(첫 줄 정확히 ``# `improve-signals` ``, `src/main.rs:3226`), `src/help.rs` 의 `GROUPS`·`EXAMPLES` 행, 도움말 golden 네 개가 필요해요. 장부는 `ledger::open_read_only`(없으면 `Absent`, 아무것도 안 만듦), 모든 노트 본문은 `card::export_db::read_snapshot`. 감사 판정을 읽는 기존 파서는 PR 댓글용(`AUDIT #` 로 시작, `src/lane_next.rs:347`)이라 카드 노트용 파서는 새로 필요해요. `LANE-BLOCKED kind=` 를 읽는 파서는 없어요(`src/card/inbox.rs:56` 은 접두사만 봄). 플레이키 기록은 카드 제목 `flaky: <test>`(`skills/moai/workflows/verify.md:53`). 점검 실패의 종료 코드를 남기는 곳은 없어요(`src/verify.rs:1-5`, `src/check_affected.rs` 는 워크트리 `target/` 로그).

## Plan

조각 순서(보고서 §6.1). 이번 변경은 S1 만이에요.

1. **S1 신호 수집 — `improve-signals`(읽기 전용)**: 이 SPEC 의 REQ-1~REQ-7.
2. **S2 제안과 정적 게이트 — `improve-plan` 과 쓰기 쪽 신호 기록 `improve-note`**: REQ-8·REQ-9. 등급 A 대상의 델타 제안, `generation-plan` 경유 미리보기, 정적 게이트, `plan_digest`. 수용: 허용 목록 밖 거절, 크기 예산 N/N+1(200줄·500줄·1,024자·32 KiB), 테스트 삭제·약화 거절, 가리키는 명령·경로가 없으면 `drift` 거절, 같은 입력 같은 digest. 범위 밖: AGENTS.md·AGENTS.local.md 쓰기, 평가 실행.
3. **S3 평가 게이트 — `improve-eval` 과 `improve` 워크플로의 재생 단계**: REQ-10. 수용: 이전 통과 사례가 실패로 바뀌면 `reject`, 촉발 사례가 그대로 실패면 `reject`, 결정 못 하면 `needs_human`, 평가 파일이 바뀌면 `blocked`, 재생 횟수 N/N+1.
4. **S4 적용·되돌리기·원장·속도 제한 — `improve-apply`/`improve-revert` 와 `skills/moai/workflows/improve.md`**: REQ-11~REQ-13·REQ-15. 수용: digest·permission reference 없으면 거절, revert 뒤 바이트 동일, 속도 제한 N/N+1, 원장 줄에 결정 기록 필드 전부, 두 호스트의 대상 경로가 `src/generation.rs` 와 같음.
5. **S5 유지보수 루프 연결(저장소 전용)**: REQ-14. `tools/improve/maintainer-signals.json` 어댑터, `AGENTS.local.md:74` 규칙을 엔진 명령으로 바꾸는 문장, 배포 텍스트 분리 ratchet 스캔(늘어나지 않게만), 보고서 §4.4 의 네 사례를 평가 세트로.

S1 단계:

1. RED: `tests/improve_signals_cli.rs` 를 `TempHome` 으로 먼저 쓰고(부류별 픽스처, 경계, 주장 제외, 스냅샷, FIFO, 비밀 가림, 심볼릭 링크), 실패를 확인해요.
2. GREEN: `src/ops/improve_signals.rs` 안에 순수 함수(분류 키 추출, 집계, 경계 상수)와 그 단위 시험을 두고, 읽는 쪽(장부·메모 파일)은 같은 파일의 얇은 함수로 둬요. 노트는 `ledger::open_read_only` + `read_snapshot`, 메모 파일은 `state_root::path` 로 위치를 구하고 심볼릭 링크 거절·크기 상한으로 읽어요.
3. 운영 문서 조각, 도움말 행·예시·golden 4개를 더해요(`MOAI_HELP_GOLDEN_WRITE=1 cargo test --test help_layout`).
4. 변형 5개 이하, 이웃 검사(`home_isolation_scan`, `repo_scans` 목록 전부, clippy 전체, `--lib`·`--bins`), `spec check`·`trailer-check` 를 돌리고 Result 에 적어요.

## Decisions

각 결정은 (무엇 · 선택지 · 기본값 · 되돌릴 수 있음 · 범주 · 기한 · 결정자) 순서예요. D1~D3 은 리더가 2026-11-20T21:30Z(카드 노트 4592)에 정했어요. D4~D9 는 이 SPEC 초안의 제안이에요(결정자: SPEC 승인 때 사용자·리더, 기한: 이 SPEC 의 병합).

- D1 (리더 결정, 되돌릴 수 있음, 범주 none, 기한 없음): `AGENTS.local.md` 에 `moai-lessons` 관리 블록을 허용해요. 사용자 글은 블록 밖에서 절대 건드리지 않고, `src/agents.rs:6` 의 "한 번 만들고 고치지 않음" 규칙은 블록 안만 예외로 고쳐요(REQ-15, 조각 S4). 대안: 제안 파일만 만들고 사용자가 붙임 / 개인 파일은 대상에서 뺌. 기본값: 허용.
- D2 (리더 결정, 되돌릴 수 있음, 범주 none, 기한 없음): 명령 이름은 `moai improve`, 하위는 signals·plan·eval·apply·revert. 대안: `learn`, `lessons`.
- D3 (리더 결정, 되돌릴 수 있음, 범주 guarded, 기한 없음): 호스트 메모리(Claude 자동 메모리, Codex `~/.codex/memories`) 읽기는 기본 꺼짐이고 프로젝트별 옵트인이에요. 실제 HOME 을 읽는 일이라 켜는 것은 사용자의 선택이에요. 이번 변경은 호스트 메모리를 전혀 읽지 않아요(REQ-5).
- D4 (제안, 되돌릴 수 있음, 범주 none): D2 의 이름 규칙을 이 저장소의 명령 규칙에 맞춰 평평한 이름으로 써요: `improve-signals`, `improve-plan`, `improve-eval`, `improve-apply`, `improve-revert`, `improve-note`. 이유: `src/ops/<word>.rs` 한 파일이 하위 명령이 되는 구조라 중첩 하위 명령(`moai improve signals`)은 `src/main.rs` 의 명령 정의를 바꿔야 해요. 사용자에게 보이는 이름은 스킬 루트 `/moai improve`(Codex `$moai improve`)로 같아요. 대안: 중첩 하위 명령. 기본값: 평평한 이름. 되돌리기: 이름만 바꾸면 돼요(아직 배포 전).
- D5 (제안, 되돌릴 수 있음, 범주 none): 부류 키 규칙(결정적, LLM 없음). 감사 노트: 첫 줄이 `AUDIT ` 로 시작하고 `verdict: FAIL` 을 담은 노트(카드 노트 모양 `AUDIT C-<n> … — verdict: FAIL`, PR 댓글 모양 `AUDIT #<n> …` 도 받아요)만 신호이고, 본문에서 `[A-Za-z0-9_./-]+\.(rs|md|py|json|yml|yaml|toml|sh|txt)` 모양의 경로 토큰(뒤의 `:줄` 은 뗌)마다 `audit_fail:<경로>` 한 신호, 경로가 없으면 `audit_fail:uncited`. 같은 노트가 같은 경로를 두 번 말해도 한 번이에요. 차단 노트: 첫 줄에 `LANE-BLOCKED` 가 있고 `kind=<단어>` 가 있으면 `block:<kind>`(없으면 `block:unknown`). 플레이키: 제목이 `flaky: ` 로 시작하는 카드의 시험 이름(제목에서 접두사 뒤 첫 공백 앞)으로 `flake:<test>`. 메모: `improve/notes.jsonl` 줄의 `kind`·`key` 로 `<kind>:<key>`. 대안: 감사 본문을 LLM 으로 분류(결정적이 아니라 거절). 기본값: 경로 토큰.
- D6 (제안, 되돌릴 수 있음, 범주 none): 플레이키 횟수 규칙. 플레이키 기록은 시험마다 카드 하나(`verify.md:53`)라, 횟수는 "그 카드 1개 + 그 카드에 달린 `observation`·`report` 종류 노트 수"로 세요(재발은 노트로 남는다는 병합 열차의 관행). 노트의 형식을 정한 문서가 없어 이 규칙은 근사값이에요. 대안: 카드 수만 셈(한 시험은 항상 1). 기본값: 위 규칙. 되돌리기: 조각 S2 가 `improve-note` 로 정확한 재발을 남기면 이 규칙을 대체해요.
- D7 (제안, 되돌릴 수 있음, 범주 none): `improve/notes.jsonl` 형식. 줄마다 JSON 한 개: `{"v":1,"at":"<UTC RFC 3339>","kind":"correction|failure|false_claim","key":"<주제 또는 명령, 80자 이하>","text":"<200자 이하>","by":"<관찰한 쪽>"}`. 읽을 때 알 수 없는 `kind`·깨진 줄은 `skipped`(`bad_line`)로 세고 멈추지 않아요. 위치는 상태 폴더 아래 `improve/notes.jsonl`(`state_root::path(root, "improve/notes.jsonl")`, 철자를 직접 쓰지 않음), 관리형 `.gitignore` 블록이 이미 상태 폴더의 나머지를 무시해요. 이 파일을 쓰는 명령(`improve-note`)은 조각 S2 예요.
- D8 (제안, 되돌릴 수 있음, 범주 none): 점검 실패의 종료 코드는 저장된 곳이 없어서(`src/verify.rs:1-5`, `check-affected` 로그는 워크트리 `target/` 아래) 조각 S1 은 읽지 않아요. 같은 정보는 `improve-note` 의 `failure` 종류로만 들어와요(조각 S2). 대안: `target/check-affected-*/` 로그를 읽기(워크트리마다 달라 불안정). 기본값: 읽지 않음.
- D9 (제안, 되돌릴 수 있음, 범주 none): 신호의 두 번 규칙(보고서 P9)은 조각 S1 에서 고정 2 예요(경계 N=1/N=2). 프로필로 바꾸는 것은 조각 S4 의 속도 제한과 함께.
- D10 (제안, 되돌릴 수 있음, 범주 none): 호스트별 차이(보고서 §3.8)를 고정해요. 지침 로딩: Claude Code 는 CLAUDE.md 가 있으면 AGENTS.md 를 읽지 않으니 `@AGENTS.md` import 를 둔 CLAUDE.md 를 유지, Codex 는 루트부터 현재 폴더까지 이어 붙이고 32 KiB 한도. 개인 파일: 둘 다 SKILL.md 가 `AGENTS.local.md` 를 명시적으로 읽어요. 프로젝트 스킬: `.claude/skills/prj-*`(파일 감시로 즉시 반영) / `.agents/skills/prj-*`(안 보이면 재시작 → 재생은 새 `codex exec` 프로세스). 에이전트: `.claude/agents/prj-*.md`(재생은 Agent 도구 + `isolation: worktree`, `disallowedTools` 로 쓰기 막기) / `.codex/agents/prj-*.toml`(세 키만, 에이전트 파일이 sandbox 를 못 덮으니 부모 세션 sandbox 를 읽기 전용으로). 강제 규칙: 지침은 "맥락이지 강제 설정이 아님"이라 반복 위반은 훅·테스트 승격을 제안해요(등급 D, `guarded` 라 사용자 결정).

### 근거(보고서 인용, 읽은 날 2026-11-20)

- 루프와 평가: Darwin Gödel Machine(arXiv 2505.22954, §1·§3·§4.2·§5), ADAS(2408.08435), SICA(2504.15228), OPRO(2309.03409), GEPA(2507.19457).
- 반성과 신호: Reflexion(2303.11366), Self-Refine(2303.17651), Huang et al.(2310.01798, 외부 피드백 없는 자기 교정 실패), ExpeL(2308.10144), Agent Workflow Memory(2409.07429).
- 맥락·지침 관리: ACE(arXiv 2510.04618v3, 델타·결정적 병합·맥락 붕괴), Dynamic Cheatsheet(2504.07952), Anthropic "Effective context engineering"(주의 예산), Claude Code Memory 문서(CLAUDE.md 200줄 아래, 팀 공유와 개인 파일), Codex AGENTS.md 문서(32 KiB).
- 스킬·에이전트 공식 문서: Claude Code Skills·Subagents, Anthropic Agent Skills 모범 사례("평가를 먼저 만들어라", 제안자와 평가자 분리), Codex Build skills·Subagents·Memories.
- 보상 해킹과 안전: Sakana "DGM and AI Safety"(탐지 표식 삭제·로그 위조), METR(2025-06-05), ImpossibleBench(2510.20270), Sycophancy to Subterfuge(2406.10162), OpenAI CoT 모니터링(2503.11926), STOP(2310.02304).
- 확인하지 못한 것(보고서 부록 B): DGM 부록 H 원문, Claude 자동 메모리 저장 경로, "검색" 표시 행의 수치(초록 기준).

## Risks

- 감사 노트의 경로 토큰은 거친 신호예요. 다른 문장 속 경로(예: 예시로 든 파일)도 세어 부풀릴 수 있어서, 이 명령은 부류와 횟수와 예시 다섯 개만 보여 주고 아무것도 제안하지 않아요(제안은 S2, 사람이 선택). 노이즈가 크면 D5 의 규칙을 바꾸는 것이 되돌리기예요.
- 판정 줄 형식이 바뀌면 신호가 조용히 줄어요. 맞는 줄은 `skipped` 에 `no_rule` 로 세어져 보이고, 형식의 출처(`skills/moai/workflows/audit.md`, `AGENTS.md`)를 운영 문서 조각에 적어요.
- 플레이키 횟수(D6)는 근사값이에요. 이 근사값이 "두 번째 플레이키"를 잘못 만들 수 있으나 이 명령은 읽기 전용이라 해가 없고, S2 에서 정확한 기록으로 바꿔요.
- 노트 본문에는 개인 정보나 비밀이 있을 수 있어요. 출력은 120자·가림을 거치고(REQ-6), 파일에 저장하지 않으며(REQ-4), 네트워크로 보내지 않아요(비목표).
- 큰 장부는 느려요. 최근 20,000 개·본문 8,192 바이트 상한(REQ-7)과 경계 테스트로 막아요.
- 호스트 메모리는 실제 HOME 읽기라 이번 변경은 읽지 않고(REQ-5, FIFO 테스트로 증명), 켜는 것은 D3 에 따라 사용자 옵트인 뒤 별도 조각에서 해요.
- 루프 전체의 위험(보고서 §5): 평가기 속이기, 체크리스트 변조, 자기 강화 오류, 맥락 붕괴, 지침 비대화, 지침은 강제가 아님, 감독 신호 최적화, 경로 의존성·비용 폭주, 팀 지침 충돌, 비밀 유출, 외부 스킬 공급망, 메모리를 규칙으로 착각, 샌드박스 탈출. 각 장치는 REQ-8~REQ-15 에 들어 있고, 구현하는 조각이 자기 수용 기준으로 증명해요.

## Result

조각 S1(`improve-signals`)의 결과예요. REQ-12~REQ-25 는 뒤 조각의 몫이라 수용 줄이 아직 없어요.

- AC-1: met — head 70bb7e360c 에서 `cargo test --offline --locked --test improve_signals_cli` exit 0 (11 passed: 네 출처의 부류·횟수·시각·예시, 없는 장부·메모 파일은 `absent`, 심볼릭 링크·이름 붙은 파이프·깨진 장부는 거절). 고치기 전(명령 없음)에는 같은 명령이 exit 101 (11 failed, `unrecognized subcommand 'improve-signals'`).
- AC-2: met — 같은 실행의 `one_signal_is_observed_once_and_two_are_candidates_for_every_source`, 단위 `improve_signals_status_flips_at_two` 통과
- AC-3: met — 같은 실행의 `an_agents_own_claim_is_never_a_signal`(일곱 줄이 `skipped.no_rule` 7, `false_claim` 메모만 셈)와 `an_audit_note_cites_each_path_once_and_a_note_without_a_path_is_uncited` 통과
- AC-4: met — 같은 실행의 `nothing_is_written_with_or_without_a_ledger_and_an_absent_source_is_reported`(TempHome 스냅샷 전후 같음, 장부·메모 파일이 없을 때 폴더·파일이 생기지 않음) 통과
- AC-5: met — 같은 실행의 `host_memory_is_never_opened`(Claude 자동 메모리·Codex memories 자리에 이름 붙은 파이프를 두고 30 s 한도 안에 종료 0, `host_memory: "off"`) 통과
- AC-6: met — 같은 실행의 `texts_are_masked_and_cut_to_120_characters` 와 단위 `improve_signals_masks_then_cuts_and_never_keys_a_secret` 통과
- AC-7: met — `cargo test --offline --locked --bin moai-cli improve_signals` exit 0 (7 passed: N 개는 `truncated: false`, N+1 개는 N 개만 읽고 `true`; 본문 8,192 바이트는 그대로, 8,193 은 잘림)
- AC-9: met — 변형 5개를 하나씩 적용해 `cargo test --offline --locked --no-fail-fast --test improve_signals_cli --bin moai-cli` 를 돌렸고 모두 빨갛게 바뀐 뒤 파일을 복원했어요(`git status` 깨끗).

  | 변형 | 바꾼 곳 | 잡은 시험 |
  | --- | --- | --- |
  | threshold-3 | `THRESHOLD` 2 → 3 | `improve_signals_status_flips_at_two`, `every_source_becomes_a_class_with_count_times_and_examples` |
  | examples-6 | `EXAMPLES` 5 → 6 | `a_class_shows_at_most_five_examples_and_the_newest_ones` |
  | truncated-ge | `notes.len() > limit` → `>=` | `improve_signals_reads_the_newest_n_notes_and_says_when_it_cut` |
  | audit-no-prefix | `AUDIT ` 접두사 조건 제거 | `improve_signals_reads_the_audit_and_blocked_lines_only`, `an_agents_own_claim_is_never_a_signal` |
  | fifo-allowed | `!meta.is_file()` 거절 제거 | `a_broken_ledger_and_an_unsafe_notes_file_are_refused` |

  처음 계획한 다섯째 변형(심볼릭 링크 검사 제거)은 다음 검사(`!is_file`)가 링크도 거절해서 겉으로 같은 동작이었어요(동등 변형). 그래서 위 표로 바꿨어요. 모든 변형 실행에서 도움말 시험 둘(`help_table_names_every_command_once`, `every_command_has_examples`)도 빨갰는데 이는 도움말 행이 아직 없어서 원래 실패하는 것이라 위 표의 근거로 세지 않았어요.
