# t406 판정문 — release/v3.1.4 CI golden 6개 적색 분류·처방

작성: lane (WT-golden-six, base origin/develop `9145806d8`) · 2026-09-02
카드: t406 (Tier S, 브랜치 수리까지 — 태그·release.sh·GoReleaser 제외, t204 게이트)
측정 대상: origin/release/v3.1.4 @ `26898312e`, CI run 33382844571 (Test, ubuntu) · 33382844521 (Release Verify ×3 + Multi-OS Gate) · 33382844542 (graph-freshness)

---

## Claim (주장)

1. **golden 6개 실패는 전부 버전 스탬프 축이다.** doctor 집계 수치('5 ok, 0 warn' / '9 ok, 3 warn' / '3 ok, 8 warn' / 'Pass 17 Warn 11 Fail 0')는 got·want 양쪽에서 바이트 동일하며, 유일한 차이는 `MoAI Version: moai-adk v3.1.4`(got) vs `v3.1.3`(want) 한 줄이다. 카드가 의심한 "집계 수치 환경 의존" 가설은 기각 — 집계는 환경 의존이 아니다. lane-4의 "UPDATE_GOLDEN=1 재생성" 처방은 증상 치료일 뿐이다.
2. **근본 원인은 릴리스 브랜치가 버전 고정 수정을 안고 자랐기 때문이다.** release/v3.1.4는 `9328a5242`(2026-08-31 18:23)에서 잘렸고, golden 테스트의 버전 고정(`version.Version = "v0.0.0-test"`, SPEC-VERSION-STAMP-PREDICATE-001 REQ-VSP-008)은 develop에 `96bfa0c99`(2026-09-01 18:10, t392 M4)로 그 이후 착지했다. 릴리스 브랜치의 테스트는 버전을 고정하지 않아 범프된 실제 버전(v3.1.4)을 렌더링하고, golden은 범프 전 생성분(v3.1.3)이라 어긋난다.
3. **처방 = `96bfa0c99`를 release/v3.1.4에 cherry-pick.** 재생성이 아니다. 재생성 시 golden에 v3.1.4가 구워지고 다음 범프마다 재발하며, golden이 develop(v0.0.0-test)과 영구 갈린다. cherry-pick이면 릴리스 브랜치 테스트가 develop과 동일해져 버전 값과 무관하게 초록이고 golden도 수렴한다.

## Evidence (증거 — 이번 런에서 실측한 명령 + 출력)

본 디렉터리 첨부물 (전부 이번 런에서 `gh run view <run> --log-failed` / 로컬 `go test` 산출):

- `ci_failed.log` (run 33382844571): 6개 실패 확인. `grep -c 'moai-adk v3.1.4'` = 6, `grep -c 'moai-adk v3.1.3'` = 6 — 6테스트 × (got 1 + want 1), 집계 수치 개입 0.
- `doctor_dark_got.txt` / `doctor_dark_want.txt`: 로그 697-744행 / 745-792행에서 타임스탬프 프리픽스를 벗겨 추출한 got·want 블록. `diff` 결과 **1 hunk, 17행뿐**:
  ```
  17c17
  < │    ok      MoAI Version           moai-adk v3.1.4   │
  ---
  > │    ok      MoAI Version           moai-adk v3.1.3   │
  ```
  나머지 45행(집계 4곳 포함) 바이트 동일. status-dark도 로그 993-1025행에서 동일 패턴 육안 확인(`- **ADK**: moai-adk v3.1.4` vs `v3.1.3`, 그 외 전부 동일).
