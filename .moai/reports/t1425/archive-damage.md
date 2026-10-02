# t1425 — damage check of the live retention archives (read-only, lane measurement)

The card text asked: "확인 필요: 2026-09.jsonl.gz 중복 이벤트 여부". The lane measured the live archives read-only on 2026-10-02 (source files untouched; copies and scripts live in machine-local scratch). Archive directory: `/Users/goos/MoAI/moai-adk-go/.moai/harness/learning-history/archive/` (primary checkout, not part of any tree this card changes).

## Observed

1. Standard decompression fails on three archives (each command run separately, exit code as printed):
   - `gzip -t .../2026-09.jsonl.gz` → `gzip: data stream error` / `uncompress failed`, exit 1 (file 11,764,350 bytes, mtime Oct 2 15:24)
   - `gzip -t .../2026-08.jsonl.gz` → same error, exit 1 (14,398,585 bytes, mtime Oct 1 07:43)
   - `gzip -t .../2026-07.jsonl.gz` → same error, exit 1 (4,179,609 bytes, mtime Aug 31 09:15)
   Not tested: `2026-04`, `2026-05`, `2026-06` (149 B, 12,995 B, 1,371 B).
2. Positive control, so the failure is not a tool limit: two valid gzip members (`printf 'a\n' | gzip -c`, `printf 'b\n' | gzip -c`) concatenated with `cat` → `gzip -t` exit 0 and `gzip -dc` printed `a` and `b`. The same `gzip` therefore reads concatenated members, which is the format `appendToGzip` produces.
3. Member walk of `2026-09.jsonl.gz` from byte 0: `zlib error in member 1 at offset 0: Error -3 while decompressing data: invalid block type`, 0 bytes recovered by a sequential read.
4. Salvage estimate (try to decode a gzip member at every gzip header magic `1f 8b 08` offset, skip offsets inside an already-decoded member; the script is below). Identical lines are counted as duplicates; timestamps carry sub-second precision for real events, which makes two genuinely distinct identical lines unlikely, but this was not verified line by line:

| archive | magic offsets | members decoded | undecodable candidates | recovered lines | unique lines | duplicate lines |
|---|---|---|---|---|---|---|
| 2026-09 | 7,523 | 7,348 | 175 | 775,051 | 37,271 | 737,780 (95.2 %) |
| 2026-08 | 53,794 | 53,382 | 409 | 297,722 | 176,467 | 121,255 (40.7 %) |
| 2026-07 | 16,565 | 16,439 | 126 | 107,392 | 94,102 | 13,290 (12.4 %) |

"Undecodable candidates" mixes truly corrupt members with magic bytes that occur by chance inside compressed data; the two cannot be told apart from this scan.

## What it means, and what it does not

- The suspicion in the card is confirmed for 2026-09: about 95 percent of the recoverable lines are repeats, consistent with N concurrent pruners each appending the same stale events (the observed RED of this card shows the same mechanism: interleaved appends corrupt the gzip stream and one event is archived once per racing process).
- The corruption of 2026-08 and 2026-07 shows the concurrent-append defect also fired before 2026-10-02 (2026-07 was last written Aug 31), on a smaller scale.
- NOT established: how many unique events are lost versus merely unreachable by a sequential reader (salvage recovered members, but the corrupt members' content is gone); whether any consumer outside Go reads these archives. A grep of `internal` and `cmd` for `learning-history/archive`, `archiveDir` and `.jsonl.gz` found only the writer (`internal/harness/retention.go`) and the hook code that builds the path (`internal/cli/hook.go` lines 820, 918, 1055, 1219); no Go reader.
- The repair of this card only stops new corruption (one pruner per interval, stamp before the work). It does NOT repair existing archives. Salvage or replacement of the three archives is a leader decision (operator-visible data); the lane touched nothing.

## Scripts (run as `python3 <script> <archive>`, read-only)

```python
# scan_archive.py — sequential member walk, stops at the first bad member
import sys, zlib
path = sys.argv[1]; data = open(path, "rb").read()
pos = members = good_bytes = 0; lines = []; err = None
while pos < len(data):
    if data[pos:pos + 3] != b"\x1f\x8b\x08":
        err = f"bad member header at offset {pos}"; break
    d = zlib.decompressobj(31)
    try:
        chunk = d.decompress(data[pos:])
    except zlib.error as e:
        err = f"zlib error in member {members + 1} at offset {pos}: {e}"; break
    good_bytes += len(chunk); lines.extend(chunk.splitlines()); members += 1
    used = len(data[pos:]) - len(d.unused_data)
    if not d.eof:
        err = f"member {members} truncated at offset {pos}"; break
    pos += used
print(members, good_bytes, len(lines), len(set(lines)), err)
```

```python
# salvage_scan.py — decode a member at every gzip magic offset
import sys, zlib
MAGIC = b"\x1f\x8b\x08"; data = open(sys.argv[1], "rb").read(); mv = memoryview(data)
offsets = []; i = data.find(MAGIC)
while i != -1:
    offsets.append(i); i = data.find(MAGIC, i + 1)
ok = bad = 0; lines = []; covered_to = -1
for off in offsets:
    if off < covered_to:
        continue
    d = zlib.decompressobj(31)
    try:
        chunk = d.decompress(mv[off:])
    except zlib.error:
        bad += 1; continue
    if not d.eof:
        bad += 1; continue
    ok += 1; covered_to = off + (len(data) - off - len(d.unused_data)); lines.extend(chunk.splitlines())
u = len(set(lines))
print(len(offsets), ok, bad, len(lines), u, len(lines) - u)
```
