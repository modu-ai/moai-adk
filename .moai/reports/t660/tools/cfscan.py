# t660 invisible-character scan: counts Unicode Cf (format) characters per file.
# The control string holds exactly one Cf character and must print 1.
import sys
import unicodedata

ctl = "a" + chr(0x200B) + "b"
print("control", sum(1 for c in ctl if unicodedata.category(c) == "Cf"))
for p in sys.argv[1:]:
    t = open(p, encoding="utf-8").read()
    print(p, "chars", len(t), "Cf", sum(1 for c in t if unicodedata.category(c) == "Cf"))
