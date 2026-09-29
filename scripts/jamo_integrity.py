#!/usr/bin/env python3
"""jamo_integrity.py — detect and repair Korean jamo-level corruption in text files.

Why this exists
---------------
The model backend output channel (observed on the GLM lane) occasionally emits
Korean words whose Hangul syllables carry one or more wrong jamo slots
(leading / medial / trailing consonant-vowel components), e.g. a composed
syllable rendered with a wrong or spurious jongseong, or a jungseong shifted
to a neighboring vowel. The corrupted syllables are still VALID precomposed
syllables (U+AC00-D7A3), so this is NOT an NFC/NFD normalization problem and
``unicodedata.normalize('NFC', ...)`` can neither detect nor repair it.

This tool:
  check    - scan text files (or --stdin) and report suspicious Korean words
             with line/column, ranked suggestions, and frequency evidence.
  restore  - propose replacements (--dry-run, default) or auto-replace
             (--apply) ONLY when the nearest vocabulary word is unique at
             jamo-slot distance 1; everything else stays flagged as ambiguous.

The detector is HEURISTIC. It compares out-of-vocabulary Hangul words against
a frequency-ranked vocabulary built at runtime from the repo's git-tracked
Korean prose. It can miss rare-word corruption and it can flag legitimate
rare words. Treat every finding as a candidate, not a verdict.

Stdlib only. Python 3.8+. No network.
"""

import argparse
import os
import re
import subprocess
import sys
import unicodedata
from collections import Counter

# ---------------------------------------------------------------------------
# Hangul arithmetic
# ---------------------------------------------------------------------------

SYL_MIN, SYL_MAX = 0xAC00, 0xD7A3
JAMO_BLOCK_MIN, JAMO_BLOCK_MAX = 0x1100, 0x11FF      # Hangul Jamo (NFD flavor)
COMPAT_JAMO_MIN, COMPAT_JAMO_MAX = 0x3130, 0x318F    # compatibility jamo

N_CHO, N_JUNG, N_JONG = 19, 21, 28

# Compact jamo name tables (index into Unicode jamo order).
CHO_CHARS = "ㄱㄲㄴㄷㄸㄹㅁㅂㅃㅅㅆㅇㅈㅉㅊㅋㅌㅍㅎ"
JUNG_CHARS = "ㅏㅐㅑㅒㅓㅔㅕㅖㅗㅘㅙㅚㅛㅜㅝㅞㅟㅠㅡㅢㅣ"
JONG_CHARS = ["", "ㄱ", "ㄲ", "ㄳ", "ㄴ", "ㄵ", "ㄶ", "ㄷ", "ㄹ", "ㄺ", "ㄻ",
              "ㄼ", "ㄽ", "ㄾ", "ㄿ", "ㅀ", "ㅁ", "ㅂ", "ㅄ", "ㅅ", "ㅆ",
              "ㅇ", "ㅈ", "ㅊ", "ㅋ", "ㅌ", "ㅍ", "ㅎ"]


def is_syllable(ch):
    return SYL_MIN <= ord(ch) <= SYL_MAX


def decompose(ch):
    """Precomposed syllable -> (cho, jung, jong) jamo indices; jong 0 = none."""
    i = ord(ch) - SYL_MIN
    cho, rem = divmod(i, N_JUNG * N_JONG)
    jung, jong = divmod(rem, N_JONG)
    return cho, jung, jong


def compose(cho, jung, jong):
    """(cho, jung, jong) indices -> precomposed syllable."""
    return chr(SYL_MIN + (cho * N_JUNG + jung) * N_JONG + jong)


def describe_syllable(ch):
    """Human-readable jamo breakdown, e.g. '채넍[넍=ㄴ+ㅓ+ㄵ]' pieces."""
    if not is_syllable(ch):
        return ch
    cho, jung, jong = decompose(ch)
    parts = CHO_CHARS[cho] + JUNG_CHARS[jung]
    if jong:
        parts += "+" + JONG_CHARS[jong]
    return "%s(%s)" % (ch, parts)


