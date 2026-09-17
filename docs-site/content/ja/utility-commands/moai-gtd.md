---
title: /moai gtd
weight: 29
draft: false
new: true
---

# /moai gtd

`/moai gtd` と `moai gtd` は、仕事を収集し、実行可能かを判断してから既存の開発キューへつなぐ正式な GTD 入口です。従来の `todo` SQLite DB、カード ID、並び順、`queued`・`picked`・`dropped` 状態、アーカイブ・復元の意味は変わりません。`todo` は同じコマンドツリーを使う互換名として残ります。

```bash
moai gtd add "認証のエラー経路を整理"
moai gtd list
moai gtd next t1 --spec SPEC-AUTH-001
moai gtd done t1 --expect "認証"
```

従来の `moai todo ...` も同じ結果になります。キューの全動詞とフラグは [todo 互換コマンドのリファレンス](/ja/utility-commands/moai-todo) を参照してください。

GTD 専用の 5 動詞は、同じ SQLite 項目を順に引き継ぎます。

```bash
moai gtd capture "認証のエラー経路を整理" --event inbox-42 --source user --sensitivity private
moai gtd clarify <gtd-id> --disposition action --outcome "マージ済み" --evidence "テストとマージ SHA" --authority "queue,commit" --trusted
moai gtd organize <gtd-id> --class action --context computer --depends-on <gtd-id>
moai gtd reflect --rebuild-projection --json
moai gtd engage <gtd-id> --approve --fresh --dependencies-ready --resources --pick --dispatch --lane lane-10 --run-id mission-42
```

`capture` では再試行の重複を防ぐ安定した `--event` が必須です。`engage` からの配車には `--pick`、`--lane`、`--run-id` がすべて必要で、承認・鮮度・依存関係・資源の確認が一つでも欠ければ効果を作りません。操作 receipt を先に SQLite へ準備し、中断後は同じ操作 ID の実状態を読み直してから再試行を決めます。

明示的 opt-in の GTD v2 export/import は、項目・関係に加えて封印契約、ミッション、イベント、操作 receipt を保存し、import 時に整合性を検査します。カードの archive/reopen は同じ GTD identity を維持し、`reflect` はアーカイブ済み完了・再開・取消・source revision の変化を読み直して、古い根拠と blocked successor を表示します。

## 5 つの手順

GTD は開発ボードの列ではなく、**仕事を整理して選ぶための手順**です。

| 手順 | 判断と保存結果 |
|---|---|
| Capture | 出典・機密区分・重複識別子を記録し、まだ開発カードは作りません。 |
| Clarify | 望む結果、完了根拠、権限、出典の信頼性を確認します。不明な項目があれば発行を保留します。 |
| Organize | プロジェクト・行動・参考・保留の分類、実行コンテキスト、見直し時期、関係を整理します。依存循環は拒否します。 |
| Reflect | 根拠の変更、アーカイブ・再開・取消、予定された見直しを再評価します。取消を前提作業の完了とは扱いません。 |
| Engage | 承認範囲、根拠の鮮度、依存関係、レーン所有権、資源上限をすべて通った項目だけを推薦またはキューへ接続します。 |

開発は引き続き `backlog → plan → run → sync → done` で進みます。Capture から Engage はこれらの段階を置き換えません。

## 関係と非公開グラフ

実行を止める関係は `depends_on` だけです。`part_of`、`supported_by`、`related_to` は非ブロッキングで、既存の `contains`、`absorbs`、`replaces`、`conflicts` の意味も維持されます。GTD グラフはキュー DB の隣に置く非公開の派生物で、リポジトリ、ログ、テレメトリ、標準エクスポート、標準バックアップには含まれません。メタデータまたは原本 revision が異なれば stale と判定します。

## 自律運用と安全境界

LLM と `mission-governor` は構造化された提案だけを作り、ファイル、Git、キュー、配車状態を直接変更しません。通常コードが、ユーザー承認済みの封印された目標・範囲・許可行為・完了根拠・資源上限・停止条件・最新 snapshot を検査してから操作を準備します。範囲拡大、古い根拠、未許可行為は暗黙に承認せず `blocked` で止めます。

現時点の実装は、決定論的なポリシー・復旧・配車・明示パスのコミット・local develop `--no-ff` マージを担う owner adapter を提供します。コミットには現在の HEAD に対応するリポジトリ内 `0600` テスト receipt が必要で、ローカルマージでは manager-git 役割、`WT-*` ブランチ、基準 SHA、`.git` 配下の `0600` lease を再確認します。バックアップ・復元・エクスポートは明示的 opt-in の場合だけ GTD 拡張を含み、private projection は SQLite revision から再生成します。実プロバイダーで開始・再接続・交代・資格情報・プロセス識別の全能力が確認されるまでは、セッション終了後の継続実行やリモート push・PR・マージ完了を保証しません。

関連: [`/moai goal --auto`](/ja/utility-commands/moai-goal#auto-ミッションモード) · [Kanban Mode](/ja/advanced/kanban-mode)
