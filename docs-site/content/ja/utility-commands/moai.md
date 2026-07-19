---
title: /moai
weight: 20
draft: false
---

完全自律の自動化コマンドです。ユーザーが目標を提供すると MoAI が **plan → run → sync** パイプラインを自律的に実行します。

{{< callout type="info" >}}
  **一行要約**: `/moai` は「完全自律の自動化」コマンドです。ユーザーは望む
  機能を自然言語で説明するだけで、MoAI が SPEC 生成から実装、ドキュメント化まで **すべての
  プロセスを自動的に** 行います。
{{< /callout >}}

{{< callout type="info" >}}
**スラッシュコマンド対応**: MoAI のすべてのサブコマンドはスキルでラップされているので、`/moai` だけ入力すると利用可能なサブコマンド一覧が表示されます。各サブコマンドは `/moai:fix`、`/moai:loop`、`/moai:review` などの形式ですぐに実行することもできます。
{{< /callout >}}

## 概要

`/moai` は MoAI-ADK の **完全自律の自動化ワークフロー** コマンドです。下位コマンドを別々に実行する必要なく、たった 1 回のコマンドで開発プロセス全体が自動化されます:

1. **SPEC 生成** (manager-spec)
2. **DDD/TDD 実装** (manager-develop — quality.yaml の development_mode に応じて)
3. **ドキュメント同期** (manager-docs)

## Analyze-First ルーティング

v3 から `/moai` の基本ルーティングは **Analyze-First** — 言語独立的な意図分析です。英語のキーワードマッチングではなくリクエストの意味を分類するので、どの `conversation_language` でリクエストしても同じ品質でルーティングされます。

ルーティングは次の順序で進行します:

1. **意図分析**: ユーザーリクエストの意図を分類 (入力言語とは無関係)
2. **コンテキスト十分性の確認**: 不十分なら Socratic インタビューで明確化
3. **実行計画の構成**: スキル / エージェント / 動的ワークフローのチェーンを選択
4. **オーケストレーションモードの選択** (Phase 4): solo-sequential / parallel-subagents / dynamic-workflow

つまり `/moai "ログインバグを直して"` のようにサブコマンドなしで自然言語だけ入力しても、意図分析を経て適切なワークフロー (修正なら fix 系列、新機能なら plan→run→sync パイプライン) につながります。

## 使い方

```bash
# 基本的な使い方
> /moai "実装したい機能の説明"

# ワークツリーと一緒に
> /moai "機能の説明" --worktree

# ブランチと一緒に
> /moai "機能の説明" --branch

# ループモードの有効化
> /moai "機能の説明" --loop

# 既存 SPEC の再開
> /moai --resume SPEC-AUTH-001
```

## 対応フラグ

| フラグ              | 説明                             | 例                           |
| ------------------- | -------------------------------- | ------------------------------ |
| `--loop`            | 実装後に自動反復修正を有効化    | `/moai "機能" --loop`          |
| `--max N`           | ループ反復上限の指定 (デフォルト値は設定ベース、`ralph.yaml` `loop.max_iterations` = 10) | `/moai "機能" --loop --max 20` |
| `--sequential`      | Phase 1 の探索エージェントを並列の代わりに順次実行 | `/moai "機能" --sequential`    |
| `--branch`          | 自動 feature ブランチの生成         | `/moai "機能" --branch`        |
| `--pr`              | 完了後の自動 PR 生成             | `/moai "機能" --pr`            |
| `--issue`           | SPEC 生成 (plan ステップ) 後の GitHub Issue 生成の opt-in (なければ late-branch opt-in ポリシーに従いスキップ) | `/moai "機能" --issue` |
| `--resume SPEC-XXX` | 既存 SPEC 作業の再開              | `/moai --resume SPEC-AUTH-001` |
| `--solo`            | 下位エージェントモードを強制 (順次実行) | `/moai "機能" --solo`          |
| `--team`            | (引退) `MODE_TEAM_UNAVAILABLE` とともに下位エージェントモードにフォールバック | `/moai "機能" --team`          |

