# Progress — SPEC-WEB-BROWSER-OBS-001

카드 t1081 · 트리 WT-cdp-observation @ cd99336bf · 측정 전용 SPEC (tier M)

## §E.1 Plan-phase Audit-Ready Signal

_<pending plan-audit>_ — plan-phase 산출물 4종(spec.md / plan.md / acceptance.md / progress.md) 작성 완료. plan-auditor 실행 시 `plan_status: audit-ready` + `plan_complete_at` 기록 예정.

## §E.2 Run-phase Evidence

### AC PASS/FAIL matrix (detail: `.moai/reports/t1081/verdict.md`, confidence class `browser-observed`)

| AC | Status | Actual output (verbatim basis, this run, tree 06bd697c6) |
|----|--------|----------------------------------------------------------|
| AC-BO-001 | PASS | Failed-submit slot reached as visible text: `save__msg save__msg--error` `role="alert"`, display flex / visibility visible / 686.56×16.67px, text `could not save profile preferences: write preferences: open /tmp/t1081-home/.moai/claude-profiles/preferences.yaml: permission denied` (stable seam-1 phrase present; generic fallback absent) — live CDP session, no string probing |
| AC-BO-002 | PASS | Tri-class = **htmx-swap**: `signal_a_real_nav_count: 0` (only `Page.navigatedWithinDocument(record-only)` pushState events), signal (b) slot childList mutation captured by the injected MutationObserver; href axis recorded (`/settings` → `/save?profile=default`), never judged |
| AC-BO-003 | PASS | Measured (vs t1051 premises fail 2 / success 0 / ~110KB — kept separate): failed-submit console errors **0**, success-submit console errors **0** (Runtime/Log channels empty), failed-response body **112,972 bytes** (`Network.getResponseBody`, status 200) |
| AC-BO-004 | PASS | Incidental: validation-reject submit → status **400**, body 113,168 B discarded, `htmx:responseError` fired, no swap, no nav (classification `swap-failed-no-nav`); lead-routing note: consistent with t1051's discard premise, contradicts nothing |
| AC-BO-005 | PASS | `git diff --name-only $(git merge-base develop HEAD)..HEAD` → 0 `internal/` paths; verdict cites no t1080/t1051 verdict as basis; REQ-BO-007 not triggered (observation matches REQ-A) |
| AC-BO-006 | PASS | Orphan probes empty (ports 3477/9478: 0 LISTEN; 0 own processes; temp chrome dir gone); evidence resident under `.moai/reports/t1081/` (gitignored) |

### Session record

- Binary `moai-bin` built in this run from tree 80ed77d28 (docs-only delta from base cd99336bf — `git diff --name-only cd99336bf..80ed77d28 -- internal/ | wc -l` → 0).
- Environment per plan §C.1/§D: Chrome standard-path discovery, `--headless=new`, dedicated ports 3477 (web) / 9478 (CDP), liveness-checked, `timeout 600` wrappers, `mktemp` user-data-dir, HOME isolated to `/tmp/t1081-home`, project root `/tmp/t1081-project` (temp scaffold).
- Failure induction (plan M1.2): seam 1 `writePreferences`, `os.chmod(preferences.yaml, 0o444)` in the isolated HOME store; expected stable phrase per t1051 §5.1 row 1 — observed phrase matched.
- Instrument integrity (`verification-completeness.md` §1.1): deliberate wrong-input check observed FAIL in every run (`integrity-negative-control`), framework-saw-failure PASS.
- Regression guard (plan §E5): `unset MOAI_KANBAN … && go test ./internal/web/...` → `ok github.com/modu-ai/moai-adk/internal/web 25.779s`.
- Console-error attribution: probe-origin vs page-origin separated per acceptance §C.4; page-load window carried one unrelated 404 console error (context only).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-22
run_commit_sha: be3d4131d   # backfilled after the M2-M4 record commit landed (D3 backfill window)
run_status: complete
ac_pass_count: 6
ac_fail_count: 0
preserve_list_post_run_count: 4
new_warnings_or_lints_introduced: 0   # zero production code changed; golangci-lint baseline untouched
cross_platform_build: not-applicable  # measurement-only SPEC; no code built for release
total_run_phase_files: 0              # zero tracked source files; evidence is gitignored local artifacts
m1_to_mn_commit_strategy: per-milestone commits on WT-cdp-observation (M1 06bd697c6; M2-M4 record commit)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

_<pending orchestrator — run 단계 진입 전 기록>_

---

## Run-phase 참고 메모 (plan-phase 작성)

- MCP 서버 project root 는 spawn-frozen 이므로, run 중 `mcp__moai__*` 호출은 전부 `project_root` = 본 워크트리의 `git rev-parse --show-toplevel` 절대 경로를 명시 전달한다(spec.md HARD-7).
- 판정 기록 거처: `.moai/reports/t1081/verdict.md`. 프로브·캡처도 같은 디렉터(무추적).
- 레인 의무: 완료 보고에 카드 id · 브랜치와 HEAD · 로컬 병합 SHA · 증거 경로 — push 는 리드 일괄(CLAUDE.local.md §4.1).
