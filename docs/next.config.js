const nextra = require("nextra");
const withNextra = nextra.default || nextra;

/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    webpackBuildWorker: true,
  },
  images: {
    unoptimized: false,
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'moai-adk.gooslab.ai',
      },
    ],
  },
  compiler: {
    removeConsole: false,
    reactRemoveProperties: false,
  },
  compress: true,
  poweredByHeader: false,
};

module.exports = withNextra({
  theme: "nextra-theme-docs",
  themeConfig: "./theme.config.tsx",
  staticImage: true,
  latex: true,
  codeHighlight: true,
  defaultShowCopyCode: true,
  search: {
    codeblocks: false,
  },
})(nextConfig);