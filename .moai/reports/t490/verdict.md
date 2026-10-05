# t490 판정 — spec-lint ERROR 2건 수리 (SPEC-SPECLINT-ARTIFACT-STATUS-001)

> 카드: t490 · 브랜치: `WT-speclint-status-transition` · 측정 트리: 본 워크트리 `.claude/worktrees/t490`
> base: `615d18c1f` (= origin/develop, CI run 34014859906 잡 `spec-lint`가 판정한 그 트리)
> 커밋 체인: RED baseline+plan `a957d8308` → 수리 `a75041e37` → 본 판정서 (§2.3 baseline-first 순서를 커밋 그래프가 증언)
> blob: 대상 파일 수리 전 `15d5ef6ae…`/`13d830137…` → 수리 후 `564eae267…`/`9d27ce22a…`
> 관측: 2026-09-06/07 · lane-15 본 세션 실측 (단일 측정자 — 레인이 모든 측정 수행)

## Claim (주장)

- **C1**: CI spec-lint 잡 FAILURE의 원인은 **ERROR 2건**뿐이다 — CLI는 error 전용으로 exit 1을 낸다(`internal/cli/spec_lint.go:97` `report.HasErrors()` → `error-severity findings detected`). 경고 4,333건은 종료코드에 불식되지 않는다.
- **C2**: 그 2건(`ArtifactStatusFieldForbidden` — `SPEC-CODEX-E2E-MEASURE-001/plan.md:5`·`acceptance.md:5`)의 근원은 카드 t462의 3-phase close 커밋 `185ce7d57`가 형제 아티팩트에 `status: completed`를 복제한 것이다(Artifact Statelessness 위반). 수리는 그 두 줄 삭제뿐이고 나머지 frontmatter(`id`·`title`·`version`·`created`·`updated`·`author`·`tier`)는 바이트 동일 보존됐다.
- **C3**: 4회 측정 — **RED 2 → GREEN 0 → 뮤턴트 2 → 재확인 0**. 뮤턴트가 실제로 뒤집혀 「고쳤다」와 「검사가 못 본다」가 구별됐다.
- **C4**: WARNING 계열은 본 카드 범위 밖으로 기록했다 — StatusTransitionInvalid 99(48건이 2026-05-16 #939 `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001`의 **의도적 일괄 정리 커밋 `ce779f9ee`** 인용 — 경고는 역사를 올바르게 보고), MissingExclusions 25(terminal-status 강등, 비차단). 수리는 규칙 의미론 설계 판단이 필요해 별도 소관이다.
- **C5**: plan-audit은 2회 반복 — 1차 FAIL 0.81(MP-3 `lifecycle` enum) → **운영자 승인 예외**(Tier S 상한 1 소모, 델타 확인 한정 — 예외 기록은 plan-audit.md 헤더) → 2차 **PASS 0.94**.

## Evidence (증거 — 명령 + 실측 출력, 본 세션 본 트리)

명령: `go run ./cmd/moai spec lint --strict` — CI(`.github/workflows/spec-lint.yml:58`)와 동일 호출, 파일 redirect 후 종료코드 직접 포획(파이프 없음), **판별식은 `^ERROR`/`^WARNING` 계수**.

| Run | 트리 상태 | `^ERROR` | `^WARNING` | exit |
|---|---|---|---|---|
| A — RED (baseline, `red-baseline.md`) | `615d18c1f` | **2** | 4,333 | 1 |
| B — GREEN | 수리 적용(`a75041e37`) | **0** | 4,333 | 1 |
| C — MUTANT | `git checkout 615d18c1f -- <두 파일>` | **2** | 4,333 | 1 |
| D — 재확인 | `git checkout HEAD -- <두 파일>` | **0** | 4,333 | 1 |

- Run B/C/D 전문: 본 세션 `/tmp/t490-speclint-{green,mutant,regreen}.txt`. RED 전문: `red-baseline.md`(`2 errors / 4333 warnings` — lane-8 인용값과 독립 일치, 감사인 3중 재현).
- exit 1은 경고 잔존 때문의 **예상값**이며 판정이 아니다(RQ-006 · AC-GREEN의 명시 조항).
- 폐쇄 검사: 수리 커밋 `a75041e37`은 **정확히 2파일·2삭제**(`-status: completed` × 2)이고 다른 파일을 만지지 않는다(AC-SCOPE).

## Baseline-attribution (귀속)

- **명령**: 위 표의 단일 명령 — **트리 소스에서 빌드**(CI와 동일 근거, VCI §2.2). 실행: 2026-09-06/07 본 세션.
- **트리 상태**: Run A = `615d18c1f` · Run B = 워킹트리(수리 후 커밋 `a75041e37`) · Run C = `a75041e37` + 두 파일만 base blob · Run D = `a75041e37`(파일 HEAD와 동일 — `git status`에서 두 파일 침묵 확인).
- **base ref 근거**: lint의 main 체인 최우선 `main` = `origin/main` = `7ad9f8534`(fetch 후 비교 실측) — CI의 `git fetch origin main:main` 결과와 동일 커밋.
- **plan-audit**: 1차 FAIL 0.81 → 운영자 예외 → 2차 **PASS 0.94** — `.moai/reports/t490/plan-audit.md`(예외 기록 포함).

## Gaps (미검증)

- **원격 CI 재판정은 미관측** — 리드 일괄 push 후 판독 소관. 예상: error 0이므로 잡은 **경고 4,333이 남아 있어도 통과**(exit 0)해야 한다 — 잡 로그에 WARNING이 계속 보이는 것은 정상이다.
- WARNING 4,333건의 전수 분류 미수행 — 계열 2개(StatusTransitionInvalid 99 · MissingExclusions 25)만 계수했다.
- StatusTransitionInvalid의 처리 방침(간선 추가/역사 면제/개별 수정)은 규칙 설계 판단 사항으로 별도 카드 후보다.

## Residual-risk (잔여 위험)

- 새 SPEC 유입이 error를 재유입시키면 잡은 다시 적색이 된다 — 본 판정은 이 트리 위에서만 성립한다.
- go run의 Go 툴체인과 CI의 setup-go(go.mod 기준) 버전은 동일 지정을 가정했고 별도 대조하지 않았다.
- 병합 시 develop 측에서 이 SPEC이나 대상 파일이 움직였다면 창에서 재측정한다(흡수 tip 대비 diff로 확인).
