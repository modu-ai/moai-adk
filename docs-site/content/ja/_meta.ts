import type { MetaRecord } from "nextra";

/**
 * MoAI-ADK ドキュメント - 6つのセクション構成
 *
 * 1. はじめに - インストール、基本設定、クイックスタート
 * 2. コア概念 - MoAI-ADKとは、SPEC、DDD、TRUST 5
 * 3. ワークフローコマンド - /moai project ~ /moai sync
 * 4. ユーティリティコマンド - /moai、loop、fix、feedback
 * 5. 詳細 - スキル、エージェント、ビルダー、フック、設定
 * 6. Git Worktree - 完全なワークツリー CLI ガイド
 */
const meta: MetaRecord = {
  "getting-started": "はじめに",
  "core-concepts": "コア概念",
  "workflow-commands": "ワークフローコマンド",
  "utility-commands": "ユーティリティコマンド",
  "quality-commands": "品質コマンド",
  agency: "Agency",
  advanced: "詳細",
  worktree: "Git Worktree",
};

export default meta;
