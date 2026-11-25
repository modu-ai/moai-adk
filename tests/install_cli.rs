//! Installation receipt and install-inspect fixtures (04-A #1998). New-file discipline:
//! existing test files are untouched.
//!
//! Go updater characterization port (AC-04A-03) cites the sibling read-only reference
//! checkout `moai-adk-go` `internal/update/checker.go` — L34-89 (list selection: prefix
//! `go-v`/`v` filter, first item, no draft/prerelease fields parsed), L324-347
//! (compareSemver: prefix strip, prerelease suffix stripped per segment, non-numeric
//! segment counts as 0) — and `internal/cli/deps.go:430-450` (channel URL choice:
//! `rc`/`alpha`/`beta`/`dev`/`go-v` in the current version select the `/releases` list,
//! everything else selects `/releases/latest`). Fixture provenance:
//! `tests/fixtures/go-releases-sample.json` was recorded from
//! `https://api.github.com/repos/modu-ai/moai-adk/releases?per_page=15` (authenticated
//! `gh` call, 2026-11-08T10:57:53Z), filtered to the fields the checker reads plus
//! `draft`/`prerelease` (present in the real response, unmodeled by the checker).

#![cfg(unix)]

mod common;

use common::CleanEnv;
use moai_cli::install::{
    plan_receipt, record_receipt, validate_receipt, InstallReceipt, InstallerOwner, ReceiptPlan,
};
use moai_cli::tree_digest;
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

// Both fixtures are exec'd by install-inspect, so they are written by a child process (#2283).
fn cli_copy(dest: &Path) {
    common::copy_executable(Path::new(EXE), dest);
}

fn script(dest: &Path, body: &str) {
    common::write_executable(dest, body);
}

fn run_inspect(home: &Path, path_env: &Path, cwd: &Path) -> std::process::Output {
    Command::new(EXE)
        .args(["install-inspect", "--json"])
        .env("HOME", home)
        .env("PATH", path_env)
        .current_dir(cwd)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .output()
        .expect("spawn install-inspect")
}

fn ok_result(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).expect("envelope JSON on stdout");
    assert_eq!(
        v["ok"],
        json!(true),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    v["result"].clone()
}

fn receipt_fixture() -> InstallReceipt {
    InstallReceipt {
        schema_version: 1,
        product: "moai-cli".into(),
        generation: "rust-rc".into(),
        channel: "rc".into(),
        version: "2026.11.0-rc.1".into(),
        binary_path: "~/.moai-cli/bin/moai-cli".into(),
        binary_sha256: format!("sha256:{}", "a".repeat(64)),
        installer_owner: InstallerOwner::Unknown,
        protocol_version: 1,
        created_at: "2026-11-08T10:00:00Z".into(),
        updated_at: "2026-11-08T10:00:00Z".into(),
    }
}

/// AC-04A-01: the REQ-04A-01 schema round-trips, refuses unknown keys and refuses
/// credential-shaped or malformed field values.
#[test]
fn install_receipt_schema_roundtrip_rejects_unknown_and_secrets() {
    let receipt = receipt_fixture();

    // Round trip keeps every REQ field with its declared name.
    let bytes = serde_json::to_string(&receipt).unwrap();
    for key in [
        "schema_version",
        "product",
        "generation",
        "channel",
        "version",
        "binary_path",
        "binary_sha256",
        "installer_owner",
        "protocol_version",
        "created_at",
        "updated_at",
    ] {
        assert!(bytes.contains(&format!("\"{key}\"")), "{key} missing");
    }
    let back: InstallReceipt = serde_json::from_str(&bytes).unwrap();
    assert_eq!(back, receipt);

    // Unknown keys are refused on read (deny_unknown_fields).
    let mut value = serde_json::to_value(&receipt).unwrap();
    value["extra"] = json!("x");
    assert!(serde_json::from_value::<InstallReceipt>(value).is_err());

    // Every installer_owner variant round-trips under its lowercase name.
    for (owner, word) in [
        (InstallerOwner::Direct, "direct"),
        (InstallerOwner::Cargo, "cargo"),
        (InstallerOwner::Homebrew, "homebrew"),
        (InstallerOwner::Msi, "msi"),
        (InstallerOwner::Unknown, "unknown"),
    ] {
        let json = serde_json::to_string(&owner).unwrap();
        assert_eq!(json, format!("\"{word}\""));
        assert_eq!(
            serde_json::from_str::<InstallerOwner>(&json).unwrap(),
            owner
        );
    }
    assert!(serde_json::from_str::<InstallerOwner>("\"brew\"").is_err());

    // Validation negatives: credential shapes, malformed paths/digests/timestamps.
    let invalid = |mutate: &dyn Fn(&mut InstallReceipt)| {
        let mut r = receipt_fixture();
        mutate(&mut r);
        assert!(
            validate_receipt(&r).is_err(),
            "expected rejection: {}",
            serde_json::to_string(&r).unwrap()
        );
    };
    invalid(&|r| r.version = "sk-abc123".into());
    // C-90 audit N2: a receipt counts `sk-` anywhere, after a letter or an encoded `=` too.
    invalid(&|r| r.version = "key%3Dsk-abc123".into());
    invalid(&|r| r.channel = "xsk-abc".into());
    invalid(&|r| r.binary_path = "~/bin/key%3dsk-abc/moai-cli".into());
    invalid(&|r| r.channel = "token=abc".into());
    invalid(&|r| r.binary_path = "~/x/ghp_secret/binary".into());
    invalid(&|r| r.binary_path = "bin/moai-cli".into()); // relative
    invalid(&|r| r.binary_path = "~~/bin/moai-cli".into());
    invalid(&|r| r.binary_path = format!("~/.moai-cli/bin/{}", "x".repeat(2100)));
    invalid(&|r| r.binary_sha256 = "deadbeef".into());
    invalid(&|r| r.binary_sha256 = format!("sha256:{}", "g".repeat(64)));
    invalid(&|r| r.created_at = "2026-11-08 10:00:00".into());
    invalid(&|r| r.updated_at = "not-a-time".into());
    invalid(&|r| r.schema_version = 2);
    invalid(&|r| r.product = "moai".into());
    invalid(&|r| r.generation = "go".into());
    invalid(&|r| r.protocol_version = 2);
    invalid(&|r| r.channel = String::new());
    invalid(&|r| r.version = String::new());

    // Validation positives for the binary path shapes the schema allows.
    for shape in [
        "~/.moai-cli/bin/moai-cli",
        "/usr/local/bin/moai-cli",
        "C:\\Users\\goos\\.moai-cli\\bin\\moai-cli.exe",
    ] {
        let mut r = receipt_fixture();
        r.binary_path = shape.into();
        assert!(validate_receipt(&r).is_ok(), "{shape}");
    }
}

