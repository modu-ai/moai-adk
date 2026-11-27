//! SPEC-20261124-update-linked-config-targets: `moai update` plans a Claude config file that is a
//! link into the default folder, and a blocked style target says why and what to do next. Every
//! test runs the binary in a temporary HOME: the HOME, the account home and the config folder are
//! given to the child only; the real user Claude folder, the Orca profile and the real HOME are
//! never read or written.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    os::unix::fs::symlink,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const PERM: &str = "test-permission";

struct Fixture {
    _tmp: TempDir,
    home: PathBuf,
    /// A package holding only what a style plan reads: `SKILL.md` and the style source.
    package: PathBuf,
}

impl Fixture {
    fn default_dir(&self) -> PathBuf {
        self.home.join(".claude")
    }
    fn profile(&self) -> PathBuf {
        self.home.join("profiles/p")
    }
    fn outside(&self) -> PathBuf {
        self.home.join("outside")
    }
}

fn fixture(prepare: impl FnOnce(&Path)) -> Fixture {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(&home).unwrap();
    prepare(&home);
    let package = root.join("package");
    fs::create_dir_all(&package).unwrap();
    fs::write(package.join("SKILL.md"), "package\n").unwrap();
    let source = package.join(moai_cli::home_style::SOURCE);
    fs::create_dir_all(source.parent().unwrap()).unwrap();
    fs::write(&source, "## Style\n\nOne.\n").unwrap();
    Fixture {
        _tmp: tmp,
        home,
        package,
    }
}

/// A temporary HOME on the earlier generation's state folder (no marker of this generation).
fn earlier_home() -> Fixture {
    fixture(|home| fs::create_dir_all(home.join(moai_cli::HOME_STATE_DIR)).unwrap())
}

/// A temporary HOME whose state folder is this generation's (the marker is there).
fn current_home() -> Fixture {
    fixture(|home| assert!(moai_cli::home_root::adopt_new(home, "test").unwrap()))
}

fn write(path: &Path, text: &str) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// The Orca-like layout: the profile's `settings.json` and `output-styles` are links into the
/// default folder, whose `settings.json` holds `{"a": 1}`.
fn orca_layout(f: &Fixture) {
    write(&f.default_dir().join("settings.json"), "{\"a\": 1}\n");
    fs::create_dir_all(f.default_dir().join("output-styles")).unwrap();
    fs::create_dir_all(f.profile()).unwrap();
    symlink(
        f.default_dir().join("settings.json"),
        f.profile().join("settings.json"),
    )
    .unwrap();
    symlink(
        f.default_dir().join("output-styles"),
        f.profile().join("output-styles"),
    )
    .unwrap();
}

/// One `home-style-*` run in the profile folder (`None`: the default folder).
fn op(f: &Fixture, folder: Option<&Path>, args: &[&str]) -> (bool, Value) {
    let out = Command::new(EXE)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &f.home)
        .env("MOAI_ACCOUNT_HOME", &f.home)
        .env("LC_ALL", "C")
        .envs(folder.map(|p| ("CLAUDE_CONFIG_DIR", p.to_path_buf())))
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let v = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    (out.status.success(), v)
}

fn plan(f: &Fixture, folder: Option<&Path>) -> Value {
    let (h, p) = (f.home.to_str().unwrap(), f.package.to_str().unwrap());
    let (ok, v) = op(f, folder, &["home-style-plan", h, p]);
    assert!(ok, "{v}");
    v["result"].clone()
}

fn apply(f: &Fixture, folder: Option<&Path>, plan: &Value) -> (bool, Value) {
    let (h, p) = (f.home.to_str().unwrap(), f.package.to_str().unwrap());
    let digest = plan["plan_digest"].as_str().unwrap();
    op(
        f,
        folder,
        &[
            "home-style-apply",
            h,
            p,
            "--plan-digest",
            digest,
            "--permission-reference",
            PERM,
        ],
    )
}

fn target<'a>(plan: &'a Value, kind: &str) -> &'a Value {
    plan["targets"]
        .as_array()
        .unwrap()
        .iter()
        .find(|t| t["kind"] == kind)
        .unwrap_or_else(|| panic!("no {kind} target in {plan}"))
}

