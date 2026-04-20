import type { MetaRecord } from "nextra";

/**
 * 유틸리티 명령어 - 자동화 및 피드백 명령어
 *
 * 새로운 형식: /moai [subcommand]
 * 레거시 형식(/moai:XXXX)도 호환됨
 */
const meta: MetaRecord = {
  index: { title: "유틸리티 명령어", display: "hidden" },
  moai: "/moai",
  "moai-github": "/moai github",
  "moai-loop": "/moai loop",
  "moai-fix": "/moai fix",
  "moai-clean": "/moai clean",
  "moai-mx": "/moai mx",
  "moai-feedback": "/moai feedback",
};

export default meta;
