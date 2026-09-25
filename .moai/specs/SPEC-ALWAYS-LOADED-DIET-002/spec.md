---
id: SPEC-ALWAYS-LOADED-DIET-002
title: "지시문 예산 두 축 — 18파일 합계 150,000자 미만 + 룰 파일당 40,000자 미만, 구속 조항 무손실"
version: "0.5.0"
status: draft
created: 2026-09-25
updated: 2026-09-25
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "CLAUDE.md, AGENTS.md, .claude/rules/moai, internal/template/templates/.claude/rules/moai"
lifecycle: spec-anchored
tags: "always-loaded, instruction-budget, per-file-limit, context-diet, rule-split, binding-clause-integrity, template-mirror"
tier: L
era: V3R6
related_specs: [SPEC-ALWAYS-LOADED-DIET-001, SPEC-AGENTS-MD-CANON-001, SPEC-V3R6-RULES-PATH-SCOPE-001]
---

# SPEC: always-loaded 지시문 18파일 표면을 150,000자 아래로

## HISTORY

| 날짜 | 버전 | 변경 | 작성자 |
|---|---|---|---|
| 2026-09-25 | 0.1.0 | 최초 작성 (card t1175, plan phase). 근거는 `.moai/reports/t1175/discovery.md` | manager-spec |
| 2026-09-25 | 0.5.1 | §A 에 배포-템플릿 교차 확인을 추가. 격리된 `moai init` 기본 트리(CC 2.1.282)가 동일 경고(18파일 249.2k > 150.0k)를 내므로 이 경고는 도그푸드 한정이 아니다. 출처 `.moai/reports/t1184/verdict.md` (card t1184 관측 인용) | MoAI |
| 2026-09-25 | 0.5.0 | plan-audit iter4 (PASS-WITH-DEBT 0.890) 수리. **N7** — 범위 의존 술어의 **닫힌 열거 자체가 결함**이었다. D7→N2→N7 세 번 모두 인용된 사례는 닫혔는데 클래스가 열린 채 남았고, 원인은 처방이 "대상은 이 셋"이었던 것. 열린 정의("이 절이 사라지면 이 조항이 구속하는 대상이 달라지는가")로 교체하고 참조·인접·용어를 **예시로 강등**. 확인 사례: 룰 전체의 적용 범위를 세우는 절은 임의 거리에서 지배해 근접성 기반 세 가지에 모두 안 걸린다. **D9 종결** — §E 에 전용 AC 열을 추가해 여덟 행 전부 병기(종전 5~8 만). 앵커 무결성을 삭제 경로로 확장, 감사 자기 Gap 3건을 §D.3.3 에 기록 | manager-spec |
| 2026-09-25 | 0.4.0 | plan-audit iter3 (FAIL 0.845, +0.040) 수리. **N6** 재배치 가능 풀 172,363 → **170,847** 정정 — 측정 스크립트가 문단 구분자 2자를 재배치 쪽에 귀속시키던 생성기 산물(`goal-directive` 6,877 은 파일 크기 6,875 를 넘는 산출 불가 값이었다). **D4 회귀** — §2 는 템플릿, §2.1 은 라이브였는데 수리가 재측정 없이 반대로 단언했다. 두 트리 재측정 후 라이브로 통일, 분기 6/14 명시. **N2** 조항 선택 술어를 참조 → **지배(참조∪인접∪용어)** 로 교체. **N3** §4 스키마에 `기제` 열 추가, 삭제 행 규정. N4·N5·N1 수리, 율 표 복제 상태(감사 부록이 14/14 재현 — 잔여 위험 아닌 경위 기록)와 재유입 입도 잔여를 공개에 편입 | manager-spec |
| 2026-09-25 | 0.3.0 | plan-audit iter2 (FAIL 0.805) 수리. **D5 는 설계 변경** — 포인터 줄 재유입 항을 파일별 가중으로 실측(종전 목표 8,796 / 개정 목표 9,019 — 이동량에 비례)해 투영을 gross/net 으로 재구성하고 저재유입 파일 8개의 목표를 상향(107,800 → 117,200 gross, net 108,181, 여유 11,238). **D7** 삭제 경로를 재배치 표 발동 조건에 편입(REQ-017)하고 stub 포인터 줄 검사를 신설(REQ-018). D1·D2·D3 및 마이너 4건 수리 | manager-spec |
| 2026-09-25 | 0.2.0 | 리드 판정으로 **파일당 40,000자 축**을 범위에 편입(REQ-014·015·016, AC-009). 이 축과 충돌한 `design.md §2`("새 companion 을 만들지 않는다")를 개정 — 목적지는 수용량으로 고른다. 이미 초과한 룰 4개의 수리는 범위 밖, 악화 금지만 진다 | manager-spec |

