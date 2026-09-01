# SPEC-VERSION-STAMP-PREDICATE-001 — 진행 기록

카드 t392 · Tier M · 워크트리 `.claude/worktrees/t392`(브랜치 `WT-version-stamp-predicate`)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-01
tier: M
plan_audit_iteration: 3
plan_audit_iteration_note: "iteration 3 exists under an operator override of the Tier M two-iteration ceiling"
requirements: 15
acceptance_criteria: 15
artifacts:
  - .moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/spec.md
  - .moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/plan.md
  - .moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/acceptance.md
  - .moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/progress.md
measurement_tree: 9a3e2dabe9ab12a4a7313db2d8ab5c0247b24bb4
spec_id_regex_check: PASS
```

plan-phase에서 반증된 전제 1건: 운영자 지시가 지목한 `origin/release/v3.1.4` 는 이 검사의
RED이 **아니다**(그 트리의 권위 토큰은 `v3.1.4` 이고 낡은 golden 은 `v3.1.3` 을 담으므로
현재-토큰 스윕에 걸리지 않는다). 완결성 단언의 RED은 pin 이전의 이 트리에 있다.
근거: `spec.md` §1.1, `plan.md` §D.

### iter-2 (plan-audit FAIL 0.79 → 수리)

`.moai/reports/t392/plan-audit.md` 의 D1-D8에 답했다. 설계가 바뀐 것 둘:

1. **스윕 모집단 = 추적 파일 집합**(`git ls-files`). 파일시스템 walk를 택하지 않은 이유와
   그 결과를 실측으로 기록했다(`spec.md` §2.0). REQ-VSP-010이 「git 호출 없음」에서
   「이력·타 ref·네트워크 없음」으로 좁혀졌다.
2. **「스윕 개수 = 28」 상수 제거.** 범프 numstat 실측(`eba919e44` 6파일 전부 스탬프)으로
   그 상수가 범프마다 무너질 뿐 아니라 **애초에 틀렸음**을 확인했다 — 서술이 옛 릴리스를
   인용한 상태는 이 SPEC이 정상이라고 판정한 상태다. 상수는 등록부의 성질 둘(28 · 7)로
   옮겼고, 감사가 D5로 지목한 M4→M5 파손은 그 변경으로 **소멸**했다(`plan.md` §D.0).

plan-phase에서 반증된 감사 주장 1건: 미추적 토큰 보유 파일은 6이 아니라 **7**이며 전부 이
카드 자신의 산출물이다(`comm -23` 실측, `acceptance.md` AC-VSP-012). 형제 워크트리 수도
181이 아니라 **183**이다. 둘 다 감사 보고 자신이 예고한 드리프트다.

### iter-3 (plan-audit PASS-WITH-DEBT 0.84 → 수리, **운영자가 2회 상한 해제**)

[HARD] Tier M은 plan-audit 2회가 상한이고 iter-2가 그 두 번째였다. 이 라운드는 **운영자가
이번 한 번에 한해 해제해** 성립한 것이며 통상 경로가 아니다. `plan-audit-iter2.md` 의
N1-N6 · D6에 답했다.

설계가 바뀐 것 하나: **REQ/AC-VSP-015 — 스윕 도달 범위.** iter-2의 뺄셈(「스윕 개수 = 28」
제거)이 옳았지만, 그 상수가 스윕이 얼마나 넓게 닿았는가에 대한 **유일한 하한**이기도 했다.
크기 리터럴(감사가 제시한 10,048)은 범프에 불변이나 **개발에 불변이 아니어서** 채택하지
않고, 양변을 실행 시점에 얻는 단언 둘로 대체했다: 등록부 ⊆ 모집단 · 판정 수 = 넘긴 수
(`spec.md` §6.2). 이 둘이 못 덮는 한 칸은 §8 R-9로 이름 붙여 열어 뒀다.

나머지는 산출물 편집이다. 뿌리가 하나인 결함 둘(N2 · N3)이 있었다 — **영문 문서에 한국어
문안과 한국어 판정 grep을 못박고 있었다**(`grep -cP '[가-힣]' .moai/docs/version-management.md`
→ 0). `plan.md` §E를 영문으로 재작성하고, 판정 구절을 §E 문안에만 존재하는 영문 구절로 다시
못박았으며(현재 전부 0건 = 미리 못박은 RED), 닫힌-수 정규식을 구조형으로 바꿔 뮤턴트 여덟에
전부 걸리고 열린 문안에 0건임을 이 세션에서 실측했다.

iter-3 판정: **PASS-WITH-DEBT 0.90**(보고서 `.moai/reports/t392/plan-audit-iter3.md`, opus에서 작성 후 weekly-limit 중단으로 디스크에서 복구됨) — D1(015 카운트 키 이중 고정 `examined_of`/`handed` + §D.2 행 부재; 정본 키 `handed`로 통일)·D2(`현 L83·L90` → `현 L82·L88` 좌표 정정)는 이번 라운드에서 수리했고, D3-D5는 감사자 권고에 따라 선택적 부채로 남긴다.

## §E.2 Run-phase Evidence

모든 측정은 이 워크트리(브랜치 `WT-version-stamp-predicate`)에서 실행됐다. 작업 시작 시점 HEAD
`051f209b0`(plan-phase 최종 커밋). 중간에 GLM 429로 한 차원 중단됐고 리드의 재개 지시로 재앵커링
했다 — 중단 시점은 006 뮤턴트 관측 직후였고, 그 관측 자체는 중단 전에 완료돼 있었다(아래 기록).
재개 후 `git status`로 리드의 측정과 트리 상태가 일치함을 확인한 뒤 이어갔다.

### M0 — 착수 전 재측정 (HEAD `051f209b0`)

```
$ grep -n 'Version = ' pkg/version/version.go
8:	Version = "v3.1.3"
```

거부 목록 적용 스윕: **34** (plan 핀 `9a3e2dabe`와 동일 — 불변 확인). 34 = 스탬프 7 + golden 6 +
서술 21. `git ls-files | wc -l` → 10,062.

그룹별 감춰진 파일(권위 토큰 `v3.1.3` 기준, 그룹마다 `git grep -lF v3.1.3 -- <그룹> | wc -l`):
`.moai/reports/` 61 · `.moai/specs/` 62 · `.moai/release-notes/` 1 · `CHANGELOG.md` 1 ·
`*_test.go` 4 · changelog-pages 0. 합 **129**. 합집합을 한 pathspec으로 재측정 → **129**(서로소,
항등식이 아니라 측정으로). 전체 추적 토큰 보유 **163**. 귀속 확인: **129 + 34 = 163** ✓.

plan 핀(`9a3e2dabe`)에서의 값(57/58/121/155)과의 차분은 plan-phase 커밋 4개가 이 카드 자신의
산출물을 제외 그룹 안에 추가한 것 — `plan.md` §B가 **미리 선언한 규칙** 그대로다. 정수 예측은
두지 않았고 34의 불변과 귀속 등식만 확인했다. **상수(등록부 28 · stamp 7) 영향 없음.**

부수 확인: `git ls-files -s | awk '$1 == "160000"' | wc -l` → **0**(서브모듈 gitlink 없음 —
판정 수=넘긴 수가 clean 트리에서 성립). golden 6개 각각 `v3.1.3` 출현 1회. 문서 한글 0건(영문
전용 전제 유지).

### M1 — 등록부 자료 모형 (커밋 `3c7b55adf`)

`internal/cli/version_stamp_registry_test.go`: 분류 어휘 `stamp`/`prose` 둘, 정확 경로 28항목
리터럴, 상수 둘(`expectedRegistryEntries = 28` · `expectedStampEntries = 7` — 서로 독립 보유,
스윕 개수 상수 없음), 형태 가드 `TestVersionStampRegistryShape`(개수·클래스·exact-path·유일성).
같은 커밋에서 spec.md `status: draft → in-progress` 전이.

### M2 — 순수 판정 코어 + 합성 RED 일곱 (커밋 `043e21033`)

**RED 관측(구현 전, 빈 코어에 대해 한 번의 실행):**

```
$ go test ./internal/cli/ -run 'TestVersionStampSynthetic|TestVersionStampSweepByContent' -count=1 -v
=== RUN   TestVersionStampSyntheticFreshness
=== RUN   TestVersionStampSyntheticFreshness/catches_a_stale_stamp
    version_stamp_registry_test.go:251: check did not emit expected failure: registered stamp does not carry the authoritative token: docs-site/hugo.toml (got 0 findings: [])
