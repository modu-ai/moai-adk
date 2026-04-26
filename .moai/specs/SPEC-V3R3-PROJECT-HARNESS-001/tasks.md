# SPEC-V3R3-PROJECT-HARNESS-001 — Tasks

Granular work items per phase. Each task references the REQ-IDs it implements. Dependencies on SPEC-V3R3-HARNESS-001 are marked explicitly with `[DEP:HARNESS-001]`.

Task ID format: `T-P{phase}-{NN}` where `{phase}` is the phase number (1-5) and `NN` is the in-phase sequence.

---

## Phase 1 — Socratic Interview (16Q / 4 Round)

### T-P1-01: Interview prompt template authoring (Round 1: Q1-Q4)
- **Type**: Design
- **REQ-IDs**: REQ-PH-001, REQ-PH-002
- **Dependencies**: None
- **Output**: `.claude/skills/moai/workflows/project.md`에 Phase 5 헤더 추가 + Round 1 명세. Q1 (도메인), Q2 (기술스택), Q3 (규모), Q4 (팀구성). 각 질문에 4 옵션 + 첫 옵션 "(권장)" + 상세 설명.
- **Done when**: 4 질문 모두 conversation_language (ko) 텍스트 + 권장 마커 + 옵션 차이점 명시 (예: "Mobile (iOS)" vs "Mobile (Android)" vs "Mobile (cross-platform)").

### T-P1-02: Round 2-4 prompt templates
- **Type**: Design
- **REQ-IDs**: REQ-PH-001, REQ-PH-002
- **Dependencies**: T-P1-01
- **Output**: project.md에 Round 2 (Q5 방법론, Q6 디자인툴, Q7 UI복잡도, Q8 디자인시스템), Round 3 (Q9 보안, Q10 성능, Q11 배포, Q12 외부통합), Round 4 (Q13 customization 범위, Q14 특수제약, Q15 우선순위, Q16 최종확인) 추가.
- **Done when**: 12 질문 모두 명세 + Q16의 옵션이 정확히 [Confirm/Restart/Abort] 3개.

### T-P1-03: In-memory answer buffer + abort handler
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-002, REQ-PH-010
- **Dependencies**: T-P1-02
- **Output**: `internal/harness/interview.go` 신설. `Answer` struct (question_id, round, text, recorded_at), `Buffer` struct (in-memory append-only), `Buffer.Abort()` (디스크 commit 0 보장), `Buffer.Commit()` (Phase 6 진입 시 호출).
- **Done when**: unit test가 abort 시 file system zero writes 검증; Commit() 후 buffer.Frozen() == true.

### T-P1-04: interview-results.md writer
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-003, REQ-PH-007
- **Dependencies**: T-P1-03
- **Output**: `internal/harness/interview_writer.go`. YAML frontmatter (spec_id, generated_at, project_root, conversation_language) + Round 1-4 헤더 + Q-A 쌍 + Recorded at timestamp.
- **Done when**: `yq eval '.spec_id'` 통과; `grep -c "^- Q[0-9][0-9]:"` == 16.

### T-P1-05: Phase 1 unit tests (interview)
- **Type**: Test
- **REQ-IDs**: REQ-PH-001, REQ-PH-002, REQ-PH-003, REQ-PH-010
- **Dependencies**: T-P1-03, T-P1-04
- **Output**: `internal/harness/interview_test.go`. Cases: 16Q full flow, abort at Round 2, Round 4 Restart 분기, conversation_language fallback.
- **Done when**: `go test -race ./internal/harness/...` PASS; coverage ≥ 85%.

---

## Phase 2 — meta-harness Invocation

### T-P2-01: Skill("moai-meta-harness") wrapper
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-004
- **Dependencies**: T-P1-04, [DEP:HARNESS-001]
- **Output**: `.claude/skills/moai/workflows/project.md`에 Phase 6 추가. orchestrator-side prompt template: `Skill("moai-meta-harness")` 호출 시 16개 답변을 structured context로 전달.
- **Done when**: project.md Phase 6 섹션이 Skill() 호출 형식 + answer-to-context schema 포함.