/// AC-04A-05: recording the same input twice is `created` then `unchanged`, byte-identical;
/// a different receipt for the same binary path is refused, never silently rewritten.
/// The Go generation's fake `.moai/` tree in the same home is untouched throughout
/// (REQ-04A-04).
#[test]
fn install_receipt_rerun_is_unchanged() {
    let home = TempDir::new().unwrap();
    let home = home.path();
    let exe = home.join(".moai-cli/bin/moai-cli");
    fs::create_dir_all(exe.parent().unwrap()).unwrap();
    cli_copy(&exe);

    let go_tree = home.join(".moai");
    fs::create_dir_all(go_tree.join("cache")).unwrap();
    fs::write(go_tree.join("config.yaml"), "generation: go\n").unwrap();
    fs::write(go_tree.join("cache/x.bin"), b"go state bytes").unwrap();
    let go_digest_before = tree_digest(&go_tree, &[]).unwrap();

    let plan1 = plan_receipt(&exe, home, InstallerOwner::Unknown).unwrap();
    assert_eq!(plan1.receipt.channel, "rc"); // version contains -rc
    assert_eq!(plan1.receipt.binary_path, "~/.moai-cli/bin/moai-cli");
    assert_eq!(plan1.receipt.created_at, plan1.receipt.updated_at);

    let created = record_receipt(home, &plan1).unwrap();
    assert_eq!(created["state"], json!("created"));
    let path = home
        .join(".moai-cli/installations")
        .join(format!("{}.json", plan1.install_id));
    let bytes_first = fs::read(&path).unwrap();
    let stored: InstallReceipt = serde_json::from_slice(&bytes_first).unwrap();
    assert_eq!(stored, plan1.receipt);

    // Rerun with fresh timestamps: identity fields match, so nothing is written.
    let plan2 = plan_receipt(&exe, home, InstallerOwner::Unknown).unwrap();
    assert_eq!(plan2.install_id, plan1.install_id);
    let again = record_receipt(home, &plan2).unwrap();
    assert_eq!(again["state"], json!("unchanged"));
    assert_eq!(fs::read(&path).unwrap(), bytes_first);

    // Different fields under the same id are refused (REQ-04A-06: no auto-update).
    let mut conflicting = plan1.receipt.clone();
    conflicting.binary_sha256 = format!("sha256:{}", "0".repeat(64));
    let plan3 = ReceiptPlan {
        install_id: plan1.install_id.clone(),
        receipt: conflicting,
    };
    let err = record_receipt(home, &plan3).unwrap_err();
    assert!(err.contains("conflict"), "{err}");
    assert_eq!(fs::read(&path).unwrap(), bytes_first);

    assert_eq!(tree_digest(&go_tree, &[]).unwrap(), go_digest_before);
}

