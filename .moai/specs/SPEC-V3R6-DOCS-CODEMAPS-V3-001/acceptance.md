# Acceptance — SPEC-V3R6-DOCS-CODEMAPS-V3-001

> Falsifiable acceptance criteria for the codemaps SSOT refresh +
> docs-truth checklist. Each criterion is independently verifiable via a
> concrete command. Severity, traceability, and Definition of Done at the
> bottom. iter-2 note: AC-CM-000 (docs-truth existence) added; AC-CM-001
> re-targeted from a non-existent `architecture.md` to the skill's real
> `overview.md` + `modules.md` outputs; AC-CM-005 rewritten to span the
> full `internal/cli/` tree and the full 17-file `/moai` skill set;
> AC-CM-006 expanded to the full GLM tier-models table.

## §D. Acceptance Criteria

### AC-CM-000 — docs-truth.md exists at the canonical path

**Severity**: MUST-PASS.

**Given** the docs cohort M2 milestone has run,
**When** a reviewer checks for the docs-truth checklist file,
**Then** the file `.moai/project/codemaps/docs-truth.md` exists and is
non-empty.

**Falsifiable via**:

```bash
test -s .moai/project/codemaps/docs-truth.md
```

Exit 0 ⇒ PASS.

**REQ traceability**: REQ-CM-003.

### AC-CM-001 — Refreshed codemaps collectively cover all capability layers

**Severity**: MUST-PASS.

**Given** the `/moai codemaps` skill has been run (with `--force`) against
the v3.0 codebase,
**When** a reviewer opens `.moai/project/codemaps/overview.md` AND
`.moai/project/codemaps/modules.md` together,
**Then** the UNION of the two files contains a mention/section for each
of the 10 capability layers — CLI surface, Lifecycle, Harness, Hooks,
Templates, Quality, Multi-LLM, Web, Governance, Foundation. (The skill
does NOT emit a single `architecture.md`; coverage is judged collectively
across its real outputs per §B.1.1 decision (a).)

**Falsifiable via**:

```bash
# overview.md and modules.md must both exist (refreshed)
test -s .moai/project/codemaps/overview.md
test -s .moai/project/codemaps/modules.md
# Each of the 10 capability layers must appear in AT LEAST ONE of the two files
for layer in "CLI" "Lifecycle" "Harness" "Hook" "Template" "Quality" "GLM\|Multi-LLM\|multi-llm" "Web" "Governance" "Foundation"; do
  grep -qiE "${layer}" .moai/project/codemaps/overview.md \
    || grep -qiE "${layer}" .moai/project/codemaps/modules.md \
    || { echo "MISSING capability layer across overview.md + modules.md: ${layer}"; exit 1; }
done
```

Exit 0 ⇒ PASS. Any "MISSING capability layer" ⇒ FAIL.

**REQ traceability**: REQ-CM-001.

### AC-CM-002 — docs-truth lists exactly the 8 retained agents

**Severity**: MUST-PASS.

**Given** the docs-truth checklist has been authored,
**When** a reviewer reads the Agent catalog section,
**Then** the section lists exactly 8 agents with correct class labels:
4 Manager (`manager-spec`, `manager-develop`, `manager-docs`,
`manager-git`), 2 Evaluator (`plan-auditor`, `sync-auditor`), 1 Builder
(`builder-harness`), 1 Anthropic built-in (`Explore`).

**Falsifiable via**:

```bash
# Ground-truth: 7 MoAI-custom files on disk + 1 built-in Explore
ls -1 .claude/agents/moai/*.md | wc -l     # → 7
# docs-truth must enumerate the same 7 + Explore (8 total)
grep -cE 'manager-spec|manager-develop|manager-docs|manager-git' .moai/project/codemaps/docs-truth.md  # → ≥4
grep -cE 'plan-auditor|sync-auditor' .moai/project/codemaps/docs-truth.md                              # → ≥2
grep -cE 'builder-harness' .moai/project/codemaps/docs-truth.md                                        # → ≥1
grep -cE '\bExplore\b' .moai/project/codemaps/docs-truth.md                                            # → ≥1
# Verify no archived-agent leakage
grep -cE 'manager-strategy|manager-quality|manager-brain|manager-project|claude-code-guide|researcher|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-refactoring' .moai/project/codemaps/docs-truth.md  # → 0
```