fn action(plan: &Value, kind: &str) -> String {
    target(plan, kind)["action"].as_str().unwrap().to_string()
}

fn detail(plan: &Value, kind: &str) -> String {
    target(plan, kind)["detail"]
        .as_str()
        .unwrap_or("")
        .to_string()
}

fn is_link_to(path: &Path, to: &Path) -> bool {
    fs::read_link(path).is_ok_and(|t| t == to)
}

/// The Orca case: both profile targets are links into the default folder, on a HOME whose state
/// folder is still the earlier generation's. The plan is made for the real files, the apply
/// writes them, the links stay, the owned-text record is the default folder's, and a revert puts
/// the bytes back.
#[test]
fn a_config_file_linked_into_the_default_folder_is_planned_and_written_there() {
    let f = earlier_home();
    orca_layout(&f);
    let profile = f.profile();
    let settings = f.default_dir().join("settings.json");
    let style = f.default_dir().join("output-styles/MoAI-Style.md");
    let p = plan(&f, Some(&profile));
    assert_eq!(action(&p, "style_file"), "create", "{p}");
    assert_eq!(action(&p, "setting"), "set", "{p}");
    assert_eq!(
        target(&p, "setting")["path"],
        json!(settings.to_str().unwrap()),
        "the plan names the file that is written: {p}"
    );
    assert_eq!(
        target(&p, "setting")["link"],
        json!(profile.join("settings.json").to_str().unwrap()),
        "{p}"
    );
    let (ok, applied) = apply(&f, Some(&profile), &p);
    assert!(ok, "{applied}");
    assert_eq!(applied["result"]["state"], "applied", "{applied}");
    let now: Value = serde_json::from_slice(&fs::read(&settings).unwrap()).unwrap();
    assert_eq!(now, json!({"a": 1, "outputStyle": "MoAI-Style"}));
    assert!(style.is_file());
    assert!(is_link_to(&profile.join("settings.json"), &settings));
    assert!(is_link_to(
        &profile.join("output-styles"),
        &f.default_dir().join("output-styles")
    ));
    let records = f.home.join(moai_cli::HOME_STATE_DIR).join("home-style");
    assert!(records.join("claude.json").is_file(), "the default record");
    let names: Vec<String> = fs::read_dir(&records)
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .filter(|n| n.starts_with("claude-"))
        .collect();
    assert!(
        names.is_empty(),
        "no record of the config folder: {names:?}"
    );
    // The same preview again has nothing to write.
    let again = plan(&f, Some(&profile));
    assert_eq!(action(&again, "style_file"), "unchanged", "{again}");
    assert_eq!(action(&again, "setting"), "unchanged", "{again}");
    // A revert from the same profile puts the default folder's bytes back; the links stay.
    let id = applied["result"]["receipt"].as_str().unwrap();
    let (ok, reverted) = op(
        &f,
        Some(&profile),
        &[
            "home-style-revert",
            f.home.to_str().unwrap(),
            "--receipt",
            id,
            "--permission-reference",
            PERM,
        ],
    );
    assert!(ok, "{reverted}");
    assert_eq!(reverted["result"]["state"], "reverted", "{reverted}");
    assert_eq!(fs::read_to_string(&settings).unwrap(), "{\"a\": 1}\n");
    assert!(!style.exists());
    assert!(is_link_to(&profile.join("settings.json"), &settings));
}

