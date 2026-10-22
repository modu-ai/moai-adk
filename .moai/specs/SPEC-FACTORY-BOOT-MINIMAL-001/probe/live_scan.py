"""SPEC-FACTORY-BOOT-MINIMAL-001 live-session scan (card t1923, AC-FBM-008).

Read-only. Walks a Claude Code projects directory, finds factory role sessions
(SessionStart hook stdout carrying "Role-gated rules (factory") started at or
after a given ISO timestamp, and prints one row per session:

    <timestamp> <role> <composite_utf16> <pre_role_utf16> <role_block_utf16> <overflow:bool> <file>

followed by "count N" and "overflow N". Exit 0 always; the verdict is read
from the printed rows (composite <= 10000 and overflow 0).

Usage: python3 -I live_scan.py <projects-dir> <since-iso>
"""
import json
import os
import sys
import time


def u16(s):
    return len(s.encode("utf-16-le")) // 2


def main():
    base, since_iso = sys.argv[1], sys.argv[2]
    rows = []
    for dp, _dn, fn in os.walk(base):
        if "/subagents" in dp:
            continue
        for f in fn:
            if not f.endswith(".jsonl"):
                continue
            p = os.path.join(dp, f)
            try:
                if os.path.getmtime(p) < time.time() - 30 * 86400:
                    continue
                with open(p, encoding="utf-8", errors="replace") as h:
                    for i, line in enumerate(h):
                        if i > 15:
                            break
                        if "Role-gated rules (factory" not in line:
                            continue
                        rec = json.loads(line)
                        ts = rec.get("timestamp", "")
                        if ts < since_iso:
                            break
                        so = (rec.get("attachment") or {}).get("stdout", "")
                        hj = json.loads(so)
                        ac = (hj.get("hookSpecificOutput") or {}).get("additionalContext", "")
                        k = ac.find("Role-gated rules (factory")
                        role = "lane" if "factory factory-lane" in ac else "leader"
                        over = ac.startswith("NOTE: the output above exceeds")
                        rows.append((ts, role, u16(ac), u16(ac[:k]), u16(ac[k:]), over, p.split("projects/")[-1][:70]))
                        break
            except Exception:
                continue
    rows.sort()
    for r in rows:
        print(*r)
    print("count", len(rows))
    print("overflow", sum(1 for r in rows if r[5]))


if __name__ == "__main__":
    main()
