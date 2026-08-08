# SPEC-AGENT-MODEL-ENFORCE-001 — 인수 기준

---

## §A 판정 규약

### §A.1 공허한 GREEN 방지 (AC 명령 작성 규칙)

1. 표 셀 안에서 `\|`를 정규식 교대로 쓰지 않는다 — ERE에서 리터럴 파이프가 된다. 판정 명령은 표 밖 코드 블록에 둔다.
2. `.claude/`, `.moai/specs/`, `internal/template/templates/` 하위 산문은 영어 전용이므로 패턴 언어를 파일 언어에 맞춘다.
3. 부정형(부재) 단언은 반드시 대응하는 RED 픽스처로 실패를 증명한다 — "0건"이 검사기 오작동이 아님을 보인다.
4. 기존 가드(중립성·경계·린트)의 정규식을 AC 안에서 재구현하지 않는다. 가드 실행 자체가 권위다.
5. spawn 페이로드에 관한 단언은 M1이 커밋한 실제 픽스처를 입력으로 쓴다 — 합성 페이로드 전용 테스트는 AC를 만족하지 않는다.

### §A.2 실행 기준

- 모든 명령은 워크트리 루트 `.claude/worktrees/web-redesign`에서 실행한다.
- `go test` 판정은 종료 코드로 한다. 커버리지·린트는 명시된 임계값으로 한다.
- `golangci-lint`의 종료 코드는 리포지터리 전역이므로, 신규 결함 판정은 pre-flight 베이스라인과의 차분으로 한다.

---

## §B 실측 증거 (베이스라인)

plan-phase에서 실제 실행한 명령과 관측된 출력. run-phase는 이 값들을 재측정해 회귀를 판정한다.

```bash
# B1. PreToolUse matcher — Agent 부재
grep -n '"matcher"' internal/template/templates/.claude/settings.json.tmpl
# 관측: PreToolUse 블록의 matcher는 "Write|Edit|Bash" (Agent/Task 없음)

# B2. PostToolUse Agent 분기의 존재
grep -n '"Agent"' internal/hook/post_tool.go
# 관측: 169행 — logTaskMetrics(input) 호출만, model 미참조

# B3. 그 분기의 도달 불가 (F4)
grep -n 'hook post-tool' .claude/hooks/moai/handle-post-tool.sh
grep -n 'hook harness-observe' .claude/hooks/moai/handle-harness-observe.sh
# 관측: post-tool 래퍼는 matcher "Write|Edit|MultiEdit" 블록에만 등록;
#       matcher 없는 두 번째 블록은 harness-observe(다른 서브커맨드)

# B4. model-policy 경로 스코프 (F5)
sed -n '1,3p' .claude/rules/moai/development/model-policy.md
# 관측: paths: "**/.claude/agents/**"

# B5. 항상 로드 규칙 총량 (D4 토큰 비용 근거)
wc -c .claude/rules/moai/development/model-policy.md
# 관측: 27571 bytes
# 관측(집계): paths: 없는 규칙 13개 파일 = 197215 bytes

# B6. 에이전트 frontmatter 모델 핀 (F7)
grep -rn "^model:" .claude/agents/moai/*.md
# 관측: 11개 파일 중 10개 inherit, manager-git만 sonnet

# B7. spawn 페이로드 인자 키 분포 (F6)
# 로컬 세션 트랜스크립트 54개 전수 스캔 (python3, tool_use where name in {Agent, Task})
# 관측: 호출 156건 / model 키 보유 1건
#       키 빈도 {description:156, subagent_type:156, prompt:156,
#                run_in_background:137, name:14, model:1, isolation:2}
#       유일한 model 값: "opus" (subagent_type: hns-github-specialist)
```

---

## §C 품질 게이트

| 게이트 | 임계값 | 명령 |
|---|---|---|
| 빌드 | 성공 | `go build ./...` |
| 정적 분석 | 신규 0건 | `go vet ./...` |
| 린트 | 신규 0건 (베이스라인 차분) | `golangci-lint run ./internal/hook/... ./internal/config/...` |
| 테스트 | 전체 통과 | `go test ./...` |
| 커버리지 | `internal/hook` ≥ 90% | `go test -cover ./internal/hook/...` |
| 훅 경계 | AskUserQuestion 참조 0건 | `go test ./internal/verify/...` |
| 템플릿 중립성 | 통과 | `go test ./internal/template/ -run Neutrality -run InternalContentLeak` |

