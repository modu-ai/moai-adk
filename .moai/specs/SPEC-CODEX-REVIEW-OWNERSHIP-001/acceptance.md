---
id: SPEC-CODEX-REVIEW-OWNERSHIP-001
title: "수락 기준 — codex 리뷰 소유권 재배치"
version: "0.1.0"
created: 2026-10-02
---

# SPEC-CODEX-REVIEW-OWNERSHIP-001 — 수락 기준

**트리 핀(문서 수준, 자체 핀이 없는 모든 기준에 적용): `3ae43ed8e78ffa673ca238227df6ca7202c1ce70`** — 본 트리 `git rev-parse HEAD` 출력(2026-10-02, 이 SPEC 디렉터리를 추가한 커밋; 인용 경로는 모두 직전 커밋 `c50da9c2f` 와 동일). 아래 장부의 모든 값은 이 핀에서 이 레인이 다시 실행해 얻었다(plan-audit 1회차가 같은 값을 독립 재현했다).

**개수 규칙.** 수락 기준은 `### AC-` 제목 수로 센다 = **15**(하위 ID 없음). 요구는 `### REQ-` 제목 수 = 14. 둘 다 Tier M 상한 16 이하다.

## §A 관측 규율 [HARD]

모든 AC 는 **게이트·도구가 조립한 요청의 대상 필드**(target·cwd), **GLM 으로 보낸 자료의 구성**, **설정 판독값**, **`tools/list` 입출력 스키마와 `tools:` 줄**, **감사 영수증 저장소의 호출 전후 목록**을 관측한다. verdict 값 단독은 어떤 AC 의 근거도 못 된다 — 스텁 reviewer 는 요청과 무관한 값을 돌려줄 수 있다.

라이브 codex·라이브 GLM 을 요구하는 AC 는 없다. 해상기·판독기·RPC 는 기존 주입 seam(`reviewScopeResolver`, `reviewGateChangeDetector`, `codexLookPath`, 리뷰 RPC 호출 지점, `glmKeyLoader`/HTTP 스텁)으로 검증한다. 라이브 프로브(plan.md §G 위험 5)는 AC 가 아니라 **기록되는 관측**이며 skip 은 통과가 아니라 **미관측(Gap)**이다.

## §B 픽스처

**카드 워크트리 픽스처**: 실제 git 저장소+연결 워크트리. 브랜치 `WT-fixture-card`, develop 분기 후 커밋 1건(파일 A), 추적 파일 미커밋 수정 1건(파일 B), 비추적 신규 파일 1건(파일 C), **추적된 런타임 접두 경로 미커밋 수정 1건(파일 D, 예 `.moai/state/fixture-tracked.json`)**. primary 역할 트리 — 미커밋 외부 WIP 1건(파일 F, 카드 소유 아님).

**비카드 트리 픽스처**: `develop` 브랜치 트리+미커밋 변경 1건. 변형: detached HEAD, 비 git 디렉터리. **`WT-` 기저 불가 픽스처**: 브랜치 `WT-fixture-nobase`, 로컬 `develop` 참조 없음.

**설정 픽스처**: 설정 루트의 `.moai/config/sections/workflow.yaml` 에 `workflow.codex.review_gate.enabled: true` 와 `tree_scope` 값을 변형.

## §C RED/GREEN 두 칸 규율 [HARD]

`verification-completeness.md` §2 를 따른다: 기준마다 **RED-now 칸**(구현 전 트리에서 관측된 실패, 왜 빨간지 명시)과 **GREEN 경로 칸**(어느 마일스톤이 무엇으로 뒤집는지)을 짝으로 갖는다. 이미 초록인 기준은 RED 를 가장하지 않고 **회귀 칸(preserve)**으로 분류하며 변경 전 초록 관측을 증거로 든다.

- **소스 수준 RED-now** — 아래 증거 장부(트리 핀 위 측정; 명령은 읽기 전용 단일 호출, stdout 원문, 종료 코드).
- **행동 수준 RED-now** — 요청·결정·도구 호출을 단정하는 Go 테스트는 아직 존재하지 않으므로 plan-time 에 실행해 볼 수 없다. 이 기준들은 **M1 에서 테스트를 먼저 추가해 `-v` 로 `=== RUN` 과 `--- FAIL` 을 관측하고 `.moai/reports/t1422/red/` 에 구현 전 트리 SHA 와 함께 보존할 때까지 릴리스 차단(release-blocking) 자격이 없다.** 그 전에는 PASS 로 기록하지 않는다. 각 RED 는 §D 의 "RED 이유"가 가리키는 사유(기능·도구·문구 부재) 때문에 빨간 것이어야 하며, 다른 사유(예: 픽스처 오류)로 빨간 RED 는 무효다.
- **회귀 칸(변경 전 초록이어야 하고 변경 뒤에도 초록)**: AC-003·AC-004, AC-012 의 감사 도구 보유 집합(P1), AC-010 의 `codex_audit` 대조 칸, AC-015 의 인벤토리 불변 칸(P3·P5).

