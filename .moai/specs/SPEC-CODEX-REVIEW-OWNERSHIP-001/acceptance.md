---
id: SPEC-CODEX-REVIEW-OWNERSHIP-001
title: "수락 기준 — codex 리뷰 소유권 재배치"
version: "0.1.0"
created: 2026-10-02
---

# SPEC-CODEX-REVIEW-OWNERSHIP-001 — 수락 기준

**트리 핀(문서 수준, 자체 핀이 없는 모든 기준에 적용): `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b`** — 본 트리 `git rev-parse HEAD` 출력(2026-10-02). 아래 RED-now 칸은 모두 이 트리에서 측정한 값이다.

## §A 관측 규율 [HARD]

모든 AC 는 **게이트·도구가 조립한 요청의 대상 필드**(target·cwd), **설정 판독값**, **`tools/list` 스키마와 `tools:` 줄**, **감사 영수증 저장소의 호출 전후 목록**을 관측한다. verdict 값 단독은 어떤 AC 의 근거도 못 된다 — 스텁 reviewer 는 요청과 무관한 값을 돌려줄 수 있다(SPEC-CODEX-GATE-SCOPE-001 §A 와 같은 규율).

라이브 codex·라이브 GLM 을 요구하는 AC 는 없다. 해상기·판독기·RPC 는 기존 주입 seam(`reviewScopeResolver`, `reviewGateChangeDetector`, `codexLookPath`, 리뷰 RPC 호출 지점, `glmKeyLoader`/HTTP 스텁)으로 검증한다. 라이브 프로브(plan.md M3)는 AC 가 아니라 부가 관측이며 skip 은 통과가 아니라 **미관측**이다.

## §B 픽스처

**카드 워크트리 픽스처**: 실제 git 저장소+연결 워크트리. 카드 워크트리 — 브랜치 `WT-fixture-card`, develop 분기 후 커밋 1건(파일 A), 추적 파일 미커밋 수정 1건(파일 B), 비추적 신규 파일 1건(파일 C). primary 역할 트리 — 미커밋 외부 WIP 1건(파일 F, 카드 소유 아님).

**비카드 트리 픽스처**: `develop` 브랜치 트리+미커밋 변경 1건. 변형: detached HEAD, 비 git 디렉터리.

**설정 픽스처**: 설정 루트의 `.moai/config/sections/workflow.yaml` 에 `workflow.codex.review_gate.enabled: true` 와 `tree_scope` 값을 변형(skip / Skip / review / 부재 / 빈 값 / 알 수 없는 값 / 깨진 YAML).

**env 행렬**: `MOAI_FACTORY_WORKER` 를 서로 다른 임의 문자열 2종 이상으로 바꿔 가며 쓴다(상수 `config.EnvMoaiFactoryWorker` 로만 참조). 값 의존은 곧 라벨 형식 결합이다.

## §C RED/GREEN 두 칸 규율 [HARD]

`verification-completeness.md` §2 를 따른다: 기준마다 **RED-now 칸**(구현 전 트리에서 관측된 실패, 왜 빨간지 명시)과 **GREEN 경로 칸**(어느 마일스톤이 무엇으로 뒤집는지)을 짝으로 갖는다.

- **소스 수준 RED-now** — 아래 증거 장부 L1-L7 이 본 트리에서 측정한 값이다(명령은 읽기 전용 단일 호출, stdout 원문, 종료 코드, 트리 핀 = 문서 수준 핀).
- **행동 수준 RED-now** — 요청·결정·도구 호출을 단정하는 Go 테스트는 아직 존재하지 않으므로 plan-time 에 실행해 볼 수 없다. 이 기준들은 **M1 에서 테스트를 먼저 추가해 `-v` 로 `=== RUN` 과 `--- FAIL` 을 관측하고 `.moai/reports/t1422/red/` 에 구현 전 트리 SHA 와 함께 보존할 때까지 릴리스 차단(release-blocking) 자격이 없다.** 그 전에는 PASS 로 기록하지 않는다.
- 변경 전 트리에서 초록이어야 하는 **회귀 칸**: AC-CRO-004(트리 스코프 요청 형태)·AC-CRO-014 의 감사 도구 보유 집합·AC-CRO-012 의 `codex_audit` 대조 칸·AC-CRO-001b 의 인벤토리 불변 칸(P3·P5). 이들은 변경 전에 초록을 관측하고, 변경 뒤에도 초록이어야 한다.

