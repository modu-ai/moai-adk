---
id: SPEC-SEAM-GREENFIELD-002
title: "acceptance — greenfield 씨앗의 flow 스타일 고착 (t1050)"
version: "0.1.0"
created: 2026-09-20
updated: 2026-09-20
author: GOOS
module: "internal/settings/yamlpatch"
tier: S
---

> 본 산출물은 `status:` 필드를 두지 않는다(SPEC 디렉터리의 상태축은 `spec.md` 단일 소관 — frontmatter 스키마 SSOT § Artifact Statelessness).
>
> 본 파일이 AC의 **정본**이다. `spec.md §3`은 id와 한 줄 의도만 담는 색인이며, 충돌 시 이 파일이 지배한다.

## §A 측정 규율 (모든 AC 공통)

- 검증 범위: `go test ./internal/settings/... ./internal/cli/...`. **전체 스위트 로컬 실행 금지** — 전 패키지 판정은 CI 몫이다(`spec.md` §6; 요구가 아니라 검증 규율이다).
- 모든 AC의 PASS/FAIL 보고는 **커맨드 + verbatim 출력 + 트리 SHA**를 함께 적는다. 요약만 적힌 행은 Gap이지 Claim이 아니다.
- 신규 테스트는 `t.TempDir()` 안에서만 absent 트리를 구성한다 — 프로젝트 루트에 파일을 만들지 않는다.
- 부재 주장은 `/usr/bin/grep`으로 채득하고, 같은 경로에 대한 **양성 대조**(비어 있지 않은 출력)를 나란히 인용한다.

---

## §B AC 매트릭스

### AC-SGF2-001 — greenfield 출력이 block 스타일이다 (둘 이상의 섹션)

- **Given** 섹션 파일이 존재하지 않는 임시 트리(`t.TempDir()`),
- **When** 실변경 edit(예: absent bool 키에 명시 `false`)으로 `PatchFile`을 호출해 파일을 생성하고,
- **Then** 생성된 파일의 첫 유효 행은 `<root-key>:`로 시작하는 **block 매핑**이고, 파일 전체 어디에도 루트를 감싸는 `{` … `}` flow 매핑이 없다.
- **적용 범위**: 단일 섹션으로 충족되지 않는다 — 측정된 다섯 루트 키 `mcp` / `report` / `crosssession` / `gate` / `cacheStrategy` 전부에 대해 같은 단정을 건다(REQ-SGF2-003). 한 셀만 통과하고 나머지가 미측정이면 이 AC는 FAIL이다.
- **RED-now 셀**: 수리 전 트리(`159dd30df`)에서 다섯 셀 전부 FAIL이어야 하며, 실패 출력에 실제 flow 형상(`{mcp: {tools: …}}` 꼴)이 보여야 한다 — 틀린 이유로 떨어지는 RED는 채택하지 않는다.
- **Green 셀**: 수리 후 다섯 셀 전부 PASS.

### AC-SGF2-002 — 생성된 파일의 두 번째 저장도 block을 유지한다 (고착 해소)

- **Given** AC-SGF2-001의 경로로 **방금 생성된** 섹션 파일,
- **When** 같은 파일에 두 번째 실변경 edit(첫 번째와 다른 키를 upsert하는 형태)을 가하고,
- **Then** 결과 파일이 여전히 block 스타일이며, 첫 저장에서 기록된 키와 그 값이 그대로 남아 있다.
- **왜 별도 AC인가**: 고착은 생성과 다른 시점의 성질이다 — 생성만 고치고 두 번째 저장이 다시 flow를 만드는 수리는 AC-SGF2-001만으로는 통과한다. 이 셀이 그 구멍을 막는다.
- **RED-now 셀**: 수리 전 트리에서 before/after 둘 다 flow인 것을 verbatim 채득한다(상류 t1043 §C9가 같은 방향을 보였으나 그것은 **다른 트리의 인용**이며, 이 셀은 이 트리에서 직접 잰다).

### AC-SGF2-003 — 기존 파일의 바이트 보존

- **Given** 주석·빈 줄·중첩 키·미편집 키를 담은 **기존** block 스타일 섹션 파일,
- **When** 그중 한 스칼라만 바꾸는 edit을 `PatchFile`로 가하고,
- **Then** 편집 대상 스칼라를 제외한 **모든 바이트가 불변**이다 — 주석(head/line/foot), 빈 줄 위치, 키 순서, 모델링되지 않은 unknown 키, 인용 스타일, typed 스칼라 표현이 그대로다.
- **비교 방법**: 문자열 포함 검사가 아니라 **바이트 동일성**(`want`/`got` 전체 비교 또는 sha256)으로 재고, 불일치 시 diff를 출력한다.

#### AC-SGF2-003b — deliberate-flow 보존 (과잉 수리 판별 셀, [HARD] edit 형태가 지정된다)