### 증거 장부 (트리 핀 `3ae43ed8e78ffa673ca238227df6ca7202c1ce70`)

```text
[L1] command: grep -c tree_scope internal/cli/codex_review_gate.go
     stdout:  0
     exit:    1
     red because: 게이트 핸들러가 tree_scope 를 읽지 않는다 (정책 부재) — AC-002/005/006 의 skip 행동이 없다
[L2] command: grep -c tree_scope internal/config/types.go
     stdout:  0
     exit:    1
     red because: CodexReviewGateConfig 에 필드가 없다 — AC-001 의 로더 일치
[L3] command: grep -c codex_review internal/mcp/catalog.go
     stdout:  0
     exit:    1
     red because: 카탈로그에 자기 리뷰 도구가 없다 — AC-007~011
[L4] command: grep -c 'wantCatalogSize = 47' internal/mcp/catalog_test.go
     stdout:  0
     exit:    1
     red because: 불변식이 45 로 고정돼 있다 — AC-015
[L5] command: grep -c mcp__moai__codex_review .claude/agents/moai/manager-develop.md
     stdout:  0
     exit:    1
     red because: 에이전트 tools 줄에 자기 리뷰 도구가 없다 — AC-012
[L6] command: grep -c card-review .claude/rules/moai/workflow/kanban-dispatch.md
     stdout:  0
     exit:    1
     red because: 레인 카드 단계 교리에 card-review 가 없다 — AC-013
[L7] command: grep -c 'Twenty-two tools' .claude/rules/moai/core/moai-mcp-tools.md
     stdout:  0
     exit:    1
     red because: project_root 선언 도구 수 문장이 아직 Twenty (20) — AC-015
[L8] command: grep -c tree_scope internal/template/templates/.moai/config/sections/workflow.yaml
     stdout:  0
     exit:    1
     red because: 템플릿에 주석 예시가 아직 없다 — AC-015
[L9] command: grep -c 'turn-end codex review gate' .claude/rules/moai/workflow/kanban-dispatch.md
     stdout:  0
     exit:    1
     red because: 리더 세션에 턴 종료 게이트가 없다는 교리 문장이 없다 — AC-014
[L10] command: grep -c 'f.mcp.tools.codex_review' internal/web/assets/i18n.js
     stdout:  0
     exit:    1
     red because: 두 도구의 콘솔 i18n 항목이 없다 — AC-015
[P1] command: grep -l mcp__moai__codex_audit .claude/agents/moai/manager-develop.md .claude/agents/moai/manager-docs.md .claude/agents/moai/manager-lead.md .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md
     stdout:  .claude/agents/moai/sync-auditor.md
              .claude/agents/moai/plan-auditor.md
     exit:    0
     preserve: 감사 도구 보유자는 둘뿐 — 구현 뒤에도 같은 두 파일이어야 한다
[P2] command: grep -c recordAuditReceipt internal/cli/mcp_codex.go
     stdout:  2
     exit:    0
     positive control: AC-010 의 "자기 리뷰 핵심 파일에서 0" 단정이 빈 스윕이 아님을 보이는 대조군
[P3] command: grep -c tree_scope internal/config/testdata/shipped_key_inventory.yaml
     stdout:  0
     exit:    1
     preserve: 인벤토리에 tree_scope 항목이 없다 — 구현 뒤에도 0 (AC-015)
[P4] command: grep -c 'workflow.codex.review_gate.enabled' internal/config/testdata/shipped_key_inventory.yaml
     stdout:  1
     exit:    0
     positive control: P3 의 0 이 빈 스윕이 아님 (같은 파일에서 이웃 키는 1)
[P5] command: shasum -a 256 internal/config/testdata/shipped_key_inventory.yaml
     stdout:  a3d2c397827130d71047fad9aa59c6b690d64394b222dca56adc44db149fc409  internal/config/testdata/shipped_key_inventory.yaml
     exit:    0
     preserve: 인벤토리 바이트 불변 (AC-015)
[P6] command: go test ./internal/cli/ -run '^TestCodexReviewGate_TreeScopeRequestShapeUnchanged$' -count=1 -v   (환경 세척: MOAI_KANBAN* unset)
     stdout:  === RUN   TestCodexReviewGate_TreeScopeRequestShapeUnchanged
              --- PASS: TestCodexReviewGate_TreeScopeRequestShapeUnchanged (1.22s)
     exit:    0
     preserve: 비카드 트리 요청 형태 (AC-003)
[P7] command: go test ./internal/cli/ -run '^TestCodexReviewGate_CardScopeRequestIsCardDiff$' -count=1 -v   (환경 세척)
     stdout:  === RUN   TestCodexReviewGate_CardScopeRequestIsCardDiff
              --- PASS: TestCodexReviewGate_CardScopeRequestIsCardDiff (1.74s)
     exit:    0
     preserve: 카드 세션 요청이 카드 diff (AC-004 (a))
[P8] command: go test ./internal/cli/ -run '^TestCodexReviewScope_UnidentifiedFallsToTree$' -count=1 -v   (환경 세척)
     stdout:  === RUN   TestCodexReviewScope_UnidentifiedFallsToTree
              (하위 4건 RUN: detached_HEAD · plain_develop_branch · WT-_branch_without_a_develop_base · non-git_directory)
              --- PASS: TestCodexReviewScope_UnidentifiedFallsToTree (2.73s)
     exit:    0
     preserve: WT- 브랜치+기저 불가 → 트리 클래스 폴백 (AC-004 (b))
[P9] command: grep -c 'f.mcp.tools.codex_audit' internal/web/assets/i18n.js
     stdout:  8
     exit:    0
     positive control: L10 의 0 이 빈 스윕이 아님 — 도구 하나당 title·desc × 4로케일 = 8
```

