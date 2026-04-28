# SPEC-V3R3-PROJECT-HARNESS-001 — Acceptance Criteria

본 문서는 Given-When-Then 형식으로 8개 핵심 AC와 traceability 매트릭스를 정의한다. 모든 AC는 `spec.md` REQ-PH-001 ~ REQ-PH-012와 100% 매핑된다.

---

## AC-PH-01: 16Q Interview Simulation (iOS Project Scenario)

**Covers**: REQ-PH-001, REQ-PH-002, REQ-PH-003

### Given
- 빈 프로젝트 디렉터리에서 `moai init my-ios-app && cd my-ios-app` 후 `/moai project` 실행됨.
- 기존 Phase 1-4 (project doc generation) 완료된 상태.
- `conversation_language: ko` 설정됨.

### When
- 시스템이 Phase 5 진입 → 4 라운드 인터뷰 수행:
  - Round 1 (`AskUserQuestion` call #1, Q1-Q4):
    - Q1 도메인: "Mobile (iOS)" 선택 (권장: Web — 사용자가 Mobile 선택).
    - Q2 기술 스택: "Swift + SwiftUI" 선택.
    - Q3 규모: "MVP (1-3 modules)" 선택.
    - Q4 팀 구성: "Solo developer" 선택.
  - Round 2 (`AskUserQuestion` call #2, Q5-Q8):
    - Q5 방법론: "TDD" 선택.
    - Q6 디자인툴: "Figma" 선택.
    - Q7 UI 복잡도: "Standard (lists + forms)" 선택.
    - Q8 디자인시스템: "Custom DTCG tokens" 선택.
  - Round 3 (`AskUserQuestion` call #3, Q9-Q12):
    - Q9 보안: "OAuth + Keychain" 선택.
    - Q10 성능: "60fps 일반 UI" 선택.
    - Q11 배포: "App Store" 선택.
    - Q12 외부통합: "HealthKit" 선택.
  - Round 4 (`AskUserQuestion` call #4, Q13-Q16):
    - Q13 customization: "Standard (recommended)" 선택.
    - Q14 특수 제약: "iOS 17+ minimum" 선택.
    - Q15 우선순위: "thorough harness level" 선택.
    - Q16 최종 확인: "Confirm" 선택.

### Then
- 4번의 `AskUserQuestion` 호출만 발생 (총 16 질문, max 4 per call 준수).
- 각 호출의 첫 옵션은 "(권장)" 마커 포함 + 상세 설명 존재.
- 모든 16개 답변이 in-memory 버퍼에 누적됨.
- Round 4 Q16 응답 "Confirm" 후 → Phase 6 진입.
- `.moai/harness/interview-results.md` 작성됨:
  - YAML frontmatter: `spec_id`, `generated_at` (ISO8601), `project_root`, `conversation_language: ko`.
  - 본문에 16개 Q-A 쌍 (Round 1-4 헤더 포함) + 각 답변에 답변 시점 timestamp.
  - 파일 size > 0, jq-readable YAML frontmatter.

### Verification

```bash
test -f .moai/harness/interview-results.md
yq eval '.spec_id' .moai/harness/interview-results.md   # SPEC-PROJ-INIT-NNN
grep -c "^- Q[0-9]" .moai/harness/interview-results.md  # 16
```

---

## AC-PH-02: 5-Layer Independent Unit Tests

**Covers**: REQ-PH-005, REQ-PH-008, REQ-PH-012

### Given
- 인터뷰 완료 + meta-harness 호출 완료된 상태 (Phase 5+ Phase 6).
- `.claude/agents/my-harness/`, `.claude/skills/my-harness-*/`, `.moai/harness/` 산출물 생성됨.

### When
- Phase 7에서 5-Layer를 순차 활성화:
  - **L1** activation: 생성된 `my-harness-ios-patterns/SKILL.md`의 frontmatter에 `triggers` 섹션이 inject되었는지 단위 검증.
  - **L2** activation: `.moai/config/sections/workflow.yaml`에 `harness:` 섹션이 추가되었는지 단위 검증.
  - **L3** activation: `CLAUDE.md` 끝에 `<!-- moai:harness-start id="..." -->` ~ `<!-- moai:harness-end -->` 블록이 inject되었는지 단위 검증.
  - **L4** verify: template static import line (`@.moai/harness/plan-extension.md` 등 4개 워크플로우)이 존재하는지 verify.
  - **L5** activation: `.moai/harness/` 디렉터리에 7개 필수 파일이 생성되었는지 verify (main.md, plan-extension.md, run-extension.md, sync-extension.md, chaining-rules.yaml, interview-results.md, README.md).

### Then
- L1: `paths`, `keywords`, `agents`, `phases` 4개 키 모두 frontmatter에 존재.
  - paths 예: `**/*.swift,**/Package.swift,**/*.xcodeproj`.
  - agents 예: `["manager-tdd","manager-ddd","manager-spec","expert-frontend"]`.
- L2: workflow.yaml의 `harness.enabled: true`, `harness.domain: "ios-mobile"`, `harness.custom_agents` 배열 ≥ 1개, `harness.chaining_rules` 배열 ≥ 1개.
- L3: marker block 존재. id 필드는 SPEC-PROJ-INIT-NNN 형식. @import 5줄 (workflow.yaml, main.md, agent들, skill들). idempotent — 동일 id로 재실행 시 block in-place 갱신만 발생.
- L4: 4개 workflow file 모두에 import line 존재 (plan.md / run.md / sync.md / design.md).
- L5: 7개 필수 파일 모두 존재 + non-empty + each file 첫 줄에 file purpose 명시.

### Verification

```bash
# L1
yq eval '.triggers.paths' .claude/skills/my-harness-ios-patterns/SKILL.md
# L2
yq eval '.workflow.harness.enabled' .moai/config/sections/workflow.yaml  # true
# L3
grep -c "moai:harness-start" CLAUDE.md  # 1
grep -c "moai:harness-end" CLAUDE.md    # 1
# L4
grep -l "@.moai/harness/" .claude/skills/moai/workflows/{plan,run,sync,design}.md | wc -l  # 4
# L5
ls .moai/harness/{main.md,plan-extension.md,run-extension.md,sync-extension.md,chaining-rules.yaml,interview-results.md,README.md} | wc -l  # 7
```

### Sub-AC: REQ-PH-012 (Optional design-extension.md)

- Q13 답변이 "Advanced (full custom)"일 때만 `.moai/harness/design-extension.md` 추가 생성.
- Q13 답변이 "Standard"이면 design-extension.md는 생성되지 않음.

---

## AC-PH-03: meta-harness Invocation Output Verification

**Covers**: REQ-PH-004, REQ-PH-011

### Given
- 인터뷰 완료 (16 답변 in-memory).
- SPEC-V3R3-HARNESS-001 머지 완료 → `moai-meta-harness` skill 사용 가능.
- `moai-managed area` 현재 상태 snapshot 저장됨 (Git stash or temp diff).

### When
- Phase 6에서 `Skill("moai-meta-harness")` 호출됨 (16 답변을 prompt context로 전달).
- meta-harness가 산출물 생성:
  - `.claude/agents/my-harness/ios-architect.md` 생성.
  - `.claude/agents/my-harness/swiftui-engineer.md` 생성.
  - `.claude/skills/my-harness-ios-patterns/SKILL.md` 생성.
  - `.claude/skills/my-harness-swiftui-best-practices/SKILL.md` 생성.

### Then
- 4개 산출물 모두 존재 + non-empty.
- 각 agent.md는 `name`, `description`, `model`, `tools` 필수 frontmatter 보유.
- 각 skill의 SKILL.md는 `name`, `description`, `metadata.generated_by: moai-meta-harness`, `metadata.parent_spec` 보유.
- **moai-managed area 0 changes**: `git diff .claude/agents/moai/ .claude/skills/moai-*/` 출력 비어있음.
- meta-harness가 `.claude/agents/moai/`, `.claude/skills/moai-*/` 또는 `.claude/rules/moai/`에 쓰려고 시도하면 path-prefix matcher가 reject (FROZEN guard 동작 검증).

### Verification

```bash
ls .claude/agents/my-harness/*.md | wc -l  # ≥ 2
ls -d .claude/skills/my-harness-*/ | wc -l  # ≥ 2
git diff --name-only .claude/agents/moai/ .claude/skills/moai-*/ .claude/rules/moai/  # empty
```

---

## AC-PH-04: New Session Auto-Activation

**Covers**: REQ-PH-006

### Given
- AC-PH-01 ~ AC-PH-03 완료 → 5-Layer 활성 상태.
- 새 Claude Code 세션 시작됨 (`claude` 또는 `claude -w`).

### When
- 세션 시작 시 Claude Code가 `CLAUDE.md`를 자동 로드.
- `<!-- moai:harness-start -->` ~ `<!-- moai:harness-end -->` 블록 내부의 `@.moai/harness/main.md` 등 @import 디렉티브를 follow.
- 사용자가 첫 prompt로 `/moai plan "user authentication"` 입력.

### Then
- harness customization context (main.md + workflow.yaml.harness 내용)이 conversation context에 포함됨.
- manager-spec subagent 호출 시 prompt에 harness context (custom_agents 목록, chaining_rules)가 포함됨.
- manager-spec이 paths 매칭되는 영역 (예: iOS 프로젝트 → `.swift` 파일들) 분석 시 `my-harness-ios-patterns` skill을 자동 reference (L1 frontmatter triggers 동작).

### Verification

- 수동: 새 세션에서 `/moai plan "FaceID authentication"` 실행 후 manager-spec 응답에 "iOS-specific" 또는 "ios-architect chain" 언급 존재.
- 자동: integration test가 CLAUDE.md content를 parse → @import path follow → resolved content가 `.moai/harness/main.md` 내용 포함 확인.

```bash
# integration test pseudo-flow
moai test session-replay --fixture iosproject --prompt "/moai plan ..."
# expects manager-spec response to reference my-harness/ios-architect.md
```

---

## AC-PH-05: moai update Safety — User Area Preservation

**Covers**: REQ-PH-009

### Given
- AC-PH-01 ~ AC-PH-03 완료 + AC-PH-04 검증 통과한 상태.
- `.moai/harness/`, `.claude/agents/my-harness/`, `.claude/skills/my-harness-*/` 디렉터리 존재 + 사용자가 수동으로 `.moai/harness/run-extension.md`에 1줄 comment 추가 ("# user-customized chain").

### When
- 사용자가 `moai update` 실행 (구버전 → 신버전 마이그레이션 시뮬레이션).

### Then
- `moai update` 종료 시:
  - `.claude/skills/moai/`, `.claude/agents/moai/`, `.claude/rules/moai/` 등 moai-managed area는 갱신됨.
  - `.moai/harness/`, `.claude/agents/my-harness/`, `.claude/skills/my-harness-*/` 디렉터리는 어떤 파일도 추가/수정/삭제되지 않음.
  - 사용자가 추가한 "# user-customized chain" comment 보존됨.

### Verification

```bash
# Pre-update snapshot
cp -r .moai/harness /tmp/harness-pre
cp -r .claude/agents/my-harness /tmp/myh-agents-pre
cp -r .claude/skills/my-harness-* /tmp/myh-skills-pre

moai update

# Post-update diff
diff -rq /tmp/harness-pre .moai/harness                          # empty
diff -rq /tmp/myh-agents-pre .claude/agents/my-harness            # empty
for d in /tmp/myh-skills-pre/*; do
  diff -rq "$d" ".claude/skills/$(basename $d)"
done                                                              # all empty
grep "# user-customized chain" .moai/harness/run-extension.md    # found
```

---

## AC-PH-06: moai doctor Diagnosis & Prefix Conflict Warning

**Covers**: REQ-PH-009 (간접), 5-Layer health check

### Given
- AC-PH-03 완료 상태.
- 사용자가 의도적으로 `.claude/skills/my-harness-foundation-core/`를 생성 (충돌 시뮬레이션 — 정적 `moai-foundation-core`와 의미 충돌).

### When
- `moai doctor` 실행됨.

### Then
- doctor 출력에 5-Layer 진단 섹션 추가:
  - L1 status: PASS/FAIL (각 my-harness skill의 triggers 섹션 유효성).
  - L2 status: PASS/FAIL (workflow.yaml.harness 유효성).
  - L3 status: PASS/FAIL (CLAUDE.md marker block paired).
  - L4 status: PASS/FAIL (template + local import line 존재).
  - L5 status: PASS/FAIL (필수 7파일 존재).
- prefix 충돌 경고: "WARN: my-harness-foundation-core conflicts with moai-foundation-core (semantic name overlap). Consider renaming."
- doctor exit code: WARN의 경우 0 (non-blocking), FAIL의 경우 1.

### Verification

```bash
moai doctor 2>&1 | grep -E "L[1-5] (status|PASS|FAIL)" | wc -l   # ≥ 5
moai doctor 2>&1 | grep "WARN:.*my-harness-foundation-core"     # found
```

---

## AC-PH-07: Interview Results Permanent Record

**Covers**: REQ-PH-003, REQ-PH-007

### Given
- AC-PH-01 인터뷰 완료된 상태.

### When
- 사용자가 `cat .moai/harness/interview-results.md`로 파일 확인.
- 추후 `/moai project --harness` 재실행 시 interview-results.md를 reference로 사용 (이전 답변 표시 + 변경 옵션).

### Then
- 파일 구조:
  ```
  ---
  spec_id: SPEC-PROJ-INIT-NNN
  generated_at: <ISO8601>
  project_root: <abs path>
  conversation_language: ko
  ---

  # Interview Results

  ## Round 1: Domain & Technology Foundation
  - Q01: <question text>
    - Answer: <user choice>
    - Recorded at: <ISO8601>
  ... (Q02 ~ Q16)

  ## Round 4 Final Decision
  - Q16 Answer: Confirm
  - Generated harness: ios-architect, swiftui-engineer, my-harness-ios-patterns, my-harness-swiftui-best-practices
  ```
- Q1-Q16 모두 기록 + 각 답변에 timestamp.
- conversation_language의 텍스트 보존 (한국어 답변 → 한국어 그대로 기록).

### Verification

```bash
yq eval '.spec_id' .moai/harness/interview-results.md   # SPEC-PROJ-INIT-...
grep -c "^- Q[0-9][0-9]:" .moai/harness/interview-results.md  # 16
grep -c "Recorded at:" .moai/harness/interview-results.md  # 16
```

---

## AC-PH-08: 5-Layer All-Active End-to-End (handoff §5.2 verbatim)

**Covers**: REQ-PH-004, REQ-PH-007, REQ-PH-008

### Given
- AC-PH-01 ~ AC-PH-05 모두 통과.
- iOS 프로젝트 (`my-ios-app`) 활성 상태, 새 세션 시작.

### When
- 사용자가 `/moai plan "user authentication with FaceID"` 실행.
- manager-spec이 SPEC-AUTH-001 작성 → user 승인 → 사용자가 `/moai run SPEC-AUTH-001` 실행.

### Then
- `/moai plan` Phase:
  - manager-spec subagent prompt에 harness context (workflow.yaml.harness.chaining_rules) 포함.
  - manager-spec이 plan 작성 시 ios-architect agent를 chain (chaining_rules.yaml의 `phase: plan`, `insert_after: [my-harness/ios-architect]` 적용).
  - my-harness-ios-patterns skill이 paths 매칭으로 자동 활성 (`**/*.swift,**/Package.swift`).
  - 작성된 SPEC-AUTH-001/spec.md에 iOS-specific 패턴 (Keychain, SwiftUI lifecycle, FaceID API) 반영.

- `/moai run` Phase:
  - manager-tdd가 Phase 2 시작 시 `workflow.yaml.harness.chaining_rules` read.
  - chain 순서: `ios-architect` → `expert-frontend` (또는 `expert-backend`) → `swiftui-engineer` 적용.
  - 각 chain agent가 호출되었음을 progress.md에 기록.
  - 구현 결과물에 my-harness-ios-patterns의 패턴 (예: SwiftUI ObservableObject, async/await Keychain wrapper) 반영.

### Verification

```bash
# /moai plan output verification
cat .moai/specs/SPEC-AUTH-001/spec.md | grep -E "Keychain|FaceID|SwiftUI"  # ≥ 3 hits

# /moai run progress verification
cat .moai/specs/SPEC-AUTH-001/progress.md | grep -E "ios-architect|swiftui-engineer"  # ≥ 2 hits

# Generated code reflects ios patterns
grep -r "ObservableObject\|@StateObject" src/ | wc -l  # ≥ 1
```

---

## 5. AC ↔ REQ Traceability Matrix

100% coverage of REQ-PH-001 ~ REQ-PH-012:

| AC ID    | REQ-PH-001 | REQ-PH-002 | REQ-PH-003 | REQ-PH-004 | REQ-PH-005 | REQ-PH-006 | REQ-PH-007 | REQ-PH-008 | REQ-PH-009 | REQ-PH-010 | REQ-PH-011 | REQ-PH-012 |
|----------|:---------:|:---------:|:---------:|:---------:|:---------:|:---------:|:---------:|:---------:|:---------:|:---------:|:---------:|:---------:|
| AC-PH-01 |     ✓     |     ✓     |     ✓     |           |           |           |           |           |           |     ✓     |           |           |
| AC-PH-02 |           |           |           |           |     ✓     |           |           |     ✓     |           |           |           |     ✓     |
| AC-PH-03 |           |           |           |     ✓     |           |           |           |           |           |           |     ✓     |           |
| AC-PH-04 |           |           |           |           |           |     ✓     |           |     ✓     |           |           |           |           |
| AC-PH-05 |           |           |           |           |           |           |           |           |     ✓     |           |     ✓     |           |
| AC-PH-06 |           |           |           |           |     ✓     |           |           |           |     ✓     |           |           |           |
| AC-PH-07 |           |           |     ✓     |           |           |           |     ✓     |           |           |           |           |           |
| AC-PH-08 |           |           |           |     ✓     |           |     ✓     |     ✓     |     ✓     |           |           |           |           |

**Coverage check**:
- REQ-PH-001 → AC-PH-01 ✓
- REQ-PH-002 → AC-PH-01 ✓
- REQ-PH-003 → AC-PH-01, AC-PH-07 ✓
- REQ-PH-004 → AC-PH-03, AC-PH-08 ✓
- REQ-PH-005 → AC-PH-02, AC-PH-06 ✓
- REQ-PH-006 → AC-PH-04, AC-PH-08 ✓
- REQ-PH-007 → AC-PH-07, AC-PH-08 ✓
- REQ-PH-008 → AC-PH-02, AC-PH-04, AC-PH-08 ✓
- REQ-PH-009 → AC-PH-05, AC-PH-06 ✓
- REQ-PH-010 → AC-PH-01 ✓
- REQ-PH-011 → AC-PH-03, AC-PH-05 ✓
- REQ-PH-012 → AC-PH-02 (sub-AC) ✓

**Traceability**: 12/12 REQ covered = **100%**.

---

## 6. Definition of Done

본 SPEC은 다음 모든 조건이 충족될 때 `status: completed`:

- [ ] 8개 AC 모두 통과 (AC-PH-01 ~ AC-PH-08).
- [ ] 12개 REQ 모두 traceability 매트릭스에서 ✓ 표시.
- [ ] `internal/harness/interview_test.go` 등 5개 단위 테스트 파일 통과 (`go test -race ./...`).
- [ ] handoff §5.2 iOS 시나리오 manual end-to-end 검증 완료.
- [ ] SPEC-V3R3-HARNESS-001 (depends_on) 완료 verified.
- [ ] Template-First 미러 완료 (`internal/template/templates/.../workflows/{plan,run,sync,design}.md` 4개 파일 + `make build` 후 embedded.go regenerated).
- [ ] 한국어 commit body로 conventional commit 작성: `spec(project): SPEC-V3R3-PROJECT-HARNESS-001 — 16Q 인터뷰 + 5-Layer 통합`.
- [ ] CHANGELOG.md v2.17 entry 추가.
- [ ] `moai update` 회귀 테스트 통과 (`internal/cli/update_safety_test.go`).
- [ ] `moai doctor` 5-Layer 진단 섹션 출력 검증.