def slot_distance(a, b):
    """Jamo-slot distance between two same-length Hangul words.

    Each syllable is a (cho, jung, jong) tuple; jong 0 means 'no jongseong'.
    Distance = number of differing slots across aligned syllables.
    Words of different syllable counts are incomparable -> None.
    """
    if len(a) != len(b):
        return None
    d = 0
    for ca, cb in zip(a, b):
        ta = decompose(ca)
        tb = decompose(cb)
        d += sum(1 for x, y in zip(ta, tb) if x != y)
    return d


def slot_deltas(a, b):
    """List of human-readable slot deltas between two same-length Hangul words."""
    out = []
    for i, (ca, cb) in enumerate(zip(a, b)):
        ta, tb = decompose(ca), decompose(cb)
        names = ("choseong", "jungseong", "jongseong")
        for slot, (x, y) in enumerate(zip(ta, tb)):
            if x != y:
                fa = describe_slot(slot, x)
                fb = describe_slot(slot, y)
                out.append("syllable %d %s: %s -> %s" % (i + 1, names[slot], fa, fb))
    return out


def describe_slot(slot, idx):
    if slot == 0:
        return CHO_CHARS[idx]
    if slot == 1:
        return JUNG_CHARS[idx]
    return JONG_CHARS[idx] if idx else "(none)"


HANGUL_WORD_RE = re.compile(r"[가-힣]+")

# ---------------------------------------------------------------------------
# Vocabulary construction
# ---------------------------------------------------------------------------

PROSE_SUFFIXES = (".md", ".markdown", ".txt", ".yaml", ".yml")
PROSE_DIR_HINTS = (
    ".moai/docs/", ".moai/specs/", ".claude/rules/", ".moai/project/",
    "docs/",
)


def git_tracked_prose_files(repo_root, exclude=()):
    """Git-tracked Korean-prose candidate files (md/txt/yaml), newest-weighted."""
    try:
        out = subprocess.run(
            ["git", "ls-files", "-z"], cwd=repo_root,
            capture_output=True, check=True).stdout
    except (subprocess.CalledProcessError, FileNotFoundError):
        return []
    files = []
    exclude_abs = set(os.path.abspath(p) for p in exclude)
    for name in out.decode("utf-8", "replace").split("\0"):
        if not name:
            continue
        if not name.endswith(PROSE_SUFFIXES):
            continue
        path = os.path.join(repo_root, name)
        if os.path.abspath(path) in exclude_abs:
            continue
        files.append(path)
    # Prefer the known prose roots, then the rest, to bound build time.
    ranked = [f for f in files if any(h in f.replace(os.sep, "/") for h in PROSE_DIR_HINTS)]
    rest = [f for f in files if f not in set(ranked)]
    return ranked + rest


def strip_code_regions(text):
    """Yield lines with fenced code blocks emptied; inline code spans masked."""
    lines = text.split("\n")
    fence_re = re.compile(r"^\s*(```|~~~)")
    in_fence = False
    for line in lines:
        if fence_re.match(line):
            in_fence = not in_fence
            yield ""
            continue
        if in_fence:
            yield ""
            continue
        # Mask inline code spans (odd segments between backticks).
        parts = line.split("`")
        for i in range(1, len(parts), 2):
            parts[i] = ""
        yield "`".join(parts)


def build_vocab(repo_root, exclude=(), min_len=1):
    """Counter of Hangul words across tracked prose, code regions excluded."""
    vocab = Counter()
    for path in git_tracked_prose_files(repo_root, exclude=exclude):
        try:
            with open(path, "r", encoding="utf-8", errors="replace") as fh:
                text = fh.read()
        except OSError:
            continue
        for line in strip_code_regions(text):
            for w in HANGUL_WORD_RE.findall(line):
                if len(w) >= min_len:
                    vocab[w] += 1
    return vocab


# ---------------------------------------------------------------------------
# Scanning
# ---------------------------------------------------------------------------

class Finding(object):
    __slots__ = ("path", "line", "col", "word", "candidates")

    def __init__(self, path, line, col, word, candidates):
        self.path = path
        self.line = line
        self.col = col
        self.word = word
        self.candidates = candidates  # list of (word, distance, freq)


