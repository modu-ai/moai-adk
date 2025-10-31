# DevOps Plugin

**Multi-cloud deployment with Vercel, Supabase, Render MCPs** — CI/CD automation, environment management, infrastructure as code, monitoring setup.

## 🎯 What It Does

Automate multi-cloud deployments and infrastructure management across platforms:

```bash
/plugin install moai-plugin-devops
```

**Automatically provides**:
- 🚀 Vercel deployment automation and edge function configuration
- 🗄️ Supabase PostgreSQL, authentication, and real-time setup
- 🏗️ Render backend deployment and environment management
- 🔄 CI/CD pipeline automation
- 📊 Monitoring and alerting setup
- 🔐 Environment variable and secrets management

## 🏗️ Architecture

### 4 Specialist Agents

| Agent | Role | When to Use |
|-------|------|------------|
| **Deployment Strategist** | Multi-cloud strategy, optimization | Planning deployments |
| **Vercel Specialist** | Vercel platform, edge functions, preview envs | Frontend deployment |
| **Supabase Specialist** | PostgreSQL, auth, real-time | Database and auth setup |
| **Render Specialist** | Render platform, environment config | Backend deployment |

### 6 Skills

1. **moai-saas-vercel-mcp** — Vercel deployment, edge functions, preview environments
2. **moai-saas-supabase-mcp** — PostgreSQL database, authentication, real-time subscriptions
3. **moai-saas-render-mcp** — Render deployment, environment management, production setup
4. **moai-domain-backend** — Backend patterns, scalability, security
5. **moai-domain-frontend** — Frontend deployment, optimization, performance
6. **moai-domain-devops** — CI/CD, monitoring, infrastructure patterns

## ⚡ Quick Start

### Installation

```bash
/plugin install moai-plugin-devops
```

### MCP Server Configuration

This plugin uses Model Context Protocol servers for cloud integrations:

**Required Environment Variables**:
```bash
VERCEL_API_TOKEN=<your-vercel-token>
SUPABASE_URL=<your-supabase-url>
SUPABASE_KEY=<your-supabase-key>
RENDER_API_KEY=<your-render-key>
```

### Use with MoAI-ADK

The devops plugin provides agents for:

1. **Plan deployments** - Deployment Strategist handles strategy
2. **Deploy frontend** - Vercel Specialist manages frontend
3. **Setup database** - Supabase Specialist configures database
4. **Deploy backend** - Render Specialist manages backend

## 📊 Multi-Cloud Architecture

```
Application Layer
    ├─ Frontend (Vercel)
    ├─ Backend (Render)
    └─ Database (Supabase)

Infrastructure
    ├─ CI/CD (GitHub Actions)
    ├─ Monitoring (Built-in dashboards)
    └─ Secrets Management (Environment vars)
```

## 🎨 Features

### Vercel Integration
- **Frontend Deployment** - Automated Next.js/React deployments
- **Edge Functions** - Serverless functions at edge locations
- **Preview Environments** - Automatic PR preview deployments
- **Performance Monitoring** - Built-in Web Vitals tracking
- **Custom Domains** - Domain configuration automation

### Supabase Integration
- **PostgreSQL Database** - Fully managed relational database
- **Authentication** - Built-in auth with OAuth, email, phone
- **Real-time Subscriptions** - WebSocket-based live updates
- **Row Level Security** - Fine-grained access control
- **Storage** - File storage with CDN integration

### Render Integration
- **Backend Deployment** - Docker-based app deployment
- **Environment Management** - Deploy to staging/production
- **PostgreSQL Hosting** - Managed database option
- **Cron Jobs** - Scheduled background tasks
- **Private Services** - Internal service networking

### DevOps Features
- **CI/CD Automation** - GitHub Actions workflows
- **Environment Configuration** - Multi-environment setup
- **Secrets Management** - Secure credential storage
- **Monitoring & Logging** - Application performance tracking
- **Infrastructure as Code** - Reproducible deployments

## 🚀 Typical Deployment Workflow

### Frontend Deployment (Vercel)

```
Push to main
    ↓
GitHub Actions trigger
    ↓
Build frontend (Next.js)
    ↓
Deploy to Vercel
    ├─ Production environment
    ├─ Build optimization
    └─ Performance monitoring
    ↓
Automatic DNS update (if custom domain)
```

### Backend Deployment (Render)

```
Push to main
    ↓
GitHub Actions trigger
    ↓
Build Docker image
    ↓
Push to Render
    ├─ Environment variables
    ├─ Health checks
    └─ Auto-scaling setup
    ↓
Database migration (if needed)
```

### Database Setup (Supabase)

```
Create Supabase project
    ↓
Configure PostgreSQL
    ├─ User roles
    ├─ Row level security
    └─ Extensions
    ↓
Setup authentication
    ├─ OAuth providers
    ├─ Email verification
    └─ MFA options
    ↓
Enable real-time
    ├─ WebSocket connections
    ├─ Broadcast channels
    └─ Presence tracking
```

## 📚 Skills Explained

### moai-saas-vercel-mcp
Vercel platform automation:
- **Deployments** - Automated Next.js deployments
- **Edge Functions** - Serverless API endpoints
- **Previews** - Automatic PR preview links
- **Performance** - Web Vitals monitoring
- **Custom Domains** - DNS and SSL setup

### moai-saas-supabase-mcp
Supabase database and auth:
- **PostgreSQL** - Managed relational database
- **Authentication** - OAuth, email, SMS auth
- **Real-time** - Live data subscriptions
- **Storage** - File uploads with CDN
- **Vectors** - Embedding search (pgvector)

### moai-saas-render-mcp
Render deployment platform:
- **Web Services** - Application hosting
- **Background Workers** - Job processing
- **Cron Jobs** - Scheduled tasks
- **PostgreSQL** - Database hosting
- **Private Services** - Internal networking

### moai-domain-devops
DevOps patterns and practices:
- **CI/CD Pipelines** - Automated testing/deployment
- **Infrastructure as Code** - Reproducible environments
- **Monitoring** - Application health tracking
- **Logging** - Centralized log aggregation
- **Secrets Management** - Secure credential handling

## 🔗 Integration with Other Plugins

- **Backend Plugin**: Deploy FastAPI backends to Render
- **Frontend Plugin**: Deploy Next.js frontends to Vercel
- **Technical Blog Plugin**: Publish content to Supabase

## 🔧 Common DevOps Tasks

### Setup CI/CD Pipeline
1. Create GitHub Actions workflow
2. Configure build and test steps
3. Deploy to staging environment
4. Deploy to production environment

### Configure Environment Variables
1. Set up `.env.local` for development
2. Configure Render environment variables
3. Setup Vercel environment variables
4. Manage Supabase project settings

### Database Management
1. Create Supabase project
2. Run migrations
3. Configure backups
4. Setup row-level security

### Monitoring and Alerts
1. Configure Render alerts
2. Setup Vercel Web Vitals monitoring
3. Enable Supabase metrics
4. Create dashboards

## 📖 Documentation

- Vercel Docs: https://vercel.com/docs
- Supabase Docs: https://supabase.com/docs
- Render Docs: https://render.com/docs
- GitHub Actions: https://docs.github.com/actions

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## 📄 License

MIT License - See LICENSE file for details

---

**Created by**: GOOS
**Version**: 1.0.0-dev
**Status**: Development
**Updated**: 2025-10-31
