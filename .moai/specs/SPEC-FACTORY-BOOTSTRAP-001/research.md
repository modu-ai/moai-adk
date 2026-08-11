# research.md — SPEC-FACTORY-BOOTSTRAP-001

> Measurement record. Every fact in `spec.md` §A and `plan.md` §A.4 is re-traced here with the command run, the observed output, and the baseline attribution (HEAD `94025ce0a` in worktree `~/.moai/worktrees/kanban`, branch `feat/factory-bootstrap-guidance`). The verification-claim-integrity.md §2 baseline-attribution rule binds every row: a fact is valid only while the command + observed output remain attributable to this baseline.

---

## §A. Baseline attribution

- **Worktree root**: `/Users/goos/.moai/worktrees/kanban` (`git rev-parse --show-toplevel`).
- **Branch**: `feat/factory-bootstrap-guidance` (`git branch --show-current`).
- **HEAD**: `94025ce0a` (`git rev-parse --short HEAD`).
- **Tree status**: clean (`git status --short` → no rows at measurement time).
- **Base**: `chore/revert-kanban-rename` (`24c4674b5`) ← `origin/main`; local ahead by 2, no race.

---

## §B. Prior-art commit `94025ce0a` — file inventory

**Command**: `git show --stat 94025ce0a`

**Observed** (abridged): 12 files / +724 lines, including:

- `internal/cli/factory.go` (entry-switch functions)
- `internal/factory/bootstrap.go` (vocabulary)
- `internal/hook/session_start_factory.go` (SessionStart notice)
- `internal/config/envkeys.go` (env-key constants)
- `internal/cli/cc.go`, `internal/cli/glm.go` (dispatch wiring)
- `internal/cli/launcher_blockcap_infinite.go` (block-cap inject)
- `internal/cli/factory_bootstrap_test.go`, `internal/factory/bootstrap_test.go`, `internal/hook/session_start_factory_test.go`, `internal/cli/launcher_blockcap_infinite_test.go`

**Commit message abridged**: "feat(factory): announce companion session bootstrap from the SessionStart hook". The message explicitly states the lead is the only session carrying `-f` (chain seed) and that companions launch under `--name <role>-<run-id>` taking the raised block cap through `MOAI_FACTORY_LABEL`. The message also records: "Pre-commit gate overridden with SKIP_MOAI_PRECOMMIT=1. Attribution check: the gate reports 22,506 repo-wide ast-grep findings."

---

## §C. `crossSessionInbound` zero-presence

**Command C.1**: `grep -rn crossSessionInbound internal/template/templates/ | wc -l`

**Observed**: `0`.

**Command C.2**: `grep -rn crossSessionInbound internal/ pkg/ cmd/ | wc -l`

**Observed**: `0`.

**Interpretation**: the accept/hold/refuse ladder is not satisfied from any settings layer moai writes today, and no Go code reads the field. This is the load-bearing finding for REQ-FB-006 (the transient `--settings` injection is the only path that can force `accept` given the stricter-tier-wins merge rule).

---

## §D. `--settings` zero-presence in moai code

**Command**: `grep -rn '\-\-settings' internal/ pkg/ cmd/`

**Observed**: 2 matches, both inside `internal/template/templates/.claude/skills/moai-foundation-cc/reference/`:

- `claude-code-cli-reference-official.md:200` — `claude -p "Task" --settings '{"model": "opus"}'`
- `claude-code-headless-official.md:208` — same form

**Interpretation**: the matches are in Claude Code reference documentation that moai ships as a skill resource; they are **not** moai Go code parsing or forwarding `--settings`. Moai's own launcher does not handle `--settings` at all today, so the injection surface is genuinely new (REQ-FB-006 introduces it).

**Cross-check**: `grep -rn 'settings' internal/cli/tool_policy.go` returns matches for `--local-only` / `--template-only` flags only (`internal/cli/tool_policy.go:172-173`), confirming the `--settings` surface is distinct from any existing settings-handling moai code.

---

## §E. Dispatch site — `cc.go` and `glm.go`

**Command E.1**: `sed -n '88,112p' internal/cli/cc.go`

**Observed** (abridged):

```go
// SPEC-FACTORY-MODE-001: --factory / -f seeds a plan -> run -> verify -> sync
// chain in the launched session. ...
if specID, factoryEnabled, factoryArgs := parseFactoryFlag(filteredArgs); factoryEnabled {
    filteredArgs = factoryArgs
    defer enterFactoryMode(specID)()
    recordFactorySession(specID, factory.BackendClaude)
} else if label, isCompanion := parseCompanionLabel(filteredArgs); isCompanion {
    // A companion of a factory run: same raised Stop-hook block cap as the
    // lead, no chain seed. ...
    defer enterFactoryCompanionMode(label)()
}
```

