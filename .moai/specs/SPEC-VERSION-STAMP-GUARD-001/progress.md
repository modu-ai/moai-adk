# SPEC-VERSION-STAMP-GUARD-001 — 진행 기록

카드: t388 · Tier S · 워크트리 `WT-version-sync-list`

## §E.1 Plan-phase Audit-Ready Signal

- plan-phase 산출: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- 기준선: `.moai/reports/t388/baseline.md` (트리 `9328a5242`)

### iter-1 (0.1.0)

- 요구 7 / 수락 7. 토큰 술어 가드 + 문서 수정을 한 카드로 묶었다.

### iter-2 (0.2.0) — plan-audit FAIL 대응

- 감사: `.moai/reports/t388/plan-audit.md` — **FAIL 0.77**(Tier S 임계 0.75를 넘겼으나 D2가
  critical/blocking)
- D1·D3·D4·D5-D10 정정. D2는 미해결 blocker로 기록.

### iter-3 (0.3.0) — 운영자 판정으로 카드 분할

- **토큰 술어 가드 축 전체가 카드 t392로 이관**(Tier M, 이 카드 착지에 의존). D2는 수리가
  아니라 **범위 이관으로 해소**됐고, 측정치(2,225줄 / 592파일 + 토큰 히스토그램)는
  `spec.md` §2에 보존했다.
- 삭제된 요구: 0.2.0의 REQ-VSG-004(목록 비의존 스윕) · REQ-VSG-005(토큰 술어) ·
  REQ-VSG-006(토큰 가드 비공허성). 딸린 §3 거부 목록 · §4 `eba919e44` RED 픽스처 ·
  `internal/versionstamp/` 패키지 · AC 3건도 함께 삭제(이월 아님).
- D1·D3는 §3과 함께 소멸. **D4는 살아남아 AC-VSG-001에서 닫았다**(리터럴 7경로가 판정자,
  `61921f1ba`는 출처 인용으로 강등). D7은 살아남은 검사에 재적용(합성 입력
  `docs-site/nonexistent-stamp.toml` 고정), D8은 REQ-VSG-004의 `Where` 절로 이월.
- 남은 산출: Go 파일 하나(`internal/cli/version_sync_list_test.go`) + 문서 하나.
- **회귀 보장은 절반**임을 `spec.md` §4 + REQ-VSG-006 + AC-VSG-006으로 명시. 누락 방향은
  t392 소관.
- 요구 6 / 수락 6 (Tier S 상한 8/8 이내).

### iter-4 (0.4.0) — plan-audit iter-2 FAIL 0.80 대응

- 감사: `.moai/reports/t388/plan-audit-iter2.md` — **FAIL 0.80**(Tier S 임계 0.75를 넘겼으나
  D1이 major/blocking). 결함 9건(D1 major/blocking · D2-D5 minor/blocking ·
  D6-D9 minor/optional).
- **D1 — 주 수리.** 검사의 RED 증거를 유령에 귀속시키던 서술을 걷어냈다. 앵커가 가리키는
  스탬프 소제목은 M2가 만들므로 M1 트리에서 파싱은 0건이고, 그때 우는 것은 **개수 단언**이지
  존재 단언이 아니다. 두 단언이 **각자의 RED을 따로 관측**하도록 재설계했다:
  E3-a(M1, `parsed=0 expected=7`) · E3-b(M2.1, 두 단언 동시 — 보조 관측) ·
  E3-c(M2.3, 치환으로 존재 단언 단독). 기대 RED 리터럴을 측정 **전에** `plan.md` §D의 단언
  메시지 계약으로 못박았다. 새 절 `spec.md` §5.1이 근거를 [HARD]로 적는다.
- **AC-VSG-004는 추가가 아니라 치환**으로 바뀌었다 — 한 줄을 더하면 개수가 8이 되어 원인이
  둘이 된다. `docs-site/hugo.toml` → `docs-site/nonexistent-stamp.toml` 치환은 개수를 7로
  유지해 존재 단언 단독의 빨강을 만든다. D6도 같은 수리로 함께 닫힌다.
- D5: REQ-VSG-005의 순환 비교를 「검사가 상수로 보유한 기대 개수 7」로 교체.
- D4: AC-VSG-006 3항을 판단에서 계기로 교체 — 양성 존재(`partial`, `does not detect`) +
  리터럴 거부 목록 5건, 둘 다 grep 판정.
- D2: `spec.md`의 걸린 `§3의 단위 고정 조항` 포인터 제거(사유를 인라인으로).
- **D3: 측정 정정.** 「이름에 버전 토큰이 든 파일이 둘」은 거짓이었다 — 아래 재측정 참조.
- D7: AC-VSG-006의 Given을 문서로 한정하고, SPEC 절반은 존재 판정(이미 성립)으로 커버리지 유지.
- D8: `항목 71-78행` → `71-74·77-78행`(75 공백, 76 라벨). D9: REQ-VSG-004의 `Where` 절을
  본절로 접었다.
