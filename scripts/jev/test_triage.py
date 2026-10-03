"""Regression guard for triage.py's ancestor-query classification (card t1392).

scripts/jev/ is local-only (never distributed); this test drives the real
measure() against a throwaway git fixture repo — no mocks, no network.

The 3-way classification: `git merge-base --is-ancestor` exit 0 = ancestor,
exit 1 = genuinely not an ancestor (a measured answer), any other exit = the
lookup itself failed and the ancestry is unmeasured — which must never render
as "NOT an ancestor".
"""
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import triage  # noqa: E402

GIT_CONFIG = [
    "-c", "user.email=triage-test@local",
    "-c", "user.name=triage-test",
    "-c", "init.defaultBranch=main",
]


def _git(cwd, *args):
    subprocess.run(
        ["git", *GIT_CONFIG, *args],
        cwd=cwd, capture_output=True, text=True, check=True,
    )


def _fixture_repo(tmp_path, symbol):
    work = tmp_path / "repo"
    work.mkdir()
    (work / "probe.go").write_text(f"package probe\n\nfunc {symbol}() {{}}\n")
    _git(work, "init", "-q", ".")
    _git(work, "add", "probe.go")
    _git(work, "commit", "-q", "-m", "add probe")
    return work


def test_lookup_failure_is_unmeasured_not_not_ancestor(tmp_path, monkeypatch):
    # No `develop` ref exists: the ancestry query fails (exit 128) and must
    # read as unmeasured, never as a wrong-reason "NOT an ancestor" answer.
    work = _fixture_repo(tmp_path, "orphan_probe_fn")
    monkeypatch.chdir(work)
    monkeypatch.setattr(triage, "INTEGRATION", "develop")
    _, obs = triage.measure("t0000", "premise `orphan_probe_fn` in probe.go")
    assert "ANCESTOR LOOKUP FAILED" in obs
    assert "NOT an ancestor" not in obs


def test_real_ancestor_is_reported_as_ancestor(tmp_path, monkeypatch):
    work = _fixture_repo(tmp_path, "merged_probe_fn")
    _git(work, "branch", "develop")
    monkeypatch.chdir(work)
    monkeypatch.setattr(triage, "INTEGRATION", "develop")
    _, obs = triage.measure("t0000", "premise `merged_probe_fn` in probe.go")
    assert "ancestor of develop" in obs


def test_real_non_ancestor_is_reported_not_ancestor(tmp_path, monkeypatch):
    work = _fixture_repo(tmp_path, "base_probe_fn")
    _git(work, "branch", "develop")
    _git(work, "checkout", "-q", "-b", "side")
    (work / "side.go").write_text("package probe\n\nfunc side_probe_fn() {}\n")
    _git(work, "add", "side.go")
    _git(work, "commit", "-q", "-m", "side work")
    monkeypatch.chdir(work)
    monkeypatch.setattr(triage, "INTEGRATION", "develop")
    _, obs = triage.measure("t0000", "premise `side_probe_fn` in side.go")
    assert "NOT an ancestor" in obs


def test_symbol_is_counted_as_a_fixed_string_not_a_regex(tmp_path, monkeypatch):
    # Card t1419: a card symbol is a literal, never a pattern. `foo.bar`
    # treated as a regex also matches `fooXbar` (the `.` is a wildcard), so
    # the file count over-reports presence. Only lit.go holds the literal.
    work = tmp_path / "repo"
    work.mkdir()
    (work / "lit.go").write_text("package probe\n\n// foo.bar(\n")
    (work / "wild.go").write_text("package probe\n\n// fooXbar\n")
    _git(work, "init", "-q", ".")
    _git(work, "add", "lit.go", "wild.go")
    _git(work, "commit", "-q", "-m", "add probes")
    _git(work, "branch", "develop")
    monkeypatch.chdir(work)
    monkeypatch.setattr(triage, "INTEGRATION", "develop")
    _, obs = triage.measure("t0000", "premise `foo.bar` in lit.go")
    assert "'foo.bar': 1 file(s) contain it in develop" in obs, obs
