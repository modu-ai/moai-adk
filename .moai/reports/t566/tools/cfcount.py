# t566: count Unicode category Cf characters in the given files (and stdin text
# when "-" is passed), with a positive control that must report 1.
import sys
import unicodedata


def cf(text):
    return sum(1 for ch in text if unicodedata.category(ch) == "Cf")


total = 0
for p in sys.argv[1:]:
    text = sys.stdin.read() if p == "-" else open(p, encoding="utf-8").read()
    n = cf(text)
    total += n
    print(p, "Cf=%d" % n)
print("control Cf=%d (expected 1)" % cf("a" + chr(0x200B) + "b"))
print("TOTAL Cf=%d" % total)
