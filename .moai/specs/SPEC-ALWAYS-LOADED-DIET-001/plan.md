# Implementation Plan — SPEC-ALWAYS-LOADED-DIET-001

> 순서 원칙: **되돌리기 어려운 결정이 먼저**. §B 결정 사항은 아직 확정되지 않았거나 사람의 판단이 필요한 순으로, §D 마일스톤은 기계적 작업이 뒤로 가도록 배치했다. 리뷰 시간은 §B에 쓰는 것이 맞다.

## §A Context

- 대상 리포: `/Users/goos/MoAI/moai-adk-go`, 브랜치 `main`
- SPEC 산출물: `.moai/specs/SPEC-ALWAYS-LOADED-DIET-001/{spec,plan,acceptance,progress}.md` (Tier M = 3 산출물 + progress)
- 근거: spec.md §1(실측 baseline), §2(재발 증거), §5(미검증 항목). **이 plan은 수치를 다시 유도하지 않고 인용한다.**
- 편집 대상 파일 4개 + 미러 4개 + M3 신설 1개(+미러) = 최대 10개 파일

주요 선례 2개를 작업 전에 통독한다.

- `.claude/rules/moai/workflow/goal-directive.md` (6,531 B) — 스텁 쪽 선례
- `.claude/rules/moai/workflow/goal-directive-detail.md` (17,334 B) — 컴패니언 쪽 선례

선례가 확립한 분할선: **스텁**은 정의 + 메커니즘 + 위반이 안전 실패가 되는 HARD 불변식 + hard-precondition 목록 + 트리거를 한 줄로 압축한 형태 + cross-reference 를 갖는다. **컴패니언**은 표, 워크드 예시, 복붙 템플릿, 근거, 통합 열거를 가져간다.

## §B 결정 사항 — 변경 가능성 높은 순

### D1 (최상위 결정) — M3 재발 통제 메커니즘 선택

후보 3개를 spec.md §2의 분해 수치에 대고 평가한다. 그 수치가 이 결정의 유일한 판별자다: **증가분의 45.3%는 기존 파일의 자체 성장이고 54.7%만 신규 유입**이다.

| 후보 | 커버하는 성장 모드 | 비용 | 판정 |
|---|---|---|---|
| A. 작성 규칙 — 새 always-loaded 규칙은 자기 bytes와 비용 정당화를 진술해야 함 | 신규 유입만 (54.7%) | 문서만, Go 무변경 | **원안은 부족** |
| B. `AlwaysLoadedTokenBudget` 하향 | 양쪽 (100%) | Go 상수 1줄 | 이 SPEC 밖 — 별도 백로그 카드 (D4-1) |
| C. 가드가 pass/fail 대신 현재 여유를 보고 | 양쪽 (100%) | Go 테스트 메시지 | 미채택 — 실패 시점에만 보임 (D4-2) |

**선택: A를 확장한 형태.** 작성 규칙을 두되, 신규 생성뿐 아니라 **기존 always-loaded 파일의 증가**에도 진술 의무를 건다. 근거 셋:

1. A 원안 그대로면 측정된 성장의 절반을 놓친다 — 위 표의 판정이 그 이유다.
2. 새 파일을 `.claude/rules/moai/development/rule-authoring.md` 로 두고 **가드 슬롯 전체를 키잉**하면, always-loaded 표면을 편집하려는 사람은 정의상 그 경로를 건드리고 있으므로 규칙이 정확히 그 순간 붙는다. 채택 글롭은 아래 세 경로다.

   ```
   paths: "**/.claude/rules/**,**/CLAUDE.md,**/.claude/output-styles/**,**/MEMORY.md"
   ```

   이 글롭의 도달 범위는 **295,044 bytes = 100%** 다. 종전 안이던 `**/.claude/rules/**` 단독은 규칙 파일 210,753 bytes 에만 닿아 **71.4%** 에 그친다 — 나머지 `CLAUDE.md`(19,040) + `.claude/output-styles/moai/moai.md`(65,251) = 84,291 bytes 가 **28.6%** 이고, `moai.md` 는 단일 파일로는 표면 최대 기여자다(규칙 1위 `agent-common-protocol.md` 39,455 의 1.65배). 규칙 파일만 키잉하는 안은 통제가 선언한 적용 범위(REQ-ALD-013: 신규 생성 + 기존 증가 양쪽)를 실제 도달 범위가 따라가지 못하는 상태였고, 그것이 이번에 교정된 지점이다. 확장해도 이 파일 자신은 `paths:` 를 갖고 있어 always-loaded 비용은 여전히 0 이다.

   세 번째 글롭은 **슬롯보다 넓다.** `.claude/output-styles/moai/` 에는 `moai.md`(65,251) 외에 `moai-easy.md`(32,342)·`moai-learn.md`(31,395)도 있지만, 가드가 세는 것은 `moai.md` 하나뿐이다 — 위 295,044 에 나머지 둘은 들어 있지 않다. 넓은 쪽으로 틀린 것이라 해는 없다(슬롯이 아닌 output-style 을 편집할 때도 규칙이 붙을 뿐이고, 그 파일은 `paths:` 스코프라 always-loaded 비용이 0 이다). 좁은 글롭(`**/.claude/output-styles/moai/moai.md`)으로 정확히 맞출 수도 있으나, output-style 이 늘어날 때 슬롯 편입을 놓치는 쪽이 더 비싼 실수라 넓은 형태를 택했다.

   부수 관찰(미검증) — `coding-standards.md:25` 는 `CLAUDE.md` 40,000자 한도를 "세션 시작 시 전량 로드되는 project-local instruction 파일"로 확장해 적고 있고, 그 기준에 `moai.md`(65,251)는 25,251 초과다. 다만 그 줄은 스스로를 "CI-**enforceable** heuristic"이라 부를 뿐이며 기계적 강제 여부는 확인하지 않았다. 이 관찰은 확장의 **동기**일 뿐 작동 중인 통제가 아니다(spec.md §5).
3. B/C는 Go 변경을 수반해 리스크를 갖고, 이 SPEC의 순감소 판정을 흐린다.

B는 기각이 아니라 **별도 백로그 카드로 분리한다** — 하향할 값은 M1/M2가 착지해야 알 수 있고, M4가 기록하는 여유가 그 카드의 입력이다. C는 이 SPEC 안에서 채택하지 않는다. 확정 내용과 근거는 아래 D4.

### D2 — `kanban-dispatch-detail.md` 의 `paths:` 키

실측한 컴패니언 4개의 키잉 형태:

| 컴패니언 | `paths:` | 형태 |
|---|---|---|
| `agent-common-protocol-reference.md` | `**/agent-common-protocol.md` | self-keyed |
| `askuser-protocol-reference.md` | `**/askuser-protocol.md` | self-keyed |
| `session-handoff-examples.md` | `**/session-handoff.md` | self-keyed |
| `goal-directive-detail.md` | `**/.moai/state/goal/**,**/.claude/skills/moai/workflows/goal.md,**/goal-directive.md,**/goal-directive-detail.md` | **domain-keyed** |

self-keyed 3개는 부모 규칙 파일을 편집할 때만 붙는다 — 즉 실질적으로 문서 포인터다. domain-keyed 는 실제 그 도메인 작업 중에도 붙는다. lead 세션이 카드를 옮기는 동안 board/card-class/dispatch 구간이 필요하므로 **domain-keyed 를 채택**한다(REQ-ALD-003).

채택 키 3개 — 모두 이번 세션에 존재를 확인했다:

```
paths: "**/kanban-dispatch*.md,**/.claude/agents/moai/manager-kanban.md,**/.claude/skills/moai/workflows/todo.md"
```

- `.claude/agents/moai/manager-kanban.md` — 존재(19,984 B), 미러 존재(19,989 B)
- `.claude/skills/moai/workflows/todo.md` — 존재(5,099 B), 미러 존재(5,099 B)