/// Only a link straight at a regular file inside the default folder is followed. Every other
/// shape keeps today's refusal, with a next step, and nothing is written through it.
#[test]
fn only_a_link_straight_to_a_regular_file_inside_the_default_folder_is_followed() {
    // (name, how the profile's settings.json link is made: the link target as a path)
    type Make = fn(&Fixture) -> PathBuf;
    // Every case that has an outside file keeps it at `outside/settings.json`: the loop below
    // reads it before and after, so a write through the link shows as a changed byte.
    let cases: [(&str, Make); 6] = [
        ("outside the default folder", |f| {
            write(&f.outside().join("settings.json"), "{\"keep\": 1}\n");
            // The default folder is there, so the refusal is the place of the target.
            fs::create_dir_all(f.default_dir()).unwrap();
            f.outside().join("settings.json")
        }),
        ("a relative link that leaves the default folder", |f| {
            write(&f.outside().join("settings.json"), "{\"keep\": 1}\n");
            fs::create_dir_all(f.default_dir()).unwrap();
            // From `home/profiles/p`: two folders up is the HOME, then `outside`.
            PathBuf::from("../../outside/settings.json")
        }),
        ("a link to a link", |f| {
            write(&f.outside().join("settings.json"), "{\"keep\": 1}\n");
            fs::create_dir_all(f.default_dir()).unwrap();
            symlink(
                f.outside().join("settings.json"),
                f.default_dir().join("other.json"),
            )
            .unwrap();
            f.default_dir().join("other.json")
        }),
        ("a link to nothing", |f| {
            fs::create_dir_all(f.default_dir()).unwrap();
            f.default_dir().join("missing.json")
        }),
        ("a link to a folder", |f| {
            fs::create_dir_all(f.default_dir().join("sub")).unwrap();
            f.default_dir().join("sub")
        }),
        (
            "an absolute link that leaves the default folder by ..",
            |f| {
                write(&f.outside().join("settings.json"), "{\"keep\": 1}\n");
                fs::create_dir_all(f.default_dir()).unwrap();
                f.default_dir().join("../outside/settings.json")
            },
        ),
    ];
    for (name, make) in cases {
        let f = current_home();
        let to = make(&f);
        fs::create_dir_all(f.profile()).unwrap();
        symlink(&to, f.profile().join("settings.json")).unwrap();
        let before = fs::read(f.outside().join("settings.json")).ok();
        let p = plan(&f, Some(&f.profile()));
        assert_eq!(action(&p, "setting"), "blocked", "{name}: {p}");
        let d = detail(&p, "setting");
        assert!(d.contains("settings.json"), "{name}: names the link: {d}");
        assert!(d.contains("next:"), "{name}: names a next step: {d}");
        let (ok, a) = apply(&f, Some(&f.profile()), &p);
        assert!(ok, "{name}: {a}");
        assert_eq!(
            fs::read(f.outside().join("settings.json")).ok(),
            before,
            "{name}: nothing is written through the link"
        );
        assert!(
            fs::read_to_string(f.profile().join("settings.json"))
                .map_or(true, |t| !t.contains("outputStyle")),
            "{name}"
        );
    }
    // A link inside the default folder itself is no redirect: it stays refused.
    let f = current_home();
    write(&f.default_dir().join("other.json"), "{}\n");
    symlink(
        f.default_dir().join("other.json"),
        f.default_dir().join("settings.json"),
    )
    .unwrap();
    let p = plan(&f, None);
    assert_eq!(action(&p, "setting"), "blocked", "{p}");
    assert!(detail(&p, "setting").contains("next:"), "{p}");
    // Control: a regular file of another name inside the default folder is followed.
    let f = current_home();
    write(&f.default_dir().join("work.json"), "{\"w\": 1}\n");
    fs::create_dir_all(f.default_dir().join("output-styles")).unwrap();
    fs::create_dir_all(f.profile()).unwrap();
    symlink(
        f.default_dir().join("work.json"),
        f.profile().join("settings.json"),
    )
    .unwrap();
    symlink(
        f.default_dir().join("output-styles"),
        f.profile().join("output-styles"),
    )
    .unwrap();
    let p = plan(&f, Some(&f.profile()));
    assert_eq!(action(&p, "setting"), "set", "{p}");
}

/// A relative link is read from the folder the link lies in: one that lands on a regular file
/// inside the default folder is followed (and stays relative), one that leaves it is refused (the
/// table above holds that case).
#[test]
fn a_relative_link_into_the_default_folder_is_followed_and_stays_relative() {
    let f = current_home();
    let settings = f.default_dir().join("settings.json");
    write(&settings, "{\"a\": 1}\n");
    fs::create_dir_all(f.default_dir().join("output-styles")).unwrap();
    fs::create_dir_all(f.profile()).unwrap();
    // From `home/profiles/p`: two folders up is the HOME, then `.claude`.
    let relative = PathBuf::from("../../.claude/settings.json");
    symlink(&relative, f.profile().join("settings.json")).unwrap();
    symlink(
        f.default_dir().join("output-styles"),
        f.profile().join("output-styles"),
    )
    .unwrap();
    let p = plan(&f, Some(&f.profile()));
    assert_eq!(action(&p, "setting"), "set", "{p}");
    assert_eq!(
        target(&p, "setting")["path"],
        json!(settings.to_str().unwrap()),
        "{p}"
    );
    let (ok, applied) = apply(&f, Some(&f.profile()), &p);
    assert!(ok, "{applied}");
    let now: Value = serde_json::from_slice(&fs::read(&settings).unwrap()).unwrap();
    assert_eq!(now, json!({"a": 1, "outputStyle": "MoAI-Style"}));
    assert_eq!(
        fs::read_link(f.profile().join("settings.json")).unwrap(),
        relative,
        "the link itself is not changed"
    );
}

