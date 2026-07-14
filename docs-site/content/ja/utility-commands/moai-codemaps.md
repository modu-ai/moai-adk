---
title: /moai codemaps
weight: 50
draft: false
---

コードベースをスキャンして **アーキテクチャドキュメント** を自動生成するコマンドです。

{{< callout type="info" >}}
**一行要約**: `/moai codemaps` は「アーキテクチャ地図製作者」です。コードベースを分析してモジュールマップ、依存性グラフ、進入点カタログなど **構造ドキュメントを自動生成** します。
{{< /callout >}}

{{< callout type="info" >}}
**スラッシュコマンド**: Claude Code で `/moai:codemaps` と入力すると、このコマンドをすぐに実行できます。`/moai` だけ入力すると、利用可能なすべてのサブコマンド一覧が表示されます。
{{< /callout >}}

## 概要

新しいプロジェクトに参加したり大規模なコードベースを把握したりするとき、アーキテクチャを理解することが最も重要です。`/moai codemaps` はコードベースを自動的に分析してモジュールマップ、依存性グラフ、進入点カタログ、データフロードキュメントを生成します。

生成されたドキュメントは `.moai/project/codemaps/` ディレクトリに保存され、人間と AI エージェントの両方がコードベースを素早く理解できるよう助けます。ハーネスエンジニアリングの用語では **コンテキストマップ** — エージェントが毎セッションでアーキテクチャを再発見する代わりに常に参照できるファイルベースの地図です。反復探索コストをドキュメント生成 1 回で置き換える点でトークン節約効果も大きいです。

## 使い方

```bash
# コードベース全体のアーキテクチャドキュメント生成
> /moai codemaps

# 既存ドキュメントを無視して再生成
> /moai codemaps --force

# 特定の領域のみ分析
> /moai codemaps --area api

# Mermaid ダイアグラム含む
> /moai codemaps --format mermaid

# 探索の深さ制限
> /moai codemaps --depth 3
```

## 対応フラグ

| フラグ | 説明 | 例 |
|-------|------|------|
| `--force` (または `--regenerate`) | 既存ドキュメントを無視してすべてのコードマップを再生成 | `/moai codemaps --force` |
| `--area AREA` | 特定の領域に集中して分析 | `/moai codemaps --area auth` |
| `--format FORMAT` | 出力形式 (markdown, mermaid, json, デフォルト値: markdown) | `/moai codemaps --format mermaid` |
| `--depth N` | 最大ディレクトリ探索の深さ (デフォルト値: 4) | `/moai codemaps --depth 3` |

### --force フラグ

既存のコードマップドキュメントをすべて削除して最初から作り直します:

```bash
> /moai codemaps --force
```

コードベースに大きな変化があったときに有用です。

### --area フラグ

特定の領域とその依存性のみ分析します:

```bash
# API モジュールのみ分析
> /moai codemaps --area api

# 認証モジュールのみ分析
> /moai codemaps --area auth
```

結果は `.moai/project/codemaps/{area}/` に保存されます。

### --format フラグ

出力形式を指定します:

```bash
# Mermaid ダイアグラム含む
> /moai codemaps --format mermaid

# JSON 形式を追加生成
> /moai codemaps --format json
```

## 実行プロセス

`/moai codemaps` は 5 ステップで実行されます。

```mermaid
flowchart TD
    Start["/moai codemaps 実行"] --> Phase1["ステップ 1: コードベース探索"]
    Phase1 --> Explore["Explore エージェント"]

    Explore --> Phase2["ステップ 2: アーキテクチャ分析"]
    Phase2 --> Analyze["モジュール分類<br/>依存性マッピング<br/>循環参照検出"]

    Analyze --> Phase3["ステップ 3: マップ生成"]
    Phase3 --> Generate["overview.md<br/>modules.md<br/>dependencies.md<br/>entry-points.md<br/>data-flow.md"]

    Generate --> Phase4["ステップ 4: 検証"]
    Phase4 --> Verify["ファイル存在確認<br/>依存性一貫性検査<br/>進入点アクセス性確認"]

    Verify --> Phase5["ステップ 5: 報告書"]
```

### ステップ 1: コードベース探索

`Explore` エージェントがコードベースを深く探索します:

| 探索対象 | 説明 |
|-----------|------|
| ディレクトリ構造 | トップレベルおよび重要なサブディレクトリのマッピング |
| モジュール境界 | パッケージ/モジュールの境界と責任の識別 |
| 進入点 | メイン進入点の探索 (main.go, index.ts, app.py など) |
| 公開 API | エクスポートされた関数、型、インターフェースの一覧 |
| 依存性グラフ | モジュール間の依存性マッピング (import, require) |
| 外部依存性 | サードパーティ依存性のカタログ |
| 設定ファイル | ビルド、デプロイ、設定ファイルの識別 |

### ステップ 2: アーキテクチャ分析

