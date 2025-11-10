const nextra = require("nextra");
const withNextra = nextra.default || nextra;

module.exports = withNextra({
  theme: "nextra-theme-docs",
  themeConfig: "./theme.config.tsx",
  staticImage: true,
  latex: true,
  codeHighlight: true,
});
