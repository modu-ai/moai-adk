# Evaluation Report — SPEC-V3R3-RETIRED-AGENT-001
Iteration: 1/3
Mode: final-pass (standard harness)
Evaluator: evaluator-active v1.0
Date: 2026-05-04
Overall Verdict: **FAIL**
Overall Score: 0.67

---

## Dimension Scores (rubric-anchored)

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 65/100 | FAIL | AC-RA-09 명확한 미구현; AC-RA-06/18 partial fail |
| Security (25%) | 88/100 | PASS | path traversal 방지, sensitive info leak 없음, YAML injection N/A |
| Craft (20%) | 82/100 | PASS | error wrapping, godoc, 한국어 주석 준수, coverage 양호 |
| Consistency (15%) | 78/100 | PASS | handler 패턴 일관, test naming 준수; factory comment stale |

---

## Dimension 상세

### Functionality (40%): 65/100 — FAIL

#### AC별 평가 (18 ACs)

**AC-RA-01** (manager-cycle.md exists + full frontmatter): **PASS**
- 증거: `internal/template/templates/.claude/agents/moai/manager-cycle.md` 11385B 존재
- frontmatter 검증: name, description, tools, model(sonnet), permissionMode(bypassPermissions), memory, skills, hooks 모두 확인됨
- `TestManagerCyclePresentInEmbeddedFS` PASS (min 5000B 충족)

**AC-RA-02** (manager-tdd.md retired stub 5개 표준 필드): **PASS**
- 증거: `internal/template/templates/.claude/agents/moai/manager-tdd.md` 확인
  - `retired: true` (boolean) ✓
  - `retired_replacement: manager-cycle` ✓
  - `retired_param_hint: "cycle_type=tdd"` ✓
  - `tools: []` ✓
  - `skills: []` ✓
- legacy `status: retired` 필드 REMOVED ✓
- body에 retirement reason + replacement + migration guide 포함 ✓

**AC-RA-03** (make build embedded FS): **PASS**
- 증거: progress.md `make build ✅ embedded.go regenerated`
- `TestManagerCyclePresentInEmbeddedFS` + `TestAgentFrontmatterAudit` 모두 PASS

**AC-RA-04** (SubagentStart hook block decision): **PASS**
- 증거: `agentStartHandler.Handle()` 구현됨 (subagent_start.go L186-246)
- retired: true → DecisionBlock, reason 포함 ✓
- TestAgentStartHandler_BlocksRetiredAgent PASS

**AC-RA-05** (launcher.go empty-object worktreePath 거부): **PASS**
- 증거: `internal/cli/worktree_validation.go` 구현 확인
- `validateWorktreeReturn(nil, "worktree", agentName)` → WORKTREE_PATH_INVALID ✓
- `validateWorktreeReturn(&{WorktreePath:""}, "worktree", agentName)` → WORKTREE_PATH_INVALID ✓
- TestValidateWorktreeReturn_RejectsEmptyObject, _RejectsNullPath PASS

**AC-RA-06** (path interpolation text/template): **PARTIAL FAIL**
- 증거: `constructPathUnsafe`는 test-only 함수로만 존재 (실제 callsite 0개)
- text/template 시연: `TestPathTemplateRejectsNonStringValue` PASS (typed struct 사용 시 "{}" 없음 확인)
- **FAIL 근거**: `validateWorktreeReturn`은 `"{}"` 문자열(비어있지 않음)을 통과시킴
  - `launcher_worktree_validation_test.go:142-145`: `"{}" worktreePath는 비어있지 않으므로 validation 통과 (향후 강화 예정)` 명시
  - AC-RA-06 요구: "empty object `{}` values produce a typed error rather than the literal string `{}` in the path"
  - 현재 구현에서는 worktreePath="{}" 인 경우 validation을 통과함

**AC-RA-07** (retired-rejection guard JSON + exit 2): **PASS**
- 증거: `agentStartHandler.Handle()` DecisionBlock 반환 + reason 구조 확인
- reason에 agentName, replacement, param_hint 포함 ✓
- TestAgentStartHandler_OutputFormat PASS (JSON 직렬화 + decision/reason 필드 확인)
- **주의**: 실제 hook shell exit code 2는 UNVERIFIED (Go layer에서만 검증됨)

**AC-RA-08** (unknown agent bypass): **PASS**
- 증거: `loadAgentFrontmatter` file not found → found=false → allow
- TestAgentStartHandler_AllowsUnknownAgent PASS

