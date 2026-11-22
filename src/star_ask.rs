//! The star ask and the consented reports (SPEC-20261122-star-and-reports):
//! one polite line asks for a star when the state is unknown or not starred —
//! checked by exactly one read-only GET through the user's gh login, never a
//! write — and a dismissal is remembered in the home moai state (the
//! current-generation folder only; a home whose resolver answers the retired
//! folder writes nothing and says why). A report leaves the machine only on
//! the user's yes to its exact full text, created with gh under the user's
//! account, with the prefilled issue page URL when gh is absent. The command
//! runs only when something calls it explicitly — nothing calls it from a
//! hook, the status line or the install and update flows yet.

use crate::forge::transport::HttpRequest;
use std::path::{Path, PathBuf};

/// The repository the star line names.
pub const REPOSITORY: &str = "modu-ai/moai-adk";

/// The full URL of the one read-only GET (pinned by its own test, the
/// audit's N1: a mutant changing the URL must not survive).
pub const STAR_CHECK_URL: &str = "https://api.github.com/user/starred/modu-ai/moai-adk";

/// The star states the check answers.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum StarState {
    /// The GET answered 204: the user starred the repository.
    Starred,
    /// The GET answered 404: not starred.
    NotStarred,
    /// gh is absent, not logged in, or the check failed: unknown, and the
    /// line still shows (REQ-2's failure answer).
    Unknown,
}

/// The dismissal record in the home moai state: one file whose presence
/// means "never ask again" (the user starred, or asked not to be asked).
fn dismissal_file(home: &Path) -> PathBuf {
    crate::home_root::path(home).join("star-ask-dismissed")
}

/// The write outcome of a state write: done, or refused with its reason.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum WriteOutcome {
    /// The marker landed in the current-generation home folder.
    Written,
    /// The resolver answers the retired folder (`.moai-cli`): nothing new is
    /// ever created there and nothing was written (B2).
    RetiredFolder,
    /// The folder exists but the write failed (a read-only folder): nothing
    /// was written and the run that asked is never blocked (B3).
    WriteRefused(String),
}

/// Remember the dismissal (REQ-3): one empty marker file in the home moai
/// state — the current-generation folder only. A new HOME adopts `~/.moai`
/// first (the marker), so the state file's name cannot flip the folder
/// classification; a home whose resolver answers the retired folder writes
/// nothing and says why.
pub fn dismiss_star_ask(home: &Path) -> WriteOutcome {
    write_state_marker(home, "star-ask-dismissed")
}

/// Whether the star line is suppressed (a dismissal is recorded).
pub fn star_ask_dismissed(home: &Path) -> bool {
    dismissal_file(home).is_file()
}

/// The shared state-write path (B2, B3): write one empty marker file into
/// the current-generation home folder, creating the folder and its marker
/// for a new home; refuse to create anything under the retired folder; say
/// the truth when the write fails.
fn write_state_marker(home: &Path, name: &str) -> WriteOutcome {
    let _ = crate::home_root::adopt_new(home, name);
    let path = crate::home_root::path(home).join(name);
    // B2: when the resolver answers the retired folder, nothing new is ever
    // created there — write nothing and say why.
    if crate::home_root::dir(home) != crate::home_root::MOAI_DIR {
        return WriteOutcome::RetiredFolder;
    }
    if let Some(parent) = path.parent() {
        if let Err(e) = std::fs::create_dir_all(parent) {
            return WriteOutcome::WriteRefused(format!("the state folder cannot be created: {e}"));
        }
    }
    match std::fs::write(&path, b"") {
        Ok(()) => WriteOutcome::Written,
        Err(e) => WriteOutcome::WriteRefused(format!("the marker cannot be written: {e}")),
    }
}

/// The one read-only star check (REQ-2): exactly one GET of the user's
/// starred endpoint through the forge transport (the test seam swaps gh);
/// anything but 204 and 404 answers unknown, and no write call exists on
/// this path.
pub fn check_star(transport: &dyn crate::forge::transport::Transport) -> StarState {
    let request = HttpRequest::new("GET", STAR_CHECK_URL);
    match transport.send(&request) {
        Ok(response) => match response.status {
            204 => StarState::Starred,
            404 => StarState::NotStarred,
            _ => StarState::Unknown,
        },
        Err(_) => StarState::Unknown,
    }
}

/// The star ask line in both languages (B4: the text lives in a two-language
/// pair, not a single English constant; it joins the shared UI text table
/// when `src/ui_text.rs` is free — a pair request stands).
pub fn star_ask_line(lang: crate::ui_text::Lang) -> &'static str {
    match lang {
        crate::ui_text::Lang::En => STAR_ASK_LINE_EN,
        crate::ui_text::Lang::Ko => STAR_ASK_LINE_KO,
    }
}