### 증거 장부 (소스 수준 RED-now, 트리 핀 `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b`)

```text
[L1] command: grep -c tree_scope internal/cli/codex_review_gate.go
     stdout:  0
     exit:    1
     red because: 게이트 핸들러가 tree_scope 를 어디서도 읽지 않는다 (정책 부재)
[L2] command: grep -c tree_scope internal/config/types.go
     stdout:  0
     exit:    1
     red because: CodexReviewGateConfig 에 필드가 없다
[L3] command: grep -c codex_review internal/mcp/catalog.go
     stdout:  0
     exit:    1
     red because: 카탈로그에 자기 리뷰 도구가 없다
[L4] command: grep -c 'wantCatalogSize = 47' internal/mcp/catalog_test.go
     stdout:  0
     exit:    1
     red because: 불변식이 45 로 고정돼 있다 (두 도구 추가 뒤 47)
[L5] command: grep -c mcp__moai__codex_review .claude/agents/moai/manager-develop.md
     stdout:  0
     exit:    1
     red because: 에이전트 tools 줄에 자기 리뷰 도구가 없다
[L6] command: grep -c card-review .claude/rules/moai/workflow/kanban-dispatch.md
     stdout:  0
     exit:    1
     red because: 레인 카드 단계 목록에 card-review 단계 교리가 없다
[L7] command: grep -c 'Twenty-two tools' .claude/rules/moai/core/moai-mcp-tools.md
     stdout:  0
     exit:    1
     red because: project_root 선언 도구 수 문장이 아직 Twenty (20)
[L8] command: grep -c tree_scope internal/template/templates/.moai/config/sections/workflow.yaml
     stdout:  0
     exit:    1
     red because: 템플릿에 주석 예시가 아직 없다 (AC-CRO-001b 의 "예시는 있다" 칸)
[P1] command: grep -l mcp__moai__codex_audit .claude/agents/moai/manager-develop.md .claude/agents/moai/manager-docs.md .claude/agents/moai/manager-lead.md .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md
     stdout:  .claude/agents/moai/sync-auditor.md
              .claude/agents/moai/plan-auditor.md
     exit:    0
     preserve: 감사 도구 보유자는 둘뿐 — 구현 뒤에도 같은 두 파일이어야 한다 (회귀 칸)
[P3] command: grep -c tree_scope internal/config/testdata/shipped_key_inventory.yaml
     stdout:  0
     exit:    1
     preserve: 인벤토리에 tree_scope 항목이 없다 — 구현 뒤에도 0 이어야 한다 (회귀 칸, AC-CRO-001b)
[P4] command: grep -c 'workflow.codex.review_gate.enabled' internal/config/testdata/shipped_key_inventory.yaml
     stdout:  1
     exit:    0
     positive control: P3 의 0 이 비어 있는 스윕이 아님을 보이는 대조군 (같은 파일에서 이웃 키는 1)
[P5] command: shasum -a 256 internal/config/testdata/shipped_key_inventory.yaml
     stdout:  a3d2c397827130d71047fad9aa59c6b690d64394b222dca56adc44db149fc409  internal/config/testdata/shipped_key_inventory.yaml
     exit:    0
     preserve: 인벤토리 파일 바이트 불변 — 구현 뒤 같은 명령의 해시가 같아야 한다
[P2] command: grep -c recordAuditReceipt internal/cli/mcp_codex.go
     stdout:  2
     exit:    0
     positive control: AC-CRO-012 의 "자기 리뷰 핵심 파일에서 0" 단정이 비어 있는 스윕이 아님을 보이는 대조군
```

### 변이 점검(mutant probe) 요약

