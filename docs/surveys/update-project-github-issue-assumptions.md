# Survey B: GitHub-Issue assumptions in the update, project and setup scope

Card C-51, criterion 2. Read at head `eea59f4ff1` (branch `spec/20261118-update-project-surveys`). SPEC: `SPEC-20261118-update-project-surveys`.

GitHub Issues are no longer the work ledger; the local card ledger (`moai card`) is (user decision 2026-11-16, card C-51). A row is a defect when it assumes an Issue is the unit of work (owner, report comment, claim trailer, "Issue scope"); it is kept when Issues are named only as a place secrets must not go, as a neutral fact about GitHub, or as a provenance comment.

The method and the limits are those of Survey A (`docs/surveys/update-project-user-requests.md`, "How the rows were made"): every row's cited line was read from the file at this head and contains the row's key word; the class and the planned change are a reading.

## Counts

- 22 rows: 16 defects, 6 kept.
- Not found by search: `Moai-Claim` and `Refs #` do not appear in `src/agents.rs`, `src/generation.rs` or `templates/`. The interview and question-bank text is not in `src/`: it lives in `skills/moai/` and Survey A lists it as met (`W6`, `W8`, `W11`, `W12`).
- `development_methods` accepts free strings; no Issue or ledger value exists in `src/` (B22).

## Rows

