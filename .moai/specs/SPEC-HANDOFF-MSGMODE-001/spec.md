---
id: SPEC-HANDOFF-MSGMODE-001
title: "핸드오프 메시지 오케스트레이션-모드 내장 (message-v2)"
version: "0.1.0"
status: in-progress
created: 2026-07-04
updated: 2026-07-05
author: MoAI
priority: P2
phase: "v3.0.0"
module: "docs-handoff"
lifecycle: spec-anchored
tags: "session-handoff, orchestration-mode, message-v2, mode-seed, handoff-protocol, tier-s, epic-handoff-v2"
era: V3R6
related_specs: [SPEC-HANDOFF-CTXGUIDE-001]
---

# SPEC-HANDOFF-MSGMODE-001 — 핸드오프 메시지 오케스트레이션-모드 내장 (message-v2)

> Epic "Handoff-v2" M2/4. 선행: M1 SPEC-HANDOFF-CTXGUIDE-001 (256K 밴드 로직, origin `60db8e721`에서 **completed**). 후속: M3 auto-resume(handoff.yaml landing) · M4 threshold-guidance 완성.
> 본 SPEC은 **doctrine-only**(문서 전용)이다. Go 코드·config·상태 파일 변경 0. paste-ready session-handoff 메시지 포맷에 오케스트레이션 **모드 시드(seed)**를 내장한다.

## §1 배경 · 목표 · 범위

### §1.1 배경 — 확인된 문제

현재 paste-ready session-handoff 메시지(`.claude/rules/moai/workflow/session-handoff.md` 6-block Canonical Format)는 다음 세션이 **어떤 오케스트레이션 모드로 재개해야 하는지**에 대한 신호를 담지 않는다. `/clear` 이후 재개 세션의 오케스트레이터는 Phase 0.95 모드 선택을 처음부터 다시 추정해야 하며, 직전 세션이 이미 도출했던 mode 판단(예: dynamic-workflow fan-out, agent-team)이 소실된다.

부수적으로, 현행 doctrine에 3가지 정합성 결함(B1/B2/B3)이 관측된다.

- **B1 (placement-clause 편재)**: Block 1의 지시어(directive) 배치 규칙이 `"immediately after `ultrathink.`"` 로 고정되어 있어(`session-handoff.md` L79-L80 + 템플릿 미러), 신규 `mode:` 라인이 추가될 때 배치 순서가 모호해진다.
- **B2 (solo-sequential emission — design-intent 모호성, 사전 존재 conflict 아님)**: `mode: solo-sequential`은 **본 SPEC이 신규 도입**하는 필드다 — live doctrine 4개 표면(SH/M8/양 미러) 어디에도 "never emitted"·"parse-accept" 문구는 부재(grep 0). 따라서 B2는 제거할 사전 conflict가 아니라, **저작 시 회피해야 할 design-intent 모호성**이다: solo-sequential 발화 정책을 저술할 때 "never emitted"(발화 안 함) vs "parse-accepted"(읽기 허용)의 상충 서술을 **도입하지 않고** 단일 "emit-discouraged" 프레이밍으로 저작해야 한다.
- **B3 (self-check 표면 attribution 불일치)**: pre-emit self-check가 표면마다 **서로 다른 concern**으로 존재한다 — SSOT(`session-handoff.md` § Diet Constraints)는 **"paste-ready budget" 9 items**만 보유하고, 렌더 표면(`moai.md §8`)은 이와 **별개 concern**인 **"session-handoff template completeness" 10 items**를 참조하되 이를 `session-handoff.md`가 SSOT인 것처럼 귀속한다(그러나 SH에는 template-completeness self-check 섹션이 부재 — § Cross-references sentinel만 3개 qualifier를 명명). M2 범위: M8의 template-completeness self-check에 `mode:` 검증 1개를 추가(10→11)하고, SH의 paste-ready-budget 9-item self-check는 **불변 유지**하며, SSOT-attribution 불일치 해소는 후속 doctrine-reconciliation chore로 이연한다(§1.3).

