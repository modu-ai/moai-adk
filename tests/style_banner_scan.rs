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

    /// The seven steps of the flow of one piece of work, in flow order, then the four icons that
    /// stand beside it (the one icon set of the progress board, the status line and the lane
    /// reports: `skills/moai/references/home/output-style.md`).
    const FLOW_ICONS: [&str; 7] = ["🗂️", "🟡", "📨", "🔍", "🔁", "🛬", "✅"];
    const BESIDE_ICONS: [&str; 4] = ["🚧", "🙋", "🔴", "❌"];
    /// Icons that left the set: the old board used them for not started, done, blocked, review
    /// and the like. `⏸` is the base character, so the text form and the emoji form both match.
    const RETIRED_ICONS: [&str; 6] = ["⬜", "🟢", "⏸", "🔵", "👀", "⛔"];

    #[test]
    fn the_three_style_templates_use_the_one_icon_set() {
        for rel in [
            "templates/claude/output-styles/moai.md",
            "templates/claude/output-styles/moai-easy.md",
            "templates/claude/output-styles/moai-learn.md",
        ] {
            let text = fs::read_to_string(Path::new(ROOT).join(rel)).unwrap();
            for old in RETIRED_ICONS {
                assert!(!text.contains(old), "{rel} still carries the old icon {old}");
            }
            // the flow is one sentence: the line that starts it (🗂️) holds every step in order,
            // so a banner emoji such as ✅ earlier in the file does not count as the step
            let line = text
                .lines()
                .find(|l| l.contains(FLOW_ICONS[0]))
                .unwrap_or_else(|| panic!("{rel} does not name the flow icon {}", FLOW_ICONS[0]));
            // each step is searched after the one before it, so a banner emoji earlier in the
            // same line does not count as the step
            let mut rest = line;
            for icon in FLOW_ICONS {
                let at = rest.find(icon).unwrap_or_else(|| {
                    panic!("{rel} flow line does not name {icon} after the step before it")
                });
                rest = &rest[at + icon.len()..];
            }
            for icon in BESIDE_ICONS {
                assert!(text.contains(icon), "{rel} does not name the icon {icon}");
            }
        }
    }
}
