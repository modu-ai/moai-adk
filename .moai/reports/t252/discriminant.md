# t252 — SPEC 종결 전이 백필: 판별식 + 실측

측정 트리: worktree `.claude/worktrees/t252`, HEAD `18ba3cddb` (= `origin/develop`, 배차 기준선과 일치)
측정일: 2026-09-02

## Claim

`implemented` 정체 SPEC 중 **실제 종결 부채는 3건**이다. 배차문이 인계한 148건(재측정 145건)은
"status=implemented 개수"이지 "종결 부채 개수"가 아니다 — 단위가 다르다.

## Evidence

### 1. 현재 status 분포 (배차 148 → 재측정)

```
$ git grep -h -E '^status:' origin/develop -- '.moai/specs/**/spec.md' | ... | uniq -c
 547 completed      145 implemented     31 archived      13 draft
  10 in-progress      9 superseded       1 retired        1 rejected  ...
```

### 2. 판별식의 정본은 grep 이 아니라 `moai spec audit`

era 분류 + 드리프트 판정은 착지된 Go 정책이다(`internal/spec/era.go` `ClassifyEra()`,
`internal/spec/audit.go` `checkV3R6Drift()`). SSOT 문서: `.claude/rules/local/lifecycle-sync-gate.md`.

```
$ moai spec audit --json   → rc=0
total_specs 740 / grandfathered 286 / modern_era_clean 449 / findings 516
findings 내역:  509 INFO EraAutoDetected · 5 MUST-FIX SyncStatusDrift · 2 INFO AuditError
```

**grandfather clause 가 결정적이다.** V2.x / V3R2-R4 / V3R5 는 `era_final: true` 로
드리프트 판정에서 제외된다 — 소급 정규화는 정책상 금지다. `implemented` 145건 중
절대다수가 여기 속한다. **일괄 전이는 규율 위반이다.**

### 3. MUST-FIX 5건 (도구 판정)

| SPEC | sync_commit_sha | 착지 |
|---|---|---|
| SPEC-V3R6-AUDIT-MODEL-PIN-001 | pending-backfill-sync | `baa100ce5` (PR #1642) ancestor rc=0 |
| SPEC-STATUSLINE-PROFILE-RESPECT-001 | `2114ed981` | ancestor rc=0 |
| SPEC-VACUOUS-FLOOR-GUARD-001 | pending-backfill-sync | `85c59f155` ancestor rc=0 |
| SPEC-CI-FLAKE-SERIES-001 | `3a7aaab37` | **ancestor rc=1** (squash 전 sha, 기록 부정확) |
| SPEC-FULL-SUITE-DOCTRINE-001 | `b904de9a3` | ancestor rc=0 |

### 4. 도구가 못 보는 축 — 의도적 연기

`checkV3R6Drift` 의 면제는 terminal state(`completed`/`superseded`/`archived`/`rejected`) 뿐이다.
**"의도적 연기"에 해당하는 면제가 없다.** 그래서 아래 2건이 MUST-FIX 로 올라온다:

- **SPEC-CI-FLAKE-SERIES-001** — progress.md: *"`completed` 전환은 의도적 연기 (운영자 결정 2026-08-27) —
  AC-CFS-007(b) 관측 창이 충족된 뒤 후속 세션에서 종결."* 카드 t278 소관.
- **SPEC-FULL-SUITE-DOCTRINE-001** — progress.md: *"`completed` 전환은 의도적 연기 —
  AC-FSD-011 의 관측 창이 아직 열려 있다."*

이 2건을 닫으면 카드가 경고한 함정 1 — "안 끝난 SPEC 을 끝난 것으로 만듦" — 이 그대로 성립한다.

## 판별식 (3조건 AND)

`implemented` SPEC 이 종결 가능하려면:

1. `moai spec audit` 이 **MUST-FIX SyncStatusDrift** 로 판정 (era V3R6 + §E.2 + §E.4 + sync_commit_sha)
2. **AND** 착지 커밋이 `origin/develop` 의 조상 (`git merge-base --is-ancestor` rc=0)
3. **AND** progress.md 에 **의도적 연기 표지가 없음** (관측 창 대기 · 운영자 연기 결정)

조건 3 은 도구가 볼 수 없는 축이며, 사람이 읽어야 한다.

## Baseline-attribution

모든 수치는 이 트리(`18ba3cddb`)에서 이 실행으로 측정. raw: `.moai/reports/t252/spec-audit-raw.json`.

## 결론 — 종결 대상 3건

- SPEC-V3R6-AUDIT-MODEL-PIN-001 (카드가 명시한 t225 항목)
- SPEC-STATUSLINE-PROFILE-RESPECT-001
- SPEC-VACUOUS-FLOOR-GUARD-001

## Gaps (미검증)

- 3건의 acceptance AC 충족 여부를 개별 재검증하지 않았다 — sync 페이즈 기록을 신뢰했다.
- `moai spec close --backfill-only` 의 실제 거동(§E.4 sync_commit_sha 를 채우는지)을 아직 실행/관측하지 않았다.
- 136건 legacy 의 개별 내용은 표본 3건(SPEC-CORE-001·SPEC-MX-001·SPEC-V3R6-SKILL-GEARS-ALIGN-001)만 확인했다.

## Residual-risk

- 배차 이후에도 다른 레인이 SPEC 을 닫으면 5건 집합이 다시 밀린다 — 커밋 직전 `moai spec audit` 재실행 필요.
- `moai spec close` 가 §E.5 를 만들지 않는 알려진 거동(카드 명시) — 닫은 뒤 tail 직독으로 확인해야 한다.

## 부수 발견 (카드 범위 밖 — 별도 카드 후보)

1. **sync_commit_sha 파서 결함**: `extractProgressField` 가 인라인 주석/따옴표를 값에 흡수한다.
   5건 중 3건에서 관측 — 예: `"3a7aaab37\" — backfill 커밋에서 실제 short SHA로 교체 ..."`.
   조상성 검증을 자동화하려면 선행 수리가 필요하다.
2. **의도적 연기의 기계 가독 표지 부재**: 연기가 prose 로만 존재해 매 감사마다 MUST-FIX 로 재부상한다.
   다음 레인이 실수로 닫을 위험이 구조적으로 남는다.
3. **AuditError 2건**: `SPEC-V3R4-CC2X-ADOPT-001` / `-002` 는 `research.md` 만 있고 `spec.md` 가 없다.
4. **SPEC-CI-FLAKE-SERIES-001 의 sync_commit_sha 기록 부정확**: `3a7aaab37` 는 `origin/develop` 조상이 아니다.
