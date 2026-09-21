---
id: SPEC-SEAM-GREENFIELD-002
title: "progress — greenfield 씨앗의 flow 스타일 고착 (t1050)"
version: "0.1.0"
created: 2026-09-20
updated: 2026-09-22
author: GOOS
module: "internal/settings/yamlpatch"
tier: S
---

> 본 산출물은 `status:` 필드를 두지 않는다(SPEC 디렉터리의 상태축은 `spec.md` 단일 소관).

## §E.1 Plan-phase Audit-Ready Signal

- **카드**: t1050 (Class B — 결함, 원인 특정 완료).
- **트리**: 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1050`, 브랜치 `WT-greenfield-style`, plan-phase 기준 HEAD `159dd30df` (2026-09-20).
- **산출물**: `spec.md`(REQ-SGF2-001..009, 증거 사슬 E1..E12) · `plan.md`(M1..M5) · `acceptance.md`(AC-SGF2-001..006 정본) · 본 파일.
- **plan-phase에서 코드 파일은 수정하지 않았다** — Go 소스 0건.
- **배차 전제 대비 정정 3건**을 `spec.md` §1.4에 기록했다(등록부 12 엔트리 / seam 도달 6섹션 / `@MX:ANCHOR` 호출점 분포).
- **측정 귀속 구분**: E1·E2·E12는 배차 레인 세션의 임시 프로브 실측(프로브는 삭제됨), E4~E11은 본 저작 세션의 직접 재측정, E3은 상류 카드 t1043(다른 트리 `0bf27ea69`)의 인용, E8은 구조적 추론. 각 행에 관측 방법을 명기했다.
- **미해결 NEEDS CLARIFICATION 마커**: 없음. (대괄호를 풀어 적는다 — 게이트의 탐지식이 대괄호 포함 문자열을 훑으므로, 부재 선언이 탐지어를 그대로 품으면 기계 소비자가 존재로 읽는다.)
- **plan-audit iter1 (2026-09-20)**: FAIL 0.74 (Tier S 임계 0.75). blocking 4건(D1 판별 셀 공허 · D2 REQ 9>8 · D2-b 미커버 REQ · D3 측정 기록 미인용) + optional 4건(D5-D8). 판정서: `.moai/reports/t1050/plan-audit-verdict.md`. 2026-09-21 수정 완료 — blocking 4건 + optional 4건 전부 반영(상세는 아래 「수정 기록」).
- **수정 기록 (2026-09-21)**: D1 = `acceptance.md` AC-SGF2-003b 신설(edit을 upsert로 고정, 단정을 원본 대비 바이트 비교로 고정, 코디네이터 프로브 verbatim 4행 인라인 보존) + `plan.md` B-a/M2-3/M4-1 동기화 · D2 = REQ-SGF2-009 삭제(§6로 접음), REQ 8건 · D2-b = §3 색인표에 「커버하는 REQ」 열 + traceability 단정 추가 · D3 = §1.3 E1/E2/E12 행과 §7에 `.moai/reports/t1050/plan-measurements.md` 인용(그 파일 §C3은 기각된 해석임을 명시) · D4 = REQ-SGF2-003의 "seam section root key" → "`PatchFile` root key" · D5 = 스테일 수치 4곳 전부 열거 · D6 = "관례적" 주장 철회 + 사용자 스토리 축소 · D7 = 본 행의 대괄호 해제 · D8 = 001 패치의 `updated` 동결과 `—` 표기가 의도임을 001 HISTORY 행에 명시.

## §E.2 Run-phase Evidence

### M1 — 수리 형태의 경계 확정 (트리 `0a23f6a5d`, 2026-09-22)

**결정: `spec.md` §4의 대안 A — greenfield 경로를 취했음을 지역 플래그로 기록하고, 파싱 후 루트 노드의 flow 스타일만 해제한다.** 대안 B(씨앗 리터럴 자체를 block 형태로 교체)는 **측정으로 기각됐다**.

레인 세션이 「읽어서 얻은 소견」으로 전달한 대안 B 차단 가설을 그대로 채택하지 않고 프로브로 재측정했다. 프로브(`internal/settings/yamlpatch/zz_probe_m1_test.go`)는 측정 후 삭제했다.

**M1(c) — 대안 B는 구조적으로 막혀 있다 (측정 확인).** YAML은 빈 매핑을 `{}` 로만 철자하므로, block 형태의 빈 문서 씨앗은 `PatchFile` 자신의 가드에 걸린다:

```
M1c seed="{}\n"     unmarshalErr=<nil> kind=1(doc=1) contentLen=1
M1c seed=""         unmarshalErr=<nil> kind=0(doc=1) contentLen=0
M1c seed="\n"       unmarshalErr=<nil> kind=0(doc=1) contentLen=0
M1c seed="---\n"    unmarshalErr=<nil> kind=1(doc=1) contentLen=1
M1c PatchFile on pre-written seed ""     => err=yamlpatch: …/x.yaml: empty or non-document YAML
M1c PatchFile on pre-written seed "\n"   => err=yamlpatch: …/x.yaml: empty or non-document YAML
M1c PatchFile on pre-written seed "---\n" => err=yamlpatch: …/x.yaml: top-level node is not a mapping
```

`""` · `"\n"` 은 `Kind=0`(DocumentNode 아님) + `Content` 비어 있음 → `empty or non-document YAML`. `"---\n"` 은 DocumentNode 이지만 루트가 null 스칼라 → `top-level node is not a mapping`. 세 후보 모두 씨앗으로 쓸 수 없다. 레인의 읽기 기반 소견은 **측정으로 확인됐다**(가설 → 관측).

**M1(a) — greenfield 문서는 항상 재직렬화 폴백으로 간다 (측정, 가정 아님).** 경로 길이 3종에서 `lineSplice` 가 모두 `ok=false` 를 냈다:

```
M1a single-segment => ok=false err=<nil> out=""
M1a two-segment    => ok=false err=<nil> out=""
M1a four-segment   => ok=false err=<nil> out=""
M1a POSCTRL        => ok=true err=<nil> out="mcp:\n    enabled: false\n"
```

양성 대조(`POSCTRL`)가 `ok=true` 를 내므로 위 세 개의 `ok=false` 는 계측기 고장이 아니다. 기제: greenfield 문서는 키가 0개이므로 `findKey` 가 **첫 세그먼트에서** `-1` 을 반환하고 `lineSplice` 는 `idx < 0` 에서 upsert 로 판정해 폴백을 지시한다(`yamlpatch.go:147`). 경로 모양과 무관하게 도달 불가능하다.

**M1(c-배치) — 조건은 폴백 한쪽에만 걸면 충분하고, 빠른 경로에는 걸 것이 없다.** 이것은 「한쪽을 빠뜨렸다」가 아니라 구조적 사실이다: `lineSplice` 는 **재직렬화를 하지 않는다** — 원본 바이트의 대상 라인만 다시 쓰고 나머지는 그대로 둔다(`yamlpatch.go:173-182`). 그 경로에서 스타일은 결정되는 것이 아니라 바이트로 보존되므로, 걸 조건 자체가 존재하지 않는다. 게다가 M1(a)가 보였듯 greenfield 는 그 경로에 도달하지도 않는다. 두 이유는 독립이며 어느 하나만으로도 충분하다.

**M1(b) — `detectIndent([]byte("{}\n"))` = 4. 측정만 하고 고치지 않는다(범위 규율).**

```
M1b detectIndent("{}\n") = 4
M1b detectIndent("mcp:\n  a: 1\n") = 2
M1b detectIndent("mcp:\n    a: 1\n") = 4
M1b detectIndent("") = 4
```

`{}\n` 에는 들여쓰기 라인이 없어 `indentRe` 가 매치하지 않고 함수의 기본값 4로 떨어진다(`yamlpatch.go:332-341`). 형제 관례와 **어긋나지 않는다** — 관례가 단일하지 않기 때문이다: `spec.md` §1.1 의 plan-audit D6 실측이 seam 6종에서 2-space 4건 / 4-space 2건, 로컬 30개 전체에서 15 대 15 로 갈려 있음을 기록한다. 4는 그 두 값 중 하나이며 이상치가 아니다. 본 SPEC 의 AC 는 들여쓰기 폭을 단정하지 않으므로 이 값은 **기록 대상이지 수정 대상이 아니다**(`acceptance.md` §D). 들여쓰기 폭 축을 정하려면 별도 카드가 필요하다.

### M2 — RED 가드 채득 (트리 `0a23f6a5d`, 수리 이전)

신설 파일 `internal/settings/yamlpatch/greenfield_style_test.go` — 가드 4종. 커맨드:

```
$ go test ./internal/settings/yamlpatch/ -run 'TestPatchFileGreenfield|TestPatchFileExistingBlockByteInvariant|TestPatchFileDeliberateFlow' -v -count=1
EXIT=1
```

전체 verbatim: `.moai/reports/t1050/m2-red-capture.log`. 발췌 — **실패 출력에 실제 flow 형상이 보인다**(틀린 이유로 떨어지는 RED 가 아니다):

```
--- FAIL: TestPatchFileGreenfieldOutputIsBlockStyle (0.00s)
    --- FAIL: TestPatchFileGreenfieldOutputIsBlockStyle/mcp (0.00s)
    --- FAIL: TestPatchFileGreenfieldOutputIsBlockStyle/cacheStrategy (0.00s)
    --- FAIL: TestPatchFileGreenfieldOutputIsBlockStyle/report (0.00s)
    --- FAIL: TestPatchFileGreenfieldOutputIsBlockStyle/gate (0.00s)
    --- FAIL: TestPatchFileGreenfieldOutputIsBlockStyle/crosssession (0.00s)
