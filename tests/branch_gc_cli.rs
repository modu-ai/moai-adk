//! SPEC-20261123-branch-gc: `branch-gc-plan` lists the local branches that are merged and
//! unused, `branch-gc-apply` deletes the planned ones with `git branch -d` after a digest match.
//! Every repository here is a temporary one under a temporary HOME; the real branches, ledger
//! and remote of this clone are never planned or touched.
#![cfg(unix)]
mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Output;

const SESSION: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const PERMISSION: &str = "fixture-approved";

fn git(dir: &Path, args: &[&str]) -> String {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all)
}

fn commit_file(dir: &Path, name: &str) {
    fs::write(dir.join(name), format!("{name}\n")).unwrap();
    git(dir, &["add", name]);
    git(dir, &["commit", "-q", "-m", name]);
}

struct Fx {
    home: TempHome,
    project: PathBuf,
    wts: PathBuf,
}

/// A registered project that is a git repository. `main` and the branches below sit at the first
/// commit `C0`; `integ` is the integration branch, `C0` plus a no-fast-forward merge of
/// `feat/done`, and is checked out in the primary checkout.
///
/// - merged and unused: `feat/at-base`, `feat/done`, `card/released` (its reservation was
///   released), `release/x`
/// - preserved: `main`, `master`, `v2026.11.0`, `2026.11.0-rc.1`, `archive/old`, `integ`
/// - `feat/open` has a commit `integ` lacks; `work/linked` is checked out in a linked worktree
/// - `card/held` and `card/review` have an active reservation; `alias` is a symbolic ref
fn fixture() -> Fx {
    let home = TempHome::new();
    let project = fs::canonicalize(home.project("demo")).unwrap();
    ledger::register(home.path(), "demo").unwrap();
    git(&project, &["init", "-q", "-b", "main"]);
    fs::write(
        project.join(".gitignore"),
        format!("/{}/\n/wts/\n", moai_cli::PROJECT_STATE_DIR),
    )
    .unwrap();
    git(&project, &["add", ".gitignore"]);
    commit_file(&project, "a.txt");
    for b in [
        "master",
        "v2026.11.0",
        "2026.11.0-rc.1",
        "archive/old",
        "feat/at-base",
        "card/held",
        "card/review",
        "card/released",
        "work/linked",
        "release/x",
    ] {
        git(&project, &["branch", b]);
    }
    git(&project, &["checkout", "-q", "-b", "feat/done"]);
    commit_file(&project, "done.txt");
    git(&project, &["checkout", "-q", "-b", "feat/open", "main"]);
    commit_file(&project, "open.txt");
    git(&project, &["checkout", "-q", "-b", "integ", "main"]);
    git(&project, &["merge", "-q", "--no-ff", "-m", "land", "feat/done"]);
    git(
        &project,
        &["symbolic-ref", "refs/heads/alias", "refs/heads/feat/done"],
    );
    let wts = fs::canonicalize(home.path()).unwrap().join("wts");
    fs::create_dir_all(&wts).unwrap();
    git(
        &project,
        &[
            "worktree",
            "add",
            "-q",
            wts.join("linked").to_str().unwrap(),
            "work/linked",
        ],
    );
    let fx = Fx { home, project, wts };
    let held = fx.reserve("card/held");
    let review = fx.reserve("card/review");
    let released = fx.reserve("card/released");
    fx.ledger(&["card", "review", &review.0, "--generation", &review.1]);
    fx.ledger(&["card", "release", &released.0, "--generation", &released.1]);
    drop(held);
    fx
}

impl Fx {
    fn moai(&self, args: &[&str]) -> Output {
        self.home
            .command(args)
            .env("CLAUDE_CODE_SESSION_ID", SESSION)
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .output()
            .expect("run moai")
    }

