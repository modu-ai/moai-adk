//! SPEC-20261121-path-root-moai-default (REQ-1..REQ-10): the three state-folder resolvers answer
//! `.moai` for a tree that has no earlier `.moai-cli`, keep `.moai-cli` where it exists, never
//! write a Go folder, and read only. Every case builds its tree in a temporary folder; the real
//! HOME and the real project state are never read or written.
#![cfg(unix)]

mod common;
use common::home::TempHome;
use moai_cli::{clone_root, home_root, state_root};
use std::{fs, os::unix::fs::symlink, path::Path};

const EARLIER: &str = ".moai-cli";
const NOW: &str = ".moai";

fn write(root: &Path, rel: &str, body: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

fn marker(root: &Path) {
    write(
        root,
        ".moai/moai.json",
        std::str::from_utf8(&state_root::marker_bytes("test", &[])).unwrap(),
    );
}

fn earlier(root: &Path) {
    write(
        root,
        ".moai-cli/project.yaml",
        "{\"schema_version\":2,\"project_id\":\"p\"}\n",
    );
}

fn verdict(root: &Path) -> &'static str {
    state_root::classify(root).generation
}

fn exists(p: &Path) -> bool {
    fs::symlink_metadata(p).is_ok()
}

// ---- project ---------------------------------------------------------------------------------

#[test]
fn project_with_a_marker_resolves_to_moai_even_beside_the_earlier_folder() {
    let t = TempHome::new();
    marker(t.path());
    assert_eq!(verdict(t.path()), "rust");
    assert_eq!(state_root::dir(t.path()), NOW, "marker alone");
    earlier(t.path());
    assert_eq!(state_root::dir(t.path()), NOW, "marker beside .moai-cli");
    assert_eq!(state_root::path(t.path(), "x"), t.path().join(".moai/x"));
    assert_eq!(state_root::rel(t.path(), "x"), ".moai/x");
}

#[test]
fn project_with_only_the_earlier_folder_stays_on_it() {
    let t = TempHome::new();
    earlier(t.path());
    assert_eq!(verdict(t.path()), "none");
    assert_eq!(state_root::dir(t.path()), EARLIER);
    assert_eq!(state_root::rel(t.path(), "x"), ".moai-cli/x");
}

#[test]
fn project_with_both_folders_and_no_marker_stays_on_the_earlier_one() {
    let t = TempHome::new();
    earlier(t.path());
    write(t.path(), ".moai/project.yaml", "tracked\n");
    write(t.path(), ".moai/specs/S/spec.md", "# s\n");
    assert_eq!(verdict(t.path()), "unknown");
    assert_eq!(
        state_root::dir(t.path()),
        EARLIER,
        "a legacy tree is unchanged"
    );
    let empty = TempHome::new();
    earlier(empty.path());
    fs::create_dir(empty.path().join(".moai")).unwrap();
    assert_eq!(verdict(empty.path()), "empty");
    assert_eq!(state_root::dir(empty.path()), EARLIER);
}

#[test]
fn project_with_neither_folder_resolves_to_moai() {
    let t = TempHome::new();
    assert_eq!(verdict(t.path()), "none");
    assert_eq!(state_root::dir(t.path()), NOW);
    assert_eq!(state_root::path(t.path(), ""), t.path().join(".moai"));
    assert!(
        !exists(&t.path().join(EARLIER)),
        "a resolver creates nothing"
    );
}

#[test]
fn project_with_a_specs_only_moai_resolves_to_moai() {
    let t = TempHome::new();
    write(t.path(), ".moai/specs/S/spec.md", "# s\n");
    assert_eq!(
        verdict(t.path()),
        "none",
        "a specs-only .moai reads as absent"
    );
    assert_eq!(state_root::dir(t.path()), NOW);
}

#[test]
fn project_worktree_without_a_marker_resolves_to_moai() {
    // A fresh worktree of a repository that tracks files under .moai/ and keeps no state.
    let t = TempHome::new();
    write(t.path(), ".moai/project.yaml", "tracked\n");
    write(t.path(), ".moai/evals/e/case.md", "# e\n");
    write(t.path(), ".moai/specs/S/spec.md", "# s\n");
    assert_eq!(verdict(t.path()), "unknown");
    assert_eq!(state_root::dir(t.path()), NOW);
    assert!(!exists(&t.path().join(EARLIER)));
}

