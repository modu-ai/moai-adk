# SPEC-PHASE-FRONTMATTER-OWNER-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-02
spec_version: "0.5.0"
amendment: scope-reduction   # v0.4.1 → v0.5.0, 사용자 명시 승인
audit_iteration: 4     # iter1 0.79 → iter2 0.76(STOP) → iter3 0.78(상승, STOP 미발동) → iter4 PASS 0.86
audit_verdict: PASS    # 0.86 — v0.4.1 (Tier M 임계값 0.80) 기준. v0.5.0은 범위 축소이므로 재감사 대상
tier: S                # M → S (범위 축소 후 문서 2개 + 미러 2개 = 4파일, Go 변경 0)
artifacts: [spec.md, plan.md]   # acceptance.md 삭제 — AC는 spec.md §3에 인라인 (Tier S 계약)
req_count: 7           # REQ-PFO-001/002/003/004/006/007/016
nfr_count: 1           # NFR-PFO-001
req_total: 8           # Tier S 상한 8에 정확히 도달
ac_count: 8            # AC-PFO-001/002/003/004/006/007/014/016 — Tier S 상한 8에 정확히 도달
dropped_milestones: [M4, M5]    # lint 규칙 + 9개 실물 정정 — SPEC-PHASE-FIELD-VALIDATION-001이 선착지
dropped_requirements: [REQ-PFO-005, REQ-PFO-008, REQ-PFO-009, REQ-PFO-010, REQ-PFO-011, REQ-PFO-012, REQ-PFO-013, NFR-PFO-002, NFR-PFO-003]
                       # 005는 REQ-PFO-004에 병합(삭제 아님); 나머지는 M4/M5 소관이라 삭제
remaining_milestones: [M7, M8]  # M1~M3은 이미 브랜치에 착지 (AC 001~004·006·007 PASS)
open_clarifications: 0 # plan.md §I
```

**남은 미착지 작업은 M7 한 건이다.** AC 8개 중 7개(001·002·003·004·006·007·014)는 이미 이 브랜치에서 PASS 관측되며, 현재 FAIL인 유일 항목은 **AC-PFO-016**이다 — 배포 문서가 착지한 가드를 오기하고 있고 그 정정이 아직 적용되지 않았다(spec.md §1.13). 016 없이 얻는 "8개 중 7개 PASS"는 이번 개정이 아무것도 바꾸지 않았다는 것과 구별되지 않는다.

**M4·M5 삭제 근거**: 병렬 SPEC `SPEC-PHASE-FIELD-VALIDATION-001`(커밋 `998744216`, PR #1285)이 lint 가드와 9개 실물 정정을 먼저 착지시켰다. 두 SPEC이 같은 매처를 소유하면 중복 원천이 되므로 가드 축을 형제 SPEC에 넘기고, 이 SPEC은 그 위층(저작 시점 계약·정정 소유자·문서 정확성)만 유지한다(spec.md §1.12).

**§1.8 정정**: 이전 판이 결정 축으로 지목한 "대소문자 구분"은 틀렸다. 실측 결과 결정 축은 **부분 문자열 대 값 전체 일치**이며, 값 전체 일치 위에서는 대소문자를 무시해도 오탐이 0이다(착지한 매처가 그 형태이며 이 SPEC이 명세했던 것보다 엄격하다).

미해결 클라리피케이션 마커는 **0건**이다(plan.md §I).

## §F Phase 4 Mode Selection

**입력 파라미터**

| 항목 | 값 |
|---|---|
| tier | M |
| scope (파일 수) | 약 8 — 규칙 문서 2 + 템플릿 미러 2 + `lint_phase.go` + `lint_phase_test.go` + `lint.go` 등록 + SPEC 프론트매터 9건(기계적 1행 편집) |
| 도메인 수 | 3 (rule 마크다운 / Go 소스 + 테스트 / SPEC 아티팩트) |
| 파일 언어 구성 | 마크다운 + Go 혼합 |
| 동시성 이득 | LOW — 구현 중심 작업이며 M4→M5 순서 의존이 존재 |

**모드 평가**

| 모드 | 선택 | 사유 |
|---|---|---|
| 1 trivial | 미선택 | 신규 lint 규칙 + 테스트 저작이 포함되어 의미 변경이 있다 |
| 2 background | 미선택 | Write/Edit를 수반하므로 read-only 조건 불충족 |
| 3 agent-team | 미선택 | RETIRED (tombstone) |
| 4 parallel | 미선택 | 도메인 3개이나 research-heavy가 아니라 coding-heavy — Anthropic coding-task parallelism caveat에 따라 Mode 5 우선 |
| 5 sub-agent | **선택** | 기본 폴백이자 coding-heavy 작업의 안전 경로 |
| 6 workflow | 미선택 | ~8 파일로 `≥ ~30 파일` 문턱 미달이며, 단일 균일 변환 규칙도 아니다(M1~M5가 서로 다른 성격) |

**Decision: sub-agent**

**근거**: 이 SPEC의 run-phase는 신규 Go lint 규칙 저작 + 테이블 드리븐 테스트 + 문서 계약 강화가 섞인 coding-heavy 작업이다. Anthropic의 coding-task parallelism caveat("most coding tasks involve fewer truly parallelizable tasks than research")에 따라 병렬 팬아웃보다 순차 sub-agent가 안전하다. 특히 M4(lint 규칙 + 반증 3축 관측)와 M5(실물 9건 정정) 사이에는 **역전 불가능한 순서 의존**이 있다 — M5가 코퍼스에서 라이프사이클 토큰을 제거하고 나면 규칙이 실물 데이터에서 발화하는 것을 관측할 창이 영구히 닫힌다. 병렬 실행은 이 순서를 보장하지 못한다.

**Implementation Kickoff Approval**: 통과 (사용자 명시 승인). 진행 방식은 **자율(autonomous)** — `/moai goal`로 AC 수렴 조건을 arm한 뒤 턴 단위 중단 없이 진행한다. 승인은 게이트 통과이며, 자율 진행 선택은 게이트 이후의 진행 방식 축이다(게이트 우회가 아니다).

## §E.2 Run-phase Evidence

작업 위치: 격리 워크트리 `.claude/worktrees/phase-frontmatter/`, 브랜치 `spec/phase-frontmatter-owner`, base `origin/main` SHA `54f92ef4b`.

### 사전 점검 (§C)

```
$ git rev-list --count --left-right origin/main...HEAD
0	0

