---
title: moai worktree ワークツリー
weight: 25
draft: false
---

`moai worktree` (エイリアス `moai wt`) は、並列 SPEC 開発に使う Git ワークツリーを管理します。同期、完了処理、削除、整理、レジストリ復旧、そして隔離されたエージェント実行を包む状態ガードまで、8 つのサブコマンドを提供します。

## ワークツリーへの進入と一覧表示はこのコマンドの役目ではありません

`moai worktree` はワークツリーを**管理**するだけで、その中へ入ったり一覧を表示したりはしません。

| やりたいこと | 使うコマンド |
|-------------|-------------|
| ワークツリーの中で作業を始める | `moai cc -w <name>` (または `moai glm -w` / `moai cg -w`) |
| 現在のセッションは残したまま新しい tmux ウィンドウで開く | `moai cc -w <name> --spawn` |
| ワークツリーの一覧を確認する | `git worktree list` |
| ワークツリーを新しく作る | `moai cc -w <name>` (`.claude/worktrees/<name>/` を自動生成) または `git worktree add` |

`-w` に短い名前を渡すと `.claude/worktrees/<name>/` の下で解決され、存在しなければ新しく作られます。絶対パスを渡した場合は `~/.moai/worktrees/` または `<プロジェクト>/.claude/worktrees/` 配下の既存ワークツリーへ再進入します。それ以外の絶対パスは拒否されます。

## サブコマンド

| コマンド | 説明 |
|--------|------|
| `moai worktree sync [branch-name]` | ベースブランチの変更をワークツリーへ取り込む |
| `moai worktree done <branch-name>` | ブランチに紐づくワークツリーを削除、任意でブランチも削除 |
| `moai worktree remove <path>` | 指定したパスのワークツリーを削除 |
| `moai worktree clean` | stale な参照の整理、マージ済み・放置されたワークツリーの整理 |
| `moai worktree recover` | ワークツリーレジストリを復旧 |
| `moai worktree snapshot` | 作業ツリーの状態をスナップショットとして取得 |
| `moai worktree verify` | 現在の作業ツリーをスナップショットと照合 |
| `moai worktree restore` | 作業ツリーをスナップショット HEAD の状態へ戻す |

## moai worktree sync

```bash
moai worktree sync [branch-name]
```

ブランチ名を渡すとそのブランチのワークツリーを、省略すると現在のディレクトリのワークツリーを同期します。

| フラグ | 説明 |
|--------|------|
| `--base <branch>` | 基準ブランチ (デフォルト: `main`) |
| `--strategy <mode>` | `merge` (デフォルト) または `rebase` |

## moai worktree done

```bash
moai worktree done <branch-name>
```

ブランチ名は必須です。そのブランチを使っているワークツリーを探して削除し、必要ならブランチも削除します。**マージは行いません** — ベースブランチへのマージは `git merge` や PR で別途完了させてください。

| フラグ | 説明 |
|--------|------|
| `--force` | 未コミットの変更があっても強制削除 |
| `--delete-branch` | ワークツリー削除後にブランチも削除 |
| `--auto` | 自動化用の無出力モード (例: PR マージ後の整理)。ワークツリーが見つからなくてもエラー終了しません |

## moai worktree remove

```bash
moai worktree remove <path>
```

引数はブランチ名ではなく**ファイルシステムのパス**です。

| フラグ | 説明 |
|--------|------|
| `--force` | 未コミットの変更があっても強制削除 |

## moai worktree clean

```bash
moai worktree clean [--merged-only | --stale] [--yes] [--json] [--base <branch>]
```

フラグなしで実行すると、stale なワークツリー参照だけを prune します。ランチ台帳(`~/.moai/claude-profiles/launch.yaml` の `projects:` 項目)のうち、プロジェクトディレクトリが消えた項目も一緒に整理し、削除した項目数を出力します。`moai worktree remove` と `moai worktree done` も、破棄するツリーの項目を同じ方法で回収します。

