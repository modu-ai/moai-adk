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
| `-k, --kanban [SPEC-ID]` | カンバンリードとして進入 — `plan → run → sync` チェーンをこのセッションにシード。SPEC-ID を付けるとその SPEC を目標に |
| `-k --name <role>` | 開いているカンバンランに同伴セッションとして合流。ロールは `plan` · `run` · `sync`。同じロール名の生存セッションがあれば次の番号が付く (`plan-1`, `plan-2`, …) |
| `-f, --factory` | **ファクトリーリード**として進入 — ワーカー 1 人(`worker-1`)でファクトリーランを開きます。リードは運営者が選んだカードをセッション間メッセージで空きワーカーに配分 |
| `-f worker` | 次の空き番号にワーカーを1人自動合流させ、実行中のファクトリーのリードソケットに接続 |
| `-f worker-<n>` | その番号ちょうど(`worker-<n>`)のワーカーを1人追加で立ち上げ。番号が生存中の正規ワーカーと重なれば次の空き番号に、生存中のレガシーワーカー(`agent-<n>`/`lane-<n>`)と重なれば名前を挙げて拒否されます。`moai glm -f worker` / `-f worker-<n>` も GLM バックエンドで同じように動作 |
| `-k <N>` / `-k <N> --name worker-<i>` | v1.2.0 の統一形式で現在も有効 — `-k <N>` はワーカー N 人のランのリード、`-k <N> --name worker-<i>` はそのうちのワーカー `<i>`。N なしで `-k --name worker-<i>` だけなら既定 8 ワーカー |
| `-f agent` · `-f lane-<n>` · `--name lane-<n>` | **非推奨エイリアス。** それぞれ `-f worker` / `-f worker-<n>` / `--name worker-<n>` と全く同じに動作しますが、起動のたびに正規の綴りへの切り替えを促すヒントを表示します |

{{< callout type="info" >}} `-k` はカンバンチェーンのトークン、`-f` は**ファクトリーモード** (Factory Mode) 専用の進入トークンです。 `-k` ひとつが 3 つの形に解釈されるのはそのまま — 引数なし・SPEC-ID はカンバンリード、`--name <ロール>` はカンバン同伴、数値はワーカーランです。 1 回の実行に付けられる進入トークンはひとつだけなので、`-k` と `-f` を一緒に使うとエラーになります。 詳細な契約は[カンバンモード](/ja/advanced/kanban-mode)と[manager-lead リードコーディネーター](/ja/advanced/manager-lead)を参照してください。 {{< /callout >}}

カードの回り方はカンバンとは違います。カンバンでは1枚のカードが `plan → run → sync` の列を移っていきますが、ファクトリーでは1枚のカードが丸ごと1人のワーカーに入り、そのワーカーの中で3段階を順に通過します。段階ごとにそのセッションが `Agent()` サブエージェントを立ち上げ、書き込みを担うスポーンは `isolation: "worktree"` で隔離します。1人のワーカーが同時に立ち上げられるサブエージェントは最大10個で、ランチャーがワーカー・同伴セッションに `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` としてこの値を仕込むため、ワーカー N 人がマシンの容量を分け合う構造は運営者の自制ではなく設定で保証されます。ワーカーは一度に全員を立ち上げないでください — まず最初のワーカーを上げ、実際に出力が出始めたのを確認してから残りを立ち上げます。

バックエンドの組み合わせは、トークンの空きをまず見て決めます。ひとつの出発点は、リードは GLM、plan は Claude(Opus)、run は GLM、sync は Claude(Opus)と置き、判断の重い段階にだけ Opus を配置するやり方です。別の組み合わせを使うのも、片方のバックエンドに統一するのも同じように問題ありません。

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
