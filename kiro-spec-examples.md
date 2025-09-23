# SPEC 파일 예제 모음

## 📁 3대 핵심 SPEC 파일

### 1. requirements.md - 요구사항 명세

```markdown
# requirements.md - Community Platform

## 📌 User Stories

### US-001: User Authentication

**As a** user  
**I want to** register and login to the platform  
**So that** I can access personalized community features

#### Acceptance Criteria (EARS Format)

- **WHEN** the user provides valid email and password  
  **THEN** the system shall create a new account within 2 seconds
- **IF** the email already exists in the database  
  **THEN** the system shall return HTTP 409 with message "Email already registered"
- **WHILE** the user is logged in  
  **THEN** the system shall maintain JWT session for 24 hours
- **WHERE** authentication is required  
  **THEN** the system shall redirect to login page with return URL
- **UBIQUITOUS** all API calls shall include authentication token in header

### US-002: Post Management

**As a** community member  
**I want to** create, edit, and delete posts  
**So that** I can share knowledge with the community

#### Acceptance Criteria

- **WHEN** creating a post  
  **THEN** the system shall support MDX format with syntax highlighting
- **IF** the post contains images larger than 5MB  
  **THEN** the system shall compress to < 500KB maintaining 80% quality
- **WHILE** editing a post  
  **THEN** the system shall auto-save draft every 30 seconds
- **WHERE** markdown is used  
  **THEN** the system shall sanitize HTML to prevent XSS attacks

### US-003: Real-time Chat

**As a** community member  
**I want to** chat in real-time with other members  
**So that** I can collaborate instantly

#### Acceptance Criteria

- **WHEN** a message is sent  
  **THEN** all connected users shall receive it within 100ms
- **IF** the recipient is offline  
  **THEN** the system shall queue message for delivery on next login
- **WHILE** typing  
  **THEN** the system shall show typing indicator to other users
```

### 2. design.md - 기술 설계 문서

```markdown
# design.md - Technical Architecture

## 🏗️ System Architecture

### Data Models

\`\`\`typescript
// User Entity
interface User {
id: string; // UUID v4
email: string; // Unique, validated
passwordHash: string; // bcrypt 12 rounds
profile: {
displayName: string;
avatar?: string;
bio?: string;
};
role: 'user' | 'moderator' | 'admin';
createdAt: Date;
updatedAt: Date;
}

// Post Entity
interface Post {
id: string;
title: string; // Max 200 chars
content: string; // MDX format
slug: string; // URL-friendly
author: User;
status: 'draft' | 'published' | 'archived';
tags: string[];
metadata: {
views: number;
likes: number;
readTime: number; // Minutes
};
comments: Comment[];
createdAt: Date;
updatedAt: Date;
}

// Comment Entity
interface Comment {
id: string;
content: string;
author: User;
postId: string;
parentId?: string; // For nested comments
reactions: Reaction[];
createdAt: Date;
updatedAt: Date;
}
\`\`\`

### API Design

\`\`\`typescript
// RESTful Endpoints
POST /api/auth/register // User registration
POST /api/auth/login // User login
POST /api/auth/logout // User logout
GET /api/auth/session // Get current session

GET /api/posts // List posts (paginated)
POST /api/posts // Create new post
GET /api/posts/:id // Get single post
PUT /api/posts/:id // Update post
DELETE /api/posts/:id // Delete post

POST /api/posts/:id/comments // Add comment
GET /api/posts/:id/comments // Get comments
DELETE /api/comments/:id // Delete comment

// WebSocket Events
socket.on('join-room') // Join chat room
socket.on('leave-room') // Leave chat room
socket.on('send-message') // Send message
socket.on('typing') // Typing indicator
socket.emit('new-message') // Broadcast message
socket.emit('user-typing') // Broadcast typing
\`\`\`

### Database Schema

\`\`\`sql
-- Users table
CREATE TABLE users (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
email VARCHAR(255) UNIQUE NOT NULL,
password_hash VARCHAR(255) NOT NULL,
display_name VARCHAR(100),
avatar_url TEXT,
bio TEXT,
role VARCHAR(20) DEFAULT 'user',
created_at TIMESTAMP DEFAULT NOW(),
updated_at TIMESTAMP DEFAULT NOW()
);

-- Posts table
CREATE TABLE posts (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
title VARCHAR(200) NOT NULL,
content TEXT NOT NULL,
slug VARCHAR(255) UNIQUE NOT NULL,
author_id UUID REFERENCES users(id) ON DELETE CASCADE,
status VARCHAR(20) DEFAULT 'draft',
views INTEGER DEFAULT 0,
created_at TIMESTAMP DEFAULT NOW(),
updated_at TIMESTAMP DEFAULT NOW()
);

-- Comments table
CREATE TABLE comments (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
content TEXT NOT NULL,
author_id UUID REFERENCES users(id) ON DELETE CASCADE,
post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
created_at TIMESTAMP DEFAULT NOW(),
updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_comments_post ON comments(post_id);
\`\`\`

### Component Architecture

\`\`\`
src/
├── components/
│ ├── auth/
│ │ ├── LoginForm.tsx
│ │ ├── RegisterForm.tsx
│ │ └── AuthGuard.tsx
│ ├── posts/
│ │ ├── PostCard.tsx
│ │ ├── PostEditor.tsx
│ │ ├── PostList.tsx
│ │ └── PostDetail.tsx
│ ├── chat/
│ │ ├── ChatRoom.tsx
│ │ ├── MessageList.tsx
│ │ └── MessageInput.tsx
│ └── common/
│ ├── Layout.tsx
│ ├── Header.tsx
│ └── Footer.tsx
├── lib/
│ ├── api/
│ │ ├── auth.ts
│ │ ├── posts.ts
│ │ └── websocket.ts
│ └── utils/
│ ├── validation.ts
│ └── sanitization.ts
└── pages/
├── api/
│ ├── auth/
│ └── posts/
└── app/
\`\`\`
```

