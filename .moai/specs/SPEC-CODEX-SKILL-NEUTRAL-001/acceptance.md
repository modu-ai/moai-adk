# SPEC-CODEX-SKILL-NEUTRAL-001 — 수용 기준

## §D. AC 표

측정 트리는 이 SPEC 이 도는 워크트리, 측정 시점 HEAD 는 판정 당시 `git rev-parse HEAD`. 모든 AC 는 그 HEAD 에서 재측정한 값으로 판정하며, spec.md §A 의 값을 그대로 옮겨 쓰지 않는다.

### AC-CSN-001 — 코덱스 거동이 실제로 관측됐다 (MUST)

**Given** 축 A 의 우선순위가 "코덱스는 없는 도구명을 조용히 대체한다"는 추론 위에 서 있고,
**When** run-phase 가 축 A 의 본문을 편집하기 전에 코덱스 세션에 그 도구명을 담은 지시를 물리면,
**Then** `.moai/reports/t196/codex-behavior.log` 가 존재하고 그 안에 세션의 축자 출력이 있으며, 그 출력이 추론과 일치하는지 어긋나는지가 한 문장으로 판정되어 있다.

판정: 파일 존재 + 축자 출력 포함 + 판정 문장 존재. 어긋난 관측이 기록되고 blocker 가 올라간 경우도 이 AC 는 **통과**다 — 이 AC 가 요구하는 것은 특정 결과가 아니라 관측 그 자체다.

### AC-CSN-002 — 중립 어휘를 새로 만들지 않았다 (MUST)

**Given** `internal/template/agentemit/agents-codex.yaml` 의 `tool_classes` 가 중립 클래스 이름을 이미 정의하고 있고,
**When** `AGENTS.md` 의 결속표에서 능력 이름을 읽으면,
**Then** 표에 등장하는 모든 중립 능력 이름이 `tool_classes` 의 값 집합에 속한다.

```bash
# 결속표가 쓰는 이름 집합 ⊆ tool_classes 의 값 집합
```

판정: 두 집합의 차집합이 공집합. 차집합이 비지 않으면 FAIL 이며, 그 원소를 명명한다.

### AC-CSN-003 — 결속표가 3열을 모두 담는다 (MUST)

**Given** 결속표가 `AGENTS.md` 에 실려 있고,
**When** 각 행을 읽으면,
**Then** 모든 행이 (a) 중립 이름, (b) Claude 구현, (c) 능력 부재 시 행동 세 칸을 비어 있지 않게 채우고 있다.

판정: 표 파싱 후 빈 칸 0개. (c) 열이 비어 있는 행이 하나라도 있으면 FAIL — 그 열이 이 SPEC 의 전부이기 때문이다.

### AC-CSN-004 — 질문 채널 부재 시 행동이 blocker 반환으로 적혀 있다 (MUST)

**Given** 결속표에 질문 채널 행이 있고,
**When** 그 행의 "능력 부재 시 행동" 칸을 읽으면,
**Then** 그 칸이 blocker 보고 반환을 지시하며, 사용자에게 직접 묻는 것을 지시하지 않는다.

판정: 해당 칸의 축자 인용. 사람이 읽고 판정한다 — 문면 판정이며 기계적 문자열 일치가 아니다. 이 사실을 감춘 정규식 검사를 만들지 않는다.

### AC-CSN-005 — 두 사본이 동일하고 둘 다 상한 이하 (MUST)

**Given** `AGENTS.md` 가 루트와 `internal/template/templates/` 두 곳에 있고,
**When** 결속표 추가 후 두 파일을 비교하고 바이트 가드를 돌리면,
**Then** `cmp AGENTS.md internal/template/templates/AGENTS.md` 가 차이 없이 종료하고, `go test ./internal/config/ -run TestCodexContractByteCeiling` 이 통과하며, 그 `-v` 출력의 바이트·여유 수치가 기록된다.

```bash
cmp AGENTS.md internal/template/templates/AGENTS.md
go test ./internal/config/ -run TestCodexContractByteCeiling -v
```

판정: `cmp` 종료코드 0 + 테스트 통과 + `-v` 출력의 두 사본 바이트 수치 축자 기록. **착수 전 값과의 차분**을 함께 적는다(plan.md §C).

### AC-CSN-006 — 실행 인자 자리의 변수가 사라졌다 (MUST)

**Given** 셸·노드 호출의 인자로 `${CLAUDE_SKILL_DIR}` 을 쓰던 자리가 있었고,
**When** 치환 후 그 형태를 다시 세면,
**Then** 일치하는 줄이 0이다.

