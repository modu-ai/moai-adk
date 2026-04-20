---
title: /moai feedback
weight: 80
draft: false
---
# /moai feedback

MoAI-ADK へのフィードバックやバグレポートを送信するコマンド。

{{< callout type="info" >}}

**新しいコマンド形式**

`/moai:9-feedback` は `/moai feedback` に変更されました。

{{< /callout >}}

{{< callout type="info" >}}
**一言でいうと**: `/moai feedback` は MoAI-ADK 自体への改善提案やバグレポートについて**GitHub Issue を自動作成**するコマンドです。
{{< /callout >}}

{{< callout type="info" >}}
**スラッシュコマンド**: Claude Code で `/moai:feedback` と入力すると、このコマンドを直接実行できます。`/moai` だけ入力すると、利用可能なすべてのサブコマンドの一覧が表示されます。
{{< /callout >}}

## 概要

MoAI-ADK を使用中にバグを発見した場合、新機能が必要な場合、または改善アイデアがある場合にこのコマンドを使用します。GitHub に直接アクセスする必要はありません - Claude Code 内から直接フィードバックを送信できます。

{{< callout type="info" >}}
**重要**: このコマンドは**プロジェクトコードを変更するためのものではありません**。MoAI-ADK ツール自体についてのフィードバックを開発チームに伝えるためのものです。
{{< /callout >}}

## 使用方法

```bash
# 標準形式
> /moai feedback

# 短いエイリアス
> /moai fb
> /moai bug
> /moai issue
```

コマンドを実行すると、フィードバックタイプの選択と内容の入力をガイドされます。

## サポートされるフラグ

| フラグ | 説明 | 例 |
|------|-------------|---------|
| `--type {bug,feature,question}` | フィードバックタイプを直接指定 | `/moai feedback --type bug` |
| `--title "<title>"` | タイトルを直接指定 | `/moai feedback --title "エラーレポート"` |
| `--dry-run` | Issue 作成なしで内容のみ確認 | `/moai feedback --dry-run` |

## 仕組み

`/moai feedback` を実行すると、以下のプロセスが実行されます：

```mermaid
flowchart TD
    A["/moai feedback を実行"] --> B["フィードバックタイプを選択"]
    B --> C["内容を記述"]
    C --> D["環境情報を自動収集"]
    D --> E["GitHub Issue を自動作成"]
    E --> F["Issue URL を返す"]
```

### 自動収集される情報

フィードバック送信時、開発チームが問題を迅速に理解できるよう、以下の情報が自動的に含まれます。

| 収集項目 | 説明 | 例 |
|----------------|-------------|---------|
| MoAI-ADK バージョン | 現在インストールされているバージョン | v10.8.0 |
| OS 情報 | オペレーティングシステムとバージョン | macOS 15.2 |
| Claude Code バージョン | 使用中の Claude Code バージョン | 1.0.30 |
| 現在の SPEC | 作業中の SPEC ID | SPEC-AUTH-001 |
| エラーログ | 最近のエラー (ある場合) | TypeError: ... |

## フィードバックタイプ

### バグレポート

MoAI-ADK を使用中に遭遇したエラーや予期しない動作を報告します。

```bash
> /moai feedback
# タイプ選択: バグレポート
# タイトル: /moai run 実行時にキャラクタリゼーションテストが作成されない
# 説明: SPEC-AUTH-001 で /moai run を実行しましたが、キャラクタリゼーション
#        テストが PRESERVE フェーズで作成されず、直接 IMPROVE に進みました。
# 再現: /moai run SPEC-AUTH-001 を実行
```

### 機能リクエスト

MoAI-ADK に追加してほしい新機能を提案します。

```bash
> /moai feedback
# タイプ選択: 機能リクエスト
# タイトル: /moai loop で特定ファイルのみ対象にするオプションを追加
# 説明: /moai loop がプロジェクト全体ではなく、特定のディレクトリや
#        ファイルのみを対象にできると便利です。
# 例: /moai loop --path src/auth/
```

### 改善提案

既存機能を改善するアイデアを提案します。

```bash
> /moai feedback
# タイプ選択: 改善提案
# タイトル: /moai fix 実行結果で前後の diff を表示
# 説明: /moai fix が自動修正を diff 形式で表示すれば、
#        どのような変更が行われたかを一目で確認できます。
```

## エージェント委任チェーン

`/moai feedback` コマンドのエージェント委任フロー：

```mermaid
flowchart TD
    User["ユーザーリクエスト"] --> Orchestrator["MoAI オーケストレータ"]
    Orchestrator --> Collect["環境情報を収集"]

    Collect --> Info1["MoAI-ADK バージョン"]
    Collect --> Info2["OS 情報"]
    Collect --> Info3["Claude Code バージョン"]
    Collect --> Info4["現在の SPEC"]
    Collect --> Info5["エラーログ"]

    Info1 --> Format["Issue をフォーマット"]
    Info2 --> Format
    Info3 --> Format
    Info4 --> Format
    Info5 --> Format

    Format --> GitHub["manager-quality エージェント<br/>GitHub Issue 作成"]
    GitHub --> Complete["Issue URL を返す"]
```

**エージェントの役割:**

| エージェント | 役割 | 主なタスク |
|-------|------|------------|
| **MoAI オーケストレータ** | フィードバックプロセスをガイド |
| **manager-quality** | GitHub 統合 | Issue 作成、URL 返却 |

## 実践例

### 状況: コマンド実行中の予期しないエラー

```bash
# エラーが発生した状況
> /moai "決済機能を実装" --branch
# エラー: ブランチ作成失敗 - 権限が拒否されました

# フィードバックを送信
> /moai feedback
```

MoAI オーケストレータが順番にフィードバックタイプ、タイトル、説明を尋ねます。回答を入力すると、GitHub Issue が自動的に作成され、Issue URL が返されます。

```
GitHub Issue が作成されました:
https://github.com/anthropics/moai-adk/issues/1234

開発チームがレビュー後に回答します。
```

{{< callout type="info" >}}
**フィードバックはいつでも歓迎します!** 小さな不便でもフィードバックを送ってください。MoAI-ADK の改善に役立ちます。
{{< /callout >}}

## よくある質問

### Q: フィードバック内容を編集または削除できますか?

はい、GitHub で直接 Issue を編集またはクローズできます。Issue URL が提供されるため、いつでもアクセスできます。

### Q: 同じ問題を複数回報告できますか?

心配ありません - GitHub は重複した Issue をチェックします。問題が既に報告されている場合、既存の Issue に案内されます。

### Q: フィードバックへの回答はいつ受け取れますか?

開発チームは週次で Issue をレビューしてコメントします。複雑な問題は解決に時間がかかる場合があります。

### Q: `/moai feedback` と直接 GitHub Issue を作成の違いは何ですか?

`/moai feedback` は環境情報を自動収集するため、開発チームが問題をより迅速に理解できます。手動で Issue を作成するよりも効率的です。

## 関連ドキュメント

- [/moai - 完全自律自動化](/utility-commands/moai)
- [/moai loop - 反復修正ループ](/utility-commands/moai-loop)
- [/moai fix - ワンショット自動修正](/utility-commands/moai-fix)
