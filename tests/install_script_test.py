"""Tests for install/install.sh and install/install.ps1 (#2935, #2871 I1-a/I1-d).

Every run uses a temporary HOME and a fake GitHub release served from `file://`: a feed
(`releases.atom`), per-tag archives in the cargo-dist shape (`moai-cli-<triple>.tar.xz` plus
`.sha256`), and a fake `moai-cli` that answers `install-inspect` and `--version` only. Nothing
reaches the network, the real HOME or the real PATH.
"""

import hashlib
import io
import os
import platform
import shutil
import stat
import subprocess
import tarfile
import tempfile
import unittest

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SH = os.path.join(REPO_ROOT, "install", "install.sh")
PS1 = os.path.join(REPO_ROOT, "install", "install.ps1")

FAKE_CLI = """#!/bin/sh
case "$1" in
  install-inspect)
    printf '{"protocol_version":1,"operation":"install.inspect","ok":true,"result":{"home_state_root":"%s","path_first_moai":{"path":null,"generation":"%s"},"writes":false}}\\n' "${FAKE_STATE_ROOT:-$HOME/.moai-cli}" "${FAKE_GEN:-absent}"
    ;;
  --version) echo "moai-cli @VERSION@ (rust)" ;;
  *) exit 2 ;;
esac
"""


def host_triple():
    machine = platform.machine().lower()
    arch = {"arm64": "aarch64", "aarch64": "aarch64", "x86_64": "x86_64", "amd64": "x86_64"}[machine]
    system = platform.system()
    if system == "Darwin":
        return f"{arch}-apple-darwin"
    if system == "Linux":
        return f"{arch}-unknown-linux-gnu"
    raise unittest.SkipTest(f"no installer triple for {system}")


def feed(tags):
    entries = "".join(
        f'<entry><link rel="alternate" type="text/html" '
        f'href="https://github.com/modu-ai/moai-adk/releases/tag/{t}"/><title>{t}</title></entry>'
        for t in tags
    )
    return f'<?xml version="1.0" encoding="UTF-8"?><feed xmlns="http://www.w3.org/2005/Atom">{entries}</feed>'


def snapshot(root):
    """Every path under root with its kind and content (bytes digest, link target)."""
    out = {}
    for dirpath, dirnames, filenames in os.walk(root):
        for name in dirnames + filenames:
            p = os.path.join(dirpath, name)
            rel = os.path.relpath(p, root)
            st = os.lstat(p)
            if stat.S_ISLNK(st.st_mode):
                # The link's own time too: re-making the same link is a write.
                out[rel] = ("link", os.readlink(p), st.st_mtime_ns)
            elif stat.S_ISDIR(st.st_mode):
                out[rel] = ("dir",)
            else:
                with open(p, "rb") as f:
                    out[rel] = ("file", hashlib.sha256(f.read()).hexdigest(), st.st_mtime_ns)
    return out


class Release:
    """A fake release folder: `<root>/download/<tag>/<asset>` and `<root>/releases.atom`."""

    def __init__(self, root):
        self.root = root
        self.download = os.path.join(root, "download")
        os.makedirs(self.download)
        self.triple = host_triple()
        self.asset = f"moai-cli-{self.triple}.tar.xz"

    def add(self, version, cli=None, sha256=None, with_sha=True):
        tag = f"moai-cli-v{version}"
        d = os.path.join(self.download, tag)
        os.makedirs(d, exist_ok=True)
        body = (cli or FAKE_CLI.replace("@VERSION@", version)).encode()
        buf = io.BytesIO()
        with tarfile.open(fileobj=buf, mode="w:xz") as tar:
            for name in ("moai-cli", "moai"):
                info = tarfile.TarInfo(f"moai-cli-{self.triple}/{name}")
                info.size = len(body)
                info.mode = 0o755
                tar.addfile(info, io.BytesIO(body))
        data = buf.getvalue()
        with open(os.path.join(d, self.asset), "wb") as f:
            f.write(data)
        if with_sha:
            digest = sha256 if sha256 is not None else hashlib.sha256(data).hexdigest()
            with open(os.path.join(d, self.asset + ".sha256"), "w") as f:
                f.write(f"{digest}  {self.asset}\n")
        return tag

    def publish_feed(self, tags):
        with open(os.path.join(self.root, "releases.atom"), "w") as f:
            f.write(feed(tags))


class InstallShTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.mkdtemp(prefix="moai-install-")
        self.home = os.path.join(self.tmp, "home")
        os.makedirs(self.home)
        self.rel = Release(os.path.join(self.tmp, "release"))
        self.rel.add("1.2.3")
        self.rel.add("1.3.0-rc.1")
        # Newest first, as GitHub's feed lists them; a Go tag in between is never taken.
        self.rel.publish_feed(["moai-cli-v1.3.0-rc.1", "go-v3.1.2", "v3.1.2", "moai-cli-v1.2.3"])

    def tearDown(self):
        shutil.rmtree(self.tmp, ignore_errors=True)

    def env(self, **extra):
        env = {
            "HOME": self.home,
            "PATH": "/usr/bin:/bin",
            "LC_ALL": "C",
            "MOAI_RELEASE_BASE": "file://" + self.rel.download,
            "MOAI_RELEASE_FEED": "file://" + os.path.join(self.rel.root, "releases.atom"),
        }
        env.update(extra)
        return env

    def run_sh(self, *args, **env):
        return subprocess.run(
            ["/bin/sh", SH, *args], env=self.env(**env), cwd=self.home,
            capture_output=True, text=True, timeout=120,
        )

    def state(self, *parts):
        return os.path.join(self.home, ".moai-cli", *parts)

    def test_help_lists_the_go_options_and_unknown_options_fail(self):
        r = self.run_sh("--help")
        self.assertEqual(r.returncode, 0, r.stderr)
        for opt in ("--version", "--install-dir", "--channel", "-h, --help"):
            self.assertIn(opt, r.stdout)
        r = self.run_sh("--nope")
        self.assertNotEqual(r.returncode, 0)
        self.assertIn("Unknown option: --nope", r.stdout + r.stderr)
        self.assertEqual(snapshot(self.home), {})

    def test_latest_stable_installs_into_the_version_folder(self):
        r = self.run_sh()
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        cli = self.state("versions", "1.2.3", "bin", "moai-cli")
        moai = self.state("versions", "1.2.3", "bin", "moai")
        self.assertTrue(os.access(cli, os.X_OK))
        # `moai` is a hard link to `moai-cli` (src/bin_names.rs layout): one file, two names.
        self.assertEqual(os.stat(cli).st_ino, os.stat(moai).st_ino)
        for name in ("moai-cli", "moai"):
            link = self.state("path", name)
            self.assertTrue(os.path.islink(link), name)
            self.assertEqual(os.path.realpath(link), os.path.realpath(self.state("versions", "1.2.3", "bin", name)))
        self.assertIn("moai-cli 1.2.3 (rust)", r.stdout)
        self.assertIn("Checksum verified", r.stdout)
        # No partial folder is left behind.
        self.assertEqual(sorted(os.listdir(self.state("versions"))), ["1.2.3"])

    def test_channel_rc_and_a_pinned_version(self):
        r = self.run_sh("--channel", "rc")
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertTrue(os.path.isfile(self.state("versions", "1.3.0-rc.1", "bin", "moai-cli")))
        r = self.run_sh("--version", "1.2.3")
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertEqual(os.path.realpath(self.state("path", "moai-cli")),
                         os.path.realpath(self.state("versions", "1.2.3", "bin", "moai-cli")))
        r = self.run_sh("--channel", "nightly")
        self.assertNotEqual(r.returncode, 0)

    def test_no_stable_release_in_the_feed_fails_with_the_rc_hint(self):
        self.rel.publish_feed(["moai-cli-v1.3.0-rc.1", "go-v3.1.2"])
        before = snapshot(self.home)
        r = self.run_sh()
        self.assertNotEqual(r.returncode, 0)
        self.assertIn("--channel rc", r.stdout + r.stderr)
        self.assertEqual(snapshot(self.home), before)

    def assert_refused_without_writes(self, *args, **env):
        before = snapshot(self.home)
        r = self.run_sh(*args, **env)
        self.assertNotEqual(r.returncode, 0, r.stdout)
        self.assertEqual(snapshot(self.home), before)
        return r

    def test_a_wrong_missing_or_malformed_checksum_stops_with_nothing_written(self):
        self.rel.add("2.0.0", sha256="0" * 64)
        r = self.assert_refused_without_writes("--version", "2.0.0")
        self.assertIn("Checksum mismatch", r.stdout + r.stderr)
        self.rel.add("2.0.1", with_sha=False)
        r = self.assert_refused_without_writes("--version", "2.0.1")
        self.assertIn("checksum", (r.stdout + r.stderr).lower())
        # A malformed checksum file is named as such, not reported as a mismatch: each half
        # of "a SHA-256 digest" on its own — hex of the wrong length, 64 characters that are
        # not hex.
        for v, digest in (("2.0.2", "not-a-digest"), ("2.0.3", "a" * 63), ("2.0.4", "z" * 64)):
            self.rel.add(v, sha256=digest)
            r = self.assert_refused_without_writes("--version", v)
            self.assertIn("not a SHA-256 digest", r.stdout + r.stderr, v)

    def test_a_version_that_is_not_a_version_is_refused(self):
        for bad in ("../x", "1.2.3;rm", "", "1 2"):
            self.assert_refused_without_writes("--version", bad)
        # Refused even when a release under that text exists: the check is the version's
        # shape, not a failed download.
        for bad in ("a1.2.3", "1.2.3_x", "1..2"):
            self.rel.add(bad)
            self.assert_refused_without_writes("--version", bad)

    def test_a_go_install_on_path_is_left_alone_and_pointed_to_update(self):
        gobin = os.path.join(self.home, "go", "bin")
        os.makedirs(gobin)
        go_moai = os.path.join(gobin, "moai")
        with open(go_moai, "w") as f:
            f.write("#!/bin/sh\necho go moai\n")
        os.chmod(go_moai, 0o755)
        before = snapshot(gobin)
        r = self.run_sh(FAKE_GEN="go", PATH=gobin + ":/usr/bin:/bin")
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertEqual(snapshot(gobin), before)
        self.assertIn("moai-cli update", r.stdout)

    def test_a_second_run_writes_nothing_and_other_bytes_are_refused(self):
        self.assertEqual(self.run_sh("--version", "1.2.3").returncode, 0)
        before = snapshot(self.home)
        r = self.run_sh("--version", "1.2.3")
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertIn("already installed", r.stdout)
        self.assertEqual(snapshot(self.home), before)
        # The same version with other bytes is never written over the installed one.
        self.rel.add("1.2.3", cli=FAKE_CLI.replace("@VERSION@", "1.2.3") + "# other\n")
        r = self.assert_refused_without_writes("--version", "1.2.3")
        self.assertIn("different", (r.stdout + r.stderr).lower())

    def test_the_state_folder_comes_from_the_binary(self):
        target = os.path.join(self.home, ".moai")
        r = self.run_sh("--version", "1.2.3", FAKE_STATE_ROOT=target)
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertTrue(os.path.isfile(os.path.join(target, "versions", "1.2.3", "bin", "moai-cli")))
        self.assertFalse(os.path.exists(self.state()))
        # A state folder the binary names that is not an absolute path is refused.
        self.assert_refused_without_writes("--version", "1.2.3", FAKE_STATE_ROOT="relative/dir")

    def test_install_dir_holds_the_command_links(self):
        d = os.path.join(self.home, "tools")
        r = self.run_sh("--install-dir", d)
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertTrue(os.path.islink(os.path.join(d, "moai-cli")))
        self.assertTrue(os.path.islink(os.path.join(d, "moai")))
        self.assertFalse(os.path.exists(self.state("path")))

    def test_a_command_file_that_is_not_a_link_is_left_alone(self):
        os.makedirs(self.state("path"))
        shim = self.state("path", "moai")
        with open(shim, "w") as f:
            f.write("#!/bin/sh\n# someone's shim\n")
        r = self.run_sh()
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        with open(shim) as f:
            self.assertEqual(f.read(), "#!/bin/sh\n# someone's shim\n")
        self.assertIn("left as it is", r.stdout)
        self.assertTrue(os.path.islink(self.state("path", "moai-cli")))

    def test_shellcheck_passes(self):
        if shutil.which("shellcheck") is None:
            self.skipTest("shellcheck not installed on this machine")
        r = subprocess.run(["shellcheck", "-s", "sh", SH], capture_output=True, text=True)
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)

    def test_the_script_parses_as_posix_sh(self):
        r = subprocess.run(["/bin/sh", "-n", SH], capture_output=True, text=True)
        self.assertEqual(r.returncode, 0, r.stderr)