---

## §D AC 매트릭스

### §D.1 M1 — 페이로드 실측 게이트

**AC-AME-001**: Agent/Task spawn에 대한 실제 PreToolUse 페이로드 픽스처가 커밋되어 있다.
```bash
test -f internal/hook/testdata/agent_pretool_payload.json && \
  python3 -c "import json,sys; d=json.load(open('internal/hook/testdata/agent_pretool_payload.json')); print(sorted(d.keys()))"
# 기대: hook_event_name == "PreToolUse" AND tool_name ∈ {Agent, Task} AND tool_input 존재
```

**AC-AME-002**: 픽스처의 `tool_input`이 에이전트 식별자를 담는다.
```bash
go test ./internal/hook/ -run TestAgentSpawnFixtureCarriesSubagentType -v
# 테스트 본문: 픽스처 tool_input을 파싱해 subagent_type 비어있지 않음을 단언
```

**AC-AME-003**: REQ-AME-002의 3문항 관측 결과가 progress.md §E.2에 기록되어 있다.
```bash
grep -c "PreToolUse 발화\|subagent_type\|model 키" .moai/specs/SPEC-AGENT-MODEL-ENFORCE-001/progress.md
# 기대: 3 이상 (3문항 각각의 관측된 답 + 그 근거 명령)
```

**AC-AME-004**: 픽스처에 prompt 본문이 유출되지 않는다 (R5).
```bash
python3 -c "
import json; d=json.load(open('internal/hook/testdata/agent_pretool_payload.json'))
ti=json.loads(d['tool_input']) if isinstance(d['tool_input'],str) else d['tool_input']
p=ti.get('prompt',''); assert len(p)<=200, f'prompt not truncated: {len(p)}'; print('OK', len(p))"
```

### §D.2 M2 — 관측 계층

**AC-AME-010**: PreToolUse에 Agent/Task 전용 matcher 블록이 존재하고, 기존 Write|Edit|Bash 블록의 matcher 문자열은 변경되지 않았다.
```bash
go test ./internal/template/ -run TestSettingsPreToolUseAgentMatcher -v
# 테스트 본문: 렌더된 settings JSON을 파싱해
#   (a) PreToolUse 배열에 matcher가 Agent/Task를 포함하는 블록이 정확히 1개
#   (b) matcher "Write|Edit|Bash" 블록이 문자열 동등성으로 그대로 존재
#   (c) RED fixture: 기존 블록의 matcher를 넓힌 변형에서 (b)가 실패함을 별도 케이스로 증명
```

**AC-AME-011**: 핸들러가 Agent/Task spawn에서 에이전트 식별자와 선언 모델을 추출한다.
```bash
go test ./internal/hook/ -run TestExtractAgentSpawn -v
# 서브테스트: (a) 픽스처(실제 페이로드) (b) model 보유 합성 (c) model 부재 합성
#             (d) tool_input 파싱 실패 (e) subagent_type 부재
```

**AC-AME-012**: 판정이 해석기를 호출하며, 모델 별칭이나 에이전트별 기대값을 리터럴로 재선언하지 않는다.
```bash
grep -c "ResolveAgentModelEffort" internal/hook/agent_model_guard.go
# 기대: 1 이상
grep -nE '"(opus|sonnet|haiku|fable)"' internal/hook/agent_model_guard.go
# 기대: 0건 (AP-2 — 별칭 리터럴 재선언 금지)
```

**AC-AME-013**: 4치 판정이 명세대로 산출된다.
```bash
go test ./internal/hook/ -run TestClassifyAgentModel -v
# 테이블 테스트 4행: unmapped(카탈로그 밖) / missing(해석 구체·선언 부재)
#                    / mismatch(선언≠해석) / ok(선언==해석)
```

