import sys, re
files = [
 "CLAUDE.md","AGENTS.md",
 ".claude/rules/moai/core/agent-common-protocol.md",
 ".claude/rules/moai/core/askuser-protocol.md",
 ".claude/rules/moai/core/moai-constitution.md",
 ".claude/rules/moai/core/moai-mcp-tools.md",
 ".claude/rules/moai/core/native-idiom-and-register.md",
 ".claude/rules/moai/core/verification-claim-integrity.md",
 ".claude/rules/moai/workflow/cache-aware-execution.md",
 ".claude/rules/moai/workflow/context-window-management.md",
 ".claude/rules/moai/workflow/cross-session-messaging.md",
 ".claude/rules/moai/workflow/goal-directive.md",
 ".claude/rules/moai/workflow/kanban-dispatch.md",
 ".claude/rules/moai/workflow/main-checkout-branch-guard.md",
 ".claude/rules/moai/workflow/session-handoff.md",
 ".claude/rules/moai/workflow/skill-routing.md",
]
BIND = re.compile(r'\[HARD\]|MUST|shall ')
maxlvl = int(sys.argv[1]) if len(sys.argv)>1 else 3
for f in files:
    txt = open(f, encoding='utf-8').read()
    lines = txt.split("\n")
    infence = False
    heads = []
    for i,l in enumerate(lines):
        if l.lstrip().startswith("```"):
            infence = not infence
            continue
        if infence: continue
        m = re.match(r'^(#{1,6}) (.*)$', l)
        if m and len(m.group(1)) <= maxlvl:
            heads.append((i, len(m.group(1)), m.group(2)))
    bounds = [h[0] for h in heads] + [len(lines)]
    print("=== %s  (%d chars)" % (f, len(txt)))
    for (a,lvl,title),b in zip(heads, bounds[1:]):
        body = "\n".join(lines[a:b])
        nb = sum(1 for l in lines[a:b] if BIND.search(l))
        print("  %6d  b=%-3d L%d %s%s" % (len(body), nb, lvl, "  "*(lvl-1), title[:88]))