def scan_text(text, path_label, vocab, max_distance=2, min_cand_freq=2,
              flag_jamo_blocks=True):
    """Scan one text blob; return findings outside code regions."""
    findings = []
    vocab_by_len = {}
    for w in vocab:
        vocab_by_len.setdefault(len(w), []).append(w)

    for lineno, line in enumerate(strip_code_regions_for_scan(text), 1):
        for m in HANGUL_WORD_RE.finditer(line):
            word = m.group()
            # Detector (a): true-decomposition / stray jamo inside composed text.
            if flag_jamo_blocks:
                jf = find_stray_jamo(line, m.start(), m.end())
                if jf is not None:
                    findings.append(Finding(path_label, lineno, m.start() + 1,
                                            word, [("JAMO-BLOCK-CHAR", 0, 0)]))
                    continue
            if word in vocab:
                continue
            cands = []
            for cand in vocab_by_len.get(len(word), ()):
                freq = vocab[cand]
                if freq < min_cand_freq:
                    continue
                d = slot_distance(word, cand)
                if d is not None and 1 <= d <= max_distance:
                    cands.append((cand, d, freq))
            if not cands:
                continue
            cands.sort(key=lambda t: (t[1], -t[2], t[0]))
            findings.append(Finding(path_label, lineno, m.start() + 1, word, cands))
    return findings


def strip_code_regions_for_scan(text):
    """Same fenced/inline masking as strip_code_regions, as a list of lines."""
    return list(strip_code_regions(text))


def find_stray_jamo(line, start, end):
    """Return the index of a stray jamo char inside the [start,end) word span.

    - U+1100-11FF (NFD jamo) are flagged unconditionally: normal prose never
      carries them.
    - Compatibility jamo (U+3130-318F) are flagged only when sandwiched
      between precomposed syllables (e.g. inside a Hangul word).
    """
    for i in range(start, end):
        o = ord(line[i])
        if JAMO_BLOCK_MIN <= o <= JAMO_BLOCK_MAX:
            return i
        if COMPAT_JAMO_MIN <= o <= COMPAT_JAMO_MAX:
            prev_ok = i > 0 and is_syllable(line[i - 1])
            next_ok = i + 1 < len(line) and is_syllable(line[i + 1])
            if prev_ok or next_ok:
                return i
    return None


def find_stray_jamo_global(text):
    """Detector (a) over a whole blob (used when word-boundary jamo appears)."""
    hits = []
    for lineno, line in enumerate(text.split("\n"), 1):
        for i, ch in enumerate(line):
            o = ord(ch)
            if JAMO_BLOCK_MIN <= o <= JAMO_BLOCK_MAX:
                hits.append((lineno, i + 1, ch))
    return hits


def is_binary_file(path):
    """NUL-byte heuristic; a chunk-boundary UTF-8 split is NOT binary."""
    try:
        with open(path, "rb") as fh:
            chunk = fh.read(8192)
    except OSError:
        return True
    return b"\0" in chunk


def scan_files(paths, vocab, max_distance, min_cand_freq):
    findings = []
    for path in paths:
        if not os.path.isfile(path):
            print("skip (not a file): %s" % path, file=sys.stderr)
            continue
        if is_binary_file(path):
            print("skip (binary/non-UTF-8): %s" % path, file=sys.stderr)
            continue
        with open(path, "r", encoding="utf-8", errors="replace") as fh:
            text = fh.read()
        findings.extend(scan_text(text, path, vocab, max_distance, min_cand_freq))
        # Whole-blob NFD-jamo sweep (catches jamo outside word regex spans too).
        for lineno, col, ch in find_stray_jamo_global(text):
            findings.append(Finding(path, lineno, col,
                                    "U+%04X" % ord(ch),
                                    [("JAMO-BLOCK-CHAR", 0, 0)]))
    return findings


# ---------------------------------------------------------------------------
# Commands
# ---------------------------------------------------------------------------

def dedupe(findings):
    seen = set()
    out = []
    for f in findings:
        key = (f.path, f.line, f.col)
        if key in seen:
            continue
        seen.add(key)
        out.append(f)
    return out