$ go build ./... && GOOS=windows GOARCH=amd64 go build ./...
BUILD OK

$ golangci-lint run --timeout=2m
0 issues.

$ go test -count=1 ./internal/spec/...
ok  	github.com/modu-ai/moai-adk/internal/spec	20.392s

$ git ls-tree -r --name-only HEAD -- .moai/specs | grep -c '/spec\.md$'
563

$ git ls-files 'internal/template/templates/**/spec-assembly.md' 'internal/template/templates/**/spec-frontmatter-schema.md'
internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md
internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md
```

lint baseline은 `0 issues` — 이후 등장하는 어떤 지적도 NEW다.

### M1 — 정정 소유자 (AC-PFO-006 / AC-PFO-007)

```
$ S=.claude/rules/moai/development/spec-frontmatter-schema.md
$ grep -n -i 'phase.*correction\|phase 값 정정' "$S" | head
119:**Owner: `manager-spec`, reached through orchestrator re-delegation.** A `phase:` value correction — and any other non-transition frontmatter correction — is authored by `manager-spec`, the agent that already owns `spec.md` as canonical body content. ...

$ grep -c -i 'amendment.*not required\|amendment를 요구하지' "$S"      # (a)
1
$ grep -c '^#\+ .*Non-transition frontmatter corrections' "$S"          # (b0)
1
$ awk '/^## Status Transition Ownership Matrix/{f=1;next} /^## /{f=0} f' "$S" \
    | grep -c '^#\+ .*Non-transition frontmatter corrections'           # (b)
0
$ awk '/^## Status Transition Ownership Matrix/{f=1;next} /^## /{f=0} f' "$S" \
    | grep -v '^#' | grep -c -i 'Non-transition frontmatter corrections\|비전이 프론트매터 정정'   # (c)
1
```

AC-PFO-006 PASS (지정 에이전트 `manager-spec` — 유지 카탈로그 소속). AC-PFO-007 PASS — (b0) `1`과 함께 읽은 (b) `0`이므로 미구현이 아니라 배치 준수다.

### M2 — 값 계약 (AC-PFO-001 / AC-PFO-002 / AC-PFO-003)

```
$ grep -c 'typically release target' "$S"
0
$ grep -c -i '`phase:`.*\(prohibit\|MUST NOT\|금지\)\|\(prohibit\|MUST NOT\|금지\).*`phase:`' "$S"
1
$ sed -n '/Prohibited phase values/,/^## /p' "$S" | grep -oE '\b(plan|run|sync|mx)\b' | sort -u | wc -l | tr -d ' '
4
$ grep -c 'matchesModernPhase\|H-5' "$S"
1
```

AC-PFO-001 / 002 / 003 전부 PASS.

### M3 — 저작 시점 방지 (AC-PFO-004 / AC-PFO-005)

```
$ A=.claude/skills/moai/workflows/plan/spec-assembly.md
$ grep -n '`phase:' "$A" | grep -ci 'NEVER\|금지\|NOT a lifecycle'
1
$ sed -n '/Pre-write gate behavior/,/^- \.moai/p' "$A" | grep -ci 'prohibited value\|금지값\|lifecycle token'
1
```

AC-PFO-004 / 005 PASS.

### M4 — lint 규칙

**RED (구현 이전 관측 — E8).** 테스트와 픽스처만 존재하고 `PhaseValueRule`이 없는 상태에서 실행했다.

```
$ go test -run 'TestPhaseValueRule' -count=1 -v ./internal/spec/
=== RUN   TestPhaseValueRule
=== RUN   TestPhaseValueRule/plan_fires
    lint_phase_test.go:50: expected exactly 1 PhaseValueInvalid finding, got 0: []
