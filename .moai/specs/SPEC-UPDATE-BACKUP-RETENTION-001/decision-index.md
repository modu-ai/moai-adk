# Decision Index — SPEC-UPDATE-BACKUP-RETENTION-001

Stateless artifact: it carries no `status:` field; the SPEC's lifecycle lives in `spec.md`.

Routing labels: `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, `FOUNDER`. Every row states what is unresolved and why. No row carries a recommendation for a judgment call; a `Default:` line is the published Default rule applied (preserve current behavior; else single-revert undo; else smaller user-visible surface), not a recommendation.

Authority note. The operator rulings this card rests on (decision board records a263, fd8b, a041, 242e; read with `moai decision read --scope card:t1687 --all`) and the t1687 audit reports under `.moai/reports/t1687/` are not in the committed tree. By the authority register they are evidence, quoted in `research.md` § Operator rulings, and never an authority anchor. A row whose only support is one of them routes `FOUNDER`.

Resolution note (revision 0.2.0, 2026-10-16). Rows Q1, Q2 and Q3 carry the operator's answers (`decided_by=operator`, relayed by the factory leader; verdict source "operator via brain"). The answers are recorded as resolved decisions below; they are not committed authority anchors, so the rows keep the label `FOUNDER` with the verdict recorded in the `Operator verdict:` line. The operator principle for release 3.2 that governs this revision: **the SPEC adds no behavior that deletes data.** Every requirement in `spec.md` was re-read against it (research.md § Principle check).

### Q1: How does the state-residue pass avoid removing a zero-byte database a process holds open (finding F21)?

Label: FOUNDER
Class: product-level
Authority anchor: none
Why unresolved (at authoring): The current pass removes a zero-byte `*.db` under `.moai/state` that has no `-wal` / `-shm` sidecar and has sat unmodified for ten minutes (operator-approved rule, board 73f8; `internal/cli/update_state_residue.go:19`, `:32-62`). A process that opened such a file without writing to it takes `disk I/O error (1802)` on its next write (t1687 audit iter3, A3-P3; cited, not re-measured here). Options were: (a) keep the rule; (b) report zero-byte `.db` files and never delete them; (c) narrow removal to an enumerated legacy list.
Operator verdict: RESOLVED 2026-10-16 — option (b), REPORT ONLY. Do not delete a zero-byte database; surface "empty DB file present" in the doctor / update reporting. (decided_by=operator; verdict source: operator via brain.) Consequence: REQ-UBR-018 is a reporting requirement with no removal; milestone M4 implements it; AC-UBR-18 verifies it. The zero-byte `backlog.json.lock` removal named in the same function is outside this ruling and is unchanged.

### Q2: May protected excess runs exceed keepCount, or must the total run count stay at or under keepCount?

Label: FOUNDER
Class: product-level
Authority anchor: none
Why unresolved (at authoring): Operator ruling a263 (C) says to keep the newest N past versions per file and clean up the rest; the card's completion text says both backup roots end at or under their cap. A protected version may live only in an old run, so the two cannot both hold for every input. Options: (A) soft cap; (B) hard cap with protected versions carried forward into a retained location (writes new content into backups); (C) hard cap with a limit on retained excess runs (data loss beyond the limit).
Operator verdict: RESOLVED 2026-10-16 — option (A), SOFT CAP. The window keeps keepCount runs; where the only protecting version of some file lives only in an older run, that run is kept beyond the window and the excess count is reported. The last copy of a protected file version is never deleted. (decided_by=operator; verdict source: operator via brain.) Reading note: "protected" is the REQ-UBR-003 set (the newest N past versions per path, N = `state.backup_past_versions_keep`, default 3); the operator's budget ruling a263 (C) is not withdrawn, so a version older than the budget is not protected. This reading is stated in the leader report for confirmation; it is not an open marker. Consequence: REQ-UBR-010 and AC-UBR-06 are written for the soft cap, with a verifiable bound (AC-UBR-06 (a)-(c)).

### Q3: Is the retired-skill archive and `.agents/skills` link removal reintroduced?

Label: FOUNDER
Class: product-level
Authority anchor: none
Why unresolved (at authoring): The t1687 progress record handed "reintroduction and repair" of the archive and link-removal path to this card, and board a041 moved F16 and F17 here. Reintroduction changes a shipped command's default user-visible behavior (a folder moves, a link disappears). Options: (A) reintroduce with the root and parent-chain checks; (B) keep report-only.
Operator verdict: RESOLVED 2026-10-16 — option (B), do NOT reintroduce. Report only; the linked or outside-root content is left untouched. The authority basis for any cleanup of retired skill folders and links is deferred to card 3.3. (decided_by=operator; verdict source: operator via brain.) Principle recorded with this answer: in 3.2 the SPEC adds no data-deleting behavior. Consequence: REQ-UBR-015 and REQ-UBR-016 are report-only guards; the earlier archive-path and exact-link-removal requirements are withdrawn; AC-UBR-04 and AC-UBR-05 become regression guards (green now, and red under a mutant that reintroduces removal).
Measurement note: the tree's `skillresidue` package and `update_state_residue.go` do not exist on main `ec5a4f392` (`git diff --stat ec5a4f392 e669da81b` shows both as pure insertions). The report-only behavior is therefore the behavior t1687 landed, not main's.

### Q4: How many excess runs does the irredundant retention leave on the real mo.ai.kr backup roots?

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved (at authoring): Whether the soft cap of Q2 option (A) is acceptable depended on a count that did not yet exist.
Operator verdict: CLOSED 2026-10-16 — not a decision. Q2 is ruled, so no decision depends on the figure (audit D5: the measurement and the ruling are no longer in a before/after relationship). AC-UBR-06 still prints the retained-excess count as a measured output for the record; it gates nothing except the bound stated in AC-UBR-06 (b).

### Q5: Is a backed-up file version keyed by its path inside the backup directory, or by a normalized project-relative path?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: A copy under `pre-clean/.claude/x` and a copy under `.claude/x` in a namespace backup are the same project file at two paths inside backups. Keying by the path inside the backup treats them as different versions; normalizing the `pre-clean/` prefix away would let one satisfy the other. No behavior of the current tree exists to preserve (the tree has no version key at all); the restore path resolves a file by its path inside the backup.
Default: key by the path inside the backup directory (rule: smaller user-visible surface — no normalization table to maintain and no case where one backup copy stands in for another the restore would not read)
Alternate: normalized project-relative path
Operator verdict: DEFAULT-APPLIED 2026-10-15T22:51:25Z manager-spec+plan

### Q6: When `.agents` or `.agents/skills` is a symbolic link, does the skill-residue cleanup refuse the whole pass or only the link and archive steps?

Label: FOUNDER
Class: product-level
Authority anchor: none
Why unresolved (at authoring): Today a symbolic `.moai`, `.claude` or `.claude/skills` refuses the whole pass (`internal/cli/update/skillresidue/residue.go:54-56`, `:299-310`), but a symbolic `.agents` or `.agents/skills` does not: probe E4 shows the pass runs and reports the foreign link as the project's retired link (research.md E4). Adding a refusal would change a shipped command's behavior, so the row is product-level.
Default: keep the current non-refusing, report-only behavior (rule: preserves current behavior — the measured behavior in E4)
Alternate: refuse the whole pass when `.agents` or `.agents/skills` is a symbolic link
Operator verdict: CLOSED 2026-10-16 by the Q3 ruling — the principle "add no new data-deleting behavior" is met by the default: with no removal path (Q3), a symbolic `.agents/skills` can lead nowhere, and a refusal is not needed to keep B's content untouched. A refusal removes nothing, so it would also be consistent with the principle, but it is a behavior change the ruling does not ask for; the SPEC does not add it. The default is therefore recorded as consistent with the principle, not as `DEFAULT-APPLIED`. (Corrects the earlier row, which labelled the refusal "preserves current behavior" and `implementation-level` — audit D6.) The operator may add the refusal later as a separate card.