첫 키의 글롭이 스텁과 컴패니언을 함께 잡는 것은 의도다(`goal-directive-detail.md` 가 자기 자신을 키에 넣은 것과 같은 형태).

### D3 — G7 의 겉보기 모순 해소 방식

`agent-common-protocol.md` § Per-Spawn Model Injection 은 "스폰마다 모델을 명시하라"고 하고, `cache-aware-execution.md` 지시 5는 "세션 모델을 상속하라"고 한다. 읽는 사람에게 모순으로 보인다.

해소: **두 축이 다르다.** 전자는 서브에이전트가 프로파일이 정한 모델로 도는지를 다루고(에이전트 정의 대부분이 `model: inherit` 이므로 미지정 스폰은 부모 모델로 조용히 떨어진다), 후자는 메인 세션이 쌓아온 캐시가 스폰 단위 오버라이드로 쪼개지는 비용을 다룬다. G7 지시문은 이 구분을 한 문장으로 명시하고 양쪽 SSOT를 cross-reference 한다. 어느 쪽도 개정하지 않는다.

### D4 — 확정된 결정 (사용자 판단, 재논의 금지)

이 SPEC에 미해결 항목은 없다. 종전에 이 자리에 있던 미해결 3건은 아래와 같이 확정됐다.

**D4-1 — post-diet 예산 ratchet: 이 SPEC 범위 밖.** `AlwaysLoadedTokenBudget` 은 전 구간에서 75,000 을 유지한다. M4는 착지 후 여유를 **측정해 기록**하고, 그 기록값은 상수를 낮출지 판단하는 **별도 백로그 카드**의 입력이 된다. 근거: 종료 조건이 "완료 후 여유 > 완료 전 여유"인 SPEC 안에서 그 비교의 기준이 되는 예산 상수까지 함께 움직이면 AC가 무엇을 재는지가 흐려진다. baseline 은 75,000 상수 기준 1,239 토큰으로 고정한다. (D1 후보 B)

**D4-2 — M3 재발 통제는 문서 전용, Go 변경 없음.** M3의 산출물은 `.claude/rules/moai/development/rule-authoring.md` 와 그 템플릿 미러가 전부다. 가드 출력은 지금 형태(초과분이 실패 시에만 보임)를 그대로 둔다. 가드가 여유를 함께 보고하도록 바꾸는 안(D1 후보 C)은 **채택하지 않는다**; 원한다면 별도 카드다. 근거: `internal/config` 를 건드리지 않으므로 순감소 판정이 문서 변경만을 재고, 새 CI 리스크가 들어오지 않는다.

**D4-3 — 증가 진술 임계값은 1,000 bytes, 단일 편집 기준.** `rule-authoring.md` 의 진술 의무는 **한 번의 편집이 기존 always-loaded 규칙 파일을 1,000 bytes 넘게 늘릴 때** 발화한다. 그 아래면 발화하지 않는다. 누적 델타를 보는 2차 트리거는 두지 않는다. 기각한 두 값: 500 bytes(문단 하나에 발화하고, 너무 잦은 의무는 형식으로 전락한다), 2,000 bytes(800 bytes 급 추가가 반복돼도 걸리지 않는다). 교정 기준 — 오타 ~100 B와 한 줄 추가 ~200 B는 통과, 문단 ~800 B도 통과, 새 HARD 절 ~1,200 B는 발화, 새 `##` 절 ~2,500 B는 발화. spec.md §2의 측정된 성장(기존 5개 파일 합계 +71,347 bytes)에 대면 1,000 bytes 임계는 약 70회 발화에 해당한다.

## §C 근거 재현 명령

가드 공식 재현(이번 세션에서 실행해 baseline 과 정확히 일치함을 확인):

```bash
tot=0
for f in $(grep -rL '^paths:' .claude/rules/moai --include='*.md'); do tot=$((tot+$(wc -c < "$f"))); done
for f in CLAUDE.md .claude/output-styles/moai/moai.md; do tot=$((tot+$(wc -c < "$f"))); done
[ -f MEMORY.md ] && tot=$((tot+$(head -200 MEMORY.md | head -c 25600 | wc -c)))
echo "bytes=$tot tokens=$((tot/4)) headroom=$((75000 - tot/4))"
# baseline 관측: bytes=295044 tokens=73761 headroom=1239
```

