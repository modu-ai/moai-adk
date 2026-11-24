//! The output style banners carry the moai mark: `<emoji> MoAI 🗿 <name> ─…` (user instruction
//! 2026-11-21). Nothing shipped keeps the old `MoAI-Easy ★` or `MoAI ★` wordings, no bare `★`
//! banner marker remains in the shipped styles, and the style references pin the new form.
//! Walks the shipped tree like the other repo scans; listed in tools/repo_scans.json.

mod style_banner {
    use std::{
        fs,
        path::{Path, PathBuf},
    };

    const ROOT: &str = env!("CARGO_MANIFEST_DIR");
    const OLD_BANNERS: [&str; 2] = ["MoAI-Easy ★", "MoAI ★"];

    fn files(dir: &Path, out: &mut Vec<PathBuf>) {
        for e in fs::read_dir(dir).unwrap() {
            let p = e.unwrap().path();
            if p.is_dir() {
                files(&p, out);
            } else {
                out.push(p);
            }
        }
    }

    #[test]
    fn shipped_paths_keep_no_old_banner_wording() {
        let root = Path::new(ROOT);
        // `templates/` holds the shipped style templates; `skills/moai/` is the shipped skill
        // package whose references state the banner.
        let mut shipped = Vec::new();
        for dir in ["templates", "skills/moai"] {
            files(&root.join(dir), &mut shipped);
        }
        assert!(
            shipped.len() > 30,
            "scan found too few files: {}",
            shipped.len()
        );
        let mut offenders = Vec::new();
        for path in shipped {
            let Ok(text) = fs::read_to_string(&path) else {
                continue;
            };
            let hits: Vec<&str> = OLD_BANNERS
                .into_iter()
                .filter(|old| text.contains(old))
                .collect();
            if !hits.is_empty() {
                offenders.push(format!("{}: {}", path.display(), hits.join(", ")));
            }
        }
        assert!(
            offenders.is_empty(),
            "old banner wording or bare star marker in shipped paths:\n{}",
            offenders.join("\n")
        );
    }

    /// SPEC-20261124-card-flow-icons REQ-10: no shipped text (the templates, the style texts
    /// among them, and the skill package with its agents and references) holds an icon of the
    /// earlier card flow; the table `moai_cli::statusline::icons` names them.
    #[test]
    fn no_shipped_text_holds_an_icon_of_the_earlier_card_flow() {
        let root = Path::new(ROOT);
        let mut shipped = Vec::new();
        for dir in ["templates", "skills/moai"] {
            files(&root.join(dir), &mut shipped);
        }
        assert!(shipped.len() > 30, "scan found too few files");
        let mut offenders = Vec::new();
        for path in shipped {
            let Ok(text) = fs::read_to_string(&path) else {
                continue;
            };
            let hits: Vec<&str> = moai_cli::statusline::icons::RETIRED_FLOW
                .into_iter()
                .filter(|icon| text.contains(icon))
                .collect();
            if !hits.is_empty() {
                offenders.push(format!("{}: {}", path.display(), hits.join(" ")));
            }
        }
        assert!(
            offenders.is_empty(),
            "icons of the earlier card flow in shipped text:\n{}",
            offenders.join("\n")
        );
    }

    #[test]
    fn the_style_references_pin_the_new_banner_form() {
        for rel in [
            "skills/moai/references/home/output-style.md",
            "templates/claude/output-styles/moai-easy.md",
            "templates/claude/output-styles/moai.md",
            "templates/claude/output-styles/moai-learn.md",
        ] {
            let text = fs::read_to_string(Path::new(ROOT).join(rel)).unwrap();
            assert!(text.contains("MoAI 🗿"), "{rel} lost the new banner mark");
            // every banner mark in a shipped style is the moai emoji: a bare star left in a
            // style file is a banner the change missed
            assert!(!text.contains('★'), "{rel} still carries a star marker");
        }
    }

    /// The card-flow icon set of the progress board, the status line and the lane reports is the
    /// table `moai_cli::statusline::icons` (SPEC-20261124-card-flow-icons): the three templates
    /// spell the flow sentence and the beside sentence the table builds, and carry none of the
    /// icons that left the set.
    #[test]
    fn the_three_style_templates_use_the_one_icon_set() {
        let flow = moai_cli::statusline::icons::flow_sentence();
        let beside = moai_cli::statusline::icons::beside_sentence();
        for rel in [
            "templates/claude/output-styles/moai.md",
            "templates/claude/output-styles/moai-easy.md",
            "templates/claude/output-styles/moai-learn.md",
        ] {
            let text = fs::read_to_string(Path::new(ROOT).join(rel)).unwrap();
            for old in moai_cli::statusline::icons::RETIRED {
                assert!(
                    !text.contains(old),
                    "{rel} still carries the old icon {old}"
                );
            }
            assert!(text.contains(&flow), "{rel} does not spell `{flow}`");
            assert!(text.contains(&beside), "{rel} does not spell `{beside}`");
        }
    }
}
