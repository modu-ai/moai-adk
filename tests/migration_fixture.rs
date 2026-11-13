use moai_cli::migrate::*;
use std::fs;
use tempfile::TempDir;

#[test]
fn credential_container_and_token_name_variants_are_not_clearance() {
    for bytes in [
        br#"{"clientCredentials":{"value":"synthetic-audit-value"}}"#.as_slice(),
        br#"{"clientToken":"synthetic-audit-value"}"#,
        br#"{"tokenValue":"synthetic-audit-value"}"#,
        br#"{"\u0063lient\u0043redentials":{"value":"synthetic-audit-value"}}"#,
    ] {
        excluded(bytes, "config.json");
    }
}

// AUDIT-2016-1340: serde_json::Value keeps only the last duplicate key, while
// backup/import copy the raw bytes. The parser must not be the only witness.
#[test]
fn duplicate_json_string_must_not_hide_escaped_secret() {
    let bytes = br#"{"ordinary":"\u0073ecrets","ordinary":"ko"}"#;
    assert_eq!(
        serde_json::from_slice::<serde_json::Value>(bytes).unwrap()["ordinary"],
        "ko"
    );
    excluded(bytes, "config.json");
}

#[test]
fn duplicate_json_object_must_not_hide_nested_credentials() {
    let bytes = br#"{"ordinary":{"\u0063redentials":{"value":"synthetic-audit-value"}},"ordinary":{"language":"ko"}}"#;
    assert!(
        serde_json::from_slice::<serde_json::Value>(bytes).unwrap()["ordinary"]
            .get("credentials")
            .is_none()
    );
    excluded(bytes, "config.json");
}

#[test]
fn nested_duplicate_json_key_hiding_escaped_token_is_excluded() {
    for bytes in [
        br#"{"a":[{"k":"\u0073ecret","k":"x"}]}"#.as_slice(),
        br#"{"k":"\u0074oken","k":1}"#,
    ] {
        excluded(bytes, "config.json");
    }
}