| 기준 | 요구를 어기면서 기준을 만족하는 변이 | 이 변이를 죽이는 칸 |
|---|---|---|
| AC-CRO-002 (skip) | `enabled` 를 전역으로 끄는 구현 — 트리 세션도 카드 세션도 리뷰 안 함 | AC-CRO-005(카드 세션은 skip 에서도 리뷰된다) |
| AC-CRO-002 (skip) | 요청을 조립한 뒤 버리는 구현(reviewer 는 부르지 않지만 셀프게이트는 돈다) | 셀프게이트 호출 횟수 0 단정 + reviewer 호출 0 단정 |
| AC-CRO-009 (카드 해상) | 도구가 자체 `WT-` 판별·merge-base 를 계산하는 구현 | 스텁 해상기가 반환한 sentinel 이 요청에 실려야 한다(두 번째 판별기는 sentinel 을 못 싣는다) |
| AC-CRO-012 (영수증 없음) | 영수증을 쓰고 지우는 구현 | 정적 단정(`recordAuditReceipt` 호출 0, 양성 대조 P2)이 병행 |
| AC-CRO-014 (보유) | 모든 에이전트에 도구를 주는 구현 | `tools:` 집합 동치 단정(허용 목록 밖은 부재여야 함) |

## §D AC 매트릭스

| AC | 요구 | 시작 상태 | 관측 대상 | RED-now 근거 |
|---|---|---|---|---|
| AC-CRO-001a | REQ-CRO-001 | RED(행동) | 판독기 진리표 + 로더 일치 | L1, L2 |
| AC-CRO-001b | REQ-CRO-001 | RED(소스: 예시 부재) + 회귀 **초록**(라이브 키·인벤토리 불변) | 템플릿 파싱 키 집합·인벤토리 바이트 | L8, P3-P5 |
| AC-CRO-002 | REQ-CRO-002 | RED(행동) | Claude 경로: reviewer·셀프게이트 호출 0, 로그 행 | L1 |
| AC-CRO-003 | REQ-CRO-002 | RED(행동) | Codex 멤버 6: 상태·영수증 미판독 | L1 |
| AC-CRO-004 | REQ-CRO-003 | **회귀(초록)** | 트리 스코프 요청 target·cwd | — |
| AC-CRO-005 | REQ-CRO-004 | RED(행동) | 카드 스코프 요청 target·cwd (skip 설정 하) | L1 |
| AC-CRO-006 | REQ-CRO-005 | RED(행동) | env 행렬 결정 동일성 | L1 |
| AC-CRO-007 | REQ-CRO-006 | RED(행동) | 두 경로 동일 결정·설정 루트·명시 생산자 비적용 | L1 |
| AC-CRO-008 | REQ-CRO-007 | RED(소스+행동) | `tools/list` 스키마·카탈로그 | L3 |
| AC-CRO-009 | REQ-CRO-008 | RED(행동) | 카드 요청/자료·sentinel·흡수 후 재계산 | L3 |
| AC-CRO-010 | REQ-CRO-008 | RED(행동) | 비카드 → inconclusive·reviewer 호출 0 | L3 |
| AC-CRO-011 | REQ-CRO-009 | RED(행동) | uncommitted 요청 target·cwd | L3 |
| AC-CRO-012 | REQ-CRO-010 | RED(행동) + 대조 **초록** | advisory 필드·영수증 목록·`gate_unmet` 부재 | L3, P2 |
| AC-CRO-013 | REQ-CRO-011 | RED(행동) | CLI JSON 필드·종료 코드 | L3 |
| AC-CRO-014 | REQ-CRO-012 | RED(소스) + 회귀 **초록** | `tools:` 줄 집합 | L5, P1 |
| AC-CRO-015 | REQ-CRO-013 | RED(소스) | 교리 문구 앵커 | L6 |
| AC-CRO-016 | REQ-CRO-014 | RED(소스) | 카탈로그 크기·project_root 문서·i18n | L4, L7 |

REQ→AC 누락 없음: 001→001a·001b · 002→002·003 · 003→004 · 004→005 · 005→006 · 006→007 · 007→008 · 008→009·010 · 009→011 · 010→012 · 011→013 · 012→014 · 013→015 · 014→016.

---

### AC-CRO-001a — `tree_scope` 판독은 `skip` 만 skip 으로 읽는다

