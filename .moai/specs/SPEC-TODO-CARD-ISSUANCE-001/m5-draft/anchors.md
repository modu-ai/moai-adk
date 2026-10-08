# M5 Mode B Draft — Insertion Anchors

> The gate read CLOSED (t1453 unlanded at the M5 reads), so the six rule
> files and their template copies stay UNTOUCHED on this branch. This draft
> holds what a follow-up card applies when the gate opens — or whenever the
> operator orders the rules amendment regardless of the gate. Each entry
> names the file, the anchor text already in it, and what goes where. The
> reachability sentence (entry 1) is a NET-ZERO edit by compression; the
> merge-doctrine exception (entry 2) is APPEND-ONLY after a pinned sentence.

## 1. Reachability — the always-loaded pointer

- **File**: `.claude/rules/moai/workflow/factory-dispatch.md` (+ byte-identical template copy).
- **Anchor** (the deputy-powers sentence, currently reading "…operator gates,
  card issuance and `done` (`moai gtd` mutations), …"): compress the
  enumeration to "…operator gates, card issuance (the card-issuance rule),
  and `done` (`moai gtd` mutations), …" — the parenthetical names the new
  rule, net character delta 0 (the compression funds it).
- **Why**: a path-scoped rule without this pointer never reaches a session
  that opens neither anchored path. The compressed sentence carries the
  reachability to every session.

## 2. Merge doctrine — the operator-verb exception

- **File**: `.claude/skills/moai/workflows/gtd.md` (+ template copy).
- **Anchor**: the `[HARD]` sentence beginning "Analysis never folds one card
  into another, never reorders the queue," — PINNED VERBATIM by a doc test;
  it is never edited or removed.
- **Insert AFTER that sentence's paragraph** one new sentence: an operator-
  invoked `todo merge` is the named exception — the fold is the operator's
  decision, made through the verb that records the relation and the drop;
  no analysis, relate, or lane path gains the fold.
- **Second anchor** (the relation-table row documenting `absorbs`): append
  "the operator merge verb supersedes an absorbs suggestion" to the row's
  trailing note — a record-only suggestion points at the verb that acts.

## 3. Sync review — the follow-up split and the ledger

- **File**: `.claude/agents/moai/sync-auditor.md` (+ template copy).
- **Anchor**: the blocking/optional classification section (the passage that
  defines PASS-WITH-DEBT).
- **Insert after it** two sentences: (i) a defect the card's own change
  introduced blocks in-card repair — it is never a follow-up card; a defect
  the change merely exposed goes to the per-component debt ledger with a
  one-line reproduction; (ii) the debt ledger is the queue's debt-origin
  rows, grouped per component — the reviewer's PASS-WITH-DEBT names the
  ledger rows it relied on.

## 4. Queue manager — follow-up and depth

- **File**: `.claude/agents/moai/manager-todo.md` (+ template copy).
- **Anchor**: the dispatch-guidance section (the passage advising when a
  card is issued vs split).
- **Insert after it** two sentences: (i) below the size lower bound, issue
  as a bundle member or folded work — never as a standalone row; above the
  upper bound, split; (ii) a spawn chain deeper than the derivation bound
  issues its descendant as a top-level card — the thresholds rule the
  card-issuance pointer names holds both values.

## 5. Mirrors

- Byte-identical pairs (`gtd.md`, `factory-dispatch-detail.md`,
  `factory-dispatch-mechanics.md`, `manager-todo.md`): the same edit lands
  on both copies, byte-identically. Entries 2 and 4 touch `gtd.md` and
  `manager-todo.md` — their template mirrors get the identical bytes.
- Forked pairs (`factory-dispatch.md`, `sync-auditor.md`): the SAME new
  text lands on both copies of each pair; the existing fork hunks are
  preserved untouched, and the fork hunk count must equal the pre-edit
  count.
- Emissions: after the edits, `make agents-emit` (the two agent files),
  `make plugin-emit` (`gtd.md`), and the catalog-hash regeneration all run
  before the same commit stages.
- Budget: `gtd.md` measures over the 40,000-character budget today. The
  application card MUST bring it under budget in the same edit that lands
  entry 2 — by compression that does not move content out of the
  always-loaded surface (reachability first), and the card-id baseline pair
  the compression touches is repaired in the same commit.

## 6. What this draft deliberately does NOT hold

The thresholds file holds the numbers; this file holds no number, no card
id, and no repository-internal path in the insertion texts themselves — the
draft is the boundary between "what the rule will say" and "what this
repository measured", and mixing them here is how a measured value leaks
into a distributed copy.