| id | path:line | the line as read | class | planned change |
| --- | --- | --- | --- | --- |
| B1 | `src/profile.rs:111` | `/// The project chose GitHub Issues, so 'github-plan' generates the GitHub templates` | defect: assumes GitHub Issues are the ledger | Stop treating `github_issues` as "Issues are my ledger"; keep reading it so old profiles stay valid; the meaning becomes "generate the GitHub templates" |
| B2 | `src/github_templates.rs:18` | `//! Generation is opted in by the profile key 'github_issues' (Issue #2106) or by the` | defect: assumes GitHub Issues are the ledger | Same as B1; the opt-in note text changes. Pins: `tests/github_profile_cli.rs:119, 161, 176, 181`, `tests/github_cli.rs:36, 111, 468`, `tests/gitignore_block_cli.rs:33` |
| B3 | `src/github_templates.rs:294` | `notes.push("not generated: not opted in; an existing .github/ never opts in".into());` | keep | Keep: a neutral fact |
| B4 | `src/github_templates.rs:336` | `"GitHub reads Issue forms only from the default branch; rendering is not_run".into(),` | keep | Keep: a fact about GitHub |
| B5 | `templates/github/en/ISSUE_TEMPLATE/moai-feature.yml:5` | `description: A new feature or behaviour. One execution owner per Issue, progress reported in a single owner...` | defect: assumes GitHub Issues are the ledger | The forms use the Issue as the work ledger (execution owner, owner report comment, leader-added labels); decide whether the GitHub forms stay at all, otherwise rewrite as card-neutral intake. Same line 5 in `moai-bugfix.yml`, `moai-investigation.yml`, `moai-followup.yml`, and the ko twins |
| B6 | `templates/github/en/ISSUE_TEMPLATE/moai-feature.yml:12` | `value: 'One Issue is one piece of work and has only one execution owner at a time. This form adds only the ...` | defect: assumes GitHub Issues are the ledger | Drop "execution owner" (also `moai-bugfix.yml:12`, ko twins) |
| B7 | `templates/github/en/ISSUE_TEMPLATE/moai-feature.yml:96` | `- [ ] Owner report comment updated'` | defect: assumes GitHub Issues are the ledger | Remove the owner-report-comment checklist item (also `moai-bugfix.yml:105`, `moai-investigation.yml:100`, `moai-followup.yml:88`, ko twins) |
| B8 | `templates/github/en/ISSUE_TEMPLATE/moai-feature.yml:76` | `label: Prerequisite Issues` | defect: assumes GitHub Issues are the ledger | Rename to "Prerequisites" and drop the leader/native-dependency text (also `moai-bugfix.yml:85`, `moai-investigation.yml:80`, `moai-followup.yml:68`, ko twins) |
| B9 | `templates/github/en/ISSUE_TEMPLATE/moai-followup.yml:10` | `value: 'Describe one piece of work split off from the original Issue or PR. Do not rewrite the body of the ...` | defect: assumes GitHub Issues are the ledger | The follow-up form assumes the Issue/PR chain: rewrite or drop |
| B10 | `templates/github/en/ISSUE_TEMPLATE/moai-feature.yml:12` | `value: 'One Issue is one piece of work and has only one execution owner at a time. This form adds only the ...` | defect: assumes GitHub Issues are the ledger | Remove the leader-labelling model (all 8 forms, around lines 10-15) |
| B11 | `templates/github/labels.yml:1` | `# MoAI label list for a project that uses MoAI's Issue forms.` | defect: assumes GitHub Issues are the ledger | Rewrite the header; `needs:decision`, `type:` and `priority:` labels only make sense if the forms stay |
| B12 | `templates/github/manifest.json:23` | `"generation_condition": "request.github_issues == true",` | defect: assumes GitHub Issues are the ledger | Update with B5-B11; every `template_sha256` must be recomputed. Pins: `tests/github_templates_test.py:41, 84-92`, `tests/github_assets_test.py:46-48`, `tests/github_cli.rs:259-307, 440-464`, `tests/labels_sync.rs` |
| B13 | `src/agents.rs:290` | `Confirm the approved SPEC and Issue scope, then proceed plan → implement → real verification → independent ...` | defect: assumes GitHub Issues are the ledger | Say "card scope" (ko :269 "카드 범위"); both lines are in the generated AGENTS.md managed section. Pins: `tests/product_cli.rs:3763`, `tests/language_setting.rs:259`, repo `AGENTS.md:73` (regenerate the managed section with `moai agents-plan` and `agents-apply`) |
| B14 | `src/agents.rs:235` | `(_, "tracked") => "Never share credentials or raw secret logs in the repo or in Issues.",` | keep | Keep: Issues are named only as a place secrets must not go (also :233, :311, :313). Pin: `tests/agents_local_merge_cli.rs:24` |
| B15 | `src/agents.rs:337` | `/// Issues (user decision 2026-11-17, SDD always; replaces decision 6892663669). 'specs/' stays` | keep | Keep: a comment citing a decision |
| B16 | `src/hooks.rs:37` | `/// 'claim-verify . --hook': the claim write boundary (rc.2 A9 #2255). Allows every call` | defect: assumes GitHub Issues are the ledger | The generated hook set ships a guard that checks Issue-claim trailers (`src/claim_verify.rs:994-998`): point it at the card reservation or make it optional. Pins: `tests/hooks_cli.rs:122, 745`, `tests/claim_verify_cli.rs:1115` |
| B17 | `src/hook_legacy.rs:243` | `the Rust build ({binary_state}); the claim guard would stay off. Put the Rust moai \` | defect: assumes GitHub Issues are the ledger | Follows B16 (also `install.rs:837, 866`, `hooks.rs:250, 286`) |
| B18 | `src/hooks.rs:47` | `/// folder, lane branch, local claim record) the model reads at session start (#2936 R9 S1).` | defect: assumes GitHub Issues are the ledger | Follows B16 |
| B19 | `src/language.rs:55` | `label_en: "Issue and PR records",` | defect: assumes GitHub Issues are the ledger | Relabel to "PR and card records"; keep the id `issue_pr_records` for compatibility. Pins: `tests/language_setting.rs:186, 238, 281`, `schemas/protocol/v1/profile.schema.json:42` |
| B20 | `src/ops/github_plan.rs:8` | `/// ('github_issues' opts in); read-only.` | defect: assumes GitHub Issues are the ledger | Follows B1 |
| B21 | `src/profile.rs:167` | `pub const WORK_FLOWS: &[&str] = &["m-flow", "github-flow", "trunk-based", "git-flow"];` | keep | Keep: git branching flows, not a ledger method |
| B22 | `src/launcher_init.rs:457` | `"development_methods": ["unspecified"],` | keep | Keep: free strings, no Issue or ledger method value exists in `src/` |

## Findings the parent should read first

1. **The generated AGENTS.md says "Issue scope" (B13).** It reaches every project's managed section, and the repository's own `AGENTS.md:73` carries it. Changing it needs the three pins named in the row and a regenerated managed section.
2. **The generated hook set still checks Issue-claim trailers (B16-B18).** `claim-verify` refuses a commit whose message has no `Moai-Claim: <issue>#<id>` on an `issue/<n>-...` lane branch. With cards as the ledger the guard must read the card reservation or be optional.
3. **The GitHub Issue forms are a work-ledger design (B5-B12).** Eight forms and `labels.yml` assume an owner, an owner report comment, leader-added labels and native dependencies. Whether the forms stay at all is a decision for the leader and the user, because the embedded `template_sha256` values and four test families pin them.

## Follow-up groups (for the leader to card; a lane adds no card)

1. Generated AGENTS.md text: B13 (with its pins and the managed-section regeneration).
2. The claim guard: B16-B18.
3. The GitHub templates: B1, B2, B5-B12, B20 (a decision first).
4. Language label: B19.
