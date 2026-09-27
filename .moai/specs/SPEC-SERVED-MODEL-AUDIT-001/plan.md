# SPEC-SERVED-MODEL-AUDIT-001 — 구현 계획

---

## §A 맥락

- Tier M · 카드 클래스 C (hook·auditreceipt·config·cli·에이전트 정의 5개 영역에 걸친 설계 변경)
- 개발 방식: `constitution.development_mode` 를 따른다. 신규 동작은 테스트를 먼저 쓴다(RED → GREEN).
- 근거: `.moai/reports/t1282/verdict.md`(착수 판정서) + spec.md §A.2 plan 단계 실측.
- 리드 결정 (c) 는 재론하지 않는다: 관측 2지점, 기본 기록·경고, 감사관 한정 opt-in 채택 거부, 설정 템플릿 OFF·로컬 ON, 감사관 첫 줄 자기 보고.

## §B 결정 사항 (변경 가능성이 높은 순)

### D1. 감사 로그 행 스키마 확장 (데이터 모델)

기존 `agentModelAuditRecord`(`internal/hook/agent_model_guard.go:160`)에 omitempty 필드를 더한다: `source`(`"subagent_stop"` — PreToolUse 행은 필드 부재로 구별), `agent_id`, `served_models`(서빙 집합, 정렬된 배열), `self_reported_model`. 판정 필드 `verdict` 는 새 값 `served_drift`·`unknown` 을 받는다. 기존 PreToolUse 행은 바이트 단위로 그대로다(REQ-SMA-006). 소비자(`prune_logs.go` 의 에이징)는 파일 단위라 영향 없다.

대안(별도 `served-model-audit.jsonl`)은 기각: 같은 spawn 의 선언과 서빙을 한 파일에서 `agent`·`session_id` 로 이을 수 있어야 한다.

### D2. 설정 키 — 형제 키 `workflow.served_model_gate.enabled`

`agent_model_guard` 의 하위 키가 아니라 형제 키로 둔다. 근거: `internal/config/types.go` 의 `SettingsDriftGate` 주석이 밝힌 원칙 — 한 플래그가 서로 다른 표면의 두 거부를 함께 켜면 운영자가 어느 쪽을 끄려 했는지 말할 수 없다. `agent_model_guard.enabled` 는 PreToolUse spawn 거부이고, 이 키는 SubagentStop 감사관 판정 채택 거부다.

- `internal/config/types.go`: `ServedModelGate ServedModelGateConfig \`yaml:"served_model_gate"\`` + `ServedModelGateConfig{Enabled bool}`
- `internal/config/defaults.go`: `Enabled: false`, 기존 가드 패밀리와 같은 중립성 주석
- 템플릿 `internal/template/templates/.moai/config/sections/workflow.yaml`: `served_model_gate: enabled: false` 와 중립 주석(날짜·SPEC ID·카드 ID 금지)
- 로컬 `.moai/config/sections/workflow.yaml`: `served_model_gate: enabled: true`, 주석에 "로컬 opt-in, 템플릿 기본 false" 형태로 근거 기록

SubagentStop 핸들러(`subagentStopHandler`, `subagent_stop.go:23`)는 설정 제공자를 들고 있지 않다. 감사관이 실행된 트리의 설정을 `auditreceipt.rawCodexGate`(`internal/auditreceipt/store.go:166`) 와 같은 원시 YAML 읽기로 판정한다 — 트리는 `auditreceipt.TreeRootFromCWD(input.CWD)`, 설정 고아 워크트리는 `auditreceipt.StoreRoot` 가 가리키는 primary 로 해석(SPEC-WORKTREE-STATE-ROOT-001 과 동일 규칙).

### D3. 채택 거부가 읽히는 경로 — 기존 감사 영수증 저장소와 PreToolUse 소비자 재사용

새 차단 표면을 만들지 않는다. 채택 거부는 기존 경로 그대로 읽힌다:

1. SubagentStop(`subagent_stop.go:38` `Handle`) — 게이트 감사관이고 게이트 ON 이며 서빙 판정이 `served_drift`/`unknown` 이면 `auditreceipt.Rejection` 을 트리 저장소에 기록한다. 원인(`Cause`)은 새 상수(예: `CauseServedModelDrift`, `CauseServedModelUnknown`)로 서빙 집합과 기대 모델을 담는다. SPEC 은 감사관 마지막 줄 판정 줄(`auditreceipt.ParseVerdictLine`, `store.go:423`)에서 읽고, 없으면 `auditreceipt.UnknownSpec`.
2. PreToolUse(`pre_tool.go:724` → `checkAuditReceiptSpawn`, `audit_receipt_guard.go:241`) — 현재 스코프 함수 `auditReceiptScope`(`audit_receipt_guard.go:72`)는 codex 게이트가 `required` 일 때만 참이다. 스코프를 "codex 게이트 required **또는** served_model_gate ON" 으로 넓히되, 목록에서 거부를 셀 때 served 게이트가 OFF 면 served 원인 기록은 무시하고 codex 게이트가 아니면 영수증 원인 기록은 무시한다. 거부 사유 문자열은 원인을 그대로 나열하므로 served 원인이 드러난다(REQ-SMA-013). 새 센티널 `SERVED_MODEL_VIOLATION` 을 사유 앞에 붙여 오케스트레이터가 영수증 거부와 구별하게 한다.
3. 되돌림 없음(REQ-SMA-012) — SubagentStop 은 감사관에게 `Decision: "block"` 을 돌려주지 않는다. 서빙 모델은 같은 인스턴스를 계속 돌려도 바뀌지 않으므로 block 은 루프만 만든다. 기록과 경고(`SystemMessage`)만 한다.

**[주의] 레코드 키 충돌.** `rejectionFileName(agentType, specID)`(`store.go:508`)는 역할+SPEC 으로 파일을 정한다. 같은 SPEC 에 영수증 거부와 서빙 거부가 동시에 있으면 한 파일을 덮는다. 서빙 거부는 키 공간을 분리해야 한다(예: 레코드에 `Kind` 필드를 더하고 파일명에 kind 접미). 이 분리가 REQ-SMA-014 의 원인별 해제의 전제다.

**[주의] 기존 해제 경로.** `ClearRejectionsForRoleInTree`(`store.go:394`)는 역할 단위로 전부 지운다 — 영수증 증명 PASS 가 서빙 거부까지 지우게 된다. 해제를 kind 로 한정하는 변형을 쓰거나 인자를 추가한다.

### D4. 사후 스캔 표면 — `moai doctor` 점검 `"Served Model"`

`internal/cli/doctor.go:201` `moaiChecks` 에 `servedModelCheckName`(상수) 점검을 등록한다. 읽기 전용, 상태는 발견 시 warn, 발견이 없으면 ok, 트랜스크립트 디렉터리가 없으면 ok + "관측 대상 없음" 메시지(오류 아님). doctor 종료 코드를 바꾸지 않는다(REQ-SMA-017).

- 트랜스크립트 루트: `internal/escalation/roots.go:21` `MemorySlug` 와 같은 slug 규칙, 기준 디렉터리는 `$CLAUDE_CONFIG_DIR` 과 `~/.claude`(같은 파일 35-58행 `memoryRoots` 의 기준 집합), 키는 현재 트리와 primary 체크아웃. `memory` 하위 대신 `*/subagents/agent-*.jsonl` 을 본다. `memoryRoots` 가 비공개이므로 기준·키 계산을 공유 함수로 끌어내거나 동일 규칙을 재사용한다 — 규칙을 손으로 복제하지 않는다.
- 분류 로직은 SubagentStop 과 **한 구현**을 공유한다(hook 패키지의 export 함수 또는 별도 소패키지). 두 곳에서 따로 판정하면 REQ-SMA-003 이 두 벌이 된다.
- **[HARD] 점검 이름을 `internal/cli/binary_lag_test.go:198` `namesAddedAfterBaseline` 에 `"servedModelCheckName": true` 로 등록한다.** 누락 시 `TestBinaryLag_DoctorCheckNameSetIsUnchanged` 가 CI 에서 붉어진다(메모리 교훈 t1251).

### D5. 기대 모델 대조 규칙

기대 모델 = meta.json `model`(선언) 이 비어 있지 않으면 그것, 아니면 `resolveAgentModel`(`agent_model_guard.go:113`) 결과. 둘 다 없으면 `unmapped`. 일치 판정: 대소문자 무시 동일, 또는 기대가 순수 계열 별칭(`opus`·`sonnet`·`haiku` 등 — 해석기가 내는 별칭 집합을 기준으로 하되 목록을 손으로 재기술하지 않는다)이고 서빙 식별자가 하이픈 구분 토큰으로 그 계열 토큰을 포함하는 Claude 모델 식별자일 때. 실측 값 `claude-opus-5-5`·`claude-opus-5` 는 `opus` 와 일치, `glm-5.3-flash`·`glm-5.3` 은 불일치.

