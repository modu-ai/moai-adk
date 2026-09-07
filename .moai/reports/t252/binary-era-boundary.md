# t252 — era 경계 이동 관측 + `moai spec close` 결함 재현

card: t252
branch: WT-spec-close-backfill
측정 트리: `fce4eff27` (= `1dc29226d` + merge `origin/develop` `65196a5a7`)

## 1. Claim

동일 트리에서 `moai` 바이너리만 교체해 `moai spec audit` 을 두 번 재면,
**드리프트 판정(MUST-FIX)은 변하지 않는데 era 분류가 23건 이동한다** —
`V3R5` → `V3R6`. V3R5 는 grandfather 보호 대상이고 V3R6 은 아니므로,
이 23건은 보호 밖으로 나왔다.

## 2. Evidence

명령: `moai spec audit`

| 항목 | 구 바이너리 `v3.1.2` | 신 바이너리 `v3.1.2-1308-g65196a5a7` |
|---|---|---|
| Total SPECs | 740 | 740 |
| Drift findings | 513 | 513 |
| MUST-FIX | 2 | 2 |
| MUST-FIX 대상 | CI-FLAKE-SERIES-001, FULL-SUITE-DOCTRINE-001 | 동일 |
| Grandfathered | 286 | **263** |
| Modern-era clean | 452 | **475** |

소견 대상 SPEC id 집합은 512개로 **완전히 동일**하다 (old-only 0, new-only 0).
두 실행의 차이는 오직 era 라벨이다.

병합 후 재측정(`fce4eff27`)도 신 바이너리 값과 동일하다
— 740 / 263 / 475 / 513, MUST-FIX 같은 2건. 병합이 드리프트를 만들지 않았다.

### `V3R5` → `V3R6` 재분류 23건 (전수)

```
SPEC-ASTGREP-BREADTH-001              SPEC-ASTGREP-LANG16-001
SPEC-CLOCAL-AUDIT-001                 SPEC-CONFIG-DEAD-SWEEP-001
SPEC-DEVPROT-REQUIRED-001             SPEC-DWF-CODEMAPS-PILOT-001
SPEC-GITFLOW-DOCTRINE-ALIGN-001       SPEC-GLM-EFFORT-REBALANCE-001
SPEC-HARNESS-LEARNING-EVO-001         SPEC-HARNESS-LEARNING-EVO-002
SPEC-HOOK-DISCIPLINE-WIRING-001       SPEC-HOOK-PREEDIT-INVESTIGATE-001
SPEC-INTERNAL-ARCH-001                SPEC-KANBAN-BOOTSTRAP-001
SPEC-KANBAN-PR-CARD-TRACEABILITY-001  SPEC-KANBAN-QUEUE-PR-SYNC-001
SPEC-KANBAN-TODO-CLI-001              SPEC-KANBAN-WORKTREE-001
SPEC-LSPMCP-RETIRE-001                SPEC-PREPUSH-WIRING-001
SPEC-SEC-SCAN-SURFACE-001             SPEC-UPDATE-VERSION-FLAG-001
SPEC-V3R6-SESSION-HANDOFF-AUTO-001
```

지금은 23건 모두 clean 이라 MUST-FIX 로 올라오지 않는다. 다만 앞으로
이들에 드리프트가 생기면 **면제 없이 MUST-FIX 가 된다.** 종결 판별식의
조상성 조건을 쓰는 후속 작업은 이 경계가 움직였다는 사실을 전제로 삼아야 한다.

## 3. Baseline-attribution

- 구 값: `.moai/reports/t252/spec-audit-after.json` (`audited_at` `2026-09-02T05:54:50Z`)
- 신 값: 같은 워크트리에서 `moai spec audit` 재실행 (`~/go/bin/moai` = `65196a5a7` 빌드, `built 2026-09-02T05:53:08Z`)
- 두 실행 사이 이 트리에 커밋 없음 — 트리 축은 통제됨

## 4. Gaps

- **귀속에 ~100초 모호 구간이 있다.** 신 바이너리의 *빌드* 시각(`05:53:08Z`)과
  구 값의 *감사* 시각(`05:54:50Z`) 이 겹친다. 빌드 시각과 설치 시각은 다르므로,
  구 값이 정말 구 바이너리로 잰 것이라는 보장은 시각만으로는 서지 않는다.
  라벨 이동의 원인이 바이너리라는 것은 **강한 추정이지 증명이 아니다.**
  다만 트리 축이 배제되므로 남은 후보는 바이너리뿐이다.
- 재분류의 *원인 코드*(era 판정 로직의 어느 변경인지)는 확인하지 않았다.

## 5. Residual-risk

23건이 지금 clean 이라는 것은 현재 트리 기준이다. 다른 브랜치에서 이 23건 중
하나를 건드리면 그 브랜치에서는 곧바로 MUST-FIX 로 뜰 수 있다.

---

## 부록 — `moai spec close --backfill-only` 결함 재현 (신 바이너리)

명령: `moai spec close <SPEC-ID> --backfill-only --dry-run` (쓰기 없음)

대상 2건 모두 동일한 출력:

```
Would apply the following transitions:
  spec.md:frontmatter.status → completed
  progress.md:§E.3.status → completed
  progress.md:§E.5.mx_commit_sha → <derived-from-recent-mx-commit>
```

| 결함 | 신 바이너리 상태 | 근거 |
|---|---|---|
| ① 은퇴한 §E.5 mx 신호를 기입 | **재현됨** | 전이 목록에 `§E.5.mx_commit_sha` 가 그대로 있다. `--help` 본문도 여전히 "§E.5 Mx section present" 를 전제로 적혀 있다 |
| ② `updated:` 미갱신 | **재현됨** | 전이 목록에 `spec.md:frontmatter.updated` 가 없다 |
| ③ sync SHA 를 병합 커밋으로 오유도 | **재현 못 함** | 이 축을 태울 픽스처가 이 트리에 없다 — `implemented` 상태이면서 `sync_commit_sha` 가 비어 있는 SPEC 이 0건. 후보로 잡힌 것들은 전부 `[noop] already at status: completed` 로 빠진다. **반증이 아니라 미검증이다** |

### 부수 관측 — 종결 함정이 그대로 있다

`SPEC-CI-FLAKE-SERIES-001` 에 dry-run 을 걸면 **닫겠다고 답한다.**
이 SPEC 은 관측 창이 아직 열려 있어 의도적으로 연기된 것이다
(종결 하한 `2026-09-02T18:05:57Z`). 도구는 그 연기를 읽지 못한다 —
terminal state 만 면제하기 때문이다. 즉 **의도적 연기를 기계가 읽을 수 있는
표지가 없는 한, 다음 작업자가 이 SPEC 을 실수로 닫는 경로는 열려 있다.**
오늘은 사람 판단이 막았지만, 사람 판단은 다음번에 없을 수 있다.
