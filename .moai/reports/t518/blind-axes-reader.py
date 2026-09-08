"""Corpus reader for SPEC-SPEC-LINT-BLIND-AXES-001 §A / §I figures.

Read-only. Run from the repository root (or worktree root):
    python3 .moai/reports/t518/blind-axes-reader.py
"""
import glob
import os
import re

# Collector regex — verbatim from internal/spec/lint_req_widen.go reqLineWidePattern
WIDE = re.compile(r'^\s*[-*]\s+\**\s*(REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+)\s*\**\s*(?:\([^)]*\)\s*\**\s*)?(?:—|:)\s*(.*)$')
# Table-row regex — spec.md §G
ROW = re.compile(r'^\s*\|\s*\**\s*(REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+)\s*\**\s*\|')
# A cell holding exactly one bare ID token
IDTOK = re.compile(r'^(?:REQ|AC|SPEC)-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+[a-z]?$')

# L1 — narrow modality lexicon (candidate C-d). Substring match, case-sensitive.
L1 = ["SHALL", "해야 한다", "해서는 안 된다"]
# L2 — widened modality lexicon (candidate C-e). L1 + Korean obligation endings
#      + GEARS pattern names. Substring match, case-sensitive.
L2 = L1 + ["않아야 한다", "되어야 한다", "여야 한다",
           "말아야 한다", "않는다", "이어야 한다",
           "MUST", "SHOULD", "Ubiquitous", "Event-driven", "State-driven",
           "Unwanted", "Optional", "When ", "While ", "Where "]


def cells(line):
    s = line.strip()
    if s.startswith('|'):
        s = s[1:]
    if s.endswith('|'):
        s = s[:-1]
    return [c.strip().strip('*').strip() for c in s.split('|')]


def all_single_id(cs):
    return all(IDTOK.match(c) for c in cs if c != "")


def all_id_list(cs):
    for c in cs:
        if c == "":
            continue
        parts = [p.strip() for p in c.split(',')]
        if not all(IDTOK.match(p) for p in parts if p):
            return False
    return True


def has(line, lex):
    return any(t in line for t in lex)


dirs = sorted(glob.glob('.moai/specs/SPEC-*'))
specs = [d for d in dirs if os.path.isfile(os.path.join(d, 'spec.md'))]
wide_total = 0
blind = []
blindrows = 0
per = {}
for d in specs:
    body = open(os.path.join(d, 'spec.md'), encoding='utf-8').read().split('\n')
    w = [ln for ln in body if WIDE.match(ln)]
    rows = [ln for ln in body if ROW.match(ln)]
    wide_total += len(w)
    if len(w) == 0 and len(rows) > 0:
        blind.append(d)
        blindrows += len(rows)
        per[os.path.basename(d)] = len(rows)

print("dirs %d   spec.md %d" % (len(dirs), len(specs)))
print("definition lines(wide): %d" % wide_total)
print("blind specs: %d   rows in blind: %d" % (len(blind), blindrows))
top = sorted(per.items(), key=lambda kv: -kv[1])[:6]
print("top: " + " | ".join("%s %d" % (k, v) for k, v in top))

surv = {"C-a": 0, "C-b": 0, "C-d": 0, "C-e": 0}
binlag = {"C-a": 0, "C-b": 0, "C-d": 0, "C-e": 0}
for d in blind:
    body = open(os.path.join(d, 'spec.md'), encoding='utf-8').read().split('\n')
    isb = 'BINLAG-INVOCATION-001' in d
    for ln in body:
        if not ROW.match(ln):
            continue
        cs = cells(ln)
        rej = {
            "C-a": all_single_id(cs),
            "C-b": all_id_list(cs),
            "C-d": not has(ln, L1),
            "C-e": all_single_id(cs) or (not has(ln, L2)),
        }
        for k, v in rej.items():
            if not v:
                surv[k] += 1
            if isb and v:
                binlag[k] += 1
print("survivors: %s" % surv)
print("BINLAG rejected out of 16: %s" % binlag)
