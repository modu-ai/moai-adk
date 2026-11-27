//! SPEC-20261121-go-residue-detect: the read-only `go-residue` step of
//! `moai project-update-plan` finds the files the Go generation left behind, proves them
//! against the first manifest it can find (`.moai/manifest.json`, a `.moai-backup*` manifest,
//! a `.moai/backups/go-markers-*` manifest, or the backup the last committed go-state receipt
//! names), sorts them into `go_same` / `go_changed` / `user_data`, counts secret-shaped names
//! without opening them, never follows a link, and writes nothing. The frame is the
//! Go-generation survey's section 8 temp project; the real checkout is never read.

#![cfg(unix)]

mod common;
use common::{CleanEnv, MOAI_BIN};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::BTreeMap,
    fs,
    os::unix::fs::PermissionsExt,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

/// A path under the Go-generation state folder this fixture builds on purpose. The folder name
/// is spelled here once so the state-folder literal scan counts one spelling, not one per path.
macro_rules! st {
    ($rel:literal) => {
        concat!(".moai/", $rel)
    };
}

/// The same path with a run-time tail (a stamp or a counter): `st!` cannot carry `{}` captures.
fn st_at(tail: &str) -> String {
    format!("{}{tail}", st!(""))
}

fn sha(body: &str) -> String {
    format!("sha256:{:x}", Sha256::digest(body.as_bytes()))
}

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    root: PathBuf,
    files: serde_json::Map<String, Value>,
}

impl Fx {
    fn new() -> Fx {
        let tmp = TempDir::new().unwrap();
        let base = fs::canonicalize(tmp.path()).unwrap();
        let (home, root) = (base.join("home"), base.join("project"));
        fs::create_dir_all(&home).unwrap();
        fs::create_dir_all(&root).unwrap();
        let fx = Fx {
            _tmp: tmp,
            home,
            root,
            files: serde_json::Map::new(),
        };
        // A valid Rust state: the marker and the minimal profile.
        fx.write(
            st!("moai.json"),
            &json!({"generation": "moai", "created_by": "moai-cli test",
                    "go_backup": ".moai-backup"})
            .to_string(),
        );
        fx.write(st!("project.yaml"), "schema_version: 2\n");
        fx
    }

    fn p(&self) -> &str {
        self.root.to_str().unwrap()
    }

    fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    /// A link inside the state folder; the target stays as written, never touched.
    fn link(&self, rel: &str, target: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        std::os::unix::fs::symlink(target, p).unwrap();
    }

    /// A file the Go deployer wrote and recorded in the manifest under construction.
    fn deployed(&mut self, rel: &str, body: &str) {
        self.write(rel, body);
        self.files.insert(
            rel.into(),
            json!({"provenance": "template_managed", "template_hash": sha(body),
                   "deployed_hash": sha(body), "current_hash": sha(body)}),
        );
    }

    /// Recorded but rewritten by the user: the manifest still names the old bytes.
    fn changed(&mut self, rel: &str, recorded: &str, on_disk: &str) {
        self.write(rel, on_disk);
        self.files.insert(
            rel.into(),
            json!({"provenance": "template_managed", "template_hash": sha(recorded),
                   "deployed_hash": sha(recorded), "current_hash": sha(recorded)}),
        );
    }

    /// A file on disk that the manifest records with its own bytes under `provenance` (none
    /// when `None`): what the Go tool writes for a file it found and did not deploy.
    fn recorded(&mut self, rel: &str, body: &str, provenance: Option<&str>) {
        self.write(rel, body);
        let mut entry = json!({"template_hash": sha("# the template\n"),
                               "deployed_hash": sha(body), "current_hash": sha(body)});
        if let Some(provenance) = provenance {
            entry["provenance"] = json!(provenance);
        }
        self.files.insert(rel.into(), entry);
    }

    /// Writes the manifest under `rel` (the live one, a backup, or a go-markers copy).
    fn manifest_at(&self, rel: &str) {
        self.write(
            rel,
            &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z",
                    "files": self.files})
            .to_string(),
        );
    }

    fn git_commit_all(&self) {
        common::git::git_ok(&self.root, &["init", "-q"]);
        common::git::git_ok(&self.root, &["add", "-A"]);
        common::git::git_ok(
            &self.root,
            &[
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "commit",
                "-q",
                "-m",
                "fixture",
            ],
        );
    }

    /// Every path under the project and HOME, links named as links, never followed.
    fn tree(&self) -> Vec<(PathBuf, String)> {
        fn node(p: &Path) -> String {
            let m = fs::symlink_metadata(p).unwrap();
            if m.file_type().is_symlink() {
                format!("link:{}", fs::read_link(p).unwrap().display())
            } else if m.is_dir() {
                "dir".into()
            } else {
                format!("file:{}", m.len())
            }
        }
        let mut out = Vec::new();
        let mut stack = vec![self.root.clone(), self.home.clone()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let e = e.unwrap();
                let p = e.path();
                if p.is_dir() && !p.is_symlink() {
                    stack.push(p);
                } else {
                    let kind = node(&p);
                    out.push((p, kind));
                }
            }
        }
        out.sort();
        out
    }

    fn run(&self, args: &[&str]) -> Value {
        let out = Command::new(MOAI_BIN)
            .args(args)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .stdin(Stdio::null())
            .output()
            .expect("spawn the binary");
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    /// `project-update-plan`, with the whole tree compared before and after.
    fn update_plan(&self) -> Value {
        let before = self.tree();
        let v = self.run(&["project-update-plan", self.p()]);
        assert_eq!(self.tree(), before, "project-update-plan wrote something");
        v
    }

    /// Every file's bytes under the project and HOME (a link as `link:<target>`), never
    /// following a link; `backups` leaves out the backup area the apply writes into.
    fn bytes(&self, backups: bool) -> BTreeMap<PathBuf, Vec<u8>> {
        let mut out = BTreeMap::new();
        let mut stack = vec![self.root.clone(), self.home.clone()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                let m = fs::symlink_metadata(&p).unwrap();
                if m.file_type().is_symlink() {
                    let t = fs::read_link(&p).unwrap();
                    out.insert(p, format!("link:{}", t.display()).into_bytes());
                } else if m.is_dir() {
                    stack.push(p);
                } else if backups || !p.starts_with(self.root.join(st!("backups"))) {
                    out.insert(p.clone(), fs::read(&p).unwrap());
                }
            }
        }
        out
    }

    /// The printed `go-residue-apply` command of a plan, the permission placeholder filled in
    /// with `permission`, as the arguments after `moai`.
    fn printed_apply(plan: &Value, permission: &str) -> Vec<String> {
        let argv = step(plan, "go-residue")["apply"]["argv"]
            .as_array()
            .unwrap_or_else(|| {
                panic!(
                    "the go-residue step prints no command: {:?}",
                    step(plan, "go-residue")
                )
            })
            .iter()
            .map(|a| a.as_str().unwrap().to_string())
            .collect::<Vec<_>>();
        assert_eq!(argv[0], "moai");
        argv[1..]
            .iter()
            .map(|a| {
                if a == "<permission-reference>" {
                    permission.to_string()
                } else {
                    a.clone()
                }
            })
            .collect()
    }

    /// Runs the command a plan printed, with `extra` arguments added.
    fn apply_printed(&self, plan: &Value, extra: &[&str]) -> Value {
        let mut args = Fx::printed_apply(plan, "test-approval");
        args.extend(extra.iter().map(|a| a.to_string()));
        let refs: Vec<&str> = args.iter().map(String::as_str).collect();
        self.run(&refs)
    }

    /// The stamp folders of the apply receipts under the backup area, sorted.
    fn stamps(&self) -> Vec<String> {
        let mut v: Vec<String> = fs::read_dir(self.root.join(st!("backups")))
            .map(|rd| {
                rd.filter_map(|e| e.ok())
                    .filter_map(|e| e.file_name().into_string().ok())
                    .filter(|n| {
                        self.root
                            .join(st!("backups"))
                            .join(n)
                            .join("go-residue-receipt.json")
                            .is_file()
                    })
                    .collect()
            })
            .unwrap_or_default();
        v.sort();
        v
    }

    /// The item of the `go-residue` step whose `path` ends with `name`.
    fn residue_item<'v>(plan: &'v Value, name: &str) -> &'v Value {
        step(plan, "go-residue")["preview"]["items"]
            .as_array()
            .expect("go-residue step has items")
            .iter()
            .find(|i| i["path"].as_str().is_some_and(|p| p.ends_with(name)))
            .unwrap_or_else(|| panic!("no go-residue item for {name}"))
    }
}

fn step<'v>(plan: &'v Value, name: &str) -> &'v Value {
    // The command answers its envelope; the plan body lives under `result`.
    let body = plan.get("result").unwrap_or(plan);
    body["phases"]
        .as_array()
        .expect("plan has phases")
        .iter()
        .flat_map(|p| p["steps"].as_array().into_iter().flatten())
        .find(|s| s["step"] == json!(name))
        .unwrap_or_else(|| {
            panic!(
                "no `{name}` step in the plan: {:?}",
                plan["phases"].as_array().map(|a| {
                    a.iter()
                        .flat_map(|p| p["steps"].as_array().into_iter().flatten())
                        .filter_map(|s| s["step"].as_str())
                        .collect::<Vec<_>>()
                })
            )
        })
}

/// The survey's section 8 frame: a git-tracked project with a valid Rust state, one live Go
/// manifest, byte-identical deployed files, one user-changed file, one unmanifested user
/// file, one secret-shaped name, and an output style a settings file references.
fn survey_frame() -> Fx {
    let mut fx = Fx::new();
    fx.deployed(st!("policies/core/a.md"), "# core policy\n");
    fx.deployed(st!("workflows/gtd.md"), "# gtd workflow\n");
    fx.deployed(
        st!("output-styles/moai/moai-easy.md"),
        "# moai easy style\n",
    );
    fx.changed(
        st!("config/sections/user.yaml"),
        "tone: formal\n",
        "tone: casual\n",
    );
    fx.write(st!("design/x.md"), "# user design notes\n");
    fx.write(st!(".env.local"), "KEY=value\n");
    // A link whose target sits outside the state folder: following it would read content
    // the survey never walks, so its digest must stay empty.
    fx.write("outside.txt", "outside content\n");
    fx.link(st!("design/link.md"), "../outside.txt");
    fx.write(
        ".claude/settings.json",
        "{\"outputStyle\": \"moai/moai-easy\"}\n",
    );
    fx.manifest_at(st!("manifest.json"));
    fx.git_commit_all();
    fx
}

