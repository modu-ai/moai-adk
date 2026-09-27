# SPEC-SERVED-MODEL-AUDIT-001 — 구현 계획

---

## §A 맥락

- Tier M · 카드 클래스 C (hook·auditreceipt·config·cli 4개 패키지와 에이전트 정의 미러)
- REQ 16 · AC 16 (Tier M 예산 한도 안). 파일 수 대역 초과에 대한 판단은 spec.md §D "Tier 판정".
- 개발 방식: `constitution.development_mode` 를 따른다. 신규 동작은 테스트를 먼저 쓴다(RED → GREEN).
- 근거: `.moai/reports/t1282/verdict.md`(착수 판정서) + spec.md §A.2 plan 단계 실측.
- 리드 결정 (c) 는 재론하지 않는다: 관측 2지점, 기본 기록·경고, 감사관 한정 opt-in 채택 거부, 설정 템플릿 OFF·로컬 ON, 감사관 첫 줄 자기 보고.
- plan-audit iter1 FAIL 0.75 반영본(v0.2.0). 결함별 반영 위치는 progress.md §E.1.

## §B 결정 사항 (변경 가능성이 높은 순)

### D1. 감사 로그 행 스키마 확장 (데이터 모델)

기존 `agentModelAuditRecord`(`internal/hook/agent_model_guard.go:160`)에 omitempty 필드를 더한다: `source`(`"subagent_stop"` — PreToolUse 행은 필드 부재로 구별), `agent_id`, `served_models`(서빙 집합, 정렬된 배열), `self_reported_model`. `verdict` 는 새 값 `served_drift`·`unknown` 을 받는다. 기존 PreToolUse 행은 바이트 단위로 그대로다(REQ-SMA-006).

대안(별도 `served-model-audit.jsonl`)은 기각: 같은 spawn 의 선언과 서빙을 한 파일에서 `agent`·`session_id` 로 이을 수 있어야 한다.

### D2. 거부 레코드의 kind 분리 (데이터 모델)

`auditreceipt.Rejection`(`internal/auditreceipt/store.go:123`)에 `Kind string \`json:"kind,omitempty"\`` 를 더한다. 값은 `served` 만 새로 쓴다. **`Kind` 가 비어 있는 레코드는 `receipt` 로 해석한다**(REQ-SMA-011) — 오늘 영수증 가드가 쓰는 모든 기존 레코드가 여기에 해당하므로 마이그레이션이 필요 없다.

- 파일명: `rejectionFileName(agentType, specID)`(`store.go:508`)는 역할+SPEC 으로만 파일을 정해 두 종류가 서로 덮어쓴다. served-kind 는 다른 파일명(예: `<role>--<spec>--served.json`)을 쓰고, receipt-kind 는 기존 파일명을 그대로 쓴다(레거시 호환).
- 해제: `ClearRejectionsForRoleInTree`(`store.go:394`)는 `<role>--` 파일명 접두로 전부 지운다(403-405행). served 파일도 이 접두에 걸리므로, kind 인자를 받는 해제 함수를 두고 영수증 경로는 receipt(= Kind 비어 있음 포함)만, 서빙 경로는 served 만 지우게 한다. 기존 함수의 호출부(`audit_receipt_guard.go:153`)는 receipt 한정 변형으로 바꾼다.
- 목록: `ListRejectionsForTree`(`store.go:377`)는 두 종류를 모두 돌려주고, 소비자가 kind 로 거른다.

### D3. 서빙 게이트 전용 스코프 — 영수증 가드 스코프는 건드리지 않는다

`auditReceiptScope`(`audit_receipt_guard.go:72`)는 `recordAuditorStart`(:98)·`checkAuditorStop`(:120)·`checkAuditReceiptSpawn`(:246) 세 곳이 공유한다. 이 술어를 넓히면 이 저장소처럼 codex 게이트를 선언하지 않은 트리에서 서빙 게이트만 켜도 영수증 가드가 무장해 모든 감사관 PASS 가 `block` 된다(iter1 D2). 따라서 **넓히지 않는다**(REQ-SMA-013).

