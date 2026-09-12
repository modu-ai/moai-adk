# t659 clause census (read-only): for every zone-registry entry, count how often
# its clause occurs verbatim in the whole source file and whether the anchor's
# heading slug exists. Informs the source-file replacement rule. Also reports how
# the evolution-log loader's "---" split sees the real log.
#
# Usage: python3 clause_census.py
import os
import re

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "..", ".."))
REG = os.path.join(ROOT, ".claude", "rules", "moai", "core", "zone-registry.md")
LOG = os.path.join(ROOT, ".moai", "research", "evolution-log.md")
FENCE = "`" * 3


def slug(heading):
    s = heading.strip().lower()
    s = re.sub(r"[^\w\- ]", "", s)
    return "#" + s.replace(" ", "-")


text = open(REG, encoding="utf-8").read()
start = text.index(FENCE + "yaml") + len(FENCE + "yaml")
block = text[start:text.index(FENCE, start)]
entries = []
cur = None
for line in block.split("\n"):
    m = re.match(r"^- id: (\S+)", line)
    if m:
        cur = {"id": m.group(1)}
        entries.append(cur)
        continue
    m = re.match(r'^  (file|anchor|clause): "?(.*?)"?$', line)
    if m and cur is not None:
        cur[m.group(1)] = m.group(2)

buckets = {"missing_file": 0, "clause_0": 0, "clause_1": 0, "clause_2plus": 0}
anchor_found = 0
multi = []
zero = []
for e in entries:
    path = os.path.join(ROOT, e.get("file", ""))
    if not os.path.isfile(path):
        buckets["missing_file"] += 1
        continue
    src = open(path, encoding="utf-8").read()
    n = src.count(e.get("clause", ""))
    if n == 0:
        buckets["clause_0"] += 1
        zero.append(e["id"])
    elif n == 1:
        buckets["clause_1"] += 1
    else:
        buckets["clause_2plus"] += 1
        multi.append("%s x%d" % (e["id"], n))
    slugs = {slug(h) for h in re.findall(r"(?m)^#{1,6} (.+)$", src)}
    if e.get("anchor") in slugs:
        anchor_found += 1

print("registry_entries", len(entries))
print("buckets", buckets)
print("anchor_slug_found", anchor_found)
print("clause_2plus_ids", multi)
print("clause_0_ids", zero)


def norm(s):
    # Same shape as validator.normalizeWhitespace: collapse runs of whitespace.
    return " ".join(s.split())


by_id = {e["id"]: e for e in entries}
for rid in zero:
    e = by_id[rid]
    src = open(os.path.join(ROOT, e["file"]), encoding="utf-8").read()
    print("normalized_count", rid, norm(src).count(norm(e.get("clause", ""))), repr(e.get("clause", "")[:60]))

log = open(LOG, encoding="utf-8").read()
parts = log.split("---")
odd = [p for i, p in enumerate(parts) if i % 2 == 1]
print("evolution_log_dash_split_parts", len(parts))
print("odd_segments_with_top_level_id_key", sum(1 for p in odd if re.search(r"(?m)^id: ", p)))
print("odd_segment_heads", [p.strip()[:30] for p in odd])