### D6. 트랜스크립트 읽기 예산

한 줄씩 스트리밍으로 읽고, 줄 수·바이트 상한(구현 상수, 테스트에서 주입 가능)을 넘으면 `unknown`. 판정서 세션 `d46e0166` 의 3.4MB 트랜스크립트는 예산 안에 들어와야 한다(M6 에서 실측).

## §C 사전 점검 (run 착수 전)

1. `git -C .claude/worktrees/t1282 rev-list --count --left-right develop...HEAD` 로 기준 확인.
2. `go list -deps ./internal/auditreceipt/... | grep moai-adk/internal/hook` 출력 없음 확인 — hook → auditreceipt 방향만 존재해야 한다(순환 방지).
3. `go test ./internal/hook/... ./internal/auditreceipt/... ./internal/config/... -count=1` 기준선 기록.
4. `golangci-lint --version` 이 CI 판 v2.1.6 인지 확인(메모리 교훈 — 로컬판이 staticcheck 를 놓쳐 CI 적색 2회).

## §D 제약

- 템플릿 퍼스트: 템플릿 먼저, `make build`, 로컬 미러.
- C3(`internal/template/templates/.codex/agents/moai/*.toml`) 손편집 금지 — `make agents-emit` 으로만 재생성.
- `.claude/rules/**` 편집은 세션 프리픽스 캐시를 깬다 — 문서 갱신은 마일스톤 마지막에 모은다.
- 전체 스위트 로컬 실행 금지 — 영향 패키지만, 전 패키지 판정은 CI.

## §E 자기 검증 (run-phase 종료 조건)

```bash
go test ./internal/hook/... -count=1
go test ./internal/auditreceipt/... ./internal/config/... -count=1
go test ./internal/cli/ -run '^TestBinaryLag_.*$' -count=1 -v
go test ./internal/cli/ -run '^TestServedModelCheck_.*$' -count=1 -v
go test ./internal/template/... -count=1
make agents-emit-check
golangci-lint run ./internal/hook/... ./internal/auditreceipt/... ./internal/config/... ./internal/cli/...
```

acceptance.md 의 AC 별 명령이 권위다. 위 목록은 묶음 실행용 요약이다.

## §F 마일스톤 (결정 변경 가능성이 높은 순서, 시간 추정 없음)

### M1 — 판정 코어 (Priority High)

- 서빙 집합 추출(REQ-SMA-001), 기대 모델 대조(D5), 판정 4값(REQ-SMA-003/004), 읽기 예산(REQ-SMA-008)을 순수 함수로 구현.
- 테스트 먼저: `TestServedModel_MatchIsOK`, `TestServedModel_DeclaredOpusServedGLMIsDrift`, `TestServedModel_MixedServedIsDrift`, `TestServedModel_MissingTranscriptIsUnknown`, `TestServedModel_EmptyModelFieldIsUnknown`, `TestServedModel_SyntheticOnlyIsUnknown`, `TestServedModel_ReadBudgetExceededIsUnknown`, `TestServedModel_UnmappedWithoutExpectation`.
- 픽스처는 `internal/hook/testdata/` 에 최소 JSONL 로 둔다. 실제 트랜스크립트 본문(프롬프트)은 복사하지 않는다.

### M2 — 설정 키 (Priority High)

- D2 대로 types/defaults/템플릿/로컬 4곳. 테스트 `TestServedModelGate_DefaultOff`(internal/config).

### M3 — SubagentStop 관측 배선 (Priority High)

- `subagent_stop.go:38` `Handle` 에 관측 추가(REQ-SMA-002/005/006/007/019). 경고는 `SystemMessage` 로, 기존 `mergeAuditorStopGuard`(53행)가 guard 의 SystemMessage 로 덮어쓰는 동작과 충돌하지 않게 이어붙인다.
- 테스트: `TestServedModel_SubagentStopAppendsServedRecord`, `TestServedModel_TranscriptPathFallback`, `TestServedModel_GateOffWarnsWithoutRejection`, `TestServedModel_ObservationFailOpen`, `TestServedModel_RecordsSelfReportedModel`.

### M4 — 채택 거부와 소비 (Priority High)

