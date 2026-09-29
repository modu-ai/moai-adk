"""Cross-verify corrupt forms quoted in characterization.md against the material.

The characterization quotes corrupt forms; since this session's own output
channel is the same GLM lane, the transcription itself must be verified
byte-identical to the material — mechanically, not by eye.
"""
import re
import sys

sys.path.insert(0, "scripts")
import jamo_integrity as J

material = open(".moai/reports/t1336/materials/t1333-verdict-original.md",
                encoding="utf-8").read()
charac = open(".moai/reports/t1336/characterization.md", encoding="utf-8").read()

# Corrupt forms as flagged by the checker on the material (source of truth).
mat_words = set(re.findall(r"[가-힣]+", material))
# The known-corrupt forms identified in the analysis (present in material,
# absent from tracked-prose vocab or flagged).
corrupt_forms = ["나먲지", "엄트리", "불볇확인", "신귴", "바로잠펴", "신귵",
                 "숴환", "브랿치", "바이너리", "채넍", "채넎이", "채넞", "엉터닐",
                 "바이넄리", "큜", "팕", "밖라", "읠", "쓛"]

ok = True
for form in corrupt_forms:
    in_mat = form in mat_words
    in_char = form in charac
    status = "OK" if (in_mat and in_char) else "MISMATCH"
    if not (in_mat and in_char):
        ok = False
    print("%s  %-8s material=%s characterization=%s" % (status, form, in_mat, in_char))

# The visually-identical stale variant: verify the exact codepoints quoted in
# characterization.md line for the stale case match the material's occurrence.
mat_stale = None
for line in material.split("\n"):
    pass
# Find the 3-syllable stale-shaped word in material line 14 that differs from 스테일.
for ln in material.split("\n"):
    for m in re.finditer(r"[가-힣]{3}", ln):
        w = m.group()
        if J.slot_distance(w, "스테일") == 2 and w != "스테일":
            mat_stale = w
print("material stale-variant form: %s (codepoints %s)"
      % (mat_stale, [hex(ord(c)) for c in mat_stale] if mat_stale else None))
# characterization quotes the corrupt syllable as `슠`테일 (split by quotes),
# so check the leading syllable's presence rather than the joined form.
stale_char = mat_stale[0] if mat_stale else None
print("stale corrupt syllable %s present in characterization: %s"
      % (stale_char, stale_char in charac if stale_char else "n/a"))

print("CROSS-VERIFY:", "PASS" if ok and stale_char in charac else "FAIL")
