auditor-model: claude-haiku-5-5

# Plan audit: SPEC-CAPACITY-RECLAIM-001 (factory card t1631)

## Raw audit_multi return

Saved first, verbatim, as returned by `mcp__moai__audit_multi`. Call parameters: `project_root` = `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1631`, `target` = `.moai/specs/SPEC-CAPACITY-RECLAIM-001`, `focus` = plan-phase adversarial audit, no `gates` argument. Every raw line carries the `> ` prefix.

> {"per_backend_verdicts":[{"backend":"codex","gate":"required","verdict":"fail","summary":"Verdict: fail\n\n- [P1] registry 등록만으로 실제 usage-log를 회전시키거나 소유권을 이전할 수 없다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:11\n  기존 rotator는 `.moai/logs/<name>`을 검사하지만 대상은 `.moai/harness/usage-log.jsonl`이다. 기존 retention-prune 호출도 남는다.\n  신뢰도: 높음. 권고: 대상 경로와 기존 prune의 종료 또는 직렬화까지 전환 계획에 명시해.\n\n- [P1] 파괴적 실행 경로의 unpushed/live-session 보호가 검증되지 않는다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:130\n  AC-013은 report 모드만 실행하고 live-session fixture가 없다. AC-015도 predicate를 통과하는 후보만 검사하므로, `--yes`에서 보호를 우회하는 구현을 놓친다.\n  신뢰도: 높음. 권고: owner-confirmed unpushed/live fixture를 실제 삭제 경로에 넣고 보존을 검증해.\n\n- [P2] 회전을 수행하지 않는 구현도 AC-CAP-007을 통과한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:85\n  초과 크기 primary를 그대로 두면 primary 1개, rotation 0개, truncate 없음 조건을 모두 만족한다. 식 평가로 허위 통과를 확인했다.\n  신뢰도: 높음. 권고: 실제 회전 발생과 기존 내용의 보존을 검증해야 해.\n\n- [P2] 정상적인 승인 거절도 완료 불가능한 상태로 만든다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:153\n  모든 AC의 통과를 요구하지만 AC-007/015는 Q4/Q2 승인을 전제한다. Q4의 기본 선택인 기존 owner 유지나 Q2 거절에는 완료 분기가 없다.\n  신뢰도: 높음. 권고: 결정별 적용 AC와 N/A 처리를 정의해.\n\n- [P2] 삭제 결과를 결정하는 keep-3 정책과 실행 명령이 미확정이다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:37\n  현재 release 포함 여부, cap 설정 키, apply 확인 표면을 run으로 미뤘다. AC-010/011도 실행 명령을 명시하지 않는다. 현재 코드는 current 외에 3개를 보존하므로 총 3개 요구사항과 다르다.\n  신뢰도: 높음. 권고: 보존 불변식, 설정값 검증, 확인 경로와 fixture 명령을 계획에서 확정해.\n\n- [P2] 주석만 추가해도 ignore 인수조건을 통과한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:62\n  주석만 있는 fixture에서 grep는 exit 0이었지만 `git check-ignore`는 exit 1이었다. 실제 제외 동작을 검사하지 않는다.\n  신뢰도: 높음. 권고: 대표 경로와 corrupt 변형을 `git check-ignore --no-index`로 검증해.\n\n- [P2] 필드가 없는 JSON도 worktree list 인수조건을 통과한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:138\n  git과 같은 개수의 `{}` 배열이 조건을 만족한다. REQ-CAP-013의 path, branch, size와 table 출력은 검증하지 않는다.\n  신뢰도: 높음. 권고: 경로·브랜치 집합, 크기 필드와 table 출력을 함께 검사해.\n\n- [P2] corrupt archive의 로컬 삭제를 보존 검사가 놓친다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:74\n  REQ-CAP-003도 연결하지만 검사 대상은 두 state 파일뿐이다. archive 세 개를 로컬에서도 삭제하면 tracked-count 검사는 통과하면서 보존 요구사항을 위반한다.\n  신뢰도: 높음. 권고: archive 세 개의 존재와 내용 hash도 M2 전후에 비교해.\n\n- [P2] 용량 측정이 공백 경로를 누락하고 오류를 exit 0으로 숨긴다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:44\n  실제 144 KiB인 공백 경로 fixture에서 `$2`가 경로를 잘랐고, 명령은 du 오류와 빈 출력을 내면서 exit 0을 반환했다.\n  신뢰도: 높음. 권고: 경로 전체를 보존하고 개별 측정 실패를 전파해.\n  감사 요약: MP-1–7과 MP-9 PASS, MP-8 N/A. 명확성 0.50, 완전성 0.75, 테스트 가능성 0.50, 추적성 1.00; 평균 0.6875로 Tier M 문턱 0.80 미달.\n  명령·출력·미관측 범위: [상세 감사 보고서](/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1631/.moai/reports/t1631/adversarial-review-40cc1031c.md).","findings":[{"severity":"P1","title":"registry 등록만으로 실제 usage-log를 회전시키거나 소유권을 이전할 수 없다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:11","body":"registry 등록만으로 실제 usage-log를 회전시키거나 소유권을 이전할 수 없다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:11\n기존 rotator는 `.moai/logs/<name>`을 검사하지만 대상은 `.moai/harness/usage-log.jsonl`이다. 기존 retention-prune 호출도 남는다.\n신뢰도: 높음. 권고: 대상 경로와 기존 prune의 종료 또는 직렬화까지 전환 계획에 명시해.","file":".moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md","line":11,"confidence":0,"recommendation":""},{"severity":"P1","title":"파괴적 실행 경로의 unpushed/live-session 보호가 검증되지 않는다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:130","body":"파괴적 실행 경로의 unpushed/live-session 보호가 검증되지 않는다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:130\nAC-013은 report 모드만 실행하고 live-session fixture가 없다. AC-015도 predicate를 통과하는 후보만 검사하므로, `--yes`에서 보호를 우회하는 구현을 놓친다.\n신뢰도: 높음. 권고: owner-confirmed unpushed/live fixture를 실제 삭제 경로에 넣고 보존을 검증해.","file":".moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md","line":130,"confidence":0,"recommendation":""},{"severity":"P2","title":"회전을 수행하지 않는 구현도 AC-CAP-007을 통과한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:85","body":"회전을 수행하지 않는 구현도 AC-CAP-007을 통과한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:85\n초과 크기 primary를 그대로 두면 primary 1개, rotation 0개, truncate 없음 조건을 모두 만족한다. 식 평가로 허위 통과를 확인했다.\n신뢰도: 높음. 권고: 실제 회전 발생과 기존 내용의 보존을 검증해야 해.","file":".moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md","line":85,"confidence":0,"recommendation":""},{"severity":"P2","title":"정상적인 승인 거절도 완료 불가능한 상태로 만든다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:153","body":"정상적인 승인 거절도 완료 불가능한 상태로 만든다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:153\n모든 AC의 통과를 요구하지만 AC-007/015는 Q4/Q2 승인을 전제한다. Q4의 기본 선택인 기존 owner 유지나 Q2 거절에는 완료 분기가 없다.\n신뢰도: 높음. 권고: 결정별 적용 AC와 N/A 처리를 정의해.","file":".moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md","line":153,"confidence":0,"recommendation":""},{"severity":"P2","title":"삭제 결과를 결정하는 keep-3 정책과 실행 명령이 미확정이다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:37","body":"삭제 결과를 결정하는 keep-3 정책과 실행 명령이 미확정이다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:37\n현재 release 포함 여부, cap 설정 키, apply 확인 표면을 run으로 미뤘다. AC-010/011도 실행 명령을 명시하지 않는다. 현재 코드는 current 외에 3개를 보존하므로 총 3개 요구사항과 다르다.\n신뢰도: 높음. 권고: 보존 불변식, 설정값 검증, 확인 경로와 fixture 명령을 계획에서 확정해.","file":".moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md","line":37,"confidence":0,"recommendation":""},{"severity":"P2","title":"주석만 추가해도 ignore 인수조건을 통과한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:62","body":"주석만 추가해도 ignore 인수조건을 통과한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:62\n주석만 있는 fixture에서 grep는 exit 0이었지만 `git check-ignore`는 exit 1이었다. 실제 제외 동작을 검사하지 않는다.\n신뢰도: 높음. 권고: 대표 경로와 corrupt 변형을 `git check-ignore --no-index`로 검증해.","file":".moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md","line":62,"confidence":0,"recommendation":""},{"severity":"P2","title":"필드가 없는 JSON도 worktree list 인수조건을 통과한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:138","body":"필드가 없는 JSON도 worktree list 인수조건을 통과한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:138\ngit과 같은 개수의 `{}` 배열이 조건을 만족한다. REQ-CAP-013의 path, branch, size와 table 출력은 검증하지 않는다.\n신뢰도: 높음. 권고: 경로·브랜치 집합, 크기 필드와 table 출력을 함께 검사해.","file":".moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md","line":138,"confidence":0,"recommendation":""},{"severity":"P2","title":"corrupt archive의 로컬 삭제를 보존 검사가 놓친다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:74","body":"corrupt archive의 로컬 삭제를 보존 검사가 놓친다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:74\nREQ-CAP-003도 연결하지만 검사 대상은 두 state 파일뿐이다. archive 세 개를 로컬에서도 삭제하면 tracked-count 검사는 통과하면서 보존 요구사항을 위반한다.\n신뢰도: 높음. 권고: archive 세 개의 존재와 내용 hash도 M2 전후에 비교해.","file":".moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md","line":74,"confidence":0,"recommendation":""},{"severity":"P2","title":"용량 측정이 공백 경로를 누락하고 오류를 exit 0으로 숨긴다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:44","body":"용량 측정이 공백 경로를 누락하고 오류를 exit 0으로 숨긴다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:44\n실제 144 KiB인 공백 경로 fixture에서 `$2`가 경로를 잘랐고, 명령은 du 오류와 빈 출력을 내면서 exit 0을 반환했다.\n신뢰도: 높음. 권고: 경로 전체를 보존하고 개별 측정 실패를 전파해.\n감사 요약: MP-1–7과 MP-9 PASS, MP-8 N/A. 명확성 0.50, 완전성 0.75, 테스트 가능성 0.50, 추적성 1.00; 평균 0.6875로 Tier M 문턱 0.80 미달.\n명령·출력·미관측 범위: [상세 감사 보고서](/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1631/.moai/reports/t1631/adversarial-review-40cc1031c.md).","file":".moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md","line":44,"confidence":0,"recommendation":""}],"next_steps":[]}],"overall_verdict":"fail","disagreement_flag":null,"participant_count":1,"residual_risk_note":"required-backend FAIL: codex","fail_open_backends":[],"audit_receipt":"rcpt-6aa817b9d99b9627dac1b670","plan_source":"config","build_commit":"4f8aba061","tree_root":"/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1631"}

