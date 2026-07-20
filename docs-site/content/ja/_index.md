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

- **🪙 トークノミクス** — コンテキストダイエットとプロンプトキャッシングで推論コストを60-70%削減します。[マルチ LLM](/multi-llm)、[コスト最適化](/cost-optimization)、[高度な使い方/トークノミクス概要](/advanced/tokenomics-overview)を参照してください。

- **🧠 再帰的自己学習** — 意思決定メモリと自律エージェントシステムで自律改善ループを実現します。[自己進化システム](/advanced/self-evolving)、[自律ループ](/advanced/autonomous-loops)、[意思決定メモリ](/advanced/decision-memory)を参照してください。

- **🛡️ エージェント型ハーネス** — スキル、フック、MCPで構成可能な実行環境で拡張可能なエージェントオーケストレーションを提供します。[コアコンセプト](/core-concepts)、[ワークフローコマンド](/workflow-commands)、[エージェントガイド](/advanced/agent-guide)を参照してください。

## 主な機能

- **MoAI Orchestrator**: 専門エージェントによる戦略的タスク委譲
- **SPEC ベース TDD/DDD**: 自動方法論選択 — 新規プロジェクトはTDD、レガシーはDDD
- **TRUST 5 Framework**: テスト・可読性・統一性・セキュリティ・追跡性の5原則
- **Progressive Disclosure**: 3段階スキルロードで67%トークン削減

## はじめに

MoAI-ADK を始めるには、[はじめに](/getting-started)セクションを参照してください。

## ドキュメント構成

- [はじめに](/getting-started) - インストール、基本設定、クイックスタート
- [コアコンセプト](/core-concepts) - SPEC 形式、エージェント、ワークフロー
- [高度な使い方](/advanced) - 高度なパターン、スキルの使用、パフォーマンス最適化
- [Git Worktree](/worktree) - Git Worktree CLI 完全ガイド