### §1.2 목표

paste-ready 메시지 Block 1에 **조건부 `mode:` 라인**을 도입해 오케스트레이션 모드를 **시드(seed)**로 내장한다. 핵심 불변식:

1. `mode:` 값은 Phase 0.95 모드 카탈로그와 1:1 대응하는 4-enum이다(아래 표).
2. `solo-sequential`(기본값)일 때 라인을 **생략** → v1 포맷과 **byte-identical**(공통 케이스 zero-diff).
3. `mode:` 시드는 **permission grant가 아니다** — Implementation Kickoff Approval(plan→run HUMAN GATE)은 시드된 모드와 무관하게 여전히 필수다.

4-enum ↔ Phase 0.95 카탈로그 매핑(`.claude/rules/moai/workflow/orchestration-mode-selection.md` §A 6-mode 카탈로그 대조):

| `mode:` 값 (seed) | Phase 0.95 카탈로그 | 발화 여부 | 비고 |
|-------------------|---------------------|-----------|------|
| `solo-sequential` | Mode 5 (sub-agent, default fallback) | **생략**(기본) | 생략 = v1 byte-identical |
| `parallel-subagents` | Mode 4 (parallel, 3-5 concurrent `Agent()`) | 발화 | mode ≠ solo-sequential |
| `agent-team` | Mode 3 (agent-team, implicit team) | 발화 | Block 5 run에 `--team` 부가 |
| `dynamic-workflow` | Mode 6 (workflow, orchestrator fan-out) | 발화 | opener에 bare `ultracode` 부가 |

**제외**: Mode 1 (trivial) / Mode 2 (background)는 handoff-relevant seed가 아니다 — 핸드오프는 trivial/background를 1차 재개 모드로 삼지 않는다(문서에 제외 근거 명시).

**임계값 재사용**: mode 시드는 신규 임계값을 도입하지 않는다. Phase 0.95의 기존 auto-select 임계값(domains ≥ 3 / files ≥ 10 / score ≥ 7, `orchestration-mode-selection.md` §B.1 참조)에서 파생한다.

**Block 1 지시어(directive) 바인딩 규칙**:

| 지시어 | 발화 조건 |
|--------|-----------|
| `ultrathink.` | **항상**(v1 불변) |
| bare `ultracode` (opener 라인, 예: `ultrathink. ultracode`) | mode = `dynamic-workflow`일 때만 |
| `# /goal <completion-condition>` | 다음 phase = run **AND** machine-verifiable end-state일 때만 (v1 조건 불변; 자율 run 진입 미승인) |
| `--team` (Block 5 run 커맨드 말미) | mode = `agent-team`일 때만 |

**ultracode 재정합(bare vs slash)**: bare `ultracode`는 per-prompt 트리거 키워드(v2.1.160+, `ultrathink`와 동류)로 **붙여넣는 순간 발화**한다. 반면 `# /effort ultracode` 슬래시-주석 라인은 hook가 붙여넣기 시점에 실행할 수 없다. 따라서 opener 기본형은 **bare `ultracode`**를 사용하고, session-persistent `/effort ultracode` 슬래시 형태는 "세션 전체 지속 필요" 별도 변형으로만 유지한다. 이는 현행 `session-handoff.md`/`moai.md §8` Block 1의 `# /effort ultracode` 주석 라인을 정정한다.

### §1.3 Out of Scope (Exclusions)

이 절은 spec-lint Exclusions(h3) 요구를 만족한다. 다음은 M2 범위 밖이며 후속 SPEC/사전 chore 소관이다.

