//! The one write check of the HOME state folder. A command that writes under HOME calls
//! [`check`] before its first write: on a HOME whose state folder belongs to the Go generation,
//! is mixed, was left by a stopped move or cannot be used, it answers with a refusal that says
//! nothing was written and which step clears the state. It reads only.

use crate::state_root::{classify_home, MOAI_DIR};
use crate::ui_text::Lang;
use std::fmt;
use std::path::Path;

/// The generations of `state_root::classify_home` that this check refuses.
pub const REFUSED: [&str; 4] = ["go", "mixed", "interrupted", "refuse"];

/// Why a HOME was refused, and so which way out its text names.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Way {
    /// `go` or `mixed`: the Go files wait for the clean-up step of `moai update`.
    CleanUp { markers: Vec<String> },
    /// `interrupted`, the Go folder still waits in a backup folder and the state folder is gone.
    MoveBack { backups: Vec<String> },
    /// `interrupted`, the state folder holds the record of a move that did not finish.
    FinishRevert,
    /// `refuse`: the state folder is a link.
    Link,
    /// `refuse`: the state folder is not a folder.
    NotAFolder,
    /// `refuse`: a newer build wrote the marker.
    NewerMarker,
    /// `refuse`: the marker file cannot be read.
    UnreadableMarker,
    /// `refuse`: the folder layout is wrong; the text is the layout's own reason.
    Layout(String),
}

/// The answer of [`check`] for a HOME that must not be written.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct HomeWriteRefused {
    pub generation: &'static str,
    pub way: Way,
}

impl HomeWriteRefused {
    /// The refusal in the language of the stream it is shown on.
    pub fn text(&self, lang: Lang) -> String {
        let (en, ko) = match &self.way {
            Way::CleanUp { markers } => {
                let shown = markers_shown(markers);
                (
                    format!(
                        "~/.moai still holds files of the Go generation ({shown}). Run \
                         `moai update`, then run the clean-up step it shows on its own: \
                         `moai update --confirm-home <token>`. When that is done, run this \
                         command again. If the preview says a Go session is still running, \
                         close that session first."
                    ),
                    format!(
                        "~/.moai 에 Go 세대 파일이 남아 있어요({shown}). `moai update` 를 실행해 \
                         보이는 정리 단계만 먼저 실행해 주세요: \
                         `moai update --confirm-home <토큰>`. 끝나면 이 명령을 다시 실행해 \
                         주세요. 미리보기가 Go 세션이 아직 돌고 있다고 하면 그 세션을 먼저 \
                         닫아 주세요."
                    ),
                )
            }
            Way::MoveBack { backups } => {
                let names = backups.join(", ");
                (
                    format!(
                        "A move of ~/.moai stopped halfway; your Go folder is still in \
                         {names} beside it. Rename that folder back to ~/.moai, then run \
                         `moai update` and its clean-up step \
                         (`moai update --confirm-home <token>`), and run this command again."
                    ),
                    format!(
                        "~/.moai 를 옮기다 중간에 멈췄어요. Go 폴더가 옆의 {names} 에 그대로 \
                         있어요. 그 폴더 이름을 ~/.moai 로 되돌린 뒤 `moai update` 와 그 정리 \
                         단계(`moai update --confirm-home <토큰>`)를 하고, 이 명령을 다시 \
                         실행해 주세요."
                    ),
                )
            }
            Way::FinishRevert => (
                "~/.moai holds the record of a move that did not finish. Run \
                 `moai go-state-revert-plan --backup <UTC> <folder>` to see what undoing it \
                 would do, then run this command again."
                    .to_string(),
                "~/.moai 에 끝나지 않은 옮기기 기록이 있어요. \
                 `moai go-state-revert-plan --backup <UTC> <폴더>` 로 되돌리면 무엇이 바뀌는지 \
                 먼저 확인한 뒤 이 명령을 다시 실행해 주세요."
                    .to_string(),
            ),
            Way::Link => (
                "~/.moai is a link, and a link is never followed. Remove the link or replace \
                 it with the real folder, then run this command again."
                    .to_string(),
                "~/.moai 가 링크예요. 링크는 따라가지 않아요. 링크를 지우거나 진짜 폴더로 \
                 바꾼 뒤 이 명령을 다시 실행해 주세요."
                    .to_string(),
            ),
            Way::NotAFolder => (
                "~/.moai is a file, not a folder. Move that file aside, then run this command \
                 again."
                    .to_string(),
                "~/.moai 가 폴더가 아니라 파일이에요. 그 파일을 옆으로 옮긴 뒤 이 명령을 \
                 다시 실행해 주세요."
                    .to_string(),
            ),
            Way::NewerMarker => (
                "~/.moai was written by a newer moai. Use that newer moai, or move the folder \
                 aside, then run this command again."
                    .to_string(),
                "~/.moai 를 더 새로운 moai 가 썼어요. 그 moai 를 쓰거나 폴더를 옆으로 옮긴 뒤 \
                 이 명령을 다시 실행해 주세요."
                    .to_string(),
            ),
            Way::UnreadableMarker => (
                "The marker file inside ~/.moai cannot be read. Fix that file or move it \
                 aside, then run this command again."
                    .to_string(),
                "~/.moai 안의 표식 파일을 읽을 수 없어요. 그 파일을 고치거나 옆으로 옮긴 뒤 \
                 이 명령을 다시 실행해 주세요."
                    .to_string(),
            ),
            Way::Layout(reason) => (
                format!(
                    "~/.moai cannot be used: {reason}. Fix or move that item, then run this \
                     command again."
                ),
                format!(
                    "~/.moai 를 쓸 수 없어요: {reason}. 그 항목을 고치거나 옮긴 뒤 이 명령을 \
                     다시 실행해 주세요."
                ),
            ),
        };
        match lang {
            Lang::En => format!("Nothing was written. {en}"),
            Lang::Ko => format!("아무것도 쓰지 않았어요. {ko}"),
        }
    }
}

