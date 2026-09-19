---
id: SPEC-LATE-BRANCH-REDESIGN-001
title: "Late-branch 재설계 — plan 워크트리 진입과 SPEC 당 PR 1개"
version: "0.1.0"
status: draft
created: 2026-09-19
updated: 2026-09-19
author: GOOS
priority: P1
phase: "v3.1.5"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "late-branch, worktree, pr-route, frozen-amendment, spec-workflow"
tier: M
---

# SPEC-LATE-BRANCH-REDESIGN-001 — Late-branch 재설계

## §A 배경

### A.1 무엇이 문제인가

`Route B`(PR 경로)의 절차 문서가 **단계마다 PR 을 하나씩** 여는 체제를 규정한다. 그 결과 하나의 SPEC 이 plan·run·sync 세 개의 PR 을 낳고, 각 PR 이 독립적으로 리뷰·머지된다. 이 체제는 두 가지를 강제한다.

1. **plan 단계는 main 체크아웃에서 수행해야 한다** — 워크트리 사용이 금지된다. 공유 체크아웃의 브랜치 상태를 건드리지 않으려면 plan 산출물을 어디에 둘지가 애매해지고, 병렬 세션이 같은 트리를 본다.
2. **폐기 조건이 두 PR 에 묶인다** — 워크트리는 "run PR 과 sync PR 이 **모두** 머지된 뒤"에만 폐기할 수 있다. 단계 사이에 폐기하면 다음 단계가 깨진다.

### A.2 왜 지금 바꾸는가

런처가 워크트리 진입을 제공하면서(`moai cc -w <name>` · `EnterWorktree(<path>)`) plan 단계를 격리된 트리에서 수행할 수 있게 됐다. 그러면 위 두 강제가 모두 불필요해진다 — plan 커밋이 공유 체크아웃에 닿지 않으므로 브랜치 상태 오염이 없고, PR 이 SPEC 당 하나면 폐기 조건도 하나가 된다.

### A.3 이 SPEC 의 성격

**절차 문서를 고치는 SPEC 이며 코드를 옮기지 않는다.** 다만 대상 문장 중 둘이 `[ZONE:Frozen]` 으로 등재돼 있어 문서 수정만으로는 닫히지 않는다 — `moai constitution amend` 의 5단 게이트를 거치는 개정이 필요하다.

## §B 범위

### B.1 In Scope

| 항목 | 대상 | 현재 상태 |
|---|---|---|
| **B1** | `spec-workflow.md` Route B 표 — 단계 당 PR → **SPEC 당 PR 1개** | 살아 있음 |
| **B2 / OD-1** | `delivery.md` Step 3.3.5 "Return to Base Branch" 은퇴 · plan 단계 워크트리 진입 | 살아 있음 |
| **B6** | `worktree-integration.md` 의 PR 전제 문장 정렬 | 살아 있음 |
| **B3** | 위 변경이 건드리는 Frozen 조항 2건을 amend 5단 게이트로 개정 | 개정 수단 |

### B.2 Out of Scope

- **`REQ-WBG-011`** — 이 트리에서 **소멸**을 확인했다(적중 0). 근거가 사라졌으므로 이 SPEC 이 다루지 않는다.
- **amend 적용 함수 구현** — 이미 구현돼 있다(`internal/constitution/apply_commit.go` · `apply_transform.go`, `TestApply_ExactlyOnce_Success` 외). 이 SPEC 은 그 기능을 **사용**할 뿐 만들지 않는다.
- **`internal/` 아래 코드 변경** — 이 SPEC 은 절차 문서와 zone-registry 항목만 건드린다.
- **docs-site 4로케일 문서** — 절차 변경이 착지한 뒤 별도 문서 카드가 소유한다.

### B.3 Out of Scope — 같은 이름이지만 다른 것

- 이 저장소의 **git-flow 레인 체제**(카드 PR 없음, develop 병합, 리드 일괄 push)는 이 SPEC 의 대상이 아니다. 로컬 전용 규칙(`gitflow-lane-protocol.md`)이 소유하며, 여기서 다루는 것은 **배포되는 기본 체제**의 Route B 다.
- `moai worktree` 의 L1/L2 계층 정의 자체는 바뀌지 않는다. 바뀌는 것은 그 계층이 **언제 폐기 가능한가**의 조건이다.

## §C 요구사항

### REQ-LBR-001 — SPEC 당 PR 1개