    fn envelope(out: &Output) -> Value {
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "not JSON: {} / {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    /// A ledger command that must succeed; returns its result.
    fn ledger(&self, args: &[&str]) -> Value {
        let mut all = args.to_vec();
        let project = self.project.to_str().unwrap();
        all.extend(["--project", project]);
        let v = Self::envelope(&self.moai(&all));
        assert_eq!(v["ok"], json!(true), "{args:?}: {v}");
        v["result"].clone()
    }

    /// A new card whose reservation names `branch`: returns its name and generation.
    fn reserve(&self, branch: &str) -> (String, String) {
        let scope = format!("src/{}.rs", branch.replace('/', "_"));
        let added = self.ledger(&["card", "add", "--title", "t", "--scope", &scope]);
        let name = added["name"].as_str().unwrap().to_string();
        self.ledger(&["card", "ready", &name]);
        let taken = self.ledger(&[
            "card",
            "take",
            &name,
            "--lease-secs",
            "3600",
            "--worktree",
            self.wts.join("none").to_str().unwrap(),
            "--branch",
            branch,
        ]);
        let generation = taken["reservation"]["generation"].as_i64().unwrap();
        (name, generation.to_string())
    }

    fn plan_args<'a>(&'a self, integration_ref: &'a str) -> Vec<&'a str> {
        vec![
            "branch-gc-plan",
            self.project.to_str().unwrap(),
            "--integration-ref",
            integration_ref,
        ]
    }

    fn plan_envelope(&self, integration_ref: &str) -> Value {
        Self::envelope(&self.moai(&self.plan_args(integration_ref)))
    }

    fn plan(&self) -> Value {
        let v = self.plan_envelope("integ");
        assert_eq!(v["ok"], json!(true), "{v}");
        assert_eq!(v["operation"], json!("branch.gc.plan"), "{v}");
        v["result"].clone()
    }

    fn apply_args<'a>(&'a self, digest: &'a str, permission: &'a str) -> Vec<&'a str> {
        vec![
            "branch-gc-apply",
            self.project.to_str().unwrap(),
            "--integration-ref",
            "integ",
            "--plan-digest",
            digest,
            "--permission-reference",
            permission,
        ]
    }

    fn apply(&self, digest: &str) -> Value {
        Self::envelope(&self.moai(&self.apply_args(digest, PERMISSION)))
    }

    fn branches(&self) -> Vec<String> {
        git(
            &self.project,
            &["for-each-ref", "--format=%(refname:short)", "refs/heads"],
        )
        .lines()
        .map(str::to_string)
        .collect()
    }

    fn tip(&self, branch: &str) -> String {
        git(&self.project, &["rev-parse", "--verify", branch])
            .trim()
            .to_string()
    }
}

/// Every candidate of a plan: branch → (action, reasons).
fn rows(plan: &Value) -> BTreeMap<String, (String, Vec<String>)> {
    plan["candidates"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| {
            (
                c["branch"].as_str().unwrap().to_string(),
                (
                    c["action"].as_str().unwrap().to_string(),
                    c["reasons"]
                        .as_array()
                        .unwrap()
                        .iter()
                        .map(|r| r.as_str().unwrap().to_string())
                        .collect(),
                ),
            )
        })
        .collect()
}

fn deleted_set(plan: &Value) -> Vec<String> {
    rows(plan)
        .into_iter()
        .filter(|(_, (a, _))| a == "delete")
        .map(|(b, _)| b)
        .collect()
}

fn kept(plan: &Value, branch: &str) -> Vec<String> {
    let all = rows(plan);
    let (action, reasons) = all
        .get(branch)
        .unwrap_or_else(|| panic!("no row for {branch}: {plan}"));
    assert_eq!(action, "keep", "{branch}: {plan}");
    reasons.clone()
}

fn strs(v: &[&str]) -> Vec<String> {
    v.iter().map(|s| s.to_string()).collect()
}

// ---- AC-1: what is deleted and what is kept

