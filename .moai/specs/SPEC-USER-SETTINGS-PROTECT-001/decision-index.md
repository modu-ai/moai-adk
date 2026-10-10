# Decision Index — SPEC-USER-SETTINGS-PROTECT-001

This index is stateless on the status axis. It carries no status field; the SPEC lifecycle lives in spec.md alone. It is required because `interview.decision_gate` is `on` (plan.md E-17).

Each row states what is unresolved and why, and carries no recommendation for a judgment call. Routing uses exactly four labels: DECIDED, POLICY-COVERED, EVIDENCE-NEEDED, FOUNDER. A `Default:` line appears only where the published Default rule ranks the options, and no row in this revision carries one, because no Default-rule clause is cited here. FOUNDER rows of class product-level are never applied automatically. The `Operator verdict:` line is empty at authoring for FOUNDER and EVIDENCE-NEEDED rows; DECIDED rows name the decider and the pinned citation.

Authority register: committed artifacts only (prior completed SPECs' HISTORY and Amendments rows, `.moai/config/sections/*.yaml`, the constitution, `.moai/project/product.md`), plus a decision-board record only when it is pinned: cited as `board:<record-id>#<digest-prefix>` with the cited line copied verbatim in a fenced block. Requirement text is not in the register. Untracked material (for example `.moai/reports/`) is not cited as authority.

Digest-prefix (definition approved by the leader in ruling `board:d-20261010T075613Z-5e49`): the first 12 hex characters of the sha256 over the bytes of the cited verbatim file. For `board:d-20261010T073547Z-07ae` the verbatim file is `.moai/reports/t1630/ruling-d-20261010T073547Z-07ae.txt`, whose sha256 begins `9c93e47809f9`. That file holds the header line only. The decisions sit in the record's body line, which each DECIDED row copies verbatim beside the header. The body line was matched to the board listing as a whole line (`moai decision read --all`, grep -x -F, one match). The 12-hex digest does not cover the body line, and the board listing carries no digest field, so the digest identifies the lane's copy rather than the board's record (plan.md G-18).

## Summary

| Row | Label | Class | Effect on run |
|---|---|---|---|
| Q1 | DECIDED | product-level | an existing USER-scope defaultMode is kept and written only when absent; supersedes SPEC-AUT-PERMMODES-001 spec.md:83 for the defaultMode write path only (spec.md §2 note on REQ-003; AC-001, AC-003 part B) |
| Q2 | DECIDED | product-level | the template carries a defaultMode; a fresh init writes it only when absent; update never touches it (spec.md §2 note on REQ-004; AC-004). The value is `"default"` (lane pin, progress.md §G; plan.md G-17) |
| Q3 | DECIDED | product-level | Out of Scope - moved to t1666 (REQ-011, REQ-012, AC-011); row text retained: run items are candidates only when no live record references them; db items are listed and deleted only with an explicit --yes, which is the existing `--force` flag (ruling d-20261010T081810Z-49e5; G-19) |
| Q4 | EVIDENCE-NEEDED | — | does not block run; closed by the evidence produced at M2 |
| Q5 | DECIDED | product-level | user-added permissions.allow entries are kept by set union (REQ-005, AC-005). Detection is diff-based against a sidecar record (lane pin, progress.md §G; G-20) |
| Q6 | EVIDENCE-NEEDED | — | does not block run; closed by the evidence produced at M2 |
| Q7 | FOUNDER | implementation-level | Out of Scope - moved to t1666 (sandbox mechanism; REQ-006 to REQ-010); row text retained: open, no default applied |
| Q8 | DECIDED | — | restores the committed contract (REQ-001, AC-001) |

---

### Q1: When `moai init` finds an existing USER-scope `permissions.defaultMode` that differs from the resolved tier default, should it overwrite the value with the tier default or keep the existing value?

Label: DECIDED
Authority anchor: `board:d-20261010T073547Z-07ae#9c93e47809f9` (decision-board record; decided_by 영실이 판단, operator-delegated). Pinned lines, copied verbatim:

```text
decision record: decided_by=영실이 판단(운영자 위임) — session 영실이 (moai-adk-go-a0) under 3db2 and d37b; recorded by leader session 359ce07d evidence_refs=.moai/worktrees/t1630/.moai/reports/t1630/operator-brief.md (leader read); 3.2 plan G5 'init/update가 사용자 설정 보존' and the 0-1 completion check 'init 전후 settings diff 0' (영실이 cited) ladder_path=delegated decision id=d-20261010T073547Z-07ae scope=card:t1630 kind=ruling
  body: 영실이 판단(운영자 위임) 10-10 — Q1 b · Q2 b(부재 시만) · Q3 c(+db 목록만) · Q5 b. Principle: a value the user changed is never overwritten on any path. Q1 (b): an existing defaultMode is kept; the new SPEC states that it supersedes SPEC-AUT-PERMMODES-001 spec.md:83 ('NOTHING else'), and tier_render.go:106 writes only when the value is absent. Q2 (b): the template carries the default; a fresh init writes it only when absent; update never touches it (t1567's intent: prevent sessions starting in auto mode when defaultMode is unset). Q3 (c): run/ items are delete candidates only when no live record (active run, lease, unclosed card) references them; db/ is listed only and deleted only with an explicit --yes. Q5 (b): user-added permissions.allow entries are kept by set union with the regenerated managed block (marker- or diff-based) and are removed only by the user. The run phase stays after t1578, t1591 and t1619 land.
```

Decided behaviour (body line, Q1 (b)): an existing USER-scope `defaultMode` is kept, and the writer writes the key only when it is absent. The new SPEC supersedes SPEC-AUT-PERMMODES-001 spec.md:83 ("NOTHING else") for the defaultMode write path only, as the supersession note in the spec.md §2 note on REQ-003 states. The record's remark that tier_render.go:106 "writes only when the value is absent" describes the rule, not the current code: the current writer overwrites the value (AC-003 part B RED cell, plan.md G-22).
Why unresolved: not open. Recorded because the record decides the question; the run implements the decided rule.
Class: product-level (changes the default user-visible behaviour of `moai init` for an existing differing value)
Operator verdict: decided_by 영실이 판단(운영자 위임); pinned citation board:d-20261010T073547Z-07ae#9c93e47809f9.

### Q2: Should the settings template ship a `permissions.defaultMode` value (card t1567)?

Label: DECIDED
Authority anchor: `board:d-20261010T073547Z-07ae#9c93e47809f9` (decision-board record; decided_by 영실이 판단, operator-delegated). Pinned lines, copied verbatim:

```text
decision record: decided_by=영실이 판단(운영자 위임) — session 영실이 (moai-adk-go-a0) under 3db2 and d37b; recorded by leader session 359ce07d evidence_refs=.moai/worktrees/t1630/.moai/reports/t1630/operator-brief.md (leader read); 3.2 plan G5 'init/update가 사용자 설정 보존' and the 0-1 completion check 'init 전후 settings diff 0' (영실이 cited) ladder_path=delegated decision id=d-20261010T073547Z-07ae scope=card:t1630 kind=ruling
  body: 영실이 판단(운영자 위임) 10-10 — Q1 b · Q2 b(부재 시만) · Q3 c(+db 목록만) · Q5 b. Principle: a value the user changed is never overwritten on any path. Q1 (b): an existing defaultMode is kept; the new SPEC states that it supersedes SPEC-AUT-PERMMODES-001 spec.md:83 ('NOTHING else'), and tier_render.go:106 writes only when the value is absent. Q2 (b): the template carries the default; a fresh init writes it only when absent; update never touches it (t1567's intent: prevent sessions starting in auto mode when defaultMode is unset). Q3 (c): run/ items are delete candidates only when no live record (active run, lease, unclosed card) references them; db/ is listed only and deleted only with an explicit --yes. Q5 (b): user-added permissions.allow entries are kept by set union with the regenerated managed block (marker- or diff-based) and are removed only by the user. The run phase stays after t1578, t1591 and t1619 land.
```

Decided behaviour (body line, Q2 (b)): the template carries the default; a fresh `moai init` writes it only when absent; `moai update` never touches it. The record gives the purpose: t1567's intent is to prevent sessions starting in auto mode when defaultMode is unset.
Not named by the record: the literal value of the template default. The lane pins it to `"default"` (progress.md §G; plan.md G-17); that pin is a lane decision, not an authority anchor.
Why unresolved: not open as to whether the template ships a value. The literal value is pinned by the lane (G-17).
Class: product-level (changes a template default)
Operator verdict: decided_by 영실이 판단(운영자 위임); pinned citation board:d-20261010T073547Z-07ae#9c93e47809f9.

### Q3: Which entries under `run/` and `db/` may `moai clean --home` treat as deletable candidates, and under which age or liveness rule?

Out of Scope - moved to t1666 by the operator-delegated split recorded in spec.md HISTORY (revision 2): this decision moves with REQ-011, REQ-012, and AC-011. The text below is retained unchanged.

Label: DECIDED
Authority anchor: `board:d-20261010T073547Z-07ae#9c93e47809f9` (decision-board record; decided_by 영실이 판단, operator-delegated). Pinned lines, copied verbatim:

```text
decision record: decided_by=영실이 판단(운영자 위임) — session 영실이 (moai-adk-go-a0) under 3db2 and d37b; recorded by leader session 359ce07d evidence_refs=.moai/worktrees/t1630/.moai/reports/t1630/operator-brief.md (leader read); 3.2 plan G5 'init/update가 사용자 설정 보존' and the 0-1 completion check 'init 전후 settings diff 0' (영실이 cited) ladder_path=delegated decision id=d-20261010T073547Z-07ae scope=card:t1630 kind=ruling
  body: 영실이 판단(운영자 위임) 10-10 — Q1 b · Q2 b(부재 시만) · Q3 c(+db 목록만) · Q5 b. Principle: a value the user changed is never overwritten on any path. Q1 (b): an existing defaultMode is kept; the new SPEC states that it supersedes SPEC-AUT-PERMMODES-001 spec.md:83 ('NOTHING else'), and tier_render.go:106 writes only when the value is absent. Q2 (b): the template carries the default; a fresh init writes it only when absent; update never touches it (t1567's intent: prevent sessions starting in auto mode when defaultMode is unset). Q3 (c): run/ items are delete candidates only when no live record (active run, lease, unclosed card) references them; db/ is listed only and deleted only with an explicit --yes. Q5 (b): user-added permissions.allow entries are kept by set union with the regenerated managed block (marker- or diff-based) and are removed only by the user. The run phase stays after t1578, t1591 and t1619 land.
```

Decided behaviour (body line, Q3 (c), with db/ listed only): a `run/` item is a delete candidate only when no live record (active run, lease, unclosed card) references it. A `db/` item is listed only, and is deleted only with an explicit `--yes`.
Not named by the record: the current deletion flag is `--force` (`internal/cli/clean.go:106`), and the CLI has no `--yes` flag. The record's term is transcribed as written. Ruling d-20261010T081810Z-49e5 (item 2) reads the record's `--yes` as the existing `--force` flag, with no new flag (plan.md G-19).
Why unresolved: not open as to the candidate rule. The flag naming is resolved by ruling d-20261010T081810Z-49e5 (G-19).
Class: product-level (changes the default dry-run output and the deletion scope of `moai clean --home`)
Operator verdict: decided_by 영실이 판단(운영자 위임); pinned citation board:d-20261010T073547Z-07ae#9c93e47809f9.

### Q4: Should the USER-scope write target the settings file of the active `CLAUDE_CONFIG_DIR` profile instead of `<home>/.claude/settings.json`?

Label: EVIDENCE-NEEDED
Authority anchor: none admitted. SPEC-USER-ASSET-INSTALL-001 REQ-002 states the per-profile rule, but requirement text is outside the authority register.
Why unresolved: which file a profile-launched session reads was not measured. The init call site builds the path from the home directory (plan.md E-5).
Evidence needed: a run-phase observation, at M2, of the settings file a profile-launched session reads, taken on a throwaway profile, compared with the path the writer uses. Closure happens at M2 and does not block run.
Operator verdict:

### Q5: For a project that carries a tool-policy document, should the PROJECT-scope full-bundle path keep user-added `permissions.allow` entries or regenerate the list from the document?

Label: DECIDED
Authority anchor: `board:d-20261010T073547Z-07ae#9c93e47809f9` (decision-board record; decided_by 영실이 판단, operator-delegated). Pinned lines, copied verbatim:

```text
decision record: decided_by=영실이 판단(운영자 위임) — session 영실이 (moai-adk-go-a0) under 3db2 and d37b; recorded by leader session 359ce07d evidence_refs=.moai/worktrees/t1630/.moai/reports/t1630/operator-brief.md (leader read); 3.2 plan G5 'init/update가 사용자 설정 보존' and the 0-1 completion check 'init 전후 settings diff 0' (영실이 cited) ladder_path=delegated decision id=d-20261010T073547Z-07ae scope=card:t1630 kind=ruling
  body: 영실이 판단(운영자 위임) 10-10 — Q1 b · Q2 b(부재 시만) · Q3 c(+db 목록만) · Q5 b. Principle: a value the user changed is never overwritten on any path. Q1 (b): an existing defaultMode is kept; the new SPEC states that it supersedes SPEC-AUT-PERMMODES-001 spec.md:83 ('NOTHING else'), and tier_render.go:106 writes only when the value is absent. Q2 (b): the template carries the default; a fresh init writes it only when absent; update never touches it (t1567's intent: prevent sessions starting in auto mode when defaultMode is unset). Q3 (c): run/ items are delete candidates only when no live record (active run, lease, unclosed card) references them; db/ is listed only and deleted only with an explicit --yes. Q5 (b): user-added permissions.allow entries are kept by set union with the regenerated managed block (marker- or diff-based) and are removed only by the user. The run phase stays after t1578, t1591 and t1619 land.
```

Decided behaviour (body line, Q5 (b)): user-added `permissions.allow` entries are kept by set union with the regenerated managed block, and are removed only by the user. The record leaves the managed-block detection open ("marker- or diff-based"); the lane pins it to diff-based detection against a sidecar record (progress.md §G; plan.md G-20).
Why unresolved: not open as to retention. The detection method is pinned by the lane (G-20).
Class: product-level (changes the behaviour of `moai init` for projects with a policy document)
Operator verdict: decided_by 영실이 판단(운영자 위임); pinned citation board:d-20261010T073547Z-07ae#9c93e47809f9.

### Q6: Does `moai init --force` overwrite an existing project `.claude/settings.json` permissions block?

Label: EVIDENCE-NEEDED
Authority anchor: none admitted.
Why unresolved: not measured. The deployer's force-update mode overwrites existing files without a manifest check (`internal/template/deployer.go:80` and `:254`), but that is the update path, not a measured init run.
Evidence needed: a run-phase `moai init --force` on a throwaway project with a throwaway home, comparing the permissions block before and after. Closure happens at M2 and does not block run.
Operator verdict:

### Q7: Which mechanism installs the test HOME and MOAI_HOME sandbox: a shared helper called from each package's TestMain (the card's wording), or a fail-closed guard inside the single home resolver?

Out of Scope - moved to t1666 by the operator-delegated split recorded in spec.md HISTORY (revision 2): this decision has no subject once REQ-006 to REQ-010 and AC-006 to AC-010 move. The text below is retained unchanged.

Label: FOUNDER
Class: implementation-level
Authority anchor: none admitted for the choice. Current practice: TestMain sandboxes in `internal/cli` and `internal/hook` (plan.md E-8).
Why unresolved: both mechanisms are implementation options with no user-visible effect. The choice changes which files the run touches (plan.md §A.4 Basis). No published Default-rule clause is cited here, so no default is applied; the choice stays open until an operator verdict is recorded.
Operator verdict:

### Q8: Must the USER-scope settings write preserve the user's `allow`, `ask`, and `deny` lists and every other key verbatim on the distributed-default path?

Label: DECIDED
Authority anchor: `.moai/specs/SPEC-INIT-WIZARD-REPAIR-001/spec.md`, HISTORY row "0.1.1" (2026-08-22), which pins the "§4 key-scoped USER-write constraint". Section §4 (line 96) states that the distributed-default path writes exactly the `permissions.defaultMode` key and "preserves every other region (PATH, hooks, env, allow, deny, ask) verbatim", and that whole-file overwrite is prohibited. The file is tracked (plan.md E-23).
Why unresolved: not open. Recorded because the current writer violates the pinned condition (AC-001 RED cell, AC-003 part A RED cell), so the run restores the committed contract rather than deciding it.
Operator verdict:
