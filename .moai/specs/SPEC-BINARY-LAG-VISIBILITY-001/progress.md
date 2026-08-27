---
id: SPEC-BINARY-LAG-VISIBILITY-001
title: "배포 지연 가시성 — 진행 기록"
version: "0.4.0"
status: in-progress
created: 2026-08-27
updated: 2026-08-28
author: manager-spec
priority: High
phase: "v3.1.4 target"
module: internal/cli, internal/hook, build
lifecycle: spec-anchored
tags: "deployment-lag, doctor, session-start, version-stamp, observability, fail-open"
tier: M
---

# SPEC-BINARY-LAG-VISIBILITY-001 — 진행 기록

카드: t326 · 워크트리: `.claude/worktrees/t326` · 브랜치: `WT-integration-lock-identity`

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (영향 파일 6건 전수 열거 — `spec.md` §4). Tier S의 `< 5 files` 기준 위반으로 M 채택.
- 산출물: `spec.md` / `plan.md` / `acceptance.md` / `progress.md`, `status: draft`.
- 요구 9 (REQ-BLV-001..009) / 수락 9 (AC-BLV-001..009). 추적표는 `spec.md` §3.1 — 모든 REQ가 ≥1개 AC를 가지며 **엄밀한 1:1은 아니다**(REQ-BLV-002 → AC-BLV-001 + AC-BLV-002). 감사 지적(표현 부정확)을 반영해 「1:1」 표현을 정정했다.
- SPEC ID 정규식 자가 점검: 실행됨, 출력 `PASS`.
- `moai spec lint --strict spec.md` → `✓ No findings`. **이 결과의 증명 범위는 좁다** (감사 N8 실측): `internal/spec/lint.go:748-773`은 12개 필드의 **비어 있지 않은 존재 여부만** 검사하고 enum은 검증하지 않는다. 따라서 lint 통과는 **필드 존재의 근거이지 enum 적합성의 근거가 아니다** — frontmatter 적합성은 스키마와 직접 대조해 확인했다(v0.4.0에서 `priority: HIGH`→`High` 교정).
- **전제 반증 1건**: 카드 지시문의 「Link 1 미커버」가 거짓임을 실측으로 확인하고 범위를 재구성했다(`spec.md` §1.4). 감사 시 이 절을 먼저 읽을 것.
- **선례 인용 정정 1건 (v0.2.0)**: v0.1.0이 `computeDeferredAdvisory`를 **발화** 선례로 인용했으나, 그 권고는 `HookOutput.Data`(`internal/hook/types.go:394`, `json:"-"`)로 들어가 도달하지 않는다. 선례를 「일정」과 「발화」로 분리하고 REQ/AC-BLV-008을 신설했다(`spec.md` §1.7.1). 이 정정을 놓친 구현은 결손을 그대로 재생산한다.
- **clarification 2건 모두 종결** (리드·운영자 결정): (a) 발화 표면 = `additionalContext` 단독, (b) `VERSION` 무변경 + 별도 `BUILD_ID`. `[NEEDS CLARIFICATION]` 잔여 0건.
- **잔여 위험 1건 명시**: `VERSION`이 그대로이므로 `moai version` 제목 줄은 `v3.1.2`로 남고 제목 줄만 읽는 오독 가능성이 남는다(`spec.md` §7.5). 제목 줄 축은 **별도 카드**로 이관 — 이 SPEC에서 넓히지 않는다.

### plan-audit iter-1 결과 (v0.3.0에서 반영 완료)

- 판정: **PASS-WITH-DEBT 0.80** (Tier M 임계 0.80 — 경계선). 보고서: `.moai/reports/t326/plan-audit-iter1.md`, 감사 트리 HEAD `25f7b0fe9`.
- must-pass 7건 전부 통과. 중심 분석(비교 존재 → 자동 호출 0건 → 운반 표면이 `json:"-"`)이 **독립 재측정으로 확인**됨.
- 결함 9건 중 **8건 수리 완료**(blocking 3 + optional 5, 리드가 optional까지 확대 지시). D9는 정보성으로 유지 — t317이 이 트리에 없다는 사실은 경로·스냅샷 일자와 함께 이미 공개돼 있다(`spec.md` §5 각주).
- **핀 판정**: AC-BLV-004의 `343399d2f` 핀은 감사관이 **독립적으로** 「올바른 규율이며 staleness가 아니다」로 판정(보고서 42-43행, 근거 `git merge-base --is-ancestor 343399d2f 25f7b0fe9` → true). 리드 판정과 일치하며 **핀은 그대로 유지**한다. 502로 갱신하지 않는다.
- 감사 미관측 항목(Gaps)은 보고서에 기록돼 있으며, 그중 「비-git 트리 실물 `moai doctor`」는 `plan.md` M4의 좁힌 형태 실측이 닫는다.

