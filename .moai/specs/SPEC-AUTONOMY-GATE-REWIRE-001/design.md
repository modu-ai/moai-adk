# design.md — SPEC-AUTONOMY-GATE-REWIRE-001

## §1. 설계의 뼈대 — 추가형 블록 + 단일 SSOT

두 제약이 설계를 정한다.

1. **guided 바이트 보존.** 배포 기본값 `guided` 에서 오케스트레이터가 읽는 지시문이 바뀌면 안 된다. 기존 문장을 고치면 그 보장을 기계로 확인할 방법이 없다. 그래서 **기존 텍스트는 한 바이트도 고치지 않고**, contract 모드 지시는 마커로 감싼 블록으로만 덧붙인다. 블록을 걷어내면 기준 파일과 바이트 동일해야 한다(REQ-GR-002) — 이 성질 하나로 「guided 경로 불변」을 스크립트 한 줄로 판정한다.
2. **always-loaded 예산.** t1175 가 always-loaded 표면을 줄이는 중이다. 계약 모드의 본문을 always-loaded 파일에 쓰면 그 작업을 거스른다. 그래서 **본문은 path-scoped 신규 규칙 하나(SSOT)** 에 두고, always-loaded 파일에는 한두 문장 포인터만 둔다.

### §1.1 마커 규약

```text
<!-- moai:contract-mode-start id="<slug>" -->
Where `workflow.autonomy.mode: contract` — <지시>. See `.claude/rules/moai/workflow/contract-autonomy.md` § <절>.
<!-- moai:contract-mode-end -->
```

- 두 마커는 각각 **한 줄 전체**를 차지한다. 블록 안쪽의 빈 줄은 허용, 블록 **바깥에 새 줄을 만들지 않는다** — 걷어낸 결과가 기준과 바이트 동일해야 하기 때문이다. 실무 규칙: 블록은 기존 빈 줄 **바로 뒤에** 시작 마커를 두고, 끝 마커 바로 다음 줄은 원래 있던 줄(대개 다음 절 제목이나 문단)이다. 블록 본문과 끝 마커 사이의 빈 줄은 블록 안에 둔다. CommonMark 에서 `<!-- … -->` 가 한 줄에 닫히는 HTML 블록은 그 줄에서 끝나므로 마커 다음 줄의 제목·문단은 정상적으로 렌더된다.
- `<slug>` 는 `[a-z0-9-]+`, 파일 안에서 유일하다.
- 첫 문장은 적용 조건 `workflow.autonomy.mode: contract` 를 명시한다(REQ-GR-003).
- 블록은 `moai:evolvable-start/end` 구간 **밖**에 둔다(REQ-GR-001). `internal/merge/evolvable_zone.go` 의 정규식은 `moai:evolvable-start` 만 잡으므로 새 마커와 충돌하지 않지만, evolvable 구간 **안**에 두면 사용자 진화 구간으로 보존되어 `moai update` 가 블록을 갱신하지 않게 된다.
- 새 마커 이름 `moai:contract-mode-*` 는 기존 마커 종류(`moai:evolvable-*` 85건, `moai:harness-*` 4건, `moai:learned-*`)와 겹치지 않는다(`research.md §2`).

걷어내기 정의(AC 가 쓰는 것과 동일):

```bash
awk '/^<!-- moai:contract-mode-start id="[a-z0-9-]+" -->$/{s=1} !s{print} /^<!-- moai:contract-mode-end -->$/{s=0}' FILE
```

시작 마커부터 끝 마커까지(양 끝 포함)를 지운다. 이 정의가 바이트 보존의 판정식이므로, 삽입 규칙은 이 정의로 지웠을 때 원본이 복원되도록 정해져 있다. 블록 추출(동등 비교용)은 같은 범위만 출력한다:

```bash
awk '/^<!-- moai:contract-mode-start id="[a-z0-9-]+" -->$/{s=1} s{print} /^<!-- moai:contract-mode-end -->$/{s=0}' FILE
```

