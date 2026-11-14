//! #2871 I1-a: `install/install.sh`, the Rust installer behind `curl -fsSL … | sh`.
//!
//! Every case runs `sh install/install.sh` against a fake release folder served through
//! `MOAI_RELEASE_BASE=file://…` (`releases.json` in the GitHub list shape, assets under
//! `download/<tag>/`), with a temporary HOME and TMPDIR. No network call, no real HOME, no
//! real PATH entry: the script must only write `<state home>/versions/<v>/bin/` and
//! `<state home>/path/`, verify every archive against its `.sha256` (fail closed), and leave
//! shell files alone.
#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    collections::BTreeSet,
    fs,
    os::unix::fs::MetadataExt as _,
    path::PathBuf,
    process::{Command, Output, Stdio},
};

const SCRIPT: &str = concat!(env!("CARGO_MANIFEST_DIR"), "/install/install.sh");
/// The system folders only: the script must work with the tools a fresh machine has.
const SYSTEM_PATH: &str = "/usr/bin:/bin:/usr/sbin:/sbin";
const RC_FILES: [&str; 4] = [".zshrc", ".bashrc", ".profile", ".bash_profile"];

/// The asset triple the script must pick on this machine.
fn host_triple() -> String {
    let os = match std::env::consts::OS {
        "macos" => "apple-darwin",
        "linux" => "unknown-linux-gnu",
        other => panic!("no installer target for {other}"),
    };
    format!("{}-{os}", std::env::consts::ARCH)
}

struct Fixture {
    _dir: tempfile::TempDir,
    root: PathBuf,
    home: PathBuf,
    tmp: PathBuf,
    release: PathBuf,
}

/// One release entry of the fake list; every `moai-cli-v` tag gets assets.
struct Rel {
    tag: &'static str,
    draft: bool,
    prerelease: bool,
}

const RELEASES: [Rel; 5] = [
    // The Go generation's tags share the list and must be skipped by prefix.
    Rel {
        tag: "v3.1.2",
        draft: false,
        prerelease: false,
    },
    Rel {
        tag: "moai-cli-v0.3.0-rc.1",
        draft: false,
        prerelease: true,
    },
    Rel {
        tag: "moai-cli-v0.2.0",
        draft: true,
        prerelease: false,
    },
    Rel {
        tag: "moai-cli-v0.1.0",
        draft: false,
        prerelease: false,
    },
    Rel {
        tag: "moai-cli-v0.0.9",
        draft: false,
        prerelease: false,
    },
];

impl Fixture {
    /// `None` when this machine cannot write a `.tar.xz` (no `tar -J` and no `xz`).
    fn new() -> Option<Self> {
        let dir = tempfile::TempDir::new().unwrap();
        let root = fs::canonicalize(dir.path()).unwrap();
        let home = root.join("home");
        let tmp = root.join("tmp");
        let release = root.join("release");
        for d in [&home, &tmp, &release] {
            fs::create_dir_all(d).unwrap();
        }
        for rc in RC_FILES {
            fs::write(home.join(rc), format!("# {rc} before\n")).unwrap();
        }
        let f = Fixture {
            _dir: dir,
            root,
            home,
            tmp,
            release,
        };
        let mut list = Vec::new();
        for r in &RELEASES {
            if r.tag.starts_with("moai-cli-v") && !f.publish(r.tag)? {
                return None;
            }
            list.push(json!({
                "url": format!("https://api.github.com/repos/x/y/releases/{}", r.tag),
                "author": {"login": "someone", "type": "User"},
                "tag_name": r.tag,
                "name": r.tag,
                "draft": r.draft,
                "prerelease": r.prerelease,
                "assets": [{"name": format!("moai-cli-{}.tar.xz", host_triple()), "state": "uploaded"}],
                // Quoted keys inside a string must not count as fields.
                "body": "notes: \"tag_name\": \"moai-cli-v9.9.9\", \"draft\": false, \"prerelease\": false {[",
            }));
        }
        fs::write(
            f.release.join("releases.json"),
            serde_json::to_string_pretty(&Value::Array(list)).unwrap(),
        )
        .unwrap();
        Some(f)
    }

