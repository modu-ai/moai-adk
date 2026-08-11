# acceptance.md — SPEC-FACTORY-BOOTSTRAP-001

> Acceptance criteria in **Given-When-Then** format. Each `AC-FB-XXX` is binary-testable (a command exits 0 / 1, a file exists / does not, a string is present / absent, a test passes / fails). Criteria are grouped by the REQ they verify; severity (MUST / SHOULD) and traceability live in §D and §D.1.

---

## §A. Verification Approach

Every criterion names either:

- **a grep / file-existence check** against a source file in this worktree (a substring is present or absent at a named path), or
- **a Go test** in `internal/cli/` or `internal/hook/` (a `*_test.go` function exits 0), or
- **a build artifact** (`make build` exits 0 and the binary contains/omits a string), or
- **a docs-site artifact** (a path exists under `docs-site/content/{en,ko,ja,zh}/` and a `main.yaml` entry exists).

The baseline-attribution (verification-claim-integrity.md §2) for every grep-based criterion is: the command `grep -nE '<pattern>' <path>` run at HEAD of `feat/factory-bootstrap-guidance` after the milestone lands.

---

## §B. AC Enumeration

### AC-FB-001 — `-f` alone classifies as lead

**Given** the launcher `moai cc`, **When** the operator runs `moai cc -f` (no `--name`), **Then** the launched session's environment carries `MOAI_FACTORY=<non-empty>` and `MOAI_FACTORY_ID=<non-empty>`, and `MOAI_FACTORY_LABEL` is unset.

### AC-FB-002 — `-f --name <companion>` classifies as companion

**Given** the launcher `moai cc`, **When** the operator runs `moai cc -f --name run-abc123`, **Then** the launched session's environment carries `MOAI_FACTORY_LABEL=run-abc123` and `MOAI_FACTORY` is unset, `MOAI_FACTORY_ID` is unset, `MOAI_FACTORY_SPEC` is unset.

### AC-FB-003 — `--name <non-companion>` alone is a no-op

**Given** the launcher `moai cc`, **When** the operator runs `moai cc --name mysession` (mysession does not match `<role>-<run-id>`), **Then** no `MOAI_FACTORY*` environment variable is set and the existing `--name` passthrough behavior is unchanged (verified by the prior-art `factory_bootstrap_test.go` no-op case continuing to pass).

### AC-FB-004 — `-f --name <non-companion>` is lead

**Given** the launcher `moai cc`, **When** the operator runs `moai cc -f --name mysession` (mysession does not match `<role>-<run-id>`), **Then** the session is classified as a **lead** (because the `--name` value does not have the companion shape): `MOAI_FACTORY` is set, `MOAI_FACTORY_LABEL` is unset. This is the `!isCompanion` column of the §A.2 truth table.

### AC-FB-005 — Dispatch evaluates both flags (no short-circuit)

**Given** the dispatch site `internal/cli/cc.go`, **When** the file is read after M2 lands, **Then** `grep -nE 'else if .*parseCompanionLabel' internal/cli/cc.go` returns no match (the `else if` is gone) and `grep -nE 'parseFactoryFlag' internal/cli/cc.go` and `grep -nE 'parseCompanionLabel' internal/cli/cc.go` both return matches (both are called unconditionally). The same property holds for `internal/cli/glm.go`.

### AC-FB-006 — Companion does NOT set `MOAI_FACTORY`

**Given** a Go test in `internal/cli/factory_bootstrap_test.go` that launches with `-f --name run-abc123`, **When** the test inspects the post-launch environment, **Then** it asserts `MOAI_FACTORY == ""` and `MOAI_FACTORY_LABEL == "run-abc123"`. (Extends the prior-art test with the new combination.)

### AC-FB-007 — Lead sets `MOAI_FACTORY` (preserve SPEC-FACTORY-MODE-001)

**Given** the prior-art test that launches `moai cc -f`, **When** it runs after M2, **Then** it still passes (the lead branch is preserved), confirming REQ-FB-004.

### AC-FB-008 — Block-cap raise preserved for both branches

**Given** the two tests `TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal` and `TestAC003_BlockCapDoctrineClauseSpecific` in `internal/cli/launcher_blockcap_infinite_test.go`, **When** `go test ./internal/cli/ -run AC003` runs after M2, **Then** both pass, and `grep -nE 'EnvMoaiFactoryLabel' internal/cli/launcher_blockcap_infinite.go` continues to return a match (the OR-branch is preserved).

