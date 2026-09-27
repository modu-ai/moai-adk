# SPEC-SERVED-MODEL-AUDIT-001 — 인수 기준

---

## §A 판정 규약

1. 모든 명령은 워크트리 루트(`.claude/worktrees/t1282`)에서 실행한다.
2. `go test` 기반 AC 는 종료 코드 0 **그리고** `-v` 출력에 `--- PASS: <테스트 이름> ` (이름 뒤 공백 포함) 줄이 있을 때만 PASS 다. `testing: warning: no tests to run` 이 나오면 FAIL 이다.
3. 모든 `-run` 패턴은 `^…$` 로 고정하고 AC 하나에 최상위 테스트 하나를 둔다. 형제 사례는 그 테스트 안의 표 기반 하위 테스트로 묶고, 표의 각 행은 아래 시나리오에 적힌 대로 존재해야 한다(행 누락은 FAIL — sync-auditor 가 표를 대조한다).
4. 부정형 단언(“기록되지 않는다”, “거부되지 않는다”)을 담는 표는 같은 테스트 안에 양성 대조 행을 둔다. 0건이 검사기 고장이 아님을 보이기 위함이다.
5. 린트는 CI 판 golangci-lint v2.1.6 으로 잰다.

---

## §B AC 매트릭스

| AC | REQ | 요지 |
|----|-----|------|
| AC-SMA-001 | REQ-SMA-001, REQ-SMA-003 | 판정 표: ok / served_drift(선언 opus·서빙 GLM) / 혼재 drift / unmapped |
| AC-SMA-002 | REQ-SMA-001, REQ-SMA-004 | unknown 표: 트랜스크립트 부재·model 빈 값·필드 부재·synthetic 전용·예산 초과·설정 없음 — 절대 ok 아님 |
| AC-SMA-003 | REQ-SMA-003, REQ-SMA-004 | 해석 모델은 활성 프로필·agent_overrides 를 따르는 설정 제공자에서 나옴 |
| AC-SMA-004 | REQ-SMA-001, REQ-SMA-002, REQ-SMA-006 | SubagentStop 기록 행(필드·기존 행 불변·경로 파생·fail-open) |
| AC-SMA-005 | REQ-SMA-005, REQ-SMA-008 | 게이트 OFF → drift·unknown 모두 경고만, 거부·차단 없음 |
| AC-SMA-006 | REQ-SMA-009, REQ-SMA-012 | 게이트 ON → 감사관 drift·unknown 은 채택 거부, 비감사관은 없음, 되돌림 없음 |
| AC-SMA-007 | REQ-SMA-010 | phase 진입 spawn 거부 — 센티널·병합 사유 |
| AC-SMA-008 | REQ-SMA-011 | 해제는 kind 한정, Kind 부재 레코드는 receipt 로 취급, 상호 덮어쓰기 없음 |
| AC-SMA-009 | REQ-SMA-013 | 서빙 게이트 ON · codex 게이트 미선언 → 영수증 가드 비활성 유지 |
| AC-SMA-010 | REQ-SMA-007 | 설정 기본값 OFF · 템플릿 false · 로컬 true |
| AC-SMA-011 | REQ-SMA-014 | doctor 사후 스캔: 워크트리 slug 포함·빈 스캔 비-ok·읽기 전용·종료 코드 불변 |
| AC-SMA-012 | REQ-SMA-014 | 새 doctor 점검 이름의 binary_lag 허용 목록 등록 |
| AC-SMA-013 | REQ-SMA-015 | 감사관 정의 C1·C2·C3 가 보고서 파일과 최종 메시지 첫 줄 요구 |
| AC-SMA-014 | REQ-SMA-016 | 자기 보고는 `last_assistant_message` 첫 줄에서 읽어 기록, 판정은 바꾸지 않음 |
| AC-SMA-015 | REQ-SMA-006, REQ-SMA-010, REQ-SMA-013 | 기존 hook·auditreceipt·config 회귀 없음, 린트 0 |
| AC-SMA-016 | REQ-SMA-003, REQ-SMA-014 | 재현 세션 d46e0166 실측 served_drift |

---

## §C 시나리오

### AC-SMA-001 — 판정 표

