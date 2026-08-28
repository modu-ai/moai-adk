---
id: SPEC-BINARY-LAG-VISIBILITY-001
title: "배포 지연 가시성 — 진행 기록"
version: "0.4.0"
status: completed
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

### AC-BLV-009 — t317 착지가 두 설계 판단을 실제로 입증했다

이 SPEC v0.4.0이 `moaiChecks`를 인용한 시점 이후, plan-phase 중에 t317이 자기 항목을
**같은 슬라이스에, 상수 식별자로** 등록하며 착지했다. 실측:

```
$ grep -n 'moaiChecks := \|"Binary Freshness"\|mcpServerVersionCheckName\|agentEmitEmbedCheckName\|workspaceChecks := ' internal/cli/doctor.go
197:	moaiChecks := []checkFunc{
201:		{"Binary Freshness", checkBinaryFreshness},
203:		{mcpServerVersionCheckName, func(v bool) DiagnosticCheck { return checkMCPServerVersion(cwd, v) }},
207:		{agentEmitEmbedCheckName, func(v bool) DiagnosticCheck { return checkAgentEmitEmbed(cwd, v) }},
220:	workspaceChecks := []checkFunc{
```

SPEC이 인용한 창(`moaiChecks` `:195`~`:212`, `workspaceChecks` `:214`)과 어긋난다. 두 원인이
겹쳐 있고, 둘 다 이 기준의 판정과 무관하다:

- **t317의 항목 4줄** (주석 3 + 항목 1). 기준점 blob에서 이미 확인된다 — `git show
  22f90b1c7:internal/cli/doctor.go`에서 `workspaceChecks`가 `:218`(SPEC 인용 `:214` 대비 +4).
- **이 SPEC의 import 2줄** (`context`, `binlag`). `:218` → `:220`의 나머지 +2가 그것이다.

**기준점 blob에 t317 항목이 이미 들어 있다**는 것이 결정적이다:

```
$ git show 22f90b1c7:internal/cli/doctor.go | grep -c 'agentEmitEmbedCheckName'
1
```

즉 `agentEmitEmbedCheckName`은 before 집합과 after 집합 **양쪽에** 있고 델타에 0을 기여한다.
AC-BLV-009가 통과하는 것은 이 SPEC이 이름을 추가하지 않았기 때문이지, 슬라이스가 그대로여서가
아니다 — 슬라이스는 그대로가 아니다.

두 설계 판단이 가설에서 **입증**으로 바뀌었다:

- **절대 개수·내용이 아니라 전후 델타로 판정한 것**(감사 D3). 슬라이스의 절대 상태를 단언했다면
  t317이 착지하는 순간, 이 SPEC과 무관한 이유로 깨졌을 것이다.
- **추출 단위를 「항목의 이름 표현식」으로 잡은 것**(감사 N1). 상수 식별자 등록은 이제
  `mcpServerVersionCheckName` 하나가 아니라 **둘**이다 — 뮤턴트 3이 방어하는 형태의
  살아 있는 실례가 같은 슬라이스 안에 있다.

행 인용이 어긋난 것은 드리프트가 아니라 **이 기준이 제 일을 하고 있다는 증거**다. 판정이 행
구간이 아니라 슬라이스 리터럴 전체를 `go/ast`로 읽고 델타를 보기 때문에, 창이 밀려도 결과가
바뀌지 않는다.

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

### 실물 end-to-end 관측 — 스텁 없이, 실제 지연 상태에서

증거 커밋(`1c042b93a`)이 착지하면서 트리 HEAD가 설치된 바이너리의 빌드 커밋(`c70c6aed9`)을
앞질렀다. 즉 **이 SPEC이 닫으려는 바로 그 상태가 실제로 발생했고**, 그 상태에서 실물 바이너리로
두 표면을 모두 관측했다. 스텁도, 주입도 없다.

doctor 표면:

```
$ ./bin/moai doctor --check "Binary Freshness"
rc=0
  warn  Binary Freshness  binary is behind source tree (binary: c70c6aed9, HEAD: 1c042b93a)
  Pass 0  Warn 1  Fail 0
```

요청 없는 발화 표면 — 실제 훅 실행, 직렬화된 출력의 `hookSpecificOutput.additionalContext`:

```
$ ./bin/moai hook session-start < <session-start JSON>
rc=0
$ jq -r '.hookSpecificOutput.additionalContext' <출력>
...
Factory Mode: joined the factory run as lane-18.

moai binary lag: the installed binary was built from commit c70c6aed9, an ancestor of this tree's HEAD 1c042b93a.
Fixes committed after c70c6aed9 are NOT in the binary you are running, so its output describes older code.
Rebuild before trusting any moai CLI result: make build && make install
```

세 가지가 한 번에 확인된다: (1) 아무도 진단 명령을 치지 않았는데 판정이 도달했다,
(2) 두 SHA와 수리 명령이 모두 담겼다, (3) 기존 Factory Mode 공지를 **덮어쓰지 않고 덧붙였다**
(append-if-non-empty 패턴 보존).

exit 상태는 doctor에서 0으로 유지된다 — 지연은 `Warn`이지 `Fail`이 아니므로
`doctorExitStatus`를 승격시키지 않는다.

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

```yaml
sync_complete_at: 2026-08-28
sync_commit_sha: d3454f1e6   # 이 커밋의 바로 다음 커밋에서 backfill(커밋은 자기 해시를 알 수 없다)
sync_status: COMPLETE — CHANGELOG 항목 1건 + @MX 주석 + 3-phase close. 통합은 리드 소관(push 안 함, PR 안 냄)
sync_audit_verdict: "PASS-WITH-DEBT 0.86 (가중 조화평균) — 보고서 .moai/reports/t326/sync-audit.md. 차원별 Functionality 0.95 / Security 1.00 / Craft 0.65 / Consistency 0.80; must-pass 방화벽(Functionality·Security) 통과. 지적 6건 F1-F6 중 F1·F2 가 blocking"
sync_audit_score_baseline: "[HARD] 0.86 은 **수리 이전 트리 bfd09fcfc** 에 대고 잰 값이다. F1(레이스) 과 F2(D8 누락) 를 수리한 뒤 트리(a213f161b)에 대한 재측정은 **하지 않았다** — 점수가 올랐을 것이라고 추정할 수는 있으나 아무도 재지 않았으므로 개선된 점수를 주장하지 않는다"
sync_audit_findings_disposition:
  F1: "[HIGH][blocking] `go test -race ./internal/hook/` 레드(DATA RACE 1건) — **수리됨**. 커밋 a213f161b, 테스트 전용. 재측정 rc 0(verification.race_hook)"
  F2: "[MEDIUM][blocking] 스캔이 자기 선언 범위 안의 어긋남을 놓쳤고 그 측정을 확인으로 기록 — **수리됨**. D8 추가 + D7 note 재작성 + divergences_found 갱신(이 커밋)"
  F3: "[MEDIUM][optional] 스캔 method 가 plan.md 를 제외 — **흡수함**. method 를 plan.md 로 넓히고 D9-D13 추가, 감사가 지목한 5건 전부 이 트리에서 직접 재측정(이 커밋)"
  F4: "[LOW][optional] D2 처분은 옳으나 논거가 한 걸음 넓다 — 조치 없음. 감사 자신이 '부채이지 결함 아님'으로 분류했고 D1..D13 일괄 manager-spec 재위임을 대안으로 제시한다. 리드 소관"
  F5: "[LOW][optional] `go run ./cmd/moai doctor` 의 fail 2건과 테스트의 `doctor: 1 check(s) failed` 가 서로 다른 환경 측정인데 한 문단에 있다 — 조치 없음, 리드 판단 대기"
  F6: "[INFO] residual_risk 3건은 정확 — 조치 없음"
lane_verdict_report: .moai/reports/t326/verdict.md   # run-phase 레인 판정(카드 전제 반증 기록 포함)
plan_audit_reports: [.moai/reports/t326/plan-audit-iter1.md, .moai/reports/t326/plan-audit-iter2.md]  # PASS-WITH-DEBT 0.80 → 0.85 단조

b12_self_test_a: "사전 grep — `grep -c 'SPEC-BINARY-LAG-VISIBILITY-001' CHANGELOG.md` 가 쓰기 전 0 반환(병렬 세션의 중복 항목 없음)"
b12_self_test_b: "AC 개수 일치 — `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l` 가 9 반환(AC-BLV-001..009); CHANGELOG 항목도 9건으로 적었다"
b12_self_test_c: "파일 경로 확인 — CHANGELOG 항목이 인용한 경로 전부 `ls -1` 로 존재 확인(internal/binlag/binlag.go, internal/cli/doctor.go, internal/cli/version.go, internal/hook/session_start.go, internal/hook/session_start_binary_lag.go, pkg/version/version.go, Makefile, spec.md). 인용한 행번호 `session_start.go:479` 도 grep로 재확인 — 초안의 `:461` 은 틀렸고 고쳤다"
changelog_entry_position: "CHANGELOG.md [Unreleased] → ### Added 첫 항목(12행)"

frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed (3-phase close, 이 단일 sync 커밋에 병합); `updated:` 2026-08-28"
  plan_md: "in-progress → implemented → completed; `updated:` 2026-08-28"
  acceptance_md: "in-progress → implemented → completed; `updated:` 2026-08-28"
  progress_md: "in-progress → implemented → completed; `updated:` 2026-08-28"

mx_annotations_added:
  - "internal/binlag/binlag.go — Comparer에 @MX:ANCHOR(단일 구현 seam 불변식, SPEC/REASON/TEST 하위줄), Evaluate에 @MX:ANCHOR(fan_in 3: doctor.go:520 + session_start_binary_lag.go:55,:67), gitCompare에 @MX:WARN(git 외부 호출 + 다섯 관용 경로 축소 금지, REASON 하위줄), Advisory에 @MX:NOTE"
  - "internal/cli/doctor.go — checkBinaryFreshness에 @MX:ANCHOR(행 이름 고정 + Fail 승격 금지, REASON/TEST 하위줄). 파일 상단의 기존 @MX:NOTE는 그대로 둠"
  - "internal/hook/session_start_binary_lag.go — binaryLagJoinBound(250ms 상수)에 @MX:NOTE. binaryLagAdvisory의 기존 @MX:WARN은 run-phase가 이미 달았고 그대로 둠"
  - "pkg/version/version.go — GetBuildID에 @MX:NOTE(Version으로 폴백하지 않는다)"
mx_fan_in_basis: "grep -rn 'binlag\\.' internal pkg --include='*.go' | grep -v _test.go — Evaluate 3개 호출점 관측"

docs_review:
  readme: "변경 없음 — 4개 로케일 README 어디에도 doctor 검사 행 목록이나 세션 시작 훅 내부가 없다. 이 기능이 바꾸는 것은 유지자가 보는 표면이고, README가 서술하는 사용자 여정에는 걸리지 않는다"
  docs_site: "변경 없음 — `docs-site/content/ko/cli-reference/doctor.md` 를 `Binary\\|Freshness\\|바이너리` 로 훑었고 검사 행 목록을 열거하지 않는다(유일한 히트는 무관한 「릴리즈 개수」 행). AC-BLV-009 가 doctor 검사 이름이 하나도 늘지 않았음을 기계적으로 판정하므로 문서에 추가할 행 자체가 없다. `BUILD_ID`/`BuildID` 전수 grep도 docs-site·README 히트 0"
  verdict: "문서 수정 불필요 — 편집하지 않았다. 판단을 명시로 남긴다"

template_first_check:
  claim: "이 SPEC은 템플릿 미러 의무를 남기지 않았다"
  evidence: "run-phase 전체 커밋(c70c6aed9 + 증거 4건)이 건드린 파일 10개는 모두 Go 소스·Makefile·SPEC 산출물이다. `internal/template/templates/` 하위에 Makefile도 pkg/version 미러도 존재하지 않는다(find 결과 0건). sync-phase가 건드린 CHANGELOG.md 역시 템플릿 대응물이 없다"
  make_build: "실행하지 않음 — 임베드 대상 파일을 아무것도 바꾸지 않았으므로 재생성할 것이 없다"

verification:
  vet: "`go vet ./internal/binlag/... ./internal/cli/... ./internal/hook/... ./pkg/version/...` → rc 0 (증거: .moai/state/verify/t326-sync/vet.txt)"
  gofmt: "`gofmt -l` on binlag + doctor.go + session_start_binary_lag.go + pkg/version → 빈 목록(clean)"
  build: "`go build ./internal/binlag/... ./internal/cli/... ./internal/hook/... ./pkg/version/...` → rc 0"
  test_3pkg: "`go test ./internal/binlag/... ./pkg/version/... ./internal/hook/...` → rc 0, 전 패키지 ok (증거: .moai/state/verify/t326-sync/test-3pkg.txt)"
  test_cli: "`go test ./internal/cli/...` → rc 1, FAIL 9건 (증거: .moai/state/verify/t326-sync/test-cli.txt)"
  test_cli_attribution: "FAIL 9건은 §E.3이 기록한 `pre_existing_baseline_red: true` 의 9건과 개수가 일치하고, 전부 doctor 스위트(TestRunDoctor_* 6 + TestDoctorCmd_* 3)다. `go run ./cmd/moai doctor` 로 실제 실패 검사를 직접 판독했다 — `Agent Emit Embed`, `Harness 5-Layer` 2건으로 §E.3의 `pre_existing_doctor_fails` 와 같고, `Binary Freshness` 는 `ok — development build (no commit metadata)` 로 통과한다. regression_delta 0"
  race_hook: "`go test -race -count=1 ./internal/hook/` → **rc 0**, `ok github.com/modu-ai/moai-adk/internal/hook 33.223s` (증거: .moai/reports/t326/f1-evidence/manager-docs-recheck-race-hook.txt — a213f161b 트리에서 이 레코드 작성자가 직접 재실행). 리드 레인의 독립 재측정도 rc 0 (.moai/reports/t326/f1-evidence/lane4-recheck-race-hook.txt, 31.521s)"
  race_hook_gap_closed: "이 게이트는 run-phase 에도 1차 sync 에도 **실행되지 않았다** — CLAUDE.local.md §6 이 goroutine·channel 을 건드리는 코드에 `go test -race` 를 규정하고 이 SPEC 은 goroutine 을 추가하는데도 빠뜨렸다. sync-audit F1 [HIGH][blocking] 이 그 누락을 잡아냈고, 당시 이 게이트는 **레드**였다(DATA RACE 1건, 이 SPEC 자신의 신규 테스트 `TestSessionStart_BlockingComparerDoesNotStallSessionStart` 에서 stub goroutine 을 join 하지 않고 `binlag.Comparer` 를 복원 — 증거 .moai/reports/t326/f1-evidence/red-race-hook.txt). 커밋 a213f161b(테스트 전용 수정)로 초록이 됐다. 메워진 결손이며, 메워졌다는 사실과 함께 결손이었다는 사실을 남긴다"
  ac_blv_004c_recheck: "`git show 22f90b1c7:Makefile | grep '^VERSION ?='` 와 현재 트리의 같은 줄을 diff — 동일. sync-phase 편집이 이 불변식을 건드리지 않았음을 재확인"

sync_phase_files: "CHANGELOG.md(항목 1건), .moai/specs/SPEC-BINARY-LAG-VISIBILITY-001/{spec.md·plan.md·acceptance.md·progress.md frontmatter, progress.md §E.4}, internal/binlag/binlag.go + internal/cli/doctor.go + internal/hook/session_start_binary_lag.go + pkg/version/version.go(@MX 주석만 — 실행 코드 무변경)"
push_state: "push 안 함, PR 안 냄 — 통합은 리드 소관(dispatch 지시)"
spec_body_untouched: "spec.md / plan.md / acceptance.md 본문 0줄 변경 — frontmatter `status:` + `updated:` 만"

spec_body_vs_code_divergence_scan:
  mandate: "리드 addendum(2026-08-28) — 이 카드의 전제는 plan·run에서 두 번 반증됐고 v0.4.1이 §5-vs-REQ-BLV-003 모순을 기록했다. 따라서 sync에서 남은 본문 주장과 착지한 코드의 어긋남을 능동적으로 훑는다. 발견은 **Gap으로 기록만** 한다 — 본문 편집은 manager-docs 소관 밖이고, 검증되지 않은 화해는 기록된 Gap보다 나쁘다"
  method: "spec.md §3(REQ 9건)·§4(파일 표 + [HARD] 구속 조항)·§5(적용가능성)·§6(C-1..C-3), acceptance.md AC-BLV-001..009, **plan.md 전체(§실측 근거·[HARD] 구속 조항·인용 목록 144-152행)** 의 코드 주장·경로·행번호를 grep/sed로 직접 재측정. 인용 행번호는 SPEC이 자기 기준점으로 지목한 22f90b1c7 와 v0.4.1 커밋 bf1a19813 양쪽에서도 재측정해 드리프트 귀속을 갈랐다. **plan.md 범위 확장(D9-D13)과 D8 은 sync-audit F2/F3 지적을 받아 a213f161b 트리에서 재측정해 추가했다** — 1차 스캔(D1-D7)은 f9c96c381 트리 기준이었고 plan.md 를 method 에서 제외했었다. 재측정 수치는 감사 보고서에서 옮겨온 것이 아니라 이 트리에서 직접 잰 값이다"
  divergences_found: 13   # 1차 7건 + F2 수리 1건(D8) + F3 범위 확장 5건(D9-D13)
  D1:
    file: ".moai/specs/SPEC-BINARY-LAG-VISIBILITY-001/spec.md §4 표 행 2"
    claim: "「`internal/cli/`의 신규 파일 1건 (예: `binary_lag.go`)」 — 단일 비교 구현이 internal/cli 에 놓인다"
    code: "구현은 신규 패키지 `internal/binlag/binlag.go`. internal/cli → internal/hook import 사이클 때문에 internal/cli 의 seam 은 훅 핸들러에서 도달 불가"
    note: "§E.3 `spec_deviations` 가 사유를 기록했고 §5 본문은 이미 `internal/binlag/binlag.go` 를 인용한다 — 즉 §4와 §5가 **같은 문서 안에서 서로 다른 경로를 주장한다**. 본문 미수리 상태이며 화해하지 않았다"
    attribution: "authoring/run-phase. 이 sync 커밋과 무관"
  D2:
    file: "spec.md §5"
    claim: "`internal/binlag/binlag.go:101`(rev-parse HEAD) · `:111`(merge-base --is-ancestor)"
    code: "현재 트리에서 각각 **121행 · 131행**"
    note: "bf1a19813 에서는 101·111 로 **정확했다**. 이 sync 커밋 `d3454f1e6` 이 그 위쪽에 @MX 주석 20줄을 넣어 밀어냈다 — **내가 만든 어긋남**이다. 인용이 가리키는 코드 자체는 동일하며 의미 변화는 없다"
    attribution: "이 sync 커밋 d3454f1e6"
  D3:
    file: "spec.md §4 [HARD] 구속 조항"
    claim: "`session_start.go:266`(maps.Copy) · `:277`(marshal) · `:301`(HookOutput{Data}) · `:574`(2차 병합)"
    code: "`internal/hook/session_start.go` 실측 — maps.Copy(data, advisory) 는 **258행과 276행 2곳**, `json.Marshal(data)` 는 **287행**, `out := &HookOutput{Data: jsonData}` 는 **311행**. 574행 부근에는 해당 구조가 없다"
    note: "SPEC이 스스로 지목한 기준점 `22f90b1c7` 에서도 258/276/287/311 로 **동일** — 즉 병렬 레인 드리프트가 아니라 **작성 시점의 인용 오차**다. 조항이 말하는 사실(Data 가 json:\"-\" 라 그 경로는 직렬화되지 않는다)은 참이고 `types.go:394` 인용도 정확하다. 틀린 것은 행번호뿐"
    attribution: "authoring. 코드 변경과 무관"
  D4:
    file: "spec.md §4 표 행 3"
    claim: "「`AdditionalContext` append 지점(`:343-346`·`:369` 패턴)에 덧붙인다」"
    code: "실제 append 블록은 353-356 · 379-382 · 430-433 · 454-457 이고, 착지한 권고는 그중 어느 곳도 아닌 **479행**에서 신규 헬퍼 `appendAdditionalContext(out, …)` 로 붙는다(Handle 말미)"
    note: "요구(REQ-BLV-008: additionalContext 에 쓴다)는 충족한다. 어긋난 것은 인용한 위치와 「덧붙이는 자리」의 서술이다"
    attribution: "authoring"
  D5:
    file: "spec.md §5"
    claim: "`internal/cli/doctor.go:140` 이 `doctorExitStatus`"
    code: "`func doctorExitStatus` 는 **142행**(140 은 그 doc 주석 안). bf1a19813 에서도 142"
    note: "2행 오차. 의미 변화 없음"
    attribution: "authoring"
  D6:
    file: "acceptance.md AC-BLV-006"
    claim: "배제 대상 ctx-wrap 형태의 선례로 `session_start.go:622-624` 를 지목"
    code: "`computeDeferredAdvisory` 안의 `context.WithTimeout(context.Background(), driftTimeout)` 는 **652행**(기준점 22f90b1c7 에서는 632행). 622-624 는 함수 선언·doc 주석 자리"
    note: "이 인용 자체가 v0.4.0 감사 N2 의 **정정 결과**였는데, 정정된 번호도 여전히 어긋난다. 다만 AC 가 채택한 선례(`:243-257` timer+select)는 **정확하다** — 253행 `joinTimer := time.NewTimer(deferredScanJoinBound)` + 254행 select 가 인용 창 안에 있다. 판정력의 원천은 무사하고, 대비용으로 든 반례의 주소만 틀렸다"
    attribution: "authoring"
  D7:
    file: "acceptance.md AC-BLV-009"
    claim: "「`:245-249` 가 세 슬라이스를 `checkGroup` 으로 묶는다」"
    code: "실제 묶는 자리는 **254-258**(`return []checkGroup{` 254, System/MoAI-ADK/Workspace 3행 255-257). 245-249 는 무관한 observer 루프. 기준점 22f90b1c7 에서는 252행"
    note: "같은 AC 의 다른 인용 `:93-95`(allChecks 평탄화)는 **정확**하다. 세 레지스트리 선언은 실재하지만 그 실측값(189 systemChecks / 197 moaiChecks / 220 workspaceChecks)은 **같은 AC 가 인용한 187/195/214 와 어긋난다** — 1차 스캔은 이 수치를 재고 나서 '실재한다'는 존재 확인으로만 적어 정확성을 함의하게 뒀는데, 존재는 참이고 그 함의는 거짓이다. 어긋남 자체는 D8 로 분리해 기록한다. 판정 내용은 성립하며 근거 인용의 주소만 틀렸다"
    attribution: "authoring"
  D8:
    file: "acceptance.md AC-BLV-009 「판정 근거 위치」 (178행) · plan.md [HARD] 구속 조항 (86행) · plan.md 인용 목록 (144행)"
    claim: "세 레지스트리를 `systemChecks`(`:187`) · `moaiChecks`(`:195` 선언 ~ `:212` 닫는 괄호, 항목 `:196-211`) · `workspaceChecks`(`:214`) 로 지목. `Binary Freshness` 등록은 `:199`"
    code: "a213f161b 트리 실측 — `grep -n 'systemChecks\\|moaiChecks\\|workspaceChecks' internal/cli/doctor.go` 가 **189 / 197 / 220**. `moaiChecks` 는 197 선언, 항목 **198-217**, 닫는 괄호 **218**. `{\"Binary Freshness\", checkBinaryFreshness}` 는 **201행**"
    note: "선언 주소가 각각 2 / 2 / 6 어긋나고, 인용한 항목 구간 `:196-211` 은 실제 항목 구간 198-217 의 뒤쪽 4건(`Harness 5-Layer` 213 · `Migration` 214 · `Plugin Deployment` 215 · `Home Disk Usage` 217)을 창 밖에 남긴다 — AC-BLV-009 자신의 [HARD] 주석이 경고한 바로 그 실패 형태(「행 구간으로 추출하면 이 기준이 잡으려는 바로 그 뮤턴트가 창 밖에 놓인다」)다. **판정력은 무사하다**: 테스트는 슬라이스 리터럴을 행이 아니라 식별자로 읽으므로(`binary_lag_test.go` `checkNamesFromSource`) 이 주소들이 틀려도 AC 판정은 영향받지 않는다. 결함은 가드가 아니라 기록에 있다. 1차 스캔은 이 수치를 D7 note 에서 이미 쟀으면서 어긋남으로 적지 않았다(F2)"
    attribution: "authoring(인용) + 1차 sync 스캔의 누락(F2 로 지적받아 이번에 수리)"
  D9:
    file: "plan.md 25행 · 144행"
    claim: "`internal/cli/doctor.go:495` 가 `checkBinaryFreshness`"
    code: "`grep -n 'func checkBinaryFreshness' internal/cli/doctor.go` → **518행**"
    note: "23행 어긋남. @MX 주석 추가분(이 sync 커밋)이 일부 기여하나 재측정은 a213f161b 트리 기준값만 기록한다. 함수 자체는 실재하고 의미 변화 없음"
    attribution: "authoring + 이 sync 커밋의 주석 삽입 혼재(분리 미측정 — Gap)"
  D10:
    file: "plan.md 33행 · 148행"
    claim: "`internal/hook/session_start.go:593` 이 `computeDeferredAdvisory`"
    code: "`grep -n 'computeDeferredAdvisory' internal/hook/session_start.go` → 선언은 **623행**(`func (h *sessionStartHandler) computeDeferredAdvisory(`), 호출은 274·604행"
    note: "30행 어긋남. 함수는 실재하며 plan 이 이 함수에서 끌어낸 「일정 선례」 논지는 성립한다"
    attribution: "authoring"
  D11:
    file: "plan.md 150행"
    claim: "`internal/hook/session_start.go:305` 이 `AdditionalContext`"
    code: "305행은 doc 주석 안(`// The AdditionalContext is the serialized field per Claude Code SessionStart` 는 304행). 실제 필드 대입은 **315행**(`AdditionalContext: fmt.Sprintf(`)"
    note: "10행 어긋남. 같은 행이 함께 인용한 append 패턴 `:343-346`·`:369` 는 D4 와 동일한 어긋남(실측 353-356·379-382)이다"
    attribution: "authoring"
  D12:
    file: "plan.md 93행 · 152행"
    claim: "`Makefile:14` 가 `RELEASE_BINARY` 에 보간"
    code: "`grep -n 'RELEASE_BINARY' Makefile` → 정의는 **25행**(`RELEASE_BINARY := moai-$(VERSION)-$(PLATFORM)`), 소비는 64·65·66·68행. 14행에는 해당 정의가 없다"
    note: "11행 어긋남. VERSION 이 바깥으로 뻗는다는 논지(무변경 유지 근거)는 성립한다"
    attribution: "authoring"
  D13:
    file: "plan.md 93행 · 152행"
    claim: "`Makefile:35` 가 `version.json` 에 쓴다"
    code: "`grep -n 'version.json' Makefile` → 쓰기는 **66행**(`@echo '{\"version\":...}' > $(LOCAL_RELEASE_DIR)/version.json`); 17행은 주석 안 언급"
    note: "31행 어긋남. 같은 문장이 인용한 `internal/update/local.go:65` 는 **정확하다** — 65행이 `versionFile := filepath.Join(c.config.ReleasesDir, \"version.json\")`"
    attribution: "authoring"
  plan_md_scope_note: "F3 지적(1차 스캔의 method 가 plan.md 를 제외했다)을 받아 plan.md 를 훑었다. 감사가 지목한 5건은 **전부 이 트리에서 재현됐고**(D9-D13) 감사 보고서의 수치와도 일치하나, 위 값은 옮겨 적은 것이 아니라 a213f161b 에서 직접 잰 것이다. plan.md 의 나머지 인용은 이미 기록된 어긋남의 재진술이거나 정확하다 — `:266`·`:574`·`:277`·`:301`(D3 재진술) · `:245-249`(D7 재진술) · `:622-624`(D6 재진술) · `doctor.go:140`(D5 재진술) · `:343-346`·`:369`(D4 재진술) · `:187`/`:195`/`:212`/`:214`/`:199`(D8 재진술) · `session_start.go:243-257` timer+select 선례(**정확** — 253행 `joinTimer := time.NewTimer(...)` + 254행 select) · `doctor.go:93-95` allChecks(**정확** — 93행 `var allChecks []DiagnosticCheck`) · `Makefile:6` `VERSION ?=`(**정확** — AC-BLV-004(c) 가 기계 판정) · `internal/update/local.go:65`(**정확**). 본문은 한 줄도 고치지 않았다 — plan.md 는 closed 이며 본문 편집은 manager-docs 소관 밖이다"
  claims_re_verified_as_accurate:
    - "REQ-BLV-009 / C-2 — 검사 이름 집합 불변: 세 레지스트리 전수 판독에서 `Binary Freshness` 는 있고 `Binary Lag` 류 신규 이름은 없다"
    - "REQ-BLV-008 / plan.md M1 운영자 결정 — 권고는 additionalContext 단독. `session_start_binary_lag.go` 에 SystemMessage 쓰기 0건이고, session_start.go 의 SystemMessage 쓰기 2곳(386·437)은 무관한 기능의 운영자 통지"
    - "AC-BLV-008 인용 `types.go:394`(Data json:\"-\") · `types.go:366`(SystemMessage json 태그) 정확"
    - "AC-BLV-006 인용 `main_test.go:47`(deferredScansAsync=false) · `session_start_parallel_test.go:315-321`(origAsync 보존 → t.Cleanup 복원) 정확"
    - "REQ-BLV-004 — Makefile `VERSION ?=` 행이 22f90b1c7 대비 바이트 동일"
    - "§4 표 행 1·4·5·6 — doctor.go 환원, Makefile BUILD_ID, internal/cli/binary_lag_test.go, internal/hook 테스트 1건: 전부 실재"
  disposition: "13건 전부 **기록만** 했다. spec.md / plan.md / acceptance.md 본문은 한 줄도 고치지 않았다. D1 은 실질(경로가 다르다), D2-D13 은 인용 주소 오차이며 어느 것도 착지한 코드의 동작을 바꾸지 않는다. 본문 수리가 필요하다고 판단되면 manager-spec 재위임 소관이다 — 감사 권고 4번(D1..D13 을 한 배치로 묶어 재위임)이 그 형태를 제안한다"

gaps_explicitly_not_observed:
  - "**수리 후 트리의 감사 점수** — sync-audit 은 실시됐으나(PASS-WITH-DEBT 0.86) 그 측정은 bfd09fcfc 기준이다. F1·F2 수리 후 트리 a213f161b 에 대한 재감사는 배차되지 않았고 점수를 재지 않았다"
  - "**D9 의 귀속 분리** — `doctor.go:495` → 518 의 23행 어긋남에서 authoring 오차와 이 sync 커밋의 @MX 주석 삽입분이 각각 몇 행인지 갈라 재지 않았다. 22f90b1c7·bf1a19813 대비 재측정을 D2 처럼 수행하지 않았다"
  - "**감사가 남긴 Gap 은 이 레코드가 닫지 않았다** — 교차 플랫폼, `internal/cli` 의 -race, 기준선 22f90b1c7 에서 9건 FAIL 재현 여부, `make build` 실물 바이너리, `moai mx query` 수확. 보고서 §5 가 정본이며 이 sync 는 그중 어느 것도 새로 관측하지 않았다"
  - "교차 플랫폼 — 전부 darwin/arm64 관측. GOOS=windows / linux 빌드·테스트 미실행(CI 몫)"
  - "전체 스위트 미실행 — 레인-로컬 규율에 따라 4개 패키지로 범위를 좁혔다. 저장소 전체 판정은 origin/develop CI 몫이며 보고 시점에 PENDING"
  - "@MX 주석은 주석이므로 런타임 동작을 바꾸지 않는다는 것을 build/vet/test 재실행으로 확인했을 뿐, 별도의 MX 스캐너(`moai mx query`)로 태그가 실제 수확되는지는 관측하지 않았다"
  - "CHANGELOG 산문의 서술 정확성은 코드·SPEC 재판독으로 세웠지만, 문장 단위로 판정하는 기계적 게이트는 없다"

residual_risk:
  - "`Binary Freshness` 행은 doctor 스위트 9건이 이미 레드인 패키지 안에 있다. 이 SPEC이 그 레드를 늘리지도 줄이지도 않았지만, 그 스위트가 초록으로 돌아오기 전까지는 이 행의 회귀가 스위트 수준 신호로는 드러나지 않는다 — 전용 테스트(internal/cli/binary_lag_test.go)가 유일한 가드다"
  - "advisory 는 250ms 안에 못 끝나면 조용히 사라진다. 병리적으로 느린 저장소에서는 지연이 있어도 아무 말도 하지 않으며, 그것이 설계다(fail-open) — 다만 「말이 없다」가 「지연이 없다」로 읽힐 위험은 남는다"
  - "`moai version` 제목 줄은 여전히 태그 바닥값을 읽는다. SPEC이 수용한 비용이고 별도 카드 소관"
```
