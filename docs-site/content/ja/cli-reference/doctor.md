---
title: moai doctor 診断
weight: 60
draft: false
---

`moai doctor` は包括的なシステム診断を実行します。Claude Code の設定、依存関係、プロジェクト構造、言語別の開発ツール、環境を検査し、検出された問題に対する修正案を提示できます。

## 概要

```bash
moai doctor [OPTIONS]
```

## フラグ

| フラグ | 説明 |
|--------|------|
| `-v, --verbose` | 詳細な診断情報 (ツールのバージョン、言語検出) を表示 |
| `--fix` | 検出された問題の修正案を提示 |
| `--export` | 診断結果を JSON ファイルにエクスポート |
| `--check <tool>` | 特定のチェックのみ実行 (例: git, go, config) |

## サブコマンド

`moai doctor` は特定領域を掘り下げるサブコマンドを提供します。

| コマンド | 説明 |
|----------|------|
| `moai doctor config` | 設定診断 — マージされた設定を provenance 付きで検査 |
| `moai doctor hook` | 27 イベントのフックカバレッジ表を表示 |
| `moai doctor permission` | 権限解決の診断 |
| `moai doctor sandbox` | サンドボックスバックエンドの可用性診断 |

`moai doctor config` はさらに `dump` (マージ設定のダンプ) と `diff <tier-a> <tier-b>` (2 つの設定ティアの比較) を提供します。

## 例

```bash
moai doctor                            # 全体診断
moai doctor --verbose                  # 詳細診断
moai doctor --export diagnostics.json  # 結果をエクスポート
moai doctor hook                       # フックカバレッジ表
```

---

関連: [プロジェクト状態](/cli-reference/status) · [CLI 概要](/getting-started/cli)
