# SPEC-INIT-QUIET-WIZARD-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-11

Plan 단계 산출물 작성 완료 (manager-spec, 카드 t583, Tier L): spec.md v0.1.4 (REQ 16개, GEARS 5패턴), plan.md, acceptance.md (인수 기준 17건 — AC-IQW-001~016, 007 은 a/b 분할), design.md, research.md, spec-compact.md. 워크트리 `.claude/worktrees/t583`, 브랜치 `WT-init-quiet-wizard`, HEAD `120436f58`.

plan 감사 1회차(`.moai/reports/t583/plan-audit.md`): FAIL 0.79(Tier L 기준 0.85 미달, must-pass 전부 통과). v0.1.2 에서 D1~D8 을 반영했다 — D1 은 리드 결정(셸 설정 단계 시접 + 스파이 실행 관측)으로 바꿨다. v0.1.3 에서 리드 조건 두 가지를 반영했다: 시접을 바꾸는 테스트의 병렬 금지와 `t.Cleanup` 원복을 AC-IQW-016 으로 올렸고, AC-IQW-003 과 완료 정의의 카드 범위 diff 를 리터럴 핀에서 흡수한 로컬 `develop` 과의 merge-base 기준으로 바꿨다.

plan 감사 2회차(`.moai/reports/t583/plan-audit-iter2.md`): FAIL 0.86(Tier L 기준 0.85 는 넘었으나 blocking 결함 D9·D10 이 남음, must-pass 전부 통과). v0.1.4 에서 D9~D16 을 반영했다 — D9 는 AC-IQW-005 대상 목록의 검증 시점 스윕, D10 은 AC-IQW-004 본문 보존 관측.

plan 감사 3회차(`.moai/reports/t583/plan-audit-iter3.md`, 마지막): FAIL 0.87(점수 추이 0.79 → 0.86 → 0.87, must-pass 전부 통과). D10~D16 해소, D9 부분 해소, D1~D8 회귀 없음. 남은 blocking 결함은 D17 하나다 — AC-IQW-005 의 스윕이 추가 파일 단위라, 테스트 단위로 걸리는 REQ-IQW-012 보다 좁다. 선택 결함은 D18~D20 이다.

plan 감사 처분 — PASS-with-debt (리드·운영자 결정, 2026-09-11): D17 을 부채로 안고 run 단계로 넘긴다. 조건은 둘이다. (1) run 위임문에 "새 init 실행 테스트는 새 파일에만 추가한다" 는 제약을 싣는다. (2) M6 마감 때 §E.2 에, 기존 `internal/cli` 테스트 파일에 추가된 `^+func Test` 줄 수가 0 임을 보이는 명령과 그 출력을 그대로 기록한다. D18~D20 을 run 단계에서 채택하면 그 기록은 `.moai/reports/t583/verdict.md` 에 남긴다. 재감사는 하지 않는다. Implementation Kickoff Approval 은 운영자 결정 대기 중이며, run 단계 진입은 그 승인을 기다린다.

Gap — lint 판정 빌드 좌표: 1·2회차와 이 문서의 `moai spec lint` 종료 코드 0 은 설치본 `/Users/goos/go/bin/moai`(`v3.2.0-rc.7`, `…-ged71054d3-dirty`)가 낸 것이다. 이 빌드는 워크트리 HEAD `120436f58` 의 조상이 아니며 그 역도 아니다(plan 감사 2회차 관측, 2026-09-11 재측정: `git merge-base --is-ancestor ed71054d3 HEAD` 종료 코드 1, `git merge-base --is-ancestor HEAD ed71054d3` 종료 코드 1). 따라서 lint 판정은 이 트리로 만든 빌드에 귀속되지 않는다.

확인 항목 해소 (2026-09-11): 기존 테스트 소급 범위는 리드 결정 (a) 로 정해졌다 — 코드 점검표는 이 SPEC 이 새로 쓰거나 본문을 다시 쓰는 init 실행 테스트에만 적용하고, init 실행 테스트를 돌리는 모든 run 슬롯은 실제 홈 지문 절차(spec.md §4.3, REQ-IQW-014·016, AC-IQW-015)를 따른다. plan.md 에 남은 확인 필요 표식 없음.