**Command E.2**: `sed -n '160,182p' internal/cli/glm.go`

**Observed**: parallel structure with the same `else if`.

**Interpretation**: the `else if` short-circuits companion classification whenever `-f` is present. A launch of `moai cc -f --name run-abc` enters the lead branch and sets `MOAI_FACTORY`, never reaching `enterFactoryCompanionMode`. This is the structural finding REQ-FB-002 addresses (the dispatch revision to evaluate both flags together).

---

## §F. Block-cap inject — both env vars OR'd

**Command**: `sed -n '40,60p' internal/cli/launcher_blockcap_infinite.go`

**Observed** (abridged):

```go
// SPEC-FACTORY-MODE-001 REQ-FM-023: the factory branch is UNCONDITIONAL and
// sits ahead of the goal read. ...
if os.Getenv(config.EnvMoaiFactory) != "" || os.Getenv(config.EnvMoaiFactoryLabel) != "" {
    return setStopHookBlockCap(base, DefaultRaisedStopHookBlockCap)
}
```

**Interpretation**: the cap raise fires for both lead (`EnvMoaiFactory`) and companion (`EnvMoaiFactoryLabel`). The `-f` redefinition changes which var a companion sets, but the OR keeps the cap raise for the companion path. This is the property REQ-FB-005 / REQ-FB-018 preserve; touching this wiring is explicitly out of scope.

---

## §G. Companion notice — role clause present

**Command**: `sed -n '60,72p' internal/hook/session_start_factory.go`

**Observed**:

```go
func factoryCompanionNotice(label string) string {
    role, runID, ok := factory.SplitCompanionLabel(label)
    if !ok {
        return ""
    }
    return fmt.Sprintf("Factory Mode: joined run %s as the %s companion.", runID, role)
}
```

