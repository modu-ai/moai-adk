# SPEC-AGENT-MODEL-ENFORCE-001 — 구현 계획

> 대상 SPEC: SPEC-AGENT-MODEL-ENFORCE-001
> Tier: M — 산출물 3종(spec / plan / acceptance) + progress
> 방법론: `.moai/config/sections/quality.yaml`의 `constitution.development_mode` 준수

---

## §A Context

### §A.1 작업 위치

- 워크트리: `.claude/worktrees/web-redesign` (브랜치 `feat/web-console-redesign`)
- 주 편집 대상: `internal/hook/`, `internal/config/`, `internal/template/templates/.claude/settings.json.tmpl`, `.claude/rules/moai/`
- 모든 `.claude/**` 편집은 `internal/template/templates/` 미러 + `make build`를 **같은 커밋**에 담는다.

### §A.2 아키텍처 요약 (실측)

```
Agent 도구 spawn
  └─ tool_input: {description, subagent_type, prompt, run_in_background, [name], [model], [isolation]}
        │  (실측: model은 156건 중 1건에만 존재)
        ▼
  PreToolUse 훅  ← 현재 matcher "Write|Edit|Bash" 라서 여기에 도달하지 않음 (F2)
        │
        ▼  (신설)
  moai hook pre-tool → preToolHandler.Handle
        ├─ ToolName ∈ {Agent, Task} 분기
        ├─ subagent_type 추출 → template.ResolveAgentModelEffort(cfg.LLM, agent)
        ├─ 판정 {ok|missing|mismatch|unmapped}
        ├─ .moai/logs/agent-model-audit.jsonl 추가 (M2)
        ├─ 권고 메시지 (M3)
        └─ workflow.agent_model_guard.enabled && mismatch → deny (M4)
```

### §A.3 PRESERVE 목록 (건드리지 않는다)

| 대상 | 이유 |
|---|---|
| PreToolUse `"matcher": "Write\|Edit\|Bash"` 블록 | Write/Edit/Bash 경로의 지연·타임아웃 특성 회귀 금지 (REQ-AME-010) |
| `checkBranchState` / `checkBashCommand` 순서 | 기존 deny 우선순위 계약 |
| `template.defaultProfileMatrix` 33셀 | 본 SPEC은 소비자 (§C Out of Scope) |
| `model-policy.md`의 `paths:` 스코프 | REQ-AME-043 — 상세 규칙 로드 동작 유지 |
| `post_tool.go`의 Agent/Task 분기 | F4로 도달 불가임이 밝혀졌으나 제거는 별도 SPEC 소관 |
| `.claude/agents/moai/*.md` frontmatter | 모델 핀 재작성은 본 SPEC 범위 밖 |

---

## §B 되돌리기 어려운 결정 (먼저 검토할 것)

### D1 — 무엇을 "드리프트"로 판정할 것인가 (브리프 전제 정정)

브리프와 issue #1376은 "주입된 모델이 프로파일과 불일치"를 상정한다. 실측(spec.md §A.1 F6)은 다른 그림이다: **156건 중 155건이 `model` 인자 자체를 갖지 않는다.**

결정: 판정을 4치로 정의한다 — `ok` / `missing` / `mismatch` / `unmapped`. 지배적 케이스는 `missing`이며, 이것이 issue가 신고한 "침묵 폴백"의 실제 형태다.

되돌리기 어려운 이유: 이 판정 어휘가 로그 스키마와 차단 정책(D3)을 동시에 규정한다. 2치(`ok`/`violation`)로 출발하면 `missing`과 `mismatch`를 구분할 수 없고, 차단 정책이 곧바로 전면 봉쇄로 붕괴한다.

### D2 — 관측 지점: PreToolUse vs PostToolUse vs SubagentStart

| 후보 | 장점 | 실측된 문제 | 판정 |
|---|---|---|---|
| PreToolUse + `Agent\|Task` matcher | 유일하게 **차단 가능**한 지점. `tool_input` 원본 접근 | Agent에 대한 발화 여부 **미검증**(§A.3 정직성 계약) | **채택** — 단 M1 측정이 선행 조건 |
| PostToolUse Agent 분기 재활용 | 코드가 이미 존재 | F4: `handle-post-tool.sh`가 `Write\|Edit\|MultiEdit` 블록에만 등록 → 도달 불가. 배선 확장이 선행돼야 하고 그래도 차단 불가 | 기각 (대체안으로만 보류) |
| SubagentStart | `agent_id` / `agent_type`를 확실히 받음 | `HookInput`에 모델 필드 없음 — 관측 대상 자체를 볼 수 없음 | 기각 |

