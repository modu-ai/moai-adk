# SPEC-SEC-DEEPSCAN-001 — Acceptance Criteria

All criteria are **observable** — verified by grep / diff / file-existence / test-run, never by subjective judgment. Because the run-phase deliverable is shipped markdown (a playbook) plus an argument-hint edit, most ACs assert the presence and reachability of specific playbook content in BOTH the local and template trees.

## §D.1 AC-to-REQ Traceability Matrix

| AC | REQ | Gist | Verification kind |
|----|-----|------|-------------------|
| AC-SDS-001 | REQ-SDS-001 | `--deep` flag documented in `review.md` (both trees) | grep |
| AC-SDS-002 | REQ-SDS-002 | No `/moai security` subcommand / alias reintroduced | grep (absence) |
| AC-SDS-003 | REQ-SDS-003 | Job menu maps to `--repo` / diff / `--commit` / `--patch` | grep |
| AC-SDS-004 | REQ-SDS-004 | `--patch` opt-in, default-off documented | grep |
| AC-SDS-010 | REQ-SDS-010 | Six phases documented in order | grep -n ordering |
| AC-SDS-011 | REQ-SDS-011 | Hunt-phase `Skill()` injection of 4 ref skills | grep |
| AC-SDS-012 | REQ-SDS-012 | Recon/hunt/verify phases marked read-only | grep |
| AC-SDS-020 | REQ-SDS-020 | 3-voter panel + 2-of-3 quorum admission | grep |
| AC-SDS-021 | REQ-SDS-021 | Non-unanimous → confidence capped at "medium" | grep |
| AC-SDS-022 | REQ-SDS-022 | Voter independence / REFUTE-skewed | grep |
| AC-SDS-023 | REQ-SDS-023 | Rejected candidates excluded from confirmed body | grep |
| AC-SDS-030 | REQ-SDS-030 | Scratch-clone drafting via `Agent(isolation:"worktree")` | grep |
| AC-SDS-031 | REQ-SDS-031 | Independent reviewer 3-claim vouch | grep |
| AC-SDS-032 | REQ-SDS-032 | Vouch-failure → note instead of patch | grep |
| AC-SDS-033 | REQ-SDS-033 | Patches never auto-applied; 1 finding=1 patch=1 PR | grep |
| AC-SDS-040 | REQ-SDS-040 | Results dir under `.moai/reports/`, not `.moai/specs/` | grep |
| AC-SDS-041 | REQ-SDS-041 | Human report schema (F-IDs + 5 fields) documented | grep |
| AC-SDS-042 | REQ-SDS-042 | JSONL one-finding-per-line documented | grep |
| AC-SDS-043 | REQ-SDS-043 | Revision stamp (commit + effort + working-tree bool) | grep |
| AC-SDS-044 | REQ-SDS-044 | Results dir ships own `.gitignore` | grep |
| AC-SDS-050 | REQ-SDS-050 | Dynamic Workflows prerequisite + signal documented | grep |
| AC-SDS-051 | REQ-SDS-051 | 3-rung degradation ladder documented | grep |
| AC-SDS-052 | REQ-SDS-052 | Degraded path preserves 2-of-3 quorum | grep |
| AC-SDS-060 | REQ-SDS-060 | Template ↔ local byte-parity for workflow skill | diff |
| AC-SDS-061 | REQ-SDS-061 | Neutrality: no internal SPEC-ID/date/SHA/Go-privilege | grep (absence) + test |
| AC-SDS-062 | REQ-SDS-062 | No new agent / no Go change / no shipped `.js` | test + ls |
| AC-SDS-063 | REQ-SDS-063 | No agent-prompts-user directive in the playbook | grep (absence) |

## §D.2 Given-When-Then Scenarios

