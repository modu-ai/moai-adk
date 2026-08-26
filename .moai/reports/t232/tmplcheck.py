import json
import os
d = json.load(open('.moai/reports/t232/analysis-repro.json'))
files = sorted({e['file'] for e in d})
base = 'internal/template/templates'
missing = [f for f in files if not os.path.exists(os.path.join(base, f))]
print("distinct_files", len(files), "missing_in_template", len(missing))
for m in missing:
    print("  MISSING", m)
