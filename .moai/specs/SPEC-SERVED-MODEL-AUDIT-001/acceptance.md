# SPEC-SERVED-MODEL-AUDIT-001 — 인수 기준

---

## §A 판정 규약

1. 모든 명령은 워크트리 루트(`.claude/worktrees/t1282`)에서 실행한다.
2. `go test` 기반 AC 는 종료 코드 0 **그리고** `-v` 출력에 `--- PASS: <테스트 이름> ` (이름 뒤 공백 포함) 줄이 있을 때만 PASS 다. 테스트가 선택되지 않아 `testing: warning: no tests to run` 이 나오면 FAIL 이다.
3. 모든 `-run` 패턴은 `^…$` 로 고정한다.
4. 부정형 단언(“기록되지 않는다”, “거부되지 않는다”)을 담는 테스트는 같은 테스트 안에서 양성 대조(같은 픽스처로 게이트 ON 또는 감사관일 때는 기록됨)를 함께 단언한다. 0건이 검사기 고장이 아님을 보이기 위함이다.
5. 린트는 CI 판 golangci-lint v2.1.6 으로 잰다.

---

## §B AC 매트릭스

| AC | REQ | 요지 |
|----|-----|------|
| AC-SMA-001 | REQ-SMA-001, 003 | 선언 opus · 서빙 claude-opus 계열 → `ok` |
| AC-SMA-002 | REQ-SMA-003 | 선언 opus · 서빙 glm-5.3-flash → `served_drift` |
| AC-SMA-003 | REQ-SMA-003 | 서빙 집합에 기대 계열과 비계열 혼재 → `served_drift` |
| AC-SMA-004 | REQ-SMA-004 | 트랜스크립트 부재 → `unknown` (절대 `ok` 아님) |
| AC-SMA-005 | REQ-SMA-001, 004 | model 필드 빈 값·부재 → `unknown` |
| AC-SMA-006 | REQ-SMA-001, 004 | `<synthetic>` 행만 존재 → `unknown` |
| AC-SMA-007 | REQ-SMA-003 | 선언·해석 모두 없음 → `unmapped` |
| AC-SMA-008 | REQ-SMA-008 | 읽기 예산 초과 → `unknown` |
| AC-SMA-009 | REQ-SMA-002, 006 | SubagentStop 이 서빙 행 1개 추가, 기존 행 불변 |
| AC-SMA-010 | REQ-SMA-007 | `agent_transcript_path` 부재 시 파생 경로로 판정 |
| AC-SMA-011 | REQ-SMA-005, 010 | 게이트 OFF → 경고만, 거부 기록 없음, spawn 허용 |
| AC-SMA-012 | REQ-SMA-006 | 관측 실패 시 fail-open, 훅 오류 없음 |
| AC-SMA-013 | REQ-SMA-011 | 게이트 ON + 감사관 `served_drift` → 채택 거부 기록 |
| AC-SMA-014 | REQ-SMA-011 | 게이트 ON + 감사관 `unknown` → 채택 거부 기록 |
| AC-SMA-015 | REQ-SMA-013 | 거부 기록 존재 시 phase 진입 spawn 거부 |
| AC-SMA-016 | REQ-SMA-015 | 게이트 ON 이어도 비감사관은 거부 없음 |
| AC-SMA-017 | REQ-SMA-012 | 채택 거부는 block 결정 없음·보고서 불변 |
| AC-SMA-018 | REQ-SMA-014 | 해제는 kind 한정 |
| AC-SMA-019 | REQ-SMA-009 | 설정 기본값 OFF · 템플릿 false · 로컬 true |
| AC-SMA-020 | REQ-SMA-016, 017 | doctor 점검이 drift/unknown 을 보고, 읽기 전용, 종료 코드 불변 |
| AC-SMA-021 | REQ-SMA-016 | 새 doctor 점검 이름이 binary_lag 허용 목록에 등록 |
| AC-SMA-022 | REQ-SMA-018 | 감사관 정의 C1·C2·C3 가 첫 줄 `auditor-model:` 요구 |
| AC-SMA-023 | REQ-SMA-019 | 자기 보고는 기록되되 판정을 바꾸지 않음 |
| AC-SMA-024 | REQ-SMA-006, 013 | 기존 hook·auditreceipt·config 패키지 회귀 없음, 린트 0 |
| AC-SMA-025 | REQ-SMA-003, 016 | 재현 세션 d46e0166 을 실측으로 `served_drift` 판정 |

