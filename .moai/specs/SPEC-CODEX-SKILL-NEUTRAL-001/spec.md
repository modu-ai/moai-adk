---
id: SPEC-CODEX-SKILL-NEUTRAL-001
title: "하네스 중립 지시 계층 — 코덱스에서도 Claude 와 같은 신뢰로 지시를 실행할 수 있게 한다"
version: "0.3.1"
status: completed
created: 2026-08-31
updated: 2026-09-01
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/template
lifecycle: spec-anchored
tier: M
tags: "codex, dual-harness, skills, agents, neutrality, instruction-layer"
related_specs: [SPEC-CODEX-SKILLS-CANONICAL-001, SPEC-CODEX-DUAL-AGENTS-001]
---

# SPEC-CODEX-SKILL-NEUTRAL-001 — 하네스 중립 지시 계층

## HISTORY

- 2026-09-01 (run-phase 문면 정정 배치, v0.3.1) — run-phase 중 리드 판정(2026-09-01 verdict) 4건을 본문에 반영한다. **요구사항·판정 개수 불변**(REQ 15 / AC 13 — 범위·문면 정정이지 예산 변경이 아니다).
  1. **REQ-CSN-003 문면 정정** — 결속표의 행 집합을 "각 중립 능력"(인벤토리)이 아니라 **이 표를 읽는 하네스에 존재하지 않는 모든 `tool_classes` 능력**(파생 기준)으로 못박는다. 오늘의 4행은 결과이지 기준이 아니다. 측정 근거는 신설 §B.D7, 증거 전문은 `.moai/reports/t196/req-csn-003-budget-remeasure.md`.
  2. **AC-CSN-012 무가드 파일 4 → 5** — 어휘 축 리드 판정(11번째 클래스 신설 = `tool_classes` 에 대한 보충)으로 `internal/template/agentemit/agents-codex.yaml` 이 M2 편집 대상에 합류했다. 범위 확장이 무가드 파일을 만들어낸 두 번째 사례다(iter-2 D11 이 규칙 트리 2파일로 첫 번째) — AC 에 재확인 의무를 붙였다.
  3. **AC-CSN-009 셀렉터 명기 의무** — 접두 셀렉터 `-run TestSkillDirToken` 은 **0매치**로 `no tests to run` PASS 를 냈고, 실제 테스트는 `TestSkillTreeHasNoClaudeSkillDirToken`(**1매치**, `internal/template/skill_dir_token_guard_test.go:41`)이다. 판정 기록이 셀렉터와 매치 수를 명시하도록 [HARD] 로 못박았다 — 0매치 셀렉터의 초록은 판정이 아니다.
  4. **plan.md M2·M3 패리티 판별 메모** — `agents-codex.yaml` 은 `rule_template_mirror_test.go` 의 어느 바이트 패리티 목록에도 없다(단일 트리 편집). 판별축은 "로컬 vs 템플릿"이 아니라 **바이트 패리티 등재 여부**이며, 이 구분이 이 카드의 적색 스위트(`TestRuleTemplateMirrorDrift`)를 만든 축이다.
  부수 정정: §A.5 의 `tool_classes` 값 목록 9개 → **10개**(`cross-session-messaging` 누락 — 이 트리 재측정), §E.1 파일 수 14 → **15**(위 2번의 합류). 리드가 준 후보 표 수치(11행 822 B / 4행 391 B)는 이 트리 재측정에서 재현되지 않아 **재측정값으로 갈아썼다**(§B.D7: 11행 797–814 B, 4행 373 B, 여유 201 tokens 는 가드 본인 측정으로 정확히 확인).

- 2026-09-01 (plan-phase, iter-2 최종, v0.3.0) — plan-audit iter-2 **PASS-WITH-DEBT 0.800** 이 남긴 blocking 5건을 닫는다. 최종 반복이므로 이 판본이 run-phase 로 넘어간다. **요구사항·판정 개수는 불변**(REQ 15 / AC 13) — 전부 기존 항목의 범위·문면 수정이다.

  **D11 — iter-2 수리가 만든 결함.** D10 을 닫으면서 규칙 트리 2파일을 범위에 들였는데, 그 둘이 CI 가드의 클래스 범위(스킬 트리 한정)에도 AC-CSN-012 의 파일 범위(`AGENTS.md` 2개)에도 안 들어갔다. 감사가 프로브로 실증했고 대조군까지 붙였다 — 같은 경로에서 이 SPEC 의 토큰은 안 잡히고 날짜·SHA 는 잡힌다. 즉 경로 도달 실패가 아니라 **클래스 불일치**다. 재도달 가능한 mutant 이 구체적이었다: M3 이 규범 문장을 뒤집을 때 **왜 뒤집었는지를 SPEC ID 로 적는 것**이 가장 자연스러운 덧붙임인데, 그것이 REQ-CSN-013 위반이면서 아무 기계도 잡지 못한다. **iter-1 D1 과 같은 형태가 범위만 줄여 재발한 것이다.** AC-CSN-012 의 파일 목록을 4개로 넓히고 측정된 사전값(`skill-authoring.md` 2 / `worktree-integration.md` 0)과 함께 pre/post 쌍으로 고쳤다. 사전값이 0 이 아닌 것은 그 2건이 프론트매터 예시 안의 날짜 리터럴 — 정당한 문서 내용이고 지울 대상이 아니다.

  **D13(승격) + D16 — 자기참조 수치가 계통 결함이었다.** AC-CSN-012 의 양성 대조값 `34` 는 **어느 단위로도 재현되지 않는다**(현재 `grep -cE` 45 / `grep -coE` 56; `34` 는 REQ/AC 팔 하나만 줄 단위로 센 값). 스테일 위에 단위 불일치가 겹친 것이고, 그 자리가 하필 **이 SPEC 의 공허성 방지 기제**여서 감사가 blocking 으로 승격했다. 수치를 지우고 "판정과 같은 명령으로 실행해 0 이 아닐 것"으로 바꿨다. 감사가 판별자를 더 날카롭게 정정해 주었다 — 부패한 3건과 멀쩡한 15건을 가르는 것은 "명령+출력이냐"가 아니라 **측정의 주어가 이 SPEC 자신의 산출물을 포함하느냐**다(자기주어 3/3 부패, 외부주어 0/15). 기제는 부주의가 아니라 문서 형식이다: **자기참조 측정값은 그 값을 인용한 문서를 편집하는 행위 자체가 무효화**하므로 쓰는 순간이 곧 깨지는 순간이고, 그래서 셋 다 lint·MUST-PASS·자체 스윕을 통과했다. 규율을 `acceptance.md` §D.4 와 `plan.md` §G AP-14 에 걸었다 — run-phase 가 `progress.md` 와 §E.2 를 쓰므로 **구조상 같은 형태를 계속 생산하기 때문**이다.

  **D12 — 잘못된 표제가 문서 하나 건너 살아남았다.** `plan.md` §A 가 "**측정 결과** … 둘 다 **조용히** 깨진다"로 적혀 있었는데, "조용히 깨진다"는 §A.4 가 방금 `추론 — 미관측` 으로 이름 붙인 바로 그 주장이고 plan 자신의 AP-1 도 추론이라 부른다. iter-1 D5 가 지목한 형태가 그대로 남았고, 하필 **run-phase 독자가 가장 먼저 만나는 서술**이었다. 측정된 부분(결합 두 가지와 자취 크기)과 추론된 부분(실패 방식)을 갈라 적었다.

  **D14·D15 채택.** 16진수 팔이 평범한 영어 단어에 걸린다(실측: `defaced`·`feedbed`) — 거짓 적색을 진짜 적색과 한 단계에서 가르도록 실패 경로에서 `-c` 를 빼고 매치 텍스트를 기록하게 했다. `worktree-integration.md:386` 은 AC 가 양쪽을 허용하는데 plan M3 는 수정을 지시해 **손대지 않고도 통과하는 틈**이 있었다 — 기대 분류를 표로 못박았다. M1 의 REQ-CSN-009 닫힘은 `기록으로 닫음` 이라 주석했다(cwd 에 *대한* 증거가 아니라 cwd 를 *적어 두었다*는 증거).

  **감사가 정정해 준 내 판단 하나 — 과소주장이었다.** `plan.md` 의 "날짜·SHA 클래스는 프로브 안 했으니 주장 안 함" 자제는 방향은 옳았으나 **있는 가드를 없는 것처럼 읽히게** 했다. 소스를 직접 읽어 정정했다: `S1-internal-date`·`S2-short-sha-sentence-final` 은 `skillBodyScoped` 를 설정하지 않고, CI 가 `MOAI_TEMPLATE_LEAK_STRICT: '1'` 로 strict tier 를 돌린다 — **날짜와 짧은 SHA 는 전 트리에서 실제로 가드된다.** 무가드인 것은 SPEC-ID·REQ/AC 토큰 클래스뿐이다.

  **감사가 확인해 준 것은 건드리지 않았다** — `:219` 의 사실/규범 분리(원문 표 문맥까지 읽어 확인됨: 6행 표의 한 행, "Available Since" 열 보유 — 지우면 표가 **틀려진다**), AC-CSN-011 이 개수로 환원되지 않는 진짜 집합 판정이라는 것, D3 종결의 전수 재도출(14 + 부채 1 = 15), D2 의 라인 핀 정확성.

