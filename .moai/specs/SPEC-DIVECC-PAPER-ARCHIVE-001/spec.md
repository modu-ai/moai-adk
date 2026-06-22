---
id: SPEC-DIVECC-PAPER-ARCHIVE-001
title: "VILA-Lab \"Dive into Claude Code\" 논문 아카이브 (CC v2.1.88 internals 권위 참조 + 분산 인용 통합)"
version: "0.1.0"
status: completed
created: 2026-06-22
updated: 2026-06-22
author: manager-spec
priority: P3
phase: "v3.0.0"
module: ".moai/research"
lifecycle: spec-anchored
tags: "research-archive, claude-code-internals, provenance, citation-consolidation, dogfooding, divecc"
era: V3R6
tier: S
---

# SPEC-DIVECC-PAPER-ARCHIVE-001 — "Dive into Claude Code" 논문 아카이브

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-22 | manager-spec | 초기 plan-phase draft. Epic Dive-into-CC 후보 N7 (LOW). 사소(trivial)-archival 전제. 논문 인용은 sibling N5 (SPEC-DIVECC-COMPACTION-LAYER-NAMING-001)에서 이미 VERIFIED-by-citation; 본 SPEC은 in-repo 인용 표면 4곳을 단일 durable 아카이브 뒤로 통합. Go 없음, 동작 변경 없음. |

---

## §A. Background

### A.1 Epic provenance (Dive-into-CC dogfooding)

본 SPEC은 **Epic Dive-into-CC** (domain token `DIVECC`)의 후보 **N7**이다. Epic Dive-into-CC는 Claude Code 내부 구조를 reverse-engineering한 학술 분석 결과를 moai-adk 자신의 harness doctrine에 적용하는 dogfooding(자기개선) 연습이다. Epic roadmap은 `.moai/specs/SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/ROADMAP.md` (§N7 후보 상세)에 있다.

분석 대상 (하나의 출판물, 두 표면):

- **arXiv:2604.14228** — "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (Liu, Zhao, Shang, Shen, 2026, cs.SE). Claude Code v2.1.88 internals reverse-engineering.
- **github.com/VILA-Lab/Dive-into-Claude-Code** — companion repository + "Build Your Own AI Agent: A Design Guide".

논문의 중심 명제는 **"98.4% infrastructure, 1.6% AI"** — agent loop은 사소한 `while`-loop이고, harness가 차별점이라는 것이다. 이 명제 figure와 open-direction 열거는 논문 자신의 주장(Epic의 외부 grounding)으로 기록되며, moai-adk verification claim이 아니다 (ROADMAP Epic Origin note 참조).

### A.2 N7이 해결하는 문제 — 분산된 인용 (scattered citation)

Epic Dive-into-CC의 선행 후보들이 본 논문을 in-repo 여러 표면에 인용했으나, 인용을 묶어두는 **단일 durable 아카이브 항목이 없다**. 현재 4개 canonical 인용 표면(§B.1에서 plan-phase에 Read로 검증)이 각각 arXiv:2604.14228을 인용하지만, 논문 자체의 전모(서지정보, moai-adk가 consume한 CC-internals 내용, 인용 표면 목록)를 한 곳에 보존하지 않는다.

N7은 다음을 한다:

1. `.moai/research/` 아래 **하나의 substantive 아카이브 항목**을 만들어, 논문을 Claude Code v2.1.88 architecture/internals 권위 참조로 보존한다 (서지정보 + moai-adk가 consume한 CC-internals 내용 + 인용 표면 목록).
2. 분산된 canonical 인용들을 단일 durable 아카이브로 연결하는 **하나의 rule cross-reference 라인**을 추가한다.

### A.3 Critical framing boundary — 아카이브이지 재검증이 아니다

[ZONE:Frozen-spirit] **논문 인용은 sibling N5에서 이미 VERIFIED-by-citation 되었다.** 본 SPEC은 arXiv 논문을 WebFetch/WebSearch로 **재검증하지 않는다** — 인용을 established fact로 취급하고, in-repo 인용 표면들이 서로 일관됨만 Read로 확인한다 (§B.1). 아카이브 항목이 기록하는 "98.4% / 1.6%" 같은 figure와 5-layer compaction taxonomy는 **논문 자신의 명제**로 framing되며 moai-adk 측정/행동 주장이 아니다 (verification-claim-integrity invariant §1.1 surface 3과 정합).

### A.4 Anti-over-engineering boundary

