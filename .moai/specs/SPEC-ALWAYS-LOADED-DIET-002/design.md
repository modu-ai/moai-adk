# SPEC-ALWAYS-LOADED-DIET-002 — 설계

파일별 분리 계획과 재배치 표. 근거 측정은 `.moai/reports/t1175/discovery.md` 와 이 문서 §1 의 자체 실측.

---

## §1. 파일별 실측 (이 실행, 이 트리)

문단 단위 분류. 문단이 `[HARD]` / `MUST` / `MUST NOT` / `shall ` 중 하나라도 담으면 **구속**, 아니면 **재배치 가능**으로 센다. `discovery.md` §5 의 집계(173,286)를 파일 단위로 푼 것이며, 차이(172,363 vs 173,286)는 discovery 가 룰 14개를 템플릿 트리에서 잰 반면 여기서는 라이브 트리에서 쟀기 때문이다 — 같은 축의 같은 방법이다.

| 파일 | 현재 | 구속 | 재배치 가능 | 가능 비율 |
|---|---:|---:|---:|---:|
| `workflow/kanban-dispatch.md` | 39,077 | 18,945 | 20,134 | 51.5% |
| `core/agent-common-protocol.md` | 28,381 | 7,905 | 20,478 | 72.2% |
| `core/askuser-protocol.md` | 24,466 | 8,742 | 15,726 | 64.3% |
| `workflow/session-handoff.md` | 19,908 | 6,412 | 13,498 | 67.8% |
| `CLAUDE.md` | 19,305 | 3,033 | 16,274 | 84.3% |
| `core/moai-constitution.md` | 17,272 | 7,180 | 10,094 | 58.4% |
| `core/verification-claim-integrity.md` | 16,636 | 8,046 | 8,592 | 51.6% |
| `AGENTS.md` | 16,374 | 2,817 | 13,559 | 82.8% |
| `workflow/cross-session-messaging.md` | 16,330 | 2,497 | 13,835 | 84.7% |
| `workflow/context-window-management.md` | 11,447 | 2,561 | 8,887 | 77.6% |
| `core/moai-mcp-tools.md` | 7,400 | 651 | 6,751 | 91.2% |
| `workflow/cache-aware-execution.md` | 6,907 | 1,861 | 5,048 | 73.1% |
| `workflow/goal-directive.md` | 6,875 | 0 | 6,877 | 100.0% |
| `workflow/main-checkout-branch-guard.md` | 6,814 | 658 | 6,158 | 90.4% |
| `workflow/skill-routing.md` | 5,709 | 2,465 | 3,246 | 56.9% |
| `core/native-idiom-and-register.md` | 3,815 | 842 | 2,975 | 78.0% |
| `language.yaml` + `user.yaml` | 227 | 0 | 227 | — |
| **합계** | **246,943** | **74,615** | **172,363** | **69.8%** |

`goal-directive.md` 의 구속 0 은 그 파일이 이미 stub 화되어 구속 조항을 detail 과 `goal.md` 로 **참조**만 하고 있기 때문이다 — 분리 선례로 삼을 파일이지, 마음껏 잘라도 되는 파일이 아니다.

---

## §2. 목적지 — 여유가 있는 기존 companion 을 쓰고, 없으면 새로 만든다

> **[개정 — 2026-09-25]** 이 절의 원래 결정은 "새 companion 파일을 만들지 않는다"였고, 근거는 "always-loaded 룰 14개가 전부 이미 companion 을 하나씩 갖고 있다"였다. 그 근거는 목적지의 **존재**를 논했을 뿐 **수용량**을 논하지 않았다. 파일당 40,000자 축(§2.1)이 범위에 들어오면서 존재와 수용량이 다른 것임이 드러났고, 원래 결정을 그대로 실행하면 이 SPEC 자신의 인수 조건을 위반하게 된다. 개정한다. 원래 판단이 틀렸다기보다 **전제가 불완전**했다.