## §2. 편집 대상 집합 (정본)

모든 행은 로컬 경로이며, 같은 경로를 `internal/template/templates/` 접두 아래에도 **같은 커밋에서** 편집한다. 블록 문자열은 두 사본에서 바이트 동일하다(REQ-GR-070).

| # | 파일 | 블록 id | 게이트 | always-loaded |
|---|---|---|---|---|
| 1 | `.claude/rules/moai/workflow/contract-autonomy.md` | (신규 파일 전체) | 전부 — SSOT | 아니오 (path-scoped) |
| 2 | `CLAUDE.md` | `contract-signing-pipeline` (§2 ④ 뒤), `contract-safe-dev` (§7 Rule 5 뒤) | G1, G2, G3 | **예** |
| 3 | `.claude/rules/moai/core/askuser-protocol.md` | `contract-ambiguity` (§ Ambiguity Triggers and Exceptions › The Five Exceptions 뒤) | G2 | **예** |
| 4 | `.claude/rules/moai/workflow/goal-directive.md` | `contract-signing-goal` (§ Goal-Presentation Timing 끝) | G1 | **예** |
| 5 | `.claude/rules/moai/workflow/orchestration-mode-selection.md` | `contract-signing` (머리의 `[ZONE:Frozen] [HARD]` 문단 뒤, `> Cross-reference:` 인용 앞) | G1 | 아니오 |
| 6 | `.claude/skills/moai/SKILL.md` | `contract-signing-router` (goal 항목 Progression mode 줄 뒤) | G1 | 아니오 |
| 7 | `.claude/skills/moai/workflows/moai.md` | `contract-pipeline-gates` (§ Pipeline Gates 목록 뒤), `contract-merged-round` (Step 11.3 뒤) | G1, G7/G8 | 아니오 |
| 8 | `.claude/skills/moai/workflows/plan.md` | `contract-clarification` (§ [NEEDS CLARIFICATION] Marker Usage 끝) | G1, G9 | 아니오 |
| 9 | `.claude/skills/moai/workflows/plan/spec-assembly.md` | `contract-draft` (Phase 10 끝), `contract-signing-review` (Step 2.3.3b 끝), `contract-audit-retry` (Step 2.3.5 끝), `contract-quality-gate` (Phase 15 HUMAN GATE 끝) | G1, G7, G8 | 아니오 |
| 10 | `.claude/skills/moai/workflows/run.md` | `contract-signing-run` (§ Run-phase Autonomy › 1. 끝), `contract-lifecycle-run` (§ Run-phase Autonomy 절 끝, § Recursive Self-Diagnosis Loop 앞) | G1, 생명주기 1~4 | 아니오 |
| 11 | `.claude/skills/moai/workflows/goal.md` | `contract-progression` (§ Progression Mode 끝) | G1 | 아니오 |
| 12 | `.claude/skills/moai/workflows/sync.md` | `contract-sync-gates` (§ HUMAN GATE Map 끝) — G11 + 생명주기 5~7 | G11, 생명주기 5~7 | 아니오 |
| 13 | `.claude/skills/moai/workflows/sync/doc-execution.md` | `contract-doc-scope` (Step 1.6 선택지 목록 뒤 — `gate-sync-2` evolvable 구간 밖) | G11 | 아니오 |
| 14 | `.claude/skills/moai/workflows/sync/delivery.md` | `contract-next-steps` (§ Context-Aware Next Steps 머리), `contract-error-flow` (§ Error Flow 끝) | G11, Push | 아니오 |
| 15 | (조건부, NC-2) `.claude/skills/moai/workflows/plan/clarity-interview.md` | `contract-interview` | G2 | 아니오 |
| 16 | `internal/template/contract_mode_blocks_test.go` (신규, 템플릿 사본 없음) | — | 가드 (REQ-GR-072) | — |
| 17 | `internal/contract/receipt/` (신규 패키지 — 저장소·해시 체인·검증) + 테스트 | — | REQ-GR-024 | — |
| 18 | `internal/contract/kickoff/` (신규 패키지 — 전제조건 평가·합의 규칙·decide) + 테스트 | — | REQ-GR-016~019 | — |
| 19 | `internal/contract/revoke/` (신규 패키지) + 테스트 | — | REQ-GR-091·092 | — |
| 20 | `internal/cli/contract_decide.go`, `internal/cli/contract_revoke.go` + 테스트 (A1 의 `contract` Cobra 명령에 하위 명령 추가) | — | REQ-GR-019·091 | — |