def cmd_check(args):
    repo_root = find_repo_root()
    targets = list(args.paths)
    vocab_excludes = list(targets)
    if args.stdin:
        targets.append("-")
    # Exclude target files from the vocabulary so a file's own corruption
    # cannot legitimize itself.
    vocab = build_vocab(repo_root, exclude=vocab_excludes)
    findings = []
    for path in targets:
        if path == "-":
            text = sys.stdin.read()
            findings.extend(scan_text(text, "<stdin>", vocab,
                                      args.max_distance, args.min_cand_freq))
            for lineno, col, ch in find_stray_jamo_global(text):
                findings.append(Finding("<stdin>", lineno, col, "U+%04X" % ord(ch),
                                        [("JAMO-BLOCK-CHAR", 0, 0)]))
        else:
            findings.extend(scan_files([path], vocab,
                                       args.max_distance, args.min_cand_freq))
    findings = dedupe(findings)
    if not findings:
        print("OK: no suspicious Korean words (%d vocab words, %d file(s))"
              % (len(vocab), len(targets)))
        return 0
    print("SUSPICIOUS: %d finding(s)  [heuristic detector — candidates, not verdicts]"
          % len(findings))
    print("vocab: %d words from git-tracked prose (targets excluded)" % len(vocab))
    for f in findings:
        loc = "%s:%d:%d" % (f.path, f.line, f.col)
        if f.candidates and f.candidates[0][1] == 0:
            print("%s: JAMO %s  <- stray jamo-block character (NFD flavor)"
                  % (loc, f.word))
            continue
        print("%s: %s" % (loc, f.word))
        for cand, d, freq in f.candidates[:args.top]:
            print("    d=%d  %s  (freq %d)" % (d, cand, freq))
        if len(f.candidates) > args.top:
            print("    ... %d more candidate(s)" % (len(f.candidates) - args.top))
    return 1


def is_likely_compound(word, vocab, min_part_freq=20):
    """True if the word splits into two vocabulary words (compound guard).

    Legit compounds (e.g. verb-noun fusions) are frequent false positives of
    the distance-1 rule; never auto-repair a word that parses as two known
    words. Still flagged by check, just protected from --apply.
    """
    for cut in range(1, len(word)):
        head, tail = word[:cut], word[cut:]
        if vocab.get(head, 0) >= min_part_freq and vocab.get(tail, 0) >= min_part_freq:
            return True
    return False


def unambiguous_repairs(findings, vocab=None):
    """Findings whose best candidate is UNIQUE at slot-distance 1.

    With a vocabulary passed in, distance-1 hits that parse as a compound of
    two frequent vocabulary words are withheld from auto-repair (compound
    guard) — they stay flagged for human review.
    """
    repairs = {}
    for f in findings:
        if not f.candidates:
            continue
        best = f.candidates[0]
        if best[1] != 1:
            continue
        same_dist = [c for c in f.candidates if c[1] == 1]
        if len(same_dist) != 1:
            continue
        if vocab is not None and is_likely_compound(f.word, vocab):
            continue
        repairs.setdefault(f.word, best[0])
    return repairs


def apply_repairs(text, repairs):
    """Replace corrupt words outside fenced blocks and inline code spans."""
    lines = text.split("\n")
    fence_re = re.compile(r"^\s*(```|~~~)")
    in_fence = False
    changed = 0
    for i, line in enumerate(lines):
        if fence_re.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        parts = line.split("`")
        for j in range(0, len(parts), 2):  # even segments = outside inline code
            segment = parts[j]

            def sub(m):
                return repairs.get(m.group(), m.group())

            if repairs:
                new = re.sub(r"[가-힣]+", sub, segment)
                if new != segment:
                    parts[j] = new
        new_line = "`".join(parts)
        if new_line != line:
            lines[i] = new_line
            changed += 1
    return "\n".join(lines), changed


