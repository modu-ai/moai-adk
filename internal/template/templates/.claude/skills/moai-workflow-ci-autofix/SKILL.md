---
name: moai-workflow-ci-autofix
description: CI auto-fix loop skill — receives Wave 2 ci-watch failure handoff, classifies failures (mechanical vs semantic), attempts automated patch (max 3 iterations), and escalates via AskUserQuestion. HARD invocation contract in .claude/rules/moai/workflow/ci-autofix-protocol.md.
version: "1.0.0"
tools: Bash,Read
level1_tokens: 120
level2_tokens: 5200
triggers:
  - "ci.*fail.*auto.*fix"
  - "moai.*autofix"
  - "T3.*loop"
  - "ci.*autofix"
paths:
  - ".claude/rules/moai/workflow/ci-autofix-protocol.md"
---

# CI Auto-Fix Loop (`moai-workflow-ci-autofix`)

## Quick Reference

Wave 2 CI watch loop(`moai-workflow-ci-watch`)가 required check 실패(exit 2 + JSON handoff)를 emit하면,
이 skill이 인수인계받아 다음을 수행한다:

1. `scripts/ci-autofix/log-fetch.sh`로 실패 로그 + PR diff 캡처
2. `scripts/ci-autofix/classify.sh`로 실패 유형 분류 (mechanical / semantic / unknown)
3. OQ2 cadence matrix에 따라 `expert-debug` subagent에 진단 위임
4. mechanical 실패: patch 제안 → orchestrator가 AskUserQuestion → apply → push → ci-watch 재호출
5. semantic 실패: 즉시 AskUserQuestion 에스컬레이션 (patch 시도 없음)
6. 최대 3회 반복 후 mandatory blocking AskUserQuestion

**[HARD]** AskUserQuestion은 orchestrator(이 skill을 실행하는 메인 세션)만 호출한다.
`expert-debug` subagent는 절대 AskUserQuestion을 호출하지 않는다.

### 상태 파일

```
.moai/state/ci-autofix-<PR번호>.json
```

Wave 2의 `.moai/state/ci-watch-active.flag`와 동일한 디렉터리에 위치한다.
파일명에 PR 번호가 포함되므로 동일 PR 중복 진입 및 서로 다른 PR 간 충돌이 방지된다.

---

## Implementation Guide

### Wave 2 → 3 Handoff 소비

Wave 2 (`scripts/ci-watch/run.sh`)가 exit 2로 종료하면 stdout에 JSON이 출력된다:

```json
{
  "prNumber": 785,
  "branch": "feat/my-feature",
  "failedChecks": [
    {
      "name": "Lint",
      "runId": "12345678",
      "logUrl": "https://github.com/.../actions/runs/12345678",
      "conclusionDetail": ""
    }
  ],
  "auxiliaryFailCount": 1,
  "totalRequired": 6
}
```

Schema source: `internal/ciwatch/handoff.go::Handoff` (Wave 2, commit `5d3f6a4c1`).
`failedChecks[]`는 required 실패만 포함한다. auxiliary 실패는 `auxiliaryFailCount`에만 카운트된다.

**[HARD]** orchestrator는 위 JSON의 `failedChecks[].runId`와 `logUrl` 필드를
`scripts/ci-autofix/log-fetch.sh`에 인자로 넘긴다.
schema 필드명 변경 시 Wave 2 + Wave 3 양쪽 동시 수정 필수.

### 상태 파일 초기화 및 관리

```bash
# Wave 2 handoff JSON을 파싱하여 상태 파일 생성
PR_NUMBER=$(printf '%s\n' "$HANDOFF_JSON" | jq -r '.prNumber')
STATE_FILE=".moai/state/ci-autofix-${PR_NUMBER}.json"

# 상태 파일 초기화 (iteration=1)
cat > "$STATE_FILE" <<EOF
{
  "pr_number": $PR_NUMBER,
  "iteration": 1,
  "max_iterations": 3,
  "last_classification": "",
  "last_subclass": "",
  "last_action": "",
  "started_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "history": []
}
EOF
```

crash recovery: 24시간 이상된 stale 상태 파일은 새 invocation이 덮어쓸 수 있다.

### 로그 캡처 및 분류

```bash
# failed check의 첫 번째 runId를 추출
RUN_ID=$(printf '%s\n' "$HANDOFF_JSON" | jq -r '.failedChecks[0].runId')
PR_NUMBER=$(printf '%s\n' "$HANDOFF_JSON" | jq -r '.prNumber')

# 로그 + diff 캡처
LOG_AND_DIFF=$(sh scripts/ci-autofix/log-fetch.sh "$RUN_ID" "$PR_NUMBER")

# 분류
CLASSIFICATION=$(printf '%s\n' "$LOG_AND_DIFF" | sh scripts/ci-autofix/classify.sh)
CLASS=$(printf '%s\n' "$CLASSIFICATION" | grep '^classification=' | cut -d= -f2)
SUB=$(printf '%s\n' "$CLASSIFICATION" | grep '^sub_class=' | cut -d= -f2)
```

