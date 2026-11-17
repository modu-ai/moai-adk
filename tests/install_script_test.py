"""Tests for install/install.sh and install/install.ps1 (#2935, #2871 I1-a/I1-d).

Every run uses a temporary HOME and a fake GitHub release served from `file://`: a feed
(`releases.atom`), per-tag archives in the cargo-dist shape (`moai-cli-<triple>.tar.xz` plus
`.sha256`), and a fake `moai-cli` that answers `install-inspect`, `--version` and (#2935 L1)
`language-set`. Nothing reaches the network, the real HOME or the real PATH.
"""

import hashlib
import io
import json
import os
import platform
import pty
import select
import shutil
import stat
import subprocess
import tarfile
import tempfile
import time
import unittest

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SH = os.path.join(REPO_ROOT, "install", "install.sh")
PS1 = os.path.join(REPO_ROOT, "install", "install.ps1")

# `language-set` (#2935 L1) is what the real moai-cli does, in miniature: record the argv one
# argument per line (ended by `--`) in $FAKE_LOG, write config/language.yaml only when it is not
# there, and answer in the real envelope's shape.
FAKE_CLI = """#!/bin/sh
case "$1" in
  install-inspect)
    printf '{"protocol_version":1,"operation":"install.inspect","ok":true,"result":{"home_state_root":"%s","path_first_moai":{"path":null,"generation":"%s"},"writes":false}}\\n' "${FAKE_STATE_ROOT:-$HOME/.moai-cli}" "${FAKE_GEN:-absent}"
    ;;
  --version) echo "moai-cli @VERSION@ (rust)" ;;
  language-set)
    if [ -n "${FAKE_LOG:-}" ]; then { for a in "$@"; do printf '%s\\n' "$a"; done; echo --; } >> "$FAKE_LOG"; fi
    root="${FAKE_STATE_ROOT:-$HOME/.moai-cli}"
    if [ -e "$root/config/language.yaml" ]; then
      printf '{"ok":true,"result":{"written":false,"reason":"exists"}}\\n'
    else
      mkdir -p "$root/config"
      printf 'language:\\n  conversation_language: %s\\n' "$3" > "$root/config/language.yaml"
      printf '{"ok":true,"result":{"written":true,"reason":null}}\\n'
    fi
    ;;
  *) exit 2 ;;
esac
"""