**Given** 설정 픽스처(`skip`, 앞뒤 공백이 낀 `Skip`, `SKIP`, `review`, 키 부재, 빈 값, 알 수 없는 값 `never`, 깨진 YAML, 파일 부재)가 주어지고,
**When** 판독기가 각 설정 루트를 읽으면,
**Then** `skip` 계열 셋만 skip 으로, 나머지 전부는 `review` 로 읽힌다. 또 설정 로더(`config.Loader`)가 같은 파일에서 돌려주는 값과 판독기 값이 일치한다(`TestReviewGateReaders_AgreeWithConfigLoader` 계열 핀).

### AC-CRO-001b — 키는 템플릿에 주석 예시로만 있고 라이브 shipped 키·인벤토리는 변하지 않는다

(AC 하위 ID 규약: 한 논리 AC 안의 짝 기준이라 `a`/`b` 접미를 쓴다. Tier M 의 AC 상한 16 은 논리 AC 기준으로 지킨다.)

**Given** 트리 핀 `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b` 의 템플릿 `internal/template/templates/.moai/config/sections/workflow.yaml` 과 `internal/config/testdata/shipped_key_inventory.yaml`,
**When** 구현 뒤 템플릿을 YAML 로 파싱하고 인벤토리를 읽으면,
**Then** (i) 파싱된 템플릿의 `workflow.codex.review_gate` 맵은 키 `enabled`(값 `false`)만 갖고 문서 어디에도 `tree_scope` 라이브 키가 없다 — 변경 전 파싱 결과와 키·값 집합이 같다(주석 줄만 늘었다: 핀 대비 템플릿 diff 의 추가 줄이 전부 `#` 로 시작). (ii) 템플릿 텍스트에는 주석 예시가 있다(`grep -c tree_scope` ≥ 1 — L8 이 0 에서 ≥1 로). (iii) 인벤토리 파일의 해시가 P5 와 같고 `tree_scope` 항목이 없으며(P3 = 0, 대조군 P4 = 1) `internal/settings/schema_sections.go`·`internal/web/fieldsets.templ`·`internal/web/assets/i18n.js` 는 `tree_scope` 를 포함하지 않는다. (iv) `internal/config` 구조체는 `tree_scope` 를 파싱한다(AC-CRO-001a 의 로더 일치 핀이 증명). **RED/GREEN 짝**: RED-now = L8(예시 부재, 마일스톤 M5 가 뒤집음), 회귀 칸 = P3·P5(변경 전 초록, 변경 뒤에도 초록). "바이트 불변"은 인벤토리에 대해서만 문자 그대로 성립한다 — 템플릿은 주석 줄이 늘므로 파싱 내용 동일로 판정한다.

### AC-CRO-002 — 트리 스코프 세션은 skip 에서 리뷰·셀프게이트를 거치지 않는다 (Claude 경로)

**Given** `enabled: true`+`tree_scope: skip` 인 설정 루트, 비 `WT-` 브랜치 트리의 미커밋 변경, 호출 횟수를 세는 `codexLookPath`·리뷰 RPC·셀프게이트 검출기 스텁이 주어지고,
**When** `HandleCodexReviewGate` 가 그 세션의 Stop 입력을 처리하면,
**Then** 결과는 ALLOW 이고 `codexLookPath`·리뷰 RPC·셀프게이트 검출기 호출이 전부 0 이며, stderr 로그에 `scope=tree` 와 skip 사유(키 값 포함)가 한 행으로 남는다. **대조 칸**: 같은 픽스처의 `tree_scope: review` 는 리뷰 RPC 를 정확히 1회 부르고 target `uncommittedChanges`, cwd = 세션 트리다.

### AC-CRO-003 — 같은 결정을 Codex Stop 체인 멤버 6 이 내린다

**Given** AC-CRO-002 와 같은 설정·트리 픽스처,
**When** Codex Stop 체인의 codex 리뷰 멤버가 평가되면,
**Then** 결과는 Allow, 상태 `not-applicable`, 사유에 `tree_scope` 가 들어가고, 영수증을 읽지 않으며(`ReceiptRead` false) 리뷰 RPC·`codexLookPath` 호출이 0 이다. 대조 칸: `review` 는 영수증 판독 경로(`ReceiptRead` true 또는 receipt 부재로 인한 unmeasured 지시)로 간다.