All counts match ⇒ PASS. Any mismatch ⇒ FAIL.

**REQ traceability**: REQ-CM-004.

### AC-CM-003 — docs-truth status enum matches `internal/spec/status.go`

**Severity**: MUST-PASS.

**Given** the docs-truth checklist has been authored,
**When** a reviewer reads the Status enum section,
**Then** the section records the exact lowercase 8-value set from
`internal/spec/status.go` `ValidStatuses`: `draft`, `planned`,
`in-progress`, `implemented`, `completed`, `superseded`, `archived`,
`rejected`.

**Falsifiable via**:

```bash
for status in draft planned in-progress implemented completed superseded archived rejected; do
  grep -q "\`${status}\`" .moai/project/codemaps/docs-truth.md \
    || { echo "MISSING status: ${status}"; exit 1; }
done
# Cross-verify count against source
test "$(grep -cE '\"draft\"|\"planned\"|\"in-progress\"|\"implemented\"|\"completed\"|\"superseded\"|\"archived\"|\"rejected\"' internal/spec/status.go)" -ge 8
```

Exit 0 ⇒ PASS.

**REQ traceability**: REQ-CM-005.

### AC-CM-004 — docs-truth lists the 12 required frontmatter fields

**Severity**: MUST-PASS.

**Given** the docs-truth checklist has been authored,
**When** a reviewer reads the Frontmatter schema section,
**Then** the section records the exact 12 required fields from
`internal/spec/lint.go` `FrontmatterSchemaRule` `required` slice: `id`,
`title`, `version`, `status`, `created`, `updated`, `author`, `priority`,
`phase`, `module`, `lifecycle`, `tags`.

**Falsifiable via**:

```bash
for field in id title version status created updated author priority phase module lifecycle tags; do
  grep -q "\`${field}\`" .moai/project/codemaps/docs-truth.md \
    || { echo "MISSING field: ${field}"; exit 1; }
done
# Cross-verify against lint.go source — count entries in the required slice.
# The slice lives in FrontmatterSchemaRule.Check (~lines 586-602) as a
# []struct{...}; each entry is `{"<field>", fm.<Field>}`.
test "$(awk '/^func \(r \*FrontmatterSchemaRule\) Check/,/^func [a-zA-Z]/' internal/spec/lint.go \
      | grep -cE '^\s*\{"[a-z]+", fm\.')" -eq 12
```

Exit 0 ⇒ PASS.

**REQ traceability**: REQ-CM-006.

### AC-CM-005 — docs-truth CLI surface cross-verified (human-facing verbs + full skill set)

**Severity**: MUST-PASS.

**Given** the docs-truth checklist has been authored,
**When** a reviewer reads the CLI subcommand surface section,
**Then** the section enumerates the HUMAN-FACING `moai` terminal verb
names (e.g. `init`, `update`, `version`, `glm`, `cc`, `cg`, `web`,
`session`, `spec`, `harness`, `worktree`, `hook`, `agent`, `research`,
`workflow`, `migrate`, `constitution`, `statusline`, `telemetry`,
`profile`, `doctor`, `lsp`, `github`, `brain`, `loop`, `mx`, `clean`,
…) cross-verified against the FULL `internal/cli/*.go` tree (where
`rootCmd.AddCommand` spans 26 files, not only `root.go`) AND the
rendered `moai --help` verb list, AND enumerates the COMPLETE `/moai`
Claude Code skill set (17 files in `.claude/commands/moai/`:
brain, clean, codemaps, coverage, design, e2e, feedback, fix, gate,
harness, loop, mx, plan, project, review, run, sync). The checklist
lists human-facing verb names, NOT internal Go identifiers like
`worktree.WorktreeCmd`.