#[test]
fn project_with_go_state_and_no_earlier_folder_stays_on_the_earlier_name() {
    let t = TempHome::new();
    write(t.path(), ".moai/manifest.json", "{}");
    assert_eq!(verdict(t.path()), "go");
    assert_eq!(
        state_root::dir(t.path()),
        EARLIER,
        "Go's folder is never written"
    );
    let mixed = TempHome::new();
    marker(mixed.path());
    write(mixed.path(), ".moai/manifest.json", "{}");
    assert_eq!(verdict(mixed.path()), "mixed");
    assert_eq!(state_root::dir(mixed.path()), EARLIER);
}

#[test]
fn project_with_a_stopped_or_refused_moai_and_no_earlier_folder_stays_on_the_earlier_name() {
    let stopped = TempHome::new();
    stopped.make_interrupted_generation();
    assert_eq!(verdict(stopped.path()), "interrupted");
    assert_eq!(state_root::dir(stopped.path()), EARLIER);
    let file = TempHome::new();
    file.make_refused_generation();
    assert_eq!(verdict(file.path()), "refuse");
    assert_eq!(state_root::dir(file.path()), EARLIER);
    let link = TempHome::new();
    let elsewhere = TempHome::new();
    symlink(elsewhere.path(), link.path().join(".moai")).unwrap();
    assert_eq!(verdict(link.path()), "refuse");
    assert_eq!(state_root::dir(link.path()), EARLIER);
}

#[test]
fn project_counts_an_earlier_folder_that_is_a_link_as_present_without_following_it() {
    let t = TempHome::new();
    symlink(t.path().join("nowhere"), t.path().join(EARLIER)).unwrap();
    assert!(!t.path().join(EARLIER).exists(), "the link dangles");
    assert_eq!(state_root::dir(t.path()), EARLIER);
    let target = TempHome::new();
    earlier(target.path());
    let linked = TempHome::new();
    symlink(target.path().join(EARLIER), linked.path().join(EARLIER)).unwrap();
    assert_eq!(state_root::dir(linked.path()), EARLIER);
    assert!(
        fs::symlink_metadata(linked.path().join(EARLIER))
            .unwrap()
            .file_type()
            .is_symlink(),
        "the link is still a link"
    );
}

#[test]
fn project_new_project_still_gets_its_marker_once() {
    let t = TempHome::new();
    assert!(state_root::is_new_project(t.path()));
    assert_eq!(state_root::create_dir(t.path()), NOW);
    assert_eq!(state_root::adopt_new(t.path(), "test"), Ok(true));
    assert_eq!(verdict(t.path()), "rust");
    assert_eq!(state_root::dir(t.path()), NOW);
    assert!(t.path().join(".moai/moai.json").is_file());
    assert_eq!(
        state_root::adopt_new(t.path(), "test"),
        Ok(false),
        "only once"
    );
    let specs = TempHome::new();
    write(specs.path(), ".moai/specs/S/spec.md", "# s\n");
    assert_eq!(state_root::adopt_new(specs.path(), "test"), Ok(true));
    assert!(
        specs.path().join(".moai/specs/S/spec.md").is_file(),
        "specs kept"
    );
    let worktree = TempHome::new();
    write(worktree.path(), ".moai/project.yaml", "tracked\n");
    assert!(!state_root::is_new_project(worktree.path()));
    assert_eq!(state_root::adopt_new(worktree.path(), "test"), Ok(false));
    assert!(
        !worktree.path().join(".moai/moai.json").exists(),
        "no marker for a non-new project"
    );
}

// ---- HOME ------------------------------------------------------------------------------------

#[test]
fn home_empty_resolves_to_moai() {
    let h = TempHome::new();
    assert_eq!(home_root::dir(h.path()), NOW);
    assert_eq!(home_root::path(h.path()), h.path().join(".moai"));
    assert!(
        !exists(&h.path().join(EARLIER)),
        "a resolver creates nothing"
    );
}

