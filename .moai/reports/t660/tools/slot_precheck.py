# t660 slot pre-check: records load and the go test processes already running
# before an internal/cli compile. Prints only matching lines, never the raw
# machine-wide snapshot. Usage: python3 slot_precheck.py <label>
import datetime
import subprocess
import sys

label = sys.argv[1] if len(sys.argv) > 1 else "run"
print("slot pre-check before the %s run" % label)
print(datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"))
print(subprocess.run(["uptime"], capture_output=True, text=True).stdout.rstrip())
snap = subprocess.run(["ps", "-axo", "pid,etime,comm,args"], capture_output=True, text=True).stdout
lines = snap.splitlines()
print(lines[0])
print("(snapshot lines: %d)" % len(lines))
# Observer control: the ps process must see itself.
print("observer control - the ps process itself:",
      sum(1 for ln in lines[1:] if ln.split(None, 3)[2:3] == ["ps"]))
hits = []
for ln in lines[1:]:
    parts = ln.split(None, 3)
    if len(parts) < 4:
        continue
    comm, args = parts[2], parts[3]
    if "zsh" in comm or "python" in comm:
        continue
    if args.startswith("go test") or ".test -test." in args:
        hits.append(ln)
print("go test/.test lines (zsh/python excluded): %d" % len(hits))
for ln in hits:
    print(ln[:160])
