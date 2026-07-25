---
title: moai ast-grep / ast-edit 構造検索・置換
weight: 77
draft: false
---

`moai ast-grep` はコードを構文木の単位でスキャンし、`moai ast-edit` はマッチしたコードを実際に置換します。テキストベースの `grep` と違い構文構造でマッチするため、空白・改行・変数名の違いに左右されません。

どちらも [ast-grep](https://ast-grep.github.io/) CLI（`sg`）を利用します。`sg` が未インストールの場合、両コマンドともエラーにはならず案内メッセージを出して終了します。

> **読み取りと書き込みが分離されています。** `ast-grep` はファイルを一切変更せず、`ast-edit` は変更します。別コマンドなので、`Bash(moai ast-grep:*)` を許可しても書き込み権限までは開きません。

## moai ast-grep — スキャン（読み取り専用）

| フラグ | 説明 |
|--------|------|
| `--format` | 出力形式: `text`（既定）・`json`・`sarif` |
| `--lang` | 指定した言語のみスキャン（例: `go`, `python`, `typescript`） |
| `--severity` | 表示する最小重大度（`error`・`warning`・`info`） |
| `--rules-dir` | ルールディレクトリのパス（既定 `.moai/config/astgrep-rules`） |
| `--dry` | 適用されるルール一覧のみ出力し、実際のスキャンは省略 |

```bash
# プロジェクト全体をスキャン
moai ast-grep ./

# Go のみ、SARIF で出力（GitHub code scanning へのアップロード用）
moai ast-grep --format=sarif --lang=go ./internal/

# error 重大度のみ表示
moai ast-grep --severity=error ./
```

## moai ast-edit — 置換（ファイルを変更）

`--dry` なしで実行すると **ファイルを直接変更します。** まず `--dry` で変更内容を確認することを推奨します。

| フラグ | 説明 |
|--------|------|
| `--dry` | ファイルを変更せず、変更される内容のみ出力 |
| `--pattern` | マッチさせる ast-grep パターン（`--rewrite` と併用） |
| `--rewrite` | 置換後のパターン（`--pattern` と併用） |
| `--rule` | 指定した ID のルールのみ適用（ルールモード） |
| `--lang` | 対象コードの言語 |
| `--rules-dir` | ルールディレクトリのパス（既定 `.moai/config/astgrep-rules`） |
| `--format` | 出力形式: `text`（既定）・`json` |

### パターンモード

`--pattern` と `--rewrite` を同時に指定すると、マッチしたコードをすべて置換します。片方のみの指定はエラーとして拒否されます。

```bash
# まずプレビュー
moai ast-edit --dry --pattern 'foo($A)' --rewrite 'bar($A)' --lang go ./internal/

# 確認後に実際に適用
moai ast-edit --pattern 'foo($A)' --rewrite 'bar($A)' --lang go ./internal/
```

### ルールモード

`--pattern` なしで実行すると、ルールディレクトリを読み込み `fix:` フィールドを宣言したルールのみを適用します。`fix:` を持たないルールは検出専用のためスキップされ、その件数が報告されます。

```bash
# すべての fix: ルールを適用（プレビュー）
moai ast-edit --dry ./internal/

# 特定のルールのみ適用
moai ast-edit --rule my-rule-id ./internal/
```

同梱のルールセット（`go/hardcoding`, `security/credentials`, `security/crypto`, `security/injection`）はすべて **検出専用** です。自動置換が意味を変えたりコンパイルを壊したりする恐れがあるため、意図的に `fix:` を入れていません。自動置換が必要な場合はプロジェクト側のルールに `fix:` を宣言してください。

## ルールファイルの場所

どちらのコマンドも既定で `.moai/config/astgrep-rules/` を読み込みます。`sgconfig.yml` が有効なルールディレクトリを定義します。

## 関連ドキュメント

- [moai loop](/ja/cli-reference/loop) — 診断ベースの反復修正ループ
- [CLI 概要](/ja/getting-started/cli)