/// AC-04A-02: `install-inspect --json` distinguishes the fake PATH `moai` (Go imitation)
/// from `moai-cli`, reports the Go-generation path existence, and leaves the temporary
/// HOME+project tree byte-identical (tree digest before/after).
#[test]
fn install_inspect_is_read_only_and_distinguishes_generations() {
    let root = TempDir::new().unwrap();
    let root = root.path();
    let home = root.join("home");
    let fakebin = home.join("fakebin");
    let proj = root.join("proj");
    fs::create_dir_all(&fakebin).unwrap();
    fs::create_dir_all(proj.join(".moai")).unwrap();
    // Go generation imitation in the temporary home.
    fs::create_dir_all(home.join(".moai/cache")).unwrap();
    fs::write(home.join(".moai/config.yaml"), "generation: go\n").unwrap();
    fs::write(home.join(".moai/cache/blob"), b"original bytes").unwrap();
    script(&fakebin.join("moai"), "#!/bin/sh\necho 'moai 3.1.5 (go)'\n");
    let cli_path = fakebin.join("moai-cli");
    cli_copy(&cli_path);

    let before = tree_digest(root, &[]).unwrap();
    let out = run_inspect(&home, &fakebin, &proj);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let result = ok_result(&out);

    let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(envelope["protocol_version"], json!(1));
    assert_eq!(envelope["operation"], json!("install.inspect"));

    let moai = &result["path_resolution"]["moai"];
    let moai_cli = &result["path_resolution"]["moai_cli"];
    assert_eq!(moai["found"], json!(true));
    assert!(moai["path"].as_str().unwrap().ends_with("fakebin/moai"));
    assert_eq!(moai["version_output"], json!(true));
    assert_eq!(moai_cli["found"], json!(true));
    assert!(moai_cli["path"]
        .as_str()
        .unwrap()
        .ends_with("fakebin/moai-cli"));
    assert_eq!(moai_cli["version_output"], json!(true));
    assert_ne!(moai["sha256"], moai_cli["sha256"]);
    assert_ne!(moai["path"], moai_cli["path"]);

    assert_eq!(result["go_generation"]["home_moai_exists"], json!(true));
    assert_eq!(result["go_generation"]["project_moai_exists"], json!(true));
    let canonical_proj = fs::canonicalize(&proj).unwrap();
    assert_eq!(
        result["go_generation"]["project_root"],
        json!(canonical_proj.to_str().unwrap())
    );
    // No receipts recorded yet: nothing on PATH is treated as self (REQ-04A-03).
    assert_eq!(result["self"]["state"], json!("unrecorded"));
    assert_eq!(result["installations"], json!([]));
    assert_eq!(result["writes"], json!(false));

    let after = tree_digest(root, &[]).unwrap();
    assert_eq!(after, before, "install-inspect must be read-only");
}

/// REQ-04A-03 and REQ-04A-06: self-identification requires the receipt's path AND digest
/// to match the running executable; the PATH-first `moai` is never mistaken for self; a
/// digest mismatch is reported `stale_receipt` and the receipt is never rewritten. The
/// subject is the CLI binary itself (its `current_exe`), recorded under a temporary home.
#[test]
fn install_receipt_self_match_and_stale() {
    let home_dir = TempDir::new().unwrap();
    let home = home_dir.path();
    let exe = PathBuf::from(EXE);
    let plan = plan_receipt(&exe, home, InstallerOwner::Unknown).unwrap();
    record_receipt(home, &plan).unwrap();

    let fakebin = home.join("fakebin");
    fs::create_dir_all(&fakebin).unwrap();
    script(&fakebin.join("moai"), "#!/bin/sh\necho 'moai 3.1.5 (go)'\n");
    let proj = home.join("proj");
    fs::create_dir_all(&proj).unwrap();

    // The running CLI matches its own receipt; the PATH `moai` is reported but is
    // never treated as self.
    let out = run_inspect(home, &fakebin, &proj);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let result = ok_result(&out);
    assert_eq!(result["self"]["state"], json!("self"));
    assert_eq!(
        result["self"]["current_exe"],
        json!(plan.receipt.binary_path)
    );
    // The running binary's own name (`moai` since #2792 R4): cargo links `moai` and
    // `moai-cli` separately, so the two are not byte-identical and `moai` keeps its own path.
    assert!(result["self"]["current_exe"]
        .as_str()
        .unwrap()
        .ends_with(exe.file_name().unwrap().to_str().unwrap()));
    assert_eq!(result["installations"].as_array().unwrap().len(), 1);
    assert_eq!(result["installations"][0]["match_state"], json!("self"));
    assert_eq!(result["path_resolution"]["moai"]["found"], json!(true));
    assert!(result["path_resolution"]["moai"]["path"]
        .as_str()
        .unwrap()
        .ends_with("fakebin/moai"));
    assert_ne!(
        result["path_resolution"]["moai"]["path"],
        result["self"]["current_exe"]
    );

    // An empty PATH resolves nothing, yet self still comes from the receipt.
    let out = run_inspect(home, Path::new(""), &proj);
    let result = ok_result(&out);
    assert_eq!(result["path_resolution"]["moai"]["found"], json!(false));
    assert_eq!(result["self"]["state"], json!("self"));

    // Tamper the recorded digest (simulates a replaced binary): stale_receipt, no rewrite.
    let receipt_path = moai_cli::home_root::path(home)
        .join("installations")
        .join(format!("{}.json", plan.install_id));
    let mut value: Value = serde_json::from_slice(&fs::read(&receipt_path).unwrap()).unwrap();
    value["binary_sha256"] = json!(format!("sha256:{}", "0".repeat(64)));
    let mut tampered = serde_json::to_vec_pretty(&value).unwrap();
    tampered.push(b'\n');
    fs::write(&receipt_path, &tampered).unwrap();
    let out = run_inspect(home, Path::new(""), &proj);
    let result = ok_result(&out);
    assert_eq!(result["self"]["state"], json!("stale_receipt"));
    assert_eq!(
        result["installations"][0]["match_state"],
        json!("stale_receipt")
    );
    assert_eq!(fs::read(&receipt_path).unwrap(), tampered, "no auto-update");
}

