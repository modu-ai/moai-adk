---
title: MoAI-ADK ドキュメント
weight: 99
draft: false
---

MoAI-ADK (Agentic Development Kit) は、Claude Code 用の戦略的オーケストレーションフレームワークです。

> **現在のバージョン:** {{< version >}} — バージョン情報は `hugo.toml` の `params.version` 単一の信頼できる情報源 (SSOT) を参照します。

![MoAI-ADK](/og.jpg)

![ドキュメント構造マップ](/images/sections/doc-map-ja.png)

## v3.1 の新機能 — カンバンモード {{< new-badge v3.1 >}}

セッションはコンテキストウィンドウを1つしか持たず、長いSPECはそれを埋め尽くす — 後に続く作業は先行したすべてを背負ったまま進む。カンバンモードは1つの作業を**5つのターミナル**に分ける。リードセッションがチェーンを進め、4つの同伴セッションが`plan`・`run`・`review`・`sync`を1列ずつ受け持ち、自分の列の文脈だけを背負う。上限が消えるわけではないが、どのセッションも4フェーズ分の履歴を抱え込まないため、同じ予算がはるかに遠くまで届く。

![カンバンモードのひとつのラン — リードセッションと4つの同伴セッションが、それぞれのターミナルで、それぞれのモデルとeffortで動いている](/images/profile/kanban-five-sessions.png)

列ごとにバックエンドとeffortを変えられる。上の画面ではPlanをOpus 5のhighで、RunをGLM 5.2のxhighで、SyncをGLM 5.2で動かしている。

{{< terminal title="kanban mode" raw="true" >}}
moai cc -k                          # リード — run-id を知らせ、チェーンを敷く
moai cc -k --name plan-<run-id>     # 同伴セッション、それぞれ別のターミナルで
moai cc -k --name run-<run-id>
moai cc -k --name review-<run-id>
moai cc -k --name sync-<run-id>
{{< /terminal >}}

ボードは`backlog → plan → run → review → sync → done`の6列で、`backlog`には意図的に担当セッションを置いていない — 作業は[`/moai todo`](/ja/utility-commands/moai-todo)で人が入れたときにだけボードに入る。リードはカードの`progress.md`から自分で読んだ証拠だけでカードを進め、同伴セッションの返信では進めない。

`moai web`を立ち上げると、カンバン画面で5セッションのチェーンとSPECパイプラインを並べて見られる。

![moai web コンソールのOverview画面 — SPEC集計、進行中SPEC一覧、セッションレジストリ](/images/profile/web-console-v31-overview.png)

詳しくは: [カンバンモード](/ja/advanced/kanban-mode) · [manager-kanban エージェント](/ja/advanced/manager-kanban) · [`/moai todo`](/ja/utility-commands/moai-todo) · [moai web コンソール](/ja/advanced/moai-web-console)

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
