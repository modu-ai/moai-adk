# \[Feature\] SPEC Review에 Human Decision Authority Guardrail 도입 제안

## Problem Statement

MoAI-ADK를 실제 프로젝트에 적용하면서 SPEC Review 단계에서 두 가지
decision-authority failure를 반복적으로 경험했습니다.

1.  **SPEC이 PRD에서 아직 결정하지 않은 제품 수준의 사항을 조용히
    확정하는 문제**
2.  **사용자가 독립적인 판단을 형성하기 전에 AI recommendation을 먼저
    보게 되는 문제**

현재 개인 프로젝트에서는 이를 다음 원칙으로 다루고 있습니다.

> **AI audits, interrogates, routes, analyses and executes. The founder
> is the only decider.**

핵심은 AI의 reasoning capability를 제한하는 것이 아니라, **문서 안에
정답이 존재하는 문제와 실제 Human judgment가 필요한 문제를 분리하는
것**입니다.

``` text
PRD             product intent, scope, phase boundaries, founder-level decisions
  |
spec.md         implementable requirements
  |
acceptance.md   observable acceptance behaviour
  |
plan.md         implementation strategy
  |
code
```

이 authority chain에서 **SPEC은 PRD가 결정하지 않은 product decision을
임의로 확정하지 않습니다.**

------------------------------------------------------------------------

## Proposed Solution

SPEC Review에 선택적으로 적용할 수 있는 **Human Decision Authority
Guardrail**을 제안합니다.

``` text
Stage 1: Mechanical Verification
        ↓
Stage 2: Adversarial Interrogation
        ↓
Decision Index / Authority Routing
        ↓
Human Decision Gate
        ↓
Stage 3: Apply Confirmed Verdict
```

### 1. Mechanical Verification과 Judgment Point 분리

문서 안에서 정답을 기계적으로 도출할 수 있는 문제는 Stage 1에서
처리하고, 문서 어디에도 결정 authority가 없는 문제는 Stage 2 judgment
point로 올립니다.

예를 들어:

-   PRD `CONFIRMED`와 SPEC의 직접 모순 → Stage 1
-   PRD `DEFERRED`를 SPEC이 임의로 확정 → Stage 2
-   SPEC이 PRD에 없는 product-level behavior를 도입 → Stage 2
-   mechanical contradiction이지만 수정 방향이 여러 개이고 product
    effect가 달라짐 → `FOUNDER`

Stage 2 Agent는 **질문을 발견하고 설명할 뿐 스스로 resolution을
추천하거나 결정하지 않습니다.**

> **Detect → Explain → Ask, but never decide.**

### 2. `decision-index.md`를 통한 Authority Routing

발견된 모든 flag를 Human에게 그대로 넘기는 대신, `decision-index.md`에서
**importance ranking이 아니라 authority routing**을 수행합니다.

  ------------------------------------------------------------------------
  Type                    Meaning                 Founder action
  ----------------------- ----------------------- ------------------------
  `DECIDED`               이전 Founder verdict가  `NONE`
                          동일 조건을 이미 결정   

  `POLICY-COVERED`        명시적                  `NONE_UNLESS_OVERRIDE`
                          PRD/ADR/engineering     
                          policy가 그대로 적용    

  `EVIDENCE-NEEDED`       결정을 위해 아직 없는   `REQUEST_EVIDENCE` /
                          데이터/측정 필요        `TEMPORARY_VERDICT` /
                                                  `DEFER`

  `FOUNDER`               기존 authority가 답하지 `DECIDE` /
                          못함                    `NEED_ANALYSIS` /
                                                  `NEED_EVIDENCE` /
                                                  `DEFER`
  ------------------------------------------------------------------------

`DECIDED`와 `POLICY-COVERED`는 repository 안에 실제 authority와 정확한
source/section이 존재하고 applicability가 명확한 경우에만 허용합니다.

> **An LLM "best practice" is not a policy.**

판단이 애매하면 자동으로 implementation detail로 낮추지 않습니다.

> **When uncertain, escalate. Never downgrade.**

### 3. Analysis is Pull, Not Push

AI recommendation에 Human 판단이 anchoring되는 것을 줄이기 위해
analysis와 recommendation도 기본적으로 push하지 않습니다.

``` text
FLAGGED → TRIAGED → FOUNDER → INITIAL_JUDGMENT
                              ├─ decide → FINAL_VERDICT
                              └─ unsure
                                   ↓
                              NEED_ANALYSIS
                                   ↓
                              AI ANALYSIS
                                   ↓
                              FINAL_VERDICT
```

Human이 `NEED_ANALYSIS`를 요청해야 AI analysis가 생성되며, single
recommendation은 별도의 `REQUEST_RECOMMENDATION`이 있을 때만 제공합니다.

> **An AI preferred answer must never reach the founder before their
> independent judgment or an explicit request for it.**

### 4. Canonical Human Verdict와 Product Authority 보존

`audit-report.md`, `interrogation-draft.md`, `decision-index.md`,
`analysis/{Qn}.md`, AI recommendation은 decision authority가 아닙니다.

Canonical verdict는 Human이 작성한 `answer.md`에 남깁니다.

또한 verdict를 적용하기 전에:

