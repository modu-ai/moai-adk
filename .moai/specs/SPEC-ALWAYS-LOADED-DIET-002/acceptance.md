# SPEC-ALWAYS-LOADED-DIET-002 — 인수 조건

모든 AC 는 Given-When-Then 이며 이진 판정된다. 명령은 워크트리 루트에서 실행한다.

---

## §D. AC 매트릭스

| AC | 대상 REQ | 판정 | 심각도 |
|---|---|---|---|
| AC-ALD2-001 | REQ-ALD2-001 | 기계적 | MUST-PASS |
| AC-ALD2-002 | REQ-ALD2-002, -003, -011, -012 | 기계적 | MUST-PASS |
| AC-ALD2-003 | REQ-ALD2-007, -008 | 기계적 | MUST-PASS |
| AC-ALD2-004 | REQ-ALD2-010 | 문서 + 검토 | MUST-PASS |
| AC-ALD2-005 | REQ-ALD2-009 | 기계적 | MUST-PASS |
| AC-ALD2-006 | REQ-ALD2-004 | 기계적 + 검토 | MUST-PASS |
| AC-ALD2-007 | REQ-ALD2-005 | 검토 | SHOULD-PASS |
| AC-ALD2-008 | REQ-ALD2-006 | 기계적 | MUST-PASS |
| AC-ALD2-009 | REQ-ALD2-014, -015, -016 | 기계적 | MUST-PASS |

---

### AC-ALD2-001 — 18파일 합계가 한도 아래

**Given** 구현이 완료된 워크트리에서,
**When** 경고가 세는 18개 경로에 `wc -m` 을 실행하고 `total` 줄을 읽을 때,
**Then** 합계는 **150,000 미만**이다.

```bash
wc -m CLAUDE.md AGENTS.md \
  .moai/config/sections/user.yaml .moai/config/sections/language.yaml \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/askuser-protocol.md \
  .claude/rules/moai/core/moai-constitution.md \
  .claude/rules/moai/core/moai-mcp-tools.md \
  .claude/rules/moai/core/native-idiom-and-register.md \
  .claude/rules/moai/core/verification-claim-integrity.md \
  .claude/rules/moai/workflow/cache-aware-execution.md \
  .claude/rules/moai/workflow/context-window-management.md \
  .claude/rules/moai/workflow/cross-session-messaging.md \
  .claude/rules/moai/workflow/goal-directive.md \
  .claude/rules/moai/workflow/kanban-dispatch.md \
  .claude/rules/moai/workflow/main-checkout-branch-guard.md \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/rules/moai/workflow/skill-routing.md | tail -1
```

기준선(2026-09-25, 이 트리): `246943 total`. 목표: `< 150000`.

---

### AC-ALD2-002 — 구속 조항 무손실 (동결 해시)

**Given** 작업 전 기준선 해시 `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3` (170줄),
**When** 구현 후 같은 16개 마크다운 파일에서 `[HARD]` / `MUST` / `shall ` 을 담은 줄을 뽑아 앞뒤 공백을 정규화하고 정렬한 sha256 을 구할 때,
**Then** 해시는 기준선과 **바이트 동일**하다.

```bash
grep -rhE '\[HARD\]|MUST|shall ' CLAUDE.md AGENTS.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/askuser-protocol.md \
  .claude/rules/moai/core/moai-constitution.md \
  .claude/rules/moai/core/moai-mcp-tools.md \
  .claude/rules/moai/core/native-idiom-and-register.md \
  .claude/rules/moai/core/verification-claim-integrity.md \
  .claude/rules/moai/workflow/cache-aware-execution.md \
  .claude/rules/moai/workflow/context-window-management.md \
  .claude/rules/moai/workflow/cross-session-messaging.md \
  .claude/rules/moai/workflow/goal-directive.md \
  .claude/rules/moai/workflow/kanban-dispatch.md \
  .claude/rules/moai/workflow/main-checkout-branch-guard.md \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/rules/moai/workflow/skill-routing.md \
  | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' | sort | shasum -a 256
```