/// The default folder is itself a link: a file link of a config folder is not followed through
/// it, and nothing is written.
#[test]
fn a_default_folder_that_is_a_link_blocks_a_linked_file() {
    let f = current_home();
    write(&f.outside().join("settings.json"), "{\"keep\": 1}\n");
    symlink(f.outside(), f.default_dir()).unwrap();
    fs::create_dir_all(f.profile()).unwrap();
    symlink(
        f.default_dir().join("settings.json"),
        f.profile().join("settings.json"),
    )
    .unwrap();
    let p = plan(&f, Some(&f.profile()));
    assert_eq!(action(&p, "setting"), "blocked", "{p}");
    let d = detail(&p, "setting");
    assert!(
        d.contains(f.default_dir().to_str().unwrap()) && d.contains("is a symlink"),
        "the detail names the default folder that is a link: {d}"
    );
    assert!(d.contains("next:"), "{d}");
    let (ok, a) = apply(&f, Some(&f.profile()), &p);
    assert!(ok, "{a}");
    assert_eq!(
        fs::read_to_string(f.outside().join("settings.json")).unwrap(),
        "{\"keep\": 1}\n",
        "nothing is written through the link of the default folder"
    );
}

/// A revert whose link was changed since the apply skips that file, names it as a file that is
/// not plain, and leaves the bytes behind the new link alone; the other target is reverted.
#[test]
fn a_revert_skips_a_linked_file_whose_link_changed_and_names_a_file() {
    let f = earlier_home();
    orca_layout(&f);
    let profile = f.profile();
    let settings = f.default_dir().join("settings.json");
    let p = plan(&f, Some(&profile));
    let (ok, applied) = apply(&f, Some(&profile), &p);
    assert!(ok, "{applied}");
    write(&f.outside().join("settings.json"), "{\"keep\": 1}\n");
    fs::remove_file(profile.join("settings.json")).unwrap();
    symlink(
        f.outside().join("settings.json"),
        profile.join("settings.json"),
    )
    .unwrap();
    let id = applied["result"]["receipt"].as_str().unwrap();
    let (ok, reverted) = op(
        &f,
        Some(&profile),
        &[
            "home-style-revert",
            f.home.to_str().unwrap(),
            "--receipt",
            id,
            "--permission-reference",
            PERM,
        ],
    );
    assert!(ok, "{reverted}");
    let skipped = reverted["result"]["skipped"].as_array().unwrap();
    let setting = skipped
        .iter()
        .find(|s| s["target"] == "claude:setting")
        .unwrap_or_else(|| panic!("the setting is not skipped: {reverted}"));
    assert!(
        setting["reason"]
            .as_str()
            .unwrap()
            .starts_with("file_not_plain: "),
        "a file link is a file that is not plain: {reverted}"
    );
    assert_eq!(
        fs::read_to_string(f.outside().join("settings.json")).unwrap(),
        "{\"keep\": 1}\n"
    );
    let still: Value = serde_json::from_slice(&fs::read(&settings).unwrap()).unwrap();
    assert_eq!(
        still["outputStyle"], "MoAI-Style",
        "the setting was not reverted"
    );
    assert!(
        reverted["result"]["reverted"]
            .as_array()
            .unwrap()
            .contains(&json!("claude:style_file")),
        "{reverted}"
    );
}

