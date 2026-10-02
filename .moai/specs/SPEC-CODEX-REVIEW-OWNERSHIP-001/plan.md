---
id: SPEC-CODEX-REVIEW-OWNERSHIP-001
title: "구현 계획 — codex 리뷰 소유권 재배치"
version: "0.1.0"
created: 2026-10-02
---

# SPEC-CODEX-REVIEW-OWNERSHIP-001 — 구현 계획

## §A 맥락

트리 `.moai/worktrees/t1422`, 브랜치 `WT-codex-review-lane-scope`, 측정 HEAD `3ae43ed8e78ffa673ca238227df6ca7202c1ce70`. 카드 t1422(운영자 지시 2026-10-02, Class C). `quality.yaml` `development_mode: tdd`(`.moai/config/sections/quality.yaml:2`) → cycle_type=tdd.

**Tier 판정.** 개수 규칙 — 요구는 `### REQ-` 제목 수 14, 수락 기준은 `### AC-` 제목 수 16(하위 ID 없음; 둘 다 Tier M 상한 16 이하 — AC 는 상한에 닿았다: plan-audit 2회차 R1 이 AC-010 을 둘로 가르며 AC-016 을 더했다. 이후 더 늘리려면 Tier 재분류나 SPEC 분리를 먼저 판단한다). 한 제목 아래 번호 붙인 독립 단정 수는 acceptance.md §E 에 정직하게 센다. LOC 는 비테스트 ≈235 + 테스트 ≈350-420 으로 Tier M(300-1000) 안이다. 파일 수는 §H: 코드+테스트 ≈15(Tier M 대역 5-15 의 상단), 미러·문서 ≈25(기계적), 합 ≈40 이라 합계만 보면 Tier L(>15) 대역이다. 이 불일치를 조용히 넘기지 않고 §G 의 Open decision O-1 로 올린다 — 권고는 Tier M 유지(근거 §G).

마일스톤 순서는 §I, 결정 순서는 §B(**되돌리기 가장 어려운 외부 표면을 앞에 둔다**). 나머지는 기계적 귀결이다. 형제 SPEC(Claude Stop 게이트 asyncRewake 전환, 같은 카드)과의 경계는 §G.

---

## §B 되돌리기 어려운 결정 (외부 표면 먼저)

### B.1 자기 리뷰 도구 계약 — 결정 D2 (Jev 0.44 → 리더 조건부 승인, 조건 충족)

**조건.** 전용 도구는 재사용 대안과의 비교 표에서 **측정 가능한 근거로** 이겨야 한다. 이 표는 plan-audit 1회차(D11) 지적에 따라 기존 명령 `moai verify codex-review` 를 세 번째 선택지로 더했고 권한 범위 행을 추가했다.

| 기준 | (가) 전용 도구 `codex_review`/`glm_review` | (나) 감사 도구 확장 | (다) 기존 `moai verify codex-review` | 인용 |
|---|---|---|---|---|
| 감사 영수증 격리 | 영수증을 쓰지 않는다. 쓰기 호출 지점 0 | `codex_audit` 의 모든 종료 경로가 영수증을 쓴다(호출 지점 2). 비감사관 호출에서 끄려면 두 경로에 분기+플래그 | 감사 영수증이 아니라 verify 영수증을 쓴다(`codex_review_receipt.go:163` `verify.RecordReceipt`) | `mcp_codex.go:1931,1954` (`grep -c recordAuditReceipt internal/cli/mcp_codex.go` → 2); 가드 `internal/auditreceipt/store.go:596-601` |
| `required` 게이트 의미 | 적용 안 됨. codex 부재 = `inconclusive` | `applyGateUnmet` 이 두 경로에 적용 → `required` 프로젝트에서 자기 리뷰가 `fail`+`gate_unmet`. 억제 플래그 필요 | 적용 안 됨(핀 없는 게이트 드라이버) | `mcp_codex.go:1929,1952`, 정의 `:1974` |
| **권한 범위(쓰기 분류)** | 읽기 전용 힌트 true, 카탈로그 `WriteCapable: false` — codex 의 `writes` 승인 모드에서 프롬프트가 없다 | `codex_audit`·`audit_multi` 는 영수증 때문에 **쓰기 가능**으로 분류돼 있다. 자기 리뷰에 쓰면 쓰기 도구 보유가 넓어진다 | MCP 도구가 아니다. Bash 권한 규칙을 따르고 `.moai/state/verify/` 아래에 쓴다 | `internal/mcp/catalog_test.go:30-37`; `internal/mcp/catalog.go` 머리 주석 |
| 감사 도구 보유 경계 | 보유자 `{plan-auditor, sync-auditor}` 불변 | 보유자 2→≥5 | 해당 없음(Bash 보유 에이전트가 직접 실행) | `plan-auditor.md:7`, `sync-auditor.md:9` |
| 완료 SPEC 계약 접촉 | 0 — 게이트 해상기를 직접 호출 | `coerceCodexReviewTarget`/`resolveReview*` 수정. 두 체인은 "갈라지면 안 된다"(SPEC-CODEX-REVIEW-TARGET-001) | 0 | `mcp_review_material.go:111`(주석 "They must not diverge"), `:92`, `:131`; `mcp_codex.go:1218-1236` |
| GLM 다리 | 있음 | 있음 | **없음** | — |
| MCP·서브에이전트 도달 | 있음(`tools:` 부여) | 있음 | 없음 | — |
| 조언 표식·범위 메타데이터 | `advisory`, `scope`, `base`, `tree`, `truncated`, `excluded_untracked` | 없음(감사 의미) | 영수증 JSON 뿐 | `codex_review_receipt.go:200-212` |
| 스코프 재정(`uncommitted` 대 `card`) | 호출자가 고른다 | `target` 열거(`baseBranch` 는 서버 해상) | 해상기가 정한다 — `WT-` 트리에서 미커밋만 요청할 수 없다 | `codex_review_receipt.go:134` |
| 정합 비용(불리) | 도구 수 45→47, `project_root` 도구 20→22, i18n 16항목 | 0 | 0 | `internal/mcp/catalog_test.go:19`; `mcp_project_root_doc_test.go:90-95` |
| 구현 규모(불리) | 도구 부분 ≈165 LOC | ≈100 LOC | 0 LOC | §H |

**결론: (가) 전용 도구 — 리더가 세 조건(비교 표·결함 (b) 수리·결과 advisory)을 걸어 승인했고 본 절이 충족한다(Q3).** (다)가 codex 다리에서 이미 하는 일은 카드 diff 해상과 JSON 판정 출력이다. `codex_review` 가 (다) 대비 **추가하는 것**: MCP 도달(Bash 허가 없는 서브에이전트·다른 하네스), `advisory` 표식과 범위 메타데이터, 게이트 영수증 부작용 없음(Codex Stop 멤버 6 이 읽는 verify 영수증을 의도치 않게 만들지 않는다), 호출자가 고르는 `scope`, `model` 입력. **GLM 다리는 (다)에 대응물이 없다** — 이 SPEC 이 새 코드로 얻는 가장 큰 부분이다. 불리한 두 행(정합 비용·규모)은 측정 가능하고 실재하며 §H 가 3배 점검을 정직하게 보고한다. (나)로 되돌릴 때의 범위: REQ-CRO-007~010 을 "감사 도구에 `advisory` 입력과 `cardDiff` 대상"으로 다시 쓰고 REQ-CRO-011 을 감사 도구 보유 확대로 바꾼다.

**계약 (고정 대상).**