### AC-FB-009 — Lead sets block cap via `MOAI_FACTORY`; companion via `MOAI_FACTORY_LABEL`

**Given** a Go test that builds the env slice through `injectStopHookBlockCapForGoal` with `MOAI_FACTORY_LABEL=run-abc123` set and `MOAI_FACTORY` unset, **When** the test inspects the returned slice, **Then** it contains `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200`. Symmetrically for `MOAI_FACTORY` set / `MOAI_FACTORY_LABEL` unset.

### AC-FB-010 — Transient settings file written when no operator `--settings`

**Given** the launcher `moai cc -f`, **When** the launch path runs and the operator did not pass `--settings`, **Then** a file matching `os.TempDir()/moai-factory-*.json` exists at launch time containing `{"crossSessionInbound": "accept"}`, and the `claude` argv contains `--settings <that file>`.

### AC-FB-011 — Operator-supplied `--settings` honored (no injection)

**Given** the launcher `moai cc -f --settings /tmp/operator.json`, **When** the launch path runs, **Then** no `os.TempDir()/moai-factory-*.json` file is created and the `claude` argv carries the operator's `--settings /tmp/operator.json` unchanged. A Go test in `factory_bootstrap_test.go` asserts both.

### AC-FB-012 — Lead notice instructs operator to verify `accept` when their `--settings` is supplied

**Given** the lead SessionStart hook runs in a session where the operator supplied `--settings`, **When** `factoryLeadNotice` is rendered, **Then** the notice text contains a substring matching `verify.*crossSessionInbound.*accept` (case-insensitive) instructing the operator to confirm the field is present in their supplied file.

### AC-FB-013 — Lead notice content (happy path)

**Given** the lead SessionStart hook runs with `MOAI_FACTORY_ID=abc123` and `MOAI_FACTORY_SPEC=SPEC-FOO-001` and the transient settings file injected, **When** `factoryLeadNotice` is rendered, **Then** the output contains, in order: (a) the run id `abc123`; (b) four lines each matching `moai cc -f --name (plan|run|review|sync)-abc123`; (c) a non-empty line carrying the leader socket path (the address on the cross-session messaging substrate that companions send to; concrete producer is run-phase-derived per spec.md §A.5 — the AC asserts a non-empty Unix-path-shaped or URL-shaped string at that position, NOT a specific env var name); (d) an inbound-automation notice mentioning auto-accept; (e) the SPEC identifier `SPEC-FOO-001`.

### AC-FB-014 — Lead notice omits SPEC line when `MOAI_FACTORY_SPEC` unset

**Given** the lead SessionStart hook runs with `MOAI_FACTORY_ID=abc123` and `MOAI_FACTORY_SPEC` unset, **When** `factoryLeadNotice` is rendered, **Then** the output does NOT contain any `SPEC-` prefixed identifier and does NOT contain an empty placeholder line where the SPEC would be.

### AC-FB-015 — Companion launch lines carry `-f`

**Given** the rendered lead notice, **When** the four companion lines are inspected, **Then** each matches the regex `^moai (cc|glm) -f --name (plan|run|review|sync)-<run-id>$`. The bare `--name` form the prior-art notice prints today is gone.

### AC-FB-016 — Companion notice is role-less

**Given** the companion SessionStart hook runs with `MOAI_FACTORY_LABEL=run-abc123`, **When** `factoryCompanionNotice` is rendered, **Then** the output contains the run id `abc123` and does NOT contain the word `companion` or any of the four role names (`plan`, `run`, `review`, `sync`) other than as part of the run id. The prior-art clause `"as the %s companion"` is removed.

**AC-FB-016a (fail-open sub-clause, per C8 / EC-4).** **Given** the companion SessionStart hook runs with `MOAI_FACTORY_LABEL` unset OR with a label `SplitCompanionLabel` returns `ok=false` for (e.g. `run-` — empty run-id portion per EC-6), **When** `factoryCompanionNotice` is rendered, **Then** the output is the empty string (no notice emitted, no error raised, the launch proceeds), matching the fail-open stance of `factoryBootstrapNotice` (`session_start_factory.go:27-35`) and the C8 constraint. This covers the two fail-open paths: (i) the env var is absent (the hook was invoked in a non-companion session that nonetheless reached the companion branch), and (ii) the label does not parse (the operator passed a malformed `--name` that survived `isCompanionShape` at dispatch but fails the stricter `SplitCompanionLabel` at notice time).