- **Given** 표의 각 행: (a) 선언 `opus`, 서빙 전부 `claude-opus-5-5` · (b) 선언 `opus`, 서빙 전부 `glm-5.3-flash`(판정서 재현 모양) · (c) 기대 `opus`, 서빙 `claude-opus-5` 와 `glm-5.3` 혼재 · (d) 선언 없음, 에이전트 유형이 프로필에 매핑되지 않음, 서빙 `claude-opus-5-5`
- **When** 각 행을 판정하면
- **Then** (a) `ok`, (b) `served_drift`, (c) `served_drift` 이고 서빙 집합에 두 값 모두 있음, (d) `unmapped` 이고 서빙 집합은 기록됨.

```bash
go test ./internal/hook/ -run '^TestServedModel_Classify$' -count=1 -v
```

### AC-SMA-002 — unknown 표 (절대 ok 아님)

- **Given** 표의 각 행: (a) 존재하지 않는 트랜스크립트 경로 · (b) assistant 행 `message.model` 이 `""` · (c) `model` 필드 부재 · (d) assistant 행 전부 `<synthetic>` · (e) 테스트가 주입한 작은 읽기 예산을 넘는 일치 픽스처 · (f) 선언 없음 + 설정 제공자가 nil 설정을 돌려줌 · 양성 대조 (g) `<synthetic>` 1행 + `claude-opus-5-5` 1행 · (h) (e) 와 같은 픽스처를 기본 예산으로
- **When** 각 행을 판정하면
- **Then** (a)-(f) 는 모두 `unknown`, (g)(h) 는 `ok` 이며 (g) 의 서빙 집합에 `<synthetic>` 이 없다.

```bash
go test ./internal/hook/ -run '^TestServedModel_UnknownNeverOK$' -count=1 -v
```

### AC-SMA-003 — 해석 모델 출처

- **Given** 기본 매트릭스로는 manager-develop 을 `opus` 로 해석하지만, 설정 제공자가 돌려주는 LLM 설정은 (a) 다른 활성 프로필로 `sonnet`, (b) `agent_overrides` 로 `haiku` 를 지정하는 두 픽스처. 두 경우 모두 선언 모델은 없고 서빙은 `claude-sonnet-5` 다
- **When** SubagentStop 핸들러가 그 설정 제공자로 생성되어 실행되면
- **Then** (a) 기록 행의 `resolved_model` 은 `sonnet`, 판정 `ok` · (b) `resolved_model` 은 `haiku`, 판정 `served_drift`. 같은 입력을 기본값(zero) 설정으로 판정했을 때와 결과가 다름을 테스트가 함께 단언한다.

```bash
go test ./internal/hook/ -run '^TestServedModel_ResolvedFromActiveProfile$' -count=1 -v
```

### AC-SMA-004 — SubagentStop 기록 행

- **Given** 임시 루트의 `.moai/logs/agent-model-audit.jsonl` 에 PreToolUse 행 1개가 있음. 표의 각 행: (a) `agent_transcript_path` 가 drift 픽스처를 가리킴 · (b) `agent_transcript_path` 는 비고 `transcript_path=<dir>/<session>.jsonl`, `agent_id` 가 주어지며 `<dir>/<session>/subagents/agent-<id>.jsonl` 에 drift 픽스처 · (c) 로그 디렉터리 쓰기 불가 · (d) 손상된 JSON 행이 섞인 트랜스크립트 · (e) 빈 CWD
- **When** SubagentStop 핸들러가 실행되면
- **Then** (a)(b) 파일은 정확히 2행, 1행은 원래 바이트와 같고, 2행은 `session_id`, `agent_id`, `agent`, `declared_model`, `resolved_model`, `served_models`, `self_reported_model`, `source`, `verdict: "served_drift"` 를 담음((b) 는 파생 경로가 실제로 읽혔다는 증거) · (c)(d)(e) 핸들러는 오류를 돌려주지 않고 패닉하지 않으며, 행이 기록된 경우 그 판정은 `ok` 가 아니다.

```bash
go test ./internal/hook/ -run '^TestServedModel_SubagentStopRecord$' -count=1 -v
```

### AC-SMA-005 — 게이트 OFF: 경고만