---

## §A. 배경과 문제

Claude Code 런타임이 세션 시작 시 다음 경고를 낸다 — 운영자 세션이 실제로 띄운 문구이며, 단위와 한도가 문구 자체에 적혀 있다(추론이 아니다).

> ⚠ 18 instruction files add up to 244.7k chars, over the 150.0k-char total limit

**단위는 문자(chars), 한도는 합계 150,000자.** 이 트리에서 그 18개 파일을 `wc -m` 으로 실측하면 **246,943자**다(`.moai/reports/t1175/discovery.md` §2·§4). 따라서 **96,943자를 덜어내야 한다.**

18개 파일의 구성: always-loaded 룰 14개 + `CLAUDE.md` + `AGENTS.md` + `@`-import 되는 설정 2개(`.moai/config/sections/user.yaml`, `language.yaml`). 출력 스타일(`.claude/output-styles/moai/moai.md`)은 이 집합에 **없다**.

### 이 경고는 도그푸드 트리만의 현상이 아니다 — 배포 템플릿에서도 뜬다

이 트리에서 잰 246,943자가 저장소 고유의 사정이 아닌지는 **별도 카드(t1184)가 격리 환경에서 교차 확인**했다. `/tmp` 에 `moai init --non-interactive` 로 띄운 기본 프로젝트(CC 2.1.282, `HOME`·`CLAUDE_CONFIG_DIR` 를 `/tmp` 로 돌려 실제 사용자 설정과 격리)의 기동 화면이 같은 경고를 냈다:

> ⚠ 18 instruction files add up to 249.2k chars, over the 150.0k-char total limit

즉 **배포 템플릿 그대로 쓰는 모든 사용자에게 매 세션 뜬다.** 파일 수 18개가 일치하고 합계도 이 트리의 246,943자와 같은 자릿수여서, 이 카드가 겨눈 표면은 도그푸드 사본이 아니라 배포되는 표면 자체다. 근거: `.moai/reports/t1184/verdict.md`(primary 체크아웃의 로컬 증거 — 이 워크트리에서 재측정하지 않았고, 수치는 t1184 의 관측을 인용한 것이다).

### 이 카드가 겨누는 계량기가 어느 것인지

같은 저장소에 크기를 재는 계량기가 둘 있고, **표면도 단위도 서로 다르다.** 둘을 뭉뚱그리면 잘못된 파일을 잘못된 단위로 줄이게 되므로 여기서 갈라 둔다.

| 계량기 | 표면 | 단위 | 한도 | 이 카드 |
|---|---|---|---|---|
| Claude Code 런타임 경고 | 18파일(출력 스타일 제외, yaml import 포함) | 문자 | 150,000 | **겨눈다** |
| `AlwaysLoadedTokenBudget` (`internal/config/token_budget_guard.go`) | 출력 스타일 포함, yaml import 제외 | 토큰 = floor(bytes/4) | 77,600 | 건드리지 않는다 |

---

## §B. 설계를 지배하는 제약 — 구속 조항은 옮기지 않는다

`SPEC-AGENTS-MD-CANON-001` REQ-AMC-002 는 `[HARD]` 조항, 그리고 `MUST` / `MUST NOT` / `shall` 의무를 skill·lazy companion·그 밖의 on-demand 표면으로 **재배치하는 것을 금지한다.** 옮겨도 되는 것은 근거(rationale), 절차, 예시, 사고 기록, 교차참조 표뿐이다. 해당 SPEC `spec.md:475` 의 문장: "No `[HARD]` demotion. A rule that is not always present cannot bind every turn."

실측된 재배치 가능 풀은 **170,847자**(라이브 트리)이고 요구 감축량은 96,943자 — 풀이 요구치의 **1.76배**이므로, **구속 조항을 단 하나도 강등하지 않고 목표에 도달할 수 있다**(§F 표). 계획된 gross 감축 117,200 은 풀의 68.6%이며, 16개 파일 전부에서 목표가 그 파일의 풀 안에 들어간다(`design.md §1`).