### AC-FB-017 — Companion notice join-only

**Given** the companion notice rendered in AC-FB-016, **When** the text is inspected, **Then** it is a single line acknowledging the join (matching `joined run <id>`) and does NOT print the four-companion launch block (the lead-only block).

### AC-FB-018 — CLI help documents lead entry

**Given** the command `moai cc --help`, **When** the help text is rendered, **Then** the output contains the flag token `-f` or `--factory` and contains a substring stating that the flag seeds a `plan → run → verify → sync` chain. The same property holds for `moai glm --help`.

### AC-FB-019 — CLI help documents companion entry

**Given** the command `moai cc --help`, **When** the help text is rendered, **Then** the output contains a companion-entry example matching `-f --name <role>-<run-id>` and enumerates the four roles (`plan`, `run`, `review`, `sync`). The same property holds for `moai glm --help`.

### AC-FB-020 — No `cmd.Flags()` registration for `-f`

**Given** the file `internal/cli/cc.go`, **When** `grep -nE 'cmd\.Flags\(\).*factory|BoolP.*factory|StringVar.*factory' internal/cli/cc.go` runs, **Then** it returns no match. `-f` / `--factory` is documented in `Use:` / `Long:` only. The same for `glm.go`.

### AC-FB-021 — docs-site 4-locale page exists

**Given** the docs-site tree, **When** `ls docs-site/content/{en,ko,ja,zh}/multi-llm/factory-mode.md` runs after M5, **Then** all four files exist and each carries frontmatter with `title:`, `weight:`, and `draft: false`.

### AC-FB-022 — Menu entry with 4-locale `name` map

**Given** `docs-site/data/menu/main.yaml`, **When** `grep -nE 'factory-mode' docs-site/data/menu/main.yaml` runs after M5, **Then** the match is inside the `multi-llm` section's `sub:` list and the entry's `name:` carries all four keys `ko`, `en`, `ja`, `zh`, and `ref: /multi-llm/factory-mode`.

### AC-FB-023 — No new docs-site section

**Given** `docs-site/content/en/_meta.yaml`, **When** the file is read after M5, **Then** no new section entry referencing `factory-mode` or `Factory` is added at the section level; the new page is a `sub:`-level entry under the existing `multi-llm` section. `layouts/partials/menu.html` is unchanged (no new icon case).

### AC-FB-024 — Template neutrality (no SPEC/REQ/SHA leak)

**Given** the template source tree, **When** `grep -rnE 'SPEC-FACTORY-BOOTSTRAP-001|REQ-FB-|94025ce0a' internal/template/templates/` runs after every milestone, **Then** it returns no match (CLAUDE.local.md §25; spec.md C2).

### AC-FB-025 — No edits to `.moai/specs/SPEC-KANBAN-*`

**Given** the git diff of the branch against base, **When** `git diff --name-only chore/revert-kanban-rename..HEAD -- .moai/specs/SPEC-KANBAN-` runs, **Then** it returns no match (C3; the sibling is off-limits).

### AC-FB-026 — Worktree isolation

**Given** the artifact paths, **When** `ls .moai/specs/SPEC-FACTORY-BOOTSTRAP-001/` runs from the worktree root, **Then** all six artifacts (`spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md`) are present; and `ls /Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/ 2>/dev/null` returns no such directory in the primary checkout (C6 — work only in this worktree).

### AC-FB-027 — Companion-shape `--name` alone is a no-op (breaking change from `94025ce0a`)

**Given** the launcher `moai cc`, **When** the operator runs `moai cc --name run-abc123` (companion-shape `--name`, no `-f`), **Then** the launcher does NOT set `MOAI_FACTORY`, does NOT set `MOAI_FACTORY_LABEL`, does NOT set `MOAI_FACTORY_ID` or `MOAI_FACTORY_SPEC`, does NOT write a factory session record (no `recordFactorySession` call), and `--name run-abc123` is passed through to claude untouched (the session launches as if `--name` were an ordinary claude flag). This is the 4th row of the §A.2 truth table — the path that was companion entry under `94025ce0a` (commit message: "Companions launch under `--name <role>-<run-id>`") and is reclassified as a no-op by REQ-FB-001's no-`-f` clause (see §A.2.1). This AC is distinct from AC-FB-003 (non-companion-shape `--name` alone → no-op) because this case carries the companion-shape label, which the prior-art commit treated as companion entry; a Go test in `internal/cli/factory_bootstrap_test.go` asserts all four env-var unset checks plus the passthrough.

