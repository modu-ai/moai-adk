---
title: "CG Mode"
draft: true
aliases: ["/ko/multi-llm/cg-mode"]
---

This page is only available in Korean.

> **Security note**: CG mode tmux environment variable injection now routes GLM tokens through a `0o600` temp file via `tmux source-file`, keeping the token out of `ps auxe` and `/proc/<pid>/cmdline` (CWE-214). See [Security Notes — CWE-214](/zh/advanced/security-notes/#cwe-214) for the full threat model and self-audit steps.
