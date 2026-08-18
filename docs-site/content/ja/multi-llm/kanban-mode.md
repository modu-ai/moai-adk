---
title: カンバンモード
weight: 30
draft: false
---

{{< callout type="info" >}}
カンバンモードの全体概要と Origin-Trail Chain 設計方向は[カンバンモード](/ja/advanced/kanban-mode)をご覧ください。このページはマルチセッション（リード + コンパニオン）運用手順を扱います。
{{< /callout >}}

## カンバンモードとは?

カンバンモードは1つの**リード**セッションが `plan -> run -> sync` チェーンを主導し、3つの**コンパニオン**セッションが同じランに合流して作業を並列に分散させます。レビュー判定は独立した段階ではなく、sync ゲートが吸収します。ランのすべてのセッション（リードとコンパニオン双方）は引き上げられた Stop-フックブロック上限を受け、セッション途中で設定したゴールがデフォルトの連続ブロック上限を越えて実行され続けます。

リードがチェーンをシードし、コンパニオンはそうしません。各コンパニオンはカンバンメンバーシップフラグ（`-k`）とロールラベル（`--name <role>`）を一緒に渡すため、ディスパッチャーが正しく分類し SessionStart フックがメンバーシップを案内します。

## 進入スイッチ

### リード進入

```bash
moai cc -k                     # Claude バックエンドリード
moai cc -k SPEC-AUTH-001       # SPEC に束ねられたリード
moai glm -k                    # GLM バックエンドリード
```

リードセッションは:

{{< icon check-circle ok >}} `MOAI_KANBAN` + `MOAI_KANBAN_ID` を設定（チェーンシード）。
{{< icon check-circle ok >}} SessionStart でラン id と3つのコンパニオン実行コマンドを出力。
{{< icon x-circle danger >}} `MOAI_KANBAN_LABEL` は設定しない（コンパニオンのシグナル）。

### コンパニオン進入

```bash
moai cc -k --name plan    # plan コンパニオン
moai cc -k --name run     # run コンパニオン
moai cc -k --name sync    # sync コンパニオン
moai glm -k --name run    # GLM バックエンドで同様
```

コンパニオンの名前はロール名だけで付き、3つのロールは
`plan`, `run`, `sync` です。`<run-id>` はリードセッションだけの識別子で、コンパニオン名には入りません — 同じロール名がすでに生きていれば次の番号が付きます。

コンパニオンセッションは:

{{< icon check-circle ok >}} `MOAI_KANBAN_LABEL` を設定（メンバーシップ + ロールラベル）。
{{< icon check-circle ok >}} リードと同じ引き上げられた Stop-フックブロック上限。
{{< icon x-circle danger >}} `MOAI_KANBAN` は設定しない — チェーンをシードしません。

### 無処理（変更のないセッション）

```bash
moai cc --name mysession         # -k なし、カンバンメンバーシップなし
moai cc --name run               # コンパニオンのロール名だが -k なし → 無処理
```

`-k` がなければ `--name` 形式にかかわらずディスパッチャーは何も行い
ません。`--name` フラグはそのまま Claude に渡されます。

## マルチセッションブートストラップフロー

```
端末 1 (リード)            端末 2-4 (コンパニオン)
─────────────────          ────────────────────────
moai cc -k                 moai cc -k --name plan
                           moai cc -k --name run
                           moai cc -k --name sync
```

ブートストラップは手動です: セッションは別のセッションを起動できません。リード
SessionStart 通知がコピーすべき3つのコマンドを正確に出力します。各
コンパニオンを GLM バックエンドで実行するには `moai cc` を `moai glm` に
変えればよいです。

## クロスセッションメッセージング

セッション間通信は Claude Code のクロスセッションメッセージング（`ListAgents` /
`SendMessage`）を使います。`crossSessionInbound` 設定フィールドがインバウンド
メッセージを受諾するか、保留するか、拒否するかを制御します。

### 可用性の制約

クロスセッションメッセージングはすべての環境で使える機能ではありません。カンバンモードはリードとコンパニオンがこのチャネルだけで接続されるため、チャネルがなければモード自体が成立しません。開始前に以下の制約を確認してください。

{{< icon warning warn >}} **OS**: macOS と Linux（WSL 2 内の Linux を含む）でのみ使用できます。ネイティブ Windows では Claude Code はクロスセッションメッセージングを提供しません。
{{< icon warning warn >}} **プロバイダー**: Amazon Bedrock、Claude Platform on AWS、Agent Platform on Google Cloud、Microsoft Foundry では使用できません。
{{< icon warning warn >}} **バージョン**: Claude Code v2.1.224 以降が必要です。マシン間会話を先に開始するには v2.1.225 以降、@メンションと /config 行の表示には v2.1.232 以降です。
{{< icon warning warn >}} **フラグ**: `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`、`DISABLE_TELEMETRY`、`DO_NOT_TRACK`、`DISABLE_GROWTHBOOK` のいずれか一つでも機能フラグ評価を無効にすると、メッセージングは警告なくオフになります。

簡単な診断: `/list-agents` コマンドが認識されればこの機能があり、認識されなければありません。

カンバンモードはインバウンドメッセージを自動受諾します: ランチャーが
`{"crossSessionInbound": "accept"}` を含む一時設定ファイルを書き
`--settings` でバックエンドに渡します。ファイルはセッション専用で（終了時
整理）恒久設定を変更しません。

### 運用者提供 `--settings`

コマンドラインに `--settings <file>` を渡せばランチャーは自身の設定ファイルを
注入しません。ファイルに以下の内容があることを確認してください:

```json
{
  "crossSessionInbound": "accept"
}
```

リード SessionStart 通知はランチャーが注入していないとき確認を
促す案内を出力します。

## SessionStart 通知

リード通知はラン id、3つのコンパニオン実行コマンド、リーダーソケットパス、
インバウンド自動化状態を案内します。コンパニオン通知は合流を確認する
役割のない1行です。両通知ともプロンプトせず情報提供用の stdout です。
