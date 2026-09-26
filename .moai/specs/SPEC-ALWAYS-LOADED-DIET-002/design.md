# SPEC-ALWAYS-LOADED-DIET-002 — 설계

파일별 분리 계획과 재배치 표. 근거 측정은 `.moai/reports/t1175/discovery.md` 와 이 문서 §1 의 자체 실측.

---

## §1. 파일별 실측 (라이브 트리, 이 실행)

> **[개정 — 2026-09-25, plan-audit iter3 N6]** 종전 표에는 **생성기 산물**이 있었다. 측정 스크립트가 문단마다 `len(para)+2` 로 구분자 2자를 더해, 파일마다 합이 `현재+2` 가 되고 총계가 부풀었다. 가장 뚜렷한 증상은 `goal-directive` 의 재배치 6,877 — **파일 자체가 6,875자라 어느 트리에서도 산출 불가능한 값**이었다. 구분자를 어느 쪽에도 귀속시키지 않는 방식으로 다시 도출했다.

측정 대상은 **라이브 트리**(`.claude/rules/moai/`, 그리고 저장소 루트의 `CLAUDE.md`·`AGENTS.md`). 문단 단위 분류 — 문단이 `[HARD]` / `MUST` / `MUST NOT` / `shall ` 중 하나라도 담으면 **구속**, 아니면 **재배치 가능**. 문단 사이 빈 줄은 **어느 쪽도 아니므로** 두 열의 합은 `현재` 보다 작다.

| 파일 | 현재 | 구속 | 재배치 가능 | 구분자 | 가능 비율 |
|---|---:|---:|---:|---:|---:|
| `workflow/kanban-dispatch.md` | 39,077 | 18,867 | 19,974 | 236 | 51.1% |
| `core/agent-common-protocol.md` | 28,381 | 7,861 | 20,284 | 236 | 71.5% |
| `core/askuser-protocol.md` | 24,466 | 8,702 | 15,554 | 210 | 63.6% |
| `workflow/session-handoff.md` | 19,908 | 6,392 | 13,392 | 124 | 67.3% |
| `CLAUDE.md` | 19,305 | 3,025 | 16,120 | 160 | 83.5% |
| `core/moai-constitution.md` | 17,272 | 7,152 | 9,962 | 158 | 57.7% |
| `core/verification-claim-integrity.md` | 16,636 | 8,012 | 8,516 | 108 | 51.2% |
| `AGENTS.md` | 16,374 | 2,807 | 13,437 | 130 | 82.1% |
| `workflow/cross-session-messaging.md` | 16,330 | 2,485 | 13,747 | 98 | 84.2% |
| `workflow/context-window-management.md` | 11,447 | 2,549 | 8,823 | 75 | 77.1% |
| `core/moai-mcp-tools.md` | 7,400 | 649 | 6,709 | 42 | 90.7% |
| `workflow/cache-aware-execution.md` | 6,907 | 1,851 | 5,018 | 38 | 72.7% |
| `workflow/goal-directive.md` | 6,875 | 0 | 6,833 | 42 | 99.4% |
| `workflow/main-checkout-branch-guard.md` | 6,814 | 652 | 6,096 | 66 | 89.5% |
| `workflow/skill-routing.md` | 5,709 | 2,455 | 3,212 | 42 | 56.3% |
| `core/native-idiom-and-register.md` | 3,815 | 838 | 2,943 | 34 | 77.1% |
| `language.yaml` + `user.yaml` | 227 | 0 | 227 | 0 | — |
| **합계** | **246,943** | **74,297** | **170,847** | 1,799 | **69.2%** |

검산: 74,297 + 170,847 + 1,799 = 246,943 ✓ (열 셋이 파일 전체를 남김없이 분할한다).

**종전 값과의 차이**: 재배치 가능 풀 172,363 → **170,847** (−1,516). 구분자 1,799자 중 대부분이 종전에 재배치 가능으로 잘못 귀속돼 있었다. 이 수정은 §B 가 인용하는 헤드라인 수치를 움직이므로 자릿수만 고치지 않고 그 주장도 다시 검토했다 — 결론은 §B 참조: 풀 170,847 은 요구치 96,943 의 **1.76배**(종전 1.78배)이고, 구속 조항을 강등하지 않고 도달 가능하다는 결론은 유지된다.

**개정 목표(§3.2) 대비 파일별 적합성도 다시 확인했다** — gross 117,200 은 풀의 68.6%이고, 16개 파일 전부에서 목표가 그 파일의 풀 안에 들어간다. 가장 빡빡한 것은 `moai-constitution`(8,500 / 9,962 = 85%)이다.

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

**수치의 트리 표기 (D4 — iter3 에서 회귀 판정, 재측정 후 정정)**

> **[개정 — iter3 D4]** 종전 수리는 "이 절과 §2.1 의 모든 자수는 라이브 트리 실측"이라고 **단언했는데 재측정 없이 쓴 문장이었고, 사실과 반대였다.** 실제로는 이 절의 표가 **템플릿** 값(`agent-common-protocol-reference` 31,174 등), §2.1 이 **라이브** 값(31,207 등)이었다. 미측정 귀속은 아무것도 주장하지 않아 정직하지만, 측정한 것처럼 들리는 틀린 귀속은 그렇지 않다 — Gap 을 위반된 주장으로 승격시킨 셈이다. 두 트리를 다시 재고 아래에 쓴다.

**두 트리를 모두 쟀다.** 아래 표와 §2.1·§2.2 의 companion 자수는 전부 **라이브 트리**(`.claude/rules/moai/`)로 통일했다. §1 의 18파일 실측도 라이브다.

**두 트리가 갈리는 companion 이 14개 중 6개 있다**(라이브 − 템플릿): `main-checkout-branch-guard-detail` +362 · `moai-constitution-detail` +145 · `verification-claim-integrity-detail` +64 · `cross-session-messaging-detail` +61 · `agent-common-protocol-reference` +33 · `cache-aware-execution-reference` +25. 나머지 8개는 바이트 동일.

이 분기는 이 카드가 만든 것이 아니라 **선행 상태**다. 다만 이 카드가 그중 어느 것이든 목적지로 쓰면 REQ-ALD2-007(만든 변경의 미러)이 걸리므로, 해당 companion 은 수정 시 **이 카드가 만든 구조 변경**을 두 사본에 같은 커밋으로 미러한다.

> **[개정 2 — 리드 판정 2026-09-26] 종전 문구의 딸림말은 「REQ-ALD2-007(미러 동등)」이었고, 문장 끝은 「분기를 물려받은 채 덮지 않는다」였다.** 폐기 사유(한 줄): 개정 REQ-ALD2-007 은 「동등」을 술어에서 내렸고, `spec.md §D` 의 `### Out of Scope — 승계된 template/live 분기의 해소` 가 승계 분기의 해소를 명시적으로 **면제**하므로, 종전 문장은 현행 정책과 반대를 말했다. D5 가 닫은 여섯 자리와 §6 에 이미 붙은 같은 표시가 여기까지 오지 않은 일곱 번째 자리다(plan-audit iter6 **D18**).

수용량 판정(§2.1)에 라이브를 쓰는 이유: 파일당 한도를 발화시키는 것은 **세션이 실제로 로드하는 사본**이고, 이 저장소에서 그것은 라이브다. 템플릿 사본은 사용자 프로젝트에서 같은 역할을 하므로 REQ-ALD2-007 이 따로 덮는다.

| stub (always-loaded) | 기존 companion (path-scoped) | companion 현재 크기 (라이브) |
|---|---|---:|
| `workflow/kanban-dispatch.md` | `workflow/kanban-dispatch-detail.md` | 41,036 |
| `core/agent-common-protocol.md` | `core/agent-common-protocol-reference.md` | 31,207 |
| `core/askuser-protocol.md` | `core/askuser-protocol-reference.md` | 15,646 |
| `workflow/session-handoff.md` | `workflow/session-handoff-examples.md` | 41,616 |
| `core/moai-constitution.md` | `core/moai-constitution-detail.md` | 7,644 |
| `core/verification-claim-integrity.md` | `core/verification-claim-integrity-detail.md` | 21,347 |
| `workflow/cross-session-messaging.md` | `workflow/cross-session-messaging-detail.md` | 15,221 |
| `workflow/context-window-management.md` | `workflow/context-window-management-detail.md` | 7,486 |
| `core/moai-mcp-tools.md` | `core/moai-mcp-tools-catalogue.md` | 11,424 |
| `workflow/cache-aware-execution.md` | `workflow/cache-aware-execution-reference.md` | 5,841 |
| `workflow/goal-directive.md` | `workflow/goal-directive-detail.md` | 20,283 |
| `workflow/main-checkout-branch-guard.md` | `workflow/main-checkout-branch-guard-detail.md` | 9,984 |
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

### §2.2 목적지 배정 — 절 단위 재측정 반영

§2 규칙 3("이미 넘긴 companion 에는 옮기지 않는다")과 §4.1 의 절 단위 재배치 표를 적용한 결과다. §2.1 의 유입량은 절 단위 판정 전의 초기 목표이므로 실제 배정량은 아래 값을 따른다.

| 출발 stub | 유입량 | 개정된 목적지 | 이후 |
|---|---:|---|---:|
| `kanban-dispatch.md` | 5,121 | **신규** companion (주제 경계 분리) | 5,121 |
| `session-handoff.md` | 5,442 | **신규** companion | 5,442 |
| `agent-common-protocol.md` | 7,108 | 기존 reference | 38,315 |
| 나머지 stub | §4.1 표대로 | 기존 companion 유지 | §4.1 표 |

신규 companion 2개가 생긴다. 각각 REQ-ALD2-004·005·006 을 그대로 지며, AC-ALD2-008(증식 금지)에서 "작업 전 크기"는 0 으로 계산한다.

`agent-common-protocol-reference.md` 의 38,315 는 한도 아래지만 여유가 1,685자뿐이다. run phase 에서 실제 이동량을 다시 재고 40,000자 미만 제약을 확인한다.

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
| `kanban-dispatch.md` | 19,974 | 16,000 | 16,000 | 17.44% | 2,790 | 13,210 |
| `agent-common-protocol.md` | 20,284 | 15,000 | 15,000 | 12.88% | 1,932 | 13,068 |
| `askuser-protocol.md` | 15,554 | 11,000 | 11,000 | 6.17% | 679 | 10,321 |
| `session-handoff.md` | 13,392 | 10,000 | 10,000 | 8.26% | 826 | 9,174 |
| `cross-session-messaging.md` | 13,747 | 9,000 | 9,000 | 7.85% | 706 | 8,294 |
| `moai-constitution.md` | 9,962 | 7,000 | **8,500** | 1.88% | 160 | 8,340 |
| `CLAUDE.md` | 16,120 | 8,000 | 8,000 | 8.67% | 694 | 7,306 |
| `context-window-management.md` | 8,823 | 5,000 | **6,500** | 3.10% | 202 | 6,298 |
| `verification-claim-integrity.md` | 8,516 | 6,000 | 6,000 | 8.18% | 491 | 5,509 |
| `moai-mcp-tools.md` | 6,709 | 4,000 | **5,500** | 0.86% | 47 | 5,453 |
| `main-checkout-branch-guard.md` | 6,096 | 3,500 | **5,000** | 1.69% | 84 | 4,916 |
| `goal-directive.md` | 6,833 | 3,000 | **4,500** | 3.75% | 169 | 4,331 |
| `AGENTS.md` | 13,437 | 4,000 | 4,000 | **0%** | 0 | 4,000 |
| `cache-aware-execution.md` | 5,018 | 3,000 | **4,000** | 1.87% | 75 | 3,925 |
| `skill-routing.md` | 3,212 | 1,800 | **2,200** | 4.37% | 96 | 2,104 |
| `native-idiom-and-register.md` | 2,943 | 1,500 | **2,000** | 3.39% | 68 | 1,932 |
| yaml 2개 | — | 0 | 0 | — | 0 | 0 |
| **합계** | | 107,800 | **117,200** | | **9,019** | **108,181** |

