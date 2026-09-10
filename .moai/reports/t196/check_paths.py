import os, re

# The 9 skill-tree files edited by this card (self-contained so the check stays
# re-runnable; the same list is reproducible via
#   git diff --name-only -- internal/template/templates/.claude/skills
# before the change is committed).
_T = 'internal/template/templates/.claude/skills/'
files = [
    _T + 'moai-domain-svg-infographic/SKILL.md',
    _T + 'moai-workflow-project/references/navigator-audit.md',
    _T + 'moai-workflow-project/references/navigator.md',
    _T + 'moai-workflow-testing/SKILL.md',
    _T + 'moai/SKILL.md',
    _T + 'moai/workflows/harness-build-entry.md',
    _T + 'moai/workflows/harness-builder.md',
    _T + 'moai/workflows/project/meta-harness.md',
    _T + 'moai/workflows/run/task-decomposition.md',
]
pat = re.compile(r'\.claude/skills/[A-Za-z0-9_./-]+')
paths = set()
for f in files:
    for m in pat.findall(open(f, encoding='utf-8').read()):
        m = m.rstrip('.),`"\'')
        if m.endswith('/') or '<' in m:
            continue
        paths.add(m)

ok = bad = 0
for p in sorted(paths):
    d = os.path.exists(p)
    t = os.path.exists(os.path.join('internal/template/templates', p))
    flag = 'OK  ' if (d and t) else 'MISS'
    ok += 1 if (d and t) else 0
    bad += 0 if (d and t) else 1
    print(f"{flag} {p}  deployed={d} template={t}")
print(f"\nresolved-in-both={ok}  unresolved={bad}  total={len(paths)}")
