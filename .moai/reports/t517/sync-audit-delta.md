# Sync-Audit (Delta) — SPEC-WEB-WRITE-SAFETY-001 (card t517)

- 감사자: sync-auditor (독립 판정)
- 감사 시점: 2026-09-07, 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t517` (HEAD `a8042c630`)
- 델타 범위: `43f9ab202..a8042c630` (2커밋 — `146372d4b` F1 수리 + F2/F4 정정, `a8042c630` F3 정정) — 초회 감사(`.moai/reports/t517/sync-audit.md`)의 결함 델타 재감사
- 선수: 초회 감사 FAIL 사유는 F1(absent 극성)·Craft 하드 기준이었다. 본 델타 재감사는 열거된 결함의 수리를 검증하고, 수리가 **또 다른 결함을 만들지 않았는지**를 같은 독립 기준으로 본다.

---

## Evaluation Report (Delta)

SPEC: SPEC-WEB-WRITE-SAFETY-001
Overall Verdict: **FAIL** (신규 차단 결함 1건 — F1-D. 수리 방향은 옳으며 잔여 델타는 렌더 1식 + 테스트)

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 0.50 | **FAIL** (차단 F1-D) | 델타 전 타깃 스위트 GREEN — `go test -count=1 ./internal/settings/... ./internal/web/... ./internal/config/...` → `ok` ×7 (exit 0, 본 런 본 트리 `a8042c630`). F1 수리의 명시 방향(OFF 기록, ON no-op)은 신규 테스트 3건+선언 테스트로 성립. 그러나 렌더-게이트 극성 불일치로 **무접촉 저장이 `todo.enabled: false`를 기록**하는 것이 본 감사 오버레이 실험으로 기계 확인됨 — 아래 F1-D |
| Security (25%) | 1.00 | PASS | 델타는 보안 표면을 변경하지 않음 — reject 흐름·splice 검증·경로 통제 모두 기존 감사 판정 유지 |
| Craft (20%) | 0.75 | PASS (기준 미달 1건 기록) | `golangci-lint run --timeout=2m` → `0 issues.` 그러나 `gofmt -l`이 델타 테스트 파일 2건 지적(write_safety_test.go ×2 — 주석 정렬), 그리고 신규 "SeamAbsentKey" 테스트 2건이 실제로는 absent 분기를 검증하지 않음(아래 F3-D) |
| Consistency (15%) | 0.75 | PASS | F2/F3/F4 정정이 모두 동일 커맨드 재측정과 일치(본 감사 독립 재측정으로 확인 — 아표). 테스트 명칭 오기 1건(F3-D) |

가중 평균 71.3 — 차단 F1-D로 **FAIL**.

### 델타 검증 결과 — 리드 지정 4판정 포인트

**(1) F1 수리가 또 다른 것을 깨지 않았는가 — 부분 실패.**

- 수리 형태 자체는 초회 감사 처방과 일치: `FieldDef.AbsentDefault`(string, bool 전용) + `withAbsentDefault` 25필드(todo 1 + mcp 24 — grep 실측 일치) + absent 분기의 유효-기본값 판정(`sectionapply.go`: 선언 없으면 effective="false" → 기존 default-off 게이트 동작 불변).
- **default-off 키**: 미선언 필드(gate, branch_guard, worktree.*, codex 토글)는 `effective="false"`로 종전과 동일. 전 타깃 스위트 GREEN + 라운드트립 계약 갱신(`schema_sections_test.go` bool 제출값을 극성 기반으로) 확인.
- **5건 기존 수리의 양성 통제**: `TestApplySchemaEditsGitStrategyRealChangeStillRewrites` 등 전부 PASS(스위트 실측) — 무손상.
- **경곗값 "빈 문자열 제출 + 유효 기본값 true"**: bool 파서는 edits에 "true"/"false"만 넣으므로(`schemaform.go` bool 분기) 도달 불가 — 문제 없음. 파일-부재가 아닌 **파스-실패**로 ok=false인 경우: default-on + OFF 제출이 기록을 시도 → PatchFile이 파스 오류로 실패 → 500(fail-closed, 무음 손상 없음).
- **그러나 — F1-D (blocking)**: 렌더와 게이트의 극성이 갈라졌다.
  - 일반 스키마 bool 렌더: `@schemaToggleRow(f, view.SchemaValues[f.Name] == "true")` (`fieldsets.templ:250`) — 부재 키는 `""` → **OFF 렌더**.
  - MCP 도구 렌더: `@mcpToolRow(f, view.SchemaValues[f.Name] != "false")` (`fieldsets.templ:687`) — 부재 키 → **ON 렌더** (fail-open 극성 일관).
  - 게이트: AbsentDefault="true" 선언 필드는 absent를 **ON**으로 해석.
  - 결론: **mcp 24종은 렌더·게이트 일치로 완전 일관**(무접촉 저장 "1"→skip ✓, 명시 OFF→기록 ✓). **todo.enabled만 불일치** — 부재 렌더 OFF → 무접촉 저장이 `""`(→"false")를 제출 → 게이트가 "실변경"으로 기록.
  - **기계적 실험 (본 감사, 오버레이 테스트 — 트리 무변경)**: todo 블록이 없는 fixture의 workflow.yaml에 렌더가 산출하는 정확한 무접촉 폼 본문(`workflow.todo.enabled__present=1`, `workflow.todo.enabled=`)을 POST /save로 제출 → 결과:
    ```
    --- before ---
    workflow:
        project:
            continuation: card
    --- after ---
    workflow:
        project:
            continuation: card
        todo:
            enabled: false
    ```
    **사용자가 아무것도 건드리지 않은 저장이 런타임 todo 안내(SessionStart 백로그 라인·statusline TODO 세그먼트)를 무음 비활성화한다.** 이것은 이 카드의 존재 이유 그 자체인 결함 부류(AC-WWS-001/003의 형태 — 의도 없는 저장이 런타임 상태를 바꿈)를 todo 필드에서 재도입한 것이다. 반대 방향으로 명시 ON 저장은 absent+true==effective로 스킵되어 무음 무효(표시는 OFF 유지) — 저장이 사용자 의도를 삼키는 초회 F1의 형태도 방향만 바꿔 잔존한다.
  - 근거 위치: 렌더 `internal/web/fieldsets.templ:250`(vs `:687`), 게이트 `internal/settings/sectionapply.go` absent 분기, 증거는 본 보고서 위 실험 출력(오버레이 스크치는 감사 후 삭제, 재현 커맨드는 본문 기재).
  - **수리 방향 (1식)**: 일반 bool 렌더의 checked 산출에 AbsentDefault 반영 — 현재값 부재(`""`) 시 `f.AbsentDefault == "true"`면 ON 체크(mcpToolRow의 `!= "false"`와 동일 의미). 부재 렌더가 ON이 되면: 무접촉 저장 "1"→skip ✓, 명시 OFF→기록 ✓, 명시 ON(부재)→skip이되 표시도 ON이라 무해. + 전체 경로 가드 테스트(위 실험을 커밋 형태로: 부재 default-on 필드에 무접촉 폼 본문 → byte-identical 보장 — 현재 RED).

**(2) RED-first 채득 — 성립(경미한 귀속 갭 1건).** `RED-f1-absent-polarity.log`: 3 동작 테스트 FAIL(FalseStillWrites/McpFalseStillWrites/TrueIsNoOp) + 선언 테스트 PASS + 패키지 FAIL — "극성 선언만 있고 게이트 미수리" 상태와 내부적으로 정확히 일치(올바른 이유의 RED). 다만 §D.1 4요소 중 **트리 SHA가 로그/progress 어디에도 없다**(중간 상태는 커밋 불가 특성상 working-tree 측정 — progress.md가 "커밋된 수리 코드 이전 트리"로만 기술). failure 메시지 형태와 최종 GREEN(`a8042c630`)의 대칭으로 신뢰성은 확보되나, 귀속 기록에 SHA 대신 상태 기술의 명시를 권고(minor).

**(3) 뮤턴트-D 재현 — 성립.** 본 감사가 독립 재현: absent 분기 비교를 `==`→`!=`로 뒤집는 오버레이 뮤턴트 → `TestApplySchemaEditsAbsentDefaultOnBoolFalseStillWrites`·`...OnMcpFalseStillWrites`·`...OnBoolTrueIsNoOp` **3건 전부 RED** 확인 후 복원.

**(4) F2/F3/F4 기록-재측정 일치 — 전부 성립.**
- **F2** ✓: progress.md M4 (iv)행·E1 AC-WWS-006 행이 "모든 중복 reject(동의 여부 무관)"으로 정정 — 코드(`schemaform.go:325-332`)와 커밋된 테스트와 일치.
- **F3** ✓: 본 감사의 동일 커맨드 재측정 — acceptance.md `grep -oE 'AC-WWS-[0-9]+' | sort -u | wc -l` → **8**, CHANGELOG → **8**(엔트리가 8 AC를 명시 나열로 전환한 것이 반영됨), SPEC-ID 중복 `grep -c` → **1**. §E.4 정정문이 오기를 명시하고 수리 선택(나열)과 재측정 귀속을 기록 — VCI §2 충족.
- **F4** ✓: M1(d) 기제 서사 정정 — "중복 첫-값 채택" 철회(도구 구조상 중복 제출 없음 — 본 감사 실측 인용), "false 폴백" 지지, effort 소실 **미해결** 명시. 초회 감사 판정과 정확히 일치.

### Findings (delta)

- **F1-D [blocking]** `internal/web/fieldsets.templ:250` — 일반 스키마 bool 렌더가 absent를 OFF로 표시(`== "true"`)하는 반면 게이트는 AbsentDefault="true" 필드의 absent를 ON으로 해석. `workflow.todo.enabled`(템플릿이 todo 블록을 배포하지 않는 near-universal 상태)에서 **무접촉 저장이 `todo.enabled: false`를 기록해 런타임 안내를 무음 비활성화**하고, 명시 ON 저장은 무음 스킵된다. MCP 24종은 자체 렌더(`:687`)가 일치해 영향 없음. 커밋된 극성 테스트 3건은 ApplySchemaEdits 직접 호출이라 렌더→파서→게이트 루프를 통과하지 않아 이 결함을 못 잡는다. 필수 수리: 렌더 checked 산출에 absent 시 AbsentDefault 반영(1식) + 전체 경로 무접촉 저장 가드 테스트(현재 RED).
- **F3-D [minor]** `internal/settings/write_safety_test.go` `TestApplySchemaEditsSeamAbsentKeyFalseIsNoOp/TrueStillWrites` — 명칭·주석은 "key is ABSENT on disk"라 하나 `testdata/sections/gate.yaml:10-11`가 `pre_commit.enabled: false`를 **포함**해 두 테스트 모두 present-key 경로만 검증한다. 즉 default-off absent 분기(수리 (i)(b)의 원래 대상)의 직접 커버리지는 0이고, 본 감사의 MUTANT-D(absent 분기 변형)가 이 테스트들을 전혀 건드리지 않은 것도 그 때문. 수리: pre_commit 블록이 없는 gate fixture 변형 + absent 분기 2테스트(false→skip, true→write).
- **F4-D [minor]** `gofmt -l` — `internal/settings/write_safety_test.go`, `internal/web/write_safety_test.go` 포맷 이탈(주석 정렬). `gofmt -w` 1회 수리.
- **F5-D [minor]** RED-f1 로그의 §D.1 트리 SHA 결원(중간 상태 특성) — progress.md에 상태 기술 명시 권고.
- (초회 F5/F6/F7 advisory·optional 항목은 본 델타에서 변화 없음 — 유지.)

### Gaps (관측하지 못한 것)

- 실물 서버 경유 무접촉 저장은 미재현 — in-process handleSave 전체 경로 실험(동일 코드)로 대체했다.
- mcpToolRow 경로의 무접촉 저장("1" 제출 → skip)은 코드 판독에 의한 확정이지 실행 실험은 아니다(게이트 skip은 테스트 (c)가 동일 분기를 커버).
- `origin/develop` CI 전체 판정은 미관측 — 리드 push 소관.

### Residual Risk

- F1-D 수리(렌더 정렬) 전까지 todo 블록 부재 프로젝트에서 콘솔 저장 한 번이 todo 안내를 끈다 — develop 병합 전 수리 권장.
- 렌더 정렬 후에도 "콘솔 표시(부재=ON)와 파일 부재"의 관계는 여전히 추상적이다 — 사용자가 ON을 저장해도 파일에는 아무것도 생기지 않는다(런타임 동등하므로 무해하나 UI 관점 혼란 여지).

---

## 판정 요약

F1 수리의 게이트 방향은 정확하고 F2/F3/F4 기록 정정은 전부 동일 커맨드 재측정으로 입증됐다. 뮤턴트-D 재현 성립, 기존 수리 무손상, 전 스위트 GREEN. 그러나 렌더 절반이 남아 있어 todo.enabled에서 이 카드의 원래 결함 부류(무접촉 저장의 무음 런타임 변화)가 살아 있고, 본 감사의 전체 경로 실험으로 기계 확인됐다. **렌더 1식 + 전체 경로 가드 테스트 + gofmt/fixture 소정리** 후 델타 재감사(그 1식의 착지만 확인)로 PASS 가능하다.