`AGENTS.md` 의 재유입이 0인 것은 M2(제자리 압축)가 포인터를 만들지 않기 때문이다 — companion 이 없으므로 가리킬 것도 없다. 이 축에서는 M2 가 M1 보다 효율적이라는 뜻이며, 기존 29% 목표가 낮아 보였던 것과는 별개의 사실이다.

### §3.3 판정 — 목표는 여전히 통과하되, 재배분이 필요했다

> **[개정 2 — 리드 판정 2026-09-26] 이 절 전체는 폐기된 목표 150,000 에 대고 계산된 plan-phase 투영이다. 표를 지우지 않고 남기되, 현행 기준이 아님을 여기서 못 박는다.**
>
> - 이 절의 `여유` 열(**+2,061** · **+11,238**)과 아래 민감도 표의 `여유` 열은 전부 `net − 96,943` 이고, 그 96,943 은 `246,943 − 150,000` 이다. **150,000 이 도달 불가로 확정되면서 이 「여유」는 정의되지 않는 수치가 됐다** — 근거는 `acceptance.md §AC-ALD2-001` 개정 표시.
> - `:183` 의 post-state **138,762** 와 §4.4 `:581` 의 post-state **137,569** 는 둘 다 `acceptance.md §AC-ALD2-001.1` 이 같은 census 에서 도출한 **최저 하한 `F₀ = 149,195` 보다 낮다**(각각 10,433 · 11,626 아래). 즉 **구조상 달성 불가한 post-state 를 달성 가능으로 단언한 문장**이며, 아래 본문은 그대로 두되 이 사실을 나란히 읽어야 한다.
> - **현행 판정 기준**은 문턱이 아니라 기록 술어다 — `acceptance.md §AC-ALD2-001`. 이 절의 어떤 수치도 그 AC 의 PASS·FAIL 을 가르지 않는다.
> - 실측 결과(이 카드): 합계 **197,897**, 감축 49,046, 런타임 한도 150,000 에 대한 잔여 **47,897**. 이 절의 투영(감축 gross 117,200 / net 108,181)은 실측의 **약 2.2배**였다.

**재유입은 상수가 아니라 이동량에 비례하므로, 목표 집합마다 값이 다르다.** 한 칸에 합쳐 적으면 D5 를 만든 바로 그 오해 — 재유입이 이동량과 무관하다는 직관 — 를 그 수정 안에 다시 심게 된다. 두 행을 따로 적는다.

| 목표 집합 | gross | 재유입 | net | 요구 | 여유 |
|---|---:|---:|---:|---:|---:|
| 종전 (v0.2.0) | 107,800 | **8,796** | 99,004 | 96,943 | +2,061 |
| **개정 (v0.3.0)** | **117,200** | **9,019** | **108,181** | 96,943 | **+11,238** |

**행별 산출식**: 재유입 = Σ(파일별 목표 × 파일별 재유입률), 실제로 이동하는 파일에 대해서만 — `AGENTS.md` 는 M2 라 율 0% 로 합에 기여하지 않는다. §3.2 표의 재유입 열이 개정 행의 항별 값이고, 종전 행은 같은 율에 종전 목표를 곱한 것이다. 두 행의 차이 223자는 gross 증가분 9,400자가 **저재유입 파일에만** 배정된 결과다(9,400 × 평균 2.4%).

post-state: 246,943 − 108,181 = **138,762** < 150,000. **[개정 2 — 폐기]** 이 단언은 성립하지 않는다 — 138,762 는 `F₀ = 149,195` 보다 **10,433 낮아** 구조상 도달 불가한 post-state 다. 문장을 지우지 않고 남기는 이유는 이 투영이 그때 무엇을 믿고 있었는지가 기록이기 때문이다. 현행 기준은 `acceptance.md §AC-ALD2-001`(기록 술어).

**감사가 예측한 5,443자 부족은 성립하지 않는다.** 감사의 추정(~16,300)이 과대계상이었고, 실측 가중치는 종전 목표에서 8,796·개정 목표에서 9,019다. 다만 **종전 목표를 그대로 뒀다면** net 은 99,004, 여유는 10,857 → **2,061** 로 깎였다. 그 여유는 §B 가 이미 공개한 휴리스틱 오차(문단 분류가 풀을 과대평가)를 흡수하기에 부족하다 — 여유가 존재했던 이유가 사라지지 않았으므로, 목표를 올려 여유를 복원했다.

**재배분 원칙**: 재유입률이 낮고(0.86%~4.37%) 풀에 잔량이 있는 파일 **8개**의 목표만 올렸다 — `moai-constitution`(7,000→8,500)·`context-window-management`(5,000→6,500)·`moai-mcp-tools`(4,000→5,500)·`main-checkout-branch-guard`(3,500→5,000)·`goal-directive`(3,000→4,500)·`cache-aware-execution`(3,000→4,000)·`skill-routing`(1,800→2,200)·`native-idiom`(1,500→2,000). 증가분 합 9,400자. 고율 파일(`kanban-dispatch` 17.44%, `agent-common-protocol` 12.88%)은 **올리지 않았다** — 거기서 더 옮기면 재유입이 같이 늘어 순이득이 적다.

민감도 (실측률이 하한이므로 필수) — **[개정 2 — 폐기] 아래 표의 `여유` 열과 `판정` 열은 전부 폐기된 목표 150,000 기준이므로 현행 판정이 아니다.** 네 행의 「통과」는 어느 것도 이 카드의 판정을 뜻하지 않는다:

| 실제 재유입률 | 재유입 | net | 여유 | 판정 |
|---|---:|---:|---:|---|
| 실측 그대로 | 9,019 | 108,181 | +11,238 | 통과 |
| 실측 ×1.2 | 10,822 | 106,378 | +9,435 | 통과 |
| 실측 ×1.5 | 13,528 | 103,672 | +6,729 | 통과 |
| 실측 ×2.0 | 18,037 | 99,163 | +2,220 | 통과 |

각 행: 재유입 = round(9,018.6 × 배수), net = 117,200 − 재유입, 여유 = net − 96,943.

실측의 2배여도 통과한다. **종전 목표로는 ×1.5 에서 이미 미달이었다** — 재유입 13,194, net 94,606, 부족 2,337.

**이것은 문서 편집이 아니라 설계 변경이다.** 파일별 목표 **8개**가 바뀌었고, `plan.md` 의 마일스톤 목표 수치도 함께 바뀐다. 감사 보고서의 "모든 수리는 문서 편집, 설계 변경 없음" 분류는 이 항목에 대해서는 맞지 않다.

`AGENTS.md` 의 30%(4,000/13,437)는 다른 행보다 낮다. M2 만 허용되기 때문이며(REQ-ALD2-011), 절충이 아니라 그 요구사항의 직접적 귀결이다. `CLAUDE.md` 의 50%(8,000/16,120)도 같은 이유로 보수적이다 — 중복 제거는 "정본이 실제로 다른 곳에 있는가"를 건건이 확인해야 하고, 확인되지 않은 문단은 그대로 둔다.

---

## §4. 재배치 표 (M1 에서 채워짐 — 2026-09-25, HEAD `172ef22eb`, 이 워크트리)

REQ-ALD2-010 이 요구하는 산출물이다. plan phase 에는 **양식**만 있었고, M1 이 절 이름 단위로 채웠다. 아래 모든 자수는 **이 실행에서 이 트리를 직접 재서** 얻은 값이다. 계획값(실측이 아니라 배정량)인 칸은 `(계획)` 으로 표시하고, `discovery.md` 나 §3.2 투영에서 인용한 값은 그렇게 밝힌다.

**표를 채우기 전에 `acceptance.md` AC-ALD2-004 의 선택 술어를 읽었다**(iter4 N7 선행 조건). 범위 의존 칸은 세 가지 검사를 돌린 결과가 아니라 **"이 절이 사라지면 이 조항이 구속하는 대상이 달라지는가"** 를 행마다 판단한 결과이며, `없음(판단)` 은 그 정의에 비추어 판단했다는 뜻이다.

### §4.0 측정 방법과 네 조건의 증거

절 경계는 fence(백틱 3개) 밖의 `#`~`###` 제목으로 잡았다 — bash 주석 줄(`# …`)이 제목으로 오인되던 것을 fence 추적으로 제거했다. 구속 조항 판정식은 AC-ALD2-002 와 같은 `\[HARD\]|MUST|shall ` 이다.

```bash
# 조건 1 — 절마다 구속 조항 줄 개수 (0 이어야 이동 가능)
python3 /tmp/claude-501/sec.py 3      # 16파일 절별 자수 + bind= 개수, fence 인식
# 조건 2 — 역방향 인용 열거 (절 제목의 §-인용 형태)
grep -rIn -F -- "§ <절 제목>" --include='*.md' . \
  | grep -v '^\./\.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/'
# 조건 2 부속 — 그중 초과 파일 4개 안의 인용 (§5 선례 대상)
grep -rIn -F -- "§ <절 제목>" --include='*.md' \
  .claude/rules/moai/workflow/worktree-integration.md \
  .claude/rules/moai/workflow/session-handoff-examples.md \
  .claude/rules/moai/workflow/kanban-dispatch-detail.md \
  .claude/rules/moai/workflow/spec-workflow.md
# 조건 3 — 목적지 paths: 현재 값
grep -H -m2 '^paths:' .claude/rules/moai/<companion>.md
# 조건 4 — 목적지 현재 크기 (라이브, 이 실행)
wc -m .claude/rules/moai/<companion>.md
```

#### 조건 1 의 집계 — M1 이 찾아낸 가장 큰 사실

파일별 "구속 조항 줄이 0인 절"의 자수 합(L1 서문 제외)을 §3.2 의 gross 목표에 대고 재면:

| 파일 | §3.2 gross 목표 | **절 단위 이동 가능 풀** | 차이 |
|---|---:|---:|---:|
| `kanban-dispatch.md` | 16,000 | 6,415 | −9,585 |
| `agent-common-protocol.md` | 15,000 | 7,240 | −7,760 |
| `askuser-protocol.md` | 11,000 | 8,777 | −2,223 |
| `session-handoff.md` | 10,000 | 6,925 | −3,075 |
| `cross-session-messaging.md` | 9,000 | 7,857 | −1,143 |
| `moai-constitution.md` | 8,500 | 4,628 | −3,872 |
| `CLAUDE.md` | 8,000 | 14,859 | +6,859 |
| `context-window-management.md` | 6,500 | 7,200 | +700 |
| `verification-claim-integrity.md` | 6,000 | **0** | −6,000 |
| `moai-mcp-tools.md` | 5,500 | 4,529 | −971 |
| `main-checkout-branch-guard.md` | 5,000 | 3,945 | −1,055 |
| `goal-directive.md` | 4,500 | 6,291 | +1,791 |
| `AGENTS.md` | 4,000 | 8,489 | +4,489 |
| `cache-aware-execution.md` | 4,000 | 1,270 | −2,730 |
| `skill-routing.md` | 2,200 | 1,471 | −729 |
| `native-idiom-and-register.md` | 2,000 | 2,436 | +436 |
| **합계** | **117,200** | **93,641** | **−23,559** |