#[test]
fn home_with_only_its_earlier_folder_stays_on_it() {
    let h = TempHome::new();
    earlier(h.path());
    assert_eq!(home_root::dir(h.path()), EARLIER);
    assert_eq!(home_root::path(h.path()), h.path().join(".moai-cli"));
}

#[test]
fn home_with_both_folders_and_no_marker_stays_on_the_earlier_one() {
    let h = TempHome::new();
    earlier(h.path());
    write(h.path(), ".moai/reports/r.md", "# r\n");
    assert_eq!(home_root::dir(h.path()), EARLIER);
    let go = TempHome::new();
    earlier(go.path());
    go.make_go_generation();
    assert_eq!(home_root::dir(go.path()), EARLIER);
}

#[test]
fn home_with_unrecognised_files_in_moai_and_no_earlier_folder_stays_on_the_earlier_name() {
    // A Go-like HOME whose items the HOME Go markers do not list: not ours to write into.
    let h = TempHome::new();
    write(h.path(), ".moai/config.yaml", "generation: go\n");
    write(h.path(), ".moai/cache/blob", "go state bytes");
    assert_eq!(state_root::classify_home(h.path()).generation, "unknown");
    assert_eq!(home_root::dir(h.path()), EARLIER);
    assert_eq!(home_root::path(h.path()), h.path().join(".moai-cli"));
}

#[test]
fn home_with_only_names_this_generation_writes_in_moai_resolves_to_moai() {
    // A first write on a fresh HOME made .moai before any marker; the next call must find it.
    let h = TempHome::new();
    write(h.path(), ".moai/home-style/receipts/r.json", "{}");
    write(h.path(), ".moai/installations/i.json", "{}");
    write(h.path(), ".moai/projects/p/ledger.db", "x");
    write(h.path(), ".moai/build/b", "x");
    write(h.path(), ".moai/edit-diagnostics/e", "x");
    assert_eq!(state_root::classify_home(h.path()).generation, "unknown");
    assert_eq!(home_root::dir(h.path()), NOW);
    assert_eq!(home_root::path(h.path()), h.path().join(".moai"));
    // One name that is not this generation's is enough to leave it alone.
    write(h.path(), ".moai/kanban/k", "x");
    assert_eq!(home_root::dir(h.path()), EARLIER);
}

#[test]
fn home_with_a_go_state_folder_and_no_earlier_folder_stays_on_the_earlier_name() {
    let h = TempHome::new();
    h.make_go_generation();
    assert_eq!(
        home_root::dir(h.path()),
        EARLIER,
        "Go's ~/.moai is never written"
    );
    let mixed = TempHome::new();
    mixed.make_mixed_generation();
    assert_eq!(
        home_root::dir(mixed.path()),
        EARLIER,
        "schema 1 beside Go files is mixed"
    );
    let stopped = TempHome::new();
    stopped.make_interrupted_generation();
    assert_eq!(home_root::dir(stopped.path()), EARLIER);
    let file = TempHome::new();
    file.make_refused_generation();
    assert_eq!(home_root::dir(file.path()), EARLIER);
}

#[test]
fn home_with_a_marker_resolves_to_moai() {
    let h = TempHome::new();
    h.make_rust_generation();
    assert_eq!(home_root::dir(h.path()), NOW);
    earlier(h.path());
    assert_eq!(
        home_root::dir(h.path()),
        NOW,
        "marker beside the earlier folder"
    );
    let v2 = TempHome::new();
    v2.make_go_generation();
    write(
        v2.path(),
        ".moai/moai.json",
        std::str::from_utf8(&state_root::home_marker_bytes("test", "stamp")).unwrap(),
    );
    assert_eq!(home_root::dir(v2.path()), NOW);
    assert_eq!(
        home_root::path(v2.path()),
        moai_cli::home_layout::rust_path(v2.path(), "")
    );
}

#[test]
fn home_link_as_the_earlier_folder_counts_as_present() {
    let h = TempHome::new();
    symlink(h.path().join("nowhere"), h.path().join(EARLIER)).unwrap();
    assert_eq!(home_root::dir(h.path()), EARLIER);
}

