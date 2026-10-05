# t677 — 웹 콘솔 탭 이름 문서·콘솔 불일치 판별·수리

브랜치: `WT-console-tab-names` (기점: 로컬 develop `16f3b8a81`)
일자: 2026-09-13 · 실행: lane 세션 (리드 배차)

## Claim (주장)

1. moai-web-console.md 4로케일의 번호 목록 탭 이름 42자리 중 카드가 열거한 8자리 불일치는 **전부 문서 측 드리프트**다 — 콘솔 코드(i18n.js) 결함은 없으며, 8자리 모두 문서를 콘솔 렌더 라벨로 정렬해 수리했다.
2. t530 검증기(`docs_tab_contract_test.go` names 층)의 두 무방비 지점(현지어 접두어 미비교, 산문 줄 미검사)을 확장해 8자리가 회귀 시 적색으로 뜨게 했다.
3. 산문 축(D9~D12, 각 로케일 :165 GLM 정직 배지 문단)은 현재 4자리 모두 정확하며, 바닥(floor) 단정으로 검사 대상에 편입됐다 — 문단 삭제와 i18n 이름 변경은 즉시 적색.

## Evidence (증거)

- **RED 재현** (확장 후, 수리 전): `go test ./internal/web/ -run TestDocsTabContract` → 정확히 8 FAIL, 카드 열거와 자리까지 일치:
  - ko tab 8 `"Codex"` vs `"Codex 설정"` / tab 12 `"교차 세션"` vs `"세션 간 메시지"`
  - ja tab 8 `"Codex"` vs `"Codex 設定"` / tab 12 `"交差セッション"` vs `"セッション間メッセージ"`
  - zh tab 1 `"用户信息"` vs `"身份"` / tab 6 `"Git 与工作树"` vs `"Git·工作树"` / tab 8 `"Codex"` vs `"Codex 设置"` / tab 12 `"跨会话"` vs `"跨会话消息"`
- **GREEN** (수리 후): 동일 명령 → `--- PASS: TestDocsTabContract (literals/allowlist/names)` + 산문 축 로그 `1 anchored GLM mention(s)` ×4 페이지.
- **패키지 전체**: `go test ./internal/web/` → `ok github.com/modu-ai/moai-adk/internal/web 23.661s`
- **정적 검사**: `go vet ./internal/web/` + `gofmt -l internal/web/docs_tab_contract_test.go` → 이상 없음.

## Baseline-attribution (baseline 귀속)

모든 측정은 이 워크트리(`.claude/worktrees/t677`)의 `WT-console-tab-names` 브랜치에서, 수리 전 커밋 기점 `16f3b8a81` 상태의 문서에 대해 이번 실행으로 관측한 출력이다. 콘솔 라벨 원본은 `internal/web/assets/i18n.js`의 렌더 값과 `consoleTabs()`(`internal/web/schemaform.go`) 순서를 직접 읽어 대조했다.

## 수리 내역

| 파일 | 자리 | 변경 |
|---|---|---|
| ko | 8 | `**Codex**` → `**Codex 설정(Codex)**` |
| ko | 12 | `**교차 세션(Cross-Session)**` → `**세션 간 메시지(Cross-Session)**` |
| ja | 8 | `**Codex**` → `**Codex 設定（Codex）**` |
| ja | 12 | `**交差セッション（Cross-Session）**` → `**セッション間メッセージ（Cross-Session）**` |
| zh | 1 | `**用户信息（Identity）**` → `**身份（Identity）**` |
| zh | 6 | `**Git 与工作树（Git & Worktree）**` → `**Git·工作树（Git & Worktree）**` |
| zh | 8 | `**Codex**` → `**Codex 设置（Codex）**` |
| zh | 12 | `**跨会话（Cross-Session）**` → `**跨会话消息（Cross-Session）**` |

검증기 확장(`internal/web/docs_tab_contract_test.go`):
- **현지어 비교 축**: `extractConsoleNames`이 굵은 항목의 접두어(현지어)를 반환 → i18n 로케일 라벨과 대조 (기존에는 영어 괄호만 대조).
- **맨 ASCII 항목 규칙**: 비-README(콘솔) 페이지의 무괄호 ASCII 항목은 그 로케일의 렌더 라벨과 같아야 한다 — 맨 `Codex`가 영어 baseline으로 통과하던 구멍을 막음. LLM·MCP처럼 전 로케일 동일 라벨은 계속 통과.
- **산문 축**: 렌더링되는 GLM 탭 라벨 + 탭 명사 형태의 산문 언급을 페이지당 ≥1로 바닥 단정. en·ko는 미고정 후보(라벨도 noise 목록도 아닌 명사 인접 텍스트)를 적색으로, ja·zh는 로그 관측만.

## Gaps (미검증)

- **ja·zh 산문 축은 관측 전용**: 무공백 스크립트는 조사·내용어 경계의 기계적 분리가 불가능해, 미고정 후보 ja 8건·zh 7건을 로그로만 남기고 단정하지 않는다. ja·zh 탭 이름의 권위 있는 사실은 번호 목록 축(A)이 전수 고정한다.
- **전체 스위트 미실행**: 레인 로컬 규율(CLAUDE.local.md §4)에 따라 변경 영향 패키지(`./internal/web`)만 실행했다. 전체 판정은 develop push 후 CI 몫이다.
- **docs-site 4로케일 동시성은 편집으로 충족**했으나, hugo 빌드(sitemap·섹션 수 parity)는 이 세션에서 돌리지 않았다 — 변경이 기존 문단의 이름 토큰 8개에 국한되므로 구조 변화 없음.

## Residual-risk (잔여 위험)

- proseNoise 목록(en·ko)은 이 기점의 산문 표현을 계측한 것이라, 새 산문 문구가 추가되면 적색이 난다 — 의도된 서있는 비용(standing cost)이며, 확장은 검토 가능한 명시적 편집이다.
- docs-site가 포함된 배치이므로 develop push 시 Vercel 프리뷰/프로덕션 바인딩 반응 확인이 리드 측 필요 관찰 항목이다(CLAUDE.local.md §4.1.3).
- README 4본은 탭 이름을 영어로만 적어 이번 축의 영향 밖이다(카드 범위 밖과 일치).