| フラグ | 説明 |
|--------|------|
| `--merged-only` | ブランチがベースへマージ済みのワークツリーのみ削除 |
| `--stale` | 失うもののない放置されたワークツリーをまとめて整理 (デフォルトはプレビュー) |
| `--yes` | `--stale` のプレビューではなく実際の削除を実行 |
| `--json` | `--stale` と併用: 保護対象外のすべてのワークツリーを、保持理由と四つの判定（dirty・マージ・アンカー・無視されたコンテンツ）とともに JSON で出力。何も削除せず、`--yes` より優先されます |
| `--base <branch>` | `--merged-only` · `--stale` の判定基準ブランチ (デフォルト: `origin/main`) |

`--stale` と `--merged-only` は同時に使えません。

### --stale の安全ルール

ワークツリーは次の 2 つの条件を**すべて**満たすときにのみ削除対象になります。

1. 作業ツリーがクリーンである — 未コミットの変更も untracked ファイルもない
2. ブランチにベースを超える固有のコミットがない

1 つでも外れるとそのワークツリーは維持され、維持した理由が併せて出力されます。**ブランチは決して削除しない**ため、ワークツリーのディレクトリが消えてもコミットはブランチ名のまま残ります。メインチェックアウトと、コマンドを実行中のワークツリーは常に保護対象です。

`--stale` はプレビューがデフォルトです。実際に削除するには `--yes` を付けてください。

## moai worktree recover

```bash
moai worktree recover
```

`git worktree repair` でワークツリー管理ファイルを修復したうえで stale な参照を prune し、最終的に認識されたワークツリーの一覧を出力します。フラグはありません。

## moai worktree snapshot

```bash
moai worktree snapshot
```

HEAD、ブランチ、porcelain 状態、`.moai/specs/` 配下の untracked ファイルを取得し、`.moai/state/` へ JSON として記録します。隔離されたエージェントを呼び出す直前に取っておくための用途です。

| フラグ | 説明 |
|--------|------|
| `--out <path>` | スナップショットの保存先パス (デフォルト: `.moai/state/worktree-snapshot-<id>.json`) |
| `--agent-name <name>` | エージェント名を記録 (後の verify 段階で参照) |

## moai worktree verify

```bash
moai worktree verify --snapshot <path>
```

現在の作業ツリーをスナップショットと照合します。`--snapshot` は**必須**です。

| フラグ | 説明 |
|--------|------|
| `--snapshot <path>` | 事前スナップショット JSON のパス (必須) |
| `--agent-response <path>` | エージェント応答 JSON — 空の `worktreePath` の検出用 |
| `--agent-name <name>` | divergence · suspect のログに記録するエージェント名 |

| 終了コード | 意味 |
|-----------|------|
| `0` | clean |
| `1` | divergence を検出 |
| `2` | suspect (空の `worktreePath`) |
| `3` | 両方 |

## moai worktree restore

```bash
moai worktree restore --snapshot <path>
```

`git restore --source=<スナップショット HEAD> --staged --worktree :/` を実行し、追跡中のファイルをスナップショット HEAD の状態へ戻します。**untracked ファイルは git で復元できない**ためパスの一覧が示されるだけで、自分で作り直す必要があります。

| フラグ | 説明 |
|--------|------|
| `--snapshot <path>` | スナップショット JSON のパス (必須) |
| `--dry-run` | 実行せずに、実行予定の git コマンドだけを出力 |

## 例

```bash
# ワークツリーを作りながらそのまま進入 (.claude/worktrees/feat-auth/)
moai cc -w feat-auth

# 現在のセッションを保ったまま新しい tmux ウィンドウで GLM チームメイトを起動
moai cg -w feat-auth --spawn

# ワークツリーの一覧
git worktree list

# 現在のワークツリーを main と同期 (merge)
moai worktree sync

# 特定のワークツリーを rebase で同期
moai worktree sync feature/SPEC-AUTH-001 --strategy rebase

# 放置されたワークツリーをまずプレビューし、確認してから実際に削除
moai worktree clean --stale
moai worktree clean --stale --yes

# マージが終わったらワークツリーを整理 + ブランチ削除
moai worktree done feature/SPEC-AUTH-001 --delete-branch
```

## 関連ドキュメント

- [Git Worktree 概要](/ja/worktree/) — 概念とワークフロー
- [完全ガイド](/ja/worktree/guide) — コマンド別の詳しい使い方
- [CG モード](/ja/multi-llm/cg-mode) — Claude リーダー + GLM チームメイトのハイブリッド
- [CLI 概要](/ja/getting-started/cli)
