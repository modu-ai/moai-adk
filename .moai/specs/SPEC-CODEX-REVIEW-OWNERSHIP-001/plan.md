---
id: SPEC-CODEX-REVIEW-OWNERSHIP-001
title: "구현 계획 — codex 리뷰 소유권 재배치"
version: "0.1.0"
created: 2026-10-02
---

# SPEC-CODEX-REVIEW-OWNERSHIP-001 — 구현 계획

## §A 맥락

트리 `.moai/worktrees/t1422`, 브랜치 `WT-codex-review-lane-scope`, 베이스 `develop` `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b`. 카드 t1422(운영자 지시 2026-10-02, Class C). `quality.yaml` `development_mode: tdd`(`.moai/config/sections/quality.yaml:2`) → cycle_type=tdd. Tier **M** — REQ 14건·AC 16건은 Tier M 상한(각 16)의 안쪽이지만 상한에 닿아 있다. 파일 수는 §H 추정으로 15를 넘는 쪽이라 Tier L 후보로도 읽히나, 그 대부분이 도구 목록·문서·미러 정합(기계적)이고 새 설계 표면은 REQ-CRO-001·007·008·013 네 곳이라 M 으로 판정한다. plan-audit 가 이 판정을 점검 대상으로 삼아도 좋다.

마일스톤 순서는 §I, 결정 순서는 §B(**되돌리기 가장 어려운 외부 표면을 앞에 둔다**). 나머지는 기계적 귀결이다.

---

## §B 되돌리기 어려운 결정 (외부 표면 먼저)

### B.1 자기 리뷰 도구 계약 — 결정 D2 (Jev 0.44, 임계 0.5 미만, 코디네이터 조건부 구속)

**조건.** 코디네이터 지시: 전용 도구는 "재사용(`codex_audit`/`glm_audit` 에 카드 diff 대상을 추가하고 보유를 넓힘)" 대비 비교 표에서 **측정 가능한 근거로** 이길 때만 채택한다. 근거가 약하면 재사용을 권고한다. 결함 (b)는 어느 쪽이든 수리한다.

| 기준 | (가) 전용 도구 `codex_review`/`glm_review` | (나) 감사 도구 확장 | 인용 |
|---|---|---|---|
| 감사 영수증 격리 | 영수증을 쓰지 않는다. 쓰기 호출 지점 0. 호출 전후 `.moai/state/audit-receipts/` 목록 동일(AC-CRO-012) | `codex_audit` 의 모든 종료 경로가 영수증을 쓴다(호출 지점 2곳). 비감사관 호출에서 쓰기를 끄려면 두 경로에 분기와 입력 플래그가 필요 | `internal/cli/mcp_codex.go:1931,1954` (`grep -c recordAuditReceipt internal/cli/mcp_codex.go` → 2, 본 트리); 가드 인정 조건 `internal/auditreceipt/store.go:596-601`(같은 트리·감사관 시작 이후·`codex_audit`/`audit_multi`) |
| `required` 게이트 의미 | 적용되지 않는다. codex 부재 = 언제나 `inconclusive` | `applyGateUnmet` 이 두 종료 경로에 적용되어 `required` 프로젝트의 자기 리뷰가 codex 부재 시 `fail`+`gate_unmet` 를 돌려준다. 억제 플래그가 필요 | `mcp_codex.go:1929,1952` 호출, `:1974` 정의 |
| 감사 도구 보유 경계 | 보유자 집합 `{plan-auditor, sync-auditor}` 불변. 신규 도구만 manager-develop·manager-docs·manager-lead 에 부여 | 감사 도구를 ≥3 에이전트에 넓혀 보유자 집합이 2→5. 카탈로그 소비자 칸도 바뀐다 | `.claude/agents/moai/plan-auditor.md:7`, `sync-auditor.md:9`; `.claude/rules/moai/core/moai-mcp-tools-catalogue.md:64-65` |
| 완료 SPEC 계약 접촉 | 0 — `resolveReviewBaseBranchName`·`resolveReviewMergeBase` 불변. 신규 경로가 게이트의 해상기를 직접 호출 | 결함 (b)를 감사 도구 안에서 고치면 두 해상 체인 또는 `coerceCodexReviewTarget` 을 건드린다. 두 체인은 "갈라지면 안 된다"고 SPEC-CODEX-REVIEW-TARGET-001 이 못 박았다 | `internal/cli/mcp_review_material.go:111`(주석 "They must not diverge"), `:92`, `:131`; `mcp_codex.go:1218-1236` |
| 정합 비용(불리) | 도구 수 45→47, `project_root` 선언 도구 20→22 | 두 항목 모두 불변 | `internal/mcp/catalog_test.go:19`; `internal/cli/mcp_project_root_doc_test.go:90-95`(`docCountWords` 가 `Twenty` 에서 끝남); `internal/web/assets/i18n.js` 4로케일×도구당 title·desc |
| 구현 규모(불리) | 도구 부분 ≈ 170 LOC (+CLI ≈ 70) | ≈ 100 LOC | §H 추정 |

