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
            ".moai/moai.json",
            &json!({"generation": "moai", "created_by": "moai-cli test",
                    "go_backup": ".moai-backup"})
            .to_string(),
        );
        fx.write(".moai/project.yaml", "schema_version: 2\n");
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
                } else if backups || !p.starts_with(self.root.join(".moai/backups")) {
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
        let mut v: Vec<String> = fs::read_dir(self.root.join(".moai/backups"))
            .map(|rd| {
                rd.filter_map(|e| e.ok())
                    .filter_map(|e| e.file_name().into_string().ok())
                    .filter(|n| {
                        self.root
                            .join(".moai/backups")
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
    fx.deployed(".moai/policies/core/a.md", "# core policy\n");
    fx.deployed(".moai/workflows/gtd.md", "# gtd workflow\n");
    fx.deployed(
        ".moai/output-styles/moai/moai-easy.md",
        "# moai easy style\n",
    );
    fx.changed(
        ".moai/config/sections/user.yaml",
        "tone: formal\n",
        "tone: casual\n",
    );
    fx.write(".moai/design/x.md", "# user design notes\n");
    fx.write(".moai/.env.local", "KEY=value\n");
    // A link whose target sits outside the state folder: following it would read content
    // the survey never walks, so its digest must stay empty.
    fx.write("outside.txt", "outside content\n");
    fx.link(".moai/design/link.md", "../outside.txt");
    fx.write(
        ".claude/settings.json",
        "{\"outputStyle\": \"moai/moai-easy\"}\n",
    );
    fx.manifest_at(".moai/manifest.json");
    fx.git_commit_all();
    fx
}

/// AC-1 (REQ-1, REQ-7): the step sorts the frame — `go_same` for the byte-identical files,
/// `go_changed` for the rewritten one, `user_data` for the unmanifested one — counts the one
/// secret-shaped name without opening it, marks tracked files, carries the item fields, and a
/// second plan run reads the same.
#[test]
fn the_go_residue_step_classifies_the_survey_frame() {
    let fx = survey_frame();
    let first = fx.update_plan();
    let found = step(&first, "go-residue");
    for name in [
        "policies/core/a.md",
        "workflows/gtd.md",
        "output-styles/moai/moai-easy.md",
    ] {
        let item = Fx::residue_item(&first, name.split('/').next_back().unwrap());
        assert_eq!(item["class"], json!("go_same"), "{name}");
        assert_eq!(item["tracked"], json!(true), "{name}");
    }
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
    fs::create_dir_all(fx.root.join(".moai/backups/go-markers-20261120T234501Z")).unwrap();
    fs::rename(
        fx.root.join(".moai/manifest.json"),
        fx.root
            .join(".moai/backups/go-markers-20261120T234501Z/manifest.json"),
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
    fx.write(".moai/design/x.md", "# user design notes\n");
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

/// AC-3 (REQ-6): with no manifest anywhere the step reports no proof with zero candidates
/// and the plan still succeeds.
#[test]
fn the_go_residue_step_reports_no_proof_when_no_manifest_exists() {
    let fx = survey_frame();
    fs::remove_file(fx.root.join(".moai/manifest.json")).unwrap();
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

const PROVEN: [&str; 3] = [
    ".moai/policies/core/a.md",
    ".moai/workflows/gtd.md",
    ".moai/output-styles/moai/moai-easy.md",
];

/// Where the apply put `rel`: `.moai/backups/<stamp>/go-residue/<same path>`.
fn backup_of(fx: &Fx, stamp: &str, rel: &str) -> PathBuf {
    fx.root
        .join(format!(".moai/backups/{stamp}/go-residue/{rel}"))
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
/// the three byte-identical files into the backup area and nothing else; the changed file is
/// asked about and moves only with an explicit pick; once nothing proven or changed is left
/// the step reads `unchanged` although an earlier-generation folder is still listed.
#[test]
fn applying_the_proven_bundle_moves_only_the_go_same_files() {
    let fx = survey_frame();
    fs::create_dir_all(fx.root.join(".moai-backups/20261101T000000Z")).unwrap();
    let before = fx.bytes(true);
    let plan = fx.update_plan();
    assert_eq!(step(&plan, "go-residue")["state"], json!("needs_decision"));

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
                .join(format!(".moai/backups/{stamp}/go-residue-receipt.json")),
        )
        .unwrap(),
    )
    .unwrap();
    assert_eq!(receipt["state"], json!("committed"));
    assert_eq!(receipt["moves"].as_array().unwrap().len(), 3);
    // REQ-12: the three moved files are tracked; the result says so and warns.
    assert_eq!(r["tracked_moved"], json!(3), "{r}");
    assert!(
        r["warning"]
            .as_str()
            .is_some_and(|w| w.contains("checkout")),
        "the warning names the checkout: {r}"
    );

    // The changed file is still asked about, and nothing proven is left to move.
    let second = fx.update_plan();
    let found = step(&second, "go-residue");
    assert_eq!(found["state"], json!("needs_decision"), "{found}");
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
        ".moai/config/sections/user.yaml",
    ]);
    assert_eq!(picked["ok"], json!(true), "{picked}");
    assert_eq!(fx.stamps().len(), 2, "two applies keep two receipts");
    let stamp2 = picked["result"]["receipt"].as_str().unwrap();
    assert_eq!(
        fs::read_to_string(backup_of(&fx, stamp2, ".moai/config/sections/user.yaml")).unwrap(),
        "tone: casual\n"
    );
    assert!(fs::symlink_metadata(fx.root.join(".moai/config/sections/user.yaml")).is_err());

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
    fs::write(fx.root.join(".moai/workflows/gtd.md"), "# edited later\n").unwrap();
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
        ".moai/design/x.md",
        ".moai/.env.local",
        ".moai/policies/core/a.md",
        ".moai/design/link.md",
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
    fx.deployed(".moai/policies/core/a.md", "# core policy\n");
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
        (".moai/linkdir/y.md", "# outside y\n"),
        (".moai/policies/core/l.md", "# outside x\n"),
    ] {
        fx.files.insert(
            key.into(),
            json!({"provenance": "template_managed", "template_hash": sha(body),
                   "deployed_hash": sha(body), "current_hash": sha(body)}),
        );
    }
    fx.link(".moai/linkdir", "../../outside_dir");
    fx.link(".moai/policies/core/l.md", "../../../../outside/x.md");
    fx.manifest_at(".moai/manifest.json");
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
        fs::read_link(fx.root.join(".moai/linkdir")).unwrap(),
        PathBuf::from("../../outside_dir")
    );
    assert_eq!(
        fs::read_link(fx.root.join(".moai/policies/core/l.md")).unwrap(),
        PathBuf::from("../../../../outside/x.md")
    );
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
        .join(format!(".moai/backups/{stamp}/go-residue-receipt.json"));
    let original = fs::read_to_string(&receipt_path).unwrap();
    let abs = fx.root.parent().unwrap().join("outside/abs.md");
    for forged in [
        "../outside/x.md".to_string(),
        abs.to_str().unwrap().to_string(),
        ".moai/linkdir/y.md".to_string(),
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
            &format!(".moai/policies/bulk/f{i:04}.md"),
            &format!("# bulk {i}\n"),
        );
    }
    fx.manifest_at(".moai/manifest.json");
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
            .join(format!(".moai/backups/{stamp}/go-residue-receipt.json")),
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
    let backups = fx.root.join(".moai/backups");
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
    assert_eq!(
        step(&replan, "go-residue")["state"],
        json!("needs_decision")
    );
}

/// AC-6 (REQ-9): a new file at an original path, or a backup copy that differs, makes only
/// that item a conflict; neither side is touched and the rest is restored.
#[test]
fn a_new_file_or_a_changed_backup_leaves_only_that_item_as_a_conflict() {
    let fx = survey_frame();
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
        fs::read_to_string(fx.root.join(PROVEN[2])).unwrap(),
        "# moai easy style\n",
        "the third file is restored"
    );
}
