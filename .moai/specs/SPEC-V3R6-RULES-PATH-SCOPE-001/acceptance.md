---
spec_id: SPEC-V3R6-RULES-PATH-SCOPE-001
artifact: acceptance
created: 2026-05-22
updated: 2026-05-22
---

# Acceptance Criteria — SPEC-V3R6-RULES-PATH-SCOPE-001

> Binary PASS/FAIL acceptance criteria for Tier M plan-auditor verification.
> Each AC includes: rationale, verification command, expected outcome.
> 100% REQ↔AC traceability matrix in §3.

---

## 1. Acceptance Criteria

### AC-RPS-001 — zone-registry frontmatter 추가 (REQ-RPS-001, REQ-RPS-005)

**Rationale**: `.claude/rules/moai/core/zone-registry.md` 가 path-scoped 로 전환되었음을 검증.

**Verification command**:
```bash
head -4 .claude/rules/moai/core/zone-registry.md
```

**Expected outcome** (exact):
- Line 1: `---`
- Line 2: starts with `description: ` (one-line, non-empty)
- Line 3: `paths: ".claude/**,.moai/specs/**,.claude/rules/**"` (exact CSV string match)
- Line 4: `---`

**Pass/Fail rule**: 4 라인 모두 위 조건 만족 → PASS. 1 라인이라도 불일치 → FAIL.

---

### AC-RPS-002 — design/constitution frontmatter 추가 (REQ-RPS-002, REQ-RPS-005)

**Rationale**: `.claude/rules/moai/design/constitution.md` 가 path-scoped 로 전환되었음을 검증.

**Verification command**:
```bash
head -4 .claude/rules/moai/design/constitution.md
```

**Expected outcome**:
- Line 1: `---`
- Line 2: starts with `description: `
- Line 3: `paths: ".moai/design/**,.moai/specs/SPEC-*-DESIGN-*/**,.moai/project/brand/**,.claude/skills/moai/**/design*.md,.claude/skills/moai/**/brand*.md"` (exact CSV)
- Line 4: `---`

**Pass/Fail rule**: 4 라인 모두 위 조건 만족 → PASS.

---

### AC-RPS-003 — manager-develop-prompt-template frontmatter 추가 (REQ-RPS-003, REQ-RPS-005)

**Rationale**: `.claude/rules/moai/development/manager-develop-prompt-template.md` 가 path-scoped 로 전환되었음을 검증.

**Verification command**:
```bash
head -4 .claude/rules/moai/development/manager-develop-prompt-template.md
```

**Expected outcome**:
- Line 1: `---`
- Line 2: starts with `description: `
- Line 3: `paths: ".moai/specs/**,.claude/agents/moai/manager-develop.md,.claude/skills/moai/workflows/run.md"` (exact CSV)
- Line 4: `---`

**Pass/Fail rule**: 4 라인 모두 위 조건 만족 → PASS.

---

### AC-RPS-004 — agent-teams-pattern frontmatter 추가 (REQ-RPS-004, REQ-RPS-005)

**Rationale**: `.claude/rules/moai/workflow/agent-teams-pattern.md` 가 path-scoped 로 전환되었음을 검증.

**Verification command**:
```bash
head -4 .claude/rules/moai/workflow/agent-teams-pattern.md
```

**Expected outcome**:
- Line 1: `---`
- Line 2: starts with `description: `
- Line 3: `paths: ".moai/config/sections/workflow.yaml,.claude/agents/moai/manager-strategy.md,.claude/skills/moai/team/**"` (exact CSV)
- Line 4: `---`

**Pass/Fail rule**: 4 라인 모두 위 조건 만족 → PASS.

---

### AC-RPS-005 — 5 Keep-Always Rule frontmatter 비추가 (REQ-RPS-007)

**Rationale**: 5 out-of-scope rule 이 frontmatter 추가되지 않았음 (keep-always 유지) 검증.

**Verification command**:
```bash
for f in .claude/rules/moai/core/agent-common-protocol.md .claude/rules/moai/workflow/session-handoff.md .claude/rules/moai/workflow/context-window-management.md .claude/rules/moai/workflow/verification-batch-pattern.md .claude/rules/moai/NOTICE.md; do
  head -1 "$f"
done
```

**Expected outcome**: 5 라인 모두 첫 줄이 `# ` 또는 `>` 또는 일반 텍스트로 시작 (즉, `---` 가 아님). YAML frontmatter 부재.