### plan-audit iter-2 결과 (v0.4.0에서 반영 완료)

- 판정: **PASS-WITH-DEBT 0.85** (Tier M 임계 0.80, 여유 있게 통과). iter-1 0.80 → iter-2 0.85 **단조 개선**. 보고서: `.moai/reports/t326/plan-audit-iter2.md`, 감사 트리 HEAD `5fc676bbe`.
- must-pass 7건 전부 통과. iter-1 결함 9건 전수 재검사 — 5건 완전 종결, D6 반쯤 종결(뮤턴트는 수리됨, 선례·테스트 경로가 남았고 N2·N3로 승계), D9는 정보성 유지.
- **REQ/AC-BLV-009가 백지 상태의 anti-vacuity 기준을 통과했다**: 감사관이 뮤턴트가 실제로 RED를 만드는지 독립 확인했고, 집합 형태가 「항목 추가 없음」보다 **엄격**하다는 점(rename·removal도 포착)을 추가로 지적했다.
- 신규 결함 8건(N1-N8) **전수 수리**. 상세는 `spec.md` HISTORY v0.4.0.
- **핀 판정 재확인**: AC-BLV-004의 `343399d2f` 핀은 iter-2가 **새 근거로** 유지 판정했다 — 그 SHA는 `moai doctor --check "Binary Freshness"`가 지금 보고하는 **설치된 바이너리의 빌드 커밋**(살아 있는 앵커)이다. 494 → 514 표류는 정상 커밋 누적. **핀은 그대로.**
- **SPEC의 논지가 감사 중 실물로 재현됐다**: 감사 시점에 `moai doctor --check "Binary Freshness"`가 `binary is behind source tree (binary: 343399d2f, HEAD: 5fc676bbe)`를 rc=0으로 출력했고, **아무도 그것을 요청하지 않았다.** 이 카드가 닫으려는 결손 그 자체다.
- Tier M 반복 상한(2/2) 도달 — iter-3 없음. 리드가 상한을 넘기지 않기로 결정했고, N1-N8은 모두 단일 셀 편집이며 그 근거 측정을 리드가 직접 확인했다.
- 감사 미관측 항목(Gaps)은 보고서에 기록. 「비-git 트리 실물 `moai doctor`」는 iter-1과 동일하게 여전히 미관측이며 `plan.md` M4가 소유한다.

_Implementation Kickoff Approval 대기_

## §E.2 Run-phase Evidence

측정 트리: 워크트리 `.claude/worktrees/t326`, 브랜치 `WT-integration-lock-identity`.
기준점(이 SPEC 첫 커밋의 부모) = **`22f90b1c7`** — AC-BLV-009의 전후 대조가 여기에 고정된다.
구현 커밋 = `c70c6aed9`.

증거 원본은 `.moai/state/verify/t326/`(gitignored)에 있고, 판정에 쓰인 출력은 아래에 직접 옮겨 적었다.

### 착수 전 기준선 — 이미 RED였다 (plan §C에 따라 리드에 보고)

```
$ go test ./internal/cli/... ./internal/hook/...
rc=1
--- FAIL: TestRunDoctor_WithExport (3.87s)
    coverage_improvement_test.go:715: runDoctor error: doctor: 1 check(s) failed
  (같은 형태 9건, 전부 internal/cli. 다른 패키지는 모두 ok)

$ go run ./cmd/moai doctor
fail  Agent Emit Embed   no readable binary to judge at .../worktrees/t326/bin/moai (11 committed artifacts to compare)
fail  Harness 5-Layer    L1:FAIL L2:FAIL L3:FAIL L4:FAIL L5:FAIL L6:FAIL
Pass 24  Warn 2  Fail 2
```

