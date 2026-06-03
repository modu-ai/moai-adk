# Progress — SPEC-GO-TOOLCHAIN-SEC-001

## Status: in-progress (run-phase)

Tier S (minimal). Hybrid Trunk 1-person OSS → main-direct, no PR.

## Phase Tracker

| Phase | Status | Owner | Commit |
|-------|--------|-------|--------|
| Plan | done | manager-spec | c5bd27fcf + 06c666a50 |
| Run (M1–M3) | done | manager-develop | 36c190fd0 |
| Sync (M4) | pending | manager-docs | — |
| Mx (close) | pending | orchestrator/manager-docs | — |

## Plan-phase notes

- SPEC ID self-check: `SPEC | GO | TOOLCHAIN | SEC | 001 → PASS` (canonical regex
  `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`; digit-only end anchor satisfied).
- Ground-truth verified at plan time:
  - `go.mod:3` = `go 1.26` (no patch, no toolchain directive)
  - installed toolchain = go1.26.0
  - 9 hardcoded `go-version: "1.26"` pins (claude.yml:42, ci.yml:87/179/205/250/292,
    codeql.yml:85, release-pr-multi-os.yml:59, release.yml:31)
  - 3 `go-version-file` users (template-neutrality-check:52, spec-lint:15,
    spec-status-auto-sync:25) — no edit needed
- Design decisions (resolved in plan.md): CI pin strategy = (b) migrate 9 → `go-version-file`
  (SSOT); toolchain acquisition = explicit `toolchain go1.26.4` directive + GOTOOLCHAIN=auto.

## Plan-patch (plan-auditor PASS-WITH-DEBT 0.88 → D1/D2/D3)

- D1 (SHOULD-FIX): REQ-GTS-006 was orphan (no AC) → added **AC-GTS-005** (`awk` toolchain/go
  directive equality guard). acceptance.md.
- D2 (SHOULD-FIX): the EC-1 false-PASS guard (verify `go version` ≥ go1.26.4 before trusting
  govulncheck 0-affecting) was only narrative (EC-1 / D.3 / progress checkbox) → promoted to
  first-class **AC-GTS-006** (toolchain pre-check gate) that GATES AC-GTS-001 evidence
  acceptance. Cross-reference note added on AC-GTS-001. acceptance.md.
- D3 (MINOR): added `tier: S` to spec.md frontmatter (was defaulting to Tier L 0.85; now
  authoritative Tier S 0.75, consistent with plan.md:L3 + progress.md:L5).
- AC count: 4 → **6** (AC-GTS-001..006). status remains `draft` (still plan-phase).
- Scope discipline held: no third-party / vitest / CI-gate scope added — pure
  traceability + testability hardening.

## §E.2 Run-phase Evidence (AC matrix — manager-develop)

6 ACs total (AC-GTS-001..006). AC-GTS-006 captured FIRST — it gates AC-GTS-001.
All evidence captured under verified effective toolchain `go version go1.26.4 darwin/arm64`
(auto-acquired via GOTOOLCHAIN=auto from the `go 1.26.4` directive).