- 새 술어 `servedGateScope(input) (guardTree, bool)`: 트리는 `auditreceipt.TreeRootFromCWD(input.CWD)`, 저장소는 `auditreceipt.StoreRoot(tree)`(설정 고아 워크트리는 primary — SPEC-WORKTREE-STATE-ROOT-001 과 동일), 활성 조건은 그 트리 설정의 `workflow.served_model_gate.enabled == true` 하나뿐. 설정은 `auditreceipt.rawCodexGate`(`store.go:166`)와 같은 원시 YAML 읽기로 판정한다(`ServedGateEnabled(treeRoot)` 신설). 저장소를 식별할 수 없으면 서빙 거부는 기록하지 않고 경고만 한다(영수증 가드의 fail-closed 가정은 codex 게이트 전용이므로 가져오지 않는다).
- SubagentStop(`subagent_stop.go:38` `Handle`): 기존 `checkAuditorStop` 호출은 그대로 두고, 그 뒤에 `checkServedModelStop` 을 따로 평가한다. 둘의 `SystemMessage` 는 이어붙인다(`mergeAuditorStopGuard` 는 guard 의 메시지로 덮어쓰므로 병합 규칙을 확장). 서빙 쪽은 절대 `Decision` 을 설정하지 않는다 — 서빙 모델은 같은 인스턴스를 계속 돌려도 바뀌지 않으므로 block 은 루프만 만든다(REQ-SMA-009).
- 기록: 게이트 감사관(`auditreceipt.IsAuditorAgent`)이고 서빙 판정이 `served_drift`/`unknown` 이면 served-kind `Rejection` 을 쓴다. 원인 상수 `CauseServedModelDrift`, `CauseServedModelUnknown` 에 서빙 집합·기대 모델을 덧붙인다. SPEC 은 `auditreceipt.ParseVerdictLine`(`store.go:423`)으로 최종 메시지에서 읽고, 없으면 `auditreceipt.UnknownSpec`. 서빙 `ok` 면 그 역할의 served-kind 만 해제.
- PreToolUse(`pre_tool.go:724` 옆): 기존 `checkAuditReceiptSpawn` 은 그대로 두고, 그 **안의** 목록 소비를 kind 로 거르게 바꾼 뒤(receipt 만 셈) 새 `checkServedModelSpawn` 을 형제로 추가한다 — `servedGateScope` 가 참이고 served-kind 기록이 있으면 `SERVED_MODEL_VIOLATION` 센티널로 거부. 두 종류가 함께 남은 경우의 사유 결합(REQ-SMA-010): 두 검사를 모두 평가해 거부 사유를 `AUDIT_RECEIPT_VIOLATION: …; SERVED_MODEL_VIOLATION: …` 로 이어 한 번에 돌려준다.

### D4. 해석 모델 출처 — PreToolUse 와 같은 설정 제공자 주입

`subagentStopHandler` 는 오늘 `struct{}`(`subagent_stop.go:23`)이고 `internal/cli/deps.go:283` 에서 인자 없이 등록된다. SubagentStart 는 이미 `NewSubagentStartHandlerWithConfig(deps.Config)`(`deps.go:273`)로 같은 제공자를 받는다. 같은 모양으로:

- `NewSubagentStopHandlerWithConfig(cfg ConfigProvider)` 추가, `deps.go:283` 을 `hook.NewSubagentStopHandlerWithConfig(deps.Config)` 로 교체. 기존 `NewSubagentStopHandler()` 는 nil 제공자로 위임(테스트 호환).
- 해석은 `resolveAgentModel(cfg.Get().LLM, agentType)`(`agent_model_guard.go:113`) — PreToolUse 의 `llmConfig`(`:222`)와 같은 출처이므로 활성 프로필·`agent_overrides`·프로필 미러가 모두 반영되고, 같은 spawn 의 PreToolUse 행 `resolved_model` 과 SubagentStop 행 `resolved_model` 이 같아진다.
- 제공자가 nil 이거나 `Get()` 이 nil 이면: 선언 모델이 있으면 그것으로 판정, 없으면 `unknown`(REQ-SMA-004). PreToolUse 처럼 zero 값 기본 매트릭스로 떨어지지 않는다 — 활성 프로필과 다른 기대치로 거짓 drift/ok 를 만들 수 있기 때문이다.

