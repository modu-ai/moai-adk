# Wave 3 Execution Strategy (Phase 1 Output)

> Audit trail. manager-strategy output for Wave 3 of SPEC-V3R3-CI-AUTONOMY-001 — T3 Auto-Fix Loop on CI Fail.
> Generated: 2026-05-07. Methodology: TDD (RED-GREEN-REFACTOR). Mode: Standard.
> Base: `origin/main 5d3f6a4c1` (Wave 2 merged via PR #788; Wave 1 PR #785 → `79f313c4a` already in main).

---

## 1. Goal

Wave 3는 5-PR sweep (2026-05-05) 사례에서 노출된 R3 (auto-fix 부재) 를 종결한다. Wave 2가 만든 CI watch loop가 required check 실패를 감지하면 (exit 2 + JSON handoff), Wave 3 skill이 그 metadata를 받아 `expert-debug` subagent로 진단을 위임하고, mechanical 실패는 patch 제안 → 오케스트레이터 `AskUserQuestion` 컨펌 → apply → push → re-watch 루프를 최대 3회 반복한다. semantic 실패(test assertion / data race / deadlock / panic)는 patch 시도 없이 즉시 사용자 에스컬레이션한다.

본 Wave는 사용자가 PR을 열어둔 채 다른 작업을 하더라도 mechanical CI 실패가 자동 해결되도록 만드는 자율성 핵심 단계다. 5-PR sweep 시나리오 중 P1 (i18n test literal), P2 (errcheck/unused), P3 (Windows-only 빌드 실패)가 mechanical 분류로 들어와 1-2 iteration 안에 처리되고, P5 (ETXTBSY race)는 즉시 escalate되는 경로가 본 Wave의 직접 검증 대상이다. AC-CIAUT-006 (mechanical resolve), AC-CIAUT-007 (semantic escalate), AC-CIAUT-008 (iteration cap 3 + AskUserQuestion 강제) 세 acceptance가 Wave 3 외에서는 충족 불가능하다.

## 2. Scope (In / Out)

### In Scope (Wave 3 deliverables — 5 files)

1. `internal/template/templates/.claude/skills/moai-workflow-ci-autofix/SKILL.md` (new) — orchestrator-side state machine + AskUserQuestion cadence wiring
2. `internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md` (new) — HARD invocation rules (max iter, force-push 금지, semantic escalation)
3. `internal/template/templates/.claude/agents/expert-debug.md` (extend) — CI failure log + PR diff interpretation 섹션 추가 (기존 본문 보존)
4. `scripts/ci-autofix/classify.sh` (new) — POSIX sh classifier: mechanical vs semantic + trivial sub-classifier
5. `scripts/ci-autofix/log-fetch.sh` (new) — `gh run view --log-failed` wrapper + PR diff capture

User-facing copies (`.claude/skills/...`, `.claude/rules/...`, `.claude/agents/...`) are produced by `make build` re-rendering `internal/template/embedded.go` — Phase 2 의무. Wave 3 strategy doc는 template 경로만 deliverable로 카운트한다 (Template-First, CLAUDE.local.md §2).

### Out of Scope

- Wave 4 (T4 auxiliary workflow cleanup)
- Wave 5 (T6 worktree state guard)
- Wave 6 (T7 i18n validator)
- Wave 7 (T8 BODP)
- semantic 실패에 대한 자동 patch (영구 prohibited per spec.md §2 Exclusions)
- release/tag automation (영구 prohibited per `feedback_release_no_autoexec.md`)
- multi-PR 동시 auto-fix (single-PR-at-a-time, Wave 2 watch loop 모델 계승)
- semantic patch ML/heuristic 학습 (사용자 결정에 항상 위임)
- T3 트리거 외 경로에서의 expert-debug 강화 (기존 expert-debug 동작은 보존, additive only)

## 3. Files

| # | Template-source path (Phase 2 write target) | User-facing path (post-`make build`) | New / Extend |
|---|---------------------------------------------|--------------------------------------|--------------|
| 1 | `internal/template/templates/.claude/skills/moai-workflow-ci-autofix/SKILL.md` | `.claude/skills/moai-workflow-ci-autofix/SKILL.md` | New |
| 2 | `internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md` | `.claude/rules/moai/workflow/ci-autofix-protocol.md` | New |
| 3 | `internal/template/templates/.claude/agents/expert-debug.md` | `.claude/agents/expert-debug.md` | Extend (additive section "CI Failure Interpretation") |
| 4 | `scripts/ci-autofix/classify.sh` | `scripts/ci-autofix/classify.sh` | New (POSIX sh, repo-rooted, no template mirror — same file) |
| 5 | `scripts/ci-autofix/log-fetch.sh` | `scripts/ci-autofix/log-fetch.sh` | New (POSIX sh, same as #4) |

Tests are companions, not separate deliverables (consistent with Wave 1/2 convention):
- `scripts/ci-autofix/test/classify_test.sh` (POSIX harness, mock log fixtures)
- `scripts/ci-autofix/test/log_fetch_test.sh` (mock `gh` injection via `MOAI_AUTOFIX_GH=`)

State file (added by orchestrator at runtime, declared but not written by Phase 2 code other than path constant):
- `.moai/state/ci-autofix-<PR>.json` — iteration counter + last-classification + last-iter-trivial (per defect D5)

## 4. Approach

### 4.1 Classifier Design (`scripts/ci-autofix/classify.sh`)

Pure stdin-in / stdout-out POSIX sh script. Reads CI log lines from stdin, emits a single-line classification verdict to stdout:

```
classification=<mechanical|semantic|unknown>
sub_class=<trivial|non-trivial|none>
```

Pattern set (named constants — Phase 2 extracts to top-of-file `readonly` declarations, no inline literals; CLAUDE.local.md §14):

| Constant name | Regex meaning | Maps to |
|---------------|---------------|---------|
| `RX_TRIVIAL_GOFMT` | `gofmt|goimports needs|import order` | mechanical + trivial |
| `RX_TRIVIAL_WHITESPACE` | `trailing whitespace|file does not end with newline` | mechanical + trivial |
| `RX_MECH_ERRCHECK` | `errcheck|err.*not checked|unused (variable|import)` | mechanical + non-trivial |
| `RX_MECH_LINT_ST` | `staticcheck.*(ST1005|QF1003|SA[0-9]+)` | mechanical + non-trivial |
| `RX_MECH_TYPO_IMPORT` | `undeclared name|undefined: |missing import` | mechanical + non-trivial |
| `RX_SEMANTIC_RACE` | `data race|race detected|--- FAIL: TestRace` | semantic |
| `RX_SEMANTIC_PANIC` | `panic: |goroutine [0-9]+ \[.*\]:` | semantic |
| `RX_SEMANTIC_ASSERT` | `--- FAIL: Test.* (assertion|expected|got)` | semantic |
| `RX_SEMANTIC_DEADLOCK` | `fatal error: all goroutines are asleep|deadlock` | semantic |

Order matters: semantic 우선 매칭 → mechanical 매칭 → 모두 미매칭 시 `unknown` (orchestrator는 unknown을 semantic처럼 escalate). 이유: 한 로그에 lint warning + race가 동시에 보일 수 있고, race가 우선 처리되어야 한다.

Test mode: `MOAI_AUTOFIX_FIXTURE_DIR=/path/to/fixtures` 시 fixture 파일에서 입력 받음 (mock 주입). 16-language neutrality는 자연스럽게 충족 — 패턴은 CI 로그 텍스트 (Go 위주지만 Python/Node/Rust도 동일한 lint/race 메시지 family를 emit) 에 매칭하므로 언어별 AST 의존 0.

### 4.2 Iteration State Machine (orchestrator-side, skill body 안에 명시)

State file `.moai/state/ci-autofix-<PR>.json`:
```json
{
  "pr_number": 785,
  "iteration": 1,
  "max_iterations": 3,
  "last_classification": "mechanical",
  "last_subclass": "trivial",
  "last_action": "applied|escalated|aborted",
  "started_at": "2026-05-07T08:30:00Z",
  "history": [
    {"iter": 1, "classification": "mechanical", "subclass": "non-trivial", "result": "applied", "patch_sha": "abc1234"}
  ]
}
```

State machine (skill body는 이를 산문으로 명시; Wave 3 코드는 state 읽기/쓰기를 orchestrator-driven shell/Bash 호출로 처리, 별도 Go runtime 추가 없음):

```
[wave2-handoff-received]
  → state file 생성 (iteration=1)
  → log-fetch.sh로 failed log + PR diff 캡처
  → classify.sh로 분류
  → if classification == "semantic" || "unknown" → [escalate-immediate]
  → if iteration == 1 → [confirm-required]   (OQ2 cadence: iter 1 항상 confirm)
  → if subclass == "trivial" && iteration >= 2 → [silent-apply]
  → else → [confirm-required]

[confirm-required]
  → AskUserQuestion presents patch (1순위 권장 = "patch apply")
  → user accepts → expert-debug patch apply → git add/commit/push → invoke ci-watch (Wave 2)
  → user rejects → [escalate-user-choice]

[silent-apply]
  → expert-debug patch apply → git commit/push → ci-watch
  → log iter result to .moai/reports/ci-autofix/<PR>-<DATE>.md (no AskUserQuestion)

[ci-watch returns]
  → exit 0 (green) → state file 삭제 + skill exit success
  → exit 2 (still failing) → iteration++; goto top (until iteration > 3)

[iteration > 3]
  → AskUserQuestion mandatory (no silent timeout): continue manually / revise SPEC / abandon PR
  → blocking call (no timer); waits indefinitely for user response
  → record outcome to log

[escalate-immediate / escalate-user-choice]
  → AskUserQuestion with diagnosis (no patch attempt)
  → record outcome to log
```

### 4.3 Wave 2 → Wave 3 Handoff Consumption

Wave 2 watch loop emits exit code 2 + JSON to stdout. Skill body (orchestrator)는 다음 단계로 소비한다:

1. Bash 도구로 `scripts/ci-watch/run.sh <PR> <branch>` 호출 → stdout JSON 캡처 (Wave 2 contract)
2. exit 2 인 경우 stdout JSON을 parse (Wave 2 schema 검증된 Go-emitted, 4.5절 schema와 동일)
3. `failedChecks[0].runId` 와 `logUrl` 을 `scripts/ci-autofix/log-fetch.sh` 에 인자로 넘김
4. log + PR diff 를 `scripts/ci-autofix/classify.sh` 에 stdin 으로 pipe
5. classification 결과 + 원본 handoff JSON 을 `expert-debug` subagent spawn prompt 에 그대로 주입

Phase 2 implementation은 **Wave 2의 Go-emitted JSON 형식을 변경하지 않는다** (read-only consumer). 만일 Wave 3 진행 중 schema 확장 필요가 발견되면, Wave 2의 `internal/ciwatch/handoff.go`에 추가 필드를 더하는 follow-up commit으로 처리하고 본 Wave 3는 옛 schema에서 동작하도록 설계한다.

### 4.4 AskUserQuestion Cadence (OQ2 Wiring)

[HARD] orchestrator-only — 모든 AskUserQuestion 호출은 skill 본문(orchestrator surface)에서, subagent는 호출 금지 (`.claude/rules/moai/core/askuser-protocol.md` §Orchestrator–Subagent Boundary).

Cadence 표는 §5 OQ2 wiring matrix 에 명시. expert-debug subagent는 진단 + patch 후보를 Markdown 으로 반환만 하고, 사용자 컨펌은 orchestrator가 별도 `ToolSearch(query: "select:AskUserQuestion")` 후 `AskUserQuestion` 호출로 처리한다. semantic 실패는 expert-debug의 결과가 "diagnosis only" 모드로 반환되며, 같은 흐름으로 escalation AskUserQuestion 으로 전달된다.

### 4.5 Escalation Paths

| Trigger | Path |
|---------|------|
| classification == semantic / unknown (iter 1) | escalate-immediate; expert-debug = diagnosis only; no patch |
| iter 1, mechanical, user rejects patch | escalate-user-choice; record reason; halt loop |
| iter == 3 + still failing (mechanical) | mandatory AskUserQuestion (continue manually / revise SPEC / abandon PR) |
| log-fetch.sh failure (gh auth, network) | exit 1 + remediation hint; user에게 status 보고 후 manual fallback |
| classify.sh returns unknown for 2 consecutive iters | escalate "classifier unable to interpret" — user manual takeover |

## 5. Wave 2 → 3 Handoff Schema (Authoritative Contract)

Phase 2 Wave 3 코드는 다음 스키마를 read-only consumer로 신뢰한다. Schema 변경 시 Wave 2 follow-up commit 우선, Wave 3 후속 조정.

```json
{
  "prNumber": <int>,                      // GitHub PR number, > 0
  "branch": "<string>",                   // head branch name
  "failedChecks": [                       // required failures only (auxiliary 제외)
    {
      "name": "<string>",                 // check context (e.g. "Lint", "Test (ubuntu-latest)")
      "runId": "<string>",                // numeric string, gh API run ID
      "logUrl": "<string>",               // direct URL to failed run log
      "conclusionDetail": "<string>"      // optional human-readable summary
    }
  ],
  "auxiliaryFailCount": <int>,            // >= 0; informational, must NOT trigger T3
  "totalRequired": <int>                  // total count of tracked required checks
}
```

검증된 source: `internal/ciwatch/handoff.go::Handoff` struct (Wave 2, commit `5d3f6a4c1`). `internal/template/templates/.claude/skills/moai-workflow-ci-watch/SKILL.md` §"Wave 3 Handoff Schema" 와 100% 일치 확인 완료 (Phase 1 strategy 작성 시 Wave 2 산출물 직접 검증).

Wave 3 에서 추가로 필요한 필드 (orchestrator-managed, schema 외부):
- `iteration` (1..3) — Wave 3 state file 에서 관리, handoff payload 자체는 변경 없음
- `last_classification` / `last_subclass` — 동일

[HARD] Wave 3은 위 스키마의 `name`, `runId`, `logUrl` 필드를 stable contract로 취급. rename/remove 발생 시 Wave 2 → Wave 3 양쪽 동시 수정 필수.

## 6. OQ2 Cadence Wiring (State Machine Matrix)

이 표는 plan.md §5 W3-T01의 authoritative source. spec.md L156 (REQ-CIAUT-015) 의 "shall attempt at most 3 iterations of (debug → propose patch → user-confirm via AskUserQuestion → apply → push → re-watch) before mandatory escalation" 문구는 plan.md §7 OQ2 결정 ("iter 1 always confirm + iter 2-3 trivial silent")로 정정된 상태다. 충돌 시 plan.md OQ2 가 우선한다 (defect D1 처리, §7 참조).

| iteration | classification | sub_class | action | AskUserQuestion? |
|-----------|---------------|-----------|--------|------------------|
| 1 | mechanical | trivial | confirm + apply | YES (1순위 "patch apply") |
| 1 | mechanical | non-trivial | confirm + apply | YES (1순위 "patch apply") |
| 1 | semantic / unknown | n/a | escalate (no patch) | YES (diagnosis report only) |
| 2 | mechanical | trivial | silent + apply + log | NO (auto, log to report) |
| 2 | mechanical | non-trivial | confirm + apply | YES (1순위 "patch apply") |
| 2 | semantic / unknown | n/a | escalate (no patch) | YES |
| 3 | mechanical | trivial | silent + apply + log | NO (auto, log to report) |
| 3 | mechanical | non-trivial | confirm + apply | YES |
| 3 | semantic / unknown | n/a | escalate (no patch) | YES |
| post-iter-3 | any (still failing) | n/a | mandatory escalation | YES (blocking call, no timer) |

"trivial" 정의 (인용, plan.md §5 W3-T01): whitespace, gofmt/goimports, import-order. classify.sh의 `RX_TRIVIAL_*` 패턴 셋과 1:1 매칭.

## 7. Honest Scope Adjustments (review-2 defect treatment)

plan-auditor review-2 의 비차단 finding 4건에 대한 처리:

### D1 (MEDIUM) — REQ-CIAUT-015 vs §7 OQ2 모순

**상황**: spec.md L156의 REQ-CIAUT-015 본문은 unconditional confirm을 시사하나, §7 OQ2 결정은 "iter 1 always confirm + iter 2-3 trivial silent" 이다.

**Resolution**: plan.md §5 W3-T01 (= 본 문서 §6 OQ2 cadence wiring matrix) 가 authoritative. spec.md L156 문구 정정 (예: "shall attempt at most 3 iterations following the cadence defined in plan.md §5 W3-T01") 은 documentation hygiene으로 분류, **Wave 3 scope 외 follow-up commit 처리**. 이유: Wave 3은 implementation deliverable; spec rewording은 별도 docs PR 적합. 본 strategy는 코드 동작이 OQ2 matrix를 따르도록 보장.

### D2 (LOW) — Wave 2→3 metadata schema가 plan.md에만 존재

**Resolution**: 본 strategy §5 (Wave 2 → 3 Handoff Schema) 가 Phase 2의 de-facto contract. 향후 schema 진화 시 본 §5와 Wave 2 SKILL.md "Wave 3 Handoff Schema" 양쪽 동시 갱신.

### D3 (LOW) — AC-CIAUT-008 "silent timeout 금지" invariant가 W3-T07에 명시되지 않음

**Resolution**: tasks-wave3.md 의 W3-T07 acceptance criterion에 "AskUserQuestion blocking call (no timer); 사용자 응답 전까지 무한 대기" 문구를 명시 (tasks-wave3.md row 참조). 본 strategy §4.2 state machine "[iteration > 3]" 분기에도 동일 invariant 명시.

### D5 (INFO) — Iteration count persistence 위치 미정

**Resolution**: `.moai/state/ci-autofix-<PR>.json` 채택 (Wave 2의 `.moai/state/ci-watch-active.flag` 파일 인접 위치, 동일 디렉터리는 이미 gitignore). 형식은 §4.2 state machine 의 JSON. crash recovery 는 mtime 기반 staleness check 로 Wave 2 패턴 그대로 (24h 이상 stale 시 새 invocation 이 reclaim 가능).

## 8. Risks & Mitigations

| ID | Risk | Mitigation |
|----|------|-----------|
| W3-R1 | classify.sh 가 false negative — semantic을 mechanical로 잘못 분류 → 잘못된 patch 제안 | iter 1 always confirm (OQ2)으로 사용자 1차 검증; semantic regex 우선 매칭 ordering; user reject 시 escalate-user-choice 즉시 진입 |
| W3-R2 | classify.sh 가 false positive — mechanical을 semantic으로 분류 → patch 안 됐는데 escalate | unknown은 semantic 처럼 escalate (intentional 보수성); user가 "manually fix" 선택 후 다음 watch 사이클로 회복 가능 |
| W3-R3 | log-fetch.sh 가 큰 로그 (50MB+) 를 메모리에 로드 → 토큰 폭증 | `gh run view --log-failed | head -c 200000` (200KB 캡); 더 필요하면 user manual takeover 안내 |
| W3-R4 | iteration 도중 user가 다른 push 함 → state 어긋남 | state file 의 last commit SHA 기록 → 매 iteration 시작 시 `git rev-parse HEAD` 와 비교, drift 감지 시 abort + 사용자 알림 |
| W3-R5 | concurrent T3 invocation (다른 PR에 대해) → state 파일 충돌 | 파일명에 `<PR>` 포함 (`ci-autofix-<PR>.json`) → PR별 독립; 동일 PR 중복 진입은 mtime check |
| W3-R6 | expert-debug 가 patch 제안 없이 빈 응답 | 빈 응답 = subagent 진단 실패 → unknown classification 처럼 escalate; max 1 retry 후 user takeover |

## 9. Verification Plan

각 task 별 acceptance + AC 매핑 (자세한 테이블은 tasks-wave3.md):

- W3-T01 (state machine + cadence) → AC-CIAUT-008 (iteration cap 3 + mandatory escalation; no silent timeout)
- W3-T02 (classifier patterns) → AC-CIAUT-006 (mechanical resolve), AC-CIAUT-007 (semantic escalate)
- W3-T03 (log-fetch.sh) → AC-CIAUT-006 entry condition
- W3-T04 (expert-debug extension) → AC-CIAUT-006 patch quality
- W3-T05 (orchestrator iteration loop) → AC-CIAUT-006 + AC-CIAUT-008
- W3-T06 (semantic immediate escalation) → AC-CIAUT-007
- W3-T07 (mandatory escalation at iter 3) → AC-CIAUT-008 ("silent timeout 금지" invariant)
- W3-T08 (logging schema/writer) → AC-CIAUT-006/007/008 audit trail (`.moai/reports/ci-autofix/<PR>-<DATE>.md`)
- W3-T09 (Wave 2 metadata consumption) → AC-CIAUT-006 entry; cross-wave contract integration
- W3-T10 (ci-autofix-protocol.md HARD rule) → AC-CIAUT-008 invariant (no force-push, every patch is new commit)

Wave-level DoD:
- `make ci-local` 통과
- 5-PR sweep replay scenario 테스트:
  - PR #739 errcheck (mechanical) → 1-2 iter 안에 자동 해결, AskUserQuestion confirm 1회
  - PR #747 ETXTBSY race (semantic) → iter 1 즉시 escalate, patch 시도 0
  - 가상 시나리오: 3-iter mechanical 실패 → mandatory escalation AskUserQuestion 발생
- `internal/template/embedded.go` 재생성 (Template-First 검증)
- shellcheck -s sh 통과 (POSIX-only, no bashisms)
- expert-debug.md extension은 기존 본문 보존 (additive only) — diff 검증
- ci-autofix-protocol.md 의 모든 [HARD] 마커가 paths frontmatter로 적절히 auto-load 됨

## 10. Dependencies on Wave 2

[HARD] Wave 3 deliverables는 Wave 2의 다음 산출물에 의존하며, Wave 2 main 머지 후에만 Phase 2 시작 가능 (이미 충족: `5d3f6a4c1` main):

- `scripts/ci-watch/run.sh` — exit 2 + JSON stdout (Wave 3 entry trigger)
- `internal/ciwatch/handoff.go::Handoff` 구조체 — JSON schema source of truth (§5 참조)
- `internal/template/templates/.claude/skills/moai-workflow-ci-watch/SKILL.md` §"Wave 3 Handoff Schema" — Wave 3가 cross-reference 하는 스키마 문서
- `.claude/rules/moai/workflow/ci-watch-protocol.md` §"T3 Handoff Format" — Wave 3 protocol 이 후속 cross-link 하는 contract 문서
- `.github/required-checks.yml` (Wave 1 SSoT, Wave 2가 read-only consume) — Wave 3는 직접 참조하지 않음 (transitive — log 분석은 check 이름과 무관하게 패턴 매칭)

cross-reference 무결성: Wave 3 의 `ci-autofix-protocol.md` 는 Wave 2 의 `ci-watch-protocol.md` §"T3 Handoff Format" 을 인용하며 (단방향 의존), Wave 2 산출물 본문은 Wave 3 머지 후에도 변경하지 않는다. 만일 schema 진화 필요 시 Wave 2 산출물에 minor version bump 후 Wave 3 도 동기 갱신 — 본 strategy 는 그 절차를 §7 D2 처리에 명시.

---

Version: 0.1.0
Status: pending Phase 2 (manager-tdd 위임 대기)
Last Updated: 2026-05-07