**결론: (가) 전용 도구를 권고한다.** 근거는 앞 네 행이다 — 영수증 쓰기 지점 0 대 2, `required` 변환 적용 지점 0 대 2, 감사 도구 보유자 집합 불변 대 확대, 완료 SPEC 계약 접촉 0 대 2함수. 모두 코드·파일 판독으로 재현 가능한 수치다. 불리한 두 행(정합 비용·규모)도 측정 가능하며 실제로 있다: (가)가 (나)보다 도구 부분에서 ≈ 70 LOC, 파일에서 ≈ 11개 더 든다(§H, 추정). 다만 그 비용은 테스트가 큰 소리로 실패시키는 기계적 정합이고(`wantCatalogSize`, project_root 문서 테스트), 앞 네 행의 이점은 **조용히** 틀어질 수 있는 의미(영수증 오인용·`fail` 오보)를 구조로 막는다. 격차가 압도적이라고는 말하지 않는다 — 두 번째 행까지는 "한 개의 `advisory` 입력으로 억제" 설계가 신뢰할 만하면 (나)로도 닫히는 항목이고, 세 번째·네 번째 행이 (가)를 지탱한다. **Q3 해결(§G): (가) 확정 — 리더가 세 조건(비교 표·결함 (b) 수리·결과 advisory 표시)을 걸어 승인했고 본 절이 셋을 충족한다.** (나)의 규모·정합 비용 대비 근거는 위 표에 보존한다(되돌릴 때의 범위: REQ-CRO-007~011 을 "감사 도구에 `advisory` 입력과 `cardDiff` 대상" 으로 다시 쓰고 REQ-CRO-012 를 감사 도구 보유 확대로 바꾼다).

**계약 (고정 대상).**

- 입력: `scope`(필수, `card`|`uncommitted`), `focus`(선택, 문자열), `model`(선택), `project_root`(선택 — 기존 `projectRootOption()` 계약: 지정하면 레지스트리 검증, 사용 불가 경로는 거부, 부재면 서버 기본 루트).
- 출력 `SelfReviewOutput`: 기존 `ReviewOutput` 필드(`verdict`/`summary`/`findings`/`next_steps`) + `advisory`(항상 true) + `scope` + `base`(카드: 재계산 merge-base SHA, 그 외 빈 문자열) + `backend`(`codex`|`glm`) + `tree`(실제로 리뷰한 정규 루트) + `excluded_untracked`(GLM 카드 스코프에서 자료에 못 넣은 비추적 비런타임 경로, 비면 생략). 감사 도구의 `ReviewOutput` 스키마는 바꾸지 않는다(출력 스키마 타입을 새로 둔다).
- 주석: `ReadOnlyHint` true(트리에 쓰지 않는다) — 카탈로그 `WriteCapable: false` 와 일치(`mcp_annotation_guard_test.go` 가 양쪽 동치를 강제).
- 모델: 두 도구 모두 감사 핀(`workflow.audit.*`)을 읽지 않는다 — codex 는 게이트와 같은 핀 없는 드라이버(`runCodexReviewRPC`, `mcp_codex.go:982`), GLM 은 기본 모델. 자기 점검이 운영자의 감사 독립성 핀에 말없이 종속되지 않게 한다. 모델 = 호출자의 선택 입력 `model` 이 있으면 그것, 없으면 백엔드 기본. **Q6 은 잠정 해결**(Jev 0.22 < 0.5, 리더에 통지됨). **되돌리는 비용은 해상기 호출 한 번이다**: codex 핸들러에서 핀 없는 `runCodexReviewRPC`(`mcp_codex.go:982`) 대신 감사 핀을 읽는 `runCodexAuditReviewRPC` 를, GLM 핸들러에서 `resolveGLMAuditModelEffort(root)`(`mcp_glm.go:214`)를 호출하면 핀이 적용된다 — 도구 계약·출력·테스트 구조는 그대로다.
- 한 호출은 동기식이다. 호스트 도구 타임아웃(900s 리뷰 예산 대비)에 걸릴 위험은 §G.

### B.2 설정 키 — 결정 D1 (Jev 1.0)

- 이름·값: `workflow.codex.review_gate.tree_scope` = `review`(기본) | `skip`. 대안으로 `non_card`(역할 지향 이름)를 검토했으나 `tree_scope` 로 정한다 — 게이트 로그 행의 `scope: tree`(REQ-CGS-010)와 같은 단어라 로그에서 키를 역으로 찾을 수 있다.
- 읽는 위치: 설정 루트는 경로마다 `enabled` 와 **같은 값**이다 — Claude 훅 `reviewGateConfigRoot(projectDir)`(`codex_review_gate.go:200,220-226`), Codex 체인 `c.root`(`codex_stop_chain.go:618`). 한 판독 함수 `readCodexReviewGateTreeScope(root)` 를 `readCodexReviewGateEnabled`(`mcp_codex.go:2385`) 옆에 두고, 판정은 한 헬퍼 `(scope, configRoot) → (skip bool, 사유)` 가 하며 두 경로가 그것만 부른다(REQ-CRO-006).
- Go 쪽: `CodexReviewGateConfig`(`internal/config/types.go:903`)에 `TreeScope string \`yaml:"tree_scope"\`` 필드, 값 이름 상수와 기본 `review` 는 `internal/config` 에 단일 원천(`defaults.go:1235` 부근), 로더-판독기 일치 핀은 `review_gate_config_key_test.go` 의 `TestReviewGateReaders_AgreeWithConfigLoader` 에 한 줄.
- 배포 형태: `internal/template/templates/.moai/config/sections/workflow.yaml:143-150` 의 `codex.review_gate` 블록에 **주석 예시**(`# tree_scope: review  # review | skip`)만 싣는다. 살아 있는 키로 싣지 않는다 — 그러면 `shipped_key_inventory.yaml`(804항목)·`shipped_key_reader_test.go`(shipped 키는 `types.go` 밖 프로덕션 읽기를 가진 구조체 필드로 해상돼야 함)·콘솔 필드(`schema_sections.go:417`, `fieldsets.templ:726`)·i18n 4로케일이 한꺼번에 따라온다. **Q1 해결(§G): 주석 예시로만 배포하며 라이브 키 승격은 하지 않는다.** 구조체(`CodexReviewGateConfig.TreeScope`)는 키를 파싱한다. 인벤토리·설정 스키마·콘솔·i18n 은 바뀌지 않는다(AC-CRO-001b 가 고정). 이 저장소의 추적된 `.moai/config/sections/workflow.yaml` 은 건드리지 않는다(Go 기본 `review`); 이 저장소에서 리더가 쓰는 `skip` 은 **primary 체크아웃 로컬 사본**에 운영자가 적는 값이다(커밋 대상 아님, Q5 해결). 그 파일은 운영자 소유이며 이 SPEC 과 run 은 편집하지 않는다.
- 셸 래퍼는 건드리지 않는다. 래퍼는 `enabled` 만 보고 `moai hook codex-review-gate` 를 부르며, `skip` 판정은 Go 핸들러 안에서 일어난다(`skip` 일 때 프로세스 1회 기동 비용은 남는다 — 허용).

