---
id: SPEC-20261125-base-agent-contract
title: "The base agent contract hands implementation over inside the lane worktree, gives the audit agent one tier table and leaves no unused read"
status: approved
revision: 1
created: 2026-11-25
approval:
  by: user
  at: 2026-11-25
  basis: "사용자가 리더 대화의 선택지에서 직접 고른 답(카드 노트 9243): \"승인, D-1~D-7 기본값 A (Recommended)\"; 리더가 묻기 직전과 기록 직전에 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:664a91146c207c53525353520f263b8d4c8386413bf9cf0e7590a87201161e87"
---

# The base agent contract hands implementation over inside the lane worktree, gives the audit agent one tier table and leaves no unused read

## Interview

- Q1: 스킬 지침을 어떻게 손볼까요? → A: 사용자 지시 그대로(스킬 전수 조사, 2026-11-25): "moai 스킬 지침 전체 전수 조사해서 검색 말고 전체 개별로 읽고 개선점을 찾아서 모두 업데이트 개선 보고 후 진행". 이 SPEC 은 그 조각 중 "기본 에이전트 계약"(보고서 `~/.moai/reports/moai-adk/skill-survey-20261125.md` 의 고치는 순서 9)이에요.
- Q2: 구현 에이전트는 어디서 일하나요? → A: 레인의 카드 워크트리에서요(사용자 결정 2026-11-25, 조사 보고서 "오늘 결정과 어긋나는 글"). `agents/impl.md` 의 `isolation: none` 과 "부모가 일하는 워크트리" 문장은 이미 그렇게 고쳐져 있어요(커밋 `0d6ff2cf41`). 남은 일은 워크트리·범위·generation 을 넘기는 입력과 변경을 돌려주는 길이에요.
- Q3: 감사 관문(FAIL 네 가지, 작성자는 감사하지 않음, 판정은 감사 세션이 기록)을 바꾸나요? → A: 아니요. `agents/audit.md` 는 감사 관문 경로라서 tier 2 이고, 이 SPEC 은 글이 감사 계약(`references/audit.md`)과 어긋난 곳만 맞춰요. 승인은 사용자만 해요.

## Intent

기본 에이전트 글 다섯과 `chains/index.md` 를 실제로 읽은 사실(2026-11-25, 기준 NeXT `4ddb7c9c31`):

- `agents/impl.md` Input 은 `task`, `scope`, `base`, `refs` 뿐이에요. 부모가 일하는 워크트리의 절대 경로, 카드와 generation 을 넘기는 줄이 없고, 구현이 끝난 뒤 변경(커밋하지 않은 파일)을 부모가 어떻게 받는지 글이 말하지 않아요. `workflows/work.md` 는 "에이전트에게 `moai card check` 가 허용한 경로만 주라"고 하지만 에이전트 글에는 그 입력이 없어요.
- `agents/audit.md` 의 tier 네 줄(45-48행)은 `references/audit.md` 의 말을 그대로 옮긴 글인데, 한 표로 "무엇을 돌리고 무엇을 인용하는지"를 보여 주지 않아요. 그리고 같은 글 안에서 어긋나요: 59행 "모든 기준에 내가 돌린 검사가 있거나 `Not run` 에 있다"와 79행 "모든 검사를 돌렸을 때 `ok`" 는 light audit(아무것도 돌리지 않음, 45-46행)과 맞지 않아요. 47행의 "아래 floor 목록"은 이 파일에 없어요(목록은 `references/audit.md` 의 "Audit tiers"). 52행은 내부 SPEC 이름(`SPEC-20261119-check-guard`)을 남겨요. 44행 docs-only 줄은 "lander 가 읽고 착지한다"로 남아 있는데, 감사 계약은 이미 "감사자가 영수증 없이 PASS 를 기록한다"로 고쳐졌어요(`SPEC-20261125-audit-text-facts`).
- `agents/docs.md` Input 에 `head` 가 없어요. `base` 만 있고 "변경이 있는 커밋"이라고만 적어서, 문서 에이전트가 어느 커밋의 변경을 문서화하는지 확인하는 줄이 없어요(impl 은 `base` 가 HEAD 와 같은지 확인해요).
- `reads:` 목록의 몇 항목은 에이전트 본문이 부르지 않아요: `moai-plan` 의 `internal/interview`(PROJECT 경로의 질문 은행), `moai-research` 의 `adapters/aside`·`adapters/orca`(브라우저 도구, 본문은 호스트의 웹 도구만 말해요). 생성되는 "Read first" 줄이 에이전트에게 읽으라고 시키므로 쓰지 않는 항목은 매번 읽는 비용이에요.
- `chains/index.md:19` "Repair after a FAIL verdict" 줄은 `moai-impl, moai-audit` 뿐이에요. 같은 파일의 규칙("SYNC writes the SPEC Result, documents … on the same branch")과 `work.md` 의 FAIL 뒤 절차(`card reopen` → 같은 브랜치에서 수리 → 6단계부터 다시)는 문서 갱신을 포함하는데 `moai-docs` 가 줄에 없어요.