`grep -rL '^paths:'` 는 가드의 frontmatter-내부 판정보다 느슨한 근사다. 현 트리에서는 파일 14개 / 210,753 bytes 로 가드와 동일한 결과를 낸다. **권위 있는 판정은 언제나 `go test ./internal/config/ -run TestAlwaysLoaded`** 이며, 위 스니펫은 그 앞뒤 델타를 보기 위한 보조 계측이다.

구간 바이트 재현(M1 분류의 근거):

```bash
f=.claude/rules/moai/workflow/kanban-dispatch.md
stay=0; for r in 1,6 7,12 60,60 93,123 153,175 176,210 211,218 219,230; do stay=$((stay+$(sed -n "${r}p" "$f" | wc -c))); done
move=0; for r in 13,31 32,43 44,59 61,65 66,92 124,138 139,152; do move=$((move+$(sed -n "${r}p" "$f" | wc -c))); done
echo "stay=$stay move=$move sum=$((stay+move)) file=$(wc -c < "$f")"
# 관측(2026-08-16 재실행): stay=12672 move=8331 sum=21003 file=21003
```

**구간 범위가 더 이상 깨끗한 절 경계와 일치하지 않는다** — REQ-ALD-001 이 지정한 Class B 문단(60행)이 LEAD-ONLY 절 한복판에서 STAY 로 넘어가기 때문이다. 처리 방식은 **스니펫을 다시 짜지 않고 범위만 재유도**하는 쪽을 택했다: STAY 에 `60,60` 을 추가하고, MOVE 의 `44,65` 를 `44,59` 와 `61,65` 로 쪼갰다. 12개였던 범위가 13개가 됐을 뿐 형태는 그대로다.

그 이유는 이 스니펫이 재는 것이 절 구조가 아니라 **분할선의 바이트 귀속**이기 때문이다. 범위 목록은 여전히 1~230행을 빈틈·중복 없이 덮고, 그래서 `sum` 이 파일 총계와 정확히 일치하는 성질이 보존된다 — 이 성질이야말로 분류가 파일 전체를 실제로 덮는지를 판정하는 유일한 기계적 근거다. 절 단위 표현으로 바꾸면 읽기는 쉬워지지만 그 성질을 잃는다.

종전 대비 델타: STAY +495, MOVE −495. 495 bytes 는 60행의 실측 크기다(`sed -n '60p' "$f" | wc -c`). 감사 보고서의 "약 450 B" 는 근사치였고, 아래 투영·코너 케이스는 전부 실측 495 를 쓴다.

## §D 마일스톤

우선순위 라벨만 쓴다(시간 추정 금지). 순서는 M1 → M2 → M3.

### M1 (Priority High) — `kanban-dispatch.md` 분리

1. `kanban-dispatch-detail.md` 생성. D2의 `paths:` + `description:` frontmatter, 그리고 자기 쪽 소유 경계 선언(`goal-directive-detail.md` 도입부 형태).
2. LEAD-ONLY 6구간을 원본에서 **잘라내어** 컴패니언에 붙인다. 재작성이 아니라 이동이다(REQ-ALD-005).
3. 단, `Card classes` 안의 Class B 문단(원본 60행)은 컴패니언이 아니라 **STAY 구간 `## Completion is read, never trusted` 로** 옮긴다(REQ-ALD-001). 줄을 쪼개지 말고 통째로 옮긴다 — 쪼개면 AC-ALD-006 이 거짓 실패한다. 옮긴 자리에서 앞뒤 빈 줄이 겹치면 하나로 정리하고(1 byte 수준, 오버헤드 허용 범위), `Card classes` 쪽에는 대체 문장을 **새로 쓰지 않는다**(REQ-ALD-005 — 컴패니언·스텁 어느 쪽도 원본에 없던 내용을 얻지 않는다).
4. 스텁에 포인터 줄 삽입. 옮긴 6구간을 이름으로 열거하고 과업 형태 트리거를 진술한다("카드를 컬럼 간 이동시키거나 카드 클래스를 판정하거나 리뷰 렌즈를 고를 때 로드").
5. 스텁 푸터의 `Classification:` 줄 아래 버전 줄에 분리 사실 기록.
6. 계측: 위 §C 첫 스니펫 재실행 → 델타 확인.

