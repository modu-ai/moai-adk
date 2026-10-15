# Decision Index — SPEC-UPDATE-BACKUP-RETENTION-001

Stateless artifact: it carries no `status:` field; the SPEC's lifecycle lives in `spec.md`.

Routing labels: `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, `FOUNDER`. Every row states what is unresolved and why. No row carries a recommendation for a judgment call; a `Default:` line is the published Default rule applied (preserve current behavior; else single-revert undo; else smaller user-visible surface), not a recommendation.

Authority note. The operator rulings this card rests on (decision board records a263, fd8b, a041, 242e; read with `moai decision read --scope card:t1687 --all`) and the t1687 audit reports under `.moai/reports/t1687/` are not in the committed tree. By the authority register they are evidence, quoted in `research.md` § Operator rulings, and never an authority anchor. A row whose only support is one of them routes `FOUNDER`.

### Q1: How does the state-residue pass avoid removing a zero-byte database a process holds open (finding F21)?

Label: FOUNDER
Class: product-level
Authority anchor: none
Why unresolved: The current pass removes a zero-byte `*.db` under `.moai/state` that has no `-wal` / `-shm` sidecar and has sat unmodified for ten minutes (operator-approved rule, board 73f8; `internal/cli/update_state_residue.go:19`, `:32-62`). A process that opened such a file without writing to it takes `disk I/O error (1802)` on its next write (t1687 audit iter3, A3-P3; cited, not re-measured here). Options: (a) keep the rule and accept the residual race; (b) report zero-byte `.db` files and never delete them (the zero-byte `backlog.json.lock` removal stays); (c) narrow removal to an enumerated list of legacy paths that no current binary opens. A partial grep found only migration-read openings of `.moai/state/todo/backlog.db` in the current binary (`internal/cli/migrate_home_state.go:55`, `:382`, `:576`); the list for (c) does not exist. Any change from (a) alters a shipped command's cleanup behavior.
Default: (a) (rule: preserves current behavior)
Alternate: (b)
Operator verdict:

### Q2: May protected excess runs exceed keepCount, or must the total run count stay at or under keepCount?

Label: FOUNDER
Class: product-level
Authority anchor: none
Why unresolved: Operator ruling a263 (C) says to keep the newest N past versions per file and clean up the rest; the card's completion text says both backup roots end at or under their cap. A protected version may live only in an old run, so the two cannot both hold for every input. Options: (A) soft cap: the window stays at keepCount and the irredundant set of excess runs that hold protected versions is retained on top, reported by name (REQ-UBR-010 as drafted); (B) hard cap: total runs at most keepCount, with protected versions carried forward into a retained location, which writes new content into backups; (C) hard cap with a limit on retained excess runs, dropping the oldest protected versions beyond it (data loss beyond the limit). The published Default rule does not rank these three. The number of excess runs option (A) retains on the mo.ai.kr roots is unmeasured (Q4).
Operator verdict:

### Q3: Is the retired-skill archive and `.agents/skills` link removal reintroduced?

Label: FOUNDER
Class: product-level
Authority anchor: none
Why unresolved: Main's behavior, and the behavior t1687 reverted to, reports a retired skill folder and its link and moves or removes neither (`internal/cli/update/skillresidue/residue.go:75-94`, `:167-192`). The t1687 progress record hands "reintroduction and repair" of the archive and link-removal path to this card, and board a041 moves F16 and F17 here; this SPEC drafts REQ-UBR-015 to REQ-UBR-017 on that basis. Reintroduction changes a shipped command's default user-visible behavior (a folder moves, a link disappears). Options: (A) reintroduce with the root and parent-chain checks of REQ-UBR-015 to REQ-UBR-017; (B) keep report-only and add the `.agents` root check as a guard on the reporting path, leaving P4b vacuous. Neither authority is in the committed tree.
Default: (B) (rule: preserves current behavior)
Alternate: (A)
Operator verdict:

### Q4: How many excess runs does the irredundant retention leave on the real mo.ai.kr backup roots?

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: Whether the soft cap of Q2 option (A) is acceptable depends on a count that does not exist yet. The t1687 audit's model estimate (30 of 42 namespace runs retained, per-version newest-holder model, iter2 residual-risk) predates the irredundancy requirement and was not measured against it. AC-UBR-06 produces the figure on a modification-time-preserving clone at run-phase close.
Operator verdict:

### Q5: Is a backed-up file version keyed by its path inside the backup directory, or by a normalized project-relative path?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: A copy under `pre-clean/.claude/x` and a copy under `.claude/x` in a namespace backup are the same project file at two paths inside backups. Keying by the path inside the backup treats them as different versions; normalizing the `pre-clean/` prefix away would let one satisfy the other. The restore path resolves a file by its path inside the backup, which is what the pre-revert key (`git show 8e43f064a^:internal/cli/update/backup/backup.go`, `backupFileKey`) used.
Default: key by the path inside the backup directory (rule: preserves current behavior)
Alternate: normalized project-relative path
Operator verdict: DEFAULT-APPLIED 2026-10-15T22:51:25Z manager-spec+plan

### Q6: When `.agents` or `.agents/skills` is a symbolic link, does the skill-residue cleanup refuse the whole pass or only the link and archive steps?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: Today a symbolic `.moai`, `.claude` or `.claude/skills` refuses the whole pass and returns one `refused-symlinked-root` action (`internal/cli/update/skillresidue/residue.go:54-56`, `:299-310`). Extending that list is the smallest change; refusing only the link and archive steps would keep empty-folder cleanup and renames working for a user who links `.agents/skills` on purpose, at the cost of a second refusal path.
Default: refuse the whole pass (rule: preserves current behavior)
Alternate: refuse only the link and archive steps
Operator verdict: DEFAULT-APPLIED 2026-10-15T22:51:25Z manager-spec+plan
