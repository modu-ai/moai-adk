# t433 판정서 — web i18n 사전에 `graph_shortest_path` 키 2개 추가 (t413 잔여 수리)

작성: t433 레인 (`WT-web-i18n-keys`, base `ad272be20` = `origin/develop` 적색 baseline)
일시: 2026-09-02 (KST)

## 판정

**수리 완료.** `internal/web/assets/i18n.js` 4로케일에 `f.mcp.tools.graph_shortest_path.enabled.title` / `.desc` 8항목 추가.
형제 결함 전수 스윕 결과 **추가 결함 0건** — 이 키 2개가 유일한 결함이다.
적색이던 `internal/web` 패키지 전체가 GREEN.

## 1. 재현 → 수리 → 재측정 사슬 (Reproduction-First)

**RED** (수정 전, 이 워크트리 `ad272be20`):

```
--- FAIL: TestDataI18nKeysSubsetOfDictionary
    i18n_test.go:271: data-i18n key "f.mcp.tools.graph_shortest_path.enabled.title" ... absent from the dictionary (R6)
    i18n_test.go:271: data-i18n key "f.mcp.tools.graph_shortest_path.enabled.desc" ... absent from the dictionary (R6)
--- FAIL: TestI18nKeySetParity
    schema_label_test.go:106: i18n.js missing key "...title" in all 4 locales (schema field "mcp.tools.graph_shortest_path.enabled")
    schema_label_test.go:106: i18n.js missing key "...desc" in all 4 locales (schema field "mcp.tools.graph_shortest_path.enabled")
```
(리드의 로컬 재현과 동일한 2키 — 재현 일치)

**수리**: `graph_trace_calls` 형제 항목 바로 뒤(4개 로케일 블록 각각, catalog 선언 순서 유지)에 삽입:

| 로케일 | title | desc |
|---|---|---|
| en | Graph shortest path | Find the shortest call path between two symbols. |
| ko | 그래프 최단 경로 | 두 심볼 사이의 최단 호출 경로를 찾습니다. |
| ja | グラフ最短経路 | 2つのシンボル間の最短呼び出し経路を検索します。 |
| zh | 图谱最短路径 | 查找两个符号之间的最短调用路径。 |

문구는 툴 정식 설명(`internal/cli/mcp_server.go:515` "Shortest call path from one symbol to another over code-call edges")과
형제 3종(`graph_file_api`/`graph_find_code`/`graph_trace_calls`)의 로케일별 문체에 맞췄다.

**GREEN** (수정 후):
- `go test ./internal/web -run 'TestDataI18nKeysSubsetOfDictionary|TestI18nKeySetParity' -count=1` → `ok` (0.775s)
- 키 카운트: title 4 · desc 4 (4로케일 전부)
- `go test ./internal/web -count=1` (패키지 전체, 재측정 범위) → **`ok` (3.149s)**

## 2. 형제 결함 전수 스윕 [HARD] — 추가 결함 0건

세 갈래로 측정, 세 갈래 모두 일치:

1. **독립 대조** — `internal/mcp/catalog.go` 전체 28개 툴의 `f.mcp.tools.<name>.enabled.title/.desc` 등재 수를 i18n.js 에서 직접 세었다:
   27개 툴 title=4·desc=4, **`graph_shortest_path` 만 title=0·desc=0**. 유일 결함.
2. **테스트 열거** — `TestI18nKeySetParity` 는 `settings.AllFields()` 전 필드의 .title/.desc/option 키/empty-label 키를
   4-로케일 강제하며, `TestI18nKeySetSubsetOfDictionary`(R6)는 렌더된 페이지의 data-i18n 전 키를 사전 대조한다.
   `TestI18n|TestSchema|TestData` 패밀리 전체 실행에서 결함은 이 2키뿐.
3. **TUI 절반** — `internal/cli` 의 `TestI18nKeySetParity`(브릿지 절반, 별도 사전)는 **수정 전부터 통과** — 결함이 web 사전에 한정됨을 뒷받침.

## 3. 범위·관례 확인

- 변경 파일은 `internal/web/assets/i18n.js` 1개 (8행 추가). Go 코드 변경 없음 — lint/vet 대상 아님.
- `internal/web/assets` 는 바이너리 임베드 자산(`go:embed`)이며 `internal/template/templates/` 에 미러가 없다 —
  Template-First 사이클·`make build` 불필요 (최근 커밋 `80d9e7e5b` 등도 직접 편집 관례).
- 원인 SPEC: SPEC-GRAPH-REPORT-001 M1 (`31566c117`) — MCP 툴 + schema 필드는 추가했으나 web 사전 항목을 빠뜨린 t413 잔여.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- **Claim**: `graph_shortest_path` i18n 키 2개(4로케일 8항목) 추가로 `internal/web` 적색 해소. 형제 결함 없음.
- **Evidence**: §1 RED→GREEN 사슬(명령+출력), §2 삼갈래 스윕. 명령은 이 워크트리에서 재실행 가능.
- **Baseline-attribution**: base `ad272be20`(= `origin/develop` 적색 run 33568757908 의 헤드) — 리드 보고 baseline 과 동일 트리에서 재현→수리→재측정. 로컬 darwin 측정.
- **Gaps**: (1) 러너 재판정은 미확인 — develop push 는 리드 일괄 몫이며, 최종 판정은 push 가 일으키는 CI 실행(리드 판독). (2) 사전 문구의 번역 품질은 형제 문체 정합으로만 자체 검증(별도 검수 절차 없음).
- **Residual-risk**: settings schema 가 향후 툴을 추가할 때 web 사전 항목이 다시 빠질 수 있는 구조는 그대로다 — 다만 TestI18nKeySetParity 가 이 결함 형태를 CI 에서 전수 강제하므로 다음 결함은 적색 run 이 아니라 PR 단계에서 잡힌다(실제로 이번 결함도 그렇게 잡혔다).
