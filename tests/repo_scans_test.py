"""Tests for tools/repo_scans.json and tools/repo_scans.py (#2681).

The table lists tests whose input is the checked-in tree. RepoScansTableTest keeps it complete:
a test file that names the repository root and walks a directory must be in the table or in
EXEMPT with a reason. The heuristic sees direct walks only; a scan hidden inside product code
(an op that reads skills/moai through package_digest, spec_links::extract running
`git ls-files`) is listed by hand and kept by the rule-file existence check below.
"""

import json
import os
import re
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
TOOL = REPO_ROOT / "tools" / "repo_scans.py"
sys.path.insert(0, str(REPO_ROOT / "tools"))
import repo_scans  # noqa: E402

# Files the heuristic flags that only walk temporary directories or named files.
EXEMPT = {
    "tests/atlas_query_cli.rs": "walks the TempDir tree it built from named atlas fixtures",
    "tests/ci_install_ripgrep_test.py": "rglob over the test's own temporary runner folder",
    "tests/decision_append_cli.rs": "read_dir of a TempDir docs folder",
    "tests/e2e_host_fixture.rs": "clones into a temp dir, then checks named route and workflow files",
    "tests/forge_cli.rs": "recursive walks of TempDir only",
    "tests/github_cli.rs": "hashes temp dirs; templates are named include_str! files",
    "tests/heavy_slot_cli.rs": "read_dir of a TempDir slot folder; reads the named "
                               "tools/heavy_check_slot.py",
    "tests/mutate_test.py": "ls-files and scandir on temporary clones",
    "tests/update_route.rs": "walks its TempDir project; reads the named update.md and routes.json",
    "tests/worktree_report.rs": "read_dir of temporary worktrees",
}

RUST_ROOT = re.compile(r"CARGO_MANIFEST_DIR")
RUST_WALK = re.compile(r"read_dir\(|ls-files|package_digest\(|stage_package\(|package_agents\("
                       r"|walk_files\(|package_files\(")
PY_ROOT = re.compile(r"__file__|REPO_ROOT")
PY_WALK = re.compile(r"\.rglob\(|\.glob\(|os\.walk\(|ls-files|\"grep\", \"-r")

# The incident (#2681 body): #2642 added a walk of src/**/*.rs, #2632 added a file it rejects.
PR_2642 = ["src/agents.rs", "src/apply.rs", "src/archive.rs", "src/claude.rs", "src/contract.rs",
           "src/decision.rs", "src/decision_log.rs", "src/forge/audit_status.rs",
           "src/forge/report.rs", "src/generation.rs", "src/github_templates.rs", "src/hooks.rs",
           "src/hosts/codex.rs", "src/hosts/codex_config.rs", "src/key_vault.rs", "src/layout.rs",
           "src/message.rs", "src/migrate.rs", "src/profile.rs", "src/recovery.rs",
           "src/update.rs", "src/worktree_gc.rs"]
PR_2632 = ["docs/design-review/appendix/operations/agents_home_apply.md",
           "docs/design-review/appendix/operations/agents_home_plan.md",
           "docs/design-review/appendix/operations/agents_home_revert.md",
           "src/agent_install.rs", "src/lib.rs", "src/ops/agents_home_apply.rs",
           "src/ops/agents_home_plan.rs", "src/ops/agents_home_revert.rs",
           "tests/agents_home_cli.rs"]


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


def run_tool(stdin):
    return subprocess.run([sys.executable, "-I", str(TOOL)], input=stdin,
                          capture_output=True, text=True, cwd=REPO_ROOT)


class RepoScansTableTest(unittest.TestCase):
    def setUp(self):
        self.scans = repo_scans.load()

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