되돌리기 어려운 이유: matcher 블록 신설은 배포 settings 템플릿을 바꾼다. 잘못 고르면 모든 사용자 세션의 spawn 경로에 무의미한 5초 타임아웃 훅이 붙는다.

### D3 — 차단 여부와 기본값

결정: **관측(기본 ON, 차단 없음) → 권고(기본 ON) → 차단(opt-in, 기본 OFF)** 3계층. 차단 대상은 `mismatch` **한정**이며 `missing`은 제외한다(REQ-AME-034).

`missing`을 차단하면 실측 기준 spawn의 99.4%가 막힌다 — 집행이 아니라 서비스 거부다. 게이트 기본값 `false`는 `Workflow.BranchGuard.Enabled`의 선례를 그대로 따른다(`internal/config/defaults.go`).

되돌리기 어려운 이유: 배포된 기본값을 나중에 뒤집으면 업그레이드한 사용자의 세션이 갑자기 막힌다.

### D4 — 규칙 가시성 해소 방식 [NEEDS CLARIFICATION: 스텁 배치 위치]

실측: `model-policy.md` = 27,571 bytes. 현재 항상 로드 규칙 = 13개 파일 197,215 bytes. 전체 always-load 전환은 항상 로드 총량을 **+14%** 시킨다.

결정: 전체 전환을 기각하고, per-spawn 모델 주입 의무만 담은 **압축 스텁**(목표 ≤ 2KB)을 항상 로드 표면에 신설한다. 상세는 `model-policy.md` 교차 참조로 남긴다.

[NEEDS CLARIFICATION: 스텁 배치 위치] 스텁을 (a) 신규 항상 로드 규칙 파일 `.claude/rules/moai/development/agent-spawn-model-injection.md`로 만들지, (b) 이미 항상 로드되는 `agent-common-protocol.md`의 한 절로 삽입할지. (a)는 파일 수를 늘리고 (b)는 기존 파일을 키운다. 토큰 비용은 사실상 동일하며 발견 가능성/소유권 관점의 선택이다. 구현 착수 승인 전에 결정한다.

### D5 — 감사 로그 형식 [NEEDS CLARIFICATION: jsonl vs log]

`branch-guard-audit.log`는 텍스트, `task-metrics.jsonl`은 JSON Lines다. 본 SPEC은 집계(에이전트별 드리프트율)를 목적으로 하므로 jsonl을 기본 선택한다. 다만 훅 감사 로그의 하우스 컨벤션이 `.log`라면 그쪽을 따른다. [NEEDS CLARIFICATION: jsonl vs log] — M2 착수 전 `.moai/logs/` 기존 파일 컨벤션 재확인으로 확정.

---

## §C Pre-flight (구현 착수 전 실행)

```bash
# 0. 작업 위치 확인
pwd && git rev-parse --show-toplevel && git branch --show-current

# 1. 베이스라인
go build ./... && go vet ./...

# 2. 현재 커버리지 (internal/hook 90% 목표의 출발점)
go test -cover ./internal/hook/... 2>&1 | tail -5

# 3. 린트 베이스라인 (NEW vs pre-existing 구분용)
golangci-lint run ./internal/hook/... ./internal/config/... 2>&1 | tail -20

# 4. matcher 실측 재확인 (라인 앵커가 아닌 content-token으로)
grep -n '"matcher"' internal/template/templates/.claude/settings.json.tmpl

# 5. 훅 AskUserQuestion 가드 베이스라인
go test ./internal/verify/... -run Boundary

# 6. 프로파일 해석기 임포트 가능성 확인 (internal/hook → internal/template 순환 없음)
go list -deps ./internal/hook/... | grep -c 'moai-adk/internal/template'
go list -deps ./internal/template/... | grep 'moai-adk/internal/hook' || echo "no reverse dep (OK)"

# 7. 로그 age-out 등록 지점 확인
grep -n "task-metrics.jsonl\|PruneObservationLogs" internal/hook/prune_logs.go
```

