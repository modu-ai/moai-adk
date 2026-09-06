#!/usr/bin/env python3
"""t485 post-processor: classify watcher lsof captures and attribute holders.

Reads lsof-<epoch>-<tag>.txt files in the raw evidence dir, classifies each as
REAL (lock fd open, holder named), GONE (lock vanished between stat and lsof),
or EMPTY, then resolves each real holder PID (and its parent chain) against the
nearest co-temporal ps-<epoch>.txt snapshot. Emits summary.md + summary.jsonl.
"""
import glob
import json
import os
import re
import sys

RAW = sys.argv[1] if len(sys.argv) > 1 else "."
NAME_RE = re.compile(r"lsof-(\d+)-(.+)\.txt$")


def parse_ps_line(line):
    """Parse a `ps -Ao pid,ppid,lstart[,etime],command` row.

    Full snapshots carry etime between lstart and command (9 leading fields);
    filtered eps/fps snapshots do not (8). lstart itself is "Fri Sep  4
    18:19:43 2026". Return (pid, ppid, command) or None.
    """
    parts = line.rstrip("\n").split(None, 8)
    if len(parts) < 8 or not parts[0].isdigit():
        return None
    if len(parts) == 9:
        # full format: parts[7] is etime like "02-11:42:21" or "00:01"
        return parts[0], parts[1], parts[8]
    return parts[0], parts[1], parts[7]


def load_ps_index():
    """Map ps snapshot epochs to their parsed pid -> (ppid, command) tables."""
    tables = {}
    for path in glob.glob(os.path.join(RAW, "ps-*.txt")):
        m = re.search(r"ps-(\d+)\.txt$", path)
        if not m:
            continue
        table = {}
        with open(path, errors="replace") as fh:
            for line in fh:
                parsed = parse_ps_line(line)
                if parsed:
                    pid, ppid, cmd = parsed
                    table[int(pid)] = (ppid, cmd)
        tables[int(m.group(1))] = table
    return dict(sorted(tables.items()))


def nearest_ps(tables, epoch):
    best, bestd = None, None
    for ts, table in tables.items():
        d = abs(ts - epoch)
        if bestd is None or d < bestd:
            best, bestd = (ts, table), d
    return best if best and bestd <= 60 else (None, None)


def load_eps_table(epoch, tag):
    """Parse the concurrent eps/fps filtered snapshot taken at the same
    event instant (written by watcher v3/v4 while lsof was scanning)."""
    for prefix in ("eps", "fps"):
        path = os.path.join(RAW, f"{prefix}-{epoch}-{tag}.txt")
        if not os.path.exists(path):
            continue
        table = {}
        with open(path, errors="replace") as fh:
            for line in fh:
                parsed = parse_ps_line(line)
                if parsed:
                    table[int(parsed[0])] = (parsed[1], parsed[2])
        return table, prefix
    return None, None


def main():
    tables = load_ps_index()
    real, gone, empty, other = [], [], 0, 0
    for path in sorted(glob.glob(os.path.join(RAW, "lsof-*.txt"))):
        m = NAME_RE.search(os.path.basename(path))
        if not m:
            continue
        epoch, tag = int(m.group(1)), m.group(2)
        with open(path, errors="replace") as fh:
            content = fh.read()
        lines = content.splitlines()
        if not lines or not lines[0].strip():
            empty += 1
            continue
        if lines[0].startswith("COMMAND"):
            eps_table, eps_kind = load_eps_table(epoch, tag)
            for row in lines[1:]:
                cols = row.split()
                if len(cols) < 2 or not cols[1].isdigit():
                    continue
                pid = int(cols[1])
                # Prefer the co-temporal concurrent snapshot (eps) over the
                # nearest throttled full snapshot.
                if eps_table is not None and pid in eps_table:
                    ppid, cmd = eps_table[pid]
                    pp_cmd = (eps_table.get(int(ppid), ("", "(parent not in eps)"))[1]
                              if ppid.isdigit() else "?")
                    src, dist = f"concurrent-{eps_kind}", 0
                else:
                    (ps_epoch, table) = nearest_ps(tables, epoch)
                    ppid, cmd = (table.get(pid, ("?", "(not in snapshot)")) if table
                                 else ("?", "(no snapshot)"))
                    pp_cmd = table.get(int(ppid), ("", "(not in snapshot)"))[1] \
                        if table and ppid.isdigit() else "?"
                    src, dist = "throttled-ps", ((ps_epoch - epoch) if ps_epoch else None)
                real.append({
                    "epoch": epoch, "tree": tag, "holder_pid": pid,
                    "holder_cmd": cmd, "holder_ppid": ppid,
                    "ppid_cmd": pp_cmd, "ps_epoch": src,
                    "ps_distance_s": dist,
                })
        elif "No such file or directory" in content:
            gone.append({"epoch": epoch, "tree": tag})
        else:
            other += 1

    with open(os.path.join(RAW, "summary.jsonl"), "w") as fh:
        for row in real:
            fh.write(json.dumps(row) + "\n")

    # Aggregate by parent command.
    agg = {}
    for row in real:
        key = (row["holder_cmd"].split()[0] if row["holder_cmd"] else "?",
               row["ppid_cmd"][:80])
        agg[key] = agg.get(key, 0) + 1

    with open(os.path.join(RAW, "summary.md"), "w") as fh:
        fh.write("# t485 capture summary\n\n")
        fh.write(f"- REAL holder sightings: {len(real)}\n")
        fh.write(f"- GONE (lock vanished sub-second before lsof): {len(gone)}\n")
        fh.write(f"- EMPTY lsof output: {empty}\n")
        fh.write(f"- UNCLASSIFIED: {other}\n\n")
        fh.write("## Holder -> parent attribution (aggregate)\n\n")
        fh.write("| holder | parent (ppid_cmd) | count |\n|---|---|---|\n")
        for (holder, parent), n in sorted(agg.items(), key=lambda x: -x[1]):
            fh.write(f"| `{holder}` | `{parent}` | {n} |\n")
        fh.write("\n## Per-sighting detail\n\n")
        fh.write("| epoch UTC | tree | holder pid | holder cmd | ppid | parent cmd | ps dist |\n")
        fh.write("|---|---|---|---|---|---|---|\n")
        for row in real:
            ts = row["epoch"]
            fh.write(f"| {ts} | {row['tree']} | {row['holder_pid']} "
                     f"| `{row['holder_cmd'][:60]}` | {row['holder_ppid']} "
                     f"| `{row['ppid_cmd'][:60]}` | {row['ps_distance_s']} |\n")
    print(f"real={len(real)} gone={len(gone)} empty={empty} other={other}")


if __name__ == "__main__":
    main()