---

## §C 시나리오

### AC-SMA-001 — 일치 → ok

- **Given** meta.json `model: "opus"` 와 assistant 행 전부 `claude-opus-5-5` 인 트랜스크립트 픽스처
- **When** 서빙 판정을 수행하면
- **Then** 판정은 `ok` 이고 서빙 집합은 `["claude-opus-5-5"]` 이다.

```bash
go test ./internal/hook/ -run '^TestServedModel_MatchIsOK$' -count=1 -v
```

### AC-SMA-002 — 선언 opus · 서빙 GLM → served_drift

- **Given** meta.json `model: "opus"` 와 assistant 행 전부 `glm-5.3-flash` 인 픽스처(판정서 재현 모양)
- **When** 서빙 판정을 수행하면
- **Then** 판정은 `served_drift` 이며 `ok` 가 아니다.

```bash
go test ./internal/hook/ -run '^TestServedModel_DeclaredOpusServedGLMIsDrift$' -count=1 -v
```

### AC-SMA-003 — 혼재 → served_drift

- **Given** assistant 행이 `claude-opus-5` 와 `glm-5.3` 을 섞어 담은 픽스처, 기대 `opus`
- **When** 서빙 판정을 수행하면
- **Then** 판정은 `served_drift` 이고 서빙 집합에 두 값이 모두 있다.

```bash
go test ./internal/hook/ -run '^TestServedModel_MixedServedIsDrift$' -count=1 -v
```

### AC-SMA-004 — 트랜스크립트 부재 → unknown

- **Given** 존재하지 않는 트랜스크립트 경로와, 같은 테스트 안의 양성 대조로 읽을 수 있는 일치 픽스처
- **When** 두 경로 각각에 서빙 판정을 수행하면
- **Then** 부재 경로는 `unknown`, 양성 대조는 `ok` 이다.

```bash
go test ./internal/hook/ -run '^TestServedModel_MissingTranscriptIsUnknown$' -count=1 -v
```

### AC-SMA-005 — model 필드 빈 값·부재 → unknown

- **Given** assistant 행의 `message.model` 이 `""` 인 픽스처와, 필드 자체가 없는 픽스처
- **When** 각각 판정하면
- **Then** 둘 다 `unknown` 이다.

```bash
go test ./internal/hook/ -run '^TestServedModel_EmptyModelFieldIsUnknown$' -count=1 -v
```

### AC-SMA-006 — synthetic 전용 → unknown

- **Given** assistant 행이 전부 `<synthetic>` 인 픽스처, 그리고 `<synthetic>` 1행과 `claude-opus-5-5` 1행이 섞인 픽스처
- **When** 각각 판정하면
- **Then** 앞의 것은 `unknown`, 뒤의 것은 `ok` 이며 서빙 집합에 `<synthetic>` 이 없다.

```bash
go test ./internal/hook/ -run '^TestServedModel_SyntheticOnlyIsUnknown$' -count=1 -v
```

### AC-SMA-007 — 비교 대상 없음 → unmapped

- **Given** meta.json 에 model 이 없고 에이전트 유형이 프로필에 매핑되지 않은 입력
- **When** 판정하면
- **Then** 판정은 `unmapped` 이고 서빙 집합은 기록된다.

```bash
go test ./internal/hook/ -run '^TestServedModel_UnmappedWithoutExpectation$' -count=1 -v
```

### AC-SMA-008 — 읽기 예산 초과 → unknown

- **Given** 테스트가 주입한 작은 읽기 예산과 그 예산을 넘는 크기의 일치 픽스처
- **When** 판정하면
- **Then** 판정은 `unknown` 이고, 같은 픽스처를 기본 예산으로 판정하면 `ok` 다(양성 대조).