Pre-flight 6번은 **차단 조건**이다: `internal/template` → `internal/hook` 역방향 의존이 존재하면 REQ-AME-012의 해석기 직접 호출이 순환 임포트를 만든다. 그 경우 해석 로직을 제3 패키지로 승격하거나 인터페이스 주입으로 우회한다.

---

## §D 밀스톤

### M1 — 페이로드 실측 게이트 (선행 조건, 기능 구현 아님)

목적: PreToolUse가 Agent/Task에 대해 발화하는지, 그 `tool_input`이 무엇을 담는지를 **관측**한다.

절차:
1. 임시로 PreToolUse에 `"matcher": "Agent|Task"` 블록을 추가하고, 핸들러는 stdin 원본을 `.moai/logs/agent-payload-probe.jsonl`에 그대로 덤프하는 최소 프로브로 구현한다.
2. 실제 Agent spawn을 1회 이상 발생시킨다.
3. 덤프를 확인하고, 민감 필드(`prompt` 본문)를 절삭한 뒤 `internal/hook/testdata/agent_pretool_payload.json` 픽스처로 커밋한다.
4. 관측 결과 3문항(REQ-AME-002)의 답을 progress.md §E.2에 기록한다.

파일: `internal/hook/pre_tool.go`(프로브 분기), `internal/template/templates/.claude/settings.json.tmpl`, `internal/hook/testdata/agent_pretool_payload.json`

분기 처리(REQ-AME-003): 발화하지 않으면 M2 이후를 **중단**하고 D2 표의 대체안으로 재라우팅한 뒤 재승인을 받는다.

### M2 — 관측 계층

파일:
- `internal/hook/agent_model_guard.go` (신규) — `extractAgentSpawn` / `classifyAgentModel` / `appendAgentModelAudit`
- `internal/hook/pre_tool.go` — `ToolName ∈ {Agent, Task}` 분기 추가 (브랜치 가드 호출 지점 이후, 기본 allow 폴스루 이전)
- `internal/hook/prune_logs.go` — 신규 로그 age-out 등록 (REQ-AME-054)
- `internal/template/templates/.claude/settings.json.tmpl` + `.claude/settings.json` — Agent/Task matcher 블록 정식화
- `internal/hook/agent_model_guard_test.go` (신규)

제약: 이 밀스톤 종료 시점에 deny/ask 반환 경로는 코드에 **존재하지 않는다**(REQ-AME-015).

### M3 — 권고 계층

파일: `internal/hook/agent_model_guard.go`(권고 문자열 생성), `internal/hook/response.go` 소비(기존 advisory 채널 재사용)

제약: 종료 코드 불변(REQ-AME-021), `AskUserQuestion` 참조 0건(REQ-AME-022).

### M4 — opt-in 차단 게이트

파일:
- `internal/config/types.go` — `AgentModelGuardConfig` 추가 + `WorkflowConfig`에 필드
- `internal/config/defaults.go` — 기본값 `false`
- `internal/config/loader_*.go` — 섹션 로더 반영
- `internal/hook/pre_tool.go` — `agentModelGuardEnabled()` 헬퍼 (branchGuardEnabled 패턴 복제)
- `.moai/config/sections/workflow.yaml` + 템플릿 미러
- `internal/hook/agent_model_guard_test.go` — 게이트 ON/OFF × 4판정 매트릭스

제약: 차단은 `mismatch` 한정(REQ-AME-034), fail-open은 모든 불확실에 적용(REQ-AME-033), 비활성 경로에 추가 I/O 없음(REQ-AME-035).

### M5 — 규칙 가시성

파일 (D4 확정에 따라 (a) 또는 (b)):
- (a) `.claude/rules/moai/development/agent-spawn-model-injection.md` 신규 + 템플릿 미러
- (b) `.claude/rules/moai/core/agent-common-protocol.md` 절 추가 + 템플릿 미러
- 어느 경우든 `model-policy.md`는 `paths:` 스코프 유지, 스텁이 이를 교차 참조

제약: 스텁 ≤ 2KB, 매트릭스 셀·별칭 목록 재선언 금지(REQ-AME-042).

