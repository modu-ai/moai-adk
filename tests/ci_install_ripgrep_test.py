"""Exercise the Linux installer with isolated commands, archives and runner state."""
from contextlib import contextmanager
import hashlib
import io
import json
import os
from pathlib import Path
import shutil
import shlex
import subprocess
import sys
import tarfile
import tempfile
import unittest

import yaml

ROOT = Path(__file__).resolve().parent.parent
SCRIPT = ROOT / "tools/ci-install-ripgrep.sh"
PINS = {
    "x86_64-unknown-linux-musl": "4cf9f2741e6c465ffdb7c26f38056a59e2a2544b51f7cc128ef28337eeae4d8e",
    "aarch64-unknown-linux-gnu": "c827481c4ff4ea10c9dc7a4022c8de5db34a5737cb74484d62eb94a95841ab2f",
}

DRIVER = r'''
import hashlib,json,os,shutil,sys,tarfile,tempfile
from pathlib import Path
root=Path(os.environ['FIXTURE_ROOT']); cmd=Path(sys.argv[0]).name; a=sys.argv[1:]
with (root/'calls.jsonl').open('a') as f:f.write(json.dumps([cmd,a])+'\n')
mode=os.environ.get('FIXTURE_APT_MODE','update_failure')
if cmd=='sudo':
 assert a.pop(0)=='-n';os.execvp(a[0],a)
elif cmd=='timeout':
 assert a[:2]==['--kill-after=5s','45s']
 if mode=='timeout':sys.exit(124)
 os.execvp(a[2],a[2:])
elif cmd=='apt-get':
 assert 'Acquire::Retries=0' in a
 if 'update' in a:
  assert '--error-on=any' in a
  state=root/'updates';n=int(state.read_text())+1 if state.exists() else 1;state.write_text(str(n))
  sys.exit(100 if mode=='update_failure' or (mode=='eventual' and n<3) else 0)
 assert a[-3:]==['install','--yes','ripgrep'],a
 if mode=='install_failure':sys.exit(100)
 (root/'bin/rg').write_text('#!/bin/bash\necho "ripgrep 14.1.1"\n')
 (root/'bin/rg').chmod(0o755)
elif cmd=='mktemp':
 assert a==['-d',str(root/'runner/moai-ripgrep.XXXXXX')],a
 print(tempfile.mkdtemp(prefix='moai-ripgrep.',dir=root/'runner'))
elif cmd=='tar':
 assert len(a)==5 and a[0]=='-xzf' and a[2]=='-C',a
 destination=Path(a[3]);member=a[4]
 assert destination.is_relative_to(root/'runner')
 assert member in ['ripgrep-14.1.1-'+arch+'/rg' for arch in ['x86_64-unknown-linux-musl','aarch64-unknown-linux-gnu']]
 with tarfile.open(a[1],'r:gz') as archive:
  info=archive.getmember(member);assert info.isfile()
  target=destination/member;target.parent.mkdir(parents=True)
  target.write_bytes(archive.extractfile(info).read());target.chmod(info.mode)
elif cmd=='mkdir':
 assert len(a)==1 and Path(a[0]).is_relative_to(root/'runner'),a
 Path(a[0]).mkdir()
elif cmd=='install':
 assert len(a)==4 and a[:2]==['-m','0755'],a
 assert all(Path(v).is_relative_to(root/'runner') for v in a[2:])
 shutil.copyfile(a[2],a[3]);Path(a[3]).chmod(0o755)
elif cmd=='cat':sys.stdout.write(sys.stdin.read())
elif cmd=='brew':
 assert a==['install','ripgrep'];(root/'bin/rg').write_text('#!/bin/bash\necho "ripgrep 14.1.1"\n');(root/'bin/rg').chmod(0o755)
elif cmd=='sleep':pass
elif cmd=='uname':
 assert a==['-m'];print(os.environ.get('FIXTURE_ARCH','x86_64'))
elif cmd=='curl':
 if os.environ.get('FIXTURE_CURL_FAIL'):sys.exit(22)
 url=a[-1];arch='aarch64-unknown-linux-gnu' if 'aarch64' in url else 'x86_64-unknown-linux-musl'
 assert url=='https://github.com/BurntSushi/ripgrep/releases/download/14.1.1/ripgrep-14.1.1-'+arch+'.tar.gz'
 assert '--fail' in a and '--location' in a and a[0]=='--disable'
 shutil.copyfile(root/(arch+'.tar.gz'),a[a.index('--output')+1])
elif cmd=='sha256sum':
 expected,name=sys.stdin.read().split();assert name=='archive.tar.gz'
 sys.exit(0 if hashlib.sha256(Path(name).read_bytes()).hexdigest()==expected else 1)
else:raise AssertionError(cmd)
'''