// --- Go 3.x updater characterization (REQ-04A-05, AC-04A-03) -------------------------
// The functions below are a faithful port of the Go rules, cited line by line. They are
// test-side evidence only; none of this logic ships in src/.

/// checker.go L86-89: keep `go-v`/`v` prefixed tags; L96: the FIRST kept item is "latest".
/// draft/prerelease fields are not parsed at all (releaseResponse, checker.go L30-35).
fn go_rc_channel_selects(releases: &[Value]) -> Option<String> {
    releases
        .iter()
        .map(|r| r["tag_name"].as_str().unwrap_or_default().to_string())
        .find(|tag| tag.starts_with("go-v") || tag.starts_with('v'))
}

/// GitHub `/releases/latest` semantics (docs 2022-11-28, verified in the Issue body):
/// the most recent non-draft non-prerelease release; modeled as the first such entry in
/// the recorded list order. The checker then reads the single object with NO prefix
/// filter (checker.go L121-127).
fn go_stable_channel_selects(releases: &[Value]) -> Option<String> {
    releases
        .iter()
        .filter(|r| r["draft"] == json!(false) && r["prerelease"] == json!(false))
        .map(|r| r["tag_name"].as_str().unwrap_or_default().to_string())
        .next()
}

/// deps.go:430-450: the current Go version's shape picks the feed URL.
fn go_channel_of(current: &str) -> &'static str {
    let dev = current == "dev"
        || current.contains("rc")
        || current.contains("alpha")
        || current.contains("beta")
        || current.starts_with("go-v");
    if dev {
        "rc"
    } else {
        "stable"
    }
}

/// checker.go L324-347: strip `go-v` then `v`, split into at most 3 dot segments, cut
/// each segment at the first `-`/`+` (prerelease/build suffix dropped), parse as integer,
/// a non-numeric segment counts as 0.
fn go_semver_parts(v: &str) -> [i64; 3] {
    let v = v.strip_prefix("go-v").unwrap_or(v);
    let v = v.strip_prefix('v').unwrap_or(v);
    let mut parts = [0i64; 3];
    for (i, seg) in v.splitn(3, '.').enumerate() {
        let seg = match seg.find(['-', '+']) {
            Some(idx) => &seg[..idx],
            None => seg,
        };
        if let Ok(n) = seg.parse::<i64>() {
            parts[i] = n;
        }
    }
    parts
}

fn go_compare_semver(a: &str, b: &str) -> std::cmp::Ordering {
    go_semver_parts(a).cmp(&go_semver_parts(b))
}

/// checker.go L325-334 (`IsUpdateAvailable`): update only when latest is strictly greater.
fn go_update_available(latest: Option<&str>, current: &str) -> Option<bool> {
    latest.map(|tag| go_compare_semver(tag, current) == std::cmp::Ordering::Greater)
}

fn feed_with_candidate(sample: &[Value], tag: &str, prerelease: bool) -> Vec<Value> {
    let mut feed = sample.to_vec();
    feed.insert(
        0,
        json!({
            "tag_name": tag,
            "name": tag,
            "draft": false,
            "prerelease": prerelease,
            "published_at": "2026-11-08T00:00:00Z",
            "assets": [],
        }),
    );
    feed
}