### M6 — 횡단 마감

파일: `internal/template/templates/**` 미러 전수, `make build` 산출물, 커버리지 보강 테스트

작업: 템플릿 중립성 가드 실행, `go test ./...` 전체 수트, `internal/hook` 커버리지 90% 달성, 훅 경계 가드 재실행.

---

## §E 밀스톤 의존 관계

```
M1 (측정 게이트) ──▶ M2 (관측) ──▶ M3 (권고) ──▶ M4 (차단 게이트)
                        │
M5 (규칙 가시성) ───────┴────────────────────────▶ M6 (횡단 마감)
```

- M1은 M2-M4의 **절대 선행**이다. M1 실패 시 M2 이후는 재설계 대상이다.
- M5는 M1-M4와 독립이며 병렬 가능하다 — Go 코드를 건드리지 않는다.
- M6은 전 밀스톤의 미러·빌드·커버리지를 한 번에 마감한다.

---

## §F 위험

| # | 위험 | 완화 |
|---|------|------|
| R1 | PreToolUse가 Agent에 발화하지 않음 | M1이 전용 게이트. 발화하지 않으면 진행 중단 + 재라우팅(REQ-AME-003) |
| R2 | 모든 spawn에 5초 타임아웃 훅이 붙어 지연 누적 | 별도 matcher 블록 + 핸들러 조기 반환. M6에서 spawn 왕복 지연 측정 |
| R3 | 순환 임포트 (`internal/hook` → `internal/template`) | Pre-flight 6번이 차단 조건 |
| R4 | 감사 로그 무한 성장 | REQ-AME-054 age-out 등록 |
| R5 | `prompt` 전문이 로그·픽스처로 유출 | 레코드는 에이전트명·모델·판정만 담고 prompt를 담지 않는다. 픽스처는 절삭 후 커밋 |
| R6 | 게이트 활성 시 오탐 차단으로 세션 wedge | `mismatch` 한정 + fail-open + 기본 OFF의 3중 방어 |
| R7 | 템플릿 미러에 SPEC 토큰 유출 | M6 중립성 가드; 훅 코드 주석의 SPEC ID는 `internal/` 소스에만 허용 |

---

## §G 안티패턴

- **AP-1 — 미검증 능력 위에 쌓기**: "PreToolUse가 Agent의 model을 본다"를 전제하고 M2부터 착수하는 것. M1이 존재하는 이유다.
- **AP-2 — 매트릭스 재선언**: 훅 안에 모델 별칭 목록이나 에이전트별 기대 모델을 리터럴로 적는 것. `ResolveAgentModelEffort` 호출이 유일한 경로다.
- **AP-3 — `missing` 차단**: 실측 99.4%를 막는다. 권고에 머문다.
- **AP-4 — 기존 matcher 확장**: `"Write|Edit|Bash"`를 `"Write|Edit|Bash|Agent"`로 넓히면 Agent spawn마다 Bash 전용 로직(브랜치 가드 등)의 조건 평가가 함께 돌고, 회귀 시 blast radius가 Write/Edit까지 번진다.
- **AP-5 — 인자 자동 변형**: 훅이 spawn 인자를 고쳐 모델을 주입하는 것. 관측 불가능한 침묵 개입이며 §C Out of Scope다.
- **AP-6 — 공허한 GREEN**: 판정 로직 테스트를 픽스처 없이 합성 페이로드로만 작성하는 것. M1 픽스처가 최소 1개의 실제 페이로드 기반 테스트를 강제한다.

---

## §H 교차 참조

- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` § Mechanical Enforcement — opt-in 게이트 + fail-open + 센티널 접두사의 하우스 선례
- `CLAUDE.local.md` §2 Template-First / §6 커버리지 / §7 훅 정책 / §25 템플릿 중립성
- `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary — 훅의 AskUserQuestion 금지
- `.claude/rules/moai/core/verification-claim-integrity.md` §1 — M1이 존재하는 근거(미관측 능력 주장 금지)
- GitHub issue #1376

---

## §G.1 Tier 편차 기록

Tier M의 표준 산출물은 spec.md + plan.md + acceptance.md(+ progress.md)이며 본 SPEC은 이를 그대로 따른다. 편차 없음.