> **[개정 2 — 리드 판정 2026-09-26] 이 표의 풀 열 합계는 재현되지 않는다 — Gap 으로 적어 둔다.**
>
> 16행의 풀 열을 합산하면 **92,332** 이고 적힌 합계는 **93,641**, 차이 **1,309** 다. 재측정 명령은 **줄번호가 아니라 표 머리와 합계 행에 앵커한다** — 줄번호를 박으면 이 주석을 끼워 넣는 것만으로 명령이 다른 구간을 재게 된다:
>
> ```bash
> awk '/^\| 파일 \| §3\.2 gross 목표/,/^\| \*\*합계\*\*/' design.md | grep -v '합계' \
>   | awk -F'|' '{gsub(/[ ,*`]/,"",$4); if ($4 ~ /^[0-9]+$/) {s+=$4; n++}} END {print "rows="n" POOL_COLUMN_SUM="s}'
> rows=16 POOL_COLUMN_SUM=92332
> ```
>
> 같은 명령의 필드를 `$4` → `$3`(gross 열)로 바꾸면 `rows=16 GROSS_COLUMN_SUM=117200` 이 나와 적힌 합계와 일치하므로, 어긋난 것은 풀 열 하나다 — 이것이 「합산 방식이 틀렸다」가 아니라 「이 열이 어긋난다」를 가르는 대조군이다.
>
> 1,309 는 `§4.3` 기각 표의 `kanban-dispatch.md ## Scope — when this rule is live` 행 자수와 같고, 이 표의 kanban 칸 `6,415 + 1,309 = 7,724` 는 `pool-census.md` 의 kanban 풀과 정확히 일치한다. 두 재조정이 모두 가능하다 — **(a)** kanban 칸이 기각분을 선차감했고 열 합 92,332 가 맞다, **(b)** 적힌 93,641 이 맞고 그 1,309 가 `J` 에도 들어가 **이중 차감**돼 있다.
>
> **이 카드는 둘 중 하나를 고르지 않는다.** 고르려면 이 절이 쓴 절 분할 스크립트를 원 트리에서 다시 돌려야 하고, 그것이 M1 측정의 재실행이다. 귀결: 구조적 하한 F 는 단일값이 아니라 **세 값의 구간** `{170,528 · 171,695 · 172,863}` 으로 적힌다(`acceptance.md §AC-ALD2-001.1`). F 가 이 카드의 판정에서 내려왔으므로(같은 절) 이 미해결은 판정을 뒤집지 않으며, 확정은 `A_adm` 실측 후속 카드가 같은 자리에서 함께 한다.

§1 의 풀 170,847 은 **문단** 단위 분류이고, `plan.md §C` M1 조건 1 은 **절** 단위 판정이다("그 절 안에 구속 조항 줄이 한 줄도 없다"). 두 입도의 차이 77,206자는 **구속 조항 줄과 같은 절 안에 있는 비구속 산문**이며, 조건 1 때문에 **M1 으로는 손댈 수 없다.**

귀결 셋:

1. **M2 는 보조 기제가 아니라 필수 기제다.** 16파일 중 10개에서 §3.2 목표가 절 단위 풀을 넘고, `verification-claim-integrity.md` 는 이동 가능한 절이 **하나도 없다**(모든 절이 구속 조항 줄을 담는다 — §1 의 재배치 가능 8,516 은 전량 문단 단위다). 그 파일의 목표 6,000 은 100% M2 다.
2. **`moai-constitution.md` 가 가장 빡빡하다** — 절 단위 풀 4,628 에 목표 8,500 이고, §1 이 이미 지목한 85% 문단 비율과 같은 방향이다.
3. **§2.1 의 초기 목표에 따른 신규 companion 3개 가정 중 하나는 불필요하다.** `agent-common-protocol.md` 의 절 단위 이동량은 7,108 이고 기존 `agent-common-protocol-reference.md`(31,207)에 전량 넣어도 **38,315**(여유 1,685)로 한도 아래다. 초기 "기존 8,000 + 신규 7,000" 분할은 이동량 15,000 을 전제했지만 §4.1 의 절 단위 판정에서는 성립하지 않는다. §2.2 의 최종 배정은 신규 2개다. `kanban-dispatch.md`·`session-handoff.md` 용 신규 2개는 기존 목적지가 이미 초과라 **여전히 필요하다.**

#### 조건 3 의 집계 — self-keyed 목적지가 14개 중 5개

이 실행에서 목적지 14개의 `paths:` 를 전부 읽었다. **부모 stub 경로만 담은 self-keyed 가 5개**다: `agent-common-protocol-reference`(`**/agent-common-protocol.md`) · `askuser-protocol-reference`(`**/askuser-protocol.md`) · `cache-aware-execution-reference`(`**/cache-aware-execution.md`) · `session-handoff-examples`(`**/session-handoff.md`) · `verification-claim-integrity-detail`(`**/verification-claim-integrity*.md`, 자기 가족 글롭). 이 카드가 쓰는 목적지는 REQ-ALD2-004 에 따라 넓히며, 각 파일 절에 수정안을 적었다. 나머지 9개는 이미 domain-keyed 다.

---

### §4.1 M1 행 — companion 으로 이동

목적지 칸은 `현재 → 이후` 이며 두 수 모두 이 실행의 라이브 실측이다.

#### `core/agent-common-protocol.md` → `core/agent-common-protocol-reference.md` (31,207 → 38,315, 여유 1,685)

`paths:` 현재 `**/agent-common-protocol.md` — **self-keyed, 수정 필요**: `**/agent-common-protocol.md,**/.claude/agents/moai/*.md,**/.claude/skills/moai/workflows/*.md`. **domain-keyed 근거**(이 파일 행 공통): 이 companion 이 필요해지는 작업은 에이전트 정의 저작과 run/sync 워크플로 실행이며, 그 작업이 건드리는 경로가 이 둘이다.

| # | 기제 | 제거되는 절 | 자수 | 인용 | 범위 의존 — 지배받는 stub 내 구속 조항 · Q1 도달 / Q2 폭 |
|---|---|---|---:|---:|---|
| R-01 | M1 | `## Skeptical Evaluation Stance` | 795 | 13 | `없음(판단)` — reviewer 자세의 서술이며 stub 내 어떤 조항도 이 절의 용어·표·범위에 의존하지 않는다. 사라져도 구속 대상이 달라지는 조항이 없다 |
| R-02 | M1 | `### File Operations Pattern` | 498 | 3 | `[HARD] Agents must follow tool usage patterns…`(`## Tool Usage Guidelines`) · Q1 **도달** — 조항은 "패턴을 따르라"이고 패턴 목록은 stub 포인터가 이름으로 호명한다 / Q2 **불변** — 이 절은 범위를 좁히지 않고 내용을 채운다 |
| R-03 | M1 | `### Search Pattern` | 333 | 0 | 같은 조항 1건 · Q1 도달 / Q2 불변 |
| R-04 | M1 | `### Tool Selection by Task` | 753 | 20 | 같은 조항 1건 · Q1 도달 / Q2 불변. **주의**: `moai-constitution.md § Tool Selection Priority` 가 이 절을 SSOT 로 지목하며 인용 20건 — 이 파일에서 앵커 수리 부담이 가장 크다(§5), 그리고 R-71 과 연쇄한다 |
| R-05 | M1 | `### Bash Timeout` | 179 | 0 | 같은 조항 1건 · Q1 도달 / Q2 불변 |
| R-06 | M1 | `### Error Recovery Pattern` | 857 | 28 | 같은 조항 1건 + `### Ledger Closure` 의 `[HARD]` · Q1 **도달** — Ledger Closure 는 abort 를 자기 문장에서 정의하고 이 절을 참조하지 않는다 / Q2 **불변**. 인용 28건으로 이 카드 전체 최다 |
| R-07 | M1 | `### Super-Advisor Escalation (E1-E4)` | 1,114 | 8 | 같은 조항 1건 · Q1 도달 / Q2 불변 — E1~E4 는 진입 조건 열거이며 어떤 조항의 한정어도 아니다 |
| R-08 | M1 | `### Read-only verification batching` | 249 | 0 | `[HARD] The orchestrator MUST execute every read-only verification batch as a single-turn multi-Bash call`(`## Parallel Execution`) · Q1 **도달** — stub 이 세 의무(한 턴 배치·파일 리다이렉트 계약·증거 반출)를 제자리에 유지한다 / Q2 **불변** |
| R-09 | M1 | `### Attributable diff-check doctrinal switch` | 1,176 | 0 | 같은 `## Parallel Execution` 조항 · Q1 도달 / Q2 **불변** — 기본값 역전의 서술이며 의무 범위를 좁히지 않는다 |
| R-10 | M1 | `### Orchestrator Obligations` + `### Re-delegation Procedure` | 1,026 | 2 | `[HARD] Subagents MUST NOT prompt the user`(`### Subagent Prohibitions`) · Q1 **도달** — 금지는 서브에이전트 측이고 이 둘은 오케스트레이터 측 절차다 / Q2 **불변** — 금지의 대상이 넓어지지 않는다 |
| R-11 | M1 | `## CLAUDE.md Reference` | 128 | 0 | `없음(판단)` — 한 문장 포인터 |
| | | **소계** | **7,108** | | 목적지 이후 **38,315**(여유 1,685) |

`## User Interaction Boundary` L2 머리(132)는 **옮기지 않는다** — 이 파일에서 서브에이전트/오케스트레이터 **비대칭의 범위를 세우는** 자리이고, 떠나면 `### Subagent Prohibitions` 의 `[HARD]` 가 어느 비대칭 안에서 읽혀야 하는지가 always-loaded 표면에서 사라진다(Q1 실패).

#### `core/askuser-protocol.md` → `core/askuser-protocol-reference.md` (15,646 → 21,258, 여유 18,742)

`paths:` 현재 `**/askuser-protocol.md` — **self-keyed, 수정 필요**: `**/askuser-protocol.md,**/.claude/skills/moai/workflows/*.md,**/.claude/output-styles/moai/*.md`. **domain-keyed 근거**: 질문 채널이 필요해지는 작업은 워크플로 실행과 출력 스타일(배너·질문 렌더) 저작이다.

| # | 기제 | 제거되는 절 | 자수 | 인용 | 범위 의존 |
|---|---|---|---:|---:|---|
| R-12 | M1 | `### General Rule for Deferred Tools` | 600 | 0 | `### Mandatory Preload Step` 의 `MUST` · Q1 **도달** — 필수 단계 자체(`ToolSearch(query: "select:AskUserQuestion")`)가 제자리에 남는다 / Q2 **불변** — 일반화 규칙이 떠나도 필수 단계의 대상은 그대로다 |
| R-13 | M1 | `### The Four Triggers` + `### The Five Exceptions` + `### The Unknowns 4-Quadrant Lens` + `### First-Action Sequence After Trigger` | 2,301 | 0 | `### Structural Constraints` 의 `[HARD]`(Socratic 라운드 6제약) · Q1 **도달** — 제약은 라운드의 형태를 구속하고, **언제** 발동하는지는 `## Ambiguity Triggers and Exceptions` L2 머리의 SSOT 선언이 제자리에 남아 포인터로 호명한다 / Q2 **불변** — 트리거와 예외가 **한 묶음으로 함께** 이동하므로 한쪽만 남아 범위가 넓어지는 모양이 생기지 않는다. **이 쌍을 갈라 옮기는 것은 금지**(예외만 떠나면 Q2 실패) |
| R-14 | M1 | `## Blind Spot Pass` | 488 | 6 | `없음(판단)` — 선택적 기법이고 어떤 조항도 이 절을 범위로 삼지 않는다 |
| R-15 | M1 | `### Directive and Recovery` | 861 | 0 | `## Non-ASCII Tool-Call Encoding` 머리의 `[HARD]`(native UTF-8 의무 + `\uXXXX` 금지) · Q1 **도달** — 의무문과 금지문이 L2 머리에 있고 이 절은 실행·복구 절차다 / Q2 **불변** |
| R-16 | M1 | `### Pre-Emit Self-Check (…non-ASCII…) — 3 items` | 713 | 0 | 같은 `[HARD]` · Q1 도달 / Q2 불변 — 체크리스트는 의무의 재확인이며 한정어가 아니다 |
| R-17 | M1 | `### Pre-emit self-check (report-before-ask) — 5 items` | 649 | 0 | `## Report-Before-Ask Gate` 머리 + `### Requested-Deliverable Primacy` + `### Report Completeness Criteria` + `### Report-Promise Fulfillment` 의 `[HARD]` 4건 · Q1 **도달** — 게이트 본문과 완전성 기준이 전부 제자리에 남는다 / Q2 **불변** |
| | | **소계** | **5,612** | | 목적지 이후 **21,258**(여유 18,742) |

#### `workflow/cross-session-messaging.md` → `workflow/cross-session-messaging-detail.md` (15,221 → 22,062, 여유 17,938)

`paths:` 현재 `**/cross-session-messaging*.md,**/kanban-dispatch*.md` — 이미 domain-keyed(kanban 레인 운용이 이 채널을 쓴다). **수정 불필요.**