### 변이 점검(mutant probe) 요약

| 기준 | 요구를 어기면서 기준을 만족하는 변이 | 이 변이를 죽이는 칸 |
|---|---|---|
| AC-002 (skip) | 전 스코프 클래스에 skip | AC-004 (a) |
| AC-002 (skip) | `WT-` 접두 세션의 기저 불가 폴백까지 skip | AC-004 (b) |
| AC-002 (skip) | 요청을 만든 뒤 버리고 셀프게이트는 도는 구현 | 셀프게이트·reviewer·`codexLookPath` 호출 0 단정 |
| AC-001 | 평면형·주석·다른 블록의 키까지 읽는 판독기 | AC-001 의 오배치 픽스처 |
| AC-005 | 정책이 `MOAI_KANBAN_LABEL` 등을 읽어 "leader" 값을 면제 | 전체 env 집합 행렬(같은 결정이어야 함) |
| AC-008 | 도구가 자체 `WT-` 판별·기저 계산 — codex 요청 또는 **GLM 자료**가 `git merge-base origin/develop HEAD` 등을 쓴다 | sentinel 전파가 codex 요청과 GLM 자료 양쪽에서 확인됨; 흡수 후 양쪽 기저 이동 |
| AC-008 | GLM 자료에 추적된 런타임 접두 파일 D 가 들어감 | 자료에 D 부재 단정 |
| AC-010 | advisory·scope·base·backend·tree 를 inconclusive 생성자에서만 채움 | pass·fail·inconclusive × 두 백엔드 전부에서 단정 |
| AC-010 | pass/fail 결과가 비어 있지 않은 `gate_unmet` 를 담음 | 세 verdict 모두 `gate_unmet` 부재 단정 |
| AC-010 | 영수증을 쓰고 지움 | 정적 단정(호출 0, 양성 대조 P2)이 병행 |
| AC-012 (보유) | 모든 에이전트에 도구 부여 | `tools:` 집합 동치 단정 |
| AC-013 | 앵커 문자열이 부정 문맥·잘못된 위치에 있음 | 순서 있는 목록의 토큰 위치 단정 + 부정문 부재 단정 |

## §D AC 매트릭스