### --loop フラグ

実装が完了した後に自動的に反復修正を実行してすべてのエラーを修正します:

```bash
> /moai "JWT 認証システム" --loop
```

このオプションを使うと:

1. SPEC 生成
2. DDD 実装
3. **自動ループ実行** (LSP エラー、テスト失敗、カバレッジ不足の解決)
4. ドキュメント同期
5. PR 生成

{{< callout type="info" >}}
  `--loop` オプションは **実装後の整理作業を完全に自動化** して生産性を
  最大化します。
{{< /callout >}}

### --solo フラグとオーケストレーションモード

フラグなしで実行すると MoAI が作業規模を見てオーケストレーションモードを自動選択します:

**自動選択の基準** (フラグがないとき):

- 影響ドメイン >= 3 個 → 並列実行
- 修正ファイル >= 10 個 → 並列実行
- 複雑度スコア >= 7 → 並列実行
- それ以外 → 下位エージェントモード (順次実行)

| フラグ | 動作 |
| ------ | ---- |
| `--solo` | 下位エージェントモードを強制 (順次実行) |
| `--team` | (引退) `MODE_TEAM_UNAVAILABLE` とともに下位エージェントモードにフォールバック |
| (なし) | 複雑度ベースの自動選択 |

{{< callout type="warning" >}}
**v3.0.0 の変更**: Agent Teams 静的オーケストレーション階層は **引退** しました。`--team` を強制しても `MODE_TEAM_UNAVAILABLE` とともに下位エージェントモードにフォールバックします。並列実行は並列下位エージェントのファンアウトと動的ワークフロー 2 種 (plan-phase の研究並列ファンアウト、sync-phase の 4 次元品質評価) が担当し、ネイティブ teammate ランタイム (`moai cg` の tmux pane) はそのまま維持されます。
{{< /callout >}}

並列実行はエージェントごとに独立したコンテキストウィンドウを使うのでトークン使用量が増えます。単純な単一ドメイン作業には `--solo` (順次) の方が経済的です — 規模ベースの自動選択がデフォルト値である理由です。

## 実行プロセス

`/moai` が内部的に行う全体のプロセスです:

```mermaid
flowchart TD
    A["コマンド実行<br/>/moai '機能の説明'"] --> B{--resume?}
    B -->|はい| C["SPEC ロード<br/>続けて作業"]
    B -->|いいえ| D["Phase 0<br/>並列探索"]

    subgraph D["Phase 0: 並列探索 (15-30 秒)"]
        D1["Explore 下位エージェント<br/>コードベース分析"]
        D2["Research 下位エージェント<br/>外部ドキュメント調査"]
        D3["Quality 下位エージェント<br/>品質基準線の確認"]
    end

    D --> E{"単一ドメイン?"}
    E -->|はい| F["専門家エージェントに<br/>直接委任"]
    E -->|いいえ| G["Phase 1 継続"]

    C --> G["Phase 1<br/>SPEC 生成"]
    G --> H["manager-spec 呼び出し"]
    H --> I["GEARS 形式の SPEC 生成"]
    I --> J[".moai/specs/SPEC-XXX/spec.md"]

    J --> K["Phase 2<br/>DDD 実装"]

    K --> L["manager-develop 呼び出し<br/>DDD/TDD 循環 (quality.yaml に応じて)"]
    L --> M{"実装完了?"}
    M -->|いいえ| L
    M -->|はい| N{"--loop?"}

    N -->|はい| O["自動ループ実行"]
    O --> P["すべての問題を解決"]
    N -->|いいえ| P

    P --> Q["Phase 3<br/>ドキュメント同期"]

    Q --> R["manager-docs 呼び出し<br/>ドキュメント生成"]
    R --> S{"--pr?"}
    S -->|はい| T["PR 生成"]
    S -->|いいえ| U["完了シグナル"]
    T --> U
```

**核心ポイント:**

