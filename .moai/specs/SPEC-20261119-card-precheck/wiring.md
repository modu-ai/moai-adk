# The `src/ops/card.rs` wiring (the leader applies this after C-43's piece)

`src/ops/card.rs` is inside C-43's reservation, so this branch carries the logic
(`src/card/precheck.rs`) and the tests; the two edits below connect the CLI action. They are the
whole wiring: one `Action` variant, one match arm. The error mapping needs no new code — the
`Refusal` messages already use the prefixes the operations layer derives its `kind` from
(`card precheck refused:` → `invalid_input`, `card not found:` → `not_found`, `ledger blocked:`
→ `blocked`).

## 1. The `Action` variant (beside `Show`)

```rust
    /// Reads the ledger only. Show what a dispatch should know before it starts: commits and
    /// PRs already naming the card, its SPEC folders, and decision notes colliding with the
    /// card's scope or criteria
    ///
    /// Reads the ledger only. Also reads the project's git history (nothing is written, no
    /// network is used): commits of `--integration-ref` (default: the current branch's
    /// upstream) whose subject, body or `Moai-Spec:` trailer names the card, the SPEC the
    /// ledger links to it, the SPEC folders naming it, and every `decision` note sharing a
    /// path with the card's write scope or a phrase from its criteria (card C-96, W3).
    Precheck {
        /// `C-12` or `#3066`.
        card: String,
        /// The integration ref to scan (default: the current branch's upstream).
        #[arg(long)]
        integration_ref: Option<String>,
        /// Scan at most this many commits (default 500; the report names a cut-short scan).
        #[arg(long)]
        max_commits: Option<usize>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
```

## 2. The match arm in `run` (beside `Action::Show`)

```rust
        Action::Precheck {
            card,
            integration_ref,
            max_commits,
            project,
            home,
        } => c::precheck::precheck(
            &card,
            integration_ref.as_deref(),
            max_commits,
            &project,
            home.as_deref(),
        )
        .map_err(|r| r.message),
```

## After applying

- `cargo test --offline --locked --test card_precheck_cli` — the three RED tests (AC-1..AC-3)
  turn GREEN; they fail today only with `unrecognized subcommand 'precheck'`.
- `cargo fmt --check`, `cargo clippy --offline --locked --lib -- -D warnings` — both clean on
  this branch with the logic in place (measured 2026-11-19).