**Pass/Fail rule**: 5 파일 모두 첫 줄이 `---` 가 아님 → PASS. 1 개라도 `---` 시작 → FAIL.

---

### AC-RPS-006 — 4 Rule Body Byte Preservation (REQ-RPS-NF-014)

**Rationale**: frontmatter prepend (4 lines `---` + desc + paths + `---` + 1 blank separator line = **5 lines total**) 외 body 1 byte 변경 없음 검증. plan.md §1.1 diff sample 의 5-line convention 과 일치 (agent-hooks.md L4 blank line 선례 모방).

**Verification command** (4 rule 각각):
```bash
# 예시 (zone-registry)
git show HEAD:.claude/rules/moai/core/zone-registry.md > /tmp/zr-orig.md
diff <(tail -n +6 .claude/rules/moai/core/zone-registry.md) /tmp/zr-orig.md
```

**Frontmatter convention**: line 1 `---` / line 2 `description: ...` / line 3 `paths: ...` / line 4 `---` / line 5 blank / line 6+ body. `tail -n +6` 이 body 첫 줄 (원 baseline 의 line 1) 부터 출력 → diff 가 empty 여야 정확.

**Expected outcome**: 4 rule 각각 `diff` 결과 empty (0 line difference). 본 SPEC 머지 전 baseline (HEAD) body 와 비교 시 frontmatter 5 라인 (4 frontmatter + 1 blank separator) prepend 외 본문 byte-identical.

**Pass/Fail rule**: 4 rule 모두 `diff` empty → PASS. 1 rule 이라도 body 차이 발견 → FAIL.

---

### AC-RPS-007 — 4 Template Mirror Sync (modulo §1.4.1 baseline drift) (REQ-RPS-006)

**Rationale**: local `.claude/rules/moai/` 4 rule 과 `internal/template/templates/.claude/rules/moai/` 4 mirror 의 sync 검증 (CLAUDE.local.md §2 [HARD] Template-First Rule). **단, spec.md §1.4.1 pre-existing drift baseline modulo** — 2 file 은 strict byte-identical, 2 file 은 본 SPEC frontmatter prepend 만 동기화 (baseline drift 보존, 별도 SPEC 흡수).

**Verification command (Group A — strict byte-identical, 2 files)**:
```bash
# pre-existing byte-identical pair (drift 없음)
for sub in design/constitution.md workflow/agent-teams-pattern.md; do
  echo "=== $sub (strict) ==="
  diff ".claude/rules/moai/$sub" "internal/template/templates/.claude/rules/moai/$sub"
done
```

**Verification command (Group B — drift modulo, 2 files)**:
```bash
# pre-existing drift pair: 본 SPEC frontmatter prepend 만 동기화. baseline drift 보존.
for sub in core/zone-registry.md development/manager-develop-prompt-template.md; do
  echo "=== $sub (drift modulo) ==="
  # post-SPEC frontmatter 5 lines 만 byte-identical 검증
  diff <(head -5 ".claude/rules/moai/$sub") <(head -5 "internal/template/templates/.claude/rules/moai/$sub")
  # 본 SPEC 으로 인해 body drift 가 확대되지 않았는지 검증
  pre_drift=$(git show HEAD:.claude/rules/moai/$sub 2>/dev/null | diff - <(git show HEAD:internal/template/templates/.claude/rules/moai/$sub 2>/dev/null) | wc -l)
  post_drift=$(diff <(tail -n +6 ".claude/rules/moai/$sub") <(tail -n +6 "internal/template/templates/.claude/rules/moai/$sub") | wc -l)
  echo "pre-SPEC drift=$pre_drift lines, post-SPEC body drift=$post_drift lines"
  test $post_drift -le $pre_drift && echo "PASS (drift not expanded)" || echo "FAIL (drift expanded)"
done
```

**Expected outcome**: Group A 2 pair `diff` empty. Group B 2 pair frontmatter 5 lines diff empty AND body drift ≤ pre-SPEC baseline ("PASS (drift not expanded)" 출력).

**Pass/Fail rule**: Group A 2 pair 모두 empty AND Group B 2 pair "PASS (drift not expanded)" → PASS. Group B drift 가 본 SPEC 으로 인해 확대 시 → FAIL.