```bash
go test ./internal/hook/ -run '^TestServedModel_ReadBudgetExceededIsUnknown$' -count=1 -v
```

### AC-SMA-009 — SubagentStop 이 서빙 행 1개를 추가하고 기존 행은 그대로

- **Given** 임시 프로젝트 루트의 `.moai/logs/agent-model-audit.jsonl` 에 PreToolUse 행 1개가 있고, SubagentStop 입력이 drift 픽스처를 가리킬 때
- **When** SubagentStop 핸들러가 실행되면
- **Then** 파일은 정확히 2행이고, 1행은 원래 바이트와 같으며, 2행은 `source`, `agent_id`, `agent`, `declared_model`, `resolved_model`, `served_models`, `verdict: "served_drift"` 를 담는다.

```bash
go test ./internal/hook/ -run '^TestServedModel_SubagentStopAppendsServedRecord$' -count=1 -v
```

### AC-SMA-010 — 트랜스크립트 경로 파생

- **Given** `agent_transcript_path` 가 비어 있고 `transcript_path=<dir>/<session>.jsonl` 과 `agent_id` 가 주어지며 `<dir>/<session>/subagents/agent-<id>.jsonl` 에 drift 픽스처가 있을 때
- **When** SubagentStop 이 실행되면
- **Then** 기록된 판정은 `served_drift` 다(파생 경로가 실제로 읽혔다는 증거).

```bash
go test ./internal/hook/ -run '^TestServedModel_TranscriptPathFallback$' -count=1 -v
```

### AC-SMA-011 — 게이트 OFF → 경고만

- **Given** `workflow.served_model_gate.enabled` 가 없거나 false 인 트리에서 plan-auditor 가 drift 픽스처로 멈출 때
- **When** SubagentStop 과 이어지는 manager-develop spawn 의 PreToolUse 가 실행되면
- **Then** SubagentStop 출력은 `SystemMessage` 에 에이전트 유형·기대 모델·서빙 집합을 담고 `Decision` 은 비어 있으며, 저장소에 서빙 거부 기록이 없고, spawn 은 거부되지 않는다. 같은 테스트에서 게이트를 켜면 기록이 생긴다(양성 대조).

```bash
go test ./internal/hook/ -run '^TestServedModel_GateOffWarnsWithoutRejection$' -count=1 -v
```

### AC-SMA-012 — fail-open

- **Given** 로그 디렉터리를 쓸 수 없는 루트, 손상된 JSON 행이 섞인 트랜스크립트, 빈 CWD 입력
- **When** SubagentStop 핸들러가 실행되면
- **Then** 핸들러는 오류를 돌려주지 않고 패닉하지 않으며, 판정 가능한 경우의 판정은 `ok` 가 아닌 값(`unknown`)으로 남는다.

```bash
go test ./internal/hook/ -run '^TestServedModel_ObservationFailOpen$' -count=1 -v
```

### AC-SMA-013 — 게이트 ON + 감사관 drift → 채택 거부

- **Given** 게이트 ON 트리에서 sync-auditor 가 `AUDIT-VERDICT: PASS spec=SPEC-X-001 receipts=none` 로 끝나는 메시지와 drift 픽스처로 멈출 때
- **When** SubagentStop 이 실행되면
- **Then** 그 트리 저장소에 역할 `sync-auditor`, SPEC `SPEC-X-001`, 서빙 원인을 담은 거부 기록이 1건 생긴다.

```bash
go test ./internal/hook/ -run '^TestServedModel_GateOnRejectsAuditorDrift$' -count=1 -v
```

### AC-SMA-014 — 게이트 ON + 감사관 unknown → 채택 거부

- **Given** 게이트 ON 트리에서 plan-auditor 의 트랜스크립트가 부재할 때
- **When** SubagentStop 이 실행되면
- **Then** 서빙 원인이 `unknown` 을 가리키는 거부 기록이 생긴다.

```bash
go test ./internal/hook/ -run '^TestServedModel_GateOnRejectsAuditorUnknown$' -count=1 -v
```

### AC-SMA-015 — 거부가 phase 진입 spawn 을 막는다