### AC-CRO-004 — `skip` 이 아닐 때 트리 스코프 요청은 변하지 않는다 (회귀)

**Given** `tree_scope` 가 부재이거나 `review` 이고 비카드 트리의 미커밋 변경이 주어지고,
**When** 게이트가 요청을 조립하면,
**Then** 요청은 사전 형태와 shape-identical 이다 — 대상 `uncommittedChanges`, cwd = 해상 트리(SPEC-CODEX-REVIEW-TARGET-001 REQ-CRT-006 직렬화). 변경 전 트리에서 이미 초록인 `TestCodexReviewGate_TreeScopeRequestShapeUnchanged` 가 변경 뒤에도 초록이다.

### AC-CRO-005 — 카드 세션은 `skip` 설정에서도 카드 diff 로 리뷰된다

**Given** 카드 워크트리 픽스처(§B)와 primary 외부 WIP(F), 설정 `enabled: true`+`tree_scope: skip`,
**When** 카드 세션의 턴이 게이트에 닿으면,
**Then** reviewer 가 1회 호출되고 요청 target 은 `baseBranch`, `branch` 값은 호출 시점의 `git merge-base develop HEAD` 출력, cwd 는 카드 워크트리다.

### AC-CRO-006 — 정책 결정은 env·역할을 읽지 않는다

**Given** (a) `develop` 트리+`MOAI_FACTORY_WORKER` 임의 값 X, (b) 같은 트리+임의 값 Y(≠X), (c) 같은 트리+env 부재, (d) `WT-` 트리+env 부재, (e) `WT-` 트리+임의 값 X, 모두 `tree_scope: skip`,
**When** 각각이 게이트에 닿으면,
**Then** (a)(b)(c)는 동일하게 skip, (d)(e)는 동일하게 리뷰된다. 판정이 env 값에 따라 달라지지 않는다는 **양성 관측**(값 둘에서 동일 결과)으로 판정한다 — 코드에 라벨 패턴이 없다는 부재 주장 단독이 아니다.

### AC-CRO-007 — 두 자동 경로와 명시 생산자

**Given** 비카드 트리·`WT-` 트리 × (`skip`, `review`) 네 조합과, 설정 값을 `enabled` 가 true 인 루트 / 그 외 루트에 따로 심은 변형,
**When** Claude 훅 경로와 Codex 체인 멤버 6 이 각각 평가되고, 이어 `produceCodexReviewReceipt` 가 비카드 트리·`skip` 에서 실행되면,
**Then** 네 조합 모두에서 두 경로의 skip/리뷰 결정이 일치하고, 두 경로 각각 `enabled` 를 읽는 바로 그 루트의 `tree_scope` 가 결정에 쓰이며(다른 루트의 값은 영향 없음), `produceCodexReviewReceipt` 는 `skip` 에서도 리뷰 RPC 를 1회 부르고 영수증을 기록한다.

### AC-CRO-008 — 자기 리뷰 도구의 표면

**Given** `moai mcp-server` 인프로세스 클라이언트로 `tools/list` 를 받으면,
**Then** `codex_review`·`glm_review` 가 모두 있고, 각 입력 스키마에서 `scope` 는 enum `{card, uncommitted}` 이며 `required` 에 들어 있고 `project_root` 속성이 선언돼 있다. 두 도구의 읽기 전용 힌트는 true, 카탈로그 선언은 `WriteCapable: false`, 설명에는 "advisory" 가 들어 있다. `codex_audit`·`glm_audit`·`claude_audit`·`audit_multi` 의 입력 스키마(`target` enum 포함)와 출력 스키마는 변경 전과 동일하다.

### AC-CRO-009 — 카드 스코프는 게이트와 같은 해상기가 정하고 기저는 핀되지 않는다