---

## §C. Edge Cases

- **EC-1 — Two leads launched in the same Unix second.** `NewRunID` is base36-of-Unix-second; two `-f` launches within the same second produce the same run id. This residual is **prior art** (the `94025ce0a` commit explicitly leaves it standing) and is NOT in scope for this SPEC. Documented in `bootstrap.go`'s `NewRunID` doc comment.
- **EC-2 — Operator passes `--settings` to a companion, not just the lead.** REQ-FB-007 applies symmetrically; the companion launcher also suppresses injection and prints the same advisory. AC-FB-011 / AC-FB-012 cover the lead case; the companion case is structurally identical and tested by the same pattern.
- **EC-3 — Operator passes `-f --name run-abc123 --settings /tmp/op.json`.** The companion branch is selected (truth table row 2) AND the operator-supplied-settings branch is selected. The companion notice (role-less per AC-FB-016) is emitted; the injected settings are suppressed; the advisory is printed.
- **EC-4 — Transient settings file write fails (disk full, permission denied).** Fail-open (C8): the launch proceeds without the injected `--settings`; the bootstrap notice prints the operator-advisory (same form as AC-FB-012) because moai cannot guarantee `accept`.
- **EC-5 — `MOAI_FACTORY_SPEC` is set to an empty string.** Treated as unset (AC-FB-014); the SPEC line is omitted, not printed as an empty placeholder.
- **EC-6 — `--name` value has the companion role but not the run-id shape (e.g., `--name run-`).** `SplitCompanionLabel` returns `ok=false` (the run-id portion fails `isRunIDShape` because it is empty); the session falls through to the `!isCompanion` column. For `-f --name run-` this is the lead branch (AC-FB-004); for `--name run-` alone this is the unchanged no-op (AC-FB-003).

---

## §D. AC Matrix (severity + traceability)

| AC | REQ | Severity | Surface | Verifier |
|---|---|---|---|---|
| AC-FB-001 | REQ-FB-001, REQ-FB-004 | MUST | env | Go test |
| AC-FB-002 | REQ-FB-001, REQ-FB-003 | MUST | env | Go test |
| AC-FB-003 | REQ-FB-001 | MUST | env | Go test (prior-art extended) |
| AC-FB-004 | REQ-FB-001, REQ-FB-002 | MUST | env | Go test |
| AC-FB-005 | REQ-FB-002 | MUST | source | grep |
| AC-FB-006 | REQ-FB-003 | MUST | env | Go test |
| AC-FB-007 | REQ-FB-004 | MUST | env | Go test (prior-art unchanged) |
| AC-FB-008 | REQ-FB-005, REQ-FB-018 | MUST | test | `go test -run AC003` |
| AC-FB-009 | REQ-FB-005 | MUST | env | Go test |
| AC-FB-010 | REQ-FB-006 | MUST | filesystem + argv | Go test |
| AC-FB-011 | REQ-FB-007 | MUST | filesystem + argv | Go test |
| AC-FB-012 | REQ-FB-007 | MUST | notice text | Go test |
| AC-FB-013 | REQ-FB-008, REQ-FB-009, REQ-FB-011 | MUST | notice text | Go test |
| AC-FB-014 | REQ-FB-009 | MUST | notice text | Go test |
| AC-FB-015 | REQ-FB-011 | MUST | notice text | Go test |
| AC-FB-016 (incl. AC-FB-016a fail-open sub-clause) | REQ-FB-010, (C8 fail-open) | MUST | notice text | Go test |
| AC-FB-017 | REQ-FB-010 | MUST | notice text | Go test |
| AC-FB-018 | REQ-FB-012 | MUST | help text | `moai cc --help` |
| AC-FB-019 | REQ-FB-013 | MUST | help text | `moai cc --help` |
| AC-FB-020 | REQ-FB-014 | SHOULD | source | grep |
| AC-FB-021 | REQ-FB-015 | MUST | docs-site | `ls` + frontmatter read |
| AC-FB-022 | REQ-FB-016 | MUST | docs-site | grep `main.yaml` |
| AC-FB-023 | REQ-FB-017 | MUST | docs-site | `_meta.yaml` + `menu.html` diff |
| AC-FB-024 | (C2 neutrality) | MUST | template source | grep |
| AC-FB-025 | (C3 sibling off-limits) | MUST | git diff | `git diff --name-only` |
| AC-FB-026 | (C6 worktree isolation) | MUST | filesystem | `ls` both trees |
| AC-FB-027 | REQ-FB-001 | MUST | env + filesystem + argv | Go test |

