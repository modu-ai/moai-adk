# Decision Index — SPEC-USER-SETTINGS-PROTECT-001

This index is stateless on the status axis. It carries no status field; the SPEC lifecycle lives in spec.md alone. It is required because `interview.decision_gate` is `on` (plan.md E-17).

Each row states what is unresolved and why, and carries no recommendation for a judgment call. Routing uses exactly four labels: DECIDED, POLICY-COVERED, EVIDENCE-NEEDED, FOUNDER. A `Default:` line appears only where the published Default rule ranks the options. That line is a policy application, not a recommendation. FOUNDER rows of class product-level are never applied automatically. The `Operator verdict:` line is empty at authoring, except where the Default rule applies to an implementation-level row at plan close.

Authority register: committed artifacts only (prior completed SPECs' HISTORY and Amendments rows, `.moai/config/sections/*.yaml`, the constitution, `.moai/project/product.md`). Requirement text is not in the register. Untracked material (for example `.moai/reports/`) is not cited as authority.

## Summary

| Row | Label | Class | Effect on run |
|---|---|---|---|
| Q1 | FOUNDER | product-level | blocks the init writer's defaultMode case (REQ-003) |
| Q2 | FOUNDER | product-level | blocks the t1567 template item (REQ-004) |
| Q3 | FOUNDER | product-level | blocks the run and db candidate rule (REQ-011) |
| Q4 | EVIDENCE-NEEDED | — | run-phase observation before the profile question closes |
| Q5 | FOUNDER | product-level | blocks the PROJECT-scope allow case (REQ-005) |
| Q6 | EVIDENCE-NEEDED | — | run-phase observation before the force question closes |
| Q7 | FOUNDER | implementation-level | default applied at plan close |
| Q8 | DECIDED | — | restores the committed contract (REQ-001) |

---

### Q1: When `moai init` finds an existing USER-scope `permissions.defaultMode` that differs from the resolved tier default, should it overwrite the value with the tier default or keep the existing value?

Label: FOUNDER
Authority anchor: none verifiable for this case. Checked: SPEC-AUT-PERMMODES-001 REQ-004 (its spec.md, the text at line 83) sets the semi-auto write to `defaultMode: "acceptEdits"` and nothing else; SPEC-INIT-WIZARD-REPAIR-001 §4 limits the write to the `defaultMode` key; no HISTORY row decides the case of an existing differing value (plan.md E-23).
Why unresolved: the card's judgement requires a zero diff in the permissions section, which holds only when the existing value already equals the tier default. The committed text sets the key but not the disposition of an existing differing value, and the card and the committed text point in different directions on that case.
Class: product-level (changes the default user-visible behaviour of `moai init`)
Default: overwrite the existing value with the tier default (rule: preserves current behavior; the current writer overwrites, plan.md E-3)
Alternate: keep the existing value
Operator verdict:

### Q2: Should the settings template ship a `permissions.defaultMode` value (card t1567)?

Label: FOUNDER
Authority anchor: none verifiable. The t1567 card text is not in the tree (plan.md G-1). Two code comments state that the template stopped shipping a defaultMode default (plan.md E-13), and no committed HISTORY row decides it.
Why unresolved: the card names the change, but its text is absent from the tree. A template default is a product-level choice.
Class: product-level (changes a template default)
Default: ship no defaultMode in the template (rule: preserves current behavior; the template has none, plan.md E-12)
Alternate: ship a defaultMode value
Operator verdict:

### Q3: Which entries under `run/` and `db/` may `moai clean --home` treat as deletable candidates, and under which age or liveness rule?

Label: FOUNDER
Authority anchor: none verifiable for the run and db rule. The home-retention setting covers the existing categories (`internal/cli/clean_home.go`, the `state.home_retention_days` loader), but no committed HISTORY row decides these two. The db area holds the contract store (`internal/contract/receipt/dir.go:19`).
Why unresolved: the card requires coverage. Which entries are safe to delete is a product-level choice, and a wrong setting deletes evidence.
Options: (a) run/ entries by age; db/ entries listed but never deleted. (b) run/ and db/ entries by age. (c) run/ and db/ entries deletable only when no live record references them.
Class: product-level (changes the default dry-run output and the deletion scope of `moai clean --home`)
Default: option (a) (rule: smaller user-visible surface; steps one and two of the Default rule do not separate the options, because no option preserves current behaviour under the card's requirement, and each option's undo is a single revert)
Alternate: option (b)
Operator verdict:

### Q4: Should the USER-scope write target the settings file of the active `CLAUDE_CONFIG_DIR` profile instead of `<home>/.claude/settings.json`?

Label: EVIDENCE-NEEDED
Authority anchor: none verifiable. SPEC-USER-ASSET-INSTALL-001 REQ-002 states the per-profile rule, but requirement text is outside the authority register.
Why unresolved: which file a profile-launched session reads was not measured in plan phase. The init call site builds the path from the home directory (plan.md E-5).
Evidence needed: a run-phase observation of the settings file a profile-launched session reads, taken on a throwaway profile, with the path compared to the path the writer uses.
Operator verdict:

### Q5: For a project that carries a tool-policy document, should the PROJECT-scope full-bundle path keep user-added `permissions.allow` entries or regenerate the list from the document?

Label: FOUNDER
Authority anchor: none for `allow`. SPEC-INIT-WIZARD-REPAIR-001 §4 states that the full-bundle path regenerates `deny` and `ask` within the same block (per SPEC-AUTONOMY-TIERS-001 REQ-003), and it does not address `allow` (plan.md E-23).
Why unresolved: the current code sets the PROJECT block's `Allow` from the document (`internal/config/toolpolicy/tier_render.go:76-81`, plan.md G-12). Keeping user entries changes a maintainer path.
Class: product-level (changes the behaviour of `moai init` for projects with a policy document)
Default: regenerate the list from the document (rule: preserves current behavior)
Alternate: merge user-added entries into the regenerated list
Operator verdict:

### Q6: Does `moai init --force` overwrite an existing project `.claude/settings.json` permissions block?

Label: EVIDENCE-NEEDED
Authority anchor: none verifiable.
Why unresolved: not measured. The deployer's force-update mode overwrites existing files without a manifest check (`internal/template/deployer.go:80` and `:254`), but that is the update path, not a measured init run.
Evidence needed: a run-phase `moai init --force` on a throwaway project with a sandboxed home, comparing the permissions block before and after.
Operator verdict:

### Q7: Which mechanism installs the test HOME and MOAI_HOME sandbox: a shared helper called from each package's TestMain (the card's wording), or a fail-closed guard inside the single home resolver?

Label: FOUNDER
Class: implementation-level
Authority anchor: none verifiable for the choice. The current practice is TestMain sandboxes in `internal/cli` and `internal/hook` (plan.md E-8).
Why unresolved: both mechanisms are implementation options with no user-visible effect. The choice is recorded because it changes which files the run touches (plan.md §A.4 Basis).
Default: shared helper called from each TestMain (rule: preserves current behavior; it follows the TestMain pattern that `internal/cli` and `internal/hook` already use)
Alternate: fail-closed guard inside the home resolver
Operator verdict: DEFAULT-APPLIED 2026-10-10T07:02:43Z manager-spec (plan close)

### Q8: Must the USER-scope settings write preserve the user's `allow`, `ask`, and `deny` lists and every other key verbatim on the distributed-default path?

Label: DECIDED
Authority anchor: `.moai/specs/SPEC-INIT-WIZARD-REPAIR-001/spec.md`, HISTORY row "0.1.1" (2026-08-22), which pins the "§4 key-scoped USER-write constraint". Section §4 (line 96) states that the distributed-default path writes exactly the `permissions.defaultMode` key and "preserves every other region (PATH, hooks, env, allow, deny, ask) verbatim", and that whole-file overwrite is prohibited. The file is tracked (plan.md E-23).
Why unresolved: not open. Recorded because the current writer violates the pinned condition (plan.md E-3 and E-4), so the run restores the committed contract rather than deciding it.
Operator verdict:
