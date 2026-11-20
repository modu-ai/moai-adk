#!/usr/bin/env python3
"""False-positive report of `moai-cli hook-dispatch --event security-scan` over a repository.

A measuring tool, not a test (SPEC-20261120-security-scan-notify REQ-12): it feeds every tracked
text file of a repository, as one Write call each, to the security-scan duty in a throwaway
project with its own HOME (the rule in its default observe mode), counts the `would_notify`
lines the duty writes, and prints one JSON object: files fed, files skipped and why, findings
per class, and the files with the most findings per class. It writes nothing into the
repository and never touches the real HOME. Nothing here decides pass or fail on a number: the
report is read by a person before the rule is turned to enforce.

Usage:
  python3 -I tools/security_scan_corpus.py --bin PATH --repo PATH [--jobs N] [--top N] [--limit N]

Exit codes: 0 = report printed, 1 = the binary or the repository could not be used, 2 = usage.
"""

import argparse
import concurrent.futures
import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path

MAX_BYTES = 1024 * 1024  # the duty skips longer text itself; the tool counts it as skipped


def tracked_files(repo):
    out = subprocess.run(
        ["git", "-C", str(repo), "ls-files", "-z"], capture_output=True, check=True
    ).stdout
    return [p for p in out.decode("utf-8", "surrogateescape").split("\0") if p]


def read_text(path):
    """(text, None) for a text file the duty would scan, or (None, why) for one it would not."""
    try:
        data = path.read_bytes()
    except OSError:
        return None, "unreadable"
    if len(data) > MAX_BYTES:
        return None, "over 1 MiB"
    if b"\0" in data:
        return None, "binary"
    try:
        return data.decode("utf-8"), None
    except UnicodeDecodeError:
        return None, "not utf-8"


class Worker:
    """One throwaway project and HOME; feeds files one at a time and reads the log growth."""

    def __init__(self, binary, root):
        self.binary = binary
        # The real path: the duty judges "inside the project" on the path the call names, and
        # this tool never creates the files it feeds.
        root = Path(root).resolve()
        self.project = root / "project"
        self.home = root / "home"
        (self.project / ".moai-cli").mkdir(parents=True)
        (self.home / ".moai-cli").mkdir(parents=True)
        # The would_notify lines have a log of their own since SPEC-20261120-security-scan-wire.
        self.log = self.project / ".moai-cli/receipts/security-scan.jsonl"
        self.seen = 0

    def feed(self, rel, text):
        """The would_notify line the duty wrote for this file, or None."""
        call = {
            "session_id": "corpus",
            "hook_event_name": "PostToolUse",
            "tool_name": "Write",
            "tool_use_id": "t",
            "tool_input": {"file_path": str(self.project / rel), "content": text},
            "tool_response": {"success": True},
            "cwd": str(self.project),
        }
        env = {"PATH": os.environ.get("PATH", ""), "HOME": str(self.home)}
        run = subprocess.run(
            [self.binary, "hook-dispatch", "--host", "claude", "--event", "security-scan"],
            input=json.dumps(call).encode(),
            capture_output=True,
            cwd=self.project,
            env=env,
        )
        if run.returncode != 0:
            raise RuntimeError(run.stderr.decode(errors="replace").strip())
        lines = self.log.read_text().splitlines() if self.log.exists() else []
        fresh, self.seen = lines[self.seen:], len(lines)
        for line in fresh:
            record = json.loads(line)
            if record.get("kind") == "would_notify":
                return record
        return None


def scan_chunk(binary, repo, files):
    with tempfile.TemporaryDirectory() as root:
        worker = Worker(binary, root)
        results = []
        for rel in files:
            text, why = read_text(Path(repo) / rel)
            if text is None:
                results.append((rel, None, why))
            else:
                results.append((rel, worker.feed(rel, text), None))
        return results


def report(results, top):
    skipped, per_class, per_file, fed, with_hits = {}, {}, {}, 0, 0
    for rel, record, why in results:
        if why:
            skipped[why] = skipped.get(why, 0) + 1
            continue
        fed += 1
        if record is None:
            continue
        with_hits += 1
        for name in record["classes"]:
            per_class[name] = per_class.get(name, 0) + 1
            per_file.setdefault(name, []).append((rel, record["findings"]))
    return {
        "files_fed": fed,
        "files_with_findings": with_hits,
        "files_skipped": dict(sorted(skipped.items())),
        "files_per_class": dict(sorted(per_class.items())),
        "top_files_per_class": {
            name: [{"path": p, "findings_on_file": n} for p, n in sorted(rows, key=lambda r: (-r[1], r[0]))[:top]]
            for name, rows in sorted(per_file.items())
        },
        "note": "findings_on_file counts all findings of the file, not only this class; the "
        "duty's would_notify line carries class names and a total, never a path or text",
    }


def main(argv):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--bin", required=True, help="the moai-cli binary")
    parser.add_argument("--repo", required=True, help="a git repository")
    parser.add_argument("--jobs", type=int, default=4)
    parser.add_argument("--top", type=int, default=10)
    parser.add_argument("--limit", type=int, default=0, help="feed at most N files (0 = all)")
    args = parser.parse_args(argv)
    if args.jobs < 1 or args.top < 1 or args.limit < 0:
        parser.error("--jobs and --top are at least 1, --limit at least 0")
    if not os.access(args.bin, os.X_OK):
        print(f"not an executable: {args.bin}", file=sys.stderr)
        return 1
    try:
        files = tracked_files(args.repo)
    except (subprocess.CalledProcessError, OSError) as err:
        print(f"cannot list the tracked files of {args.repo}: {err}", file=sys.stderr)
        return 1
    if args.limit:
        files = files[: args.limit]
    chunks = [files[i :: args.jobs] for i in range(args.jobs)]
    try:
        with concurrent.futures.ThreadPoolExecutor(args.jobs) as pool:
            parts = list(pool.map(lambda c: scan_chunk(args.bin, args.repo, c), chunks))
    except RuntimeError as err:
        print(f"the duty failed: {err}", file=sys.stderr)
        return 1
    results = [r for part in parts for r in part]
    print(json.dumps(report(results, args.top), indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