class InstallPs1Test(InstallShTest):
    """The PowerShell installer, run by `pwsh` on this machine: the same fake release and the
    same expectations. Windows itself (zip, `.exe`, User PATH) is not run here."""

    def setUp(self):
        if shutil.which("pwsh") is None:
            self.skipTest("pwsh not installed on this machine")
        super().setUp()

    def run_sh(self, *args, **env):
        e = self.env(**env)
        # pwsh needs its own folders under HOME; keep them out of the snapshot's way.
        return subprocess.run(
            [shutil.which("pwsh"), "-NoProfile", "-NonInteractive", "-File", PS1, *args],
            env={**e, "PATH": e["PATH"] + ":" + os.path.dirname(shutil.which("pwsh")),
                 "POWERSHELL_TELEMETRY_OPTOUT": "1", "XDG_CACHE_HOME": os.path.join(self.tmp, "cache"),
                 "XDG_DATA_HOME": os.path.join(self.tmp, "data"),
                 "XDG_CONFIG_HOME": os.path.join(self.tmp, "config")},
            cwd=self.home, capture_output=True, text=True, timeout=300,
        )

    def test_shellcheck_passes(self):
        self.skipTest("shell-only check")

    def test_the_script_parses_as_posix_sh(self):
        r = subprocess.run(
            [shutil.which("pwsh"), "-NoProfile", "-NonInteractive", "-Command",
             "$e=$null; [System.Management.Automation.Language.Parser]::ParseFile("
             f"'{PS1}', [ref]$null, [ref]$e) | Out-Null; if ($e.Count) {{ $e; exit 1 }}"],
            capture_output=True, text=True,
        )
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)


if __name__ == "__main__":
    unittest.main()