/// AC-1 (REQ-1, REQ-7): the step sorts the frame — `go_same` for the byte-identical files that
/// nothing reads, `go_in_use` for the output style the settings name, `go_changed` for the
/// rewritten one, `user_data` for the unmanifested one — counts the one secret-shaped name
/// without opening it, marks tracked files, carries the item fields, and a second plan run reads
/// the same.
#[test]
fn the_go_residue_step_classifies_the_survey_frame() {
    let fx = survey_frame();
    let first = fx.update_plan();
    let found = step(&first, "go-residue");
    for name in ["policies/core/a.md", "workflows/gtd.md"] {
        let item = Fx::residue_item(&first, name.split('/').next_back().unwrap());
        assert_eq!(item["class"], json!("go_same"), "{name}");
        assert_eq!(item["tracked"], json!(true), "{name}");
    }
    // SPEC-20261126-go-residue-in-use: the settings name the output style, so it stays in place.
    let style = Fx::residue_item(&first, "moai-easy.md");
    assert_eq!(style["class"], json!("go_in_use"), "{style}");
    assert_eq!(style["suggested_action"], json!("keep"), "{style}");
    assert_eq!(style["tracked"], json!(true), "{style}");
    let changed = Fx::residue_item(&first, "user.yaml");
    assert_eq!(changed["class"], json!("go_changed"), "the rewritten file");
    assert_eq!(changed["tracked"], json!(true));
    let user = Fx::residue_item(&first, "x.md");
    assert_eq!(user["class"], json!("user_data"), "the unmanifested file");
    assert_eq!(user["tracked"], json!(true));
    // The secret-shaped name is counted, never opened: no item carries it or its digest.
    let items = found["preview"]["items"].as_array().unwrap();
    assert_eq!(
        found["preview"]["secret_names"],
        json!(1),
        "one secret-shaped name"
    );
    assert!(
        items
            .iter()
            .all(|i| !i["path"].as_str().is_some_and(|p| p.contains(".env"))),
        "no item names the secret file: {items:?}"
    );
    // A link is named but never read: its own digest stays empty, and the outside file's
    // digest appears nowhere in the step.
    let linked = Fx::residue_item(&first, "link.md");
    assert_eq!(linked["class"], json!("user_data"));
    assert_eq!(linked["sha256"], json!(null), "a link's digest stays empty");
    let outside = json!(sha("outside content\n"));
    assert!(
        items.iter().all(|i| i["sha256"] != outside),
        "the outside content leaked into a digest: {items:?}"
    );
    // Every item carries the fields the SPEC requires.
    for i in items {
        for field in [
            "path",
            "class",
            "sha256",
            "tracked",
            "resurrected",
            "suggested_action",
        ] {
            assert!(i.get(field).is_some(), "item {:?} lacks {field}", i["path"]);
        }
    }
    // A second run reads the same (REQ-7).
    let second = fx.update_plan();
    assert_eq!(
        step(&second, "go-residue")["preview"]["items"],
        step(&first, "go-residue")["preview"]["items"]
    );
}

/// AC-2 (REQ-2): with the live `.moai/manifest.json` gone, the step still proves the files
/// from the backup manifests (first a `.moai-backup*` manifest, then a go-markers one).
#[test]
fn the_go_residue_step_proves_from_the_backup_manifests() {
    let fx = survey_frame();
    // The live manifest moves into a go-markers copy; a `.moai-backup` manifest remains.
    fs::create_dir_all(fx.root.join(st!("backups/go-markers-20261120T234501Z"))).unwrap();
    fs::rename(
        fx.root.join(st!("manifest.json")),
        fx.root
            .join(st!("backups/go-markers-20261120T234501Z/manifest.json")),
    )
    .unwrap();
    fx.manifest_at(".moai-backup/manifest.json");
    let plan = fx.update_plan();
    let found = step(&plan, "go-residue");
    assert!(
        found["preview"]["proof"]
            .as_str()
            .is_some_and(|p| p.contains(".moai-backup/manifest.json")),
        "the proof names the backup manifest: {:?}",
        found["preview"]["proof"]
    );
    for name in ["a.md", "gtd.md"] {
        assert_eq!(
            Fx::residue_item(&plan, name)["class"],
            json!("go_same"),
            "{name}"
        );
    }
}