**Interpretation**: the prior-art notice prints `"as the %s companion"` with the role. This is the clause REQ-FB-010 removes; rationale at spec.md §A.6 (collision with the sibling's `REQ-KS-006` role-declaration carrier).

---

## §H. AC003 preserve-tests

**Command**: `grep -rln 'TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal\|TestAC003_BlockCapDoctrineClauseSpecific' internal/`

**Observed**: `internal/cli/launcher_blockcap_infinite_test.go` (single file, both tests).

**Interpretation**: the two named tests live in one package (`internal/cli`). REQ-FB-018 binds them as the regression guard for the block-cap wiring being left untouched.

---

## §I. docs-site structure — section inventory and page count

**Command I.1**: `ls docs-site/content/en/`

**Observed** (14 directories + 2 files): `_index.md`, `_meta.yaml`, `advanced/`, `changelog/`, `claude-code/`, `cli-reference/`, `contributing/`, `core-concepts/`, `cost-optimization/`, `getting-started/`, `guides/`, `multi-llm/`, `resources/`, `utility-commands/`, `workflow-commands/`, `worktree/`.

**Command I.2**: `ls docs-site/content/en/multi-llm/`

**Observed**: existing `multi-llm/` pages (cg-mode, glm, teammate runtime, etc.).

**Command I.3**: `grep -rn 'factory-mode\|Factory Mode' docs-site/content/en/`

**Observed**: 0 page-level matches (the 4 `factory` grep hits cited in the task are all `ExecutionFactory` code-API examples in `claude-code/agentic/best-practices.md`, not Factory Mode pages).

**Command I.4**: `head -60 docs-site/data/menu/main.yaml`

**Observed**: top-level `main:` list with section entries carrying 4-locale `name` maps (`ko`/`en`/`ja`/`zh`) + `ref` + `icon` + optional `sub:` list. The `multi-llm` section has a `sub:` list of pages.

**Interpretation**: a new Factory Mode page lands as a `sub:` entry under the existing `multi-llm` section. The page frontmatter is `title` / `weight` / `draft:false` across four locale files. The `main.yaml` entry carries the 4-locale `name` map and `ref: /multi-llm/factory-mode`. No new section, no new icon case (the `multi-llm` icon already exists in the `menu.html:28-44` SVG switch). This grounds REQ-FB-015 / REQ-FB-016 / REQ-FB-017 and design.md §5.

---

## §J. Env keys — `envkeys.go` factory cluster

**Command**: `grep -nE 'EnvMoaiFactory' internal/config/envkeys.go`

**Observed**: four constants, `EnvMoaiFactory` (`MOAI_FACTORY`), `EnvMoaiFactorySpec` (`MOAI_FACTORY_SPEC`), `EnvMoaiFactoryID` (`MOAI_FACTORY_ID`), `EnvMoaiFactoryLabel` (`MOAI_FACTORY_LABEL`).

**Interpretation**: the four env-var names are already canonicalized; REQ-FB-003 / REQ-FB-004 reference them by constant, not by string literal (CLAUDE.local.md §14 hardcoding-prevention).

---

## §K. CLI `Use:` field — today's text

**Command**: `grep -nE 'Use:' internal/cli/cc.go internal/cli/glm.go`

**Observed**: `cc.go` `Use: "cc [-p profile] [-- claude-args...]"`; `glm.go` similar. Neither mentions `-f` / `--factory`.

**Interpretation**: the help surface today documents neither the lead entry nor the companion entry. REQ-FB-012 / REQ-FB-013 / REQ-FB-014 ground the revision (documentation in `Use` / `Long`, not `cmd.Flags()`).

---

## §L. Sibling boundary — `SPEC-KANBAN-BOOTSTRAP-001` (read-only)

**Command**: `sed -n '1,20p' .moai/specs/SPEC-KANBAN-BOOTSTRAP-001/spec.md`

**Observed**: `id: SPEC-KANBAN-BOOTSTRAP-001`, `status: draft`, `tier: L`, 25 REQ / 30 AC at the Tier L ceiling. The sibling's HISTORY shows it has been through four plan-audit repair rounds (v0.1.0 → v0.5.0) and is deferred to the next release.

**Command**: `grep -nE 'REQ-KS-006|REQ-KS-007|REQ-KS-013|REQ-KS-014' .moai/specs/SPEC-KANBAN-BOOTSTRAP-001/spec.md | head`

**Observed**: the named REQs exist and govern role-declaration, topology-config-gated quorum wait, multi-backend command emission, and distinct-label generation respectively.

**Interpretation**: the boundary this SPEC states (spec.md §C) is one-sided. The sibling owns the role-declaration carrier (`REQ-KS-006`), quorum (`REQ-KS-007` / `REQ-KS-012`), dispatch (`REQ-KS-018` / `REQ-KS-019`), and topology-config-gated multi-backend emission (`REQ-KS-013`). This SPEC takes only the bootstrap-notice **emit mechanism** (the SessionStart hook surface) and the `-f` companion-entry semantics, and is consumed (not re-authored) by the sibling when it lands. The sibling's files are not edited from this side (C3).

---

## §M. Residual risk

- **M-1 — Same-second run id collision.** `NewRunID` is base36-of-Unix-second; two `-f` launches within the same second collide. Left standing by `94025ce0a`; out of scope here. Risk: low (manual two-terminal operation, one-second window).
- **M-2 — `--settings` merge semantics change.** REQ-FB-006's injection path rests on Claude Code's documented stricter-tier-wins merge rule for `--settings`. If that rule changes (the file becomes weakest-tier-wins, or project settings take precedence over file), the injection becomes ineffective and REQ-FB-006 / REQ-FB-007 must be re-measured. Forward-looking check: acceptance.md §D.4.
- **M-3 — Transient settings file cleanup race.** The `defer os.Remove(...)` runs at launcher exit, but a crashed launcher leaves the file behind. Mitigation: session-private naming (`<pid>-<rand>`) prevents cross-session collision; OS tmpdir reaper eventually cleans.
- **M-4 — Sibling's notice supersedes this SPEC's notice on different terms.** When `SPEC-KANBAN-BOOTSTRAP-001` lands, the unconditional → conditional upgrade is the sibling's deliverable. If the sibling's terms differ from this SPEC's forward reference (spec.md §C), the boundary will need renegotiation at that point. Risk: low (this SPEC's emit mechanism is consumed, not re-authored; the sibling inherits the function trio as-is).
- **M-5 — Operator-supplied `--settings` collision with future moai-injected flags.** If moai later injects additional settings fields beyond `crossSessionInbound`, the operator-supplied `--settings` suppression (REQ-FB-007) becomes coarser than needed — the operator might want moai's `crossSessionInbound` injected but their own `model` honored. Out of scope for this SPEC; flagged for the follow-up that adds the second injected field.