투영: 스텁 12,672 + 컴패니언 8,331 = 21,003 (정확히 원본과 같음). 절감 8,331 B (39.7%)에서 포인터 주석 + 컴패니언 frontmatter 오버헤드 600~900 B 가 되돌아온다 — **추정치이며 측정치가 아니다**(spec.md §5). 종전 투영(12,177 / 8,826 / 42.0%)은 Class B 문단 재배치 이전 값이며 더 이상 유효하지 않다.

순감소 코너 케이스(오버헤드 300~900 B × M2 증가 1,000~2,000 B, 재배치 반영):

| 스텁 오버헤드 | M2 증가 | bytes | tokens | headroom |
|---|---|---|---|---|
| +900 | +2,000 | 289,613 | 72,403 | **2,597** |
| +900 | +1,000 | 288,613 | 72,153 | 2,847 |
| +300 | +2,000 | 289,013 | 72,253 | 2,747 |
| +300 | +1,000 | 288,013 | 72,003 | 2,997 |

최악(오버헤드 900 + M2 상한 2,000)에서도 여유는 2,597 토큰으로 baseline 1,239 대비 1,358 토큰의 슬랙을 남긴다. 재배치로 절감이 495 B 줄어 최악값이 2,721 → 2,597 로 내려갔지만 순감소 판정(AC-ALD-001)은 모든 코너에서 성립한다.

리스크: `session-handoff-examples.md` 는 40,891 B 로 부모(23,251)보다 크다. 분리가 축소를 보장하지 않는다는 증거이며, 2단계에서 "정리하면서 살 붙이기"를 하면 그대로 재현된다. 이동 후 스텁+컴패니언 합계가 21,003 ± 오버헤드 범위를 벗어나면 살이 붙은 것이다.

### M2 (Priority High) — G3~G7 편입

1. `cache-aware-execution.md` 에 지시 6~10을 추가. 각각 한 문단, `[ZONE:Evolvable] [HARD]` 접두, 현행 지시 1~5와 같은 **형태**.

   "같은 형태"는 크기가 아니라 모양이다(번호 매김 + 접두 + 한 문단). 크기는 오히려 더 작아야 한다 — 실측하면 현행 1~5는 508/614/490/501/355 B, 합계 2,468 B(평균 494)이고, **그 크기로 다섯 개를 더 쓰면 AC-ALD-011 상한 2,000 B 를 넘겨 FAIL 한다.** 현행 지시가 494 B 인 이유는 근거를 문단 안에 품고 있기 때문이고, 새 지시는 `REQ-ALD-011` 에 따라 근거·수치·예시를 컴패니언으로 보내므로 그보다 작아야 정상이다. 역산한 예산은 **문단당 약 200~400 B**.
2. `cache-aware-execution-reference.md` 생성. `paths: "**/cache-aware-execution.md"` 최소치에 더해 실제 캐시 관련 작업 경로를 함께 키잉할지는 D2와 같은 판단을 적용한다 — self-keyed 로 둔다(이 파일은 근거 보관용이며 도메인 작업 중 필요하지 않다).
3. 인용 수치(read 0.1x / write 최대 2x / output ≈ input 5x / TTL 구독 1h·API 5m)는 컴패니언에 두고, **인용 출처값**임을 그 자리에서 명시(REQ-ALD-012).
4. G7 문단에 D3 해소 문장 + 양쪽 SSOT cross-reference.
5. 계측: `cache-aware-execution.md` 증가분이 +1,000~2,000 B 안에 드는지 확인(AC-ALD-011). 하한을 1,500 에서 낮춘 이유는 그쪽에 있다 — 다섯 규율을 현행 지시 1~5 형태로 간결하게 쓰면 1,200~1,400 B 로도 다 들어가며, 그때 종전 하한은 **잘 쓴 구현을 거짓 실패시킨다**. 진술 누락 탐지는 AC-ALD-010 의 6개 패턴이 단독으로 맡는다. 상한 2,000 은 순감소 보호에 필요하므로 유지한다.