### T-P2-02: Path-prefix matcher (FROZEN guard)
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-011
- **Dependencies**: T-P2-01
- **Output**: `internal/harness/frozen_guard.go`. `IsAllowedPath(path string) bool` — `.claude/agents/my-harness/`, `.claude/skills/my-harness-*/`, `.moai/harness/` 만 true. moai-managed area는 false. 모든 write 호출 first check.
- **Done when**: unit test가 `.claude/agents/moai/foo.md` write 시도 → reject (error: FROZEN_VIOLATION).

### T-P2-03: Cleanup on partial output failure
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-004, REQ-PH-010
- **Dependencies**: T-P2-02
- **Output**: meta-harness 호출 중 mid-failure 시 cleanup 로직. `cleanup_on_failure: true`. 부분 산출물 (예: agent 1개만 생성됨) 자동 삭제.
- **Done when**: meta-harness mock failure 주입 후 `.claude/agents/my-harness/` empty 확인.

### T-P2-04: Phase 2 integration test
- **Type**: Test
- **REQ-IDs**: REQ-PH-004, REQ-PH-011
- **Dependencies**: T-P2-01, T-P2-02, T-P2-03, [DEP:HARNESS-001]
- **Output**: `internal/harness/meta_invocation_test.go`. iOS 답변 fixture → meta-harness 호출 → 4개 산출물 + moai-managed area 0 changes 검증.
- **Done when**: `go test -race -run TestMetaInvocationIOS` PASS; AC-PH-03 verification 통과.

---

## Phase 3 — 5-Layer Activation

### T-P3-01: L1 frontmatter triggers verifier
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-005, REQ-PH-008
- **Dependencies**: T-P2-04
- **Output**: `internal/harness/layer1.go`. `VerifyTriggers(skillPath string) error` — keywords/agents/phases/paths 4개 키 존재 확인. meta-harness 산출물에 inject 책임은 HARNESS-001이지만 본 SPEC은 verify 책임.
- **Done when**: my-harness-ios-patterns/SKILL.md에 4개 키 모두 존재 확인 PASS.

### T-P3-02: L2 workflow.yaml.harness 갱신
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-005, REQ-PH-008
- **Dependencies**: T-P3-01
- **Output**: `internal/harness/layer2.go`. `UpdateWorkflowYAML(domain, specID string, agents []AgentRef, skills []SkillRef, chains []ChainRule) error`. 기존 workflow.yaml의 team 섹션 보존 + harness 섹션 idempotent merge (`yaml.Node` 사용).
- **Done when**: `yq eval '.workflow.harness.enabled'` == true; 기존 `.workflow.team` 섹션 손상 없음.

### T-P3-03: L3 CLAUDE.md @import marker injector
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-005, REQ-PH-006
- **Dependencies**: T-P3-02
- **Output**: `internal/harness/layer3.go`. `InjectMarker(claudeMdPath, specID, domain string, paths []string) error`. `<!-- moai:harness-start id="..." -->` block insert (or replace if same id exists). Idempotent.
- **Done when**: re-run으로 marker 1개만 존재 (duplicate 없음); `grep -c "moai:harness-start"` == 1.

### T-P3-04: L5 .moai/harness/ scaffolder
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-005, REQ-PH-008, REQ-PH-012
- **Dependencies**: T-P3-03
- **Output**: `internal/harness/layer5.go`. 7개 파일 (main.md, plan-extension.md, run-extension.md, sync-extension.md, chaining-rules.yaml, interview-results.md, README.md). Q13 답변이 "Advanced"이면 design-extension.md 추가 (REQ-PH-012). 각 파일 첫 줄에 file purpose 명시.
- **Done when**: `ls .moai/harness/{main,plan-extension,run-extension,sync-extension,chaining-rules,interview-results,README}.{md,yaml}` 모두 존재; design-extension.md는 Q13 분기에 따라 조건부.

### T-P3-05: chaining-rules.yaml writer
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-005, REQ-PH-007 (chain 정의 영구 기록)
- **Dependencies**: T-P3-04
- **Output**: `internal/harness/chaining_rules.go`. machine-readable yaml (version, chains[].phase, chains[].when, chains[].insert_before, chains[].insert_after). manager-tdd가 read.
- **Done when**: yq parse OK; chain 1개 이상 (run phase 기본).

