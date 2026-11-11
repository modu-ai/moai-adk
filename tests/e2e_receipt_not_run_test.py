"""#2014 T2 pre-check (credential-free half): the host scripts' argument assembly,
receipt format, skill pre-copy and digest recording are verified by RUNNING
tests/e2e/host-claude.sh and host-codex.sh with the credential variables stripped.

Every receipt the operator will later collect must already have this exact shape,
and the credential-gated measurements must come out not_run with the reason naming
the missing credential — never a guessed pass (REQ-08C-01, AGENTS.local.md: a check
that did not run is recorded as not_run). AC-08C-5 (`claude plugin validate`) needs no credentials
and is measured for real here.

Skipped (with an explicit reason) on machines without the local binaries the
scripts shell out to; the scripts themselves are macOS-only operator tools (P12).
"""

import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import unittest

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
CLAUDE_SH = os.path.join(REPO, "tests", "e2e", "host-claude.sh")
CODEX_SH = os.path.join(REPO, "tests", "e2e", "host-codex.sh")
CREDENTIAL_VARS = ("ANTHROPIC_API_KEY", "CODEX_API_KEY")
REQUIRED_FIELDS = [
    "ac",
    "os",
    "arch",
    "host",
    "host_exact_version",
    "mode",
    "skill_digest",
    "moai_cli_sha",
    "result",
    "ran_at",
]


def run_without_credentials(script, receipt_dir):
    env = {k: v for k, v in os.environ.items() if k not in CREDENTIAL_VARS}
    return subprocess.run(
        ["bash", script, receipt_dir],
        cwd=REPO,
        env=env,
        capture_output=True,
        text=True,
        timeout=300,
    )


def read_receipts(receipt_dir):
    path = os.path.join(receipt_dir, "receipts.jsonl")
    with open(path, "r", encoding="utf-8") as f:
        return [json.loads(line) for line in f if line.strip()]


@unittest.skipUnless(
    sys.platform == "darwin" and shutil.which("claude") and os.path.exists(CLAUDE_SH),
    "needs macOS, the claude binary and the script; the scripts are macOS-only "
    "operator tools (P12)",
)
class HostClaudeScriptNotRunReceiptsTest(unittest.TestCase):
    def test_receipts_carry_not_run_and_one_real_static_measurement(self):
        with tempfile.TemporaryDirectory(prefix="t2-claude-") as receipt_dir:
            out = run_without_credentials(CLAUDE_SH, receipt_dir)
            self.assertEqual(out.returncode, 0, out.stderr)
            rows = read_receipts(receipt_dir)
            by_ac = {r["ac"]: r for r in rows}
            self.assertEqual(set(by_ac), {"AC-08C-3", "AC-08C-5", "AC-08C-6"})

            # Credential-gated measurements: not_run, reason names the credential.
            for ac in ("AC-08C-3", "AC-08C-6"):
                self.assertEqual(by_ac[ac]["result"], "not_run", by_ac[ac])
                self.assertIn("ANTHROPIC_API_KEY", by_ac[ac]["detail"])

            # AC-08C-5 needs no credentials and is measured for real.
            self.assertEqual(by_ac["AC-08C-5"]["result"], "passed", by_ac["AC-08C-5"])

            # Schema shape (REQ-08C-01) on every receipt: all required fields
            # present, digest 64-hex, RFC3339-Z timestamp, isolated local mode.
            for r in rows:
                for field in REQUIRED_FIELDS:
                    self.assertIn(field, r, f"{r.get('ac')} lacks {field}")
                self.assertRegex(r["skill_digest"], r"^[0-9a-f]{64}$")
                self.assertRegex(r["ran_at"], r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$")
                self.assertEqual(r["mode"], "local")
                self.assertEqual(r["network"] if "network" in r else "disabled", "disabled")
                # No credential value anywhere: the only credential mention is the
                # variable NAME in a not_run reason.
                blob = json.dumps(r)
                self.assertNotIn("sk-ant-", blob)

            # The skill pre-copy really happened inside the script's temp work —
            # observable through the digest the receipt recorded matching the
            # repository's shipped SKILL.md.
            import hashlib

            with open(os.path.join(REPO, "skills", "moai", "SKILL.md"), "rb") as f:
                digest = hashlib.sha256(f.read()).hexdigest()
            self.assertEqual(by_ac["AC-08C-5"]["skill_digest"], digest)


@unittest.skipUnless(
    sys.platform == "darwin" and os.path.exists(CODEX_SH),
    "needs macOS and the script; the scripts are macOS-only operator tools (P12)",
)
class HostCodexScriptNotRunReceiptTest(unittest.TestCase):
    def test_receipt_is_not_run_naming_the_missing_credential(self):
        with tempfile.TemporaryDirectory(prefix="t2-codex-") as receipt_dir:
            out = run_without_credentials(CODEX_SH, receipt_dir)
            self.assertEqual(out.returncode, 0, out.stderr)
            rows = read_receipts(receipt_dir)
            self.assertEqual([r["ac"] for r in rows], ["AC-08C-4"])
            r = rows[0]
            self.assertEqual(r["result"], "not_run", r)
            self.assertTrue(
                "CODEX_API_KEY" in r["detail"] or "codex binary" in r["detail"],
                r["detail"],
            )
            for field in REQUIRED_FIELDS:
                self.assertIn(field, r, f"{r.get('ac')} lacks {field}")
            self.assertEqual(r["host"], "codex")
            self.assertEqual(r["mode"], "local")


if __name__ == "__main__":
    unittest.main()
