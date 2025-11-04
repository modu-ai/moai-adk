const withNextra = require('nextra')({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.jsx',
  unstable_staticImage: true,
  flexsearch: true,
  mdxOptions: {
    remarkPlugins: [],
    rehypePlugins: []
  }
})

module.exports = withNextra({
  reactStrictMode: true,
  experimental: {
    appDir: false
  },
  images: {
    unoptimized: true
  }
})