### 3. tasks.md - 구현 태스크 체크리스트

```markdown
# tasks.md - Implementation Tasks

## ✅ Sprint 1: Project Foundation (Week 1)

### Setup & Configuration

- [ ] Initialize Next.js 15 with TypeScript
  - [ ] Configure App Router
  - [ ] Setup Tailwind CSS
  - [ ] Configure path aliases
- [ ] Setup Prisma ORM
  - [ ] Install dependencies
  - [ ] Configure PostgreSQL connection
  - [ ] Create initial schema
  - [ ] Run first migration
- [ ] Configure development tools
  - [ ] ESLint with Next.js config
  - [ ] Prettier with format on save
  - [ ] Husky for pre-commit hooks
  - [ ] Jest for unit testing
  - [ ] Playwright for E2E testing

### Authentication Foundation

- [ ] Install NextAuth.js
  - [ ] Configure JWT strategy
  - [ ] Setup session provider
  - [ ] Add CSRF protection
- [ ] Create auth database schema
  - [ ] Users table with constraints
  - [ ] Sessions table
  - [ ] Add necessary indexes

## ✅ Sprint 2: User Authentication (Week 2)

### Registration Flow

- [ ] Create registration API endpoint
  - [ ] Email validation (RFC 5322)
  - [ ] Password strength check (min 8 chars, 1 number, 1 special)
  - [ ] Check duplicate emails
  - [ ] Hash password with bcrypt (12 rounds)
  - [ ] Send verification email
  - [ ] Unit tests (100% coverage)
- [ ] Build registration UI
  - [ ] Registration form component
  - [ ] Real-time validation feedback
  - [ ] Loading states
  - [ ] Error handling with toast
  - [ ] Success redirect to login
  - [ ] Mobile responsive design
  - [ ] Accessibility (ARIA labels)

### Login Flow

- [ ] Create login API endpoint
  - [ ] Validate credentials
  - [ ] Generate JWT token (24h expiry)
  - [ ] Set secure HTTP-only cookie
  - [ ] Rate limiting (5 attempts/minute)
  - [ ] Integration tests
- [ ] Build login UI
  - [ ] Login form component
  - [ ] Remember me checkbox
  - [ ] Forgot password link
  - [ ] Social login buttons (Google, GitHub)
  - [ ] Loading states
  - [ ] Error messages
  - [ ] Mobile responsive

### Session Management

- [ ] Implement auth middleware
  - [ ] Token verification
  - [ ] Auto-refresh before expiry
  - [ ] Logout functionality
  - [ ] Protected route HOC
- [ ] Create user profile page
  - [ ] Display user info
  - [ ] Edit profile form
  - [ ] Avatar upload
  - [ ] Change password

## ✅ Sprint 3: Core Features (Week 3)

### Post Management

- [ ] Create posts API
  - [ ] CRUD endpoints
  - [ ] MDX content support
  - [ ] Slug generation
  - [ ] Draft/publish status
  - [ ] Pagination (cursor-based)
  - [ ] Search functionality
  - [ ] Unit tests (> 90% coverage)
- [ ] Build post UI components
  - [ ] Post editor with MDX preview
  - [ ] Rich text toolbar
  - [ ] Image upload with drag-drop
  - [ ] Auto-save drafts
  - [ ] Post list with filters
  - [ ] Post detail page
  - [ ] Share buttons
  - [ ] Loading skeletons
  - [ ] Infinite scroll

### Comment System

- [ ] Create comments API
  - [ ] Add comment endpoint
  - [ ] Nested comments support
  - [ ] Edit/delete own comments
  - [ ] Reaction system
  - [ ] Real-time updates
- [ ] Build comment UI
  - [ ] Comment form
  - [ ] Comment thread display
  - [ ] Nested comment indentation
  - [ ] Reaction picker
  - [ ] Edit/delete actions
  - [ ] Time ago display

## ✅ Sprint 4: Real-time Features (Week 4)

### WebSocket Setup

- [ ] Configure Socket.io server
  - [ ] CORS settings
  - [ ] Authentication middleware
  - [ ] Room management
  - [ ] Connection pooling
- [ ] Create WebSocket client
  - [ ] Auto-reconnect logic
  - [ ] Event handlers
  - [ ] State management

### Chat Implementation

- [ ] Build chat backend
  - [ ] Message persistence
  - [ ] Room creation/joining
  - [ ] Typing indicators
  - [ ] Online status
  - [ ] Message delivery receipts
- [ ] Create chat UI
  - [ ] Chat room list
  - [ ] Message interface
  - [ ] Emoji picker
  - [ ] File sharing
  - [ ] Voice messages
  - [ ] Unread badges
  - [ ] Push notifications

### Real-time Updates

- [ ] Implement live features
  - [ ] New post notifications
  - [ ] Comment notifications
  - [ ] Like/reaction updates
  - [ ] User presence
  - [ ] Collaborative editing

## ✅ Sprint 5: Testing & Optimization (Week 5)

### Testing

- [ ] Unit tests
  - [ ] API endpoints (100% coverage)
  - [ ] React components (> 90%)
  - [ ] Utility functions (100%)
  - [ ] Custom hooks (100%)
- [ ] Integration tests
  - [ ] Auth flow
  - [ ] Post creation flow
  - [ ] Comment system
  - [ ] WebSocket events
- [ ] E2E tests
  - [ ] Critical user paths
  - [ ] Cross-browser testing
  - [ ] Mobile testing
  - [ ] Performance testing

### Performance Optimization

- [ ] Frontend optimization
  - [ ] Code splitting
  - [ ] Lazy loading
  - [ ] Image optimization
  - [ ] Bundle analysis
  - [ ] Service worker
  - [ ] PWA manifest
- [ ] Backend optimization
  - [ ] Database query optimization
  - [ ] Caching strategy (Redis)
  - [ ] CDN setup
  - [ ] Rate limiting
  - [ ] Load balancing

### Security Hardening

- [ ] Security audit
  - [ ] Dependency scanning
  - [ ] OWASP Top 10 check
  - [ ] SQL injection prevention
  - [ ] XSS prevention
  - [ ] CSRF protection
  - [ ] Content Security Policy
  - [ ] HTTPS enforcement

## ✅ Sprint 6: Deployment (Week 6)

### Production Setup

- [ ] Configure production environment
  - [ ] Environment variables
  - [ ] Database migration
  - [ ] Backup strategy
  - [ ] Monitoring setup
  - [ ] Error tracking (Sentry)
  - [ ] Analytics (GA4)
- [ ] CI/CD pipeline
  - [ ] GitHub Actions workflow
  - [ ] Automated testing
  - [ ] Build optimization
  - [ ] Deployment to Vercel
  - [ ] Post-deployment tests
- [ ] Documentation
  - [ ] API documentation
  - [ ] README updates
  - [ ] Deployment guide
  - [ ] Contributing guide
```