- **Given** AC-SMA-013 의 거부 기록이 남아 있는 트리
- **When** 그 트리에서 `manager-develop`, `manager-docs`, `manager-git` spawn 의 PreToolUse 가 각각 실행되고, 대조로 `Explore` spawn 도 실행되면
- **Then** 세 phase 진입 spawn 은 `SERVED_MODEL_VIOLATION` 으로 시작하는 사유로 거부되고 사유에 역할·SPEC·서빙 원인이 있으며, `Explore` 는 거부되지 않는다.

```bash
go test ./internal/hook/ -run '^TestServedModel_RejectionDeniesPhaseEntrySpawn$' -count=1 -v
```

### AC-SMA-016 — 비감사관은 거부 없음

- **Given** 게이트 ON 트리에서 manager-develop 이 drift 픽스처로 멈출 때
- **When** SubagentStop 과 이어지는 manager-docs spawn 이 실행되면
- **Then** 감사 로그에는 `served_drift` 행과 경고가 있으나 거부 기록은 없고 spawn 은 허용된다. 같은 픽스처로 plan-auditor 가 멈추면 기록이 생긴다(양성 대조).

```bash
go test ./internal/hook/ -run '^TestServedModel_NonAuditorNeverRejected$' -count=1 -v
```

### AC-SMA-017 — 채택 거부만, 되돌림 없음

- **Given** 게이트 ON 트리, 감사관 보고서 파일 1개(`.moai/reports/...`)와 SPEC 디렉터리가 있는 임시 트리
- **When** drift 감사관의 SubagentStop 이 실행되면
- **Then** 출력의 `Decision` 은 비어 있고(`block` 아님), 보고서 파일·SPEC 파일의 바이트는 실행 전과 같으며, 새로 생긴 파일은 감사 로그 행과 거부 기록뿐이다.

```bash
go test ./internal/hook/ -run '^TestServedModel_RejectionIsAdoptionRefusalOnly$' -count=1 -v
```

### AC-SMA-018 — 해제는 kind 한정

- **Given** 같은 역할·같은 SPEC 에 영수증 거부 1건과 서빙 거부 1건이 함께 있는 저장소
- **When** (a) 같은 역할의 감사관이 서빙 `ok`·영수증 미증명으로 멈추면, (b) 따로 영수증 증명 PASS·서빙 drift 로 멈추면
- **Then** (a) 에서는 서빙 거부만 사라지고 영수증 거부는 남으며, (b) 에서는 영수증 거부만 사라지고 서빙 거부는 남는다. 두 기록은 서로를 덮어쓰지 않는다.

```bash
go test ./internal/hook/ -run '^TestServedModel_ClearingIsKindScoped$' -count=1 -v
```

### AC-SMA-019 — 설정 기본값

- **Given** 설정 파일이 없는 트리
- **When** 기본 설정을 적재하면
- **Then** `Workflow.ServedModelGate.Enabled` 는 false 다.

```bash
go test ./internal/config/ -run '^TestServedModelGate_DefaultOff$' -count=1 -v
grep -n -A1 'served_model_gate:' internal/template/templates/.moai/config/sections/workflow.yaml
grep -n -A1 'served_model_gate:' .moai/config/sections/workflow.yaml
go test ./internal/template/... -count=1
```

- 두 번째 명령 출력의 다음 줄은 `enabled: false`, 세 번째는 `enabled: true` 여야 한다. 네 번째(템플릿 중립성 가드 포함)는 exit 0.

### AC-SMA-020 — doctor 사후 스캔

- **Given** 가짜 설정 기준 디렉터리 아래 drift 1건·unknown 1건·ok 1건의 감사관 트랜스크립트가 있는 픽스처 트리
- **When** 점검 함수를 실행하면
- **Then** 상태는 warn, 메시지에 verdict 별 개수(ok 1 / served_drift 1 / unknown 1)와 두 감사관 실행이 나열되며, 실행 전후 픽스처 트리와 감사 로그의 바이트가 같다. 트랜스크립트 디렉터리가 없으면 상태는 ok 다.