オーケストレーターが探索結果と決定論的ツール (例: `go list -deps -json` + `go doc`、またはプロジェクト言語の等価な依存性・ドキュメント抽出器) を基に **直接** 分析します (別途エージェントスポーンなし):

- レイヤー別モジュール分類 (プレゼンテーション、ビジネス、データ、インフラ)
- 高い fan-in モジュールの識別 (`@MX:ANCHOR` 候補)
- 循環依存性の検出
- リクエスト/データフロー経路のマッピング
- ドメイン境界の識別
- アーキテクチャパターンの認識 (MVC, Clean, Hexagonal など)

### ステップ 3: マップ生成

`.moai/project/codemaps/` ディレクトリに 5 種類のドキュメントを生成します:

| ファイル | 内容 |
|------|------|
| `overview.md` | 高レベルのアーキテクチャ要約およびモジュール説明 |
| `modules.md` | 詳細なモジュールカタログ (責任、依存性) |
| `dependencies.md` | 依存性グラフ (テキストおよび Mermaid ダイアグラム) |
| `entry-points.md` | 進入点カタログおよび呼び出し経路 |
| `data-flow.md` | 主要なデータフロー経路 |

`--area` フラグ使用時:
- `.moai/project/codemaps/{area}/overview.md`
- `.moai/project/codemaps/{area}/modules.md`
- `.moai/project/codemaps/{area}/dependencies.md`

### ステップ 4: 検証

- 参照されたすべてのファイルとモジュールの実際の存在有無を確認
- 依存性関係の双方向の一貫性を検査
- 進入点のアクセス可能性を検証
- 既存のコードマップとの変更点を比較 (`--force` でない場合)

生成された地図が実際のコードと一致するかを機械的に確認するステップです — ドキュメントも「生成した」ではなく検証を経てこそ完了と判定されます。

### ステップ 5: 報告書

```
## コードマップ生成報告書

### 生成されたファイル
- .moai/project/codemaps/overview.md
- .moai/project/codemaps/modules.md
- .moai/project/codemaps/dependencies.md
- .moai/project/codemaps/entry-points.md
- .moai/project/codemaps/data-flow.md

### アーキテクチャのハイライト
- パターン: Clean Architecture
- モジュール数: 12 個
- 進入点: 3 個 (API サーバー、CLI、ワーカー)

### 潜在的な問題
- 循環依存性: pkg/auth <-> pkg/user
- 高い結合度: pkg/core (fan_in: 8)
- 孤立したモジュール: pkg/legacy (使用場所なし)
```

## エージェント委任チェーン

`/moai codemaps` の唯一のエージェントスポーンはステップ 1 の `Explore` (読み取り専用) です。ステップ 2・3 の分析とドキュメント生成、ステップ 4 の検証はすべてオーケストレーターが直接行います。

```mermaid
flowchart TD
    User["ユーザーリクエスト"] --> MoAI["MoAI オーケストレーター"]
    MoAI --> Phase1["ステップ 1: 探索"]
    Phase1 --> Explore["Explore エージェント<br/>(読み取り専用、唯一のスポーン)"]

    Explore --> Phase23["ステップ 2-3: 分析および生成<br/>オーケストレーター直接"]

    Phase23 --> Phase4["ステップ 4: 検証<br/>オーケストレーター直接"]

    Phase4 --> Report["ステップ 5: 報告書"]
```

**エージェントの役割:**

| エージェント | 役割 | 主な作業 |
|----------|------|----------|
| **Explore** | コードベース探索 (読み取り専用) — 唯一の Agent() スポーン | ディレクトリ構造、モジュール境界、依存性マッピング |
| **MoAI オーケストレーター** | 分析・生成・検証・報告書 (すべて直接) | 探索結果 + 決定論的ツールでモジュール分類・依存性分析・コードマップファイル作成、検証、ユーザー相互作用 |

## よくある質問

### Q: コードマップはどのくらいの頻度で再生成すべきですか?

大規模なリファクタリングや新しいモジュール追加の後に再生成するのが良いです。`/moai sync` を実行するとコードマップも自動的に更新されます。

### Q: --area フラグで生成したコードマップは全体のコードマップと衝突しますか?

いいえ。`--area` で生成したコードマップは別のサブディレクトリに保存されます。全体のコードマップと独立して管理されます。

### Q: 生成されたコードマップを直接修正してもいいですか?

はい、手動で修正できます。ただし `--force` フラグで再生成すると手動修正が上書きされます。`--force` なしで実行すると既存ドキュメントを参考にして増分更新します。

### Q: どのアーキテクチャパターンを認識しますか?

MVC, Clean Architecture, Hexagonal, Layered Architecture など主要なパターンを認識します。認識されたパターンは `overview.md` に記録されます。

## 関連ドキュメント

- [/moai clean - デッドコード除去](/utility-commands/moai-clean)
- [/moai feedback - フィードバック提出](/utility-commands/moai-feedback)