## Coordinator correction applied

- Lease deadline used: 2026-10-10T08:10:36Z (the lease was extended). Last clock read in this audit: 07:24:01Z.
- The raw audit_multi return is saved above, first, with every line prefixed `> `.
- codex stayed in the audit as the required backend. The gate set came from `moai verify audit-plan`: claude off, codex required, glm off.
- No SPEC-author reasoning was supplied. The coordinator correction changed only the deadline, the save order, and the backend requirement.

## Verdict

verdict: FAIL
audited_sha: 40cc1031c2b18781377f1bdf2c590cd39f08244e
overall_score: 0.625
must_pass_failed: 1
blocking_count: 14
plan_artifact_hash: b82e700bcbbee75bf6f9cb2b4410ccea1a471f91750cabb1c638f22b984ea014
scope: full
fix_scope: spec.md#REQ-CAP-004; spec.md#REQ-CAP-008; spec.md#REQ-CAP-009; spec.md#REQ-CAP-010; spec.md#REQ-CAP-011; spec.md#REQ-CAP-012; plan.md#M1; plan.md#M2; plan.md#M3; plan.md#M5; acceptance.md#L7; acceptance.md#AC-CAP-002; acceptance.md#AC-CAP-004; acceptance.md#AC-CAP-006; acceptance.md#AC-CAP-007; acceptance.md#AC-CAP-010; acceptance.md#AC-CAP-011; acceptance.md#AC-CAP-013; acceptance.md#AC-CAP-014; acceptance.md#AC-CAP-015; acceptance.md#§D.6
defect_class: D1=design; D2=ac-wording; D3=ac-wording; D4=design; D5=design; D6=ac-wording; D7=ac-wording; D8=traceability; D9=ac-wording; D10=traceability; D11=format; D12=other; D13=design; D14=other
convergence_overall: fail
required_backend: codex fail
required_backend_fail: codex

