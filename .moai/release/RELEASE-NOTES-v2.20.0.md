# MoAI-ADK v2.20.0 Release Notes — Self-Evolving Harness v2 Foundation

> Target release: v2.20.0 (v3.0.0-rc1 candidate)
> Authoritative SPEC: `SPEC-V3R4-HARNESS-001`
> Breaking change ID: `BC-V3R4-HARNESS-001-CLI-RETIREMENT`

---

## TL;DR

V3R4 self-evolving harness 아키텍처의 **foundation** 릴리스. 세 개의 V3R3 harness SPEC을 단일 V3R4 family로 통합하고, `moai harness <verb>` Go CLI verb path를 공식적으로 폐기합니다. harness 라이프사이클은 이제 `/moai:harness` 슬래시 커맨드 + skill workflow body + Claude Code hook만으로 동작합니다 — Go 바이너리는 더 이상 호출되지 않습니다. 5-Layer Safety 아키텍처와 FROZEN zone은 비트 단위로 보존됩니다.

---

## Self-Evolving Harness v2 Foundation

### Why this release exists

V3R3 시기에 만들어진 세 개의 harness SPEC (`SPEC-V3R3-HARNESS-001`, `SPEC-V3R3-HARNESS-LEARNING-001`, `SPEC-V3R3-PROJECT-HARNESS-001`)은 서로 다른 Phase letter (C / D) 하에 다른 release target (v2.17.0 / v2.19.0)으로 별도 작성되었습니다. 같은 아키텍처 도메인을 다루지만 V3R4 self-evolution vision이 정착되기 전에 쓰여, 표면이 V3R4 통합 아키텍처와 깔끔하게 매핑되지 않았습니다. 이 릴리스는 그 세 개를 단일 foundation SPEC 아래로 통합 (`supersedes:` frontmatter)하여 downstream SPEC 002-008 작성자에게 단일 contract를 제공합니다.

근거 자료: `/Users/goos/MoAI/moai-adk-go/.moai/brain/IDEA-004/proposal.md` (8-SPEC decomposition), `research.md` (24 cited external sources — Reflexion arXiv:2303.11366, Voyager arXiv:2305.16291, Constitutional AI arXiv:2212.08073, LangGraph reflection production patterns 2026, Anthropic Claude Code Skills/Agents 공식 문서 2026, revfactory/harness Apache-2.0).

### CLI retirement migration story (BC-V3R4-HARNESS-001-CLI-RETIREMENT)

V3R3 시절에 `moai harness <verb>` 셸 커맨드를 사용한 사용자가 있다면 다음과 같이 마이그레이션해야 합니다:

| 이전 (V3R3) | 신규 (V3R4) |
|-------------|-------------|
| `moai harness status` | `/moai:harness status` (Claude Code 세션 내) |
| `moai harness apply` | `/moai:harness apply` (Claude Code 세션 내) |
| `moai harness rollback <date>` | `/moai:harness rollback <YYYY-MM-DD>` (Claude Code 세션 내) |
| `moai harness disable` | `/moai:harness disable` (Claude Code 세션 내) |

`v2.20.0` 설치 후 `moai harness status`를 셸에서 호출하면 다음 진단이 출력됩니다:

```
Error: unknown command "harness" for "moai"
```

이것은 의도된 동작입니다. 슬래시 커맨드 표면은 V3R3 시절과 동일하므로 Claude Code 세션 내 머슬 메모리에는 영향이 없습니다. 슬래시 커맨드 thin wrapper (`.claude/commands/moai/harness.md`)는 `moai` 스킬 본문의 `workflows/harness.md` workflow module로 라우팅하며, workflow body는 모든 4개 verb (`status` / `apply` / `rollback` / `disable`)를 file-system 연산으로 직접 구현합니다 — 어떤 Go 바이너리도 호출되지 않습니다.

기존 `.moai/harness/usage-log.jsonl` 항목은 마이그레이션 / 삭제 / 수정되지 않습니다. 사용자는 누적된 관측 데이터를 그대로 가지고 V3R4로 진입합니다.

### Preserved without modification

