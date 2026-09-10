# t329 판정서 — git 독트린의 존재하지 않는 패키지 인용 정리

- 판정: **FIXED** — 카드 전제 확인. `.moai/docs/git-workflow-doctrine.md` §18.12가 `internal/bodp/relatedness.go`의 `Check()`/`applyMatrix()`를 함수명·파일경로까지 대며 인용했으나 그 패키지는 존재하지 않는다.
- 카드 범위 준수: 기본값 `origin/main` 자체는 **변경하지 않음**(카드가 명시한 범위 밖 — 배포 룰은 GitHub Flow 다운스트림 사용자를 위해 그 값이 옳음).

## 전제 검증 (본 런 실측, develop `2660bcd09`)

- `git ls-tree -r --name-only develop | grep -i bodp` → `.moai/bodp/plan/SPEC-V3R4-HARNESS-003-2026-05-15.json` · `.moai/reports/expert-debug/bodp-audit-trail-leak-2026-05-09.md` 2건뿐 — `internal/bodp/` 추적 파일 **0건** (리드 실측과 일치).
- 발견 경로 문서(`.moai/reports/t310/verdict.md` Residual-risk)의 계열을 추적해 수정.

## 실재하는 표면 (인용의 새 목적지)

- `.claude/rules/moai/development/branch-origin-protocol.md:26` — [HARD] "When no signal fires, the recommendation is `origin/main`"
- `...branch-origin-protocol.md:29` — § Algorithm (3-Signal Evaluation)
- `...branch-origin-protocol.md:39` — § Decision Matrix (verbatim 8-row truth table)
- 템플릿 미러 존재(`internal/template/templates/.../branch-origin-protocol.md`) — 배포 사용자도 읽을 수 있는 SSOT.
- `internal/cli/doctor.go:908` — `checkBODPConfig`: `.moai/branches/` os.Stat 존재 확인 전용, base 선택 로직 부재 (카드 좌표 :894-906에서 드리프트 — 실측 908로 갱신).
- `branch-origin-protocol.md:60`의 `internal/bodp` 언급은 **은퇴 서술**(호출자 부재 사실 서술)로 정확 — 유지.

## 변경 (이름 나열)

- `.moai/docs/git-workflow-doctrine.md` — §18.12 두 인용 행을 실재 표면으로 교체:
  - :402 `internal/bodp/relatedness.go Check()` → `branch-origin-protocol.md § Algorithm` + 패키지 부재 실측 주석
  - :412 `internal/bodp/relatedness.go applyMatrix()` → `branch-origin-protocol.md § Decision Matrix` + :26 [HARD] 기본값 근거 + `doctor.go:908` 존재-확인 전용 명시
- 동류 스윕 결과(수정 불필요): doctrine 내 나머지 `internal/` 인용은 `:212`(`internal/template/templates/`, 존재)뿐; `git-local-workflow-doctrine.md`는 `internal/` 인용 0건; branch-origin-protocol.md의 코드 인용은 은퇴 서술 1건(정확).

## 5-섹션

**Claim**: §18.12의 모든 코드 경로 인용이 실재하는 표면을 가리킨다. 코드 인용으로서의 `internal/bodp/` 참조는 제거됐고, 패키지 **부재를 선언하는** 주석 1건만 의도적으로 잔존한다.

**Evidence**: 교체 후 `grep -n "internal/bodp" .moai/docs/git-workflow-doctrine.md` → `:402` 1건(부재 선언문 — "존재하지 않는다(… 추적 파일 0건)"); `grep -n "internal/" …` → `:212`(유효)·`:402`(부재 선언)·`:412`(신규 — `doctor.go:908` 인용, 실재 확인 `sed -n '908p'` → `func checkBODPConfig`). `branch-origin-protocol.md:26` [HARD] 기본값 실재 확인.

**Baseline-attribution**: `WT-bodp-stale-citation` 워크트리, develop `2660bcd09`, 2026-09-02 본 런.

**Gaps**: doctrine은 로컬 전용 파일이므로 make build·템플릿 사이클 불필요(템플릿 미러 부재 실측으로 확인). t310 판정서 원문은 `.moai/reports/t310/verdict.md`(브랜치 트리) 판독 없이 카드 본문 서술에 의존 — 카드가 요약을 제공했고 전제는 자체 실측으로 독립 확인.

**Residual-risk**: §18.12의 8행 행렬 표 본문(:414-421)은 룰 파일 :39의 행렬과 동일 서술로 유지 — 두 문서가 같은 행렬을 이중 서술하는 상태는 그대로(단일화는 독트린 개편 소관, 카드 범위 밖). 룰 파일 쪽이 정본이고 doctrine은 §18.12 서두에서 룰 파일로 인용을 돌렸으므로 드리프트 시 룰 파일이 우선한다.