**개정된 목적지 선택 규칙**: 절을 옮길 목적지는 **이동 후 40,000자 미만으로 끝나야 한다.**

1. 기존 companion 에 여유가 있으면 그것을 쓴다(기본 — 새 `paths:` 설계와 stub 푸터 작업이 줄어든다).
2. 여유가 없거나 일부만 있으면, 넘치는 분량은 **주제 경계를 따라 새 companion 으로 분리**한다. 새 companion 도 REQ-ALD2-004(domain-keyed `paths:`)·REQ-ALD2-005(stub 3요소)·REQ-ALD2-006(원본에 없던 내용 금지)을 똑같이 진다.
3. **이미 한도를 넘긴 companion 에는 옮기지 않는다.** 옮기는 순간 이 카드가 그 파일의 초과에 기여하게 되고, 그러면 그 파일의 수리가 이 카드의 몫이 된다(§2.2).

한 stub 이 companion 을 둘 갖는 것은 허용된다 — 오히려 각 companion 의 `paths:` 가 더 좁고 정확해진다.

아래 표는 **각 stub 이 이미 가진 companion 의 목록**이지 최종 배정이 아니다. 최종 배정은 수용량을 반영한 **§2.2** 이며, 이 표의 세 행(`kanban-dispatch`·`session-handoff`·`agent-common-protocol`)은 거기서 바뀐다.

**수치의 트리 표기 (D4)**: 이 절과 §2.1 의 모든 자수는 **라이브 트리**(`.claude/rules/moai/`) 실측이다. 파일당 한도 4개 목록은 두 트리에서 동일함을 확인했으나(`spec.md §D`), companion 자수는 두 트리가 갈릴 수 있으므로 어느 쪽을 잰 것인지 밝힌다. §1 의 18파일 실측도 라이브 트리 기준이며, `discovery.md §5` 의 집계(173,286)만 템플릿 트리 기준이라 0.5% 차이가 난다.

| stub (always-loaded) | 기존 companion (path-scoped) | companion 현재 크기 |
|---|---|---:|
| `workflow/kanban-dispatch.md` | `workflow/kanban-dispatch-detail.md` | 41,036 |
| `core/agent-common-protocol.md` | `core/agent-common-protocol-reference.md` | 31,174 |
| `core/askuser-protocol.md` | `core/askuser-protocol-reference.md` | 15,646 |
| `workflow/session-handoff.md` | `workflow/session-handoff-examples.md` | 41,616 |
| `core/moai-constitution.md` | `core/moai-constitution-detail.md` | 7,499 |
| `core/verification-claim-integrity.md` | `core/verification-claim-integrity-detail.md` | 21,283 |
| `workflow/cross-session-messaging.md` | `workflow/cross-session-messaging-detail.md` | 15,160 |
| `workflow/context-window-management.md` | `workflow/context-window-management-detail.md` | 7,486 |
| `core/moai-mcp-tools.md` | `core/moai-mcp-tools-catalogue.md` | 11,424 |
| `workflow/cache-aware-execution.md` | `workflow/cache-aware-execution-reference.md` | 5,816 |
| `workflow/goal-directive.md` | `workflow/goal-directive-detail.md` | 20,283 |
| `workflow/main-checkout-branch-guard.md` | `workflow/main-checkout-branch-guard-detail.md` | 9,622 |
| `workflow/skill-routing.md` | `workflow/skill-routing-detail.md` | 1,351 |
| `core/native-idiom-and-register.md` | `core/native-idiom-and-register-detail.md` | 2,509 |
| `CLAUDE.md` | 없음 — 기존 path-scoped 룰을 가리킨다 | — |
| `AGENTS.md` | **없음, 만들지 않는다** (REQ-ALD2-011) | — |

`CLAUDE.md` 의 목적지가 새 companion 이 아닌 이유: §1-§17 의 비구속 산문 대부분이 이미 path-scoped 룰에 정본으로 존재하는 내용의 **재서술**이다. 따라서 옮기는 것이 아니라 **중복을 지우고 정본을 가리킨다** — M1 의 변형이며 새 파일을 만들지 않는다.