    fn asset_dir(&self, tag: &str) -> PathBuf {
        self.release.join("download").join(tag)
    }

    fn archive(&self, tag: &str) -> PathBuf {
        self.asset_dir(tag)
            .join(format!("moai-cli-{}.tar.xz", host_triple()))
    }

    fn checksum(&self, tag: &str) -> PathBuf {
        self.asset_dir(tag)
            .join(format!("moai-cli-{}.tar.xz.sha256", host_triple()))
    }

    /// Writes `download/<tag>/moai-cli-<triple>.tar.xz` (a cargo-dist shaped archive whose
    /// `moai-cli` prints its version) and its `.sha256`. `Some(false)` when `tar -J` failed
    /// on a machine without `xz`.
    fn publish(&self, tag: &str) -> Option<bool> {
        let version = tag.strip_prefix("moai-cli-v").unwrap();
        let stage = self.root.join("stage").join(tag);
        let top = stage.join(format!("moai-cli-{}", host_triple()));
        fs::create_dir_all(&top).unwrap();
        let body = format!("#!/bin/sh\necho \"moai-cli {version} (rust)\"\n");
        common::write_executable(&top.join("moai-cli"), &body);
        common::write_executable(&top.join("moai"), &body);
        fs::write(top.join("README.md"), "readme\n").unwrap();
        fs::create_dir_all(self.asset_dir(tag)).unwrap();
        let out = Command::new("tar")
            .clean_env()
            .env("PATH", SYSTEM_PATH)
            .env("LC_ALL", "C")
            .arg("-cJf")
            .arg(self.archive(tag))
            .arg("-C")
            .arg(&stage)
            .arg(format!("moai-cli-{}", host_triple()))
            .output()
            .unwrap();
        if !out.status.success() {
            let has_xz = Command::new("sh")
                .args(["-c", "command -v xz"])
                .env("PATH", SYSTEM_PATH)
                .output()
                .unwrap()
                .status
                .success();
            assert!(
                !has_xz,
                "tar -cJf failed although xz exists: {}",
                String::from_utf8_lossy(&out.stderr)
            );
            return Some(false);
        }
        let bytes = fs::read(self.archive(tag)).unwrap();
        let hex = moai_cli::digest(&bytes)
            .strip_prefix("sha256:")
            .unwrap()
            .to_string();
        fs::write(
            self.checksum(tag),
            format!("{hex} *moai-cli-{}.tar.xz\n", host_triple()),
        )
        .unwrap();
        Some(true)
    }

    fn state(&self) -> PathBuf {
        self.home.join(".moai-cli")
    }

    fn run_env(&self, args: &[&str], env: &[(&str, &str)]) -> Output {
        let mut cmd = Command::new("sh");
        cmd.clean_env()
            .arg(SCRIPT)
            .args(args)
            .current_dir(&self.root)
            .env("PATH", SYSTEM_PATH)
            .env("HOME", &self.home)
            .env("TMPDIR", &self.tmp)
            .env("LC_ALL", "C")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env(
                "MOAI_RELEASE_BASE",
                format!("file://{}", self.release.display()),
            )
            .stdin(Stdio::null());
        for (k, v) in env {
            cmd.env(k, v);
        }
        cmd.output().unwrap()
    }

    fn run(&self, args: &[&str]) -> Output {
        self.run_env(args, &[])
    }