- `release_verify_failed.log` (run 33382844521): 실패 분포 — **macos 6 / ubuntu 6 / windows 154**. ubuntu·macos의 실패 목록은 `grep -o 'Test[A-Za-z_0-9]*'`로 추출 시 golden 6개와 정확히 일치 → **Release Verify (ubuntu·macos)는 스탬프 축의 하류**. windows 154개는 별개 축(하류 아님): internal/cli 119 · internal/kanban 22 · internal/graph 6 · internal/statusline 5 · internal/web 2. 참고: Windows 런타임 테스트는 release/* PR에서만 도는 ci.yml의 의도적 절충(ci.yml 주석, t10 카드 기록)이라 develop에선 안 보이던 부채가 릴리스 게이트에서 집계 노출된 것.
- Multi-OS Gate: `##[error]Release PR multi-OS verification FAILED on at least one OS` — Release Verify들의 집계 게이트 → 하류.
- `graph_freshness_failed.log` (run 33382844542): `##[error]codemaps provenance stamp 7fc0af324cf47f65d343ac3f936ca4f22bde9c51 is NOT an ancestor of PR base origin/main — orphan-bound stamp`. 실측: `git merge-base --is-ancestor 7fc0af324 origin/main` rc=1, `--is-ancestor 7fc0af324 9328a5242` rc=0 — develop 전용 머지 커밋("merge(develop): absorb fef7a4b9b for batch cleanup", 2026-08-31 15:45)이 codemaps provenance에 스탬프돼 있는데 릴리스 PR의 베이스는 main이라 실패. **golden 축과 무관한 별도 적색 축.**
- **처방 검증 (로컬)**: 워크트리 내 스크래치 브랜치( origin/release/v3.1.4 기반)에서 `git cherry-pick 96bfa0c99` → 충돌 없이 `d601bae16` (8 files, +51/-6) → `go test -run 'TestDoctorGolden|TestStatus' -count=1 ./internal/cli/` → **6개 전부 PASS**. pick 적합성 사전 확인: `git diff --quiet origin/release/v3.1.4 96bfa0c99^ -- internal/cli/doctor_golden_test.go internal/cli/status_golden_test.go internal/cli/testdata/` rc=0 (8파일 내용이 pick 부모와 바이트 동일 → clean pick 보장). 스크래치 브랜치는 검증 후 폐기했고 릴리스 브랜치는 건드리지 않았다.
- develop 베이스(본 카드 브랜치)에서 동일 6테스트 = PASS (고정 코드 존재) — 실패가 릴리스 브랜치 상태에만 존재함을 확인.

## Baseline-attribution (귀속)

- RED 근거: CI run 33382844571 / 33382844521 / 33382844542, head `26898312e` (`gh run view --json headSha` = 26898312eee07e9ff28d635a05b074594713e2d3, conclusion failure, event pull_request).
- GREEN 근거(처방 검증): 본 워크트리, 스크래치 브랜치 `origin/release/v3.1.4 + 96bfa0c99 cherry-pick`, `go test -run 'TestDoctorGolden|TestStatus' -count=1 ./internal/cli/` 2026-09-02 로컬 실행, 결과 `--- PASS` 6/6.
- 로컬 GREEN은 **최종 회귀 판정이 아니다**(카드 mutant 조항). 최종 판정은 리드가 릴리스 브랜치에 pick을 적용·push한 뒤 그 조용한 head의 CI에서 내는 것.

## 처방 (리드 실행 항목)

```
git -C <release-체크아웃> cherry-pick 96bfa0c99   # 8파일, 충돌 없음 (본 판정문이 검증)
git push origin release/v3.1.4                     # 이후 CI 재판독
```
- pick 커밋 메시지에 t406 참조를 덧붙이는 것 권장(추적성 — AGENTS.md §3 3중 운반기).
- CI 재판독 시 축별 기대: golden 6 → Test(ubuntu)·Release Verify(ubuntu/macos) 소멸, Multi-OS Gate는 windows·graph-freshness 때문에 여전히 적색일 수 있음(하류 집계 게이트).
- 이후 남는 적색: **windows 154**(별개 축 — 별도 카드 권장: internal/cli 119가 본체) · **graph-freshness**(codemaps provenance가 릴리스 브랜치에서 develop 전용 스탬프를 가짐 — provenance 재생성 또는 릴리스 PR 베이스 특례, 별도 카드).

## Gaps (명시적으로 관측하지 않은 것)

- windows 154개 실패의 개별 원인 — 미규명(본 카드 축 1 밖). 개별 원인은 별도 카드 소관.
- graph-freshness의 처방 선택(provenance 재생성 vs 가드 특례) — 판단만 하고 어떤 쪽이 옳은지는 미검증(별도 카드 소관).
- Release Verify (windows) 154개 중 golden 6을 제외한 148개가 macos/ubuntu에서 동반 실패하지 않는 이유(플랫폼 고유 vs 환경) — 분류만 하고 원인 미규명.
- cherry-pick 후 CI에서의 실제 GREEN — 리드 push 전까지 존재하지 않는 증거.

## Residual-risk (잔여 위험)

- cherry-pick은 golden 6을 고치지만 windows 154는 그대로다. Multi-OS Gate·Release PR의 최종 초록은 windows 축·graph-freshness 축이 별도로 수리돼야 한다.
- 릴리스 브랜치가 develop을 흡수(merge origin/develop)하면 96bfa0c99 본체와 pick 쌍둥이가 함께 오지만 내용 동일이라 충돌 없음 — 다만 그 흡수 자체는 다른 대상이므로 본 카드에서 다루지 않았다.
- `96bfa0c99` 이후 develop에서 해당 8파일이 다시 바뀌었다면 pick 쌍둥이와 미세 갈림이 생길 수 있다 — 본 판정 시점 기준 develop tip(`9145806d8`)에서 8파일은 pick 결과와 동일함을 확인(로컬 테스트 PASS로 간접 확인).