--- FAIL: TestPatchFileGreenfieldSecondSaveStaysBlock (0.00s)

greenfield_style_test.go:70: root line = "{mcp: {tools: {session_list: {enabled: false}}}}", want "mcp:" (block mapping opener) — got:
    {mcp: {tools: {session_list: {enabled: false}}}}
greenfield_style_test.go:70: flow-mapping brace present in a block-style document — got:
    {mcp: {tools: {session_list: {enabled: false}}}}
greenfield_style_test.go:104: root line = "{mcp: {tools: {session_list: {enabled: false}, spec_audit: {enabled: false}}}}", want "mcp:" (block mapping opener) — got:
    {mcp: {tools: {session_list: {enabled: false}, spec_audit: {enabled: false}}}}
```

- **AC-SGF2-001**: 다섯 셀(`mcp`/`report`/`crosssession`/`gate`/`cacheStrategy`) **전부 FAIL**. 각 셀의 출력이 그 섹션의 실제 flow 한 줄을 그대로 보인다.
- **AC-SGF2-002**: FAIL. before/after 둘 다 flow이며, 두 키가 한 줄에 들어간 형상이 출력에 남았다.
- **AC-SGF2-003** / **AC-SGF2-003b**: 수리 전 **PASS**. 이것은 「정상」이 아니라 **수리가 아직 없어 재포맷할 주체가 없기 때문**이다(`plan.md` M2-3). 두 셀의 역할은 RED 채득이 아니라 M4 뮤턴트 아래에서 FAIL 하는 것이며, 003b 의 판별력은 M4 (b) 채득으로만 증명된다.
- **판별 셀 형태 확인**: `TestPatchFileDeliberateFlowPreservedOnUpsert` 의 edit 은 **upsert**(`mcp.tools.b.enabled` — 원본에 없는 키)이고 단정은 **전체 리터럴 `want` 바이트 비교**다(`acceptance.md` AC-SGF2-003b 의 허용 형태 2). 수리 전 PASS 가 그 리터럴(`{mcp: {tools: {a: {enabled: true}, b: {enabled: false}}}}\n`)이 코디네이터 프로브 측정과 일치함을 동시에 확인해 준다.

### M3 — 수리 + 스테일 주석 정정 (수리 커밋, RED 커밋 `d289130f1`의 뒤)

**수리 형태**: `PatchFile` 에 지역 플래그 `greenfield` 를 두고 absent 분기에서 `true` 로 세운 뒤, 루트 매핑 가드 직후 `if greenfield { root.Style = 0 }` 한 줄. 조건은 **재직렬화 폴백 쪽에만** 걸린다 — M1(c-배치)가 보인 대로 `lineSplice` 는 재직렬화를 하지 않으므로 그 경로에는 걸 조건이 존재하지 않는다.

**M2 가드 4종이 PASS 로 뒤집혔다**:

```
$ go test ./internal/settings/yamlpatch/ -run 'TestPatchFileGreenfieldOutputIsBlockStyle|TestPatchFileGreenfieldSecondSaveStaysBlock|TestPatchFileExistingBlockByteInvariant|TestPatchFileDeliberateFlowPreservedOnUpsert' -v -count=1
EXIT=0

