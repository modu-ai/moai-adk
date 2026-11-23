---
id: SPEC-20261124-moai-leader-route
title: "`/moai leader` is a ninth route: the window that types it registers as the coordination window, reads the ledger, lands audited heads and releases the registration when it ends"
status: approved
revision: 1
created: 2026-11-24
parent: SPEC-20261122-next-flow-skills-agents
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7838 (2026-11-23T19:53:19Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 7837 에 답함): 판 1 승인, D-1 부터 D-6 그대로; 사용자 선택 2026-11-22T23:49:36Z 를 함께 근거로 해요; 레인은 그 노트를 인용만 해요"
  digest: "sha256:80f3cd406000f62c6a34abd4096312db32326bee5272c79b74ff22e0ae2908c5"
---

# The LEADER route of the moai skill

## Intent

사용자 선택(2026-11-22T23:49:36Z, 완독 보고서 `skill-command-structure-readthrough-2026-11-23.md` 를 본 뒤)에 따라 `/moai leader` 는 상주 역할이 아니라 "등록된 조정 창"이에요. 친 창이 조정 창구로 등록되고(호스트 세션 id 포함, 두 번째 창은 거절), 레인의 완료 보고를 받아 원장을 먼저 읽은 뒤 `/clear` 와 다음 배차를 하고, 창을 닫으면 등록이 끝나서 레인은 원장만 보고 스스로 돌아가요. 같은 선택이 정한 순서는 리더 등록(C-144, 착지) → 완료 핸드셰이크(C-373, 글 착지) → 경로(이 SPEC) → 레인 수 추천·시작이에요.

앞의 두 조각은 명령과 문장을 남겼어요: `moai session-register --mode leader` 는 호스트 세션 id 를 항목에 적고 두 번째 살아 있는 leader 를 거절하며(`SPEC-20261123-leader-registration-lease`), `flows/next-flow` 의 "Completion message" 와 "Transition" 은 레인이 보내는 알림과 리더가 원장을 먼저 읽는 순서를 적었어요(`SPEC-20261123-lane-completion-handshake`). 비어 있는 것은 그 명령과 문장을 한 길로 묶는 경로 글 하나와 그 경로를 고르는 낱말이에요. 이 SPEC 은 그 경로 글만 다뤄요: 새 명령도, 새 코드도 없어요.

바뀌는 것: (1) 경로표가 아홉 개가 돼요(`LEADER` 추가, 부모 SPEC REQ-1 의 여덟을 고쳐요). (2) 요청의 첫 낱말이 `leader` 이면 LEADER 를 골라요(`work` 와 같은 방식). (3) `workflows/leader.md` 한 파일이 등록 → 라운드 → 착지 → 끝(해제)을 적어요. 흐름 전문은 `flows/next-flow` 에만 있고 이 글은 그 절을 가리키기만 해요(부모 REQ-11). (4) 체인표와 Orca 어댑터 표에 LEADER 행을 더해요.

### What exists (tip `5aa64681a2`, 잰 시각 2026-11-23T19:49Z)