- 2026-09-01 (plan-phase, iter-2, v0.2.0) — plan-audit iter-1 **FAIL 0.75** 의 blocking 7건을 닫는다. MUST-PASS 7개는 전부 통과했고 결함은 판정 계층에 몰려 있었으므로 델타 수정이며, 살아남은 절반은 손대지 않았다. 요구사항 14 → **15**, 판정 10 → **13**.

  **가장 무거운 것 — D10: 토큰의 실제 자취가 이 SPEC 이 센 것보다 넓다.** 템플릿 트리 전체 **50줄 / 11파일**인데 이 SPEC 은 46줄 / 9파일만 세고 있었다. 남은 4줄 2파일이 **배포되는 규칙 트리**에 있고, 그중 `skill-authoring.md:226` 은 스킬 저자에게 "`${CLAUDE_SKILL_DIR}` 를 **상대 경로 대신** 쓰라, 그쪽이 더 믿을 만하다"고 가르친다 — §B.D5 가 채택한 설계를 정확히 뒤집은 문장이 규범으로, 모든 사용자 프로젝트에 실려 나간다. **범위 안으로 들였다**(REQ-CSN-015, §A.8). 사유는 §B.D6.

  **D1 — 무가드인 요구사항을 "가드가 잡는다"고 적어 놓았었다.** `plan.md` §B.3 과 AP-7 이 템플릿 중립성 CI 가드가 결속표의 SPEC ID·REQ 토큰을 잡는다고 적었는데 **거짓**이다. 잡을 수 있는 클래스는 `skillBodyScoped: true` 라 `.claude/skills/` 아래에서만 발화하고, 전 트리 클래스는 접두 한정(`SPEC-(V3R[2-6]|AGENCY|WORKTREE)-` · `(REQ|AC)-(ATR|WO|COORD|UNP|LNC|TII|HRN|ORC)-`)이라 이 SPEC 자신의 토큰에 둘 다 안 맞는다. 감사가 mutant probe 로 실측했고, 본 SPEC 은 클래스 표를 직접 읽어 같은 결론에 도달했다. 게다가 `acceptance.md` §D.2 가 REQ-CSN-013 을 AC-CSN-005 에 매핑하고 있었는데 그 AC 는 두 사본 동일성과 바이트 상한만 본다 — **위반하면서 통과하는 mutant 가 성립했다.** `plan.md` 를 사실대로 정정하고 REQ-CSN-013 에 전용 판정(AC-CSN-012)을 주었다. 날짜·SHA 클래스는 프로브하지 않았으므로 그쪽에 대해서는 아무 주장도 하지 않는다.

  **D2 — 결론은 맞았으나 근거가 틀렸다.** §A.7 은 상대 경로가 두 하네스에서 같게 해석되는 이유를 "미러가 심볼릭 링크라서"로 댔다. **미러 모드와 무관하다** — 루트 기준 경로는 정본을 직접 지목하므로 미러가 링크든 복사든 부재든 상관없다. 실제로 기대는 전제는 **읽는 프로세스의 cwd 가 프로젝트 루트라는 것**이고, 그 전제는 `internal/cli/codex_launcher.go:245-250` 한 곳에서만 지지되며 루트 해석 실패 시 프로세스 cwd 로 강등된다. 귀결로 REQ-CSN-009 는 **깨질 수 없는 팔**(복사 폴백)에 관측 의무를 걸고 정작 깨질 수 있는 팔은 안 묶고 있었다. §A.7 을 다시 세우고 REQ-CSN-009 를 cwd 팔로 재조준했다.

  **D5 — "실측" 표제 아래 미관측 추론이 있었다.** §A.4 의 줄 수(6/40)는 실측이지만 "실패 방식" 열은 관측된 적이 없다. 이 SPEC 은 축 A 의 같은 계열 추론을 REQ-CSN-001 로 묶어 놓고 **구조가 동일한 축 B 의 이 주장은 안 묶었다.** 열에 추론 표시를 달고, §B.D4 의 기각이 이 추론에 기대지 않는다는 것을 본문에 못박았다 — 빈 전개는 시끄럽든 조용하든 틀렸다.

  **D3·D4·D6** — plan 마일스톤 닫힘 조건이 acceptance §D.3 의 부채 선언과 모순됐고(REQ-CSN-009·012), AC-CSN-010 의 `<base>` 가 미해결 자리표시자였으며, M1 blocker 절이 관측 반증 시 **설계 기록을 다시 열라는 지시**를 빠뜨렸다. 셋 다 닫았다.

  **감사가 확인해 준 것은 건드리지 않았다** — §B.D3 의 (c) 기각(코드까지 따라가 확인됨), §D.3 의 부채 명시(정직한 범위 설정으로 판정됨), AC-CSN-009 의 RED 관측 절차(올바르게 구성됨).

  **리드 가설 하나가 반증됐고 그 정정을 받았다.** `renderer.go` passthrough 는 중괄호 형태도 같은 등록을 탄다(`renderer.go:110-113` 이 재구성). 다만 그 등록은 **검증 억제 목록**이지 치환 경로가 아니라 "런타임 해석용으로 남겼다"까지만 증명하고 "코덱스에서 미설정"은 증명하지 않는다 — 후자는 런처의 exporter 부재가 진다. §A.4 에서 약한 근거가 앞에 서 있던 순서를 바꿨다.