- **Phase 0 (並列探索)**: 3 つのエージェントが同時に実行され 2-3 倍の速度向上
- **単一ドメインルーティング**: 単純な作業は専門家エージェントに直接委任して SPEC をスキップ
- **完了シグナル**: 作業完了時に完了報告書に作業完了を明示

## Phase 別の詳細

### Phase 0: 並列探索 (オプション)

3 つのエージェントが **同時に** 実行されてプロジェクトの文脈を素早く把握します:

| エージェント     | 役割            | 作業                                     |
| ------------ | --------------- | ---------------------------------------- |
| **Explore**  | コードベース分析 | 関連ファイル、アーキテクチャパターン、既存実装の発見 |
| **Research** | 外部ドキュメント調査  | 公式ドキュメント、API ドキュメント、類似実装の例      |
| **Quality**  | 品質基準線     | テストカバレッジ、リント状態、技術的負債    |

**速度向上:** 並列実行で順次実行に比べ 2-3 倍速い (15-30 秒 vs 45-90 秒)

**単一ドメインルーティング:**

- 単一ドメイン作業 (例: 「SQL 最適化」): SPEC 生成なしで専門家エージェントに直接委任
- 多重ドメイン作業: ワークフロー全体を進行

### Phase 1: SPEC 生成

**manager-spec** 下位エージェントが GEARS 形式の SPEC ドキュメントを生成します:

- .moai/specs/SPEC-XXX/spec.md
- GEARS 形式の要件
- Given-When-Then 受け入れ基準
- conversation_language で書かれたコンテンツ

{{< callout type="info" >}}
**GEARS 形式** が現在の SPEC 要件の正式な形式です。以前のドキュメント・構成で見られる **EARS** はレガシー名称で、GEARS に置き換えられました。
{{< /callout >}}

### Phase 2: DDD/TDD 実装ループ

**manager-develop** 下位エージェントが SPEC を基に実装を行います:

- DDD 循環: ANALYZE-PRESERVE-IMPROVE (既存コードのリファクタリング)
- TDD 循環: RED-GREEN-REFACTOR (新機能開発)
- ドメインコンテキストの自動注入 (バックエンド、フロントエンド、セキュリティ、データベースなど)

**quality.yaml development_mode 設定:**

- `development_mode: ddd` → DDD 循環を使用 (既存コードの改善)
- `development_mode: tdd` → TDD 循環を使用 (新機能開発、デフォルト値)

**ループ動作 (--loop または loop.enabled が true のとき):**

```
問題が存在 AND 反復 < 最大値:
  1. 診断実行 (LSP エラー、テスト失敗、カバレッジ)
  2. manager-develop に修正を委任
  3. 修正結果の検証
  4. 完了条件の充足有無を確認
  5. 完了文の検出時にループを終了
```

### Phase 3: ドキュメント同期

**manager-docs** 下位エージェントが実装とドキュメントを同期します:

- API ドキュメントの生成
- README の更新
- CHANGELOG の追加
- 成功時に作業完了を明示

## TODO 管理

**[HARD] TodoWrite ツールが必須:** すべての作業追跡に TodoWrite の使用が必須

- 課題発見時: TodoWrite (pending 状態)
- 作業開始前: TodoWrite (in_progress 状態)
- 作業完了後: TodoWrite (completed 状態)
- TODO リストをテキストで出力するのは禁止

## 完了シグナル

すべてのワークフローステップが正常に完了すると、MoAI は完了報告書 (バナー/散文) に作業完了を明示して結果を明確にします。

## LLM モードルーティング

トークノミクスの核心的な装置です。llm.yaml 設定に応じてステップごとに Claude と GLM を自動ルーティングします — 戦略・計画は Claude が、大量の実装は低コストの GLM が担当するハイブリッドが可能です。

| モード          | Plan ステップ      | Run ステップ       |
| ------------- | -------------- | -------------- |
| `claude-only` | Claude         | Claude         |
| `hybrid`      | Claude         | GLM (worktree) |
| `glm-only`    | GLM (worktree) | GLM (worktree) |