#[test]
fn ac1_only_a_merged_unused_branch_is_a_delete_row() {
    let fx = fixture();
    let plan = fx.plan();
    assert_eq!(
        deleted_set(&plan),
        strs(&["card/released", "feat/at-base", "feat/done", "release/x"]),
        "{plan}"
    );
    assert_eq!(kept(&plan, "feat/open"), strs(&["not_merged"]));
    assert_eq!(kept(&plan, "work/linked"), strs(&["checked_out"]));
    assert_eq!(kept(&plan, "card/held"), strs(&["card_reserved"]));
    assert_eq!(kept(&plan, "card/review"), strs(&["card_reserved"]));
    assert_eq!(kept(&plan, "alias"), strs(&["symbolic_ref"]));
    let totals = &plan["totals"];
    assert_eq!(totals["delete"], json!(4), "{plan}");
    assert_eq!(totals["branches"], json!(fx.branches().len()), "{plan}");
    assert_eq!(plan["integration_tip"], json!(fx.tip("integ")), "{plan}");
}

#[test]
fn ac1_the_branch_checked_out_in_the_primary_checkout_is_kept() {
    let fx = fixture();
    git(&fx.project, &["checkout", "-q", "feat/at-base"]);
    let plan = fx.plan();
    assert_eq!(kept(&plan, "feat/at-base"), strs(&["checked_out"]), "{plan}");
    assert!(!deleted_set(&plan).contains(&"feat/at-base".to_string()));
}

#[test]
fn ac1_a_reservation_that_was_released_does_not_keep_its_branch() {
    let fx = fixture();
    let plan = fx.plan();
    assert_eq!(rows(&plan)["card/released"].0, "delete", "{plan}");
}

// ---- AC-2: preserved refs

#[test]
fn ac2_the_preserved_refs_are_kept_though_merged_and_unused() {
    let fx = fixture();
    let plan = fx.plan();
    for b in [
        "main",
        "master",
        "v2026.11.0",
        "2026.11.0-rc.1",
        "archive/old",
    ] {
        assert_eq!(kept(&plan, b), strs(&["preserved"]), "{b}: {plan}");
    }
    // The integration branch is preserved and, being checked out here, kept for that too.
    assert_eq!(kept(&plan, "integ"), strs(&["preserved", "checked_out"]));
}

#[test]
fn ac2_the_local_branch_the_integration_ref_names_is_preserved() {
    let fx = fixture();
    let v = fx.plan_envelope("release/x");
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(kept(&v["result"], "release/x"), strs(&["preserved"]));
}

// ---- AC-3: an unreadable profile or ledger rules no reservation out

#[test]
fn ac3_an_unreadable_profile_or_ledger_keeps_every_would_be_delete() {
    let fx = fixture();
    let profile = fx
        .project
        .join(moai_cli::PROJECT_STATE_DIR)
        .join("project.yaml");
    let good = fs::read(&profile).unwrap();
    fs::write(&profile, b"not a profile").unwrap();
    let plan = fx.plan();
    assert!(deleted_set(&plan).is_empty(), "{plan}");
    for b in ["feat/done", "feat/at-base", "card/released", "release/x"] {
        assert_eq!(kept(&plan, b), strs(&["reservation_check_unavailable"]));
    }
    // A branch that is kept for another reason keeps that reason only.
    assert_eq!(kept(&plan, "feat/open"), strs(&["not_merged"]));

    fs::write(&profile, good).unwrap();
    let file = ledger::ledger_path(fx.home.path(), "demo").unwrap();
    fs::write(&file, b"not a ledger").unwrap();
    let plan = fx.plan();
    assert!(deleted_set(&plan).is_empty(), "{plan}");
    assert_eq!(
        kept(&plan, "feat/done"),
        strs(&["reservation_check_unavailable"])
    );
}

#[test]
fn ac3_a_project_without_a_profile_has_no_reservations() {
    let fx = fixture();
    fs::remove_file(
        fx.project
            .join(moai_cli::PROJECT_STATE_DIR)
            .join("project.yaml"),
    )
    .unwrap();
    let plan = fx.plan();
    assert!(deleted_set(&plan).contains(&"card/held".to_string()), "{plan}");
}