```bash
grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/.claude/skills | grep -E ':(bash|node) ' | wc -l
```

판정: 출력이 `0`. **착수 전 같은 명령의 값을 함께 기록한다** — 0 만 적으면 원래 0 이었는지 닫아서 0 인지 구별되지 않는다.

[HARD] **이 AC 의 사후 단언은 AC-CSN-008 에 포섭된다.** 전수 0(AC-CSN-008)이 성립하면 이 파이프라인의 첫 단이 비어 결과가 무조건 0 이 되므로, 이 AC 는 독립적으로 실패할 수 없다. **이 AC 의 실제 내용은 pre/post 값 쌍**이며 사후 초록은 독립 증거가 아니다 — 나중 독자가 그렇게 읽지 않도록 여기 적어 둔다.

### AC-CSN-007 — 산문 경로 자리도 함께 사라졌다 (MUST)

**Given** 산문 경로 지시로 `${CLAUDE_SKILL_DIR}` 을 쓰던 자리가 실행 자리보다 많았고,
**When** 치환 후 전체를 다시 세면,
**Then** 실행 자리와 산문 자리가 **함께** 0이다.

이 AC 가 "조용한 깨짐"을 기계적으로 판정할 수 있는 이유를 분명히 적어 둔다: **여기서 고치는 것이 의미 판단이 아니라 토큰 제거이기 때문이다.** 산문 자리의 실패는 조용하지만, 그 자리의 *존재*는 조용하지 않다 — `grep` 이 센다. 따라서 이 AC 는 공허하지 않다.

판정: AC-CSN-008 의 전수 계수와 함께 판정한다.

### AC-CSN-008 — 잔여 토큰 전수 0 (MUST)

**Given** 치환이 끝났고,
**When** 스킬 트리 전체를 훑으면,
**Then** `CLAUDE_SKILL_DIR` 토큰이 하나도 남아 있지 않다.

```bash
grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/.claude/skills | wc -l
```

판정: 출력이 `0`, 그리고 착수 전 값(측정 시점 기준 46)이 함께 기록되어 있다.

### AC-CSN-009 — 회귀 가드가 붉게 뜨는 것이 관측됐다 (MUST)

**Given** 회귀 가드 테스트가 추가됐고,
**When** 감시 대상에 `CLAUDE_SKILL_DIR` 토큰을 일부러 한 줄 심고 가드를 돌리면,
**Then** 가드가 실패하고, 되돌린 뒤 다시 돌리면 통과한다. 심기 전 census(현재 0)와 심은 뒤 census(1)가 함께 기록된다.

판정: RED 관측 출력 + GREEN 복귀 출력 + 양쪽 census. **RED 를 보지 않은 가드는 이 AC 를 통과하지 못한다** — 초록은 감시가 작동한다는 증거가 아니라 위반이 없다는 증거일 뿐이고, 두 상태는 아무것도 안 보는 가드에서 구별되지 않는다.

[HARD] **판정 기록은 실행한 셀렉터와 그 매치 수를 명시한다** [v0.3.1]. 접두 셀렉터 `-run TestSkillDirToken` 은 **0매치**다 — `testing: warning: no tests to run` 을 출력하고 PASS·exit 0 으로 끝난다(이 트리 실측; 축자 출력은 `.moai/reports/t196/req-csn-003-budget-remeasure.md` §5). 실제 테스트는 `TestSkillTreeHasNoClaudeSkillDirToken` 1건이며(`internal/template/skill_dir_token_guard_test.go:41`), `-run TestSkillTreeHasNoClaudeSkillDirToken` 은 **1매치**다. **0매치 셀렉터의 초록은 판정이 아니다** — 훑은 집합이 빈 통과는 아무것도 집지 못한 통과와 구별되지 않는다. 판정 기록에 셀렉터 문자열과 매치 수가 빠지면 이 AC 의 판정은 이루어지지 않은 것으로 본다.

### AC-CSN-010 — 범위 밖 표면이 실제로 안 건드려졌다 (MUST)

**Given** 이 SPEC 이 런처와 미러 생산자와 생성기를 범위 밖에 뒀고,
**When** 계획 단계 기준 트리 SHA 부터 현재까지의 변경 집합을 읽으면,
**Then** `internal/cli/codex_launcher.go` · `internal/template/skill_mirror.go` · `internal/template/agentemit/**` 에 변경이 없다.