| AC | 요구 | 시작 상태 | RED 이유 / 회귀 근거 | 관측 대상 |
|---|---|---|---|---|
| AC-001 | REQ-CRO-001 | RED(행동) | 판독기·필드 부재 (L1, L2) | 판독기 진리표+로더 일치 |
| AC-002 | REQ-CRO-002 | RED(행동) | skip 분기 부재 (L1) | 두 자동 경로의 reviewer·셀프게이트 호출 0, 로그 행 |
| AC-003 | REQ-CRO-003 | **회귀(초록)** | P6 | 비카드 트리 요청 target·cwd |
| AC-004 | REQ-CRO-004 | **회귀(초록)** | P7, P8 | 카드 요청, `WT-`+기저 불가 요청·로그 (skip 설정 하) |
| AC-005 | REQ-CRO-005 | RED(행동) | skip 분기 부재 (L1) — 비카드 트리 행이 skip 되지 않아 빨강 | env 집합 행렬 결정 동일성 |
| AC-006 | REQ-CRO-006 | RED(행동) | skip 분기 부재 (L1) | 두 경로 결정 일치·설정 루트·명시 생산자 |
| AC-007 | REQ-CRO-007 | RED(소스+행동) | 도구 부재 (L3) | `tools/list` 입출력 스키마·카탈로그 |
| AC-008 | REQ-CRO-008 | RED(행동) | 도구 부재 (L3) | codex 요청·GLM 자료 구성·기저 이동·sentinel·비카드 |
| AC-009 | REQ-CRO-009, 010 | RED(행동) | 도구 부재 (L3) | 미커밋 요청/자료·비추적 목록·`truncated` |
| AC-010 | REQ-CRO-010 | RED(행동) + 대조 **초록** | 도구 부재 (L3); 대조 `codex_audit` 는 현행 | advisory·메타데이터·`gate_unmet`·영수증 목록 (두 백엔드×세 verdict) |
| AC-011 | REQ-CRO-007, 010 | RED(행동) | 도구 부재 (L3) | `model` 입력·핀 미적용 |
| AC-012 | REQ-CRO-011 | RED(소스) + 회귀 **초록** | 보유 부재 (L5); P1 | `tools:` 줄 집합 |
| AC-013 | REQ-CRO-012 | RED(소스) | 교리 부재 (L6) | 순서 있는 목록·앵커 |
| AC-014 | REQ-CRO-013 | RED(소스) | 문장 부재 (L9) | 리더 문장·GAP 규칙의 위치 |
| AC-015 | REQ-CRO-014 | RED(소스) + 회귀 **초록** | L4, L7, L8, L10; P3, P5 | 카탈로그 크기·`project_root` 문서·템플릿·인벤토리·i18n |

REQ→AC 누락 없음: 001→001 · 002→002 · 003→003 · 004→004 · 005→005 · 006→006 · 007→007·011 · 008→008 · 009→009 · 010→009·010·011 · 011→012 · 012→013 · 013→014 · 014→015.

---

### AC-001 — `tree_scope` 판독은 `workflow` 아래 중첩 경로의 `skip` 만 skip 으로 읽는다

**Given** 설정 픽스처 — `skip`, 앞뒤 공백이 낀 `Skip`, `SKIP`, `review`, 키 부재, 빈 값, 알 수 없는 값 `never`, 깨진 YAML, 파일 부재; **오배치** 픽스처 — `workflow` 루트 없는 평면형(`codex:\n  review_gate:\n    tree_scope: skip`), 주석 처리된 `# tree_scope: skip`, `workflow.multi.review_gate.tree_scope: skip`, `workflow.codex.task.tree_scope: skip`,
**When** 판독기가 각 설정 루트를 읽으면,
**Then** 올바른 중첩 경로의 `skip` 계열 셋(`skip`·` Skip `·`SKIP`)만 skip 으로, 나머지 전부는 `review` 로 읽힌다(오배치 네 건이 `review` 인 것이 양성 대조 — 같은 파일에서 올바른 경로의 `skip` 은 skip). 설정 로더(`config.Loader`)가 같은 파일에서 돌려주는 값과 판독기 값이 일치한다(`TestReviewGateReaders_AgreeWithConfigLoader` 계열 핀, 기존 `enabled` 판독기의 평면형 거부 테스트 `flatCodexOn` 선례).

### AC-002 — `WT-` 브랜치 증거가 없는 트리 스코프 세션은 skip 에서 리뷰·셀프게이트를 거치지 않는다 (두 자동 경로)

**Given** `enabled: true`+`tree_scope: skip` 인 설정 루트, 비 `WT-` 브랜치 트리의 미커밋 변경(변형: `develop`·detached HEAD·비 git 디렉터리), 호출 횟수를 세는 `codexLookPath`·리뷰 RPC·셀프게이트 검출기 스텁,
**When** (i) `HandleCodexReviewGate` 가 그 세션의 Stop 입력을 처리하고 (ii) Codex Stop 체인의 codex 리뷰 멤버가 같은 상태로 평가되면,
**Then** (i) 결과는 ALLOW 이고 `codexLookPath`·리뷰 RPC·셀프게이트 검출기 호출이 전부 0 이며 stderr 로그에 `scope=tree` 와 skip 사유(키 값 포함)가 한 행으로 남는다. (ii) 결과는 Allow, 상태 `not-applicable`, 사유에 `tree_scope` 가 들어가고 영수증을 읽지 않으며(`ReceiptRead` false) RPC·`codexLookPath` 호출이 0 이다. **대조 칸**: 같은 픽스처의 `tree_scope: review` 는 (i) 에서 리뷰 RPC 를 정확히 1회 부르고 target `uncommittedChanges`, cwd = 세션 트리이며 (ii) 에서 영수증 판독 경로로 간다.

