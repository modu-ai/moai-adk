# 자율 무중단 개발 모드 SPEC — `/moai --mode-leader` · `/moai --mode-agent` (#2138, DR-02)

> 현행 상태(#2167): 구현된 형태는 `/moai --mode-leader [--autonomy full|semi]`·`/moai --mode-agent`(#2160, `skills/moai/SKILL.md`)다. 이 문서의 표기는 #2286(사용자 결정 Issue #2389 댓글 6774899518 행 36)에서 이 두 형태로 고쳤다. 설계 당시 이름 `--leader-mode`·`--agent-mode`와 그 뒤의 `--mode leader`·`--mode agent`는 더 쓰지 않는다. §3의 한 회차는 `moai-cli queue-next`(#2161), §5의 한 tick은 `moai-cli leader-tick`(#2162)이 실행한다. 이 저장소에만 해당하는 운영 규칙은 `AGENTS.local.md` § Agent loop (self-claim)·§ Leader supervision·§ Gate에 있고, 그 이전 제안은 `docs/lane-loop-contract.md`(2026-11-08)다. 본문의 `AGENTS.local.md:행` 인용은 `604792666` tree 기준이며 #2167 재구성 뒤 행 번호가 바뀌었다.

- 상태: SPEC 초안(Class C, 설계 변경). 이 문서는 구현하지 않는다. 코드·skill 파일은 바꾸지 않았고, 구현은 §10의 후속 카드가 맡는다.
- 입력: 종합 개선 보고서 §10.4(`references/chatgpt-space/synthesis-improvement-report-2026-11-09.md`, 2026-11-09 작성, 저장소에 아직 tracked 아님). 사용자 결정 2026-11-09 00:0x~00:3x(모드 skill 진입, 완전/반자율, self-claim 기본, 리더 5분 PR 점검·타 에이전트 수리 지시, `/goal` 붙여넣기 안내).
- 대조 기준: `origin/v2026.11.0` = `604792666`. 이 문서의 `경로:행`은 모두 이 tree에서 읽은 값이다. `<primary>/.claude/…`는 Go 세대 규칙 파일이며 이 저장소에 tracked가 아니다(이 tree의 `git ls-files .claude` 0건). 행 번호는 2026-11-09 primary checkout에서 읽었다.
- 표기: **있음**은 그 단계의 동작이 RC1 코드나 tracked 규칙에 실제로 있다는 뜻이다. **부분**은 쓸 수 있는 부품은 있으나 이 단계로 연결되지 않았다는 뜻이고, **신규**는 RC1에 대응물이 없다는 뜻이다. 근거 없는 추정은 적지 않았다. 확인하지 못한 것은 §11에 따로 모았다.

## 1. 진입 방식

론처는 역할을 모른다. `moai-cli cc|glm|codex`를 `-l`/`-a` 없이 실행하면 역할을 추측하지 않는 평범한 host 세션이 뜬다(`src/main.rs:81-82`, #2053). 역할은 세션 안에서 공통 skill 인자로 켠다.

| 형태 | host | RC1 대조 |
|---|---|---|
| `/moai --mode-leader [--autonomy full\|semi]` | Claude Code | **신규**. `skills/moai/SKILL.md`에는 모드 인자가 없고, `skills/moai/references/routes.json`의 route는 8개(`AUDIT PLAN READ RUN SETUP SYNC UPDATE VERIFY`)뿐이다 |
| `/moai --mode-agent` | Claude Code | **신규**. 위와 같다 |
| 같은 인자를 Codex skill 호출로 전달 | Codex CLI | **미확인**. Codex에서 공통 skill에 인자를 넘기는 형태는 이번에 확인하지 않았다(§11) |

지금은 `-l`/`-a` 론처 플래그가 세션 이름과 첫 prompt만 바꾼다(`src/main.rs:85-95`, `src/session.rs:13-15`). 역할 행동은 `AGENTS.local.md` § Lane loop(`AGENTS.local.md:28-39`)와 SessionStart 주입에 기대고 있다. 모드 skill은 이 암묵적 주입을 명시적 진입으로 바꾼다. `-l`/`-a`를 남길지는 §8의 결정 항목이다.

## 2. 모드별 책임

| 책임 | 에이전트 모드 | 리더 모드 |
|---|---|---|
| 작업 획득 | self-claim(§3). 리더 배차를 기다리지 않는다 | 배차하지 않는다. 미배차 목록을 보충한다(보고서 → Issue 변환, 수리 Issue 발행) |
| 진행 | 카드 하나를 worktree에서 끝내고 PR·영속 보고 뒤 다음 카드로 간다. 병합을 기다리지 않는다(`AGENTS.local.md:32`, `:36`) | 주기적 PR 점검(§5), 수리 지시, 배정 기록 대조, 목록이 빌 때 대기 사유 통지 |
| 완료 판정 | 하지 않는다 | 증거(PR 병합, verdict 파일)를 읽어서만 한다. 메시지나 idle 알림으로는 판정하지 않는다 |
| 통합 | 하지 않는다(작성자 ≠ 감사자 ≠ 통합자) | 게이트 절차를 따른다(`AGENTS.local.md:39`: verdict 첫 줄 → head == 감사 SHA → `moai/audit` → auto-merge) |
| 특정 레인 지정 | — | Issue 본문이나 댓글에 `assignee-hint: agent-N`을 적는다. 그래도 소유권은 그 에이전트의 claim 댓글이 성립해야 생긴다 |

## 3. self-claim 7단계 — RC1 대조

| 단계 | 내용 | RC1 | 근거(경로:행) | 실패 분기 |
|---|---|---|---|---|
| ① 목록 갱신 | 미배차 Issue 번호 목록을 원격에서 다시 읽어 갱신한다(읽기 전용 스냅샷) | **부분** | 원격 읽기는 있다. `forge-read`가 열린 Issue와 댓글을 조건부 GET으로 읽는다(`src/forge/mod.rs:1-7`, `:347`, CLI `src/main.rs:801`). 반면 "미배차 목록"을 만드는 코드는 없다. `src/` 전체에서 `dispatch_journal`·`assignments`·`github-issues.json`을 찾으면 0건이다. `.moai-cli/manifests/github-issues.json`은 리더가 손으로 관리하는 기록이고, 코드가 읽지 않는다 | 원격 읽기 실패 → 목록을 갱신하지 않고 이전 스냅샷을 쓰지 않는다. 그 회차는 claim하지 않고 대기를 기록한다 |
| ② 후보 선택 | 우선순위와 ready 조건(선행 Issue 종료)으로 1건을 고른다 | **신규** | Issue 큐용 선택 코드는 없다. `plan-admit`(`src/plan.rs:1-6`)은 SPEC task DAG의 `ready/stale/blocked`를 계산하는 것이라 Issue 큐 선택과는 다르다. 재사용할 수 있는지는 후속 카드에서 판단한다 | 후보 0건 → `idle: backlog empty <time>`을 한 번 기록하고 멈춘다. 폴링을 늘리지 않는다(Go 규칙 `<primary>/.claude/rules/moai/workflow/factory-dispatch.md:123`) |
| ③ 같은 호스트 직렬화 | `<common dir>/.moai-cli/claims/` journal을 열고 advisory lock을 잡는다 | **있음** | `src/claim.rs:5-13`(journal 위치·authority·advisory lock), `:595-596`(`try_lock`), `:1120-1133`(lock을 못 잡으면 즉시 `busy`, 기다리지 않음) | `busy` → 기다리지 않고 다른 후보로 가거나 다음 회차로 미룬다(REQ-06A-05). 이미 활성 기록이 있으면 `already_owned`, lease가 만료됐으면 `expired`(자동 takeover 금지, `:787-811`) |
| ④ 댓글 claim | `moai-claim: session= epoch= lease_until=` 댓글을 정확히 1건 게시한다 | **있음** | 게시 전에 먼저 읽고, 살아 있는 남의 claim이 있으면 쓰지 않는다(`src/claim.rs:821-841`). POST는 `:846-854`, `post_claim`은 `:506-538`(claim의 유일한 GitHub 쓰기) | POST 실패 → journal에 아무것도 쓰지 않는다. 오류는 envelope `kind`로 보고한다 |
| ⑤ 승자 판정 | 댓글 목록을 다시 읽어 가장 작은 live 댓글 id가 이긴다 | **있음** | `src/claim.rs:856-918`(다시 읽기로 판정), `:447-450`(`resolve_winner` = 최소 id), 모듈 설명 `:15-23` | 다시 읽기 실패, 또는 내 댓글이 목록에 안 보임 → `uncertain`. journal은 쓰지 않는다(`:864-895`). 진 경우 → `already_owned`, journal은 쓰지 않고 내 댓글은 남긴다(규칙상 효력 없음, `:856-859`, `:896-912`) |
| ⑥ 표시(projection) | label(`moai:in-progress` 등)과 소유 댓글의 `상태:` 줄을 갱신한다. 소유권의 근거가 아니다 | **부분** | claim은 `projection_intents`를 기록에 의도로만 남긴다(`src/claim.rs:141`, `:920-924`). label 추가·삭제 부품은 있으나 CLI에 노출되지 않은 라이브러리다(`src/forge/write.rs:1-6`, `:103-135`). 소유 보고 댓글은 `forge-report`(`src/forge/report.rs:1-21`, CLI `src/main.rs:776`). claim 뒤에 이것들을 부르는 연결은 없다 | 표시 실패는 소유권에 영향이 없다. 실패를 기록하고 다음 회차에 다시 맞춘다 |
| ⑦ 실행·반복 | `EnterWorktree` → 구현·검사 → PR → 영속 보고 → `ExitWorktree(keep)` → 병합을 기다리지 않고 ①로 | **부분** | worktree 생성은 `worktree-apply`(`src/worktree.rs:1-9`, CLI `src/main.rs:966`). 출입 규칙은 `AGENTS.local.md:33`, 반복·대기 금지는 `AGENTS.local.md:32`, `:36`. 이 순서를 스스로 도는 루프 구동부는 없다 | 검사 실패 → PR을 열지 않고 보고한다. 같은 단계가 3번 실패하면 멈춘다(`AGENTS.local.md:37`) |

보고서와 다른 점: 보고서 §10.4는 ⑤에서 "진 쪽은 로컬 기록을 `lost`로 닫는다"고 적었다. RC1에서 진 쪽은 journal에 **아무것도 쓰지 않는다**(`src/claim.rs:906-911` "no journal record was written"). 상태 목록 `ACTIVE_STATES`(`src/claim.rs:50-59`)에도 `lost`가 없다. 이 SPEC은 RC1 동작을 따른다.

### 현재 선택 경계: 지정 handoff (#2390 S4a)

`queue-next`는 유효한 `moai-claim-handoff: comment=<id> to=<session>`의 수신자를 세션 id의 **정확한 문자열 일치**로 구분한다. 지정되지 않은 세션은 그 카드를 `skipped`의 `reason: handed_to_other`로 남기고 다음 후보로 간다. 로컬 journal에 활성 기록이 있어도 새 원격 handoff를 우회하지 못한다. `pr_ready` 카드의 기존 건너뛰기는 유지한다.

handoff는 소유권 부여가 아니다. 지정된 세션도 claim 게시와 재조회·승자 판정을 거쳐야 한다. 일반 release는 누구나 다시 고를 수 있게 하며, 수신자가 새 claim 뒤 release한 과거 handoff가 카드를 영구히 예약하지 않는다. lease 만료만으로는 이전하지 않는다. 원격 댓글의 eventual consistency와 `cross_host_atomic: false` 한계는 그대로다.

이 경계는 기존 댓글을 읽는 후보 선택만 다룬다. 메시지 저장 위치·ACK 경보 시간·host wake 연동·구조화된 resume/checkpoint는 구현하거나 결정하지 않는다.

## 4. 동시성 논증

중재 수단은 **Issue 댓글 id**다. 로컬 파일이나 label은 중재 수단이 될 수 없다.

1. **로컬 파일은 lock이 될 수 없다.** journal과 lock은 clone 하나의 git common dir 아래에 있어서(`src/claim.rs:5-8`) 다른 PC나 다른 clone의 세션은 볼 수 없다. 같은 호스트에서도 advisory lock이라 `moai-cli`를 거치지 않는 프로세스는 직렬화되지 않는다(`src/claim.rs:11-13`, 실행 시 notes `:947-949`). 그래서 journal은 같은 호스트 직렬화와 캐시에만 쓴다.
2. **label은 lock이 될 수 없다.** GitHub는 unsafe 메서드에 조건부 요청(If-Match)을 지원하지 않는다. RC1의 쓰기 보호는 "쓰기 전에 다시 읽기"이지 원자적 CAS가 아니다(`src/forge/write.rs:4-6`). 두 세션이 같은 label을 동시에 붙여도 둘 다 성공한다. 저장소 규칙도 label·assignee는 표시이며 lock이 아니라고 적고 있다(`docs/lane-loop-contract.md:8`).
3. **댓글 id 단조성은 가정이다.** RC1은 "살아 있는 claim 가운데 가장 작은 댓글 id가 이긴다"로 판정한다(`src/claim.rs:15-19`, `:447-450`). 이 판정은 먼저 게시된 댓글의 id가 더 작다는 가정에 기댄다. 이 가정이 GitHub 공식 문서에 보장으로 적혀 있는지는 이번에 확인하지 않았다(§11). 판정은 게시 뒤에 다시 읽어서 하므로, POST 결과 하나로 결정하지 않는다(`src/claim.rs:856-859`).
4. **가정이 깨질 때의 fallback은 `uncertain`이다.** 다시 읽기가 실패하거나, 방금 게시한 내 댓글이 목록에 보이지 않으면(read-your-writes 불확실) 승자를 판정하지 않는다. 결과는 `uncertain`이고 journal은 쓰지 않는다(`src/claim.rs:864-895`). 에이전트는 그 카드를 소유한 것으로 행동하지 않는다. 사람이나 리더가 `claim-show`와 댓글을 보고 정리한다.
5. **lease는 각 호스트의 시계로 판정한다.** `is_live`는 호출한 쪽의 현재 시각(`clock.now()`, `src/claim.rs:786`)과 비교한다. 호스트 간 시계가 어긋나면 live 판정이 달라질 수 있다. lease 수치 자체가 HOLD다(`src/claim.rs:25-28`).
6. **세션 id가 겹치면 판정이 무너진다.** 승자의 `session`이 내 `owner.id`와 같으면, 그 댓글을 내 claim으로 "채택"한다(`src/claim.rs:829-845`). 같은 이름의 세션이 둘이면 서로의 claim을 자기 것으로 받아들일 수 있다. 2026-11-09에 `agent-1` 이름의 세션 두 개가 동시에 살아 있는 것이 관측됐다. #2138의 claim은 리더가 `session=agent-1-e29c15`로 구분해 게시했다. self-claim 모드는 세션 id가 유일하다는 것을 전제 조건으로 삼아야 한다(§9 REQ-AM-04).

## 5. 리더 감독 루프

- 주기: **5분**. 사용자가 지정한 값이며 측정해서 정한 값이 아니다. 실제로 맞는 주기인지는 운영하면서 기록해 다시 정한다.
- 한 회차에서 읽는 증거: 열린 PR(head·base·병합 상태·CI·감사 verdict·충돌), 각 claim 댓글과 배정 기록, `ListAgents`(살아 있는 세션·중복 이름). 리더 상시 점검 규칙과 같다(`AGENTS.local.md:35`).
- 행동: PASS면 게이트를 진행한다(`AGENTS.local.md:39`). 수리가 필요하면 수리 Issue나 댓글을 만들고 **원 작성자가 아닌** 레인을 `assignee-hint`로 지목한다. 목록이 비었는데 대기 중인 에이전트가 있으면 사유를 알린다.
- 깨우기 수단: Claude runner는 스스로 다음 깨우기를 예약한다(`ScheduleWakeup`, 상대 지연). 반복 cron과 절대 시각 one-shot은 쓰지 않는다(Go 규칙 `<primary>/.claude/rules/moai/workflow/factory-dispatch.md:123`). Codex runner에는 스스로 깨우는 도구가 없다. 대체 수단은 증거 읽기와 디스크 대기 기록이다(Go 규칙 `<primary>/.claude/rules/moai/workflow/auto-semantics.md:200-204`). 따라서 Codex 세션은 리더 모드를 맡을 수 없거나, 사람이 깨우는 제한을 갖는다. 이것은 §8의 결정 항목이다.
- 변화가 없는 PR 때문에 폴링을 늘리지 않는다.

## 6. 결정 등급과 keep-set

| 등급 | 리더가 스스로 결정 | 사람에게 넘김 |
|---|---|---|
| full(완전자율) | keep-set 밖의 모든 결정. 결정 사다리를 따른다: 디스크 증거 → decision board → 감사 교차 → `jev_ask`(켜져 있을 때만) → 자체 판단과 기록(`<primary>/.claude/rules/moai/workflow/auto-semantics.md:148-163`) | keep-set 3종만: 환경상 불가, 운영자 보유, 되돌릴 수 없는 외부 공유 작업(`auto-semantics.md:208-213`) |
| semi(반자율) | 위에서 아래 목록을 뺀 나머지 | keep-set 3종 + Issue 신규 발행, HOLD 해제, 병합 정책 변경, release |

- Jev: RC1에서 Jev는 P10 HOLD다. RC1의 `decision.record`는 `authority`가 `none` 하나뿐이라 아무것도 승인하지 못한다(`src/decision.rs:1-6`). Go 규칙에서도 Jev 답은 참고 입력일 뿐 결정이 아니다(`auto-semantics.md:358-368`). 모드는 "Jev를 쓸 수 있는가"를 먼저 확인하고, 쓸 수 없으면 자체 판단으로 간다. 어느 쪽이든 결정 기록을 남긴다.
- 안전 경계는 두 등급 모두 같다. ruleset과 `moai/audit` 독립 감사, 비밀 미노출, hook·권한 우회 금지, Go 상태 보존, release/deploy 금지를 지킨다(`docs/decisions-2026-11-08.md:90`의 "HOLD만 해제" 범위).
- RC1 대조: 결정 사다리와 keep-set은 Go 세대 규칙에만 있다. RC1 Rust와 tracked 문서에서 이것을 모드에 연결한 곳은 없다(**신규**).

## 7. `/goal` 안내 — 자동 입력 가능 여부

| 항목 | 존재 | 모델이 대신 입력할 수 있는가 | 근거(확인일 2026-11-09) |
|---|---|---|---|
| ① Claude Code 내장 `/goal` | **확인됨**. "`/goal` is a wrapper around a session-scoped prompt-based Stop hook". 조건이 충족되거나, 모델이 불가능하다고 판단하거나, 고쳐야 할 오류로 턴이 실패하면 해제된다 | **미확인**. 공식 문서에 언급이 없다. 공식 문서는 내장 명령 대부분이 "built-in commands whose behavior is coded into the CLI"이고 skill과 다르다고만 적는다 | https://code.claude.com/docs/en/goal , https://code.claude.com/docs/en/commands (HTTP 200) |
| ② MoAI `goal_arm` | Go 세대에는 **있다**. "Arming is programmatic"(`<primary>/.claude/rules/moai/workflow/goal-directive.md:26`). RC1 Rust에는 **없다**: `routes.json`에 goal route가 없고 `skills/moai/`에 goal 절차도 없다 | Go 세대 한정으로 가능. RC1에서는 쓸 수 없다 | `skills/moai/references/routes.json`(route 8개) |
| ③ Codex CLI `/goal` | **확인됨**. "Set, edit, pause, resume, view, or clear a task goal", "Codex keeps the goal attached to the active chat while work continues". 조건이 충족될 때까지 턴 종료를 막는지는 문서에 없다 | **미확인** | https://learn.chatgpt.com/docs/cli/slash-commands (HTTP 200) |

설계 결정: 모델이 내장 명령을 입력할 수 있다는 근거가 없으므로, 두 모드 모두 **붙여넣기 지시문**을 안내한다. 리더 모드는 켜질 때 열린 Issue와 PR로 완료 조건 문장을 만든다(예: "milestone X의 Issue·PR이 모두 병합되거나 종료됨"). 그리고 사용자가 그대로 붙여 넣을 `/goal <조건>` 한 줄을 출력한다. 에이전트 모드는 자기 카드 범위의 조건을 같은 방식으로 안내한다. Codex 세션에도 같은 한 줄을 안내하되, Codex `/goal`이 계속 진행을 강제하는지는 확인되지 않았다. 그래서 그 효과를 보장한다고 적지 않는다. 보고서 §10.4는 Codex 항목을 "미확인"이라고 적었지만, 명령이 있다는 것은 이번에 공식 문서로 확인했다.

RC1 implementation (A4 #2163): `tools/gen_goal_block.py` is a deterministic, standard-library-only fixture generator. `tests/goal_block_test.py` pins the complete JSON output for issues only, two open PRs with no open issues, and a keep-set overlapping actual open items. Exclude keep-set numbers from both Issue and PR lists before computing counts and agent lines; include their reasons in the condition (REQ-AM-08). If no eligible items remain, return `goal_line: null` for leader and Codex, with an explicit no-open-items condition, rather than offering an already-satisfied `/goal`. Reject malformed numbers, lists and reasons without a traceback. Six tests cover these snapshots, input permutations, empty scopes and invalid input. The paste-only contract matches the user-input goal rows in `docs/host-support-matrix.md`: the model cannot enter `/goal`, and RC1 does not use a Codex self-arming path.

## 8. 결정 항목(사용자 몫)

- **D-1 론처 `-l`/`-a`**: (a) 호환용으로 남긴다. 모드 skill과 같은 세션 이름 기본값을 유지한다. (b) 폐기한다. 역할은 모드 skill로만 켠다. 어느 쪽이든 `moai-cli cc|glm|codex`만 실행하면 역할을 추측하지 않는다는 #2053의 동작은 유지한다.
- **D-2 5분 주기**: 사용자 지정값으로 시작한다. 측정 기록이 쌓이면 다시 정할지 정한다.
- **D-3 Codex 리더 모드**: 스스로 깨우는 도구가 없으므로 (a) Codex는 에이전트 모드만 허용한다. (b) 사람이 깨우는 제한을 명시하고 허용한다.
- **D-4 미배차 목록의 위치**: `.moai-cli/manifests/github-issues.json`의 `dispatch_journal`을 확장한다. 또는 옆에 `assignments.json`을 새로 둔다(보고서 §10.4). 어느 쪽이든 원격 Issue가 원본이고 로컬 파일은 스냅샷이다.

## 9. GEARS 요구와 수용 기준 초안

- REQ-AM-01 (WHEN) 사용자가 세션 안에서 `/moai --mode-agent`를 입력하면, 시스템은 §3의 ①~⑦을 리더 배차 없이 반복해야 한다. 각 회차의 결과(`claimed`, `already_owned`, `busy`, `expired`, `uncertain`, 목록 없음)를 디스크에 기록해야 한다.
- REQ-AM-02 (WHEN) 사용자가 `/moai --mode-leader [--autonomy full|semi]`를 입력하면, 시스템은 Issue를 직접 배차하지 않아야 한다. 대신 §5 감독 루프와 §6 결정 등급으로 동작해야 하고, `--autonomy`가 없으면 `semi`로 시작해야 한다.
- REQ-AM-03 (Ubiquitous) 소유권은 claim 댓글의 최소 live id 판정(RC1 `claim-acquire`)으로만 성립해야 한다. label, assignee, 로컬 파일, 메시지는 소유권의 근거가 되지 않아야 한다.
- REQ-AM-04 (IF … THEN) 같은 `session` 값을 쓰는 살아 있는 다른 세션이 감지되면, 시스템은 claim하지 않고 중복을 보고해야 한다.
- REQ-AM-05 (IF … THEN) claim 결과가 `uncertain`이면, 시스템은 그 카드에 쓰지 않아야 한다. 판정에 필요한 근거(댓글 id, 다시 읽기 실패 사유)를 보고해야 한다.
- REQ-AM-06 (WHILE) 미배차 후보가 없는 동안, 시스템은 `idle: backlog empty <time>`을 한 번만 기록해야 하고 폴링을 늘리지 않아야 한다.
- REQ-AM-07 (Ubiquitous) 두 모드 모두 `/goal`을 직접 실행하지 않아야 한다. 사용자가 붙여 넣을 `/goal <조건>` 한 줄을 출력해야 한다.
- REQ-AM-08 (Ubiquitous) full 등급에서도 keep-set 3종의 결정은 사람에게 넘겨야 한다. 모든 결정은 결정 기록을 남겨야 한다.
- REQ-AM-09 (Ubiquitous) 리더 감독 주기는 설정값으로 두어야 하며, 기본값은 사용자 지정 5분이어야 한다. 문서와 출력은 이 값을 측정값이라고 부르지 않아야 한다.

수용 기준 초안(후속 구현 카드가 그대로 쓴다):

- AC-AM-1: 두 세션이 같은 Issue에 동시에 self-claim하면 정확히 한 세션만 `claimed`이고, 다른 세션은 `already_owned` 또는 `uncertain`이다. 진 쪽 journal에는 기록이 없다. 검증은 fixture transport로 POST 순서를 바꿔 가며 한다.
- AC-AM-2: 같은 `session` 값의 다른 세션이 살아 있으면 claim을 시도하지 않고 중복을 보고한다.
- AC-AM-3: 후보 0건에서 `idle` 기록이 정확히 1번 남고, 이후 회차가 새 GitHub 요청을 늘리지 않는다.
- AC-AM-4: 리더 모드 출력에 `/goal <조건>` 한 줄이 있고, 세션이 그 명령을 스스로 실행한 흔적이 없다.
- AC-AM-5: semi 등급에서 Issue 신규 발행이 사람 확인 없이 일어나지 않는다. full 등급에서 keep-set 결정이 사람 확인 없이 일어나지 않는다.

## 10. 후속 구현 카드 분해

- **(a) skill 인자와 모드 진입 절차.** `skills/moai/SKILL.md`와 `references/routes.json`에 `--mode-leader`, `--mode-agent`, `--autonomy full|semi`를 추가한다. `skills/moai/workflows/`에 모드별 진입 절차를 둔다. 지금 route 8개에는 모드가 없으므로 새 route를 둘지, 기존 route의 인자로 둘지 먼저 정한다. Codex에서 skill 인자를 넘기는 형태를 공식 문서로 확인하는 일이 이 카드의 선행 조사다. REQ-AM-01·02·07을 맡는다.
- **(b) 리더 감독 절차 문서.** §5의 5분 PR 점검, 수리 재배차(원 작성자 제외), 목록 보충, 대기 사유 통지를 factory-dispatch 문서에 넣는다. 리더 상시 점검(`AGENTS.local.md:35`)과 게이트 절차(`:39`)를 바꾸지 않고 참조한다. 깨우기는 Claude는 `ScheduleWakeup`이고, Codex는 D-3 결정에 따른다. REQ-AM-09를 맡는다.
- **(c) assignments 연동과 self-claim 루프.** §3의 ①(미배차 목록 산출), ②(후보 선택), ⑥(claim 뒤 label·소유 댓글 표시 연결)이 이 카드의 코드 작업이다. ③~⑤는 RC1 `claim-acquire`를 그대로 부른다. 목록 위치는 D-4 결정을 따르고, 원격 Issue가 원본이다. 세션 id 중복 감지(REQ-AM-04)와 `uncertain` 처리(REQ-AM-05)를 함께 넣는다. AC-AM-1~3을 맡는다.
- **(d) 결정 등급과 결정 사다리 연결.** full/semi를 결정 사다리와 keep-set에 연결한다. semi의 추가 보류 목록(Issue 발행, HOLD 해제, 병합 정책, release)을 gate 목록에 적는다. Jev 사용 여부는 "켜져 있고 질문형이 맞을 때만"으로 두고, 결정 기록 형식은 RC1 `decision.record`(`src/decision.rs`)를 쓴다. 사다리가 지금은 Go 세대 규칙에만 있으므로, RC1 tracked 문서로 옮길지가 이 카드의 결정 항목이다. REQ-AM-08과 AC-AM-5를 맡는다.
- **(e) 론처 결정.** D-1에 따라 `-l`/`-a`를 남기거나 폐기한다. 남기면 모드 skill과 이름 기본값을 맞춘다. 폐기하면 `src/main.rs` `HostLaunch`의 역할 플래그와 관련 fixture·문서를 정리한다. 어느 쪽이든 역할 없는 실행이 역할을 추측하지 않는 #2053 동작은 회귀 시험으로 지킨다.

## 11. 미확인·한계

- 모델이 Claude Code 내장 `/goal`을 직접 실행할 수 있는지. 공식 문서에 언급이 없다. 보고서는 "이전 세션에서 사용자가 설정한 goal을 리더가 해제하지 못했다"는 관찰을 적었지만, 이 문서에서 재현하지 않았다.
- Codex `/goal`이 조건 충족까지 턴 종료를 막는지. 모델이 입력할 수 있는지도 확인하지 않았다.
- Codex에서 공통 skill에 인자를 넘기는 형태(보고서의 `$moai` 호출).
- GitHub Issue 댓글 id가 게시 순서대로 커진다는 것이 공식 문서에 보장으로 적혀 있는지(§4-3).
- GitHub Actions Ubuntu 러너 이미지 문서. #2029 때 504로 읽지 못했고, 이 카드와는 직접 관련이 없어 다시 시도하지 않았다.
- `<primary>/.claude/…` 인용은 Go 세대 파일이라 저장소 이력으로 행 번호를 다시 확인할 수 없다. 이 문서의 행 번호는 2026-11-09 primary checkout 기준이다.