- **Given** 사용자가 **일부러 flow 스타일로 적은 기존** 섹션 파일(예: `{mcp: {tools: {a: {enabled: true}}}}\n`),
- **When** **그 문서에 아직 없는 키를 추가하는 upsert edit**(예: `mcp.tools.b.enabled = false`)을 `PatchFile`로 가하고,
- **Then** 결과 파일이 **flow 스타일로 남아 있다** — 그리고 그 단정은 **원본 대비 바이트 비교**로 표현한다(아래 [HARD] 참조).

**[HARD] edit은 반드시 upsert여야 한다 — 스칼라 교체는 이 셀을 공허하게 만든다.** 스칼라 교체 edit은 `lineSplice` 빠른 경로에서 성립해 `atomicWrite`로 즉시 반환하므로 **스타일이 결정되는 재직렬화 경로에 도달하지 않는다.** 그러면 올바른 greenfield-한정 수리에서도, 과잉 적용 뮤턴트(b)에서도 출력이 같아져 판별력이 0이 된다. upsert는 `lineSplice`가 `idx < 0`에서 `ok=false`를 내므로 재직렬화 경로에 진입하고, 거기서 비로소 두 구현이 갈린다.

측정된 근거(배차 코디네이터 세션 프로브, 트리 `159dd30df`, 2026-09-20 — 씨앗 `{mcp: {tools: {a: {enabled: true}}}}\n`):

```
baseline  SCALAR-REPLACE on flow => "{mcp: {tools: {a: {enabled: false}}}}\n"
baseline  UPSERT         on flow => "{mcp: {tools: {a: {enabled: true}, b: {enabled: false}}}}\n"

mutant(b) SCALAR-REPLACE on flow => "{mcp: {tools: {a: {enabled: false}}}}\n"      <- baseline과 BYTE-IDENTICAL
mutant(b) UPSERT         on flow => "mcp: {tools: {a: {enabled: true}, b: {enabled: false}}}\n"   <- 다르다
```

**[HARD] 단정은 "block 스타일이 아니다"가 아니라 "원본 대비 바이트 비교"로 쓴다.** 위 측정이 보이듯 뮤턴트(b)에서 de-flow되는 것은 **루트 노드뿐**이고 중첩 컬렉션은 flow로 남는다(`mcp: {tools: {a: …}}`). 따라서 "개행이 들어갔는가", "들여쓰기가 있는가" 같은 substring·개행 휴리스틱은 **이 뮤턴트를 놓친다**(둘 다 뮤턴트(b) 출력에서 참이다) — 가장 먼저 떠오르는 단정이 하필 실패하는 단정이다. 유효한 형태는 둘뿐이다:

1. upsert 이전 원본과 upsert 이후 결과를 **바이트 단위로 비교**해, 차이가 삽입된 키에 국한되고 루트 줄 형상(`{`로 열리는 한 줄 매핑)이 보존됐음을 보인다. 또는
2. 기대 출력 전체를 리터럴로 고정한다(`want` = 위 `baseline UPSERT` 문자열).

- **판별 대상**: 이 셀은 `spec.md` §4의 "인코딩 직전 무조건 스타일 해제" 구현을 기각하는 **유일한** 판별 증거이며, AC-SGF2-005의 (b) 방향 뮤턴트가 자신의 검출기로 지목하는 셀이다. 두 장치가 이 한 셀로 수렴하므로, edit 형태와 단정 형태를 바꾸는 것은 곧 두 장치를 함께 무력화하는 일이다.
- **수리 전 트리에서의 상태**: 이 셀은 수리 전에도 PASS한다. 그것은 "정상"이 아니라 **수리가 아직 없어서 재포맷할 주체가 없기 때문**이며, 이 셀의 역할은 RED 채득이 아니라 AC-SGF2-005 (b) 뮤턴트 아래에서 FAIL하는 것이다. 수리 후 PASS만으로는 이 셀이 살아 있다는 증거가 되지 않는다 — (b) 뮤턴트 FAIL 채득이 그 증거다.

### AC-SGF2-004 — 기존 통제군 6건 무수정 GREEN

- **Given** 커밋된 수리 트리,
- **When** 다음 6개 테스트를 **수정 없이** 실행하고,
- **Then** 전부 PASS한다:
  - `TestPatchFileValueInvariantPreservesBytes` — `internal/settings/write_safety_test.go:29`
  - `TestPatchFileScalarChangePreservesPresentation` — `:52`
  - `TestPatchFileSpliceFallsBackForUpsert` — `:320`
  - `TestPatchFileSpliceQuotedScalarChange` — `:341`
  - `TestPatchFileGreenfieldCreation` — `internal/settings/yamlpatch/yamlpatch_test.go:395`
  - `TestAtomicWriteStatErrorNotWidened` — `:426`
- **[HARD] 기대값 전환 금지**: 이 6건의 기대값을 바꿔 통과시키는 것은 수리가 아니라 통제군 파괴다. 만약 어느 하나가 수리와 충돌한다면 그것은 **수리 형태가 틀렸다는 신호**이며, 기대를 고치기 전에 그 충돌을 `progress.md`에 적고 판단을 받는다.
- **행 앵커 주의**: 위 줄 번호는 트리 `159dd30df` 실측값이다 — run 착수 시 `func <name>` content-token으로 재검증한다(줄 인용은 낡는다).

