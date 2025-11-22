# Frontend Best Practices — Performance & Deployment

_Last updated: 2025-11-22_

## Performance Optimization

### Code Splitting & Dynamic Imports
```typescript
// Lazy load components
import { lazy, Suspense } from 'react';

const HeavyComponent = lazy(() => import('./HeavyComponent'));

export function App() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <HeavyComponent />
    </Suspense>
  );
}

// Next.js dynamic imports
import dynamic from 'next/dynamic';

const Component = dynamic(() => import('../components/Heavy'), {
  loading: () => <div>Loading...</div>
});
```

### Image Optimization
```typescript
import Image from 'next/image';

// Next.js Image component handles optimization
<Image
  src="/profile.png"
  alt="Profile"
  width={300}
  height={300}
  priority={false}
  loading="lazy"
/>

// Web standard picture element
<picture>
  <source srcSet="image.webp" type="image/webp" />
  <img src="image.jpg" alt="Description" />
</picture>
```

### Bundle Analysis
```json
{
  "scripts": {
    "analyze": "ANALYZE=true next build"
  }
}
```

## Web Vitals & Monitoring

### Core Web Vitals Integration
```typescript
import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals';

export function reportWebVitals(metric: any) {
  console.log(metric);
  // Send to analytics
  fetch('/api/analytics', {
    method: 'POST',
    body: JSON.stringify(metric)
  });
}

getCLS(reportWebVitals);
getFID(reportWebVitals);
getFCP(reportWebVitals);
getLCP(reportWebVitals);
getTTFB(reportWebVitals);
```

## Error Handling

### Error Boundary
```typescript
import { ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error(error, errorInfo);
    // Log to error tracking service
  }

  render() {
    if (this.state.hasError) {
      return <div>Something went wrong: {this.state.error?.message}</div>;
    }

    return this.props.children;
  }
}
```

### Error Tracking
```typescript
import * as Sentry from "@sentry/react";

Sentry.init({
  dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
  environment: process.env.NODE_ENV,
  tracesSampleRate: 1.0
});

// Captures errors automatically
export default Sentry.withProfiler(App);
```

## SEO & Meta Tags

### Next.js Metadata
```typescript
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'My App',
  description: 'App description',
  keywords: ['nextjs', 'react', 'seo'],
  openGraph: {
    type: 'website',
    url: 'https://example.com',
    title: 'My App',
    description: 'App description',
    images: [{
      url: '/og-image.png',
      width: 1200,
      height: 630
    }]
  }
};

export default function Page() {
  return <div>Content</div>;
}
```

### Structured Data (JSON-LD)
```typescript
export function ProductStructuredData() {
  const data = {
    '@context': 'https://schema.org/',
    '@type': 'Product',
    name: 'Product Name',
    description: 'Product description',
    image: 'image.jpg',
    offers: {
      '@type': 'Offer',
      price: '29.99',
      priceCurrency: 'USD'
    }
  };

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(data) }}
    />
  );
}
```

## Testing Best Practices

### Component Testing with Vitest
```typescript
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Button } from './Button';

describe('Button', () => {
  it('renders correctly', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  it('calls onClick handler', async () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click</Button>);

    const button = screen.getByRole('button');
    await button.click();

    expect(handleClick).toHaveBeenCalled();
  });
});
```

## Deployment Strategies

### Environment Variables
```bash
# .env.local
DATABASE_URL=postgresql://...
NEXT_PUBLIC_API_URL=https://api.example.com
```

### Docker Deployment
```dockerfile
FROM node:20-alpine

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY .next ./.next
COPY public ./public

EXPOSE 3000
CMD ["npm", "start"]
```

### Performance Checklist
- [ ] Enable compression (gzip)
- [ ] Use CDN for static assets
- [ ] Implement caching headers
- [ ] Lazy load images and components
- [ ] Minimize bundle size (<100KB gzipped)
- [ ] Optimize fonts (avoid @import)
- [ ] Remove unused dependencies
- [ ] Monitor Core Web Vitals
- [ ] Implement error tracking
- [ ] Set up performance budgets

---

**Last Updated**: 2025-11-22