### B.3 skip 의 판정 위치와 로그

`HandleCodexReviewGate` 는 (3) 스코프 해상·로그(`codex_review_gate.go:81-82`) 직후, 셀프게이트(:83) 앞에서 `treeScopeSkipped` 를 부른다. 멤버 6 은 `codex_stop_chain.go:624-625` 직후, 셀프게이트(:626) 앞이다 — 상태는 `stopStatusNotApplicable`, 이유는 `tree_scope=skip`, **영수증을 읽지 않는다**(`ReceiptRead` false). 로그는 `reviewGateScopeLogger` 와 같은 stderr JSON 행 한 줄(`gate`, `scope: tree`, `tree_scope: skip`, `basis`). `produceCodexReviewReceipt`(`codex_review_receipt.go:126`)는 정책을 읽지 않는다(명시 요청).

---

## §C 최소 설계 — 재사용 지점

새로 쓰는 것은 얇은 접착뿐이다. 새 의존성·새 판별기·새 diff 측정식·새 영수증 형식은 없다.

| 필요 | 재사용 | 새로 쓰는 것 |
|---|---|---|
| 카드 판별·재계산 기저 | `reviewScopeResolver`(`codex_review_scope.go:78`) | 없음 — 도구가 같은 변수를 부른다 |
| codex 요청 | `reviewRequestParams`(:169) — 카드: `baseBranch`+SHA+카드 cwd, 미커밋: 트리 형태 | `scope: uncommitted` 는 해상기를 거치지 않고 트리 클래스 `reviewScope` 를 직접 만들어 `reviewRequestParams` 에 넘긴다(WT- 트리에서도 미커밋 요청을 보장) |
| codex 드라이버 | `runCodexReviewRPC`(`mcp_codex.go:982`, 게이트와 동일) | 없음 |
| GLM 자료·호출 | `callGLMAudit`(`mcp_glm.go:302`), `glmKeyLoader`, `truncateDiff`(`mcp_review_material.go:174`), `runReviewGit` | 카드 스코프 자료 `git diff <MergeBase>`(추적 합집합) + 비추적 비런타임 경로 목록(`ls-files --others` + `isRuntimeManagedPath`) |
| 루트 해상 | `resolveToolProjectRoot`(`mcp_project_root.go:92`), 옵션 `projectRootOption()`(:67) | 없음 |
| 등록·카탈로그 | `add(...)`/`projectRootOption()`(`mcp_server.go`), `moaiMCPTools`(`internal/mcp/catalog.go`) | 두 도구 선언, 카탈로그 두 줄 |
| 출력 | `ReviewOutput` | `SelfReviewOutput`(내장+필드 5개) |
| CLI | cobra 등록 패턴, JSON 출력 헬퍼(`verifyEmitJSON` 선례) | 명령 1개 `moai self-review <codex|glm> --scope …` — 이름은 run 이 확정(`moai verify codex-review` 와 충돌하지 않게 영수증 계열 밖에 둔다) |
| 정책 | `reviewScopeResolver` 결과 + 새 판독기 | `readCodexReviewGateTreeScope`, `treeScopeSkipped` |

**[HARD] 도구 두 개와 게이트 두 경로의 스코프 판별은 하나의 변수를 지난다.** 도구가 자체 `WT-` 판별이나 자체 merge-base 계산을 갖는 순간 REQ-CRO-008 이 깨진다.

---

## §D t1404 항목별 처분 — 결정 D3 검증

카드 본문은 `moai gtd` 로 읽었다(`moai gtd 2>&1 | grep -A3 '^t1404'`, 상태 `queued`). 본문 인용은 리더 제공이며, 원 처분 보고서 `.moai/reports/t1395/gate-block-disposition*.md` 는 본 트리와 primary 경로 어디에도 없어 판독하지 못했다 — **그 파일에 의존하는 전제는 모두 미관측(Gap)**.