### T-P3-06: Phase 3 unit tests (5-Layer)
- **Type**: Test
- **REQ-IDs**: REQ-PH-005, REQ-PH-008, REQ-PH-012
- **Dependencies**: T-P3-01 ~ T-P3-05
- **Output**: `internal/harness/layer{1,2,3,5}_test.go`. 각 Layer 단위 + idempotency + Q13 분기 (REQ-PH-012).
- **Done when**: AC-PH-02 verification 통과.

---

## Phase 4 — Integration & Verification

### T-P4-01: New session simulation test
- **Type**: Test
- **REQ-IDs**: REQ-PH-006, REQ-PH-008
- **Dependencies**: T-P3-06
- **Output**: `internal/harness/session_replay_test.go`. CLAUDE.md re-load → @import follow → harness context 포함 검증. iOS fixture 사용.
- **Done when**: AC-PH-04 verification 통과 (manager-spec response가 ios-architect chain 언급).

### T-P4-02: moai update safety regression
- **Type**: Test
- **REQ-IDs**: REQ-PH-009
- **Dependencies**: T-P3-06
- **Output**: `internal/cli/update_safety_test.go`. pre-update snapshot → moai update → post-update diff (`diff -rq`). user area 0 changes 보장. 사용자 customization comment 보존 검증.
- **Done when**: AC-PH-05 verification 통과.

### T-P4-03: moai doctor 5-Layer diagnosis
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-009 (간접)
- **Dependencies**: T-P3-06
- **Output**: `internal/cli/doctor.go` 확장. 5-Layer 각각 PASS/FAIL 출력 + my-harness-* prefix 충돌 경고. exit code: WARN=0, FAIL=1.
- **Done when**: AC-PH-06 verification 통과 (`moai doctor` 출력에 5 lines + WARN line).

### T-P4-04: my-harness-* prefix conflict detector
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-009 (간접)
- **Dependencies**: T-P4-03
- **Output**: `internal/harness/prefix_conflict.go`. `my-harness-foundation-core`처럼 `moai-foundation-core` 의미 충돌 감지 (Levenshtein distance ≤ 2 + suffix match).
- **Done when**: 충돌 시 doctor가 "WARN: my-harness-X conflicts with moai-X" 출력.

### T-P4-05: end-to-end iOS scenario validation
- **Type**: Test
- **REQ-IDs**: REQ-PH-004, REQ-PH-006, REQ-PH-007, REQ-PH-008
- **Dependencies**: T-P4-01, T-P4-02
- **Output**: handoff §5.2 verbatim 시나리오 자동화. `internal/harness/e2e_ios_test.go`. moai init → /moai project → 인터뷰 → meta-harness → 새 세션 → /moai plan → /moai run → progress.md에 chain 기록 확인.
- **Done when**: AC-PH-08 verification 통과.

---

## Phase 5 — Template Mirror & Release

### T-P5-01: Template static import line — plan.md
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-005 (Layer 4)
- **Dependencies**: T-P4-05
- **Output**: `internal/template/templates/.claude/skills/moai/workflows/plan.md` 마지막에 한 줄 정적 import:
  ```
  ## Custom Harness Extension (Optional)

  @.moai/harness/plan-extension.md

  *(이 파일은 `/moai project --harness`로 생성됩니다. 파일이 없으면 자동으로 skip됩니다.)*
  ```
- **Done when**: template + local copy 동일 내용; `make build` 후 `internal/template/embedded.go` regenerated.

### T-P5-02: Template static import — run.md
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-005
- **Dependencies**: T-P5-01
- **Output**: 동일 패턴. `@.moai/harness/run-extension.md` import line 추가.
- **Done when**: 동일 검증.

### T-P5-03: Template static import — sync.md
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-005
- **Dependencies**: T-P5-02
- **Output**: 동일 패턴. `@.moai/harness/sync-extension.md`.
- **Done when**: 동일 검증.

