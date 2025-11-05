const withNextra = require('nextra')({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.cjs'
})

module.exports = withNextra({
  reactStrictMode: true,
  swcMinify: true
})
