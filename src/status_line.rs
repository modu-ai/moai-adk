//! The Claude Code `statusLine` MoAI writes into a project calls `moai statusline`; the one the
//! releases before #3004 P4-3 wrote calls `moai-cli statusline`. This module generates both
//! spellings, recognises the old one, previews its rewrite and holds the gate that decides whether
//! a plan may write the new one.
//!
//! Recognition is equality with the legacy generator's own output ([`command_of`] called with
//! [`LEGACY_BINARY`]), the same rule `hook_legacy` applies to the hook sets: another argument, a
//! path prefix, other whitespace or another object shape is not MoAI's old value. It is kept as it
//! is and its text is never printed (a preview names a reason code only). A value is MoAI's to
//! rewrite only when the ownership receipt records it as MoAI's and the file still holds it.
//!
//! Nothing here writes a settings file. The one writer that already owns the key is
//! `claude::settings_apply_in` (three-way against the ownership receipt); [`apply_in`] only runs
//! the gate in front of it and keeps the old file next to it. A pristine old install is an
//! `update`; an edited one is not rewritten at all.
//!
//! The gate (decision Q4 of the #3004 P4 plan) is the hook gate's PATH idea plus one more fact. The
//! new command is `moai statusline`, so the `moai` first on PATH must be the Rust build
//! (`hook_binary.state == resolved`, probed by `--version`) AND must answer `statusline --help` as
//! the Rust build does. The second probe exists because a routing script can answer `--version`
//! as Rust and still send `statusline` to the Go build; a version probe alone cannot tell. The
//! probe is the same safe runner as the version probe ([`crate::install::probe_args_env`]) and
//! runs only for a `moai` outside the project. A hold never reaches the hooks, whose gate is
//! `hook_legacy`'s and does not read this one.

use crate::claude::{self, SettingsScope};
use crate::hooks::{self, HOOK_BINARY, LEGACY_BINARY};
use serde_json::{json, Map, Value};
use std::path::Path;

/// The subcommand every generated status line command ends with.
pub const SUBCOMMAND: &str = "statusline";

/// The first stdout line of `moai statusline --help` on the Rust build (the command's summary in
/// `src/main.rs`); a test runs the built binary and compares. The Go build prints something else.
pub const RUST_HELP_LINE: &str = "Render the Claude Code status line";

/// The status line object a project gets for the binary named `binary`.
pub fn command_of(binary: &str) -> Value {
    json!({"type": "command", "command": format!("{binary} {SUBCOMMAND}")})
}

/// What the generator writes now.
pub fn current() -> Value {
    command_of(HOOK_BINARY)
}

/// What the generator of the releases before #3004 P4-3 wrote.
pub fn legacy() -> Value {
    command_of(LEGACY_BINARY)
}

/// What a plan would do with the status line.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Kind {
    /// MoAI's old value is replaced by the new one.
    Rewrite,
    /// A project with no status line and no ownership of one gets the new one.
    Create,
}

impl Kind {
    fn name(self) -> &'static str {
        match self {
            Kind::Rewrite => "rewrite",
            Kind::Create => "create",
        }
    }
}

/// What MoAI does with the status line of a project, from the receipt's value and the file's.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Intent {
    Write(Kind),
    /// Left as it is, for this reason code.
    Keep(&'static str),
}

/// The decision: `owned` is the receipt's `statusLine`, `file` the settings file's.
pub fn intent(owned: Option<&Value>, file: Option<&Value>) -> Intent {
    let (old, new) = (legacy(), current());
    match (owned, file) {
        (None, None) => Intent::Write(Kind::Create),
        (None, Some(c)) if *c == new => Intent::Keep("current"),
        (None, Some(_)) => Intent::Keep("foreign"),
        (Some(o), Some(c)) if *o == old && *c == old => Intent::Write(Kind::Rewrite),
        (Some(o), Some(_)) if *o == old => Intent::Keep("edited"),
        (Some(o), Some(c)) if *o == new && *c == new => Intent::Keep("current"),
        (Some(o), Some(_)) if *o == new => Intent::Keep("edited"),
        (Some(_), None) => Intent::Keep("removed"),
        (Some(_), Some(_)) => Intent::Keep("other"),
    }
}

/// The kind of write a request makes: `desired` is MoAI's new status line exactly and the files
/// say [`Intent::Write`]. `None` for every other request (another value is the caller's).
pub fn classify(
    owned: Option<&Value>,
    file: Option<&Value>,
    desired: Option<&Value>,
) -> Option<Kind> {
    if desired != Some(&current()) {
        return None;
    }
    match intent(owned, file) {
        Intent::Write(kind) => Some(kind),
        Intent::Keep(_) => None,
    }
}