```bash
git diff --stat 297a21ea73b24e6605280625e576555e4316263e..HEAD -- \
  internal/cli/codex_launcher.go internal/template/skill_mirror.go internal/template/agentemit/
```

[HARD] **기준은 브랜치 이름이 아니라 트리 SHA 로 못박는다.** 범위 밖 표면 단언은 불변식 주장이므로, 움직이는 ref 를 기준으로 삼으면 기준이 단언 아래에서 이동해 만든 적 없는 변경을 보고하거나 만든 변경을 감춘다. 기준값은 계획 단계 HEAD `297a21ea73b24e6605280625e576555e4316263e` 다. **run-phase 가 다른 HEAD 에서 착수하면 착수 시점 HEAD 를 재측정해 이 자리에 적고, 바꿨다는 사실을 progress.md 에 남긴다** — 조용히 교체하지 않는다.

판정: 출력이 비어 있다. `--stat` 을 쓰는 이유는 접두사 필터가 마크다운 목록 줄을 버리는 형태의 오판을 피하기 위해서다.

### AC-CSN-011 — 규칙 트리에 규범 문장이 남지 않았다 (MUST)

**Given** 배포되는 규칙 트리가 `${CLAUDE_SKILL_DIR}` 을 상대 경로보다 선호하라고 가르치고 있었고(착수 시점 4줄 / 2파일),
**When** 수정 후 규칙 트리의 토큰 보유 줄을 전수 열거하면,
**Then** 남은 줄이 **사실 기술 집합과 정확히 일치**하고, 규범 문장(`skill-authoring.md:226` · `:301` 형태)은 하나도 없다.

```bash
grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/.claude/rules
```

판정: 출력의 **각 줄을 열거해 기록**하고, 각 줄이 사실 기술인지 규범 문장인지 한 단어로 분류한다. 규범 문장이 하나라도 남으면 FAIL.

기대 잔존 집합(착수 시점 기준 예상):

| 줄 | 기대 분류 | 기대 처분 |
|---|---|---|
| `skill-authoring.md:219` | 사실 기술 | **잔존** — 능력 표의 한 행이며 참이다(§B.D6) |
| `skill-authoring.md:226` | 규범 문장 | 제거 또는 채택 설계에 맞게 수정 |
| `skill-authoring.md:301` | 규범 문장 | 동일 |
| `worktree-integration.md:386` | 예시 | **수정** — plan M3 가 지시한다 |

[HARD] **`:386` 의 기대 분류를 못박는다.** 종전 판본은 "잔존하거나 사라질 수 있으며 어느 쪽이든 분류 결과를 적는다"로 열어 두었는데, plan M3 는 그 줄의 수정을 **지시**한다 — 그러면 `:386` 을 손대지 않고 `예시` 로 분류하기만 해도 이 AC 는 통과하면서 plan 지시는 건너뛰어진다. 수정 후에도 토큰이 남는 형태(예: 다른 표현으로 바꾸며 변수명을 인용)라면 **분류는 `예시(수정됨)`** 이어야 하고, 원문 그대로 남아 있으면 FAIL 이다.

[HARD] **개수 판정이 아니라 집합 판정이다.** "4 → 1" 이라는 수는 근거가 되지 않는다 — 규범 문장 하나를 남기고 사실 기술 둘을 지워도 같은 수가 나오기 때문이다. 남은 **줄의 내용**이 판정 대상이다.

### AC-CSN-012 — 이 SPEC 이 편집하는 무가드 파일 5개에 금지 토큰이 새로 들어오지 않았다 (MUST)

**Given** 템플릿 중립성 CI 가드가 이 SPEC 자신의 토큰 형태(`SPEC-CODEX-…` / `REQ-CSN-…`)를 잡는 클래스를 `.claude/skills/` 아래로만 한정하고(§B.D1 정정), **이 SPEC 이 편집하는 파일 중 스킬 트리 밖에 있는 것이 5개**이며 — `AGENTS.md` 2 사본(M2) + 규칙 트리 2파일(M3, §B.D6 로 편입) + `internal/template/agentemit/agents-codex.yaml`(M2 어휘 보충 — spec.md §B.D7, 리드 판정 2026-09-01) — 근거를 인용하고 싶어지는 자리가 바로 그 다섯이고,
**When** M2·M3 편집 후 그 5개를 검사하면,
**Then** SPEC ID · REQ/AC 토큰 · 커밋 해시 · ISO 날짜가 **착수 시점보다 늘지 않았고**, 규칙 트리에 남은 매치는 열거된 기존 항목과 정확히 일치한다.

