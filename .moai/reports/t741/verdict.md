# t741 판정서 — 세션 워크트리 전환 시 배경 에이전트 Bash 거부 (앵커 판별)

- card: t741 (리드 배차 · 운영자 선택 · 2026-09-14 집행)
- class: 리드 배차 기준 B(재현 우선). 카드 본문 표기는 "Tier S~M, Class C" — 배차가 지배하되 불일치 사실을 기록한다.
- branch: `WT-anchor-flip-agents` @ `d416f8162` (로컬 develop = origin 배치 8 head와 동일 트리에서 분기)
- 측정 세션: lane-4 (source_session_id=443d0408-02cd-4c2b-a3c1-aab19600e56f)

## 판정 요약

**원인 확정.** 거부 가드는 Claude Code 런타임 소속(`internal/hook/post_tool_failure.go:42-46` 주석 명시: "The guard belongs to the Claude Code runtime, not to this codebase")이며, **세션의 현재 앵커**를 기준으로 판정한다. 배경 서브에이전트는 **자기 앵커가 없다** — 스폰 시점 트리에 고정되지 않고, 호출 시점마다 세션의 현재 앵커를 따라 cwd가 재고정된다(1차 탐침 12/12 실측). 거부는 명령 자체가 앵커와 다른 트리를 **git 축으로 참조**할 때(`-C`, `--git-dir`, cwd 해석 불일치) 발화한다. 비-git 절대경로 인자는 거부되지 않는다.

**수리 방향 판정: (c) — 그리고 (c)는 이미 착지돼 있다.**

- (a) 가드가 옳고 운영 규율로 막을 일 — **부분 채택**. 가드의 목적(한 세션이 여러 트리에 쓰는 것 차단)은 유지 가치가 있다. 다만 "세션 이동은 배경 에이전트가 없을 때만"이라는 규율만으로는 감사 퇴화를 막지 못한다(규율 위반 시 결과가 조용하기 때문).
- (b) 에이전트별 앵커 등록 — **기각**. ① 1차 탐침이 서브에이전트 앵커 부재(부동 cwd)를 실측 — 앵커 등록은 런타임(Claude Code 바이너리) 변경으로 이 저장소 손이 닿지 않는다. ② 의미론적 역효과: 스폰 트리에 고정된 배경 에이전트는 세션이 그 트리를 버린 뒤에도 옛 트리에 쓰기를 계속한다 — 운영자가 폐기했다 믿는 트리의 익명 쓰기 주체를 재도입하는 꼴이라 가드 목적과 충돌한다.
- (c) 거부 사실의 산출물 강제 기록 — **채택, 그러나 신규 기제 불필요**. 카드 발행(2026-09-07) 이후 t529가 두 층을 이미 착지했다: (i) 정책층 — `verification-claim-integrity.md` §3.1 "Refused-tool degradation — a refusal is a Gap, never a silent substitution"(이 트리 140행, provenance card t529): 거부가 재현 명령을 막으면 판정서 Gaps 절에 거부를 명시하라고 규정. (ii) 기계층 — `internal/hook/post_tool_failure.go:33-47`의 `WorktreeGuardRefusal` 범주가 거부를 실패 이벤트로 기록. **lane-3 당시의 에이전트 자발 보고는 우연이었고, 지금은 규약이다.** 잔여 갭은 규약이 정책층이라 게이트가 없다는 것과, 리드가 판정서의 Gaps를 읽어야 한다는 것 — 후자는 kanban-dispatch "Completion is read, never trusted"가 이미 담당한다.

**부수 관측(통합 창 락 브랜치 표기) — 현행 트리에서 이미 수리됨.** `internal/cli/integration.go:105-145` `resolveIntegrationTarget`이 해석 순서(`--branch` 플래그 → 설정된 git-flow develop → 호출자 트리)와 출처 라벨(`BranchSource`, card t449·t637)을 갖추고, 폴백 시 stderr 경고(`integrationFallbackWarning`, card t637)까지 낸다. 카드가 목격한 "세션 앵커 브랜치가 락에 남는" 형태는 현재 설정(develop 구성)에서 재발 경로가 없다.