- 입력: `scope`(필수, `card`|`uncommitted`), `model`(선택), `project_root`(선택 — 기존 `projectRootOption()` 계약, `mcp_project_root.go:67,92`). `focus` 는 두지 않는다(요구 없음).
- 출력 `SelfReviewOutput`: 기존 `ReviewOutput` 필드(`verdict`/`summary`/`findings`/`next_steps`) + `advisory`(항상 true) + `scope` + `base`(카드: 재계산 merge-base SHA, 그 외 빈 문자열) + `backend`(`codex`|`glm`) + `tree`(실제 리뷰한 정규 루트) + `truncated`(GLM 자료가 200,000 바이트 상한에서 잘렸는가; codex 는 항상 false) + `excluded_untracked`(GLM 두 스코프에서 자료에 못 넣은 비추적 비런타임 경로; codex 는 빈 배열). 감사 도구의 `ReviewOutput` 스키마는 바꾸지 않는다. 출력 스키마는 `mcp.WithOutputSchema[SelfReviewOutput]()` 로 선언하고 임베드 구조체가 스키마에 평탄화되는지는 run 이 관측한다(안 되면 필드를 풀어 쓴다). **모든 반환 경로가 한 생성자를 지난다**(pass·fail·inconclusive, `card`·`uncommitted`, 비카드 조기 반환, 빈 자료, 오류) — 필드를 일부 경로에서만 채우는 구현을 구조로 막기 위해서다(AC-010 이 경로 전부에서 값으로 단정). 빈 자료(spec.md REQ-CRO-010 정의)는 두 백엔드 모두 리뷰어 호출 0 회·`inconclusive` 다.
- 주석: `ReadOnlyHint` true — 카탈로그 `WriteCapable: false` 와 일치(`mcp_annotation_guard_test.go`).
- 모델(Q6 잠정, Jev 0.22): 두 도구 모두 `workflow.audit.*` 핀을 읽지 않는다. codex 는 게이트와 같은 핀 없는 드라이버 `runCodexReviewRPC`(`mcp_codex.go:982`), GLM 은 작업 위임 경로의 기본 모델(`mcp_glm.go:63` `glmTaskDefaultModel`). `model` 입력이 있으면 그것. **되돌리는 비용은 해상기 호출 한 번이다**: codex 핸들러에서 `runCodexAuditReviewRPC`(`mcp_codex.go:990`) 를, GLM 핸들러에서 `resolveGLMAuditModelEffort(root)`(`mcp_glm.go:214`)를 부르면 핀이 적용된다 — 계약·테스트 구조는 그대로다.
- 한 호출은 동기식이다. 호스트 도구 타임아웃·오래된 서버 프로세스에 대한 우회로가 **GLM 쪽에는 없다**(§G).

### B.2 설정 키 — 결정 D1 (Jev 1.0), Q1·Q2 확정

- 이름·값: `workflow.codex.review_gate.tree_scope` = `review`(기본) | `skip` (리더 최종). 로그 행의 `scope: tree`(REQ-CGS-010)와 같은 어휘다.
- 읽는 위치: 설정 루트는 경로마다 `enabled` 와 **같은 값**이다 — Claude 훅 `reviewGateConfigRoot(projectDir)`(`codex_review_gate.go:200,220-226`), Codex 체인 `c.root`(`codex_stop_chain.go:618`). 판독 함수 `readCodexReviewGateTreeScope(root)` 와 판정 헬퍼 `(scope, configRoot) → (skip bool, 사유)` 는 **신규 파일 `internal/cli/codex_review_tree_scope.go`** 에 둔다(`readCodexReviewGateEnabled`, `mcp_codex.go:2385`, 의 판독 방식을 따르되 `mcp_codex.go` 를 건드리지 않는다). 이 파일 배치는 정적 가드를 위한 것이다 — 정책 판정의 입력은 해상 결과와 키 값뿐이라는 REQ-CRO-005 를 "이 파일이 `os.Getenv`·`os.LookupEnv`·`os.Environ`·`config.Env*` 를 참조하지 않는다"는 소스 판독 시험(AC-005 (b))으로 고정하려면 정책 코드가 환경을 읽는 다른 코드(`mcp_codex.go` 는 환경을 읽는다)와 한 파일에 섞이면 안 된다. 판독은 `workflow:` 루트 아래의 중첩 경로만 읽는다(평면형·다른 블록 아래·주석 처리된 키는 `review`). 두 경로가 이 헬퍼만 부른다.
- **핸들러 배선 결정(D15).** `HandleCodexReviewGate(input, enabled, projectDir)` 시그니처는 **바꾸지 않는다**(테스트 호출 23곳 — `grep -rn "HandleCodexReviewGate(" internal --include='*_test.go' | wc -l` → 23, 감사 보고서 실측). 스코프 클래스가 트리일 때만 핸들러 안에서 기존 `reviewGateConfigRoot(projectDir)` 로 루트를 다시 구해 주입 가능한 판독기 변수(`reviewScopeResolver` 전례)로 키를 읽는다. 따라서 `:100-102` 주석("projectDir remains the CONFIG root only")과 헤더의 "free of config I/O" 문구를 run 이 정정한다. 형제 SPEC 이 같은 핸들러의 시그니처·루트 배선을 바꿀 수 있으므로 이 SPEC 은 그 결정을 선취하지 않는다(§G).
- Go 쪽: `CodexReviewGateConfig`(`internal/config/types.go:903`)에 `TreeScope string \`yaml:"tree_scope"\``, 값 이름 상수와 기본 `review` 는 `internal/config` 단일 원천(`defaults.go:1235` 부근), 로더-판독기 일치 핀은 `review_gate_config_key_test.go` 의 `TestReviewGateReaders_AgreeWithConfigLoader` 에 한 줄.
- **배포 형태(Q1 확정): 주석 예시만.** `internal/template/templates/.moai/config/sections/workflow.yaml:143-150` 의 `codex.review_gate` 블록에 `# tree_scope: review  # review | skip` 주석만 싣는다. 구조체는 파싱하고 인벤토리(`shipped_key_inventory.yaml`)·설정 스키마(`schema_sections.go:417`)·콘솔 필드(`fieldsets.templ:726`)·i18n 은 바뀌지 않는다(AC-015 가 고정).
- 이 저장소의 추적된 `.moai/config/sections/workflow.yaml` 은 건드리지 않는다(Go 기본 `review`). 리더가 쓰는 `skip` 은 primary 체크아웃의 **추적 파일 로컬 수정본**(§A.3)에 착지 **뒤** 리더가 쓰는 운영자 행위이며 이 SPEC 과 run 은 그 파일을 편집하지 않는다.
- 셸 래퍼는 건드리지 않는다. 래퍼는 `enabled` 만 보고 `moai hook codex-review-gate` 를 부르며, `skip` 판정은 Go 핸들러 안에서 일어난다(`skip` 일 때 프로세스 1회 기동 비용은 남는다 — 허용).

### B.3 skip 의 판정 위치·조건·로그 (D4 반영)

`HandleCodexReviewGate` 는 (3) 스코프 해상·로그(`codex_review_gate.go:81-82`) 직후, 셀프게이트(:83) 앞에서 헬퍼를 부른다. 멤버 6 은 `codex_stop_chain.go:624-625` 직후, 셀프게이트(:626) 앞이다 — 상태는 `stopStatusNotApplicable`, 이유는 `tree_scope=skip`, **영수증을 읽지 않는다**(`ReceiptRead` false). **skip 조건은 "클래스 트리 **그리고** 해상된 브랜치가 `WT-` 접두를 갖지 않음"이다**(`cardScopeFromBranch(scope.Branch)` false — `scope.Branch` 는 `WT-` 접두이나 기저 불가인 경우에도 채워진다, `codex_review_scope.go:101`). `WT-` 접두 세션은 기저를 계산하지 못해 클래스가 트리로 떨어져도 skip 하지 않고 오늘처럼 전체 트리 리뷰로 간다(REQ-CRO-004). 로그는 `reviewGateScopeLogger` 와 같은 stderr JSON 행 한 줄(`gate`, `scope: tree`, `tree_scope: skip`, `basis`). `produceCodexReviewReceipt`(`codex_review_receipt.go:126`)는 정책을 읽지 않는다(명시 요청).

---

## §C 최소 설계 — 재사용 지점

새로 쓰는 것은 얇은 접착뿐이다. 새 의존성·새 판별기·새 기저 계산·새 영수증 형식은 없다.