| # | 기제 | 제거되는 절 | 자수 | 인용 | 범위 의존 |
|---|---|---|---:|---:|---|
| R-18 | M1 | `## What the channel is` — 여는 서술과 세 성질의 **해설부만**(세 성질의 단정문은 제자리 유지) | 900 (계획; 절 전체 1,689) | 0 | `## Rules` 의 `[HARD]` 4건(사용자 결정 우회 금지 · 권한 우회 금지 · 사실만 전송 · 정지된 팀메이트 호명 금지) · Q1 **도달** — 조항이 쓰는 용어(channel / peer / message)의 정의 단정문을 stub 에 남기고 해설만 옮긴다 / Q2 **불변**. **절 전체 이동은 금지** — 세 성질은 규칙의 범위를 정하는 한정어이므로 함께 떠나면 Q2 실패 |
| R-19 | M1 | `## Availability constraints` | 951 | 17 | 같은 `[HARD]` 4건 · Q1 **도달** / Q2 **불변** — 채널의 **부재** 조건을 서술하며, 부재하면 규칙이 구속할 대상 자체가 없다. 규칙의 폭을 좁히는 한정어가 아니다 |
| R-20 | M1 | `## Where it sits among MoAI's existing mechanisms` | 691 | 2 | `없음(판단)` — 기제 비교표 |
| R-21 | M1 | `## Integration with the concurrency checks` | 1,034 | 0 | `없음(판단)` — `agent-common-protocol.md` 쪽 검사와의 합성 절차이며 이 파일의 조항을 한정하지 않는다 |
| R-22 | M1 | `## Addressing and configuration` | 1,057 | 0 | `## Rules` 의 `[HARD]`(정지된 팀메이트 호명 금지) · Q1 **도달** — 주소 지정 상세가 떠나도 "이름으로 호명하지 말라"의 대상이 달라지지 않는다 / Q2 **불변** |
| R-23 | M1 | `## Anti-patterns` | 1,036 | 0 | `## Rules` 의 `[HARD]` 4건 전부 · Q1 **도달** — 안티패턴은 조항 위반 사례의 열거이고 조항 본문이 제자리다 / Q2 **불변**. REQ-ALD2-013 확인: **의무의 사본이 아니라 위반 예시**이므로 중복 제거 금지에 걸리지 않는다 |
| R-24 | M1 | `## Codex broker path (session messaging tools)` | 672 | 0 | `없음(판단)` — 도구 표면 포인터 |
| R-25 | M1 | `## Cross-references` 5개 중 3개 | 500 (계획; 절 전체 727) | 0 | `없음(판단)` — 남기는 둘은 `askuser-protocol.md`(질문 채널 독점)와 `worktree-integration.md`(쓰기 충돌의 구조적 해소) |
| | | **소계** | **6,841** | | 목적지 이후 **22,062**(여유 17,938) |

#### `workflow/context-window-management.md` → `workflow/context-window-management-detail.md` (7,486 → 14,325, 여유 25,675)

`paths:` 현재 `**/context-window-management*.md,**/session-handoff*.md,**/internal/statusline/**` — 이미 domain-keyed. **수정 불필요.**

| # | 기제 | 제거되는 절 | 자수 | 인용 | 범위 의존 |
|---|---|---|---:|---:|---|
| R-26 | M1 | `## Why This Matters` | 640 | 0 | `없음(판단)` — 근거 서술. `## Context Window Targets` 의 `[HARD]` 는 표 자체에 서고 이 절을 참조하지 않는다 |
| R-27 | M1 | `### GLM-5.3 context window (Issue #653)` | 891 | 0 | `## Context Window Targets` 의 `[HARD]`(모델별 임계 표) · Q1 **도달** — GLM 행이 표 안에 남고 이 절은 그 행의 근거·상류 오보고 해설이다 / Q2 **불변** |
| R-28 | M1 | `## Reduction Ladder — cheaper moves before /clear` | 1,802 | 5 | `## User Responsibilities` · `## Orchestrator Responsibilities` 의 `[HARD]` 6건 · Q1 **도달** — 임계 도달 시 `/clear` 의무가 제자리다 / Q2 **불변** — 사다리는 `/clear` 를 선택지로 **늘리지** 않는다. 절 자신의 문장("임계에서의 `/clear` 는 여전히 필수")을 **stub 포인터 줄에 남긴다**; 남기지 않으면 Q2 실패 |
| R-29 | M1 | `### Multi-session work: resume rather than re-establish` | 501 | 0 | `없음(판단)` |
| R-30 | M1 | `## Detection Heuristics` | 2,387 | 16 | `[HARD]`(오케스트레이터 사전 공지 4단계) · Q1 **도달** — 공지 의무는 "임계에 접근하면"이고 임계는 표가 정한다. 추정 방법은 의무의 대상이 아니다 / Q2 **불변** |
| R-31 | M1 | `## Applies To` | 118 | 0 | `없음(판단)` |
| R-32 | M1 | `## Cross-references` 4개 중 2개 | 500 (계획; 절 전체 861) | 0 | `없음(판단)` — 남기는 둘은 `cache-aware-execution.md`·`session-handoff.md` |
| | | **소계** | **6,839** | | 목적지 이후 **14,325**. 목표 6,500 을 **절 단위만으로 충족**(+339) |

#### `workflow/goal-directive.md` → `workflow/goal-directive-detail.md` (20,283 → 24,798, 여유 15,202)

`paths:` 현재 4키 domain-keyed(AC-ALD2-006 의 유일 선례). **수정 불필요.**

| # | 기제 | 제거되는 절 | 자수 | 인용 | 범위 의존 |
|---|---|---|---:|---:|---|
| R-33 | M1 | `## What It Is` — 상태 파일·상한·정지 가드 해설부 | 1,400 (계획; 절 전체 1,969) | 1 | `없음(판단)` — 이 파일의 구속 조항 줄은 **0개**다(§1 실측). 외부 의존은 R-34 가 진다 |
| R-34 | M1 | `## Goal-Presentation Timing` | 1,650 | 17 | **외부 파일의 구속 조항이 지배받는다** — `session-handoff.md § Canonical Format` Block 5 의 `[HARD]`(「`/moai goal` 은 arm-only 이므로 Block 5 의 단독 행이 되지 않는다」)가 이 절의 arm-only 성질에 의존한다 · Q1 **도달** — 그 조항이 이 절을 **이름으로** 교차참조하고 stub 포인터가 절 이름을 유지한다 / Q2 **불변** — arm-only 성질이 companion 으로 가도 조항이 구속하는 대상(Block 5 한 줄)은 그대로다. 인용 17건 중 1건이 초과 파일 내부(§5 선례) |
| R-35 | M1 | `## Proactive Recommendation Triggers` | 965 | 1 | `없음(판단)` — T1~T4 는 권고 트리거이며 어떤 조항의 한정어도 아니다 |
| R-36 | M1 | `## Cross-references` 4개 중 2개 | 500 (계획; 절 전체 748) | 0 | `없음(판단)` |
| | | **소계** | **4,515** | | 목적지 이후 **24,798**. 목표 4,500 충족(+15) |

#### `core/moai-mcp-tools.md` → `core/moai-mcp-tools-catalogue.md` (11,424 → 15,491, 여유 24,509)

`paths:` 현재 3키 domain-keyed. **수정 불필요.**

| # | 기제 | 제거되는 절 | 자수 | 인용 | 범위 의존 |
|---|---|---|---:|---:|---|
| R-37 | M1 | `## Tool families (35 of the 39 tools…)` | 2,206 | 0 | `## The project_root input` 의 `[HARD]`(워크트리 세션은 반드시 전달) · Q1 **도달** — 그 조항은 13개 도구를 **자기 절 안에서** 열거한다 / Q2 **불변** |
| R-38 | M1 | `### Session messaging broker (Claude ↔ Codex)` | 1,361 | 0 | 같은 `[HARD]` · Q1 도달 / Q2 불변 — 브로커 4도구는 `project_root` 를 받지 않는다 |
| R-39 | M1 | `## Unwired-by-design` | 500 | 1 | `없음(판단)` — `goal_arm` 미배선의 근거 |
| | | **소계** | **4,067** | | 목적지 이후 **15,491**. 목표 5,500 대비 **−1,433 → M2**(D-30) |

`## MCP-over-CLI rule`(462)은 옮기지 않는다 — 이 파일의 유일한 선택 규칙이고, 떠나면 도구 목록만 남아 "언제 MCP 를 고르는가"가 사라진다.

#### `workflow/main-checkout-branch-guard.md` → `workflow/main-checkout-branch-guard-detail.md` (9,984 → 13,612, 여유 26,388)

`paths:` 현재 3키 domain-keyed. **수정 불필요.**

| # | 기제 | 제거되는 절 | 자수 | 인용 | 범위 의존 |
|---|---|---|---:|---:|---|
| R-40 | M1 | `## Why This Matters` | 415 | 0 | `없음(판단)` — 경쟁이 조용한 이유의 서술. 이 파일의 범위는 L1 서문("primary 체크아웃은 공유된다")이 세우고 그 서문은 남는다 |
| R-41 | M1 | `## Procedure — Isolate With a Worktree` | 574 | 0 | `## Rules` 의 `[HARD]`(primary 체크아웃에서 브랜치 상태 변경 금지) · Q1 **도달** — 금지가 제자리에 있고 이 절은 대안 절차다 / Q2 **불변** |
| R-42 | M1 | `## Verification` | 283 | 0 | `## Staleness Rule` 의 `[HARD]` · Q1 **도달** — 재확인 의무와 두 명령이 그 절 안에 있다 / Q2 **불변** |
| R-43 | M1 | `## Mechanical Enforcement` | 1,856 | 12 | `## Rules` 의 `[HARD]` · Q1 **도달** — 기계 층은 독트린의 **조건부 적용**을 서술하며 금지의 대상을 정하지 않는다 / Q2 **불변**. 「면제는 도구 spawn 서브에이전트에서 도달 불가」한 줄은 stub 포인터에 남긴다 |
| R-44 | M1 | `## Cross-references` 5개 중 2개 | 500 (계획; 절 전체 817) | 0 | `없음(판단)` |
| | | **소계** | **3,628** | | 목적지 이후 **13,612**. 목표 5,000 대비 **−1,372 → M2**(D-31) |

#### `core/native-idiom-and-register.md` → `core/native-idiom-and-register-detail.md` (2,509 → 4,729, 여유 35,271)

`paths:` 현재 2키 domain-keyed. **수정 불필요.**

| # | 기제 | 제거되는 절 | 자수 | 인용 | 범위 의존 |
|---|---|---|---:|---:|---|
| R-45 | M1 | `## Two registers — do not conflate` | 715 | 0 | `## The Invariant` 의 `[HARD]` + `## Mechanism — when to invoke humanize` 의 `[HARD]` · Q1 **도달** — 두 조항이 "chat 은 구어체, artifact 는 문어체"를 자기 문장 안에서 이미 말한다 / Q2 **불변** |
| R-46 | M1 | `## Why calques survive, and what they look like` | 746 | 0 | 같은 두 조항 · Q1 도달 / Q2 불변 — calque 예시는 금지의 사례이지 한정어가 아니다 |
| R-47 | M1 | `## Pre-emit self-check (non-English output only)` | 459 | 0 | 같은 두 조항 · Q1 도달 / Q2 불변 |
| R-48 | M1 | `## Cross-references` 3개 중 2개 | 300 (계획; 절 전체 516) | 0 | `없음(판단)` |
| | | **소계** | **2,220** | | 목적지 이후 **4,729**. 목표 2,000 충족(+220) |

#### `workflow/cache-aware-execution.md` → `workflow/cache-aware-execution-reference.md` (5,841 → 7,630, 여유 32,370)

`paths:` 현재 `**/cache-aware-execution.md` — **self-keyed, 수정 필요**: `**/cache-aware-execution.md,**/.claude/agents/moai/*.md,**/.claude/rules/moai/workflow/*.md`. **domain-keyed 근거**: 캐시 경제가 문제가 되는 작업은 spawn 순서 설계(에이전트 정의)와 always-loaded 룰 편집 시점 결정이다.

