import type { MetaRecord } from "nextra";

/**
 * ワークフローコマンド - 4フェーズ開発サイクルコマンド
 *
 * 新形式：/moai [subcommand]
 * レガシー形式（/moai:X-YYYY）もサポート
 */
const meta: MetaRecord = {
  index: { title: "ワークフローコマンド", display: "hidden" },
  "moai-project": "/moai project",
  "moai-plan": "/moai plan",
  "moai-run": "/moai run",
  "moai-sync": "/moai sync",
};

export default meta;