**AC-AME-014**: 각 spawn이 감사 로그에 한 줄씩 기록되며 필수 필드를 모두 담는다.
```bash
go test ./internal/hook/ -run TestAppendAgentModelAudit -v
# 테스트 본문: t.TempDir() 프로젝트 루트에 2회 spawn 처리 → 로그 2줄,
#   각 줄이 timestamp/session_id/agent/declared_model/resolved_model/verdict 보유
#   AND prompt 필드 부재(R5)
```

**AC-AME-015**: M2 단계에서 어떤 판정도 deny/ask를 반환하지 않는다.
```bash
go test ./internal/hook/ -run TestAgentModelObserveNeverBlocks -v
# 테스트 본문: 4판정 전부에 대해 Handle 결과가 allow 폴스루임을 단언
```

**AC-AME-016**: 로그 쓰기 실패가 spawn을 막지 않는다.
```bash
go test ./internal/hook/ -run TestAgentModelAuditWriteFailureFailsOpen -v
# 테스트 본문: 쓰기 불가 경로(읽기 전용 디렉터리 또는 미해석 프로젝트 루트)에서
#   Handle이 오류 없이 allow를 반환함을 단언
```

**AC-AME-017**: 신규 로그가 age-out 대상에 등록되어 있다.
```bash
go test ./internal/hook/ -run TestPruneObservationLogsIncludesAgentModelAudit -v
```

### §D.3 M3 — 권고 계층

**AC-AME-020**: `missing` / `mismatch`에 대해 권고 메시지가 방출되고, 메시지가 에이전트명과 해석 모델을 담는다.
```bash
go test ./internal/hook/ -run TestAgentModelAdvisoryContent -v
# 테스트 본문: missing/mismatch 각각에 대해 메시지가 비어있지 않고
#   에이전트 이름 부분문자열 + 해석 별칭 부분문자열을 포함함을 단언
#   AND ok/unmapped에 대해서는 메시지가 비어 있음을 단언
```

**AC-AME-021**: 권고 방출이 결정을 바꾸지 않는다.
```bash
go test ./internal/hook/ -run TestAgentModelAdvisoryDoesNotBlock -v
```

**AC-AME-022**: 훅 도메인 코드에 AskUserQuestion 참조가 0건이다.
```bash
go test ./internal/verify/...
# 기존 경계 가드 실행이 권위 (§A.1 규칙 4 — 정규식 재구현 금지)
```

### §D.4 M4 — opt-in 차단 게이트

**AC-AME-030**: 설정 키가 존재하고 배포 기본값이 false다.
```bash
go test ./internal/config/ -run TestDefaultAgentModelGuardDisabled -v
# 테스트 본문: DefaultConfig().Workflow.AgentModelGuard.Enabled == false
```

**AC-AME-031**: 게이트 OFF에서 4판정 전부 allow다.
```bash
go test ./internal/hook/ -run TestAgentModelGuardDisabledNeverBlocks -v
```

**AC-AME-032**: 게이트 ON + `mismatch`에서만 deny가 발화하고 사유가 센티널 접두사를 갖는다.
```bash
go test ./internal/hook/ -run TestAgentModelGuardEnabledDenyMatrix -v
# 테이블: (enabled × 4판정) 8행.
#   enabled=true & mismatch → deny AND reason 접두사 "AGENT_MODEL_VIOLATION:"
#   그 외 7행 → allow
```

**AC-AME-033**: 불확실 케이스가 전부 allow 폴스루한다 (fail-open).
```bash
go test ./internal/hook/ -run TestAgentModelGuardFailsOpen -v
# 서브테스트: tool_input 파싱 실패 / subagent_type 부재 / 미매핑 에이전트
#            / ConfigProvider nil / Config nil / 프로젝트 루트 미해석
# 전부 allow 단언
```

**AC-AME-034**: 게이트 ON이어도 `missing`은 차단되지 않는다.
```bash
go test ./internal/hook/ -run TestAgentModelGuardMissingNeverBlocked -v
# AC-AME-032의 매트릭스와 중복이 아니라, AP-3 회귀 전용 핀
```

**AC-AME-035**: 게이트 OFF 경로에 추가 파일 I/O가 없다.
```bash
go test ./internal/hook/ -run TestAgentModelGuardDisabledNoExtraIO -v
# 테스트 본문: 게이트 OFF에서 Handle 실행 후 t.TempDir() 하위에 신규 파일이
#   감사 로그 1개 외에는 생성되지 않음을 단언 (설정 재로드 등 부수 I/O 부재)
```

