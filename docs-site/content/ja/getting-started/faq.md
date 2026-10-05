---
title: よくある質問
weight: 100
draft: false
---

MoAI-ADK 使用中によくある質問と回答です。


---

## Q: `moai` と `/moai` は何が違いますか?

まったく別の 2 つです。最もよくある混同なので先に押さえます。

| | `moai` (ターミナル CLI) | `/moai` (スラッシュサブコマンド) |
|---|---|---|
| **実行場所** | ターミナルシェル | Claude Code の対話画面 |
| **正体** | Go バイナリ | Claude Code スキル呼び出し |
| **用途** | プロジェクト設定、テンプレート配布 | AI エージェント開発ワークフロー |
| **例** | `moai init my-project` | `/moai plan "認証機能"` |

- ターミナルで `moai plan` を実行しても動作しません — `/moai plan` は Claude Code の中でのみ有効です。
- Claude Code で `/moai init` と入力しても動作しません — `moai init` はターミナルコマンドです。

---

## Q: statusline のバージョン表示は何を意味しますか?

MoAI statusline はバージョン情報とアップデート通知を併せて表示します:

```text
🗿 v3.1.2 -> 🗿 v3.1.3
```

- **`🗿 v3.1.2`**: 現在インストールされているバージョン
- **`-> 🗿 v3.1.3`**: アップデート可能な新しいバージョン (ASCII 矢印 `->` でつながれます)

最新バージョンを使用中のときはバージョン番号だけが表示されます:

```text
🗿 v3.1.3
```

**アップデート方法**: `moai update` 実行時にアップデート通知が消えます。

{{< callout type="info" >}}
**参考**: Claude Code のビルトインバージョン表示 (`🔅 v2.1.38`) とは異なります。MoAI の表示は MoAI-ADK のバージョンを追跡し、Claude Code は自身のバージョンを別途表示します。
{{< /callout >}}

---

## Q: statusline に表示されるセグメントをカスタマイズするには?

statusline はセグメント単位でオン・オフします。各セグメントを個別にトグルして、欲しい情報だけを残してください。ディスプレイプリセットはありません — 設定はテーマとセグメントの2つだけです。

`moai init` または `moai update -c` ウィザードで設定するか、`.moai/config/sections/statusline.yaml` を直接編集します:

```yaml
statusline:
  segments:
    model: true
    context: true
    output_style: false
    directory: false
    git_status: true
    claude_version: false
    moai_version: false
    git_branch: true
```

`segments:` ブロックが無ければ、すべてのセグメントがデフォルトで有効になります。