**Given** 카드 워크트리 픽스처(A 커밋, B 미커밋, C 비추적, F 외부 WIP),
**When** `codex_review`·`glm_review` 를 `scope: card`, `project_root: <카드 트리>` 로 부르면,
**Then** (i) codex 요청은 `reviewRequestParams(reviewScopeResolver(<카드 트리>))` 와 `reflect.DeepEqual` 이고 `branch` 값은 호출 시점의 `git merge-base develop HEAD` 출력, cwd 는 카드 트리다. (ii) GLM 으로 보낸 diff 자료에 A 와 B 는 있고 F·C 는 없으며, 결과 `excluded_untracked` 에 C 의 경로가 있다. (iii) 카드가 develop 신규 커밋을 흡수한 뒤 같은 호출의 `branch` 값이 새 merge-base 로 바뀐다(흡수 전 값이 남지 않는다). (iv) `reviewScopeResolver` 를 sentinel MergeBase 를 돌려주는 스텁으로 바꾸면 도구의 요청이 그 sentinel 을 싣는다 — 도구가 두 번째 판별기를 갖지 않는다는 양성 관측이다.

### AC-CRO-010 — 카드 트리가 아니면 다른 리뷰를 대신하지 않는다

**Given** `develop` 브랜치 트리 / detached HEAD / 비 git 디렉터리,
**When** `scope: card` 로 부르면,
**Then** verdict 는 `inconclusive`, 요약에 카드 워크트리가 아니라는 원인(브랜치 이름 또는 해상 실패 사유)이 있고, 리뷰 RPC 호출 0·GLM HTTP 호출 0 이다. 미커밋 변경을 대신 리뷰하지 않는다.

### AC-CRO-011 — `uncommitted` 스코프는 게이트의 트리 요청과 같은 형태다

**Given** `WT-` 카드 트리(커밋분 있음)와 비카드 트리,
**When** `scope: uncommitted` 로 `codex_review` 를 부르면,
**Then** 두 트리 모두 요청이 `{target: uncommittedChanges, cwd: <그 트리>}` 와 DeepEqual 이고(카드 트리에서도 `baseBranch` 요청이 아니다), `glm_review` 의 자료는 `git diff HEAD` 출력이다.

### AC-CRO-012 — 자기 리뷰는 조언이며 영수증·`required` 게이트와 무관하다

**Given** 트리 설정에 `workflow.audit.gates.codex: required`, codex 바이너리 부재(스텁 `codexLookPath` 가 오류), 호출 전 `.moai/state/audit-receipts/` 목록 스냅숏,
**When** `codex_review` 를 부르면,
**Then** 결과는 verdict `inconclusive`, `advisory` true, `gate_unmet` 부재 또는 빈 값이고 `scope`·`base`·`backend`·`tree` 필드가 채워져 있으며, 호출 뒤 영수증 목록이 호출 전과 동일하다. GLM 키 부재(`glmKeyLoader` 스텁 빈 문자열)에서 `glm_review` 도 같다. 정적 단정: `grep -c recordAuditReceipt internal/cli/mcp_selfreview.go` → 0 이고, 같은 명령의 양성 대조군 `internal/cli/mcp_codex.go` 는 2 이다(P2). **대조 칸(회귀 초록)**: 같은 픽스처의 `codex_audit` 는 변경 뒤에도 verdict `fail` 과 비어 있지 않은 `gate_unmet` 를 돌려준다.

### AC-CRO-013 — CLI 거울

**Given** 두 백엔드 × 두 스코프를 스텁 seam 으로 구동할 수 있고, 스텁이 verdict `fail` 을 돌려줄 때,
**When** `moai self-review`(최종 명령 이름은 run 이 확정)를 실행하면,
**Then** 출력 JSON 의 필드 집합이 MCP 도구의 구조화 결과와 같고, 종료 코드는 verdict `fail` 에서도 0 이다. `--scope` 누락과 사용할 수 없는 프로젝트 루트는 비0 이다.

### AC-CRO-014 — 도구 보유는 허용 목록과 정확히 일치한다