- 경로표 `skills/moai/references/routes.json`: READ, PLAN, RUN, AUDIT, SYNC, PROJECT, UPDATE, WORK 여덟. `SKILL.md:11` 은 `work` 만 요청의 첫 낱말로 경로를 골라요.
- 여덟 경로를 세는 핀: `tests/next_flow_skill_scan.rs:41-43`(`ROUTES: [&str; 8]`), `tests/product_cli.rs:2138-2140`, `tests/agent_defs_package.rs:530`(체인 행 8), `tests/orca_optional_3005.rs:196-200`(Orca 표 행 8). 어느 파일이 더 있는지는 본 조각의 첫 단계(목록 작성)가 정해요.
- 금지어(`tests/next_flow_skill_scan.rs:242-264`): `resident leader`, `leader mode`, `mode-leader` 와 `mode-agent` 는 출하 글에 쓰지 않아요. `leader` 낱말 하나와 `registered leader` 는 이미 `flows/next-flow` 와 `workflows/work.md` 가 써요.
- 등록부: `moai session-register --show --name leader` → `status: free`(잰 시각 2026-11-23T19:49Z, 지금 등록된 조정 창 없음). 등록 코드는 `src/leader_round/registration.rs`, 시험은 `tests/leader_registration_cli.rs`. `internal/loop` 2항이 이미 `moai session-register . --name leader --mode leader` 로 등록하고 `conflict` 거절만 "다른 세션이 리더"로 읽으라고 적어요.
- 완료 알림 글: `flows/next-flow.md` "Completion message (a lane to the leader)" 와 "Transition" 앞 세 문장(리더는 원장을 먼저 읽고, 원장이 받치지 않는 알림은 아무것도 시작하지 않는다).
- 보내는 명령: `moai lane-send`(Orca 터미널 전용, `orca terminal send` 를 부름)가 있고, 호스트 간 메시지는 호스트 도구(`SendMessage`)예요. 알림을 보내는 `moai` 명령은 아직 없어요(`flows/next-flow.md` 가 그 간격을 적었어요).
- `adapters/orca.md` 표는 경로마다 행 하나(`With Orca`, `Without Orca`, 호스트 차이)를 요구하고, `tests/orca_optional_3005.rs` 가 경로표와 행 집합이 같은지 검사해요.

## Interview

- Q1: Is `/moai leader` a standing role? -> A: No (user pick 2026-11-22T23:49:36Z): a registered coordination window; `resident leader` and `leader mode` stay forbidden words.
- Q2: How does the skill pick the route? -> A: The first word of the request after the skill name is `leader`, like `work` (D-1); a dash option stays plain text (parent REQ-4), so the retired mode flags and the card that would rename them (C-214) are not used.
- Q3: Which command renews the lease? -> A: `moai session-register . --name leader --instance <instance>` (D-2): `leader-round` needs an observations file of Orca screens, and the renewal must work without Orca.
- Q4: How does the window send the dispatch? -> A: The text names the host's message tool and, only when the user chose Orca for the lanes, `moai lane-send` (D-3); no `moai` command sends a cross-session message yet.
- Q5: Does this SPEC change the profile value `m-flow`? -> A: No (D-5): the profile value `next-flow` and the move of existing `m-flow` profiles are code and a migration; they get their own SPEC.

## Non-goals

- A new `moai` command or a change of an existing one: the route calls `session-register`, `card inbox`, `card wait`, `card next`, `card land` and `leader-round` as they are.
- The completion message's text and its order: `flows/next-flow` owns them; the route points to them and repeats none of it (parent REQ-11).
- The lane count recommender and the start of lanes without Orca (C-319 step 6, C-348, C-315), and the profile value `next-flow` with the migration of `m-flow` profiles (user pick of 2026-11-22, item 3): other SPECs.
- Replacing the installed copy under `~/.claude` or any push: the user's pick only (that note's item 2).
- A timer that repeats the round: `internal/loop` decides who may arm one; the route says "read `internal/loop` before arming a timer" and nothing else about timers.
- The flag rename card C-214: it is the opposite direction (parent REQ-4 keeps dash options as plain text) and is not taken here.

## Requirements

