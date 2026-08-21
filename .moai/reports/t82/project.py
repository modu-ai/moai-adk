#!/usr/bin/env python3
"""M1 Arm A / Arm B projections. Every input is a measured figure from a sibling
script in this directory; nothing here is hand-entered except the SPEC's own
constants and the two stated structure/diet assumptions, which are labelled.
"""

# --- measured inputs -------------------------------------------------------
CLAUSE_TOTAL = 51639      # surface_r3.py / summarize.py: 97 clause blocks
CODEX_REL = 16135         # classify_report.py: 35 blocks classed C
CLAUDE_ONLY = 35147       # classify_report.py: 61 blocks classed K
PROSE = 357               # classify_report.py: 1 block classed P
RATIO = 2848 / 5355       # pilot_measure.py aggregate, 11 clauses
SURFACE_TOKENS = 71207    # surface_r3.py, guard-exact (sum of per-file len/4)
R3_MOVABLE = 161482       # surface_r3.py, rules + CLAUDE.md, excludes R1 and R2

# --- SPEC constants --------------------------------------------------------
CEILING = 24576           # REQ-AMC-004
BUDGET = 32768            # measured codex project_doc_max_bytes default
RATCHET = 66371           # spec.md C.4

# --- stated assumptions (labelled, not measured) ---------------------------
# ~14 sections x (heading ~40 B + 1-2 framing sentences ~150 B) + preamble ~600 B
STRUCTURE = 3300
# conservative empirical stub-split reduction, the harder of two measured
# precedents: kanban-dispatch 21,003 -> 13,027 (38.0%); goal-directive
# 25,755 -> 6,531 (74.6%).
DIET_LOW, DIET_HIGH = 0.380, 0.746

print("=== Arm A -- contract fits its ceiling")
comp = CODEX_REL * RATIO
proj = comp + STRUCTURE
print("clause-block [HARD] corpus            %8d B" % CLAUSE_TOTAL)
print("  less Claude-only (61 blocks)        %8d B" % -CLAUDE_ONLY)
print("  less prose mention (1 block)        %8d B" % -PROSE)
print("= Codex-relevant verbatim (35)        %8d B" % CODEX_REL)
print("  x measured compression %.4f       %8d B" % (RATIO, comp))
print("  + document structure (assumption)   %8d B" % STRUCTURE)
print("= projected AGENTS.md                 %8d B" % proj)
print("ceiling %d  -> %s (spare %d B, %.0f%% of ceiling used)" % (
    CEILING, "REACHABLE" if proj <= CEILING else "NOT REACHABLE",
    CEILING - proj, 100 * proj / CEILING))
print("budget  %d  -> headroom %d B (floor 8192)" % (BUDGET, BUDGET - proj))
nocomp = CODEX_REL + STRUCTURE
print("sensitivity: zero compression         %8d B -> %s" % (
    nocomp, "still fits" if nocomp <= CEILING else "over"))
b = (CEILING - STRUCTURE) / RATIO
print("sensitivity: break-even Codex-relevant verbatim = %d B (measured %d, %.1fx margin)" % (
    b, CODEX_REL, b / CODEX_REL))
print()

print("=== Arm B -- the diet reaches the ratchet ceiling")
for label, agents_b in (("measured projection", proj), ("REQ-AMC-004 ceiling", CEILING)):
    at = int(agents_b) // 4
    with_contract = SURFACE_TOKENS + at
    cut = with_contract - RATCHET
    print("%-20s |AGENTS.md| %6d B = %5d tok | surface+contract %6d | required cut %6d tok (%d B)"
          % (label, agents_b, at, with_contract, cut, cut * 4))
print()
print("movable R3 mass, rules + CLAUDE.md      %8d B = %6d tok" % (R3_MOVABLE, R3_MOVABLE // 4))
for label, agents_b in (("measured projection", proj), ("REQ-AMC-004 ceiling", CEILING)):
    cut = SURFACE_TOKENS + int(agents_b) // 4 - RATCHET
    print("  required cut as a share of movable R3: %-20s %5.1f%%" % (label, 100 * cut * 4 / R3_MOVABLE))
print("empirical stub-split precedents: %.1f%% (kanban-dispatch) .. %.1f%% (goal-directive) whole-file"
      % (100 * DIET_LOW, 100 * DIET_HIGH))
print("projected yield at the conservative %.1f%% applied to movable R3: %d B = %d tok"
      % (100 * DIET_LOW, DIET_LOW * R3_MOVABLE, DIET_LOW * R3_MOVABLE // 4))
print()

# Tightest defensible bound: apply the conservative whole-file precedent ONLY to
# always-loaded files that have never been stub-split (the five with an existing
# companion have had their easy detail removed already, so their second pass is
# excluded from this bound entirely).
NEVER_STUBBED = {
    "CLAUDE.md": 20523,
    "moai-constitution.md": 18958,
    "cross-session-messaging.md": 16672,
    "verification-claim-integrity.md": 13140,
    "context-window-management.md": 13009,
    "main-checkout-branch-guard.md": 11865,
    "moai-mcp-tools.md": 7357,
    "skill-routing.md": 5825,
    "native-idiom-and-register.md": 4967,
}
ns = sum(NEVER_STUBBED.values())
print("never-stub-split always-loaded files (%d): %d B" % (len(NEVER_STUBBED), ns))
print("  at the conservative %.1f%% whole-file precedent: %d B = %d tok"
      % (100 * DIET_LOW, DIET_LOW * ns, DIET_LOW * ns // 4))
for label, agents_b in (("measured projection", proj), ("REQ-AMC-004 ceiling", CEILING)):
    cut = SURFACE_TOKENS + int(agents_b) // 4 - RATCHET
    yld = DIET_LOW * ns // 4
    print("  vs required cut (%-20s) %6d tok -> %s (margin %+d tok)"
          % (label, cut, "clears" if yld >= cut else "SHORT", yld - cut))