=== RUN   TestVersionStampSyntheticFreshness/exempts_a_stale_prose_entry
--- FAIL: TestVersionStampSyntheticFreshness (0.00s)
    --- FAIL: TestVersionStampSyntheticFreshness/catches_a_stale_stamp (0.00s)
    --- PASS: TestVersionStampSyntheticFreshness/exempts_a_stale_prose_entry (0.00s)
=== RUN   TestVersionStampSyntheticVacuity
=== RUN   TestVersionStampSyntheticVacuity/reports_an_empty_sweep_naming_every_stamp
    version_stamp_registry_test.go:285: check did not emit expected failure: registered stamp missing from sweep: README.md
    version_stamp_registry_test.go:285: check did not emit expected failure: registered stamp missing from sweep: README.ko.md
    version_stamp_registry_test.go:285: check did not emit expected failure: registered stamp missing from sweep: README.ja.md
    version_stamp_registry_test.go:285: check did not emit expected failure: registered stamp missing from sweep: README.zh.md
    version_stamp_registry_test.go:285: check did not emit expected failure: registered stamp missing from sweep: .moai/config/sections/system.yaml
    version_stamp_registry_test.go:285: check did not emit expected failure: registered stamp missing from sweep: docs-site/hugo.toml
    version_stamp_registry_test.go:285: check did not emit expected failure: registered stamp missing from sweep: pkg/version/version.go
=== RUN   TestVersionStampSyntheticVacuity/holds_the_registry_entry_count
    version_stamp_registry_test.go:300: check did not emit expected failure: registry entries=27 expected=28 (got [])
=== RUN   TestVersionStampSyntheticVacuity/holds_the_stamp_classification_count
    version_stamp_registry_test.go:321: check did not emit expected failure: stamp entries=6 expected=7 (got [])
--- FAIL: TestVersionStampSyntheticVacuity (0.00s)
    --- FAIL: TestVersionStampSyntheticVacuity/reports_an_empty_sweep_naming_every_stamp (0.00s)
    --- FAIL: TestVersionStampSyntheticVacuity/holds_the_registry_entry_count (0.00s)
    --- FAIL: TestVersionStampSyntheticVacuity/holds_the_stamp_classification_count (0.00s)
=== RUN   TestVersionStampSweepByContent
    version_stamp_registry_test.go:343: sweep must hold only the content carrier, got []
    version_stamp_registry_test.go:346: judged=0 handed=2
--- FAIL: TestVersionStampSweepByContent (0.00s)
=== RUN   TestVersionStampSyntheticDocCrossCheck
=== RUN   TestVersionStampSyntheticDocCrossCheck/names_a_stamp_the_document_dropped
    version_stamp_registry_test.go:368: check did not emit expected failure: stamp set differs from documentation list: pkg/version/version.go (got [])
=== RUN   TestVersionStampSyntheticDocCrossCheck/passes_with_prose_entries_the_document_never_lists
--- FAIL: TestVersionStampSyntheticDocCrossCheck (0.00s)
    --- FAIL: TestVersionStampSyntheticDocCrossCheck/names_a_stamp_the_document_dropped (0.00s)
    --- PASS: TestVersionStampSyntheticDocCrossCheck/passes_with_prose_entries_the_document_never_lists (0.00s)
=== RUN   TestVersionStampSyntheticGhost
=== RUN   TestVersionStampSyntheticGhost/names_a_registry_entry_with_no_file_behind_it
    version_stamp_registry_test.go:397: check did not emit expected failure: registry entry does not resolve to a file: docs-site/content/en/retired-page.md (got [])
    version_stamp_registry_test.go:400: check did not emit expected failure: registry path missing from population: docs-site/content/en/retired-page.md (got [])
=== RUN   TestVersionStampSyntheticGhost/passes_when_every_entry_resolves
--- FAIL: TestVersionStampSyntheticGhost (0.00s)
    --- FAIL: TestVersionStampSyntheticGhost/names_a_registry_entry_with_no_file_behind_it (0.00s)
    --- PASS: TestVersionStampSyntheticGhost/passes_when_every_entry_resolves (0.00s)