/// A link that is changed after the plan makes the apply refuse and write nothing.
#[test]
fn a_link_changed_after_the_plan_is_refused_at_write_time() {
    let f = current_home();
    orca_layout(&f);
    let settings = f.default_dir().join("settings.json");
    let p = plan(&f, Some(&f.profile()));
    assert_eq!(action(&p, "setting"), "set", "{p}");
    write(&f.outside().join("settings.json"), "{\"keep\": 1}\n");
    fs::remove_file(f.profile().join("settings.json")).unwrap();
    symlink(
        f.outside().join("settings.json"),
        f.profile().join("settings.json"),
    )
    .unwrap();
    let (ok, a) = apply(&f, Some(&f.profile()), &p);
    assert!(!ok, "{a}");
    assert_eq!(fs::read_to_string(&settings).unwrap(), "{\"a\": 1}\n");
    assert_eq!(
        fs::read_to_string(f.outside().join("settings.json")).unwrap(),
        "{\"keep\": 1}\n"
    );
    assert!(!f.default_dir().join("output-styles/MoAI-Style.md").exists());
}

/// The default folder's `settings.json` is read by the default folder's Claude Code too: it is
/// pointed at the style only when the style file is written in the default folder as well.
#[test]
fn a_linked_setting_waits_for_a_style_file_in_the_default_folder() {
    let f = current_home();
    let settings = f.default_dir().join("settings.json");
    write(&settings, "{\"a\": 1}\n");
    fs::create_dir_all(f.profile()).unwrap();
    symlink(&settings, f.profile().join("settings.json")).unwrap();
    // `output-styles` is a real folder of the profile: the style file is written there.
    let p = plan(&f, Some(&f.profile()));
    assert_eq!(action(&p, "style_file"), "create", "{p}");
    assert_eq!(action(&p, "setting"), "skipped", "{p}");
    assert!(detail(&p, "setting").contains("default folder"), "{p}");
    let (ok, a) = apply(&f, Some(&f.profile()), &p);
    assert!(ok, "{a}");
    assert_eq!(fs::read_to_string(&settings).unwrap(), "{\"a\": 1}\n");
    assert!(f.profile().join("output-styles/MoAI-Style.md").is_file());
    assert!(!f.default_dir().join("output-styles/MoAI-Style.md").exists());
}

/// A blocked style target names the file, the reason and one next step, in its detail and in a
/// notice (the `notices` of `moai update` come from `home_style::notices` over this plan).
#[test]
fn a_blocked_style_target_gives_a_reason_and_a_next_step_in_detail_and_notices() {
    let f = earlier_home();
    // `settings.json` links outside the default folder; `output-styles` is the profile's own
    // folder on a HOME whose state folder is the earlier generation's.
    write(&f.outside().join("settings.json"), "{\"keep\": 1}\n");
    fs::create_dir_all(f.profile()).unwrap();
    symlink(
        f.outside().join("settings.json"),
        f.profile().join("settings.json"),
    )
    .unwrap();
    let p = plan(&f, Some(&f.profile()));
    let targets = p["targets"].as_array().unwrap();
    let blocked: Vec<&Value> = targets
        .iter()
        .filter(|t| t["action"] == "blocked")
        .collect();
    assert_eq!(blocked.len(), 2, "both targets are blocked: {p}");
    let parsed: moai_cli::home_style::Plan = serde_json::from_value(p.clone()).unwrap();
    let project = f.home.join("project");
    fs::create_dir_all(&project).unwrap();
    let notices = moai_cli::home_style::notices(&parsed, &project);
    for t in blocked {
        let (path, why) = (t["path"].as_str().unwrap(), t["detail"].as_str().unwrap());
        assert!(why.contains("next:"), "the detail names a next step: {why}");
        let notice = notices
            .iter()
            .filter_map(|n| n["reason"].as_str())
            .find(|r| r.contains(path))
            .unwrap_or_else(|| panic!("no notice for {path}: {notices:?}"));
        assert!(notice.contains(why), "{notice}");
    }
    assert_eq!(
        fs::read_to_string(f.outside().join("settings.json")).unwrap(),
        "{\"keep\": 1}\n"
    );
}

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).unwrap();
    for entry in fs::read_dir(from).unwrap() {
        let p = entry.unwrap().path();
        let dest = to.join(p.file_name().unwrap());
        if p.is_dir() {
            copy_tree(&p, &dest);
        } else {
            fs::copy(&p, &dest).unwrap();
        }
    }
}