| # | 기제 | 제거되는 절 | 자수 | 인용 | 범위 의존 |
|---|---|---|---:|---:|---|
| R-49 | M1 | L1 서문의 캐시 기제 해설 문단(접두 일치·TTL·1시간 설정 서술) | 800 (계획; 서문 전체 1,233) | — | 지시 6·7·8·9 의 `[HARD]` 4건 · Q1 **도달** — 각 지시가 자기 안에서 의무를 완결한다 / Q2 **불변**, 단 **조건부**: 서문의 마지막 문장(「이 규칙들은 언제·어떤 순서로 행동할지를 지배하며 게이트 의미를 바꾸지 않는다」)을 **남기는 경우에만** 성립한다. 그것이 이 파일 전체의 적용 범위를 세우는 문장이고, 떠나면 10개 지시가 게이트 의미까지 구속하는 것으로 읽힌다 |
| R-50 | M1 | `## Non-goals` | 489 | 0 | 같은 4건 · Q1 도달 / Q2 **주의** — 이 절은 "게이트 의미를 바꾸지 않는다"를 **반복**하는 자리다. R-49 가 서문의 같은 문장을 남기므로 이동이 한정어를 없애지 않는다. 서문 문장을 남기지 않으면 **이 행도 금지** |
| R-51 | M1 | `## Cross-references` 5개 중 2개 | 500 (계획; 절 전체 781) | 0 | `없음(판단)` |
| | | **소계** | **1,789** | | 목적지 이후 **7,630**. 목표 4,000 대비 **−2,211 → M2**(D-28) |

#### `workflow/skill-routing.md` → `workflow/skill-routing-detail.md` (1,351 → 2,598, 여유 37,402)

`paths:` 현재 3키 domain-keyed. **수정 불필요.** 이 파일의 **템플릿 사본** frontmatter 부재는 M6 이 별도로 고친다(REQ-ALD2-008) — 그것은 `paths:` 의 **키 구성** 문제가 아니라 **부재** 문제이므로 조건 3 과 다른 축이다.

| # | 기제 | 제거되는 절 | 자수 | 인용 | 범위 의존 |
|---|---|---|---:|---:|---|
| R-52 | M1 | `## 2. Agent Obligation` | 388 | 0 | §1 + §1.1 의 `[HARD]` 5건 · Q1 **도달** — 오케스트레이터 의무가 제자리에 있고 이 절은 에이전트 측 로딩 관행이다 / Q2 **불변** |
| R-53 | M1 | `## 3. Rationale` | 459 | 0 | 같은 5건 · Q1 도달 / Q2 불변 — 비용 프로필 근거 |
| R-54 | M1 | `## 4. Cross-references` 4개 중 2개 | 400 (계획; 절 전체 624) | 0 | `없음(판단)` |
| | | **소계** | **1,247** | | 목적지 이후 **2,598**. 목표 2,200 대비 **−953 → M2**(D-32) |

#### `workflow/kanban-dispatch.md` → **신규 companion** (0 → 5,580)

§2.2 배정대로 신규다 — 기존 `kanban-dispatch-detail.md` 는 41,036 으로 이미 초과이고 REQ-ALD2-015 가 목적지 사용을 금지한다. 파일명 제안 `workflow/kanban-dispatch-mechanics.md`(기존 `-detail` 과 역할이 갈린다: detail = 서사·사례·사고 기록, mechanics = 보드·렌즈·모드 기제). `paths:` 신규 선언 `**/kanban-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.moai/state/integration/**`. **domain-keyed 근거**: 이 companion 이 필요해지는 작업은 리드 세션의 배차와 통합 창 운용이며, 그 작업이 건드리는 경로가 이 셋이다.

| # | 기제 | 제거되는 절 | 자수 | 인용 | 범위 의존 |
|---|---|---|---:|---:|---|
| R-55 | M1 | `## The board` | 446 | 7 | `## Entry into the board is an operator act` 의 `[HARD]` 7건 · Q1 **도달** — 컬럼 이름이 조항 본문 안에 쓰이고 stub 포인터가 5컬럼 열거를 호명한다 / Q2 **불변** |
| R-56 | M1 | `## The dispatch cycle` L2 머리 | 777 | 11 | `### The delegation channel is the queue` 2건 + `### Dispatch format` 2건의 `[HARD]` · Q1 **도달** — 하위 절들이 제자리에 남는다 / Q2 **불변** |
| R-57 | M1 | `## Review lens selection` | 438 | 4 | `없음(판단)` — 렌즈 표는 이미 companion 에 있고 이 절은 포인터다. `--deep --patch` 가 두 번 opt-in 이라는 문장은 stub 포인터에 남긴다 |
| R-58 | M1 | `### Serializing a heavy run across lanes` | 1,060 | 0 | `## Verification load is lane-local` 의 `[HARD]` 2건(카드 범위 검증 · 백그라운드 부하 금지) · Q1 **도달** / Q2 **불변** — 임대 절차는 금지를 좁히지 않는다 |
| R-59 | M1 | `## Factory Mode — the card travels whole` — 기제 서술부(**레인 spawn 권한 문단은 제자리 유지**) | 1,200 (계획; 절 전체 1,793) | 1 | 레인 spawn 권한 문단이 **실질 의무**(카드 단계별 전문가 spawn 권한, depth-1 한정)이므로 함께 옮기지 않는다 · Q1/Q2 는 그 문단을 남기는 조건에서만 성립. 남기지 않으면 kanban·factory 컴패니언 세션의 권한 근거가 always-loaded 에서 사라진다 |
| R-60 | M1 | `## Boundaries — what this protocol does not do` — 5항목 중 3항목 | 600 (계획; 절 전체 1,059) | 0 | 파일 내 `[HARD]` 전부 · Q1 **도달** / Q2 **주의** — 이 절은 "이 프로토콜이 하지 않는 것"을 열거해 **범위를 좁힌다.** `게이트 우회 없음`·`질문 위임 없음` 두 항목은 **stub 에 남기고**, 나머지 셋(보드 상태 저장소 없음 · 세션 spawn 없음 · 빈 역할은 결함)만 옮긴다. **전량 이동은 Q2 실패** |
| R-61 | M1 | `## Cross-references` 6개 중 4개 | 600 (계획; 절 전체 842) | 0 | `없음(판단)` |
| | | **소계** | **5,121** | | 신규 companion **5,121**(한도 아래). 목표 16,000 대비 **−10,879 → M2**(D-22) |

`## Scope — when this rule is live`(1,309)는 **옮기지 않는다** — iter4 N7 이 확인한 원거리 지배 사례다. 이 절이 떠나면 `[HARD] Never spawn background load` 가 kanban 레인 한정이던 것이 **매 세션 구속**으로 읽힌다(Q2 실패). 이 파일의 **모든** 구속 조항이 이 절에 지배받으므로 표에 올리지 않는다. §4.3 재수록.

#### `workflow/session-handoff.md` → **신규 companion** (0 → 5,442)

기존 `session-handoff-examples.md` 는 41,616 으로 초과 — 목적지 금지(REQ-ALD2-015). 파일명 제안 `workflow/session-handoff-format.md`(examples = 예시·부록·안티패턴 카탈로그, format = 마커 규격·로케일 표·활성화 행렬). `paths:` 신규 `**/session-handoff*.md,**/.claude/output-styles/moai/*.md,**/.moai/state/handoff/**`. **domain-keyed 근거**: 이 companion 이 필요해지는 작업은 핸드오프 블록 렌더와 출력 스타일 §8 저작이다.

| # | 기제 | 제거되는 절 | 자수 | 인용 | 범위 의존 |
|---|---|---|---:|---:|---|
| R-62 | M1 | `## Why This Matters` | 318 | 0 | `없음(판단)` |
| R-63 | M1 | `### Cut-line Marker Specification` | 589 | 21 (초과 파일 내 2) | `## Canonical Format` 의 `[HARD]`(6블록 구조 + 컷라인 경계) · Q1 **도달** — 조항이 이 절을 이름으로 참조하고 stub 포인터가 유지한다 / Q2 **불변** |
| R-64 | M1 | `### Localization Table` | 1,621 | 18 (초과 파일 내 1) | 같은 `[HARD]` + `## Auto-Memory Integration` 의 `[HARD]` · Q1 **도달** — 두 조항이 "locale renderings per § Localization Table" 로 이름 참조한다 / Q2 **불변** |
| R-65 | M1 | `## Paste-Time Activation Matrix` | 549 | 5 (초과 파일 내 1) | `없음(판단)` — 이미 companion 을 가리키는 요약 절 |
| R-66 | M1 | `## Auto-Injected Resume Flow (mode=auto)` L2 머리 | 845 | 19 (초과 파일 내 2) | `## Output Surface` 의 `[HARD]` + `## When To Generate` 의 `[HARD]` · Q1 **도달** — 하위 `### Invariants (both modes)` 를 **남긴다**(Kickoff 게이트 불변 · manual 회귀 동일 · fail-open) / Q2 **불변**. **Invariants 를 함께 옮기면 Q2 실패** |
| R-67 | M1 | `### Pre-emit self-check (emission surface) — 3 items` | 406 | 0 | `## Output Surface (User-Facing)` 의 `[HARD]` · Q1 도달 / Q2 불변 |
| R-68 | M1 | `## Anti-Patterns` | 214 | 0 | `없음(판단)` — 포인터 한 줄 |
| R-69 | M1 | `## Cross-references` 9개 중 5개 | 900 (계획; 절 전체 1,549) | 0 | `없음(판단)` — 남기는 넷은 `context-window-management.md`(임계 SSOT)·`goal-directive.md`·`moai-constitution.md`·`handoff.yaml` |
| | | **소계** | **5,442** | | 신규 companion **5,442**. 목표 10,000 대비 **−4,558 → M2**(D-27) |

#### `core/moai-constitution.md` → `core/moai-constitution-detail.md` (7,644 → 9,520, 여유 30,480)

`paths:` 현재 3키 domain-keyed. **수정 불필요.**

| # | 기제 | 제거되는 절 | 자수 | 인용 | 범위 의존 |
|---|---|---|---:|---:|---|
| R-70 | M1 | `## Parallel Execution` 중 세 원시 기제(sub-agents / Agent Teams / dynamic workflows) 해설부 | 1,000 (계획; 절 전체 1,638) | 0 | `## MoAI Orchestrator` 의 `[HARD]`(AskUserQuestion 독점) + `## Agent Core Behaviors` 의 `[HARD]` 7건 · Q1 **도달** / Q2 **불변**, 단 **조건부**: fanout 상한 문장과 "독립 호출은 병렬로" 규칙 불릿을 **남기는 경우에만** 성립한다. 그것이 이 절의 규범부다 |
| R-71 | M1 | `## Tool Selection Priority` | 380 | 6 | `없음(판단)` — 본문이 이미 "정본 표는 `agent-common-protocol.md` 에 있다"고 적는 포인터다. **주의**: R-04 가 그 정본 표를 companion 으로 옮기므로 두 이동이 연쇄한다 — 포인터의 대상 경로를 **같은 커밋에서** 갱신한다 |
| R-72 | M1 | `## URL Verification` 중 실행 절차 3단계 서술 | 496 → 실이동 **300** (계획) | 1 | `없음(판단)` — 절차 서술은 `glm-web-tooling.md` 와 `CLAUDE.md §10` 의 재서술이다. **다만 "검증 없는 URL 생성 금지"는 이 파일과 `CLAUDE.md` 어디에도 `[HARD]` 로 없어** REQ-ALD2-013 의 취지대로 **always-loaded 에 보존**한다 — 실제 이동은 절차 서술뿐 |
| | | **소계** | **1,680** | | 목적지 이후 **9,324**. 목표 8,500 대비 **−6,820 → M2**(D-24). **16파일 중 가장 빡빡** |

**M1 소계 합**: 7,108 + 5,612 + 6,841 + 6,839 + 4,515 + 4,067 + 3,628 + 2,220 + 1,789 + 1,247 + 5,121 + 5,442 + 1,680 = **56,109**

