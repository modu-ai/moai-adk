#[path = "../src/migrate.rs"]
mod migrate;

use migrate::*;
use moai_cli::apply;
use std::fs;
use tempfile::TempDir;

fn fixture() -> (TempDir, SyntheticHome, Selection) {
    let d = TempDir::new().unwrap();
    fs::create_dir_all(d.path().join(".moai/config")).unwrap();
    fs::write(
        d.path().join(".moai/config/language.json"),
        b"{\"language\":\"ko\"}",
    )
    .unwrap();
    let home = SyntheticHome::new(d.path()).unwrap();
    let item = Selection {
        source: ".moai/config/language.json".into(),
        target: ".moai-cli/imports/language.json".into(),
        recorded_sha256: Some(hash(b"{\"language\":\"ko\"}")),
        project_current: None,
    };
    (d, home, item)
}

fn preview(h: &SyntheticHome, items: &[Selection]) -> Plan {
    h.plan(
        &[Host::Claude, Host::Codex, Host::Glm, Host::OnPrem],
        Mode::Default,
        items,
    )
    .unwrap()
}

#[test]
fn global_baseline_project_conflict_and_spec_selection_are_explicit() {
    let (d, h, mut item) = fixture();
    fs::create_dir_all(d.path().join("project/.moai/config")).unwrap();
    fs::write(
        d.path().join("project/.moai/config/language.json"),
        b"project override",
    )
    .unwrap();
    item.project_current = Some("project/.moai/config/language.json".into());
    let before = moai_cli::tree_digest(d.path(), &[]).unwrap();
    let p = preview(&h, std::slice::from_ref(&item));
    assert_eq!(p.entries[0].action, "conflict");
    assert!(p.entries[0].project_sha256.is_some());
    assert!(h.apply(&p, &p.digest).is_err());
    assert_eq!(before, moai_cli::tree_digest(d.path(), &[]).unwrap());
    fs::create_dir_all(d.path().join(".moai/specs")).unwrap();
    fs::write(
        d.path().join(".moai/specs/SPEC-1.md"),
        b"# Selected specification",
    )
    .unwrap();
    let spec = Selection {
        source: ".moai/specs/SPEC-1.md".into(),
        target: ".moai-cli/imports/SPEC-1.md".into(),
        recorded_sha256: Some(hash(b"# Selected specification")),
        project_current: None,
    };
    let p = preview(&h, &[spec]);
    h.apply(&p, &p.digest).unwrap();
    assert!(!d.path().join(&item.target).exists());
    assert!(d.path().join(".moai/specs/SPEC-1.md").is_file());
}

#[test]
fn preexisting_backup_and_changed_recovery_plan_never_overwrite() {
    let (d, h, item) = fixture();
    let p = preview(&h, &[item]);
    let r = h.apply(&p, &p.digest).unwrap();
    let recovery = h.recovery_plan(&r).unwrap();
    fs::write(
        d.path().join(".moai-cli/imports/language.json"),
        b"late edit",
    )
    .unwrap();
    assert!(h.recover(&recovery, &recovery.digest).is_err());
    assert_eq!(
        fs::read(d.path().join(".moai-cli/imports/language.json")).unwrap(),
        b"late edit"
    );
    assert!(h.apply(&p, &p.digest).is_err());
}

#[test]
fn preview_writes_nothing_and_requires_final_summary_confirmation() {
    let (d, h, item) = fixture();
    let before = moai_cli::tree_digest(d.path(), &[]).unwrap();
    let p = preview(&h, &[item]);
    assert_eq!(before, moai_cli::tree_digest(d.path(), &[]).unwrap());
    assert!(p.summary.contains("confirmation"));
    assert!(h.apply(&p, "wrong").is_err());
    assert_eq!(before, moai_cli::tree_digest(d.path(), &[]).unwrap());
}

#[test]
fn verified_backup_selected_copy_and_restore_keep_go_source() {
    let (d, h, item) = fixture();
    let p = preview(&h, &[item]);
    let r = h.apply(&p, &p.digest).unwrap();
    let target = d.path().join(".moai-cli/imports/language.json");
    assert_eq!(
        fs::read(&target).unwrap(),
        fs::read(d.path().join(".moai/config/language.json")).unwrap()
    );
    assert!(d
        .path()
        .join(".moai-go-backup")
        .join(&p.digest)
        .join("manifest.json")
        .is_file());
    let recovery = h.recovery_plan(&r).unwrap();
    assert!(h.recover(&recovery, "wrong").is_err());
    assert_eq!(
        h.recover(&recovery, &recovery.digest).unwrap(),
        vec!["removed_import"]
    );
    assert!(!target.exists());
    assert!(d.path().join(".moai/config/language.json").is_file());
}

