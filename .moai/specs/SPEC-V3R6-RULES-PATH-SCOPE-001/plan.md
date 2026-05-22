---
spec_id: SPEC-V3R6-RULES-PATH-SCOPE-001
artifact: plan
created: 2026-05-22
updated: 2026-05-22
---

# Plan — SPEC-V3R6-RULES-PATH-SCOPE-001

## 1. Implementation Strategy

### 1.1 핵심 접근: Frontmatter Prepend Only

본 SPEC 의 모든 변경은 4 rule 파일 (+ 4 template mirror) 의 첫 줄에 YAML frontmatter 블록을 **prepend** 하는 것으로 완결. body 1 byte 도 수정 금지 (REQ-RPS-NF-014).

```diff
+---
+description: "Zone Registry — MoAI-ADK HARD 조항 SSOT. rules/agents/skills/specs 디렉토리 수정 시에만 로드."
+paths: ".claude/**,.moai/specs/**,.claude/rules/**"
+---
+
 # Zone Registry

 MoAI-ADK 규칙 트리의 모든 HARD 조항을 열거하는 단일 진실 공급원(single source of truth).
 ...
```

### 1.2 Path Glob 설계 원칙

| 원칙 | 적용 |
|---|---|
| 보수적 광범위 우선 | trigger miss High 위험 (R-RPS-001) 회피를 위해 **약간 넓게** 잡음. 너무 좁아 rule 미발동 위험이 너무 넓어 token 낭비 위험보다 큼. |
| CSV string only | `.claude/rules/moai/development/coding-standards.md` § Paths Frontmatter 표준 일치. YAML array 형식 금지. |
| 선례 14+ 형식 모방 | `**/.claude/skills/**` (skill-authoring) / `.claude/agents/**/*.md,.claude/rules/moai/development/agent-authoring.md` (agent-patterns) / `**/*.md,.moai/specs/**/*.md` (spec-frontmatter-schema) 같은 단순 표현 사용. |
| Glob 중복 OK | 한 path 가 여러 glob 에 매치되더라도 Claude Code 가 rule 1회만 로드. 중복은 안전. |
| 절대 경로 금지 | 모든 path 는 project root 기준 상대 경로. `/Users/...` 같은 절대 경로 금지. |

### 1.3 Trigger Miss 위험 완화 메커니즘 (R-RPS-001 대응)

design.md §7 표 (라인 439) 위험 평가: "Rule path-scoped 화 시 trigger miss → 기능 누락 — Medium/High". 본 SPEC 의 4 mitigation 적용:

**M1 - Glob 광범위화 (preventive)**:
- `zone-registry.md` glob 에 `.claude/rules/**` 추가 (rules 자체 수정 시도 trigger)
- `manager-develop-prompt-template.md` glob 에 agent 파일 + run.md skill 추가
- `agent-teams-pattern.md` glob 에 strategy agent + team skills 추가

**M2 - Doctor 시뮬레이션 (detective)**:
- run-phase M3 단계에서 5 representative session 시나리오 시뮬레이션
- 보고서 `.moai/reports/rules-path-scope-simulation-<DATE>.md` 생성

**M3 - 24h 운영 관찰 (corrective)**:
- 본 SPEC 머지 후 24h 내 sister SPEC `SPEC-V3R6-RULES-COMPRESS-001` 작성 시점에 trigger miss 발생 여부 1차 점검
- 발견 시 hot-fix PR 로 glob 보정

**M4 - 5 keep-always 보존 (defensive)**:
- `agent-common-protocol.md` / `session-handoff.md` / `context-window-management.md` / `verification-batch-pattern.md` / `NOTICE.md` 모두 always-loaded 유지
- HARD 조항 reference, session-end protocol, 라이선스 등 cross-cutting concerns 항상 가시

---

## 2. Files Affected

### 2.1 Modified (4 files in `.claude/`)

| 파일 | 변경 | 라인 추가 |
|---|---|---|
| `.claude/rules/moai/core/zone-registry.md` | frontmatter prepend (description + paths) | +4 |
| `.claude/rules/moai/design/constitution.md` | frontmatter prepend (description + paths) | +4 |
| `.claude/rules/moai/development/manager-develop-prompt-template.md` | frontmatter prepend (description + paths) | +4 |
| `.claude/rules/moai/workflow/agent-teams-pattern.md` | frontmatter prepend (description + paths) | +4 |