The verdict is FAIL on four independent grounds, and any one of them is sufficient:

1. Must-pass MP-2 fails on the requirement layer (REQ-CAP-008 and REQ-CAP-009, D11).
2. The required backend codex returned `fail`, and convergence is `fail` (receipt `rcpt-6aa817b9d99b9627dac1b670`). No PASS may be admitted.
3. overall_score 0.625 is below the Tier M threshold of 0.80.
4. Fourteen blocking findings remain (D1–D14).

Agreement and disagreement with codex: codex reached FAIL on the same convergence ground and reported the same nine findings I verified. I differ on two points. I fail MP-2, which codex passed (D11). I score Traceability 0.75, which codex scored 1.00 (D8, D10, and the disposal and field-coverage gaps in D2 and D7). I also add D3–D5, D9, D12, D13, and D14, which codex did not report.

## Evidence-bearing report

### Claim

The plan does not pass as written. The defects concentrate in the verification layer (criteria that pass on the wrong thing) and in the decision-gated outcomes (the Definition of Done has no branch for a legitimate decision). The four open decisions are operator gates and are not decided here.

### Evidence

- E1 `date -u` → `Sat Oct 10 07:10:52 UTC 2026` (start); `Sat Oct 10 07:24:01 UTC 2026` (later read).
- E2 `git rev-parse HEAD` → `40cc1031c2b18781377f1bdf2c590cd39f08244e`; `git status --short` → empty.
- E3 `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1631` → `config_status: ok`, `cross_model_active: true`, `enforced_required: ["codex"]`; backends: claude `off`, codex `required`, glm `off`.
- E4 `mcp__moai__spec_audit` (filter SPEC-CAPACITY-RECLAIM-001) → `total_specs: 1`, `modern_era_clean: 1`, one INFO `EraAutoDetected` (H-5).
- E5 `mcp__moai__spec_drift` → 136,378 characters, saved to a tool-results file. A grep for the SPEC ID matches one line of that file (the INFO entry). A search for hash keys found none.
- E6 `mcp__moai__audit_multi` (project_root = worktree, target = SPEC directory, no `gates` argument) → `overall_verdict: fail`; codex `verdict: fail`; `participant_count: 1`; `audit_receipt: rcpt-6aa817b9d99b9627dac1b670`; `build_commit: 4f8aba061`.
- E7 `grep -n -E 'NEEDS CLARIFICATION' plan.md` → no output, exit 1. research.md is absent (Tier M).
- E8 `grep -c -i 'syscall'` over spec.md, plan.md, acceptance.md, decision-index.md → 0, 0, 0, 0.
- E9 `grep -n -i -E 'gopls|pylsp|rust-analyzer|golangci|eslint|pytest|jest|cargo|ruff|mypy'` over spec.md, plan.md, acceptance.md → no output, exit 1.
- E10 `grep -n -E '^(created_at|updated_at|labels|spec_id):' spec.md` → no output, exit 1.
- E11 `sed -n '1,16p' spec.md` → the 12 canonical fields (id, title, version "0.1.1", status draft, created and updated 2026-10-10, author, priority P0, phase "v3.2.0 target", module, lifecycle spec-anchored, tags string) plus `tier: M` and `related_specs`.
- E12 `grep -n -E '^status:|^plan_status:'` → spec.md:5 `status: draft`; progress.md:3 `status: draft`; progress.md:4 `plan_status: audit-ready`.
- E13 Referenced SPEC statuses: SPEC-MOAI-HYGIENE-001 `status: completed` (spec.md:5), SPEC-WORKTREE-SWEEP-001 `completed`, SPEC-WORKTREE-REAPER-001 `completed`.
- E14 `grep -n -E '^- \*\*REQ-CAP-[0-9]{3}' spec.md` → 13 lines, REQ-CAP-001 through REQ-CAP-013, at lines 56–77.
- E15 `grep -n -E '^\| AC-CAP-' acceptance.md` → 15 rows (lines 15–29). `grep -c -E '^### AC-CAP-[0-9]{3}' acceptance.md` → 15.
- E16 `grep -n -i 'release-blocking' acceptance.md` → line 7: "No criterion here is classified release-blocking, so the four-element cell of `verification-completeness.md` §2.1 is not required in this file."
- E17 `grep -n -E '^### M[0-9]' plan.md` → M1 (line 9), M2 (16), M3 (22), M4 (27), M5 (33), M6 (39). `grep -n -i 'Exit' plan.md` → no output.
- E18 Ordering candidates, `grep -n -i -E 'before|after|first|prior to|pre-change'`: acceptance.md lines 20, 21, 23, 29, 47, 67, 74, 76, 84, 122, 124, 145, 160; plan.md lines 5, 18, 19, 20, 29, 30, 31, 68.
- E19 RED re-observations on this tree: `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl | xargs -0 wc -c | tail -1` → `128526661 total`; `git ls-files -- '*.corrupt*' | wc -l` → 3; `grep -n -E 'usage-log|lessons-inbox|\.corrupt' internal/template/templates/.gitignore` → no output, exit 1; `grep -n -i -E 'owner|unpushed' internal/cli/worktree/sweep.go` → no output, exit 1; `grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'` → no output, exit 1.
- E20 Rotator and registry: `grep -n -E 'logs|harness|usage-log|lessons' internal/hygiene/sinks.go` → line 90 (lessons-inbox owner exception), line 91 (usage-log owner exception, "internal/harness/retention.go (30-day prune to monthly gz)"). `grep -n -E 'logs|harness|Dir' internal/hygiene/rotate.go` → line 26–27 `LogDir` is `<projectRoot>/.moai/logs`; line 102 `path := filepath.Join(r.LogDir, name)`.
- E21 `grep -n -E 'usage-log|retention|Prune|prune' internal/harness/retention.go` → lines 59–64: "Retention archives and cleans up old entries in usage-log.jsonl", and `PruneStaleEntries` is "the retention-prune child verb and the tests call it".
- E22 `git log --oneline -3` → 40cc1031c (audited), 2aab5f797, 695d056f6. `git diff --stat 2aab5f797 40cc1031c2b18781377f1bdf2c590cd39f08244e` → 5 files changed, 432 insertions, 0 deletions, all under SPEC-CAPACITY-RECLAIM-001.
- E23 SPEC-MOAI-HYGIENE-001 spec.md:39 (REQ-HYG-001): bound `1 + audit_log_kept_rotations`, 10 MiB default threshold, kept rotations 1.
- E24 `grep -n -E 'DefaultReleaseKeep|DefaultProfileMaxBytes' internal/config/defaults.go internal/cli/clean_home.go` → defaults.go:451 `DefaultReleaseKeep = 3` (non-current count, lines 447–451); defaults.go:455 `DefaultProfileMaxBytes = 5 * 1024 * 1024 * 1024`; clean_home.go:243 and :552 use them.
- E25 `grep -n -E 'GiB|10,485,760|10 GiB' spec.md plan.md acceptance.md decision-index.md` → acceptance.md:16, :41, :45, :116; spec.md:36, :113. No REQ line states a size target.
- E26 `grep -n -i 'owner'` over spec.md, plan.md, acceptance.md → plan.md:24 says "Add an owner predicate: an unconfirmed owner is PRESERVE", and no line defines how an owner is determined or confirmed. The same holds for "live session" (spec.md:75).
- E27 `grep -n -E '^(Default|Alternate|Operator verdict):|^### Q' decision-index.md` → Q1 Default line 11, Alternate line 12; Q4 Default line 36, Alternate line 37; Q2 (lines 15–21) and Q3 (lines 23–28) have no Default or Alternate line. All four Operator verdict fields are empty (lines 13, 21, 28, 38).
- E28 `ls .moai/reports/t1631/` → `adversarial-review-40cc1031c.md` (11,200 bytes), written by the codex tool.
- E29 plan-artifact hash replay. `wc -c` → acceptance.md 16006, decision-index.md 3346, plan.md 6202, spec.md 18534 (design.md, research.md, tasks.md absent). Framed stream, following `internal/runtime/audit_cache.go:121-141` (`name:len` + NUL, bytes, NUL, in list order) → 44162 bytes. `shasum -a 256` → `b82e700bcbbee75bf6f9cb2b4410ccea1a471f91750cabb1c638f22b984ea014`.
- E30 `moai verify --help` → the verb set includes `audit-plan`. The output is not the legacy signature, so the plan surface is reachable.
- E31 `git merge-base --is-ancestor 4f8aba061 40cc1031c…` → exit 1. `git merge-base --is-ancestor 40cc1031c… 4f8aba061` → exit 1. `git merge-base 4f8aba061 40cc1031c…` → `2aab5f797b75`.