| 필요 | 재사용 | 새로 쓰는 것 |
|---|---|---|
| 카드 판별·재계산 기저 | `reviewScopeResolver`(`codex_review_scope.go:78`) | 없음 — 도구가 같은 변수를 부른다 |
| codex 요청 | `reviewRequestParams`(:169) — 카드: `baseBranch`+SHA+카드 cwd, 미커밋: 트리 형태 | `scope: uncommitted` 는 해상기를 거치지 않고 트리 클래스 `reviewScope` 를 직접 만들어 `reviewRequestParams` 에 넘긴다 |
| codex 드라이버 | `runCodexReviewRPC`(`mcp_codex.go:982`, 게이트와 동일) | 없음 |
| GLM 호출 | `callGLMAudit`(`mcp_glm.go:302`), `glmKeyLoader`, `runReviewGit`(`mcp_review_material.go:162`) | 카드: `git diff <MergeBase> -- . <런타임 접두 exclude>`, 미커밋: `git diff HEAD -- . <exclude>`. exclude 는 `reviewGateRuntimePrefixes`(`codex_review_gate.go:36-43`)에서 파생(단일 원천). 비추적 목록 `git ls-files --others --exclude-standard` + `isRuntimeManagedPath`. 잘림 표식: 길이가 `reviewDiffMaxBytes`(`mcp_review_material.go:33`) 초과인지 `truncateDiff`(:174) 호출 전에 판정 |
| 루트 해상 | `resolveToolProjectRoot`(`mcp_project_root.go:92`), `projectRootOption()`(:67) | 없음 |
| 등록·카탈로그 | `add(...)`(`mcp_server.go`), `moaiMCPTools`(`internal/mcp/catalog.go`) | 두 도구 선언, 카탈로그 두 줄 |
| 출력 | `ReviewOutput` | `SelfReviewOutput`(내장+필드 6개) |
| 정책 | `reviewScopeResolver` 결과 + 새 판독기 | `readCodexReviewGateTreeScope`, 헬퍼 |

**[HARD] 도구 두 개와 게이트 두 경로의 스코프 판별은 하나의 변수를 지난다.** 도구가 자체 `WT-` 판별이나 자체 merge-base 계산을 갖는 순간 REQ-CRO-008 이 깨진다.

---

## §D t1404 항목별 처분 — 결정 D3 검증

카드 본문은 `moai gtd` 로 읽었다(`moai gtd 2>&1 | grep -A3 '^t1404'`, 상태 `queued`). 본문 인용은 리더 제공이며, 원 처분 보고서 `.moai/reports/t1395/gate-block-disposition*.md` 는 본 트리와 primary 경로 어디에도 없어 판독하지 못했다 — **그 파일에 의존하는 전제는 모두 미관측(Gap)**.

| # | t1404 항목 | 처분 | 근거(코드 인용) |
|---|---|---|---|
| T1 | 워크트리를 나와 primary 에 앉은 레인 세션의 턴 종료 게이트가 스테일 main 사본을 검토해 오탐 블록 | **흡수** — 단 `tree_scope: skip` 을 설정한 배포에서만 해소. 기본 `review` 는 현행 | primary cwd → 비 `WT-` 브랜치는 트리 클래스(`codex_review_scope.go:94-95`), 요청은 primary 미커밋 전체(:176-179), 셀프게이트도 primary 전체(:190). skip 이 이 세 곳 앞에서 끊는다(REQ-CRO-002). 블록 3건 자체는 미관측 |
| T2 | ③ 방향 A: 검토 트리의 develop 대비 진부함 감지 | **분리**(미구축) | **미관측 상태 표기**: `grep -c -i 'stale\|staleness' internal/cli/codex_review_gate.go internal/cli/codex_review_scope.go` → 0 / 1(1은 env 라벨 잔존을 설명하는 주석, :9 부근). 양성 대조: `grep -c -i 'detector' internal/cli/codex_review_gate.go` → 4(파일은 검색 가능). 0 에 가까운 출력은 감지기 **부재의 증명이 아니라** 이 판독에서 찾지 못했다는 관측이다. 새 메커니즘(트리 HEAD vs develop 비교+임계)이며 스코핑이 아니다 |
| T3 | ③ 방향 B: primary 체크아웃 비카드 검토 제외 | **흡수** | REQ-CRO-002. 단 "primary 인가"를 검출하지 않는다 — 해상 결과+설정으로만 정한다(REQ-CRO-005) |
| T4 | 비카드 스코프에서 런타임 관리 파일(settings 드리프트류) 제외 | **분리** — 관측하지 못한 전제 | 현행 접두 목록은 `codex_review_gate.go:36-43`(6개). "settings 드리프트류"가 어떤 경로인지 본 레인은 관측하지 못했다(원 보고서 부재). skip 배포에서는 무관 |
| T5 | 장부(`ledger.jsonl`) 조회로 기지류 사전 분류 | **분리** | **미관측 상태 표기**: `grep -c -i ledger internal/cli/codex_review_gate.go internal/cli/codex_review_scope.go` → 0 / 0, 양성 대조는 T2 와 같은 파일 검색 가능성. 소비자가 이 두 파일에서 보이지 않는다는 관측일 뿐 부재의 증명은 아니다. 새 메커니즘(발견 장부 읽기·분류)이며 스코핑이 아니다 |
| T6 | `WCI_EXCLUDES` 의 `.moai/reports/**` 제외 누락("본 카드 흡수 후보") | **분리**(D3) | 다른 게이트다: `.claude/hooks/moai/sync-phase-quality-gate.sh:257-266` 의 배열(본 트리 실측, 닫는 괄호 :266). `grep -n '\.moai/reports' .claude/hooks/moai/sync-phase-quality-gate.sh` → `:894` 주석 한 줄뿐. codex 리뷰 게이트의 접두 목록은 `.moai/reports/` 를 이미 제외한다(`codex_review_gate.go:39`) |
| T7 | `moai gpt` 문서-CLI 드리프트 | **분리**(카드 t1406) | `internal/cli/launcher.go:132-136` 이 "moai gpt is removed" 를 반환하고 `AGENTS.md:322` 표가 `moai gpt` 를 안내한다 |
| T8 | 뿌리: 검토자 CLI 가 primary(main) 소스 빌드라 develop 기능이 미구현으로 관측 | **분리** | 도구 출처(provenance) 문제 — `verification-claim-integrity.md` §2.2. 라이브 재현은 하지 않았다 |
| T9 | t1383 과 겹침 | **판정** | t1383 = SPEC-CODEX-GATE-SCOPE-001(completed). 카드 스코프는 이미 착지(`codex_review_scope.go:97-103`). t1404 의 남은 부분은 그 SPEC이 의도적으로 보존한 REQ-CGS-003 의 설정 탈출구이며 이 SPEC이 연다 |

**잔여.** 이 SPEC 착지 뒤 t1404 에는 T2·T4·T5·T8 이 남는다. **Q4 해결: 착지 뒤 리더가 t1404 를 이 잔여 항목으로 편집한다 — 착지 전에는 닫지 않는다.** 이는 sync 단계 인계 항목(리더 행위)이며 코드 변경이 아니다.

---

## §E 레인의 카드 리뷰 단계 — 구체형과 Stop 게이트 유지 판단

**구체형 (REQ-CRO-012·013).**