**Falsifiable via**:

```bash
# (1) Cross-verify: grep the FULL internal/cli/ tree for AddCommand
#     (spans 26 files, not only root.go). The regex captures BOTH bare
#     identifiers AND paren'd constructor calls (NewAstGrepCmd(),
#     newConstitutionCmd(), newStateCmd(), newCleanCmd(),
#     newHarnessRouterCmd()).
test "$(grep -rnE '\.AddCommand\(' internal/cli/ | grep -v '_test.go' | wc -l)" -ge 30
# (2) docs-truth MUST list human-facing verb names — sanity-check a
#     representative subset that lives in files OTHER than root.go
#     (these are the verbs most easily missed by a root.go-only grep).
for verb in glm cc cg web session spec harness worktree hook agent research workflow; do
  grep -qiE "\b${verb}\b" .moai/project/codemaps/docs-truth.md \
    || { echo "MISSING human-facing verb in docs-truth: ${verb}"; exit 1; }
done
# (3) /moai skill-set parity: docs-truth MUST list all 17 command files.
#     Use ls parity, NOT a hardcoded allowlist (a hardcoded allowlist
#     silently passes while omitting codemaps/clean/mx/coverage/e2e/
#     review/brain/gate/harness).
expected=$(ls -1 .claude/commands/moai/*.md | xargs -n1 basename | sed 's/\.md$//')
for skill in ${expected}; do
  grep -qiE "/moai ${skill}\b|\`${skill}\`" .moai/project/codemaps/docs-truth.md \
    || { echo "MISSING /moai skill in docs-truth: ${skill}"; exit 1; }
done
# (4) Sanity: command-file count is exactly 17
test "$(ls -1 .claude/commands/moai/*.md | wc -l)" -eq 17
```

Exit 0 ⇒ PASS.

**REQ traceability**: REQ-CM-007.

### AC-CM-006 — docs-truth GLM→Claude tier mapping (full tier-models table)

**Severity**: MUST-PASS.

**Given** the docs-truth checklist has been authored,
**When** a reviewer reads the GLM→Claude tier mapping section,
**Then** the section records the FULL tier-models table from
`internal/config/defaults.go` lines 40-57 — `DefaultGLMBaseURL`,
`DefaultGLMHigh`, `DefaultGLMMedium`, `DefaultGLMLow`,
`DefaultGLMSonnet`, `DefaultGLMHaiku`, `DefaultGLMOpus` — with the
glm-5.2[1m] activation reflected in the High and Opus slots.

**Falsifiable via**:

```bash
grep -q 'glm-5\.2\[1m\]' .moai/project/codemaps/docs-truth.md         # reflects the recent activation
grep -q 'api\.z\.ai/api/anthropic' .moai/project/codemaps/docs-truth.md
# Full tier-models table — Medium/Low/Sonnet/Haiku MUST appear too
grep -qiE 'DefaultGLMMedium.*glm-4\.7' .moai/project/codemaps/docs-truth.md
grep -qiE 'DefaultGLMLow.*glm-4\.5-air' .moai/project/codemaps/docs-truth.md
grep -qiE 'DefaultGLMSonnet.*glm-4\.7' .moai/project/codemaps/docs-truth.md
grep -qiE 'DefaultGLMHaiku.*glm-4\.5-air' .moai/project/codemaps/docs-truth.md
# Cross-verify against source
grep -q 'DefaultGLMHigh.*=.*"glm-5\.2\[1m\]"' internal/config/defaults.go
grep -q 'DefaultGLMOpus.*=.*"glm-5\.2\[1m\]"' internal/config/defaults.go
grep -q 'DefaultGLMMedium.*=.*"glm-4\.7"' internal/config/defaults.go
grep -q 'DefaultGLMLow.*=.*"glm-4\.5-air"' internal/config/defaults.go
grep -q 'DefaultGLMSonnet.*=.*"glm-4\.7"' internal/config/defaults.go
grep -q 'DefaultGLMHaiku.*=.*"glm-4\.5-air"' internal/config/defaults.go
```