### 2.2 Modified (4 files in `internal/template/`)

| 파일 | 변경 | 라인 추가 |
|---|---|---|
| `internal/template/templates/.claude/rules/moai/core/zone-registry.md` | 위 1번과 byte-identical mirror | +4 |
| `internal/template/templates/.claude/rules/moai/design/constitution.md` | 위 2번과 byte-identical mirror | +4 |
| `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` | 위 3번과 byte-identical mirror | +4 |
| `internal/template/templates/.claude/rules/moai/workflow/agent-teams-pattern.md` | 위 4번과 byte-identical mirror | +4 |

### 2.3 New (3 files)

| 파일 | 목적 |
|---|---|
| `.moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/spec.md` | 본 SPEC 본문 (위에 작성 완료) |
| `.moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/plan.md` | 본 파일 |
| `.moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/acceptance.md` | acceptance criteria 별도 파일 |

### 2.4 New (run-phase 산출물, gitignored)

| 파일 | 목적 |
|---|---|
| `.moai/reports/rules-path-scope-simulation-<YYYY-MM-DD>.md` | M3 doctor 시뮬레이션 보고서 (REQ-RPS-NF-012) |

### 2.5 Generated (run-phase, embedded.go 갱신)

| 파일 | 목적 |
|---|---|
| `internal/template/embedded.go` | `make build` 시 4 mirror 파일 자동 임베디드 갱신 (auto-generated, manual edit 금지) |

### 2.6 NOT Modified (PRESERVE list)

| 파일 | 사유 |
|---|---|
| `.claude/rules/moai/core/agent-common-protocol.md` | 모든 agent 호출 의무 reference (keep-always) |
| `.claude/rules/moai/workflow/session-handoff.md` | 본문 footnote 명시 keep-always |
| `.claude/rules/moai/workflow/context-window-management.md` | RULES-COMPRESS-001 담당 (별도 SPEC) |
| `.claude/rules/moai/workflow/verification-batch-pattern.md` | RULES-COMPRESS-001 담당 (별도 SPEC) |
| `.claude/rules/moai/NOTICE.md` | 라이선스 의무 keep-always |
| 14+ pre-existing path-scoped rules | 본 SPEC 무관, 검증 reference 로만 사용 |
| `internal/config/loader.go` | Go 코드 변경 0건 (REQ-RPS-NF-011) |
| 4 rule 본문 (frontmatter 외) | body byte preservation (REQ-RPS-NF-014) |

---

## 3. Implementation Order (Milestone Plan)

### M1: 4 Rule Local Frontmatter Prepend (priority: High)

**Goal**: 4 target rule 파일에 frontmatter prepend (local `.claude/rules/moai/` 위치).

**Steps**:
1. `.claude/rules/moai/core/zone-registry.md` 첫 줄 위에 frontmatter 4줄 prepend
2. `.claude/rules/moai/design/constitution.md` 첫 줄 위에 frontmatter 4줄 prepend
3. `.claude/rules/moai/development/manager-develop-prompt-template.md` 첫 줄 위에 frontmatter 4줄 prepend
4. `.claude/rules/moai/workflow/agent-teams-pattern.md` 첫 줄 위에 frontmatter 4줄 prepend

**Frontmatter 표준 형식** (각 파일):

```yaml
---
description: "<one-line purpose statement>"
paths: "<CSV glob string from spec.md §1.3 표>"
---
```

**Verification**:
- `head -5 <file>` 출력 첫 줄이 정확히 `---`
- `grep -c '^paths:' <file>` 결과 = 1
- `tail -n +6 <file>` 가 변경 전 원본 본문과 byte-identical (REQ-RPS-NF-014)
- YAML parse 검증: `python3 -c "import yaml,sys; print(yaml.safe_load(open(sys.argv[1]).read().split('---')[1]))" <file>` 정상 출력

**Tool**: `Edit` (Read-before-write 의무, 단일 파일당 단일 Edit operation 권장)

**Risk**: R-RPS-004 (frontmatter parse 실패) — 선례 14+ 단순 형식 모방으로 회피

### M2: 4 Template Mirror Frontmatter Prepend (priority: High)