1. **이름·위치와 순서 목록.** 교리(`kanban-dispatch-detail.md` 새 절)에 레인 카드 마감 단계의 **순서 있는 목록**을 둔다: `[run-exit]` run 수렴·레인-로컬 검증 → `[card-review]` → `[integration]` 통합 창 진입·병합 → `[report]` 완료 보고. AC-013 이 이 목록에서 세 앵커의 순서를 구조적으로 검사한다. 레인이 카드 접수 때 `TaskCreate` 로 등록하는 단계 목록(`kanban-dispatch.md` § The lane's task list carries the card's stages)에 한 줄이 늘어난다. 팩토리 4단계 표(`workflows/factory.md` plan/run/verify/sync)는 바꾸지 않는다.
2. **호출.** 레인 오케스트레이터가 직접 또는 단계 서브에이전트(manager-docs/manager-develop)가 `codex_review`(필수 백엔드)를 `scope: card`, `project_root: <자기 `git rev-parse --show-toplevel`>` 로 부른다. GLM 은 선택(`glm_review`). 도구가 보이지 않으면(오래된 MCP 서버) codex 다리는 `moai verify codex-review --project-root <tree>` 로 대체하고 GLM 다리는 "사용 불가"로 기록한다.
3. **증거.** 결과를 `.moai/reports/<card-id>/card-review.md` 에 쓴다 — 백엔드·`base` SHA·`tree`·verdict·findings·각 finding 의 처분(수정함/이월/불채택+이유). 경로를 카드 진행 기록(`progress.md` §E.2)에 인용한다.
4. **리더 측 GAP 규칙과 연속 발화.** 리더의 완료 판독이 읽는 증거 경로 목록에 `card-review.md` 가 들어간다. 파일이 없고 진행 기록이 사유를 적지 않았다면 gap 이며 카드는 그 칼럼에 남는다(`kanban-dispatch.md` § Completion is read, never trusted). 단계가 조용히 멈추면 리더의 판독에서 부재가 보인다 — 이것이 verification-completeness §1.3 의 "멈췄을 때 무엇이 달라 보이는가"에 대한 답이다. (`.moai/reports/*` 는 gitignore 대상이라 증거는 로컬이다.)
5. **재리뷰 상한.** 수리 뒤 재리뷰는 최대 2회(run 출구 verify 게이트의 재진입 상한과 같은 수, `workflows/run/mode-orchestration.md` § Verify Exit Gate). 상한에서 미해결이면 처분을 적고 리더에게 올린다 — 레인이 상한을 넓히지 않는다.
6. **조언성.** 결과는 구속력이 없다. 카드의 PASS/FAIL 은 여전히 리더의 증거 판독과 독립 감사관(sync-auditor)이 정한다. codex 미설치·`inconclusive` 는 "리뷰 못 함"으로 적고 PASS 로 쓰지 않는다. `Stop` 훅이 아니므로 턴을 막지 않는다.
7. **리더.** 교리 문장은 **조건부**로 쓴다: "`tree_scope: skip` 을 리더의 체크아웃에 설정하면 리더 세션은 turn-end codex review gate 를 갖지 않고 같은 도구로 직접 리뷰한다" — 설정 없는 저장소에서 리더는 게이트를 받으므로 무조건 문장은 거짓이다(plan-audit 2회차 N5). L9 의 앵커 `turn-end codex review gate` 는 이 조건부 문장 안에 있다. 리더가 자기 내부 산출물을 점검하려면 같은 도구를 `scope: uncommitted` 로 직접 부른다(primary 에서는 공유 트리 전체가 대상 — 한계로 명시).

**카드 스코프 Stop 게이트를 레인에 계속 둘 것인가 — 권고: 코드는 유지, 이 저장소의 레인에서는 켜지 않는다(Q5 `do_not_commit`).**

- 근거 1(관측): spec.md §A.3 — 추적된 `workflow.yaml` 에 `review_gate` 키가 없으므로 이 저장소의 레인에서 Go 측 `enabled` 는 off 로 읽힌다(코드·파일 판독). primary 의 로컬 수정본 값은 이동하는 운영자 상태라 이 판단은 그 값에 의존하지 않는다(spec.md §A.3 의 시각 표기 관측).
- 근거 2(구조): 켜면 같은 카드 diff 가 단계와 Stop 게이트에서 두 번 리뷰된다. Stop 게이트는 diff 상태가 바뀔 때마다 다시 돈다(REQ-CGS-007).
- 근거 3: 코드를 지우면 배포 사용자의 옵트인 기능(REQ-CGS-002)이 사라진다 — 범위 밖.
- 레인 Stop 게이트를 켜고 싶은 저장소는 `enabled: true` 를 커밋하고 비카드 세션은 `tree_scope: skip` 으로 가르면 된다 — 이것이 새 키의 존재 이유다.

---

## §F 파일 목록·Template-First 미러·정합 목록 (마일스톤 배정 포함)

### F.1 Go 프로덕션 (미러 없음)

| 파일 | 변경 | 마일스톤 |
|---|---|---|
| `internal/cli/codex_review_gate.go` | skip 호출·헬퍼·주석 정정 | M2 |
| `internal/cli/codex_stop_chain.go` | 멤버 6 skip | M2 |
| **신규** `internal/cli/codex_review_tree_scope.go` | 판독기+정책 헬퍼(환경 미참조 — AC-005 (b) 정적 가드 대상) | M2 |
| `internal/config/types.go`, `defaults.go` | 필드·상수·기본 | M2 |
| `internal/cli/mcp_server.go` | 두 도구 등록 | M3 |
| `internal/mcp/catalog.go` | 두 줄 | M3 |
| **신규** `internal/cli/mcp_selfreview.go` | 핵심+핸들러+출력 타입 | M3 |

### F.2 테스트·불변식 (도구 수·목록을 세거나 열거하는 전부)

| 대상 | 변경 | 마일스톤 |
|---|---|---|
| `internal/cli/review_gate_config_key_test.go`, `codex_review_gate_test.go`(또는 신규 `codex_review_ownership_test.go`), `codex_stop_fixture_test.go` 계열 | 정책 진리표·skip·카드 무영향·env 행렬 | M1(RED)→M2 |
| 신규 `internal/cli/mcp_selfreview_test.go` | 도구 표면·카드/미커밋 요청·GLM 자료·advisory·truncated | M1(RED)→M3 |
| `internal/mcp/catalog_test.go:19` `wantCatalogSize` 45→47, `:30-37` 주석의 읽기 전용 수(25→27) | 카탈로그 불변식 | M3 |
| `internal/cli/mcp_jev_catalog_doc_test.go` — `TestMCPToolCatalogueDocsStayMirrorIdentical`(두 규칙 문서와 각 미러의 **바이트 동일**), `TestMCPToolCatalogueFiguresMatchRegistry`(총 수 문장 `(\d+) tools exposed by the self-hosted`, 머리 `Tool families \((\d+) of the (\d+) tools` = 등록 수 − `session_msg_*` 수, 카탈로그 전체의 `(\d+)-tool`·`\((\d+) tools\)`·`of the (\d+) tools` 수치) | 시험 자체는 수정하지 않는다 — **두 규칙 문서 편집이 이 시험을 초록으로 유지해야 한다**(모든 수치 45→47, 가족 머리 41→43, 사본 쌍 바이트 동일). 이 시험들이 이 문서 쌍의 유일한 기계 강제다(시험 머리 주석) | M3 |
| `internal/cli/codex_review_ownership_test.go` 안의 시험(파일을 늘리지 않는다 — 코드+테스트 ≈15 유지) | `TestTreeScopePolicy_SourceReadsNoEnvironment` — 정책 파일을 읽어 `os.Getenv`·`os.LookupEnv`·`os.Environ`·`config.Env` 부재 단정, 같은 스캔을 `codex_review_scope.go` 에 돌려 ≥1 건 발견(양성 대조: 스캐너가 눈이 먼 것이 아님, 측정 P10 = 2) | M1(RED: 파일 부재)→M2 |
| `internal/cli/mcp_console_test.go:174` `TestMoaiMCPServer_RegistrationMatchesCatalog`, `mcp_annotation_guard_test.go` | 자동(카탈로그 파생) | M3 |
| `internal/cli/mcp_project_root_doc_test.go` | `docCountWords`(:90-95)에 `Twenty-one`·`Twenty-two`, `projectRootDocSentence` 의 `(\w+)`→`([\w-]+)`, `TestDocsSiteProjectRootMatchesServer` 의 로케일 개수 문구 4건 | **M3** |
| 에이전트 `tools:` 집합 검사(신규, `internal/template` 에이전트 테스트 옆) | 보유 허용 목록 동치 | M4 |
| 웹 i18n: `internal/web/i18n_test.go:255` `TestDataI18nKeysSubsetOfDictionary`, `i18n_governance_test.go:408` `TestI18nKeyCoverageForward`, `:217` `TestI18nUntranslatedValues`, `mcp_console_test.go:32` `TestMCPConsoleRendersAllTools` | 새 도구 행의 키가 4로케일에 있어야 초록. 이 테스트들이 per-tool 항목을 강제한다는 것은 **테스트 소스 판독에 의한 추론**이며 빨간 상태는 실행하지 않았다 — run 이 M3 에서 실행해 관측한다 | **M3** |
| `make build` 선행 검사 | `agents-emit-check`, `commands-emit-check`, `tool-policy-drift-check`(`Makefile:34`) | M4 |

