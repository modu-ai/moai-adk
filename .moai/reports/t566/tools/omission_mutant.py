# t566 omission-removal mutant on the real codemaps documents.
#   apply:  pick the first omission unit whose matching lines name no other
#           judged unit, back up every document it touches, delete those lines.
#   revert: restore every backed-up document and prove the bytes match.
#
# Usage: python3 omission_mutant.py apply|revert
import glob
import hashlib
import json
import os
import shutil
import sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "..", ".."))
CODEMAPS = os.path.join(ROOT, ".moai", "project", "codemaps")
JUDG = os.path.join(CODEMAPS, "fold-judgments.txt")
BACKUP_DIR = os.path.join(os.path.dirname(__file__), "omission-backup")
STATE = os.path.join(BACKUP_DIR, "state.json")


def sha(path):
    return hashlib.sha256(open(path, "rb").read()).hexdigest()


def judgments():
    units = []
    for line in open(JUDG, encoding="utf-8"):
        parts = line.split()
        if len(parts) == 2 and parts[0] in ("fold", "omission"):
            units.append((parts[0], parts[1]))
    return units


def apply():
    if os.path.exists(BACKUP_DIR):
        raise SystemExit("backup already exists; revert first")
    units = judgments()
    all_units = [u for _, u in units]
    docs = sorted(glob.glob(os.path.join(CODEMAPS, "*.md")))
    for kind, unit in units:
        if kind != "omission":
            continue
        touched = {}
        clean = True
        for d in docs:
            lines = open(d, encoding="utf-8").read().split("\n")
            hits = [l for l in lines if unit in l]
            for l in hits:
                if any(o != unit and o in l for o in all_units):
                    clean = False
            if hits:
                touched[d] = len(hits)
        if clean and touched:
            break
    else:
        raise SystemExit("no omission unit with isolated lines")
    os.makedirs(BACKUP_DIR)
    record = {"unit": unit, "files": {}}
    for d, n in touched.items():
        name = os.path.basename(d)
        shutil.copyfile(d, os.path.join(BACKUP_DIR, name))
        before = sha(d)
        data = open(d, encoding="utf-8").read()
        kept = "\n".join(l for l in data.split("\n") if unit not in l)
        with open(d, "w", encoding="utf-8") as f:
            f.write(kept)
        record["files"][name] = {"removed_lines": n, "original": before, "mutated": sha(d)}
    with open(STATE, "w", encoding="utf-8") as f:
        json.dump(record, f, indent=2)
    print(json.dumps(record, indent=2))


def revert():
    if not os.path.exists(STATE):
        raise SystemExit("no backup to restore from")
    record = json.load(open(STATE, encoding="utf-8"))
    ok = True
    for name, info in record["files"].items():
        target = os.path.join(CODEMAPS, name)
        mutated = sha(target)
        shutil.copyfile(os.path.join(BACKUP_DIR, name), target)
        restored = sha(target)
        match = restored == info["original"]
        ok = ok and match
        print(name, "mutated", mutated, "original", info["original"], "restored", restored, "match", match)
    shutil.rmtree(BACKUP_DIR)
    print("unit", record["unit"])
    print("RESTORED_MATCHES_ORIGINAL", ok)


{"apply": apply, "revert": revert}[sys.argv[1]]()