/// An installed copy of the binary with the real package and agent sources beside it, below the
/// temporary HOME's state folder (the way `moai update` finds its package).
fn installed_copy(f: &Fixture) -> PathBuf {
    let folder = moai_cli::home_root::path(&f.home).join("versions/a");
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    copy_tree(&repo.join("skills/moai"), &folder.join("skills/moai"));
    copy_tree(&repo.join("agents"), &folder.join("agents"));
    exe
}

/// `moai update` (a preview) of the installed copy with `CLAUDE_CONFIG_DIR` naming the profile;
/// returns the one `claude_folders` notice.
fn folders_notice_of_update(f: &Fixture, exe: &Path) -> Value {
    let out = Command::new(exe)
        .arg("update")
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &f.home)
        .env("MOAI_ACCOUNT_HOME", &f.home)
        .env("LC_ALL", "C")
        .env("CLAUDE_CONFIG_DIR", f.profile())
        .current_dir(&f.home)
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    let all: Vec<Value> = v["result"]["notices"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|n| n["kind"] == json!("claude_folders"))
        .cloned()
        .collect();
    assert_eq!(all.len(), 1, "one claude_folders notice: {v}");
    all[0].clone()
}

/// SPEC-20261127-update-linked-profile-display AC-4 (REQ-7, REQ-8): when every entry of the
/// profile that the update writes is a link into the default folder, the notice names the default
/// folder's paths and does not warn that the update writes into the profile; one entry that is a
/// real folder in the profile names that place and keeps the warning.
#[test]
fn a_fully_linked_profile_is_shown_as_the_default_folder() {
    let f = fixture(|_| {});
    let exe = installed_copy(&f);
    orca_layout(&f);
    for entry in ["skills", "agents", "rules"] {
        fs::create_dir_all(f.default_dir().join(entry)).unwrap();
        symlink(f.default_dir().join(entry), f.profile().join(entry)).unwrap();
    }
    let default = f.default_dir();
    let profile = f.profile();

    let n = folders_notice_of_update(&f, &exe);
    let shown = n.to_string();
    assert_eq!(n["warning"], json!(false), "{n}");
    for part in ["skills", "agents", "rules"] {
        assert_eq!(
            n["folders"][part],
            json!(default.to_str().unwrap()),
            "{part}: {n}"
        );
    }
    assert_eq!(
        n["folders"]["style"],
        json!(default.to_str().unwrap()),
        "{n}"
    );
    assert!(
        !shown.contains(profile.to_str().unwrap()),
        "no profile path: {n}"
    );
    assert!(shown.contains(default.to_str().unwrap()), "{n}");
    let reason = n["reason"].as_str().unwrap();
    assert!(!reason.contains("CLAUDE_CONFIG_DIR names"), "{reason}");

    // One entry that is a real folder in the profile: that place and the warning.
    fs::remove_file(profile.join("skills")).unwrap();
    fs::create_dir_all(profile.join("skills")).unwrap();
    let n = folders_notice_of_update(&f, &exe);
    assert_eq!(n["warning"], json!(true), "{n}");
    assert_eq!(
        n["folders"]["skills"],
        json!(profile.to_str().unwrap()),
        "{n}"
    );
    let reason = n["reason"].as_str().unwrap();
    assert!(reason.contains("CLAUDE_CONFIG_DIR names"), "{reason}");
    assert!(reason.contains(profile.to_str().unwrap()), "{reason}");
}

/// The keys destination comes from the one HOME state folder function: below the earlier state
/// folder while the HOME has no marker of this generation (the move belongs to the open card that merges the two state folders), below the
/// current one once it has. The switch link folder comes from the same function
/// (`src/update/switch.rs`, pinned by `update_switch_cli`).
#[test]
fn keys_follow_the_home_state_folder() {
    let scope = moai_cli::key_vault::Scope::Shared;
    let f = earlier_home();
    let dest = moai_cli::key_vault::entry_path(&f.home, &scope, "glm");
    assert!(
        dest.starts_with(f.home.join(moai_cli::HOME_STATE_DIR)),
        "{dest:?}"
    );
    let f = current_home();
    let dest = moai_cli::key_vault::entry_path(&f.home, &scope, "glm");
    assert!(
        dest.starts_with(f.home.join(moai_cli::home_root::MOAI_DIR)),
        "{dest:?}"
    );
}