**휴리스틱의 한계를 숨기지 않고 적는다.** 170,847 이라는 수치는 문단 단위 분류(`[HARD]`·`MUST`·`shall ` 포함 여부)에서 나온 값이고, 이 방식은 풀을 **과대평가한다** — 구속 조항이 의존하는 문맥(참조하는 표, 범위를 정하는 제목)을 담은 문단도 "비구속"으로 분류되기 때문이다. 170,847 은 **적격성의 상한이지 달성 가능량의 상한이 아니다.**

이 수치는 iter3 N6 에서 172,363 → 170,847 로 정정됐다 — 측정 스크립트가 문단 구분자 2자를 재배치 가능 쪽에 잘못 귀속시키고 있었다(`design.md §1` 개정 표시). 결론의 방향은 바뀌지 않았으나, 헤드라인 수치였으므로 자릿수만 고치지 않고 주장 자체를 다시 확인했다.

### 감축 기제는 둘이며, 둘 다 구속 조항을 건드리지 않는다

| 기제 | 정의 | 구속 조항 취급 |
|---|---|---|
| M1 — 재배치 | 비구속 산문을 path-scoped companion 으로 옮기고 stub 에 포인터를 남긴다. 목적지는 **수용량이 있는** 기존 companion, 없으면 신규 companion(§D 파일당 한도 축 — `design.md §2`) | 옮기지 않는다 |
| M2 — 제자리 압축 | 비구속 산문을 의미 손실 없이 줄인다. companion 으로 옮기지 않는다 | 한 글자도 고치지 않는다 |

M2 가 필요한 이유는 `AGENTS.md` 다 — 그 파일은 "다른 어떤 지시 파일도 로드되지 않았다고 가정한다"는 자기충족성을 스스로 선언하고, Claude 가 아닌 하네스(Codex)가 companion 기제 없이 그대로 읽는다. 따라서 `AGENTS.md` 에는 M1 을 적용하지 않는다.

---

## §C. GEARS 요구사항