[ZONE:Frozen-spirit] 본 SPEC은 사소한 Tier S 아카이브다. 산출물은 정확히 **2가지**: 아카이브 파일 1개 + cross-reference 라인 1개. 가상의 요구사항 발명 금지, template-mirror 작업 제안 금지 (`.moai/research/`는 dev-local — `internal/template/templates/` 아래 없음; §B.1에서 plan-phase에 `find`로 확인됨), 4개 인용 표면을 cross-reference 라인 1개를 넘어 문서 리팩터로 확장 금지. AC는 기계 검증 가능(file-existence check + cross-reference grep)이어야 한다.

---

## §B. Problem Statement & Grounding

### B.1 Grounding (citation established + moai-tree observations this plan-phase)

전제는 **trivial (archival)**이다 — 논문 인용은 N5에서 이미 established. 본 SPEC은 인용을 재검증하지 않고, in-repo 인용 표면 일관성만 Read로 확인한다. 다음 moai-tree 관측을 본 plan-phase(Read/grep, 2026-06-22)에 재현했다:

**관측 — `.moai/research/` 디렉터리 (dev-local 확인):**
- `find internal/template/templates -path '*moai/research*'` → empty. ⇒ `.moai/research/`는 dev-local이며 template-distributed 아님 → mirror 불필요, neutrality 제약 없음.
- 선행 precedent `.moai/research/gears-paper-validation.md` 존재 (10328 bytes ≈ 10KB) — house style/length target.

**관측 — 4개 canonical 인용 표면 (각각 arXiv:2604.14228 인용 확인):**

