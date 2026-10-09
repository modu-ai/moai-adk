auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: 2e37b963a

# SPEC Review Report: SPEC-ALWAYS-LOADED-BUDGET-001
Iteration: 4/3+1 (델타 + 회귀 확인, CN-4 전체 재독) — 리더가 허용한 마지막 연장 회차(decision-index Q9)
Verdict: FAIL (최종 에스컬레이션)
Overall Score: 0.81 (Tier L 문턱 0.85 미달; iter3 0.80 → iter4 0.81, 하락 아님 → STOP 신호 없음)
Plan Artifact Hash: sha256 `24649b3eafc328ace209a7450434de68e056f1f696f5c28897dfb39f8207650c` (spec·plan·acceptance·design·research·decision-index 를 이 순서로 이어 붙인 것)
Auditor Version: plan-auditor (card t1469, iter4)

`audited_sha` 는 위임 프롬프트가 알려 준 HEAD 이고, 이 세션이 직접 잰 값이 아니다. 이 세션은 git 이 거부됐다(Gaps 참고). 대신 codex 백엔드의 `build_lag` 보고에 "this tree's HEAD 2e37b963a" 가 찍혀 있어 이를 교차 근거로 삼는다.

Reasoning context ignored per M1 Context Isolation. 운영자 판정(Q1 동결 해제, v3.2.0 이후 착지)과 리더 판정(Q5, Q9)은 다시 따지지 않았다.

## 1. Claim

iter3 blocking 4건은 모두 **실질적으로 닫혔다**.
- **D3-1**: 명령을 전체 경로로 적었고 그대로 실행된다. 다만 AC-015 의 stdout 줄 순서가 축자가 아니어서 잔여가 있다(minor).
- **D3-2**: 원장 범위를 「원장 대상 파일의 모든 절·단위」로 넓혔고, goal-directive 변이를 AC-020(d) 로 넣었다.
- **D3-3**: 앵커 종류 필드를 두고, 종류 불일치와 `companion:` 텍스트 대조를 실패 조건으로 넣었다.
- **D3-4**: AC-021 에 면제 범위와 40 코드 단위 조각 규칙을 넣었다.

optional D3-5~D3-9 도 모두 해소했다. 다만 D3-6 은 다른 줄(L15)에 잔여가 남았다.

이번 회차에 **새 blocking 결함 3건**을 찾았다. 두 required 백엔드(claude 앵커, codex)가 독립적으로 이 중 둘을 짚었고, 이 감사가 원문으로 확인했다.
1. `always:` 위치가 실제 상시 로드 표면의 구성원인지 검사하지 않는다(N4-1).
2. 분할 대상인 기존 companion `kanban-dispatch-detail.md` 안의 `MUST` 의무가 원장 범위 밖이다(N4-2).
3. AC-019 초록 조건에 정의가 사라진 「구속 절」이 남아, REQ-015 보다 좁다(N4-3).

세 건 모두 문서 몇 줄로 고칠 수 있다. 그래도 REQ-012 무손실 보장을 기계적으로 우회할 수 있는 구멍이라 → **FAIL**.

### Must-Pass
- [PASS] MP-1: REQ-ALB-001~025 가 spec.md:L72–L108 에 정의돼 있다. 결번·중복은 없다. 023~025 가 011 과 012 사이(L86–L88)에 있는 점은 iter2 부터의 D-N8 로, 무해하다.
- [PASS] MP-2 (요구 층에서만 판정):
  - 25개 REQ 전부 GEARS 또는 레거시 `shall not` 형이다. 레거시 형은 008·013·014·019 넷이다.
  - REQ-ALB-015(L95) 는 Ubiquitous 에 실패 조건을 열거한 형태다.
  - AC 는 검증 층 Given-When-Then(acceptance §D.1) 이라 이 판정에서 뺐다.
