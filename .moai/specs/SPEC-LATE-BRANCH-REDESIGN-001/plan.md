# SPEC-LATE-BRANCH-REDESIGN-001 — 실행 계획

> 기준 트리: `develop @ 0cca34439`. 모든 좌표는 문구로 재측정한 값이다(spec.md §G).

## §A 전제와 순서

### A.1 Frozen 개정이 먼저다

`CONST-V3R5-027` · `-028` 은 `canary_gate: true` 로 등재돼 있다. 본문을 먼저 고치면 **등재 문언과 본문이 어긋난 상태**가 생기고, 그 상태에서 amend 를 돌리면 게이트가 무엇을 기준으로 판정할지 모호해진다. 따라서 순서는 **등재 개정 → 본문 정렬**이다.

### A.2 승인 게이트

[HARD] M2·M3 은 Frozen 개정을 수행한다. **운영자 승인 없이 착수하지 않는다.** 승인은 리드를 통해 온다(레인은 사용자 질문 채널을 갖지 않는다).

M1 과 M5 는 Frozen 을 건드리지 않으므로 승인 전에도 수행 가능하다.

### A.3 단일 writer

이 SPEC 의 대상은 상시로드 규칙과 배포 템플릿이다. 한 트리에 한 writer 로 수행하며 병렬 write 를 하지 않는다.

## §B 마일스톤

### M1 — 비-Frozen 문장 정렬 (승인 불요)

Frozen 등재에 걸리지 않는 문장만 먼저 정렬해 범위를 줄인다.

| 대상 | 변경 |
|---|---|
| `spec-workflow.md:26` | "opens a PR per phase" → SPEC 당 1개 PR 서술 |
| `spec-workflow.md:47` | "one squash commit per phase" → SPEC 당 1개 서술 |
| `worktree-integration.md:45` | L2 정의의 "after both run + sync PRs merge" → 단일 PR 조건 |
| `delivery.md:344` | Step 3.3.5 은퇴 |
| `quality-gates-context.md:100` | Step 3.3.5 참조 제거 — 포인터 정렬 |

두 사본(로컬·템플릿) 동반. 기존 파일을 `cp` 로 통째 미러하지 않는다 — 두 사본이 이미 분기한 지점이 있을 수 있으므로 둘 다 읽고 각각 편집한다.

**RED**: 각 문구가 현재 적중함을 측정(양성 대조 포함).
**GREEN**: 같은 프로브가 0 적중, 그리고 새 문구가 적중.

### M2 — CONST-V3R5-028 개정 (승인 필요)

`Step 4 (cleanup) ... ONLY after BOTH run AND sync PRs are merged` → 단일 PR 조건으로.

1. `moai constitution amend` 로 등재 문언 개정 — 5단 게이트 통과 관측
2. `spec-workflow.md:53` 본문을 개정된 문언에 맞춤
3. `worktree-integration.md:623`(Frozen disposal contract) 정렬

**RED**: 개정 전 등재 문언 축자 기록.
**GREEN**: amend 종료 코드 0 + 게이트 5단 통과 로그 + 등재↔본문 일치 확인.

### M3 — CONST-V3R5-027 개정 (승인 필요)

`Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step` → plan 단계 워크트리 진입 허용으로.

1. `moai constitution amend` 로 등재 문언 개정
2. `spec-workflow.md:50` 본문 정렬
3. `worktree-integration.md:612`(Frozen per-step applicability) 정렬

**RED / GREEN**: M2 와 같은 형식.

### M4 — 교차 정합 확인

개정 후 세 문서가 한 방향인지 확인한다. 특히 **상시로드 규칙**과의 모순을 본다.

| 확인 축 | 대상 |
|---|---|
| PR 개수 | `spec-workflow.md` ↔ `worktree-integration.md` ↔ `delivery.md` |
| 폐기 조건 | `worktree-integration.md:45` ↔ `:623` ↔ `spec-workflow.md` Step 4 |
| plan 위치 | `spec-workflow.md` Step 1 ↔ `worktree-integration.md` SPEC-to-Worktree 매핑 |
| 등재 ↔ 본문 | `zone-registry.md` 두 항목 ↔ `spec-workflow.md` 해당 문장 |

결과를 표로 판정서에 남긴다. **"범위 밖"과 "초록"을 같은 칸에 넣지 않는다.**

### M5 — 두 사본 diff 와 방출물 (승인 불요)

1. 대상 4파일의 로컬↔템플릿 diff 를 측정하고, 남은 차이가 **의도된 분기**인지 판정해 기록
2. 템플릿 변경이 생성물에 미치는 영향 확인 — `make agents-emit-check`
3. catalog 해시 축 확인 — `TestCatalogHashParity` · `TestManifestHashFormat`

> **두 축은 다르다.** `agents-emit-check` 는 `.md → .codex toml` 이고, catalog 해시는 `catalog.yaml` 저장 해시 ↔ `.md` 계산 해시다. 한쪽이 초록이어도 다른 쪽이 적색일 수 있다(t784 실측).

`make build` 는 레인이 돌리지 않는다 — 배치 종료 시 리드가 빌드·embed-check 한다. 이 SPEC 이 남기는 것은 **소스 축 증거뿐**이며, 임베드 축은 Gap 으로 명시한다.

## §C 계기 계약

- 파이프 없이 **파일 리다이렉트 + 종료 코드**를 따로 잡는다. `cmd > file 2>&1; echo "EXIT=$?"` 형태.
- 관측기가 부재를 보고하기 전에 **존재하는 대상을 먼저 잡는다**(양성 대조).
- 프로브는 **영어 원문·식별자**로 잡는다. 한국어 문구와 카드의 줄 번호는 프로브가 아니다(spec.md §G).
- 마크다운·커밋 메시지에 역슬래시-u 이스케이프를 쓰지 않는다.

## §D 되돌리기

| 마일스톤 | 되돌리는 법 |
|---|---|
| M1 · M5 | 문서 변경이므로 커밋 되돌리기로 충분 |
| M2 · M3 | **amend 는 등재를 바꾼다** — 되돌리려면 역방향 amend 가 필요하며, 이것이 승인 게이트를 두는 이유다 |

## §E 진행 기록

착수 전 판정과 측정은 `.moai/reports/t810/verdict.md` 에 있다. 마일스톤 진행은 `progress.md` 가 소유한다.

---

🗿 MoAI
