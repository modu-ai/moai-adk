# t566 census: measures the SPEC-CODEMAPS-REFRESH-002 fold / omission units
# against the current .moai/project/codemaps/*.md, using that SPEC's own hit
# conventions, plus the §A.3(a) hit-0 package census. Read-only.
#
# Usage: python3 census.py <label>   (writes census-<label>.json next to this file)
import glob
import json
import os
import subprocess
import sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "..", ".."))
CODEMAPS = os.path.join(ROOT, ".moai", "project", "codemaps")

# From .moai/reports/t475/verdict.md observation 2 (the run's final judgment).
FOLD = [
    "internal/core/git",
    "internal/cli/doctor_hook_delivery.go",
    "internal/hook/quality/step_git_env.go",
    "internal/kanban/prlink_landedref.go",
    "internal/web/fieldsets_codex_templ.go",
]
OMISSION = [
    "internal/chain",
    "internal/stateanchor",
    "internal/settings/yamlpatch",
    "internal/template/agentemit",
    "internal/template/commandemit",
    "internal/template/published_skills.go",
    "internal/template/skill_mirror_repair.go",
    "internal/cli/skills.go",
    "internal/cli/codex_skills_disable.go",
    "internal/cli/codex_skills_prune.go",
    "internal/cli/integration_settings_drift.go",
    "internal/cli/update_mirror_heal.go",
    "internal/kanban/settings_drift.go",
    "internal/statusline/state_anchor.go",
    "internal/web/codexmirror.go",
]


def main():
    label = sys.argv[1] if len(sys.argv) > 1 else "run"
    md_files = sorted(glob.glob(os.path.join(CODEMAPS, "*.md")))
    texts = {os.path.basename(p): open(p, encoding="utf-8").read() for p in md_files}
    concat_lines = []
    for name in sorted(texts):
        concat_lines.extend(texts[name].splitlines())

    def line_hits(unit):
        # SPEC §A.3(a) convention: grep -c -F over the concatenated six documents.
        return sum(1 for line in concat_lines if unit in line)

    def file_hits(unit):
        # AC-CM2-004 convention: grep -rl -F <unit> .moai/project/codemaps/ (files returned).
        return sorted(name for name, t in texts.items() if unit in t)

    units = {}
    for kind, names in (("fold", FOLD), ("omission", OMISSION)):
        for u in names:
            units[u] = {"kind": kind, "line_hits": line_hits(u), "files": file_hits(u)}

    proc = subprocess.run(
        ["go", "list", "./internal/...", "./cmd/...", "./pkg/..."],
        cwd=ROOT, capture_output=True, text=True, timeout=300,
    )
    pkgs = []
    for line in proc.stdout.splitlines():
        parts = line.split("/", 3)
        if len(parts) == 4:
            pkgs.append(parts[3])
    joined = "\n".join(concat_lines)
    zero = sorted(p for p in pkgs if p not in joined)

    result = {
        "label": label,
        "codemaps_files": sorted(texts),
        "go_list_exit": proc.returncode,
        "go_list_packages": len(pkgs),
        "hit_zero_packages": zero,
        "hit_zero_count": len(zero),
        "fold_hit_zero": [u for u in FOLD if units[u]["line_hits"] == 0],
        "fold_with_hits": [u for u in FOLD if units[u]["line_hits"] > 0],
        "omission_without_hits": [u for u in OMISSION if not units[u]["files"]],
        "omission_count": len(OMISSION),
        "fold_count": len(FOLD),
        "fold_packages_in_hit_zero_census": [u for u in FOLD if u in zero],
        "omission_packages_in_hit_zero_census": [u for u in OMISSION if u in zero],
        "units": units,
    }
    out = os.path.join(os.path.dirname(__file__), "census-%s.json" % label)
    with open(out, "w", encoding="utf-8") as f:
        json.dump(result, f, ensure_ascii=False, indent=2)
    print(json.dumps({k: result[k] for k in (
        "label", "go_list_exit", "go_list_packages", "hit_zero_count",
        "fold_hit_zero", "fold_with_hits", "omission_without_hits",
        "fold_packages_in_hit_zero_census", "omission_packages_in_hit_zero_census")},
        ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