/// The first Go markers by name, with how many more there are.
fn markers_shown(markers: &[String]) -> String {
    const SHOWN: usize = 4;
    let head = markers
        .iter()
        .take(SHOWN)
        .cloned()
        .collect::<Vec<_>>()
        .join(", ");
    match markers.len().saturating_sub(SHOWN) {
        0 => head,
        more => format!("{head} and {more} more"),
    }
}

impl fmt::Display for HomeWriteRefused {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(&self.text(Lang::En))
    }
}

impl std::error::Error for HomeWriteRefused {}

impl From<HomeWriteRefused> for String {
    fn from(refused: HomeWriteRefused) -> String {
        refused.to_string()
    }
}

/// `Ok` when `home` may take a write, `Err` with the refusal when it may not. A HOME that does
/// not exist passes: each writer keeps its own answer for that. Reads only; the environment is
/// never read.
pub fn check(home: &Path) -> Result<(), HomeWriteRefused> {
    let c = classify_home(home);
    let way = match c.generation {
        "go" | "mixed" => Way::CleanUp {
            markers: c.go_markers,
        },
        // A stopped move that left its Go folder in a backup names the backup; one that left
        // only the state folder with the record of the move names the revert preview.
        "interrupted" => {
            if !c.interrupted.is_empty() && c.interrupted.iter().all(|n| n == MOAI_DIR) {
                Way::FinishRevert
            } else {
                Way::MoveBack {
                    backups: c.interrupted,
                }
            }
        }
        "refuse" => match (c.moai_dir, c.rust_marker) {
            ("link", _) => Way::Link,
            ("other", _) => Way::NotAFolder,
            (_, "newer") => Way::NewerMarker,
            (_, "unreadable") => Way::UnreadableMarker,
            _ => Way::Layout(c.reason),
        },
        _ => return Ok(()),
    };
    Err(HomeWriteRefused {
        generation: c.generation,
        way,
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::state_root::{classify_home, MOAI_DIR};
    use std::fs;

    /// The generations a HOME can be judged, each with a fixture that makes it.
    const NINE: [(&str, &str); 9] = [
        ("none", "none"),
        ("empty", "empty"),
        ("unknown", "unknown"),
        ("rust", "rust"),
        ("rust-first", "rust-first"),
        ("go", "go"),
        ("mixed", "mixed"),
        ("interrupted", "interrupted-backup"),
        ("refuse", "refuse-file"),
    ];

    fn marker(schema: u64) -> String {
        format!(r#"{{"schema_version":{schema},"generation":"moai","created_by":"test"}}"#)
    }

    /// A temporary HOME made into the named shape; the shapes are named for what they are.
    fn home_of(shape: &str) -> tempfile::TempDir {
        let dir = tempfile::tempdir().unwrap();
        let moai = dir.path().join(MOAI_DIR);
        let go_state = |m: &Path| {
            fs::create_dir_all(m.join("state")).unwrap();
            fs::write(m.join("state").join("active-sessions.json"), "[]").unwrap();
        };
        match shape {
            "none" => {}
            "empty" => fs::create_dir_all(&moai).unwrap(),
            "unknown" => {
                fs::create_dir_all(&moai).unwrap();
                fs::write(moai.join("notes.txt"), "x").unwrap();
            }
            "rust" => {
                fs::create_dir_all(&moai).unwrap();
                fs::write(moai.join("moai.json"), marker(1)).unwrap();
            }
            "rust-first" => {
                fs::create_dir_all(&moai).unwrap();
                fs::write(moai.join("moai.json"), marker(2)).unwrap();
            }
            "go" => go_state(&moai),
            "mixed" => {
                go_state(&moai);
                fs::write(moai.join("moai.json"), marker(1)).unwrap();
            }
            "interrupted-backup" => {
                let backup = dir.path().join(".moai-backup");
                fs::create_dir_all(&backup).unwrap();
                fs::write(backup.join("manifest.json"), "{}").unwrap();
            }
            "interrupted-receipt" => {
                let stamp = moai.join("backups").join("20261113T150000Z");
                fs::create_dir_all(&stamp).unwrap();
                fs::write(
                    stamp.join("go-state-receipt.json"),
                    r#"{"state":"applied"}"#,
                )
                .unwrap();
            }
            "refuse-file" => fs::write(&moai, "a file where the folder belongs").unwrap(),
            "refuse-link" => {
                let target = dir.path().join("elsewhere");
                fs::create_dir_all(&target).unwrap();
                #[cfg(unix)]
                std::os::unix::fs::symlink(&target, &moai).unwrap();
            }
            "refuse-newer" => {
                fs::create_dir_all(&moai).unwrap();
                fs::write(moai.join("moai.json"), marker(3)).unwrap();
            }
            "refuse-unreadable" => {
                fs::create_dir_all(&moai).unwrap();
                fs::write(moai.join("moai.json"), "not json").unwrap();
            }
            other => panic!("no such shape: {other}"),
        }
        dir
    }

    #[test]
    fn the_check_refuses_the_four_generations_and_passes_the_five_others() {
        for (generation, shape) in NINE {
            let home = home_of(shape);
            assert_eq!(classify_home(home.path()).generation, generation, "{shape}");
            let answer = check(home.path());
            if REFUSED.contains(&generation) {
                let refused = answer.expect_err(shape);
                assert_eq!(refused.generation, generation, "{shape}");
            } else {
                assert_eq!(answer, Ok(()), "{shape}");
            }
        }
    }

    #[test]
    fn the_check_and_the_two_older_refusals_agree_on_every_fixture() {
        for (_, shape) in NINE {
            let home = home_of(shape);
            let ours = check(home.path()).is_err();
            let backup = crate::home_root::backup_root(home.path()).is_err();
            let ledger = crate::ledger::open::check_home_ready(home.path()).is_err();
            assert_eq!(ours, backup, "{shape}: backup_root");
            assert_eq!(ours, ledger, "{shape}: check_home_ready");
        }
    }

    #[test]
    fn a_home_that_does_not_exist_passes_and_nothing_is_created() {
        let dir = tempfile::tempdir().unwrap();
        let gone = dir.path().join("no-home-here");
        assert_eq!(check(&gone), Ok(()));
        assert!(!gone.exists());
    }

    /// Each refused shape, the words its English and Korean text must hold, and the way it is
    /// named.
    const WAYS: [(&str, &str, &str); 7] = [
        ("go", "--confirm-home", "--confirm-home"),
        ("interrupted-backup", ".moai-backup", ".moai-backup"),
        (
            "interrupted-receipt",
            "go-state-revert-plan",
            "go-state-revert-plan",
        ),
        ("refuse-link", "is a link", "링크"),
        ("refuse-file", "is a file", "파일"),
        ("refuse-newer", "newer moai", "더 새로운 moai"),
        ("refuse-unreadable", "cannot be read", "읽을 수 없어요"),
    ];

    #[test]
    fn the_refusal_says_nothing_was_written_and_names_the_way_out() {
        for (shape, english, korean) in WAYS {
            let home = home_of(shape);
            let refused = check(home.path()).expect_err(shape);
            let en = refused.text(Lang::En);
            let ko = refused.text(Lang::Ko);
            assert!(en.starts_with("Nothing was written."), "{shape}: {en}");
            assert!(ko.starts_with("아무것도 쓰지 않았어요."), "{shape}: {ko}");
            assert!(en.contains(english), "{shape}: {en}");
            assert!(ko.contains(korean), "{shape}: {ko}");
            assert_eq!(
                en,
                refused.to_string(),
                "{shape}: Display is the English text"
            );
            for text in [&en, &ko] {
                let lower = text.to_lowercase();
                assert!(!lower.contains("verdict"), "{shape}: {text}");
                assert!(!lower.contains("generation_"), "{shape}: {text}");
            }
        }
    }

    #[test]
    fn a_go_home_text_names_the_markers_and_the_wait_for_a_running_session() {
        let home = home_of("go");
        let refused = check(home.path()).unwrap_err();
        let Way::CleanUp { markers } = &refused.way else {
            panic!("{:?}", refused.way);
        };
        assert!(
            markers.iter().any(|m| m.ends_with("active-sessions.json")),
            "{markers:?}"
        );
        let en = refused.text(Lang::En);
        assert!(en.contains("active-sessions.json"), "{en}");
        assert!(en.contains("moai update"), "{en}");
        assert!(en.contains("Go session"), "{en}");
        assert!(
            refused.text(Lang::Ko).contains("Go 세션"),
            "{}",
            refused.text(Lang::Ko)
        );
    }

    #[test]
    fn a_layout_error_shows_its_own_reason_and_the_refused_text_stays_plain() {
        let refused = HomeWriteRefused {
            generation: "refuse",
            way: Way::Layout("~/.moai/agents is not a folder".into()),
        };
        assert!(refused
            .text(Lang::En)
            .contains("~/.moai/agents is not a folder"));
        assert!(refused
            .text(Lang::Ko)
            .contains("~/.moai/agents is not a folder"));
    }

    #[test]
    fn the_string_conversion_is_the_display_text() {
        let home = home_of("refuse-file");
        let refused = check(home.path()).unwrap_err();
        let text = refused.to_string();
        assert_eq!(String::from(refused), text);
    }
}