### Baseline-attribution

- Tree measured: HEAD `40cc1031c2b18781377f1bdf2c590cd39f08244e` on branch `WT-capacity-reclaim`, clean at the start of the audit.
- Code measurements (E19, E20, E24) were taken in this run on this worktree. The pin named in acceptance.md:5 is `2aab5f797`. E22 shows the audited commit differs from that pin only by the five SPEC files, so the code observations carry to the audited commit.
- Judging build: the audit_multi `build_commit` is `4f8aba061` (moai-adk v3.2.0-rc.29 per the MCP banner). It is not an ancestor of `40cc1031c` and does not contain it (E31; merge-base `2aab5f797`). Under verification-claim-integrity §2.2 this lag is named here: the codex result and the `audit-plan` config read are attributed to build `4f8aba061` beside tree HEAD `40cc1031c`. The defect findings rest on file reads and greps in this worktree, not on the build's output, so the lag does not change them. It does limit the codex and `audit-plan` results to the build that produced them.

### Gaps

- G1 Refused tools. The awk-based verification verbs (the Group A traceability verb and the Group B CN-4 verb) were refused by the worktree-isolation guard, which could not verify the awk programs. The refusals are recorded here as required by verification-claim-integrity §3.1. Replacements: plain grep and manual reading. The traceability count (13 REQ definitions, 15 AC rows, complete mapping) and the MP-9 ordering judgment are therefore hand measurements, not verb output.
- G2 plan_artifact_hash. No tool output in this session carried a plan-artifact hash: audit_multi, spec_audit, spec_drift, and audit-plan did not, and the drift file has no hash key. The value is derived by replaying the ComputeHash byte stream (E29). The in-tree Go function was not executed, so the value is my replay of its documented framing, not a tool output.
- G3 spec_drift was saved to a file of 136,378 characters. I searched it for hash keys and for the audited SPEC ID, and I did not read it in full. spec_audit was read in full.
- G4 The codex auxiliary report `.moai/reports/t1631/adversarial-review-40cc1031c.md` (11,200 bytes) was written by the codex tool, not by this auditor. It is untracked and is not one of the five SPEC artifacts. I did not read it in full. The lane should decide whether to commit it.
- G5 Single backend. Only codex participated (`participant_count: 1`). claude and glm are off in the gate set, so the MP-9 judgment and the rubric scores have no second opinion.
- G6 Criteria not re-observed on this tree. AC-CAP-007, -010, -011, and -015 are "RED by source", and I did not re-observe them. AC-CAP-003 needs a fixture run, which is not possible at plan phase. D9 (the AC-CAP-002 command) rests on the shell and awk semantics plus codex's fixture measurement; I did not reproduce the fixture.
- G7 Not executed: the Go ComputeHash, the sweep, and the rotator. D1 rests on reading `rotate.go` and `sinks.go`.
- G8 Provenance lag (see Baseline-attribution). The judging build is not an ancestor of the audited HEAD.