두 Fail 모두 이 SPEC 소관이 아니다. `Agent Emit Embed`는 t317이 착지시킨 검사이고, 이 워크트리에
`bin/moai`가 없어서 실패했다(이후 `make build`로 해소). `Harness 5-Layer`는 별개의 선행 실패다.
따라서 이 SPEC의 회귀 판정은 **절대 상태가 아니라 델타**로 한다 — 구현 후에도 같은 9건, 신규 0건.

### 구현 위치 — SPEC §4 행 2에서의 이탈 1건

§4 행 2는 seam의 자리를 `internal/cli/binary_lag.go`로 예시했다(「예:」). 그 위치는 **빌드되지
않는다**: `internal/cli/hook.go:26`이 `internal/hook`을 import하므로, 세션 시작 핸들러도 불러야
하는 seam을 `internal/cli`에 두면 import 사이클이다. seam은 두 호출자 아래의 새 패키지
`internal/binlag`에 두었다. REQ-BLV-005가 요구하는 것(구현 1개 + 대체 가능한 seam 1개)은 그대로이며,
오히려 AC-BLV-005가 판정 가능해진다 — 패키지마다 seam 변수를 두면 한쪽에 심은 스텁을 다른 쪽이
보지 못해 기준이 아무것도 구분하지 못한다. 파일 수와 Tier M 근거는 불변. 리드에 보고 완료.

### 적용가능성 기준점 해석 — §5 문구와의 차이 1건

§5는 t317 D9 선례를 따라 `.moai/` marker 상향 탐색을 채택하라고 적는다. 이 축에서는 그 탐색이
**해롭다**: `git rev-parse`가 이미 상위로 걸어 올라가므로 하위 디렉터리 문제는 그것으로 닫히는 반면,
`.moai/` 존재를 추가 조건으로 걸면 `.moai/`가 없는 정상 git 트리가 적용 불가로 뒤집혀 §5가 금지하는
관용 축소가 된다. AC-BLV-003의 하위 디렉터리 하위 케이스는 git의 상향 탐색으로 GREEN이다
(`TestEvaluate_SubdirectoryStaysApplicable`).

### AC별 판정 — GREEN과 대표 뮤턴트 RED

각 행: 기준 → GREEN 판정 명령 → 심은 뮤턴트 → 관측된 RED.

**AC-BLV-001** (요청 없는 발화)

```
GREEN $ go test ./internal/hook/ -run TestSessionStart_LagAdvisoryReachesAdditionalContext
--- PASS (0.06s)

뮤턴트: 판정은 계산하되 어떤 출력 필드에도 넣지 않는다
        appendAdditionalContext(out, binaryLagAdvisory(...)) → _ = binaryLagAdvisory(...)
RED   --- FAIL: TestSessionStart_LagAdvisoryReachesAdditionalContext (0.06s)
          session_start_binary_lag_test.go:66: additionalContext does not name "aaaaaaaaa":
          session_start_binary_lag_test.go:66: additionalContext does not name "fffffffff":
```

**AC-BLV-002** (대조군: 일치 시 침묵)

```
GREEN $ go test ./internal/hook/ -run TestSessionStart_NoLagAdvisoryWhenBinaryMatchesHead
--- PASS (0.07s)

뮤턴트: 무조건 방출 — Advisory()에서 `if v.Status != StatusBehind { return "" }` 제거
RED   --- FAIL: TestSessionStart_NoLagAdvisoryWhenBinaryMatchesHead (0.06s)
          session_start_binary_lag_test.go:86: lag advisory emitted for a binary that matches HEAD:
```

**AC-BLV-003** (적용 불가가 exit 상태를 바꾸지 않는다)

```
GREEN $ go test ./internal/cli/ -run TestBinaryLag_NonGitDirectoryKeepsDoctorExitZero
--- PASS (0.00s)
      $ go test ./internal/binlag/ -run 'TestEvaluate_NonGitDirectoryIsNotApplicable|TestEvaluate_SubdirectoryStaysApplicable'
--- PASS (하위 디렉터리 앵커 포함)

뮤턴트: 적용 불가를 Fail로 — doctor.go default 분기 CheckOK → CheckFail
RED   --- FAIL: TestBinaryLag_NonGitDirectoryKeepsDoctorExitZero (0.00s)
          binary_lag_test.go:90: non-git directory reported Fail: "development build (no commit metadata)"
```

실물 확인 (실행 바이너리, 적용 불가 분기, 좁힌 `--check` 형태):

