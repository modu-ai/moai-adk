---
description: >
  Scan codebase and generate architecture documentation in .moai/project/codemaps/ directory.
  Creates module maps, dependency graphs, and entry point references.
  Supports full regeneration and area-specific focus.
  Use when generating architecture documentation or visualizing codebase structure.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "codemaps, architecture, documentation, visualization, codebase-analysis"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 4000

# MoAI Extension: Triggers
triggers:
  keywords: ["codemaps", "architecture", "codebase map", "module map", "dependency graph"]
  agents: ["Explore"]
  phases: ["codemaps"]
---

# Workflow: Codemaps - Architecture Documentation Generation

Purpose: Scan the codebase and generate architecture documentation in `.moai/project/codemaps/` directory. Creates human-readable and AI-consumable maps of modules, dependencies, entry points, and interaction patterns.

Flow: Explore Codebase -> Analyze Architecture -> Generate Maps -> Verify -> Report

## Supported Flags

- --force (alias --regenerate): Regenerate all codemaps even if they already exist
- --area AREA: Focus on a specific area (e.g., --area api, --area auth, --area cli)
- --format FORMAT: Output format (default: markdown). Options: markdown, mermaid, json
- --depth N: Maximum directory depth for exploration (default: 4)

## Pipeline Contract (Agentless Classification)

<!-- @MX:NOTE - Agentless fixed-pipeline classification; localize→repair→validate contract. See spec-workflow.md#subcommand-classification. -->

This subcommand is classified as **Agentless fixed-pipeline**.
It executes a deterministic 3-phase contract: **localize → repair → validate**.

- **Phase mapping**: localize ← Phase 1; repair ← Phase 2+3; validate ← Phase 4
- **No LLM-driven control flow**: Agent() invocations exist for executor delegation within phases (e.g., `Explore` for codebase analysis) but never select the next phase.
- **No-op exit**: When the localize phase finds zero targets, the pipeline exits with status `no-op` and exit code 0, skipping repair and validate.
- **Fail-fast**: When repair encounters an unresolvable error, the pipeline terminates and reports the error. There is no multi-agent fallback.
- **`--mode` flag handling**: Any `--mode` flag passed to this subcommand is ignored. The system logs `MODE_FLAG_IGNORED_FOR_UTILITY` at info level and proceeds with the fixed pipeline.
- **Repeatability**: Even when the parent invocation supplies `--mode loop`, the pipeline runs once per command invocation. Re-entry requires explicit user re-invocation.