### F.3 Template-First 미러 (C1=로컬, C2=배포 미러, C3=기계 방출)

배포 미러(C2)는 **중립 본문**이다 — SPEC ID·REQ 토큰·카드 번호·날짜·커밋 SHA 를 쓰지 않는다. **바이트 동일이 강제되는 쌍**: `moai-mcp-tools.md` 와 `moai-mcp-tools-catalogue.md` 의 C1·C2 (`TestMCPToolCatalogueDocsStayMirrorIdentical` — 오늘 둘 다 `diff` 출력 없음, 이 개정에서 관측; 두 사본 모두 중립 본문이어야 한다). **의도된 분기가 허용되는 쌍**: `kanban-dispatch.md`(오늘 C1 에만 한 문장이 더 있다 — `diff` 한 hunk, 이 개정에서 관측)·에이전트 정의(`tools:` 줄 번호가 다르다). `kanban-dispatch-detail.md` 쌍은 오늘 바이트 동일이다(`diff` 출력 없음) — 동일 유지를 시험이 강제하지는 않지만 C1·C2 에 같은 절을 넣는다. (이전 판의 "C1 과 C2 는 바이트 동일 관계가 아니다"는 MCP 규칙 두 쌍에 대해 틀렸다 — plan-audit 2회차 N1.)

| C1 (로컬) | C2 (`internal/template/templates/…`) | C3 / 후속 | 마일스톤 |
|---|---|---|---|
| `.claude/agents/moai/manager-develop.md` `tools:`(C1 :9) | 동명 (C2 :10) | `make agents-emit` → `internal/template/templates/.codex/agents/moai/manager-develop.toml` | M4 |
| `…/manager-docs.md` `tools:`(C1 :9) | 동명 (C2 :9) | `…/manager-docs.toml` | M4 |
| `…/manager-lead.md` `tools:`(C1 :10) | 동명 (C2 :10) | `…/manager-lead.toml` | M4 |
| `.claude/rules/moai/core/moai-mcp-tools.md`(:3 "45 tools", :22 `project_root` 문장 `Twenty`→`Twenty-two`, :27 "Four of the twenty"→"Four of the twenty-two"(줄바꿈은 `REQUIRE` 앞에서 — AC-015 의 grep 토큰이 한 줄에 남게), :71 "(45 tools)") | 동명 — **바이트 동일** | — | **M3** |
| `.claude/rules/moai/core/moai-mcp-tools-catalogue.md`(:2 "45-tool" — frontmatter description, :10 "of the 45 tools", :14 "Tool catalogue (45 tools)", :64-65 감사 행 옆에 자기 리뷰 행 신설, :221 "Tool families (41 of the 45 tools" → "(43 of the 47 tools"; 계산 = 47 − `session_msg_*` 4 = 43) | 동명 — **바이트 동일** | — | **M3** |
| `docs-site/content/{ko,en,ja,zh}/guides/mcp-server.md`(도구 표 행+`project_root` 문장) | (docs-site 는 미러 없음) | — | **M3** (테스트가 강제) |
| `internal/web/assets/i18n.js` 16항목 | (임베드 자산) | — | **M3** |
| `.claude/rules/moai/workflow/kanban-dispatch.md`(단계 목록 [HARD] 한 줄+리더 문장) | 동명 | — | M4 |
| `.claude/rules/moai/workflow/kanban-dispatch-detail.md`(새 §카드 리뷰 단계, 순서 목록) | 동명 | — | M4 |
| (없음 — 키는 주석 예시) | `.moai/config/sections/workflow.yaml:143-150` `codex.review_gate` 주석 | — | M4 |
| `CHANGELOG.md`, README ×4(선택), `advanced/multi-model-audit.md` ×4(선택), `moai-ref-cross-model-audit` 스킬 한 줄(선택) | — | `internal/template/catalog.yaml` 해시 갱신 | M5 |

- `make build` 가 `gen-catalog-hashes.go --all` 로 `internal/template/catalog.yaml` 의 에이전트·스킬 해시를 갱신한다 — 갱신된 `catalog.yaml` 과 방출된 `.codex/agents/moai/*.toml` 3개를 같은 커밋에 넣는다. 에이전트 C2 를 고치고 `make agents-emit` 을 빠뜨리면 `make build` 선행 `agents-emit-check` 가 실패한다(`AGENTS.local.md` §2.0).
- 로컬 전용(미러하지 않음): `.claude/rules/local/*`, `AGENTS.local.md`. 이 SPEC은 둘 다 건드리지 않는다.
- `docs-site/.../autonomous-loops.md` 는 codex 게이트를 sibling 언급(:113)으로만 다뤄 변경 없이 둔다(`grep -n 'codex' docs-site/content/en/advanced/autonomous-loops.md` 는 audit_multi 단락뿐).

---

## §G 위험·인계 항목·해결된 결정·Open decisions·병합 순서

### 위험 (상태 표기)