- **REQ-ALD2-001** (Ubiquitous) — The always-loaded 지시문 표면(경고가 세는 18개 파일) shall `wc -m` 합계 150,000자 미만을 유지한다.
- **REQ-ALD2-002** (Unwanted) — The 이 SPEC 의 구현 shall not `[HARD]` 조항 또는 `MUST` / `MUST NOT` / `shall` 의무를 companion·skill·그 밖의 on-demand 표면으로 재배치한다.
- **REQ-ALD2-003** (Ubiquitous) — The 18파일 표면의 구속 조항 줄(`[HARD]` / `MUST` / `shall ` 을 포함하는 줄) shall 작업 전후에 개수와 텍스트가 동일하다. 즉 구속 조항 줄은 재배치도 재작성도 되지 않고 **축자 동결**된다.
- **REQ-ALD2-004** (When) — **When** 한 절이 stub 에서 companion 으로 옮겨질 때, the companion shall `paths:` 를 **domain-keyed** 로 선언한다 — 부모 룰 파일만 가리키는 self-keyed `paths:` 는 그 룰을 편집할 때만 로드되므로 금지한다(`SPEC-ALWAYS-LOADED-DIET-001` REQ-ALD-003; 선례 `goal-directive-detail.md`).
- **REQ-ALD2-005** (When) — **When** stub/companion 분리가 수행될 때, the stub shall 세 요소를 모두 갖춘다: (a) 옮겨진 절을 **이름으로 호명**하고 작업 모양의 로드 트리거를 적은 포인터 줄, (b) companion 이 선언하는 자기 소유 경계, (c) 분리를 기록하는 stub 푸터 버전 줄(REQ-ALD-004).
- **REQ-ALD2-006** (Unwanted) — The companion shall not 원본에 없던 내용을 새로 획득한다(REQ-ALD-005). 분리는 이동이지 증식이 아니다.
- **REQ-ALD2-007** (Ubiquitous) — The 이 SPEC 이 수정하는 모든 룰 파일 shall 라이브 사본(`.claude/rules/**`)과 템플릿 사본(`internal/template/templates/.claude/rules/**`)이 같은 커밋에서 동등해진다 — 이미 존재하는 분기의 해소를 포함한다.
- **REQ-ALD2-008** (When) — **When** `skill-routing.md` 의 템플릿 미러가 수정될 때, the 템플릿 사본 shall 라이브 사본과 동일한 `paths:` frontmatter 를 갖는다. 현재 템플릿 사본에는 frontmatter 가 전혀 없어 사용자 프로젝트에 always-loaded 로 배포된다.
- **REQ-ALD2-009** (Ubiquitous) — The 옮겨진 절을 가리키는 모든 교차참조 shall 이동 후에도 해소된다 — 끊긴 앵커를 남기지 않는다.
- **REQ-ALD2-010** (Ubiquitous) — The 구현 shall **always-loaded 표면에서 제거된 모든 비구속 산문**을 재배치 표에 한 행씩 나열한다 — 기제가 M1(companion 으로 이동)이든 M1′(중복 제거)이든 M2(제자리 압축에 의한 삭제)든 무관하다. 이동 행은 목적지 companion·그 `paths:` 키·domain-keyed 근거를 채우고, **삭제 행은 목적지 칸에 `삭제(목적지 없음)` 를 적는다.** 두 종류 모두 범위 의존 칸(Q1/Q2)을 채운다.
- **REQ-ALD2-011** (Where) — **Where** 대상 파일이 `AGENTS.md` 인 경우, the 구현 shall M2(제자리 압축)만 적용하고 M1(companion 재배치)을 적용하지 않는다 — 그 파일의 자기충족성 선언과 비-Claude 하네스의 읽기 경로 때문이다.
- **REQ-ALD2-012** (While) — **While** M2 제자리 압축이 수행되는 동안, the 구현 shall 의미를 보존한다 — 의무·조건·예외·수치 중 어느 것도 삭제하지 않는다.
- **REQ-ALD2-013** (Unwanted) — The 구현 shall not always-loaded 표면의 의무 사본을, 같은 의무가 path-scoped 룰에도 존재한다는 이유로 삭제한다. 그 룰이 로드되지 않는 턴에는 의무가 사라지므로, 중복 제거로 위장한 REQ-AMC-002 강등이다. 중복 제거(M1′)는 **비구속 산문에만** 적용된다.
- **REQ-ALD2-014** (Ubiquitous) — The 이 SPEC 이 내용을 쓰는 모든 룰 파일 shall 작업 후 40,000자 미만으로 끝난다 — `moai hook instructions-loaded` 가 로드되는 파일마다 강제하는 한도다(`internal/hook/instructions_loaded.go:86-104`). 이 축은 REQ-ALD2-001(18파일 합계)과 **별개**이며, 어느 쪽의 통과도 다른 쪽을 함의하지 않는다.
- **REQ-ALD2-015** (Unwanted) — The 구현 shall not 40,000자를 이미 초과한 룰 파일에 내용을 추가한다. 초과 파일을 목적지로 쓰면 그 파일의 초과에 기여하게 되고, 수리 의무가 이 카드로 넘어온다.
- **REQ-ALD2-016** (Ubiquitous) — The 40,000자를 초과하는 룰 파일의 개수 shall 작업 전후로 늘지 않는다. 기준선은 두 트리 각각 **4개**다(실측). 이는 신규 companion 이 스스로 한도를 넘기는 경우를 잡는 래칫이다.
- **REQ-ALD2-017** (Unwanted) — The 재배치 표 shall not 삭제 경로를 누락한다. 표의 발동 조건이 "옮겼을 때"뿐이면 `AGENTS.md`(M2)와 `CLAUDE.md`(M1′)는 **정의상 행을 만들지 않으며**, 두 파일의 감축분(계획의 11% 이상)이 범위 이탈 검토를 통째로 건너뛴다.
- **REQ-ALD2-018** (When) — **When** stub 에 새 포인터 줄이 쓰일 때(REQ-ALD2-005a 가 요구하는 바), the 포인터 줄 shall 옮긴 절을 이름으로 호명하는 것 외의 새 규범적 내용을 담지 않는다. REQ-ALD2-006 은 companion 에만 걸리므로, 이 조항이 없으면 **stub 은 어떤 기준도 검사하지 않는 표면**이 된다.

---

## §D. 범위 밖

### Out of Scope — 출력 스타일

- `.claude/output-styles/moai/moai.md` 는 경고가 세는 18파일 집합에 없다. 이 카드는 그 파일을 읽지도 고치지도 않는다.

### Out of Scope — `AlwaysLoadedTokenBudget` 가드

- `internal/config/token_budget_guard.go` 의 `AlwaysLoadedTokenBudget = 77600`, 그 0.22% 여유(77,600 중 168 토큰), 그리고 76,000 → 77,600 상향 사슬은 이 카드가 해소하지 않는다.
- 그 가드는 **다른 표면**(출력 스타일 포함, yaml import 제외)을 **다른 단위**(토큰 = floor(bytes/4))로 잰다. 두 계량기를 한데 묶지 않는다.
- 부수 효과로 그 가드의 수치가 내려갈 수는 있으나, 이 SPEC 의 어떤 AC 도 그것을 판정 대상으로 삼지 않는다.