**AC-RA-09** (factory.go dispatch for agent-start): **FAIL**
- 증거: `internal/hook/agents/factory.go` 실제 코드 확인:
  ```go
  switch agent {
  case "ddd": ...
  case "tdd": ...  // ← manager-tdd legacy case 여전히 존재
  // "cycle", "agent" case 없음
  }
  ```
- **factory.go에 "agent-start" 또는 "cycle" case 없음**
- AC-RA-09 요구: "file contains a switch case branch dispatching case 'agent-start' to NewAgentStartHandler()"
- progress.md 정당화: "NewAgentStartHandler() 생성자 존재만으로 REQ-RA-009 acceptance criterion 만족"
- 이 정당화는 AC 텍스트와 불일치: AC는 명시적으로 factory.go에 case 추가를 요구함
- TestAgentStartHandler_RoutesViaFactory는 factory dispatch를 검증하지 않고 생성자만 검증 (AC 요구사항보다 약한 테스트)
- **추가 발견**: factory.go의 `@MX:NOTE` comment에 여전히 `tdd` handler가 나열됨 (L21): "Supported agents: ddd, tdd, backend..." — cycle/agent-start 미추가

**AC-RA-10** (WORKTREE_PATH_INVALID sentinel with context): **PASS**
- 증거: `WorktreePathInvalidError.Error()` = `"WORKTREE_PATH_INVALID: agent=%q reason=%q"`
- agentName + reason 포함 ✓
- `errors.Is(err, ErrWorktreePathInvalid)` 지원 ✓
- TestValidateWorktreeReturn_RejectsEmptyObject + TestValidateWorktreeReturn_RejectsNullPath PASS

**AC-RA-11** (retired stub body: reason + replacement + migration): **PASS**
- 증거: manager-tdd.md body 확인
  - retirement reason: "Retired (SPEC-V3R3-RETIRED-AGENT-001)" ✓
  - replacement: "manager-cycle with cycle_type=tdd" ✓
  - migration table: | Old Invocation | New Invocation | ✓
  - "Why This Change" 섹션 ✓

**AC-RA-12** (retired-rejection guard ≤500ms): **PASS**
- 증거: progress.md "100회 반복 평균: 0ms (총 5.556ms), 단일 호출 평균: ~0.056ms"
- TestAgentStartHandler_PerformanceUnder500ms PASS
- 목표 ≤500ms 대비 약 9000배 여유

**AC-RA-13** (6개 문서 참조 substituted): **PASS (with caveat)**
- 증거: TestNoOrphanedManagerTDDReference PASS — 6개 파일에서 "manager-tdd" 문자열 0건
- progress.md M5: 7 references across 6 files 완료
- **Caveat (minor)**: template CLAUDE.md §4 Manager Agents 목록 (L108): `"spec, ddd, tdd, docs, quality, project, strategy, git"` — `tdd` 단축어가 제거되지 않고 `cycle`이 추가되지 않음
  - 그러나 AC-RA-13의 grep 기준은 `"manager-tdd"` 전체 문자열이므로 기술적으로 PASS
  - 실질적으로 §4 목록이 incomplete updated 상태

**AC-RA-14** (moai agents list --retired 서브커맨드): **DEFERRED — PASS (Branch B)**
- 증거: progress.md에 명시적 deferral 기록 없으나 spec.md §5.4 REQ-RA-014는 "P2; deferred"로 명시
- Implementation은 deferral로 처리됨 (구현 안 됨)
- Branch B 조건: 사용자 명시적 승인 필요 — UNVERIFIED (session 외 결정)
- Evaluator 판단: P2 Optional으로 명시된 REQ이므로 deferral acceptable

**AC-RA-15** (CI rejects RETIREMENT_INCOMPLETE): **PASS**
- 증거: `TestRetirementCompletenessAssertion` 구현됨
  - manager-tdd replacement manager-cycle 존재 확인 ✓
  - 모든 retired agents의 교체 파일 존재 확인 ✓
  - TestAgentFrontmatterAudit PASS ✓

**AC-RA-16** (manager-cycle spawn via Agent() succeeds): **UNVERIFIED**
- 증거: integration test 언급됨 (acceptance.md: "Integration test at M5 in /tmp/test-project-retired-agent-001")
- progress.md에 manual integration test 완료 기록 없음
- Evaluator: 코드 레벨에서 검증 불가 (Claude Code runtime 의존). UNVERIFIED

