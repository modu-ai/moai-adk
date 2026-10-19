"""measure-caps.py — read-only byte measurement for SPEC-LANE-READ-VERBS-001 (plan § C).

Usage:
    python3 -I .moai/specs/SPEC-LANE-READ-VERBS-001/measure-caps.py <judging-binary>

The script runs only read verbs (`todo list --json`, `factory runs`,
`factory status --json`, `session list --json`) and writes nothing. It first checks
its encoder emulation against each live output, byte for byte, and then prints the
byte counts of the candidate row caps: live data, worst case at the declared field
widths, and the options that were not chosen.
"""

import json
import os
import subprocess
import sys
from collections import Counter

BIN = sys.argv[1]
BOUND = 2048
WIDE = 9999999999        # 10-digit integers (declared width)
WIDE6 = 999999           # 6-digit counts (declared width)
WORST_ID = "t" + "9" * 7  # 8-byte identifier (declared width)
FIELDS = ["run_id", "card_id", "state", "stage", "lease_holder", "lease_expired"]


def run(args):
    proc = subprocess.run([BIN] + args, capture_output=True)
    return proc.returncode, proc.stdout


def esc(text):
    # Go encoding/json (HTML-safe) escapes <, >, & and U+2028 / U+2029.
    text = text.replace("<", "\\u003c").replace(">", "\\u003e").replace("&", "\\u0026")
    return text.replace("\u2028", "\\u2028").replace("\u2029", "\\u2029")


def compact_bytes(obj):
    # json.Marshal followed by a newline: the todo default and the --compact form.
    return (esc(json.dumps(obj, separators=(",", ":"), ensure_ascii=False)) + "\n").encode("utf-8")


def indent_bytes(obj):
    # json.Encoder or json.MarshalIndent with a two-space indent, then a newline.
    return (esc(json.dumps(obj, indent=2, ensure_ascii=False)) + "\n").encode("utf-8")


def prefix_of(text):
    # todoTextPrefix: todoPRCell (tab, newline, carriage return become spaces),
    # then the first 40 runes followed by "..." when the text is longer.
    text = text.replace("\t", " ").replace("\n", " ").replace("\r", " ")
    runes = list(text)
    return text if len(runes) <= 40 else "".join(runes[:40]) + "..."


def row(label, size, bounded=True):
    if not bounded:
        verdict = "waived"
    else:
        verdict = "PASS" if size <= BOUND else "FAIL"
    print(f"  {label:<46} {size:>6} bytes  {verdict}")


def probe(pid):
    # Signal 0 only: no signal is delivered. Only ESRCH means positively gone.
    if pid <= 0:
        return "nonpositive"
    try:
        os.kill(pid, 0)
        return "ok"
    except ProcessLookupError:
        return "esrch"
    except PermissionError:
        return "eperm"
    except Exception:
        return "other"


# ---- live outputs and validation ---------------------------------------------
rc_t, raw_t = run(["todo", "list", "--json"])
todo = json.loads(raw_t)
rc_f, raw_f = run(["factory", "status", "--json"])
fac = json.loads(raw_f)
rc_s, raw_s = run(["session", "list", "--json"])
sess = json.loads(raw_s)
rc_r, raw_r = run(["factory", "runs"])

print("== validation: encoder emulation vs live output (byte-exact)")
print("  todo default    rc", rc_t, "bytes", len(raw_t), "exact", compact_bytes(todo) == raw_t)
print("  factory status  rc", rc_f, "bytes", len(raw_f), "exact", indent_bytes(fac) == raw_f)
print("  session list    rc", rc_s, "bytes", len(raw_s), "exact", indent_bytes(sess) == raw_s)

# ---- todo --compact ----------------------------------------------------------
live = [i for i in todo["items"] if i.get("state") != "dropped"]
dropped_n = len(todo["items"]) - len(live)


def todo_env(rows, live_n, last_seq, version, findings, archived, dropped):
    return {
        "project_uuid": todo["project_uuid"],
        "version": version,
        "last_seq": last_seq,
        "live": live_n,
        "withheld": live_n - len(rows),
        "items": rows,
        "findings_count": findings,
        "archived_count": archived,
        "dropped_count": dropped,
    }