/// The English star ask line.
pub const STAR_ASK_LINE_EN: &str = "Stars are real encouragement for this project — https://github.com/modu-ai/moai-adk if it helps you";
/// The Korean star ask line.
pub const STAR_ASK_LINE_KO: &str =
    "별은 이 프로젝트에 실제 힘이 돼요 — 도움이 되셨다면 https://github.com/modu-ai/moai-adk";

/// The star line's decision (REQ-1): shown when the state is unknown or not
/// starred and no dismissal stands, at most one line per run. The text comes
/// in the caller's language (B4).
pub fn star_line(
    state: StarState,
    home: &Path,
    lang: crate::ui_text::Lang,
) -> Option<&'static str> {
    if star_ask_dismissed(home) {
        return None;
    }
    match state {
        StarState::Starred => None,
        StarState::NotStarred | StarState::Unknown => Some(star_ask_line(lang)),
    }
}

/// The report consent choice of REQ-4, stored in the home moai state.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ReportConsent {
    /// Ask each time (the recommended default the install question offers).
    AskEachTime,
    /// Never send anything.
    Never,
}

impl ReportConsent {
    fn file(home: &Path) -> PathBuf {
        crate::home_root::path(home).join("report-consent")
    }

    /// Store the choice; changeable by the same command later. The same
    /// B2/B3 rules as the dismissal marker apply.
    pub fn store(self, home: &Path) -> WriteOutcome {
        let _ = crate::home_root::adopt_new(home, "report-consent");
        let path = Self::file(home);
        if crate::home_root::dir(home) != crate::home_root::MOAI_DIR {
            return WriteOutcome::RetiredFolder;
        }
        if let Some(parent) = path.parent() {
            if let Err(e) = std::fs::create_dir_all(parent) {
                return WriteOutcome::WriteRefused(format!(
                    "the state folder cannot be created: {e}"
                ));
            }
        }
        let bytes = match self {
            ReportConsent::AskEachTime => b"ask-each-time".as_slice(),
            ReportConsent::Never => b"never".as_slice(),
        };
        match std::fs::write(&path, bytes) {
            Ok(()) => WriteOutcome::Written,
            Err(e) => {
                WriteOutcome::WriteRefused(format!("the consent file cannot be written: {e}"))
            }
        }
    }

    /// The stored choice; a missing file reads as ask-each-time (nothing was
    /// answered yet, and nothing sends without a fresh yes anyway).
    pub fn load(home: &Path) -> ReportConsent {
        match std::fs::read_to_string(Self::file(home)) {
            Ok(text) if text.trim() == "never" => ReportConsent::Never,
            _ => ReportConsent::AskEachTime,
        }
    }
}

