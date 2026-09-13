"""Print the five probe cells side by side: pre-fix (HEAD) vs post-fix."""
import json
import pathlib

here = pathlib.Path(__file__).parent
base = json.load(open(here / "probe-baseline.json"))
after = json.load(open(here / "probe-after.json"))

rows = [
    ("P0 control: 2nd call re-ran checks",
     lambda d: d["P0_control_pass"]["second"]["checks_ran"], False),
    ("P1 fail->recall: 2nd call re-ran checks",
     lambda d: d["P1_fail_recall"]["second"]["checks_ran"], False),
    ("P1 fail->recall: stdout byte-identical",
     lambda d: d["P1_fail_recall"]["stdout_identical_1_2"], True),
    ("P2a interrupted fresh: re-ran checks",
     lambda d: d["P2a_interrupted_fresh"]["checks_ran"], False),
    ("P2b interrupted stale: re-ran checks",
     lambda d: d["P2b_interrupted_stale"]["checks_ran"], True),
    ("P3 repaired tree: re-ran checks",
     lambda d: d["P3_fixed_same_head"]["after_fix"]["checks_ran"], True),
    ("P3 repaired tree: stale block re-delivered",
     lambda d: d["P3_fixed_same_head"]["stale_block_redelivered"], False),
    ("P3 repaired tree: HEAD unchanged",
     lambda d: d["P3_fixed_same_head"]["head_unchanged"], True),
]

print("%-46s %-9s %-9s %s" % ("cell", "pre-fix", "post-fix", "want(post)"))
for label, get, want in rows:
    b, a = get(base), get(after)
    mark = "OK" if a == want else "MISMATCH"
    print("%-46s %-9s %-9s %s  %s" % (label, b, a, want, mark))

print()
print("P3 post-fix stdout: %r" % after["P3_fixed_same_head"]["after_fix"]["stdout"])
print("P3 post-fix record: %s" % after["P3_fixed_same_head"]["after_fix"]["record"])
print("P1 post-fix record: %s" % after["P1_fail_recall"]["second"]["record"])
print("P1 pre-fix  record: %s" % base["P1_fail_recall"]["second"]["record"])
