# SPEC-LEAD-AUTOPUSH-001 — Acceptance Criteria

> Card t1346 · Tier M · Given-When-Then, binary-testable · 2026-09-29

## §D AC Matrix

| AC | REQ | Criterion (binary) | Class | RED-now | Green path |
|---|---|---|---|---|---|
| AC-001 | REQ-004 | §4 names `lead_push_threshold` + the count command | release-blocking | E-1 (0 hits) | M1 |
| AC-002 | REQ-002/003 | §4 carries the green-conditional (window re-measurement gate + last-push CI hold) | release-blocking | E-2 (no green wording) | M1 |
| AC-003 | REQ-001 | §7 batch bullet references the threshold trigger (no bare-discretion-only wording) | release-blocking | E-4 (discretion wording present, trigger absent) | M2 |
| AC-004 | REQ-005 | Disabled (0/absent) fallback documented in §4 | release-blocking | E-1 (0 hits) | M1 |
| AC-005 | REQ-004 | Docs carry NO duplicated numeric value in ANY form (YAML spelling `lead_push_threshold: 20` OR word form `초기값 20`); config retains the value | release-blocking | E-3 positive control (config=20) + E-9 (word form present at AGENTS.local.md:206 — the mutant D2 closed) | M1+M3 |
| AC-006 | REQ-004 | `AGENTS.local.md` §4.1 item 6 cross-references the green-conditional (predicate: `초록 조건부|green-conditional`) | planned | E-11 probe (0 hits, exit 1; bare `초록\|green` is NOT the predicate — line 177 matches it, E-5) | M3 |
| AC-007 | — | FREEZE: lanes-never-push wording in §4 ¶1 unchanged | regression-guard | E-6 (wording present pre-edit) | re-run post-edit |
| AC-008 | — | FREEZE (token-presence): landing-verification tokens survive in BOTH §4 and §7 after M2 edits the shared bullet | regression-guard | E-10 (token count = 2 pre-edit) | re-run post-edit |

AC-007/AC-008 are regression-guards (historical-state claims, not flippable gates) — they are
verified by direct re-execution before and after the edit, never by inference.

## §D.1 RED-now Evidence Ledger

All measurements this run, worktree `.moai/worktrees/t1346`, tree `51abf337a` (2026-09-29).

| Id | Command (single invocation) | Verbatim stdout | Exit |
|---|---|---|---|
| E-1 | `grep -c "lead_push_threshold" .claude/rules/local/gitflow-lane-protocol.md` | `0` | 1 |
| E-2 | `grep -n "초록\|green\|CI" .claude/rules/local/gitflow-lane-protocol.md` | raw output: ledger entry E-2 below | 0 |
| E-3 | `grep -n "lead_push\|manual:" .moai/config/sections/git-strategy.yaml` | `8:    manual:` / `26:        lead_push_threshold: 20` | 0 |
| E-4 | `grep -n "배치를 닫을 시점" .claude/rules/local/gitflow-lane-protocol.md` | raw output: ledger entry E-4 below | 0 |
| E-5 | **iter2 — retracts the iter1 cell, which recorded a command that was not executed** | raw output: ledger entry E-5 below | 0 |
| E-5c | command: `grep -n "초록\|green" AGENTS.local.md` (fresh, iter2) | see E-5 ledger entry — line 177 matches, exit 0 | 0 |
| E-6 | §4 ¶1 + §4 landing paragraph + §7 last bullet read (lines 77-102) | lanes-never-push + landing-verification wording present | 0 |
| E-7 | `grep -rn "LeadPushThreshold" --include="*.go" .` | raw output: ledger entry E-7 below (3 lines — zero runtime consumers) | 0 |
| E-8 | `git rev-list --count origin/develop..develop` | `12` (origin/develop `6bcc7e4ac` → develop `51abf337a`) | 0 |
| E-9 | `grep -n "초기값 20" AGENTS.local.md` | raw output: ledger entry E-9 below (line 206 — the D2 word-restatement) | 0 |
| E-10 | `grep -nE "git fetch" .claude/rules/local/gitflow-lane-protocol.md \| grep -c "origin/develop"` | `2` (§4 landing ¶ + §7 landing step) | 0 |
| E-11 | `grep -nE "초록 조건부\|green-conditional" AGENTS.local.md` | no output | 1 |