// Kills a mutant that drops the duplicate-key refusal: no value here is
// sensitive, so only the duplicate rule refuses these documents.
#[test]
fn benign_duplicate_keys_are_refused_before_any_write() {
    for (bytes, name) in [
        (
            br#"{"language":"ko","language":"en"}"#.as_slice(),
            "config.json",
        ),
        (
            br#"{"outer":[{"language":"ko","language":"en"}]}"#,
            "config.json",
        ),
        (b"language: ko\nlanguage: en", "config.yml"),
        (br#"{"k":1,"k":2}"#, "config.json"),
        (b"\"k\": a\nk: b", "config.yml"),
    ] {
        let (d, h, s) = independent_fixture(bytes, name);
        let before = moai_cli::tree_digest(d.path(), &[]).unwrap();
        let err = h.plan(&[Host::Claude], Mode::Default, &[s]).unwrap_err();
        assert!(err.contains("duplicate"), "{name}: {err}");
        assert_eq!(before, moai_cli::tree_digest(d.path(), &[]).unwrap());
    }
}

#[test]
fn duplicate_yaml_key_hiding_escaped_secret_is_excluded() {
    excluded(b"ordinary: \"\\u0073ecrets\"\nordinary: ko", "config.yml");
}

// toml_edit 0.25.3 must reject every duplicate definition form, so a TOML
// document can never hide an earlier value from the decoded walk.
#[test]
fn toml_duplicate_definitions_are_parser_errors_and_never_copied() {
    for text in [
        "k = \"\\u0073ecrets\"\nk = \"ko\"",
        "[t]\nk = \"\\u0073ecrets\"\n[t]\nk = \"ko\"",
        "t = { k = \"\\u0073ecrets\", k = \"ko\" }",
        "t.k = \"\\u0073ecrets\"\nt.k = \"ko\"",
    ] {
        assert!(text.parse::<toml_edit::DocumentMut>().is_err(), "{text}");
        excluded(text.as_bytes(), "config.toml");
    }
}

// The raw source filter refuses these substrings; an escape hides them from
// that filter, so the decoded policy must cover at least the same words.
#[test]
fn escaped_raw_filter_words_are_excluded_after_decoding() {
    for word in [
        "db\\u0070assword",
        "my\\u0061pi_key",
        "x\\u0061pi-key",
        "my\\u0061pikey",
        "my\\u0073ecret",
        "\\u0062earer ",
        "my\\u0061ccess_token",
        "my\\u0061uth_token",
        "\\u0070rivate key",
    ] {
        let json = format!(r#"{{"{word}":"synthetic-audit-value"}}"#);
        excluded(json.as_bytes(), "config.json");
        let toml = format!(r#""{word}" = "synthetic-audit-value""#);
        excluded(toml.as_bytes(), "config.toml");
        let yaml = format!(r#""{word}": "synthetic-audit-value""#);
        excluded(yaml.as_bytes(), "config.yml");
        let value = format!(r#"{{"ordinary":"{word}"}}"#);
        excluded(value.as_bytes(), "config.json");
    }
}

// The raw-text pass misses this spelling; only the decoded key walk sees it.
#[test]
fn decoded_key_without_escape_is_excluded() {
    excluded(br#"{"mytoken":"synthetic-audit-value"}"#, "config.json");
    excluded(b"mytoken = \"synthetic-audit-value\"", "config.toml");
}

// Boundary pair for the 262144-byte source limit (literal values on purpose).
#[test]
fn source_size_limit_accepts_262144_and_refuses_262145_bytes() {
    for (len, accepted) in [(262_144usize, true), (262_145, false)] {
        let mut bytes = br#"{"language":"ko"}"#.to_vec();
        bytes.resize(len, b' ');
        let (_d, h, s) = independent_fixture(&bytes, "config.json");
        let plan = h.plan(&[Host::Claude], Mode::Default, &[s]);
        assert_eq!(plan.is_ok(), accepted, "len {len}: {plan:?}");
    }
}

// Boundary pair for the 128-selection limit (literal values on purpose).
#[test]
fn selection_limit_accepts_128_and_refuses_129() {
    for (count, accepted) in [(128usize, true), (129, false)] {
        let (_d, h, s) = independent_fixture(br#"{"language":"ko"}"#, "config.json");
        let items: Vec<_> = (0..count)
            .map(|i| Selection {
                target: format!(".moai-cli/imports/config-{i}.json"),
                ..s.clone()
            })
            .collect();
        let plan = h.plan(&[Host::Claude], Mode::Default, &items);
        assert_eq!(plan.is_ok(), accepted, "count {count}");
    }
}

#[test]
fn decoded_policy_covers_nested_keys_values_arrays_and_tables() {
    for (bytes, name) in [
        (
            br#"{"nested":[{"\u0063redentials":{"value":"synthetic"}}]}"#.as_slice(),
            "config.json",
        ),
        (br#"{"ordinary":["\u0073ecrets"]}"#, "config.json"),
        (
            br#"[nested]
"\u0061pi_key" = "synthetic""#,
            "config.toml",
        ),
        (
            br#"[[nested]]
value = "\u0073ecrets""#,
            "config.toml",
        ),
        (
            br#"nested = [{"\u0063redentials" = "synthetic"}]"#,
            "config.toml",
        ),
        (br#""\u0061pi_key": "synthetic""#, "config.yml"),
        (br#"ordinary: "\u0073ecrets""#, "config.yml"),
    ] {
        excluded(bytes, name);
    }
}

#[test]
fn unclassifiable_formats_are_rejected_before_any_copy() {
    for (bytes, name) in [
        (br#"{"language":"ko",}"#.as_slice(), "config.json"),
        (br#""\q" = "synthetic""#, "config.toml"),
        (br#""\x61pi_key": "synthetic""#, "config.yml"),
        (b"nested:\n  language: ko", "config.yml"),
        (b"language: &alias ko", "config.yml"),
        (b"opaque input", "config.unknown"),
    ] {
        let (d, h, s) = independent_fixture(bytes, name);
        let before = moai_cli::tree_digest(d.path(), &[]).unwrap();
        assert!(h.plan(&[Host::Claude], Mode::Default, &[s]).is_err());
        assert_eq!(before, moai_cli::tree_digest(d.path(), &[]).unwrap());
    }
}

#[test]
fn ordinary_settings_remain_migratable_in_each_supported_format() {
    for (bytes, name) in [
        (
            br#"{"language":"ko","max_tokens":4096,"tokenizer":"default"}"#.as_slice(),
            "config.json",
        ),
        (
            b"language = \"ko\"\nmax_tokens = 4096\ntokenizer = \"default\"",
            "config.toml",
        ),
        (
            b"language: ko\nmax_tokens: 4096\ntokenizer: default",
            "config.yml",
        ),
    ] {
        let (d, h, s) = independent_fixture(bytes, name);
        let target = s.target.clone();
        let source = s.source.clone();
        let p = h.plan(&[Host::Claude], Mode::Default, &[s]).unwrap();
        assert_eq!(p.entries[0].action, "create");
        h.apply(&p, &p.digest).unwrap();
        assert_eq!(fs::read(d.path().join(target)).unwrap(), bytes);
        assert_eq!(fs::read(d.path().join(source)).unwrap(), bytes);
    }
}

fn independent_fixture(bytes: &[u8], name: &str) -> (tempfile::TempDir, SyntheticHome, Selection) {
    let d = tempfile::tempdir().unwrap();
    fs::create_dir_all(d.path().join(".moai/config")).unwrap();
    let source = format!(".moai/config/{name}");
    fs::write(d.path().join(&source), bytes).unwrap();
    let h = SyntheticHome::new(d.path()).unwrap();
    let s = Selection {
        source,
        target: ".moai-cli/imports/config.json".into(),
        recorded_sha256: Some(hash(bytes)),
        project_current: None,
    };
    (d, h, s)
}
fn excluded(bytes: &[u8], name: &str) {
    let (d, h, s) = independent_fixture(bytes, name);
    let before = moai_cli::tree_digest(d.path(), &[]).unwrap();
    let p = h.plan(&[Host::Glm], Mode::Default, &[s]);
    if let Ok(p) = p {
        let result = h.apply(&p, &p.digest);
        assert!(
            result.is_err(),
            "credential copied: import={} backup={}",
            d.path().join(".moai-cli/imports/config.json").exists(),
            d.path().join(".moai-go-backup").exists()
        );
    }
    assert_eq!(
        before,
        moai_cli::tree_digest(d.path(), &[]).unwrap(),
        "excluded input caused writes"
    );
}
#[test]
fn raw_token_excluded() {
    excluded(b"token: synthetic-audit-value", "config.yml");
}
#[test]
fn escaped_token_excluded() {
    excluded(br#"{"\u0074oken":"synthetic-audit-value"}"#, "config.json");
}
#[test]
fn env_path_excluded() {
    excluded(b"UNRELATED=synthetic-audit-value", ".env.glm");
}
#[test]
fn credentials_path_excluded() {
    excluded(b"synthetic-audit-value", "credentials.json");
}
#[test]
fn secrets_path_excluded() {
    excluded(b"synthetic-audit-value", "secrets.json");
}
#[test]
fn env_content_excluded() {
    excluded(b"ANTHROPIC_AUTH_TOKEN=synthetic-audit-value", "config.env");
}
#[test]
fn raw_secrets_content_excluded() {
    excluded(
        br#"{"secrets":{"x":"synthetic-audit-value"}}"#,
        "config.json",
    );
}
#[test]
fn credentials_content_excluded() {
    excluded(
        br#"{"credentials":{"username":"synthetic-user","value":"synthetic-audit-value"}}"#,
        "config.json",
    );
}
#[test]
fn escaped_api_key_excluded() {
    let b = br#"{"\u0061pi_key":"synthetic-audit-value"}"#;
    assert!(serde_json::from_slice::<serde_json::Value>(b)
        .unwrap()
        .get("api_key")
        .is_some());
    excluded(b, "config.json");
}
#[test]
fn escaped_secrets_excluded() {
    let b = br#"{"\u0073ecrets":{"x":"synthetic-audit-value"}}"#;
    assert!(serde_json::from_slice::<serde_json::Value>(b)
        .unwrap()
        .get("secrets")
        .is_some());
    excluded(b, "config.json");
}
#[test]
fn quoted_toml_token_excluded() {
    let b = br#""\u0074oken" = "synthetic-audit-value""#;
    assert!(std::str::from_utf8(b)
        .unwrap()
        .parse::<toml_edit::DocumentMut>()
        .unwrap()
        .get("token")
        .is_some());
    excluded(b, "config.toml");
}
#[test]
fn ordinary_config_import_and_preexisting_copy_are_safe() {
    let (d, h, s) = independent_fixture(b"language: ko", "language.yml");
    fs::create_dir_all(d.path().join(".moai-cli/imports")).unwrap();
    fs::write(d.path().join(&s.target), b"language: ko").unwrap();
    let p = h
        .plan(&[Host::Claude], Mode::Default, std::slice::from_ref(&s))
        .unwrap();
    assert_eq!(p.entries[0].action, "unchanged");
    let r = h.apply(&p, &p.digest).unwrap();
    let rp = h.recovery_plan(&r).unwrap();
    assert_eq!(rp.actions, vec!["preserved"]);
    h.recover(&rp, &rp.digest).unwrap();
    assert_eq!(fs::read(d.path().join(&s.target)).unwrap(), b"language: ko");
}

#[test]
fn excluded_token_variants_never_reach_backup_or_import() {
    for bytes in [
        b"token = 'synthetic-value'".as_slice(),
        br#"{"TOKEN":"synthetic-value"}"#,
        br#"{"\u0074oken":"synthetic-value"}"#,
    ] {
        let (d, h, mut item) = fixture();
        fs::write(d.path().join(&item.source), bytes).unwrap();
        item.recorded_sha256 = Some(hash(bytes));
        let p = preview(&h, &[item]);
        assert_eq!(p.entries[0].action, "excluded_secret");
        assert!(h.apply(&p, &p.digest).is_err());
        assert!(!d.path().join(".moai-go-backup").exists());
        assert!(!d.path().join(".moai-cli/imports").exists());
    }
}

#[test]
fn unconfirmed_publish_is_an_explicit_conflict_and_preserves_both_copies() {
    let (d, h, item) = fixture();
    let p = preview(&h, &[item]);
    let mut receipt = h.apply(&p, &p.digest).unwrap();
    receipt.published[0] = false;
    let manifest = d
        .path()
        .join(".moai-go-backup")
        .join(&p.digest)
        .join("manifest.json");
    fs::write(&manifest, serde_json::to_vec(&receipt).unwrap()).unwrap();
    for body in [b"{\"language\":\"ko\"}".as_slice(), b"late user edit"] {
        let target = d.path().join(".moai-cli/imports/language.json");
        fs::write(&target, body).unwrap();
        let recovery = h.recovery_plan(&receipt).unwrap();
        assert_eq!(recovery.actions, vec!["conflict_unconfirmed_publish"]);
        assert_eq!(
            h.recover(&recovery, &recovery.digest).unwrap(),
            vec!["conflict_unconfirmed_publish"]
        );
        assert_eq!(fs::read(target).unwrap(), body);
        assert!(d.path().join(".moai/config/language.json").is_file());
    }
}

#[test]
fn receipt_digest_tamper_is_refused_before_any_recovery_write() {
    let (d, h, item) = fixture();
    let p = preview(&h, &[item]);
    let mut receipt = h.apply(&p, &p.digest).unwrap();
    let manifest = d
        .path()
        .join(".moai-go-backup")
        .join(&p.digest)
        .join("manifest.json");
    receipt.plan.summary.push_str("tampered summary");
    fs::write(&manifest, serde_json::to_vec(&receipt).unwrap()).unwrap();
    let before = moai_cli::tree_digest(d.path(), &[]).unwrap();
    assert!(h.recovery_plan(&receipt).is_err());
    assert_eq!(before, moai_cli::tree_digest(d.path(), &[]).unwrap());
}

fn audit_fixture() -> (tempfile::TempDir, SyntheticHome, Selection) {
    let d = tempfile::tempdir().unwrap();
    fs::create_dir_all(d.path().join(".moai/config")).unwrap();
    fs::write(d.path().join(".moai/config/lang.json"), b"language: ko").unwrap();
    let h = SyntheticHome::new(d.path()).unwrap();
    let s = Selection {
        source: ".moai/config/lang.json".into(),
        target: ".moai-cli/imports/lang.json".into(),
        recorded_sha256: Some(hash(b"language: ko")),
        project_current: None,
    };
    (d, h, s)
}
#[test]
fn tampered_manifest_must_not_delete_a_preexisting_user_import() {
    let (d, h, s) = audit_fixture();
    let p = h.plan(&[Host::Claude], Mode::Default, &[s]).unwrap();
    let mut r = h.apply(&p, &p.digest).unwrap();
    let user = ".moai-cli/imports/user.txt";
    fs::write(d.path().join(user), b"language: ko").unwrap();
    r.plan.entries[0].selection.target = user.into();
    let dir = d.path().join(".moai-go-backup").join(&p.digest);
    fs::write(dir.join("manifest.json"), serde_json::to_vec(&r).unwrap()).unwrap();
    let rp = h.recovery_plan(&r);
    let rejected = match rp {
        Err(_) => true,
        Ok(rp) => h.recover(&rp, &rp.digest).is_err(),
    };
    assert!(
        rejected && d.path().join(user).exists(),
        "tampered digest accepted; user_import_exists={}",
        d.path().join(user).exists()
    );
}
#[test]
fn interrupted_publish_must_not_silently_preserve_import() {
    let (d, h, s) = audit_fixture();
    let p = h.plan(&[Host::Claude], Mode::Default, &[s]).unwrap();
    let mut r = h.apply(&p, &p.digest).unwrap();
    r.published[0] = false;
    fs::write(
        d.path()
            .join(".moai-go-backup")
            .join(&p.digest)
            .join("manifest.json"),
        serde_json::to_vec(&r).unwrap(),
    )
    .unwrap();
    let rp = h.recovery_plan(&r).unwrap();
    assert_ne!(
        rp.actions[0], "preserved",
        "published bytes exist, durable marker false; actions={:?}",
        rp.actions
    );
}
#[test]
fn generic_token_content_must_be_excluded() {
    let (d, h, mut s) = audit_fixture();
    let bytes = b"token: synthetic-audit-value";
    fs::write(d.path().join(&s.source), bytes).unwrap();
    s.recorded_sha256 = Some(hash(bytes));
    let p = h.plan(&[Host::Glm], Mode::Default, &[s]).unwrap();
    let applied = h.apply(&p, &p.digest);
    assert!(
        applied.is_err(),
        "generic token copied into import={}",
        d.path().join(".moai-cli/imports/lang.json").is_file()
    );
}
#[test]
fn fifo_and_traversal_are_rejected_before_read() {
    let (d, h, s) = audit_fixture();
    let fifo = d.path().join(".moai/config/pipe");
    assert!(std::process::Command::new("/usr/bin/mkfifo")
        .arg(&fifo)
        .status()
        .unwrap()
        .success());
    for source in [".moai/config/pipe", ".moai/config/../../../escape"] {
        let bad = Selection {
            source: source.into(),
            ..s.clone()
        };
        assert!(h.plan(&[Host::Claude], Mode::Default, &[bad]).is_err());
    }
}

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
        &[Host::Claude, Host::Codex, Host::Glm],
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
        .plan(&[Host::Codex], Mode::Automatic, std::slice::from_ref(&item))
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

fn release(hashes: &[String]) -> GoRelease {
    GoRelease {
        version: "v3.2.1".into(),
        sha256: hashes.to_vec(),
    }
}

/// #2016 decision 6784520717: `owned` needs the recorded hash AND the Go release list.
#[test]
fn release_list_ownership_grades_are_conservative() {
    let (d, h, item) = fixture();
    let src = hash(b"{\"language\":\"ko\"}");
    let other = hash(b"another release file");
    let grade = |rel: &GoRelease, s: &Selection| {
        let p = h
            .plan_release(&[Host::Claude], Mode::Default, rel, std::slice::from_ref(s))
            .unwrap();
        (p.entries[0].ownership.clone(), p.entries[0].action.clone())
    };
    // recorded + listed: the only importable grade.
    assert_eq!(
        grade(&release(std::slice::from_ref(&src)), &item),
        ("owned".into(), "create".into())
    );
    // recorded but not shipped by the release: preserved.
    assert_eq!(
        grade(&release(std::slice::from_ref(&other)), &item),
        ("unreleased".into(), "preserved".into())
    );
    // shipped by the release but no install record: preserved, reported only.
    let unrecorded = Selection {
        recorded_sha256: None,
        ..item.clone()
    };
    assert_eq!(
        grade(&release(std::slice::from_ref(&src)), &unrecorded),
        ("release_match_unrecorded".into(), "preserved".into())
    );
    assert_eq!(
        grade(&release(std::slice::from_ref(&other)), &unrecorded),
        ("unknown".into(), "preserved".into())
    );
    let modified = Selection {
        recorded_sha256: Some(other.clone()),
        ..item.clone()
    };
    assert_eq!(
        grade(&release(&[src.clone(), other.clone()]), &modified),
        ("modified".into(), "preserved".into())
    );
    // The list is bound into the digest: a plan whose list is swapped is refused.
    let before = moai_cli::tree_digest(d.path(), &[]).unwrap();
    let p = h
        .plan_release(
            &[Host::Claude],
            Mode::Default,
            &release(std::slice::from_ref(&src)),
            std::slice::from_ref(&item),
        )
        .unwrap();
    let mut swapped = p.clone();
    swapped.go_release = Some(release(&[src.clone(), other.clone()]));
    assert!(h.apply(&swapped, &p.digest).is_err());
    let unreleased = h
        .plan_release(
            &[Host::Claude],
            Mode::Default,
            &release(&[other]),
            std::slice::from_ref(&item),
        )
        .unwrap();
    assert!(h.apply(&unreleased, &unreleased.digest).is_err());
    assert_eq!(before, moai_cli::tree_digest(d.path(), &[]).unwrap());
    h.apply(&p, &p.digest).unwrap();
    assert!(d.path().join(&item.target).is_file());
}

#[test]
fn release_list_bounds_and_format_are_checked() {
    let (_d, h, item) = fixture();
    let ok = hash(b"x");
    let mut bad = vec![
        release(&[]),
        release(&[ok.clone(), ok.clone()]),
        release(&[ok.to_uppercase()]),
        release(&[ok[..63].to_string()]),
        GoRelease {
            version: String::new(),
            ..release(std::slice::from_ref(&ok))
        },
        GoRelease {
            version: "v3/../x".into(),
            ..release(std::slice::from_ref(&ok))
        },
        GoRelease {
            version: "v".repeat(65),
            ..release(std::slice::from_ref(&ok))
        },
    ];
    // 768 / 769 distinct hashes (RELEASE_HASH_LIMIT, written as literals).
    let many: Vec<String> = (0..769u32).map(|i| hash(&i.to_le_bytes())).collect();
    bad.push(release(&many));
    for rel in &bad {
        assert!(h
            .plan_release(
                &[Host::Claude],
                Mode::Default,
                rel,
                std::slice::from_ref(&item)
            )
            .is_err());
    }
    let edge = GoRelease {
        version: "v".repeat(64),
        ..release(&many[..768])
    };
    assert!(h
        .plan_release(
            &[Host::Claude],
            Mode::Default,
            &edge,
            std::slice::from_ref(&item)
        )
        .is_ok());
}

#[test]
fn onprem_host_and_duplicate_hosts_are_refused() {
    let (_d, h, item) = fixture();
    assert!(serde_json::from_str::<Host>("\"OnPrem\"").is_err());
    assert!(h
        .plan(
            &[Host::Claude, Host::Claude],
            Mode::Default,
            std::slice::from_ref(&item)
        )
        .is_err());
}

/// #2449 audit F2: an existing backup folder is a conflict (exclusive create), not an I/O
/// error worth retrying; nothing is imported.
#[test]
fn an_existing_backup_folder_is_a_conflict() {
    let (d, h, item) = fixture();
    let p = h
        .plan(&[Host::Claude], Mode::Default, std::slice::from_ref(&item))
        .unwrap();
    fs::create_dir_all(d.path().join(".moai-go-backup").join(&p.digest)).unwrap();
    let err = h.apply(&p, &p.digest).unwrap_err();
    assert!(err.contains("backup already exists"), "{err}");
    assert_eq!(
        moai_cli::contract::classify(&err),
        moai_cli::contract::Kind::Conflict
    );
    assert!(!d.path().join(&item.target).exists());
}