=== RUN   TestVersionStampSyntheticPopulationReach
=== RUN   TestVersionStampSyntheticPopulationReach/names_a_registry_path_missing_from_the_population
    version_stamp_registry_test.go:435: check did not emit expected failure: registry path missing from population: .moai/docs/version-management.md (got [])
=== RUN   TestVersionStampSyntheticPopulationReach/reports_a_handed_path_the_core_did_not_examine
    version_stamp_registry_test.go:445: check did not emit expected failure: judged=28 handed=29 (got [])
=== RUN   TestVersionStampSyntheticPopulationReach/stays_silent_on_a_consistent_population
--- FAIL: TestVersionStampSyntheticPopulationReach (0.00s)
    --- FAIL: TestVersionStampSyntheticPopulationReach/names_a_registry_path_missing_from_the_population (0.00s)
    --- FAIL: TestVersionStampSyntheticPopulationReach/reports_a_handed_path_the_core_did_not_examine (0.00s)
    --- PASS: TestVersionStampSyntheticPopulationReach/stays_silent_on_a_consistent_population (0.00s)
=== RUN   TestVersionStampSyntheticTracking
=== RUN   TestVersionStampSyntheticTracking/an_untracked_carrier_stays_invisible
=== RUN   TestVersionStampSyntheticTracking/a_tracked_carrier_is_named
    version_stamp_registry_test.go:479: check did not emit expected failure: unregistered file carries the authoritative token: notes/stray-version-note.md (got [])
--- FAIL: TestVersionStampSyntheticTracking (0.00s)
    --- PASS: TestVersionStampSyntheticTracking/an_untracked_carrier_stays_invisible (0.00s)
    --- FAIL: TestVersionStampSyntheticTracking/a_tracked_carrier_is_named (0.00s)
```

§D.2 핀 문자열이 일곱 RED 전부의 실패 줄에 그대로 나타난다. 개수 키는 **`handed`**(D1 폐쇄대로
`judged=<n> handed=<n>` 단일 키). 013의 합성 상태는 유령 단언과 도달 단언이 **구성상 함께**
우는 상태라(내용 없는 경로는 모집단에도 없고 내용 맵에도 없다) 양쪽 기대 줄을 함께 단언했다 —
한 RED를 다른 RED의 증거로 쓰지 않고, 상태의 성질을 테스트 주석에 적었다. 통과 방향(stale prose
면제 · prose 문서 미기재 무시 · 전 항목 해소 · 일관 모집단 침묵 · 미추적 무시)은 **같은 실행에서**
전부 PASS.

**GREEN(동일 명령, 구현 후):** `ok github.com/modu-ai/moai-adk/internal/cli 0.910s` — 7개
테스트 함수 + 양방향 서브테스트 전부 PASS. 캐치 단언은 기대 finding 문자열과의 **완전 일치**
(포함이 아니라)로 검증한다 — 검사가 §D.2 문자열을 정확히 내는지가 이 실행에서 증명된다.

**006 뮤턴트(내용→경로 판정 전환):** 첫 시도에서 합성 경로명(`RELEASE-NOTES-v2.17.0.md`)이
현재 토큰을 담지 않아 경로-판정 뮤턴트가 집지 못했다(빈 스윕 RED). 이는 테스트의 허점이지 구현의
허점이 아니므로 **테스트를 강화**했다 — 이름에 현재 토큰을 담은 경로와 옛 토큰만 담은 경로를
둘 두어 경로-판정 뮤턴트와 일반-형태 뮤턴트 양쪽을 다 잡게 했다. 강화 후 뮤턴트:

```
--- FAIL: TestVersionStampSweepByContent (0.00s)
    version_stamp_registry_test.go:476: path named with a token but carrying none must not enter the sweep: RELEASE-NOTES-v9.9.9-synthetic.md
    version_stamp_registry_test.go:480: sweep must hold only the content carrier, got [RELEASE-NOTES-v9.9.9-synthetic.md]