/// AC-04A-03: the contamination table. With the recorded feed plus one hypothetical
/// Rust release at the top, reports per candidate namespace what a Go stable-channel
/// client (`/releases/latest`) and a Go RC-channel client (`/releases` list, first item)
/// select and whether that reads as an available update. The table is decision
/// evidence, not the decision.
#[test]
fn go_updater_feed_contamination_table() {
    let sample: Vec<Value> =
        serde_json::from_str(include_str!("fixtures/go-releases-sample.json")).unwrap();
    let stable_current = "v3.2.1"; // pkg/version.go default build version shape
    let rc_current = "go-v3.2.1-rc.2";

    // The recorded feed itself: the FIRST entry is a draft (v3.1.3, draft=true) and the
    // rule does not check drafts, so an authenticated RC-channel client selects it.
    // Anonymous callers do not see drafts, so their first entry is v3.1.2.
    assert_eq!(go_rc_channel_selects(&sample).as_deref(), Some("v3.1.3"));
    assert_eq!(sample[0]["draft"], json!(true));

    let rows: Vec<(String, bool)> = vec![
        ("moai-adk-v2026.11.0-rc.1".into(), true), // decided namespace (B)
        ("v2026.11.0-rc.1".into(), true),
        ("v2026.11.0".into(), false),
        ("rust-v2026.11.0-rc.1".into(), true),
    ];

    println!("Go 3.x updater contamination table (fixture recorded 2026-11-08T10:57:53Z)");
    println!("rc current = {rc_current}, stable current = {stable_current}");
    println!(
        "{:<32} {:<10} {:<22} {:<22} {:<14} {:<14}",
        "candidate tag",
        "prerelease",
        "stable (/latest)",
        "rc (/releases first)",
        "stable upd?",
        "rc upd?"
    );
    for (tag, prerelease) in &rows {
        let feed = feed_with_candidate(&sample, tag, *prerelease);
        let stable_sel = go_stable_channel_selects(&feed);
        let rc_sel = go_rc_channel_selects(&feed);
        let stable_upd = go_update_available(stable_sel.as_deref(), stable_current);
        let rc_upd = go_update_available(rc_sel.as_deref(), rc_current);
        println!(
            "{:<32} {:<10} {:<22} {:<22} {:<14} {:<14}",
            tag,
            prerelease,
            stable_sel.as_deref().unwrap_or("-"),
            rc_sel.as_deref().unwrap_or("-"),
            format!("{:?}", stable_upd),
            format!("{:?}", rc_upd),
        );
    }
    println!(
        "separate repository: the Go checker's feed URL is compiled to\n\
         api.github.com/repos/modu-ai/moai-adk/releases (deps.go:36, :443-450); a release\n\
         in any other repository never enters either channel's feed.\n\
         note: a selected candidate with no checksums.txt asset makes buildVersionInfo\n\
         return ErrChecksumUnavailable (checker.go L187-217), so the Go updater errors\n\
         instead of downloading."
    );

    // Decided namespace B: invisible to both channels.
    let feed = feed_with_candidate(&sample, "moai-adk-v2026.11.0-rc.1", true);
    assert_ne!(
        go_rc_channel_selects(&feed).as_deref(),
        Some("moai-adk-v2026.11.0-rc.1")
    );
    assert_ne!(
        go_stable_channel_selects(&feed).as_deref(),
        Some("moai-adk-v2026.11.0-rc.1")
    );
    assert_eq!(
        go_update_available(go_rc_channel_selects(&feed).as_deref(), rc_current),
        Some(false)
    );

    // Plain `v` prefix, prerelease: selected by the RC channel's list, reads as an update.
    let feed = feed_with_candidate(&sample, "v2026.11.0-rc.1", true);
    assert_eq!(
        go_rc_channel_selects(&feed).as_deref(),
        Some("v2026.11.0-rc.1")
    );
    assert_eq!(
        go_update_available(Some("v2026.11.0-rc.1"), rc_current),
        Some(true)
    );
    assert_ne!(
        go_stable_channel_selects(&feed).as_deref(),
        Some("v2026.11.0-rc.1")
    );

    // Plain `v` prefix, non-prerelease: selected by BOTH channels, reads as an update.
    let feed = feed_with_candidate(&sample, "v2026.11.0", false);
    assert_eq!(go_rc_channel_selects(&feed).as_deref(), Some("v2026.11.0"));
    assert_eq!(
        go_stable_channel_selects(&feed).as_deref(),
        Some("v2026.11.0")
    );
    assert_eq!(
        go_update_available(Some("v2026.11.0"), rc_current),
        Some(true)
    );
    assert_eq!(
        go_update_available(Some("v2026.11.0"), stable_current),
        Some(true)
    );

    // Alternative prefix: invisible to both channels, like the decided namespace.
    let feed = feed_with_candidate(&sample, "rust-v2026.11.0-rc.1", true);
    assert_ne!(
        go_rc_channel_selects(&feed).as_deref(),
        Some("rust-v2026.11.0-rc.1")
    );
    assert_ne!(
        go_stable_channel_selects(&feed).as_deref(),
        Some("rust-v2026.11.0-rc.1")
    );

    // The prerelease suffix does not protect a `v`-tagged candidate in the version
    // comparison: parseSemverParts strips it (checker.go L356-370).
    assert_eq!(
        go_compare_semver("v2026.11.0-rc.1", "v2026.11.0"),
        std::cmp::Ordering::Equal
    );

    // Sanity of the channel rule itself (deps.go:430-450).
    assert_eq!(go_channel_of("v3.2.1"), "stable");
    assert_eq!(go_channel_of("go-v3.2.1-rc.2"), "rc");
    assert_eq!(go_channel_of("dev"), "rc");
}