- 2026-08-31 (plan-phase, iter-1, v0.1.0) — Tier M 최초 작성. 카드 t196 의 전제 4건을 이 트리(`297a21ea7`)에서 전부 재측정했고 **2건을 반증했다**(스킬 21종 → 34종, 도구명 참조 9종 → 14종). 세 번째는 단위 불일치였고 두 값 모두 유효하다. 재측정 전문은 `.moai/reports/t196/premise-remeasure.md`.

  **카드가 틀린 구조를 하나 더 정정한다.** 카드는 "미러된 스킬"을 중립화 대상으로 지목하지만 **`.codex/skills/` 는 존재하지 않는다** — 코덱스가 읽는 스킬은 `.claude/skills/**` 를 가리키는 심볼릭 링크이며 정본과 바이트 동일하다. 따라서 스킬 축의 편집은 코덱스 쪽 손질이 아니라 **Claude 쪽 정본 편집**이고 폭발 반경이 넓다. 에이전트 축만이 진짜 이중 발행이다. 두 축은 처방도 폭발 반경도 다르므로 요구사항 군과 마일스톤을 분리했다(§A.3, §C).

  **Tier 를 카드 등록값 M~L 에서 M 으로 확정했다.** 근거는 §E.1 의 파일 수 산정(권장안 채택 시 12파일)이며, Tier L 로 밀어 올렸을 유일한 항목 — 도구명을 담은 14개 `SKILL.md` 전부의 산문 재작성 — 을 §D 에서 명시적으로 범위 밖에 두었기 때문이다. 그 배제는 예산 때문이 아니라 §B.D2 의 설계 판단 때문이다.

  **관측하지 않은 것을 요구사항의 근거로 삼은 자리가 하나 있다.** "코덱스는 없는 도구명을 만나면 조용히 개선한다"는 것은 추론이지 관측이 아니다. 그 추론이 §C 의 우선순위(조용한 실패를 시끄러운 실패보다 먼저 닫는다)를 정하고 있으므로, REQ-CSN-001 이 그 관측을 run-phase 의 선행 의무로 못박았다.

---

## §A. 검증된 기준선 (실측)

전 항목 측정 트리 `.claude/worktrees/t196`, HEAD `297a21ea73b24e6605280625e576555e4316263e`. 명령과 축자 출력은 `.moai/reports/t196/premise-remeasure.md` 에 있다.

### A.1 카드 전제 재측정 — 2건 반증, 1건 단위 불일치, 1건 확인

| # | 카드 주장 | 측정값 | 판정 |
|---|---|---|---|
| ① | 미러 스킬 21종 | **34** | 반증 |
| ② | 도구명 참조 9종 | **14** | 반증 |
| ③ | `${CLAUDE_SKILL_DIR}` 의존 3종 | `SKILL.md` 단위 **3** / 디렉터리 단위 **4 스킬·9 파일·46 줄** | 단위 불일치, 둘 다 유효 |
| ④ | 에이전트 TOML 11종 전부 | **11/11** (정본 `.md` 도 11/11) | 확인 |

34 는 `SPEC-CODEX-SKILLS-CANONICAL-001` §A.1 이 확정한 값과 일치한다. 그 SPEC 은 자기보다 앞선 두 값(32, 36)을 정정한 이력을 남겼고, 36 은 `ls | wc -l` 이 별칭 때문에 `.` `..` 를 더해 만든 값이었다. 본 SPEC 은 `find` 로 독립 측정해 같은 34 에 도달했다.

③ 의 판정 단위는 **디렉터리 단위(4 스킬 / 9 파일 / 46 줄)** 다. 좁은 단위는 실제로 실행이 깨지는 자리 하나(`moai-workflow-project/references/navigator.md:88`)를 놓친다.

### A.2 코덱스가 읽는 스킬은 정본과 바이트 동일하다 — 설계를 구속하는 사실

`internal/template/templates/.codex/` 아래에는 `agents/` 하나뿐이며 `skills/` 는 없다.

`internal/template/skill_mirror.go` 가 배포 시점에 `.agents/skills/<name>` 을 `../../.claude/skills/<name>` 로 향하는 **상대 심볼릭 링크**로 만든다(`mirrorLinkTarget`). 링크 생성이 불가한 플랫폼에서는 실 디렉터리 복사로 폴백한다(`MirrorModeCopy`). 이 함수는 `Deploy` 당 한 번, `WithSkillMirror` 를 통해 호출된다.

**귀결: 중립화할 "코덱스판 스킬 텍스트"는 존재하지 않는다.** 스킬 본문을 코덱스에 맞게 바꾸는 유일한 방법은 정본 `.claude/skills/**` 를 바꾸는 것이고, 그것은 Claude 쪽 동작에도 그대로 반영된다.

### A.3 두 축은 폭발 반경이 다르다

| 축 | 코덱스가 읽는 실물 | 실제 편집 지점 | 폭발 반경 |
|---|---|---|---|
| **축 A — 능력 이름 결합** (`AskUserQuestion` · `Agent(` · `Skill(` · `Task*`) | 스킬: 정본과 동일 / 에이전트: 생성 TOML | 정본 `.claude/skills/**` + 정본 `.claude/agents/**` | 넓다 (Claude 동작 동반 변경) |
| **축 B — 경로 변수 결합** (`${CLAUDE_SKILL_DIR}`) | 정본과 동일 | 정본 `.claude/skills/**` 9파일 | 중간 (경로 표기만) |

에이전트 TOML 은 `internal/template/agentemit/emit.go` 의 `EmitAll` 이 정본 `.md` 에서 결정적으로 생성한다. `developer_instructions` 는 정본 본문을 **그대로 옮긴다** — 즉 생성기는 변환 지점으로 쓸 수 **있으나** 현재 의도적으로 항등 변환이다.

### A.4 `${CLAUDE_SKILL_DIR}` 은 코덱스에서 빈 문자열이다 — 그리고 깨짐에 두 등급이 있다