1. **형제 SPEC과의 충돌(D15, N11 갱신).** 같은 카드의 형제 SPEC `SPEC-CODEX-REVIEW-ASYNC-001`(Claude Stop 게이트 asyncRewake 전환; **이미 작성돼 draft 로 존재**하며 이 SPEC 에 `depends_on`)은 같은 `HandleCodexReviewGate` 를 건드린다. **이 SPEC 이 먼저 착지**하고 형제의 결정을 선취하지 않는다 — 시그니처 불변, 트리 클래스일 때만 핸들러 안에서 루트 재해상(§B.2). 교차 확인(형제 plan.md §B.1 판독): 형제도 시그니처를 바꾸지 않고 실행기(`runCodexReviewGate`)의 출력 매핑·핸들러의 codex 조회 뒤(트리 루트·락·상태 키)에 삽입한다 — 이 SPEC 의 삽입(스코프 해상 직후, 셀프게이트 앞)과 위치가 겹치지 않는다. 형제가 착지하면 이 SPEC 이 넣은 판독 호출 한 줄을 그쪽 배선에 맞춰 옮긴다.
2. **t1399 충돌.** 카드 t1399 가 `MOAI_KANBAN*` env 계열을 삭제한다. 이 SPEC 은 env 를 읽지 않지만 기존 `reviewGateEnvContext`(`codex_review_scope.go:328-334`)가 `config.EnvMoaiFactoryWorker` 를 참조하므로 상수가 사라지면 그 파일이 컴파일되지 않는다(이 SPEC 의 변경 아님; 병합 순서 인지 사항). AC-005 가 env 상수를 열거하므로 t1399 가 상수를 지우면 그 AC 의 목록도 같이 갱신돼야 한다.
3. **t1424 충돌.** 카드 t1424 가 `manager-develop.md` 의 `tools:` 줄(C1 :9, C2 :10)에 codex·glm 위임 도구를 추가한다 — 같은 줄·C3 방출. 나중 착지 카드가 줄을 합친다.
4. **t1423.** plan-auditor/sync-auditor 와 cross-model 문서를 만진다 — 본 SPEC 선택 항목(스킬 한 줄)과 겹칠 수 있다.
5. **(a) 미관측 외부 전제.** 카드 스코프 codex 요청은 `baseBranch` 의 `branch` 필드에 merge-base **SHA** 를 넣는다(`codex_review_scope.go:169-174`). codex 가 SHA 를 받는지, `baseBranch` 대상이 미커밋·비추적 파일까지 읽는지는 관측되지 않았다 — 앱 서버 스키마는 `BaseBranchReviewTarget` 을 "Review changes between the current branch and the given base branch" 로, `UncommittedChangesReviewTarget` 을 "staged, unstaged, and untracked files" 로 기술한다(감사 보고서가 `codex app-server generate-json-schema` 로 인용; 이 레인은 스키마를 직접 생성하지 않았다). 기존 라이브 테스트는 `turn/started` 에서 잘린다(`codex_review_target_live_test.go`). 거부되면 카드 스코프 리뷰는 항상 `inconclusive` 가 되어 조용히 무력해진다. **M3 에 기록 관측 단계를 둔다:** 일회용 저장소(커밋 변경 1·추적 미커밋 1·비추적 1, 각각 구별되는 문자열)에 `{type: baseBranch, branch: <merge-base SHA>}` 를 보내 어떤 문자열이 리뷰 산문에 나오는지 읽어 `.moai/reports/t1422/live-probe/<날짜>.md` 에 기록한다. codex 가 없으면 **Gap 으로 기록**하고 진행하며 AC 의 통과로 세지 않는다.
6. **(b) 호스트 도구 타임아웃 — 우회로 없음(GLM).** `codex_review`/`glm_review` 는 동기식이고 리뷰 예산은 900s 다(`config.DefaultCodexReviewGateTimeout`). 호스트 도구 타임아웃이 먼저 끊으면 호출 결과가 사라지고 아무것도 기록되지 않으며 재시도 장치가 없다(타임아웃을 호스트가 어떻게 보고하는지는 미관측). `codex_role_audit` 이 그래서 만들어졌다(`mcp_server.go:425` 주석). CLI 거울 삭제로 **GLM 쪽에는 우회로가 없다.** codex 쪽 대체는 기존 `moai verify codex-review --project-root <tree>` (`--project-root` 플래그 실측)이다. **진행 알림(heartbeat, N9):** 감사 도구는 긴 호출 중 `notifyMCPProgress`(`internal/cli/mcp_progress.go:3-25` — `notifications/message` 와 클라이언트가 `progressToken` 을 준 경우의 `notifications/progress`)로 Claude Code 의 idle watchdog(stdio 기본 30분)을 재설정한다. **새 도구는 이 알림을 호출하지 않는다**(결정, decision-index Q13): 한 호출의 상한 900s 는 그 창보다 짧다. 호스트 도구 타임아웃은 별개 층이고 이 알림이 그것을 늘리는지는 미관측이다. 알림을 넣으려면 두 호출 지점에 한 줄씩이며 REQ 와 AC 가 하나씩 는다(REQ 14→15, AC 16→17 — 상한 16 초과이므로 Tier 재분류나 SPEC 분리를 먼저 판단해야 한다).
7. **(d) 오래된 MCP 서버 프로세스 — 우회로 없음(GLM).** 서버 프로세스가 오래된 빌드로 떠 있으면 새 도구가 `tools/list` 에 없다. 이 세션이 그 상태다: 세션 서버 안내문은 `build v3.2.0-rc.23 (commit: d194083fb)` 인데 설치된 `moai version` 은 `v3.2.0-rc.24` 다(둘 다 이 세션에서 관측). GLM 쪽은 재연결 외에 방법이 없고 codex 쪽은 위 기존 명령으로 대체한다.
8. **(c) GLM 자료.** `git diff` 기반이라 비추적 신규 파일이 빠진다(`excluded_untracked` 로 명시). 레인은 커밋 뒤에 단계를 돌리므로 보통 비어 있다.
9. **리더 제외의 지속성(D5) — 미해결.** §A.3 참조. 코드가 아니라 운영자 행위이며, primary 가 `main` 에 있고 `main` 이 릴리스 PR 로만 전진하므로 릴리스가 `workflow.yaml` 을 바꿀 때 로컬 수정(최초 측정 429 줄 차이)과 부딪힐 수 있다. 시도하지 않았다. **primary 로컬 설정의 현재 값은 이동하는 상태다** — 2026-10-02 16:39 KST 관측으로 `codex.review_gate.enabled: false`(수정 15:32)이며 이 SPEC 은 그 값에 의존하지 않는다(spec.md §A.3). `tree_scope: skip` 은 primary 가 다시 `enabled: true` 가 될 때만 필요하다.
10. **`scope: uncommitted` 의 귀속 한계.** primary 에서 부르면 공유 작업 트리 전체가 대상이다(spec.md §0 이 지적한 귀속 문제). 경로 제한 입력은 두지 않는다(범위 밖) — 한계를 도구 설명·교리에 적는다.
11. **분리된 HEAD(detached)의 카드 워크트리(N8).** 리베이스 중처럼 detached HEAD 인 카드 워크트리는 해상기가 `Branch == ""`("no card branch (unreadable or detached)", `codex_review_scope.go:88-92`)로 읽어 `WT-` 증거가 없다고 본다. `tree_scope: skip` 이면 그 세션의 턴 종료 게이트가 건너뛰어진다. spec.md 가 "읽을 수 없거나 detached" 근거를 명시하므로 설계 선택이지만 리뷰 방향이 꺼지는 경우이므로 위험으로 올린다 — 카드 리뷰 단계(레인 교리)가 `scope: card` 도구 호출로 같은 세션을 다시 해상하며, 그 경로도 detached 에서는 비카드 `inconclusive` 를 돌려준다(AC-008 (v) 변형).

### 인계 항목 (이 SPEC 의 코드 밖, progress.md「인계 항목」과 동일)

- 착지 **뒤** 리더: primary 의 `workflow.yaml`(추적 파일의 로컬 수정본)을 확인한다. 2026-10-02 16:39 KST 관측으로는 그 파일이 `codex.review_gate.enabled: false` 라 게이트가 이미 꺼져 있고 리더 제외에 `tree_scope: skip` 이 필요하지 않다 — **`enabled: true` 로 다시 켜는 시점에만** `tree_scope: skip` 을 함께 기록한다. 운영자 행위이며 이 SPEC 은 편집하지 않는다; 값은 이동하는 상태라 착지 시점에 다시 읽는다; 지속성 미해결.
- 착지 뒤 리더: t1404 를 잔여(T2·T4·T5·T8)로 편집, 착지 전에는 닫지 않음(Q4).
- 리더 발행: 감사 도구 `baseBranch` 정렬 카드 **t1426**(Q8).
- 오케스트레이터: 형제 SPEC(Claude Stop 게이트 asyncRewake) 작성 — 이 SPEC 착지 후.

### 해결된 결정 (2026-10-02)

처분 원문은 progress.md「Decision Log」, 권위 표기는 decision-index.md.

| Q | 결정 | 처분 | 신뢰도 / 승인 |
|---|---|---|---|
| Q1 | `tree_scope` 배포 형태 | `comment_example` — 템플릿에 주석 예시로만. 구조체는 파싱, 인벤토리·스키마·콘솔·i18n 불변 | 0.74 |
| Q2 | 키 이름·값 | `workflow.codex.review_gate.tree_scope` = `review`\|`skip` (리더 최종) | 1.00 |
| Q3 | D2 표면 | 전용 도구 — 리더 조건부 승인, 조건 충족(§B.1) | 승인 |
| Q4 | t1404 | 착지 뒤 리더가 잔여로 편집, 착지 전 닫지 않음 | 0.92 |
| Q5 | 추적 `workflow.yaml` 의 `enabled: true` | 커밋하지 않음. 레인 수단은 카드 리뷰 단계뿐. primary 의 `tree_scope: skip` 은 착지 뒤 리더 행위 | 0.97 |
| Q6 | 자기 리뷰 모델 해상 | `no_pins` — 감사 핀 미적용(되돌림 비용: 해상기 호출 한 번). 리더 수용 | 0.22 (< 0.5, 잠정) |
| Q7 | CLI 거울 | **삭제**(2차 Jev 배치 0.83; 1차 포함 판정 0.38 을 대체). 리더 수용 | 0.83 |
| Q8 | 감사 도구 `baseBranch` | 이 SPEC 불변. 별도 카드 **t1426** 이 소유 | 1.00 |
| S | 구조 | Claude Stop 게이트 asyncRewake 전환은 형제 SPEC | 0.93 |

### O-1 ~ O-3 처분 (2026-10-02, Jev `jev-1.13.0` + 리더 — decision-index 의 Q10=O-1, Q11=O-2, Q12=O-3; S 는 Q9)