--- PASS: TestPatchFileExistingBlockByteInvariant (0.00s)
--- PASS: TestPatchFileGreenfieldOutputIsBlockStyle (0.00s)
    --- PASS: TestPatchFileGreenfieldOutputIsBlockStyle/report (0.00s)
    --- PASS: TestPatchFileGreenfieldOutputIsBlockStyle/mcp (0.00s)
    --- PASS: TestPatchFileGreenfieldOutputIsBlockStyle/cacheStrategy (0.00s)
    --- PASS: TestPatchFileGreenfieldOutputIsBlockStyle/gate (0.00s)
    --- PASS: TestPatchFileGreenfieldOutputIsBlockStyle/crosssession (0.00s)
--- PASS: TestPatchFileDeliberateFlowPreservedOnUpsert (0.00s)
--- PASS: TestPatchFileGreenfieldSecondSaveStaysBlock (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.131s
```

**빈 스윕 대조**: `grep -c -- '--- PASS'` = **9** (최상위 4 + 서브테스트 5). 0 이 아니므로 이 초록은 「아무것도 고르지 않은」 초록이 아니다 — `-run` 필터가 없는 이름을 골랐다면 exit 0 + `ok` 를 그대로 내면서 0을 셌을 것이다.

**AC-SGF2-001 의 단정 형태**: 「루트만 block」이 아니라 **파일 어디에도 `{`/`}` 가 없다**를 함께 건다. 이 두 번째 조건이 루트만 de-flow 되고 중첩은 flow 로 남는 형상(`mcp: {tools: …}`)을 기각한다 — M4 (b) 뮤턴트가 정확히 그 형상이다.

**범위 패키지 전량**:

```
$ go test ./internal/settings/... -count=1
EXIT=0
ok  	github.com/modu-ai/moai-adk/internal/settings	0.621s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm	0.401s
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.408s
```

**AC-SGF2-006 — `@MX:ANCHOR` 주석 정정 (이 세션이 이 트리에서 직접 잰 값)**. 본 SPEC E5/E11 의 수치를 옮겨 적지 않았다. 재측정 커맨드와 출력(트리 `d289130f1`):

```
$ /usr/bin/grep -rn 'yamlpatch\.PatchFile(' --include='*.go' internal/ cmd/ pkg/
internal/settings/write_safety_test.go:35:	err := yamlpatch.PatchFile(path, []yamlpatch.KeyEdit{
internal/settings/write_safety_test.go:58:	err := yamlpatch.PatchFile(path, []yamlpatch.KeyEdit{
internal/settings/write_safety_test.go:326:	err := yamlpatch.PatchFile(path, []yamlpatch.KeyEdit{
internal/settings/write_safety_test.go:354:	err := yamlpatch.PatchFile(path, []yamlpatch.KeyEdit{
internal/settings/sectionwrite.go:18:// yamlpatch.PatchFile(노드 수술)만 사용한다. 라우팅 판정은 RouteForSection이
internal/settings/sectionwrite.go:76:	return yamlpatch.PatchFile(path, edits)
internal/cli/init_workflow_flags_test.go:234:	if err := yamlpatch.PatchFile(path, edits); err != nil {
internal/cli/init_workflow_flags.go:97:	if err := yamlpatch.PatchFile(workflowPath, edits); err != nil {

$ /usr/bin/grep -rn 'yamlpatch' internal/core/project/
exit=1                                  ← 무출력

$ /usr/bin/grep -c 'func' internal/core/project/initializer_expansion.go
13                                      ← 양성 대조: 계측기가 그 경로에 도달한다
```

적중 8행 중 **`sectionwrite.go:18` 은 호출이 아니라 주석 안의 언급**이다(적중 줄이 곧 선언 줄이 아니다 — 그 줄은 `// yamlpatch.PatchFile(노드 수술)만 사용한다` 로 시작한다). 그 한 줄을 제외한 판정:

- 프로덕션 호출점 = **2파일 2호출점** — `internal/settings/sectionwrite.go:76`, `internal/cli/init_workflow_flags.go:97`
- 테스트 호출점 = **2파일 5호출점** — `write_safety_test.go` ×4, `init_workflow_flags_test.go` ×1
- `internal/core/project` 의 `yamlpatch` 참조 = **0건**(exit 1 무출력, 양성 대조 13행이 나란히 있으므로 미측정이 아니라 부재다)

종전 주석의 「호출 파일 3개 5호출점 … initializer_expansion ×3」 은 세 축 모두 틀렸고, 정정본은 위 세 줄을 그대로 적는다.

**AC-SGF2-004 — 통제군 6건 무수정 GREEN**. 먼저 무수정임을 기계로 확인했다:

```
$ git diff --stat HEAD -- internal/settings/write_safety_test.go internal/settings/yamlpatch/yamlpatch_test.go
(무출력 — 두 파일 모두 미변경)
```

행 앵커도 content-token 으로 재검증했고 `spec.md` E10 의 기록값과 일치한다(29 / 52 / 320 / 341 / 395 / 426). 실행:

```
$ go test ./internal/settings/ -run 'TestPatchFileValueInvariantPreservesBytes|TestPatchFileScalarChangePreservesPresentation|TestPatchFileSpliceFallsBackForUpsert|TestPatchFileSpliceQuotedScalarChange' -v -count=1
EXIT=0
--- PASS: TestPatchFileScalarChangePreservesPresentation (0.00s)
--- PASS: TestPatchFileValueInvariantPreservesBytes (0.00s)
--- PASS: TestPatchFileSpliceQuotedScalarChange (0.00s)
--- PASS: TestPatchFileSpliceFallsBackForUpsert (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/settings	0.184s

$ go test ./internal/settings/yamlpatch/ -run 'TestPatchFileGreenfieldCreation|TestAtomicWriteStatErrorNotWidened' -v -count=1
EXIT=0
--- PASS: TestAtomicWriteStatErrorNotWidened (0.00s)
--- PASS: TestPatchFileGreenfieldCreation (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.131s
```

6건 전부 PASS 이고 기대값 전환은 없었다(무수정 확인이 그 증거다). 충돌이 하나도 없었다는 것은 수리 형태가 기존 계약과 어긋나지 않는다는 신호다.

### M4 — 뮤턴트 2방향 채득 (AC-SGF2-005, 트리 `5fbc3baee`)

두 뮤턴트 모두 커밋 트리 `5fbc3baee` 위의 오버레이로 적용했고, 채득 후 복원했다. 오버레이는 git 을 거치지 않고 `sed` 정·역 치환으로 넣고 뺐다 — 복원 증명은 `git diff --stat` 무출력이다.

#### (a) 스타일 해제 제거 — 결함 복귀를 잡는가

오버레이: `root.Style = 0` → `_ = root.Style // MUTANT-A: style-clearing removed`. (해제문만 없애고 `greenfield` 는 조건에 남겨 컴파일을 유지한다.)

```
$ go test ./internal/settings/yamlpatch/ -run 'TestPatchFileGreenfieldOutputIsBlockStyle|TestPatchFileGreenfieldSecondSaveStaysBlock|TestPatchFileExistingBlockByteInvariant|TestPatchFileDeliberateFlowPreservedOnUpsert' -v -count=1
EXIT=1

--- PASS: TestPatchFileExistingBlockByteInvariant (0.00s)
--- PASS: TestPatchFileDeliberateFlowPreservedOnUpsert (0.00s)
--- FAIL: TestPatchFileGreenfieldOutputIsBlockStyle (0.00s)
    --- FAIL: TestPatchFileGreenfieldOutputIsBlockStyle/gate (0.00s)
    --- FAIL: TestPatchFileGreenfieldOutputIsBlockStyle/cacheStrategy (0.00s)
    --- FAIL: TestPatchFileGreenfieldOutputIsBlockStyle/crosssession (0.00s)
    --- FAIL: TestPatchFileGreenfieldOutputIsBlockStyle/mcp (0.00s)
    --- FAIL: TestPatchFileGreenfieldOutputIsBlockStyle/report (0.00s)
--- FAIL: TestPatchFileGreenfieldSecondSaveStaysBlock (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.293s
```

**잡았다** — AC-SGF2-001 다섯 셀 전부 + AC-SGF2-002 가 FAIL. 보존 셀 둘은 영향 없이 PASS 이며, 이 뮤턴트가 보존 축이 아니라 생성 축의 변이임을 그 대비가 보인다. 전량: `.moai/reports/t1050/m4-mutant-a.log`.

#### (b) 조건 없이 무조건 해제 — 과잉 수리를 잡는가

**[HARD] 돌리기 전에 검출기 셀의 형태를 다시 읽고 확인했다**: `TestPatchFileDeliberateFlowPreservedOnUpsert` 의 edit 은 `mcp.tools.b.enabled` 이고 `b` 는 `before` (`{mcp: {tools: {a: {enabled: true}}}}\n`) 에 없다 → **upsert** 다. 단정은 전체 리터럴 `want` 와의 `got != want` → **바이트 비교**다. 두 성질 모두 살아 있으므로 이 실행은 「검출기가 없었다」가 아니라 실제 검출이다.

오버레이: `if greenfield {` → `if greenfield || true { // MUTANT-B: …`.

```
$ (같은 커맨드)
EXIT=1

--- PASS: TestPatchFileExistingBlockByteInvariant (0.00s)
--- FAIL: TestPatchFileDeliberateFlowPreservedOnUpsert (0.00s)
--- PASS: TestPatchFileGreenfieldOutputIsBlockStyle (0.00s)
    --- PASS: … /mcp /report /cacheStrategy /crosssession /gate (5 cells)
--- PASS: TestPatchFileGreenfieldSecondSaveStaysBlock (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.305s

greenfield_style_test.go:189: deliberate flow style not preserved across an upsert.
    --- got  ---
    mcp: {tools: {a: {enabled: true}, b: {enabled: false}}}
    --- want ---
    {mcp: {tools: {a: {enabled: true}, b: {enabled: false}}}}
```

**잡았다 — 그리고 AC-SGF2-003b **하나만**이 잡았다.** 생성 축 여섯 셀은 전부 PASS 로 남는다. `spec.md` §4 의 「인코딩 직전 무조건 해제」 구현이 기각되는 근거가 이 한 줄이며, 그 셀을 무력화하면 기각 근거가 함께 사라진다.

`got` 은 `spec.md` E13 의 코디네이터 측정과 **바이트 동일**하다 — 루트만 de-flow 되고 중첩 컬렉션은 flow 로 남는다(`mcp: {tools: {a: …}}`). 그리고 이 출력이 휴리스틱 단정이 왜 실패하는지도 함께 보인다: 뮤턴트 출력은 여전히 **한 줄**이고 **들여쓰기가 없다**. 「개행이 안 생겼는가 / 들여쓰기가 없는가」로 flow 보존을 단정했다면 뮤턴트 아래에서도 그 두 조건이 참이라 **PASS 하고 놓쳤을** 것이다. 바이트 비교만이 이 차이를 본다. 전량: `.moai/reports/t1050/m4-mutant-b.log`.

#### 복원 완전성

두 오버레이 각각의 채득 직후 역치환으로 복원했고, 최종 상태:

```
$ /usr/bin/grep -n 'MUTANT' internal/settings/yamlpatch/yamlpatch.go
grep-exit=1                             ← 무출력, 오버레이 잔재 0

$ git diff --stat
(무출력)

$ git status --short
(무출력)

$ git rev-parse --short HEAD
5fbc3baee

$ go test ./internal/settings/... -count=1
EXIT=0
```

#### 잡지 못한 뮤턴트 (REQ-SGF2-008)

**돌린 두 방향은 모두 잡혔다** — 기록할 미검출 실행은 없다. 이것은 「경계가 없다」는 뜻이 아니므로, 돌리지 않았고 따라서 **측정되지 않은** 가드 경계를 아래에 적는다. 셋 다 추론이며 관측으로 표시하지 않는다:

- **사용자가 디스크에 리터럴 `{}` 를 적어 둔 파일.** 그 문서는 `greenfield=false` 로 읽히므로 수리가 스타일을 건드리지 않고 flow 로 남는다. 어느 가드도 이 경우를 덮지 않는다 — 본 SPEC 이 고치는 것은 **씨앗에서 태어난 파일**의 형상이고, 사용자가 적은 `{}` 는 `spec.md` §5 의 「소급 재포맷 금지」 쪽에 속한다. 의도된 경계이지 구멍이 아니다.
- **중첩 노드 스타일까지 함께 해제하는 변이.** greenfield 문서의 중첩 노드는 `applyEdit` 이 만든 무스타일 노드라 해제가 no-op 이다. 가드는 이 변이를 구별하지 못하며, 구별할 필요도 없다(출력이 동일하다).
- **들여쓰기 폭 변이.** `detectIndent` 가 4 대신 2를 내도 모든 가드가 PASS 한다 — AC 가 폭을 단정하지 않기 때문이다(M1(b), `acceptance.md` §D). 그 축은 본 SPEC 의 가드 밖이며 별도 카드 소관이다.

## §E.3 Run-phase Audit-Ready Signal

> **측정 귀속 규율**: 아래 각 항목은 **누가 쟀는지**를 행마다 밝힌다. `[구현 세션]` 은 본 run-phase 구현 세션의 관측이고, `[레인]` 은 레인 세션이 직접 재서 전달한 관측이다. 구현 세션이 재지 않은 것을 자기 관측으로 적지 않는다 — 전달받은 수치를 무표기로 옮기면 2차 주장이 1차 관측으로 승격된다.

### E1 — AC PASS/FAIL 매트릭스

기준 트리: **`3386b8ae1`** (clean, `git status --porcelain` 무출력). 커맨드·verbatim 출력은 §E.2 의 해당 절에 있으며 여기서 중복하지 않는다.

| AC | 결과 | 커맨드 | 관측 요지 | 귀속 |
|----|------|--------|-----------|------|
| AC-SGF2-001 | **PASS** | `go test ./internal/settings/yamlpatch/ -run 'TestPatchFileGreenfieldOutputIsBlockStyle\|…' -v -count=1` | exit 0. 다섯 루트 키(`mcp`/`report`/`crosssession`/`gate`/`cacheStrategy`) 전부 PASS. 빈 스윕 대조 `grep -c -- '--- PASS'` = **9**(최상위 4 + 서브 5) | [구현 세션] |
| AC-SGF2-002 | **PASS** | 같음 | exit 0. 두 번째 저장 후에도 block 유지 + 첫 키 잔존 | [구현 세션] |
| AC-SGF2-003 | **PASS** | 같음 | exit 0. 기존 block 파일 스칼라 편집이 전체 리터럴 `want` 와 바이트 일치 | [구현 세션] |
| AC-SGF2-003b | **PASS** | 같음 | exit 0. upsert edit + 전체 리터럴 바이트 비교로 deliberate-flow 보존 | [구현 세션] |
| AC-SGF2-004 | **PASS** | `go test ./internal/settings/ -run '…4건' -v` + `go test ./internal/settings/yamlpatch/ -run '…2건' -v` | 양쪽 exit 0, 4 + 2 = 6 PASS. **무수정 증명**: `git diff d289130f1 -- write_safety_test.go yamlpatch_test.go` 무출력, 같은 형태가 `yamlpatch.go` 에는 29 insertions 를 내므로 공허한 무출력이 아니다 | [구현 세션] |
| AC-SGF2-005 | **PASS** | §E.2 M4 | 뮤턴트 2방향 채득 + 복원 완전성 증명. 아래 E9 | [구현 세션] + [레인] 독립 재현 |
| AC-SGF2-006 | **PASS** | `/usr/bin/grep -rn 'yamlpatch\.PatchFile(' …` 외 2건 | 주석 수치가 run 시점 재측정값이다. 양성 대조 동반 | [구현 세션] |

**`ac_pass_count` = 6 (논리 AC) / 라벨 셀 7. `ac_fail_count` = 0.**

### E2 — 크로스플랫폼 빌드 [레인]

```
$ GOOS=windows GOARCH=amd64 go build ./...
exit 0 (무출력)
```

트리 `3386b8ae1`. **이 항목은 레인 세션이 실행한 관측이며 구현 세션은 재지 않았다.**

### E3 — 범위 패키지 커버리지 [구현 세션]

```
$ go test -cover ./internal/settings/... -count=1
EXIT=0
ok  	github.com/modu-ai/moai-adk/internal/settings	0.613s	coverage: 90.6% of statements
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm	0.412s	coverage: 85.4% of statements
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.782s	coverage: 82.9% of statements
```

**[HARD] `yamlpatch` 82.9% 는 quality.yaml 의 패키지 목표 85% 에 2.1%p 미달이다 — 숨기지 않고 적는다.** 다만 **NEW 가 아니라 기존 baseline 이다**, 그리고 그 판정은 추측이 아니라 프로파일로 잰 것이다:

```
$ go tool cover -func=/tmp/t1050-cover.out
… PatchFile 90.0% · lineSplice 86.9% · replaceScalarInLine 71.4% · renderScalar 75.0%
   applyEdit 96.0% · findKey 100.0% · setScalar 25.0% · detectIndent 85.7%
   encode 88.9% · atomicWrite 72.0%
total: 82.9%
```

미달분은 `setScalar`(25.0%) · `atomicWrite`(72.0%) · `replaceScalarInLine`(71.4%) · `renderScalar`(75.0%) 네 함수에 몰려 있고 **넷 다 본 카드가 건드리지 않은 함수다**. 본 카드가 수정한 `PatchFile` 은 90.0% 이며, 프로파일의 count=0 블록 목록에 **내가 추가한 세 문장(`greenfield := false` · `greenfield = true` · `root.Style = 0`)은 들어 있지 않다** — 전부 새 가드가 실행한다. `PatchFile` 의 미커버 3블록(`70.26,72.4` · `95.59,97.3` · `132.16,134.3`)은 각각 non-absent read 오류 / lineSplice 오류 / encode 오류 분기이며 전부 기존 것이다.

산술 역산(같은 프로파일에서 추가 3문장을 빼면): **82.6% → 82.9%, 본 카드가 0.3%p 올렸다.** 이 값은 측정이 아니라 현재 프로파일로부터의 역산이며, 변경 전 트리에서 직접 재지 않았다 — 그 사실은 아래 Gap 에 적는다.

### E5 — Lint [레인]

```
$ golangci-lint run ./internal/settings/yamlpatch/...
exit 0
0 issues.
```

트리 `3386b8ae1`. **이 항목은 레인 세션이 실행한 관측이며 구현 세션은 재지 않았다.**

**[HARD] 이것은 패키지 한정 실행이지 저장소 전체 실행이 아니다.** 따라서 저장소 전역 lint baseline 은 **미측정**으로 남고, 그 판정은 CI 몫이다. 이 0 을 「저장소에 지적이 없다」로 읽으면 재지 않은 범위를 잰 것으로 승격시키는 오독이다. 구현 세션이 부수적으로 돌린 `go vet ./internal/settings/...`(exit 0)과 `gofmt -l`(무출력) 역시 범위 한정이며 lint 를 대체하지 않는다.

### E6 — 브랜치 / HEAD / 미푸시

```
$ git branch --show-current
WT-greenfield-style
$ git rev-parse --short HEAD
3386b8ae1
$ git rev-list --count develop..HEAD
4
$ git status --porcelain
(무출력)
```

**푸시하지 않았다** — 레인 규율대로 로컬 병합 SHA 보고까지가 소관이고 `origin/develop` push 는 리드의 일괄 행위다. 착지한 run-phase 커밋 3건(순서대로):

| SHA | 성격 |
|---|---|
| `d289130f1` | **RED** — 가드 4종, 수리 없음 |
| `5fbc3baee` | **수리** + `@MX:ANCHOR` 정정 |
| `3386b8ae1` | M4 뮤턴트 기록 |

**순서는 커밋 그래프가 증언한다** — RED 산출물이 수리 커밋보다 앞선 별도 커밋에 있으므로, 같은 커밋에 담았을 때 영구히 검증 불가가 되는 순서 주장이 여기서는 재확인 가능하다(`verification-claim-integrity.md` §2.3).

**변경 파일 전량 (vs plan 커밋 `0a23f6a5d`)**: 6개 — SPEC 산출물 4 + `greenfield_style_test.go`(신설) + `yamlpatch.go`. **PRESERVE 목록 잔존 확인**: `git diff --stat 0a23f6a5d..HEAD -- internal/settings/sectionapply.go internal/settings/sectionroute.go internal/settings/sectionwrite.go internal/web/ .moai/specs/SPEC-SEAM-GREENFIELD-001/` → **무출력**. `sectionwrite.go` 의 「8개 섹션」 스테일 독스트링 4곳과 등록부 12 vs 라우팅 6 불일치는 **관측만 하고 손대지 않았다**(`plan.md` §D 금지 사항).

### E8 — RED verbatim 채득

전량은 `.moai/reports/t1050/m2-red-capture.log`, 인용은 §E.2 M2 절에 있다 — 여기서 중복하지 않는다. 요지: exit 1, AC-SGF2-001 다섯 셀 + AC-SGF2-002 FAIL, **실패 출력이 실제 flow 형상(`{mcp: {tools: …}}`)을 그대로 보인다**. 틀린 이유로 떨어진 RED 가 아니다.

**[HARD] 귀속 정밀화**: 이 RED 는 HEAD `0a23f6a5d` + **가드 파일 미커밋 상태**에서 채득했다(그 내용이 곧 `d289130f1` 이 됐다). 커밋된 SHA 위에서 재실행한 값이 아니다 — 그러려면 체크아웃이 필요했고 하지 않았다. 순서 주장은 커밋 그래프가 받치지만, RED **출력**의 귀속 대상은 커밋이 아니라 워킹 트리다.

### E9 — 뮤턴트 2방향 + 미검출 뮤턴트

상세와 verbatim 은 §E.2 M4. 요지:

| 방향 | 오버레이 | 결과 | 귀속 |
|---|---|---|---|
| (a) 해제 제거 | `root.Style = 0` → `_ = root.Style` | exit 1 — AC-SGF2-001 다섯 셀 + AC-SGF2-002 FAIL, 보존 셀 2개 무영향 | [구현 세션] |
| (b) 무조건 해제 | `if greenfield {` → `if greenfield \|\| true {` | exit 1 — **AC-SGF2-003b 단독 FAIL**, greenfield 6셀 전부 PASS | [구현 세션] |
| (b) 재현 | `if greenfield {` 래퍼 삭제 + `_ = greenfield` | exit 1 — 같은 판정, 같은 `got` | **[레인] 독립 재현** |

**[레인] 재현이 드러낸 주의점 — 「삭제」 철자 하나에 국한된 이야기다.** (b) 는 두 철자 모두 유효하며, 둘 중 하나가 특별히 옳은 것은 아니다. 다만 **`if greenfield {` 래퍼를 통째로 삭제하는 철자**를 고르면 `greenfield` 가 미사용 변수가 되어 Go 가 빌드를 거부한다 — 그 철자를 쓰려면 변수를 살려 두는 줄(예: `_ = greenfield`)을 함께 넣어야 하고, 넣지 않으면 실행이 `[build failed]` 로 끝나 **아무것도 측정하지 않는다**. 구현 세션이 쓴 `if greenfield || true` 철자는 조건 안에서 변수를 그대로 쓰므로 **이 주의가 적용되지 않는다** — 별도의 줄이 필요 없다.

이것은 **(b) 뮤턴트의 성질이 아니라 한 철자의 성질**이다. 레인은 삭제 철자를 골라 `_ = greenfield` 로 해소했고, 구현 세션은 `|| true` 철자라 해소할 것이 없었다. 두 철자가 같은 판정과 같은 `got` 에 도달했다는 점이 이 결과를 한 세션의 우연한 표현에 의존하지 않게 만든다.

**이 주의를 기록하는 이유**: `[build failed]` 는 non-zero 로 끝나므로 「뮤턴트가 잡혔다」로 **오독되기 쉽다** — 가드가 FAIL 한 것이 아니라 가드가 **돌지 않은** 것이고, 그것은 검출이 아니라 미측정이다. 삭제 철자를 고르는 다음 사람이 이 지점을 그대로 밟는다.

**복원 완전성**: 구현 세션은 역 `sed` 치환(`grep -n 'MUTANT'` 무출력 + `git diff --stat` 무출력), 레인은 파일 복사(`git diff --stat` + `git status --porcelain` 양쪽 무출력). 두 세션 모두 git 을 거치지 않고 오버레이를 넣고 뺐다.

**미검출 뮤턴트 (REQ-SGF2-008): 없다** — 돌린 세 실행(구현 2 + 레인 1)이 모두 잡혔다. 「경계가 없다」는 뜻이 아니므로, **돌리지 않아 측정되지 않은** 가드 경계 3건을 §E.2 M4 절에 추론으로 명시해 뒀다(사용자가 적은 리터럴 `{}` 파일 / 중첩 스타일 동시 해제 / 들여쓰기 폭 변이). 셋 다 관측으로 표시하지 않았다.

### 선행 SPEC 패치 (`plan.md` §F M5-2)

**plan 단계에서 이미 적용돼 있어 재적용하지 않았다** — 「현재 상태부터 읽는다」는 M5-2 의 지시대로 읽기 전용으로 확인만 했다:

- `related_specs` 에 `SPEC-SEAM-GREENFIELD-002` 존재(`spec.md:15`)
- HISTORY 에 소유권 이관 행 존재(`spec.md:23`)
- `status: completed` 유지, `updated: 2026-09-08` **동결** — 그 행 자신이 「정확히 두 가지」 경계를 깨지 않기 위한 의도적 선택이라고 기록하고 있다

본 세션은 `SPEC-SEAM-GREENFIELD-001/` 을 **한 바이트도 수정하지 않았다**(E6 의 PRESERVE 무출력이 그 증거다).

### Audit-Ready Signal

```yaml
run_complete_at: 2026-09-22
run_commit_sha: pending-backfill-run   # 이 §E.3 커밋 자신의 해시는 커밋 시점에 알 수 없다
run_landed_commits: [d289130f1, 5fbc3baee, 3386b8ae1]
run_status: audit-ready
ac_pass_count: 6          # 논리 AC (라벨 셀 7 — 003b는 003의 하위 셀)
ac_fail_count: 0
preserve_list_post_run_count: 6        # plan.md §D PRESERVE 6항목 전부 잔존 (측정: E6 무출력)
new_warnings_or_lints_introduced: 0    # 단, 패키지 한정 측정 — 전역은 CI 몫 (E5)
cross_platform_build:
  windows_amd64: pass                  # [레인] 측정
  measured_by: lane
total_run_phase_files: 6               # SPEC 산출물 4 + 신설 테스트 1 + 수리 1
m1_to_mN_commit_strategy: three-commits-red-first
coverage_scope_packages:
  settings: 90.6
  settings_agentfm: 85.4
  settings_yamlpatch: 82.9             # 85% 목표 미달 — 기존 baseline, 본 카드가 +0.3%p
uncaught_mutants: 0
```

### Gaps — run 단계가 관측하지 않은 것

- **`./internal/cli/...` 미실행.** 구현 세션은 지시에 따라 돌리지 않았고, 레인이 백그라운드로 전량 실행 중이다. `internal/cli/init_workflow_flags.go:97` 이 수정된 함수의 **살아 있는 프로덕션 호출점**이므로 이 패키지는 실제로 영향권이며, 이 보고 시점에 미검증이다.
- **저장소 전역 lint 미측정** — E5 는 `./internal/settings/yamlpatch/...` 한정이다.
- **변경 전 커버리지를 직접 재지 않았다** — 82.6% 는 현재 프로파일로부터의 역산이지 이전 트리의 측정이 아니다.
- **RED 출력의 귀속 대상이 커밋이 아니라 워킹 트리다**(E8).
- **증거 로그가 gitignore 대상이다** — `.gitignore:227` 의 `.moai/reports/*` 가 `.moai/reports/t1050/*.log` 를 덮는다(`git check-ignore -v` 로 확인). 어떤 클론·CI 에도 닿지 않으므로 verbatim 을 추적 파일인 본 `progress.md` 에 인라인해 뒀고, `.log` 사본은 편의용이다.
- **`go test ./...` 미실행** — 레인 부하 규율에 따라 의도적이며, 전 패키지 판정은 CI 몫이다.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
