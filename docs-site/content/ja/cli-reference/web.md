---
title: moai web コンソール
weight: 50
draft: false
---

`moai web` はブラウザベースの設定エディタ **MoAI Web Console** を起動します。ターミナルのプロファイルウィザード (`moai profile`) と同じ検証・保存ロジックを再利用し、プロファイル設定とプロジェクトの user / language / statusline セクションを Web UI で編集できます。

## 概要

```bash
moai web [OPTIONS]
```

コンソールは **ループバック (127.0.0.1) のみにバインド** します。外部データベース、認証、ネットワーク公開は一切ありません。デフォルトでは、対象ポートを古い moai インスタンスが占有している場合、それを終了して再バインドします。moai 以外の外部プロセスは決して終了せず、その場合はエラーを報告し `--port` の使用を提案します。

## フラグ

| フラグ | 説明 |
|--------|------|
| `--port <N>` | 127.0.0.1 にバインドする TCP ポート (デフォルト: `3041`) |
| `--no-open` | ブラウザを自動で開かない |
| `--no-reuse` | 古い moai インスタンスからポートを回収せず、ポート競合時に失敗する |

## 例

```bash
moai web                 # 127.0.0.1:3041 にバインドしてブラウザを開く
moai web --port 9000     # 別のポートにバインド
moai web --no-open       # ブラウザを開かずに起動
moai web --no-reuse      # ポート使用中なら回収せず失敗
```

## 編集対象

Web コンソールは次を編集します。

- **プロファイル設定** — モデル・言語・表示設定などプロファイルごとの設定
- **プロジェクト設定** — `.moai/config/sections/` の user / language / statusline セクション

保存時にはターミナルウィザードと同じバリデーションを通るため、どちらの経路を使っても結果が一貫します。

## コンソール画面

コンソールのインターフェース言語は、ヘッダー右側のセレクタで English · 한국어 · 日本語 · 中文 の中から選びます。以下の画面は English にした状態のため、この節では画面に表示される英語表記を括弧に併記します。

ヘッダーにはプロジェクト名と現在のプロファイル、主要な設定の要約 (`lang · model · effort · dev`) が横に並びます。その下にプロファイルバーが続き、プロファイルセレクタのすぐ隣に追加・名称変更・削除コントロールがあるため、プロファイルの全ライフサイクルが1行に収まります（独立したプロファイルカードはありません）。プロファイルバーの下に、ユーザー情報(Identity) · 言語(Language) · LLM · サードパーティ LLM(3rd Party LLM) · ワークフロー(Workflow) · Git・ワークツリー(Git & Worktree) · 監査(Audit) · エージェント(Agents) · レポート(Report) の 9 つのタブが並びます。変更した値は最下部の設定を保存(Save settings) ボタンで記録します。

![MoAI Web Console の初期画面。ヘッダーのプロジェクト名とプロファイル、プロファイルバー、9 つのタブ、ユーザー情報(Identity) タブの表示名(Display name) 入力欄、設定を保存(Save settings) ボタン](/images/profile/web-console-overview.png)

プロファイルバーではプロファイルを切り替え、名称変更し、削除(Delete)し、新しいプロファイル名(New profile name) を入力してプロファイルを作成(Create profile) で新規作成できます。別のプロファイルを選ぶとヘッダーのプロファイル表示も一緒に変わります。以下は `moai-cowork` プロファイルに切り替えたあとに言語(Language) タブを開いた状態です。

![moai-cowork プロファイルに切り替えたコンソールの言語(Language) タブ。会話言語(Conversation language)、コミットメッセージ言語(Commit message language)、コードコメント言語(Code comment language)、ドキュメント言語(Documentation language) の 4 項目](/images/profile/web-console-switch.png)

LLM タブでは権限モード(Permission mode) とモデル(Model)、推論強度(Effort level) を変更します。ターミナルウィザードの "Model Settings" ステップが扱う値と同じです。

![moai-adk プロファイルの LLM タブ。権限モード(Permission mode)、モデル(Model)、推論強度(Effort level) の 3 項目](/images/profile/web-console-llm-tab.png)

## プロファイル記録の範囲

コンソールでプロファイルを切り替えると、その選択が `~/.moai/claude-profiles/launch.yaml` に現在のプロジェクトの記録として残ります。同じプロジェクトで `-p` なしに `moai cc` を実行するとき、この値が使われます。

{{< callout type="note" >}}
プロジェクト単位の記録は次のリリースに含まれます。現在配布されているバージョンはプロジェクトを区別せず、グローバル記録 1 つだけを扱います。
{{< /callout >}}

コンソールが読む値も書く値もすべて現在のプロジェクトを基準とするため、画面に表示されるプロファイルと実際に記録されるプロファイルは常に同じです。ただし `moai cc -p X` で開始したセッションの中でコンソールを開くと、`CLAUDE_CONFIG_DIR` がすでに決まっているため、記録とは無関係に `X` をそのまま表示します。

選択順序と制約は [プロファイル管理](/ja/cli-reference/profile#プロファイルの自動選択) で詳しく扱います。

---

関連: [プロファイル管理](/ja/cli-reference/profile) · [CLI 概要](/ja/getting-started/cli)