집계: 문서 로컬 14 + 템플릿 14 + 가드 테스트 1 = 29 파일(조건부 포함 31), 여기에 Go 코드 네 묶음(17~20행). 패키지 경계는 A1 초안의 Package Layout(`internal/contract/` 순수 코어, `internal/contract/sign/`, `internal/cli/contract.go`)을 따른다 — **[A1 감사 통과본으로 재확인]**. 발화 지점 10개(`research.md §1.2` 의 E) 전부가 2~11 행 중에 있다.

**편집하지 않는 파일**(REQ-GR-080): `moai-constitution.md`, `zone-registry.md`, `.claude/agents/**`, `.claude/output-styles/**`, `ci-autofix-protocol.md`, `context-window-management.md`, `agent-common-protocol.md`, 그리고 `research.md §1.2` 의 참조 지점 R 전부.

## §3. SSOT — `contract-autonomy.md` 의 내용 개요

path-scoped 규칙. frontmatter 초안:

```yaml
---
description: "Contract-mode gate rewiring — which MoAI gates a signed SPEC contract replaces, which it never touches, and the one-pass lifecycle"
paths: "**/.moai/specs/**/contract.yaml,**/.moai/config/sections/workflow.yaml,**/.claude/skills/moai/workflows/run.md,**/.claude/skills/moai/workflows/sync.md,**/.claude/skills/moai/workflows/plan.md,**/.claude/skills/moai/workflows/moai.md"
---
```

self-keyed 항목(`**/contract-autonomy.md`)만으로 구성하지 않는다 — t1175 의 domain-keyed 요건과 같은 이유다.

절 구성(영어로 작성 — 규칙 본문 언어 정책):

