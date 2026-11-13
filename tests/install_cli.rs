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
    let receipt_path = home
        .join(".moai-cli/installations")
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
        ("moai-cli-v2026.11.0-rc.1".into(), true), // decided namespace (B)
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
    let feed = feed_with_candidate(&sample, "moai-cli-v2026.11.0-rc.1", true);
    assert_ne!(
        go_rc_channel_selects(&feed).as_deref(),
        Some("moai-cli-v2026.11.0-rc.1")
    );
    assert_ne!(
        go_stable_channel_selects(&feed).as_deref(),
        Some("moai-cli-v2026.11.0-rc.1")
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