| # | t1404 항목 | 처분 | 근거(코드 인용) |
|---|---|---|---|
| T1 | 워크트리를 나와 primary 에 앉은 레인 세션의 턴 종료 게이트가 스테일 main 사본을 검토해 오탐 블록 | **흡수** — 단 `tree_scope: skip` 을 설정한 배포에서만 해소. 기본 `review` 는 현행 | primary cwd → 비 `WT-` 브랜치는 트리 클래스(`codex_review_scope.go:94-95`), 요청은 primary 미커밋 전체(:176-179), 셀프게이트도 primary 전체(:190). skip 이 이 세 곳 앞에서 끊는다(REQ-CRO-002). 블록 3건 자체는 미관측 |
| T2 | ③ 방향 A: 검토 트리의 develop 대비 진부함 감지 | **분리**(미구축) | 감지기 없음 — `grep -c -i 'stale' internal/cli/codex_review_gate.go internal/cli/codex_review_scope.go` → 0 / 1(1은 env 라벨 잔존을 설명하는 주석, :9 부근, 감지기 아님). 새 메커니즘(트리 HEAD vs develop 비교+임계)이며 스코핑이 아니다 |
| T3 | ③ 방향 B: primary 체크아웃 비카드 검토 제외 | **흡수** | REQ-CRO-002. 단 "primary 인가"를 검출하지 않는다 — 스코프 클래스(트리)+설정으로만 정한다(REQ-CRO-005). primary 에 앉은 레인과 리더는 구별되지 않고, 구별하지 않는 것이 설계다 |
| T4 | 비카드 스코프에서 런타임 관리 파일(settings 드리프트류) 제외 | **분리** — 관측하지 못한 전제 | 현행 접두 목록은 `codex_review_gate.go:36-43`(6개). "settings 드리프트류"가 어떤 경로인지 본 레인은 관측하지 못했다(원 보고서 부재). skip 배포에서는 무관하고, `review` 배포의 목록 확장은 경로 증거가 생기면 한 줄 변경이다 |
| T5 | 장부(`ledger.jsonl`) 조회로 기지류 사전 분류 | **분리** | 새 메커니즘(발견 장부 읽기·분류). `grep -c -i ledger internal/cli/codex_review_gate.go internal/cli/codex_review_scope.go` → 0 / 0, 소비자가 없다 |
| T6 | `WCI_EXCLUDES` 의 `.moai/reports/**` 제외 누락("본 카드 흡수 후보") | **분리**(D3) | 다른 게이트다: `.claude/hooks/moai/sync-phase-quality-gate.sh:257-262` 의 배열. 본 트리에서 `grep -n '\.moai/reports' .claude/hooks/moai/sync-phase-quality-gate.sh` → `:894` 주석 한 줄뿐(배열에 없음). codex 리뷰 게이트의 접두 목록은 `.moai/reports/` 를 이미 제외한다(`codex_review_gate.go:39`) |
| T7 | `moai gpt` 문서-CLI 드리프트 | **분리**(카드 t1406) | `internal/cli/launcher.go:132-136` 이 "moai gpt is removed" 를 반환하고 `AGENTS.md:322` 표가 `moai gpt` 를 안내한다. 리뷰 게이트와 무관한 문서 결함 |
| T8 | 뿌리: 검토자 CLI 가 primary(main) 소스 빌드라 develop 기능이 미구현으로 관측 | **분리** | 도구 출처(provenance) 문제 — `verification-claim-integrity.md` §2.2(어느 빌드가 트리를 판정했는가). 게이트 스코핑이 아니다. 라이브 재현은 하지 않았다 |
| T9 | t1383 과 겹침 | **판정** | t1383 = SPEC-CODEX-GATE-SCOPE-001(completed). 카드 스코프는 이미 착지(`codex_review_scope.go:97-103`). t1404 의 남은 부분은 그 SPEC이 의도적으로 보존한 REQ-CGS-003 의 설정 탈출구이며 이 SPEC이 연다 |

**잔여.** 이 SPEC 착지 뒤 t1404 에는 T2·T4·T5·T8 이 남는다. **Q4 해결: 착지 뒤 리더가 t1404 를 이 잔여 항목으로 편집한다 — 착지 전에는 닫지 않는다.** 이는 sync 단계 인계 항목(리더 행위)이며 코드 변경이 아니다.

---

## §E 레인의 카드 리뷰 단계 — 구체형과 Stop 게이트 유지 판단

**구체형 (REQ-CRO-013).**

1. **이름·위치.** 카드 단계 목록에 `card-review` 를 둔다. 레인이 카드 접수 때 `TaskCreate` 로 등록하는 단계 목록(`kanban-dispatch.md` § The lane's task list carries the card's stages)에 한 줄이 늘어난다. 위치는 run 수렴·레인-로컬 검증 뒤, 통합 창 진입(병합) 앞 — 즉 sync 단계의 첫 단계다. 팩토리 4단계 표(`workflows/factory.md` plan/run/verify/sync)는 바꾸지 않는다(verify 는 run 의 출구 게이트이고 보안 리뷰다).
2. **호출.** 레인 오케스트레이터가 직접 또는 단계 서브에이전트(manager-docs/manager-develop)가 `codex_review`(필수 백엔드)를 `scope: card`, `project_root: <자기 `git rev-parse --show-toplevel`>` 로 부른다. GLM 은 선택(`glm_review`). 서브에이전트 보유는 REQ-CRO-012.
3. **증거.** 결과를 `.moai/reports/<card-id>/card-review.md` 에 쓴다 — 백엔드·`base` SHA·`tree`·verdict·findings·각 finding 의 처분(수정함/이월/불채택+이유). 경로를 카드 진행 기록(`progress.md` §E.2)에 인용한다. 리더는 완료 판독 때 이 파일을 progress.md 와 함께 읽는다(`kanban-dispatch.md` § Completion is read, never trusted) — 파일이 없고 진행 기록도 사유를 적지 않았다면 gap 이다.
4. **재리뷰 상한.** 수리 뒤 재리뷰는 최대 2회(run 출구 verify 게이트의 재진입 상한과 같은 수, `workflows/run/mode-orchestration.md` § Verify Exit Gate). 상한에서 미해결이면 처분을 적고 리더에게 올린다 — 레인이 상한을 넓히지 않는다.
5. **조언성.** 결과는 구속력이 없다. 카드의 PASS/FAIL 은 여전히 리더의 증거 판독과 독립 감사관(sync-auditor)이 정한다. codex 미설치·`inconclusive` 는 "리뷰 못 함"으로 적고 PASS 로 쓰지 않는다. `Stop` 훅이 아니므로 턴을 막지 않는다.
6. **리더.** 리더 세션에는 턴 종료 codex 리뷰 게이트를 두지 않는다. 리더가 자기 내부 산출물(큐·배차 문서·설정 편집 등)을 점검하려면 같은 도구를 `scope: uncommitted`/`card` 로 직접 부른다.