### §2.1 수용량 실측 — 목적지 14개를 40,000자 한도에 대고 재다

파일당 40,000자는 `moai hook instructions-loaded` 가 강제한다. 소스를 직접 읽어 확인한 것(`internal/hook/instructions_loaded.go:86-104`): `checkCharacterBudget` 는 **로드되는 파일마다** `utf8.RuneCount` 를 재고 40,000 초과 시 위반 메시지를 낸다. 즉 **디스크에 앉아 있는 상태가 아니라 로드 시점에 발화하며, path-scoped companion 도 로드되면 똑같이 걸린다.** 이것이 86개 룰 전부가 이 축의 대상인 이유다.

계획된 유입량을 각 목적지에 더한 결과(2026-09-25, 라이브 트리 실측):

| 목적지 companion | 현재 | 유입 | 이후 | 판정 |
|---|---:|---:|---:|---|
| `kanban-dispatch-detail.md` | 41,036 | 16,000 | 57,036 | **초과 17,036** (이미 초과 상태) |
| `session-handoff-examples.md` | 41,616 | 10,000 | 51,616 | **초과 11,616** (이미 초과 상태) |
| `agent-common-protocol-reference.md` | 31,207 | 15,000 | 46,207 | **초과 6,207 — 이 카드가 넘긴다** |
| `askuser-protocol-reference.md` | 15,646 | 11,000 | 26,646 | 여유 13,354 |
| `verification-claim-integrity-detail.md` | 21,347 | 6,000 | 27,347 | 여유 12,653 |
| `cross-session-messaging-detail.md` | 15,221 | 9,000 | 24,221 | 여유 15,779 |
| `goal-directive-detail.md` | 20,283 | 3,000 | 23,283 | 여유 16,717 |
| `moai-mcp-tools-catalogue.md` | 11,424 | 4,000 | 15,424 | 여유 24,576 |
| `moai-constitution-detail.md` | 7,644 | 7,000 | 14,644 | 여유 25,356 |
| `main-checkout-branch-guard-detail.md` | 9,984 | 3,500 | 13,484 | 여유 26,516 |
| `context-window-management-detail.md` | 7,486 | 5,000 | 12,486 | 여유 27,514 |
| `cache-aware-execution-reference.md` | 5,841 | 3,000 | 8,841 | 여유 31,159 |
| `native-idiom-and-register-detail.md` | 2,509 | 1,500 | 4,009 | 여유 35,991 |
| `skill-routing-detail.md` | 1,351 | 1,800 | 3,151 | 여유 36,849 |

**셋이 걸리고, 셋의 성격이 같지 않다.** 앞의 둘은 이 카드가 손대기 전부터 초과였다. 세 번째 `agent-common-protocol-reference.md` 는 **이 카드의 계획이 만드는 신규 초과**이며, 이미 초과한 파일만 훑는 방식으로는 보이지 않는다 — 나머지 11개도 같은 이유로 매번 실측해야 한다.

### §2.2 개정된 목적지 배정

§2 규칙 3("이미 넘긴 companion 에는 옮기지 않는다")을 적용한 결과:

| 출발 stub | 유입량 | 개정된 목적지 | 이후 |
|---|---:|---|---:|
| `kanban-dispatch.md` | 16,000 | **신규** companion (주제 경계 분리) | ~16,000 |
| `session-handoff.md` | 10,000 | **신규** companion | ~10,000 |
| `agent-common-protocol.md` | 15,000 | 기존 reference 에 8,000 (→39,207) + **신규** companion 에 7,000 | 39,207 / ~7,000 |
| 나머지 11개 | 각 표대로 | 기존 companion 유지 | §2.1 표 |

신규 companion 3개가 생긴다. 각각 REQ-ALD2-004·005·006 을 그대로 지며, AC-ALD2-008(증식 금지)에서 "작업 전 크기"는 0 으로 계산한다.