## Claim

1. 배경 서브에이전트의 Bash cwd는 스폰 시점 트리에 고정되지 않고 세션의 현재 앵커를 실시간 따라간다.
2. 세션 앵커 ≠ 목표 트리인 `git -C` 명령은 가드가 거부하며, 같은 명령이라도 자기 트리 대상이면 통과한다.
3. 비-git 절대경로 인자(`ls` 등)는 크로스 트리여도 거부되지 않는다 — 거부 축은 git 한정.
4. 거부 메시지 문안은 이 저장소 테스트 픽스처가 고정한 4변형 복합 명령 전체가 함께 거부된다(비-git 부분 포함).

## Evidence

**1차 탐침(t741-probe, 경로 없는 명령 12회)** — 스폰 앵커 t741, 실행 중 세션을 primary → t741-rb로 이동:

- call 1-3 (17:05:59~17:06:31): pwd = primary 체크아웃 — 통과 3/3
- call 4-12 (17:08:33~17:10:27): pwd = `.claude/worktrees/t741-rb` — 통과 9/9
- 거부 0건. **cd 없이 cwd가 호출 사이에 두 번 바뀜** — 스폰 시점 고정이 없음을 직접 실측. call 3→4 사이 약 2분 갭(세션 이동 창)에서 에이전트 턴이 정지했다가 재개 — 이동 창 동안 배경 에이전트가 멈추는 관측도 함께 기록.

**2차 탐침(t741-probe2, t741 절대경로 참조 복합 명령 8회)** — 스포느로부터 첫 호출까지의 창이 짧아 call 1이 이동 후에 실행됨(대조쌍 실패 — Gaps 참조):

- call 2-8 (7/7) 전부 동일 거부, 원문:

> This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t741-rb, but this command redirects git to the shared checkout via -C. Refusing to run it — a worktree-isolated session's git operations must target its own worktree. Run the equivalent from /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t741-rb without the redirect.

- 복합 명령 안의 `git -C`가 명령 전체(sleep·date·pwd·ls 포함)를 죽였다.

**대조 3건(본 세션, 앵커=t741-rb, 직접 재어)**:

| 명령 | 결과 |
|---|---|
| `ls -la /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t741/go.mod` (크로스 트리, 비-git) | **통과** |
| `git -C …/worktrees/t741-rb log --oneline -1` (자기 트리) | **통과** (`d416f8162`) |
| `git -C …/worktrees/t741 log --oneline -1` (남 트리) | **거부** — 위 원문과 동일 템플릿 |

**코드·문서 근거(전부 이 트리 `d416f8162` 판독)**:

- `internal/hook/post_tool_failure.go:33-47` — `WorktreeGuardRefusal` 범주(card t529), 앵커 토큰 `"isolated in the worktree"`, 가드 소속이 "Claude Code runtime"으로 명시.
- `internal/hook/worktree_guard_refusal_test.go:25-40` — 거부 문안 4변형(-C / --git-dir / too-complex / cwd-resolved)이 실측 인용 픽스처로 고정. 카드 사례 원문(t526/t508)이 `sampleGuardRefusalCwdCardQuoted`로 그대로 있음. 2차 탐침의 거부는 `sampleGuardRefusalDashC`와 경로만 다른 동일 문안.
- `.claude/rules/moai/workflow/worktree-integration.md:472-501` — "Which guard refused" 표(이 저장소 코드 아님) + 트리거 관측표(비완결 명시).
- `.claude/rules/moai/core/verification-claim-integrity.md:140` — §3.1 Refused-tool degradation (provenance card t529).
- `internal/cli/integration.go:105-145, 244-245` — 통합 창 락 브랜치 해석 순서 + 출처 라벨 + 렌더링.