Exit 0 ⇒ PASS.

**REQ traceability**: REQ-CM-008.

### AC-CM-007 — codemaps neutrality (zero internal-token leakage)

**Severity**: MUST-PASS.

**Given** the codemaps directory has been refreshed,
**When** a reviewer greps `.moai/project/codemaps/` for internal SPEC
tokens,
**Then** the only matches are self-references (docs-truth.md citing THIS
SPEC's own ID `SPEC-V3R6-DOCS-CODEMAPS-V3-001`); there are ZERO
`REQ-[A-Z]+-[0-9]+` tokens, ZERO `Audit [0-9]+ Finding` citations, and
ZERO references to OTHER SPEC IDs.

**Falsifiable via**:

```bash
# REQ tokens must be absent entirely
! grep -rE 'REQ-[A-Z]+-[0-9]+' .moai/project/codemaps/
# Audit-citation style must be absent entirely
! grep -rE 'Audit [0-9]+ Finding' .moai/project/codemaps/
# Other SPEC IDs must be absent (self-reference to THIS SPEC is allowed inside docs-truth.md)
test "$(grep -rEoh 'SPEC-[A-Z0-9-]+-[0-9]{3}' .moai/project/codemaps/ | sort -u | wc -l)" -le 1
test "$(grep -rEoh 'SPEC-[A-Z0-9-]+-[0-9]{3}' .moai/project/codemaps/ | sort -u)" = "SPEC-V3R6-DOCS-CODEMAPS-V3-001" \
  -o "$(grep -rEoh 'SPEC-[A-Z0-9-]+-[0-9]{3}' .moai/project/codemaps/ | sort -u | wc -l)" -eq 0
```

All sub-checks pass ⇒ PASS.

**REQ traceability**: REQ-CM-009.

### AC-CM-008 — Zero new build/vet/lint regressions

**Severity**: MUST-PASS.

**Given** the pre-SPEC baseline passed `go build ./...`,
`go vet ./...`, and `golangci-lint run`,
**When** the same three commands are re-run after the codemaps refresh,
**Then** each command shows zero NEW failures relative to the pre-SPEC
baseline (`.moai/project/codemaps/` is additive Markdown output that the
Go toolchain does not compile).

**Falsifiable via**:

```bash
go build ./...           # exit 0, no new errors
go vet ./...             # exit 0, no new warnings
golangci-lint run        # baseline-finding-count preserved
```

Exit 0 on all three ⇒ PASS.

**REQ traceability**: REQ-CM-010, REQ-CM-011.

### AC-CM-009 — Named capability-anchor package set covered (SHOULD)

**Severity**: SHOULD-PASS (not a merge blocker, but expected).

**Given** the codemaps have been refreshed,
**When** a reviewer reads `.moai/project/codemaps/modules.md`,
**Then** the named capability-anchor package set (spec, cli, config,
statusline, hook, template, harness, session, web — per REQ-CM-002
selection rule) is covered.

**Falsifiable via**:

```bash
found=0
for pkg in spec cli config statusline hook template harness session web; do
  grep -qiE "(^|[^a-z])${pkg}(/|[^a-z]|$)" .moai/project/codemaps/modules.md && found=$((found + 1))
done
test "${found}" -ge 9
```

`found ≥ 9` ⇒ PASS.

**REQ traceability**: REQ-CM-002.

## §D.1 Severity Classification

- **MUST-PASS** (AC-CM-000 through AC-CM-008): merge blocker. Any
  failure prevents `draft → in-progress` progression and requires
  remediation before the next milestone.
- **SHOULD-PASS** (AC-CM-009): expected but not a blocker. Failure
  triggers a documented debt entry, not a remediation hold.

## §D.2 Traceability Matrix

| AC | REQ | Milestone |
|----|-----|-----------|
| AC-CM-000 | REQ-CM-003 | M2, M3 |
| AC-CM-001 | REQ-CM-001 | M1, M3 |
| AC-CM-002 | REQ-CM-004 | M2, M3 |
| AC-CM-003 | REQ-CM-005 | M2, M3 |
| AC-CM-004 | REQ-CM-006 | M2, M3 |
| AC-CM-005 | REQ-CM-007 | M2, M3 |
| AC-CM-006 | REQ-CM-008 | M2, M3 |
| AC-CM-007 | REQ-CM-009 | M3 |
| AC-CM-008 | REQ-CM-010, REQ-CM-011 | M3 |
| AC-CM-009 | REQ-CM-002 | M1 |

> Note: AC-CM-000 (docs-truth.md existence, REQ-CM-003) and AC-CM-010
> are the same criterion referenced from two angles — AC-CM-000 is the
> existence check; plan.md M2/M3 also bind AC-CM-010 as the same
> existence gate at milestone level. Both names refer to the single
> REQ-CM-003 existence requirement; no separate AC-CM-010 row is
> emitted to avoid double-counting.

## §D.3 Indirect Verification (Where direct grep is insufficient)

- **Docs-truth fact freshness**: AC-CM-002..006 each include a
  cross-verification against the primary source file (`ls`, `grep` on
  `internal/spec/status.go`, etc.). This guards against docs-truth
  recording stale facts that have drifted from source since M2.
- **Insight-vs-extraction labeling**: REQ-CM-002 + plan.md §B require
  insight-level prose to be labeled. Verified indirectly via the
  `modules.md` structure (each package's "## Public Surface" and
  "## Dependencies" subsections carry deterministic content; an
  "## Architectural Insight" subsection carries the augmented judgment).
  The skill does NOT emit `architecture.md`, so labeling is judged
  within `modules.md` / `overview.md`.

## §D.4 Closure Gates (Definition of Done)

A SPEC is "done" at run-phase close when ALL of the following hold:

- [ ] All MUST-PASS ACs (AC-CM-000..008) PASS with verbatim command
      output recorded in `progress.md` §E.2.
- [ ] AC-CM-009 (SHOULD) either PASSes OR carries a documented debt
      entry with rationale.
- [ ] `progress.md` §E.2..§E.5 headings are populated per their owning
      phase (run/sync/Mx).
- [ ] The single commit on `main` is attributable to manager-develop
      via the `Authored-By-Agent:` trailer (per the Status Transition
      Ownership Matrix).
- [ ] `moai spec audit` on this SPEC returns zero MUST-FIX findings
      (INFO-level `EraAutoDetected` is acceptable ONLY if `era: V3R6` is
      somehow missing; with `era: V3R6` present, even that INFO should
      be suppressed).

## §D.5 Forward-Looking Checks (for later cohort SPECs)

These are NOT merge blockers for this SPEC but are checks the later
README / DOCSITE / COVERAGE / i18n SPECs SHOULD perform before closing:

- [ ] The later SPEC cites `.moai/project/codemaps/docs-truth.md` AND
      re-verifies the cited fact against the primary source (per
      REQ-CM-012 — docs-truth is a navigation aid, not a new SSOT).
- [ ] The later SPEC does NOT introduce a contradicting agent count,
      status value, or frontmatter field that differs from docs-truth
      without an explicit rationale.
- [ ] If the later SPEC discovers that a docs-truth fact has drifted
      from its primary source, the later SPEC updates docs-truth FIRST
      (in a separate commit) before proceeding with its own rewrite.