**AC-RA-17** (manager-tdd spawn blocked at SubagentStart layer): **PASS (unit), UNVERIFIED (integration)**
- 증거: TestAgentStartHandler_BlocksRetiredAgent PASS — retired stub detection 검증됨
- Go hook handler 구현 확인됨
- 실제 Claude Code SubagentStart hook 동작 (exit code 2 propagation) 은 integration test 필요 — UNVERIFIED

**AC-RA-18** (empty-object worktreePath → WORKTREE_PATH_INVALID): **PARTIAL FAIL**
- 증거: `TestPathTemplateRejectsNonStringValue` 내 명시적 인정 (L142-145):
  ```go
  // "{}"가 통과되는 것은 현재 스펙. 향후 path sanitization에서 처리.
  t.Logf("현재 구현: '{}' worktreePath는 비어있지 않으므로 validation 통과 (향후 강화 예정)")
  ```
- AC-RA-18 critical assertion: "the validation layer raises WORKTREE_PATH_INVALID error before any path interpolation"
- 현재 구현에서 worktreePath="{}" 는 validateWorktreeReturn 통과 → path interpolation에 진입 가능
- 이것은 Layer 4 (path string interpolation `{}/{}`) 방어에 gap이 있음을 의미

#### 5-Layer Defect Chain 차단 평가

| Layer | 차단 효과 | 판정 |
|-------|-----------|------|
| Layer 1 (retired stub frontmatter) | manager-tdd.md 표준화 + SubagentStart guard | **EFFECTIVE** |
| Layer 2 (worktree allocation timing) | validateWorktreeReturn nil/empty string 차단 | **PARTIAL** (nil/empty 차단, "{}" 미차단) |
| Layer 3 (auto-fallback propagation) | validateWorktreeReturn 존재, 그러나 callsite 0개 | **NOT WIRED UP** |
| Layer 4 (path interpolation `{}/{}`) | text/template 시연됨, validateWorktreeReturn이 "{}" 통과 | **PARTIAL** |
| Layer 5 (stream idle) | 명시적 out-of-scope | N/A |

---

### Security (25%): 88/100 — PASS

**YAML frontmatter injection (OWASP A03)**: PASS
- `parseAgentFrontmatter`는 user-controlled `agentName`으로 파일 경로를 구성하지만 YAML 파싱 결과를 struct에만 담음
- agentName → file path → yaml.Unmarshal → struct 필드 참조: injection 경로 없음

**Path traversal (OWASP A01)**: PASS
- `subagent_start.go L194-198`: `strings.Contains(agentName, "/") || strings.Contains(agentName, "..")` 검증
- `filepath.Join(projectDir, ".claude", "agents", "moai", agentName+".md")`: slash 없는 agentName → traversal 불가

**Sensitive info leak**: PASS
- reason 필드에 포함: `agentName`, `fm.RetiredReplacement`, `fm.RetiredParamHint` (모두 agent 파일에서 읽은 값)
- stdin JSON의 `last_assistant_message` 등 민감 필드는 reason에 포함되지 않음 (subagent_start.go SPEC 제약 준수)

**Sentinel error info leak**: PASS
- `WorktreePathInvalidError.Error()`: agentName과 reason만 포함, 내부 경로 노출 없음

**Minor finding**:
- `subagent_start.go L260-264`: `os.IsNotExist(err)` 체크 후 `return nil, true, fmt.Errorf(...)` — file exists but unreadable 케이스에서 found=true 반환. 이것은 적절한 보안 결정이나 caller가 error를 무시(allow로 처리)하는 것은 잠재적 bypass 가능성이 있으나 fail-safe 관점에서 acceptable.

---

### Craft (20%): 82/100 — PASS

**Test coverage (추정)**: PASS
- 신규 함수별 직접 테스트 확인됨:
  - `agentStartHandler.Handle()`: 5개 테스트 (retired/active/unknown/perf/format)
  - `parseAgentFrontmatter()`: 간접 (TestAgentFrontmatterAudit)
  - `loadAgentFrontmatter()`: 간접 (Handle 테스트를 통해)
  - `validateWorktreeReturn()`: 5개 직접 테스트
  - `TestManagerCyclePresentInEmbeddedFS`, `TestManagerCycleFrontmatterValid`: 각 1개
- 추정 coverage: ~80-85% (factory.go dispatch 미구현 부분 포함 시)