---

### §4.2 삭제 행 — M1′(중복 제거)과 M2(제자리 압축)

목적지·`paths:`·domain-keyed 근거 세 칸은 `삭제(목적지 없음)` / `해당없음` 이다(§4.5 조건 0). **범위 의존 칸은 기제와 무관하게 채운다.**

#### `CLAUDE.md` — M1′ (중복 제거, 목표 8,000)

`CLAUDE.md` 는 companion 을 만들지 않는다 — §1~§17 의 비구속 산문 대부분이 path-scoped 룰에 정본으로 있는 내용의 **재서술**이므로, 옮기지 않고 지우고 정본을 가리킨다. 아래 자수는 **계획 삭제량**이며 절 전체 자수를 함께 적는다.

| # | 기제 | 제거되는 절 | 삭제량 (절 전체) | 정본 위치 | 범위 의존 |
|---|---|---|---:|---|---|
| D-01 | M1′ | `## 15. Agent Teams (Re-allowed, experimental) + CG Mode` | 1,500 (1,603) | `orchestration-mode-selection.md §C.1` · `glm-web-tooling.md` | `없음(판단)` — 이 절에 `[HARD]` 없음. §14 의 `[HARD]`(백그라운드 기본·동시성 안전장치)는 자기 절 안에서 완결된다 · Q2 **불변**. **주의**: 라이브 사본의 CG 서술이 낡았다는 사실은 이 카드가 고치지 않는다(§B-1) — 삭제는 낡은 주장을 **없애는** 방향이므로 범위 밖 수정이 아니지만, 정정 문구를 새로 쓰지는 않는다 |
| D-02 | M1′ | `## 16. Context Search Protocol` | 1,100 (1,252) | `context-window-management.md` · `session-handoff.md`(절 자신이 그렇게 적는다) | `없음(판단)` · Q2 불변 |
| D-03 | M1′ | `## 10. Web Search Protocol` | 700 (805) | `moai-constitution.md § URL Verification` · `glm-web-tooling.md` · `dynamic-workflows.md` | `없음(판단)`. **REQ-ALD2-013 확인**: "WebSearch 결과에 없는 URL 을 생성하지 말라"는 의무 문장은 **남긴다** — 정본이 path-scoped 룰이라는 이유로 지우면 그 룰이 로드되지 않는 턴에 의무가 사라진다 |
| D-04 | M1′ | `## 11. Error Handling` | 550 (653) | `agent-common-protocol.md § Error Recovery Pattern` · `archived-agent-rejection.md` · `session-handoff.md` | `없음(판단)` · Q2 불변 |
| D-05 | M1′ | `## 12. MCP Servers & Deep Analysis Modes` | 550 (650) | `dynamic-workflows.md` · `settings-management.md` · `moai-foundation-thinking` 스킬 | `없음(판단)` · Q2 불변 |
| D-06 | M1′ | `## 13. Progressive Disclosure System` | 330 (400) | `skill-authoring.md § Progressive Disclosure`(절 자신이 "Canonical rule" 로 적는다) | `없음(판단)` · Q2 불변 |
| D-07 | M1′ | `## 2. Request Processing Pipeline` — 5단계 해설부 | 700 (1,359) | `.claude/skills/moai/SKILL.md`(Intent Router) · `orchestration-mode-selection.md` | `없음(판단)` · Q2 **불변**, 단 **조건부**: 5단계의 **순서 열거**(①~⑤)는 남기고 각 단계의 해설만 지운다. ④의 Kickoff 게이트 언급은 §8 의 `[HARD]` 와 짝이므로 남긴다 |
| D-08 | M1′ | `## 4. Agent Catalog` L2 머리의 nesting 주석 | 600 (781) | `agent-authoring.md` · `agent-patterns.md`(주석 자신이 그렇게 적는다) | `없음(판단)`. `### Selection Decision Tree`(b=1)는 **건드리지 않는다** |
| D-09 | M1′ | `## 5. SPEC-Based Workflow` — 방법론·MX 해설부 | 450 (709) | `spec-workflow.md` · `mx-tag-protocol.md` | `없음(판단)` · Q2 불변. 커맨드 흐름 3줄은 남긴다 |
| D-10 | M1′ | `## 9. Configuration Reference` — 룰 카테고리·설계 시스템 해설부 | 500 (734) | `moai-memory.md` · 각 config 섹션 | `없음(판단)` · Q2 불변. **`@.moai/config/sections/user.yaml` / `@…/language.yaml` 두 import 줄은 절대 지우지 않는다** — 그 둘이 18파일 집합의 두 yaml 을 로드하는 유일한 경로다 |
| D-11 | M1′ | `## 3. Command Reference` 하위 서브커맨드 열거 | 300 (468) | `.claude/skills/moai/SKILL.md` | `없음(판단)` · Q2 불변 |
| D-12 | M2 | `## 7. Safe Development Protocol` — Rule 1~5 본문의 **제자리 압축** | 300 (2,845) | (삭제 아님 — 압축) | Rule 1~5 는 `[HARD]` 토큰이 없지만 **실질 의무**다(§1 의 `[HARD] Rules (Mandatory)` 가 이 절을 확장이라고 밝힌다) · Q1 **도달** / Q2 **불변** — 의무·조건·예외·수치를 하나도 지우지 않고 산문만 줄인다(REQ-ALD2-012). **M1′ 로 취급하지 않는다** — 정본이 다른 곳에 있다는 이유로 지우면 REQ-ALD2-013 위반 |
| D-13 | M2 | `## 17. Troubleshooting` 제자리 압축 | 420 (809) | (압축) | `없음(판단)` — 표 2행 + 디버그 명령. 정본이 따로 없으므로 삭제하지 않고 압축한다 |
| | | **소계** | **8,000** | | 삭제 7,280 + 압축 720 |

#### `AGENTS.md` — M2 전용 (목표 4,000, REQ-ALD2-011)

이 파일은 companion 을 만들지 않는다. 절이 사라지지 않고 **절 안의 비구속 산문이 줄어든다.** 따라서 아래 행은 "제거되는 절" 칸에 압축 대상 절을 적고, 삭제량은 계획 배정이다.

| # | 기제 | 압축되는 절 | 배정 삭제량 (절 전체) | 범위 의존 |
|---|---|---|---:|---|
| D-14 | M2 | L1 서문 — 예산 경고 해설 + 능력 바인딩 표의 산문부 | 700 (2,372) | `§1` 의 `MUST NOT`·`§5`·`§6` 의 `MUST` 전부 · Q1 **도달** / Q2 **주의** — 서문은 "이 파일은 자기충족적이며 다른 지시 파일을 전제하지 않는다"로 **이 파일 전체의 적용 범위를 세운다.** 그 선언 문장과 능력 바인딩 표 자체는 **남기고** 해설만 줄인다. 선언을 지우면 모든 조항이 다른 하네스에서 어떻게 읽혀야 하는지가 사라진다(Q2 실패) |
| D-15 | M2 | `## 2. Git, branches, and the shared checkout` | 700 (2,305) | 이 절의 금지 목록은 `[HARD]` 토큰 없이 실질 의무다 · Q1 **도달** / Q2 **불변** — 금지 표와 재확인 명령은 그대로 두고 근거 산문만 줄인다 |
| D-16 | M2 | `## 3. Worktrees` | 700 (2,080) | 같음 · Q1 도달 / Q2 **주의** — "카드 브랜치는 미푸시이므로 워크트리가 유일본"이라는 문장이 폐기 금지의 **이유이자 범위**다. 남긴다 |
| D-17 | M2 | `## 4. How verification is run` | 400 (1,493) | 같음 · Q1 도달 / Q2 불변 — CodeRabbit 두 조건은 수치·조건이라 REQ-ALD2-012 가 보존을 요구한다 |
| D-18 | M2 | `## 5. Core behaviors` | 600 (2,606) | 이 절의 `[HARD]`(간이화 안전 예외) 1건 · Q1 **도달** / Q2 **불변** — 6개 행동의 규범 문장과 사다리 7단계는 보존하고 예시 산문만 줄인다 |
| D-19 | M2 | `## 6. Output, language, and format` | 400 (1,588) | 이 절의 `MUST` 2건 · Q1 **도달** / Q2 **불변** |
| D-20 | M2 | `## 7. Tools and command output` | 300 (1,678) | `없음(판단)` — 이 절에 `[HARD]` 없음 · Q2 불변 |
| D-21 | M2 | `## 8. Harness-local instructions` | 200 (933) | `없음(판단)` · Q2 불변 |
| | | **소계** | **4,000** | 전량 제자리 압축, 포인터 재유입 **0** |

#### 나머지 파일의 M2 보충 — 절 단위 풀이 목표에 못 미치는 분

각 행의 압축 대상은 **구속 조항 줄을 담은 절**이며, 구속 조항 줄은 축자 동결된다(REQ-ALD2-003). 압축 대상은 그 절 안의 비구속 문단뿐이다.

| # | 파일 | 압축 대상 절(들) | 배정 삭제량 | 범위 의존 |
|---|---|---|---:|---|
| D-22 | `kanban-dispatch.md` | `## Entry into the board is an operator act`(4,764/b=7) · `## Isolation is provisioned by MoAI…`(6,511/b=9) · `## Integration into the release branch is self-served`(4,207/b=4) · `## Deputy dispatch surface`(1,907/b=5) · `## Completion is read, never trusted`(1,359/b=1) · `## Card classes`(2,032/b=1) | **10,879** | 각 절의 `[HARD]` 전부 · Q1 **도달**(조항 줄 불변) / Q2 **불변** — 압축은 근거·서사·사고 기록만 줄이고 한정어 문장(`## Scope` 참조, 레인 한정 표현)은 보존한다. **이 파일은 목표의 68%가 M2** 이므로 REQ-ALD2-012 판정 부담이 가장 크다 |
| D-23 | `agent-common-protocol.md` | `### Pre-Spawn Sync Check`(3,456/b=4) · `### Pre-Edit Sync Check`(3,013/b=2) · `## Background Agent Execution`(1,922/b=2) · `### Hook Invocation Surface`(2,317/b=1) · `### Verbatim batch…`(2,383/b=1) · `### Per-Spawn Model Injection`(1,231/b=1) | **7,892** | 각 절의 `[HARD]`/`MUST` · Q1 도달 / Q2 **불변** — 판정 행렬(divergence 4행, 세션 2행)과 명령은 수치·조건이라 보존 대상이다 |
| D-24 | `moai-constitution.md` | `## Opus 5.5 Prompt Philosophy`(2,030/b=2) · `## Lessons Protocol`(2,757/b=2) · `## Agent Core Behaviors` 6개 하위(총 5,419/b=7) · `## Quality Gates`(677) · `## Worktree Isolation`(502) | **6,820** | 각 절의 `[HARD]` · Q1 도달 / Q2 **불변**. `## Quality Gates`·`## Worktree Isolation` 은 **M1 으로 옮기지 않고 압축한다** — TRUST 5 항목과 worktree 경로 규칙은 실질 의무이고 REQ-ALD2-013 이 정본 중복을 이유로 한 삭제를 금지한다 |
| D-25 | `verification-claim-integrity.md` | 전 절(`§1`·`§1.1`·`§2`·`§2.1`·`§2.2`·`§2.3`·`§3`·`§3.1` — 모두 구속 조항 줄 보유) | **6,000** | 각 절의 `[HARD]` 17건 · Q1 도달 / Q2 **불변**. **이 파일은 이동 가능한 절이 0개**이므로 목표 전량이 M2 다 — 판정은 REQ-ALD2-012(의미 보존) 검토 몫이고 기계 검사가 없다(§D.4) |
| D-26 | `askuser-protocol.md` | `## Recommendation Placement Principles`(1,558, 이동 금지분) · `### Recommendation mode`(1,499/b=2) · `## Report-Before-Ask Gate` 하위(2,795/b=4) · `### Completion-Report Next-Step Discipline`(1,604/b=1) | **5,388** | 각 절의 `[HARD]` · Q1 도달 / Q2 불변 — 5원칙의 **번호와 제목**은 보존한다(`### Recommendation mode` 가 원칙 5 를 이름으로 지목한다) |
| D-27 | `session-handoff.md` | `### Emission-Time Save Obligation`(1,914/b=2) · `### Field-by-Field Specification`(2,891/b=2) · `## Auto-Memory Integration`(1,697/b=1) · `## Output Surface`(1,703/b=1) · `## When To Generate`(1,244/b=1) | **4,558** | 각 절의 `[HARD]` · Q1 도달 / Q2 불변 — 5트리거 표와 6블록 스켈레톤은 수치·구조라 보존 |
| D-28 | `cache-aware-execution.md` | `## Directives`(4,401/b=5) — 지시 1~10 의 근거 산문 | **2,211** | 지시 6~10 의 `[HARD]` 5건 · Q1 도달 / Q2 불변 — 지시 문장 자체는 동결, 근거만 줄인다 |
| D-29 | `cross-session-messaging.md` | `## Rules`(4,782/b=4) · `## A send result has three shapes…`(2,054/b=1) | **2,159** | 각 절의 `[HARD]` · Q1 도달 / Q2 불변 — 3-shape 표는 판정 행렬이라 보존 |
| D-30 | `moai-mcp-tools.md` | `## The project_root input`(2,368/b=1) | **1,433** | 그 절의 `[HARD]` · Q1 도달 / Q2 불변 — 3행 판정 표와 거절 사유는 보존 |
| D-31 | `main-checkout-branch-guard.md` | `## Rules`(1,289/b=1) · `## Staleness Rule`(527/b=1) · `## Detecting Concurrent Sessions`(520/b=1) | **1,372** | 각 절의 `[HARD]` · Q1 도달 / Q2 불변 — 금지 표 6행과 두 명령은 보존 |
| D-32 | `skill-routing.md` | `## 1. Orchestrator Obligation`(891/b=1) · `### §1.1`(3,000/b=4) — 예시 3줄·안티패턴 산문 | **953** | 각 절의 `[HARD]` 5건 · Q1 도달 / Q2 불변 — 6행 매핑 표와 `report.format` 결합 규칙은 보존 |
| | | **소계** | **49,665** | 포인터 재유입 **0**(목적지가 없다) |