run 단계 안내: §E.2 는 슬롯 표(슬롯마다 한 행, 행 머리는 슬롯 번호)로 시작하고, 각 행에 지문 명령·두 지문·종료 코드를 남긴다. 형식과 판정식은 acceptance.md AC-IQW-015 에 있다.

Pre-spawn divergence origin/develop...HEAD = 17 1 at 2026-09-11 (first plan-time measurement); the 17 commits (t657 web, t545 settings form, t564 AC ids) touch none of internal/cli/wizard, internal/cli/init.go, internal/cli/update_wizard.go, internal/core/project, internal/shell (git diff --stat empty); absorb deferred to the integration window.

plan 작성 중 재측정 (같은 워크트리, 2026-09-11):

```
$ git fetch origin develop -q; git rev-list --count --left-right origin/develop...HEAD
17	1
$ git diff --stat HEAD...origin/develop -- internal/cli/wizard internal/cli/init.go internal/cli/update_wizard.go internal/core/project internal/shell | tail -2; echo "diffstat-exit=$?"
diffstat-exit=0
```

(diff --stat 출력 없음, 종료 코드 0.)

v0.1.3 직전 재측정 (같은 워크트리, HEAD `120436f58`, 2026-09-11): `origin/develop...HEAD` = `118 1`. SPEC 대상 경로 가운데 develop 이 바꾼 파일은 `internal/cli/update_wizard.go` 하나이며, 커밋은 `c4990eea7`(t587, `applyWizardConfig` 의 system.yaml 오류 처리, `update_wizard.go:312-340` 에 +13/-4)다. research.md·design.md·plan.md·spec.md 가 인용하는 `update_wizard.go` 줄(`:64`, `:133`, `:307-310`)은 `git show origin/develop:internal/cli/update_wizard.go` 에서도 같은 줄 번호에 같은 내용이다. 흡수는 여전히 통합 창으로 미룬다. AC-IQW-003 과 완료 정의의 위저드 범위 확인은 이제 흡수한 `develop` 과의 merge-base 부터 재므로, t587 을 흡수해도 그 때문에 빨개지지 않는다.

```
$ git fetch origin develop -q 2>&1; git rev-list --count --left-right origin/develop...HEAD
118	1
$ git diff --stat HEAD...origin/develop -- internal/cli/wizard internal/cli/init.go internal/cli/update_wizard.go internal/core/project internal/shell
 internal/cli/update_wizard.go | 17 +++++++++++++----
 1 file changed, 13 insertions(+), 4 deletions(-)
$ git log --oneline HEAD..origin/develop -- internal/cli/update_wizard.go
c4990eea7 fix(update): repair recovery hint, git-mode render, archive order, wizard errors (t587)
$ git merge-base develop HEAD
93182d137159c4facbf87c66dea3fd69160a6b8b
$ git diff --name-only develop...HEAD | wc -l
       0
$ git diff --quiet 120436f58 develop -- internal/cli/update_wizard.go; echo "pinned-vs-develop-exit=$?"
pinned-vs-develop-exit=1
$ git diff --quiet develop...HEAD -- internal/cli/update_wizard.go; echo "mergebase-exit=$?"
mergebase-exit=0
```

(마지막 두 줄: 옛 리터럴 핀 형식은 로컬 develop 에 든 t587 때문에 이미 종료 코드 1 이다. merge-base 형식의 0 은 카드 커밋이 아직 없어 대조군이 0 인 상태의 값이므로 "측정 불가" 이며, 통과 근거가 아니다.)

SPEC ID 자기 검사:

```
$ ID="SPEC-INIT-QUIET-WIZARD-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL; ls .moai/specs | grep -c QUIET
PASS
0
```

## §E.2 Run-phase Evidence