**미설정이라는 판단의 근거는 두 개이고, 무게가 다르다. 강한 쪽을 먼저 적는다.**

1. **(강함) `internal/cli/codex_launcher.go` 에 환경변수를 내보내는 코드가 없다** (`grep -n 'Env\|env'` → 출력 없음). 코덱스 아래에서 이 변수를 채우는 주체가 없다.
2. **(약함) `internal/template/renderer.go` 의 `claudeCodePassthroughTokens` 가 `$CLAUDE_SKILL_DIR` 을 미치환 통과시킨다.** 이것이 증명하는 것은 **"런타임 해석용으로 일부러 남겼다"까지**다 — 이 목록은 **검증 억제 목록**이지 치환 경로가 아니며, "코덱스에서 미설정"을 증명하지 않는다. 근거 (1)이 그것을 진다.

(참고: `renderer.go:110-113` 이 등록된 토큰에서 중괄호 형태를 재구성하므로 `$X` 와 `${X}` 가 등록 하나를 함께 탄다 — 두 철자가 별개 등록을 요구한다는 가설은 반증됐다.)

| 등급 | 줄 수 (실측) | 형태 (실측) | 실패 방식 (**추론 — 미관측**) |
|---|---|---|---|
| **HARD (시끄러움)** | **6** | `bash "${CLAUDE_SKILL_DIR}/…"` · `node ${CLAUDE_SKILL_DIR}/…` | 잘못된 경로로 실행 실패. 오류가 화면에 뜰 것으로 본다 |
| **SOFT (조용함)** | **40** | `Read ${CLAUDE_SKILL_DIR}/workflows/plan.md` 형태의 산문 경로 지시 | 실행이 실패하지 않고 모델이 추론으로 메울 것으로 본다 |

[HARD] **마지막 열은 실측이 아니다.** 줄 수와 형태는 이 트리에서 잰 값이지만, "무슨 일이 일어나는가"는 코덱스를 돌려 본 적이 없는 추론이다. §A 의 표제가 "실측"이므로 이 사실을 열 제목에 붙여 둔다 — 축 A 의 같은 계열 추론(REQ-CSN-001)만 묶고 구조가 동일한 이 주장을 안 묶으면 비대칭이다.

HARD 6줄은 2개 스킬 · 3개 스크립트 대상에 걸쳐 있다. SOFT 40줄 중 19줄이 `moai/SKILL.md` 한 파일에 몰려 있다(`Read ${CLAUDE_SKILL_DIR}/workflows/*.md`).

**어느 쪽이 더 큰 위험인가 — 추론이 맞다면 SOFT 다.** HARD 는 실패가 즉시 보이고 사용자가 "코덱스는 이 스킬을 못 돌린다"를 곧바로 배운다. SOFT 는 산출물이 정상처럼 보이고 그것이 워크플로 본문을 읽지 않은 결과라는 사실이 아무 신호도 내지 않는다. 카드가 말하는 "단순 호출은 가능하나 전체 오케스트레이션을 동일 신뢰 불가"는 이 계열이 만드는 상태다. **부재하는 실패 신호는 성공의 근거가 아니다.**

**이 추론이 정하는 것은 우선순위이지 처방이 아니다** — §B.D4 가 그 구분에 선다. 축 A 전체도 같은 계열의 추론 위에 있으며, REQ-CSN-001 이 그것을 관측 의무로 묶는다.

### A.5 중립 어휘는 이미 하나 존재한다 — 새로 만들지 않아도 된다

`internal/template/agentemit/agents-codex.yaml` 의 `tool_classes` 가 Claude 도구 토큰을 하네스 중립 클래스로 이미 매핑한다: `file-read` · `file-write` · `shell` · `web` · `task-list` · `skill-loader` · `subagent-spawn` · `design-sync` · `cross-session-messaging` · `moai-mcp` — **값 집합 10개**(`awk '/^tool_classes:/{f=1;next} /^[^ ]/{f=0} f&&NF==2{gsub(":","",$2);print $2}' internal/template/agentemit/agents-codex.yaml | sort -u | wc -l` → 10, 이 트리 실측). [v0.3.1 정정] 종전 판본은 9개로 적었다 — `cross-session-messaging`(SendMessage·ListAgents 매핑)을 빠뜨린 누락이다.

다만 이 어휘는 **`tools:` 프론트매터에만** 적용된다. 본문 산문은 이 매핑을 통과하지 않는다. 카드가 남긴 구멍이 정확히 여기다.

### A.6 `AGENTS.md` 는 중립 어휘를 이미 선언했고, 결속표를 담을 여유가 있다

`AGENTS.md:14` 는 이미 능력을 중립 이름으로 부른다 — "the question channel, subagent spawning, skills, session handoff" — 그리고 그 구현을 `CLAUDE.md` 로 미룬다. **없는 것은 "하네스가 그 능력을 못 가졌을 때 무엇을 하는가"를 적은 결속표다.**

용량: `AGENTS.md` 14,229 바이트 (루트·템플릿 사본 동일), 상한 `CodexContractByteCeiling = 24576` → **여유 10,347 바이트**. `TestCodexContractByteCeiling` 이 두 사본을 실 트리에서 감시하며 초과 시 빌드를 실패시킨다(코덱스가 꼬리를 조용히 자르므로 다른 신호가 없다).

### A.7 루트 기준 경로가 성립하는 진짜 근거는 cwd 이지 미러 모드가 아니다

**프로젝트 루트 기준 상대 경로 `.claude/skills/<name>/…` 은 두 하네스에서 같은 파일로 해석된다.** 코덱스용 대체 환경변수를 신설할 필요가 없다 — 변수를 **없애는** 쪽이 성립한다.

**단, 그 이유는 미러가 심볼릭 링크라는 것이 아니다.** 루트 기준 경로는 정본 `.claude/skills/<name>/…` 을 **직접 지목**하므로 `.agents/skills/<name>` 이 링크든 실 복사든 아예 없든 해석에 관여하지 않는다. 미러 모드는 이 성질과 무관하다.

실제로 기대는 전제는 하나다 — **읽는 프로세스의 작업 디렉터리가 프로젝트 루트라는 것.**

그 전제는 한 경로에서만 지지되고 다른 경로에서는 방어되지 않는다. `internal/cli/codex_launcher.go:245-250` 이 실행 디렉터리를 프로젝트 루트로 잡는다:

```go
dir := ""
if root, rerr := findProjectRootFn(); rerr == nil && root != "" {
    dir = root
} else if cwd, gerr := os.Getwd(); gerr == nil {
    dir = cwd
}
```

주석(`:242-244`)이 의도를 명시한다 — 하위 디렉터리에서 불러도 루트에서 띄우며, **루트 해석이 실패하면 거부하지 않고 프로세스 cwd 로 강등한다**(`:248-250`). 그리고 `moai codex` 를 거치지 않고 직접 띄운 코덱스 세션은 사용자의 cwd 를 그대로 물려받는다 — 이 저장소에는 그것을 묶는 장치가 없다.

