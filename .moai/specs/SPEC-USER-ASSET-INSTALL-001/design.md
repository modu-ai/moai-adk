---
id: SPEC-USER-ASSET-INSTALL-001
title: "design.md — user-folder asset install architecture"
version: "0.6.4"
created: 2026-10-05
updated: 2026-10-05
author: manager-spec
---

# Design — User-Folder Common-Asset Install

## 1. Target Model

One install axis changes: WHERE common skills and agents live.

```
BEFORE (today, two modes)                    AFTER (this SPEC)
├─ project .claude/skills + commands         ├─ USER ~/.claude/skills, ~/.claude/agents   (Claude)
│   (local mode) OR plugin carrier           ├─ USER $HOME/.agents/skills                 (Codex skills)
│   (plugin mode)                            ├─ USER ~/.codex/agents                      (Codex agents)
├─ project .claude/agents, .codex/agents     ├─ project: settings, AGENTS.md, lock file,
├─ project .agents/skills mirror                 project-only harness (hooks, .mcp.json,
└─ plugin marketplace + install step             output-styles, rules, commands wrappers)
                                             └─ NO plugin anything
```

Per-profile settings folders (`~/.moai/claude-profiles/<name>`, the
`CLAUDE_CONFIG_DIR` roots) remain per-profile for SETTINGS and are never an
install target for the shared asset tree (decision-index D-Q3: closed at plan
phase — profile sessions do not see the shared user assets in v1; premise P6).

## 2. Components

### 2.1 User-asset installer (new, `internal/template` sibling or subpackage)

- Reads the same embedded tree + catalog (`//go:embed all:templates`,
  `catalog.yaml`) the project deployer reads — zero new asset sources.
- Writes to four user roots, resolved at run time:
  - Claude: `$HOME/.claude/skills/`, `$HOME/.claude/agents/`
  - Codex: `$HOME/.agents/skills/`, `$HOME/.codex/agents/`
- Per-file rule set — the four-state hash truth table (iter2 D16; state ×
  action for a manifest-tracked file):
  - **manifest-match** (current == manifest hash, ≠ shipped): refresh →
    rewrite to shipped bytes + re-record hash/version (REQ-008); removal →
    remove (REQ-009).
  - **up-to-date** (current == manifest hash == shipped): refresh → no-op,
    not counted; removal → remove.
  - **manifest-stale** (current ≠ manifest hash, == shipped bytes — e.g. a
    prior refresh crashed before the manifest write): refresh → repair the
    manifest entry to the shipped state WITHOUT a file rewrite, counted under
    "refreshed" (REQ-011); removal → remove — the bytes are moai's own, which
    is exactly REQ-009's shipped-bytes alternative (iter4 D25: the one
    removal rule is REQ-009 as extended — current == manifest hash OR current
    == shipped bytes where a shipped source exists; this table arm and the
    REQ now state the same rule).
  - **divergence** (current equals NEITHER manifest hash NOR shipped bytes):
    preserve (REQ-023) — at refresh, back up the shipped replacement to the
    backup home `~/.moai/backups/<root-slug>/<relpath>` (iter4 D33: the
    layout is ROOT-SLUG-PREFIXED — each root gets a fixed slug, e.g.
    `claude-skills`, `claude-agents`, `agents-skills`, `codex-agents`, so
    `~/.claude/skills/x` and `$HOME/.agents/skills/x` cannot collide on
    `skills/x`; the `<relpath>` is Cleaned and `..`-rejected before joining,
    and the backup home is judged on RESOLVED paths exactly like the four
    roots — resolved once per run, destination resolved immediately before
    the write, an escape from the resolved backup root refused; C2's sole
    out-of-root write carve-out for user-folder asset writes — NOT
    "alongside the user folder", which would sit outside the four roots or
    pollute them as an untracked file); at removal of a file dropped from
    every bundle, NO shipped-bytes backup exists, so the file is preserved
    in place + reported; the tracked path is unmodified by refresh AND by
    removal; always reported.
  - **missing** (tracked in manifest, absent on disk): refresh → reinstall +
    record; removal → drop the manifest entry, count as removed.
  - target exists AND manifest does NOT track it → SKIP + report (collision;
    the file may be the user's — mirrors `rehomeOneSkill`'s skip-and-report and
    `UserCreated` provenance semantics)
  - target absent AND untracked → install
- Init trigger semantics (round-5 F3; fold A1 correcting the first cut):
  init's install is PER-ASSET-STATE based, not manifest-absence based — and
  the per-asset judgment applies the §2.1 truth table EXACTLY as update
  does, never a blanket reinstall: an ABSENT tracked target is installed;
  a present tracked file whose bytes differ from its manifest record is
  classified by the truth table — manifest-hash match (stale install) →
  refresh per REQ-008; current == shipped (manifest-stale) → manifest
  repair, no rewrite; NEITHER (user edit) → REQ-023 divergence preserve +
  backup + report, bytes untouched; a present UNTRACKED target → REQ-010
  collision skip + report. "Bytes differ" alone is ambiguous between
  stale-install and user-edit — the fold-A1 gate measured user bytes
  clobbered on exactly that wording. A manifest left by a PARTIAL install
  therefore does not suppress the run: init completes the shortfall
  idempotently, sharing the shortfall-append semantics with update's
  upgrade branch (§2.4 state (a)). The former "no per-user install"
  wording (v0.4.0 and earlier) left the partial-install retry skipping the
  missing assets with no project fallback after M4.