/// #2696 C4b: the HOME state root follows `home_root` — the earlier `~/.moai-cli` until
/// `~/.moai` holds this generation's marker, the Go generation's `~/.moai` never.
#[test]
fn install_inspect_names_the_home_state_root_by_its_marker() {
    let root = TempDir::new().unwrap();
    let home = root.path().join("home");
    let fakebin = home.join("fakebin");
    let proj = root.path().join("proj");
    fs::create_dir_all(&fakebin).unwrap();
    fs::create_dir_all(&proj).unwrap();
    let state_root = |home: &Path| -> String {
        let out = run_inspect(home, &fakebin, &proj);
        ok_result(&out)["home_state_root"]
            .as_str()
            .unwrap()
            .to_string()
    };
    // Go generation `~/.moai` (no marker): the earlier root.
    fs::create_dir_all(home.join(".moai/state")).unwrap();
    fs::write(home.join(".moai/state/active-sessions.json"), "{}").unwrap();
    assert!(state_root(&home).ends_with("/.moai-cli"), "go HOME");
    // This generation's marker alone: `~/.moai`.
    let fresh = root.path().join("fresh");
    fs::create_dir_all(fresh.join(".moai")).unwrap();
    fs::write(
        fresh.join(".moai/moai-cli.json"),
        r#"{"schema_version":1,"generation":"moai-cli"}"#,
    )
    .unwrap();
    let marked = state_root(&fresh);
    assert!(marked.ends_with("/.moai"), "{marked}");
}

/// #2867 (R5a): `path_first_moai` names the first `moai` on PATH with its generation and the
/// evidence — a Go shim is `go`, a Rust build is `rust`, an executable that gives no version
/// line is `unknown` (each `go`/`unknown` carries the note); no `moai` on PATH is `absent`, and
/// a repository-content candidate is reported but never run.
#[test]
fn the_first_moai_on_path_is_reported_with_its_generation() {
    let root = TempDir::new().unwrap();
    let home = root.path().join("home");
    let proj = root.path().join("proj");
    fs::create_dir_all(&home).unwrap();
    fs::create_dir_all(&proj).unwrap();
    let first = |bin: &Path| -> Value {
        let out = run_inspect(&home, bin, &proj);
        ok_result(&out)["path_first_moai"].clone()
    };
    let go = root.path().join("go-bin");
    fs::create_dir_all(&go).unwrap();
    script(&go.join("moai"), "#!/bin/sh\necho 'moai 3.1.5 (go)'\n");
    let v = first(&go);
    assert_eq!(v["generation"], json!("go"), "{v}");
    assert_eq!(v["basis"], json!("--version: moai 3.1.5 (go)"), "{v}");
    let note = v["note"].as_str().unwrap();
    assert!(note.contains("is not the Rust build"), "{v}");
    assert!(note.contains("generated hooks call moai"), "{v}");
    assert!(
        !note.contains("until R5") && !note.contains("moai-cli"),
        "{v}"
    );
    assert!(v["path"].as_str().unwrap().ends_with("go-bin/moai"), "{v}");

    let rust = root.path().join("rust-bin");
    fs::create_dir_all(&rust).unwrap();
    cli_copy(&rust.join("moai"));
    let v = first(&rust);
    assert_eq!(v["generation"], json!("rust"), "{v}");
    assert_eq!(v["note"], Value::Null, "{v}");

    let silent = root.path().join("silent-bin");
    fs::create_dir_all(&silent).unwrap();
    script(&silent.join("moai"), "#!/bin/sh\nexit 3\n");
    let v = first(&silent);
    assert_eq!(v["generation"], json!("unknown"), "{v}");
    assert_eq!(v["basis"], json!("--version gave no line"), "{v}");
    assert!(v["note"].is_string(), "{v}");

    let empty = root.path().join("empty-bin");
    fs::create_dir_all(&empty).unwrap();
    let v = first(&empty);
    assert_eq!(
        v,
        json!({"path": null, "generation": "absent", "probed": false,
            "basis": "no moai on PATH", "note": null})
    );

    // A repository-content candidate (`$CLAUDE_PROJECT_DIR/.moai/bin/moai`) is never run: a
    // marker file proves the script did not execute.
    let ran = root.path().join("ran");
    fs::create_dir_all(proj.join(".moai/bin")).unwrap();
    script(
        &proj.join(".moai/bin/moai"),
        &format!("#!/bin/sh\ntouch '{}'\necho 'moai 3.1.5'\n", ran.display()),
    );
    let out = Command::new(EXE)
        .args(["install-inspect", "--json"])
        .env("HOME", &home)
        .env("PATH", &empty)
        .env("CLAUDE_PROJECT_DIR", &proj)
        .current_dir(&proj)
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let r = ok_result(&out);
    assert_eq!(r["path_first_moai"]["generation"], json!("absent"), "{r}");
    assert!(!ran.exists(), "the repository candidate was run");
    let item = r["moai_candidates"]["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"].as_str().unwrap().ends_with(".moai/bin/moai"))
        .unwrap_or_else(|| panic!("{r}"))
        .clone();
    assert_eq!(item["probed"], json!(false), "{item}");
}

// --- The install scripts read the release tag family of the release plan -----------------
// SPEC-20261121-moai-name, slice 6: the scripts pick `moai-adk-v<version>` releases and no
// longer the earlier family. Every run below uses a temporary HOME and a release list served
// from file://; the real HOME, the network and the installed binary are never touched, and a
// run stops at the download because no archive exists.

/// The earlier tag family, joined at run time so no line of this file spells it.
fn earlier_tag_prefix() -> String {
    ["moai", "cli", "v"].join("-")
}

