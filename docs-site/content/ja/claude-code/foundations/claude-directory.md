---
title: .claude ディレクトリ
weight: 60
draft: false
description: "Claude Code が CLAUDE.md、settings.json、スキル、サブエージェント、hook を読み込むプロジェクトごとの設定ルート、.claude ディレクトリの構造とスコープを整理します。"
---

# .claude ディレクトリ

`.claude` ディレクトリは、Claude Code がプロジェクトごとに指示、設定、拡張機能を読み込む単一の設定ルートです。

{{< callout type="info" >}}
**ひとことで言うと**: `.claude` は Claude Code が毎セッション開始時に覗き込むプロジェクト専用の「操作盤」であり、大半は git にコミットしてチームと共有し、個人用ファイルだけを別に隔離します。
{{< /callout >}}

大半のユーザーは `CLAUDE.md` と `settings.json` の 2 ファイルを編集するだけで十分です。残りのスキル、rules、サブエージェントは必要になったときに 1 つずつ追加すればよいのです。

## .claude ディレクトリの役割

Claude Code は 2 か所から設定を読みます。1 つは作業中のプロジェクトの `.claude/` ディレクトリ、もう 1 つはホームディレクトリの `~/.claude/` です。プロジェクト内のファイルは git にコミットしてチームと共有し、`~/.claude/` のファイルはすべてのプロジェクトに適用される個人設定として残ります。

- **プロジェクトコンテキストの伝達**: `CLAUDE.md` のように Claude が「読んで従う」指針
- **動作の強制**: `settings.json` の権限 (permissions) と hook のように、Claude の遵守と無関係に「執行される」設定
- **拡張機能の保管**: スキル、サブエージェント、ダイナミックワークフローなど再利用可能な資産

ここでの重要な区分は **指針** (guidance) と **設定** (configuration) です。`CLAUDE.md` や rules は Claude が参考にする案内文であり常に守られる保証はありませんが、hook と permissions はランタイムが直接執行するため決定論的です。確実な動作が必要なら、指針ではなく hook または permissions で実装すべきです。この区分こそがハーネスエンジニアリングの最初の設計判断です — MoAI-ADK も `moai init` 一発でこのディレクトリにオーケストレーター指針 (CLAUDE.md)、品質ゲート hook、エージェント・スキル資産を配布し、プロジェクト専用のハーネスを構成します。

## プロジェクト .claude/ ディレクトリの構造

| 項目 | 場所 | コミット | 役割 |
| --- | --- | --- | --- |
| `CLAUDE.md` | プロジェクトルートまたは `.claude/` | ✓ | 毎セッションのコンテキストとしてロードされるプロジェクト指針 |
| `settings.json` | `.claude/` | ✓ | 権限、hook、環境変数、デフォルトモデルなど執行される設定 |
| `settings.local.json` | `.claude/` | - | 個人用の設定オーバーライド (自動 gitignore) |
| `rules/` | `.claude/` | ✓ | トピック別に分割した指針、ファイルパスで条件付きロード可能 |
| `skills/` | `.claude/` | ✓ | `/name` で呼ぶか Claude が自動呼び出しするスキル |
| `commands/` | `.claude/` | ✓ | 単一ファイルのプロンプト (スキルと同じ仕組み) |
| `agents/` | `.claude/` | ✓ | 独立したコンテキストウィンドウを持つサブエージェント定義 |
| `workflows/` | `.claude/` | ✓ | 複数のサブエージェントを調律するダイナミックワークフロースクリプト |
| `hooks/` | `.claude/` | ✓ | hook が実行するスクリプト (settings.json で登録) |
| `agent-memory/` | `.claude/` | ✓ | サブエージェント専用の永続メモリ |
| `.mcp.json` | プロジェクトルート | ✓ | チーム共有の MCP サーバー構成 |
| `.worktreeinclude` | プロジェクトルート | ✓ | worktree 作成時にコピーする gitignore パターン |

### 指針ファイル (Claude が読むもの)

**`CLAUDE.md`**: プロジェクトのルール、よく使うコマンド、アーキテクチャの背景を収めます。毎セッション全文がコンテキストとしてロードされるため 200 行以下が推奨され、長くなったら rules へ分離します。

**`rules/*.md`**: `paths:` フロントマターがなければセッション開始時にロードされ、`paths:` グロブがあれば該当ファイルがコンテキストに入るときのみロードされます。`CLAUDE.md` が 200 行に近づいたらトピック別の rule に分割するのがベストプラクティスです。

### 執行設定 (Claude Code が強制するもの)