**Error wrapping**: PASS
- `fmt.Errorf("에이전트 파일 읽기 실패 %s: %w", path, err)` ✓
- `fmt.Errorf("frontmatter 파싱 실패 %s: %w", path, err)` ✓
- `fmt.Errorf("YAML 언마샬 실패: %w", err)` ✓

**Nil-pointer guards**: PASS
- `if result == nil` (worktree_validation.go L74)
- `if h.cfg == nil` (subagent_start.go L75)

**context.Context propagation**: PASS
- `Handle(ctx context.Context, input *HookInput)` 서명 준수

**godoc completeness**: PASS
- 모든 exported 타입/함수에 주석 존재 (agentStartHandler, NewAgentStartHandler, WorktreePathInvalidError, ErrWorktreePathInvalid, validateWorktreeReturn)

**코드 주석 언어**: PASS
- `language.yaml code_comments: ko` 준수 — 한국어 주석 일관되게 사용됨

**minor craft issues**:
- `_ = strings.Contains` (agent_start_test.go L270) — 컴파일러 경고 방지 위한 불필요한 코드, 사소하지만 clean code 원칙 미준수
- `worktree_validation.go L19-21`: 주석에 "M4 완료 후 test 파일의 로컬 정의를 이 타입 참조로 대체한다" — 이미 대체가 완료된 상태인데 미래 시제로 기록됨 (staleness)

---

### Consistency (15%): 78/100 — PASS

**Handler 패턴 준수**: PASS
- `agentStartHandler`가 `subagentStartHandler`와 동일한 `EventType() + Handle()` 인터페이스 구현
- file 위치도 동일 패키지(`package hook`) 내 통합 — Option A 결정으로 clean integration

**Naming convention**: PASS
- `agentStartHandler`, `NewAgentStartHandler`, `agentFrontmatter` — 기존 패턴 일관
- `WorktreePathInvalidError`, `ErrWorktreePathInvalid` — Go 관용 error naming 준수

**Test naming**: PASS
- `TestAgentStartHandler_BlocksRetiredAgent`, `TestAgentStartHandler_AllowsUnknownAgent` — 기존 `TestXxx_<Behavior>` 패턴 준수

**Pattern inconsistency (FAIL portion)**:
- `factory.go @MX:NOTE` (L21): "Supported agents: ddd, tdd, backend..." — `tdd` handler가 남아있고 `cycle`이 추가되지 않음 (stale comment)
- factory.go에 `case "tdd":` 여전히 존재 (manager-tdd는 retired되었으나 factory에 남아있어 cycle으로 교체됨을 반영 안 됨)
- local `.claude/rules/moai/core/agent-hooks.md` (project root): 여전히 `manager-tdd` 행 포함 (template은 manager-cycle로 교체됨) — Template-First 패턴에서 예상 가능하나 inconsistency 존재

**Template-First discipline**: PASS
- `make build` embedded.go 재생성 확인됨
- template 파일들의 변경이 embedded FS에 반영됨

---

## Anti-Pattern Findings (cap-triggering)

해당 없음 — cap-triggering anti-pattern 발견되지 않음

단, 다음 주의사항 기록:

**Premature abstraction 경계선**: `validateWorktreeReturn` + `constructPathUnsafe` — callsite 0개로 현재 wire-up 없음. progress.md에 "후속 SPEC에서 wire-up 예정"으로 문서화되어 있어 적절한 planning scope cut으로 판단. SPEC §2.1 In Scope에 REQ-RA-005 명시되어 있으므로 구현 자체는 required이나, wire-up 없이 standalone helper로만 존재하는 것은 기능 완성도 gap.

**Claiming Without Evidence**: AC-RA-09 관련 — progress.md가 "NewAgentStartHandler() 생성자 존재만으로 REQ-RA-009 만족"을 주장하나 AC 텍스트와 불일치. 이것은 evidence 없이 AC 충족을 선언한 사례.

---

## Defects Found

