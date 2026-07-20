---
title: moai cc / cg / glm ランチャー
weight: 15
draft: false
---

`moai cc`、`moai cg`、`moai glm` は Claude Code を異なるバックエンド構成で起動する 3 つのランチャーです。3 コマンドとも設定を調整したうえで `exec` で現在のプロセスを Claude Code に置き換えます。どのモデルがどの仕事をするかがそのままコストになるため、ランチャー選択はトークノミクスの最初の決定です。

## 3 ランチャーの比較

| ランチャー | バックエンド | 用途 |
|------------|--------------|------|
| `moai cc` | Claude 専用 | 標準実行 — すべてのエージェントが Claude モデルを使用 |
| `moai glm` | GLM 専用 | すべてのエージェントが Z.AI プロキシ経由で GLM モデルを使用 |
| `moai cg` | Claude + GLM ハイブリッド | リーダーは Claude、チームメイトは GLM (60-70% のコスト削減) |

## moai cc — Claude バックエンド

```bash
moai cc [-p profile] [-- claude-args...]
```

`.claude/settings.local.json` から GLM 専用の環境変数を取り除き、team モードが有効だった場合はリセットしたうえで Claude Code を起動します。

| フラグ | 説明 |
|--------|------|
| `-p, --profile <name>` | 名前付き Claude プロファイルを使用 (`~/.moai/claude-profiles/<name>/`) |
| `--permission-mode <mode>` | 権限モードを指定 |
| `-b, --bypass` | `--permission-mode bypassPermissions` のショートハンド |
| `-c, --continue` | 前回のセッションを継続 |
| `-m, --model <model>` | モデル選択をオーバーライド |
| `--chrome` / `--no-chrome` | Chrome MCP のトグル |

権限モードは `default`、`acceptEdits`(プロジェクトデフォルト)、`plan`、`auto`、`bypassPermissions`、`dontAsk` のいずれかです。`auto` モードはバックグラウンド分類器が動作を検査するもので、Team プラン + Sonnet/Opus 4.6 以上が必要です。

## moai glm — GLM バックエンド

```bash
moai glm setup <api-key>   # API キーを保存 (初回のみ)
moai glm                   # GLM バックエンドで起動
moai glm -p work           # 'work' プロファイルで起動
moai glm status            # 資格情報の状態を確認
```

`~/.moai/.env.glm` から GLM 資格情報を読み取り、`ANTHROPIC_AUTH_TOKEN`、`ANTHROPIC_BASE_URL` などの環境変数を注入したうえで Claude Code を起動します。

| サブコマンド | 説明 |
|--------------|------|
| `moai glm setup [api-key]` | GLM API キーを保存 |
| `moai glm status` | 現在の GLM 資格情報の状態を表示 |

{{< callout type="warning" >}}
GLM は `auto` 権限モードをサポートしません (サードパーティプロバイダ)。`auto` が必要な場合は `moai cc` または `moai cg` を使用してください。また Z.AI は同時リクエスト上限が低い (有料ティアで 1-3 in-flight) ため、複数エージェントの並列実行は `moai cg` ハイブリッドモードのほうが安定します。
{{< /callout >}}

## moai cg — Claude + GLM ハイブリッド

```bash
moai cg [-p profile]
```

CG は "Claude + GLM" の略で、コスト最適化のチーム構成です。

- **リーダー** (現在の tmux ペイン): Claude モデルを使用 (opus/sonnet)
- **チームメイト** (新しい tmux ペイン): Z.AI プロキシ経由で GLM モデルを使用

実行時に tmux セッションを検証し、リーダーペインでは GLM 環境を取り除き (Claude)、tmux セッションには GLM 環境を注入し (チームメイト)、`teammateMode=tmux` と `team_mode: cg` を設定します。

**前提条件**:

1. `moai glm setup <api-key>` で GLM API キーを設定
2. ペイン単位の環境隔離のため tmux セッション内部で実行

## プロファイル (`-p` フラグ)

3 ランチャーとも `-p <name>` で名前付きプロファイルを指定すると `CLAUDE_CONFIG_DIR` が `~/.moai/claude-profiles/<name>/` に設定されます。複数のアカウント・設定セットを分離して運用するときに使用します。

## 関連ドキュメント

- [CG モード (Claude + GLM)](/ja/multi-llm/cg-mode)
- [プロファイル管理](/ja/cli-reference/profile)
- [セキュリティノート](/ja/advanced/security-notes) — GLM 資格情報パスのセキュリティモデル
- [CLI 概要](/ja/getting-started/cli)
