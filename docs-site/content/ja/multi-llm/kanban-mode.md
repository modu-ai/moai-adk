---
title: カンバンモード
weight: 30
draft: false
---

{{< callout type="info" >}}
カンバンモードの全体概要と Origin-Trail Chain 設計方向は[カンバンモード](/ja/advanced/kanban-mode)をご覧ください。このページはマルチセッション（リード + コンパニオン）運用手順を扱います。
{{< /callout >}}

## カンバンモードとは?

カンバンモードは1つの**リード**セッションが `plan -> run -> verify -> sync` チェーンを主導し、4つの**コンパニオン**セッションが同じランに合流して作業を並列に分散させます。ランのすべてのセッション（リードとコンパニオン双方）は引き上げられた Stop-フックブロック上限を受け、セッション途中で設定したゴールがデフォルトの連続ブロック上限を越えて実行され続けます。

リードがチェーンをシードし、コンパニオンはそうしません。各コンパニオンはカンバンメンバーシップフラグ（`-k`）とロールラベル（`--name <role>-<run-id>`）を一緒に渡すため、ディスパッチャーが正しく分類し SessionStart フックがメンバーシップを案内します。

## 進入スイッチ

### リード進入

```bash
moai cc -k                     # Claude バックエンドリード
moai cc -k SPEC-AUTH-001       # SPEC に束ねられたリード
moai glm -k                    # GLM バックエンドリード
```

リードセッションは:

{{< icon check-circle ok >}} `MOAI_KANBAN` + `MOAI_KANBAN_ID` を設定（チェーンシード）。
{{< icon check-circle ok >}} SessionStart でラン id と4つのコンパニオン実行コマンドを出力。
{{< icon x-circle danger >}} `MOAI_KANBAN_LABEL` は設定しない（コンパニオンのシグナル）。

### コンパニオン進入

```bash
moai cc -k --name plan-abc123    # plan コンパニオン
moai cc -k --name run-abc123     # run コンパニオン
moai cc -k --name review-abc123  # review コンパニオン
moai cc -k --name sync-abc123    # sync コンパニオン
moai glm -k --name run-abc123    # GLM バックエンドで同様
```

`<run-id>` はリードが開始時に案内した識別子で、4つのロールは
`plan`, `run`, `review`, `sync` です。

コンパニオンセッションは:

{{< icon check-circle ok >}} `MOAI_KANBAN_LABEL` を設定（メンバーシップ + ロールラベル）。
{{< icon check-circle ok >}} リードと同じ引き上げられた Stop-フックブロック上限。
{{< icon x-circle danger >}} `MOAI_KANBAN` は設定しない — チェーンをシードしません。

### 無処理（変更のないセッション）

```bash
moai cc --name mysession         # -k なし、カンバンメンバーシップなし
moai cc --name run-abc123        # コンパニオン形式だが -k なし → 無処理
```

`-k` がなければ `--name` 形式にかかわらずディスパッチャーは何も行い
ません。`--name` フラグはそのまま Claude に渡されます。

## マルチセッションブートストラップフロー

```
端末 1 (リード)            端末 2-5 (コンパニオン)
─────────────────          ────────────────────────
moai cc -k                 moai cc -k --name plan-<run-id>
                           moai cc -k --name run-<run-id>
                           moai cc -k --name review-<run-id>
                           moai cc -k --name sync-<run-id>
```

ブートストラップは手動です: セッションは別のセッションを起動できません。リード
SessionStart 通知がコピーすべき4つのコマンドを正確に出力します。各
コンパニオンを GLM バックエンドで実行するには `moai cc` を `moai glm` に
変えればよいです。

## クロスセッションメッセージング

セッション間通信は Claude Code のクロスセッションメッセージング（`ListAgents` /
`SendMessage`）を使います。`crossSessionInbound` 設定フィールドがインバウンド
メッセージを受諾するか、保留するか、拒否するかを制御します。

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

リード通知はラン id、4つのコンパニオン実行コマンド、リーダーソケットパス、
インバウンド自動化状態を案内します。コンパニオン通知は合流を確認する
役割のない1行です。両通知ともプロンプトせず情報提供用の stdout です。
