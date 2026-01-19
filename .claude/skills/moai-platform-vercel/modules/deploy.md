# Instant Deployment Module

Deploy any project to Vercel instantly without authentication. This module provides zero-configuration deployment with automatic framework detection and claimable deployment URLs.

---

## Overview

The Instant Deployment feature enables rapid project deployment to Vercel's global CDN without requiring authentication or account setup. Projects are deployed as claimable deployments that can be transferred to a Vercel account later.

### Key Features

- Zero authentication required for deployment
- Automatic framework detection from package.json
- Support for 40+ frameworks including Next.js, React, Vue, Svelte, and more
- Preview URLs for immediate access to deployed sites
- Claim URLs to transfer deployments to personal Vercel accounts
- Static HTML project support with automatic index.html handling

### How It Works

1. Project files are packaged into a tarball (excludes node_modules and .git)
2. Framework is auto-detected from package.json dependencies
3. Package is uploaded to Vercel's claimable deployment service
4. Preview URL and Claim URL are returned for immediate access

---

## Usage

### Basic Deployment

Deploy the current directory:

```bash
bash .claude/skills/moai-platform-vercel/scripts/deploy.sh
```

Deploy a specific project directory:

```bash
bash .claude/skills/moai-platform-vercel/scripts/deploy.sh /path/to/project
```

Deploy an existing tarball:

```bash
bash .claude/skills/moai-platform-vercel/scripts/deploy.sh /path/to/project.tgz
```

### Integration with Claude Code

When users request deployment actions, invoke the deploy script:

```bash
# Deploy user's project
bash .claude/skills/moai-platform-vercel/scripts/deploy.sh "$PROJECT_PATH"
```

Common trigger phrases:
- "Deploy my app"
- "Deploy this to production"
- "Create a preview deployment"
- "Deploy and give me the link"
- "Push this live"

---

## Output Format

### Console Output (stderr)

```
Preparing deployment...
Detected framework: nextjs
Creating deployment package...
Deploying...

Deployment successful!

Preview URL: https://skill-deploy-abc123.vercel.app
Claim URL:   https://vercel.com/claim-deployment?code=...
```

### JSON Output (stdout)

The script outputs JSON to stdout for programmatic use:

```json
{
  "previewUrl": "https://skill-deploy-abc123.vercel.app",
  "claimUrl": "https://vercel.com/claim-deployment?code=...",
  "deploymentId": "dpl_...",
  "projectId": "prj_..."
}
```

### Response Fields

| Field | Description |
|-------|-------------|
| previewUrl | Live URL where the deployed site is accessible |
| claimUrl | URL to transfer the deployment to a Vercel account |
| deploymentId | Unique identifier for this deployment |
| projectId | Unique identifier for the project |

---

## Framework Detection

The deploy script automatically detects frameworks from package.json dependencies. Detection order prioritizes more specific frameworks before generic ones.

### React Ecosystem

| Framework | Detection Key |
|-----------|--------------|
| Blitz.js | blitz |
| Next.js | next |
| Gatsby | gatsby |
| Remix | @remix-run/ |
| React Router | @react-router/ |
| TanStack Start | @tanstack/start |
| Create React App | react-scripts |
| Ionic React | @ionic/react |
| Preact | preact |

### Vue Ecosystem

| Framework | Detection Key |
|-----------|--------------|
| Nuxt | nuxt |
| VitePress | vitepress |
| VuePress | vuepress |
| Gridsome | gridsome |

### Svelte Ecosystem

| Framework | Detection Key |
|-----------|--------------|
| SvelteKit | @sveltejs/kit |
| Svelte | svelte |
| Sapper (legacy) | sapper |

### Other Frontend Frameworks

| Framework | Detection Key |
|-----------|--------------|
| Astro | astro |
| Angular | @angular/core |
| Ionic Angular | @ionic/angular |
| Ember | ember-cli, ember-source |
| Docusaurus | @docusaurus/core |
| Solid Start | @solidjs/start |
| Stencil | @stencil/core |
| Dojo | @dojo/framework |
| Polymer | @polymer/ |
| UmiJS | umi |

### Backend Frameworks

| Framework | Detection Key |
|-----------|--------------|
| Express | express |
| Hono | hono |
| Fastify | fastify |
| NestJS | @nestjs/core |
| Elysia | elysia |
| h3 | h3 |
| Nitro | nitropack |

### Build Tools and Other

| Framework | Detection Key |
|-----------|--------------|
| Vite | vite |
| Parcel | parcel |
| Hydrogen (Shopify) | @shopify/hydrogen |
| RedwoodJS | @redwoodjs/ |
| Hexo | hexo |
| Eleventy | @11ty/eleventy |
| Sanity v3 | sanity |
| Sanity | @sanity/ |
| Storybook | @storybook/ |
| Saber | saber |

### Static HTML Projects

For projects without a package.json:
- Framework is set to null
- If there's exactly one HTML file not named index.html, it gets renamed automatically
- This ensures the page is served at the root URL

---

## Presenting Results to Users

Always display both URLs to users after successful deployment:

```
Deployment successful!

Preview URL: https://skill-deploy-abc123.vercel.app
Claim URL:   https://vercel.com/claim-deployment?code=...

View your site at the Preview URL.
To transfer this deployment to your Vercel account, visit the Claim URL.
```

---

## Troubleshooting

### Network Egress Error

If deployment fails due to network restrictions (common in sandboxed environments):

```
Deployment failed due to network restrictions. To fix this:

1. Go to https://claude.ai/admin-settings/capabilities
2. Add *.vercel.com to the allowed domains
3. Try deploying again
```

### Common Issues

**Error: Input must be a directory or a .tgz file**
- Ensure the path points to a valid directory or tarball
- Check that the path exists and is accessible

**Error: Could not extract preview URL from response**
- Network connectivity issues may have occurred
- The deployment service may be temporarily unavailable
- Check the raw response for error details

**Framework not detected**
- Ensure package.json exists in the project root
- Verify the framework dependency is listed in dependencies or devDependencies
- For unsupported frameworks, deployment still works with framework set to null

### Excluded Files

The following are automatically excluded from deployment:
- node_modules/ - Dependencies are installed during build
- .git/ - Git history is not needed for deployment

---

## Best Practices

### Pre-Deployment Checklist

1. Ensure build scripts are defined in package.json
2. Verify environment variables are configured in vercel.json if needed
3. Test the build locally before deploying
4. Remove sensitive files from the project directory

### Performance Optimization

- Keep project size minimal by excluding unnecessary files
- Use .vercelignore to exclude additional files from deployment
- Consider using Vercel's build cache for faster deployments

### Security Considerations

- Never include secrets or API keys in deployed code
- Use Vercel environment variables for sensitive configuration
- Review the claim URL before sharing - it allows account transfer
