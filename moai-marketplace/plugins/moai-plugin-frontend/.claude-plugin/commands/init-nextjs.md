# init-nextjs
Initialize Next.js 16 project with React 19, TypeScript, shadcn/ui, Tailwind CSS.

## Usage
```
/init-nextjs [project-name] [--app-router] [--shadcn]
```

## Options
- `--app-router`: Use App Router (default: true)
- `--shadcn`: Install shadcn/ui components
- `--tailwind`: Setup Tailwind CSS (default: true)

## What It Does
1. Creates Next.js project with latest features
2. Configures TypeScript strict mode
3. Sets up Tailwind CSS with custom config
4. Installs shadcn/ui if requested
5. Creates project structure
6. Sets up ESLint and Prettier

## Example
```bash
/init-nextjs my-app --shadcn
```