REQ-CSN-009 는 이 팔을 겨눈다. 이전 판본은 복사 폴백 모드를 관측 의무로 걸고 있었는데, 그것은 **어떤 변이로도 깨질 수 없는 팔**이다(경로가 미러에 닿지 않으므로). 깨질 수 있는 팔은 cwd 쪽이다.

**부수 관측**: 복사 폴백에서 미러 본문은 1회차 내용에 고착되지만 그것이 지목하는 경로는 신선한 정본으로 해석된다. 그 섞인 신선도는 현 상태보다 엄격히 낫지만 §B.D5 의 귀결이며 이 SPEC 이 만든 것이 아니다.

이 성질이 형제 카드 t391(`moai codex` 런처)과의 표면 겹침도 해소한다: 런처가 환경변수를 내보내야 할 이유가 사라지므로 본 SPEC 은 `internal/cli/codex_launcher.go` 에 손대지 않는다(읽기만 했다).

### A.8 토큰의 실제 자취는 스킬 트리보다 넓고, 밖에 있는 4줄이 처방과 정반대다

```
$ grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/ | wc -l              → 50
$ grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/.claude/skills | wc -l → 46
$ grep -rl 'CLAUDE_SKILL_DIR' internal/template/templates/ | wc -l              → 11
```

**50줄 / 11파일**이다. §A.1 이 센 46줄 / 9파일은 스킬 트리 부분집합이었다. 나머지 **4줄 / 2파일**은 배포되는 **규칙 트리**에 있다:

```
.claude/rules/moai/development/skill-authoring.md:219
  | `${CLAUDE_SKILL_DIR}` | Absolute path to the skill's own directory | v2.1.69 |
.claude/rules/moai/development/skill-authoring.md:226
  Use `${CLAUDE_SKILL_DIR}` for referencing files within the skill directory
  instead of relative paths. This is more reliable across different invocation contexts.
.claude/rules/moai/development/skill-authoring.md:301
  - Use `${CLAUDE_SKILL_DIR}` for self-referencing paths within skill content
.claude/rules/moai/workflow/worktree-integration.md:386
  | Read-only references | Skills, configs via `${CLAUDE_SKILL_DIR}` | YES | ... |
```

`226` 은 §B.D5 가 채택한 설계를 **정확히 뒤집은 문장**이다 — "상대 경로 **대신** 쓰라, 그쪽이 더 믿을 만하다" — 그것도 스킬 저자에게 주는 규범으로, 모든 사용자 프로젝트에 배포되는 파일 안에서. 그 정당화("more reliable across different invocation contexts")는 §A.4 가 코덱스에서 거짓임을 잰 바로 그 주장이다.

**세 줄의 성격이 다르다.** `219` 는 "Claude Code 가 이런 변수를 제공한다"는 **사실 기술**이며 참이다 — 유지 대상이다. `226` 과 `301` 은 **규범적 선호**이며 뒤집힌 쪽이다. `worktree-integration.md:386` 은 워크트리에서의 읽기 전용 참조 예시로 토큰을 인용한다.

**이 4줄을 그대로 두면 재도입 경로가 열린 채로 남는다.** 46줄을 지우고 REQ-CSN-010 의 가드를 걸어도, 다음 스킬 저자가 이 저장소의 문서화된 규칙을 따르면 가드가 그 저자에게 발화한다. 가드도 옳고 저자도 옳으며, 틀린 것은 규칙이다. 그래서 범위 안으로 들였다(REQ-CSN-015, 사유 §B.D6).

---

## §B. 설계 결정 기록

**결정해야 할 것**: "하네스 중립 지시 계층"이 실제로 무엇인가.

### B.D1 — 채택: 중립 어휘 + `AGENTS.md` 단일 결속표

능력을 중립 이름으로 부르고, "이 하네스에 그 능력이 없으면 무엇을 하는가"를 **한 곳에** 적는다. 그 한 곳은 `AGENTS.md` 다.

채택 근거 셋, 전부 실측에 기반한다:

1. **어휘가 이미 있다.** `agents-codex.yaml` 의 `tool_classes` (§A.5). 두 번째 어휘를 만들면 두 어휘가 갈리는 것을 막을 장치가 없다.
2. **선언 지점이 이미 있다.** `AGENTS.md:14` 가 능력을 중립 이름으로 부르고 있다(§A.6). 빠진 것은 결속표 한 개다.
3. **담을 자리가 있다.** 여유 10,347 바이트, 그리고 그 여유를 지키는 가드가 이미 돌고 있다(§A.6).

`AGENTS.md` 는 코덱스가 실제로 읽는 파일이고 **자기 충족적**이도록 쓰여 있다(다른 지시 파일이 하나도 안 실려도 성립). 결속표가 거기 있으면 코덱스 세션은 그것을 항상 갖는다.

### B.D2 — 기각: 스킬 본문마다 하네스 조건절을 심는 방식

"질문 채널이 있으면 쓰고, 없으면 blocker 보고를 반환한다"를 각 스킬 본문에 적는 방식.

기각 사유: 조건절이 **최대 25개 파일**(스킬 **≤14** + 에이전트 11)에 복제된다. 14 는 §A.1 이 잰 **상한**이고 그중 몇 자리가 실제 도구 호출 지시인지는 분해하지 않았다 — 그러나 아래 (a)(b)(c) 는 전부 개수와 무관하므로 상한이 낮아져도 기각은 그대로 선다. 복제된 규칙은 (a) 감사할 단일 지점이 없고, (b) 새 스킬 저자가 매번 기억해야 하며, (c) 항상 로드되는 스킬 본문을 부풀린다. **§A.4 가 밝힌 것은 "지시가 흩어져 있다"가 아니라 "결속이 어디에도 적혀 있지 않다"이므로, 처방은 흩뿌리기가 아니라 한 곳에 적기다.**

이 기각이 Tier 를 M 으로 유지시킨다 — 채택했다면 25파일이 되어 Tier L 이었다. 다만 **예산이 기각 사유가 아니다**; 위 (a)(b)(c) 가 사유이고 Tier 는 그 결과다.

### B.D3 — 기각: 미러 생산자에서 빌드타임 변환

`skill_mirror.go` 가 코덱스 쪽 사본을 만들 때 식별자를 치환하는 방식.

기각 사유는 측정된 것이다: **바이트 동일 성질이 끝난다.** 현재 미러는 심볼릭 링크이며, 변환을 하려면 **전 플랫폼에서 복사 경로를 강제**해야 한다. 그러면 `SPEC-CODEX-SKILLS-CANONICAL-001` §D 가 잔존 결함으로 명시한 상태 — 복사 미러는 2회차 배포부터 `REQ-CSC-014` 의 "실 항목" 분기에 걸려 건너뛰어지고 1회차 내용에 고착된다 — 가 예외가 아니라 **기본 동작**이 된다. 즉 이 안은 코덱스에 정확한 텍스트를 주려다 **낡은 텍스트를 주는 쪽**으로 귀결한다.