- **`orchestration-mode-selection.md` mirror-drift resync** (live 40597B vs 템플릿 미러 37523B, 3074B drift) → 별도 사전 chore. M2 AC 아님.
- **null session_id filename nonce** → M3(auto-resume handoff.yaml landing) 소관.
- **`handoff.yaml` / `HandoffConfig` struct+loader** → M3 landing · M4 소비.
- **Go 코드 / config / 상태 파일 변경** → M2는 doctrine-only. JSON-twin은 개념(§2 REQ-MSGMODE-007)만 문서화.
- **`dynamic-workflows.md` / `goal-directive.md` / `orchestration-mode-selection.md` 본문 편집** → 병렬 세션 dirty 또는 cross-ref-only. M2는 READ만.
- **`session-handoff.md` template-completeness self-check SSOT-attribution drift** → `moai.md §8`이 "session-handoff template completeness" self-check를 `session-handoff.md`에 존재하는 것으로 귀속하나, SH는 해당 섹션이 부재하고 `paste-ready budget` self-check만 보유한다(§ Cross-references sentinel은 3개 qualifier `paste-ready budget`/`localization render`/`session-handoff template completeness`를 명명). 이 귀속 불일치 해소(SH에 template-completeness self-check 신설 또는 M8 귀속 문구 정정)는 별도 doctrine-reconciliation chore 소관 → M2 아님.

## §2 요구사항 (GEARS/EARS)