```bash
go test ./internal/cli/ -run '^TestServedModelCheck_ReportsDriftAndUnknown$' -count=1 -v
go test ./internal/cli/ -run '^TestServedModelCheck_ReadOnly$' -count=1 -v
go test ./internal/cli/ -run '^TestServedModelCheck_NoTranscriptsIsOK$' -count=1 -v
```

### AC-SMA-021 — binary_lag 허용 목록 등록

- **Given** `moaiChecks` 에 새 점검이 등록된 트리
- **When** binary_lag 가드 계열을 실행하면
- **Then** 전부 통과하고, `namesAddedAfterBaseline` 에 `servedModelCheckName` 키가 있다.

```bash
go test ./internal/cli/ -run '^TestBinaryLag_.*$' -count=1 -v
grep -n '"servedModelCheckName"' internal/cli/binary_lag_test.go
```

### AC-SMA-022 — 감사관 정의의 첫 줄 요구

- **Given** 편집된 C2 원본과 C1 로컬 사본
- **When** 방출 일치 검사와 문자열 검사를 실행하면
- **Then** `make agents-emit-check` 가 exit 0 이고, 아래 6개 파일 각각에 `auditor-model:` 이 1회 이상 나오며, C2 두 파일에는 SPEC ID·카드 ID 가 없다.

```bash
make agents-emit-check
grep -c 'auditor-model:' internal/template/templates/.claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md internal/template/templates/.codex/agents/moai/plan-auditor.toml internal/template/templates/.codex/agents/moai/sync-auditor.toml
go test ./internal/template/agentemit/... -count=1
```

### AC-SMA-023 — 자기 보고는 기록만

- **Given** 첫 줄이 `auditor-model: opus` 인 감사관 메시지와 glm 서빙 픽스처
- **When** SubagentStop 이 실행되면
- **Then** 기록 행의 `self_reported_model` 은 `opus`, `verdict` 는 `served_drift` 다.

```bash
go test ./internal/hook/ -run '^TestServedModel_RecordsSelfReportedModel$' -count=1 -v
```

### AC-SMA-024 — 회귀 없음

```bash
go test ./internal/hook/... -count=1
go test ./internal/auditreceipt/... ./internal/config/... -count=1
golangci-lint run ./internal/hook/... ./internal/auditreceipt/... ./internal/config/... ./internal/cli/...
```

- 모두 exit 0, 린트 신규 이슈 0(v2.1.6).

### AC-SMA-025 — 재현 세션 실측 (개발 머신 한정)

- **Given** 착수 판정서의 세션 `d46e0166` 트랜스크립트가 이 머신의 프로필 디렉터리에 있음
- **When** 새로 빌드한 바이너리로 doctor 점검을 실행하면
- **Then** 출력에 manager-develop 한 건이 `served_drift`(서빙 `glm-5.3-flash`)로 나타난다. 판정서 `agent-model-audit.jsonl` 의 같은 세션 행은 여전히 `ok` 로 남아 있다(소급 기록 없음).

```bash
go run ./cmd/moai doctor --check "Served Model" --verbose
```

- 결과는 progress.md §E.2 에 명령과 출력 원문으로 남긴다. 이 AC 는 해당 트랜스크립트가 있는 머신에서만 판정 가능하며, 없으면 Gap 으로 보고한다.

---

## §D 품질 게이트

- TRUST 5: 신규 코드 패키지 커버리지 85% 이상(`go test -cover ./internal/hook/`), gofmt 무변경, 린트 0.
- 템플릿 중립성 가드와 `make agents-emit-check` 통과.
- @MX: 새 SubagentStop 관측 진입점과 kind 분리 저장소 함수에 `@MX:ANCHOR` 검토(fan_in 확인), 트랜스크립트 스트리밍 읽기에 `@MX:NOTE`.

## §E Definition of Done

- AC-SMA-001..024 전부 PASS, AC-SMA-025 는 PASS 또는 명시적 Gap.
- progress.md §E.2 에 각 AC 의 명령·출력 원문과 측정 트리 HEAD 가 기록됨.
- C3 는 `make agents-emit` 으로만 바뀌었음(손편집 흔적 없음 — `make agents-emit-check` exit 0).
