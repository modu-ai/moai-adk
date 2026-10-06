---
id: SPEC-USER-ASSET-INSTALL-001
title: "progress.md — phase progress record"
created: 2026-10-05
updated: 2026-10-05
author: manager-spec
---

# Progress — SPEC-USER-ASSET-INSTALL-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-05
- Plan-phase artifacts complete (Tier L set: spec.md, plan.md, acceptance.md,
  design.md, research.md, progress.md, decision-index.md) @ worktree
  `WT-user-asset-copy`, authored against HEAD `6643c7bba` (research
  baseline), repaired in the iter1-defect-closure commit.
- Plan audit: iter1 FAIL 0.64 (Tier L threshold 0.85; MP-8 firewall — no
  RED-now cells). Iter1 defects D1-D13 closed in the v0.2.0 repair; iter2
  delta re-audit FAIL 0.75 (improving, no STOP signal): 12/13 iter1 defects
  verified RESOLVED; the D12 residue (renamed D19) plus new findings
  D14-D23 closed in the v0.3.0 repair. Iter3 is the last numbered round.
- v0.3.0 iter2 repair closure map: D19/D12-residue — plan M4 + design §2.6 +
  research V13 repoint list corrected (`probeCodexReadiness`/
  `countCodexAgentTOMLs`, codex_readiness.go:131/:215-217; the
  `codexStaleSkillFinding` attribution withdrawn, doctor_codex.go:857-870
  reads user-layer `[[skills.config]]`, no agent-count input); D14 — REQ-024
  upgrade arm + REQ-020 same-run gating + design §2.4 no-manifest branch
  relabeled + M3/M4 wiring + AC-020 extended (Blocker) with REQ-024
  secondary; D15 — AC-011 rebuilt on catalog-derived placements (38 skill
  dirs = 37 `moai-*` + plain `moai`; `/bin/ls` measured); D16 — REQ-023
  truth table + `~/.moai/` backup home with the C2 carve-out + REQ-011
  manifest-repair count + AC-006/008 arms; D17 — C2 resolved-root/leaf/
  TOCTOU edges + design §2.1 write posture + AC-025 arms; D18 — REQ-004
  selection surface (`--bundles`, `moai bundle add|remove`, manifest
  `bundles:` list) + design §2.3 + AC-018 + M2/M3 assignment; D20 — §D.2
  enumeration, AC-002 REQ-024, research §3 D-Q3 closure text, EV-014 green
  path, proxy-cell notes; D21 — versions 0.3.0, baseline-SHA policy in plan
  §C.2, M0 heading + D-Q2, decision-index iter1-D4 prefix, research V16
  38-dir correction; D22 — AC-009 repoint binding, AC-013 release-chain
  grep, AC-016 advisory row; D23 — REQ-021 unknown-field preservation +
  design §2.2 note + AC-021 arm.