- **REQ-MSGMODE-001**: While 다음 세션의 오케스트레이션 모드가 `solo-sequential`이 아닐 때, the system shall session-handoff 메시지 Block 1에 정확히 하나의 `mode: <value>` 라인을 포함해야 한다. Where 모드가 `solo-sequential`(기본)이면 the system shall 해당 라인을 생략하여 메시지를 v1 포맷과 byte-identical하게 유지해야 한다.
- **REQ-MSGMODE-002**: The doctrine shall 4-enum(`solo-sequential`↔Mode 5, `parallel-subagents`↔Mode 4, `agent-team`↔Mode 3, `dynamic-workflow`↔Mode 6)과 Phase 0.95 카탈로그 간 1:1 매핑 표를 SSOT(`session-handoff.md`)와 렌더 표면(`moai.md §8`) 양쪽에 제공해야 한다.
- **REQ-MSGMODE-003**: The doctrine shall Mode 1(trivial) 및 Mode 2(background)를 handoff-relevant seed에서 제외하고 그 제외 근거를 명시해야 한다.
- **REQ-MSGMODE-004**: Where mode 시드를 도출할 때, the system shall Phase 0.95의 기존 auto-select 임계값(domains ≥ 3 / files ≥ 10 / score ≥ 7)을 재사용해야 하며, 신규 임계값을 도입하지 않아야 한다.
- **REQ-MSGMODE-005**: The doctrine shall Block 1 지시어를 mode 시드에 바인딩해야 한다 — `ultrathink.`는 항상; bare `ultracode`는 mode=`dynamic-workflow`일 때만 opener에; `# /goal`은 run-phase + machine-verifiable end-state일 때만; `--team`은 mode=`agent-team`일 때 Block 5 run 커맨드에.
- **REQ-MSGMODE-006**: The doctrine shall opener 기본형으로 bare `ultracode` 키워드(붙여넣기 시점 발화)를 사용하고, session-persistent `/effort ultracode` 슬래시 형태를 "세션 전체 지속" 별도 변형으로 구분해야 하며, 현행 `# /effort ultracode` Block 1 주석 라인을 이에 맞춰 정정해야 한다.
- **REQ-MSGMODE-007**: The `mode:` seed shall SEED(다음 세션 오케스트레이터용 신호)로 문서화되어야 하며 permission grant가 아니어야 한다 — Implementation Kickoff Approval(plan→run HUMAN GATE)은 시드된 모드(dynamic-workflow/agent-team 포함)와 무관하게 필수로 유지되어야 하고, 시드된 모드는 자율 run-phase 진입을 승인하지 않는다.
- **REQ-MSGMODE-008**: Where resume 메시지의 JSON-twin 표현이 존재하거나 이후 도입되면, the system shall 그 `schema_version`을 `2`로 하여 `mode` 필드를 담아야 한다. While 현재 코드베이스에 JSON twin이 부재하므로(doctrine 개념), the system shall 이를 forward-compatibility doctrine note로만 기록하고 코드 변경을 발생시키지 않아야 한다.
- **REQ-MSGMODE-009**: The `mode:` 값 shall `plan|run|sync|mx`와 같은 protocol token으로서 전 locale에 verbatim 보존되어야 하며, the doctrine shall 어떤 localization/translation 표에도 신규 행(row)을 추가하지 않아야 한다.
- **REQ-MSGMODE-010** (B1): The doctrine shall Block 1 배치 규칙 `"immediately after `ultrathink.`"` 를 `"immediately after `ultrathink.` (or after the `mode:` line when present)"` 로 일반화해야 하며, grep으로 확인된 모든 표면(live `session-handoff.md` + 템플릿 미러; 해당 시 `moai.md §8` + 미러)에서 일관되게 갱신해야 한다.
- **REQ-MSGMODE-011** (B2 — forward-authoring guard): The doctrine shall solo-sequential 발화 정책을 단일 **"emit-discouraged"** 프레이밍으로 저작해야 한다 — solo-sequential 라인은 발화하지 않고(Block 1 생략 → v1 byte-identical), 명시적 `mode: solo-sequential` 라인을 만나면 parse-accept(forward-compatible)한다. `mode: solo-sequential`은 본 SPEC 신규 필드이므로 이는 **사전 존재 conflict의 제거가 아니라 저작 시 모순 도입 금지**다: the doctrine shall "never emitted" vs "parse-accepted"의 상충 쌍을 **도입하지 않아야 한다**(단일 emit-discouraged 프레이밍 유지).
- **REQ-MSGMODE-012** (B3): The doctrine shall 렌더 표면(`moai.md §8`)의 **"session-handoff template completeness"** self-check에 Block 1 `mode:` 라인 검증 항목("present iff mode ≠ solo-sequential AND Phase 0.95 카탈로그 일치")을 추가하여 그 항목 수를 **10→11 items**로 갱신해야 한다. SSOT(`session-handoff.md`)의 별개 concern인 **"paste-ready budget" 9-item** self-check는 **불변 유지**되어야 하며(M2에서 SH에 신규 template-completeness self-check 섹션을 생성하지 않는다), 두 self-check는 concern-name qualifier(`paste-ready budget` vs `session-handoff template completeness`)로 명확히 구분 표기되어야 한다. `moai.md §8`이 template-completeness self-check를 `session-handoff.md` SSOT로 귀속하나 SH에 해당 섹션이 부재한 **SSOT-attribution 불일치**는 M2 범위 밖(후속 chore, §1.3)이다.
- **REQ-MSGMODE-013**: When live doctrine 표면을 편집하면, the system shall 동일 편집을 `internal/template/templates/...` 미러에 반영해야 하며(§24 mirror-parity), 미러 콘텐츠는 internal-content-neutral(SPEC-ID·내부 날짜·commit SHA·REQ 토큰 부재)이어야 한다(§25 neutrality).

## §3 인수 기준 (Tier S 인라인)

