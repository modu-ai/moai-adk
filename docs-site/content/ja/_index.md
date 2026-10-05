---
title: MoAI-ADK ドキュメント
weight: 99
draft: false
---

MoAI-ADK (Agentic Development Kit) は、Claude Code 用の戦略的オーケストレーションフレームワークです。

> **現在のバージョン:** {{< version >}} — バージョン情報は `hugo.toml` の `params.version` 単一の信頼できる情報源 (SSOT) を参照します。

![MoAI-ADK](/og.jpg)

![ドキュメント構造マップ](/images/sections/doc-map-ja.png)

## v3.2 の新機能 — ファクトリーモード {{< new-badge v3.2 >}}

セッションはコンテキストウィンドウを1つしか持たず、長いSPECはそれを埋め尽くす。後に続く作業は先行したすべてを背負ったまま進む。ファクトリーモードは作業を**複数のターミナル**に分ける。リーダーセッションが選ばれたカードを割り当て、レーンごとに立ち上げた別のセッションが、カード1枚を`plan`・`run`・`sync`まで丸ごと受け持つ。レビュー判定は独立した段階ではなく、syncゲートが吸収する。上限が消えるわけではないが、どのセッションもカード1枚分を超える履歴を抱え込まないため、同じ予算がはるかに遠くまで届く。

レーンごとにバックエンドを変えられる。リーダーと各レーンは、`cc`・`glm`・`codex`の中からそれぞれ選べる。

{{< terminal title="factory mode" raw="true" >}}
moai cc -f                    # リーダー: ファクトリーのランを開く
moai cc -l                    # レーン、それぞれ別のターミナルで: 次の空きレーンとして合流する
moai glm -l                   # GLMバックエンドのレーン
moai codex -l                 # Codexのレーン
{{< /terminal >}}

委任のチャネルはキューである。`backlog`には意図的に担当セッションを置いておらず、作業は[`/moai todo`](/ja/utility-commands/moai-todo)で人が入れたときにだけ入る。リーダーはカードの`progress.md`から自分で読んだ証拠だけでカードを進め、レーンの返信では進めない。

`moai web`を立ち上げると、ファクトリー画面でファクトリーのレーンとSPECパイプラインを並べて見られる。

![moai web コンソールのOverview画面 — SPEC集計、進行中SPEC一覧、セッションレジストリ](/images/profile/web-console-v31-overview.png)

詳しくは: [ファクトリーモード](/ja/advanced/factory-mode) · [manager-lead リーダーコーディネーター](/ja/advanced/manager-lead) · [`/moai todo`](/ja/utility-commands/moai-todo) · [moai web コンソール](/ja/advanced/moai-web-console)

## MoAI 3.1の3つのコアバリュー

- {{< icon database primary >}} **トークノミクス** — コンテキストダイエットとプロンプトキャッシングで推論コストを60-70%削減します。[マルチ LLM](/ja/multi-llm)、[コスト最適化](/ja/cost-optimization)、[高度な使い方/トークノミクス概要](/ja/advanced/tokenomics-overview)を参照してください。

- {{< icon rotate primary >}} **エージェンティック・ループ・エンジニアリング** — 意思決定メモリと自律エージェントシステムで自律改善ループを実現します。[自己進化システム](/ja/advanced/self-evolving)、[自律ループ](/ja/advanced/autonomous-loops)、[意思決定メモリ](/ja/advanced/decision-memory)を参照してください。

- {{< icon package primary >}} **エージェント型ハーネス** — スキル、フック、MCPで構成可能な実行環境で拡張可能なエージェントオーケストレーションを提供します。[コアコンセプト](/ja/core-concepts)、[ワークフローコマンド](/ja/workflow-commands)、[エージェントガイド](/ja/advanced/agent-guide)を参照してください。

## 主な機能

- **MoAI Orchestrator**: 専門エージェントによる戦略的タスク委譲
- **SPEC ベース TDD/DDD**: 自動方法論選択 — 新規プロジェクトはTDD、レガシーはDDD
- **TRUST 5 Framework**: テスト・可読性・統一性・セキュリティ・追跡性の5原則
- **Progressive Disclosure**: 3段階スキルロードで67%トークン削減

## はじめに

MoAI-ADK を始めるには、[はじめに](/ja/getting-started)セクションを参照してください。

## ドキュメント構成

- [はじめに](/ja/getting-started) - インストール、基本設定、クイックスタート
- [コアコンセプト](/ja/core-concepts) - SPEC 形式、エージェント、ワークフロー
- [高度な使い方](/ja/advanced) - 高度なパターン、スキルの使用、パフォーマンス最適化
- [Git Worktree](/ja/worktree) - Git Worktree CLI 完全ガイド