def cmd_restore(args):
    repo_root = find_repo_root()
    vocab = build_vocab(repo_root, exclude=list(args.paths))
    findings = dedupe(scan_files(args.paths, vocab,
                                 args.max_distance, args.min_cand_freq))
    repairs = unambiguous_repairs(findings, vocab=vocab)
    ambiguous = [f for f in findings
                 if f.word not in repairs and
                 (not f.candidates or f.candidates[0][1] != 0)]
    mode = "DRY-RUN" if not args.apply else "APPLY"
    print("restore [%s]: %d unambiguous repair(s) proposed, %d flagged ambiguous"
          % (mode, len(repairs), len(ambiguous)))
    for wrong, right in sorted(repairs.items()):
        print("  %s -> %s" % (wrong, right))
    if ambiguous:
        print("ambiguous (human review required):")
        for f in ambiguous:
            cands = ", ".join("%s(d=%d,f=%d)" % (c, d, fq)
                              for c, d, fq in f.candidates[:3])
            print("  %s:%d:%d: %s   %s" % (f.path, f.line, f.col, f.word, cands))
    if not args.apply:
        print("dry-run only; pass --apply to write the unambiguous repairs")
        return 1 if findings else 0
    applied = 0
    by_path = {}
    for path in args.paths:
        with open(path, "r", encoding="utf-8", errors="replace") as fh:
            text = fh.read()
        new_text, n_lines = apply_repairs(text, repairs)
        if new_text != text:
            with open(path, "w", encoding="utf-8") as fh:
                fh.write(new_text)
            applied += 1
            by_path[path] = n_lines
    print("applied repairs to %d file(s): %s"
          % (applied, ", ".join("%s(%d lines)" % kv for kv in by_path.items())))
    remaining = [f for f in dedupe(scan_files(args.paths, vocab,
                                              args.max_distance,
                                              args.min_cand_freq))
                 if f.word not in repairs]
    print("remaining flagged after apply: %d" % len(remaining))
    return 0 if not remaining else 1


# ---------------------------------------------------------------------------
# Selftest
# ---------------------------------------------------------------------------

def corrupt_word(word, syl_index, slot, new_value):
    """Mechanically corrupt ONE jamo slot of ONE syllable (selftest builder).

    syl_index: 1-based syllable position. slot: 0 choseong, 1 jungseong,
    2 jongseong (index into JONG_CHARS; 0 = remove). Substituting the empty
    jongseong (0) with a consonant models the observed jongseong-insertion
    corruption class. Recomposition stays a valid precomposed syllable.
    """
    out = []
    for i, ch in enumerate(word, 1):
        cho, jung, jong = decompose(ch)
        if i == syl_index:
            if slot == 0:
                cho = new_value
            elif slot == 1:
                jung = new_value
            else:
                jong = new_value
        out.append(compose(cho, jung, jong))
    return "".join(out)


