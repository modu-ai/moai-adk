---
title: moai pr PR 監視
weight: 96
draft: false
---

`moai pr` は CI/CD ワークフローで pull request を監視・管理するコマンドです。

## サブコマンド

| コマンド | 説明 |
|--------|------|
| `moai pr watch <PR_NUMBER>` | PR の CI チェックを監視 (または `--abort` でアクティブな監視を中断) |

## moai pr watch

```bash
moai pr watch 123
```

指定した PR 番号に対して `gh pr checks` をモニタリングします。

| フラグ | 説明 |
|--------|------|
| `--abort` | アクティブな CI 監視ループを中断 |
| `--report` | 当該 PR 番号に対するマージ準備レポートを出力 |
| `--branch <name>` | レポートのコンテキスト用ブランチ名 (デフォルト: main) |

## 例

```bash
# PR の CI チェックを監視
moai pr watch 42

# アクティブな監視を中断
moai pr watch 42 --abort

# マージ準備レポート
moai pr watch 42 --report
```

## 関連ドキュメント

- [moai github](/ja/cli-reference/github)
- [CLI 概要](/ja/getting-started/cli)