- REQ-1: The route table shall name exactly the routes READ, PLAN, RUN, AUDIT, SYNC, PROJECT, UPDATE, WORK and LEADER, each with a reference file that the package ships. · This changes REQ-1 of the parent from eight routes to nine (D-4).
- REQ-2: When the first word of a request after the skill name is `leader`, the skill shall pick the LEADER route. · failure: none — any other first word picks its route from the table, and a dash option stays plain text.
- REQ-3: When the LEADER route starts, the workflow shall read the registry entry with `moai session-register --show --name leader` first, then register this session with `moai session-register . --name leader --mode leader` and keep the `instance` the registration returns. · failure: none — a read that does not answer `free` registers nothing (REQ-4).
- REQ-4: IF the entry is held by another session or the registration is refused with `conflict`, THEN the workflow shall stop the route with one line that names the holder's lease end, register nothing and report a refusal of any other kind as an error. · failure: none — the stop is the outcome.
- REQ-5: While the route runs, the workflow shall renew the registration with `moai session-register . --name leader --instance <instance>` once per round, before it reads the ledger.
- REQ-6: The workflow shall take its input from the ledger (`moai card inbox`, `moai card wait`, `moai card next`) and treat a lane's completion message as a notice that it checks against the ledger before it acts, pointing to the "Completion message" and "Transition" sections of `flows/next-flow` without repeating their sentences.
- REQ-7: The workflow shall send the audit of a head to a session that wrote none of the change, then land that head with `moai card land <card> --head <40 hex> --by leader` after a verdict `PASS` for that exact head. · failure: none — a head without that verdict is not landed.
- REQ-8: When a lane's note is `LANE-BLOCKED kind=decision`, the workflow shall sort the decision in the order of the skill entry's open-question rules: a reversible choice of category none is decided and recorded by this window, and one that touches a HOLD category or a `guarded` setting goes to the user and waits. · failure: none — the choice that goes to the user waits for the user's pick.
- REQ-9: When the route ends, the workflow shall release the registration with `moai session-register . --name leader --release --instance <instance>` and say that a window that is closed releases nothing, so its lease ends by itself and the lanes then write the ledger note only. · failure: none — a lease that ends is the outcome.
- REQ-10: The LEADER workflow shall describe the window as a registered coordination window that holds no standing role.
- REQ-11: The chain table and the Orca adapter table shall each carry one LEADER row, the Orca row with its three status words.
- REQ-12: The LEADER workflow shall be at most 6144 bytes.

## Acceptance