---

### §4.3 기각 — 옮기지 않기로 판단한 절과 그 이유

조건 1(토큰)은 통과하지만 **지배 관계 판단에서 걸린** 절들이다. `plan.md §C` M1 조건 1 단서 (a)/(b)에 해당한다.

| 파일 | 절 | 자수 | 기각 사유 |
|---|---|---:|---|
| `kanban-dispatch.md` | `## Scope — when this rule is live` | 1,309 | **(b) 원거리 지배** — 이 파일의 모든 구속 조항을 임의 거리에서 지배한다. 떠나면 `[HARD] Never spawn background load` 가 kanban 레인 한정에서 매 세션 구속으로 읽힌다(Q2 실패). iter4 N7 의 확인 사례 |
| `agent-common-protocol.md` | `## User Interaction Boundary` L2 머리 | 132 | **(a)** 비대칭의 범위를 세운다 — 떠나면 `### Subagent Prohibitions` 의 `[HARD]` 가 어느 비대칭 안에서 읽히는지 사라진다(Q1 실패) |
| `askuser-protocol.md` | `## Recommendation Placement Principles` L2 머리 | 1,558 | **(a)** `### Recommendation mode` 의 `[HARD]` 가 원칙 5 를 **이름으로 지목**한다 — 머리가 떠나면 그 조항이 무엇을 일반화하는지 사라진다(Q1 실패). M2 로 압축(D-26) |
| `askuser-protocol.md` | `### Exceptions (gate does not apply)` | 447 | **(b)** 게이트의 `[HARD]` 를 **좁히는** 예외 목록 — 떠나면 게이트가 넓게 읽힌다(Q2 실패, §D.3 방향 2) |
| `askuser-protocol.md` | `## Socratic Interview Structure` L2 머리 | 366 | **(a)** 하위 `### Structural Constraints` 의 6제약이 서는 전제(라운드의 정의) |
| `askuser-protocol.md` | `## Ambiguity Triggers and Exceptions` L2 머리 | 223 | **(a)** SSOT 선언 — R-13 의 포인터가 서는 자리 |
| `askuser-protocol.md` | `## Orchestrator–Subagent Boundary` L2 머리 · `### Blocker Report Format / Re-delegation Procedure` | 291 | **(a)** 비대칭 선언 + 소유 경계 포인터 |
| `moai-constitution.md` | `## Output Format` · `## Error Handling Protocol` · `## Security Boundaries` · `## Quality Gates` · `## Worktree Isolation` | 1,910 | **REQ-ALD2-013** — 정본이 path-scoped 룰에 있다는 이유로 지울 수 없는 실질 의무 블록. M2 로 압축(D-24) |
| `moai-constitution.md` | `## Agent Core Behaviors` L2 머리 | 204 | **(a)** 6개 행동의 범위 선언 |
| `goal-directive.md` | `## Hard Preconditions for Every Recommendation` | 959 | **(b)** Kickoff 선행 의무와 안전 경계 불변을 실질적으로 진술 — 떠나면 arm 이 게이트를 대신하지 않는다는 한정이 사라진다 |
| `moai-mcp-tools.md` | `## MCP-over-CLI rule` | 462 | **(a)** 이 파일의 유일한 선택 규칙 — 떠나면 도구 목록만 남는다 |
| `session-handoff.md` | `### Invariants (both modes)` | 834 | **(b)** Kickoff 게이트 불변·manual 회귀 동일·fail-open 을 진술 — R-66 과 함께 옮기면 Q2 실패 |
| `cache-aware-execution.md` | L1 서문의 마지막 문장 | (문장) | **(b)** 파일 전체의 적용 범위("게이트 의미를 바꾸지 않는다") |
| `kanban-dispatch.md` | `## Boundaries…` 중 `게이트 우회 없음`·`질문 위임 없음` | (2항목) | **(b)** 범위를 좁히는 항목 |
| `kanban-dispatch.md` | `## Factory Mode` 중 레인 spawn 권한 문단 | (문단) | **(b)** 실질 의무(권한 부여와 depth-1 한정) |
| `CLAUDE.md` | `## 0. Standing Contract` · `### HARD Rules (Mandatory)` · `## 1. Core Identity` · `## 6. Quality Gates` · `## 8`·`## 14` | 3,331 | `@AGENTS.md` import(§0) · 의무 요약(§1) · 이미 순수 포인터(§6) · 구속 조항 보유(§8·§14) |
| `CLAUDE.md` | `## 9` 의 두 `@` import 줄 | (2줄) | 18파일 집합의 두 yaml 을 로드하는 유일 경로 |
| `AGENTS.md` | 전 절(M1 자체가 금지) | — | **REQ-ALD2-011** — companion 을 만들지 않는다 |

---

### §4.4 투영 — 마일스톤별 감축량과 **포인터 재유입의 입도 위험**

| 마일스톤 | 대상 | 기제 | 이 표가 내는 gross |
|---|---|---|---:|
| **M2** | `AGENTS.md` + `CLAUDE.md` | M2 4,000 + M1′ 8,000 | **12,000** |
| **M3** | `kanban-dispatch` · `agent-common-protocol` · `askuser-protocol` · `session-handoff` | M1 23,283 + M2 28,717 | **52,000** |
| **M4** | `moai-constitution` · `verification-claim-integrity` · `cross-session-messaging` · `context-window-management` | M1 15,360 + M2 14,979 | **30,339** |
| **M5** | `moai-mcp-tools` · `cache-aware-execution` · `goal-directive` · `main-checkout-branch-guard` · `skill-routing` · `native-idiom` | M1 17,466 + M2 5,969 | **23,435** |
| **M6.5** | 신규 companion 2개 생성(kanban 5,121 · session-handoff 5,442) | (M3 의 목적지) | 0 (이중 계산 아님) |
| | | **합계** | **117,774** |

§3.2 의 gross 117,200 대비 **+574**. 파일별 배정도 §3.2 와 일치한다(각 파일 절의 소계 참조).

#### 재유입 — 네 가지 산정과 두 가지 실패 시나리오

**재유입은 옮긴 문자량이 아니라 stub 에 쓰는 포인터 줄의 개수에 비례한다**(§D.3.2 가 남긴 잔여). 이 표가 그 개수를 처음으로 셀 수 있게 했다: **M1 행 72개**, 그 행들이 담는 **절 75개**(R-13 이 4개 절을 한 행에 묶는다), 목적지 **companion 13개**(기존 11 + 신규 2), 그리고 `CLAUDE.md` M1′ 가 가리키는 **정본 룰 11개**. 기존 분리의 평균 포인터 길이 **292.4자/줄**은 §3.1 의 실측 인용이다.

| 산정 | 포인터 줄 | 재유입 | net(117,774 −) | 요구 96,943 대비 | 판정 |
|---|---:|---:|---:|---:|---|
| **A. 목적지당 1줄** — companion 13 + `CLAUDE.md` 정본 11, 한 줄이 여러 절을 함께 호명(약 350자/줄) | 24 | 8,400 | 109,374 | **+12,431** | 통과 |
| **B. §3.2 율을 M1·M1′ 분에만 적용**(M2 는 포인터를 만들지 않는다) | — | 4,506 | 113,268 | **+16,325** | 통과 |
| **C. M1 행마다 1줄** (72 × 292.4) | 72 | 21,053 | 96,721 | **−222** | **미달** |
| **D. 절마다 1줄** (75 × 292.4) | 75 | 21,930 | 95,844 | **−1,099** | **미달** |

**C·D 가 미달한다는 것이 이 표의 두 번째 큰 사실이다.** 옮기는 절(또는 행)마다 포인터 줄을 따로 쓰면 요구 감축량에 222~1,099자 못 미친다. **[개정 2]** 이 표의 `요구 96,943 대비` 열과 `판정` 열도 §3.3 과 같은 이유로 폐기됐다(96,943 = 246,943 − 150,000). 산정 A 를 고른 **설계 결정 자체는 유효하다** — 포인터 줄을 목적지당 하나로 묶는 규율(아래 문단)은 재유입을 줄이는 방향이므로 목표가 무엇이든 옳다. 폐기된 것은 「그러면 통과한다」는 결론뿐이다. §3.2 의 투영(여유 11,238)은 재유입을 **자수 비례**로 계산했으므로 이 위험을 원리상 볼 수 없었다 — §D.3.2 가 "판정까지 살아남지 않는다"고 본 잔여가 **실제로는 판정을 뒤집는 구간에 닿아 있다.** 그 평가를 정정한다.

따라서 M2~M6.5 에 구속하는 설계 결정 하나를 여기서 정한다: **포인터 줄은 목적지당 하나로 묶고, 그 한 줄이 옮긴 절들을 이름으로 함께 호명한다**(산정 A). REQ-ALD2-005a 는 "옮겨진 절을 이름으로 호명"을 요구하고 "절마다 한 줄"을 요구하지 않으므로 이 묶음은 허용된다. REQ-ALD2-018(포인터 줄에 새 규범 내용 금지)도 그대로 지킨다 — 묶인 줄도 절 이름과 로드 트리거만 담는다.

post-state(산정 A): 246,943 − 109,374 = **137,569** < 150,000. 산정 C 를 만나도 post-state 는 150,222 로 **한도를 넘는다** — 이것이 이 결정을 규율로 적어 두는 이유다.

> **[개정 2 — 리드 판정 2026-09-26 · 폐기] 위 post-state 단언은 성립하지 않는다.** 137,569 는 `acceptance.md §AC-ALD2-001.1` 의 최저 하한 `F₀ = 149,195` 보다 **11,626 낮아** 구조상 도달 불가하다. 산정 A 를 고른 설계 결정(포인터 줄을 목적지당 하나로 묶는다)은 유효하고, 폐기된 것은 그 결정이 한도를 넘기지 **않는다**는 결론이다. 실측 post-state 는 **197,897** 이었다(런타임 한도 대비 잔여 47,897). 현행 판정 기준은 `acceptance.md §AC-ALD2-001` 의 기록 술어다.