### D5. 사후 스캔 표면 — `moai doctor` 점검 `"Served Model"`

`internal/cli/doctor.go:201` `moaiChecks` 에 `servedModelCheckName`(상수) 점검을 등록한다.

- **기준 디렉터리**: `internal/escalation/roots.go:35` `memoryRoots` 의 기준 집합(`$CLAUDE_CONFIG_DIR`, `~/.claude`). 비공개이므로 기준 계산을 export 함수로 끌어내 공유한다.
- **slug 열거 규칙**(REQ-SMA-014): 각 기준의 `projects/` 아래 디렉터리 중 이름이 ① primary 체크아웃 slug 와 같음, ② `<primary slug>--claude-worktrees-` 로 시작함, ③ `git worktree list --porcelain` 이 나열한 경로의 slug 와 같음 — 셋 중 하나를 만족하는 것. slug 계산은 `escalation.MemorySlug`(`roots.go:21`) 한 곳만 쓴다. primary 는 `canonicalProjectRoot` 로 얻는다(워크트리에서 doctor 를 돌려도 같은 집합). ②는 §A.2 실측에서 363개 slug 디렉터리가 이 모양임을 확인한 규칙이고, ③은 L2 워크트리처럼 다른 접두를 가진 살아 있는 트리를 덮는다. 둘 다에 없는 과거 slug 는 Out of Scope.
- 트랜스크립트: 각 slug 아래 `*/subagents/agent-*.jsonl`. 분류 로직은 SubagentStop 과 **한 구현**을 공유한다(hook 패키지 export 함수). 해석 모델은 doctor 가 적재한 프로젝트 설정으로 계산한다.
- **명시 호출 전용(run 단계 결정)**: 기본 `moai doctor` 는 이 점검 자리에 info 행 1개(`run with --check "Served Model"`)만 내고 스캔하지 않는다. 스캔은 `moai doctor --check "Served Model"` 에서만 돈다(`servedModelDoctorEntry`, `internal/cli/doctor_served_model.go`). 기본 실행에 넣으면 이 머신(트랜스크립트 약 2.3k개)에서 doctor 한 번에 5.7–10.2초가 더해지기 때문이다. 점검 이름은 `moaiChecks` 에 계속 등록되므로 `namesAddedAfterBaseline` 등록은 그대로 필요하다.
- 출력(명시 호출 시): 상태는 발견 시 warn, 발견 없음 ok, **스캔 대상 0개면 info**(ok 아님). 메시지는 항상 swept 개수와 검색한 기준·slug 를 담는다. 읽기 전용이며 doctor 종료 코드를 바꾸지 않는다.
- **[HARD] 점검 이름을 `internal/cli/binary_lag_test.go:198` `namesAddedAfterBaseline` 에 `"servedModelCheckName": true` 로 등록한다.** 누락 시 `TestBinaryLag_DoctorCheckNameSetIsUnchanged` 가 CI 에서 붉어진다.

### D6. 설정 키 — 형제 키 `workflow.served_model_gate.enabled`

`agent_model_guard` 의 하위 키가 아니라 형제 키로 둔다. `internal/config/types.go` 의 `SettingsDriftGate` 주석 원칙 — 한 플래그가 서로 다른 표면의 두 거부를 함께 켜면 운영자가 어느 쪽을 끄려 했는지 말할 수 없다.

- `internal/config/types.go`: `ServedModelGate ServedModelGateConfig \`yaml:"served_model_gate"\`` + `ServedModelGateConfig{Enabled bool}`
- `internal/config/defaults.go`: `Enabled: false`, 가드 패밀리와 같은 중립성 주석
- 템플릿 `internal/template/templates/.moai/config/sections/workflow.yaml`: `served_model_gate: enabled: false` + 중립 주석
- 로컬 `.moai/config/sections/workflow.yaml`: `served_model_gate: enabled: true` + "로컬 opt-in, 템플릿 기본 false" 주석

### D7. 기대 모델 대조 규칙