### Out of Scope — 경고를 낸 트리의 식별

- 경고의 파일별 수치(36.1k / 27.5k / 24.3k)는 `main` 에서도 `develop` 에서도 재현되지 않으며 제3의 트리를 가리킨다(`discovery.md` §3). 이 카드는 그 트리를 특정하지 않는다 — 한도 150,000 은 선언된 값이지 도출된 값이 아니고, 고쳐야 할 트리는 배포되는 트리이기 때문이다.

### Out of Scope — kanban-dispatch 의 조건부 주입

- `kanban-dispatch.md` 를 always-loaded 표면에서 떼어내 `-k`/`-f` 세션에만 SessionStart `AdditionalContext` 로 주입하는 방안은 기계적으로 가능하지만 REQ-AMC-002 가 금지하고, §B 에 따라 필요하지도 않다. 재제안되지 않도록 여기에 기각 사실을 남긴다.

### Out of Scope — `CLAUDE.md` 템플릿/라이브 분기의 해소

- 두 사본은 **의미 수준에서** 다르다. 라이브 루트 `CLAUDE.md` §15 는 "CG Mode (`moai cg`, requires tmux)" 를 살아 있는 기능으로 서술하고 §10 의 GLM 라우팅 문장도 같은 축이며, 템플릿 사본은 "`moai cg` is retired. Run `moai migrate cg`" 로 서술한다.
- **어느 쪽이 낡았는지는 실측됐다.** 설치된 바이너리에서 `moai cg --help` 는 `moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration` 을 출력하고, `moai migrate cg` 는 존재한다(2026-09-25, 이 트리에서 확인). 따라서 **템플릿 사본이 옳고 라이브 루트 사본이 낡았다** — 은퇴한 기능을 살아 있다고 서술하고 있다.
- 그럼에도 이 카드는 그 분기를 해소하지 않는다. 낡음은 **적재 바이트 수를 전혀 건드리지 않으므로** 지시문 예산 다이어트와 직교하는 축이고, 내용 정정을 같은 diff 에 접으면 커밋 하나가 질문 둘에 답하게 되어 감사가 둘 다를 판정해야 한다. 별도 카드로 제기된다.
- 대조되는 판단 하나를 나란히 둔다 — `skill-routing.md` 의 미러 결함은 **범위 안**이다. 그 결함은 always-loaded 배포를 직접 일으키므로 이 카드의 주제 자체이기 때문이다. 두 결함의 취급이 다른 것은 일관성 없음이 아니라 이 판별식의 적용이다.
- 귀결: AC-ALD2-003(미러 동등)의 판정 대상에서 `CLAUDE.md` 를 제외한다. **이 제외는 범위 결정이지, 두 사본이 일치한다는 주장이 아니다.**

### Out of Scope — path-scoped 룰 72개

- 나머지 72개 path-scoped 룰(1,047,263자)은 **합계 축(REQ-ALD2-001)** 의 대상이 아니다 — always-loaded 표면에 없기 때문이다. companion 이 커지는 것 자체는 그 축에서 의도된 결과다.
- **다만 파일당 축(REQ-ALD2-014·015·016)은 이들에게도 구속한다.** 파일당 한도는 로드되는 파일마다 발화하므로 path-scoped 여부를 가리지 않는다. 즉 "companion 이 커지는 것은 위반이 아니다"는 **합계 축에 한정된 진술**이며, 40,000자를 넘기면서 커지는 것은 위반이다. REQ-ALD2-006(원본에 없던 내용 금지)도 여전히 구속한다.

### Out of Scope — 이미 40,000자를 넘긴 룰 4개의 수리