#[test]
fn home_install_inspect_in_an_empty_temporary_home_reports_moai() {
    let h = TempHome::new();
    let out = h
        .command_for(env!("CARGO_BIN_EXE_moai-cli"), &["install-inspect"])
        .stdin(std::process::Stdio::null())
        .output()
        .unwrap();
    let text = String::from_utf8(out.stdout).unwrap();
    let json: serde_json::Value = serde_json::from_str(
        text.lines()
            .find(|l| l.starts_with('{'))
            .expect("a JSON line"),
    )
    .unwrap();
    let root = json["result"]["home_state_root"]
        .as_str()
        .unwrap_or_else(|| panic!("no home_state_root in {json}"));
    assert!(root.ends_with("/.moai"), "reported {root}");
    assert!(
        !exists(&h.path().join(EARLIER)) && !exists(&h.path().join(NOW)),
        "it only reports"
    );
}

// ---- clone-shared state ----------------------------------------------------------------------

#[test]
fn clone_with_nothing_resolves_to_moai() {
    let c = TempHome::new();
    assert_eq!(clone_root::dir(c.path()), NOW);
    assert_eq!(clone_root::path(c.path()), c.path().join(".moai"));
}

#[test]
fn clone_with_only_the_earlier_folder_stays_on_it() {
    let c = TempHome::new();
    fs::create_dir_all(c.path().join(".moai-cli/claims")).unwrap();
    assert_eq!(clone_root::dir(c.path()), EARLIER);
}

#[test]
fn clone_with_a_marker_resolves_to_moai() {
    let c = TempHome::new();
    marker(c.path());
    assert_eq!(clone_root::dir(c.path()), NOW);
    fs::create_dir_all(c.path().join(".moai-cli/claims")).unwrap();
    assert_eq!(
        clone_root::dir(c.path()),
        NOW,
        "marker beside the earlier folder"
    );
}

#[test]
fn clone_with_files_in_moai_and_no_marker_and_no_earlier_folder_resolves_to_moai() {
    // The common folder is inside .git: no other tool writes there, and a first write that
    // started on .moai before any marker must be found again by the next call.
    let c = TempHome::new();
    write(c.path(), ".moai/sessions/agent-1.live", "x\n");
    assert_eq!(state_root::classify(c.path()).generation, "unknown");
    assert_eq!(clone_root::dir(c.path()), NOW);
}

#[test]
fn clone_with_an_empty_moai_follows_the_earlier_folder_only_where_it_exists() {
    let beside = TempHome::new();
    fs::create_dir_all(beside.path().join(".moai-cli/claims")).unwrap();
    fs::create_dir(beside.path().join(".moai")).unwrap();
    assert_eq!(
        clone_root::dir(beside.path()),
        EARLIER,
        "an empty .moai is not ours"
    );
    let alone = TempHome::new();
    fs::create_dir(alone.path().join(".moai")).unwrap();
    assert_eq!(clone_root::dir(alone.path()), NOW);
}

// ---- read only -------------------------------------------------------------------------------

#[test]
fn reads_only_and_never_creates_the_earlier_folder() {
    let trees: Vec<TempHome> = (0..8).map(|_| TempHome::new()).collect();
    earlier(trees[1].path());
    marker(trees[2].path());
    write(trees[3].path(), ".moai/project.yaml", "tracked\n");
    trees[4].make_go_generation();
    trees[5].make_mixed_generation();
    trees[6].make_interrupted_generation();
    symlink(
        trees[7].path().join("nowhere"),
        trees[7].path().join(EARLIER),
    )
    .unwrap();
    for t in &trees {
        let before = t.snapshot();
        let had_earlier = exists(&t.path().join(EARLIER));
        let _ = (
            state_root::dir(t.path()),
            state_root::path(t.path(), "a/b"),
            state_root::rel(t.path(), "a/b"),
            state_root::create_dir(t.path()),
            state_root::is_new_project(t.path()),
            home_root::dir(t.path()),
            home_root::path(t.path()),
            home_root::is_new_home(t.path()),
            clone_root::dir(t.path()),
            clone_root::path(t.path()),
        );
        assert_eq!(
            t.snapshot(),
            before,
            "a resolver call changed {:?}",
            t.path()
        );
        assert_eq!(
            exists(&t.path().join(EARLIER)),
            had_earlier,
            "{:?}",
            t.path()
        );
    }
}