| AC | Status | Verification command | Actual Output |
|----|--------|----------------------|---------------|
| AC-GTS-006 | PASS | `go version \| grep -qE 'go1\.26\.(4\|[5-9]\|[1-9][0-9])' && echo GUARD-OK` | `GUARD-OK` (effective `go1.26.4`; gates AC-GTS-001 — satisfied first) |
| AC-GTS-001 | PASS | `govulncheck ./...` (under go1.26.4) | `No vulnerabilities found. Your code is affected by 0 vulnerabilities.` (19 affecting stdlib → 0). 13 residual in "modules you require but your code doesn't appear to call" = NON-affecting, out of scope per spec §E. |
| AC-GTS-002 | PASS | `go build ./...` ; `go test ./...` (under go1.26.4) | `go build ./...` exit 0. `go test ./...` green except ONE pre-existing unrelated drift `TestOutputStylesTemplateLiveParity` (einstein.md template/live parity — NOT in this SPEC's changed-file set; uncommitted working-tree drift independent of toolchain change; `go test ./internal/template/ -skip TestOutputStylesTemplateLiveParity` = `ok`). |
| AC-GTS-003 | PASS | `grep -rn 'go-version: *"1.26"' .github/workflows/` ; `go.mod` go directive | 0 residual matches (9 migrated to `go-version-file: go.mod`) AND `go 1.26.4` ≥ 1.26.4. |
| AC-GTS-004 | PASS | `git diff go.mod` | Only the `go` directive line changed (`go 1.26` → `go 1.26.4`); zero `require` block lines; no `.go` source file modified. Changed-file set = go.mod + 9 workflow steps (ci.yml ×5, claude/codeql/release-pr-multi-os/release ×1 each). |
| AC-GTS-005 | PASS | `awk '/^go /{g=$2} /^toolchain /{t=$2} END{exit (t=="" \|\| t=="go"g)?0:1}' go.mod` | exit 0 (no `toolchain` directive present → `t==""` branch; see Decision note below). |

## §E.2.1 Run-phase Decision — `toolchain` directive dropped (go1.26.x tooling conflict)

**Decision**: the explicit `toolchain go1.26.4` directive (plan §D-DESIGN Decision 2 /
REQ-GTS-006) was **NOT retained**. The final go.mod carries only `go 1.26.4`.

**Empirical ground truth (go1.26.4)**: when the `go` directive already pins the full
patch (`go 1.26.4`), go1.26.x treats an additional `toolchain go1.26.4` directive as
**redundant** and:
- `go build ./...` / `govulncheck ./...` FAIL with `go: updates to go.mod needed; to update
  it: go mod tidy` while the directive is present.
- `go mod tidy -diff` proposes exactly one change: **remove** the `toolchain` directive.
  `go mod verify` = "all modules verified" (zero dependency-graph change).

**Why this stays within SPEC scope (no body change required)**:
- REQ-GTS-002 ("go directive ≥ go1.26.4") is fully satisfied by `go 1.26.4` alone.
- REQ-GTS-006 is a **Capability gate** ("**Where** an explicit `toolchain` directive is
  added…") — conditional on the directive being present; it does NOT mandate adding it.
- AC-GTS-005 is written to PASS when no toolchain directive exists (`t==""` → exit 0).
- AC-GTS-002 (build/test pass) is **only satisfiable** without the directive under go1.26.x.
- `GOTOOLCHAIN=auto` still auto-acquires go1.26.4 from the `go 1.26.4` directive (proven:
  `go: downloading go1.26.4` on first invocation, then `go version go1.26.4`). CI
  `actions/setup-go` with `go-version-file: go.mod` likewise reads the `go` directive.

The plan §D-DESIGN Decision 2 narrative assumed the directive would be inert; that
assumption is empirically false on go1.26.x. Resolution confined to the SPEC's declared
`module: go.mod` scope — a same-scope decision, not a scope expansion. No SPEC body
(spec.md/plan.md/acceptance.md) modified.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-03
run_commit_sha: 36c190fd0
run_status: implemented
ac_pass_count: 6
ac_fail_count: 0
preserve_list_post_run_count: 0   # no PRESERVE-list file outside go.mod + 9 workflows touched
l44_pre_commit_fetch: "git fetch origin main → origin/main...HEAD = 0 3 (local ahead, clean)"
l44_post_push_fetch: n/a   # orchestrator handles push after sync-phase (Tier S main-direct, no PR)
new_warnings_or_lints_introduced: false
cross_platform_build:
  darwin_arm64: pass   # go build ./... exit 0 under go1.26.4
  note: "toolchain bump is platform-agnostic; no syscall/build-tag code touched"
total_run_phase_files: 6   # go.mod + ci.yml + claude.yml + codeql.yml + release-pr-multi-os.yml + release.yml
m1_to_mN_commit_strategy: "single M1 commit (Tier S minimal; all of M1+M2+M3 in one run-phase commit)"
```
