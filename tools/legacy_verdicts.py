#!/usr/bin/env python3
"""Regenerate the verdict of every row of .moai-cli/manifests/legacy-asset-map.json.

User decision 2026-11-17 (Go generation inheritance, D8 = A): concepts are redesigned, files
are not ported, and every row gets `adopted(<file>)`, `rejected(<reason>)` or
`deferred(<design or SPEC id>)`; no row stays `unchecked`. The verdict is a pure function of
the row's path, so the map can be rebuilt byte for byte.

Usage: python3 -I tools/legacy_verdicts.py [--write]
Without --write it compares the regenerated map with the committed one and exits 1 on any
difference. Either way it prints the verdict counts as one JSON line.
"""
import collections
import json
import pathlib
import sys

ROOT = pathlib.Path(__file__).resolve().parents[1]
MAP = ROOT / ".moai-cli/manifests/legacy-asset-map.json"
TEMPLATE = "internal/template/templates/"
PURPOSE = (
    "Legacy (Go-generation) assets of the 02-A inventory (#1991). Every row carries a verdict "
    "(user decision 2026-11-17, Go generation inheritance; D8 = A): adopted(<file that carries "
    "it>), rejected(<reason>) or deferred(<new design or SPEC id>); no row stays unchecked. "
    "Regenerate with tools/legacy_verdicts.py."
)

ATLAS = "design:code-atlas-read-plan"
GUIDES = "design:language-domain-guides"
DIAGNOSTICS = "design:edit-time-diagnostics-hook"
GENERATION = "design:project-agent-skill-generation"

# The 16 rows already shipped (card note 6783558175): each source exists twice, once in the
# repository and once under the Go installer's template folder.
ADOPTED = {
    ".claude/skills/moai/SKILL.md": "skills/moai/SKILL.md",
    ".claude/skills/moai/workflows/plan.md": "skills/moai/workflows/plan.md",
    ".claude/skills/moai/workflows/run.md": "skills/moai/workflows/run.md",
    ".claude/skills/moai/workflows/sync.md": "skills/moai/workflows/sync.md",
    "AGENTS.md": "src/agents.rs",
    "AGENTS.md.tmpl": "src/agents.rs",
    ".claude/output-styles/moai/moai.md": "templates/claude/output-styles/moai.md",
    ".claude/output-styles/moai/moai-easy.md": "templates/claude/output-styles/moai-easy.md",
    ".claude/output-styles/moai/moai-learn.md": "templates/claude/output-styles/moai-learn.md",
}

NOT_PORTED = "not ported (D8 = A)"
R_ENTRY = "Go-era entry point; the single /moai skill skills/moai/SKILL.md replaces it"
R_ROUTER = "Go-era router workflow with no Rust route; " + NOT_PORTED
R_EMPTY = "empty placeholder file of a Go-era folder; nothing to port"
R_MX = "Go-era @MX tag workflow; superseded by the ACC @cc: design (decision 2026-11-17)"
R_RULE = (
    "Go-era rule file; " + NOT_PORTED + "; the Rust rules live in AGENTS.md and "
    "skills/moai/references, with no host-capabilities.json, resources/ or adapters/codex/ "
    "destination"
)
R_HOOK_RULE = (
    "Go-era rule for the Go hook scripts; " + NOT_PORTED + "; the Rust hook engine is src/hooks.rs"
)
R_ROLE = "Go-era role definition; " + NOT_PORTED + "; the Rust base agents are in agents/"
R_METHOD = "Go-era method skill; " + NOT_PORTED
R_PROJECT = (
    "Go-era project method skill; " + NOT_PORTED + "; the /moai project route is "
    "skills/moai/workflows/project.md"
)
R_FOUNDATION = "Go-era foundation knowledge skill; " + NOT_PORTED
R_REPO = "repository-specific Go-era tooling; not part of the product"
R_OPS = "Go-era operations skill; " + NOT_PORTED
R_ARCHIVE = "retired in the Go generation before the Rust product"
R_CATALOG = "Go installer catalog; " + NOT_PORTED