### M3 (Priority Medium) — 재발 통제

1. `.claude/rules/moai/development/rule-authoring.md` 생성. `paths:` 는 가드 슬롯 전체를 덮는 확장 글롭이다(D1 근거 2).

   ```
   paths: "**/.claude/rules/**,**/CLAUDE.md,**/.claude/output-styles/**,**/MEMORY.md"
   ```

2. 내용: (a) 새 always-loaded 파일(= top-level `paths:` 없는 규칙, 그리고 `CLAUDE.md`·output-style 처럼 슬롯 자체인 파일) 생성 시 bytes 와 비용 정당화를 본문에 진술할 것, (b) **한 번의 편집**이 기존 always-loaded 파일을 **1,000 bytes 넘게** 늘릴 때 같은 진술을 할 것, (c) 진술은 **그 파일을 한 번도 필요로 하지 않는 세션이 지불하는 비용**을 다뤄야 함(REQ-ALD-014), (d) `paths:` 스코프로 옮길 수 있는지 먼저 물을 것. (a)~(d) 는 규칙 파일뿐 아니라 가드가 세는 슬롯 4개 전부에 건다 — 규칙 파일만 다루면 표면의 28.6%가 통제 밖에 남는다.
3. (c)의 근거로 현행 실태를 인용: `session-handoff.md` 가 비용을 이름으로 부르는 가장 가까운 사례이고, `native-idiom-and-register.md` 의 "영어 세션에 부담 zero" 는 동작에 대해 참·컨텍스트 바이트에 대해 거짓이다.
4. 임계값은 **1,000 bytes, 단일 편집 기준**으로 못박아 쓴다(D4-3). 규칙 본문이 이 값을 읽는 사람에게 납득시키도록 교정 기준을 함께 적는다 — 오타 ~100 B·한 줄 추가 ~200 B·문단 ~800 B는 통과, 새 HARD 절 ~1,200 B·새 `##` 절 ~2,500 B는 발화. 누적 델타를 보는 2차 트리거는 두지 않는다.
5. M3는 **문서 전용**이다(D4-2). `internal/config` 의 가드 코드와 출력 형식은 손대지 않으며, 이 마일스톤이 만드는 Go 변경은 없다.

### M4 (Priority Medium) — 미러 + 빌드 + 최종 계측

1. 생성·수정 파일 전부를 `internal/template/templates/` 아래 대응 경로에 미러.
2. `make build`.
3. 미러본 중립성 확인 — SPEC ID / REQ 토큰 / 내부 날짜 / 커밋 SHA 부재.
4. `go test ./internal/config/ -run TestAlwaysLoaded` + §C 스니펫 재실행 → **여유 > 1,239** 확인(REQ-ALD-015). 이때 관측된 여유 값을 progress.md 에 기록한다 — 예산 상수 하향을 판단할 별도 백로그 카드의 입력이 되는 값이다(D4-1). 이 SPEC 안에서 `AlwaysLoadedTokenBudget` 은 바꾸지 않는다.

## §E 리스크