All measurements this run (iter2 re-measurement batch), tree `51abf337a`, worktree
`.moai/worktrees/t1346`. The iter1 E-5 cell is RETRACTED: the command it named was not
executed before its output was recorded; E-5c is the fresh execution with the observed
result (the auditor's re-run — line 177 "초록불로" — is confirmed).

### Evidence ledger (long verbatim outputs)

```text
E-2 | command: grep -n "초록\|green\|CI" .claude/rules/local/gitflow-lane-protocol.md | exit 0
2:description: "git-flow lane protocol (repo-local) — card worktrees branch from develop, lanes merge into a single develop integration worktree, origin/develop is the CI verdict surface, rc builds are cut from develop, release/vX.Y.Z is the only path to main"
44:- 통합 브랜치는 `develop` 이며 push 대상은 `origin/develop` 이다(§4의 CI 판정 면).
83:**원격 CI(`origin/develop`)가 통합 판정의 주체다.** 로컬 통과는 조기 신호일 뿐이다 — 깨끗한 환경도, darwin/windows 매트릭스도 아니다.
106:[HARD] 자기 변경이 영향 줄 수 있는 테스트만 돌리고, push 후 `origin/develop` CI가 전체 스위트를 돌리게 한다.

E-4 | command: grep -n "배치를 닫을 시점" .claude/rules/local/gitflow-lane-protocol.md | exit 0
102:- **develop push는 리더의 일괄 소관이다(2026-09-02).** 레인 완료 보고에서 카드 id와 로컬 병합 SHA를 모은다 → 배치를 닫을 시점을 리더가 판단한다 → `git push origin develop`을 **한 번** 실행한다 → `git fetch`와 `git rev-parse origin/develop`으로 원격 착지를 검증한다 → 그 뒤에야 카드 done과 워크트리 폐기 승인을 낸다.

E-5 | command: grep -n "초록\|green" AGENTS.local.md | exit 0
177:카드별로 각각 검증해 머지했는데 **합쳐진 상태는 아무도 보지 않는** 구멍을 막는다. 2026-08-15에 PR 12개가 각각 초록불로 main에 들어갔고, 합류 후에야 `moai update`가 로컬 전용 파일을 지운다는 사실이 드러났다.

E-7 | command: grep -rn "LeadPushThreshold" --include="*.go" . | exit 0
internal/config/types.go:133:	// LeadPushThreshold is the manual-mode batch-push trigger (SPEC-MAIN-COMMIT-BAN-001
internal/config/types.go:140:	LeadPushThreshold int `yaml:"lead_push_threshold"` // manual mode only; 0 = disabled
internal/config/defaults.go:896:			LeadPushThreshold: 0,

E-9 | command: grep -n "초기값 20" AGENTS.local.md | exit 0
206:6. **로컬 `main`은 commit-dead다 (SPEC-MAIN-COMMIT-BAN-001, 카드 t1337).** 어느 세션도 primary 체크아웃의 `main` 안에서 커밋하지 않는다 — `git commit` / `git revert` / `git cherry-pick`은 BranchGuard(`workflow.branch_guard.deny_commits_on: [main]`)가 거부하고, 커밋은 `develop`에서 분기한 카드 워크트리에서만 만든다. main의 잔여물을 처분하는 절차(운영자 터미널 전용)는 `.moai/docs/gitflow-integration-chain.md`가 소유한다. 리더의 develop push는 배치 트리거로 닫는다 — `git rev-list --count origin/develop..develop`이 `git_strategy.manual.lead_push_threshold`(초기값 20)에 닿으면 배치를 닫는다. 현재값의 원천은 설정 파일이고 이 규율은 키와 계수 명령을 명명할 뿐이다.
```

E-8 is context, not an AC: the current unpushed backlog is below threshold, so the trigger
would not fire today — no operational interference with this card's landing.

## §D.2 Given-When-Then Scenarios

- **AC-001** — **Given** the lane protocol at §4, **When** a reader looks for the batch-close trigger, **Then** the text names `git_strategy.manual.lead_push_threshold` and `git rev-list --count origin/develop..develop` (grep ≥1 hit for each), and no numeric value appears in the prose.
- **AC-002** — **Given** a batch candidate whose last push left `origin/develop` CI red, **When** the lead reads §4, **Then** the doc states the next push is withheld until the red is repaired (grep ≥1 hit for the hold wording), and states the window re-measurement as the per-card pre-push gate.
- **AC-003** — **Given** §7's batch bullet, **When** read post-edit, **Then** the bullet references the threshold trigger / §4 instead of asserting bare discretion alone (the phrase "배치를 닫을 시점을 리더가 판단한다" no longer stands without a trigger reference beside it).
- **AC-004** — **Given** `lead_push_threshold: 0` (the shipped template default), **When** the lead reads §4, **Then** the doc states the trigger is disabled and batch timing falls back to the lead's judgment, without erroring.
- **AC-005** — **Given** the edited docs, **When** `grep -nE "lead_push_threshold: 20|초기값 20" .claude/rules/local/gitflow-lane-protocol.md AGENTS.local.md` runs, **Then** exit is 1 (no numeric restatement in ANY form — the word form at line 206 is deleted by M3) while E-3's config grep still returns the value — key named, number delegated.
- **AC-006** — **Given** `AGENTS.local.md` §4.1 item 6, **When** `grep -nE "초록 조건부|green-conditional" AGENTS.local.md` runs post-edit, **Then** ≥1 hit naming the lane protocol's green-conditional. The bare `초록\|green` pattern is NOT the predicate — line 177's unrelated "초록불로" matches it (E-5, exit 0).
- **AC-007** — **Given** the pre-edit §4 ¶1 (E-6), **When** the post-edit file is diffed, **Then** the lanes-never-push sentence is byte-unchanged.
- **AC-008** — **Given** the pre-edit landing-verification tokens (E-10: count 2), **When** the post-edit file is checked with `grep -cE "git fetch.*origin/develop" .claude/rules/local/gitflow-lane-protocol.md`, **Then** the count is ≥2 — both the §4 landing paragraph and the §7 landing step retain the landing-verification tokens (token-presence form; a byte-freeze on §7 line 102 is impossible because M2 edits that same bullet).

## §D.3 Edge Cases

- Threshold reached WHILE a window is open: the trigger fires after the window closes —
  push never happens mid-window (the window discipline is untouched; docs wording must not
  imply a push inside `acquire`/`release`).
- Threshold reached but last-push CI red: REQ-002 wins — the batch waits. The docs state the
  precedence (green-conditional overrides the count trigger).
- Threshold key present with a huge value: same as disabled-in-practice; docs need no
  special wording beyond the 0-disabled fallback.
- Weekend/no-CI: last-push CI "not yet established" is not red — REQ-002 holds only on red;
  the docs must not require a CI verdict that does not exist yet (first push of a fresh
  develop, for instance).

## §D.4 Quality Gates

- No Go code changes → no package tests owed; verification is doc-grep + diff-based.
- Lint: no lintable code surface touched (`.md` only).
- Plan-audit gate on the plan artifacts (this set) before run-phase entry.

## §D.5 Definition of Done

- AC-001..006 flipped green by M1-M3 with the E-ledger commands re-run, outputs recorded in
  `.moai/reports/t1346/`.
- AC-007/008 freeze re-executions recorded (pre/post diff evidence).
- Out-of-scope declarations verified untouched: no verb, no goal wiring, no Go consumer
  (E-7 re-run post-edit must be unchanged), no template mirror.