각 AC는 구현 후 대상 표면에 대한 **grep으로 기계 검증** 가능하다. 대상 표면 약칭: **SH**=`.claude/rules/moai/workflow/session-handoff.md`, **M8**=`.claude/output-styles/moai/moai.md`, **SH-mir**/**M8-mir**=각 `internal/template/templates/...` 미러.

| AC | REQ | Given (표면) | When (검증) | Then (기대) |
|----|-----|--------------|-------------|-------------|
| AC-MSGMODE-001 | 001 | SH Block 1 스펙 | `grep -n 'mode:' SH` (Block 1 field spec) | 조건부 `mode:` 라인 스펙 존재 + "omit"/"solo-sequential"/"byte-identical" 서술 존재 |
| AC-MSGMODE-002 | 002 | SH, M8 | 4 enum 토큰 grep | `solo-sequential`·`parallel-subagents`·`agent-team`·`dynamic-workflow` **AND** `Mode 3`·`Mode 4`·`Mode 5`·`Mode 6`가 SH(전체 매핑표)와 M8(compact 참조) 양쪽에 존재 |
| AC-MSGMODE-003 | 003 | SH | `grep -i 'trivial\|background' SH` | Mode 1 trivial + Mode 2 background 제외 서술("not handoff-relevant"/"excluded") 존재 |
| AC-MSGMODE-004 | 004 | SH | 임계값 grep | `domains ≥ 3`·`files ≥ 10`·`score ≥ 7` 재사용 참조(+ `orchestration-mode-selection.md` cross-ref); 신규 임계값 도입 문구 부재 |
| AC-MSGMODE-005 | 005 | SH | 4 바인딩 규칙 grep | `ultrathink.`(항상)·bare `ultracode`(iff dynamic-workflow)·`/goal`(iff run+verifiable)·`--team`(iff agent-team) 네 규칙 모두 존재 |
| AC-MSGMODE-006 | 006 | SH, M8 | ultracode 재정합 grep (전수) | bare `ultracode` opener 규칙 문서화 **AND** `/effort ultracode` session-persistence 변형 구분 서술 존재; Block 1 default 스켈레톤에서 `# /effort ultracode` opener-default 제거. **전수 일관성**: `grep -n '/effort ultracode\|bare .ultracode' SH M8`로 bare-opener(붙여넣기 발화) vs slash-`/effort ultracode`(세션 지속) 구분이 **모든** occurrence에 일관 적용(SH ~4곳 L33/L79/L80/L168 [L80은 `/goal` 배치 절이 `/effort ultracode`를 2회 참조] · M8 ~2곳 L682/L720)됨을 확인 — 일부만 갱신된 partial-edit drift(구 opener-default `# /effort ultracode`가 어느 한 표면·어느 한 occurrence에 잔존) 0 |
| AC-MSGMODE-007 | 007 | SH | SEED 불변식 grep | "SEED" + "not a permission grant"(또는 동등) + "Implementation Kickoff Approval" 존재; "자율 run 진입 미승인" 서술 존재 |
| AC-MSGMODE-008 | 008 | SH | `grep 'schema_version' SH` | `schema_version: 2` JSON-twin forward-compat note 존재 + "no JSON twin currently"(doctrine-only) 서술 존재 |
| AC-MSGMODE-009 | 009 | SH, M8 | locale 표 컬럼 카운트 | Localization/Cut-line/Header translation 표의 locale 컬럼 = 정확히 4(en/ko/ja/zh) 유지(신규 행 0); `mode:`가 protocol token으로 verbatim 보존 서술 존재 |
| AC-MSGMODE-010 | 010 | SH, SH-mir(+M8/미러) | placement clause grep (directive + /goal sibling) | `"(or after the `mode:` line when present)"` 가 구 clause가 있던 **모든** 표면에 존재(grep count 일치); 확장 없는 구 clause 잔여 0. **/goal sibling 정합**: `/goal` 배치 절(SH ~L80 "immediately after the `/effort ultracode` line")이 신규 배치 불변식 `opener → mode → /goal` 순서를 반영하도록 갱신됨 — `grep -n 'goal' SH`로 `/goal` 배치 절이 `mode:` 라인 순서(opener→mode→goal)를 참조함을 확인(구 ordering `ultrathink.` 직후만 가리키는 잔여 절 0) |
| AC-MSGMODE-011 | 011 | SH, M8 | authored-policy 내부 일관성 grep | 저작된 solo-sequential 정책에 `"emit-discouraged"` 존재 **AND** `"parse-accept"`(forward-compatible, 또는 동등) 존재 **AND** 저작 텍스트가 `"never emitted"` vs `"parse-accepted"` 상충 쌍을 **도입하지 않음**(forward-authoring guard — 사전 존재 conflict 제거가 아니라 신규 모순 미도입; `grep 'never emitted' SH M8` = 0은 저작 후에도 상충 문구가 authored되지 않았음을 확인) |
| AC-MSGMODE-012 | 012 | SH, M8 | self-check count grep (M8-only 갱신 + SH 보존) | `grep -c '11 items' M8` ≥ 1 **AND** M8 "session-handoff template completeness" self-check에 `mode:` 검증 항목 존재 **AND** M8 해당 self-check의 `10 items` 잔여 0; **SH 보존**: `grep -c '9 items' SH` ≥ 1 (paste-ready-budget 9-item self-check 불변). SH에 대한 `11 items` 요구 없음(M2에서 SH 신규 self-check 섹션 미생성) |
| AC-MSGMODE-013 | 013 | SH-mir, M8-mir | mirror parity + neutrality | 미러가 live와 동일 doctrine 편집 반영(`diff` 델타는 neutrality-구동만); `grep -E 'SPEC-HANDOFF-MSGMODE|SPEC-MSGMODE|2026-07-04' SH-mir M8-mir` = 0 (internal-content 무유출) |
| AC-MSGMODE-014 | 001 | SH | v1 byte-identity 서술 | solo-sequential 케이스가 v1과 byte-identical(생략-라인 불변식) 문서화 — `grep 'byte-identical\|zero-diff\|v1' SH` ≥ 1 |