### Residual-risk

- MP-2 is a strict reading. If the operator reads "On apply," as an event-driven trigger, MP-2 becomes PASS, with D11 kept as a format defect. The verdict stays FAIL through the convergence result, the score, and the blocking count.
- MP-9 was judged by hand. A multi-line ordering record that the verb would have joined could be missed. I read every keyword-bearing line in both files.
- plan_artifact_hash is derived (G2). A framing error in my replay would show as a mismatch with the in-tree function. I checked the framing against the source's byte order and the header lengths (20, 23, 13, 14 bytes, plus four separators).
- Codex is the only backend (G5). A second backend could disagree on any rubric band.
- D14 depends on how the operator classifies release-blocking criteria. If the 15 criteria are not all gating, D14 narrows to a documentation fix.

## Must-pass results

| Criterion | Result | Evidence and layer |
|---|---|---|
| MP-1 REQ number consistency | PASS | REQ-CAP-001 to REQ-CAP-013 in order at spec.md:56–77 (E14). No gap, no duplicate, three-digit padding throughout. Requirement layer. |
| MP-2 GEARS format compliance | FAIL | REQ-CAP-008 (spec.md:69) and REQ-CAP-009 (spec.md:70) begin "On apply," under a "(Ubiquitous)" label. "On" is not a GEARS trigger keyword (When, While, Where), and the Ubiquitous shape has no precondition. Neither entry matches exactly one pattern, and the trigger "apply" is undefined (D11). REQ-CAP-005 (spec.md:63) and REQ-CAP-011 (spec.md:75) are "(Unwanted)", counted as legacy equivalents under M3, not failures. Judged against the requirement layer (spec.md) only; acceptance.md is the verification layer. |
| MP-3 YAML frontmatter validity | PASS | spec.md:2–15 carries all 12 canonical fields with the correct types (E11). No snake_case alias (E10). |
| MP-4 Language neutrality | PASS | No language-specific tool name in spec, plan, or acceptance (E9). The SPEC is also Go-project-scoped, so the criterion is N/A-equivalent. |
| MP-5 D7 cross-SPEC reconciliation | PASS | All three referenced SPECs are `completed` (E13). None is retired, superseded, or archived, so no REVIEW or BLOCKING line arises. |
| MP-6 D8 cross-platform discipline | PASS | `syscall` appears 0 times in every SPEC file (E8), so D8-4 auto-passes. |
| MP-7 clarification gate | PASS | No `[NEEDS CLARIFICATION` in plan.md (E7). research.md is absent in Tier M, so only plan.md applies. The open decisions Q1–Q4 are decision-index rows, not clarification markers. |
| MP-8 RED-now cell re-execution | N/A | acceptance.md:7 classifies no criterion as release-blocking (E16), so MP-8 is N/A on the file's own classification. This N/A is contested by D14: the DoD at acceptance.md:153 gates completion on all 15 criteria. |
| MP-9 cross-artifact ordering | PASS (hand-judged) | Milestones in plan order: M1 (plan.md:9) < M2 (16) < M3 (22) < M4 (27) < M5 (33) < M6 (39) (E17). plan.md has no `Exit:` line, so no criterion is bound to a milestone mechanically, and no CONFLICT line can arise. Each ordering candidate (E18) binds consistently with plan order: acceptance.md:74 and :76 "immediately before M2" matches plan.md:19; acceptance.md:84 "after the registration that Q4 authorizes" matches M1; acceptance.md:145 "written before its disposal" matches plan.md:31; plan.md:20 "before this milestone's commit is pushed" matches the Q3 gate. No conflict found. The mechanical verb was refused (G1). |

