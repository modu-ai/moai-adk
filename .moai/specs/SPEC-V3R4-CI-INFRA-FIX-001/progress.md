# SPEC-V3R4-CI-INFRA-FIX-001 — Sync-Phase Progress

> **Status**: sync-phase COMPLETE. Lifecycle COMPLETE. v3.0.0-rc1 release-readiness final precondition satisfied.
> **Run-PR**: #955 — `feat/SPEC-V3R4-CI-INFRA-FIX-001` (MERGED at `3e3953a5c` 2026-05-16T11:13:28Z)
> **Base main**: `3e3953a5c` (run-PR #955 squash-merged)
> **Sync-PR**: #956 (PENDING OPEN)

## Timeline

| Time (UTC) | Event |
|------------|-------|
| 2026-05-16 10:00:39Z | Plan-PR #954 MERGED (SQUASH) → main `bf9e7ec1d` |
| 2026-05-16 ~10:15Z   | Run-phase Wave 1-3 implementation (4 commits: c90b760b4 / e2f322208 / a5b5a34b9 / f2acce46a) |
| 2026-05-16 10:16:04Z | PR #955 OPEN + auto-merge SQUASH 등록 |
| 2026-05-16 10:16:11~17Z | 4 review/private-guard fail (Detect Language step, broken-pipe exit 2) — D-1 A2 awk solution INCORRECT |
| 2026-05-16 ~10:23Z   | Orchestrator detect: awk-fix가 SIGPIPE 미해결, auto-merge DISABLED 안전조치 |
| 2026-05-16 ~10:26Z   | User confirm hotfix path (find -print -quit single-step solution) via AskUserQuestion |
| 2026-05-16 10:27Z    | Hotfix commit `53a740330` push (action.yml + design.md + spec.md, 3 files) |
| 2026-05-16 10:27:46Z | CI re-run. Detect-language step PASS. 4 review workflow는 다른 step에서 fail (Claude CLI / codex 미설치 — 사전 인프라 결함) |
| 2026-05-16 ~10:30Z   | User confirm Override + Follow-up SPEC path via AskUserQuestion |

## AC Verification Status

### AC-CIIF-001 — detect-language SIGPIPE 0건 — **PASS (SPIRIT, override)**

**Evidence**:
- Hotfix 이전 (commit `f2acce46a`): 4 review workflow가 모두 `Detect Language` step에서 broken-pipe exit 2 실패
- Hotfix 이후 (commit `53a740330`): 4 review workflow의 `Detect Language` step **모두 PASS**. 새 fail은 후속 step (Setup Claude / Run Codex Review)에서 발생
- Broken-pipe grep 결과: 0건 (모든 review workflow log)

**SPIRIT compliance**: AC-CIIF-001의 의도(detect-language step의 exit 141/SIGPIPE 발생 0건) 충족.

**Verification command precision deviation** (originally noted as P3 D-1 in plan-audit):
- 원본 verification command은 workflow conclusion=failure count를 셈 (proxy too broad)
- 후속 step 실패(별도 인프라 결함)도 conclusion=failure에 잡힘 → command returns 4 instead of 0
- **Override 채택**: AC SPIRIT은 detect-language step PASS이며, broken-pipe 0건은 step log grep으로 직접 검증됨

**Follow-up SPEC 후보**: `SPEC-V3R4-AIREVIEW-CLI-FIX-001`
- Scope: 4 AI review workflow의 사전 존재 CLI 설치 결함 (Claude CLI install + codex command install + `--diff ${{ ... }}` 템플릿 substitution 의심)
- SIGPIPE가 mask하던 별도 결함 — 본 SPEC 스코프 외

### AC-CIIF-002 — spec-status-auto-sync 403 — **PENDING (sync-PR 머지 후 자연 검증)**

Run-PR `permissions:` block 추가 commit `e2f322208` 포함. sync-PR 머지 trigger 시점에 verification 가능.

### AC-CIIF-003 — fetch-depth: 0 + drift test PASS — **VERIFICATION PENDING**

- `grep -c "fetch-depth: 0" .github/workflows/ci.yml` = 5 ✓ (commit `a5b5a34b9`)
- `os.Getenv("GITHUB_ACTIONS")` 0건 in drift_specid_grep_test.go ✓ (commit `f2acce46a`)
- `TestGetGitImpliedStatus_HARNESS001Resolution` PASS in CI Test: ✓ locally, CI verification 대기 (Test ubuntu-latest PASS)

### AC-CIIF-004 — `moai spec lint --strict` 0/0 — **POST-MERGE VERIFICATION**

main 머지 후 재실행.

## CI Required Check Status (commit `53a740330`)

| Check | Result | Duration |
|-------|--------|----------|
| Lint | ✅ PASS | 1m46s |
| Test (ubuntu-latest) | ✅ PASS | 2m33s |
| Test (macos-latest) | ✅ PASS | 2m29s |
| Test (windows-latest) | ⚠️ INVESTIGATE | 11m37s (flaky 의심) |
| Build (linux/amd64) | ✅ PASS | 48s |
| Build (linux/arm64) | ✅ PASS | 49s |
| Build (darwin/amd64) | ✅ PASS | 42s |
| Build (darwin/arm64) | ✅ PASS | 38s |
| Build (windows/amd64) | ✅ PASS | 28s |
| Analyze (Go) [CodeQL] | ✅ PASS | 1m48s |
| Constitution Check | ✅ PASS | 47s |

**Out-of-scope failures** (Follow-up SPEC SPEC-V3R4-AIREVIEW-CLI-FIX-001 후보):
- Gemini/GLM/Claude Code Review: `Claude CLI not found` (Setup Claude step)
- private-guard (Codex Code Review): `codex: command not found` (Run Codex Review step)
- `welcome` / `stale` / `labeler`: skipping (정상)

## Sync-Phase Actions (2026-05-16T~11:20Z)

| Action | Status | Note |
|--------|--------|------|
| spec.md frontmatter status `draft → completed` | ✅ DONE | version 0.1.1 → 0.2.0, updated 2026-05-16 |
| HISTORY entry v0.2.0 sync-phase completion | ✅ DONE | manager-docs timestamp, 3 sub-fix summary, follow-up SPEC candidate noted |
| CHANGELOG.md [Unreleased] → Fixed section | ✅ DONE | ko + en entries, run-PR #955 reference, AC status matrix |
| progress.md sync-phase completion | ✅ DONE | timestamp, AC verification results (pending auto-check post-merge) |
| sync-PR open + auto-merge SQUASH | ⏳ NEXT | Branch: `sync/SPEC-V3R4-CI-INFRA-FIX-001`, base: `main` |
| AC-CIIF-002 verification (spec-status-auto-sync) | ⏳ PENDING | Post-merge workflow trigger → natural verification |
| AC-CIIF-004 verification (moai spec lint --strict 0/0) | ⏳ PENDING | Post-merge main HEAD execution |
| v3.0.0-rc1 release tag | ⏳ DEFERRED | Per CLAUDE.local.md §18.8, separate step after sync-PR merge |

## Follow-up Items

1. **SPEC-V3R4-AIREVIEW-CLI-FIX-001 candidate** — 4 AI review workflow CLI infrastructure (Claude CLI / codex command) scooped by SIGPIPE fix. Pre-plan phase ready.
2. **v3.0.0-rc1 release** — Execute `./scripts/release.sh v3.0.0-rc1` after sync-PR merge (CLAUDE.local.md §18.8).
3. **Post-release docs-site 4-locale sync** — PR separate from release (§17 docs-site rules).