**Given** `.claude/agents/moai/*.md`(C1)와 `internal/template/templates/.claude/agents/moai/*.md`(C2)의 모든 에이전트 정의,
**When** `tools:` 줄을 파싱하면,
**Then** `mcp__moai__codex_review`·`mcp__moai__glm_review` 는 manager-develop·manager-docs·manager-lead 에만 들어 있고, `mcp__moai__codex_audit`·`mcp__moai__glm_audit` 는 plan-auditor·sync-auditor 에만 들어 있으며(변경 전과 같은 두 파일 — P1), C1 과 C2 의 `tools:` 집합이 에이전트별로 같다. `make agents-emit-check` 는 종료 코드 0 이다.

### AC-CRO-015 — 교리가 단계와 리더를 규정한다

**Given** `.claude/rules/moai/workflow/kanban-dispatch.md`·`kanban-dispatch-detail.md` 와 C2 미러,
**When** 문구 앵커를 검사하면,
**Then** 스텁에 `card-review` 단계 이름, 증거 경로 `.moai/reports/<card-id>/card-review.md`, "advisory", 리더 세션에 턴 종료 codex 리뷰 게이트가 없다는 문장이 있고, detail 에 단계 절(위치·호출·증거·재리뷰 상한 2·조언성)이 있다. C2 미러에는 SPEC ID·카드 번호·날짜 토큰이 없고 기존 템플릿 중립성 검사(`.github/workflows/template-neutrality-check.yaml`)가 초록이다.

### AC-CRO-016 — 정합 대상이 새 도구를 같게 서술한다

**Given** 두 도구가 등록된 서버,
**Then** `wantCatalogSize` 는 47, `TestMoaiMCPServer_RegistrationMatchesCatalog`·`TestMoaiMCPServer_AnnotationsMatchCatalog`·`TestMoaiMCPTools_WriteCapableSet`(쓰기 가능 집합 불변)이 초록이고, `TestProjectRootDocMatchesServer`·`TestDocsSiteProjectRootMatchesServer` 가 `project_root` 선언 도구 22 개(문서 문장 `Twenty-two`, 로케일별 개수 문구 포함)로 초록이다. `internal/web/assets/i18n.js` 에 두 도구의 `f.mcp.tools.<name>.enabled.title/.desc` 가 4로케일 모두에 있다 — 이 항목을 강제하는 웹 패키지 테스트가 어느 것인지는 본 레인이 특정하지 못했다(Gap): run 이 `grep -c` 로 16항목을 직접 센다.

---

## §E 품질 게이트와 DoD

- **DoD**: §D 전 AC 최종 상태 PASS · 행동 수준 RED 관측 기록이 `.moai/reports/t1422/red/` 에 구현 전 트리 SHA 와 함께 존재 · `go vet`(darwin)+`GOOS=windows GOARCH=amd64 go build ./...` 신규 0 · 수정 파일 커버리지 `quality.yaml` `test_coverage_target`(85) 이상 · `make agents-emit-check`·`make build` 통과 후 갱신 산출물(`catalog.yaml`, `.codex/agents/moai/*.toml`) 커밋 포함 · §D 추적표 상 미매핑 REQ 없음.
- **Tier 상한**: 논리 AC 16 (001a/001b 는 한 논리 AC) — Tier M 상한 16 에 닿아 있다. 폐기 가능한 CLI 거울(Q7)을 빼면 15.
- **간접 검증**: AC-CRO-006 은 양성 귀결(값 둘에서 동일 결과)을 포함한다. AC-CRO-009 (iv)는 "두 번째 판별기가 없다"를 sentinel 의 양성 전파로 검증한다. AC-CRO-012 의 정적 0 단정은 양성 대조군 P2(2)와 함께만 읽는다.
- **선행 폐쇄 게이트**: M1 회귀 칸(AC-CRO-004, AC-CRO-014 의 P1, AC-CRO-012 의 `codex_audit` 대조)이 변경 전 트리에서 초록으로 관측되지 않으면 이후 마일스톤을 시작하지 않는다.
- **측정 규율**: 소관 패키지 단위로 재측정한다(`-run` 이름 패턴 단독 재측정 금지 — 정문 가드를 0개 고르고도 `ok` 로 보인다). 전체 스위트 로컬 금지.
