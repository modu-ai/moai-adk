# t490 RED baseline — spec-lint `--strict` 전수 계수 (수리 전 측정)

> 카드: t490 · 브랜치: `WT-speclint-status-transition` · 측정 트리: 본 워크트리 `.claude/worktrees/t490`
> base/HEAD: `615d18c1f` (= origin/develop, CI run 34014859906 잡 `spec-lint`가 판정한 그 트리)
> 관측: 2026-09-06 · 본 세션(lane-15) 실측

## Claim (주장)

- **C1**: 이 트리 전수 `moai spec lint --strict` 결과는 **error 2건 / warning 4,333건**(총 발견 4,335)이고, 잡 FAILURE를 만드 것은 **error 2건**이다 — CLI의 종료코드 로직이 error 전용이기 때문이다(`internal/cli/spec_lint.go:97` `spec lint: error-severity findings detected` → exit 1).
- **C2**: error 2건은 **`ArtifactStatusFieldForbidden`** 으로 한 SPEC의 형제 아티팩트에 몰려 있다 — `SPEC-CODEX-E2E-MEASURE-001/plan.md:5` 와 `acceptance.md:5` 의 `status: completed` (Artifact Statelessness 위반, 스키마 § Artifact Statelessness).
- **C3**: WARNING 계열은 이 카드 범위 밖이다 — StatusTransitionInvalid **99건**(그중 48건이 커밋 `ce779f9ee` 인용), MissingExclusions **25건**. 경고는 종료코드에 불식되지 않는다.
- **C4**: `ce779f9ee` 판독 — 2026-05-16 `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` (#939), **의도적 일괄 status 정리**(completed→implemented 47건 등 77건 일괄). 경고 99건은 이 역사적 사실의 올바른 보고이며, 수리는 규칙 의미론 설계 판단이 필요한 별도 소관이다.

## Evidence (증거 — 명령 + 실측 출력)

명령: `go run ./cmd/moai spec lint --strict` — CI(`.github/workflows/spec-lint.yml:58`)와 동일 호출, 파일 redirect 후 종료코드 직접 포획(파이프 없음). 전체 출력: 본 세션 `/tmp/t490-speclint-baseline.txt` (4,340행 — 헤더 2행 + 발견 4,335 + 기타), 종료코드 **1**.

```
exit-code:1
ERROR     ArtifactStatusFieldForbidden  .../SPEC-CODEX-E2E-MEASURE-001/plan.md        5   `plan.md` carries `status: completed` in its frontmatter. ...
ERROR     ArtifactStatusFieldForbidden  .../SPEC-CODEX-E2E-MEASURE-001/acceptance.md  5   `acceptance.md` carries `status: completed` in its frontmatter. ...
```

계열 계수(본 파일에서 산출): `ERROR` 2 · `WARNING` 4,333 · StatusTransitionInvalid 99 · MissingExclusions 25 · `ce779f9ee` 인용 48.

대상 blob 핀: plan.md `15d5ef6ae478b97fc301f9c778e472aa7bd6c8d8` · acceptance.md `13d830137cd34b4fcb3616a92f0099e4880f2d54` (`HEAD:` 기준, 2026-09-06).

## Baseline-attribution (귀속)

- **명령**: `go run ./cmd/moai spec lint --strict` — **트리 소스에서 빌드된 툴**(CI와 동일한 근거, VCI §2.2 충족). 판별식은 종료코드가 아니라 error/warning 계수다.
- **트리**: `.claude/worktrees/t490` @ `615d18c1f` — 2026-09-06 본 세션 실측.
- **base ref 근거**: lint의 main-branch 체인은 `["main", "origin/main", "master", "origin/master"]`(`internal/spec/gitquery_cache.go:129`)이고 본 트리의 local `main` = `origin/main` = `7ad9f8534`(fetch 후 비교 실측) — CI의 `git fetch origin main:main` 결과와 동일 커밋을 가리켜 재현이 정확하다. CI의 fetch 방식은 로컬에서 불가(local main이 primary에 체크아웃 — `refusing to fetch into branch` 확인)지만 결과 근거가 동일하므로 무실이다.
- **lane-8 인용값 대조**: 「2 errors / 4333 warnings」는 t472 창 트리에서 재 값으로, 본 독립 재측정과 **정확히 일치**한다. 리드가 「직접 재세요」라고 지시한 부분의 이행.
- **과거 CI 판정**: run `34014859906` 잡 `spec-lint` FAILURE — 로그 표본(Agency-Absorb 등)은 WARNING이고 잡을 뒤집은 줄은 본 2건이다.

## Gaps (미검증)

- WARNING 4,333건의 전수 분류는 미수행 — 본 카드 판정에 필요한 계열(StatusTransitionInvalid 99 · MissingExclusions 25)만 셌다.
- CI 원격 재판정은 미관측 — 리드 일괄 push 후 판독 소관이다.
- StatusTransitionInvalid의 처리 방침(간선 추가/면제/개별 수정)은 별도 카드·운영자 판단 사항으로 본 카드에서 결정하지 않는다.

## Residual-risk (잔여 위험)

- go run 빌드는 트리 소스 기반이므로 CI와 동일 근거지만, Go 툴체인 마이너 버전 차이(go.mod 지정 동일 가정)는 별도 확인하지 않았다.
- error 2건 제거 후에도 새 SPEC 유입에 따라 error가 재유입될 수 있다 — 본 카드의 판정은 이 트리 위에서만 성립한다.