부수적으로 `//go:embed` 는 심볼릭 링크를 무음으로 버리므로(같은 SPEC §A.2) 링크 계층은 이미 취약하다. 그 위에 변환을 얹는 것은 취약한 계층에 의존을 하나 더 다는 일이다.

### B.D4 — 기각: HARD 6줄만 고치기

**기각은 실패 방식과 무관하게 성립한다 — 이것이 1차 근거다.** `${CLAUDE_SKILL_DIR}` 이 코덱스에서 비어 전개되면 그 경로는 **틀렸다.** 시끄럽게 틀렸든 조용하게 틀렸든 틀린 것이고, 40줄을 남기는 안은 틀린 경로 40줄을 남기는 안이다. 이 근거는 §A.4 의 미관측 열에 전혀 기대지 않는다.

**우선순위 근거는 2차이며 추론에 기댄다.** SOFT 계열이 더 위험하다는 판단(§A.4)은 관측되지 않았다. 그것이 정하는 것은 "무엇을 먼저 닫는가"이지 "무엇을 닫는가"가 아니다. **관측이 이 추론을 반증하면 §C 의 우선순위와 이 절의 2차 근거는 다시 열린다**(M1 blocker 절). 1차 근거는 그 관측과 무관하게 남는다.

HARD 6줄은 **채택안에 포함**된다(REQ-CSN-006) — 기각한 것은 "그것만" 하는 범위이지 그 작업이 아니다.

### B.D5 — 축 B 의 채택 형태: 변수를 대체하지 않고 없앤다

§A.7 의 실측에 따라, `${CLAUDE_SKILL_DIR}/X` 는 코덱스용 대체 변수가 아니라 **프로젝트 루트 기준 상대 경로 `.claude/skills/<name>/X`** 로 바꾼다. 두 하네스가 같은 파일에 도달하고, 새 변수도 런처 변경도 필요 없다.

이것은 B.D1 의 어휘 원칙을 경로에 적용한 것이지 별개 결정이 아니다 — 하네스 고유 이름(`CLAUDE_SKILL_DIR`)을 양쪽이 아는 이름(저장소 경로)으로 바꾸는 같은 동작이다.

### B.D6 — 채택: 규칙 트리 4줄을 범위 안으로 들인다

§A.8 의 4줄(규칙 트리)을 범위 밖에 둘 수도 있었다. 범위 안으로 들인 근거 둘:

1. **밖에 두면 완료 상태가 자기모순이다.** §E.2 는 스킬 트리 토큰 0 을 완료로 선언하는데, 그 상태에서 배포된 규칙은 여전히 다음 저자에게 토큰을 쓰라고 가르친다. 그러면 REQ-CSN-010 의 가드는 **저장소의 문서화된 규칙을 따른 저자에게** 발화한다. 재도입 경로가 열린 채 "완료"라고 부르는 것은 완료가 아니다.
2. **§A.1 의 단위 규율이 한 단계 위에서 같은 힘으로 적용된다.** 좁은 단위(`SKILL.md` 3종)를 버리고 넓은 단위(9파일)를 택한 이유는 좁은 쪽이 실제 깨짐 자리를 놓쳤기 때문이었다. 스킬 트리(46) 대 템플릿 트리(50)도 같은 형태이며, 넓은 쪽에 있는 것이 이번에는 깨짐이 아니라 **깨짐을 재생산하는 규범**이다.

비용은 2파일이다. §E.1 의 파일 수가 12 → 14 로 오르며 Tier M 범위(5–15) 안에 남는다.

**세 줄을 같은 처분으로 다루지 않는다.** `skill-authoring.md:219` 는 사실 기술이므로 **유지**하고, `226`·`301` 의 규범적 선호와 `worktree-integration.md:386` 의 예시만 고친다. "규칙 트리 토큰 0" 이라는 목표를 세우지 않는 이유가 이것이다 — 참인 사실 기술을 지우는 것은 문서를 나쁘게 만든다.

**따라서 REQ-CSN-010 의 가드는 스킬 트리에 그대로 둔다.** 전 트리로 넓히면 `219` 에 발화한다. 규칙 트리 쪽은 가드가 아니라 **잔존 줄 집합을 열거한 판정**(AC-CSN-011)으로 닫는다.

### B.D7 — 채택: 결속표의 행 기준은 인벤토리가 아니라 부재다 [v0.3.1, run-phase 리드 판정]

**판정 출처**: run-phase 중 리드 verdict (2026-09-01). 같은 판정이 어휘 축도 확정했다 — 신설 11번째 클래스 `question-channel` 은 `agents-codex.yaml` 의 `tool_classes` 에 대한 **보충**이지 두 번째 어휘가 아니다(REQ-CSN-002 의 금지 대상은 별개 어휘다; 확장은 기존 명명 관례를 따르는 매핑 행 + `classes:` 처분 행으로 들어가고, rationale 필드가 그 클래스가 무엇을 덮는지 서술한다 — 매니페스트 로더가 이 일관성을 강제한다, `internal/template/agentemit/manifest.go:121-124`).

**REQ-CSN-003 을 "각 중립 능력에 대해"로 읽으면 결속표가 능력 인벤토리가 된다.** 그 표는 `AGENTS.md` 에 실리고, `AGENTS.md` 는 always-loaded 측정 표면의 고정 슬롯이다(`internal/config/token_budget_guard.go:195-200`). 이 트리에서 재측정한 예산:

- **여유 201 tokens = 804 bytes** — 가드 본인의 측정: `always-loaded surface = 75799 tokens (budget 76000, headroom 201, 18 entries)` (`go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v` → `token_budget_guard_test.go:69` 로그). `estimateTokens` 는 `len/4`(`token_budget_guard.go:105-107`). `CodexContractByteCeiling`(§A.6 의 여유 10,347 B)은 이 표에 대해 먼저 걸리지 않는다 — **구속 기준은 토큰 예산이다.**
- **인벤토리 형태(11행 = `tool_classes` 값 10개 + `question-channel`)는 문구와 무관하게 여유의 끝에 붙는다.** 3열 후보 표 실측: 가장 압축한 구성 **797 B = 199 tokens(여유 2)**, 보유 능력 행의 (c) 칸에 정직한 문구를 쓴 구성 **814 B = 203 tokens(75,799+203 = 76,002 — 가드 트립)**. 같은 11행이 17바이트 문구 차이로 트립과 비트립을 갈라놓는다는 것은, 인벤토리 형태의 가용성이 셀 문구 운에 맡겨진다는 뜻이다. 회귀 트립와이어의 남은 여유 전부를 오늘의 표 한 장이 소비하는 기준은 기준이 아니라 예산 회피권이다.
- **부재 형태(4행 = 코덱스에 없는 능력만)는 같은 후보 형태로 373 B = 93 tokens — 여유 108 tokens.** 행 수가 부재를 따라가므로 능력 인벤토리가 늘어도 표는 늘지 않는다.

