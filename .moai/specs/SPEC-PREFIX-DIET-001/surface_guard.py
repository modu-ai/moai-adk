# Path-allowlist + content guard for SPEC-PREFIX-DIET-001 (run-phase, pre-merge evaluation only).
# Usage (worktree root): python3 .moai/specs/SPEC-PREFIX-DIET-001/surface_guard.py <BASE-SHA> [forbidden-surface ...]
# Surfaces: output-styles, agents, codex-tomls, tests, docs, catalog-hashes (run-phase D1: only the `hash:`
# lines of edited agents in internal/template/catalog.yaml). A forbidden surface turns any change on it
# into a violation (positive control). Anything outside the allowlist is a violation by default --
# both settings files are deliberately NOT on the allowlist, so ANY settings change is a violation.
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
    ("agents", r"^(internal/template/templates/)?\.claude/agents/moai/[a-z-]+\.md$"),
    ("codex-tomls", r"^internal/template/templates/\.codex/agents/moai/[a-z-]+\.toml$"),
    ("tests", r"^internal/template/((output_style|prefix_diet|agent_description)[a-z_]*_test\.go|testdata/(output_style|prefix_diet)[a-z_]*\.(json|txt))$"),
    ("docs", r"^\.moai/(specs/SPEC-PREFIX-DIET-001|reports/t1450)/"),
    # Run-phase addition (plan-audit iter3 D1, observed at M0 in progress.md section E.2): after an agent
    # description edit, `gen-catalog-hashes --all` (part of `make build`) rewrites that agent's `hash:` line
    # in catalog.yaml. Allowed ONLY as hash lines of agents whose template .md changed in this diff.
    ("catalog-hashes", r"^internal/template/catalog\.yaml$"),
]


def catalog_hash_violations(path, changed):
    """Return violations for catalog.yaml: only `hash:` lines of edited agent templates may change."""
    diff = git("diff", "-U0", base, "--", path)
    new_lines = open(path, encoding="utf-8").read().split("\n")
    bad = []
    new_no = 0
    for ln in diff.split("\n"):
        m = re.match(r"^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@", ln)
        if m:
            new_no = int(m.group(1))
            continue
        if ln.startswith("---") or ln.startswith("+++") or not ln or ln[0] not in "+-":
            continue
        if not re.match(r"^[+-]\s+hash: [0-9a-f]{64}$", ln):
            bad.append("non-hash-line " + ln.strip())
            if ln[0] == "+":
                new_no += 1
            continue
        if ln[0] == "+":
            entry = ""
            for k in range(new_no - 1, -1, -1):
                s = new_lines[k].strip()
                if s.startswith("path:"):
                    entry = s[len("path:"):].strip()
                    break
            new_no += 1
            if "internal/template/" + entry not in changed or "/.claude/agents/moai/" not in entry:
                bad.append("hash-of-unedited-entry " + entry)
    return bad

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
    if surf == "catalog-hashes":
        cv = catalog_hash_violations(p, names)
        if cv:
            msg = "VIOLATION catalog-change-beyond-edited-agent-hashes (" + "; ".join(cv) + ")"
    if surf in ("output-styles", "agents"):
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
    if msg != "ok":
        bad += 1
    print(msg + " " + surf + " " + p)
print("surface-guard=" + ("FAIL" if bad else "PASS"))
sys.exit(1 if bad else 0)