**해시 불일치는 어느 방향이든 FAIL 이다** — 조항이 companion 으로 옮겨졌든(REQ-ALD2-002 위반), 재작성됐든(REQ-ALD2-003 위반), 새로 추가됐든 마찬가지다. 구속 조항 줄은 이 카드에서 **축자 동결**된다.

부수 검사(불일치 시 원인 지목용): 줄 수가 `170` 인지, 그리고 `grep -c` 로 센 파일별 개수가 기준선과 같은지.

#### 이 해시가 구속하는 것 — 순서가 아니라 다중집합

파이프라인이 `grep … | sed … | sort | shasum` 이므로 `sort` 가 순서를 정규화한다. 따라서 해시가 구속하는 것은 **정규화된 줄의 다중집합(multiset)** 이며, 줄의 순서도 출신 파일도 구속하지 않는다. 읽는 사람이 파이프라인에서 직접 도출하도록 두지 않고 여기에 명시한다 — 재배치 중 조항 순서를 바꿔도 되는지가 이 한 줄에 달려 있기 때문이다.

**깨지지 않는다(허용)**
- 한 파일 안에서 구속 조항의 순서를 바꾸는 것.
- 구속 조항 줄을 측정 대상 16파일 **중 두 파일 사이**에서 옮기는 것 — 해시는 출신 파일을 보지 않는다.

**깨진다(금지)**
- 구속 조항 줄을 16파일 **밖으로** 내보내는 것(companion 재배치가 여기 해당한다 — REQ-ALD2-002).
- 삭제하는 것(REQ-ALD2-013 포함 — 같은 의무가 path-scoped 룰에 있다는 이유로 지우는 경우).
- 문구를 고치는 것, 앞뒤 공백 밖의 어떤 변형이든(REQ-ALD2-003).

**중복 줄은 보존되고 그 개수가 해시의 일부다** — `sort` 는 `sort -u` 가 아니다. 동일한 구속 조항 줄이 둘 있을 때 하나를 지우면 해시가 깨진다. 이것이 직관에 반하는 쪽이라 적어 둔다. 다만 현재 기준선에는 중복이 **없다**(170줄, `sort -u` 후에도 170 — 실측). 따라서 이 조항은 현재 상태를 서술하는 것이 아니라, 작업 중 중복이 **생기거나 사라지는 경우**에 대해 구속한다.

**AC-ALD2-003 에 대한 귀결**: 해시가 출신 파일에 눈이 멀어 있으므로, "어느 조항이 어느 파일에 있는가"는 AC-ALD2-002 가 **검증하지 않는다.** 그 축은 AC-ALD2-004 의 재배치 표가 떠안는다 — 표가 비면 파일 단위 귀속은 아무 데서도 판정되지 않는다.

---

### AC-ALD2-003 — template ↔ live 미러 동등

**Given** 이 SPEC 이 수정한 룰 파일 목록이,
**When** 각 파일에 대해 라이브 사본과 템플릿 사본을 `diff` 할 때,
**Then** 모든 쌍이 차이 0 이다. 특히 `skill-routing.md` 의 템플릿 사본은 라이브와 동일한 `paths:` frontmatter 를 갖는다.

**판정 대상에서 `CLAUDE.md` 는 제외된다.** 두 사본에는 이 카드보다 앞선 의미 분기가 있고(라이브 루트가 낡은 쪽 — `spec.md §D`), 그 해소는 별도 카드다. 제외는 **범위 결정이지 두 사본이 일치한다는 주장이 아니다** — `CLAUDE.md` 의 두 사본은 각각 독립적으로 감축되고, 각각 AC-ALD2-002 의 구속 조항 동결을 지킨다.