```bash
grep -cE 'SPEC-[A-Z0-9]+-[A-Z0-9-]*[0-9]{3}|\b(REQ|AC)-[A-Z0-9]+-[0-9]{3}\b|\b[0-9a-f]{7,40}\b|20[0-9]{2}-[0-9]{2}-[0-9]{2}' \
  AGENTS.md \
  internal/template/templates/AGENTS.md \
  internal/template/templates/.claude/rules/moai/development/skill-authoring.md \
  internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md \
  internal/template/agentemit/agents-codex.yaml
```

**착수 시점 측정값** (이 트리, HEAD `297a21ea73b24e6605280625e576555e4316263e`):

| 파일 | 사전값 | 성격 |
|---|---|---|
| `AGENTS.md` | 0 | M2 가 새로 들여오는 것만 잡는다 |
| `internal/template/templates/AGENTS.md` | 0 | 동일 |
| `…/development/skill-authoring.md` | **2** | 기존 — `:45`·`:89` 의 프론트매터 예시 ISO 날짜 |
| `…/workflow/worktree-integration.md` | 0 | — |
| `internal/template/agentemit/agents-codex.yaml` | **구현 착수 시점 측정** | [v0.3.1] M2 어휘 보충으로 합류 — 이 파일의 사전값은 지금 재지 않고 **구현 착수 시점에 잰다**(그 전에 yaml 이 바뀔 수 있다) |

[HARD] **단순 `0` 단언이 아니라 pre/post 쌍이다.** `skill-authoring.md` 의 사전값이 0 이 아니기 때문이다 — 그 2건은 프론트매터 예시 안의 날짜 리터럴이고, 정당한 문서 내용이며 **지울 대상이 아니다**(CI 가드도 S1 클래스의 날짜 carve-out 으로 통과시킨다). 판정은 AC-CSN-006/008/011 이 이미 쓰는 형태를 따른다: **사후값 ≤ 사전값**, 그리고 `skill-authoring.md` 에 남은 매치가 `:45`·`:89` 두 줄과 일치할 것.

[HARD] **왜 이 5개인가 — 가드가 닿지 않는 정확한 집합이다.** 이 SPEC 의 다른 편집 대상은 전부 `templates/.claude/skills/**` 아래에 있고 거기서는 `skillBodyScoped` 클래스가 발화한다. 스킬 트리 밖 편집 대상은 이 5개뿐이므로, 이 AC 의 파일 목록이 REQ-CSN-013 의 실질 노출면을 덮는다. **M3 이 규칙 트리를 범위에 들이면서 생긴 구멍이며**, 그것을 닫지 않으면 "규범 문장을 왜 뒤집었는지"를 SPEC ID 로 적는 가장 자연스러운 수정이 아무 데도 안 걸린다.

[v0.3.1] **다섯 번째 파일이 왜 생겼는가 — 범위 확장이 무가드 파일을 계속 만들어낸다.** 이 카드에서 두 번 그랬다: iter-2 D11 이 규칙 트리 2파일을, run-phase 리드 판정이 `agents-codex.yaml` 을 이 목록에 합류시켰다. **이 SPEC 의 범위가 다시 넓어지면 가드 커버리지 재확인이 편집보다 먼저다** — 새 편집 대상이 이 목록과 `skillBodyScoped` 클래스(`.claude/skills/` 한정) 어느 쪽에도 닿지 않는지 확인 없이 편집하지 않는다.

[HARD] **정규식이 공허하지 않다는 것을 양성 대조로 확인한 뒤 판정한다.** 대조는 **판정과 같은 명령**(위 `grep -cE`)을 이 SPEC 의 `spec.md` 에 걸어 실행하며, 결과가 **0 이 아니면** 통과다. 대조가 0 이면 정규식이 죽은 것이므로 4개 파일의 값은 아무 의미가 없다.

  구체적 수치를 적지 않는다 — 대조 대상이 **이 SPEC 자신의 문서**라서 편집할 때마다 값이 움직이고, 통과 조건은 "0 이 아닐 것"이므로 수치가 필요하지 않다. 대조는 판정 시점에 실행해 읽는다(§D.4).
  **대조에 판정과 다른 명령을 쓰지 않는다** — 특히 `-c`(줄 수)와 `-co`(출현 수)는 값이 다르므로, 다른 플래그로 잰 값을 대조로 제시하면 실행자가 재현되지 않는 숫자를 만난다.

