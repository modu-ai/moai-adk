import type { MetaRecord } from "nextra";

/**
 * 品質コマンド - コードレビュー、カバレッジ、E2E、アーキテクチャ分析
 *
 * 新形式：/moai [subcommand]
 * レガシー形式（/moai:XXXX）もサポート
 */
const meta: MetaRecord = {
  index: { title: "品質コマンド", display: "hidden" },
  "moai-review": "/moai review",
  "moai-coverage": "/moai coverage",
  "moai-e2e": "/moai e2e",
  "moai-codemaps": "/moai codemaps",
};

export default meta;