- **Given** `workflow.served_model_gate.enabled` 가 없거나 false 인 트리. 표의 각 행: (a) plan-auditor · drift 픽스처 · (b) plan-auditor · 트랜스크립트 부재(unknown) · 양성 대조 (c) 같은 트리에서 게이트만 true 로 바꾼 (a)
- **When** SubagentStop 과 이어지는 manager-develop spawn 의 PreToolUse 가 실행되면
- **Then** (a)(b) SubagentStop 출력의 `SystemMessage` 에 에이전트 유형·기대 모델·서빙 집합이 있고 `Decision` 은 비어 있으며, 저장소에 served-kind 거부 기록이 없고 spawn 은 허용된다 · (c) served-kind 거부 기록이 생긴다.

```bash
go test ./internal/hook/ -run '^TestServedModel_GateOffWarnsOnly$' -count=1 -v
```

### AC-SMA-006 — 게이트 ON: 채택 거부, 되돌림 없음

- **Given** 게이트 ON 트리, 감사관 보고서 파일 1개와 SPEC 디렉터리가 있는 임시 트리. 표의 각 행: (a) sync-auditor, 마지막 줄 `AUDIT-VERDICT: PASS spec=SPEC-X-001 receipts=none`, drift 픽스처 · (b) plan-auditor, 트랜스크립트 부재 · (c) manager-develop, drift 픽스처
- **When** SubagentStop 이 실행되면
- **Then** (a) 역할 `sync-auditor`, SPEC `SPEC-X-001`, kind `served`, 서빙 원인을 담은 거부 기록 1건 · (b) 원인이 `unknown` 을 가리키는 served-kind 거부 기록 1건 · (c) 감사 로그에 `served_drift` 행과 경고는 있으나 거부 기록 0건. 모든 행에서 출력의 `Decision` 은 비어 있고(`block` 아님), 보고서·SPEC 파일의 바이트는 실행 전과 같으며, 새로 생긴 파일은 감사 로그 행과 거부 기록뿐이다.

```bash
go test ./internal/hook/ -run '^TestServedModel_GateOnAdoptionRefusal$' -count=1 -v
```

### AC-SMA-007 — phase 진입 spawn 거부

- **Given** 표의 각 행: (a) served-kind 거부 1건만 남은 게이트 ON 트리 · (b) 같은 트리에 codex 게이트 `required` 와 receipt-kind 거부 1건이 함께 있음 · 대조 (c) (a) 트리에서 게이트만 false 로 바꿈
- **When** `manager-develop`, `manager-docs`, `manager-git`, `Explore` spawn 의 PreToolUse 가 각각 실행되면
- **Then** (a) 세 phase 진입 spawn 은 `SERVED_MODEL_VIOLATION` 으로 시작하고 역할·SPEC·서빙 원인을 담은 사유로 거부, `Explore` 는 허용 · (b) 사유에 `AUDIT_RECEIPT_VIOLATION` 과 `SERVED_MODEL_VIOLATION` 이 모두 있고 각 센티널 뒤에 자기 종류의 기록이 나열됨 · (c) 네 spawn 모두 served 근거로는 거부되지 않음.

```bash
go test ./internal/hook/ -run '^TestServedModel_RefusalDeniesPhaseEntry$' -count=1 -v
```

### AC-SMA-008 — kind 한정 해제와 레거시 레코드

- **Given** 서빙 게이트 ON 이고 codex 게이트 `required` 인 트리(두 해제 경로가 모두 살아 있어야 하므로). 같은 역할·같은 SPEC 에 served-kind 거부 1건과, `Kind` 필드가 없는 기존 형식의 거부 1건(오늘의 영수증 가드가 쓰는 형태)이 함께 있는 저장소. 두 기록이 파일 2개로 따로 존재함을 먼저 단언한다
- **When** (a) 같은 역할의 감사관이 서빙 `ok`·영수증 미증명으로 멈추면, (b) 따로 영수증 증명 PASS·서빙 drift 로 멈추면
- **Then** (a) served-kind 만 사라지고 Kind 부재 레코드는 남는다 · (b) Kind 부재 레코드(receipt 로 취급)만 사라지고 served-kind 는 남는다. 어느 쪽에서도 한 기록이 다른 기록을 덮어쓰지 않는다.