See [Subcommand Classification matrix](../../rules/moai/workflow/spec-workflow.md#subcommand-classification) for the full pipeline-vs-multi-agent contract.

## Phase 1: Codebase Exploration

[HARD] Delegate codebase exploration to the Explore subagent.

Exploration Objectives passed to Explore agent:

- Directory Structure: Map all top-level and significant subdirectories
- Module Boundaries: Identify package/module boundaries and their responsibilities
- Entry Points: Find main entry points (main.go, index.ts, app.py, etc.)
- Public APIs: List exported functions, types, and interfaces
- Dependency Graph: Map inter-module dependencies (imports, requires)
- External Dependencies: Catalog third-party dependencies with purposes
- Configuration Files: Identify build, deployment, and config files

If --area flag: Limit exploration to the specified area and its dependencies.

Expected Output from Explore agent:

- Module inventory with purpose descriptions
- Dependency adjacency list (who imports whom)
- Entry point catalog
- Technology stack summary
- Architecture pattern identification (MVC, Clean, Hexagonal, etc.)

### High-Count Extraction Fan-Out (capability-gated)

**Where** `.claude/workflows/codemaps-extract.js` exists on disk **AND** the runtime supports dynamic workflows, **and** the source-package count is high — approaching the full codebase, where parallel wall-clock speed offsets per-agent cost — the orchestrator MAY launch it to layer per-package architecture insight on top of Phase 1's output. **Where** any of those conditions is absent, including the ordinary low-package-count case, Phase 1 and Phase 2 run their existing path with no error, no warning, and no interruption.

The scoping is augmentation, never replacement: `go list -deps -json` + `go doc` — or the project language's equivalent dependency and doc extractors — remain the mechanically complete source for the dependency graph and public surface. The fan-out contributes only review-grade inference (coupling risk, latent contracts, layering judgments, negative-space gaps) that the deterministic baseline cannot produce, so a run that reduces to restating import edges has added nothing. The orchestrator launches the script itself (scaling, not subagent nesting); every extractor is read-only and returns markdown — one missing required input returns a structured blocker report, and no extractor ever prompts the user. Launching it is executor delegation inside Phase 1, so the Agentless pipeline contract above is unaffected — it does not select the next phase.

## Phase 2: Architecture Analysis

The orchestrator performs architecture analysis directly (no Agent() spawn) from the Phase 1 exploration output plus deterministic tooling (e.g., `go list -deps -json` + `go doc`, or the project language's equivalent dependency/doc extractors) — replacing the former analysis delegation spawn.

Analysis inputs:

- Explore agent results from Phase 1
- Existing .moai/project/codemaps/ content (if --force not set, for incremental updates — narrative quality only, NEVER as an authority for package existence: stale docs carry phantom package names that re-enter the output as context, which is how renamed/removed packages survive regeneration)
- Output format preference (from --format flag)

Package existence is decided ONLY by the working tree (deterministic extractors + the Phase 1 exploration), never by what a previous codemaps document says.

Analysis Tasks:

- Classify modules by layer (presentation, business, data, infrastructure)
- Identify high fan-in modules (potential @MX:ANCHOR candidates)
- Detect circular dependencies
- Map request/data flow paths
- Identify domain boundaries

## Phase 3: Map Generation

The orchestrator generates the maps directly (no Agent() spawn) from the Phase 2 analysis — replacing the former generation delegation spawn. Phase 1's Explore spawn remains the workflow's single Agent() invocation.

Output Files in `.moai/project/codemaps/` directory:

- `overview.md`: High-level architecture summary with module descriptions
- `modules.md`: Detailed module catalog with responsibilities and dependencies
- `dependencies.md`: Dependency graph (text and/or mermaid diagram)
- `entry-points.md`: Entry point catalog with invocation paths
- `data-flow.md`: Key data flow paths through the system

If --area flag: Generate only area-specific maps:
- `.moai/project/codemaps/{area}/overview.md`
- `.moai/project/codemaps/{area}/modules.md`
- `.moai/project/codemaps/{area}/dependencies.md`

If --format mermaid: Include mermaid diagrams in documentation.
If --format json: Generate machine-readable JSON alongside markdown.

Fold and omission judgments: when a refresh classifies units as `fold` (folded into a parent's description, so the unit itself must stay unnamed in the documents) or `omission` (left out by an earlier refresh and now owed a description), record every judgment in `.moai/project/codemaps/fold-judgments.txt`, one judgment per line:

```text
# blank lines and lines starting with # are ignored
fold <unit>
omission <unit>
```

A unit is the exact string a reader would search for: a package or module directory path, or a single file path. The file is the input of the Phase 4 fold and omission judgment check. It is a `.txt` file on purpose: every `*.md` file in the directory is scanned for hits, so a judgments file inside that set would match every unit it names. The map generation step writes only the named documents above and never rewrites this file.

## Phase 4: Verification

- Verify all referenced files and modules actually exist — runnable check: `moai graph check` (read the `citations` row; it reports `positive-cited-path-absence`, threshold 0, so any positively-cited absent path is red). A freshness-gate green stamp is NOT accuracy proof — freshness and accuracy are separate axes
- Negative citations — paths cited precisely because they do NOT exist (removed-package records, rename histories, warning notes) — MUST use blockquote form (a `>`-prefixed line): the citations check exempts blockquote lines as negative context. This is a form mandate for negative citations, not a claim that blockquotes may only carry negative citations
- Check that dependency relationships are bidirectionally consistent
- Validate entry points are reachable
- Compare with existing .moai/project/codemaps/ to highlight changes (if not --force)

### Fold and Omission Judgment Check

Runnable check for the judgments recorded in Phase 3, executed from the project root. Each recorded unit is checked on its own. A census of packages with zero hits cannot stand in for this check: a file-granularity unit never appears in such a census, so a fold unit given prose there goes unnoticed.

```bash
# Check recorded fold / omission judgments unit by unit; pass = no UNCOVERED, FOLD-PROSE, or GAP line
dir=".moai/project/codemaps"
judg="$dir/fold-judgments.txt"
if [ ! -r "$judg" ]; then
  echo "GAP: $judg is not readable — fold and omission judgments were not observed"
else
  set -- "$dir"/*.md
  if [ ! -e "$1" ]; then
    echo "GAP: no generated *.md document under $dir — hits were not observed"
  else
    awk '
      { sub(/\r$/, "") }
      FILENAME == ARGV[1] {
        line = $0
        sub(/^[ \t]+/, "", line)
        sub(/[ \t]+$/, "", line)
        if (line == "" || substr(line, 1, 1) == "#") next
        if (line ~ /^(fold|omission)[ \t]+[^ \t]+$/) {
          n++
          kind[n] = line
          sub(/[ \t].*$/, "", kind[n])
          unit[n] = line
          sub(/^[^ \t]+[ \t]+/, "", unit[n])
          if (kind[n] == "fold") { folds++ } else { omissions++ }
        } else {
          bad[++nbad] = "GAP: malformed judgment line " FNR ": " line
        }
        next
      }
      { for (i = 1; i <= n; i++) if (index($0, unit[i]) > 0) hits[i]++ }
      END {
        printf "COLLECTED: fold=%d omission=%d\n", folds, omissions
        for (i = 1; i <= nbad; i++) print bad[i]
        if (n == 0) { print "GAP: 0 judgments collected — fold and omission units were not observed"; exit }
        for (i = 1; i <= n; i++) {
          if (kind[i] == "omission" && hits[i] == 0) print "UNCOVERED: " unit[i]
          if (kind[i] == "fold" && hits[i] > 0) print "FOLD-PROSE: " unit[i] " " hits[i]
        }
      }
    ' "$judg" "$@"
  fi
fi
```

The script narrows where to look; the orchestrator decides. It never prints PASS and always exits 0. The check passes when the output carries no `UNCOVERED:`, `FOLD-PROSE:`, or `GAP:` line.

- `COLLECTED: fold=N omission=M` — the judgments read from `fold-judgments.txt`. Record it with the verification result: it is the measurement the check rests on, and a count that differs from the number of judgments the refresh made means a judgment was never written down.
- `UNCOVERED: <unit>` — an omission unit that no line of the generated documents names. The refresh owed it a description and did not write one.
- `FOLD-PROSE: <unit> <lines>` — a fold unit named on that many document lines. A folded unit gets no prose of its own: remove those lines, or revise the judgment in `fold-judgments.txt` if the unit should be described after all.
- `GAP: …` — the check was not observed: the judgments file is unreadable, it holds zero judgments, a line is neither `fold <unit>` nor `omission <unit>`, or no generated document exists. Report it as a gap, never read it as a pass. A refresh that made no fold or omission judgment also reports this GAP; state that reason rather than creating an empty file.

A hit is a fixed-string line match (the `grep -c -F` reading) over the top-level generated `*.md` documents in `.moai/project/codemaps/` only; area-specific subdirectories and the judgments file are outside the scanned set. A unit matches whether it appears as a whole path or inside a longer one, so a fold unit that is a prefix of a described unit reads as prose: revise the judgment or the wording, not the check.

## Phase 5: Report

Display completion summary in user's conversation_language:

- Files generated: List of created/updated codemaps
- Architecture highlights: Key patterns and notable findings
- Potential issues: Circular dependencies, orphaned modules, high coupling

Next Steps (AskUserQuestion):

- Write SPEC for improvements (Recommended): Create a SPEC to address any architectural issues found. Useful if circular dependencies or high coupling were detected.
- Generate project documentation: Run /moai project to create or update product.md, structure.md, tech.md alongside the new codemaps.
- Review codemaps manually: Open the generated files in .moai/project/codemaps/ directory for manual review and editing.

## Task Tracking

[HARD] Task management tools mandatory:
- Each map file creation tracked as a pending task via TaskCreate
- Before each generation: change to in_progress via TaskUpdate
- After each generation: change to completed via TaskUpdate

## Agent Chain Summary

- Phase 1: Explore subagent (codebase exploration, read-only) — the workflow's single Agent() spawn
- Phase 2-3: MoAI orchestrator (analysis and generation, orchestrator-direct from the Phase 1 exploration output + deterministic tooling)
- Phase 4: MoAI orchestrator (verification checks)

## Execution Summary

1. Parse arguments (extract flags: --force, --area, --format, --depth)
2. Check for existing .moai/project/codemaps/ directory content
3. Delegate codebase exploration to Explore subagent (the single Agent() spawn)
4. Perform architecture analysis and map generation orchestrator-direct (exploration output + deterministic tooling)
5. Verify generated maps for consistency
6. TaskCreate/TaskUpdate for all generated files
7. Report results with next step options

---

Version: 1.0.0