## 実践例

### 例: JWT 認証システムの完全自動化

**ステップ 1: コマンド実行**

```bash
> /moai "JWT ベースのユーザー認証システム: 会員登録、ログイン、トークン更新" --worktree --loop --pr
```

**ステップ 2: Phase 0 - 並列探索**

```
[並列探索開始]
  Explore 下位エージェント: src/auth/ を分析中...
  Research 下位エージェント: JWT best practices を調査中...
  Quality 下位エージェント: テストカバレッジ 32% を確認...

[探索完了 - 23 秒]
  発見ファイル: 4 個
  推奨ライブラリ: PyJWT, bcrypt
  基準線: LSP 0 エラー、カバレッジ 32%
```

**ステップ 3: Phase 1 - SPEC 生成**

```
[manager-spec 呼び出し]
  SPEC ID: SPEC-AUTH-001
  要件: 5 個 (GEARS 形式)
  受け入れ基準: 3 個のシナリオ

  ユーザー承認: 完了
```

**ステップ 4: Phase 2 - DDD 実装**

```
[manager-spec]
  作業分解: 7 個のタスク
  戦略計画完了

[manager-develop]
  ANALYZE: コード構造の分析完了
  PRESERVE: 特性化テスト 12 個を作成
  IMPROVE: 7 個のタスクの実装完了

[sync-auditor]
  TRUST 5: すべての柱を通過
  カバレッジ: 89%
  状態: PASS
```

**ステップ 5: 自動ループ (--loop)**

```
[ループ開始 - 反復 1/100]
  診断: 型エラー 2 個を発見
  修正: manager-develop 下位エージェントに委任
  検証: すべてのエラーを解決

[ループ終了 - 1 回反復]
  完了条件を充足!
```

**ステップ 6: Phase 3 - ドキュメント同期**

```
[manager-docs]
  API ドキュメント: docs/api/auth.md を生成
  README: 使い方セクションを更新
  CHANGELOG: v1.1.0 項目を追加
  SPEC-AUTH-001: ACTIVE → COMPLETED
```

**ステップ 7: 完了**

```
[完了]
  SPEC: SPEC-AUTH-001
  コミット: 7 個
  テスト: 36/36 通過
  カバレッジ: 89%
  PR: #42 生成 (Draft → Ready)

<moai:COMPLETE />
```

## よくある質問

### Q: `/moai` と下位コマンドの違いは何ですか?

| コマンド       | 範囲        | 使用タイミング                    |
| ------------ | ----------- | ---------------------------- |
| `/moai`      | 全体の自動化 | 素早い完全自動化を望むとき     |
| `/moai plan` | SPEC 生成のみ | SPEC を先にレビューしたいとき |
| `/moai run`  | 実装のみ      | SPEC がすでにあるとき          |
| `/moai sync` | ドキュメント化のみ    | 実装後にドキュメントだけ更新するとき |

### Q: --loop フラグはいつ使うべきですか?

実装後に自動的にすべてのエラーを修正したいときに使います。特に大規模なリファクタリング後の整理作業に有用です。

### Q: 単一ドメインルーティングとは何ですか?

単一ドメイン作業 (例: 「SQL クエリの最適化」) は SPEC 生成なしでその領域の専門家エージェントに直接委任して時間を節約します。

### Q: 英語以外の言語でリクエストしてもいいですか?

はい。Analyze-First ルーティングは言語独立的な意図分析なので、韓国語・日本語・中国語などどの言語でリクエストしても同じように動作します。

## 関連ドキュメント

- [/moai plan](/workflow-commands/moai-plan) - SPEC 生成の詳細
- [/moai run](/workflow-commands/moai-run) - DDD 実装の詳細
- [/moai sync](/workflow-commands/moai-sync) - ドキュメント同期の詳細
- [/moai loop](/utility-commands/moai-loop) - 反復修正ループの詳細
- [/moai fix](/utility-commands/moai-fix) - 一回限りの自動修正の詳細
