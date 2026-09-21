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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