class JudgeTest(unittest.TestCase):
    def setUp(self):
        self.scans = repo_scans.load()

    def judge(self, pr, tip):
        return repo_scans.judge(pr, tip, self.scans)

    def test_the_2642_2632_incident_requires_a_remerge_either_way_round(self):
        for pr, tip, rules_side in ((PR_2642, PR_2632, "pr"), (PR_2632, PR_2642, "tip")):
            remerge, hits = self.judge(pr, tip)
            self.assertTrue(remerge)
            hit = [h for h in hits if h["id"] == "permission-reference-one-copy"]
            self.assertEqual(len(hit), 1, hits)
            self.assertEqual(hit[0]["rules_side"], rules_side)
            self.assertEqual(hit[0]["rules"], ["src/contract.rs"])
            self.assertIn("src/agent_install.rs", hit[0]["scope"])

    def test_scope_on_both_sides_without_a_rule_change_needs_no_remerge(self):
        self.assertEqual(self.judge(["src/agent_install.rs"], ["src/update_x.rs"]), (False, []))

    def test_a_rule_change_alone_needs_no_remerge(self):
        self.assertEqual(self.judge(["src/contract.rs"], ["README.md"]), (False, []))
        self.assertEqual(self.judge(["src/contract.rs"], []), (False, []))

    def test_scope_boundaries(self):
        # suffix filter: a non-.rs file under src/ is outside the permission scan
        remerge, hits = self.judge(["src/contract.rs"], ["src/notes.md"])
        self.assertNotIn("permission-reference-one-copy", [h["id"] for h in hits])
        # prefix is a path prefix: "srcx/" is not "src/"
        remerge, hits = self.judge(["src/contract.rs"], ["srcx/a.rs"])
        self.assertNotIn("permission-reference-one-copy", [h["id"] for h in hits])
        # one level deeper still counts
        remerge, hits = self.judge(["src/contract.rs"], ["src/forge/new.rs"])
        self.assertIn("permission-reference-one-copy", [h["id"] for h in hits])

    def test_an_expected_set_file_counts_as_a_rule(self):
        remerge, hits = self.judge(["skills/moai/references/source-manifest.json"],
                                   ["skills/moai/workflows/plan.md"])
        self.assertTrue(remerge)
        self.assertIn("skill-package-files", [h["id"] for h in hits])

    def test_empty_input_needs_no_remerge(self):
        self.assertEqual(self.judge([], []), (False, []))
        self.assertEqual(self.judge([""], [""]), (False, []))

    def test_checks_are_the_hit_scans_in_table_order_without_duplicates(self):
        _, hits = self.judge(PR_2642, PR_2632)
        checks = repo_scans.checks_for(hits, self.scans)
        self.assertIn("cargo test --offline --locked --lib permission_reference_rule_has_one_copy",
                      checks)
        self.assertEqual(len(checks), len(set(checks)))
        order = [c for s in self.scans for c in s["checks"]]
        self.assertEqual(checks, sorted(checks, key=order.index))
        # two hits of one scan (both directions) give its checks once
        twice = [{"id": "repo-scans-table"}, {"id": "repo-scans-table"}]
        self.assertEqual(repo_scans.checks_for(twice, self.scans),
                         ["python3 -m unittest tests.repo_scans_test"])
        self.assertEqual(repo_scans.checks_for([], self.scans), [])
        # two scans sharing one command give it once, at its first place
        shared = [{"id": "a", "checks": ["x", "y"]}, {"id": "b", "checks": ["y", "z"]}]
        self.assertEqual(repo_scans.checks_for([{"id": "b"}, {"id": "a"}], shared), ["x", "y", "z"])


class CliTest(unittest.TestCase):
    def test_incident_through_stdin(self):
        out = run_tool(json.dumps({"pr": PR_2642, "tip": PR_2632}))
        self.assertEqual(out.returncode, 0, out.stderr)
        report = json.loads(out.stdout)
        self.assertTrue(report["remerge"])
        self.assertIn("permission-reference-one-copy", [h["id"] for h in report["hits"]])

    def test_disjoint_change_sets_pass(self):
        out = run_tool(json.dumps({"pr": ["README.md"], "tip": ["notes/x.txt"]}))
        self.assertEqual(json.loads(out.stdout),
                         {"remerge": False, "hits": [], "checks": [], "reasons": []})

    def test_refs_read_the_change_sets_from_git(self):
        # the same commit on both sides: nothing changed since the merge-base
        out = subprocess.run([sys.executable, "-I", str(TOOL), "--refs", "HEAD", "HEAD"],
                             capture_output=True, text=True, cwd=REPO_ROOT)
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertEqual(json.loads(out.stdout),
                         {"remerge": False, "hits": [], "checks": [], "reasons": []})

    def test_bad_input_fails_safe_with_every_check(self):
        every = [c for s in repo_scans.load() for c in s["checks"]]
        every = [c for i, c in enumerate(every) if c not in every[:i]]
        cases = [("", []), ("not json", []), (json.dumps({"pr": "src/a.rs", "tip": []}), []),
                 (json.dumps({"pr": []}), []), ("", ["--refs", "HEAD"]),
                 ("", ["--refs", "no-such-ref-2681", "HEAD"]), ("", ["--bogus"])]
        for stdin, args in cases:
            out = subprocess.run([sys.executable, "-I", str(TOOL), *args], input=stdin,
                                 capture_output=True, text=True, cwd=REPO_ROOT)
            self.assertEqual(out.returncode, 0, (stdin, args))
            report = json.loads(out.stdout)
            self.assertTrue(report["remerge"], (stdin, args))
            self.assertEqual(report["checks"], every, (stdin, args))
            self.assertTrue(report["reasons"][0].startswith("judge internal error"), (stdin, args))


GIT_ENV = {"GIT_CONFIG_GLOBAL": os.devnull, "GIT_CONFIG_NOSYSTEM": "1",
           "GIT_AUTHOR_DATE": "2026-11-13T00:00:00Z", "GIT_COMMITTER_DATE": "2026-11-13T00:00:00Z"}
INCIDENT_CHECK = "cargo test --offline --locked --lib permission_reference_rule_has_one_copy"


