const nextra = require("nextra");
const withNextra = nextra.default || nextra;

module.exports = withNextra({
  staticImage: true,
  latex: true,
  codeHighlight: true,
  output: 'export',
  trailingSlash: true,
  images: {
    unoptimized: true
  }
});