```bash
for f in $(git diff --name-only HEAD~1 -- '.claude/rules/moai/**' | sed 's|^\.claude/|internal/template/templates/.claude/|'); do :; done
# 실제 판정: 수정한 룰마다
diff .claude/rules/moai/<path> internal/template/templates/.claude/rules/moai/<path> && echo OK
head -4 internal/template/templates/.claude/rules/moai/workflow/skill-routing.md   # paths: 존재 확인
```

기준선: 현재 `skill-routing.md` 템플릿 사본에는 frontmatter 가 **없다**(실측 확인됨). 목표: 라이브와 동일.

---

### AC-ALD2-004 — 재배치 표가 존재하고 완전

**Given** 구현이 절을 옮겼을 때,
**When** `design.md §4` 의 재배치 표를 읽을 때,
**Then** 옮겨진 모든 절이 정확히 한 행씩 있고, 각 행이 여섯 칸을 모두 채우고 있다: 출발 stub · 절 이름 · 목적지 companion · 그 companion 의 `paths:` 값 · 그 키가 self-keyed 가 아니라 domain-keyed 인 이유 · **범위 의존(scope dependency)**. 빈 칸이 하나라도 있으면 FAIL.

**범위 의존 칸**이 담는 것(§D.3 의 위험을 이 표가 떠안는다): 그 절을 참조하는 stub 내 구속 조항을 **열거**하고, 조항마다 **두 질문에 모두** 답한다.

| # | 질문 | 잡는 방향 |
|---|---|---|
| Q1 | 이동 후에도 이 조항이 자기 범위 문맥에 **도달하는가**? | 축소 — 참조 대상이 떠남 |
| Q2 | 이 조항의 범위가 이동 전보다 **넓어지지 않았는가**? | 확대 — 한정어가 떠남 |

**Q1 만 묻는 칸은 한정어가 절과 함께 떠난 조항을 통과시킨다.** 그 조항은 여전히 "도달"하지만 더 넓은 것을 구속하게 되며, 결손이 아니라 증가 방향이라 어떤 손실 탐지기도 발화하지 않는다(§D.3 방향 2).

참조하는 구속 조항이 없으면 `없음`이라 적되, 빈칸으로 두지 않는다 — `없음`은 확인했다는 뜻이고 빈칸은 보지 않았다는 뜻이다.

판정: 열거된 구속 조항 각각에 대해, 검토자가 stub 을 읽었을 때 그 조항이 무엇을 구속하는지가 이동 전과 **같아야** 한다 — 좁아져도 넓어져도 FAIL 이다. 해시(AC-ALD2-002)가 통과하더라도 그렇다.

**AC-ALD2-002 가 눈이 먼 축은 둘이다.** 하나는 출신 파일(위 문단), 다른 하나는 **범위 폭**이며 후자는 **양방향 모두**에 눈이 멀어 있다 — `sort` 된 줄 목록은 그 줄을 둘러싼 한정어가 있든 없든 같기 때문이다. 따라서 AC-ALD2-002 의 PASS 가 확립하는 것은 "구속 줄의 다중집합이 동일하다"뿐이고, "각 조항이 같은 것을 구속한다"는 **이 표만이** 확립한다.

**이 표가 단독으로 떠안는 축 — 파일 단위 귀속**: AC-ALD2-002 의 해시는 `sort` 를 거쳐 **출신 파일에 눈이 멀어 있다**(AC-ALD2-002 § 이 해시가 구속하는 것). 따라서 "어느 구속 조항이 16파일 중 어느 파일에 있는가"는 그 검사가 **전혀 검증하지 않으며**, 오직 이 재배치 표만이 그 축을 담는다. 표가 비거나 행이 누락되면 파일 단위 귀속은 이 SPEC 어디에서도 판정되지 않는다 — 해시는 통과한 채로.

---

### AC-ALD2-005 — 앵커 무결성