- REQ/AC accounting (unchanged counts, stated per ceiling): REQ stays 24
  (all new assertions folded into REQ-004/009/011/020/021/023/024). AC stays
  25 — the auditor's "add a Blocker AC for the upgrade case" (D14) is
  satisfied as an EXTENSION of Blocker AC-020 (Verifies + REQ-024 secondary,
  GWT extended), not a 26th AC; no AC was swapped out (none was orphanable —
  every AC is its REQ's sole or primary coverage).
- Evidence ledger: all 22 cells + 2 positive controls RE-EXECUTED verbatim
  on tree cfb9033582eff27f9031e1a6438d8558aaa48115 (source bytes identical
  across b965a3912 → cfb903358 → this repair: only SPEC artifacts touched);
  document pin re-bound accordingly; proxy-cell notes added (D20e).
- Deferred (none blocking): D22's "add Major ACs for the repointed
  diagnostics and the advisory row" was folded as EXTENSIONS of AC-009 and
  AC-016 respectively (25-AC ceiling; stated here per the fold-and-state
  rule). No other optional deferred — D21/D22/D23 all taken.
- Source verification: 19 rows (V1-V13, V17, V18, V19 CONFIRMED; V14
  UNRESOLVED-routed; V15 AMBIGUOUS-routed; V16 partially-confirmed with 2
  corrections) — research.md.
- Decision gates: D-Q1/D-Q2/D-Q4/D-Q5 RESOLVED 2026-10-05 (operator/leader
  adjudication relayed with the iter4 authorization; verdicts in
  decision-index.md — D-Q2/D-Q4/D-Q5 leader defaults, operator-contestable;
  D-Q1 a full operator reading). D-Q3/D-Q6 closed at plan phase by
  constraint (POLICY-COVERED; premises P5/P6). No gate blocks M0/M1 at run
  entry.
- REQ/AC: 24 / 25 (ceilings 25/25 respected).
- RED-now baseline: all 22 release-blocking ACs carry executed RED cells
  (acceptance.md §D.2b; first measured on tree
  `b965a3912c0e97ef81aeeea773019e633591e1cd`, re-executed in full and
  re-pinned on `cfb9033582eff27f9031e1a6438d8558aaa48115` — the document
  pin; iter4 D31a correcting this bullet's stale first-measurement pin).
- Plan-audit trajectory (card t1509): iter1 FAIL 0.64
  (`.moai/reports/t1509/plan-audit-iter1.md`, 13 findings) → repair
  cfb90335 (v0.2.0) → iter2 FAIL 0.75 (`plan-audit-iter2.md`, 12/13
  RESOLVED verified; D12-residue + D14-D23) → repair 082b7daa5 (v0.3.0) →
  iter3 **FAIL 0.79 with CEILING HIT** (`plan-audit-iter3.md`, receipt
  rcpt-da1471a704401a2a2e1ed9cf) — iter2's ten findings all RESOLVED at
  their fix-route demands (D14's AC-020 fold judged real coverage); new
  blocking D24-D29 (upgrade-gate machine states, manifest-stale-at-removal
  cross-artifact conflict, TOCTOU posture claim false, foreign-schema
  preserve unreachable, conflicting bundle-removal criteria, command-wrapper
  disposition) + optional D31-D34 — per the auditor, all six are
  wording+arm fixes inside the existing ceilings. Per the leader's
  dispatch discipline ("plan 감사 상한에 닿으면 보고"), the lane STOPS at
  3/3 numbered rounds and REPORTS: escalation options per the verdict §
  Recommendation are PASS-with-debt (debt inventory in the verdict) /
  scope-reduction / one authorized delta round (iter4 fix surface named in
  the verdict). Cross-model: claude FAIL (11) + codex FAIL (4), glm
  inconclusive; two claude claims rejected with evidence. Run-phase entry
  AWAITS the leader's disposition.
- v0.4.0 iter4 delta round (2026-10-05; the ONE authorized round per the
  operator disposition on the iter3 ceiling hold; fix surface exactly the
  iter3 verdict's enumeration): D24 removal gate re-keyed per-asset
  (REQ-020/024 rewrite, design §2.4 three machine states, AC-020 GWT +
  matrix arms); D25/D28 one removal rule — REQ-009 extended to manifest-hash
  OR shipped-bytes, design §2.4 re-keyed to the selection-based criterion,
  design §2.1 manifest-stale removal arm aligned, AC-006 + AC-018 arms; D26
  TOCTOU "closed" → "narrowed" + declared parent-swap limitation + AC-025
  posture arm; D27 REQ-021 unknown-field preservation extended to ALL
  manifest writes + AC-021 foreign-schema round-trip arm; D29 published
  command skills dispositioned (D-Q4/D-Q5 propagation: design §2.3/§2.5,
  spec §6, AC-011 note); D30 EV-009 → M4+M5, EV-018 → M0+M2+M3, per-cell
  flip expectations added (EV-005/006/008/009/010/016/017/018); D31 stale
  §E.1 pin re-bound + M0 heading de-overstated; D32 repoint-clean/advisory
  rows folded into REQ-014/REQ-019 + §D.2 note; D33 backup home pinned
  (root-slug layout, resolved-path judgment, sanitization, AC-025 arm);
  D34 C2 sole-write clause scoped to asset writes + manifest named as the
  SPEC's own state file; carried nit taken — EV-011 command pinned to
  `/bin/ls` and re-executed on this tree (plain listing; the unpinned form's
  long format was the environment alias, observed three times). Gates
  D-Q1/D-Q2/D-Q4/D-Q5 adjudicated and recorded. Baseline re-pinned
  post-absorption: 6643c7bba → 51976e651 (develop a158b4b5f absorbed; the
  load-bearing pins re-verified holding; the one anchor drift —
  `inspectSkillMirror` comment `:425` — touches no live citation, the func
  line `:429` unchanged, re-measured this run).
- v0.5.0 round-5 gate-fix round (2026-10-05; the codex review gate's four
  findings, operator-authorized disposition (i), artifacts frozen at HEAD
  064ff9960): F1 — L0's transitive runtime skill closure enumerated in the
  catalog L0 view (REQ-003/REQ-004, design §2.3 eight-skill two-tier table
  + M0 drift guard over the sources; verified at source, not transcribed —
  plan-auditor contributes NOTHING to the static preload union, its body
  says so; research §2b W1-W3); "default-install-runs" absorbed into
  AC-017. F2 — dispatcher references rebind at SOURCE level (REQ-001
  user-side mirror `$HOME/.agents/skills/moai/`; design §2.5: sources
  `.claude/commands/moai/` → emitter `internal/template/commandemit`
  (CommandsRoot, commandemit.go:51) → `make commands-emit`, drift guard
  `commands-emit-check` in the build chain, Makefile:34/:51-60;
  AGENTS.md.tmpl:40-41 rebind — the gate's "§3" label corrected, the
  sentences sit in the unnumbered preamble); loading verification absorbed
  into AC-002 (green path M2+M4). F3 — init re-keyed per-asset-state
  (REQ-024, design §2.1); partial-failure-retry absorbed into AC-001. F4 —
  manifest read-modify-write serialized per user (REQ-006, design §2.2);
  cross-run coverage absorbed into AC-018. Absorption map written into
  acceptance §D.2; no new REQ, 25/25 AC ceiling holds.
- Round-5 ADDENDUM (fold batches A1-A4 + B1, operator-approved while the
  core round was mid-flight; VERSION STAYS 0.5.0 — one version per round):
  A1 — init's per-asset judgment reworded to the REQ-023 truth table (the
  first cut's "bytes differ → reinstall" clobbered user edits; design
  §2.1 + REQ-024 + AC-001 edited-file preservation arm); A2 — the
  dispatcher's EIGHTEEN internal `Read .claude/skills/moai/workflows/*.md`
  references (the fold named the L0 three; the same-class sweep found 18)
  rebind at template source to installed-skill-relative paths (the
  dispatcher is a source template, NOT a commandemit output — precision
  recorded); AC-017's executable arm extended to LOADS, not merely
  resolves; A3 — `checkSkillsAllowlist` (doctor.go:957-958) joins M4's
  repoint list, repointed NOT removed; AC-009 enumeration extended; A4 —
  AC-005's verdict basis converted to the M3 behavior test, the EV-005
  grep demoted to auxiliary with the conversion record written into the
  cell (the auditor makes the final call at the gate re-run); B1 — the L0
  closure restated as the TEN-skill three-tier union (static preload ∪
  dispatcher routing ∪ on-demand invoke sites: + moai-workflow-testing,
  moai-workflow-worktree) with moai-ref-*/moai-domain-* per-mission
  injections explicitly classified out (4 sites swept: ref-cross-model-
  audit ×2 agents, ref-owasp-checklist, ref-testing-pyramid,
  domain-html-report); REQ-003 + design §2.3 + plan M0 drift guard pin the
  union and the classification-out. Sweep evidence: research §2b W6-W8.
- v0.6.0 FINAL CLASS ROUND (2026-10-06; operator disposition α — the LAST
  plan→run branch; STANDING RULE recorded: any NEW gate finding after
  this round is run-phase debt, no further plan folds). The CLASS CLAUSE
  pinned verbatim (design §2.5): "every project-relative reference in the
  user-scope deployed tree rebinds to its installed location, verified by
  a raw-pattern sweep + run-phase loading ACs" — replacing
  layer-by-layer enumerations for the whole rebind family. Items 1-8
  closed: recursive workflows-tree rebind (broad sweep measured 274 raw
  occurrences over 26 files + 5 subdirs; AC-017 step-document loading
  arm); all-17 command-skill rebind (source split 13 sources + 4
  emitter-injected — goal/gtd/sync/todo — W11); six remaining grep cells
  converted to behavior-test verdict bases (EV-006/008/009/010/017/018 +
  §D.2b conversion note, extending the adjudicated-SOUND EV-005 ruling);
  manager-git policy DECIDED require-a-bundle (Route B precondition +
  `moai bundle add` remediation, AC-018 arm); pending-install recovery
  journal designed (expected-hash match claims the run's own installs
  without absorbing user files — REQ-006, AC-001 arm); mirror-repair
  rollback terminated (update.go:535 / skill_mirror_repair.go:89,:113
  verified re-growing the 17 published copies — AC-011 repeated-update
  arm); `moai-ref-cross-model-audit` joined L0 under the DEFAULT-FLOW
  REACHABILITY criterion (closure ELEVEN; corrects fold B1's prefix
  classification); stale "eight-skill" labels re-pointed here-exterior
  (plan M0 + research §2b; this file's earlier entries stay as history).
- IN-ROUND EXTENSION (same final class round, v0.6.0 — two more item-7
  criterion instances from the gate's mid-flight review): E1 —
  manager-lead pinned as the FACTORY entry's declared agent dependency
  (factory-dispatch.md:104 [HARD] resident deputy, template mirror
  identical; NOT a sixth core agent — D-Q1's five stands); closure table
  + drift guard + REQ-003 + AC-017 carry it. E2 —
  `moai-ref-owasp-checklist` + `moai-ref-testing-pyramid` reclassified IN
  (the DEFAULT Phase 7 evaluation scores Security + test coverage by
  default — the fold-B1 per-mission reading was wrong for these two);
  DESIGN CALL: both join L0, no degraded absent-path (one criterion,
  no second standard); closure THIRTEEN skills + the factory agent
  dependency; `moai-domain-html-report` alone stays out. Evidence:
  research §2b W12.
- v0.6.1 DIRECTED REPAIR (leader order, post-convergence; the leader
  audit's two P2s promoted into plan content — the last edit before the
  auditor spawn + fresh receipt ceremony): R-a — the class-clause sweep
  now classifies hits into TWO populations checked differently (design
  §2.5 + plan M4): MOVED-asset references (skills/agents) rewritten with
  zero-hit-after-rebind on that population ONLY; PROJECT-RETAINED
  references (rules/hooks staying project-side — run/phase-execution.md:
  208-210, run.md `.claude/rules` ×13 + trace-ledger.sh :28, sync.md:43
  quality-gate hook, plan.md:47 spec-workflow) verified
  present-and-correct, never rewritten, never swept to zero; sweep
  evidence records both populations + the scoping decision. R-b — the
  manager-git bundle precheck extended to ALL entry points: sync delivery
  Route B + the run flow's Route B (task-decomposition.md:302-303,
  Phase 19, W13); design §2.5 policy block + plan M3 + AC-018 arm
  verifies BOTH paths.
- v0.6.2 CLASS REPAIR (leader 2nd-audit order — class sweeps ending the
  instance parade; the last edit before the 3rd receipt ceremony): R-c —
  the install-coverage DERIVATION MATRIX promoted to plan content (design
  §2.3 + plan M0): M0 mechanically sweeps every loading instruction
  across the deployed workflow tree (run/sync paths incl.
  task-decomposition, phase-execution, quality-gates, delivery,
  spec-assembly) AND the L0 agent bodies, crosses each invoked role and
  loaded skill against the L0+bundle install set, and every gap becomes a
  remediation row; the MATRIX is the M0 drift guard's source of truth
  (the derived set, never a hand list — absorbs JD-8). Derived rows:
  api-patterns/react-patterns/domain-database → per-mission domain
  bundle rows (step 4b, W14); manager-git → the R-b precheck row;
  `moai-ref-secops` → IN via the default Phase 8 delegate path
  (quality-gates-quality.md:135 — same reachability as E2's owasp) —
  closure FOURTEEN skills. R-d — the rewrite-scope extension (design
  §2.5/§2.6 + plan M4): rebind target extends to skills tree + agent
  sources + GENERATED TOMLs (`make agents-emit`; measured instance
  e2e-tester.md:141 whose emitted TOMLs inherit the moved-asset
  reference), the RETAINED-FILE→MOVED-ASSET class (templates/CLAUDE.md:
  31/:47 rewritten), and the doctor consumer completion — `runHarnessCheck`
  L4 (doctor_harness.go:20/:70; the dispatch's "plan.md:243" label did
  not resolve — verified anchors cited) joins the M4 repoint list with
  the healthy-install regression arm (AC-009/AC-017 + absorption map).
- v0.6.3 DIRECTED REPAIR (leader 3rd-audit order — JD-1 promoted from run
  debt into plan content; closes the journal finding in the documents):
  the pending-install journal's COMPLETENESS specified (design §2.2 +
  REQ-006 + AC-001 + plan M1) — per entry AND per run the journal records
  (1) the bundle-SELECTION delta (an `init --bundles` interrupted before
  manifest-save recovers with the selection INTACT, never `[]` — the
  gate's temp model reproduced `recovered bundle selection: []` → REQ-009
  re-pruning the just-installed files; that sequence is the red), (2) the
  FULL manifest-entry provenance per file (bundle, moai_version,
  installed_at — path+sha256 alone makes two same-byte installs by
  different binaries indistinguishable, defeating REQ-006's per-file
  version), (3) the ownership evidence (write-completion flag per entry,
  reconciled atomically — an unflagged entry is NOT claimed; the
  ambiguous file stays a REQ-010 collision). AC-001 gains the
  interrupted-init --bundles arm (the reproduction as a GWT: selection
  restored intact, provenance restored, `moai update` does NOT re-prune
  the recovered files).
- IN-ROUND EXTENSION (same R-e round, v0.6.3 — two gate mid-flight
  findings, E1/E2 precedent): E3 — bundle removal preserves L0-shared
  assets: the removal target is the COMPLEMENT (entries of the removed
  bundle NOT in L0 ∪ the remaining opted-in selections; shared assets
  survive with a report note) — measured: the historical devops pack
  (catalog.yaml:225-251) carries the L0 trio owasp-checklist(:230)/
  cross-model-audit(:240)/secops(:250), and "remove exactly that bundle's
  entries" would delete them (gate reproduced the deletion); design
  §2.3 + REQ-004 + plan M3 + AC-018 shared-asset arm. E4 — the
  rename→written-flag window: the journal's STAGING record (written
  pre-rename: path+sha256+provenance) IS the intent-and-content proof;
  rename(2) is atomic, so a final-path file hash-matching a staged entry
  is claimed as OWN whether or not the flag write landed — never a
  permanent REQ-010 collision on any number of retries (gate reproduced
  repeated-retry non-recovery); the flag only separates
  completed-vs-uncompleted at reconciliation (no final-path file →
  reinstall from the journal's provenance); design §2.2 claim rule +
  REQ-006 + AC-001 rename→flag-window arm. Evidence: research §2b W15.
- IN-ROUND EXTENSION E5 (same R-e round, v0.6.3 — the gate sharpened the
  recovery lattice; user_bytes_preserved=False reproduced): the recovery
  branch's hash-MISMATCH case NEVER reinstalls — a mismatch is not
  evidence of an incomplete write (the user may have edited after the
  interrupted install; reinstalling on mismatch overwrites the user's
  edit). The recovery lattice is EXACTLY THREE CASES (design §2.2 +
  REQ-006 + AC-001 mismatch arm + plan M1): absent → install from the
  journal entry; present hash-matching → claim as own (E4 staging
  proof); present mismatching → NEVER reinstall, preserve as REQ-023
  divergence (flag-complete) or REQ-010 collision (unflagged) — both
  preserve the user's bytes.
- v0.6.4 DIRECTED REPAIR R-f (leader 4th-audit order, OPERATOR HARD
  SCOPE — the last edit before the 5th/final ceremony): R-f-① —
  cross-binary payload recovery DECIDED, option (b)
  reinstall-from-current-version + honest provenance re-stamp (design
  §2.2: the retry writes its own bytes for the absent target and records
  its own version per recovered file — REQ-006 stays truthful; option
  (a) journal payload-preservation rejected on the simplicity ladder —
  doubles `~/.moai/` storage + shadow-tree lifecycle for fidelity to a
  superseded binary's payload; gate observation `replay with current
  embedded bytes matches staged sha256: False` recorded); AC-001
  cross-binary case-1 arm. R-f-② — dependency maintenance for
  preserved assets: the removal step re-runs the derivation matrix's
  dependency rows and DEFERS the deletion of any entry that is a
  declared dependency of a preserved asset (kept + reported,
  re-evaluated next removal/update); preserved assets' consumers are
  user-folder references by the class clause, loading verified
  post-removal — design §2.3 dependency-maintenance block + plan M3 +
  AC-020 preserved-asset consumer loading arm.

## §E.2 Run-phase Evidence

### Run start (2026-10-06)

- Run-phase entry @ worktree `.moai/worktrees/t1509` (branch
  `WT-user-asset-copy`), HEAD `ac0c72ece` — the lane-measured tree at
  dispatch time; plan artifacts v0.6.4, hash
  `4368c9f199121733fc305f63651e1e2715213df3f0c73097dbec765812559dca`
  (dispatch-pinned; hash change ⇒ blocker, not a silent edit).
- Pre-flight baseline (this run, this tree): `go build ./...` exit 0;
  `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (B1 hold);
  `git branch --show-current` → `WT-user-asset-copy`,
  `git rev-parse HEAD` → `ac0c72ece73c4eca5ac0e01dcf623dead71fa3f0`.
- JD-5 re-pin (measured THIS run, THIS tree ac0c72ece, per dispatch):
  `grep -n "only reinstalls" internal/cli/doctor.go` →
  `986:		// 'moai update' (which only reinstalls the manifest's own skills) would be`
  exit 0 — matches the lane's earlier measurement on `03306622e`. AC-020's
  RED arm re-measures this at flip time (current-tree value at the M4 flip).

### M0 — bundle taxonomy and L0 resolution (2026-10-06)

- RED (captured before the catalog restructure, tree ac0c72ece):
  `go test ./internal/template/ -run TestUserInstallView` → 4/5 FAIL with
  the named requirement drift: "catalog core.agents carries non-L0 agents
  [e2e-tester manager-design manager-git manager-todo super-advisor] —
  reclassify them into bundles (D-Q5)"; "factory entry moai-factory-foreman
  does not declare its manager-lead agent dependency"; 17× "published
  command skill X is not a catalog entry — the catalog is the membership
  SSOT"; "L0 command skill \"moai-plan\" missing from catalog";
  "MATRIX_GAP_UNRESOLVED ... moai-ref (wildcard)"; "matrix row missing for
  manager-git"; "moai-domain-html-report ... sits in the L0 view". (E8
  verbatim output held in the lane transcript; the guard file is
  internal/template/catalog_user_install_view_test.go.)
- GREEN: same command → `ok github.com/modu-ai/moai-adk/internal/template`.
  The catalog user-install view: `catalog.core` = L0 (19 skills = the
  FOURTEEN-skill closure + factory pair + published plan/run/sync trio;
  6 agents = D-Q1 five + manager-lead under the factory entry's
  depends_agents), per-entry `depends_skills`/`depends_agents` edges
  (REQ-004; the dispatcher's dep list carries the matrix's cross-reference
  set for R-f-② maintenance), six standing packs unchanged (devops keeps
  the L0 trio — the E3 shared case), five D-Q5 theme bundles (commands 14,
  consult 4, creative 3, delivery 2, ops-tools 8), hashes regenerated via
  `gen-catalog-hashes --all` (script mirrors extended for the new fields).
- DERIVATION MATRIX (mechanically derived this tree; the guard pins it):
  8 rows — moai-ref-api-patterns→backend, moai-ref-react-patterns→frontend,
  moai-domain-database→backend, moai-ref-seo→frontend,
  moai-ref-supply-chain→devops, moai-workflow-loop→ops-tools,
  moai-domain-frontend→frontend (all conditional per-mission injections on
  the default chain) + manager-git→delivery (the R-b precheck row). The
  plan's four known rows are present; the remaining four are the same class
  derived by the same sweep. html-report classified OUT (explicit list,
  JD-4); the derived L0 closure equals exactly the fourteen-skill union.
- Fallout updates (in-scope reclassification consequences, each named in
  the diff): slim_fs computeDenySet never denies a path that is also a core
  entry (E3 shared wins); catalog audit tests re-derived (tier counts 25/44,
  per-root population comparison, E3-shared duplicate rule); two cli
  placement fixtures re-pointed (init_codex_only, init_force_manifest —
  the pre-SPEC full-17 project placement the SPEC retires).
- Verification (this run, this tree): `go test ./internal/template/
  -count=1` ok; `go test ./internal/userassets/` ok (see M1);
  `go test ./internal/cli/ -run 'TestInit|TestUpdate|TestSkillMirror|
  TestMirror|TestPublished|TestSlim|TestDeployer'` ok;
  `GOOS=windows GOARCH=amd64 go build ./...` exit 0; full internal/cli
  suite under slot lease — output in the M0 commit's evidence block.

### M1 — per-user manifest subsystem (2026-10-06)

- New package `internal/userassets` (design §2.2): manifest at
  `~/.moai/user-assets.json` (schema_version=1, bundles, files with
  per-file sha256/bundle/installed_at/moai_version — no top-level version,
  REQ-006; collisions), the pending-install journal with the FULL
  completeness record (selection delta + full provenance + write-completion
  flag, R-e/E4/E5), the three-case recovery lattice documented at the
  journal's head (incl. R-f-① honest re-stamp), the user-level lock
  (O_EXCL + stale takeover, no flock — B1 portability), C2 root slugs +
  relpath hygiene, REQ-021 unknown-field preservation on every write
  (top-level + per-file; value-receiver MarshalJSON because map values are
  not addressable), schema-refusal gate (CanRemove), corrupt-JSON typed
  error (never auto-delete).
- RED: `go test ./internal/userassets/` → build failure "undefined: Load /
  ManifestPath / SchemaVersion ..." — the M1 surface did not exist (E8).
- GREEN: 11/11 tests ok; two in-round defects caught by the tests and fixed
  (lock-home mkdir; map-value marshal receiver). `go vet` clean; gofmt
  clean; GOOS=windows build exit 0.

### Leader mid-run guidance (2026-10-06, codex gate on the in-flight M0 tree)

- P1 SEQUENCING DECISION (option a — stated per the dispatch): the M3
  `moai bundle add|remove` command is PULLED FORWARD into the M2 commit
  series, landing with the installer that consumes it. The catalog
  exclusion of manager-git takes deployment effect at M0 as restructured;
  from M2 on, the recovery command `moai bundle add delivery` exists, so
  the no-remediation window closes at the next milestone rather than
  staying open until M3. The plan's milestone ORDER is unchanged (M0→M1→
  M2→...); one M3 item moves earlier into M2's commit, and the remaining
  M3 scope (update phase, REQ-009 removal, upgrade branch, R-b precheck
  wiring) stays in M3.
- P2 FIX (harness_fs relocation filter): `newCodexOnlyDeployer` and
  `newProfileDeployer` derived the relocation catalog root from the RAW
  embed — every skill directory re-homed into `.agents/skills` regardless
  of tier (the leader reproduced moai-workflow-loop deploying with no
  bundle selection). Fixed: both constructors now derive the relocation
  root from the SlimFS-filtered skills root. RED observed first:
  `TestCodexOnlyRelocationRespectsCatalogFilter` → "CATALOG_FILTER_LEAK:
  .agents/skills/moai-workflow-loop/SKILL.md visible in codex-only
  deployment" (reproducing the finding) → GREEN after the fix, with the
  absence assertions pinning the unselected-bundle skill OUT and the L0
  skills IN.

### M2 — user-folder installer, init trigger, bundle command (2026-10-06)

- Installer engine (`internal/userassets` install.go + remove.go): the
  per-asset-state judgment over the four roots (skills land whole-tree in
  BOTH harness roots; agents land flat in both — the Claude body from
  .claude/agents/moai/<name>.md, the Codex body from the emitted
  .codex/agents/moai/<name>.toml, REQ-022); the REQ-023 truth table
  (absent/up-to-date/manifest-match/manifest-stale/divergent/collision);
  the REQ-013 per-file fail-open; the pending-install journal
  reconciliation before any collision judgment (the three-case E5 lattice
  incl. R-f-① honest re-stamp and the R-e selection adoption); the C2
  confinement writer (resolved roots once per run, confinedMkdir building
  the chain WITHOUT following an escaping symlink — the plain-MkdirAll
  variant was caught by the parent-symlink sentinel test writing a
  directory through the symlink; leaf-symlink refusal; temp+rename with
  the parent re-validated immediately before it; the REQ-023 backup home
  resolved like the four roots — the macOS /var→/private/var form caught
  by the divergence test). Bundle removal carries the one REQ-009 rule
  (manifest-hash OR shipped-bytes), the E3 complement, and the R-f-②
  deferral — whose deferral set EXCLUDES the dispatcher's matrix-class
  conditional deps (documented at the RemoveBundle head: including them
  would defeat the D28 selection-based prune, AC-018's shipped-but-
  deselected arm).
- RED evidence (E8): `go test ./internal/cli/ -run TestBundle` → 4 FAIL
  ("unknown command" surface absent — `runBundleAdd` undefined at
  authorship); the installer tests went RED first as the absent-surface
  build failure, then through four implementation defects each caught by
  a named arm (fixture source prefix; dirTargets walk prefix; backup-home
  creation; /var symlink resolution; MkdirAll-through-symlink).
- GREEN: internal/userassets 17/17 ok; `go test ./internal/cli/ -run
  'TestBundle'` 5/5 ok (add exactness + selection record; the E3
  complement arm over the REAL devops pack — the L0 trio survives with
  report notes, llm-security/supply-chain removed; divergence honored at
  removal; unknown-name refusal with the valid set; the R-f-② deferral
  over the moai-e2e→e2e-tester edge); `go test ./internal/cli/ -run
  'TestInitEnsuresUserAssets|TestInitBundlesFlagRecordsSelection'` ok —
  init installs L0 user-side end-to-end (AC-001/002 presence arms) and
  `--bundles` records the selection (REQ-004).
- Fold A2 dispatcher rebind at SOURCE: all 19
  `.claude/skills/moai/workflows/` references in
  templates/.claude/skills/moai/SKILL.md rebound to the
  installed-skill-relative `workflows/` form (one form resolves in
  ~/.claude/skills/moai/ and $HOME/.agents/skills/moai/ alike); catalog
  hashes regenerated (`gen-catalog-hashes --all`).
- init wiring: the `--bundles` flag registered on initCmd (+ the test
  cmd mirror); the ensure call after the project deploy succeeds —
  systemic failures fail init with the idempotent-retry hint; per-file
  failures surface in the summary.
- Suite verdict attribution (this run, this tree): the change-scoped
  families `TestInit|TestUpdate|TestBundle|TestCodexOnly|TestSkillMirror|
  TestPublished|TestSlim` → exit 0, 0 failures (20m timeout envelope);
  `golangci-lint run internal/cli/` 0 issues; `GOOS=windows` build exit 0.
  The FULL internal/cli suite exceeded its wall-clock envelope TWICE on
  this loaded machine (10m default kill mid-package; a 25m kill at
  1500.9s with ZERO test-level failures — the package never reached
  completion; the two earlier named failures
  (TestStopChainMemberCostWithinBudget, TestCodexTaskBackgroundHandshake-
  HonorsTaskBound) are pre-existing load-sensitive timing tests that pass
  standalone and on -count=3, touch no code this SPEC changes). The
  repository-wide test verdict is owned by the CI run on origin/develop —
  PENDING at report time.

### M3 — `moai update` user-asset phase (2026-10-06)

- runUserAssetUpdatePhase (user_asset_phase.go): refresh (REQ-008 — the
  installer's manifest-match arm), the SELECTION-BASED prune (REQ-009 —
  PruneUnselected: every manifest-tracked file whose owning entry left
  L0 ∪ the recorded selection, under the one removal rule incl. the
  missing-file entry-drop arm and REQ-023 preserve+backup; the recorded
  selection is passed back to Install so it is honored, never reset),
  REQ-011 summary, all under the user lock, placed BEFORE the project
  phase (REQ-024 upgrade-arm ordering). Journal reconciliation at run
  start is the installer's own first act. ensureGlobalSettingsEnv
  untouched.
- RED (E8): `go test ./internal/userassets/ -run TestPrune` →
  "in.PruneUnselected undefined" (absent surface); cli
  TestUpdatePhase* authored red at the same surface.
- GREEN: prune tests 3/3 (D28 flip arm over the deselect artifact,
  L0-never-a-candidate + REQ-023 preserve, missing-file entry drop);
  update-phase tests 2/2 (AC-005 verdict basis — current==manifest≠
  shipped → rewritten to shipped bytes + hash re-recorded; REQ-004
  update-honors + the deselect prune). R-b Route B precheck: the guard
  test RED observed ("ROUTEB_PRECHECK_MISSING ... does not carry the
  `moai bundle add delivery` remediation") → both entry points wired
  (sync/delivery.md Route B row + run/task-decomposition.md Route B /
  Phase 19): verify user-side role body + refuse with the named
  remediation, C4.
- Verification: TestRouteBPrecheck ok; userassets full ok; cli
  M3-scoped families ok; lint 0 issues (this run, this tree).

### M4 — project slimming, migration, doctor repoints (2026-10-06)

- THE CLASS-CLAUSE REBIND (design §2.5, two populations): the MOVED-asset
  population — 110 project-relative occurrences across the user-scope
  deployed tree (workflows tree incl. subdirectories, agent sources,
  dispatcher internals) rewritten to user-folder paths
  (`.claude/skills/…` → `~/.claude/skills/…` etc.); zero-hit-after-rebind
  measured on THAT population (raw-pattern grep → 0). The
  PROJECT-RETAINED population — run/phase-execution.md rules refs, run.md
  `.claude/rules` refs + trace-ledger hook, sync.md quality-gate hook,
  plan.md spec-workflow pointer — untouched and verified present at their
  project paths. Specific pins: all 17 command sources rebound
  (`~/.agents/skills/moai/SKILL.md`), regenerated via COMMAND_EMIT_UPDATE
  (the golden dispatch-branch pin updated to the user-folder form);
  AGENTS.md.tmpl:40-41 skill-path sentences rebound; templates/CLAUDE.md
  :31/:47 (retained-file→moved-asset) rebound with post-rewrite
  verification; agent sources swept (8 files incl. the e2e-tester.md
  measured instance) + `AGENTEMIT_UPDATE=1` TOML regeneration; the
  dispatcher's 19 internal workflow refs rebind at source (M2, restated
  here as part of the family).
- REQ-005 slimming: `isCommonAssetRoot` excludes `.claude/skills/`,
  `.claude/commands/moai/`, `.claude/agents/moai/`, `.agents/skills/`,
  `.codex/agents/moai/` from the deploy walk in EVERY mode — the project
  payload carries no common skill or agent file.
- MIRROR-REPAIR TERMINATION (final-class item 6): update.go's
  `repairSkillMirrorBestEffort` call removed; `update_mirror_heal.go` +
  `skill_mirror_repair.go` + their three test files deleted (Path B
  re-created the 17 published copies, undoing the migration);
  `pluralMirrorEntries` helper relocated to update_migrate.go.
- REQ-020 MIGRATION: `migrateProjectCommonAssets` (wired into runUpdate
  after the user-asset phase — the ordering IS the per-asset gate):
  provenance-classified removal — template_managed files removed ONLY
  after the user counterpart confirms (user-manifest tracked with hash);
  user_modified preserved + reported (C6); user_created untouched; a
  file whose counterpart is unconfirmed stays project-side + reported
  (REQ-024 upgrade arm, machine states a/b/c honored by the phase
  ordering + the gate).
- DOCTOR REPOINTS (fold A3 + R-d ii + JD-10 + JD-20):
  `checkSkillsAllowlist` and `runHarnessCheck`'s L4 read the user
  install when the PROJECT dir is absent (fallback shape keeps test
  fixtures and pre-migration projects on the project reading — no
  real-HOME leak); `inspectSkillMirror` passes clean when the user
  install is present; `countCodexAgentTOMLs` prefers
  `~/.codex/agents/*.toml`; `agentDirsFor` (web console) prepends
  `~/.claude/agents` so post-migration rows do not vanish.
- Template-guard fallout handled in-scope: SPEC-ID scrub from 4 template
  files (the C1 leak guard), dogfood-mirror sync (rule-mirror +
  pipeline-carry + late-branch pairs), the codex-only profile test
  converted to the AC-011 placement set (absence of the project
  placement + presence of the user set), the relocation-filter test's
  L0-only expectation.
- Verification (this run, this tree): template family + web +
  userassets ok; cli families (Init/UpdatePhase/Bundle/CodexOnly/
  Doctor/Harness/CheckSkills) rerun after the two fixture fixes →
  broad run in flight at record time; GOOS=windows build exit 0;
  golangci-lint 0 issues.

### M7 — deployer_mode retirement (2026-10-06)

- REMOVED (REQ-018): DeployModePlugin + PluginMirrorPolicy constants,
  WithDeployMode/WithPluginMirrorPolicy options, the deployer's mode/
  policy fields, pluginModeExcluded + isPluginExcludedPath walk branch,
  stripMoaiFromMcpJSON, the plugin re-home path (pluginRehomedMirror +
  rehomeOneSkill), RehomeExistingMirrorEntries, and
  deployer_mode_test.go with the surface (the plan's deletion binding).
  The retired mirror-policy test files whose subject died with the
  project-side placement went with them (published_skills_deploy_test,
  skill_mirror_test, skill_mirror_fallback_test,
  skill_mirror_release_test, skill_mirror_manifest_test,
  codex_agents_deploy_test + the plugin-mode update tests), with the
  shared fixture helpers (threeSkillFS, sameStringSlice) recovered into
  mirror_test_helpers_test.go for the surviving tests. Dead code that
  lost its only callers went in the same change: updateDroppedRootTargets,
  classifiedCleanTarget, migrationTemplateContext, the migration-plan
  count fields, and the mirror syscall seams.
- WHAT REMAINS: DeployMode survives as the project's deployment_mode
  RECORD surface only (REQ-018: update never flips the record) —
  DeployModeLocal is the single payload shape; isCommonAssetRoot stays
  (the REQ-005 walk exclusion); resolveUpdateDeployMode resolves LOCAL
  on every arm; init/update construct deployers with NO mode options.
- JD-11 disposition: the carrier's absence is the EXPECTED post-M6
  state — doctor carries no failure-signaling retirement check; the
  Plugin Migration advisory row (M5) separates the manual uninstall
  guidance from any failure signal, and the removed carrier rows cannot
  fire. The recorded exit-1 grep semantics apply to residue sweeps
  (boundary greps below), not to deleted files.
- Boundary greps (this run, this tree): zero references to the retired
  identifiers in non-test Go code (grep exit 1, 0 hits — the tombstone
  comment reworded so even comment literals read zero);
  `go build ./...` + GOOS=windows exit 0; golangci-lint 0 issues;
  template family (Deployer/SlimFS/UserInstallView/CodexOnly/
  EmbeddedSkill/HarnessProfiles) ok; cli Update family ok (43s).
- 429-interruption note: the mid-surgery tree was resumed and completed
  in this session; the scratch generator scripts at the repo root were
  deleted (B8), and the llm.yaml runtime drift was restored to HEAD
  (not M7 scope).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
