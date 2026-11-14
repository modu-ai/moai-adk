//! The home layer of `moai update` (#2871 I2 U2b, PLAN comment 6879946980): the Go generation's
//! `~/.moai` items are previewed beside the user layer ([`crate::home_retire::plan`]) and moved
//! to the dated backup only by `moai update --confirm-home <token>`.
//!
//! Like the cleanup (#2872 stage 2) the part has its own token: neither `--confirm <digest>`
//! nor the terminal yes applies it, and it is not part of the layer's `plan_digest`. The token
//! is the preview's clock-free `content_digest`, so a preview and the confirming run may happen
//! at different times; the apply recomputes the preview and refuses a token that is not its
//! own before anything is written.
//!
//! The part is shown only for a `~/.moai` the verdict table calls `go`, `mixed` or `rust-first`
//! (a HOME without Go items gets exactly the preview it had before), with the entries of the
//! preview and, when something would move, the confirming flag. A blocked preview (a live Go
//! session, `MOAI_HOME` elsewhere, ...) is shown with its reasons and no token.

use crate::home_retire;
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The token prefix, so a token of another part is never taken for this one.
const PREFIX: &str = "home:";

/// The home layer of one `moai update` preview.
#[derive(Debug)]
pub struct Layer {
    /// The part's preview JSON; `None` when the HOME has no Go items to show.
    preview: Option<Value>,
    /// The confirming token; `None` unless confirming would move something.
    token: Option<String>,
}

/// The token of a preview's `content_digest` (`sha256:<hex>`): sixteen hex digits.
fn token_of(content_digest: &str) -> String {
    let hex = content_digest
        .strip_prefix("sha256:")
        .unwrap_or(content_digest);
    format!("{PREFIX}{}", hex.chars().take(16).collect::<String>())
}

/// The retire preview of `home` as the update's home part. Writes nothing.
pub fn plan(home: &Path) -> Result<Layer> {
    let moai_home = std::env::var("MOAI_HOME").ok();
    let p = home_retire::plan(home, moai_home.as_deref(), &crate::audit::utc_now())?;
    Ok(layer_of(home, &p))
}

fn layer_of(home: &Path, p: &Value) -> Layer {
    let generation = p["state"]["generation"].as_str().unwrap_or("");
    let active = matches!(generation, "go" | "mixed" | "rust-first");
    let action = p["action"].as_str().unwrap_or("none");
    // Shown when something would move, when the part is blocked, and when Go items exist but
    // every one is in use (so the person sees why nothing is offered).
    let in_use = p["counts"]["skip-in-use"].as_u64().unwrap_or(0) > 0;
    let shown = active && (matches!(action, "retire" | "blocked") || in_use);
    if !shown {
        return Layer {
            preview: None,
            token: None,
        };
    }
    let token = (action == "retire")
        .then(|| p["content_digest"].as_str().map(token_of))
        .flatten();
    let preview = json!({
        "action": action,
        "generation": generation,
        "backup_root": home.join(home_retire::RETIRED_DIR).display().to_string(),
        "counts": p["counts"],
        "total": p["total"],
        "blocks": p["blocks"],
        "warnings": p["warnings"],
        "entries": p["entries"].as_array().into_iter().flatten().map(|e| json!({
            "name": e["name"], "kind": e["kind"], "action": e["action"], "reason": e["reason"],
        })).collect::<Vec<_>>(),
        "confirm_token": token,
        "apply_flag": token.as_ref().map(|t| format!("--confirm-home {t}")),
        "original_kept": true,
        "writes": false,
    });
    Layer {
        preview: Some(preview),
        token,
    }
}

impl Layer {
    /// The part's preview, when there is one.
    pub fn to_json(&self) -> Option<&Value> {
        self.preview.as_ref()
    }

    /// The token `--confirm-home` takes, when confirming would move something.
    pub fn token(&self) -> Option<&str> {
        self.token.as_deref()
    }

    /// The terminal question's lines for this part; empty when there is no part.
    pub fn summary(&self) -> String {
        let Some(p) = &self.preview else {
            return String::new();
        };
        let n = |k: &str| p["counts"][k].as_u64().unwrap_or(0);
        let mut out = String::from("Go items in ~/.moai\n");
        match p["action"].as_str() {
            Some("retire") => out.push_str(&format!(
                "  retire {} · keep {} · skipped in use {} · confirm with: moai update {}\n",
                n("retire"),
                n("keep"),
                n("skip-in-use"),
                p["apply_flag"].as_str().unwrap_or("")
            )),
            Some("blocked") => out.push_str(&format!(
                "  blocked: {}\n",
                p["blocks"]
                    .as_array()
                    .into_iter()
                    .flatten()
                    .filter_map(Value::as_str)
                    .collect::<Vec<_>>()
                    .join(", ")
            )),
            _ => out.push_str(&format!(
                "  nothing moves now: keep {} · in use {}\n",
                n("keep"),
                n("skip-in-use")
            )),
        }
        out
    }
}

/// Refuse a `--confirm-home` token that is not the preview's, before anything is written.
pub fn check_token<'a>(layer: &Layer, given: Option<&'a str>) -> Result<Option<&'a str>> {
    let Some(t) = given else {
        return Ok(None);
    };
    if layer.token() != Some(t) {
        return Err(
            "the home token differs from the preview's; nothing applied, preview again".into(),
        );
    }
    Ok(Some(t))
}