**카드 스코프 Stop 게이트를 레인에 계속 둘 것인가 — 권고: 코드는 유지, 이 저장소의 레인에서는 켜지 않는다.**

- 근거 1(관측): §A.3 — 추적된 `workflow.yaml` 에 `review_gate` 키가 없으므로 이 저장소의 레인에서 Go 측 `enabled` 는 off 로 읽힌다(코드·파일 판독). 현상 유지 = 레인은 Stop 게이트가 없고 카드 리뷰 단계가 유일한 수단이다.
- 근거 2(구조): 켜면 같은 카드 diff 가 단계와 Stop 게이트에서 두 번 리뷰된다. Stop 게이트는 diff 상태가 바뀔 때마다 다시 돈다(REQ-CGS-007 — 판정이 카드 diff 상태 키에 묶여 상태 변화 시 재검토). 단계는 카드당 한 번+상한 2회로 바뀌지 않는다.
- 근거 3: 코드를 지우면 배포 사용자의 옵트인 기능(REQ-CGS-002)이 사라진다 — 범위 밖(§E 범위 밖).
- 따라서 이 저장소의 추적 `workflow.yaml` 에 `enabled: true` 를 커밋하지 않는다(Q5 해결). 레인 Stop 게이트를 켜고 싶은 저장소는 `enabled: true` 를 커밋하고 비카드 세션은 `tree_scope: skip` 으로 가르면 된다 — 이것이 새 키의 존재 이유다.

---

## §F 파일 목록·Template-First 미러·정합 목록

### F.1 Go 프로덕션 (미러 없음)

`internal/cli/codex_review_gate.go`(skip 호출+헬퍼 배치), `internal/cli/codex_stop_chain.go`(멤버 6 skip), `internal/cli/mcp_codex.go`(판독기), `internal/config/types.go`·`defaults.go`(필드·상수·기본), `internal/cli/mcp_server.go`(두 도구 등록), `internal/mcp/catalog.go`(두 줄), **신규** `internal/cli/mcp_selfreview.go`(핵심+핸들러+출력 타입), **신규** `internal/cli/selfreview_cmd.go`(CLI).

### F.2 테스트·불변식 (도구 수·목록을 세거나 열거하는 전부)

- `internal/mcp/catalog_test.go:19` `wantCatalogSize` 45→47, `TestMoaiMCPTools_WriteCapableSet` 주석의 읽기 전용 수(25→27 — 쓰기 가능 집합 자체는 불변).
- `internal/cli/mcp_console_test.go:174` `TestMoaiMCPServer_RegistrationMatchesCatalog`(자동), `internal/cli/mcp_annotation_guard_test.go` `TestMoaiMCPServer_AnnotationsMatchCatalog`(두 도구 `ReadOnlyHint` true).
- `internal/cli/mcp_project_root_doc_test.go`: `docCountWords`(:90-95)에 `Twenty-one`·`Twenty-two` 추가, 문장 정규식 `projectRootDocSentence` 의 `(\w+)` 가 하이픈을 못 잡으므로 `([\w-]+)` 로, `TestDocsSiteProjectRootMatchesServer` 의 로케일별 개수 문구 4건(Twenty→Twenty-two, 20→22 등).
- `internal/cli/review_gate_config_key_test.go`: `tree_scope` 판독 진리표+로더 일치 핀. `internal/cli/codex_review_gate_test.go`(또는 신규 `codex_review_ownership_test.go`): skip·카드 무영향·env 행렬. `internal/cli/codex_stop_fixture_test.go` 계열: 멤버 6 skip.
- 신규 `internal/cli/mcp_selfreview_test.go`, `selfreview_cmd_test.go`, 에이전트 보유 집합 검사(`internal/template` 의 에이전트 테스트 옆 — `tools:` 집합 동치).
- `make build` 선행 검사: `agents-emit-check`, `commands-emit-check`, `tool-policy-drift-check`(`Makefile:34`) — `tool-policy.yaml` 에는 `mcp__moai`/`codex_audit` 항목이 0(`grep -c` → 0, 본 트리)이라 영향 없음.

### F.3 Template-First 미러 (C1=로컬, C2=배포 미러, C3=기계 방출)

배포 미러(C2)는 **중립 본문**이다 — SPEC ID·REQ 토큰·카드 번호·날짜·커밋 SHA 를 쓰지 않는다. C1 과 C2 는 바이트 동일 관계가 아니다(의도된 분기).

| C1 (로컬) | C2 (`internal/template/templates/…`) | C3 / 후속 |
|---|---|---|
| `.claude/agents/moai/manager-develop.md` `tools:`(:9) | 동명 파일 | `make agents-emit` → `.codex/agents/moai/manager-develop.toml` |
| `…/manager-docs.md` `tools:`(:9) | 동명 | `…/manager-docs.toml` |
| `…/manager-lead.md` `tools:`(:10) | 동명 | `…/manager-lead.toml` |
| `.claude/rules/moai/core/moai-mcp-tools.md`(:3 "45 tools", :22 `project_root` 문장, :71) | 동명 | — |
| `.claude/rules/moai/core/moai-mcp-tools-catalogue.md`(:2,:10,:14 "45", :64-65 감사 행 옆에 자기 리뷰 행 신설, :221 "41 of the 45") | 동명 | — |
| `.claude/rules/moai/workflow/kanban-dispatch.md`(단계 목록 [HARD] 한 줄+리더 문장) | 동명 | — |
| `.claude/rules/moai/workflow/kanban-dispatch-detail.md`(새 §카드 리뷰 단계) | 동명 | — |
| (없음 — 키는 주석 예시) | `.moai/config/sections/workflow.yaml:143-150` `codex.review_gate` 주석 | — |
| (선택) `.claude/skills/moai-ref-cross-model-audit/SKILL.md` — "자기 리뷰는 감사가 아니다" 한 줄 | 동명 | catalog.yaml 해시 갱신 |

