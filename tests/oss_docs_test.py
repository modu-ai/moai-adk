"""Offline checks for the public open-source documents (Issue #2435, AC-6).

The public files must not carry secrets, personal absolute paths, e-mail addresses or internal
operation records (lane names, pane ids, claim lines). Referral links are allowed only as exact,
listed URLs. Relative links must point at files that exist on this branch. LICENSE must stay the
Apache-2.0 text from `main` byte for byte, and decisions that are still pending must stay visible
as placeholders instead of being filled in by a document edit.

These checks read repository files only. They do not prove that GitHub renders the files, that the
community profile counts them, or that external links resolve.
"""
import hashlib
import pathlib
import re
import tomllib
import unittest

ROOT = pathlib.Path(__file__).resolve().parent.parent

PUBLIC_FILES = [
    "README.md",
    "README.ko.md",
    "CONTRIBUTING.md",
    "SUPPORT.md",
    "SECURITY.md",
    "CODE_OF_CONDUCT.md",
    "LICENSE",
    "docs/llms.txt",
]

# sha256 of `LICENSE` on `main` (Apache-2.0, recognised by GitHub as `Apache-2.0`).
LICENSE_SHA256 = "955dae34deb3b002225a6e3fee61c7b7570b289a8a6957f78cf8cf4da2102707"

FORBIDDEN = {
    # Internal operation records.
    "lane name": re.compile(r"\b(?:agent|pr|omp)-\d+\b|\blanes?\b", re.IGNORECASE),
    "pane id": re.compile(r"\bpane[ _-]?ids?\b|(?<![\w.])%\d+\b", re.IGNORECASE),
    "claim line": re.compile(r"moai-claim(?:-handoff)?:|moai-audit-claim:", re.IGNORECASE),
    # Personal absolute paths.
    "personal path": re.compile(r"/Users/|/home/[a-z]|[A-Za-z]:\\Users\\|~/Library/"),
    # Secret shapes.
    "secret shape": re.compile(
        r"gh[pousr]_[A-Za-z0-9]{8,}|github_pat_|\bsk-[A-Za-z0-9_-]{8,}|AKIA[0-9A-Z]{16}"
        r"|-----BEGIN [A-Z ]*PRIVATE KEY|xox[abprs]-"
    ),
    "e-mail address": re.compile(r"[\w.+-]+@[\w-]+(?:\.[\w-]+)+"),
}

# URLs that may carry a query string (referral or sponsor codes). Exact match only.
ALLOWED_QUERY_URLS = {
    "https://z.ai/subscribe?ic=1NDV03BGWU",
}
URL = re.compile(r"https?://[^\s)<>\"'`]+")

# Decisions that are still on HOLD must stay visible as placeholders.
PLACEHOLDERS = {
    "CODE_OF_CONDUCT.md": "PENDING MAINTAINER DECISION",
    "SECURITY.md": "PENDING MAINTAINER DECISION",
}
# A concrete response-time promise would decide a HOLD by editing a document.
TIME_PROMISE = re.compile(r"\b\d+\s*(?:hours?|days?|business days?|시간|일 이내)\b", re.IGNORECASE)
TIME_PROMISE_FILES = ["SECURITY.md", "CONTRIBUTING.md", "SUPPORT.md", "CODE_OF_CONDUCT.md"]

MD_LINK = re.compile(r"\]\(([^)\s]+)\)")


def read(rel):
    return (ROOT / rel).read_text(encoding="utf-8")


def forbidden_hits(text):
    """Return (label, matched text) for every forbidden pattern found in text."""
    hits = []
    for label, pattern in FORBIDDEN.items():
        hits.extend((label, m.group(0)) for m in pattern.finditer(text))
    return hits


def query_url_hits(text):
    """Return URLs that carry a query string and are not exactly allow-listed."""
    return [u for u in (m.group(0).rstrip(".,") for m in URL.finditer(text))
            if "?" in u and u not in ALLOWED_QUERY_URLS]


def broken_relative_links(rel, text):
    """Return relative link targets in a markdown file that do not exist on disk."""
    base = (ROOT / rel).parent
    broken = []
    for target in MD_LINK.findall(text):
        if re.match(r"^(?:[a-z][a-z0-9+.-]*:|#)", target, re.IGNORECASE):
            continue
        path = target.split("#", 1)[0]
        if path and not (base / path).exists():
            broken.append(target)
    return broken