기대 모델 = meta.json `model` 이 비어 있지 않으면 그것, 아니면 D4 해석 결과. 둘 다 없으면 `unmapped`(해석기가 매핑 없음을 보고) 또는 `unknown`(설정 제공자 없음, D4). 일치: 대소문자 무시 동일, 또는 기대가 순수 계열 별칭이고 서빙 식별자가 하이픈 구분 토큰으로 그 계열 토큰을 포함하는 Claude 모델 식별자. 별칭 집합은 해석기가 내는 값을 기준으로 하되 손으로 재기술하지 않는다. 실측 값 `claude-opus-5-5`·`claude-opus-5` 는 `opus` 와 일치, `glm-5.3-flash`·`glm-5.3` 은 불일치.

### D8. 트랜스크립트 읽기 예산과 자기 보고

- 한 줄씩 스트리밍으로 읽고 줄 수·바이트 상한(구현 상수, 테스트에서 주입 가능)을 넘으면 `unknown`. 3.4MB 트랜스크립트는 예산 안이어야 한다(AC-SMA-016 에서 실측).
- 자기 보고는 `input.LastAssistantMessage` 의 첫 비공백 줄이 `auditor-model: ` 로 시작할 때만 그 값을 쓴다. 보고서 파일은 읽지 않는다(REQ-SMA-016).

## §C 사전 점검 (run 착수 전)

1. `git -C .claude/worktrees/t1282 rev-list --count --left-right develop...HEAD` 로 기준 확인.
2. `go list -deps ./internal/auditreceipt/... | grep moai-adk/internal/hook` 출력 없음 확인(순환 방지).
3. `go test ./internal/hook/... ./internal/auditreceipt/... ./internal/config/... -count=1` 기준선 기록.
4. `golangci-lint --version` 이 CI 판 v2.1.6 인지 확인.

## §D 제약

- 템플릿 퍼스트: 템플릿 먼저, `make build`, 로컬 미러.
- C3(`internal/template/templates/.codex/agents/moai/*.toml`) 손편집 금지 — `make agents-emit` 으로만 재생성.
- 전체 스위트 로컬 실행 금지 — 영향 패키지만, 전 패키지 판정은 CI.

## §E 자기 검증 (run-phase 종료 조건)

```bash
go test ./internal/hook/... -count=1
go test ./internal/auditreceipt/... ./internal/config/... -count=1
go test ./internal/cli/ -run '^TestBinaryLag_.*$' -count=1 -v
go test ./internal/cli/ -run '^TestServedModelCheck_Sweep$' -count=1 -v
go test ./internal/template/... -count=1
make agents-emit-check
golangci-lint run ./internal/hook/... ./internal/auditreceipt/... ./internal/config/... ./internal/cli/...
```

acceptance.md 의 AC 별 명령이 권위다. 위 목록은 묶음 실행용 요약이다.

## §F 마일스톤 (결정 변경 가능성이 높은 순서, 시간 추정 없음)

### M1 — 판정 코어 (Priority High)

- 서빙 집합 추출, 기대 모델 대조(D7), 판정 4값, 읽기 예산(D8)을 순수 함수로.
- 테스트 먼저: `TestServedModel_Classify`(AC-SMA-001), `TestServedModel_UnknownNeverOK`(AC-SMA-002).

### M2 — 설정 키와 설정 제공자 주입 (Priority High)

- D6 네 곳 + D4 의 `NewSubagentStopHandlerWithConfig`·`deps.go:283` 교체.
- 테스트: `TestServedModelGate_DefaultOff`(AC-SMA-010), `TestServedModel_ResolvedFromActiveProfile`(AC-SMA-003).

### M3 — SubagentStop 관측 배선 (Priority High)

- 기록 행(D1), 경로 파생, 경고, fail-open, 자기 보고(D8).
- 테스트: `TestServedModel_SubagentStopRecord`(AC-SMA-004), `TestServedModel_GateOffWarnsOnly`(AC-SMA-005), `TestServedModel_SelfReportFromFinalMessage`(AC-SMA-014).

### M4 — kind 분리와 채택 거부 (Priority High)