1. **Scope and activation** — 적용 조건 `workflow.autonomy.mode: contract`. 키 없음·무효값 → guided(A1 설정 판독 규칙과 일치). `MOAI_AUTONOMY_TIER`(권한 번들 축)와는 다른 축이라는 한 줄. 모드 판독 방법(NC-1).
2. **The signing gate (replaces Implementation Kickoff Approval)** — plan-audit PASS/PASS-WITH-DEBT → 오케스트레이터가 서명 명령과 요약을 보고하고 턴을 닫음 → 운영자가 대화형 터미널에서 `moai contract sign <SPEC-ID>` → 다음 턴에 `moai contract verify <SPEC-ID>` exit 0 을 확인하고 run 진입. verify 실패 시 `reasons` 를 담아 에스컬레이션. 사람의 개입은 사라지지 않고 서명 시점으로 옮겨진다(`cadence-bridge.md` 의 human-only 성질과 정합).
3. **Equivalence clause** — 「Implementation Kickoff Approval 이 통과했다/얻었다」를 전제로 하는 모든 참조는 contract 모드에서 검증된 서명으로 충족된다. Phase 4 모드 선택, `sweep` 능력 게이트의 「Kickoff 통과 + 선호 수집 완료」 기록, design 단계 진입, manager-develop 의 `draft → in-progress` 전제가 여기에 해당한다. 「선호 수집」은 계약의 필드가 대신한다.
4. **Gate disposition table** — 게이트 이름별 처분(guided / contract). 소크라테스 인터뷰, 접근 승인, 가정 확인, plan-audit FAIL, SPEC 품질 게이트, sync 확인 질문.
5. **Gates a contract never touches** — 질문 채널 독점, CI autofix 3회 후 질문, 컨텍스트 한도 `/clear`, 헌법, sync-auditor must-pass, main/release(스키마 금지 토큰), 카드 선택, goal 상한, 파괴적 명령 확인, Report-Before-Ask.
6. **Escalation routing** — 에스컬레이션은 질문이 아니라 보고다. Report-Before-Ask 형식 재사용. `escalate_on` 6종과 이 SPEC 이 추가하는 라우팅 사유(verify 실패, 감사 재시도 상한, 두 번째 리뷰 미수행, 미해결 `[NEEDS CLARIFICATION]`). 감지는 A2 소관이라는 경계.
7. **One-pass lifecycle** — Discovery → RED → GREEN → Qualification → Closure → Integration → Push. 단계별 입력·증거·진행 조건 표. Push 조건: `push-develop` ∈ `actions` ∧ `workflow.autonomy.contract.push_develop: true` ∧ 통합 창 보유 ∧ (`second_review: required` 이면 리뷰 증거 존재).
8. **Guided mode** — 이 파일 전체가 guided 에서 비활성이라는 선언. A1 서명기의 guided 안내(서명해도 Kickoff 대체 안 함)와 일치.
9. **Autonomous Kickoff** — 결정자 설정(`human` 기본 / `llm+jev`), 활성 조건 네 가지와 활성 순서(§7.1), 여섯 전제조건, 합의 규칙, 작성자 배제, 「측정되지 않음 → 사람」, Jev 단독·LLM 단독 불허. 오케스트레이터의 절차: 전제조건 판단은 moai 가 한다 → 주 LLM 은 계약 줄별 판단을 판단 파일로 쓴다 → `moai contract decide` 가 Jev 를 부르고 영수증을 발급한다 → `outcome: approve` 면 A1 의 영수증 기반 서명 경로로 서명한다, 아니면 에스컬레이션 보고. 어느 단계에도 `AskUserQuestion` 은 없다.
10. **Revocation** — `moai contract revoke <card>` 의 효과(에스컬레이션 기록 1건 → 다음 단계 경계에서 정지)와 하지 않는 일.

## §4. always-loaded 예산

| 파일 | 현재 `wc -m` (ca1d5dc43) | 이 SPEC 블록 상한 |
|---|---|---|
| `CLAUDE.md` | 19,367 | 두 블록 합 600자 |
| `askuser-protocol.md` | 24,466 | 450자 |
| `goal-directive.md` | 6,875 | 450자 |
| **합계 증가 상한** | — | **1,500자** |

- 파일당 40,000자 한도(t1175 REQ 축)는 셋 모두 넉넉하다. 합계 1,500자는 t1175 의 감축 목표(`96,943자`) 대비 1.5% 이며, 포인터 3개로 정당화되는 최소치로 잡았다. 수치의 적정성은 NC-6.
- 측정 명령은 `LC_ALL=en_US.UTF-8 wc -m` — `LC_ALL=C` 는 바이트를 센다.
- 상한은 t1175 흡수 **후** `BASE` 에서의 크기 대비 증가분으로 잰다.

## §5. 검증 기제