```
$ GIT_DIR=/nonexistent-t326 ./bin/moai doctor --check "Binary Freshness"
rc=0
  ok   Binary Freshness   not in a git source tree (skipped)
  Pass 1  Warn 0  Fail 0
```

> **Gap (미관측)**: 저장소 **밖 cwd**에서의 실물 실행은 관측하지 못했다. 워크트리 격리 가드가 트리
> 밖으로의 cwd 변경(`cd`, `env -C`, `sh -c`)을 모두 거부한다 — plan-audit이 감사관에 대해 기록한
> 것과 같은 제약이다. 위 실행은 `GIT_DIR`을 깨뜨려 **같은 분기**(`git rev-parse` 실패 → 적용 불가)를
> 태운 것이며, 원인이 「비-git 디렉터리」가 아니라 「깨진 GIT_DIR」이라는 점에서 대체물이다. 분기와
> exit 상태는 관측됐고, cwd 형태는 관측되지 않았다.

**AC-BLV-004** (빌드 신원 단조성, `VERSION` 불변)

```
RED-now (사전 구현 트리, 트리 SHA 22f90b1c7):
      $ go test ./internal/cli/ -run TestBuildIdentity_IsMonotoneAcrossAnAncestorRelation
--- FAIL: TestBuildIdentity_IsMonotoneAcrossAnAncestorRelation (0.04s)
    binary_lag_test.go:274: the Makefile defines no BUILD_ID; there is no monotone build identity to derive
  RED인 이유: BUILD_ID 부재(기능 부재)이며 무관한 파일 때문이 아니다.

GREEN (M3 이후):
--- PASS: TestBuildIdentity_VersionDerivationUnchanged (0.04s)
--- PASS: TestBuildIdentity_IsMonotoneAcrossAnAncestorRelation (0.48s)

뮤턴트 1: VERSION 파생에서 --abbrev=0 제거해 단조성을 얻는다
RED   --- FAIL: TestBuildIdentity_VersionDerivationUnchanged (0.03s)
           got: VERSION ?= $(shell git describe --tags 2>/dev/null || ...)
          want: VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || ...)

뮤턴트 2: 시각을 신원으로 — BUILD_ID := $(shell git show -s --format=%cI HEAD ...)
RED   --- FAIL: TestBuildIdentity_IsMonotoneAcrossAnAncestorRelation (0.53s)
          binary_lag_test.go:318: BUILD_ID collapses two commits in an ancestor relation onto
                                  "2026-08-28T03:35:34+09:00"; a build identity that cannot
                                  separate them cannot report lag
          binary_lag_test.go:324: BUILD_ID "2026-08-28T03:35:34+09:00" carries no component
                                  identifying commit b55db7b
```

실물 확인 — `make build` rc=0, ldflags에 실린 값:

```
-X .../pkg/version.Version=v3.1.2 -X .../pkg/version.BuildID=v3.1.2-539-gc70c6aed9

$ ./bin/moai version
 v3.1.2   v3.1.2-539-gc70c6aed9   built 2026-08-27T18:38:32Z
```

제목 줄이 `v3.1.2`로 남는 것은 spec.md §7.5가 명시한 **감수 비용**이며 결함이 아니다. 스탬프 줄이
이제 두 빌드를 구분한다.

**AC-BLV-005** (단일 구현 seam)

```
GREEN $ go test ./internal/cli/ -run TestBinaryLag_OneSeamServesBothSurfaces
--- PASS (0.07s)   두 표면 모두 스텁 판정을 반영

뮤턴트: 훅 쪽에 조상 판정 복사본을 심는다(rev-parse + merge-base --is-ancestor 직접 호출)
RED   --- FAIL: TestBinaryLag_OneSeamServesBothSurfaces (0.09s)
          binary_lag_test.go:68: session-start advisory did not reflect the stub verdict;
                                 it is not reading the same seam
```

**AC-BLV-006** (실패 개방 + 시간 제한)

```
GREEN $ go test ./internal/hook/ -run TestSessionStart_BlockingComparerDoesNotStallSessionStart
--- PASS (0.32s)   context 취소를 무시하는 블록 스텁 아래에서 기한 안에 반환, err == nil, 권고 부재

뮤턴트: 기한 없이 조인 — timer+select 제거하고 `return <-result`
RED   panic: test timed out after 1m0s
      github.com/modu-ai/moai-adk/internal/hook.binaryLagAdvisory(...)   ← 매달린 프레임
```

