# SPEC-ERA-H3-NARROWING-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- tier: S
- artifacts: spec.md + plan.md (+ progress.md) — AC는 spec.md §3에 인라인
- baseline_tree: f72c0bf0f (worktree `.claude/worktrees/t382`, branch `WT-era-plan-phase`) — re-measured after the plan artifacts landed, because this SPEC is itself an instance of the defect (V3R5 23 -> 24, grandfathered 285 -> 286)
- superseded_baseline_tree: 9328a5242 (pre-artifact; background only)
- measurement_tool: `./bin/moai` (this tree, `make build` rc=0). No PATH binary.
- decision_evidence: `.moai/reports/t382/red-evidence.md` (R1-R4 — command + verbatim stdout + exit code + tree SHA per verification-completeness.md 2.1). Probe bodies: `red_probe.py` (R1), `drift_probe.py` (R4). Per-item attribution: R1-R3 at `1f10f5e8d`, R4 at `f967089ba`
- background_evidence: `.moai/reports/t382/measurements-9328a5242.md` (M1-M13), `v3r5-population.txt`, `drift-before-9328a5242.txt` — no verbatim stdout or exit codes, so NOT cited as decision basis
- red_state_at_plan_close: R1 exit 1 (`misclassified: 23` of 24 swept); R4 exit 1 (`unearned exemption: 22` of 23 era-exempt rows)
- plan_audit: iter 1/1 (Tier S ceiling), verdict PASS-WITH-DEBT, score 0.825 vs Tier S baseline 0.75. Report `.moai/reports/t382/plan-audit-iter1.md` (auditor-owned, not modified by this card)
- audit_debt_repaid: D1-D4 (blocking) + D5-D10 (optional) all closed at v0.4.0. Auditor findings independently re-verified before adoption: D1 via `drift_probe.py`, D2 via `d2_check.py`, D7 via `d7_check.py`, D6/D9/D10 via git and grep

## §F Phase 4 Mode Selection

Decision: **serial** (단일 sub-agent 순차 spawn)

**입력** — tier S / scope 3파일(`era.go`, `era_test.go`, `lifecycle-sync-gate.md`) / domain 1(Go 분류기 + 그 규칙 문서) / concurrency benefit LOW.

| 모드 | 선택 | 사유 |
|---|---|---|
| `direct` | 미선택 | 오타 수정이 아니다. 분류 술어 변경 + 테스트 신설이고 양방향 판정이 걸려 있다 |
| `serial` | **선택** | 아래 |
| `fanout` | 미선택 | 3파일 단일 도메인이라 나눌 축이 없다. research-heavy 도 아니다 |
| `sweep` | 미선택 | ~30파일 문턱에 두 자릿수 미달, 단일 균일 변환도 아니다 |

**정당화.** 편집이 3파일뿐이고 서로 의존한다 — `era.go` 가 도입하는 헬퍼의 이름과 술어가 곧 `era_test.go` 의 판정 대상이고, `lifecycle-sync-gate.md` 는 그 결과를 서술한다. 나눌 것이 없는 크기에 fan-out 을 쓰면 조율 비용만 남는다.

**경계 사례 없음.** scope 3파일은 Tier S 구간(<5)의 한가운데이고, plan 감사가 Tier S 로 확정했다. §C.3 같은 미결 결정이 이 카드에는 없다.

## §E.2 Run-phase Evidence

측정 트리 `c0cdb2fd3` (worktree `.claude/worktrees/t382`, branch `WT-era-plan-phase`).
도구는 이 트리에서 `make build` 한 `./bin/moai` 와 수정 전 사본 `./bin/moai-pre-t382`.
`origin/develop` = `297a21ea7`, 발산 `5 6` — **흡수하지 않았다**. 아래 수치는 전부 흡수 이전
트리의 것이며, 통합 창에서 병합 트리 재측정이 남아 있다.

**착수 시점 기준선 재측정** (spec.md §3.2 오른쪽 열이 그대로 재현됐다):
Total 715 / Grandfathered 286 / V3R5 24 / V3R6 429(뺄셈) / `EraUnclassified` 0.
R1 exit 1 (`misclassified: 23`), R4 exit 1 (`unearned exemption: 22`).