1. **guided 바이트 보존** — 편집 대상 13개 기존 파일 × 2 사본 = 26개 각각에 §1.1 걷어내기를 적용하고 `git show "$BASE:<path>"` 와 `cmp`. 전부 동일해야 한다.
2. **로컬↔템플릿 블록 동등** — 각 파일 쌍에서 블록 추출(걷어내기의 역) 결과를 `diff`. 신규 SSOT 는 `cmp`.
3. **승계 분기 불변** — 각 쌍에서 `diff <(git show BASE:local) <(git show BASE:tmpl)` 와 `diff <(strip local) <(strip tmpl)` 가 같다.
4. **Frozen·제외 파일 불변** — `git diff --quiet "$BASE" -- <REQ-GR-080 제외 목록>` exit 0. `moai constitution validate` 가 기준선과 같은 `OK` 줄.
5. **발화 지점 전수** — `research.md §1.2` 의 E 10개 파일 각각에 블록 ≥ 1. BASE 에서 Kickoff 를 가진 파일 집합 = E ∪ R ∪ H ∪ 로컬 전용 — 차집합이 비어야 한다(t1175 이후 목록이 바뀌면 M0 에서 재분류).
6. **생명주기 순서** — run.md 블록에서 네 단계명, sync.md 블록에서 세 단계명, SSOT 에서 일곱 단계명이 순서대로 나타난다.
7. **상시 가드** — `internal/template/contract_mode_blocks_test.go` 가 `templates/` 아래 모든 `.md` 에 대해: 시작·끝 마커 짝 맞춤, 중첩 없음, evolvable 구간 밖, 블록 안에 금지 클래스(SPEC ID·REQ/AC 토큰·카드 id·날짜) 없음을 검사한다. 로컬 사본과의 대조는 테스트가 아니라 AC 스크립트가 한다(로컬 트리는 배포 산출물이 아니다).

가드 테스트의 RED 는 「마커가 없어 검사 대상이 0건」이 아니라 **알려진 불량 입력**으로 본다 — 짝이 안 맞는 픽스처, evolvable 구간 안의 블록 픽스처, SPEC ID 를 담은 블록 픽스처(`verification-completeness.md §1.1`). 또한 검사 대상 블록이 0개면 테스트가 실패하도록 해 공허한 초록을 막는다(같은 규칙 §1.1 「empty sweep」).

## §6. 결정과 잔여 위험

| 결정 | 근거 | 잔여 위험 |
|---|---|---|
| 기존 텍스트 무수정, 블록 추가만 | guided 불변을 기계로 판정할 유일한 방식 | guided 세션도 블록 문장을 읽는다. 조건문이 첫 문장에 있어도 모델이 오독할 가능성은 0 이 아니다 → 블록 첫 문장을 조건으로 시작하는 규칙으로 완화 |
| `orchestration-mode-selection.md` 머리의 `[ZONE:Frozen] [HARD]` 문단은 두고 등가 블록만 추가 | 운영자 결정 A-Q1; zone-registry 미등록 확인 | plan-auditor 가 이를 「Frozen 조항 재해석」으로 볼 수 있다. 문단 본문은 무수정이며 결정 A-Q1 이 근거라는 점을 SSOT 에 적는다 |
| `moai-constitution.md` 무수정 | 헌법 개정은 범위 밖 | G4 원문이 여전히 「wait for user confirmation」을 말한다. contract 모드의 예외는 SSOT·CLAUDE.md 블록에만 있다 → NC-3 |
| 에이전트 파일 무수정 | Codex 사본 재방출(`make agents-emit`) 부담, 참조 지점일 뿐 | manager-develop 이 전제를 「Kickoff」로 읽는다. 오케스트레이터가 위임 프롬프트 Section A 에 「contract signature verified」를 적어 전달하도록 SSOT 에 명시 |
| 초안 `contract.yaml` 지시를 plan 워크플로 위임문에 싣는다 | A1 이 A3 로 넘긴 일; 에이전트 무수정 원칙 | manager-spec 에이전트 본문에는 아무 말도 없다 → NC-5 |
| SSOT 를 path-scoped 로 | t1175 예산 | 오케스트레이터가 해당 경로를 읽지 않은 채 게이트에 도달하면 SSOT 가 로드되지 않는다 → always-loaded 포인터 3곳이 경로를 명시해 읽게 한다 |
| 에스컬레이션을 「카드 → needs-decision」로 기술 | 설계 문서 | `moai todo` 에 needs-decision 상태가 없다(`moai todo --help` 에 해당 동사·상태 없음, `internal/` Go 소스에 `needs-decision` 0건). **v0.2.0 해소**: 리드 결정으로 needs-decision 은 큐 상태가 아니라 A2 형식의 에스컬레이션 기록 존재다. 이 SPEC 은 그 기록을 읽고(전제조건 e, 단계 경계) revoke 가 1건 쓴다 — 형식은 A2 소관 **[A2 스키마 재확인]** |
| 결정자 쌍을 영수증으로 묶는다 | 에이전트가 쓴 JSON 은 위조 가능(A1 이 제기한 잔여 위험) | 완전한 방지는 불가능. 체인·Jev 원문·저장소 밖 영수증 거부로 흔적만 남긴다(§8.3) |
| decide 는 결과와 무관하게 exit 0 | `SPEC-JEV-CORE-001` REQ-JEVC-007(「Jev 사용 불가의 소비자는 exit 0 으로 계속」) | 호출자가 exit 코드만 보고 승인으로 오독할 위험 → 서명 경로는 `outcome == approve` 인 저장소 기록만 받는다(A1 요구) |

