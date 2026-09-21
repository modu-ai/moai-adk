---
id: SPEC-SEAM-GREENFIELD-002
title: "greenfield 씨앗의 flow 스타일 고착 — `{}\\n` 한 줄 문서가 seam 섹션 파일 형상을 결정한다 (t1050)"
version: "0.1.0"
status: draft
created: 2026-09-20
updated: 2026-09-20
author: GOOS
priority: P2
phase: "v3.2.0 target"
module: "internal/settings/yamlpatch"
lifecycle: spec-anchored
tier: S
tags: "yamlpatch, greenfield, flow-style, block-style, seam, section-file, byte-preservation, mutant-capture, defect"
related_specs: [SPEC-SEAM-GREENFIELD-001, SPEC-WEB-WRITE-SAFETY-001, SPEC-MCP-CONSOLE-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-20 | GOOS | 최초 draft. 카드 t1050 (Class B — 결함, 원인 특정 완료). 상류 증거는 카드 t1043 판정서 `.claude/worktrees/t1043/.moai/reports/t1043/verdict.md` §C9(「greenfield 파일의 flow 스타일 고착 (신규)」) — 그 카드가 범위 밖으로 남긴 신규 결함이 본 SPEC의 대상이다. Tier S이나 디스패치 지시에 따라 `acceptance.md`를 함께 둔다(§3 참조). 배차 전제 대비 정정 3건을 §1.4에 기록했다 — seam 섹션 등록부 개수(11→12), `WriteSectionViaSeam`으로 실제 도달 가능한 섹션 수(6), `@MX:ANCHOR` 스테일 주석의 정확한 호출점 분포. |

---

## §1 Context & Motivation

### §1.1 관측된 결함

seam 섹션 파일이 **존재하지 않는** 탭에서 첫 저장을 하면 파일은 정상적으로 생성된다(그 생성 자체는 SPEC-SEAM-GREENFIELD-001이 수리한 결과다). 그러나 생성된 파일은 **flow 스타일 한 줄 YAML**이다:

```
{mcp: {tools: {session_list: {enabled: false}}}}
```

block 스타일이 아니다 — 아래는 대비용 예시이며 "관례"의 제시가 아니다. 들여쓰기 폭은 형제 섹션 파일들 사이에서 갈려 있어(plan-audit 판정서 D6 실측: seam 6종 중 2-space 4건 / 4-space 2건, 로컬 30개 전체로는 15 대 15) 단일 관례가 존재하지 않는다. 본 SPEC이 단정하는 것은 **flow 대 block**이라는 축 하나이고, 들여쓰기 폭 축은 run 단계가 재서 기록만 한다(`acceptance.md` §D):

```
mcp:
    tools:
        session_list:
            enabled: false
```

그리고 그 형상은 **고착된다** — 같은 파일에 두 번째, 세 번째 저장을 가해도 flow 스타일이 유지된다. 한 번 잘못된 형상으로 태어난 파일은 사람이 손으로 고치기 전까지 그 형상을 물려준다.

### §1.2 원인 — `{}\n` 씨앗의 스타일이 루트 노드에 상속된다

`PatchFile`은 absent 파일을 greenfield 문서로 시작하면서 `data = []byte("{}\n")`을 씨앗으로 쓴다(`internal/settings/yamlpatch/yamlpatch.go:74`). `{}`는 YAML에서 **flow 스타일 매핑**이고, yaml.v3 인코더는 루트 노드의 스타일을 보존한다. 따라서 그 씨앗에서 자란 문서는 인코딩될 때 flow로 직렬화된다. 씨앗 한 줄이 파일의 평생 형상을 정한다.

고착의 기제는 별도의 결함이 아니라 같은 결함의 2차 효과다 — 두 번째 저장은 이미 flow인 파일을 읽어 그 스타일을 다시 보존한다.

### §1.3 측정된 증거 사슬

| # | 사실 | 위치 | 어떻게 관측했는가 |
|---|------|------|------------------|
| E1 | greenfield 출력이 flow 스타일이다 — 다섯 seam 섹션 루트 키 전부에서 동일. `GREENFIELD mcp => "{mcp: {tools: {x: {enabled: false}}}}\n"`, 동일 형태로 `report` / `crosssession` / `gate` / `cacheStrategy` | `internal/settings/yamlpatch/yamlpatch.go` `PatchFile` | **배차 레인 세션의 임시 프로브 테스트 실측** (트리 `159dd30df`, 워크트리 t1050, 2026-09-20). 프로브는 측정 후 삭제됐고 측정 종료 시점 `git status --porcelain` 0행. **verbatim 출력의 유일한 잔존 기록**: `.moai/reports/t1050/plan-measurements.md` §C1 |
| E2 | 양성 대조 — 같은 edit을 **기존 block 스타일 씨앗** 위에 가하면 block으로 유지된다: `"mcp:\n    tools:\n        a:\n            enabled: true\n        b:\n            enabled: false\n"` | 같음 | **배차 레인 세션의 같은 프로브 실측** (같은 트리·같은 실행). E1의 무-block 출력이 계측기 고장이 아니라 결함임을 가르는 증거. verbatim: `.moai/reports/t1050/plan-measurements.md` §C2 |
| E3 | 고착 — 생성된 flow 파일에 두 번째 저장을 가해도 flow가 유지된다 (`before: '{mcp: {tools: {session_list: {enabled: false}}}}\n'` → POST /save 200 → `after: '{mcp: {tools: {session_list: {enabled: false}, spec_audit: {enabled: false}}}}\n'`), 같은 편집의 block-씨앗 양성 대조는 block 유지 | 같음 | **상류 카드 t1043의 실측 인용** — `.claude/worktrees/t1043/.moai/reports/t1043/verdict.md` §C9 (그 판정서의 baseline은 워크트리 t1043, HEAD `0bf27ea69`). 본 SPEC 세션에서 재측정하지 않았다 |
| E4 | greenfield 씨앗 리터럴은 `data = []byte("{}\n")` 한 곳이다 | `internal/settings/yamlpatch/yamlpatch.go:74` | **본 SPEC 저작 세션 직접 판독** (`/usr/bin/grep -n 'data = \[\]byte("{}' …` → 1행) |
| E5 | `PatchFile`은 공유 진입점이다 — 프로덕션 호출점은 **2파일 2호출점**: `internal/settings/sectionwrite.go:76`(`WriteSectionViaSeam` 말미), `internal/cli/init_workflow_flags.go:97`. 테스트 호출점 5건은 별도(`internal/settings/write_safety_test.go` ×4, `internal/cli/init_workflow_flags_test.go` ×1) | 같음 | **본 SPEC 저작 세션 직접 측정** — `/usr/bin/grep -rn 'yamlpatch\.PatchFile(' --include='*.go' internal/ cmd/ pkg/` |
| E6 | seam 섹션 등록부 `sectionRootKeys`는 **12 엔트리**다(cache, crosssession, feedback, gate, handoff, harness, mcp, observability, ralph, report, security, workflow) | `internal/settings/sectionwrite.go` | **본 SPEC 저작 세션 직접 측정** — `awk '/^var sectionRootKeys/,/^}/' … \| /usr/bin/grep -cE '^\t"[a-z]+":'` → `12` (비어 있지 않은 출력이 곧 양성 대조) |
| E7 | 그 12 중 `WriteSectionViaSeam` 게이트(`sectionwrite.go:63`, `RouteForSection(section) != RouteSeam` → 거부)를 실제로 통과하는 것은 **6개**다: `report`, `mcp`, `crosssession`, `workflow`, `feedback`, `gate`. 나머지 6개(harness, ralph, observability, security, handoff, cache)는 `RouteExcluded`라 seam 경로로 도달하지 않는다 | `internal/settings/sectionroute.go`, `internal/settings/sectionwrite.go:63` | **본 SPEC 저작 세션 직접 판독** — `/usr/bin/grep -rn 'RouteSeam' internal/settings/*.go` + 맵 본문 판독 |
| E8 | **구조적 사실**(측정이 아님): 씨앗은 `PatchFile` 한 곳에 있으므로, `PatchFile`에 도달하는 모든 경로 — E7의 seam 6섹션과 E5의 `init_workflow_flags.go` 경로 — 가 같은 씨앗을 공유한다. E1이 측정한 것은 그중 다섯 루트 키이고, 나머지 경로의 동일성은 코드 형상에서 따라 나오는 추론이다 | 같음 | **구조적 추론** — 측정으로 표시하지 않는다 |
| E9 | 현행 테스트 중 flow 형상을 기대값으로 인코딩한 것은 **없다** | `internal/**/*_test.go` | **본 SPEC 저작 세션 직접 측정** — `/usr/bin/grep -rn '{mcp:\|{report:\|{gate:\|{crosssession:\|{cacheStrategy:' internal/ --include='*_test.go'` → 0행. 양성 대조: 같은 도구로 `mcp` 단어를 같은 두 파일에서 세면 2건·7건이 나온다(계측기 도달 확인) |
| E10 | 통제군 6개 테스트가 명명된 위치에 실재한다 — `TestPatchFileValueInvariantPreservesBytes`(`internal/settings/write_safety_test.go:29`), `TestPatchFileScalarChangePreservesPresentation`(`:52`), `TestPatchFileSpliceFallsBackForUpsert`(`:320`), `TestPatchFileSpliceQuotedScalarChange`(`:341`), `TestPatchFileGreenfieldCreation`(`internal/settings/yamlpatch/yamlpatch_test.go:395`), `TestAtomicWriteStatErrorNotWidened`(`:426`) | 같음 | **본 SPEC 저작 세션 직접 측정** — 함수 선언 앵커(`func <name>`) grep |
| E11 | `PatchFile`의 `@MX:ANCHOR` 주석이 스테일하다 — "호출 파일 3개 5호출점(sectionwrite, initializer_expansion ×3, init_workflow_flags)"이라고 적혀 있으나, `internal/core/project`에는 `yamlpatch` 참조가 **0건**이다 | `internal/settings/yamlpatch/yamlpatch.go:52-53` | **본 SPEC 저작 세션 직접 측정** — `/usr/bin/grep -rn 'yamlpatch' internal/core/project/` exit 1(무출력). 양성 대조: 같은 경로에 `/usr/bin/grep -c 'func' internal/core/project/initializer_expansion.go` → `13`(공백 포함 `'func '`는 `11`) — 계측기가 그 경로에 도달함을 보인다 |
| E12 | 수리 방향 검증 — greenfield 씨앗 경로를 취했음을 기록하고 파싱 후 루트 노드의 flow 스타일을 해제(`root.Style = 0`)하면, 다섯 섹션 전부 block 출력으로 바뀌고 E2의 block-씨앗 양성 대조는 바이트 불변이며 `go test ./internal/settings/... -count=1`이 세 패키지(`settings`, `settings/agentfm`, `settings/yamlpatch`) 모두 GREEN이다. 검증 후 되돌렸다 | 같음 | **배차 레인 세션의 실측 후 revert** (같은 트리·같은 실행). verbatim + 패키지별 실행 시간: `.moai/reports/t1050/plan-measurements.md` §C4. 구현 형태 확정은 run 단계 소관 — 본 SPEC은 이를 plan-phase 소견으로만 적는다 |

| E13 | 판별력 측정 — 기존 flow 씨앗 `{mcp: {tools: {a: {enabled: true}}}}\n` 위에서, **스칼라 교체** edit은 올바른 수리와 과잉 적용 뮤턴트(b) 아래 출력이 **바이트 동일**하고, **upsert** edit만 둘을 가른다. 또한 뮤턴트(b)에서 de-flow되는 것은 **루트 노드뿐**이고 중첩 컬렉션은 flow로 남는다(`mcp: {tools: {a: …}}`) — 개행·들여쓰기 휴리스틱으로는 이 뮤턴트를 놓친다 | `internal/settings/yamlpatch/yamlpatch.go` `lineSplice` / 재직렬화 분기 | **배차 코디네이터 세션의 프로브 실측** (트리 `159dd30df`, 2026-09-20). 프로브 삭제됨. verbatim 4행은 `acceptance.md` AC-SGF2-003b에 인라인 보존 — 그 블록이 이 측정의 잔존 기록이다 |

**Baseline 귀속**: E1·E2·E12는 배차 레인 세션이 이 워크트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1050`, 브랜치 `WT-greenfield-style`, HEAD `159dd30df`, 2026-09-20)에서 임시 프로브로 채득한 뒤 프로브를 삭제한 값이다. E4~E11은 본 SPEC 저작 세션이 같은 트리에서 직접 재측정한 값이다. E13은 배차 코디네이터 세션이 같은 트리에서 채득한 실측이다. E3은 상류 카드 t1043의 다른 트리(`0bf27ea69`) 측정의 인용이며 본 세션에서 재측정하지 않았다 — 인용으로 표시한다. E8은 측정이 아니라 구조적 추론이다.

**측정 기록 파일**: E1·E2·E12의 프로브는 측정 후 삭제됐으므로 그 verbatim 출력이 남은 곳은 `.moai/reports/t1050/plan-measurements.md`(§C1·§C2·§C4) 하나뿐이다. run 단계와 감사는 그 파일을 열어 본 표의 인용을 대조한다. [HARD] 같은 파일의 §C3(「등재 11종 전부가 `WriteSectionViaSeam`을 거친다」)은 **2026-09-20자 정정 블록으로 기각된 해석**이다 — 관측(§C1·§C2의 프로브 출력)은 그대로 유효하고 해석만 틀렸다. 정정된 사실이 본 문서 §1.4-1·2이며(등재 12종, seam 도달 6종), §C3 원문을 인용하지 않는다. 잔여 위험: 그 파일은 `.gitignore`의 `.moai/reports/*`로 무시되어 어떤 클론·CI에도 닿지 않는다 — 추적 여부는 이 저장소의 증거 보관 정책 축이며 본 SPEC이 푸는 문제가 아니다.

### §1.4 배차 전제 대비 정정 (3건)

1. **"11 seam sections registered in `sectionRootKeys`"** → 실측 **12** 엔트리다(E6).
2. **"all 11 … reach `PatchFile` through `WriteSectionViaSeam`"** → `WriteSectionViaSeam`은 `RouteForSection != RouteSeam`을 거부하므로 그 경로로 실제 도달하는 섹션은 **6개**다(E7). 등록부 등재와 seam 도달 가능성은 다른 축이다. E1이 측정한 다섯 루트 키 중 `cacheStrategy`(cache)는 `RouteExcluded`이므로 `WriteSectionViaSeam`으로는 도달하지 않는다 — 그 셀은 `PatchFile` 직접 호출로 측정된 것이며, 씨앗 결함이 `PatchFile` 층에 있다는 사실을 오히려 더 넓게 보인다.
3. **`@MX:ANCHOR` 주석의 호출점 분포** → "3파일 5호출점 … initializer_expansion ×3"은 스테일하다. 프로덕션 호출점은 2파일 2호출점이고(E5) `initializer_expansion`에는 `yamlpatch` 참조가 0건이다(E11). 이 주석은 고쳐질 바로 그 함수 위에 앉아 있으므로 **본 SPEC의 범위 안**이다.

### §1.5 판정 방향 — 씨앗이 형상 계약을 은밀히 결정한다

`{}\n`은 "빈 매핑 문서"라는 **의미**를 전달할 의도의 리터럴이었고, 그것이 **스타일**까지 전달한다는 것은 의도된 계약이 아니다. 근거: 선행 SPEC-SEAM-GREENFIELD-001의 `§5 Out of Scope`는 세 가지 제외(seam 외 쓰기 경로 / renderer / absent 쌍 전수 스윕)를 명시하는데 **출력 스타일은 그중에 없다.** 즉 flow 형상은 검토 후 수용된 결정이 아니라 **검토되지 않은 귀결**이다. 따라서 001은 `completed`로 남고(그 SPEC이 선언한 범위는 실제로 닫혔다), 형상 축은 본 SPEC이 소유한다.

---

## §2 Requirements (GEARS)

**사용자 스토리**: "나(moai-adk 사용자)가 아직 섹션 파일이 없는 탭에서 처음 설정을 저장하면, 생성된 `.moai/config/sections/<section>.yaml`은 한 줄 flow가 아니라 **block 스타일 YAML**이다 — 손으로 형상을 고칠 일이 없고, 이후 저장도 그 형상을 유지한다." (「형제 파일과 같은」이라고 쓰지 않는다 — 들여쓰기 축에서는 형제 관례가 갈려 있어 그 약속이 결정 불가능하다. §1.1 참조.)

### REQ-SGF2-001 (Ubiquitous)

The `PatchFile` function of the `internal/settings/yamlpatch` package shall write a greenfield-created section file as a block-style YAML document — the flow-style single-line form shall not be the created shape.

### REQ-SGF2-002 (Event-driven)

When `PatchFile` creates a file from the greenfield seed and that file is subsequently edited again, the second write shall also produce block-style output — the created shape shall not require a manual reformat to recover.

### REQ-SGF2-003 (Ubiquitous)

The greenfield block-style guarantee shall hold for more than one **`PatchFile` root key** — the measured set is `mcp`, `report`, `crosssession`, `gate`, `cacheStrategy` (E1). The wording is deliberately NOT "seam section root key": `cacheStrategy` (the `cache` section) is `RouteExcluded` and never reaches `PatchFile` through `WriteSectionViaSeam` (E7), so that cell was measured against `PatchFile` directly. The broader term is the accurate one — the seed defect lives below the routing gate — and the measured set is the probe's incidental selection, neither the seam set (which is `report` / `mcp` / `crosssession` / `workflow` / `feedback` / `gate`) nor a principled subset of it.

### REQ-SGF2-004 (Unwanted)

A write to an **existing** section file shall not change its bytes beyond the edited scalars — comments, blank lines, key order, unknown keys, quoted style, and typed scalars shall survive unchanged, and an existing flow-style document that a user authored deliberately shall not be reformatted by this change.

### REQ-SGF2-005 (State-driven)

While the fix is reverted (a mutant overlay on the committed tree), the new greenfield-style guard shall FAIL — the guard shall be shown to be non-vacuous rather than merely green.

### REQ-SGF2-006 (Ubiquitous)

The control family shall stay green unmodified: `TestPatchFileValueInvariantPreservesBytes`, `TestPatchFileScalarChangePreservesPresentation`, `TestPatchFileSpliceFallsBackForUpsert`, `TestPatchFileSpliceQuotedScalarChange` (`internal/settings/write_safety_test.go`), `TestPatchFileGreenfieldCreation`, `TestAtomicWriteStatErrorNotWidened` (`internal/settings/yamlpatch/yamlpatch_test.go`).

### REQ-SGF2-007 (Ubiquitous)

The stale `@MX:ANCHOR` contract comment on `PatchFile` (`yamlpatch.go:52-53`) shall be corrected to the measured call-site distribution (E5, E11) — it sits on the very function this SPEC changes.

### REQ-SGF2-008 (Event-driven)

When a mutant probe reverts the fix and a guard fails to catch the reversion, the actor shall record the uncaught mutant in `progress.md` as a statement of that guard's boundary — an uncaught mutant is recorded evidence, never a silently dropped run.

> **REQ 개수 상한**: Tier S의 요구 상한은 8이다(`spec-workflow.md` § SPEC Complexity Tier). 위 8건이 전부이며, 초안에 있던 아홉 번째 요구(검증 범위를 건드린 패키지로 한정하고 전체 스위트 로컬 실행을 금지)는 **삭제했다** — §6 첫 bullet과 `acceptance.md` §A가 같은 문장을 이미 담고 있어 순수 중복이었고, 상한 초과는 예산을 완화할 신호가 아니라 접거나 쪼갤 신호다. 그 제약은 요구가 아니라 검증 규율로 §6에 산다.

---

## §3 Acceptance Criteria — 소재 선언

본 SPEC은 Tier S이나(artifact set = spec.md + plan.md, AC는 통상 spec.md 인라인), 디스패치 지시에 따라 **AC의 정본은 `acceptance.md`**에 둔다. 중복 저작은 드리프트를 만들므로 여기에는 id와 한 줄 의도만 두고 Given-When-Then 본문은 두지 않는다.

| AC | 커버하는 REQ | 한 줄 의도 |
|----|-------------|-----------|
| AC-SGF2-001 | REQ-SGF2-001, REQ-SGF2-003 | greenfield 생성 출력이 block 스타일이다 — 둘 이상의 `PatchFile` 루트 키에서 |
| AC-SGF2-002 | REQ-SGF2-002 | 생성된 파일에 대한 두 번째 저장도 block을 유지한다 — 고착 해소 |
| AC-SGF2-003 | REQ-SGF2-004 | 기존 block 파일의 바이트 보존 — 주석·빈 줄·키 순서·미편집 키 불변 |
| AC-SGF2-003b | REQ-SGF2-004 | deliberate-flow 보존 — **upsert** edit + 원본 대비 **바이트 비교**. 과잉 수리를 기각하는 유일한 판별 셀 |
| AC-SGF2-004 | REQ-SGF2-006 | 통제군 6건 무수정 GREEN |
| AC-SGF2-005 | REQ-SGF2-005, REQ-SGF2-008 | 뮤턴트 2방향 채득 — 되돌림은 RED, 과잉 적용은 AC-SGF2-003b가 FAIL |
| AC-SGF2-006 | REQ-SGF2-007 | `@MX:ANCHOR` 주석이 실측 호출점 분포와 일치한다 |

**Traceability 단정**: REQ-SGF2-001..008 여덟 건 전부가 위 표의 어느 행엔가 나타난다 — 번호 붙은 AC 없이 남는 요구는 없다. 기계 확인: `spec.md`에서 `/usr/bin/grep -o 'REQ-SGF2-[0-9]\{3\}' | sort -u`의 집합이 위 "커버하는 REQ" 열의 합집합과 같아야 한다. AC 개수는 논리 AC 6건(라벨 셀 7개 — `AC-SGF2-003b`는 003의 하위 셀이다)으로 Tier S 상한 8 이내다.

---

## §4 Plan-phase 소견 — 수리 방향

**소견(결정 아님)**: greenfield 씨앗 경로를 취했다는 사실을 기록해 두고, 파싱 후 루트 노드의 flow 스타일을 해제하는 방향이 측정으로 확인됐다(E12) — 다섯 섹션이 block으로 바뀌고, block-씨앗 양성 대조는 바이트 불변이며, 세 패키지 테스트가 GREEN이었다. 검증 후 되돌렸으므로 트리에는 남아 있지 않다.

**구현 형태는 run 단계의 판단이다.** 특히 다음 두 축은 run 단계가 정한다:

- 스타일 해제를 **greenfield 경로에서만** 할 것인가(씨앗 플래그), 아니면 인코딩 직전에 무조건 할 것인가. 후자는 REQ-SGF2-004를 위반한다 — 사용자가 일부러 flow로 적은 기존 파일을 재포맷하게 된다. E12가 검증한 형태는 전자다.
- 씨앗 리터럴 자체를 block 형태(예: 빈 문자열이나 개행 한 줄)로 바꾸는 대안. 이 대안은 `lineSplice`·파서 가드의 빈-문서 처리에 닿으므로 run 단계가 그 파급을 직접 재야 한다 — 본 SPEC은 이 대안을 기각하지도 채택하지도 않는다.

---

## §5 Out of Scope

### Out of Scope — 기존 flow 스타일 파일의 소급 재포맷

- 이미 flow 스타일로 태어난 사용자 트리의 섹션 파일을 찾아 block으로 고쳐 쓰는 마이그레이션·스윕은 본 SPEC이 다루지 않는다. 본 SPEC은 **새로 만들어지는 파일의 형상**만 고친다. 소급 재포맷은 REQ-SGF2-004(바이트 보존)와 정면으로 충돌하므로, 필요하다면 별도 카드에서 명시적 사용자 동의 표면과 함께 설계해야 한다.

### Out of Scope — `PatchFile` 밖의 쓰기 경로

- typed 섹션 경로(`internal/config/manager.go` `Save()`), 프로필 스토어, 템플릿 렌더러 등 `PatchFile`을 거치지 않는 쓰기 경로의 출력 스타일은 본 SPEC의 범위가 아니다.

### Out of Scope — seam 라우팅 축

- `sectionRootKeys` 12 엔트리와 `RouteSeam` 6 섹션의 불일치(E6/E7) — 등록돼 있으나 seam으로 도달하지 않는 6개 — 는 본 SPEC이 **관측만** 한다. 등록부 정리나 라우팅 변경은 다른 결정 축이며 별도 카드 소관이다.

### Out of Scope — `sectionwrite.go`의 "8개 섹션" 스테일 수치 4곳

- 같은 스테일 수치가 `internal/settings/sectionwrite.go`의 **네 곳**에 있다: `:3`(파일 헤더), `:12`, `:14`(`@MX:REASON` 블록 안), `:57`(`WriteSectionViaSeam` 독스트링). 넷 다 실측 6(E7)과 어긋난다. 그러나 이 주석들은 본 SPEC이 고치는 함수 위에 있지 않으므로(`@MX:ANCHOR`와 달리) 범위 규율상 손대지 않는다 — 네 곳을 모두 열거하는 이유는 이 관측을 이어받을 다음 카드가 1/4만 물려받지 않게 하려는 것이다.

---

## §6 Verification & Quality Gates

- 검증 범위는 건드린 패키지로 한정한다: `go test ./internal/settings/... ./internal/cli/...`. 전체 스위트(`go test ./...`)는 로컬 금지(레인 부하 규율), 전 패키지 판정은 CI 몫이다.
- RED-first: AC-SGF2-001·002는 수리 전 트리에서 RED 채득을 먼저 하고, 채득 로그에 커맨드 + exit code + 트리 SHA를 명기한다.
- 뮤턴트 채득(AC-SGF2-005)은 부재-형상 가드의 유일한 판별 증거다 — RED-now만으로는 채택하지 않는다.
- 부재 주장의 증거 절차는 `/usr/bin/grep`이다(이 트리의 셸 `grep`은 ugrep 래퍼다). 무출력 주장은 같은 경로에 대한 양성 대조와 나란히 인용한다.
- LSP 게이트: run 단계 임계(zero errors/type-errors/lint-errors) 적용.

---

## §7 Related SPECs & 선행 SPEC 패치

- **SPEC-SEAM-GREENFIELD-001** (`status: completed`, 유지): greenfield 생성 자체를 가능하게 만든 선행 결함 수리. 그 SPEC의 `§5 Out of Scope`는 출력 스타일을 제외 항목으로 명명하지 **않았으므로**, flow 형상은 수용된 결정이 아니라 검토되지 않은 귀결이다(§1.5). 001은 **정확히 두 가지만** 덧붙여 패치한다 — (1) greenfield 출력 스타일의 소유가 SPEC-SEAM-GREENFIELD-002임을 알리는 HISTORY 한 행, (2) `related_specs`에 `SPEC-SEAM-GREENFIELD-002` 추가. `status`·본문 섹션·`progress.md`는 건드리지 않는다.
- **SPEC-WEB-WRITE-SAFETY-001** (`status: completed`): 바이트 보존 불변(REQ-WWS-005)과 뮤턴트-오버레이 방법론의 원천. `completed`이므로 열린 형제로 흡수하는 선택지는 없다 — 본 SPEC이 새 소유자다.
- **SPEC-MCP-CONSOLE-001** (`status: completed`): `mcp` seam 섹션의 도입처. E1/E3의 측정 대상 표면이 여기서 생겼다.
- **상류 증거**: `.claude/worktrees/t1043/.moai/reports/t1043/verdict.md` §C9 — 카드 t1043이 "한 줄로 고칠 수 있으나 이 카드에서 고치지 않았다"고 남긴 신규 결함.
- **본 카드 측정 기록**: `.moai/reports/t1050/plan-measurements.md` — E1(§C1)·E2(§C2)·E12(§C4)의 verbatim 출력과 §C5(`@MX:ANCHOR` 스테일 측정). §C3은 기각된 해석이며 같은 파일의 정정 블록이 지배한다(§1.3 「측정 기록 파일」).

---

## §8 계열 관측 — 씨앗 리터럴이 계약을 넘어 형상까지 전달한 사례

SPEC-SEAM-GREENFIELD-001이 기록한 계열은 **계층 간 absent 의미론의 발산**이었다(읽기는 greenfield, 쓰기는 stat 실패). 본 결함은 그 계열의 연장이 아니라 인접한 다른 형태다: **하나의 리터럴이 의도한 의미(빈 매핑)와 함께 의도하지 않은 속성(flow 스타일)을 전달했고, 그 속성이 산출물에 영구 고착됐다.**

공통점은 "greenfield 경로는 보존할 원본이 없으므로 모든 기본값이 암묵적 선택이 된다"는 점이다 — 001이 권한 모드 0644를 **명시적 선택으로 문서화**해야 했던 것과 같은 이유로, 스타일도 명시적 선택이어야 한다. 001 §4가 모드를 다뤘고 본 SPEC §4가 스타일을 다룬다. 세 번째 암묵 기본값(줄바꿈·들여쓰기 폭 등)이 남아 있는지는 후속 카드의 질문이며 본 SPEC의 범위가 아니다.