```

복원 후 GREEN 재확인. AC-VSP-006 RED 셀의 `RELEASE-NOTES-v2.17.0.md 류`는 강화된 테스트에서
`agedName` 상수로 그대로 유지된다.

### M3 — 모집단 드라이버 + 실트리 완결성 RED (커밋 `4462f7788`)

드라이버: `gitLsFilesArgv = []string{"git", "ls-files"}` — 파일에 **정확히 하나**의 이름 붙은
argv 리터럴(AP-12). 제외 그룹 여섯 리터럴 + D4 매처(아래 결정 참조). 권위 토큰을
`pkg/version/version.go` 작업 트리 값에서 추출.

**AC-VSP-003 RED — pin 이전 트리(이 커밋)에서 관측:**

```
$ go test ./internal/cli/ -run 'TestVersionStampRegistry$' -count=1
--- FAIL: TestVersionStampRegistry (0.26s)
    version_stamp_registry_test.go:818: unregistered file carries the authoritative token: internal/cli/testdata/doctor-dark.golden
    version_stamp_registry_test.go:818: unregistered file carries the authoritative token: internal/cli/testdata/doctor-light.golden
    version_stamp_registry_test.go:818: unregistered file carries the authoritative token: internal/cli/testdata/doctor-nocolor.golden
    version_stamp_registry_test.go:818: unregistered file carries the authoritative token: internal/cli/testdata/status-dark.golden
    version_stamp_registry_test.go:818: unregistered file carries the authoritative token: internal/cli/testdata/status-light.golden
    version_stamp_registry_test.go:818: unregistered file carries the authoritative token: internal/cli/testdata/status-nocolor.golden