print(f"== todo --compact, live queue ({len(live)} live items)")
for cap in [6, 7, 8, 10, 15, 20]:
    rows = [{"id": i["id"], "state": i["state"], "prefix": prefix_of(i["text"])} for i in live[:cap]]
    env = todo_env(rows, len(live), todo["last_seq"], todo["version"],
                   len(todo["findings"]), len(todo["archived"]), dropped_n)
    row(f"cap {cap}", len(compact_bytes(env)))

wrow = {"id": WORST_ID, "state": "picked", "prefix": "<" * 40 + "..."}
print("== todo --compact, worst case (declared widths; each prefix rune at 6 bytes)")
for cap in [5, 6, 7, 8, 10]:
    row(f"cap {cap}", len(compact_bytes(todo_env([wrow] * cap, WIDE, WIDE, 1, WIDE, WIDE, WIDE))))

crow = {"id": WORST_ID, "state": "picked", "prefix": "\u4e2d" * 40 + "..."}
print("== todo --compact, CJK-only prefixes (3 bytes per rune, no escapes)")
for cap in [10, 11]:
    row(f"cap {cap}", len(compact_bytes(todo_env([crow] * cap, WIDE, WIDE, 1, WIDE, WIDE, WIDE))))

arow = {"id": WORST_ID, "state": "picked", "prefix": "<" * 6 + '"' * 2 + "..."}
print("== todo --compact, option not chosen: prefix of at most 40 encoded bytes plus '...'")
for cap in [20, 21]:
    row(f"cap {cap}", len(compact_bytes(todo_env([arow] * cap, WIDE, WIDE, 1, WIDE, WIDE, WIDE))))

# ---- factory status ----------------------------------------------------------
active = [line.split()[0] for line in raw_r.decode("utf-8").splitlines()[1:]
          if len(line.split()) > 1 and line.split()[1] == "active"]
print(f"== factory runs: active run ids {active} (rc {rc_r})")

QUOTA = fac["quota"]


def fac_env(rows, withheld, run_value, unavail, skipped, durations):
    env = {"run": run_value, "cards": rows, "withheld": withheld, "unavailable_count": unavail}
    if skipped is not None:
        env["unavailable_skipped"] = skipped
    env["lease_duration"], env["renew_interval"] = durations
    env["quota"] = QUOTA
    return env


WROW = {"run_id": "x" * 6, "card_id": WORST_ID, "state": "merged-local",
        "stage": "merge-ready", "lease_holder": "h" * 16, "lease_expired": False}
WORST_DUR = ("9" * 12, "9" * 12)
print("== factory default, worst case (declared widths; quota block as live)")
for cap in [6, 7, 8, 10]:
    row(f"cap {cap}", len(indent_bytes(fac_env([WROW] * cap, WIDE6, "x" * 6, WIDE6, WIDE6, WORST_DUR))))

if len(active) == 1:
    act = [c for c in fac["cards"] if c["run_id"] == active[0]]
    print(f"== factory default, live active run {active[0]} ({len(act)} cards)")
    for cap in [6, 7, 8]:
        rows = [{k: c[k] for k in FIELDS} for c in act[:cap]]
        env = fac_env(rows, len(act) - cap, "", len(fac["unavailable"]), None,
                      (fac["lease_duration"], fac["renew_interval"]))
        row(f"cap {cap}", len(indent_bytes(env)))

print("== factory options not chosen: B = run_id dropped from rows; C = compact encoding")
for cap in [7, 8]:
    rows = [{k: WROW[k] for k in FIELDS[1:]}] * cap
    row(f"B cap {cap} (indent)", len(indent_bytes(fac_env(rows, WIDE6, "x" * 6, WIDE6, WIDE6, WORST_DUR))))
for cap in [8, 9, 10, 11]:
    row(f"C cap {cap} (compact)", len(compact_bytes(fac_env([WROW] * cap, WIDE6, "x" * 6, WIDE6, WIDE6, WORST_DUR))))

# ---- session list ------------------------------------------------------------
states = Counter(probe(e["pid"]) for e in sess)
live_rows = [e for e in sess if probe(e["pid"]) != "esrch"]
print(f"== session list, live store: {len(sess)} entries {dict(states)}")
row("session --all (every entry)", len(indent_bytes(sess)), bounded=False)
row("session default under S-2 (live entries)", len(indent_bytes(live_rows)), bounded=False)
