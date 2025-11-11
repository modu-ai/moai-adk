const nextra = require("nextra");
const withNextra = nextra.default || nextra;

/** @type {import('next').NextConfig} */
const nextConfig = {
  // Development mode - no static export
  
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

// Use pages directory (default for Nextra)
module.exports = withNextra({
  staticImage: true,
  latex: true,
  codeHighlight: true,
  defaultShowCopyCode: true,
  search: {
    codeblocks: false,
  },
})(nextConfig);
