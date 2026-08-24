# CR Round-1 Response — PR #1642 (t225, SPEC-V3R6-AUDIT-MODEL-PIN-001)

- Review: CodeRabbit `Review completed` @ head 9c5294763, Merge Risk 🟡 Moderate, 10 inline comments
- Triage: 8 ACCEPTED / 2 REJECTED (each rejection grounded in measurement)
- Fix commits: `3ee1da61c` (CR #4-#8, code+tests — manager-develop) · this commit (CR #2 docs — manager-spec; CR #3 + #10 — lane)
- All anchors verified in-tree by the lane BEFORE acceptance; no comment taken on trust.

## Per-comment disposition

| # | Location | Severity | Verdict | Disposition |
|---|----------|----------|---------|-------------|
| 1 | sync-audit-review-1.md:34 | Minor/Security | **REJECTED** | Phantom: line 34 reads "Self-consistent: tests rewritten… 26.3s green in this tree." — no username, no workstation path. `grep -n "goos\|/Users/"` across ALL 3 tracked t225 reports = **0 matches**. The cited exposure does not exist. |
| 2 | plan.md:147 + 183-185; spec.md §H | Major | **ACCEPTED** | Staleness the author had self-flagged at amendment time. plan.md M3: delivery field = hypothesis B (top-level `reasoning_effort`), hypothesis A measured true null (1.02 < 1.1 bound); M5: 1.25× rule + null-control discriminant cross-referenced to the acceptance.md amendment record; spec.md §H: overlay claim REVERSED per the live differential. Residual-stale grep 0 matches; `moai spec lint` rc=0. |
| 3 | progress.md:453 | Minor (MD024) | **ACCEPTED** | Duplicate `## §E.3` placeholder heading removed (lane edit); the completed §E.3 record remains the single anchor. Gap line updated to record the removal. |
| 4 | audit_pin_live_test.go:277 | Minor | **ACCEPTED** | Evidence line now cites the measured input: `git diff 63e10bc1b..HEAD -- internal/cli/mcp_codex.go (%d bytes)` (CR's suggested form). The "~19-line cache.go" figure belonged to attempts 1-3 (const comment :44-46). |
| 5 | audit_pin_live_test.go:338 | **Major** | **ACCEPTED** | `probeWriteTranscript` moved INSIDE the `tapPtr != nil && *tapPtr != nil` guard — nil runner no longer panics the package binary (CR's diff adopted as-is). |
| 6 | mcp_glm_audit_pin_test.go | Trivial | **ACCEPTED** | 3 tests consolidated into table-driven `TestGLMAuditPin_RequestBody`, 5 states — adding the 2 previously-uncovered wire states (`high` verbatim; model-pinned+empty-effort → directive omitted, model still pinned). no-thinking-object assertion applied per-row. |
| 7 | mcp_glm_audit_pin_test.go:161 | Trivial | **ACCEPTED** | Absent-pin assertion narrowed from whole-body `"reasoning"` to `"reasoning_effort"`/`"thinking"` field presence — diff content can no longer false-fail the test. |
| 8 | mcp_glm.go:170 (+ convergence :463) | **Major** | **ACCEPTED** | Real worktree-tree mismatch (same family as the `project_root` MCP doctrine): `resolveGLMAuditModelEffort()` read `projectDirResolver()` (primary checkout) while the reviewed diff came from the caller-named tree. Now takes `projectRoot string` (non-empty wins, `projectDirResolver()` fallback); `handleGLMAudit` resolves AFTER `resolveToolProjectRoot` and passes that root; `mcp_convergence.go` passes its already-computed root. 5 call sites updated (incl. 3 fallback tests asserting the same path via ""). Matches codex counterpart semantics (mcp_codex.go prefers `params["cwd"]`). |
| 9 | audit_models.go:78 | Major/Heavy | **REJECTED** | The sibling seam `loadLLMSectionOnly` lives at `internal/cli/glm.go:624` — the established precedent for section-only CLI loaders IS internal/cli. The new loader mirrors it in `internal/cli/audit_pin.go`; moving it alone would break the symmetry the codebase already has. Package-wide Loader-API unification is a legitimate follow-up card candidate (recorded), not a minimal CR fix. |
| 10 | template workflow.yaml:64 | Minor | **ACCEPTED** | Comment overclaimed: "an unservable model id falls back to the SSOT path" is codex-only (servability filter); GLM sends a non-empty pin verbatim and a wrong id degrades via z.ai fail-open to inconclusive (author's design decision D3, deliberate). Comment scoped per-backend; template regenerated (`make build`), values still ship EMPTY. Local workflow.yaml comment never carried the claim (verified). |

## Verification (lane-measured @ 3ee1da61c + this working tree)

- `go vet ./internal/cli/` → rc=0; `go build ./internal/cli/` → ok
- `go test ./internal/cli/ -run 'TestGLMAuditPin|TestAuditPin|TestGLMAuditParse|TestCodexAuditPin|TestDoctorCmd' -count=1` → **ok 44.439s**
- manager-develop's battery: `TestGLMAuditPin_RequestBody` (5 subtests incl. the 2 new states) PASS; surrounding sweep `ok 10.913s`; unfiltered full suite (F-pass tree) `ok 293.649s`
- `moai spec lint .moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/spec.md` → rc=0 `✓ No findings`
- Merge-integrity (lead question 1): t200 doctor commit 294b4b6ab shares **0 files** with our diff (comm-verified); behavioral proof on the merged tree: `TestDoctorCmd|TestExitCodeContract|TestDoctorConstitution` → **ok 24.873s**, pin battery → **ok 2.549s** — both sides alive.
- CR status (lead question 2): review genuinely ran (`Review completed`, Merge Risk pinned `up to 9c529` = current head); earlier "rate limited" resolved by the lead's retrigger.

## Residual

- Moderate → re-review needed on the new head after this push (all 8 accepted fixes + 2 evidence-grounded rejections); if it stays Moderate on non-blocking residue, the lead requests operator override.
- Follow-up card candidates recorded: llm.yaml user-modification preservation on `moai update` (B-plan, lead queued); template overlay-doc correction (§H reversal); Loader-API unification (#9 rejection rationale).