- 두 트리에서 동일하게 4개가 한도를 넘어 있다(2026-09-25 실측): `workflow/worktree-integration.md` 61,435 · `workflow/session-handoff-examples.md` 41,616 · `workflow/kanban-dispatch-detail.md` 41,036 · `workflow/spec-workflow.md` 40,799.
- **이 카드는 넷 중 어느 것도 수리하지 않는다.** 넷 다 이 카드 이전부터 초과였고, 개정된 목적지 배정(`design.md §2.2`)에 따라 이 카드는 그중 어디에도 쓰지 않는다 — 따라서 어느 것도 악화시키지 않는다.
- 수리를 접지 않는 이유는 규모다. `worktree-integration.md` 하나만 해도 21,435자(3분의 1 이상)를 덜어내야 하고, 그 작업은 자체의 REQ-ALD-003/004/005 의무를 지는 별도 카드 크기다. 접으면 커밋 하나가 질문 둘에 답하게 된다 — `CLAUDE.md` CG 낡음을 범위 밖으로 둔 것과 **같은 판별식**이며, 두 판정이 같은 근거로 서는 것은 일관성이다.
- 대신 이 카드가 지는 것은 **악화 금지**다: REQ-ALD2-015(초과 파일에 추가 금지)와 REQ-ALD2-016(초과 파일 개수 4개에서 늘지 않음). 별도 카드로 제기된다.
- **명시**: 이 범위 결정은 "넷이 괜찮다"는 주장이 아니다. 넷 다 현재 `moai hook instructions-loaded` 위반 상태이며, 그중 `spec-workflow.md` 는 이번 세션에서 실제로 위반 메시지를 냈다.

---

## §E. 성공 기준

| # | 기준 | AC | 판정 |
|---|---|---|---|
| 1 | 18파일 `wc -m` 합계 < 150,000 | AC-ALD2-001 | 기계적 |
| 2 | 구속 조항 줄 목록의 정규화·정렬 sha256 이 작업 전후 동일 | AC-ALD2-002 | 기계적 |
| 3 | 수정한 모든 파일의 template ↔ live 동등 (`CLAUDE.md` 제외) | AC-ALD2-003 | 기계적 |
| 4 | **제거된** 절을 가리키는 교차참조가 전부 해소 | AC-ALD2-005 | 기계적 + 선례 |
| 5 | 재배치 표가 완전하고 모든 companion `paths:` 가 domain-keyed | AC-ALD2-004 · AC-ALD2-006 | 문서 + 검토 |
| 6 | companion 이 원본에 없던 내용을 획득하지 않음 | AC-ALD2-008 | 기계적 |
| 7 | 이 카드가 쓴 모든 룰 파일이 40,000자 미만 | AC-ALD2-009 (a) | 기계적 |
| 8 | 40,000자 초과 룰 파일 개수가 4개에서 늘지 않음 | AC-ALD2-009 (b) | 기계적 |

이 표는 `acceptance.md §D.1` 의 MUST-PASS 8개를 **남김없이 덮는다.** 판정의 정본은 `acceptance.md` 이고, 이 표는 요약이다.

**[iter4 D9 종결]** 종전에는 행 5~8 에만 AC 번호가 붙어 있고 1~4 에는 없었다 — 부분 병기라 대조가 절반만 가능했다. 여덟 행 전부에 전용 `AC` 열로 병기해 닫는다. 같은 수정에서 행 4 의 "옮긴 절"을 "제거된 절"로 고쳤다(N3 의 이동 전용 어법이 남아 있던 자리 — 삭제 경로도 앵커를 끊을 수 있다).

**일대일은 아니다**(iter3 N1): 행 5가 AC-004 와 AC-006 둘을 덮고, AC-009 는 행 7·8 에 걸친다. 종전에 "일대일로 대응한다"고 적은 것은 덮개(coverage)와 대응(bijection)을 혼동한 것이다. 덮개는 성립하므로 기준이 누락되지는 않으며, 고친 것은 **주장**이지 덮개가 아니다.

---

## §F. 기준선 (이 실행, 이 트리)

명령: `wc -m <18 paths>` — 2026-09-25, `.claude/worktrees/t1175`, base develop `a0b78213d`.

```
246943 total
```

구속 조항 줄 기준선 — 16개 마크다운 파일(yaml 2개는 구속 조항 0):

```
170 lines
sha256 d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3
```

토큰 출현 내역: `[HARD]` 128, `MUST` 57, `MUST NOT` 23, `shall` 1.

---

## §G. 교차참조

- `.moai/reports/t1175/discovery.md` — 이 SPEC 의 근거 전량(계량기 동정, 18파일 구성, 재현 Gap, 타당성, 기각안, 선례)
- `.moai/reports/t1175/rules-inventory.tsv` — 템플릿 룰 86개의 크기·로딩 범위
- `SPEC-AGENTS-MD-CANON-001` REQ-AMC-002 — 구속 조항 재배치 금지
- `SPEC-ALWAYS-LOADED-DIET-001` REQ-ALD-003 / -004 / -005 — 분리 규약
- `.claude/rules/moai/development/rule-authoring.md` — always-loaded 표면 정의와 scope-first 의무

🗿 MoAI
