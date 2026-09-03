# Audit Artifact Convention — where audit verdicts live and who makes them stick

> Where audit verdicts (plan-audit, sync-audit) are exported, when they must
> be written, and what a lead reads when deciding whether a phase closed.
> This convention removes the export-on-request dependency: exporting is part
> of completing an audit, not a courtesy.
>
> Audience: any MoAI-ADK project (template-distributed)

---

## Why

An audit verdict exists only in the session that produced it unless it is
exported to a file. When that session ends — by `/clear`, by a handoff, or by
a machine boundary — the verdict's basis is unrecoverable: the score survives
as a claim, but nothing a later reader can re-check. A verdict that cannot be
re-read is an unattributed claim under the evidence-integrity doctrine.

Experience shows the determining factor is not auditor capability but whether
an export was requested. Asked, an auditor produces a complete report file
immediately in the same turn; unasked, the verdict dies in the transcript.
Left uncodified, export stays a habit — and habits do not survive instruction
turnover across sessions and lanes.

This convention therefore removes the per-request dependency: the export is
part of what it means for an audit to be complete. An audit response without
an exported file is an incomplete audit. The failure is not a discipline
failure but a missing default — actors who know the rule and value evidence
still lose the export when nothing in the workflow asks for it. A convention
exists so that remembering is not the mechanism.

---

## Where

- Card-scoped audit artifacts go to `.moai/reports/<card-id>/`. File names:
  - `plan-audit.md` — single-iteration plan audit
  - `plan-audit-iter<N>.md` — one file per audit iteration
  - `sync-audit.md` — sync-phase audit (or `sync-audit-verdict*.md` where an
    existing workflow already names it so)

  The `plan-audit-iter<N>.md` / `sync-audit.md` family is the dominant
  existing convention; new artifacts stay in that family.

- SPEC-scoped audits produced without a card may use
  `.moai/reports/<SPEC-ID>/` instead.

- FORBIDDEN: `.moai/reports/plan-audit/`. That directory is deliberately
  gitignored — only its `.gitkeep` is tracked, and the repository's
  `.gitignore` comment marks the directory as local artifacts. A verdict
  written there is disposed of, not exported. Do not repurpose the directory.

---

## When

The audit agent writes the file **in the same turn it renders the verdict**.
Deferring the export to "later cleanup" is prohibited — timing deferral is
the exact failure mode this convention closes, because the session holding
the verdict is the one that may not survive to write it.

The audit agent exports itself. A lead or lane may relay the export only when
the auditor cannot write files, and must record in the artifact that it did
so.

---

## What

An exported audit artifact carries at minimum:

- The verdict token and score (for example `PASS`, `PASS-WITH-DEBT`, `FAIL`,
  plus the numeric score where the audit type uses one)
- Per-defect findings — each defect's identifier, severity, and a one-line
  summary
- The commands run with their observed outputs, in evidence-bearing format:
  Claim / Evidence / Baseline-attribution / Gaps / Residual-risk
- Iteration history when the audit ran more than once — what changed between
  iterations and why

An inline response summary alone does not satisfy this convention. The
summary points to the file; it does not substitute for it.

---

## Side-talk (advice attached to verdicts)

An audit report sometimes carries advice beyond its verdict — operational
warnings, "measure this next time" notes, policy suggestions. The verdict
itself must carry evidence (§ What); the advice that follows it has no such
requirement, yet a reader consumes both at the same confidence. Two observed
failures shape this rule: advice that turned out to be exactly the valuable
warning it needed to be, and advice whose cited rule was true but whose
inference drawn from it was false.

The convention:

- Advice lives in a **separate section titled as unverified** (for example
  "Operational Notes (unverified)") at the end of the artifact — never woven
  into the verdict, the dimension scores, or the defect list it follows.
- Advice is written as a **measurement instruction, not a conclusion**.
  "Measure X — command Y" survives on its own, because the reader can run Y;
  "X is false" may only be stated once X has been verified. A cited rule
  being true is not the same as the advice drawn from it being true — the
  inference is what needs measuring, not the rule.
- Each advice line carries a **status label**: `measured` (the command and
  its output are recorded, per § What), `inferred` (the reasoning rule is
  named so the reader can check it), or `assumption` (a naked claim — the
  weakest standing). A `measured` label without its recorded command is a
  mislabel, not evidence.
- Banning advice outright is rejected by design: valuable warnings arrive as
  advice first, and forbidding them costs the next reader the very warning
  that would have prevented a wasted re-measurement.

This section is itself subject to the rule above: the incidents that shaped
it are recorded in the artifacts of the cards that raised them, not asserted
here.

---

## Committing

Audit artifacts are tracked files. They reach the integration branch with the
card's evidence commit — never left as uncommitted files in a worktree. A
worktree holding the only copy of a verdict is a disposal hazard: the tree is
removed when the card closes, and the verdict goes with it.

---

## What makes the convention stick

Three enforcement layers, none of which depends on a person remembering:

- **Auditor side.** The plan-auditor and sync-auditor agent definitions carry
  the export as a HARD completion condition: an audit response without an
  exported file is an incomplete audit.
- **Lead side.** The lead advances a card by reading its evidence, never by
  trusting a companion's reply. When a phase's evidence should include an
  audit verdict and the file is absent, that is a gap — the card stays put
  and the lead reports why.
- **Mechanical check.** `ls .moai/reports/<card-id>/` — a missing verdict
  file is the detection.

---

## Cross-references

- `.claude/rules/moai/workflow/kanban-dispatch.md` § Completion is read,
  never trusted — the lead-side reading obligation
- `.claude/rules/moai/core/verification-claim-integrity.md` — the
  unattributed-claim invariant this convention operationalizes
- `.claude/agents/moai/plan-auditor.md`,
  `.claude/agents/moai/sync-auditor.md` — the export-mandate clauses
