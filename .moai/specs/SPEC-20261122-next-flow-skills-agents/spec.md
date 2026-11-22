---
id: SPEC-20261122-next-flow-skills-agents
title: "The moai skill has eight routes and /moai work, one flow reference and one audit contract, and the update removes what the package no longer ships"
status: draft
revision: 1
created: 2026-11-22
parent: null
---

# The NeXT-Flow skill package: routes, one flow reference, one audit contract

## Intent

사용자 결정(2026-11-22)에 따라 설정 폴더에 설치하는 `/moai` 스킬 묶음을 NeXT-Flow 에 맞춰 줄여요. 이 SPEC 은 스킬 글만 다루고, 기본 에이전트와 렌더러는 자식 SPEC `SPEC-20261122-next-flow-base-agents`, 갱신의 낡은 파일 정리와 실행기 역할 플래그는 자식 SPEC `SPEC-20261122-next-flow-update-and-launcher` 가 맡아요. 이 조각은 SPEC 만 쓰고 코드는 없어요.

바뀌는 것은 다섯 가지예요. (1) 경로(route)가 READ·PLAN·RUN·AUDIT·SYNC·PROJECT·UPDATE·WORK 여덟 개로 줄어요: SETUP 은 PROJECT 로, VERIFY 는 RUN 으로, OPTIMIZE 는 UPDATE 의 미리보기 단계로 흡수되고, 리더·에이전트 두 모드 글(합쳐 105 KB)과 `--mode-*` 플래그는 없어져요. (2) 진입은 `/moai` 와 `/moai work` 둘이에요. WORK 는 카드를 받아 PLAN → RUN → SYNC → AUDIT → 착지 명령까지 돌고 카드가 없을 때까지 되풀이해요. (3) 흐름 글은 `references/flows/next-flow.md` 한 파일이에요(약 5 KB; 네 글자의 뜻을 설명하는 출하 글은 이 한 곳뿐). (4) 감사 계약은 `references/audit.md` 한 파일이에요(약 5 KB; `review.md` 를 흡수). 받는 것, 돌려주는 것, 누가 판정을 적는가를 정하고, 로컬 `/code-review <base>..<head>` 를 증거로 허용하되 `--fix`·`--comment` 는 막고, "줄이 없으면 판정 없음"을 적어요. (5) 새 결정과 부딪히는 내부 참조를 고치고, Claude Code 에서 없어진 도구 이름을 글에서 지워요.

이 묶음은 설정 폴더(Claude 설정 폴더마다 한 벌)에 설치돼요(사용자 결정 2026-11-22). 설치 자리와 순서는 이 SPEC 의 범위가 아니에요(전역 설치 조각이 맡아요). 사용자의 세 언어 질문(대화·문서·커밋)은 설치 때 한 번 묻고 사용자 범위에 저장하며 프로젝트는 그 값을 쓴다는 결정(2026-11-22)은 PROJECT 글에 "이미 사용자 범위에 있으면 다시 묻지 않는다"로 들어가요(REQ-10); 저장 코드는 전역 설치 조각의 일이에요.

### What exists (tip `78c8a5a247`, 잰 시각 2026-11-22T07:15Z)

- 스킬 글(`find skills agents -type f | xargs wc -c`): 파일 전체 485,958 바이트 중 스킬이 467,115, 기본 에이전트 다섯이 18,843. 워크플로 12개가 251,319 바이트(mode-agent 56,705, mode-leader 48,464, project 35,253, update 26,884, verify 26,307, audit 19,859, sync 15,981 …).
- 경로표 `skills/moai/references/routes.json`: READ, SETUP, PROJECT, UPDATE, PLAN, RUN, VERIFY, AUDIT, SYNC, OPTIMIZE, MODE-LEADER, MODE-AGENT 열두 개. WORK 는 없어요. `SKILL.md`(6,081 바이트)는 `AGENTS.md` 를 읽으라 하고 `--mode-agent`·`--mode-leader` 를 모드 철자로 못 박아요.
- 체인표 `references/chains/index.md`: 마지막 행이 `Keep working toward a goal | MODE-AGENT`, 규칙 한 줄이 MODE-AGENT 를 가리켜요. `Review a pull request | AUDIT | internal/review`.
- 모드 글 이름을 담은 파일: `mode-agent`·`mode-leader` 문자열을 담은 파일이 tests·src·skills·agents 에 61곳(`grep -rlE 'mode-agent|mode-leader'`, 그중 tests/*.rs 26, src 11), docs 에 14곳. 워크플로 파일 `setup.md`·`verify.md`·`optimize.md` 를 경로로 부르는 파일이 24곳. 규칙 문장을 고정하는 시험이 많아요(보고서 기준 mode-agent 규칙 60줄 중 57줄).
- 이름이 새 결정과 부딪히는 내부 참조 열한 개(보고서 8-2, 2차 읽기 조사): build-folder, commit-trailers(Codex 변수, PR 감사 관문, 리더 리뷰), decide("리더에게 묻는다"), git-safety(모드, 홈 작업트리, 끝 병합, Codex 행), interview, quality, questions(SETUP, Codex 한도), review(PR CI 인용), sdd("SPEC = PR 하나"), self-check(이 저장소 도구, Python 슬롯, PR 본문), versioning(자율 단계, 통합 브랜치 이름). 부딪힘 없는 여덟과 덧붙이기만 하나는 그대로예요.
- 설치본과의 간격: 설치된 스킬은 51개 파일 중 27개가 저장소와 다르고 기본 에이전트 둘이 달라요(보고서 2-3, 실측 2026-11-22). 이 간격은 갱신이 메워요(자식 SPEC).
- 갱신의 낡은 파일: 패키지에서 빠진 파일을 `remove_blocked` 로 보고만 하고 지우지 않아요(`src/update.rs:1295-1318`, `src/agent_install.rs:259-275` 를 읽음).

## Non-goals

- 설치 자리·순서·설정 폴더 목록: 전역 설치 조각의 일. 이 SPEC 은 패키지 안의 글을 정해요.
- 기본 에이전트 정의와 렌더러: 자식 SPEC. 갱신 코드와 실행기 플래그: 자식 SPEC.
- 레인 띄우기·중단·`/goal` 조건의 글과 명령(`moai lane`): 레인 운영 SPEC 들의 일이에요. WORK 글은 그 명령 이름을 부를 뿐 정의하지 않아요.
- 규칙 모델(`rules/moai/flow.md`, `rules/moai.md`)의 글: 설정 폴더 규칙 SPEC 이 정해요. 이 SPEC 의 흐름 글은 스킬 안의 참조예요.
- 감사 실행기(세션 띄우기·헤드리스 호출)의 방법: 감사 계약은 받는 것과 돌려주는 것만 정해요.
- 이 저장소의 프로젝트 하네스 파일(`prj-reviewer`, `prj-local-check`)을 지우거나 얇게 하는 일: 추적되지 않는 사용자 파일이라 SPEC 의 산출물이 아니에요. 출하 글이 그 이름을 부르지 않게만 해요(REQ-8).

## Requirements

- REQ-1: The route table shall name exactly the routes READ, PLAN, RUN, AUDIT, SYNC, PROJECT, UPDATE and WORK, each with a reference file that the package ships.
- REQ-2: The skill package shall hold none of the files workflows/mode-agent.md, workflows/mode-leader.md, workflows/setup.md, workflows/verify.md, workflows/optimize.md, references/adapters/herdr.md, references/internal/build-folder.md, references/internal/review.md and references/rule-history.md.
- REQ-3: When a request after the skill name starts with the word work, the skill shall pick the WORK route, which takes one card, plans it, runs it, syncs it, audits it and lands it with the landing command, and repeats until no card is left. · failure: none — a request without that word picks its route from the route table, and a card that needs a user decision ends in a blocked note on the card, never in a question to the user.
- REQ-4: The skill shall treat --mode-agent, --mode-leader, -a, -l and --autonomy in a request as plain text that picks no route.
- REQ-5: The package shall carry exactly one flow reference, references/flows/next-flow.md of at most 6144 bytes, that states the worktree place, that a lane submits its head and base only, the exactly-once landing command, the protection hook, the fast-forward push after landing and that remote checks never block landing, and that explains the four letters of the flow name once.
- REQ-6: The package shall carry exactly one audit contract, references/audit.md of at most 6144 bytes, that lists what the auditor receives, what it returns, which session records the verdict and that the author of a change never audits it.
- REQ-7: The audit contract shall state five side rules: the local code review of the base..head range is allowed evidence without its fix or comment options, a re-audit goes to the first auditor, a lane's tree-bound check receipt may be cited, no forked or advisor session audits, and a finding that names no file and no line counts as no finding.
- REQ-8: No shipped skill text shall name Codex, OMP, a resident leader, a mode flag, a pull-request audit gate, a merge train, AGENTS.md as a file moai writes, prj-reviewer or prj-local-check.
- REQ-9: No shipped skill or agent text shall name a tool or option Claude Code no longer has (TodoWrite, TaskCreate, TaskOutput, BashOutput, the resume and mode inputs of the Agent tool, the agents wizard), assume the Grep or Glob tool exists, hold a shell-injection command block, or set a model key in a skill file.
- REQ-10: The PROJECT and UPDATE route texts shall absorb the retired SETUP and OPTIMIZE routes: PROJECT carries the steps of SETUP with the local tool language question and asks no conversation, document or commit language that the user scope already holds, and UPDATE carries a preview of leftover space that deletes nothing.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4): the route table holds the eight routes and each reference file exists, the nine files are absent from the package, and SKILL.md names no mode flag as a route.
  verify: `cargo test --locked --offline --test next_flow_skill_scan routes_`
- AC-2 (REQ-3): the WORK reference lists, in order, take a card, plan, run, sync, audit, land and repeat, and says that a decision goes to the card as a blocked note.
  verify: `cargo test --locked --offline --test next_flow_skill_scan work_`
- AC-3 (REQ-5, REQ-6, REQ-7): the flow reference and the audit contract exist, stay within their byte budget and hold each stated item; exactly one shipped text explains the four letters.
  verify: `cargo test --locked --offline --test next_flow_skill_scan refs_`
- AC-4 (REQ-8, REQ-9): the scan finds none of the named words or tools in the shipped skill and agent texts, and planting one in a copy fails it.
  verify: `cargo test --locked --offline --test next_flow_skill_scan scrub_`
- AC-5 (REQ-10): the PROJECT text holds the SETUP steps and the language rule, and the UPDATE text holds the preview step that says it deletes nothing.
  verify: `cargo test --locked --offline --test next_flow_skill_scan absorb_`
- AC-6: every older test that read a removed file or a removed rule sentence is moved to the sentence's new place or retired with a written reason, and the neighbouring checks pass.
  verify: `cargo test --locked --offline --test agent_defs_package`
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- this slice, write: `.moai/specs/SPEC-20261122-next-flow-skills-agents/**`, `.moai/specs/SPEC-20261122-next-flow-base-agents/**`, `.moai/specs/SPEC-20261122-next-flow-update-and-launcher/**`
- later slices, write: `skills/moai/**` (texts, `references/routes.json`, `references/chains/index.md`, `references/source-manifest.json`), `docs/**` (the moved rule history), `tests/next_flow_skill_scan.rs`, and every test that reads a removed file or a removed sentence (the inventory is the first step of the text slice)
- read-only: `src/update.rs`, `src/agent_install.rs`, `tests/style_banner_scan.rs`
- forbidden: `AGENTS.md`, `AGENTS.local.md` of this repository, the real HOME, the network, the integration branch, `.claude/` of this checkout

## Plan

1. (done) The three SPECs (this slice).
2. Inventory: list every test and source line that names a removed file or a removed rule sentence (61 files name a mode file, 24 name `setup`, `verify` or `optimize`), and for each decide: the rule still matters → its sentence moves to the new place and the pin follows; the rule retires with the route → the test is retired with its reason written where it stood.
3. Text slice A: routes and WORK: `routes.json`, `SKILL.md`, `chains/index.md`, the WORK reference, PROJECT absorbs SETUP, RUN absorbs VERIFY, UPDATE absorbs OPTIMIZE; the removed files go; `next_flow_skill_scan.rs` tests `routes_`, `work_`, `absorb_` are written first and fail first.
4. Text slice B: the flow reference and the audit contract (the review text moves in); `refs_` tests first.
5. Text slice C: the scrub of the internal references that named the old flow (commit-trailers, decide, git-safety, interview, quality, questions, self-check, sdd, versioning; build-folder and review leave the package in slices A and B), of the workflow texts and SKILL.md, and of the stale tool names; `scrub_` tests first; the source manifest is regenerated.
6. Rule history moves from the package to `docs/`, and the guard test that pinned it follows.

## Risks

- The mode files carry many pinned sentences. Retiring a rule with its route is right only when the rule belongs to the retired route; the inventory step decides each one and writes the reason, so a rule that still matters (for example a safety line) moves instead of vanishing.
- The flow reference and the flow-name rename slice both touch the single text that explains the four letters. Whoever lands second adjusts the other's test to name `references/flows/next-flow.md` as the one explaining text.
- WORK needs the lane-start and `/goal` texts of the lane operation SPECs. This SPEC writes WORK so that it calls those commands by name and keeps working when they are absent (it then runs in the current session).
- Three unverified flow lines: the protection hook, the fast-forward push after landing and "remote checks never block landing" are decisions in the notes of the flow cards; the reference states them as the decided rules and the slice checks each against the landing code before it ships.
- Measured sizes of the flow reference and the audit contract are estimates (about 5 KB each); the 6,144-byte limit is a ceiling the test enforces, not a prediction.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The eight route tokens are READ, PLAN, RUN, AUDIT, SYNC, PROJECT, UPDATE, WORK. | The whole-docs report 4-3, the user's pick of the entry design. | Add a route. |
| D-2 | The flow reference lives at `references/flows/next-flow.md` and explains the four letters; the rule file in the config folder stays a short operational git rule. | The flow-name SPEC allows exactly one explaining text. | Move the letters. |
| D-3 | `references/rule-history.md` moves to `docs/`. | The report 4-3 and the coverage note: nothing reads it while a route runs. | Keep it in the package. |
| D-4 | `review.md` is absorbed into the audit contract and removed. | The report 4-3: one audit contract, about 5 KB. | Keep both. |
| D-5 | A re-audit goes to the first auditor, a lane's receipt may be cited, no fork or advisor audits. | The coverage note on the audit side rules. | Loosen per rule. |
| D-6 | The audit contract fixes what the auditor receives and returns, not how the session is started. | Starting audit sessions is the lane and headless SPECs' work. | Fold it in later. |
| D-7 | The shipped texts do not name `prj-reviewer` or `prj-local-check`. | They are this repository's own harness files; a product text must not depend on them. | Name them again. |

## Result

(Filled when the work is done. This is the SPEC-only slice: no code and no text changed yet.)
