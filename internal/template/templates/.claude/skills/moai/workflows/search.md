---
name: moai-workflow-search
description: >
  JSONL session search with SQLite FTS5. Provides full-text search
  across indexed Claude Code session history via moai search CLI.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-03-06"
  tags: "search, sqlite, fts5, session, jsonl"
triggers:
  keywords: ["search", "recall", "find session", "session history"]
  agents: ["manager-strategy"]
  phases: ["run"]
---

# Session Search Workflow

## Overview

moai search provides full-text search across Claude Code session history
using SQLite FTS5 with trigram tokenizer for CJK language support.

## Usage

### CLI Commands

moai search "query" - Search indexed sessions
moai search "query" --branch feat/auth - Filter by branch
moai search "query" --role user - Filter by role
moai search "query" --since 2026-01-01T00:00:00Z --until 2026-03-01T00:00:00Z - Date range
moai search "query" --limit 50 - Limit results

### Auto-Indexing

Sessions are automatically indexed via SessionEnd hook.
Manual indexing: moai search --index-session <session-id>

### Database Location

~/.moai/search/sessions.db (global, all projects)

## Architecture

- Parser: internal/search/parser.go (JSONL to Messages)
- Indexer: internal/search/indexer.go (Messages to SQLite FTS5)
- Searcher: internal/search/searcher.go (Query to Results)
- CLI: internal/cli/search.go (cobra command)
- Hook: internal/hook/session_end.go (auto-index trigger)

## FTS5 Trigram Notes

The trigram tokenizer requires queries of 3+ Unicode characters.
For CJK languages (Korean, Chinese, Japanese), queries of 3+ characters use
FTS5 MATCH (BM25-ranked, high performance).

Queries of 1-2 CJK characters (e.g., Korean "인증", Japanese "認証") automatically
fall back to LIKE-based search. This is slower but ensures short CJK terms are
still searchable. The fallback is transparent to the user.