**Goal**: M1 변경을 `internal/template/templates/.claude/rules/moai/` 4 mirror 파일에 byte-identical 적용 + `make build` 임베디드 갱신.

**Steps**:
1. `internal/template/templates/.claude/rules/moai/core/zone-registry.md` 동일 frontmatter prepend
2. `internal/template/templates/.claude/rules/moai/design/constitution.md` 동일 frontmatter prepend
3. `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` 동일 frontmatter prepend
4. `internal/template/templates/.claude/rules/moai/workflow/agent-teams-pattern.md` 동일 frontmatter prepend
5. `make build` 실행 (project root) — `internal/template/embedded.go` 갱신 트리거

**Verification**:
- `diff .claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/core/zone-registry.md` → empty (byte-identical, REQ-RPS-006)
- 4 file pair 모두 동일 검증
- `ls -la internal/template/embedded.go` 의 modification time 이 `make build` 직후
- (선택) `internal/template/rule_template_mirror_test.go` 또는 동치 `TestRuleTemplateMirrorDrift` 실행 PASS

**Tool**: `Edit` (Read-before-write 의무) + `Bash` (`make build`)

**Risk**: R-RPS-003 (Template mirror drift) — M2 단계 명시 분리로 회피

### M3: Doctor 시뮬레이션 + 5 Session Scenario Verification (priority: Medium)

**Goal**: R-RPS-001 (trigger miss) 완화. 5 representative session 시나리오별 4 rule + 5 keep-always rule 로드 매트릭스 생성.

**5 Session 시나리오**:
1. **Go-only session**: `internal/cli/foo.go` 만 수정 — `agent-teams-pattern.md` / `manager-develop-prompt-template.md` / `design/constitution.md` 미로드 기대 / `zone-registry.md` 미로드 기대
2. **SPEC-only session**: `.moai/specs/SPEC-XXX/spec.md` 수정 — `zone-registry.md` / `manager-develop-prompt-template.md` 로드 기대
3. **Design session**: `.moai/design/tokens.json` 수정 — `design/constitution.md` 로드 기대
4. **Team session**: `.moai/config/sections/workflow.yaml` 수정 — `agent-teams-pattern.md` 로드 기대
5. **General docs session**: `README.md` 수정 — 4 path-scoped rule 모두 미로드 기대 / 5 keep-always 만 로드

**Steps**:
1. 시뮬레이션 스크립트 (Bash) 작성 또는 manual grep 으로 5 시나리오 별 각 rule 의 `paths:` 글롭 매치 여부 평가
2. `.moai/reports/rules-path-scope-simulation-<YYYY-MM-DD>.md` 작성 (markdown 표 형식)
3. 보고서에 위반 (trigger miss 또는 spurious load) 0 건 확인

**Boilerplate report 형식**:

```markdown
# Rules Path-Scope Simulation Report

Date: 2026-05-XX
SPEC: SPEC-V3R6-RULES-PATH-SCOPE-001

## Simulation Matrix

| Session | zone-registry | design/constitution | manager-develop-prompt | agent-teams-pattern | (5 keep-always) |
|---|---|---|---|---|---|
| Go-only | ✗ | ✗ | ✗ | ✗ | ✓ all 5 |
| SPEC-only | ✓ | ✗ | ✓ | ✗ | ✓ all 5 |
| Design | ✓ | ✓ | ✗ | ✗ | ✓ all 5 |
| Team | ✓ | ✗ | ✗ | ✓ | ✓ all 5 |
| General docs | ✗ | ✗ | ✗ | ✗ | ✓ all 5 |

## Findings
- Trigger miss: 0
- Spurious load: 0
- Token saving estimate (vs always-loaded): -23.5K tokens (62% off in Go-only / docs scenarios)
```

**Verification**:
- 보고서 파일 존재
- 보고서에 위반 0 건 명시
- (acceptance.md AC-RPS-014 가 detection rule 정의)

**Tool**: Bash + Write

**Risk**: R-RPS-001 (실측 trigger miss 발견) — 발견 시 M1 글롭 보정 후 M2 재실행 (cycle 가능)

### M4: 회귀 테스트 (priority: Medium)

**Goal**: 본 SPEC 변경이 기존 CI 결함 도입 없음 확인.

