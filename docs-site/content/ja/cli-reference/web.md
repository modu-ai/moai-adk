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

---

関連: [プロファイル管理](/cli-reference/profile) · [CLI 概要](/getting-started/cli)
