import json, sys

path = sys.argv[1]
out = sys.argv[2]

def u16(s):
    return len(s.encode("utf-16-le")) // 2

with open(path, encoding="utf-8") as f, open(out, "w", encoding="utf-8") as w:
    for ln, line in enumerate(f):
        if ln > 12:
            break
        try:
            rec = json.loads(line)
        except Exception:
            continue
        att = rec.get("attachment") or {}
        so = att.get("stdout")
        if not so or "hookSpecificOutput" not in so:
            continue
        try:
            hj = json.loads(so)
        except Exception as e:
            w.write(f"LINE {ln} stdout not json: {e}\n")
            continue
        ac = (hj.get("hookSpecificOutput") or {}).get("additionalContext", "")
        sm = hj.get("systemMessage", "")
        w.write(f"LINE {ln} command={att.get('command','')[:120]!r}\n")
        w.write(f"additionalContext UTF16={u16(ac)} systemMessage UTF16={u16(sm)}\n")
        # producer segmentation: split on blank line, report segment head and size
        segs = ac.split("\n\n")
        for i, s in enumerate(segs):
            head = s[:110].replace("\n", " | ")
            w.write(f"  seg{i:03d} len={u16(s):6d} head={head!r}\n")
        w.write("=====FULL=====\n" + ac + "\n=====END=====\n")