- **5-Layer Safety 아키텍처** (`.claude/rules/moai/design/constitution.md` §5) — L1 Frozen Guard / L2 Canary Check / L3 Contradiction Detector / L4 Rate Limiter (≤3/week, ≥24h cooldown) / L5 Human Oversight (AskUserQuestion at Tier 4) 전체가 비트 단위 보존 (REQ-HRN-FND-005, AC-HRN-FND-004).
- **FROZEN zone** (constitution.md §2) — `.claude/agents/moai/`, `.claude/skills/moai-`, `.claude/rules/moai/`, `.moai/project/brand/` 4개 path prefix가 L1 Frozen Guard에 의해 보호됩니다. 위반 시도는 `.moai/harness/learning-history/frozen-guard-violations.jsonl`에 silent audit (REQ-HRN-FND-006, REQ-HRN-FND-014, AC-HRN-FND-005).
- **4-tier observation ladder** (Observation 1 / Heuristic 3 / Rule 5 / Auto-update 10) — `SPEC-V3R3-HARNESS-LEARNING-001` REQ-HL-002의 임계값이 그대로 유지됩니다 (REQ-HRN-FND-011, AC-HRN-FND-008).
- **PostToolUse observer 스키마** — ISO-8601 timestamp + event_type + subject + context_hash 4 필드 (REQ-HRN-FND-010).
- **AskUserQuestion orchestrator monopoly** — subagent는 user prompt 불가; 필요 시 structured blocker report로 orchestrator에 위임 (REQ-HRN-FND-015, `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary).

### New runtime behavior

- **`/moai:harness` 슬래시 커맨드**: 4개 verb 전부 workflow body에서 직접 구현. CLI 호출 0건.
- **Tier-4 AskUserQuestion 4-option pattern**: Apply (권장) / Modify / Defer / Reject. 첫 옵션은 항상 `(권장)` 또는 `(Recommended)` 접미사.
- **Tier-4 rate limit**: 프로젝트당 7일 rolling window 1회 (REQ-HRN-FND-012). 향후 adaptive expansion이 가능하지만 1회 floor는 절대 하향 불가 (REQ-HRN-FND-018).
- **PostToolUse observer no-op gate**: `learning.enabled: false` 설정 시 observer가 stdin도 읽지 않고 즉시 exit 0. 기존 `.moai/harness/usage-log.jsonl` 항목은 보존됩니다 (REQ-HRN-FND-009).
- **Conflict resolution contract**: Reflexion 자체-비판 (downstream SPEC-V3R4-HARNESS-004) 과 `evaluator-active` 채점이 충돌할 때 evaluator-active가 binding gate; Reflexion은 advisory pre-screen만 (REQ-HRN-FND-017).

### CI regression guard

신규 테스트 `internal/cli/harness_retirement_test.go`:

- `TestHarnessRetirement` — `rootCmd.Commands()` 중 `Use: "harness"` 항목이 발견되면 즉시 `t.Fatalf`로 차단. 진단 메시지는 SPEC-V3R4-HARNESS-001 + REQ-HRN-FND-002 + BC-V3R4-HARNESS-001-CLI-RETIREMENT를 명시.
- `TestHarnessFactoryStillCompiles` — `newHarnessCmd()` factory가 deprecation marker로 트리에 남아 있는지 검증. 향후 follow-up SPEC이 물리적 제거를 수행하기 전까지 factory는 호출 가능 상태로 유지됩니다.

신규 단위 테스트 `internal/cli/hook_harness_observe_test.go` (10 cases table-driven):

- `TestIsHarnessLearningEnabled` (7 cases) — gate function 자체 검증 (missing config / empty / no learning block / no enabled key / true / false / invalid YAML).
- `TestRunHarnessObserve_NoOpWhenLearningDisabled` — `learning.enabled: false`에서 `.moai/harness/usage-log.jsonl` 미생성 확인.
- `TestRunHarnessObserve_PreservesExistingLogWhenDisabled` — 기존 log entry 보존 확인.
- `TestRunHarnessObserve_RecordsWhenEnabled` — `learning.enabled: true`에서 JSONL 1건 append + 4-field schema 검증.

### Downstream SPECs enumeration (foundation 이후 계획)

이 foundation SPEC은 자체적으로 self-evolution 기능을 도입하지 않습니다. 다음 7개 SPEC이 점진적으로 V3R4 아키텍처를 확장합니다:

1. **SPEC-V3R4-HARNESS-002** — Multi-event observer: Stop / SubagentStop / UserPromptSubmit hook 통합 + 통합 관측 스키마.
2. **SPEC-V3R4-HARNESS-003** — Embedding-cluster pattern detection: 빈도 카운트 분류자를 임베딩 클러스터링으로 대체.
3. **SPEC-V3R4-HARNESS-004** — Actor + Evaluator + Self-Reflection trio: Reflexion 자체-비판 3-iteration cap + 자연어 reflection 에피소드 메모리.
4. **SPEC-V3R4-HARNESS-005** — Constitution principle 파서 + 자체-채점 rubric + pre-screen 통합.
5. **SPEC-V3R4-HARNESS-006** — Multi-objective scoring tuple (quality + token cost + latency + iteration count) + auto-rollback-on-regression.
6. **SPEC-V3R4-HARNESS-007** — Voyager 스킬 라이브러리: 임베딩 인덱스 + top-K 검색 + 합성 스킬 재사용.
7. **SPEC-V3R4-HARNESS-008** — 익명화 레이어 + opt-in cross-project federation + namespace isolation (개인정보 민감; 002-007 single-project 트랙 레코드 확인 후 plan-phase 진입).

각 downstream SPEC은 이 foundation SPEC의 REQ ID를 인용하여 binding constraint로 사용합니다.

### Out of scope (intentional non-goals)

- `internal/cli/harness.go` 또는 `internal/cli/harness_test.go`의 물리적 삭제 — 본 SPEC은 deprecation marker로 남기고 follow-up SPEC이 downstream 002-008 머지 이후 정리.
- 세 superseded V3R3 SPEC 파일의 frontmatter 수정 — `.moai/specs/SPEC-V3R4-HARNESS-001/follow-up.md` 지침에 따라 `manager-git`이 별도 commit으로 처리.
- `.claude/rules/moai/design/constitution.md` 수정 (FROZEN file).
- 기존 `usage-log.jsonl` 항목의 스키마 마이그레이션.
- 평가 history inspection GUI / 대시보드 / 웹 클라이언트.
- 네트워크 호출 / telemetry 업로드 / 외부 API 통합.

---

## Upgrade Notes

- **Existing users**: 기존 `.moai/harness/usage-log.jsonl` 상태는 그대로 유지됩니다. 첫 슬래시 커맨드 호출 (`/moai:harness status`)은 누적된 관측 데이터를 그대로 표시합니다.
- **`learning.enabled: false` 사용자**: observer가 no-op이 되었음을 확인하려면 `Edit` / `Write` 호출 후 `.moai/harness/usage-log.jsonl` 의 line count가 변하지 않는지 비교하세요.
- **CI 파이프라인**: 새 테스트 두 건 (`TestHarnessRetirement`, `TestHarnessFactoryStillCompiles`)이 추가됩니다 — 별도 설정 불필요.

---

## Detailed References

- SPEC: `.moai/specs/SPEC-V3R4-HARNESS-001/spec.md` (REQ-HRN-FND-001 ~ REQ-HRN-FND-018)
- Acceptance: `.moai/specs/SPEC-V3R4-HARNESS-001/acceptance.md` (12 AC + 6 edge cases + 7 scenarios)
- Plan: `.moai/specs/SPEC-V3R4-HARNESS-001/plan.md`
- Tasks: `.moai/specs/SPEC-V3R4-HARNESS-001/tasks.md` (5 Wave / 17 tasks)
- Follow-up: `.moai/specs/SPEC-V3R4-HARNESS-001/follow-up.md` (post-merge `manager-git` V3R3 status-transition 지침)
- Workflow body: `.claude/skills/moai/workflows/harness.md`
- Slash command thin wrapper: `.claude/commands/moai/harness.md`
- Hook code: `internal/cli/hook.go` (`isHarnessLearningEnabled`, `runHarnessObserve`)
- CI guard: `internal/cli/harness_retirement_test.go`
- Observer test: `internal/cli/hook_harness_observe_test.go`
- Brain artifacts: `.moai/brain/IDEA-004/proposal.md`, `ideation.md`, `research.md`