| # | 인용 표면 | 추가 SPEC | plan-phase grep 확인 |
|---|-----------|-----------|----------------------|
| 1 | `.claude/output-styles/moai/moai.md` | N3 (DELEGATION-TOKEN-COST) | line 142: "Dive into Claude Code" paper (arXiv:2604.14228) ~7× delegation token-cost signal |
| 2 | `.claude/rules/moai/workflow/context-window-management.md` | N5 (COMPACTION-LAYER-NAMING) | line 17: 5-layer compaction naming, arXiv:2604.14228 + github.com/VILA-Lab/Dive-into-Claude-Code |
| 3 | `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §1 | N5 (COMPACTION-LAYER-NAMING) | line 24: "Convergent second source" arXiv:2604.14228 + VILA-Lab |
| 4 | `.claude/rules/moai/development/agent-authoring.md` | N2 (EXTENSION-COST-LADDER) | line 309: extension-mechanism context-cost ladder, arXiv:2604.14228 |

4개 표면 모두 동일 논문(arXiv:2604.14228)을 인용함이 확인되었다 — 인용 간 불일치 없음.

### B.2 논문이 consume된 CC-internals 내용 (아카이브 항목이 기록할 내용)

아카이브 항목은 moai-adk가 이미 consume한 다음 CC-internals 내용을 포착한다 (각 항목은 위 4개 표면 중 하나가 인용한 논문 명제):

- **5-layer graduated-compaction taxonomy**: `Budget Reduction → Snip → Microcompact → Context Collapse → Auto-Compact` (표면 2/3이 consume).
- **AI-agent-system design-space taxonomy**: 논문의 설계 공간 분류 + open-direction 열거 (중심 명제 "98.4% infrastructure").
- **query-loop / withheld-recoverable-error framing**: runtime-recovery-doctrine.md §1이 의존하는 input-governance 사전 단계 (표면 3이 consume).
- **delegation token-cost (~7×) signal**: SkillTool(context 주입, 저렴) vs AgentTool(isolated context spawn, ~7× token) (표면 1이 consume).
- **extension-mechanism context-cost ladder**: Hooks(zero) → Skills(low) → Plugins(medium) → MCP(high) (표면 4가 consume).

### B.3 Run-phase 파일 scope (Tier S 확정)

총 **2개 산출물**:

1. **아카이브 파일 1개** — `.moai/research/<archive-name>.md` (NEW). substantive(bare bibliographic stub 아님); precedent `.moai/research/gears-paper-validation.md`를 구조적으로 모델로 함(house style/length ≈ 10KB). 서지정보 + consume된 CC-internals 내용(§B.2) + 인용 표면 열거(§B.1 표).
2. **cross-reference 라인 1개** — 분산된 canonical 인용들을 단일 durable 아카이브로 연결. 정확한 target은 plan.md에서 제안.

dev-local (no mirror, no neutrality constraint), doc-only, no Go, no behavior change → Tier S 확정.

---

## §C. Requirements (GEARS)

> Notation: GEARS (current). `<subject>`는 일반화됨.

**REQ-PA-001 (Ubiquitous)** — 아카이브 파일 1개가 `.moai/research/` 아래에 **존재해야 한다** (substantive — bare bibliographic stub 아님; precedent `gears-paper-validation.md` house style을 따른다).

**REQ-PA-002 (Ubiquitous)** — 아카이브 항목은 논문의 full bibliographic citation을 **기록해야 한다**: title("Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems"), authors(VILA-Lab), arXiv:2604.14228, companion repo URL(github.com/VILA-Lab/Dive-into-Claude-Code).

**REQ-PA-003 (Ubiquitous)** — 아카이브 항목은 moai-adk가 이미 consume한 논문의 CC-internals 내용(§B.2 — 5-layer graduated-compaction taxonomy, AI-agent-system design-space taxonomy, query-loop/withheld-recoverable-error framing)을 **포착해야 한다**.

**REQ-PA-004 (Ubiquitous)** — 아카이브 항목은 논문을 consume하는 in-repo 표면들(§B.1의 4개 canonical 인용 표면)을 **열거해야 한다**.

**REQ-PA-005 (Ubiquitous)** — 하나의 rule cross-reference 라인이 분산된 canonical 인용들을 단일 durable 아카이브로 **연결해야 한다** (정확한 target은 plan.md에서 제안).

**REQ-PA-006 (Where capability gate / consume-not-implement)** — **Where** 아카이브 항목이 5-layer graduated-compaction taxonomy를 기록하는 경우, 그 항목은 5개 layer를 moai-adk가 **consume하는** Claude Code의 graduated-compaction layer로 framing해야 하며, moai-adk가 구현하는 layer로 **암시해서는 안 된다** (sibling N5의 consume-not-implement boundary와 정합).

**REQ-PA-007 (Unwanted behavior)** — run-phase 편집은 §B.3에 명명된 2개 산출물(아카이브 파일 1개 + cross-reference 라인 1개) 외 어떤 파일도 **수정해서는 안 된다**; `.moai/research/`의 template mirror를 **생성해서는 안 된다** (dev-local); 4개 인용 표면을 cross-reference 라인 1개를 넘어 **리팩터해서는 안 된다**.

**REQ-PA-008 (Unwanted behavior)** — run-phase는 WebFetch/WebSearch로 arXiv 논문을 **재검증해서는 안 된다** (인용은 N5에서 established); 어떤 compaction/recovery/runtime 동작도 **변경해서는 안 된다** (moai-adk는 CC compaction을 consume — 변경 불가); 어떤 Go 코드도 **변경해서는 안 된다** (없음 — doc-only 아카이브 SPEC).

---

## §D. Acceptance Criteria (inline reference — Tier S)

> Tier S이지만 orchestrator 지시에 따라 본 SPEC은 별도 `acceptance.md`를 동반한다. §D는 AC 요지의 inline 참조이며, 완전한 Given-When-Then 시나리오 + edge case + DoD는 `acceptance.md`가 SSOT다. 각 AC는 observable/mechanical (grep / file-existence). 이 AC들은 **run-phase** 결과를 bind하며 plan-phase에서 충족되지 않는다.

| AC | 요지 | bind REQ | 검증 방식 |
|----|------|----------|-----------|
| AC-PA-001 | 아카이브 파일이 `.moai/research/` 아래 존재 + non-stub (≥ 일정 byte) | REQ-PA-001 | file-existence + `wc -c` |
| AC-PA-002 | full bibliographic citation 4요소 present | REQ-PA-002 | `grep -F` 4-term for-loop |
| AC-PA-003 | consume된 CC-internals 내용 present (5-layer taxonomy 등) | REQ-PA-003 | `grep` 5-layer 이름 + framing 키워드 |
| AC-PA-004 | 4개 인용 표면 모두 아카이브에 열거됨 | REQ-PA-004 | `grep -F` 4-path for-loop |
| AC-PA-005 | cross-reference 라인이 단일 아카이브로 연결 | REQ-PA-005 | `grep` archive 경로 in target file |
| AC-PA-006 | consume-not-implement framing co-located with layer name | REQ-PA-006 | `grep -niE` co-location anchor |
| AC-PA-007 | run-commit diff ⊆ {archive file, cross-ref target file}; no rmirror | REQ-PA-007 | `git show --stat` + `find` mirror absent |
| AC-PA-008 | Go 변경 0; arXiv 재검증 흔적 없음 | REQ-PA-008 | `git show --stat` (no `.go`) |

전체 AC 시나리오·검증 명령·edge case는 `acceptance.md` 참조.

---

## §E. Lifecycle Progress Markers

> Plan-phase는 §E 섹션 skeleton(placeholder heading만)을 emit한다. §E.2–§E.4는 manager-develop(run) 및 manager-docs(sync)가 populate하며 본 plan-phase author가 아니다. canonical skeleton은 progress.md 참조.

- **§E.1 Plan-phase Audit-Ready Signal** — progress.md §E.1 참조.
- §E.2 Run-phase Evidence — _pending run-phase_.
- §E.3 Run-phase Audit-Ready Signal — _pending run-phase_.
- §E.4 Sync-phase Audit-Ready Signal — _pending sync-phase_.

---

## §F. Out of Scope

이 섹션은 본 SPEC이 다루지 **않는** 범위를 명시한다 (`OutOfScopeRule` 충족 / `MissingExclusions` 회피).

### Out of Scope — arXiv 논문 재검증

- WebFetch/WebSearch로 arXiv:2604.14228 논문을 재검증·재-fetch하는 작업. 인용은 sibling N5 (SPEC-DIVECC-COMPACTION-LAYER-NAMING-001)에서 이미 VERIFIED-by-citation으로 established이며, 본 SPEC은 인용을 established fact로 취급한다 (REQ-PA-008). 본 SPEC이 하는 일관성 확인은 in-repo 인용 표면을 Read하는 것뿐이다.

### Out of Scope — template mirror 생성

- `.moai/research/` 아카이브 파일의 template mirror를 만드는 작업. `.moai/research/`는 dev-local이며 `internal/template/templates/` 아래에 배포되지 않는다 (plan-phase `find`로 확인). mirror 생성은 명시적으로 out of scope (REQ-PA-007).

### Out of Scope — 4개 인용 표면 문서 리팩터

- 4개 canonical 인용 표면(moai.md / context-window-management.md / runtime-recovery-doctrine.md / agent-authoring.md)을 cross-reference 라인 1개 추가를 넘어 리팩터·재구성·내용 변경하는 작업. 산출물은 cross-reference 라인 1개로 한정된다 (REQ-PA-007). 기존 인용 본문은 손대지 않는다.

### Out of Scope — Go 코드 변경 / 동작 변경

- 어떤 Go 코드 변경도 없다 — doc-only 아카이브 SPEC. 어떤 compaction/recovery/runtime 동작도 변경하지 않는다. moai-adk는 Claude Code의 graduated compaction(Budget Reduction / Snip / Microcompact / Context Collapse / Auto-Compact)을 CONSUME하며 그 layer를 구현·변경할 수 없다 (REQ-PA-006/008).

### Out of Scope — 논문 내용의 독립 검증 / 행동 주장

- 5-layer compaction taxonomy나 "98.4%/1.6%" figure 등 논문 명제를 moai-adk 트리에서 독립 검증·벤치마크·행동 주장하는 작업. 아카이브 항목은 이를 **논문 자신의 명제**(arXiv:2604.14228의 인용)로 기록할 뿐, moai-adk 측정·행동 주장이 아니다 (REQ-PA-003 + verification-claim-integrity invariant §1.1 surface 3).

### Out of Scope — 다른 Epic Dive-into-CC 후보

- N1 (hook failure-mode audit, closed), N2 (extension-cost ladder, closed), N3 (delegation token-cost, closed), N4 (observability loop), N5 (compaction-layer naming, closed), N6 (unified inventory). 각각 자체 SPEC이며 본 SPEC은 그 표면을 건드리지 않는다 (단, N7 아카이브는 N2/N3/N5가 추가한 인용 표면을 **읽어** 열거 목록을 만든다 — 이는 read-only이며 그 표면을 수정하지 않는다).

---

## §G. Cross-References

- **arXiv:2604.14228** — "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (아카이브 대상 논문).
- **github.com/VILA-Lab/Dive-into-Claude-Code** — companion repository.
- `.moai/specs/SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/ROADMAP.md` — Epic Dive-into-CC roadmap (§N7 후보 상세).
- `.moai/research/gears-paper-validation.md` — precedent 아카이브 항목 (house style / length ≈ 10KB 모델).
- `.claude/output-styles/moai/moai.md` — 인용 표면 #1 (N3, delegation ~7× token-cost).
- `.claude/rules/moai/workflow/context-window-management.md` — 인용 표면 #2 (N5, 5-layer compaction naming).
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` — 인용 표면 #3 (N5, convergent second source).
- `.claude/rules/moai/development/agent-authoring.md` — 인용 표면 #4 (N2, extension-cost ladder).
- `.moai/specs/SPEC-DIVECC-COMPACTION-LAYER-NAMING-001/spec.md` — sibling N5 (인용 provenance + era convention 모델).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 — provenance discipline (논문-naming citation vs moai-tree 행동 주장 구분).
