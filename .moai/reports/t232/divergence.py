import json
import os
import subprocess
d = json.load(open('.moai/reports/t232/analysis-repro.json'))
files = sorted({e['file'] for e in d})
base = 'internal/template/templates'
for f in files:
    t = os.path.join(base, f)
    r = subprocess.run(['diff', '-q', f, t], capture_output=True)
    n = sum(1 for e in d if e['file'] == f)
    print(("SAME " if r.returncode == 0 else "DIFF "), f"entries={n:3d}", f)