**Given** 재배치 표의 각 절 제목에 대해,
**When** 저장소 전체에서 그 제목 문자열을 역방향 grep 하고 각 히트가 가리키는 파일을 확인할 때,
**Then** 모든 인용이 해소된다 — 옮겨진 절을 가리키는 끊긴 앵커가 0 이다.

```bash
grep -rn "<옮긴 절 제목>" --include='*.md' . | grep -v '^\./\.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/'
```

각 히트에 대해 판정: 인용이 (a) stub 의 포인터 줄을 경유하거나, (b) companion 을 직접 가리키거나 둘 중 하나여야 한다. 어느 쪽도 아닌 히트는 끊긴 앵커다.

---

### AC-ALD2-006 — companion `paths:` 가 domain-keyed

**Given** 이 SPEC 이 수정한 각 companion 파일에 대해,
**When** 그 `paths:` 값을 읽을 때,
**Then** 값이 부모 stub 경로 **하나만** 담고 있지 않다 — 그 companion 이 다루는 작업이 건드리는 도메인 경로를 함께 담는다.

```bash
grep -m1 '^paths:' .claude/rules/moai/<companion>.md
```

FAIL 예(self-keyed): `paths: ".claude/rules/moai/workflow/kanban-dispatch.md"` — 부모 룰을 **편집할 때만** 로드되므로 그 룰이 지배하는 작업 세션에는 도달하지 않는다.
PASS 선례: `goal-directive-detail.md` (유일한 domain-keyed 선례).

---

### AC-ALD2-007 — stub 3요소

**Given** 분리가 수행된 각 stub 에 대해,
**When** stub 본문을 읽을 때,
**Then** 세 요소가 모두 있다: (a) 옮긴 절을 **이름으로** 호명하고 작업 모양의 로드 트리거를 적은 포인터 줄, (b) companion 이 선언하는 자기 소유 경계, (c) 분리를 기록하는 stub 푸터 버전 줄.

"자세한 내용은 companion 참조" 같은 절 이름 없는 포인터는 (a) 미충족으로 FAIL.

---

### AC-ALD2-008 — companion 이 내용을 획득하지 않음

**Given** 각 companion 의 작업 전 크기와 옮긴 절들의 문자 합계가,
**When** 작업 후 companion 크기를 `wc -m` 으로 잴 때,
**Then** `작업후 − 작업전 ≤ 옮긴절합계 + 500` 이다. 500자 여유는 companion 측 절 제목·소유 경계 문장 같은 접합부만 허용한다. 초과분은 원본에 없던 내용의 획득이므로 FAIL(REQ-ALD2-006).

---

### AC-ALD2-009 — 파일당 40,000자 (합계 축과 별개)

**Given** 두 트리(`internal/template/templates/.claude/rules/`, `.claude/rules/moai/`)의 모든 룰 파일에 대해,
**When** 각 파일의 문자 수를 재고 40,000 이상인 것을 열거할 때,
**Then** (a) 이 카드가 내용을 쓴 파일은 그 목록에 **하나도 없고**, (b) 목록의 길이가 트리마다 **4를 넘지 않는다**.

```bash
# (a)+(b) 한 번에 — 두 트리 각각
find internal/template/templates/.claude/rules -name '*.md' -exec wc -m {} + \
  | sort -rn | awk '$1>=40000 && $2!="total"'
find .claude/rules/moai -name '*.md' -exec wc -m {} + \
  | sort -rn | awk '$1>=40000 && $2!="total"'
```

기준선(2026-09-25, 두 트리 동일 — 실측):

```
61435  workflow/worktree-integration.md
41616  workflow/session-handoff-examples.md
41036  workflow/kanban-dispatch-detail.md
40799  workflow/spec-workflow.md
```

판정 규칙 셋:

- **(a) 이 카드가 쓴 파일이 목록에 있으면 FAIL** (REQ-ALD2-014). 신규 companion 이 스스로 넘긴 경우도 포함한다.
- **(b) 목록 길이가 4를 넘으면 FAIL** (REQ-ALD2-016). 래칫이며, 어떤 경로로 늘었든 무관하다.
- **(c) 위 4개가 그대로 남아 있는 것은 PASS 다** (REQ-ALD2-015 — `spec.md §D`). 이 카드는 넷을 수리하지 않고 **악화만 금지**한다. 넷이 줄어드는 것은 당연히 PASS 이며, 이 카드가 그것을 목표로 삼지 않을 뿐이다.

**이 AC 는 AC-ALD2-001 과 독립이다.** 합계가 150,000 아래여도 파일 하나가 40,000을 넘으면 FAIL 이고, 그 역도 같다. 두 축을 한 기준에 접지 않는 이유는 계량기가 서로 다르기 때문이다 — 합계는 Claude Code 런타임이, 파일당은 `moai hook instructions-loaded` 가 잰다(`spec.md §A` 표).

**이 AC 가 통과해도 확립되지 않는 것**: 40,000 미만이라는 사실은 그 파일이 **적절한 크기**라는 뜻이 아니다. 한도는 상한이지 목표가 아니며, 39,207자로 끝나는 파일(`design.md §2.2` 의 `agent-common-protocol-reference.md`)은 여유가 793자뿐이라 다음 카드가 조금만 더해도 다시 넘긴다.

---

## §D.1 판정 게이트

- **MUST-PASS 8개**(AC-001~006, 008, 009) 중 하나라도 FAIL 이면 SPEC 은 FAIL 이다. 점수로 상쇄되지 않는다.
- AC-ALD2-007 은 SHOULD-PASS — FAIL 시 PASS-WITH-DEBT 로 기록하고 후속 카드를 낸다.

## §D.2 추적성

| REQ | AC |
|---|---|
| REQ-ALD2-001 | AC-ALD2-001 |
| REQ-ALD2-002 | AC-ALD2-002 |
| REQ-ALD2-003 | AC-ALD2-002 |
| REQ-ALD2-004 | AC-ALD2-006 |
| REQ-ALD2-005 | AC-ALD2-007 |
| REQ-ALD2-006 | AC-ALD2-008 |
| REQ-ALD2-007 | AC-ALD2-003 |
| REQ-ALD2-008 | AC-ALD2-003 |
| REQ-ALD2-009 | AC-ALD2-005 |
| REQ-ALD2-010 | AC-ALD2-004 |
| REQ-ALD2-011 | AC-ALD2-002 (간접 — `AGENTS.md` 에서 절이 사라지면 해시가 깨진다) |
| REQ-ALD2-012 | AC-ALD2-002 (간접 — 구속 줄은 동결; 비구속 의미 보존은 검토 판정) |
| REQ-ALD2-013 | AC-ALD2-002 (의무 사본을 지우면 해시가 깨진다) + AC-ALD2-004 범위 의존 칸 |
| REQ-ALD2-014 | AC-ALD2-009 (a) |
| REQ-ALD2-015 | AC-ALD2-009 (c) |
| REQ-ALD2-016 | AC-ALD2-009 (b) |

## §D.3 잔여 위험 — 범위 이탈 (scope detachment)

**통과하는 기계 검사에 가려지는 위험이므로 별도 항목으로 세운다.**

AC-ALD2-002 의 동결 해시는 구속 *줄*을 지킨다. 그 줄을 **범위 짓는 문맥**은 지키지 않는다. 절 제목·한정하는 표·참조되는 열거가 companion 으로 옮겨지고 구속 줄만 stub 에 남으면, 줄은 바이트 동일로 살아남아 **해시가 통과하는데 그 줄이 구속하는 대상은 조용히 바뀐다.**

**이 위험은 방향이 둘이고, 둘째는 손실 탐지기에 아예 걸리지 않는다.**