/// The archive target the shell script detects on this machine; `None` on a platform the
/// installer does not support (the script stops there before it reads a release).
fn installer_target() -> Option<&'static str> {
    let say = |flag: &str| {
        let out = Command::new("uname").arg(flag).output().unwrap();
        String::from_utf8(out.stdout).unwrap().trim().to_string()
    };
    match (say("-s").as_str(), say("-m").as_str()) {
        ("Darwin", "arm64") => Some("aarch64-apple-darwin"),
        ("Linux", "x86_64") => Some("x86_64-unknown-linux-gnu"),
        _ => None,
    }
}

struct ReleaseFixture {
    _tmp: TempDir,
    root: PathBuf,
    home: PathBuf,
    scratch: PathBuf,
}

impl ReleaseFixture {
    fn url(&self, name: &str) -> String {
        format!("file://{}", self.root.join(name).display())
    }
}

/// A release list in the two shapes the scripts read: `releases.json` (the REST list, newest
/// first, left out when `rest_readable` is false so the script falls back to the feed) and
/// `releases.atom`. `base/` stays empty: no archive exists, so every run ends at the download.
fn release_fixture(releases: &[(String, bool)], rest_readable: bool) -> ReleaseFixture {
    let tmp = TempDir::new().unwrap();
    let root = tmp.path().join("fixture");
    let home = tmp.path().join("home");
    let scratch = tmp.path().join("tmp");
    for dir in [&root, &home, &scratch, &root.join("base")] {
        fs::create_dir_all(dir).unwrap();
    }
    if rest_readable {
        let list: Vec<Value> = releases
            .iter()
            .map(|(tag, pre)| json!({"tag_name": tag, "draft": false, "prerelease": pre}))
            .collect();
        fs::write(root.join("releases.json"), Value::Array(list).to_string()).unwrap();
    }
    let entries: String = releases
        .iter()
        .map(|(tag, _)| {
            format!(
                "<entry><link rel=\"alternate\" type=\"text/html\" \
                 href=\"https://github.com/modu-ai/moai-adk/releases/tag/{tag}\"/></entry>"
            )
        })
        .collect();
    fs::write(
        root.join("releases.atom"),
        format!("<?xml version=\"1.0\"?><feed>{entries}</feed>"),
    )
    .unwrap();
    ReleaseFixture {
        _tmp: tmp,
        root,
        home,
        scratch,
    }
}

fn install_env(cmd: &mut Command, fx: &ReleaseFixture, path: &str) {
    cmd.clean_env()
        .env("PATH", path)
        .env("HOME", &fx.home)
        .env("TMPDIR", &fx.scratch)
        .env("NO_COLOR", "1")
        .env("POWERSHELL_TELEMETRY_OPTOUT", "1")
        .env("MOAI_LANGUAGE", "en")
        .env("MOAI_RELEASE_API", fx.url("releases.json"))
        .env("MOAI_RELEASE_FEED", fx.url("releases.atom"))
        .env("MOAI_RELEASE_BASE", fx.url("base"))
        .current_dir(&fx.root)
        .stdin(Stdio::null());
}

fn script_path(name: &str) -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("install")
        .join(name)
}