### AC-SGF2-005 — 뮤턴트 채득 (가드 비-공허성의 유일한 판별 증거)

- **Given** 커밋된 수리 트리,
- **When** 수리 되돌림 오버레이를 커밋 트리 위에 적용하고 AC-SGF2-001·002의 신규 가드를 실행하고,
- **Then** 다음 네 가지를 모두 채득한다:
  1. 오버레이 상태에서 신규 가드가 **FAIL**로 떨어지는 것을 verbatim(커맨드 + exit code + 트리 SHA 명기),
  2. 오버레이를 복원하고 같은 커맨드가 **PASS**임을 확인,
  3. `git diff --stat` 무출력으로 복원 완전성 증명,
  4. 가드가 **잡지 못한** 뮤턴트가 하나라도 있으면 그 내용과 함께 `progress.md`에 기록(REQ-SGF2-008) — 못 잡은 뮤턴트는 삭제할 실패가 아니라 그 가드의 경계를 그리는 기록이다.
- **권장 뮤턴트 최소 2방향**: (a) 스타일 해제 호출을 통째로 제거(가드가 결함 복귀를 잡는가), (b) 스타일 해제를 greenfield 조건 없이 **무조건** 수행(가드가 **AC-SGF2-003b**의 deliberate-flow 셀로 과잉 수리를 잡는가). 한 방향만 돌리면 "위험한 변이"와 "무해한 no-op"이 구별되지 않는다.
- **[HARD] (b) 방향의 검출기는 AC-SGF2-003b뿐이고, 그 셀이 upsert edit + 바이트 비교 형태일 때만 검출한다.** (b)를 돌리기 전에 그 셀이 여전히 그 형태인지 확인한다 — 스칼라 교체로 되돌아간 셀은 (b) 아래에서도 PASS하므로, 뮤턴트가 "잡혔다"가 아니라 "검출기가 없었다"가 된다. 그 경우는 항목 4의 기록 대상이다.

### AC-SGF2-006 — `@MX:ANCHOR` 주석이 실측 호출점 분포와 일치한다

- **Given** `internal/settings/yamlpatch/yamlpatch.go`의 `PatchFile` 위 `@MX:ANCHOR` 블록(`:52-53`, 현재 "호출 파일 3개 5호출점(sectionwrite, initializer_expansion ×3, init_workflow_flags)"),
- **When** 수리가 착지하고,
- **Then** 그 주석이 **재측정된 분포**를 서술한다 — 프로덕션 호출점 2파일 2호출점(`internal/settings/sectionwrite.go`, `internal/cli/init_workflow_flags.go`)이며, `internal/core/project`는 `yamlpatch`를 참조하지 않는다.
- **증거 요구**: 주석에 적히는 수치는 run 단계가 **그 시점 트리에서 직접 잰 값**이어야 한다 — 본 SPEC의 E5/E11 수치를 그대로 옮겨 적지 않는다(트리가 움직였을 수 있다). 재측정 커맨드와 출력, 그리고 `internal/core/project` 부재 주장에 대한 양성 대조를 함께 채득한다.

---

## §C 완료 게이트 (Definition of Done)

- [ ] AC-SGF2-001..006 전부 PASS — 각 행이 커맨드 + verbatim 출력 + 트리 SHA를 인용
- [ ] AC-SGF2-001·002의 **RED-now** 채득이 수리 커밋보다 앞선 커밋에 존재(순서 주장은 커밋 그래프만이 증언한다)
- [ ] AC-SGF2-005 뮤턴트 2방향 채득 + 복원 완전성(`git diff --stat` 무출력)
- [ ] `go test ./internal/settings/... ./internal/cli/...` GREEN — NEW 실패와 기존 baseline 실패를 구분해 보고
- [ ] `golangci-lint run` NEW 지적 0 (기존 baseline과 구분)
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- [ ] 커밋 메시지에 카드 id `t1050` + Conventional Commits + `🗿 MoAI` 트레일러
- [ ] `progress.md` §E.2/§E.3에 증거 기록 — 못 잡은 뮤턴트가 있으면 그 사실 포함

## §D 전방 점검 (run 단계가 먼저 물어야 할 것)

- 수리를 **greenfield 경로에만** 거는 조건 표현이 `lineSplice` 빠른 경로와 재직렬화 폴백 **양쪽**에 올바로 걸리는가 — greenfield 문서는 통상 폴백으로 가지만(키 미해소), 그 경로 가정을 측정 없이 전제하지 않는다.
- `detectIndent(data)`가 `{}\n` 씨앗에 대해 무엇을 반환하는지 — 들여쓰기 폭 역시 greenfield 경로의 암묵 기본값이다(§8 계열 관측). 본 SPEC의 AC는 들여쓰기 폭을 단정하지 않으므로, 측정 결과가 형제 파일 관례와 어긋나면 **고치지 말고 기록**한 뒤 별도 카드로 올린다.
