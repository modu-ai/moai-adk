---
title: MoAI-ADK ドキュメント
weight: 99
draft: false
---

MoAI-ADK (Agentic Development Kit) は、Claude Code 用の戦略的オーケストレーションフレームワークです。

> **現在のバージョン:** {{< version >}} — バージョン情報は `hugo.toml` の `params.version` 単一の信頼できる情報源 (SSOT) を参照します。

![MoAI-ADK](/og.jpg)

![ドキュメント構造マップ](/images/sections/doc-map-ja.png)

## MoAI 3.0の3つのコアバリュー

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