```bash
go test ./internal/hook/ -run '^TestServedModel_ClearingIsKindScoped$' -count=1 -v
```

### AC-SMA-009 — 영수증 가드 비결합

- **Given** 서빙 게이트 ON, `workflow.audit.gates.codex` 미선언 트리. plan-auditor 가 SubagentStart 를 거쳐 마지막 줄 `AUDIT-VERDICT: PASS spec=SPEC-X-001 receipts=none` 과 서빙 `ok` 픽스처로 멈추고, 이어서 manager-develop spawn 이 실행됨. 양성 대조: 같은 입력에 codex 게이트 `required` 를 선언한 트리
- **When** SubagentStart · SubagentStop · PreToolUse 가 차례로 실행되면
- **Then** 서빙 게이트만 켠 트리에서는 SubagentStop 출력 `Decision` 이 비어 있고, 시작 마커 0개, 거부 기록 0개, spawn 허용. 양성 대조 트리에서는 영수증 가드가 기존대로 `block` 과 receipt-kind 거부를 낸다.

```bash
go test ./internal/hook/ -run '^TestServedModel_GateDoesNotArmReceiptGuard$' -count=1 -v
```

### AC-SMA-010 — 설정 기본값

- **Given** 설정 파일이 없는 트리
- **When** 기본 설정을 적재하면
- **Then** `Workflow.ServedModelGate.Enabled` 는 false 다. 템플릿 workflow.yaml 의 `served_model_gate:` 다음 줄은 `enabled: false`, 로컬은 `enabled: true` 이며, 템플릿 가드 테스트(중립성 포함)는 exit 0.

```bash
go test ./internal/config/ -run '^TestServedModelGate_DefaultOff$' -count=1 -v
grep -n -A1 'served_model_gate:' internal/template/templates/.moai/config/sections/workflow.yaml
grep -n -A1 'served_model_gate:' .moai/config/sections/workflow.yaml
go test ./internal/template/... -count=1
```

### AC-SMA-011 — doctor 사후 스캔

- **Given** 가짜 설정 기준 디렉터리 두 개(`CLAUDE_CONFIG_DIR` 과 홈 `.claude` 대역)와 가짜 primary 경로 `/x/proj` 를 둔 픽스처. 표의 각 행: (a) primary slug 아래 drift 감사관 1건·ok 1건 + `<primary slug>--claude-worktrees-lane1` 아래 drift 감사관 1건 + `git worktree list` 대역으로 주입한 경로의 slug 아래 unknown 감사관 1건 + 무관한 slug `-x-other` 아래 drift 1건 · (b) 트랜스크립트가 하나도 없음
- **When** 점검 함수를 실행하면
- **Then** (a) 상태 warn, 메시지에 swept 4, 검색한 기준·slug 목록, verdict 별 개수(ok 1 / served_drift 2 / unknown 1)와 세 감사관 실행(워크트리 slug 건 포함)이 나열되고 `-x-other` 건은 세지 않음 · (b) 상태는 `ok` 가 아닌 info 이며 메시지에 swept 0 과 검색한 기준·slug 가 있음. 두 행 모두 실행 전후 픽스처 트리와 감사 로그의 바이트가 같고, 이 점검 결과로 doctor 종료 코드가 바뀌지 않음(같은 테스트가 doctor 진입 함수의 종료 코드를 대조군과 비교).

```bash
go test ./internal/cli/ -run '^TestServedModelCheck_Sweep$' -count=1 -v
```

### AC-SMA-012 — binary_lag 허용 목록 등록

- **Given** `moaiChecks` 에 새 점검이 등록된 트리
- **When** binary_lag 가드 계열을 실행하면
- **Then** 전부 통과하고, `namesAddedAfterBaseline` 에 `servedModelCheckName` 키가 있다.

```bash
go test ./internal/cli/ -run '^TestBinaryLag_.*$' -count=1 -v
grep -n '"servedModelCheckName"' internal/cli/binary_lag_test.go
```

### AC-SMA-013 — 감사관 정의의 첫 줄 요구