**Note**: §1.4.1 documented baseline drift 2건 (zone-registry 4 lines + manager-develop-prompt-template 20 lines) 의 정정은 별도 `SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001` (V3R6 lesson #25 가칭) 범위. 본 SPEC 는 frontmatter prepend 만 적용하고 drift 자체는 보존.

---

### AC-RPS-008 — Always-Loaded Word Count 7,500+ 감소 (REQ-RPS-NF-010)

**Rationale**: 본 SPEC 머지 후 always-loaded rule 부담이 baseline 대비 최소 7,500 단어 감소.

**Verification command**:
```bash
# Baseline (9-rule always-loaded sum, pre-SPEC)
git show HEAD~1:.claude/rules/moai/core/agent-common-protocol.md HEAD~1:.claude/rules/moai/workflow/session-handoff.md HEAD~1:.claude/rules/moai/workflow/context-window-management.md HEAD~1:.claude/rules/moai/workflow/verification-batch-pattern.md HEAD~1:.claude/rules/moai/NOTICE.md HEAD~1:.claude/rules/moai/core/zone-registry.md HEAD~1:.claude/rules/moai/design/constitution.md HEAD~1:.claude/rules/moai/development/manager-develop-prompt-template.md HEAD~1:.claude/rules/moai/workflow/agent-teams-pattern.md 2>/dev/null | wc -w
# Expected baseline: ~13,510 words

# Target (5-rule keep-always sum, post-SPEC)
wc -w .claude/rules/moai/core/agent-common-protocol.md .claude/rules/moai/workflow/session-handoff.md .claude/rules/moai/workflow/context-window-management.md .claude/rules/moai/workflow/verification-batch-pattern.md .claude/rules/moai/NOTICE.md | tail -1
# Expected target: ~5,700 words

# Saving
echo "Saved words: $((13510 - 5700)) ≈ 7810"
```

**Expected outcome**: Baseline word count 약 13,500 + ε. Target 약 5,700 + ε. 차이 ≥ 7,500.

**Pass/Fail rule**: 감소량 ≥ 7,500 → PASS. < 7,500 → FAIL.

**Note**: 정확한 word count 는 ± few words 변동 (description text 추가분). 7,500 threshold 는 보수적 margin (실측 ~7,810 보다 약간 낮게 설정).

---

### AC-RPS-009 — 4 Rule Frontmatter YAML Parse Validity (REQ-RPS-008, R-RPS-004)

**Rationale**: 4 rule 의 frontmatter 블록이 유효한 YAML 임 검증 (parsing 실패 시 rule 전체 미로드 방지).

**Verification command**:
```bash
for f in .claude/rules/moai/core/zone-registry.md .claude/rules/moai/design/constitution.md .claude/rules/moai/development/manager-develop-prompt-template.md .claude/rules/moai/workflow/agent-teams-pattern.md; do
  echo "=== $f ==="
  python3 -c "
import sys, yaml
with open('$f') as fh:
    content = fh.read()
parts = content.split('---', 2)
if len(parts) < 3:
    print('FAIL: no frontmatter')
    sys.exit(1)
fm = yaml.safe_load(parts[1])
assert 'description' in fm, 'missing description'
assert 'paths' in fm, 'missing paths'
assert isinstance(fm['paths'], str), 'paths must be CSV string'
print('PASS:', fm)
"
done
```

**Expected outcome**: 4 file 모두 `PASS: {...}` 출력. `FAIL` 또는 `AssertionError` 없음.

**Pass/Fail rule**: 4 rule 모두 PASS 출력 → PASS.

---

### AC-RPS-010 — Go-only Session 시나리오 시뮬레이션 (R-RPS-001 mitigation, REQ-RPS-NF-012)

**Rationale**: 일반 Go 코드 session 에서 4 path-scoped rule 모두 미로드 + 5 keep-always 만 로드 검증.

**Verification command**:
```bash
# .moai/reports/rules-path-scope-simulation-<DATE>.md 에서 Go-only row 추출
grep -A 1 "^| Go-only " .moai/reports/rules-path-scope-simulation-*.md
```

**Expected outcome**: 매트릭스 row 가 `✗ | ✗ | ✗ | ✗ | ✓ all 5` 형식. 4 path-scoped 모두 ✗ + 5 keep-always 모두 ✓.

**Pass/Fail rule**: 위 형식 일치 → PASS. 1 셀 다름 → FAIL.

---

### AC-RPS-011 — SPEC-only Session 시나리오 시뮬레이션 (R-RPS-001 mitigation)

**Rationale**: `.moai/specs/SPEC-XXX/spec.md` 수정 session 에서 `zone-registry.md` + `manager-develop-prompt-template.md` 로드, `design/constitution` 미로드, `agent-teams-pattern` 미로드 검증.

**Verification command**:
```bash
grep -A 1 "^| SPEC-only " .moai/reports/rules-path-scope-simulation-*.md
```

**Expected outcome**: `✓ | ✗ | ✓ | ✗ | ✓ all 5` 형식.

**Pass/Fail rule**: 일치 → PASS.

---

### AC-RPS-012 — Design Session 시나리오 시뮬레이션 (R-RPS-001 mitigation)

**Rationale**: `.moai/design/tokens.json` 또는 brand 디렉토리 수정 session 에서 `design/constitution.md` 로드 검증.

**Verification command**:
```bash
grep -A 1 "^| Design " .moai/reports/rules-path-scope-simulation-*.md
```

**Expected outcome**: design column = ✓.

**Pass/Fail rule**: design column ✓ → PASS.

---

### AC-RPS-013 — Team Session 시나리오 시뮬레이션 (R-RPS-001 mitigation)

**Rationale**: `.moai/config/sections/workflow.yaml` 또는 team skill 수정 session 에서 `agent-teams-pattern.md` 로드 검증.

**Verification command**:
```bash
grep -A 1 "^| Team " .moai/reports/rules-path-scope-simulation-*.md
```

**Expected outcome**: agent-teams-pattern column = ✓.

**Pass/Fail rule**: agent-teams-pattern column ✓ → PASS.

---

### AC-RPS-014 — Trigger Miss / Spurious Load 0 건 (R-RPS-001 mitigation)

**Rationale**: 5 session 시나리오 전체에서 trigger miss (필요한데 미로드) 또는 spurious load (불필요한데 로드) 0 건 검증.

**Verification command**:
```bash
grep -E "Trigger miss: 0|Spurious load: 0" .moai/reports/rules-path-scope-simulation-*.md | wc -l
```

**Expected outcome**: 결과 = `2` (두 줄 모두 발견).

**Pass/Fail rule**: = 2 → PASS. < 2 → FAIL (즉시 M1 글롭 보정 후 M2 재실행).

---

### AC-RPS-015 — No Go Code Change (REQ-RPS-NF-011)

**Rationale**: `internal/rules/`, `internal/loader/` 디렉토리 신규 작성 없음 + `internal/config/loader.go` 미수정 검증.

**Verification command**:
```bash
git diff --stat HEAD~1 HEAD -- 'internal/**/*.go' | grep -v '_test.go' | grep -v '^$'
ls internal/rules/ internal/loader/ 2>/dev/null
```

**Expected outcome**:
- 첫 명령: empty 출력 (Go non-test file 변경 0 건) — 단, `internal/template/embedded.go` 는 `make build` auto-generated 이므로 제외
- 둘째 명령: `internal/rules/` 와 `internal/loader/` 모두 `No such file or directory` 또는 empty

**Pass/Fail rule**: 두 검증 모두 expected → PASS.

**Note**: `internal/template/embedded.go` 는 `make build` 시 자동 갱신 — 이는 REQ-RPS-006 의 의도된 결과로 본 AC 의 위반 대상 아님. AC measurement 시 `grep -v 'embedded.go'` 적용.

---

### AC-RPS-016 — Cross-Platform Build (REQ-RPS-NF-013, baseline regression)

**Rationale**: 본 SPEC 변경이 cross-platform Go build 결함 도입 없음.

**Verification command**:
```bash
go build ./... 2>&1 | tail -3
GOOS=windows GOARCH=amd64 go build ./... 2>&1 | tail -3
```

**Expected outcome**: 두 명령 모두 exit 0 + tail -3 출력에 build error 없음.

**Pass/Fail rule**: 두 build 모두 exit 0 → PASS.

---

### AC-RPS-017 — Pre-existing CI Baseline Preserved (REQ-RPS-NF-013)

**Rationale**: 본 SPEC 가 새 CI 결함 도입 없음 (baseline 외 NEW 0 건).

**Verification command**:
```bash
go test ./internal/template/... 2>&1 | tail -10
golangci-lint run --timeout=2m 2>&1 | tail -10
```

**Expected outcome**:
- `go test` 결과: pre-existing baseline failures (`TestRuleTemplateMirrorDrift`, `TestLateBranchTemplateMirror`, `TestSkillsContainPlanAuditGateMarkers` 등 `project_v3r6_template_mirror_drift_audit_2026_05_22` memory 기록 베이스라인) 외 새 FAIL 0 건. 단, 본 SPEC 가 `TestRuleTemplateMirrorDrift` 의 4 mirror sync 정정으로 일부 baseline failure **개선** 가능.
- `golangci-lint`: pre-existing baseline 외 NEW issue 0 건.

**Pass/Fail rule**: 새 결함 0 건 → PASS. NEW 결함 있을 시 → FAIL + 정정 후 재검증.

---

### AC-RPS-018 — C-HRA-008 Subagent Boundary (baseline)

**Rationale**: 본 SPEC 가 `internal/harness/` 또는 `internal/hook/` 파일을 건드리지 않으므로 C-HRA-008 boundary 무관 — 0 매치 유지 검증.

**Verification command**:
```bash
grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ internal/hook/ 2>/dev/null | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//" | wc -l
```

**Expected outcome**: 결과 = `0`.

**Pass/Fail rule**: 0 매치 → PASS. (본 SPEC 무관 baseline, 회귀 검증용)

---

## 2. Edge Cases (검증 시나리오)

### EC-RPS-001: Glob 미매치 session 에서 SPEC body 가 zone-registry HARD 조항 인용

**Scenario**: README.md 수정 session 에서 사용자가 "zone-registry HARD 조항 12 위반 우려" 발화.

**Expected**: orchestrator 가 README 작업 시 `zone-registry.md` 자동 로드 안 됨 (glob `.claude/**,.moai/specs/**,.claude/rules/**` 모두 미매치). 사용자 발화 시 orchestrator 가 명시적으로 인용 정보 부족함을 알림, 또는 (선택) `.claude/rules/**` glob 매치를 위해 zone-registry 를 명시적으로 Read tool 호출.

**완화**: 본 edge case 는 critical 아님 — orchestrator 가 명시적으로 zone-registry Read 호출하면 즉시 해소. 시뮬레이션 보고서에서 "general docs" row 가 이 시나리오 cover.

### EC-RPS-002: 시뮬레이션 보고서 자체 작성 시 trigger

**Scenario**: M3 단계에서 `.moai/reports/rules-path-scope-simulation-<DATE>.md` 작성 — 이는 `.moai/reports/**` glob 에 매치되는 paths 가 4 rule 중 어디에도 없음 (보고서 영역은 path-scoped 화 대상 외).

**Expected**: 보고서 작성 session 에서 4 path-scoped rule 모두 미로드. 5 keep-always 만 로드. 보고서 작성 자체는 단순 Write 작업이라 rule 의무 없음 — 정상.

### EC-RPS-003: Template mirror sync 누락 발견

**Scenario**: M1 만 적용 + M2 skip 시 `TestRuleTemplateMirrorDrift` 가 BLOCKING ERROR.

**Expected**: M2 의무 강제 (REQ-RPS-006). AC-RPS-007 `diff` empty 의무 검증. 발견 시 즉시 M2 재실행.

---

## 3. REQ ↔ AC Traceability Matrix

100% coverage: 14 REQ 모두 최소 1 개 AC 에 매핑. 14 AC 모두 최소 1 개 REQ 에 매핑.

| REQ | AC | Notes |
|---|---|---|
| REQ-RPS-001 | AC-RPS-001 | zone-registry path-scoped |
| REQ-RPS-002 | AC-RPS-002 | design/constitution path-scoped |
| REQ-RPS-003 | AC-RPS-003 | manager-develop-prompt path-scoped |
| REQ-RPS-004 | AC-RPS-004 | agent-teams-pattern path-scoped |
| REQ-RPS-005 | AC-RPS-001, AC-RPS-002, AC-RPS-003, AC-RPS-004 | frontmatter prepend (4 rule 모두 cover) |
| REQ-RPS-006 | AC-RPS-007 | template mirror sync (4 pair byte-identical) |
| REQ-RPS-007 | AC-RPS-005 | 5 keep-always frontmatter 비추가 |
| REQ-RPS-008 | AC-RPS-009 | YAML parse + CSV string 표준 |
| REQ-RPS-009 | AC-RPS-001, AC-RPS-002, AC-RPS-003, AC-RPS-004 | description: 표준화 (각 AC 의 "starts with `description:`" 검증) |
| REQ-RPS-NF-010 | AC-RPS-008 | Word count -7,500+ |
| REQ-RPS-NF-011 | AC-RPS-015 | No Go code change |
| REQ-RPS-NF-012 | AC-RPS-010, AC-RPS-011, AC-RPS-012, AC-RPS-013, AC-RPS-014 | Doctor simulation 5 시나리오 + trigger miss 0 |
| REQ-RPS-NF-013 | AC-RPS-016, AC-RPS-017 | Cross-platform build + CI baseline |
| REQ-RPS-NF-014 | AC-RPS-006 | Body byte preservation (4 rule diff empty) |

**Coverage**: 14/14 REQ → AC 매핑 (100%). 14/14 AC → REQ 매핑 (100%).

**Risk 매핑** (informational, AC 가 mitigation 검증):

| Risk | Mitigation AC |
|---|---|
| R-RPS-001 (Trigger miss) | AC-RPS-010 ~ AC-RPS-014 (5 시나리오 + 위반 0) |
| R-RPS-002 (Claude Code 런타임 회귀) | AC-RPS-009 (YAML parse) + 14+ 선례 운영 reference |
| R-RPS-003 (Template mirror drift) | AC-RPS-007 (4 pair diff empty) |
| R-RPS-004 (Frontmatter parse 실패) | AC-RPS-009 (Python yaml.safe_load) |
| R-RPS-005 (Word count 측정 오차) | AC-RPS-008 (보수적 7,500 threshold) |

---

## 4. Definition of Done

본 SPEC 가 `status: implemented` 로 전환되는 조건:

- [ ] M1 완료: 4 rule local frontmatter prepend (AC-RPS-001, AC-RPS-002, AC-RPS-003, AC-RPS-004 모두 PASS)
- [ ] M2 완료: 4 template mirror byte-identical sync + `make build` (AC-RPS-007 PASS)
- [ ] M3 완료: 5 session 시나리오 시뮬레이션 보고서 작성 + 위반 0 (AC-RPS-010 ~ AC-RPS-014 모두 PASS)
- [ ] M4 완료: 회귀 테스트 (AC-RPS-006, AC-RPS-008, AC-RPS-009, AC-RPS-015, AC-RPS-016, AC-RPS-017, AC-RPS-018 모두 PASS)
- [ ] M5 완료: spec.md `status: implemented`, version `0.2.0` + progress.md 작성
- [ ] sync-phase: Hybrid Trunk Tier M 의무 = feat branch + PR (Wave 1 Lane A 다른 SPEC 와 batch sync 가능)
- [ ] PR 머지 시 origin/main 정렬

총 18 AC 가 PASS (100%) 또는 N/A (해당 없음, 본 SPEC 에는 N/A 없음).

---

## 5. Quality Gates

### 5.1 plan-auditor Tier M Threshold

- **PASS threshold**: 0.80
- **Self-estimated score**: 0.87 (plan.md §9.1)
- **BLOCKING criteria**: D1 / D2 / D3 / D4 모두 ≥ 0.70 (Tier M 정의)
- **STOP signal**: iter(N+1) < iter(N) → 스코프 축소 제안 의무

### 5.2 TRUST 5 (CLAUDE.md §6)

- **Tested**: AC-RPS-010 ~ AC-RPS-014 (5 시나리오 doctor 시뮬레이션) + AC-RPS-006, AC-RPS-009, AC-RPS-017 (회귀 검증)
- **Readable**: 4 rule frontmatter 가 14+ 선례 표준 형식 일치 + `description:` 한국어 1줄 명확
- **Unified**: 4 rule 동일 frontmatter 형식 (description + paths CSV)
- **Secured**: 본 SPEC 는 security 무관 (rule 로딩 메커니즘만 변경)
- **Trackable**: Conventional Commits + 🗿 MoAI trailer + `SPEC-V3R6-RULES-PATH-SCOPE-001` reference

### 5.3 SPEC-LINT

- 12-field canonical frontmatter (spec.md §자체) PASS
- ## Out of Scope (h3 sub-section §3.3.1/§3.3.2) 명시 → MissingExclusions WARN 회피
- HISTORY (v0.1.0) 명시

---

## 6. Cross-references

- [spec.md](./spec.md) — 9 EARS REQs + 5 Risks + Out of Scope
- [plan.md](./plan.md) — M1-M5 milestone + Pre-flight + PRESERVE list + Self-audit
- `.moai/research/v3.0-design-2026-05-22.md` §Layer 1 (라인 165-184) — 청사진
- `.claude/rules/moai/development/coding-standards.md` § Paths Frontmatter — CSV string 표준
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — SPEC frontmatter SSOT
- CLAUDE.local.md §2 [HARD] Template-First Rule