측정 전문(후보 표 텍스트·명령·축자 출력): `.moai/reports/t196/req-csn-003-budget-remeasure.md` · `csn003-table-{11row,11row-honest,4row}.txt`. 바이트 수치는 셀 문구에 좌우되므로 **이 결정은 특정 바이트 수에 기대지 않는다** — 201 tokens 여유가 먼저고, 부재 형태만이 그 여유를 예산이 아니라 설계로 남겨 둔다.

**교리 문장**: 결속표는 **부재를 채우기 위해 존재하는 것이지 능력 인벤토리를 복제하기 위해 존재하는 것이 아니다.** 예산이 무한대라도, 이 하네스에 없는 능력은 **없다는 이유 하나만으로** 정확히 한 행을 얻는다 — 그리고 이 하네스에 있는 능력은 행을 얻지 못한다. 표가 인벤토리를 복제하면 두 비용이 함께 온다: 능력이 바뀔 때마다 표도 고쳐야 하는 유지 부담과, AC-CSN-003 이 판정하는 (c) 칸을 빈 칸이나 "해당 없음"으로 채우는 행이다.

---

## §C. 요구사항 (GEARS)

### C.1 축 A — 능력 이름 결합

- **REQ-CSN-001** — **While** 코덱스가 이름을 아는 도구 없이 지시를 만났을 때의 거동이 관측되지 않은 상태다, run-phase 는 축 A 의 어떤 본문도 편집하기 **전에** 그 거동을 최소 1건 관측하고 그 출력을 증거 경로에 남겨야 한다.
- **REQ-CSN-002** — 하네스 중립 능력 어휘는 `internal/template/agentemit/agents-codex.yaml` 의 `tool_classes` 클래스 이름을 그대로 써야 하며, 두 번째 어휘를 새로 만들어서는 안 된다.
- **REQ-CSN-003** — `AGENTS.md` 는 결속표를 실어야 한다. 표의 행 집합은 **이 표를 읽는 하네스(코덱스)에 존재하지 않는 모든 `tool_classes` 능력**이다 — 파생 기준이지 고정 행 수가 아니다. 현재 측정값 4행은 결과이지 기준이 아니다: 미래의 하네스가 `file-read` 를 잃으면 표에 그 행이 새로 생기고, 코덱스가 어떤 능력을 얻으면 그 행은 표에서 사라진다. 각 행은 (a) 그 능력의 중립 이름, (b) Claude 하네스에서의 구현, (c) **그 능력이 없는 하네스에서 취할 행동** 세 칸을 담는다. 근거(예산 측정)는 §B.D7.
- **REQ-CSN-004** — **When** 하네스에 질문 채널이 없다, 그 하네스의 실행자는 사용자에게 질문하는 대신 blocker 보고를 반환해야 한다. 결속표는 이 규칙을 명시해야 한다.
- **REQ-CSN-005** — 결속표는 `AGENTS.md` 루트 사본과 `internal/template/templates/AGENTS.md` 사본 **양쪽**에 동일 내용으로 실려야 하며, 두 사본은 각각 `CodexContractByteCeiling` 이하로 유지되어야 한다.

### C.2 축 B — 경로 변수 결합

- **REQ-CSN-006** — `internal/template/templates/.claude/skills/**` 안에서 셸 또는 노드 호출의 인자로 쓰인 `${CLAUDE_SKILL_DIR}` 은 프로젝트 루트 기준 상대 경로로 대체되어야 한다.
- **REQ-CSN-007** — `internal/template/templates/.claude/skills/**` 안에서 산문 경로 지시로 쓰인 `${CLAUDE_SKILL_DIR}` 도 같은 형태의 프로젝트 루트 기준 상대 경로로 대체되어야 한다. 시끄러운 자리만 고치고 조용한 자리를 남겨서는 안 된다.
- **REQ-CSN-008** — 대체 후 `internal/template/templates/.claude/skills/**` 전체에 `CLAUDE_SKILL_DIR` 토큰이 **하나도** 남아 있지 않아야 한다.
- **REQ-CSN-009** — 대체된 루트 기준 경로가 기대는 전제는 읽는 프로세스의 작업 디렉터리가 프로젝트 루트라는 것이며, SPEC 은 그 전제가 지지되는 경로와 지지되지 않는 경로를 각각 기록해야 한다. (미러 모드는 이 해석과 무관하며 관측 의무의 대상이 아니다 — §A.7.)

### C.3 교차 관심사

- **REQ-CSN-010** — 회귀 가드는 `internal/template/templates/.claude/skills/**` 에 `CLAUDE_SKILL_DIR` 이 재도입되면 빌드를 실패시켜야 한다.
- **REQ-CSN-011** — **When** 회귀 가드가 도입된다, 그 가드는 감시 대상을 실제로 나쁘게 만들었을 때 붉게 뜨는 것이 관측되어야 한다. 관측 없이 초록인 가드를 통과 근거로 삼아서는 안 된다.
- **REQ-CSN-012** — 모든 편집은 `internal/template/templates/**` 를 원본으로 수행해야 하며, 로컬 `.claude/**` 사본을 원본으로 편집해서는 안 된다.
- **REQ-CSN-013** — `internal/template/templates/**` 에 실리는 어떤 내용도 SPEC ID · REQ 토큰 · 내부 날짜 · 커밋 해시를 담아서는 안 된다. 결속표는 이 제약을 만족하는 형태로 쓰여야 한다.
- **REQ-CSN-014** — 이 SPEC 은 `internal/cli/codex_launcher.go` 를 변경해서는 안 된다.
- **REQ-CSN-015** — 배포되는 규칙 트리에서 `${CLAUDE_SKILL_DIR}` 을 상대 경로보다 선호하라고 지시하는 규범 문장은 제거되거나 채택된 설계에 맞게 바뀌어야 한다. 그 변수를 Claude Code 가 제공한다는 **사실 기술**은 유지되어야 한다.

---

## §D. 범위 밖 (Out of Scope)

### Out of Scope — 도구명을 담은 14개 `SKILL.md` 의 산문 재작성

- 도구 식별자가 등장하는 14개 `SKILL.md` 와 11개 에이전트 `.md` 의 본문 문장을 중립 표현으로 고쳐 쓰는 작업은 이 SPEC 이 하지 않는다. 사유는 §B.D2 — 결속을 한 곳에 적는 것이 처방이고, 25개 파일에 조건절을 복제하는 것은 처방이 아니다.
- `grep -rlE 'Agent\('` 이 세는 14 는 **상한**이다. 그중 몇 자리가 실제 도구 호출 지시이고 몇 자리가 역할 서술(`Agent(general-purpose)` 같은)인지는 이 SPEC 이 분해하지 않았다. 자리별 분해가 필요해지면 별도 카드다.