- `make build` 가 `gen-catalog-hashes.go --all` 로 `internal/template/catalog.yaml` 의 에이전트·스킬 해시를 갱신한다 — 갱신된 `catalog.yaml` 과 방출된 `.codex/agents/moai/*.toml` 3개를 같은 커밋에 넣는다. 에이전트 C2 를 고치고 `make agents-emit` 을 빠뜨리면 `make build` 선행 `agents-emit-check` 가 실패한다(`AGENTS.local.md` §2.0).
- 로컬 전용(미러하지 않음): `.claude/rules/local/*`, `AGENTS.local.md`. 이 SPEC은 둘 다 건드리지 않는다.

### F.4 문서·i18n (sync 단계 소관)

- docs-site 4로케일(ko/en/ja/zh) 동시: `guides/mcp-server.md`(도구 표 행 + `project_root` 문장 — 위 테스트가 강제). 선택: `advanced/multi-model-audit.md` 에 "자기 리뷰 vs 감사" 단락. `advanced/autonomous-loops.md` 는 codex 게이트를 sibling 언급(:113)으로만 다뤄 변경 없이 둔다 — 카드가 가리킨 이 문서에는 codex 게이트 본문이 없다(`grep -n 'codex' docs-site/content/en/advanced/autonomous-loops.md` 는 audit_multi 단락뿐).
- `internal/web/assets/i18n.js`: `f.mcp.tools.codex_review.enabled.title/.desc`·`glm_review` × 4로케일 = 16항목(설정 스키마가 카탈로그에서 도구별 필드를 파생한다 — `internal/mcp/catalog.go` 머리 주석).
- README 4개 파일의 크로스모델 감사 행(README.md:360)은 선택: 자기 리뷰 행 신설.
- CHANGELOG `[Unreleased]` 항목(3-phase close 서술) — manager-docs.

---

## §G 위험·해결된 결정·병합 순서

**병합·충돌 위험.** (1) 카드 t1424 가 `manager-develop.md:9` 의 `tools:` 줄에 codex·glm 위임 도구를 추가한다 — 같은 줄·같은 C2 미러·같은 C3 방출. 병합 순서에 따라 충돌하므로 나중 착지 카드가 줄을 합친다. (2) 카드 t1399 가 `MOAI_KANBAN*` env 계열을 삭제한다 — 이 SPEC은 env 를 읽지 않지만 기존 `reviewGateEnvContext`(`codex_review_scope.go:328-334`)가 `config.EnvMoaiFactoryWorker` 를 참조하므로 상수가 사라지면 그 파일이 깨진다(이 SPEC의 변경 아님, 순서 인지 사항). (3) 카드 t1423 은 plan-auditor/sync-auditor 와 cross-model 문서를 만진다 — 본 SPEC의 선택 항목(cross-model 스킬 한 줄)과 겹칠 수 있다.

**잔여 위험.** (a) 카드 스코프 codex 요청은 `baseBranch` 의 `branch` 필드에 merge-base **SHA** 를 넣는다(`codex_review_scope.go:169-174`). 이름이 아니라 SHA 를 codex 가 받는지는 이 레인이 라이브로 확인하지 않았다 — 기존 라이브 테스트(`codex_review_target_live_test.go:136`)는 이름 없는 `baseBranch` 문자열 대상이다. SPEC-CODEX-GATE-SCOPE-001 은 거부 시 fail-open 이라고 규정했으나(`codex_review_scope.go:164-168` 주석) 거부되면 **카드 스코프 리뷰가 항상 `inconclusive`** 가 되어 조용히 무력해진다. M3 에 라이브 프로브(codex 부재 시 skip)를 한 건 둔다. (b) MCP 호출은 동기식이고 리뷰 예산은 900s 다 — 호스트 도구 타임아웃이 먼저 끊을 수 있다. `codex_role_audit` 가 그래서 만들어졌다(`mcp_server.go:425` 주석). 레인은 서브에이전트에서 호출하거나 CLI 거울을 쓴다. 백그라운드 잡 모델은 범위 밖. (c) GLM 자료는 `git diff <MergeBase>` 라 비추적 신규 파일이 빠진다(`excluded_untracked` 로 명시). 레인은 커밋 뒤에 단계를 돌리므로 보통 비어 있다. (d) MCP 서버 프로세스는 오래된 빌드로 떠 있으면 새 도구가 `tools/list` 에 없다(서버 안내문) — CLI 거울이 우회로다.

**전제 점검(밀어붙임).** 리더가 당장 겪는 문제(리더 턴마다 9~12분 리뷰)는 primary 로컬 `workflow.yaml` 에서 `enabled: false` 로 지금도 멈출 수 있다 — 코드 변경 없이. `tree_scope` 가 더하는 가치는 `enabled: true` 가 저장소 전체에 걸린 상태에서 카드 세션만 리뷰하고 비카드 세션은 건너뛰는 **스코프 분리**다. D1 은 확정이므로 따르되, 가치 범위가 이렇다는 점을 기록한다(관측: §A.3 의 설정 루트 판독, 라이브 실행은 하지 않았다).

### 해결된 결정 (2026-10-02, 운영자 위임 Jev `jev-1.13.0` + 리더 승인)

Q1-Q8 은 전부 해결됐다. **Q6·Q7 은 신뢰도 0.5 미만이라 잠정**이며 리더에 통지됐다. 처분 원문은 progress.md「Decision Log」, 권위 표기는 decision-index.md.

