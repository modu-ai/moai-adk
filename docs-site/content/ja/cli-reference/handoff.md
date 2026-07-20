---
title: moai handoff ハンドオフレコード
weight: 68
draft: false
---

`moai handoff` は auto-resume ハンドオフの pending レコードを管理します。セッション境界 (`/clear`) を越えて作業を継続するための paste-ready resume 本文を保存または消去します。保存されたレコードは `handoff.mode: auto` 設定時に次のセッション開始時に自動注入されます。

## サブコマンド

| コマンド | 説明 |
|--------|------|
| `moai handoff save` | paste-ready resume 本文を pending レコードとして保存 |
| `moai handoff clear` | pending レコードを除去 |

共通フラグとして `--project-dir <path>` (プロジェクトルート、デフォルト: 現在のディレクトリ) を受け取ります。

## moai handoff save

```bash
moai handoff save --stdin --spec SPEC-AUTH-001 --phase run < resume.txt
```

| フラグ | 説明 |
|--------|------|
| `--body <text>` | resume 本文 (verbatim 6-block paste-ready) |
| `--stdin` | `--body` の代わりに stdin から本文を読み込む |
| `--spec <id>` | このハンドオフが再開する SPEC id |
| `--phase <plan\|run\|sync>` | フェーズ |
| `--session <uuid>` | saved_by_session uuid (帰属) |
| `--lang <lang>` | conversation_language のスナップショット |
| `--ultrathink` | ultrathink 指示を記録 (復元案内用) |
| `--ultracode` | ultracode 指示を記録 (復元案内用) |
| `--goal <condition>` | `/goal` 条件を記録 (復元案内用) |

## moai handoff clear

```bash
moai handoff clear
```

pending ハンドオフレコードを除去します。

## Fail-open 保証

`moai` CLI が PATH に存在しない場合、または `moai handoff save` が non-zero で終了した場合でも、オーケストレーターの paste-ready 出力は変更されず維持されます。保存の失敗はハンドオフの出力を決して妨げず、手動 paste 経路は保存なしでも完全に機能します — 保存は付加的な永続化ステップであり、ゲートではありません。

## 関連ドキュメント

- [自律継続ループ](/ja/advanced/autonomous-loops)
- [moai goal](/ja/cli-reference/goal)
- [CLI 概要](/ja/getting-started/cli)