#[test]
fn user_edit_after_apply_is_preserved_by_recovery() {
    let (d, h, item) = fixture();
    let p = preview(&h, &[item]);
    let r = h.apply(&p, &p.digest).unwrap();
    let target = d.path().join(".moai-cli/imports/language.json");
    fs::write(&target, b"user edit").unwrap();
    let recovery = h.recovery_plan(&r).unwrap();
    assert_eq!(
        h.recover(&recovery, &recovery.digest).unwrap(),
        vec!["conflict"]
    );
    assert_eq!(fs::read(&target).unwrap(), b"user edit");
}

#[test]
fn existing_project_settings_conflict_and_unknown_or_modified_sources_are_preserved() {
    let (d, h, mut item) = fixture();
    fs::create_dir_all(d.path().join(".moai-cli/imports")).unwrap();
    fs::write(d.path().join(&item.target), b"project setting").unwrap();
    let p = preview(&h, &[item.clone()]);
    assert_eq!(p.entries[0].action, "conflict");
    assert!(h.apply(&p, &p.digest).is_err());
    item.recorded_sha256 = None;
    assert_eq!(preview(&h, &[item.clone()]).entries[0].ownership, "unknown");
    item.recorded_sha256 = Some(hash(b"old release"));
    let p = preview(&h, &[item]);
    assert_eq!(p.entries[0].ownership, "modified");
    assert_eq!(p.entries[0].action, "preserved");
}

#[test]
fn stale_plan_tampered_plan_and_cross_home_replay_are_refused() {
    let (d, h, item) = fixture();
    let p = preview(&h, &[item]);
    let (_other, other, _) = fixture();
    assert!(other.apply(&p, &p.digest).is_err());
    let mut tampered = p.clone();
    tampered.summary.push_str("altered");
    assert!(h.apply(&tampered, &p.digest).is_err());
    fs::write(d.path().join(".moai/config/language.json"), b"changed").unwrap();
    assert!(h.apply(&p, &p.digest).is_err());
    assert!(!d.path().join(".moai-go-backup").exists());
}

#[test]
fn secrets_traversal_duplicates_and_unsafe_mode_are_not_silently_applied() {
    let (d, h, item) = fixture();
    for source in [
        "../escape",
        ".moai/.env.glm",
        ".moai/db/data",
        ".moai/worktrees/x",
        ".moai/config/../key",
    ] {
        let bad = Selection {
            source: source.into(),
            ..item.clone()
        };
        assert!(h.plan(&[Host::Glm], Mode::Default, &[bad]).is_err());
    }
    assert!(h
        .plan(
            &[Host::Claude],
            Mode::Default,
            &[item.clone(), item.clone()]
        )
        .is_err());
    let automatic = h
        .plan(
            &[Host::OnPrem],
            Mode::Automatic,
            std::slice::from_ref(&item),
        )
        .unwrap();
    assert!(automatic.summary.contains("isolation"));
    assert!(automatic.summary.contains("does not change"));
    assert!(h.apply(&automatic, "not confirmed").is_err());
    fs::write(
        d.path().join(&item.source),
        b"{\"api_key\":\"synthetic-secret\"}",
    )
    .unwrap();
    assert!(h.plan(&[Host::Glm], Mode::Default, &[item]).is_err());
}

#[cfg(unix)]
#[test]
fn symlinks_special_files_and_corrupt_backups_are_refused() {
    use std::os::unix::fs::symlink;
    let (d, h, item) = fixture();
    let p = preview(&h, std::slice::from_ref(&item));
    let r = h.apply(&p, &p.digest).unwrap();
    let backup = d.path().join(".moai-go-backup").join(&p.digest).join("0");
    fs::write(backup, b"corrupt").unwrap();
    assert!(h.recovery_plan(&r).is_err());
    let link = d.path().join(".moai/config/link");
    symlink(d.path().join(&item.source), &link).unwrap();
    assert!(preview_result(
        &h,
        Selection {
            source: ".moai/config/link".into(),
            ..item
        }
    )
    .is_err());
    let bad = Selection {
        source: ".moai/config".into(),
        target: ".moai-cli/imports/directory".into(),
        recorded_sha256: None,
        project_current: None,
    };
    assert!(preview_result(&h, bad).is_err());
}

#[cfg(unix)]
fn preview_result(h: &SyntheticHome, item: Selection) -> Result<Plan, String> {
    h.plan(&[Host::Claude], Mode::Default, &[item])
}