### §D.1 Severity policy

- **MUST** (25 criteria): failure blocks merge. Every MUST criterion corresponds to either a hard contract (REQ-FB-001..018) or a HARD constraint (C2/C3/C6/C8). AC-FB-016a is a fail-open sub-clause paired with AC-FB-016 (per the AC sub-ID convention) — it MUST pass but does not inflate the criterion count separately from AC-FB-016.
- **SHOULD** (1 criterion: AC-FB-020): failure emits an advisory. The grep against `cmd.Flags()` registration is a defense against a documentation anti-pattern that does not break behavior; it is SHOULD because the inert registration would not affect functionality, only reader comprehension.

### §D.2 Indirect verification

REQ-FB-018 (AC003 preserve) is verified indirectly by AC-FB-008: the two named tests passing IS the preservation. No additional indirect verification is required because the block-cap behavior is fully covered by the prior-art test pair.

### §D.3 Closure gates (Definition of Done)

1. All 24 MUST criteria PASS.
2. AC-FB-020 (SHOULD) PASS or advisory-accepted with recorded rationale.
3. `go test ./...` exits 0 at HEAD of `feat/factory-bootstrap-guidance` after every milestone.
4. `make build` exits 0 (M6 verification).
5. No SPEC/REQ/SHA leak into `internal/template/templates/` (AC-FB-024).
6. No edits to `.moai/specs/SPEC-KANBAN-*` (AC-FB-025).

### §D.4 Forward-looking checks (post-merge)

- **When `SPEC-KANBAN-BOOTSTRAP-001` lands**, its topology-config-gated notice supersedes the unconditional notice shipped here (spec.md §C). At that point, REQ-FB-008 / REQ-FB-010 / REQ-FB-011 are CONSUMED by the sibling and the corresponding AC (AC-FB-013 / AC-FB-016 / AC-FB-015) become the sibling's responsibility. The emit mechanism (`factoryBootstrapNotice` / `factoryLeadNotice` / `factoryCompanionNotice`) stays; only the unconditional → conditional upgrade moves.
- **When Claude Code's `--settings` merge semantics change** (the stricter-tier-wins rule that motivates REQ-FB-006), REQ-FB-006 / REQ-FB-007 and AC-FB-010 / AC-FB-011 / AC-FB-012 must be re-measured. The injection path is only load-bearing while that semantics holds.

### §D.5 Quality-gate criteria (TRUST 5)

- **Tested**: AC-FB-001..020 are Go tests or greps; coverage of `internal/cli/factory.go`, `internal/hook/session_start_factory.go`, `internal/factory/bootstrap.go` stays at or above the package baseline measured at HEAD `94025ce0a`.
- **Readable**: the dispatch rewrite in `cc.go` / `glm.go` carries the four-way truth-table comment; the notice functions carry doc comments naming the removal rationale (spec.md §A.6).
- **Unified**: `gofmt` + `golangci-lint` clean on touched packages.
- **Secured**: the transient settings file is written under `os.TempDir()` with session-private naming (PID + random); it is cleaned up on session exit. No secrets are written to it.
- **Trackable**: every commit on `feat/factory-bootstrap-guidance` follows Conventional Commits (`feat(factory): ...`, `docs(site): ...`, `test(factory): ...`).

### §D.6 Indirect-traceability gaps

None. Every REQ-FB-001..018 maps to at least one AC-FB-XXX in §D; no REQ is without a criterion.

### §D.7 Audit-ready summary (for plan-auditor)

- **27 criteria** total (25 MUST + 1 SHOULD + AC-FB-026 worktree-isolation which is MUST but meta). AC-FB-016a is a sub-clause of AC-FB-016, not a separate criterion.
- **18 requirements** (REQ-FB-001..018), all traced to ≥ 1 AC.
- **Tier L**: 18 of 25 REQ ceiling used; 27 AC against 18 REQ (9-criterion surplus reflects the multi-surface span — env / source / notice / help / docs-site / template-neutrality / sibling-boundary / worktree-isolation / breaking-change-pin).
