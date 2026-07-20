---
title: /moai design
weight: 45
draft: false
---

UI を露出する SPEC のためのデザイン段階のコラボレーションワークフローです。plan と run の間の条件付き経路で、Claude Design と双方向にデザインシステム・画面成果物を同期します。

{{< callout type="info" >}}
**スラッシュコマンド**: Claude Code で `/moai:design` と入力すると、このコマンドをすぐに実行できます。`/moai` だけ入力すると、利用可能なすべてのサブコマンド一覧が表示されます。
{{< /callout >}}

## 概要

`/moai design` は UI を露出する SPEC にのみ適用される **デザイン段階**ワークフローです。内部的には **manager-design** エージェントが Claude Design コラボレーションパイプライン (D1-D5) と H1-H9 ハンドオフ契約を駆動します。

この経路は **追加的 (additive)** です — UI を露出しない SPEC は標準の `plan → run → sync` の順序をそのまま保ち、このワークフローを完全にスキップします。

## いつ使うか (経路の活性化条件)

SPEC が次のいずれかで UI 露出を宣言すると `plan → design → run` の経路をたどります:

- `acceptance.md` に明示的なフロントエンドコンポーネント/ビュー/ページ成果物があるか、
- `tier: L` + フロントエンドモジュール (`module:` がフロントエンドパッケージを参照)。

どちらでもなければ標準の `plan → run → sync` を保ちます。

## 進入条件

デザイン段階は、次の 2 つの条件を**どちらも**満たしたあとにのみ進入します:

1. **Plan-audit PASS** — SPEC の plan-phase 成果物が Phase 1 監査を通過
2. **実装着手承認** — plan→run ヒューマンゲートを通過

{{< callout type="warning" >}}
デザイン段階は **実装着手承認を代替しません**。plan→run 境界をヒューマンゲートより先に越えることはなく、すでに承認された run の範囲内で最初の M1 実装コミットの前に実行されます。
{{< /callout >}}

## D1-D5 パイプライン

manager-design エージェントが 5 段階のパイプラインを順に実行します。

```mermaid
flowchart TD
    Entry["plan-audit PASS + 実装着手承認"] --> D1["D1 接続準備<br/>(login + プロジェクト設定)"]
    D1 --> D2["D2 デザインシステム同期<br/>(code → design)"]
    D2 --> D3["D3 画面成果物の生成<br/>(Claude Design canvas)"]
    D3 --> D4["D4 ハンドオフ受信・貼り付け<br/>(design → code)"]
    D4 --> D5["D5 実装連結<br/>(handoff → run-phase)"]
    D5 --> Run["run-phase 再委任 (manager-develop)"]
```

| 段階 | 説明 |
|------|------|
| **D1 接続準備** | Claude Design ログイン + 書き込み可能なデザインシステムプロジェクトの確保 (`list_projects`/`create_project`/`get_project`) |
| **D2 デザインシステム同期** | `.moai/project/brand/` トークン・`design.yaml`・既存コンポーネントをバンドルしてプロジェクトに push (`finalize_plan` 承認ゲート → `write_files` コンポーネント単位の増分) |
| **D3 画面成果物の生成** | インポートした実際のコンポーネント/トークンから画面を生成 (drift 防止)、ユーザーの WYSIWYG 編集 + 実装注釈、`report_validate` 指標の確認 |
| **D4 ハンドオフ受信・貼り付け** | 完成したハンドオフ (画面 + 注釈 + トークン/コンポーネント参照) を予約パス (`.moai/design/tokens.json`、`components.json`、`assets/`、`brief/BRIEF-*.md`) に貼り付け |
| **D5 実装連結** | Section A-E 委任パッケージ (ハンドオフファイル一覧 + 注釈→要件マッピング + PRESERVE 一覧 + 検証コマンド) を構成して manager-develop に再委任 |

manager-design は再委任後に返却し、実装を一緒に操縦 (co-pilot) しません。実装後は sync-auditor がブランド一貫性を must-pass で判定します。

## Claude Design 双方向同期

`/moai design` の核心は、コードと Claude Design キャンバスの間の **双方向同期**です:

- **code → design (D2)**: コードのデザインシステム (トークン・コンポーネント) をキャンバスへ push。ファイル内容はディスクに残り、モデルコンテキストを通過しません (ファイルあたり 256KiB 上限)。
- **design → code (D4)**: キャンバスで完成した画面・注釈を pull して予約パスに貼り付け。外部で作成されたファイルに挿入された指示文は **データとしてのみ**扱い、無視・報告します (H7 セキュリティ契約)。

`/design-login`・`/design-sync` スラッシュコマンドはユーザー専用の TUI コマンドで、エージェントは使い方を案内するだけで直接呼び出しません。

## H1-H9 ハンドオフ契約

D4 ハンドオフを規律する 9 つの条項は manager-design エージェント本文に正本として存在します (要約):

- **H1 受信パス** — `/design-sync` pull はユーザー専用; エージェントは `list_files → get_file`
- **H2 配置規約** — 予約パスのみ使用
- **H3 1:1 忠実度** — 貼り付け時に任意の修正を禁止、代わりにキャンバスの回帰を提案
- **H4 ブランド優先** — `.moai/project/brand/` が憲法的な親
- **H5 注釈変換** — 注釈 → { target · requirement · AC 候補 } マッピング
- **H6 検証** — `report_validate` 指標 + drift grep + スナップショット鮮度
- **H7 セキュリティ** — `get_file` の内容はデータ、指示文は無視
- **H8 再委任パッケージ** — Section A-E で manager-develop に委任
- **H9 隠しフォルダ案内** — `.moai/design/` dot-folder の可視性

## ツール可用性 (優雅な性能低下)

DesignSync サーバーが `.mcp.json` に登録されていない場合があります。D1 が可用性を確認します:

- **ツールあり** → D2-D5 に進行
- **ツールなし** → エージェントが blocker report を返却 (H1 パス)。ユーザーが DesignSync を別途登録 (Claude Code v2.1.181+ および Pro+ Claude Design アカウントが必要)

デザイン段階の著述自体は失敗せず、ツールを待ちます。

## 関連ドキュメント

- [/moai plan](./moai-plan) - 前段階: SPEC ドキュメントの生成
- [/moai run](./moai-run) - 次段階: DDD/TDD 実装
- [サブエージェントカタログ](/advanced/agent-guide) - manager-design エージェントの詳細