## 📁 Steering 파일 (프로젝트 가이드)

### product.md - 제품 비전

```markdown
# product.md - Product Vision

## 🎯 Product Overview

Community is a modern, AI-powered community platform where developers can share knowledge, collaborate in real-time, and learn from each other.

## 👥 Target Users

- Primary: Web developers learning AI integration
- Secondary: AI enthusiasts sharing projects
- Tertiary: Companies looking for AI talent

## 🌟 Key Features

1. **AI-Powered Content**: MDX support with AI code highlighting
2. **Real-time Collaboration**: Live chat and pair programming
3. **Knowledge Sharing**: Posts, tutorials, and code snippets
4. **Community Building**: Follow users, create groups, host events

## 📈 Success Metrics

- User engagement: 5+ posts/user/month
- Retention: 60% monthly active users
- Growth: 20% month-over-month
- Performance: < 3s page load time
```

### structure.md - 프로젝트 구조

```markdown
# structure.md - Project Structure

## 📂 Directory Organization

moai-community/
├── src/
│ ├── app/ # Next.js 15 App Router
│ ├── components/ # React components
│ ├── lib/ # Business logic
│ ├── hooks/ # Custom React hooks
│ ├── utils/ # Helper functions
│ └── types/ # TypeScript definitions
├── prisma/ # Database schema
├── public/ # Static assets
├── tests/ # Test files
└── docs/ # Documentation

## 📝 Naming Conventions

- Components: PascalCase (UserProfile.tsx)
- Utilities: camelCase (formatDate.ts)
- Constants: UPPER_SNAKE_CASE
- CSS Modules: kebab-case
- Database: snake_case

## 🎨 Code Style

- 2 spaces indentation
- Single quotes for strings
- Semicolons required
- Max line length: 80 chars
```

### tech.md - 기술 스택

```markdown
# tech.md - Technology Stack

## 🔧 Core Technologies

- **Framework**: Next.js 15 (App Router)
- **Language**: TypeScript 5.0
- **Styling**: Tailwind CSS 3.4
- **Database**: PostgreSQL 16
- **ORM**: Prisma 5.0
- **Authentication**: NextAuth.js 5.0

## 📦 Key Dependencies

- **State Management**: Zustand
- **Forms**: React Hook Form + Zod
- **UI Components**: Radix UI
- **Real-time**: Socket.io
- **Testing**: Jest + React Testing Library
- **E2E Testing**: Playwright

## 🛠️ Development Tools

- **Package Manager**: pnpm
- **Code Quality**: ESLint + Prettier
- **Git Hooks**: Husky + lint-staged
- **CI/CD**: GitHub Actions
- **Deployment**: Vercel
- **Monitoring**: Datadog

## ☁️ Infrastructure

- **Hosting**: Vercel (Frontend)
- **Database**: Supabase (PostgreSQL)
- **File Storage**: AWS S3
- **CDN**: CloudFront
- **Email**: SendGrid
```