class RefsTableTest(unittest.TestCase):
    """#2693 audit B1: with --refs the table comes from both refs, not from the checkout that
    runs the script, so a scan added by the pull request itself is judged."""

    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory(prefix="repo-scans-refs-")
        self.repo = Path(self.tmp.name) / "repo"
        self.repo.mkdir()
        self.env = dict(os.environ, **GIT_ENV)
        self.git("init", "-q", "-b", "base")
        full = repo_scans.load()
        self.entry = next(s for s in full if s["id"] == "permission-reference-one-copy")
        self.without = [s for s in full if s["id"] != "permission-reference-one-copy"]
        self.write("src/lib.rs", "")
        self.write_table(self.without)
        # the judge as committed: each checkout runs its own copy next to its own table
        self.write("tools/repo_scans.py", TOOL.read_text(encoding="utf-8"))
        self.commit("base")
        # tip side = #2632: a new file in src/ (the scope)
        self.git("checkout", "-q", "-b", "tip")
        self.write("src/agent_install.rs", "// a copy of the rule\n")
        self.commit("tip")
        # pull request side = #2642: the scanning test and its table entry
        # (a branch named "head" would be HEAD on a case-insensitive disk)
        self.git("checkout", "-q", "-b", "pull", "base")
        self.write("src/contract.rs", "// the walking test\n")
        self.write_table(self.without + [self.entry])
        self.commit("head")

    def tearDown(self):
        self.tmp.cleanup()

    def git(self, *args):
        subprocess.run(["git", "-c", "user.name=fixture", "-c", "user.email=fixture@example.invalid",
                        "-c", "commit.gpgsign=false", *args], cwd=self.repo, env=self.env,
                       check=True, capture_output=True)

    def write(self, rel, text):
        path = self.repo / rel
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(text, encoding="utf-8")

    def write_table(self, scans):
        self.write("tools/repo_scans.json", json.dumps({"scans": scans}, indent=2) + "\n")

    def commit(self, message):
        self.git("add", "-A")
        self.git("commit", "-q", "-m", message)

    def judge(self, tip, head, tool=None):
        """Run the checkout's own judge (its table is that checkout's), unless `tool` is given."""
        tool = tool or self.repo / "tools" / "repo_scans.py"
        out = subprocess.run([sys.executable, "-I", str(tool), "--refs", tip, head],
                             capture_output=True, text=True, cwd=self.repo, env=self.env)
        self.assertEqual(out.returncode, 0, out.stderr)
        return json.loads(out.stdout)

    def test_a_scan_the_pull_request_adds_is_judged_from_any_checkout(self):
        for checkout in ("tip", "pull", "base"):
            self.git("checkout", "-q", checkout)
            report = self.judge("tip", "pull")
            self.assertTrue(report["remerge"], checkout)
            self.assertIn("permission-reference-one-copy", [h["id"] for h in report["hits"]],
                          checkout)
            self.assertIn(INCIDENT_CHECK, report["checks"], checkout)

    def test_a_scan_only_the_tip_has_is_judged_too(self):
        # reverse roles: the tip carries the scan, the pull request adds the file
        report = self.judge("pull", "tip")
        self.assertIn(INCIDENT_CHECK, report["checks"])

    def test_refs_without_any_table_fall_back_to_the_script_table(self):
        self.git("checkout", "-q", "-b", "old", "base")
        self.git("rm", "-q", "tools/repo_scans.json")
        self.write("src/contract.rs", "// rule\n")
        self.commit("old head without a table")
        self.git("checkout", "-q", "-b", "old-tip", "base")
        self.git("rm", "-q", "tools/repo_scans.json")
        self.write("src/agent_install.rs", "// copy\n")
        self.commit("old tip without a table")
        # the gate's own (current) judge, whose table has the entry
        report = self.judge("old-tip", "old", tool=TOOL)
        self.assertIn(INCIDENT_CHECK, report["checks"])

    def test_a_broken_table_at_a_ref_fails_safe(self):
        self.git("checkout", "-q", "pull")
        self.write("tools/repo_scans.json", "{ not json")
        self.commit("broken table")
        report = self.judge("tip", "pull", tool=TOOL)
        self.assertTrue(report["remerge"])
        self.assertTrue(report["reasons"][0].startswith("judge internal error"))
        self.assertIn(INCIDENT_CHECK, report["checks"])

    def test_no_readable_table_at_all_says_hold(self):
        # N1: a judge copy with no table beside it and bad input has nothing to run
        lone = Path(self.tmp.name) / "lone" / "repo_scans.py"
        lone.parent.mkdir()
        lone.write_bytes(TOOL.read_bytes())
        out = subprocess.run([sys.executable, "-I", str(lone)], input="not json",
                             capture_output=True, text=True, cwd=self.repo, env=self.env)
        report = json.loads(out.stdout)
        self.assertEqual((out.returncode, report["remerge"], report["checks"]), (0, True, []))
        self.assertTrue(report["reasons"][-1].endswith("hold the merge"), report["reasons"])


if __name__ == "__main__":
    unittest.main()
