# Progress — SPEC-GO-TOOLCHAIN-SEC-001

## Status: draft (plan-phase)

Tier S (minimal). Hybrid Trunk 1-person OSS → main-direct, no PR.

## Phase Tracker

| Phase | Status | Owner | Commit |
|-------|--------|-------|--------|
| Plan | done | manager-spec | (this plan-phase commit) |
| Run (M1–M3) | pending | manager-develop | — |
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

## Run-phase evidence (to be filled by manager-develop)

6 ACs total (AC-GTS-001..006). Capture AC-GTS-006 evidence FIRST — it gates AC-GTS-001.

- [ ] AC-GTS-006 — `go version | grep -qE 'go1\.26\.(4|[5-9]|[1-9][0-9])' && echo GUARD-OK` (expect `GUARD-OK`; gates AC-GTS-001)
- [ ] AC-GTS-001 — `govulncheck ./...` output (expect 0 affecting; accepted only after AC-GTS-006 PASS)
- [ ] AC-GTS-002 — `go build ./...` + `go test ./...` exit codes (expect 0)
- [ ] AC-GTS-003 — `grep -rn 'go-version: *"1.26"' .github/workflows/` (expect 0 matches) + `go.mod` `go` ≥ 1.26.4
- [ ] AC-GTS-004 — `git diff go.mod` (directive-only; no `require` change)
- [ ] AC-GTS-005 — `awk '/^go /{g=$2} /^toolchain /{t=$2} END{exit (t=="" || t=="go"g)?0:1}' go.mod` (expect exit 0)