-   **SPEC-level** --- retry count, internal algorithm, query shape 등 →
    SPEC에 적용
-   **Product-level** --- MVP/phase scope, tier behavior, UX flow,
    pricing, privacy/security promise 등 → PRD와 SPEC을 함께 reconcile

로 구분합니다.

Product-level verdict를 SPEC에만 반영하여 PRD drift를 다시 만드는 것을
방지합니다.

### 5. Traceability와 Zero-Flag Rule

결정의 전체 경로를 파일만으로 재구성할 수 있도록 합니다.

``` text
PRD section → REQ id → interrogation Qn → decision-index Qn
            → analysis/Qn.md (optional) → answer.md Qn
            → changed REQ / scenario / plan → SPEC HISTORY
            → PRD update (product-level verdict only)
```

또한 Stage 2에서 flag가 0개라고 해서 `PASS`로 간주하지 않습니다.

**Zero flags는 "이번 pass에서 model이 judgment point를 발견하지
못했다"는 의미일 뿐 Human approval이 아닙니다.**

------------------------------------------------------------------------

## Alternatives Considered

### SPEC/Planning Agent가 ambiguity를 직접 해결

가장 빠르지만, PRD에서 결정되지 않은 product decision까지 reasonable
default로 확정할 수 있습니다.

### 모든 SPEC을 Human이 직접 재검토

Human authority는 보존되지만 Agentic workflow의 review automation 이점이
크게 줄어듭니다. 제안하는 구조에서는 AI가 audit/interrogation/routing
비용을 부담하고, Human은 실제 authority가 필요한 decision surface에
집중합니다.

### 모든 flag에 AI Recommendation 자동 제공

Human이 최종 승인하더라도 AI의 preferred answer를 먼저 보게 되어
anchoring될 수 있으므로, analysis와 recommendation을 explicit pull
방식으로 제한합니다.

### Stage 1과 Stage 2 통합

Mechanical defect와 unresolved judgment point는 서로 다른 failure mode를
다루므로 역할을 분리합니다.

> `plan-auditor` answers what is settled inside the authoritative
> documents; `spec-interrogator` finds what is not settled anywhere.

------------------------------------------------------------------------

## Use Case

PRD에서 아직 결정되지 않은(`DEFERRED`) product behavior를 SPEC이 하나의
requirement로 확정했다고 가정합니다.

``` text
PRD DEFERRED
     ↓
SPEC resolves it
```

이를 자동 수정하지 않고 Stage 2 judgment point로 올립니다.

`decision-index.md`에서:

-   기존 Founder verdict가 정확히 적용되면 `DECIDED`
-   명시적 policy가 그대로 적용되면 `POLICY-COVERED`
-   추가 데이터가 필요하면 `EVIDENCE-NEEDED`
-   기존 authority가 없으면 `FOUNDER`

로 routing합니다.

`FOUNDER`인 경우 Human은 직접 `DECIDE`하거나, 필요할 때만
`NEED_ANALYSIS`, `NEED_EVIDENCE`, `DEFER`를 선택합니다.

최종 verdict가 Product-level이면 PRD와 SPEC을 함께 reconcile하고,
SPEC-level이면 SPEC에만 적용합니다.

이 구조의 목적은 Human decision을 최대화하는 것이 아니라 **각 decision을
올바른 authority에게 routing하는 것**입니다.

------------------------------------------------------------------------

## Additional Context

현재 개인 프로젝트에서 다음 요소를 포함한 customized SPEC Review
workflow를 실험하고 있습니다.

-   Authority chain
-   Stage 1 mechanical verification boundary
-   Stage 2 adversarial interrogation
-   `decision-index.md` authority routing
-   `DECIDED` / `POLICY-COVERED` / `EVIDENCE-NEEDED` / `FOUNDER`
-   conservative escalation
-   reversibility classification
-   pull-based AI analysis / recommendation
-   canonical Human verdict
-   Product-level verdict의 PRD propagation
-   end-to-end traceability
-   zero-flag ≠ pass

### ⚠️ Compatibility Note

**현재 이 설계와 구현은 MoAI-ADK v3.1.2 이전의 legacy EARS 기반 SPEC
workflow를 기준으로 작성한 것입니다.**

따라서 현재 버전의 SPEC 체계에 기존 구현을 그대로 적용하는 것을 제안하는
것은 아닙니다.

특히 다음 요소는 현행 architecture에 맞게 재설계가 필요할 수 있습니다.

-   EARS-specific terminology 및 requirement structure
-   legacy SPEC artifact layout
-   `plan-auditor` / `spec-interrogator` 역할
-   Stage 1 / Stage 2 / Stage 3 orchestration
-   `docs/review/{SPEC-ID}` artifact structure
-   현행 SPEC workflow와의 integration

이번 Issue에서 우선 제안드리고 싶은 것은 legacy EARS 구현 자체가 아니라,
그 과정에서 사용한 **Human Decision Authority model**입니다.

> **Delegate reasoning, not authority.**

이러한 authority layer가 현재 MoAI-ADK의 SPEC workflow에도 유용한
방향인지 maintainer분들의 의견을 먼저 여쭙고 싶습니다.

방향성이 적절하다면 현재 `develop` 구조에 맞게 재설계하여 기여해보고
싶습니다.