`agent-common-protocol-reference.md` 의 39,207 은 한도 아래지만 여유가 793자뿐이다. run phase 에서 실측이 이보다 나쁘면 분할 비율을 조정한다 — 8,000/7,000 은 고정값이 아니라 **"기존 파일이 40,000 아래로 끝난다"는 제약의 한 해**다.

---

## §3. 파일별 감축 목표와 투영 (v0.3.0 개정 — 재유입 항 반영)

> **[개정 — 2026-09-25, plan-audit iter2 D5]** 종전 투영은 **포인터 줄 재유입 항을 누락**했다. REQ-ALD2-005a 는 절을 옮길 때마다 stub 에 포인터 줄을 쓰라고 요구하고, 그 줄은 always-loaded 표면으로 **되돌아오는 문자**다. 종전 표의 "감축 목표"는 총량(gross)이었을 뿐 순감축(net)이 아니었다. 아래는 재유입을 실측해 다시 세운 투영이다.

### §3.1 재유입률 실측

감사가 제시한 ~16,300 추정치를 물려받지 않고 직접 쟀다. 감사의 grep 은 **다른 룰로 가는 교차참조까지** 세므로 과대계상된다(감사 보고서 자신이 밝힌 한계). 과대계상을 제거한 측정: 각 stub 에서 **자기 companion 파일명을 담은 줄**만 포인터로 세고 길이를 합산했다.

이미 분리된 stub 14개 기준: **포인터 69줄, 20,175자, 평균 292.4자/줄**. 집계 재유입률(포인터 자수 ÷ companion 자수) **8.67%**.

파일별 편차가 크다 — `moai-mcp-tools` 0.86% 부터 `kanban-dispatch` 17.44% 까지. 이 카드가 가장 많이 옮기는 파일이 정확히 고율 파일이므로 **집계율로 투영하면 과소평가된다.** 따라서 파일별 가중으로 계산한다.

**이 측정이 하한인 이유**: 분모인 companion 자수에는 companion 이 원본 외에 획득한 내용이 섞여 있다(`session-handoff-examples.md` 가 부모보다 큰 선례 — REQ-ALD-005 가 금지하는 바로 그 현상). 분모가 부푼 만큼 **실제 재유입률은 실측치보다 높을 수 있다.** §3.3 의 민감도는 그래서 붙였다.

### §3.2 개정된 파일별 목표

| 파일 | 풀 | 종전 | **개정 목표(gross)** | 재유입률 | 재유입 | **net** |
|---|---:|---:|---:|---:|---:|---:|
| `kanban-dispatch.md` | 20,134 | 16,000 | 16,000 | 17.44% | 2,790 | 13,210 |
| `agent-common-protocol.md` | 20,478 | 15,000 | 15,000 | 12.88% | 1,932 | 13,068 |
| `askuser-protocol.md` | 15,726 | 11,000 | 11,000 | 6.17% | 679 | 10,321 |
| `session-handoff.md` | 13,498 | 10,000 | 10,000 | 8.26% | 826 | 9,174 |
| `cross-session-messaging.md` | 13,835 | 9,000 | 9,000 | 7.85% | 706 | 8,294 |
| `moai-constitution.md` | 10,094 | 7,000 | **8,500** | 1.88% | 160 | 8,340 |
| `CLAUDE.md` | 16,274 | 8,000 | 8,000 | 8.67% | 694 | 7,306 |
| `context-window-management.md` | 8,887 | 5,000 | **6,500** | 3.10% | 202 | 6,298 |
| `verification-claim-integrity.md` | 8,592 | 6,000 | 6,000 | 8.18% | 491 | 5,509 |
| `moai-mcp-tools.md` | 6,751 | 4,000 | **5,500** | 0.86% | 47 | 5,453 |
| `main-checkout-branch-guard.md` | 6,158 | 3,500 | **5,000** | 1.69% | 84 | 4,916 |
| `goal-directive.md` | 6,877 | 3,000 | **4,500** | 3.75% | 169 | 4,331 |
| `AGENTS.md` | 13,559 | 4,000 | 4,000 | **0%** | 0 | 4,000 |
| `cache-aware-execution.md` | 5,048 | 3,000 | **4,000** | 1.87% | 75 | 3,925 |
| `skill-routing.md` | 3,246 | 1,800 | **2,200** | 4.37% | 96 | 2,104 |
| `native-idiom-and-register.md` | 2,975 | 1,500 | **2,000** | 3.39% | 68 | 1,932 |
| yaml 2개 | — | 0 | 0 | — | 0 | 0 |
| **합계** | | 107,800 | **117,200** | | **9,019** | **108,181** |

