---
title: moai github GitHub 連携
weight: 92
draft: false
---

`moai github` は GitHub イシューのパース、SPEC リンク、ワークフロー自動化コマンドを提供します。

共通フラグとして `--dry-run` (変更なしで実行内容のみ表示) を受け取ります。

## サブコマンド

| コマンド | 説明 |
|--------|------|
| `moai github parse-issue <number>` | GitHub イシューをパースして内容を表示 |
| `moai github link-spec <issue-number> <spec-id>` | GitHub イシューと SPEC ドキュメントを双方向にリンク |

## moai github parse-issue

```bash
moai github parse-issue 123
```

GitHub イシューをパースして内容を表示します。

## moai github link-spec

```bash
moai github link-spec 123 SPEC-AUTH-001
```

GitHub イシューと SPEC ドキュメントの間に双方向リンクを作成します。

## 例

```bash
# イシューをパース
moai github parse-issue 42

# イシューを SPEC にリンク (プレビュー)
moai github link-spec 42 SPEC-AUTH-001 --dry-run

# 実際にリンク
moai github link-spec 42 SPEC-AUTH-001
```

## 関連ドキュメント

- [moai pr](/ja/cli-reference/pr) — PR CI 監視
- [CLI 概要](/ja/getting-started/cli)