- AC-1 (REQ-1, REQ-2): the route table holds the nine routes and `workflows/leader.md` exists, the skill entry names `leader` as a first word beside `work`, and a request that starts with a dash option picks no route · verify: `cargo test --locked --offline --test next_flow_skill_scan routes_`
- AC-2 (REQ-3, REQ-4): the workflow reads `session-register --show --name leader` before it registers, registers with `--mode leader` only on `free`, keeps the `instance`, stops on `conflict` naming the lease end, and reports any other refusal as an error · verify: `cargo test --locked --offline --test leader_route register_`
- AC-3 (REQ-5, REQ-6, REQ-7, REQ-8): the workflow renews with `--instance` before it reads the ledger, names `card inbox`, `card wait` and `card next` as its input, points to the two sections of `flows/next-flow` and holds none of their sentences, sends the audit to another session, lands with `--by leader` after a `PASS` for the exact head, and sorts a decision note by the open-question order · verify: `cargo test --locked --offline --test leader_route round_`
- AC-4 (REQ-9, REQ-10): the workflow releases with `--release --instance`, says a closed window releases nothing and the lanes then write the ledger note only, describes no standing role, and the scrub of the shipped texts (parent REQ-8: no resident leader, no leader mode, no mode flag) passes with the new file · verify: `cargo test --locked --offline --test leader_route end_` and `cargo test --locked --offline --test next_flow_skill_scan scrub_`
- AC-5 (REQ-11, REQ-12): the chain table has the LEADER row, the Orca table has the LEADER row, `leader.md` is within 6144 bytes and the source manifest carries the digest of every package file the change adds or edits · verify: `cargo test --locked --offline --test agent_defs_package chains_`, `cargo test --locked --offline --test orca_optional_3005`, `cargo test --locked --offline --test product_cli skill_package::` and `cargo test --locked --offline --test source_manifest_facts`
- AC-6: every older pin that counted eight routes now counts nine, and the neighbouring checks pass · verify: `cargo fmt --check` and `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- this slice, write: `.moai/specs/SPEC-20261124-moai-leader-route/**`
- text slice, write (after the leader's approval): `skills/moai/workflows/leader.md` (new), `skills/moai/SKILL.md`, `skills/moai/references/routes.json`, `skills/moai/references/chains/index.md`, `skills/moai/references/adapters/orca.md` (one table row), `skills/moai/references/source-manifest.json`, `tests/leader_route.rs` (new), `tests/next_flow_skill_scan.rs`, `tests/product_cli.rs`, `tests/agent_defs_package.rs`, `tests/orca_optional_3005.rs`, the row of the parent SPEC's REQ-1 and AC-1 (`.moai/specs/SPEC-20261122-next-flow-skills-agents/spec.md`, one revision), and any other test the inventory of step 1 finds that counts the routes
- read-only: `skills/moai/references/flows/next-flow.md`, `skills/moai/workflows/work.md`, `skills/moai/references/internal/loop.md`, `src/leader_round/registration.rs`, `src/lane_registry.rs`
- forbidden: `skills/moai/references/flows/next-flow.md`, `skills/moai/workflows/work.md`, the folder `skills/moai/references/home`, every Rust source file, the real HOME, the network, the integration branch

## Plan

1. Inventory: run `grep -rn` for each route name and each count of eight over `tests/`, `src/`, `tools/`, `skills/` and `docs/` outside the closed SPECs, and list each pin that must follow (the four named under "What exists" and any other).
2. RED: `tests/leader_route.rs` with the prefixes `register_`, `round_` and `end_`, and the nine-route form of the old pins; each fails on the current text.
3. GREEN: `workflows/leader.md` (outline: Input, Output, Procedure with the five stages register → round → dispatch → land → end, Exceptions and Permissions, Completion Check), the route table entry, the skill entry sentence, the chain row, the Orca row.
4. The parent SPEC's REQ-1 and AC-1 change to nine routes in one revision, with decision D-21 naming this SPEC.
5. Manifest digests with `python3 -I tools/source_manifest.py write` then `check`; one-word guard mutants for each new pin; `moai spec check`; `moai trailer-check`; one `check affected` at the READY head.

## Risks

- A route that talks about a coordinator invites the forbidden words (`resident leader`, `leader mode`); the scrub test runs over the new file, and the text says "registered coordination window" and "no standing role".
- A window that is closed cannot release its registration; the lease (default four hours, at most 86400 seconds) is then the only end, and for that time a lane reads the entry as held and waits for a dispatch that will not come. The route text says so and tells the user the entry can be released with `session-register --release --by leader` and the instance `--show` prints.
- `moai card land` accepts only `--by leader` (`flows/next-flow.md`, Landing): the LEADER window is that caller today. The text says "today" so it does not freeze the rule.
- The completion message has no `moai` command (the gap `flows/next-flow.md` names). The route names the host tool and the Orca command and claims neither as a guarantee; the ledger watch is the way a dispatch is found when no message arrives.
- Four pins count eight routes; a fifth that the inventory misses fails the run, not the text. The first step lists them before the tests are written.
- `internal/loop` already tells how to register for a timer; the route leaves that to it and does not copy the paragraph.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The route is picked by the first word `leader` of the request, like `work`; the token is `LEADER`. | User pick 2026-11-22T23:49:36Z (item 1) and the report's recommendation to take `leader` as a request word and keep `resident leader` and `leader mode` forbidden. | Pick the route another way. |
| D-2 | The lease is renewed with `session-register --instance`, not `leader-round --renew-instance`. | `leader-round` requires an observations file of Orca screens (`moai leader-round --help`, read 2026-11-24), and the route must work without Orca. | Name `leader-round` as the renewal. |
| D-3 | The dispatch names the host's message tool and `moai lane-send` for Orca lanes, and claims no `moai` command for the cross-session message. | `moai lane-send --help` (Orca only) and the gap in `flows/next-flow.md`. | Add the command's name once a `moai` command exists. |
| D-4 | REQ-1 of the parent SPEC changes from eight routes to nine. | The user's pick puts the route in this card; `routes_` pins and the Orca table follow. | Remove the route. |
| D-5 | The profile value `next-flow` and the move of `m-flow` profiles are not in this SPEC. | They are code (`src/profile.rs`) and a migration with their own tests; the route text needs neither. | Fold the migration in. |
| D-6 | The route text points to `flows/next-flow` and repeats none of its sentences. | Parent REQ-11: the full flow text lives in one place. | Copy the sentences. |

## Result

(Filled when the work is done. This is the SPEC-only slice.)
