---
title: moai epic エピック進捗
weight: 100
draft: false
description: "複数 SPEC に分かれたエピックのマイルストーン進捗をディスクから計算する moai epic status コマンド。"
---

`moai epic status` は 1 つのエピックが今どこまで進んだかを、**ディスクにあるものだけを読んで**計算します。進捗用の保存先を別に持たず、その都度 SPEC ドキュメントを辿ってマイルストーンマップを組み直します。読み取り専用なので、どのファイルも変更しません。

エピック 1 つが複数の SPEC に分かれると「いくつ終わったのか」がすぐ分からなくなります。各 SPEC を開いて status を確認する代わりに、このコマンドが 1 行で答えます。

## 概要

```bash
moai epic status <prefix> [OPTIONS]
```

`<prefix>` は必須の引数で、エピックを識別する SPEC-ID の接頭辞です。たとえば `KANBAN` を渡すと `.moai/specs/SPEC-KANBAN-*/spec.md` が対象になります。

## 何を読むか

マイルストーンマップは 3 つの情報から作られます。

1. **SPEC frontmatter** — 各 `spec.md` の `status` の値。マイルストーンが終わったかを判断する根拠です。
2. **タイトルのマイルストーン標識** — SPEC タイトルに書かれた `(TOKEN Mx)` 形式の標識。どの SPEC がどのマイルストーンかを結びつけます。
3. **デザインレポート（任意）** — 存在すれば正規マイルストーン一覧の出所として使います。自動探索され、`--design-report` で明示的に指定できます。

標識のない SPEC は捨てられず、`untracked_specs` として別に報告されます — 黙って外すと「その SPEC は存在しない」と読まれるからです。

## フラグ

| フラグ | 説明 |
|--------|------|
| `--json` | 固定形状の JSON を stdout に出力 |
| `--design-report <path>` | デザインレポートの自動探索に代えてパスを指定 |
| `--marker <token>` | 推定されたエピックトークンを明示指定（例: `BAS`） |
| `--base-dir <path>` | プロジェクトルート（既定値: カレントディレクトリ） |

## 例

既定の出力は人が読む進捗ボードです。

```bash
$ moai epic status KANBAN
🎯 KANBAN ▓▓▓▓▓░░░░░ 2/4 (50%)
Epic progress:   KANBAN
  🟢 M0 M0                            SPEC-KANBAN-RENAME-001 (completed)
  ⬜ M1 M1                             SPEC-KANBAN-BOOTSTRAP-001 (draft)
  ⬜ M2 M2                             SPEC-KANBAN-WORKTREE-001 (draft)
  🟢 M3 M3                            SPEC-KANBAN-BOARD-001 (completed)
```

標識が 1 つもない場合はその事実をそのまま書き、代わりに引っかかった SPEC を並べます。

```bash
$ moai epic status DESIGN-DOCS
🎯 DESIGN-DOCS — 2 SPEC(s) matched, none carrying a (TOKEN Mx) milestone marker
Epic progress:   DESIGN-DOCS
untracked_specs: SPEC-DESIGN-DOCS-001, SPEC-DESIGN-DOCS-V31-001
```

`--json` はスクリプトから使いやすい固定形状を出します。

```bash
$ moai epic status KANBAN --json
{
  "epic": "KANBAN",
  "epic_token": "KANBAN",
  "milestones": [
    {
      "id": "M0",
      "label": "M0",
      "status": "done",
      "covered": true,
      "spec_id": "SPEC-KANBAN-RENAME-001",
      "spec_status": "completed",
      "sync_commit_sha": "144573336d07da19f4b8a50aa26c38db2704afb5"
    }
  ],
  "done": 2,
  "total": 4,
  "pct": 50,
  "extra_mx": [],
  "untracked_specs": ["SPEC-KANBAN-TODO-CLI-001"],
  "baseline_attribution": "3b9b3bf9959669c4bfc43da313e25bca61f910a2"
}
```

## JSON フィールド

| フィールド | 説明 |
|------------|------|
| `epic` | 引数として渡した接頭辞 |
| `epic_token` | タイトル標識から見つけたエピックトークン。見つからなければ空文字列 |
| `milestones` | マイルストーンの配列。各項目は id・ラベル・状態・担当 SPEC・その SPEC の status・sync コミット SHA |
| `done` / `total` / `pct` | 完了数、全体数、百分率 |
| `extra_mx` | 正規一覧にはないが標識だけがあるマイルストーン |
| `untracked_specs` | 接頭辞には一致したがマイルストーン標識がない SPEC |
| `baseline_attribution` | 計算時点の git コミット SHA |

`baseline_attribution` が併記されるのは、この値が**どの時点のツリーを読んだ結果か**を残しておくためです。進捗だけを書き留めても、いつ測った値かが分かりません。

## 読み取り専用

このコマンドは観測しかしません。SPEC の status を変えず、進捗をどこにも保存せず、実行のたびにディスクから計算し直します。だからこそ結果が実際のファイルとずれることがありません。

計算結果はウェブコンソールのモニター領域にもエピックパネルとして現れます。[MoAI Web Console](/ja/advanced/moai-web-console/) を参照してください。

---

関連: [moai spec ドキュメント管理](/ja/cli-reference/spec/) · [MoAI Web Console](/ja/advanced/moai-web-console/)