이 SPEC 이 하는 일: impl 계약에 워크트리·카드·generation 입력과 변경을 돌려주는 길을 쓰고, audit 계약에 tier 별 한 표를 두고 안의 모순 셋을 고치고, docs 에 head 입력을 넣고, 쓰지 않는 reads 를 지우거나 본문이 부르게 하고, 수리 줄에 `moai-docs` 를 넣어요.

## Non-goals

- 관문을 느슨하게 하거나 바꾸지 않아요: FAIL 네 가지, 작성자는 감사하지 않음, 판정은 감사 세션이 기록, tier floor 목록(코드 `src/gate.rs` 가 읽는 목록)은 그대로예요.
- `skills/moai/workflows/*.md`, `skills/moai/references/audit.md` 와 Rust 코드는 건드리지 않아요. 이 SPEC 의 글은 그 글들의 현재 내용을 따라가요.
- 에이전트의 출력 첫 줄(`RESULT <role> <ok|blocked|not_run> <target>`)의 모양은 바꾸지 않아요.
- 에이전트가 다른 에이전트를 시작하지 않는다는 규칙과 `moai-impl` 의 `isolation: none` 은 그대로예요.

## Requirements

<!-- This SPEC edits the audit agent's text (tier 2: a path of the audit gate) and keeps every blocking rule it states; no check that blocks is added or loosened. -->
- REQ-1: The `## Input` of `agents/impl.md` shall name `worktree` (the absolute path of the folder the parent works in), `card` and `generation` beside `task`, `scope`, `base` and `refs`, and say that the parent gives only paths `moai card check` allowed. · failure: none — a missing input is found by the contract test.
- REQ-2: IF `worktree` is not the folder `git rev-parse --show-toplevel` prints, or `base` is not `git rev-parse HEAD`, THEN `moai-impl` shall end with `RESULT impl blocked` before it writes. · failure: `blocked` — nothing is written.
- REQ-3: When `moai-impl` is about to write a path for the first time, it shall run `moai card check <card> --generation <generation> --path <path>` and take exit 0 as its permission to write. · failure: `blocked` — REQ-13 says how it ends.
- REQ-4: `agents/impl.md` shall say that the change stays as uncommitted files in the parent's worktree, that `RESULT impl` names the HEAD the change sits on, and that `Files written` plus the output of `git status --short` are what the parent reads before it runs its own `moai card check` and commits. · failure: none — the parent re-reads the tree.
- REQ-5: `agents/audit.md` shall hold one table with a row per tier line (docs-only, test-only, code, high-risk) whose cells say what the auditor runs, what it cites and what the verdict then needs, in the words of the audit contract's "Audit tiers". · failure: none — the contract test finds a missing row or cell.
- REQ-6: `agents/audit.md` shall give the floor list's place as the audit contract's section "Audit tiers" instead of "the list below". · failure: none — the contract test finds the phrase.
- REQ-7: The `## Verification` and `## Handoff` of `agents/audit.md` shall ask for the checks the tier names, not for every check; a light audit that runs nothing returns `ok` with `Not run` naming what its tier leaves out. · failure: none — the contract test finds the old sentence.
- REQ-8: `agents/docs.md` shall name `head` (the 40-hex commit that holds the change) in its Input. · failure: none — the contract test finds a missing input.
- REQ-9: Every `reads:` entry of the five agents shall be named by that agent's body at the step that uses it, or be removed from `reads:`. · failure: none — the contract test finds an entry the body does not name.
- REQ-10: The `Repair after a FAIL verdict` row of `chains/index.md` shall list `moai-impl, moai-docs, moai-audit` under the routes `RUN, SYNC, AUDIT`. · failure: none — the existing chains test reads the row.
- REQ-11: The five agent files and `chains/index.md` shall name no Issue number, no card number, no internal SPEC name and no retired host. · failure: none — a leftover is a non-blocking finding.
- REQ-12: A test that pins a sentence these files lose shall pin the new sentence or be deleted with its reason in the commit message. · failure: none — an unexplained deletion is a finding.
- REQ-13 (REQ-3): IF `moai card check` does not exit 0 for a path, THEN `moai-impl` shall end with `RESULT impl blocked` naming the answer as it is, before it writes that path. · failure: `blocked` — the answer is reported as it is.
- REQ-14 (REQ-5): The four tier statements shall stand in `agents/audit.md` only as the rows of that table, not again as bullet lines. · failure: none — the contract test finds a second statement.
- REQ-15 (REQ-6): `agents/audit.md` shall name no internal SPEC. · failure: none — the contract test finds a `SPEC-` name.
- REQ-16 (REQ-8): IF `git rev-parse HEAD` differs from `head`, THEN `moai-docs` shall end with `RESULT docs blocked` before it writes. · failure: `blocked` — nothing is written.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-13): `agents/impl.md` Input names `worktree`, `card`, `generation`; Steps say the toplevel and `base` checks end in `blocked`, the `moai card check` call and its `blocked` answer; the text says the change stays uncommitted and what the parent reads · verify: `cargo test --offline --locked --test base_agent_contract ac1_`
- AC-2 (REQ-5, REQ-6, REQ-7, REQ-14, REQ-15): `agents/audit.md` holds one table with the four tier rows and their cells, no "list below", no `SPEC-`, no "every check" in Verification and Handoff · verify: `cargo test --offline --locked --test base_agent_contract ac2_`
- AC-3 (REQ-5): the table's rows state the same four facts as `references/audit.md` "Audit tiers" (what is built or run, what is cited, who audits, what tier 2 adds); removing the high-risk row's "guard-mutant lines" or "two named test modules" makes the test fail (guard mutant) · verify: `cargo test --offline --locked --test base_agent_contract ac3_`
- AC-4 (REQ-8, REQ-16): `agents/docs.md` Input names `head` and Steps say a differing HEAD ends in `blocked` · verify: `cargo test --offline --locked --test base_agent_contract ac4_`
- AC-5 (REQ-9): each agent's `reads:` entries all appear in its body by name, and `reads:` lists are the ones D-4 chooses · verify: `cargo test --offline --locked --test base_agent_contract ac5_`
- AC-6 (REQ-10): the repair row lists `moai-impl, moai-docs, moai-audit` and routes `RUN, SYNC, AUDIT` · verify: `cargo test --offline --locked --test base_agent_contract ac6_`
- AC-7 (REQ-11): the six files hold no `Issue #`, no `#` followed by four digits, no card number, no `SPEC-2`, no `Codex`, `OMP`, `Kiro` or `Cursor` · verify: `cargo test --offline --locked --test base_agent_contract ac7_`
- AC-8 (REQ-3): the one hand-run that proves the claim behind D-1: a `moai-impl` subagent started by the Agent tool in the card worktree runs `moai card check` for an allowed and a refused path and gets exit 0 and non-zero; the commands and exit codes go into the card note · verify: `moai card check <card> --generation <g> --path agents/impl.md` run by the subagent
- AC-9 (REQ-12): every pinned sentence that changed is pinned again or its test is gone with a reason · verify: `moai check affected --card <base>..<head>`