| # | 리스크 | 완화 |
|---|---|---|
| R1 | 컴패니언이 적치장이 됨 | M1 이동 후 합계 바이트를 원본과 대조(AC-ALD-009) |
| R2 | 스텁에서 안전 관련 HARD 절이 사라짐 | 이동 전후 HARD 절 grep 대조(AC-ALD-004) |
| R3 | M2 증가가 M1 절감을 상쇄 | 순감소를 단일 종료 조건으로 고정(AC-ALD-001) |
| R4 | 미러 누락으로 배포본 깨짐 | M4에서 파일별 대조 + `make build` |
| R5 | `paths:` 부착이 예상과 다르게 동작 | 관측 불가(spec.md §5). 부착 실패 시 최악의 결과는 "컴패니언이 안 붙음"이며 스텁이 모든 구속 절을 갖고 있으므로 안전 실패다 — 이것이 REQ-ALD-001의 존재 이유 |
| R6 | 병렬 세션이 같은 체크아웃에서 규칙 파일을 편집 | 워크트리 격리 + 명시 pathspec 스테이징(§F) |

## §F 제약

- **PR 필수**: 이 리포는 `main` 이 `enforce_admins: true` 로 보호돼 있어 Route A(main 직행)가 불가하다. Tier M이라도 feature 브랜치 + PR 로 간다(`.claude/rules/local/repo-local-pr-policy.md`).
- **브랜치 상태 변경 금지**: primary 체크아웃에서 `git switch` / `git branch` / `git stash` 금지. 격리는 `moai cc -w <name>` 로 진입하고 `moai worktree done` 으로 폐기한다.
- **스테이징은 명시 pathspec**: `git add -A` / `git add .` / `git commit -a` 금지.
- **Template-First**: `.claude/` 하위 신규 파일은 템플릿 소스에 먼저 넣고 `make build` 후 로컬 동기화. 단 미러본은 중립화가 필요한 sanitized-pair 일 수 있으므로 **`cp` 로 통째 복사하지 않는다**.
- **로컬 전체 스위트 금지**: `go test ./...` 대신 영향 패키지(`./internal/config/`)만 돌리고 전 패키지 판정은 CI에 맡긴다.
- **plan-phase 커밋 제목은 `feat(` 접두**: `docs(` 를 쓰면 generic docs 로 오분류돼 `StatusGitConsistency` 경고가 상존한다.

## §G Anti-patterns

- **AP-1 — 분리를 축소로 착각.** 파일을 둘로 나누는 것 자체는 always-loaded 표면을 줄이지 않는다. 컴패니언에 `paths:` 가 붙어야 줄어든다. `session-handoff-examples.md` 가 반례다.
- **AP-2 — "정리하는 김에" 문장 다듬기.** 이동은 이동이다. 살이 붙으면 R1이 현실화되고 AC-ALD-009가 실패한다.
- **AP-3 — 공백 압축으로 바이트 줄이기.** 금지. 감축은 내용 이동으로만 얻는다.
- **AP-4 — self-keyed 컴패니언을 domain-keyed 라 부르기.** 키 문자열이 부모 규칙 파일만 가리키면 실제 kanban 작업 중에는 붙지 않는다. D2 표가 이 구분의 근거다.
- **AP-5 — 인용 수치를 실측치처럼 서술.** M2의 배수·TTL은 원문 인용이다(REQ-ALD-012).
- **AP-6 — M3를 always-loaded 파일에 넣기.** 재발 통제가 스스로 재발 원인이 된다. `paths:` 스코프가 필수다.

## §H Cross-references

- `internal/config/token_budget_guard.go` / `token_budget_guard_test.go` — 가드 구현과 트립와이어
- `.claude/rules/moai/workflow/goal-directive.md` + `goal-directive-detail.md` — 스텁/컴패니언 분리 선례
- `.claude/rules/moai/workflow/kanban-dispatch.md` — M1 대상
- `.claude/rules/moai/workflow/cache-aware-execution.md` — M2 대상
- `.claude/rules/moai/core/agent-common-protocol.md` § Per-Spawn Model Injection — D3의 한쪽 축
- `.claude/rules/moai/workflow/context-window-management.md` § Reduction Ladder — G6가 붙는 자리
- `SPEC-V3R6-RULES-PATH-SCOPE-001` — 이전 다이어트(재발의 기준점)
- `.claude/rules/local/repo-local-pr-policy.md` — 전 Tier PR 필수
