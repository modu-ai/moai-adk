import type { MetaRecord } from "nextra";

/**
 * ユーティリティコマンド - 自動化とフィードバックコマンド
 *
 * 新形式：/moai [subcommand]
 * レガシー形式（/moai:XXXX）もサポート
 */
const meta: MetaRecord = {
  index: { title: "ユーティリティコマンド", display: "hidden" },
  moai: "/moai",
  "moai-github": "/moai github",
  "moai-loop": "/moai loop",
  "moai-fix": "/moai fix",
  "moai-clean": "/moai clean",
  "moai-mx": "/moai mx",
  "moai-feedback": "/moai feedback",
};

export default meta;
