"""AC-HWD-010 independent basis: count hook COMMAND entries, not matcher groups.

settings.json shape is  hooks[event] -> [ {matcher, hooks:[{type,command},...]} ]
so the entry count is the sum over matcher groups of len(group["hooks"]),
NOT len(hooks[event]).
"""

import json

d = json.load(open(".claude/settings.json"))
h = d["hooks"]

groups = 0
entries = 0
for event, glist in h.items():
    for g in glist:
        groups += 1
        entries += len(g.get("hooks", []))

print("events        :", len(h))
print("matcher groups:", groups)
print("hook entries  :", entries)
print("statusLine present:", "statusLine" in d)