# Router workflows and support files whose job a Rust route does.
ROUTER_REPLACED = {
    ".claude/skills/moai/workflows/moai.md": "skills/moai/SKILL.md",
    ".claude/skills/moai/workflows/project.md": "skills/moai/workflows/project.md",
    ".claude/skills/moai/workflows/project/doc-generation.md": "skills/moai/workflows/project.md",
    ".claude/skills/moai/workflows/project/mode-detection.md": "skills/moai/workflows/project.md",
    ".claude/skills/moai/workflows/review.md": "skills/moai/workflows/audit.md",
    ".claude/skills/moai/workflows/gate.md": "skills/moai/workflows/run.md",
    ".claude/skills/moai/workflows/plan/clarity-interview.md": "skills/moai/references/internal/interview.md",
    ".claude/skills/moai/workflows/plan/spec-assembly.md": "skills/moai/workflows/plan.md",
}
ROUTER_REPLACED_DIRS = {
    ".claude/skills/moai/workflows/run/": "skills/moai/workflows/run.md",
    ".claude/skills/moai/workflows/sync/": "skills/moai/workflows/sync.md",
}
METHOD_REPLACED = {
    "moai-workflow-tdd": "skills/moai/references/internal/test-design.md and tdd-results.md",
    "moai-workflow-spec": "moai spec (src/spec_doc.rs)",
    "moai-workflow-worktree": "moai worktree-plan (src/main.rs)",
    "moai-workflow-docs-claim-check": "skills/moai/references/internal/doc-claims.md",
}
REF_REJECTED = {
    "moai-ref-aside-browser": "Go-era Aside browser reference; skills/moai/references/adapters/aside.md replaces it",
    "moai-ref-cross-model-audit": "Go-era cross-model audit reference for the Go auditor roles; " + NOT_PORTED,
    "moai-ref-jev-question-design": "Go-era System One question-design reference; " + NOT_PORTED,
}

ATLAS_PATHS = {
    ".claude/skills/moai/workflows/codemaps.md",
    ".claude/skills/moai/workflows/project/codebase-analysis.md",
    ".claude/skills/moai/workflows/plan/context-discovery.md",
    ".agents/skills/moai-codemaps/SKILL.md",
    ".claude/commands/moai/codemaps.md",
    ".claude/commands/moai/codemaps.md.tmpl",
}
DIAGNOSTICS_PATHS = {
    ".claude/skills/moai/workflows/fix.md",
    ".claude/skills/moai/workflows/loop.md",
}
GENERATION_PATHS = {
    ".claude/skills/moai/workflows/harness.md",
    ".claude/skills/moai/workflows/harness-builder.md",
    ".claude/skills/moai/workflows/harness-build-entry.md",
    ".claude/skills/moai/workflows/project/meta-harness.md",
    ".agents/skills/moai-harness/SKILL.md",
    ".claude/commands/moai/harness.md",
    ".claude/commands/moai/harness.md.tmpl",
    ".claude/agents/moai/builder-harness.md",
    ".codex/agents/moai/builder-harness.toml",
    ".claude/rules/moai/development/agent-authoring.md",
    ".claude/rules/moai/development/skill-authoring.md",
    ".claude/rules/moai/development/skill-writing-craft.md",
}
GENERATION_SKILLS = {"moai-meta-harness", "moai-harness-learner", "moai-jev-skill-suggestion"}
MX_PATHS = {
    ".claude/skills/moai/workflows/mx.md",
    ".claude/skills/moai/references/mx-tag.md",
    ".claude/rules/moai/workflow/mx-tag-protocol.md",
    ".agents/skills/moai-mx/SKILL.md",
    ".claude/commands/moai/mx.md",
    ".claude/commands/moai/mx.md.tmpl",
}
HOOK_RULES = {
    ".claude/rules/moai/core/hooks-system.md",
    ".claude/rules/moai/core/agent-hooks.md",
}
FOUNDATION = {"moai-foundation-core", "moai-foundation-cc", "moai-foundation-thinking"}
OPS = {"moai-lane-watchdog", "moai-factory-foreman", "moai-fetch-web"}


def source(path):
    """The row's path with the Go installer's template prefix removed."""
    return path[len(TEMPLATE):] if path.startswith(TEMPLATE) else path