- **Given** 편집된 C2 원본, C1 로컬 사본, `make agents-emit` 으로 재생성한 C3
- **When** 방출 일치 검사와 문자열 검사를 실행하면
- **Then** `make agents-emit-check` exit 0 · 아래 6개 파일 각각에서 두 번째 명령의 개수가 1 이상 — 의무 문장 `The first line of the report file and of your final message MUST be` 가 있다는 뜻 · 세 번째 명령(C2 에서 그 의무 문장이 든 줄 중 SPEC ID·카드 ID 가 섞인 줄 수)은 0 · 에이전트 방출 테스트 exit 0.

```bash
make agents-emit-check
grep -cF 'The first line of the report file and of your final message MUST be' internal/template/templates/.claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md internal/template/templates/.codex/agents/moai/plan-auditor.toml internal/template/templates/.codex/agents/moai/sync-auditor.toml
grep -hF 'The first line of the report file and of your final message MUST be' internal/template/templates/.claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md | grep -cE 'SPEC-[A-Z]|\bt[0-9]{3,}\b'
go test ./internal/template/agentemit/... -count=1
```

- plan 단계 측정: 두 번째 명령은 6개 파일 모두 0(RED).

### AC-SMA-014 — 자기 보고는 최종 메시지에서, 판정은 불변

- **Given** 서빙 `glm-5.3-flash` 픽스처. 표의 각 행: (a) `last_assistant_message` 첫 줄이 `auditor-model: opus` · (b) `last_assistant_message` 에 그 줄이 없고, 같은 트리의 보고서 파일 첫 줄에만 `auditor-model: opus` 가 있음 · (c) 첫 줄이 아닌 셋째 줄에 `auditor-model: opus`
- **When** SubagentStop 이 실행되면
- **Then** (a) 기록 행 `self_reported_model` 은 `opus`, `verdict` 는 `served_drift` · (b)(c) `self_reported_model` 은 빈 값, `verdict` 는 `served_drift`. 관측기는 보고서 파일을 읽지 않는다.

```bash
go test ./internal/hook/ -run '^TestServedModel_SelfReportFromFinalMessage$' -count=1 -v
```

### AC-SMA-015 — 회귀 없음

```bash
go test ./internal/hook/... -count=1
go test ./internal/auditreceipt/... ./internal/config/... -count=1
golangci-lint run ./internal/hook/... ./internal/auditreceipt/... ./internal/config/... ./internal/cli/...
```

- 모두 exit 0, 린트 신규 이슈 0(v2.1.6). 기존 영수증 가드 테스트가 수정 없이 통과해야 한다.

### AC-SMA-016 — 재현 세션 실측 (개발 머신 한정)

- **Given** 착수 판정서의 세션 `d46e0166` 트랜스크립트가 이 머신의 프로필 디렉터리에 있음
- **When** 새로 빌드한 바이너리로 doctor 점검을 실행하면
- **Then** 출력에 manager-develop 한 건이 `served_drift`(서빙 `glm-5.3-flash`)로 나타난다. `agent-model-audit.jsonl` 의 같은 세션 PreToolUse 행은 여전히 `ok` 다(소급 기록 없음).

```bash
go run ./cmd/moai doctor --check "Served Model" --verbose
```

- 결과는 progress.md §E.2 에 명령과 출력 원문으로 남긴다. 해당 트랜스크립트가 없는 머신에서는 Gap 으로 보고한다.

---

## §D 품질 게이트

- TRUST 5: 신규 코드 패키지 커버리지 85% 이상(`go test -cover ./internal/hook/`), gofmt 무변경, 린트 0.
- 템플릿 중립성 가드와 `make agents-emit-check` 통과.
- @MX: 새 SubagentStop 관측 진입점과 kind 분리 저장소 함수에 `@MX:ANCHOR` 검토(fan_in 확인), 트랜스크립트 스트리밍 읽기에 `@MX:NOTE`.

## §E Definition of Done

- AC-SMA-001..015 전부 PASS, AC-SMA-016 은 PASS 또는 명시적 Gap.
- progress.md §E.2 에 각 AC 의 명령·출력 원문과 측정 트리 HEAD 가 기록됨.
- C3 는 `make agents-emit` 으로만 바뀌었음(`make agents-emit-check` exit 0).