## Decisions

각 결정은 되돌릴 수 있고 분류는 none 이에요. 추천(A)이 맨 앞이고 사용자가 고르지 않으면 기본으로 A 를 써요.

- D-1 (impl 이 카드 검사를 스스로 하나): (A, 기본·권장) 입력에 `worktree`·`card`·`generation` 을 받고, 경로마다 첫 쓰기 전에 `moai card check` 를 읽기 전용으로 돌려요. 부모는 넘기기 전과 커밋 전에 다시 검사해요. 하위 에이전트의 세션 id 가 부모와 같다는 가정은 이 글을 쓰는 시점에 확인되지 않았어요(`card check` 는 "이 세션"이 예약을 쥔 때만 0 을 돌려줘요): AC-8 의 실제 실행이 가정을 닫고, 거절이 나오면 B 로 바꾸고 이유를 노트에 써요. (B) 부모만 검사하고 impl 은 `card` 와 `generation` 을 받지 않아요(경로만 받음).
- D-2 (변경을 돌려주는 길): (A, 기본·권장) 새 워크트리나 복사본 없이 부모 워크트리에 커밋하지 않은 파일로 남기고, 부모가 `git status --short` 와 `Files written` 을 대조한 뒤 자기 커밋을 만들어요. 에이전트는 작업이 시키지 않는 한 커밋하지 않아요(지금 글 그대로). (B) impl 이 커밋해서 head 를 돌려줘요 — 커밋 트레일러와 `card check` 가 레인 세션 몫이라 맞지 않아요.
- D-3 (audit 표의 열): (A, 기본·권장) 열은 tier · 바뀐 경로 · 돌리는 것 · 인용하는 것 · 판정에 필요한 것 다섯이에요. 칸의 말은 `references/audit.md` "Audit tiers" 에서 그대로 가져와요. (B) 네 줄을 그대로 두고 표 아래에 한 문단을 붙여요 — 모순 셋은 고치지만 한 표는 아니에요.
- D-4 (쓰지 않는 reads): (A, 기본·권장) 본문이 부르지 않는 항목을 `reads:` 에서 지워요: `moai-plan` 에서 `internal/interview`, `moai-research` 에서 `adapters/aside` 와 `adapters/orca`. 나머지 항목은 본문이 쓰는 단계에서 이름으로 불러요(impl: `test-design` 은 버그 수정의 테스트, `git-safety` 는 커밋 앞, `tdd-results` 는 실패·통과 기록, `self-check` 는 넘기기 앞; audit: `check-completeness`·`boundary-checks`·`security-checklist` 는 tier 2 의 단계; docs: `doc-claims` 는 사용자용 문서를 쓸 때, `sdd` 는 SPEC Result; plan·research: `code-atlas` 는 코드를 찾기 전). (B) 항목을 지우지 않고 본문에서 모두 불러요 — plan 이 PROJECT 질문 은행을 읽게 돼요.
- D-5 (docs 의 head): (A, 기본·권장) `base` 와 `head` 를 둘 다 받고 `head` 를 HEAD 와 대조해요. 변경 범위는 `base..head` 예요. (B) `base` 를 `head` 로 이름만 바꿔요.
- D-6 (수리 줄): (A, 기본·권장) `RUN, SYNC, AUDIT` 과 `moai-impl, moai-docs, moai-audit`. (B) 줄을 그대로 두고 규칙 문장만 더해요.
- D-7 (적용 순서): `agents/impl.md` 는 흐름 규칙 글을 코드와 맞추는 카드(지금 검토 중)의 범위에 있어요. 그 카드가 착지한 뒤에 이 SPEC 의 impl 부분을 적용해요. 그 전에는 SPEC 만 쓰고, `agents/audit.md`·`docs.md`·`plan.md`·`research.md`·`chains/index.md` 는 겹치지 않으므로 승인 뒤 바로 적용할 수 있어요.