### OQ2 Cadence Matrix (Authoritative)

이 표가 iteration별 행동의 단일 진실 공급원이다:

| iteration | classification | sub_class | action | AskUserQuestion? |
|-----------|---------------|-----------|--------|------------------|
| 1 | mechanical | trivial | confirm + apply | YES (1순위 "패치 적용 (권장)") |
| 1 | mechanical | non-trivial | confirm + apply | YES (1순위 "패치 적용 (권장)") |
| 1 | semantic / unknown | n/a | escalate (no patch) | YES (진단 보고서 포함) |
| 2 | mechanical | trivial | silent + apply + log | NO (자동, 보고서에만 기록) |
| 2 | mechanical | non-trivial | confirm + apply | YES |
| 2 | semantic / unknown | n/a | escalate (no patch) | YES |
| 3 | mechanical | trivial | silent + apply + log | NO |
| 3 | mechanical | non-trivial | confirm + apply | YES |
| 3 | semantic / unknown | n/a | escalate (no patch) | YES |
| post-iter-3 | any (still failing) | n/a | mandatory escalation | YES (blocking, no timer) |

"trivial" 정의: whitespace, gofmt/goimports, import-order. `classify.sh`의 `RX_TRIVIAL_*` 패턴과 1:1 매칭.

### Orchestrator 상태 머신 흐름

```
[wave2-handoff-received]
  → 상태 파일 생성 (iteration=1)
  → log-fetch.sh로 failed log + PR diff 캡처
  → classify.sh로 분류
  → if classification == "semantic" || "unknown" → [escalate-immediate]
  → OQ2 matrix 조회
  → if iteration == 1 → [confirm-required]   (iter 1 항상 confirm)
  → if subclass == "trivial" && iteration >= 2 → [silent-apply]
  → else → [confirm-required]

[confirm-required]
  → ToolSearch(query: "select:AskUserQuestion")
  → AskUserQuestion 패치 내용 제시 (1순위 = "패치 적용 (권장)")
  → user accepts → expert-debug patch apply → git add/commit/push (force-push 금지)
                 → ci-watch run.sh 재호출
  → user rejects → [escalate-user-choice]

[silent-apply]
  → expert-debug patch apply → git commit/push (force-push 금지)
  → audit log에 iteration 결과 기록 (AskUserQuestion 없음)
  → ci-watch run.sh 재호출

[ci-watch returns]
  → exit 0 (green) → 상태 파일 삭제 + skill 성공 종료
  → exit 2 (still failing) → iteration++ → state file 갱신 → 상태 머신 처음으로

[iteration > 3]
  → ToolSearch(query: "select:AskUserQuestion")
  → AskUserQuestion mandatory (blocking, no timer, 사용자 응답 전까지 무한 대기)
  → 옵션 3개 + Other: (권장) 직접 수동 수정 / SPEC 수정 / PR 포기
  → 모든 iteration 결과를 audit log에 첨부하여 보고

[escalate-immediate / escalate-user-choice]
  → ToolSearch(query: "select:AskUserQuestion")
  → AskUserQuestion 진단 보고서 포함 (patch 시도 없음)
  → audit log에 escalation 이유 기록
```

### expert-debug Subagent 호출 패턴

orchestrator가 `expert-debug`를 호출할 때 spawn prompt에 주입할 컨텍스트:

```
## CI Auto-Fix Context

**Wave 2 Handoff JSON:**
<handoff_json>

**Classification Result:**
- classification: <mechanical|semantic|unknown>
- sub_class: <trivial|non-trivial|none>

**Failed CI Log + PR Diff:**
<log_and_diff_content>

**Mode:**
- If classification == "mechanical": propose a patch in unified diff format
- If classification == "semantic" or "unknown": return diagnosis only (NO patch)

[HARD] AskUserQuestion 호출 금지. patch 제안 또는 diagnosis를 Markdown으로 반환만 한다.
```

expert-debug는 mechanical 실패 시 patch를 제안하고, semantic 실패 시 진단 보고서만 반환한다.
orchestrator가 patch를 받은 후 AskUserQuestion으로 사용자 컨펌을 받는다.

### Audit Log 작성

모든 iteration이 `.moai/reports/ci-autofix/<PR-NNN>-<YYYY-MM-DD>.md`에 기록된다:

```markdown
# CI Auto-Fix Log — PR #785 — 2026-05-07

## Iteration 1

- **classification**: mechanical
- **sub_class**: non-trivial
- **action**: applied
- **patch_sha**: abc1234
- **escalation_reason**: N/A

## Iteration 2

- **classification**: mechanical
- **sub_class**: trivial
- **action**: applied (silent)
- **patch_sha**: def5678
- **escalation_reason**: N/A
```