- [PASS] MP-3: spec.md:L1–L16 에 12개 필드가 모두 있다. `version: "0.5.0"` 은 인용됐고, `status: draft`, `priority: P1`, `lifecycle: spec-anchored`, `tags` 는 문자열이다. 거부 별칭은 없다.
- [N/A→PASS] MP-4: 다언어 툴링 SPEC 이 아니다. REQ-ALB-019 가 16개 언어 중립성을 요구한다.
- [PASS] MP-5: 참조 SPEC 5개(DIET-001·DIET-002·HEADROOM-001·INSTRUCTION-BUDGET-SCOPE-001·INSTRUCTIONS-BUDGET-001) 모두 `status: completed` 다(`grep -m1 '^status:'` 5줄 관측). retired·superseded·archived 는 없으므로 D7 BLOCKING 도 없다.
- [PASS] MP-6: `grep -c syscall spec.md` → `0`.
- [PASS] MP-7: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` 가 출력 없이 exit 1 이다.
- [PASS, 잔여 minor] MP-8: 대상 트리 루트에서 RB 명령을 표에 적힌 그대로 재실행했다.
  - AC-001·002·003·004·005·006·007 은 `grep -rl TestDeployedAlwaysLoadedCharBudget internal/template` 이 출력 없이 exit 1 이다. 재현된다.
  - AC-008 은 `…/kanban-dispatch.md:1` / `…/cross-session-messaging.md:1` 를 내고 exit 0 이다. 순서까지 표와 같아 축자다.
  - AC-009, AC-014, AC-023 은 출력 없이 exit 1 이다. 재현된다.
  - AC-015 는 `/usr/bin/grep` 으로 돌리면 `manager-lead.md:0` / `moai-kanban-foreman/SKILL.md:0` / `gtd.md:0` 를 내고 exit 1 이다. RED 는 재현된다. 다만 표(L36)는 SKILL → manager-lead → gtd 순서로 적었다. 줄 내용은 바이트 단위로 같고 순서만 다르다 → N4-4 (minor, blocking class — L14 의 「출력은 축자」 계약과 어긋난다).
  - AC-016~018 은 AC-015 를 승계한다.
  - AC-019 는 stdout 없이 stderr `ls: internal/template/testdata/binding_ledger.json: No such file or directory`, exit 1 이다. 표기까지 정확하다.
  - AC-024 는 `1`, exit 0 이다. AC-025 는 `2`, exit 0 이다.

  인용된 RED 는 모두 재현되므로 MP-8 의 판정 조건(RED 재현)은 충족한다. 순서 불일치는 문서 정합 결함으로 따로 접는다.
- [PASS] MP-9: CN-4 awk verb 는 세션 가드가 `awk -f` 를 거부해 실행하지 못했다(GAP). 그래서 손으로 대조했다.
  - plan 마일스톤 순서: M0(L55) → M1(L68) → M2(L75) → M3(L85) → M4(L91) → M5(L96). `Exit:` 바인딩은 0개다.
  - acceptance 의 영문 순서 키워드는 L46 `--first-parent` 의 `first` 1건뿐이다. AC-025 「재고정 커밋이 그 흡수 뒤 첫 규칙 편집보다 앞섬」은 plan §C.2 4단계(L50, 「다른 편집보다 먼저 별도 커밋」)와 합치한다.
  - 한국어 순서 절도 대조했다. AC-002(L23) 「M1 RED 커밋(앵커 뒤 착지)」는 plan L71 과 합치하고, AC-025 「하한 > 115,000 이면 M1 이후 커밋 없이 리드 보고」는 plan L65 와 합치한다.
  - 충돌은 없다.

### Scores
| Dim | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.80 | 0.75 | 개선: 펜스 불투명 규칙이 §B L58 에 들어갔고, 진입점 식별자 형식이 L62 에 정의됐다. 감점: AC-019(L40) 에 정의 없는 「구속 절」이 남았고, L15 「AC-ALB-002 RED-now 가 그 실례」는 이제 grep 이라 틀린 참조다 |
| Completeness | 0.75 | 0.75 | 원장 범위가 원장 대상 파일 전 단위로 넓어졌다(L61). 감점: `always:` 위치의 상시 표면 구성원 검사가 없다(N4-1). 분할 대상 companion 의 기존 `MUST` 의무가 범위 밖이다(N4-2) |
| Testability | 0.80 | 0.75 | AC-020(c)(d) 는 이제 REQ-015 조건(L95, 종류 불일치·텍스트 부재)으로 구현할 수 있다. AC-021 은 정상 원장에서 통과한다. 감점: AC-019 는 REQ-015 보다 좁은 단정이다. `always:` 원본 잔류 변이가 실패하지 않는다 |
| Traceability | 0.90 | 1.00 근접 | 수기 대조 결과 REQ-001~025 전부 AC 가 1개 이상이고, 고아 REQ 는 0이다(AC-019→012·015·025, AC-020→015·012, AC-021→013, AC-025→021·022·015). 감점: AC-019 의 범위 어휘가 REQ-015 와 다르다 |

평균은 0.81 이고 Tier L 문턱 0.85 에 못 미친다. iter3 0.80 보다 올랐으므로 LEAN STOP 신호는 없다.

## 2. Evidence

이 감사가 직접 실행·관측한 것은 다음과 같다(대상 트리 루트에서, git 이 아닌 명령만).
- RED-now 재실행 출력: 위 MP-8 항목에 축자로 적었다.
- `type grep` → `grep is a shell function from …/shell-snapshots/snapshot-zsh-….sh`.
  - 래퍼 첫 실행에서는 AC-015 출력이 manager-lead → gtd → SKILL 순서였고, 다음 3회는 인자 순서였다. `/usr/bin/grep` 은 인자 순서(manager-lead → SKILL → gtd)를 낸다.
  - 표의 순서(SKILL → manager-lead → gtd)는 관측한 어느 순서와도 다르다.
- 13개 경로 `grep -l '^paths:'`(AC-022) → 출력 없음, exit 1.
- `grep -n '구속 절' *.md` 결과:
  - acceptance.md:40 (AC-019 초록 조건) — §B 에 그 용어 정의가 없다(spec L58–L64 를 Read 로 확인).
  - design.md:28 (근거 문장, 낡음).
  - spec.md:25 (HISTORY, 정상).
- goal-directive.md:L25 「Arming a goal does not authorize autonomous run-phase entry …」, L35 「Safety boundary unchanged」가 있고, `[HARD]`·`MUST` 는 0줄이다. AC-020(d)·design L29 가 가리키는 대상이 실재한다.
- cross-session-messaging.md:L41·L44 `STOPPED_TEAMMATE_VIOLATION` 문단이 실재한다(AC-020(c)).
- kanban-dispatch-detail.md:
  - L1–L4 frontmatter 에 최상위 `paths:` 가 있다. 따라서 spec L61 정의상 원장 대상 파일이 아니다(상시 표면 아님).
  - L175 「**Write-capable sub-agents spawned in parallel MUST carry `isolation: "worktree"`**」, L266 「Contents are never printed and never committed」, L268 「Nothing is ever restored, reverted, or deleted」.
  - `grep -c -e '[HARD]' -e MUST` → `9`.
  - plan.md:L87 은 M3 에서 이 파일의 「40,000 미만 분할」을 요구한다.
- spec.md:L95 REQ-ALB-015 의 위치 검사는 「`always:` 또는 `companion:` 행의 변경 후 텍스트가 그 배포 파일에 없을 때」뿐이다. `always:<경로>` 가 배포 표면 구성원인지는 묻지 않는다.
  - spec.md:L63 위치 어휘는 `always:<배포 경로>` 의 경로를 아무것으로도 제한하지 않는다.
  - Q5 로 역할 한정 규칙의 전체 본문은 같은 경로에 `paths:` 를 단 채 남는다(design §3 L43).
- 결정 수 정정: spec L159 「결정 9건(해소 6: Q1·Q2·Q5·Q6·Q7·Q9 …)」 ↔ decision-index Q1~Q9 의 RESOLVED 6, OUT OF SCOPE 1, EVIDENCE-NEEDED 2. 일치한다.
- research.md:L11 은 §1.1 이 `d7112d005` 라고 명시하고, progress.md:L3 은 「plan base … run M0 re-anchors」다. D3-8 이 해소됐다.
- 다른 수정 확인: AC-004(L25) RED 는 이제 grep 이다(D3-9 해소). AC-002(L23) 에 「17개」 하드코딩은 없고 「원장 머리의 구성원 목록」이다.

audit_multi (project_root=agent-a6d780d7ba07bc24c, target=baseBranch):
- `overall_verdict: fail`, `disagreement_flag: false`, `participant_count: 2`, `fail_open_backends: [glm]`.
- `residual_risk_note`: 「required-backend FAIL: claude, codex」.
- codex 지적 4건:
  - [P2] `always:` 표면 구성원 미검사 (spec:95)
  - [P2] 기존 companion 의무 원장 제외 (spec:61, detail L175·L267)
  - [P2] AC-019 범위 미반영 (acceptance:40)
  - [P3] AC-015 출력 순서 (acceptance:36)
- codex 는 AC-015 명령을 그대로 재실행해 manager-lead → SKILL → gtd 순서와 exit 1 을 보고했다. 이 감사의 `/usr/bin/grep` 관측과 같다.
- glm 은 「z.ai response carried no text content」로 inconclusive 였다.
- `audit_receipt` 는 발급되지 않았다.

## 3. Baseline-attribution

- 대상: 워크트리 agent-a6d780d7ba07bc24c, 브랜치 `WT-always-loaded-budget`. HEAD `2e37b963a` 는 codex 백엔드 `build_lag` 문면("this tree's HEAD 2e37b963a")과 위임 프롬프트로 확인했다. 이 세션의 git 직접 관측은 아니다.
- RED-now 와 원문 대조는 이 세션이 같은 트리의 작업본을 직접 읽고 실행한 것이다.
- 템플릿 바이트가 `b5815ca80` 과 같다는 문서 핀(acceptance L14)은 이 세션이 다시 재지 못했다(git 거부). 다만 모든 RED 명령이 기록대로 재현됐으므로, 그 명령들이 읽는 파일은 핀과 같은 상태로 관측됐다.
- moai MCP 서버 빌드 `45600e4ee` 는 HEAD 의 조상이다(build_lag). audit_multi 의 판정은 diff 와 트리 파일 기반이라 영향이 없다.

## 4. Defects (structured defect-list)

**새로 찾은 결함 (iter4)**

N4-1. ALWAYS-LOCATION-MEMBERSHIP-UNCHECKED (신규)
- 위치: spec.md:L63, L95. design.md:L82.
- 문제: `always:<배포 경로>` 의 경로가 배포 상시 표면의 구성원인지 검사하는 조건이 없다. Q5 로 역할 한정 규칙의 전체 본문은 같은 경로에 최상위 `paths:` 를 달고 남는다. 그래서 binding 행을 `always:.claude/rules/moai/workflow/kanban-dispatch.md`(전체 본문, 이제 paths: 한정) 로 적으면, 변경 후 텍스트가 그 파일에 있으니 REQ-015 의 모든 조건을 통과한다. 그런데도 그 의무는 stub 에도 역할 core 에도 없다. 바로 REQ-012 의 무손실 보장이 막아야 할 형태다.
- Severity: major. Class: blocking.
- Required fix: REQ-ALB-015 에 실패 조건 두 개를 추가한다.
  - 「`always:` 행의 경로가 예산 테스트가 도출한 배포 표면 구성원 목록(REQ-ALB-004)에 없을 때」.
  - 「`companion:` 행의 경로가 최상위 `paths:` 를 가진 배포 파일이 아닐 때」.
- 아울러 AC-019 초록 조건에 변이 픽스처를 하나 넣는다: binding 행 위치를 역할 한정 규칙 전체 본문 경로로 바꾼 변이 → `--- FAIL`.

N4-2. EDITED-COMPANION-OBLIGATIONS-OUT-OF-SCOPE (신규)
- 위치: spec.md:L61. plan.md:L87. 원문 kanban-dispatch-detail.md:L175, L266, L268.
- 문제: 원장 대상 파일은 「상시 표면 파일 가운데 이 SPEC 이 편집하는 파일」로 정의돼 있다. 그런데 이 SPEC 은 REQ-003·M3 으로 `paths:` companion `kanban-dispatch-detail.md`(MUST 류 9줄)를 분할한다. 그 안의 기존 구속 의무(병렬 쓰기 에이전트 worktree 격리, 설정 파일 내용 출력·커밋 금지, 자동 복원 금지)는 분할 중에 사라져도 어느 검사에도 걸리지 않는다. 운영자 Q1 조건(「구속 의무는 하나도 떨어뜨리지 않는다」)은 이 SPEC 의 편집 전체에 걸린다.
- Severity: major. Class: blocking.
- Required fix: 원장 대상 파일 정의를 「이 SPEC 이 편집하는 배포 규칙 파일 전부(상시 표면과 기존 companion 포함)」로 넓힌다. 기존 companion 의 binding·normative 단위는 위치 `companion:` 를 허용하는 예외 행으로 원장에 넣는다(REQ-013 은 「새로 옮기는 것」만 금하므로, 원래 companion 에 있던 단위가 분할 후 companion 에 남는 것은 무손실 조건으로만 검사한다).
- 최소 대안: M3 분할 대상 companion 에 대해 「분할 전 단위 집합 = 분할 후 companion 단위 합집합(축자)」 검사를 AC-023 또는 AC-021 에 추가한다. 새 AC 번호는 필요 없다.

N4-3. AC019-STALE-SCOPE-TERM (신규. iter3 D3-2 수리의 미반영 잔여)
- 위치: acceptance.md:L40 (+ design.md:L28 근거 문장).
- 문제: 초록 조건이 「앵커의 구속 절 단위 전부에 행」이다. §B 에서 「구속 절」 정의는 삭제됐고, REQ-015 는 「every ledger-scope file 의 모든 단위」를 요구한다. 구현자는 토큰 있는 절만 검사하는 테스트로 AC-019 를 초록으로 만들 수 있다. AC-020(d) 는 goal-directive 문단 하나만 덮는다.
- Severity: minor. Class: blocking (내부 정합).
- Required fix: 「원장 대상 파일 전부의 모든 단위(spec §B 단위 경계)에 정확히 1행」으로 고친다. design L28 「구속 절 안에 있으므로」는 「원장 대상 파일 안에 있으므로」로 고친다.

N4-4. AC015-STDOUT-ORDER (D3-1 잔여)
- 위치: acceptance.md:L36 (+ L37–L39 인용).
- 문제: 표의 stdout 순서(SKILL → manager-lead → gtd)가 실제 순서(manager-lead → SKILL → gtd, `/usr/bin/grep` 과 codex 가 모두 관측)와 다르다. L14 「출력은 축자」 계약을 어긴다. RED 자체는 재현된다.
- Severity: minor. Class: blocking (문서가 스스로 정한 계약).
- Required fix: 실제 순서대로 다시 적는다. 이 환경의 `grep` 은 쉘 함수 래퍼라 순서가 흔들릴 수 있으므로, RED-now 명령은 `/usr/bin/grep` 으로 적거나 순서 무관 비교라고 명시하는 것을 고려한다.

N4-5. L15-STALE-EXAMPLE-REF (D3-6 잔여)
- 위치: acceptance.md:L15.
- 문제: 「(아래 AC-ALB-002 RED-now 가 그 실례)」 — AC-002 RED-now 는 이제 `grep` 이다. `ok` 줄 실례는 문서에 없다.
- Severity: minor. Class: optional.
- Required fix: 괄호를 지우거나 L17 로 참조를 돌린다.

N4-6. ANCHOR-KIND-SAME-ROW (D3-3 잔여)
- 위치: spec.md:L62, design.md:L31.
- 문제: 앵커 종류가 같은 원장 행의 열이다. 그래서 `kind` 와 `anchor kind` 를 같은 비재고정 커밋에서 함께 바꾸면 테스트는 통과한다. 「재고정 커밋에서만 바뀜」은 검사되지 않는 규범이다. AC-020(c) 가 정의한 변이(종류만 변경)는 정확히 실패하므로 iter3 의 핵심 결함은 닫혔다.
- Severity: minor. Class: optional.
- Required fix: (선택) 앵커 종류를 별도 고정물로 두거나, AC-025 에 「anchor kind 변경은 재고정 커밋에만 있음」 git log 대조를 추가한다. 그렇지 않으면 잔여 위험으로 둔다.

N4-7. FRAGMENT-RULE-SKILLS (D3-4 잔여)
- 위치: acceptance.md:L42.
- 문제: 조각 규칙 (2) 는 companion 만 검사하고, REQ-013 이 함께 금지하는 skill 표면은 검사하지 않는다.
- Severity: minor. Class: optional.
- Required fix: (선택) 조각 검사 대상에 배포 skill 파일을 더한다.

**iter2 부터 유지 중인 결함**

D-N8 — REQ 순서. optional, 무해. 유지한다.

## Regression Check (iter3 → iter4)
- D3-1 RED-NOW-LITERAL — **RESOLVED (잔여 N4-4 minor)**. AC-008·015 를 전체 경로로 적었다. AC-008 은 순서까지 축자다. AC-019 는 stderr 로 표기했다. 재실행은 모두 기록 exit 와 같다. 남은 것은 AC-015 줄 순서뿐이다.
- D3-2 TOKENLESS-SECTION-OBLIGATIONS — **RESOLVED (잔여 N4-3)**. spec L61 원장 대상 파일 = 편집하는 상시 파일의 모든 절·단위, L60 규범 문단 정의에 goal-directive 예, REQ-012/015 문면, design L29, plan L59, AC-020(d)(L41) 삭제·companion 이동 변이 각각 `--- FAIL` 을 확인했다. AC-019 문구는 아직 옛 범위다. 새로 드러난 것은 상시 표면 밖 편집 파일 범위(N4-2)다.
- D3-3 KIND-RECLASSIFICATION-UNGUARDED — **RESOLVED (잔여 N4-6 optional)**. REQ-015(L95) 에 「row's kind differs from its anchor kind」와 「`companion:` row's after-text absent」가 들어갔다. §D.1(L51) 도 AC 본문과 같아졌다.
- D3-4 AC021-LOCATION-CONTRADICTION — **RESOLVED (잔여 N4-7 optional)**. AC-021(L42) (1) 은 binding·normative 로 한정하고, rationale→companion 과 role-core 를 명시적으로 면제한다. (2) 는 40 UTF-16 코드 단위 이상인 줄을 조각으로 삼고, 앵커 시점에 이미 있던 줄은 제외한다.
- D3-5 — RESOLVED. spec L62·design L33–L35 식별자 형식, REQ-015 레지스트리·표지 파일 해석 조건.
- D3-6 — RESOLVED at L17. 잔여는 L15(N4-5).
- D3-7 — RESOLVED. spec L58 펜스 불투명, AC-020(a) 펜스 픽스처, §D.2 L67 은 §B 인용.
- D3-8 — RESOLVED (spec L159, AC-002, research L11, progress L3).
- D3-9 — RESOLVED (AC-004 RED = grep).

Stagnation: 네 회차 모두에 같은 모습으로 남은 결함은 없다. 원장 범위 계열(D3 → N3 → D3-2 → N4-2)은 회차마다 다른 면으로 옮겨 가며 줄고 있다. 다만 원장 범위 정의가 매번 새 우회로를 하나씩 남겼다는 것은 구조적 신호다(Residual-risk 참고).

## Recommendation (최종 에스컬레이션 — 리더/운영자)

iter4 가 허용된 마지막 회차다. 따라서 이 보고는 에스컬레이션으로 끝난다. 리더가 운영자 채널로 올릴 선택지는 세 가지다.
1. **PASS-with-debt**: N4-1~N4-4 를 run M0 의 첫 커밋(원장 기준선 커밋보다 앞)에서 문서 수리하는 조건으로 진행한다.
2. **범위 축소**: `kanban-dispatch-detail.md` 분할(REQ-003 의 이 파일 적용)을 별도 SPEC 으로 떼면 N4-2 가 사라진다.
3. **명시적 추가 연장**.

감사자 관점에서 blocking 4건은 모두 문서 몇 줄 수리이고, 새 REQ·AC 번호가 필요 없다(25/25 유지).
1. N4-1: REQ-015 에 `always:` 표면 구성원 조건과 `companion:` paths: 조건을 추가한다. AC-019 에 원본 잔류 변이를 넣는다.
2. N4-2: 원장 대상 파일에 편집하는 기존 companion 을 포함하거나, 분할 무손실 검사를 AC-023/021 에 넣는다.
3. N4-3: AC-019 「구속 절」을 「원장 대상 파일의 모든 단위」로 바꾼다. design L28 도 정정한다.
4. N4-4: AC-015 stdout 을 실제 순서로 적는다.

N4-5~N4-7 은 optional 이다.

## 5. Gaps
- 이 세션의 Bash 는 git 이 들어간 명령(`-C` 또는 `cd` 뒤 git), `awk -f`, 복합 파이프라인이 워크트리 격리 가드에 거부됐다. 거부된 것과 대체 수단은 다음과 같다.
  - HEAD·`git status`·`git diff b5815ca80..HEAD -- internal/template/templates`·iter3 대비 문서 diff(`6ade6fe38..2e37b963a`) — 대체: codex build_lag 의 HEAD 문면, 그리고 HISTORY v0.5.0(spec L24) 과 본문 직독.
  - 추적성 awk verb — 대체: AC 표 수기 대조. COLLECTED 수치는 없지만 REQ 25개를 손으로 셌다.
  - CN-4 awk verb — 대체: 마일스톤 순서와 순서 절 수기 대조.
  - D7 verb 일부 — 대체: 참조 SPEC 5개 `status:` 를 직접 grep.
- RED-now 명령 자체는 git 이 아니어서 모두 직접 재실행됐다.
- Glob 도구는 이 세션에 없다.
- glm inconclusive. `audit_receipt` 는 발급되지 않았다.
- `audited_sha` 는 이 세션이 git 으로 관측한 값이 아니다(위 참고).

## 6. Residual-risk
- audit_multi residual_risk_note (verbatim): "required-backend FAIL: claude, codex".
- 원장 범위 정의는 네 회차 연속 새 우회로를 하나씩 드러냈다(줄 단위 → 토큰 없는 문단 → 토큰 없는 파일 → 표면 밖 위치·편집 companion). N4-1·N4-2 를 고쳐도, 원장 행 분류(normative/rationale)와 `rewrite` 행 의미 보존은 사람 판정으로 남는다.
- research §2 개산상 M0 가 하한 > 115,000 으로 멈출 가능성이 여전히 높다(Q6 경로). 원장 범위를 넓히면(N4-2) M0 작업량이 더 커진다.
- 이 환경의 `grep` 은 쉘 함수 래퍼이고, 출력 순서가 실행마다 달라지는 것을 한 번 관측했다. RED-now 축자 계약은 래퍼와 실제 grep 사이에서 흔들릴 수 있다.

## Iteration History
- iter1 (HEAD 5bfff5682): FAIL 0.69, blocking D1–D8.
- iter2 (HEAD 15de1ad21): FAIL 0.81, blocking D-N1–D-N6.
- iter3 (HEAD 6ade6fe38): FAIL 0.80 (STOP 하락, 상한 도달), blocking D3-1–D3-4.
- iter4 (HEAD 2e37b963a, 리더 연장 Q9): FAIL 0.81. D3-1–D3-9 해소(잔여 minor). 새 blocking 은 N4-1·N4-2(major), N4-3·N4-4(minor), optional 은 N4-5–N4-7. 최종 에스컬레이션.