## Category scores (rubric-anchored)

| Dimension | Score | Band | Evidence | Codex |
|---|---|---|---|---|
| Clarity | 0.50 | multiple requirements need interpretation | "On apply" (spec.md:69–70); "owner", "confirmed owner", and "live session" are undefined (spec.md:75–76; plan.md:24; E26); "after the operator confirms" (spec.md:69) has no confirmation surface, and plan.md:37 defers it. | 0.50 (agree) |
| Completeness | 0.75 | all sections present; frontmatter complete; Out of Scope has H3 headings with bullets (spec.md:81–105) | Outcome branches are missing: DoD at acceptance.md:153 has no decision-outcome branch; Q2 and Q3 lack Default and Alternate (E27). | 0.75 (agree) |
| Testability | 0.50 | several criteria pass on the wrong thing | Comment-only ignore edit passes AC-CAP-004 (acceptance.md:62). Non-rotating run passes AC-CAP-007 (:85). Count-only check in AC-CAP-014 (:138). Masked exit in AC-CAP-002 (:44). AC-CAP-010 and AC-CAP-011 name no command or fixture. | 0.50 (agree) |
| Traceability | 0.75 | mapping complete; depth incomplete | 13 of 13 REQs have an AC and 15 of 15 ACs map to a REQ (E15; no orphan, no uncovered). Depth gaps: REQ-CAP-003 archive keep (acceptance.md:72–77); REQ-CAP-011 `--yes` path (:127–146); REQ-CAP-013 fields (:134–139). AC-CAP-002 measures a 10 GiB target that no REQ states (D10). | 1.00 (disagree) |

Overall score = mean of the four dimensions = (0.50 + 0.75 + 0.50 + 0.75) / 4 = 0.625. Threshold for Tier M: 0.80. Not met.

## Defects found

Blocking (blocking_count = 14):

D1. REQ-CAP-004 / M1 rotation mechanism — plan.md:11; spec.md:62 — Registering `usage-log.jsonl` in `sinkRegistry` cannot make the rotator act on `.moai/harness/usage-log.jsonl`. The rotator resolves every registered name under `LogDir` = `<projectRoot>/.moai/logs` (rotate.go:26–27, 102), and the existing retention prune `PruneStaleEntries` (retention.go:59–64) stays in force. Neither the path mechanism nor the prune's disposition is stated. — Severity: critical — Class: blocking — Required fix: state how a non-logs sink is resolved (or move the file under the rotator's directory); retire or serialize the existing prune in M1; add an AC that rotates a fixture at the real path.

