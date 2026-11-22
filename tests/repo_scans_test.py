"""Tests for tools/repo_scans.json (#2681).

The judge `moai-cli repo-scans` (#2708) is tested through the built binary in
tests/repo_scans_cli.rs (SPEC-20261120-python-binary-tests-to-rust); this module keeps the cases
that read the table and the repository's test files and run no binary.

The table lists tests whose input is the checked-in tree. RepoScansTableTest keeps it complete:
a test file that names the repository root and walks a directory must be in the table or in
EXEMPT with a reason. The heuristic sees direct walks only; a scan hidden inside product code
(an op that reads skills/moai through package_digest, spec_links::extract running
`git ls-files`) is listed by hand and kept by the rule-file existence check below.
"""

import json
import re
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
TABLE = REPO_ROOT / "tools" / "repo_scans.json"


def load(path=TABLE):
    return json.loads(Path(path).read_text(encoding="utf-8"))["scans"]

# Files the heuristic flags that only walk temporary directories or named files.
EXEMPT = {
    "tests/atlas_query_cli.rs": "walks the TempDir tree it built from named atlas fixtures",
    "tests/ci_install_ripgrep_test.py": "rglob over the test's own temporary runner folder",
    "tests/codex_home_update_removed.rs": "read_dir of temporary homes (before/after trees of the "
                                          "Codex and OMP folders); reads no checked-in tree",
    "tests/decision_append_cli.rs": "read_dir of a TempDir docs folder",
    "tests/e2e_host_fixture.rs": "clones into a temp dir, then checks named route and workflow files",
    "tests/forge_cli.rs": "recursive walks of TempDir only",
    "tests/github_cli.rs": "hashes temp dirs; templates are named include_str! files",
    "tests/heavy_slot_cli.rs": "read_dir of a TempDir slot folder; reads the named "
                               "tools/heavy_check_slot.py",
    "tests/home_style_ownership_cli.rs": "read_dir of the temporary home it built (before/after "
                                         "trees); reads the named template and fixtures",
    "tests/home_update_notice_follow_ups.rs": "read_dir copies the named skills/moai and agents "
                                              "folders into a temporary home; reads named fixtures",
    "tests/install_script_test.py": "os.walk of the test's temporary HOME and link folders; runs "
                                    "the named install/install.sh and install.ps1",
    "tests/launcher_settings_receipt.rs": "read_dir of the temp home's receipts folder; reads the "
                                          "named operations document models_revert.md",
    "tests/local_check_test.py": "os.walk of the test's own temporary clone (a before/after "
                                 "snapshot of its working folder); reads the named "
                                 "tools/local_check.sh and .github/workflows/ci.yml",
    "tests/mutate_test.py": "ls-files and scandir on temporary clones",
    "tests/project_flow_closing.rs": "reads only the named skills/moai/workflows/project.md; "
                                    "`git ls-files --error-unmatch` runs in TempDir repos",
    "tests/retired_hosts_layer.rs": "read_dir of temporary homes (before/after trees of a host "
                                    "folder); reads the named src files of the backup move and "
                                    "the skills/moai and agents packages",
    "tests/status_line_cli.rs": "read_dir of a TempDir settings-backups folder",
    "tests/ui_korean_terminal_cli.rs": "read_dir of the TempDir project it set up; reads the "
                                       "named golden screens under tests/fixtures/ui",
    "tests/update_route.rs":"walks its TempDir project; reads the named update.md and routes.json",
    "tests/worktree_report.rs": "read_dir of temporary worktrees",
}

RUST_ROOT = re.compile(r"CARGO_MANIFEST_DIR")
RUST_WALK = re.compile(r"read_dir\(|ls-files|package_digest\(|stage_package\(|package_agents\("
                       r"|walk_files\(|package_files\(")
PY_ROOT = re.compile(r"__file__|REPO_ROOT")
PY_WALK = re.compile(r"\.rglob\(|\.glob\(|os\.walk\(|ls-files|\"grep\", \"-r")

def flagged():
    files = (list(REPO_ROOT.glob("tests/*.rs")) + list(REPO_ROOT.glob("src/**/*.rs"))
             + list(REPO_ROOT.glob("tests/*_test.py")))
    out = set()
    for path in files:
        text = path.read_text(encoding="utf-8", errors="replace")
        if path.suffix == ".rs":
            hit = RUST_ROOT.search(text) and RUST_WALK.search(text)
        else:
            hit = PY_ROOT.search(text) and PY_WALK.search(text)
        if hit:
            out.add(path.relative_to(REPO_ROOT).as_posix())
    return out