| ID | 처분 | 신뢰도 | 상태 |
|---|---|---|---|
| O-1 | **Tier M 유지** (`keep_tier_m` 0.84 / `retier_l` 0.16) | 0.68 | 해결 |
| O-2 | **현 크기 수용 — 리더 수용(2026-10-02)으로 크기 질문은 더 이상 잠정이 아니다.** Jev 0.45(< 0.5)의 잠정 표기는 Jev 판정으로서 보존한다. 근거: 아래「3배 초과 사유」. 되돌림 경로(리더가 뒤집는 경우): 새 codex 리뷰 도구 `codex_review` 를 뺀다 — ≈165 LOC, 가장 좁은 기준선 대비 2.4배(REQ-CRO-007~011·AC 중 codex 몫 축소, `glm_review`·정책은 유지) | Jev 0.45 (< 0.5) + 리더 수용 | 해결(리더 수용) |
| O-3 | **코드로 미해결.** 리더 제외의 지속성은 명시적 운영자/리더 인계 항목으로 남는다 — 리더가 primary `workflow.yaml` 로컬 수정과 동기화 지속성 인계를 수용했고 첫 릴리스 동기화 때 확인한다(progress.md 인계 항목 1). 이 SPEC 이 만족시킨 요구가 아니다 | — | 미해결(인계, 리더 수용) |

원 권고·근거(보존):

| ID | 결정 | 권고 | 근거 |
|---|---|---|---|
| O-1 | Tier — M 유지 vs L 로 재분류(임계 0.85, design.md+research.md 추가) | **M 유지** | LOC ≈235+테스트 ≈400 은 M(300-1000) 안. 코드+테스트 파일 ≈15 는 M 대역 상단. 합계 ≈40 의 나머지 ≈25 는 도구 목록·교리·번역 미러의 기계적 편집이라 L 의 설계·조사 산출물이 새 정보를 더하지 않는다. 상한을 조용히 풀지 않고 여기 올린다 |
| O-2 | 크기 3배 점검 — 가장 좁은 기준선(정책 ≈70 LOC + 기존 `moai verify codex-review` 를 레인 수단으로)에 대해 ≈235/70 ≈ 3.4배로 3배 초과 | 수용(리더가 Q3 에서 전용 도구를 승인) — 또는 `codex_review` 만 빼고 `glm_review` 만 만든다(≈ 70+95 = 165, 2.4배) | 좁은 기준선은 GLM 다리·MCP 도달·advisory 표식·scope 재정을 못 준다. (나) 기준선에 대해서는 ≈1.4배 |
| O-3 | 리더 제외의 지속성(primary `workflow.yaml` 로컬 수정본이 릴리스 PR 병합을 견디는가) | 리더가 첫 릴리스 동기화 때 확인 | 시도하지 않음(§G 위험 9) |

---

## §H 최소 변경 규모 추정과 3배 점검

(전부 **추정** — 이 레인은 구현하지 않았다. LOC 는 비테스트 Go 기준. CLI 거울 삭제 반영.)

| 구성 | 추정 |
|---|---|
| 정책(필드·상수·판독기·헬퍼·두 경로 배선·로그·주석 정정) | ≈ 70 LOC |
| 자기 리뷰 도구(핵심·codex·GLM 자료·출력 타입·등록 2건·카탈로그) | ≈ 165 LOC |
| 합계 | **≈ 235 LOC** |
| 테스트 | ≈ 350-420 LOC |

**파일 수(코드+테스트와 미러·문서를 분리해 센다).**

| 구분 | 개수 | 내역 |
|---|---|---|
| 코드+테스트 | ≈ 15 | 프로덕션 8(신규 1: `mcp_selfreview.go`) + 테스트 ≈7(신규 ≈2) |
| 미러·문서(기계적) | ≈ 25 | 에이전트 C1·C2·C3 9, 규칙 `moai-mcp-tools`·catalogue·kanban-dispatch·detail 각 C1+C2 8, 템플릿 `workflow.yaml` 1, docs-site 가이드 4, `i18n.js` 1, `catalog.yaml` 1, CHANGELOG 1 |
| 합계 | ≈ 40 | 선택 항목(README ×4, 스킬 ×2, multi-model-audit ×4) 제외 |

**3배 초과 사유.** 가장 좁은 기준선(정책 ≈70 LOC + 기존 `moai verify codex-review`)에 대해 제안 크기 ≈235 LOC 는 ≈3.4배로 3배 트리거를 넘는다. 이 기준선은 운영자 요구 셋을 충족하지 못해 유효한 비교가 아니다: (1) **GLM 리뷰** — 기존 명령에는 GLM 다리가 없다; (2) **어느 세션에서든의 MCP 도달** — 기존 명령은 MCP 도구가 아니라 Bash 권한이 있는 에이전트만 부를 수 있고 서브에이전트·다른 하네스에서 부를 수 없다; (3) **advisory 표식** — 기존 명령은 게이트 영수증을 남기고 구속력 없음을 결과에 싣지 않으며 범위 메타데이터(`base`·`tree`·`truncated`·`excluded_untracked`)도 없다. 리더가 2026-10-02 에 이 크기를 수용했다(O-2).

**3배 점검의 기준선(두 가지를 모두 보고한다).** (i) 가장 좁은 기준선 = 정책 ≈70 + 기존 `moai verify codex-review` 를 레인 카드 리뷰 수단으로 삼는 안(새 코드 0) → ≈235/70 ≈ **3.4배, 3배 초과**. 이 기준선은 GLM 다리·MCP 도달·advisory 표식·범위 메타데이터·scope 재정을 주지 못하며 카드 요구 ③("codex·glm 에 요청")을 GLM 에 대해 충족하지 못한다. (ii) 감사 도구 확장 안 = 정책 ≈70 + ≈100 → ≈170 → ≈235/170 ≈ **1.4배**. 초과분은 §G O-2 로 올렸고 줄이는 가장 작은 방법(`codex_review` 제외)을 함께 적었다.

---

## §I 마일스톤 (우선순위 순서, 시간 추정 없음)

### M1 — 회귀선 고정과 RED 확립 (우선순위 High)

- 변경 전 트리에서 초록 관측(고정할 것): 트리 스코프 `uncommittedChanges` 요청 형태(`TestCodexReviewGate_TreeScopeRequestShapeUnchanged`), 카드 스코프 요청(`TestCodexReviewGate_CardScopeRequestIsCardDiff`), `WT-`+기저 불가 폴백의 **해상기 반쪽**(`TestCodexReviewScope_UnidentifiedFallsToTree` — 해상기만 부른다, 게이트를 부르지 않음), fail-open, 감사 도구 보유 집합, `required` 게이트의 `codex_audit` 거동(대조군), `wantCatalogSize=45`, 인벤토리 바이트, 두 카탈로그 문서 가드(`TestMCPToolCatalogueDocsStayMirrorIdentical`·`TestMCPToolCatalogueFiguresMatchRegistry`).
- **게이트 수준 보존 시험 신규 작성(N2).** `WT-` 브랜치+`develop` 참조 없음 픽스처로 `HandleCodexReviewGate` 를 호출해 리뷰어 1회 호출·요청 `{target: uncommittedChanges, cwd: <트리>}`·스코프 로그 basis "merge base unavailable" 을 단정하는 시험(예: `TestCodexReviewGate_WTBranchWithoutBaseReviewsWholeTree`)을 **skip 키 없이** 추가해 변경 전 트리에서 `--- PASS` 를 관측하고 `.moai/reports/t1422/red/` 에 보존한다. 오늘 이 동작을 게이트 수준에서 단정하는 시험은 없다(`grep -rn 'merge base unavailable' internal/cli --include='*_test.go'` → 출력 없음, exit 1 — 이 개정 관측). 이 시험이 초록으로 관측되기 전에는 AC-004 (b) 를 회귀 칸으로 세지 않는다.
- **감사 도구 스키마 스냅숏(N4).** 변경 전 서버의 `tools/list` 에서 `codex_audit`·`glm_audit`·`claude_audit`·`audit_multi` 네 도구의 입출력 스키마를 JSON 으로 받아 `.moai/reports/t1422/red/audit-tools-schema-pre.json` 에 저장하고 sha256 을 progress.md 에 적는다 — AC-007 (f) 가 이 스냅숏과 변경 뒤 스키마를 값으로 비교한다. (체크인 골든 대신 스냅숏을 택한 이유: 골든은 시험 파일과 testdata 를 새로 늘리고 도구 스키마가 의도적으로 바뀔 때 갱신 부담을 만든다. 스냅숏은 로컬 증거라 CI 에서 재현되지 않는다는 한계가 있고 DoD 선행 게이트에 올려 이 SPEC 의 run 안에서만 강제한다.)
- RED: 신규 동작 AC 를 구현 없이 먼저 추가해 `-v` 로 `=== RUN` 과 `--- FAIL` 을 관측하고 `.moai/reports/t1422/red/` 에 구현 전 트리 SHA 와 함께 보존한다. 컴파일 실패 RED 는 스텁 선언 뒤 런타임 RED 로 한 번 더 내린다.
- 산출: 회귀 초록 출력+RED 로그. 회귀가 초록이 아니면 즉시 중단·보고.

