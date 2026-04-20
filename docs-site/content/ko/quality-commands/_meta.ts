import type { MetaRecord } from "nextra";

/**
 * 품질 명령어 - 코드 리뷰, 커버리지, E2E, 아키텍처 분석
 *
 * 새로운 형식: /moai [subcommand]
 * 레거시 형식(/moai:XXXX)도 호환됨
 */
const meta: MetaRecord = {
  index: { title: "품질 명령어", display: "hidden" },
  "moai-review": "/moai review",
  "moai-coverage": "/moai coverage",
  "moai-e2e": "/moai e2e",
  "moai-codemaps": "/moai codemaps",
};

export default meta;