### Scenario 1 — Deep scan reachable and security-focused by default
- **Given** the shipped `review` workflow skill and command wrapper (both trees),
- **When** a user types `/moai review --deep`,
- **Then** the `argument-hint` advertises `--deep`, the `review.md` playbook contains a `--deep` mode section, and a bare `--deep` is documented as security-focused (composes with `--security`).
- **Verify**:
  ```bash
  grep -q '\-\-deep' internal/template/templates/.claude/commands/moai/review.md.tmpl && echo "hint OK"
  grep -qiE '^#+ .*--deep' .claude/skills/moai/workflows/review.md && echo "playbook OK"
  ```

### Scenario 2 — A candidate finding must pass the 2-of-3 adversarial panel before it is reported
- **Given** the `--deep` playbook's adversarial-verification phase,
- **When** a candidate finding is produced by the hunt phase,
- **Then** the playbook requires a 3-voter panel (REACHABILITY / IMPACT / DEFENSES) with a 2-of-3 quorum for admission, and a non-unanimous verdict caps the reported confidence at "medium".
- **Verify**:
  ```bash
  grep -qiE 'REACHABILITY|IMPACT|DEFENSES' .claude/skills/moai/workflows/review.md
  grep -qiE '2[- ]of[- ]3|2/3|quorum' .claude/skills/moai/workflows/review.md
  grep -qiE 'non-unanimous.*medium|cap.*confidence.*medium' .claude/skills/moai/workflows/review.md
  ```

### Scenario 3 — Patch drafting is reviewer-vouched and never auto-applied
- **Given** `--patch` present and a confirmed finding,
- **When** a patch is drafted in a scratch clone,
- **Then** an independent reviewer vouches for the 3 claims (only-this-finding / no-new-vuln / behavior-unchanged); on vouch-failure a note is emitted instead of a patch; the patch is never applied/committed/pushed (one finding = one patch = one PR, user applies via `git apply`).
- **Verify**:
  ```bash
  grep -qiE 'isolation.*worktree|scratch clone' .claude/skills/moai/workflows/review.md
  grep -qiE 'independent reviewer|vouch' .claude/skills/moai/workflows/review.md
  grep -qiE 'never.*(apply|applied)|git apply' .claude/skills/moai/workflows/review.md
  ```

### Scenario 4 — Results are a self-contained, gitignore-protected report
- **Given** a completed deep scan,
- **When** it writes outputs,
- **Then** it writes to a timestamped `.moai/reports/security-deepscan-<ts>/` directory containing a human `.md` (F-IDs + impact/exploit/severity/confidence/recommendation), a `.jsonl` (one finding per line), a revision stamp (scanned commit + effort tier + working-tree-included), and its own `.gitignore`.
- **Verify**:
  ```bash
  grep -qE '\.moai/reports/security-deepscan' .claude/skills/moai/workflows/review.md
  grep -qiE 'jsonl|one finding per line' .claude/skills/moai/workflows/review.md
  grep -qiE 'revision stamp|scanned commit|working tree' .claude/skills/moai/workflows/review.md
  grep -qE '\.gitignore' .claude/skills/moai/workflows/review.md
  ```

### Scenario 5 — Graceful degradation when Dynamic Workflows are unavailable
- **Given** a runtime with Dynamic Workflows disabled or below v2.1.154,
- **When** `/moai review --deep` is invoked,
- **Then** the playbook degrades to a Mode-4 bounded parallel fan-out (preserving the 2-of-3 quorum) and, when even that is not viable, to single-pass `/moai review --security` + native `/security-review` — never a hard failure.
- **Verify**:
  ```bash
  grep -qiE 'v2\.1\.154|dynamic workflows.*(require|unavailable)' .claude/skills/moai/workflows/review.md
  grep -qiE 'degrad|fallback' .claude/skills/moai/workflows/review.md
  grep -qiE 'security-review' .claude/skills/moai/workflows/review.md
  ```

## §D.3 Severity Classification