D2. REQ-CAP-011 disposal guard unverified on the destructive path — acceptance.md:127–132 (AC-CAP-013); :141–146 (AC-CAP-015) — AC-CAP-013 runs report mode only and has no live-session fixture. AC-CAP-015 uses candidates that already pass the predicate chain, so a `--yes` path that skips the unpushed or live check still passes. — Severity: critical — Class: blocking — Required fix: add `--yes` fixtures for an unpushed worktree, an unowned worktree, and a live-session worktree, each asserting that it is preserved and not removed.

D3. AC-CAP-007 passes without a rotation — acceptance.md:81–86 — The Then clause ("one primary file and at most one kept rotation; no file truncated in place") holds when nothing rotates. — Severity: major — Class: blocking — Required fix: require that a rotation occurred (a displaced `.1` file exists with the former primary content) and that the primary's content is preserved.

D4. The Definition of Done cannot be met under a legitimate decision outcome — acceptance.md:150–153 (§D.6) — The DoD requires AC-CAP-001 to AC-CAP-015 to pass. AC-CAP-007 is gated by Q4, whose default keeps the owner. AC-CAP-015 is gated by Q2, which can be rejected. AC-CAP-009 has an unapproved branch. No N/A or outcome branch is defined. — Severity: major — Class: blocking — Required fix: map each decision outcome to the criteria it applies to, and state the N/A treatment in §D.6.

D5. Keep-3 rule and existing code — spec.md:69 (REQ-CAP-008); spec.md:47 and :118 (§B row, A10); plan.md:37 — REQ-CAP-008 says "the three most recent entries". The existing rule keeps the current release plus three non-current ones (defaults.go:447–451; spec.md:47), so whether "current" counts toward the three is undecided and deferred to run. AC-CAP-010 and AC-CAP-011 name no executable command or fixture. — Severity: major — Class: blocking — Required fix: decide the count rule (whether the current release is among the three), name the cap configuration key and its validation, and give each AC a command and a fixture definition.

D6. Ignore-rule criterion passes on comments — acceptance.md:58–63 (AC-CAP-004) — The grep pattern matches comment text. spec.md:34 already records comment-line matches (240, 258, 263, 282) for the broader pattern, so a comment-only edit passes while `git check-ignore` exits 1 (codex fixture). — Severity: major — Class: blocking — Required fix: verify ignore behavior with `git check-ignore --no-index` on one representative path and one corrupt-variant path.

D7. AC-CAP-014 is count-only — acceptance.md:134–139 — A list of `{}` entries satisfies the count. REQ-CAP-013 requires path, branch, size, and a table form (spec.md:77). — Severity: major — Class: blocking — Required fix: compare the path and branch sets with `git worktree list --porcelain`, require a size field, and check the table output.

D8. REQ-CAP-003 archive keep is not verified — acceptance.md:72–77 (AC-CAP-006) — Only the two state files are checked for local survival. The three `.corrupt` archives (spec.md:58) have no presence or hash check. — Severity: major — Class: blocking — Required fix: extend AC-CAP-006 to the three archives, recording presence and content hash before and after M2.

D9. AC-CAP-002 command masks errors and truncates paths — acceptance.md:44 — `awk '/^worktree /{print $2}'` truncates any worktree path that contains a space. The pipeline has no pipefail, so its exit status is the last awk's (0) even when `du` fails (codex measured a 144 KiB path with exit 0). — Severity: major — Class: blocking — Required fix: list paths NUL-safely, propagate per-path `du` failures, and check the pipeline's exit status.

D10. AC-CAP-002 measures a target no REQ states — acceptance.md:16, :41–47; spec.md:76 (REQ-CAP-012) — The "10 GiB" target appears only in spec.md:36 and :113 (A5), the AC title, and the AC body. REQ-CAP-012 says nothing about size, so the AC is traced to a requirement whose text it does not measure. — Severity: major — Class: blocking — Required fix: add a REQ for the worktree-size outcome, or re-trace AC-CAP-002 to the REQ it actually measures.

D11. MP-2 requirement notation — spec.md:69–70 (REQ-CAP-008, REQ-CAP-009) — "On apply," precedes a "(Ubiquitous)" requirement. Neither matches a GEARS pattern. The trigger "apply" is undefined: `moai clean --audit-logs --apply` (spec.md:38; plan.md:13) and the hygiene auto path could both be meant. — Severity: major (critical for MP-2) — Class: blocking — Required fix: rewrite each as "When the hygiene apply path runs under <named switch>, the backups and releases directories shall …" (and the same for REQ-CAP-009), naming the trigger.

D12. Baseline-first ordering cannot be witnessed by the commit graph (verification-claim-integrity §2.3) — acceptance.md:74–76 (AC-CAP-006); plan.md:19 (M2); acceptance.md:47 (AC-CAP-002 "before any disposal") — The byte-count baseline "immediately before M2" is recorded in the same milestone as the change. The plan specifies no preceding baseline commit, so the graph cannot witness the ordering. — Severity: major — Class: blocking — Required fix: specify a baseline-only commit that precedes the M2 commit, or rewrite the ordering clause to what the graph can witness.