/// The preflight of a report body (REQ-5): the exact text the user will see,
/// refused when it carries code, a file path beyond the version and the OS,
/// a project name, or anything the secret rule blocks. The check runs before
/// any yes is possible. (N4 of the audit: this is the star slice's shape of
/// the check; the report slice tightens it to what the report actually
/// sends.)
pub fn report_body_check(text: &str) -> Result<(), &'static str> {
    if text.contains("```") || text.contains("    fn ") {
        return Err("report refused: the body carries code");
    }
    if text.contains('/') && text.contains('.') && text.contains(':') {
        return Err("report refused: the body carries a file path");
    }
    if crate::redact::redact_outgoing(text).1 > 0 {
        return Err("report refused: the body carries a value with a secret's shape");
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    /// A fake transport answering with a fixed status and recording the URL
    /// (N1: a mutant changing the URL must not survive).
    struct FixedTransport(u16, bool, std::cell::RefCell<Vec<String>>);
    impl crate::forge::transport::Transport for FixedTransport {
        fn send(
            &self,
            req: &HttpRequest,
        ) -> std::result::Result<crate::forge::transport::HttpResponse, String> {
            self.2.borrow_mut().push(req.url.clone());
            if self.1 {
                Err("no gh".into())
            } else {
                Ok(crate::forge::transport::HttpResponse::new(self.0))
            }
        }
    }

    fn answering(status: u16) -> FixedTransport {
        FixedTransport(status, false, std::cell::RefCell::new(vec![]))
    }

    fn failing() -> FixedTransport {
        FixedTransport(0, true, std::cell::RefCell::new(vec![]))
    }

    fn home() -> tempfile::TempDir {
        tempfile::tempdir().unwrap()
    }

    /// REQ-2 and N1: the check reads exactly one GET, answers by status, and
    /// the URL is pinned.
    #[test]
    fn the_star_check_answers_by_status_and_pins_the_url() {
        let transport = answering(204);
        assert_eq!(check_star(&transport), StarState::Starred);
        assert_eq!(
            transport.2.borrow().as_slice(),
            [STAR_CHECK_URL],
            "exactly one GET to the pinned URL"
        );

        assert_eq!(check_star(&answering(404)), StarState::NotStarred);
        assert_eq!(check_star(&answering(401)), StarState::Unknown);
        assert_eq!(check_star(&failing()), StarState::Unknown);
    }

    /// REQ-1, REQ-3, B1: the line shows for unknown and not-starred, never
    /// for starred, and a dismissal suppresses it forever.
    #[test]
    fn the_star_line_follows_the_state_and_the_dismissal() {
        let home = home();
        assert!(
            star_line(StarState::NotStarred, home.path(), crate::ui_text::Lang::En).is_some(),
            "not starred shows the line"
        );
        assert!(
            star_line(StarState::Unknown, home.path(), crate::ui_text::Lang::En).is_some(),
            "unknown shows the line"
        );
        assert!(
            star_line(StarState::Starred, home.path(), crate::ui_text::Lang::En).is_none(),
            "starred never shows the line"
        );
        assert_eq!(
            dismiss_star_ask(home.path()),
            WriteOutcome::Written,
            "the dismissal lands on a new home"
        );
        assert!(star_line(StarState::NotStarred, home.path(), crate::ui_text::Lang::En).is_none());
        assert!(star_line(StarState::Unknown, home.path(), crate::ui_text::Lang::En).is_none());
    }

    /// B2: on a home whose resolver answers the retired folder, the marker
    /// is not written and the outcome says why.
    #[test]
    fn a_retired_folder_home_writes_nothing() {
        let home = home();
        std::fs::create_dir_all(home.path().join(crate::HOME_STATE_DIR)).unwrap();
        assert_eq!(
            dismiss_star_ask(home.path()),
            WriteOutcome::RetiredFolder,
            "nothing is created under the retired folder"
        );
        assert!(!star_ask_dismissed(home.path()));
    }

    /// B3: a read-only home reports the write refusal with its reason.
    #[test]
    fn a_read_only_home_reports_the_refusal() {
        let home = home();
        // Create the folder and its marker, then remove the marker and make
        // the folder read-only: the second write must be refused.
        dismiss_star_ask(home.path()); // creates the folder and the marker
        let state = crate::home_root::path(home.path());
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt as _;
            // Remove the marker while the folder is still writable.
            std::fs::remove_file(state.join("star-ask-dismissed")).unwrap();
            let mut perms = std::fs::metadata(&state).unwrap().permissions();
            perms.set_mode(0o555);
            std::fs::set_permissions(&state, perms).unwrap();
        }
        match dismiss_star_ask(home.path()) {
            WriteOutcome::WriteRefused(reason) => {
                assert!(reason.contains("cannot be written"), "{reason}");
            }
            other => panic!("expected a refusal, got {other:?}"),
        }
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt as _;
            let mut perms = std::fs::metadata(&state).unwrap().permissions();
            perms.set_mode(0o755);
            std::fs::set_permissions(&state, perms).unwrap();
        }
    }

    /// REQ-4: the consent choice stores and loads, and a missing file reads
    /// as ask-each-time.
    #[test]
    fn the_report_consent_stores_and_loads() {
        let home = home();
        assert_eq!(ReportConsent::load(home.path()), ReportConsent::AskEachTime);
        assert_eq!(
            ReportConsent::Never.store(home.path()),
            WriteOutcome::Written
        );
        assert_eq!(ReportConsent::load(home.path()), ReportConsent::Never);
        assert_eq!(
            ReportConsent::AskEachTime.store(home.path()),
            WriteOutcome::Written
        );
        assert_eq!(ReportConsent::load(home.path()), ReportConsent::AskEachTime);
    }

    /// REQ-5: the preflight refuses code, a path and a secret's shape before
    /// any yes is possible.
    #[test]
    fn the_report_body_preflight_refuses_code_paths_and_secrets() {
        assert!(report_body_check(
            "The update preview showed three lines.\nmoai 4.2647.17 on macOS."
        )
        .is_ok());
        assert!(report_body_check("look:\n```rust\nfn main() {}\n```").is_err());
        assert!(report_body_check("see a made-up path like src/main.rs:12 with dots").is_err());
        assert!(report_body_check("my key sk-abcdef0123456789").is_err());
    }
}