/// Why a write is held: a short code for the preview and a sentence for the person.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Hold {
    pub code: String,
    pub message: String,
}

/// How `moai statusline` answers on the first `moai` on PATH.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Reach {
    /// `statusline --help` printed the Rust build's first line.
    Rust,
    /// It printed something else: a Go build, or a script that routes `statusline` away.
    RoutedElsewhere,
    /// It printed nothing within the probe limit.
    Unprobed,
}

/// Probes `moai statusline --help` on the `moai` the version probe resolved (`moai["path"]`,
/// outside the project). HOME is the one variable the child gets, as for the version probe.
pub fn reach(moai: &Value, home: Option<&Path>) -> Reach {
    let Some(path) = moai["path"].as_str() else {
        return Reach::Unprobed;
    };
    let extra: Vec<(&str, &Path)> = home.map(|h| ("HOME", h)).into_iter().collect();
    match crate::install::probe_args_env(Path::new(path), &[SUBCOMMAND, "--help"], &extra) {
        None => Reach::Unprobed,
        Some(line) if line == RUST_HELP_LINE => Reach::Rust,
        Some(_) => Reach::RoutedElsewhere,
    }
}

fn hold(kind: Kind, code: String, why: &str) -> Hold {
    let what = match kind {
        Kind::Rewrite => "status line not rewritten",
        Kind::Create => "status line not added",
    };
    Hold {
        message: format!(
            "{what}: it would call moai statusline, and {why}; put the Rust moai first on PATH \
             and plan again (nothing was changed)"
        ),
        code,
    }
}

/// The gate decision for a write of `kind` given the state of the `moai` first on PATH
/// (`moai`, as `hooks::moai_state` returns it) and what its `statusline` answers.
///
/// - A rewrite needs `resolved` and a Rust `statusline`: the old status line works today, and a
///   `moai` that is not the Rust build would swap it for a command that reaches something else.
/// - A new default is refused on `wrong_generation`, and on a `resolved` `moai` whose
///   `statusline` is not the Rust build. With no `moai` on PATH (`binary_unresolved`) or one
///   inside the project (`not_probed`) it is written, as a hook install is: nothing would run.
pub fn check(kind: Kind, moai: &Value, reach_of: impl FnOnce(&Value) -> Reach) -> Result<(), Hold> {
    let state = moai["state"].as_str().unwrap_or("wrong_generation");
    match (kind, state) {
        (_, "wrong_generation") => {
            return Err(hold(
                kind,
                "moai_wrong_generation".into(),
                "the moai first on PATH is not the Rust build (wrong_generation)",
            ))
        }
        (Kind::Create, "binary_unresolved" | "not_probed") => return Ok(()),
        (Kind::Rewrite, s) if s != "resolved" => {
            return Err(hold(
                kind,
                format!("moai_{s}"),
                &format!("the moai first on PATH is not the Rust build ({s})"),
            ))
        }
        _ => {}
    }
    match reach_of(moai) {
        Reach::Rust => Ok(()),
        Reach::RoutedElsewhere => Err(hold(
            kind,
            "statusline_routed_elsewhere".into(),
            "the moai first on PATH answers `statusline` as something other than the Rust build \
             (a routing script?)",
        )),
        Reach::Unprobed => Err(hold(
            kind,
            "statusline_unprobed".into(),
            "the moai first on PATH did not answer `statusline --help` within the probe limit",
        )),
    }
}

/// [`check`] for the `moai` on `path_env` as seen from `project`. Probes only when it must.
pub fn gate(kind: Kind, project: &Path, home: Option<&Path>, path_env: &str) -> Result<(), Hold> {
    let moai = hooks::moai_state(project, path_env, home);
    check(kind, &moai, |m| reach(m, home))
}

/// The `status_line_migration` object of a plan: what the status line would become and whether
/// the gate holds it. It carries reason codes and the two binary names, never a command text.
pub fn migration(intent: Intent, verdict: Option<&Result<(), Hold>>) -> Value {
    let mut out = json!({
        "from": LEGACY_BINARY,
        "to": HOOK_BINARY,
        "state": "none",
        "reason": null,
        "blocked_by": null,
    });
    match (intent, verdict) {
        (Intent::Keep(reason), _) => out["reason"] = json!(reason),
        (Intent::Write(kind), Some(Err(h))) => {
            out["state"] = json!("held");
            out["held_action"] = json!(kind.name());
            out["reason"] = json!(match kind {
                Kind::Rewrite => "legacy_default",
                Kind::Create => "no_status_line",
            });
            out["blocked_by"] = json!(h.code);
            out["note"] = json!(h.message);
        }
        (Intent::Write(kind), _) => {
            out["state"] = json!(kind.name());
            out["reason"] = json!(match kind {
                Kind::Rewrite => "legacy_default",
                Kind::Create => "no_status_line",
            });
        }
    }
    out
}

