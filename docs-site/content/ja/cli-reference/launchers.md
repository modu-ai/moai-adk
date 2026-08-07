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
moai cc [-p profile] [-w [name]] [-- claude-args...]
```

`.claude/settings.local.json` から GLM 専用の環境変数を取り除き、team モードが有効だった場合はリセットしたうえで Claude Code を起動します。

| フラグ | 説明 |
|--------|------|
| `-p, --profile <name>` | 名前付き Claude プロファイルを使用 (`~/.moai/claude-profiles/<name>/`) |
| `--permission-mode <mode>` | 権限モードを指定 |
| `-b, --bypass` | `--permission-mode bypassPermissions` のショートハンド |
| `-c, --continue` | 前回のセッションを継続 |
| `-m, --model <model>` | モデル選択をオーバーライド |
| `-w, --worktree [name]` | 隔離された git worktree (`.claude/worktrees/<name>/`) で起動 — 名前を省略すると自動生成 |
| `--spawn` | 現在のプロセスを置き換える代わりに **新しい tmux ウィンドウ** にセッションを開き、呼び出し元セッションはそのまま維持 (下の "Team window spawn" 参照) |
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

## 隔離 worktree (`-w` フラグ)

3 ランチャーとも `-w [name]` で隔離された git worktree の中からセッションを開始できます。`cd` でディレクトリを移動してから起動する 2 ステップを 1 コマンドにまとめます。

```bash
moai cc -w feat-login    # .claude/worktrees/feat-login/ で開始
moai cc -w               # 名前を自動生成
moai glm -w feat-login   # GLM バックエンドでも同様
moai cg -w feat-login    # ハイブリッドでも同様
```

動作ルール:

- worktree のパスは `.claude/worktrees/<name>/` です。`<name>` は **worktree 名**であり、ブランチ名でも SPEC ID でもありません。
- 同名の worktree が既に存在する場合は**新規作成せず再利用**します。そのため、以前のセッションが作業していたツリーへの再入場パスとしても使えます。
- 名前を省略すると Claude Code が自動的に命名します。
- `-w=name`、`--worktree name`、`--worktree=name` の表記もすべて同じ意味として受け付けます。
- `--` の後ろの引数はそのまま Claude Code に渡され、この書き換えの影響を受けません。

{{< callout type="info" >}}
セッション引き継ぎで worktree 名を SPEC ID と同じにしておくと (`moai cc -w SPEC-XXX-001`)、次のセッションが 1 行で同じ作業ツリーに復帰できます。
{{< /callout >}}

## Team window spawn (`--spawn` フラグ)

3 ランチャー (`moai cc`、`moai cg`、`moai glm`) はすべて `--spawn` フラグを受け付けます。`-w` と組み合わせると、現在のプロセスを Claude Code で置き換える代わりに **新しい tmux ウィンドウ** にセッションを開き、呼び出し元のセッションはそのまま維持します。チームモードでチームメイトのウィンドウを開くときに使います。

```bash
moai cc -w feat-login --spawn        # Claude チームメイトウィンドウ (現在のセッション維持)
moai cg -w feat-auth --spawn         # GLM チームメイトウィンドウ (現在のセッション維持)
moai glm -w feat-auth --spawn        # GLM セッション (現在のセッション維持)
```

動作ルール:

- tmux セッション内部でのみ動作します。tmux の外で呼ぶと **フォールバックなしで拒否** します — 黙って in-place で置き換えると呼び出し元のセッションを上書きする結果になるためです。
- `tmux` と `moai` バイナリが両方必要です。どちらかが欠けていても拒否します。
- `--spawn` なしで `-w <name>` だけを使うと、従来どおり現在のプロセスを置き換えます。

`--spawn` は CLI ランチャーの決定を `AskUserQuestion` で問わないという HARD ルールに従います — すべての結果は観察可能な状態 (tmux セッションの有無、バイナリの有無) から決定され、終了コードと stderr で報告されます。

## 関連ドキュメント

- [CG モード (Claude + GLM)](/ja/multi-llm/cg-mode)
- [プロファイル管理](/ja/cli-reference/profile)
- [セキュリティノート](/ja/advanced/security-notes) — GLM 資格情報パスのセキュリティモデル
- [CLI 概要](/ja/getting-started/cli)