### 추가 검증 (회귀 / 게이트)

- **AC-MSGMODE-015 (spec-lint)**: `moai spec lint .moai/specs/SPEC-HANDOFF-MSGMODE-001/spec.md` **AND** `moai spec lint .moai/specs/SPEC-HANDOFF-MSGMODE-001/plan.md` 가 각각 "No findings"(exit 0), 신규 ERROR 0(기존 baseline 대비 무증가). 검증은 **파일 경로 형식**을 사용한다 — `moai spec lint` CLI는 파일 경로를 읽으며 디렉터리 형식(`.../SPEC-HANDOFF-MSGMODE-001/`)은 콘텐츠와 무관하게 `ParseFailure ... is a directory`(exit 1)를 반환하므로 clean lint을 입증할 수 없다(구조적 미검증). 이 파일 경로 형식은 resume precondition이 이미 사용하는 형식이다.
- **AC-MSGMODE-016 (parity sentinel)**: `session-handoff.md` § Cross-references 및 `moai.md §8` drift-mitigation sentinel의 concern-name qualifier(`paste-ready budget`/`localization render`/`session-handoff template completeness`)와 locale 컬럼 카운트(4)가 편집 후에도 양 표면에서 상호 일치.

## §4 접근 (요약)

Block 1 6-block Canonical Format을 확장한다(개념 스켈레톤 — 실제 편집은 run-phase):

```
ultrathink.[ ultracode]            ← ultracode는 mode=dynamic-workflow일 때만 opener에 부가(붙여넣기 발화 키워드)
mode: <value>                      ← mode ≠ solo-sequential일 때만; value ∈ {parallel-subagents|agent-team|dynamic-workflow} → Phase 0.95 Mode 4/3/6
# /goal <completion-condition>     ← next phase=run AND machine-verifiable end-state일 때만 (자율 run 진입 미승인)
applied lessons: ...
source_session_id: ...
```

배치 순서 불변식: `ultrathink.`(opener) → `mode:`(옵션) → `# /goal`(옵션) → `applied lessons:` → `source_session_id:`. Block 5 run 커맨드는 mode=agent-team일 때 `--team` 부가. `/effort ultracode` 슬래시 라인은 default 스켈레톤에서 제거되고 "session-persistence 변형"으로만 별도 문서화. 상세 편집 계획은 plan.md 참조.