FAIL
```

golden 6경로만 이름으로, 개수 단언은 침묵 — plan §M3의 설계대로다(우는 것은 완결성 단언 하나).

**파일 수준 판정(AC-VSP-010 a/b · AC-VSP-012):**

```
$ awk '/^func versionStampSweep/,/^}/' <file> | grep -cE '\bexec\.|os/exec'   → 0
$ awk '/^func judgeVersionStampRegistry/,/^}/' <file> | grep -cE '\bexec\.|os/exec'  → 0
$ grep -c '"git"' <file>                       → 1
$ grep -cF '[]string{"git", "ls-files"}' <file> → 1
$ grep -cE 'filepath\.Walk|WalkDir|ReadDir' <file> → 0
```

**AC-VSP-010 뮤턴트 배터리(셋 전부 — 넷째는 무수정 통과 방향):**

1. argv → `[]string{"git", "ls-tree", "-r", "HEAD"}`: 허용 목록 grep(`[]string{"git", "ls-files"}`) → **0건**(rc=1) — 판정 실패로 적발. **iter-2 거부 목록이 통과시켰던 바로 그 뮤턴트**다.
2. 코어에 `exec.Command("git", "show", "HEAD", ...)` 한 줄 추가: 코어 블록 grep → **1건**(acceptance 예측의 "첫째 grep 2건"은 파일 전체 grep 기준값 — 코어 블록 한정 판정에서는 1건이고, 파일 전체 `"git"` 행 수는 **2건**으로 (b)도 함께 실패. 양 판정 모두 발화).
3. ls-files 호출을 코어로 이동: 코어 블록 grep → **1건** — (a) 적발.
4. 무수정 원본: (a) 0건 · (b) 1건 · 1건 — **통과 방향**.

**AC-VSP-015 뮤턴트 배터리(셋 전부, 실트리 검사로 관측):**

1. 드라이버를 `internal/cli`로 좁힘: `registry path missing from population` **28건**(등록부 전체가 이름으로 지목됨).
2. 제외 그룹에 `docs-site/` 한 줄 추가: 같은 단언 **21건**(서술 20 + 스탬프 `docs-site/hugo.toml`).
3. 드라이버가 앞 50개만 넘기도록 절단: 같은 단언 **28건**.
   셋 다 네 단언이 초록인 iter-2 설계에서는 통과했을 상태들이다.
   [수리 기록] 뮤턴트 2의 되돌리기 편집에서 U+200B(ZERO WIDTH SPACE)가 `*_test.go` 행에
   유입됐다 — od로 바이트 확인(0342 0200 0213) 후 행 번호 기준 perl 교체로 제거했고,
   `git diff`가 비어 **HEAD와 바이트 동일 복원**임을 확인했다.

**AC-VSP-012 양방향** — 합성 모집단 둘로 색인을 건드리지 않고 관측(M2의 Tracking 테스트:
미추적 무시 PASS + 추적 시 이름 적발). 판정 grep(WalkDir/ReadDir 0건)은 위와 같다.

**run-phase 결정 둘(iter-3가 열어 둔 것):**

- **D5(순수 코어 함수명 핀):** `versionStampSweep` + `judgeVersionStampRegistry`. (a) 판정은
  위처럼 awk 행 범위 한정 grep으로 실행한다.
- **D4(제외 집합 매처 의미론):** git pathspec 의미론으로 대조한다 — 끝이 `/`면 경로 접두사,
  별표 없는 이름은 정확 일치, `*`는 **세그먼트를 넘는다**(정규식 `.*` 환산, 컴파일 캐시).
  근거: 제외 수치의 측정 도구가 git pathspec이므로 검사가 같은 집합을 봐야 측정-판정 갈림이
  원천 차단된다. `filepath.Match`의 세그먼트 한정 `*`은 측정이 센 중첩 `_test.go`를 놓치므로
  기각. `TestVersionStampExclusionMatcher`가 접두사/정확/별표와 -h 함정 형태(`.moai/release/`
  아래 이름-토큰 파일은 제외되지 않음)를 표로 고정 — PASS.

### M4 — pin + 픽스처 재생성 (커밋 `96bfa0c99`)

`status_golden_test.go` · `doctor_golden_test.go`의 golden 테스트 함수 **여섯 각각**에 pin
(선례 `version_test.go:180-186` 형태: 원값 보존 → `version.Version = "v0.0.0-test"` 대입 →
`defer` 복원). 헬퍼 함수가 아닌 테스트 함수마다 넣은 이유: `captureStatusCmdWithPkgDir`/
`captureDoctorCmd`는 golden이 아닌 테스트 3개(status_specview_render, stdout_clean,
doctor_render)가 공유하므로 헬퍼 pin은 영향 반경이 SPEC 밖으로 나간다.

재생성 + **diff 판독**(AP-5):

```
$ git diff --stat
 internal/cli/doctor_golden_test.go          | 22 +++++++++++++++++++++++
 internal/cli/status_golden_test.go          | 23 +++++++++++++++++++++++
 internal/cli/testdata/doctor-dark.golden    |  2 +-
 internal/cli/testdata/doctor-light.golden   |  2 +-
 internal/cli/testdata/doctor-nocolor.golden |  2 +-
 internal/cli/testdata/status-dark.golden    |  2 +-
 internal/cli/testdata/status-light.golden   |  2 +-
 internal/cli/testdata/status-nocolor.golden |  2 +-