`AGENTS.md` 의 재유입이 0인 것은 M2(제자리 압축)가 포인터를 만들지 않기 때문이다 — companion 이 없으므로 가리킬 것도 없다. 이 축에서는 M2 가 M1 보다 효율적이라는 뜻이며, 기존 29% 목표가 낮아 보였던 것과는 별개의 사실이다.

### §3.3 판정 — 목표는 여전히 통과하되, 재배분이 필요했다

**재유입은 상수가 아니라 이동량에 비례하므로, 목표 집합마다 값이 다르다.** 한 칸에 합쳐 적으면 D5 를 만든 바로 그 오해 — 재유입이 이동량과 무관하다는 직관 — 를 그 수정 안에 다시 심게 된다. 두 행을 따로 적는다.

| 목표 집합 | gross | 재유입 | net | 요구 | 여유 |
|---|---:|---:|---:|---:|---:|
| 종전 (v0.2.0) | 107,800 | **8,796** | 99,004 | 96,943 | +2,061 |
| **개정 (v0.3.0)** | **117,200** | **9,019** | **108,181** | 96,943 | **+11,238** |

**행별 산출식**: 재유입 = Σ(파일별 목표 × 파일별 재유입률), 실제로 이동하는 파일에 대해서만 — `AGENTS.md` 는 M2 라 율 0% 로 합에 기여하지 않는다. §3.2 표의 재유입 열이 개정 행의 항별 값이고, 종전 행은 같은 율에 종전 목표를 곱한 것이다. 두 행의 차이 223자는 gross 증가분 9,400자가 **저재유입 파일에만** 배정된 결과다(9,400 × 평균 2.4%).

post-state: 246,943 − 108,181 = **138,762** < 150,000.

**감사가 예측한 5,443자 부족은 성립하지 않는다.** 감사의 추정(~16,300)이 과대계상이었고, 실측 가중치는 종전 목표에서 8,796·개정 목표에서 9,019다. 다만 **종전 목표를 그대로 뒀다면** net 은 99,004, 여유는 10,857 → **2,061** 로 깎였다. 그 여유는 §B 가 이미 공개한 휴리스틱 오차(문단 분류가 풀을 과대평가)를 흡수하기에 부족하다 — 여유가 존재했던 이유가 사라지지 않았으므로, 목표를 올려 여유를 복원했다.

**재배분 원칙**: 재유입률이 낮고 풀에 잔량이 있는 파일 7개의 목표만 올렸다(`moai-constitution`·`context-window-management`·`moai-mcp-tools`·`main-checkout-branch-guard`·`goal-directive`·`cache-aware-execution`·`skill-routing`·`native-idiom`). 고율 파일(`kanban-dispatch` 17.44%, `agent-common-protocol` 12.88%)은 **올리지 않았다** — 거기서 더 옮기면 재유입이 같이 늘어 순이득이 적다.

민감도 (실측률이 하한이므로 필수):

| 실제 재유입률 | 재유입 | net | 여유 | 판정 |
|---|---:|---:|---:|---|
| 실측 그대로 | 9,019 | 108,181 | +11,238 | 통과 |
| 실측 ×1.2 | 10,822 | 106,378 | +9,435 | 통과 |
| 실측 ×1.5 | 13,528 | 103,672 | +6,729 | 통과 |
| 실측 ×2.0 | 18,037 | 99,163 | +2,220 | 통과 |