| Q | 결정 | 처분 | 신뢰도 | 후속 영향 |
|---|---|---|---|---|
| Q1 | `tree_scope` 배포 형태 | `comment_example` — 템플릿에 주석 예시로만. 구조체는 파싱, 인벤토리·스키마·콘솔·i18n 불변 | 0.74 | AC-CRO-001b (템플릿 라이브 키·인벤토리 불변) |
| Q2 | 키 이름·값 | `workflow.codex.review_gate.tree_scope` = `review`\|`skip` (최종) | 1.00 | — |
| Q3 | D2 표면 | 전용 도구 — 리더가 조건부 승인(비교 표·결함 (b) 수리·advisory), 조건 충족(§B.1) | 승인 조건 충족 (Jev D2 0.44 는 승인 이전 기록) | — |
| Q4 | t1404 | `edit_to_residual` — 착지 뒤 리더가 잔여(T2·T4·T5·T8)로 편집, 착지 전 닫지 않음 | 0.92 | sync 단계 인계 항목(코드 변경 아님) |
| Q5 | 추적 `workflow.yaml` 의 `enabled: true` | `do_not_commit` — 커밋하지 않음. 리더가 primary 의 비추적 로컬 설정에 `tree_scope: skip` 을 적는다(운영자 소유 파일, 이 SPEC 은 편집하지 않음). 레인 수단은 카드 리뷰 단계뿐 | 0.97 | — |
| Q6 | 자기 리뷰 모델 해상 | `no_pins` — 감사 핀 미적용, 모델 = 선택 입력 `model` 또는 백엔드 기본. **잠정** | 0.22 (< 0.5) | 되돌림 비용: 해상기 호출 한 번 (§B.1 모델 항목) |
| Q7 | CLI 거울 | 포함하되 **가장 낮은 우선순위·폐기 가능 마일스톤** (M4). **잠정** | 0.38 (< 0.5) | 폐기 규칙: 아래 |
| Q8 | 감사 도구 `baseBranch` 정렬 | `split_followup` — 감사 도구는 이 SPEC 에서 불변. 후속 카드 필요(카드 발행은 리더 소관, 레인은 발행하지 않음). 결함 (b)는 자기 리뷰 한정으로 닫힘 | 1.00 | spec.md §E 범위 밖에 명시 |

**Q7 폐기 규칙.** M4 는 다른 모든 마일스톤이 초록이 된 뒤에만 착수한다. 폐기하면 REQ-CRO-011·AC-CRO-013·M4 를 함께 삭제하고(REQ 14→13, AC 16→15), §H 합계가 ≈ 70 LOC 줄며, MCP 서버 노후·MCP 미보유 자리의 우회로가 사라진다. 폐기는 코드 계약(MCP 도구)·다른 REQ 에 영향이 없다. 폐기 판단은 리더가 한다.

**전제 점검 잔여.** 결정이 모두 해결됐어도 plan-audit 가 점검할 항목은 남는다: Tier 판정(§A), (a) SHA 를 `baseBranch` 가 받는지(미관측), Q6·Q7 의 잠정 상태.

---

## §H 최소 변경 규모 추정과 3배 점검

(전부 **추정** — 이 레인은 구현하지 않았다. LOC 는 비테스트 Go 기준.)

| 구성 | 추정 |
|---|---|
| 정책(필드·상수·판독기·헬퍼·두 경로 배선·로그) | ≈ 70 LOC |
| 자기 리뷰 도구(핵심·codex·GLM·출력 타입·등록 2건·카탈로그) | ≈ 170 LOC |
| CLI 거울 | ≈ 70 LOC |
| 합계 | **≈ 310 LOC** (CLI 제외 ≈ 240) |
| 테스트 | ≈ 350-420 LOC |
| 파일 | Go 프로덕션 9(신규 2)+테스트 ≈ 7+미러·교리 ≈ 18+문서·i18n·CHANGELOG ≈ 7 ⇒ **약 41개**(재사용 대안 약 30) |

"가장 적게 쓸 수 있는 형태"(정책+재사용 대안, CLI 없음) ≈ 170 LOC. 제안 합계 310 은 그 **≈ 1.8배**(CLI 제외 ≈ 1.4배)로 3배 임계 안이다. 크기는 도구 두 개와 정합 파일 수가 만든다 — 위 임계는 LOC 기준이고 파일 수 기준은 §B.1 비용 행이 이미 공개했다.

---

## §I 마일스톤 (우선순위 순서, 시간 추정 없음)

### M1 — 회귀선 고정과 RED 확립 (우선순위 High)

- 변경 전 트리에서 초록 관측(고정할 것): 트리 스코프 `uncommittedChanges` 요청 형태(`TestCodexReviewGate_TreeScopeRequestShapeUnchanged`), 카드 스코프 요청, fail-open, 감사 도구 보유 집합(`codex_audit`/`glm_audit` 는 plan-auditor·sync-auditor 만), `required` 게이트의 `codex_audit` 거동(대조군: codex 부재+`required` ⇒ `fail`+`gate_unmet`), `wantCatalogSize=45`.
- RED: AC-CRO-001~016 중 신규 동작을 구현 없이 먼저 추가해 `-v` 로 `=== RUN` 과 `--- FAIL` 을 관측하고 `.moai/reports/t1422/red/` 에 구현 전 트리 SHA(`git rev-parse HEAD` 출력)와 함께 보존한다. 컴파일 실패 RED 는 스텁 선언 뒤 런타임 RED 로 한 번 더 내린다(형제 SPEC 의 RED-1/RED-2 패턴).
- 산출: 회귀 초록 출력+RED 로그. 회귀가 초록이 아니면 즉시 중단·보고.

### M2 — 정책 `tree_scope` (우선순위 High)

- REQ-CRO-001~006: 판독기·헬퍼·두 경로 배선·로그. 멤버 6 skip 은 영수증을 읽지 않는다. 명시 생산자는 정책을 읽지 않는다.
- 로더-판독기 일치 핀, env 행렬(서로 다른 임의 값 2종)·카드 무영향.

