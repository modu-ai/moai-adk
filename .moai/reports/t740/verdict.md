# t740 verdict — SPEC Lint 게이트: 카드 전제는 t525 가 이미 수리, 현재 초록 확인

- 날짜: 2026-09-14
- 카드: t740 (Class B — defect, cause unknown; plan 생략, run → sync)
- 브랜치: `WT-spec-lint-gate` @ b1bd81b23 (local develop 팁 = origin/develop a404132e7 + t748 미푸시)
- 코드 변경: **0건** — 원인 확립 + 상류 수리(t525) 검증이 카드의 전부였다

---

## Claim

카드 명제 — "develop CI SPEC Lint 게이트가 경고 4,344건 exit 1 로 5회 연속 실패(상시 적색), 상시 적색인 게이트는 게이트가 아니다" — 는 **09-07 시점까지는 사실**이었으나, 그 수리는 **t525(SPEC-SPECLINT-GATE-SIGNAL-001)가 09-08~09-11 사이 이미 착지**시켰다. 본 카드가 배차된 09-14 시점의 develop 에서 게이트는:

1. **초록**이다 — 원격 팁(a404132e7)과 로컬 팁(b1bd81b23) 모두에서 오류 0, 비-advisory 기준선 위반 0.
2. **신호를 낸다** — 오류는 무조건 차단, 비-advisory 경고가 기록치를 넘어오르면 규칙명+델타를 이름대고 차단, standing advisory 재고(3,147건)는 보이되 차단하지 않는다. 카드의 성공 기준("새 적색을 알아볼 수 있게")이 충족된 상태다.

따라서 본 카드는 **코드 변경 없이 확정 종료**한다. 원인 확립과 검증이 카드의 산출물이다.

## 원인 확립 — 카드 3선지 중 (b) 정책 결함이었고 t525 가 고쳤다

| 선지 | 판정 | 근거 |
|---|---|---|
| (a) 경고 4,344건 = 진짜 부채, 게이트 옳음 | **아님** | t518 흡수 후 전량 advisory 재분류(377d98c5a 기준선 사유문: "0 non-advisory warnings (3133 warnings, all advisory)") |
| (b) 경고로 exit 1 을 내는 정책이 결함 | **이것이었다** | 09-07 실패 run 의 명령이 `go run ./cmd/moai spec lint --strict` — standing 경고를 전부 차단 사항으로 승격하는 플래그가 워크플로에 명시돼 있었다 |
| (c) 경고 자체가 오탐 | 부분 해당, t518 소관 | 수집기 맹축(표 형식·한국어 서술어·맹디렉터리)은 t518/SPEC-SPEC-LINT-BLIND-AXES-001 이 수리 |

시계열:

- **09-03~09-07**: `--strict` 시대 — 재현 run 34088825415 (d4162b368): 단계 "Run SPEC lint", 마지막 줄 `0 error(s), 4344 warning(s)` → `##[error]Process completed with exit code 1.` 카드가 목격한 5연속 적색.
- **09-08 09:03** (9fb52f746): t525 M4 — SPEC-V3R4-CC2X-ADOPT-001/002 맹디렉터리 2건 폐쇄 (카드 부수 항목 SpecsDirMissingSpecFile 해소).
- **09-08 14:03** (9e1744469): t525 M3 — baseline 배선 + 초기 기준선. 카드 발행(09-07)이 이 착지보다 앞선 카드였다.
- **09-11 11:46** (377d98c5a): t525 M3.4 — t518 흡수 후 gated 재기준선. baseline `rules: {}`, 전체 3,133건 advisory.
- **09-13**: CI develop 완결 run 5건 전부 success (아래 Evidence 2).

## Evidence

1. **로컬 baseline 게이트** (트리 b1bd81b23, 2026-09-14 측정):
   명령 `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json` → **rc=0**.
   꼬리: `0 error(s), 3147 warning(s)` / `baseline: OK — .moai/spec-lint-baseline.json` /
   `inventory: 3147 warning(s) total (advisory included), 0 non-advisory tracked across 0 recorded rule(s)`.
   전문(3,361행): `.moai/reports/t740/lint-baseline-b1bd81b23.txt`