/// Retire the previewed Go items (the confirmed `token`): the preview is computed again and
/// must give the same token, then [`home_retire::apply_confirmed`] moves the entries. The
/// result is the retire result plus the revert command.
pub fn apply(home: &Path, token: &str, permission_reference: &str) -> Result<Value> {
    let moai_home = std::env::var("MOAI_HOME").ok();
    let now = crate::audit::utc_now();
    let fresh = home_retire::plan(home, moai_home.as_deref(), &now)?;
    let content = fresh["content_digest"].as_str().unwrap_or_default();
    if fresh["action"] != "retire" || token_of(content) != token {
        return Err(
            "the home token differs from the preview's; nothing applied, preview again".into(),
        );
    }
    // The result already names the revert command, with the `--home` it ran against.
    home_retire::apply_confirmed(
        home,
        moai_home.as_deref(),
        &now,
        content,
        permission_reference,
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_token_is_sixteen_hex_digits_of_the_content_digest() {
        let t = token_of("sha256:0123456789abcdef0123456789abcdef");
        assert_eq!(t, "home:0123456789abcdef");
        assert_eq!(token_of("0123456789abcdef9"), "home:0123456789abcdef");
    }

    #[test]
    fn only_a_go_home_with_something_to_say_has_a_part() {
        let home = Path::new("/h");
        let p = |generation: &str, action: &str| {
            layer_of(
                home,
                &json!({
                    "state": {"generation": generation}, "action": action,
                    "content_digest": "sha256:aaaaaaaaaaaaaaaaaaaa", "counts": {}, "total": 0,
                    "blocks": [], "warnings": [], "entries": [],
                }),
            )
        };
        let shown = |l: &Layer| l.to_json().is_some();
        assert!(shown(&p("go", "retire")) && p("go", "retire").token().is_some());
        assert!(shown(&p("mixed", "retire")) && shown(&p("rust-first", "retire")));
        // Blocked: shown, no token.
        let b = p("go", "blocked");
        assert!(shown(&b) && b.token().is_none());
        // Go items that are all in use: shown (so the person sees why), no token.
        let u = layer_of(
            home,
            &json!({
                "state": {"generation": "go"}, "action": "none",
                "content_digest": "sha256:aaaaaaaaaaaaaaaaaaaa",
                "counts": {"retire": 0, "keep": 1, "skip-in-use": 2}, "total": 3,
                "blocks": [], "warnings": [], "entries": [],
            }),
        );
        assert!(shown(&u) && u.token().is_none());
        assert!(u.summary().contains("in use 2"), "{}", u.summary());
        // Nothing to move, or a HOME that is not Go: no part.
        for (g, a) in [
            ("go", "none"),
            ("none", "none"),
            ("rust", "none"),
            ("refuse", "blocked"),
            ("interrupted", "blocked"),
            ("unknown", "none"),
        ] {
            assert!(!shown(&p(g, a)), "{g} {a}");
        }
    }

    #[test]
    fn plan_shows_a_go_home_with_its_token() {
        let dir = tempfile::tempdir().unwrap();
        let home = std::fs::canonicalize(dir.path()).unwrap();
        let moai = home.join(".moai");
        std::fs::create_dir_all(moai.join("state")).unwrap();
        std::fs::create_dir_all(moai.join("cache")).unwrap();
        std::fs::write(moai.join("user-assets.json"), "{}").unwrap();
        std::fs::write(moai.join("state/active-sessions.json"), "[]").unwrap();
        let l = plan(&home).unwrap();
        let p = l.to_json().expect("a Go HOME has a home part");
        assert_eq!(p["action"], json!("retire"), "{p}");
        let t = l.token().expect("a token");
        assert!(t.starts_with("home:") && t.len() == 5 + 16, "{t}");
        assert_eq!(p["apply_flag"], json!(format!("--confirm-home {t}")));
    }

    /// `apply` recomputes the preview and takes only its own token: a wrong or stale one writes
    /// nothing; the right one moves the items and names the revert.
    #[test]
    fn apply_takes_only_the_previews_own_token() {
        let dir = tempfile::tempdir().unwrap();
        let home = std::fs::canonicalize(dir.path()).unwrap();
        let moai = home.join(".moai");
        std::fs::create_dir_all(moai.join("state")).unwrap();
        std::fs::create_dir_all(moai.join("cache")).unwrap();
        std::fs::write(moai.join("user-assets.json"), "{}").unwrap();
        std::fs::write(moai.join("state/active-sessions.json"), "[]").unwrap();
        let token = plan(&home).unwrap().token().unwrap().to_string();
        for bad in ["home:0000000000000000", "cleanup:aaaaaaaaaaaaaaaa", ""] {
            let e = apply(&home, bad, "unit test").unwrap_err();
            assert!(e.contains("home token"), "{e}");
            assert!(moai.join("cache").is_dir() && !home.join(".moai-go-retired").exists());
        }
        // A new retire candidate appeared after the preview: the token is stale.
        std::fs::create_dir_all(moai.join("logs")).unwrap();
        assert!(apply(&home, &token, "unit test").is_err());
        assert!(moai.join("cache").is_dir());
        // The fresh token moves the Go items and names the revert. An entry that stays
        // appearing in between (the layer's own writes are such) does not stale it.
        let token = plan(&home).unwrap().token().unwrap().to_string();
        std::fs::write(moai.join("appeared-later"), "x").unwrap();
        let r = apply(&home, &token, "unit test").unwrap();
        assert_eq!(r["state"], json!("applied"), "{r}");
        assert!(!moai.join("cache").exists() && moai.join("appeared-later").is_file());
        let revert = r["revert"].as_str().unwrap();
        assert!(revert.contains(&format!("--stamp {}", r["stamp"].as_str().unwrap())));
        assert!(
            revert.contains(&format!("--home {}", home.display())),
            "{revert}"
        );
    }

    #[test]
    fn a_missing_or_foreign_token_is_refused() {
        let l = Layer {
            preview: None,
            token: Some("home:aaaaaaaaaaaaaaaa".into()),
        };
        assert_eq!(check_token(&l, None), Ok(None));
        assert_eq!(
            check_token(&l, Some("home:aaaaaaaaaaaaaaaa")),
            Ok(Some("home:aaaaaaaaaaaaaaaa"))
        );
        for bad in ["home:bbbbbbbbbbbbbbbb", "cleanup:aaaaaaaaaaaaaaaa", ""] {
            assert!(check_token(&l, Some(bad)).is_err(), "{bad}");
        }
        let none = Layer {
            preview: None,
            token: None,
        };
        assert!(check_token(&none, Some("home:aaaaaaaaaaaaaaaa")).is_err());
    }
}