각 행: 재유입 = round(9,018.6 × 배수), net = 117,200 − 재유입, 여유 = net − 96,943.

실측의 2배여도 통과한다. **종전 목표로는 ×1.5 에서 이미 미달이었다** — 재유입 13,194, net 94,606, 부족 2,337.

**이것은 문서 편집이 아니라 설계 변경이다.** 파일별 목표 7개가 바뀌었고, `plan.md` 의 마일스톤 목표 수치도 함께 바뀐다. 감사 보고서의 "모든 수리는 문서 편집, 설계 변경 없음" 분류는 이 항목에 대해서는 맞지 않다.

`AGENTS.md` 의 30%(4,000/13,559)는 다른 행보다 낮다. M2 만 허용되기 때문이며(REQ-ALD2-011), 절충이 아니라 그 요구사항의 직접적 귀결이다. `CLAUDE.md` 의 49% 도 같은 이유로 보수적이다 — 중복 제거는 "정본이 실제로 다른 곳에 있는가"를 건건이 확인해야 하고, 확인되지 않은 문단은 그대로 둔다.

---

## §4. 재배치 표 (계획 — run phase 에서 절 이름 단위로 채워진다)

REQ-ALD2-010 이 요구하는 산출물의 **양식**이다. 실제 절 이름은 run phase 에서 각 파일을 열어 확정한다 — plan phase 에서 절 목록을 미리 못 박으면 열어 보고 달라졌을 때 SPEC 을 고쳐야 한다.

| # | 출발 stub | 옮기는 절 | 목적지 companion | companion `paths:` | domain-keyed 근거 | 범위 의존 |
|---|---|---|---|---|---|---|
| R-01 | `kanban-dispatch.md` | *(run 에서 확정)* | `kanban-dispatch-detail.md` | *(run 에서 확정)* | *(run 에서 확정)* | *(run 에서 확정)* |
| … | … | … | … | … | … | … |

각 행이 만족해야 하는 조건:

1. **옮기는 절**은 `[HARD]` / `MUST` / `MUST NOT` / `shall ` 을 한 줄도 포함하지 않는다(REQ-ALD2-002·003).
2. **`paths:`** 는 부모 stub 경로만 담지 않는다. 그 companion 이 다루는 작업을 하는 동안 세션이 실제로 건드리는 경로를 함께 키로 잡는다(REQ-ALD2-004).
3. **domain-keyed 근거**는 "이 companion 이 필요해지는 작업은 무엇이며 그 작업이 어느 경로를 건드리는가"를 한 문장으로 답한다.
4. **범위 의존**은 그 절을 참조하는 stub 내 구속 조항을 **열거**하고, 조항마다 두 질문에 모두 답한다 — **Q1** 이동 후에도 자기 범위 문맥에 도달하는가(축소 방향), **Q2** 범위가 이동 전보다 넓어지지 않았는가(확대 방향). 참조하는 조항이 없으면 `없음` 이라 적는다 — 빈칸으로 두지 않는다. `없음` 은 확인했다는 뜻이고 빈칸은 보지 않았다는 뜻이며, 둘은 같지 않다.

   Q2 가 따로 필요한 이유: 범위를 **좁히던** 비구속 문장이 절과 함께 떠나면 남은 구속 조항이 더 넓게 읽힌다. 결손이 아니라 증가 방향의 실패라 손실 탐지기가 발화하지 않고, Q1 만 묻는 칸은 그것을 통과시킨다.

4번 칸이 존재하는 이유는 §7 의 첫 번째 위험(범위 이탈)을 **행 단위 산출물로 끌어내리기 위해서**다. 동결 해시는 구속 줄을 지키지만 그 줄을 범위 짓는 문맥은 지키지 않으므로, 이 칸이 비면 그 위험은 아무 데서도 판정되지 않는다. 전문: `acceptance.md §D.3`.

