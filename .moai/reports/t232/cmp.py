import json
import os
import subprocess
d = json.load(open('.moai/reports/t232/analysis-repro.json'))
g = set()
for e in d:
    if os.path.exists(e['file']) and subprocess.run(['grep', '-F', '-q', '--', e['clause'], e['file']]).returncode == 0:
        g.add(e['id'])
vok = {e['id'] for e in d if e['clause_ok']}
print("validator_ok", len(vok), "grepF_ok", len(g))
print("validator_ok_but_not_grep", len(vok - g), sorted(vok - g))
print("grep_ok_but_not_validator", len(g - vok))