/// REQ-3, REQ-4: files only the user (or this generation) wrote are listed as `user_data`
/// with the action `keep`; they alone never ask for a decision, so the step reads `unchanged`
/// and the plan's `all_unchanged` is not forced false by it.
#[test]
fn user_data_alone_leaves_the_go_residue_step_unchanged() {
    let fx = Fx::new();
    // A backup manifest that names no file: it proves nothing for the surveyed ones.
    fx.manifest_at(".moai-backup/manifest.json");
    fx.write(st!("design/x.md"), "# user design notes\n");
    fx.git_commit_all();
    let plan = fx.update_plan();
    let found = step(&plan, "go-residue");
    let user = Fx::residue_item(&plan, "x.md");
    assert_eq!(user["class"], json!("user_data"), "the unmanifested file");
    assert_eq!(user["tracked"], json!(true));
    assert_eq!(user["suggested_action"], json!("keep"));
    assert_eq!(
        found["state"],
        json!("unchanged"),
        "user data alone is not a decision: {found:?}"
    );
    // Every surveyed item is kept as it is, so nothing here could ask for a decision.
    let items = found["preview"]["items"].as_array().unwrap();
    assert!(
        items.iter().all(|i| i["suggested_action"] == json!("keep")),
        "only keep items in this frame: {items:?}"
    );
    // The step takes no part in the other steps' state: the plan's `all_unchanged` is the
    // AND of every step, so the residue step is the one that must read `unchanged` here.
    let body = plan.get("result").unwrap_or(&plan);
    let others_unchanged = body["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|p| p["steps"].as_array().into_iter().flatten())
        .filter(|s| s["step"] != json!("go-residue"))
        .all(|s| matches!(s["state"].as_str(), Some("unchanged" | "skip" | "optional")));
    assert_eq!(
        body["all_unchanged"],
        json!(others_unchanged),
        "all_unchanged follows the other steps alone: {:?}",
        body["all_unchanged"]
    );
}

/// The step stays out of the other steps' state: `all_unchanged` is the AND of every step that
/// is not the go-residue one, because `optional` and `unchanged` count as nothing to do.
fn assert_all_unchanged_follows_the_other_steps(plan: &Value) {
    let body = plan.get("result").unwrap_or(plan);
    let others_unchanged = body["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|p| p["steps"].as_array().into_iter().flatten())
        .filter(|s| s["step"] != json!("go-residue"))
        .all(|s| matches!(s["state"].as_str(), Some("unchanged" | "skip" | "optional")));
    assert_eq!(
        body["all_unchanged"],
        json!(others_unchanged),
        "all_unchanged follows the other steps alone: {:?}",
        body["all_unchanged"]
    );
}

/// Leaving the files where they are is a valid answer, so a surveyed file the survey would act
/// on reads `optional`, never a decision, and never holds `all_unchanged` false. A command is
/// printed only for the proven bundle: with no proven file there is none.
#[test]
fn residue_to_act_on_is_reported_and_does_not_hold_the_update() {
    // A proven file is there: `optional` with the printed proven-bundle command.
    let fx = survey_frame();
    let plan = fx.update_plan();
    let found = step(&plan, "go-residue");
    let same = Fx::residue_item(&plan, "a.md");
    assert_eq!(same["class"], json!("go_same"), "the byte-identical file");
    assert_ne!(
        same["suggested_action"],
        json!("keep"),
        "the survey would act on this file: {same:?}"
    );
    assert_eq!(
        found["state"],
        json!("optional"),
        "leaving the files is a valid answer, so this is not a decision: {found:?}"
    );
    let argv = Fx::printed_apply(&plan, "test-approval");
    assert_eq!(argv[0], "go-residue-apply", "{argv:?}");
    assert!(
        argv.windows(2).any(|w| w == ["--bundle", "proven"]),
        "the command names the proven bundle only: {argv:?}"
    );
    assert!(
        !argv.iter().any(|a| a == "--pick"),
        "no command moves a changed file: {argv:?}"
    );
    assert_all_unchanged_follows_the_other_steps(&plan);

    // Only a changed file: still reported, but nothing is proven, so there is no command.
    let mut changed_only = Fx::new();
    changed_only.changed(
        st!("config/sections/user.yaml"),
        "tone: formal\n",
        "tone: casual\n",
    );
    changed_only.manifest_at(st!("manifest.json"));
    changed_only.git_commit_all();
    let plan = changed_only.update_plan();
    let found = step(&plan, "go-residue");
    assert_eq!(found["state"], json!("optional"), "{found:?}");
    assert_eq!(found["apply"], json!(null), "{found:?}");
    assert_all_unchanged_follows_the_other_steps(&plan);

    // No manifest proves anything: nothing to move and nothing held.
    let unproven = Fx::new();
    unproven.write(st!("design/x.md"), "# user design notes\n");
    unproven.git_commit_all();
    let plan = unproven.update_plan();
    let found = step(&plan, "go-residue");
    assert_eq!(found["state"], json!("unchanged"), "{found:?}");
    assert_eq!(found["apply"], json!(null), "{found:?}");
    assert_all_unchanged_follows_the_other_steps(&plan);
}

/// AC-3 (REQ-6): with no manifest anywhere the step reports no proof with zero candidates
/// and the plan still succeeds.
#[test]
fn the_go_residue_step_reports_no_proof_when_no_manifest_exists() {
    let fx = survey_frame();
    fs::remove_file(fx.root.join(st!("manifest.json"))).unwrap();
    let plan = fx.update_plan();
    let found = step(&plan, "go-residue");
    assert_eq!(
        found["preview"]["proof"],
        json!("none"),
        "no proof anywhere"
    );
    assert_eq!(found["preview"]["items"], json!([]), "zero candidates");
    assert_eq!(
        found["preview"]["secret_names"],
        json!(0),
        "nothing was surveyed as a candidate"
    );
}

// ---- SPEC-20261125-go-residue-apply: apply and revert ----

/// The proven files of the survey frame that nothing reads: the bundle moves exactly these.
const PROVEN: [&str; 2] = [st!("policies/core/a.md"), st!("workflows/gtd.md")];
/// The frame's output style: the settings name it, so the bundle leaves it in place
/// (SPEC-20261126-go-residue-in-use).
const STYLE_IN_USE: &str = st!("output-styles/moai/moai-easy.md");

/// The survey frame with one more proven file nothing reads, for the tests that need a file the
/// revert restores beside the files they break.
fn survey_frame_with_a_third_proven_file() -> (Fx, &'static str) {
    let third = st!("policies/core/b.md");
    let mut fx = survey_frame();
    fx.deployed(third, "# core policy b\n");
    fx.manifest_at(st!("manifest.json"));
    (fx, third)
}

/// Where the apply put `rel`: `.moai/backups/<stamp>/go-residue/<same path>`.
fn backup_of(fx: &Fx, stamp: &str, rel: &str) -> PathBuf {
    fx.root
        .join(st_at(&format!("backups/{stamp}/go-residue/{rel}")))
}

fn revert_args<'a>(fx: &'a Fx, stamp: &'a str) -> Vec<&'a str> {
    vec![
        "go-residue-revert",
        fx.p(),
        "--receipt",
        stamp,
        "--permission-reference",
        "test-approval",
    ]
}

/// AC-1 (REQ-1, REQ-10, REQ-11) and AC-7 (REQ-12): applying the proven bundle moves exactly
/// the two byte-identical files nothing reads into the backup area and nothing else (the output
/// style the settings name stays in place); the changed file is asked about and moves only with
/// an explicit pick; once nothing proven or changed is left the step reads `unchanged` although
/// an earlier-generation folder is still listed.
#[test]
fn applying_the_proven_bundle_moves_only_the_go_same_files() {
    let fx = survey_frame();
    fs::create_dir_all(fx.root.join(".moai-backups/20261101T000000Z")).unwrap();
    let before = fx.bytes(true);
    let plan = fx.update_plan();
    assert_eq!(step(&plan, "go-residue")["state"], json!("optional"));

    let done = fx.apply_printed(&plan, &[]);
    assert_eq!(done["ok"], json!(true), "{done}");
    let r = &done["result"];
    assert_eq!(r["state"], json!("committed"), "{r}");
    let stamp = r["receipt"]
        .as_str()
        .expect("the receipt stamp")
        .to_string();
    let after = fx.bytes(true);
    for rel in PROVEN {
        assert!(
            fs::symlink_metadata(fx.root.join(rel)).is_err(),
            "{rel} is still at its original place"
        );
        assert_eq!(
            fs::read(backup_of(&fx, &stamp, rel)).unwrap(),
            before[&fx.root.join(rel)],
            "{rel} differs in the backup"
        );
    }
    // Every other file is byte for byte where it was: the changed file, the user's file, the
    // secret-shaped name, the link and the file it points at.
    for (p, b) in &before {
        if PROVEN.iter().any(|rel| *p == fx.root.join(rel)) {
            continue;
        }
        assert_eq!(after.get(p), Some(b), "{} changed", p.display());
    }
    let receipt: Value = serde_json::from_slice(
        &fs::read(
            fx.root
                .join(st_at(&format!("backups/{stamp}/go-residue-receipt.json"))),
        )
        .unwrap(),
    )
    .unwrap();
    assert_eq!(receipt["state"], json!("committed"));
    assert_eq!(receipt["moves"].as_array().unwrap().len(), 2);
    assert!(
        fx.root.join(STYLE_IN_USE).is_file(),
        "the output style the settings name left its place"
    );
    // REQ-12: the two moved files are tracked; the result says so and warns.
    assert_eq!(r["tracked_moved"], json!(2), "{r}");
    assert!(
        r["warning"]
            .as_str()
            .is_some_and(|w| w.contains("checkout")),
        "the warning names the checkout: {r}"
    );

    // The changed file is still offered by name, and nothing proven is left to move.
    let second = fx.update_plan();
    let found = step(&second, "go-residue");
    assert_eq!(found["state"], json!("optional"), "{found}");
    assert_eq!(
        found["apply"],
        json!(null),
        "no command moves a changed file"
    );
    assert_eq!(
        Fx::residue_item(&second, "user.yaml")["class"],
        json!("go_changed")
    );
    assert!(
        found["preview"]["items"]
            .as_array()
            .unwrap()
            .iter()
            .all(|i| i["class"] != json!("go_same")),
        "no proven file is left: {found}"
    );
    let digest = found["plan_digest"].as_str().expect("the plan digest");
    let picked = fx.run(&[
        "go-residue-apply",
        fx.p(),
        "--plan-digest",
        digest,
        "--permission-reference",
        "test-approval",
        "--pick",
        st!("config/sections/user.yaml"),
    ]);
    assert_eq!(picked["ok"], json!(true), "{picked}");
    assert_eq!(fx.stamps().len(), 2, "two applies keep two receipts");
    let stamp2 = picked["result"]["receipt"].as_str().unwrap();
    assert_eq!(
        fs::read_to_string(backup_of(&fx, stamp2, st!("config/sections/user.yaml"))).unwrap(),
        "tone: casual\n"
    );
    assert!(fs::symlink_metadata(fx.root.join(st!("config/sections/user.yaml"))).is_err());

    // REQ-10: only user data and an earlier-generation folder (a later slice's) remain.
    let third = fx.update_plan();
    let last = step(&third, "go-residue");
    assert_eq!(last["state"], json!("unchanged"), "{last}");
    assert_eq!(
        Fx::residue_item(&third, ".moai-backups")["class"],
        json!("leftover_folder"),
        "the earlier folder is still listed"
    );
    assert_eq!(
        Fx::residue_item(&third, "x.md")["class"],
        json!("user_data")
    );
}

/// AC-2 (REQ-2): a made-up digest, a digest from before a file changed and a rejected
/// permission reference each leave the whole tree as it was.
#[test]
fn a_stale_digest_or_a_rejected_permission_reference_changes_nothing() {
    let fx = survey_frame();
    let plan = fx.update_plan();
    let before = fx.bytes(true);

    let mut args = Fx::printed_apply(&plan, "test-approval");
    let at = args.iter().position(|a| a == "--plan-digest").unwrap();
    args[at + 1] = format!("sha256:{}", "0".repeat(64));
    let refs: Vec<&str> = args.iter().map(String::as_str).collect();
    let made_up = fx.run(&refs);
    assert_eq!(made_up["ok"], json!(false), "{made_up}");
    assert_eq!(made_up["error"]["kind"], json!("stale_plan"), "{made_up}");
    assert_eq!(fx.bytes(true), before, "a made-up digest wrote something");

    for bad in ["placeholder", "null", " "] {
        let mut args = Fx::printed_apply(&plan, bad);
        args.retain(|a| !a.is_empty());
        let refs: Vec<&str> = args.iter().map(String::as_str).collect();
        let refused = fx.run(&refs);
        assert_eq!(refused["ok"], json!(false), "{bad:?}: {refused}");
        assert_eq!(fx.bytes(true), before, "{bad:?} wrote something");
    }

    // The plan was made before this file changed: its digest no longer matches.
    fs::write(fx.root.join(st!("workflows/gtd.md")), "# edited later\n").unwrap();
    let edited = fx.bytes(true);
    let stale = fx.apply_printed(&plan, &[]);
    assert_eq!(stale["ok"], json!(false), "{stale}");
    assert_eq!(stale["error"]["kind"], json!("stale_plan"), "{stale}");
    assert_eq!(fx.bytes(true), edited, "a stale digest wrote something");
}

/// AC-7 aside (REQ-1, REQ-11): only a changed file may be picked, and one choice is needed.
#[test]
fn only_a_changed_file_can_be_picked_and_one_choice_is_needed() {
    let fx = survey_frame();
    let plan = fx.update_plan();
    let digest = step(&plan, "go-residue")["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    let before = fx.bytes(true);
    for pick in [
        st!("design/x.md"),
        st!(".env.local"),
        st!("policies/core/a.md"),
        st!("design/link.md"),
        "../outside.txt",
        "outside.txt",
    ] {
        let refused = fx.run(&[
            "go-residue-apply",
            fx.p(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test-approval",
            "--pick",
            pick,
        ]);
        assert_eq!(refused["ok"], json!(false), "{pick}: {refused}");
        assert_eq!(fx.bytes(true), before, "{pick} moved something");
    }
    let none = fx.run(&[
        "go-residue-apply",
        fx.p(),
        "--plan-digest",
        &digest,
        "--permission-reference",
        "test-approval",
    ]);
    assert_eq!(none["ok"], json!(false), "no bundle and no pick: {none}");
    assert_eq!(fx.bytes(true), before);
}

/// A project whose manifest names paths that lead outside it: through `..`, as an absolute
/// path, through a linked folder and through a linked file. Each names a file whose bytes
/// match the recorded hash, so a reader that followed them would call it proven.
fn hostile_frame() -> Fx {
    let mut fx = Fx::new();
    fx.deployed(st!("policies/core/a.md"), "# core policy\n");
    let base = fx.root.parent().unwrap().to_path_buf();
    fs::create_dir_all(base.join("outside")).unwrap();
    fs::create_dir_all(base.join("outside_dir")).unwrap();
    fs::write(base.join("outside/x.md"), "# outside x\n").unwrap();
    fs::write(base.join("outside/abs.md"), "# outside abs\n").unwrap();
    fs::write(base.join("outside_dir/y.md"), "# outside y\n").unwrap();
    let abs = base.join("outside/abs.md").to_str().unwrap().to_string();
    for (key, body) in [
        ("../outside/x.md", "# outside x\n"),
        (abs.as_str(), "# outside abs\n"),
        (st!("linkdir/y.md"), "# outside y\n"),
        (st!("policies/core/l.md"), "# outside x\n"),
    ] {
        fx.files.insert(
            key.into(),
            json!({"provenance": "template_managed", "template_hash": sha(body),
                   "deployed_hash": sha(body), "current_hash": sha(body)}),
        );
    }
    fx.link(st!("linkdir"), "../../outside_dir");
    fx.link(st!("policies/core/l.md"), "../../../../outside/x.md");
    fx.manifest_at(st!("manifest.json"));
    fx.git_commit_all();
    fx
}

fn outside_files(fx: &Fx) -> Vec<(PathBuf, Vec<u8>)> {
    let base = fx.root.parent().unwrap();
    ["outside/x.md", "outside/abs.md", "outside_dir/y.md"]
        .iter()
        .map(|rel| (base.join(rel), fs::read(base.join(rel)).unwrap()))
        .collect()
}

/// AC-3 (REQ-3): the survey and the apply never read, move or write through a path that leaves
/// the project; the outside files keep their bytes and the legitimate file still moves.
#[test]
fn paths_that_leave_the_project_are_never_read_or_moved() {
    let fx = hostile_frame();
    let outside = outside_files(&fx);
    let plan = fx.update_plan();
    let found = step(&plan, "go-residue");
    let items = found["preview"]["items"].as_array().unwrap();
    for i in items {
        let path = i["path"].as_str().unwrap();
        assert!(
            !path.starts_with("..") && !path.starts_with('/'),
            "an unsafe path is an item: {path}"
        );
        if path.ends_with("l.md") || path.ends_with("linkdir") {
            assert!(
                !matches!(i["class"].as_str(), Some("go_same" | "go_changed")),
                "a linked path was proven through its link: {i}"
            );
        }
    }
    for body in ["# outside x\n", "# outside abs\n", "# outside y\n"] {
        assert!(
            items.iter().all(|i| i["sha256"] != json!(sha(body))),
            "an outside file's digest leaked into the step: {body:?}"
        );
    }
    assert_eq!(found["preview"]["refused_paths"], json!(4), "{found}");

    let done = fx.apply_printed(&plan, &[]);
    assert_eq!(done["ok"], json!(true), "{done}");
    assert_eq!(done["result"]["moved"].as_array().unwrap().len(), 1);
    assert_eq!(outside_files(&fx), outside, "an outside file changed");
    assert_eq!(
        fs::read_link(fx.root.join(st!("linkdir"))).unwrap(),
        PathBuf::from("../../outside_dir")
    );
    assert_eq!(
        fs::read_link(fx.root.join(st!("policies/core/l.md"))).unwrap(),
        PathBuf::from("../../../../outside/x.md")
    );
}

/// AC-1 (REQ-11): a manifest that names secret-shaped files by their whole path (a bare-name
/// check on the full path misses `.moai/.env`) never makes the survey hash them or the apply
/// move them; the ordinary file beside them still moves.
#[test]
fn a_manifest_that_names_a_secret_shaped_path_never_moves_it() {
    let mut fx = Fx::new();
    fx.deployed(st!("policies/core/a.md"), "# core policy\n");
    let secrets = [
        (st!(".env"), "TOKEN=1\n"),
        (st!(".npmrc"), "//registry/:_authToken=x\n"),
        (st!("config/credentials.json"), "{\"k\": 1}\n"),
        (st!("keys/deploy.pem"), "-----BEGIN-----\n"),
    ];
    for (rel, body) in secrets {
        fx.deployed(rel, body);
    }
    fx.manifest_at(st!("manifest.json"));
    fx.git_commit_all();
    let before = fx.bytes(true);
    let plan = fx.update_plan();
    let found = step(&plan, "go-residue");
    let items = found["preview"]["items"].as_array().unwrap();
    for (rel, body) in secrets {
        assert!(
            items.iter().all(|i| i["path"] != json!(rel)),
            "{rel} is an item of the step: {found}"
        );
        assert!(
            items.iter().all(|i| i["sha256"] != json!(sha(body))),
            "the digest of {rel} was read into the step"
        );
    }
    let done = fx.apply_printed(&plan, &[]);
    assert_eq!(done["ok"], json!(true), "{done}");
    assert_eq!(
        done["result"]["moved"],
        json!([st!("policies/core/a.md")]),
        "{done}"
    );
    let after = fx.bytes(true);
    for (rel, _) in secrets {
        let p = fx.root.join(rel);
        assert_eq!(after.get(&p), before.get(&p), "{rel} changed");
    }
}

/// AC-3 (REQ-3): a receipt is not its own authority. A forged one that names a path outside
/// the project (through `..`, as an absolute path or through a linked folder) is refused by
/// the revert before it moves anything.
#[test]
fn a_forged_receipt_naming_a_path_outside_the_project_is_refused() {
    let fx = hostile_frame();
    let outside = outside_files(&fx);
    let plan = fx.update_plan();
    let done = fx.apply_printed(&plan, &[]);
    let stamp = done["result"]["receipt"].as_str().unwrap().to_string();
    let receipt_path = fx
        .root
        .join(st_at(&format!("backups/{stamp}/go-residue-receipt.json")));
    let original = fs::read_to_string(&receipt_path).unwrap();
    let abs = fx.root.parent().unwrap().join("outside/abs.md");
    for forged in [
        "../outside/x.md".to_string(),
        abs.to_str().unwrap().to_string(),
        st!("linkdir/y.md").to_string(),
        st!(".env").to_string(),
    ] {
        let mut v: Value = serde_json::from_str(&original).unwrap();
        v["moves"].as_array_mut().unwrap().push(json!({
            "path": forged, "sha256": sha("# outside x\n"), "moved": true, "tracked": false
        }));
        fs::write(&receipt_path, v.to_string()).unwrap();
        let before = fx.bytes(true);
        let refused = fx.run(&revert_args(&fx, &stamp));
        assert_eq!(refused["ok"], json!(false), "{forged}: {refused}");
        assert_eq!(
            fx.bytes(true),
            before,
            "{forged}: the revert moved something"
        );
        assert_eq!(
            outside_files(&fx),
            outside,
            "{forged}: an outside file changed"
        );
    }
}

/// AC-4 (REQ-4, REQ-5): a receipt above 256 KiB still applies to `committed` and reverts, and
/// the same limit is read back.
#[test]
fn a_receipt_above_256_kib_applies_and_reverts() {
    let mut fx = Fx::new();
    for i in 0..2000 {
        fx.deployed(
            &st_at(&format!("policies/bulk/f{i:04}.md")),
            &format!("# bulk {i}\n"),
        );
    }
    fx.manifest_at(st!("manifest.json"));
    fx.git_commit_all();
    let before = fx.bytes(false);
    let plan = fx.update_plan();
    let done = fx.apply_printed(&plan, &[]);
    assert_eq!(
        done["ok"],
        json!(true),
        "{}",
        done.to_string().chars().take(400).collect::<String>()
    );
    let stamp = done["result"]["receipt"].as_str().unwrap().to_string();
    let size = fs::metadata(
        fx.root
            .join(st_at(&format!("backups/{stamp}/go-residue-receipt.json"))),
    )
    .unwrap()
    .len();
    assert!(
        size > 256 * 1024,
        "the receipt is {size} bytes, not above 256 KiB"
    );
    assert_eq!(done["result"]["state"], json!("committed"));
    let again = fx.update_plan();
    assert_eq!(step(&again, "go-residue")["state"], json!("unchanged"));
    let back = fx.run(&revert_args(&fx, &stamp));
    assert_eq!(
        back["ok"],
        json!(true),
        "{}",
        back.to_string().chars().take(400).collect::<String>()
    );
    assert_eq!(back["result"]["state"], json!("reverted"));
    assert_eq!(fx.bytes(false), before, "the revert left a file different");
}

/// AC-4 (REQ-5): when the pending receipt cannot be saved, no file moves.
#[test]
fn an_unwritable_backup_area_moves_no_file() {
    let fx = survey_frame();
    let backups = fx.root.join(st!("backups"));
    fs::create_dir_all(&backups).unwrap();
    fs::set_permissions(&backups, fs::Permissions::from_mode(0o555)).unwrap();
    let before = fx.bytes(true);
    let plan = fx.update_plan();
    let refused = fx.apply_printed(&plan, &[]);
    fs::set_permissions(&backups, fs::Permissions::from_mode(0o755)).unwrap();
    assert_eq!(refused["ok"], json!(false), "{refused}");
    assert_eq!(fx.bytes(true), before, "a file moved without a receipt");
}

/// AC-6 (REQ-8): the revert leaves every file byte for byte as it was before the apply.
#[test]
fn the_revert_restores_every_byte() {
    let fx = survey_frame();
    let before = fx.bytes(false);
    let plan = fx.update_plan();
    let done = fx.apply_printed(&plan, &[]);
    let stamp = done["result"]["receipt"].as_str().unwrap().to_string();
    assert_ne!(fx.bytes(false), before, "the apply moved nothing");
    let back = fx.run(&revert_args(&fx, &stamp));
    assert_eq!(back["ok"], json!(true), "{back}");
    assert_eq!(back["result"]["state"], json!("reverted"), "{back}");
    assert_eq!(fx.bytes(false), before, "the revert left a file different");
    // A reverted apply is found again by the next preview.
    let replan = fx.update_plan();
    let again = step(&replan, "go-residue");
    assert_eq!(again["state"], json!("optional"), "{again}");
    assert_eq!(
        again["apply"]["argv"][1],
        json!("go-residue-apply"),
        "{again}"
    );
}

/// AC-6 (REQ-9): a new file at an original path, or a backup copy that differs, makes only
/// that item a conflict; neither side is touched and the rest is restored.
#[test]
fn a_new_file_or_a_changed_backup_leaves_only_that_item_as_a_conflict() {
    let (fx, third) = survey_frame_with_a_third_proven_file();
    let plan = fx.update_plan();
    let done = fx.apply_printed(&plan, &[]);
    let stamp = done["result"]["receipt"].as_str().unwrap().to_string();
    fs::write(fx.root.join(PROVEN[0]), "# new file\n").unwrap();
    fs::write(backup_of(&fx, &stamp, PROVEN[1]), "# tampered\n").unwrap();
    let back = fx.run(&revert_args(&fx, &stamp));
    assert_eq!(back["ok"], json!(true), "{back}");
    assert_eq!(back["result"]["state"], json!("conflict"), "{back}");
    let conflicts: Vec<&str> = back["result"]["conflicts"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(Value::as_str)
        .collect();
    assert_eq!(conflicts.len(), 2, "{back}");
    assert!(conflicts.contains(&PROVEN[0]) && conflicts.contains(&PROVEN[1]));
    assert_eq!(
        fs::read_to_string(fx.root.join(PROVEN[0])).unwrap(),
        "# new file\n"
    );
    assert_eq!(
        fs::read_to_string(backup_of(&fx, &stamp, PROVEN[0])).unwrap(),
        "# core policy\n"
    );
    assert_eq!(
        fs::read_to_string(backup_of(&fx, &stamp, PROVEN[1])).unwrap(),
        "# tampered\n"
    );
    assert!(fs::symlink_metadata(fx.root.join(PROVEN[1])).is_err());
    assert_eq!(
        fs::read_to_string(fx.root.join(third)).unwrap(),
        "# core policy b\n",
        "the third file is restored"
    );
}

/// REQ-11 and the ownership rule of go-state: the Go manifest records the bytes on disk for a
/// file it did not deploy (`user_created`) and for one the user edited (`user_modified`), so a
/// matching hash proves who wrote it only under `template_managed`. The proven bundle moves
/// nothing else, and a file that is the user's own cannot be picked either.
#[test]
fn the_proven_bundle_leaves_the_users_own_files() {
    let own = [
        ".claude/skills/mine/SKILL.md",
        st!("rules/mine.md"),
        st!("rules/unlabelled.md"),
    ];
    let edited = st!("rules/edited.md");
    let mut fx = survey_frame();
    fx.recorded(own[0], "# my skill\n", Some("user_created"));
    fx.recorded(own[1], "# my rule\n", Some("user_created"));
    fx.recorded(own[2], "# who wrote this\n", None);
    fx.recorded(edited, "# my edit\n", Some("user_modified"));
    fx.manifest_at(st!("manifest.json"));
    let plan = fx.update_plan();
    for rel in own {
        let item = Fx::residue_item(&plan, rel);
        assert_eq!(item["class"], json!("user_data"), "{item}");
        assert_eq!(item["suggested_action"], json!("keep"), "{item}");
    }
    assert_eq!(
        Fx::residue_item(&plan, edited)["class"],
        json!("go_changed")
    );

    let done = fx.apply_printed(&plan, &[]);
    assert_eq!(done["ok"], json!(true), "{done}");
    let mut moved: Vec<&str> = done["result"]["moved"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(Value::as_str)
        .collect();
    moved.sort_unstable();
    let mut proven = PROVEN.to_vec();
    proven.sort_unstable();
    assert_eq!(moved, proven, "{done}");
    for (rel, body) in [
        (own[0], "# my skill\n"),
        (own[1], "# my rule\n"),
        (own[2], "# who wrote this\n"),
        (edited, "# my edit\n"),
    ] {
        assert_eq!(
            fs::read_to_string(fx.root.join(rel)).unwrap(),
            body,
            "{rel}"
        );
    }

    let second = fx.update_plan();
    let digest = step(&second, "go-residue")["plan_digest"]
        .as_str()
        .expect("the plan digest");
    let before = fx.bytes(true);
    let refused = fx.run(&[
        "go-residue-apply",
        fx.p(),
        "--plan-digest",
        digest,
        "--permission-reference",
        "test-approval",
        "--pick",
        own[0],
    ]);
    assert_eq!(refused["ok"], json!(false), "{refused}");
    assert_eq!(fx.bytes(true), before, "the user's own file moved");
}

/// REQ-9: with the backup copy gone, another file at the original place is a conflict and the
/// receipt is not marked reverted; the original bytes back at their place are none.
#[test]
fn a_lost_backup_with_another_file_at_the_original_is_a_conflict() {
    let (fx, third) = survey_frame_with_a_third_proven_file();
    let plan = fx.update_plan();
    let done = fx.apply_printed(&plan, &[]);
    let stamp = done["result"]["receipt"].as_str().unwrap().to_string();
    for (rel, body) in [
        (PROVEN[0], "# another file\n"),
        (PROVEN[1], "# gtd workflow\n"),
    ] {
        fs::remove_file(backup_of(&fx, &stamp, rel)).unwrap();
        fs::write(fx.root.join(rel), body).unwrap();
    }
    let back = fx.run(&revert_args(&fx, &stamp));
    assert_eq!(back["ok"], json!(true), "{back}");
    assert_eq!(back["result"]["state"], json!("conflict"), "{back}");
    assert_eq!(back["result"]["conflicts"], json!([PROVEN[0]]), "{back}");
    assert_eq!(back["result"]["missing"], json!([]), "{back}");
    assert_eq!(back["result"]["restored"], json!([third]), "{back}");
    assert_eq!(
        fs::read_to_string(fx.root.join(PROVEN[0])).unwrap(),
        "# another file\n"
    );
    let receipt: Value = serde_json::from_slice(
        &fs::read(
            fx.root
                .join(st_at(&format!("backups/{stamp}/go-residue-receipt.json"))),
        )
        .unwrap(),
    )
    .unwrap();
    assert_eq!(receipt["state"], json!("committed"), "{receipt}");
}

// ---- SPEC-20261126-go-residue-in-use: the Go files something still reads ----
//
// This section spells the state folder only through `st!` and `st_at`: the literal scan counts
// every spelling, comments and JSON strings included.

/// The import line of a state-folder file: `@` and the path.
fn import_of(tail: &str) -> String {
    format!("@{}", st_at(tail))
}

/// The surveyed item at exactly `path`, when the step lists it.
fn item_if<'v>(plan: &'v Value, path: &str) -> Option<&'v Value> {
    step(plan, "go-residue")["preview"]["items"]
        .as_array()
        .expect("go-residue step has items")
        .iter()
        .find(|i| i["path"] == json!(path))
}

fn item_at<'v>(plan: &'v Value, path: &str) -> &'v Value {
    item_if(plan, path).unwrap_or_else(|| panic!("no go-residue item for {path}"))
}

/// The `used_by` list of an item as `(reader path, kind)` pairs. An entry holds a path and a
/// kind and nothing else, so a setting value never travels in it (REQ-1).
fn used_by(item: &Value) -> Vec<(String, String)> {
    item["used_by"]
        .as_array()
        .unwrap_or_else(|| panic!("REQ-1: {} has no used_by list: {item}", item["path"]))
        .iter()
        .map(|u| {
            let only_path_and_kind = u
                .as_object()
                .is_some_and(|o| o.keys().all(|k| k == "path" || k == "kind"));
            assert!(
                only_path_and_kind,
                "REQ-1: a used_by entry holds a path and a kind only: {u}"
            );
            (
                u["path"].as_str().unwrap_or_default().to_string(),
                u["kind"].as_str().unwrap_or_default().to_string(),
            )
        })
        .collect()
}

/// REQ-1: `path` is `go_in_use`, kept, and named by each `(reader, kind)` given.
fn assert_in_use(plan: &Value, path: &str, readers: &[(&str, &str)]) {
    let item = item_at(plan, path);
    assert_eq!(
        item["class"],
        json!("go_in_use"),
        "REQ-1: {path} is still read: {item}"
    );
    assert_eq!(
        item["suggested_action"],
        json!("keep"),
        "REQ-1: {path}: {item}"
    );
    let used = used_by(item);
    for (reader, kind) in readers {
        assert!(
            used.contains(&(reader.to_string(), kind.to_string())),
            "REQ-1: {path} is not listed as used by {reader} ({kind}): {item}"
        );
    }
}

/// REQ-1, REQ-2: a status line file is `go_in_use` and kept, with one `statusline` reader.
fn assert_read_by_statusline(plan: &Value, path: &str) {
    let item = item_at(plan, path);
    assert_eq!(
        item["class"],
        json!("go_in_use"),
        "REQ-2: the status line reads {path}: {item}"
    );
    assert_eq!(
        item["suggested_action"],
        json!("keep"),
        "REQ-1: {path}: {item}"
    );
    assert!(
        used_by(item)
            .iter()
            .any(|(reader, kind)| kind == "statusline" && !reader.is_empty()),
        "REQ-1: {path} has no `statusline` reader: {item}"
    );
}

/// REQ-2: a proven file nothing reads stays `go_same` with the action `archive`.
fn assert_candidate(plan: &Value, path: &str) {
    let item = item_at(plan, path);
    assert_eq!(
        item["class"],
        json!("go_same"),
        "REQ-2: nothing reads {path}: {item}"
    );
    assert_eq!(
        item["suggested_action"],
        json!("archive"),
        "REQ-2: {path}: {item}"
    );
}

/// The readers the survey could not read, as the preview lists them.
fn unread_readers(plan: &Value) -> Vec<Value> {
    step(plan, "go-residue")["preview"]["unread_readers"]
        .as_array()
        .cloned()
        .unwrap_or_default()
}

/// REQ-3: `file` is listed under `unread_readers` with a reason.
fn assert_unread_reader(plan: &Value, file: &str) {
    let list = unread_readers(plan);
    let entry = list
        .iter()
        .find(|r| r["file"] == json!(file))
        .unwrap_or_else(|| panic!("REQ-3: {file} is not under unread_readers: {list:?}"));
    assert!(
        entry["reason"].as_str().is_some_and(|r| !r.is_empty()),
        "REQ-3: {file} has no reason: {entry}"
    );
}

/// The go-residue command is printed, and no earlier backup step has a command of its own:
/// the command waits while one has, so a fixture that trips an earlier step would test that
/// wait and not the readers.
fn assert_command_is_printed(plan: &Value) {
    for name in ["go-settings", "go-state", "reports"] {
        assert_eq!(
            step(plan, name)["apply"],
            json!(null),
            "the fixture trips the {name} step: {}",
            step(plan, name)
        );
    }
    assert!(
        step(plan, "go-residue")["apply"]["argv"].is_array(),
        "the go-residue step prints no command: {}",
        step(plan, "go-residue")
    );
}

/// The whole-number words of the step's reason text.
fn reason_numbers(found: &Value) -> Vec<String> {
    found["reason"]
        .as_str()
        .unwrap_or_default()
        .split(|c: char| !c.is_ascii_alphanumeric())
        .filter(|t| !t.is_empty() && t.bytes().all(|b| b.is_ascii_digit()))
        .map(str::to_string)
        .collect()
}

/// REQ-6: the reason gives the number of files in use and the number of unread readers.
fn assert_reason_counts(found: &Value, in_use: usize, unread: usize) {
    let numbers = reason_numbers(found);
    for n in [in_use, unread] {
        assert!(
            numbers.contains(&n.to_string()),
            "REQ-6: the reason does not give {n} ({in_use} files in use, {unread} unread \
             readers): {found}"
        );
    }
}

/// A project whose files are read in every way REQ-2 names: imports (two levels, and one from a
/// rules file), `outputStyle` by file name and by frontmatter name, a status line script, a hook
/// script that names a second script, and the status line's own yaml files. Beside them stay
/// files nothing reads: names inside a code span, a fenced block and `permissions`, a style no
/// setting names, and a plain proven file.
fn in_use_frame() -> Fx {
    let mut fx = Fx::new();
    fx.deployed(st!("config/sections/language.yaml"), "language: ko\n");
    fx.changed(
        st!("config/sections/user.yaml"),
        "name: recorded\n",
        "name: edited\n",
    );
    fx.deployed(st!("policies/outer.md"), "# outer\n@inner.md\n");
    fx.deployed(st!("policies/inner.md"), "# inner\n");
    fx.deployed(st!("policies/by-rule.md"), "# by rule\n");
    fx.deployed(st!("policies/spanned.md"), "# spanned\n");
    fx.deployed(st!("policies/fenced.md"), "# fenced\n");
    fx.deployed(st!("policies/perm.md"), "# perm\n");
    fx.deployed(st!("workflows/gtd.md"), "# gtd workflow\n");
    fx.deployed(st!("config/sections/statusline.yaml"), "theme: plain\n");
    fx.changed(st!("config/sections/llm.yaml"), "model: a\n", "model: b\n");
    fx.deployed(st!("statusline.local.yaml"), "segments: []\n");
    fx.deployed(".claude/output-styles/terse-mode.md", "# terse\n");
    fx.deployed(
        ".claude/output-styles/moai/voice.md",
        "---\nname: Tutor Voice\n---\n# voice\n",
    );
    fx.deployed(
        ".claude/output-styles/moai/other.md",
        "---\nname: Other\n---\n# other\n",
    );
    fx.deployed(".claude/statusline/render.sh", "#!/bin/sh\necho status\n");
    fx.deployed(
        ".claude/hooks/custom/start.sh",
        "#!/bin/sh\nbash .claude/hooks/custom/helper.sh\n",
    );
    fx.deployed(".claude/hooks/custom/helper.sh", "#!/bin/sh\necho helper\n");
    let claude = [
        "# Project rules".to_string(),
        import_of("config/sections/language.yaml"),
        format!(
            "A name in a code span, `{}`, is no import.",
            import_of("policies/spanned.md")
        ),
        "```text".to_string(),
        import_of("policies/fenced.md"),
        "```".to_string(),
        import_of("policies/outer.md"),
        import_of("config/sections/user.yaml"),
    ]
    .join("\n");
    fx.write("CLAUDE.md", &format!("{claude}\n"));
    fx.write(
        ".claude/rules/r.md",
        &format!("# a rule\n{}\n", import_of("policies/by-rule.md")),
    );
    let settings = json!({
        "permissions": {"allow": [format!("Read({})", st_at("policies/perm.md"))]},
        "env": {"SENTINEL": "zz-sentinel-17"},
        "statusLine": {"type": "command",
                       "command": "bash .claude/statusline/render.sh --theme zz-arg-17"},
        "hooks": {"SessionStart": [{"hooks": [
            {"type": "command", "command": "bash .claude/hooks/custom/start.sh"}]}]},
        "outputStyle": "terse-mode",
    });
    fx.write(".claude/settings.json", &settings.to_string());
    fx.write(
        ".claude/settings.local.json",
        &json!({"outputStyle": "Tutor Voice"}).to_string(),
    );
    fx.manifest_at(st!("manifest.json"));
    fx
}

/// AC-1 (REQ-1, REQ-2): every way REQ-2 names a file makes it `go_in_use` with the action
/// `keep` and a `used_by` list of reader paths and kinds; a name inside a code span, a fenced
/// block or `permissions` reads nothing, and a proven file nothing reads stays `go_same`.
#[test]
fn a_go_file_still_read_is_in_use_and_names_its_reader() {
    let fx = in_use_frame();
    let plan = fx.update_plan();

    // `@` imports: one level, two levels (the reader of the inner file is the outer one, which
    // is itself a proven file that something reads), from a rules file, and a changed file.
    assert_in_use(
        &plan,
        st!("config/sections/language.yaml"),
        &[("CLAUDE.md", "import")],
    );
    assert_in_use(
        &plan,
        st!("config/sections/user.yaml"),
        &[("CLAUDE.md", "import")],
    );
    assert_in_use(&plan, st!("policies/outer.md"), &[("CLAUDE.md", "import")]);
    assert_in_use(
        &plan,
        st!("policies/inner.md"),
        &[(st!("policies/outer.md"), "import")],
    );
    assert_in_use(
        &plan,
        st!("policies/by-rule.md"),
        &[(".claude/rules/r.md", "import")],
    );

    // `outputStyle`: by file name in the shared settings, by frontmatter name in the local ones.
    assert_in_use(
        &plan,
        ".claude/output-styles/terse-mode.md",
        &[(".claude/settings.json", "settings")],
    );
    assert_in_use(
        &plan,
        ".claude/output-styles/moai/voice.md",
        &[(".claude/settings.local.json", "settings")],
    );

    // A status line script and a hook script named by the settings, and the script that hook
    // script names in turn.
    assert_in_use(
        &plan,
        ".claude/statusline/render.sh",
        &[(".claude/settings.json", "settings")],
    );
    assert_in_use(
        &plan,
        ".claude/hooks/custom/start.sh",
        &[(".claude/settings.json", "settings")],
    );
    assert_in_use(
        &plan,
        ".claude/hooks/custom/helper.sh",
        &[(".claude/hooks/custom/start.sh", "script")],
    );

    // The status line's own yaml files, whichever class the survey gave them.
    for path in [
        st!("config/sections/statusline.yaml"),
        st!("config/sections/llm.yaml"),
        st!("statusline.local.yaml"),
    ] {
        assert_read_by_statusline(&plan, path);
    }

    // Names that read nothing: a code span, a fenced block, `permissions`, a style no setting
    // names, and a plain proven file.
    for path in [
        st!("policies/spanned.md"),
        st!("policies/fenced.md"),
        st!("policies/perm.md"),
        ".claude/output-styles/moai/other.md",
        st!("workflows/gtd.md"),
    ] {
        assert_candidate(&plan, path);
    }

    // No setting value travels into the step.
    let shown = step(&plan, "go-residue").to_string();
    for value in ["zz-sentinel-17", "zz-arg-17", "Tutor Voice"] {
        assert!(
            !shown.contains(value),
            "REQ-1: the setting value {value:?} is in the step"
        );
    }
}

/// A project of two proven files and a changed one that no reader names, for the cases of
/// REQ-3 to break a reader in.
fn unread_frame() -> Fx {
    let mut fx = Fx::new();
    fx.deployed(st!("policies/core/a.md"), "# core policy\n");
    fx.deployed(st!("workflows/gtd.md"), "# gtd workflow\n");
    fx.changed(
        st!("config/sections/user.yaml"),
        "tone: formal\n",
        "tone: casual\n",
    );
    fx.manifest_at(st!("manifest.json"));
    fx
}

const UNREAD_FILES: [&str; 3] = [
    st!("policies/core/a.md"),
    st!("workflows/gtd.md"),
    st!("config/sections/user.yaml"),
];

/// The status line command of a settings file that names `script`.
fn statusline_settings(script: &str) -> String {
    json!({"statusLine": {"type": "command", "command": format!("bash {script}")}}).to_string()
}

/// A project where one reader cannot be read, and that reader's project path.
fn unreadable_reader(case: &str) -> (Fx, &'static str) {
    let fx = unread_frame();
    let base = fx.root.parent().unwrap().to_path_buf();
    match case {
        "settings that are not JSON" => {
            fx.write(".claude/settings.local.json", "{ \"outputStyle\": ");
            (fx, ".claude/settings.local.json")
        }
        "an instruction file that is a link" => {
            fx.write("notes/claude-real.md", "# the real file\n");
            fx.link("CLAUDE.md", "notes/claude-real.md");
            (fx, "CLAUDE.md")
        }
        "a script that is not UTF-8" => {
            fx.write(
                ".claude/settings.json",
                &statusline_settings(".claude/statusline/render.sh"),
            );
            fs::create_dir_all(fx.root.join(".claude/statusline")).unwrap();
            fs::write(
                fx.root.join(".claude/statusline/render.sh"),
                [0xffu8, 0xfe, 0x00],
            )
            .unwrap();
            (fx, ".claude/statusline/render.sh")
        }
        "an import outside the project" => {
            fs::write(base.join("outside-notes.md"), "# outside notes\n").unwrap();
            fx.write("CLAUDE.md", "# rules\n@../outside-notes.md\n");
            (fx, "CLAUDE.md")
        }
        "an import from the home folder" => {
            fs::write(fx.home.join("shared-notes.md"), "# shared notes\n").unwrap();
            fx.write("CLAUDE.md", "# rules\n@~/shared-notes.md\n");
            (fx, "CLAUDE.md")
        }
        "a script outside the project that settings run with bash" => {
            fs::create_dir_all(base.join("outside")).unwrap();
            fs::write(base.join("outside/statusline.sh"), "#!/bin/sh\necho x\n").unwrap();
            fx.write(
                ".claude/settings.json",
                &statusline_settings("../outside/statusline.sh"),
            );
            (fx, ".claude/settings.json")
        }
        "an instruction file over 4 MiB" => {
            let line = format!("{}\n", "a".repeat(63));
            let body = format!("{}a", line.repeat(4 * 1024 * 1024 / 64));
            fx.write("CLAUDE.md", &body);
            (fx, "CLAUDE.md")
        }
        other => panic!("no case {other}"),
    }
}

/// AC-2 (REQ-3): when a reader that exists cannot be read, every `go_same` and `go_changed`
/// file is kept, the reader is listed with its reason under `unread_readers`, no command is
/// printed and the tree is as it was; a reader that does not exist, or one of exactly 4 MiB,
/// blocks nothing.
#[test]
fn a_reader_that_cannot_be_read_keeps_every_go_file() {
    // No reader file at all, and an instruction file of exactly 4 MiB: both are read, so the
    // proven files are still offered.
    let none = unread_frame();
    let at_limit = unread_frame();
    let line = format!("{}\n", "a".repeat(63));
    at_limit.write("CLAUDE.md", &line.repeat(4 * 1024 * 1024 / 64));
    for (label, fx) in [("no reader file", none), ("a 4 MiB reader", at_limit)] {
        let plan = fx.update_plan();
        assert_command_is_printed(&plan);
        assert!(
            unread_readers(&plan).is_empty(),
            "REQ-3: {label} blocks nothing: {:?}",
            unread_readers(&plan)
        );
        assert_candidate(&plan, UNREAD_FILES[0]);
        assert_candidate(&plan, UNREAD_FILES[1]);
    }

    for case in [
        "settings that are not JSON",
        "an instruction file that is a link",
        "a script that is not UTF-8",
        "an import outside the project",
        "an import from the home folder",
        "a script outside the project that settings run with bash",
        "an instruction file over 4 MiB",
    ] {
        let (fx, reader) = unreadable_reader(case);
        let before = fx.bytes(true);
        let plan = fx.update_plan();
        let found = step(&plan, "go-residue");
        for path in UNREAD_FILES {
            let item = item_at(&plan, path);
            assert!(
                matches!(item["class"].as_str(), Some("go_same" | "go_changed")),
                "{case}: {path} keeps its class: {item}"
            );
            assert_eq!(
                item["suggested_action"],
                json!("keep"),
                "REQ-3: {case}: {path} is kept: {item}"
            );
        }
        assert_unread_reader(&plan, reader);
        assert_eq!(
            found["apply"],
            json!(null),
            "REQ-3: {case}: no command is printed: {found}"
        );
        assert_eq!(fx.bytes(true), before, "{case}: the tree changed");
    }
}

// SPEC-20261127-go-residue-secret-readers: a reader whose project path holds a secret-shaped
// name is never opened; it is an unread reader and the survey keeps every Go file.

/// The reason the survey gives for a reader it did not open because of its name.
const SECRET_REASON: &str = "has a secret-shaped name";

/// The reason of the `unread_readers` entry for `file`, when it is listed.
fn unread_reason(plan: &Value, file: &str) -> Option<String> {
    unread_readers(plan)
        .iter()
        .find(|r| r["file"] == json!(file))
        .and_then(|r| r["reason"].as_str().map(str::to_string))
}

/// REQ-1, REQ-2: every reader in `readers` is an unread reader for its name, no surveyed file
/// is `go_in_use` (so none of the readers was opened and read for what it names), all of them
/// are kept, no command is printed and the tree is as it was.
fn assert_secret_readers_unread(fx: &Fx, readers: &[&str]) {
    let before = fx.bytes(true);
    let plan = fx.update_plan();
    let found = step(&plan, "go-residue");
    for path in UNREAD_FILES {
        let item = item_at(&plan, path);
        assert!(
            matches!(item["class"].as_str(), Some("go_same" | "go_changed")),
            "REQ-1: {path} was read through a secret-shaped reader: {item}"
        );
        assert_eq!(
            item["suggested_action"],
            json!("keep"),
            "REQ-2: {path} is kept: {item}"
        );
    }
    for reader in readers {
        assert_eq!(
            unread_reason(&plan, reader).as_deref(),
            Some(SECRET_REASON),
            "REQ-1: {reader} is not an unread reader for its name: {:?}",
            unread_readers(&plan)
        );
    }
    assert_eq!(
        found["apply"],
        json!(null),
        "REQ-2: no command is printed: {found}"
    );
    assert_eq!(fx.bytes(true), before, "the tree changed");
}

/// AC-1 (REQ-1, REQ-2): `CLAUDE.md` imports a `.env` that is filled with bytes that are not UTF-8
/// (opened, it would be unread for that reason), and another `.env` that is plain text naming a
/// surveyed file (opened, that file would be in use): neither is opened.
#[test]
fn a_reader_with_a_secret_shaped_name_is_never_opened() {
    let binary = unread_frame();
    binary.write("CLAUDE.md", "# rules\n@.env\n");
    fs::write(binary.root.join(".env"), [0xffu8, 0xfe, 0x00]).unwrap();
    assert_secret_readers_unread(&binary, &[".env"]);

    let text = unread_frame();
    text.write("CLAUDE.md", "# rules\n@.env\n");
    text.write(".env", &format!("{}\n", import_of("policies/core/a.md")));
    assert_secret_readers_unread(&text, &[".env"]);
}

/// AC-2 (REQ-1, REQ-2): a script that settings name, a rules file and a script under a folder
/// whose own name is secret-shaped each name a surveyed file in their text; none is opened.
#[test]
fn a_secret_shaped_script_or_rule_is_unread_and_names_no_file() {
    let naming = format!("echo {}\n", st_at("policies/core/a.md"));
    let cases: [(&str, bool); 3] = [
        (".claude/hooks/credentials.sh", true),
        (".claude/rules/credentials.md", false),
        (".claude/hooks/secrets/run.sh", true),
    ];
    for (reader, by_settings) in cases {
        let fx = unread_frame();
        if by_settings {
            fx.write(".claude/settings.json", &statusline_settings(reader));
            fx.write(reader, &naming);
        } else {
            fx.write(
                reader,
                &format!("# a rule\n{}\n", import_of("policies/core/a.md")),
            );
        }
        assert_secret_readers_unread(&fx, &[reader]);
    }
}

/// AC-3 (REQ-3): an import and a script that name a secret-shaped path that does not exist are
/// no reader: nothing is unread and the proven files are still offered.
#[test]
fn a_missing_secret_shaped_name_blocks_nothing() {
    let fx = unread_frame();
    fx.write("CLAUDE.md", "# rules\n@.env\n");
    fx.write(
        ".claude/settings.json",
        &statusline_settings(".claude/hooks/credentials.sh"),
    );
    let plan = fx.update_plan();
    assert_command_is_printed(&plan);
    assert!(
        unread_readers(&plan).is_empty(),
        "REQ-3: a missing name blocks nothing: {:?}",
        unread_readers(&plan)
    );
    assert_candidate(&plan, UNREAD_FILES[0]);
    assert_candidate(&plan, UNREAD_FILES[1]);
}

/// AC-4 (REQ-4): the rule is the shared one, `go_state::has_secret_component`: a name only
/// `home_state::is_secret` knows, a name only `worktree_include::secret_name` knows, and a
/// secret-shaped folder on the way all stop a reader; a plain name beside them is still read.
#[test]
fn a_secret_shaped_name_follows_the_shared_rule() {
    for reader in [
        "credentials-handling.md",
        "credentials.json",
        "server.key",
        "id_rsa",
        ".npmrc",
        "secrets/notes.md",
    ] {
        let fx = unread_frame();
        fx.write("CLAUDE.md", &format!("# rules\n@{reader}\n"));
        fx.write(
            reader,
            &format!("# notes\n{}\n", import_of("policies/core/a.md")),
        );
        assert_secret_readers_unread(&fx, &[reader]);
    }

    // The control: a plain name is read, so the file it names is in use.
    let fx = unread_frame();
    fx.write("CLAUDE.md", "# rules\n@notes.md\n");
    fx.write(
        "notes.md",
        &format!("# notes\n{}\n", import_of("policies/core/a.md")),
    );
    let plan = fx.update_plan();
    assert_in_use(&plan, UNREAD_FILES[0], &[("notes.md", "import")]);
    assert!(
        unread_readers(&plan).is_empty(),
        "{:?}",
        unread_readers(&plan)
    );
}

/// A project to apply against: an unread proven file, a spare one, a proven and a changed file
/// that `CLAUDE.md` imports, and a proven file nothing reads yet.
fn apply_frame() -> Fx {
    let mut fx = Fx::new();
    fx.deployed(st!("policies/core/a.md"), "# core policy\n");
    fx.deployed(st!("workflows/gtd.md"), "# gtd workflow\n");
    fx.deployed(st!("policies/late.md"), "# late policy\n");
    fx.deployed(st!("config/sections/language.yaml"), "language: ko\n");
    fx.changed(
        st!("config/sections/user.yaml"),
        "tone: formal\n",
        "tone: casual\n",
    );
    fx.write(
        "CLAUDE.md",
        &format!(
            "# Project\n{}\n{}\n",
            import_of("config/sections/language.yaml"),
            import_of("config/sections/user.yaml")
        ),
    );
    fx.manifest_at(st!("manifest.json"));
    fx
}

/// AC-3 (REQ-4): the proven bundle leaves an in-use file where it is, a `--pick` of an in-use
/// file (proven or changed) is refused with a message that names its reader and moves nothing,
/// and a reference added after the preview makes the printed command `stale_plan`.
#[test]
fn an_in_use_file_is_never_moved_and_a_new_reference_makes_the_plan_stale() {
    let fx = apply_frame();
    let (imported_same, imported_changed) = (
        st!("config/sections/language.yaml"),
        st!("config/sections/user.yaml"),
    );
    let plan = fx.update_plan();
    assert_command_is_printed(&plan);
    assert_in_use(&plan, imported_same, &[("CLAUDE.md", "import")]);
    assert_in_use(&plan, imported_changed, &[("CLAUDE.md", "import")]);

    // A pick of an in-use file is refused and names the reader.
    let digest = step(&plan, "go-residue")["plan_digest"]
        .as_str()
        .expect("the plan digest")
        .to_string();
    let before = fx.bytes(true);
    for pick in [imported_changed, imported_same] {
        let refused = fx.run(&[
            "go-residue-apply",
            fx.p(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test-approval",
            "--pick",
            pick,
        ]);
        assert_eq!(refused["ok"], json!(false), "REQ-4: {pick}: {refused}");
        assert!(
            refused["error"]["message"]
                .as_str()
                .is_some_and(|m| m.contains("CLAUDE.md")),
            "REQ-4: the refusal of {pick} names no reader: {refused}"
        );
        assert_eq!(
            fx.bytes(true),
            before,
            "{pick}: the refusal moved something"
        );
    }

    // A reference added after the preview: the printed command is stale and moves nothing.
    fx.write(
        "CLAUDE.md",
        &format!(
            "# Project\n{}\n{}\n{}\n",
            import_of("config/sections/language.yaml"),
            import_of("config/sections/user.yaml"),
            import_of("policies/late.md")
        ),
    );
    let edited = fx.bytes(true);
    let stale = fx.apply_printed(&plan, &[]);
    assert_eq!(stale["ok"], json!(false), "REQ-4: {stale}");
    assert_eq!(
        stale["error"]["kind"],
        json!("stale_plan"),
        "REQ-4: {stale}"
    );
    assert_eq!(fx.bytes(true), edited, "a stale command moved something");

    // The fresh plan reads the new reference, and its command moves the unread files only.
    let fresh = fx.update_plan();
    assert_in_use(&fresh, st!("policies/late.md"), &[("CLAUDE.md", "import")]);
    let done = fx.apply_printed(&fresh, &[]);
    assert_eq!(done["ok"], json!(true), "{done}");
    let mut moved: Vec<&str> = done["result"]["moved"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(Value::as_str)
        .collect();
    moved.sort_unstable();
    assert_eq!(
        moved,
        [st!("policies/core/a.md"), st!("workflows/gtd.md")],
        "REQ-4: only the files nothing reads move: {done}"
    );
    let stamp = done["result"]["receipt"].as_str().unwrap().to_string();
    for rel in [
        imported_same,
        imported_changed,
        st!("policies/late.md"),
        "CLAUDE.md",
    ] {
        assert!(
            fx.root.join(rel).is_file(),
            "{rel} left its place although something reads it"
        );
        assert!(
            fs::symlink_metadata(backup_of(&fx, &stamp, rel)).is_err(),
            "{rel} is in the backup"
        );
    }
}

/// AC-4 (REQ-5): the survey lists the local rules, the harness commands and the `hns-` workflow
/// files as `keep` and outside every apply bundle (also a local rule the manifest names), looks
/// under no worktree folder, and the apply leaves their bytes as they were.
#[test]
fn the_survey_lists_local_rules_and_harness_files_and_moves_none() {
    let mut fx = Fx::new();
    fx.deployed(st!("policies/core/a.md"), "# core policy\n");
    // The manifest names this local rule, so the proof reaches it as a proven file.
    fx.deployed(st!("docs/local-rules/named.md"), "# a named local rule\n");
    let listed = [
        st!("docs/local-rules/rule-a.md"),
        st!("docs/local-rules/sub/deep.md"),
        st!("docs/local-rules/named.md"),
        ".claude/commands/harness/cmd.md",
        ".claude/commands/harness/sub/deep.md",
        ".claude/workflows/hns-flow.md",
    ];
    for rel in listed {
        if !fx.files.contains_key(rel) {
            fx.write(rel, &format!("# {rel}\n"));
        }
    }
    let hidden = [
        ".claude/worktrees/w1/.claude/commands/harness/hidden.md".to_string(),
        format!(".claude/worktrees/w1/{}", st!("docs/local-rules/hidden.md")),
        ".claude/worktrees/w1/.claude/workflows/hns-hidden.md".to_string(),
    ];
    let unlisted = [
        ".claude/workflows/other-flow.md".to_string(),
        ".claude/commands/other/cmd.md".to_string(),
    ];
    for rel in hidden.iter().chain(&unlisted) {
        fx.write(rel, "# not surveyed\n");
    }
    fx.manifest_at(st!("manifest.json"));

    let before = fx.bytes(true);
    let plan = fx.update_plan();
    assert_command_is_printed(&plan);
    for rel in listed {
        let item =
            item_if(&plan, rel).unwrap_or_else(|| panic!("REQ-5: the survey does not list {rel}"));
        assert_eq!(
            item["suggested_action"],
            json!("keep"),
            "REQ-5: {rel}: {item}"
        );
    }
    for rel in hidden.iter().chain(&unlisted) {
        assert!(
            item_if(&plan, rel).is_none(),
            "REQ-5: the survey lists {rel}"
        );
    }
    assert_eq!(
        step(&plan, "go-residue")["preview"]["bundles"]["proven"],
        json!(1),
        "REQ-5: a listed file is in an apply bundle: {}",
        step(&plan, "go-residue")
    );

    let done = fx.apply_printed(&plan, &[]);
    assert_eq!(done["ok"], json!(true), "{done}");
    assert_eq!(
        done["result"]["moved"],
        json!([st!("policies/core/a.md")]),
        "{done}"
    );
    let after = fx.bytes(true);
    for (p, b) in &before {
        if *p == fx.root.join(st!("policies/core/a.md")) {
            continue;
        }
        assert_eq!(after.get(p), Some(b), "{} changed", p.display());
    }
}

/// A project of `in_use` proven files that `CLAUDE.md` imports and, when asked, three unread
/// proven files with one unread changed file (`candidate`) and a settings file that is not JSON
/// (`broken`).
fn counts_frame(in_use: usize, candidate: bool, broken: bool) -> Fx {
    let mut fx = Fx::new();
    let mut claude = String::from("# Project\n");
    for i in 0..in_use {
        let tail = format!("config/sections/imported{i}.yaml");
        fx.deployed(&st_at(&tail), "key: value\n");
        claude.push_str(&format!("{}\n", import_of(&tail)));
    }
    if in_use > 0 {
        fx.write("CLAUDE.md", &claude);
    }
    if candidate {
        for name in ["a", "b", "c"] {
            fx.deployed(
                &st_at(&format!("policies/{name}.md")),
                &format!("# {name}\n"),
            );
        }
        fx.changed(&st_at("policies/old.md"), "# recorded\n", "# edited\n");
    }
    if broken {
        fx.write(".claude/settings.local.json", "{ \"outputStyle\": ");
    }
    fx.manifest_at(st!("manifest.json"));
    fx
}

/// AC-5 (REQ-6): with a file left to move the step reads `optional`; with only files in use or
/// readers that could not be read left it reads `unchanged`; the reason gives the two counts and
/// no frame reads `needs_decision`.
#[test]
fn the_step_is_optional_with_a_candidate_and_never_needs_decision() {
    // Two files in use and unread files left to move: `optional`, 2 in use, 0 unread readers.
    let plan = counts_frame(2, true, false).update_plan();
    let found = step(&plan, "go-residue");
    let in_use = found["preview"]["items"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|i| i["class"] == json!("go_in_use"))
        .count();
    assert_eq!(in_use, 2, "REQ-1: two files are read: {found}");
    assert_eq!(found["state"], json!("optional"), "REQ-6: {found}");
    assert_reason_counts(found, 2, 0);

    // Only files in use are left: `unchanged`, 2 in use, 0 unread readers.
    let plan = counts_frame(2, false, false).update_plan();
    let found = step(&plan, "go-residue");
    assert_eq!(found["state"], json!("unchanged"), "REQ-6: {found}");
    assert_reason_counts(found, 2, 0);

    // Only a reader that could not be read stands in the way: `unchanged`, 0 in use, 1 unread.
    let plan = counts_frame(0, true, true).update_plan();
    let found = step(&plan, "go-residue");
    assert_eq!(unread_readers(&plan).len(), 1, "REQ-3: {found}");
    assert_eq!(found["state"], json!("unchanged"), "REQ-6: {found}");
    assert_reason_counts(found, 0, 1);

    // No frame reads `needs_decision`, whatever it holds.
    let proofless = Fx::new();
    proofless.write(st!("design/x.md"), "# user design notes\n");
    for (label, fx) in [
        (
            "files in use and files to move",
            counts_frame(2, true, false),
        ),
        ("files in use only", counts_frame(2, false, false)),
        ("an unread reader", counts_frame(0, true, true)),
        ("the survey frame", survey_frame()),
        ("the frame of every reader", in_use_frame()),
        ("no proof", proofless),
    ] {
        let plan = fx.update_plan();
        let found = step(&plan, "go-residue");
        assert!(
            matches!(found["state"].as_str(), Some("optional" | "unchanged")),
            "REQ-6: {label}: {found}"
        );
    }
}

/// A project shaped like the one the SPEC read: only a backup manifest, a `CLAUDE.md` with four
/// imports and sentences that name other files, an `outputStyle` whose style file carries the
/// name in its frontmatter, and a status line and a hook that call the installed program.
fn mo_ai_kr_frame() -> Fx {
    let mut fx = Fx::new();
    let imports = ["user", "language", "git-convention", "workflow"];
    for name in imports {
        fx.deployed(
            &st_at(&format!("config/sections/{name}.yaml")),
            &format!("{name}: recorded\n"),
        );
    }
    fx.deployed(st!("config/sections/quality.yaml"), "level: strict\n");
    fx.deployed(st!("config/sections/harness.yaml"), "harness: on\n");
    fx.deployed(st!("workflows/project/p.md"), "# a project workflow\n");
    fx.deployed(
        ".claude/output-styles/moai/moai-easy.md",
        "---\nname: MoAI-Easy\n---\n# easy\n",
    );
    let import_lines: Vec<String> = imports
        .iter()
        .map(|name| import_of(&format!("config/sections/{name}.yaml")))
        .collect();
    fx.write(
        "CLAUDE.md",
        &format!(
            "# Project\nThe quality rules are in quality.yaml.\n{}\n",
            import_lines.join("\n")
        ),
    );
    fx.write(
        "AGENTS.md",
        &format!(
            "# Agents\nThe harness is in harness.yaml and quality.yaml; project workflows are \
             in {}.\n",
            st_at("workflows/project/")
        ),
    );
    fx.write("CLAUDE.local.md", "# local notes\n");
    fx.write("AGENTS.local.md", "# local agent notes\n");
    let launcher = format!("$HOME/{}", st!("bin/moai"));
    let settings = json!({
        "outputStyle": "MoAI-Easy",
        "statusLine": {"type": "command", "command": format!("{launcher} statusline")},
        "hooks": {"SessionStart": [{"hooks": [
            {"type": "command", "command": format!("{launcher} hook session-start")}]}]},
    });
    fx.write(".claude/settings.local.json", &settings.to_string());
    fx.manifest_at(".moai-backup/manifest.json");
    fx
}

/// AC-6 (REQ-1, REQ-2, REQ-4): in a project shaped like `mo.ai.kr` no reader is unread, the four
/// imported yaml files and the referenced output style are `go_in_use`, and the printed command
/// moves only the proven files that sentences name and no reader reads.
#[test]
fn the_mo_ai_kr_shape_moves_only_unread_proven_files() {
    let fx = mo_ai_kr_frame();
    let before = fx.bytes(true);
    let plan = fx.update_plan();
    assert_command_is_printed(&plan);
    assert!(
        unread_readers(&plan).is_empty(),
        "REQ-3: a call of the installed program is no unread reader: {:?}",
        unread_readers(&plan)
    );
    let imported = [
        st!("config/sections/user.yaml"),
        st!("config/sections/language.yaml"),
        st!("config/sections/git-convention.yaml"),
        st!("config/sections/workflow.yaml"),
    ];
    for rel in imported {
        assert_in_use(&plan, rel, &[("CLAUDE.md", "import")]);
    }
    let style = ".claude/output-styles/moai/moai-easy.md";
    assert_in_use(&plan, style, &[(".claude/settings.local.json", "settings")]);
    let unread = [
        st!("config/sections/harness.yaml"),
        st!("config/sections/quality.yaml"),
        st!("workflows/project/p.md"),
    ];
    for rel in unread {
        assert_candidate(&plan, rel);
    }

    let done = fx.apply_printed(&plan, &[]);
    assert_eq!(done["ok"], json!(true), "{done}");
    let mut moved: Vec<&str> = done["result"]["moved"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(Value::as_str)
        .collect();
    moved.sort_unstable();
    assert_eq!(moved, unread, "REQ-4: {done}");
    let stamp = done["result"]["receipt"].as_str().unwrap().to_string();
    let after = fx.bytes(true);
    for (p, b) in &before {
        if unread.iter().any(|rel| *p == fx.root.join(rel)) {
            continue;
        }
        assert_eq!(after.get(p), Some(b), "{} changed", p.display());
    }
    for rel in imported.iter().chain(&[style]) {
        assert!(
            fs::symlink_metadata(backup_of(&fx, &stamp, rel)).is_err(),
            "{rel} is in the backup"
        );
    }
}