// ---- AC-4: read-only, and the digest

/// Everything a read-only command must leave alone: refs, worktree list, the ledger file and
/// every file of the project folder outside `.git`.
fn state_of(fx: &Fx) -> (String, String, Vec<u8>, BTreeMap<String, Vec<u8>>) {
    let refs = git(
        &fx.project,
        &["for-each-ref", "--format=%(refname) %(objectname) %(symref)"],
    );
    let list = git(&fx.project, &["worktree", "list", "--porcelain"]);
    let file = ledger::ledger_path(fx.home.path(), "demo").unwrap();
    let ledger_bytes = fs::read(file).unwrap();
    fn walk(dir: &Path, root: &Path, out: &mut BTreeMap<String, Vec<u8>>) {
        for e in fs::read_dir(dir).unwrap() {
            let p = e.unwrap().path();
            if p.file_name().is_some_and(|n| n == ".git") {
                continue;
            }
            if p.is_dir() {
                walk(&p, root, out);
            } else {
                out.insert(
                    p.strip_prefix(root).unwrap().display().to_string(),
                    fs::read(&p).unwrap(),
                );
            }
        }
    }
    let mut files = BTreeMap::new();
    walk(&fx.project, &fx.project, &mut files);
    (refs, list, ledger_bytes, files)
}

#[test]
fn ac4_the_plan_writes_nothing_and_names_itself_read_only() {
    let fx = fixture();
    let before = state_of(&fx);
    let plan = fx.plan();
    assert_eq!(state_of(&fx), before);
    assert_eq!(plan["writes"], json!(false), "{plan}");
    assert_eq!(plan["network"], json!("disabled"), "{plan}");
}

#[test]
fn ac4_the_digest_follows_the_integration_tip_and_the_delete_rows_only() {
    let fx = fixture();
    let base = fx.plan()["plan_digest"].as_str().unwrap().to_string();
    assert!(base.starts_with("sha256:"), "{base}");
    assert_eq!(fx.plan()["plan_digest"], json!(base), "a second run");

    // A keep row that changes without changing the integration tip leaves the digest.
    git(&fx.project, &["checkout", "-q", "feat/open"]);
    commit_file(&fx.project, "open2.txt");
    git(&fx.project, &["checkout", "-q", "integ"]);
    assert_eq!(fx.plan()["plan_digest"], json!(base), "a keep row moved");

    // A new delete row changes it.
    git(&fx.project, &["branch", "feat/extra", "main"]);
    let with_extra = fx.plan();
    assert_ne!(with_extra["plan_digest"], json!(base));
    // A delete row that becomes a keep row changes it.
    git(&fx.project, &["checkout", "-q", "feat/extra"]);
    let kept_now = fx.plan();
    assert_ne!(kept_now["plan_digest"], with_extra["plan_digest"]);
    git(&fx.project, &["checkout", "-q", "integ"]);
    // The integration tip moving changes it, though the rows are the same.
    let before_move = fx.plan();
    commit_file(&fx.project, "next.txt");
    let after_move = fx.plan();
    assert_eq!(deleted_set(&before_move), deleted_set(&after_move));
    assert_ne!(before_move["plan_digest"], after_move["plan_digest"]);
}

// ---- AC-5: apply deletes the plan and nothing else