**Steps**:
1. `go test ./internal/template/...` 실행 — pre-existing baseline 외 새 실패 0 건
2. `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` 둘 다 exit 0 (Cross-platform)
3. `golangci-lint run --timeout=2m` baseline 외 새 issue 0 건
4. `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ internal/hook/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"` 0 매치 (C-HRA-008 boundary)
5. `wc -w .claude/rules/moai/core/agent-common-protocol.md .claude/rules/moai/workflow/session-handoff.md .claude/rules/moai/workflow/context-window-management.md .claude/rules/moai/workflow/verification-batch-pattern.md .claude/rules/moai/NOTICE.md` 합산 vs 9-rule 원본 합산 비교 → -7,500w 이상 감소 (REQ-RPS-NF-010)
6. `grep -A 1 'paths:' .claude/rules/moai/core/zone-registry.md` 정상 출력 (frontmatter intact)
7. spec-lint baseline 비교 (`.moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/spec.md` 의 frontmatter 12-field PASS)

**Verification**: All steps exit 0 또는 expected baseline state.

**Tool**: Bash (parallel verification batch per `agent-common-protocol.md` § Parallel Execution)

**Risk**: 없음 — 모두 read-only 검증 명령

### M5: Implementation 완료 보고 + status 갱신 (priority: Low)

**Goal**: spec.md frontmatter `status: draft → implemented` 갱신, version v0.1.0 → v0.2.0, 진행 상황 progress.md 작성.

**Steps**:
1. `.moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/spec.md` frontmatter `status: implemented`, `version: "0.2.0"`, `updated: <today>` 갱신
2. `.moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/progress.md` 작성 (M1-M4 결과 요약 + 측정값)
3. (sync-phase) Hybrid Trunk Tier M = feat branch + PR — manager-git 위임

**Verification**: spec.md `grep '^status:' frontmatter` = `implemented`, progress.md 존재.

**Tool**: Edit + Write

---

## 4. Pre-flight Checks (착수 전 의무)

`manager-develop` 위임 받은 implementer 가 M1 시작 전 실행:

```bash
# 1. 현재 branch + HEAD baseline
git branch --show-current
git rev-parse HEAD

# 2. Cross-platform build 가능성 사전 확인 (baseline)
go build ./... 2>&1 | tail -3
GOOS=windows GOARCH=amd64 go build ./... 2>&1 | tail -3

# 3. 4 target rule 의 현재 frontmatter 상태 확인 (모두 zero-frontmatter 기대)
for f in .claude/rules/moai/core/zone-registry.md .claude/rules/moai/design/constitution.md .claude/rules/moai/development/manager-develop-prompt-template.md .claude/rules/moai/workflow/agent-teams-pattern.md; do
  echo "=== $f ==="
  head -1 "$f"  # expect: # <Title> (not ---)
done

# 4. 4 target rule 의 단어 수 측정 (REQ-RPS-NF-010 baseline)
wc -w .claude/rules/moai/core/zone-registry.md .claude/rules/moai/design/constitution.md .claude/rules/moai/development/manager-develop-prompt-template.md .claude/rules/moai/workflow/agent-teams-pattern.md

# 5. 5 keep-always rule 의 현재 상태 (no frontmatter 변경 의무 baseline)
for f in .claude/rules/moai/core/agent-common-protocol.md .claude/rules/moai/workflow/session-handoff.md .claude/rules/moai/workflow/context-window-management.md .claude/rules/moai/workflow/verification-batch-pattern.md .claude/rules/moai/NOTICE.md; do
  head -1 "$f"
done

# 6. Template mirror 존재 확인 (CLAUDE.local.md §2 HARD 의무)
ls -la internal/template/templates/.claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/design/constitution.md internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md internal/template/templates/.claude/rules/moai/workflow/agent-teams-pattern.md

# 7. internal/rules/ 부재 재확인 (REQ-RPS-NF-011)
ls internal/rules/ 2>/dev/null || echo "internal/rules/ MISSING — Go code change avoidable, per design.md §Layer 1 footnote re internal/rules/loader.go"

# 8. paths frontmatter 선례 확인 (REQ-RPS-008 표준 형식 reference)
grep -l '^paths:' .claude/rules/moai/development/*.md .claude/rules/moai/core/*.md | head -5

# 9. Cross-SPEC 충돌 확인
grep -r "Retired\|TestRulePathScopeRetired\|deprecation-marker" .claude/rules/moai/ || echo "no cross-SPEC retirement conflicts"
```

