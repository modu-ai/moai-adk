import json
import os
import subprocess

d = json.load(open('.moai/reports/t232/analysis-repro.json'))
once = 0
multi = []
zero = 0
lengths = []
for e in d:
    c = e['clause']
    lengths.append(len(c))
    if not os.path.exists(e['file']):
        continue
    r = subprocess.run(['grep', '-F', '-c', '--', c, e['file']], capture_output=True, text=True)
    n = int(r.stdout.strip() or 0)
    if n == 0:
        zero += 1
    elif n == 1:
        once += 1
    else:
        multi.append((e['id'], n))
lengths.sort()
print("hit_zero", zero, "hit_once", once, "hit_multi", len(multi), multi)
print("clause_len_median", lengths[len(lengths) // 2], "min", lengths[0], "max", lengths[-1])
print("len_under_20", sum(1 for x in lengths if x < 20))
