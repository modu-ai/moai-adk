# SPEC-GO-DEPS-UPDATE-001 — Acceptance Criteria

> SSOT for the tester gate. 9 ACs. Tier S. All verification commands run from project root
> under go1.26.4.

## D. Acceptance Criteria Matrix

| AC | Requirement | Verification Command | Expected |
|----|-------------|----------------------|----------|
| AC-GDU-001 | REQ-GDU-001 — Phase 1 (patch) applied | `grep -E 'charmbracelet/x/powernap v0\.1\.6|go-playground/validator/v10 v10\.30\.3|mattn/go-runewidth v0\.0\.24' go.mod` | 3 matches |
| AC-GDU-002 | REQ-GDU-002 — Phase 2 (x/* minor) applied | `grep 'golang.org/x/sys v0.45' go.mod` AND `go list -m golang.org/x/crypto golang.org/x/net golang.org/x/tools golang.org/x/mod` shows bumped minors | x/sys ≥ v0.45.0; 4 indirect x/* at latest minor |
| AC-GDU-003 | REQ-GDU-003 — clean tidy tree | `go mod tidy && git diff --quiet go.mod go.sum` after a SECOND `go mod tidy` | second tidy is a no-op (no diff) |
| AC-GDU-004 | REQ-GDU-004 — cross-platform build | `go build ./...` ; `GOOS=windows GOARCH=amd64 go build ./...` ; `GOOS=linux GOARCH=amd64 go build ./...` | all exit 0 |
| AC-GDU-005 | REQ-GDU-005 — tests, no NEW regression | `go test ./...` | pass except §D.1 known pre-existing failures |
| AC-GDU-006 | REQ-GDU-006 — vuln floor preserved | `govulncheck ./...` | affecting third-party count == 0 (unchanged from baseline) |
| AC-GDU-007 | REQ-GDU-007 — `.go`-edit capability gate | `git diff --name-only` ; IF any `.go` file is listed THEN confirm `progress.md` records the file + the API-change reason | go.mod/go.sum-only (expected) OR (if any `.go` changed) a matching note exists in progress.md |
| AC-GDU-008 | REQ-GDU-008 — scope guard (diff bounded) | `git diff --name-only` | only `go.mod` + `go.sum` (plus any `.go` edit already accounted for by AC-GDU-007) |
| AC-GDU-009 | REQ-GDU-008 — no new dependency | compare `go mod edit -json \| jq '.Require[].Path'` pre vs post | no net-new module path in `require` (transitive tidy churn allowed) |

## D.1 Known pre-existing non-regression failures (excluded from AC-GDU-005)

A tester MUST exclude these from the AC-GDU-005 gate — they exist on the pre-bump baseline
and are NOT caused by the dependency update:

- `internal/hook` — `TestHookWrapper_MoaiBinaryFallback`, `TestHookWrapper_ValidJSON`
  (~5s timing-flaky; pass on retry).
- `internal/template` — `TestOutputStylesTemplateLiveParity` (einstein.md drift, unrelated
  to dependencies).

Verification that an AC-GDU-005 failure is a KNOWN failure (not a regression):

```bash
# Run the full suite, then confirm any failure is in the §D.1 known-failure set ONLY.
# NOTE: `--- FAIL:` is NOT anchored to line start — Go sub-test failures are indented
# (`    --- FAIL:`), so anchoring would miss them. The `^FAIL` alternative stays anchored
# (package-level FAIL lines are always at column 0).
go test ./... 2>&1 | grep -E '^FAIL|--- FAIL:' \
  | grep -vE 'TestHookWrapper_MoaiBinaryFallback|TestHookWrapper_ValidJSON|TestOutputStylesTemplateLiveParity'
# Expected: no output (no failure outside the known set). Any line here is a NEW regression.
```

## D.2 Definition of Done

- [ ] AC-GDU-001 through AC-GDU-009 all PASS (AC-GDU-005 evaluated against the §D.1
      exclusion list).
- [ ] `git diff` is limited to `go.mod` + `go.sum` (scope guard).
- [ ] No `charmbracelet/x/*` indirect module was manually bumped (Phase 3 stayed excluded).
- [ ] Run-phase commit is main-direct, Conventional Commits, no PR, specific-path `git add`.
- [ ] If any `.go` source edit was required by a minor bump (expected none), it is recorded
      in progress.md per REQ-GDU-007.

## D.3 Edge cases

- **EC-1 — tidy proposes a charmbracelet/x/* indirect bump**: if `go mod tidy` itself
  surfaces a `charmbracelet/x/*` change driven by the legitimate patch/minor resolution
  (NOT a manual edit), that is acceptable as long as it is a natural consequence of tidy
  and not a force-bump. The AC-GDU-008 scope guard still holds (diff stays in go.mod/go.sum).
- **EC-2 — govulncheck newly flags a bumped module**: AC-GDU-006 FAILS. The bump that
  introduced the affecting vuln must be reverted or pinned; the maintenance update must not
  regress the 0-affecting floor.
- **EC-3 — windows/linux cross-build fails but host build passes**: AC-GDU-004 FAILS. A
  bumped dependency introduced a platform-specific break; investigate the offending module.