def named_files(scans):
    names = set()
    for scan in scans:
        names.update(scan["rules"])
        for test in scan["tests"]:
            names.add(test.split("::")[0].split(" (")[0])
    return names


class RepoScansTableTest(unittest.TestCase):
    def setUp(self):
        self.scans = load()

    def test_entries_are_well_formed(self):
        ids = [s["id"] for s in self.scans]
        self.assertEqual(len(ids), len(set(ids)), "duplicate id")
        for scan in self.scans:
            self.assertTrue(scan["rules"] and scan["scope"] and scan["tests"], scan["id"])
            for item in scan["scope"]:
                self.assertIsInstance(item["prefix"], str, scan["id"])

    def test_every_named_file_exists(self):
        missing = sorted(n for n in named_files(self.scans) if not (REPO_ROOT / n).is_file())
        self.assertEqual(missing, [], "rule or test file named in tools/repo_scans.json is gone")

    def test_each_scanning_test_file_is_one_of_its_rule_files(self):
        # a change to the test itself is a rule change of its scan
        for scan in self.scans:
            for test in scan["tests"]:
                self.assertIn(test.split("::")[0].split(" (")[0], scan["rules"], scan["id"])

    def test_every_walking_test_file_is_listed_or_exempt(self):
        unlisted = sorted(flagged() - named_files(self.scans) - set(EXEMPT))
        self.assertEqual(unlisted, [], "a test walks the repository tree: add it to "
                         "tools/repo_scans.json with its rule files and scope, or to EXEMPT "
                         "with the reason it only walks temporary or named files")

    def test_every_check_names_a_real_target_and_test(self):
        """Static: the gate runs these on the merge result, so a check that selects nothing
        would pass silently (cargo exits 0 on 'running 0 tests'). No cargo is spawned here
        (#2661)."""
        cargo, python = "cargo test --offline --locked ", "python3 -m unittest "
        for scan in self.scans:
            self.assertTrue(scan["checks"], scan["id"])
            for command in scan["checks"]:
                if command.startswith(python):
                    for module in command[len(python):].split():
                        self.assertTrue((REPO_ROOT / (module.replace(".", "/") + ".py")).is_file(),
                                        command)
                    continue
                self.assertTrue(command.startswith(cargo), command)
                words = command[len(cargo):].split()
                sources, filters, i, after_sep = [], [], 0, False
                while i < len(words):
                    if after_sep:
                        filters.append(words[i])
                        i += 1
                    elif words[i] == "--test":
                        path = REPO_ROOT / "tests" / f"{words[i + 1]}.rs"
                        self.assertTrue(path.is_file(), command)
                        sources.append(path)
                        i += 2
                    elif words[i] == "--lib":
                        sources += [p for p in (REPO_ROOT / "src").rglob("*.rs") if p.name != "main.rs"]
                        i += 1
                    elif words[i:i + 2] == ["--bin", "moai-cli"]:
                        sources.append(REPO_ROOT / "src" / "main.rs")
                        i += 2
                    elif words[i] == "--":
                        after_sep = True
                        i += 1
                    else:
                        # cargo takes one test name before `--`; more go after it to libtest
                        self.assertFalse(words[i].startswith("-"), command)
                        self.assertEqual(filters, [], f"{command}: put several names after --")
                        filters.append(words[i])
                        i += 1
                self.assertTrue(sources, command)
                text = "\n".join(p.read_text(encoding="utf-8") for p in sources)
                fns = re.findall(r"\bfn ([a-z_0-9]+)", text)
                for f in filters:
                    if f.endswith("::"):
                        self.assertRegex(text, rf"\bmod {re.escape(f[:-2])}\b", command)
                    else:
                        self.assertTrue(any(f in name for name in fns), f"{command}: {f}")

    def test_exemptions_are_not_stale(self):
        stale = sorted(set(EXEMPT) - flagged())
        self.assertEqual(stale, [], "EXEMPT names a file the heuristic no longer flags")


if __name__ == "__main__":
    unittest.main()