### M2 — 정책 `tree_scope` (우선순위 High)

- REQ-CRO-001~006: 신규 `codex_review_tree_scope.go` 에 판독기(평면형·다른 블록·주석 fixture 포함)·헬퍼, 두 경로 배선·로그·핸들러 주석 정정. 멤버 6 skip 은 영수증을 읽지 않는다. 명시 생산자는 정책을 읽지 않는다. 로더-판독기 일치 핀. 정책 파일 환경 미참조 정적 가드 시험(AC-005 (b)).
- M2 종료: `./internal/cli` 소관 정책 테스트 초록.

### M3 — 자기 리뷰 도구와 **파리티 강제 편집 일체** (우선순위 High)

- REQ-CRO-007~010: 핵심 함수·codex/GLM 핸들러·`SelfReviewOutput`·등록·카탈로그.
- **같은 마일스톤에서 함께 고치는 파리티 강제 편집(M3 가 초록으로 끝나기 위한 조건):** `catalog_test.go` 크기·주석, `mcp_project_root_doc_test.go`(`docCountWords`·정규식·로케일 문구), 규칙 파일 두 사본(`moai-mcp-tools.md` 의 `Twenty-two` 문장+"Four of the twenty-two" 문장 정정, catalogue 의 모든 수치와 가족 머리 43 of 47) — **각 쌍은 바이트 동일**, docs-site 가이드 4개, `i18n.js` 16항목. 같은 편집이 `TestMCPToolCatalogueDocsStayMirrorIdentical`·`TestMCPToolCatalogueFiguresMatchRegistry` 를 초록으로 유지해야 한다.
- 기록 관측: §G 위험 5 의 라이브 프로브(codex 부재 시 Gap 으로 기록).
- M3 종료 조건(소관 패키지 단위 재측정): `./internal/cli/...`, `./internal/mcp/...`, `./internal/web/...` 의 위 테스트 전부 초록. `TestProjectRootDocMatchesServer`·`TestDocsSiteProjectRootMatchesServer`·`TestMCPToolCatalogueDocsStayMirrorIdentical`·`TestMCPToolCatalogueFiguresMatchRegistry`·웹 i18n 테스트가 **M3 안에서** 초록이어야 한다. M3 안에서 두 도구 등록 직후·문서 편집 전의 빨강 이름을 한 번 기록한다(`.moai/reports/t1422/red/` — 이 가드들이 정말 빨개지는지의 관측, 감사 보고서의 운영 노트).

### M4 — 보유·교리·템플릿 (우선순위 Medium)

- REQ-CRO-011~013: C1·C2 에이전트 `tools:` 줄, 교리(스텁 한 줄 [HARD]+detail 절·순서 목록·리더 문장·GAP 규칙), C2 중립 본문 미러, 템플릿 `workflow.yaml` 주석 예시, `make agents-emit` → `make build` → 갱신 산출물(`catalog.yaml`, `.codex/agents/moai/*.toml`) 포함. REQ-CRO-014 의 템플릿·인벤토리 불변 단정.
- 교리 스텁(`kanban-dispatch.md`)은 상시 로드라 늘리는 줄 수를 최소화한다(상세는 detail 로).

### M5 — 문서·인계 (우선순위 Low, 기계적)

- CHANGELOG, 선택 README·multi-model-audit·스킬 한 줄. sync 단계(manager-docs) 소관이 일부 겹친다.
- 인계(코드 변경 아님): §G 인계 항목 4건을 완료 보고에 옮긴다.

---

## §J 자기 검증

| 항목 | 명령 |
|---|---|
| 대상 테스트 | 이름은 M1 에서 확정된다. 먼저 `go test ./internal/cli/ -list 'Ownership\|SelfReview\|CodexReviewGate\|CodexReviewScope\|ReviewGateReaders\|ProjectRootDoc\|MoaiMCPServer\|MCPToolCatalogue\|TreeScopePolicy'` 로 이름을 열거하고, 그 정확한 이름들로 `go test ./internal/cli/ -run '^(<이름1>\|<이름2>\|…)$' -count=1 -v` 를 돌린다 — 끝 앵커까지 두르고 `=== RUN` 행 수를 열거 수와 대조해 0매칭 초록을 막는다 |
| 카탈로그 | `go test ./internal/mcp/ -count=1` |
| 웹 i18n | `go test ./internal/web/ -list 'I18n\|MCPConsole\|DataI18n'` 로 열거한 정확한 이름을 같은 `^(…)$` 형태로 실행 |
| 설정 | `go test ./internal/config/ -list 'ReviewGate\|ShippedKey'` 로 열거한 정확한 이름을 같은 `^(…)$` 형태로 실행 |
| 정적 (darwin) | `go vet ./internal/cli/... ./internal/mcp/... ./internal/config/... ./internal/web/...` |
| 정적 (windows) | `GOOS=windows GOARCH=amd64 go build ./...` |
| 에이전트 방출 | `make agents-emit-check` |
| 범위 침범 | 변경 파일 통계에 `internal/auditreceipt/`, `internal/hook/audit_receipt_guard.go`, `mcp_review_material.go`, `multi_review_gate.go`, `sync-phase-quality-gate.sh` 가 없어야 한다 |
| 환경 세척 | `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …` (한 번의 복합 호출) |

전체 스위트는 로컬에서 돌리지 않는다. 소관 패키지 전체 재측정은 마일스톤 종료 시 1회(`-timeout 30m`).

## §K 안티패턴

1. verdict 값으로 AC 검증하기 — 관측 대상은 요청의 target·cwd, GLM 자료 구성, 영수증 저장소 목록, 도구 스키마, `tools:` 줄이다.
2. 도구 안에 자체 `WT-` 판별·merge-base 계산을 두기 — 단일 해상기 위반(GLM 자료 포함).
3. `scope` 에 기본값 두기 — 레인이 빠뜨리면 커밋분 없는 조용한 다른 리뷰가 된다.
4. 자기 리뷰가 영수증을 쓰거나 `required` 게이트를 적용받기 — 감사 의미 오염. 또는 advisory 표식을 inconclusive 경로에서만 채우기.
5. 리더·레인을 env/역할로 검출하기 — REQ-CRO-005 의 shall not.
6. `tree_scope` 를 라이브 키로 싣기 — 인벤토리·가드·i18n 연쇄(Q1 은 주석 예시로 확정).
7. 명시 실행 `moai verify codex-review` 에 skip 적용하기 — 요청된 리뷰를 설정이 끈다.
8. `WT-` 접두 세션을 skip 으로 삼키기 — 기저 계산 실패 때 오늘의 리뷰 방향이 꺼진다(REQ-CRO-004).
9. C3 를 손으로 고치기, C2 에 SPEC ID·카드 번호·날짜 쓰기.
10. merge-base 핀. 0매칭 초록·전체 스위트 로컬 실행.
11. codex 가 SHA 를 받는지·미커밋 파일을 읽는지 관측 없이 단정하기 — 기록 관측 전까지 Gap.

## §L 참조

- spec.md §A(측정)·§B(요구)·§E(범위 밖)·§F(제약)
- acceptance.md — AC·RED/GREEN 두 칸 규율
- decision-index.md — Q1-Q8 권위 표기와 verdict(§G 해결된 결정과 번호 일치)
- progress.md — Decision Log(Jev 판정 원문)·인계 항목
- `.claude/rules/moai/development/verification-completeness.md` §1-2 — 관측된 실패·두 칸 채택
- SPEC-CODEX-GATE-SCOPE-001 plan.md §B-§D — 해상기·요청·영수증 설계(재사용 원천)
- `.moai/reports/t1422/plan-audit.md` — plan-audit 1회차(D1-D17)