```

전체 diff를 읽었다 — 6개 golden 전부 **정확히 1행**(버전 행)만 변경: doctor 3개는
`moai-adk v3.1.3` → `moai-adk v0.0.0-test`(MoAI Version 행, `ok` 상태 유지 — **ok/warn 집계
수 불변**, release/v3.1.4 사고 형태 없음), status 3개는 `- **ADK**: moai-adk v3.1.3` →
`v0.0.0-test`. 그 외 무변화.

```
$ go test ./internal/cli/ -run 'TestVersionStampRegistry$' -count=1 -v   → PASS (M3 RED flip)
$ grep -c 'v3\.1\.3' internal/cli/testdata/*.golden                       → 0 0 0 0 0 0
```

M3·M4를 **같은 push**로 올렸다(§D.1 — `git push -u origin WT-version-stamp-predicate`, M1-M4
포함, 신규 브랜치 생성 확인).

### M5 — 문서 절 (커밋 `8c71cd423`)

`.moai/docs/version-management.md`: 거부 목록 열거 표(여섯 그룹 · 사유 · 감춘 파일 수, 트리
`051f209b0` 핀 + 재도출 방아쇠 문장) + plan §E 영문 교체 문안(두 검사 · 열린 열거 6항목 ·
`this list is not exhaustive` 표지 · `None of this means the list can no longer rot.` ·
등록부 유지 계약). 서술 21경로는 문서에 옮겨 적지 않았다.

**판정 명령 관측(모두 이 세션 실행):**

```
AC-VSP-011(b) 7구절: aged-out token 1 · registered as `prose` 1 · inlined inside a file the
  exclusion set hides 1 · not a version token at all 1 · renders the version rather than
  carrying it 1 · the repository does not track 1 · this list is not exhaustive 1
AC-VSP-011(d) 닫힌-수 정규식 → 0건(rc=1) · 뮤턴트 8종(re2.txt/mut8.txt) → **8/8 적중**
AC-VSP-014   version_stamp_registry_test.go 1 · edits the registry in that same commit 1 ·
  the check fails naming the path 1 · A version bump does not touch the registry 1 ·
  docs-site/content 유출 → **0건**
영문 전용: grep -cP '[가-힣]' → 0
AC-VSP-007   표 데이터 행 6(헤더 제외) = 검사의 리터럴 그룹 6 · 트리 SHA `051f209b0` 표 근처 1건
AC-VSP-014 양방향 뮤턴트: 문서 사본에 docs-site/content 경로 1행 추가 → 다섯째 판정 **1건으로 발화** ✓
```

M5 이후 스윕은 27(문서가 토큰 행을 잃어 스윕에서 빠짐 — 등록된 prose이므로 §D.0대로 무위반)이고
`t388 검사 + t392 검사` 동시 실행 PASS.

### 최종 검증 (HEAD `8c71cd423`)

```
$ go test ./internal/cli/... -count=1
ok  github.com/modu-ai/moai-adk/internal/cli        286.774s
ok  ... (17패키지 전부 ok)  — FAIL 0
$ go vet ./internal/cli/ → 정상 (각 마일스톤 커밋 전에도 실행)
$ awk '/^var versionStampRegistry/,/^}/' <file> | grep -cE '"[^"]*[*?\[]' → 0  (AC-VSP-002)
```

M0 수치와 상수는 최종 트리에서 재확인: 34(불변) · 등록부 28 · stamp 7 — 병합 트리 재측정(R-7)은
통합 창의 몫.

**Gaps:** CI 판정(origin/develop과 동일 조건의 전체 스위트·매트릭스)은 이 세션 밖 — push된
브랜치 head의 CI를 리드가 읽어야 한다. R-9의 잔여(등록부 경로만 남기는 모집단)는 설계상 열려
있고 이 run은 그것을 잠그지 않았다(SPEC이 주장하지 않은 것과 동일).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-09-01
tier: M
requirements: 15
acceptance_criteria: 15
measurement_tree_run_start: 051f209b0 (plan-phase final commit — not the 9a3e2dabe plan pin; the card's own plan commits shifted the tree)
branch: WT-version-stamp-predicate
commits:
  - "3c7b55adf feat(t392): M1 — version-stamp registry data model and classification vocabulary"
  - "043e21033 feat(t392): M2 — pure judgment core and seven synthetic REDs"
  - "4462f7788 feat(t392): M3 — tracked-files population driver and the real-tree check"
  - "96bfa0c99 feat(t392): M4 — pin golden fixtures off the version predicate"
  - "8c71cd423 docs(t392): M5 — sweep exclusion enumeration and the partial-guarantee replacement"
pushed: origin/WT-version-stamp-predicate (M1-M4 @ 96bfa0c99 확인; M5+증거는 최종 push)
red_observations: synthetic 7 (M2) + real completeness 6-path (M3, pre-pin tree) — §D.2 문자열 verbatim
mutant_batteries_complete:
  ac_vsp_010: "4/4 (allowlist + core grep + whole-file git count + clean pass)"
  ac_vsp_011: "8/8 (closed-count regex mutants)"
  ac_vsp_015: "3/3 (scoped driver / over-broad exclusion / truncated driver)"
  ac_vsp_006: "1/1 (path-based membership mutant)"
  ac_vsp_014_leak: "1/1 (prose-path leak fires the fifth judgment)"
run_decisions:
  d5_pure_core_names: "versionStampSweep, judgeVersionStampRegistry (awk-bounded block grep)"
  d4_exclusion_matcher: "git pathspec semantics — trailing slash prefix / exact / segment-crossing star"
message_contract: "plan.md §D.2 — count key handed, single key across all sites"
evidence_paths:
  - ".moai/reports/t392/scratch/re2.txt (tracked, plan-phase afed3c8f8)"
  - ".moai/reports/t392/scratch/mut8.txt (tracked, plan-phase afed3c8f8)"
remaining_for_lead: "origin CI verdict on the pushed head; develop integration window; sync-phase"
```

run-phase에서 반증·수정된 전제 1건: 합성 006 테스트의 초기 경로명이 현재 토큰을 담지 않아
경로-판정 뮤턴트를 집지 못했다 — 구현 결함이 아니라 테스트 결함이었고, 강화 후 뮤턴트가 적발된다
(위 006 뮤턴트 기록). 운영 사고 1건: 015 뮤턴트 2 되돌리기 중 U+200B 유입 — 바이트 확인 후
제거, `git diff` 공복로 복원 검증(위 M3 절).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-09-01
tier: M
requirements: 15
acceptance_criteria: 15
sync_commit_sha: pending-backfill-sync
sync_commit_sha_note: |
    A commit cannot cite its own hash. Two-step backfill per the dispatch and
    the D3 SHA-placeholder exemption: this sync commit lands first with the
    placeholder, and the real SHA is backfilled into this field in the
    immediately following commit on this branch. No push, no PR — develop
    integration is the lead's window.
branch: WT-version-stamp-predicate
changelog_entry_position: "CHANGELOG.md [Unreleased] > ### Added, first entry (inserted above SPEC-MEMORY-STORE-RECONCILE-001)"
b12_self_test_a: "pre-emission grep — `grep -c 'SPEC-VERSION-STAMP-PREDICATE-001' CHANGELOG.md` = 0 before writing (no duplicate from a parallel session)"
b12_self_test_b: "AC count — `grep -c '^### AC-VSP-' acceptance.md` = 15, all live (no [RETIRED]/[REF] marks), matching spec.md §4.1 (요구 15 / 수락 15); the CHANGELOG entry states 15/15"
b12_self_test_c: "file-path verification — every path named in the CHANGELOG entry resolves in this tree (internal/cli/version_stamp_registry_test.go, internal/cli/testdata/*.golden ×6, .moai/docs/version-management.md, pkg/version/version.go, docs-site/hugo.toml, .moai/config/sections/system.yaml, README×4)"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (single sync commit, 3-phase close)"
  plan_md: "none — no `status:` field (ArtifactStatusFieldForbidden, card t357)"
  acceptance_md: "none — no `status:` field"
  progress_md: "none — no frontmatter"
  updated_field: "spec.md `updated:` already 2026-09-01 (sync-commit date); unchanged, not re-stamped"
sync_phase_files_touched_count: 3
sync_phase_files_touched:
  changelog: "CHANGELOG.md — [Unreleased] > Added entry (repo convention: one prose entry per sync-phase close)"
  spec_frontmatter: ".moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/spec.md — `status:` only, zero body edits"
  progress: ".moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/progress.md — this §E.4 block only"
mx_tag_changes:
  added_count: 0
  detail: "docs-only sync — no production code touched this phase; annotation state of the run-phase test files is run-phase's record, not edited here."
docs_surfaces_this_phase: ".moai/docs/version-management.md untouched by sync — its partial-guarantee replacement was run-phase M5 (8c71cd423); this sync only describes it in the CHANGELOG entry."
readme_4_locale: "README{,.ko,.ja,.zh}.md untouched — they are registered stamp entries and this card changes no version (dispatch HARD constraint)."
remaining_for_lead: "origin CI verdict on the pushed head; develop integration window (sync made no push)"
```

sync-phase에서 반증·수정된 전제: 없다. plan/run이 남긴 수치와 상수(등록부 28 · stamp 7 ·
스윕 34 불변)를 이 단계에서 다시 건드리지 않았고, 이 커밋의 전부는 CHANGELOG 항목 · 이
§E.4 블록 · spec.md 프론트매터 전이 셋이다. 병합 트리에서의 재측정(R-7)은 통합 창의 몫으로
남는다.