### AC-003 — `skip` 이 아닐 때 비카드 트리 요청은 변하지 않는다 (회귀 칸)

**Given** `tree_scope` 가 부재이거나 `review` 이고 비카드 트리의 미커밋 변경,
**When** 게이트가 요청을 조립하면,
**Then** 요청은 사전 형태와 shape-identical 이다 — 대상 `uncommittedChanges`, cwd = 해상 트리(SPEC-CODEX-REVIEW-TARGET-001 REQ-CRT-006 직렬화). 변경 전 트리에서 이미 초록인 P6 의 테스트가 변경 뒤에도 초록이다. 이 기준은 RED 가 아니라 **회귀 칸**이다.

### AC-004 — `WT-` 브랜치 세션은 `skip` 설정에서도 오늘처럼 리뷰된다 (회귀 칸)

**Given** `enabled: true`+`tree_scope: skip` 과 (a) 카드 워크트리 픽스처(기저 계산 가능), (b) `WT-` 기저 불가 픽스처,
**When** 각 세션의 턴이 게이트에 닿으면,
**Then** (a) reviewer 가 1회 호출되고 요청 target 은 `baseBranch`, `branch` 값은 호출 시점의 `git merge-base develop HEAD` 출력, cwd 는 카드 워크트리다. (b) reviewer 가 1회 호출되고 요청은 `{target: uncommittedChanges, cwd: <트리>}` 이며 스코프 로그 행의 basis 에 "merge base unavailable" 이 남는다(skip 로그 행은 없다). 두 칸 모두 변경 전 트리에서 skip 키가 무시되므로 **이미 초록**이다(P7·P8 이 같은 동작을 skip 설정 없이 초록으로 보인다) — 이 기준은 skip 구현이 `WT-` 세션을 삼키지 못하게 하는 회귀 가드다.

### AC-005 — 정책 결정은 env 집합 전체에 둔감하다

**Given** `tree_scope: skip` 과 다음 env 행렬 — `config.EnvMoaiKanban`·`EnvMoaiKanbanID`·`EnvMoaiKanbanLabel`·`EnvMoaiKanbanLeadAddr`·`EnvMoaiKanbanSettingsInjected`(레인 env 세척 집합)와 `envkeys.go` 의 나머지 `EnvMoaiKanban*`·`EnvMoaiFactoryWorker`·`EnvMoaiFactoryWorkers` 를 (R1) 임의 값 X 로 전부 설정, (R2) 다른 임의 값 Y 로 전부 설정, (R3) 전부 부재; 각 행을 (T1) `develop` 트리, (T2) `WT-` 카드 트리에서 평가,
**When** 각 조합이 게이트에 닿으면,
**Then** T1 은 R1·R2·R3 모두 동일하게 skip, T2 는 R1·R2·R3 모두 동일하게 리뷰된다. 판정이 어떤 env 값에도 따라 달라지지 않는다는 **양성 관측**(값 셋에서 동일 결과)으로 판정한다. 상수 목록은 `envkeys.go` 에서 파생한다(카드 t1399 가 상수를 지우면 같은 변경에서 이 목록도 갱신).

### AC-006 — 두 자동 경로와 설정 루트, 명시 생산자

**Given** 비카드 트리·`WT-` 트리 × (`skip`, `review`) 네 조합, 그리고 `tree_scope` 를 `enabled` 가 true 인 루트에만 / 그 외 루트에만 심은 변형; **설정 고아 워크트리 픽스처**(`.moai` 를 추적하지 않는 저장소의 연결 워크트리 — Claude 경로의 `StoreRoot` 가 projectDir 와 달라지는 유일한 경우),
**When** Claude 훅 경로와 Codex 체인 멤버 6 이 각각 평가되고, 이어 `produceCodexReviewReceipt` 가 비카드 트리·`skip` 에서 실행되면,
**Then** 일반 픽스처에서 네 조합 모두 두 경로의 skip/리뷰 결정이 일치하고, 각 경로는 자기가 `enabled` 를 읽는 바로 그 루트의 `tree_scope` 로 결정한다(다른 루트에 심은 값은 영향 없음). 고아 워크트리 픽스처에서 Claude 경로는 primary 의 키로 결정한다(Codex 체인은 `c.root` 를 읽으므로 같은 픽스처에서 `enabled` 를 false 로 읽어 not-applicable — 기존 비대칭이며 각 경로가 자기 `enabled` 와 같은 루트를 쓴다는 단정만 한다). `produceCodexReviewReceipt` 는 `skip` 에서도 리뷰 RPC 를 1회 부르고 영수증을 기록한다.

