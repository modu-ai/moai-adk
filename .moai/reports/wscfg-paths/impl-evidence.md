# wscfg-paths — Windows first-exposure cluster ② (paths/overrides) — Implementation Evidence

Card: wscfg-paths · Branch: `WT-wscfg-paths` · Base: `origin/release/v3.1.1` @ `a39646a91` (verified equal to dispatch base) · Impl commit: `44ea01575`

## 1. Claim

1. The 10 Windows-CI failures (job 95500006316) share **one root cause**, and it is TEST-side, not production-side: fixtures supply unix-style absolute literals (`"/tmp/moai-home-override"` ×2 sites in `internal/paths/paths_test.go`; `"/opt/shared/astgrep-rules"` in `internal/cli/astgrep_rulesdir_test.go`). On Windows `filepath.IsAbs("/tmp/...")` is correctly `false`, so `MOAI_HOME` was disregarded (XDG semantics) and `rules_dir` was joined onto the project root — the assertion then compared against the wrong arm.
2. Production is already correct and is **unchanged**: `internal/paths/paths.go` `MoaiHome` and `internal/hook/quality/astgrep_gate.go` `ResolveRulesDirPath` both gate on per-GOOS `filepath.IsAbs`. The lead's hypothesis ("absolute-path detection bound to unix filepath semantics") is **refuted for production** and confirmed for the test fixtures — the same family as 3d51358f2's `promote_test` fix ("45 down to 2" left these stragglers behind).
3. Fix: platform-native fixtures (`filepath.Join(os.TempDir(), …)`), single-quoted `rules_dir` YAML (double quotes would eat `C:\Users\…` backslashes as escapes), and GOOS=windows-guarded volume-letter regression tests including 8.3 short-form variants (`PROGRA~1`, `RUNNER~1` — cluster-① worker's observation folded in as pinned contract: nothing canonicalizes short paths).

## 2. Evidence

Affected fixture sites (pre-fix):

```
internal/paths/paths_test.go:62   t.Setenv("MOAI_HOME", "/tmp/moai-home-override")        # TestMoaiHome_OverrideAbsolute
internal/paths/paths_test.go:185  t.Setenv("MOAI_HOME", "/tmp/moai-home-override")        # TestSubAccessors /override arm ×8
internal/cli/astgrep_rulesdir_test.go:56  writeGateYAML(t, dir, gateYAMLWithRulesDir("/opt/shared/astgrep-rules"))
```

Production predicates (verified, unchanged):

```
internal/paths/paths.go:69        if v := os.Getenv(EnvHome); v != "" && filepath.IsAbs(v) { return v, nil }
internal/hook/quality/astgrep_gate.go:59  if filepath.IsAbs(dir) { return dir }
internal/cli/astgrep.go:84-86     flag path returns verbatim (explains why the "flag wins" subtest passed on Windows)
```

Post-fix runs (darwin, env-scrubbed compound form):

```
$ unset MOAI_KANBAN … && go test ./internal/paths/ -count=1 -v
--- PASS: TestMoaiHome_OverrideAbsolute (0.00s)
--- SKIP: TestMoaiHome_WindowsVolumeAbsolute (0.00s)      # designed skip on non-windows
--- PASS: TestSubAccessors_OverrideAndFallback (0.00s)     # 8 fallback + 8 override subtests
… 10 PASS + 1 SKIP total
ok  github.com/modu-ai/moai-adk/internal/paths

$ unset MOAI_KANBAN … && go test ./internal/cli/ -run 'TestResolveRulesDir|TestAstGrepCmdRulesDirFlagNoDefault|TestGAR' -count=1 -v
--- PASS: TestResolveRulesDir (0.00s)   # incl. reworked absolute subtest + windows volume subtest (skipped)
--- PASS: TestGAR_AC008/009/010 + TestGAR_D3   # writeGateYAML consumers unaffected by quoting change
ok  github.com/modu-ai/moai-adk/internal/cli

$ unset MOAI_KANBAN … && GOOS=windows go vet ./internal/paths/ ./internal/cli/
FINAL_VET_EXIT=0    # windows cross-compile clean (tests included)
```

Diff surface: `2 files changed, 69 insertions(+), 9 deletions(-)` — test files only.

## 3. Baseline-attribution

- All measurements above are from THIS tree (`WT-wscfg-paths` @ `a39646a91`), post-fix, run in this session; pre-fix fixture contents cited from the same tree before editing.
- The failure list itself (job 95500006316) is the lead-supplied baseline; I could NOT execute the Windows tests locally (darwin host — Windows filepath semantics not directly reproducible), so the Windows-side verdict is deferred to Windows CI per the dispatch's 검증 제약.

## 4. Gaps

- Windows runtime execution NOT observed locally — `GOOS=windows go vet` proves compilation only; the volume-letter/8.3 tests execute solely on Windows CI, and the original 10 failures are expected to clear there (final 판정 = windows CI).
- No full-suite local run (lane-local discipline; touched packages run directly).
- The "flag wins over config" subtest still uses the unix literal `"/explicit/flag/path"` as a FLAG value — left as-is because the flag returns verbatim (no IsAbs involved) and it is outside the failing list (scope discipline); noted as latent-but-inert.

## 5. Residual-risk

- If Windows CI surfaces OTHER unix-literal fixtures in these or sibling packages (cluster ③+), they will fail the same way; this card fixed the named failing groups plus the latent sites in the same two files.
- The 8.3 hint from cluster ① does not interact with this fix's diagnosis (8.3 forms are still volume-absolute and round-trip verbatim), but if some OTHER code path string-compares a short-form env path against a long-form derived path, that comparison can still mismatch — outside this card's measured surface.
- `t.Setenv("HOME", …)` behavior on Windows depends on the process env actually carrying HOME; tests that rely on HOME-empty fallback already characterize both arms, unchanged by this fix.