=== RUN   TestPhaseValueRule/planning_does_not_fire
=== RUN   TestPhaseValueRule/sync_layer_does_not_fire
=== RUN   TestPhaseValueRule/runtime_hardening_does_not_fire
--- FAIL: TestPhaseValueRule (0.55s)
    --- FAIL: TestPhaseValueRule/plan_fires (0.15s)
    --- PASS: TestPhaseValueRule/planning_does_not_fire (0.13s)
    --- PASS: TestPhaseValueRule/sync_layer_does_not_fire (0.14s)
    --- PASS: TestPhaseValueRule/runtime_hardening_does_not_fire (0.13s)
=== RUN   TestPhaseValueRule_Registered
    lint_phase_test.go:82: PhaseValueInvalid absent — the rule is not registered in defaultRules
--- FAIL: TestPhaseValueRule_Registered (0.14s)
=== RUN   TestPhaseValueRule_StrictExitCode
    lint_phase_test.go:104: strict=true: HasErrors()=false, want true
--- FAIL: TestPhaseValueRule_StrictExitCode (0.25s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/spec	1.395s
```

세 비발화 서브테스트가 RED에서 이미 PASS인 것은 규칙 부재의 자명한 귀결이며, 판별력을 갖는 RED 신호는 `plan_fires` FAIL·등록 FAIL·strict FAIL 세 건이다.

**GREEN.**

```
$ go test -run 'TestPhaseValueRule' -count=1 -v ./internal/spec/
--- PASS: TestPhaseValueRule (0.98s)
    --- PASS: TestPhaseValueRule/plan_fires (0.18s)
    --- PASS: TestPhaseValueRule/planning_does_not_fire (0.50s)
    --- PASS: TestPhaseValueRule/sync_layer_does_not_fire (0.15s)
    --- PASS: TestPhaseValueRule/runtime_hardening_does_not_fire (0.14s)
=== RUN   TestPhaseValueRule_Registered
--- PASS: TestPhaseValueRule_Registered (0.15s)
=== RUN   TestPhaseValueRule_StrictExitCode
--- PASS: TestPhaseValueRule_StrictExitCode (0.30s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/spec	1.860s
```

**AC-PFO-009 (판정 4항).**

```
$ grep -c 'PhaseValueRule' internal/spec/lint.go                                    # (1)
1
$ go test -run 'TestPhaseValueRule' -count=1 -v ./internal/spec/ | grep -c '^--- PASS'  # (2)
3
$ grep -c 'func TestPhaseValueRule' internal/spec/lint_phase_test.go                # (3)
3
$ go test -run 'TestPhaseValueRule' -count=1 -v ./internal/spec/ \
    | grep -c -- '--- PASS: TestPhaseValueRule/.*does_not_fire'                     # (4)
3
```

AC-PFO-009 PASS — (3)이 `3`이므로 (2)의 PASS는 0매칭 공허가 아니다.

**[판정 시점 계약] AC-PFO-008 — M4 시점, 실물 대상.** M5가 이 파일을 정정하면 이 관측은 되돌릴 수 없이 사라진다.

```
$ go run ./cmd/moai spec lint .moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/spec.md
SEVERITY  CODE                  FILE                                               LINE  MESSAGE
--------  ----                  ----                                               ----  -------
WARNING   StatusGitConsistency  .moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/spec.md  1     SPEC SPEC-UPDATE-YAML-PRESERVE-001 frontmatter status 'draft' disagrees with git-implied status 'in-progress'
WARNING   PhaseValueInvalid     .moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/spec.md  1     phase "plan" is a lifecycle stage name, not a release or milestone target label; use the release this SPEC is aimed at (e.g. "v3.0.0")

0 error(s), 2 warning(s)

$ ... | grep -c 'PhaseValueInvalid'
1
```

AC-PFO-008 **PASS (계수 `1`)** — M4 시점 실물 관측. AC-PFO-010의 결합항은 이 기록을 참조한다.

**[판정 시점 계약] AC-PFO-011 — M4 시점, 실물 대상.**

```
$ T=.moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/spec.md
$ go run ./cmd/moai spec lint "$T" | grep 'PhaseValueInvalid' | grep -c 'WARNING'   # (a)
1
$ go run ./cmd/moai spec lint "$T" >/dev/null 2>&1; echo $?                          # (b)
0
$ go run ./cmd/moai spec lint --strict "$T" >/dev/null 2>&1; echo $?                 # (c)
1
```

AC-PFO-011 PASS — (a) `1` / (b) `0` / (c) `1`. 심각도 문자열은 strict에서도 `WARNING`으로 남으며, 승격의 실체가 종료 코드라는 §1.11의 실측과 일치한다.

**AC-PFO-008F — 반증 3축, 전부 관측.**

```
# (a) defaultRules 등록 제거
$ go run ./cmd/moai spec lint .moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/spec.md | grep -c 'PhaseValueInvalid'
0
# (a) 되돌린 뒤
1

# (b) Check 본문 무력화 (선두 `return nil`)
$ go run ./cmd/moai spec lint .moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/spec.md | grep -c 'PhaseValueInvalid'
0
# (b) 되돌린 뒤
1

# (c) 매칭 축 — 값 전체 일치를 strings.Contains 부분 문자열로 퇴화
$ go test -run 'TestPhaseValueRule' -count=1 -v ./internal/spec/ | grep -E '^    --- (PASS|FAIL)'
    --- PASS: TestPhaseValueRule/plan_fires (0.29s)
    --- FAIL: TestPhaseValueRule/planning_does_not_fire (0.21s)
    --- FAIL: TestPhaseValueRule/sync_layer_does_not_fire (0.19s)
    --- PASS: TestPhaseValueRule/runtime_hardening_does_not_fire (0.18s)
$ ... | grep -c -- '--- PASS: TestPhaseValueRule/.*does_not_fire'
1
# (c) 되돌린 뒤
3
```

세 축 모두 변형 시 판정이 무너지고 되돌린 뒤 복원되는 것을 관측했다. (c)에서 `runtime_hardening`이 여전히 PASS인 것은 대소문자 구분 부분 문자열이 `Run`을 잡지 못하기 때문이며, spec.md §1.8이 "대소문자 구분 부분 문자열도 현재 코퍼스에서는 우연히 9건"이라고 기록한 바와 정확히 일치한다 — 그 우연을 깨뜨리는 것이 `planning`과 `sync layer` 두 합성 픽스처다.

```
$ grep -c 'FALSIFICATION' internal/spec/lint_phase.go internal/spec/lint.go
internal/spec/lint_phase.go:0
internal/spec/lint.go:0
```

변형 잔재 0 — 세 반증의 되돌림이 완전하다.

### M5 — 9개 정정 (AC-PFO-012 / AC-PFO-013)

```
$ ... 9개 spec.md의 phase: 값 계수
9                       # AC-PFO-012 PASS

$ git diff --numstat -- .moai/specs/   (이 SPEC 자신 제외)
1	1	.moai/specs/SPEC-CI-LOOP-DEVONLY-001/spec.md
2	2	.moai/specs/SPEC-ENVKEY-ANTHROPIC-SSOT-001/spec.md
1	1	.moai/specs/SPEC-PIPELINE-FANOUT-ACTIVATION-001/spec.md
1	1	.moai/specs/SPEC-REF-SEO-ABSORB-001/spec.md
1	1	.moai/specs/SPEC-UPDATE-GUARD-EFFICACY-001/spec.md
2	2	.moai/specs/SPEC-UPDATE-REINSTALL-LOOP-002/spec.md
2	2	.moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/spec.md
2	2	.moai/specs/SPEC-WORKTREE-BRANCH-GUARD-001/spec.md
2	2	.moai/specs/SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001/spec.md
```

1행 파일은 `updated:`가 이미 실행일이라 `phase:` 한 줄만 바뀐 경우다. 어느 파일도 2행을 넘지 않는다 — §D 범위 규율(두 줄만) 준수.

```
$ LIFECYCLE_SCAN (§A 헬퍼, git show 형태) | wc -l
0                       # AC-PFO-013 PASS
$ git ls-tree -r --name-only HEAD -- .moai/specs | grep -c '/spec\.md$'
564                     # 판정 시점 모집단 N — 이 SPEC 산출물이 커밋되어 563 → 564
```

**B7 catalog.yaml 해시.** `make build` 후 `git status`에 catalog 변경 없음 — 9개 SPEC은 catalog 해시 대상이 아니므로 무효화가 발생하지 않았다. 재생성 판단 불요.

### M5 이후 재판정 (판정 시점 계약)

```
$ go run ./cmd/moai spec lint internal/spec/testdata/phase-token-plan/spec.md | grep -c 'PhaseValueInvalid'
1                       # AC-PFO-008 PASS (§C 발화 픽스처 대상)
$ T=internal/spec/testdata/phase-token-plan/spec.md
$ ... | grep 'PhaseValueInvalid' | grep -c 'WARNING'   → 1     # (a)
$ go run ./cmd/moai spec lint "$T" >/dev/null 2>&1; echo $?   → 0     # (b)
$ go run ./cmd/moai spec lint --strict "$T" >/dev/null 2>&1; echo $?  → 1     # (c)
                        # AC-PFO-011 PASS
```

### M6 — 회귀 검증 (AC-PFO-010 / AC-PFO-014 / AC-PFO-015)

```
$ go run ./cmd/moai spec lint > /tmp/pfo-full.txt 2>&1; echo $?
0
$ grep -c 'PhaseValueInvalid' /tmp/pfo-full.txt
0                       # AC-PFO-010 계수 — 결합항 AC-PFO-008은 위 M4 시점 기록 참조 → PASS
$ tail -1 /tmp/pfo-full.txt
0 error(s), 62 warning(s)   # AC-PFO-015 baseline 복귀 (error 0, warning 62)

$ go test -count=1 ./...
exit=0, FAIL 행 0개
  (1회차에서 internal/web가 `signal: terminated`로 종료되었으나 단독 재실행 `ok 1.707s`,
   전체 재실행 exit 0 / FAIL 0행 — 기존에 알려진 web 패키지 flaky이며 이 변경에 귀속되지 않는다.
   내 변경은 internal/web를 건드리지 않는다.)

$ go vet ./...                                → exit 0
$ golangci-lint run --timeout=2m              → 0 issues.   (C3 baseline 0 issues 동일, NEW 0건)
$ go build ./...                              → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...    → exit 0
$ go test -cover -count=1 ./internal/spec/... → coverage: 89.0% of statements
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/spec/ | grep -v '_test.go' | grep -v '//'
(출력 없음)

# AC-PFO-014 미러 패리티
$ diff .claude/rules/moai/development/spec-frontmatter-schema.md \
       internal/template/templates/.../spec-frontmatter-schema.md | wc -l   → 0
$ diff .claude/skills/moai/workflows/plan/spec-assembly.md \
       internal/template/templates/.../spec-assembly.md | wc -l             → 0
$ go test -count=1 -run 'TestInternalContentLeak|TestTemplateNeutrality|TestSplitHarnessNamespaceNoLeak' ./internal/template/
ok  	github.com/modu-ai/moai-adk/internal/template	3.204s
```

AC-PFO-014는 AC-PFO-001~005 PASS와 결합해서만 유효하며, 그 다섯 건은 위에 기록되어 있다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: BLOCKED-SUPERSEDED
run_complete_at: 2026-08-02
run_commit_sha: d27a274de      # 로컬 브랜치 HEAD — 미push (아래 차단 사유)
base_sha: 54f92ef4b
branch: spec/phase-frontmatter-owner
ac_pass_count: 16              # AC-PFO-001~015 + AC-PFO-008F 전부 PASS (판정 시점 계약 준수)
ac_fail_count: 0
preserve_list_post_run_count: 0   # 기존 lint 규칙 판정 동작 무변경 — 전수 lint 총계 0/62 동일
l44_pre_commit_fetch: "0 0"       # M1 착수 시점
l44_post_push_fetch: "1 6"        # push 이전 fetch — origin/main이 1커밋 전진, 아래 충돌의 발견 지점
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues (baseline 동일), 전수 lint error 0 / warning 62
cross_platform_build:
  host: exit 0
  windows_amd64: exit 0
coverage_internal_spec: "89.0%"
total_run_phase_files: 24
m1_to_mN_commit_strategy: "마일스톤당 1커밋 (M1~M5) + 플랜 산출물 랜딩 1커밋 = 6커밋, 전부 미push"
pushed: false
```

### 차단 사유 — 병렬 SPEC이 같은 결함을 이미 해결한 채 main에 착지했다

M6 검증 직후 push 직전 fetch에서 `origin/main`이 1커밋 전진한 것을 관측했고, 그 커밋이 이 SPEC의 M4·M5와 **같은 결함을 같은 방식으로** 해결한 병렬 SPEC이었다.

```
$ git log --oneline HEAD..origin/main
998744216 feat(SPEC-PHASE-FIELD-VALIDATION-001): validate the phase frontmatter value shape (#1285)
```

착지한 구현(`origin/main:internal/spec/lint.go`):

```go
if phase := strings.TrimSpace(fm.Phase); phase != "" && phaseWorkflowStageTokens[strings.ToLower(phase)] {
    findings = append(findings, Finding{
        ...
        Severity: SeverityError,
        Code:     "FrontmatterPhaseInvalid",
```

겹치는 축과 겹치지 않는 축을 구분해 기록한다.

| 축 | 이 SPEC (M1~M5) | 착지한 SPEC-PHASE-FIELD-VALIDATION-001 | 상태 |
|---|---|---|---|
| 값 계약 문서(schema SSOT) | M2가 규범화 + 금지값 절 + era H-5 결속 | **미변경** | **충돌 없음 — 이 SPEC만 보유** |
| 저작 체크리스트 / pre-write gate | M3 | **미변경** | **충돌 없음 — 이 SPEC만 보유** |
| 정정 소유자 규정 | M1 (manager-spec, 매트릭스 밖 별도 절) | **미변경** | **충돌 없음 — 이 SPEC만 보유** |
| lint 가드 | M4 `PhaseValueRule` / `PhaseValueInvalid` / warning + `Advisory:false` / 대소문자 **구분** | `FrontmatterSchemaRule` 내 분기 / `FrontmatterPhaseInvalid` / **error** / 대소문자 **무시** | **중복 — 같은 결함에 두 규칙** |
| 9개 실물 정정 | M5 | **이미 동일하게 `"v3.0.2"`로 정정 완료** | **중복 — 재적용** |
| 파일 경로 | `internal/spec/lint_phase_test.go` (package `spec_test`) | **같은 경로** (package `spec`, 전혀 다른 내용) | **하드 충돌 — merge conflict 확정** |

세 가지가 결정을 요구한다.

1. **경로 하드 충돌.** `internal/spec/lint_phase_test.go`가 양쪽에 서로 다른 내용으로 존재한다. 이 경로는 acceptance.md AC-PFO-009의 **명명 앵커 산출물 요건**이므로, 파일명을 바꾸면 AC가 fail-closed로 떨어진다 — 즉 재작업 없이 리베이스만으로 해소되지 않는다.
2. **가드 이중화.** 같은 결함에 코드가 다른 두 finding이 방출된다. 착지한 쪽이 error 심각도라 더 강한 게이트이고, `eraDemotableCodes` 제외로 grandfather 강등도 피한다 — 이 SPEC의 `Advisory:false` 설계와 목적은 같고 수단이 다르다.
3. **spec.md §1.8의 근거가 부분적으로 반증되었다.** §1.8은 "대소문자 무시 매칭은 실물 8건(`Runtime` 포함)에 즉시 오탐을 낸다"고 단언했으나, 그것은 **부분 문자열 매칭과 결합했을 때**만 참이다. 착지한 구현은 대소문자 무시 + **값 전체 일치**이므로 `"v3.0.0 — Phase 2 — Runtime Hardening"`은 `"run"`과 같지 않아 오탐이 없다. 실제로 그쪽이 이 SPEC보다 엄격하다 — `PLAN` / `Sync` 같은 대소문자 변형 오타까지 잡는다. 이 SPEC이 배제한 축이 실은 배제할 이유가 없었다.

따라서 push하지 않고 중단한다. 6개 커밋은 브랜치에 보존되어 있으며 워킹 트리는 깨끗하다(추적되지 않은 plan-audit 반복 보고서 4건 제외).

### 해소 — 범위 축소 (사용자 결정)

위 차단은 **범위 축소**로 해소되었다. 사용자가 네 선택지(범위 축소 / 전면 폐기 / 둘 다 유지 / 보류) 중 범위 축소를 명시적으로 선택했고, SPEC은 v0.5.0에서 Tier M → **Tier S**로 강등되었다(REQ 8 / AC 8, `acceptance.md` 삭제 후 spec.md §3 인라인).

```yaml
run_status: COMPLETE-REDUCED     # BLOCKED-SUPERSEDED 를 대체한다
resolution: scope-reduction
dropped_milestones: [M4, M5]     # 병렬 SPEC 998744216 이 동일 결함을 선행 해결
surviving_milestone: M7          # 값 계약이 착지한 가드를 정확히 기술하도록 정정
base_sha: 998744216              # origin/main 재베이스 (54f92ef4b 아님)
final_change_surface: 7          # 문서 2 + 미러 2 + SPEC 산출물 3
go_code_changed: 0
```

**§E.2의 M4·M5 증거는 삭제하지 않고 보존한다.** 그 작업은 실제로 수행되었고 16개 AC 전부 PASS했으며 반증 3축까지 관측되었다 — 다만 병렬 SPEC이 먼저 착지해 **이 PR에는 포함되지 않는다**. 기록으로서는 참이고 배송 주장으로서는 거짓이므로, 이 절이 그 경계를 명시한다.

브랜치는 `origin/main`(`998744216`) 위로 재구성되었다. M4가 만든 `internal/spec/lint_phase.go`·testdata 4건은 삭제되었고, `internal/spec/lint.go`·`internal/spec/lint_phase_test.go`와 M5가 정정한 9건은 전부 `origin/main` 내용으로 원복되어 형제 SPEC의 작업을 되돌리지 않는다.

**정정 반영**: 위 3번(§1.8 반증)은 v0.5.0에서 SPEC 본문에 반영되었다 — 결정 축은 대소문자가 아니라 **부분 문자열 대 값 전체 일치**이며, 스키마 문서의 값 계약도 착지한 가드의 실제 동작(대소문자 무시 + 값 전체 일치)을 기술하도록 M7에서 교체되었다.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: COMPLETE
sync_complete_at: 2026-08-02
sync_commit_sha: pending-backfill   # self-referential — populated by the follow-up backfill commit
run_commit_sha: d27a274de           # pre-squash local branch HEAD; origin's merge is 850de684c (PR #1286)
tier: S                             # M → S 범위 축소 (v0.5.0) — M4·M5는 SPEC-PHASE-FIELD-VALIDATION-001 선착지로 제외
changelog_entry_position: "### Added 최상단 (CHANGELOG.md)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed"   # 유일한 YAML frontmatter 소유 산출물 (Tier S: spec/plan/progress 중 spec.md만 frontmatter 보유)
  plan_md: "n/a — markdown header convention, no YAML frontmatter"
  progress_md: "n/a — markdown header convention, no YAML frontmatter"
  acceptance_md: "n/a — Tier S 계약: acceptance.md는 존재하지 않음 (AC는 spec.md §3에 인라인)"
canary_compliance_check:
  b12_pre_emission_duplicate: "grep -c 'SPEC-PHASE-FRONTMATTER-OWNER-001' CHANGELOG.md → 0 (사전) → 1 (커밋 후)"
  b12_ac_count_match: "spec.md §3 AC 인라인 = 8 (AC-PFO-001/002/003/004/006/007/014/016); §E.3 `ac_pass_count: 16`은 M4·M5 포함 pre-scope-reduction 작업 카운트 — v0.5.0 SSOT는 8"
  b12_file_paths_verified: "ls .moai/specs/SPEC-PHASE-FRONTMATTER-OWNER-001/{spec,plan,progress}.md → 3 files (acceptance.md 부재 확인)"
```

### Claim (주장)

이 커밋은 SPEC-PHASE-FRONTMATTER-OWNER-001의 **3단계 종료(sync-phase close)** 를 수행한다: 유일한 YAML-frontmatter 산출물인 `spec.md`의 `status:` 를 `in-progress → implemented → completed` 로 전이시키고, 본 `progress.md §E.4` 를 채우며, `CHANGELOG.md [Unreleased] → ### Added` 최상단에 엔트리를 추가한다. **코드 변경은 없다** — run-phase 코드는 PR #1286 (squash merge `850de684c`) 로 이미 `origin/main` 에 착지했다. 이 커밋은 산출물 전이 + §E.4 시그널 + CHANGELOG 엔트리만 운반한다.

### Evidence (증거)

동기화 커밋이 존재하기 전에는 자기 SHA를 알 수 없으므로, `sync_commit_sha` 필드는 `pending-backfill` 자리표시자로 두고 후속 커밋에서 채운다 (`9ec8d8464` "docs(SPEC-HOOK-TRACE-FLUSH-001): backfill sync_commit_sha 9d976c95b" 의 정준 패턴).

frontmatter 전이 (Edit 전/후):

```
$ git -C .claude/worktrees/phase-frontmatter-sync diff -- .moai/specs/SPEC-PHASE-FRONTMATTER-OWNER-001/spec.md
-status: in-progress
+status: completed
```

`updated:` 필드는 이미 `2026-08-02` (당일) 이므로 부가적인 갱신이 없다 — `spec.md` 프론트매터의 다른 필드는 manager-docs 금지 영역(본문)이거나 이미 당일 값이다.

검증 명령 (worktree 내):

```
$ git -C .claude/worktrees/phase-frontmatter-sync status --porcelain
 M .moai/specs/SPEC-PHASE-FRONTMATTER-OWNER-001/spec.md
 M .moai/specs/SPEC-PHASE-FRONTMATTER-OWNER-001/progress.md
 M CHANGELOG.md
$ git -C .claude/worktrees/phase-frontmatter-sync diff --name-only origin/main HEAD | grep '^internal/template/templates/' || echo NONE
NONE
```

### Baseline-attribution (baseline 귀속)

- **측정 기준**: 이 워크트리의 `sync/phase-frontmatter-owner` 브랜치 HEAD `850de684c` (== `origin/main`, divergence `0 0`)에서의 트리 상태.
- **run-phase 코드 baseline**: PR #1286 squash merge `850de684c`. `progress.md §E.3`의 `run_commit_sha: d27a274de`는 squash 이전 로컬 브랜치 HEAD이며, origin에 착지한 형태는 squash 메인 라인이다 — 이 분기는 `progress.md §E.3` 라인 350의 주석과 §E.4 YAML의 `run_commit_sha` 필드에 이미 기록되어 있어 "정정" 대상이 아니다.
- **AC 판정 baseline**: spec.md §3에 인라인된 8개 AC (AC-PFO-001/002/003/004/006/007/014/016). §E.3 `ac_pass_count: 16` 은 M4·M5를 포함한 pre-scope-reduction 작업 카운트며 v0.5.0 SSOT(8)와 충돌하므로, 본 §E.4는 **spec.md §3 = 8/8 PASS** 를 권위적 수치로 취급한다 (Gaps 절에 명시).

### Gaps (미검증)

- **acceptance.md 부재**: 이 SPEC은 Tier S 계약(AC를 spec.md §3에 인라인)으로 `acceptance.md` 를 두지 않았다 (`progress.md §E.1` 라인 13 명시). 따라서 매트릭스의 "4-artifact transition"은 이 SPEC에서 **3-artifact transition** (spec/plan/progress) 으로 실제로 축소되며, 이 중 YAML frontmatter를 가진 것은 `spec.md` 뿐이다. plan.md / progress.md는 마크다운 헤더 관례(`# SPEC-... — ...`)를 따르며 YAML frontmatter가 없다 — 전이할 `status:` 필드가 없다.
- **AC 카운트 분기**: spec.md §3 SSOT = 8개 AC. §E.3 `ac_pass_count: 16`은 v0.4.1 (Tier M) 시점의 작업 카운트가 v0.5.0 (Tier S) 축소 후에 갱신되지 않은 잔재이다. 두 수치가 모두 문서에 존재하며, 본 §E.4는 spec.md §3(8)을 권위로 삼고 §E.3의 16은 시대착오적 잔재로 기록한다 — 이 분기는 sync-phase 범위 밖(spec.md 본문 / §E.3 본문 수정은 manager-docs 금지)이므로, 기록만 하고 고치지 않는다.
- **M4·M5 증거 보존 경계**: `progress.md §E.2` (M4·M5 증거, 라인 ~269–420) 와 `§E.3` (run_complete_at / run_commit_sha / 차단 사유 / 해소 절)은 본 커밋에서 수정하지 않는다 — 그 영역은 manager-develop 소유이며 본 SPEC의 범위 축소 내역을 서술한다. 증거는 삭제되지 않고 보존되며, "이 PR에는 포함되지 않는다"는 경계가 §E.2 라인 420에 명시되어 있다.
- **`moai spec audit` / `moai spec lint`**: sync-phase 품질 게이트 검증은 본 §E.4 작성 시점에 실행하지 않았다 — 이 커밋의 diff가 `.moai/specs/` 마크다운과 `CHANGELOG.md` 만을 다루고 Go 코드를 0건 변경하므로, 코드 품질 게이트(`go test`, `golangci-lint`)는 baseline `850de684c` 와 동일하다. spec-lint가 `completed` 전이를 거부하는지 여부는 오케스트레이터 검증 배치에서 확인한다.

### Residual-risk (잔여 위험)

- **CHANGELOG 충돌 (병렬 세션)**: `[Unreleased] → ### Added` 최상단에 삽입한다. 병렬 BATCH-SYNC 세션이 동시에 같은 위치에 엔트리를 삽입하면 충돌이 발생할 수 있다 — pre-emission grep (`grep -c 'SPEC-PHASE-FRONTMATTER-OWNER-001' CHANGELOG.md` → 0) 을 커밋 직전에 수행해 중복을 방지했다. push 전 단계에서 추가 재검증이 필요하다.
- **§E.3 ↔ §E.4 카운트 분기의 향후 영향**: §E.3의 `16`과 §3의 `8` 분기는 drift detector나 lint가 카운트를 기계적으로 비교하는 경우 오탐을 유발할 수 있다 — 현재 그런 검사기는 없으며, era.go는 SHA 필드 매칭에 그치므로 분기가 era 분류에 미치는 영향은 0이다.
- **`completed → in-progress (amendment)` 가능성**: 본 SPEC이 `completed` 로 전이된 후, 만약 M4·M5의 영역(`internal/spec/lint_phase.go` 파일 경로 하드 충돌)이 향후 별도 SPEC으로 재추진된다면 이 SPEC 본문이 그 참조 대상이 될 수 있다 — 그 시나리오는 본 커밋 범위 밖이다.