#### 이 투영이 확립하지 않는 것

- 배정 삭제량(`(계획)` 표시 칸과 §4.2 의 M2 행 전부)은 **측정이 아니라 배정**이다. 실제 압축량은 run 중 절마다 재야 확정된다. §4.0 의 절 단위 풀(93,641)만이 실측이다.
- M2 의 49,665자는 **REQ-ALD2-012(의미 보존) 기계 검사가 없는 분량**이다(§D.4). 이 카드 감축의 **42%** 가 그 축에 놓인다 — §3.2 투영에서는 이 비중이 보이지 않았다.
- 파일당 40,000자 축: 이 표의 목적지 13개 중 **하나도** 40,000 을 넘지 않는다(최대 `agent-common-protocol-reference` 38,315, 여유 1,685). 초과 파일 4개는 목적지로 쓰지 않았다(REQ-ALD2-015). 신규 2개는 5,121 · 5,442 로 한도에서 멀다.
- 앵커: 이 표의 절 제목을 가리키는 `§`-형태 인용 중 **10건이 초과 파일 4개 안에** 있다(R-14 1 · R-34 1 · R-57 1 · R-63 2 · R-64 1 · R-65 1 · R-66 2 · R-56 1). §5 의 선례에 따라 길이 중립 수리를 먼저 시도하고, 불가하면 수리하지 않고 `초과 파일 앵커 미수리` 로 기록한다.

---

### §4.5 스키마 — 각 행이 만족해야 하는 조건 (변경 없음)

0. **`기제`** 는 `M1`(companion 으로 이동) · `M1′`(중복 제거) · `M2`(제자리 압축에 의한 삭제) 중 하나다. 이 값이 아래 조건 2·3 의 적용 여부를 정한다.
1. **제거되는 절**은 `[HARD]` / `MUST` / `MUST NOT` / `shall ` 을 한 줄도 포함하지 않는다(REQ-ALD2-002·003).
2. *(M1 행에만 적용)* **`paths:`** 는 부모 stub 경로만 담지 않는다. 그 companion 이 다루는 작업을 하는 동안 세션이 실제로 건드리는 경로를 함께 키로 잡는다(REQ-ALD2-004).
3. *(M1 행에만 적용)* **domain-keyed 근거**는 "이 companion 이 필요해지는 작업은 무엇이며 그 작업이 어느 경로를 건드리는가"를 한 문장으로 답한다.

   **M1′·M2 행에서 2·3 은 `해당없음` 이다** — 목적지가 없으므로 답할 대상이 없다. `해당없음` 은 빈칸이 아니며, 조건 4(범위 의존)는 **기제와 무관하게 모든 행에** 적용된다.
4. **범위 의존**은 제거되는 절에 **지배받는** stub 내 구속 조항을 **열거**하고, 조항마다 두 질문에 모두 답한다 — **Q1** 이동 후에도 자기 범위 문맥에 도달하는가(축소), **Q2** 범위가 이동 전보다 넓어지지 않았는가(확대).

   **선택 술어 — 닫힌 목록이 아니라 열린 정의다.**

   > 제거되는 절에 **어떤 경로로든 범위를 의존하는** stub 내 구속 조항을 전부 열거한다.

   판정 기준은 **"이 절이 사라지면 이 조항이 구속하는 대상이 달라지는가"** 하나다. 달라지면 열거 대상이고, 아니면 아니다.

   흔한 경로 셋을 **예시로** 든다 — 이 셋이 정의가 아니라 정의의 사례다: (a) 그 절을 명시적으로 **참조**하는 조항, (b) 제거 지점에 **인접**한 조항, (c) 그 절이 정의·한정하던 **용어**를 쓰는 조항.

   **셋은 전부 근접성을 요구하므로, 근접하지 않은 지배를 놓친다.** 확인된 사례: 룰 **전체의 적용 범위를 세우는 절**(`kanban-dispatch.md` 의 `## Scope — when this rule is live`)은 그 파일의 모든 조항을 **임의의 거리에서** 지배한다. 그 절이 떠나면 `[HARD] Never spawn background load` 는 kanban 레인 한정이던 것이 매 세션 구속으로 읽힌다 — (a)(b)(c) 어디에도 걸리지 않으면서.

   따라서 **세 가지 검사를 돌리지 말고 지배 관계를 판단하라.** 목록에 없는 경로를 찾는 것이 검토자의 일이며, `없음` 은 정의에 비추어 판단한 결과일 때만 적는다 — 셋을 훑고 적은 `없음` 은 빈칸과 같다.

   이 표를 채우며 실제로 목록 밖에서 찾은 두 경로를 기록해 둔다: **외부 파일의 조항이 지배하는 경우**(R-34 — `session-handoff.md` 의 `[HARD]` 가 `goal-directive.md` 의 절에 의존한다. (a)(b)(c) 는 모두 같은 파일 안을 본다), 그리고 **조건부 성립**(R-49·R-50·R-70 — 같은 절의 한 문장을 남기는지에 따라 Q2 의 답이 뒤집힌다. 절을 원자 단위로 보는 셋은 이 갈림을 표현할 수 없다).

   Q2 가 따로 필요한 이유: 범위를 **좁히던** 비구속 문장이 절과 함께 떠나면 남은 구속 조항이 더 넓게 읽힌다. 결손이 아니라 증가 방향의 실패라 손실 탐지기가 발화하지 않고, Q1 만 묻는 칸은 그것을 통과시킨다.

4번 칸이 존재하는 이유는 §7 의 첫 번째 위험(범위 이탈)을 **행 단위 산출물로 끌어내리기 위해서**다. 동결 해시는 구속 줄을 지키지만 그 줄을 범위 짓는 문맥은 지키지 않으므로, 이 칸이 비면 그 위험은 아무 데서도 판정되지 않는다. 전문: `acceptance.md §D.3`.

선례로 삼을 `paths:` — `goal-directive-detail.md` 가 유일한 domain-keyed 선례다. self-keyed 반례: `paths` 에 부모 룰 파일 하나만 적으면 그 룰을 **편집할 때만** 로드되므로, 그 룰이 지배하는 작업을 하는 세션에는 영원히 도달하지 않는다. 이 카드가 쓰는 목적지 중 3개(`agent-common-protocol-reference` · `askuser-protocol-reference` · `cache-aware-execution-reference`)가 현재 그 반례 상태이며, 위 각 절에 수정안을 적었다.

---

## §5. 앵커 무결성 (REQ-ALD2-009)

절이 stub 에서 companion 으로 옮겨지면, 그 절을 `§`·절 제목·파일명+절 형태로 가리키던 교차참조가 끊길 수 있다. 끊김은 조용하다 — 어떤 빌드도 실패하지 않는다.

점검 방식(run phase 에서 실행):

- **제거된 절마다**(이동·삭제 무관) 그 절 제목 문자열을 저장소 전체에서 역방향 grep 한다. 삭제 경로도 앵커를 끊으며, 그쪽은 경유할 stub 포인터조차 없어 더 조용히 끊긴다.
- 히트한 모든 인용에 대해, 인용이 가리키는 파일이 **이동 후의 위치**를 가리키는지 확인한다.
- **이동 행**: stub 에 남는 포인터 줄이 옮긴 절을 **이름으로** 호명하므로(REQ-ALD2-005a), stub 을 경유하는 인용은 자동으로 살아난다 — 확인 대상은 stub 의 포인터를 경유하지 않는 인용, 즉 절 제목만 인용하거나 행 번호·앵커로 직접 가리키던 것이다.
- **삭제 행**: 경유할 포인터가 없으므로 그 절을 가리키던 인용은 **전부** 수리 대상이다.

---

## §6. template ↔ live 동등 (REQ-ALD2-007)

이 카드가 건드리는 모든 파일은 사본이 둘이다. 두 사본은 같은 커밋에서 함께 바뀐다.

> **[개정 2 — 리드 판정 2026-09-26] 「동등」은 `diff` 0 이 아니다.** 이 절의 표제와 위 문장이 함의하던 `diff` 0 은 템플릿 중립성이 금지하는 내용을 라이브 사본이 정당하게 담으므로 **동시에 성립할 수 없었다**(기준선 ref 에서 이미 30쌍 중 11쌍이 그 사유로 분기해 있었다). 현행 술어: **이 카드가 만든 구조 변경을 같은 커밋에서 빠짐없이 미러하고, 남는 분기는 전부 중립성 허용 클래스로 설명된다.** 승계된 분기의 해소는 이 카드가 지지 않는다(`spec.md §D` 의 `### Out of Scope — 승계된 template/live 분기의 해소`). 판정 기제는 `acceptance.md §AC-ALD2-003.1`~`.4`, 개정된 요구사항 문구는 `spec.md` REQ-ALD2-007.

주의할 기존 분기 하나가 이미 발견됐다 — `skill-routing.md`: 라이브 사본은 `paths:` frontmatter 를 갖는데 템플릿 사본에는 frontmatter 가 **아예 없다**. 그대로 두면 사용자 프로젝트에 always-loaded 로 배포된다. 이 카드가 고친다(REQ-ALD2-008).

`CLAUDE.md` / `AGENTS.md` 도 템플릿 미러를 갖는지 run phase 착수 시 확인한다 — 확인 전에 가정하지 않는다.

---

## §7. 위험

| 위험 | 완화 |
|---|---|
| **범위 이탈 (방향 1 — 축소)** — 절 제목·한정 표·참조 열거가 떠나고 구속 줄만 남아, 해시는 통과하는데 그 줄이 구속할 대상이 사라짐 | `§4` 범위 의존 칸 **Q1** + `plan.md §C` M1 조건 1 단서(구속 조항이 의존하는 표·제목인 절은 옮기지 않는다). 해시로는 안 잡히므로 판정은 검토 몫 |
| **범위 이탈 (방향 2 — 확대)** — 범위를 **좁히던** 비구속 문장이 지워져 남은 구속 조항이 이전보다 넓게 읽힘. 결손이 아니라 **증가** 방향이라 손실 탐지기가 발화하지 않는다 | `§4` 범위 의존 칸 **Q2** + 같은 단서의 확대판(범위를 좁히는 문장은 비구속이어도 지우지 않는다). 전문 `acceptance.md §D.3` |
| companion 이 비대해짐 | **합계 축에서만** 허용된다 — companion 은 always-loaded 가 아니다. **파일당 축에서는 40,000자에서 막힌다**(REQ-ALD2-014). REQ-ALD2-006(원본에 없던 내용 금지)도 함께 구속 |
| 파일별 목표 중 일부 미달 | **[개정 2 — 리드 판정 2026-09-26 · 이 완화는 폐기됐다. 현행 정책이 아니다.]** 종전 완화: 「개정 투영(§3.3)에 **net 여유 11,238자**, 재유입이 실측의 2배여도 통과(+2,220)」. 폐기 사유: 그 여유는 `net − 96,943` 이고 96,943 은 `246,943 − 150,000` 이므로, **150,000 이 도달 불가로 확정되면서 여유 자체가 정의되지 않는 수치가 됐다**. 완화는 더 이상 이 축을 흡수하지 않는다. **현행**: 미달은 완화되지 않고 **기록된다** — AC-ALD2-001 이 합계·감축·런타임 한도 잔여를 귀속과 함께 적고(`acceptance.md §AC-ALD2-001`), 잔여의 성격은 `A_adm` 미측정 때문에 확립되지 않은 채로 남는다(`§AC-ALD2-001.3`). 회수처 지목 자체는 유효하다 — **재유입률이 낮고** 풀에 잔량이 있는 파일(`moai-mcp-tools.md` 0.86% · `main-checkout-branch-guard.md` 1.69%); 고율 파일(`kanban-dispatch` 17.44%)은 순이득이 작아 쓰지 않는다 |
| `AGENTS.md` 압축이 의미를 깎음 | M2 는 구속 조항 줄을 못 건드리고, 구속 조항 해시가 그것을 강제한다. 비구속 산문의 의미 보존은 REQ-ALD2-012 가 구속하며 기계 검사가 없다 — 검토로 판정 |

🗿 MoAI