    /// Every file, link and folder under the temporary root, relative to it.
    fn tree(&self) -> BTreeSet<String> {
        let mut out = BTreeSet::new();
        let mut todo = vec![self.root.clone()];
        while let Some(d) = todo.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                out.insert(p.strip_prefix(&self.root).unwrap().display().to_string());
                if fs::symlink_metadata(&p).unwrap().is_dir() {
                    todo.push(p);
                }
            }
        }
        out
    }

    fn assert_rc_files_untouched(&self) {
        for rc in RC_FILES {
            assert_eq!(
                fs::read_to_string(self.home.join(rc)).unwrap(),
                format!("# {rc} before\n"),
                "{rc} changed"
            );
        }
    }

    fn assert_nothing_installed(&self, out: &Output) {
        assert!(!out.status.success(), "{}", text(out));
        assert!(!self.state().exists(), "state home created: {}", text(out));
        assert_eq!(
            fs::read_dir(&self.tmp).unwrap().count(),
            0,
            "TMPDIR left files"
        );
        self.assert_rc_files_untouched();
    }

    /// The installed layout of `version`, and that `path/` names it.
    fn assert_installed(&self, version: &str) {
        let bin = self.state().join("versions").join(version).join("bin");
        let cli = bin.join("moai-cli");
        let moai = bin.join("moai");
        let cli_meta = fs::symlink_metadata(&cli).unwrap();
        assert!(cli_meta.is_file());
        assert_eq!(cli_meta.mode() & 0o111, 0o111, "moai-cli is not executable");
        assert_eq!(
            fs::symlink_metadata(&moai).unwrap().ino(),
            cli_meta.ino(),
            "bin/moai is not a hard link to bin/moai-cli"
        );
        for name in ["moai", "moai-cli"] {
            let link = self.state().join("path").join(name);
            assert!(fs::symlink_metadata(&link).unwrap().is_symlink());
            assert_eq!(fs::read_link(&link).unwrap(), bin.join(name));
            let out = Command::new(&link).clean_env().output().unwrap();
            assert_eq!(
                String::from_utf8_lossy(&out.stdout),
                format!("moai-cli {version} (rust)\n")
            );
        }
    }
}

