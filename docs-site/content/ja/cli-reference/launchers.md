---
title: moai cc / glm ランチャー
weight: 15
draft: false
---

`moai cc` と `moai glm` は選択したバックエンドで Claude Code を起動します。旧 CG 設定は起動前に移行が必要です。

## 対応ランチャーの比較

| ランチャー | バックエンド | 用途 |
|------------|--------------|------|
| `moai cc` | Claude 専用 | 標準実行 — すべてのエージェントが Claude モデルを使用 |
| `moai glm` | GLM 専用 | すべてのエージェントが Z.AI プロキシ経由で GLM モデルを使用 |

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
| `--chrome` / `--no-chrome` | Claude Code にそのまま渡します。ランチャーはどちらも自動では付けないため、`--no-chrome` を指定しない限り `/chrome` で接続できます |
| `-f, --factory` | **ファクトリーリーダー**として進入します。引数は取りません。リーダーは運用者が選んだカードをクロスセッションメッセージで空きレーンに丸ごと割り当て、レーンは `-l` で合流させます |
| `-l, --lane` | 稼働中のファクトリーに**レーン**として合流し、次の `lane-<n>` 番号（生きているレーンの最大番号の次）を自動で受け取ります。引数は取らず、稼働中のファクトリーがなければ拒否されます。`moai glm -l` と `moai codex -l` も同じ動作です |
| `--leader <name>` | `-l` または `--lane` とだけ併用します。合流先のリーダーセッションを指定します (既定値 `leader`、旧綴り `lead` は拒否)。実行の記録が欠けているか退役済みでも、生きているリーダーがいれば、合流はそのリーダーを検証 (pid + プロセス開始) してその実行を復元します |
| `--factory-run <run-id>` | `-l` と併用: 合流する実行を id で指定します。`--leader` とは併用できません |
| `--clear-policy <value>` | `moai cc -l` · `moai glm -l` と併用: カードを終えたあとコンテキストを空ける方法です (`clear-each` が既定、`clear-when-full`、`relaunch`) |
| `--no-auto-dispatch` | `moai cc -l` · `moai glm -l` と併用: レーンを手動モードで起動します。既定は、キューの次のカードを自分で借り受ける自己配車レーンです |

{{< callout type="info" >}} 進入トークンは `-f` (リーダー) と `-l` (レーン) の2つだけで、どちらも引数を取りません。1回の起動に付けられるトークンは1つなので、`-f` と `-l` を同時に付けるとエラーです。`-f <値>`、`-l lane-2` のように値を付ける形や、Codex でリーダーを求める `moai codex` の `-f` は、すべて1行のエラーで拒否されます。廃止された `-k` 進入も同じ扱いで、`-f` と `-l` を案内して拒否されます。詳しい契約は[ファクトリーモード](/ja/advanced/factory-mode)と [manager-lead リーダーコーディネーター](/ja/advanced/manager-lead)を参照してください。 {{< /callout >}}

カードは丸ごと1つのレーンに入り、そのレーンの中で `plan → run → sync` の3段階を順番に通ります。段階ごとにそのセッションが `Agent()` サブエージェントを起動し、書き込みを担当する起動は `isolation: "worktree"` で分離します。レーン1つが同時に動かせるサブエージェントは最大10個で、ランチャーがレーンのセッションに `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` としてこの値を設定します。そのため、N 個のレーンがマシンの容量を分け合う構造は、運用者の自制ではなく設定で保証されます。レーンは一斉に立ち上げないでください。最初のレーンを先に起動し、実際に出力が出始めたのを確かめてから残りを立ち上げます。

リーダーとレーンは別々のバックエンドで起動できます。バックエンドの組み合わせは、トークンの空きをまず見て決めます。判断が重い場所にだけ Opus を置き、実装中心のレーンは GLM で動かすのがひとつの出発点です。別の組み合わせでも、ひとつのバックエンドにそろえても、同じように問題ありません。