@unittest.skipUnless(os.name == "posix" and Path("/bin/bash").is_file(), "requires a Unix bash fixture")
class InstallerTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="moai-rg-fixture-")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.bin = self.root / "bin"
        self.bin.mkdir()
        (self.root / "runner").mkdir()
        self.path_file = self.root / "github-path"
        self.path_file.write_text("")
        for name in ("sudo", "timeout", "apt-get", "sleep", "uname", "curl", "sha256sum", "brew",
                     "mktemp", "tar", "install", "mkdir", "cat"):
            f = self.bin / name
            f.write_text(f"#!{sys.executable}\n" + DRIVER)
            f.chmod(0o755)
        # Only the shell is native; file tools cannot inherit host-specific dependencies.
        for name in ("bash",):
            path = shutil.which(name)
            self.assertIsNotNone(path, name)
            (self.bin / name).symlink_to(path)
        source = SCRIPT.read_text()
        for arch, expected in PINS.items():
            data = b'#!/bin/bash\necho "ripgrep 14.1.1"\n'
            archive = self.root / (arch + ".tar.gz")
            with tarfile.open(archive, "w:gz") as tar:
                info = tarfile.TarInfo(f"ripgrep-14.1.1-{arch}/rg")
                info.size = len(data)
                info.mode = 0o755
                tar.addfile(info, io.BytesIO(data))
            source = source.replace(expected, hashlib.sha256(archive.read_bytes()).hexdigest())
        # Pins are substituted only in this isolated copy, never via a production override.
        self.script = self.root / "installer.sh"
        self.script.write_text(source)
        self.env = {"PATH": str(self.bin), "HOME": str(self.root), "RUNNER_OS": "Linux",
                    "RUNNER_TEMP": str(self.root / "runner"), "GITHUB_PATH": str(self.path_file),
                    "FIXTURE_ROOT": str(self.root), "PYTHONDONTWRITEBYTECODE": "1"}

    def set_payload_version(self, text, exit_code=0):
        source = self.script.read_text()
        for arch in PINS:
            archive = self.root / (arch + ".tar.gz")
            old = hashlib.sha256(archive.read_bytes()).hexdigest()
            data = ("#!/bin/bash\nprintf '%s\\n' " + shlex.quote(text) + "\nexit " + str(exit_code) + "\n").encode()
            with tarfile.open(archive, "w:gz") as tar:
                info = tarfile.TarInfo(f"ripgrep-14.1.1-{arch}/rg")
                info.size = len(data)
                info.mode = 0o755
                tar.addfile(info, io.BytesIO(data))
            source = source.replace(old, hashlib.sha256(archive.read_bytes()).hexdigest())
        self.script.write_text(source)

    def run_installer(self, **env):
        return subprocess.run(["/bin/bash", str(self.script)], env={**self.env, **env},
                              capture_output=True, text=True, timeout=15)

    def calls(self, name):
        path = self.root / "calls.jsonl"
        rows = [json.loads(line) for line in path.read_text().splitlines()] if path.exists() else []
        return [args for cmd, args in rows if cmd == name]

    def test_apt_failure_falls_back_after_three_backoff_retries(self):
        result = self.run_installer()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("ripgrep 14.1.1", result.stdout)
        self.assertIn("apt attempt 4/4", result.stdout)
        self.assertIn("fallback download", result.stdout)
        self.assertIn("SHA256 verified", result.stdout)
        self.assertEqual(len(self.calls("apt-get")), 4)
        self.assertEqual(self.calls("sleep"), [["5"], ["10"], ["20"]])
        self.assertEqual(len(self.calls("curl")), 1)
        installed = Path(self.path_file.read_text().strip()) / "rg"
        self.assertTrue(installed.is_file())
        check = subprocess.run([str(installed), "--version"], capture_output=True, text=True)
        self.assertEqual(check.returncode, 0)
        self.assertIn("14.1.1", check.stdout)

    def test_fallback_needs_no_host_tar_gzip_or_file_utilities(self):
        self.assertFalse((self.bin / "gzip").exists())
        result = self.run_installer()
        self.assertEqual(result.returncode, 0, result.stderr)
        for name in ("mktemp", "tar", "mkdir", "install"):
            self.assertFalse((self.bin / name).is_symlink(), name)
            self.assertEqual(len(self.calls(name)), 1, name)
        installed = Path(self.path_file.read_text().strip()) / "rg"
        self.assertEqual(installed.stat().st_mode & 0o777, 0o755)

    def test_existing_rg_never_uses_apt_or_network(self):
        rg = self.bin / "rg"
        rg.write_text('#!/bin/bash\necho "ripgrep 13.0.0"\n')
        rg.chmod(0o755)
        result = self.run_installer()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("13.0.0", result.stdout)
        self.assertEqual(self.calls("apt-get"), [])
        self.assertEqual(self.calls("curl"), [])
        self.assertEqual(self.path_file.read_text(), "")

    def test_existing_broken_rg_is_not_silently_accepted(self):
        rg = self.bin / "rg"
        rg.write_text('#!/bin/bash\nexit 13\n')
        rg.chmod(0o755)
        self.assertEqual(self.run_installer().returncode, 13)
        self.assertEqual(self.calls("apt-get"), [])
        self.assertEqual(self.calls("curl"), [])

    def test_successful_apt_path_never_downloads(self):
        result = self.run_installer(FIXTURE_APT_MODE="success")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(len(self.calls("apt-get")), 2)
        self.assertEqual(self.calls("sleep"), [])
        self.assertEqual(self.calls("curl"), [])
        self.assertEqual(self.path_file.read_text(), "")

    def test_transient_apt_failure_recovers_before_fallback(self):
        result = self.run_installer(FIXTURE_APT_MODE="eventual")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.calls("sleep"), [["5"], ["10"]])
        self.assertEqual(self.calls("curl"), [])

    def test_install_failure_and_timeout_both_reach_verified_fallback(self):
        for mode in ("install_failure", "timeout"):
            with self.subTest(mode=mode), installer_fixture() as other:
                result = other.run_installer(FIXTURE_APT_MODE=mode)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(len(other.calls("curl")), 1)
                self.assertEqual(other.calls("sleep"), [["5"], ["10"], ["20"]])
                self.assertTrue(other.path_file.read_text().strip())

    def test_arm64_selects_the_arm_asset(self):
        result = self.run_installer(FIXTURE_ARCH="aarch64")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("aarch64-unknown-linux-gnu.tar.gz", self.calls("curl")[0][-1])

    def test_checksum_mismatch_never_extracts_executes_or_exports_path(self):
        archive = self.root / "x86_64-unknown-linux-musl.tar.gz"
        expected = hashlib.sha256(archive.read_bytes()).hexdigest()
        self.script.write_text(self.script.read_text().replace(expected, "0" * 64))
        result = self.run_installer()
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(self.path_file.read_text(), "")
        self.assertEqual(list((self.root / "runner").rglob("rg")), [])
        self.assertNotIn("ripgrep 14.1.1", result.stdout)
        self.assertNotIn("SHA256 verified", result.stdout)

    def test_download_failure_does_not_publish_a_path(self):
        result = self.run_installer(FIXTURE_CURL_FAIL="1")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(self.calls("sha256sum"), [])
        self.assertEqual(self.path_file.read_text(), "")

    def test_unsupported_architecture_never_downloads(self):
        result = self.run_installer(FIXTURE_ARCH="armv7l")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(self.calls("curl"), [])
        self.assertEqual(self.path_file.read_text(), "")

    def test_helper_refuses_other_hosts_before_any_install(self):
        for host in ("macOS", "Windows"):
            result = self.run_installer(RUNNER_OS=host)
            self.assertNotEqual(result.returncode, 0)
            self.assertEqual(self.calls("sudo"), [])
            self.assertEqual(self.calls("curl"), [])

    def test_workflow_macos_brew_path_is_still_executable(self):
        flow = yaml.safe_load((ROOT / ".github/workflows/rc-ci.yml").read_text())
        step = next(s for s in flow["jobs"]["check"]["steps"] if s.get("name") == "Install and verify ripgrep (C10a baseline)")
        result = subprocess.run(["/bin/bash", "-e", "-o", "pipefail", "-c", step["run"]],
                                env={**self.env, "RUNNER_OS": "macOS"}, cwd=ROOT,
                                capture_output=True, text=True, timeout=15)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.calls("brew"), [["install", "ripgrep"]])
        self.assertEqual(self.calls("sudo"), [])
        self.assertEqual(self.calls("curl"), [])

    def test_pinned_version_accepts_exact_token_and_official_revision_suffix(self):
        for text in ("ripgrep 14.1.1", "ripgrep 14.1.1 (rev 4649aa9)",
                     "ripgrep 14.1.1 (rev 4649aa9)\n-SIMD -AVX\nPCRE2 is available"):
            with self.subTest(output=text), installer_fixture() as other:
                other.set_payload_version(text)
                result = other.run_installer()
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertIn("SHA256 verified", result.stdout)
                self.assertIn("ripgrep executable first line: " + text.splitlines()[0], result.stdout)
                self.assertTrue(other.path_file.read_text().strip())

    def test_pinned_version_rejects_prefix_aliases_and_wrong_first_line(self):
        for text in ("ripgrep 14.1.10 (rev x)", "ripgrep 14x1x1", "ripgrep 14.1.1-dev",
                     "ripgrep 14.1.1x", "not-ripgrep 14.1.1", "wrong first line\nripgrep 14.1.1"):
            with self.subTest(output=text), installer_fixture() as other:
                other.set_payload_version(text)
                result = other.run_installer()
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("SHA256 verified", result.stdout, "must reach the actual version gate")
                self.assertIn("ripgrep executable first line: " + text.splitlines()[0], result.stdout)
                self.assertEqual(other.path_file.read_text(), "")

    def test_pinned_version_command_failure_logs_first_line_and_refuses_path(self):
        self.set_payload_version("ripgrep 14.1.1 (rev 4649aa9)", exit_code=13)
        result = self.run_installer()
        self.assertEqual(result.returncode, 13)
        self.assertIn("ripgrep executable first line: ripgrep 14.1.1 (rev 4649aa9)", result.stdout)
        self.assertEqual(self.path_file.read_text(), "")


@contextmanager
def installer_fixture():
    case = InstallerTest()
    case.setUp()
    try:
        yield case
    finally:
        case.doCleanups()


if __name__ == "__main__":
    unittest.main()