**Where** 사용자가 Route B(Tier L 또는 명시적 `--pr`)로 SPEC 을 수행할 때, 절차 문서는 **SPEC 당 하나의 PR** 을 규정해야 한다. 단계(plan·run·sync)마다 별도 PR 을 여는 서술은 남지 않아야 한다.

### REQ-LBR-002 — plan 단계의 워크트리 진입

**Where** Route B 의 plan 단계가 시작될 때, 절차 문서는 런처를 통한 워크트리 진입을 허용해야 하며, "main 체크아웃에서 수행하고 워크트리를 쓰지 않는다"는 금지를 남기지 않아야 한다.

### REQ-LBR-003 — 폐기 조건의 단일화

**While** 워크트리 폐기 조건이 서술될 때, 그 조건은 **하나의 PR 머지**를 기준으로 해야 한다. "run PR 과 sync PR 이 모두 머지된 뒤"라는 두-PR 전제는 남지 않아야 한다.

### REQ-LBR-004 — Step 3.3.5 은퇴

**When** `--pr` 경로가 PR 생성을 마쳤을 때, 기준 브랜치로 되돌아가는 별도 단계(Step 3.3.5)는 수행되지 않아야 한다. 워크트리 체제에서는 되돌아갈 기준 브랜치가 그 트리에 없다.

### REQ-LBR-005 — Frozen 조항의 정규 개정

**Where** 이 SPEC 의 변경이 `[ZONE:Frozen]` 으로 등재된 조항의 문언을 바꿀 때, 그 개정은 `moai constitution amend` 의 5단 게이트(FrozenGuard → Canary → ContradictionDetector → RateLimiter → HumanOversight)를 통과해야 한다. 등재 문언과 문서 본문이 어긋난 채로 남지 않아야 한다.

### REQ-LBR-006 — 두 사본 동반

**While** 대상 문서가 로컬(`.claude/`)과 템플릿(`internal/template/templates/.claude/`) 두 사본을 가질 때, 변경은 두 사본에 함께 적용돼야 한다. 다만 두 사본이 이미 의도적으로 분기한 지점은 그 분기를 보존해야 한다.

## §D 대상 인벤토리 (측정된 좌표)

> 좌표는 `develop @ 0cca34439` 에서 **문구로** 측정했다. 카드가 인용한 줄 번호는 기준 트리(`main`)가 달라 사용하지 않았다 — §G 참조.

### D.1 B1 — `spec-workflow.md`

| 줄 | 문구 | 요구 |
|---|---|---|
| `:26` | "opens a PR per phase (`gh pr create`); phase transitions are triggered by PR merges" | REQ-LBR-001 |
| `:47` | "one squash commit per phase yields clean, revertable SPEC history" | REQ-LBR-001 |
| `:50` | "Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step" | REQ-LBR-002 |
| `:53` | "Step 4 (cleanup) ... ONLY after BOTH run AND sync PRs are merged" | REQ-LBR-003 |

### D.2 B2 / OD-1 — `delivery.md`

| 줄 | 문구 | 요구 |
|---|---|---|
| `:344` | "Step 3.3.5: Return to Base Branch (Post-PR Cleanup)" | REQ-LBR-004 |

`quality-gates-context.md:100` 이 이 단계를 참조한다 — 포인터 정렬 대상.

### D.3 B6 — `worktree-integration.md`

| 줄 | 문구 | 성격 | 요구 |
|---|---|---|---|
| `:45` | L2 정의 — "disposed only via `moai worktree done SPEC-XXX` after both run + sync PRs merge" | 일반 | REQ-LBR-003 |
| `:612` | `[ZONE:Frozen] [HARD]` per-step worktree applicability (spec-workflow.md 가 canonical) | **Frozen** | REQ-LBR-005 |
| `:623` | `[ZONE:Frozen] [HARD]` Disposal contract — "MUST run only after BOTH run PR AND sync PR are merged" | **Frozen** | REQ-LBR-003 · 005 |

### D.4 B3 — `zone-registry.md` Frozen 등재 (축자)

```yaml
- id: CONST-V3R5-027
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/spec-workflow.md
  anchor: "#spec-phase-discipline"
  clause: "Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step"
  canary_gate: true

- id: CONST-V3R5-028
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/spec-workflow.md
  anchor: "#spec-phase-discipline"
  clause: "Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged"
  canary_gate: true
```