#[test]
fn ac5_apply_deletes_exactly_the_delete_rows() {
    let fx = fixture();
    let plan = fx.plan();
    let delete = deleted_set(&plan);
    let before = fx.branches();
    let tips: Vec<String> = delete.iter().map(|b| fx.tip(b)).collect();
    let v = fx.apply(plan["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["operation"], json!("branch.gc.apply"), "{v}");
    let r = &v["result"];
    assert_eq!(r["status"], json!("applied"), "{v}");
    assert_eq!(r["writes"], json!(true), "{v}");
    assert_eq!(r["network"], json!("disabled"), "{v}");
    let results = r["results"].as_array().unwrap();
    assert_eq!(results.len(), delete.len(), "{v}");
    for (row, (branch, tip)) in results.iter().zip(delete.iter().zip(&tips)) {
        assert_eq!(row["branch"], json!(branch), "{v}");
        assert_eq!(row["status"], json!("deleted"), "{v}");
        assert_eq!(row["tip"], json!(tip), "{v}");
    }
    let mut want: Vec<String> = before
        .into_iter()
        .filter(|b| !delete.contains(b))
        .collect();
    want.sort();
    let mut have = fx.branches();
    have.sort();
    assert_eq!(have, want);
    // The commits stay: `integ` still holds them.
    for tip in tips {
        git(&fx.project, &["merge-base", "--is-ancestor", &tip, "integ"]);
    }
    // A second run finds nothing to do.
    let again = fx.plan();
    assert!(deleted_set(&again).is_empty(), "{again}");
}

// ---- AC-6: refusals before any write

#[test]
fn ac6_a_wrong_or_stale_digest_or_no_permission_refuses_and_deletes_nothing() {
    let fx = fixture();
    let plan = fx.plan();
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let before = fx.branches();

    let wrong = fx.apply("sha256:0000");
    assert_eq!(wrong["ok"], json!(false), "{wrong}");
    assert!(
        wrong["error"]["message"].as_str().unwrap().contains("digest"),
        "{wrong}"
    );
    assert_eq!(fx.branches(), before);

    // No permission reference: the argument is required.
    let missing = fx.moai(&[
        "branch-gc-apply",
        fx.project.to_str().unwrap(),
        "--integration-ref",
        "integ",
        "--plan-digest",
        &digest,
    ]);
    assert!(!missing.status.success());
    assert_eq!(fx.branches(), before);

    // An empty one is not a permission.
    let empty = Fx::envelope(&fx.moai(&fx.apply_args(&digest, "")));
    assert_eq!(empty["ok"], json!(false), "{empty}");
    assert_eq!(fx.branches(), before);

    // A plan made before another branch became deletable is stale.
    git(&fx.project, &["branch", "feat/late", "main"]);
    let stale = fx.apply(&digest);
    assert_eq!(stale["ok"], json!(false), "{stale}");
    assert_eq!(fx.branches().len(), before.len() + 1);
    assert!(fx.branches().contains(&"feat/done".to_string()));
}

// ---- AC-7: the branch is read again right before it is deleted

#[test]
fn ac7_a_branch_that_moved_or_was_checked_out_after_the_digest_check_is_refused() {
    let fx = fixture();
    let plan = fx.plan();
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let moved_to = fx.tip("feat/open");
    let project = fx.project.clone();
    let wt = fx.wts.join("late");
    let v = moai_cli::branch_gc::apply(
        &fx.project,
        "integ",
        &digest,
        PERMISSION,
        Some(fx.home.path()),
        &|branch| match branch {
            "feat/at-base" => {
                git(&project, &["update-ref", "refs/heads/feat/at-base", &moved_to]);
            }
            "release/x" => {
                git(
                    &project,
                    &["worktree", "add", "-q", wt.to_str().unwrap(), "release/x"],
                );
            }
            _ => {}
        },
    )
    .unwrap();
    let status = |b: &str| {
        v["results"]
            .as_array()
            .unwrap()
            .iter()
            .find(|r| r["branch"] == json!(b))
            .unwrap_or_else(|| panic!("no result for {b}: {v}"))["status"]
            .clone()
    };
    assert_eq!(status("feat/at-base"), json!("refused"), "{v}");
    assert_eq!(status("release/x"), json!("refused"), "{v}");
    assert_eq!(status("feat/done"), json!("deleted"), "{v}");
    assert_eq!(v["status"], json!("partial"), "{v}");
    let have = fx.branches();
    assert!(have.contains(&"feat/at-base".to_string()));
    assert!(have.contains(&"release/x".to_string()));
    // The moved branch kept its new tip.
    assert_eq!(fx.tip("feat/at-base"), moved_to);
}

#[test]
fn ac7_a_card_or_a_checkout_after_the_plan_changes_the_digest_and_refuses_the_run() {
    let fx = fixture();
    let digest = fx.plan()["plan_digest"].as_str().unwrap().to_string();
    fx.reserve("feat/done");
    let refused = fx.apply(&digest);
    assert_eq!(refused["ok"], json!(false), "{refused}");
    assert!(fx.branches().contains(&"feat/done".to_string()));

    let digest = fx.plan()["plan_digest"].as_str().unwrap().to_string();
    git(
        &fx.project,
        &[
            "worktree",
            "add",
            "-q",
            fx.wts.join("co").to_str().unwrap(),
            "feat/at-base",
        ],
    );
    let refused = fx.apply(&digest);
    assert_eq!(refused["ok"], json!(false), "{refused}");
    assert!(fx.branches().contains(&"feat/at-base".to_string()));
}

// ---- AC-8: git branch -d, never -D

#[test]
fn ac8_a_branch_git_refuses_to_delete_stays_and_the_rest_go_on() {
    let fx = fixture();
    // HEAD is `main`: it holds C0 but not the merge of `feat/done`, so `git branch -d` calls
    // `feat/done` not fully merged although the integration branch has it.
    git(&fx.project, &["checkout", "-q", "main"]);
    let plan = fx.plan();
    assert!(deleted_set(&plan).contains(&"feat/done".to_string()), "{plan}");
    let v = fx.apply(plan["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    let r = &v["result"];
    assert_eq!(r["status"], json!("partial"), "{v}");
    let results = r["results"].as_array().unwrap();
    let by = |b: &str| results.iter().find(|x| x["branch"] == json!(b)).unwrap();
    assert_eq!(by("feat/done")["status"], json!("refused"), "{v}");
    assert!(
        by("feat/done")["detail"]
            .as_str()
            .unwrap()
            .contains("not fully merged"),
        "{v}"
    );
    assert_eq!(by("feat/at-base")["status"], json!("deleted"), "{v}");
    let have = fx.branches();
    assert!(have.contains(&"feat/done".to_string()));
    assert!(!have.contains(&"feat/at-base".to_string()));
}

// ---- AC-9: the remote is not touched

#[test]
fn ac9_apply_leaves_the_remote_and_the_remote_tracking_refs_alone() {
    let fx = fixture();
    let remote = fx.home.path().join("origin.git");
    git(&fx.project, &["init", "-q", "--bare", remote.to_str().unwrap()]);
    git(
        &fx.project,
        &["remote", "add", "origin", remote.to_str().unwrap()],
    );
    git(&fx.project, &["push", "-q", "origin", "--all"]);
    git(&fx.project, &["push", "-q", "-u", "origin", "feat/done"]);
    git(&fx.project, &["fetch", "-q", "origin"]);
    let remote_before = git(&remote, &["for-each-ref"]);
    let tracking_before = git(&fx.project, &["for-each-ref", "refs/remotes"]);
    assert!(remote_before.contains("refs/heads/feat/done"));

    let plan = fx.plan();
    let v = fx.apply(plan["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["result"]["status"], json!("applied"), "{v}");
    assert!(!fx.branches().contains(&"feat/done".to_string()));

    assert_eq!(git(&remote, &["for-each-ref"]), remote_before);
    assert_eq!(
        git(&fx.project, &["for-each-ref", "refs/remotes"]),
        tracking_before
    );
}

// ---- AC-10: help and operation names

#[test]
fn ac10_both_commands_are_in_the_help_listing_with_a_summary() {
    let fx = fixture();
    let out = fx.moai(&["help", "all"]);
    let text = String::from_utf8_lossy(&out.stdout).into_owned();
    for name in ["branch-gc-plan", "branch-gc-apply"] {
        let line = text
            .lines()
            .find(|l| l.trim_start().starts_with(name))
            .unwrap_or_else(|| panic!("no help line for {name}: {text}"));
        assert!(line.trim().len() > name.len() + 3, "{line}");
    }
}