- D3 전부: kind 분리, 원인 상수, SubagentStop 기록(REQ-SMA-011), PreToolUse 소비 확장(REQ-SMA-013), 원인별 해제(REQ-SMA-014), 비감사관 제외(REQ-SMA-015), 되돌림 없음(REQ-SMA-012).
- 테스트: `TestServedModel_GateOnRejectsAuditorDrift`, `TestServedModel_GateOnRejectsAuditorUnknown`, `TestServedModel_RejectionDeniesPhaseEntrySpawn`, `TestServedModel_NonAuditorNeverRejected`, `TestServedModel_RejectionIsAdoptionRefusalOnly`, `TestServedModel_ClearingIsKindScoped`.
- 기존 영수증 가드 회귀: `go test ./internal/hook/ -run '^TestAudit.*$' -count=1`(이름 집합은 run 착수 시 `go test -list` 로 확정) 와 `go test ./internal/auditreceipt/... -count=1`.

### M5 — 사후 스캔 doctor 점검 (Priority Medium)

- D4 전부. 테스트 `TestServedModelCheck_ReportsDriftAndUnknown`, `TestServedModelCheck_ReadOnly`, `TestServedModelCheck_NoTranscriptsIsOK` (internal/cli).
- `namesAddedAfterBaseline` 등록 후 `go test ./internal/cli/ -run '^TestBinaryLag_.*$' -count=1 -v`.

### M6 — 감사관 자기 보고 첫 줄 (Priority Medium)

- C2 `internal/template/templates/.claude/agents/moai/{plan,sync}-auditor.md` 의 Output Format 절에 첫 줄 `auditor-model: <served model>` 요구 추가 → C1 `.claude/agents/moai/{plan,sync}-auditor.md` 동일 반영(C1↔C2 는 바이트 동일 관계가 아님 — 해당 절만 맞춘다).
- **`make agents-emit` 실행**으로 C3 재생성 → **`make agents-emit-check` exit 0** 확인. C3 손편집 금지.
- 문구 중립성: SPEC ID·카드 ID·날짜 없음. `go test ./internal/template/... -count=1`(중립성·내부 누출 가드 포함).
- 종단 실측: 판정서 재현 세션 `d46e0166` 을 대상으로 `go run ./cmd/moai doctor --check "Served Model" --verbose` 가 `served_drift` 를 보고하는지 확인(개발 머신 한정, AC-SMA-020).

### M7 — 문서·정리 (Priority Low)

- `.claude/rules/moai/core/agent-common-protocol.md` § Per-Spawn Model Injection 의 감사 로그 문장에 "SubagentStop 이 서빙 모델을 같은 로그에 기록한다" 한 문장 추가 — 로컬·템플릿 양쪽.
- `make build` 후 영향 패키지 재측정, golangci-lint v2.1.6.

## §G 위험

| # | 위험 | 대응 |
|---|------|------|
| R1 | SubagentStop 페이로드에 `agent_transcript_path` 가 없을 수 있음(§A.4) | REQ-SMA-007 파생 경로. 둘 다 실패하면 `unknown` — 조용한 `ok` 없음 |
| R2 | 발화 시점에 마지막 assistant 행이 아직 안 쓰였을 수 있음 | 집합 판정이라 대부분 무영향. 행 0건이면 `unknown`. M3 에서 실세션 1회 관측해 progress.md 에 기록 |
| R3 | 5초 훅 timeout 초과 | D6 예산. 예산 초과는 `unknown` |
| R4 | GLM 백엔드 사용자에게 매 서브에이전트마다 경고 | 기본 OFF 라 거부는 없음. 경고 소음은 Out of Scope 로 남기고 sync 에서 후속 카드 후보로 보고 |
| R5 | 영수증 거부와 서빙 거부의 파일 키 충돌 / 해제 교차 | D3 의 kind 분리와 kind 한정 해제. `TestServedModel_ClearingIsKindScoped` 가 판정 |
| R6 | `moai update` 가 `.moai/config` 를 재배포하며 로컬 `enabled: true` 를 잃을 가능성 | 기존 로컬 가드 키와 같은 운명. sync 에서 update 후 키 생존 여부를 확인하고 보고 |
| R7 | 판정 로직 이중 구현(hook vs doctor) | D4 — 한 구현 공유 |

## §H 교차 참조

- `.moai/reports/t1282/verdict.md` — 착수 판정서(재현 근거)
- SPEC-AGENT-MODEL-ENFORCE-001 — PreToolUse 선언 모델 감사의 기원
- SPEC-CODEX-AUDIT-GATE-AXES-001 — 감사관 판정 채택 거부와 phase 진입 차단의 기원
- SPEC-WORKTREE-STATE-ROOT-001 — 설정 고아 워크트리의 저장소 해석