fn text(out: &Output) -> String {
    format!(
        "status {:?}\nstdout:\n{}\nstderr:\n{}",
        out.status.code(),
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

macro_rules! fixture_or_skip {
    () => {
        match Fixture::new() {
            Some(f) => f,
            None => {
                eprintln!("SKIP: this machine has no `tar -J` and no `xz`; cannot build a .tar.xz fixture");
                return;
            }
        }
    };
}

#[test]
fn installs_the_newest_stable_release_into_its_version_folder_and_links_path() {
    let f = fixture_or_skip!();
    let before = f.tree();
    let out = f.run(&[]);
    assert!(out.status.success(), "{}", text(&out));
    f.assert_installed("0.1.0");
    let stdout = String::from_utf8_lossy(&out.stdout);
    let state = f.state().display().to_string();
    assert!(
        stdout.contains(&format!("export PATH=\"{state}/path:$PATH\"")),
        "{stdout}"
    );
    assert!(stdout.contains("moai install-path-plan"), "{stdout}");
    f.assert_rc_files_untouched();
    // Only the state home is new; TMPDIR is cleaned.
    let added: Vec<String> = f.tree().difference(&before).cloned().collect();
    assert!(!added.is_empty());
    for p in &added {
        assert!(
            p.starts_with("home/.moai-cli"),
            "written outside the state home: {p}"
        );
    }
    assert_eq!(
        fs::read_dir(&f.tmp).unwrap().count(),
        0,
        "TMPDIR left files"
    );
}

#[test]
fn a_checksum_mismatch_stops_and_installs_nothing() {
    let f = fixture_or_skip!();
    fs::write(
        f.checksum("moai-cli-v0.1.0"),
        format!("{} *x\n", "0".repeat(64)),
    )
    .unwrap();
    let out = f.run(&[]);
    f.assert_nothing_installed(&out);
    assert!(
        String::from_utf8_lossy(&out.stderr).contains("checksum"),
        "{}",
        text(&out)
    );
}

#[test]
fn a_tampered_archive_stops_and_installs_nothing() {
    let f = fixture_or_skip!();
    // Archive of another version under the pinned version's checksum.
    fs::copy(f.archive("moai-cli-v0.0.9"), f.archive("moai-cli-v0.1.0")).unwrap();
    let out = f.run(&[]);
    f.assert_nothing_installed(&out);
}

#[test]
fn a_missing_checksum_file_stops_and_installs_nothing() {
    let f = fixture_or_skip!();
    fs::remove_file(f.checksum("moai-cli-v0.1.0")).unwrap();
    let out = f.run(&[]);
    f.assert_nothing_installed(&out);
    let stderr = String::from_utf8_lossy(&out.stderr);
    assert!(stderr.contains("checksum"), "{}", text(&out));
    assert!(!stderr.contains("skipp"), "{}", text(&out));
}

#[test]
fn an_empty_or_malformed_checksum_file_stops() {
    let f = fixture_or_skip!();
    for body in ["", "not-a-hash *x\n", &format!("{} *x\n", "a".repeat(63))] {
        fs::write(f.checksum("moai-cli-v0.1.0"), body).unwrap();
        let out = f.run(&[]);
        f.assert_nothing_installed(&out);
    }
}

#[test]
fn a_rerun_is_idempotent_and_a_second_version_keeps_the_first() {
    let f = fixture_or_skip!();
    let first = f.run(&[]);
    assert!(first.status.success(), "{}", text(&first));
    let tree = f.tree();
    let cli = f.state().join("versions/0.1.0/bin/moai-cli");
    let ino = fs::metadata(&cli).unwrap().ino();
    let again = f.run(&[]);
    assert!(again.status.success(), "{}", text(&again));
    f.assert_installed("0.1.0");
    assert_eq!(f.tree(), tree, "a re-run changed the layout");
    assert_eq!(
        fs::metadata(&cli).unwrap().ino(),
        ino,
        "a re-run rewrote the binary"
    );
    // Another version moves path/ and keeps the first folder.
    let pinned = f.run(&["--version", "0.0.9"]);
    assert!(pinned.status.success(), "{}", text(&pinned));
    f.assert_installed("0.0.9");
    assert!(cli.is_file());
    f.assert_rc_files_untouched();
}

#[test]
fn a_different_build_under_an_installed_version_is_refused() {
    let f = fixture_or_skip!();
    assert!(f.run(&[]).status.success());
    let cli = f.state().join("versions/0.1.0/bin/moai-cli");
    fs::remove_file(f.state().join("versions/0.1.0/bin/moai")).unwrap();
    fs::remove_file(&cli).unwrap();
    common::write_executable(&cli, "#!/bin/sh\necho other\n");
    let out = f.run(&[]);
    assert!(!out.status.success(), "{}", text(&out));
    assert_eq!(fs::read_to_string(&cli).unwrap(), "#!/bin/sh\necho other\n");
}

#[test]
fn version_pin_by_option_or_environment_picks_that_release() {
    for (args, env) in [
        (vec!["--version", "0.0.9"], vec![]),
        (vec!["--version=moai-cli-v0.0.9"], vec![]),
        (vec![], vec![("MOAI_VERSION", "v0.0.9")]),
        // A pin wins over the channel.
        (vec!["--version", "0.0.9", "--channel", "rc"], vec![]),
    ] {
        let f = fixture_or_skip!();
        let out = f.run_env(&args, &env);
        assert!(out.status.success(), "{args:?} {env:?}\n{}", text(&out));
        f.assert_installed("0.0.9");
        assert!(!f.state().join("versions/0.1.0").exists(), "{args:?}");
    }
}

#[test]
fn a_pin_to_a_release_without_assets_fails() {
    let f = fixture_or_skip!();
    let out = f.run(&["--version", "4.0.0"]);
    f.assert_nothing_installed(&out);
}

#[test]
fn pre_releases_are_skipped_unless_the_rc_channel_is_asked_and_drafts_always() {
    let f = fixture_or_skip!();
    let stable = f.run(&["--channel", "stable"]);
    assert!(stable.status.success(), "{}", text(&stable));
    f.assert_installed("0.1.0");
    assert!(!f.state().join("versions/0.3.0-rc.1").exists());
    assert!(!f.state().join("versions/0.2.0").exists());
    let rc = f.run(&["--channel", "rc"]);
    assert!(rc.status.success(), "{}", text(&rc));
    f.assert_installed("0.3.0-rc.1");
    assert!(!f.state().join("versions/0.2.0").exists());
}

#[test]
fn a_list_with_no_matching_release_fails() {
    let f = fixture_or_skip!();
    fs::write(
        f.release.join("releases.json"),
        r#"[{"tag_name":"v3.1.2","draft":false,"prerelease":false},{"tag_name":"moai-cli-v0.3.0-rc.1","draft":false,"prerelease":true}]"#,
    )
    .unwrap();
    let out = f.run(&[]);
    f.assert_nothing_installed(&out);
    // The same compact list does serve the rc channel.
    let rc = f.run(&["--channel", "rc"]);
    assert!(rc.status.success(), "{}", text(&rc));
    f.assert_installed("0.3.0-rc.1");
}

#[test]
fn moai_state_home_moves_the_install_and_nothing_lands_in_home() {
    let f = fixture_or_skip!();
    let elsewhere = f.root.join("state");
    let out = f.run_env(&[], &[("MOAI_STATE_HOME", elsewhere.to_str().unwrap())]);
    assert!(out.status.success(), "{}", text(&out));
    assert!(elsewhere.join("versions/0.1.0/bin/moai-cli").is_file());
    assert_eq!(
        fs::read_link(elsewhere.join("path/moai")).unwrap(),
        elsewhere.join("versions/0.1.0/bin/moai")
    );
    assert!(!f.state().exists());
}

#[test]
fn an_existing_file_in_path_is_kept_aside_not_deleted() {
    let f = fixture_or_skip!();
    let path = f.state().join("path");
    fs::create_dir_all(&path).unwrap();
    fs::write(path.join("moai"), "old shim\n").unwrap();
    let out = f.run(&[]);
    assert!(out.status.success(), "{}", text(&out));
    f.assert_installed("0.1.0");
    let kept: Vec<String> = fs::read_dir(&path)
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().to_string())
        .filter(|n| n.starts_with("moai.before-install-"))
        .collect();
    assert_eq!(kept.len(), 1, "{kept:?}");
    assert_eq!(
        fs::read_to_string(path.join(&kept[0])).unwrap(),
        "old shim\n"
    );
}

#[test]
fn unknown_options_and_channels_are_usage_errors() {
    let f = fixture_or_skip!();
    for args in [
        vec!["--bogus"],
        vec!["--channel", "nightly"],
        vec!["--version"],
    ] {
        let out = f.run(&args);
        assert_eq!(out.status.code(), Some(2), "{args:?}\n{}", text(&out));
        assert!(!f.state().exists());
    }
    let help = f.run(&["--help"]);
    assert!(help.status.success(), "{}", text(&help));
    assert!(String::from_utf8_lossy(&help.stdout).contains("--channel"));
}

#[test]
fn the_script_runs_from_a_pipe_like_curl_sh() {
    let f = fixture_or_skip!();
    let script = fs::read(SCRIPT).unwrap();
    let mut child = Command::new("sh")
        .clean_env()
        .args(["-s", "--", "--version", "0.0.9"])
        .current_dir(&f.root)
        .env("PATH", SYSTEM_PATH)
        .env("HOME", &f.home)
        .env("TMPDIR", &f.tmp)
        .env("LC_ALL", "C")
        .env(
            "MOAI_RELEASE_BASE",
            format!("file://{}", f.release.display()),
        )
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, &script);
    let out = child.wait_with_output().unwrap();
    assert!(out.status.success(), "{}", text(&out));
    f.assert_installed("0.0.9");
    f.assert_rc_files_untouched();
}

#[test]
fn the_script_is_posix_sh_without_rc_or_path_edits() {
    let script = fs::read_to_string(SCRIPT).unwrap();
    assert!(script.starts_with("#!/bin/sh\n"));
    for rc in RC_FILES {
        assert!(!script.contains(rc), "the script names {rc}");
    }
    // The state home is resolved in one place, so stage I2 changes one line.
    assert_eq!(script.matches(".moai-cli").count(), 1);
}