## §7. Kickoff 자율 승인

### §7.1 활성 조건과 순서

결정자 설정이 `llm+jev` 여도 다음 넷이 모두 참일 때만 활성이고, 아니면 사람 서명 경로다(REQ-GR-016).

| 순서 | 조건 | 소유 | 확인 방법 |
|---|---|---|---|
| 1 | `moai contract revoke` 사용 가능 | A3 (이 SPEC) | `moai contract revoke --help` exit 0 |
| 2 | moai 발급 영수증 — `moai contract decide` + 저장소 | A3 (이 SPEC) | `moai contract decide --help` exit 0, 저장소 체인 검증 통과 |
| 3 | A1 의 영수증 기반 서명 경로(저장소 기록만 인정) | A1 **[A1 감사 통과본으로 재확인]** | A1 이 정하는 명령 형태 |
| 4 | A2 의 push 직렬화 강제 + 에이전트발 대화형 `sign` 차단 | A2 **[A2 스키마 재확인]** | A2 가 정하는 형태 |
| 5 | `SPEC-JEV-CORE-001` REQ-JEVC-012 개정 착지 | 별도 카드(NC-7) | 해당 SPEC 의 개정 기록 |

개발 순서도 같다: plan.md 마일스톤은 revoke(M6) → 저장소·decide(M7) → 문서의 자율 Kickoff 블록(M8) 순이며, AC-GR-022 가 커밋 순서로 그것을 확인한다.

### §7.2 역할

| 역할 | 누구 | 하지 않는 일 |
|---|---|---|
| 주 LLM 결정자 | 리드 세션, 또는 새 컨텍스트의 판단 역할(모델은 SPEC 작성자와 다른 것을 선호) | SPEC 을 쓴 에이전트(`manager-spec`)가 아니다 |
| Jev 결정자 | `moai contract decide` 가 직접 호출 | 판단 문장을 생성하지 않는다 — 형식화된 질문에 대한 확률 답 |
| 전제조건 판정 | `moai contract decide` 안의 결정적 코드 | LLM 판단에 맡기지 않는다 |
| 서명 | A1 의 영수증 기반 경로 | 이 SPEC 이 구현하지 않는다 |

`mission-governor` 재사용은 그 에이전트 정의의 「NOT for: … approval」과 충돌한다(`research.md §9.3`). 에이전트 파일 수정은 범위 밖이므로 NC-8 에서 정한다. 결정 전까지 SSOT 는 「리드 세션 또는 새 컨텍스트 판단 역할」로만 적는다.

### §7.3 Jev 질문 형태

Jev 는 한 상태에 대한 형식화된 질문 묶음에 답한다(`internal/jev` `Question`: noul / choice / score). decide 가 보내는 상태는 계약 요약(행동·소유권·예산·수락 기준 해시·전제조건 결과)과 주 LLM 판단의 결론 줄이며, 질문은 「이 계약으로 시작해도 되는가」(noul)와 「가장 약한 줄은」(choice)이다. 비밀 선별·크기 한도·재시도 정책은 `internal/jev` 가 이미 가진 것을 그대로 쓴다 — 두 번째 클라이언트를 만들지 않는다(REQ-JEVC-002).