### T-P5-04: Template static import — design.md
- **Type**: Implementation
- **REQ-IDs**: REQ-PH-005
- **Dependencies**: T-P5-03
- **Output**: 동일 패턴. `@.moai/harness/design-extension.md`. 파일 없으면 graceful skip 동작 강조 (Claude Code @import 자동 스킵).
- **Done when**: 4개 workflow file 모두 import line 보유 (`grep -l "@.moai/harness/" workflows/{plan,run,sync,design}.md | wc -l` == 4).

### T-P5-05: make build + commands_audit_test
- **Type**: Test
- **REQ-IDs**: REQ-PH-005
- **Dependencies**: T-P5-04
- **Output**: `make build` 실행 + `go test ./internal/template/...` 통과. `commands_audit_test.go`가 새 import line을 검증하도록 갱신.
- **Done when**: 모든 테스트 PASS, embedded.go에 4개 import line 포함.

### T-P5-06: HISTORY 갱신 + version bumps
- **Type**: Documentation
- **REQ-IDs**: 모든 REQ
- **Dependencies**: T-P5-05
- **Output**: spec.md HISTORY 추가 (`0.2.0` post-implementation), v2.17 release notes draft (`.moai/release/v2.17.0-draft.md`)에 본 SPEC entry 추가, CHANGELOG.md `## [Unreleased]` 섹션에 추가.
- **Done when**: 3개 파일 모두 갱신 완료.

### T-P5-07: Final commit
- **Type**: Git
- **REQ-IDs**: 모든 REQ
- **Dependencies**: T-P5-06
- **Output**: 한국어 commit body + conventional message: `spec(project): SPEC-V3R3-PROJECT-HARNESS-001 — 16Q 인터뷰 + 5-Layer 통합`. branch: `feat/SPEC-V3R3-HARNESS-LEARNING-001` (현재 브랜치, commit 분리).
- **Done when**: `git log` 첫 commit이 본 SPEC 메시지; pre-commit hook PASS (lint, vet, test).

---

## 6. Cross-Phase Dependencies Summary

```
[Phase 1: Interview]
   T-P1-01 → T-P1-02 → T-P1-03 → T-P1-04 → T-P1-05

[Phase 2: meta-harness]
   T-P2-01 → T-P2-02 → T-P2-03 → T-P2-04
   (T-P2-01 depends on T-P1-04 + [DEP:HARNESS-001])

[Phase 3: 5-Layer]
   T-P3-01 → T-P3-02 → T-P3-03 → T-P3-04 → T-P3-05 → T-P3-06
   (T-P3-01 depends on T-P2-04)

[Phase 4: Integration]
   T-P4-01, T-P4-02, T-P4-03 (parallel)
   T-P4-04 (depends on T-P4-03)
   T-P4-05 (depends on T-P4-01, T-P4-02)

[Phase 5: Template Mirror]
   T-P5-01 → T-P5-02 → T-P5-03 → T-P5-04 → T-P5-05 → T-P5-06 → T-P5-07
   (T-P5-01 depends on T-P4-05)
```

전체 critical path: **T-P1-01 → T-P1-04 → T-P2-01 → T-P2-04 → T-P3-01 → T-P3-06 → T-P4-05 → T-P5-04 → T-P5-07** (총 9단계).

---

## 7. Done Criteria

본 tasks.md는 다음이 모두 완료되면 `tasks_complete: true`:

- [ ] Phase 1: 5 tasks 완료, AC-PH-01, AC-PH-07 PASS
- [ ] Phase 2: 4 tasks 완료, AC-PH-03 PASS
- [ ] Phase 3: 6 tasks 완료, AC-PH-02 PASS
- [ ] Phase 4: 5 tasks 완료, AC-PH-04, AC-PH-05, AC-PH-06, AC-PH-08 PASS
- [ ] Phase 5: 7 tasks 완료, Template-First mirror + release prep 완료
- [ ] 총 27개 tasks 완료
- [ ] AC-PH-01 ~ AC-PH-08 모두 PASS (8/8)
- [ ] REQ-PH-001 ~ REQ-PH-012 모두 traceable (12/12)
- [ ] `go test -race ./internal/harness/... ./internal/cli/...` PASS
- [ ] `make build && make test && make lint` PASS
- [ ] 한국어 conventional commit 작성 + push 완료