/// The output of both streams in one text: the scripts print errors on stderr (shell) or on
/// the host stream (PowerShell).
fn text_of(out: &std::process::Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn run_install_sh(fx: &ReleaseFixture, args: &[&str]) -> std::process::Output {
    let mut cmd = Command::new("sh");
    cmd.arg(script_path("install.sh")).args(args);
    install_env(&mut cmd, fx, "/usr/bin:/bin:/usr/sbin:/sbin");
    cmd.output().unwrap()
}

/// A run that stops before the archive exists writes nothing: the temporary HOME is empty and
/// the script's own temporary folder is gone.
fn assert_nothing_written(fx: &ReleaseFixture, text: &str) {
    for dir in [&fx.home, &fx.scratch] {
        let left: Vec<_> = fs::read_dir(dir).unwrap().flatten().collect();
        assert!(left.is_empty(), "{dir:?} holds {left:?}\n{text}");
    }
}

fn tag(prefix: &str, version: &str) -> String {
    format!("{prefix}{version}")
}

#[test]
fn install_sh_finds_no_release_when_only_the_earlier_tag_family_exists() {
    if installer_target().is_none() {
        eprintln!("unsupported platform: install.sh stops before it reads a release");
        return;
    }
    let earlier = earlier_tag_prefix();
    for rest_readable in [true, false] {
        let fx = release_fixture(&[(tag(&earlier, "4.2649.0"), false)], rest_readable);
        let out = run_install_sh(&fx, &[]);
        let text = text_of(&out);
        assert!(!out.status.success(), "{text}");
        assert!(
            text.contains("No moai-adk-v* release found for channel stable"),
            "rest_readable={rest_readable}\n{text}"
        );
        assert!(!text.contains("Latest version"), "{text}");
        assert_nothing_written(&fx, &text);
    }
}

#[test]
fn install_sh_picks_the_moai_adk_tag_family_and_reads_the_version_argument() {
    if installer_target().is_none() {
        eprintln!("unsupported platform: install.sh stops before it reads a release");
        return;
    }
    let earlier = earlier_tag_prefix();
    let releases = [
        (tag(&earlier, "4.2650.0"), false),
        (tag("moai-adk-v", "4.2649.1"), true),
        (tag("moai-adk-v", "4.2649.0"), false),
    ];
    // The REST list marks a pre-release, so the stable channel skips 4.2649.1 and the earlier
    // family; the rc channel takes it.
    let fx = release_fixture(&releases, true);
    let text = text_of(&run_install_sh(&fx, &[]));
    assert!(text.contains("Latest version (stable): 4.2649.0"), "{text}");
    let text = text_of(&run_install_sh(&fx, &["--channel", "rc"]));
    assert!(text.contains("Latest version (rc): 4.2649.1"), "{text}");
    // The feed has no pre-release mark: the newest moai-adk-v tag is taken.
    let fx = release_fixture(&releases, false);
    let text = text_of(&run_install_sh(&fx, &[]));
    assert!(text.contains("Latest version (stable): 4.2649.1"), "{text}");
    assert!(
        !text.contains("4.2650.0"),
        "the earlier family was read: {text}"
    );
    // `--version` strips a leading `v` or the tag prefix, and refuses the earlier prefix.
    for given in ["4.2649.0", "v4.2649.0", "moai-adk-v4.2649.0"] {
        let text = text_of(&run_install_sh(&fx, &["--version", given]));
        assert!(
            text.contains("Installing version: 4.2649.0"),
            "{given}: {text}"
        );
    }
    let given = tag(&earlier, "4.2649.0");
    let out = run_install_sh(&fx, &["--version", &given]);
    let text = text_of(&out);
    assert!(!out.status.success(), "{text}");
    assert!(text.contains("Not a version"), "{text}");
    assert!(!text.contains("Installing version"), "{text}");
    assert_nothing_written(&fx, &text);
}

#[test]
fn install_sh_names_the_moai_adk_tag_in_the_download_address_and_writes_nothing_without_the_archive(
) {
    let Some(target) = installer_target() else {
        eprintln!("unsupported platform: install.sh stops before it reads a release");
        return;
    };
    let fx = release_fixture(&[(tag("moai-adk-v", "4.2649.0"), false)], true);
    let out = run_install_sh(&fx, &["--version", "4.2649.0"]);
    let text = text_of(&out);
    let base = fx.root.join("base");
    let address = format!(
        "Downloading from: file://{}/moai-adk-v4.2649.0/moai-cli-{target}.tar.xz",
        base.display()
    );
    assert!(text.contains(&address), "{text}");
    assert!(!out.status.success(), "{text}");
    assert!(text.contains("Download failed"), "{text}");
    assert_nothing_written(&fx, &text);
}

#[test]
fn install_ps1_reads_the_same_tag_family_when_pwsh_exists() {
    let Some(target) = installer_target() else {
        eprintln!("unsupported platform: install.ps1 stops before it reads a release");
        return;
    };
    // PowerShell writes its own cache and telemetry files under HOME, so even the probe runs
    // in a temporary one.
    let probe_home = TempDir::new().unwrap();
    let probe = Command::new("pwsh")
        .args(["-NoProfile", "-NonInteractive", "-Command", "exit 0"])
        .env("PATH", "/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin")
        .env("HOME", probe_home.path())
        .env("POWERSHELL_TELEMETRY_OPTOUT", "1")
        .stdin(Stdio::null())
        .output();
    if probe.map(|o| !o.status.success()).unwrap_or(true) {
        eprintln!("pwsh not found: install.ps1 not run");
        return;
    }
    let run = |fx: &ReleaseFixture, args: &[&str]| {
        let mut cmd = Command::new("pwsh");
        cmd.args(["-NoProfile", "-NonInteractive", "-File"])
            .arg(script_path("install.ps1"))
            .args(args);
        install_env(
            &mut cmd,
            fx,
            "/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin",
        );
        cmd.output().unwrap()
    };
    let earlier = earlier_tag_prefix();
    let fx = release_fixture(&[(tag(&earlier, "4.2649.0"), false)], true);
    let out = run(&fx, &[]);
    let text = text_of(&out);
    assert!(!out.status.success(), "{text}");
    assert!(
        text.contains("No moai-adk-v* release found for channel stable"),
        "{text}"
    );
    let fx = release_fixture(
        &[
            (tag(&earlier, "4.2650.0"), false),
            (tag("moai-adk-v", "4.2649.0"), false),
        ],
        true,
    );
    let out = run(&fx, &["--version", "moai-adk-v4.2649.0"]);
    let text = text_of(&out);
    let address = format!(
        "Downloading from: file://{}/moai-adk-v4.2649.0/moai-cli-{target}.tar.xz",
        fx.root.join("base").display()
    );
    assert!(text.contains(&address), "{text}");
    assert!(text.contains("Download failed"), "{text}");
    let out = run(&fx, &[]);
    assert!(
        text_of(&out).contains("Latest version (stable): 4.2649.0"),
        "{}",
        text_of(&out)
    );
}