- 요구 6 / 수락 6 불변. Tier S 상한 8/8 이내.

### iter-4 재측정 (전부 이 워크트리, 트리 `9328a5242`에서 직접 실행)

- 문서 소제목: `### Files Requiring Version Sync`(66) · `**Documentation Files:**`(70) ·
  `**Configuration Files:**`(76). **스탬프/산출물 축은 없다** — D1의 근거
- 목록 항목 행: 71-74(문서 라벨 아래 4건) · 77-78(설정 라벨 아래 2건). 75 공백, 76 라벨 — D8
- 유령은 78행, `**Configuration Files:**` 아래 — 앵커 한정이 읽지 않는 위치
- `test -e internal/cli/version_sync_list_test.go` → 종료 1 (AC-VSG-004/005의 RED-now)
- `test -e docs-site/hugo.toml` → 종료 0 (치환 대상이 실재함)
- **D3 재측정** — 거부 목록 범위에서 이름에 버전 토큰이 든 파일은 **8개**(2가 아니다),
  매치 줄 합 **113**:
  `docs/design/v2.14.0-release-plan.md` 40 · `.moai/release/RELEASE-NOTES-v2.17.0.md` 24 ·
  `.moai/release/MIGRATION-v2.17.0.md` 16 · `.moai/release/RELEASE-NOTES-v2.16.0.md` 12 ·
  `.moai/marketing/awesome-lists/github-release-v2.12.0-enhanced.md` 7 ·
  `.moai/release/RELEASE-NOTES-v2.15.0.md` 6 · `.moai/release/v2.15.0-draft.md` 4 ·
  `.moai/release/RELEASE-NOTES-v2.20.0.md` 4
  - 교차 검증: `-n` 출력에서 뽑은 전체 출현 **2607** − `-h` 기반 **2494** = **113**. 정확히 일치
  - `.moai/release/` 6개가 66줄을 차지한다(`grep -c '^\.moai/release/'` → 66). 이 거부 목록은
    `.moai/release-notes/`만 제외하고 `.moai/release/`는 제외하지 않는다 — t392가 밟을 자리
  - 토큰별: `v2.14.0` 72→112(+40) · `v2.12.0` 83→90(+7) · `v2.17.0` 25→65(+40, 두 파일 합산)
- 면적·히스토그램 재확인: **2225줄 / 592파일**, 출현 총 **2494**
  (v3.0.0 270 · v2.12.0 83 · v3.1.1 80 · v2.1.219 80 · v2.14.0 72 · v2.1.198 68) — 0.3.0과 동일
- AC-VSG-006 RED-now: 문서 전체에 `partial` / `does not detect` **0건**
- REQ-VSG-006 SPEC 절반: `grep -cF '이 카드가 착지해도 목록은 여전히 썩을 수 있다' spec.md` → 1

### iter-3 세션의 재측정 (전부 트리 `9328a5242`에서 직접 실행)

- 스탬프 집합: `61921f1ba` numstat → 7파일 9줄
- 유령: `test -e internal/template/templates/.moai/config/config.yaml` → 종료 1
- 유령이 목록에 존재: 해당 경로 grep → 1건, 종료 0
- 현재 라벨 축: `**Documentation Files:**`(70행) / `**Configuration Files:**`(76행)
- 릴리스 산출물 플레이스홀더 부재: `.moai/release-notes/vX.Y.Z.ko.md` 없음
  (실재 `v3.1.0.ko.md`·`v3.1.3.ko.md`)
- 토큰 술어 면적: **2225줄 / 592파일** (단위=줄)
- 토큰 히스토그램(단위=출현, 전체 2494): v3.0.0 270 · v2.12.0 83 · v3.1.1 80 · v2.1.219 80 ·
  v2.14.0 72 · v2.1.198 68
  - 정정: 초판은 `-n` 출력에서 토큰을 뽑아 **파일 이름 속 버전이 중복 계수**됐다
    (`v2.14.0` 112→72 — `docs/design/v2.14.0-release-plan.md` 매치 40줄이 경로에서 유입;
    `v2.12.0` 90→83 — `github-release-v2.12.0-enhanced.md` 매치 7줄). `-h`로 재측정해 교체.
    면적 수치는 영향 없음
- 도달 가능성: `61921f1ba` 조상 rc=1 · `eba919e44` 조상 rc=0
- `spec-lint.yml` fetch-depth 없음(rc=1) · `ci.yml` checkout 7개 중 6개 `fetch-depth: 0`
- 합성 입력 부재 확인: `test -e docs-site/nonexistent-stamp.toml` → 종료 1

상태: `draft` — Implementation Kickoff Approval 대기

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
