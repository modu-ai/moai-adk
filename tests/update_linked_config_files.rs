//! SPEC-20261124-update-linked-config-targets: `moai update` plans a Claude config file that is a
//! link into the default folder, and a blocked style target says why and what to do next. Every
//! test runs the binary in a temporary HOME: the HOME, the account home and the config folder are
//! given to the child only; the real `~/.claude`, the Orca profile and the real HOME are never
//! read or written.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    os::unix::fs::{symlink, PermissionsExt},
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const PERM: &str = "test-permission";
/// The switch script the update step knows (a preserved copy of the v2 SPEC).
const SCRIPT: &[u8] =
    include_bytes!("../.moai/specs/SPEC-20261119-switch-script-v2/preview/moai.v1");

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
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

struct Fixture {
    _tmp: TempDir,
    home: PathBuf,
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

/// A temporary HOME on the earlier generation's state folder (`~/.moai-cli`, no marker).
fn earlier_home() -> Fixture {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    Fixture { _tmp: tmp, home }
}

/// A temporary HOME whose `~/.moai` is this generation's (the marker is there).
fn current_home() -> Fixture {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(&home).unwrap();
    assert!(moai_cli::home_root::adopt_new(&home, "test").unwrap());
    Fixture { _tmp: tmp, home }
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
    let (h, p) = (f.home.to_str().unwrap(), package());
    let (ok, v) = op(f, folder, &["home-style-plan", h, p.to_str().unwrap()]);
    assert!(ok, "{v}");
    v["result"].clone()
}

fn apply(f: &Fixture, folder: Option<&Path>, plan: &Value) -> (bool, Value) {
    let (h, p) = (f.home.to_str().unwrap(), package());
    let digest = plan["plan_digest"].as_str().unwrap();
    op(
        f,
        folder,
        &[
            "home-style-apply",
            h,
            p.to_str().unwrap(),
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
    let style = f.default_dir().join("output-styles/MoAI-Easy.md");
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
    assert_eq!(now, json!({"a": 1, "outputStyle": "MoAI-Easy"}));
    assert!(style.is_file());
    assert!(is_link_to(&profile.join("settings.json"), &settings));
    assert!(is_link_to(
        &profile.join("output-styles"),
        &f.default_dir().join("output-styles")
    ));
    let records = f.home.join(".moai-cli/home-style");
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
    let cases: [(&str, Make); 5] = [
        ("outside the default folder", |f| {
            write(&f.outside().join("settings.json"), "{\"keep\": 1}\n");
            f.outside().join("settings.json")
        }),
        ("a link to a link", |f| {
            write(&f.outside().join("real.json"), "{\"keep\": 1}\n");
            write(&f.default_dir().join("other.json"), "{}\n");
            fs::remove_file(f.default_dir().join("other.json")).unwrap();
            symlink(
                f.outside().join("real.json"),
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
    assert!(!f.default_dir().join("output-styles/MoAI-Easy.md").exists());
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
    assert!(f.profile().join("output-styles/MoAI-Easy.md").is_file());
    assert!(!f.default_dir().join("output-styles/MoAI-Easy.md").exists());
}

/// A version folder with this build, the package and the agent sources; `state` is the HOME
/// state folder name the folder sits in.
fn install(f: &Fixture, state: &str) -> PathBuf {
    let folder = f.home.join(state).join("versions/a");
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    copy_tree(&repo.join("skills/moai"), &folder.join("skills/moai"));
    copy_tree(&repo.join("agents"), &folder.join("agents"));
    exe
}

fn update_preview(f: &Fixture, exe: &Path, folder: Option<&Path>) -> Value {
    let work = f.home.parent().unwrap().join("work");
    fs::create_dir_all(&work).unwrap();
    let out: Output = Command::new(exe)
        .arg("update")
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &f.home)
        .env("MOAI_ACCOUNT_HOME", &f.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_CEILING_DIRECTORIES", work.parent().unwrap())
        .env("LC_ALL", "C")
        .envs(folder.map(|p| ("CLAUDE_CONFIG_DIR", p.to_path_buf())))
        .current_dir(&work)
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// A blocked style target names the file, the reason and one next step, in its detail and in a
/// notice of the update preview.
#[test]
fn a_blocked_style_target_gives_a_reason_and_a_next_step_in_detail_and_notices() {
    let f = earlier_home();
    let exe = install(&f, ".moai-cli");
    // `settings.json` links outside the default folder; `output-styles` is the profile's own
    // folder on a HOME whose state folder is the earlier generation's.
    write(&f.outside().join("settings.json"), "{\"keep\": 1}\n");
    fs::create_dir_all(f.profile()).unwrap();
    symlink(
        f.outside().join("settings.json"),
        f.profile().join("settings.json"),
    )
    .unwrap();
    let p = update_preview(&f, &exe, Some(&f.profile()));
    let targets = p["style"]["targets"].as_array().unwrap();
    let blocked: Vec<&Value> = targets
        .iter()
        .filter(|t| t["action"] == "blocked")
        .collect();
    assert_eq!(blocked.len(), 2, "both targets are blocked: {p}");
    let notices = p["notices"].as_array().unwrap();
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

/// `keys.destination` and `switch.file` come from the one HOME state folder function: below
/// `~/.moai-cli` while the HOME has no marker of this generation (the move is card C-199's), below
/// `~/.moai` once it has.
#[test]
fn keys_and_switch_follow_the_home_state_folder() {
    for (state, f) in [(".moai-cli", earlier_home()), (".moai", current_home())] {
        let exe = install(&f, state);
        let script = f.home.join(state).join("path/moai");
        write(&script, "");
        fs::write(&script, SCRIPT).unwrap();
        fs::set_permissions(&script, fs::Permissions::from_mode(0o755)).unwrap();
        symlink(
            "../versions/a/bin/moai-cli",
            script.with_file_name("moai-cli"),
        )
        .unwrap();
        // A legacy key file makes the keys layer say where its vault entry would go.
        let key = f.home.join(".moai-cli/.env.glm");
        write(&key, "GLM_API_KEY=DUMMY-GLM-KEY-123\n");
        fs::set_permissions(&key, fs::Permissions::from_mode(0o600)).unwrap();
        let p = update_preview(&f, &exe, None);
        let root = f.home.join(state);
        let dest = PathBuf::from(
            p["keys"]["destination"]
                .as_str()
                .unwrap_or_else(|| panic!("{state}: no keys destination in {p}")),
        );
        assert!(dest.starts_with(&root), "{state}: keys {dest:?}");
        let file = PathBuf::from(
            p["switch"]["file"]
                .as_str()
                .unwrap_or_else(|| panic!("{state}: no switch file in {p}")),
        );
        assert!(file.starts_with(&root), "{state}: switch {file:?}");
    }
}