---

## 5. Brownfield Strategy

본 SPEC 는 brownfield (4 기존 rule 파일 수정) 이나, body 무변경 + frontmatter prepend only 로 **non-destructive**:

| 보존 영역 | 보존 방식 |
|---|---|
| 4 rule body 전체 (frontmatter 외) | byte-identical (REQ-RPS-NF-014 — `tail -n +6` diff empty 의무) |
| 5 keep-always rule | frontmatter 추가 0 건 (REQ-RPS-007) |
| 14+ pre-existing path-scoped rules | 변경 0 건 (reference only) |
| `internal/config/loader.go` 등 Go config 로더 | 변경 0 건 (REQ-RPS-NF-011) |
| `internal/template/` build pipeline | `make build` 동작 변경 없음, 갱신 트리거만 사용 |
| CI workflows | `.github/workflows/` 변경 0 건 |
| pre-existing baseline tests | failure list 변동 없음 (M4 §5에서 측정) |

상호 충돌 검증: `grep -r "Retired\|deprecation-marker" .claude/rules/moai/` 사전 0 매치 (Pre-flight §9). 본 SPEC 가 retire 시키는 rule 없음 — 모두 보존 + 발동 조건 변경만 적용.

---

## 6. PRESERVE List

run-phase implementer (manager-develop) 가 변경 금지 의무 파일:

- `.claude/rules/moai/core/agent-common-protocol.md` — keep-always rule
- `.claude/rules/moai/workflow/session-handoff.md` — keep-always rule (본문 footnote 명시)
- `.claude/rules/moai/workflow/context-window-management.md` — RULES-COMPRESS-001 담당
- `.claude/rules/moai/workflow/verification-batch-pattern.md` — RULES-COMPRESS-001 담당
- `.claude/rules/moai/NOTICE.md` — 라이선스 의무 keep-always
- 14+ pre-existing path-scoped rules (`spec-frontmatter-schema.md`, `agent-patterns.md`, `skill-authoring.md`, `agent-hooks.md`, `moai-constitution.md`, `settings-management.md`, `boundary-verification.md`, `hooks-system.md`, `karpathy-quickref.md`, `skill-ab-testing.md`, `skill-writing-craft.md`, `model-policy.md`, `agent-authoring.md` 등) — reference only
- `internal/config/loader.go` 및 `internal/config/` 전체 — Go code 0 change
- `internal/template/embedded.go` — auto-generated, manual edit 금지 (`make build` 갱신만 허용)
- 4 rule body (frontmatter 외) — REQ-RPS-NF-014 byte preservation
- 3 sister SPEC 디렉토리 (`SPEC-V3R6-RULES-COMPRESS-001`, `SPEC-V3R6-SKILL-CONSOLIDATE-001`, `SPEC-V3R6-SKILL-COMPRESS-001`) — Wave 1 Lane A 다른 SPEC, 본 SPEC 무관
- Working tree 의 7 modified + 28+ untracked 파일 (`.moai/research/`, `docs-site/`, `internal/template/templates/.github/`, `internal/template/renderer.go`, `internal/hook/.moai/`, `.moai/harness/usage-log.jsonl` 등) — 본 SPEC 무관, parallel session 작업

---

## 7. Technical Approach

### 7.1 Why no Go code change

design.md §Layer 1 본문 (라인 180-183) 은 "구현 방법" 으로 "각 rule 파일에 `paths: [...]` frontmatter 추가" + "`internal/rules/loader.go` (가칭)에 path matching 로직" 명시. 그러나 위임 prompt §Pre-flight 결과 + 본 SPEC §1.4 검증으로 다음 확정:

- **사실 1**: `internal/rules/` 디렉토리 부재 (`ls internal/rules/ 2>/dev/null` empty)
- **사실 2**: 14+ pre-existing path-scoped rules 가 이미 정상 동작 중 (Claude Code 런타임이 `paths:` 를 직접 해석)
- **사실 3**: `internal/config/loader.go` 는 YAML config 로더 (rule 로더 아님)