감사 N3 전제는 실제로 물렸다: 이 테스트는 `deferredScansAsync`를 `true`로 뒤집고 `t.Cleanup`에서
복원한다(선례 `session_start_parallel_test.go:315-321`). 뒤집지 않으면 인라인 분기에는 기한 조인이
없어 구현이 옳든 그르든 매달린다.

**AC-BLV-007** (판정 근거는 커밋 조상이지 버전 문자열이 아니다)

```
GREEN $ go test ./internal/binlag/ -run TestEvaluate_VersionStringDoesNotDecideTheVerdict
--- PASS (0.51s)   낮은 semver를 보고하는 후손 = 뒤처짐 아님, 높은 semver를 보고하는 조상 = 뒤처짐

뮤턴트 1: semver 비교로 판정 — merge-base 분기를 `if req.BinaryVersion < "v3.1.3"`로 교체
RED   --- FAIL: TestEvaluate_VersionStringDoesNotDecideTheVerdict (0.41s)
          binlag_test.go:157: ancestor build with the higher semver: status = "divergent", want "behind"
```

뮤턴트 2(`strings | grep`로 기능 존재 추론)는 심지 않았다 — 구현이 바이너리 문자열을 **읽지 않으므로**
심을 자리가 없다. 그 형태의 부재는 소스에서 직접 확인된다: `internal/binlag/binlag.go`에 `strings`
패키지의 `Contains`/외부 `strings(1)` 호출이 없고, 판정은 `rev-parse` + `merge-base --is-ancestor`
두 호출뿐이다(위 뮤턴트 1이 그 두 호출이 실제 판정임을 반증으로 보였다).

**AC-BLV-008** (발화 표면이 관측자에게 도달한다)

```
RED-now: 사전 구현 트리에는 지연 권고가 없어 직렬화 출력에 그 문자열이 없다(기능 부재).

GREEN $ go test ./internal/hook/ -run TestSessionStart_LagAdvisorySerializesUnderAdditionalContext
--- PASS (0.06s)   json.Marshal → 재-unmarshal → hookSpecificOutput.additionalContext 키에서 판독

뮤턴트 1: 도달하지 않는 필드 — 권고를 HookOutput.Data(json:"-")에 넣는다
RED   --- FAIL (0.06s) lag advisory absent from the serialized hookSpecificOutput.additionalContext:
        {"systemMessage":"Factory Mode: ...","hookSpecificOutput":{"hookEventName":"SessionStart",
         "additionalContext":"moai session attribution: ..."}}
        ← 권고가 직렬화 문서에 아예 없다

뮤턴트 2: 도달하지만 잘못된 키 — 권고를 SystemMessage에 쓴다
RED   --- FAIL (0.07s) lag advisory absent from the serialized hookSpecificOutput.additionalContext:
        {"systemMessage":"moai binary lag: the installed binary was built from commit aaaaaaaaa,
          an ancestor of this tree's HEAD fffffffff.\n...","hookSpecificOutput":{...
          "additionalContext":"moai session attribution: ..."}}
        ← 권고가 문서에는 있다. 문서 전체 부분 문자열 검색이었다면 통과했을 것이며,
          키를 지정한 unmarshal만이 이 뮤턴트를 잡는다(감사 D7이 요구한 판별력).
```

**AC-BLV-009** (doctor 검사 이름이 하나도 늘지 않는다)

이 기준은 설계상 처음부터 GREEN인 **회귀 가드**다. 유효성의 유일한 근거는 아래 세 뮤턴트가 실제로
RED가 되는 것이다. 판정은 `go/ast`로 세 슬라이스 리터럴 **전체**를 읽고 각 항목의 **이름 표현식**을
집합 원소로 삼아, `git show 22f90b1c7:internal/cli/doctor.go`와 현재 파일을 대조한다.