### AC-007 — 자기 리뷰 도구의 표면

**Given** `moai mcp-server` 인프로세스 클라이언트로 `tools/list` 를 받으면,
**Then** `codex_review`·`glm_review` 가 모두 있고, 각 입력 스키마에서 `scope` 는 enum `{card, uncommitted}` 이며 `required` 에 들어 있고 `project_root`·`model` 속성이 선언돼 있으며 `focus` 속성은 없다. 각 출력 스키마는 속성 `advisory`·`scope`·`base`·`backend`·`tree`·`truncated`·`excluded_untracked` 를 선언한다. 읽기 전용 힌트는 true, 카탈로그 선언은 `WriteCapable: false`, 설명에는 "advisory" 가 들어 있다. `codex_audit`·`glm_audit`·`claude_audit`·`audit_multi` 의 입력·출력 스키마(`target` enum 포함)는 M1 에서 기록한 변경 전 스냅숏과 동일하다.

### AC-008 — 카드 스코프는 게이트와 같은 해상기가 정하고 기저는 핀되지 않는다 (요청·자료·비카드)

**Given** 카드 워크트리 픽스처(A 커밋, B 추적 미커밋, C 비추적, D 추적 런타임 접두 미커밋, F 외부 WIP)와 비카드 트리 픽스처,
**When** `codex_review`·`glm_review` 를 `scope: card`, `project_root: <트리>` 로 부르면,
**Then** (i) codex 요청은 `reviewRequestParams(reviewScopeResolver(<카드 트리>))` 와 `reflect.DeepEqual` 이고 `branch` 값은 호출 시점의 `git merge-base develop HEAD` 출력, cwd 는 카드 트리다. (ii) GLM 으로 보낸 자료는 테스트가 독립 계산한 `git diff <그 merge-base> -- . <런타임 접두 exclude>` 출력과 같고, A·B 는 있으며 F·C·D 는 없다. (iii) 카드가 develop 신규 커밋을 흡수한 뒤 같은 호출에서 codex `branch` 값과 GLM 자료의 기저가 **둘 다** 새 merge-base 로 바뀐다(흡수 전 값이 남지 않는다). (iv) `reviewScopeResolver` 를 sentinel MergeBase(실재하는 조상 커밋)를 돌려주는 스텁으로 바꾸면 codex 요청이 그 sentinel 을 싣고 GLM 자료가 `git diff <sentinel>` 이다 — 도구가 두 번째 판별기·두 번째 기저 계산을 갖지 않는다는 양성 관측이다. (v) 비카드 트리(`develop`·detached·비 git)에서는 verdict `inconclusive`, 요약에 카드 워크트리가 아니라는 원인, 리뷰 RPC 호출 0·GLM HTTP 호출 0 이며 미커밋 변경을 대신 리뷰하지 않는다. **이 기준이 주장하지 않는 것**: codex 가 `baseBranch` 대상으로 어떤 작업 트리 파일을 읽는지, SHA 를 받는지 — 라이브 프로브의 기록 관측이다.

### AC-009 — 미커밋 스코프, 비추적 목록, 잘림 표식

**Given** `WT-` 카드 트리(커밋분 있음)와 비카드 트리, 비추적 비런타임 파일 C·C2 와 비추적 런타임 접두 파일, 200,000 바이트 상한을 넘는 추적 변경 하나를 가진 별도 픽스처,
**When** `scope: uncommitted` 와 `scope: card` 로 두 도구를 부르면,
**Then** (i) `codex_review` `scope: uncommitted` 의 요청은 두 트리 모두 `{target: uncommittedChanges, cwd: <그 트리>}` 와 DeepEqual 이다(카드 트리에서도 `baseBranch` 요청이 아니다). (ii) `glm_review` `scope: uncommitted` 의 자료는 독립 계산한 `git diff HEAD -- . <exclude>` 와 같다. (iii) `glm_review` 는 **두 스코프 모두** `excluded_untracked` 에 C·C2 를 담고 비추적 런타임 접두 경로는 담지 않는다. (iv) 상한 초과 픽스처에서 `glm_review` 결과의 `truncated` 는 true 이고 자료 길이가 상한(+고정 잘림 안내문) 이내이며, 상한 미만 픽스처에서는 false 다. (v) `codex_review` 결과는 항상 `truncated` false, `excluded_untracked` 빈 배열이다.

### AC-010 — 자기 리뷰는 모든 verdict 에서 조언이며 영수증·`required` 게이트와 무관하다