| AC | 판정 | 명령 | 관측 |
|---|---|---|---|
| AC-EH3-001 | PASS | `go test -run TestClassifyEra ./internal/spec/` | RED `era = V3R5 ... want V3R6` → GREEN `ok`. 코퍼스 R1 exit 1 → 0 |
| AC-EH3-002 | PASS | 같음 + 「헬퍼 날짜만으로 축소」 뮤테이션 | 뮤턴트에서 이 서브테스트만 실패. 같은 뮤턴트에 대해 코퍼스 R1 은 exit **0** (`misclassified: 0`) — R1 이 phase 경로에 공허하다는 spec.md §3.3 의 주장이 이 트리에서 독립 재현됐다 |
| AC-EH3-003 | PASS | 같음 + 「H-3 무조건 스킵」 뮤테이션 | 뮤턴트에서 `era = unclassified, want V3R5` 로 실패 |
| AC-EH3-004 | PASS | 같음 + 같은 뮤테이션 | 뮤턴트에서 `H-6 (no heuristic matched)` 로 실패. 코퍼스 `EraUnclassified` 0 → 0 |
| AC-EH3-005 | PASS | `./bin/moai spec audit --json` 전후 대조 (`attribution_probe.py`) | 총계 286→263 / 24→1 / V3R6 429→452 / V2.x 144·V3R2-R4 118 불변 / `EraUnclassified` 0→0. 시대 변경 **23건**이 모집단−INIT-WIZARD 와 원소 단위 일치(`diff` 무출력). 23건 전부 근거 보유. 표본 7건 `created` ↔ git 최초 커밋 날짜 대조 |
| AC-EH3-006 | PASS | `inject_probe.sh` (임시 사본, `.moai/specs/` 무손상) | 수정 전 rc **0** + `warning` / `advisory: true` / `[grandfathered era — downgraded to warning]`; 수정 후 rc **1** + `error` / advisory 없음 |
| AC-EH3-007 | **FAIL** | `spec drift --no-cache` 전후 + `drift_probe.py` | `unearned exemption: 22 → 0`, exit 1 → 0, `era-exempt` 행 250 → 228 (명제 1 충족). 그러나 명제 3 위반 — 새 `DRIFT` 4행 중 1행이 **오탐**이다. 판정: `.moai/reports/t382/m3-drift-row-judgment.md` |
| AC-EH3-008 | PASS | `go test -run TestClassifyEra_NoV3R5WhileModernSignal ./internal/spec/` | 유예절 제거 뮤턴트에서 실패 관측 후 되돌려 재통과. 스윕 20조합(빈 스윕 가드 내장) |

**AC-EH3-007 이 실패인 이유와, 그 원인이 이 카드의 것이 아닌 이유.** 새로 `DRIFT` 로 뜬 4행 중
3행(`SPEC-HARNESS-LEARNING-EVO-002` · `SPEC-KANBAN-TODO-CLI-001` · `SPEC-UPDATE-VERSION-FLAG-001`)은
진짜 불일치다 — `main` 에 close 커밋이 아예 없거나 frontmatter 가 착지한 작업보다 뒤처져 있다.
4행째 `SPEC-V3R6-SESSION-HANDOFF-AUTO-001` 은 **오탐**이다: 실제 close 는 `e979a4d13
chore(SPEC group C): Mx-phase close` 로 존재하나 그 subject 가 **결합 범위**(prefix only)라
검출기가 귀속하지 못하고, `--grep` 이 커밋 **본문**까지 매치해 다른 SPEC 의 plan 커밋
`7cffb9717` 을 최신으로 집는다. 두 원인 모두 이 카드 이전부터 있었고
`spec-frontmatter-schema.md` § Close-subject full-ID mandate 가 이미 금지한 형태다 — 면제가
가리고 있었을 뿐 막고 있던 것이 아니다. 그럼에도 AC-EH3-007 명제 3 은 「오탐 1건이면 실패」라고
적혀 있으므로, 기준을 재해석해 통과시키지 않고 FAIL 로 보고한다.

**예상 밖 관측 2건.**