def cmd_selftest(_args):
    results = []

    def check(name, cond):
        results.append((name, bool(cond)))
        print("%s  %s" % ("PASS" if cond else "FAIL", name))

    # Controlled fixture: a tiny vocabulary, corrupted words built IN CODE so
    # no output channel can corrupt them between here and the detector.
    vocab = Counter({
        "채널": 40, "바이너리": 30, "순환": 20, "엉터리": 10,
        "브랜치": 25, "나머지": 15, "엔트리": 12, "불변": 8,
    })

    # Corrupt: substitute one jamo slot per word (mechanical).
    cases = {
        "채널": corrupt_word("채널", 2, 2, JONG_CHARS.index("ㅈ")),      # jong ㄹ->ㅈ
        "순환": corrupt_word("순환", 2, 2, JONG_CHARS.index("ㅋ")),      # jong - ->ㅋ
        "브랜치": corrupt_word("브랜치", 2, 1, JUNG_CHARS.index("ㅑ")),  # jung ㅐ->ㅑ
        "나머지": corrupt_word("나머지", 3, 2, JONG_CHARS.index("ㅆ")),  # jong - ->ㅆ
    }

    # 1. the corrupted forms must be valid precomposed syllables (not NFD).
    all_valid = all(is_syllable(ch) for w in cases.values() for ch in w)
    check("corrupt fixtures are valid precomposed syllables (non-NFD class)", all_valid)

    # 2. NFC normalization must NOT undo the corruption (root-cause assertion).
    nfc_unchanged = all(
        unicodedata.normalize("NFC", c) == c for c in cases.values())
    check("NFC normalization does not repair the corruption", nfc_unchanged)

    # 3. check must flag every corrupted form.
    findings = scan_text("\n".join(cases.values()), "selftest", vocab)
    flagged = set(f.word for f in findings)
    check("check flags all %d corrupted forms" % len(cases),
          flagged == set(cases.values()))

    # 4. restore must map each corrupted form to its unique original.
    repairs = unambiguous_repairs(findings)
    expected = {v: k for k, v in cases.items()}
    check("restore resolves all corrupted forms to their originals (unique d=1)",
          repairs == expected)

    # 5. apply_repairs must rewrite text and never touch code regions.
    body = ("본문 %s 은 수리 대상.\n" % cases["채널"]
            + "```text\n fence %s 는 불변\n```\n"
            + "inline `%s` 도 불변.\n") % (cases["순환"], cases["브랜치"])
    new_text, n = apply_repairs(body, repairs)
    check("apply rewrites prose line", cases["채널"] not in new_text)
    check("apply never rewrites fenced block", cases["순환"] in new_text)
    check("apply never rewrites inline code span", cases["브랜치"] in new_text)

    # 6. clean words must NOT be flagged on this fixture.
    clean_findings = scan_text("채널 바이너리 순환 엉터리 브랜치", "selftest-clean", vocab)
    check("no false positive on clean vocabulary words", len(clean_findings) == 0)

    # 7. stray NFD jamo detection (detector a).
    nfd_text = "챈널 테스트"  # compat jamo ㄴ sandwiched inside composed text
    stray = find_stray_jamo_global(nfd_text)
    check("stray jamo (NFD flavor) detected", len(stray) == 1)

    failed = [n for n, ok in results if not ok]
    print("selftest: %d/%d assertions passed" % (len(results) - len(failed), len(results)))
    return 1 if failed else 0


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def find_repo_root():
    try:
        out = subprocess.run(["git", "rev-parse", "--show-toplevel"],
                             capture_output=True, check=True).stdout
        return out.decode("utf-8").strip()
    except (subprocess.CalledProcessError, FileNotFoundError):
        return os.getcwd()


def main(argv=None):
    ap = argparse.ArgumentParser(
        prog="jamo_integrity.py",
        description=("Detect and repair Korean jamo-level corruption. "
                     "HEURISTIC: suggestions are candidates ranked by jamo-slot "
                     "distance and corpus frequency, not verdicts."))
    ap.add_argument("--max-distance", type=int, default=2,
                    help="max jamo-slot distance to consider (default 2)")
    ap.add_argument("--min-cand-freq", type=int, default=2,
                    help="minimum vocabulary frequency for a suggestion (default 2)")
    ap.add_argument("--top", type=int, default=5,
                    help="max suggestions shown per finding (default 5)")
    ap.add_argument("--selftest", action="store_true", dest="selftest_flag",
                    help="run the built-in mechanical selftest (same as 'selftest')")
    sub = ap.add_subparsers(dest="cmd")

    p_check = sub.add_parser("check", help="scan files/stdin for suspicious Korean words")
    p_check.add_argument("paths", nargs="*", help="text files to scan")
    p_check.add_argument("--stdin", action="store_true",
                         help="read text from stdin instead of / in addition to paths")

    p_restore = sub.add_parser("restore",
                               help="propose (--dry-run, default) or apply (--apply) repairs")
    p_restore.add_argument("paths", nargs="+", help="text files to repair")
    group = p_restore.add_mutually_exclusive_group(required=True)
    group.add_argument("--apply", action="store_true",
                       help="write only the unambiguous distance-1 repairs")
    group.add_argument("--dry-run", action="store_true",
                       help="print proposed repairs without writing (default posture)")

    sub.add_parser("selftest", help="run the built-in mechanical selftest")

    args = ap.parse_args(argv)
    if getattr(args, "selftest_flag", False) or args.cmd == "selftest":
        return cmd_selftest(args)
    if args.cmd == "check":
        if not args.paths and not args.stdin:
            ap.error("check needs file paths or --stdin")
        return cmd_check(args)
    if args.cmd == "restore":
        return cmd_restore(args)
    if args.cmd == "selftest":
        return cmd_selftest(args)
    ap.print_help()
    return 2


if __name__ == "__main__":
    sys.exit(main())