## Baseline-attribution

모든 실측은 이 실행에서, `WT-anchor-flip-agents` = `d416f8162` 트리 기준으로 수행됐다(2026-09-14 02:04-02:15 KST 전후). 코드·문서 인용은 같은 트리 판독. 실행 명령은 Evidence 표와 본문에 그대로 기재.

## 수리 방향 상세 (리드 사전 승인 불필요 범위)

1. **문서 갱신 (후속 카드 권고)** — `worktree-integration.md` §"Refused Commands"에 (i) `-C`/`--git-dir`/cwd-resolved 변형 문안을 "too complex" 변형과 나란히 기록, (ii) 서브에이전트 앵커 부재(부동 cwd) 행동을 기록, (iii) 대조 판별식(git 축 한정, 비-git 경로 인자 통과)을 기록. 가드를 약화하지 않는 관측 기록이며, lane-3형 감사 퇴화의 인지 비용을 낮춘다.
2. **운영 규율 보강 (문서에 한 줄 권고)** — 감사(auditor) 계열 에이전트를 배경으로 띄운 세션은 그 에이전트 종료 전에 워크트리를 옮기지 않는다. 불가피하다면 판정서 Gaps에 거부 기록이 반드시 들어가는지 리드가 확인한다(VCI §3.1).
3. **비권고** — (b) 에이전트별 앵커 등록(런타임 변경 + 의미론 역효과), 가드 완화 일체.

## Gaps

- **cwd-resolved 변형(t526/t508형)은 직접 재현하지 못했다.** 1차 탐침은 거부 0건, 2차 탐침은 `-C` 변형만 발화. cwd 해석 불일치가 정확히 어떤 명령 형태에서 오는지(예: 명령 내 `cd` 포함 여부)는 미측정 — 원인 **등급** 확정(세션 앵커 판별)에는 충족하지만 변형별 트리거 경계는 문서 표(:493)와 마찬가지로 미완이다.
- lane-3 원사례(t508→t526)의 당시 런타임 버전과 현 시점 런타임 버전의 차이는 측정하지 않았다. 부동 cwd 행동이 당시에도 동일했는지는 단언할 수 없다(다만 거부 문안 픽스처가 동일 계열임은 확인).
- 2차 탐침의 이동-전 대조 호출(앵커=t741에서 같은 `git -C t741` 통과)은 확보하지 못했다 — 본 세션이 앵커=t741일 때 동일 명령이 통과했음을 별도 관측으로 남기지 않았다. 대신 자기 트리 `git -C`(앵커=t741-rb 기준 t741-rb 대상) 통과가 통제 역할을 대신했다.
- "the shared checkout" 오표기(워크트리 경로를 shared checkout이라 부르는 메시지 결함)는 관측됐으나 가드가 런타임 소속이라 이 저장소에서 수리 불가 — 문서 기록으로만 대응 가능.

## Residual-risk

- 거부 문안은 런타임 의존 문자열이라 Claude Code 업데이트로 변할 수 있다 — `worktreeGuardAnchor` 주석이 이 의존성을 이미 경고하며, 바뀌면 `WorktreeGuardRefusal` 탐지가 조용히 0건이 된다(verification-completeness §1.3 absent-execution 형태).
- VCI §3.1은 정책층이라 게이트가 없다 — 판정서가 Gaps 규약을 어겨도 기계적 실패 신호는 없고, `WorktreeGuardRefusal` 인박스 기록과 판정서 본문의 부일치는 리드의 읽기에만 의존한다.
- 부동 cwd는 거부가 아니라 **침묵하는 재앵커링**이다 — 경로 없는 명령을 계속 돌리는 배경 에이전트는 자기가 다른 트리에서 일하고 있음을 스스로 알 수 없다(1차 탐침이 보여준 형태). 감사 퇴화의 이 경로는 거부 기록으로 잡히지 않는다.

🗿 MoAI
