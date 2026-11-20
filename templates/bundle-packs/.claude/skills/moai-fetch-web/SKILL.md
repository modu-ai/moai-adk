---
name: moai-fetch-web
description: >
  Verbatim web-excerpt fetcher — the moai fetch web command and fetch_web MCP
  tool. Fetches a page, scores its blocks against a question in two Jev
  stages, and returns the exact passages that answer the question, unmodified.
  Use when a task needs the exact passage (a quote, a version string, a
  config line) rather than a summary. Gated behind workflow.jev.enabled
  (shipped default false).
---

# moai-fetch-web — verbatim web excerpts

`moai fetch web <url> "<question>"` (MCP: the `fetch_web` tool) fetches a
page, splits it into blocks — paragraphs, list items, table rows, code
blocks, each with its heading path — and keeps the blocks the question is
about, word for word. No paraphrase, no truncation: the output is evidence a
reader can quote.

## When to prefer it over WebFetch

- Prefer it when the task needs the EXACT passage: quoting a doc, extracting
  a version or a config value, verifying what a page literally says.
- Prefer WebFetch when a SUMMARY is wanted, when Jev is not enabled, or when
  offline. WebFetch rewrites the page through a summary model — it can never
  certify a verbatim quote.
- The capability is gated behind `workflow.jev.enabled` (shipped default
  false) and needs a TypeSafe credential. With the gate off the command
  makes ZERO network requests, prints a WebFetch fallback notice, and exits
  2 — it never silently degrades into a paraphrase.

## Reading the output

The first line begins `#` and carries one of five statuses:

| Status | Meaning | Exit code |
|---|---|---|
| `filtered` | relevant blocks kept | 0 |
| `passthrough` | whole short document kept (no scoring) | 0 |
| `no_match` | nothing scored relevant | 1 |
| `empty` | the page split to zero blocks | 1 |
| `fetch_error` | the page (or scoring) failed | 2 |

Each kept block follows as a `##` heading-path line plus the block's verbatim
text.

Handling `no_match` (exit 1): it means the question found no relevant passage
ON THAT PAGE — it is not an error and not a negative answer about anything
else. Do not retry the same page; rephrase the question or try another page.
Do not treat exit 1 as "the answer is no".

Handling `--json`: pass it when a caller parses the output. The object
carries `status`, `exit_code`, `blocks` (kind, heading, verbatim text), and
approximate token counts; branch on `status`, not on parsing the markdown
shape.

## Codex note — network_access

Under Codex, this command performs outbound network I/O (the target page and
the TypeSafe endpoint). Run it from a profile with `network_access` enabled;
without it the command fails with `fetch_error` and the condition names the
refused dial.

## Cost discipline

Every scored call spends Jev credits once the gate is on. Short documents
(under ~2,000 approximate tokens) pass through with NO scoring call. Cache
hits under `~/.moai/cache/jev` are served without a network request.