### §D.5 M5 — 규칙 가시성

**AC-AME-040**: per-spawn 모델 주입 의무를 담은 조각이 항상 로드 표면에 있다.
```bash
# 스텁 파일 경로는 plan.md D4 확정값을 사용
head -5 <스텁 경로> | grep -c '^paths:'
# 기대: 0 (paths: frontmatter 부재 = 항상 로드)
grep -c "model" <스텁 경로>
# 기대: 1 이상
```

**AC-AME-041**: 스텁이 압축돼 있고 model-policy.md 전체가 항상 로드로 전환되지 않았다.
```bash
wc -c <스텁 경로>
# 기대: 2048 이하
head -3 .claude/rules/moai/development/model-policy.md | grep -c '^paths:'
# 기대: 1 (기존 스코프 유지 — REQ-AME-043)
```

**AC-AME-042**: 스텁이 매트릭스 셀·별칭 목록을 재선언하지 않는다.
```bash
grep -cE 'opus|sonnet|haiku|fable' <스텁 경로>
# 기대: 0
grep -c "model-policy.md" <스텁 경로>
# 기대: 1 이상 (교차 참조 존재)
```

### §D.6 M6 — 횡단

**AC-AME-050**: `.claude/` 편집이 템플릿 미러와 같은 커밋에 있다.
```bash
git show --stat HEAD --name-only | grep -E '^\.claude/|^internal/template/templates/'
# .claude/ 편집이 있으면 같은 출력에 internal/template/templates/ 대응 경로가 함께 나타나야 한다
```

**AC-AME-051**: 배포 산출물에 내부 토큰이 없다.
```bash
go test ./internal/template/ -run InternalContentLeak -v
go test ./internal/template/ -run Neutrality -v
# 기존 가드 실행이 권위
```

**AC-AME-052**: 훅 배선이 하우스 규약을 따른다.
```bash
go test ./internal/template/ -run TestHookWrapperConventions -v
# 테스트 본문: 신규 PreToolUse 항목이
#   (a) 셸 래퍼 경로를 가리키고 (b) $CLAUDE_PROJECT_DIR가 따옴표로 감싸져 있고
#   (c) timeout == 5
```

**AC-AME-053**: 커버리지 임계값 충족.
```bash
go test -cover ./internal/hook/... 2>&1 | grep -E 'coverage:' | tail -1
# 기대: 90.0% 이상
```

**AC-AME-054**: 전체 수트 통과.
```bash
go test ./...
# 기대: 종료 코드 0
```

### §D.7 회귀 핀

**AC-AME-060**: Write/Edit/Bash PreToolUse 경로의 동작이 회귀하지 않는다.
```bash
go test ./internal/hook/ -run 'TestPreTool|TestBranchGuard' -v
# 기존 pre_tool_test.go / branch_guard_test.go 전량 통과
```

**AC-AME-061**: 브랜치 가드 결정 우선순위가 유지된다.
```bash
go test ./internal/hook/ -run TestPreToolDecisionPrecedence -v
# 테스트 본문: Bash + 위험 패턴 + 브랜치 상태 변경이 동시 성립할 때
#   기존 deny 사유가 우선함을 단언 (신규 Agent 분기 삽입이 순서를 흔들지 않음)
```

---

## §E Definition of Done

1. M1 픽스처가 커밋되어 있고, 3문항 관측 결과가 progress.md §E.2에 근거 명령과 함께 기록되어 있다.
2. §D의 모든 AC가 실행되어 PASS 증거(명령 + 출력)를 남겼다.
3. §C 품질 게이트 전 항목 통과.
4. `.claude/**` 편집과 `internal/template/templates/` 미러 + `make build` 산출물이 동일 커밋에 있다.
5. plan.md §B의 [NEEDS CLARIFICATION] 2건(D4 스텁 배치, D5 로그 형식)이 구현 착수 승인 전에 해소되고 결정이 plan.md에 기록되어 있다.
6. 게이트 기본값 `false` 상태에서 실제 세션의 Agent spawn이 차단되지 않음을 1회 이상 실측 확인.