[HARD] **실패 시 매치 텍스트를 기록한 뒤 판정한다.** 위 명령은 `-c` 라 개수만 내놓는데, 16진수 팔(`\b[0-9a-f]{7,40}\b`)은 평범한 영어 단어에도 걸린다 — 실측: `defaced`, `feedbed`. 커밋 해시가 없는데도 붉어질 수 있으므로, **사후값이 사전값을 넘으면 `-c` 를 빼고 다시 돌려 매치 문자열을 기록**한다:

```bash
grep -nE '<위와 동일한 정규식>' <초과한 파일>
```

거짓 적색(평범한 단어)과 진짜 적색(실제 토큰)을 한 단계에서 갈라야 하며, 매치 텍스트 없이 개수만으로 FAIL 을 보고해서는 안 된다.

### AC-CSN-013 — cwd 전제의 지지·미지지 경로가 기록됐다 (MUST)

**Given** 루트 기준 경로가 기대는 전제가 "읽는 프로세스의 cwd 가 프로젝트 루트"라는 것이고,
**When** SPEC §A.7 을 읽으면,
**Then** (a) 그 전제를 지지하는 코드 위치가 인용돼 있고, (b) 그 코드의 **강등 분기**가 함께 인용돼 있으며, (c) `moai codex` 를 거치지 않은 직접 실행 세션이 이 SPEC 의 어떤 요구사항으로도 묶이지 않는다는 사실이 부채로 적혀 있다.

판정: 세 항목의 존재를 문면으로 확인한다. **문면 판정이며 기계적 문자열 일치가 아니다** — 그 사실을 감춘 정규식 검사를 만들지 않는다.

(c) 를 요구하는 이유: 이 SPEC 은 cwd 팔을 **닫지 않는다.** 닫히지 않은 것을 닫혔다고 읽히게 두지 않는 것이 이 AC 의 전부다.

---

## §D.1 심각도

전 항목 MUST. nice-to-have 항목은 두지 않았다 — 13개 중 어느 하나가 빠져도 "같은 신뢰"라는 목표가 성립하지 않는다.

## §D.2 요구사항 추적

| AC | 판정하는 REQ |
|---|---|
| AC-CSN-001 | REQ-CSN-001 |
| AC-CSN-002 | REQ-CSN-002 |
| AC-CSN-003 | REQ-CSN-003 |
| AC-CSN-004 | REQ-CSN-004 |
| AC-CSN-005 | REQ-CSN-005 |
| AC-CSN-006 | REQ-CSN-006 |
| AC-CSN-007 | REQ-CSN-007 |
| AC-CSN-008 | REQ-CSN-008 |
| AC-CSN-009 | REQ-CSN-010, REQ-CSN-011 |
| AC-CSN-010 | REQ-CSN-014 |
| AC-CSN-011 | REQ-CSN-015 |
| AC-CSN-012 | REQ-CSN-013 |
| AC-CSN-013 | REQ-CSN-009 |

**iter-1 판정이 잡은 거짓 매핑 하나를 정정했다.** 종전 표는 `AC-CSN-005 → REQ-CSN-005, REQ-CSN-013` 로 적었으나 AC-CSN-005 는 두 사본 `cmp` 동일성과 바이트 상한만 본다 — REQ-CSN-013(템플릿 중립성)을 **위반하면서 통과하는 mutant 가 성립**했다. REQ-CSN-013 은 이제 전용 판정 AC-CSN-012 를 갖는다.

REQ-CSN-012(Template-First)만이 전용 AC 를 갖지 않는다. 사유는 §D.3.

## §D.3 기계적으로 판정할 수 없는 것 — 공허한 검사를 만들지 않는다

세 항목은 이 SPEC 의 판정 계층이 **기계적으로 닫을 수 없다.** 닫을 수 있는 척하는 검사를 만드는 대신 그 사실을 적는다.

> iter-2 변경: 종전 (2)번은 "복사 폴백 모드에서의 경로 해석"이었고 REQ-CSN-009 를 부채로 지목했다. 그 팔은 **애초에 깨질 수 없는 팔**이었으므로(루트 기준 경로는 미러에 닿지 않는다 — spec.md §A.7) 부채가 아니라 **잘못 겨눈 요구사항**이었다. REQ-CSN-009 를 cwd 팔로 재조준하고 AC-CSN-013 을 붙였으며, 그 자리에 진짜 미판정 항목인 (2) 직접 실행 세션의 cwd 를 넣었다.