### Out of Scope — 규칙 트리의 사실 기술 줄

- `skill-authoring.md:219` 의 능력 표 항목(`${CLAUDE_SKILL_DIR}` 이 스킬 디렉터리 절대 경로라는 서술)은 **참이므로 유지**한다. 이 SPEC 은 규칙 트리에서 토큰 0 을 목표로 삼지 않는다 — 참인 사실 기술을 지우는 것은 문서를 나쁘게 만든다(§B.D6).
- 그 결과 REQ-CSN-010 의 기계 가드는 스킬 트리에 한정된다. 규칙 트리에 토큰이 남는 것 자체는 위반이 아니며, 위반은 **규범 문장**이 남는 것이다 — 그 구분은 기계가 아니라 열거된 줄 집합으로 판정한다(AC-CSN-011).

### Out of Scope — 미러 생산자·생성기의 변환 계층

- `internal/template/skill_mirror.go` 에 빌드타임 텍스트 변환을 넣지 않는다(§B.D3 — 바이트 동일 성질과 미러 신선도를 잃는다).
- `internal/template/agentemit/` 의 `developer_instructions` 항등 전달을 바꾸지 않는다. 생성기가 변환 지점으로 **쓸 수 있다**는 것은 §A.3 이 기록한 사실이지 이 SPEC 의 처방이 아니다.

### Out of Scope — 런처와 형제 카드 표면

- `moai codex` 런처 동작은 카드 t391 소관이다. 이 SPEC 은 `internal/cli/codex_launcher.go` 를 읽기만 하고 쓰지 않는다(REQ-CSN-014).
- 코덱스용 대체 환경변수 신설도 하지 않는다 — §A.7 이 그것을 불필요하게 만들었다.

### Out of Scope — dev-only 로컬 스킬

- 로컬 `.claude/skills/` 의 `hns-*` 10종은 배포 대상이 아니므로 이 SPEC 의 어떤 수치에도 포함되지 않고, 편집 대상도 아니다.

---

## §E. 성공 기준

### E.1 Tier 산정 근거

권장안 채택 시 편집 대상:

| 항목 | 파일 수 |
|---|---|
| `AGENTS.md` (루트 + 템플릿 사본) | 2 |
| `${CLAUDE_SKILL_DIR}` 보유 스킬 트리 파일 | 9 |
| `${CLAUDE_SKILL_DIR}` 보유 규칙 트리 파일 (§A.8) | 2 |
| `internal/template/agentemit/agents-codex.yaml` (M2 어휘 보충 — §B.D7) | 1 |
| 회귀 가드 테스트 (신규) | 1 |
| **합계** | **15** |

Tier M 범위(5–15 파일)의 **상한 한 자리**에 든다. [v0.3.1] 합계 14 → 15 는 run-phase 리드 판정(11번째 클래스 보충)으로 합류한 행이다 — 이로 Tier M 상한에 정확히 닿았으므로 **추가 범위 확장은 Tier 재판정 대상이다.** 카드 등록값 `M~L` 에서 **M 으로 확정**한다. Tier L 로 넘겼을 항목(최대 25파일 산문 재작성)은 §D 에서 배제됐고, 배제 사유는 예산이 아니라 §B.D2 다.

### E.2 완료 상태

- `internal/template/templates/.claude/skills/**` 에 `CLAUDE_SKILL_DIR` 토큰 0건.
- 규칙 트리에 남은 `CLAUDE_SKILL_DIR` 줄이 **열거된 사실 기술 집합과 정확히 일치**하며, 규범 문장은 하나도 남아 있지 않다. 완료 상태가 배포된 지침으로 §B.D5 를 반박하지 않는다.
- `AGENTS.md` 두 사본이 동일한 결속표를 담고, 각각 바이트 상한 이하이며, 금지 토큰 0건.
- 회귀 가드가 존재하고, **붉게 뜨는 것이 관측된 상태로** 존재한다.
- 코덱스 거동 관측 1건이 증거 경로에 남아 있다.
- cwd 전제의 지지/미지지 경로가 기록돼 있고, 미지지 팔은 부채로 명시돼 있다.

---

## §F. 교차 참조

- `.moai/reports/t196/premise-remeasure.md` — 본 SPEC §A 의 명령·축자 출력 전문
- `SPEC-CODEX-SKILLS-CANONICAL-001` — 미러의 출처 SPEC. §A.1 의 34종 인벤토리, §A.2 의 `//go:embed` 링크 무음 소실, §D 의 복사 미러 고착 잔존 결함
- `SPEC-CODEX-DUAL-AGENTS-001` — 에이전트 이중 발행의 출처 SPEC
- `internal/template/skill_mirror.go` — 미러 생산자 (`mirrorLinkTarget`, `WithSkillMirror`)
- `internal/template/agentemit/agents-codex.yaml` — `tool_classes` 중립 어휘 (REQ-CSN-002 의 재사용 대상)
- `internal/config/token_budget_guard.go` — `CodexContractByteCeiling` (`:41`), `AlwaysLoadedTokenBudget` (`:32`), `estimateTokens` len/4 (`:105-107`), 고정 표면 슬롯의 AGENTS.md (`:195-200`) — §B.D7 의 예산 근거
- `.moai/reports/t196/req-csn-003-budget-remeasure.md` — §B.D7·AC-CSN-009 셀렉터 정정의 명령·축자 출력 전문 (v0.3.1 배치)
- `internal/template/agentemit/manifest.go:121-124` — 매핑 값↔처분 행 일관성 강제 (§B.D7 의 11번째 클래스 편집 형태)
- `internal/template/renderer.go` — `claudeCodePassthroughTokens` (`:41`), 중괄호 재구성 (`:110-113`)
- `internal/template/internal_content_leak_test.go` — 누출 클래스 표. 전 트리 클래스는 접두 한정(`:170-176`), 이 SPEC 의 토큰을 잡는 클래스는 `skillBodyScoped: true` (§B.D1 정정의 근거)
- `internal/cli/codex_launcher.go:242-250` — 실행 cwd 를 프로젝트 루트로 잡는 분기와 그 강등 분기 (§A.7 의 근거)
- `internal/template/templates/.claude/rules/moai/development/skill-authoring.md:219,226,301` · `.../workflow/worktree-integration.md:386` — §A.8 의 규칙 트리 4줄
- `.moai/reports/t196/plan-audit-iter1.md` · `.moai/reports/t196/_addendum.md` — iter-1 판정 전문
- 카드 t391 — `moai codex` 런처 동작 (표면 겹침, REQ-CSN-014 로 회피)