1. `moai spec drift` 는 HEAD SHA 로 결과를 캐시한다. 이 변경이 미커밋이라 HEAD 가 움직이지 않아
   수정 후 첫 실행이 **수정 전 캐시**를 그대로 냈다 — drift 출력이 바이트 동일(`diff` 무출력)인데
   `spec audit` 은 이미 새 분류를 보고하는 상태였다. 그때도 R4 프로브는 exit 0 을 냈다(스윕
   모집단이 V3R5 집합에서 나오는데 그 집합이 이미 1건으로 줄었기 때문). **spec.md §3 과 plan.md
   §C 어디에도 `--no-cache` 가 없다** — 계획된 그대로 실행했으면 이 공허한 초록을 판정으로
   삼았을 것이다. 모든 수치는 `--no-cache` 재측정본이다.
2. 재분류된 23건 중 **11건은 drift 표에서 행 자체가 사라진다**. `main` 에 그 SPEC-ID 를 언급하는
   커밋이 없어 `getGitImpliedStatus` 가 `no git history found` 를 내고 record 가 drop 된다.
   면제도 정렬도 아닌 **부재**이며, SPEC 이 예상하지 않은 상태다.

**범위 — 3파일이 아니라 4파일이다.** `git diff --stat c0cdb2fd3 -- internal/ .claude/` 는
`era.go` · `era_test.go` · `lifecycle-sync-gate.md` 에 더해 **`internal/spec/audit_test.go`** 를
낸다. `makeSpecMD` 헬퍼가 모든 픽스처에 `phase: "v3.0.0"` 을 박는데, 그 값이 곧 modern-era
신호라 `TestAudit_EraClassification5Buckets` 의 V3R5 픽스처가 수정 후 V3R6 으로 옮겨가
스위트가 적색이 됐다(`era V3R5 not observed`). 픽스처의 phase 를 비-modern 값으로 덮는
1줄(+주석)로 고쳤다 — 이미 같은 파일이 `fixtures[4]` 에 쓰던 방식 그대로다. spec.md §5 옵션 C 의
「기존 테스트를 하나도 깨지 않는다」는 **측정으로 반증됐다**. REQ-EH3-004 의 실질 금지는
지켜졌다: `lint.go` · `audit.go` · `drift.go` 무수정, `eraDemotableCodes` ·
`applyEraDemotion` · `EraFinal()` · `IsModern()` diff 0건.

**뮤테이션 왕복 기록** — 3종 모두 심고, 실패를 관측하고, 되돌려 재통과했다.

| 뮤테이션 | 깨진 것 | 증거 |
|---|---|---|
| H-3 무조건 스킵 | AC-003 · AC-004 + 기존 H-3 케이스 3건 + 5-bucket 커버리지 | `m2-mutant-a-h3-skip.txt` |
| 유예절 제거 | AC-008 불변식 + AC-001 + AC-002, 코퍼스 R1 exit 1 복귀 | `m2-mutant-b-no-deferral.txt` · `m2-mutant-b-r1-corpus.txt` |
| 헬퍼 날짜만으로 축소 | AC-002 (+ 기존 H-5 phase 케이스). 코퍼스 R1 은 **못 잡음**(exit 0) | `m2-mutant-c-date-only.txt` · `m2-mutant-c-r1-corpus-blind.txt` |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-31
run_commit_sha: pending-backfill-run
run_status: complete-with-one-ac-fail
ac_pass_count: 7
ac_fail_count: 1          # AC-EH3-007 (명제 3 — 오탐 1건, 원인은 선행 결함)
preserve_list_post_run_count: 0   # .moai/reports/t382/** 무수정 (신규 증거만 추가)
l44_pre_commit_fetch: "origin/develop=297a21ea7, 발산 5 6 — 흡수하지 않음(리드 소관)"
l44_post_push_fetch: "n/a — push 하지 않음"
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues, go vet rc 0
cross_platform_build:
  host: rc 0
  windows_amd64: rc 0
coverage_internal_spec: 89.6%
total_run_phase_files: 4   # era.go, era_test.go, audit_test.go(픽스처 연쇄), lifecycle-sync-gate.md
m1_to_mN_commit_strategy: single-commit   # Tier S, 4파일
```


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