# An older moai-cli: no `language-set` (an unknown command exits 2).
FAKE_CLI_OLD = """#!/bin/sh
case "$1" in
  install-inspect)
    printf '{"protocol_version":1,"operation":"install.inspect","ok":true,"result":{"home_state_root":"%s","path_first_moai":{"path":null,"generation":"%s"},"writes":false}}\\n' "${FAKE_STATE_ROOT:-$HOME/.moai-cli}" "${FAKE_GEN:-absent}"
    ;;
  --version) echo "moai-cli @VERSION@ (rust)" ;;
  *) if [ -n "${FAKE_LOG:-}" ]; then echo "$*" >> "$FAKE_LOG"; fi; exit 2 ;;
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


# The fake for `install-inspect` and `--version`, with `language-set` handed to the real moai-cli
# this tree built ($REAL_CLI), so the installer's call is answered by the real command.
FAKE_CLI_REAL = FAKE_CLI_OLD.replace(
    "  *) if",
    '  language-set) echo "$*" >> "${FAKE_LOG:-/dev/null}"; exec "$REAL_CLI" "$@" ;;\n  *) if',
)


def real_cli():
    """The moai-cli built from this tree, or None: `MOAI_TEST_CLI`, else the cargo target dir."""
    target = os.environ.get("CARGO_TARGET_DIR") or os.path.join(REPO_ROOT, "target")
    path = os.environ.get("MOAI_TEST_CLI") or os.path.join(target, "debug", "moai-cli")
    return path if os.access(path, os.X_OK) else None


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

    def publish_api(self, releases):
        """GitHub's REST release list, newest first, in its own shape: an author object before
        the tag, assets after it, and a body whose text holds commas and an escaped
        `"tag_name": "moai-cli-v9.9.9"` that must not be read as a release."""
        items = []
        for tag, draft, pre in releases:
            items.append({
                "url": f"https://api.github.com/repos/modu-ai/moai-adk/releases/{len(items)}",
                "author": {"login": "someone", "type": "User", "site_admin": False},
                "node_id": "RE_x", "tag_name": tag, "target_commitish": "main", "name": tag,
                "draft": draft, "immutable": False, "prerelease": pre,
                "created_at": "2026-11-14T00:00:00Z",
                "assets": [{"name": "checksums.txt", "label": "", "uploader": {"login": "bot"}}],
                "body": 'Notes, with commas, and a quote: "tag_name": "moai-cli-v9.9.9", "draft": false',
            })
        with open(os.path.join(self.root, "releases.json"), "w") as f:
            json.dump(items, f, indent=2)


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
            # Outside HOME, so the second-run snapshots do not see it (see FAKE_CLI).
            "FAKE_LOG": os.path.join(self.tmp, "fake.log"),
            "MOAI_RELEASE_BASE": "file://" + self.rel.download,
            "MOAI_RELEASE_FEED": "file://" + os.path.join(self.rel.root, "releases.atom"),
            # Never the real API: until a test publishes it, this file is missing, which is how
            # an unreadable API looks (the feed is used then).
            "MOAI_RELEASE_API": "file://" + os.path.join(self.rel.root, "releases.json"),
        }
        env.update(extra)
        return env

    def test_the_rest_list_finds_a_release_the_feed_has_pushed_out(self):
        # Audit #2943 B1: the feed holds the 10 newest entries; ten Go releases after the Rust
        # one push it out. The REST list (100 newest) still has it as the 11th.
        go = [f"v3.{n}.0" for n in range(10, 0, -1)]
        self.rel.publish_feed(go)
        before = snapshot(self.home)
        r = self.run_sh()
        self.assertNotEqual(r.returncode, 0, "feed only: the 11th is out of reach")
        self.assertEqual(snapshot(self.home), before)
        self.rel.publish_api([(t, False, False) for t in go] + [("moai-cli-v1.2.3", False, False)])
        r = self.run_sh()
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertTrue(os.path.isfile(self.state("versions", "1.2.3", "bin", "moai-cli")))

    def test_the_rest_list_skips_drafts_and_prereleases_and_never_reads_the_body(self):
        self.rel.add("1.4.0")
        self.rel.add("1.5.0")
        self.rel.publish_api([
            ("moai-cli-v1.5.0", True, False),        # a draft
            ("moai-cli-v1.4.0", False, True),        # a prerelease without a `-`
            ("v3.1.2", False, False),
            ("moai-cli-v1.3.0-rc.1", False, True),
            ("moai-cli-v1.2.3", False, False),
        ])
        r = self.run_sh()
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertEqual(sorted(os.listdir(self.state("versions"))), ["1.2.3"])
        r = self.run_sh("--channel", "rc")
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        # rc takes the newest non-draft: 1.4.0 (flagged prerelease), never the draft 1.5.0 or
        # the 9.9.9 written in a body.
        self.assertEqual(sorted(os.listdir(self.state("versions"))), ["1.2.3", "1.4.0"])

    def test_release_sources_must_be_https_or_file(self):
        for var in ("MOAI_RELEASE_BASE", "MOAI_RELEASE_FEED", "MOAI_RELEASE_API"):
            r = self.assert_refused_without_writes(**{var: "http://127.0.0.1:9/x"})
            self.assertIn("https://", r.stdout + r.stderr, var)

    def test_a_v_or_tag_prefix_on_the_version_is_accepted(self):
        for given in ("v1.2.3", "moai-cli-v1.2.3"):
            r = self.run_sh("--version", given)
            self.assertEqual(r.returncode, 0, given + r.stdout + r.stderr)
        self.assertEqual(sorted(os.listdir(self.state("versions"))), ["1.2.3"])

    def test_a_link_to_something_else_is_left_alone(self):
        # Codex review P1 on #2943: a `moai` link in --install-dir that points to the Go
        # edition is someone else's, not a link to replace.
        d = os.path.join(self.home, "go", "bin")
        os.makedirs(d)
        go = os.path.join(d, "moai.go")
        with open(go, "w") as f:
            f.write("#!/bin/sh\necho go\n")
        os.symlink(go, os.path.join(d, "moai"))
        r = self.run_sh("--install-dir", d)
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertEqual(os.readlink(os.path.join(d, "moai")), go)
        self.assertIn("left as it is", r.stdout)
        # Its own link from an earlier version is switched to the new one.
        self.assertTrue(os.path.islink(os.path.join(d, "moai-cli")))
        self.assertEqual(self.run_sh("--install-dir", d, "--channel", "rc").returncode, 0)
        self.assertIn("1.3.0-rc.1", os.readlink(os.path.join(d, "moai-cli")))
        self.assertEqual(os.readlink(os.path.join(d, "moai")), go)
        self.assertEqual([n for n in os.listdir(d) if n.startswith(".")], [])

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
        for opt in ("--version", "--install-dir", "--channel", "--language", "-h, --help"):
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

    def test_the_end_of_an_install_names_the_next_steps_in_order(self):
        # #2935 L3-b: after the install, a project folder, then a session there, then
        # `/moai project` in it. `moai init` is being removed (PR #2949) and is never named.
        r = self.run_sh()
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        out = r.stdout
        done = out.find("Installation complete!")
        self.assertNotEqual(done, -1, out)
        folder = out.find("my-project", done)
        session = out.find("moai cc", done)
        setup = out.find("/moai project", done)
        self.assertNotEqual(folder, -1, out)
        self.assertTrue(folder < session < setup, out)
        self.assertNotIn("moai init", out + r.stderr)

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

    # ---- the conversation language (#2935 L1) ---------------------------------------------

    def calls(self):
        """The `moai-cli` calls the fake recorded, as lists of arguments."""
        if not os.path.exists(self.log):
            return []
        with open(self.log) as f:
            text = f.read()
        return [chunk.split("\n") for chunk in text.split("--\n") if chunk]

    @property
    def log(self):
        return os.path.join(self.tmp, "fake.log")

    def language_file(self):
        return self.state("config", "language.yaml")

    def assert_language(self, code, *args, **env):
        """Runs the installer and checks the one `language-set --home <code>` call it made."""
        if os.path.exists(self.log):
            os.remove(self.log)
        r = self.run_sh(*args, **env)
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertEqual(self.calls(), [["language-set", "--home", code, ""]], r.stdout)
        return r

    def test_the_language_flag_is_recorded_with_the_exact_call(self):
        r = self.assert_language("ko", "--language", "ko")
        self.assertIn("Conversation language recorded: ko", r.stdout)
        with open(self.language_file()) as f:
            self.assertEqual(f.read(), "language:\n  conversation_language: ko\n")
        # The call is made after the commands are installed and linked, by the installed binary.
        self.assertTrue(os.path.islink(self.state("path", "moai-cli")))

    def test_the_language_comes_from_the_flag_then_the_environment_then_the_locale(self):
        self.assert_language("ja", MOAI_LANGUAGE="ja")
        shutil.rmtree(self.state())
        # The flag beats the environment variable, which beats the locale.
        self.assert_language("zh", "--language", "zh", MOAI_LANGUAGE="ja", LC_ALL="ko_KR.UTF-8")
        shutil.rmtree(self.state())
        self.assert_language("ja", MOAI_LANGUAGE="ja", LC_ALL="ko_KR.UTF-8")
        shutil.rmtree(self.state())
        # A code is passed on exactly as typed: no case folding, no region trimming.
        self.assert_language("KO", "--language", "KO")
        shutil.rmtree(self.state())
        self.assert_language("pt-BR", "--language", "pt-BR")
        shutil.rmtree(self.state())
        self.assert_language("es-419", MOAI_LANGUAGE="es-419")

    def test_the_locale_is_read_from_lc_all_then_lc_messages_then_lang(self):
        cases = [
            ({"LC_ALL": "ko_KR.UTF-8", "LC_MESSAGES": "ja_JP.UTF-8", "LANG": "zh_CN.UTF-8"}, "ko"),
            ({"LC_ALL": "", "LC_MESSAGES": "ja_JP.UTF-8", "LANG": "zh_CN.UTF-8"}, "ja"),
            ({"LC_ALL": "", "LC_MESSAGES": "", "LANG": "zh_CN.UTF-8"}, "zh"),
            ({"LC_ALL": "", "LANG": "en_US.UTF-8@euro"}, "en"),
            # C and POSIX are no language: English. The first locale variable that is set decides.
            ({"LC_ALL": "C", "LANG": "ko_KR.UTF-8"}, "en"),
            ({"LC_ALL": "", "LC_MESSAGES": "POSIX", "LANG": "ko_KR.UTF-8"}, "en"),
            ({"LC_ALL": "C.UTF-8"}, "en"),
            # Not a language code: English.
            ({"LC_ALL": "toolong_XX"}, "en"),
            ({"LC_ALL": ""}, "en"),
        ]
        for env, want in cases:
            with self.subTest(env=env):
                shutil.rmtree(self.state(), ignore_errors=True)
                self.assert_language(want, **env)

    def test_a_language_that_is_not_a_code_is_refused_before_anything_is_written(self):
        # The edges of `^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$` and of the 35-character limit: the
        # last text that passes and the first that does not, for the language part, for a `-`
        # group and for the whole code.
        longest = "ko" + "-abcdefgh" * 3 + "-abcde"
        too_long = longest + "f"
        self.assertEqual((len(longest), len(too_long)), (35, 36))
        for good in ("ko", "kor", "ko-KR", "ko-ab", "ko-abcdefgh", "zh-Hant-TW", "KO-kr", longest):
            with self.subTest(good=good):
                shutil.rmtree(self.state(), ignore_errors=True)
                self.assert_language(good, "--language", good)
        for bad in ("", "k", "abcd", "ko-", "ko-a", "ko-abcdefghi", "ko--KR", "ko_KR", "ko KR",
                    "ko;id", "$(id)", "../ko", "한국어", "ko\n", "ko\nxx", "-ko", "--home",
                    "ko'", 'ko"', "ko\t", too_long):
            with self.subTest(bad=bad):
                if os.path.exists(self.log):
                    os.remove(self.log)
                r = self.assert_refused_without_writes("--language", bad)
                self.assertIn("Not a language code", r.stdout + r.stderr)
                self.assertEqual(self.calls(), [], "the binary was never run")
        r = self.assert_refused_without_writes("--language")
        self.assertIn("--language needs a value", r.stdout + r.stderr)
        for bad in ("xx!", "ko\n", "abcd"):
            with self.subTest(env=bad):
                r = self.assert_refused_without_writes(MOAI_LANGUAGE=bad)
                self.assertIn("Not a language code", r.stdout + r.stderr)
        # An empty MOAI_LANGUAGE is no choice at all.
        self.assert_language("en", MOAI_LANGUAGE="")

    def test_a_moai_cli_without_language_set_still_installs_with_a_warning(self):
        old = "1.0.0"
        self.rel.add(old, cli=FAKE_CLI_OLD.replace("@VERSION@", old))
        r = self.run_sh("--version", old, "--language", "ko")
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertIn("Could not record the conversation language (ko)", r.stdout)
        self.assertIn("Installation complete!", r.stdout)
        self.assertIn("moai-cli 1.0.0 (rust)", r.stdout)
        self.assertTrue(os.path.islink(self.state("path", "moai-cli")))
        self.assertFalse(os.path.exists(self.language_file()))
        # The old binary was asked, once, and said no.
        with open(self.log) as f:
            self.assertEqual(f.read(), "language-set --home ko\n")
        # The second run on top of it writes nothing either.
        before = snapshot(self.home)
        r = self.run_sh("--version", old, "--language", "ko")
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertEqual(snapshot(self.home), before)

    def test_a_second_run_writes_nothing_new_and_a_recorded_language_stays(self):
        self.assertEqual(self.run_sh("--language", "ko").returncode, 0)
        with open(self.language_file()) as f:
            self.assertEqual(f.read(), "language:\n  conversation_language: ko\n")
        before = snapshot(self.home)
        r = self.run_sh("--language", "ja")
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertIn("already recorded", r.stdout)
        self.assertEqual(snapshot(self.home), before)
        self.assertEqual(self.calls()[-1], ["language-set", "--home", "ja", ""])
        # A HOME that already had its own value keeps it through a first install.
        shutil.rmtree(self.state())
        os.makedirs(self.state("config"))
        with open(self.language_file(), "w") as f:
            f.write("language:\n  conversation_language: zh\n")
        self.assertEqual(self.run_sh("--language", "ko").returncode, 0)
        with open(self.language_file()) as f:
            self.assertEqual(f.read(), "language:\n  conversation_language: zh\n")

    def test_the_real_language_set_records_the_choice_and_keeps_it(self):
        real = real_cli()
        if real is None:
            self.skipTest("no moai-cli built from this tree (set MOAI_TEST_CLI)")
        self.rel.add("1.6.0", cli=FAKE_CLI_REAL.replace("@VERSION@", "1.6.0"))
        r = self.run_sh("--version", "1.6.0", "--language", "ko", REAL_CLI=real)
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertIn("Conversation language recorded: ko", r.stdout)
        with open(self.language_file(), "rb") as f:
            self.assertEqual(f.read(), b"language:\n  conversation_language: ko\n")
        # Only the one file is new under the state folder's config.
        self.assertEqual(os.listdir(self.state("config")), ["language.yaml"])
        before = snapshot(self.home)
        r = self.run_sh("--version", "1.6.0", "--language", "ja", REAL_CLI=real)
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertIn("already recorded", r.stdout)
        self.assertEqual(snapshot(self.home), before)

    def test_the_real_language_set_stores_a_code_moai_has_no_name_for(self):
        real = real_cli()
        if real is None:
            self.skipTest("no moai-cli built from this tree (set MOAI_TEST_CLI)")
        self.rel.add("1.6.0", cli=FAKE_CLI_REAL.replace("@VERSION@", "1.6.0"))
        want = b"language:\n  conversation_language: pt-BR\n"
        # By the flag, by the environment, and by menu choice 5 typing the code: the fake binary
        # records `language-set --home pt-BR` and the real one writes the file as typed.
        for how in ("flag", "environment", "menu"):
            with self.subTest(how=how):
                shutil.rmtree(self.state(), ignore_errors=True)
                if os.path.exists(self.log):
                    os.remove(self.log)
                if how == "flag":
                    r = self.run_sh("--version", "1.6.0", "--language", "pt-BR", REAL_CLI=real)
                    code, out = r.returncode, r.stdout + r.stderr
                elif how == "environment":
                    r = self.run_sh("--version", "1.6.0", MOAI_LANGUAGE="pt-BR", REAL_CLI=real)
                    code, out = r.returncode, r.stdout + r.stderr
                else:
                    code, out = self.run_pty(["--version", "1.6.0"], "5\npt-BR\n", REAL_CLI=real)
                self.assertEqual(code, 0, out)
                self.assertIn("Conversation language recorded: pt-BR", out)
                with open(self.log) as f:
                    self.assertEqual(f.read(), "language-set --home pt-BR\n", out)
                with open(self.language_file(), "rb") as f:
                    self.assertEqual(f.read(), want)
                self.assertEqual(os.listdir(self.state("config")), ["language.yaml"])

    def test_the_installers_own_output_stays_english(self):
        r = self.assert_language("ko", "--language", "ko", LC_ALL="ko_KR.UTF-8")
        text = r.stdout + r.stderr
        self.assertTrue(all(ord(c) < 128 for c in text), text)

    # The menu needs a terminal on stdin and stdout: a pseudo-terminal stands in for one.

    def pty_argv(self):
        return ["/bin/sh", SH]

    def pty_env(self, **env):
        return self.env(NO_COLOR="1", **env)

    def run_pty(self, args, answers, timeout=300, after=None, **env):
        """Run the installer at a terminal and type `answers`: at once, or, with `after`, only
        once that text has been printed (the prompt is waiting then, as it is for a person)."""
        master, slave = pty.openpty()
        proc = subprocess.Popen(
            [*self.pty_argv(), *args], stdin=slave, stdout=slave, stderr=slave,
            env=self.pty_env(**env), cwd=self.home, close_fds=True, start_new_session=True,
        )
        os.close(slave)
        if after is None:
            os.write(master, answers.encode())
        out = b""
        deadline = time.monotonic() + timeout
        while time.monotonic() < deadline:
            ready, _, _ = select.select([master], [], [], 0.5)
            if ready:
                try:
                    chunk = os.read(master, 4096)
                except OSError:
                    break
                if not chunk:
                    break
                out += chunk
                if after is not None and after.encode() in out:
                    os.write(master, answers.encode())
                    after = None
            elif proc.poll() is not None:
                break
        else:
            proc.kill()
        proc.wait(timeout=30)
        os.close(master)
        return proc.returncode, out.decode(errors="replace").replace("\r\n", "\n")

    def stdin_argv(self):
        """argv and stdin bytes of a run whose script comes in on stdin, as `curl ... | sh` does;
        None when the shell has no such run."""
        with open(SH, "rb") as f:
            return ["/bin/sh", "-s", "--"], f.read()

    def run_half_terminal(self, stdin_is_terminal, argv=None, stdin_data=b"", timeout=300, **env):
        """Run the installer with a terminal on stdin only, or on stdout and stderr only: the
        other side is a pipe. Returns (exit code, stdout, stderr); with the terminal on the
        output side, stdout and stderr come back together as stdout."""
        master, slave = pty.openpty()
        argv = argv or [*self.pty_argv()]
        common = dict(env=self.pty_env(**env), cwd=self.home, close_fds=True,
                      start_new_session=True)
        if stdin_is_terminal:
            # stdout and stderr go to files: the terminal's own output (the echo of what is typed)
            # must be read off the master side, or a process that has the terminal as its stdin
            # cannot finish exiting.
            out_path = os.path.join(self.tmp, "half.out")
            err_path = os.path.join(self.tmp, "half.err")
            with open(out_path, "wb") as fo, open(err_path, "wb") as fe:
                proc = subprocess.Popen(argv, stdin=slave, stdout=fo, stderr=fe, **common)
                os.close(slave)
                # A menu that wrongly appeared would read these empty answers (the default).
                os.write(master, b"\n\n\n")
                deadline = time.monotonic() + timeout
                while proc.poll() is None and time.monotonic() < deadline:
                    ready, _, _ = select.select([master], [], [], 0.5)
                    if ready:
                        try:
                            os.read(master, 4096)
                        except OSError:
                            break
                if proc.poll() is None:
                    proc.kill()
                proc.wait(timeout=30)
            os.close(master)
            with open(out_path, "rb") as fo, open(err_path, "rb") as fe:
                return (proc.returncode, fo.read().decode(errors="replace"),
                        fe.read().decode(errors="replace"))
        proc = subprocess.Popen(argv, stdin=subprocess.PIPE, stdout=slave, stderr=slave, **common)
        os.close(slave)
        try:
            proc.stdin.write(stdin_data)
            proc.stdin.close()
        except BrokenPipeError:
            pass
        out = b""
        deadline = time.monotonic() + timeout
        while time.monotonic() < deadline:
            ready, _, _ = select.select([master], [], [], 0.5)
            if ready:
                try:
                    chunk = os.read(master, 4096)
                except OSError:
                    break
                if not chunk:
                    break
                out += chunk
            elif proc.poll() is not None:
                break
        else:
            proc.kill()
        proc.wait(timeout=30)
        os.close(master)
        return proc.returncode, out.decode(errors="replace").replace("\r\n", "\n"), ""

    def assert_no_menu_for_a_half_terminal(self, code, out, err):
        # The default (the locale's language) is recorded once and the install runs to its end:
        # nothing was waiting for an answer, and the script text was not eaten as one.
        text = out + err
        self.assertEqual(code, 0, text)
        self.assertNotIn("Choose 1-5", text)
        self.assertNotIn("Which language should the model use", text)
        self.assertEqual(self.calls(), [["language-set", "--home", "ko", ""]], text)
        self.assertIn("Installation complete!", out)
        self.assertIn("Next steps:", out)
        self.assertEqual(os.listdir(os.path.join(self.tmp, "scratch")), [], text)

    def prepare_half_terminal_run(self):
        if os.path.exists(self.log):
            os.remove(self.log)
        shutil.rmtree(self.state(), ignore_errors=True)
        scratch = os.path.join(self.tmp, "scratch")
        shutil.rmtree(scratch, ignore_errors=True)
        os.makedirs(scratch)
        return scratch

    def test_no_menu_when_stdin_is_a_pipe_even_if_the_output_is_a_terminal(self):
        # `curl ... | sh`: the script (or nothing) comes in on stdin, the output is the screen.
        scratch = self.prepare_half_terminal_run()
        code, out, err = self.run_half_terminal(
            False, TMPDIR=scratch, LC_ALL="", LANG="ko_KR.UTF-8")
        self.assert_no_menu_for_a_half_terminal(code, out, err)
        piped = self.stdin_argv()
        if piped is not None:
            argv, script = piped
            scratch = self.prepare_half_terminal_run()
            code, out, err = self.run_half_terminal(
                False, argv=argv, stdin_data=script, TMPDIR=scratch, LC_ALL="",
                LANG="ko_KR.UTF-8")
            self.assert_no_menu_for_a_half_terminal(code, out, err)

    def test_no_menu_when_the_output_is_a_pipe_even_if_stdin_is_a_terminal(self):
        scratch = self.prepare_half_terminal_run()
        code, out, err = self.run_half_terminal(
            True, TMPDIR=scratch, LC_ALL="", LANG="ko_KR.UTF-8")
        self.assert_no_menu_for_a_half_terminal(code, out, err)

    def assert_menu_choice(self, answers, want, prompts=1, after=None, **env):
        if os.path.exists(self.log):
            os.remove(self.log)
        shutil.rmtree(self.state(), ignore_errors=True)
        # The installer's own temporary folder goes here, so a leftover shows.
        scratch = os.path.join(self.tmp, "scratch")
        shutil.rmtree(scratch, ignore_errors=True)
        os.makedirs(scratch)
        code, out = self.run_pty([], answers, after=after, TMPDIR=scratch, **env)
        self.assertEqual(code, 0, out)
        self.assertEqual(out.count("Choose 1-5"), prompts, out)
        self.assertEqual(self.calls(), [["language-set", "--home", want, ""]], out)
        # The install ran to its end whatever the answers were: the tail (the next steps) is
        # printed and the temporary folder is gone.
        self.assertIn("Installation complete!", out)
        self.assertIn("Next steps:", out)
        self.assertEqual(os.listdir(scratch), [], out)
        return out

    def test_the_menu_at_a_terminal_lists_the_five_choices_in_english(self):
        out = self.assert_menu_choice("2\n", "ko")
        for line in ("1) English", "2) 한국어", "3) 日本語", "4) 中文", "5) Other"):
            self.assertIn(line, out)
        self.assertEqual(out.count("Which language should the model use"), 1)

    def test_each_menu_number_picks_its_language(self):
        for answers, want in (("1\n", "en"), ("2\n", "ko"), ("3\n", "ja"), ("4\n", "zh")):
            with self.subTest(answers=answers):
                self.assert_menu_choice(answers, want)

    def test_menu_choice_five_takes_a_typed_code_and_asks_again_for_a_bad_one(self):
        self.assert_menu_choice("5\nfr-CA\n", "fr-CA")
        out = self.assert_menu_choice("5\nbad!\n5\npt-BR\n", "pt-BR", prompts=2)
        self.assertIn("not a language code", out)

    def test_a_wrong_menu_answer_is_asked_again_three_times_at_most(self):
        # Two wrong answers, then a right one.
        self.assert_menu_choice("9\nx\n3\n", "ja", prompts=3)
        # Three wrong answers: the locale's language, or English, is used.
        out = self.assert_menu_choice("9\n0\nx\n", "ja", prompts=3, LC_ALL="", LANG="ja_JP.UTF-8")
        self.assertIn("Using ja.", out)
        out = self.assert_menu_choice("9\n9\n9\n", "en", prompts=3)
        self.assertIn("Using en.", out)
        # The answer after the third wrong one is never read as an answer.
        self.assert_menu_choice("9\n9\n9\n2\n", "en", prompts=3)

    def test_an_empty_menu_answer_takes_the_default_and_the_end_of_input_too(self):
        self.assert_menu_choice("\n", "en")
        self.assert_menu_choice("\n", "zh", LC_ALL="", LANG="zh_CN.UTF-8")
        self.assert_end_of_input_takes_the_default()

    def assert_end_of_input_takes_the_default(self):
        # Ctrl-D on an empty line is the end of input; it is typed once the prompt is waiting.
        out = self.assert_menu_choice(
            "\x04", "ko", after="Choose 1-5", LC_ALL="", LANG="ko_KR.UTF-8")
        self.assertIn("Using ko.", out)

    def test_no_menu_when_the_language_is_given_or_there_is_no_terminal(self):
        for args, env in ((["--language", "ko"], {}), ([], {"MOAI_LANGUAGE": "ko"})):
            with self.subTest(args=args, env=env):
                shutil.rmtree(self.state(), ignore_errors=True)
                code, out = self.run_pty(args, "3\n", **env)
                self.assertEqual(code, 0, out)
                self.assertNotIn("Choose 1-5", out)
                self.assertEqual(self.calls()[-1], ["language-set", "--home", "ko", ""])
        # Piped output (every other test here) never shows it either.
        r = self.run_sh()
        self.assertNotIn("Choose 1-5", r.stdout)

    # --- SPEC-20261117-supported-platforms: exactly macOS arm64, Linux x86_64, Windows x86_64 ---

    SUPPORTED_LINE = "Supported platforms: macOS arm64 (Apple silicon), Linux x86_64, Windows x86_64"

    def fake_platform(self, system, machine, decoy_translated=None):
        """PATH with a fake `uname` first. With `decoy_translated`, also a decoy `sysctl` on PATH
        that answers `sysctl.proc_translated` (the Rosetta 2 flag) with that value: the
        installers ask `/usr/sbin/sysctl` by absolute path, so the decoy must never decide."""
        d = os.path.join(self.tmp, f"fakebin-{system}-{machine}-{decoy_translated}")
        os.makedirs(d, exist_ok=True)
        tools = {"uname": f'#!/bin/sh\ncase "$1" in -m) echo {machine} ;; *) echo {system} ;; esac\n'}
        if decoy_translated is not None:
            tools["sysctl"] = ('#!/bin/sh\n[ "$1 $2" = "-n sysctl.proc_translated" ] && '
                               f'{{ echo {decoy_translated}; exit 0; }}\nexit 1\n')
        for name, body in tools.items():
            path = os.path.join(d, name)
            with open(path, "w") as f:
                f.write(body)
            os.chmod(path, 0o755)
        return d + ":/usr/bin:/bin"

    def test_the_three_supported_platforms_are_detected(self):
        for system, machine, triple in [
            ("Darwin", "arm64", "aarch64-apple-darwin"),
            ("Darwin", "aarch64", "aarch64-apple-darwin"),
            ("Linux", "x86_64", "x86_64-unknown-linux-gnu"),
            ("Linux", "amd64", "x86_64-unknown-linux-gnu"),
        ]:
            with self.subTest(platform=f"{system} {machine}"):
                r = self.run_sh(PATH=self.fake_platform(system, machine))
                out = r.stdout + r.stderr
                self.assertIn(f"Detected platform: {triple}", out)
                self.assertNotIn("Unsupported", out)

    def rosetta_argv(self):
        """The installer run in an x86_64 process under Rosetta 2, or None when that cannot run
        here (not Apple silicon, no Rosetta 2)."""
        if platform.system() != "Darwin" or platform.machine() != "arm64":
            return None
        if subprocess.run(["/usr/bin/arch", "-x86_64", "/usr/bin/true"]).returncode != 0:
            return None
        return ["/usr/bin/arch", "-x86_64", "/bin/sh", SH]

    def test_an_x86_64_shell_under_rosetta_on_apple_silicon_is_arm64(self):
        """The real uname and the real /usr/sbin/sysctl, in a translated x86_64 process."""
        argv = self.rosetta_argv()
        if argv is None:
            self.skipTest("needs Apple silicon with Rosetta 2 (and an x86_64-capable installer host)")
        r = subprocess.run(argv, env=self.run_env(), cwd=self.home, capture_output=True,
                           text=True, timeout=300)
        out = r.stdout + r.stderr
        self.assertIn("Detected platform: aarch64-apple-darwin", out)
        self.assertNotIn("Unsupported", out)

    def run_env(self, **env):
        return self.env(**env)

    def test_an_unsupported_platform_stops_with_the_supported_list_and_writes_nothing(self):
        for system, machine, decoy in [
            ("Darwin", "x86_64", None),  # Intel Mac
            # A `sysctl` earlier on PATH that claims Rosetta 2 is not asked: still an Intel Mac.
            ("Darwin", "x86_64", "1"),
            ("Linux", "aarch64", None),
            ("Linux", "arm64", None),
            ("Linux", "i686", None),
            ("FreeBSD", "amd64", None),
        ]:
            with self.subTest(platform=f"{system} {machine} decoy={decoy}"):
                r = self.assert_refused_without_writes(
                    PATH=self.fake_platform(system, machine, decoy))
                self.assertEqual(r.returncode, 1, r.stdout + r.stderr)
                out = r.stdout + r.stderr
                self.assertIn(f"Unsupported platform: {system} {machine}", out)
                self.assertIn(self.SUPPORTED_LINE, out)
                self.assertNotIn("Detected platform", out)
                self.assertNotIn("Downloading", out)

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

    def pwsh_env(self, **env):
        e = self.env(**env)
        # pwsh needs its own folders under HOME; keep them out of the snapshot's way.
        return {**e, "PATH": e["PATH"] + ":" + os.path.dirname(shutil.which("pwsh")),
                "POWERSHELL_TELEMETRY_OPTOUT": "1", "XDG_CACHE_HOME": os.path.join(self.tmp, "cache"),
                "XDG_DATA_HOME": os.path.join(self.tmp, "data"),
                "XDG_CONFIG_HOME": os.path.join(self.tmp, "config")}

    def run_sh(self, *args, **env):
        return subprocess.run(
            [shutil.which("pwsh"), "-NoProfile", "-NonInteractive", "-File", PS1, *args],
            env=self.pwsh_env(**env), cwd=self.home, capture_output=True, text=True, timeout=300,
        )

    # The menu is read with [Console]::In.ReadLine(): the terminal runs get an interactive host,
    # as a person at a terminal has.
    def pty_argv(self):
        return [shutil.which("pwsh"), "-NoProfile", "-File", PS1]

    def pty_env(self, **env):
        return self.pwsh_env(NO_COLOR="1", **env)

    def stdin_argv(self):
        # `pwsh -File` takes the script from the file; the stdin-is-a-pipe run has empty input.
        return None

    def assert_end_of_input_takes_the_default(self):
        # Not asserted for PowerShell: whether Ctrl-D at a pty prompt closes the input depends
        # on the host (pwsh on Linux waits, pwsh on macOS does not always see it), so the harness
        # cannot drive it reliably. Enter (the empty answers above) takes the default; and a
        # closed input is `$null` in `Read-Answer`, which takes it too. The sh variant keeps it.
        pass

    def test_shellcheck_passes(self):
        self.skipTest("shell-only check")

    def rosetta_argv(self):
        # The pwsh on a machine may be arm64 only; the translated case of install.ps1 is
        # `Get-UnixTarget` with Translated "1", checked from the syntax tree below.
        return None

    def run_env(self, **env):
        return self.pwsh_env(**env)

    def call_function(self, name, body):
        """Takes function `name` from install.ps1's own syntax tree and runs `body` with it."""
        probe = (
            "$a = [System.Management.Automation.Language.Parser]::ParseFile("
            f"'{PS1}', [ref]$null, [ref]$null); "
            "$f = $a.Find({ param($n) $n -is "
            "[System.Management.Automation.Language.FunctionDefinitionAst] -and "
            f"$n.Name -eq '{name}' }}, $true); "
            ". ([scriptblock]::Create($f.Extent.Text)); " + body
        )
        r = subprocess.run([shutil.which("pwsh"), "-NoProfile", "-NonInteractive", "-Command", probe],
                           capture_output=True, text=True, timeout=120)
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        return r.stdout.splitlines()

    def test_unix_targets_take_rosetta_as_arm64_and_nothing_else(self):
        lines = self.call_function("Get-UnixTarget", (
            "foreach ($c in 'Darwin:arm64:','Darwin:aarch64:','Darwin:x86_64:1','Darwin:x86_64:0',"
            "'Darwin:x86_64:','Linux:x86_64:','Linux:amd64:','Linux:aarch64:1','Linux:arm64:',"
            "'FreeBSD:amd64:') { $os, $m, $t = $c.Split(':'); "
            "'{0} -> {1}' -f $c, (Get-UnixTarget $os $m $t) }"))
        self.assertEqual(lines, [
            "Darwin:arm64: -> aarch64-apple-darwin",
            "Darwin:aarch64: -> aarch64-apple-darwin",
            "Darwin:x86_64:1 -> aarch64-apple-darwin",
            "Darwin:x86_64:0 -> ",
            "Darwin:x86_64: -> ",
            "Linux:x86_64: -> x86_64-unknown-linux-gnu",
            "Linux:amd64: -> x86_64-unknown-linux-gnu",
            "Linux:aarch64:1 -> ",
            "Linux:arm64: -> ",
            "FreeBSD:amd64: -> ",
        ])

    def test_the_rosetta_question_goes_to_the_system_sysctl(self):
        text = open(PS1, encoding="utf-8").read()
        self.assertIn("& /usr/sbin/sysctl -n sysctl.proc_translated", text)
        self.assertNotRegex(text, r"&\s+sysctl\b")

    def test_windows_takes_only_x86_64(self):
        """SPEC-20261117-supported-platforms: the Windows branch cannot run here, so its decision,
        `Get-WindowsTarget`, is taken from the script's own syntax tree and called with each
        architecture Windows reports (PROCESSOR_ARCHITECTURE / ARCHITEW6432)."""
        lines = self.call_function("Get-WindowsTarget", (
            "foreach ($c in 'AMD64:1','x64:1','amd64:1','x86:1','ARM64:1','x86:0','IA64:1','') { "
            "$arch, $bits = $c.Split(':'); "
            "'{0}:{1} -> {2}' -f $arch, $bits, (Get-WindowsTarget $arch ($bits -eq '1')) }"))
        self.assertEqual(lines, [
            "AMD64:1 -> x86_64-pc-windows-msvc",
            "x64:1 -> x86_64-pc-windows-msvc",
            "amd64:1 -> x86_64-pc-windows-msvc",
            # A 32-bit PowerShell on 64-bit x86 Windows.
            "x86:1 -> x86_64-pc-windows-msvc",
            "ARM64:1 -> ",
            "x86:0 -> ",
            "IA64:1 -> ",
            ": -> ",
        ])

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