D13. Owner, confirmed-owner, and live-session predicates are undefined — spec.md:75–76 (REQ-CAP-011, REQ-CAP-012); plan.md:24 (M3) — "Owner" is never defined, and no predicate says how ownership is confirmed or how a live session is detected. The REQ-CAP-012 gate and the "unowned" half of AC-CAP-013 both depend on these predicates. — Severity: major — Class: blocking — Required fix: define the owner source, the confirmation rule, and the live-session detection predicate (name the existing `sweep.go` predicate if it is reused), and cite them from REQ-CAP-010 to REQ-CAP-012.

D14. The release-blocking classification contradicts the DoD — acceptance.md:7 vs acceptance.md:153 — Line 7 says no criterion is release-blocking, yet §D.6 gates completion on all 15 passing. If the gate is intended, MP-8 applies: only AC-CAP-014's RED cell records an exit code (acceptance.md:28); the RED commands of AC-CAP-001, -002, and -005 use pipes, which fall outside the single-invocation form (verification-completeness §2.1); AC-CAP-007, -010, -011, and -015 are "RED by source". — Severity: major — Class: blocking — Required fix: classify the release gate explicitly. If all 15 gate, rewrite the RED cells in the four-element form and remove pipes from the cited commands.

Optional (not counted in blocking_count):

O1. Legacy "Unwanted" form — spec.md:63 (REQ-CAP-005), spec.md:75 (REQ-CAP-011) — Counted as legacy equivalents under M3, but the rubric says never to steer toward this form. Rewrite as Event-driven or Ubiquitous at the next edit. — Severity: minor — Class: optional.

O2. Decision records incomplete — decision-index.md:15–28 (Q2 and Q3) — No Default or Alternate line, unlike Q1 (lines 11–12) and Q4 (lines 36–37). The fail-closed outcome for Q2 is stated nowhere (see D4). — Severity: minor — Class: optional.

O3. Lifecycle state duplicated outside spec.md — progress.md:3 (`status: draft`) beside spec.md:5 — The schema's single-source wording (spec-frontmatter-schema.md, Artifact Statelessness) names spec.md as the one place. progress.md is outside that rule, so this is an observation, not a violation. — Severity: minor — Class: optional.

O4. Q3 names no measurement — decision-index.md:23–28 — The EVIDENCE-NEEDED row does not say how the operator would count the clones. The count cannot be taken from this worktree. A short pre-flight instruction would help. — Severity: minor — Class: optional.

O5. Pin wording — acceptance.md:5 names `2aab5f797`. The audited commit adds only the five SPEC files (E22), so the code measurements still hold. The pin wording could say so. — Severity: minor — Class: optional.

## Open decisions (operator-owned; not decided here)

- Q1 (FOUNDER, hygiene default `report` or `apply`): OPEN. Operator verdict empty (decision-index.md:13). Default `report`. Presence is not a plan defect: REQ-CAP-006 is conditioned on it. Its consequence for the DoD is D4.
- Q2 (FOUNDER, bulk worktree-sweep approval): OPEN. Operator verdict empty (:21). No Default or Alternate line (O2). It gates AC-CAP-015 and REQ-CAP-012. A rejection has no completion branch (D4). Presence is not a defect; the missing outcome branch is.
- Q3 (EVIDENCE-NEEDED, clones with local data): OPEN. Operator verdict empty (:28). It gates the push of the M2 commit (plan.md:20; acceptance.md:153). The count cannot be measured from this tree (O4). Presence is not a defect.
- Q4 (FOUNDER, usage-log rotation owner): OPEN. Operator verdict empty (:38). Default keeps `internal/harness/retention.go` as owner (:36). It gates AC-CAP-007 and REQ-CAP-004. D1 shows the mechanism fails even if Q4 is approved as written.

Judgment on the presence of the open decisions: they are operator gates that the plan states explicitly, and REQ-CAP-004, REQ-CAP-006, and REQ-CAP-012 are conditioned on them. Their presence is therefore not a plan defect. The defects that follow are D1 (the Q4 mechanism), D4 (no outcome branches), and O2 (missing Q2 and Q3 defaults).

## Regression check

Not applicable: this is iteration 1 of the plan audit.

## Recommendation

Revise the SPEC (manager-spec, via the orchestrator). Address D1 to D14 in the fix_scope above. Then re-audit with scope `full`, because D1 and D4 change the criteria and the DoD structure, and a re-read of individual hunks would not be enough. Before the re-audit, the operator should record verdicts for Q1 to Q4 or confirm that each open decision keeps its default outcome.

## Operational Notes (unverified)

- Measure the AC-CAP-002 pipeline on a worktree path that contains a space. Label: inferred (shell and awk field splitting); codex reports a measured instance (G6).
- Compare the rotator's `LogDir` with the real path of usage-log.jsonl before planning M1. Label: measured by this audit (rotate.go:26–27, 102). This is a fact, not advice.
- Land a baseline-only commit before the M2 commit. Label: assumption (not tested); D12 is the governing rule.

reread_hunks is omitted deliberately: a re-read of individual hunks does not suffice, because the verdict rests on 14 blocking findings across spec.md, plan.md, acceptance.md, and decision-index.md, and on a required-backend fail.