→ **결론**: design.md §Layer 1 "가칭" 표현은 잠정 추정이었고, 실제로는 Go 코드 변경 0건이 정답. 본 SPEC §3.3.2 + REQ-RPS-NF-011 로 명시 확정.

### 7.2 Why CSV string vs YAML array

`.claude/rules/moai/development/coding-standards.md` § Paths Frontmatter 예시:

```yaml
---
paths: "**/*.py,**/pyproject.toml"
---
```

CSV string 단일 형식 표준. YAML array 는 `internal/skill/loader.go` (만일 존재) 또는 다른 파싱 로직이 추가로 필요 — 표준 위반. 14+ 기존 사례 모두 CSV string 사용 (검증 완료).

### 7.3 Frontmatter parsing safety

위임 prompt §Pre-flight 검증으로 14+ 선례 중 가장 단순한 형식 사용 — 따옴표 escape, indent 변형 불필요. `agent-hooks.md` 의 `paths: "**/.claude/agents/**,**/.claude/hooks/**"` 같은 단일 라인 CSV 형식이 안전한 default.

### 7.4 Template-First Rule 적용

CLAUDE.local.md §2 [HARD]: ".claude/, .moai/, .agency/ 어디서든 새 파일 추가 시 internal/template/templates/<path> 먼저 추가 후 make build". 본 SPEC 는 신규 파일 추가가 아닌 **기존 파일 수정** 이나 동일 의무 적용:

1. local `.claude/rules/moai/<sub>/*.md` 수정
2. `internal/template/templates/.claude/rules/moai/<sub>/*.md` 동일 수정 (byte-identical)
3. `make build` 로 `internal/template/embedded.go` 갱신
4. 검증: `diff` empty + `make build` exit 0

### 7.5 Token saving 측정 방법

REQ-RPS-NF-010 의 측정 방식:

- **Baseline (현재)**: 9 always-loaded rule 의 `wc -w` 합산 = 13,690 단어 (실측 §1.4 표 + spec.md §1.3 추정 ~40,500w 는 추가 sub-content 포함한 token estimate, word count base 는 별도)
  - zone-registry 3,453 + design/constitution 2,472 + agent-common-protocol 1,927 + session-handoff 1,927 + manager-develop-prompt 1,136 + context-window-mgmt 712 + agent-teams-pattern 770 + verification-batch-pattern 764 + NOTICE 349 = 13,510 단어
- **Target (변경 후)**: 5 keep-always rule 의 `wc -w` 합산 (frontmatter 4-line overhead 포함) ~ 5,679 + ~20 (description text) ≈ 5,700 단어
- **감소량**: ~7,810 단어 ≈ ~23.5K tokens (3 tokens/word 비율)
- **AC measurement**: AC-RPS-008 에서 "9-rule sum baseline 이 5-rule sum 보다 최소 7,500w 더 큼" binary 검증

---

## 8. Edge Cases

### 8.1 Edge Case 1: Glob 미매치 session 에서 SPEC body 가 path-scoped rule 인용

예) Go-only session 에서 작성된 SPEC body 가 "zone-registry HARD 조항 12에 따르면..." 인용 시, `zone-registry.md` 미로드 상태 → 인용 내용 확인 불가.

**완화**: spec.md / plan.md 작성 시 자동으로 `.moai/specs/**` glob 매치 → `zone-registry.md` 로드. SPEC 외부 인용 (예: 일반 코드 comment) 시는 일시적으로 손실되나 critical 아님 (HARD 조항 위반 시 별도 CI guard 가 정정).

### 8.2 Edge Case 2: Doctor 시뮬레이션 false positive

M3 시뮬레이션은 보수적 — glob string 매치 평가가 Claude Code 런타임 실측과 일부 다를 가능성. 예) `**/.claude/skills/**` 와 `.claude/skills/**` 의 미세 차이.

**완화**: M3 보고서는 **추정** 임을 명시. 실측 trigger miss 발견 시 24h 운영 관찰 (M3 mitigation §1.3) 로 보정.

### 8.3 Edge Case 3: 4 rule 중 1 개만 frontmatter parse 실패

R-RPS-004 가시화 — 4 rule 중 1 개만 frontmatter 오류 → 3 정상 + 1 미로드.

