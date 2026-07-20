---
title: "GitHub 連携ガイド"
description: "moai github サブコマンドで Issue をパースし SPEC と連携する"
draft: false
weight: 10
---

MoAI-ADK の GitHub 連携機能は、GitHub Issue をパースし SPEC ドキュメントと
連携する軽量な CLI ツールを提供します。すべてのコマンドはローカルに
インストールされた `gh` CLI を通じて、現在のリポジトリの Issue データを
取得します。

> **範囲の案内**: このページは実際に配布される `moai github` サブコマンドと
> 一緒に提供される GitHub Actions 資産のみを扱います。複数の LLM を PR に
> パネルとして付ける「マルチ LLM レビューパネル」は現在の配布リリースには
> 含まれていません。

## 前提要件

- MoAI-ADK インストール (macOS · Linux · Windows)
- GitHub CLI (`gh`) のインストールおよび認証 (`gh auth login`)
- GitHub リポジトリ

## moai github サブコマンド

`moai github` は 2 つのアクティブなサブコマンドを提供します。共通で
`--dry-run` フラグをサポートし、実際の変更なしで行う作業だけを事前に
確認できます。

### Issue パース: `moai github parse-issue`

```bash
moai github parse-issue 123
```

`gh` CLI で指定した番号の Issue を取得し、番号・タイトル・作成者・ラベル・本文
要約・コメント数をカード形式で出力します。

### SPEC 連携: `moai github link-spec`

```bash
moai github link-spec 123 SPEC-ISSUE-123
```

GitHub Issue と SPEC ドキュメントの間に双方向リンクを作り、そのマッピングを
`.moai/github-spec-registry.json` に保存します。SPEC ID は保存前に
形式検証を経ます。

```bash
# 実際の変更なしで計画のみ確認
moai github link-spec 123 SPEC-ISSUE-123 --dry-run
```

## 一緒に配布される GitHub Actions 資産

`moai init` は `.github/` 配下に次の 2 つの資産を配布します。

### Label Sync ワークフロー (`.github/workflows/label-sync.yml`)

`.github/labels.yml` を単一の真実の供給源としてリポジトリのラベルを
同期します。

- **トリガー**: `workflow_dispatch` (手動、`dry_run` 入力対応) または
  `.github/labels.yml` / ワークフローファイルが `main` に push されたときに自動実行
- **権限**: `issues: write`、`pull-requests: write`、`contents: read`
- **動作**: EndBug/label-sync アクションで `labels.yml` → リポラベルに反映

### detect-language コンポジットアクション (`.github/actions/detect-language/action.yml`)

リポジトリの最初のソースファイル拡張子を基準に主言語を検出し、
`language` 出力値として送り出します。

- **対応言語 (16 個)**: Go, Python, TypeScript, JavaScript, Rust, Java,
  Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift
- **実装メモ**: `find ... -print -quit` で最初のマッチ後に即座に終了し、
  `set -o pipefail` 環境で broken-pipe 失敗を避けます

## トラブルシューティング

### `gh` コマンドが見つからないとき

`moai github` サブコマンドはローカルの `gh` CLI に依存します。`gh --version` で
インストールを確認し、`gh auth login` で認証を済ませてください。

### Issue を取得できないとき

現在のディレクトリが対象リポジトリの作業ツリー内にあるか、そして `gh` が
該当リポにアクセス権を持っているか確認してください。

### SPEC ID 検証失敗

`link-spec` は `SPEC-` 接頭辞に従う有効な SPEC ID のみを受け付けます。ID
形式を確認した後、再実行してください。

## 次のステップ

- [CLI リファレンス参照](/ja/workflow-commands/)
- [Workflow 設定参照](/ja/advanced/settings-json/)
- [セキュリティポリシー確認](/ja/advanced/security-notes/)