Bash 도구로 append-only 기록. 첫 iteration 시 헤더 + Iteration 1 섹션 신규 작성.
이후 iteration은 `## Iteration N` 섹션을 append.

### 보고서 디렉터리 생성

```bash
mkdir -p .moai/reports/ci-autofix/
REPORT_FILE=".moai/reports/ci-autofix/PR-${PR_NUMBER}-$(date +%Y-%m-%d).md"

# 첫 iteration: 헤더 작성
if [ ! -f "$REPORT_FILE" ]; then
    printf '# CI Auto-Fix Log — PR #%s — %s\n\n' \
        "$PR_NUMBER" "$(date +%Y-%m-%d)" > "$REPORT_FILE"
fi

# 각 iteration: append
printf '\n## Iteration %s\n\n- **classification**: %s\n- **sub_class**: %s\n- **action**: %s\n' \
    "$ITER" "$CLASS" "$SUB" "$ACTION" >> "$REPORT_FILE"
```

---

## Advanced Patterns

### 에스컬레이션 경로 요약

| 트리거 | 경로 |
|--------|------|
| classification == semantic / unknown (any iter) | escalate-immediate: expert-debug = diagnosis only; patch 없음 |
| iter 1, mechanical, user rejects patch | escalate-user-choice: 이유 기록 + loop halt |
| iter == 3 + still failing (mechanical) | mandatory AskUserQuestion (blocking, no timer) |
| log-fetch.sh 실패 (gh auth, 네트워크) | exit 1 + remediation hint: user에게 manual fallback 안내 |
| classify.sh가 2 연속 unknown | "classifier 해석 불가" → user manual takeover |

### Force-Push 금지 (Invariant)

[HARD] 모든 auto-fix patch는 PR 브랜치에 새 commit으로 추가된다.
`git push --force`, `git push -f`, `git push --force-with-lease` 사용 절대 금지.
이 invariant는 `ci-autofix-protocol.md`에 HARD 규칙으로 명시된다.

### Wave 2 Cross-Reference

- Schema: `internal/ciwatch/handoff.go::Handoff`
- SKILL: `internal/template/templates/.claude/skills/moai-workflow-ci-watch/SKILL.md` §"Wave 3 Handoff Schema"
- Protocol: `internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md` §"T3 Handoff Format"

---

## Works Well With

- `moai-workflow-ci-watch` → Wave 2 handoff 소스
- `expert-debug` → 실패 로그 진단 + patch 제안
- `scripts/ci-autofix/classify.sh` → mechanical vs semantic 분류
- `scripts/ci-autofix/log-fetch.sh` → 실패 로그 + PR diff 캡처
- `.claude/rules/moai/workflow/ci-autofix-protocol.md` → HARD invocation contract

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "semantic 실패도 auto-patch 시도해보자" | semantic 실패(race condition, test assertion)는 컨텍스트 없이 auto-patch 불가. 잘못된 patch가 더 큰 문제를 만든다. |
| "iter 3 이후에 timeout 걸면 편하다" | AC-CIAUT-008 invariant — silent timeout 금지. 사용자가 반드시 결정해야 한다. |
| "force-push로 히스토리 정리하면 깔끔하다" | PR 브랜치에 force-push하면 리뷰어가 diff를 잃는다. 항상 새 commit. |
| "trivial fix는 confirm 없이 바로 적용하자" | iter 1은 항상 confirm (OQ2 matrix). 사용자가 처음 patch를 검토해야 한다. |
<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- expert-debug subagent가 AskUserQuestion을 호출하려 함 (HARD 위반: orchestrator 전용)
- iteration 4에서 auto-continue (no AskUserQuestion) — blocking call 누락
- git push --force 또는 push -f 사용 — force-push 금지 invariant 위반
- semantic 실패에 대해 patch 생성 시도 — 즉시 escalate 경로 누락
- 상태 파일 미생성 — iteration 카운터 유실, 무한 루프 위험
<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Wave 2 exit 2 + JSON이 state file 초기화를 트리거하는가
- [ ] iter 1 모든 classification이 AskUserQuestion을 발생시키는가
- [ ] iter 2-3 trivial mechanical이 silent apply + log only인가
- [ ] semantic/unknown이 patch 없이 즉시 AskUserQuestion 에스컬레이션하는가
- [ ] iter > 3 진입 시 blocking AskUserQuestion (no timer)가 발생하는가
- [ ] `.moai/reports/ci-autofix/<PR>-<DATE>.md`에 모든 iteration이 기록되는가
- [ ] git push --force / -f가 사용되지 않는가 (`grep -r 'push -f\|push --force' scripts/ci-autofix/`)
<!-- moai:evolvable-end -->