/// `claude::settings_plan_in` with the `status_line_migration` preview next to it (outside the
/// plan digest, which binds the files and not the PATH). The object is present only when the
/// plan concerns MoAI's status line: a request for the new one, or an old one MoAI owns.
pub fn plan_in(
    project: &Path,
    desired: &Map<String, Value>,
    scope: SettingsScope,
    home: Option<&Path>,
    path_env: &str,
) -> Result<Value, String> {
    let mut out = claude::settings_plan_in(project, desired, scope)?;
    let (owned, current) = claude::scope_view(project, scope)?;
    let (o, c) = (owned.get("statusLine"), current.get("statusLine"));
    // A request that keeps the status line as it is (another key changes, or the old value is
    // carried over) does not rewrite it, so the preview says `not_requested`, not `rewrite`.
    let said = match (intent(o, c), classify(o, c, desired.get("statusLine"))) {
        (Intent::Write(_), None) => Intent::Keep("not_requested"),
        (said, _) => said,
    };
    let verdict =
        classify(o, c, desired.get("statusLine")).map(|kind| gate(kind, project, home, path_env));
    if verdict.is_some() || o == Some(&legacy()) {
        out["status_line_migration"] = migration(said, verdict.as_ref());
    }
    Ok(out)
}

/// `claude::settings_apply_in` behind the gate: a request that writes MoAI's new status line is
/// refused, before anything is written, when [`gate`] holds it now (the plan digest does not bind
/// the PATH, so the plan-time verdict is not enough). After a rewrite the old settings file is
/// kept byte for byte under `.moai-cli/settings-backups/`, and the result says where.
#[allow(clippy::too_many_arguments)]
pub fn apply_in(
    project: &Path,
    desired: &Map<String, Value>,
    plan_digest: &str,
    permission: &str,
    scope: SettingsScope,
    confirm_replace: Option<&str>,
    home: Option<&Path>,
    path_env: &str,
) -> Result<Value, String> {
    let (owned, current) = claude::scope_view(project, scope)?;
    let kind = classify(
        owned.get("statusLine"),
        current.get("statusLine"),
        desired.get("statusLine"),
    );
    if let Some(kind) = kind {
        gate(kind, project, home, path_env).map_err(|h| h.message)?;
    }
    let before = match kind {
        Some(Kind::Rewrite) => claude::read_existing(&project.join(scope.file()))?,
        _ => None,
    };
    let mut out = claude::settings_apply_in(
        project,
        desired,
        plan_digest,
        permission,
        scope,
        confirm_replace,
    )?;
    let Some(kind) = kind else {
        return Ok(out);
    };
    let wrote = out["files_changed"]
        .as_array()
        .is_some_and(|f| !f.is_empty());
    let mut note = json!({
        "from": LEGACY_BINARY,
        "to": HOOK_BINARY,
        "state": match (kind, wrote) {
            (_, false) => "unchanged",
            (Kind::Rewrite, true) => "rewritten",
            (Kind::Create, true) => "created",
        },
    });
    if let (Kind::Rewrite, true) = (kind, wrote) {
        match claude::keep_original(project, scope, before.as_deref()) {
            Ok(dir) => {
                let name = scope.file().rsplit('/').next().unwrap_or("settings.json");
                note["backup_dir"] = json!(dir);
                note["revert"] = json!(format!(
                    "from the project folder: cp '{dir}/{name}' '{}' (MoAI then treats the restored value as your change and leaves it as it is)",
                    scope.file()
                ));
            }
            // The settings file is already written; say that the old copy could not be kept.
            Err(e) => note["backup_error"] = json!(e),
        }
    }
    out["status_line_migration"] = note;
    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn lit(command: &str) -> Value {
        json!({"type": "command", "command": command})
    }

    #[test]
    fn the_two_generators_differ_only_in_the_binary_name() {
        assert_eq!(current(), lit("moai statusline"));
        assert_eq!(legacy(), lit("moai-cli statusline"));
    }

    /// The old value is recognised by equality and by nothing looser.
    #[test]
    fn intent_follows_the_receipt_and_the_file() {
        let (old, new) = (legacy(), current());
        assert_eq!(intent(Some(&old), Some(&old)), Intent::Write(Kind::Rewrite));
        assert_eq!(intent(None, None), Intent::Write(Kind::Create));
        assert_eq!(intent(Some(&new), Some(&new)), Intent::Keep("current"));
        assert_eq!(intent(None, Some(&new)), Intent::Keep("current"));
        assert_eq!(intent(None, Some(&old)), Intent::Keep("foreign"));
        assert_eq!(intent(Some(&old), None), Intent::Keep("removed"));
        for edited in [
            lit("moai-cli statusline --foo"),
            lit("moai-cli  statusline"),
            lit("moai-cli statusline "),
            lit("/x/moai-cli statusline"),
            json!({"type": "command", "command": "moai-cli statusline", "padding": 1}),
            json!("moai-cli statusline"),
        ] {
            assert_eq!(
                intent(Some(&old), Some(&edited)),
                Intent::Keep("edited"),
                "{edited}"
            );
            assert_eq!(
                intent(None, Some(&edited)),
                Intent::Keep("foreign"),
                "{edited}"
            );
            // A receipt that names an edited text is not the old generator's output either.
            assert_eq!(
                intent(Some(&edited), Some(&edited)),
                Intent::Keep("other"),
                "{edited}"
            );
        }
    }

    #[test]
    fn only_a_request_for_the_new_value_is_classified() {
        let (old, new) = (legacy(), current());
        assert_eq!(
            classify(Some(&old), Some(&old), Some(&new)),
            Some(Kind::Rewrite)
        );
        assert_eq!(classify(None, None, Some(&new)), Some(Kind::Create));
        assert_eq!(classify(Some(&old), Some(&old), Some(&old)), None);
        assert_eq!(classify(Some(&old), Some(&old), Some(&lit("mine"))), None);
        assert_eq!(classify(Some(&old), Some(&old), None), None);
        assert_eq!(
            classify(Some(&old), Some(&lit("moai-cli statusline -x")), Some(&new)),
            None
        );
    }

    fn moai(state: &str) -> Value {
        json!({"state": state, "path": "/x/moai"})
    }

    /// The gate in both directions: what it lets through and what it holds.
    #[test]
    fn the_gate_lets_a_rust_statusline_through_and_holds_the_rest() {
        let code = |k, state: &str, r| check(k, &moai(state), |_| r).map_err(|h| h.code);
        for kind in [Kind::Rewrite, Kind::Create] {
            assert_eq!(code(kind, "resolved", Reach::Rust), Ok(()));
            assert_eq!(
                code(kind, "resolved", Reach::RoutedElsewhere),
                Err("statusline_routed_elsewhere".to_string())
            );
            assert_eq!(
                code(kind, "resolved", Reach::Unprobed),
                Err("statusline_unprobed".to_string())
            );
            assert_eq!(
                code(kind, "wrong_generation", Reach::Rust),
                Err("moai_wrong_generation".to_string())
            );
        }
        for state in ["binary_unresolved", "not_probed"] {
            assert_eq!(
                code(Kind::Rewrite, state, Reach::Rust),
                Err(format!("moai_{state}"))
            );
            assert_eq!(code(Kind::Create, state, Reach::RoutedElsewhere), Ok(()));
        }
    }

    #[test]
    fn the_statusline_is_not_probed_when_the_version_probe_already_held() {
        for (kind, state) in [
            (Kind::Rewrite, "wrong_generation"),
            (Kind::Rewrite, "binary_unresolved"),
            (Kind::Create, "wrong_generation"),
            (Kind::Create, "binary_unresolved"),
        ] {
            let _ = check(kind, &moai(state), |_| panic!("{kind:?} {state} probed"));
        }
    }

    #[test]
    fn a_preview_never_carries_a_command_text() {
        let held = Err(Hold {
            code: "statusline_unprobed".into(),
            message: "m".into(),
        });
        for i in [
            Intent::Keep("edited"),
            Intent::Write(Kind::Rewrite),
            Intent::Write(Kind::Create),
        ] {
            let text = migration(i, Some(&held)).to_string();
            assert!(!text.contains("statusline\""), "{text}");
        }
        let m = migration(Intent::Write(Kind::Rewrite), Some(&held));
        assert_eq!(
            (m["state"].clone(), m["held_action"].clone()),
            (json!("held"), json!("rewrite"))
        );
        let ok = migration(Intent::Write(Kind::Rewrite), Some(&Ok(())));
        assert_eq!(ok["state"], json!("rewrite"));
        assert_eq!(ok["blocked_by"], Value::Null);
    }
}
