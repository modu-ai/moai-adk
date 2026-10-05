# t204 — v3.1.3 documentation preparation (partial verdict)

## Claim

The four locale copies of `advanced/codex-dual-harness.md` now describe the Codex and Claude Code desktop setup paths and keep the v3.1.3 release gate explicit. This proves documentation preparation only. It does not close t204 or authorize a release.

## Evidence

Measured in `WT-v313-docs-prep` from `9da000bf282d9aa8109dc8cb573a8fb47a3e64d3` on 2026-09-28:

```text
$ git diff --check
exit 0
$ git diff --numstat
8  0  docs-site/content/en/advanced/codex-dual-harness.md
8  0  docs-site/content/ja/advanced/codex-dual-harness.md
8  0  docs-site/content/ko/advanced/codex-dual-harness.md
8  0  docs-site/content/zh/advanced/codex-dual-harness.md
$ DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh
exit 0
Errors:   0
Warnings: 0
OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.
```

Hugo was run with `--source docs-site --destination <TemporaryDirectory> --minify`; the temporary destination was automatically cleaned up. The observed result was `exit= 0 sitemap= True warning_or_error_lines= 0`. The output was checked for `WARN` and `ERROR` lines.

The Codex trust and `/hooks` text was checked against the [official Codex hooks documentation](https://developers.openai.com/codex/hooks); `/hooks` is described as a CLI action. The shared Claude Code Desktop configuration and remote SSH execution paths were checked against the [official Claude Code Desktop documentation](https://code.claude.com/docs/en/desktop#shared-configuration).

## Baseline-attribution

The new worktree was created from local `develop` at `9da000bf282d9aa8109dc8cb573a8fb47a3e64d3`; local `develop` and `origin/develop` matched at that SHA before editing. The older, dirty `.claude/worktrees/t204` was read as a draft and left untouched. Only the new eight-line section per locale was applied to the current baseline, preserving intervening documentation changes.

## Gaps

- The operator has not declared completion of their own MoAI, Codex, and desktop-app tests. This is t204's stated completion signal.
- No release tag, GoReleaser run, deployment, remote push, or production docs-site rendering was performed here.
- The original t204 worktree remains dirty and has not been reconciled or disposed.

## Residual-risk

The four locale paragraphs describe configuration and release conditions, not observed desktop behavior. The operator's actual tests may require correcting the wording before publication.
