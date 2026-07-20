---
title: moai tool-policy ツールポリシー
weight: 88
draft: false
---

`moai tool-policy` はツール/権限ポリシーの SSOT を管理します。`.moai/config/sections/tool-policy.yaml` が唯一の真実源であり、ここから `settings.json` の permissions ブロックを生成 (codegen) し、ポリシー項目を照会します。

## サブコマンド

| コマンド | 説明 |
|--------|------|
| `moai tool-policy build` | tool-policy.yaml から settings.json permissions ブロックを再生成 |
| `moai tool-policy list` | tool-policy 項目を一覧表示 (thin query) |

## moai tool-policy build

```bash
moai tool-policy build
moai tool-policy build --local-only
```

ローカル `.claude/settings.json` とテンプレート `settings.json.tmpl` の permissions ブロックを再生成します。

| フラグ | 説明 |
|--------|------|
| `--repo-root <path>` | リポジトリルート (デフォルト: cwd) |
| `--policy <path>` | tool-policy.yaml のパス (デフォルト: `<repo-root>/.moai/config/sections/tool-policy.yaml`) |
| `--local-only` | ローカル `.claude/settings.json` のみ再生成 (テンプレート .tmpl を省略) |
| `--template-only` | テンプレート settings.json.tmpl のみ再生成 (ローカルを省略) |
| `--default-mode <mode>` | `permissions.defaultMode` を上書き (デフォルト: 既存値を保持) |
| `--json` | 結果を JSON で出力 |

## moai tool-policy list

```bash
moai tool-policy list
moai tool-policy list --risk-tier irreversible --decision deny
```

| フラグ | 説明 |
|--------|------|
| `--risk-tier <read\|write\|irreversible>` | リスクティアでフィルタ |
| `--decision <allow\|deny\|ask>` | 決定でフィルタ |
| `--tool <name>` | ツール名でフィルタ (完全一致) |
| `--format <text\|json>` | 出力形式 |
| `--repo-root <path>` | リポジトリルート (デフォルト: cwd) |
| `--policy <path>` | tool-policy.yaml のパス |

## 関連ドキュメント

- [settings.json ガイド](/ja/advanced/settings-json) — permissions ブロック詳細
- [config セクションリファレンス](/ja/advanced/config-sections)
- [CLI 概要](/ja/getting-started/cli)
