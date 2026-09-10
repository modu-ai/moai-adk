# t566: insert the fold/omission judgment text into the codemaps workflow
# (template copy first), then write the identical bytes to the local copy.
import hashlib
import os

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "..", ".."))
TEMPLATE = os.path.join(ROOT, "internal", "template", "templates", ".claude", "skills", "moai", "workflows", "codemaps.md")
LOCAL = os.path.join(ROOT, ".claude", "skills", "moai", "workflows", "codemaps.md")

FENCE = "`" * 3

PHASE3_ANCHOR = "If --format json: Generate machine-readable JSON alongside markdown.\n"
PHASE3_ADD = """
Fold and omission judgments: when a refresh classifies units as `fold` (folded into a parent's description, so the unit itself must stay unnamed in the documents) or `omission` (left out by an earlier refresh and now owed a description), record every judgment in `.moai/project/codemaps/fold-judgments.txt`, one judgment per line:

{fence}text
# blank lines and lines starting with # are ignored
fold <unit>
omission <unit>
{fence}

A unit is the exact string a reader would search for: a package or module directory path, or a single file path. The file is the input of the Phase 4 fold and omission judgment check. It is a `.txt` file on purpose: every `*.md` file in the directory is scanned for hits, so a judgments file inside that set would match every unit it names. The map generation step writes only the named documents above and never rewrites this file.
""".format(fence=FENCE)

PHASE4_ANCHOR = "- Compare with existing .moai/project/codemaps/ to highlight changes (if not --force)\n"
PHASE4_ADD = """
### Fold and Omission Judgment Check

Runnable check for the judgments recorded in Phase 3, executed from the project root. Each recorded unit is checked on its own. A census of packages with zero hits cannot stand in for this check: a file-granularity unit never appears in such a census, so a fold unit given prose there goes unnoticed.

{fence}bash
# Check recorded fold / omission judgments unit by unit; pass = no UNCOVERED, FOLD-PROSE, or GAP line
dir=".moai/project/codemaps"
judg="$dir/fold-judgments.txt"
if [ ! -r "$judg" ]; then
  echo "GAP: $judg is not readable — fold and omission judgments were not observed"
else
  set -- "$dir"/*.md
  if [ ! -e "$1" ]; then
    echo "GAP: no generated *.md document under $dir — hits were not observed"
  else
    awk '
      {{ sub(/\\r$/, "") }}
      FILENAME == ARGV[1] {{
        line = $0
        sub(/^[ \\t]+/, "", line)
        sub(/[ \\t]+$/, "", line)
        if (line == "" || substr(line, 1, 1) == "#") next
        if (line ~ /^(fold|omission)[ \\t]+[^ \\t]+$/) {{
          n++
          kind[n] = line
          sub(/[ \\t].*$/, "", kind[n])
          unit[n] = line
          sub(/^[^ \\t]+[ \\t]+/, "", unit[n])
          if (kind[n] == "fold") {{ folds++ }} else {{ omissions++ }}
        }} else {{
          bad[++nbad] = "GAP: malformed judgment line " FNR ": " line
        }}
        next
      }}
      {{ for (i = 1; i <= n; i++) if (index($0, unit[i]) > 0) hits[i]++ }}
      END {{
        printf "COLLECTED: fold=%d omission=%d\\n", folds, omissions
        for (i = 1; i <= nbad; i++) print bad[i]
        if (n == 0) {{ print "GAP: 0 judgments collected — fold and omission units were not observed"; exit }}
        for (i = 1; i <= n; i++) {{
          if (kind[i] == "omission" && hits[i] == 0) print "UNCOVERED: " unit[i]
          if (kind[i] == "fold" && hits[i] > 0) print "FOLD-PROSE: " unit[i] " " hits[i]
        }}
      }}
    ' "$judg" "$@"
  fi
fi
{fence}

The script narrows where to look; the orchestrator decides. It never prints PASS and always exits 0. The check passes when the output carries no `UNCOVERED:`, `FOLD-PROSE:`, or `GAP:` line.

- `COLLECTED: fold=N omission=M` — the judgments read from `fold-judgments.txt`. Record it with the verification result: it is the measurement the check rests on, and a count that differs from the number of judgments the refresh made means a judgment was never written down.
- `UNCOVERED: <unit>` — an omission unit that no line of the generated documents names. The refresh owed it a description and did not write one.
- `FOLD-PROSE: <unit> <lines>` — a fold unit named on that many document lines. A folded unit gets no prose of its own: remove those lines, or revise the judgment in `fold-judgments.txt` if the unit should be described after all.
- `GAP: …` — the check was not observed: the judgments file is unreadable, it holds zero judgments, a line is neither `fold <unit>` nor `omission <unit>`, or no generated document exists. Report it as a gap, never read it as a pass. A refresh that made no fold or omission judgment also reports this GAP; state that reason rather than creating an empty file.

A hit is a fixed-string line match (the `grep -c -F` reading) over the top-level generated `*.md` documents in `.moai/project/codemaps/` only; area-specific subdirectories and the judgments file are outside the scanned set. A unit matches whether it appears as a whole path or inside a longer one, so a fold unit that is a prefix of a described unit reads as prose: revise the judgment or the wording, not the check.
""".format(fence=FENCE)


def sha(b):
    return hashlib.sha256(b).hexdigest()


text = open(TEMPLATE, encoding="utf-8").read()
before = sha(text.encode("utf-8"))
for anchor in (PHASE3_ANCHOR, PHASE4_ANCHOR):
    if text.count(anchor) != 1:
        raise SystemExit("anchor count != 1: %r" % anchor)
if "### Fold and Omission Judgment Check" in text:
    raise SystemExit("already patched")
text = text.replace(PHASE3_ANCHOR, PHASE3_ANCHOR + PHASE3_ADD)
text = text.replace(PHASE4_ANCHOR, PHASE4_ANCHOR + PHASE4_ADD)
data = text.encode("utf-8")
with open(TEMPLATE, "wb") as f:
    f.write(data)
with open(LOCAL, "wb") as f:
    f.write(data)
print("template before", before)
print("template after ", sha(open(TEMPLATE, "rb").read()))
print("local after    ", sha(open(LOCAL, "rb").read()))
print("IDENTICAL", open(TEMPLATE, "rb").read() == open(LOCAL, "rb").read())