{{< callout type="info" >}}
詳しい内容は [SPEC-STATUSLINE-001](https://github.com/modu-ai/moai-adk/blob/main/.moai/specs/SPEC-STATUSLINE-001/spec.md) を参照してください。
{{< /callout >}}

---

## Q: モデルポリシーをどう選びますか?

v3.2 から、エージェントごとのモデルポリシーにはもう選ぶものがありません。**サブエージェントはメインセッションのモデルと推論深度をそのまま引き継ぎます** — spawn するとき `model` も `effort` も渡さず、MoAI のエージェント定義はどちらも宣言しません。残ったのはセッションレベルの選択 1 つです。`moai profile setup` の**セッションモデルポリシー**は、このプロファイルで起動する Claude セッションの既定の推論強度(推論強度を別に選ばなかったときに適用)を決めます。

### セッションモデルポリシーの比較

| 値 | 意味 |
|------|------|
| **high** | セッション effort フォールバック `high` |
| **medium** (デフォルト値) | セッション effort フォールバック `medium` — コスト/スコア曲線の膝 |
| **low** | セッション effort フォールバック `low` — 同じモデルの中での経済運用 |

{{< callout type="warning" >}}
**なぜ重要ですか?** effort を下げて変わるのは、モデルクラスではなく主に *推論深度* です。長期にわたるエージェンティックな作業では、Opus の `low` effort が `max` を含むあらゆる effort の Sonnet よりもスコアが高く、作業あたりのコストも低くなります — 請求額を決めるのはトークン単価ではなく、モデルが完了までに費やしたステップ数です。この経済運用の座が今はセッション effort にあり、かつてのエージェント別割り当て表は退きました。
{{< /callout >}}

### エージェント別割り当ての時代から変わったこと

v3.1 まで MoAI-ADK はプロファイルマトリクスで 13 個のカタログエージェントそれぞれに `{model, effort}` を割り当て、ティアがその列を選んでいました。この仕組みは SPEC-AGENT-MODEL-INHERIT-001 で退きました — 呼び出しに model 引数が付いたケースが 1% にも満たないという実測が出て、割り当ての席はセッション自身へ移ったからです。かつての `--model-policy`、`--profile`、`--high`、`--medium-alias`、`--low` フラグは警告だけを出して効果のない廃止済みスタブとして残っています。

### 設定方法

```bash
# セッションモデルポリシーの設定 (セッションモデルポリシーの質問)
moai profile setup

# 既存プロジェクトの再設定
moai update -c                # 設定ウィザードの再実行
```

{{< callout type="info" >}}
既定の effort フォールバックは `medium` です。`moai profile setup` で変えるか、`/effort` や `ultrathink` でセッション effort をその都度調整してください — 以降のサブエージェント呼び出しはすべてその値に従います。
{{< /callout >}}

---

## Q: "Allow external CLAUDE.md file imports?" 警告が表示されます

プロジェクトを開くとき、Claude Code が外部ファイル import に関するセキュリティプロンプトを表示することがあります:

```
External imports:
  /Users/<user>/.moai/config/sections/quality.yaml
  /Users/<user>/.moai/config/sections/user.yaml
  /Users/<user>/.moai/config/sections/language.yaml
```

{{< callout type="info" >}}
**推奨する対応:** **"No, disable external imports"** を選択してください。
{{< /callout >}}

**理由:**
- プロジェクトの `.moai/config/sections/` にすでにこれらのファイルが存在します
- プロジェクト別設定がグローバル設定より優先適用されます
- 必須設定はすでに CLAUDE.md テキストに含まれています
- 外部 import を無効化する方が安全で、機能に影響を与えません

**ファイルの説明:**
- `quality.yaml`: TRUST 5 フレームワークおよび開発方法論の設定
- `language.yaml`: 言語設定 (会話、コメント、コミット)
- `user.yaml`: ユーザー名 (オプション、Co-Authored-By 表示用)

---

## Q: TDD と DDD 方法論の違いは何ですか?

MoAI-ADK v2.5.0+ では、方法論を **TDD または DDD のどちらか** からのみ選べます。明確性と一貫性のため hybrid モードは除去されました。

TDD はテストを先に書いてそのテストを通す順序なので新規開発に向いており、DDD は既存の動作を特性テストで押さえたうえで少しずつ手を入れる順序なので、テストがほとんどないコードに向いています。各サイクルのステップごとの手順は [SPEC ベース開発](/ja/core-concepts/spec-based-dev) と [DDD](/ja/core-concepts/ddd) で扱います。

### 方法論選択表

| プロジェクト状態 | テストカバレッジ | 推奨方法論 | 理由 |
|--------------|---------------|-------------|------|
| 新規プロジェクト | N/A | TDD | テスト優先開発 |
| 既存プロジェクト | 50%+ | TDD | テスト基盤がある |
| 既存プロジェクト | 10-49% | TDD | テスト拡張が可能 |
| 既存プロジェクト | < 10% | DDD | 段階的な特性テストが必要 |

### 設定方法

```bash
# プロジェクト初期化時に自動検出
moai init my-project          # --mode <ddd|tdd> フラグで指定可能

# 手動設定
# .moai/config/sections/quality.yaml を編集
development_mode: tdd         # または ddd
```


---

## Q: 自分のコードに @MX タグがない理由は?

これは **完全に正常** です。@MX タグシステムは AI が最初に注目すべき最も危険で重要なコードだけを表示するよう設計されています。

| 質問 | 回答 |
|------|------|
| タグがなければ問題ですか? | **いいえ。** ほとんどのコードにはタグが必要ありません。 |
| タグはいつ追加されますか? | **高い fan_in** (呼び出し元 >= 3)、**複雑なロジック** (複雑度 >= 15)、**危険なパターン** (context のない goroutine) にのみ追加されます。 |
| すべてのプロジェクトが似ていますか? | **はい。** すべてのプロジェクトでほとんどのコードにはタグがありません。 |

### タグの優先順位

| 優先順位 | 条件 | タグの種類 |
|---------|------|----------|
| **P1 (致命的)** | fan_in >= 3 | `@MX:ANCHOR` |
| **P2 (危険)** | goroutine、複雑度 >= 15 | `@MX:WARN` |
| **P3 (コンテキスト)** | マジック定数、godoc なし | `@MX:NOTE` |
| **P4 (欠落)** | テストファイルなし | `@MX:TODO` |

コードベースを @MX タグでスキャンするには:

```bash
/moai mx --all        # 全体スキャン
/moai mx --dry        # プレビュー
/moai mx --priority P1  # 致命的な項目のみ
```

---

## さらに質問がありますか?

- [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions) — 質問、アイデア、フィードバック
- [Issues](https://github.com/modu-ai/moai-adk/issues) — バグレポート、機能リクエスト
- [Discord コミュニティ](https://discord.gg/Z7E7Mdc5aN) — リアルタイムのやり取り、ヒント共有