2. **CI 이력** (`gh run list --workflow "SPEC Lint" --branch develop --limit 8`):
   재기준선 후 완결 run 전부 success — 62fbd6baf(09-13 11:30) · 7a7a08f20(12:49) · b66789479(13:13) · d416f8162(16:58) · a404132e7(18:34 시작, 18:44:40Z success, REST API `gh api .../actions/runs/34775055880` 직접 확인). cancelled 3건은 concurrency cancel-in-progress 정상 동작.
3. **09-07 실패 run 로그 tail** (`gh run view 34088825415 --log-failed`): `0 error(s), 4344 warning(s)` → `##[error]Process completed with exit code 1.` — 실패 단계명 "Run SPEC lint", d4162b368 시점 워크플로(`git show d4162b368:.github/workflows/spec-lint.yml`)의 명령행 `go run ./cmd/moai spec lint --strict`.
4. **기준선 파일** (b1bd81b23): `tree_sha: 4ac93f755`, `updated_at: 2026-09-11`, `rules: {}` — 기록된 비-advisory 부채 0.
5. **SpecsDirMissingSpecFile**: 현재 트리에서 0건 (측정 출력 grep count=0). 09-07 run 에서 지적됐던 2개 디렉터리는 부재 — 9fb52f746(t525 M4)가 폐쇄. "왜 그 상태였는지"에 대한 추가 조사 불요: 카드 발행 3일 뒤 t525 가 같은 관측으로 닫았다.
6. **위상** (`git fetch origin develop` 후): `origin/develop = a404132e7`, `a404132e7` 는 `b1bd81b23` 의 조상 — 본 카드 트리는 원격 팁 + t748(미푸시) 1병합 차등.

## Baseline-attribution

전 실측은 2026-09-14, 본 카드 워크트리(`.claude/worktrees/t740`, 브랜치 `WT-spec-lint-gate` @ b1bd81b23)에서 **이번 run 으로 직접 수행**했다. 인용한 수치 중 이 run 밖에서 온 것은 없다. 카드 본문의 4,344/4,346 등 과거 수치는 재인용하지 않았고 원인 서사의 역사적 사실로만 둔다(카드 경고 "코퍼스 수는 얼리지 마라" 준수).

## Gaps

- 3,133(기준선 시점)→3,147(현재) advisory 증분 +14건의 규칙별 분해는 하지 않았다 — advisory 는 게이트 판정 불변량이 아니어서 무관하다.
- ~~plain(error-only) 경로의 rc 미측정~~ → 측정 완료: **rc=0**, `0 error(s), 3147 warning(s)` — 두 게이트 경로 모두 초록 확인.
- t518 수집기의 advisory 재분류 자체가 옳은지는 본 카드가 재판정하지 않았다 — t518 소관.
- CI 의 `--strict` 미사용 경로에서 잠재 error 급 규칙 ~600건 노출 시나리오(카드 본문 언급)는 시뮬레이션하지 않았다 — 현재 0 error 실측이 있어 발화 조건이 아니다.

## Residual-risk

- **재기준선은 수동 절차로 남는다.** SPEC 착지가 비-advisory 경고를 올리면 매번 수동 rebaseline 커밋이 필요하고, 잊으면 새 적색이 쌓인다. 다만 그 적색은 규칙명+델타를 이름대므로 09-07 의 무신호 적색과는 다른, 읽을 수 있는 적색이다. 절차 개선(자동 rebaseline 제안 등)이 필요하면 별도 카드.
- **advisory 재분류를 신뢰한다.** 재분류가 틀려 실제 차단 급이 advisory 로 묶였다면 게이트가 조용히 통과한다 — 이 축의 판정은 t518 에 귀속되며 본 카드는 재검증하지 않았다.
- 운영자 판정 대상 **없음** — 본 카드는 임계·정책을 완화하지 않았다(수리는 기착지 t525 소관).