**(1) 결속표가 코덱스를 실제로 Claude 와 같게 행동하게 하는가.**
이것이 카드의 진짜 목표이고, 어떤 `grep` 도 판정하지 못한다. 판정하려면 같은 지시를 두 하네스에 물리고 산출물을 비교하는 대조 실행이 필요하며, 그 비교의 기준선 자체가 이 SPEC 에 없다. **AC-CSN-002~004 가 판정하는 것은 결속표가 존재하고 형태를 갖췄다는 것뿐이다** — 그것이 효과를 낸다는 주장은 이 SPEC 이 하지 않는다. 효과 판정은 대조 실행 하네스를 갖춘 별도 카드의 몫이다.

**(2) 직접 실행된 코덱스 세션의 작업 디렉터리.**
루트 기준 경로는 읽는 프로세스의 cwd 가 프로젝트 루트라는 전제 위에 선다(spec.md §A.7). `moai codex` 로 띄운 세션은 런처가 그것을 보장하지만 — 그것도 루트 해석이 성공할 때만이고, 실패하면 프로세스 cwd 로 강등된다 — **`moai codex` 를 거치지 않고 직접 띄운 세션은 사용자의 cwd 를 그대로 물려받는다.** 이 SPEC 은 그 팔을 묶지 않는다. AC-CSN-013 이 요구하는 것은 이 사실이 **기록되는 것**이지 닫히는 것이 아니다. 닫으려면 런처 밖에서 cwd 를 강제하는 장치가 필요하고, 런처 표면은 카드 t391 소관이다.

**(3) REQ-CSN-012 — Template-First 준수.**
"로컬 사본을 원본으로 편집하지 않았다"는 것은 변경 집합만 보고는 판정되지 않는다 — 두 경로에 같은 내용이 있으면 어느 쪽이 원본이었는지 파일이 말해주지 않는다. `moai update` 후 로컬이 템플릿과 일치하는지로 간접 확인할 수는 있으나, 그것은 순서가 아니라 최종 상태를 판정한다. 부채로 남긴다.

## §D.4 종료 조건 (Definition of Done)

- AC-CSN-001 ~ AC-CSN-013 전부 통과, 각 판정에 명령과 축자 출력이 붙어 있다.
- 착수 전 값과 착수 후 값이 **쌍으로** 기록되어 있다(AC-CSN-005, 006, 008, 009, 011, 012).
- §D.3 의 세 부채가 progress.md 에 부채로 적혀 있고, 통과로 위장되어 있지 않다.

### [HARD] 자기참조 수치 규율 — run-phase 에 그대로 걸린다

**주어가 이 SPEC 자신의 산출물을 포함하는 수치는, 인용하는 그 시점에 다시 재거나 아예 수치로 인용하지 않는다.**

plan-phase 에서 이 형태로 3건이 썩었다. 기제는 부주의가 아니라 문서 형식이다 — **자기참조 측정값은 그 값을 인용한 문서를 편집하는 행위 자체가 무효화하므로, 쓰는 순간이 곧 깨지는 순간**이다. 그래서 셋 다 lint·MUST-PASS·자체 스윕을 통과했다.

- **주어가 외부인 수치**(템플릿 트리, 런처, 바이트 가드, 에이전트 집합)는 이 규율의 대상이 아니다 — 이 SPEC 을 편집해도 움직이지 않는다. 계획 단계에서 15건 전부 재현됐다.
- **주어가 자기 산출물인 수치**는 (a) 판정 시점에 재측정해 기록하거나, (b) 수치를 적지 않고 조건만 적는다. AC-CSN-012 의 양성 대조가 (b) 를 택한 사례다.
- **출처 사실은 워킹트리가 아니라 ref 에 대고 읽는다** — `git ls-tree <ref>` / `git grep <pattern> <ref> -- <paths>`. 잰 트리가 명령에 박히므로 라벨을 잃을 수 없고, 이 SPEC 자신의 디렉터리는 미추적이라 ref 판독은 자기포함에 **구조적으로 면역**이다.

**run-phase 가 이 규율을 상속한다.** `progress.md` 와 §E.2 증거 절은 **구조상 자기참조**이므로(자기 산출물을 대상으로 세는 값이 계속 생긴다), 이 규율이 없으면 같은 계열이 계속 생산된다. 위반은 판정 실패가 아니라 **기록 무결성 실패**이며, 발견 즉시 재측정해 정정한다.
- `make build` 가 통과하고, 템플릿 중립성 가드가 통과한다.