### M3 — 자기 리뷰 도구 (우선순위 High)

- REQ-CRO-007~010: 핵심 함수(해상기 호출·요청 조립·GLM 자료)·codex/GLM 핸들러·`SelfReviewOutput`·등록·카탈로그. 카탈로그 크기·주석 가드·project_root 문서 테스트를 같은 마일스톤에서 갱신해 초록으로 맞춘다.
- 라이브 프로브 1건(codex 가 `baseBranch`+SHA 를 받는지; codex 부재 시 skip 이며 skip 은 미관측으로 기록).

### M4 — CLI 거울 (우선순위 Low — 가장 낮음, 폐기 가능, 잠정)

- REQ-CRO-011. 핵심 함수를 그대로 부른다. 종료 코드 규칙 검사. 번호와 무관하게 **M1·M2·M3·M5·M6 이 초록이 된 뒤 마지막에 착수**한다. 폐기 규칙은 §G(Q7).

### M5 — 보유·교리·미러 (우선순위 Medium)

- REQ-CRO-012·013: C1·C2 에이전트 `tools:` 줄, 교리(스텁 한 줄 [HARD]+detail 절), C2 중립 본문 미러, `make agents-emit` → `make build` → 갱신 산출물(`catalog.yaml`, `.codex/agents/moai/*.toml`) 포함.
- 교리 스텁(`kanban-dispatch.md`)은 상시 로드라 늘리는 줄 수를 최소화한다(상세는 detail 로).

### M6 — 정합 문서 (우선순위 Low, 기계적)

- REQ-CRO-014 잔여: docs-site 4로케일, `i18n.js`, README(선택), CHANGELOG. sync 단계(manager-docs) 소관이 일부 겹친다.
- sync 인계(코드 변경 아님): 착지 뒤 리더가 t1404 를 잔여 항목(T2·T4·T5·T8)으로 편집(Q4). 감사 도구 `baseBranch` 정렬 후속 카드는 리더가 발행한다(Q8) — 이 레인·run 은 카드를 발행하지 않는다.

---

## §J 자기 검증

| 항목 | 명령 |
|---|---|
| 대상 테스트 | 테스트 이름은 M1 에서 확정된다. 먼저 `go test ./internal/cli/ -list 'Ownership\|SelfReview\|CodexReviewGate\|CodexReviewScope\|ReviewGateReaders\|ProjectRootDoc\|MoaiMCPServer'` 로 이름을 열거하고, 그 정확한 이름들로 `go test ./internal/cli/ -run '^(<이름1>\|<이름2>\|…)$' -count=1 -v` 를 돌린다 — 끝 앵커까지 두르고 `=== RUN` 행 수를 열거 수와 대조해 0매칭 초록을 막는다 |
| 카탈로그 | `go test ./internal/mcp/ -count=1` |
| 설정 | `go test ./internal/config/ -list 'ReviewGate\|ShippedKey'` 로 열거한 정확한 이름을 같은 `^(…)$` 형태로 실행 |
| 정적 (darwin) | `go vet ./internal/cli/... ./internal/mcp/... ./internal/config/...` |
| 정적 (windows) | `GOOS=windows GOARCH=amd64 go build ./...` |
| 에이전트 방출 | `make agents-emit-check` |
| 범위 침범 | `git diff --stat` — `internal/auditreceipt/`, `internal/hook/audit_receipt_guard.go`, `mcp_review_material.go`, `multi_review_gate.go`, `sync-phase-quality-gate.sh` 가 목록에 없어야 한다 |
| 환경 세척 | `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …` (한 번의 복합 호출) |

전체 스위트(`go test ./...`)는 로컬에서 돌리지 않는다. 소관 패키지 전체 재측정은 마일스톤 종료 시 1회(`-timeout 30m`, 형제 SPEC 의 소관 패키지 단위 규율).

## §K 안티패턴

1. verdict 값으로 AC 검증하기 — 스텁 verdict 는 요청과 무관하다. 관측 대상은 요청의 target·cwd, 영수증 저장소 목록, 도구 스키마, `tools:` 줄이다.
2. 도구 안에 자체 `WT-` 판별·merge-base 계산을 두기 — 단일 해상기 위반.
3. `scope` 에 기본값 두기 — 레인이 빠뜨리면 커밋분 없는 조용한 다른 리뷰가 된다.
4. 자기 리뷰가 영수증을 쓰거나 `required` 게이트를 적용받기 — 감사 의미 오염.
5. 리더·레인을 env/역할로 검출하기 — REQ-CRO-005 의 shall not.
6. `tree_scope` 를 라이브 키로 싣기 — 인벤토리·가드·i18n 연쇄(Q1 은 주석 예시로 확정).
7. 명시 실행 `moai verify codex-review` 에 skip 적용하기 — 요청된 리뷰를 설정이 끈다.
8. C3(`.codex/agents/moai/*.toml`)를 손으로 고치기 — 다음 방출에서 덮인다.
9. 배포 미러(C2)에 SPEC ID·카드 번호·날짜 쓰기 — 템플릿 중립성 CI 위반.
10. merge-base 핀 — gitflow-lane-protocol §8.
11. 0매칭 초록·전체 스위트 로컬 실행.

## §L 참조

- spec.md §A(측정)·§B(요구)·§E(범위 밖)·§F(제약)
- acceptance.md — AC·RED/GREEN 두 칸 규율
- decision-index.md — Q1-Q8 권위 표기와 verdict(§G 해결된 결정과 번호 일치)
- progress.md — Decision Log(Jev 판정 원문)
- `.claude/rules/moai/development/verification-completeness.md` §1-2 — 관측된 실패·두 칸 채택
- SPEC-CODEX-GATE-SCOPE-001 plan.md §B-§D — 해상기·요청·영수증 설계(재사용 원천)
