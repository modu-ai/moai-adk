# Progress — SPEC-SERVED-MODEL-NORM-001 (card t1287)

## §A Premise (CLAUDE.local.md §30, three steps)

- Base: local develop `37dc766b9` (worktree `.claude/worktrees/t1287`, branch `WT-served-model-norm`).
- Premise alive. Binary built from this tree, `moai doctor --check "Served Model" --verbose`: `swept 2334 subagent transcripts: ok 1295, served_drift 982, unknown 11, unmapped 46`.
- Drift shapes: `expected=claude-opus-5[1m] served=[claude-opus-5]` ×295, `expected=claude-fable-5-1[1m] served=[claude-fable-5-1]` ×2, `expected=inherit served=[claude-sonnet-5]` ×1 → 298 false drifts (t1282 recorded 292). `expected=opus[1m] served=[glm-5.3-flash]` ×125 are true drifts.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-28

## §F Phase 4 Mode Selection

Decision: direct — Tier S, one function plus one test table, lane-executed under §31 autonomous kickoff.

## §E.2 Run-phase Evidence

- RED: `go test ./internal/hook/ -run TestServedModel_Classify -count=1` → 3 FAIL (e, f, h: `verdict = "served_drift"`); control g passed.
- GREEN: `go test ./internal/hook/ -run TestServedModel -count=1` → `ok github.com/modu-ai/moai-adk/internal/hook 5.241s`.
- Sweep with fixed binary: `swept 2335 …: ok 1592, served_drift 684, unknown 12, unmapped 47`; drift rows with `expected=…[1m]` or `expected=inherit`: 0.

## §E.3 Run-phase Audit-Ready Signal

이 통합 브랜치에 원 담당자 구현 커밋 `73d900b97`을 `-x`로 가져온 결과는 `13481c8ba`다. 최신 `develop` 위에서 재측정했다.

```text
$ go test ./internal/hook -run '^TestServedModel_(Classify|UnknownNeverOK)$' -count=1 -timeout 180s
ok  \tgithub.com/modu-ai/moai-adk/internal/hook\t0.609s
$ go test ./internal/cli -run 'TestRunDiagnosticChecks|TestServedModelCheck_|TestBinaryLag_' -count=1 -timeout 180s
ok  \tgithub.com/modu-ai/moai-adk/internal/cli\t8.318s
$ go vet ./internal/hook
(출력 없음, exit 0)
$ CLAUDE_CONFIG_DIR=/Users/goos/.moai/claude-profiles/moai-adk /tmp/moai-t1287-doctor doctor --check 'Served Model' --verbose
swept 2363 subagent transcripts: ok 1599, served_drift 706, unknown 11, unmapped 47
drift_detail_rows=706 legacy_false_drift_rows=0
```

마지막 두 줄은 doctor 출력을 `/tmp/t1287-served-doctor.log`에 저장한 뒤 요약과 drift 상세 행을 집계한 결과다. doctor exit 0. 이 결과는 현재 로컬 프로필 스냅샷에만 적용한다.