class PublicDocsTest(unittest.TestCase):
    def test_public_files_exist_and_are_not_empty(self):
        for rel in PUBLIC_FILES:
            with self.subTest(file=rel):
                self.assertTrue((ROOT / rel).is_file(), rel)
                self.assertTrue(read(rel).strip(), rel)

    def test_no_internal_records_paths_or_secrets(self):
        for rel in PUBLIC_FILES:
            with self.subTest(file=rel):
                self.assertEqual(forbidden_hits(read(rel)), [])

    def test_query_urls_are_allow_listed_exactly(self):
        for rel in PUBLIC_FILES:
            with self.subTest(file=rel):
                self.assertEqual(query_url_hits(read(rel)), [])

    def test_relative_links_resolve(self):
        for rel in PUBLIC_FILES:
            if rel.endswith(".md"):
                with self.subTest(file=rel):
                    self.assertEqual(broken_relative_links(rel, read(rel)), [])

    def test_license_is_main_apache_text(self):
        digest = hashlib.sha256((ROOT / "LICENSE").read_bytes()).hexdigest()
        self.assertEqual(digest, LICENSE_SHA256)

    def test_pending_decisions_stay_placeholders(self):
        for rel, marker in PLACEHOLDERS.items():
            with self.subTest(file=rel):
                self.assertIn(marker, read(rel))
        for rel in TIME_PROMISE_FILES:
            with self.subTest(file=rel):
                self.assertEqual(TIME_PROMISE.findall(read(rel)), [])

    def test_code_of_conduct_is_contributor_covenant_3(self):
        text = read("CODE_OF_CONDUCT.md")
        self.assertIn("# Contributor Covenant 3.0 Code of Conduct", text)
        self.assertIn("licensed under CC BY-SA 4.0", text)
        self.assertNotIn("[NOTE:", text)

    def test_security_points_to_private_reporting(self):
        self.assertIn("https://github.com/modu-ai/moai-adk/security/advisories/new", read("SECURITY.md"))

    def test_readmes_link_each_other_and_keep_referral_link(self):
        en, ko = read("README.md"), read("README.ko.md")
        self.assertIn("(README.ko.md)", en)
        self.assertIn("(README.md)", ko)
        for text in (en, ko):
            self.assertIn("https://z.ai/subscribe?ic=1NDV03BGWU", text)


def cargo():
    return tomllib.loads(read("Cargo.toml"))


def binary_names():
    """Every `[[bin]]` name in Cargo.toml (`moai` and its alias `moai-cli`, #2790 R1, #2812 W3)."""
    return [entry["name"] for entry in cargo().get("bin", [])]


def version_signature():
    """What `--version` prints: the clap command name, the package version and the generation
    mark that `VERSION_SIGNATURE` in src/lib.rs appends (#2790 R1). Read from the source, so a
    README that quotes an older output fails here."""
    name = re.search(r'#\[command\((?:(?!\)\]).)*?\bname = "([^"]+)"', read("src/main.rs"), re.S)
    mark = re.search(
        r'VERSION_SIGNATURE: &str = concat!\(env!\("MOAI_BUILD_VERSION"\), "([^"]*)"\)', read("src/lib.rs")
    )
    if not name or not mark:
        raise AssertionError("command name or VERSION_SIGNATURE not found in the source")
    return f"{name.group(1)} {cargo()['package']['version']}{mark.group(1)}"


def one_line(rel, prefix):
    found = [line for line in read(rel).splitlines() if line.startswith(prefix)]
    if len(found) != 1:
        raise AssertionError(f"{rel} must have exactly one line starting {prefix!r}, found {len(found)}")
    return found[0]


# The README lines that name the binary: the Facts line and the compatibility table row.
BINARY_LINES = {
    "README.md": ["- RC1 implementation language: ", "| Command | "],
    "README.ko.md": ["- RC1 구현 언어: ", "| 명령 | "],
}


class ProductFactsTest(unittest.TestCase):
    """The public documents state the binary names and the `--version` output the code has now."""

    def test_readmes_name_every_binary(self):
        names = binary_names()
        self.assertIn("moai", names)
        for rel, prefixes in BINARY_LINES.items():
            for prefix in prefixes:
                line = one_line(rel, prefix)
                if prefix.startswith("|"):
                    # Only the RC1 cell: the Go 3.x cell after it is also `moai`.
                    line = line.split("|")[2]
                for name in names:
                    with self.subTest(file=rel, line=prefix, binary=name):
                        self.assertIn(f"`{name}`", line)

    def test_readmes_quote_the_version_signature(self):
        signature = version_signature()
        stale = signature.rsplit(" ", 1)[0]
        for rel in ("README.md", "README.ko.md", "SUPPORT.md"):
            with self.subTest(file=rel):
                text = read(rel)
                self.assertNotIn(f"`{stale}`", text)
                if rel.startswith("README"):
                    self.assertIn(f"`{signature}`", text)

    def test_signature_reader_reads_the_source(self):
        signature = version_signature()
        self.assertTrue(signature.startswith("moai-cli "), signature)
        self.assertTrue(signature.endswith(" (rust)"), signature)


class DetectorSelfTest(unittest.TestCase):
    """The detectors must flag known-bad samples, so a pass above is not a blind pass."""

    def test_forbidden_samples_are_flagged(self):
        samples = {
            "lane name": "handed to agent-9 by the leader",
            "pane id": "send to pane %12 now",
            "claim line": "moai-claim: session=x epoch=1",
            "personal path": "see /Users/someone/project",
            "secret shape": "token ghp_ABCDEFGHijklmnop1234",
            "e-mail address": "write to someone@example.com",
        }
        for label, sample in samples.items():
            with self.subTest(label=label):
                self.assertIn(label, [h[0] for h in forbidden_hits(sample)])

    def test_query_url_outside_allow_list_is_flagged(self):
        self.assertEqual(query_url_hits("https://z.ai/subscribe?ic=OTHER"),
                         ["https://z.ai/subscribe?ic=OTHER"])
        self.assertEqual(query_url_hits("[x](https://z.ai/subscribe?ic=1NDV03BGWU)"), [])

    def test_broken_relative_link_is_flagged(self):
        self.assertEqual(broken_relative_links("README.md", "[x](no-such-file.md)"), ["no-such-file.md"])
        self.assertEqual(broken_relative_links("README.md", "[x](LICENSE) [y](#facts) [z](https://a.b)"), [])

    def test_time_promise_is_flagged(self):
        self.assertTrue(TIME_PROMISE.findall("We respond within 48 hours."))
        self.assertTrue(TIME_PROMISE.findall("14 days"))


if __name__ == "__main__":
    unittest.main()