## Sentence table

| # | 지금 위치와 문장(요약) | 새 위치 |
| --- | --- | --- |
| 1 | `audit.md:44` docs-only: "the lander reads the change and lands it, and nothing is built" | 표 1행: 감사자가 영수증 없이 PASS 를 기록하고 아무것도 빌드하지 않는다 |
| 2 | `audit.md:45` test-only: 읽고 영수증을 읽고 아무것도 돌리지 않는다 | 표 2행 |
| 3 | `audit.md:46` code: 영수증을 인용하고 head 를 다시 돌리지 않고 뮤턴트를 돌리지 않는다 | 표 3행 |
| 4 | `audit.md:47` high-risk: 코드 단계 + 경계·보안·데이터 손실 검사, 레인의 guard-mutant 줄을 읽고 이름 붙은 테스트 모듈 둘까지 | 표 4행 |
| 5 | `audit.md:42`, `:47` "the floor list below" | `references/audit.md` 의 "Audit tiers" 절을 이름으로 가리킨다 |
| 6 | `audit.md:52` `(SPEC-20261119-check-guard)` | 지운다(규칙 문장은 그대로: 풀 리퀘스트면 base 의 `REVIEW.md` 와 `agents/audit.md` 를 읽는다) |
| 7 | `audit.md:59` "Every acceptance criterion has a check you ran, or is listed under `Not run`" | "tier 가 이름 붙인 검사는 돌렸거나 `Not run` 에 이유와 함께 있다" |
| 8 | `audit.md:79` "`ok` when every check ran" | "`ok` when every check the tier names ran, and what the tier leaves out is under `Not run`" |
| 9 | `impl.md` Input 네 줄 | 일곱 줄: `task`, `scope`, `worktree`, `card`, `generation`, `base`, `refs` |
| 10 | `impl.md` Steps 1(워크트리·base 확인) | `worktree`·toplevel·`base` 확인 + 경로마다 `moai card check` |
| 11 | `docs.md` Input `base` "the 40-hex commit the change is on, when given" | `base` 와 `head`(변경이 있는 커밋, HEAD 와 대조) |
| 12 | `chains/index.md:19` `moai-impl, moai-audit` | `moai-impl, moai-docs, moai-audit` |