## §8. `moai contract decide <card>` (가칭)

### §8.1 입력

- `<card>` — 카드 id(`t` + 숫자). 영수증의 키.
- `--spec <SPEC-ID>` — 필수. 카드→SPEC 매핑 저장소가 아직 없으므로(F1) 명시로 받는다.
- `--judgement <file|->` — 주 LLM 판단. JSON: `decider{agent, model, session}`, `verdict`(`approve`/`reject`/`escalate`), 계약 줄별 `reasons[]`. 형식 오류는 exit 2, 기록 없음.
- `--json` — 결과 객체 출력.

### §8.2 절차

1. 저장소 체인을 검증한다. 깨졌으면 exit 1, 아무것도 쓰지 않는다.
2. 입력 파일 해시를 잰다(`contract.yaml`, `acceptance.md`, plan-audit 판정 파일, 판단 파일).
3. 여섯 전제조건을 평가한다(REQ-GR-017). 하나라도 실패하면 Jev 를 부르지 않고 `outcome: human`, `reason: precondition:<a..f>` 영수증을 쓴다.
4. 작성자 배제: SPEC 의 plan-phase 커밋 `Authored-By-Agent:` 트레일러를 읽고 판단의 `decider` 와 비교한다. 같거나 `manager-spec` 이면 `outcome: human`, `reason: author-decider-conflict`.
5. Jev 를 호출한다. 결과가 `available` 이 아니면(disabled / no-credential / unreachable / …) 원시 상태를 기록하고 `outcome: human`, `reason: jev-not-measured`. 신뢰도 < `jev_min_confidence` 도 같다.
6. 합의 규칙(REQ-GR-018)으로 `outcome` 을 정한다.
7. 영수증을 체인에 추가하고, 카드 증거 경로에 사본을 쓴다. exit 0.

### §8.3 영수증 저장소 — 위치 선택

| 기준 | 프라이머리 체크아웃 `.moai/state/…` | `~/.moai/db/<project-key>/contract/` |
|---|---|---|
| 에이전트가 쓰는 워크트리 밖인가 | 워크트리 밖이지만 프라이머리 세션의 에이전트가 늘 쓰는 트리 안 | 모든 프로젝트 트리 밖 |
| 워크트리 폐기·`moai update` 에서 살아남는가 | 살아남음(단 `.moai/` 청소 대상 변동 위험) | 살아남음 |
| 같은 저장소의 모든 워크트리가 같은 곳을 보는가 | 프라이머리 경로를 따로 풀어야 함 | `homestate.ProjectKey` 가 이미 공유 키를 준다 |
| 선례 | — | 큐 `backlog.db`, factory db 가 같은 방식(`internal/homestate`) |
| 권한 | 저장소 기본 | 디렉터리 0700 |

**선택: `~/.moai/db/<project-key>/contract/receipts.jsonl`** (`internal/homestate` 경로 헬퍼 재사용). 근거는 위 표의 첫째·셋째·넷째 줄이다. 운영자 판단이 필요한 성격이 아니라 기존 선례를 따르는 선택이라 NC 로 올리지 않았다.

변조 흔적의 세 겹:

1. **해시 체인** — 각 줄이 `prev` (직전 줄의 sha256)와 자기 `hash` 를 담는다. 중간 줄을 고치면 이후 모든 줄의 검증이 깨진다.
2. **Jev 원문 보관** — 요청·응답 본문을 base64 로 담고 해시를 따로 둔다. 판단 파일만 바꿔 승인을 만들려면 Jev 응답까지 위조해야 한다.
3. **저장소 밖 영수증 거부** — A1 의 서명 경로는 저장소 체인 안의 `approve` 기록 id 만 받는다(A1 요구). 카드 경로의 사본을 고쳐도 서명되지 않는다.