```
GREEN $ go test ./internal/cli/ -run TestBinaryLag_DoctorCheckNameSetIsUnchanged
--- PASS (0.06s)   추가 0, 제거 0, "Binary Freshness" 양쪽 존재

뮤턴트 1: moaiChecks에 {"Binary Lag", checkBinaryFreshness} 추가
RED   binary_lag_test.go:177: this SPEC added doctor check name "Binary Lag"; ...

뮤턴트 2: 이웃 레지스트리로 회피 — 같은 항목을 workspaceChecks에 추가
RED   binary_lag_test.go:177: this SPEC added doctor check name "Binary Lag"; ...
      ← moaiChecks만 보는 판정이었다면 통과했을 것(감사 N7)

뮤턴트 3: 상수 식별자로 회피 — {binaryLagCheckName, checkBinaryFreshness}
RED   binary_lag_test.go:177: this SPEC added doctor check name binaryLagCheckName; ...
      ← 인용 부호가 없다. 따옴표 문자열만 추출했다면 델타가 비어 통과했을 것(감사 N1)
```

### 회귀 확인

```
$ go test ./internal/cli/... ./internal/hook/... ./internal/binlag/...
rc=1 — 실패 9건, 전부 착수 전 기준선과 동일한 `doctor: 1 check(s) failed` 형태.
       신규 실패 0건. 델타 = 0.

$ go vet ./internal/cli/... ./internal/hook/... ./internal/binlag/... ./pkg/version/...
(출력 없음 — 통과)

$ go build ./...
(출력 없음 — 통과)

$ make build
rc=0

$ ./bin/moai doctor --check "Binary Freshness"     # 이 저장소
rc=0   ok  Binary Freshness  binary matches source HEAD (c70c6aed9)
```

`moai doctor` **전체** 실행은 이 워크트리에서 exit 1이며, 그 원인은 위 기준선의 선행 Fail 2건이다.
이 SPEC이 일으키지도 통제하지도 않는 검사이므로, 완료 정의가 이미 좁혀 놓은 `--check "Binary Freshness"`
형태로 판정한다(감사 D8이 비-git 케이스에 대해 세운 것과 같은 근거가 소스 트리에서 실현된 것).

### 범위 준수

- t317 워크트리에 쓰기 0건 — 읽기 전용 인용만.
- 신규 doctor 검사 이름 0건(AC-BLV-009가 기계적으로 판정).
- `Makefile`의 `VERSION ?=` 행 바이트 동일(AC-BLV-004 (c)가 기계적으로 판정, 기준 SHA `22f90b1c7`).
- 통합 락·재설치 자동화·임베드 축·실행 중 프로세스 축·릴리스 버전 정책·`moai version` 제목 줄 —
  전부 손대지 않음.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-28
run_commit_sha: c70c6aed9        # 구현 커밋. 증거 기록 커밋이 뒤따른다
run_status: complete
ac_pass_count: 9
ac_fail_count: 0
mutants_planted_red_reverted: 13
baseline_sha: 22f90b1c7          # AC-BLV-009 전후 대조의 고정 기준점
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run    # 이 카드는 push하지 않는다(통합은 오케스트레이터 소관)
l44_post_push_fetch: not-run
new_warnings_or_lints_introduced: 0
cross_platform_build:
  status: partial
  observed: go build ./... (darwin/arm64) 통과, go vet 4개 패키지 통과
  not_observed: GOOS=windows / GOOS=linux 교차 빌드 — CI 몫
total_run_phase_files: 8         # binlag 2 + cli 3 + hook 2 + pkg/version 1 (+ Makefile)
m1_to_mN_commit_strategy: 2-commit  # 구현 1건 + 증거 1건
pre_existing_baseline_red: true  # 착수 전부터 internal/cli 9건 FAIL — 이 SPEC 소관 아님
pre_existing_doctor_fails: ["Agent Emit Embed", "Harness 5-Layer"]
regression_delta: 0
spec_deviations:
  - seam 위치: SPEC §4 행 2의 예시(`internal/cli/binary_lag.go`) → `internal/binlag`
    사유: internal/cli → internal/hook import 사이클. 리드 보고 완료
  - 적용가능성 기준점: §5의 `.moai/` marker 상향 탐색 대신 git 자체의 상향 탐색
    사유: marker 조건 추가는 §5가 금지하는 관용 축소가 된다
unobserved:
  - 저장소 밖 cwd에서의 실물 doctor 실행(워크트리 격리 가드가 cwd 변경을 거부).
    같은 분기를 GIT_DIR 파괴로 태워 exit 0 관측함
  - windows/linux 교차 빌드 및 전체 스위트 — CI 몫
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
