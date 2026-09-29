"""Temporary debug harness for selftest failure diagnosis (not committed)."""
import sys
from collections import Counter

sys.path.insert(0, "scripts")
import jamo_integrity as J

vocab = Counter({
    "채널": 40, "바이너리": 30, "순환": 20, "엉터리": 10,
    "브랜치": 25, "나머지": 15, "엔트리": 12, "불변": 8,
})
cases = {
    "채널": J.corrupt_word("채널", 2, J.JONG_CHARS.index("ㅈ")),
    "순환": J.corrupt_word("순환", 2, J.JONG_CHARS.index("ㅋ")),
    "브랜치": J.corrupt_word("브랜치", 1, J.JUNG_CHARS.index("ㅑ")),
    "나머지": J.corrupt_word("나머지", 2, J.JONG_CHARS.index("ㅆ")),
}
print("corrupt forms:", list(cases.values()))
findings = J.scan_text("\n".join(cases.values()), "selftest", vocab)
for f in findings:
    print("flagged:", repr(f.word), "line", f.line, "col", f.col, f.candidates[:3])
for name, w in cases.items():
    print("distance %s -> %s:" % (w, name), J.slot_distance(w, name))