Slot table (AC-IQW-015). Every `go test` call against `./internal/cli` or `./internal/core/project/...` is declared here before it runs. Slots are granted by the lead and executed by the lane orchestrator (card t583); the fingerprint command is acceptance.md AC-IQW-015's reference form (`stat -f %m`, macOS), identical before and after except the file name. Tree for M1 slots: HEAD `1b3666cc1` plus the uncommitted M1 change (initializer.go seam, three new test files).

| Slot | Declared | Command | go test exit | before.out lines | .err bytes (before/after) | Result |
|---|---|---|---|---|---|---|
| SLOT-1 | 2026-09-11 | `go test ./internal/core/project/... -run '^(TestInitializer_ShellConfigSeamGate\|TestConfigureShellEnvFn_DefaultIsProductionFunc)$' -count=1 -v` | 0 (`--- PASS:` top-level 2, both subtests PASS, `no tests to run` 0; `.moai/state/verify/t583/ac004-gate.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-2 | 2026-09-11 | same selector, `-count=2 -v` | 0 (`--- PASS:` top-level 4, `--- FAIL` 0, `no tests to run` 0; `.moai/state/verify/t583/ac016-project.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-3 | 2026-09-11 | `go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=1 -timeout 600s -v` | 1 — TOOL_FAILURE (tdd-result-contract): `[build failed]`, `fileSHA256 redeclared in this block` (`internal/cli/update_preserve_my_harness_test.go:20:6` vs new `internal/cli/init_home_guard_test.go:130:6`); no test ran (`.moai/state/verify/t583/ac004-primary.txt`, 25 lines) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-4 | 2026-09-11 | `go test ./internal/cli -run '^(TestHomeGuard_RejectsPathInsideRealHome\|TestHomeGuard_AcceptsTempDir)$' -count=1 -timeout 600s -v` | not run — same package build failure as SLOT-3; slot returned to the lead, re-declared after the fix | — | — | — |
| SLOT-5 | 2026-09-11 | `go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=2 -timeout 600s -v` | not run — same reason as SLOT-4 | — | — | — |
| SLOT-6 | 2026-09-11 | SLOT-3 rerun after the rename: `go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=1 -timeout 600s -v` | 0 (`--- PASS:` 1, `--- FAIL` 0, `no tests to run` 0, `ok … 1.629s`; `.moai/state/verify/t583/ac004-primary.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-7 | 2026-09-11 | SLOT-4 command: `go test ./internal/cli -run '^(TestHomeGuard_RejectsPathInsideRealHome\|TestHomeGuard_AcceptsTempDir)$' -count=1 -timeout 600s -v` | 0 (`--- PASS:` 2, `--- FAIL` 0, `no tests to run` 0, `ok … 0.869s`; `.moai/state/verify/t583/ac005.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-8 | 2026-09-11 | SLOT-5 command: `go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=2 -timeout 600s -v` | 0 (`--- PASS:` 2, `--- FAIL` 0, `no tests to run` 0, `ok … 2.259s`; `.moai/state/verify/t583/ac016-cli.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |

SLOT-6 is the first slot that runs an init execution test through the new home-safety helper; its home-diff-exit=0 is the first runtime confirmation that the helper plus seam spy leave the real home untouched on this machine. Behavioral RED for M1 (mutants A, B, C1, C2, D) is still owed and needs further internal/cli and internal/core/project slots.

Fingerprint form deviation (recorded, not hidden). The worktree-isolation guard refused AC-IQW-015's reference fingerprint command as a single invocation ("too complex to verify that it stays inside the worktree"), so each fingerprint is taken as three plain commands, byte-identical before and after except `before`/`after` in the file names: (A) `shasum -a 256 "$HOME/.claude/settings.json" "$HOME/.zshenv" "$HOME/.zshrc" "$HOME/.zprofile" "$HOME/.profile" "$HOME/.bashrc" "$HOME/.bash_profile" > home-SLOT-<n>-<phase>-sha.out 2> …-sha.err`; (B) `stat -f '%N mtime=%m' "$HOME/.zshenv" "$HOME/.zshrc" "$HOME/.zprofile" "$HOME/.profile" "$HOME/.bashrc" "$HOME/.bash_profile" > …-mtime.out 2> …-mtime.err`; (C) `ls -d "$HOME/.claude/hooks/moai" > …-hooks.out 2> …-hooks.err`. The 8 AC items are covered: settings.json sha256 (A), hooks/moai presence (C), six rc files sha256 (A) and mtime (B). All seven files existed and `hooks/moai` was absent at SLOT-1 (checked with `ls -ld` first), so (C) exits 1 by design with a fixed "No such file" message on stderr; for (C) the comparison is `diff` of the two `.err` files instead of the zero-byte rule. Expected line counts in this form: sha 7, mtime 6. The per-slot `home-diff-exit` above is 0 only when all three diffs are 0.

VET (not a slot, no test execution): `go vet ./internal/core/project/...` → `vet-exit=0` (`.moai/state/verify/t583/m1-vet-project.txt`).

SLOT-3 repair (2026-09-11): the new helper `fileSHA256` was renamed to `homeGuardFileSHA256` in `internal/cli/init_home_guard_test.go` only. manager-develop swept all 11 top-level identifiers of the two new internal/cli test files: each has exactly one declaration hit and zero word hits outside the new files (the sweep uses `([^A-Za-z0-9_]|$)` because `\b` is not a word boundary in `git grep -E` — a control on the old name found `update_preserve_my_harness_test.go:20` only with the corrected form). Compile check allowed without a slot by the lead: `go vet ./internal/cli/` → `vet-exit=0`, output 0 bytes (`.moai/state/verify/t583/m1-vet-cli.txt`). SLOT-4/5 and the SLOT-3 rerun are re-declared when the next internal/cli slot is granted.

M1 non-slot observation — AC-IQW-004 body preservation (lane, 2026-09-11, tree above): `git show 120436f58:internal/core/project/initializer.go > .moai/state/verify/t583/ac004-initializer-base.go` → `show-exit=0`; base and head `ConfigOptions` extractions both 4 lines; `diff ac004-body-base.txt ac004-body-head.txt` → `body-diff-exit=0`.

M1 commit: status draft → in-progress landed with the M1 commit (manager-develop, 2026-09-12); mutant RED slots (A, B, revert, C1, C2, D) and mutant E are still owed against this committed tree.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Recorded by the lane orchestrator (card t583) before the first run-phase `Agent()` spawn, after Implementation Kickoff Approval (operator, relayed by the lead, 2026-09-11) and the develop absorb (`2723447be`, HEAD^2 `4c99d973e`).

Input parameters:
- tier: L (spec.md; 6 milestones M1-M6)
- scope: about 15 files across `internal/cli/wizard`, `internal/cli`, `internal/core/project` (plan.md §F)
- domain count: 1 (Go CLI init path and its tests); no template, docs, or hook work
- file language mix: Go source and Go tests
- concurrency benefit: LOW — coding-heavy, milestones depend in order (M1 seam lands before M2 tests), and every `internal/cli` / `internal/core/project` compile or test run is a lead-granted slot with a real-home fingerprint (AC-IQW-015)
- Agent Teams: not requested

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | semantic multi-file change, not a trivial edit |
| serial | **yes** | one `manager-develop` (cycle_type=tdd) per milestone, lane runs slot-gated verification between steps |
| fanout | no | single domain, coding-heavy; parallel writers would race on `init.go` and the wizard files |
| sweep | no | not a uniform mechanical transform |
| manager-lead | no | entry predicate needs cross-domain fan-out; this card is one domain with serial dependencies |

Decision: serial

Justification: the milestones are ordered by a hard dependency (REQ-IQW-011 seam before any init execution test), and verification cannot run inside the writer because test slots are granted by the lead one at a time. A single sequential writer with lane-run slots keeps one writer on the tree and keeps every `go test` call declared and fingerprinted.

Boundary case: none of the numeric thresholds is at ±1; the deciding factor is the slot-gated verification, not file count.
