import subprocess, sys
ref = sys.argv[1]
def git(*a):
    return subprocess.run(["git"]+list(a), capture_output=True, text=True).stdout
files = [f for f in git("ls-tree","-r","--name-only",ref).splitlines()
         if f.startswith(".claude/rules/moai/") and f.endswith(".md")]
def has_paths(blob):
    lines = blob.split("\n")
    if not lines or lines[0].rstrip() != "---":
        return False
    for ln in lines[1:]:
        if ln.rstrip() == "---":
            return False
        if ln.startswith("paths:"):
            return True
    return False
total = 0
rules = 0
for f in sorted(files):
    b = git("show", f"{ref}:{f}")
    if not has_paths(b):
        total += len(b.encode()); rules += 1
fixed = ["CLAUDE.md", ".claude/output-styles/moai/moai.md", "MEMORY.md"]
for f in fixed:
    b = git("show", f"{ref}:{f}")
    total += len(b.encode())
print(f"{ref}: rules={rules} bytes={total} tokens~{total//4}")
