---
title: ファクトリーモード
weight: 30
draft: false
---

## ファクトリーモードとは

ファクトリーモードは、1つの**リード**セッションが `plan -> run -> verify ->
sync` チェーンを主導し、4つの**コンパニオン**セッションが同じランに参加して
作業を並列化します。ラン内のすべてのセッション（リード・コンパニオン問わず）
は、引き上げられた Stop-hook ブロック上限を受け取り、セッション途中で設定した
ゴールがデフォルトの連続ブロック上限を超えて実行され続けます。

リードがチェーンをシードし、コンパニオンはシードしません。各コンパニオンは
ファクトリーメンバーシップフラグ（`-f`）とロールラベル（`--name
<role>-<run-id>`）を一緒に渡すので、ディスパッチャが正しく分類し、
SessionStartフックがメンバーシップを案内します。

## エントリスイッチ

### リードエントリ

```bash
moai cc -f                     # Claude バックエンドのリード
moai cc -f SPEC-AUTH-001       # SPEC に紐づくリード
moai glm -f                    # GLM バックエンドのリード
```

リードセッションは:

{{< icon check-circle ok >}} `MOAI_FACTORY` + `MOAI_FACTORY_ID` を設定（チェーンシード）。
{{< icon check-circle ok >}} SessionStart でラン id と4つのコンパニオン起動コマンドを出力。
{{< icon x-circle danger >}} `MOAI_FACTORY_LABEL` は設定しない（コンパニオンのシグナル）。

### コンパニオンエントリ

```bash
moai cc -f --name plan-abc123    # plan コンパニオン
moai cc -f --name run-abc123     # run コンパニオン
moai cc -f --name review-abc123  # review コンパニオン
moai cc -f --name sync-abc123    # sync コンパニオン
moai glm -f --name run-abc123    # GLM バックエンドで同じ
```

`<run-id>` はリードが起動時に案内した識別子で、4つのロールは
`plan`, `run`, `review`, `sync` です。

コンパニオンセッションは:

{{< icon check-circle ok >}} `MOAI_FACTORY_LABEL` を設定（メンバーシップ + ロールラベル）。
{{< icon check-circle ok >}} リードと同じ引き上げられた Stop-hook ブロック上限。
{{< icon x-circle danger >}} `MOAI_FACTORY` は設定しない — チェーンをシードしません。

### 無処理（変更なしセッション）

```bash
moai cc --name mysession         # -f なし、ファクトリーメンバーシップなし
moai cc --name run-abc123        # コンパニオン形式だが -f なし → 無処理
```

`-f` がないと、`--name` の形式にかかわらずディスパッチャは何もしません。
`--name` フラグはそのまま Claude に渡されます。

## マルチセッションブートストラップフロー

```
ターミナル 1 (リード)        ターミナル 2-5 (コンパニオン)
─────────────────          ────────────────────────
moai cc -f                 moai cc -f --name plan-<run-id>
                           moai cc -f --name run-<run-id>
                           moai cc -f --name review-<run-id>
                           moai cc -f --name sync-<run-id>
```

ブートストラップは手動です: セッションは別のセッションを起動できません。
リードの SessionStart 通知がコピーする4つのコマンドを正確に出力します。
各コンパニオンを GLM バックエンドで実行するには、`moai cc` を `moai glm`
に置き換えてください。

## クロスセッションメッセージング

セッション間通信は Claude Code のクロスセッションメッセージング
（`ListAgents` / `SendMessage`）を使用します。`crossSessionInbound` 設定
フィールドがインバウンドメッセージを受け入れるか、保留するか、拒否するかを
制御します。

ファクトリーモードはインバウンドメッセージを自動受諾します: ランチャーが
`{"crossSessionInbound": "accept"}` を含む一時設定ファイルを書き込み、
`--settings` でバックエンドに渡します。ファイルはセッション専用で（終了時に
クリーンアップ）、永続設定を変更しません。

### オペレーター提供の `--settings`

コマンドラインに `--settings <file>` を渡すと、ランチャーは自身の設定ファイルを
注入しません。ファイルに以下が含まれていることを確認してください:

```json
{
  "crossSessionInbound": "accept"
}
```

リードの SessionStart 通知は、ランチャーが注入しなかった場合に確認を促す
アドバイスを出力します。

## SessionStart 通知

リード通知はラン id、4つのコンパニオン起動コマンド、リーダーソケットパス、
インバウンド自動化ステータスを案内します。コンパニオン通知は参加を確認する
ロールなしの1行です。どちらの通知もプロンプトせず、情報提供用の stdout です。