**Given** 스텁 reviewer 가 `pass`·`fail`·`inconclusive`(오류/부재 포함)를 각각 돌려주는 세 변형을 **두 백엔드(codex·GLM) 모두**에 구성하고, 트리 설정에 `workflow.audit.gates.codex: required`, 호출 전 `.moai/state/audit-receipts/` 목록 스냅숏,
**When** `scope: card` 로 각 변형을 부르면(codex 부재 변형 포함),
**Then** 여섯 결과 모두 `advisory == true`, `scope`·`base`(카드: 비어 있지 않음)·`backend`·`tree` 가 채워져 있고, `gate_unmet` 가 부재이거나 비어 있으며, 호출 뒤 영수증 목록이 호출 전과 동일하다. codex 부재+`required` 에서도 verdict 는 `inconclusive` 이지 `fail` 이 아니다. 출력 스키마가 같은 필드를 선언한다(AC-007). 정적 단정: `grep -c recordAuditReceipt internal/cli/mcp_selfreview.go` → 0 이고 같은 명령의 양성 대조군 `internal/cli/mcp_codex.go` 는 2 이다(P2). **대조 칸(회귀 초록)**: 같은 `required` 픽스처의 `codex_audit` 는 변경 뒤에도 verdict `fail` 과 비어 있지 않은 `gate_unmet` 를 돌려준다.

### AC-011 — 모델은 호출자 입력 또는 백엔드 기본이며 감사 핀은 적용되지 않는다

**Given** 설정에 `workflow.audit.codex.model` 과 `workflow.audit.glm.model` 핀이 있고,
**When** (a) `model` 입력 없이 (b) `model: X` 로 두 도구를 부르면,
**Then** (a) GLM 요청 본문의 `model` 은 작업 위임 경로의 기본 모델이며 핀 값이 아니고, codex 쪽 스텁이 관측한 모델 값은 게이트 경로가 같은 픽스처에서 쓰는 값과 같다(핀 아님). (b) 두 백엔드 모두 요청에서 관측한 모델이 `X` 다. (Q6 잠정 — 핀을 적용하려면 해상기 호출 한 번을 바꾸고 이 기준을 뒤집는다.)

### AC-012 — 도구 보유는 허용 목록과 정확히 일치한다

**Given** `.claude/agents/moai/*.md`(C1)와 `internal/template/templates/.claude/agents/moai/*.md`(C2)의 모든 에이전트 정의,
**When** `tools:` 줄을 파싱하면,
**Then** `mcp__moai__codex_review`·`mcp__moai__glm_review` 는 manager-develop·manager-docs·manager-lead 에만 들어 있고, `mcp__moai__codex_audit`·`mcp__moai__glm_audit` 는 plan-auditor·sync-auditor 에만 들어 있으며(변경 전과 같은 두 파일 — P1), C1 과 C2 의 `tools:` 집합이 에이전트별로 같다. `make agents-emit-check` 는 종료 코드 0 이다.

### AC-013 — 카드 리뷰 단계는 순서 있는 목록에서 run 출구 검증과 통합 사이에 놓인다

**Given** `.claude/rules/moai/workflow/kanban-dispatch.md`·`kanban-dispatch-detail.md` 와 C2 미러,
**When** 문구를 검사하면,
**Then** 스텁에 `card-review` 단계 이름, 증거 경로 `.moai/reports/<card-id>/card-review.md`, "advisory" 가 있다. detail 의 순서 있는 단계 목록(항목 토큰 `[run-exit]`, `[card-review]`, `[integration]`, `[report]`)에서 `[run-exit]` < `[card-review]` < `[integration]` < `[report]` 순서이고, 해당 항목 줄과 같은 절에 "no card-review"·"not required" 류 부정 문구가 없으며, 재리뷰 상한 2 가 적혀 있다. C2 미러는 같은 단정을 만족하고 SPEC ID·카드 번호·날짜 토큰이 없으며 템플릿 중립성 검사(`.github/workflows/template-neutrality-check.yaml`)가 초록이다.

### AC-014 — 리더 측 규칙: 게이트 부재 문장, GAP 규칙, 연속 발화

**Given** `kanban-dispatch.md` 와 C2 미러,
**When** 리더 완료 판독 절(`Completion is read, never trusted`)과 카드 단계 절을 추출하면,
**Then** 완료 판독 절 안에 (i) 리더의 증거 경로 목록에 `.moai/reports/<card-id>/card-review.md` 가 들어 있고, (ii) 그 파일을 인용하지 않고 부재 사유도 적지 않은 카드는 gap 이며 칼럼에 남는다는 문장이 있고, 카드 단계 절에 (iii) 리더 세션에는 `turn-end codex review gate` 가 없고 같은 도구로 직접 리뷰한다는 문장(L9)이 있다. 세 문장이 부정·위치 오류 문맥에 있지 않다(추출한 절 안에 위치).

