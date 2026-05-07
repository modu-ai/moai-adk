# Wave 3 Atomic Tasks (Phase 1 Output)

> Companion to strategy-wave3.md. SPEC-V3R3-CI-AUTONOMY-001 Wave 3 — Auto-Fix Loop on CI Fail (T3).
> Generated: 2026-05-07. Methodology: TDD. Wave Base: `origin/main 5d3f6a4c1` (Wave 2 PR #788 merged).

## Atomic Task Table

| Task ID | Description | Files (provisional) | REQ | AC | Status |
|---------|-------------|---------------------|-----|----|--------|
| W3-T01 | Iteration cadence state machine wiring (OQ2 matrix: iter 1 always confirm; iter 2-3 trivial silent + log; non-trivial confirm). State persisted to `.moai/state/ci-autofix-<PR>.json` (path constant only — no Go runtime added). Skill body 안에 산문 + flowchart로 명시. | `internal/template/templates/.claude/skills/moai-workflow-ci-autofix/SKILL.md` (Implementation 섹션 일부) | REQ-CIAUT-014, 015, 016 | AC-CIAUT-006, AC-CIAUT-008 | completed |
| W3-T02 | Failure classifier — `scripts/ci-autofix/classify.sh` (POSIX sh). Top-of-file `readonly` 패턴 상수 9개 (`RX_TRIVIAL_*`, `RX_MECH_*`, `RX_SEMANTIC_*` per strategy-wave3.md §4.1). 출력: `classification=<mechanical|semantic|unknown>` + `sub_class=<trivial|non-trivial|none>`. semantic 우선 매칭, unknown은 semantic처럼 처리. | `scripts/ci-autofix/classify.sh`, `scripts/ci-autofix/test/classify_test.sh` | REQ-CIAUT-016, 017 | AC-CIAUT-006 (errcheck → mechanical), AC-CIAUT-007 (race → semantic) | completed |
| W3-T03 | Log fetcher — `scripts/ci-autofix/log-fetch.sh` (POSIX sh). `gh run view <run-id> --log-failed | head -c 200000` 캡; 추가로 `gh pr diff <PR>` 캡처 후 결합 출력. mock 주입은 `MOAI_AUTOFIX_GH=` env로. | `scripts/ci-autofix/log-fetch.sh`, `scripts/ci-autofix/test/log_fetch_test.sh` | REQ-CIAUT-014 | AC-CIAUT-006 (entry condition: log + diff 캡처 가능) | completed |
| W3-T04 | `expert-debug` agent extension — 기존 본문 보존하면서 "CI Failure Interpretation" 섹션 additive 추가. 입력 형식 (Wave 2 Handoff JSON + classify.sh 결과 + log-fetch.sh 출력) + 출력 형식 (mechanical: patch 제안 markdown / semantic: diagnosis only) 명시. AskUserQuestion 호출 금지 reminder. | `internal/template/templates/.claude/agents/expert-debug.md` (extend, not rewrite) | REQ-CIAUT-014, 016, 017 | AC-CIAUT-006 (patch quality), AC-CIAUT-007 (diagnosis only) | completed |
| W3-T05 | Orchestrator iteration loop (skill body 안의 Implementation section). 흐름: Wave 2 exit 2 수신 → log-fetch → classify → expert-debug → cadence 분기 → AskUserQuestion 또는 silent apply → git commit/push (force-push 금지) → ci-watch 재호출 → state file iter++ → 반복. Pure-orchestrator surface — 모든 AskUserQuestion 호출은 `ToolSearch(query: "select:AskUserQuestion")` 선행. | `internal/template/templates/.claude/skills/moai-workflow-ci-autofix/SKILL.md` (Implementation 섹션 메인) | REQ-CIAUT-014, 015, 016 | AC-CIAUT-006, AC-CIAUT-008 | completed |
| W3-T06 | Semantic immediate escalation 경로. `expert-debug` 는 진단 markdown 만 반환, patch field 비움. orchestrator는 AskUserQuestion 으로 사용자 결정 받음 (continue manually / revise SPEC / abandon PR). iter 1 도 동일 분기 (semantic은 iteration 카운트 무관 즉시 escalate). | `internal/template/templates/.claude/skills/moai-workflow-ci-autofix/SKILL.md` (Implementation 분기), `internal/template/templates/.claude/agents/expert-debug.md` (diagnosis-only 모드 명시) | REQ-CIAUT-017 | AC-CIAUT-007 | completed |
| W3-T07 | Mandatory escalation at iteration 3. iter > 3 진입 시 AskUserQuestion **blocking call (no timer)** — 사용자 응답 전까지 무한 대기 (silent timeout 금지). 옵션 3개 + Other (1순위 "(권장) 직접 수동 수정"). 모든 iteration 결과를 보고서로 첨부. | `internal/template/templates/.claude/skills/moai-workflow-ci-autofix/SKILL.md` (Implementation 분기), `internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md` (HARD rule) | REQ-CIAUT-018 | AC-CIAUT-008 ("silent timeout 금지" invariant) | completed |
| W3-T08 | Audit log writer schema — `.moai/reports/ci-autofix/<PR-NNN>-<YYYY-MM-DD>.md` per spec REQ-CIAUT-019. 매 iteration entry: iteration #, classification, sub_class, action (applied/escalated/aborted), patch SHA, escalation reason. Skill body 가 Bash 도구로 append. 첫 호출 시 헤더 작성, 이후 append-only. | `internal/template/templates/.claude/skills/moai-workflow-ci-autofix/SKILL.md` (logging 섹션) | REQ-CIAUT-019 | AC-CIAUT-006/007/008 audit trail | completed |
| W3-T09 | Wave 2 → 3 handoff integration. strategy-wave3.md §5 schema 를 skill 본문에 명시 + `internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md` §"T3 Handoff Format" cross-reference. JSON parse는 `jq` (이미 Wave 2가 의존, `moai doctor`에서 검증 중) 사용. schema 의 `failedChecks[].runId` 와 `logUrl` 필드를 `log-fetch.sh` 입력으로 매핑. | `internal/template/templates/.claude/skills/moai-workflow-ci-autofix/SKILL.md` (Wave 2 contract 섹션) | REQ-CIAUT-014 | AC-CIAUT-006 (cross-wave contract 무결성) | completed |
| W3-T10 | `ci-autofix-protocol.md` HARD rule 작성. paths frontmatter 로 skill 활성 시 자동 로드. 핵심 [HARD]: max 3 iterations, 모든 patch는 새 commit (force-push 금지), AskUserQuestion 은 orchestrator-only, semantic 자동 patch 금지, iter 3 이후 blocking AskUserQuestion (no silent timeout), audit log 강제. | `internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md` | REQ-CIAUT-014..019 (전체) | AC-CIAUT-006/007/008 invariants | completed |

## File Ownership Assignment (Agent Teams 모드 시; Solo 모드도 동일 경계)

### Implementer Scope (write access)

```
internal/template/templates/.claude/skills/moai-workflow-ci-autofix/      # SKILL.md (신규)
internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md  # 신규
internal/template/templates/.claude/agents/expert-debug.md                # additive 확장 (기존 본문 보존)
scripts/ci-autofix/                                                       # classify.sh, log-fetch.sh (신규)
```

### Tester Scope (test 파일만 write, production 코드 read-only)

```
scripts/ci-autofix/test/                                                  # classify_test.sh, log_fetch_test.sh
```

### Read-Only Scope (analyst/reviewer)

```
internal/ciwatch/                                                         # Wave 2 산출물 — schema source of truth (read-only)
internal/template/templates/.claude/skills/moai-workflow-ci-watch/        # Wave 2 SKILL.md — cross-reference 소스
internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md  # Wave 2 rule — cross-reference 소스
.github/required-checks.yml                                               # Wave 1 SSoT — transitive (직접 참조 안 함)
scripts/ci-watch/                                                         # Wave 2 — read-only consumer
```

### Implicit Read Access (모든 role)

- `.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/` (spec.md, plan.md, acceptance.md, strategy-wave3.md, 본 파일)
- `.claude/rules/moai/**` (auto-loaded rules — 특히 askuser-protocol.md, agent-common-protocol.md)
- Wave 1 + Wave 2 산출물 (read-only consumer)

## AC Mapping

| Wave 3 Task | Drives AC | Validation |
|-------------|-----------|------------|
| W3-T02 + W3-T03 + W3-T04 + W3-T05 | AC-CIAUT-006 (mechanical resolve, replay PR #739 errcheck) | classify.sh 가 errcheck 패턴 → mechanical/non-trivial 분류; expert-debug patch 제안 → user accept → apply → 다음 ci-watch green; iter 1 confirm AskUserQuestion 발생 |
| W3-T02 + W3-T06 | AC-CIAUT-007 (semantic immediate escalation, replay PR #747 ETXTBSY) | classify.sh 가 race 패턴 → semantic 분류; expert-debug diagnosis-only 반환; orchestrator iter 1 즉시 escalation AskUserQuestion (patch 시도 0) |
| W3-T01 + W3-T07 + W3-T08 | AC-CIAUT-008 (iteration cap 3 + mandatory escalation; no silent timeout) | 가상 시나리오: 3-iter 모두 fail; iter 4 진입 시 mandatory blocking AskUserQuestion (no timer); 모든 iter 결과 audit log 에 기록 |
| W3-T09 | Wave 2 ↔ Wave 3 contract 무결성 | Wave 2 emit JSON 그대로 consume; schema field rename 없음; log-fetch.sh 가 runId/logUrl 정확히 인입 |
| W3-T10 | 모든 AC invariants (force-push 금지, orchestrator-only AskUserQuestion, no auto-patch for semantic) | rule paths frontmatter auto-load; HARD 마커 검증 (`grep -c '\[HARD\]'`); cross-reference 링크 검증 |

## TRUST 5 Targets (Wave 3 SPEC-Level DoD)

| Pillar | Target | Verification |
|--------|--------|--------------|
| **Tested** | shell test harness 가 classify.sh 9개 패턴 모두 커버 (mechanical: 5 / semantic: 4) + log-fetch.sh mock 시나리오 4개 (success, gh-auth-fail, large-log-cap, missing-run-id) | `bash scripts/ci-autofix/test/classify_test.sh && bash scripts/ci-autofix/test/log_fetch_test.sh` |
| **Readable** | shellcheck -s sh clean on `scripts/ci-autofix/**/*.sh`; SKILL.md ≤ 500 lines (Progressive Disclosure 분리 시 modules/ 사용) | `shellcheck -s sh scripts/ci-autofix/**/*.sh`; `wc -l SKILL.md` |
| **Unified** | classify.sh 패턴 상수 명명이 Wave 1 `_common.sh` 컨벤션 따름 (UPPER_SNAKE, `readonly` prefix); SKILL.md frontmatter 가 skill-authoring.md schema 100% 준수 | manual code review + frontmatter validator (`internal/template/skills_audit_test.go` 가 자동 검증) |
| **Secured** | expert-debug 가 patch 제안 시 절대 secrets/keys 변경 안 함 (skill rule 명시); auto-fix loop 가 .env, credentials 파일 modify 금지; 모든 patch 는 새 commit (force-push 금지, no `git push -f`) | grep -rE 'force-push\|push -f' rules — Wave 3 산출물에 검출 0; rule 본문 explicit invariant 검증 |
| **Trackable** | 모든 commit 이 SPEC-V3R3-CI-AUTONOMY-001 W3 reference 포함; Conventional Commits + 🗿 MoAI co-author trailer | `git log --grep='SPEC-V3R3-CI-AUTONOMY-001 W3'` |

## Per-Wave DoD Checklist

- [ ] 모든 10개 W3 task 완료 (위 표)
- [ ] Template-First mirror 검증: 5개 deliverable 중 3개 (.claude/skills, .claude/rules, .claude/agents) 가 `internal/template/templates/` 하위에 위치; scripts/ci-autofix/ 는 repo-rooted (mirror 불필요 — Wave 1 패턴과 동일)
- [ ] `make build` 가 `internal/template/embedded.go` 재생성 성공 (변경된 3개 .claude/ 파일 반영)
- [ ] `make ci-local` 통과 (Wave 1 framework 가 Wave 3 변경에 회귀 없음)
- [ ] `go test ./...` 통과 (Wave 2 ciwatch 패키지 회귀 검증)
- [ ] Wave 3 specific acceptance: AC-CIAUT-006/007/008 manual replay 검증
  - [ ] PR #739 errcheck replay → mechanical 자동 처리, AskUserQuestion 1회 (iter 1 confirm)
  - [ ] PR #747 ETXTBSY replay → semantic 즉시 escalation, patch 시도 0
  - [ ] 가상 3-iter mechanical fail → mandatory blocking AskUserQuestion
- [ ] No hardcoded URLs/models/regex literals in code paths (모든 패턴은 `readonly` 상수)
- [ ] No release/tag automation 도입 (negative grep on Wave 3 산출물)
- [ ] expert-debug.md 기존 본문 100% 보존 (additive only — `git diff` 검증)
- [ ] PR labeled with `type:feature`, `priority:P0`, `area:ci`, `area:templates`
- [ ] Conventional Commits + 🗿 MoAI co-author trailer 모든 commit
- [ ] CHANGELOG.md 에 Wave 3 머지 entry

## Out-of-Scope (Wave 3 Exclusions)

- Wave 4 (T4 auxiliary workflow cleanup)
- Wave 5 (T6 worktree state guard)
- Wave 6 (T7 i18n validator)
- Wave 7 (T8 BODP)
- semantic 실패의 자동 patch 시도 (영구 prohibited)
- multi-PR 동시 auto-fix (single-PR-at-a-time, Wave 2 모델 계승)
- semantic patch ML 학습 / heuristic 누적
- expert-debug 의 비-CI 동작 변경 (additive only — 기존 진단 capability 보존)
- spec.md L156 (REQ-CIAUT-015) wording 정정 — defect D1, follow-up commit 처리
- release tag/GoReleaser 자동 실행 (`feedback_release_no_autoexec.md` 영구 금지)
- Cross-platform PowerShell variant of `scripts/ci-autofix/*.sh` (Windows 사용자는 git-bash; Wave 1/2 패턴 계승)

## Honest Scope Concerns

1. **Classifier false-negative on novel CI failure patterns**: 현재 9개 regex 패턴 셋은 5-PR sweep 사례 + 일반 Go/lint 출력에 기반. 새로운 lint tool 추가나 Go 언어 진화로 미커버 패턴 발생 가능. Mitigation: unknown classification 은 semantic 처럼 escalate (보수적 default) — 사용자가 manual takeover 후 패턴 추가 가능; follow-up SPEC 으로 패턴 자동 학습 검토 (Wave 3 scope 외).
2. **State file race on concurrent T3 invocation**: PR별 파일 분리 (`ci-autofix-<PR>.json`) 로 충돌 방지하나, 동일 PR 에 대해 Wave 2 watch loop 가 두 번 실행되어 두 번 exit 2 emit 시 race 가능. Mitigation: Wave 2 의 single-watch-per-repo 모델이 상위에서 차단 — Wave 3 는 single watch 신뢰. 만일 회귀 발견 시 fcntl 기반 lock 추가 (Wave 3 scope 외 follow-up).
3. **`gh run view --log-failed` 가 매우 큰 로그를 emit 하는 경우**: 200KB 캡으로 mitigated 하나, 캡 이전에 발생한 실제 root cause 가 잘리는 경우 classifier 가 잘못된 분류. Mitigation: log-fetch.sh 가 캡 적용 시 stderr 로 "log truncated to 200KB; manual review may be needed" 경고 출력; user 가 이 경고를 보고 manual takeover 가능.
4. **iter 3 이후 mandatory escalation 의 blocking 무한 대기**: 사용자가 응답 안 하면 orchestrator session이 멈춰있음. silent timeout 도입은 AC-CIAUT-008 invariant 위반이므로 불가. Mitigation: blocking 은 의도된 design (강제 user attention); 사용자가 명시적으로 abandon 옵션 선택 가능; session 자체 종료는 사용자 자유.

No hard blockers identified. Wave 3 ready for Phase 2 (manager-tdd) delegation upon strategy + tasks approval.

---

Version: 0.1.0
Status: pending implementation
Last Updated: 2026-05-07