## Scope

- write: `.moai/specs/SPEC-20261125-base-agent-contract/**`, `agents/impl.md`, `agents/audit.md`, `agents/docs.md`, `agents/plan.md`, `agents/research.md`, `skills/moai/references/chains/index.md`, `tests/base_agent_contract.rs`
- write (고정 문장이 바뀔 때만, 읽어서 정확한 경로를 노트로 올려요): `tests/agent_defs_package.rs`, `tests/agent_delegation_text.rs`, `tests/base_agents_nested.rs`, `tests/plan_output_sections.rs`, `tests/agents_flat_ownership.rs`, `tests/core_rules_2876.rs`, `tests/next_flow_skill_scan.rs`, `tests/sdd_guidance.rs`
- write (한 줄, line-lease): `skills/moai/references/source-manifest.json`(`references/chains/index.md` 의 sha256), `tools/repo_scans.json`(새 시험 파일 등록), `tests/areas.toml`, `tests/fixtures/update_tux/*.json`(해시가 바뀔 때만)
- read-only: `skills/moai/references/audit.md`, `skills/moai/workflows/work.md`, `skills/moai/workflows/audit.md`, `src/agent_defs.rs`, `src/gate.rs`
- forbidden: 관문을 느슨하게 하는 변경, `references/audit.md` 와 `workflows/*.md` 의 수정, 전역 설치본과 실제 HOME, 네트워크, Python

## Plan

1. SPEC 승인을 기다려요(tier 2: `agents/audit.md` 는 floor 목록의 경로예요). 승인 전에는 글과 코드를 쓰지 않아요.
2. 범위를 넓혀요(`card scope`): 위 `write (고정 문장이 바뀔 때만)` 의 시험은 읽어서 필요한 것만이에요. `agents/impl.md` 는 그 카드 착지 뒤(D-7).
3. RED: `tests/base_agent_contract.rs`(AC-1..AC-7)가 지금 글에서 실패해요. `reads_lists_match_the_design…`(`agent_defs_package.rs`)의 기대 목록을 D-4 대로 바꿔요.
4. 글을 고쳐요(대조표대로). 표는 `references/audit.md` "Audit tiers" 와 말이 같아야 해요. GREEN.
5. AC-8: `moai-impl` 하위 에이전트로 `moai card check` 를 실제로 돌려요.
6. 가드 뮤턴트(AC-3: high-risk 행에서 "guard-mutant lines" 또는 "two named test modules" 를 지우면 죽는다), 매니페스트 해시, `cargo fmt --all`, `moai check affected --card`, READY.

## Risks

- R-1: 에이전트 글이 많은 시험에 고정돼 있어요(위 목록). 문장 하나가 빠지면 그 시험이 실패하므로 하나씩 새 문장에 다시 고정해요.
- R-2: D-1 의 가정(하위 에이전트의 세션 id 가 부모와 같다)이 틀리면 impl 이 `card check` 에서 거절돼요. AC-8 이 구현 전에 밝히고, 거절이면 B 로 바꿔요.
- R-3: audit 표를 쓰다가 tier 2 의 검사 줄이 빠지면 관문이 약해져요. AC-3 의 가드 뮤턴트가 그 줄이 빠지면 실패하게 해요.
- R-4: `references/audit.md` 의 tier 줄이 `SPEC-20261125-audit-text-facts` 로 바뀐 직후라서 그 SPEC 이 되돌려지면 표가 어긋나요. AC-3 이 두 글의 사실을 대조해요.
- R-5: `agents/impl.md` 가 다른 카드의 범위에 있어 같은 파일을 두 카드가 고치면 머지가 충돌해요. D-7 의 순서로 피해요.