- Per-file failure (permissions, EISDIR, …) → continue + surface in summary
  (fail-open per file, loud at the end; never a silent partial install).
- Four-root confinement is judged on RESOLVED paths, covering all three
  second-order edges (iter2 D17):
  1. **Symlinked root** — each root is resolved once at install start
     (`filepath.EvalSymlinks`); a root that is itself a symlink (dotfile-manager
     `~/.claude` → elsewhere) is a legal boundary at its RESOLVED location —
     containment is judged resolved-root vs resolved-destination, so such a
     setup neither collapses into wholesale refusal nor escapes the boundary.
  2. **Leaf symlink inside a root** — a managed leaf that is itself a symlink
     is resolved; if it resolves outside its own root's resolved tree it is
     NEVER written through: refused at install, classified as divergence
     (REQ-023 preserve + report) at refresh/removal.
  3. **TOCTOU (check-then-open)** — the resolve-then-write window is NARROWED
     by posture, not closed by it (iter4 D26 downgrading the earlier
     "closed" claim, which was false as stated): the payload is written to a
     temp file created inside the validated RESOLVED directory and moved
     onto the final path with an atomic rename — `rename(2)` does not follow
     a symlink on its destination's final component, so a leaf swapped in
     after validation is replaced, not followed — with the resolved parent
     re-validated immediately before the rename (the
     O_NOFOLLOW-equivalent semantics for the update path). DECLARED
     LIMITATION: `rename(2)` still traverses INTERMEDIATE directory
     components, so a parent directory swapped to an outside-pointing
     symlink after that re-validation but before the rename can redirect the
     rename outside the root (codex reproduced the sequence on a /tmp model,
     iter3 D26). The remaining window is the instant between the
     re-validation and the rename; full closure needs descriptor-relative
     operations (held dirfd + the openat/renameat family, or
     O_NOFOLLOW|O_DIRECTORY on the parent), which are POSIX-only and are
     declared out of scope for v1 (the SPEC pins no OS-specific
     system-call dependency — D8; C1's cross-platform determinism). The
     limitation is
     doc-visible (acceptance §D.7) and AC-025 asserts the posture itself.
     Manifest writes keep the existing `atomicWriteFile` pattern.

### 2.2 Per-user manifest (new)

- One JSON document at `~/.moai/user-assets.json` (D-Q2, adjudicated
  2026-10-05 — leader default, operator-contestable; `~/.moai/` is already
  moai's user-level state home per `internal/paths/paths.go`). The manifest
  is the SPEC's own state file — a separately named out-of-root write
  destination alongside the REQ-023 backup home, distinct from the C2
  user-folder asset-write surface (iter4 D34).
- Schema: `schema_version`, `installed_at`, `bundles: [opted-in bundle
  names]`, `files: {path → {sha256, bundle,
  installed_at, moai_version}}`, `collisions: [{path, first_seen_at}]`. The
  installing moai version is PER FILE (REQ-006): the field records the moai
  build that successfully wrote that file; a failed write leaves the prior
  entry and its version untouched, so REQ-013's partial-failure continuation
  yields accurate mixed-version history. There is deliberately no top-level
  `moai_version` — a single top-level value cannot record that state. The
  `bundles:` list is the recorded opt-in selection (REQ-004, iter2 D18) —
  the state `moai bundle add|remove` mutates and `moai update` reads.
- Unknown-field preservation (iter2 D23; iter4 D27 extending its scope):
  EVERY manifest write carries through fields the writing binary does not
  understand (top-level and per-file), whatever schema_version it reads —
  including the append-only install/refresh writes REQ-021 permits against
  an UNKNOWN schema_version, which are exactly the older-binary-rewrites-
  newer-manifest case this rule exists to make safe — the
  consumers-ignore-unknown-fields premise of acceptance §D.7. An
  implementation whose decode would drop unknown fields refuses the write
  instead of silently shrinking the document.
- Reuses the sha256 hex convention of `catalog.yaml` entries and the triple-hash
  spirit of the project manifest (`internal/manifest/types.go`) minus the parts
  the user scope does not need (no 3-way merge at user scope: collision = skip).
- Read-modify-write serialization (round-5 F4): manifest mutation is
  serialized per user — a user-level lock (or equivalent single-writer
  discipline) spans manifest READ → asset changes → manifest SAVE. Without
  it, concurrent init/update/`moai bundle` operations from different
  projects of the same user interleave read-modify-write cycles and the
  last writer silently erases the earlier run's selections (bundle list
  entries, file records). The lock is a user-level concern (one lock file
  beside the manifest under `~/.moai/`), not a project-level one — the
  concurrent writers are different projects sharing one manifest.
- Pending-install recovery journal (final-class item 5): the
  READ → asset-changes → SAVE ordering has a crash window — an install
  interrupted AFTER asset writes but BEFORE the manifest save leaves
  moai-written files in the user folders that the manifest does not track,
  and the retry classifies them as REQ-010 collision-skips FOREVER (the
  gate reproduced it): moai's own installs become permanent untracked
  squatters, and their real owners never install. Design: before the
  asset-write phase, the run writes a PENDING-INSTALL JOURNAL (a
  manifest-path sibling under `~/.moai/`) recording the intended delta —
  each target path with its intended sha256; on run start, an existing
  journal is RECONCILED before any install/collision judgment: a
  user-folder file whose bytes hash to the journal's recorded value is
  COMPLETED (its manifest entry is committed — the retry identified its
  OWN install), and a file NOT matching the journal (or not in it) falls
  through to the normal collision/divergence path — the journal's
  expected-hash match is what lets retry claim its own installs WITHOUT
  absorbing user files. The journal is cleared atomically with the
  manifest save.
  Journal COMPLETENESS (directed repair R-e, v0.6.3 — the gate's temp
  model reproduced the empty-selection recovery: `recovered bundle
  selection: []` followed by REQ-009 re-pruning the files the interrupted
  run had just installed — that exact sequence is the red this
  completeness closes). Per entry AND for the run, the journal records:
  (1) the bundle-SELECTION delta — the `bundles:` list change the
  interrupted run intended — so an `init --bundles` interrupted before
  the manifest save recovers WITH the recorded selection intact, never
  empty, and the recovered files are never REQ-009 re-pruned on the next
  update; (2) the FULL manifest-entry provenance per file — bundle,
  moai_version, installed_at, not merely path+sha256 — so a
  different-binary replay restores the SAME versions (path+sha256 alone
  makes two same-byte installs by different binaries indistinguishable,
  which would defeat REQ-006's per-file installing version); (3) the
  ownership evidence — a write-completion flag per entry, reconciled
  atomically with the manifest save. The flag's EXACT meaning and the
  rename→flag window (in-round extension E4, v0.6.3 — the gate
  reproduced permanent REQ-010 collisions from an interruption exactly
  between the rename and the flag save): the STAGING record — written
  BEFORE the rename, carrying path + intended sha256 + full provenance —
  is itself the intent-and-content proof; `rename(2)` is atomic, so a
  file standing at the FINAL path whose bytes hash to a staged entry's
  recorded value is a COMPLETED write of the intended content whether or
  not the flag write landed. The recovery lattice is EXACTLY THREE CASES
  (in-round extension E5 sharpening E4 — a hash mismatch is NOT evidence
  the write never completed: after an interruption-before-manifest-save
  the USER may have edited the file, and reinstalling on mismatch
  overwrites the user's edit — the gate reproduced user_bytes_preserved=
  False): (1) target ABSENT → install from the journal entry (the
  staging record's provenance); (2) target PRESENT and hash == the
  journal's sha256 → claim as OWN (the staging record is the
  intent-and-content proof, flag or no flag — E4); (3) target PRESENT
  and hash != the journal's sha256 → NEVER reinstall: preserve the
  file's bytes and classify per the ownership evidence — a
  flag-complete entry (the run's write was observed, then the file
  changed underneath) is REQ-023 divergence (backup + report); an
  unflagged entry (cannot distinguish user-edit-after-write from
  user-file-never-written) is REQ-010 collision — both preserve the
  user's bytes, and no recovery path ever overwrites on a mismatch.
  Cross-binary payload recovery for case (1) (directed repair R-f-① —
  DESIGN DECISION, option (b) REINSTALL-FROM-CURRENT-VERSION chosen,
  documented per the dispatch): a retry from a DIFFERENT binary cannot
  restore the interrupted run's original bytes from path+hash+provenance
  alone (`replay with current embedded bytes matches staged sha256:
  False` — the gate's observation). The chosen rule is HONEST RE-STAMP:
  the retry writes the CURRENT binary's bytes for the absent target and
  refreshes the journal/manifest provenance to record the binary that
  actually wrote the recovered bytes. REQ-006 interaction: the per-file
  installing version stays TRUTHFUL — it names the build that produced
  the bytes on disk, which after a cross-binary recovery is the retrying
  binary; REQ-013's mixed-version manifest already makes mixed
  provenance states real, so the re-stamp adds no new state class.
  Option (a) — persisting the payload bytes in the journal (a staging
  file retained until reconciliation) — was REJECTED on the simplicity
  ladder: it doubles the install-set storage under `~/.moai/` and adds a
  shadow-tree retention/cleanup lifecycle, buying only byte-fidelity to
  a SUPERSEDED binary's payload; no user bytes are ever at risk in case
  (1) (the target is absent — nothing to preserve), and the mismatch
  cases are governed by E5's never-reinstall lattice. C1 determinism
  binds the same binary against the same tree and is unaffected.
- Schema-version gate: unknown `schema_version` → refuse manifest-driven
  removal (REQ-021); install/refresh may still proceed in append-only fashion.

### 2.3 Bundle taxonomy (catalog extension)

- `catalog.yaml` gains a user-install view: L0 core (REQ-003's content — the
  five agents per the resolved D-Q1: manager-spec, manager-develop,
  manager-docs, plan-auditor, sync-auditor; the moai-plan/moai-run/moai-sync
  published command skills per D-Q4; the hook payload deploys project-side;
  factory) + opt-in bundles (REQ-004). The existing `optional_packs`
  structure is the natural carrier for bundles: the six existing optional
  packs stand as-is, and the current-`core` entries that are NOT L0
  re-bundle by theme (resolved D-Q5, 2026-10-05 — leader default,
  operator-contestable). The published Codex command-skill set folds into
  the same catalog view (iter4 D29): the plan/run/sync three ride L0 (D-Q4)
  and the remaining fourteen take their bundles from the reclassification —
  `publishedSkillNames` does not survive as a second name list (the catalog
  is the single membership SSOT — the `publishedSkillNames` hardening
  pattern shows why drift guards exist).
- Factory (multi-lane operation) is L0 per the operator decision D3:
  `moai-factory-foreman` + `moai-lane-watchdog` ride the core bundle; the
  factory doctor check (`checkFactoryRun`) keeps reading project state as
  today.
- Selection surface (iter2 D18): bundle MEMBERSHIP is declared in the
  catalog (above); the user's opt-in SELECTION is the `bundles:` list in the
  per-user manifest (§2.2). `moai init --bundles <name,...>` sets the
  initial selection (default: L0 only); `moai bundle add <name>` applies
  the install of exactly that bundle's catalog entries; `moai bundle
  remove <name>` applies the COMPLEMENT removal (in-round extension E3,
  v0.6.3): the removal target is the entries of the removed bundle that
  are NOT in (L0 ∪ the union of the remaining opted-in selections) —
  entries the removed bundle SHARES with L0 or with another still-opted
  bundle survive with a report note, because "remove exactly that
  bundle's entries" would delete required L0 skills (measured: the
  historical `devops` pack carries `moai-ref-owasp-checklist`
  (catalog.yaml:230), `moai-ref-cross-model-audit` (:240), and
  `moai-ref-secops` (:250) — all three are L0 since E2/R-c; the gate
  reproduced their deletion). Both respect REQ-010 collision and
  REQ-023 divergence semantics. DEPENDENCY MAINTENANCE for preserved
  assets (directed repair R-f-②): before deleting, the removal step
  re-runs the derivation matrix's dependency rows — an entry that is a
  declared dependency of a PRESERVED asset (any member of L0 ∪ the
  remaining opted-in selections) has its deletion DEFERRED (kept +
  reported, re-evaluated at the next update/removal) rather than
  orphaning the preserved asset; the preserved assets' own consumer
  references are user-folder references by the class clause (§2.5), and
  the loading of the preserved assets' consumers is verified
  post-removal (AC-020 arm). `moai update` honors the recorded selection —
  installs/refreshes L0 + opted-in bundles, prunes per REQ-009 (no longer in
  L0 nor any opted-in bundle). Milestones: the `--bundles` init flag in M2;
  the `moai bundle` command and update honoring in M3.
- L0 transitive runtime skill closure (round-5 F1; fold B1 extending the
  enumeration to on-demand invoke sites): catalog L0 entries carry
  PER-ENTRY skill dependencies, and the L0 view explicitly enumerates the
  resolved transitive closure — the installer copies the enumerated set; it
  never discovers dependencies by walking skill bodies at runtime. The
  closure is the UNION of three source classes, verified against the
  template sources on tree `064ff9960` (research §2b W1-W3, W7):
  - Tier 1 — invocation/static preload: `moai` (the dispatcher — invoked
    by all three published command skills), `moai-foundation-core` (static
    preload of manager-spec, manager-develop, manager-docs),
    `moai-workflow-spec` (static preload of manager-spec),
    `moai-foundation-quality` (static preload of sync-auditor).
    plan-auditor declares NO static `skills:` preload (its body says so
    verbatim) — it contributes nothing to the static union.
  - Tier 2 — dispatcher routing rows (delegation-injected by the default
    flows): `moai-foundation-thinking` (plan row), `moai-workflow-tdd` and
    `moai-workflow-ddd` (run rows), `moai-workflow-project` (sync row).
  - Tier 3 — on-demand `Skill("...")` invoke sites in the L0 agent bodies
    (fold B1: the default TDD flow invokes skills at need — e.g.
    manager-develop.md:237 "invoke Skill(\"moai-workflow-testing\")"):
    `moai-workflow-testing` (manager-spec:248, manager-develop:237),
    `moai-workflow-worktree` (manager-spec:250, manager-develop:242),
    `moai-ref-cross-model-audit` (plan-auditor:764, sync-auditor:226 —
    final-class item 7 CORRECTING fold B1's exclusion: the DEFAULT audit
    plan exercises the cross-model path whenever a GPT/GLM session obtains
    a Claude verdict), `moai-ref-owasp-checklist` (sync-auditor:223) and
    `moai-ref-testing-pyramid` (sync-auditor:224) — in-round extension E2
    CORRECTING fold B1's exclusion a second time: the shipped sync flow
    runs Phase 8 Security Scan and Phase 10 Coverage Analysis as STANDARD
    default phases (sync.md:50 phase routing table), and
    quality-gates-quality.md:70-73 wires the 4-dimension judges into the
    shared snapshot — Functionality 40% / Security 25% (HARD threshold) /
    Craft 20% / Consistency 15% — so Security and test-coverage scoring
    are default-path, and both skills are default-flow reachable, not
    per-mission emphasis (lead adjudication + anchors; confirms the
    in-round E2 call). `moai-ref-secops` (quality-gates-quality.md:135 —
    class repair R-c): the Phase 8 delegate — the documented security
    replacement path on the STANDARD Phase 8 — loads it alongside
    `moai-ref-owasp-checklist`, so the same reachability criterion admits
    it. DESIGN CALL (stated per the dispatch): these JOIN
    L0 — documenting a degraded absent-path would be a second
    classification standard, exactly what the single reachability
    criterion exists to prevent.
  UNION TOTAL: FOURTEEN skills. Classification criterion (final-class item 7,
  upgrading fold B1's prefix heuristic): **DEFAULT-FLOW REACHABILITY** — a
  skill is IN the closure when a documented default-configuration path of
  the plan/run/sync chain invokes it, and OUT otherwise, regardless of
  name prefix. Classified OUT (mission-type injection, no default path):
  `moai-domain-html-report` (manager-docs:224 — HTML-rendering missions
  only; no default plan/run/sync path renders HTML).
- THE INSTALL-COVERAGE DERIVATION MATRIX (class repair R-c — promoted to
  plan content; absorbs the JD-8 mechanical-derivation requirement): M0
  MECHANICALLY DERIVES the install matrix by sweeping EVERY loading
  instruction across (i) the deployed workflow tree — run/sync paths
  INCLUDING `run/task-decomposition.md`, `run/phase-execution.md`,
  `sync/quality-gates-*.md`, `sync/delivery.md`, `plan/spec-assembly` —
  and (ii) the L0 agent bodies. Each invoked ROLE (manager-*/auditor) and
  each loaded SKILL (`moai-*`) crosses against the L0+bundle install set;
  every GAP becomes a matrix ROW carrying its install-verification and
  the `moai bundle add <bundle>` remediation. Known rows the matrix MUST
  capture (measured this tree): `moai-ref-api-patterns`
  (run/phase-execution.md:252 — step 4b's per-domain mapping, with its
  siblings `moai-ref-react-patterns` and `moai-domain-database`: all
  three are per-mission DOMAIN injections → BUNDLE rows, not L0);
  `moai-ref-secops` (quality-gates-quality.md:135 — default Phase 8 path
  → IN, the fourteenth member); `manager-git` (delivery Route B +
  task-decomposition:302 — the R-b precheck row). THE MATRIX IS THE M0
  DRIFT GUARD'S SOURCE OF TRUTH: the guard pins the DERIVED set, never a
  hand list — a hand-maintained closure cannot survive the next skill
  added to a workflow body.
  AGENT DEPENDENCY (in-round extension E1): the factory entry (P2 —
  `moai-factory-foreman` + `moai-lane-watchdog` in L0) carries
  `manager-lead` as its OWN declared dependency — factory-dispatch.md:104
  [HARD]: "The deputy is resident, not optional... spawns exactly one
  UNNAMED background `Agent()` running manager-lead as its coordination
  deputy" (template mirror :104 identical). manager-lead is NOT a sixth
  core agent (D-Q1's five stands for the core-agent answer); it rides the
  catalog's per-entry dependency mechanism (REQ-004) under the factory
  entry, its role body landing in `~/.claude/agents/` with the rest of
  the agent set. The drift guard carries it under the factory entry.
  A catalog drift guard pins this enumeration to its sources — the agent
  frontmatter `skills:` unions, the dispatcher's routing-table Skills
  lines, the command skills' dispatcher references, AND the on-demand
  invoke-site sweep of the L0 agent bodies (minus the classified-out
  moai-ref-*/moai-domain-* prefixes) — so the closure cannot rot silently
  (the same drift-guard pattern as the L0 list guard).
- User-side dispatcher mirror (round-5 F2; fold A2 extending the rebind to
  the dispatcher's own internals): the Codex skill root
  `$HOME/.agents/skills/` carries `moai/SKILL.md` — the dispatcher mirror
  the published command skills reference — the user-folder successor of the
  retired project-side mirror; without it a post-M4 Codex CLI harness
  cannot resolve the `read .agents/skills/moai/SKILL.md` instruction, which
  is why the reference itself is also rebound at source (§2.5). Fold A2:
  the dispatcher's INTERNAL workflow references are also project-relative —
  `Read .claude/skills/moai/workflows/<name>.md` appears EIGHTEEN times in
  templates/.claude/skills/moai/SKILL.md (:126 plan, :134 run, :142 sync,
  and the same pattern for gate/e2e/goal/gtd/fix and the remaining rows;
  the RAW pattern including the in-prose `harness-builder.md` mention at
  SKILL.md:282 is NINETEEN — final-class item 8) —
  and break identically post-M4 (the gate observed FileNotFoundError with
  user-folder copies present, because the path resolves against the project
  root). The rebind is at TEMPLATE SOURCE — the dispatcher is not a
  commandemit output; templates/.claude/skills/moai/SKILL.md IS its source
  layer (deployed verbatim) — and rebinds the WHOLE reference family (all
  nineteen raw occurrences, not only the L0 plan/run/sync three the fold
  named) to paths
  relative to the installed skill directory, so one form resolves in
  `~/.claude/skills/moai/` and `$HOME/.agents/skills/moai/` alike.

### 2.4 `moai update` integration

- New user-asset phase in the update flow: refresh (only when the file's
  current hash equals its manifest hash and differs from the shipped bytes) →
  removal (manifest-tracked files no longer in L0 nor any bundle recorded as
  opted-in in the manifest — the SELECTION-based criterion, verbatim
  REQ-009; the sole general removal rule — iter4 D28 re-keying the former
  catalog-based wording; precondition: current hash == manifest hash, or ==
  shipped bytes where a shipped source still exists, iter4 D25) →
  tracked-file divergence preserve + backup + report (REQ-023) → summary
  counts (installed/refreshed/removed/collision-skipped/divergence-preserved
  — REQ-011).
- Removal is manifest-driven ONLY: a file in a user folder that the manifest
  does not track is never a removal candidate (REQ-010 collision rule covers
  it; deletion of untracked files is out of scope).
- Upgrade branch (iter2 D14; iter4 D24 re-keying the removal gate per-asset
  — the former labels "init never ran on the machine" and "the same run has
  completed the first install" were both wrong keys: the gate is per-asset
  presence, not a run-level flag): the phase branches on project provenance
  AND on per-asset user-side presence. A project file is removed only when
  its user counterpart is CONFIRMED PRESENT — manifest-tracked with a
  current hash matching the installed bytes (REQ-020). The machine states:
  - (a) manifest ALREADY EXISTS (written by another project's update, or by
    a partial install) on a machine whose project still carries prior-model
    assets: the arm does not stall — missing counterparts are installed
    append-only (REQ-021-safe), present-and-matching counterparts are reused
    as-is, and removal proceeds per-asset.
  - (b) optional-pack assets in the upgrading project: the upgrade first
    install covers L0 ONLY (no manifest, no `--bundles` → the selection is
    L0 per §2.3's default). A non-L0 project asset has no confirmed user
    counterpart, so it STAYS project-side (reported) until the user opts
    into its bundle — never silently installed into a bundle the user did
    not pick, never removed while unconfirmed.
  - (c) partial first-install failure (REQ-013 continues past per-file
    failures): a failed user-side write leaves that file's project
    counterpart UN-REMOVED (its presence is unconfirmed), and the summary
    reports the skipped removal; on a retry the unconfirmed counterparts are
    re-attempted and the removal completes only for files whose counterparts
    then confirm.
  - No manifest AND no prior-model project assets → the advisory; there is
    nothing to install from and nothing to strand (the first install belongs
    to init, REQ-024).
- Ordering: the user-asset phase runs BEFORE the project phase so a
  user-side failure changes no project behavior beyond the per-asset gate
  itself; within the upgrade case the ordering PLUS the per-asset gate is
  the stranding guard — a project file's removal runs only after its own
  counterpart's install has landed and confirmed, so "a mid-update failure
  leaves the project phase untouched" reads per-asset: the affected file's
  counterpart stays project-side, the rest of the phase proceeds.

### 2.5 Project slimming

- Project deploy stops emitting common skills/agents entirely (both former
  modes): the project keeps settings, AGENTS.md/CLAUDE.md, the lock file
  (`.moai/manifest.json` unchanged in role), hooks, `.mcp.json`, output-styles,
  rules, command wrappers. "Command wrappers" names NON-SKILL command files
  only (iter4 D29): the 17 published Codex command skills (research V4) are
  common skills and move to user folders — the moai-plan/moai-run/moai-sync
  three via L0 (D-Q4), the remaining fourteen via the D-Q5 re-bundling (§2.3)
  — so AC-011's `no .agents/skills/moai*` placement ban holds unchanged over
  the post-M4 project tree.
- THE CLASS CLAUSE (final-class round, operator wording, verbatim — the
  governing principle for this ENTIRE rebind family, replacing
  layer-by-layer enumerations): **"every project-relative reference in the
  user-scope deployed tree rebinds to its installed location, verified by a
  raw-pattern sweep + run-phase loading ACs."** Two instruments, by design
  so the family cannot re-sprout: (a) the RAW-PATTERN SWEEP — a boundary
  grep over the user-scope deployed sources
  (`templates/.claude/skills/**`, `templates/.agents/skills/**`) for the
  project-relative path patterns (`.claude/`, `.agents/`, `.codex/`
  path-shaped references) — CLASSIFIED INTO TWO POPULATIONS THAT ARE
  CHECKED DIFFERENTLY (directed repair R-a, v0.6.1): (1) MOVED-ASSET
  references — skills/agents paths M4 relocates user-side — are REWRITTEN,
  and zero-hit-after-rebind is measured on THIS POPULATION ONLY; (2)
  PROJECT-RETAINED references — rules/settings/hooks that STAY
  project-side — are verified PRESENT-AND-CORRECT at their project paths
  and are NEVER rewritten and NEVER swept to zero (measured instances:
  `workflows/run/phase-execution.md:208-210` `.claude/rules/moai/
  languages/*.md`, run.md's `.claude/rules` references (13 measured this
  tree) + the `trace-ledger.sh` hook call (:28), `sync.md:43` the
  quality-gate hook, `plan.md:47` the spec-workflow pointer); the sweep
  evidence records BOTH populations and the scoping decision; (b) the
  RUN-PHASE LOADING ACs (AC-002/AC-017 arms). Standing rule (operator,
  recorded in the verdict file): any NEW gate finding after this round is
  run-phase debt — no further plan folds.
- REWRITE-SCOPE EXTENSION (class repair R-d): the rebind/rewrite-check
  target is the WHOLE user-scope deployed surface, not only workflow
  markdown — it extends to the SKILLS TREE, the AGENT SOURCES
  (`templates/.claude/agents/moai/*.md`), and the GENERATED TOMLs they
  emit, with `make agents-emit` regeneration in the M4 procedure (the
  agentemit pipeline: the .md layer is the source, the .codex TOMLs are
  golden-pinned outputs — a source edit obliges the explicit
  regeneration). Measured instance: `e2e-tester.md:141` carries a
  moved-asset reference (`.claude/skills/moai-workflow-testing/references/
  e2e-desktop-native-recipes.md`) — the emitted TOMLs inherit it, so the
  rebind at source + agents-emit is required or the Codex-side agent
  keeps the stale project path. Two more same-class instances enumerated
  inside the sweep: (i) RETAINED-FILE→MOVED-ASSET references get
  REWRITTEN, not merely verified — `templates/CLAUDE.md:31` and `:47`
  route `.claude/skills/moai/SKILL.md`; CLAUDE.md itself STAYS
  project-side, but its moved-asset references rebind to the installed
  location, with post-rewrite verification; (ii) DOCTOR CONSUMER
  completion — `runHarnessCheck`'s L4 inspects the PROJECT
  `.claude/skills/moai/workflows` directory
  (`internal/cli/doctor_harness.go:20`, L4 wiring at `:70`) and joins the
  M4 repoint list alongside `checkSkillsAllowlist` and the Codex mirror
  diagnostics, with a healthy-install regression test: post-migration,
  L4 must PASS with the workflows read from the user folder.
- Sub-workflow recursive rebind (final-class item 1): the class clause
  reaches the dispatcher's POINTED-TO documents — the workflows tree
  itself carries project-relative references (top-level `.md` files with
  skill-path references measured: 12 files / 32 narrow-pattern
  occurrences, plan.md and sync.md among them; the run/ and sync/
  subdirectory step documents included in the broad-pattern sweep), so the
  rebind is RECURSIVE through the workflows tree, and AC-017's executable
  arm extends to end-to-end STEP-DOCUMENT loading (the run/sync flow's
  sub-documents load from the user folders), not merely the dispatcher
  body.
- Command-skill reference widening (final-class item 2): the dispatcher
  reference lives in ALL SEVENTEEN generated copies (measured 17/17) —
  the round-5 fix naming only {plan,run,sync} was instance-scoped. Source
  split (measured): 13 command sources carry the literal
  (`.claude/commands/moai/{clean,codemaps,e2e,feedback,fix,gate,harness,
  loop,mx,plan,project,review,run}.md`) and rebind at source; the
  remaining four (goal, gtd, sync, todo) carry the line via the EMITTER's
  injected fallback (their sources lack the literal) and rebind through
  the emitter's injected-line template. `make commands-emit` regenerates
  all seventeen.
- Mirror-repair rollback termination (final-class item 6):
  `runUpdate` calls `repairSkillMirrorBestEffort()` OUTSIDE deploy
  (`internal/cli/update.go:535`), and `RepairSkillMirror`
  (`internal/template/skill_mirror_repair.go:89,:113`) Path B re-creates
  the seventeen published `.agents/skills/moai-<command>/SKILL.md` files
  restore-missing-only — so a repeated `moai update` after M4 actively
  re-grows the project placement M4 removed. M4 TERMINATES this path: the
  update-time repair call is removed (the mirror concept it served is
  retired with the project-side placement; no user-side equivalent is
  needed — the user folders are the primary, not a mirror), with a
  repeated-update regression test (AC-011 arm).
- Manager-git policy (final-class item 4 — the SPEC's design decision):
  Route B (Tier L OR explicit `--pr`) of the sync delivery
  (`workflows/sync/delivery.md` Route B row) invokes `manager-git`, which
  is NOT one of the L0 five (D-Q1). CHOSEN POLICY: REQUIRE-A-BUNDLE, not
  re-route — re-routing Tier L git operations to an installed role
  (manager-docs) would blur the DRI the agent catalog owns (PR/branch
  delivery specialty), and D-Q1's five-agent answer is settled. manager-git's
  role body ships in an OPT-IN bundle (the D-Q5 re-bundling assigns it —
  the git/delivery theme), and the Route B rows gain an entry PRECONDITION:
  the flow verifies the manager-git role body is installed and, when it is
  not, refuses with the named remediation `moai bundle add <bundle>` per
  C4's actionable-report rule — never a missing-file error mid-flow. The
  precheck covers ALL manager-git ENTRY POINTS (directed repair R-b,
  v0.6.1): the sync delivery Route B row AND the run flow's Route B
  (`workflows/run/task-decomposition.md:302-303` — "Route B — Tier L OR
  explicit `--pr`: Agent: manager-git subagent", wired at Phase 19) — the
  same verify-or-refuse + remediation on both. The default Route A flow
  (manager-docs + lane self-delivery) needs nothing beyond L0. Anchors:
  design §2.3 (bundle assignment), plan M3 (the
  precondition check rides the `moai bundle` command milestone), AC-018
  (the requires-unopted-bundle arm over BOTH entry points).
- Dispatcher reference rebind at SOURCE level (round-5 F2): the published
  command skills are GENERATED artifacts — the command sources under
  `.claude/commands/moai/` are consumed READ-ONLY by the
  `internal/template/commandemit` emitter (`CommandsRoot:
  ".claude/commands/moai"`, commandemit.go:51), whose golden-pinned output
  is the committed `templates/.agents/skills/moai-<command>/SKILL.md` set;
  the project-relative `read .agents/skills/moai/SKILL.md` instruction
  (moai-plan/moai-run/moai-sync SKILL.md:9) is therefore fixed in the
  COMMAND SOURCES (to the user-folder path `~/.agents/skills/moai/
  SKILL.md`) and the copies are regenerated with `make commands-emit` —
  never hand-edited; the read-only drift check `commands-emit-check` rides
  the `build:` prerequisite chain (Makefile:34, :51-60) and would reject a
  hand-edited copy at the next build. The AGENTS.md skill-path sentences
  (AGENTS.md.tmpl:40-41 — "the deployed skill is in `.agents/skills/<name>/
  SKILL.md` for Codex and `.claude/skills/<name>/SKILL.md` for Claude")
  rebind the same way at source: post-M4 the Codex sentence names the
  user-folder dispatcher path.
- Migration of EXISTING projects (REQ-020): update classifies current
  project-local `moai-*` skills/agents via the provenance classes —
  `template_managed` → removable (offer + apply), `user_modified` → preserved
  with a report (3-way decision is NOT silent), `user_created` → untouched.
  The existing namespace contract (V12) is the classifier.

### 2.6 `moai doctor` integration

- New checks (workspace group): (1) user-install integrity — every
  manifest-tracked path exists with matching hash; untracked files in the four
  roots listed as informational; (2) project-vs-lock comparison — drift in
  both directions (project file absent from lock; lock entry absent from
  project). Both read-only, report-only (matching the doctor house style).
- Retired-carrier checks (`checkPluginDeployment`, `checkPluginVersion`) are
  repointed to the user-manifest comparison or removed with their SPEC's REQs
  cited in the commit (REQ-019). The existing project-scope Codex asset
  diagnostics are repointed to the user-install path in M4 — the same
  milestone that stops emitting project assets — with a clean-on-correct-
  install regression test extending to the readiness output (iter1 D12
  residue, corrected by iter2 D19): the PROJECT-root readers are
  `inspectSkillMirror` (`internal/cli/doctor_codex.go:429`) and the readiness
  pair `probeCodexReadiness`/`countCodexAgentTOMLs`
  (`internal/cli/codex_readiness.go:131` consumer, `:215-217` definition —
  the agent-TOML count reads `.codex/agents/moai/*.toml` under the project
  root). `codexStaleSkillFinding` (`doctor_codex.go:857-870`) is NOT one of
  them: it reads user-layer `[[skills.config]]` entries and has no
  agent-count input — iter1's attribution to it was wrong; M5 judges whether
  that user-layer check needs its own repoint. `checkSkillsAllowlist`
  (`internal/cli/doctor.go:957-958`) joins the repoint list (round-5 fold
  A3): it reads `.claude/skills` under the PROJECT root and warns
  ".claude/skills/ not found" on a healthy post-M4 install — it is
  REPOINTED to the user-install path in M4 (the same milestone that stops
  emitting project assets), not removed, so the allowlist integrity check
  survives scoped to the user folders. `runHarnessCheck`'s L4 (which
  inspects the project `.claude/skills/moai/workflows` directory —
  `internal/cli/doctor_harness.go:20`, L4 wiring `:70`) joins the same
  repoint list (class repair R-d, instance ii): post-migration L4 must
  PASS with the workflows read from the user folder, covered by a
  healthy-install regression test. Left alone the project-root
  readers — the allowlist check, and the harness L4 — would misreport
  every correct install as drift.

### 2.7 Plugin carrier disposition

- `internal/template/pluginemit/` (generator + 8 test files; 13 .go files
  total), the committed `plugins/moai/` tree, `.claude-plugin/marketplace.json`,
  `.agents/plugins/marketplace.json`, Makefile `plugin-emit`/`plugin-emit-check`
  targets (and their `build:` prerequisite), the roster-guard sweep-skip note
  (`internal/harness/rosterguard/check.go:338-341`), and the plugin install
  step (`internal/cli/plugin_install.go` + init/update call sites) go away.
- The release-chain gates go with them (iter1 D4): release.yml provenance
  Check 8 (`release.yml:123-128`, invoking `scripts/check-plugin-version.sh`,
  citing SPEC-PLUGIN-MARKETPLACE-001 REQ-024), `scripts/check-plugin-version.sh`
  (reads `plugins/moai/.claude-plugin/plugin.json`), and
  `scripts/check-plugin-discoverable.sh` (reads `.claude-plugin/marketplace.json`
  + `plugins/moai/.mcp.json`) — all deleted with the carrier (hard delete,
  P5); no orphaned release check remains (C5). SPEC-PLUGIN-MARKETPLACE-001
  REQ-024 is recorded as retired in spec.md §7.
- `deployer_mode.go`'s split (DeployModePlugin, PluginMirrorPolicy, the
  exclusion walk, the MCP strip, the plugin re-home) is retired; the deployer
  returns to a single project payload shape (the slim one).
- `.mcp.json` rendering returns to always carrying the `moai` entry (the
  plugin was its alternative carrier).

## 3. Sequencing Logic (why the milestones order this way)

Decision-reversibility: M0/M1 pin the data model (bundle taxonomy + manifest
schema + the two decision gates D-Q1/D-Q2) — the highest-change-likelihood
decisions, reviewed first. M2-M5 build the new behavior on the frozen schema.
M6/M7 are the demolitions, last, because they are the milestones the operator
would drop if the direction reverses, and because deleting the carrier before
the replacement installs would strand users between milestones. M8 closes with
cross-harness evidence.

## 4. Risks and Mitigations

| Risk | Mitigation |
|---|---|
| Wide retire blast radius (8 test files / 13 .go files, build chain, roster guard, release chain) | M6/M7 isolated; each deletion cites the owning REQ; golden tests deleted WITH the generator in the same commit; release-chain gates dispositioned in M6 (iter1 D4) |
| Users left with plugin installs from the old model | Migration report names the manual `claude plugin uninstall` step (doctor informational row); no silent divergence |
| Profile sessions not seeing user assets (D-Q3 closed: P6) | Declared limitation documented in M8; doc-visible so a later SPEC can lift it |
| User edits to tracked files silently overwritten or deleted (iter1 D3) | Refresh gated on current-hash == manifest-hash; removal on the one REQ-009 rule (manifest-hash OR shipped-bytes where a shipped source exists — iter4 D25); divergence → preserve (backup) + report + a REQ-011 count category; doctor's "modified" row lists it persistently |
| Manifest corruption / partial write | Atomic write (same pattern as `atomicWriteFile`); schema-version refusal for removal; corrupt-JSON recovery path (AC-021) |
