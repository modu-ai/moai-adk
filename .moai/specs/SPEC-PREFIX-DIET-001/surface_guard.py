# Path-allowlist + content guard for SPEC-PREFIX-DIET-001 (run-phase, pre-merge evaluation only).
# Usage (worktree root): python3 .moai/specs/SPEC-PREFIX-DIET-001/surface_guard.py <BASE-SHA> [forbidden-surface ...]
# Surfaces: output-styles, settings, agents, tests, docs. A forbidden surface turns any change on it into a
# violation (positive control). Anything outside the allowlist is a violation by default.
# Exit 0 + final line "surface-guard=PASS" only when every changed or new file passes; otherwise exit 1.
import re
import subprocess
import sys

base = sys.argv[1]
forbid = set(sys.argv[2:])


def git(*a):
    return subprocess.run(["git", *a], capture_output=True, text=True, check=True).stdout


def show(path):
    r = subprocess.run(["git", "show", base + ":" + path], capture_output=True, text=True)
    return r.stdout if r.returncode == 0 else None


def frontmatter(t):
    return t.split("---")[1]


def body(t):
    return t.split("---", 2)[2]


def drop_description(fm):
    return re.sub(r"^description:.*?(?=^\S)", "", fm, flags=re.S | re.M)


SURFACES = [
    ("output-styles", r"^(internal/template/templates/)?\.claude/output-styles/moai/(moai|moai-easy|moai-learn)\.md$"),
    ("settings", r"^(internal/template/templates/\.claude/settings\.json\.tmpl|\.claude/settings\.json)$"),
    ("agents", r"^(internal/template/templates/)?\.claude/agents/moai/[a-z-]+\.md$"),
    ("tests", r"^internal/template/((output_style|prefix_diet|agent_description)[a-z_]*_test\.go|testdata/(output_style|prefix_diet)[a-z_]*\.(json|txt))$"),
    ("docs", r"^\.moai/(specs/SPEC-PREFIX-DIET-001|reports/t1450)/"),
]

names = set(git("diff", "--name-only", base).split())
names |= set(git("ls-files", "--others", "--exclude-standard").split())
bad = 0
for p in sorted(names):
    surf = next((s for s, rx in SURFACES if re.search(rx, p)), None)
    if surf is None:
        print("VIOLATION outside-allowlist " + p)
        bad += 1
        continue
    if surf in forbid:
        print("VIOLATION forbidden-surface " + surf + " " + p)
        bad += 1
        continue
    msg = "ok"
    if surf in ("output-styles", "agents", "settings"):
        old = show(p)
        new = open(p, encoding="utf-8").read()
        if old is None:
            msg = "VIOLATION new-file-on-guarded-surface"
        elif surf == "output-styles" and frontmatter(old) != frontmatter(new):
            msg = "VIOLATION frontmatter-changed"
        elif surf == "agents" and (
            body(old) != body(new) or drop_description(frontmatter(old)) != drop_description(frontmatter(new))
        ):
            msg = "VIOLATION agent-body-or-nondescription-frontmatter-changed"
        elif surf == "settings":
            d = git("diff", "-U0", base, "--", p).splitlines()
            ch = [l for l in d if l[:1] in "+-" and not l.startswith(("+++", "---"))]
            if any("skillListingBudgetFraction" not in l for l in ch):
                msg = "VIOLATION settings-line-other-than-skillListingBudgetFraction"
    if msg != "ok":
        bad += 1
    print(msg + " " + surf + " " + p)
print("surface-guard=" + ("FAIL" if bad else "PASS"))
sys.exit(1 if bad else 0)