**`settings.json`**: `permissions` (ツール・コマンドの許可/ブロック)、`hooks` (イベント時点のスクリプト実行)、`statusLine`、`model`、`env`、`outputStyle` のキーを収めます。

**`settings.local.json`**: 同じスキーマですが個人用で、コミットしません。チームのデフォルトと異なる権限が必要なときに使います。

### 拡張資産

**`skills/<name>/SKILL.md`**: フォルダ単位のスキルで、参考ドキュメント・テンプレート・スクリプトを一緒にバンドルできます。

**`commands/*.md`**: 単一ファイルのプロンプトです。公式にはスキルと同じ仕組みであり、新しいワークフローはスキルとして書くことが推奨されます。

**`agents/*.md`**: 独自のシステムプロンプトとツールアクセス権限を持つサブエージェントです。新しいコンテキストウィンドウで実行され、メイン会話をきれいに保ちます。

**`workflows/*.js`**: 多数のサブエージェントをスポーン・調律するダイナミックワークフロースクリプトです。

## グローバル ~/.claude/ ディレクトリの構造

| 項目 | 場所 | 役割 |
| --- | --- | --- |
| `CLAUDE.md` | `~/.claude/` | すべてのプロジェクトに適用される個人指針 |
| `settings.json` | `~/.claude/` | すべてのプロジェクトのデフォルト設定 (プロジェクト設定で上書き) |
| `keybindings.json` | `~/.claude/` | カスタムキーボードショートカット |
| `skills/` | `~/.claude/` | すべてのプロジェクトで使える個人スキル |
| `commands/` | `~/.claude/` | すべてのプロジェクトで使える個人コマンド |
| `agents/` | `~/.claude/` | すべてのプロジェクトで使える個人サブエージェント |
| `workflows/` | `~/.claude/` | すべてのプロジェクトで使える個人ワークフロー |
| `output-styles/` | `~/.claude/` | 個人の出力スタイル |
| `projects/` | `~/.claude/` | プロジェクト別のセッション記録、会話トランスクリプト、自動メモリ |

## 設定のスコープと優先順位

同じ設定が複数の場所に存在でき、より具体的なスコープが優先されます。スコープはエンタープライズ、ユーザー、プロジェクトの 3 段階に分かれます。

| スコープ | 場所 | 適用範囲 |
| --- | --- | --- |
| エンタープライズ | `managed-settings.json` (OS ごとのシステムパス) | 組織全体 (ユーザーによるオーバーライド不可、最優先) |
| ユーザー (グローバル) | `~/.claude/` | すべてのプロジェクト (個人のデフォルト) |
| プロジェクト | `.claude/` | 現在のプロジェクト (チーム共有) |
| プロジェクトローカル | `.claude/settings.local.json` | 現在のプロジェクト、個人 (ユーザー編集ファイルの中で最優先) |

**配列の設定** (`permissions.allow` など) はすべてのスコープの値が **マージされます**。**スカラーの設定** (`model` など) は最も具体的なスコープの **単一値が使われます**。

## バージョン管理の対象 vs 除外

| ファイル | コミット | 理由 |
| --- | --- | --- |
| `CLAUDE.md`, `rules/`, `settings.json` | ✓ | チームが共有するコンテキストとポリシー |
| `skills/`, `commands/`, `agents/`, `workflows/` | ✓ | チームが共有する拡張資産 |
| `.mcp.json` | ✓ | チーム共有の MCP サーバー構成 |
| `settings.local.json` | - | 個人のオーバーライド (自動 gitignore) |
| `~/.claude/` 全体 | - | すべてのプロジェクトに適用される個人設定 |
| `CLAUDE.local.md` | - | プロジェクト別の個人指針 (手動作成後 `.gitignore` に追加) |

Claude Code は `settings.local.json` を初めて作成するとき、`.gitignore` へ自動的に追加します。

## 関連ドキュメント

- [settings.json ガイド](/ja/advanced/settings-json)
- [CLAUDE.md ガイド](/ja/advanced/claude-md-guide)
- [Statusline システム](/ja/advanced/statusline)

## 参考資料

- [Explore the .claude directory (Claude Code 公式ドキュメント)](https://code.claude.com/docs/en/claude-directory)

{{< callout type="tip" >}}
新しいプロジェクトなら、まず `CLAUDE.md` と `settings.json` の 2 ファイルだけを埋め、チームの権限・hook はプロジェクトの `settings.json` に、自分だけが使う権限は `settings.local.json` に置けば、git の衝突なくきれいに始められます。
{{< /callout >}}