権限モードは `default`、`acceptEdits`(`moai init` の既定値)、`plan`、`auto`、`bypassPermissions`、`dontAsk` のいずれかです。`auto` モードではバックグラウンドの分類器が操作を審査します。対応するプランとモデルは [Claude Code の権限モードのドキュメント](https://code.claude.com/docs/en/permission-modes) を参照してください。

## moai glm — GLM バックエンド

```bash
moai glm setup <api-key>   # API キーを保存 (初回のみ)
moai glm --key <api-key>   # API キーを保存 (フラグ形式 — setup と同じ保存先)
moai glm                   # GLM バックエンドで起動
moai glm -p work           # 'work' プロファイルで起動
moai glm status            # 資格情報の状態を確認
```

`~/.moai/.env.glm` から GLM 資格情報を読み取り、`ANTHROPIC_AUTH_TOKEN`、`ANTHROPIC_BASE_URL` などの環境変数を注入したうえで Claude Code を起動します。

| サブコマンド | 説明 |
|--------------|------|
| `moai glm setup [api-key]` | GLM API キーを保存 |
| `moai glm --key <api-key>` | GLM API キーを保存 (フラグ形式 — 両形式とも同じファイルに記録し、後から保存した値が残ります) |
| `moai glm status` | 現在の GLM 資格情報の状態を表示 |

Jev (TypeSafe) 資格情報は `moai jev --key <credential>` で保存します — `~/.moai/.env.typesafe`(モード 0600)に記録され、`moai doctor` と Web コンソールが同じファイルを読み取ります。`--key` を付けずに `moai jev` を実行するとヘルプだけを表示します。

{{< callout type="warning" >}}
GLM は `auto` 権限モードに対応しません。このモードは利用条件を満たす Claude セッションで選んでください。廃止された CG は並列実行の代替手段ではありません。
{{< /callout >}}

## CG の廃止と設定の移行

Claude や GLM を起動せず、移行案内を表示して終了します。 `moai cc` の別名ではありません。 `llm.team_mode: cg` が残るプロジェクトでは、セッションを起動する前に移行先を明示的に選ぶ必要があります。 [CG の廃止と設定の移行](/ja/multi-llm/cg-mode/) CG は廃止されました。`moai migrate cg` で移行先を確認してください。

```bash
moai migrate cg
moai migrate cg --target claude-only --apply --accept-role-change
```

`llm.team_mode: claude`、`llm.gateway.teammate_mode: in-process`、`llm.gateway.teammate_provider: inherit` を保存します。従来の混合構成の役割分担を解除する変更です。Claude リーダーと GLM チームメイトのペインを維持する移行ではありません。

`claude-glm` は Claude リーダーと tmux 内の GLM チームメイトを表します。現在は TEAMMATE の統合検証を通過していないため、適用と起動は利用できず、プレビューのみ可能です。tmux のインストールや `verified: true` の設定では、この制限は解除されません。

## プロファイル (`-p` フラグ)

対応ランチャーとも `-p <name>` で名前付きプロファイルを指定すると `CLAUDE_CONFIG_DIR` が `~/.moai/claude-profiles/<name>/` に設定されます。複数のアカウント・設定セットを分離して運用するときに使用します。

## 隔離 worktree (`-w` フラグ)

対応ランチャーとも `-w [name]` で隔離された git worktree の中からセッションを開始できます。`cd` でディレクトリを移動してから起動する 2 ステップを 1 コマンドにまとめます。

```bash
moai cc -w feat-login    # .claude/worktrees/feat-login/ で開始
moai cc -w               # 名前を自動生成
moai glm -w feat-login   # GLM バックエンドでも同様
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

## 関連ドキュメント

- [CG の廃止と設定の移行](/ja/multi-llm/cg-mode/)
- [プロファイル管理](/ja/cli-reference/profile)
- [セキュリティノート](/ja/advanced/security-notes) — GLM 資格情報パスのセキュリティモデル
- [CLI 概要](/ja/getting-started/cli)