| ID | Severity | Location | 설명 | 권장 조치 |
|----|----------|----------|------|-----------|
| D-EVAL-01 | HIGH | `internal/hook/agents/factory.go` | **AC-RA-09 FAIL**: factory.go에 `case "agent-start":` (또는 `case "cycle":`) 브랜치가 없음. REQ-RA-009는 factory dispatch를 명시적으로 요구하며, AC-RA-09는 "file contains a switch case branch dispatching 'agent-start' to NewAgentStartHandler()"를 검증 방법으로 명시했으나 미구현. `TestAgentStartHandler_RoutesViaFactory`가 factory dispatch가 아닌 생성자만 검증하여 gap이 test에 의해 드러나지 않음. | factory.go에 `case "cycle":` 또는 `case "agent-start":` 추가. 관련 factory `@MX:NOTE` 주석 업데이트 |
| D-EVAL-02 | MEDIUM | `internal/cli/worktree_validation.go:82-87`, `internal/cli/launcher_worktree_validation_test.go:142-145` | **AC-RA-06/18 Partial**: `validateWorktreeReturn`은 empty string("")은 거부하지만 `"{}"` 문자열은 통과시킴. AC-RA-18 critical assertion: "the validation layer raises WORKTREE_PATH_INVALID before any path interpolation" — `"{}"` worktreePath에 대해 성립하지 않음. 테스트가 이를 명시적으로 인정하고 향후 처리로 deferral. | `validateWorktreeReturn`에 `strings.Contains(result.WorktreePath, "{}")` 또는 `/^\{.*\}$/` 패턴 검증 추가 |
| D-EVAL-03 | LOW | `internal/template/templates/CLAUDE.md:108` | CLAUDE.md §4 Manager Agents 목록: `"spec, ddd, tdd, docs, quality, project, strategy, git"` — `tdd` 단축어가 `cycle`로 교체되지 않음. AC-RA-13 grep 기준(`manager-tdd` 전체 문자열)은 통과하지만 실질적으로 §4 목록이 stale 상태. | `tdd` → `cycle` 교체 |
| D-EVAL-04 | LOW | `internal/hook/agents/factory.go:21-22` (`@MX:NOTE`), `factory.go:35` (`case "tdd":`) | factory.go의 MX annotation과 switch case에 `tdd`가 남아있음. manager-tdd retirement와 manager-cycle 추가가 factory에 반영되지 않아 future maintainer가 혼동할 수 있음. | `@MX:NOTE` 업데이트, `case "tdd":` 유지 여부 재검토 (cycle-* actions은 `case "cycle":` 추가 필요) |
| D-EVAL-05 | INFO | `internal/template/templates/.claude/rules/moai/core/agent-hooks.md` vs `.claude/rules/moai/core/agent-hooks.md` | template 버전은 manager-cycle 행으로 교체됨. project root local 버전은 여전히 manager-tdd 행 포함. Template-First 원칙에서 예상 가능한 sync lag이나 `make install` 또는 수동 sync로 해소 권장. | `moai update` 실행 또는 manual copy |

---

## Verdict Rationale

### FAIL 판정 근거

1. **D-EVAL-01 (HIGH)**: AC-RA-09는 factory.go에 명시적 dispatch case 추가를 요구했다. 이것이 구현되지 않았다. progress.md의 "생성자 존재만으로 충분" 정당화는 AC 텍스트를 충족하지 않는다. REQ-RA-009의 "factory shall route to the new handler" 요구사항이 구현되지 않은 것이다.

2. **D-EVAL-02 (MEDIUM)**: AC-RA-18의 critical assertion이 `"{}"` 문자열에 대해 성립하지 않는다. 이것은 5-layer defect chain의 Layer 4 방어가 partial임을 의미한다. 테스트가 이 gap을 명시적으로 인정하고 있어 "claiming without evidence" 없이 고의적 scope cut으로 기록되어 있으나, AC 요구사항은 충족되지 않았다.

### PASS 항목 (강점)

- Layer 1 (retired stub frontmatter) 차단: 완전하고 효과적
- SubagentStart hook 성능: 9000배 여유 (0.056ms vs 500ms target)
- error 처리 패턴: 일관되고 idiomatic Go
- 16개 REQ 중 12개 AC 명확 PASS
- TDD RED-GREEN-REFACTOR 사이클 준수 (TestNoOrphanedManagerTDDReference RED→GREEN 확인됨)

### Recommendation

**requires fix** — 두 가지 수정이 필요하다:
1. (필수) factory.go에 `case "cycle":` 또는 `case "agent-start":` 추가 + TestAgentStartHandler_RoutesViaFactory 강화
2. (권장) validateWorktreeReturn에 `"{}"` worktreePath 거부 로직 추가

수정 후 re-evaluation 필요. Layer 1-4 방어 완성도가 높아질 것으로 예상됨.

---

VERDICT: FAIL
