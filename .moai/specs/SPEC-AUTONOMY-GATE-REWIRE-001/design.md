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
| 16 | `internal/template/contract_mode_blocks_test.go` (신규, 템플릿 사본 없음) | — | 가드 (REQ-GR-073) | — |

집계: 로컬 14 + 템플릿 14 + 테스트 1 = **29 파일**(조건부 포함 31). 발화 지점 10개(`research.md §1.2` 의 E) 전부가 2~11 행 중에 있다.

**편집하지 않는 파일**(REQ-GR-090): `moai-constitution.md`, `zone-registry.md`, `.claude/agents/**`, `.claude/output-styles/**`, `ci-autofix-protocol.md`, `context-window-management.md`, `agent-common-protocol.md`, 그리고 `research.md §1.2` 의 참조 지점 R 전부.

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
4. **Frozen·제외 파일 불변** — `git diff --quiet "$BASE" -- <REQ-GR-090 목록>` exit 0. `moai constitution validate` 가 기준선과 같은 `OK` 줄.
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
| 에스컬레이션을 「카드 → needs-decision」로 기술 | 설계 문서 | `moai todo` 에 needs-decision 상태가 없다(`moai todo --help` 에 해당 동사·상태 없음, `internal/` Go 소스에 `needs-decision` 0건). 상태 도입은 A2/F1 소관 → 블록 문구는 「보고를 남기고 멈춘다」까지만 규정하고 상태 이름은 SSOT 에 A2 참조로만 둔다 |
