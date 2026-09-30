"""Compare the two visually-identical stale-loanword occurrences (lines 11, 14)."""
import re
import sys

sys.path.insert(0, "scripts")
import jamo_integrity as J

text = open(".moai/reports/t1336/materials/t1333-verdict-original.md",
            encoding="utf-8").read()
lines = text.split("\n")

occ = []
for ln in (11, 14):
    for m in re.finditer(r"[가-힣]+", lines[ln - 1]):
        occ.append((ln, m.start() + 1, m.group()))

# Pick the 5-syllable loanword occurrences (스테일-shaped = 3 syllables ending 테일/스테일).
cands = [o for o in occ if len(o[2]) == 3 and "일" in o[2]]
print("3-syllable words containing 일 on lines 11/14:", cands)

w11 = [o[2] for o in cands if o[0] == 11][0]
w14 = [o[2] for o in cands if o[0] == 14][0]
print()
print("line 11 form:", w11, "codepoints:", [hex(ord(c)) for c in w11])
print("line 14 form:", w14, "codepoints:", [hex(ord(c)) for c in w14])
print("identical strings:", w11 == w14)
print("slot distance:", J.slot_distance(w11, w14))
print("deltas:", J.slot_deltas(w11, w14))
for ch in set(w11 + w14):
    print(" ", J.describe_syllable(ch))