| Severity | ACs | Gate effect |
|----------|-----|-------------|
| MUST-PASS (blocks close) | AC-SDS-001, -002, -010, -020, -021, -030, -031, -033, -040, -044, -060, -061, -062 | Any FAIL blocks sync-close |
| SHOULD-PASS | AC-SDS-003, -004, -011, -012, -022, -023, -032, -041, -042, -043, -050, -051, -052, -063 | FAIL → PASS-WITH-DEBT + tracked |

## §D.4 Edge Cases

- **EC-1 — Zero confirmed findings**: the scan completes cleanly, writes a results dir stating "0 confirmed findings", and drafts no patches. The playbook MUST document the empty-result path (no error, no empty-patch).
- **EC-2 — Confirmed finding but `--patch` absent**: report is written; no patch phase runs (REQ-SDS-004). Verify the playbook stops after phase 5 without `--patch`.
- **EC-3 — Reviewer cannot vouch**: a note (not a patch) is emitted for that finding; other findings' patches are unaffected (REQ-SDS-032).
- **EC-4 — Non-unanimous panel on a high-severity finding**: severity may remain high, but stated CONFIDENCE is capped at "medium" (severity and confidence are independent axes — the cap binds confidence only).
- **EC-5 — Dynamic Workflows disabled mid-availability-check**: the availability signal is checked BEFORE launch; degradation is chosen up front, not mid-run (workflow agents cannot prompt the user mid-run).
- **EC-6 — Large repo, hundreds of candidate findings**: the primary Workflow path (up to 16 concurrent / 1000-total) handles scale; the degraded Mode-4 path batches findings within the 3-5 concurrent ceiling. Neither path drops the quorum.

## §D.5 Quality Gates (TRUST 5, scaled to a markdown-shipping SPEC)

- **Tested**: every MUST-PASS AC has a runnable grep/diff/test command with expected output (this file). No Go coverage delta (no Go change).
- **Readable**: playbook prose is English, 16-language-neutral, and scannable (section headings per phase).
- **Unified**: template ↔ local byte-parity (`diff` clean) for the workflow skill; command wrapper rendered consistently.
- **Secured**: the shipped content is defensive-only; no offensive tooling; no secrets; AskUserQuestion boundary preserved.
- **Trackable**: Conventional Commits per milestone (`feat(SPEC-SEC-DEEPSCAN-001): M<N> …`); PR per Tier L Route B.

## §D.6 Definition of Done

- [ ] All 27 ACs evaluated; every MUST-PASS AC = PASS.
- [ ] `review.md` template ↔ local `diff` clean (byte-parity).
- [ ] Command `argument-hint` includes `--deep` and `--patch` (rendered + `.tmpl`).
- [ ] Neutrality grep clean on edited template files (no internal SPEC-ID / date / SHA; no Go-privileging); `internal_content_leak_test.go` + `template-neutrality-check.yaml` pass.
- [ ] `go build ./...` exit 0; `go test ./internal/template/...` PASS (catalog counts unchanged — no new agent).
- [ ] `moai spec lint --strict` for this SPEC → 0 errors (repo-wide residual debt reported separately, not attributed to this SPEC).
- [ ] No `/moai security` revival; no shipped `.js`; no Go runtime change.
- [ ] Out of Scope §C intact; SPEC-2 layering boundary documented.

## §D.7 Indirect / Forward-Looking Checks

- **Indirect**: because the deep scan is runtime-constructed (no shipped script), the "scan actually runs" behavior is NOT verifiable by a static grep — it is verified indirectly by (a) the playbook completeness ACs above and (b) a manual smoke invocation recorded in progress.md §E.2 during run-phase (invoke `/moai review --deep --repo` on a fixture and confirm a results dir with the required artifacts is produced). This smoke check is the run-phase evidence for REQ-SDS-010/-040/-041/-042/-043/-044 beyond the prose ACs.
- **Forward-looking**: SPEC-2 (always-on guardian) will consume the same results-dir schema; the JSONL schema defined in M3 SHOULD be forward-compatible with SPEC-2's inline findings (recorded as a design note, not gated here).
