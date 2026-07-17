---
title: moai worktree ワークツリー
weight: 25
draft: false
---

`moai worktree` (エイリアス `moai wt`) は並列 SPEC 開発のための Git ワークツリーを管理します。ワークツリーの作成・一覧表示・切り替え・同期・削除・整理を行うサブコマンドを提供します。

## サブコマンド

| コマンド | 説明 |
|--------|------|
| `moai worktree new [branch]` | 新しいワークツリーを作成 |
| `moai worktree list` | アクティブなワークツリーを一覧表示 |
| `moai worktree status` | ワークツリーの状態を表示 |
| `moai worktree switch [branch]` | ワークツリーへ切り替え |
| `moai worktree go [branch]` | シェル移動用のワークツリーパスを出力 |
| `moai worktree sync [branch]` | ベースブランチとワークツリーを同期 |
| `moai worktree done [branch]` | ワークツリーを完了して整理 |
| `moai worktree remove [path]` | ワークツリーを削除 |
| `moai worktree clean` | stale なワークツリー参照を整理 |
| `moai worktree recover` | ワークツリーレジストリを復旧 |

## moai worktree new

```bash
moai worktree new [branch-name]
```

| フラグ | 説明 |
|--------|------|
| `--path <dir>` | ワークツリーパスを指定 (デフォルト: SPEC ID の場合 `.moai/worktrees/<SPEC-ID>`、それ以外は `../<branch-name>`) |
| `--base <branch>` | ベースブランチ (デフォルト: `origin/main`、自動 fetch)。`--base main` はローカル専用コミット用 |
| `--from-current` | 現在の HEAD をワークツリーのベースとして使用 (`git fetch origin main` を省略) |
| `--tmux` | ワークツリー作成後に tmux セッションを作成 |
| `--team` | 新しいワークツリーで Claude/GLM セッションを起動 (tmux+CG → `moai glm` ウィンドウ、tmux+CC → `moai cc` ウィンドウ、no-tmux → インプロセス、no-flag → ハンドオフ案内) |

## moai worktree done

```bash
moai worktree done [branch-name]
```

| フラグ | 説明 |
|--------|------|
| `--force` | 未コミットの変更があっても強制削除 |
| `--delete-branch` | ワークツリー削除後にブランチも削除 |
| `--auto` | 自動モード — 自動化用の無出力実行 (例: PR マージ後) |

## 例

```bash
# SPEC 用のワークツリーを作成 (origin/main ベース)
moai worktree new SPEC-AUTH-001

# 現在の HEAD を基準にしたローカルワークツリー
moai worktree new feature-x --from-current

# アクティブなワークツリーを一覧表示
moai worktree list

# シェルからワークツリーへ移動
cd "$(moai worktree go SPEC-AUTH-001)"

# 完了後の整理 + ブランチ削除
moai worktree done SPEC-AUTH-001 --delete-branch
```

## 関連ドキュメント

- [ワークツリーワークフロー](/ja/advanced/autonomous-loops) — 並列開発パターン
- [CG モード](/ja/multi-llm/cg-mode) — `--team` ハイブリッド実行
- [CLI 概要](/ja/getting-started/cli)