**방향 1 — 축소 (참조 대상이 떠난다).** 구속 조항이 의존하던 정의·표·열거가 companion 으로 떠나고 조항만 남는다. 예: "every row it marks PASS MUST correspond to an actually-observed command output" 는 제자리에 남고, **어떤 row 가 존재하는지를 정의하던 표**가 떠난다. 조항은 글자 그대로 멀쩡한데 `row` 가 무엇인지가 always-loaded 표면에서 사라진다.

**방향 2 — 확대 (한정어가 떠난다).** 범위를 **좁히던** 비구속 문장이 지워지면, 뒤따르는 구속 조항이 이전보다 **넓게** 읽힌다. 예: `이 절은 primary 체크아웃에서만 적용된다` 가 사라지면, 바로 아래의 `[HARD]` 금지 조항이 워크트리에까지 걸리는 것으로 읽힌다. 조항 줄은 손대지 않았고 해시도 통과한다.

방향 2 가 더 어렵다. **결손이 아니라 증가 방향의 실패**여서, "무엇이 사라졌는가"를 보는 어떤 탐지기도 발화하지 않는다 — 사라진 것은 의무가 아니라 의무의 **울타리**이고, 울타리가 사라지면 의무는 줄지 않고 늘어난다.

> **이 방향 2 는 이 SPEC 의 작성자가 자기 공개의 구멍으로 찾은 것이다** — 감사가 지적해서가 아니라, plan-audit 대기 중에 스스로 예상 변이를 적어 보다가 §D.3 이 축소 방향만 예시로 들고 있음을 발견했다. 어떻게 발견됐는지를 적어 두는 이유는, 구성상 완전해 **보이는** 공개보다 발견 경위를 밝힌 공개가 나중에 읽는 사람에게 더 믿을 만하기 때문이다.

이 위험이 §D.4 의 REQ-ALD2-012 미검증 항목보다 **더 위험하다.** 후자는 검사가 없다고 스스로 밝히지만, 이것은 **통과한 기계 검사가 안전하다는 인상을 준다.**

완화는 두 겹이며 둘 다 사람/에이전트 검토에 걸린다:

1. **AC-ALD2-004 의 범위 의존 칸** — 옮기는 절마다, 그 절을 참조하는 stub 내 구속 조항을 열거하고 **양방향으로** 단언한다: 이동 후에도 도달하는가(방향 1), 그리고 범위가 이동 전보다 넓어지지 않았는가(방향 2). 위험을 진술문에서 **행 단위 산출물**로 끌어내리는 장치다.
2. **`plan.md §C` M1 의 조건 1 단서** — 구속 조항이 없지만 구속 조항이 의존하는 표·제목인 절은 **옮기지 않는다.** 방향 2 에 대해서는 같은 단서가 이렇게 읽힌다: **범위를 좁히는 문장은 비구속이어도 지우지 않는다.**

이 축의 판정 주체는 sync-audit 검토다. AC-ALD2-002 의 PASS 가 이 축의 검증을 대신하지 않는다.

## §D.4 간접 검증의 한계 — 숨기지 않고 적는다

REQ-ALD2-012(제자리 압축 시 의미 보존)에는 **기계 검사가 없다.** 구속 조항 해시는 구속 *줄*만 지키고, 비구속 산문에서 조건·예외·수치가 조용히 사라지는 것은 잡지 못한다. 같은 이유로 `design.md §7` 이 "문맥 손실은 해시로 잡히지 않는다"를 위험으로 올려 두었다. 이 항목의 판정은 sync-audit 의 사람/에이전트 검토이며, AC 가 통과했다는 사실이 이 축의 검증을 대신하지 않는다.

## §D.5 완료 정의

MUST-PASS 7개 통과 + 재배치 표 완비 + 두 사본 동등 + `go test ./internal/config/... ./internal/template/...` 통과(템플릿 사본 변경이 임베드·중립성 검사를 건드리므로).

🗿 MoAI
