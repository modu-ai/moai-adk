import json
import os
import subprocess
d = json.load(open('.moai/reports/t232/analysis-repro.json'))
base = 'internal/template/templates'
ok = 0
for e in d:
    t = os.path.join(base, e['file'])
    if os.path.exists(t) and subprocess.run(['grep', '-F', '-q', '--', e['clause'], t]).returncode == 0:
        ok += 1
print("template_grepF_ok", ok, "of", len(d))