- D2 + D3 전부.
- 테스트: `TestServedModel_GateOnAdoptionRefusal`(AC-SMA-006), `TestServedModel_RefusalDeniesPhaseEntry`(AC-SMA-007), `TestServedModel_ClearingIsKindScoped`(AC-SMA-008), `TestServedModel_GateDoesNotArmReceiptGuard`(AC-SMA-009).
- 기존 영수증 가드 테스트는 수정 없이 통과해야 한다: `go test ./internal/hook/... ./internal/auditreceipt/... -count=1`(AC-SMA-015).

### M5 — 사후 스캔 doctor 점검 (Priority Medium)

- D5 전부. 테스트 `TestServedModelCheck_Sweep`·`TestServedModelCheck_DefaultRunShowsHintOnly`(AC-SMA-011).
- `namesAddedAfterBaseline` 등록 후 `go test ./internal/cli/ -run '^TestBinaryLag_.*$' -count=1 -v`(AC-SMA-012).

### M6 — 감사관 자기 보고 첫 줄 (Priority Medium)

- C2 `internal/template/templates/.claude/agents/moai/{plan,sync}-auditor.md` 의 Output Format 절에 다음 의무 문장을 그대로 넣는다: `The first line of the report file and of your final message MUST be \`auditor-model: <served model>\`` (뒤에 짧은 설명을 붙여도 되나 이 문장 머리는 바이트 그대로). C1 `.claude/agents/moai/{plan,sync}-auditor.md` 도 같은 문장을 반영(C1↔C2 는 바이트 동일 관계가 아님 — 해당 절만 맞춘다).
- **`make agents-emit` 실행**으로 C3 재생성 → **`make agents-emit-check` exit 0**. C3 손편집 금지.
- 문구 중립성: SPEC ID·카드 ID·날짜 없음(AC-SMA-013 세 번째 명령).
- 종단 실측: `go run ./cmd/moai doctor --check "Served Model" --verbose` 로 세션 `d46e0166` 이 `served_drift` 로 보고되는지 확인(AC-SMA-016).
- `make build` 후 영향 패키지 재측정, golangci-lint v2.1.6(AC-SMA-015).

## §G 위험

| # | 위험 | 대응 |
|---|------|------|
| R1 | SubagentStop 페이로드에 `agent_transcript_path` 가 없을 수 있음 | REQ-SMA-001 파생 경로. 둘 다 실패하면 `unknown` |
| R2 | 발화 시점에 마지막 assistant 행이 아직 안 쓰였을 수 있음 | 집합 판정이라 대부분 무영향. 행 0건이면 `unknown`. M3 에서 실세션 1회 관측해 progress.md 에 기록 |
| R3 | 5초 훅 timeout 초과 | D8 예산. 예산 초과는 `unknown` |
| R4 | GLM 백엔드 사용자에게 매 서브에이전트마다 경고 | 기본 OFF 라 거부는 없음. 경고 소음은 Out of Scope, sync 에서 후속 카드 후보로 보고 |
| R5 | 영수증·서빙 거부의 파일 키 충돌 / 해제 교차 | D2. AC-SMA-008 |
| R6 | `moai update` 가 `.moai/config` 를 재배포하며 로컬 `enabled: true` 를 잃을 가능성 | 기존 로컬 가드 키와 같은 운명. sync 에서 update 후 생존 여부를 확인하고 보고 |
| R7 | 판정 로직 이중 구현(hook vs doctor) | D5 — 한 구현 공유 |
| R8 | 서빙 게이트가 영수증 가드를 무장시키는 결합 | D3 전용 술어. AC-SMA-009 |
| R9 | 활성 프로필과 다른 기대치로 거짓 판정 | D4 같은 설정 제공자, 제공자 없으면 `unknown`. AC-SMA-003 |

## §H 교차 참조

- `.moai/reports/t1282/verdict.md` — 착수 판정서(재현 근거)
- `.moai/reports/plan-audit/SPEC-SERVED-MODEL-AUDIT-001-review-1.md` — plan-audit iter1
- SPEC-AGENT-MODEL-ENFORCE-001 — PreToolUse 선언 모델 감사의 기원
- SPEC-CODEX-AUDIT-GATE-AXES-001 — 감사관 판정 채택 거부와 phase 진입 차단의 기원
- SPEC-WORKTREE-STATE-ROOT-001 — 설정 고아 워크트리의 저장소 해석