**하지 못하는 것.** 같은 사용자 권한의 행위자는 체인 전체를 새로 쓸 수 있다. 이 SPEC 은 그것을 막지 않는다. 체인 머리 해시를 카드 증거·커밋 메시지에 남기면 사후 대조로 드러나지만, 그것도 흔적이지 방지가 아니다.

## §9. `moai contract revoke <card>`

| 상황 | 동작 | exit |
|---|---|---|
| 카드에 `approve` 영수증이 있고 그 뒤 revoke 없음 | 체인에 revoke 기록 추가 + A2 형식 에스컬레이션 기록 1건 | 0 |
| 이미 revoke 됨(그 뒤 새 `approve` 없음) | 아무것도 쓰지 않음, 「already revoked」 | 0 |
| `approve` 영수증 없음(시작된 run 없음) | 아무것도 쓰지 않음 | 1 |
| 사용법·I/O·체인 무결성 오류 | 아무것도 쓰지 않음 | 2 |

- 옵션: `--reason <text>`(영수증에 기록), `--json`.
- 에스컬레이션 기록 경로·형식은 A2 가 정한다. 리드가 전달한 형태는 `.moai/reports/<card>/escalations/<ts>.json` 이고, A2 초안(`ee9f57151`) design.md §A 는 `.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md` + 색인 줄을 제안한다. 두 형태가 다르므로 A2 감사 통과본을 따른다 **[A2 스키마 재확인]**. A2 의 9개 종류에 revoke 가 없으므로 종류 하나의 추가를 A2 에 요청한다.
- 진행 중인 run 을 강제로 끊지 않는다. 각 단계 경계가 열린 에스컬레이션 기록을 확인하므로(REQ-GR-062) 다음 경계에서 멈춘다. 경계 사이의 작업은 끝까지 간다 — 잔여 위험(NC-9).
- 하지 않는 일(REQ-GR-092): 워크트리 삭제, 브랜치 삭제·개명, push, 큐 변경, `contract.yaml`·서명·SPEC 수정, 프로세스 종료, git 쓰기.
- 누가 실행해도 된다(운영자·에이전트). 안전한 방향의 동작이기 때문이다.

## §10. 테스트 설계 (RED 먼저)

| 대상 | RED 로 관측할 실패 입력 | 녹색 조건 |
|---|---|---|
| 저장소 체인 | 중간 줄 한 글자 변조 픽스처 | 검증이 깨짐을 보고, decide·revoke 가 exit 1/2 로 쓰기 거부 |
| decide 전제조건 | (a)~(f) 각각 하나씩 깬 픽스처 6종 | 각 경우 `outcome: human`, `reason: precondition:<x>`, **Jev 클라이언트 생성 0회**(`internal/cli/mcp_jev.go` 의 생성 이음매 방식) |
| decide Jev 불가 | disabled / no-credential / unreachable 스텁 | exit 0, `outcome: human`, `reason: jev-not-measured`, 원시 상태 기록 |
| decide 신뢰도 | 0.49 스텁 | `outcome: human` |
| decide 합의 | (승인,승인)·(승인,거절)·(에스컬레이션,승인) × `on_disagree` 두 값 | 표대로 |
| 작성자 배제 | 판단 `decider.agent: manager-spec`, 트레일러와 같은 식별자 | `outcome: human`, `reason: author-decider-conflict` |
| revoke 동작 | `approve` 영수증 픽스처 | exit 0, 체인 +1, 에스컬레이션 기록 정확히 1건; 두 번째 호출 exit 0·추가 0 |
| revoke 무영수증 | 빈 저장소 | exit 1, 쓰기 0 |
| revoke 금지 사항 | 워크트리·브랜치·원격·`backlog.db` 가 있는 임시 git 저장소 | 실행 전후 `git worktree list`·`git for-each-ref`·원격 ref·`backlog.db` 바이트 동일; git 실행 이음매에 push 호출 0회 |

모든 테스트는 `t.TempDir()` 와 격리된 `MOAI_HOME` 을 쓴다. 실제 `~/.moai` 에 쓰지 않는다.