**두 항목 모두 `canary_gate: true`** 이므로 개정에 카나리 검증이 붙는다.

### D.5 카드 항목 ↔ Frozen 대응

| 카드 항목 | 바꿀 것 | 걸리는 Frozen |
|---|---|---|
| OD-1 · B2 | plan 도 워크트리로 · Step 3.3.5 은퇴 | **CONST-V3R5-027** |
| B1 | Route B 를 SPEC 당 PR 1개로 | **CONST-V3R5-028** |
| B6 | `worktree-integration` `:45` · `:612` · `:623` 정렬 | `:612` · `:623` |
| B3 | 위 둘의 개정 수단 | — |

### D.6 파일 인벤토리 (2사본 확인 완료)

| 파일 | 로컬 | 템플릿 |
|---|---|---|
| `rules/moai/workflow/spec-workflow.md` | OK | OK |
| `skills/moai/workflows/sync/delivery.md` | OK | OK |
| `rules/moai/workflow/worktree-integration.md` | OK | OK |
| `rules/moai/core/zone-registry.md` | OK | OK |

## §E 자체 검증 전략

각 마일스톤은 **수정 전 적색 관측 → 수정 → 초록 관측**을 남긴다. 절차 문서 변경이라 컴파일 실패로 드러나지 않으므로, 검증은 (1) 문구 잔존 측정, (2) Frozen 등재 문언과 본문의 일치, (3) 두 사본 diff 로 수행한다.

빈 스윕 금지: 모든 문구 프로브는 같은 회차에 **양성 대조**(존재가 확인된 문자열)를 함께 잡아 관측기가 살아 있음을 보인다.

## §F 제외 근거 기록

### F.1 REQ-WBG-011 — 소멸

```
$ grep -rn "REQ-WBG-011" .claude .moai/docs internal/template/templates | wc -l
0
```

카드도 "근거 소멸"로 이미 인지한 항목이다. 소멸했으므로 이 SPEC 의 요구사항으로 옮기지 않는다.

### F.2 amend 적용 함수 — 구현 완료

```
$ ls internal/constitution/
amendment.go  apply_commit.go  apply_transform.go  apply_test.go  ...
```

`moai constitution amend` 는 `internal/cli/constitution.go:448-532` 에 5단 게이트로 배선돼 있다. 이 SPEC 은 사용자이지 구현자가 아니다.

## §G 이 SPEC 의 좌표가 카드와 다른 이유

카드는 `2026-09-10` 에 발행됐고 그 근거 보고서는 `main` 기준이다. 작업 트리인 `develop` 과는 기준이 다르므로 **카드가 인용한 줄 번호를 그대로 쓸 수 없다.** 실제로 두 번 어긋났다.

1. 카드의 `worktree-integration.md 543-556` 은 develop 에서 **워크트리 가드 거부 형태 카탈로그**를 가리킨다. 실제 대상은 `:45` · `:612` · `:623` 이다.
2. 카드의 한국어 문구 `기준 브랜치 복귀` 로 grep 하면 **0 적중**이다. 문서의 실제 문구는 영어(`Return to Base Branch`)이며, 같은 회차의 `Step 3.3.5` 프로브가 4 적중을 내어 모순이 드러났다.

따라서 §D 의 좌표는 전부 **문구·식별자로 재측정한 값**이며, 줄 번호는 그 결과의 기록이지 입력이 아니다.

## §H 선행 카드 인용의 정정

t622 SPEC §D 274-275 는 이 범위를 「카드 t658」로, 그 선행(B4)을 「카드 t659」로 위임했다고 적는다. **그 인용은 현재 큐의 t658·t659 와 다른 것을 가리킨다.**

```
$ git log develop --grep=t658 --oneline | wc -l
12
```

12건이 전부 `SPEC-TODO-QUEUE-HOME-CANON-001`(home-canonical queue) 의 배달이다 — `033323529` · `d26091f5b` · `da3e72ea5` · `a2590d81a` · `68246a364` 등. **late-branch 재설계 산출물은 한 건도 없다.**

`archived` 는 「수행됨」을 뜻하지 않는다. 닫힌 카드가 *무엇을 배달했는지* 는 `git log --grep=<카드id>` 한 번으로 확인된다. 이 기록을 남기는 이유는 다음 사람이 「t658 이 이미 했다」는 추론으로 이 SPEC 을 닫지 못하게 하기 위함이다.

---

🗿 MoAI