### AC-015 — 정합 대상이 새 도구를 같게 서술하고 템플릿·인벤토리는 라이브 키를 늘리지 않는다

**Given** 두 도구가 등록된 서버와 구현 뒤 템플릿·인벤토리,
**Then** (i) `wantCatalogSize` 는 47(L4), `TestMoaiMCPServer_RegistrationMatchesCatalog`·`TestMoaiMCPServer_AnnotationsMatchCatalog`·`TestMoaiMCPTools_WriteCapableSet`(쓰기 가능 집합 불변)이 초록. (ii) `TestProjectRootDocMatchesServer`·`TestDocsSiteProjectRootMatchesServer` 가 `project_root` 선언 도구 22 개(문장 `Twenty-two`, 로케일별 개수 문구 포함, 하이픈을 읽는 정규식)로 초록이고 두 규칙 파일 사본의 "Four of the twenty REQUIRE it" 문장이 현재 수에 맞게 고쳐져 있다. (iii) 파싱된 템플릿의 `workflow.codex.review_gate` 맵은 `enabled`(false)만 갖고 어디에도 `tree_scope` 라이브 키가 없으며(변경 전 대비 키·값 집합 동일, 추가된 줄은 전부 `#` 로 시작) 템플릿 텍스트에는 주석 예시가 있다(`grep -c tree_scope` 가 L8 의 0 에서 ≥1). (iv) 인벤토리 파일 해시가 P5 와 같고 `tree_scope` 항목이 없으며(P3 = 0, 대조군 P4 = 1) `internal/settings/schema_sections.go`·`internal/web/fieldsets.templ`·`internal/web/assets/i18n.js` 는 `tree_scope` 를 포함하지 않는다. (v) `i18n.js` 에 두 도구의 `f.mcp.tools.<name>.enabled.title/.desc` 가 4로케일 모두에 있다 — `grep -c 'f.mcp.tools.codex_review\|f.mcp.tools.glm_review'` → 16(L10 이 0, 대조군 P9 는 도구 하나당 8) — 그리고 웹 패키지의 `TestDataI18nKeysSubsetOfDictionary`(`i18n_test.go:255`)·`TestI18nKeyCoverageForward`(`i18n_governance_test.go:408`)·`TestI18nUntranslatedValues`(:217)·`TestMCPConsoleRendersAllTools`(`mcp_console_test.go:32`)가 초록이다. 이 테스트들이 per-tool 항목을 강제한다는 것은 테스트 소스 판독에 의한 추론이며(빨간 상태 미실행) run 이 M3 에서 실행해 관측한다.

---

## §E 품질 게이트와 DoD

- **DoD**: §D 전 AC 최종 상태 PASS · 행동 수준 RED 관측 기록이 `.moai/reports/t1422/red/` 에 구현 전 트리 SHA 와 함께 존재 · **M3 종료 시 `TestProjectRootDocMatchesServer`·`TestDocsSiteProjectRootMatchesServer`·웹 i18n 테스트가 초록**(plan.md §I) · `go vet`(darwin)+`GOOS=windows GOARCH=amd64 go build ./...` 신규 0 · 수정 파일 커버리지 `quality.yaml` `test_coverage_target`(85) 이상 · `make agents-emit-check`·`make build` 통과 후 갱신 산출물(`catalog.yaml`, `internal/template/templates/.codex/agents/moai/*.toml`) 커밋 포함 · §D 추적표 상 미매핑 REQ 없음 · 라이브 프로브 기록 파일 존재 또는 codex 부재 Gap 기록.
- **간접 검증**: AC-005 는 양성 귀결(값 셋에서 동일 결과)을 포함한다. AC-008 (iv)는 "두 번째 판별기가 없다"를 sentinel 의 양성 전파로 codex·GLM 양쪽에서 검증한다. AC-010 의 정적 0 단정은 양성 대조군 P2(2)와 함께만 읽는다.
- **선행 폐쇄 게이트**: M1 회귀 칸(AC-003·AC-004, AC-012 의 P1, AC-010 의 `codex_audit` 대조, AC-015 의 P3·P5)이 변경 전 트리에서 초록으로 관측되지 않으면 이후 마일스톤을 시작하지 않는다.
- **측정 규율**: 소관 패키지 단위로 재측정한다(`-run` 이름 패턴 단독 재측정 금지 — 정문 가드를 0개 고르고도 `ok` 로 보인다). 전체 스위트 로컬 금지.