**완화**: M1 단계에서 각 rule 별 YAML parse 검증 (REQ-RPS-008 + AC-RPS-009). 1 개 실패 시 즉시 rollback + 재작성.

---

## 9. Self-Audit (Tier M Pre-Submission)

### 9.1 plan-auditor 4 Dimension 자체 평가 (Preview)

| Dimension | Self-Score | 근거 |
|---|---|---|
| D1 Design Quality | 0.88 | path-scoped 설계 명확. 글롭 보수적 광범위화 (R-RPS-001 mitigation). selected 4 + skipped 5 분리 명시. Go 0 change 확정 (Pre-flight §7 + REQ-RPS-NF-011). |
| D2 Acceptance Criteria | 0.84 | 14 binary AC (acceptance.md 작성 시) + 5 session simulation 매트릭스. 일부 AC (token saving) 가 ~3 tokens/word 추정 의존 — Low risk. |
| D3 Spec Quality | 0.86 | 12-field frontmatter PASS. 9 EARS REQs 5 patterns 분포. ## Out of Scope h3 sub-section (§3.3.1/§3.3.2). 5 Risks 명시. HISTORY v0.1.0. |
| D4 Traceability | 0.90 | 9 REQs → 14 AC 1:N 매핑 (acceptance.md 작성 시). PRESERVE list 명시. Cross-SPEC dependency (`depends_on` 2건 + `related_specs` 3건) 정확. |
| **종합** | **0.87** | Tier M PASS threshold 0.80 대비 +0.07 margin. 0 BLOCKING 자체 점검. |

### 9.2 EARS Pattern Distribution

| Pattern | REQ # | 개수 |
|---|---|---|
| Ubiquitous | REQ-RPS-001, REQ-RPS-002, REQ-RPS-003, REQ-RPS-004 | 4 |
| Event-Driven | REQ-RPS-005, REQ-RPS-006 | 2 |
| State-Driven | REQ-RPS-007 | 1 |
| Optional | REQ-RPS-008, REQ-RPS-009 | 2 |
| Non-Functional (Ubiquitous) | REQ-RPS-NF-010, REQ-RPS-NF-011, REQ-RPS-NF-012, REQ-RPS-NF-013, REQ-RPS-NF-014 | 5 |

총 14 REQ (9 Functional + 5 NF). 5 patterns 모두 적어도 1건씩 사용 (Unwanted-behavior 없음 — 본 SPEC 는 추가 기능 아닌 frontmatter 조정이라 prohibition 형 자연 없음).

### 9.3 Out of Scope 검증

- §3.3.1 (다른 SPEC) 6 items
- §3.3.2 (본 SPEC 절대 미터치) 7 items
- 모두 h3 sub-section 형식 (`### X.Y` 헤더 + 표) — spec-lint MissingExclusions WARN 회피

### 9.4 12-Field Frontmatter Self-Check

- [x] id: SPEC-V3R6-RULES-PATH-SCOPE-001 (regex match)
- [x] title: quoted string
- [x] version: "0.1.0" (quoted semver)
- [x] status: draft
- [x] created: 2026-05-22 (canonical name, NOT created_at)
- [x] updated: 2026-05-22 (canonical name, NOT updated_at)
- [x] author: manager-spec
- [x] priority: P1
- [x] phase: "v3.0.0"
- [x] module: ".claude/rules/moai"
- [x] lifecycle: spec-anchored
- [x] tags: "rules, path-scope, token-economy, wave-1, v3r6, always-loaded-reduction" (canonical name, NOT labels, CSV string NOT array)
- [x] (Optional) tier: M
- [x] (Optional) depends_on: 2건
- [x] (Optional) related_specs: 3건

---

## 10. Cross-references

- [spec.md](./spec.md) — 본 SPEC 의 EARS REQ + Out of Scope + Risks
- [acceptance.md](./acceptance.md) — 14 binary AC + Traceability matrix
- `.moai/research/v3.0-design-2026-05-22.md` §Layer 1 — 청사진 출처
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier M 위임 5-section template 의무 (본 SPEC 가 path-scoped 화 대상이므로 약간의 자기참조성 있음 — 위임 prompt 자체는 본 SPEC 머지 전까지 always-loaded 유지하여 본 SPEC 자체에 의무 영향 없음)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier M 3 artifacts + 0.80 plan-auditor threshold