선례로 삼을 `paths:` — `goal-directive-detail.md` 가 유일한 domain-keyed 선례다. self-keyed 반례: `paths` 에 부모 룰 파일 하나만 적으면 그 룰을 **편집할 때만** 로드되므로, 그 룰이 지배하는 작업을 하는 세션에는 영원히 도달하지 않는다.

---

## §5. 앵커 무결성 (REQ-ALD2-009)

절이 stub 에서 companion 으로 옮겨지면, 그 절을 `§`·절 제목·파일명+절 형태로 가리키던 교차참조가 끊길 수 있다. 끊김은 조용하다 — 어떤 빌드도 실패하지 않는다.

점검 방식(run phase 에서 실행):

- 옮긴 절마다 그 절 제목 문자열을 저장소 전체에서 역방향 grep 한다.
- 히트한 모든 인용에 대해, 인용이 가리키는 파일이 **이동 후의 위치**를 가리키는지 확인한다.
- stub 에 남는 포인터 줄이 옮긴 절을 **이름으로** 호명하므로(REQ-ALD2-005a), stub 을 경유하는 인용은 자동으로 살아난다 — 확인 대상은 companion 을 직접 가리키지 않던 인용이다.

---

## §6. template ↔ live 동등 (REQ-ALD2-007)

이 카드가 건드리는 모든 파일은 사본이 둘이다. 두 사본은 같은 커밋에서 함께 바뀐다.

주의할 기존 분기 하나가 이미 발견됐다 — `skill-routing.md`: 라이브 사본은 `paths:` frontmatter 를 갖는데 템플릿 사본에는 frontmatter 가 **아예 없다**. 그대로 두면 사용자 프로젝트에 always-loaded 로 배포된다. 이 카드가 고친다(REQ-ALD2-008).

`CLAUDE.md` / `AGENTS.md` 도 템플릿 미러를 갖는지 run phase 착수 시 확인한다 — 확인 전에 가정하지 않는다.

---

## §7. 위험

| 위험 | 완화 |
|---|---|
| **범위 이탈 (방향 1 — 축소)** — 절 제목·한정 표·참조 열거가 떠나고 구속 줄만 남아, 해시는 통과하는데 그 줄이 구속할 대상이 사라짐 | `§4` 범위 의존 칸 **Q1** + `plan.md §C` M1 조건 1 단서(구속 조항이 의존하는 표·제목인 절은 옮기지 않는다). 해시로는 안 잡히므로 판정은 검토 몫 |
| **범위 이탈 (방향 2 — 확대)** — 범위를 **좁히던** 비구속 문장이 지워져 남은 구속 조항이 이전보다 넓게 읽힘. 결손이 아니라 **증가** 방향이라 손실 탐지기가 발화하지 않는다 | `§4` 범위 의존 칸 **Q2** + 같은 단서의 확대판(범위를 좁히는 문장은 비구속이어도 지우지 않는다). 전문 `acceptance.md §D.3` |
| companion 이 비대해짐 | **합계 축에서만** 허용된다 — companion 은 always-loaded 가 아니다. **파일당 축에서는 40,000자에서 막힌다**(REQ-ALD2-014). REQ-ALD2-006(원본에 없던 내용 금지)도 함께 구속 |
| 파일별 목표 중 일부 미달 | 개정 투영(§3.3)에 **net 여유 11,238자**, 재유입이 실측의 2배여도 통과(+2,220). 회수처는 **재유입률이 낮고** 풀에 잔량이 있는 파일 — `moai-mcp-tools.md`(0.86%)·`main-checkout-branch-guard.md`(1.69%). 고율 파일(`kanban-dispatch` 17.44%)에서 더 옮기는 것은 순이득이 작아 회수처로 쓰지 않는다 |
| `AGENTS.md` 압축이 의미를 깎음 | M2 는 구속 조항 줄을 못 건드리고, 구속 조항 해시가 그것을 강제한다. 비구속 산문의 의미 보존은 REQ-ALD2-012 가 구속하며 기계 검사가 없다 — 검토로 판정 |

🗿 MoAI
