//! #2286 row 35 removes the external skill installer, including hidden CLI aliases.
//! Every invocation uses a synthetic HOME and project; existing Go skill bytes stay intact.
mod common;
use common::CleanEnv;
use std::{
    fs,
    path::Path,
    process::{Command, Output, Stdio},
};

const REMOVED: [&str; 3] = [
    "skill-source-plan",
    "skill-source-apply",
    "skill-read-receipt",
];

fn run(home: &Path, project: &Path, args: &[&str]) -> Output {
    let mut command = Command::new(crate::common::MOAI_BIN);
    command
        .clean_env()
        .env("HOME", home)
        .current_dir(project)
        .args(args)
        .stdin(Stdio::null());
    for key in ["SystemRoot", "SystemDrive", "COMSPEC", "PATHEXT"] {
        if let Some(value) = std::env::var_os(key) {
            command.env(key, value);
        }
    }
    command.output().unwrap()
}

fn snapshot(root: &Path) -> Vec<(String, bool, Vec<u8>)> {
    fn walk(root: &Path, dir: &Path, result: &mut Vec<(String, bool, Vec<u8>)>) {
        for entry in fs::read_dir(dir).unwrap() {
            let path = entry.unwrap().path();
            let is_dir = path.is_dir();
            result.push((
                path.strip_prefix(root)
                    .unwrap()
                    .to_str()
                    .unwrap()
                    .to_owned(),
                is_dir,
                if is_dir {
                    Vec::new()
                } else {
                    fs::read(&path).unwrap()
                },
            ));
            if is_dir {
                walk(root, &path, result);
            }
        }
    }
    let mut result = Vec::new();
    walk(root, root, &mut result);
    result.sort();
    result
}

#[test]
fn removed_installers_are_unknown_and_preserve_home_and_project_bytes() {
    let root = tempfile::tempdir().unwrap();
    let home = root.path().join("home");
    let project = root.path().join("project");
    fs::create_dir_all(&project).unwrap();
    fs::write(project.join("notes.txt"), b"keep project bytes\n").unwrap();
    for host in [".claude", ".agents"] {
        let skill = home.join(host).join("skills/moai");
        fs::create_dir_all(&skill).unwrap();
        fs::write(
            skill.join("SKILL.md"),
            b"Go-era user skill: preserve verbatim\n",
        )
        .unwrap();
    }
    let before = snapshot(root.path());
    let p = project.to_str().unwrap();
    let dir = home.join(".claude/skills/moai");
    let cases = [
        vec![REMOVED[0], p],
        vec![
            REMOVED[1],
            p,
            "--plan-digest",
            "synthetic",
            "--permission-reference",
            "fixture",
        ],
        vec![
            REMOVED[2],
            p,
            "--dir",
            dir.to_str().unwrap(),
            "--name",
            "moai",
        ],
    ];
    for args in cases {
        let out = run(&home, &project, &args);
        assert_eq!(out.status.code(), Some(2), "{}: {:?}", args[0], out);
        assert!(
            out.stdout.is_empty(),
            "{} returned an operation envelope",
            args[0]
        );
        assert!(
            String::from_utf8_lossy(&out.stderr).contains("unrecognized subcommand"),
            "{}: {:?}",
            args[0],
            out
        );
        assert_eq!(
            snapshot(root.path()),
            before,
            "{} changed fixture state",
            args[0]
        );
    }
}

#[test]
fn help_excludes_removed_installers_and_keeps_search_atlas_and_project_generation() {
    let root = tempfile::tempdir().unwrap();
    // `help --all` lists every command (#2802 S1); the root help shows only the person-facing ones.
    let out = run(root.path(), root.path(), &["help", "--all"]);
    assert!(out.status.success(), "{:?}", out);
    let help = String::from_utf8(out.stdout).unwrap();
    for removed in REMOVED {
        assert!(
            !help.contains(removed),
            "removed installer advertised: {removed}"
        );
    }
    for preserved in [
        "search",
        "atlas-map",
        "atlas-graph",
        "project-skill-plan",
        "generation-plan",
    ] {
        assert!(
            help.lines()
                .any(|line| line.split_whitespace().next() == Some(preserved)),
            "missing preserved command: {preserved}"
        );
    }
}