def classify(path):
    """Returns (verdict, argument) for one row path; an unknown path raises ValueError."""
    p = source(path)
    if p in ADOPTED:
        return "adopted", ADOPTED[p]
    if p in ATLAS_PATHS or (p.startswith(".claude/skills/moai-workflow-project/") and "/navigator" in p):
        return "deferred", ATLAS
    if p in DIAGNOSTICS_PATHS or p.startswith(".claude/skills/moai-workflow-loop/"):
        return "deferred", DIAGNOSTICS
    if p in GENERATION_PATHS:
        return "deferred", GENERATION
    if p in MX_PATHS:
        return "rejected", R_MX
    if p.endswith("/.gitkeep"):
        return "rejected", R_EMPTY
    if p.startswith(".moai/archive/skills/"):
        return "rejected", R_ARCHIVE
    if p.startswith("internal/template/"):
        return "rejected", R_CATALOG
    if p.startswith(".claude/rules/"):
        if p.startswith(".claude/rules/local/"):
            return "rejected", R_REPO
        if p.startswith((".claude/rules/moai/languages/", ".claude/rules/moai/quality/")) or (
            p == ".claude/rules/moai/development/coding-standards.md"
        ):
            return "deferred", GUIDES
        if p in HOOK_RULES:
            return "rejected", R_HOOK_RULE
        return "rejected", R_RULE
    if p.startswith((".claude/agents/harness/", ".claude/commands/harness/")):
        return "rejected", R_REPO
    if p.startswith((".claude/agents/moai/", ".codex/agents/moai/")):
        return "rejected", R_ROLE
    if p.startswith((".claude/commands/moai/", ".agents/skills/")) or p == ".claude/skills/verify/SKILL.md":
        return "rejected", R_ENTRY
    if p.startswith(".claude/skills/moai/"):
        target = ROUTER_REPLACED.get(p) or next(
            (t for d, t in ROUTER_REPLACED_DIRS.items() if p.startswith(d)), None
        )
        if target:
            return "rejected", "Go-era router workflow; " + target + " replaces it"
        return "rejected", R_ROUTER
    if p.startswith(".claude/skills/"):
        name = p.split("/")[2]
        if name.startswith("hns-"):
            return "rejected", R_REPO
        if name in REF_REJECTED:
            return "rejected", REF_REJECTED[name]
        if name.startswith(("moai-domain-", "moai-ref-")) or name == "moai-foundation-quality":
            return "deferred", GUIDES
        if name in GENERATION_SKILLS:
            return "deferred", GENERATION
        if name in FOUNDATION:
            return "rejected", R_FOUNDATION
        if name in METHOD_REPLACED:
            return "rejected", "Go-era method skill; " + METHOD_REPLACED[name] + " replaces it"
        if name == "moai-workflow-project":
            return "rejected", R_PROJECT
        if name.startswith("moai-workflow-"):
            return "rejected", R_METHOD
        if name in OPS:
            return "rejected", R_OPS
    raise ValueError("no verdict rule for " + path)


def dump(doc):
    """The map's own layout: two-space header, one asset per line."""
    head = {k: v for k, v in doc.items() if k != "assets"}
    text = json.dumps(head, ensure_ascii=False, indent=2)
    rows = ",\n".join("    " + json.dumps(r, ensure_ascii=False) for r in doc["assets"])
    return text[:-2] + ',\n  "assets": [\n' + rows + "\n  ]\n}\n"


def regenerate(raw):
    """Returns (new text, verdict counts) for the map text `raw`."""
    doc = json.loads(raw)
    doc["purpose"] = PURPOSE
    counts = collections.Counter()
    for row in doc["assets"]:
        verdict, arg = classify(row["path"])
        if verdict == "adopted" and not (ROOT / arg).is_file():
            raise ValueError(row["path"] + ": adopted file " + arg + " is missing")
        row["status"] = verdict + "(" + arg + ")"
        counts[verdict] += 1
    return dump(doc), dict(sorted(counts.items()))


def main(argv):
    write = argv[1:] == ["--write"]
    if argv[1:] not in ([], ["--write"]):
        print("usage: python3 -I tools/legacy_verdicts.py [--write]", file=sys.stderr)
        return 2
    raw = MAP.read_text(encoding="utf-8")
    text, counts = regenerate(raw)
    print(json.dumps({"verdicts": counts, "rows": sum(counts.values())}))
    if write:
        MAP.write_text(text, encoding="utf-8")
        return 0
    if text != raw:
        print("legacy-asset-map.json differs from the regenerated map", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
