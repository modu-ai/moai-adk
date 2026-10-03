# decision-index.md — SPEC-INIT-SHRINK-001

`interview.decision_gate: on` and `interview.recommendation_mode: pull`. Each row below is a
decision surfaced while assembling card t1438 that no operator answer in the interview settled.
Rows state what is unresolved and why; none carries a preferred answer. Options for every row are
enumerated in `spec.md` §5 (Q-n is OD-n there), and `spec.md` §5's marker table lists, for each
row, the requirement clauses that carry its default and what changes if the verdict differs.
`Operator verdict` is empty at authoring and is read by the Kickoff gate. Label vocabulary:
DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER.

**Authority note (per the leader's dispatch of 2026-10-03):** operator-gate-level decisions for
this card are set by the leader on audit evidence. No verdict line on any row is invented during
authoring; a leader ruling is recorded verbatim on the row it settles.

No row is DECIDED or POLICY-COVERED: no committed artifact in the authority register
(`.moai/project/product.md`, prior completed SPECs' HISTORY/Amendments rows,
`.moai/config/sections/*.yaml`, the constitution) answers any of these questions as written. The
t1435 OD verdicts (operator batch 2026-10-03, via leader relay) are binding pins this card designs
within, but they decide t1435's questions — not the ones below; where a pin constrains a row
(OD-3 → agents stay; OD-8 → core tier only), that is stated in the row's context, and the row
still asks the question the pin does not answer.

## Q1: On the default (plugin) path, what does init write into the project `.mcp.json` for the `moai` entry?

- Label: FOUNDER
- Authority anchor: none — the card text names `.mcp.json` in the thin-deploy list without saying
  whether its `moai` entry stays, and no committed artifact decides it
- Why unresolved: the plugin carries the entry (t1435 OD-2 pin), a project copy is inert-but-duplicated
  (t1434 R06: the project copy wins and suppresses the plugin copy), and today's project-entry
  behavior is harness-keyed (`init.go:997-1004`, P-15 — `gpt` declines, `both` forces,
  non-interactive writes none). The Codex side keeps its `config.toml` wiring on every path, so a
  `gpt` + plugin project holds two registrations whose combined effect is UNMEASURED. Default path
  and `--no-plugin` path may legitimately differ; which split the operator wants is a product call
  the evidence does not make.
- Operator verdict:

## Q2: If the REQ-008 measurement shows bare names do NOT resolve to namespaced plugin components, what ships?

- Label: FOUNDER
- Authority anchor: none — the fallback policy for a measurement outcome that does not exist yet is
  named by no committed artifact
- Why unresolved: the measurement itself is a fact this card produces (REQ-008, the t1434 scratch-home
  routes); the unresolved part is the fallback. Bare-name references are load-bearing today (5
  instruction-file references, 7 rule files, P-17; every command body's `Skill("moai")`), and
  t1435 §1.5-3 named the resolution question "load-bearing at t1438" without deciding it. Rewriting
  references, keeping a partial scaffold, or holding the flip are all consistent with the card text;
  the cost and risk differ materially.
- Operator verdict:

## Q3: What does the migration remove and what does it back up?

- Label: FOUNDER
- Authority anchor: none — SPEC-UPDATE-DATA-SURVIVAL-001 (completed, in-tree) fixes that a failed
  backup aborts before any removal, but not which classes this migration removes or archives
- Why unresolved: today's pre-clean backup exempts template-carried files (P-08) — exactly the
  copies this card removes — and the raw template keeps carrying the sources as the plugin
  derivation input, so the exemption would silently apply. Removing identical copies without
  archive, archiving everything, or a two-step report-then-remove release are all defensible; the
  operator's tolerance for archive bulk vs recovery convenience is not decidable from evidence.
- Operator verdict:

## Q4: What is the migration surface for a project with no deploy-mode record (every pre-shrink project)?

- Label: FOUNDER
- Authority anchor: none — no committed artifact decides whether `moai update` may run the t1435
  install step (a potential network touch) as part of a migration
- Why unresolved: old projects never ran the install step (it attaches to init and the three
  install scripts, t1435 REQ-010/018), so a naive dedupe would leave them with no skills at all.
  Making update the migration surface changes update's character (today it deploys from embedded
  templates only); keeping update offline makes the plugin switch an explicit init re-run; a
  dedicated `moai migrate` verb adds surface. The card text asks for an 이행 경로 with 중복 제거·백업
  but does not pick the surface.
- Operator verdict:

## Q5: Where does the deploy-mode record live, so update never infers it?

- Label: FOUNDER
- Authority anchor: none — no committed setting or prior SPEC names a deploy-mode key
- Why unresolved: the `llm.harness` persistence pattern (`ApplyHarness`, `init.go:947`; update
  re-assert, `update_template_sync.go:666-668`) is the natural home and precedent, but a dedicated
  section file and a manifest-derived inference are both consistent with the card. The choice
  binds every consumer criterion (REQ-009/016 and their ACs), so it is pinned before M1 rather
  than improvised.
- Operator verdict:

## Q6: What happens to the Codex command-skill mirror (`.agents/skills`) after the shrink?

- Label: FOUNDER
- Authority anchor: none — no committed artifact decides the mirror's fate on a plugin-carrying
  deployment
- Why unresolved: Codex lists plugin skills (t1434 R01-codex PLUGIN-OK) and receives plugin
  commands as generated skills (R03-codex), but the activation observations are render-level only
  (t1434 G-f), the mirror machinery is deploy-time symlinks whose lifecycle "belongs to the clean
  path" (P-11), and a codex-only project re-homes skills as real directories (P-12). Follow-the-mode,
  always-deploy, and retire-entirely each leave a different residue for `local`-mode and codex-only
  users.
- Operator verdict:

## Q7: What does `--all` mean after the shrink?

- Label: FOUNDER
- Authority anchor: none — the flag's current meaning ("deploy all catalog entries", the slim-mode
  bypass, P-05) predates the plugin and no artifact redefines it
- Why unresolved: the plugin carries the core tier only (t1435 OD-8 pin), so 13 optional-pack
  entries have no plugin home. Whether `--all` becomes the local full-deploy escape hatch, widens
  only the tier while the plugin stays the carrier, or is deprecated is a product call with
  different docs and migration consequences.
- Operator verdict:

## Q8: Who owns instruction-file consistency beyond name resolution?

- Label: FOUNDER
- Authority anchor: none — card t1466 (leader-issued 2026-10-03, per the t1418 verdict H1) is
  queue text, not a committed artifact, and cannot anchor a scope boundary
- Why unresolved: this card's REQ-008 fixes that names referenced by shipped instructions resolve
  in each deployment mode; the broader deployed-template instruction drift is t1466's subject, and
  this card's shrink changes the deployed skill set again (which is why the dispatch names the
  interface). Absorbing t1466's moved-component scope here, deferring wholly to it, or doing only
  what the flip mechanically forces are all open; the boundary between two cards is the operator's
  to draw.
- Operator verdict:
